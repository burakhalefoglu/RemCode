# RemLinkAgent — AI coding agent for your machine, driven from your phone
# Copyright (C) 2026 Burak Halefoğlu
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

<#
.SYNOPSIS
    Windows mirror of the Makefile. GNU make is not installed by default on
    Windows, so this keeps `make <target>` workflows reachable without it.

.EXAMPLE
    .\scripts\make.ps1 help
    .\scripts\make.ps1 cli
    .\scripts\make.ps1 ci

.NOTES
    The Makefile is canonical. Any target added there should be added here.
#>

[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string]$Target = 'help'
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

# ── Build metadata ───────────────────────────────────────────────────────────

function Get-GitOr([string]$Command, [string]$Fallback) {
    try {
        $out = Invoke-Expression $Command 2>$null
        if ($LASTEXITCODE -eq 0 -and $out) { return "$out".Trim() }
    } catch { }
    return $Fallback
}

$Version   = Get-GitOr 'git describe --tags --always --dirty' '0.0.0-dev'
$Commit    = Get-GitOr 'git rev-parse --short HEAD' 'unknown'
$BuildDate = Get-GitOr 'git log -1 --format=%cI' 'unknown'

$Pkg     = 'github.com/burakhalefoglu/RemLinkAgent'
$LdFlags = "-s -w " +
           "-X '$Pkg/internal/version.version=$Version' " +
           "-X '$Pkg/internal/version.commit=$Commit' " +
           "-X '$Pkg/internal/version.date=$BuildDate'"

$Holder     = 'Burak Halefoğlu'
$Year       = '2026'
$AgplDirs   = @('cmd', 'internal', 'deploy', 'scripts')
$ApacheDirs = @('mobile/lib', 'mobile/test')
$LicIgnore  = @(
    '-ignore', '**/*.md', '-ignore', '**/*.json', '-ignore', '**/*.lock',
    '-ignore', '**/vendor/**', '-ignore', '**/.dart_tool/**', '-ignore', '**/build/**',
    '-ignore', '**/generated_plugin_registrant.dart', '-ignore', '**/*.g.dart',
    '-ignore', '**/*.freezed.dart'
)

function Test-Tool([string]$Name) { $null -ne (Get-Command $Name -ErrorAction SilentlyContinue) }
function Existing([string[]]$Dirs) { $Dirs | Where-Object { Test-Path $_ } }
function Fail([string]$Message) { Write-Host $Message -ForegroundColor Red; exit 1 }
function Step([string]$Message) { Write-Host "→ $Message" -ForegroundColor Cyan }

# Native commands set $LASTEXITCODE rather than throwing; surface real failures.
function Invoke-Checked {
    param([scriptblock]$Block, [string]$What)
    & $Block
    if ($LASTEXITCODE -ne 0) { Fail "$What failed (exit $LASTEXITCODE)" }
}

# ── Targets ──────────────────────────────────────────────────────────────────

function Target-Help {
    Write-Host ''
    Write-Host 'RemLinkAgent — make targets (PowerShell mirror)' -ForegroundColor White
    Write-Host ''
    $targets = [ordered]@{
        'build'         = 'Build both binaries'
        'cli'           = 'Build the CLI agent -> bin/rla.exe'
        'server'        = 'Build the relay server -> bin/rla-server.exe'
        'install'       = 'go install both binaries into GOPATH/bin'
        'test'          = 'Run Go tests'
        'test-race'     = 'Run Go tests with the race detector'
        'cover'         = 'Coverage report -> coverage.out / coverage.html'
        'lint'          = 'Run golangci-lint (and flutter analyze when present)'
        'vet'           = 'go vet'
        'fmt'           = 'Format Go (and Dart when present)'
        'tidy'          = 'go mod tidy'
        'mobile'        = 'Build/run the Flutter app'
        'mobile-test'   = 'Run Flutter tests'
        'docker-up'     = 'Start NATS + relay server'
        'docker-down'   = 'Stop the stack'
        'docker-logs'   = 'Tail stack logs'
        'license-check' = 'Verify every source file carries the right header'
        'license-add'   = 'Add the correct header to files missing one'
        'tools'         = 'Install dev tooling (addlicense, golangci-lint)'
        'og-image'      = 'Export docs/og-image.svg -> og-image.png'
        'docs'          = 'Verify every relative doc link and heading anchor'
        't0'            = 'Tier 0 - format, compile, vet (after every edit)'
        't1'            = 'Tier 1 - lint, tests, conformance, fake-green'
        't2'            = 'Tier 2 - coverage, spec fidelity, licences'
        't3'            = 'Tier 3 - race detector, CVE scan'
        'verify'        = 'Full sweep + checkpoint (1) - ready for a human?'
        'canary'        = 'Prove each gate still detects deliberate breakage'
        'spec'          = 'Spec artifact status'
        'version'       = 'Print build metadata'
        'clean'         = 'Remove build artefacts'
        'ci'            = 'Everything CI runs'
    }
    foreach ($k in $targets.Keys) {
        Write-Host ('  {0,-16}' -f $k) -ForegroundColor Cyan -NoNewline
        Write-Host $targets[$k]
    }
    Write-Host ''
    Write-Host 'Status: pre-alpha (P0). See docs/roadmap.md' -ForegroundColor DarkGray
    Write-Host ''
}

