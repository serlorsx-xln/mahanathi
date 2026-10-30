# syntax=docker/dockerfile:1
# Multi-stage build: front end, then one static Go binary, then a minimal
# non-root runtime. The data directory is /data, owned by the runtime user
# (65532); the binary healthchecks itself, because distroless has no shell.
#
# Two targets:
#   deploy (default) — no front-end build; the binary serves the placeholder
#     page (the national sentence from /api/v1/now). This is what runs until
#     Phase 3 ships the real UI. `make docker` builds it.
#   full — builds web/ and embeds it, for testing the whole image once the
#     front end exists. `make docker-full` builds it.

ARG VERSION=dev

FROM golang:1.25-alpine AS server
# Re-declare so the stage sees the global VERSION (or its default, "dev").
ARG VERSION
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/mahanathi ./cmd/mahanathi \
    && mkdir -p /out/data

FROM node:24-alpine AS web
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY web/package.json web/
COPY e2e/package.json e2e/
RUN pnpm install --frozen-lockfile --filter @mahanathi/web...
COPY web/ web/
# The look-dev routes fetch geodata at /geodata/...; in dev the vite plugin
# serves geodata/assets, in production the embedded front end must carry
# them, so the build bakes them into web/public (vite copies public/ to
# dist/). The 160 MB PMTiles basemap is deliberately not baked - the app
# HEAD-checks it and ships terrain-only when absent, per ADR 0007.
COPY geodata/assets/borders.topo.json geodata/assets/rivers.geojson geodata/assets/provinces.topo.json web/public/geodata/
COPY geodata/assets/dem web/public/geodata/dem
RUN pnpm --filter @mahanathi/web build

# deploy: the server alone; / serves the placeholder page.
FROM gcr.io/distroless/static-debian12:nonroot AS deploy
COPY --from=server /out/mahanathi /mahanathi
COPY --from=server --chown=65532:65532 /out/data /data
# The app reads PORT; Coolify's web-app template probes 3000, so that is
# the baked-in default (MAHANATHI_ADDR still wins if an operator sets it).
ENV MAHANATHI_DATA_DIR=/data
ENV PORT=3000
EXPOSE 3000
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s --start-period=120s --retries=3 \
  CMD ["/mahanathi", "healthcheck"]
ENTRYPOINT ["/mahanathi"]

# full: the web build embedded, once Phase 3 ships it.
FROM golang:1.25-alpine AS full-server
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

FROM gcr.io/distroless/static-debian12:nonroot AS full
COPY --from=full-server /out/mahanathi /mahanathi
COPY --from=full-server --chown=65532:65532 /out/data /data
ENV MAHANATHI_DATA_DIR=/data
ENV PORT=3000
EXPOSE 3000
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=5s --start-period=120s --retries=3 \
  CMD ["/mahanathi", "healthcheck"]
ENTRYPOINT ["/mahanathi"]
