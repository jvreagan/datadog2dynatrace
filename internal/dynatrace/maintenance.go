package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateMaintenanceWindow creates a maintenance window in Dynatrace.
func (c *Client) CreateMaintenanceWindow(mw *MaintenanceWindow) error {
	_, err := c.post("/api/config/v1/maintenanceWindows", mw)
	if err != nil {
		return fmt.Errorf("creating maintenance window: %w", err)
	}
	return nil
}

// ListMaintenanceNames returns the names of all existing maintenance windows.
func (c *Client) ListMaintenanceNames() ([]string, error) {
	data, err := c.get("/api/config/v1/maintenanceWindows")
	if err != nil {
		return nil, fmt.Errorf("listing maintenance windows: %w", err)
	}
	var resp struct {
		Values []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing maintenance window list: %w", err)
	}
	names := make([]string, len(resp.Values))
	for i, v := range resp.Values {
		names[i] = v.Name
	}
	return names, nil
}
