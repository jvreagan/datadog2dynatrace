package dynatrace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
)

// Client handles communication with the Dynatrace API.
type Client struct {
	envURL     string
	apiToken   string
	httpClient *http.Client
	limiter    *ratelimit.Limiter
}

// NewClient creates a new Dynatrace API client.
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

// Validate checks that the API credentials are valid.
func (c *Client) Validate() error {
	req, err := http.NewRequest("GET", c.envURL+"/api/v1/time", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

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

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Api-Token "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}

// get performs a GET request and returns the response body.
func (c *Client) get(path string) ([]byte, error) {
	logging.Debug("Dynatrace API GET %s", path)
	req, err := http.NewRequest("GET", c.envURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

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
	c.setHeaders(req)

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
	c.setHeaders(req)

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

// PushOptions configures the behavior of PushAll.
type PushOptions struct {
	SkipExisting bool
}

// PushAll pushes all converted resources to Dynatrace.
func (c *Client) PushAll(result *ConversionResult) []error {
	return c.PushAllWithOptions(result, PushOptions{})
}

// PushAllWithOptions pushes all converted resources with configurable behavior.
func (c *Client) PushAllWithOptions(result *ConversionResult, opts PushOptions) []error {
	var errs []error

	// Pre-fetch existing names if skip-existing is enabled
	var existingNames map[string]map[string]bool
	if opts.SkipExisting {
		existingNames = c.fetchExistingNames()
	}

	skip := func(resourceType, name string) bool {
		if !opts.SkipExisting || existingNames == nil {
			return false
		}
		if names, ok := existingNames[resourceType]; ok && names[name] {
			logging.Info("skipping existing %s: %q", resourceType, name)
			return true
		}
		return false
	}

	logging.Info("pushing %d dashboards to Dynatrace", len(result.Dashboards))
	for _, d := range result.Dashboards {
		if skip("dashboard", d.DashboardMetadata.Name) {
			continue
		}
		if err := c.CreateDashboard(&d); err != nil {
			errs = append(errs, fmt.Errorf("dashboard %q: %w", d.DashboardMetadata.Name, err))
		}
	}

	logging.Info("pushing %d metric events to Dynatrace", len(result.MetricEvents))
	for _, me := range result.MetricEvents {
		if skip("metric_event", me.Summary) {
			continue
		}
		if err := c.CreateMetricEvent(&me); err != nil {
			errs = append(errs, fmt.Errorf("metric event %q: %w", me.Summary, err))
		}
	}

	logging.Info("pushing %d SLOs to Dynatrace", len(result.SLOs))
	for _, s := range result.SLOs {
		if skip("slo", s.Name) {
			continue
		}
		if err := c.CreateSLO(&s); err != nil {
			errs = append(errs, fmt.Errorf("SLO %q: %w", s.Name, err))
		}
	}

	logging.Info("pushing %d synthetics to Dynatrace", len(result.Synthetics))
	for _, sm := range result.Synthetics {
		if skip("synthetic", sm.Name) {
			continue
		}
		if err := c.CreateSyntheticMonitor(&sm); err != nil {
			errs = append(errs, fmt.Errorf("synthetic %q: %w", sm.Name, err))
		}
	}

	logging.Info("pushing %d maintenance windows to Dynatrace", len(result.Maintenance))
	for _, mw := range result.Maintenance {
		if skip("maintenance", mw.Name) {
			continue
		}
		if err := c.CreateMaintenanceWindow(&mw); err != nil {
			errs = append(errs, fmt.Errorf("maintenance window %q: %w", mw.Name, err))
		}
	}

	logging.Info("pushing %d notifications to Dynatrace", len(result.Notifications))
	for _, n := range result.Notifications {
		if skip("notification", n.Name) {
			continue
		}
		if err := c.CreateNotification(&n); err != nil {
			errs = append(errs, fmt.Errorf("notification %q: %w", n.Name, err))
		}
	}

	logging.Info("pushing %d log processing rules to Dynatrace", len(result.LogRules))
	for _, r := range result.LogRules {
		if skip("log_rule", r.Name) {
			continue
		}
		if err := c.CreateLogProcessingRule(&r); err != nil {
			errs = append(errs, fmt.Errorf("log processing rule %q: %w", r.Name, err))
		}
	}

	logging.Info("pushing %d metric descriptors to Dynatrace", len(result.Metrics))
	for _, md := range result.Metrics {
		if skip("metric", md.MetricID) {
			continue
		}
		if err := c.CreateMetricDescriptor(&md); err != nil {
			errs = append(errs, fmt.Errorf("metric descriptor %q: %w", md.MetricID, err))
		}
	}

	logging.Info("pushing %d notebooks to Dynatrace", len(result.Notebooks))
	for _, nb := range result.Notebooks {
		if skip("notebook", nb.Name) {
			continue
		}
		if err := c.CreateNotebook(&nb); err != nil {
			errs = append(errs, fmt.Errorf("notebook %q: %w", nb.Name, err))
		}
	}

	return errs
}

// fetchExistingNames pre-fetches existing resource names from Dynatrace.
func (c *Client) fetchExistingNames() map[string]map[string]bool {
	result := map[string]map[string]bool{}

	if names, err := c.ListDashboardNames(); err == nil {
		result["dashboard"] = toNameSet(names)
	}
	if names, err := c.ListSLONames(); err == nil {
		result["slo"] = toNameSet(names)
	}
	if names, err := c.ListSyntheticNames(); err == nil {
		result["synthetic"] = toNameSet(names)
	}
	if names, err := c.ListMaintenanceNames(); err == nil {
		result["maintenance"] = toNameSet(names)
	}
	if names, err := c.ListNotificationNames(); err == nil {
		result["notification"] = toNameSet(names)
	}
	if names, err := c.ListMetricEventSummaries(); err == nil {
		result["metric_event"] = toNameSet(names)
	}
	if names, err := c.ListNotebookNames(); err == nil {
		result["notebook"] = toNameSet(names)
	}

	return result
}

func toNameSet(names []string) map[string]bool {
	s := make(map[string]bool, len(names))
	for _, n := range names {
		s[n] = true
	}
	return s
}
