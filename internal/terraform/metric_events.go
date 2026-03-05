package terraform

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// GenerateMetricEvents generates Terraform HCL for Dynatrace metric events.
func GenerateMetricEvents(events []dynatrace.MetricEvent) string {
	var sb strings.Builder
	sb.WriteString("# Metric Events (Alerts) - migrated from DataDog Monitors\n\n")

	for i, me := range events {
		name := sanitizeTFName(me.Summary)
		sb.WriteString(fmt.Sprintf("resource \"dynatrace_metric_events\" \"%s\" {\n", uniqueName(name, i)))
		sb.WriteString(fmt.Sprintf("  enabled          = %t\n", me.Enabled))
		sb.WriteString(fmt.Sprintf("  summary          = %q\n", me.Summary))
		sb.WriteString(fmt.Sprintf("  event_type       = %q\n", me.EventType))
		sb.WriteString(fmt.Sprintf("  metric_selector  = %q\n", me.MetricSelector))

		if me.Description != "" {
			sb.WriteString(fmt.Sprintf("  description      = %q\n", me.Description))
		}

		sb.WriteString("\n  model_properties {\n")
		sb.WriteString(fmt.Sprintf("    type               = %q\n", me.MonitoringStrategy.Type))
		sb.WriteString(fmt.Sprintf("    alert_condition     = %q\n", me.MonitoringStrategy.AlertCondition))
		sb.WriteString(fmt.Sprintf("    threshold          = %g\n", me.MonitoringStrategy.Threshold))
		sb.WriteString(fmt.Sprintf("    samples            = %d\n", me.MonitoringStrategy.Samples))
		sb.WriteString(fmt.Sprintf("    violating_samples  = %d\n", me.MonitoringStrategy.ViolatingSamples))
		sb.WriteString(fmt.Sprintf("    dealerting_samples = %d\n", me.MonitoringStrategy.DealertingSamples))
		sb.WriteString("  }\n")

		sb.WriteString("}\n\n")
	}

	return sb.String()
}
