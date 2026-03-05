package datadog

import (
	"encoding/json"
	"fmt"
)

// GetDowntimes retrieves all downtimes.
func (c *Client) GetDowntimes() ([]Downtime, error) {
	data, err := c.get("/api/v1/downtime")
	if err != nil {
		return nil, err
	}
	var downtimes []Downtime
	if err := json.Unmarshal(data, &downtimes); err != nil {
		return nil, fmt.Errorf("parsing downtimes: %w", err)
	}
	return downtimes, nil
}
