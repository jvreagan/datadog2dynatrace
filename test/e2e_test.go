package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/importer"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/report"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/terraform"
)

// TestEndToEndPipeline runs the full conversion pipeline:
// JSON fixtures → import → convert → terraform generate → report
func TestEndToEndPipeline(t *testing.T) {
	testdataDir := filepath.Join("testdata")

	// Step 1: Import from JSON fixtures
	ext, err := importer.ImportFromDirectory(testdataDir)
	if err != nil {
		t.Fatalf("ImportFromDirectory failed: %v", err)
	}

	// Verify extraction counts
	if len(ext.Dashboards) != 1 {
		t.Errorf("expected 1 dashboard, got %d", len(ext.Dashboards))
	}
	if len(ext.Monitors) != 6 {
		t.Errorf("expected 6 monitors, got %d", len(ext.Monitors))
	}
	if len(ext.SLOs) != 2 {
		t.Errorf("expected 2 SLOs, got %d", len(ext.SLOs))
	}
	if len(ext.Synthetics) != 2 {
		t.Errorf("expected 2 synthetics, got %d", len(ext.Synthetics))
	}
	if len(ext.LogPipelines) != 2 {
		t.Errorf("expected 2 log pipelines, got %d", len(ext.LogPipelines))
	}
	if len(ext.Downtimes) != 2 {
		t.Errorf("expected 2 downtimes, got %d", len(ext.Downtimes))
	}
	if len(ext.Notebooks) != 1 {
		t.Errorf("expected 1 notebook, got %d", len(ext.Notebooks))
	}
	if len(ext.Notifications) != 4 {
		t.Errorf("expected 4 notifications, got %d", len(ext.Notifications))
	}

	// Step 2: Convert DD → DT
	conv := converter.New(converter.Options{})
	result, errs := conv.ConvertAll(ext)

	// Log conversion errors (some are expected for complex queries)
	for _, e := range errs {
		t.Logf("conversion warning: %v", e)
	}

	// Verify conversions produced output
	if len(result.Dashboards) == 0 {
		t.Error("expected at least 1 converted dashboard")
	}
	if len(result.MetricEvents) == 0 {
		t.Error("expected at least 1 metric event")
	}
	if len(result.SLOs) == 0 {
		t.Error("expected at least 1 SLO")
	}
	if len(result.Synthetics) == 0 {
		t.Error("expected at least 1 synthetic monitor")
	}
	if len(result.LogRules) == 0 {
		t.Error("expected at least 1 log processing rule")
	}
	if len(result.Maintenance) == 0 {
		t.Error("expected at least 1 maintenance window")
	}
	if len(result.Notebooks) == 0 {
		t.Error("expected at least 1 notebook")
	}
	if len(result.Notifications) == 0 {
		t.Error("expected at least 1 notification")
	}

	// Step 3: Generate Terraform output
	tfDir := t.TempDir()
	gen := terraform.NewGenerator(tfDir)
	if err := gen.GenerateAll(result); err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Verify TF files were created
	expectedFiles := []string{
		"provider.tf",
		"dashboards.tf",
		"metric_events.tf",
		"slos.tf",
		"synthetics.tf",
		"log_processing.tf",
		"maintenance.tf",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(tfDir, f)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected TF file %s: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("TF file %s is empty", f)
		}
	}

	// Step 4: Generate migration report
	reportPath := filepath.Join(tfDir, "migration-report.md")
	rpt := report.New()
	rpt.SetSource("file", testdataDir)
	rpt.SetTarget("terraform", tfDir)
	rpt.SetDryRun(false)
	rpt.AddExtractionSummary(ext)
	rpt.AddConversionSummary(result)
	rpt.AddConversionErrors(errs)
	if err := rpt.WriteToFile(reportPath); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	reportContent, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}
	report := string(reportContent)
	if !strings.Contains(report, "Migration Report") {
		t.Error("report missing title")
	}
	if !strings.Contains(report, "Dashboards") {
		t.Error("report missing dashboard section")
	}
}

