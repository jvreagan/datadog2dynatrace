package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateNotification creates a notification integration in Dynatrace.
func (c *Client) CreateNotification(n *NotificationIntegration) error {
	if c.isGen3 {
		return c.createNotificationGen3(n)
	}
	_, err := c.post("/api/config/v1/notifications", n)
	if err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}
	return nil
}

func (c *Client) createNotificationGen3(n *NotificationIntegration) error {
	setting := NotificationSetting{
		Name:   n.Name,
		Type:   n.Type,
		Active: n.Active,
		Config: n.Config,
	}

	settings := []SettingsObjectCreate{{
		SchemaID: "builtin:problem.notifications",
		Scope:    "environment",
		Value:    setting,
	}}

	_, err := c.post("/api/v2/settings/objects", settings)
	if err != nil {
		return fmt.Errorf("creating notification via Settings 2.0: %w", err)
	}
	return nil
}

// ListNotificationNames returns the names of all existing notifications.
func (c *Client) ListNotificationNames() ([]string, error) {
	if c.isGen3 {
		return c.listNotificationNamesGen3()
	}
	data, err := c.get("/api/config/v1/notifications")
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	var resp struct {
		Values []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing notification list: %w", err)
	}
	names := make([]string, len(resp.Values))
	for i, v := range resp.Values {
		names[i] = v.Name
	}
	return names, nil
}

func (c *Client) listNotificationNamesGen3() ([]string, error) {
	data, err := c.get("/api/v2/settings/objects?schemaIds=builtin:problem.notifications&scopes=environment&pageSize=500")
	if err != nil {
		return nil, fmt.Errorf("listing notifications via Settings 2.0: %w", err)
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
