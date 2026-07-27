# 🔀 Wire Protocol

> **Protocol version:** 1 (draft — not frozen until P2 ships)
> **Status:** design · **Updated:** 2026-07-27

The contract between the CLI agent, the relay and the mobile client. All three ship on independent cadences, so **the wire format is versioned separately from the software** and both ends must be able to detect an incompatible peer and say so.

Rationale for the design constraints: [`decisions.md`](decisions.md) · attacker model: [`threat-model.md`](threat-model.md).

---

## Versioning

`Protocol` is an integer, independent of any binary's semantic version. `internal/version.Protocol` is the source of truth on the Go side.

**Bump it when:** a field is removed or renamed, a field's meaning changes, a new message type becomes mandatory, or the crypto envelope changes.

**Do not bump for:** a new optional field, a new message type both sides may ignore, or a new error code.

### Handshake

Both ends declare their version before anything else. There is no implicit negotiation and no guessing.

```jsonc
// client → relay, first frame after connect
{
  "type": "hello",
  "protocol": 1,
  "client": { "kind": "cli" | "mobile", "version": "0.3.1", "platform": "darwin/arm64" },
  "deviceId": "dev_01HQ...",
  "auth": "<device token>"
}
```

```jsonc
// relay → client
{
  "type": "hello.ack",
  "protocol": 1,
  "minProtocol": 1,
  "server": { "version": "0.3.0" },
  "sessionId": "ses_01HQ...",
  "lastSeq": 1247
}
```

If the client's `protocol` is below `minProtocol`, the relay replies with `error.protocol_unsupported` and closes. The client surfaces an explicit "update required" message. **It never falls back to guessing** — a silently degraded protocol is a correctness bug waiting to be blamed on something else.

Support policy: the relay accepts the current version and the one before it. Two behind is a hard failure with an upgrade instruction.

---

## Transport

WebSocket Secure on 443. JSON text frames; binary frames reserved for future compression.

| Property | Value |
| :--- | :--- |
| Heartbeat | ping every 30 s, pong deadline 10 s |
| Reconnect | exponential backoff 1 s → 60 s, full jitter |
| Max frame | 1 MB (matches the NATS `max_payload` ceiling) |
| Idle timeout | 5 minutes without a pong |

Transport TLS is required even though payloads are already encrypted: it protects the routing metadata and the device token, which are not inside the sealed envelope.

---

## Encryption

Every payload crossing the relay is sealed. The relay routes an envelope whose contents it cannot read ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)).

**Primitives** — NaCl `box`: X25519 key agreement, XSalsa20-Poly1305 AEAD. Chosen for one reason above all: mature, audited, hard-to-misuse implementations exist on both sides (`golang.org/x/crypto/nacl/box`, Dart `cryptography`).

### Key exchange at pairing

```
CLI                          relay                        mobile
 │                             │                             │
 │─ POST /pair/init ──────────>│                             │
 │  { cliPubKey }              │                             │
 │<─ { token, expiresAt } ─────│                             │
 │                             │                             │
 │  displays QR:               │                             │
 │  remlinkagent://pair?token=…&host=…&fp=<cliPubKey fp>     │
 │                             │                             │
 │                             │<─ POST /pair/complete ──────│
 │                             │   { token, mobilePubKey }   │
 │                             │                             │
 │<─ pair.completed ───────────│─── { cliPubKey } ──────────>│
 │   { mobilePubKey }          │                             │
 │                             │                             │
 │══ both derive a shared secret; the relay never sees it ══│
```

The relay forwards public keys. It never possesses a private key or a derived secret. The QR carries a **fingerprint** of the CLI public key so the phone can verify that the key the relay forwarded is the key the CLI actually published — this is what closes the relay-substitutes-its-own-key attack.

### Envelope

```jsonc
{
  "type": "event",
  "seq": 1248,
  "sessionId": "ses_01HQ...",
  "eventType": "message",        // routing metadata — visible to the relay
  "ts": "2026-07-27T10:15:30Z",
  "sealed": {
    "nonce": "<base64, 24 bytes>",
    "box":   "<base64 ciphertext>"
  }
}
```

