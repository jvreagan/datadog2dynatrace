package datadog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/ratelimit"
)

// Client handles communication with the DataDog API.
type Client struct {
	apiKey     string
	appKey     string
	baseURL    string
	httpClient *http.Client
	limiter    *ratelimit.Limiter
}

// NewClient creates a new DataDog API client.
func NewClient(apiKey, appKey, site string) *Client {
	if site == "" {
		site = "datadoghq.com"
	}
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return &Client{
		apiKey:     apiKey,
		appKey:     appKey,
		baseURL:    fmt.Sprintf("https://api.%s", site),
		httpClient: httpClient,
		limiter: ratelimit.New(httpClient, ratelimit.Config{
			RequestsPerSecond: 5,
			MaxRetries:        5,
			InitialBackoff:    1 * time.Second,
			MaxBackoff:        60 * time.Second,
		}),
	}
}

// Validate checks that the API credentials are valid.
func (c *Client) Validate() error {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/validate", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return fmt.Errorf("connecting to DataDog API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DataDog API validation failed (HTTP %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("DD-API-KEY", c.apiKey)
	req.Header.Set("DD-APPLICATION-KEY", c.appKey)
	req.Header.Set("Content-Type", "application/json")
}

// get performs a GET request and returns the response body.
func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.limiter.Do(req, nil)
	if err != nil {
		return nil, fmt.Errorf("API request to %s: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request to %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}

	return body, nil
}

// getPaginated performs paginated GET requests, accumulating results.
func (c *Client) getPaginated(path string, pageSize int, extractItems func([]byte) (int, error)) error {
	page := 0
	for {
		url := fmt.Sprintf("%s?page[size]=%d&page[number]=%d", path, pageSize, page)
		data, err := c.get(url)
		if err != nil {
			return err
		}

		count, err := extractItems(data)
		if err != nil {
			return fmt.Errorf("extracting items from page %d: %w", page, err)
		}

		if count < pageSize {
			break
		}
		page++
	}
	return nil
}

// ExtractAll extracts all supported resource types from DataDog.
func (c *Client) ExtractAll() (*ExtractionResult, error) {
	result := &ExtractionResult{}
	var errs []error

	if dashboards, err := c.GetDashboards(); err != nil {
		errs = append(errs, fmt.Errorf("dashboards: %w", err))
	} else {
		result.Dashboards = dashboards
	}

	if monitors, err := c.GetMonitors(); err != nil {
		errs = append(errs, fmt.Errorf("monitors: %w", err))
	} else {
		result.Monitors = monitors
	}

	if slos, err := c.GetSLOs(); err != nil {
		errs = append(errs, fmt.Errorf("SLOs: %w", err))
	} else {
		result.SLOs = slos
	}

	if synthetics, err := c.GetSynthetics(); err != nil {
		errs = append(errs, fmt.Errorf("synthetics: %w", err))
	} else {
		result.Synthetics = synthetics
	}

	if pipelines, err := c.GetLogPipelines(); err != nil {
		errs = append(errs, fmt.Errorf("log pipelines: %w", err))
	} else {
		result.LogPipelines = pipelines
	}

	if downtimes, err := c.GetDowntimes(); err != nil {
		errs = append(errs, fmt.Errorf("downtimes: %w", err))
	} else {
		result.Downtimes = downtimes
	}

	if notebooks, err := c.GetNotebooks(); err != nil {
		errs = append(errs, fmt.Errorf("notebooks: %w", err))
	} else {
		result.Notebooks = notebooks
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("extraction errors: %v", errs)
	}
	return result, nil
}

// GetDashboardList returns the list of dashboard summaries (id + title).
func (c *Client) GetDashboardList() ([]Dashboard, error) {
	data, err := c.get("/api/v1/dashboard")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Dashboards []Dashboard `json:"dashboards"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing dashboard list: %w", err)
	}
	return resp.Dashboards, nil
}
