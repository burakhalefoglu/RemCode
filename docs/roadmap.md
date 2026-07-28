# 🗺️ RemLinkAgent — Roadmap

> **Version:** 8.1 · **Updated:** 2026-07-28 · **Status:** P0 finishing, P1 next
> **Thesis:** Agentic Project Management — a control room, not a chat window. You approve a **module** in business language; agents write the specs and the code; a *different* model tries to break it; gates produce evidence. The human is **not involved** in between. See [ADR-018](decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy), [ADR-015](decisions.md#adr-015--the-humans-default-state-is-non-involvement), [ADR-020](decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional).
> **Reference design:** [`design/mobile-v2.html`](../design/mobile-v2.html)

Covers **P0–P4** and **M1**. Deferred capability is in [`vision-roadmap.md`](vision-roadmap.md); the reasoning behind every scope choice is in [`decisions.md`](decisions.md).

**Tracking:** `[ ]` todo → `[~]` in progress → `[x]` done. `⚠️` blocked, `➕` optional.

---

## 0. Two tracks, not one

This is not a build with a method attached. It is **two workstreams that feed each other**, and the phase list below is only one of them.

| | **M-track — the method** | **P-track — the product** |
| :--- | :--- | :--- |
| **What** | [`loop-engineering.md`](loop-engineering.md) as a **living best-practice document** | P0–P4, the phases below |
| **How it advances** | Run by hand on real projects. Every gap, wrong call and false alarm is folded back in. | Implements whatever the method has settled |
| **Cadence** | Continuous, no phases | Phased |
| **Relationship** | **Upstream.** The method decides what the product must do. | Downstream. Never invents method. |

The M-track is already running and predates this repository — the maintainer works this way manually today. That is the reason the product can be specified at all, and it is why an unsettled part of the method must **not** be implemented ahead of use: guessing at a rule and shipping it costs more than waiting for the manual run to produce it.

**A practical consequence.** When the two disagree, the method wins and the product changes. A product behaviour that has no counterpart in the manual practice is a feature nobody asked for.

**And the direction of proof runs the other way too.** The product is dogfooded on the maintainer's own real projects, so a method rule that survives manual use but collapses under automation is caught before anyone else meets it.

---

## 1. What is being built

An **Agentic Project Management** system. The intent hierarchy ([ADR-018](decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy)) is the spine — **Module → Spec → Feature + invariant → Code**, each layer the oracle of the one below.

Work is done by real agents that already exist, driven over the [Agent Client Protocol](https://agentclientprotocol.com/) and assigned per tier ([ADR-022](decisions.md#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer)):

| Tier | Job | Model | Bound |
| :--- | :--- | :--- | :--- |
| **M** | Intent gates M1/M2/M3 | expensive — reading meaning | per spec draft / per module |
| **T0** | Format, lint, types | **none** — deterministic, free | every save |
| **T1** | Producer, diff-scoped inner loop | cheapest capable, flat-rate sub | ~30 rounds |
| **T2** | **Adversarial verifier** | **never the same model as T1** | per feature |
| **T3** | Arbiter, on contradiction | expensive, rare | ~3 rounds |
| **T4** ➕ | Periodic sweep — CVEs, dead code | cheap | weekly, off by default |

The human appears **three times and no more**, plus arbitration: **⓪** approve the module, **①** approve a spec *only when it matters*, **②** observe the criteria once per module ([ADR-020](decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional)). Between them, silence is the success state.

**In scope**

- ✅ **Module artifacts in business language** — observable criteria, scope boundary, ⓪ ratification ([spec](../.rla/specs/module-layer.md))
- ✅ **Tier M intent gates** — M1 spec↔module fidelity, **M2 uncovered criterion**, M3 orphan spec
- ✅ **Conditional ①** — auto-ratification with a stated reason whenever it escalates
- ✅ ACP orchestration of multiple agents, each with its own model, subscription and tier
- ✅ Automatic handoff: implement → gates → cross-review → evidence
- ✅ Deterministic gates: lint, tests, coverage ratchet, spec fidelity, fake-green, canaries
- ✅ **Decision objects** — a red gate rendered as invariant, one-sentence why, and 2–3 priced options ([spec](../.rla/specs/decision-object.md))
- ✅ Every proposed command approved by a human before it runs
- ✅ End-to-end encrypted relay, lossless background sync
- ✅ **Three-tab mobile client** — queue, modules, settings, plus an async command bar ([ADR-021](decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar))
- ✅ **Interrupt counter** — interrupts per module, auto-approval ratio, time away from the desk
- ✅ Direct PAYG path against any OpenAI-compatible endpoint, as a fallback

**Out of scope** → [`vision-roadmap.md`](vision-roadmap.md)

- ❌ Our own agent loop, tool framework or model router ([ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one))
- ❌ Calling any provider's *subscription* endpoint ourselves ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us))
- ❌ **A chat thread as the top-level container** — the opposite bet ([ADR-021](decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar))
- ❌ **Scheduling-style project management** — no dates, no Gantt, no burndown. The only budget is a decision count.
- ❌ Interactive terminal mirroring → X4
- ❌ Voice, smartwatches → X2, X3
- ❌ Managed cloud, teams, billing → M1

### Invariants — not negotiable

| | |
| :--- | :--- |
| ⚖️ **Vendor-neutral** | We sell no model. That is what makes arbitration credible. |
| 🔑 **Zero-Touch AI** | API keys and model traffic never reach the relay. It is a control channel; it never proxies a model call. |
| 🎭 **No impersonation** | Agents identify honestly. We never present another tool's identity. |
| 🔐 **E2E encrypted** | The relay handles ciphertext and routing metadata. Nothing else. |
| 🟢 **Fail-loud** | `COULD NOT VERIFY` is never a pass. |
| 📐 **Evidence, not assertion** | A green gate must be able to prove it can still detect breakage. |
| 🧱 **Intent hierarchy** | Module in business language, spec in technical. Mechanism never appears at module level. |
| ✍️ **The loop cannot write a module** | Structurally denied. A loop that edits its own targets has no targets. |
| ⚔️ **Verifier ≠ producer** | T2's model may never be T1's. Same model → gate **void**, ladder halts. Never green. |
| 🧮 **Judgement accuses, never absolves** | Every intent gate has a deterministic binding half; the judged half may raise a finding, never clear one. |
| 🔕 **Silence by default** | Only a pending checkpoint may interrupt. Every avoidable interrupt is a defect. |
| 💓 **Silence is attested** | An empty queue carries when the system last confirmed it was working. Stale silence *is* the alarm. |
| 👁️ **Decision-relevance** | Nothing is displayed that does not change a decision — budget excepted. Nothing is *hidden* either. |
| 🪶 **Light** | Single static binary, < 30 MB RSS; mobile cold start < 2 s. |
| 🌍 **Localised** | TR + EN, i18n keys only. |

---

## 2. Schedule

**17–22 weeks for one developer working full time.**

| Phase | Estimate |
| :--- | :--- |
| [P0](#p0--scaffolding-gates--de-risking) — Scaffolding, gates, de-risking | 1.5–2 weeks |
| [P1](#p1--orchestrator) — Orchestrator **+ module layer + decision objects** | 6–8 weeks |
| [P2](#p2--relay-wss--nats) — Relay | 3 weeks |
| [P3](#p3--mobile-client) — Mobile client | 4–5 weeks |
| [P4](#p4--release--distribution) — Release | 3–4 weeks |
| [M1](#m1--managed-cloud-subscriptions--team) — Managed cloud | 5–6 weeks (after MVP) |

Down from 17–20 in v4.0: [ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one) removed the agent loop, tool framework, sandbox, danger classification, model router, context normalisation and provider contract tests — most of the old P1.

Back up since: **P1 has absorbed the two hardest things in the product.** The decision-object translation layer ([1.22–1.27](#decision-objects--the-translation-layer)) and the module layer with Tier M ([1.28–1.39](#module-layer--tier-m--the-intent-hierarchy)) are both new and both load-bearing — M2 in particular is the gate no competitor has, and it is judgement-based, so it will take iteration rather than implementation. P3 does not get correspondingly cheaper on paper: the saving there is *scope removed*, not effort moved.

**These numbers assume the open questions in the specs get answered, not researched.** Risk classification and the "new invariant" boundary ([`module-layer.md`](../.rla/specs/module-layer.md)) are the two most likely to expand. [P0.18](#p0--scaffolding-gates--de-risking) exists to answer them from one real run rather than from reasoning.

### The honest calendar

**Weeks of work ≠ weeks elapsed.** The table above counts full-time engineering effort. Realistically this is **8–12 months** to something usable, and the gap is not padding:

- It is one person, part-time, alongside the projects the product is meant to serve.
- The hardest components — M1/M2, the risk classifier, the translation layer — are **judgement-based**, and judgement-based work consumes iteration, not implementation. Getting M2 to fire correctly is a tuning problem measured in real modules, not in commits.
- The M-track is upstream and asynchronous: some of P1 cannot start until manual use has settled a rule.

Planning against 17–22 weeks and discovering this in month four is a worse outcome than starting from the real number.

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
- [ ] **0.16** Backward fidelity diff — behaviour with no covering requirement. **Splits in two:** the structural half (an endpoint, table, dependency or package no spec declares) is set arithmetic and belongs in the script; the semantic half needs judgement and stays a reviewer task. Design only in P0.

> **The deterministic backbone** ([`deterministic-backbone.md`](../.rla/specs/deterministic-backbone.md), ① awaiting ratification).
> Everything whose output is an exit code runs in one place, as one of two modes, and writes one
> artifact — because the alternative was measured at 10–30× the cost with an unstable verdict
> ([ADR-027](decisions.md#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it)).

- [x] **0.19** **Modes as the unit of running:** `gate fast` (tiers 0–1) and `gate full` (0–3), with every deferred check named in the output. Tiers stay the unit of classification.
- [x] **0.20** **Evidence artifact** — `.rla/state/verify-<mode>-<fingerprint>.json`: per step the verdict, exit code, duration, cache status, evidence counts, guard results; plus deferrals and the obligations owed to judgement. `gate evidence` serves it, or refuses when the fingerprint has moved on.
- [x] **0.21** **Guards** — empty-run floors, the test-count ratchet in [`.rla/test-baseline.txt`](../.rla/test-baseline.txt), step budgets, and cache hits re-judged against their recorded counts. Guards may worsen a verdict, never improve one ([ADR-028](decisions.md#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard)).
- [x] **0.22** **Measured wall clock** — `gate timings [-record]`, baseline committed in [`.rla/tool-timings.json`](../.rla/tool-timings.json), drift reported on every run. Budgets are ceilings set after measuring: `fast` 5.8 s of 2 min, `full` 10.4 s of 30 min.
- [ ] **0.23** Guard liveness in `gate canary` — trip each guard deliberately (delete a test, exceed a budget, corrupt a cached count) and prove it turns red. Guards that die are invisible until something needed them.

### De-risking spike

- [ ] **0.17** ⚠️ **ACP capability matrix.** For Qwen Code, Kimi Code and OpenCode, verify with real sessions:
  - [ ] **0.17.1** `session/request_permission` — payload shape, what is included about the proposed action, whether it can be deferred long enough for a phone round trip.
  - [ ] **0.17.2** Session lifecycle: create, resume, cancel; what survives an agent restart.
  - [ ] **0.17.3** How a completed diff is retrieved for handoff to the reviewer.
  - [ ] **0.17.4** Whether one orchestrator can hold several agent sessions concurrently without interference.
  - [ ] **0.17.5** `qwen serve` HTTP+SSE specifics: `Last-Event-ID` replay, multi-client voting, TLS.
  - [ ] **0.17.6** Results into [`protocol.md`](protocol.md#acp-capability-matrix).

  > **Why this gates P1.** The whole design assumes ACP exposes approvals to an external client and that agents can be driven concurrently. If permission requests cannot survive a phone round trip, the approval flow needs a different mechanism — and that changes P1 and P3 together.

- [ ] **0.18** ⚠️ **Manual module trial — one real module, by hand, on an existing project.**

  The cross-verification thesis was believed because it was run manually first and found real defects. The module layer has not had that test, and it is a larger bet. Same move, before the code:

  - [ ] **0.18.1** Write one module by hand: business language, 4–5 criteria, explicit Included/Excluded. **Record whether that was actually possible** — where a criterion resisted being written without naming mechanism, and where "observable" was hard to satisfy.
  - [ ] **0.18.2** Have an agent draft the specs; each declares the `MOD-K` it serves. This is M1, performed by a human.
  - [ ] **0.18.3** Implement and drive the gates green with an agent, T2 on a different model from T1.
  - [ ] **0.18.4** **Then ask M2 by hand: which criterion has no spec serving it?** Record whether it found anything, and whether that finding would otherwise have shipped.
  - [ ] **0.18.5** **Record the risk class of every spec produced**, and what made it that class. This is the corpus [1.37](#module-layer--tier-m--the-intent-hierarchy)'s classifier is built from — the alternative is inventing it.
  - [ ] **0.18.6** Count the actual ① decisions and total interrupts the module produced. Compare against the working hypotheses of 6 and 8.
  - [ ] **0.18.7** Results into [`loop-engineering.md`](loop-engineering.md) as method, and into [`module-layer.md`](../.rla/specs/module-layer.md) as answers to its open questions.

  > **Why this gates P1.** [1.28–1.39](#module-layer--tier-m--the-intent-hierarchy) is six to eight weeks of work resting on three unproven claims: that intent can be written in business language without leaking mechanism, that M2 catches something a green build would otherwise hide, and that risk classification is definable at all. A week by hand answers all three. If M2 finds nothing on a real module, the layer's central justification is wrong and it is far better to learn that now.

**Done when**

- `make cli server` builds; `make verify` and `make canary` are green.
- `make fast` finishes well inside its measured budget and **names every check it deferred**; deleting a test turns the suite red on the *baseline* guard rather than on an assertion.
- `make docker-up` brings up NATS + relay; `/healthz` answers.
- Flutter app runs on a simulator.
- The ACP matrix is written up in `protocol.md`.
- **One module has been carried end to end by hand**, and `module-layer.md`'s open questions are answered from that run rather than from reasoning.

---

## P1 — Orchestrator

**Goal:** automate, on one machine, the handoff the maintainer currently performs by hand — **and turn a red gate into something answerable from a phone** ([1.22–1.27](#decision-objects--the-translation-layer): the phase's hardest problem).
**Estimate:** 4–5 weeks · **Depends on:** P0.17

**Technical decisions**

- Agents are external processes the user installs. We orchestrate; we never redistribute.
- Each agent keeps its own credentials. The orchestrator never sees a key.
- Roles are configuration, not code paths — any ACP agent can fill any role, at either tier.
- The translation layer lives **here**, not in the mobile client. The phone renders decision objects; it does not construct them ([ADR-016](decisions.md#adr-016--the-interface-is-three-screens-not-a-chat-thread)).

### ACP client

- [ ] **1.1** ACP client: JSON-RPC 2.0 over stdio, agent lifecycle, initialisation, capability negotiation.
- [ ] **1.2** `session/request_permission` handling — surface, hold, resolve, expire.
- [ ] **1.3** Streaming session updates; tool calls and diffs surfaced as structured events.
- [ ] **1.4** `qwen serve` HTTP+SSE transport alongside stdio, where the agent offers it.
- [ ] **1.5** Concurrent sessions across several agents without interference.
- [ ] **1.6** Agent health: crash detection, restart, and a loud failure when an agent dies mid-task.

### Roles & orchestration

- [ ] **1.7** `rla setup` — detect installed agents, install missing ones via **their own official installers**, run their login flows, verify.
- [ ] **1.8** Role configuration in `~/.rla/config.yaml`: which agent fills Coder, Reviewer, Explorer — and at which **tier**, grinding or arbitration ([ADR-017](decisions.md#adr-017--cheap-models-grind-an-expensive-model-arbitrates)). Collapsing every role onto one agent stays legal.
- [ ] **1.9** **Handoff:** implement → collect diff → run gates → hand diff + spec + **the evidence artifact** to the Reviewer. The Reviewer reads it and runs no tools ([ADR-027](decisions.md#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it)); re-running a decided question costs an order of magnitude more and answers it less reliably.
- [ ] **1.10** Reviewer prompt construction: spec, diff, artifact, and an explicit instruction to disprove rather than confirm. **Trigger-bounded** — once at convergence, never per round, and its output never blocks a round.
- [ ] **1.11** Finding aggregation: reviewer output plus gate output into one evidence set.
- [ ] **1.12** Iteration loop with a bounded budget; hard stop with a loud report.
- [ ] **1.13** **Unified approval queue** — permission requests from every agent in one place, one decision each.

### Specs & gates

- [ ] **1.14** `rla spec new|list|ratify` — artifact lifecycle from the CLI.
- [ ] **1.15** Gate engine promoted from `scripts/gate` into `internal/verify` as a library — **including the run artifact, guards and measured timings** ([0.19–0.22](#verification-system--the-products-core-running-on-itself)), which are the same objects, not parallel ones.
- [ ] **1.16** Language-agnostic gate configuration (Go, Dart, Python, JS toolchains). **Per language, measure before tiering** — the guard set is adapted to what that toolchain actually hides, never ported verbatim ([ADR-028](decisions.md#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard)).
- [ ] **1.17** Evidence bundle: the run artifact plus what the reviewer found. **Raw form** — retained, and the appeal path behind every decision object. Freshness is a fingerprint: a bundle describing a tree that has moved on is not served.
- [ ] **1.18** Fail-loud end to end: an agent crash or a skipped gate blocks the readiness verdict.

### Fallback path ➕

- [ ] **1.19** Direct PAYG mode against any OpenAI-compatible endpoint, keys in the OS keychain.
- [ ] **1.20** Secret redaction across every log path, verified by test.

- [ ] **1.21** Tests: ACP client against a mock agent, handoff, aggregation, approval expiry, concurrent sessions.

### Module layer & Tier M — the intent hierarchy

> Spec: [`module-layer.md`](../.rla/specs/module-layer.md). **M2 is why this layer exists** —
> it is the only mechanism that fires while every spec is ratified, every gate is green and
> the code is correct, to report that *part of what was wanted was never built*.

- [ ] **1.28** Module artifact in `.rla/modules/<slug>.md`: `MOD-` id, why, `K01…Kn` criteria, **Included/Excluded** scope, projected ①. Hygiene gate extended.
- [ ] **1.29** ⓪ ratification lifecycle: `rla module new|list|ratify`; draft → ratified → closed.
- [ ] **1.30** **Write-deny:** `.rla/modules/**` unwritable by the loop — deny rule + `PreToolUse` hook exiting non-zero, **with a canary proving the block**.
- [ ] **1.31** Business-language check: a criterion naming an endpoint, table, class or library is rejected at draft time with the term quoted.
- [ ] **1.32** Observability **and completeness** check: the verifier must state what a person would look at, *and* attempt to name a criterion the draft is missing. Unobservable criteria never reach ⓪; the completeness attempt is surfaced there. **The only pressure that exists at the top of the hierarchy** ([ADR-026](decisions.md#adr-026--the-hierarchy-relocates-the-blind-spot-it-does-not-remove-it)).
- [ ] **1.33** **M1** — **M1a** declaration validity: the spec names an existing `MOD-K`, and **every invariant names the criterion it derives from** ([ADR-025](decisions.md#adr-025--every-invariant-declares-the-criterion-it-derives-from)). Set membership, no model, **binding**. **M1b** scope: does the subject fall in **Excluded** — judged, **advisory**. M1a carrying the decision is what keeps the highest-frequency intent gate off the expensive tier.
- [ ] **1.34** **M2** — two layers, kept apart. **M2a** declaration coverage: set arithmetic over the `MOD-K` declarations, no model, **binding**. **M2b** declaration honesty: does the declaring spec really serve it — judged, **advisory only, may raise a finding but never clear one.** An uncovered criterion fails the gate and **blocks ②**.
- [ ] **1.35** **M3** — ratified specs whose criterion no longer exists.
- [ ] **1.36** Lower-a-criterion path: returns the module to `draft`, requires written justification, re-enters ⓪. **Never one step** ([P5](../.rla/PRINCIPLES.md#p5--gates-are-immutable-to-the-loop)).
- [ ] **1.37** Risk classification (LOW/MEDIUM/HIGH) and **conditional ①**: auto-ratify on M1 green + **every invariant's parent criterion declared** + LOW/MEDIUM; escalate otherwise **with the reason recorded and displayed**. Early escalation rates are expected to be high — **the escalation log is the corpus the classifier is built from** ([ADR-025](decisions.md#adr-025--every-invariant-declares-the-criterion-it-derives-from)), and the rate falling is the measurement of whether it is definable at all.
- [ ] **1.38** ①-budget projection at ⓪; interrupt counter — per module, auto-approval ratio, time away from the desk. **The limits of 6 ① and 8 total interrupts are working hypotheses, not measurements** — they come from no data and are configurable, revisited against [0.18.6](#p0--scaffolding-gates--de-risking) and the trial.
- [ ] **1.39** **Verifier ≠ producer** enforced at ladder entry: identical T1/T2 models → T2 `COULD NOT VERIFY`, ladder halts. Canaried.

### Decision objects — the translation layer

> ⚠️ **The phase's hardest problem, and the product's principal engineering work**
> ([ADR-021](decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar)).
> Everything above produces evidence. This turns evidence into a decision — and if it
> fails, the unattended default in [ADR-015](decisions.md#adr-015--the-humans-default-state-is-non-involvement)
> is not safe to ship. Spec: [`decision-object.md`](../.rla/specs/decision-object.md).

- [ ] **1.22** Decision object model + emission: exactly one per blocking condition, deduplicated per `(feature, condition)`; zero is a defect.
- [ ] **1.23** Invariant attribution — cite the violated `SPEC-…` id, or declare `none — unspecified behaviour` as a backward-fidelity finding. Never the nearest plausible id.
- [ ] **1.24** One-sentence explanation in the spec's domain vocabulary, ≤ 200 chars; toolchain detail stays in the raw bundle.
- [ ] **1.25** Option construction: 2–3 always, with the standing *"accept and record why"* alternative injected when only one action was derived.
- [ ] **1.26** Cost annotation — time, tokens, blast radius, invariant left broken. **`unknown` where unmeasured; a fabricated number is the defect.**
- [ ] **1.27** Interrupt discipline: only a decision object notifies. Idempotent answers bound to `(object id, nonce)`; redaction over every field.

**Done when**

- Two different agents, two different subscriptions and two tiers run T1 and T2 on one task — and configuring them to the **same** model halts the ladder instead of passing it.
- A deliberately incomplete implementation is caught — by the gates, by the reviewer, or both.
- The reviewer's findings and the gate evidence arrive as one bundle — **and the reviewer ran no tools of its own**, asserted by test on the tool-call log.
- A draft spec or module blocks the readiness verdict.
- Approvals from several agents queue in one place and each resolves exactly once.
- No agent's API key is ever visible to the orchestrator (asserted by test).
- **A module with every spec ratified and every gate green still fails M2** when one acceptance criterion is served by nothing — and ② stays blocked.
- The loop cannot write a module document, **proved by canary rather than asserted**.
- Most specs ratify without reaching a human; the ones that escalate carry a stated reason.
- **A red gate emits a decision object the maintainer can act on without opening the repository**, and a full green run emits nothing at all.

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

**Goal:** the control room in a pocket — **three tabs, none of them a chat**
([ADR-021](decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar)).
**Estimate:** 4–5 weeks · **Depends on:** P2, and on **P1.22–1.39** — the phone renders
decision objects and module state; there is nothing to build here until they exist.

Riverpod · `go_router` · `drift` · Material 3 light/dark. Reference design:
[`design/mobile-v2.html`](../design/mobile-v2.html) — the items below are the behaviour
it has to satisfy, and its light/dark token set is the starting palette.

**Foundations**

- [ ] **3.1 — Skeleton:** feature-first layout, router with pairing guard, theme, L10n (TR + EN, keys only), `Failure` mapping.
- [ ] **3.2 — Network & crypto:** WSS wrapper with reconnect; envelope seal/open matching Go byte for byte; protocol version check; `drift` cache; **explicit host-unreachable state** ([ADR-009](decisions.md#adr-009--the-cli-host-must-be-reachable)).
- [ ] **3.3 — Storage:** device token and session keys in Keychain/Keystore.
- [ ] **3.4 — Pairing:** QR scan, `remlinkagent://pair?token=…`, handshake, key exchange.

**Tab 1 — Queue**

- [ ] **3.5** Pending ⓪, ①, ② and arbitrations — nothing else. **The empty state is the destination:** it says the app can be closed, reports what is still running, and shows **time away from the desk**. It also carries **when the orchestrator last confirmed it was working**, degrades as that ages, and past a threshold **becomes the alarm itself** — the one interrupt allowed with no decision attached ([ADR-024](decisions.md#adr-024--silence-is-attested-not-assumed)).
- [ ] **3.6 — ⓪ module approval:** why it exists, `K01…Kn` criteria, Included/Excluded scope, projected ① against the limit of six. Approve or amend.
- [ ] **3.7 — ① escalated spec:** **the reason it escalated, at the top of the screen** — risk class, new invariant, or M1 red. Which module criteria it serves. Estimated cost.
- [ ] **3.8 — Arbitration:** the invariant, the one-sentence why, 2–3 options with costs, `unknown` rendered honestly, and who raised the objection. Idempotent; resolved on every paired device at once.
- [ ] **3.9 — ② interface test:** criteria ticked one at a time by observation, timestamped. **A half-ticked module does not close** — it returns to the queue.
- [ ] **3.10** Permission requests from any agent, resolved here too — full command, working directory, which agent asked, visible expiry; **fails closed**.

**Tab 2 — Modules**

- [ ] **3.11** Per project: modules, complete, **scope gaps**, awaiting approval. Per module, a criterion ladder showing which are covered. Switch between orchestrator hosts — the maintainer runs 3–5 projects.
- [ ] **3.12 — Module detail:** each criterion mapped to the spec serving it, or **"no spec serves this"** in the alarm colour; M1/M2/M3 status. The **lower-a-criterion** action is present and deliberately **not** a one-tap resolution — it reopens ⓪ ([P5](../.rla/PRINCIPLES.md#p5--gates-are-immutable-to-the-loop)).
- [ ] **3.13 — Feature detail:** the T0–T4 ladder, last evidence line, and the escape hatch — full reasoning trace and raw agent stream, **complete and unrestricted**, one tap away and freely closed. Off by default; never absent.

**Tab 3 — Settings**

- [ ] **3.14** Agent assignment per tier with the per-module override and its written justification; loop method (M, T0–T4) with caps; auto-approval conditions; budget ceiling.
- [ ] **3.15** Provider keys read **"on device"** — a commitment with no UI path that could make it false ([P1](../.rla/PRINCIPLES.md#p1--zero-touch-ai)). Visibility toggles for reasoning and raw stream, both off by default.
- [ ] **3.16 — Interrupt counter:** interrupts per module by kind, auto-approval ratio, time away from the desk, and the "what should page me" thresholds.

**Command bar & platform**

- [ ] **3.17 — Command bar:** context-aware at the bottom of every screen, scoped to project / module / feature. **Asynchronous** — confirms *"sent, you can go."* **Input conversation, output structure:** the reply returns as a module draft, a spec card or an optioned decision, never a wall of text.
- [ ] **3.18 — Push:** APNs + FCM, categories, deep links, content-free payloads rendered on device. **Only a pending checkpoint may generate one.**
- [ ] **3.19 — Platform:** iOS `Info.plist`, Android manifest, permission rationale UI.
- [ ] **3.20 — Tests:** unit, widget, crypto round-trip against Go fixtures, `integration_test` E2E against a mock relay; **assertions that a full green run produces zero notifications** and that a half-ticked module cannot close.

**Done when**

- A module is drafted from a sentence typed into the command bar, and approved at ⓪ from the phone.
- A red gate is resolved from a phone, away from the desk, **without opening the repository**.
- Most specs never appear; the ones that do explain why they did.
- A full green run produces an empty queue, silently, and that reads as success rather than as a broken connection.
- **M2 red visibly blocks ②**, and lowering the criterion demonstrably costs more than one tap.
- The raw trace is reachable in two taps and is complete when opened.
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

## The trial — when this stops being a project

> Shipping P4 does not make this a product. **Being used does.**

The maintainer runs several real projects. The system will be turned on those, and the question it has to answer is the one it was built for:

> **Did it reduce time at the computer, improve what came out, and make the work easier to manage?**

If yes, this is a product. If no, it is a well-engineered project that did not deliver, and saying so is cheaper than the alternative.

### The trial is staged, and the first stage is at the end of P1

Waiting until P4 to find out puts the verdict eight to twelve months away, and it bundles two independent claims into one test.

They separate cleanly:

| Claim | Needs | Testable after |
| :--- | :--- | :--- |
| **"It produces better software with fewer interruptions"** | Orchestrator + module layer + gates, on a desktop. No phone, no relay. | **P1** |
| **"It removes the constraint that work stops when I leave the desk"** | Relay + mobile client | P3 |

**So P1 ships to an audience of one and gets used.** Not demoed — used, on a real project, as the way work actually gets done. Every number in the table below except *time away from the desk* is already measurable at that point, and every falsification condition except the last can already fire.

This is the same argument as [P0.18](#p0--scaffolding-gates--de-risking), one level up: learn the expensive thing early, on the cheapest artifact that can teach it. If the method does not hold at P1, P2–P4 are the wrong six months.

### The thresholds are set before the trial, not after

Every number below is **blank on purpose.** They are the maintainer's to fill, and they must be filled **before** the trial starts. A threshold chosen after seeing the result is not a threshold, it is a rationalisation — and this is the same rule the product applies to everyone else at ⓪: agree the target, then do the work.

The instrumentation already exists — it is [3.16](#p3--mobile-client), shipped as a feature rather than added as telemetry.

| Claim | Measured by | Threshold |
| :--- | :--- | :--- |
| **Less time at the computer** | Time away from the desk, per week, against a baseline recorded before the trial | `___` |
| **Fewer interruptions per unit of work** | Interrupts per module, by kind (⓪ / ① / ② / arbitration) | `___` |
| **The loop is trusted** | Auto-ratification ratio — how many specs never reached a human | `___` |
| **Intent survives** | Modules closed at ② **without reopening**, as a share of modules closed | `___` |
| **The layer earns its cost** | Times M2 caught a criterion that an all-green build would otherwise have shipped | `___` |

Record the baseline first. Without a before, the after means nothing.

### What would falsify it

Stated in advance, for the same reason:

- **Interrupts per module climb rather than fall.** Then this is a spec-reading job with extra steps, and [ADR-020](decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional)'s conditional ① did not work.
- **M2 never fires on real work.** Then the module layer is overhead and its central justification ([ADR-018](decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy)) was wrong. This is checked once by hand at [0.18](#p0--scaffolding-gates--de-risking) precisely so it is not first discovered here.
- **Modules reopen repeatedly after ②.** Then the criteria were not observable and ⓪ is not doing its job — the failure is at the top of the hierarchy, not in the gates.
- **The maintainer stops using it and returns to the manual method.** The single most honest signal available, and the one that needs no metric.

A failed trial is not a failed project. It converts the M-track's practice into a documented method and `scripts/gate` into a working tool, both of which stand alone. What it ends is the claim that the automation was worth building.

---

## 4. Risks

| Risk | Impact | Likelihood | Mitigation |
| :--- | :--- | :--- | :--- |
| **Cross-verification adds no signal** — two models share a blind spot | **Critical** | Medium | Manual runs already found real defects; deterministic gates carry weight where judgement cannot; measure catch rate from day one |
| **The translation layer produces confident nonsense** — a wrong summary is acted on faster than raw output | **Critical** | Medium | Raw bundle always one tap away as the appeal path; every claim cites a `SPEC-…` id; fabricated costs forbidden and `unknown` is a valid answer; decisions contradicted by their own raw evidence are a tracked defect class |
| **Silence hides a broken loop** — an unattended default manufactures confidence if the gates are weak | High | Medium | Canaries; gates compiled not configured; `COULD NOT VERIFY` blocks readiness; a run that emits nothing must still show *when it last ran* |
| **Auto-ratification silently approves a misclassified HIGH spec** | **Critical** | Medium | Classification recorded and reviewable per spec; per-module override disables auto-approval wholesale; the auto-approval ratio is displayed, so a run of silent approvals is visible. **The classifier is an open question that blocks ratification of [`module-layer.md`](../.rla/specs/module-layer.md)** |
| **Business-language criteria are written unfalsifiably** — *"the user understands why"* | High | High | The adversarial verifier must state what a person would look at; criteria it cannot ground are rejected before ⓪ (1.32) |
| **M2 accepts a spec that only sounds like it serves a criterion** | High | Medium | The gate is split: **M2a** is deterministic set arithmetic over declarations and is what binds; **M2b**'s judgement is advisory and can only add findings, never clear them. The mapping is then shown to the human at ②, criterion by criterion |
| **The trial has no pre-agreed threshold** and "it kind of works" becomes the verdict | High | **High** | Thresholds and the baseline are recorded **before** the trial starts; falsification conditions written down with them ([the trial](#the-trial--when-this-stops-being-a-project)) |
| **ACP cannot hold an approval for a phone round trip** | High | Medium | P0.17.1 answers this before P1 |
| ACP protocol churn across agents | Medium | High | Shared standard with a registry, not one vendor's interface; capability negotiation; pin tested versions |
| Provider terms change | High | Medium | Delegated path only; re-verified each release; no impersonation ever |
| Agent install friction | Medium | High | `rla setup` orchestrates their official installers |
| Incumbents add cross-verification | High | Low | A vendor cannot credibly arbitrate between models ([ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent)) |
| **Scope creep** | High | High | Adding to the MVP requires an ADR |

> **The critical risk is the first one.** Everything else is engineering. If a second model does not reliably find what the first missed, no architecture saves the product — so catch rate is instrumented from P1 and reviewed at every phase gate.

---

## M1 — Managed Cloud, subscriptions & Team

> Not a feature — the mechanism that funds the work. Self-hosting stays free and complete.

**Prerequisite:** P0–P4 shipped **and a passed [trial](#the-trial--when-this-stops-being-a-project)**. **Estimate:** 5–6 weeks

**Settled — the shape:**

| | |
| :--- | :--- |
| **One person** | **Free, full parity, permanently.** Self-hosted or managed. |
| **A team** | Paid. What it adds is **multi-person identity** — who approved which checkpoint, and a delivery record that carries more than one signature. |
| **Self-host** | Free and complete, always. |

**Not settled — the price.** Deliberately. There is no product yet, no trial result and no conversation with a real buyer; a number published now would anchor the wrong value and be hard to raise later. It comes from the first few buyer conversations, not from this document.

**What the paid tier is not.** It is not a feature held back from the free one. Everything a single user needs works free and self-hosted forever. What is paid is the part that **inherently requires shared infrastructure** — identity across people, and a record signed by more than one of them.

- [ ] **M1.1** Managed relay: multi-tenant, isolated.
- [ ] **M1.2** Accounts, device management, plan assignment.
- [ ] **M1.3** Billing; renewal, cancellation, webhooks.
- [ ] **M1.4** Subscription website: landing, account area, checkout, docs, status page, privacy.
- [ ] **M1.5** Plan gating. **Self-host: full parity, always.**
- [ ] **M1.6** Team: identity on every ⓪/①/② decision, checkpoint routing by role, handover, shared method profile.
- [ ] **M1.11** **Delivery record** — a module, its criteria, the spec that served each, gate evidence, and who confirmed what at ② and when. Exportable, readable by someone who does not code.
- [ ] **M1.7** ➕ StoreKit 2 + Play Billing with server-side receipt validation.
- [ ] **M1.8** Cost transparency: token reporting per agent and provider.
- [ ] **M1.9** Self-host ↔ cloud migration.
- [ ] **M1.10** Operations: DDoS protection, monitoring, backups, AGPL compliance audit.

---

## 5. Beyond M1

[`vision-roadmap.md`](vision-roadmap.md): deeper verification tiers (mutation, fuzzing, black-box exploration), voice control, Apple Watch and Wear OS, interactive terminal mirroring.

---

## 6. Glossary

**ACP** — Agent Client Protocol; JSON-RPC over stdio between a client and a coding agent. **Module** — the primitive unit: intent in business language, with observable acceptance criteria `K01…Kn` and an explicit Included/Excluded scope. **`MOD-K`** — one acceptance criterion of a module. **Intent hierarchy** — Module → Spec → Feature + invariant → Code, each layer the oracle of the one below. **Oracle** — an approved statement of intent a machine can check the layer below against; a chat log is not one. **Tier M** — the intent gates: **M1** spec↔module fidelity, **M2** an acceptance criterion no spec serves, **M3** orphaned ratified specs. **T0–T4** — the code ladder: instant checks, producer, adversarial verifier, arbiter, periodic sweep. **Checkpoint ⓪** — human approves the module; where critical questioning happens. **Checkpoint ①** — spec approval, *conditional*; auto-ratifies unless HIGH risk, a new invariant, or M1 red. **Checkpoint ②** — human observes the module's criteria, once, one at a time. **Arbitration** — a gate went red and the loop cannot resolve it alone. **Decision object** — a red gate rendered as violated invariant, one-sentence why, and 2–3 priced options. **Evidence bundle** — everything a run can prove, in raw form; the appeal path behind a decision object. **Cross-verification** — a different vendor's model attempting to disprove work; T2 may never be T1. **`COULD NOT VERIFY`** — a gate did not run; never a pass. **Delegated path** — a listed agent calls the provider under its own identity. **BYOK** — the user's own key or subscription, never ours.

---

*Why things are the way they are: [`decisions.md`](decisions.md). How the work gets done: [`development-loop.md`](development-loop.md).*
