# 🔁 How we work — the development loop

> **Status:** active · **Updated:** 2026-07-28
> The working method for **building RemLinkAgent** — which is also, since
> [ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent),
> **the product itself**. The full specification is [`loop-engineering.md`](loop-engineering.md).

We are shipping a system that makes one model's work provable to a different
model. Building it any other way would be an odd thing to do — so it runs on
this repository first, and we find out where it is annoying before anyone else
has to.

**What runs here today is the spec layer and the deterministic backbone.** The
module layer above it — business-language intent, `MOD-K` criteria and the
[Tier M](decisions.md#adr-019--three-intent-gates-and-a-module-document-the-loop-cannot-write)
gates M1/M2/M3 — is designed ([`module-layer.md`](../.rla/specs/module-layer.md))
and lands in [P1](roadmap.md#module-layer--tier-m--the-intent-hierarchy). Until
then this repository has **no M2**, which means it carries exactly the blind spot
M2 exists to close: everything can be green while a part of what was wanted was
never specified. Stated rather than hidden.

The cross-model half is [P1](roadmap.md#p1--orchestrator) too; until it lands,
those passes are run by hand — reading the evidence artifact, never re-running
the tools ([ADR-027](decisions.md#adr-027--a-deterministic-verdict-is-an-exit-code-and-judgement-reads-it)).

Everything here **runs today**:

```bash
go run ./scripts/gate fast      # after every change      — 5.8s measured
go run ./scripts/gate full      # at convergence          — 10.4s measured
go run ./scripts/gate verify    # full + checkpoint ①, before you test by hand
go run ./scripts/gate evidence  # the artifact a reviewer reads instead of re-running
go run ./scripts/gate timings   # measure the wall clock; -record sets the baseline
go run ./scripts/gate canary    # prove the gates still work
go run ./scripts/gate spec      # what is ratified, what is not
go run ./scripts/gate t0…t3     # one tier at a time, when that is what you want
```

`make fast` … `make verify` wrap these. On Windows: `.\scripts\make.ps1 fast`.

---

## The loop

```
        ┌─────────────────────────────────────────────┐
        │  write a spec  →  ① HUMAN RATIFIES  →  code │
        └──────────────────────┬──────────────────────┘
                               ▼
                    ┌──────────────────┐
              ┌────▶│  edit            │
              │     └────────┬─────────┘
              │              ▼
              │     ┌──────────────────┐   red
              └─────│  fast   ~6s      │───────┐
                    │  tiers 0–1       │       │
                    └────────┬─────────┘       │
                      green  ▼                 │
                    ┌──────────────────┐   red │
                    │  full   ~10s     │───────┘
                    │  tiers 0–3       │
                    └────────┬─────────┘
                      green  ▼
                    ┌──────────────────┐
                    │  judged pass     │  reads .rla/state/verify-*.json
                    │  reports, never  │  runs no tools, blocks nothing
                    │  blocks          │
                    └────────┬─────────┘
                             ▼
                    ┌──────────────────┐
                    │  verify          │
                    └────────┬─────────┘
                             ▼
                    ② HUMAN TESTS IT
```

**Modes are what you run; tiers are how checks are classified.** `fast` is the
one you invoke constantly, and it stays cheap because tiers 2 and 3 are not in
it — which it says out loud at the end of every run, by name.

Iteration budgets are the same as the product design: **`fast` generously
(20–30 attempts)** because it is cheap, **the heavy steps in `full` tightly
(2–3)** because they are not. Hitting the tight limit means stop and think, not
try harder.

---

## The gates

| Tier | Gate | What it catches |
| :--- | :--- | :--- |
| **0** | `gofmt` · `build` · `vet` | Doesn't compile, isn't formatted |
| **1** | `lint` | golangci-lint, including `nilerr` and `errcheck` |
| | `tests` | Unit tests — **and how many of them ran** |
| | `zero-touch-ai` | **Relay importing provider, agent, tool or crypto code** |
| | `secret-logging` | Credential-shaped identifiers passed to log calls |
| | `fake-green` | Tests that assert nothing |
| | `spec-hygiene` | Malformed spec artifacts |
| **2** | `spec-fidelity` | Ratified requirements with no implementation |
| | `coverage-ratchet` | Coverage regression against a committed floor |
| | `doc-links` | Broken relative links and heading anchors |
| | `licence-headers` | Missing AGPL headers |
| | `licence-boundary` | AGPL code inside Apache-2.0 `mobile/` |
| **3** | `race` | Data races |
| | `vulnerabilities` | Known CVEs in dependencies |

`zero-touch-ai` is the one worth pausing on. The project's flagship promise —
*AI traffic never reaches the relay* — is enforced as an **import-graph rule**,
not a paragraph. Relay packages cannot import provider, agent, tool or crypto
code. A promise in a document can be broken by accident on a Tuesday. An import
graph cannot.

### Three statuses, not two

| | Meaning |
| :--- | :--- |
| 🟢 `PASSED` | The gate ran and found nothing |
| 🔴 `FAILED` | The gate ran and found something |
| ⚠️ `COULD NOT VERIFY` | **The gate did not run** — a missing tool, a crash |

The third is the whole point. `COULD NOT VERIFY` is not a pass, and `verify`
refuses to declare readiness while any gate is in that state. A gate that
cannot run proves nothing, and treating that as success is precisely the
failure this system exists to prevent.

You will see it on Windows: the race detector needs cgo, so `t3` reports
`COULD NOT VERIFY` locally and CI runs it on Linux. Honest, and visible.

---

## Guards — what the exit code cannot say

`go test ./...` exits 0 when the suite passes. It also exits 0 when the suite
is empty. Every gate therefore reports **counts**, and the counts are held to
declared floors ([ADR-028](decisions.md#adr-028--counts-are-evidence-and-every-convenience-buys-a-guard)):

| Guard | Trips when | Result |
| :--- | :--- | :--- |
| **empty-run** | A gate declaring a floor examined less than it — 0 files scanned, 0 tests run | 🔴 |
| **unreported count** | A gate declares a floor and reports no number at all | ⚠️ |
| **baseline** | `tests_run` fell below [`.rla/test-baseline.txt`](../.rla/test-baseline.txt) | 🔴 |
| **step-budget** | A step exceeded the wall clock declared in its definition | 🔴 |
| **cached evidence** | A cache hit whose stored counts would fail the guards today | Hit rejected, gate re-runs |
| **stale artifact** | The report's fingerprint no longer matches the working tree | Not served |

**The baseline is a ratchet, like the coverage floor.** It may rise freely;
a fall is a deliberate, reviewable commit. That guard exists because
`go test` already exits non-zero when a package will not build — the loud case
is covered — while a suite eroding from 70 tests to 50 with everything still
green is invisible to every other gate here.

**Guards may only worsen a verdict.** A large count never rehabilitates a gate
that found a problem — the same rule the judged gates follow
([ADR-023](decisions.md#adr-023--a-judged-gate-may-add-a-finding-never-clear-one)),
one layer down.

---

## The evidence artifact

Every run writes `.rla/state/verify-<mode>-<fingerprint>.json`: per step, the
verdict, exit code, duration, cache status, evidence counts and guard results —
plus every check the mode deferred, and the obligations still owed to
judgement.

```bash
go run ./scripts/gate evidence            # the full-mode artifact for this tree
go run ./scripts/gate evidence -mode fast
```

**This is what a review pass reads.** It does not re-run the gates: re-running
a decided question costs an order of magnitude more and answers it less
reliably, because a model's verdict on the same unchanged input is not stable.

**Freshness is a fingerprint, not a timestamp.** Edit anything and `evidence`
refuses, naming the stale reports it declined to serve. A judged pass over a
report about a different tree proves nothing while looking like proof.

---

## Measured, not estimated

```bash
go run ./scripts/gate timings          # measure; writes .rla/state/
go run ./scripts/gate timings -record  # promote it to the committed baseline
```

The committed baseline is [`.rla/tool-timings.json`](../.rla/tool-timings.json);
recording a new one is deliberate, like lowering the coverage floor. Every run
prints its measured duration against both its mode budget and that baseline,
and says so when the two have drifted well apart.

The budgets — 2 minutes for `fast`, 30 for `full` — are ceilings set *after*
measuring 5.8 s and 10.4 s. A measurement climbing towards its ceiling is the
signal that something expensive moved into a layer meant to stay cheap.

The baseline names the four checks that could not run on the machine that
recorded it (`lint`, `licence-headers`, `race`, `vulnerabilities` — the tools
are not installed locally). A measurement that hid its own gaps would report
the first complete run as a regression.

---

## Specs and checkpoint ①

Anything beyond a bug fix starts with a spec in `.rla/specs/<feature>.md`
(copy [`_TEMPLATE.md`](../.rla/specs/_TEMPLATE.md)).

```markdown
---
id: provider-contract
title: Provider capability matrix and contract tests
status: draft          # → ratified, once a human agrees
phase: P0.17
---

## SPEC-provider-contract-01 — Tool schema compatibility is measured per provider
```

Then code cites the id:

```go
// SPEC-provider-contract-01: submits nested, enum and additionalProperties
// schemas to each provider and records what is accepted.
func (c *Client) probeToolSchema(...) { … }
```

**`gate full → spec-fidelity` fails until every ratified requirement is cited.**
It also fails on a citation of an id no spec declares — a typo, or a
requirement deleted while its marker stayed behind. Both are drift.

While a spec is `draft` its requirements are **not** enforced: a draft is a
proposal, and holding code to a plan nobody agreed to would invert the
checkpoint. But `gate verify` will not declare readiness while any spec is
still `draft`. Ratification is a human act — the fidelity gate proves the code
matches the plan, never that the plan is right.

Files containing SPEC-shaped strings as *test data* declare
`//gate:spec-fixtures` so fixtures never count as citations.

---

## What the gates cannot decide

`gate verify` prints these every run rather than omitting them, because an
unlisted obligation looks exactly like a met one:

| Owed | Who |
| :--- | :--- |
| **Backward spec diff** — behaviour in the code no requirement covers | AI review pass |
| **Architectural intent** — does this still match `PRINCIPLES.md` | AI review pass |
| **Black-box exploration** — adversarial probing without reading the source | AI review pass |
| **Spec correctness** — checkpoint ① | Human |
| **Interface test** — checkpoint ② | Human |

Today these passes are run by whoever is at the keyboard with an agent, using
the prompts in [`loop-engineering.md`](loop-engineering.md). Two rules apply
whether a human or [P1](roadmap.md#p1--orchestrator) drives them:

**The pass reads the artifact and runs no tools.** Hand it
`.rla/state/verify-full-<fingerprint>.json`, the diff and the spec. A reviewer
that starts running `go test` has moved deterministic work back into the
expensive, non-reproducible half.

**The pass reports; it does not block.** Blocking authority stays with the exit
codes and with `verify`. That is what lets judgement run **once**, at
convergence, instead of on every round — which is the difference between a
review pass and a nine-hour one.

---

## Gate integrity

The loop is rewarded for green gates, which creates a standing incentive to
weaken one rather than fix the code. Four defences:

**Definitions are compiled.** They live in `scripts/gate/main.go`, not a config
file. Weakening a gate is a code change that shows up in review.

**Canaries.** `gate canary` plants deliberate breakage and requires each gate to
catch it. A gate that reports green on broken input is blind, and says so:

```
🟢 zero-touch-ai      detected planted breakage
⚠️  race               no canary — this gate's ability to detect breakage is unproven
```

Four gates have no canary — `lint`, `doc-links`, `race`, `vulnerabilities`.
They wrap well-established external tools whose own correctness we take on
trust. That is a defensible position, and it is stated rather than hidden.

**The ratchets.** [`.rla/coverage-floor.txt`](../.rla/coverage-floor.txt) and
[`.rla/test-baseline.txt`](../.rla/test-baseline.txt) may rise freely; a drop
below either fails its gate. Lowering one is a deliberate, reviewable commit —
never a side effect.

**Guards are not editable either.** Floors, budgets and baselines are the price
of the conveniences elsewhere in the pipeline. A loop that can delete the price
gets the convenience for free, and the system reverts to the state that made
these guards necessary.

---

## The cache

A passing gate signs the content hash of its declared inputs, the ratified
requirement set, the gate version — **and the evidence it produced**.
Unchanged next round → skipped.

Only passes are cached. Failures and `COULD NOT VERIFY` always re-run.

**A hit is re-judged, not trusted.** Storing the counts is what makes a stale
silent green expire: the files that make a suite run zero tests never change,
so the signature keeps matching for ever. A hit whose recorded numbers would
fail the guards today is discarded and the gate runs again, saying so.

**`git commit` invalidates nothing** — committing does not change what a gate
read, so cache keys hash working-tree content and never a commit SHA. A cache
that clears on every commit is a cache nobody benefits from.

Bumping `gateVersion` in `scripts/gate/cache.go` invalidates everything: a gate
that just got stricter must not be skipped on the strength of an older, weaker
run.

```bash
go run ./scripts/gate full -no-cache    # force
go run ./scripts/gate clear-cache       # discard
```

---

## Adding a feature, end to end

```bash
cp .rla/specs/_TEMPLATE.md .rla/specs/my-feature.md
$EDITOR .rla/specs/my-feature.md          # goal, non-goals, invariants, SPEC ids
go run ./scripts/gate spec                # shows it as ① awaiting ratification

#   ① read it. Is the PLAN right? Then: status: ratified

go run ./scripts/gate full                # red — nothing implements it yet

# … implement, citing SPEC-my-feature-NN in the code …

go run ./scripts/gate fast                # after every change — seconds
go run ./scripts/gate full                # at convergence
go run ./scripts/gate evidence            # hand this to the review pass
go run ./scripts/gate verify              # → READY FOR INTERFACE TEST

#   ② use it. Does it actually work?
```

## Adopting existing code

`build-scaffolding.md` is a retro-spec: written after the code, to bring the
P0 scaffolding under the loop. That is the documented path for an existing
tree, and the project used it on itself. Scan, write drafts, mark anything that
contradicts `PRINCIPLES.md` as `⚠️ SUSPECTED DEVIATION`, ratify, then verify.

---

## The constitution

- [`.rla/PRINCIPLES.md`](../.rla/PRINCIPLES.md) — how this codebase is built.
  Not relaxable; changing one needs a superseding [ADR](decisions.md).
- [`.rla/SECURITY-BASELINE.md`](../.rla/SECURITY-BASELINE.md) — security
  invariants enforced while code is written, not discovered in review.

When a task and a principle conflict: **stop, report both readings, wait.** A
principle quietly bent once is a principle that no longer exists.

---

## In CI

`ci-go.yml → loop-gate` runs `canary` then `fast`, then tier 2, on every PR,
with tier 3's race detector covered by the Linux test matrix.

CI deliberately does **not** run `verify`: that command's job is to declare
human-readiness, and it fails while any spec is `draft` — which is the correct
state for a spec awaiting ratification, not a broken build.
