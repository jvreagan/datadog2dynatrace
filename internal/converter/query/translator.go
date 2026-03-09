package query

import (
	"fmt"
	"strings"

	"github.com/datadog2dynatrace/datadog2dynatrace/internal/logging"
)

// ToMetricSelector converts a parsed DataDog query to a Dynatrace metric selector.
func ToMetricSelector(pq *ParsedQuery) string {
	if pq == nil {
		return ""
	}

	metric := TranslateMetricName(pq.Metric)
	if metric != pq.Metric {
		logging.Debug("metric name translated: %s -> %s", pq.Metric, metric)
	}

	var sb strings.Builder
	sb.WriteString(metric)

	// Add filters
	if len(pq.Filters) > 0 {
		// Build filter groups: consecutive OR terms form a group with the preceding AND term.
		var topLevel []string // AND-combined top-level expressions
		i := 0
		for i < len(pq.Filters) {
			ft := pq.Filters[i]
			expr := filterTermExpr(ft)

			// Check if the next term(s) are OR — if so, build an or() group
			if i+1 < len(pq.Filters) && pq.Filters[i+1].Operator == "OR" {
				orGroup := []string{expr}
				i++
				for i < len(pq.Filters) && pq.Filters[i].Operator == "OR" {
					orGroup = append(orGroup, filterTermExpr(pq.Filters[i]))
					i++
				}
				topLevel = append(topLevel, fmt.Sprintf("or(%s)", strings.Join(orGroup, ",")))
			} else {
				topLevel = append(topLevel, expr)
				i++
			}
		}

		var filterExpr string
		if len(topLevel) == 1 {
			filterExpr = topLevel[0]
		} else {
			filterExpr = fmt.Sprintf("and(%s)", strings.Join(topLevel, ","))
		}
		sb.WriteString(fmt.Sprintf(":filter(%s)", filterExpr))
	}

	// Add split by (group by) — must come before aggregation in DT selectors
	if len(pq.GroupBy) > 0 {
		var dims []string
		for _, g := range pq.GroupBy {
			dims = append(dims, fmt.Sprintf("\"%s\"", translateFilterKey(g)))
		}
		sb.WriteString(fmt.Sprintf(":splitBy(%s)", strings.Join(dims, ",")))
	} else {
		sb.WriteString(":splitBy()")
	}

	// Add aggregation
	if pq.Aggregation != "" {
		dtAgg := TranslateAggregation(pq.Aggregation)
		if dtAgg != "" {
			if dtAgg != pq.Aggregation {
				logging.Debug("aggregation translated: %s -> %s", pq.Aggregation, dtAgg)
			}
			sb.WriteString(":" + dtAgg)
		}
	}

	// Add rollup
	if pq.Rollup != nil {
		dtRollup := MapRollupFunction(pq.Rollup.Method)
		if pq.Rollup.Period > 0 {
			sb.WriteString(fmt.Sprintf(":fold(%s,%d)", dtRollup, pq.Rollup.Period))
		} else {
			sb.WriteString(fmt.Sprintf(":fold(%s)", dtRollup))
		}
	}

	// Add as_count / as_rate modifier
	if pq.AsModifier == "rate" {
		sb.WriteString(":rate")
	}
	// as_count is default for count metrics in DT, no extra modifier needed

	// Add wrapping function translation
	if pq.Function != "" {
		dtFunc := MapFunction(pq.Function)
		if dtFunc != "" {
			sb.WriteString(":" + dtFunc)
		} else {
			logging.Debug("function %q has no DT equivalent, skipping", pq.Function)
		}
	}

	// Add sort/limit for top/bottom functions
	if pq.Function == "top" || pq.Function == "bottom" {
		order := "descending"
		if pq.Function == "bottom" {
			order = "ascending"
		}
		limit := "10"
		if len(pq.FuncArgs) > 0 {
			limit = pq.FuncArgs[0]
		}
		agg := "avg"
		if len(pq.FuncArgs) > 1 {
			agg = MapRollupFunction(pq.FuncArgs[1])
		}
		sb.WriteString(fmt.Sprintf(":sort(value(%s,%s)):limit(%s)", agg, order, limit))
	}

	return sb.String()
}

// DQLCompute represents a compute aggregation for DQL queries.
type DQLCompute struct {
	Aggregation string
	Facet       string
}

// DQLGroupBy represents a group-by clause for DQL queries.
type DQLGroupBy struct {
	Facet string
	Limit int
}

// ToDQL converts a DataDog log/APM query to Dynatrace Query Language (DQL).
// sourceType should be "log" or "apm".
func ToDQL(ddQuery string, sourceType string) string {
	return ToDQLFull(ddQuery, sourceType, nil, nil)
}

