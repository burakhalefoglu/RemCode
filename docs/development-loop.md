# 🔁 How we work — the development loop

> **Status:** active · **Updated:** 2026-07-27
> The working method for **building RemLinkAgent** — which is also, since
> [ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent),
> **the product itself**. The full specification is [`loop-engineering.md`](loop-engineering.md).

We are shipping a system that makes one model's work provable to a different
model. Building it any other way would be an odd thing to do — so it runs on
this repository first, and we find out where it is annoying before anyone else
has to.

The deterministic half described below **works today**. The cross-model half is
[P1](roadmap.md#p1--orchestrator); until it lands, those passes are run by hand.

Everything here **runs today**:

```bash
go run ./scripts/gate t0        # after every edit        — seconds
go run ./scripts/gate t1        # every fix iteration     — ~10s
go run ./scripts/gate t2        # at convergence          — ~30s
go run ./scripts/gate t3        # candidate-complete      — minutes
go run ./scripts/gate verify    # before you test by hand
go run ./scripts/gate canary    # prove the gates still work
go run ./scripts/gate spec      # what is ratified, what is not
```

`make t0` … `make verify` wrap these. On Windows: `.\scripts\make.ps1 t0`.

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
              │     │  T0  seconds     │───────┐
              │     └────────┬─────────┘       │
              │       green  ▼                 │
              │     ┌──────────────────┐   red │
              └─────│  T1  ~10s        │───────┤
                    └────────┬─────────┘       │
                      green  ▼                 │
                    ┌──────────────────┐   red │
                    │  T2  ~30s        │───────┤
                    └────────┬─────────┘       │
                      green  ▼                 │
                    ┌──────────────────┐   red │
                    │  T3  minutes     │───────┘
                    └────────┬─────────┘
                      green  ▼
                    ┌──────────────────┐
                    │  verify          │
                    └────────┬─────────┘
                             ▼
                    ② HUMAN TESTS IT
```

Iteration budgets are the same as the product design: **T1 generously (20–30
attempts)** because it is cheap, **T3 tightly (2–3)** because it is not. Hitting
the T3 limit means stop and think, not try harder.

---

## The gates

| Tier | Gate | What it catches |
| :--- | :--- | :--- |
| **0** | `gofmt` · `build` · `vet` | Doesn't compile, isn't formatted |
| **1** | `lint` | golangci-lint, including `nilerr` and `errcheck` |
| | `tests` | Unit tests |
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

**`gate t2 → spec-fidelity` fails until every ratified requirement is cited.**
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
the prompts in [`loop-engineering.md`](loop-engineering.md). [P1](roadmap.md#p1--orchestrator)
automates them by handing the diff, the spec and the gate output to a **different
vendor's** model. The obligation is identical either way; only the automation
changes — which is exactly why doing it by hand first was worth the effort.

---

## Gate integrity

The loop is rewarded for green gates, which creates a standing incentive to
weaken one rather than fix the code. Three defences:

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

**The coverage ratchet.** `.rla/coverage-floor.txt` may rise freely; a drop
below it (beyond 0.5% rounding slack) fails the gate. Lowering it is a
deliberate, reviewable commit — never a side effect.

---

## The cache

A passing gate signs the content hash of its declared inputs plus the ratified
requirement set plus the gate version. Unchanged next round → skipped.

Only passes are cached. Failures and `COULD NOT VERIFY` always re-run.

**`git commit` invalidates nothing** — committing does not change what a gate
read, so cache keys hash working-tree content and never a commit SHA. A cache
that clears on every commit is a cache nobody benefits from.

Bumping `gateVersion` in `scripts/gate/cache.go` invalidates everything: a gate
that just got stricter must not be skipped on the strength of an older, weaker
run.

```bash
go run ./scripts/gate t2 -no-cache    # force
go run ./scripts/gate clear-cache     # discard
```

---

## Adding a feature, end to end

```bash
cp .rla/specs/_TEMPLATE.md .rla/specs/my-feature.md
$EDITOR .rla/specs/my-feature.md          # goal, non-goals, invariants, SPEC ids
go run ./scripts/gate spec                # shows it as ① awaiting ratification

#   ① read it. Is the PLAN right? Then: status: ratified

go run ./scripts/gate t2                  # red — nothing implements it yet

# … implement, citing SPEC-my-feature-NN in the code …

go run ./scripts/gate t0                  # after each edit
go run ./scripts/gate t1                  # each iteration
go run ./scripts/gate t2                  # at convergence
go run ./scripts/gate t3                  # when candidate-complete
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

`ci-go.yml → loop-gate` runs `canary` then tiers 0–2 on every PR, with tier 3's
race detector covered by the Linux test matrix.

CI deliberately does **not** run `verify`: that command's job is to declare
human-readiness, and it fails while any spec is `draft` — which is the correct
state for a spec awaiting ratification, not a broken build.
