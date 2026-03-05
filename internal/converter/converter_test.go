package converter

import (
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
)

func TestConvertDashboard(t *testing.T) {
	dd := &datadog.Dashboard{
		Title: "Test Dashboard",
		Widgets: []datadog.Widget{
			{
				Definition: datadog.WidgetDefinition{
					Type:  "note",
					Title: "Hello",
					Content: "World",
				},
			},
		},
	}

	dt, err := ConvertDashboard(dd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.DashboardMetadata.Name != "Test Dashboard" {
		t.Errorf("expected name 'Test Dashboard', got %q", dt.DashboardMetadata.Name)
	}
	if len(dt.Tiles) != 1 {
		t.Errorf("expected 1 tile, got %d", len(dt.Tiles))
	}
	if dt.Tiles[0].TileType != "MARKDOWN" {
		t.Errorf("expected MARKDOWN tile, got %q", dt.Tiles[0].TileType)
	}
}

func TestConvertMonitor(t *testing.T) {
	critical := 90.0
	dd := &datadog.Monitor{
		Name:  "High CPU",
		Type:  "metric alert",
		Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
		Options: datadog.MonitorOptions{
			Thresholds: &datadog.Thresholds{
				Critical: &critical,
			},
		},
	}

	me, err := ConvertMonitor(dd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if me.Summary != "High CPU" {
		t.Errorf("expected summary 'High CPU', got %q", me.Summary)
	}
	if me.Threshold != 90.0 {
		t.Errorf("expected threshold 90, got %f", me.Threshold)
	}
	if me.EventType != "CUSTOM_ALERT" {
		t.Errorf("expected CUSTOM_ALERT event type, got %q", me.EventType)
	}
}

func TestConvertSLO(t *testing.T) {
	dd := &datadog.SLO{
		Name: "API Availability",
		Type: "metric",
		Query: &datadog.SLOQuery{
			Numerator:   "sum:api.requests.success{*}",
			Denominator: "sum:api.requests.total{*}",
		},
		Thresholds: []datadog.SLOThreshold{
			{Timeframe: "30d", Target: 99.9, Warning: 99.95},
		},
	}

	dt, err := ConvertSLO(dd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt.Name != "API Availability" {
		t.Errorf("expected name 'API Availability', got %q", dt.Name)
	}
	if dt.Target != 99.9 {
		t.Errorf("expected target 99.9, got %f", dt.Target)
	}
	if dt.Timeframe != "-1M" {
		t.Errorf("expected timeframe '-1M', got %q", dt.Timeframe)
	}
}

func TestConvertSyntheticHTTP(t *testing.T) {
	dd := &datadog.SyntheticTest{
		Name:   "Homepage Check",
		Type:   "api",
		Status: "live",
		Config: datadog.SyntheticConfig{
			Request: &datadog.SyntheticRequest{
				Method: "GET",
				URL:    "https://example.com",
			},
		},
		Options: datadog.SyntheticOptions{
			TickEvery: 300,
		},
		Locations: []string{"aws:us-east-1"},
	}

	sm, err := ConvertSynthetic(dd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm.Type != "HTTP" {
		t.Errorf("expected HTTP type, got %q", sm.Type)
	}
	if sm.FrequencyMin != 5 {
		t.Errorf("expected frequency 5 min, got %d", sm.FrequencyMin)
	}
	if !sm.Enabled {
		t.Error("expected monitor to be enabled")
	}
}

func TestConvertNotificationSlack(t *testing.T) {
	dd := &datadog.NotificationChannel{
		Name: "Team Slack",
		Type: "slack",
		Config: map[string]interface{}{
			"url":     "https://hooks.slack.com/xxx",
			"channel": "#alerts",
		},
	}

	ni, err := ConvertNotification(dd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ni.Type != "SLACK" {
		t.Errorf("expected SLACK type, got %q", ni.Type)
	}
	if ni.Config["channel"] != "#alerts" {
		t.Errorf("expected channel '#alerts', got %q", ni.Config["channel"])
	}
}

func TestConvertAllEmpty(t *testing.T) {
	c := New()
	result, errs := c.ConvertAll(&datadog.ExtractionResult{})
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if len(result.Dashboards) != 0 {
		t.Errorf("expected no dashboards, got %d", len(result.Dashboards))
	}
}