// TestEndToEndDashboardContent verifies the converted dashboard has correct DT structure.
func TestEndToEndDashboardContent(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Dashboards) == 0 {
		t.Fatal("no dashboards converted")
	}

	dash := result.Dashboards[0]

	// Check dashboard metadata
	if dash.DashboardMetadata.Name != "Production Infrastructure Overview" {
		t.Errorf("dashboard name: got %q, want %q", dash.DashboardMetadata.Name, "Production Infrastructure Overview")
	}

	// Check tiles were created (one per widget)
	if len(dash.Tiles) < 5 {
		t.Errorf("expected at least 5 tiles, got %d", len(dash.Tiles))
	}

	// Verify tile types exist
	tileTypes := make(map[string]int)
	for _, tile := range dash.Tiles {
		tileTypes[tile.TileType]++
	}

	if tileTypes["CUSTOM_CHARTING"] == 0 && tileTypes["DATA_EXPLORER"] == 0 {
		t.Error("expected at least one charting or data explorer tile")
	}
	if tileTypes["MARKDOWN"] == 0 {
		t.Error("expected at least one markdown tile")
	}
}

// TestEndToEndMonitorConversion verifies monitors become metric events with correct selectors.
func TestEndToEndMonitorConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.MetricEvents) == 0 {
		t.Fatal("no metric events converted")
	}

	// Check the CPU monitor
	found := false
	for _, me := range result.MetricEvents {
		if strings.Contains(me.Summary, "CPU") {
			found = true
			if me.MetricSelector == "" {
				t.Error("CPU metric event has empty metric selector")
			}
			if !me.Enabled {
				t.Error("expected metric event to be enabled")
			}
			if me.Threshold != 90.0 {
				t.Errorf("threshold: got %f, want 90.0", me.Threshold)
			}
			break
		}
	}
	if !found {
		t.Error("CPU monitor not found in converted metric events")
	}

	// Check the K8s pod restarts monitor
	for _, me := range result.MetricEvents {
		if strings.Contains(me.Summary, "Pod Restarts") || strings.Contains(me.Summary, "Restart") {
			if me.MetricSelector == "" {
				t.Error("K8s metric event has empty metric selector")
			}
			break
		}
	}
}

// TestEndToEndSLOConversion verifies SLOs are converted correctly.
func TestEndToEndSLOConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.SLOs) == 0 {
		t.Fatal("no SLOs converted")
	}

	for _, slo := range result.SLOs {
		if slo.Name == "" {
			t.Error("SLO has empty name")
		}
		if slo.Target <= 0 || slo.Target > 100 {
			t.Errorf("SLO %q target out of range: %f", slo.Name, slo.Target)
		}
		if slo.Timeframe == "" {
			t.Errorf("SLO %q has empty timeframe", slo.Name)
		}
	}
}

// TestEndToEndSyntheticConversion verifies both HTTP and browser synthetics.
func TestEndToEndSyntheticConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Synthetics) < 2 {
		t.Fatalf("expected at least 2 synthetics, got %d", len(result.Synthetics))
	}

	typeCount := make(map[string]int)
	for _, sm := range result.Synthetics {
		typeCount[sm.Type]++

		if sm.Name == "" {
			t.Error("synthetic has empty name")
		}
		if sm.FrequencyMin <= 0 {
			t.Errorf("synthetic %q has invalid frequency: %d", sm.Name, sm.FrequencyMin)
		}
		if len(sm.Locations) == 0 {
			t.Errorf("synthetic %q has no locations", sm.Name)
		}
	}

	if typeCount["HTTP"] == 0 {
		t.Error("expected at least one HTTP synthetic")
	}
	if typeCount["BROWSER"] == 0 {
		t.Error("expected at least one BROWSER synthetic")
	}
}

// TestEndToEndLogPipelineConversion verifies log pipelines become processing rules.
func TestEndToEndLogPipelineConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.LogRules) == 0 {
		t.Fatal("no log processing rules converted")
	}

	for _, rule := range result.LogRules {
		if rule.Name == "" {
			t.Error("log rule has empty name")
		}
		if !rule.Enabled {
			t.Error("expected log rule to be enabled")
		}
	}
}

