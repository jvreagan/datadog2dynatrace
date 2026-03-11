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

func TestGenerateDashboards(t *testing.T) {
	dashboards := []dynatrace.Dashboard{
		{
			DashboardMetadata: dynatrace.DashboardMetadata{Name: "Test Dashboard"},
			Tiles: []dynatrace.Tile{
				{Name: "CPU", TileType: "DATA_EXPLORER"},
			},
		},
	}
	result := GenerateDashboards(dashboards)
	if !strings.Contains(result, "dynatrace_json_dashboard") {
		t.Error("expected dynatrace_json_dashboard resource type")
	}
	if !strings.Contains(result, "Test Dashboard") {
		t.Error("expected dashboard name in output")
	}
}

func TestGenerateSynthetics(t *testing.T) {
	monitors := []dynatrace.SyntheticMonitor{
		{
			Name:    "Health Check",
			Type:    "HTTP",
			Enabled: true,
		},
	}
	result := GenerateSynthetics(monitors)
	if !strings.Contains(result, "dynatrace_http_monitor") {
		t.Error("expected dynatrace_http_monitor resource type")
	}
	if !strings.Contains(result, "Health Check") {
		t.Error("expected monitor name in output")
	}
}

func TestGenerateSyntheticsBrowser(t *testing.T) {
	monitors := []dynatrace.SyntheticMonitor{
		{
			Name:    "Browser Test",
			Type:    "BROWSER",
			Enabled: true,
		},
	}
	result := GenerateSynthetics(monitors)
	if !strings.Contains(result, "dynatrace_browser_monitor") {
		t.Error("expected dynatrace_browser_monitor resource type")
	}
}

func TestGenerateLogProcessing(t *testing.T) {
	rules := []dynatrace.LogProcessingRule{
		{
			Name:    "Parse Nginx",
			Enabled: true,
			Query:   "process_group_instance:nginx",
		},
	}
	result := GenerateLogProcessing(rules)
	if !strings.Contains(result, "dynatrace_log_processing") {
		t.Error("expected dynatrace_log_processing resource type")
	}
	if !strings.Contains(result, "Parse Nginx") {
		t.Error("expected rule name in output")
	}
	if !strings.Contains(result, "process_group_instance:nginx") {
		t.Error("expected query in output")
	}
}

func TestGenerateMaintenance(t *testing.T) {
	windows := []dynatrace.MaintenanceWindow{
		{
			Name:        "Weekly Maintenance",
			Type:        "PLANNED",
			Suppression: "DETECT_PROBLEMS_DONT_ALERT",
			Schedule: dynatrace.MaintenanceSchedule{
				RecurrenceType: "WEEKLY",
				Start:          "2024-01-01 00:00",
				End:            "2024-01-01 02:00",
				ZoneID:         "UTC",
			},
		},
	}
	result := GenerateMaintenance(windows)
	if !strings.Contains(result, "dynatrace_maintenance") {
		t.Error("expected dynatrace_maintenance resource type")
	}
	if !strings.Contains(result, "schedule") {
		t.Error("expected schedule block in output")
	}
}

func TestGenerateNotifications(t *testing.T) {
	notifications := []dynatrace.NotificationIntegration{
		{
			Name:   "Slack Alert",
			Type:   "SLACK",
			Active: true,
			Config: map[string]interface{}{"url": "https://hooks.slack.com/test", "channel": "#alerts"},
		},
		{
			Name:   "Webhook",
			Type:   "WEBHOOK",
			Active: true,
			Config: map[string]interface{}{"url": "https://example.com/hook"},
		},
		{
			Name:   "PD Oncall",
			Type:   "PAGER_DUTY",
			Active: true,
			Config: map[string]interface{}{"account": "Production Oncall", "integrationKey": "pd-key-123"},
		},
		{
			Name:   "OpsGenie Alert",
			Type:   "OPS_GENIE",
			Active: true,
			Config: map[string]interface{}{"apiKey": "og-key-456"},
		},
		{
			Name:   "VictorOps Alert",
			Type:   "VICTOR_OPS",
			Active: true,
			Config: map[string]interface{}{"apiKey": "vo-key-789"},
		},
	}
	result := GenerateNotifications(notifications)
	if !strings.Contains(result, "dynatrace_slack_notification") {
		t.Error("expected dynatrace_slack_notification resource type")
	}
	if !strings.Contains(result, "dynatrace_webhook_notification") {
		t.Error("expected dynatrace_webhook_notification resource type")
	}
	if !strings.Contains(result, "dynatrace_pagerduty_notification") {
		t.Error("expected dynatrace_pagerduty_notification resource type")
	}
	if !strings.Contains(result, "integration_key") {
		t.Error("expected integration_key in PagerDuty resource")
	}
	if !strings.Contains(result, "pd-key-123") {
		t.Error("expected PagerDuty integration key value in output")
	}
	if !strings.Contains(result, "Production Oncall") {
		t.Error("expected PagerDuty account name in output")
	}
	if !strings.Contains(result, "dynatrace_opsgenie_notification") {
		t.Error("expected dynatrace_opsgenie_notification resource type")
	}
	if !strings.Contains(result, "og-key-456") {
		t.Error("expected OpsGenie API key value in output")
	}
	if !strings.Contains(result, "dynatrace_victorops_notification") {
		t.Error("expected dynatrace_victorops_notification resource type")
	}
	if !strings.Contains(result, "vo-key-789") {
		t.Error("expected VictorOps API key value in output")
	}
	// Verify no "Unsupported" comments for known types
	if strings.Contains(result, "Unsupported") {
		t.Error("unexpected 'Unsupported' comment in output for known notification types")
	}
}

func TestGenerateNotebooks(t *testing.T) {
	notebooks := []dynatrace.DynatraceNotebook{
		{
			Name: "Investigation Notebook",
			Sections: []dynatrace.NotebookSection{
				{Type: "markdown", Content: "# Overview"},
			},
		},
	}
	result := GenerateNotebooks(notebooks)
	if !strings.Contains(result, "dynatrace_document") {
		t.Error("expected dynatrace_document resource type")
	}
	if !strings.Contains(result, "Investigation Notebook") {
		t.Error("expected notebook name in output")
	}
	if !strings.Contains(result, "notebook") {
		t.Error("expected type = notebook in output")
	}
}

func TestUniqueName(t *testing.T) {
	if got := uniqueName("test", 0); got != "test" {
		t.Errorf("uniqueName(test, 0) = %q, want %q", got, "test")
	}
	if got := uniqueName("test", 1); got != "test_1" {
		t.Errorf("uniqueName(test, 1) = %q, want %q", got, "test_1")
	}
	if got := uniqueName("test", 5); got != "test_5" {
		t.Errorf("uniqueName(test, 5) = %q, want %q", got, "test_5")
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
