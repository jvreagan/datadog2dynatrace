package datadog

import (
	"encoding/json"
	"fmt"
)

// GetMonitors retrieves all monitors.
func (c *Client) GetMonitors() ([]Monitor, error) {
	data, err := c.get("/api/v1/monitor")
	if err != nil {
		return nil, err
	}
	var monitors []Monitor
	if err := json.Unmarshal(data, &monitors); err != nil {
		return nil, fmt.Errorf("parsing monitors: %w", err)
	}
	return monitors, nil
}

// GetMonitor retrieves a single monitor by ID.
func (c *Client) GetMonitor(id int64) (*Monitor, error) {
	data, err := c.get(fmt.Sprintf("/api/v1/monitor/%d", id))
	if err != nil {
		return nil, err
	}
	var m Monitor
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing monitor: %w", err)
	}
	return &m, nil
}
