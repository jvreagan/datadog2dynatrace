package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateNotification creates a notification integration in Dynatrace.
func (c *Client) CreateNotification(n *NotificationIntegration) error {
	_, err := c.post("/api/config/v1/notifications", n)
	if err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}
	return nil
}

// ListNotificationNames returns the names of all existing notifications.
func (c *Client) ListNotificationNames() ([]string, error) {
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
