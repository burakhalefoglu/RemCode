# RemLinkAgent — mobile client

> **Licence: Apache-2.0** — deliberately different from the rest of the repository.

Everything else in RemLinkAgent is AGPL-3.0-or-later. This directory is not, because GPL-family terms conflict with the Apple App Store's distribution rules, and an AGPL iOS build is a licence problem waiting to happen. Full reasoning: [ADR-002](../docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile).

## Rules for this directory

- Every file here starts with the **Apache-2.0** header ([`LICENSE_HEADER_APACHE`](../LICENSE_HEADER_APACHE)). `make license-add` applies it automatically based on path.
- **No AGPL code may be copied into `mobile/`.** Apache-2.0 is one-way compatible with AGPL — code moves core-ward, never mobile-ward.
- **No copyleft dependencies.** Every Flutter/Dart package must be Apache-2.0, MIT, BSD or similar.
- Wire types are **not** shared with the Go side as code. Both ends implement [`docs/protocol.md`](../docs/protocol.md) independently. Duplication here is intentional — it keeps the licence boundary clean.

## Status

Not bootstrapped yet — this is [P0.7](../docs/roadmap.md#p0--scaffolding-gates--de-risking):

```bash
flutter create --org com.remlinkagent --project-name rla_mobile \
               --platforms=ios,android mobile
```

Then restore this README and `LICENSE`, which `flutter create` will not overwrite but which must survive the bootstrap.

Architecture for the app (feature-first, Riverpod, go_router) is in [`architecture.md`](../docs/architecture.md#mobile-client).
