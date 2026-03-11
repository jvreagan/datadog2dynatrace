package terraform

import (
	"os"
	"path/filepath"
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

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("/tmp/test-output")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.outputDir != "/tmp/test-output" {
		t.Errorf("outputDir: got %q, want %q", g.outputDir, "/tmp/test-output")
	}
}

func TestGenerateAll(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	result := &dynatrace.ConversionResult{
		Dashboards: []dynatrace.Dashboard{
			{DashboardMetadata: dynatrace.DashboardMetadata{Name: "Test"}},
		},
		MetricEvents: []dynatrace.MetricEvent{
			{Summary: "CPU", Enabled: true, EventType: "CUSTOM_ALERT", MetricSelector: "builtin:host.cpu.user",
				MonitoringStrategy: dynatrace.MonitoringStrategy{Type: "STATIC_THRESHOLD", AlertCondition: "ABOVE", Threshold: 90, Samples: 5, ViolatingSamples: 3, DealertingSamples: 5}},
		},
		SLOs: []dynatrace.SLO{
			{Name: "Avail", Enabled: true, MetricExpression: "(100)*(good)/(total)", EvaluationType: "AGGREGATE", Target: 99.9, Timeframe: "-1M"},
		},
		Synthetics: []dynatrace.SyntheticMonitor{
			{Name: "Health", Type: "HTTP", Enabled: true, FrequencyMin: 5, Locations: []string{"GEOLOCATION-1"}},
		},
		LogRules: []dynatrace.LogProcessingRule{
			{Name: "Parse", Enabled: true, Query: "source:nginx"},
		},
		Maintenance: []dynatrace.MaintenanceWindow{
			{Name: "Deploy", Type: "PLANNED", Suppression: "DETECT_PROBLEMS_DONT_ALERT",
				Schedule: dynatrace.MaintenanceSchedule{RecurrenceType: "ONCE", Start: "2024-01-01 00:00", End: "2024-01-01 02:00", ZoneID: "UTC"}},
		},
		Notifications: []dynatrace.NotificationIntegration{
			{Name: "Slack", Type: "SLACK", Active: true, Config: map[string]interface{}{"url": "https://hooks.slack.com/test"}},
		},
		Notebooks: []dynatrace.DynatraceNotebook{
			{Name: "NB", Sections: []dynatrace.NotebookSection{{Type: "markdown", Content: "# Test"}}},
		},
	}

	if err := g.GenerateAll(result); err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	expectedFiles := []string{
		"provider.tf", "dashboards.tf", "metric_events.tf", "slos.tf",
		"synthetics.tf", "log_processing.tf", "maintenance.tf",
		"notifications.tf", "notebooks.tf",
	}
	for _, f := range expectedFiles {
		info, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			t.Errorf("expected file %s: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s is empty", f)
		}
	}
}

func TestGenerateAllEmptyResult(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	result := &dynatrace.ConversionResult{}
	if err := g.GenerateAll(result); err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Only provider.tf should exist
	if _, err := os.Stat(filepath.Join(dir, "provider.tf")); err != nil {
		t.Error("expected provider.tf to exist")
	}

	// Optional files should not be created for empty results
	optionalFiles := []string{"dashboards.tf", "metric_events.tf", "slos.tf",
		"synthetics.tf", "log_processing.tf", "maintenance.tf",
		"notifications.tf", "notebooks.tf"}
	for _, f := range optionalFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Errorf("did not expect %s for empty result", f)
		}
	}
}

func TestGenerateAllBadDirectory(t *testing.T) {
	g := NewGenerator("/proc/nonexistent/impossible/path")
	result := &dynatrace.ConversionResult{}
	if err := g.GenerateAll(result); err == nil {
		t.Error("expected error for bad output directory")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	g := NewGenerator(dir)

	if err := g.writeFile("test.tf", "content here"); err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "test.tf"))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(data) != "content here" {
		t.Errorf("file content: got %q, want %q", string(data), "content here")
	}
}

func TestGenerateSyntheticsWithLocations(t *testing.T) {
	monitors := []dynatrace.SyntheticMonitor{
		{
			Name:         "API Check",
			Type:         "HTTP",
			Enabled:      true,
			FrequencyMin: 10,
			Locations:    []string{"GEOLOCATION-1", "GEOLOCATION-2"},
		},
	}
	result := GenerateSynthetics(monitors)
	if !strings.Contains(result, "GEOLOCATION-1") {
		t.Error("expected location in output")
	}
	if !strings.Contains(result, "locations") {
		t.Error("expected locations field")
	}
	if !strings.Contains(result, "frequency = 10") {
		t.Error("expected frequency in output")
	}
}

func TestGenerateSyntheticsWithScript(t *testing.T) {
	monitors := []dynatrace.SyntheticMonitor{
		{
			Name:    "Scripted Check",
			Type:    "HTTP",
			Enabled: true,
			Script: &dynatrace.SyntheticScript{
				Version: "1.0",
				Type:    "availability",
				Requests: []dynatrace.ScriptRequest{
					{Description: "GET /health", URL: "https://example.com/health"},
				},
			},
		},
	}
	result := GenerateSynthetics(monitors)
	if !strings.Contains(result, "script") {
		t.Error("expected script block in output")
	}
	if !strings.Contains(result, "jsonencode") {
		t.Error("expected jsonencode for script")
	}
	if !strings.Contains(result, "https://example.com/health") {
		t.Error("expected script URL in output")
	}
}

