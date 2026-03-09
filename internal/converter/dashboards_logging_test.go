package converter

import (
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
)

func TestLoggingDashboardWidgetType(t *testing.T) {
	buf := setupLogging(t)

	dd := &datadog.Dashboard{
		Title: "Test Dashboard",
		Widgets: []datadog.Widget{
			{
				Definition: datadog.WidgetDefinition{
					Type:  "timeseries",
					Title: "CPU Over Time",
					Requests: []datadog.WidgetRequest{
						{Query: "avg:system.cpu.user{*}"},
					},
				},
			},
		},
	}

	ConvertDashboard(dd, false)

	out := buf.String()
	if !strings.Contains(out, `[DEBUG] converting widget type "timeseries"`) {
		t.Errorf("expected debug log for timeseries widget type, got:\n%s", out)
	}
	if !strings.Contains(out, `"CPU Over Time"`) {
		t.Errorf("expected widget title in log, got:\n%s", out)
	}
}

func TestLoggingDashboardUnsupportedWidget(t *testing.T) {
	buf := setupLogging(t)

	dd := &datadog.Dashboard{
		Title: "Mixed Dashboard",
		Widgets: []datadog.Widget{
			{
				Definition: datadog.WidgetDefinition{
					Type:  "note",
					Title: "A Note",
					Content: "hello",
				},
			},
			{
				Definition: datadog.WidgetDefinition{
					Type:  "geomap",
					Title: "Unsupported Widget",
				},
			},
		},
	}

	ConvertDashboard(dd, false)

	out := buf.String()
	if !strings.Contains(out, `unsupported widget type "geomap"`) {
		t.Errorf("expected unsupported widget type log, got:\n%s", out)
	}
	if !strings.Contains(out, "falling back to MARKDOWN") {
		t.Errorf("expected MARKDOWN fallback log, got:\n%s", out)
	}
}
