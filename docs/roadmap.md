# 🗺️ RemLinkAgent — Roadmap

> **Version:** 5.0 · **Updated:** 2026-07-27 · **Status:** P0 finishing, P1 next
> **Thesis:** one model implements, a different model tries to disprove it against a ratified spec, deterministic gates produce the evidence, a human decides. See [ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent).

Covers **P0–P4** and **M1**. Deferred capability is in [`vision-roadmap.md`](vision-roadmap.md); the reasoning behind every scope choice is in [`decisions.md`](decisions.md).

**Tracking:** `[ ]` todo → `[~]` in progress → `[x]` done. `⚠️` blocked, `➕` optional.

---

## 1. What is being built

A **multi-model cross-verification system**. Roles are filled by real agents that already exist, driven over the [Agent Client Protocol](https://agentclientprotocol.com/):

| Role | Runs | Uses |
| :--- | :--- | :--- |
| **Coder** | e.g. Qwen Code | the user's Qwen subscription |
| **Reviewer** | e.g. Kimi Code | the user's Kimi subscription |
| **Explorer** ➕ | e.g. OpenCode | any provider, including those with no CLI |
| **Gates** | native, deterministic | no tokens |

**In scope**

- ✅ Spec artifacts with human ratification — checkpoint ①
- ✅ ACP orchestration of multiple agents, each with its own model and subscription
- ✅ Automatic handoff: implement → gates → cross-review → evidence
- ✅ Deterministic gates: lint, tests, coverage ratchet, spec fidelity, fake-green, canaries
- ✅ Every proposed command approved by a human before it runs
- ✅ End-to-end encrypted relay, lossless background sync
- ✅ Mobile client for checkpoint ② and approvals
- ✅ Direct PAYG path against any OpenAI-compatible endpoint, as a fallback

**Out of scope** → [`vision-roadmap.md`](vision-roadmap.md)

- ❌ Our own agent loop, tool framework or model router ([ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one))
- ❌ Calling any provider's *subscription* endpoint ourselves ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us))
- ❌ Interactive terminal mirroring → X4
- ❌ Voice, smartwatches → X2, X3
- ❌ Managed cloud, teams, billing → M1

### Invariants — not negotiable

| | |
| :--- | :--- |
| ⚖️ **Vendor-neutral** | We sell no model. That is what makes arbitration credible. |
| 🔑 **Zero-Touch AI** | API keys and model traffic never reach the relay. |
| 🎭 **No impersonation** | Agents identify honestly. We never present another tool's identity. |
| 🔐 **E2E encrypted** | The relay handles ciphertext and routing metadata. Nothing else. |
| 🟢 **Fail-loud** | `COULD NOT VERIFY` is never a pass. |
| 📐 **Evidence, not assertion** | A green gate must be able to prove it can still detect breakage. |
| 🪶 **Light** | Single static binary, < 30 MB RSS; mobile cold start < 2 s. |
| 🌍 **Localised** | TR + EN, i18n keys only. |

---

## 2. Schedule

**14–18 weeks for one developer working full time.**

| Phase | Estimate |
| :--- | :--- |
| [P0](#p0--scaffolding-gates--de-risking) — Scaffolding, gates, de-risking | 1.5–2 weeks |
| [P1](#p1--orchestrator) — Orchestrator | 3–4 weeks |
| [P2](#p2--relay-wss--nats) — Relay | 3 weeks |
| [P3](#p3--mobile-client) — Mobile client | 4–5 weeks |
| [P4](#p4--release--distribution) — Release | 3–4 weeks |
| [M1](#m1--managed-cloud-subscriptions--team) — Managed cloud | 5–6 weeks (after MVP) |

Down from 17–20 in v4.0. [ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one) removed the agent loop, tool framework, sandbox, danger classification, model router, context normalisation and provider contract tests — most of the old P1.

---

## 3. Stack

| Layer | Choice | Notes |
| :--- | :--- | :--- |
| Orchestrator | **Go** | Static binary; macOS/Linux/Windows, amd64 + arm64 |
| Agent transport | **ACP** (JSON-RPC 2.0 over stdio) | Plus `qwen serve` HTTP+SSE where available |
| Relay | **Go** | `net/http` + `gorilla/websocket` |
| Queue | **NATS JetStream** | Persistent, replayable |
| Database | **SQLite** | Zero-config |
| Mobile | **Flutter** | iOS + Android |
| Crypto | **`nacl/box`** + `cryptography` (Dart) | X25519 + XSalsa20-Poly1305 |
| Direct path ➕ | **`sashabaranov/go-openai`** | PAYG fallback ([ADR-006](decisions.md#adr-006--provider-neutral-core-first-class-zaiqwenkimi)) |
| Licence | **AGPL-3.0-or-later** core, **Apache-2.0** mobile | [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile) |

Detail: [`architecture.md`](architecture.md) · Wire formats: [`protocol.md`](protocol.md)

---

## P0 — Scaffolding, gates & de-risking

**Goal:** a repository that verifies itself, and one answered question that decides P1's shape.
**Estimate:** 1.5–2 weeks

### Scaffolding

- [x] **0.1** Monorepo layout; Go module; `cmd/rla`, `cmd/rla-server`.
- [x] **0.2** `Makefile` + `scripts/make.ps1`; `golangci-lint`; CI on Linux/macOS/Windows.
- [x] **0.3** Governance: `CONTRIBUTING`, `SECURITY`, `CHANGELOG`, `CLA`, templates, `CODEOWNERS`.
- [x] **0.4** Dual licence headers, boundary guard, `license-check` CI.
- [x] **0.5** ADR log ([`decisions.md`](decisions.md)).
- [x] **0.6** `deploy/` — NATS JetStream + relay compose.
- [x] **0.7** Documentation link/anchor checker in CI.
- [ ] **0.8** Cobra + Viper, replacing the skeleton dispatch.
- [ ] **0.9** SQLite migrations ([`goose`](https://github.com/pressly/goose)).
- [ ] **0.10** Flutter bootstrap: `flutter create --org com.remlinkagent --project-name rla_mobile --platforms=ios,android mobile`.

### Verification system — the product's core, running on itself

- [x] **0.11** Tiered gates 0–3 with selective regression cache ([`gate`](../scripts/gate)).
- [x] **0.12** Spec artifacts, `SPEC-…` ids, forward fidelity diff, ratification checkpoint.
- [x] **0.13** Canaries — every gate proves it detects planted breakage.
- [x] **0.14** Fake-green detection; `COULD NOT VERIFY` as a first-class status.
- [x] **0.15** Coverage ratchet; structural invariant enforcement via the import graph.
- [ ] **0.16** Backward fidelity diff — behaviour with no covering requirement. Judgement-based; design only in P0.

### De-risking spike

- [ ] **0.17** ⚠️ **ACP capability matrix.** For Qwen Code, Kimi Code and OpenCode, verify with real sessions:
  - [ ] **0.17.1** `session/request_permission` — payload shape, what is included about the proposed action, whether it can be deferred long enough for a phone round trip.
  - [ ] **0.17.2** Session lifecycle: create, resume, cancel; what survives an agent restart.
  - [ ] **0.17.3** How a completed diff is retrieved for handoff to the reviewer.
  - [ ] **0.17.4** Whether one orchestrator can hold several agent sessions concurrently without interference.
  - [ ] **0.17.5** `qwen serve` HTTP+SSE specifics: `Last-Event-ID` replay, multi-client voting, TLS.
  - [ ] **0.17.6** Results into [`protocol.md`](protocol.md#acp-capability-matrix).

  > **Why this gates P1.** The whole design assumes ACP exposes approvals to an external client and that agents can be driven concurrently. If permission requests cannot survive a phone round trip, the approval flow needs a different mechanism — and that changes P1 and P3 together.

**Done when**

- `make cli server` builds; `make verify` and `make canary` are green.
- `make docker-up` brings up NATS + relay; `/healthz` answers.
- Flutter app runs on a simulator.
- The ACP matrix is written up in `protocol.md`.

---

## P1 — Orchestrator

**Goal:** automate, on one machine, the handoff the maintainer currently performs by hand.
**Estimate:** 3–4 weeks · **Depends on:** P0.17

**Technical decisions**

- Agents are external processes the user installs. We orchestrate; we never redistribute.
- Each agent keeps its own credentials. The orchestrator never sees a key.
- Roles are configuration, not code paths — any ACP agent can fill any role.

### ACP client

- [ ] **1.1** ACP client: JSON-RPC 2.0 over stdio, agent lifecycle, initialisation, capability negotiation.
- [ ] **1.2** `session/request_permission` handling — surface, hold, resolve, expire.
- [ ] **1.3** Streaming session updates; tool calls and diffs surfaced as structured events.
- [ ] **1.4** `qwen serve` HTTP+SSE transport alongside stdio, where the agent offers it.
- [ ] **1.5** Concurrent sessions across several agents without interference.
- [ ] **1.6** Agent health: crash detection, restart, and a loud failure when an agent dies mid-task.

### Roles & orchestration

- [ ] **1.7** `rla setup` — detect installed agents, install missing ones via **their own official installers**, run their login flows, verify.
- [ ] **1.8** Role configuration in `~/.rla/config.yaml`: which agent fills Coder, Reviewer, Explorer.
- [ ] **1.9** **Handoff:** implement → collect diff → run gates → hand diff + spec + gate findings to the Reviewer.
- [ ] **1.10** Reviewer prompt construction: spec, diff, gate output, and an explicit instruction to disprove rather than confirm.
- [ ] **1.11** Finding aggregation: reviewer output plus gate output into one evidence set.
- [ ] **1.12** Iteration loop with a bounded budget; hard stop with a loud report.
- [ ] **1.13** **Unified approval queue** — permission requests from every agent in one place, one decision each.

### Specs & gates

- [ ] **1.14** `rla spec new|list|ratify` — artifact lifecycle from the CLI.
- [ ] **1.15** Gate engine promoted from `scripts/gate` into `internal/verify` as a library.
- [ ] **1.16** Language-agnostic gate configuration (Go, Dart, Python, JS toolchains).
- [ ] **1.17** Evidence bundle: what ran, what passed, what could not be verified, what the reviewer found.
- [ ] **1.18** Fail-loud end to end: an agent crash or a skipped gate blocks the readiness verdict.

### Fallback path ➕

- [ ] **1.19** Direct PAYG mode against any OpenAI-compatible endpoint, keys in the OS keychain.
- [ ] **1.20** Secret redaction across every log path, verified by test.

- [ ] **1.21** Tests: ACP client against a mock agent, handoff, aggregation, approval expiry, concurrent sessions.

**Done when**

- Two different agents, two different subscriptions, run Coder and Reviewer on one task.
- A deliberately incomplete implementation is caught — by the gates, by the reviewer, or both.
- The reviewer's findings and the gate evidence arrive as one bundle.
- A draft spec blocks the readiness verdict.
- Approvals from several agents queue in one place and each resolves exactly once.
- No agent's API key is ever visible to the orchestrator (asserted by test).

---

## P2 — Relay (WSS + NATS)

**Goal:** a relay that cannot read what it carries.
**Estimate:** 3 weeks · **Depends on:** P0

Design unchanged from v4.0 — the relay is indifferent to what the orchestrator does.

- [ ] **2.1** HTTP/WSS gateway ([`gorilla/websocket`](https://github.com/gorilla/websocket)); TLS termination.
- [ ] **2.2** SQLite migrations: `devices`, `sessions`, `session_events`, `pairing_tokens`, `push_tokens`, `audit_log`.
- [ ] **2.3** **Wire protocol v1** per [`protocol.md`](protocol.md), with version handshake and explicit incompatibility errors.
- [ ] **2.4** **QR pairing:** single-use token, ~60 s TTL, HMAC-signed, challenge-response, public-key exchange.
- [ ] **2.5** Device registration, per-device tokens, revocation.
- [ ] **2.6** **Multi-host:** several phones and several orchestrator hosts under one account — the maintainer runs 3–5 projects.
- [ ] **2.7** **E2E encryption**, proven by a test asserting the relay cannot produce plaintext ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)).
- [ ] **2.8** JetStream: durable stream per session, sequenced events, ack-all.
- [ ] **2.9** **Retention policy** with documented limits ([`architecture.md`](architecture.md#sessions-streams-retention)).
- [ ] **2.10** **SYNC:** `lastSeq` → incremental replay, gap detection surfaced to the user.
- [ ] **2.11** Streaming relay: deltas ephemeral, completed messages persisted.
- [ ] **2.12** Session lifecycle: create / pause / resume / close.
- [ ] **2.13** **Push gateway** (APNs + FCM) — event type and id only; text rendered on-device after decryption.
- [ ] **2.14** **Approval binding:** device id + command hash + nonce; replay and cross-command reuse rejected.
- [ ] **2.15** Rate limiting; heartbeat; reconnect backoff.
- [ ] **2.16** **Observability:** structured logs (no payloads), `/metrics`, `/healthz`, `/readyz`.
- [ ] **2.17** Dependency CVE scanning in CI.
- [ ] **2.18** Tests: integration (testcontainers), replay, gap detection, approval replay attempts.

**Done when**

- Pairing by QR; events flow; an offline phone replays from `lastSeq` losing nothing.
- Tests prove the relay never sees plaintext and never reaches a model provider.
- A replayed approval is rejected.
- Multiple project hosts are addressable from one phone.

---

## P3 — Mobile client

**Goal:** make checkpoints ① and ② work away from the desk.
**Estimate:** 4–5 weeks · **Depends on:** P2

Riverpod · `go_router` · `drift` · Material 3 light/dark.

- [ ] **3.1 — Skeleton:** feature-first layout, router with pairing guard, theme, L10n (TR + EN, keys only), `Failure` mapping.
- [ ] **3.2 — Network & crypto:** WSS wrapper with reconnect; envelope seal/open matching Go byte for byte; protocol version check; `drift` cache; **explicit host-unreachable state** ([ADR-009](decisions.md#adr-009--the-cli-host-must-be-reachable)).
- [ ] **3.3 — Storage:** device token and session keys in Keychain/Keystore.
- [ ] **3.4 — Pairing:** QR scan, `remlinkagent://pair?token=…`, handshake, key exchange.
- [ ] **3.5 — ① Spec ratification:** read a draft spec, requirement by requirement; ratify, request changes, or reject.
- [ ] **3.6 — ② Evidence review:** what each gate reported, what the reviewer found, which requirements are covered, what could not be verified. **The screen the product exists for.**
- [ ] **3.7 — Approvals:** push + in-app modal with full command, working directory, which agent asked, visible expiry; fails closed.
- [ ] **3.8 — Run view:** which agent holds which role, current step, streamed output, cancel.
- [ ] **3.9 — Multi-project:** switch between orchestrator hosts; per-project status at a glance.
- [ ] **3.10 — Push:** APNs + FCM, categories, deep links, content-free payloads.
- [ ] **3.11 — Platform:** iOS `Info.plist`, Android manifest, permission rationale UI.
- [ ] **3.12 — Tests:** unit, widget, crypto round-trip against Go fixtures, `integration_test` E2E against a mock relay.

**Done when**

- A spec is ratified from the phone and the orchestrator proceeds.
- Evidence from a real cross-verification run is readable and actionable on a phone screen.
- Approving from the phone executes; rejecting does not.
- Background → foreground loses nothing; killing the orchestrator produces a clear unreachable state.
- Light/dark and TR/EN both correct.

---

## P4 — Release & distribution

**Goal:** get it into people's hands.
**Estimate:** 3–4 weeks · **Depends on:** P1–P3

- [ ] **4.1** Cross-compile macOS/Linux/Windows × amd64/arm64; checksums + signatures.
- [ ] **4.2** Install script at `https://remlinkagent.com/install.sh`.
- [ ] **4.3** ➕ Homebrew tap / Scoop / apt.
- [ ] **4.4** Self-host package: compose, deployment docs, reverse-proxy TLS example.
- [ ] **4.5** **Privacy policy and data inventory** ([`privacy.md`](privacy.md)) for store labels.
- [ ] **4.6** App Store + Play Store submission, with review notes explaining the BYOK and relay model.
- [ ] **4.7** User docs: install, agent setup, roles, specs, FAQ.
- [ ] **4.8** E2E suite: setup → pair → ratify → implement → verify → approve.
- [ ] **4.9** Performance: orchestrator < 30 MB RSS, mobile cold start < 2 s.
- [ ] **4.10** Security review against [`threat-model.md`](threat-model.md).
- [ ] **4.11** **Re-verify provider terms** — they change without notice ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).
- [ ] **4.12** `CHANGELOG`, GitHub release, tags.
- [ ] **4.13** Post-MVP review: real usage data decides what comes next.

---

## 4. Risks

| Risk | Impact | Likelihood | Mitigation |
| :--- | :--- | :--- | :--- |
| **Cross-verification adds no signal** — two models share a blind spot | **Critical** | Medium | Manual runs already found real defects; deterministic gates carry weight where judgement cannot; measure catch rate from day one |
| **ACP cannot hold an approval for a phone round trip** | High | Medium | P0.17.1 answers this before P1 |
| ACP protocol churn across agents | Medium | High | Shared standard with a registry, not one vendor's interface; capability negotiation; pin tested versions |
| Provider terms change | High | Medium | Delegated path only; re-verified each release; no impersonation ever |
| Agent install friction | Medium | High | `rla setup` orchestrates their official installers |
| Incumbents add cross-verification | High | Low | A vendor cannot credibly arbitrate between models ([ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent)) |
| **Scope creep** | High | High | Adding to the MVP requires an ADR |

> **The critical risk is the first one.** Everything else is engineering. If a second model does not reliably find what the first missed, no architecture saves the product — so catch rate is instrumented from P1 and reviewed at every phase gate.

---

## M1 — Managed Cloud, subscriptions & Team

> Not a feature — the mechanism that funds the work. Self-hosting stays free and complete; the hosted tier sells uptime, push certificates and upgrades.

**Prerequisite:** P0–P4 shipped. **Estimate:** 5–6 weeks
**Settled:** Pro $5/mo · Team $15/mo for 5 seats · own subscription site primary, in-app purchase secondary.

- [ ] **M1.1** Managed relay: multi-tenant, isolated.
- [ ] **M1.2** Accounts, device management, plan assignment.
- [ ] **M1.3** Stripe billing; renewal, cancellation, proration; webhooks.
- [ ] **M1.4** Subscription website: landing, account area, checkout, docs, status page, privacy.
- [ ] **M1.5** Plan gating. **Self-host: full parity, always.**
- [ ] **M1.6** Team: seats, shared specs, roles, audit log.
- [ ] **M1.7** ➕ StoreKit 2 + Play Billing with server-side receipt validation.
- [ ] **M1.8** Cost transparency: token reporting per agent and provider.
- [ ] **M1.9** Self-host ↔ cloud migration.
- [ ] **M1.10** Operations: DDoS protection, monitoring, backups, AGPL compliance audit.

---

## 5. Beyond M1

[`vision-roadmap.md`](vision-roadmap.md): deeper verification tiers (mutation, fuzzing, black-box exploration), voice control, Apple Watch and Wear OS, interactive terminal mirroring.

---

## 6. Glossary

**ACP** — Agent Client Protocol; JSON-RPC over stdio between a client and a coding agent. **Cross-verification** — a different model attempting to disprove work against a ratified spec. **Checkpoint ①** — human ratifies the plan. **Checkpoint ②** — human confirms the outcome. **Evidence bundle** — everything a run can prove. **`COULD NOT VERIFY`** — a gate did not run; never a pass. **Delegated path** — a listed agent calls the provider under its own identity. **BYOK** — the user's own key or subscription, never ours.

---

*Why things are the way they are: [`decisions.md`](decisions.md). How the work gets done: [`development-loop.md`](development-loop.md).*
