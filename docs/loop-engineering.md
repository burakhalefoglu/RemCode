# 🔄 Loop Engineering Framework

> **Version:** 4.1 · **Updated:** 2026-07-28
> **Status:** ⭐ **This is the product.** Promoted from deferred vision item to core specification by [ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent).
> **Premise:** "everything is green" means "it follows the rules" — not "it does the right thing".

The system that requires an AI not merely to write code, but to demonstrate that the code matches an agreed specification — and that puts a **different vendor's model** in the seat that judges it.

**Two halves, with different levels of certainty.**

**Deterministic gates** — lint, types, tests, coverage ratchet, spec fidelity, fake-green detection, canaries. No tokens, no judgement. **These already run**: this repository is built with them ([`development-loop.md`](development-loop.md), [ADR-011](decisions.md#adr-011--the-project-is-built-with-its-own-loop)).

**Cross-model verification** — a different model attempting to disprove finished work against the spec. Judgement-based, not deterministic, and its weakness is real: two models can share a blind spot. What justifies building it anyway is measurement, not theory — a project one model declared complete was put through a verification pass driven by another, which found missing translations, unwritten tests, a coverage regression, and real bugs.

> **Why a model vendor cannot ship this.** The honest output of cross-verification is sometimes *"our own model got this wrong."* Only a party that sells no model can credibly arbitrate between models.

---

## Measured first — the cost was in the ceremony

The division of labour here is by **who decides**, not only by how often a check runs. That distinction came from a measurement, and it is worth stating before the design that rests on it ([ADR-027](decisions.md#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it)).

On the codebase where this method was first run by hand, every deterministic tool was timed individually. The complete deterministic pass — lint, types, format, static security rules, the change-scoped tests — took **about two minutes**. The same tools, invoked by a model acting as a gate, took **ten to twenty-one minutes per gate**. One round of forty-one gate runs consumed roughly **nine hours** and ended without a verdict.

Two things were producing that bill, and neither was verification:

**Context rediscovery.** The overwhelming majority of a gate's time went not to running a tool but to working out, again, which tool to run and why. Every gate relearned the project on every round.

**Judgement production.** The gate reported its own opinion rather than the tool's exit code — and the same unchanged code passed one round and failed the next. Four such reversals were recorded. A layer whose verdict is unstable never converges, which makes *"iterate until everything is green"* an instruction with no terminating condition.

**The rule that follows.** Before handing a check to a model, ask whether its output is an exit code. If it is, a model must not run it: the cost rises by an order of magnitude and the verdict stops being reproducible.

**And this repository's own numbers**, measured on 2026-07-28 (Windows, Go 1.26), recorded in [`.rla/tool-timings.json`](../.rla/tool-timings.json) by `gate timings`:

| Mode | Measured | Budget | Of budget |
| :--- | ---: | ---: | ---: |
| `fast` — tiers 0–1, after every change | **5.8 s** | 2 min | 5% |
| `full` — tiers 0–3, at convergence | **10.4 s** | 30 min | 1% |

Four checks could not run on that machine — `lint`, `licence-headers`, `race`, `vulnerabilities` all need tooling absent locally — and the baseline names them, because a measurement that hides its gaps reports the first complete run as a regression.

The budgets are ceilings set after measuring, not targets. The number that matters is the measurement, and its **drift** is the early warning: a layer that was seconds and now takes half its budget did not get slower by accident.

---

## Verdict, and who issues it

| | **Deterministic backbone** | **Judged layer** |
| :--- | :--- | :--- |
| Runs | `scripts/gate` (P1: `internal/verify`) | A model, a different vendor's |
| Scope | Lint, types, tests, coverage, spec fidelity (forward), structural invariants, licence rules, CVEs, guards, Tier M set arithmetic | Architectural intent, backward spec diff, black-box exploration, M1b/M2b honesty |
| Verdict | **The tool's exit code** | A report |
| Frequency | Every change (`fast`) · at convergence (`full`) | Once, at a defined trigger |
| Authority | Blocks | Advises |

**Two rules, and both are cost decisions before they are design ones.**

**A judged pass reads the artifact; it runs no tools.** Re-running what a script already decided costs an order of magnitude more and answers a question that was already answered reproducibly. A reviewer invoking `go test` is a design regression, not a thorough one.

**A judged pass does not block the iteration.** Its output is a finding. Blocking authority lives with the exit codes and with the readiness verdict — which is also why the loop can keep moving while judgement happens once, at the end, rather than on every round.

This is [ADR-023](decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one) applied one layer down: judgement may accuse, never absolve.

---

## Two modes, four tiers

Tiers classify: they say how often a check is worth running. **Modes are what a person actually invokes.** Conflating the two produces a pipeline nobody runs, because every invocation costs as much as its most expensive member.

| | `fast` | `full` |
| :--- | :--- | :--- |
| When | After every change | At convergence, before anyone is asked for anything |
| Contains | Tiers 0–1 | Tiers 0–3 |
| Excludes | Coverage, licences, race, CVEs — **named in the output, every run** | Nothing |
| Measured / budget | 5.8 s / 2 min | 10.4 s / 30 min |
| Verdict means | This round is finished | The readiness verdict may be issued |

**Deferred work is named, not implied.** A check that was skipped and a check that passed are indistinguishable in a green summary, so `gate fast` closes by listing every gate it did not run. *"The fast layer is green"* says nothing about data races, and the output says so in those words.

### Which gate runs when

| Gate | Tier | Mode | Runner | Status |
| :--- | :---: | :--- | :--- | :--- |
| Format, lint, type/compile check | **0** | fast | Script — exit code | ✅ **Working** |
| Changed-file tests, structural conformance, fake-green, guards | **1** | fast | Script — exit code | ✅ **Working** |
| Full tests, coverage ratchet, forward spec fidelity, licences | **2** | full | Script — exit code | ✅ **Working** |
| Race detector, CVE scan | **3** | full | Script — exit code | ✅ **Working** |
| **Cross-model review** — reviewer disproves against the spec | **2** | once, at convergence | Model — reads the artifact | 🚧 **P1** |
| Backward spec diff — behaviour no requirement covers | 2 | once, at convergence | Model — reads the artifact | 🚧 P1 |
| Mutation, fuzzing, black-box exploration | 3 | once, candidate-complete | Model | 📋 X1 |
| Self-audit, cross-feature exploration | 4 | periodic, manual | Human, with a model | 📋 X1 |

The script rows run today via `go run ./scripts/gate fast|full`. The model rows are what P1 adds — see [`roadmap.md`](roadmap.md#p1--orchestrator).

---

## Tier 0 — the innermost loop

Deterministic, and running today.

After every agent edit: format, lint, type-check or compile. Sub-second, no tokens, no judgement. Failures return to the agent loop as tool results, so it fixes its own mistakes before proceeding.

```
agent edits a file
      ↓
Tier 0: format → lint → typecheck
      ↓
   pass? ─ yes → continue
      │
      └─ no → feed diagnostics back → agent fixes → rerun
```

Small, but it is the difference between an agent that produces plausible-looking code and one that produces code that at least compiles. And because it costs nothing, it can run on every single edit rather than at checkpoints.

---

## Guards — what an exit code cannot say

An exit code answers *"did what ran pass?"*. It cannot answer *"did anything run?"* — and both silent greens found in the field were of the second kind: a suite selecting **zero tests** and exiting 0, and a collection error dropping **all 18,034 tests** behind innocent-looking output.

So every gate also reports counts, and the counts are held to declared floors ([ADR-028](decisions.md#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard)).

| Guard | Condition | Verdict |
| :--- | :--- | :--- |
| **Empty run** | A check declaring a floor examined less than it | 🔴 FAILED |
| **Unreported count** | A check declares a floor and reports no count | ⚠️ COULD NOT VERIFY |
| **Baseline** | Tests run fell below the committed baseline | 🔴 FAILED |
| **Step budget** | A step exceeded its declared wall clock | 🔴 FAILED |
| **Cached evidence** | A cache hit whose recorded counts would fail today | Hit rejected, gate re-runs |
| **Stale report** | An artifact's fingerprint ≠ the working tree's | ⚠️ Not served at all |
| **Silent skip** | A deferred check is unnamed in the report | Structurally impossible — the runner emits them |

**Guards may only worsen a verdict.** A passing count never rehabilitates a gate that found something.

**On collection errors.** The Python original guards them separately, because a collection failure there can drop a suite while the command still looks calm. In Go a package that will not build makes `go test` exit non-zero, so the loud case is already covered by the exit code. What is *not* covered is the quiet one — a suite eroding from 70 tests to 50 while every remaining test passes — and that is exactly what the baseline ratchet is for. Porting the guard list verbatim would have added a check that could never fire and omitted the one that can.

**Why guards exist at all:** every convenience in this system — caching, tier splitting, deferring — trades certainty for speed, and each is only defensible while the counts are still checked. Guards are the price of the speed, which is why the loop cannot edit them ([P5](../.rla/PRINCIPLES.md#p5--gates-are-immutable-to-the-loop)).

---

## The evidence artifact

Every run writes `.rla/state/verify-<mode>-<fingerprint>.json`: per step, its verdict, exit code, duration, cache status, evidence counts and guard results — plus the checks the mode deferred and the obligations still owed to judgement.

This file is what the judged layer reads. It is the direct answer to context rediscovery: the reviewer is handed what ran and what it saw, instead of rediscovering the project and re-running its tools.

**Freshness is a fingerprint, never a timestamp.** The file name carries the hash of the tree it describes, so a report about code that has since changed cannot be opened as though it were current. `gate evidence` serves the artifact for the working tree or refuses, naming the stale reports it declined to serve — a judged pass over stale evidence proves nothing while looking like proof.

---

## The cache, and how a silent green becomes permanent

When a gate passes it signs the content hashes of every file it read, the ratified requirement set, the gate version — **and the evidence it produced**.

That last component is not decoration. A suite that runs zero tests does not change any file, so the signature keeps matching and the empty pass is served indefinitely. The cache is where a silent green goes to become permanent unless every hit is re-judged against its recorded counts. It is, and a hit whose numbers would fail today is discarded and the gate re-runs.

**Cache keys never bind to a commit SHA.** `git commit` does not change the working tree, so it must not invalidate anything. Binding to commits would make the cache useless in normal development — the exact moment it needs to work.

At 7 seconds, the fast layer's cache is a convenience rather than a necessity. The heavy checks are where it earns its keep, and knowing which is which is one more reason to measure.

---

## Gate integrity

The loop is rewarded for turning gates green. That creates a direct incentive to weaken gates rather than fix code, and the design has to assume it.

**Gate definitions, thresholds, guards and baselines are immutable to the loop.** It cannot edit them. Not "should not" — cannot. The list includes the coverage floor, the test baseline, step budgets and the guard definitions themselves: if the loop can delete the price of a convenience, the convenience becomes free and the system quietly reverts.

**Canary testing.** Every gate must return a failure on deliberately broken input. A gate that reports success on nothing is broken, and the system needs to know that before trusting a pass.

**Fail-loud.** A crashed gate reports `COULD NOT VERIFY` — never `PASSED`. A green that cannot be substantiated is worse than a red, because it stops anyone looking.

**Fake-green hunting.** Detects tests with no assertions, coverage inflated by exercising code without checking behaviour, and SAST configured so narrowly it reports zero findings. If a gate cannot produce evidence, it counts as failed.

---

## Spec artifacts and fidelity

Before any code is written, the Architect model produces `.rla/specs/<feature>.md`. Every requirement gets a unique id: `SPEC-{feature}-NN`.

The fidelity gate diffs code against spec in **both directions**:

1. **Forward** — every SPEC id maps to an implementation. Missing → 🔴 FAIL.
2. **Backward** — behaviour in the code with no corresponding SPEC id → 🔴 DEVIATION.

The backward direction is the interesting one. Most architectural violations are not missing work; they are *extra* work — code nobody asked for, quietly added. Only a reverse diff finds it.

> Forward is tractable and deterministic — it ships today. Backward splits: the **structural** part (a new endpoint, table, dependency or package that no spec declares) is set arithmetic and belongs in the script; the **semantic** part — *does this code really satisfy that requirement* — needs judgement and becomes a reviewer task in [P1](roadmap.md#p1--orchestrator).

---

## Black-box explorer

A subagent whose isolation is enforced at the tool-permission level: **it cannot read implementation source.** It only sees the public interface, and it tries to break it — malicious inputs, boundary conditions, unexpected sequences.

Findings become permanent regression tests.

The isolation is the whole point. An explorer that can read the implementation tests what the code does. One that cannot tests what the interface promises — and the gap between those two is where bugs live.

This is one to keep in the judged layer: its output is not an exit code, and no amount of scripting produces it. What it is not is frequent — once, when a feature is candidate-complete. Running an explorer every round is the most expensive form of the mistake this document opens with.

---

## Observability and security gates

**No secrets in logs.** Static analysis rejects raw PII or credential interpolation into log statements.

**Dependency CVEs.** A change to `go.mod` or `pubspec.yaml` triggers a scan automatically.

**Project invariants.** Rules like "no query without a tenant filter" or "cross-tenant access returns 404" are enforced at write time via PreToolUse hooks, blocking the edit rather than reporting it later.

---

## Daily flow — adding a feature

```
/new-feature <name>
      ↓
Architect writes the spec (status: draft) and STOPS
      ↓
① HUMAN CHECKPOINT — you read and ratify it on your phone
      ↓
implement → gate fast → implement → gate fast …   (seconds per round)
      ↓
gate full — every tier, nothing deferred
      ↓
judged pass — reads the artifact, runs no tools, reports
      ↓
② HUMAN CHECKPOINT — "ready for UI test" notification
```

Request format:

```text
/new-feature dealer-commission-deduction

Goal: deduct commission from the dealer balance when a transaction completes.

Invariants:
- Balance can never go negative.
- The same transaction cannot be deducted twice (idempotent).

Acceptance:
- Insufficient balance rejects the transaction.
- An audit record is written.

Produce the spec first and wait for my approval.
```

**Iteration budgets.** `fast` gets a generous limit (20–30 rounds) because it is cheap. The heavy steps in `full` get a tight one (2–3) because they are not. Hitting the tight limit means stop and reconsider the spec — not try harder.

---

## Adopting into an existing codebase

You do not write specs for existing features by hand. RemLinkAgent scans the codebase and produces `status: draft` specs for each logical feature, marking anything that contradicts stated principles as `⚠️ SUSPECTED DEVIATION`. You review and ratify, then verification begins.

**Measure before any of it.** Time each deterministic tool on the target codebase first; the tier boundaries and budgets are downstream of those numbers, and setting them from intuition is how a pipeline ends up defending a cost nobody checked.

---

## Porting this to another project

The order is not decorative. Skipping ahead — particularly to step 4 — reproduces the nine-hour run described at the top.

1. **Measure.** Time every deterministic tool individually, and time an agent doing the same work. Do not estimate. No tier decision is made before those two numbers sit side by side.
2. **Push everything deterministic into one runner.** Anything that produces an exit code belongs to the script, not to a model.
3. **Split into modes.** `fast` and `full`, with budgets set from step 1 and the measurement committed.
4. **Relax the blocking hook last.** Only after 1–3 does the per-round block become unnecessary; the blocking point moves to the readiness verdict.
5. **Pair every relaxation with a guard.** Otherwise the speed is bought with certainty, which is the one trade this system exists to refuse.

---

## Self-audit — the three numbers

Monthly, and after any major refactor. The point is not to admire the system but to catch it reverting:

**Duration drift.** What fraction of its budget does the fast layer now consume, against the committed baseline? Growth means work was quietly added to the layer that runs a hundred times a day.

**Backbone leakage.** Is any judged pass running tools? A reviewer that calls `go test` has taken deterministic work back into the judged layer, and that single change restores both the 10–30× cost and the unstable verdict.

**Guard liveness.** Trip each guard deliberately — delete a test, break a step's budget, corrupt a cached count — and confirm each turns red. Guards are the price of the conveniences; when one dies, only a deliberate violation reveals it.

Every regression here arrives as a convenience. *"Just ask the model to check it, that's quicker"* is the first step of the nine-hour run, every time.

---

## The jobs that stay human

There is no perfect system — only one that knows where its blind spots are. Three responsibilities cannot be delegated:

**1. Is the spec right?** The fidelity gate proves the code matches the plan. It cannot prove the plan is correct. For anything touching money, measurement or safety, questioning the specification's logic is yours alone.

**2. Does it actually work, and keep working?** Every gate green still does not confirm the end-to-end experience is right. Testing it, and watching production afterwards, is the outermost net — and it is human.

**3. Measuring.** A system cannot time itself into telling you it costs too much. Every structural decision in this document came from someone opening a stopwatch instead of reasoning about what ought to be fast, and that reflex is the one thing here that cannot be automated — it is what does the automating.

> When a production incident is found, it is not closed until the scenario exists as a permanent regression test — and until one further question is answered: *was there a gate that should have caught this, and why didn't it?* If the answer is "it was green but it examined nothing", the fix is a new guard, not only a new test. Both guards in the table above were found exactly that way.

---

*The deterministic gates run today ([`development-loop.md`](development-loop.md)). Cross-model verification is [P1](roadmap.md#p1--orchestrator). Heavy tiers are [X1](vision-roadmap.md#x1--deeper-verification-tiers).*
