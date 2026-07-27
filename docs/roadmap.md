# 🗺️ RemLinkAgent — MVP Roadmap

> **Version:** 4.0 · **Updated:** 2026-07-27 · **Status:** P0 starting
> **Goal:** ship a working, store-ready MVP — then learn from real users.

Covers **P0–P4** (the MVP) and **M1** (managed cloud, the sustainability step). Everything beyond that lives in [`vision-roadmap.md`](vision-roadmap.md). The reasoning behind the scope choices is in [`decisions.md`](decisions.md).

**Tracking:** `[ ]` todo → `[~]` in progress → `[x]` done. `⚠️` blocked, `➕` optional.

---

## 1. What the MVP is

An **AI coding agent** that runs on the developer's machine and is driven from their phone.

**In scope**

- ✅ QR pairing between CLI host and phone
- ✅ BYOK — API keys in the OS keychain, never transmitted
- ✅ Agent loop: read/write files, run commands, iterate on the output
- ✅ **Model hot-swap** mid-conversation with context preserved — the core differentiator
- ✅ Dangerous-command detection with mobile approval
- ✅ **Tier 0 verification** — format/lint/type-check after every agent edit ([ADR-008](decisions.md#adr-008--loop-engineering-tier-0-ships-in-the-mvp))
- ✅ **End-to-end encryption** of everything crossing the relay ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads))
- ✅ Streaming responses, end to end
- ✅ Background sync — JetStream replay, no lost events
- ✅ Push notifications for approvals and completion
- ✅ Single project

**Out of scope** → [`vision-roadmap.md`](vision-roadmap.md)

