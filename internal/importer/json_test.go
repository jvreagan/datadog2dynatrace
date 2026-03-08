package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportDashboardsJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCount int
	}{
		{
			name:      "array format",
			json:      `[{"title":"Dashboard 1","widgets":[]},{"title":"Dashboard 2","widgets":[]}]`,
			wantCount: 2,
		},
		{
			name:      "single object",
			json:      `{"title":"My Dashboard","widgets":[]}`,
			wantCount: 1,
		},
		{
			name:      "wrapper format",
			json:      `{"dashboards":[{"title":"Wrapped","widgets":[]}]}`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, dir, "dashboards.json", tt.json)
			result, err := ImportFromDirectory(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Dashboards) != tt.wantCount {
				t.Errorf("got %d dashboards, want %d", len(result.Dashboards), tt.wantCount)
			}
		})
	}
}

func TestImportMonitorsJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantCount int
	}{
		{
			name:      "array format",
			json:      `[{"name":"Mon1","type":"metric alert","query":"avg:system.cpu.user{*}"},{"name":"Mon2","type":"metric alert","query":"avg:system.mem.used{*}"}]`,
			wantCount: 2,
		},
		{
			name:      "single object",
			json:      `{"name":"Mon1","type":"metric alert","query":"avg:system.cpu.user{*}"}`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, dir, "monitors.json", tt.json)
			result, err := ImportFromDirectory(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Monitors) != tt.wantCount {
				t.Errorf("got %d monitors, want %d", len(result.Monitors), tt.wantCount)
			}
		})
	}
}

func TestImportSLOsJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "slos.json", `{"data":[{"name":"API Avail","type":"metric","thresholds":[{"timeframe":"7d","target":99.9}]}]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SLOs) != 1 {
		t.Errorf("got %d SLOs, want 1", len(result.SLOs))
	}
}

func TestImportSyntheticsJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "synthetics.json", `{"tests":[{"name":"Health Check","type":"api","public_id":"abc-123"}]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("got %d synthetics, want 1", len(result.Synthetics))
	}
}

func TestImportLogPipelinesJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "log_pipelines.json", `[{"name":"Pipeline1","is_enabled":true}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("got %d log pipelines, want 1", len(result.LogPipelines))
	}
}

func TestImportDowntimesJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "downtimes.json", `[{"id":1,"scope":["env:prod"],"message":"Maintenance"}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("got %d downtimes, want 1", len(result.Downtimes))
	}
}

func TestImportNotebooksJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notebooks.json", `[{"id":1,"name":"My Notebook","cells":[]}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("got %d notebooks, want 1", len(result.Notebooks))
	}
}

func TestAutoImportDashboard(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"title":"Auto Dashboard","widgets":[]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("got %d dashboards, want 1", len(result.Dashboards))
	}
}

func TestAutoImportMonitor(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"name":"Auto Monitor","type":"metric alert","query":"avg:system.cpu.user{*}"}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Monitors) != 1 {
		t.Errorf("got %d monitors, want 1", len(result.Monitors))
	}
}

func TestAutoImportSLO(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"name":"Auto SLO","thresholds":[{"timeframe":"7d","target":99.9}]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SLOs) != 1 {
		t.Errorf("got %d SLOs, want 1", len(result.SLOs))
	}
}

func TestAutoImportSynthetic(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"name":"Auto Syn","public_id":"abc-123","type":"api"}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("got %d synthetics, want 1", len(result.Synthetics))
	}
}

func TestAutoImportUnknown(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"foo":"bar","baz":123}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 0 || len(result.Monitors) != 0 || len(result.SLOs) != 0 {
		t.Error("expected no resources imported for unknown format")
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
}
