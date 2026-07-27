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
| [ADR-005](#adr-005--mvp-is-an-agent-not-a-terminal-mirror) | MVP is an agent, not a terminal mirror | **Superseded by [ADR-013](#adr-013--the-product-is-cross-verification-not-an-agent)** | — |
| [ADR-006](#adr-006--provider-neutral-core-first-class-zaiqwenkimi) | Provider-neutral core, first-class Z.AI/Qwen/Kimi | Locked | ✅ Easy |
| [ADR-007](#adr-007--phase-naming-scheme) | Phase naming `P0–P4` / `M1` / `X1–X4` | Locked | ✅ Easy |
| [ADR-008](#adr-008--loop-engineering-tier-0-ships-in-the-mvp) | Loop Engineering Tier 0 ships in the MVP | **Superseded by [ADR-013](#adr-013--the-product-is-cross-verification-not-an-agent)** — all tiers move to the core | — |
| [ADR-009](#adr-009--the-cli-host-must-be-reachable) | The CLI host must be reachable — stated, not hidden | Locked | ✅ Easy |
| [ADR-010](#adr-010--public-docs-in-english-internal-prd-stays-private) | Public docs in English; internal PRD stays private | Locked | ✅ Easy |
| [ADR-011](#adr-011--the-project-is-built-with-its-own-loop) | The project is built with its own Loop Engineering method | Locked | ✅ Easy |
| [ADR-012](#adr-012--subscription-access-goes-through-listed-agents-never-through-us) | Subscription access goes through listed agents, never through us | Locked | ⚠️ Medium |
| [ADR-013](#adr-013--the-product-is-cross-verification-not-an-agent) | The product is cross-verification, not an agent — **supersedes ADR-005, ADR-008** | Locked | ❌ Hard (thesis) |
| [ADR-014](#adr-014--orchestrate-acp-agents-do-not-build-one) | Orchestrate ACP agents; do not build one | Locked | ⚠️ Medium |

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

**Decision.** Session payloads are encrypted on the device and on the CLI host. The relay stores and forwards **opaque ciphertext plus routing metadata**. Ships in the MVP ([P2](roadmap.md#p2--relay-wss--nats)), not deferred.

**Context.** "Zero-Touch AI" is true of AI API traffic — that flows directly from the CLI host to the provider. But chat messages, model output, proposed shell commands and their results all pass through the relay and are *persisted* in JetStream for the retention window. On managed cloud, that is the maintainer's server.

The public claim "we never touch your traffic" would therefore have been read by users as a stronger promise than the architecture delivered. In a project whose core pitch is "the licence lets you audit us", the first competent reader finds that gap, and the credibility loss is out of proportion to the gap itself.

**Why implement rather than reword.** Key exchange already happens at pairing, so the marginal cost is small and it converts a liability into the strongest claim the project has: *a compromised relay — ours or anyone's — discloses nothing but timing and size.* No comparable tool offers this.

**Consequences.** Push notification bodies cannot contain message content; they carry an event type and an identifier, and the client renders the text locally after decryption. Server-side search and any future web client are constrained by this and must be designed around it. What the relay still observes — connection times, message sizes, device identifiers — is documented in [`privacy.md`](privacy.md) rather than hand-waved.

---

## ADR-005 — MVP is an agent, not a terminal mirror

**Decision.** The MVP ships an **AI coding agent** with tool-based command execution and captured output. Full interactive terminal mirroring — PTY allocation, ANSI escape handling, resize, a mobile terminal emulator — is **out of MVP scope** and moves to [X5](vision-roadmap.md#x4--interactive-terminal-mirroring).

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

**Decision.** Tier 0 (format, lint, type/compile check after every agent edit — no AI cost) ships in the MVP as [P1.12](roadmap.md#p1--orchestrator). Tiers 1–4 remain in [X6](vision-roadmap.md#x1--deeper-verification-tiers).

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

---

## ADR-011 — The project is built with its own loop

**Decision.** RemLinkAgent is developed using the Loop Engineering method it will later ship as a feature. The deterministic parts run today via `go run ./scripts/gate`; the parts needing judgement are declared as owed rather than skipped. Working method: [`development-loop.md`](development-loop.md). Constitution: [`.rla/PRINCIPLES.md`](../.rla/PRINCIPLES.md) and [`.rla/SECURITY-BASELINE.md`](../.rla/SECURITY-BASELINE.md).

**Context.** [`loop-engineering.md`](loop-engineering.md) describes a tiered pipeline that makes an AI demonstrate its work matches an agreed spec. As a product feature it is deferred to [X6](vision-roadmap.md#x1--deeper-verification-tiers) because tiers 1–4 depend on LLM judgement and are the riskiest thing in the plan. But the *deterministic* half — tiered gates, spec artifacts with human ratification, content-hash caching, canaries, fake-green hunting, fail-loud — needs no LLM at all.

**Rationale.** Shipping a tool that makes AI verify its own work while building it some other way would be a poor advertisement, and more practically it would mean discovering the method's rough edges after users do rather than before. Two things follow that a document alone could not deliver:

The flagship invariant becomes **structural**. "AI traffic never reaches the relay" is enforced as an import-graph rule: relay packages cannot import provider, agent, tool or crypto code. A promise in a document can be broken by accident; an import graph cannot.

And the gates get **audited**. `gate canary` plants deliberate breakage and requires each gate to catch it. Thirteen of sixteen do. The other four wrap external tools and say so — which is the honest position, not a gap being hidden.

**Consequences.** `.rla/` exists in this repository with the same layout the product will later create in a user's project, so the format is exercised before it is shipped. `COULD NOT VERIFY` is a first-class status distinct from both pass and fail, and blocks readiness. The coverage floor is a committed ratchet. CI runs the canary before the gates, because a gate that cannot detect breakage makes every later green meaningless.

Running it immediately paid for itself: the first full `verify` surfaced nine stdlib CVEs in the declared Go version, three false-positive spec citations coming from test fixtures, and a race gate silently unable to run on Windows. Each is now fixed or reported honestly.

**What this is not.** Tiers 1–4 of the *product* remain deferred. The judgement gates here are run by a human with an agent, not by the pipeline. That distinction is printed on every `gate verify` run so it cannot quietly erode.

---

## ADR-012 — Subscription access goes through listed agents, never through us

**Decision.** RemLinkAgent never calls a provider's subscription endpoint. Two supported paths:

| Path | How | Economics |
| :--- | :--- | :--- |
| **Direct** | Our code calls an OpenAI-compatible endpoint with the user's own pay-as-you-go key | PAYG rates |
| **Delegated** | We drive a provider-listed agent (Qwen Code, Kimi Code, OpenCode…) over ACP; that agent calls the provider under its own identity with the user's own subscription | Flat subscription |

**Context.** Research on 2026-07-27 against the three providers' own terms and documentation found:

- **Z.AI** — "The GLM Coding Plan is limited to use within the following officially supported tools and product environments; users may not use their subscription benefits for tools or scenarios outside of this scope." Enforcement is documented: throttle → suspension → permanent ban on the third violation.
- **Qwen / Alibaba** — supports "any third-party programming tool compatible with OpenAI or Anthropic API protocols", but excludes "custom applications: automated scripts, application backends calling the API directly". A dedicated error exists: **"Only available for Coding Agents"** — the server recognises and rejects unsupported clients.
- **Kimi** — "Kimi Code subscriptions are for interactive use only", plus an explicit routing instruction for our exact situation: *"If you need to invoke large model capabilities in your own product, visit the Kimi Platform."* Altering client identity is prohibited by name.

**Why the direct path cannot use subscriptions.** Our own client is, by definition, a third-party integration that appears on no provider's list. The only technical route to making it work is to present the identity of a listed tool — which Kimi prohibits explicitly, and which this project ruled out in its own research brief before the question arose. There is no application process to find: the providers' documented answer to product developers is "use the pay-as-you-go platform".

**Why the delegated path is different in kind.** The call is made by a listed tool, under its own real identity, with a key we never see. No impersonation, no misrepresentation. Both `qwen serve` and `kimi acp` are **first-party, documented external-control interfaces** — using a published interface as published is not circumvention. The user's subscription is consumed by the tool it was licensed for.

**Consequences.** The "cheap flat rate" promise is only available on the delegated path, which is what makes [ADR-014](#adr-014--orchestrate-acp-agents-do-not-build-one) load-bearing rather than a convenience. Z.AI ships no CLI, so it is reachable only through a provider-agnostic listed agent such as OpenCode. Public copy says "your own API key", never "your own subscription", until a provider lists us.

> Provider terms change without notice. This entry is dated and must be re-verified at every release.

---

## ADR-013 — The product is cross-verification, not an agent

> **Supersedes [ADR-005](#adr-005--mvp-is-an-agent-not-a-terminal-mirror) and [ADR-008](#adr-008--loop-engineering-tier-0-ships-in-the-mvp).**

**Decision.** RemLinkAgent is a **multi-model cross-verification system**. One model implements; a different model verifies the result against a ratified spec and produces evidence; a human ratifies the plan and confirms the outcome. Loop Engineering moves from a deferred vision item to the centre of the product. The mobile client is the interface to checkpoint ②, not the product.

**Context.** The project was scoped as a mobile-controlled coding agent. Two findings ended that framing:

*It is already built, twice.* Z.AI's ZCode ships multi-model selection, multi-agent support, BYOK, QR pairing to a desktop session, phone control, and a 1.5× quota bonus for its own subscribers — a subsidy no third party can match. OpenClaw is an open-source agent runtime with ~369k stars and ~3.2M users that already dispatches sub-agents over ACP. Competing on "a nicer mobile agent" is not winnable.

*The remaining gap was measured, not guessed.* The maintainer ran the workflow by hand: a Z.AI-driven agent declared a project complete; a verification pass driven by a **different** model, against the spec and the tier gates, found missing translations, unwritten tests, coverage regressions **and real bugs**. The gap between "the model says it is done" and "the work is correct" is measurable, and a second model finds it.

**Why this is defensible where the previous thesis was not.** Neither incumbent does adversarial cross-verification. ZCode lets you *choose* a model; OpenClaw lets an agent *delegate* to sub-agents. Neither takes finished work and has a different vendor's model try to disprove it against a specification.

And a vendor structurally cannot: the honest output of cross-verification is sometimes *"our own model got this wrong."* Z.AI's 1.5× bonus exists precisely to steer users toward GLM. **Only a party that sells no model can credibly arbitrate between models** — that is the moat, and it is unavailable to every well-resourced competitor.

**Economics.** Cross-verification costs 2–3× the tokens of a single pass. At Western API prices that is prohibitive; on cheap flat-rate subscriptions it is free at the margin. The cost positioning that began as marketing turns out to be a technical prerequisite — which is why [ADR-012](#adr-012--subscription-access-goes-through-listed-agents-never-through-us) matters so much.

**Consequences.** [`loop-engineering.md`](loop-engineering.md) is promoted from X6 reference to core specification. P1 becomes the orchestrator rather than an agent loop. The relay and mobile client keep their designs but move behind the orchestrator in sequence — they address the second half of the same pain ("when I leave my desk, work stops"), not a separate product. The README claim that the project "ships its own native CLI agent" is withdrawn.

**Known risk, stated plainly.** Cross-verification rests on model judgement, and two models can share a blind spot. The maintainer's manual runs are evidence, not proof. The gates that produce *deterministic* evidence — spec fidelity, coverage, tests, lint — carry the weight where judgement cannot.

---

## ADR-014 — Orchestrate ACP agents; do not build one

**Decision.** RemLinkAgent speaks the [Agent Client Protocol](https://agentclientprotocol.com/) and drives existing agents. It implements no agent loop, no tool framework, no model router and no provider adapters.

**Context.** ACP is JSON-RPC 2.0 over stdio with `session/request_permission` — the agent asks the client to authorise sensitive operations. Roughly 28 agents implement it, including Qwen Code, Kimi Code, Gemini CLI, Codex CLI, OpenCode and Copilot CLI. Qwen additionally ships `qwen serve`: the same session over HTTP + SSE, with `POST /permission/:requestId` for approval votes, `Last-Event-ID` replay over an 8000-frame ring buffer, multi-client sessions, and TLS intended for mobile access.

**Rationale.** The permission request *is* the approval flow this product needs; it did not have to be invented. Wrapping the protocol rather than a provider means one integration reaches every ACP agent, and the provider-agnostic ones (OpenCode, Goose, Crush) reach dozens of models each — including Z.AI, which ships no CLI of its own.

It also removes the part of the plan that could not be won. An agent loop written by one person will not beat one with years of investment behind it; the orchestration and verification layer above it is unclaimed.

**Consequences.** P1 loses the agent loop, tool framework, sandbox, danger classification, model router, context normalisation and provider contract tests — most of the original phase. What remains is the ACP client, multi-agent orchestration, handoff between agents, and integration with the existing gates.

The cost is a dependency on other projects' protocols and release cadence. ACP being a shared standard with a registry — rather than one vendor's interface — is what makes that acceptable. Agents are configured by the user, installed through their own official installers; we orchestrate rather than redistribute, which keeps licensing and supply-chain surface out of our build.

**What this does not change.** [ADR-006](#adr-006--provider-neutral-core-first-class-zaiqwenkimi) still holds for the direct PAYG path, which remains as a fallback and as the way contract tests run without a second tool installed.
