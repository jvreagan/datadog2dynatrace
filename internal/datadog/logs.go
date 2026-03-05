package datadog

import (
	"encoding/json"
	"fmt"
)

// GetLogPipelines retrieves all log pipelines.
func (c *Client) GetLogPipelines() ([]LogPipeline, error) {
	data, err := c.get("/api/v1/logs/config/pipelines")
	if err != nil {
		return nil, err
	}
	var pipelines []LogPipeline
	if err := json.Unmarshal(data, &pipelines); err != nil {
		return nil, fmt.Errorf("parsing log pipelines: %w", err)
	}
	return pipelines, nil
}
