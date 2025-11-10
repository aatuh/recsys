#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:8000}
ORG_ID=${ORG_ID:-00000000-0000-0000-0000-000000000001}
NAMESPACE=${NAMESPACE:-marketplace_demo}
ENV_PROFILE=${ENV_PROFILE:-dev}
CUSTOMER_LABEL=${CUSTOMER_LABEL:-marketplace-demo}
FIXTURE=${FIXTURE:-analysis/fixtures/templates/marketplace.json}
GUARDRAILS_FILE=${GUARDRAILS_FILE:-guardrails.yml}

python analysis/scripts/run_simulation.py \
  --customer "$CUSTOMER_LABEL" \
  --base-url "$BASE_URL" \
  --org-id "$ORG_ID" \
  --namespace "$NAMESPACE" \
  --env-profile "$ENV_PROFILE" \
  --fixture-path "$FIXTURE" \
  --guardrails-file "$GUARDRAILS_FILE"
