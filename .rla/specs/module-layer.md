---
id: module-layer
title: The module — intent in business language, and the three gates that enforce it
status: draft
phase: P1
created: 2026-07-27
ratified:
---

> ① **AWAITING RATIFICATION.** `gate verify` will refuse to declare readiness
> while this is `draft`. Read it and decide whether the *plan* is right — not
> whether code matches it — then set `status: ratified`.

## Goal

A ratified spec is an oracle for code, and that is not enough. When every spec is ratified, every Tier 0–3 gate is green and the code is entirely correct, nothing in the system can see that **a part of what was actually wanted was never specified.** There is no artifact above the spec to compare against.

This spec introduces that artifact. A **module** states intent in the language of the business, its acceptance criteria are outcomes a non-technical person can observe, and three gates (`M1`, `M2`, `M3`) hold the layers to each other. Reasoning: [ADR-018](../../docs/decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy), [ADR-019](../../docs/decisions.md#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write).

It also relocates critical questioning to where competence lives. At ⓪ the user reads *"a dealer with insufficient balance cannot transact and understands why"* and knows immediately whether that is right. At ① they would have read an idempotency invariant and taken it on trust.

## Non-goals

- **Not project management in the scheduling sense.** No dates, no estimates in time, no dependencies between modules, no burndown. The only budget is a **decision count**.
- **Not a requirements database.** A module is small enough to read on a phone. Anything needing a hierarchy of sub-modules is two modules.
- **Not a replacement for specs.** The module says *what* and *for whom*; the spec says *how it must behave*. Collapsing them is the failure this exists to prevent.
- **Not automatic decomposition.** The loop may *draft* a module; only a human may ratify one.
- **Not UI.** The queue, the modules tab and ② are [ADR-021](../../docs/decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar); their spec is separate.

## Invariants

- **The implementation loop can never write a module document.** Not to add a criterion, not to soften one, not to widen scope.
- Mechanism never appears in a module: no endpoint, table, class, library or field name.
- Every acceptance criterion is **observable** — a human can state what they would look at to confirm it.
- Every ratified spec serves at least one named module criterion, and violates no scope boundary.
- **② cannot start while M2 is red.**
- Lowering a criterion is never one action: it returns the module to `draft` and re-enters ⓪.
- A module that cannot state its scope boundary is not ratifiable — an unwritten boundary cannot be violated detectably.

## Requirements

## SPEC-module-layer-01 — The module artifact has a fixed, checkable shape

`.rla/modules/<slug>.md`, mirroring the layout the product will later create in a user's repository:

| Field | Rule |
| :--- | :--- |
| `id` | `MOD-<slug>`, equal to the filename |
| `status` | `draft` → `ratified` → `closed` |
| **Why it exists** | The problem, in the user's own words |
| **Acceptance criteria** | `K01…Kn`, each one sentence, observable |
| **Scope boundary** | **Included** and **Excluded**, both required and both non-empty |
| **Projected ①** | Estimated conditional checkpoints, shown at ⓪ |

A module missing any field, or with an empty Excluded list, fails `spec-hygiene` and is not ratifiable.

**Verified by:** parser unit tests per field; hygiene gate extended with module artifacts.

## SPEC-module-layer-02 — A module document is not writable by the implementation loop

Enforced **structurally**: a deny rule on `.rla/modules/**` plus a `PreToolUse` hook exiting non-zero. Not by instruction, not by prompt, not by convention.

A loop that can edit its own targets has no targets.

**Verified by:** a canary that attempts a module write from inside the loop and requires the attempt to be blocked. **A gate that cannot prove it blocks the write protects nothing** — an unproven block reports `COULD NOT VERIFY`, not green.

## SPEC-module-layer-03 — Acceptance criteria name outcomes, never mechanism

A criterion mentioning an endpoint, table, class, library, field or protocol is rejected at draft time with the offending term quoted.

*"A dealer can see their own balance correctly"* — accepted.
*"`GET /dealers/{id}/wallet` returns 200"* — rejected: that is a spec, one layer down.

**Verified by:** unit tests over a fixture corpus of accepted and rejected criteria; the Tier M model performs the judgement call and its rejection cites the term.

## SPEC-module-layer-04 — Every criterion is observable, and unobservable ones are rejected

For each criterion the adversarial verifier must be able to state *what a person would look at to confirm it*. Where it cannot, the criterion is rejected before ⓪ with that objection shown.

*"The user understands why"* is rejected. *"When the transaction is refused, the screen states that the balance is insufficient"* is accepted.

This is the mitigation for the known risk in [ADR-018](../../docs/decisions.md#adr-018--the-primitive-unit-is-a-module-and-intent-is-a-hierarchy): business language can be written vaguely enough to be unfalsifiable.

**The same pass also attempts completeness.** Given the module's *why it exists* statement, the verifier tries to **name a criterion the draft is missing**, and surfaces its attempt at ⓪.

This is the only pressure that exists at the top of the hierarchy. Nothing is the oracle of the module: M2 catches *a criterion with no spec*, and nothing catches *a requirement with no criterion* — the identical defect one layer up, made worse by the draft being written by an agent and merely approved by a human ([ADR-026](../../docs/decisions.md#adr-026--the-hierarchy-relocates-the-blind-spot-it-does-not-remove-it)). It cannot be closed. It can be pressured, and currently is not at all.

**Verified by:** fixture corpus; integration test asserting a vague criterion never reaches ⓪ unchallenged; a fixture module with a deliberately omitted criterion, asserting the completeness pass raises it as a suggestion rather than silently approving.

## SPEC-module-layer-05 — A module projects its ① checkpoints against a stated budget

The projection is computed and shown at ⓪. A module projected above the budget is reported as too large, with the suggestion to split, **before** approval.

The budget is denominated in **decisions, not calendar time**. Duration is an output of the decision count and is never an input to it.

> ⚠️ **The current figures — 6 ① and 8 total interrupts per module — are working hypotheses, not measurements.** They come from no data. They are configurable, they are displayed as such, and they are revisited against [P0.18.6](../../docs/roadmap.md#p0--scaffolding-gates--de-risking) and the trial. Treating them as invariants would be the exact error this spec warns about elsewhere: a confident number with nothing behind it.

**Verified by:** unit test on the projection; integration test asserting the ⓪ payload carries projection, limit, and the fact that the limit is a configured default rather than a rule.

## SPEC-module-layer-06 — M1 runs on every spec draft, and its binding half needs no model

Each spec declares which `MOD-K` criteria it serves, **with justification**. Like M2, M1 has two layers that fail differently ([ADR-023](../../docs/decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)):

**M1a — declaration validity. Deterministic, no model, binding.** The spec names at least one `MOD-K`, and every id it names exists in the module. Set membership. A spec naming nothing, or naming a criterion that does not exist, is an **orphan spec** and fails.

**M1b — scope. Judged, advisory.** Does the spec's subject fall in the module's **Excluded** list? A **scope violation** is raised as a finding; it can never clear an M1a failure.

M1 runs on every spec draft, making it the highest-frequency intent gate in the system. **M1a carrying the binding decision is what keeps it off the expensive tier** — the judged half fires only where a scope question actually arises.

M1 green is one of the conditions for automatic ① ratification ([ADR-020](../../docs/decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional)).

**Verified by:** unit tests for orphan and scope-violation paths; a test asserting M1b cannot turn a red M1a green; canary planting one of each.

## SPEC-module-layer-11 — Every invariant declares the criterion it derives from

Each invariant in a spec names the `MOD-K` it derives from. **An invariant with no declared parent is by definition new intent**, and its spec escalates to ① ([ADR-025](../../docs/decisions.md#adr-025--every-invariant-declares-the-criterion-it-derives-from)).

This replaces asking a model *"does this spec introduce a new invariant?"* — a question that cannot work, because [SPEC-03](#spec-module-layer-03--acceptance-criteria-name-outcomes-never-mechanism) forbids the module from naming mechanism, so **every technical invariant is absent from the module by construction.** The real question is entailment, and entailment asked open-endedly fails asymmetrically: generous models approve silently, strict ones escalate everything.

The parent check is **mechanical** (M1a). Whether the declared derivation is *honest* is M1b: advisory, never clearing.

**Early escalation rates will be high, and that is the intended behaviour.** Every escalation a human waves through is a labelled example of a legitimate derivation, and that log **is** the corpus the risk classifier is built from. The rate falling over time is the measurement of whether the classifier is definable at all.

**Verified by:** unit test asserting an undeclared invariant escalates; test asserting a declared-but-nonexistent parent fails M1a; integration test asserting the escalation log records the human's decision against each derivation.

## SPEC-module-layer-07 — M2 detects an uncovered criterion and blocks ②

At module candidate-complete, M2 maps every criterion to the specs serving it. **The check has two layers, and they are kept separate because they fail differently.**

**M2a — declaration coverage. Deterministic, no model.** Every spec declares the `MOD-K` it serves (SPEC-06). A criterion that **no spec declares** fails the gate and is named. This is set arithmetic; it cannot be argued with, and it is the floor M2 rests on.

**M2b — declaration honesty. Judged.** For each declared pairing, does the spec actually serve that criterion? This needs the Tier M model, and **its failure mode is silent**: a compliant-sounding spec accepted for a criterion it does not really serve produces a green M2 on an incomplete module — the exact outcome the gate exists to prevent, wearing a badge.

Therefore **M2a is binding and M2b is advisory**: M2b can raise a finding, never clear one. A green M2 means *"every criterion is declared, and nothing was flagged as a false declaration"* — never *"a model confirmed the coverage is real."* The honest confirmation happens at ② where a human reads the criterion→spec mapping (SPEC-10).

**This is the only mechanism in the system that fires while every spec is ratified, every Tier 0–3 gate is green and the code is entirely correct.** Every other gate compares code against a spec; here the defect is that the spec was never written.

While M2 is red, **② cannot start**. Interface-testing a knowingly incomplete product is not a test.

**Verified by:** integration test — a module with all specs green, all gates green, and one criterion deliberately undeclared; assert M2a red and ② blocked. Separate test asserting M2b can add a finding but can never turn a red M2a green. Canary on both shapes.

## SPEC-module-layer-08 — M3 detects orphaned ratified specs

A ratified spec whose module criterion was removed or renamed is reported. Drift can run in either direction, and a spec left serving nothing is dead weight that still costs gate time and still reads as intent.

**Verified by:** unit test removing a criterion from a ratified module and asserting the serving spec is reported.

## SPEC-module-layer-09 — Lowering a criterion returns the module to draft

M2 red may legitimately be closed by lowering an acceptance criterion — sometimes the goal really was too ambitious. This is **never a single action.** It:

1. returns the module to `status: draft`,
2. requires a **written justification** recorded in the module,
3. re-enters **⓪** for human approval.

Lowering an acceptance criterion to clear M2 is the human-hand version of what [P5](../PRINCIPLES.md#p5--gates-are-immutable-to-the-loop) forbids the loop, and one tap in a queue between meetings is the most dangerous place in the product for it.

**Verified by:** integration test asserting a lower-the-criterion action cannot resolve M2 in one step, and that the resulting module carries the justification.

## SPEC-module-layer-10 — ② is per module, by observation, and does not close half-ticked

The interface test presents the module's criteria one at a time. Each is ticked as **observed**, with a timestamp. A module with any criterion unticked stays in the queue and does not close.

Ticking is a **claim of observation**, which is why SPEC-04 exists: an unobservable criterion makes the tick meaningless.

**Verified by:** integration test asserting a partially ticked module does not close and reappears in the queue; test asserting each tick records a timestamp.

---

## Acceptance

Checkpoint ② — the interface test no gate can perform:

- A module drafted from a one-paragraph problem statement is **readable and judgeable by someone who does not code**, and they can say whether it is right.
- On a real module where everything is green, **M2 catches a genuinely forgotten requirement** — measured at least once on real work, not on a fixture.
- The rejection messages from SPEC-03 and SPEC-04 are useful rather than pedantic: the author agrees the rejected criterion was wrong.
- Attempting to lower a criterion **feels** like a decision with weight, not like dismissing a dialog.
- Across a month, interrupts per module land under eight, and the modules that exceed it are recognisably the ones that were too large.

## Open questions

**A spec with open questions is not ratifiable.** These block ①:

- **Who classifies a spec LOW / MEDIUM / HIGH, and against what?** [ADR-020](../../docs/decisions.md#adr-020--three-checkpoints-and-the-middle-one-is-conditional) makes auto-ratification depend on this class, so the classifier is load-bearing for a safety property, and it is currently undefined. A misclassified HIGH spec is silently approved. **Partly addressed:** [SPEC-11](#spec-module-layer-11--every-invariant-declares-the-criterion-it-derives-from) makes the escalation log the corpus it is built from, and [P0.18.5](../../docs/roadmap.md#p0--scaffolding-gates--de-risking) seeds that corpus by hand. It does not say what the classifier *is*.

- **Can a spec serve criteria across two modules?** Real work says yes; it would weaken M2's coverage arithmetic and make scope violations ambiguous. Undecided.
- **What happens to in-flight specs when a ratified module is reopened at ⓪?** Halt, continue, or invalidate — each is defensible and they differ materially in cost.

> ✅ **Answered — what counts as "an invariant absent from the module".** Resolved by [ADR-025](../../docs/decisions.md#adr-025--every-invariant-declares-the-criterion-it-derives-from): the question is not asked semantically at all. Every invariant declares its parent `MOD-K`, and an undeclared parent means new intent — mechanically. Kept here rather than deleted, because the reasoning that killed it is more useful than its absence.
