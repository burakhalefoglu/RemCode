---
id: acp-orchestrator
title: ACP orchestrator — coder, gates, reviewer, evidence
status: draft
phase: P1
created: 2026-07-27
ratified:
---

> ① **AWAITING RATIFICATION.** `gate verify` will refuse to declare readiness
> while this is `draft`. Read it and decide whether the *plan* is right — not
> whether code matches it — then set `status: ratified`.

## Goal

Automate, on one machine, the handoff currently performed by hand: one model implements a ratified spec, deterministic gates run over the result, a **different vendor's model** tries to disprove it against that spec, and the findings arrive as one evidence bundle.

This is the product ([ADR-013](../../docs/decisions.md#adr-013--the-product-is-cross-verification-not-an-agent)). Everything else — relay, mobile, watches — carries this to where the human is.

## Non-goals

- **Building an agent.** Roles are filled by Qwen Code, Kimi Code, OpenCode and other ACP agents ([ADR-014](../../docs/decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one)).
- **Handling anyone's credentials.** Each agent authenticates itself.
- **Calling a provider's subscription endpoint.** Ever ([ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)).
- Remote access — that is P2 and P3. This spec must be fully usable from one terminal.

## Invariants

- The orchestrator never sees an API key or subscription token. Asserted by test.
- Agents identify as themselves. No header, User-Agent or identifier is ever altered.
- A run that cannot be verified does not report success. `COULD NOT VERIFY` blocks the verdict exactly as a failure does.
- The reviewer is always a **different agent** from the coder. Same-agent review is refused with an explicit error, not silently allowed.
- Every run is bounded. No unbounded loop between two paid models.

## Requirements

## SPEC-acp-orchestrator-01 — ACP sessions are established and supervised

The orchestrator starts an ACP agent, negotiates capabilities, and maintains the session. An agent that dies mid-task produces a loud failure naming the agent and the step, never a silent retry that looks like success.

**Verified by:** `internal/acp` unit tests against a mock agent; a kill-the-agent integration test.

## SPEC-acp-orchestrator-02 — Permission requests are surfaced, held and resolved exactly once

`session/request_permission` from any agent enters a single queue with the requesting agent, its role, the proposed action and an expiry. Each request resolves exactly once; expiry **fails closed**.

The hold must survive long enough for a human decision — locally now, and a phone round trip in P3.

**Verified by:** approval unit tests including double-resolve and expiry; hold-time measured against real agents in [P0.17](../../docs/roadmap.md#p0--scaffolding-gates--de-risking).

## SPEC-acp-orchestrator-03 — Roles bind agents to responsibilities per project

`~/.rla/config.yaml` binds Coder, Reviewer and optionally Explorer to installed agents, per project. Any ACP agent can fill any role. Binding the same agent to Coder and Reviewer is rejected with an explanation.

**Verified by:** config round-trip tests; a rejection test for same-agent review.

## SPEC-acp-orchestrator-04 — Completed work is extracted for handoff

After the Coder reports completion, the orchestrator obtains the resulting diff and the list of touched files without reading the agent's internal state.

**Verified by:** integration test against a mock agent producing a known diff.

## SPEC-acp-orchestrator-05 — Gates run between implementation and review

The gate engine runs over the produced diff before the reviewer is invoked. Its output becomes part of the reviewer's input: deterministic findings are established fact, not something a model needs to rediscover.

**Verified by:** orchestration test asserting gate output reaches the reviewer prompt.

## SPEC-acp-orchestrator-06 — The reviewer is instructed to disprove, not confirm

The reviewer receives the ratified spec, the diff and the gate findings, with an explicit instruction to identify where the work fails to satisfy specific requirement ids — or to state that it found nothing. An empty or unparseable response is `COULD NOT VERIFY`, never a pass.

**Verified by:** prompt construction tests; a test asserting an unparseable review blocks the verdict.

## SPEC-acp-orchestrator-07 — Findings become one evidence bundle

Gate results and reviewer findings are merged into a single artifact stating, per requirement: covered, not covered, or not verifiable. It records what ran, what did not, and why.

**Verified by:** bundle schema tests; a golden-file test over a known run.

## SPEC-acp-orchestrator-08 — Runs are bounded and interruptible

An iteration budget caps coder↔reviewer cycles. Exhausting it stops the run and reports the state loudly. Cancellation takes effect at the next step boundary.

**Verified by:** budget-exhaustion test; cancellation test asserting no work starts after the boundary.

## SPEC-acp-orchestrator-09 — Draft specs block the verdict

A run against an unratified spec is refused. A run whose spec is edited mid-flight is invalidated. Checkpoint ① is not bypassable.

**Verified by:** refusal test; mid-run spec mutation test.

## SPEC-acp-orchestrator-10 — Agent setup is orchestrated, never redistributed

`rla setup` detects installed agents, installs missing ones through **their own official installers**, runs their login flows and verifies the result. No third-party binary is bundled or fetched by us.

**Verified by:** setup detection tests; a check that no agent binary appears in our release artifacts.

## SPEC-acp-orchestrator-11 — Catch rate is instrumented from the first run

Every run records what the gates caught, what the reviewer caught, and what the human caught afterwards. Without this number the central thesis stays an anecdote, and [X1](../../docs/vision-roadmap.md#x1--deeper-verification-tiers) has no gate to open.

**Verified by:** a metrics record per run; a report command summarising catch rate by source.

---

## Acceptance

Checkpoint ② for this phase, on one machine and one terminal:

- [ ] Two different agents on two different subscriptions run Coder and Reviewer against one ratified spec.
- [ ] A deliberately incomplete implementation is caught — by the gates, by the reviewer, or both — and the bundle says which.
- [ ] Approvals from both agents queue in one place; each resolves once; an expired one fails closed.
- [ ] Killing an agent mid-run produces `COULD NOT VERIFY`, not a pass.
- [ ] A draft spec refuses to run.
- [ ] `rla catch-rate` reports real numbers from real runs.
- [ ] No API key is visible to the orchestrator process at any point.

## Open questions

Why this is still `draft` — ratifying now would agree to a plan not yet made:

- **How long can an ACP permission request be held?** If agents time out locally in seconds, the phone approval flow in P3 needs a different mechanism — and that changes P1 and P3 together. [P0.17.1](../../docs/roadmap.md#p0--scaffolding-gates--de-risking) answers this and **this spec should not be ratified before it does.**
- **Does the reviewer need the conversation, or is the diff enough?** Diff-plus-spec is cheaper and keeps the agents independent. If the catch rate is materially worse without context, the handoff design changes.
- **What is the minimum acceptable catch rate?** Stated before measuring, or the answer will be fitted to whatever the data shows.
