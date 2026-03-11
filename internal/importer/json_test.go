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

// --- Notification import tests ---

func TestImportNotificationsArray(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notifications.json", `[
		{"id":1,"name":"Slack","type":"slack","config":{"url":"https://hooks.slack.com/test","channel":"#alerts"}},
		{"id":2,"name":"PD","type":"pagerduty","config":{"service_name":"Prod","service_key":"key-1"}}
	]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notifications) != 2 {
		t.Errorf("got %d notifications, want 2", len(result.Notifications))
	}
}

func TestImportNotificationsSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notifications.json", `{"id":1,"name":"Webhook","type":"webhook","config":{"url":"https://example.com"}}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notifications) != 1 {
		t.Errorf("got %d notifications, want 1", len(result.Notifications))
	}
}

// --- Single-object import tests for types that only tested array ---

func TestImportLogPipelineSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "log_pipelines.json", `{"id":"p1","name":"Single Pipeline","is_enabled":true}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("got %d log pipelines, want 1", len(result.LogPipelines))
	}
}

func TestImportDowntimeSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "downtimes.json", `{"id":1,"scope":["env:prod"],"message":"Deploy"}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("got %d downtimes, want 1", len(result.Downtimes))
	}
}

func TestImportNotebookSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notebooks.json", `{"id":1,"name":"Single NB","cells":[]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("got %d notebooks, want 1", len(result.Notebooks))
	}
}

func TestImportSLOSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "slos.json", `{"id":"s1","name":"Single SLO","type":"metric","thresholds":[{"timeframe":"7d","target":99.5}]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SLOs) != 1 {
		t.Errorf("got %d SLOs, want 1", len(result.SLOs))
	}
}

func TestImportSyntheticSingle(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "synthetics.json", `{"name":"Single Syn","public_id":"abc","type":"api"}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("got %d synthetics, want 1", len(result.Synthetics))
	}
}

func TestImportSLOArrayFormat(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "slos.json", `[{"id":"s1","name":"SLO1","type":"metric"},{"id":"s2","name":"SLO2","type":"metric"}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.SLOs) != 2 {
		t.Errorf("got %d SLOs, want 2", len(result.SLOs))
	}
}

func TestImportSyntheticArrayFormat(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "synthetics.json", `[{"name":"Syn1","type":"api"},{"name":"Syn2","type":"browser"}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Synthetics) != 2 {
		t.Errorf("got %d synthetics, want 2", len(result.Synthetics))
	}
}

// --- Auto-import tests for uncovered detection paths ---

func TestAutoImportLogPipeline(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"name":"Auto Pipeline","processors":[{"type":"grok-parser"}],"is_enabled":true}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("got %d log pipelines, want 1", len(result.LogPipelines))
	}
}

func TestAutoImportDowntimeMonitorID(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"id":1,"scope":["env:prod"],"monitor_id":100}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("got %d downtimes, want 1", len(result.Downtimes))
	}
}

func TestAutoImportDowntimeMonitorTags(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"id":1,"scope":["env:staging"],"monitor_tags":["service:web"]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("got %d downtimes, want 1", len(result.Downtimes))
	}
}

func TestAutoImportNotebook(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `{"name":"Auto NB","cells":[],"author":{"handle":"user@test.com"}}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("got %d notebooks, want 1", len(result.Notebooks))
	}
}

func TestAutoImportInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "unknown.json", `not valid json at all`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should silently skip invalid JSON
	total := len(result.Dashboards) + len(result.Monitors) + len(result.SLOs)
	if total != 0 {
		t.Errorf("expected 0 resources from invalid JSON, got %d", total)
	}
}

// --- Error case tests ---

func TestImportInvalidDirectory(t *testing.T) {
	_, err := ImportFromDirectory("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestImportBadJSON(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"bad dashboard", "dashboards.json", `not json`},
		{"bad monitors", "monitors.json", `not json`},
		{"bad notifications", "notifications.json", `not json`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeTestFile(t, dir, tt.filename, tt.content)
			_, err := ImportFromDirectory(dir)
			if err == nil {
				t.Errorf("expected error for bad JSON in %s", tt.filename)
			}
		})
	}
}

func TestImportSkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)
	writeTestFile(t, dir, "monitors.json", `[{"name":"Mon","type":"metric alert","query":"avg:cpu{*}"}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Monitors) != 1 {
		t.Errorf("got %d monitors, want 1", len(result.Monitors))
	}
}

func TestImportSkipsNonJSONFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "readme.txt", "not a json file")
	writeTestFile(t, dir, "monitors.json", `[{"name":"Mon","type":"metric alert","query":"avg:cpu{*}"}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Monitors) != 1 {
		t.Errorf("got %d monitors, want 1", len(result.Monitors))
	}
}

func TestImportTerraformJSON(t *testing.T) {
	dir := t.TempDir()
	tfJSON := `{
		"resource": {
			"datadog_dashboard": {
				"main": {"title":"TF Dashboard","widgets":[]}
			},
			"datadog_monitor": {
				"cpu": {"name":"CPU Alert","type":"metric alert","query":"avg:system.cpu{*}"}
			},
			"datadog_service_level_objective": {
				"slo1": {"name":"API SLO","type":"metric"}
			},
			"datadog_synthetics_test": {
				"health": {"name":"Health Check","type":"api","status":"live"}
			},
			"datadog_logs_custom_pipeline": {
				"pipe1": {"name":"JSON Pipeline","is_enabled":true}
			},
			"datadog_downtime": {
				"deploy": {"id":1,"message":"Deploy","scope":["*"]}
			},
			"datadog_notebook": {
				"nb1": {"name":"My Notebook"}
			}
		}
	}`
	writeTestFile(t, dir, "main.tf.json", tfJSON)

	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("expected 1 dashboard, got %d", len(result.Dashboards))
	}
	if len(result.Monitors) != 1 {
		t.Errorf("expected 1 monitor, got %d", len(result.Monitors))
	}
	if len(result.SLOs) != 1 {
		t.Errorf("expected 1 SLO, got %d", len(result.SLOs))
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("expected 1 synthetic, got %d", len(result.Synthetics))
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("expected 1 log pipeline, got %d", len(result.LogPipelines))
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("expected 1 downtime, got %d", len(result.Downtimes))
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("expected 1 notebook, got %d", len(result.Notebooks))
	}
}

func TestImportTerraformJSONBadJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "bad.tf.json", `{not valid}`)
	_, err := ImportFromDirectory(dir)
	if err == nil {
		t.Fatal("expected error for bad tf.json")
	}
}

func TestImportTerraformHCL(t *testing.T) {
	dir := t.TempDir()
	hcl := `
resource "datadog_dashboard" "main" {
  title = "My Dashboard"
  widget {
    timeseries_definition {}
  }
}

resource "datadog_monitor" "cpu" {
  name  = "CPU Alert"
  type  = "metric alert"
  query = "avg:system.cpu{*} > 90"
}

resource "datadog_service_level_objective" "slo" {
  name = "API SLO"
}

resource "datadog_synthetics_test" "health" {
  name = "Health Check"
  type = "api"
}

resource "datadog_logs_custom_pipeline" "json" {
  name = "JSON Parser"
}

resource "datadog_downtime" "deploy" {
  message = "Deploy window"
}
`
	writeTestFile(t, dir, "main.tf", hcl)

	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("expected 1 dashboard, got %d", len(result.Dashboards))
	}
	if len(result.Monitors) != 1 {
		t.Errorf("expected 1 monitor, got %d", len(result.Monitors))
	}
	if len(result.SLOs) != 1 {
		t.Errorf("expected 1 SLO, got %d", len(result.SLOs))
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("expected 1 synthetic, got %d", len(result.Synthetics))
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("expected 1 log pipeline, got %d", len(result.LogPipelines))
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("expected 1 downtime, got %d", len(result.Downtimes))
	}
}

func TestImportDashboardsJSONBadArray(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "dashboards.json", `[{"title":"Good"},{"bad json}]`)
	_, err := ImportFromDirectory(dir)
	if err == nil {
		t.Fatal("expected error for bad JSON array")
	}
}

func TestImportNotificationsJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notifications.json", `[{"name":"Slack","type":"slack","config":{"url":"https://hooks.slack.com"}}]`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(result.Notifications))
	}
}

func TestImportBadLogPipeline(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "log_pipelines.json", `{bad json`)
	_, err := ImportFromDirectory(dir)
	if err == nil {
		t.Fatal("expected error for bad log pipeline JSON")
	}
}

func TestImportBadDowntime(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "downtimes.json", `{bad json`)
	_, err := ImportFromDirectory(dir)
	if err == nil {
		t.Fatal("expected error for bad downtime JSON")
	}
}

func TestImportBadNotebook(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notebooks.json", `{bad json`)
	_, err := ImportFromDirectory(dir)
	if err == nil {
		t.Fatal("expected error for bad notebook JSON")
	}
}

func TestImportSingleLogPipeline(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "log_pipelines.json", `{"name":"My Pipeline","is_enabled":true}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("expected 1 log pipeline, got %d", len(result.LogPipelines))
	}
}

func TestImportSingleDowntime(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "downtimes.json", `{"message":"Deploy window","scope":["env:prod"]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("expected 1 downtime, got %d", len(result.Downtimes))
	}
}

func TestImportSingleNotebook(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "notebooks.json", `{"name":"Investigation","cells":[]}`)
	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("expected 1 notebook, got %d", len(result.Notebooks))
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
}
