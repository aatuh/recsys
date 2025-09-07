FROM caddy:2.8.4-alpine

ARG MKCERT_VERSION=v1.4.4

RUN set -eux; \
    apk add --no-cache bash curl nss-tools; \
    curl -sSL \
    "https://github.com/FiloSottile/mkcert/releases/download/${MKCERT_VERSION}/mkcert-${MKCERT_VERSION}-linux-amd64" \
    -o /usr/local/bin/mkcert; \
    chmod +x /usr/local/bin/mkcert

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
