# 🔒 Privacy & Data Inventory

> **Status:** design — becomes the published privacy policy at [P4.5](roadmap.md#p4--release--distribution)
> **Updated:** 2026-07-27

This is the honest counterpart to the encryption claim. Saying "the relay cannot read your data" is only meaningful alongside a complete statement of what it *can* see. Everything the relay observes is listed here.

It also has a practical purpose: App Store and Play Store privacy labels require exactly this inventory, and writing it late means writing it hurriedly.

---

## Summary

| | |
| :--- | :--- |
| **AI traffic** | Never reaches our servers. Direct from your machine to your provider. |
| **API keys** | Never leave your machine. OS keychain, no transmission path. |
| **Message content** | End-to-end encrypted. The relay stores ciphertext. |
| **Your code** | Never uploaded anywhere except to the AI provider you chose. |
| **What we see** | Routing metadata — enumerated below. |
| **Self-hosting** | Removes us from the picture entirely. Free, full parity. |

---

## What the relay stores

### Ciphertext — unreadable to us

Messages, model output, commands and their results, file contents in tool calls, approval requests. Sealed on your device or your CLI host with keys exchanged during pairing. **We hold no key capable of opening them.** Not on the server, not in backups, not under legal compulsion — the capability does not exist.

Retention per [`architecture.md`](architecture.md#sessions-streams-retention): 24 hours to 30 days depending on plan, then deleted.

### Metadata — visible to us

This is the complete list. If something is not here, we do not have it.

| Data | Why it exists | Retention |
| :--- | :--- | :--- |
| Device identifier | Route events to the right device | Until unpaired |
| Device public key | Establish encryption | Until unpaired |
| Platform and app version | Compatibility, debugging | Until unpaired |
| Session identifier | Group events | Retention window |
| Event sequence numbers | Lossless replay | Retention window |
| **Event type** | Route push notifications | Retention window |
| Payload size | Enforce limits | Not stored |
| Connection timestamps | Rate limiting, abuse prevention | 30 days |
| IP address | Rate limiting, abuse prevention | 7 days |
| Push token (APNs/FCM) | Deliver notifications | Until unpaired |

**On event type.** The relay knows a message is an approval request rather than a chat message. It does not know what command is being approved. This exists because push notifications cannot be routed without it — an unavoidable consequence of content-free notifications. It is disclosed rather than glossed over.

**What that leaks.** Someone with full relay access could infer when you work, how often, how long sessions run, and roughly how much content moves. They could not learn what you are building, what your code says, or what commands you ran.

### Managed cloud only (M1)

Email, plan, Stripe customer id, invoice history, Team membership. Payment card details go to Stripe and never touch our systems.

---

## What we never collect

- Message or code content in readable form
- API keys, of any provider
- File contents, paths, or repository names
- Command strings or their output
- Analytics, telemetry, behavioural tracking
- Advertising identifiers
- Location
- Contacts, photos, or anything outside the app's own data

There is no analytics SDK in the mobile app. Not a privacy-preserving one, not a self-hosted one. None.

---

## On your devices

**CLI host** — API keys in the OS keychain (macOS Keychain, Windows Credential Manager, Linux libsecret); config and session state in `~/.rla/`; nothing transmitted except sealed payloads.

**Mobile** — device token and session keys in Keychain/Keystore; decrypted message cache in the app sandbox, cleared on unpair.

`rla logout` clears the host. Unpairing clears the device. Neither requires our involvement.

---

## Third parties

| Service | Receives | Why | Avoidable? |
| :--- | :--- | :--- | :--- |
| **Your AI provider** | Your prompts and code | You chose them; requests go direct | No — but you pick who |
| **Apple APNs** | Push token, content-free payload | iOS notifications | Yes — disable push |
| **Google FCM** | Push token, content-free payload | Android notifications | Yes — disable push |
| **Stripe** (M1) | Email, payment details | Billing | Yes — self-host |

Apple and Google see that a notification was sent, not what it says.

**Your AI provider sees your prompts and your code.** That is inherent to using a hosted model, and their policy governs it, not ours. Choose accordingly.

---

## Your rights

Self-hosted: you already hold everything.

Managed cloud (M1):

- **Access** — export your metadata; message content is exportable only from your own devices, since we cannot decrypt it.
- **Deletion** — delete your account; metadata is removed within 30 days, ciphertext immediately.
- **Portability** — machine-readable export.
- **Objection** — self-hosting is always available at full parity.

GDPR and KVKK apply. Requests go to the contact address published with the policy at P4.5.

---

## Store privacy labels

Draft for [P4.6](roadmap.md#p4--release--distribution).

**Apple — Data Not Linked to You:** Identifiers (device id), Usage Data (connection timestamps). **Not collected:** contact info, health, financial, location, contacts, user content, search history, diagnostics.

> *User content* is deliberately marked not collected. The relay transports ciphertext it cannot read; Apple's definition turns on whether the operator can access the content, and here it cannot.

**Google Play — Data Safety:** collects Device ID and App Activity, encrypted in transit, deletable on request. Data is **not** shared with third parties for advertising or analytics.

---

## Changes

Material changes are announced in [`CHANGELOG.md`](../CHANGELOG.md), in release notes, and in-app before taking effect.

**Two commitments that will not change:** we will never add content-visible telemetry, and we will never add a code path that lets the relay read payloads. Either would require breaking [ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads) — and the licence means you can verify we have not.
