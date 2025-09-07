#!/usr/bin/env bash
set -euo pipefail

# Check if required environment variables are provided
if [ -z "${DOMAIN+x}" ] || [ -z "${WEB_BACKEND+x}" ]; then
  echo "Error: Required environment variables are not set."
  echo "Please set DOMAIN and WEB_BACKEND environment variables."
  exit 1
fi

CAROOT="/root/.local/share/mkcert"
CERT_DIR="/certs"

mkdir -p "${CAROOT}" "${CERT_DIR}"

# Make sure CA exists inside the container
if [[ ! -f "${CAROOT}/rootCA-key.pem" ]]; then
  mkcert --install
fi

# Generate cert only when absent
if [[ ! -f "${CERT_DIR}/${DOMAIN}.crt" ]]; then
  echo "⏳ generating cert for ${DOMAIN}"
  mkcert -key-file "${CERT_DIR}/${DOMAIN}.key" \
         -cert-file "${CERT_DIR}/${DOMAIN}.crt" \
         "${DOMAIN}"
fi

cat > /etc/caddy/Caddyfile <<EOF
${DOMAIN} {
    tls ${CERT_DIR}/${DOMAIN}.crt ${CERT_DIR}/${DOMAIN}.key

    reverse_proxy ${WEB_BACKEND}
}
EOF

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
