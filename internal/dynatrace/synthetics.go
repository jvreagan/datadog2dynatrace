package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateSyntheticMonitor creates a synthetic monitor in Dynatrace.
func (c *Client) CreateSyntheticMonitor(sm *SyntheticMonitor) error {
	_, err := c.post("/api/v1/synthetic/monitors", sm)
	if err != nil {
		return fmt.Errorf("creating synthetic monitor: %w", err)
	}
	return nil
}

// ListSyntheticNames returns the names of all existing synthetic monitors.
func (c *Client) ListSyntheticNames() ([]string, error) {
	data, err := c.get("/api/v1/synthetic/monitors")
	if err != nil {
		return nil, fmt.Errorf("listing synthetic monitors: %w", err)
	}
	var resp struct {
		Monitors []struct {
			Name string `json:"name"`
		} `json:"monitors"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing synthetic monitor list: %w", err)
	}
	names := make([]string, len(resp.Monitors))
	for i, m := range resp.Monitors {
		names[i] = m.Name
	}
	return names, nil
}
