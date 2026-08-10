# syntax=docker/dockerfile:1.7

# Multi-stage production image for the Anti-Scam Trainer API.
# Build context: backend/  (see compose.yaml and README)
#
#   docker build -f backend/Dockerfile -t antiscam-backend:test ./backend

ARG GO_VERSION=1.25.12
ARG GO_IMAGE_VARIANT=alpine3.23
ARG ALPINE_VERSION=3.23.5

# -----------------------------------------------------------------------------
# Builder
# -----------------------------------------------------------------------------
FROM golang:${GO_VERSION}-${GO_IMAGE_VARIANT} AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

# Dependency layer first for better cache hits when only source changes.
# The dependency manifests are copied first so source-only changes reuse the
# module-download layer.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

COPY . .

# Static binaries: no CGO, stripped symbols, reproducible-friendly flags.
# `migrate` embeds backend/migrations/*.sql, so the schema always ships with
# the code that expects it — no separate migration layer to forget.
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build \
		-trimpath \
		-ldflags="-s -w" \
		-o /out/api \
		./cmd/api \
	&& CGO_ENABLED=0 GOOS=linux go build \
		-trimpath \
		-ldflags="-s -w" \
		-o /out/migrate \
		./cmd/migrate

# -----------------------------------------------------------------------------
# Runtime
# -----------------------------------------------------------------------------
FROM alpine:${ALPINE_VERSION} AS runtime

LABEL org.opencontainers.image.title="antiscam-backend" \
	org.opencontainers.image.description="Anti-Scam Trainer HTTP API" \
	org.opencontainers.image.source="https://github.com/sonjiwu2/copypaste_antiscum" \
	org.opencontainers.image.licenses="MIT"

# wget: HEALTHCHECK without a fat image; ca-certificates for future HTTPS clients.
RUN apk add --no-cache ca-certificates wget \
	&& adduser -D -H -u 65532 -g app app

COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/migrate /usr/local/bin/migrate

USER app:app

EXPOSE 8080

# Defaults only — override at runtime via compose/env. No secrets here.
ENV HTTP_ADDR=:8080 \
	LOG_LEVEL=info

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

# PID 1 receives SIGTERM; application already handles graceful shutdown.
ENTRYPOINT ["/usr/local/bin/api"]
