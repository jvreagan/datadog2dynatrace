package datadog

import (
	"encoding/json"
	"fmt"
)

// GetNotificationChannels retrieves all notification integrations.
// Note: DataDog doesn't have a single API for all notification types.
// This uses the webhooks integration endpoint as the primary source.
func (c *Client) GetNotificationChannels() ([]NotificationChannel, error) {
	// Get webhooks
	data, err := c.get("/api/v1/integration/webhooks/configuration/webhooks")
	if err != nil {
		// Webhooks may not be configured, return empty
		return []NotificationChannel{}, nil
	}
	var webhooks []struct {
		Name       string `json:"name"`
		URL        string `json:"url"`
		CustomHeaders string `json:"custom_headers,omitempty"`
		Payload    string `json:"payload,omitempty"`
	}
	if err := json.Unmarshal(data, &webhooks); err != nil {
		return nil, fmt.Errorf("parsing webhooks: %w", err)
	}

	var channels []NotificationChannel
	for i, wh := range webhooks {
		channels = append(channels, NotificationChannel{
			ID:   int64(i),
			Name: wh.Name,
			Type: "webhook",
			Config: map[string]interface{}{
				"url":     wh.URL,
				"payload": wh.Payload,
			},
		})
	}
	return channels, nil
}
