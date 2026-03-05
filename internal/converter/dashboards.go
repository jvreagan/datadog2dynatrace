package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertDashboard converts a DataDog dashboard to a Dynatrace dashboard.
func ConvertDashboard(dd *datadog.Dashboard) (*dynatrace.Dashboard, error) {
	dt := &dynatrace.Dashboard{
		DashboardMetadata: dynatrace.DashboardMetadata{
			Name:  dd.Title,
			Owner: "datadog2dynatrace",
		},
	}

	for i, w := range dd.Widgets {
		tile, err := convertWidget(&w, i)
		if err != nil {
			// Non-fatal: skip unsupported widgets
			continue
		}
		dt.Tiles = append(dt.Tiles, *tile)
	}

	if len(dt.Tiles) == 0 {
		return nil, fmt.Errorf("no widgets could be converted")
	}

	return dt, nil
}

func convertWidget(w *datadog.Widget, index int) (*dynatrace.Tile, error) {
	tile := &dynatrace.Tile{
		Configured: true,
		Bounds:     calculateBounds(w.Layout, index),
	}

	switch w.Definition.Type {
	case "timeseries":
		return convertTimeseriesWidget(w, tile)
	case "query_value":
		return convertQueryValueWidget(w, tile)
	case "toplist":
		return convertToplistWidget(w, tile)
	case "note":
		return convertNoteWidget(w, tile)
	case "free_text":
		return convertNoteWidget(w, tile)
	case "group":
		return convertGroupWidget(w, tile)
	case "heatmap":
		return convertTimeseriesWidget(w, tile) // Approximate with timeseries
	case "distribution":
		return convertTimeseriesWidget(w, tile)
	case "change":
		return convertQueryValueWidget(w, tile) // Approximate with query value
	case "hostmap":
		return convertHostmapWidget(w, tile)
	case "table":
		return convertTableWidget(w, tile)
	case "slo":
		return convertSLOWidget(w, tile)
	default:
		// For unsupported widgets, create a markdown tile with info
		tile.TileType = "MARKDOWN"
		tile.Name = w.Definition.Title
		tile.Markdown = fmt.Sprintf("**%s**\n\nMigrated from DataDog widget type: `%s`\n\nManual configuration required.", w.Definition.Title, w.Definition.Type)
		return tile, nil
	}
}

func convertTimeseriesWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	for i, req := range w.Definition.Requests {
		metricSelector := extractMetricSelector(&req)
		if metricSelector != "" {
			tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
				ID:             fmt.Sprintf("Q%d", i+1),
				MetricSelector: metricSelector,
			})
		}
	}

	if len(tile.Queries) == 0 {
		return nil, fmt.Errorf("no queries could be converted")
	}

	return tile, nil
}

func convertQueryValueWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	for i, req := range w.Definition.Requests {
		metricSelector := extractMetricSelector(&req)
		if metricSelector != "" {
			tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
				ID:             fmt.Sprintf("Q%d", i+1),
				MetricSelector: metricSelector,
			})
		}
	}

	if len(tile.Queries) == 0 {
		return nil, fmt.Errorf("no queries could be converted")
	}

	return tile, nil
}

func convertToplistWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	for i, req := range w.Definition.Requests {
		metricSelector := extractMetricSelector(&req)
		if metricSelector != "" {
			tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
				ID:             fmt.Sprintf("Q%d", i+1),
				MetricSelector: metricSelector + ":sort(value(avg,descending)):limit(10)",
			})
		}
	}

	return tile, nil
}

func convertNoteWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "MARKDOWN"
	tile.Name = w.Definition.Title
	tile.Markdown = w.Definition.Content
	return tile, nil
}

func convertGroupWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	// Group widgets contain nested widgets - flatten them
	tile.TileType = "HEADER"
	tile.Name = w.Definition.Title
	return tile, nil
}

func convertHostmapWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "HOSTS"
	tile.Name = w.Definition.Title
	return tile, nil
}

func convertTableWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	for i, req := range w.Definition.Requests {
		metricSelector := extractMetricSelector(&req)
		if metricSelector != "" {
			tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
				ID:             fmt.Sprintf("Q%d", i+1),
				MetricSelector: metricSelector,
			})
		}
	}

	return tile, nil
}

func convertSLOWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "SLO"
	tile.Name = w.Definition.Title
	return tile, nil
}

func extractMetricSelector(req *datadog.WidgetRequest) string {
	// Try the simple query string first
	if req.Query != "" {
		parsed, err := query.Parse(req.Query)
		if err == nil {
			return query.ToMetricSelector(parsed)
		}
	}

	// Try the newer queries/formulas format
	if len(req.Queries) > 0 {
		parsed, err := query.Parse(req.Queries[0].Query)
		if err == nil {
			return query.ToMetricSelector(parsed)
		}
	}

	return ""
}

func calculateBounds(layout *datadog.WidgetLayout, index int) dynatrace.TileBounds {
	if layout != nil {
		// DD uses a 12-column grid, DT uses pixel-based layout
		// Scale DD coordinates to DT tile bounds
		return dynatrace.TileBounds{
			Top:    layout.Y * 38,
			Left:   layout.X * 76,
			Width:  layout.Width * 76,
			Height: layout.Height * 38,
		}
	}

	// Default grid layout for ordered dashboards
	col := index % 2
	row := index / 2
	return dynatrace.TileBounds{
		Top:    row * 304,
		Left:   col * 456,
		Width:  456,
		Height: 304,
	}
}
