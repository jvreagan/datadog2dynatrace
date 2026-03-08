package query

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

func setupLogging(t *testing.T) *bytes.Buffer {
	var buf bytes.Buffer
	logging.SetOutput(&buf)
	logging.SetLevel(logging.LevelDebug)
	t.Cleanup(func() {
		logging.SetOutput(os.Stderr)
		logging.SetLevel(logging.LevelWarn)
	})
	return &buf
}

func TestLoggingMetricNameTranslation(t *testing.T) {
	buf := setupLogging(t)

	pq := &ParsedQuery{
		Metric:      "system.cpu.user",
		Aggregation: "avg",
		Filters:     map[string]string{},
	}

	ToMetricSelector(pq)

	out := buf.String()
	if !strings.Contains(out, "[DEBUG] metric name translated: system.cpu.user -> builtin:host.cpu.user") {
		t.Errorf("expected debug log for metric translation, got:\n%s", out)
	}
}

func TestLoggingAggregationTranslation(t *testing.T) {
	buf := setupLogging(t)

	pq := &ParsedQuery{
		Metric:      "system.cpu.user",
		Aggregation: "last",
		Filters:     map[string]string{},
	}

	ToMetricSelector(pq)

	out := buf.String()
	// "last" translates to "value", so it should log
	if !strings.Contains(out, "[DEBUG] aggregation translated: last -> value") {
		t.Errorf("expected debug log for aggregation translation, got:\n%s", out)
	}
}

func TestLoggingUnsupportedFunction(t *testing.T) {
	buf := setupLogging(t)

	pq := &ParsedQuery{
		Metric:   "system.cpu.user",
		Filters:  map[string]string{},
		Function: "forecast",
	}

	ToMetricSelector(pq)

	out := buf.String()
	if !strings.Contains(out, `[DEBUG] function "forecast" has no DT equivalent, skipping`) {
		t.Errorf("expected debug log for unsupported function, got:\n%s", out)
	}
}
