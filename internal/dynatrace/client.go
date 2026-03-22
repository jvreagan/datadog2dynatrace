package dynatrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
)

// Client handles communication with the Dynatrace API.
type Client struct {
	envURL      string
	apiToken    string
	httpClient  *http.Client
	limiter     *ratelimit.Limiter
	auth        authProvider
	platformURL string
	isGen3      bool
}

// NewClient creates a new Dynatrace API client using Api-Token auth.
func NewClient(envURL, apiToken string) *Client {
	return newClientWithConfig(envURL, apiToken, ratelimit.Config{
		RequestsPerSecond: 10,
		MaxRetries:        5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        60 * time.Second,
	})
}

// newClientWithConfig creates a client with custom rate limiter config (for testing).
func newClientWithConfig(envURL, apiToken string, cfg ratelimit.Config) *Client {
	envURL = strings.TrimRight(envURL, "/")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return &Client{
		envURL:     envURL,
		apiToken:   apiToken,
		httpClient: httpClient,
		limiter:    ratelimit.New(httpClient, cfg),
		auth:       &tokenAuth{token: apiToken},
	}
}

// NewOAuthClient creates a Dynatrace API client using OAuth2 client_credentials.
func NewOAuthClient(envURL, clientID, clientSecret string) *Client {
	envURL = strings.TrimRight(envURL, "/")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return &Client{
		envURL:      envURL,
		httpClient:  httpClient,
		limiter:     ratelimit.New(httpClient, ratelimit.Config{
			RequestsPerSecond: 10,
			MaxRetries:        5,
			InitialBackoff:    1 * time.Second,
			MaxBackoff:        60 * time.Second,
		}),
		auth:        newOAuthAuth(clientID, clientSecret, httpClient),
		platformURL: derivePlatformURL(envURL),
		isGen3:      true,
	}
}

// newOAuthClientWithConfig creates an OAuth client with custom rate limiter config (for testing).
func newOAuthClientWithConfig(envURL, clientID, clientSecret string, cfg ratelimit.Config) *Client {
	envURL = strings.TrimRight(envURL, "/")
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	oa := newOAuthAuth(clientID, clientSecret, httpClient)
	return &Client{
		envURL:      envURL,
		httpClient:  httpClient,
		limiter:     ratelimit.New(httpClient, cfg),
		auth:        oa,
		platformURL: derivePlatformURL(envURL),
		isGen3:      true,
	}
}

// NewTestClient creates a client with relaxed rate limiting, suitable for tests
// that use httptest servers.
func NewTestClient(envURL, apiToken string) *Client {
	return newClientWithConfig(envURL, apiToken, ratelimit.Config{
		RequestsPerSecond: 1000,
		MaxRetries:        0,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        10 * time.Millisecond,
	})
}

// derivePlatformURL converts an environment URL to the platform/apps URL.
// e.g. https://abc12345.live.dynatrace.com → https://abc12345.apps.dynatrace.com
func derivePlatformURL(envURL string) string {
	return strings.Replace(envURL, ".live.dynatrace.com", ".apps.dynatrace.com", 1)
}