// ToDQLFull converts a DataDog log/APM query to DQL with optional compute and group-by.
func ToDQLFull(ddQuery string, sourceType string, compute *DQLCompute, groupBy []DQLGroupBy) string {
	if ddQuery == "" {
		return ""
	}

	fetchTarget := "logs"
	if sourceType == "apm" {
		fetchTarget = "spans"
	}
	dql := "fetch " + fetchTarget

	// Translate log query tokens
	var filters []string
	parts := tokenizeLogQuery(ddQuery)
	for _, part := range parts {
		// Skip boolean operators — DQL uses "and"/"or" directly
		upper := strings.ToUpper(part)
		if upper == "AND" || upper == "OR" {
			filters = append(filters, strings.ToLower(part))
			continue
		}

		// Negation
		negated := false
		if strings.HasPrefix(part, "-") {
			negated = true
			part = part[1:]
		}

		translated := translateLogToken(part)
		if negated {
			translated = "NOT(" + translated + ")"
		}
		filters = append(filters, translated)
	}

	if len(filters) > 0 {
		dql += "\n| filter " + strings.Join(filters, " ")
	}

	// Append compute (summarize) clause
	if compute != nil {
		agg := mapDQLAggregation(compute.Aggregation)
		facet := compute.Facet
		if strings.HasPrefix(facet, "@") {
			facet = facet[1:]
		}
		if facet != "" {
			dql += fmt.Sprintf("\n| summarize %s(%s)", agg, facet)
		} else {
			dql += fmt.Sprintf("\n| summarize %s()", agg)
		}

		if len(groupBy) > 0 {
			var facets []string
			for _, gb := range groupBy {
				facets = append(facets, gb.Facet)
			}
			dql += fmt.Sprintf(", by {%s}", strings.Join(facets, ", "))
		}
	}

	return dql
}

