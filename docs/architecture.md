# 🏗️ Architecture

> **Status:** design · **Updated:** 2026-07-27
> Wire formats: [`protocol.md`](protocol.md) · Security: [`threat-model.md`](threat-model.md) · Rationale: [`decisions.md`](decisions.md)

---

## The shape of it

Four layers. The interesting property is how little each one is trusted with.

```
┌────────────────────────────────────────────────────────────────┐
│  MOBILE (Flutter)                               Apache-2.0     │
│  ① Ratify spec · ② Review evidence · Approve commands          │
└──────────────────────────▲─────────────────────────────────────┘
                           │ WSS · sealed envelopes
┌──────────────────────────▼─────────────────────────────────────┐
│  RELAY (Go)                                     AGPL-3.0       │
│  WSS gateway · JetStream · Push · SQLite                       │
│  Holds no keys. Sees ciphertext + routing metadata.             │
└──────────────────────────▲─────────────────────────────────────┘
                           │ WSS · sealed envelopes
┌──────────────────────────▼─────────────────────────────────────┐
│  ORCHESTRATOR (Go) — the user's machine         AGPL-3.0       │
│  ACP client · Role router · Handoff · Gates · Approval queue   │
│  Holds no API keys either.                                      │
└──────────────────────────┬─────────────────────────────────────┘
                           │ ACP (JSON-RPC/stdio, or HTTP+SSE)
        ┌──────────────────┼──────────────────┐
        ▼                  ▼                  ▼
  ┌───────────┐      ┌───────────┐     ┌───────────┐
  │ Qwen Code │      │ Kimi Code │     │ OpenCode  │
  │  CODER    │      │ REVIEWER  │     │ EXPLORER  │
  └─────┬─────┘      └─────┬─────┘     └─────┬─────┘
        │ own key/subscription, own identity  │
        ▼                  ▼                  ▼
     Qwen               Kimi            GLM / DeepSeek / …
```

Three separations do the work:

