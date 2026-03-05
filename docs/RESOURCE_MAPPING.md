# DataDog → Dynatrace Resource Mapping Reference

## Resource Type Mapping

| DataDog | Dynatrace | Notes |
|---------|-----------|-------|
| Dashboard | Dashboard | Widget-by-widget conversion |
| Monitor | Metric Event | Query and threshold translation |
| SLO | SLO | Metric and monitor-based types |
| Synthetic Test (API) | HTTP Monitor | Assertions → Validation rules |
| Synthetic Test (Browser) | Browser Monitor | Steps → Clickpath events |
| Log Pipeline | Log Processing Rule | Processors → DQL rules |
| Metric Metadata | Metric Descriptor | Unit and description mapping |
| Downtime | Maintenance Window | Schedule and scope mapping |
| Notification Channel | Notification Integration | Type-specific config mapping |
| Notebook | Notebook | Cell-by-cell conversion |

## Dashboard Widget Mapping

| DataDog Widget | Dynatrace Tile | Notes |
|----------------|----------------|-------|
| Timeseries | Data Explorer | Metric selector queries |
| Query Value | Data Explorer | Single value display |
| Top List | Data Explorer | With sort/limit |
| Table | Data Explorer | Tabular metric view |
| Note / Free Text | Markdown | Direct content transfer |
| Group | Header | Flattened in DT |
| Host Map | Hosts | Approximate mapping |
| Heatmap | Data Explorer | Approximated as timeseries |
| SLO | SLO Tile | Requires SLO entity link |

## Monitor → Metric Event Mapping

| DataDog Field | Dynatrace Field |
|---------------|-----------------|
| `name` | `summary` |
| `message` | `description` (sanitized) |
| `query` | `metricSelector` (translated) |
| `options.thresholds.critical` | `threshold` |
| `type` | `eventType` |
| Monitor comparison (`>`, `<`) | `alertCondition` (`ABOVE`, `BELOW`) |

## Notification Type Mapping

| DataDog | Dynatrace |
|---------|-----------|
| Slack | SLACK |
| PagerDuty | PAGER_DUTY |
| Email | EMAIL |
| Webhook | WEBHOOK |
| OpsGenie | OPS_GENIE |
| VictorOps | VICTOR_OPS |

## SLO Timeframe Mapping

| DataDog | Dynatrace |
|---------|-----------|
| `7d` | `-1w` |
| `30d` | `-1M` |
| `90d` | `-3M` |

## Unit Mapping (Metrics)

| DataDog | Dynatrace |
|---------|-----------|
| `byte` | `Byte` |
| `kilobyte` | `KiloByte` |
| `percent` | `Percent` |
| `millisecond` | `MilliSecond` |
| `second` | `Second` |
| `count` | `Count` |
