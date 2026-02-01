#!/usr/bin/env bash
set -euo pipefail

# Proxy configuration supports:
# - PROXY_ROUTES: "api.app.local=api:8000"
# - Fallback explicit pairs: API_DOMAIN/API_BACKEND, etc.

declare -a PAIRS=()

if [[ -n "${PROXY_ROUTES:-}" ]]; then
  IFS=', ' read -ra raw_routes <<<"${PROXY_ROUTES}"
  for route in "${raw_routes[@]}"; do
    [[ -z "${route}" ]] && continue
    domain="${route%%=*}"
    backend="${route#*=}"
    if [[ -z "${domain}" || -z "${backend}" || "${domain}" == "${backend}" ]]; then
      echo "⚠️  Skipping malformed route '${route}', expected 'domain=backend'"
      continue
    fi
    PAIRS+=("${domain}|${backend}")
  done
else
  prefixes=(WEB API SWAGGER DEMO SHOP)
  for p in "${prefixes[@]}"; do
    domain_var="${p}_DOMAIN"
    backend_var="${p}_BACKEND"
    domain="${!domain_var:-}"
    backend="${!backend_var:-}"
    if [[ -n "${domain}" && -n "${backend}" ]]; then
      PAIRS+=("${domain}|${backend}")
    fi
  done
fi

if [[ ${#PAIRS[@]} -eq 0 ]]; then
  echo "Error: no proxy routes configured. Use PROXY_ROUTES or *_DOMAIN/*_BACKEND variables."
  exit 1
fi

CAROOT="/root/.local/share/mkcert"
CERT_DIR="/certs"

mkdir -p "${CAROOT}" "${CERT_DIR}"

# Make sure CA exists inside the container
if [[ ! -f "${CAROOT}/rootCA-key.pem" ]]; then
  mkcert --install
fi

CADDYFILE="/etc/caddy/Caddyfile"
: > "${CADDYFILE}"

# Generate certs and Caddyfile for all configured routes.
for pair in "${PAIRS[@]}"; do
  domain="${pair%%|*}"
  backend="${pair##*|}"

  if [[ ! -f "${CERT_DIR}/${domain}.crt" ]]; then
    echo "⏳ generating cert for ${domain}" >&2
    mkcert -key-file "${CERT_DIR}/${domain}.key" \
           -cert-file "${CERT_DIR}/${domain}.crt" \
           "${domain}"
  fi

  cat <<SITE >> "${CADDYFILE}"
${domain} {
    tls ${CERT_DIR}/${domain}.crt ${CERT_DIR}/${domain}.key

    reverse_proxy ${backend}
}

SITE
done

echo "---- Generated Caddyfile ----"
sed -n '1,120p' /etc/caddy/Caddyfile
echo "-----------------------------"

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
