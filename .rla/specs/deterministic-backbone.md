---
id: deterministic-backbone
title: One deterministic backbone, an evidence artifact, and guards on every count
status: draft
phase: P0.19
created: 2026-07-28
ratified:
---

## Goal

Split the pipeline by **who decides**, not only by how often it runs.

Everything whose output is an exit code belongs to one script, runs as one of
two modes, and produces one machine-readable artifact. Everything requiring
judgement reads that artifact and produces a report. The two halves stop
overlapping, and the expensive half stops rediscovering the project.

The measurement that motivates this is in
[`loop-engineering.md`](../../docs/loop-engineering.md#measured-first--the-cost-was-in-the-ceremony):
running deterministic tools *through* a model costs 10–30× running them
directly, and returns a verdict that is not reproducible — the same unchanged
code passing one round and failing the next. A layer whose green is unstable
never converges, so "keep going until everything is green" becomes an
instruction that cannot terminate.

The second half of the goal is smaller and sharper. An exit code answers
*"did what ran pass?"*. It cannot answer *"did anything run?"* — and both of
this project's real silent greens were of the second kind.

## Non-goals

- **Not a replacement for the tiers.** Tiers stay as classification: they say
  how often a check is worth running. Modes are what a person invokes.
- **Not a removal of the judged gates.** Backward fidelity, architectural
  intent and black-box exploration stay judged, because their output is not an
  exit code. What changes is that they read evidence instead of regenerating it.
- **Not a mutation, fuzzing or property-testing layer.** Those remain
  [X1](../../docs/vision-roadmap.md#x1--deeper-verification-tiers).
- **No hook that blocks every iteration.** The blocking point is the readiness
  verdict, not each turn of the loop.

## Invariants

- A deterministic gate's verdict is its exit code. Nothing in this backbone
  produces an opinion, a score or a recommendation.
- A guard may worsen a verdict, never improve one. Same rule as the judged
  gates ([ADR-023](../../docs/decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)),
  applied one layer down.
- A check that was not run is named. Silence and success must never look alike.
- A cached pass is re-judged against its recorded evidence, not trusted for
  having matched a signature.
- Freshness is a fingerprint, never a timestamp.
- Budgets are set from measurements. A budget invented before a measurement is
  a guess wearing a threshold's clothes.

## Requirements

## SPEC-deterministic-backbone-01 — Every run writes one machine-readable artifact

A run of any mode writes `.rla/state/verify-<mode>-<fingerprint>.json`
containing, per step: id, tier, verdict, exit code, duration, whether it came
from the cache, its evidence counts and its guard results — plus the checks the
mode deferred, and the obligations still owed to judgement.

The file name carries the fingerprint of the tree it describes, so a report
about a different tree cannot be opened as if it were current.

**Verified by:** `TestArtifactIsWrittenUnderItsFingerprint`,
`TestArtifactIsMachineReadable`.

## SPEC-deterministic-backbone-02 — Every gate reports how much it examined, and floors are enforced

Each check records counts — tests run, files scanned, requirements checked. A
check declaring a floor for a count fails when the count falls below it, and is
`COULD NOT VERIFY` when it declares a floor and then reports no count at all.

A check may additionally ratchet a count against a committed baseline, catching
erosion rather than collapse: a floor sees a suite fall from 70 tests to 0,
only a baseline sees it fall to 50.

Steps declare a wall-clock budget, and exceeding it is a red — a step that
outgrew its budget has changed which tier it belongs to.

**Verified by:** `TestEmptyRunGuardFailsAGateThatExaminedNothing`,
`TestUnreportedEvidenceIsUnverifiedNotPassed`,
`TestBaselineGuardCatchesErosionNotJustCollapse`, `TestMissingBaselineIsUnverified`,
`TestStepBudgetIsARedNotAWarning`, `TestGuardsNeverClearARed`,
`TestParseGoTestJSONSeparatesRunFromSkipped`.

## SPEC-deterministic-backbone-03 — A cached pass is re-judged, never trusted

The cache stores each pass together with the evidence that produced it. A hit
is served only when those recorded counts still clear the check's guards;
otherwise the gate runs again and the rejection is reported.

Without this the cache is where a silent green becomes permanent: the files
that make a suite run zero tests do not change, so the signature keeps matching
and the empty pass is served indefinitely.

**Verified by:** `TestCachedPassIsRejectedWhenItsEvidenceWouldFailToday`,
`TestCacheRoundTripsEvidenceThroughDisk`.

## SPEC-deterministic-backbone-04 — A reviewer is given fresh evidence or none

`gate evidence` returns the artifact for the current working tree, or refuses
and names the stale reports it declined to serve. A judged pass reads that file
and runs no tools.

Staleness is decided by fingerprint. A report about a tree that has moved on is
not weak evidence; it is evidence about something else, arriving with the
authority of a machine-generated file.

**Verified by:** `TestArtifactForAnotherTreeIsNotFound`.

## SPEC-deterministic-backbone-05 — The wall clock is measured, and drift is reported against a committed baseline

`gate timings` measures every check with the cache bypassed and writes the
result to `.rla/state/`. With `-record` it writes `.rla/tool-timings.json`, the
committed baseline — a deliberate act, like lowering the coverage floor.

Every run afterwards reports its measured duration against both its mode budget
and that baseline, and says so when the measurement has drifted well past it.

**Verified by:** `TestDriftIsMeasuredAgainstTheRecordedBaseline`,
`TestTimingsRoundTrip`.

## SPEC-deterministic-backbone-06 — Modes are the unit of running; tiers remain the unit of classification

`fast` runs tiers 0–1 and is what runs after every change. `full` runs 0–3 and
is what a readiness verdict rests on. A mode names every check it did not run,
in its output and in its artifact.

Individual tiers stay invocable for when only one is wanted, and full mode
defers nothing.

**Verified by:** `TestFastModeDefersTheExpensiveTiers`, `TestFullModeDefersNothing`.

---

## Acceptance

Checkpoint ② — observable without reading the source:

- `gate fast` finishes in seconds and its closing lines name every deferred
  check, so a green is never mistaken for a complete one.
- Deleting a test file and re-running `gate fast` turns the suite red on the
  baseline guard, **not** on a failing assertion.
- Running `gate fast` twice serves the second from cache; corrupting the
  recorded evidence of that entry makes it re-run instead of hit.
- `gate evidence` refuses after any edit until the run is repeated.
- `gate timings` prints each check's measured cost, and the mode totals are
  consistent with the sum of their steps.

## Open questions

- **Should the mode budget be a red rather than a warning?** Today a step
  budget fails and a mode budget only reports. The case for keeping it a report
  is that a slow machine is not a defect; the case against is that every
  threshold nobody enforces eventually stops being read.
- **What replaces the baseline when subtests move?** Counting subtests makes
  the number sensitive to refactors that split a table test. The ratchet
  absorbs growth, but a large legitimate reduction needs a deliberate lowering,
  and how often that happens is not yet measured.
- **Does the evidence artifact belong in `internal/verify` now or at
  [1.17](../../docs/roadmap.md#specs--gates)?** It is the same object the
  evidence bundle needs; building it twice would be the mistake.
