# 🤝 Contributing to RemLinkAgent

> First off, thanks for taking the time to contribute! 🎉
> This document explains how to contribute to the project.

RemLinkAgent is an open-source mobile AI agent released under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. We welcome community contributions of all kinds — following this guide keeps the process smooth for everyone.

---

## 📜 License (Important — Please Read)

RemLinkAgent is licensed under **AGPL v3**. By contributing to the project, you agree to the following:

1. **Your contributions are licensed under AGPL v3.** The code, documentation, and other content you submit become part of the project under the terms of the [AGPL-3.0 license](./LICENSE). This is AGPL's "inbound = outbound" rule: contributing to an AGPL project means you automatically accept that your contribution will be distributed under the same AGPL license.

2. **No separate CLA (Contributor License Agreement) is required.** This project has **no** dual-licensing or commercial relicensing plans. Because of this, you are not asked to assign the copyright of your contributions to the project. AGPL v3 keeps everything under a single, transparent license.

3. **You retain ownership; the license is AGPL.** The copyright to the code you write remains yours; however, you grant everyone the right to use it under the terms of AGPL v3 (including modification, distribution, and network use).

> **In short:** By submitting your code, you accept that it will be shared with everyone under AGPL v3. No hidden obligations beyond that.

---

## 🚀 Ways to Contribute

You don't have to write code — there are many ways to help:

| Area | How? |
| :--- | :--- |
| 🐛 **Bug reports** | Open a new [Issue](../../issues) and describe the bug clearly |
| 💡 **Feature requests** | [Open an Issue to discuss](../../issues) first, then code |
| 🔧 **Code contributions** | Follow the development workflow below |
| 📝 **Documentation** | Improvements to README, docs/, code comments |
| 🌍 **Translations** | UI translations (TR + EN → other languages) |
| 📣 **Spread the word** | Star the project ⭐, share it, write a blog post |

---

## 🛠️ Development Environment Setup

### Prerequisites

- **Go** ≥ 1.22 (for CLI + Server)
- **Flutter** ≥ 3.22 (for the Mobile app)
- **Docker** + **Docker Compose** (for NATS/JetStream)
- **Git** (signed commits recommended)

### Running the Project

```bash
# Clone your fork
git clone https://github.com/<your-username>/RemLinkAgent.git
cd RemLinkAgent

# Add upstream
git remote add upstream https://github.com/burakhalefoglu/RemLinkAgent.git

# Install dependencies & run
make docker-up      # start NATS + server
make cli            # build the CLI
make mobile         # build the Flutter app
```

> For the full architecture and development roadmap, see [`docs/roadmap.md`](docs/roadmap.md).

---

## 📋 Development Workflow (Pull Requests)

### 1. Discuss first
Making a significant change (new feature, architectural refactor)? Open an **Issue** and discuss it first. This prevents you from writing code that may be rejected.

### 2. Create a branch
```bash
git checkout -b feat/descriptive-branch-name
# or: fix/issue-123, docs/update, chore/dependency
```

**Branch naming:**
- `feat/...` — new feature
- `fix/...` — bug fix
- `docs/...` — documentation
- `chore/...` — maintenance, dependencies, CI
- `refactor/...` — code restructuring (no behavior change)

### 3. Write code & test
```bash
make lint           # golangci-lint (Go) + flutter analyze (Dart)
make test           # all tests
make license-check  # ⚠️ license header check (mandatory!)
```

### 4. Commit conventions
Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <short description>

<longer description (optional)>
```

**Examples:**
```
feat(cli): add context normalization for model hot-swap
fix(server): fix heartbeat loss during WSS reconnect
docs(roadmap): detail mobile architecture for Phase 3
chore(deps): bump go-openai to v1.40.0
```

**Types:** `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `perf`, `ci`

### 5. Add the license header (MANDATORY)
Every **new** source file must start with the AGPL v3 header:

```bash
make license-add     # automatically adds missing headers
```

See [`docs/license-ön-eki.md`](docs/license-ön-eki.md) for the header template and per-language examples. **CI will reject the PR if a header is missing.**

### 6. Open a Pull Request
- Fill out the PR template ([`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md)).
- Link the related Issue with `Closes #123`.
- Make sure CI is green.

---

## ✅ Code Standards

### General rules
- **Go:** `gofmt` + `goimports` + `golangci-lint`. Return errors with `error`, don't panic.
- **Dart/Flutter:** `dart format` + `flutter analyze`. Follow the Effective Dart guide.
- **Tests:** Write tests for new features/fixes. Goal: coverage of changed code must not drop.
- **Comments:** Explain "why", not "what". Comment non-obvious code sections.
- **Dependencies:** Must be AGPL v3-compatible (MIT, BSD, Apache 2.0, MPL are fine; avoid GPLv2-only).

### Architectural compliance
- **Zero-Touch AI:** AI traffic and API keys must NEVER pass through the server. Contributions violating this principle will be rejected.
- **Fail-Loud:** Don't silently report "passed" on errors. The system must halt loudly.
- **Lightness:** CLI RAM < 30MB. Avoid adding unnecessary dependencies.

> For the complete architectural decisions, see [`docs/project.md`](docs/project.md) (PRD) and [`docs/loopeng.md`](docs/loopeng.md).

---

## 🐛 Bug Reports

A good bug report includes:
- A clear title (summarizing the problem in one sentence)
- **Reproduction steps** (1, 2, 3...)
- Expected behavior vs. actual behavior
- Version info (CLI/Server/OS/model)
- Screenshots / logs if available

---

## 💰 Financial Model & Contributions (Transparency)

RemLinkAgent has two usage models:

- **Self-host:** Always **free** and **fully-featured** (AGPL v3). Run it on any server you like.
- **Managed Cloud:** Coming soon — $5/mo Pro plan, offering hosting/push/uptime as a "convenience fee".

**Important:** Contributing does not entitle you to an automatic share of Managed Cloud revenue (see the License section above). AGPL v3 protects your contributions, but commercial revenue belongs to Burak Halefoğlu, who maintains the project.

That said, a voluntary sponsorship mechanism via **OpenCollective / GitHub Sponsors** may be considered in the future to give back to the community. This is optional and not currently active.

---

## ❓ Questions?

- Technical question → [GitHub Discussions](../../discussions) (if available) or an Issue
- Security vulnerability → see [`SECURITY.md`](SECURITY.md) (responsible disclosure process)
- License question → Open an Issue and label it `question`

---

## 🌟 Code of Conduct

This community expects everyone to treat each other with respect. Personal attacks, harassment, and discrimination are **not tolerated**. Be constructive and kind. If you experience an issue, report it via an Issue (or email if it needs to be confidential).

---

*Your contributions make RemLinkAgent better for everyone. Thanks again!* ❤️
