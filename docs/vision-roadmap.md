# 🔭 RemLinkAgent — Vision Roadmap

> **Version:** 3.0 · **Updated:** 2026-07-27 · **Status:** ON HOLD
> **Prerequisite:** [`roadmap.md`](roadmap.md) P0–P4 **and** M1 shipped.

Everything here increases the product's value without being required for it to work. Each phase begins by answering *"should we do this now?"* with measurements, not with the order it happens to appear in.

## What changed in v3.0

[ADR-013](decisions.md#adr-013--the-product-is-cross-verification-not-an-agent) moved the centre of the product. Two former vision phases are no longer here:

| Was | Now |
| :--- | :--- |
| **X1** Multi-model orchestration | **Core** — it is the product ([P1](roadmap.md#p1--orchestrator)) |
| **X6** Loop Engineering tiers 1–4 | Deterministic gates **ship today**; cross-model review is [P1](roadmap.md#p1--orchestrator); only the heavy tiers remain, as X1 below |

Remaining phases renumbered accordingly.

---

## Why each was deferred

| Feature | Reason | Revisit when |
| :--- | :--- | :--- |
| **Deeper verification tiers** | Mutation and fuzzing are slow and expensive; black-box exploration is research-grade. Cheaper gates must prove their catch rate first | Cross-model review's catch rate is measured and the ceiling is visible |
| **Voice control** | Extra platform surface (audio routing, interruptions); value unproven | Hands-free demand is actually measured |
| **Smartwatches** | Separate native projects; small surface for the cost | A solid user base exists |
| **Terminal mirroring** | Highest engineering risk, weakest differentiator — `ssh` and `tmux` already do it well | Users report the orchestrator alone is insufficient |

---

## X1 — Deeper verification tiers

**Goal:** push cross-verification past what a reviewing model can find by reading.
**Estimate:** 8–10 weeks — the largest and riskiest phase.
**Reference:** [`loop-engineering.md`](loop-engineering.md)

> ⚠️ **Do not start before the catch rate of P1's cross-model review is a measured number.** If a reviewing model already finds most of what matters, mutation testing buys little for a great deal of time. If it plateaus early, this is where the ceiling lifts.

- [ ] **X1.1** **Mutation testing:** deliberately break the code, confirm the tests notice. The strongest available answer to "are these tests real?"
- [ ] **X1.2** **Fuzzing:** generated inputs against public interfaces; failures become permanent regression tests.
- [ ] **X1.3** **Black-box explorer:** an agent whose isolation is enforced at the tool-permission level — **it cannot read implementation source**. It tests what the interface promises rather than what the code does, and the gap between those is where bugs live.
- [ ] **X1.4** **Property-based testing:** invariants over generated inputs rather than fixed examples.
- [ ] **X1.5** **SAST integration** with fake-green detection — a scanner configured to find nothing reports nothing.
- [ ] **X1.6** **Tier 4 periodic:** self-audit, cross-feature exploration, adversarial simulation.
- [ ] **X1.7** Cost controls: heavy tiers are opt-in per project with a visible budget.

**Done when** a deliberately introduced defect that survives Tier 2 is caught by Tier 3, and the marginal catch rate over Tier 2 is measured rather than asserted.

---

## X2 — Voice control

**Goal:** hands-free operation over AirPods or any BLE headset — check status, review findings, approve.
**Estimate:** 2 weeks

Flutter plugins only; no native code at this stage.

- [ ] **X2.1** `audio_session`: `playAndRecord`, `voiceChat`, Bluetooth A2DP + HFP.
- [ ] **X2.2** Headset detection and automatic routing.
- [ ] **X2.3** Interruption handling — an incoming call pauses and resumes cleanly.
- [ ] **X2.4** `speech_to_text`: on-device with cloud fallback, TR + EN, streaming partial results.
- [ ] **X2.5** Command grammar: `status`, `read findings`, `approve`, `reject`, `cancel run`.
- [ ] **X2.6** Intent parser; free-form speech becomes an instruction to the Coder.
- [ ] **X2.7** ➕ TTS read-back of reviewer findings — arguably the most useful part: hearing what the reviewer found while walking.
- [ ] **X2.8** Tests: intent parser, audio-route scenarios, real hardware.

> ⚠️ **Approving a destructive command by voice needs care.** Recognition is probabilistic; approval is not. Either require a second confirmation or exclude approvals from voice entirely — decide before building.

---

## X3 — Smartwatches

**Goal:** run status and approvals from the wrist.
**Estimate:** 4 weeks (Apple Watch) + 4 weeks (Wear OS)

The natural surface for this product: checkpoint ② is a glance and a decision, not a typing session.

### Apple Watch

- [ ] **X3.1** `watchos/` Xcode project (SwiftUI, watchOS 10+), App Group sharing.
- [ ] **X3.2** WatchConnectivity bridge.
- [ ] **X3.3** UI: run status per project, approval prompts, findings summary.
- [ ] **X3.4** Complication showing per-project verification state.
- [ ] **X3.5** Push + haptic patterns; actionable notifications.
- [ ] **X3.6** TestFlight and real-device testing.

### Wear OS

- [ ] **X3.7** `wearos/` Gradle module (Compose for Wear OS, API 33+).
- [ ] **X3.8** Wearable Data Layer bridge.
- [ ] **X3.9** UI parity with watchOS.
- [ ] **X3.10** Tile + complication provider.
- [ ] **X3.11** Play Console (Wear OS) distribution.

> The watch must decrypt locally too ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)). Getting key material to it securely is the hard part of this phase, not the UI.

---

## X4 — Interactive terminal mirroring

**Goal:** a real interactive terminal on the phone.
**Estimate:** 4–6 weeks

Deferred by [ADR-005](decisions.md#adr-005--mvp-is-an-agent-not-a-terminal-mirror) and unaffected by the thesis change: still the highest engineering risk and the weakest differentiator.

- [ ] **X4.1** Cross-platform PTY ([`creack/pty`](https://github.com/creack/pty); ConPTY on Windows).
- [ ] **X4.2** Streaming terminal I/O over the existing encrypted transport.
- [ ] **X4.3** Mobile terminal emulator: ANSI escapes, colour, cursor.
- [ ] **X4.4** Resize propagation, signal forwarding.
- [ ] **X4.5** On-screen key row.
- [ ] **X4.6** Session persistence across reconnects.

> Revisit only if users say the orchestrator is insufficient. If most requests are "I want to run one quick command", that is the approval queue — not this.

---

## Risks

| Risk | Impact | Likelihood | Mitigation |
| :--- | :--- | :--- | :--- |
| **Heavy tiers buy little over cross-model review** | High | Medium | Measure P1's catch rate first; X1 is gated on that number |
| Black-box explorer isolation leaks | High | Medium | Enforce at the tool-permission level, not by prompting |
| Voice approving a destructive command | **Critical** | Medium | Second confirmation, or exclude approvals from voice |
| Watch key distribution | High | Medium | Design against the E2E model before building |
| Heavy tiers make runs unaffordable | Medium | High | Opt-in per project, visible budget, hard stop |
| Scope creep back into the MVP | High | Medium | Returning to a shipped phase requires an ADR |

---

## Prioritisation criteria

Before starting any X phase, collect from P0–P4 and M1:

1. **Catch rate** — what proportion of real defects does cross-model review find? **This is the gate for X1.**
2. **Active users** — past the threshold?
3. **Request categories** — what is actually asked for?
4. **Retention** — why do people leave? Missing capability, or complexity?
5. **Contribution areas** — where are PRs arriving?

> **Rule:** at least two metrics must give a strong signal. **X1 has an additional gate:** catch rate must be instrumented and measured, not estimated.

---

## Glossary

**Mutation testing** — deliberately breaking code to confirm tests notice. **Black-box explorer** — a verifier that cannot read the implementation. **Fake-green** — a gate that passes without proving anything. **Catch rate** — proportion of real defects a verification stage finds. **WatchConnectivity** — iOS phone↔watch framework. **Wearable Data Layer** — Android equivalent.

---

*Post-MVP, post-M1. The MVP is in [`roadmap.md`](roadmap.md).*
