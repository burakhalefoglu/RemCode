---
id: <feature-slug>
title: <one line, what this delivers>
status: draft
phase: <P0.7 | P1.10 | M1.3 | …>
created: <YYYY-MM-DD>
ratified:
---

> Copy to `.rla/specs/<feature-slug>.md`. `id` **must** equal the file name.
> Delete this quote block.

## Goal

What problem this solves, in the user's terms. Not the implementation.

## Non-goals

What this deliberately does not do. The most useful section in the file —
it is what stops scope creep from looking like progress later.

## Invariants

Properties that must hold no matter how it is built. These become tests.

- …

## Requirements

Each requirement gets an id. **`gate t2 → spec-fidelity` fails until every one
of them is cited from code or tests** as `SPEC-<feature-slug>-NN`.

Write them so a reader can tell whether one is met. "Handles errors well" is
not a requirement; "an unknown provider id exits non-zero with the list of
configured providers" is.

## SPEC-<feature-slug>-01 — <short imperative title>

What must be true. Include the observable behaviour.

**Verified by:** unit test / integration test / gate name.

## SPEC-<feature-slug>-02 — <short imperative title>

…

---

## Acceptance

How a human confirms this at checkpoint ② — the interface test that no gate
can perform.

- …

## Open questions

Anything unresolved. **A spec with open questions should not be ratified** —
ratification means the plan is agreed, and an open question means it is not.

- …
