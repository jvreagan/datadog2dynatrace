package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

func TestReportExtractionSummary(t *testing.T) {
	r := New()
	ext := &datadog.ExtractionResult{
		Dashboards:   make([]datadog.Dashboard, 2),
		Monitors:     make([]datadog.Monitor, 3),
		SLOs:         make([]datadog.SLO, 1),
		Synthetics:   make([]datadog.SyntheticTest, 1),
		LogPipelines: make([]datadog.LogPipeline, 1),
		Downtimes:    make([]datadog.Downtime, 1),
		Notebooks:    make([]datadog.Notebook, 1),
	}
	r.AddExtractionSummary(ext)

	r.SetSource("file", "/tmp/dd")
	r.SetTarget("terraform", "/tmp/dt")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	for _, want := range []string{"Dashboards", "Monitors", "SLOs", "Synthetic Tests", "Log Pipelines", "Notebooks"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected report to contain %q", want)
		}
	}
}

func TestReportConversionSummary(t *testing.T) {
	r := New()
	result := &dynatrace.ConversionResult{
		Dashboards:    make([]dynatrace.Dashboard, 1),
		MetricEvents:  make([]dynatrace.MetricEvent, 2),
		SLOs:          make([]dynatrace.SLO, 1),
		Synthetics:    make([]dynatrace.SyntheticMonitor, 1),
		LogRules:      make([]dynatrace.LogProcessingRule, 1),
		Maintenance:   make([]dynatrace.MaintenanceWindow, 1),
		Notifications: make([]dynatrace.NotificationIntegration, 1),
		Notebooks:     make([]dynatrace.DynatraceNotebook, 1),
	}
	r.AddConversionSummary(result)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	for _, want := range []string{"Dashboards", "Metric Events", "SLOs", "Synthetic Monitors", "Log Processing Rules", "Maintenance Windows", "Notifications", "Notebooks"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected report to contain %q", want)
		}
	}
}

func TestReportConversionErrors(t *testing.T) {
	r := New()
	r.AddConversionErrors([]error{
		errString("dashboard conversion failed"),
		errString("monitor conversion failed"),
	})

	r.SetSource("api", "")
	r.SetTarget("terraform", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "2 conversion errors") {
		t.Error("expected error count in report")
	}
	if !strings.Contains(content, "dashboard conversion failed") {
		t.Error("expected error message in report")
	}
}

func TestReportWriteToFile(t *testing.T) {
	r := New()
	r.SetSource("file", "/tmp/dd")
	r.SetTarget("terraform", "/tmp/dt")
	r.SetDryRun(true)

	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Migration Report") {
		t.Error("expected title in report")
	}
	if !strings.Contains(content, "Dry Run") {
		t.Error("expected dry run indicator")
	}
	if !strings.Contains(content, "file") {
		t.Error("expected source in report")
	}
}

func TestReportExtractionSummaryWithNames(t *testing.T) {
	r := New()
	ext := &datadog.ExtractionResult{
		Dashboards: []datadog.Dashboard{
			{Title: "Dash A"},
			{Title: "Dash B"},
		},
		Monitors: []datadog.Monitor{
			{Name: "Mon 1"},
		},
	}
	r.AddExtractionSummary(ext)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Dash A") {
		t.Error("expected 'Dash A' in extraction summary")
	}
	if !strings.Contains(content, "Dash B") {
		t.Error("expected 'Dash B' in extraction summary")
	}
	if !strings.Contains(content, "Mon 1") {
		t.Error("expected 'Mon 1' in extraction summary")
	}
}

func TestReportDashboardDetails(t *testing.T) {
	r := New()
	dashboards := []dynatrace.Dashboard{
		{
			DashboardMetadata: dynatrace.DashboardMetadata{Name: "Test Dash"},
			Tiles: []dynatrace.Tile{
				{TileType: "DATA_EXPLORER", Name: "CPU"},
				{TileType: "DATA_EXPLORER", Name: "Memory"},
				{TileType: "MARKDOWN", Name: "Notes"},
			},
		},
	}
	r.AddDashboardDetails(dashboards)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Test Dash") {
		t.Error("expected dashboard name in details")
	}
	if !strings.Contains(content, "DATA_EXPLORER") {
		t.Error("expected DATA_EXPLORER tile type in details")
	}
	if !strings.Contains(content, "Total tiles: 3") {
		t.Error("expected 'Total tiles: 3' in details")
	}
}

func TestReportDQLQueryNotes(t *testing.T) {
	r := New()
	dashboards := []dynatrace.Dashboard{
		{
			DashboardMetadata: dynatrace.DashboardMetadata{Name: "Log Dashboard"},
			Tiles: []dynatrace.Tile{
				{TileType: "MARKDOWN", Name: "Log Errors", Markdown: "fetch logs\n| filter loglevel == \"ERROR\""},
				{TileType: "DATA_EXPLORER", Name: "CPU"},
			},
		},
	}
	r.AddDQLQueryNotes(dashboards)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "DQL Query Notes") {
		t.Error("expected DQL Query Notes section")
	}
	if !strings.Contains(content, "Log Dashboard") {
		t.Error("expected dashboard name in DQL notes")
	}
	if !strings.Contains(content, "Log Errors") {
		t.Error("expected tile name in DQL notes")
	}
}

func TestJoinResourceNamesTruncation(t *testing.T) {
	r := New()
	ext := &datadog.ExtractionResult{
		Dashboards: []datadog.Dashboard{
			{Title: "D1"}, {Title: "D2"}, {Title: "D3"},
			{Title: "D4"}, {Title: "D5"}, {Title: "D6"}, {Title: "D7"},
		},
	}
	r.AddExtractionSummary(ext)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "+2 more") {
		t.Error("expected '+2 more' for truncated names")
	}
}

