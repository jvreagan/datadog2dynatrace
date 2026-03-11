package query

import "testing"

func TestMapAlertCondition(t *testing.T) {
	tests := map[string]string{
		"above":          "ABOVE",
		"above_or_equal": "ABOVE",
		"below":          "BELOW",
		"below_or_equal": "BELOW",
		">":              "ABOVE",
		">=":             "ABOVE",
		"<":              "BELOW",
		"<=":             "BELOW",
		"unknown":        "ABOVE", // default
	}
	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			if got := MapAlertCondition(input); got != expected {
				t.Errorf("MapAlertCondition(%q) = %q, want %q", input, got, expected)
			}
		})
	}
}

func TestMapAggregationDefault(t *testing.T) {
	if got := MapAggregation("unknown_agg"); got != "avg" {
		t.Errorf("MapAggregation(unknown_agg) = %q, want %q", got, "avg")
	}
}

func TestMapFunctionUnsupported(t *testing.T) {
	if got := MapFunction("totally_unknown"); got != "" {
		t.Errorf("MapFunction(totally_unknown) = %q, want empty", got)
	}
}

func TestMapRollupFunctionDefault(t *testing.T) {
	if got := MapRollupFunction("unknown_rollup"); got != "avg" {
		t.Errorf("MapRollupFunction(unknown_rollup) = %q, want %q", got, "avg")
	}
}
