package dynatrace

import "fmt"

// CreateSLO creates an SLO in Dynatrace.
func (c *Client) CreateSLO(s *SLO) error {
	_, err := c.post("/api/v2/slo", s)
	if err != nil {
		return fmt.Errorf("creating SLO: %w", err)
	}
	return nil
}
