package dynatrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client handles communication with the Dynatrace API.
type Client struct {
	envURL     string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Dynatrace API client.
func NewClient(envURL, apiToken string) *Client {
	envURL = strings.TrimRight(envURL, "/")
	return &Client{
		envURL:   envURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Validate checks that the API credentials are valid.
func (c *Client) Validate() error {
	req, err := http.NewRequest("GET", c.envURL+"/api/v1/time", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
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

// post performs a POST request with a JSON body.
func (c *Client) post(path string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.envURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
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
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("PUT", c.envURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
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

// PushAll pushes all converted resources to Dynatrace.
func (c *Client) PushAll(result *ConversionResult) []error {
	var errs []error

	for _, d := range result.Dashboards {
		if err := c.CreateDashboard(&d); err != nil {
			errs = append(errs, fmt.Errorf("dashboard %q: %w", d.DashboardMetadata.Name, err))
		}
	}

	for _, me := range result.MetricEvents {
		if err := c.CreateMetricEvent(&me); err != nil {
			errs = append(errs, fmt.Errorf("metric event %q: %w", me.Summary, err))
		}
	}

	for _, s := range result.SLOs {
		if err := c.CreateSLO(&s); err != nil {
			errs = append(errs, fmt.Errorf("SLO %q: %w", s.Name, err))
		}
	}

	for _, sm := range result.Synthetics {
		if err := c.CreateSyntheticMonitor(&sm); err != nil {
			errs = append(errs, fmt.Errorf("synthetic %q: %w", sm.Name, err))
		}
	}

	for _, mw := range result.Maintenance {
		if err := c.CreateMaintenanceWindow(&mw); err != nil {
			errs = append(errs, fmt.Errorf("maintenance window %q: %w", mw.Name, err))
		}
	}

	for _, n := range result.Notifications {
		if err := c.CreateNotification(&n); err != nil {
			errs = append(errs, fmt.Errorf("notification %q: %w", n.Name, err))
		}
	}

	return errs
}
