package datadog

import (
	"encoding/json"
	"fmt"
)

// GetSynthetics retrieves all synthetic tests.
func (c *Client) GetSynthetics() ([]SyntheticTest, error) {
	data, err := c.get("/api/v1/synthetics/tests")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Tests []SyntheticTest `json:"tests"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing synthetic tests: %w", err)
	}
	return resp.Tests, nil
}
