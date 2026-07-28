# 🔗📱 RemLinkAgent

### Agentic Project Management — a control room, not another chat window

Nobody staffs a control room by staring at a screen for eight hours. Instruments watch, thresholds are agreed in advance, and an alarm summons a human who arrives with authority and context.

You approve a **module** — what you want, in your own words, with criteria anyone can observe. Agents write the specs and the code. Deterministic gates and a *different vendor's* model produce the evidence. Your default state is **not involved**. When something needs you, your phone says which rule broke, why in one sentence, and what your two or three options cost.

**An empty queue is the success state. The app tells you to close it.**

[![License: AGPL v3](https://img.shields.io/badge/core-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Mobile: Apache 2.0](https://img.shields.io/badge/mobile-Apache--2.0-green.svg)](mobile/LICENSE)
[![Status](https://img.shields.io/badge/status-pre--alpha%20·%20P0-orange.svg)](docs/roadmap.md)

---

## 🚦 Status — read this first

**There is no usable release yet.** The repository contains the architecture, the roadmap, the governance setup, a build skeleton, and a working implementation of the verification system itself. Phase **P1** is where the orchestrator begins.

| What | State |
| :--- | :--- |
| Architecture, protocol, threat model, decisions | ✅ [`docs/`](docs/) |
| **Verification gates** — modes, guards, evidence artifact, canaries | ✅ **Working** — `go run ./scripts/gate verify` |
| Repo scaffolding, CI, licence automation | ✅ In place |
| Orchestrator, relay, mobile app | 🚧 Not started — [P1–P4](docs/roadmap.md) |

The gates are not a demo. This repository is built with them ([ADR-011](docs/decisions.md#adr-011--the-project-is-built-with-its-own-loop)).

---

## 🎯 The problem

We used to spend hours at a computer **writing** code. Now we spend hours at a computer **watching** something else write it. The clock did not move; the agency did.

Both ends of the current spectrum are wrong. The developer stays at the centre of the work because something has to ship — that does not scale past one person's day. The product owner hands everything to the agent and looks only at the output — that has no way of finding out whether the output is right.

And the reason nobody can safely stop watching is that the agent's own "done" is not checkable. A coding agent finishes a feature and reports success. Everything is green. But green means *"it followed the rules"* — not *"it did the right thing."*

This was measured, not assumed. A project one model declared complete was put through a verification pass driven by a **different** model against its spec. That pass found missing translations, tests that were never written, a coverage regression, and real bugs.

## 🧱 Intent is a hierarchy — and each layer judges the one below

The primitive unit is not a message and not a feature. It is a **module**.

| Layer | Written in | Who | You see it |
| :--- | :--- | :--- | :--- |
| **Module** | **Business language.** Criteria anyone can observe. | **You** | **⓪ always** |
| **Spec** | Technical language. `SPEC-…` ids, invariants. | Agents | **① only when it matters** |
| **Feature + invariant** | Code citing a `SPEC-…` id | Agents | never |
| **Code** | — | Agents | never |

A module says *"a dealer with insufficient balance cannot transact and understands why."* It may **not** name an endpoint, a table or a class — that is a spec, one layer down, and writing it at module level destroys the oracle.

**This is where your judgement actually works.** Asked to ratify an idempotency invariant, you take it on trust. Asked whether a dealer should see their own balance, you know instantly.

```
        ⓪  YOU approve a module ─── criteria · scope: included / excluded
                       ▼
        M1  every spec must name the criterion it serves ── and stay in scope
                       ▼
        ①  spec — auto-ratified unless HIGH risk, a new rule, or M1 red
                       ▼            (and if it reaches you, it says why)
        T1  cheap model implements ─── T0 lint/types, instant, no tokens
                       ▼
        T2  a DIFFERENT model tries to break it ── same model = gate void
                       ▼
        T3  arbiter, only on contradiction
                       ▼
        M2  is any criterion served by NO spec? ──── blocks ② while red
                       ▼
        ②  YOU observe the criteria, one at a time. Half-ticked won't close.
```

Each role runs a **real, existing agent** — Qwen Code, Kimi Code, OpenCode — over the [Agent Client Protocol](https://agentclientprotocol.com/). We implement no agent, no model router, no provider integration ([ADR-014](docs/decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)). Each uses **your** account under **its own** identity. Keys never reach us.

Every gate runs against a selective regression cache, and **`git commit` invalidates nothing** — committing does not change what a gate read.

## 🕳️ M2 — the gate that catches what nothing else can

Every spec ratified. Every gate green. The code entirely correct. And a part of what you actually asked for was **never built** — because nobody ever wrote a spec for it.

No other gate can see this, because every other gate compares code against a spec, and here the defect *is* the missing spec. M2 compares the specs against your module's criteria and names the one nothing serves ([ADR-019](docs/decisions.md#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write)). The check that binds is **set arithmetic, not model judgement** — a judged gate here may raise a finding, never clear one ([ADR-023](docs/decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)).

**Said honestly: this relocates the blind spot rather than removing it.** Nothing is the oracle of the module itself, so a requirement you never turned into a criterion is still invisible. What changes is the price — a missing criterion costs a conversation at ⓪; a missing spec costs a rebuild. The verifier is pointed at the module draft to push on that, and it is pressure, not proof ([ADR-026](docs/decisions.md#adr-026--the-hierarchy-relocates-the-blind-spot-it-does-not-remove-it)).

While M2 is red, **② cannot start.** Interface-testing a knowingly incomplete product is not a test.

And you may lower a criterion to clear it — sometimes the goal really was too ambitious. But **never in one tap.** The module returns to draft, you write why, and it comes back through ⓪. Lowering a target to turn red green is the human-hand version of the thing the loop is forbidden to do, and a phone queue between meetings is the most dangerous place in the product for it.

## 🗣️ The hard part: making a red gate answerable from a phone

Raw gate output is unreadable on a train. So the orchestrator does not send it. It sends a decision:

> **Violated:** *the same transaction can never be charged twice*
> **Why:** the second call with the same `session_id` issues a second receipt.
> **Options:** ① atomic lock — ~15 min, no schema change · ② relax the spec — touches the module boundary, reopens ⓪ · ③ leave it to me — stays queued, I'll look at my desk.

**That translation is the product's real engineering work** ([ADR-021](docs/decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar)). It happens in the orchestrator, not the app.

## 💰 Cheap models grind — a different model breaks — an expensive one arbitrates

| Tier | Job | Model |
| :--- | :--- | :--- |
| **M** | Intent gates — reading meaning | Expensive |
| **T0** | Format, lint, types | **None.** Deterministic, free |
| **T1** | Producer | Cheapest capable, flat-rate subscription |
| **T2** | **Adversarial verifier** | **Never the same model as T1** |
| **T3** | Arbiter, on contradiction only | Expensive, rare |
| **T4** | Weekly sweep — CVEs, dead code | Cheap, off by default |

Cross-verification costs 2–3× the tokens of a single pass. One price tier everywhere fails both ways: expensive everywhere is unaffordable; cheap everywhere puts the weakest judgement exactly where your decision rests on it.

**And T2 may never be T1.** Configure them the same and the gate is **void and the ladder halts** — it does not quietly pass. A model grading its own work is the failure this product exists to correct; showing a green badge for it would be worse than running nothing ([ADR-022](docs/decisions.md#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer)).

The cheap end of the market is therefore **not a positioning choice. It is what makes the technique exist at all.**

## 🧠 Why a model vendor cannot build this

The honest output of cross-verification is sometimes *"our own model got this wrong."*

Vendors ship model *pickers* and steer you toward their own — one gives a 1.5× quota bonus for using its first-party client. **Only a party that sells no model can credibly arbitrate between models.**

---

## 📱 Three tabs — none of them is a chat

Reference design: [`design/mobile-v2.html`](design/mobile-v2.html) — open it in a browser.

| Tab | What it holds | Empty means |
| :--- | :--- | :--- |
| **Queue** | Pending ⓪, ①, ② and arbitrations — nothing else | ✅ **"Nothing is waiting for you."** And it says you can close the app — with a timestamp proving it still knows. |
| **Modules** | Per project: modules, complete, **scope gaps**, awaiting approval — and which criterion is met by which spec | Nothing in flight |
| **Settings** | Agent assignment, loop method, auto-approval conditions, budget, visibility, **interrupt counter** | — |

*"One module has a scope gap"* is a number almost nothing can produce, because it needs an **oracle** — a statement of intent above the spec layer. A chat log is not one. A ratified module with observable criteria is.

**Chat is not removed, and it is not a tab.** It is a context-aware command bar at the bottom of every screen, and it is **asynchronous** — you send and you leave; it answers *"sent, you can go."* Input is conversation; **output is structure**: a module draft, a spec card, an optioned decision. Never a wall of text.

### What you will not see

Every element has to pass one test: **does seeing this change the decision I am about to make?**

Failing it: the token counter, the agent's reasoning stream, the diff by default, the live log. Each is engaging; none changes an answer.

**Two exceptions, and no more.** Budget — it decides whether to let the loop keep grinding. And the **heartbeat** — because a dead orchestrator and a quiet one render the same empty screen, and an app still reporting *"2 modules running"* nine hours after the process died is a lie shaped like good news. Stale silence becomes the alarm itself ([ADR-024](docs/decisions.md#adr-024--silence-is-attested-not-assumed)).

**Nothing is hidden.** The full reasoning trace and the raw agent conversation are available **in their entirety, without restriction**, one tap away, freely opened and freely closed. The objection was to making them the *centre* — never to their existence ([P11](.rla/PRINCIPLES.md#p11--nothing-is-shown-that-does-not-change-a-decision)).

### The metric

Not time in the app. **Time away from the desk.** The interrupt counter ships as a feature: interrupts per module, auto-approval ratio, hours spent elsewhere. Above roughly eight interrupts per module, the product tells you what that means — *the modules are too large, or the criteria too vague.*

> The system cannot earn the right to interrupt before it has earned the right to correctly **not** interrupt.

## ⚠️ What this is not

- **Not another coding agent.** Use the ones you already have. We orchestrate them.
- **Not a chat client for your agent.** Several of those exist. This is the opposite bet.
- **Not a terminal mirror.** Interactive TUIs are out of scope — `ssh` and `tmux` already do that well.
- **Not a model provider.** Your keys, your subscriptions, your bill — paid directly to your provider.
- **Not a cloud runner.** Our server is a control-channel relay. It never proxies a model call.
- **Not project management with dates.** No Gantt, no burndown, no estimates in time. The only budget is a **count of decisions**.

**One requirement worth knowing:** the orchestrator runs on your machine, so that machine must be on. If it is not, the app shows the last known state and says so — it will not silently spin. For always-on use, run it on a machine that is always up ([headless deployment](docs/architecture.md#cli-host-availability)).

Everything crossing the relay is **end-to-end encrypted**; it handles ciphertext and routing metadata only ([ADR-004](docs/decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)). In Settings, provider keys read **"on device"** — that line is a commitment, not a preference, and there is no UI path that could make it false. Self-hosting removes us from the picture entirely.

---

## 🧩 Models and agents

Any [ACP](https://agentclientprotocol.com/) agent can fill any role. Provider-agnostic agents such as **OpenCode** reach dozens of models on their own — including providers that ship no CLI.

| Agent | Models | Licence |
| :--- | :--- | :--- |
| **Qwen Code** | Qwen | Apache-2.0 |
| **Kimi Code** | Kimi / Moonshot | Apache-2.0 |
| **OpenCode** | provider-agnostic — GLM, DeepSeek, MiniMax, … | Open source |

A direct pay-as-you-go path against any OpenAI-compatible endpoint remains available as a fallback ([ADR-012](docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).

---

## ⚡ Building from source

No release yet. What works today:

```bash
git clone https://github.com/burakhalefoglu/RemLinkAgent.git
cd RemLinkAgent

make cli && make server   # → ./bin/rla, ./bin/rla-server
make verify               # the verification system, running on itself
make canary               # proves each gate still detects breakage
```

`make help` lists every target. Windows without GNU make: `.\scripts\make.ps1 <target>`.
Prerequisites and setup: [`CONTRIBUTING.md`](CONTRIBUTING.md).

---

## 🗺️ Roadmap

| Phase | Scope | State |
| :--- | :--- | :--- |
| **P0** | Scaffolding, CI, verification gates, ACP spike | 🚧 In progress |
| **P1** | Orchestrator: ACP client, role assignment, coder → gates → reviewer handoff | 📋 Next |
| **P2** | Relay: WSS, pairing, E2E, JetStream sync, push | 📋 |
| **P3** | Mobile: spec ratification, evidence review, approvals | 📋 |
| **P4** | Signed binaries, store submission, launch | 📋 |
| **M1** | Managed cloud and a team tier — self-host stays free and complete | 📋 |
| **X** | Voice, smartwatches, deeper verification tiers | 💭 |

Detail and the assumptions behind the estimates: [`roadmap.md`](docs/roadmap.md).

---

## 📜 Licence

| Component | Licence |
| :--- | :--- |
| Orchestrator + relay (`cmd/`, `internal/`, `deploy/`, `scripts/`) | [AGPL-3.0-or-later](LICENSE) |
| Mobile client (`mobile/`) | [Apache-2.0](mobile/LICENSE) |

The split is deliberate — GPL-family terms conflict with App Store distribution ([ADR-002](docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile)). Contributions require a [CLA](CLA.md) ([why](docs/decisions.md#adr-003--cla-required)).

## 🤝 Contributing

Early, and the architecture is still soft in places. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md), then [`development-loop.md`](docs/development-loop.md) — this repository is built with the method it ships.

Security issues: **do not** open a public issue — see [`SECURITY.md`](SECURITY.md).

---

*"Everything is green" means it follows the rules. It does not mean it does the right thing.*