function Target-Cli {
    if (-not (Test-Path bin)) { New-Item -ItemType Directory bin | Out-Null }
    Step 'building bin/rla.exe'
    Invoke-Checked { go build -trimpath -ldflags $LdFlags -o bin/rla.exe ./cmd/rla } 'go build (rla)'
    Write-Host "built bin/rla.exe ($Version)" -ForegroundColor Green
}

function Target-Server {
    if (-not (Test-Path bin)) { New-Item -ItemType Directory bin | Out-Null }
    Step 'building bin/rla-server.exe'
    Invoke-Checked { go build -trimpath -ldflags $LdFlags -o bin/rla-server.exe ./cmd/rla-server } 'go build (rla-server)'
    Write-Host "built bin/rla-server.exe ($Version)" -ForegroundColor Green
}

function Target-Build { Target-Cli; Target-Server }

function Target-Install {
    Invoke-Checked { go install -trimpath -ldflags $LdFlags ./cmd/... } 'go install'
}

function Target-Test     { Step 'go test ./...'; Invoke-Checked { go test ./... } 'go test' }
function Target-TestRace  { Step 'go test -race ./...'; Invoke-Checked { go test -race ./... } 'go test -race' }
function Target-Vet      { Step 'go vet ./...'; Invoke-Checked { go vet ./... } 'go vet' }

function Target-Cover {
    Invoke-Checked { go test -coverprofile=coverage.out ./... } 'go test -cover'
    Invoke-Checked { go tool cover -html=coverage.out -o coverage.html } 'go tool cover'
    Write-Host 'report: coverage.html' -ForegroundColor Green
}

function Target-Fmt {
    Invoke-Checked { go fmt ./... } 'go fmt'
    if ((Test-Path mobile/pubspec.yaml) -and (Test-Tool flutter)) {
        Push-Location mobile; try { dart format lib test } finally { Pop-Location }
    }
}

function Target-Tidy { Invoke-Checked { go mod tidy } 'go mod tidy' }

function Target-Lint {
    Target-Vet
    if (Test-Tool golangci-lint) {
        Step 'golangci-lint run'
        Invoke-Checked { golangci-lint run ./... } 'golangci-lint'
    } else {
        Write-Host 'skip: golangci-lint not installed -> .\scripts\make.ps1 tools' -ForegroundColor Yellow
    }
    if (Test-Path mobile/pubspec.yaml) {
        Push-Location mobile; try { flutter analyze } finally { Pop-Location }
    } else {
        Write-Host 'skip: mobile/ not bootstrapped (P0.7)' -ForegroundColor Yellow
    }
}

function Target-Mobile {
    if (-not (Test-Path mobile/pubspec.yaml)) {
        Write-Host 'mobile/ is not bootstrapped yet — this is P0.7.' -ForegroundColor Yellow
        Write-Host ''
        Write-Host '  flutter create --org com.remlinkagent --project-name rla_mobile `'
        Write-Host '                 --platforms=ios,android mobile'
        Write-Host ''
        Write-Host 'Keep mobile/LICENSE and mobile/README.md — they carry the Apache-2.0'
        Write-Host 'boundary described in ADR-002.'
        exit 1
    }
    Push-Location mobile
    try { flutter pub get; flutter run } finally { Pop-Location }
}

function Target-MobileTest {
    if (-not (Test-Path mobile/pubspec.yaml)) { Fail 'mobile/ not bootstrapped (P0.7)' }
    Push-Location mobile; try { flutter test } finally { Pop-Location }
}

function Target-DockerUp {
    Push-Location deploy
    try {
        Invoke-Checked { docker compose up -d } 'docker compose up'
        Write-Host 'NATS on :4222 (monitor :8222) - server on :8080' -ForegroundColor Green
    } finally { Pop-Location }
}

function Target-DockerDown {
    Push-Location deploy; try { docker compose down } finally { Pop-Location }
}

function Target-DockerLogs {
    Push-Location deploy; try { docker compose logs -f } finally { Pop-Location }
}

function Invoke-AddLicense {
    param([string]$Template, [string[]]$Dirs, [switch]$CheckOnly)
    $dirs = Existing $Dirs
    if (-not $dirs) { return }
    $argv = @()
    if ($CheckOnly) { $argv += '-check' }
    $argv += @('-f', $Template, '-c', $Holder, '-y', $Year) + $LicIgnore + $dirs
    & addlicense @argv
    if ($LASTEXITCODE -ne 0) { Fail "licence header check failed for: $($dirs -join ', ')" }
}

