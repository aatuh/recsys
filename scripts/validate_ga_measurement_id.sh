#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${PUBLIC_GA_MEASUREMENT_ID:-}" ]]; then
  echo "PUBLIC_GA_MEASUREMENT_ID must be set for the production Pages build." >&2
  exit 1
fi

if [[ ! "${PUBLIC_GA_MEASUREMENT_ID}" =~ ^G-[A-Za-z0-9]+$ ]]; then
  echo "PUBLIC_GA_MEASUREMENT_ID must look like a GA4 Measurement ID, for example G-XXXXXXXXXX." >&2
  exit 1
fi