Deliberately **outside** the sealed box, because the relay needs it to route: `seq`, `sessionId`, `eventType`, `ts`. Everything of substance — message text, tool calls, command strings, results — is inside.

`eventType` leaks the *kind* of thing happening (a message vs. an approval request). That is the price of routing push notifications without a plaintext payload, and it is listed in [`privacy.md`](privacy.md) rather than glossed over.

---

## Message types

### `event` — CLI → mobile

Sealed contents:

```jsonc
{
  "kind": "message.delta" | "message.complete" | "tool.call" | "tool.result"
        | "approval.request" | "status" | "error",
  "payload": { /* per kind */ }
}
```

| Kind | Persisted | Notes |
| :--- | :--- | :--- |
| `message.delta` | ❌ ephemeral | Streaming tokens. Not replayed — nobody needs half a sentence from yesterday. |
| `message.complete` | ✅ | The final message. This is what replay reconstructs. |
| `tool.call` | ✅ | Tool name and arguments, plus which agent and role issued it. |
| `tool.result` | ✅ | Output, exit code, duration. Truncated above 64 KB with a marker. |
| `approval.request` | ✅ | Blocks the agent until answered. Carries the role that asked. |
| `spec.draft` | ✅ | A spec awaiting **checkpoint ①**. Requirements, ids, open questions. |
| `run.status` | ✅ | Which agent holds which role, current step, iteration count. |
| `evidence.bundle` | ✅ | **Checkpoint ②.** Gate results, reviewer findings, requirement coverage, and anything that could not be verified. |
| `error` | ✅ | Always surfaced — never swallowed. |

`evidence.bundle` is the payload the product exists to produce. It must distinguish three states per check — passed, failed, and **could not verify** — because collapsing the third into either of the others is the failure this system prevents.

### `command` — mobile → CLI

```jsonc
{
  "kind": "instruct"          // free-form instruction to the Coder
        | "spec.ratify"       // ① approve, request changes, or reject
        | "approval.respond"  // resolve one permission request
        | "run.start"         // begin a run against a ratified spec
        | "run.cancel"
        | "role.assign",      // bind an agent to a role for this project
  "payload": { /* per kind */ }
}
```

### `approval.request` / `approval.respond`

The most security-sensitive exchange in the protocol.

```jsonc
// CLI → mobile (sealed)
{
  "kind": "approval.request",
  "payload": {
    "approvalId": "apr_01HQ...",
    "command": "rm -rf ./build",
    "commandHash": "sha256:9f2b...",
    "cwd": "/Users/x/code/myproject",
    "reason": "recursive delete",
    "expiresAt": "2026-07-27T10:20:30Z"
  }
}
```

```jsonc
// mobile → CLI (sealed)
{
  "kind": "approval.respond",
  "payload": {
    "approvalId": "apr_01HQ...",
    "commandHash": "sha256:9f2b...",   // must match, or the CLI refuses
    "decision": "approve" | "reject",
    "nonce": "<the nonce from the request>",
    "deviceId": "dev_01HQ..."
  }
}
```

The CLI enforces all of:

1. `approvalId` refers to a pending request.
2. `commandHash` matches the request exactly — an approval cannot be moved onto a different command.
3. `nonce` has not been used — no replay.
4. `deviceId` is a currently paired, non-revoked device.
5. `expiresAt` has not passed — **expiry fails closed**, always reject.

Any check failing is logged as a security event, not merely refused.

### `sync` — resuming

```jsonc
// client → relay
{ "type": "sync", "sessionId": "ses_01HQ...", "lastSeq": 1200 }

// relay → client
{ "type": "sync.begin", "fromSeq": 1201, "toSeq": 1248, "gap": false }
// … events 1201 … 1248 …
{ "type": "sync.end", "lastSeq": 1248 }
```

`gap: true` means retention dropped events the client never saw. **The client must show this**, not paper over it. An incomplete history presented as complete is exactly the silent failure the project's fail-loud rule exists to prevent.