// tokenizeLogQuery splits a DD log query into tokens, respecting quoted strings.
func tokenizeLogQuery(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(query); i++ {
		c := query[i]
		if inQuotes {
			current.WriteByte(c)
			if c == quoteChar {
				inQuotes = false
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuotes = true
			quoteChar = c
			current.WriteByte(c)
			continue
		}
		if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteByte(c)
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// translateLogToken translates a single DD log search token to DQL.
func translateLogToken(token string) string {
	// Handle specific status values
	statusMap := map[string]string{
		"status:error":    "loglevel == \"ERROR\"",
		"status:warn":     "loglevel == \"WARN\"",
		"status:warning":  "loglevel == \"WARN\"",
		"status:info":     "loglevel == \"INFO\"",
		"status:ok":       "loglevel == \"INFO\"",
		"status:debug":    "loglevel == \"DEBUG\"",
		"status:critical":  "loglevel == \"ERROR\"",
		"status:emergency": "loglevel == \"ERROR\"",
	}
	if dt, ok := statusMap[strings.ToLower(token)]; ok {
		return dt
	}

	// Handle field:value patterns
	if idx := strings.Index(token, ":"); idx > 0 {
		field := token[:idx]
		value := token[idx+1:]
		// Strip quotes from value
		value = strings.Trim(value, "\"'")

		fieldMap := map[string]string{
			"source":            "dt.entity.process_group_instance",
			"service":           "dt.entity.service",
			"host":              "dt.entity.host",
			"status":            "loglevel",
			"@http.method":      "http.method",
			"@http.status_code": "http.status_code",
			"@http.url":         "http.url",
			"@http.url_details.path": "http.url.path",
			"@duration":         "duration",
			"@msg":              "content",
			"env":               "environment",
			"@env":              "environment",
		}

		dtField := field
		if mapped, ok := fieldMap[field]; ok {
			dtField = mapped
		} else if strings.HasPrefix(field, "@") {
			// DD custom attributes start with @ — strip it for DT
			dtField = field[1:]
		}

		// Handle wildcards
		if strings.Contains(value, "*") {
			pattern := strings.ReplaceAll(value, "*", "%")
			return fmt.Sprintf("%s LIKE \"%s\"", dtField, pattern)
		}

		return fmt.Sprintf("%s == \"%s\"", dtField, value)
	}

	// Plain text — content search
	return fmt.Sprintf("content CONTAINS \"%s\"", token)
}

// TranslateMetricName converts a DD metric name to a DT metric key.
func TranslateMetricName(ddMetric string) string {
	metricMap := map[string]string{
		// System metrics
		"system.cpu.user":        "builtin:host.cpu.user",
		"system.cpu.system":      "builtin:host.cpu.system",
		"system.cpu.idle":        "builtin:host.cpu.idle",
		"system.cpu.iowait":      "builtin:host.cpu.iowait",
		"system.cpu.stolen":      "builtin:host.cpu.steal",
		"system.load.1":          "builtin:host.cpu.load",
		"system.load.5":          "builtin:host.cpu.load",
		"system.load.15":         "builtin:host.cpu.load",
		"system.mem.used":        "builtin:host.mem.usage",
		"system.mem.free":        "builtin:host.mem.avail",
		"system.mem.total":       "builtin:host.mem.total",
		"system.mem.pct_usable":  "builtin:host.mem.usage",
		"system.swap.used":       "builtin:host.mem.swap.used",
		"system.swap.free":       "builtin:host.mem.swap.avail",
		"system.disk.used":       "builtin:host.disk.usedPct",
		"system.disk.free":       "builtin:host.disk.avail",
		"system.disk.in_use":     "builtin:host.disk.usedPct",
		"system.disk.read_time_pct":  "builtin:host.disk.readTime",
		"system.disk.write_time_pct": "builtin:host.disk.writeTime",
		"system.io.r_s":          "builtin:host.disk.readOps",
		"system.io.w_s":          "builtin:host.disk.writeOps",
		"system.net.bytes_rcvd":  "builtin:host.net.nic.trafficIn",
		"system.net.bytes_sent":  "builtin:host.net.nic.trafficOut",
		"system.net.packets_in.count":  "builtin:host.net.nic.packetsIn",
		"system.net.packets_out.count": "builtin:host.net.nic.packetsOut",
		// Docker / containers
		"docker.cpu.usage":           "builtin:containers.cpu.usagePercent",
		"docker.cpu.throttled":       "builtin:containers.cpu.throttledMilliseconds",
		"docker.mem.rss":             "builtin:containers.memory.residentSetSize",
		"docker.mem.limit":           "builtin:containers.memory.memoryLimit",
		"docker.mem.in_use":          "builtin:containers.memory.usagePercent",
		"docker.net.bytes_rcvd":      "builtin:containers.net.receivedBytes",
		"docker.net.bytes_sent":      "builtin:containers.net.sentBytes",
		// Kubernetes
		"kubernetes.cpu.usage.total":      "builtin:cloud.kubernetes.workload.cpuUsage",
		"kubernetes.cpu.requests":         "builtin:cloud.kubernetes.workload.cpuRequested",
		"kubernetes.cpu.limits":           "builtin:cloud.kubernetes.workload.cpuLimits",
		"kubernetes.memory.usage":         "builtin:cloud.kubernetes.workload.memoryUsage",
		"kubernetes.memory.requests":      "builtin:cloud.kubernetes.workload.memoryRequested",
		"kubernetes.memory.limits":        "builtin:cloud.kubernetes.workload.memoryLimits",
		"kubernetes.pods.running":         "builtin:cloud.kubernetes.workload.pods",
		"kubernetes.pods.desired":         "builtin:cloud.kubernetes.workload.desiredPods",
		"kubernetes.containers.running":   "builtin:cloud.kubernetes.workload.runningContainers",
		"kubernetes.containers.restarts":  "builtin:cloud.kubernetes.workload.containerRestarts",
		// APM / Traces
		"trace.servlet.request.hits":     "builtin:service.requestCount.total",
		"trace.servlet.request.errors":   "builtin:service.errors.total.count",
		"trace.servlet.request.duration": "builtin:service.response.time",
		"trace.http.request.hits":        "builtin:service.requestCount.total",
		"trace.http.request.errors":      "builtin:service.errors.total.count",
		"trace.http.request.duration":    "builtin:service.response.time",
		// Nginx
		"nginx.net.connections":        "builtin:tech.nginx.connections",
		"nginx.net.request_per_s":      "builtin:tech.nginx.requestsPerSecond",
		// Redis
		"redis.mem.used":               "builtin:tech.redis.memoryUsed",
		"redis.net.clients":            "builtin:tech.redis.connectedClients",
		"redis.keys.evicted":           "builtin:tech.redis.evictedKeys",
		// PostgreSQL
		"postgresql.connections":        "builtin:tech.postgresql.connections",
		// AWS CloudWatch - EC2
		"aws.ec2.cpuutilization":           "ext:cloud.aws.ec2.cpuUtilization",
		"aws.ec2.disk_read_ops":            "ext:cloud.aws.ec2.diskReadOps",
		"aws.ec2.disk_write_ops":           "ext:cloud.aws.ec2.diskWriteOps",
		"aws.ec2.network_in":               "ext:cloud.aws.ec2.networkIn",
		"aws.ec2.network_out":              "ext:cloud.aws.ec2.networkOut",
		"aws.ec2.status_check_failed":      "ext:cloud.aws.ec2.statusCheckFailed",
		// AWS CloudWatch - ELB
		"aws.elb.request_count":            "ext:cloud.aws.elb.requestCount",
		"aws.elb.latency":                  "ext:cloud.aws.elb.latency",
		"aws.elb.httpcode_backend_5xx":     "ext:cloud.aws.elb.httpCode5xx",
		"aws.elb.healthy_host_count":       "ext:cloud.aws.elb.healthyHostCount",
		"aws.elb.unhealthy_host_count":     "ext:cloud.aws.elb.unhealthyHostCount",
		// AWS CloudWatch - RDS
		"aws.rds.cpuutilization":           "ext:cloud.aws.rds.cpuUtilization",
		"aws.rds.database_connections":     "ext:cloud.aws.rds.databaseConnections",
		"aws.rds.free_storage_space":       "ext:cloud.aws.rds.freeStorageSpace",
		"aws.rds.read_iops":                "ext:cloud.aws.rds.readIOPS",
		"aws.rds.write_iops":               "ext:cloud.aws.rds.writeIOPS",
		// More APM / Traces
		"trace.grpc.request.hits":          "builtin:service.requestCount.total",
		"trace.grpc.request.errors":        "builtin:service.errors.total.count",
		"trace.grpc.request.duration":      "builtin:service.response.time",
		"trace.redis.request.hits":         "builtin:service.requestCount.total",
		"trace.redis.request.errors":       "builtin:service.errors.total.count",
		"trace.redis.request.duration":     "builtin:service.response.time",
		"trace.pg.request.hits":            "builtin:service.requestCount.total",
		"trace.pg.request.errors":          "builtin:service.errors.total.count",
		"trace.pg.request.duration":        "builtin:service.response.time",
		// JVM
		"jvm.heap_memory":                  "builtin:tech.jvm.memory.heap.used",
		"jvm.heap_memory_max":              "builtin:tech.jvm.memory.heap.max",
		"jvm.non_heap_memory":              "builtin:tech.jvm.memory.nonheap.used",
		"jvm.gc.major_collection_count":    "builtin:tech.jvm.memory.gc.collectionCount",
		"jvm.gc.minor_collection_count":    "builtin:tech.jvm.memory.gc.collectionCount",
		"jvm.gc.major_collection_time":     "builtin:tech.jvm.memory.gc.collectionTime",
		"jvm.gc.minor_collection_time":     "builtin:tech.jvm.memory.gc.collectionTime",
		"jvm.thread_count":                 "builtin:tech.jvm.threads.count",
		// Process
		"process.cpu.pct":                  "builtin:process.cpu.usage",
		"process.mem.rss":                  "builtin:process.memory.resident",
		"process.open_fds":                 "builtin:process.openFileDescriptors",
		// System extras
		"system.cpu.guest":                 "builtin:host.cpu.other",
		"system.uptime":                    "builtin:host.availability",
		"system.processes.count":           "builtin:host.processes",
	}

	if dt, ok := metricMap[ddMetric]; ok {
		return dt
	}

	// For custom metrics, convert DD naming to DT convention
	if strings.HasPrefix(ddMetric, "custom.") {
		return "ext:" + ddMetric
	}

	// Default: pass through as custom metric with ext: prefix
	return "ext:custom." + strings.ReplaceAll(ddMetric, ":", ".")
}

// TranslateAggregation converts a DD aggregation to DT equivalent.
func TranslateAggregation(ddAgg string) string {
	return MapAggregation(ddAgg)
}

// mapDQLAggregation maps DD aggregation names to DQL equivalents.
func mapDQLAggregation(agg string) string {
	switch strings.ToLower(agg) {
	case "count":
		return "count"
	case "avg":
		return "avg"
	case "sum":
		return "sum"
	case "min":
		return "min"
	case "max":
		return "max"
	case "cardinality":
		return "countDistinct"
	default:
		return agg
	}
}

func filterTermExpr(ft FilterTerm) string {
	dtKey := translateFilterKey(ft.Key)
	if ft.Negated {
		return fmt.Sprintf("ne(%s,\"%s\")", dtKey, ft.Value)
	}
	return fmt.Sprintf("eq(%s,\"%s\")", dtKey, ft.Value)
}

func translateFilterKey(ddKey string) string {
	keyMap := map[string]string{
		"host":          "dt.entity.host",
		"service":       "dt.entity.service",
		"env":           "environment",
		"environment":   "environment",
		"region":        "cloud.region",
		"zone":          "cloud.zone",
		"availability-zone": "cloud.zone",
		"pod_name":      "k8s.pod.name",
		"kube_pod_name": "k8s.pod.name",
		"namespace":     "k8s.namespace.name",
		"kube_namespace": "k8s.namespace.name",
		"cluster":       "k8s.cluster.name",
		"kube_cluster_name": "k8s.cluster.name",
		"container":     "container.name",
		"container_name": "container.name",
		"image":         "container.image.name",
		"device":        "dt.entity.disk",
		"interface":     "dt.entity.network_interface",
	}

	if dt, ok := keyMap[ddKey]; ok {
		return dt
	}
	return ddKey
}
