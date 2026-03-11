package ratelimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fastConfig returns a Config with very short backoffs suitable for testing.
func fastConfig() Config {
	return Config{
		RequestsPerSecond: 1000, // effectively no rate limit
		MaxRetries:        3,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        10 * time.Millisecond,
	}
}

func init() {
	// Suppress log output during tests.
	SetLogWriter(io.Discard)
}

func TestRateLimitEnforcesMinInterval(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := Config{
		RequestsPerSecond: 20, // 50ms between requests
		MaxRetries:        0,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        10 * time.Millisecond,
	}
	limiter := New(ts.Client(), cfg)

	numRequests := 5
	start := time.Now()
	for i := 0; i < numRequests; i++ {
		req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
		resp, err := limiter.Do(req, nil)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}
	elapsed := time.Since(start)

	// 5 requests at 20/sec should take at least 4 intervals of 50ms = 200ms
	expectedMin := time.Duration(numRequests-1) * (50 * time.Millisecond)
	if elapsed < expectedMin {
		t.Errorf("requests completed too fast: %v, expected at least %v", elapsed, expectedMin)
	}

	if got := atomic.LoadInt32(&count); got != int32(numRequests) {
		t.Errorf("expected %d requests, got %d", numRequests, got)
	}
}

func TestRetryOn429(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetryOn429RespectsRetryAfterHeader(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)

	start := time.Now()
	resp, err := limiter.Do(req, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Retry-After: 1 means wait 1 second.
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected to wait ~1s for Retry-After, but only waited %v", elapsed)
	}
}

func TestRetryOn5xx(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("recovered"))
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestMaxRetriesExhausted429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	cfg := fastConfig()
	cfg.MaxRetries = 2
	limiter := New(ts.Client(), cfg)
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	_, err := limiter.Do(req, nil)
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	if !strings.Contains(err.Error(), "max retries") {
		t.Errorf("expected max retries error, got: %v", err)
	}
}

func TestMaxRetriesExhausted5xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	cfg := fastConfig()
	cfg.MaxRetries = 2
	limiter := New(ts.Client(), cfg)
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	_, err := limiter.Do(req, nil)
	if err == nil {
		t.Fatal("expected error after max retries")
	}
	if !strings.Contains(err.Error(), "server error 502") {
		t.Errorf("expected server error, got: %v", err)
	}
}

func Test4xxNotRetried(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Errorf("4xx should not be retried, expected 1 attempt, got %d", got)
	}
}

func TestPostBodyRetried(t *testing.T) {
	var attempts int32
	var lastBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	body := []byte(`{"key":"value"}`)
	req, _ := http.NewRequest("POST", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if lastBody != `{"key":"value"}` {
		t.Errorf("body not preserved on retry, got: %s", lastBody)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetryAfterInvalidHeader(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "not-a-number-or-date")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// Should fall back to exponential backoff and succeed
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetryAfterHTTPDate(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			// Set Retry-After to an HTTP-date 2 seconds in the future
			futureTime := time.Now().Add(2 * time.Second)
			w.Header().Set("Retry-After", futureTime.UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)

	start := time.Now()
	resp, err := limiter.Do(req, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// HTTP-date header was parsed, so wait should be non-trivial
	// (at least 500ms, accounting for clock resolution and processing time)
	if elapsed < 500*time.Millisecond {
		t.Errorf("expected to wait for HTTP-date Retry-After, but only waited %v", elapsed)
	}
}

func TestRetryAfterPastDate(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			// Set Retry-After to a past date — should fall back to backoff
			pastTime := time.Now().Add(-10 * time.Second)
			w.Header().Set("Retry-After", pastTime.UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	limiter := New(ts.Client(), fastConfig())
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	resp, err := limiter.Do(req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSetLogWriterNil(t *testing.T) {
	// SetLogWriter(nil) should set logWriter to io.Discard
	SetLogWriter(nil)
	if logWriter != io.Discard {
		t.Error("expected logWriter to be io.Discard after SetLogWriter(nil)")
	}
}

func TestSetLogWriterCustom(t *testing.T) {
	var buf strings.Builder
	SetLogWriter(&buf)
	if logWriter != &buf {
		t.Error("expected logWriter to be set to custom writer")
	}
	// Reset
	SetLogWriter(io.Discard)
}

func TestBackoffCappedAtMax(t *testing.T) {
	cfg := Config{
		RequestsPerSecond: 1000,
		MaxRetries:        10,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        5 * time.Millisecond,
	}
	limiter := New(http.DefaultClient, cfg)

	// At a high attempt, backoff should be capped at MaxBackoff (+25% jitter max)
	d := limiter.backoff(20)
	maxAllowed := time.Duration(float64(cfg.MaxBackoff) * 1.25)
	if d > maxAllowed {
		t.Errorf("backoff %v exceeds max allowed %v", d, maxAllowed)
	}
	if d < cfg.MaxBackoff {
		t.Errorf("backoff %v should be at least MaxBackoff %v", d, cfg.MaxBackoff)
	}
}

func TestNoRetriesConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	cfg := fastConfig()
	cfg.MaxRetries = 0
	limiter := New(ts.Client(), cfg)
	req, _ := http.NewRequest("GET", ts.URL+"/test", nil)
	_, err := limiter.Do(req, nil)
	if err == nil {
		t.Fatal("expected error with 0 retries")
	}
	if !strings.Contains(err.Error(), "max retries (0)") {
		t.Errorf("expected max retries (0) in error, got: %v", err)
	}
}