### `error`

```jsonc
{
  "type": "error",
  "code": "protocol_unsupported" | "auth_failed" | "device_revoked"
        | "rate_limited" | "session_not_found" | "payload_too_large"
        | "approval_expired" | "internal",
  "message": "human-readable, no secrets",
  "retryable": true
}
```

Error messages never contain tokens, keys, or payload content.

---

## Push notifications

Because the relay cannot read payloads, notifications carry no content:

```jsonc
{
  "sessionId": "ses_01HQ...",
  "eventType": "approval.request",
  "seq": 1248
}
```

The device wakes, fetches, decrypts, and renders the visible text locally. On iOS this means a mutable-content notification with a service extension; on Android, a data message handled before display.

The cost is honest: a device that cannot reach the relay shows a generic notification instead of a specific one. That is the correct trade against the relay knowing what your commands say.

---

## ACP capability matrix

The second contract in this system, and the one P1 depends on. RemLinkAgent drives existing agents over the [Agent Client Protocol](https://agentclientprotocol.com/) — JSON-RPC 2.0 over stdio, with `session/request_permission` as the approval mechanism ([ADR-014](decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)).

The protocol is a shared standard, but coverage differs per agent. [P0.17](roadmap.md#p0--scaffolding-gates--de-risking) measures, and the results land here:

| Check | Why it matters |
| :--- | :--- |
| `session/request_permission` payload | What the client learns about the proposed action — enough to render a decision on a phone? |
| **Permission hold time** | Can a request survive a phone round trip, or does it time out locally? **This decides the whole approval design.** |
| Session lifecycle | Create, resume, cancel; what survives an agent restart |
| Diff retrieval | How completed work is extracted for handoff to the reviewer |
| Concurrent sessions | Can one orchestrator hold several agents at once without interference? |
| Transport | stdio only, or HTTP+SSE (`qwen serve`) with `Last-Event-ID` replay |
| Version negotiation | Which capabilities are advertised, and what happens when one is absent |

### Known so far

| Agent | Transport | Permission | Notes |
| :--- | :--- | :--- | :--- |
| **Qwen Code** | stdio + **HTTP/SSE** (`qwen serve`) | `permission_request` SSE → `POST /permission/:requestId`, first responder wins | 8000-frame ring buffer, `Last-Event-ID` resume, multi-client sessions, TLS for LAN/mobile, bearer auth |
| **Kimi Code** | stdio (`kimi acp`) | ACP baseline | To be measured |
| **OpenCode** | stdio | ACP baseline | Provider-agnostic — the route to providers that ship no CLI |

> **Fill this in during P0.17.** If the permission-hold row is still empty when P1 begins, the approval design is proceeding on an unverified assumption — and it is the assumption the product rests on.

### Rules for every agent integration

1. **Never alter client identity.** Agents identify as themselves. This is what keeps the delegated path legitimate ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).
2. **Never handle their credentials.** Each agent authenticates itself.
3. **Capability-negotiate, do not assume.** A missing capability degrades the run explicitly; it never silently changes behaviour.
4. **Unknown messages are ignored, not errors.** Same additive rule as the wire protocol above.

---

## Direct provider path ➕

The fallback when no agent is installed: any OpenAI-compatible endpoint with the user's own **pay-as-you-go** key. Never a subscription endpoint ([ADR-012](decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).

Contract tests for this path cover tool schema shape, parallel tool calls, streaming deltas, finish reasons and rate-limit headers — the same checks as before, but for a secondary path rather than the primary one.

---

## Reserved for later

Declared now so that adding them does not force a protocol break:

- `terminal.*` — PTY streaming ([X4](vision-roadmap.md#x4--interactive-terminal-mirroring))
- `verify.*` — heavy verification tier events ([X1](vision-roadmap.md#x1--deeper-verification-tiers))

Implementations must ignore unknown `type` and `kind` values rather than erroring. That single rule is what makes additive change possible without a version bump.
