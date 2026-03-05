package query

import (
	"testing"
)

func TestParseSimpleMetric(t *testing.T) {
	pq, err := Parse("avg:system.cpu.user{*}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pq.Aggregation != "avg" {
		t.Errorf("expected aggregation 'avg', got %q", pq.Aggregation)
	}
	if pq.Metric != "system.cpu.user" {
		t.Errorf("expected metric 'system.cpu.user', got %q", pq.Metric)
	}
}

func TestParseWithFilters(t *testing.T) {
	pq, err := Parse("sum:custom.metric{host:web01,env:prod}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pq.Aggregation != "sum" {
		t.Errorf("expected aggregation 'sum', got %q", pq.Aggregation)
	}
	if pq.Metric != "custom.metric" {
		t.Errorf("expected metric 'custom.metric', got %q", pq.Metric)
	}
	if pq.Filters["host"] != "web01" {
		t.Errorf("expected filter host=web01, got %q", pq.Filters["host"])
	}
	if pq.Filters["env"] != "prod" {
		t.Errorf("expected filter env=prod, got %q", pq.Filters["env"])
	}
}

func TestParseWithGroupBy(t *testing.T) {
	pq, err := Parse("avg:system.cpu.user{*}by{host}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pq.GroupBy) != 1 || pq.GroupBy[0] != "host" {
		t.Errorf("expected group by [host], got %v", pq.GroupBy)
	}
}

func TestParseWithFiltersAndGroupBy(t *testing.T) {
	pq, err := Parse("avg:system.cpu.user{env:prod}by{host,region}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pq.Filters["env"] != "prod" {
		t.Errorf("expected filter env=prod, got %q", pq.Filters["env"])
	}
	if len(pq.GroupBy) != 2 {
		t.Errorf("expected 2 group by dimensions, got %d", len(pq.GroupBy))
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestParseNoAggregation(t *testing.T) {
	pq, err := Parse("system.cpu.user{host:web01}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pq.Aggregation != "" {
		t.Errorf("expected empty aggregation, got %q", pq.Aggregation)
	}
	if pq.Metric != "system.cpu.user" {
		t.Errorf("expected metric 'system.cpu.user', got %q", pq.Metric)
	}
}
