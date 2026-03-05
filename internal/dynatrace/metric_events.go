package dynatrace

import "fmt"

// CreateMetricEvent creates a metric event (alert) in Dynatrace.
func (c *Client) CreateMetricEvent(me *MetricEvent) error {
	_, err := c.post("/api/config/v1/anomalyDetection/metricEvents", me)
	if err != nil {
		return fmt.Errorf("creating metric event: %w", err)
	}
	return nil
}