// Validate checks that the API credentials are valid.
func (c *Client) Validate() error {
	if c.isGen3 {
		// For OAuth, validate by hitting the v2 cluster time endpoint
		req, err := http.NewRequest("GET", c.envURL+"/api/v2/time", nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		if err := c.setHeaders(req); err != nil {
			return fmt.Errorf("setting auth: %w", err)
		}

		resp, err := c.limiter.Do(req, nil)
		if err != nil {
			return fmt.Errorf("connecting to Dynatrace API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Dynatrace API validation failed (HTTP %d): %s", resp.StatusCode, string(body))
		}
		return nil
	}

	req, err := http.NewRequest("GET", c.envURL+"/api/v1/time", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return fmt.Errorf("connecting to Dynatrace API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Dynatrace API validation failed (HTTP %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) error {
	if err := c.auth.setAuth(req); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return nil
}

// get performs a GET request and returns the response body.
func (c *Client) get(path string) ([]byte, error) {
	logging.Debug("Dynatrace API GET %s", path)
	req, err := http.NewRequest("GET", c.envURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return nil, fmt.Errorf("API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// post performs a POST request with a JSON body.
func (c *Client) post(path string, body interface{}) ([]byte, error) {
	logging.Debug("Dynatrace API POST %s", path)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.envURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// put performs a PUT request with a JSON body.
func (c *Client) put(path string, body interface{}) ([]byte, error) {
	logging.Debug("Dynatrace API PUT %s", path)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("PUT", c.envURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// getPlatform performs a GET request against the platform/apps URL.
func (c *Client) getPlatform(path string) ([]byte, error) {
	logging.Debug("Dynatrace Platform API GET %s", path)
	req, err := http.NewRequest("GET", c.platformURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return nil, fmt.Errorf("platform API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("platform API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// postPlatformMultipart posts a multipart/form-data request to the platform/apps URL.
// fields is a map of plain text form fields; contentPart is sent as the "content" file part
// with Content-Type application/json.
func (c *Client) postPlatformMultipart(path string, fields map[string]string, contentJSON []byte) ([]byte, error) {
	logging.Debug("Dynatrace Platform API POST (multipart) %s", path)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("writing multipart field %s: %w", k, err)
		}
	}

	part, err := mw.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="content"; filename="document.json"`},
		"Content-Type":        {"application/json"},
	})
	if err != nil {
		return nil, fmt.Errorf("creating multipart content part: %w", err)
	}
	if _, err := part.Write(contentJSON); err != nil {
		return nil, fmt.Errorf("writing multipart content: %w", err)
	}
	mw.Close()

	req, err := http.NewRequest("POST", c.platformURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.auth.setAuth(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return nil, fmt.Errorf("platform API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("platform API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// postPlatform performs a POST request against the platform/apps URL.
func (c *Client) postPlatform(path string, body interface{}) ([]byte, error) {
	logging.Debug("Dynatrace Platform API POST %s", path)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.platformURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return nil, fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, jsonBody)
	if err != nil {
		return nil, fmt.Errorf("platform API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("platform API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// delete performs a DELETE request.
func (c *Client) delete(path string) error {
	logging.Debug("Dynatrace API DELETE %s", path)
	req, err := http.NewRequest("DELETE", c.envURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if err := c.setHeaders(req); err != nil {
		return fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return fmt.Errorf("API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // already gone
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}
	return nil
}

// deletePlatform performs a DELETE request against the platform/apps URL.
func (c *Client) deletePlatform(path string) error {
	logging.Debug("Dynatrace Platform API DELETE %s", path)
	req, err := http.NewRequest("DELETE", c.platformURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if err := c.auth.setAuth(req); err != nil {
		return fmt.Errorf("setting auth: %w", err)
	}

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return fmt.Errorf("platform API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // already gone
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("platform API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}
	return nil
}

// IsGen3 returns whether this client is configured for Gen3/Platform APIs.
func (c *Client) IsGen3() bool {
	return c.isGen3
}

// PushOptions configures the behavior of PushAll.
type PushOptions struct {
	// SkipExisting silently skips resources that already exist in Dynatrace.
	// Takes precedence over ConflictResolver when true.
	SkipExisting bool
	// ConflictResolver is called when a resource already exists and SkipExisting is false.
	// If nil and SkipExisting is false, existing resources are re-created (may duplicate).
	ConflictResolver ConflictResolver
}

// PushAll pushes all converted resources to Dynatrace.
func (c *Client) PushAll(result *ConversionResult) []error {
	return c.PushAllWithOptions(result, PushOptions{})
}

// PushAllWithOptions pushes all converted resources with configurable behavior.
func (c *Client) PushAllWithOptions(result *ConversionResult, opts PushOptions) []error {
	var errs []error

	// Pre-fetch existing resources when we need to detect conflicts.
	var existing map[string]map[string]string // resourceType -> name -> id
	if opts.SkipExisting || opts.ConflictResolver != nil {
		existing = c.fetchExistingResources()
	}

	// handleConflict checks if a resource exists and acts according to opts.
	// Returns true if the resource should be skipped (not created).
	handleConflict := func(resourceType, name string) bool {
		if existing == nil {
			return false
		}
		byName, ok := existing[resourceType]
		if !ok {
			return false
		}
		existingID, exists := byName[name]
		if !exists {
			return false
		}
		// Resource exists.
		if opts.SkipExisting {
			logging.Info("skipping existing %s: %q", resourceType, name)
			return true
		}
		if opts.ConflictResolver == nil {
			return false // re-create without deletion
		}
		action := opts.ConflictResolver(resourceType, name)
		if action == ConflictSkip {
			logging.Info("skipping existing %s: %q", resourceType, name)
			return true
		}
		// ConflictReplace: delete then create.
		if existingID == "" {
			logging.Warn("cannot replace %s %q: ID unknown, skipping", resourceType, name)
			return true
		}
		if err := c.deleteExisting(resourceType, existingID); err != nil {
			logging.Warn("failed to delete existing %s %q (%s): %v", resourceType, name, existingID, err)
		}
		return false
	}

	logging.Info("pushing %d dashboards to Dynatrace", len(result.Dashboards))
	for _, d := range result.Dashboards {
		if handleConflict("dashboard", d.DashboardMetadata.Name) {
			continue
		}
		if err := c.CreateDashboard(&d); err != nil {
			errs = append(errs, fmt.Errorf("dashboard %q: %w", d.DashboardMetadata.Name, err))
		}
	}

	logging.Info("pushing %d metric events to Dynatrace", len(result.MetricEvents))
	for _, me := range result.MetricEvents {
		if handleConflict("metric_event", me.Summary) {
			continue
		}
		if err := c.CreateMetricEvent(&me); err != nil {
			errs = append(errs, fmt.Errorf("metric event %q: %w", me.Summary, err))
		}
	}

	logging.Info("pushing %d SLOs to Dynatrace", len(result.SLOs))
	for _, s := range result.SLOs {
		if handleConflict("slo", s.Name) {
			continue
		}
		if err := c.CreateSLO(&s); err != nil {
			errs = append(errs, fmt.Errorf("SLO %q: %w", s.Name, err))
		}
	}

	logging.Info("pushing %d synthetics to Dynatrace", len(result.Synthetics))
	for _, sm := range result.Synthetics {
		if handleConflict("synthetic", sm.Name) {
			continue
		}
		if err := c.CreateSyntheticMonitor(&sm); err != nil {
			errs = append(errs, fmt.Errorf("synthetic %q: %w", sm.Name, err))
		}
	}

	logging.Info("pushing %d maintenance windows to Dynatrace", len(result.Maintenance))
	for _, mw := range result.Maintenance {
		if handleConflict("maintenance", mw.Name) {
			continue
		}
		if err := c.CreateMaintenanceWindow(&mw); err != nil {
			errs = append(errs, fmt.Errorf("maintenance window %q: %w", mw.Name, err))
		}
	}

	logging.Info("pushing %d notifications to Dynatrace", len(result.Notifications))
	for _, n := range result.Notifications {
		if handleConflict("notification", n.Name) {
			continue
		}
		if err := c.CreateNotification(&n); err != nil {
			errs = append(errs, fmt.Errorf("notification %q: %w", n.Name, err))
		}
	}

	logging.Info("pushing %d log processing rules to Dynatrace", len(result.LogRules))
	for _, r := range result.LogRules {
		if handleConflict("log_rule", r.Name) {
			continue
		}
		if err := c.CreateLogProcessingRule(&r); err != nil {
			errs = append(errs, fmt.Errorf("log processing rule %q: %w", r.Name, err))
		}
	}

	logging.Info("pushing %d metric descriptors to Dynatrace", len(result.Metrics))
	for _, md := range result.Metrics {
		if handleConflict("metric", md.MetricID) {
			continue
		}
		if err := c.CreateMetricDescriptor(&md); err != nil {
			errs = append(errs, fmt.Errorf("metric descriptor %q: %w", md.MetricID, err))
		}
	}

	logging.Info("pushing %d notebooks to Dynatrace", len(result.Notebooks))
	for _, nb := range result.Notebooks {
		if handleConflict("notebook", nb.Name) {
			continue
		}
		if err := c.CreateNotebook(&nb); err != nil {
			errs = append(errs, fmt.Errorf("notebook %q: %w", nb.Name, err))
		}
	}

	return errs
}

// fetchExistingResources pre-fetches existing resource names and IDs from Dynatrace.
// Returns a map of resourceType -> (name -> id). ID may be empty for resource types
// that don't support deletion.
func (c *Client) fetchExistingResources() map[string]map[string]string {
	result := map[string]map[string]string{}

	if resources, err := c.ListDashboardsWithIDs(); err == nil {
		result["dashboard"] = toNameIDMap(resources)
	}
	if resources, err := c.ListMetricEventsWithIDs(); err == nil {
		result["metric_event"] = toNameIDMap(resources)
	}
	// For resource types without a delete implementation, store name with empty ID.
	if names, err := c.ListSLONames(); err == nil {
		result["slo"] = namesToMap(names)
	}
	if names, err := c.ListSyntheticNames(); err == nil {
		result["synthetic"] = namesToMap(names)
	}
	if names, err := c.ListMaintenanceNames(); err == nil {
		result["maintenance"] = namesToMap(names)
	}
	if names, err := c.ListNotificationNames(); err == nil {
		result["notification"] = namesToMap(names)
	}
	if names, err := c.ListNotebookNames(); err == nil {
		result["notebook"] = namesToMap(names)
	}

	return result
}

// deleteExisting deletes an existing resource by type and ID.
func (c *Client) deleteExisting(resourceType, id string) error {
	switch resourceType {
	case "dashboard":
		return c.DeleteDashboard(id)
	case "metric_event":
		return c.DeleteMetricEvent(id)
	default:
		return fmt.Errorf("delete not implemented for resource type %q", resourceType)
	}
}

func toNameIDMap(resources []NamedResource) map[string]string {
	m := make(map[string]string, len(resources))
	for _, r := range resources {
		m[r.Name] = r.ID
	}
	return m
}

func namesToMap(names []string) map[string]string {
	m := make(map[string]string, len(names))
	for _, n := range names {
		m[n] = "" // ID unknown
	}
	return m
}
