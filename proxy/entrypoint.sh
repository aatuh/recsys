#!/usr/bin/env bash
set -euo pipefail

# Check if required environment variables are provided
missing_vars=()
for var in WEB_DOMAIN API_DOMAIN SWAGGER_DOMAIN DEMO_DOMAIN SHOP_DOMAIN WEB_BACKEND API_BACKEND SWAGGER_BACKEND DEMO_BACKEND SHOP_BACKEND; do
  if [ -z "${!var+x}" ]; then
    missing_vars+=("$var")
  fi
done

if [ "${#missing_vars[@]}" -ne 0 ]; then
  echo "Error: Required environment variables are not set: ${missing_vars[*]}"
  echo "Please set WEB_DOMAIN, API_DOMAIN, SWAGGER_DOMAIN, DEMO_DOMAIN, SHOP_DOMAIN, WEB_BACKEND, API_BACKEND, SWAGGER_BACKEND, DEMO_BACKEND, and SHOP_BACKEND environment variables."
  exit 1
fi

CAROOT="/root/.local/share/mkcert"
CERT_DIR="/certs"

mkdir -p "${CAROOT}" "${CERT_DIR}"

# Make sure CA exists inside the container
if [[ ! -f "${CAROOT}/rootCA-key.pem" ]]; then
  mkcert --install
fi

# Generate certs for all required domains
for domain in "${WEB_DOMAIN}" "${API_DOMAIN}" "${SWAGGER_DOMAIN}" "${DEMO_DOMAIN}" "${SHOP_DOMAIN}"; do
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

${SWAGGER_DOMAIN} {
    tls ${CERT_DIR}/${SWAGGER_DOMAIN}.crt ${CERT_DIR}/${SWAGGER_DOMAIN}.key

    reverse_proxy ${SWAGGER_BACKEND}
}

${DEMO_DOMAIN} {
    tls ${CERT_DIR}/${DEMO_DOMAIN}.crt ${CERT_DIR}/${DEMO_DOMAIN}.key

    reverse_proxy ${DEMO_BACKEND}
}

${SHOP_DOMAIN} {
    tls ${CERT_DIR}/${SHOP_DOMAIN}.crt ${CERT_DIR}/${SHOP_DOMAIN}.key

    reverse_proxy ${SHOP_BACKEND}
}
EOF

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
