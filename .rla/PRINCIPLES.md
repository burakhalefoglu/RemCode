# PRINCIPLES — RemLinkAgent

> **This file is the constitution of this codebase.** The project brief decides
> *what* gets built; this decides *how*. Where the two conflict, stop and
> report — do not resolve it silently.
>
> **Not relaxable.** Not by the agent, not under deadline, not "just this once".
> Changing anything here requires a superseding entry in
> [`docs/decisions.md`](../docs/decisions.md), not an edit in passing.

---

## P0 — Vendor neutrality

We sell no model and favour none. This is not a marketing stance — it is the
only reason a verdict of *"the Coder got this wrong"* carries any weight. Every
model vendor is structurally incapable of issuing that verdict about itself.

Nothing in this codebase may privilege one provider: not in defaults, not in
prompt wording, not in how findings are ranked.

**Corollary — no impersonation.** Agents identify as themselves. No header,
User-Agent or client identifier is ever altered to resemble another tool. Not to
unlock a subscription tier, not for testing, not temporarily. Every provider
prohibits it, and the moment we do it the neutrality claim above becomes a lie.

**Enforced by:** review, and by [ADR-012](../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us).

## P1 — Zero-Touch AI

AI traffic and API keys never reach the relay. No exceptions, no debug flag, no
temporary telemetry, no "just for local development".

**Enforced by:** `gate t1 → zero-touch-ai` — relay packages may not import
provider, agent, tool or crypto packages. The import graph is the guarantee;
the sentence above is only its description.

## P2 — The relay is untrusted

Payloads crossing the relay are ciphertext. Any change that would let the relay
read plaintext is rejected regardless of what it enables.

Design as though the relay is hostile — because for a self-hosting user, some
day, it will be someone else's.

## P3 — Fail loud

An error halts and says so. A gate that crashes reports `COULD NOT VERIFY`,
never `PASSED`. A tool that fails returns a failure, never an empty result that
reads like success.

**A green you cannot substantiate is worse than a red**, because a red gets
looked at.

**Enforced by:** `nilerr` and `errcheck` in lint; `Unverified` as a first-class
gate status distinct from `Pass`.

## P4 — Evidence, not assertion

"It works" is not a claim, it is a question. Every requirement carries a
`SPEC-…` id and something in the code cites it. Every test asserts something.
Every gate can prove it still detects breakage.

**Enforced by:** `gate t2 → spec-fidelity`, `gate t1 → fake-green`,
`gate canary`.

## P5 — Gates are immutable to the loop

Gate definitions, thresholds and lint rules cannot be edited to make a failure
go away. The loop is rewarded for green; that reward must not be reachable by
lowering the bar.

Gate definitions live in compiled Go (`scripts/gate/main.go`), not in a config
file, so weakening one is a code change that appears in review. The coverage
floor is a ratchet: it may rise freely and only ever falls deliberately.

## P6 — The user is in the loop

Two checkpoints are not automatable and are never skipped:

- **① Is the plan right?** A spec stays `draft` until a human ratifies it. The
  fidelity gate proves code matches the spec; it cannot prove the spec is
  correct.
- **② Does it actually work?** Every gate green still does not mean the
  end-to-end experience is right.

**Enforced by:** `gate verify` refuses to declare readiness while any spec is
`draft`.

## P7 — Lightness

CLI target < 30 MB RSS. Every dependency is justified. The smallest dependency
surface is also the smallest attack surface — this is a security principle
wearing a performance costume.

## P8 — Nothing hard-coded that a human reads

User-facing text is an i18n key, in every language, including Turkish. This
applies to CLI output, push notification bodies and error messages alike.

## P9 — Licence boundary is one-way

Apache-2.0 code (`mobile/`) may flow into the AGPL core. Core code may never
flow into `mobile/`. See
[ADR-002](../docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile).

**Enforced by:** `gate t2 → licence-boundary`, `gate t2 → licence-headers`.

## P10 — Decisions are append-only

An [ADR](../docs/decisions.md) is superseded, never rewritten. What was decided
and later reversed is often more instructive than what currently stands.

---

## Conflict resolution

When the task and these principles disagree:

1. **Stop.** Do not pick one silently.
2. **Report** the conflict, with both readings stated plainly.
3. **Wait** for a human decision.
4. If the decision changes a principle, it lands in `docs/decisions.md` first.

A principle quietly bent once becomes a principle that no longer exists.
