package query

import (
	"testing"
)

// filterMap converts []FilterTerm to a map for backward-compatible assertions.
// Negated keys are prefixed with "!".
func filterMap(filters []FilterTerm) map[string]string {
	m := make(map[string]string)
	for _, ft := range filters {
		key := ft.Key
		if ft.Negated {
			key = "!" + key
		}
		m[key] = ft.Value
	}
	return m
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantAgg     string
		wantMetric  string
		wantFunc    string
		wantFilters map[string]string
		wantGroupBy []string
		wantRollup  *RollupDef
		wantAs      string
		wantFill    string
		wantErr     bool
	}{
		{
			name:       "simple metric with wildcard",
			query:      "avg:system.cpu.user{*}",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
		},
		{
			name:       "metric with filters",
			query:      "sum:custom.metric{host:web01,env:prod}",
			wantAgg:    "sum",
			wantMetric: "custom.metric",
			wantFilters: map[string]string{
				"host": "web01",
				"env":  "prod",
			},
		},
		{
			name:        "metric with group by",
			query:       "avg:system.cpu.user{*}by{host}",
			wantAgg:     "avg",
			wantMetric:  "system.cpu.user",
			wantGroupBy: []string{"host"},
		},
		{
			name:        "metric with spaced group by",
			query:       "avg:system.cpu.user{*} by {host, region}",
			wantAgg:     "avg",
			wantMetric:  "system.cpu.user",
			wantGroupBy: []string{"host", "region"},
		},
		{
			name:        "filters and group by",
			query:       "avg:system.cpu.user{env:prod} by {host,region}",
			wantAgg:     "avg",
			wantMetric:  "system.cpu.user",
			wantFilters: map[string]string{"env": "prod"},
			wantGroupBy: []string{"host", "region"},
		},
		{
			name:       "no aggregation prefix",
			query:      "system.cpu.user{host:web01}",
			wantAgg:    "",
			wantMetric: "system.cpu.user",
			wantFilters: map[string]string{"host": "web01"},
		},
		{
			name:       "wrapping function - top",
			query:      "top(avg:system.cpu.user{*} by {host}, 10, 'mean', 'desc')",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantFunc:   "top",
			wantGroupBy: []string{"host"},
		},
		{
			name:       "wrapping function - per_second",
			query:      "per_second(sum:my.counter{*})",
			wantAgg:    "sum",
			wantMetric: "my.counter",
			wantFunc:   "per_second",
		},
		{
			name:       "wrapping function - abs",
			query:      "abs(avg:temperature.delta{*})",
			wantAgg:    "avg",
			wantMetric: "temperature.delta",
			wantFunc:   "abs",
		},
		{
			name:       "nested function",
			query:      "per_second(avg:system.net.bytes_rcvd{host:web01})",
			wantAgg:    "avg",
			wantMetric: "system.net.bytes_rcvd",
			wantFunc:   "per_second",
			wantFilters: map[string]string{"host": "web01"},
		},
		{
			name:       "as_count modifier",
			query:      "sum:my.metric{*}.as_count()",
			wantAgg:    "sum",
			wantMetric: "my.metric",
			wantAs:     "count",
		},
		{
			name:       "as_rate modifier",
			query:      "sum:my.metric{*}.as_rate()",
			wantAgg:    "sum",
			wantMetric: "my.metric",
			wantAs:     "rate",
		},
		{
			name:       "rollup with method only",
			query:      "avg:system.cpu.user{*}.rollup(sum)",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantRollup: &RollupDef{Method: "sum"},
		},
		{
			name:       "rollup with method and period",
			query:      "avg:system.cpu.user{*}.rollup(avg, 60)",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantRollup: &RollupDef{Method: "avg", Period: 60},
		},
		{
			name:       "fill modifier",
			query:      "avg:my.metric{*}.fill(zero)",
			wantAgg:    "avg",
			wantMetric: "my.metric",
			wantFill:   "zero",
		},
		{
			name:       "fill last",
			query:      "avg:my.metric{*}.fill(last)",
			wantAgg:    "avg",
			wantMetric: "my.metric",
			wantFill:   "last",
		},
		{
			name:       "combined modifiers: rollup + fill",
			query:      "avg:my.metric{*}.rollup(sum, 300).fill(zero)",
			wantAgg:    "avg",
			wantMetric: "my.metric",
			wantRollup: &RollupDef{Method: "sum", Period: 300},
			wantFill:   "zero",
		},
		{
			name:       "combined: function + as_count",
			query:      "per_second(sum:my.counter{*}.as_count())",
			wantAgg:    "sum",
			wantMetric: "my.counter",
			wantFunc:   "per_second",
			wantAs:     "count",
		},
		{
			name:       "negated filter",
			query:      "avg:system.cpu.user{!env:staging}",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantFilters: map[string]string{"!env": "staging"},
		},
		{
			name:       "multiple filters with negation",
			query:      "avg:system.cpu.user{env:prod,!host:debug01}",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantFilters: map[string]string{
				"env":   "prod",
				"!host": "debug01",
			},
		},
		{
			name:       "bare metric name without braces",
			query:      "avg:system.cpu.user",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
		},
		{
			name:       "percentile aggregation",
			query:      "p99:trace.servlet.request.duration{env:prod}",
			wantAgg:    "p99",
			wantMetric: "trace.servlet.request.duration",
			wantFilters: map[string]string{"env": "prod"},
		},
		{
			name:       "count aggregation",
			query:      "count:my.events{*}",
			wantAgg:    "count",
			wantMetric: "my.events",
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			query:   "   ",
			wantErr: true,
		},
		{
			name:       "real-world: kubernetes CPU with group by",
			query:      "avg:kubernetes.cpu.usage.total{kube_cluster_name:prod} by {kube_namespace}",
			wantAgg:    "avg",
			wantMetric: "kubernetes.cpu.usage.total",
			wantFilters: map[string]string{"kube_cluster_name": "prod"},
			wantGroupBy: []string{"kube_namespace"},
		},
		{
			name:       "real-world: docker mem with rollup and group by",
			query:      "avg:docker.mem.rss{env:prod}.rollup(max, 300) by {container_name}",
			wantAgg:    "avg",
			wantMetric: "docker.mem.rss",
			wantFilters: map[string]string{"env": "prod"},
			wantRollup: &RollupDef{Method: "max", Period: 300},
			wantGroupBy: []string{"container_name"},
		},
		{
			name:       "real-world: top function",
			query:      "top(avg:system.cpu.user{*} by {host}, 25, 'mean', 'desc')",
			wantAgg:    "avg",
			wantMetric: "system.cpu.user",
			wantFunc:   "top",
			wantGroupBy: []string{"host"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Parse(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pq.Aggregation != tt.wantAgg {
				t.Errorf("aggregation: got %q, want %q", pq.Aggregation, tt.wantAgg)
			}
			if pq.Metric != tt.wantMetric {
				t.Errorf("metric: got %q, want %q", pq.Metric, tt.wantMetric)
			}
			if pq.Function != tt.wantFunc {
				t.Errorf("function: got %q, want %q", pq.Function, tt.wantFunc)
			}
			if pq.AsModifier != tt.wantAs {
				t.Errorf("as modifier: got %q, want %q", pq.AsModifier, tt.wantAs)
			}
			if pq.Fill != tt.wantFill {
				t.Errorf("fill: got %q, want %q", pq.Fill, tt.wantFill)
			}

			// Check filters
			if tt.wantFilters != nil {
				fm := filterMap(pq.Filters)
				for k, v := range tt.wantFilters {
					if fm[k] != v {
						t.Errorf("filter %q: got %q, want %q", k, fm[k], v)
					}
				}
				if len(pq.Filters) != len(tt.wantFilters) {
					t.Errorf("filter count: got %d, want %d (%v)", len(pq.Filters), len(tt.wantFilters), pq.Filters)
				}
			}

			// Check group by
			if tt.wantGroupBy != nil {
				if len(pq.GroupBy) != len(tt.wantGroupBy) {
					t.Errorf("group by count: got %d (%v), want %d (%v)", len(pq.GroupBy), pq.GroupBy, len(tt.wantGroupBy), tt.wantGroupBy)
				} else {
					for i, g := range tt.wantGroupBy {
						if pq.GroupBy[i] != g {
							t.Errorf("group by[%d]: got %q, want %q", i, pq.GroupBy[i], g)
						}
					}
				}
			}

			// Check rollup
			if tt.wantRollup != nil {
				if pq.Rollup == nil {
					t.Error("expected rollup, got nil")
				} else {
					if pq.Rollup.Method != tt.wantRollup.Method {
						t.Errorf("rollup method: got %q, want %q", pq.Rollup.Method, tt.wantRollup.Method)
					}
					if pq.Rollup.Period != tt.wantRollup.Period {
						t.Errorf("rollup period: got %d, want %d", pq.Rollup.Period, tt.wantRollup.Period)
					}
				}
			}
		})
	}
}

