# datadog2dynatrace

A CLI tool that converts DataDog monitoring configurations to Dynatrace Cloud equivalents. Built for teams migrating from DataDog to Dynatrace.

## Features

- **Live API migration**: Read from DataDog API and push to Dynatrace API
- **Terraform output**: Generate Dynatrace Terraform configurations
- **File import**: Import from DataDog JSON exports or Terraform files
- **Interactive selection**: Choose which resources to migrate
- **Dry run**: Preview changes before applying
- **Migration report**: Markdown report of what was converted and any issues

### Supported Resources

| DataDog | Dynatrace |
|---------|-----------|
| Dashboards | Dashboards |
| Monitors | Metric Events |
| SLOs | SLOs |
| Synthetic Tests | Synthetic Monitors |
| Log Pipelines | Log Processing Rules |
| Metric Metadata | Metric Metadata |
| Downtimes | Maintenance Windows |
| Notification Channels | Notification Integrations |
| Notebooks | Notebooks |

## Installation

```bash
go install github.com/datadog2dynatrace/datadog2dynatrace/cmd/datadog2dynatrace@latest
```

Or build from source:

```bash
git clone https://github.com/datadog2dynatrace/datadog2dynatrace.git
cd datadog2dynatrace
make build
```

## Configuration

Create `~/.datadog2dynatrace.yaml`:

```yaml
datadog:
  api_key: "your-dd-api-key"
  app_key: "your-dd-app-key"
  site: "datadoghq.com"

dynatrace:
  env_url: "https://your-env.live.dynatrace.com"
  api_token: "your-dt-api-token"
```

Credential precedence: CLI flags > config file > environment variables.

Environment variables:
- `DD_API_KEY`, `DD_APP_KEY`, `DD_SITE`
- `DT_ENV_URL`, `DT_API_TOKEN`

## Usage

### Validate Credentials

```bash
datadog2dynatrace validate
```

### Convert to Terraform (default)

```bash
datadog2dynatrace convert --source api --target terraform
```

### Convert and Push to Dynatrace API

```bash
datadog2dynatrace convert --source api --target api
```

### Dry Run

```bash
datadog2dynatrace convert --source api --target api --dry-run
```

### Import from Files

```bash
datadog2dynatrace convert --source file --input-dir ./dd-exports --target terraform
```

### Convert All (Skip Interactive Selection)

```bash
datadog2dynatrace convert --all
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--source` | Input source: `api` or `file` | `api` |
| `--input-dir` | Directory with DD export files | |
| `--target` | Output target: `api` or `terraform` | `terraform` |
| `--output-dir` | Terraform output directory | `./dynatrace-terraform/` |
| `--dry-run` | Preview without applying changes | `false` |
| `--fail-fast` | Stop on first error | `false` |
| `--all` | Convert all resources (skip selection) | `false` |
| `--report-file` | Migration report path | `./migration-report.md` |
| `--dd-api-key` | DataDog API key | |
| `--dd-app-key` | DataDog Application key | |
| `--dd-site` | DataDog site | `datadoghq.com` |
| `--dt-env-url` | Dynatrace environment URL | |
| `--dt-api-token` | Dynatrace API token | |

## License

Apache 2.0 — see [LICENSE](LICENSE).
