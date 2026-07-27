---
id: provider-contract
title: Provider capability matrix, auth paths and contract tests
status: archived
phase: P0.17
created: 2026-07-27
ratified: 2026-07-27
archived: 2026-07-27
---

> # 📦 ARCHIVED — its questions were answered, and the answers changed the product
>
> This spec was ratified to measure whether one client wrapper could serve three
> providers, and how subscription access worked. **Both questions were answered
> by desk research before any code was written**, and the answers made most of
> the spec obsolete:
>
> | Requirement | Outcome |
> | :--- | :--- |
> | 06 — subscription access model | **Answered.** All three are Type A (subscription issues an API key) — but all three place third-party clients outside the permitted scope. → [ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us) |
> | 07 — terms-of-service position | **Answered.** Z.AI: closed list. Qwen: category rule excluding application backends, with an *"Only available for Coding Agents"* server-side rejection. Kimi: interactive use only, identity spoofing prohibited by name. → [ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us) |
> | 01–05, 08–10 — tool schema, streaming, finish reasons, cassettes, containment verdict | **Obsolete on the primary path.** [ADR-014](../../docs/decisions.md#adr-014--orchestrate-acp-agents-do-not-build-one) removed the direct provider integration; we drive existing agents instead, so their vendors own these concerns. |
>
> The direct pay-as-you-go path survives as a fallback, and these checks apply
> to it — at a much lower priority. The replacement work is
> [`acp-orchestrator`](acp-orchestrator.md).
>
> Kept as a record. A spike that kills its own plan did its job.

---

## Goal

Answer, with real API calls rather than documentation, two questions that the
whole P1 design rests on:

1. **Can one client wrapper serve Z.AI, Qwen and Kimi?**
2. **How does each provider let a user authenticate with a subscription rather
   than metered API billing?**

**Why this gates P1.** The design assumes one `go-openai` wrapper plus thin
per-provider adapters. That holds for chat completion and is noticeably weaker
for tool calling and streamed tool deltas. And [ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)
commits the MVP to subscription authentication, which is not OpenAI-compatible
at all — it is a bespoke flow per provider, or no flow, depending on the
provider. Learning either of these in week 6 of P1 costs far more than learning
them now.

## Non-goals

- Building the adapters or the auth flows. This spike produces **findings and
  contract tests**, not the abstraction.
- Benchmarking answer quality or price. Capability and access only.
- Providers beyond the three first-class ones ([ADR-006](../../docs/decisions.md#adr-006--provider-neutral-core-first-class-zaiqwenkimi)).

## Decisions taken before measurement

Recorded here because a threshold chosen after seeing the data is not a
threshold.

**Compatibility verdict — adapter containment.** The single-wrapper assumption
survives if and only if **every difference can be absorbed inside a
per-provider adapter without changing the agent loop's interface.** One
difference that forces a change in the loop means the adapter layer is
redesigned before P1 starts. Not a checklist of capabilities: the question
being asked is "does one wrapper work", so that is what gets measured.

**Fixture policy — cassettes plus a scheduled live run.** Contract tests run
from committed, sanitised response recordings: deterministic, fast, and
runnable by a contributor with no keys. A scheduled live job re-runs them
against the real APIs and fails on divergence. Cassettes alone would stay green
after a provider changed its API, which is a fake-green by construction; live
alone would make every PR depend on third-party uptime.

**Missing keys.** Keys are supplied by the maintainer. If one becomes
unobtainable, the affected provider's rows are marked `UNVERIFIED` — never
inferred from documentation, never left blank.

## Invariants

- Every finding is reproducible by a test someone else can run. A sentence in a
  document saying "Qwen does X" is not a finding.
- A provider that fails a contract test gets an **adapter**, never a special
  case scattered through the agent loop.
- No API key, subscription token or account identifier appears in a fixture, a
  log, or a recorded response.
- Cassettes are sanitised at record time, not at review time.

## Requirements

## SPEC-provider-contract-01 — Tool schema compatibility is measured per provider

For each provider, submit a tool definition exercising nested object
parameters, `required` arrays, enums, and `additionalProperties: false`.
Record which constructs are accepted, rejected, or silently ignored.

**Verified by:** `internal/models/contract_test.go`, one subtest per provider.

## SPEC-provider-contract-02 — Parallel tool-call behaviour is known

Determine whether each provider can return multiple tool calls in one
assistant turn, and how it behaves when asked to. The agent loop must handle
both shapes, so the answer is recorded rather than assumed.

**Verified by:** contract test asserting the recorded behaviour per provider.

## SPEC-provider-contract-03 — Streaming tool deltas are characterised

Capture the SSE chunk sequence for a streamed tool call: how argument
fragments are split, whether the tool name arrives in the first chunk, and what
terminates the call. This determines whether partial rendering is possible or
whether tool calls must be buffered whole.

**Verified by:** contract test over a recorded chunk sequence per provider.

## SPEC-provider-contract-04 — Finish reasons are mapped

Enumerate the finish-reason values each provider emits and map them to one
internal enum. An unmapped value must fail loudly, never be treated as a normal
stop — a silently mishandled stop reason ends the agent loop early and looks
like success.

**Verified by:** contract test; exhaustive switch with an explicit default.

## SPEC-provider-contract-05 — Token accounting and rate-limit signals are located

Record where each provider reports token usage — including whether it appears
on streamed responses — and which rate-limit headers it returns, so backoff is
informed rather than blind.

**Verified by:** contract test asserting the fields exist where documented.

## SPEC-provider-contract-06 — Subscription authentication is characterised per provider

For each provider, establish which access model applies and record it:

- **Type A** — the subscription issues an API key usable against the normal
  endpoints. Nothing new is needed; this is documentation, not engineering.
- **Type B** — the subscription is only reachable through a login/OAuth flow.
  Record the flow type (device code, browser redirect), the token lifetime,
  the refresh mechanism, and whether the endpoints differ from the API-key ones.
- **Type C** — no subscription access exists for third-party clients.

The classification decides how much of [ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)
is real work versus a paragraph in the install guide.

**Verified by:** the access-model table in [`docs/protocol.md`](../../docs/protocol.md#acp-capability-matrix);
a live auth test per Type B provider.

## SPEC-provider-contract-07 — Terms-of-service position is recorded per provider

For each provider, record what its terms say about using subscription
credentials in a third-party client: **supported**, **unclear**, or
**prohibited**, with a dated citation.

This is the highest-consequence finding in the spike. Getting it wrong does not
produce a legal problem first — it produces **banned user accounts**, and the
user blames RemLinkAgent. A provider whose position is `prohibited` does not
get a subscription auth path, regardless of technical feasibility.

**Verified by:** dated citations in `docs/protocol.md`; reviewed at every
release, since terms change without notice.

## SPEC-provider-contract-08 — Contract tests run from sanitised cassettes

Tests execute against committed, sanitised recordings by default: no keys
required, deterministic, offline. Recording is an explicit opt-in mode.
Sanitisation strips credentials, account identifiers and response content at
record time.

**Verified by:** the suite passes in a clean checkout with no environment
variables set.

## SPEC-provider-contract-09 — A scheduled live run detects cassette drift

A scheduled CI job re-runs the contract tests against the real APIs and fails
when behaviour diverges from the cassettes. Absent keys make the job report
`COULD NOT VERIFY`, never a pass.

**Verified by:** the scheduled workflow; a deliberately stale cassette fails it.

## SPEC-provider-contract-10 — The adapter-containment verdict is stated

The spike concludes with an explicit verdict: does every measured difference
fit inside a per-provider adapter without changing the agent loop's interface?
A yes/no, with the differences enumerated. If no, the P1 adapter design is
revised before P1 begins.

**Verified by:** the verdict paragraph in `docs/protocol.md`; referenced by the
P1 design.

---

## Acceptance

Checkpoint ② for a research spike is a reading, not a click:

- [ ] Capability and access-model tables in `protocol.md` filled for all three
      providers, with `UNVERIFIED` where a key was unobtainable.
- [ ] ToS position recorded per provider, dated.
- [ ] Contract tests green in a clean checkout with no keys.
- [ ] Scheduled live job exists and has run at least once.
- [ ] The containment verdict is written, and if negative, P1's adapter design
      has been revised before P1 starts.

## Open questions

None. Ratified 2026-07-27 — see *Decisions taken before measurement* above, and
[ADR-012](../../docs/decisions.md#adr-012--subscription-access-goes-through-listed-agents-never-through-us)
for the scope decision that produced requirements 06 and 07.