func TestReportPushErrors(t *testing.T) {
	r := New()
	r.AddPushErrors([]error{
		errString("dashboard push failed"),
		errString("SLO push timeout"),
	})

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Push Errors") {
		t.Error("expected 'Push Errors' section")
	}
	if !strings.Contains(content, "2 push errors") {
		t.Error("expected push error count")
	}
	if !strings.Contains(content, "dashboard push failed") {
		t.Error("expected push error message")
	}
	if !strings.Contains(content, "SLO push timeout") {
		t.Error("expected second push error message")
	}
}

func TestReportPushErrorsEmpty(t *testing.T) {
	r := New()
	r.AddPushErrors(nil)
	r.AddPushErrors([]error{})

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if strings.Contains(content, "Push Errors") {
		t.Error("should not include Push Errors section when empty")
	}
}

func TestReportConversionErrorsEmpty(t *testing.T) {
	r := New()
	r.AddConversionErrors(nil)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "Conversion Errors") {
		t.Error("should not include Conversion Errors section when empty")
	}
}

func TestReportValidationResults(t *testing.T) {
	r := New()
	r.AddValidationResults(&dynatrace.ValidationResult{
		Selectors: []dynatrace.SelectorValidation{
			{Selector: "builtin:host.cpu.usage", Sources: []string{"monitor: High CPU"}, Valid: true},
			{Selector: "bad.metric", Sources: []string{"dashboard tile: Error"}, Valid: false, Error: "unknown metric"},
			{Selector: "builtin:host.availability", Sources: []string{"monitor: Composite"}, Skipped: true},
		},
		Summary: dynatrace.ValidationSummary{Total: 3, Valid: 1, Invalid: 1, Skipped: 1},
	})

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "Metric Selector Validation") {
		t.Error("expected 'Metric Selector Validation' section")
	}
	if !strings.Contains(content, "3") {
		t.Error("expected total count")
	}
	if !strings.Contains(content, "builtin:host.cpu.usage") {
		t.Error("expected valid selector")
	}
	if !strings.Contains(content, "bad.metric") {
		t.Error("expected invalid selector")
	}
	if !strings.Contains(content, "unknown metric") {
		t.Error("expected error message")
	}
	if !strings.Contains(content, "Skipped") {
		t.Error("expected skipped status")
	}
}

func TestReportValidationResultsNil(t *testing.T) {
	r := New()
	r.AddValidationResults(nil)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "Metric Selector Validation") {
		t.Error("should not include validation section when nil")
	}
}

func TestReportValidationResultsPipeEscaping(t *testing.T) {
	r := New()
	r.AddValidationResults(&dynatrace.ValidationResult{
		Selectors: []dynatrace.SelectorValidation{
			{Selector: "bad.metric", Sources: []string{"src"}, Valid: false, Error: "error|with|pipes"},
		},
		Summary: dynatrace.ValidationSummary{Total: 1, Invalid: 1},
	})

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if strings.Contains(content, "error|with") {
		t.Error("expected pipes to be escaped in markdown table")
	}
	if !strings.Contains(content, `error\|with\|pipes`) {
		t.Error("expected escaped pipes")
	}
}

func TestReportDashboardDetailsEmpty(t *testing.T) {
	r := New()
	r.AddDashboardDetails(nil)
	r.AddDashboardDetails([]dynatrace.Dashboard{})

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "Dashboard Details") {
		t.Error("should not include Dashboard Details when empty")
	}
}

func TestReportDQLQueryNotesGrail(t *testing.T) {
	r := New()
	dashboards := []dynatrace.Dashboard{
		{
			DashboardMetadata: dynatrace.DashboardMetadata{Name: "Grail Dashboard"},
			Tiles: []dynatrace.Tile{
				{
					TileType: "DATA_EXPLORER",
					Name:     "DQL Tile",
					Queries: []dynatrace.DashboardQuery{
						{DQL: "fetch logs | filter status == \"ERROR\""},
					},
				},
			},
		},
	}
	r.AddDQLQueryNotes(dashboards)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "native DQL queries") {
		t.Error("expected Grail DQL note")
	}
	if !strings.Contains(content, "Grail Dashboard") {
		t.Error("expected dashboard name in Grail note")
	}
}

func TestReportDQLQueryNotesNoNotes(t *testing.T) {
	r := New()
	dashboards := []dynatrace.Dashboard{
		{
			DashboardMetadata: dynatrace.DashboardMetadata{Name: "Simple"},
			Tiles: []dynatrace.Tile{
				{TileType: "DATA_EXPLORER", Name: "CPU"},
			},
		},
	}
	r.AddDQLQueryNotes(dashboards)

	r.SetSource("api", "")
	r.SetTarget("api", "")
	path := filepath.Join(t.TempDir(), "report.md")
	if err := r.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile error: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "DQL Query Notes") {
		t.Error("should not include DQL Query Notes when no DQL found")
	}
}

func TestJoinResourceNames(t *testing.T) {
	if got := joinResourceNames(nil); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
	if got := joinResourceNames([]string{"a"}); got != "a" {
		t.Errorf("expected 'a', got %q", got)
	}
	if got := joinResourceNames([]string{"a", "b", "c", "d", "e"}); got != "a, b, c, d, e" {
		t.Errorf("expected all 5 joined, got %q", got)
	}
	got := joinResourceNames([]string{"a", "b", "c", "d", "e", "f"})
	if !strings.Contains(got, "+1 more") {
		t.Errorf("expected truncation, got %q", got)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
