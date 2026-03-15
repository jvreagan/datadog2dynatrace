package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateMaintenanceWindow creates a maintenance window in Dynatrace.
func (c *Client) CreateMaintenanceWindow(mw *MaintenanceWindow) error {
	if c.isGen3 {
		return c.createMaintenanceWindowGen3(mw)
	}
	_, err := c.post("/api/config/v1/maintenanceWindows", mw)
	if err != nil {
		return fmt.Errorf("creating maintenance window: %w", err)
	}
	return nil
}

func (c *Client) createMaintenanceWindowGen3(mw *MaintenanceWindow) error {
	setting := MaintenanceWindowSetting{
		Name:        mw.Name,
		Description: mw.Description,
		Enabled:     true,
		Type:        mw.Type,
		Suppression: mw.Suppression,
		GeneralProperties: MaintenanceGeneralProperties{
			DisableSyntheticMonitoring: false,
			MaintenanceType:            mw.Type,
			Recurrence:                 mw.Schedule.RecurrenceType,
			StartTime:                  mw.Schedule.Start,
			EndTime:                    mw.Schedule.End,
			ZoneID:                     mw.Schedule.ZoneID,
		},
	}

	settings := []SettingsObjectCreate{{
		SchemaID: "builtin:alerting.maintenance-window",
		Scope:    "environment",
		Value:    setting,
	}}

	_, err := c.post("/api/v2/settings/objects", settings)
	if err != nil {
		return fmt.Errorf("creating maintenance window via Settings 2.0: %w", err)
	}
	return nil
}

// ListMaintenanceNames returns the names of all existing maintenance windows.
func (c *Client) ListMaintenanceNames() ([]string, error) {
	if c.isGen3 {
		return c.listMaintenanceNamesGen3()
	}
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

func (c *Client) listMaintenanceNamesGen3() ([]string, error) {
	data, err := c.get("/api/v2/settings/objects?schemaIds=builtin:alerting.maintenance-window&scopes=environment&pageSize=500")
	if err != nil {
		return nil, fmt.Errorf("listing maintenance windows via Settings 2.0: %w", err)
	}
	var resp struct {
		Items []struct {
			Value struct {
				Name string `json:"name"`
			} `json:"value"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings objects: %w", err)
	}
	names := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		names[i] = item.Value.Name
	}
	return names, nil
}
