# 📌 Architecture Decision Log

> **Status:** Living document · **Last updated:** 2026-07-28
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
| [ADR-015](#adr-015--the-humans-default-state-is-non-involvement) | The human's default state is non-involvement — **sharpens [P6](../.rla/PRINCIPLES.md#p6--the-user-is-in-the-loop)** | Locked · checkpoint set superseded by [ADR-020](#adr-020--three-checkpoints-and-the-middle-one-is-conditional) | ❌ Hard (thesis) |
| [ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread) | The interface is three screens, not a chat thread | **Superseded by [ADR-021](#adr-021--three-tabs-and-an-asynchronous-command-bar)** | — |
| [ADR-017](#adr-017--cheap-models-grind-an-expensive-model-arbitrates) | Cheap models grind; an expensive model arbitrates | **Superseded by [ADR-022](#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer)** | — |
| [ADR-018](#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy) | The primitive unit is a module; intent is a hierarchy | Locked | ❌ Hard (thesis) |
| [ADR-019](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write) | Three intent gates; the module document is not writable by the loop | Locked | ⚠️ Medium |
| [ADR-020](#adr-020--three-checkpoints-and-the-middle-one-is-conditional) | Three checkpoints; ① is conditional — **supersedes [ADR-015](#adr-015--the-humans-default-state-is-non-involvement)'s checkpoint set** | Locked | ⚠️ Medium |
| [ADR-021](#adr-021--three-tabs-and-an-asynchronous-command-bar) | Three tabs and an async command bar — **supersedes [ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread)** | Locked | ⚠️ Medium |
| [ADR-022](#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer) | Tiers M/T0–T4; verifier ≠ producer — **supersedes [ADR-017](#adr-017--cheap-models-grind-an-expensive-model-arbitrates)** | Locked | ✅ Easy |
| [ADR-023](#adr-023--a-judged-gate-may-add-a-finding-never-clear-one) | Every intent gate has a deterministic floor; judgement is advisory | Locked | ⚠️ Medium |
| [ADR-024](#adr-024--silence-is-attested-not-assumed) | Silence is attested — the heartbeat is P11's second and final exception | Locked | ✅ Easy |
| [ADR-025](#adr-025--every-invariant-declares-the-criterion-it-derives-from) | Invariants declare their parent criterion — **replaces [ADR-020](#adr-020--three-checkpoints-and-the-middle-one-is-conditional)'s second escalation test** | Locked | ⚠️ Medium |
| [ADR-026](#adr-026--the-hierarchy-relocates-the-blind-spot-it-does-not-remove-it) | The hierarchy relocates the blind spot; ⓪ gets an adversarial completeness pass | Locked | ✅ Easy |
| [ADR-027](#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it) | Deterministic verdicts are exit codes; judgement reads the artifact, reports, never blocks a round | Locked | ⚠️ Medium |
| [ADR-028](#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard) | Counts are evidence; guards are the price of caching, tiering and deferral | Locked | ⚠️ Medium |

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

---

## ADR-015 — The human's default state is non-involvement

> **Sharpens [P6 — The user is in the loop](../.rla/PRINCIPLES.md#p6--the-user-is-in-the-loop).** P6 established *that* a human decides. This establishes *how often*, and that the rest of the time the answer is "not at all".

**Decision.** The human's default state is **not involved**. Intervention begins because the system crossed a threshold and called for them — never because they were watching. There are exactly three occasions on which this product may take a person's attention:

| | Occasion | Question |
| :--- | :--- | :--- |
| **①** | Spec ratification | Is the plan right? |
| **②** | Interface test | Does it actually work? |
| **③** | Arbitration | A gate went red and the loop cannot resolve it alone. |

Everything else runs unattended, and anything that is not one of the three is not permitted to interrupt.

**Context.** Two working modes are common, and both consume the same resource: a person's continuous attention.

The developer keeps their hands on the keyboard because something has to ship, and so stays at the centre of the work. The product owner delegates everything to the agent and looks only at the output. The first does not scale past one person's day. The second has no way to find out whether the output is right.

Current tooling inherits the chat window from the first mode. The agent streams its reasoning, a person reads along, and every affordance assumes an audience. The net effect is that *hours spent writing code* became *hours spent watching code be written* — the same eight hours, less agency, and no better guarantee at the end.

**Rationale.** Control rooms settled this long ago. Nobody staffs one by staring at a console for eight hours. Instrumentation watches continuously, thresholds are agreed in advance, and an alarm summons a human who arrives with authority and context. The operator is measured by the quality of their decisions, not by their uptime in front of the screen.

That maps onto the loop without forcing: the gates are the instrumentation, the ratified spec supplies the thresholds, and a red gate is the alarm. The missing piece was never technical — it was the willingness to state that the human is **off duty** the rest of the time, and then to build the interface as though that were true.

**Consequences.**

*The primitive unit of the product is the spec and the gate, not the message.* A thread has no notion of "finished" and no notion of "correct". A spec with ratified invariants and a fidelity gate has both. This is what makes an unattended default safe rather than negligent — the system is not merely quiet, it is *checking*.

*An interface with nothing on it is a success state.* Chat trained everyone to read an empty screen as "nothing happened yet". Here it means the loop is grinding and has found nothing worth your time. The interface must say so in those words ([ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread)).

*Every avoidable interrupt is a defect,* tracked as one. A tool that pages an operator about something they cannot act on teaches them to ignore the pager, and the third alarm — the one that mattered — is the one they miss.

*Anything that only makes sense while watching cannot be the default* — live logs, a ticking token counter, a reasoning stream. Available on request, never in the way ([ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread)).

**Known risk.** If the gates are weak, an unattended default is worse than watching, because it manufactures confidence rather than merely wasting time. This is why gate integrity is treated as a first-order concern — canaries, the compiled-not-configured rule, the coverage ratchet, and `COULD NOT VERIFY` as a status that blocks readiness ([ADR-011](#adr-011--the-project-is-built-with-its-own-loop)). The thesis is only as good as the alarm.

---

## ADR-016 — The interface is three screens, not a chat thread

**Decision.** The mobile client is **three screens**. A conversation is never the top-level container.

| Screen | Contents | Empty means |
| :--- | :--- | :--- |
| **Queue** | Decisions awaiting a human, nothing else | ✅ Success — the loop is running and has nothing for you |
| **Command view** | Per project: how many features, how many ratified, how many drifting from spec, how many awaiting a decision | Nothing is being worked on |
| **Feature detail** | The spec, its invariants, gate evidence, and a conversation **scoped to the decision at hand** | — |

And the load-bearing part: **a red gate must be answerable from a phone.** So the orchestrator does not ship raw evidence to the device. It produces a decision object:

| Field | Content |
| :--- | :--- |
| **Violated invariant** | Which ratified invariant, by `SPEC-…` id |
| **Why** | One sentence, in the spec's own domain language |
| **Options** | Two or three, each with its cost stated |

**This translation — from gate output to a decision a person can make in thirty seconds on a phone — is the product's principal engineering work.** It is not presentation, and it does not live in the mobile client. The orchestrator produces the decision object; the phone renders it.

**Context.** Mobile clients for coding agents already exist — several for OpenCode alone — and they mirror the desktop session: the thread, the stream, the diff, an inline permission prompt. That is a faithful port of the chat model onto a smaller screen, and it inherits the chat model's assumption that someone is watching. It is a good answer to a different question. Given [ADR-015](#adr-015--the-humans-default-state-is-non-involvement), it is not the answer here.

**Why the command view is not a dashboard nicety.** *"How many features are drifting from their spec"* is a number almost nothing can produce, because producing it requires an **oracle**: a statement of intent a machine can check code against. A chat log is not an oracle. A repository is not an oracle. A ratified spec with `SPEC-…` ids and a fidelity gate is one. The number is the visible output of the only thing that makes this product different — which is why it gets a screen rather than a corner of one.

**What is not shown, and the test that decides.** Every element must pass:

> **Does seeing this change the decision I am about to make?**

Failing it today: the token counter, the reasoning stream, the diff by default, the live log. Each is engaging, and none of them changes an answer.

The **single exception is budget.** A spend figure is not evidence and does not describe correctness, but it genuinely changes one decision — whether to let the loop keep grinding. It earns its place on those grounds alone, and no other exception inherits from it.

**None of this is concealment.** The full reasoning trace and the raw agent conversation are available **in their entirety and without restriction**, freely opened and freely closed. The objection to the reasoning stream is that it was made the *centre*, not that it exists. A capability that is one tap away but not in the way costs nothing and settles the argument honestly — anyone who wants to watch, can.

**Consequences.**

- The translation layer belongs to **P1 (orchestrator)**, not P3. The phone gets correspondingly thinner: it renders decision objects and posts answers.
- The evidence bundle gains a second, human-facing form. Both are produced and both are retained — the raw bundle is the appeal path when the summary is disputed.
- Push payloads stay content-free ([ADR-004](#adr-004--end-to-end-encryption-of-relay-payloads)); the decision object is fetched and decrypted on device.
- "Options with costs" implies the orchestrator can *price* an option — even coarsely, as time, tokens or blast radius. Where it cannot, it says so rather than inventing a number ([P3 fail-loud](../.rla/PRINCIPLES.md#p3--fail-loud)).

**Known risk.** A summary is a lossy compression of evidence, and a *wrong* summary is more dangerous than raw output because it is easier to act on. Mitigations: the raw bundle is always one tap away; the summary cites the `SPEC-…` id it rests on; and a decision recorded against a summary later contradicted by the raw evidence is a tracked defect class, not an accident.

---

## ADR-017 — Cheap models grind; an expensive model arbitrates

**Decision.** Roles are assigned to a **price tier** as well as to an agent:

| Layer | Work | Model |
| :--- | :--- | :--- |
| **Grinding** | Implementation, re-runs, iterating until the gates go green | The cheapest capable model, on a flat-rate subscription |
| **Arbitration** | The cross-verification pass, and constructing the argument a human decides on | A stronger model, invoked rarely |

**Context.** Cross-verification costs 2–3× the tokens of a single pass ([ADR-013](#adr-013--the-product-is-cross-verification-not-an-agent)). Applying one tier uniformly fails in both directions: an expensive model everywhere makes the technique unaffordable, and a cheap model everywhere puts the weakest judgement at the exact point where a human's decision rests on it.

**Rationale.** The two jobs have different economics because they have different failure costs. Grinding is high-volume, low-stakes per call, and **recoverable** — a bad attempt is caught by the next gate run, and the correction is another cheap call. Arbitration is low-volume and unrecoverable in a specific way: it is what the human sees, and a wrong arbitration is acted upon. Spend follows stakes.

This is also why the cheap end of the market is a **technical prerequisite** rather than a positioning choice. Without flat-rate subscriptions at the grinding layer, the token multiplier makes the whole method uneconomic and there is no product ([ADR-012](#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).

**Consequences.** Role configuration carries a tier alongside an agent. Tier is a property of the agent the **user** configured — we ship no ranking of vendors and no recommended pairing that favours one, so [P0 vendor neutrality](../.rla/PRINCIPLES.md#p0--vendor-neutrality) is untouched. A user who wants one model everywhere may have it; the tiering is a default that can be collapsed, not a constraint.

---

## ADR-018 — The primitive unit is a module, and intent is a hierarchy

**Decision.** The product is **Agentic Project Management**. Its primitive unit is not a message and not a feature — it is a **module**, the top of a four-layer intent hierarchy in which **each layer is the oracle of the one below it**:

| Layer | Fixes intent in | Owner | Checkpoint |
| :--- | :--- | :--- | :--- |
| **Module** | **Business language.** Acceptance criteria are outcomes a non-technical person can observe. | Human | **⓪** approves |
| **Spec** | **Technical language.** `SPEC-…` ids, measurable requirements, invariants. | Loop | **①** conditional |
| **Feature + invariant** | Code that cites a `SPEC-…` id. | Loop | automatic |
| **Code** | — | Loop | Tier 0–3 gates produce evidence |

**A module document may not name mechanism.** No endpoint, no table, no class, no library. *"A dealer can see their own balance correctly"* is an acceptance criterion. *"`GET /dealers/{id}/wallet` returns 200"* is not — it is a spec, one layer down, and writing it at module level destroys the oracle by collapsing two layers into one.

**A module may produce at most six ① checkpoints.** The budget is denominated in **decisions, not calendar time.** Duration is an output of that number, never an input to it. A module projected to exceed six is too large and must be split before ⓪.

**Context.** [ADR-013](#adr-013--the-product-is-cross-verification-not-an-agent) made a ratified spec the oracle, which was right and insufficient. A spec is written in the vocabulary of the implementation, and a human ratifying one is being asked to audit a technical document — the exact task their domain knowledge is worst at and the loop is best at. Meanwhile the question the human is uniquely qualified to answer — *"is this the right thing to build at all?"* — was being asked in a language that made it hard to answer.

Worse, a spec-only hierarchy has a blind spot with no possible detector. When every spec is ratified, every gate is green and the code is entirely correct, nothing in the system can observe that **a part of what was actually wanted was never specified.** There is no artifact above the spec to compare against.

**Rationale.** Putting a business-language layer above the spec does three things at once, which is why it is one decision and not three:

*It moves critical questioning to where competence lives.* At ⓪ the user reads *"a dealer with insufficient balance cannot transact and understands why"* and knows immediately whether that is right. At ① they would have read an idempotency invariant and taken it on trust.

*It creates the missing oracle.* Acceptance criteria are a checkable statement of intent above the spec layer, which is what makes [M2](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write) possible at all.

*It makes "how much of what I asked for exists" computable.* Coverage of criteria by specs is a number, not an impression — and only a system that treats an approved module as an oracle can produce it.

**Consequences.** The top-level unit of the project surface is a module, not a feature ([ADR-021](#adr-021--three-tabs-and-an-asynchronous-command-bar)). Features become a sub-breakdown. The scope boundary — an explicit **Included / Excluded** pair — is part of the module document, because a boundary that is not written cannot be violated detectably. The six-① budget is enforced as an estimate at ⓪ and measured afterwards; exceeding it is a signal that modules are too big or criteria too vague, and it is shown to the user as such.

**Known risk.** Business-language criteria can be written vaguely enough to be unfalsifiable — *"the user understands why"* is not observable until it says what appears on screen. The adversarial verifier is pointed at exactly this at module-draft time, and rejects criteria it cannot state an observation for.

---

## ADR-019 — Three intent gates, and a module document the loop cannot write

**Decision.** Three gates enforce the hierarchy in [ADR-018](#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy). **They are what makes it a layer rather than a folder.**

| Gate | Runs | Catches |
| :--- | :--- | :--- |
| **M1** | Every spec draft | A spec that names no `MOD-K` criterion it satisfies, with justification → **orphan**. A spec touching the module's **Excluded** list → **scope violation**. |
| **M2** | Module candidate-complete | An acceptance criterion **no spec serves**. |
| **M3** | Continuously | A ratified spec whose module criterion no longer exists → **orphan spec**. |

**M2 is the reason the layer exists.** It fires when every spec is ratified, every Tier 0–3 gate is green and the code is entirely correct — and reports that *a part of what was asked for was never built.* No other gate in the system can see this, because every other gate compares code against a spec, and the defect is that the spec was never written.

**While M2 is red, ② cannot start.** Interface-testing a knowingly incomplete product is not a test.

**And the module document is the one artifact the implementation loop cannot write.** Enforced structurally — a deny rule plus a `PreToolUse` hook exiting non-zero — not by instruction. A loop that can edit its own targets has no targets.

**The friction is deliberate for the human too.** M2 red can legitimately be closed by lowering an acceptance criterion — the goal really was too ambitious sometimes. But that **cannot be one tap.** The module returns to `draft`, a written justification is required, and it is re-approved from ⓪.

**Context.** [P5](../.rla/PRINCIPLES.md#p5--gates-are-immutable-to-the-loop) already forbids the loop from weakening a gate to turn red green. Lowering an acceptance criterion to clear M2 is **the same act performed by the human hand**, and it is the single most dangerous thing a phone interface can make easy: one tap, in a queue, between meetings, with no record of why.

**Rationale.** The asymmetry is intentional and is the whole design. Approving is cheap because approving does not lower a bar. Lowering a bar is expensive because it is the one action that makes every downstream green meaningless. A confirmation dialog would not do — the cost has to be *writing a reason and re-entering ⓪*, because that is what produces an artifact a future reader can weigh.

**Consequences.** `.rla/modules/` joins `.rla/specs/` with the same on-disk shape the product will create in a user's repository. M1/M2/M3 join the gate ladder as **Tier M**, run by the arbitration-tier model ([ADR-022](#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer)) since reading intent is judgement, not pattern-matching. The module-write deny rule is itself covered by a canary: a gate that cannot prove it blocks the write is not protecting anything.

**Known risk.** M1 and M2 rest on model judgement about whether a spec serves a criterion, and a compliant-sounding spec could be accepted for a criterion it does not really serve. The mitigation is that the mapping is **shown to the human at ②**, criterion by criterion, with the covering spec named — so a bad mapping surfaces at the interface test rather than never.

---

## ADR-020 — Three checkpoints, and the middle one is conditional

> **Supersedes the checkpoint set in [ADR-015](#adr-015--the-humans-default-state-is-non-involvement).** That entry's thesis — the human's default state is non-involvement — stands unchanged. What changes is which occasions qualify.

**Decision.** Three checkpoints, plus arbitration:

| | Checkpoint | Scope | When |
| :--- | :--- | :--- | :--- |
| **⓪** | **Module approval** | Per module | Always. **This is where critical questioning happens.** |
| **①** | **Spec approval** | Per spec | **Conditional — usually never reaches the user.** |
| **②** | **Interface test** | Per module, once | When M2 is green. |
| **③** | Arbitration | Per red gate | The loop cannot resolve it alone. |

**① auto-ratifies** when all three hold:

1. **M1 is green** — the spec maps to a module criterion and violates no scope boundary.
2. **It introduces no invariant absent from the module document.** A new rule is new intent, and new intent is the human's.
3. **Its risk class is LOW or MEDIUM.**

It escalates on **HIGH class**, a **new invariant**, or **M1 red**.

**And when it does escalate, the screen must state why it escalated.** This is not a courtesy. An automatic approval a user cannot audit is an automatic approval they will not trust, and a user who does not trust it will disable it and read every spec — which returns the product to the thing it was built to replace. *"Risk class HIGH — money, irreversible record, concurrency. Auto-approval is disabled for this class"* is what makes the other seventeen silent approvals credible.

**Context.** [ADR-015](#adr-015--the-humans-default-state-is-non-involvement) put spec ratification at ① unconditionally. Running that against a real module shape gives roughly one ① per spec, four or five specs per module, and a user reading technical documents all day — which is watching, wearing different clothes.

**Rationale.** *The system cannot earn the right to interrupt before it has earned the right to correctly not interrupt.* Every checkpoint the product spends is drawn from a finite budget of the user's willingness to be interrupted, and spending it on a LOW-risk spec that restates a criterion the user already approved at ⓪ is spending it on nothing. ⓪ is where the user's domain knowledge does work no model can do. ① is where it mostly does not — except in the specific cases above, which is why they are the exceptions rather than the rule.

**Consequences.** The **interrupt counter is a shipped feature, not internal telemetry**: interrupts per module, auto-approval ratio, and **time spent away from the desk** — because that, and not time in the app, is the metric this product optimises. Exceeding roughly eight interrupts per module is surfaced as a diagnosis in those terms: *the modules are too large or the criteria too vague.*

② is per module, not per feature, and is completed by observing acceptance criteria one at a time. **A half-ticked module does not close** — it returns to the queue. Ticking is a claim of observation, so the criteria have to have been written observably, which is enforced back at ⓪ ([ADR-018](#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy)).

**Known risk.** Auto-ratification is a genuine transfer of authority, and its correctness rests on the risk classifier. A misclassified HIGH spec is silently approved. Mitigations: classification is recorded and reviewable per spec; the per-module override can disable auto-approval wholesale; and the auto-approval ratio is displayed, so a run of silent approvals is visible rather than invisible.

---

## ADR-021 — Three tabs and an asynchronous command bar

> **Supersedes [ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread).** The translation layer, the display test and the unrestricted-raw-access rule carry over unchanged. The screen set and chat's placement change.

**Decision.** Three tabs. Reference design: [`design/mobile-v2.html`](../design/mobile-v2.html).

| Tab | Holds | Empty means |
| :--- | :--- | :--- |
| **Queue** | Pending ⓪, ①, ② and arbitration — nothing else | ✅ Success. **The screen says you can close the app.** |
| **Modules** | Per project: modules, complete, **scope gaps**, awaiting approval; which criterion is met by which spec | Nothing in flight |
| **Settings** | Agent assignment, loop method, auto-approval conditions, budget, visibility, **interrupt counter** | — |

**Chat is not removed, and it is not a tab.** It is a **context-aware command bar at the bottom of every screen** — scoped to the project, the module or the feature the user is looking at.

It is **asynchronous.** You send and you leave; the bar confirms *"sent — you can go"* rather than opening a thread to wait in.

**Input is conversation. Output is structure.** The reply is never a wall of text. It comes back as a **module draft**, a **spec card**, or an **optioned decision** — an object with actions on it. A free-text answer would reconstitute the chat product one message at a time.

**Carried over from [ADR-016](#adr-016--the-interface-is-three-screens-not-a-chat-thread), unchanged:**

- A red gate arrives as **violated invariant · one-sentence why in domain language · two or three options with their costs.** This translation is the product's principal engineering work, and it happens in the orchestrator, not the client.
- **The display test:** *does seeing this change the decision I am about to make?* Budget is the sole exception.
- The reasoning trace and raw agent stream are available **in full, without restriction**, freely opened and closed — never the default, never the centre.

**Rationale for the top-level unit being the module.** A feature list is a technical inventory; a module list is a statement of what the product is for, readable by someone who does not code. And *"one module has a scope gap"* is the single most valuable number on the surface, because it is the one an all-green build cannot otherwise tell you ([ADR-019](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write)).

**Consequences.** The empty state is designed as the destination rather than as a placeholder, and reports **time away from the desk**. Settings states provider keys as **"on device"** — presented as a commitment, not a preference, with no UI path that could make it false ([P1](../.rla/PRINCIPLES.md#p1--zero-touch-ai)). Push bodies stay content-free on the wire and are rendered on-device after decryption ([ADR-004](#adr-004--end-to-end-encryption-of-relay-payloads)); a lock-screen card showing a module name is local rendering, not relay-visible payload.

The **"lower the criterion"** action exists in the M2 view and is deliberately **not** a one-tap resolution — it reopens ⓪ per [ADR-019](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write). "Leave it to me" is a first-class option on any arbitration: the decision stays queued for the desk, which is an answer, not a deferral bug.

---

## ADR-022 — Tiered model assignment, and the verifier is never the producer

> **Supersedes [ADR-017](#adr-017--cheap-models-grind-an-expensive-model-arbitrates).** The economic argument stands; the tier set is finer and one rule becomes mandatory.

**Decision.**

| Tier | Work | Model | Bound |
| :--- | :--- | :--- | :--- |
| **M** | Intent gates M1/M2/M3 | Expensive — reading intent is judgement | per spec draft / per module |
| **T0** | Format, lint, types | **None.** Deterministic, zero cost | every save |
| **T1** | Producer — implementation, inner loop, diff-scoped | Cheapest capable, flat-rate subscription | ceiling ~30 rounds |
| **T2** | **Adversarial verifier** — convergence, per invariant | **Must not be the producer's model** | per feature |
| **T3** | Arbiter — heavy verification, called on contradiction | Expensive, rare | ceiling ~3 rounds |
| **T4** | Periodic sweep — CVEs, dependencies, dead code | Cheap | weekly, off by default |

**The mandatory rule: T2's model may never equal T1's.** If a user configures them the same, **the T2 gate is void and the ladder halts** — it does not silently pass. A model grading its own output is the failure this product exists to correct, and permitting it while displaying a green T2 would be the worst outcome available: a fake green with a credential.

**Context.** [ADR-017](#adr-017--cheap-models-grind-an-expensive-model-arbitrates) split grinding from arbitration on cost grounds. Two things were missing. Intent gates ([ADR-019](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write)) are neither grinding nor arbitration and deserve the strong model for a different reason — they read meaning, not correctness. And "a different vendor's model verifies" was stated as a property of the product but enforced nowhere.

**Rationale for halting rather than warning.** A warning is an interrupt that changes no decision, which [P11](../.rla/PRINCIPLES.md#p11--nothing-is-shown-that-does-not-change-a-decision) forbids anyway. More importantly, `COULD NOT VERIFY` already exists precisely for a gate that cannot do its job ([P3](../.rla/PRINCIPLES.md#p3--fail-loud)), and a self-graded T2 is exactly that case. Treating it as anything other than unverified would make the status meaningless everywhere else.

**Consequences.** Assignment is per project with a **per-module override** — a module touching tax law can promote its verifier — and an override carries a **written justification**, so a future reader learns *why* this module cost 3× the default. Cost per spec is estimated and shown before the spend, not reported after it. [P0 vendor neutrality](../.rla/PRINCIPLES.md#p0--vendor-neutrality) is untouched: every tier names an agent the **user** configured, and the one constraint we impose is *difference*, which favours no vendor.

---

## ADR-023 — A judged gate may add a finding, never clear one

**Decision.** Every intent gate splits into a deterministic half that **binds** and a judged half that is **advisory**:

| Gate | `a` — deterministic, binding | `b` — judged, advisory |
| :--- | :--- | :--- |
| **M1** | Does the spec declare a `MOD-K` that exists? Set membership. | Does its subject fall in the module's **Excluded** list? |
| **M2** | Is every criterion declared by at least one spec? Set arithmetic. | Does the declaring spec actually serve it? |
| **M3** | Do all declared criteria still exist? Set arithmetic. | — |

**The `b` half may raise a finding. It may never clear one.** A green intent gate therefore means *"the declarations are complete, and nothing was flagged as false"* — never *"a model confirmed the coverage is real."*

**Context.** As first specified, M2's judgement could turn the gate green. Its failure mode is silent: a compliant-sounding spec accepted for a criterion it does not serve produces a green M2 on an incomplete module — the exact outcome the gate exists to prevent, wearing a badge. The worst case has to be a **miss**, not a false clearance.

**Rationale.** A judged gate that can only accuse has a bounded failure: it stays quiet when it should have spoken. A judged gate that can also absolve has an unbounded one: it speaks when it should have stayed quiet, and its word carries the gate's authority. Only the first is survivable in a system whose whole promise is that the human can stop watching.

**Consequences — and the economics change materially.** The deterministic halves need **no model at all**, which moves the highest-frequency intent check (M1, on every spec draft) off the expensive tier entirely. What remains there is M1b, M2b and T3 — all rare. The claim in [ADR-022](#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer) that cheap models make the method affordable did not previously cover Tier M; with this split, it does.

The honest confirmation of a `b`-half question happens at **②**, where a human reads the criterion→spec mapping. That is late, and it is the correct place: it is the only point where someone who knows what they wanted is looking.

---

## ADR-024 — Silence is attested, not assumed

**Decision.** An empty queue must carry **when the system last confirmed it was working**. Silence with no heartbeat past a configured threshold **becomes the alarm itself**.

This is the **second and final exception** to [P11](../.rla/PRINCIPLES.md#p11--nothing-is-shown-that-does-not-change-a-decision), alongside budget. It is the only interrupt permitted to fire with no decision attached, because its content *is* the decision: *stop trusting the silence.*

**Context.** The product's central promise — *nothing is waiting for you, close the app* — asks the user to read an empty screen as success. But a dead orchestrator, a sleeping machine and a dropped notification all produce exactly the same empty screen. An interface that says *"2 modules running"* nine hours after the process died is not merely unhelpful; it is a lie shaped like good news, and it is the failure that would discredit everything else.

[ADR-009](#adr-009--the-cli-host-must-be-reachable) covers the case where unreachability is *known*. This covers the case where nothing is known at all, which is the more dangerous one.

**Rationale.** Alarm engineering settled this long ago: a monitoring system must prove it is alive, because the absence of an alarm and the absence of a monitor are indistinguishable to the operator. Positive attestation is not a nicety; it is what makes the negative signal mean anything.

**Consequences.** The empty state renders a timestamp and what is still in flight, and degrades — visibly, then audibly — as the heartbeat ages. No further exception to P11 is granted by analogy: budget changes whether to keep spending, and the heartbeat changes whether to believe the screen. Nothing else has a claim of that kind.

---

## ADR-025 — Every invariant declares the criterion it derives from

> **Replaces the second escalation test in [ADR-020](#adr-020--three-checkpoints-and-the-middle-one-is-conditional).** The other two conditions and the rest of that entry stand.

**Decision.** Every invariant in a spec declares the `MOD-K` it derives from. **An invariant with no declared parent is by definition new**, and its spec escalates to ①.

**Context.** ADR-020 made auto-ratification depend on the spec introducing *"no invariant absent from the module document."* That test cannot work as written, because [ADR-018](#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy) forbids the module from naming mechanism — so **every technical invariant is absent from the module by construction.**

What the test actually asks is whether a technical rule is *entailed* by a business criterion. That is open-ended semantic judgement, put in charge of whether a human ever sees the spec. And it fails asymmetrically: a generous model approves silently, a strict one escalates everything and turns the product back into reading specs all day. There is no obviously correct setting.

**Rationale.** The same move that fixed M2 ([ADR-023](#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)) fixes this: convert the semantic question into a **declaration check**. *"Which criterion does this rule come from?"* is answerable and recordable; *"is this rule new?"* is neither.

It will over-escalate at first, and that is the point. **The escalation log becomes the corpus the risk classifier is built from** — every escalation a human waves through is a labelled example of a legitimate derivation. The alternative was inventing the classifier before seeing a single real case.

**Consequences.** Spec artifacts gain a parent declaration per invariant, checked by M1a — mechanical, no model. Whether the declared derivation is *honest* is an M1b question: advisory, never clearing. Early escalation rates will be high and are expected to fall as the corpus grows; that curve is itself the measurement of whether the classifier is definable at all, which is currently an open question in [`module-layer.md`](../.rla/specs/module-layer.md).

---

## ADR-026 — The hierarchy relocates the blind spot; it does not remove it

**Decision.** State plainly, in the documentation and in the product, that the intent hierarchy **moves** the class of defect it was built to catch rather than eliminating it. And put adversarial pressure at the new location: at module draft, a verifier attempts to **name a criterion the draft is missing**, given the module's *why it exists* statement.

**Context.** [ADR-018](#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy) introduced the module because nothing could detect *"a part of what was wanted was never specified."* [M2](#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write) detects a **criterion with no spec.** Nothing detects a **requirement with no criterion** — the identical defect, one layer up, at the layer that has no oracle above it.

It is worse than a symmetry argument suggests, because the module draft is written by an agent and merely approved by a human. A plausible, well-structured three-criterion draft gets approved; the criterion nobody thought of is exactly as invisible at ⓪ as it was everywhere else. That risk was recorded in ADR-018 as *vagueness*, which was the smaller half of it.

**Rationale.** It cannot be closed — nothing can prove a person wanted something they did not say. But it is currently not pressured at all, and the machinery already exists: the verifier objects to criterion *quality* at draft time ([SPEC-module-layer-04](../.rla/specs/module-layer.md)). Pointing the same pass at *completeness* is a small change to an existing step.

**Consequences.** The claim in the public copy changes from elimination to relocation, with the reason it is still worth it: **a missing criterion costs a conversation at ⓪; a missing spec costs a rebuild.** The blind spot moves to the cheapest place to have one. That is a real win and it survives being stated accurately, which is the test any claim in this repository has to pass.

---

## ADR-027 — A deterministic verdict is an exit code, and judgement reads it

**Decision.** Every check whose output is an exit code belongs to the script, runs as one of two modes (`fast`, `full`) and writes one machine-readable artifact. Every check requiring judgement **reads that artifact and runs no tools**, runs **once at a defined trigger**, and **reports rather than blocks**. Blocking authority stays with the exit codes and with the readiness verdict.

**Context.** The method's earlier shape gave each gate to a model that invoked the deterministic tools itself. Measured on the codebase where it was first run by hand: the complete deterministic pass took **~2 minutes**; the same work through a model took **10–21 minutes per gate**; one round of 41 gate runs took roughly **nine hours** and produced no verdict.

Two causes, and neither of them was verification. **Context rediscovery** — nearly all of a gate's time went to working out again which tool to run and why, relearning the project every round. **Judgement production** — the gate reported its own opinion instead of the tool's exit code, and the same unchanged code passed one round and failed the next, four times over.

**Rationale.** The second cause is the fatal one. *"Iterate until everything is green"* has no terminating condition in a system where green is not stable, so the loop cannot converge — not slowly, but never. Determinism is not merely cheaper here; it is what makes the stopping rule exist.

Reading is also the fix for the first cause. The artifact hands a reviewer what ran and what it saw, which is precisely the context it was previously paying to rediscover.

**Consequences.** `scripts/gate` gains `fast`, `full`, `evidence` and `timings`; tiers stay as classification while modes become the unit of running. `gate verify` prints the artifact path with the review obligations, so the reading contract is visible at the point it applies. Judged gates keep their design authority from [ADR-023](#adr-023--a-judged-gate-may-add-a-finding-never-clear-one) — accuse, never absolve — and gain a second bound: they may not block a round.

The regression to watch for is stated in [`loop-engineering.md`](loop-engineering.md#self-audit--the-three-numbers) as **backbone leakage**: a review pass calling `go test`. It always arrives disguised as convenience, and it restores both the cost and the instability in one step.

---

## ADR-028 — Counts are evidence, and every convenience buys a guard

**Decision.** Every gate reports how much it examined — tests run, files scanned, requirements checked — and those counts are held to declared floors, to a committed baseline, and to a step budget. Guards may worsen a verdict and never improve one. A **cache hit is re-judged against its recorded counts**, and an evidence artifact is served only when its fingerprint matches the working tree.

**Context.** An exit code answers *"did what ran pass?"* and cannot answer *"did anything run?"*. Both silent greens found in the field were the second kind: a suite selecting **zero tests** while exiting 0, and a collection error dropping **all 18,034 tests** behind calm-looking output. Neither was caught by a gate; both were caught by someone eventually looking at a number.

**Rationale.** Every speed decision in this pipeline — caching, tier splitting, deferring a check to a later mode — trades certainty for time. Each is defensible only while the counts are still checked, which makes guards the price of the speed rather than an addition to it. It follows that the loop must not be able to edit them: a loop that can delete the price gets the convenience free, and the system reverts silently to the state that made the guards necessary. They join the coverage floor under [P5](../.rla/PRINCIPLES.md#p5--gates-are-immutable-to-the-loop).

The cache is the sharpest case. The files that make a suite run zero tests do not change, so the signature keeps matching and the empty pass is served indefinitely — a cache is where a silent green goes to become permanent. Storing the evidence beside the verdict is what lets a hit expire.

**Consequences — and one deliberate divergence.** The guard list was **not** ported verbatim. The original guards collection errors separately because a Python collection failure can drop a suite quietly; in Go a package that will not build makes `go test` exit non-zero, so that guard could never fire here. What Go *does* hide is erosion — 70 tests becoming 50 while every remaining one passes — so the ratchet in `.rla/test-baseline.txt` carries that weight instead. Copying the list unexamined would have added a dead check and omitted a live one.

Two costs are accepted. The suite runs with `-count=1`, forgoing Go's own test cache, because a cached package emits no per-test events and would look like a suite that vanished. And the test baseline counts subtests, so splitting a table test into cases requires the ratchet to move — a mildly annoying and entirely visible act, which is the correct direction for that annoyance to point.
