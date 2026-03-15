package dynatrace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// authProvider sets authorization headers on outgoing requests.
type authProvider interface {
	setAuth(req *http.Request) error
	authType() string
}

// tokenAuth uses the classic Api-Token header.
type tokenAuth struct {
	token string
}

func (a *tokenAuth) setAuth(req *http.Request) error {
	req.Header.Set("Authorization", "Api-Token "+a.token)
	return nil
}

func (a *tokenAuth) authType() string { return "token" }

// oauthAuth uses the OAuth2 client_credentials flow with Dynatrace SSO.
type oauthAuth struct {
	clientID     string
	clientSecret string
	tokenURL     string
	scopes       string
	httpClient   *http.Client

	mu          sync.Mutex
	accessToken string
	expiry      time.Time
}

const (
	defaultTokenURL = "https://sso.dynatrace.com/sso/oauth2/token"
	oauthScopes     = "document:documents:write document:documents:read settings:objects:write settings:objects:read settings:schemas:read"
)

func newOAuthAuth(clientID, clientSecret string, httpClient *http.Client) *oauthAuth {
	return &oauthAuth{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     defaultTokenURL,
		scopes:       oauthScopes,
		httpClient:   httpClient,
	}
}

func (a *oauthAuth) setAuth(req *http.Request) error {
	token, err := a.getToken()
	if err != nil {
		return fmt.Errorf("obtaining OAuth token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *oauthAuth) authType() string { return "oauth" }

func (a *oauthAuth) getToken() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Return cached token if still valid (with 30s buffer)
	if a.accessToken != "" && time.Now().Add(30*time.Second).Before(a.expiry) {
		return a.accessToken, nil
	}

	logging.Debug("OAuth: requesting new access token from %s", a.tokenURL)

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {a.clientID},
		"client_secret": {a.clientSecret},
		"scope":         {a.scopes},
	}

	resp, err := a.httpClient.Post(a.tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	a.accessToken = tokenResp.AccessToken
	a.expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	logging.Debug("OAuth: obtained token, expires in %ds", tokenResp.ExpiresIn)
	return a.accessToken, nil
}
