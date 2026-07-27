<!--
  RemLinkAgent pull request template.
  CI runs: ci-go, ci-flutter, license-check, cla.
-->

## What this does

<!-- What problem does it solve? Keep it short. -->

## Related issue

<!-- "Closes #123" / "Refs #456", or "None". -->

Closes #

## Type of change

- [ ] 🐛 Bug fix (`fix`)
- [ ] ✨ Feature (`feat`)
- [ ] 💥 Breaking change
- [ ] 📝 Documentation (`docs`)
- [ ] ♻️ Refactor — no behaviour change
- [ ] 🔧 Maintenance / dependency / CI (`chore`)
- [ ] 🧪 Tests

## Checklist

- [ ] Follows [`CONTRIBUTING.md`](../CONTRIBUTING.md).
- [ ] `make lint` and `make test` pass locally (Windows: `.\scripts\make.ps1 ci`).
- [ ] Tests added or updated for changed behaviour.
- [ ] `make license-check` passes — new files carry the header for **their directory**.
- [ ] Documentation updated where behaviour changed.

## Licence boundary — [ADR-002](../docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile)

- [ ] Files in `cmd/`, `internal/`, `deploy/`, `scripts/` carry the **AGPL-3.0-or-later** header.
- [ ] Files in `mobile/` carry the **Apache-2.0** header.
- [ ] No AGPL code was copied into `mobile/`. (Apache → core is fine; core → mobile is not.)
- [ ] New dependencies are compatible with the licence of the directory they land in — nothing copyleft in `mobile/`.

## Project invariants

<!-- These are the reasons the project exists. A PR that breaks one is rejected regardless of what it enables. -->

- [ ] ⚖️ **Vendor-neutral** — no model or provider is favoured in code or defaults.
- [ ] 🎭 **No impersonation** — no client identifier, User-Agent or header is altered to look like another tool.
- [ ] 🔑 **Zero-Touch AI** — AI traffic and API keys still never reach the relay, and the orchestrator still holds no provider credentials.
- [ ] 🔐 **E2E encryption** — the relay still handles ciphertext only ([ADR-004](../docs/decisions.md#adr-004--end-to-end-encryption-of-relay-payloads)).
- [ ] 🟢 **Fail-Loud** — no new path reports success on an error.
- [ ] 🌍 **i18n** — no hard-coded user-facing strings, in any language.
- [ ] 📓 If this changes an [ADR](../docs/decisions.md), a superseding entry was added rather than editing history.
- [ ] 🔀 If this changes the wire format, [`docs/protocol.md`](../docs/protocol.md) and the protocol version were updated.

## Screenshots / notes

<!-- Optional. -->

---

## 📜 Licence attestation

- [ ] I wrote this, or I have the right to submit it.
- [ ] I agree it ships under the licence of the directory it lands in — **AGPL-3.0-or-later** for the core, **Apache-2.0** for `mobile/`.
- [ ] It contains no code under a licence incompatible with that (e.g. GPLv2-only, proprietary).
- [ ] I have signed the [CLA](../CLA.md), or will when the bot asks. ([Why](../docs/decisions.md#adr-003--cla-required))
- [ ] It contains no credentials, API keys or personal data.
