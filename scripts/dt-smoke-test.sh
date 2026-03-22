#!/usr/bin/env bash
set -euo pipefail

# ── Configuration ──────────────────────────────────────────────
# Set these via environment variables or pass as arguments:
#   DT_ENV_URL   — e.g. https://abc12345.live.dynatrace.com
#   DT_API_TOKEN — token with metrics.ingest, logs.ingest, events.ingest scopes

DT_ENV_URL="${DT_ENV_URL:?Set DT_ENV_URL (e.g. https://abc12345.live.dynatrace.com)}"
DT_API_TOKEN="${DT_API_TOKEN:?Set DT_API_TOKEN with metrics.ingest, logs.ingest, events.ingest scopes}"

# Strip trailing slash if present
DT_ENV_URL="${DT_ENV_URL%/}"

PASS=0
FAIL=0
DT_LAST_BODY=""

call_api() {
  local method="$1" url="$2" content_type="$3" data="$4"
  local http_code body
  http_code=$(curl -s -o /tmp/dt-smoke-body.txt -w "%{http_code}" \
    -X "$method" "$url" \
    -H "Authorization: Api-Token ${DT_API_TOKEN}" \
    -H "Content-Type: ${content_type}" \
    -d "$data")
  body=$(cat /tmp/dt-smoke-body.txt)
  echo "$http_code"
  DT_LAST_BODY="$body"
}

check() {
  local label="$1" code="$2"
  if [[ "$code" -ge 200 && "$code" -lt 300 ]]; then
    echo "  ✅ $label — HTTP $code"
    PASS=$((PASS + 1))
  else
    echo "  ❌ $label — HTTP $code"
    echo "     $DT_LAST_BODY"
    FAIL=$((FAIL + 1))
  fi
}

echo "Dynatrace smoke test — $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "Environment: $DT_ENV_URL"
echo ""

# ── 1. Metric ──────────────────────────────────────────────────
echo "1) Pushing test metric..."
TIMESTAMP=$(($(date +%s) * 1000))
CODE=$(call_api POST "${DT_ENV_URL}/api/v2/metrics/ingest" \
  "text/plain" \
  "d2d.smoke.test,source=smoke-test 1 ${TIMESTAMP}")
check "Metric (d2d.smoke.test)" "$CODE"

# ── 2. Log ─────────────────────────────────────────────────────
echo "2) Pushing test log..."
CODE=$(call_api POST "${DT_ENV_URL}/api/v2/logs/ingest" \
  "application/json; charset=utf-8" \
  "[{\"content\": \"d2d smoke test at $(date -u +%Y-%m-%dT%H:%M:%SZ)\", \"severity\": \"info\", \"log.source\": \"d2d-smoke-test\"}]")
check "Log (d2d-smoke-test)" "$CODE"

# ── 3. Event ───────────────────────────────────────────────────
echo "3) Pushing test event..."
CODE=$(call_api POST "${DT_ENV_URL}/api/v2/events/ingest" \
  "application/json" \
  "{\"eventType\": \"CUSTOM_INFO\", \"title\": \"d2d smoke test\", \"properties\": {\"source\": \"d2d-smoke-test\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}}")
check "Event (CUSTOM_INFO)" "$CODE"

# ── Summary ────────────────────────────────────────────────────
echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"
if [[ "$FAIL" -gt 0 ]]; then
  echo ""
  echo "Troubleshooting:"
  echo "  • Verify your token has these scopes: metrics.ingest, logs.ingest, events.ingest"
  echo "  • Verify DT_ENV_URL is correct (no trailing /api/...)"
  echo "  • Check if your environment requires a different auth header (e.g. SaaS vs Managed)"
  exit 1
fi
