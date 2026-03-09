package datadog

import (
	"encoding/json"
	"fmt"
)

// GetNotificationChannels retrieves all notification integrations.
// It fetches from webhooks, Slack, and PagerDuty endpoints.
// Each endpoint may return an error if the integration is not configured;
// these are handled gracefully with empty results.
func (c *Client) GetNotificationChannels() ([]NotificationChannel, error) {
	var channels []NotificationChannel
	var id int64

	// Get webhooks
	if data, err := c.get("/api/v1/integration/webhooks/configuration/webhooks"); err == nil {
		var webhooks []struct {
			Name          string `json:"name"`
			URL           string `json:"url"`
			CustomHeaders string `json:"custom_headers,omitempty"`
			Payload       string `json:"payload,omitempty"`
		}
		if err := json.Unmarshal(data, &webhooks); err != nil {
			return nil, fmt.Errorf("parsing webhooks: %w", err)
		}
		for _, wh := range webhooks {
			channels = append(channels, NotificationChannel{
				ID:   id,
				Name: wh.Name,
				Type: "webhook",
				Config: map[string]interface{}{
					"url":     wh.URL,
					"payload": wh.Payload,
				},
			})
			id++
		}
	}

	// Get Slack channels
	if data, err := c.get("/api/v1/integration/slack/configuration/accounts"); err == nil {
		var accounts []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &accounts); err == nil {
			for _, acc := range accounts {
				channels = append(channels, NotificationChannel{
					ID:   id,
					Name: acc.Name,
					Type: "slack",
					Config: map[string]interface{}{
						"account": acc.Name,
					},
				})
				id++
			}
		}
	}

	// Get PagerDuty services
	if data, err := c.get("/api/v1/integration/pagerduty/configuration/services"); err == nil {
		var services []struct {
			ServiceName string `json:"service_name"`
			ServiceKey  string `json:"service_key"`
		}
		if err := json.Unmarshal(data, &services); err == nil {
			for _, svc := range services {
				channels = append(channels, NotificationChannel{
					ID:   id,
					Name: svc.ServiceName,
					Type: "pagerduty",
					Config: map[string]interface{}{
						"service_name": svc.ServiceName,
						"service_key":  svc.ServiceKey,
					},
				})
				id++
			}
		}
	}

	return channels, nil
}
