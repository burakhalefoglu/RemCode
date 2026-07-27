# RemLinkAgent — AI coding agent for your machine, driven from your phone
# Copyright (C) 2026 Burak Halefoğlu
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
#
# Windows without GNU make? Use the mirror: .\scripts\make.ps1 <target>

SHELL       := /bin/sh
.DEFAULT_GOAL := help

BIN         := bin
CLI         := $(BIN)/rla
SERVER      := $(BIN)/rla-server

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell git log -1 --format=%cI 2>/dev/null || echo "unknown")

PKG         := github.com/burakhalefoglu/RemLinkAgent
LDFLAGS     := -s -w \
	-X '$(PKG)/internal/version.version=$(VERSION)' \
	-X '$(PKG)/internal/version.commit=$(COMMIT)' \
	-X '$(PKG)/internal/version.date=$(BUILD_DATE)'

# Licence boundary — ADR-002. Core is AGPL, mobile/ is Apache-2.0.
AGPL_DIRS   := cmd internal deploy scripts
APACHE_DIRS := mobile/lib mobile/test
HOLDER      := Burak Halefoğlu
YEAR        := 2026
LIC_IGNORE  := -ignore "**/*.md" -ignore "**/*.json" -ignore "**/*.lock" \
	-ignore "**/vendor/**" -ignore "**/.dart_tool/**" -ignore "**/build/**" \
	-ignore "**/generated_plugin_registrant.dart" -ignore "**/*.g.dart" \
	-ignore "**/*.freezed.dart"

# Only pass directories that exist — the tree is still being built out.
existing = $(foreach d,$(1),$(wildcard $(d)))

.PHONY: help
help: ## Show this help
	@echo "RemLinkAgent — make targets"
	@echo ""
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Status: pre-alpha (P0). See docs/roadmap.md"

# ── Build ────────────────────────────────────────────────────────────────────

.PHONY: build
build: cli server ## Build both binaries

.PHONY: cli
cli: ## Build the CLI agent → bin/rla
	@mkdir -p $(BIN)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(CLI) ./cmd/rla
	@echo "built $(CLI) ($(VERSION))"

.PHONY: server
server: ## Build the relay server → bin/rla-server
	@mkdir -p $(BIN)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(SERVER) ./cmd/rla-server
	@echo "built $(SERVER) ($(VERSION))"

.PHONY: install
install: ## go install both binaries into GOPATH/bin
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/...

# ── Test & lint ──────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run Go tests
	go test ./...

.PHONY: test-race
test-race: ## Run Go tests with the race detector
	go test -race ./...

.PHONY: cover
cover: ## Coverage report → coverage.out / coverage.html
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "report: coverage.html"

.PHONY: lint
lint: vet ## Run golangci-lint (and flutter analyze when present)
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "skip: golangci-lint not installed → make tools"
	@test -f mobile/pubspec.yaml \
		&& (cd mobile && flutter analyze) \
		|| echo "skip: mobile/ not bootstrapped (P0.7)"

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format Go (and Dart when present)
	go fmt ./...
	@test -f mobile/pubspec.yaml && (cd mobile && dart format lib test) || true

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

# ── Mobile ───────────────────────────────────────────────────────────────────

.PHONY: mobile
mobile: ## Build/run the Flutter app (prints bootstrap command if absent)
	@if [ ! -f mobile/pubspec.yaml ]; then \
		echo "mobile/ is not bootstrapped yet — this is P0.7."; \
		echo ""; \
		echo "  flutter create --org com.remlinkagent --project-name rla_mobile \\"; \
		echo "                 --platforms=ios,android mobile"; \
		echo ""; \
		echo "Keep mobile/LICENSE and mobile/README.md — they carry the Apache-2.0"; \
		echo "boundary described in ADR-002."; \
		exit 1; \
	fi
	cd mobile && flutter pub get && flutter run

.PHONY: mobile-test
mobile-test: ## Run Flutter tests
	@test -f mobile/pubspec.yaml || { echo "mobile/ not bootstrapped (P0.7)"; exit 1; }
	cd mobile && flutter test

# ── Docker ───────────────────────────────────────────────────────────────────

.PHONY: docker-up
docker-up: ## Start NATS + relay server
	cd deploy && docker compose up -d
	@echo "NATS on :4222 (monitor :8222) · server on :8080"

