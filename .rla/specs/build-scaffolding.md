---
id: build-scaffolding
title: Build scaffolding, version surface and fail-loud skeleton
status: ratified
phase: P0
created: 2026-07-27
ratified: 2026-07-27
---

## Goal

Give the repository a build that compiles, tests, lints and verifies itself
from day one — and a binary skeleton that is honest about how little it does.

This spec was written **after** the code, as a retro-spec ([`loop-engineering.md` §
adopting into an existing codebase](../../docs/loop-engineering.md#adopting-into-an-existing-codebase)).
That is the documented path for bringing an existing tree under the loop, and
this is the project doing it to itself.

## Non-goals

- Any actual agent, relay or pairing behaviour — those are P1 and P2.
- Cobra command structure (P0.3) and the Flutter project (P0.7).

## Invariants

- A build that reports a version must report a *real* one, or an obvious
  placeholder. Never an empty string — an unversioned bug report is unactionable.
- Nothing may claim to work when it does not. Unimplemented surfaces exit
  non-zero.
- The wire protocol version is independent of the software version.

## Requirements

## SPEC-build-scaffolding-01 — Build metadata is stamped at link time

`version`, `commit` and `date` are injected via `-ldflags -X`. An unstamped
build reports documented `dev` defaults rather than empty strings, and the
stamped values are unexported so nothing can mutate them at runtime.

**Verified by:** `internal/version` unit tests; `make version`.

## SPEC-build-scaffolding-02 — Protocol version is independent of software version

The wire protocol version is a separate integer constant, reported alongside
build metadata. CLI, relay and mobile ship on independent cadences and must
negotiate on the protocol number alone.

**Verified by:** `internal/version` unit tests; [`docs/protocol.md`](../../docs/protocol.md#versioning).

## SPEC-build-scaffolding-03 — Unimplemented CLI commands fail loudly

Every command named in `rla help` either works or exits non-zero stating the
phase that will implement it. A declared-but-absent command never exits 0.
Unknown commands exit with a usage error, distinct from not-implemented.

**Verified by:** manual smoke test; enforced by principle P3.

## SPEC-build-scaffolding-04 — The relay exposes health and version endpoints

`rla-server` serves `GET /healthz` (plain text) and `GET /version` (JSON) so
deployment plumbing can be built and verified before the WebSocket gateway
exists. The container healthcheck depends on `/healthz`.

**Verified by:** `deploy/docker-compose.yml` healthcheck; manual curl.

## SPEC-build-scaffolding-05 — The relay shuts down gracefully

SIGINT and SIGTERM drain in-flight requests within a bounded timeout rather
than severing connections, so `docker compose down` and Ctrl-C behave.

**Verified by:** manual signal test.

## SPEC-build-scaffolding-06 — Invalid configuration fails loudly

An unrecognised log level exits non-zero with the accepted values, rather than
silently defaulting. A typo in a deployment config must be visible.

**Verified by:** manual smoke test; principle P3.

## SPEC-build-scaffolding-07 — Documentation links are machine-verified

Every relative Markdown link and heading anchor in the repository resolves,
checked in CI. This exists because the project previously shipped a
contributing guide whose documentation links all 404'd.

**Verified by:** `scripts/checkdocs`; `gate t2 → doc-links`; CI job `docs`.

---

## Acceptance

- [x] `make cli && make server` produces binaries.
- [x] `./bin/rla version` prints version, commit, platform and protocol.
- [x] `./bin/rla help` lists commands and marks the unimplemented ones.
- [x] `make test lint license-check docs` is green.
- [x] `docker compose config` validates.

## Open questions

None. Ratified 2026-07-27.
