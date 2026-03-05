package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertSLO converts a DataDog SLO to a Dynatrace SLO.
func ConvertSLO(dd *datadog.SLO) (*dynatrace.SLO, error) {
	dt := &dynatrace.SLO{
		Name:           dd.Name,
		Description:    dd.Description,
		EvaluationType: "AGGREGATE",
		Enabled:        true,
	}

	// Map timeframe from first threshold
	if len(dd.Thresholds) > 0 {
		dt.Target = dd.Thresholds[0].Target
		dt.Warning = dd.Thresholds[0].Warning
		dt.Timeframe = mapSLOTimeframe(dd.Thresholds[0].Timeframe)
	} else {
		dt.Target = 99.0
		dt.Warning = 99.5
		dt.Timeframe = "-1M"
	}

	// Convert based on SLO type
	switch dd.Type {
	case "metric":
		if dd.Query != nil {
			numerator := translateSLOQuery(dd.Query.Numerator)
			denominator := translateSLOQuery(dd.Query.Denominator)
			dt.MetricExpression = fmt.Sprintf("(100)*(%s)/(%s)", numerator, denominator)
		}
	case "monitor":
		// Monitor-based SLOs don't have a direct DT equivalent with metric expressions
		// We create a placeholder that needs manual configuration
		dt.MetricExpression = "builtin:synthetic.browser.availability.location.totalPerformance"
		dt.Description += "\n\n[Migration note: This was a monitor-based SLO in DataDog. The metric expression needs manual configuration.]"
	default:
		return nil, fmt.Errorf("unsupported SLO type: %s", dd.Type)
	}

	return dt, nil
}

func translateSLOQuery(ddQuery string) string {
	parsed, err := query.Parse(ddQuery)
	if err != nil {
		return ddQuery
	}
	return query.ToMetricSelector(parsed)
}

func mapSLOTimeframe(ddTimeframe string) string {
	switch ddTimeframe {
	case "7d":
		return "-1w"
	case "30d":
		return "-1M"
	case "90d":
		return "-3M"
	case "custom":
		return "-1M"
	default:
		return "-1M"
	}
}
