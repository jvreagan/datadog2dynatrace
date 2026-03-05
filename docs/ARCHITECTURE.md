# Architecture

## Overview

`datadog2dynatrace` follows a pipeline architecture:

```
DataDog Source → Extraction → Intermediate Representation → Conversion → Output Target
```

## Pipeline Stages

### 1. Extraction (internal/datadog/)
Pulls resources from either:
- **DataDog API** (`internal/datadog/client.go`): Live API calls with pagination
- **File imports** (`internal/importer/`): JSON exports or Terraform files

All resources are normalized into Go structs defined in `internal/datadog/types.go`.

### 2. Selection (internal/ui/)
Interactive TUI (bubbletea) lets users choose which resources to migrate. Supports group selection, toggle all, and individual checkbox selection. Skipped with `--all` flag.

### 3. Conversion (internal/converter/)
Each resource type has a dedicated converter that maps DataDog concepts to Dynatrace equivalents:

- `converter.go` — Orchestrator that calls individual converters
- `dashboards.go` — Widget-by-widget dashboard translation
- `monitors.go` — Monitor → Metric Event with query translation
- `slos.go` — SLO type and threshold mapping
- `synthetics.go` — HTTP/Browser test conversion
- `logs.go` — Pipeline processor → DQL rule translation
- `metrics.go` — Metric metadata and unit mapping
- `downtimes.go` — Downtime → Maintenance window schedule mapping
- `notifications.go` — Channel type mapping (Slack, PD, email, webhook)
- `notebooks.go` — Cell-by-cell notebook translation

### 4. Query Translation (internal/converter/query/)
The most complex subsystem, handling DD query language → DT metric selector/DQL:

- `parser.go` — Parses DD queries into structured AST
- `translator.go` — Converts parsed queries to DT metric selectors or DQL
- `functions.go` — Maps DD aggregation/rollup functions to DT equivalents

### 5. Output (internal/dynatrace/, internal/terraform/)
Two output modes:
- **Dynatrace API** (`internal/dynatrace/client.go`): Direct push via REST API
- **Terraform** (`internal/terraform/`): Generates HCL files per resource type

### 6. Reporting (internal/report/)
Generates a Markdown migration report with extraction summary, conversion summary, and any errors.

## Configuration

Config loading (`internal/config/`) follows precedence: CLI flags > config file (`~/.datadog2dynatrace.yaml`) > environment variables. Implemented with Viper.

## API Layer

`api/api.go` provides a programmatic interface to the pipeline for future web UI integration.

## Error Handling

By default, errors are collected and reported at the end. With `--fail-fast`, the first error stops execution. Errors at each stage (extraction, conversion, push) are tracked separately and included in the migration report.