// TestEndToEndTerraformContent verifies generated TF files contain expected content.
func TestEndToEndTerraformContent(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	tfDir := t.TempDir()
	gen := terraform.NewGenerator(tfDir)
	if err := gen.GenerateAll(result); err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Check provider.tf
	providerContent, err := os.ReadFile(filepath.Join(tfDir, "provider.tf"))
	if err != nil {
		t.Fatalf("reading provider.tf: %v", err)
	}
	provider := string(providerContent)
	if !strings.Contains(provider, "dynatrace-oss/dynatrace") {
		t.Error("provider.tf missing dynatrace-oss/dynatrace")
	}
	if !strings.Contains(provider, "required_providers") {
		t.Error("provider.tf missing required_providers block")
	}

	// Check dashboards.tf
	dashContent, err := os.ReadFile(filepath.Join(tfDir, "dashboards.tf"))
	if err != nil {
		t.Fatalf("reading dashboards.tf: %v", err)
	}
	dash := string(dashContent)
	if !strings.Contains(dash, "dynatrace_json_dashboard") {
		t.Error("dashboards.tf missing dynatrace_json_dashboard resource")
	}
	if !strings.Contains(dash, "Production Infrastructure Overview") {
		t.Error("dashboards.tf missing dashboard name")
	}

	// Check metric_events.tf
	meContent, err := os.ReadFile(filepath.Join(tfDir, "metric_events.tf"))
	if err != nil {
		t.Fatalf("reading metric_events.tf: %v", err)
	}
	me := string(meContent)
	if !strings.Contains(me, "dynatrace_metric_events") {
		t.Error("metric_events.tf missing dynatrace_metric_events resource")
	}

	// Check slos.tf
	sloContent, err := os.ReadFile(filepath.Join(tfDir, "slos.tf"))
	if err != nil {
		t.Fatalf("reading slos.tf: %v", err)
	}
	slo := string(sloContent)
	if !strings.Contains(slo, "dynatrace_slo") {
		t.Error("slos.tf missing dynatrace_slo resource")
	}

	// Check synthetics.tf
	synContent, err := os.ReadFile(filepath.Join(tfDir, "synthetics.tf"))
	if err != nil {
		t.Fatalf("reading synthetics.tf: %v", err)
	}
	syn := string(synContent)
	if !strings.Contains(syn, "dynatrace_") {
		t.Error("synthetics.tf missing dynatrace resource")
	}
}

// TestEndToEndMaintenanceConversion verifies downtime → maintenance window conversion.
func TestEndToEndMaintenanceConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Maintenance) < 2 {
		t.Fatalf("expected at least 2 maintenance windows, got %d", len(result.Maintenance))
	}

	// Verify recurring maintenance
	foundRecurring := false
	foundOneTime := false
	for _, mw := range result.Maintenance {
		if mw.Name == "" {
			t.Error("maintenance window has empty name")
		}
		if mw.Schedule.RecurrenceType == "WEEKLY" {
			foundRecurring = true
		}
		if mw.Schedule.RecurrenceType == "ONCE" {
			foundOneTime = true
		}
	}
	if !foundRecurring {
		t.Error("expected at least one recurring maintenance window")
	}
	if !foundOneTime {
		t.Error("expected at least one one-time maintenance window")
	}
}

// TestEndToEndNotebookConversion verifies notebook cells are converted.
func TestEndToEndNotebookConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Notebooks) == 0 {
		t.Fatal("no notebooks converted")
	}

	nb := result.Notebooks[0]
	if nb.Name != "Production Incident Runbook" {
		t.Errorf("notebook name: got %q, want %q", nb.Name, "Production Incident Runbook")
	}

	if len(nb.Sections) < 2 {
		t.Errorf("expected at least 2 sections, got %d", len(nb.Sections))
	}

	sectionTypes := make(map[string]int)
	for _, s := range nb.Sections {
		sectionTypes[s.Type]++
	}

	if sectionTypes["markdown"] == 0 {
		t.Error("expected at least one markdown section")
	}
}