- ❌ Interactive terminal mirroring / PTY ([ADR-005](decisions.md#adr-005--mvp-is-an-agent-not-a-terminal-mirror)) → X5
- ❌ Loop Engineering tiers 1–4 → X6
- ❌ Role-based model delegation → X1
- ❌ Multi-project → X1
- ❌ Voice, smartwatches → X2–X4
- ❌ Managed cloud, teams, billing → M1

### Invariants — not negotiable

| | |
| :--- | :--- |
| 🔑 **Zero-Touch AI** | AI traffic and API keys never reach the relay. |
| 🔐 **E2E encrypted** | The relay handles ciphertext and routing metadata. Nothing else. |
| 🪶 **Light** | One static binary, < 30 MB RSS; mobile cold start < 2 s. |
| 🔁 **Zero-loss** | JetStream replay from the last acknowledged sequence. |
| 🟢 **Fail-loud** | Never a silent green. Halt and say so. |
| 🌍 **Localised** | TR + EN, i18n keys only — never hard-coded strings. |

---

## 2. Schedule

**17–20 weeks for one developer working full time.**

| Phase | Estimate |
| :--- | :--- |
| [P0](#p0--scaffolding--de-risking) — Scaffolding & de-risking | 1.5–2 weeks |
| [P1](#p1--cli-agent) — CLI agent | 4–5 weeks |
| [P2](#p2--server-backend-wss--nats) — Relay server | 3 weeks |
| [P3](#p3--flutter-mobile-app) — Flutter app | 5–6 weeks |
| [P4](#p4--release--distribution) — Release | 3–4 weeks |
| [M1](#m1--managed-cloud-subscriptions--team) — Managed cloud | 5–6 weeks (after MVP) |

An earlier draft said 10–12 weeks. That estimate assumed no App Store review queue (1–2 weeks on its own), no push-certificate setup, and no integration time between phases. The number above is what the same plan costs when those are counted. Dropping terminal mirroring ([ADR-005](decisions.md#adr-005--mvp-is-an-agent-not-a-terminal-mirror)) is what keeps it from being worse.

---

## 3. Stack

| Layer | Choice | Notes |
| :--- | :--- | :--- |
| CLI agent | **Go** | Static binary; macOS/Linux/Windows, amd64 + arm64 |
| Relay | **Go** | `net/http` + `gorilla/websocket` |
| Queue | **NATS JetStream** | Persistent, replayable |
| Database | **SQLite** | Zero-config |
| Transport | **WSS** | TLS on 443, E2E-encrypted payloads inside |
| Mobile | **Flutter** | iOS + Android, one codebase |
| AI client | **`sashabaranov/go-openai`** | Any OpenAI-compatible endpoint ([ADR-006](decisions.md#adr-006--provider-neutral-core-first-class-zaiqwenkimi)) |
| Crypto | **`golang.org/x/crypto/nacl/box`** + `cryptography_flutter` | X25519 + XSalsa20-Poly1305 |
| Licence | **AGPL-3.0-or-later** core, **Apache-2.0** mobile | [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile) |

Flutter packages: `flutter_riverpod`, `go_router`, `web_socket_channel`, `flutter_secure_storage`, `mobile_scanner`, `firebase_messaging`, `flutter_markdown`, `drift`, `logger`, `cryptography`.

Architecture detail: [`architecture.md`](architecture.md) · Wire format: [`protocol.md`](protocol.md)

---

## P0 — Scaffolding & de-risking

**Goal:** a repository that builds, tests and lints itself — and answers the two questions that could invalidate the P1 design.
**Estimate:** 1.5–2 weeks

### Scaffolding

- [x] **0.1** Monorepo layout: `cmd/`, `internal/`, `mobile/`, `deploy/`, `docs/`, `scripts/`.
- [x] **0.2** Go module; minimal `cmd/rla` and `cmd/rla-server`.
- [x] **0.4** `Makefile` + `scripts/make.ps1` (Windows mirror — GNU make is not standard there).
- [x] **0.5** `golangci-lint` + `gofmt`/`goimports` config.
- [x] **0.6** CI: `ci-go` (lint, vet, test, build on Linux/macOS/Windows).
- [x] **0.12** `CONTRIBUTING.md`, `SECURITY.md`, `CHANGELOG.md`, `CLA.md`, PR/issue templates, `CODEOWNERS`.
- [x] **0.14** `.editorconfig`, `.gitattributes`, `dependabot.yml`.
- [x] **0.15** Licence headers: dual templates, `make license-check`/`license-add`, `license-check` CI, boundary guard against AGPL leaking into `mobile/`.
- [x] **0.16** ADR log ([`decisions.md`](decisions.md)) — ADR-001 … ADR-010.
- [ ] **0.3** CLI framework: [`spf13/cobra`](https://github.com/spf13/cobra) + [`spf13/viper`](https://github.com/spf13/viper), replacing the skeleton dispatch.
- [ ] **0.7** Bootstrap Flutter: `flutter create --org com.remlinkagent --project-name rla_mobile --platforms=ios,android mobile`. Preserve `mobile/LICENSE` and `mobile/README.md`.
- [ ] **0.8** Flutter architecture skeleton: `app/`, `core/`, `features/`, `shared/` + `ProviderScope` + `go_router`.
- [ ] **0.9** Flutter dependencies + `analysis_options.yaml`.
- [ ] **0.10** SQLite migrations ([`pressly/goose`](https://github.com/pressly/goose)).
- [x] **0.11** `deploy/docker-compose.yml`: NATS with JetStream + relay.
- [~] **0.13** Version stamping via ldflags and `rla version` are in place; `pubspec.yaml` version still pending on 0.7.

### De-risking spikes — do these before P1 design is final

- [ ] **0.17** ⚠️ **Provider capability matrix.** For Z.AI, Qwen and Kimi, verify with real API calls:
  - [ ] **0.17.1** Tool/function calling: schema shape, parallel calls, whether `tool_choice` is honoured.
  - [ ] **0.17.2** Streaming: SSE chunk format, how tool-call deltas arrive, finish reasons.
  - [ ] **0.17.3** Context window, token accounting, rate-limit headers.
  - [ ] **0.17.4** Write the results into [`protocol.md`](protocol.md#provider-contract-tests) as executable contract tests.

  > **Why this gates everything.** P1 assumes one `go-openai` wrapper covers all three. That assumption holds for chat completion and is much weaker for tool calling and streaming tool deltas. If it breaks, the provider adapter layer needs a different design — and it is far cheaper to learn that now than in week 6.

- [ ] **0.18** ⚠️ **Cross-platform command execution.** Non-interactive execution with captured stdout/stderr, exit codes, timeouts and cancellation on Windows/macOS/Linux. Much smaller than PTY, but shell quoting and signal handling still differ per platform.

- [ ] **0.19** **E2E crypto shape.** Confirm the pairing key exchange and payload envelope in [`protocol.md`](protocol.md#encryption) round-trip between Go and Dart. A crypto mismatch discovered in P3 is very expensive.

**Done when**

- `make cli && make server` produces binaries; `make test lint license-check` is green.
- `make docker-up` brings up NATS + relay; `/healthz` answers.
- Flutter app runs on a simulator.
- All three spikes have written conclusions in `docs/`.

---

## P1 — CLI agent

**Goal:** an agent that does real work locally, with a model router behind it.
**Estimate:** 4–5 weeks · **Depends on:** P0

**Technical decisions**

- One `go-openai` wrapper; per-provider `baseURL` + model mapping + adapter for the deltas found in P0.17.
- Keyring backends: macOS Keychain, Windows Credential Manager, Linux libsecret.
- The agent loop is the core of the product. It gets built and tested first.

### Providers & keys

- [ ] **1.1** `rla login` — read keys with a hidden prompt (`term.ReadPassword`).
- [ ] **1.2** OS keychain via [`zalando/go-keyring`](https://github.com/zalando/go-keyring); no plaintext on disk; `rla logout` clears everything.
- [ ] **1.3** OpenAI-compatible client; presets for Z.AI/Qwen/Kimi; arbitrary endpoints by config.
- [ ] **1.4** Secret redaction in every log path — verified by test, not by inspection.

### Agent loop — the core

- [ ] **1.5** **Tool framework:** declaration, JSON-schema validation, dispatch, result marshalling.
- [ ] **1.6** **Filesystem tools:** read, write, list, search — confined to configured working directories. Escaping the sandbox is a security bug.
- [ ] **1.7** **Exec tool:** run a command, capture stdout/stderr/exit code, enforce a timeout, support cancellation (built on P0.18).
- [ ] **1.8** **Agent loop:** call model → execute tools → feed results back → iterate, with a configurable step ceiling and a hard stop.
- [ ] **1.9** **Streaming:** stream provider tokens through the loop so partial output is emitted as it arrives, not at the end.
- [ ] **1.10** **Dangerous-command detection:** classify before execution, route to mobile for approval, block until answered. The classifier is allowlist-oriented, not a regex blocklist — see [`threat-model.md`](threat-model.md#t4--command-execution-escape).
- [ ] **1.11** **Tier 0 gate:** run format/lint/type-check after every edit; feed failures back into the loop. Deterministic, no token cost ([ADR-008](decisions.md#adr-008--loop-engineering-tier-0-ships-in-the-mvp)).

### Model router

- [ ] **1.12** **Hot-swap:** switch provider/model mid-session without losing context.
- [ ] **1.13** **Context normalisation:** translate history — including tool calls and results — into the target model's expected shape.
- [ ] **1.14** Token accounting and per-provider cost estimation.
- [ ] **1.15** Rate-limit (429) detection, exponential backoff with jitter, per-provider.
- [ ] **1.16** Context-overflow detection with summarisation by a cheaper model before handing to a smaller-context target.

### Daemon & commands

- [ ] **1.17** Daemon with IPC (Unix socket / named pipe) for CLI ↔ daemon.
- [ ] **1.18** `rla status` — daemon, relay, paired devices, active model.
- [ ] **1.19** `rla connect` — pairing token + QR + `remlinkagent://` deep link.
- [ ] **1.20** `rla device list|revoke`.
- [ ] **1.21** Config at `~/.rla/config.yaml`: providers, active model, working-directory allowlist, daemon socket.
- [ ] **1.22** Unit tests: tool dispatch, agent loop (mocked provider), context normalisation, keyring, backoff, sandbox escape attempts.

**Done when**

- `rla login` stores three keys in the keychain — verified by grepping the disk for them.
- The agent completes a real multi-step task in a scratch repository.
- Mid-task model swap preserves context, including pending tool calls.
- A dangerous command halts and waits rather than executing.
- Tier 0 catches a deliberately introduced lint error and the agent fixes it.
- A filesystem tool cannot escape the configured working directory (tested adversarially).

---

## P2 — Server backend (WSS + NATS)

**Goal:** a relay that cannot read what it carries.
**Estimate:** 3 weeks · **Depends on:** P0

**Technical decisions**

- One durable JetStream stream per session; per-consumer `lastSeq` acknowledgement.
- Pairing token: single-use, ~60 s TTL, HMAC-signed.
- The relay never holds a decryption key. This constrains every feature that follows.

- [ ] **2.1** HTTP/WSS skeleton ([`gorilla/websocket`](https://github.com/gorilla/websocket)); TLS termination.
- [ ] **2.2** SQLite migrations: `devices`, `sessions`, `session_events`, `pairing_tokens`, `push_tokens`, `audit_log`.
- [ ] **2.3** **Wire protocol v1** implemented to [`protocol.md`](protocol.md), including the version handshake and an explicit incompatibility error.
- [ ] **2.4** **QR pairing:** single-use token, challenge-response, **public-key exchange for E2E**.
- [ ] **2.5** Device registration, per-device tokens, revocation.
- [ ] **2.6** **Multi-device:** several phones and several CLI hosts under one account; explicit routing.
- [ ] **2.7** **E2E encryption:** relay stores and forwards sealed envelopes. Verified by a test asserting the relay cannot produce plaintext ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)).
- [ ] **2.8** JetStream: durable stream per session, sequenced events, ack-all.
- [ ] **2.9** **Retention policy:** per-stream age and size limits, disk ceilings, documented in [`architecture.md`](architecture.md#retention). A self-hoster must not silently run out of disk.
- [ ] **2.10** **SYNC:** `lastSeq` → incremental replay, gap detection, resume.
- [ ] **2.11** **Streaming relay:** partial-token deltas forwarded with low latency; deltas are ephemeral, final messages persist. Naive per-token persistence would bloat the stream.
- [ ] **2.12** Session lifecycle: create / pause / resume / close.
- [ ] **2.13** **Push gateway:** APNs + FCM. Bodies carry event type and id only — content is rendered on-device after decryption.
- [ ] **2.14** **Approval binding:** every approval carries device id + command hash + nonce; replay and cross-command reuse rejected.
- [ ] **2.15** Rate limiting; WSS heartbeat / ping-pong / reconnect backoff.
- [ ] **2.16** **Observability:** structured logs, `/metrics` (connections, stream lag, error rate, push failures), `/healthz`, `/readyz`. Enough to notice an outage before a user reports it.
- [ ] **2.17** Dependency CVE scanning in CI.
- [ ] **2.18** Tests: unit + integration (testcontainers: NATS + SQLite), replay, gap detection, approval replay attempts.

**Done when**

- CLI pairs a device by QR and publishes events over WSS.
- The phone reconnects after being offline and replays from `lastSeq` with nothing missing.
- A test proves the relay never sees plaintext.
- A test proves AI traffic never traverses the relay (proxy assertion).
- A replayed approval is rejected.
- Retention limits are enforced and observable.

---

## P3 — Flutter mobile app

**Goal:** the phone becomes a usable control surface for the agent.
**Estimate:** 5–6 weeks · **Depends on:** P2

**Technical decisions:** Riverpod · `go_router` · `drift` offline · Material 3 light/dark.

- [ ] **3.1 — Skeleton**
  - [ ] **3.1.1** Feature-first layout: `features/<feature>/{presentation,domain,data}`.
  - [ ] **3.1.2** `ProviderScope` + router with a pairing guard.
  - [ ] **3.1.3** Material 3 theme, light + dark.
  - [ ] **3.1.4** L10n (`intl` + `gen-l10n`), TR + EN `.arb`. **No hard-coded user-facing strings.**
  - [ ] **3.1.5** `core/errors`: `Failure` type and error → message mapping.

- [ ] **3.2 — Network & sync**
  - [ ] **3.2.1** WSS wrapper: connect, heartbeat, reconnect with backoff.
  - [ ] **3.2.2** **Crypto layer:** key storage, envelope seal/open, matching Go byte for byte.
  - [ ] **3.2.3** Event deserialisation (freezed / json_serializable) + protocol-version check with a clear upgrade message.
  - [ ] **3.2.4** SYNC client: `lastSeq` → incremental replay on open.
  - [ ] **3.2.5** `drift`: `lastSeq`, offline queue, message cache.
  - [ ] **3.2.6** Connectivity state + UI indicator.
  - [ ] **3.2.7** **CLI-unreachable state** ([ADR-009](decisions.md#adr-009--the-cli-host-must-be-reachable)): show last known status and say the host is unreachable. Never an infinite spinner.

- [ ] **3.3 — Storage & auth**
  - [ ] **3.3.1** `flutter_secure_storage` for device token and session keys.
  - [ ] **3.3.2** Route to pairing when unpaired.

- [ ] **3.4 — Pairing**
  - [ ] **3.4.1** QR scanning (`mobile_scanner`) + camera permission flow.
  - [ ] **3.4.2** Deep link `remlinkagent://pair?token=…`.
  - [ ] **3.4.3** Handshake, key exchange, device registration.

- [ ] **3.5 — Chat**
  - [ ] **3.5.1** Message list with streaming partial rendering.
  - [ ] **3.5.2** Composer → WSS event → daemon.
  - [ ] **3.5.3** Markdown + syntax highlighting.
  - [ ] **3.5.4** **Tool-call display:** which files were read/written, which commands ran, with output.
  - [ ] **3.5.5** Message states: sending / running / done / failed.

- [ ] **3.6 — Model switcher** *(the differentiator — give it the polish)*
  - [ ] **3.6.1** Provider/model dropdown with an active-model badge.
  - [ ] **3.6.2** Hot-swap event, with an explicit "context preserved" indicator.
  - [ ] **3.6.3** Model metadata: context limit, cost badge, token usage so far.

- [ ] **3.7 — Approvals**
  - [ ] **3.7.1** Push + in-app modal: full command preview, working directory, approve/reject.
  - [ ] **3.7.2** Signed approval → WSS → CLI.
  - [ ] **3.7.3** Visible expiry countdown; expired requests fail closed.

- [ ] **3.8 — Push**
  - [ ] **3.8.1** `firebase_messaging` (APNs + FCM).
  - [ ] **3.8.2** iOS APNs certificates/provisioning; Android `google-services.json`.
  - [ ] **3.8.3** Push → deep link → correct screen.
  - [ ] **3.8.4** Categories: approval, task complete, error.
  - [ ] **3.8.5** Bodies rendered locally post-decryption — no plaintext in the payload.

- [ ] **3.9 — Platform**
  - [ ] **3.9.1** iOS `Info.plist`: camera, push, deep link.
  - [ ] **3.9.2** Android `AndroidManifest.xml`: camera, internet, push, intent filter.
  - [ ] **3.9.3** Runtime permission rationale UI.

- [ ] **3.10 — Tests**
  - [ ] **3.10.1** Domain/use-case unit tests.
  - [ ] **3.10.2** Widget tests: pairing, chat, switcher, approval.
  - [ ] **3.10.3** Crypto round-trip against Go-generated fixtures.
  - [ ] **3.10.4** `integration_test` E2E: pair → chat → swap → approve (mock relay).

**Done when**

- QR pairing, then a real agent task run from the phone.
- Model swap mid-conversation preserves context.
- Background → foreground loses nothing.
- Approving a command from the phone executes it; rejecting does not.
- Light/dark and TR/EN both correct.
- Killing the CLI produces a clear unreachable state, not a hang.

---

## P4 — Release & distribution

**Goal:** get it into people's hands.
**Estimate:** 3–4 weeks · **Depends on:** P1–P3

- [ ] **4.1** Cross-compile macOS/Linux/Windows × amd64/arm64; checksums + signatures (cosign/gpg).
- [ ] **4.2** Install script at `https://remlinkagent.com/install.sh`.
- [ ] **4.3** ➕ Homebrew tap / Scoop / apt.
- [ ] **4.4** Self-host package: `docker compose` + deployment docs + reverse-proxy TLS example.
- [ ] **4.5** **Privacy policy and data inventory** ([`privacy.md`](privacy.md)) — required for store privacy labels, and the honest counterpart to the E2E claim.
- [ ] **4.6** App Store + Play Store submission: privacy labels, screenshots, review notes explaining BYOK and the relay model.
- [ ] **4.7** User docs: install, BYOK, pairing, headless operation, FAQ.
- [ ] **4.8** E2E test suite covering setup → pairing → agent task → hot-swap → approval.
- [ ] **4.9** Performance: CLI < 30 MB RSS, mobile cold start < 2 s, streaming latency budget.
- [ ] **4.10** Security review against [`threat-model.md`](threat-model.md); resolve findings before launch.
- [ ] **4.11** `CHANGELOG.md`, GitHub release, tags.
- [ ] **4.12** **Post-MVP review:** collect real usage data and decide, with community input, which of [`vision-roadmap.md`](vision-roadmap.md) comes next.

**Done when**

- One-command install works on all three platforms; binaries verify.
- Both stores approved and live.
- Threat-model review passed.
- At least one real user feedback loop is running.

---

## 4. Risks

| Risk | Impact | Likelihood | Mitigation |
| :--- | :--- | :--- | :--- |
| **Provider tool-calling incompatibility** | High | **High** | P0.17 spike before P1 design is locked; per-provider contract tests |
| **Agent loop quality** — plausible but wrong edits | High | High | Tier 0 gate; sandboxed tools; every command approved |
| Store rejection (push, permissions, BYOK model) | High | Medium | Apache-2.0 mobile ([ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile)); early guideline review; TestFlight |
| E2E crypto interop Go ↔ Dart | High | Medium | P0.19 spike; cross-language fixtures in CI |
| Cross-platform command execution | Medium | Medium | P0.18 spike; per-platform tests |
| JetStream loss or disk exhaustion | High | Low | Durable streams + ack; retention limits (2.9); replay tests |
| CLI memory growth | Medium | Medium | Memory profiling in CI; streaming rather than buffering |
| Key leakage | Critical | Low | Keychain; log redaction verified by test; no key path to the relay |
| **Scope creep** | High | **High** | Vision items stay in `vision-roadmap.md`. Adding to the MVP requires an ADR. |

> **The largest risk remains scope creep.** Terminal mirroring and Loop Engineering are each large enough to consume the whole schedule. They are out, and they stay out until the MVP ships.

---

## M1 — Managed Cloud, subscriptions & Team

> **Why this is here and not in the vision doc.** Managed cloud is not a product feature — it is the mechanism that funds the work. Self-hosting stays free and at full parity; the hosted tier sells convenience: uptime, push certificates, upgrades. Because it pays for everything else, it comes before the nice-to-haves.

**Prerequisite:** P0–P4 shipped. **Estimate:** 5–6 weeks

**Settled:** AGPL-3.0-or-later core · Pro $5/mo · Team $15/mo for 5 seats ("five for the price of three") · own subscription site as the primary channel, in-app purchase secondary.

- [ ] **M1.1** Managed relay: multi-tenant, isolated; WSS gateway + JetStream + push.
- [ ] **M1.2** Accounts: registration, login, device management, plan assignment.
- [ ] **M1.3** Billing: Stripe; Pro and Team; renewal, cancellation, proration; webhook → plan sync.
- [ ] **M1.4** **Subscription website** (`website/`)
  - [ ] **M1.4.1** Landing: value, pricing, self-host vs managed.
  - [ ] **M1.4.2** Account area: plan, renewal, Team seat invitations, invoices.
  - [ ] **M1.4.3** Checkout (Stripe), coupons.
  - [ ] **M1.4.4** Pairing integration from the web.
  - [ ] **M1.4.5** Docs/blog — static (Astro/Next), SEO.
  - [ ] **M1.4.6** Status page, contact, GDPR/KVKK privacy.
- [ ] **M1.5** Plan gating: Pro (3 projects, 7-day retention, push); Team (5 seats, unlimited retention, audit log). **Self-host: full parity, always.**
- [ ] **M1.6** Team: seat invitations, shared projects, roles, audit log.
- [ ] **M1.7** ➕ In-app purchase: StoreKit 2 + Play Billing with server-side receipt validation.
- [ ] **M1.8** Cost transparency: per-provider token reporting, export (JSON/CSV/PDF).
- [ ] **M1.9** Self-host ↔ cloud migration: config export/import.
- [ ] **M1.10** Operations: DDoS protection, rate limiting, monitoring, backups, AGPL compliance audit.

**Done when**

- Pro/Team purchasable; Stripe billing works.
- Subscription → pairing → CLI/mobile works end to end.
- Team invitations and seat management work.
- A test proves AI traffic does not traverse the managed relay.
- E2E encryption holds on managed exactly as on self-hosted.
- Migration in both directions works.

---

## 5. Beyond M1

Once managed cloud ships, [`vision-roadmap.md`](vision-roadmap.md) covers:

- 🤖 **X1** Role delegation (Coder/Reviewer/Architect) + multi-project
- 🎙️ **X2** Voice control over AirPods and BLE headsets
- ⌚ **X3/X4** Apple Watch and Wear OS
- 🎧 **X5** Interactive terminal mirroring (PTY)
- 🔄 **X6** Loop Engineering tiers 1–4

Prioritised by real metrics, not by this ordering.

---

## 6. Glossary

**BYOK** — Bring Your Own Key. **Hot-swap** — changing model mid-conversation with context intact. **Zero-Touch AI** — AI traffic never traverses the relay. **E2E** — end-to-end encryption; the relay handles ciphertext only. **Tier 0** — deterministic post-edit verification with no token cost. **Agent loop** — model → tool → result → model, iterated. **`lastSeq`** — the last acknowledged JetStream sequence number, the basis of lossless replay.

---

*MVP (P0–P4) and M1. Everything beyond: [`vision-roadmap.md`](vision-roadmap.md). Why things are the way they are: [`decisions.md`](decisions.md).*
