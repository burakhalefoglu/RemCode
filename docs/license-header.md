# 🔏 Licence Header Standard

> **Status:** MANDATORY · **Updated:** 2026-07-27
> Every source file in this repository must begin with the header for **its directory**. CI rejects files without one.

RemLinkAgent uses two licences. Which header a file gets depends entirely on where it lives — see [ADR-002](decisions.md#adr-002--split-licensing-agpl-core-apache-20-mobile) for why.

---

## 1. Which header goes where

| Path | Licence | Template |
| :--- | :--- | :--- |
| `cmd/`, `internal/`, `deploy/`, `scripts/` | **AGPL-3.0-or-later** | [`LICENSE_HEADER`](../LICENSE_HEADER) |
| `mobile/lib/`, `mobile/test/` | **Apache-2.0** | [`LICENSE_HEADER_APACHE`](../LICENSE_HEADER_APACHE) |

`make license-add` picks the right one by path. You should not need to copy headers by hand.

> ⚠️ **The boundary is one-way.** Apache-2.0 code may move into the AGPL core. AGPL code may **never** move into `mobile/` — it would make the mobile client AGPL, which is the exact problem the split exists to avoid. A CI job greps for AGPL text under `mobile/` and fails the build if it finds any.

---

## 2. Header text

### AGPL-3.0-or-later — core

```text
RemLinkAgent — AI coding agent for your machine, driven from your phone
Copyright (C) 2026 Burak Halefoğlu

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

> "version 3 of the License, or (at your option) any later version" means `AGPL-3.0-or-later`. To pin to v3 only, the clause would end at "version 3 of the License." This repository uses **or-later**.

### Apache-2.0 — mobile

```text
RemLinkAgent mobile client
Copyright (C) 2026 Burak Halefoğlu

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

---

## 3. Comment syntax by language

### `//` — Go, Dart, Swift, Kotlin, JS/TS, C, Java

```go
// RemLinkAgent — AI coding agent for your machine, driven from your phone
// Copyright (C) 2026 Burak Halefoğlu
//
// This program is free software: you can redistribute it and/or modify
// ... (full text)

package agent
```

### `#` — YAML, Shell, TOML, Dockerfile, Makefile, PowerShell

```yaml
# RemLinkAgent — AI coding agent for your machine, driven from your phone
# Copyright (C) 2026 Burak Halefoğlu
# ...
```

**Shebang stays first.** The header goes after it:

```bash
#!/usr/bin/env bash
# RemLinkAgent — AI coding agent for your machine, driven from your phone
# ...
```

### `<!-- -->` — HTML, XML, SVG, plist

```html
<!--
  RemLinkAgent — AI coding agent for your machine, driven from your phone
  Copyright (C) 2026 Burak Halefoğlu
  ...
-->
```

### `/* */` — CSS, SCSS

```css
/*
 * RemLinkAgent — AI coding agent for your machine, driven from your phone
 * Copyright (C) 2026 Burak Halefoğlu
 * ...
 */
```

---

## 4. Automation

[`google/addlicense`](https://github.com/google/addlicense) handles comment syntax per language and preserves shebangs.

```bash
make tools           # installs addlicense + golangci-lint
make license-add     # adds the correct header where missing
make license-check   # verifies; this is what CI runs
```

Windows without GNU make:

```powershell
.\scripts\make.ps1 license-add
.\scripts\make.ps1 license-check
```

The templates use `{{ .Year }}` and `{{ .Holder }}`, filled from the `Makefile`'s `YEAR` and `HOLDER` variables. **Directory-to-template mapping lives in the `Makefile` only** — one definition, so it cannot drift.

### CI

Two jobs in [`license.yml`](../.github/workflows/license.yml):

1. **headers** — runs `make license-check`. A missing header fails the build with instructions.
2. **boundary** — greps for AGPL text under `mobile/` and verifies all four licence files exist.

---

## 5. Files that need no header

- Markdown (`*.md`)
- Lockfiles and generated output: `go.sum`, `pubspec.lock`, `package-lock.json`, `*.g.dart`, `*.freezed.dart`, `generated_plugin_registrant.dart`
- Build output
- Binary assets — images, fonts, audio
- `LICENSE` and `mobile/LICENSE` themselves
- Small config files: `.gitignore`, `.editorconfig`, `.gitattributes`

Excluded via `-ignore` flags in the `Makefile`.

---

## 6. Copyright as contributors arrive

Copyright accumulates. Significant contributions may extend the line:

```text
Copyright (C) 2026 Burak Halefoğlu
Copyright (C) 2026 Contributor Name <email>
```

Or, once there are several:

```text
Copyright (C) 2026 Burak Halefoğlu and contributors
```

The collective form is preferred. Contributors retain their copyright regardless — the [CLA](../CLA.md) grants a licence, it does not transfer ownership.

---

## 7. Licence compatibility

| Dependency licence | Core (AGPL) | `mobile/` (Apache-2.0) |
| :--- | :---: | :---: |
| MIT, BSD, ISC | ✅ | ✅ |
| Apache-2.0 | ✅ | ✅ |
| MPL-2.0 | ✅ | ⚠️ file-level copyleft — avoid |
| LGPL | ⚠️ dynamic linking only | ❌ |
| GPL-2.0-only | ❌ incompatible with AGPL-3.0 | ❌ |
| GPL-3.0 / AGPL-3.0 | ✅ | ❌ **breaks the split** |
| Proprietary | ❌ | ❌ |

A copyleft dependency in `mobile/` would defeat the entire reason for the Apache-2.0 client. CI checks `pubspec.lock` for it.

---

## 8. AGPL, BYOK and managed cloud

**API keys** belong to the user and are not part of the software. Outside the licence's scope.

**AI traffic** never traverses our server. This is an architectural fact and does not affect copyleft obligations either way.

**Managed cloud** is precisely why AGPL was chosen: it extends copyleft to network use. A competitor who takes the relay, modifies it and runs it as a service must publish their changes. GPL-3.0 would not have required this — that gap is the "SaaS loophole", and AGPL closes it.

**Mobile app** is Apache-2.0 and therefore has no such obligation. It also has no relay to hide: it is a client for a documented protocol.

**Commercial dual-licensing** remains possible because of the [CLA](../CLA.md), not because of the licence choice. See [ADR-003](decisions.md#adr-003--cla-required) — an earlier version of this document claimed the licence kept that door open, which was wrong.

> Have a lawyer review this section before relying on it commercially.