func TestGenerateSyntheticsWithAnomalyDetection(t *testing.T) {
	monitors := []dynatrace.SyntheticMonitor{
		{
			Name:    "Monitored Check",
			Type:    "HTTP",
			Enabled: true,
			AnomalyDetection: &dynatrace.AnomalyDetection{
				OutageHandling: &dynatrace.OutageHandling{
					GlobalOutage: true,
					LocalOutage:  false,
					RetryOnError: true,
				},
			},
		},
	}
	result := GenerateSynthetics(monitors)
	if !strings.Contains(result, "anomaly_detection") {
		t.Error("expected anomaly_detection block")
	}
	if !strings.Contains(result, "outage_handling") {
		t.Error("expected outage_handling block")
	}
	if !strings.Contains(result, "global_outage  = true") {
		t.Error("expected global_outage = true")
	}
	if !strings.Contains(result, "retry_on_error = true") {
		t.Error("expected retry_on_error = true")
	}
}

func TestGenerateMaintenanceWithScope(t *testing.T) {
	windows := []dynatrace.MaintenanceWindow{
		{
			Name:        "Scoped Maintenance",
			Type:        "PLANNED",
			Suppression: "DONT_DETECT_PROBLEMS",
			Description: "Monthly deploy window",
			Schedule: dynatrace.MaintenanceSchedule{
				RecurrenceType: "MONTHLY",
				Start:          "2024-01-01 00:00",
				End:            "2024-01-01 04:00",
				ZoneID:         "America/New_York",
			},
			Scope: &dynatrace.MaintenanceScope{
				Matches: []dynatrace.MaintenanceScopeMatch{
					{
						Tags: []dynatrace.METag{
							{Key: "env", Value: "production"},
							{Key: "service", Value: "api"},
						},
					},
				},
			},
		},
	}
	result := GenerateMaintenance(windows)
	if !strings.Contains(result, "filter") {
		t.Error("expected filter block")
	}
	if !strings.Contains(result, "env:production") {
		t.Error("expected tag env:production")
	}
	if !strings.Contains(result, "service:api") {
		t.Error("expected tag service:api")
	}
	if !strings.Contains(result, "Monthly deploy window") {
		t.Error("expected description in output")
	}
	if !strings.Contains(result, "DONT_DETECT_PROBLEMS") {
		t.Error("expected suppression type in output")
	}
}

func TestGenerateNotificationUnsupportedType(t *testing.T) {
	notifications := []dynatrace.NotificationIntegration{
		{
			Name:   "Unknown Channel",
			Type:   "XMATTERS",
			Active: true,
			Config: map[string]interface{}{},
		},
	}
	result := GenerateNotifications(notifications)
	if !strings.Contains(result, "Unsupported notification type: XMATTERS") {
		t.Error("expected unsupported comment for unknown type")
	}
}

func TestGenerateNotificationEmail(t *testing.T) {
	notifications := []dynatrace.NotificationIntegration{
		{
			Name:   "Team Email",
			Type:   "EMAIL",
			Active: true,
			Config: map[string]interface{}{"receivers": "team@example.com"},
		},
	}
	result := GenerateNotifications(notifications)
	if !strings.Contains(result, "dynatrace_email_notification") {
		t.Error("expected dynatrace_email_notification resource type")
	}
	if !strings.Contains(result, "team@example.com") {
		t.Error("expected email receivers in output")
	}
}

func TestGenerateMetricEventsWithDescription(t *testing.T) {
	events := []dynatrace.MetricEvent{
		{
			Summary:        "High Memory",
			Description:    "Alert when memory exceeds threshold",
			Enabled:        true,
			EventType:      "CUSTOM_ALERT",
			MetricSelector: "builtin:host.mem.usage",
			MonitoringStrategy: dynatrace.MonitoringStrategy{
				Type: "STATIC_THRESHOLD", AlertCondition: "ABOVE",
				Threshold: 85, Samples: 5, ViolatingSamples: 3, DealertingSamples: 5,
			},
		},
	}
	result := GenerateMetricEvents(events)
	if !strings.Contains(result, "Alert when memory exceeds threshold") {
		t.Error("expected description in output")
	}
}

func TestGenerateSLOsWithDescription(t *testing.T) {
	slos := []dynatrace.SLO{
		{
			Name:             "API Latency",
			Description:      "P99 latency under 500ms",
			Enabled:          true,
			MetricExpression: "(100)*(good)/(total)",
			EvaluationType:   "AGGREGATE",
			Target:           99.0,
			Warning:          99.5,
			Timeframe:        "-1w",
		},
	}
	result := GenerateSLOs(slos)
	if !strings.Contains(result, "P99 latency under 500ms") {
		t.Error("expected description in output")
	}
	if !strings.Contains(result, "-1w") {
		t.Error("expected timeframe in output")
	}
}

func TestGenerateLogProcessingWithProcessor(t *testing.T) {
	rules := []dynatrace.LogProcessingRule{
		{
			Name:      "Parse JSON",
			Enabled:   true,
			Query:     "log.source:app",
			Processor: "PARSE(content, \"JSON\")",
		},
	}
	result := GenerateLogProcessing(rules)
	if !strings.Contains(result, "PARSE(content") {
		t.Error("expected processor definition in output")
	}
}