func TestParseLogQuery(t *testing.T) {
	input := "source:nginx status:error"
	got := ParseLogQuery(input)
	if got != input {
		t.Errorf("ParseLogQuery(%q) = %q, want %q", input, got, input)
	}
}

func TestFindMatchingBraceUnclosed(t *testing.T) {
	got := findMatchingBrace("{abc", 0)
	if got != -1 {
		t.Errorf("findMatchingBrace unclosed = %d, want -1", got)
	}
}

func TestParseUnclosedBrace(t *testing.T) {
	_, err := Parse("avg:system.cpu.user{host:web01")
	if err == nil {
		t.Error("expected error for unclosed brace")
	}
}

func TestParseORFilters(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantFilters []FilterTerm
	}{
		{
			name:  "simple OR filter",
			query: "avg:system.cpu.user{host:web01 OR host:web02}",
			wantFilters: []FilterTerm{
				{Key: "host", Value: "web01", Operator: "AND"},
				{Key: "host", Value: "web02", Operator: "OR"},
			},
		},
		{
			name:  "mixed AND and OR",
			query: "avg:system.cpu.user{host:web01,env:prod OR env:staging}",
			wantFilters: []FilterTerm{
				{Key: "host", Value: "web01", Operator: "AND"},
				{Key: "env", Value: "prod", Operator: "AND"},
				{Key: "env", Value: "staging", Operator: "OR"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pq.Filters) != len(tt.wantFilters) {
				t.Fatalf("filter count: got %d, want %d", len(pq.Filters), len(tt.wantFilters))
			}
			for i, want := range tt.wantFilters {
				got := pq.Filters[i]
				if got.Key != want.Key || got.Value != want.Value || got.Negated != want.Negated || got.Operator != want.Operator {
					t.Errorf("filter[%d]: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}
