# syntax=docker/dockerfile:1
# Multi-stage build: front end, then one static Go binary, then a minimal
# non-root runtime. Phase 6 adds the HEALTHCHECK and the /data volume once
# the server has storage.

FROM node:24-alpine AS web
WORKDIR /src
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY web/package.json web/
COPY e2e/package.json e2e/
RUN pnpm install --frozen-lockfile --filter @mahanathi/web...
COPY web/ web/
RUN pnpm --filter @mahanathi/web build

FROM golang:1.25-alpine AS server
WORKDIR /src/server
COPY server/go.mod ./
RUN go mod download
COPY server/ ./
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mahanathi ./cmd/mahanathi

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/mahanathi /mahanathi
ENV MAHANATHI_ADDR=:8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/mahanathi"]
