package dynatrace

import "fmt"

// CreateMaintenanceWindow creates a maintenance window in Dynatrace.
func (c *Client) CreateMaintenanceWindow(mw *MaintenanceWindow) error {
	_, err := c.post("/api/config/v1/maintenanceWindows", mw)
	if err != nil {
		return fmt.Errorf("creating maintenance window: %w", err)
	}
	return nil
}
