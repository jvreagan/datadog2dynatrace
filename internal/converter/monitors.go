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
	case "metric alert", "query alert":
		return convertMetricMonitor(dd)
	case "log alert":
		return convertLogAlertMonitor(dd)
	case "composite":
		return convertCompositeMonitor(dd)
	case "service check":
		return convertUnsupportedMonitor(dd, "service check",
			"DataDog service checks poll agent integrations and report UP/DOWN/WARN status. "+
				"The check status cannot be expressed as a Dynatrace metric selector. "+
				"Use Dynatrace host/service availability alerting or create a custom Davis anomaly detector.")
	case "event alert":
		return convertUnsupportedMonitor(dd, "event alert",
			"DataDog event alerts monitor the DataDog event stream. "+
				"Dynatrace does not have a direct equivalent. "+
				"Consider using Dynatrace problem alerting or log monitoring rules.")
	case "apm alert":
		return convertUnsupportedMonitor(dd, "apm alert",
			"DataDog APM alerts monitor trace/span metrics such as latency and error rate. "+
				"Configure a Dynatrace service anomaly detection rule targeting the equivalent service.")
	case "rum alert":
		return convertUnsupportedMonitor(dd, "rum alert",
			"DataDog RUM alerts monitor real user monitoring data. "+
				"Configure a Dynatrace RUM threshold or session anomaly detection rule.")
	case "process alert":
		return convertUnsupportedMonitor(dd, "process alert",
			"DataDog process monitors track running processes via the Live Process agent. "+
				"Use Dynatrace process availability monitoring or create a custom metric alert.")
	case "network alert":
		return convertUnsupportedMonitor(dd, "network alert",
			"DataDog network alerts monitor NPM (Network Performance Monitoring) data. "+
				"Use Dynatrace network monitoring or create a custom metric alert for equivalent host network metrics.")
	default:
		return convertUnsupportedMonitor(dd, dd.Type,
			fmt.Sprintf("Monitor type %q has no direct Dynatrace equivalent. Manual configuration is required.", dd.Type))
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

// convertUnsupportedMonitor creates a stub metric event for monitor types that cannot
// be directly translated to a Dynatrace metric selector. The original query and a
// human-readable migration note are embedded in the description.
func convertUnsupportedMonitor(dd *datadog.Monitor, typeName, migrationNote string) (*dynatrace.MetricEvent, error) {
	desc := sanitizeDescription(dd.Message)
	note := fmt.Sprintf("\n\n--- Migration Note ---\nThis was a DataDog %s monitor.\nOriginal query: %s\n\n%s",
		typeName, dd.Query, migrationNote)
	combined := desc + note
	if len(combined) > 1000 {
		combined = combined[:997] + "..."
	}

	me := &dynatrace.MetricEvent{
		Summary:        dd.Name,
		Description:    combined,
		Enabled:        true,
		EventType:      mapMonitorSeverity(dd.Type),
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

	// Preprocess formula expressions (e.g., "100 - cpu.idle", "(1 - disk.free)")
	if preprocessed, ok := preprocessFormulaQuery(q); ok {
		logging.Debug("formula query preprocessed: %q -> %q", q, preprocessed)
		q = preprocessed
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

// preprocessFormulaQuery handles common DD formula patterns by substituting complement
// metrics. Returns (processed, true) if a substitution was made.
//
// Handles:
//   - "100 - avg:metric.idle{filters} by {groupby}" → "avg:metric.user{filters} by {groupby}"
//   - "(1 - avg:metric.free{filters}) by {groupby}" → "avg:metric.in_use{filters} by {groupby}"
//   - "forecast(<formula>, ...)" → strips forecast wrapper and applies above rules
func preprocessFormulaQuery(q string) (string, bool) {
	q = strings.TrimSpace(q)

	// Pattern: "forecast(<formula>, ...)" — DD predictive monitor; strip the wrapper
	// and preprocess the underlying metric formula.
	if strings.HasPrefix(q, "forecast(") {
		inner := extractFirstTopLevelArg(q[len("forecast("):])
		if inner != "" {
			if preprocessed, ok := preprocessFormulaQuery(inner); ok {
				return preprocessed, true
			}
			// No substitution but we still drop the forecast wrapper
			return inner, true
		}
	}

	// Pattern: "100 - <inner_query>"
	if strings.HasPrefix(q, "100 - ") || strings.HasPrefix(q, "100 -") {
		dashIdx := strings.Index(q, "-")
		inner := strings.TrimSpace(q[dashIdx+1:])
		if subst := complementSubstitute(inner); subst != "" {
			return subst, true
		}
	}

	// Pattern: "(1 - <inner_query>) [by {groupby}]"
	if strings.HasPrefix(q, "(1 - ") || strings.HasPrefix(q, "(1-") {
		closeIdx := strings.LastIndex(q, ")")
		if closeIdx > 0 {
			dashIdx := strings.Index(q, "-")
			inner := strings.TrimSpace(q[dashIdx+1 : closeIdx])
			byPart := strings.TrimSpace(q[closeIdx+1:])
			if subst := complementSubstitute(inner); subst != "" {
				// Only append byPart if it looks like a "by {groupby}" clause
				if byPart != "" && (strings.HasPrefix(byPart, "by {") || strings.HasPrefix(byPart, "by{")) {
					return subst + " " + byPart, true
				}
				return subst, true
			}
		}
	}

	return q, false
}

// extractFirstTopLevelArg returns the first comma-separated argument at depth 0
// from a function call's argument list (the string after the opening paren).
// Depth is tracked for both () and {} brackets.
func extractFirstTopLevelArg(s string) string {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(', '{':
			depth++
		case ')', '}':
			if depth == 0 {
				// Reached the closing paren of the outer function — return remainder
				return strings.TrimSpace(s[:i])
			}
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return strings.TrimSpace(s)
}

// complementSubstitute parses an inner metric query, looks up its complement metric,
// and returns a new query string with the complement substituted in.
func complementSubstitute(inner string) string {
	pq, err := query.Parse(inner)
	if err != nil {
		return ""
	}
	complement := query.ComplementMetric(pq.Metric)
	if complement == "" {
		return ""
	}
	return strings.Replace(inner, pq.Metric, complement, 1)
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
