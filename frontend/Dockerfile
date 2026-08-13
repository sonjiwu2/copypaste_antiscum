# syntax=docker/dockerfile:1.7

# Multi-stage production image for the Anti-Scam Trainer client.
# Build context: frontend/  (see compose.yaml and README)
#
#   docker build -f frontend/Dockerfile -t antiscam-frontend:test ./frontend

ARG NODE_VERSION=22.14.0
ARG NGINX_VERSION=1.27-alpine

# -----------------------------------------------------------------------------
# Builder
# -----------------------------------------------------------------------------
FROM node:${NODE_VERSION}-alpine AS builder

WORKDIR /src

# Dependency layer first: source-only changes reuse the install layer.
# `npm ci` installs exactly the lockfile, so the image cannot drift from it.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# The build typechecks first: a broken contract with the API must stop the
# image, not ship to the browser.
RUN npm run build

# -----------------------------------------------------------------------------
# Runtime
# -----------------------------------------------------------------------------
FROM nginx:${NGINX_VERSION} AS runtime

LABEL org.opencontainers.image.title="antiscam-frontend" \
	org.opencontainers.image.description="Anti-Scam Trainer web client" \
	org.opencontainers.image.source="https://github.com/sonjiwu2/copypaste_antiscum" \
	org.opencontainers.image.licenses="MIT"

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /src/dist /usr/share/nginx/html

# Port 8080, not 80: compose maps the published port onto it, and the same
# number works if the container is later moved to an unprivileged user.
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8080/index.html >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]
