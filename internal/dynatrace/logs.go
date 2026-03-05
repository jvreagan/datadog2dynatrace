package dynatrace

import "fmt"

// CreateLogProcessingRule creates a log processing rule in Dynatrace.
func (c *Client) CreateLogProcessingRule(rule *LogProcessingRule) error {
	_, err := c.post("/api/v2/logs/processing/rules", rule)
	if err != nil {
		return fmt.Errorf("creating log processing rule: %w", err)
	}
	return nil
}
