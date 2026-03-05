package datadog

import (
	"encoding/json"
	"fmt"
)

// GetSLOs retrieves all SLOs.
func (c *Client) GetSLOs() ([]SLO, error) {
	data, err := c.get("/api/v1/slo")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []SLO `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing SLOs: %w", err)
	}
	return resp.Data, nil
}
