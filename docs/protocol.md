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
| `tool.call` | ✅ | Tool name and arguments. |
| `tool.result` | ✅ | Output, exit code, duration. Truncated above 64 KB with a marker. |
| `approval.request` | ✅ | Blocks the agent until answered. |
| `status` | ✅ | Agent state transitions. |
| `error` | ✅ | Always surfaced — never swallowed. |

### `command` — mobile → CLI

```jsonc
{
  "kind": "chat.send" | "model.switch" | "approval.respond" | "session.cancel",
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

## Provider contract tests

Separate from the wire protocol, but equally a contract: RemLinkAgent assumes every provider is OpenAI-compatible. That assumption is solid for chat completion and much weaker for tool calling and streaming.

[P0.17](roadmap.md#p0--scaffolding--de-risking) verifies per provider, and the results land here:

| Check | Why it matters |
| :--- | :--- |
| Tool schema shape | Some providers reject `additionalProperties`, or nested schemas |
| Parallel tool calls | Not universally supported; the loop must cope with either |
| `tool_choice` honoured | If ignored, forcing a tool is impossible |
| Streaming tool deltas | Chunk boundaries differ; some emit arguments in fragments |
| Finish reasons | Naming varies (`tool_calls`, `function_call`, …) |
| Token accounting | Field names and whether usage appears on streamed responses |
| Rate-limit headers | Needed for informed backoff rather than blind retry |

Each provider gets an executable contract test. A provider that fails one gets an adapter — not a special case scattered through the agent loop.

> **Findings from P0.17 are recorded here.** If this section is still empty when P1 begins, the spike did not happen and P1 is proceeding on an unverified assumption.

---

## Reserved for later

Declared now so that adding them does not force a protocol break:

- `terminal.*` — PTY streaming ([X5](vision-roadmap.md#x5--interactive-terminal-mirroring-pty))
- `loop.*` — Loop Engineering tier events ([X6](vision-roadmap.md#x6--loop-engineering-tiers-14))
- `project.*` — multi-project ([X1](vision-roadmap.md#x1--multi-model-orchestration))

Implementations must ignore unknown `type` and `kind` values rather than erroring. That single rule is what makes additive change possible without a version bump.
