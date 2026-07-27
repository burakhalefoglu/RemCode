# 🔄 Loop Engineering Framework

> **Version:** 3.0 · **Updated:** 2026-07-27
> **Status:** Tier 0 ships in the MVP ([ADR-008](decisions.md#adr-008--loop-engineering-tier-0-ships-in-the-mvp)). Tiers 1–4 are deferred to [X6](vision-roadmap.md#x6--loop-engineering-tiers-14).
> **Premise:** "everything is green" means "it follows the rules" — not "it does the right thing".

The reference for RemLinkAgent's optional tiered QA pipeline: a system that requires the AI not merely to write code, but to demonstrate that the code matches an agreed specification.

**Read this as a design document, not a manual.** Only Tier 0 exists today. Tiers 1–4 rest on LLM judgement, have no determinism, and are deliberately the last thing the project will build — see [`vision-roadmap.md`](vision-roadmap.md#x6--loop-engineering-tiers-14) for why.

---

## Which gate runs when

| Gate | Tier | Runs | Cost | Status |
| :--- | :---: | :--- | :--- | :--- |
| Format, lint, type/compile check | **0** | After every edit | None — deterministic | ✅ **MVP** |
| Changed-file tests, architectural conformance, fast fidelity | 1 | Every fix iteration | Cheap model | 📋 X6 |
| Integration, coverage, invariants, full spec fidelity, fake-green hunt | 2 | Once per feature, at convergence | Strong model | 📋 X6 |
| Security, SAST, black-box exploration, mutation, property-based | 3 | Once, when a feature is candidate-complete | Large-context model | 📋 X6 |
| Self-audit, cross-feature exploration, blue-hat simulation | 4 | Periodic, manual | Varies | 📋 X6 |
| CVE scan | event | On dependency change | None | 📋 X6 |

---

## Tier 0 — what exists

The deterministic part, and the only part shipping in the MVP.

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

## Tiered verification — the design

### Why tiers

Running every check on every change is unaffordable in both time and tokens. Splitting by cost and frequency is what makes autonomous verification viable at all.

- **Tier 0 (instant)** — syntax, types, lint. Every edit. Native, free.
- **Tier 1 (inner loop)** — tests for changed files only, architectural conformance, quick fidelity check. Every fix iteration. Generous limit: 20–30 attempts.
- **Tier 2 (convergence)** — integration tests, coverage, domain invariants, full spec fidelity. Once per feature, after Tier 1 is green.
- **Tier 3 (heavy)** — mutation testing, fuzzing, SAST, black-box exploration. Once, when candidate-complete. Tight limit: 2–3 attempts.
- **Tier 4 (periodic)** — self-audit, cross-feature exploration, adversarial simulation. Manual.

### Selective regression cache

When a gate passes it signs the hashes of every file it read, the SPEC-ID set, and the config version. If all three are unchanged next round, the gate does not rerun.

**Cache keys never bind to a commit SHA.** `git commit` does not change the working tree, so it must not invalidate anything. Binding to commits would make the cache useless in normal development — the exact moment it needs to work.

---

## Gate integrity

The loop is rewarded for turning gates green. That creates a direct incentive to weaken gates rather than fix code, and the design has to assume it.

**Gate definitions, thresholds and lint rules are immutable to the loop.** It cannot edit them. Not "should not" — cannot.

**Canary testing.** Every hook must return a failure on empty input. A gate that reports success on nothing is broken, and the system needs to know that before trusting a pass.

**Fail-loud.** A crashed gate reports `COULD NOT VERIFY` — never `PASSED`. A green that cannot be substantiated is worse than a red, because it stops anyone looking.

**Fake-green hunting.** Detects tests with no assertions, coverage inflated by exercising code without checking behaviour, and SAST configured so narrowly it reports zero findings. If a gate cannot produce evidence, it counts as failed.

---

## Spec artifacts and fidelity

Before any code is written, the Architect model produces `.rla/specs/<feature>.md`. Every requirement gets a unique id: `SPEC-{feature}-NN`.

The fidelity gate diffs code against spec in **both directions**:

1. **Forward** — every SPEC id maps to an implementation. Missing → 🔴 FAIL.
2. **Backward** — behaviour in the code with no corresponding SPEC id → 🔴 DEVIATION.

The backward direction is the interesting one. Most architectural violations are not missing work; they are *extra* work — code nobody asked for, quietly added. Only a reverse diff finds it.

> Forward is tractable. Backward is genuinely hard and non-deterministic, which is why [X6](vision-roadmap.md#x6--loop-engineering-tiers-14) sequences forward first and treats backward as experimental.

---

## Black-box explorer

A subagent whose isolation is enforced at the tool-permission level: **it cannot read implementation source.** It only sees the public interface, and it tries to break it — malicious inputs, boundary conditions, unexpected sequences.

Findings become permanent regression tests.

The isolation is the whole point. An explorer that can read the implementation tests what the code does. One that cannot tests what the interface promises — and the gap between those two is where bugs live.

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
Tier 0 + 1: Coder implements, inner loop iterates
      ↓
Tier 2: Reviewer checks integration and spec fidelity
      ↓
Tier 3: mutation, fuzzing, black-box
      ↓
Full verify — evidence collected from cache
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

---

## Adopting into an existing codebase

You do not write specs for existing features by hand. RemLinkAgent scans the codebase and produces `status: draft` specs for each logical feature, marking anything that contradicts stated principles as `⚠️ SUSPECTED DEVIATION`. You review and ratify, then verification begins.

---

## The two jobs that stay human

There is no perfect system — only one that knows where its blind spots are. Two responsibilities cannot be delegated:

**1. Is the spec right?** The fidelity gate proves the code matches the plan. It cannot prove the plan is correct. For anything touching money, measurement or safety, questioning the specification's logic is yours alone.

**2. Does it actually work, and keep working?** Every gate green still does not confirm the end-to-end experience is right. Testing it, and watching production afterwards, is the outermost net — and it is human.

> When a production incident is found, it is not closed until the scenario exists as a permanent regression test.

---

*Reference document. Only Tier 0 is implemented; the rest is [X6](vision-roadmap.md#x6--loop-engineering-tiers-14).*
