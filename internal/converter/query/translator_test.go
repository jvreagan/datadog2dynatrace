package query

import (
	"strings"
	"testing"
)

func TestToMetricSelector(t *testing.T) {
	tests := []struct {
		name     string
		pq       *ParsedQuery
		contains []string
		notContains []string
	}{
		{
			name: "nil returns empty",
			pq:   nil,
		},
		{
			name: "simple metric with aggregation",
			pq: &ParsedQuery{
				Metric:      "system.cpu.user",
				Aggregation: "avg",
				Filters:     map[string]string{},
			},
			contains: []string{"builtin:host.cpu.user", ":avg"},
		},
		{
			name: "metric with positive filter",
			pq: &ParsedQuery{
				Metric:      "system.cpu.user",
				Aggregation: "avg",
				Filters:     map[string]string{"host": "web01"},
			},
			contains: []string{":filter(", "eq(dt.entity.host,\"web01\")"},
		},
		{
			name: "metric with negated filter",
			pq: &ParsedQuery{
				Metric:      "system.cpu.user",
				Aggregation: "avg",
				Filters:     map[string]string{"!env": "staging"},
			},
			contains: []string{"ne(environment,\"staging\")"},
		},
		{
			name: "metric with group by",
			pq: &ParsedQuery{
				Metric:      "system.cpu.user",
				Aggregation: "avg",
				Filters:     map[string]string{},
				GroupBy:     []string{"host"},
			},
			contains: []string{":splitBy(\"dt.entity.host\")", ":avg"},
		},
		{
			name: "rollup generates fold",
			pq: &ParsedQuery{
				Metric:  "my.metric",
				Filters: map[string]string{},
				Rollup:  &RollupDef{Method: "sum", Period: 0},
			},
			contains: []string{":fold(sum)"},
		},
		{
			name: "rollup with period emits fold with period",
			pq: &ParsedQuery{
				Metric:  "my.metric",
				Filters: map[string]string{},
				Rollup:  &RollupDef{Method: "sum", Period: 60},
			},
			contains: []string{":fold(sum,60)"},
		},
		{
			name: "as_rate generates rate",
			pq: &ParsedQuery{
				Metric:     "my.metric",
				Filters:    map[string]string{},
				AsModifier: "rate",
			},
			contains: []string{":rate"},
		},
		{
			name: "top function generates sort and limit",
			pq: &ParsedQuery{
				Metric:   "system.cpu.user",
				Filters:  map[string]string{},
				Function: "top",
				FuncArgs: []string{"10", "mean", "desc"},
			},
			contains: []string{":sort(value(avg,descending)):limit(10)"},
		},
		{
			name: "bottom function generates ascending sort",
			pq: &ParsedQuery{
				Metric:   "system.cpu.user",
				Filters:  map[string]string{},
				Function: "bottom",
				FuncArgs: []string{"5"},
			},
			contains: []string{":sort(value(avg,ascending)):limit(5)"},
		},
		{
			name: "per_second function generates rate",
			pq: &ParsedQuery{
				Metric:   "my.counter",
				Filters:  map[string]string{},
				Function: "per_second",
			},
			contains: []string{":rate"},
		},
		{
			name: "custom metric gets ext: prefix",
			pq: &ParsedQuery{
				Metric:  "custom.my_app.request_count",
				Filters: map[string]string{},
			},
			contains: []string{"ext:custom.my_app.request_count"},
		},
		{
			name: "unknown metric gets ext:custom. prefix",
			pq: &ParsedQuery{
				Metric:  "myapp.requests.count",
				Filters: map[string]string{},
			},
			contains: []string{"ext:custom.myapp.requests.count"},
		},
		{
			name: "filter keys translated correctly",
			pq: &ParsedQuery{
				Metric: "system.cpu.user",
				Filters: map[string]string{
					"pod_name":  "web-abc",
					"namespace": "production",
					"cluster":   "main",
				},
			},
			contains: []string{"k8s.pod.name", "k8s.namespace.name", "k8s.cluster.name"},
		},
		{
			name: "multiple filters use and()",
			pq: &ParsedQuery{
				Metric: "system.cpu.user",
				Filters: map[string]string{
					"host": "web01",
					"env":  "prod",
				},
			},
			contains: []string{":filter(and("},
		},
		{
			name: "single filter no and()",
			pq: &ParsedQuery{
				Metric:  "system.cpu.user",
				Filters: map[string]string{"host": "web01"},
			},
			contains:    []string{":filter(eq("},
			notContains: []string{":filter(and("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToMetricSelector(tt.pq)

			if tt.pq == nil {
				if result != "" {
					t.Errorf("expected empty string, got %q", result)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got %q", s, result)
				}
			}
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("expected result NOT to contain %q, got %q", s, result)
				}
			}
		})
	}
}