.PHONY: docker-down
docker-down: ## Stop the stack
	cd deploy && docker compose down

.PHONY: docker-logs
docker-logs: ## Tail stack logs
	cd deploy && docker compose logs -f

# ── Licence headers (ADR-002) ────────────────────────────────────────────────

.PHONY: license-check
license-check: ## Verify every source file carries the right header
	@command -v addlicense >/dev/null 2>&1 || { \
		echo "addlicense missing → make tools"; exit 1; }
	@echo "checking AGPL headers: $(call existing,$(AGPL_DIRS))"
	@addlicense -check -f LICENSE_HEADER -c "$(HOLDER)" -y $(YEAR) $(LIC_IGNORE) \
		$(call existing,$(AGPL_DIRS))
	@if [ -d mobile/lib ]; then \
		echo "checking Apache-2.0 headers: $(call existing,$(APACHE_DIRS))"; \
		addlicense -check -f LICENSE_HEADER_APACHE -c "$(HOLDER)" -y $(YEAR) $(LIC_IGNORE) \
			$(call existing,$(APACHE_DIRS)); \
	fi
	@echo "licence headers OK"

.PHONY: license-add
license-add: ## Add the correct header to files missing one
	@command -v addlicense >/dev/null 2>&1 || { \
		echo "addlicense missing → make tools"; exit 1; }
	addlicense -f LICENSE_HEADER -c "$(HOLDER)" -y $(YEAR) $(LIC_IGNORE) \
		$(call existing,$(AGPL_DIRS))
	@if [ -d mobile/lib ]; then \
		addlicense -f LICENSE_HEADER_APACHE -c "$(HOLDER)" -y $(YEAR) $(LIC_IGNORE) \
			$(call existing,$(APACHE_DIRS)); \
	fi

# ── Misc ─────────────────────────────────────────────────────────────────────

.PHONY: tools
tools: ## Install dev tooling (addlicense, golangci-lint)
	go install github.com/google/addlicense@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

.PHONY: og-image
og-image: ## Export docs/og-image.svg → og-image.png (needs a converter)
	@if command -v rsvg-convert >/dev/null 2>&1; then \
		rsvg-convert -w 1200 -h 630 docs/og-image.svg -o docs/og-image.png; \
	elif command -v magick >/dev/null 2>&1; then \
		magick -background none -density 144 docs/og-image.svg -resize 1200x630 docs/og-image.png; \
	elif command -v inkscape >/dev/null 2>&1; then \
		inkscape docs/og-image.svg -w 1200 -h 630 -o docs/og-image.png; \
	else \
		echo "no converter found — install librsvg, ImageMagick or Inkscape"; exit 1; \
	fi
	@echo "wrote docs/og-image.png — now point og:image at it in docs/index.html"

.PHONY: docs
docs: ## Verify every relative doc link and heading anchor resolves
	go run ./scripts/checkdocs

# ── Loop Engineering gates (docs/development-loop.md) ────────────────────────

.PHONY: t0
t0: ## Tier 0 — format, compile, vet (after every edit)
	go run ./scripts/gate t0

.PHONY: t1
t1: ## Tier 1 — lint, tests, conformance, fake-green (every iteration)
	go run ./scripts/gate t1

.PHONY: t2
t2: ## Tier 2 — coverage, spec fidelity, licences (at convergence)
	go run ./scripts/gate t2

.PHONY: t3
t3: ## Tier 3 — race detector, CVE scan (candidate-complete)
	go run ./scripts/gate t3

.PHONY: verify
verify: ## Full sweep + checkpoint ① — is this ready for a human to test?
	go run ./scripts/gate verify

.PHONY: canary
canary: ## Prove each gate still detects deliberate breakage
	go run ./scripts/gate canary

.PHONY: spec
spec: ## Spec artifact status
	go run ./scripts/gate spec

.PHONY: version
version: ## Print build metadata
	@echo "version : $(VERSION)"
	@echo "commit  : $(COMMIT)"
	@echo "date    : $(BUILD_DATE)"

.PHONY: clean
clean: ## Remove build artefacts
	rm -rf $(BIN) coverage.out coverage.html
	@test -f mobile/pubspec.yaml && (cd mobile && flutter clean) || true

.PHONY: ci
ci: fmt vet lint test license-check docs ## Everything CI runs