// TestEndToEndNotificationConversion verifies notification channels are converted correctly.
func TestEndToEndNotificationConversion(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	if len(ext.Notifications) != 4 {
		t.Fatalf("expected 4 notifications, got %d", len(ext.Notifications))
	}

	conv := converter.New(converter.Options{})
	result, errs := conv.ConvertAll(ext)
	for _, e := range errs {
		t.Logf("conversion warning: %v", e)
	}

	if len(result.Notifications) != 4 {
		t.Fatalf("expected 4 converted notifications, got %d", len(result.Notifications))
	}

	typeCount := make(map[string]int)
	for _, n := range result.Notifications {
		typeCount[n.Type]++

		switch n.Type {
		case "PAGER_DUTY":
			if n.Config["account"] == nil || n.Config["account"] == "" {
				t.Error("PagerDuty notification missing account")
			}
			if n.Config["integrationKey"] == nil || n.Config["integrationKey"] == "" {
				t.Error("PagerDuty notification missing integrationKey")
			}
		case "SLACK":
			if n.Config["url"] == nil || n.Config["url"] == "" {
				t.Error("Slack notification missing url")
			}
			if n.Config["channel"] == nil || n.Config["channel"] == "" {
				t.Error("Slack notification missing channel")
			}
		}
	}

	if typeCount["SLACK"] == 0 {
		t.Error("expected at least one SLACK notification")
	}
	if typeCount["PAGER_DUTY"] == 0 {
		t.Error("expected at least one PAGER_DUTY notification")
	}
	if typeCount["WEBHOOK"] == 0 {
		t.Error("expected at least one WEBHOOK notification")
	}
	if typeCount["EMAIL"] == 0 {
		t.Error("expected at least one EMAIL notification")
	}
}

// TestEndToEndReportContent verifies the migration report has all expected sections.
func TestEndToEndReportContent(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, errs := conv.ConvertAll(ext)

	rpt := report.New()
	rpt.SetSource("file", "testdata")
	rpt.SetTarget("terraform", "output")
	rpt.SetDryRun(true)
	rpt.AddExtractionSummary(ext)
	rpt.AddConversionSummary(result)
	rpt.AddConversionErrors(errs)

	reportPath := filepath.Join(t.TempDir(), "report.md")
	if err := rpt.WriteToFile(reportPath); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}

	r := string(content)
	expectedPhrases := []string{
		"Migration Report",
		"Extracted Resources",
		"Converted Resources",
		"Dashboards",
		"Monitors",
		"SLOs",
		"Synthetic",
		"Log",
		"Dry Run",
		"file",
		"terraform",
	}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(r, phrase) {
			t.Errorf("report missing expected phrase: %q", phrase)
		}
	}
}

// TestEndToEndErrorCollection verifies errors are collected (not fatal) during conversion.
func TestEndToEndErrorCollection(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, errs := conv.ConvertAll(ext)

	// Even if there are conversion errors, the pipeline should continue
	// and produce results for the resources it can handle
	totalConverted := len(result.Dashboards) + len(result.MetricEvents) +
		len(result.SLOs) + len(result.Synthetics) + len(result.LogRules) +
		len(result.Maintenance) + len(result.Notifications) + len(result.Notebooks)

	if totalConverted == 0 {
		t.Error("expected at least some conversions to succeed")
	}

	t.Logf("Converted %d resources total, with %d errors", totalConverted, len(errs))
}

// TestEndToEndLogAlertMonitor verifies log alert monitors produce metric events with DQL.
func TestEndToEndLogAlertMonitor(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	found := false
	for _, me := range result.MetricEvents {
		if strings.Contains(me.Summary, "Error Log Volume") {
			found = true
			if me.MetricSelector == "" {
				t.Error("log alert metric event has empty metric selector")
			}
			if !strings.Contains(me.Description, "DQL") {
				t.Errorf("expected DQL migration note in description, got %q", me.Description)
			}
			if !strings.Contains(me.Description, "fetch logs") {
				t.Errorf("expected DQL query in description, got %q", me.Description)
			}
			if me.Threshold != 100.0 {
				t.Errorf("expected threshold 100, got %f", me.Threshold)
			}
			break
		}
	}
	if !found {
		t.Error("log alert monitor not found in converted metric events")
	}
}

// TestEndToEndCompositeMonitor verifies composite monitors produce metric events with referenced IDs.
func TestEndToEndCompositeMonitor(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	found := false
	for _, me := range result.MetricEvents {
		if strings.Contains(me.Summary, "Composite") {
			found = true
			if me.MetricSelector == "" {
				t.Error("composite metric event has empty metric selector")
			}
			if !strings.Contains(me.Description, "10001") || !strings.Contains(me.Description, "10002") {
				t.Errorf("expected referenced monitor IDs in description, got %q", me.Description)
			}
			if !strings.Contains(me.Description, "composite") {
				t.Errorf("expected composite migration note, got %q", me.Description)
			}
			break
		}
	}
	if !found {
		t.Error("composite monitor not found in converted metric events")
	}
}

