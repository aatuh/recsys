#!/usr/bin/env bash
set -euo pipefail

# Check if required environment variables are provided
if [ -z "${WEB_DOMAIN+x}" ] || [ -z "${API_DOMAIN+x}" ] || [ -z "${WEB_BACKEND+x}" ] || [ -z "${API_BACKEND+x}" ]; then
  echo "Error: Required environment variables are not set."
  echo "Please set WEB_DOMAIN, API_DOMAIN, WEB_BACKEND, and API_BACKEND environment variables."
  exit 1
fi

CAROOT="/root/.local/share/mkcert"
CERT_DIR="/certs"

mkdir -p "${CAROOT}" "${CERT_DIR}"

# Make sure CA exists inside the container
if [[ ! -f "${CAROOT}/rootCA-key.pem" ]]; then
  mkcert --install
fi

# Generate certs for both domains
for domain in "${WEB_DOMAIN}" "${API_DOMAIN}"; do
  if [[ ! -f "${CERT_DIR}/${domain}.crt" ]]; then
    echo "⏳ generating cert for ${domain}"
    mkcert -key-file "${CERT_DIR}/${domain}.key" \
           -cert-file "${CERT_DIR}/${domain}.crt" \
           "${domain}"
  fi
done

cat > /etc/caddy/Caddyfile <<EOF
${WEB_DOMAIN} {
    tls ${CERT_DIR}/${WEB_DOMAIN}.crt ${CERT_DIR}/${WEB_DOMAIN}.key

    reverse_proxy ${WEB_BACKEND}
}

${API_DOMAIN} {
    tls ${CERT_DIR}/${API_DOMAIN}.crt ${CERT_DIR}/${API_DOMAIN}.key

    reverse_proxy ${API_BACKEND}
}
EOF

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
