# 🏗️ Architecture

> **Status:** design · **Updated:** 2026-07-27
> Wire format: [`protocol.md`](protocol.md) · Security: [`threat-model.md`](threat-model.md) · Rationale: [`decisions.md`](decisions.md)

---

## The shape of it

Three components. The important property is what the middle one *cannot* do.

```
┌──────────────────────────────────────────────────────────────┐
│  MOBILE (Flutter)                             Apache-2.0     │
│  Chat · Model switcher · Approvals · Local decryption        │
└───────────────────────────▲──────────────────────────────────┘
                            │ WSS · sealed envelopes
┌───────────────────────────▼──────────────────────────────────┐
│  RELAY (Go)                                   AGPL-3.0       │
│  WSS gateway · JetStream · Push · SQLite                     │
│  Holds no keys. Sees ciphertext + routing metadata.           │
└───────────────────────────▲──────────────────────────────────┘
                            │ WSS · sealed envelopes
┌───────────────────────────▼──────────────────────────────────┐
│  CLI AGENT (Go) — the user's machine          AGPL-3.0       │
│  Agent loop · Tools · Tier 0 · Model router · Keychain       │
└───────────────────────────┬──────────────────────────────────┘
                            │ HTTPS, direct — never via the relay
                            ▼
              AI provider (Z.AI / Qwen / Kimi / any OpenAI-compatible)
```

Two guarantees, often conflated, kept separate here:

**Zero-Touch AI.** AI requests go from the CLI host straight to the provider. API keys live in the OS keychain and have no code path to the relay. This is architectural, not a policy promise.

**End-to-end encryption.** Everything crossing the relay is sealed on the device and on the CLI host. The relay routes ciphertext. See [ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads) for why this is in the MVP rather than deferred.

---

## CLI agent

The product. Everything else is transport and UI.

```
cmd/rla/                    entry point
internal/
├── agent/                  agent loop, orchestration
├── tools/                  file / exec / search tools
├── models/                 provider adapters, router, context normalisation
├── verify/                 Tier 0 gate
├── config/                 ~/.rla/config.yaml, keychain
├── crypto/                 envelope seal/open, key management
├── transport/              WSS client, reconnect, sync
└── daemon/                 IPC, lifecycle
```

### The agent loop

```
user message
    ↓
build context (history + tool schemas)
    ↓
┌─→ call provider (streaming) ─────────────────┐
│       ↓                                      │
│   tool calls requested?                      │
│       ├─ no  → stream final answer → done    │
│       └─ yes ↓                               │
│   dangerous? ─ yes → request approval        │
│       │              (blocks until answered) │
│       ↓ no / approved                        │
│   execute tools, capture results             │
│       ↓                                      │
│   Tier 0 gate on any edits                   │
│       ↓                                      │
└── feed results back ─────────────────────────┘
        (bounded by a step ceiling)
```

Three properties this has to hold:

- **Bounded.** A step ceiling and a hard stop. An agent that loops forever burns tokens and trust in equal measure.
- **Interruptible.** Cancellation from the phone stops it at the next step boundary, not eventually.
- **Loud.** A tool that fails reports failure. It never returns an empty result that reads like success.

### Tools

Each tool is a JSON schema plus a handler. The set is deliberately small.

| Tool | Does | Constrained by |
| :--- | :--- | :--- |
| `read_file` | Read a file | Working-directory allowlist |
| `write_file` | Create or overwrite | Allowlist + approval on paths outside the project |
| `list_files` | Directory listing | Allowlist |
| `search` | Content search | Allowlist |
| `run_command` | Execute, capture output | Allowlist + timeout + danger classification |

