package dynatrace

import "fmt"

// CreateSyntheticMonitor creates a synthetic monitor in Dynatrace.
func (c *Client) CreateSyntheticMonitor(sm *SyntheticMonitor) error {
	_, err := c.post("/api/v1/synthetic/monitors", sm)
	if err != nil {
		return fmt.Errorf("creating synthetic monitor: %w", err)
	}
	return nil
}
