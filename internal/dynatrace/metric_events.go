package dynatrace

import (
	"encoding/json"
	"fmt"
)

// CreateMetricEvent creates a metric event (alert) in Dynatrace.
func (c *Client) CreateMetricEvent(me *MetricEvent) error {
	if c.isGen3 {
		return c.createMetricEventGen3(me)
	}
	_, err := c.post("/api/config/v1/anomalyDetection/metricEvents", me)
	if err != nil {
		return fmt.Errorf("creating metric event: %w", err)
	}
	return nil
}

func (c *Client) createMetricEventGen3(me *MetricEvent) error {
	detector := DavisAnomalyDetector{
		Title:       me.Summary,
		Description: me.Description,
		Enabled:     me.Enabled,
		EventTemplate: DavisEventTemplate{
			Title:       me.Summary,
			Description: me.Description,
			EventType:   mapEventType(me.EventType),
			DavisMerge:  true,
		},
		Analyzer: DavisAnalyzerConfig{
			Input: []DavisAnalyzerInput{{
				Name: "input",
				AnalyzerDef: DavisAnalyzerDefinition{
					Type:              mapModelType(me.MonitoringStrategy.Type),
					Threshold:         me.MonitoringStrategy.Threshold,
					AlertCondition:    me.MonitoringStrategy.AlertCondition,
					Samples:           me.MonitoringStrategy.Samples,
					ViolatingSamples:  me.MonitoringStrategy.ViolatingSamples,
					DealertingSamples: me.MonitoringStrategy.DealertingSamples,
				},
			}},
		},
		QueryDefinition: DavisQueryDefinition{
			Type:           "METRIC_SELECTOR",
			MetricSelector: me.MetricSelector,
		},
	}

	settings := []SettingsObjectCreate{{
		SchemaID: "builtin:davis.anomaly-detectors",
		Scope:    "environment",
		Value:    detector,
	}}

	_, err := c.post("/api/v2/settings/objects", settings)
	if err != nil {
		return fmt.Errorf("creating metric event via Settings 2.0: %w", err)
	}
	return nil
}

func mapEventType(et string) string {
	switch et {
	case "CUSTOM_ALERT":
		return "CUSTOM_ALERT"
	case "ERROR":
		return "ERROR_EVENT"
	case "INFO":
		return "CUSTOM_INFO"
	default:
		return "CUSTOM_ALERT"
	}
}

func mapModelType(mt string) string {
	switch mt {
	case "STATIC_THRESHOLD":
		return "STATIC"
	case "AUTO_ADAPTIVE":
		return "AUTO_ADAPTIVE"
	default:
		return "STATIC"
	}
}

// ListMetricEventSummaries returns the summaries of all existing metric events.
func (c *Client) ListMetricEventSummaries() ([]string, error) {
	if c.isGen3 {
		return c.listMetricEventSummariesGen3()
	}
	data, err := c.get("/api/config/v1/anomalyDetection/metricEvents")
	if err != nil {
		return nil, fmt.Errorf("listing metric events: %w", err)
	}
	var resp struct {
		Values []struct {
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing metric event list: %w", err)
	}
	names := make([]string, len(resp.Values))
	for i, v := range resp.Values {
		names[i] = v.Name
	}
	return names, nil
}

func (c *Client) listMetricEventSummariesGen3() ([]string, error) {
	data, err := c.get("/api/v2/settings/objects?schemaIds=builtin:davis.anomaly-detectors&scopes=environment&pageSize=500")
	if err != nil {
		return nil, fmt.Errorf("listing metric events via Settings 2.0: %w", err)
	}
	var resp struct {
		Items []struct {
			Value struct {
				Title string `json:"title"`
			} `json:"value"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings objects: %w", err)
	}
	names := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		names[i] = item.Value.Title
	}
	return names, nil
}
