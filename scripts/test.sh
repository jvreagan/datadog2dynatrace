#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "=== Building ==="
make build

echo ""
echo "=== Running test suite ==="
go test ./... -v -count=1
