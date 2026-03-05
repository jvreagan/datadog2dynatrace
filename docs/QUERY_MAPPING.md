# DataDog → Dynatrace Query Mapping Reference

## Metric Queries

### Format Translation

| DataDog | Dynatrace |
|---------|-----------|
| `avg:system.cpu.user{host:web01}` | `builtin:host.cpu.user:filter(eq("dt.entity.host","web01")):avg` |
| `sum:custom.metric{env:prod}by{host}` | `ext:custom.metric:filter(eq("environment","prod")):sum:splitBy("dt.entity.host")` |

### Aggregation Functions

| DataDog | Dynatrace |
|---------|-----------|
| `avg` | `avg` |
| `sum` | `sum` |
| `min` | `min` |
| `max` | `max` |
| `count` | `count` |
| `last` | `value` |
| `p50` | `percentile(50)` |
| `p95` | `percentile(95)` |
| `p99` | `percentile(99)` |

### Common Metric Name Mappings

| DataDog | Dynatrace |
|---------|-----------|
| `system.cpu.user` | `builtin:host.cpu.user` |
| `system.cpu.system` | `builtin:host.cpu.system` |
| `system.mem.used` | `builtin:host.mem.usage` |
| `system.disk.used` | `builtin:host.disk.usedPct` |
| `system.net.bytes_rcvd` | `builtin:host.net.nic.trafficIn` |
| `system.net.bytes_sent` | `builtin:host.net.nic.trafficOut` |
| `docker.cpu.usage` | `builtin:containers.cpu.usagePercent` |
| `kubernetes.cpu.usage.total` | `builtin:cloud.kubernetes.workload.cpuUsage` |
| `trace.servlet.request.hits` | `builtin:service.requestCount.total` |
| `trace.servlet.request.errors` | `builtin:service.errors.total.count` |

### Filter Key Mappings

| DataDog | Dynatrace |
|---------|-----------|
| `host` | `dt.entity.host` |
| `service` | `dt.entity.service` |
| `env` | `environment` |
| `region` | `cloud.region` |
| `pod_name` | `k8s.pod.name` |
| `namespace` | `k8s.namespace.name` |
| `cluster` | `k8s.cluster.name` |

## Log Queries

DataDog uses Lucene-like syntax; Dynatrace uses DQL.

| DataDog | Dynatrace DQL |
|---------|---------------|
| `source:nginx` | `fetch logs \| filter dt.entity.process_group_instance == "nginx"` |
| `status:error` | `fetch logs \| filter loglevel == "ERROR"` |
| `service:web @http.method:GET` | `fetch logs \| filter dt.entity.service == "web" and http.method == "GET"` |

## Functions

| DataDog | Dynatrace | Notes |
|---------|-----------|-------|
| `per_second()` | `:rate` | |
| `rollup(avg, 60)` | `:fold(avg)` | |
| `top()` | `:sort` | |
| `diff()` | `:delta` | |
| `derivative()` | `:rate` | |
| `abs()` | `:abs` | |
| `forecast()` | — | No direct equivalent |
| `anomalies()` | — | Built-in anomaly detection |
