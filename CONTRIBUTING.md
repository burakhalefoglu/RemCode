# 🤝 Contributing to RemLinkAgent

Thanks for considering it. The project is at **P0** — no feature code exists yet — which means contributions can still shape the architecture rather than work around it.

---

## 📜 Licensing (please read before your first PR)

RemLinkAgent uses **two licences**, deliberately:

| Path | Licence | Header template |
| :--- | :--- | :--- |
| `cmd/`, `internal/`, `deploy/`, tooling | **AGPL-3.0-or-later** | [`LICENSE_HEADER`](LICENSE_HEADER) |
| `mobile/` (Flutter client) | **Apache-2.0** | [`LICENSE_HEADER_APACHE`](LICENSE_HEADER_APACHE) |

The core is AGPL so that a modified hosted relay has to publish its changes. The mobile client is Apache-2.0 because GPL-family terms conflict with App Store distribution — an AGPL iOS build is a licence problem waiting to happen. Full reasoning: [ADR-002](docs/decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile).

**Practical rules:**

- Put a file in the directory whose licence you intend. `make license-add` applies the right header automatically.
- Apache-2.0 code from `mobile/` may be reused in the core. **The reverse is not permitted** — AGPL code cannot move into `mobile/`.
- Wire types are **not** shared as code. Both sides implement [`docs/protocol.md`](docs/protocol.md) independently. This is what keeps the boundary clean.

### Contributor License Agreement

