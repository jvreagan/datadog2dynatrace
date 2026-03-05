package converter

import (
	"encoding/json"
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertNotebook converts a DataDog notebook to a Dynatrace notebook.
func ConvertNotebook(dd *datadog.Notebook) (*dynatrace.DynatraceNotebook, error) {
	nb := &dynatrace.DynatraceNotebook{
		Name: dd.Name,
	}

	for _, cell := range dd.Cells {
		section, err := convertNotebookCell(&cell)
		if err != nil {
			// Non-fatal: skip cells that can't be converted
			continue
		}
		nb.Sections = append(nb.Sections, *section)
	}

	return nb, nil
}

func convertNotebookCell(cell *datadog.NotebookCell) (*dynatrace.NotebookSection, error) {
	section := &dynatrace.NotebookSection{
		ID: cell.ID,
	}

	switch cell.Type {
	case "markdown":
		section.Type = "markdown"
		if content, ok := cell.Attributes.Definition["text"]; ok {
			section.Content = fmt.Sprintf("%v", content)
		}

	case "timeseries", "toplist", "heatmap", "distribution":
		section.Type = "code"
		section.Visualization = "chart"
		// Try to extract the query
		if def, ok := cell.Attributes.Definition["requests"]; ok {
			if reqData, err := json.Marshal(def); err == nil {
				var requests []datadog.WidgetRequest
				if json.Unmarshal(reqData, &requests) == nil && len(requests) > 0 {
					if requests[0].Query != "" {
						parsed, err := query.Parse(requests[0].Query)
						if err == nil {
							section.Query = fmt.Sprintf("timeseries %s", query.ToMetricSelector(parsed))
						}
					}
				}
			}
		}
		if section.Query == "" {
			section.Query = "// Migrated from DataDog - requires manual query configuration"
		}

	case "query_value":
		section.Type = "code"
		if def, ok := cell.Attributes.Definition["requests"]; ok {
			if reqData, err := json.Marshal(def); err == nil {
				var requests []datadog.WidgetRequest
				if json.Unmarshal(reqData, &requests) == nil && len(requests) > 0 {
					if requests[0].Query != "" {
						parsed, err := query.Parse(requests[0].Query)
						if err == nil {
							section.Query = query.ToMetricSelector(parsed)
						}
					}
				}
			}
		}

	default:
		section.Type = "markdown"
		section.Content = fmt.Sprintf("**Migrated cell** (original type: `%s`)\n\nThis cell requires manual configuration in Dynatrace.", cell.Type)
	}

	return section, nil
}
