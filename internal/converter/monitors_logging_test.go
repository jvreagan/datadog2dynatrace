package converter

import (
	"strings"
	"testing"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/datadog"
)

func TestLoggingMonitorQueryParsing(t *testing.T) {
	buf := setupLogging(t)

	dd := &datadog.Monitor{
		Name:  "High CPU",
		Type:  "metric alert",
		Query: "avg(last_5m):avg:system.cpu.user{*} > 90",
	}

	ConvertMonitor(dd)

	out := buf.String()
	if !strings.Contains(out, "[DEBUG] parsing monitor query") {
		t.Errorf("expected debug log for parsing monitor query, got:\n%s", out)
	}
}

func TestLoggingMonitorQueryFallback(t *testing.T) {
	buf := setupLogging(t)

	dd := &datadog.Monitor{
		Name:  "Bad Query Monitor",
		Type:  "query alert",
		Query: "avg(last_5m):metric{unclosed > 10",
	}

	ConvertMonitor(dd)

	out := buf.String()
	if !strings.Contains(out, "[WARN] query parse failed, falling back to raw string") {
		t.Errorf("expected warn log for query parse fallback, got:\n%s", out)
	}
}