// TestEndToEndTemplateVarSubstitution verifies dashboard template vars are substituted.
func TestEndToEndTemplateVarSubstitution(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Dashboards) == 0 {
		t.Fatal("no dashboards converted")
	}

	dash := result.Dashboards[0]
	for _, tile := range dash.Tiles {
		for _, q := range tile.Queries {
			if strings.Contains(q.MetricSelector, "$env") {
				t.Errorf("tile %q still has $env in metric selector: %q", tile.Name, q.MetricSelector)
			}
			if strings.Contains(q.MetricSelector, "$host") {
				t.Errorf("tile %q still has $host in metric selector: %q", tile.Name, q.MetricSelector)
			}
		}
	}
}

// TestEndToEndNewWidgetTypes verifies sunburst, scatter, alert_value produce tiles.
func TestEndToEndNewWidgetTypes(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	if len(result.Dashboards) == 0 {
		t.Fatal("no dashboards converted")
	}

	dash := result.Dashboards[0]
	tileNames := make(map[string]bool)
	for _, tile := range dash.Tiles {
		tileNames[tile.Name] = true
	}

	expectedTiles := []string{
		"Disk Usage by Host",
		"CPU vs Memory Scatter",
		"CPU Alert Current Value",
	}
	for _, name := range expectedTiles {
		found := false
		for tileName := range tileNames {
			if strings.Contains(tileName, name) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected tile containing %q, not found in dashboard tiles", name)
		}
	}
}

// TestEndToEndValidation verifies the validation pipeline with a mock DT API.
func TestEndToEndValidation(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{})
	result, _ := conv.ConvertAll(ext)

	// Stand up a mock Dynatrace API that accepts all builtin: selectors
	// and rejects everything else.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		selector := r.URL.Query().Get("metricSelector")
		if strings.HasPrefix(selector, "builtin:") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result":[]}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"unknown metric"}}`))
	}))
	defer srv.Close()

	dtClient := dynatrace.NewTestClient(srv.URL, "test-token")
	valResult := dtClient.ValidateAll(result)

	if valResult.Summary.Total == 0 {
		t.Fatal("expected at least some selectors to validate")
	}

	// Verify summary counts add up
	sum := valResult.Summary.Valid + valResult.Summary.Invalid + valResult.Summary.Skipped
	if sum != valResult.Summary.Total {
		t.Errorf("summary counts don't add up: valid(%d) + invalid(%d) + skipped(%d) != total(%d)",
			valResult.Summary.Valid, valResult.Summary.Invalid, valResult.Summary.Skipped, valResult.Summary.Total)
	}

	// Verify the report integrates validation results
	rpt := report.New()
	rpt.SetSource("file", "testdata")
	rpt.SetTarget("terraform", "output")
	rpt.AddValidationResults(valResult)

	reportPath := filepath.Join(t.TempDir(), "validation-report.md")
	if err := rpt.WriteToFile(reportPath); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}

	r := string(content)
	if !strings.Contains(r, "Metric Selector Validation") {
		t.Error("report missing validation section")
	}
	if !strings.Contains(r, "Valid") {
		t.Error("report missing Valid status")
	}
}

// TestEndToEndGrailDashboard verifies Grail mode produces DQL DATA_EXPLORER tiles.
func TestEndToEndGrailDashboard(t *testing.T) {
	ext, err := importer.ImportFromDirectory(filepath.Join("testdata"))
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	conv := converter.New(converter.Options{EnableGrail: true})
	result, _ := conv.ConvertAll(ext)

	if len(result.Dashboards) == 0 {
		t.Fatal("no dashboards converted")
	}

	dash := result.Dashboards[0]
	hasDataExplorer := false
	for _, tile := range dash.Tiles {
		if tile.TileType == "DATA_EXPLORER" {
			hasDataExplorer = true
			break
		}
	}
	if !hasDataExplorer {
		t.Error("expected at least one DATA_EXPLORER tile in Grail mode")
	}
}
