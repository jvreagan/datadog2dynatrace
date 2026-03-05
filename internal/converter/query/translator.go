package query

import (
	"fmt"
	"strings"
)

// ToMetricSelector converts a parsed DataDog query to a Dynatrace metric selector.
func ToMetricSelector(pq *ParsedQuery) string {
	if pq == nil {
		return ""
	}

	metric := TranslateMetricName(pq.Metric)

	var sb strings.Builder
	sb.WriteString(metric)

	// Add filters
	if len(pq.Filters) > 0 {
		var posFilters, negFilters []string
		for k, v := range pq.Filters {
			if strings.HasPrefix(k, "!") {
				dtKey := translateFilterKey(k[1:])
				negFilters = append(negFilters, fmt.Sprintf("ne(%s,\"%s\")", dtKey, v))
			} else {
				dtKey := translateFilterKey(k)
				posFilters = append(posFilters, fmt.Sprintf("eq(%s,\"%s\")", dtKey, v))
			}
		}
		var allFilters []string
		allFilters = append(allFilters, posFilters...)
		allFilters = append(allFilters, negFilters...)
		if len(allFilters) == 1 {
			sb.WriteString(fmt.Sprintf(":filter(%s)", allFilters[0]))
		} else if len(allFilters) > 1 {
			sb.WriteString(fmt.Sprintf(":filter(and(%s))", strings.Join(allFilters, ",")))
		}
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
			sb.WriteString(":" + dtAgg)
		}
	}

	// Add rollup
	if pq.Rollup != nil {
		dtRollup := MapRollupFunction(pq.Rollup.Method)
		if pq.Rollup.Period > 0 {
			sb.WriteString(fmt.Sprintf(":fold(%s)", dtRollup))
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

// ToDQL converts a DataDog log query to Dynatrace Query Language (DQL).
func ToDQL(ddQuery string) string {
	if ddQuery == "" {
		return ""
	}

	dql := "fetch logs"

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
