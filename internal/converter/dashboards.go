package converter

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// ConvertDashboard converts a DataDog dashboard to a Dynatrace dashboard.
func ConvertDashboard(dd *datadog.Dashboard) (*dynatrace.Dashboard, error) {
	dt := &dynatrace.Dashboard{
		DashboardMetadata: dynatrace.DashboardMetadata{
			Name:  dd.Title,
			Owner: "datadog2dynatrace",
		},
	}

	// Convert template variables to a guidance tile
	if len(dd.TemplateVars) > 0 {
		var md strings.Builder
		md.WriteString("**Template Variables (from DataDog)**\n\n")
		md.WriteString("This dashboard used the following template variables in DataDog:\n\n")
		md.WriteString("| Name | Prefix | Default |\n")
		md.WriteString("|---|---|---|\n")
		for _, tv := range dd.TemplateVars {
			def := tv.Default
			if def == "" {
				def = "*"
			}
			md.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", tv.Name, tv.Prefix, def))
		}
		md.WriteString("\nConfigure **Dynatrace management zone filters** or **dashboard variables** to replicate this behavior.")
		dt.Tiles = append(dt.Tiles, dynatrace.Tile{
			Configured: true,
			TileType:   "MARKDOWN",
			Name:       "Template Variables",
			Markdown:   md.String(),
			Bounds:     dynatrace.TileBounds{Top: 0, Left: 0, Width: 912, Height: 152},
		})
	}

	for i, w := range dd.Widgets {
		if w.Definition.Type == "group" {
			// Emit header tile
			header := &dynatrace.Tile{
				Configured: true,
				TileType:   "HEADER",
				Name:       w.Definition.Title,
				Bounds:     calculateBounds(w.Layout, i),
			}
			dt.Tiles = append(dt.Tiles, *header)
			// Recurse into nested widgets
			for j, child := range w.Definition.Widgets {
				tile, err := convertWidget(&child, len(dd.Widgets)+j)
				if err != nil {
					continue
				}
				dt.Tiles = append(dt.Tiles, *tile)
			}
			continue
		}
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

	logging.Debug("converting widget type %q (%q)", w.Definition.Type, w.Definition.Title)

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
	case "heatmap":
		return convertApproxWidget(w, tile, "heatmap", convertTimeseriesWidget)
	case "distribution":
		return convertApproxWidget(w, tile, "distribution", convertTimeseriesWidget)
	case "change":
		return convertApproxWidget(w, tile, "change", convertQueryValueWidget)
	case "hostmap":
		return convertHostmapWidget(w, tile)
	case "table":
		return convertTableWidget(w, tile)
	case "slo":
		return convertSLOWidget(w, tile)
	default:
		logging.Debug("unsupported widget type %q, falling back to MARKDOWN", w.Definition.Type)
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

	qIdx := 0
	var lastDQL queryResult
	for _, req := range w.Definition.Requests {
		for _, qr := range extractQueries(&req) {
			if qr.MetricSelector != "" {
				qIdx++
				tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
					ID:             fmt.Sprintf("Q%d", qIdx),
					MetricSelector: qr.MetricSelector,
				})
			} else if qr.DQL != "" {
				lastDQL = qr
			}
		}
	}

	if len(tile.Queries) > 0 {
		return tile, nil
	}
	if lastDQL.DQL != "" {
		return buildDQLMarkdownTile(w.Definition.Title, lastDQL.DQL, lastDQL.SourceType, tile.Bounds), nil
	}
	return nil, fmt.Errorf("no queries could be converted")
}

func convertQueryValueWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	qIdx := 0
	var lastDQL queryResult
	for _, req := range w.Definition.Requests {
		for _, qr := range extractQueries(&req) {
			if qr.MetricSelector != "" {
				qIdx++
				tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
					ID:             fmt.Sprintf("Q%d", qIdx),
					MetricSelector: qr.MetricSelector,
				})
			} else if qr.DQL != "" {
				lastDQL = qr
			}
		}
	}

	if len(tile.Queries) > 0 {
		return tile, nil
	}
	if lastDQL.DQL != "" {
		return buildDQLMarkdownTile(w.Definition.Title, lastDQL.DQL, lastDQL.SourceType, tile.Bounds), nil
	}
	return nil, fmt.Errorf("no queries could be converted")
}

func convertToplistWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	qIdx := 0
	var lastDQL queryResult
	for _, req := range w.Definition.Requests {
		for _, qr := range extractQueries(&req) {
			if qr.MetricSelector != "" {
				qIdx++
				tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
					ID:             fmt.Sprintf("Q%d", qIdx),
					MetricSelector: qr.MetricSelector + ":sort(value(avg,descending)):limit(10)",
				})
			} else if qr.DQL != "" {
				lastDQL = qr
			}
		}
	}

	if len(tile.Queries) > 0 {
		return tile, nil
	}
	if lastDQL.DQL != "" {
		return buildDQLMarkdownTile(w.Definition.Title, lastDQL.DQL, lastDQL.SourceType, tile.Bounds), nil
	}
	return nil, fmt.Errorf("no queries could be converted")
}

func convertNoteWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "MARKDOWN"
	tile.Name = w.Definition.Title
	tile.Markdown = w.Definition.Content
	return tile, nil
}

func convertHostmapWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	// Try to extract metric queries from the hostmap widget
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	qIdx := 0
	for _, req := range w.Definition.Requests {
		for _, qr := range extractQueries(&req) {
			if qr.MetricSelector != "" {
				qIdx++
				tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
					ID:             fmt.Sprintf("Q%d", qIdx),
					MetricSelector: qr.MetricSelector,
				})
			}
		}
	}

	if len(tile.Queries) > 0 {
		return tile, nil
	}

	// Fallback to HOSTS tile
	tile.TileType = "HOSTS"
	tile.Queries = nil
	return tile, nil
}

func convertTableWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "DATA_EXPLORER"
	tile.Name = w.Definition.Title

	qIdx := 0
	var lastDQL queryResult
	for _, req := range w.Definition.Requests {
		for _, qr := range extractQueries(&req) {
			if qr.MetricSelector != "" {
				qIdx++
				tile.Queries = append(tile.Queries, dynatrace.DashboardQuery{
					ID:             fmt.Sprintf("Q%d", qIdx),
					MetricSelector: qr.MetricSelector,
				})
			} else if qr.DQL != "" {
				lastDQL = qr
			}
		}
	}

	if len(tile.Queries) > 0 {
		return tile, nil
	}
	if lastDQL.DQL != "" {
		return buildDQLMarkdownTile(w.Definition.Title, lastDQL.DQL, lastDQL.SourceType, tile.Bounds), nil
	}
	return nil, fmt.Errorf("no queries could be converted")
}

func convertSLOWidget(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error) {
	tile.TileType = "MARKDOWN"
	tile.Name = w.Definition.Title
	tile.Markdown = fmt.Sprintf("**%s** (SLO Widget)\n\nThis DataDog SLO widget has been migrated.\nLink this tile to the corresponding **Dynatrace SLO** that was created during migration.\n\nTo add an SLO tile, edit this dashboard in Dynatrace and replace this markdown tile with an SLO tile.", w.Definition.Title)
	return tile, nil
}

// convertApproxWidget wraps a converter and appends an approximation note to the tile name.
type widgetConverter func(w *datadog.Widget, tile *dynatrace.Tile) (*dynatrace.Tile, error)

func convertApproxWidget(w *datadog.Widget, tile *dynatrace.Tile, sourceType string, fn widgetConverter) (*dynatrace.Tile, error) {
	result, err := fn(w, tile)
	if err != nil {
		return nil, err
	}
	result.Name = result.Name + fmt.Sprintf(" (approx. from %s)", sourceType)
	return result, nil
}

// queryResult discriminates between MetricSelector (classic dashboard tiles)
// and DQL (which requires a Notebook or Grail-powered dashboard).
type queryResult struct {
	MetricSelector string
	DQL            string
	SourceType     string // "metric", "log", or "apm"
}

// extractQueries extracts all query results from a widget request,
// processing every entry in req.Queries (not just the first one).
func extractQueries(req *datadog.WidgetRequest) []queryResult {
	var results []queryResult

	// Try the simple query string first
	if req.Query != "" {
		parsed, err := query.Parse(req.Query)
		if err == nil {
			results = append(results, queryResult{MetricSelector: query.ToMetricSelector(parsed), SourceType: "metric"})
		}
	}

	// Process ALL entries in queries/formulas format
	for _, qd := range req.Queries {
		parsed, err := query.Parse(qd.Query)
		if err == nil {
			results = append(results, queryResult{MetricSelector: query.ToMetricSelector(parsed), SourceType: "metric"})
		}
	}

	// Log/APM queries produce DQL, not MetricSelector
	if req.LogQuery != nil && req.LogQuery.Search != nil {
		results = append(results, queryResult{DQL: query.ToDQL(req.LogQuery.Search.Query), SourceType: "log"})
	}
	if req.ApmQuery != nil && req.ApmQuery.Search != nil {
		results = append(results, queryResult{DQL: query.ToDQL(req.ApmQuery.Search.Query), SourceType: "apm"})
	}

	return results
}

// extractQuery returns the first query result from a widget request (convenience wrapper).
func extractQuery(req *datadog.WidgetRequest) queryResult {
	results := extractQueries(req)
	if len(results) > 0 {
		return results[0]
	}
	return queryResult{}
}

// buildDQLMarkdownTile creates a MARKDOWN tile with embedded DQL and guidance
// to use a Dynatrace Notebook, since classic dashboards don't support DQL tiles.
func buildDQLMarkdownTile(title, dqlQuery, sourceType string, bounds dynatrace.TileBounds) *dynatrace.Tile {
	label := "Log"
	if sourceType == "apm" {
		label = "APM"
	}
	markdown := fmt.Sprintf("**%s** (%s Query)\n\nThis widget used a DataDog %s query that was translated to DQL.\nClassic Dynatrace dashboards do not support DQL tiles.\nUse a **Dynatrace Notebook** to run this query:\n\n```\n%s\n```", title, label, label, dqlQuery)

	return &dynatrace.Tile{
		Configured: true,
		TileType:   "MARKDOWN",
		Name:       title,
		Markdown:   markdown,
		Bounds:     bounds,
	}
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
