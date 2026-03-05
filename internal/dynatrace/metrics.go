package dynatrace

import "fmt"

// CreateMetricDescriptor creates/updates metric metadata in Dynatrace.
func (c *Client) CreateMetricDescriptor(md *MetricDescriptor) error {
	_, err := c.put(fmt.Sprintf("/api/v2/metrics/%s", md.MetricID), md)
	if err != nil {
		return fmt.Errorf("creating metric descriptor: %w", err)
	}
	return nil
}
