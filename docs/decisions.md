# 📌 Architecture Decision Log

> **Status:** Living document · **Last updated:** 2026-07-27
> Every entry here is **locked**. Changing one requires a new entry that supersedes it — never edit history in place.

This log exists because several early decisions (naming, licensing, scope) are *expensive to reverse once code and contributors exist*. They are recorded here so that neither the maintainer nor a future contributor has to re-litigate them.

| # | Decision | Status | Reversible? |
| :-- | :--- | :--- | :--- |
| [ADR-001](#adr-001--product-name-and-command-surface) | Product name `RemLinkAgent`, CLI `rla` | Locked | ❌ Hard (user-visible paths) |
| [ADR-002](#adr-002--split-licensing-agpl-core-apache-20-mobile) | Split licensing: AGPL core, Apache-2.0 mobile | Locked | ❌ Hard (needs all contributors) |
| [ADR-003](#adr-003--cla-required) | CLA required for all contributions | Locked | ⚠️ Only loosenable, never tightenable |
| [ADR-004](#adr-004--end-to-end-encryption-of-relay-payloads) | End-to-end encryption of relay payloads | Locked | ⚠️ Medium (protocol change) |
| [ADR-005](#adr-005--mvp-is-an-agent-not-a-terminal-mirror) | MVP is an agent, not a terminal mirror | Locked | ✅ Easy (additive later) |
| [ADR-006](#adr-006--provider-neutral-core-first-class-zaiqwenkimi) | Provider-neutral core, first-class Z.AI/Qwen/Kimi | Locked | ✅ Easy |
| [ADR-007](#adr-007--phase-naming-scheme) | Phase naming `P0–P4` / `M1` / `X1–X6` | Locked | ✅ Easy |
| [ADR-008](#adr-008--loop-engineering-tier-0-ships-in-the-mvp) | Loop Engineering Tier 0 ships in the MVP | Locked | ✅ Easy |
| [ADR-009](#adr-009--the-cli-host-must-be-reachable) | The CLI host must be reachable — stated, not hidden | Locked | ✅ Easy |
| [ADR-010](#adr-010--public-docs-in-english-internal-prd-stays-private) | Public docs in English; internal PRD stays private | Locked | ✅ Easy |

---

## ADR-001 — Product name and command surface

**Decision.** The product is **RemLinkAgent**. Every user-visible surface derives from it:

| Surface | Value |
| :--- | :--- |
| CLI binary | `rla` |
| Server binary | `rla-server` |
| Go module | `github.com/burakhalefoglu/RemLinkAgent` |
| User config directory | `~/.rla/` |
| Project-local directory | `.rla/` |
| Deep-link scheme | `remlinkagent://` |
| Mobile application id | `com.remlinkagent.app` |
| Domain | `remlinkagent.com` |

**Context.** The project was drafted under the working name *RemCode*, which survived into the internal PRD, the roadmap, the repo layout, the config path, the deep link and the install domain — while the public README, landing page and contributing guide already said *RemLinkAgent*.

**Why it had to be settled before P0.** `~/.rla/` lands in the user's home directory and `.rla/` lands inside *the user's own git repositories*. Renaming either after release requires a migration path for every installation. The deep-link scheme is registered in `Info.plist` / `AndroidManifest.xml` and, once shipped to an app store, is effectively permanent.

**Consequences.** All `RemCode`/`remcode` references were rewritten. The historical PRD keeps its original wording and carries a banner marking it as frozen.

---

## ADR-002 — Split licensing: AGPL core, Apache-2.0 mobile

**Decision.**

| Component | License |
| :--- | :--- |
| `cmd/`, `internal/`, `deploy/` (CLI + server) | **AGPL-3.0-or-later** |
| `mobile/` (Flutter client) | **Apache-2.0** |

**Context.** The FSF states that GPL-family terms conflict with the Apple App Store's usage restrictions, and GPL applications have been removed from the store over exactly this conflict. Shipping an AGPL Flutter client to iOS would put the project in a position it cannot defend.

**Why this split and not the alternatives.** The copyleft protection that actually matters here is on the *relay server*: the risk being defended against is someone taking the server, modifying it and running a closed managed service. AGPL closes that. The mobile client carries no such risk — it is a client for a protocol. Granting an App Store exception under AGPL §7 would also have worked, but it produces a licence that no one recognises at a glance; Apache-2.0 is understood immediately and is unambiguously store-compatible.

**Compatibility.** Apache-2.0 is one-way compatible with AGPL-3.0, so mobile code may be incorporated into the core, but not the reverse. Any code intended for both sides must originate in `mobile/`. Shared wire types are **not** shared as code — both sides implement [`protocol.md`](protocol.md) independently.

**Timing.** Locked before the first third-party contribution. After that point, relicensing requires the agreement of every copyright holder.

---

## ADR-003 — CLA required

**Decision.** All contributions require a signed [Contributor License Agreement](../CLA.md), automated via CLA Assistant.

**Context.** Two project documents contradicted each other: the contributing guide said "no dual-licensing plans, therefore no CLA", while the licence standard and the vision risk table both listed a future commercial dual-licence as a mitigation for enterprise adoption.

Both cannot be true. **Dual-licensing is impossible unless one party holds the rights to relicense**, and a permission that does not exist at contribution time cannot be obtained retroactively at any reasonable cost. The claim in the old licence document that "AGPL keeps this door open (MIT/GPL cannot)" was simply wrong: the door is held open by copyright aggregation, not by licence choice.

**Decision rationale.** The commercial option was listed twice as desirable, and the revenue model already states that managed-cloud revenue belongs to the maintainer. The consistent choice is to preserve the option. A CLA imposes a one-click cost on contributors; losing the option imposes an unrecoverable cost on the project.

**Consequences.** The "no CLA" promise in the contributing guide is withdrawn — the reason is stated openly rather than quietly dropped. The CLA grants a relicensing right; it does **not** transfer copyright. Contributors keep ownership of their work.

> ⚠️ `CLA.md` is drafted from the common industry pattern and has **not** been reviewed by a lawyer. Have counsel review it before relying on it commercially.

---

## ADR-004 — End-to-end encryption of relay payloads

**Decision.** Session payloads are encrypted on the device and on the CLI host. The relay stores and forwards **opaque ciphertext plus routing metadata**. Ships in the MVP ([P2](roadmap.md#p2--server-backend-wss--nats)), not deferred.

**Context.** "Zero-Touch AI" is true of AI API traffic — that flows directly from the CLI host to the provider. But chat messages, model output, proposed shell commands and their results all pass through the relay and are *persisted* in JetStream for the retention window. On managed cloud, that is the maintainer's server.

The public claim "we never touch your traffic" would therefore have been read by users as a stronger promise than the architecture delivered. In a project whose core pitch is "the licence lets you audit us", the first competent reader finds that gap, and the credibility loss is out of proportion to the gap itself.

**Why implement rather than reword.** Key exchange already happens at pairing, so the marginal cost is small and it converts a liability into the strongest claim the project has: *a compromised relay — ours or anyone's — discloses nothing but timing and size.* No comparable tool offers this.

**Consequences.** Push notification bodies cannot contain message content; they carry an event type and an identifier, and the client renders the text locally after decryption. Server-side search and any future web client are constrained by this and must be designed around it. What the relay still observes — connection times, message sizes, device identifiers — is documented in [`privacy.md`](privacy.md) rather than hand-waved.

---

## ADR-005 — MVP is an agent, not a terminal mirror

**Decision.** The MVP ships an **AI coding agent** with tool-based command execution and captured output. Full interactive terminal mirroring — PTY allocation, ANSI escape handling, resize, a mobile terminal emulator — is **out of MVP scope** and moves to [X5](vision-roadmap.md#x5--interactive-terminal-mirroring-pty).

**Context.** The plan promised both. They are separate products with separate hard problems, and the project's own risk table already rated cross-platform PTY as its highest risk on both impact *and* likelihood.

**Rationale.** Every differentiator — BYOK, model hot-swap, first-class Z.AI/Qwen/Kimi — lives in the agent. Terminal mirroring is a solved problem with good existing answers (Tailscale + tmux, Termius, VS Code Remote) and a poor experience on a phone screen regardless of implementation quality. Carrying the hardest engineering risk in the project to deliver its weakest differentiator is the wrong trade.

**Consequences.** Non-interactive command execution with captured stdout/stderr replaces PTY work — a substantially smaller problem. The README no longer claims the terminal is "mirrored and controllable". Roughly 3–4 weeks come out of the critical path.

---

## ADR-006 — Provider-neutral core, first-class Z.AI/Qwen/Kimi

**Decision.** Any OpenAI-compatible endpoint is supported. Z.AI, Qwen and Kimi get first-class treatment: presets, tested tool-calling adapters, cost metadata.

**Context.** The product was framed as *the* client for the Chinese AI ecosystem. But an OpenAI-compatible wrapper already reaches DeepSeek, OpenRouter, Together, a local Ollama, and OpenAI itself. The restriction was a positioning choice, not an architectural one — and one that costs adoption in markets with data-residency or procurement constraints, for no engineering saving.

**Consequences.** Providers are configuration, not code paths. Adding one is a config entry plus a contract test. The narrow framing is dropped from the public copy while the cost-conscious audience stays addressed.

---

## ADR-007 — Phase naming scheme

**Decision.**

| Prefix | Meaning |
| :--- | :--- |
| `P0`–`P4` | MVP phases — [`roadmap.md`](roadmap.md) |
| `M1` | Managed Cloud + subscriptions — [`roadmap.md`](roadmap.md#m1--managed-cloud-subscriptions--team) |
| `X1`–`X6` | Vision phases — [`vision-roadmap.md`](vision-roadmap.md) |

**Context.** "V1" previously meant three different things — Loop Engineering in the PRD, Managed Cloud in the roadmap, Multi-Model Orchestration in the vision doc — and the vision doc's item numbers (`V1.1`, `V1.2`…) collided one-for-one with the Managed Cloud items. Unambiguous prefixes cost nothing and remove a whole class of misunderstanding.

---

## ADR-008 — Loop Engineering Tier 0 ships in the MVP

**Decision.** Tier 0 (format, lint, type/compile check after every agent edit — no AI cost) ships in the MVP as [P1.12](roadmap.md#p1--cli-agent). Tiers 1–4 remain in [X6](vision-roadmap.md#x6--loop-engineering-tiers-14).

**Context.** Loop Engineering was deferred wholesale, correctly: tiers 1–4 rest on LLM judgement, have no determinism, and the gap between a working demo and something trustworthy is widest there.

**Rationale.** Tier 0 is not in that category. It is deterministic, costs no tokens, and the agent loop that invokes it is being built anyway. It is close to free, and it is the difference between "another chat interface" and a tool that verifies its own edits.

---

## ADR-009 — The CLI host must be reachable

**Decision.** The requirement that the user's machine be powered on with the daemon running is stated plainly in the README, given first-class failure UX, and answered with a documented headless deployment path.

**Context.** BYOK plus a CLI-resident agent means there is no cloud fallback: if the machine is asleep, the product does not work. This follows from the architecture and is not a defect — but it was absent from the documentation, so a user arriving via "your terminal, in your pocket" would discover it at the worst possible moment.

**Consequences.** The client shows an explicit unreachable state with last-known status rather than a generic error. [`architecture.md`](architecture.md#cli-host-availability) documents running the daemon on an always-on host. Wake-on-LAN is noted as a future convenience, not a promise.

---

## ADR-010 — Public docs in English; internal PRD stays private

**Decision.** Everything under `docs/` is public and written in English. `documentation/` stays git-ignored for internal material only.

**Context.** The roadmap, vision, Loop Engineering framework and licence standard were private, yet the public contributing guide linked to all four — so every one of those links was broken for anyone who was not the maintainer. Meanwhile the public surface was English and the private docs Turkish.

**Rationale.** A public roadmap earns trust and attracts contributors, and pricing was already public in the README, so almost nothing was actually being protected. Publishing also gives these documents version control and a backup, which they previously lacked entirely.

**Consequences.** The frozen PRD stays private; the originals are archived under `documentation/_archive/`. UI strings are localised TR + EN via i18n keys — user-facing text is never hard-coded in either language.

> ⚠️ `documentation/` is git-ignored and therefore **has no backup**. Keep a copy outside this working tree.
