package converter

import (
	"fmt"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertSLO converts a DataDog SLO to one Dynatrace SLO per threshold.
// If the SLO has multiple thresholds, each DT SLO is suffixed with the timeframe
// (e.g. "API Availability (7d)"). A single threshold produces no suffix.
func ConvertSLO(dd *datadog.SLO) ([]dynatrace.SLO, error) {
	// Build the metric expression once — it's the same across all thresholds.
	var metricExpr string
	var descExtra string

	switch dd.Type {
	case "metric":
		if dd.Query != nil {
			numerator := translateSLOQuery(dd.Query.Numerator)
			denominator := translateSLOQuery(dd.Query.Denominator)
			metricExpr = fmt.Sprintf("(100)*(%s)/(%s)", numerator, denominator)
		}
	case "monitor":
		metricExpr = "builtin:synthetic.browser.availability.location.totalPerformance"
		descExtra = "\n\n[Migration note: This was a monitor-based SLO in DataDog. The metric expression needs manual configuration.]"
	default:
		return nil, fmt.Errorf("unsupported SLO type: %s", dd.Type)
	}

	// No thresholds: produce a single SLO with defaults.
	if len(dd.Thresholds) == 0 {
		slo := dynatrace.SLO{
			Name:             dd.Name,
			Description:      dd.Description + descExtra,
			EvaluationType:   "AGGREGATE",
			Enabled:          true,
			Target:           99.0,
			Warning:          99.5,
			Timeframe:        "-1M",
			MetricExpression: metricExpr,
		}
		return []dynatrace.SLO{slo}, nil
	}

	multi := len(dd.Thresholds) > 1
	result := make([]dynatrace.SLO, 0, len(dd.Thresholds))

	for _, thr := range dd.Thresholds {
		name := dd.Name
		if multi {
			name = fmt.Sprintf("%s (%s)", dd.Name, thr.Timeframe)
		}
		warning := thr.Warning
		if warning <= thr.Target {
			warning = thr.Target + 0.5
		}
		slo := dynatrace.SLO{
			Name:             name,
			Description:      dd.Description + descExtra,
			EvaluationType:   "AGGREGATE",
			Enabled:          true,
			Target:           thr.Target,
			Warning:          warning,
			Timeframe:        mapSLOTimeframe(thr.Timeframe),
			MetricExpression: metricExpr,
		}
		result = append(result, slo)
	}

	return result, nil
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
