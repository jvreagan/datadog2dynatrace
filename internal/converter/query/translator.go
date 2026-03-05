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

	// Translate metric name (DD uses dots, DT also uses dots but with different prefixes)
	metric := TranslateMetricName(pq.Metric)

	var parts []string
	parts = append(parts, metric)

	// Add filter
	if len(pq.Filters) > 0 {
		var filters []string
		for k, v := range pq.Filters {
			dtKey := translateFilterKey(k)
			filters = append(filters, fmt.Sprintf("%s(\"%s\")", dtKey, v))
		}
		parts = append(parts, fmt.Sprintf(":filter(and(%s))", strings.Join(filters, ",")))
	}

	// Add aggregation
	if pq.Aggregation != "" {
		dtAgg := TranslateAggregation(pq.Aggregation)
		if dtAgg != "" {
			parts = append(parts, ":"+dtAgg)
		}
	}

	// Add split by (group by)
	if len(pq.GroupBy) > 0 {
		var dims []string
		for _, g := range pq.GroupBy {
			dims = append(dims, fmt.Sprintf("\"%s\"", translateFilterKey(g)))
		}
		parts = append(parts, fmt.Sprintf(":splitBy(%s)", strings.Join(dims, ",")))
	}

	return strings.Join(parts, "")
}

// ToDQL converts a DataDog log query to Dynatrace Query Language (DQL).
func ToDQL(ddQuery string) string {
	if ddQuery == "" {
		return ""
	}

	// Basic translation from DD log search syntax to DQL
	// DD: source:nginx status:error @http.method:GET
	// DT: fetch logs | filter loglevel == "ERROR" and dt.source.name == "nginx"

	dql := "fetch logs"

	// Translate common field mappings
	translated := ddQuery
	replacements := map[string]string{
		"source:":            "dt.entity.process_group_instance == \"",
		"service:":           "dt.entity.service == \"",
		"status:error":       "loglevel == \"ERROR\"",
		"status:warn":        "loglevel == \"WARN\"",
		"status:info":        "loglevel == \"INFO\"",
		"@http.method:":      "http.method == \"",
		"@http.status_code:": "http.status_code == \"",
		"@http.url:":         "http.url == \"",
	}

	var filters []string
	parts := strings.Fields(translated)
	for _, part := range parts {
		matched := false
		for dd, dt := range replacements {
			if strings.HasPrefix(part, dd) {
				value := strings.TrimPrefix(part, dd)
				if strings.HasSuffix(dt, "\"") {
					filters = append(filters, dt+value+"\"")
				} else {
					filters = append(filters, dt)
				}
				matched = true
				break
			}
		}
		if !matched && part != "" {
			// Keep as content filter
			filters = append(filters, fmt.Sprintf("content == \"%s\"", part))
		}
	}

	if len(filters) > 0 {
		dql += "\n| filter " + strings.Join(filters, " and ")
	}

	return dql
}

// TranslateMetricName converts a DD metric name to a DT metric key.
func TranslateMetricName(ddMetric string) string {
	// Common DD → DT metric mappings
	metricMap := map[string]string{
		// System metrics
		"system.cpu.user":        "builtin:host.cpu.user",
		"system.cpu.system":      "builtin:host.cpu.system",
		"system.cpu.idle":        "builtin:host.cpu.idle",
		"system.cpu.iowait":      "builtin:host.cpu.iowait",
		"system.load.1":          "builtin:host.cpu.load",
		"system.load.5":          "builtin:host.cpu.load",
		"system.load.15":         "builtin:host.cpu.load",
		"system.mem.used":        "builtin:host.mem.usage",
		"system.mem.free":        "builtin:host.mem.avail",
		"system.mem.total":       "builtin:host.mem.total",
		"system.disk.used":       "builtin:host.disk.usedPct",
		"system.disk.free":       "builtin:host.disk.avail",
		"system.net.bytes_rcvd":  "builtin:host.net.nic.trafficIn",
		"system.net.bytes_sent":  "builtin:host.net.nic.trafficOut",
		// Docker
		"docker.cpu.usage":       "builtin:containers.cpu.usagePercent",
		"docker.mem.rss":         "builtin:containers.memory.residentSetSize",
		// Kubernetes
		"kubernetes.cpu.usage.total":     "builtin:cloud.kubernetes.workload.cpuUsage",
		"kubernetes.memory.usage":        "builtin:cloud.kubernetes.workload.memoryUsage",
		"kubernetes.pods.running":        "builtin:cloud.kubernetes.workload.pods",
		// APM
		"trace.servlet.request.hits":     "builtin:service.requestCount.total",
		"trace.servlet.request.errors":   "builtin:service.errors.total.count",
		"trace.servlet.request.duration": "builtin:service.response.time",
	}

	if dt, ok := metricMap[ddMetric]; ok {
		return dt
	}

	// For custom metrics, convert DD naming to DT convention
	// DD: custom.metric.name → ext:custom.metric.name
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
		"host":        "dt.entity.host",
		"service":     "dt.entity.service",
		"env":         "environment",
		"environment": "environment",
		"region":      "cloud.region",
		"zone":        "cloud.zone",
		"pod_name":    "k8s.pod.name",
		"namespace":   "k8s.namespace.name",
		"cluster":     "k8s.cluster.name",
		"container":   "container.name",
	}

	if dt, ok := keyMap[ddKey]; ok {
		return dt
	}
	return ddKey
}
