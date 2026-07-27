# 🔐 Security Policy

RemLinkAgent runs an AI agent on a developer's machine and lets a phone approve commands on it. **A compromise here is remote code execution on a developer workstation.** The project is treated accordingly.

---

## 📣 Reporting a vulnerability

**Do not open a public issue.**

Use [**GitHub Security Advisories**](https://github.com/burakhalefoglu/RemLinkAgent/security/advisories/new) — private, and it gives you a coordinated-disclosure workspace with the maintainer.

If that is unavailable to you, email the maintainer at the address on the [GitHub profile](https://github.com/burakhalefoglu) with `SECURITY` in the subject.

### Please include

- What the vulnerability allows an attacker to do, and what they need to start
- Reproduction steps or a proof of concept
- Affected component and version (`rla version`)
- Your assessment of severity

### What to expect

| Stage | Target |
| :--- | :--- |
| Acknowledgement | 48 hours |
| Initial assessment | 7 days |
| Fix or mitigation plan | 30 days for high/critical |
| Public advisory | After a fix ships, or 90 days, whichever comes first |

Reporters are credited in the advisory unless they prefer otherwise. There is no bug bounty — this is an unfunded project and pretending otherwise would be dishonest.

> **Pre-release note.** RemLinkAgent is at P0 and has shipped nothing. Reports against the *design* in [`docs/threat-model.md`](docs/threat-model.md) are as welcome as reports against code, and are more useful right now.

---

## 🎯 In scope

- CLI agent (`rla`) and relay server (`rla-server`)
- Mobile client
- Pairing, session and approval protocols — [`docs/protocol.md`](docs/protocol.md)
- End-to-end encryption design and implementation
- The dangerous-command detection and approval path
- Official Docker images and deployment manifests
- Design flaws in the documented threat model

## 🚫 Out of scope

- Vulnerabilities in AI providers themselves — report to them
- Self-hosted deployments misconfigured against our documentation (report a docs bug instead)
- The user's own machine being already compromised — outside our trust boundary
- Social engineering of the maintainer or users
- Automated scanner output with no demonstrated impact
- Missing hardening headers on the marketing site with no exploit path

---

## 🧭 Security model in brief

Full detail: [`docs/threat-model.md`](docs/threat-model.md).

**Trust boundaries**

| Component | Trusted with |
| :--- | :--- |
| CLI host | Everything — API keys, filesystem, command execution. The user's own machine. |
| Mobile device | Session keys, approval authority. Paired explicitly by the user. |
| Relay | **Nothing.** Ciphertext and routing metadata only. Assume it is hostile. |
| AI provider | Prompt content the user chose to send. Never keys belonging to anyone else. |

**Load-bearing guarantees**

1. **API keys never leave the CLI host.** OS keychain, no plaintext on disk, redacted from logs.
2. **The relay cannot read payloads.** End-to-end encrypted; keys exchanged at pairing, never uploaded. [ADR-004](docs/decisions.md#adr-004--end-to-end-encryption-of-relay-payloads).
3. **Pairing tokens are single-use, short-lived and HMAC-signed.** A leaked token is the most dangerous artefact in the system — it is the path to command execution.
4. **Approvals are bound to device identity, command hash and a nonce.** An approval cannot be replayed, nor moved onto a different command.
5. **Command execution is confined to configured working directories.** Escaping that is a vulnerability, not a feature request.
6. **Fail loud.** On an error, the pipeline halts and says so. It never reports success it cannot prove.

---

## 🔑 If you think you have leaked a key

1. Revoke and rotate it with your AI provider immediately.
2. `rla device revoke <id>` for a lost or stolen paired device (`rla device list` to find it).
3. `rla logout` clears every stored credential on the host.

Keys are never transmitted to the relay, so a relay compromise does not expose them — but a compromised **host** does.

---

## 📦 Supply chain

- Dependencies are pinned; `go.sum` and `pubspec.lock` are committed.
- Automated CVE scanning on every dependency change ([P2.14](docs/roadmap.md#p2--server-backend-wss--nats)).
- Release binaries are checksummed and signed ([P4.1](docs/roadmap.md#p4--release--distribution)).
- Verify signatures before running any binary that claims to be ours.

---

## 📜 Disclosure philosophy

The licence invites you to audit this code. That invitation is worthless if reports go into a void. Every valid report gets a real response, and every fixed vulnerability gets a public advisory — including the ones that are embarrassing.
