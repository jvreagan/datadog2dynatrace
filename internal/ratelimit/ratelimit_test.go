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