A signed [**CLA**](CLA.md) is required. [CLA Assistant](https://cla-assistant.io/) posts a link on your first PR; signing is a single click and covers every future PR.

**This reverses an earlier position, and it is worth saying why.** An earlier draft of this guide promised no CLA. It also planned to keep a future commercial dual-licence available. Those are mutually exclusive: relicensing needs someone who holds the rights to relicense, and that permission cannot be collected after the fact. Rather than quietly keep an impossible promise, the option is being preserved and the change stated openly — see [ADR-003](docs/decisions.md#adr-003--cla-required).

**You keep your copyright.** The CLA grants a licence, it does not transfer ownership. Your work stays available under AGPL / Apache-2.0 to everyone, permanently, regardless of what happens to the project.

---

## 🚀 Ways to help

| Area | How |
| :--- | :--- |
| 🏗️ **P0–P1 implementation** | The critical path. See [`roadmap.md`](docs/roadmap.md) |
| 🔬 **Provider contract tests** | Verify tool-calling and streaming for a provider — [P0.9](docs/roadmap.md#p0--scaffolding--de-risking) |
| 🔐 **Threat model review** | Poke holes in [`threat-model.md`](docs/threat-model.md); adversarial reading is genuinely wanted |
| 🐛 **Bug reports** | [Open an issue](../../issues) |
| 💡 **Feature ideas** | [Open an issue](../../issues) first — check [`vision-roadmap.md`](docs/vision-roadmap.md), it may already be planned |
| 📝 **Docs** | Anything unclear in [`docs/`](docs/) is a bug |
| 🌍 **Translations** | UI is TR + EN; more languages welcome once P3 lands |

---

## 🛠️ Development setup

### Prerequisites

| Tool | Version | Needed for |
| :--- | :--- | :--- |
| **Go** | ≥ 1.24 | CLI + server |
| **Docker** + Compose | any recent | NATS/JetStream |
| **Flutter** | ≥ 3.22 | Mobile app (P3 onward) |
| **Git** | any | Signed commits appreciated |

### Get running

```bash
git clone https://github.com/<your-username>/RemLinkAgent.git
cd RemLinkAgent
git remote add upstream https://github.com/burakhalefoglu/RemLinkAgent.git

make help          # every available target
make cli server    # → ./bin/rla, ./bin/rla-server
make test lint     # green on the skeleton
make docker-up     # NATS + server
```

The Flutter project does not exist yet — `make mobile` prints the exact `flutter create` command to bootstrap it ([P0.7](docs/roadmap.md#p0--scaffolding--de-risking)).

---

## 📋 Pull request workflow

**1. Discuss anything non-trivial first.** Open an issue before building a feature or restructuring something. It is cheaper than a rejected PR.

**2. Branch.**

```bash
git checkout -b feat/short-description
```

`feat/` · `fix/` · `docs/` · `chore/` · `refactor/` · `test/`

**3. Build and check.**

```bash
make lint            # golangci-lint + flutter analyze
make test            # all tests
make license-check   # header enforcement — CI blocks on this
```

**4. Commit** using [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat(cli): add context normalization for model hot-swap
fix(server): handle heartbeat loss during WSS reconnect
docs(protocol): document event envelope versioning
chore(deps): bump go-openai to v1.40.0
```

Types: `feat` `fix` `docs` `chore` `refactor` `test` `perf` `ci`

**5. Headers.** `make license-add` adds the correct one for each path. CI rejects PRs with missing headers.

**6. Open the PR.** Fill in the template, link the issue (`Closes #123`), get CI green.

---

## ✅ Code standards

**Go** — `gofmt` + `goimports` + `golangci-lint`. Return errors, do not panic. Wrap with context: `fmt.Errorf("pairing device: %w", err)`.

**Dart** — `dart format` + `flutter analyze`, following Effective Dart.

**Tests** — required for new behaviour. Coverage of changed code must not drop.

**Comments** — explain *why*. The *what* is already in the code.

**Dependencies** — must be compatible with the licence of the directory they land in. MIT / BSD / Apache-2.0 / MPL are fine everywhere. GPLv2-only is not compatible with AGPL-3.0. Nothing copyleft in `mobile/`.

**User-facing strings** — always i18n keys, never hard-coded, in either language. Applies to push notification bodies and CLI output alike.

### Invariants that will get a PR rejected

These are not style preferences — they are the reasons the project exists:

| Invariant | Meaning |
| :--- | :--- |
| 🔑 **Zero-Touch AI** | AI traffic and API keys never reach the relay. No exceptions, no debug flags, no "temporary" telemetry. |
| 🔐 **E2E encryption** | The relay handles ciphertext. Any change that would let it read plaintext is rejected regardless of what it enables. |
| 🟢 **Fail-Loud** | Never report success on an error path. Halt visibly. A silent green is worse than a red. |
| 🪶 **Lightness** | CLI target < 30 MB RSS. Justify new dependencies. |
| 📓 **Decisions are logged** | Changing an [ADR](docs/decisions.md) means adding a superseding entry, not editing history. |

---

## 🐛 Bug reports

Include: a one-line summary, numbered reproduction steps, expected vs. actual, versions (CLI / server / OS / model / provider), and logs or screenshots.

**Never paste an API key, pairing token or device token into an issue.** Redact before posting. If you already have, rotate the key first, then report.

---

## 🔐 Security issues

Do **not** open a public issue. Follow the disclosure process in [`SECURITY.md`](SECURITY.md).

---

## 💰 Money, honestly

- **Self-hosted:** free, full parity, permanently. No feature is reserved for the paid tier.
- **Managed cloud (M1):** $5/mo Pro. You are paying for uptime, push infrastructure and upgrades — not for features.

Managed-cloud revenue goes to the maintainer, who carries the hosting cost, the on-call burden and the legal exposure. Contributing does not create a claim on it. This is stated plainly because finding it out later feels worse than reading it now.

What contributing *does* guarantee: your work stays AGPL / Apache-2.0 forever. Nobody — including the maintainer — can make it proprietary or take it away from you or anyone else.

GitHub Sponsors / OpenCollective for redistributing to contributors may follow if revenue justifies it. Not active, and not a promise.

---

## 🌟 Code of conduct

Treat people with respect. Harassment, personal attacks and discrimination are not tolerated. Assume good faith, critique work rather than people.

Report problems by email to the maintainer if it needs to stay private.

---

*Thanks — early contributions are worth disproportionately more.* ❤️
