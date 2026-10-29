# ADR 0001: One repository, one Go binary that embeds the web build

- **Status:** accepted (Phase 1)
- **Date:** 2026-10-29
- **Deciders:** implementer, for review by the lead

## Context

มหานที is a single page with a small JSON API (`/api/v1/now`, station history, a 72-hour timeline) over one public data source, ThaiWater. The deadline is 2026-10-31 12:00 ICT. The lead deploys to the owner's Coolify server, and the brief's contract is `docker run -p 8080:8080 -v data:/data`. The owner will review the code line by line, so the number of moving parts matters as much as the code itself.

Facts measured in Phase 1 that bear on the shape of the system (details in `docs/research/thaiwater-api.md`):

- **Snapshot cost.** One `waterlevel_load` call returns all 804 stations in 1,395,146 bytes, 241,911 bytes on the wire with gzip, in 0.32 s.
- **History cost.** History comes only from `waterlevel_graph`, one call per station. A 72-hour timeline therefore needs about 800 calls. Each call is 0.9 to 2.8 KB gzipped and takes about 0.2 s.
- **Messy fields.** The data needs normalising in one place. Cadence is mixed (10-minute HII, hourly RID and others), timestamps carry no offset, and the bank references contain zero codes and a second datum (ADR 0003).
- **CORS.** ThaiWater's API reflects any `Origin` with `access-control-allow-credentials: true`. A request with `Origin: https://example.org` on 2026-10-29 was allowed. A browser could therefore call ThaiWater directly. This is observed behaviour, not a documented contract.
- **Planned tests.** The brief asks for an "API down, then the last snapshot message shows" end-to-end test, and for `/readyz` to report whether we hold data no older than 2 h.

## Options

1. **One monorepo, one Go binary that embeds `web/dist`, one container.** This is the lead's recommendation.
2. **The same monorepo, but deployed as two units:** the static web build on a static host or CDN, and the Go API in its own container.
3. **Node and TypeScript end to end** (for example Hono or a meta-framework with server routes), sharing types between the server and the client.
4. **Static only.** A scheduled job, such as a GitHub Actions cron, fetches ThaiWater every hour and publishes JSON next to the static site. There is no server.
5. **A client-only SPA** that calls ThaiWater from each visitor's browser, which the CORS behaviour above makes possible.
6. **Separate repositories** for the server and the web app. This sits on a different axis from the others and applies to options 1 to 3.

## Evaluation

| | 1 Embedded binary | 2 Static + API | 3 Node full stack | 4 Static cron | 5 Client-only | 
|---|---|---|---|---|---|
| Deploy units | 1 | 2 | 1 | CI job + static host | static host |
| ThaiWater load | one client: 24 snapshot calls a day plus backfill | same as 1 | same as 1 | same as 1 | **every visitor**: 242 KB each, and about 800 calls each for the timeline |
| 72 h timeline in ≤ 80 KB | yes, precomputed | yes | yes | yes, precomputed | no: assembled from about 800 calls in the browser |
| Last snapshot when ThaiWater is down | yes (SQLite) | yes | yes | yes (last published file) | no |
| Version skew between UI and API | impossible | possible | impossible | possible | n/a |
| Same origin (CSP `'self'`, no CORS) | yes | needs CORS or a proxy | yes | yes | needs `connect-src` to ThaiWater |
| Hourly freshness | our scheduler | our scheduler | our scheduler | GitHub's scheduled runs are best effort and can be delayed or skipped under load | live |
| Fits the brief's Go layout and review guide | yes | yes | no | partly | no |

Notes on the options that were not chosen:

- **Option 5 is rejected even though it is the cheapest to build.**
  - A share spike in a LINE group would become a spike on a public institution's API.
  - The timeline cannot meet its budget.
  - There is no fallback when ThaiWater is down.
  - We would depend on a CORS policy that looks unintended. Reflecting any origin *with credentials* is a common misconfiguration, not an invitation, and it could change without notice.
- **Option 4** loses the hourly guarantee and the `/readyz` contract, makes history a pile of published files, and moves an outward-facing scheduled job into CI. It remains a possible emergency fallback (see Consequences).
- **Option 3** would give one language and shared types. It is rejected for these reasons:
  - The brief, the package boundaries and the owner's review guide are written around Go.
  - Go gives a static binary with a pure-Go SQLite and `CGO_ENABLED=0`, so it runs on a distroless image.
  - `humanize`, the part a human must defend line by line, is a natural fit for Go table tests.
  - The cost of two languages is real but small. Only the Thai number and time formatter needs to exist on both sides, and the brief already plans for that.
- **Option 2** buys global edge caching that a Thai-audience page with `max-age=300` JSON does not need. In exchange it adds a second deploy, CORS or a proxy, and a way for the UI and the API to disagree.
- **Option 6** doubles the CI and review surface for a two-person review loop. In one repository, a change to the API and its UI lands in one commit.

## Decision

**Option 1.** One repository holds:
- `server/`: a Go module.
- `web/`: Vite, React and TypeScript.
- `e2e/`: Playwright.
- `geodata/`: asset scripts.
- `docs/`.

The JavaScript side is a pnpm workspace (`web`, `e2e`) with one lockfile at the root. The Go module path is `mahanathi`, not a URL, because the public repository URL is not known yet. Renaming it later is mechanical: run `go mod edit -module` and rewrite the imports.

The build is:
1. `web` builds to `web/dist`.
2. The build copies `web/dist` into `server/web/dist`, because `go:embed` cannot reach a parent directory.
3. `go build` embeds it.

`make build` and the `Dockerfile` both follow these steps. The server is the only ThaiWater client, so visitors never contact ThaiWater.

## Consequences

**Positive:**
- There is one artifact and one deploy.
- The UI and the API cannot drift apart.
- Everything is same-origin, so the CSP can stay at `'self'`.
- ThaiWater sees one polite client however many people visit.
- The Phase 1 image, which serves only `/healthz` so far, is 14.1 MB, runs as `nonroot` on `distroless/static-debian12`, and builds in about 52 s on this machine.

**Negative:**
- There are two toolchains, so CI has a Go job and a web job.
- The Go build depends on the output of the web build.
- Horizontal scaling would need a different store (ADR 0002). One process on one Coolify node is the intended scale.

**Must solve in Phase 2: a fresh clone without a web build.** `go test ./...` and `go vet` must keep working on a fresh clone that has no web build. `//go:embed` fails to compile when its pattern matches nothing. There are two options, both reversible:
- **A tracked placeholder.** Change `.gitignore` to `server/web/dist/*` with `!server/web/dist/.gitkeep`, embed `all:dist`, and have the server answer "web build not embedded" when `index.html` is missing.
- **A build tag.** Tag the embed file `embedweb` and build release binaries with `-tags embedweb`.

The placeholder is preferred, because a tag that is forgotten in one build path silently ships a binary with no UI.

**Must solve before go-live: the HTTP client's User-Agent.** It names the project but has no repository URL yet. It reads from config and has a default (see the open questions in the Phase 1 report).

**Local Docker builds on this machine.** This repository is on an exFAT volume, and BuildKit cannot read the extended attributes of the AppleDouble `._*` files there. `make docker` therefore streams a gzipped tar of the git-known files as the build context. On a normal Linux filesystem, such as CI or Coolify, a plain `docker build .` with `.dockerignore` should work. That has not been verified yet.

**Emergency fallback.** If the server cannot be hosted in time, the same Go ingest code could be run on a schedule to write `now.json` and `timeline.json` into the static build. That is option 4 with our code. Nothing in this decision blocks that.
