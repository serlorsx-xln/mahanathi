# syntax=docker/dockerfile:1
# Multi-stage build: front end, then one static Go binary, then a minimal
# non-root runtime. The data directory is /data, owned by the runtime user
# (65532); the binary healthchecks itself, because distroless has no shell.

FROM node:24-alpine AS web
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY web/package.json web/
COPY e2e/package.json e2e/
RUN pnpm install --frozen-lockfile --filter @mahanathi/web...
COPY web/ web/
RUN pnpm --filter @mahanathi/web build

ARG VERSION=dev

FROM golang:1.25-alpine AS server
# Re-declare so the stage sees the global VERSION (or its default, "dev").
ARG VERSION
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/mahanathi ./cmd/mahanathi \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/mahanathi /mahanathi
COPY --from=server --chown=65532:65532 /out/data /data
# PORT (what Coolify passes) or the default :8080; MAHANATHI_ADDR still wins.
ENV MAHANATHI_DATA_DIR=/data
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s --start-period=120s --retries=3 \
  CMD ["/mahanathi", "healthcheck"]
ENTRYPOINT ["/mahanathi"]
