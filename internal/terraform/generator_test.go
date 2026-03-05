package terraform

import (
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

func TestGenerateProvider(t *testing.T) {
	result := GenerateProvider()
	if !strings.Contains(result, "dynatrace-oss/dynatrace") {
		t.Error("expected provider source in output")
	}
	if !strings.Contains(result, "DYNATRACE_ENV_URL") {
		t.Error("expected env var comment in output")
	}
}

func TestGenerateMetricEvents(t *testing.T) {
	events := []dynatrace.MetricEvent{
		{
			Summary:        "High CPU",
			Enabled:        true,
			EventType:      "CUSTOM_ALERT",
			MetricSelector: "builtin:host.cpu.user",
			MonitoringStrategy: dynatrace.MonitoringStrategy{
				Type:              "STATIC_THRESHOLD",
				AlertCondition:    "ABOVE",
				Threshold:         90,
				Samples:           5,
				ViolatingSamples:  3,
				DealertingSamples: 5,
			},
		},
	}

	result := GenerateMetricEvents(events)
	if !strings.Contains(result, "dynatrace_metric_events") {
		t.Error("expected resource type in output")
	}
	if !strings.Contains(result, "High CPU") {
		t.Error("expected summary in output")
	}
	if !strings.Contains(result, "STATIC_THRESHOLD") {
		t.Error("expected model type in output")
	}
}

func TestGenerateSLOs(t *testing.T) {
	slos := []dynatrace.SLO{
		{
			Name:             "API Availability",
			Enabled:          true,
			MetricExpression: "(100)*(good)/(total)",
			EvaluationType:   "AGGREGATE",
			Target:           99.9,
			Warning:          99.95,
			Timeframe:        "-1M",
		},
	}

	result := GenerateSLOs(slos)
	if !strings.Contains(result, "dynatrace_slo_v2") {
		t.Error("expected SLO resource type in output")
	}
	if !strings.Contains(result, "99.9") {
		t.Error("expected target in output")
	}
}

func TestSanitizeTFName(t *testing.T) {
	tests := map[string]string{
		"My Dashboard!":     "my_dashboard",
		"high-cpu-alert":    "high_cpu_alert",
		"123_starts_number": "r_123_starts_number",
		"":                  "resource",
	}
	for input, expected := range tests {
		result := sanitizeTFName(input)
		if result != expected {
			t.Errorf("sanitizeTFName(%q) = %q, want %q", input, result, expected)
		}
	}
}
