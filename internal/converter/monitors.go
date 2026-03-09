package converter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// ConvertMonitor converts a DataDog monitor to a Dynatrace metric event.
func ConvertMonitor(dd *datadog.Monitor) (*dynatrace.MetricEvent, error) {
	switch dd.Type {
	case "log alert":
		return convertLogAlertMonitor(dd)
	case "composite":
		return convertCompositeMonitor(dd)
	default: // "metric alert", "query alert", "service check", "event alert"
		return convertMetricMonitor(dd)
	}
}

func convertMetricMonitor(dd *datadog.Monitor) (*dynatrace.MetricEvent, error) {
	me := &dynatrace.MetricEvent{
		Summary:     dd.Name,
		Description: sanitizeDescription(dd.Message),
		Enabled:     true,
		EventType:   mapMonitorSeverity(dd.Type),
	}

	// Parse the monitor query to extract the metric selector
	metricSelector, alertCondition, threshold := parseMonitorQuery(dd)
	me.MetricSelector = metricSelector
	me.AlertCondition = alertCondition

	// Set threshold from monitor options
	if dd.Options.Thresholds != nil && dd.Options.Thresholds.Critical != nil {
		me.Threshold = *dd.Options.Thresholds.Critical
	} else if threshold != 0 {
		me.Threshold = threshold
	}

	me.MonitoringStrategy = dynatrace.MonitoringStrategy{
		Type:              "STATIC_THRESHOLD",
		Samples:           5,
		ViolatingSamples:  3,
		DealertingSamples: 5,
		AlertCondition:    alertCondition,
		Threshold:         me.Threshold,
	}

	// Map tags
	for _, tag := range dd.Tags {
		parts := strings.SplitN(tag, ":", 2)
		t := dynatrace.METag{Key: parts[0]}
		if len(parts) == 2 {
			t.Value = parts[1]
		}
		me.Tags = append(me.Tags, t)
	}

	if me.MetricSelector == "" {
		return nil, fmt.Errorf("could not translate monitor query: %s", dd.Query)
	}

	return me, nil
}

var logQueryPattern = regexp.MustCompile(`logs\("(.+?)"\)`)

func convertLogAlertMonitor(dd *datadog.Monitor) (*dynatrace.MetricEvent, error) {
	searchQuery, alertCondition, threshold := parseLogAlertQuery(dd.Query)
	dqlQuery := query.ToDQL(searchQuery, "log")

	// Use threshold from options if available
	if dd.Options.Thresholds != nil && dd.Options.Thresholds.Critical != nil {
		threshold = *dd.Options.Thresholds.Critical
	}

	desc := sanitizeDescription(dd.Message)
	desc += fmt.Sprintf("\n\n--- Migration Note ---\nThis was a DataDog log alert. The original search query was converted to DQL:\n\n%s\n\nConfigure a Dynatrace Log event or custom metric from logs to replicate this alert.", dqlQuery)
	if len(desc) > 1000 {
		desc = desc[:997] + "..."
	}

	me := &dynatrace.MetricEvent{
		Summary:        dd.Name,
		Description:    desc,
		Enabled:        true,
		EventType:      "CUSTOM_ALERT",
		MetricSelector: "builtin:host.availability",
		AlertCondition: alertCondition,
		Threshold:      threshold,
		MonitoringStrategy: dynatrace.MonitoringStrategy{
			Type:              "STATIC_THRESHOLD",
			Samples:           5,
			ViolatingSamples:  3,
			DealertingSamples: 5,
			AlertCondition:    alertCondition,
			Threshold:         threshold,
		},
	}

	for _, tag := range dd.Tags {
		parts := strings.SplitN(tag, ":", 2)
		t := dynatrace.METag{Key: parts[0]}
		if len(parts) == 2 {
			t.Value = parts[1]
		}
		me.Tags = append(me.Tags, t)
	}

	return me, nil
}

