package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportTFJSON(t *testing.T) {
	tfJSON := `{
		"resource": {
			"datadog_dashboard": {
				"main": {
					"title": "TF Dashboard",
					"widgets": []
				}
			},
			"datadog_monitor": {
				"cpu": {
					"name": "CPU Alert",
					"type": "metric alert",
					"query": "avg:system.cpu.user{*}"
				}
			}
		}
	}`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf.json"), []byte(tfJSON), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("got %d dashboards, want 1", len(result.Dashboards))
	}
	if len(result.Monitors) != 1 {
		t.Errorf("got %d monitors, want 1", len(result.Monitors))
	}
}

func TestImportTFJSONNotebooks(t *testing.T) {
	tfJSON := `{
		"resource": {
			"datadog_notebook": {
				"nb1": {
					"name": "My Notebook",
					"cells": []
				}
			}
		}
	}`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notebooks.tf.json"), []byte(tfJSON), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Notebooks) != 1 {
		t.Errorf("got %d notebooks, want 1", len(result.Notebooks))
	}
}

func TestImportTFHCL(t *testing.T) {
	hcl := `
resource "datadog_dashboard" "main" {
  title = "HCL Dashboard"
}

resource "datadog_monitor" "cpu" {
  name  = "CPU Alert"
  type  = "metric alert"
  query = "avg:system.cpu.user{*}"
}

resource "datadog_service_level_objective" "api" {
  name = "API Availability"
}

resource "datadog_synthetics_test" "health" {
  name = "Health Check"
  type = "api"
}

resource "datadog_logs_custom_pipeline" "main" {
  name = "Main Pipeline"
}

resource "datadog_downtime" "maint" {
  message = "Scheduled maintenance"
}
`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(hcl), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFromDirectory(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Dashboards) != 1 {
		t.Errorf("got %d dashboards, want 1", len(result.Dashboards))
	}
	if len(result.Monitors) != 1 {
		t.Errorf("got %d monitors, want 1", len(result.Monitors))
	}
	if len(result.SLOs) != 1 {
		t.Errorf("got %d SLOs, want 1", len(result.SLOs))
	}
	if len(result.Synthetics) != 1 {
		t.Errorf("got %d synthetics, want 1", len(result.Synthetics))
	}
	if len(result.LogPipelines) != 1 {
		t.Errorf("got %d log pipelines, want 1", len(result.LogPipelines))
	}
	if len(result.Downtimes) != 1 {
		t.Errorf("got %d downtimes, want 1", len(result.Downtimes))
	}
}

func TestExtractHCLStringField(t *testing.T) {
	lines := []string{
		`resource "datadog_monitor" "test" {`,
		`  name  = "My Monitor"`,
		`  type  = "metric alert"`,
		`  query = "avg:system.cpu.user{*}"`,
		`}`,
	}

	tests := map[string]string{
		"name":  "My Monitor",
		"type":  "metric alert",
		"query": "avg:system.cpu.user{*}",
	}

	for field, expected := range tests {
		t.Run(field, func(t *testing.T) {
			result := extractHCLStringField(lines, field)
			if result != expected {
				t.Errorf("extractHCLStringField(%q) = %q, want %q", field, result, expected)
			}
		})
	}
}
