# 📚 RemLinkAgent Documentation

Everything here is public and current. `index.html` in this directory is the landing page served at [remlinkagent.com](https://remlinkagent.com).

---

## Start here

| Document | What it answers |
| :--- | :--- |
| [**decisions.md**](decisions.md) | *Why is it like this?* The ADR log — naming, licensing, encryption, scope. Read this before proposing a change to any of them. |
| [**roadmap.md**](roadmap.md) | *What is being built, in what order?* P0–P4 (MVP) and M1 (managed cloud). |
| [**architecture.md**](architecture.md) | *How does it fit together?* Components, agent loop, data flow, deployment. |

## Building on it

| Document | What it answers |
| :--- | :--- |
| [**protocol.md**](protocol.md) | Wire format, versioning, encryption envelope, message types. The contract between all three components. |
| [**threat-model.md**](threat-model.md) | Adversaries, trust boundaries, mitigations, and what is explicitly *not* defended against. |
| [**license-header.md**](license-header.md) | Which licence header goes in which directory, and why the boundary is one-way. |

## Product and policy

| Document | What it answers |
| :--- | :--- |
| [**privacy.md**](privacy.md) | Exactly what the relay can and cannot see. The honest counterpart to the encryption claim. |
| [**vision-roadmap.md**](vision-roadmap.md) | X1–X6 — deferred capability, and the metrics that would justify starting each. |
| [**loop-engineering.md**](loop-engineering.md) | The tiered autonomous QA design. Tier 0 ships in the MVP; the rest is X6. |

---

## Reading paths

**Evaluating the project** → [`decisions.md`](decisions.md) → [`roadmap.md`](roadmap.md) → [`privacy.md`](privacy.md)

**Contributing code** → [`../CONTRIBUTING.md`](../CONTRIBUTING.md) → [`architecture.md`](architecture.md) → [`roadmap.md`](roadmap.md) for the current phase

**Reviewing security** → [`threat-model.md`](threat-model.md) → [`protocol.md`](protocol.md) → [`../SECURITY.md`](../SECURITY.md)

**Curious about the licensing** → [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile) → [`license-header.md`](license-header.md) → [`../CLA.md`](../CLA.md)

---

## Conventions

**Decisions are append-only.** Changing an ADR means adding an entry that supersedes it, never editing history. What was decided *and later reversed* is often more useful than what stands today.

**Phase prefixes** — `P0`–`P4` MVP · `M1` managed cloud · `X1`–`X6` vision. Unambiguous on purpose: "V1" previously meant three different things in three documents.

**Status is stated.** Every document says whether it describes something implemented, designed, or deferred. Nothing here should imply more than exists.

---

*Something unclear or wrong is a bug — [open an issue](https://github.com/burakhalefoglu/RemLinkAgent/issues).*
