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
	fields := map[string]string{
		"name":      d.DashboardMetadata.Name,
		"type":      "dashboard",
		"isPrivate": "false",
	}
	_, err = c.postPlatformMultipart("/platform/document/v1/documents", fields, content)
	if err != nil {
		return fmt.Errorf("creating dashboard via Documents API: %w", err)
	}
	return nil
}

// ListDashboardNames returns the names of all existing dashboards.
func (c *Client) ListDashboardNames() ([]string, error) {
	resources, err := c.ListDashboardsWithIDs()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}
	return names, nil
}

// ListDashboardsWithIDs returns existing dashboards with their IDs and names.
func (c *Client) ListDashboardsWithIDs() ([]NamedResource, error) {
	if c.isGen3 {
		return c.listDashboardsWithIDsGen3()
	}
	data, err := c.get("/api/config/v1/dashboards")
	if err != nil {
		return nil, fmt.Errorf("listing dashboards: %w", err)
	}
	var resp struct {
		Dashboards []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"dashboards"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing dashboard list: %w", err)
	}
	resources := make([]NamedResource, len(resp.Dashboards))
	for i, d := range resp.Dashboards {
		resources[i] = NamedResource{ID: d.ID, Name: d.Name}
	}
	return resources, nil
}

func (c *Client) listDashboardsWithIDsGen3() ([]NamedResource, error) {
	data, err := c.getPlatform("/platform/document/v1/documents?filter=type%3D%3D%27dashboard%27")
	if err != nil {
		return nil, fmt.Errorf("listing dashboards via Documents API: %w", err)
	}
	var resp struct {
		Documents []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"documents"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing document list: %w", err)
	}
	resources := make([]NamedResource, len(resp.Documents))
	for i, d := range resp.Documents {
		resources[i] = NamedResource{ID: d.ID, Name: d.Name}
	}
	return resources, nil
}

// DeleteDashboard deletes a dashboard by ID.
func (c *Client) DeleteDashboard(id string) error {
	if c.isGen3 {
		return c.deletePlatform("/platform/document/v1/documents/" + id)
	}
	return c.delete("/api/config/v1/dashboards/" + id)
}
