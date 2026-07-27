# SECURITY BASELINE — RemLinkAgent

> Project-specific security invariants enforced **while code is being written**,
> not discovered in review. The full adversary analysis is
> [`docs/threat-model.md`](../docs/threat-model.md); this is the operational
> subset the gates check.

**What is at stake:** an agent that runs commands on a developer workstation. A
successful attack here is remote code execution on a machine holding source,
cloud credentials and production access. Calibrate accordingly.

---

## Invariants

### S1 — Credentials never leave the host

API keys live in the OS keychain. They are never written to disk in plaintext,
never sent to the relay, never included in a crash report.

**Gate:** `secret-logging` (Tier 1) rejects credential-shaped identifiers passed
to log or print calls. Opt out only with `//gate:allow-secret-log` and a comment
explaining why.
**Test:** a disk-grep assertion after `rla login` (P1.2).

### S2 — The relay holds no keys

No key material, no derived secret, no code path that could acquire one.

**Gate:** `zero-touch-ai` (Tier 1) forbids relay packages from importing
`internal/crypto`.

### S3 — Paths are resolved before they are checked

Every filesystem operation resolves symlinks **first**, then verifies
containment within the configured working directory. Checking before resolving
is bypassable and therefore wrong.

**Owner:** P1.6. Adversarial tests are part of the requirement, not a follow-up.

### S4 — Command classification is allowlist-oriented

Known-safe commands run. Everything else asks a human. Enumerating badness does
not work: `$(echo cm) -rf /` defeats any regex blocklist, and so will the next
thing nobody thought of.

**Owner:** P1.10.

### S5 — Approvals authorise exactly one command, once

Every approval is verified against: pending id, exact command hash, unused
nonce, paired and non-revoked device, and expiry. **Expiry fails closed.**

Any check failing is a security event in the log, not merely a refusal.

**Owner:** P1.10, P2.14.

### S6 — Pairing tokens are the crown jewels

Single-use, ~60 s TTL, HMAC-signed, bound to a CLI public key. A leaked pairing
token is the shortest path to code execution on the host.

**Owner:** P2.4.

### S7 — Dependencies are minimal and current

Every dependency is justified in the PR that adds it. CVE scanning runs on
every dependency change.

**Gate:** `vulnerabilities` (Tier 3).

### S8 — Sequence gaps are surfaced

An incomplete event history is never presented as complete. Retention dropping
events is normal; hiding that it happened is not.

**Owner:** P2.10, P3.2.4.

---

## Applies to every change

- [ ] No credential can reach a log, an error message or the relay.
- [ ] No new code path lets the relay read plaintext.
- [ ] Filesystem access stays inside the allowlist, after symlink resolution.
- [ ] New failure modes fail **closed**.
- [ ] New dependencies are justified and scanned.
- [ ] Anything security-relevant is reflected in `docs/threat-model.md`.

---

## When a gate cannot decide

Some of this is not machine-checkable — whether a classification is genuinely
allowlist-oriented, whether a failure mode really fails closed. Those are
judgement gates. `gate verify` lists them as owed rather than omitting them,
because an unlisted obligation looks exactly like a met one.