func TestTranslateMetricName(t *testing.T) {
	tests := map[string]string{
		"system.cpu.user":                  "builtin:host.cpu.user",
		"system.cpu.system":                "builtin:host.cpu.system",
		"system.mem.used":                  "builtin:host.mem.usage",
		"system.disk.used":                 "builtin:host.disk.usedPct",
		"system.net.bytes_rcvd":            "builtin:host.net.nic.trafficIn",
		"docker.cpu.usage":                 "builtin:containers.cpu.usagePercent",
		"docker.mem.rss":                   "builtin:containers.memory.residentSetSize",
		"kubernetes.cpu.usage.total":       "builtin:cloud.kubernetes.workload.cpuUsage",
		"kubernetes.memory.usage":          "builtin:cloud.kubernetes.workload.memoryUsage",
		"kubernetes.pods.running":          "builtin:cloud.kubernetes.workload.pods",
		"trace.servlet.request.hits":       "builtin:service.requestCount.total",
		"trace.servlet.request.duration":   "builtin:service.response.time",
		"trace.http.request.errors":        "builtin:service.errors.total.count",
		"custom.my_metric":                 "ext:custom.my_metric",
		"system.swap.used":                 "builtin:host.mem.swap.used",
		"kubernetes.containers.restarts":   "builtin:cloud.kubernetes.workload.containerRestarts",
	}

	for dd, expected := range tests {
		t.Run(dd, func(t *testing.T) {
			result := TranslateMetricName(dd)
			if result != expected {
				t.Errorf("TranslateMetricName(%q) = %q, want %q", dd, result, expected)
			}
		})
	}
}

func TestToDQL(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		contains []string
	}{
		{
			name:     "empty returns empty",
			query:    "",
			contains: nil,
		},
		{
			name:     "source filter",
			query:    "source:nginx",
			contains: []string{"fetch logs", "dt.entity.process_group_instance == \"nginx\""},
		},
		{
			name:     "status error",
			query:    "status:error",
			contains: []string{"loglevel == \"ERROR\""},
		},
		{
			name:     "status warning",
			query:    "status:warn",
			contains: []string{"loglevel == \"WARN\""},
		},
		{
			name:     "combined filters",
			query:    "source:nginx status:error",
			contains: []string{"dt.entity.process_group_instance", "loglevel == \"ERROR\""},
		},
		{
			name:     "http method",
			query:    "@http.method:GET",
			contains: []string{"http.method == \"GET\""},
		},
		{
			name:     "negated filter",
			query:    "-source:debug",
			contains: []string{"NOT("},
		},
		{
			name:     "custom attribute",
			query:    "@custom_field:myvalue",
			contains: []string{"custom_field == \"myvalue\""},
		},
		{
			name:     "wildcard filter",
			query:    "source:web*",
			contains: []string{"LIKE", "web%"},
		},
		{
			name:     "plain text search",
			query:    "OutOfMemoryError",
			contains: []string{"content CONTAINS \"OutOfMemoryError\""},
		},
		{
			name:     "service filter",
			query:    "service:payment-api",
			contains: []string{"dt.entity.service == \"payment-api\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToDQL(tt.query)

			if tt.query == "" {
				if result != "" {
					t.Errorf("expected empty, got %q", result)
				}
				return
			}

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}

func TestEndToEndQueryTranslation(t *testing.T) {
	// Full pipeline: raw DD query → parse → translate → DT metric selector
	tests := []struct {
		name     string
		ddQuery  string
		contains []string
	}{
		{
			name:    "simple CPU avg",
			ddQuery: "avg:system.cpu.user{*}",
			contains: []string{"builtin:host.cpu.user", ":avg"},
		},
		{
			name:    "filtered with group by",
			ddQuery: "avg:system.cpu.user{env:prod} by {host}",
			contains: []string{
				"builtin:host.cpu.user",
				"eq(environment,\"prod\")",
				":splitBy(\"dt.entity.host\")",
				":avg",
			},
		},
		{
			name:    "kubernetes metric",
			ddQuery: "avg:kubernetes.cpu.usage.total{kube_cluster_name:prod} by {kube_namespace}",
			contains: []string{
				"builtin:cloud.kubernetes.workload.cpuUsage",
				"k8s.cluster.name",
				"k8s.namespace.name",
			},
		},
		{
			name:    "per_second wrapper",
			ddQuery: "per_second(sum:system.net.bytes_rcvd{*})",
			contains: []string{
				"builtin:host.net.nic.trafficIn",
				":sum",
				":rate",
			},
		},
		{
			name:    "rollup and fill",
			ddQuery: "avg:system.cpu.user{*}.rollup(max, 300).fill(zero)",
			contains: []string{
				"builtin:host.cpu.user",
				":fold(max,300)",
			},
		},
		{
			name:    "custom metric with as_count",
			ddQuery: "sum:custom.orders.placed{env:prod}.as_count()",
			contains: []string{
				"ext:custom.orders.placed",
				":sum",
			},
		},
		{
			name:    "negated filter",
			ddQuery: "avg:system.cpu.user{env:prod,!host:debug01}",
			contains: []string{
				"eq(environment,\"prod\")",
				"ne(dt.entity.host,\"debug01\")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Parse(tt.ddQuery)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.ddQuery, err)
			}

			result := ToMetricSelector(pq)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("expected result to contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}
