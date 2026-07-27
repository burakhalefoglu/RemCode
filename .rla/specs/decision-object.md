---
id: decision-object
title: Translate a red gate into a decision a person can make on a phone
status: draft
phase: P1
created: 2026-07-27
ratified:
---

> ① **AWAITING RATIFICATION.** `gate verify` will refuse to declare readiness
> while this is `draft`. Read it and decide whether the *plan* is right — not
> whether code matches it — then set `status: ratified`.

## Goal

The human's default state is **not involved** ([ADR-015](../../docs/decisions.md#adr-015--the-humans-default-state-is-non-involvement)). That is only defensible if, when the system does call, the call is answerable in about thirty seconds on a phone held in one hand.

Raw gate output is not that. A stack trace, a coverage delta and a reviewer's prose are evidence, not a decision. This spec defines the **decision object**: the single structured artifact the orchestrator produces when the loop cannot proceed alone, and the only thing permitted to interrupt a human.

**This is the product's principal engineering work** ([ADR-021](../../docs/decisions.md#adr-021--three-tabs-and-an-asynchronous-command-bar)) — not a presentation concern, and not the mobile client's job. The orchestrator produces it; every surface merely renders it.

## Non-goals

- **Not a summariser for the whole run.** One decision object per blocking condition, not a digest of everything that happened.
- **Not a replacement for the evidence bundle.** The raw bundle is always produced and retained; the decision object references it and never supersedes it.
- **Not UI.** Layout, typography and interaction belong to a later spec against the mockup. This defines the payload and its guarantees.
- **Not an auto-resolver.** Choosing between the options is the human's act. The orchestrator may not pick one because the user is slow to answer.
- **Not a general notification system.** Push transport is [P2](../../docs/roadmap.md#p2--relay-wss--nats). This defines what is *allowed* to become a notification.

## Invariants

- A blocking condition **always** yields exactly one decision object. Blocking silently is the failure mode this whole system exists to prevent ([P3](../PRINCIPLES.md#p3--fail-loud)).
- A decision object is the **only** artifact that may interrupt a human. There is no informational push.
- Every claim in a decision object is traceable to something that actually ran. Nothing is inferred to fill a field.
- An unknown is **stated**, never estimated into existence — an unpriceable option says so.
- No decision object contains a credential, a token or a key, in any field, ever.
- The raw evidence remains reachable from the decision object for as long as the decision is retained.

## Requirements

## SPEC-decision-object-01 — Every blocking condition produces exactly one decision object

When the loop cannot proceed without a human — a red gate, an exhausted iteration budget, an agent crash, a `COULD NOT VERIFY` on a required gate — the orchestrator emits one decision object. Zero is a defect. Two for the same `(feature, blocking condition)` pair is a defect: the second is deduplicated onto the first, not queued.

**Verified by:** unit test per blocking-condition kind; integration test asserting queue depth after a run with two independent failures and one duplicate.

## SPEC-decision-object-02 — The violated invariant is cited by id, or its absence is declared

The object names the ratified requirement that was broken, by `SPEC-…` id, together with the invariant text as written in the spec.

Where a failure maps to **no** ratified requirement, the object says so explicitly (`invariant: none — unspecified behaviour`) and is flagged as a **backward-fidelity finding**: behaviour exists that no requirement covers. It is never attributed to the nearest plausible id.

**Verified by:** unit tests for both branches; a test asserting no decision object ever carries an id absent from the loaded spec set.

## SPEC-decision-object-03 — The explanation is one sentence in the spec's own vocabulary

A single sentence, ≤ 200 characters, using the domain terms of the spec that was violated — not the terms of the toolchain. *"The coupon path subtracts before the floor check"* rather than *"assertion failed at pricing_test.go:212"*.

The file, line and assertion remain available in the referenced evidence bundle.

**Verified by:** unit test on length and single-sentence structure; review pass on vocabulary, recorded as a judgement gate.

## SPEC-decision-object-04 — Two or three options, never one, never more

One option is not a decision, it is a notification. More than three is a menu, and a menu on a phone is a deferral.

Three option kinds are always available to the constructor, and at least one non-fix option is always present:

| Kind | Meaning | Cost annotation |
| :--- | :--- | :--- |
| **Fix** | Change the code to satisfy the invariant | time · blast radius |
| **Relax** | Weaken the requirement | **states that it touches the module boundary and reopens ⓪** — never a silent edit ([SPEC-module-layer-09](module-layer.md)) |
| **Defer** | *"Leave it to me"* — stays queued, handled at the desk | none; this is an answer, not a deferral bug |

**Defer is a first-class answer, not a timeout.** A user who reads the decision and decides it needs a keyboard has decided. The object leaves the alarm state and stops re-notifying.

**Verified by:** unit test asserting `2 ≤ len(options) ≤ 3` for every emitted object; test that a relax option always carries the reopen-⓪ consequence; test that defer clears the notification state without resolving the condition.

## SPEC-decision-object-05 — Every option states its cost, or states that the cost is unknown

Each option carries a cost across the dimensions that apply: elapsed time, tokens, blast radius (files, public API, tests affected), and whether it leaves an invariant violated.

An unmeasurable dimension is rendered `unknown`. **Fabricating a plausible number is the defect this requirement exists to prevent** — a decision made against an invented cost is worse than one made against no cost at all.

**Verified by:** unit test asserting no numeric cost field is populated without a recorded measurement backing it.

## SPEC-decision-object-06 — Each object references a retained raw evidence bundle

Every decision object carries the id of the evidence bundle it was derived from. That bundle is retained for at least as long as the decision, and is reachable from every surface that renders the object.

A summary is lossy compression, and a wrong summary is more dangerous than raw output because it is easier to act on. The raw bundle is the appeal path.

**Verified by:** integration test asserting the referenced bundle resolves after the decision is recorded, and after the session ends.

## SPEC-decision-object-07 — Nothing but a pending checkpoint may interrupt

Only four things may enter the queue or raise a notification: **⓪** module approval, an escalated **①**, **②** an interface test, and **arbitration**. Progress updates, completed gates, streamed tokens, an agent's reasoning and successful runs generate **neither**.

An empty queue renders as a success state and **says the app can be closed**, in those words — never as an empty-state placeholder inviting the user to start something. It reports what is still running and **how long the user has been away from the desk**, because that is the metric this product optimises.

**Verified by:** integration test asserting notification count is zero across a full green run; test asserting no queue entry is created for any non-checkpoint event; widget test for the empty-queue copy.

## SPEC-decision-object-08 — Budget is the only non-decision datum on the queue surface

A spend figure changes exactly one decision — whether to let the loop keep grinding — and is therefore permitted alongside the queue.

Token counters, elapsed timers, per-agent activity and live logs are not, on that surface, by [P11](../PRINCIPLES.md#p11--nothing-is-shown-that-does-not-change-a-decision). No further exception is granted by analogy to this one.

**Verified by:** review pass against P11 for every field added to the queue payload, recorded per change.

## SPEC-decision-object-09 — Answering is idempotent and replay is rejected

A decision is recorded against `(object id, nonce)`. Re-submitting the same answer is a no-op returning the recorded outcome; submitting a different answer to an already-resolved object is rejected with an explicit conflict, never silently applied.

An object resolved from one device is resolved on all of them.

**Verified by:** unit test for the idempotent path, conflict path and replayed nonce; integration test across two paired devices.

## SPEC-decision-object-10 — No decision object carries a secret

Every field is passed through the same redaction used on log paths ([P1](../PRINCIPLES.md#p1--zero-touch-ai)) before the object is sealed. This includes option text and cost annotations, not only the explanation.

**Verified by:** unit test seeding credential-shaped values into gate output and asserting none survives into the emitted object; the `secret-logging` gate extended to cover the decision-object constructor.

## SPEC-decision-object-11 — Every object names who raised the finding

Arbitration objects record which tier and which model produced the objection — *"Kimi K3 raised this at 09:52:11"*. Deterministic objects name the gate.

Attribution is not decoration. Under [P13](../PRINCIPLES.md#p13--a-verifier-is-never-the-producer) a verifier may never be the producer, and an unattributed finding cannot be audited against that rule after the fact.

**Verified by:** unit test asserting attribution is populated on every emitted object; test asserting an object whose verifier equals its producer is refused rather than emitted.

## SPEC-decision-object-12 — An empty queue is timestamped, and stale silence is itself the alarm

The empty state carries **when the orchestrator last confirmed it was working**, alongside what is still in flight. Past a configured heartbeat age it degrades visibly, and past a second threshold it raises an alarm.

**This is the one interrupt permitted with no decision attached**, because its content *is* the decision: *stop trusting the silence.* It is [P11](../PRINCIPLES.md#p11--nothing-is-shown-that-does-not-change-a-decision)'s second and final exception ([ADR-024](../../docs/decisions.md#adr-024--silence-is-attested-not-assumed)).

A dead orchestrator, a sleeping machine and a dropped notification all render the same empty queue as a healthy one. An interface still reporting *"2 modules running"* nine hours after the process died is not merely unhelpful — it is a lie shaped like good news, and it is the single failure that would discredit every other claim the product makes.

Distinct from [ADR-009](../../docs/decisions.md#adr-009--the-cli-host-must-be-reachable), which covers unreachability that is *known*. This covers the case where nothing is known at all.

**Verified by:** integration test freezing the heartbeat and asserting the empty state degrades and then alarms; test asserting the alarm fires with no decision object attached; test asserting a healthy empty queue always renders a timestamp.

---

## Acceptance

Checkpoint ② — the interface test no gate can perform:

- A real red gate on a real feature produces an object the maintainer can act on **from a phone, away from the desk, without opening the repository.**
- The chosen option, once applied, actually resolves the condition — measured over at least ten real decisions before this is considered met.
- Across a full green run the phone stays silent, and that silence is legible as success rather than as a broken connection.
- When an option's cost is `unknown`, that is visibly more useful than a confident wrong number would have been.

## Open questions

**A spec with open questions is not ratifiable.** These block ①:

- **Who constructs the options — the arbitration model, or deterministic rules?** [ADR-017](../../docs/decisions.md#adr-017--cheap-models-grind-an-expensive-model-arbitrates) puts arbitration on the expensive model, which implies the model. But a model-authored option set can be plausible and wrong, and SPEC-05 forbids invented costs. A hybrid — deterministic costs, model-authored option text — is likely, and is not yet decided.
- **What happens when a decision object cannot be constructed?** The failure must be loud, but "we could not explain why this is broken" is itself a decision a human has to receive, and its shape is undefined.
- **Retention.** SPEC-06 says the evidence bundle outlives the decision, without saying by how long. Interacts with the relay retention policy ([P2.9](../../docs/roadmap.md#p2--relay-wss--nats)).
- **Does an unanswered decision expire?** An approval that expires fails closed. A decision that expires has no obvious safe default, since the loop is already blocked.
