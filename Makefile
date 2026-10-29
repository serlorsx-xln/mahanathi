# One entry point for the common tasks. Every target runs from the repo root.
# `make help` lists them.

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

GO_DIR    := server
WEB       := pnpm --filter @mahanathi/web
E2E       := pnpm --filter @mahanathi/e2e
PRETTIER  := pnpm --dir web exec prettier --ignore-path ../.prettierignore
BIN       := $(GO_DIR)/bin/mahanathi
IMAGE     := mahanathi:dev

.PHONY: help install dev dev-server dev-web test test-go test-web e2e lint lint-go lint-web fmt build build-web embed-web docker docker-full clean

help: ## List targets
	@grep -E '^[a-z0-9-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

install: ## Install JS dependencies exactly as locked
	pnpm install --frozen-lockfile

dev: ## Run the Go server and the Vite dev server together (Ctrl-C stops both)
	@$(MAKE) -j2 --no-print-directory dev-server dev-web

dev-server:
	cd $(GO_DIR) && MAHANATHI_LOG_FORMAT=text go run ./cmd/mahanathi

dev-web:
	$(WEB) dev

test: test-go test-web ## Go tests (race detector) and web unit tests

test-go:
	cd $(GO_DIR) && go test -race -count=1 ./...

test-web:
	$(WEB) test

e2e: ## Playwright end-to-end tests (builds the web app first)
	$(E2E) test

lint: lint-go lint-web ## go vet, gofmt, tsc, ESLint and Prettier

lint-go:
	cd $(GO_DIR) && go vet ./...
	@unformatted="$$(cd $(GO_DIR) && gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then echo "gofmt needed:"; echo "$$unformatted"; exit 1; fi

lint-web:
	$(WEB) typecheck
	$(E2E) typecheck
	$(WEB) lint
	$(PRETTIER) --check . ../e2e

fmt: ## Format Go and front-end sources in place
	cd $(GO_DIR) && gofmt -w .
	$(PRETTIER) --write . ../e2e

build: build-web embed-web ## Build the web app, then one Go binary at server/bin/mahanathi
	cd $(GO_DIR) && CGO_ENABLED=0 go build -trimpath -o bin/mahanathi ./cmd/mahanathi
	@echo "built $(BIN)"

build-web:
	$(WEB) build

# go:embed cannot reach outside the module, so the build copies web/dist
# into server/web/dist for the server package to embed. The tracked
# .gitkeep stays, so `git status` stays clean and a fresh clone compiles.
embed-web:
	find $(GO_DIR)/web/dist -mindepth 1 ! -name .gitkeep -delete
	cp -R web/dist/. $(GO_DIR)/web/dist/
	find $(GO_DIR)/web/dist -name '._*' -delete

# The context is streamed as a tar of the files git knows about (tracked plus
# untracked-but-not-ignored). Docker cannot read the exFAT volume directly:
# BuildKit fails on the xattrs of macOS ._ side files before .dockerignore
# applies. COPYFILE_DISABLE stops tar from adding ._ entries of its own, and
# gzip gives Docker an unambiguous archive header (it misreads a plain pax
# stream from bsdtar as a Dockerfile).
docker: ## Build the deploy image (mahanathi:dev; no front end, placeholder page)
	git ls-files -z --cached --others --exclude-standard \
	  | COPYFILE_DISABLE=1 tar --null --no-mac-metadata -T - -czf - \
	  | docker build --target deploy --build-arg VERSION=$$(git describe --always --dirty) -t $(IMAGE) -

docker-full: ## Build the full image with the web build embedded (mahanathi:full)
	git ls-files -z --cached --others --exclude-standard \
	  | COPYFILE_DISABLE=1 tar --null --no-mac-metadata -T - -czf - \
	  | docker build --target full --build-arg VERSION=$$(git describe --always --dirty) -t $(IMAGE)-full -

clean: ## Remove build output
	rm -rf web/dist $(GO_DIR)/bin e2e/test-results e2e/playwright-report
	find $(GO_DIR)/web/dist -mindepth 1 ! -name .gitkeep -delete