**The allowlist is a security boundary, not a convenience.** Every path is resolved and verified to be inside a configured working directory *after* symlink resolution. Escaping it is a vulnerability — see [`threat-model.md`](threat-model.md#t4--command-execution-escape).

### Model router

Providers are configuration. Adding one is a config entry plus a contract test ([ADR-006](decisions.md#adr-006--provider-neutral-core-first-class-zaiqwenkimi)).

```yaml
# ~/.rla/config.yaml
providers:
  zai:
    base_url: https://api.z.ai/v1
    models: [glm-4.6]
  qwen:
    base_url: https://dashscope.aliyuncs.com/compatible-mode/v1
    models: [qwen-2.5-coder]
  kimi:
    base_url: https://api.moonshot.cn/v1
    models: [moonshot-v1-128k]

active: zai/glm-4.6

workspace:
  allow: ["~/code/myproject"]
```

**Hot-swapping** is the differentiator, and the hard part is not the switch — it is the history. Providers disagree on how tool calls and tool results are represented, so switching mid-conversation means translating the whole transcript into the target's shape. Where the target cannot represent something (parallel tool calls, for instance), it is flattened into a form that preserves meaning.

Context overflow when moving to a smaller window is handled by summarising older turns with a cheap model before the handoff, never by silently truncating.

### Tier 0

After every edit: format, lint, type-check or compile. Deterministic, no token cost, sub-second. Failures go back into the loop as tool results, so the agent fixes its own mess before continuing.

This is the whole of Loop Engineering that ships in the MVP. Tiers 1–4 depend on LLM judgement and are in [X6](vision-roadmap.md#x6--loop-engineering-tiers-14).

---

## Relay server

Untrusted infrastructure by design. Assume it is hostile and the guarantees still hold.

```
cmd/rla-server/
internal/
├── server/    WSS gateway, session routing
├── pairing/   token issue/verify, key exchange
├── stream/    JetStream publish/replay
├── push/      APNs + FCM
├── db/        SQLite, migrations
└── metrics/   observability
```

**What it does:** authenticates devices, routes sealed envelopes, persists them in JetStream so an offline phone can catch up, and sends content-free push notifications.

**What it cannot do:** read payloads, reach an AI provider, or see an API key. Tests assert all three.

### Sessions and streams

One JetStream stream per session. Every event carries a monotonic sequence number; the client acknowledges its `lastSeq` and replays from there on reconnect. That is the whole of the zero-loss guarantee.

Streaming token deltas are a special case: persisting each one would bloat the stream for no benefit, since nobody replays a half-finished sentence. Deltas go out on an ephemeral subject; only the completed message is persisted.

### Retention

Retention is a policy decision with a disk-space consequence, so it is explicit rather than implied by a plan tier:

| Deployment | Age limit | Size limit |
| :--- | :--- | :--- |
| Self-hosted (default) | 7 days | 1 GB per stream, 4 GB total |
| Managed Free | 24 hours | 100 MB per session |
| Managed Pro | 7 days | 1 GB per session |
| Managed Team | 30 days | 10 GB per account |

Self-hosted limits are configurable. When a stream hits its ceiling the oldest events are discarded and the client is told there is a gap, rather than being handed a silently incomplete history.

### Observability

Not deferred — a relay nobody is watching fails silently, and the first report comes from a user. From P2: structured JSON logs (no payloads, no tokens), `/metrics` (connections, stream lag, error rate, push failures, replay gaps), `/healthz`, `/readyz`.

---

## Mobile application

```
mobile/lib/
├── main.dart
├── app/            router, theme, l10n (TR + EN)
├── core/           network (WSS), crypto, storage, errors, connectivity
├── features/
│   ├── pairing/
│   ├── chat/
│   ├── models/     switcher
│   └── approvals/
└── shared/         reusable widgets
```

Riverpod for state, `go_router` for routing, `drift` for the offline cache, Material 3 in light and dark.

Apache-2.0, unlike the rest of the repository — see [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile). Nothing copyleft may enter this directory, and wire types are reimplemented here rather than shared, which keeps the boundary unambiguous.

---

## CLI host availability

**The daemon must be running for anything to work.** There is no cloud fallback; that is the direct consequence of BYOK and a local agent ([ADR-009](decisions.md#adr-009--the-cli-host-must-be-reachable)).

When the host is unreachable the app shows the last known state and says so plainly, with the time of last contact. Queued messages are held and delivered on reconnect. What it never does is spin indefinitely.

### Headless deployment

For always-on availability, run the daemon somewhere that is always on:

```bash
# systemd — a small VPS, home server, or always-on desktop
sudo tee /etc/systemd/system/rla.service <<'EOF'
[Unit]
Description=RemLinkAgent daemon
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

The trade-off is real and worth stating: a machine reachable at all times is a machine that can be attacked at all times. Keep the working-directory allowlist tight and the approval flow on.

Wake-on-LAN for sleeping hosts is a plausible convenience, not a commitment.

---

## Data storage

| Where | What | Protection |
| :--- | :--- | :--- |
| CLI host — OS keychain | AI API keys, device keys | OS-level encryption |
| CLI host — `~/.rla/` | Config, session state | Filesystem permissions |
| Relay — SQLite | Devices, sessions, push tokens | No payloads, ever |
| Relay — JetStream | Sealed event payloads | E2E encrypted; time and size limited |
| Mobile — secure storage | Device token, session keys | Keychain / Keystore |
| Mobile — drift | Decrypted message cache | App sandbox |

Full inventory of what the relay operator can observe: [`privacy.md`](privacy.md).

---

## Repository layout

```
RemLinkAgent/
├── cmd/rla/                CLI binary                    AGPL
├── cmd/rla-server/         relay binary                  AGPL
├── internal/               core packages                 AGPL
├── mobile/                 Flutter app                   Apache-2.0
├── deploy/                 compose, Dockerfile, NATS     AGPL
├── scripts/                tooling (make.ps1)            AGPL
├── docs/                   public documentation + Pages
├── LICENSE                 AGPL-3.0-or-later
├── mobile/LICENSE          Apache-2.0
├── LICENSE_HEADER*         header templates
└── Makefile
```

`website/` (M1), `watchos/` and `wearos/` (X3/X4) appear when their phases begin.
