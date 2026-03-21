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
| `p75` | `percentile(75)` |
| `p90` | `percentile(90)` |
| `p95` | `percentile(95)` |
| `p99` | `percentile(99)` |

### Common Metric Name Mappings

#### System

| DataDog | Dynatrace |
|---------|-----------|
| `system.cpu.user` | `builtin:host.cpu.user` |
| `system.cpu.system` | `builtin:host.cpu.system` |
| `system.cpu.idle` | `builtin:host.cpu.idle` |
| `system.cpu.iowait` | `builtin:host.cpu.iowait` |
| `system.cpu.stolen` | `builtin:host.cpu.steal` |
| `system.cpu.guest` | `builtin:host.cpu.other` |
| `system.load.1` | `builtin:host.cpu.load` |
| `system.load.5` | `builtin:host.cpu.load` |
| `system.load.15` | `builtin:host.cpu.load` |
| `system.mem.used` | `builtin:host.mem.usage` |
| `system.mem.free` | `builtin:host.mem.avail` |
| `system.mem.total` | `builtin:host.mem.total` |
| `system.mem.pct_usable` | `builtin:host.mem.usage` |
| `system.swap.used` | `builtin:host.mem.swap.used` |
| `system.swap.free` | `builtin:host.mem.swap.avail` |
| `system.disk.used` | `builtin:host.disk.usedPct` |
| `system.disk.free` | `builtin:host.disk.avail` |
| `system.disk.in_use` | `builtin:host.disk.usedPct` |
| `system.disk.read_time_pct` | `builtin:host.disk.readTime` |
| `system.disk.write_time_pct` | `builtin:host.disk.writeTime` |
| `system.io.r_s` | `builtin:host.disk.readOps` |
| `system.io.w_s` | `builtin:host.disk.writeOps` |
| `system.net.bytes_rcvd` | `builtin:host.net.nic.trafficIn` |
| `system.net.bytes_sent` | `builtin:host.net.nic.trafficOut` |
| `system.net.packets_in.count` | `builtin:host.net.nic.packetsIn` |
| `system.net.packets_out.count` | `builtin:host.net.nic.packetsOut` |
| `system.uptime` | `builtin:host.availability` |
| `system.processes.count` | `builtin:host.processes` |

#### Docker / Containers

| DataDog | Dynatrace |
|---------|-----------|
| `docker.cpu.usage` | `builtin:containers.cpu.usagePercent` |
| `docker.cpu.throttled` | `builtin:containers.cpu.throttledMilliseconds` |
| `docker.mem.rss` | `builtin:containers.memory.residentSetSize` |
| `docker.mem.limit` | `builtin:containers.memory.memoryLimit` |
| `docker.mem.in_use` | `builtin:containers.memory.usagePercent` |
| `docker.net.bytes_rcvd` | `builtin:containers.net.receivedBytes` |
| `docker.net.bytes_sent` | `builtin:containers.net.sentBytes` |

#### Kubernetes

| DataDog | Dynatrace |
|---------|-----------|
| `kubernetes.cpu.usage.total` | `builtin:cloud.kubernetes.workload.cpuUsage` |
| `kubernetes.cpu.requests` | `builtin:cloud.kubernetes.workload.cpuRequested` |
| `kubernetes.cpu.limits` | `builtin:cloud.kubernetes.workload.cpuLimits` |
| `kubernetes.memory.usage` | `builtin:cloud.kubernetes.workload.memoryUsage` |
| `kubernetes.memory.requests` | `builtin:cloud.kubernetes.workload.memoryRequested` |
| `kubernetes.memory.limits` | `builtin:cloud.kubernetes.workload.memoryLimits` |
| `kubernetes.pods.running` | `builtin:cloud.kubernetes.workload.pods` |
| `kubernetes.pods.desired` | `builtin:cloud.kubernetes.workload.desiredPods` |
| `kubernetes.containers.running` | `builtin:cloud.kubernetes.workload.runningContainers` |
| `kubernetes.containers.restarts` | `builtin:cloud.kubernetes.workload.containerRestarts` |

#### APM / Traces

| DataDog | Dynatrace |
|---------|-----------|
| `trace.servlet.request.hits` | `builtin:service.requestCount.total` |
| `trace.servlet.request.errors` | `builtin:service.errors.total.count` |
| `trace.servlet.request.duration` | `builtin:service.response.time` |
| `trace.http.request.hits` | `builtin:service.requestCount.total` |
| `trace.http.request.errors` | `builtin:service.errors.total.count` |
| `trace.http.request.duration` | `builtin:service.response.time` |
| `trace.grpc.request.hits` | `builtin:service.requestCount.total` |
| `trace.grpc.request.errors` | `builtin:service.errors.total.count` |
| `trace.grpc.request.duration` | `builtin:service.response.time` |
| `trace.redis.request.hits` | `builtin:service.requestCount.total` |
| `trace.redis.request.errors` | `builtin:service.errors.total.count` |
| `trace.redis.request.duration` | `builtin:service.response.time` |
| `trace.pg.request.hits` | `builtin:service.requestCount.total` |
| `trace.pg.request.errors` | `builtin:service.errors.total.count` |
| `trace.pg.request.duration` | `builtin:service.response.time` |

#### AWS CloudWatch

