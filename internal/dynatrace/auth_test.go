package dynatrace

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTokenAuthSetAuth(t *testing.T) {
	a := &tokenAuth{token: "dt0c01.test"}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	if err := a.setAuth(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := req.Header.Get("Authorization")
	if got != "Api-Token dt0c01.test" {
		t.Errorf("expected 'Api-Token dt0c01.test', got %q", got)
	}
}

func TestTokenAuthType(t *testing.T) {
	a := &tokenAuth{token: "x"}
	if a.authType() != "token" {
		t.Errorf("expected 'token', got %q", a.authType())
	}
}

func TestOAuthAuthType(t *testing.T) {
	a := newOAuthAuth("id", "secret", http.DefaultClient)
	if a.authType() != "oauth" {
		t.Errorf("expected 'oauth', got %q", a.authType())
	}
}

func TestOAuthTokenExchange(t *testing.T) {
	var gotContentType string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		r.ParseForm()
		gotBody = r.Form.Encode()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"test-bearer-token","expires_in":300,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("my-client-id", "my-client-secret", srv.Client())
	a.tokenURL = srv.URL

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	if err := a.setAuth(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := req.Header.Get("Authorization")
	if got != "Bearer test-bearer-token" {
		t.Errorf("expected 'Bearer test-bearer-token', got %q", got)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "grant_type=client_credentials") {
		t.Errorf("expected client_credentials grant_type in body, got %q", gotBody)
	}
}

func TestOAuthTokenCaching(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"cached-token","expires_in":300,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("id", "secret", srv.Client())
	a.tokenURL = srv.URL

	// First call gets a token
	token1, err := a.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Second call should use cached token
	token2, err := a.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token1 != token2 {
		t.Errorf("expected same cached token")
	}
	if callCount != 1 {
		t.Errorf("expected 1 token request (cached), got %d", callCount)
	}
}

func TestOAuthTokenRefresh(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"refreshed-token","expires_in":1,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("id", "secret", srv.Client())
	a.tokenURL = srv.URL

	// Get first token
	_, err := a.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Force expiry (set expiry to past, since 30s buffer makes 1s expire immediately)
	a.mu.Lock()
	a.expiry = time.Now().Add(-1 * time.Second)
	a.mu.Unlock()

	// Should refresh
	_, err = a.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 token requests (refreshed), got %d", callCount)
	}
}

func TestOAuthTokenExchangeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("bad-id", "bad-secret", srv.Client())
	a.tokenURL = srv.URL

	_, err := a.getToken()
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got %q", err.Error())
	}
}

func TestOAuthEmptyToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"","expires_in":300}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("id", "secret", srv.Client())
	a.tokenURL = srv.URL

	_, err := a.getToken()
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !strings.Contains(err.Error(), "empty access token") {
		t.Errorf("expected 'empty access token' in error, got %q", err.Error())
	}
}

func TestOAuthBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	a := newOAuthAuth("id", "secret", srv.Client())
	a.tokenURL = srv.URL

	_, err := a.getToken()
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("expected parsing error, got %q", err.Error())
	}
}

func TestOAuthConcurrentAccess(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"concurrent-token","expires_in":300,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	a := newOAuthAuth("id", "secret", srv.Client())
	a.tokenURL = srv.URL

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := a.getToken()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	// Due to mutex, most goroutines should find the cached token
	mu.Lock()
	defer mu.Unlock()
	if callCount > 2 {
		t.Errorf("expected at most 2 token requests (mutex-protected), got %d", callCount)
	}
}
