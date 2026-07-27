# Changelog

All notable changes to this project are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The wire protocol is versioned separately from the software — see [`docs/protocol.md`](docs/protocol.md#versioning).

---

## [Unreleased]

Pre-release. No published version yet.

### Thesis change — 2026-07-27

The product is now a **multi-model cross-verification system**, not a mobile-controlled coding agent ([ADR-013](docs/decisions.md#adr-013--the-product-is-cross-verification-not-an-agent)). Three findings forced it, in order:

1. **Subscription access is closed to third-party clients.** All three providers permit their coding plans only inside listed tools; Qwen's server returns *"Only available for Coding Agents"*, Kimi prohibits altering client identity by name, Z.AI bans on the third violation. The only technical route was impersonation, which this project had already ruled out in its own research brief ([ADR-012](docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).
2. **The original product already exists, twice.** Z.AI's ZCode ships multi-model selection, QR pairing to a desktop session, phone control and a 1.5× quota bonus no third party can match. OpenClaw is an open-source agent runtime with ~369k stars that already dispatches sub-agents over ACP.
3. **The remaining gap was measured, not guessed.** A project one model declared complete was put through a verification pass driven by a *different* model against its spec: missing translations, unwritten tests, a coverage regression, and real bugs.

What a vendor structurally cannot ship is the honest verdict *"our own model got this wrong."* That is the position this project takes.

### Added

- **ADR-012, ADR-013, ADR-014** — provider access, the thesis change, and the decision to orchestrate ACP agents rather than build one.
- [`acp-orchestrator`](.rla/specs/acp-orchestrator.md) spec (draft, P1): ACP sessions, unified approval queue, coder → gates → reviewer handoff, evidence bundle, catch-rate instrumentation.
- ACP capability matrix in [`protocol.md`](docs/protocol.md#acp-capability-matrix), including what `qwen serve` already provides: `POST /permission/:requestId`, SSE with `Last-Event-ID` replay, multi-client sessions, TLS.
- New protocol message kinds: `spec.draft`, `run.status`, `evidence.bundle`.

### Changed

- **Loop Engineering promoted from deferred vision item to the core specification.** Its deterministic gates already run; cross-model review is P1.
- **P1 is now the orchestrator.** The agent loop, tool framework, sandbox, danger classification, model router, context normalisation and provider contract tests are all removed ([ADR-014](docs/decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)). Schedule drops from 17–20 weeks to 14–18.
- **Vision phases renumbered** to X1–X4; multi-model orchestration moved to core.
- README, landing page and social preview repositioned from "coding agent on your machine" to "one model writes it, a different one proves it".
- `provider-contract` spec **archived** — its two questions were answered by research, and the answers made the other eight requirements obsolete.
- `subscription-auth` spec never written: the problem it would have solved (OAuth) does not exist, and the real problem is topology, not authentication.

### Superseded

- **ADR-005** (MVP is an agent) and **ADR-008** (Tier 0 only) — both by ADR-013.

### Known risk

Cross-verification rests on model judgement, and two models can share a blind spot. The maintainer's manual runs are evidence, not proof. Catch rate is instrumented from the first P1 run ([SPEC-acp-orchestrator-11](.rla/specs/acp-orchestrator.md)), and [X1](docs/vision-roadmap.md#x1--deeper-verification-tiers) is gated on that number.

---

## Earlier — P0 foundation

### Added — P0 foundation

- Architecture decision log ([`docs/decisions.md`](docs/decisions.md)) recording ADR-001 … ADR-010.
- Public documentation set: [`architecture.md`](docs/architecture.md), [`protocol.md`](docs/protocol.md), [`threat-model.md`](docs/threat-model.md), [`privacy.md`](docs/privacy.md), [`license-header.md`](docs/license-header.md), [`roadmap.md`](docs/roadmap.md), [`vision-roadmap.md`](docs/vision-roadmap.md), [`loop-engineering.md`](docs/loop-engineering.md).
- Governance: [`SECURITY.md`](SECURITY.md), [`CLA.md`](CLA.md), `CODEOWNERS`, issue templates.
- CI: Go, Flutter, licence-header and CLA workflows.
- Build scaffolding: `Makefile`, `go.mod`, `.golangci.yml`, `.editorconfig`, Go skeletons for `rla` and `rla-server`.
- Licence header templates for both licences: `LICENSE_HEADER` (AGPL), `LICENSE_HEADER_APACHE` (mobile).
- `deploy/` — Docker Compose for NATS JetStream and the relay server.
- Landing page: Open Graph and Twitter card metadata, waitlist capture.
- **Loop Engineering applied to this repository** ([ADR-011](docs/decisions.md#adr-011--the-project-is-built-with-its-own-loop)) — the deterministic half of the pipeline the product will later ship:
  - `scripts/gate` — 16 gates across tiers 0–3, with `verify`, `canary`, `spec` and `clear-cache`.
  - `.rla/PRINCIPLES.md` and `.rla/SECURITY-BASELINE.md` — the codebase constitution.
  - `.rla/specs/` — spec artifacts with `SPEC-…` ids and human ratification (checkpoint ①); `build-scaffolding` ratified, `provider-contract` awaiting.
  - **Zero-Touch AI enforced structurally** — relay packages cannot import provider, agent, tool or crypto code.
  - Additional conformance gates: credential-in-log detection, assertion-free test detection, spec fidelity (forward diff), licence boundary.
  - `COULD NOT VERIFY` as a status distinct from pass and fail; readiness blocked while any gate is in it.
  - Selective regression cache keyed on working-tree content — `git commit` invalidates nothing.
  - Coverage ratchet (`.rla/coverage-floor.txt`), committed and self-tightening.
  - `docs/development-loop.md` — the working method.

### Changed — P0 foundation

- **Naming unified on RemLinkAgent** (ADR-001). `remcode` → `rla`, `~/.remcode/` → `~/.rla/`, `.remcode/` → `.rla/`, `remcode://` → `remlinkagent://`, `remcode.app` → `remlinkagent.com`.
- **Split licensing** (ADR-002). Core stays AGPL-3.0-or-later; `mobile/` becomes Apache-2.0, resolving the conflict between GPL-family terms and App Store distribution.
- **End-to-end encryption moved into the MVP** (ADR-004). Relay payloads are ciphertext; README claims about relay visibility now match the architecture, with observable metadata enumerated in `privacy.md`.
- **MVP scope narrowed to the agent** (ADR-005). Interactive terminal mirroring (PTY) moved to X5; command execution is tool-based with captured output.
- **Positioning widened** (ADR-006). Any OpenAI-compatible provider is supported; Z.AI, Qwen and Kimi remain first-class.
- **Phase naming disambiguated** (ADR-007). `P0`–`P4` (MVP), `M1` (managed cloud), `X1`–`X6` (vision) — "V1" previously meant three different things across three documents.
- **Loop Engineering Tier 0 pulled into the MVP** (ADR-008); tiers 1–4 stay in X6.
- **Schedule re-estimated.** P0–P4 is 17–20 weeks for a solo developer, replacing the earlier 10–12.
- Roadmap gained the items it was missing: protocol versioning, JetStream retention, observability, multi-device pairing, streaming end to end, privacy/data inventory.
- CLA now required (ADR-003), reversing the earlier "no CLA" position; the reasoning is documented rather than silently changed.
- Public docs moved from private `documentation/` to `docs/` and translated to English (ADR-010).
- README rewritten: honest pre-release status, the always-on host requirement (ADR-009), precise relay-visibility claims, unverifiable competitor comparison removed.

### Fixed — P0 foundation

- **Nine known stdlib CVEs** — surfaced by the new Tier 3 gate on its first run. Minimum Go raised to 1.26.5 across `go.mod`, CI, Dockerfile and the contributing guide.
- Eight lint findings in the tooling, including an unbounded `exec.Command` (external gate commands now carry a 10-minute timeout) and a regex that silently failed to match tilde code fences.
- `gate verify -no-cache` silently ignored the flag — Go's flag package stops at the first non-flag argument, so the subcommand is now lifted out and flags work on either side of it.
- Every documentation link in `CONTRIBUTING.md` pointed at a git-ignored path and 404'd for anyone but the maintainer.
- `Makefile`, CI workflows, `SECURITY.md` and `CHANGELOG.md` were referenced by the contributing guide and PR template but did not exist.
- `license-ön-eki.md` renamed to `license-header.md` — non-ASCII filenames break URL references and cross-platform checkouts.
- Chat-session preamble removed from the PRD and the Loop Engineering document.
- Malformed `SPEC-<feature></feature>-NN` in the PRD, which rendered as an HTML tag.

### Known gaps

- `docs/og-image.svg` needs a 1200×630 PNG export for crawlers that reject SVG — `make og-image` on a machine with ImageMagick, Inkscape or `rsvg-convert`.
- The landing page waitlist form needs `WAITLIST_ENDPOINT` configured; until then it falls back to the GitHub watch link.
- `CLA.md` has not been reviewed by a lawyer.
- The Flutter project is not bootstrapped — `make mobile` prints the command (P0.7).

---

[Unreleased]: https://github.com/burakhalefoglu/RemLinkAgent/commits/main
