#!/usr/bin/env bash
set -euo pipefail

# Generates docs/*/index.md pages from root markdown files, so you get stable URLs
# (/pricing/, /commercial/, /licensing/, ...) without duplicating the canonical content in Git.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOCS_DIR="${ROOT_DIR}/docs"

# Derive repo URL for rewriting links to templates that are not published on the site.
# Supports:
# - git@github.com:org/repo.git
# - https://github.com/org/repo.git
get_repo_url() {
  local u
  if ! u="$(git -C "${ROOT_DIR}" remote get-url origin 2>/dev/null)"; then
    echo "https://github.com/<ORG>/<REPO>"
    return 0
  fi

  if [[ "$u" =~ ^git@github.com:(.+)$ ]]; then
    u="https://github.com/${BASH_REMATCH[1]}"
  fi
  u="${u%.git}"
  echo "$u"
}

REPO_URL="$(get_repo_url)"
BRANCH="$(git -C "${ROOT_DIR}" rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"

write_wrapper() {
  local src="$1"
  local dst_rel="$2"
  local title="$3"

  local dst="${DOCS_DIR}/${dst_rel}"
  mkdir -p "$(dirname "${dst}")"

  if [[ ! -f "${ROOT_DIR}/${src}" ]]; then
    echo "WARN: missing source ${src} (skipping)" >&2
    return 0
  fi

  # Copy content and rewrite a few known cross-links to website routes.
  # Keep this intentionally small: only rewrite the canonical "project pages".
  local tmp
  tmp="$(mktemp)"
  cp "${ROOT_DIR}/${src}" "${tmp}"

  # Normalize common links: (PRICING.md) -> (/pricing/)
  sed -E -i \
    -e 's#\((\./)?PRICING\.md\)#(/pricing/)#g' \
    -e 's#\((\./)?COMMERCIAL\.md\)#(/commercial/)#g' \
    -e 's#\((\./)?LICENSING\.md\)#(/licensing/)#g' \
    -e 's#\((\./)?EVAL_LICENSE\.md\)#(/eval-license/)#g' \
    -e 's#\((\./)?SECURITY\.md\)#(/security/)#g' \
    -e 's#\((\./)?SUPPORT\.md\)#(/support/)#g' \
    -e 's#\((\./)?CONTRIBUTING\.md\)#(/contributing-repo/)#g' \
    -e 's#\((\./)?CODE_OF_CONDUCT\.md\)#(/code-of-conduct/)#g' \
    -e 's#\((\./)?GOVERNANCE\.md\)#(/governance/)#g' \
    -e 's#\((\./)?LICENSES-README\.md\)#(/licenses/)#g' \
    "${tmp}"

  # Templates: keep in repo, but link to GitHub (not published as site pages)
  sed -E -i \
    -e "s#\((\./)?commercial/COMMERCIAL_LICENSE_TEMPLATE\.md\)#(${REPO_URL}/blob/${BRANCH}/commercial/COMMERCIAL_LICENSE_TEMPLATE.md)#g" \
    -e "s#\((\./)?commercial/ORDER_FORM_TEMPLATE\.md\)#(${REPO_URL}/blob/${BRANCH}/commercial/ORDER_FORM_TEMPLATE.md)#g" \
    "${tmp}"

  cat > "${dst}" <<EOF
<!--
  AUTO-GENERATED FILE.
  Source: /${src}
  Regenerate via: ./scripts/docs_sync.sh
-->

# ${title}

EOF

  cat "${tmp}" >> "${dst}"
  rm -f "${tmp}"
}

# Map canonical repo markdown -> published URL via docs/<slug>/index.md
write_wrapper "PRICING.md"            "pricing/index.md"           "Pricing"
write_wrapper "COMMERCIAL.md"         "commercial/index.md"        "Commercial"
write_wrapper "LICENSING.md"          "licensing/index.md"         "Licensing"
write_wrapper "EVAL_LICENSE.md"       "eval-license/index.md"      "Evaluation License"
write_wrapper "SECURITY.md"           "security/index.md"          "Security"
write_wrapper "SUPPORT.md"            "support/index.md"           "Support"
write_wrapper "CONTRIBUTING.md"       "contributing-repo/index.md" "Contributing"
write_wrapper "CODE_OF_CONDUCT.md"    "code-of-conduct/index.md"   "Code of Conduct"
write_wrapper "GOVERNANCE.md"         "governance/index.md"        "Governance"
write_wrapper "LICENSES-README.md"    "licenses/index.md"          "Licenses"

echo "Docs sync complete."
