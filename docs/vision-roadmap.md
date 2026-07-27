# 🔭 RemLinkAgent — Vision Roadmap

> **Version:** 2.0 · **Updated:** 2026-07-27 · **Status:** ON HOLD
> **Prerequisite:** [`roadmap.md`](roadmap.md) P0–P4 **and** M1 shipped.

Everything here increases the product's value without being required for it to survive. Each phase begins by answering *"should we do this now?"* with metrics and community input — not with the order they happen to appear in.

**Ordering logic**

- **Managed cloud is not here.** It funds the project, so it sits in the main roadmap as [M1](roadmap.md#m1--managed-cloud-subscriptions--team).
- **X1 first** — role delegation and multi-project is what users ask for once they are comfortable with hot-swap. Fastest return.
- **X6 last** — Loop Engineering is research-grade. Do not start it before "how often is the AI wrong?" is an actual measured number.

---

## Why each was deferred

| Feature | Reason | Revisit when |
| :--- | :--- | :--- |
| **Role delegation + multi-project** | Single-model, single-project has to be validated first — though this is the cheapest to add | Active users pass threshold → **X1** |
| **Voice control** | Extra platform surface (audio routing, interruptions); value unproven | Hands-free demand is actually measured |
| **Apple Watch / Wear OS** | Separate native projects; a small surface for the cost | A solid user base exists |
| **Advanced AirPods (CallKit)** | High App Store rejection risk | Later |
| **Terminal mirroring (PTY)** | Highest engineering risk in the project, weakest differentiator ([ADR-005](decisions.md#adr-005--mvp-is-an-agent-not-a-terminal-mirror)) | Users report the agent is not enough |
| **Loop Engineering tiers 1–4** | Research-grade. Riskiest thing in the plan | "The AI wrote the wrong thing" becomes a measured, frequent complaint → **X6** |

---

## X1 — Multi-model orchestration

**Goal:** role-based delegation and multi-project. The natural extension of MVP hot-swap.
**Estimate:** 4 weeks

Once someone is comfortable switching models by hand, the next thing they want is for it to happen automatically — a cheap model writing, a strict one reviewing — and for it to work across more than one repository.

- [ ] **X1.1** AI Roles screen: Coder / Reviewer / Architect → model assignment.
- [ ] **X1.2** Role-based routing: plan → Architect, implement → Coder, review diff → Reviewer.
- [ ] **X1.3** Multi-project binding + project switcher.
- [ ] **X1.4** Per-project model/role configuration (`model_configs` table).
- [ ] **X1.5** Cost and usage reporting per provider, summarised on mobile.
- [ ] **X1.6** Deeper observability, mobile and server.
- [ ] **X1.7** Beta polish: onboarding tour, empty states, haptics.

**Done when** roles are assignable, "add Stripe" flows Architect → Coder → Reviewer, and switching projects is seamless.

---

## X2 — Voice control

**Goal:** hands-free operation over AirPods or any BLE headset — chat, model switching, approvals.
**Estimate:** 2 weeks

Flutter plugins only; no native code at this stage.

- [ ] **X2.1** `audio_session`: `playAndRecord`, `voiceChat` mode, Bluetooth A2DP + HFP.
- [ ] **X2.2** Headset detection and automatic routing (route-change listener).
- [ ] **X2.3** Interruption handling — an incoming call pauses and resumes cleanly.
- [ ] **X2.4** `speech_to_text`: on-device with cloud fallback, TR + EN, streaming partial results.
- [ ] **X2.5** Command grammar: `chat <message>`, `switch model <name>`, `status`, `approve`, `reject`.
- [ ] **X2.6** Intent parser (regex + fuzzy); free-form speech becomes a chat message.
- [ ] **X2.7** Listening UI with live transcript.
- [ ] **X2.8** ➕ TTS read-back (`flutter_tts`), opt-in.
- [ ] **X2.9** Tests: intent parser units, audio-route scenarios, real hardware.

> ⚠️ **Approving a destructive command by voice needs care.** Voice recognition is probabilistic; command approval is not. Either require a second confirmation or exclude approvals from voice entirely — decide before building.

---

## X3 — Apple Watch

**Goal:** monitoring and approvals from the wrist.
**Estimate:** 4 weeks

- [ ] **X3.1** `watchos/` Xcode project (SwiftUI, watchOS 10+), App Group sharing.
- [ ] **X3.2** WatchConnectivity bridge: `sendMessage`, `transferUserInfo`, `updateApplicationContext`.
- [ ] **X3.3** Watch UI: status, approvals, project picker.
- [ ] **X3.4** Complication showing agent state.
- [ ] **X3.5** Push + haptic patterns; actionable notifications.
- [ ] **X3.6** Plan-based capability: Free view-only, Pro/Team full control.
- [ ] **X3.7** TestFlight and real-device testing.

> The watch app must decrypt locally too — the same E2E constraint as the phone ([ADR-004](decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)). Key material has to reach the watch securely, which is the hard part of this phase.

---

## X4 — Wear OS

**Goal:** X3 parity on Android.
**Estimate:** 4 weeks

- [ ] **X4.1** `wearos/` Gradle module (Compose for Wear OS, API 33+).
- [ ] **X4.2** Wearable Data Layer: `MessageClient`, `DataClient`, `ChannelClient`.
- [ ] **X4.3** Wear UI: status, approvals, project picker.
- [ ] **X4.4** Tile + complication provider.
- [ ] **X4.5** FCM bridge, haptics, actionable notifications.
- [ ] **X4.6** Plan-based capability.
- [ ] **X4.7** Play Console (Wear OS) distribution.

---

## X5 — Interactive terminal mirroring (PTY)

**Goal:** a real interactive terminal on the phone — the capability deliberately cut from the MVP.
**Estimate:** 4–6 weeks

Deferred by [ADR-005](decisions.md#adr-005--mvp-is-an-agent-not-a-terminal-mirror): highest engineering risk, weakest differentiator, and well served by existing tools.

- [ ] **X5.1** Cross-platform PTY allocation ([`creack/pty`](https://github.com/creack/pty); ConPTY on Windows).
- [ ] **X5.2** Streaming terminal I/O over the existing encrypted transport.
- [ ] **X5.3** Mobile terminal emulator widget: ANSI escapes, colour, cursor.
- [ ] **X5.4** Resize propagation, signal forwarding (Ctrl-C, Ctrl-D).
- [ ] **X5.5** On-screen key row: Tab, Esc, Ctrl, arrows.
- [ ] **X5.6** Session persistence across reconnects.
- [ ] **X5.7** ➕ Native `tmux` integration.

> Revisit only if users say the agent alone is insufficient. If most requests are "I want to run one quick command", that is already the exec tool — not this.

---

## X6 — Loop Engineering, tiers 1–4

**Goal:** the optional, tiered autonomous QA pipeline. The project's original ambition.
**Estimate:** 10–12 weeks — the largest and riskiest phase.
**Reference:** [`loop-engineering.md`](loop-engineering.md)

> ⚠️ **Deliberately last.** Everything here rests on LLM judgement and none of it is deterministic. The distance between a working demo and something trustworthy is widest in this phase. Do not begin before AI error rate is a measured number from real usage.

Tier 0 already shipped in the MVP ([ADR-008](decisions.md#adr-008--loop-engineering-tier-0-ships-in-the-mvp)); it is the deterministic part, and it does not belong to this risk category.

- [ ] **X6.1** `rla loop enable|disable`; `.rla/` bootstrap (`PRINCIPLES.md`, `SECURITY-BASELINE.md`, `specs/`).
- [ ] **X6.2** Spec generation: Architect writes `.rla/specs/<feature>.md` with `SPEC-{feature}-NN` ids, `status: draft`.
- [ ] **X6.3** `/new-feature <name>`.
- [ ] **X6.4** Spec ratification: pipeline halts until mobile approval; `status: ratified`.
- [ ] **X6.5** PreToolUse/PostToolUse hook framework (Go native).
- [ ] **X6.6** **Tier 1 (inner loop):** changed-file tests, architectural conformance, fast fidelity check; 20–30 iterations.
- [ ] **X6.7** Tier router — model-to-tier binding.
- [ ] **X6.8** **Selective regression cache:** file hash + SPEC-ID set + config signature; skip when all three are unchanged.
- [ ] **X6.9** Git-commit exemption — cache keys never bind to a commit SHA.
- [ ] **X6.10** Canary tests + fail-loud + `COULD NOT VERIFY` state on mobile.
- [ ] **X6.11** **Fake-green hunting:** assertion-free tests, inflated coverage, misconfigured SAST reporting zero findings.
- [ ] **X6.12** **Tier 2 (convergence):** integration tests, coverage, **bidirectional spec diff** (forward: missing → FAIL; backward: unspecified behaviour → DEVIATION).
- [ ] **X6.13** **Tier 3 (heavy):** mutation testing, fuzzing, SAST, **black-box explorer** with no source access; 2–3 iterations.
- [ ] **X6.14** Mobile Loop dashboard + spec approval screen.
- [ ] **X6.15** Adoption into an existing codebase: scan → draft specs + `⚠️ SUSPECTED DEVIATION` markers.
- [ ] **X6.16** Security additions: log-leak prevention, CVE scanning, project-specific invariants.

**Done when** a spec is generated, ratified on mobile, tiers run in order, a deviation is caught and corrected, and a crashed gate halts loudly instead of reporting green.

---

## Risks

| Risk | Impact | Likelihood | Mitigation |
| :--- | :--- | :--- | :--- |
| **Loop Engine never becomes trustworthy** | High | High | Staged rollout; Tier 0 already proven in MVP; X6 last |
| Spec-diff non-determinism | High | High | Forward direction (missing) first; backward (deviation) experimental |
| Voice approving a destructive command | **Critical** | Medium | Second confirmation, or exclude approvals from voice |
| Watch key distribution | High | Medium | Design against the E2E model before building |
| CallKit store rejection | Medium | High | Guideline review first; keep it late |
| Watch development cost | Medium | Medium | Only after a strong user base |
| Scope creep back into MVP/M1 | High | Medium | Returning to a shipped phase requires an ADR |

---

## Prioritisation criteria

Before starting any X phase, collect from MVP + M1:

1. **Active users** — past the threshold?
2. **Request categories** — what is actually asked for? (issues, discussions, email)
3. **AI error rate** — "it wrote the wrong thing" complaints. **This is the gate for X6.**
4. **Retention** — why do people leave? Missing capability, or complexity?
5. **Contribution areas** — where are PRs arriving?

> **Rule:** at least two metrics must give a strong signal. **X6 has an additional gate:** AI error rate must be instrumented and measured, not estimated.

---

## Glossary

**Loop Engineering** — autonomous, tiered QA from spec to verification. **Tier 0–4** — verification stages (lint / inner loop / convergence / heavy / periodic). **Spec artifact** — `.rla/specs/<feature>.md` with `SPEC-NN` ids. **Fake-green** — a passing gate that proves nothing. **WatchConnectivity** — iOS phone↔watch framework. **Wearable Data Layer** — Android equivalent. **CallKit** — iOS VoIP call UI framework.

---

*Post-MVP, post-M1. The MVP is in [`roadmap.md`](roadmap.md).*
