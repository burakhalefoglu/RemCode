# Changelog

All notable changes to this project are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

The wire protocol is versioned separately from the software — see [`docs/protocol.md`](docs/protocol.md#versioning).

---

## [Unreleased]

Pre-release. No published version yet; **P0** is the first implementation phase.

### Added

- Architecture decision log ([`docs/decisions.md`](docs/decisions.md)) recording ADR-001 … ADR-010.
- Public documentation set: [`architecture.md`](docs/architecture.md), [`protocol.md`](docs/protocol.md), [`threat-model.md`](docs/threat-model.md), [`privacy.md`](docs/privacy.md), [`license-header.md`](docs/license-header.md), [`roadmap.md`](docs/roadmap.md), [`vision-roadmap.md`](docs/vision-roadmap.md), [`loop-engineering.md`](docs/loop-engineering.md).
- Governance: [`SECURITY.md`](SECURITY.md), [`CLA.md`](CLA.md), `CODEOWNERS`, issue templates.
- CI: Go, Flutter, licence-header and CLA workflows.
- Build scaffolding: `Makefile`, `go.mod`, `.golangci.yml`, `.editorconfig`, Go skeletons for `rla` and `rla-server`.
- Licence header templates for both licences: `LICENSE_HEADER` (AGPL), `LICENSE_HEADER_APACHE` (mobile).
- `deploy/` — Docker Compose for NATS JetStream and the relay server.
- Landing page: Open Graph and Twitter card metadata, waitlist capture.

### Changed

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

### Fixed

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
