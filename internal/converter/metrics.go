package converter

import (
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/converter/query"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
	"github.com/datadog2dynatrace/datadog2dynatrace/internal/dynatrace"
)

// ConvertMetricMetadata converts DataDog metric metadata to a Dynatrace metric descriptor.
func ConvertMetricMetadata(dd *datadog.MetricMetadata) (*dynatrace.MetricDescriptor, error) {
	return &dynatrace.MetricDescriptor{
		MetricID:    query.TranslateMetricName(dd.Metric),
		DisplayName: metricDisplayName(dd),
		Description: dd.Description,
		Unit:        mapMetricUnit(dd.Unit, dd.PerUnit),
	}, nil
}

func metricDisplayName(dd *datadog.MetricMetadata) string {
	if dd.ShortName != "" {
		return dd.ShortName
	}
	return dd.Metric
}

func mapMetricUnit(unit, perUnit string) string {
	unitMap := map[string]string{
		"byte":        "Byte",
		"kilobyte":    "KiloByte",
		"megabyte":    "MegaByte",
		"gigabyte":    "GigaByte",
		"terabyte":    "TeraByte",
		"percent":     "Percent",
		"nanosecond":  "NanoSecond",
		"microsecond": "MicroSecond",
		"millisecond": "MilliSecond",
		"second":      "Second",
		"minute":      "Minute",
		"hour":        "Hour",
		"day":         "Day",
		"bit":         "Bit",
		"kilobit":     "KiloBit",
		"megabit":     "MegaBit",
		"gigabit":     "GigaBit",
		"count":       "Count",
		"operation":   "Count",
		"request":     "Count",
		"error":       "Count",
		"connection":  "Count",
	}

	dtUnit := "Unspecified"
	if mapped, ok := unitMap[unit]; ok {
		dtUnit = mapped
	}

	if perUnit != "" {
		if mapped, ok := unitMap[perUnit]; ok {
			dtUnit += "Per" + mapped
		}
	}

	return dtUnit
}
