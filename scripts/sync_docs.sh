#!/usr/bin/env bash
# Syncs documentation into recsys/docs for MkDocs.
# - Copies project/legal root markdown into docs/licensing and docs/project
# - Mirrors recsys-eval docs into docs/recsys-eval (README + docs/)
# - Mirrors recsys-pipelines docs into docs/recsys-pipelines (README + docs/)
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RECSYS_EVAL_DIR="${ROOT_DIR}/recsys-eval"
RECSYS_PIPELINES_DIR="${ROOT_DIR}/recsys-pipelines"

cp "${ROOT_DIR}/COMMERCIAL.md" "${ROOT_DIR}/docs/licensing/commercial.md"
cp "${ROOT_DIR}/LICENSING.md" "${ROOT_DIR}/docs/licensing/index.md"
cp "${ROOT_DIR}/PRICING.md" "${ROOT_DIR}/docs/licensing/pricing.md"
cp "${ROOT_DIR}/SUPPORT.md" "${ROOT_DIR}/docs/project/support.md"
cp "${ROOT_DIR}/GOVERNANCE.md" "${ROOT_DIR}/docs/project/governance.md"
cp "${ROOT_DIR}/SECURITY.md" "${ROOT_DIR}/docs/project/security.md"
cp "${ROOT_DIR}/CODE_OF_CONDUCT.md" "${ROOT_DIR}/docs/project/code_of_conduct.md"
cp "${ROOT_DIR}/CONTRIBUTING.md" "${ROOT_DIR}/docs/project/contributing.md"
cp "${ROOT_DIR}/LICENSES-README.md" "${ROOT_DIR}/docs/project/licenses_readme.md"

EVAL_ROOT_DST="${ROOT_DIR}/docs/recsys-eval"
EVAL_DOCS_SRC="${RECSYS_EVAL_DIR}/docs"
EVAL_DOCS_DST="${EVAL_ROOT_DST}/docs"
if [[ -d "${EVAL_DOCS_SRC}" ]]; then
  if [[ -d "${EVAL_ROOT_DST}" ]]; then
    rm -rf "${EVAL_ROOT_DST}"
  fi
  mkdir -p "${EVAL_DOCS_DST}"
  cp "${RECSYS_EVAL_DIR}/README.md" "${EVAL_ROOT_DST}/overview.md"
  cp -R "${EVAL_DOCS_SRC}/." "${EVAL_DOCS_DST}"
fi

PIPELINES_ROOT_DST="${ROOT_DIR}/docs/recsys-pipelines"
PIPELINES_DOCS_SRC="${RECSYS_PIPELINES_DIR}/docs"
PIPELINES_DOCS_DST="${PIPELINES_ROOT_DST}/docs"
if [[ -d "${PIPELINES_DOCS_SRC}" ]]; then
  if [[ -d "${PIPELINES_ROOT_DST}" ]]; then
    rm -rf "${PIPELINES_ROOT_DST}"
  fi
  mkdir -p "${PIPELINES_DOCS_DST}"
  cp "${RECSYS_PIPELINES_DIR}/README.md" "${PIPELINES_ROOT_DST}/overview.md"
  cp -R "${PIPELINES_DOCS_SRC}/." "${PIPELINES_DOCS_DST}"
fi

printf 'Docs synced.\n'
