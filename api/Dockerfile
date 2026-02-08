FROM golang:1.25.7-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG TARGETOS=linux
ARG TARGETARCH

ENV CGO_ENABLED=0

RUN GOOS=$TARGETOS GOARCH=${TARGETARCH:-$(go env GOARCH)} \
	go build -trimpath \
	-ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
	-o /out/recsys-service \
	./cmd/recsys-service

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
	&& addgroup -S app \
	&& adduser -S app -G app \
	&& mkdir -p /tmp \
	&& chown app:app /tmp

WORKDIR /app

COPY --from=builder /out/recsys-service /app/recsys-service

EXPOSE 8000

USER app

ENV TMPDIR=/tmp

HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8000/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/recsys-service"]
