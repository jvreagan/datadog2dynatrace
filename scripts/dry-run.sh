#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

OUTPUT_DIR="${1:-/tmp/d2d-dry-run}"

echo "=== Building ==="
make build

echo ""
echo "=== Dry run (Terraform) ==="
./bin/datadog2dynatrace convert \
  --source file \
  --input-dir test/testdata \
  --target terraform \
  --output-dir "$OUTPUT_DIR" \
  --all \
  --dry-run

echo ""
echo "=== Dry run (JSON) ==="
./bin/datadog2dynatrace convert \
  --source file \
  --input-dir test/testdata \
  --target json \
  --output-dir "$OUTPUT_DIR" \
  --all \
  --dry-run
