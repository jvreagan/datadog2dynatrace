package dynatrace

import "fmt"

// CreateDashboard creates a dashboard in Dynatrace.
func (c *Client) CreateDashboard(d *Dashboard) error {
	_, err := c.post("/api/config/v1/dashboards", d)
	if err != nil {
		return fmt.Errorf("creating dashboard: %w", err)
	}
	return nil
}
