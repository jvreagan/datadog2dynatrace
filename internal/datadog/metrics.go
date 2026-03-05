package datadog

import (
	"encoding/json"
	"fmt"
)

// GetMetrics retrieves metric names and their metadata.
func (c *Client) GetMetrics(query string) ([]MetricMetadata, error) {
	// First get the list of active metrics
	data, err := c.get(fmt.Sprintf("/api/v1/metrics?from=%d", 0))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Metrics []string `json:"metrics"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing metrics list: %w", err)
	}

	var metrics []MetricMetadata
	for _, name := range resp.Metrics {
		meta, err := c.GetMetricMetadata(name)
		if err != nil {
			// Non-fatal: some metrics may not have metadata
			continue
		}
		meta.Metric = name
		metrics = append(metrics, *meta)
	}
	return metrics, nil
}

// GetMetricMetadata retrieves metadata for a specific metric.
func (c *Client) GetMetricMetadata(name string) (*MetricMetadata, error) {
	data, err := c.get(fmt.Sprintf("/api/v1/metrics/%s", name))
	if err != nil {
		return nil, err
	}
	var m MetricMetadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing metric metadata: %w", err)
	}
	return &m, nil
}
