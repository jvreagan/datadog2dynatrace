package datadog

import (
	"encoding/json"
	"fmt"
)

// GetDashboards retrieves all dashboards with full details.
func (c *Client) GetDashboards() ([]Dashboard, error) {
	list, err := c.GetDashboardList()
	if err != nil {
		return nil, err
	}

	var dashboards []Dashboard
	for _, d := range list {
		full, err := c.GetDashboard(d.ID)
		if err != nil {
			return nil, fmt.Errorf("getting dashboard %s: %w", d.ID, err)
		}
		dashboards = append(dashboards, *full)
	}
	return dashboards, nil
}

// GetDashboard retrieves a single dashboard by ID.
func (c *Client) GetDashboard(id string) (*Dashboard, error) {
	data, err := c.get(fmt.Sprintf("/api/v1/dashboard/%s", id))
	if err != nil {
		return nil, err
	}
	var d Dashboard
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing dashboard: %w", err)
	}
	return &d, nil
}
