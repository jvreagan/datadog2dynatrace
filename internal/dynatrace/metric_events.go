package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateMetricEvent creates a metric event (alert) in Dynatrace.
func (c *Client) CreateMetricEvent(me *MetricEvent) error {
	_, err := c.post("/api/config/v1/anomalyDetection/metricEvents", me)
	if err != nil {
		return fmt.Errorf("creating metric event: %w", err)
	}
	return nil
}

// ListMetricEventSummaries returns the summaries of all existing metric events.
func (c *Client) ListMetricEventSummaries() ([]string, error) {
	data, err := c.get("/api/config/v1/anomalyDetection/metricEvents")
	if err != nil {
		return nil, fmt.Errorf("listing metric events: %w", err)
	}
	var resp struct {
		Values []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing metric event list: %w", err)
	}
	names := make([]string, len(resp.Values))
	for i, v := range resp.Values {
		names[i] = v.Name
	}
	return names, nil
}
