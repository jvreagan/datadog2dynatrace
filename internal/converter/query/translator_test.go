package query

import (
	"strings"
	"testing"
)

func TestToMetricSelectorSimple(t *testing.T) {
	pq := &ParsedQuery{
		Metric:      "system.cpu.user",
		Aggregation: "avg",
	}
	result := ToMetricSelector(pq)
	if !strings.Contains(result, "builtin:host.cpu.user") {
		t.Errorf("expected builtin metric, got %q", result)
	}
	if !strings.Contains(result, ":avg") {
		t.Errorf("expected avg aggregation, got %q", result)
	}
}

func TestToMetricSelectorWithFilters(t *testing.T) {
	pq := &ParsedQuery{
		Metric:      "system.cpu.user",
		Aggregation: "avg",
		Filters:     map[string]string{"host": "web01"},
	}
	result := ToMetricSelector(pq)
	if !strings.Contains(result, ":filter") {
		t.Errorf("expected filter in result, got %q", result)
	}
	if !strings.Contains(result, "dt.entity.host") {
		t.Errorf("expected translated filter key, got %q", result)
	}
}

func TestToMetricSelectorWithGroupBy(t *testing.T) {
	pq := &ParsedQuery{
		Metric:      "system.cpu.user",
		Aggregation: "avg",
		GroupBy:     []string{"host"},
	}
	result := ToMetricSelector(pq)
	if !strings.Contains(result, ":splitBy") {
		t.Errorf("expected splitBy in result, got %q", result)
	}
}

func TestTranslateMetricNameBuiltin(t *testing.T) {
	tests := map[string]string{
		"system.cpu.user":       "builtin:host.cpu.user",
		"system.mem.used":       "builtin:host.mem.usage",
		"docker.cpu.usage":      "builtin:containers.cpu.usagePercent",
	}
	for dd, expected := range tests {
		result := TranslateMetricName(dd)
		if result != expected {
			t.Errorf("TranslateMetricName(%q) = %q, want %q", dd, result, expected)
		}
	}
}

func TestTranslateMetricNameCustom(t *testing.T) {
	result := TranslateMetricName("custom.my_metric")
	if !strings.HasPrefix(result, "ext:") {
		t.Errorf("expected ext: prefix for custom metric, got %q", result)
	}
}

func TestToMetricSelectorNil(t *testing.T) {
	result := ToMetricSelector(nil)
	if result != "" {
		t.Errorf("expected empty string for nil query, got %q", result)
	}
}

func TestToDQL(t *testing.T) {
	result := ToDQL("source:nginx status:error")
	if !strings.Contains(result, "fetch logs") {
		t.Errorf("expected DQL to start with 'fetch logs', got %q", result)
	}
	if !strings.Contains(result, "ERROR") {
		t.Errorf("expected ERROR in DQL filter, got %q", result)
	}
}

func TestToDQLEmpty(t *testing.T) {
	result := ToDQL("")
	if result != "" {
		t.Errorf("expected empty string for empty query, got %q", result)
	}
}
