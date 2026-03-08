package converter

import (
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ---------------------------------------------------------------------------
// Dashboard conversion tests
// ---------------------------------------------------------------------------

func TestConvertDashboard(t *testing.T) {
	tests := []struct {
		name       string
		input      *datadog.Dashboard
		wantErr    bool
		errContain string
		check      func(t *testing.T, dt *dynatrace.Dashboard)
	}{
		{
			name: "timeseries widget",
			input: &datadog.Dashboard{
				Title: "Timeseries Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "timeseries",
							Title: "CPU Usage",
							Requests: []datadog.WidgetRequest{
								{Query: "avg:system.cpu.user{*}"},
							},
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if dt.DashboardMetadata.Name != "Timeseries Dashboard" {
					t.Errorf("expected name %q, got %q", "Timeseries Dashboard", dt.DashboardMetadata.Name)
				}
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				if dt.Tiles[0].TileType != "DATA_EXPLORER" {
					t.Errorf("expected DATA_EXPLORER tile type, got %q", dt.Tiles[0].TileType)
				}
				if dt.Tiles[0].Name != "CPU Usage" {
					t.Errorf("expected tile name %q, got %q", "CPU Usage", dt.Tiles[0].Name)
				}
				if len(dt.Tiles[0].Queries) == 0 {
					t.Error("expected at least one query in tile")
				}
				if dt.Tiles[0].Queries[0].ID != "Q1" {
					t.Errorf("expected query ID %q, got %q", "Q1", dt.Tiles[0].Queries[0].ID)
				}
				if !strings.Contains(dt.Tiles[0].Queries[0].MetricSelector, "builtin:host.cpu.user") {
					t.Errorf("expected metric selector to contain builtin:host.cpu.user, got %q", dt.Tiles[0].Queries[0].MetricSelector)
				}
			},
		},
		{
			name: "query_value widget",
			input: &datadog.Dashboard{
				Title: "Query Value Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "query_value",
							Title: "Avg CPU",
							Requests: []datadog.WidgetRequest{
								{Query: "avg:system.cpu.user{*}"},
							},
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				if dt.Tiles[0].TileType != "DATA_EXPLORER" {
					t.Errorf("expected DATA_EXPLORER tile type, got %q", dt.Tiles[0].TileType)
				}
				if len(dt.Tiles[0].Queries) == 0 {
					t.Error("expected at least one query in tile")
				}
			},
		},
		{
			name: "toplist widget appends sort/limit suffix",
			input: &datadog.Dashboard{
				Title: "Toplist Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "toplist",
							Title: "Top Hosts",
							Requests: []datadog.WidgetRequest{
								{Query: "avg:system.cpu.user{*} by {host}"},
							},
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				if dt.Tiles[0].TileType != "DATA_EXPLORER" {
					t.Errorf("expected DATA_EXPLORER, got %q", dt.Tiles[0].TileType)
				}
				if len(dt.Tiles[0].Queries) == 0 {
					t.Fatal("expected queries on toplist tile")
				}
				sel := dt.Tiles[0].Queries[0].MetricSelector
				if !strings.Contains(sel, ":sort(value(avg,descending)):limit(10)") {
					t.Errorf("expected sort/limit suffix in metric selector, got %q", sel)
				}
			},
		},
		{
			name: "note/markdown widget",
			input: &datadog.Dashboard{
				Title: "Notes Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:    "note",
							Title:   "My Note",
							Content: "## Hello World\nSome markdown content.",
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				tile := dt.Tiles[0]
				if tile.TileType != "MARKDOWN" {
					t.Errorf("expected MARKDOWN, got %q", tile.TileType)
				}
				if tile.Markdown != "## Hello World\nSome markdown content." {
					t.Errorf("unexpected markdown content: %q", tile.Markdown)
				}
			},
		},
		{
			name: "free_text widget treated as note",
			input: &datadog.Dashboard{
				Title: "Free Text Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:    "free_text",
							Title:   "Free Note",
							Content: "Just text",
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				if dt.Tiles[0].TileType != "MARKDOWN" {
					t.Errorf("expected MARKDOWN, got %q", dt.Tiles[0].TileType)
				}
				if dt.Tiles[0].Markdown != "Just text" {
					t.Errorf("expected markdown %q, got %q", "Just text", dt.Tiles[0].Markdown)
				}
			},
		},
		{
			name: "group widget becomes HEADER tile",
			input: &datadog.Dashboard{
				Title: "Grouped Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "group",
							Title: "Server Metrics",
							Widgets: []datadog.Widget{
								{
									Definition: datadog.WidgetDefinition{
										Type:  "timeseries",
										Title: "Nested TS",
										Requests: []datadog.WidgetRequest{
											{Query: "avg:system.cpu.user{*}"},
										},
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 2 {
					t.Fatalf("expected 2 tiles (HEADER + nested), got %d", len(dt.Tiles))
				}
				if dt.Tiles[0].TileType != "HEADER" {
					t.Errorf("expected HEADER tile type, got %q", dt.Tiles[0].TileType)
				}
				if dt.Tiles[0].Name != "Server Metrics" {
					t.Errorf("expected tile name %q, got %q", "Server Metrics", dt.Tiles[0].Name)
				}
				if dt.Tiles[1].TileType != "DATA_EXPLORER" {
					t.Errorf("expected nested tile type DATA_EXPLORER, got %q", dt.Tiles[1].TileType)
				}
			},
		},
		{
			name: "unsupported widget type gets markdown fallback",
			input: &datadog.Dashboard{
				Title: "Unsupported Widgets",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "iframe",
							Title: "My iFrame",
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				tile := dt.Tiles[0]
				if tile.TileType != "MARKDOWN" {
					t.Errorf("expected MARKDOWN fallback, got %q", tile.TileType)
				}
				if !strings.Contains(tile.Markdown, "iframe") {
					t.Errorf("expected markdown to mention original widget type, got %q", tile.Markdown)
				}
				if !strings.Contains(tile.Markdown, "Manual configuration required") {
					t.Errorf("expected migration note in markdown, got %q", tile.Markdown)
				}
			},
		},
		{
			name: "empty dashboard (no convertible widgets) returns error",
			input: &datadog.Dashboard{
				Title:   "Empty Dashboard",
				Widgets: []datadog.Widget{},
			},
			wantErr:    true,
			errContain: "no widgets could be converted",
		},
		{
			name: "timeseries with no parseable query is skipped, resulting in error if sole widget",
			input: &datadog.Dashboard{
				Title: "Bad Query Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:     "timeseries",
							Title:    "Bad TS",
							Requests: []datadog.WidgetRequest{},
						},
					},
				},
			},
			wantErr:    true,
			errContain: "no widgets could be converted",
		},
		{
			name: "multiple mixed widgets",
			input: &datadog.Dashboard{
				Title: "Mixed Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:    "note",
							Title:   "Heading",
							Content: "Welcome",
						},
					},
					{
						Definition: datadog.WidgetDefinition{
							Type:  "timeseries",
							Title: "CPU",
							Requests: []datadog.WidgetRequest{
								{Query: "avg:system.cpu.user{*}"},
							},
						},
					},
					{
						Definition: datadog.WidgetDefinition{
							Type:  "hostmap",
							Title: "Host Map",
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 3 {
					t.Fatalf("expected 3 tiles, got %d", len(dt.Tiles))
				}
				tileTypes := make([]string, len(dt.Tiles))
				for i, tile := range dt.Tiles {
					tileTypes[i] = tile.TileType
				}
				if tileTypes[0] != "MARKDOWN" {
					t.Errorf("tile 0: expected MARKDOWN, got %q", tileTypes[0])
				}
				if tileTypes[1] != "DATA_EXPLORER" {
					t.Errorf("tile 1: expected DATA_EXPLORER, got %q", tileTypes[1])
				}
				if tileTypes[2] != "HOSTS" {
					t.Errorf("tile 2: expected HOSTS, got %q", tileTypes[2])
				}
			},
		},
		{
			name: "layout bounds are calculated from widget layout",
			input: &datadog.Dashboard{
				Title: "Layout Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:    "note",
							Title:   "Positioned",
							Content: "hello",
						},
						Layout: &datadog.WidgetLayout{
							X:      2,
							Y:      3,
							Width:  4,
							Height: 2,
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				b := dt.Tiles[0].Bounds
				if b.Left != 2*76 {
					t.Errorf("expected left %d, got %d", 2*76, b.Left)
				}
				if b.Top != 3*38 {
					t.Errorf("expected top %d, got %d", 3*38, b.Top)
				}
				if b.Width != 4*76 {
					t.Errorf("expected width %d, got %d", 4*76, b.Width)
				}
				if b.Height != 2*38 {
					t.Errorf("expected height %d, got %d", 2*38, b.Height)
				}
			},
		},
		{
			name: "dashboard metadata owner is set",
			input: &datadog.Dashboard{
				Title: "Owner Test",
				Widgets: []datadog.Widget{
					{Definition: datadog.WidgetDefinition{Type: "note", Title: "x", Content: "y"}},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if dt.DashboardMetadata.Owner != "datadog2dynatrace" {
					t.Errorf("expected owner %q, got %q", "datadog2dynatrace", dt.DashboardMetadata.Owner)
				}
			},
		},
		{
			name: "timeseries widget with queries/formulas format",
			input: &datadog.Dashboard{
				Title: "New Format Dashboard",
				Widgets: []datadog.Widget{
					{
						Definition: datadog.WidgetDefinition{
							Type:  "timeseries",
							Title: "New Style Query",
							Requests: []datadog.WidgetRequest{
								{
									Queries: []datadog.QueryDef{
										{Name: "a", DataSource: "metrics", Query: "avg:system.mem.used{*}"},
									},
									Formulas: []datadog.Formula{
										{Formula: "a", Alias: "Memory Used"},
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, dt *dynatrace.Dashboard) {
				if len(dt.Tiles) != 1 {
					t.Fatalf("expected 1 tile, got %d", len(dt.Tiles))
				}
				if len(dt.Tiles[0].Queries) == 0 {
					t.Fatal("expected at least one query")
				}
				sel := dt.Tiles[0].Queries[0].MetricSelector
				if !strings.Contains(sel, "builtin:host.mem.usage") {
					t.Errorf("expected builtin:host.mem.usage in selector, got %q", sel)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := ConvertDashboard(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, dt)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Monitor conversion tests
// ---------------------------------------------------------------------------

func TestConvertMonitor(t *testing.T) {
	critical := 90.0
	warning := 80.0
	critical50 := 50.0
	logThreshold := 5.0

	tests := []struct {
		name       string
		input      *datadog.Monitor
		wantErr    bool
		errContain string
		check      func(t *testing.T, me *dynatrace.MetricEvent)
	}{
		{
			name: "metric alert with critical threshold",
			input: &datadog.Monitor{
				Name:  "High CPU",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if me.Summary != "High CPU" {
					t.Errorf("expected summary %q, got %q", "High CPU", me.Summary)
				}
				if me.Threshold != 90.0 {
					t.Errorf("expected threshold 90, got %f", me.Threshold)
				}
				if me.EventType != "CUSTOM_ALERT" {
					t.Errorf("expected CUSTOM_ALERT, got %q", me.EventType)
				}
				if me.AlertCondition != "ABOVE" {
					t.Errorf("expected ABOVE alert condition, got %q", me.AlertCondition)
				}
				if !me.Enabled {
					t.Error("expected monitor to be enabled")
				}
				if !strings.Contains(me.MetricSelector, "builtin:host.cpu.user") {
					t.Errorf("expected metric selector to contain builtin:host.cpu.user, got %q", me.MetricSelector)
				}
				if me.MonitoringStrategy.Type != "STATIC_THRESHOLD" {
					t.Errorf("expected STATIC_THRESHOLD strategy, got %q", me.MonitoringStrategy.Type)
				}
				if me.MonitoringStrategy.Samples != 5 {
					t.Errorf("expected 5 samples, got %d", me.MonitoringStrategy.Samples)
				}
				if me.MonitoringStrategy.ViolatingSamples != 3 {
					t.Errorf("expected 3 violating samples, got %d", me.MonitoringStrategy.ViolatingSamples)
				}
			},
		},
		{
			name: "metric alert with warning threshold (critical still used)",
			input: &datadog.Monitor{
				Name:  "CPU Warning",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 80",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
						Warning:  &warning,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if me.Threshold != 90.0 {
					t.Errorf("expected critical threshold 90, got %f", me.Threshold)
				}
			},
		},
		{
			name: "service check maps to ERROR event type",
			input: &datadog.Monitor{
				Name:  "HTTP Check Failed",
				Type:  "service check",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 50",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical50,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if me.EventType != "ERROR" {
					t.Errorf("expected ERROR event type for service check, got %q", me.EventType)
				}
			},
		},
		{
			name: "log alert maps to CUSTOM_ALERT",
			input: &datadog.Monitor{
				Name:  "Log Spike",
				Type:  "log alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 5",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &logThreshold,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if me.EventType != "CUSTOM_ALERT" {
					t.Errorf("expected CUSTOM_ALERT for log alert, got %q", me.EventType)
				}
				if me.Threshold != 5.0 {
					t.Errorf("expected threshold 5, got %f", me.Threshold)
				}
			},
		},
		{
			name: "missing thresholds uses parsed threshold",
			input: &datadog.Monitor{
				Name:    "No Threshold Options",
				Type:    "metric alert",
				Query:   "avg(last_5m):avg:system.cpu.user{*} > 90",
				Options: datadog.MonitorOptions{},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				// Threshold defaults to 0 when not parsed from query comparison
				// and no options thresholds exist
				if !strings.Contains(me.MetricSelector, "builtin:host.cpu.user") {
					t.Errorf("expected builtin:host.cpu.user in metric selector, got %q", me.MetricSelector)
				}
			},
		},
		{
			name: "message sanitization strips @mentions",
			input: &datadog.Monitor{
				Name:  "Alert with Mentions",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
				Message: "CPU is too high!\n@slack-team-alerts\n@pagerduty-oncall\nPlease investigate.",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if strings.Contains(me.Description, "@slack") {
					t.Errorf("expected @slack mention to be stripped, got %q", me.Description)
				}
				if strings.Contains(me.Description, "@pagerduty") {
					t.Errorf("expected @pagerduty mention to be stripped, got %q", me.Description)
				}
				if !strings.Contains(me.Description, "CPU is too high!") {
					t.Errorf("expected message body to be preserved, got %q", me.Description)
				}
				if !strings.Contains(me.Description, "Please investigate.") {
					t.Errorf("expected remaining text to be preserved, got %q", me.Description)
				}
			},
		},
		{
			name: "monitor with tags are mapped",
			input: &datadog.Monitor{
				Name:  "Tagged Monitor",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
				Tags:  []string{"env:prod", "team:platform", "severity"},
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if len(me.Tags) != 3 {
					t.Fatalf("expected 3 tags, got %d", len(me.Tags))
				}
				if me.Tags[0].Key != "env" || me.Tags[0].Value != "prod" {
					t.Errorf("expected env:prod tag, got %q:%q", me.Tags[0].Key, me.Tags[0].Value)
				}
				// Tag without value
				found := false
				for _, tag := range me.Tags {
					if tag.Key == "severity" && tag.Value == "" {
						found = true
					}
				}
				if !found {
					t.Error("expected tag 'severity' with empty value")
				}
			},
		},
		{
			name: "below threshold alert condition",
			input: &datadog.Monitor{
				Name:  "Low Disk",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.disk.free{*} < 10",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if me.AlertCondition != "BELOW" {
					t.Errorf("expected BELOW alert condition, got %q", me.AlertCondition)
				}
			},
		},
		{
			name: "query with filters",
			input: &datadog.Monitor{
				Name:  "CPU per host",
				Type:  "metric alert",
				Query: "avg(last_5m):avg:system.cpu.user{host:web01} > 90",
				Options: datadog.MonitorOptions{
					Thresholds: &datadog.Thresholds{
						Critical: &critical,
					},
				},
			},
			check: func(t *testing.T, me *dynatrace.MetricEvent) {
				if !strings.Contains(me.MetricSelector, "filter") {
					t.Errorf("expected filter in metric selector, got %q", me.MetricSelector)
				}
				if !strings.Contains(me.MetricSelector, "dt.entity.host") {
					t.Errorf("expected dt.entity.host in filter, got %q", me.MetricSelector)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			me, err := ConvertMonitor(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, me)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SLO conversion tests
// ---------------------------------------------------------------------------

func TestConvertSLO(t *testing.T) {
	tests := []struct {
		name       string
		input      *datadog.SLO
		wantErr    bool
		errContain string
		check      func(t *testing.T, dt *dynatrace.SLO)
	}{
		{
			name: "metric SLO with 30d threshold",
			input: &datadog.SLO{
				Name: "API Availability",
				Type: "metric",
				Query: &datadog.SLOQuery{
					Numerator:   "sum:api.requests.success{*}",
					Denominator: "sum:api.requests.total{*}",
				},
				Thresholds: []datadog.SLOThreshold{
					{Timeframe: "30d", Target: 99.9, Warning: 99.95},
				},
			},
			check: func(t *testing.T, dt *dynatrace.SLO) {
				if dt.Name != "API Availability" {
					t.Errorf("expected name %q, got %q", "API Availability", dt.Name)
				}
				if dt.Target != 99.9 {
					t.Errorf("expected target 99.9, got %f", dt.Target)
				}
				if dt.Warning != 99.95 {
					t.Errorf("expected warning 99.95, got %f", dt.Warning)
				}
				if dt.Timeframe != "-1M" {
					t.Errorf("expected timeframe -1M, got %q", dt.Timeframe)
				}
				if dt.EvaluationType != "AGGREGATE" {
					t.Errorf("expected AGGREGATE, got %q", dt.EvaluationType)
				}
				if !dt.Enabled {
					t.Error("expected SLO to be enabled")
				}
				if dt.MetricExpression == "" {
					t.Error("expected non-empty metric expression")
				}
				if !strings.Contains(dt.MetricExpression, "(100)") {
					t.Errorf("expected metric expression to contain (100) for percentage, got %q", dt.MetricExpression)
				}
			},
		},
		{
			name: "metric SLO with 7d threshold",
			input: &datadog.SLO{
				Name: "Weekly SLO",
				Type: "metric",
				Query: &datadog.SLOQuery{
					Numerator:   "sum:requests.ok{*}",
					Denominator: "sum:requests.total{*}",
				},
				Thresholds: []datadog.SLOThreshold{
					{Timeframe: "7d", Target: 99.0, Warning: 99.5},
				},
			},
			check: func(t *testing.T, dt *dynatrace.SLO) {
				if dt.Timeframe != "-1w" {
					t.Errorf("expected timeframe -1w for 7d, got %q", dt.Timeframe)
				}
				if dt.Target != 99.0 {
					t.Errorf("expected target 99.0, got %f", dt.Target)
				}
			},
		},
		{
			name: "metric SLO with 90d threshold",
			input: &datadog.SLO{
				Name: "Quarterly SLO",
				Type: "metric",
				Query: &datadog.SLOQuery{
					Numerator:   "sum:requests.ok{*}",
					Denominator: "sum:requests.total{*}",
				},
				Thresholds: []datadog.SLOThreshold{
					{Timeframe: "90d", Target: 99.5},
				},
			},
			check: func(t *testing.T, dt *dynatrace.SLO) {
				if dt.Timeframe != "-3M" {
					t.Errorf("expected timeframe -3M for 90d, got %q", dt.Timeframe)
				}
			},
		},
		{
			name: "monitor-based SLO gets placeholder expression",
			input: &datadog.SLO{
				Name:        "Monitor SLO",
				Type:        "monitor",
				Description: "Based on monitors",
				MonitorIDs:  []int64{123, 456},
				Thresholds: []datadog.SLOThreshold{
					{Timeframe: "30d", Target: 99.0},
				},
			},
			check: func(t *testing.T, dt *dynatrace.SLO) {
				if !strings.Contains(dt.MetricExpression, "builtin:synthetic") {
					t.Errorf("expected placeholder metric expression for monitor SLO, got %q", dt.MetricExpression)
				}
				if !strings.Contains(dt.Description, "monitor-based SLO") {
					t.Errorf("expected migration note in description, got %q", dt.Description)
				}
			},
		},
		{
			name: "SLO with no thresholds gets defaults",
			input: &datadog.SLO{
				Name: "Default SLO",
				Type: "metric",
				Query: &datadog.SLOQuery{
					Numerator:   "sum:requests.ok{*}",
					Denominator: "sum:requests.total{*}",
				},
				Thresholds: []datadog.SLOThreshold{},
			},
			check: func(t *testing.T, dt *dynatrace.SLO) {
				if dt.Target != 99.0 {
					t.Errorf("expected default target 99.0, got %f", dt.Target)
				}
				if dt.Warning != 99.5 {
					t.Errorf("expected default warning 99.5, got %f", dt.Warning)
				}
				if dt.Timeframe != "-1M" {
					t.Errorf("expected default timeframe -1M, got %q", dt.Timeframe)
				}
			},
		},
		{
			name: "unsupported SLO type returns error",
			input: &datadog.SLO{
				Name: "Bad SLO",
				Type: "unknown_type",
				Thresholds: []datadog.SLOThreshold{
					{Timeframe: "30d", Target: 99.0},
				},
			},
			wantErr:    true,
			errContain: "unsupported SLO type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := ConvertSLO(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, dt)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Synthetic conversion tests
// ---------------------------------------------------------------------------

func TestConvertSynthetic(t *testing.T) {
	tests := []struct {
		name       string
		input      *datadog.SyntheticTest
		wantErr    bool
		errContain string
		check      func(t *testing.T, sm *dynatrace.SyntheticMonitor)
	}{
		{
			name: "API/HTTP test basic",
			input: &datadog.SyntheticTest{
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
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.Type != "HTTP" {
					t.Errorf("expected HTTP type, got %q", sm.Type)
				}
				if sm.Name != "Homepage Check" {
					t.Errorf("expected name %q, got %q", "Homepage Check", sm.Name)
				}
				if sm.FrequencyMin != 5 {
					t.Errorf("expected frequency 5 min, got %d", sm.FrequencyMin)
				}
				if !sm.Enabled {
					t.Error("expected monitor to be enabled")
				}
				if sm.Script == nil {
					t.Fatal("expected script to be set")
				}
				if len(sm.Script.Requests) != 1 {
					t.Fatalf("expected 1 script request, got %d", len(sm.Script.Requests))
				}
				if sm.Script.Requests[0].URL != "https://example.com" {
					t.Errorf("expected URL %q, got %q", "https://example.com", sm.Script.Requests[0].URL)
				}
				if sm.Script.Requests[0].Method != "GET" {
					t.Errorf("expected method %q, got %q", "GET", sm.Script.Requests[0].Method)
				}
				if sm.AnomalyDetection == nil {
					t.Fatal("expected anomaly detection to be set")
				}
				if !sm.AnomalyDetection.OutageHandling.GlobalOutage {
					t.Error("expected global outage to be true")
				}
			},
		},
		{
			name: "API test with assertions",
			input: &datadog.SyntheticTest{
				Name:   "API Assert",
				Type:   "api",
				Status: "live",
				Config: datadog.SyntheticConfig{
					Request: &datadog.SyntheticRequest{
						Method: "POST",
						URL:    "https://api.example.com/health",
						Body:   `{"check":"health"}`,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
					},
					Assertions: []datadog.SyntheticAssertion{
						{Type: "statusCode", Operator: "is", Target: 200},
						{Type: "body", Operator: "contains", Target: "ok"},
					},
				},
				Options: datadog.SyntheticOptions{
					TickEvery:       600,
					FollowRedirects: true,
				},
				Locations: []string{"aws:us-east-1", "aws:eu-west-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.Type != "HTTP" {
					t.Errorf("expected HTTP, got %q", sm.Type)
				}
				if sm.FrequencyMin != 10 {
					t.Errorf("expected frequency 10 min for 600s tick, got %d", sm.FrequencyMin)
				}
				if sm.Script == nil || len(sm.Script.Requests) != 1 {
					t.Fatal("expected 1 script request")
				}
				req := sm.Script.Requests[0]
				if req.RequestBody != `{"check":"health"}` {
					t.Errorf("expected request body, got %q", req.RequestBody)
				}
				if req.Configuration == nil {
					t.Fatal("expected configuration for headers")
				}
				if req.Configuration.Headers["Content-Type"] != "application/json" {
					t.Errorf("expected Content-Type header, got %v", req.Configuration.Headers)
				}
				if !req.Configuration.FollowRedirects {
					t.Error("expected follow redirects to be true")
				}
				if req.Validation == nil || len(req.Validation.Rules) != 2 {
					t.Fatalf("expected 2 validation rules, got %v", req.Validation)
				}
				if req.Validation.Rules[0].Type != "httpStatusesList" {
					t.Errorf("expected httpStatusesList rule, got %q", req.Validation.Rules[0].Type)
				}
				// Check both locations mapped
				if len(sm.Locations) != 2 {
					t.Errorf("expected 2 locations, got %d", len(sm.Locations))
				}
			},
		},
		{
			name: "browser test",
			input: &datadog.SyntheticTest{
				Name:   "Browser Flow",
				Type:   "browser",
				Status: "live",
				Config: datadog.SyntheticConfig{
					Request: &datadog.SyntheticRequest{
						URL: "https://app.example.com",
					},
					Steps: []datadog.BrowserStep{
						{Name: "Click Login", Type: "click", Params: map[string]interface{}{"element": "#login-btn"}},
						{Name: "Type Username", Type: "typeText", Params: map[string]interface{}{"element": "#username"}},
					},
				},
				Options: datadog.SyntheticOptions{
					TickEvery: 900,
				},
				Locations: []string{"aws:us-west-2"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.Type != "BROWSER" {
					t.Errorf("expected BROWSER type, got %q", sm.Type)
				}
				if sm.FrequencyMin != 15 {
					t.Errorf("expected frequency 15 min for 900s tick, got %d", sm.FrequencyMin)
				}
				if sm.KeyPerformanceMetrics == nil {
					t.Fatal("expected key performance metrics to be set")
				}
				if sm.KeyPerformanceMetrics.LoadActionKPM != "VISUALLY_COMPLETE" {
					t.Errorf("expected VISUALLY_COMPLETE, got %q", sm.KeyPerformanceMetrics.LoadActionKPM)
				}
				if sm.Script == nil {
					t.Fatal("expected script to be set")
				}
				// navigate + click + keystrokes = 3 events
				if len(sm.Script.Events) != 3 {
					t.Fatalf("expected 3 script events, got %d", len(sm.Script.Events))
				}
				if sm.Script.Events[0].Type != "navigate" {
					t.Errorf("expected first event to be navigate, got %q", sm.Script.Events[0].Type)
				}
				if sm.Script.Events[1].Type != "click" {
					t.Errorf("expected second event to be click, got %q", sm.Script.Events[1].Type)
				}
				if sm.Script.Events[2].Type != "keystrokes" {
					t.Errorf("expected third event to be keystrokes, got %q", sm.Script.Events[2].Type)
				}
			},
		},
		{
			name: "missing request config returns error for API test",
			input: &datadog.SyntheticTest{
				Name:   "No Config",
				Type:   "api",
				Status: "live",
				Config: datadog.SyntheticConfig{
					Request: nil,
				},
				Options:   datadog.SyntheticOptions{TickEvery: 300},
				Locations: []string{"aws:us-east-1"},
			},
			wantErr:    true,
			errContain: "no request configuration",
		},
		{
			name: "frequency mapping: 60s -> 1 min",
			input: &datadog.SyntheticTest{
				Name: "Fast Check", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://fast.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 60},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.FrequencyMin != 1 {
					t.Errorf("expected 1 min for 60s, got %d", sm.FrequencyMin)
				}
			},
		},
		{
			name: "frequency mapping: 120s -> 2 min",
			input: &datadog.SyntheticTest{
				Name: "2min Check", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://test.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 120},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.FrequencyMin != 2 {
					t.Errorf("expected 2 min for 120s, got %d", sm.FrequencyMin)
				}
			},
		},
		{
			name: "frequency mapping: 0s -> default 15 min",
			input: &datadog.SyntheticTest{
				Name: "Default Freq", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://def.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 0},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.FrequencyMin != 15 {
					t.Errorf("expected default 15 min for 0s, got %d", sm.FrequencyMin)
				}
			},
		},
		{
			name: "frequency mapping: 3600s -> 60 min",
			input: &datadog.SyntheticTest{
				Name: "Hourly Check", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://hourly.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 3600},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.FrequencyMin != 60 {
					t.Errorf("expected 60 min for 3600s, got %d", sm.FrequencyMin)
				}
			},
		},
		{
			name: "location mapping: known locations mapped",
			input: &datadog.SyntheticTest{
				Name: "Location Test", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://loc.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 300},
				Locations: []string{"aws:us-east-1", "aws:eu-central-1", "aws:ap-northeast-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if len(sm.Locations) != 3 {
					t.Errorf("expected 3 locations, got %d", len(sm.Locations))
				}
				for _, loc := range sm.Locations {
					if !strings.HasPrefix(loc, "GEOLOCATION-") {
						t.Errorf("expected GEOLOCATION- prefix, got %q", loc)
					}
				}
			},
		},
		{
			name: "location mapping: unknown location gets default",
			input: &datadog.SyntheticTest{
				Name: "Unknown Loc", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://unk.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 300},
				Locations: []string{"gcp:us-central1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if len(sm.Locations) != 1 {
					t.Fatalf("expected 1 default location, got %d", len(sm.Locations))
				}
				if sm.Locations[0] != "GEOLOCATION-0A41430434C388A9" {
					t.Errorf("expected N. Virginia default, got %q", sm.Locations[0])
				}
			},
		},
		{
			name: "paused synthetic test is disabled",
			input: &datadog.SyntheticTest{
				Name: "Paused Test", Type: "api", Status: "paused",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://paused.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 300},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.Enabled {
					t.Error("expected paused test to have Enabled=false")
				}
			},
		},
		{
			name: "unsupported synthetic type returns error",
			input: &datadog.SyntheticTest{
				Name:      "Grpc Test",
				Type:      "grpc",
				Status:    "live",
				Locations: []string{"aws:us-east-1"},
			},
			wantErr:    true,
			errContain: "unsupported synthetic test type",
		},
		{
			name: "API test with retry config",
			input: &datadog.SyntheticTest{
				Name: "Retry Check", Type: "api", Status: "live",
				Config: datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://retry.test"}},
				Options: datadog.SyntheticOptions{
					TickEvery: 300,
					Retry:     &datadog.RetryConfig{Count: 2, Interval: 500},
				},
				Locations: []string{"aws:us-east-1"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if sm.AnomalyDetection == nil || sm.AnomalyDetection.OutageHandling == nil {
					t.Fatal("expected anomaly detection outage handling")
				}
				if !sm.AnomalyDetection.OutageHandling.RetryOnError {
					t.Error("expected RetryOnError to be true when retry config count > 0")
				}
			},
		},
		{
			name: "tags are mapped to synthetic tags",
			input: &datadog.SyntheticTest{
				Name: "Tagged Synth", Type: "api", Status: "live",
				Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://tagged.test"}},
				Options:   datadog.SyntheticOptions{TickEvery: 300},
				Locations: []string{"aws:us-east-1"},
				Tags:      []string{"env:staging", "team:qa"},
			},
			check: func(t *testing.T, sm *dynatrace.SyntheticMonitor) {
				if len(sm.Tags) != 2 {
					t.Fatalf("expected 2 tags, got %d", len(sm.Tags))
				}
				if sm.Tags[0].Key != "env" || sm.Tags[0].Value != "staging" {
					t.Errorf("expected env:staging, got %q:%q", sm.Tags[0].Key, sm.Tags[0].Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := ConvertSynthetic(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, sm)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Log pipeline conversion tests
// ---------------------------------------------------------------------------

func TestConvertLogPipeline(t *testing.T) {
	tests := []struct {
		name  string
		input *datadog.LogPipeline
		check func(t *testing.T, rule *dynatrace.LogProcessingRule)
	}{
		{
			name: "grok parser processor",
			input: &datadog.LogPipeline{
				Name:      "Web Logs",
				IsEnabled: true,
				Filter:    &datadog.LogFilter{Query: "source:nginx"},
				Processors: []datadog.LogProcessor{
					{
						Type:      "grok-parser",
						Name:      "Parse Access Log",
						IsEnabled: true,
						Grok: &datadog.GrokRule{
							MatchRules:   `%{COMBINEDAPACHELOG}`,
							SupportRules: "",
						},
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if rule.Name != "Web Logs" {
					t.Errorf("expected name %q, got %q", "Web Logs", rule.Name)
				}
				if !rule.Enabled {
					t.Error("expected rule to be enabled")
				}
				if !strings.Contains(rule.Query, "dt.source.name") {
					t.Errorf("expected translated filter with dt.source.name, got %q", rule.Query)
				}
				if !strings.Contains(rule.Processor, "PARSE") {
					t.Errorf("expected PARSE in processor, got %q", rule.Processor)
				}
				if !strings.Contains(rule.Processor, "COMBINEDAPACHELOG") {
					t.Errorf("expected grok pattern in processor, got %q", rule.Processor)
				}
			},
		},
		{
			name: "attribute remapper processor",
			input: &datadog.LogPipeline{
				Name:      "Remap Pipeline",
				IsEnabled: true,
				Processors: []datadog.LogProcessor{
					{
						Type:      "attribute-remapper",
						Name:      "Rename hostname",
						IsEnabled: true,
						Sources:   []string{"hostname"},
						Target:    "host",
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if !strings.Contains(rule.Processor, "FIELDS_RENAME") {
					t.Errorf("expected FIELDS_RENAME in processor, got %q", rule.Processor)
				}
				if !strings.Contains(rule.Processor, "hostname") {
					t.Errorf("expected source field in processor, got %q", rule.Processor)
				}
				if !strings.Contains(rule.Processor, "host") {
					t.Errorf("expected target field in processor, got %q", rule.Processor)
				}
			},
		},
		{
			name: "disabled processor is skipped",
			input: &datadog.LogPipeline{
				Name:      "Mixed Pipeline",
				IsEnabled: true,
				Processors: []datadog.LogProcessor{
					{
						Type:      "grok-parser",
						Name:      "Disabled Parser",
						IsEnabled: false,
						Grok: &datadog.GrokRule{
							MatchRules: `%{GREEDYDATA}`,
						},
					},
					{
						Type:      "attribute-remapper",
						Name:      "Active Remapper",
						IsEnabled: true,
						Sources:   []string{"src"},
						Target:    "dest",
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if strings.Contains(rule.Processor, "PARSE") {
					t.Errorf("disabled grok parser should be skipped, got %q", rule.Processor)
				}
				if !strings.Contains(rule.Processor, "FIELDS_RENAME") {
					t.Errorf("expected active remapper in processor, got %q", rule.Processor)
				}
			},
		},
		{
			name: "empty pipeline (no processors)",
			input: &datadog.LogPipeline{
				Name:       "Empty Pipeline",
				IsEnabled:  true,
				Filter:     &datadog.LogFilter{Query: "service:api"},
				Processors: []datadog.LogProcessor{},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if rule.Processor != "" {
					t.Errorf("expected empty processor, got %q", rule.Processor)
				}
				if !strings.Contains(rule.Query, "dt.entity.service") {
					t.Errorf("expected translated service filter, got %q", rule.Query)
				}
			},
		},
		{
			name: "pipeline with no filter defaults to wildcard query",
			input: &datadog.LogPipeline{
				Name:      "No Filter",
				IsEnabled: true,
				Filter:    nil,
				Processors: []datadog.LogProcessor{
					{
						Type:      "status-remapper",
						Name:      "Set Status",
						IsEnabled: true,
						Sources:   []string{"level"},
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if rule.Query != "*" {
					t.Errorf("expected wildcard query for nil filter, got %q", rule.Query)
				}
			},
		},
		{
			name: "date remapper processor",
			input: &datadog.LogPipeline{
				Name:      "Date Remap",
				IsEnabled: true,
				Processors: []datadog.LogProcessor{
					{
						Type:      "date-remapper",
						Name:      "Set Date",
						IsEnabled: true,
						Sources:   []string{"event_time"},
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if !strings.Contains(rule.Processor, "FIELDS_RENAME(event_time, timestamp)") {
					t.Errorf("expected date remapper to timestamp, got %q", rule.Processor)
				}
			},
		},
		{
			name: "message remapper processor",
			input: &datadog.LogPipeline{
				Name:      "Message Remap",
				IsEnabled: true,
				Processors: []datadog.LogProcessor{
					{
						Type:      "message-remapper",
						Name:      "Set Message",
						IsEnabled: true,
						Sources:   []string{"msg"},
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if !strings.Contains(rule.Processor, "FIELDS_RENAME(msg, content)") {
					t.Errorf("expected message remapper to content, got %q", rule.Processor)
				}
			},
		},
		{
			name: "multiple processors joined with pipe",
			input: &datadog.LogPipeline{
				Name:      "Multi Processor",
				IsEnabled: true,
				Processors: []datadog.LogProcessor{
					{
						Type:      "attribute-remapper",
						Name:      "Remap A",
						IsEnabled: true,
						Sources:   []string{"a"},
						Target:    "b",
					},
					{
						Type:      "attribute-remapper",
						Name:      "Remap C",
						IsEnabled: true,
						Sources:   []string{"c"},
						Target:    "d",
					},
				},
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if !strings.Contains(rule.Processor, "\n| ") {
					t.Errorf("expected pipe-joined processors, got %q", rule.Processor)
				}
				parts := strings.Split(rule.Processor, "\n| ")
				if len(parts) != 2 {
					t.Errorf("expected 2 processor parts, got %d", len(parts))
				}
			},
		},
		{
			name: "disabled pipeline is still converted (enabled=false)",
			input: &datadog.LogPipeline{
				Name:      "Disabled Pipeline",
				IsEnabled: false,
			},
			check: func(t *testing.T, rule *dynatrace.LogProcessingRule) {
				if rule.Enabled {
					t.Error("expected disabled pipeline to produce disabled rule")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := ConvertLogPipeline(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, rule)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Metric metadata conversion tests
// ---------------------------------------------------------------------------

func TestConvertMetricMetadata(t *testing.T) {
	tests := []struct {
		name  string
		input *datadog.MetricMetadata
		check func(t *testing.T, md *dynatrace.MetricDescriptor)
	}{
		{
			name: "byte unit",
			input: &datadog.MetricMetadata{
				Metric:      "system.mem.used",
				Description: "Memory used",
				Unit:        "byte",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Byte" {
					t.Errorf("expected unit Byte, got %q", md.Unit)
				}
				if !strings.Contains(md.MetricID, "builtin:host.mem.usage") {
					t.Errorf("expected translated metric ID, got %q", md.MetricID)
				}
				if md.Description != "Memory used" {
					t.Errorf("expected description %q, got %q", "Memory used", md.Description)
				}
			},
		},
		{
			name: "percent unit",
			input: &datadog.MetricMetadata{
				Metric: "system.cpu.user",
				Unit:   "percent",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Percent" {
					t.Errorf("expected unit Percent, got %q", md.Unit)
				}
			},
		},
		{
			name: "second unit",
			input: &datadog.MetricMetadata{
				Metric: "trace.servlet.request.duration",
				Unit:   "second",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Second" {
					t.Errorf("expected unit Second, got %q", md.Unit)
				}
			},
		},
		{
			name: "millisecond unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.latency",
				Unit:   "millisecond",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "MilliSecond" {
					t.Errorf("expected unit MilliSecond, got %q", md.Unit)
				}
			},
		},
		{
			name: "count unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.requests",
				Unit:   "count",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Count" {
					t.Errorf("expected unit Count, got %q", md.Unit)
				}
			},
		},
		{
			name: "request unit maps to Count",
			input: &datadog.MetricMetadata{
				Metric: "custom.req_count",
				Unit:   "request",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Count" {
					t.Errorf("expected request -> Count, got %q", md.Unit)
				}
			},
		},
		{
			name: "error unit maps to Count",
			input: &datadog.MetricMetadata{
				Metric: "custom.error_count",
				Unit:   "error",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Count" {
					t.Errorf("expected error -> Count, got %q", md.Unit)
				}
			},
		},
		{
			name: "kilobyte unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.kb_metric",
				Unit:   "kilobyte",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "KiloByte" {
					t.Errorf("expected KiloByte, got %q", md.Unit)
				}
			},
		},
		{
			name: "megabyte unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.mb_metric",
				Unit:   "megabyte",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "MegaByte" {
					t.Errorf("expected MegaByte, got %q", md.Unit)
				}
			},
		},
		{
			name: "gigabyte unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.gb_metric",
				Unit:   "gigabyte",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "GigaByte" {
					t.Errorf("expected GigaByte, got %q", md.Unit)
				}
			},
		},
		{
			name: "nanosecond unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.ns_metric",
				Unit:   "nanosecond",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "NanoSecond" {
					t.Errorf("expected NanoSecond, got %q", md.Unit)
				}
			},
		},
		{
			name: "bit unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.bit_metric",
				Unit:   "bit",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Bit" {
					t.Errorf("expected Bit, got %q", md.Unit)
				}
			},
		},
		{
			name: "unknown unit defaults to Unspecified",
			input: &datadog.MetricMetadata{
				Metric: "custom.weird",
				Unit:   "frobnicators",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Unspecified" {
					t.Errorf("expected Unspecified for unknown unit, got %q", md.Unit)
				}
			},
		},
		{
			name: "unit with per_unit",
			input: &datadog.MetricMetadata{
				Metric:  "custom.rate",
				Unit:    "byte",
				PerUnit: "second",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "BytePerSecond" {
					t.Errorf("expected BytePerSecond, got %q", md.Unit)
				}
			},
		},
		{
			name: "count per second",
			input: &datadog.MetricMetadata{
				Metric:  "custom.ops_per_sec",
				Unit:    "operation",
				PerUnit: "second",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "CountPerSecond" {
					t.Errorf("expected CountPerSecond, got %q", md.Unit)
				}
			},
		},
		{
			name: "short_name used as display name",
			input: &datadog.MetricMetadata{
				Metric:    "system.cpu.user",
				ShortName: "CPU User",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.DisplayName != "CPU User" {
					t.Errorf("expected display name %q, got %q", "CPU User", md.DisplayName)
				}
			},
		},
		{
			name: "metric name used as display name when short_name is empty",
			input: &datadog.MetricMetadata{
				Metric: "custom.my_metric",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.DisplayName != "custom.my_metric" {
					t.Errorf("expected display name %q, got %q", "custom.my_metric", md.DisplayName)
				}
			},
		},
		{
			name: "custom metric gets ext: prefix",
			input: &datadog.MetricMetadata{
				Metric: "custom.my_app.latency",
				Unit:   "millisecond",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if !strings.HasPrefix(md.MetricID, "ext:custom.my_app.latency") {
					t.Errorf("expected ext: prefix for custom metric, got %q", md.MetricID)
				}
			},
		},
		{
			name: "no unit and no per_unit",
			input: &datadog.MetricMetadata{
				Metric: "custom.dimensionless",
			},
			check: func(t *testing.T, md *dynatrace.MetricDescriptor) {
				if md.Unit != "Unspecified" {
					t.Errorf("expected Unspecified for no unit, got %q", md.Unit)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, err := ConvertMetricMetadata(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, md)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Downtime conversion tests
// ---------------------------------------------------------------------------

func TestConvertDowntime(t *testing.T) {
	endTime := int64(1700010000)

	tests := []struct {
		name       string
		input      *datadog.Downtime
		wantErr    bool
		errContain string
		check      func(t *testing.T, mw *dynatrace.MaintenanceWindow)
	}{
		{
			name: "one-time downtime",
			input: &datadog.Downtime{
				ID:       1,
				Message:  "Planned maintenance",
				Start:    1700000000,
				End:      &endTime,
				Scope:    []string{"host:web01"},
				Timezone: "America/New_York",
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if !strings.Contains(mw.Name, "Planned maintenance") {
					t.Errorf("expected name to contain message, got %q", mw.Name)
				}
				if mw.Type != "PLANNED" {
					t.Errorf("expected PLANNED type, got %q", mw.Type)
				}
				if mw.Suppression != "DETECT_PROBLEMS_DONT_ALERT" {
					t.Errorf("expected DETECT_PROBLEMS_DONT_ALERT, got %q", mw.Suppression)
				}
				if mw.Schedule.RecurrenceType != "ONCE" {
					t.Errorf("expected ONCE recurrence, got %q", mw.Schedule.RecurrenceType)
				}
				if mw.Schedule.ZoneID != "America/New_York" {
					t.Errorf("expected timezone America/New_York, got %q", mw.Schedule.ZoneID)
				}
				if mw.Schedule.Start == "" {
					t.Error("expected start time to be set")
				}
				if mw.Schedule.End == "" {
					t.Error("expected end time to be set")
				}
				if mw.Schedule.Recurrence != nil {
					t.Error("expected no recurrence for one-time downtime")
				}
			},
		},
		{
			name: "recurring weekly downtime",
			input: &datadog.Downtime{
				ID:       2,
				Message:  "Weekly patching",
				Start:    1700000000,
				End:      &endTime,
				Scope:    []string{"*"},
				Timezone: "UTC",
				Recurrence: &datadog.DowntimeRecurrence{
					Type:     "weeks",
					Period:   1,
					WeekDays: []string{"monday", "wednesday"},
				},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Schedule.RecurrenceType != "WEEKLY" {
					t.Errorf("expected WEEKLY recurrence, got %q", mw.Schedule.RecurrenceType)
				}
				if mw.Schedule.Recurrence == nil {
					t.Fatal("expected recurrence to be set")
				}
				if mw.Schedule.Recurrence.DayOfWeek != "MONDAY" {
					t.Errorf("expected DayOfWeek MONDAY, got %q", mw.Schedule.Recurrence.DayOfWeek)
				}
				if mw.Schedule.Recurrence.DurationMinutes <= 0 {
					t.Errorf("expected positive duration, got %d", mw.Schedule.Recurrence.DurationMinutes)
				}
			},
		},
		{
			name: "daily recurring downtime",
			input: &datadog.Downtime{
				ID:      3,
				Message: "Daily maintenance",
				Start:   1700000000,
				End:     &endTime,
				Scope:   []string{"*"},
				Recurrence: &datadog.DowntimeRecurrence{
					Type:   "days",
					Period: 1,
				},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Schedule.RecurrenceType != "DAILY" {
					t.Errorf("expected DAILY recurrence, got %q", mw.Schedule.RecurrenceType)
				}
			},
		},
		{
			name: "monthly recurring downtime",
			input: &datadog.Downtime{
				ID:      4,
				Message: "Monthly maintenance",
				Start:   1700000000,
				End:     &endTime,
				Scope:   []string{"*"},
				Recurrence: &datadog.DowntimeRecurrence{
					Type:   "months",
					Period: 1,
				},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Schedule.RecurrenceType != "MONTHLY" {
					t.Errorf("expected MONTHLY recurrence, got %q", mw.Schedule.RecurrenceType)
				}
			},
		},
		{
			name: "disabled downtime is skipped",
			input: &datadog.Downtime{
				ID:       5,
				Message:  "Disabled one",
				Disabled: true,
				Start:    1700000000,
				Scope:    []string{"*"},
			},
			wantErr:    true,
			errContain: "disabled",
		},
		{
			name: "timezone defaults to UTC when empty",
			input: &datadog.Downtime{
				ID:      6,
				Message: "No timezone",
				Start:   1700000000,
				End:     &endTime,
				Scope:   []string{"*"},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Schedule.ZoneID != "UTC" {
					t.Errorf("expected UTC default timezone, got %q", mw.Schedule.ZoneID)
				}
			},
		},
		{
			name: "downtime with no end gets 24h default",
			input: &datadog.Downtime{
				ID:      7,
				Message: "No end time",
				Start:   1700000000,
				End:     nil,
				Scope:   []string{"*"},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Schedule.Start == "" || mw.Schedule.End == "" {
					t.Error("expected both start and end to be set")
				}
				// Both should be valid formatted dates
				if !strings.Contains(mw.Schedule.Start, " ") {
					t.Errorf("expected formatted date in start, got %q", mw.Schedule.Start)
				}
			},
		},
		{
			name: "downtime with monitor tags creates scope matches",
			input: &datadog.Downtime{
				ID:          8,
				Message:     "Tagged downtime",
				Start:       1700000000,
				End:         &endTime,
				Scope:       []string{"*"},
				MonitorTags: []string{"env:prod", "team:platform"},
			},
			check: func(t *testing.T, mw *dynatrace.MaintenanceWindow) {
				if mw.Scope == nil {
					t.Fatal("expected scope to be set")
				}
				if len(mw.Scope.Matches) != 2 {
					t.Fatalf("expected 2 scope matches, got %d", len(mw.Scope.Matches))
				}
				if mw.Scope.Matches[0].Type != "TAG" {
					t.Errorf("expected TAG match type, got %q", mw.Scope.Matches[0].Type)
				}
				if mw.Scope.Matches[0].TagCombination != "OR" {
					t.Errorf("expected OR tag combination, got %q", mw.Scope.Matches[0].TagCombination)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := ConvertDowntime(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, mw)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Notification conversion tests
// ---------------------------------------------------------------------------

func TestConvertNotification(t *testing.T) {
	tests := []struct {
		name       string
		input      *datadog.NotificationChannel
		wantErr    bool
		errContain string
		check      func(t *testing.T, ni *dynatrace.NotificationIntegration)
	}{
		{
			name: "slack notification",
			input: &datadog.NotificationChannel{
				Name: "Team Slack",
				Type: "slack",
				Config: map[string]interface{}{
					"url":     "https://hooks.slack.com/services/T00/B00/xxx",
					"channel": "#alerts",
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "SLACK" {
					t.Errorf("expected SLACK, got %q", ni.Type)
				}
				if ni.Name != "Team Slack" {
					t.Errorf("expected name %q, got %q", "Team Slack", ni.Name)
				}
				if !ni.Active {
					t.Error("expected notification to be active")
				}
				if ni.Config["url"] != "https://hooks.slack.com/services/T00/B00/xxx" {
					t.Errorf("expected URL in config, got %v", ni.Config["url"])
				}
				if ni.Config["channel"] != "#alerts" {
					t.Errorf("expected channel in config, got %v", ni.Config["channel"])
				}
			},
		},
		{
			name: "pagerduty notification",
			input: &datadog.NotificationChannel{
				Name: "Oncall PD",
				Type: "pagerduty",
				Config: map[string]interface{}{
					"service_key": "abc123def456",
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "PAGER_DUTY" {
					t.Errorf("expected PAGER_DUTY, got %q", ni.Type)
				}
				if ni.Config["account"] != "abc123def456" {
					t.Errorf("expected service_key mapped to account, got %v", ni.Config["account"])
				}
			},
		},
		{
			name: "email notification",
			input: &datadog.NotificationChannel{
				Name: "Team Email",
				Type: "email",
				Config: map[string]interface{}{
					"emails": "team@example.com,oncall@example.com",
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "EMAIL" {
					t.Errorf("expected EMAIL, got %q", ni.Type)
				}
				if ni.Config["receivers"] != "team@example.com,oncall@example.com" {
					t.Errorf("expected receivers in config, got %v", ni.Config["receivers"])
				}
			},
		},
		{
			name: "webhook notification",
			input: &datadog.NotificationChannel{
				Name: "Custom Webhook",
				Type: "webhook",
				Config: map[string]interface{}{
					"url":     "https://hooks.example.com/alert",
					"payload": `{"text":"alert triggered"}`,
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "WEBHOOK" {
					t.Errorf("expected WEBHOOK, got %q", ni.Type)
				}
				if ni.Config["url"] != "https://hooks.example.com/alert" {
					t.Errorf("expected URL in config, got %v", ni.Config["url"])
				}
				if ni.Config["payload"] != `{"text":"alert triggered"}` {
					t.Errorf("expected payload in config, got %v", ni.Config["payload"])
				}
			},
		},
		{
			name: "opsgenie notification",
			input: &datadog.NotificationChannel{
				Name: "OpsGenie Alert",
				Type: "opsgenie",
				Config: map[string]interface{}{
					"api_key": "og-key-123",
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "OPS_GENIE" {
					t.Errorf("expected OPS_GENIE, got %q", ni.Type)
				}
				if ni.Config["apiKey"] != "og-key-123" {
					t.Errorf("expected apiKey in config, got %v", ni.Config["apiKey"])
				}
			},
		},
		{
			name: "victorops notification",
			input: &datadog.NotificationChannel{
				Name: "VictorOps Alert",
				Type: "victorops",
				Config: map[string]interface{}{
					"api_key": "vo-key-456",
				},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "VICTOR_OPS" {
					t.Errorf("expected VICTOR_OPS, got %q", ni.Type)
				}
				if ni.Config["apiKey"] != "vo-key-456" {
					t.Errorf("expected apiKey in config, got %v", ni.Config["apiKey"])
				}
			},
		},
		{
			name: "unsupported notification type returns error",
			input: &datadog.NotificationChannel{
				Name:   "Teams Channel",
				Type:   "microsoft-teams",
				Config: map[string]interface{}{},
			},
			wantErr:    true,
			errContain: "unsupported notification type",
		},
		{
			name: "slack with missing optional fields still works",
			input: &datadog.NotificationChannel{
				Name:   "Minimal Slack",
				Type:   "slack",
				Config: map[string]interface{}{},
			},
			check: func(t *testing.T, ni *dynatrace.NotificationIntegration) {
				if ni.Type != "SLACK" {
					t.Errorf("expected SLACK, got %q", ni.Type)
				}
				if _, exists := ni.Config["url"]; exists {
					t.Error("expected no url key when not provided")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ni, err := ConvertNotification(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, ni)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Notebook conversion tests
// ---------------------------------------------------------------------------

func TestConvertNotebook(t *testing.T) {
	tests := []struct {
		name  string
		input *datadog.Notebook
		check func(t *testing.T, nb *dynatrace.DynatraceNotebook)
	}{
		{
			name: "markdown cell",
			input: &datadog.Notebook{
				Name: "Analysis Notebook",
				Cells: []datadog.NotebookCell{
					{
						ID:   "cell-1",
						Type: "markdown",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{
								"text": "# Analysis\nSome markdown content here.",
							},
						},
					},
				},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if nb.Name != "Analysis Notebook" {
					t.Errorf("expected name %q, got %q", "Analysis Notebook", nb.Name)
				}
				if len(nb.Sections) != 1 {
					t.Fatalf("expected 1 section, got %d", len(nb.Sections))
				}
				s := nb.Sections[0]
				if s.Type != "markdown" {
					t.Errorf("expected markdown type, got %q", s.Type)
				}
				if s.Content != "# Analysis\nSome markdown content here." {
					t.Errorf("expected markdown content, got %q", s.Content)
				}
				if s.ID != "cell-1" {
					t.Errorf("expected cell ID %q, got %q", "cell-1", s.ID)
				}
			},
		},
		{
			name: "timeseries cell becomes code section with chart visualization",
			input: &datadog.Notebook{
				Name: "TS Notebook",
				Cells: []datadog.NotebookCell{
					{
						ID:   "cell-ts",
						Type: "timeseries",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{
								"requests": []interface{}{
									map[string]interface{}{
										"q": "avg:system.cpu.user{*}",
									},
								},
							},
						},
					},
				},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if len(nb.Sections) != 1 {
					t.Fatalf("expected 1 section, got %d", len(nb.Sections))
				}
				s := nb.Sections[0]
				if s.Type != "code" {
					t.Errorf("expected code type for timeseries, got %q", s.Type)
				}
				if s.Visualization != "chart" {
					t.Errorf("expected chart visualization, got %q", s.Visualization)
				}
				if s.Query == "" {
					t.Error("expected non-empty query")
				}
			},
		},
		{
			name: "unsupported cell type gets markdown fallback",
			input: &datadog.Notebook{
				Name: "Unsupported Cells",
				Cells: []datadog.NotebookCell{
					{
						ID:   "cell-unk",
						Type: "check_status",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{},
						},
					},
				},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if len(nb.Sections) != 1 {
					t.Fatalf("expected 1 section, got %d", len(nb.Sections))
				}
				s := nb.Sections[0]
				if s.Type != "markdown" {
					t.Errorf("expected markdown fallback, got %q", s.Type)
				}
				if !strings.Contains(s.Content, "check_status") {
					t.Errorf("expected original type in content, got %q", s.Content)
				}
				if !strings.Contains(s.Content, "manual configuration") {
					t.Errorf("expected migration note in content, got %q", s.Content)
				}
			},
		},
		{
			name: "empty notebook (no cells)",
			input: &datadog.Notebook{
				Name:  "Empty Notebook",
				Cells: []datadog.NotebookCell{},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if nb.Name != "Empty Notebook" {
					t.Errorf("expected name %q, got %q", "Empty Notebook", nb.Name)
				}
				if len(nb.Sections) != 0 {
					t.Errorf("expected 0 sections, got %d", len(nb.Sections))
				}
			},
		},
		{
			name: "multiple cells of mixed types",
			input: &datadog.Notebook{
				Name: "Mixed Notebook",
				Cells: []datadog.NotebookCell{
					{
						ID:   "c1",
						Type: "markdown",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{
								"text": "Introduction",
							},
						},
					},
					{
						ID:   "c2",
						Type: "timeseries",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{},
						},
					},
					{
						ID:   "c3",
						Type: "unknown_widget",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{},
						},
					},
				},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if len(nb.Sections) != 3 {
					t.Fatalf("expected 3 sections, got %d", len(nb.Sections))
				}
				if nb.Sections[0].Type != "markdown" {
					t.Errorf("section 0: expected markdown, got %q", nb.Sections[0].Type)
				}
				if nb.Sections[1].Type != "code" {
					t.Errorf("section 1: expected code, got %q", nb.Sections[1].Type)
				}
				if nb.Sections[2].Type != "markdown" {
					t.Errorf("section 2: expected markdown fallback, got %q", nb.Sections[2].Type)
				}
			},
		},
		{
			name: "timeseries cell without parseable requests gets placeholder query",
			input: &datadog.Notebook{
				Name: "No Query TS",
				Cells: []datadog.NotebookCell{
					{
						ID:   "c-nq",
						Type: "timeseries",
						Attributes: datadog.NotebookCellAttributes{
							Definition: map[string]interface{}{},
						},
					},
				},
			},
			check: func(t *testing.T, nb *dynatrace.DynatraceNotebook) {
				if len(nb.Sections) != 1 {
					t.Fatalf("expected 1 section, got %d", len(nb.Sections))
				}
				if !strings.Contains(nb.Sections[0].Query, "Migrated from DataDog") {
					t.Errorf("expected migration placeholder query, got %q", nb.Sections[0].Query)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb, err := ConvertNotebook(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, nb)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ConvertAll integration tests
// ---------------------------------------------------------------------------

func TestConvertAll(t *testing.T) {
	critical := 90.0
	endTime := int64(1700010000)

	tests := []struct {
		name         string
		input        *datadog.ExtractionResult
		wantErrCount int
		check        func(t *testing.T, result *dynatrace.ConversionResult, errs []error)
	}{
		{
			name:  "empty extraction produces empty result with no errors",
			input: &datadog.ExtractionResult{},
			check: func(t *testing.T, result *dynatrace.ConversionResult, errs []error) {
				if len(errs) != 0 {
					t.Errorf("expected no errors, got %v", errs)
				}
				if len(result.Dashboards) != 0 {
					t.Errorf("expected 0 dashboards, got %d", len(result.Dashboards))
				}
				if len(result.MetricEvents) != 0 {
					t.Errorf("expected 0 metric events, got %d", len(result.MetricEvents))
				}
				if len(result.SLOs) != 0 {
					t.Errorf("expected 0 SLOs, got %d", len(result.SLOs))
				}
				if len(result.Synthetics) != 0 {
					t.Errorf("expected 0 synthetics, got %d", len(result.Synthetics))
				}
				if len(result.LogRules) != 0 {
					t.Errorf("expected 0 log rules, got %d", len(result.LogRules))
				}
				if len(result.Metrics) != 0 {
					t.Errorf("expected 0 metrics, got %d", len(result.Metrics))
				}
				if len(result.Maintenance) != 0 {
					t.Errorf("expected 0 maintenance windows, got %d", len(result.Maintenance))
				}
				if len(result.Notifications) != 0 {
					t.Errorf("expected 0 notifications, got %d", len(result.Notifications))
				}
				if len(result.Notebooks) != 0 {
					t.Errorf("expected 0 notebooks, got %d", len(result.Notebooks))
				}
			},
		},
		{
			name: "mixed resources all convert successfully",
			input: &datadog.ExtractionResult{
				Dashboards: []datadog.Dashboard{
					{
						Title: "Main Dashboard",
						Widgets: []datadog.Widget{
							{Definition: datadog.WidgetDefinition{Type: "note", Title: "Header", Content: "Welcome"}},
						},
					},
				},
				Monitors: []datadog.Monitor{
					{
						Name:  "CPU Alert",
						Type:  "metric alert",
						Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
						Options: datadog.MonitorOptions{
							Thresholds: &datadog.Thresholds{Critical: &critical},
						},
					},
				},
				SLOs: []datadog.SLO{
					{
						Name: "Uptime SLO",
						Type: "metric",
						Query: &datadog.SLOQuery{
							Numerator:   "sum:requests.ok{*}",
							Denominator: "sum:requests.total{*}",
						},
						Thresholds: []datadog.SLOThreshold{
							{Timeframe: "30d", Target: 99.9},
						},
					},
				},
				Synthetics: []datadog.SyntheticTest{
					{
						Name: "Health Check", Type: "api", Status: "live",
						Config:    datadog.SyntheticConfig{Request: &datadog.SyntheticRequest{Method: "GET", URL: "https://health.test"}},
						Options:   datadog.SyntheticOptions{TickEvery: 300},
						Locations: []string{"aws:us-east-1"},
					},
				},
				LogPipelines: []datadog.LogPipeline{
					{
						Name:      "App Logs",
						IsEnabled: true,
						Filter:    &datadog.LogFilter{Query: "source:app"},
					},
				},
				Metrics: []datadog.MetricMetadata{
					{Metric: "system.cpu.user", Unit: "percent"},
				},
				Downtimes: []datadog.Downtime{
					{
						ID: 1, Message: "Maint", Start: 1700000000, End: &endTime,
						Scope: []string{"*"},
					},
				},
				Notifications: []datadog.NotificationChannel{
					{
						Name: "Slack Alerts", Type: "slack",
						Config: map[string]interface{}{"url": "https://hooks.slack.com/xxx", "channel": "#alerts"},
					},
				},
				Notebooks: []datadog.Notebook{
					{
						Name: "Analysis",
						Cells: []datadog.NotebookCell{
							{ID: "c1", Type: "markdown", Attributes: datadog.NotebookCellAttributes{Definition: map[string]interface{}{"text": "hi"}}},
						},
					},
				},
			},
			check: func(t *testing.T, result *dynatrace.ConversionResult, errs []error) {
				if len(errs) != 0 {
					t.Errorf("expected no errors, got %v", errs)
				}
				if len(result.Dashboards) != 1 {
					t.Errorf("expected 1 dashboard, got %d", len(result.Dashboards))
				}
				if len(result.MetricEvents) != 1 {
					t.Errorf("expected 1 metric event, got %d", len(result.MetricEvents))
				}
				if len(result.SLOs) != 1 {
					t.Errorf("expected 1 SLO, got %d", len(result.SLOs))
				}
				if len(result.Synthetics) != 1 {
					t.Errorf("expected 1 synthetic, got %d", len(result.Synthetics))
				}
				if len(result.LogRules) != 1 {
					t.Errorf("expected 1 log rule, got %d", len(result.LogRules))
				}
				if len(result.Metrics) != 1 {
					t.Errorf("expected 1 metric, got %d", len(result.Metrics))
				}
				if len(result.Maintenance) != 1 {
					t.Errorf("expected 1 maintenance, got %d", len(result.Maintenance))
				}
				if len(result.Notifications) != 1 {
					t.Errorf("expected 1 notification, got %d", len(result.Notifications))
				}
				if len(result.Notebooks) != 1 {
					t.Errorf("expected 1 notebook, got %d", len(result.Notebooks))
				}
			},
		},
		{
			name: "error collection: non-fatal errors are collected, other conversions continue",
			input: &datadog.ExtractionResult{
				Dashboards: []datadog.Dashboard{
					// This will error: empty widgets
					{Title: "Bad Dashboard", Widgets: []datadog.Widget{}},
					// This will succeed
					{
						Title: "Good Dashboard",
						Widgets: []datadog.Widget{
							{Definition: datadog.WidgetDefinition{Type: "note", Title: "ok", Content: "ok"}},
						},
					},
				},
				Notifications: []datadog.NotificationChannel{
					// This will error: unsupported type
					{Name: "Bad Channel", Type: "unsupported_type", Config: map[string]interface{}{}},
					// This will succeed
					{Name: "Good Slack", Type: "slack", Config: map[string]interface{}{"url": "https://hooks.slack.com/xxx"}},
				},
				SLOs: []datadog.SLO{
					// This will error: unsupported type
					{Name: "Bad SLO", Type: "unknown_slo_type", Thresholds: []datadog.SLOThreshold{{Timeframe: "30d", Target: 99.0}}},
				},
				Downtimes: []datadog.Downtime{
					// This will error: disabled
					{ID: 99, Message: "disabled", Disabled: true, Start: 1700000000, Scope: []string{"*"}},
				},
				Synthetics: []datadog.SyntheticTest{
					// This will error: unsupported type
					{Name: "Bad Synth", Type: "unsupported_synth", Status: "live", Locations: []string{"aws:us-east-1"}},
				},
			},
			check: func(t *testing.T, result *dynatrace.ConversionResult, errs []error) {
				// We expect errors from: bad dashboard, bad notification, bad SLO, disabled downtime, bad synthetic
				if len(errs) < 5 {
					t.Errorf("expected at least 5 errors, got %d: %v", len(errs), errs)
				}
				// But the good resources should still succeed
				if len(result.Dashboards) != 1 {
					t.Errorf("expected 1 good dashboard, got %d", len(result.Dashboards))
				}
				if len(result.Notifications) != 1 {
					t.Errorf("expected 1 good notification, got %d", len(result.Notifications))
				}
			},
		},
		{
			name: "error messages contain resource identifiers",
			input: &datadog.ExtractionResult{
				Dashboards: []datadog.Dashboard{
					{Title: "Failing Dashboard", Widgets: []datadog.Widget{}},
				},
				Notifications: []datadog.NotificationChannel{
					{Name: "Bad Notif", Type: "nonexistent", Config: map[string]interface{}{}},
				},
			},
			check: func(t *testing.T, result *dynatrace.ConversionResult, errs []error) {
				if len(errs) < 2 {
					t.Fatalf("expected at least 2 errors, got %d", len(errs))
				}
				foundDash := false
				foundNotif := false
				for _, err := range errs {
					msg := err.Error()
					if strings.Contains(msg, "Failing Dashboard") {
						foundDash = true
					}
					if strings.Contains(msg, "Bad Notif") {
						foundNotif = true
					}
				}
				if !foundDash {
					t.Error("expected error message to contain dashboard title")
				}
				if !foundNotif {
					t.Error("expected error message to contain notification name")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			result, errs := c.ConvertAll(tt.input)
			if tt.check != nil {
				tt.check(t, result, errs)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "strips @mentions on their own lines",
			input:    "Alert!\n@slack-channel\n@pagerduty-service\nDetails here",
			contains: []string{"Alert!", "Details here"},
			excludes: []string{"@slack", "@pagerduty"},
		},
		{
			name:     "preserves empty message",
			input:    "",
			contains: []string{},
			excludes: []string{},
		},
		{
			name:     "preserves message with no mentions",
			input:    "This is a clean message",
			contains: []string{"This is a clean message"},
			excludes: []string{},
		},
		{
			name:     "strips multiple @mentions",
			input:    "@admin\n@ops\nCheck the logs",
			contains: []string{"Check the logs"},
			excludes: []string{"@admin", "@ops"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeDescription(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("expected result to contain %q, got %q", want, result)
				}
			}
			for _, exclude := range tt.excludes {
				if strings.Contains(result, exclude) {
					t.Errorf("expected result NOT to contain %q, got %q", exclude, result)
				}
			}
		})
	}
}

func TestMapFrequency(t *testing.T) {
	tests := []struct {
		tickEvery int
		want      int
	}{
		{0, 15},
		{-1, 15},
		{30, 1},
		{60, 1},
		{90, 1},
		{120, 2},
		{180, 5},
		{300, 5},
		{600, 10},
		{900, 15},
		{1800, 30},
		{3600, 60},
		{7200, 60},
	}

	for _, tt := range tests {
		got := mapFrequency(tt.tickEvery)
		if got != tt.want {
			t.Errorf("mapFrequency(%d) = %d, want %d", tt.tickEvery, got, tt.want)
		}
	}
}

func TestMapSLOTimeframe(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"7d", "-1w"},
		{"30d", "-1M"},
		{"90d", "-3M"},
		{"custom", "-1M"},
		{"unknown", "-1M"},
	}

	for _, tt := range tests {
		got := mapSLOTimeframe(tt.input)
		if got != tt.want {
			t.Errorf("mapSLOTimeframe(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapRecurrenceType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"days", "DAILY"},
		{"weeks", "WEEKLY"},
		{"months", "MONTHLY"},
		{"years", "ONCE"},
		{"unknown", "ONCE"},
	}

	for _, tt := range tests {
		got := mapRecurrenceType(tt.input)
		if got != tt.want {
			t.Errorf("mapRecurrenceType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapMonitorSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"metric alert", "CUSTOM_ALERT"},
		{"query alert", "CUSTOM_ALERT"},
		{"service check", "ERROR"},
		{"event alert", "INFO"},
		{"log alert", "CUSTOM_ALERT"},
		{"composite", "CUSTOM_ALERT"},
		{"unknown_type", "CUSTOM_ALERT"},
	}

	for _, tt := range tests {
		got := mapMonitorSeverity(tt.input)
		if got != tt.want {
			t.Errorf("mapMonitorSeverity(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapMetricUnit(t *testing.T) {
	tests := []struct {
		unit    string
		perUnit string
		want    string
	}{
		{"byte", "", "Byte"},
		{"kilobyte", "", "KiloByte"},
		{"megabyte", "", "MegaByte"},
		{"gigabyte", "", "GigaByte"},
		{"terabyte", "", "TeraByte"},
		{"percent", "", "Percent"},
		{"nanosecond", "", "NanoSecond"},
		{"microsecond", "", "MicroSecond"},
		{"millisecond", "", "MilliSecond"},
		{"second", "", "Second"},
		{"minute", "", "Minute"},
		{"hour", "", "Hour"},
		{"day", "", "Day"},
		{"bit", "", "Bit"},
		{"kilobit", "", "KiloBit"},
		{"megabit", "", "MegaBit"},
		{"gigabit", "", "GigaBit"},
		{"count", "", "Count"},
		{"operation", "", "Count"},
		{"request", "", "Count"},
		{"error", "", "Count"},
		{"connection", "", "Count"},
		{"unknown_unit", "", "Unspecified"},
		{"", "", "Unspecified"},
		{"byte", "second", "BytePerSecond"},
		{"megabit", "second", "MegaBitPerSecond"},
		{"count", "second", "CountPerSecond"},
		{"operation", "minute", "CountPerMinute"},
		{"byte", "unknown_per", "Byte"},
	}

	for _, tt := range tests {
		name := tt.unit
		if tt.perUnit != "" {
			name += "_per_" + tt.perUnit
		}
		t.Run(name, func(t *testing.T) {
			got := mapMetricUnit(tt.unit, tt.perUnit)
			if got != tt.want {
				t.Errorf("mapMetricUnit(%q, %q) = %q, want %q", tt.unit, tt.perUnit, got, tt.want)
			}
		})
	}
}
