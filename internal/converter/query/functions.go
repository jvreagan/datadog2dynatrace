package query

// MapAggregation maps a DataDog aggregation to a Dynatrace equivalent.
func MapAggregation(ddAgg string) string {
	aggMap := map[string]string{
		"avg":   "avg",
		"sum":   "sum",
		"min":   "min",
		"max":   "max",
		"count": "count",
		"last":  "value",
		"p50":   "percentile(50)",
		"p75":   "percentile(75)",
		"p90":   "percentile(90)",
		"p95":   "percentile(95)",
		"p99":   "percentile(99)",
	}

	if dt, ok := aggMap[ddAgg]; ok {
		return dt
	}
	return "avg" // default
}

// MapFunction maps a DataDog function to a Dynatrace metric selector function.
func MapFunction(ddFunc string) string {
	funcMap := map[string]string{
		// Arithmetic
		"abs":   "abs",
		"log2":  "log2",
		"log10": "log10",
		"ceil":  "ceil",
		"floor": "floor",
		"round": "round",
		// Smoothing
		"ewma_3":  "smooth",
		"ewma_5":  "smooth",
		"ewma_10": "smooth",
		"ewma_20": "smooth",
		"median_3": "smooth",
		"median_5": "smooth",
		// Rollup
		"rollup": "fold",
		// Rate
		"per_second": "rate",
		"per_minute": "rate",
		"per_hour":   "rate",
		// Time shift
		"timeshift": "timeshift",
		// Top/bottom
		"top":    "sort",
		"bottom": "sort",
		// Count
		"count_nonzero":  "count",
		"count_not_null": "count",
		// Forecast / anomaly
		"forecast":  "", // No direct DT equivalent
		"anomalies": "", // Handled by DT's built-in anomaly detection
		// Diff
		"diff":       "delta",
		"derivative": "rate",
		"dt":         "rate",
		// Cumulative
		"cumsum": "rollup",
		// Clamp
		"clamp_min": "partition",
		"clamp_max": "partition",
	}

	if dt, ok := funcMap[ddFunc]; ok {
		return dt
	}
	return "" // unsupported
}

// MapRollupFunction maps DD rollup functions to DT fold aggregation.
func MapRollupFunction(ddRollup string) string {
	rollupMap := map[string]string{
		"avg":   "avg",
		"sum":   "sum",
		"min":   "min",
		"max":   "max",
		"count": "count",
	}

	if dt, ok := rollupMap[ddRollup]; ok {
		return dt
	}
	return "avg"
}

// MapAlertCondition maps a DD monitor comparison to a DT alert condition.
func MapAlertCondition(ddOperator string) string {
	condMap := map[string]string{
		"above":          "ABOVE",
		"above_or_equal": "ABOVE",
		"below":          "BELOW",
		"below_or_equal": "BELOW",
		">":              "ABOVE",
		">=":             "ABOVE",
		"<":              "BELOW",
		"<=":             "BELOW",
	}

	if dt, ok := condMap[ddOperator]; ok {
		return dt
	}
	return "ABOVE"
}