**We are not an agent.** Roles are filled by agents that already exist, driven over the [Agent Client Protocol](https://agentclientprotocol.com/). No agent loop, no tool framework, no model router here ([ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)).

**We hold no credentials.** Each agent authenticates itself, with the user's own key or subscription, under its own real identity. The orchestrator never sees a key — and neither does the relay ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).

**We sell no model.** Which is what makes a verdict of *"the Coder got this wrong"* credible ([ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent)).

---

## Orchestrator

The product. Everything else is transport and UI.

```
cmd/rla/
internal/
├── acp/          ACP client: stdio + HTTP/SSE transports, session lifecycle
├── roles/        role → agent binding, per-project configuration
├── orchestrate/  the run loop: implement → gates → review → evidence
├── verify/       the gate engine (promoted from scripts/gate)
├── spec/         spec artifacts, ids, fidelity diff, ratification state
├── approve/      unified approval queue across agents
├── crypto/       envelope seal/open, key management
├── transport/    WSS client to the relay, reconnect, sync
└── daemon/       IPC, lifecycle, multi-project supervision
```

### The run loop

```
ratified spec
     ↓
┌──────────────────────────────────────────────────┐
│ CODER agent — implement                          │
│   permission requests → approval queue → phone   │
└────────────────────┬─────────────────────────────┘
                     ▼
┌──────────────────────────────────────────────────┐
│ GATES — deterministic, no tokens                 │
│   format · lint · types · tests · coverage       │
│   spec fidelity · fake-green                     │
└────────────────────┬─────────────────────────────┘
                     ▼
┌──────────────────────────────────────────────────┐
│ REVIEWER agent — a DIFFERENT model               │
│   input: spec + diff + gate findings             │
│   task: disprove, do not confirm                 │
└────────────────────┬─────────────────────────────┘
                     ▼
              findings?  ── yes ──▶ back to CODER (bounded)
                     │
                     no
                     ▼
              evidence bundle → ② human confirms
```

Three properties this must hold:

- **Bounded.** An iteration budget and a hard stop. An unbounded loop between two models burns two subscriptions and produces confidence nobody earned.
- **Interruptible.** Cancel from the phone stops at the next step boundary.
- **Loud.** An agent that dies, a gate that could not run, a reviewer that returned nothing parseable — all block the verdict. `COULD NOT VERIFY` is never a pass.

### Why the reviewer is prompted to disprove

Asked *"is this correct?"*, a model tends to agree. Asked *"find where this fails to meet SPEC-x-01"*, it looks. The reviewer receives the spec, the diff and the gate output, and is instructed to produce findings or state explicitly that it found none.

It is a different model, from a different vendor, so it does not share the Coder's training-time blind spots. That is the entire mechanism — and its honest weakness is that two models can still be wrong together. Deterministic gates carry the weight where judgement cannot.

### Roles

Configuration, not code paths. Any ACP agent can fill any role.

```yaml
# ~/.rla/config.yaml
projects:
  myapp:
    path: ~/code/myapp
    roles:
      coder:    { agent: qwen-code }
      reviewer: { agent: kimi-code }
      explorer: { agent: opencode, model: glm-5.2 }   # optional
    iteration_budget: 8
```

Credentials are absent by design: each agent already holds its own.

### Approval queue

Every agent's `session/request_permission` lands in one queue. One decision each, bound to device id, command hash and a nonce; expiry fails closed ([`protocol.md`](protocol.md#approvalrequest--approvalrespond)).

This is what makes several agents supervisable from a phone: without a single queue, three agents mean three places to look.

### Gates

Deterministic, sub-second where possible, no tokens. Promoted from [`scripts/gate`](../scripts/gate) — the same engine this repository is built with ([ADR-011](decisions.md#adr-011--the-project-is-built-with-its-own-loop)).

| Tier | Runs | Checks |
| :--- | :--- | :--- |
| 0 | every edit | format, types, compile |
| 1 | every iteration | lint, changed-file tests, conformance, fake-green |
| 2 | at convergence | full tests, coverage ratchet, **spec fidelity** |
| 3 | candidate-complete | race, CVE scan, ➕ mutation and fuzzing |

Two properties make a green worth something: **canaries** (each gate is fed planted breakage and must catch it) and **immutability** (gate definitions are compiled, not configuration — weakening one is a reviewable code change).

Full method: [`loop-engineering.md`](loop-engineering.md).

---

## ACP integration

**Transport.** JSON-RPC 2.0 over stdio is the baseline; the orchestrator spawns the agent as a subprocess. Where an agent offers HTTP+SSE — `qwen serve` does — that transport is preferred: it survives orchestrator restarts and supports `Last-Event-ID` replay.

**What we use.** Session create/resume/cancel, streamed session updates, tool-call and diff events, and `session/request_permission`.

**What we never do.** Alter or spoof client identity. Agents identify as themselves; that honesty is what keeps the delegated path legitimate.

**Version tolerance.** Capability negotiation at session start; unknown message types are ignored rather than treated as errors. Tested agent versions are recorded in [`protocol.md`](protocol.md#acp-capability-matrix).

### Agent installation

`rla setup` detects installed agents and installs missing ones **through their own official installers** — `npm install -g …` and equivalents. We orchestrate, never redistribute: that keeps third-party licensing and binary supply-chain surface out of this project entirely.

---

## Relay

Untrusted infrastructure by design. Assume it is hostile and every guarantee still holds.

```
cmd/rla-server/
internal/
├── server/    WSS gateway, session routing
├── pairing/   token issue/verify, public-key exchange
├── stream/    JetStream publish/replay
├── push/      APNs + FCM
├── db/        SQLite, migrations
└── metrics/   observability
```

**What it does:** authenticates devices, routes sealed envelopes, persists them so an offline phone can catch up, sends content-free push notifications.

**What it cannot do:** read payloads, reach a model provider, see a credential. Tests assert all three.

### Sessions, streams, retention

One JetStream stream per session; monotonic sequence numbers; the client acknowledges `lastSeq` and replays from there. That is the whole zero-loss guarantee.

Streaming deltas go out on an ephemeral subject — nobody replays a half-finished sentence. Only completed messages persist.

| Deployment | Age | Size |
| :--- | :--- | :--- |
| **Self-hosted** (default, configurable) | 7 days | 1 GB/stream, 4 GB total |
| Managed — one person (free) | 24 hours | 100 MB/session |
| Managed — team | 30 days | 10 GB/account |

These are **resource limits on someone else's disk**, not features. Self-hosting removes them entirely and costs nothing, which is what keeps the parity promise honest.

When a stream hits its ceiling the oldest events are discarded and the client is told there is a **gap** — never handed a silently incomplete history.

### Observability

From P2: structured JSON logs (no payloads, no tokens), `/metrics` (connections, stream lag, error rate, push failures, replay gaps), `/healthz`, `/readyz`. A relay nobody watches fails silently and the first report comes from a user.

---

## Mobile client

```
mobile/lib/
├── app/            router, theme, l10n (TR + EN)
├── core/           network, crypto, storage, errors, connectivity
├── features/
│   ├── pairing/
│   ├── specs/      ① ratification
│   ├── evidence/   ② review — the screen the product exists for
│   ├── approvals/
│   └── runs/       live status, roles, cancel
└── shared/
```

Riverpod, `go_router`, `drift`, Material 3 light and dark.

Apache-2.0, unlike the rest of the repository ([ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile)). Nothing copyleft enters this directory, and wire types are reimplemented rather than shared — duplication here is deliberate and keeps the licence boundary unambiguous.

---

## CLI host availability

**The orchestrator must be running.** There is no cloud fallback; agents execute on the user's machine against the user's repository ([ADR-009](decisions.md#adr-009--the-cli-host-must-be-reachable)).

When the host is unreachable the app shows the last known state with a timestamp and says so plainly. Queued instructions are held and delivered on reconnect. It never spins indefinitely.

### Headless deployment

```bash
sudo tee /etc/systemd/system/rla.service <<'EOF'
[Unit]
Description=RemLinkAgent orchestrator
After=network-online.target

[Service]
Type=simple
User=youruser
ExecStart=/usr/local/bin/rla daemon
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now rla
```

The trade-off is real: a machine reachable at all times is attackable at all times. Keep the working-directory allowlist tight and approvals on.

### Multiple projects

One daemon supervises several projects, each with its own agent set and role bindings. The phone addresses them individually; status is per-project. This is a first-class case, not an extension — running several projects at once is the situation that motivated the product.

---

## Data storage

| Where | What | Protection |
| :--- | :--- | :--- |
| Agent processes | Their own API keys / subscription tokens | Their own storage; never ours |
| Orchestrator — `~/.rla/` | Config, role bindings, specs, gate cache | Filesystem permissions |
| Orchestrator — OS keychain | Device keys; PAYG keys **only if** the direct path is used | OS-level encryption |
| Relay — SQLite | Devices, sessions, push tokens | No payloads, ever |
| Relay — JetStream | Sealed payloads | E2E encrypted; time and size limited |
| Mobile — secure storage | Device token, session keys | Keychain / Keystore |
| Mobile — drift | Decrypted cache | App sandbox |

Complete inventory of what a relay operator can observe: [`privacy.md`](privacy.md).

---

## Repository layout

```
RemLinkAgent/
├── cmd/rla/                orchestrator binary          AGPL
├── cmd/rla-server/         relay binary                 AGPL
├── internal/               core packages                AGPL
├── mobile/                 Flutter client               Apache-2.0
├── deploy/                 compose, Dockerfile, NATS    AGPL
├── scripts/                gate engine, tooling         AGPL
├── .rla/                   this project's own specs, principles, gates
├── docs/                   public documentation + Pages
└── Makefile
```

`website/` (M1), `watchos/` and `wearos/` (X3) appear when their phases begin.
