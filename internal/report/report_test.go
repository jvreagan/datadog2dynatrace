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

type errString string

func (e errString) Error() string { return string(e) }
