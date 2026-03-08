package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateSLO creates an SLO in Dynatrace.
func (c *Client) CreateSLO(s *SLO) error {
	_, err := c.post("/api/v2/slo", s)
	if err != nil {
		return fmt.Errorf("creating SLO: %w", err)
	}
	return nil
}

// ListSLONames returns the names of all existing SLOs.
func (c *Client) ListSLONames() ([]string, error) {
	data, err := c.get("/api/v2/slo")
	if err != nil {
		return nil, fmt.Errorf("listing SLOs: %w", err)
	}
	var resp struct {
		SLOs []struct {
			Name string `json:"name"`
		} `json:"slo"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing SLO list: %w", err)
	}
	names := make([]string, len(resp.SLOs))
	for i, s := range resp.SLOs {
		names[i] = s.Name
	}
	return names, nil
}
