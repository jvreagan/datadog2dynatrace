package converter

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

func setupLogging(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	logging.SetOutput(&buf)
	logging.SetLevel(logging.LevelDebug)
	t.Cleanup(func() {
		logging.SetOutput(os.Stderr)
		logging.SetLevel(logging.LevelWarn)
	})
	return &buf
}

func TestLoggingConvertAllInfoCounts(t *testing.T) {
	buf := setupLogging(t)

	ext := &datadog.ExtractionResult{
		Dashboards: []datadog.Dashboard{
			{
				Title: "Dash1",
				Widgets: []datadog.Widget{
					{Definition: datadog.WidgetDefinition{Type: "note", Title: "N", Content: "hi"}},
				},
			},
			{
				Title: "Dash2",
				Widgets: []datadog.Widget{
					{Definition: datadog.WidgetDefinition{Type: "note", Title: "N2", Content: "bye"}},
				},
			},
		},
		Monitors: []datadog.Monitor{
			{
				Name:  "Mon1",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
			},
		},
	}

	c := New(Options{})
	c.ConvertAll(ext)

	out := buf.String()

	// Info-level: resource type counts
	if !strings.Contains(out, "[INFO] converting 2 dashboards") {
		t.Errorf("expected info log for 2 dashboards, got:\n%s", out)
	}
	if !strings.Contains(out, "[INFO] converting 1 monitors") {
		t.Errorf("expected info log for 1 monitors, got:\n%s", out)
	}
}

func TestLoggingConvertAllDebugPerResource(t *testing.T) {
	buf := setupLogging(t)

	ext := &datadog.ExtractionResult{
		Dashboards: []datadog.Dashboard{
			{
				Title: "My Dashboard",
				Widgets: []datadog.Widget{
					{Definition: datadog.WidgetDefinition{Type: "note", Title: "N", Content: "hi"}},
				},
			},
		},
	}

	c := New(Options{})
	c.ConvertAll(ext)

	out := buf.String()

	if !strings.Contains(out, `[DEBUG] converting dashboard "My Dashboard"`) {
		t.Errorf("expected debug log for dashboard name, got:\n%s", out)
	}
}

func TestLoggingConvertAllWarnOnError(t *testing.T) {
	buf := setupLogging(t)

	ext := &datadog.ExtractionResult{
		Dashboards: []datadog.Dashboard{
			{
				Title:   "Bad",
				Widgets: []datadog.Widget{}, // no widgets → conversion error
			},
		},
	}

	c := New(Options{})
	c.ConvertAll(ext)

	out := buf.String()

	if !strings.Contains(out, `[WARN] dashboard "Bad"`) {
		t.Errorf("expected warn log for failed dashboard, got:\n%s", out)
	}
}
