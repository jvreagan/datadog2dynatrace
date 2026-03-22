package dynatrace

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// staticThresholdAnalyzer is the fully-qualified name of the DT static threshold analyzer.
const staticThresholdAnalyzer = "dt.statistics.ui.anomaly_detection.StaticThresholdAnomalyDetectionAnalyzer"

// splitByPattern matches DT metric selectors of the form:
//
//	metric:splitBy("dim1","dim2"):agg
var splitByPattern = regexp.MustCompile(`^(.*):splitBy\(([^)]*)\)(?::(\w+))?$`)

// metricSelectorToDQL converts a Dynatrace v2 metric selector to a DQL timeseries expression.
// For selectors it cannot parse it wraps the expression in a best-effort timeseries call.
func metricSelectorToDQL(selector string) string {
	m := splitByPattern.FindStringSubmatch(selector)
	if m == nil {
		return "timeseries avg(" + dqlMetricName(selector) + ")"
	}
	metric := dqlMetricName(strings.TrimSpace(m[1]))
	dims := strings.ReplaceAll(m[2], `"`, "")
	dims = strings.TrimSpace(dims)
	agg := "avg"
	if m[3] != "" {
		agg = m[3]
	}
	if dims == "" {
		return fmt.Sprintf("timeseries %s(%s)", agg, metric)
	}
	return fmt.Sprintf("timeseries %s(%s), by:{%s}", agg, metric, dims)
}

// dqlMetricName converts a metric selector key to its DQL equivalent.
// In DQL, builtin metrics use the "dt." prefix instead of "builtin:".
// ext: metrics replace the colon separator with a dot.
func dqlMetricName(metric string) string {
	if strings.HasPrefix(metric, "builtin:") {
		return "dt." + metric[len("builtin:"):]
	}
	if strings.HasPrefix(metric, "ext:") {
		return "ext." + metric[len("ext:"):]
	}
	return metric
}

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
	dql := metricSelectorToDQL(me.MetricSelector)
	s := me.MonitoringStrategy

	detector := DavisAnomalyDetector{
		Title:       me.Summary,
		Description: me.Description,
		Enabled:     me.Enabled,
		Source:      "Rest-API",
		ExecutionSettings: DavisExecutionSettings{
			Actor: nil,
		},
		Analyzer: DavisAnalyzerInput{
			Name: staticThresholdAnalyzer,
			Input: []DavisAnalyzerField{
				{Key: "query", Value: dql},
				{Key: "threshold", Value: fmt.Sprintf("%g", s.Threshold)},
				{Key: "alertCondition", Value: s.AlertCondition},
				{Key: "alertOnMissingData", Value: "false"},
				{Key: "violatingSamples", Value: fmt.Sprintf("%d", s.ViolatingSamples)},
				{Key: "slidingWindow", Value: fmt.Sprintf("%d", s.Samples)},
				{Key: "dealertingSamples", Value: fmt.Sprintf("%d", s.DealertingSamples)},
			},
		},
		EventTemplate: DavisEventTemplate{
			Properties: []DavisEventProperty{
				{Key: "event.name", Value: me.Summary},
				{Key: "event.type", Value: mapEventType(me.EventType)},
			},
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
	case "ERROR":
		return "ERROR_EVENT"
	case "INFO":
		return "CUSTOM_INFO"
	default:
		return "CUSTOM_ALERT"
	}
}

// ListMetricEventSummaries returns the summaries of all existing metric events.
func (c *Client) ListMetricEventSummaries() ([]string, error) {
	resources, err := c.ListMetricEventsWithIDs()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(resources))
	for i, r := range resources {
		names[i] = r.Name
	}
	return names, nil
}

// ListMetricEventsWithIDs returns existing metric events with their IDs and names.
func (c *Client) ListMetricEventsWithIDs() ([]NamedResource, error) {
	if c.isGen3 {
		return c.listMetricEventsWithIDsGen3()
	}
	data, err := c.get("/api/config/v1/anomalyDetection/metricEvents")
	if err != nil {
		return nil, fmt.Errorf("listing metric events: %w", err)
	}
	var resp struct {
		Values []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing metric event list: %w", err)
	}
	resources := make([]NamedResource, len(resp.Values))
	for i, v := range resp.Values {
		resources[i] = NamedResource{ID: v.ID, Name: v.Name}
	}
	return resources, nil
}

func (c *Client) listMetricEventsWithIDsGen3() ([]NamedResource, error) {
	data, err := c.get("/api/v2/settings/objects?schemaIds=builtin:davis.anomaly-detectors&scopes=environment&pageSize=500")
	if err != nil {
		return nil, fmt.Errorf("listing metric events via Settings 2.0: %w", err)
	}
	var resp struct {
		Items []struct {
			ObjectID string `json:"objectId"`
			Value    struct {
				Title string `json:"title"`
			} `json:"value"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings objects: %w", err)
	}
	resources := make([]NamedResource, len(resp.Items))
	for i, item := range resp.Items {
		resources[i] = NamedResource{ID: item.ObjectID, Name: item.Value.Title}
	}
	return resources, nil
}

// DeleteMetricEvent deletes a metric event by ID.
func (c *Client) DeleteMetricEvent(id string) error {
	if c.isGen3 {
		return c.delete("/api/v2/settings/objects/" + id)
	}
	return c.delete("/api/config/v1/anomalyDetection/metricEvents/" + id)
}