func parseLogAlertQuery(q string) (searchQuery string, alertCondition string, threshold float64) {
	alertCondition = "ABOVE"

	// Extract search query from logs("...") pattern
	if m := logQueryPattern.FindStringSubmatch(q); len(m) > 1 {
		searchQuery = m[1]
	}

	// Extract threshold and condition from comparison operator
	for _, op := range []string{" >= ", " > ", " <= ", " < ", " == "} {
		if idx := strings.LastIndex(q, op); idx > 0 {
			alertCondition = query.MapAlertCondition(strings.TrimSpace(op))
			threshStr := strings.TrimSpace(q[idx+len(op):])
			if v, err := strconv.ParseFloat(threshStr, 64); err == nil {
				threshold = v
			}
			break
		}
	}

	return
}

var compositeIDPattern = regexp.MustCompile(`\b(\d+)\b`)

func convertCompositeMonitor(dd *datadog.Monitor) (*dynatrace.MetricEvent, error) {
	// Extract referenced monitor IDs from composite expression
	matches := compositeIDPattern.FindAllString(dd.Query, -1)
	var ids []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m] {
			ids = append(ids, m)
			seen[m] = true
		}
	}

	desc := sanitizeDescription(dd.Message)
	desc += fmt.Sprintf("\n\n--- Migration Note ---\nThis was a DataDog composite monitor referencing monitors: %s.\nOriginal expression: %s\nDynatrace does not support composite metric events. Create individual metric events and use Dynatrace alerting profiles to combine them.",
		strings.Join(ids, ", "), dd.Query)
	if len(desc) > 1000 {
		desc = desc[:997] + "..."
	}

	me := &dynatrace.MetricEvent{
		Summary:        dd.Name,
		Description:    desc,
		Enabled:        true,
		EventType:      "CUSTOM_ALERT",
		MetricSelector: "builtin:host.availability",
		AlertCondition: "ABOVE",
		Threshold:      0,
		MonitoringStrategy: dynatrace.MonitoringStrategy{
			Type:              "STATIC_THRESHOLD",
			Samples:           5,
			ViolatingSamples:  3,
			DealertingSamples: 5,
			AlertCondition:    "ABOVE",
			Threshold:         0,
		},
	}

	for _, tag := range dd.Tags {
		parts := strings.SplitN(tag, ":", 2)
		t := dynatrace.METag{Key: parts[0]}
		if len(parts) == 2 {
			t.Value = parts[1]
		}
		me.Tags = append(me.Tags, t)
	}

	return me, nil
}

func parseMonitorQuery(dd *datadog.Monitor) (metricSelector string, alertCondition string, threshold float64) {
	alertCondition = "ABOVE" // default

	// DD monitor queries have the format:
	// "metric_type(last_5m):aggregation:metric{filters} by {groupby} > threshold"

	q := dd.Query

	// Extract comparison and threshold
	for _, op := range []string{" >= ", " > ", " <= ", " < ", " == "} {
		if idx := strings.LastIndex(q, op); idx > 0 {
			q = q[:idx]
			op = strings.TrimSpace(op)
			alertCondition = query.MapAlertCondition(op)
			break
		}
	}

	// Strip the outer function wrapper like "avg(last_5m):"
	if idx := strings.Index(q, "):"); idx > 0 {
		q = q[idx+2:]
	}

	// Parse the remaining metric query
	logging.Debug("parsing monitor query: %s", q)
	parsed, err := query.Parse(q)
	if err != nil {
		logging.Warn("query parse failed, falling back to raw string: %s", q)
		return q, alertCondition, threshold
	}

	metricSelector = query.ToMetricSelector(parsed)
	return metricSelector, alertCondition, threshold
}

func mapMonitorSeverity(ddType string) string {
	switch ddType {
	case "metric alert", "query alert":
		return "CUSTOM_ALERT"
	case "service check":
		return "ERROR"
	case "event alert":
		return "INFO"
	case "log alert":
		return "CUSTOM_ALERT"
	case "composite":
		return "CUSTOM_ALERT"
	default:
		return "CUSTOM_ALERT"
	}
}

func sanitizeDescription(msg string) string {
	// Remove DD-specific notification handles like @slack-channel, @pagerduty-service
	lines := strings.Split(msg, "\n")
	var clean []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@") {
			continue
		}
		clean = append(clean, line)
	}
	result := strings.Join(clean, "\n")
	// Truncate if too long for DT
	if len(result) > 1000 {
		result = result[:997] + "..."
	}
	return strings.TrimSpace(result)
}
