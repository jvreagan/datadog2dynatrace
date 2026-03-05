package dynatrace

import "fmt"

// CreateNotification creates a notification integration in Dynatrace.
func (c *Client) CreateNotification(n *NotificationIntegration) error {
	_, err := c.post("/api/config/v1/notifications", n)
	if err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}
	return nil
}
