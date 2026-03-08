package ratelimit

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Config controls rate limiting and retry behavior.
type Config struct {
	// RequestsPerSecond is the maximum number of requests per second.
	RequestsPerSecond float64
	// MaxRetries is the maximum number of retry attempts for retryable errors.
	MaxRetries int
	// InitialBackoff is the starting backoff duration for exponential backoff.
	InitialBackoff time.Duration
	// MaxBackoff caps the maximum backoff duration.
	MaxBackoff time.Duration
}

// Limiter wraps an *http.Client with rate limiting and retry logic.
type Limiter struct {
	client      *http.Client
	config      Config
	minInterval time.Duration
	mu          sync.Mutex
	lastRequest time.Time
}

// New creates a new Limiter wrapping the given HTTP client.
func New(client *http.Client, cfg Config) *Limiter {
	minInterval := time.Duration(float64(time.Second) / cfg.RequestsPerSecond)
	return &Limiter{
		client:      client,
		config:      cfg,
		minInterval: minInterval,
	}
}

// Do executes the HTTP request with rate limiting and retry logic.
// The body parameter allows POST/PUT requests to be retried since the
// request body reader is consumed on the first attempt.
func (l *Limiter) Do(req *http.Request, body []byte) (*http.Response, error) {
	for attempt := 0; attempt <= l.config.MaxRetries; attempt++ {
		l.waitForRateLimit()

		// Reset the request body for retries.
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.ContentLength = int64(len(body))
		}

		resp, err := l.client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt == l.config.MaxRetries {
				return nil, fmt.Errorf("rate limited: max retries (%d) exceeded for %s", l.config.MaxRetries, req.URL.Path)
			}
			wait := l.retryAfterDuration(resp, attempt)
			fmt.Fprintf(logWriter, "rate limited on %s, retrying in %v (attempt %d/%d)\n", req.URL.Path, wait, attempt+1, l.config.MaxRetries)
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt == l.config.MaxRetries {
				return nil, fmt.Errorf("server error %d: max retries (%d) exceeded for %s", resp.StatusCode, l.config.MaxRetries, req.URL.Path)
			}
			wait := l.backoff(attempt)
			fmt.Fprintf(logWriter, "server error %d on %s, retrying in %v (attempt %d/%d)\n", resp.StatusCode, req.URL.Path, wait, attempt+1, l.config.MaxRetries)
			time.Sleep(wait)
			continue
		}

		return resp, nil
	}

	// Should not be reached, but just in case.
	return nil, fmt.Errorf("unexpected: max retries exceeded for %s", req.URL.Path)
}

// waitForRateLimit enforces the minimum interval between requests.
func (l *Limiter) waitForRateLimit() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.lastRequest.IsZero() {
		elapsed := time.Since(l.lastRequest)
		if elapsed < l.minInterval {
			time.Sleep(l.minInterval - elapsed)
		}
	}
	l.lastRequest = time.Now()
}

// retryAfterDuration parses the Retry-After header or falls back to exponential backoff.
func (l *Limiter) retryAfterDuration(resp *http.Response, attempt int) time.Duration {
	if val := resp.Header.Get("Retry-After"); val != "" {
		// Try parsing as seconds.
		if seconds, err := strconv.Atoi(val); err == nil {
			return time.Duration(seconds) * time.Second
		}
		// Try parsing as HTTP-date.
		if t, err := http.ParseTime(val); err == nil {
			d := time.Until(t)
			if d > 0 {
				return d
			}
		}
	}
	return l.backoff(attempt)
}

// backoff calculates exponential backoff with jitter.
func (l *Limiter) backoff(attempt int) time.Duration {
	backoff := float64(l.config.InitialBackoff) * math.Pow(2, float64(attempt))
	if backoff > float64(l.config.MaxBackoff) {
		backoff = float64(l.config.MaxBackoff)
	}
	// Add 0-25% random jitter.
	jitter := backoff * 0.25 * rand.Float64()
	return time.Duration(backoff + jitter)
}

// logWriter is the writer used for retry log messages.
var logWriter io.Writer = io.Discard

// SetLogWriter sets the writer for log messages. Pass nil to disable logging.
func SetLogWriter(w io.Writer) {
	if w == nil {
		logWriter = io.Discard
	} else {
		logWriter = w
	}
}
