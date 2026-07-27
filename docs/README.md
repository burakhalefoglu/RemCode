# 📚 RemLinkAgent Documentation

Everything here is public and current. `index.html` in this directory is the landing page served at [remlinkagent.com](https://remlinkagent.com).

**The thesis in one line:** one model implements, a *different vendor's* model tries to disprove it against a ratified spec, deterministic gates produce the evidence, a human decides.

---

## Start here

| Document | What it answers |
| :--- | :--- |
| [**decisions.md**](decisions.md) | *Why is it like this?* The ADR log. Read before proposing a change to naming, licensing, encryption or scope — several entries record decisions that were reversed by research, and why. |
| [**loop-engineering.md**](loop-engineering.md) | *What is the product?* The verification method: tiers, spec fidelity, cross-model review, canaries, fake-green. |
| [**roadmap.md**](roadmap.md) | *What is being built, in what order?* P0–P4 and M1. |
| [**architecture.md**](architecture.md) | *How does it fit together?* Orchestrator, ACP agents, relay, mobile. |

## Building on it

| Document | What it answers |
| :--- | :--- |
| [**development-loop.md**](development-loop.md) | *How do we work?* The gates this repository is built with — the product, running on itself. Read before your first PR. |
| [**protocol.md**](protocol.md) | Wire format to the phone, ACP capability matrix for agents, encryption envelope, message types. |
| [**threat-model.md**](threat-model.md) | Adversaries, trust boundaries, mitigations, and what is explicitly *not* defended against. |
| [**license-header.md**](license-header.md) | Which licence header goes in which directory, and why the boundary is one-way. |

## Product and policy

| Document | What it answers |
| :--- | :--- |
| [**privacy.md**](privacy.md) | Exactly what the relay can and cannot see. The honest counterpart to the encryption claim. |
| [**vision-roadmap.md**](vision-roadmap.md) | X1–X4 — deferred capability, and the metrics that would justify starting each. |

---

## Reading paths

**Evaluating the project** → [`loop-engineering.md`](loop-engineering.md) → [ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent) → [`roadmap.md`](roadmap.md)

**Contributing code** → [`../CONTRIBUTING.md`](../CONTRIBUTING.md) → [`development-loop.md`](development-loop.md) → [`architecture.md`](architecture.md)

**Reviewing security** → [`threat-model.md`](threat-model.md) → [`protocol.md`](protocol.md) → [`../SECURITY.md`](../SECURITY.md)

**Curious about the licensing** → [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile) → [`license-header.md`](license-header.md) → [`../CLA.md`](../CLA.md)

---

## Conventions

**Decisions are append-only.** Changing an ADR means adding an entry that supersedes it, never editing history. Three entries have been superseded so far, each by research that contradicted an assumption — that record is more useful than a document that only ever agreed with itself.

**Phase prefixes** — `P0`–`P4` MVP · `M1` managed cloud · `X1`–`X4` vision.

**Status is stated.** Every document says whether it describes something implemented, designed, or deferred. Nothing here should imply more than exists.

---

*Something unclear or wrong is a bug — [open an issue](https://github.com/burakhalefoglu/RemLinkAgent/issues).*
