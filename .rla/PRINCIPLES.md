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

**And every gate says how much it examined.** An exit code answers *"did what
ran pass?"*; it cannot answer *"did anything run?"* — and a suite that selects
zero tests exits 0 exactly like a healthy one. Counts are therefore part of the
evidence, held to declared floors and to a committed baseline, and a cached
pass is re-judged against the counts it recorded rather than trusted for having
matched a signature
([ADR-028](../docs/decisions.md#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard)).

**Enforced by:** `gate full → spec-fidelity`, `gate fast → fake-green`,
`gate canary`, and the guard set in `scripts/gate/evidence.go`.

## P5 — Gates are immutable to the loop

Gate definitions, thresholds and lint rules cannot be edited to make a failure
go away. The loop is rewarded for green; that reward must not be reachable by
lowering the bar.

Gate definitions live in compiled Go (`scripts/gate/main.go`), not in a config
file, so weakening one is a code change that appears in review. The coverage
floor and the test baseline are ratchets: they may rise freely and only ever
fall deliberately.

**Guards, budgets and baselines are covered by this too.** They are the price
of the conveniences elsewhere in the pipeline — caching, tier splitting,
deferral. A loop that can delete the price gets the convenience for free, and
the system reverts to the state that made the guard necessary.

**And not cheap for the human either.** Lowering an acceptance criterion to
clear a red intent gate is the same act performed by hand, and it is never one
tap: the module returns to `draft`, a written justification is required, and it
is re-approved from ⓪ ([ADR-019](../docs/decisions.md#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write)).
The module document itself is the one artifact the implementation loop cannot
write at all.

Approving is cheap; **lowering a bar is expensive.** That asymmetry is the
design, not an oversight in the ergonomics.

## P6 — The user is in the loop

> Three times, and otherwise **off duty** —
> [ADR-015](../docs/decisions.md#adr-015--the-humans-default-state-is-non-involvement),
> checkpoint set per [ADR-020](../docs/decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional).

Three checkpoints may take a human's attention, plus arbitration. **There is no fifth.**

- **⓪ Is this the right thing to build?** Per module, in **business language**.
  Always asked — this is where the user's domain knowledge does work no model
  can do.
- **① Is this spec right?** **Conditional, and usually never reaches them.** It
  auto-ratifies when M1 is green, it introduces no invariant absent from the
  module, and its risk class is LOW or MEDIUM. It escalates on HIGH class, a new
  invariant, or M1 red — **and when it escalates, the screen states why.**
- **② Does it actually work?** Once per module, criterion by criterion, by
  observation. A half-ticked module does not close.
- **Arbitration.** A gate went red and the loop cannot resolve it alone.

Between them the human is **not involved** — not watching, not supervising, not
kept informed. Intervention starts because the system crossed a threshold and
called, never because someone was looking. An empty screen is the success state,
and the metric is **time away from the desk**, never time in the app.

Every avoidable interrupt is a **defect**. A tool that pages an operator about
something they cannot act on teaches them to ignore the pager, and then the
alarm that mattered is the one that gets missed.

> **The system cannot earn the right to interrupt before it has earned the right
> to correctly not interrupt.**

**Enforced by:** `gate verify` refuses to declare readiness while any spec or
module is `draft`; ② is blocked while M2 is red.

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

## P11 — Nothing is shown that does not change a decision

Every element on every surface must answer one question:

> **Does seeing this change the decision I am about to make?**

Failing it today: token counters, the agent's reasoning stream, diffs by
default, live logs. Each is engaging. None of them changes an answer.

**Two exceptions, and no more.**

- **Budget.** A spend figure is not evidence, but it decides whether to let the
  loop keep grinding.
- **The heartbeat.** An empty queue must say *when the system last confirmed it
  was working*, and stale silence becomes the alarm itself. A dead orchestrator
  and a quiet one produce the same empty screen; without attestation the
  success state is indistinguishable from a lie
  ([ADR-024](../docs/decisions.md#adr-024--silence-is-attested-not-assumed)).

Nothing else inherits this exemption by analogy. Budget decides whether to keep
spending; the heartbeat decides whether to believe the screen. No third thing
has a claim of that kind.

This is not concealment. The reasoning trace and the raw agent conversation are
available **in full and without restriction**, freely opened and freely closed.
The objection is to making them the *centre*, not to their existence. See
[ADR-021](../docs/decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar).

## P12 — Intent is a hierarchy, and each layer is the oracle below it

**Module → Spec → Feature + invariant → Code.** Every layer states intent for
the one beneath it, and is the only thing that layer can be judged against.

- A **module** fixes intent in **business language**. Its acceptance criteria
  are outcomes a non-technical person can observe. **Mechanism is forbidden
  there** — no endpoint, table, class or library name. Writing one collapses two
  layers into one and destroys the oracle.
- A **spec** fixes intent in **technical language**: `SPEC-…` ids, measurable
  requirements, invariants.
- **Gates** prove the code stayed faithful to both, by producing **evidence, not
  opinion**.

A module may produce at most **six ① checkpoints**. The budget is denominated in
*decisions*, not calendar time; duration is an output of it.

Every spec declares the `MOD-K` it serves, and **every invariant declares the
criterion it derives from.** An invariant with no declared parent is new intent,
and new intent is the human's
([ADR-025](../docs/decisions.md#adr-025--every-invariant-declares-the-criterion-it-derives-from)).

**Enforced by:** `Tier M` — **M1** (declarations exist and are in scope), **M2**
(an acceptance criterion no spec serves — the only detector of *"correct code,
but part of what was wanted was never built"*), **M3** (orphaned ratified specs).

Each splits into a **deterministic half that binds** and a **judged half that may
accuse but never absolve**
([ADR-023](../docs/decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)).
See also [ADR-018](../docs/decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy)
and [ADR-019](../docs/decisions.md#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write).

**And it does not close the top.** Nothing is the oracle of the module itself, so
the hierarchy **relocates** this class of defect rather than removing it — to the
cheapest place to have one, where it costs a conversation at ⓪ instead of a
rebuild ([ADR-026](../docs/decisions.md#adr-026--the-hierarchy-relocates-the-blind-spot-it-does-not-remove-it)).

## P13 — A verifier is never the producer

The model at **T2** may never be the model at **T1**. If they are configured the
same, the T2 gate is **void and the ladder halts** — it never reports green.

A model grading its own output is the exact failure this product exists to
correct. Permitting it *while displaying a green gate* would be worse than not
running the gate at all: a fake green carrying a credential.

Self-graded is not passed and not failed — it is `COULD NOT VERIFY`
([P3](#p3--fail-loud)).

**Enforced by:** identity check at ladder entry; canary asserting a same-model
configuration halts rather than passes.
See [ADR-022](../docs/decisions.md#adr-022--tiered-model-assignment-and-the-verifier-is-never-the-producer).

## P14 — A deterministic verdict is an exit code, and judgement reads it

If a check's output is an exit code, **a model does not run it.** The script
runs it, the exit code *is* the verdict, and the run writes one artifact
recording what ran and what it saw.

Judgement gets what is left — architectural intent, backward fidelity,
black-box exploration — and works under two bounds: it **reads the artifact and
runs no tools**, and it **reports rather than blocks a round**. Blocking
authority stays with the exit codes and the readiness verdict.

This is a measurement, not a preference: deterministic work put through a model
cost 10–30× as much and returned an unstable verdict — the same unchanged code
passing one round and failing the next. A layer whose green is unstable never
converges, so *"iterate until everything is green"* stops having a terminating
condition.

**Enforced by:** `scripts/gate` owning every exit-code check;
`gate verify` printing the artifact path alongside the judgement obligations;
the self-audit's **backbone leakage** check
([ADR-027](../docs/decisions.md#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it)).

---

## Conflict resolution

When the task and these principles disagree:

1. **Stop.** Do not pick one silently.
2. **Report** the conflict, with both readings stated plainly.
3. **Wait** for a human decision.
4. If the decision changes a principle, it lands in `docs/decisions.md` first.

A principle quietly bent once becomes a principle that no longer exists.
