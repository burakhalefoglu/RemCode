# 🔗📱 RemLinkAgent

### An AI coding agent that runs on your machine and is driven from your phone

> **"Your terminal, in your pocket. Your models, your rules."**

RemLinkAgent pairs your phone with a coding agent running on your own machine. You bring your own API keys, the agent works in your real repository, and you review and approve what it does from your phone — including every command it wants to run.

The agent speaks to **any OpenAI-compatible provider**, with first-class support for **Z.AI, Qwen and Kimi**. AI requests go straight from your machine to your provider. Everything that passes through our relay is **end-to-end encrypted**.

[![License: AGPL v3](https://img.shields.io/badge/core-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Mobile: Apache 2.0](https://img.shields.io/badge/mobile-Apache--2.0-green.svg)](mobile/LICENSE)
[![Status](https://img.shields.io/badge/status-pre--alpha%20·%20P0-orange.svg)](docs/roadmap.md)

---

## 🚦 Project status — read this first

**There is no usable release yet.** The repository currently contains the architecture, the roadmap, the governance setup and a build skeleton. Phase **P0** is where implementation begins.

| What | State |
| :--- | :--- |
| Architecture, protocol, threat model, roadmap | ✅ Written — see [`docs/`](docs/) |
| Repo scaffolding, CI, licence automation | ✅ In place |
| Go build skeleton (`rla`, `rla-server`) | ✅ Builds — prints version, nothing more |
| CLI agent, relay server, Flutter app | 🚧 Not started — [P0–P4](docs/roadmap.md) |

If you are here to use it, [watch the repo](https://github.com/burakhalefoglu/RemLinkAgent/subscription) for the first release. If you are here to build it, start with [`CONTRIBUTING.md`](CONTRIBUTING.md).

---

## 🎯 What this actually is

An **agent**, not a remote terminal. It reads and writes files in your project, runs commands, and iterates — the same shape of work as a desktop coding agent, with the review surface moved to your phone.

**What it does**

- Reads and edits files in the repository on your machine
- Runs commands and feeds the output back into its own loop
- Verifies its own edits — format, lint, type-check after every change ([Tier 0](docs/loop-engineering.md))
- Pauses and asks your phone before running anything destructive
- Lets you switch models mid-conversation without losing context
- Keeps working while your phone is asleep, offline, or the app is closed

**What it is not**

- Not a terminal emulator or screen mirror. Interactive TUIs (`vim`, `htop`, `tmux`) are out of scope for now — see [X5](docs/vision-roadmap.md#x5--interactive-terminal-mirroring-pty). For those, `ssh` and `tmux` already work well.
- Not a cloud service that runs your code. The agent runs on **your** hardware.
- Not a model provider. You bring your own keys and pay your provider directly.

---

## ⚠️ The one requirement worth knowing up front

**Your machine has to be on, with the daemon running.** The agent lives there — that is what makes BYOK and zero-touch AI possible. If your laptop is closed, the app shows you the last known state and tells you the host is unreachable. It will not silently spin.

If you want it always available, run the daemon on a machine that is always up — a home server, a small VPS, a work desktop. See [headless deployment](docs/architecture.md#cli-host-availability).

We would rather you learn this here than on a train.

---

## 🔐 What our relay can and cannot see

The relay exists to move events between your machine and your phone when they cannot reach each other directly. Two separate guarantees:

**AI traffic never touches it.** Requests go from your machine straight to your provider. Your API keys are stored in your OS keychain and never leave the host. The relay has no code path that could carry them.

**Everything else is end-to-end encrypted.** Chat messages, model output, proposed commands, command results — all encrypted on the device and on the CLI host. The relay stores and forwards ciphertext. Keys are exchanged during QR pairing and never uploaded.

**What the relay still observes**, because it must in order to route:

- Which devices are connected, and when
- Message sizes, counts and timing
- Event types (message vs. approval request) — needed for push notification routing

That list is complete and is what [`privacy.md`](docs/privacy.md) commits to. Push notifications carry only an event type and an id; the text you see on your lock screen is rendered on your device after decryption.

Self-hosting removes the relay operator from the picture entirely. It is free, unpaywalled, and always will be.

---

## 🧩 Bring your own provider

Any OpenAI-compatible endpoint works. These three are first-class — presets, tested tool-calling adapters, cost metadata:

| Provider | Example models |
| :--- | :--- |
| 🐉 **Z.AI** | `glm-4.6` |
| ⚡ **Qwen** | `qwen-2.5-coder` |
| 🌙 **Kimi** | `moonshot-v1` |

Others — DeepSeek, OpenRouter, Together, a local Ollama, OpenAI itself — work through the same configuration path. Adding one is a config entry plus a [contract test](docs/protocol.md#provider-contract-tests), not a code change.

**Hot-swapping** is the feature we care most about: change model mid-conversation and keep the context. Plan with a large-context model, implement with a cheap fast one, review with a strict one — in one session, without re-explaining anything.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│  PHONE (Flutter)                                        │
│  Chat · Model switcher · Command approval               │
└──────────────────────▲──────────────────────────────────┘
                       │ WSS · end-to-end encrypted
┌──────────────────────▼──────────────────────────────────┐
│  RELAY (Go)                                             │
│  WSS gateway · NATS JetStream · APNs/FCM                │
│  Sees ciphertext + routing metadata only                │
└──────────────────────▲──────────────────────────────────┘
                       │ WSS · end-to-end encrypted
┌──────────────────────▼──────────────────────────────────┐
│  CLI AGENT (Go) — your machine                          │
│  Agent loop · Tools (file/exec) · Tier 0 · Model router │
└──────────────────────┬──────────────────────────────────┘
                       │ direct, BYOK — never via the relay
                       ▼
             AI provider (Z.AI / Qwen / Kimi / …)
```

| Layer | Technology |
| :--- | :--- |
| CLI agent | Go — single static binary, target < 30 MB RSS |
| Relay server | Go — `net/http` + `gorilla/websocket` |
| Message queue | NATS JetStream — persistent, replayable |
| Database | SQLite |
| Mobile | Flutter — iOS + Android |

Full detail: [`architecture.md`](docs/architecture.md) · [`protocol.md`](docs/protocol.md) · [`threat-model.md`](docs/threat-model.md)

---

## ⚡ Building from source

No release yet. What works today:

```bash
git clone https://github.com/burakhalefoglu/RemLinkAgent.git
cd RemLinkAgent

make cli && make server   # builds ./bin/rla and ./bin/rla-server
./bin/rla version         # the skeleton's only current behaviour

make docker-up            # brings up NATS + the (empty) server
make test lint            # runs green on the skeleton
```

`make help` lists every target. Prerequisites and the full development setup are in [`CONTRIBUTING.md`](CONTRIBUTING.md).

Once P4 lands, installation becomes a single command and the mobile app ships to both stores.

---

## 🗺️ Roadmap

| Phase | Scope | State |
| :--- | :--- | :--- |
| **P0** | Scaffolding, CI, provider capability spike | 🚧 Next |
| **P1** | CLI agent: agent loop, tools, model router, hot-swap, Tier 0 | 📋 |
| **P2** | Relay: WSS, pairing, E2E crypto, JetStream sync, push | 📋 |
| **P3** | Flutter app: pairing, chat, model switcher, approvals | 📋 |
| **P4** | Signed binaries, store submissions, docs, launch | 📋 |
| **M1** | Managed Cloud + subscriptions ($5/mo Pro) — self-host stays free | 📋 |
| **X1–X6** | Role delegation, voice, watches, Loop Engineering tiers 1–4 | 💭 |

Realistic solo-developer estimate for P0–P4: **17–20 weeks**. Detail and the assumptions behind that number: [`roadmap.md`](docs/roadmap.md) · [`vision-roadmap.md`](docs/vision-roadmap.md).

---

## 🏢 Self-hosted vs. managed

**Self-hosted — free, forever, full parity.** No feature is held back. Once P2 lands:

```bash
cd deploy && docker compose up -d
```

**Managed cloud — $5/month Pro, arriving in M1.** You are paying to not run a relay: uptime, push certificates, upgrades. Nothing more. AI costs always go directly to your provider. Because of end-to-end encryption, the hosted relay sees exactly as little as your own would.

---

## 📜 Licence

| Component | Licence | Why |
| :--- | :--- | :--- |
| CLI + server (`cmd/`, `internal/`, `deploy/`) | [AGPL-3.0-or-later](LICENSE) | Closes the SaaS loophole — a modified hosted relay must publish its changes |
| Mobile app (`mobile/`) | [Apache-2.0](mobile/LICENSE) | GPL-family terms conflict with App Store distribution |

The split is deliberate and explained in [ADR-002](docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile). Every source file carries its header; CI enforces it.

Contributions require a [CLA](CLA.md) — the reasoning, including what changed from the earlier "no CLA" position, is in [ADR-003](docs/decisions.md#adr-003--cla-required).

---

## 🤝 Contributing

Early, and the architecture is still soft in places — a good moment to influence it. Start with [`CONTRIBUTING.md`](CONTRIBUTING.md), then [`docs/roadmap.md`](docs/roadmap.md) for what is actually next.

Security issues: **do not** open a public issue — see [`SECURITY.md`](SECURITY.md).

---

*Built for developers who want a capable coding agent on their own hardware, on their own keys, at their own cost.*
