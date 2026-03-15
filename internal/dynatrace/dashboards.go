package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateDashboard creates a dashboard in Dynatrace.
func (c *Client) CreateDashboard(d *Dashboard) error {
	if c.isGen3 {
		return c.createDashboardGen3(d)
	}
	_, err := c.post("/api/config/v1/dashboards", d)
	if err != nil {
		return fmt.Errorf("creating dashboard: %w", err)
	}
	return nil
}

func (c *Client) createDashboardGen3(d *Dashboard) error {
	content, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("marshaling dashboard content: %w", err)
	}
	doc := DocumentRequest{
		Name:      d.DashboardMetadata.Name,
		Type:      "dashboard",
		Content:   string(content),
		IsPrivate: false,
	}
	_, err = c.postPlatform("/platform/document/v1/documents", doc)
	if err != nil {
		return fmt.Errorf("creating dashboard via Documents API: %w", err)
	}
	return nil
}

// ListDashboardNames returns the names of all existing dashboards.
func (c *Client) ListDashboardNames() ([]string, error) {
	if c.isGen3 {
		return c.listDashboardNamesGen3()
	}
	data, err := c.get("/api/config/v1/dashboards")
	if err != nil {
		return nil, fmt.Errorf("listing dashboards: %w", err)
	}
	var resp struct {
		Dashboards []struct {
			Name string `json:"name"`
		} `json:"dashboards"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing dashboard list: %w", err)
	}
	names := make([]string, len(resp.Dashboards))
	for i, d := range resp.Dashboards {
		names[i] = d.Name
	}
	return names, nil
}

func (c *Client) listDashboardNamesGen3() ([]string, error) {
	data, err := c.getPlatform("/platform/document/v1/documents?filter=type%3D%3D%22dashboard%22")
	if err != nil {
		return nil, fmt.Errorf("listing dashboards via Documents API: %w", err)
	}
	var resp struct {
		Documents []struct {
			Name string `json:"name"`
		} `json:"documents"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing document list: %w", err)
	}
	names := make([]string, len(resp.Documents))
	for i, d := range resp.Documents {
		names[i] = d.Name
	}
	return names, nil
}