function Target-LicenseCheck {
    if (-not (Test-Tool addlicense)) { Fail 'addlicense missing -> .\scripts\make.ps1 tools' }
    Step "checking AGPL headers: $((Existing $AgplDirs) -join ', ')"
    Invoke-AddLicense -Template LICENSE_HEADER -Dirs $AgplDirs -CheckOnly
    if (Test-Path mobile/lib) {
        Step 'checking Apache-2.0 headers: mobile/'
        Invoke-AddLicense -Template LICENSE_HEADER_APACHE -Dirs $ApacheDirs -CheckOnly
    }
    Write-Host 'licence headers OK' -ForegroundColor Green
}

function Target-LicenseAdd {
    if (-not (Test-Tool addlicense)) { Fail 'addlicense missing -> .\scripts\make.ps1 tools' }
    Invoke-AddLicense -Template LICENSE_HEADER -Dirs $AgplDirs
    if (Test-Path mobile/lib) {
        Invoke-AddLicense -Template LICENSE_HEADER_APACHE -Dirs $ApacheDirs
    }
}

function Target-Tools {
    Invoke-Checked { go install github.com/google/addlicense@latest } 'install addlicense'
    Invoke-Checked { go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest } 'install golangci-lint'
    Write-Host "installed into $(go env GOPATH)\bin — ensure it is on PATH" -ForegroundColor Green
}

function Target-OgImage {
    if (Test-Tool rsvg-convert)   { rsvg-convert -w 1200 -h 630 docs/og-image.svg -o docs/og-image.png }
    elseif (Test-Tool magick)     { magick -background none -density 144 docs/og-image.svg -resize 1200x630 docs/og-image.png }
    elseif (Test-Tool inkscape)   { inkscape docs/og-image.svg -w 1200 -h 630 -o docs/og-image.png }
    else { Fail 'no converter found — install librsvg, ImageMagick or Inkscape' }
    Write-Host 'wrote docs/og-image.png — now point og:image at it in docs/index.html' -ForegroundColor Green
}

function Target-Docs {
    Step 'checking documentation links'
    Invoke-Checked { go run ./scripts/checkdocs } 'checkdocs'
}

# Loop Engineering gates — docs/development-loop.md
function Invoke-Gate([string]$Command) {
    & go run ./scripts/gate $Command
    # 0 passed - 1 failed - 4 could not verify. Anything non-zero blocks.
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

function Target-T0     { Invoke-Gate 't0' }
function Target-T1     { Invoke-Gate 't1' }
function Target-T2     { Invoke-Gate 't2' }
function Target-T3     { Invoke-Gate 't3' }
function Target-Verify { Invoke-Gate 'verify' }
function Target-Canary { Invoke-Gate 'canary' }
function Target-Spec   { Invoke-Gate 'spec' }

function Target-Version {
    Write-Host "version : $Version"
    Write-Host "commit  : $Commit"
    Write-Host "date    : $BuildDate"
}

function Target-Clean {
    foreach ($p in @('bin', 'coverage.out', 'coverage.html')) {
        if (Test-Path $p) { Remove-Item -Recurse -Force $p }
    }
    if ((Test-Path mobile/pubspec.yaml) -and (Test-Tool flutter)) {
        Push-Location mobile; try { flutter clean } finally { Pop-Location }
    }
    Write-Host 'cleaned' -ForegroundColor Green
}

function Target-Ci { Target-Fmt; Target-Vet; Target-Lint; Target-Test; Target-LicenseCheck; Target-Docs }

# ── Dispatch ─────────────────────────────────────────────────────────────────

switch ($Target.ToLowerInvariant()) {
    'help'          { Target-Help }
    'build'         { Target-Build }
    'cli'           { Target-Cli }
    'server'        { Target-Server }
    'install'       { Target-Install }
    'test'          { Target-Test }
    'test-race'     { Target-TestRace }
    'cover'         { Target-Cover }
    'lint'          { Target-Lint }
    'vet'           { Target-Vet }
    'fmt'           { Target-Fmt }
    'tidy'          { Target-Tidy }
    'mobile'        { Target-Mobile }
    'mobile-test'   { Target-MobileTest }
    'docker-up'     { Target-DockerUp }
    'docker-down'   { Target-DockerDown }
    'docker-logs'   { Target-DockerLogs }
    'license-check' { Target-LicenseCheck }
    'license-add'   { Target-LicenseAdd }
    'tools'         { Target-Tools }
    'og-image'      { Target-OgImage }
    'docs'          { Target-Docs }
    't0'            { Target-T0 }
    't1'            { Target-T1 }
    't2'            { Target-T2 }
    't3'            { Target-T3 }
    'verify'        { Target-Verify }
    'canary'        { Target-Canary }
    'spec'          { Target-Spec }
    'version'       { Target-Version }
    'clean'         { Target-Clean }
    'ci'            { Target-Ci }
    default {
        Write-Host "unknown target: $Target" -ForegroundColor Red
        Target-Help
        exit 2
    }
}