| DataDog | Dynatrace |
|---------|-----------|
| `aws.ec2.cpuutilization` | `ext:cloud.aws.ec2.cpuUtilization` |
| `aws.ec2.disk_read_ops` | `ext:cloud.aws.ec2.diskReadOps` |
| `aws.ec2.disk_write_ops` | `ext:cloud.aws.ec2.diskWriteOps` |
| `aws.ec2.network_in` | `ext:cloud.aws.ec2.networkIn` |
| `aws.ec2.network_out` | `ext:cloud.aws.ec2.networkOut` |
| `aws.ec2.status_check_failed` | `ext:cloud.aws.ec2.statusCheckFailed` |
| `aws.elb.request_count` | `ext:cloud.aws.elb.requestCount` |
| `aws.elb.latency` | `ext:cloud.aws.elb.latency` |
| `aws.elb.httpcode_backend_5xx` | `ext:cloud.aws.elb.httpCode5xx` |
| `aws.elb.healthy_host_count` | `ext:cloud.aws.elb.healthyHostCount` |
| `aws.elb.unhealthy_host_count` | `ext:cloud.aws.elb.unhealthyHostCount` |
| `aws.rds.cpuutilization` | `ext:cloud.aws.rds.cpuUtilization` |
| `aws.rds.database_connections` | `ext:cloud.aws.rds.databaseConnections` |
| `aws.rds.free_storage_space` | `ext:cloud.aws.rds.freeStorageSpace` |
| `aws.rds.read_iops` | `ext:cloud.aws.rds.readIOPS` |
| `aws.rds.write_iops` | `ext:cloud.aws.rds.writeIOPS` |

#### JVM

| DataDog | Dynatrace |
|---------|-----------|
| `jvm.heap_memory` | `builtin:tech.jvm.memory.heap.used` |
| `jvm.heap_memory_max` | `builtin:tech.jvm.memory.heap.max` |
| `jvm.non_heap_memory` | `builtin:tech.jvm.memory.nonheap.used` |
| `jvm.gc.major_collection_count` | `builtin:tech.jvm.memory.gc.collectionCount` |
| `jvm.gc.minor_collection_count` | `builtin:tech.jvm.memory.gc.collectionCount` |
| `jvm.gc.major_collection_time` | `builtin:tech.jvm.memory.gc.collectionTime` |
| `jvm.gc.minor_collection_time` | `builtin:tech.jvm.memory.gc.collectionTime` |
| `jvm.thread_count` | `builtin:tech.jvm.threads.count` |

#### Process

| DataDog | Dynatrace |
|---------|-----------|
| `process.cpu.pct` | `builtin:process.cpu.usage` |
| `process.mem.rss` | `builtin:process.memory.resident` |
| `process.open_fds` | `builtin:process.openFileDescriptors` |

#### Technology-Specific

| DataDog | Dynatrace |
|---------|-----------|
| `nginx.net.connections` | `builtin:tech.nginx.connections` |
| `nginx.net.request_per_s` | `builtin:tech.nginx.requestsPerSecond` |
| `redis.mem.used` | `builtin:tech.redis.memoryUsed` |
| `redis.net.clients` | `builtin:tech.redis.connectedClients` |
| `redis.keys.evicted` | `builtin:tech.redis.evictedKeys` |
| `postgresql.connections` | `builtin:tech.postgresql.connections` |

Unmapped metrics are automatically prefixed with `ext:custom.` for custom metric ingestion.

### Filter Key Mappings

| DataDog | Dynatrace |
|---------|-----------|
| `host` | `dt.entity.host` |
| `service` | `dt.entity.service` |
| `env` / `environment` | `environment` |
| `region` | `cloud.region` |
| `zone` / `availability-zone` | `cloud.zone` |
| `pod_name` / `kube_pod_name` | `k8s.pod.name` |
| `namespace` / `kube_namespace` | `k8s.namespace.name` |
| `cluster` / `kube_cluster_name` | `k8s.cluster.name` |
| `container` / `container_name` | `container.name` |
| `image` | `container.image.name` |
| `device` | `dt.entity.disk` |
| `interface` | `dt.entity.network_interface` |

## Log Queries

DataDog uses Lucene-like syntax; Dynatrace uses DQL.

| DataDog | Dynatrace DQL |
|---------|---------------|
| `source:nginx` | `fetch logs \| filter dt.entity.process_group_instance == "nginx"` |
| `status:error` | `fetch logs \| filter loglevel == "ERROR"` |
| `service:web @http.method:GET` | `fetch logs \| filter dt.entity.service == "web" and http.method == "GET"` |

### DQL Aggregation Mapping

| DataDog | Dynatrace DQL |
|---------|---------------|
| `count` | `count` |
| `avg` | `avg` |
| `sum` | `sum` |
| `min` | `min` |
| `max` | `max` |
| `cardinality` | `countDistinct` |

## Functions

| DataDog | Dynatrace | Notes |
|---------|-----------|-------|
| `abs()` | `:abs` | |
| `log2()` | `:log2` | |
| `log10()` | `:log10` | |
| `ceil()` | `:ceil` | |
| `floor()` | `:floor` | |
| `round()` | `:round` | |
| `per_second()` | `:rate` | |
| `per_minute()` | `:rate` | Normalized to per-second rate |
| `per_hour()` | `:rate` | Normalized to per-second rate |
| `rollup(avg, 60)` | `:fold(avg)` | |
| `diff()` | `:delta` | |
| `derivative()` | `:rate` | |
| `dt()` | `:rate` | Alias for derivative |
| `cumsum()` | `:rollup` | |
| `top()` | `:sort` | With descending order + limit |
| `bottom()` | `:sort` | With ascending order + limit |
| `ewma_3/5/10/20()` | `:smooth` | Smoothing functions |
| `median_3/5()` | `:smooth` | Smoothing functions |
| `timeshift()` | `:timeshift` | |
| `count_nonzero()` | `:count` | |
| `count_not_null()` | `:count` | |
| `clamp_min()` | `:partition` | Approximate mapping |
| `clamp_max()` | `:partition` | Approximate mapping |
| `forecast()` | — | No direct equivalent |
| `anomalies()` | — | Built-in anomaly detection |
