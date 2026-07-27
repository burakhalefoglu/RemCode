# 🔗📱 RemLinkAgent

### One model writes it. A different one proves it. You approve.

Coding agents tell you they are finished. Often they are not — a translation is missing, a test asserts nothing, coverage dropped, a bug slipped through. The model that wrote the code is the worst possible judge of whether the code is right.

RemLinkAgent makes a **different model** verify the work against a specification you ratified, runs deterministic gates over the result, and shows you the evidence. From your desk or from your phone.

[![License: AGPL v3](https://img.shields.io/badge/core-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Mobile: Apache 2.0](https://img.shields.io/badge/mobile-Apache--2.0-green.svg)](mobile/LICENSE)
[![Status](https://img.shields.io/badge/status-pre--alpha%20·%20P0-orange.svg)](docs/roadmap.md)

---

## 🚦 Status — read this first

**There is no usable release yet.** The repository contains the architecture, the roadmap, the governance setup, a build skeleton, and a working implementation of the verification system itself. Phase **P1** is where the orchestrator begins.

| What | State |
| :--- | :--- |
| Architecture, protocol, threat model, decisions | ✅ [`docs/`](docs/) |
| **Verification gates** — tiers, spec fidelity, canaries | ✅ **Working** — `go run ./scripts/gate verify` |
| Repo scaffolding, CI, licence automation | ✅ In place |
| Orchestrator, relay, mobile app | 🚧 Not started — [P1–P4](docs/roadmap.md) |

The gates are not a demo. This repository is built with them ([ADR-011](docs/decisions.md#adr-011--the-project-is-built-with-its-own-loop)).

---

## 🎯 The problem

A coding agent finishes a feature and reports success. Everything is green. But green means *"it followed the rules"* — not *"it did the right thing."*

This was measured, not assumed. A project one model declared complete was put through a verification pass driven by a **different** model against its spec. That pass found missing translations, tests that were never written, a coverage regression, and real bugs.

The gap is real, and a second model finds it.

## 🔀 How it works

```
        ┌──────────────────────────────────────────────┐
        │  ①  You write a spec and ratify it           │
        └────────────────────┬─────────────────────────┘
                             ▼
        ┌──────────────────────────────────────────────┐
        │  Coder      — implements it                  │
        │               (e.g. Qwen Code, your Qwen sub)│
        └────────────────────┬─────────────────────────┘
                             ▼
        ┌──────────────────────────────────────────────┐
        │  Gates      — lint · tests · coverage ·      │
        │               spec fidelity · fake-green     │
        │               deterministic, no tokens        │
        └────────────────────┬─────────────────────────┘
                             ▼
        ┌──────────────────────────────────────────────┐
        │  Reviewer   — a DIFFERENT model tries to     │
        │               disprove it against the spec   │
        │               (e.g. Kimi Code, your Kimi sub)│
        └────────────────────┬─────────────────────────┘
                             ▼
        ┌──────────────────────────────────────────────┐
        │  ②  You read the evidence and confirm        │
        └──────────────────────────────────────────────┘
```

Each role runs a **real, existing agent** — Qwen Code, Kimi Code, OpenCode — driven over the [Agent Client Protocol](https://agentclientprotocol.com/). We do not implement an agent, a model router, or a provider integration ([ADR-014](docs/decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)).

Each agent uses **your** account and **your** subscription, under its own identity. Keys never reach us.

## 💰 Why this only works with cheap models

Cross-verification costs 2–3× the tokens of a single pass. At Western API prices that is prohibitive. On a flat-rate coding subscription it is free at the margin — you already paid.

That is why the cheap end of the market is not a marketing angle here. **It is what makes the technique affordable at all.**

## 🧠 Why a model vendor cannot build this

The honest output of cross-verification is sometimes *"our own model got this wrong."*

Vendors ship model *pickers* and steer you toward their own — one provider gives a 1.5× quota bonus for using its first-party client. **Only a party that sells no model can credibly arbitrate between models.**

---

## ⚠️ What this is not

- **Not another coding agent.** Use the ones you already have. We orchestrate them.
- **Not a terminal mirror.** Interactive TUIs are out of scope — `ssh` and `tmux` already do that well.
- **Not a model provider.** Your keys, your subscriptions, your bill — paid directly to your provider.
- **Not a cloud runner.** Everything executes on your machine, in your repository.

## 📱 Where the phone comes in

Checkpoint ② is asynchronous by nature: read the evidence, approve or redirect. That is exactly what a phone is good for — and it removes the constraint that when you leave your desk, work stops.

Every command an agent proposes is approved by you before it runs. Everything crossing our relay is **end-to-end encrypted**; the relay handles ciphertext and routing metadata only ([ADR-004](docs/decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)). Self-hosting removes us from the picture entirely.

**One requirement worth knowing:** the orchestrator runs on your machine, so that machine must be on. If it is not, the app shows the last known state and says so — it will not silently spin. For always-on use, run it on a machine that is always up ([headless deployment](docs/architecture.md#cli-host-availability)).

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
| **M1** | Managed Cloud ($5/mo) — self-host stays free and complete | 📋 |
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
