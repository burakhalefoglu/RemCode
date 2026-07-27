// RemLinkAgent — AI coding agent for your machine, driven from your phone
// Copyright (C) 2026 Burak Halefoğlu
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Command gate runs RemLinkAgent's own Loop Engineering verification tiers
// against this repository.
//
// This is dogfooding: the tiered pipeline described in docs/loop-engineering.md
// is a product feature scheduled for X6, and it is also how the project is
// built in the meantime. Only the deterministic gates live here — the ones
// requiring judgement are listed by `gate verify` as review obligations rather
// than being silently skipped.
//
//	go run ./scripts/gate t0        # after every edit
//	go run ./scripts/gate t1        # every fix iteration
//	go run ./scripts/gate t2        # at convergence, once per feature
//	go run ./scripts/gate t3        # when a feature is candidate-complete
//	go run ./scripts/gate verify    # before a human tests the interface
//	go run ./scripts/gate canary    # prove the gates can still detect breakage
//	go run ./scripts/gate spec      # spec artifact status
//
// Exit codes: 0 passed · 1 failed · 4 could not verify.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	exitOK         = 0
	exitFailed     = 1
	exitUsage      = 2
	exitUnverified = 4
)

func main() {
	// Go's flag package stops at the first non-flag argument, so `gate verify
	// -no-cache` would silently ignore the flag. Lift the subcommand out first
	// and let flags appear on either side of it — a flag that is quietly
	// dropped is worse than one that errors.
	cmd, rest := splitCommand(os.Args[1:])
	if cmd == "" {
		usage()
		os.Exit(exitUsage)
	}

	fs := flag.NewFlagSet("gate", flag.ExitOnError)
	fs.Usage = usage
	noCache := fs.Bool("no-cache", false, "ignore the selective regression cache")
	verbose := fs.Bool("v", false, "show findings for passing gates too")
	if err := fs.Parse(rest); err != nil {
		os.Exit(exitUsage)
	}
	flag.CommandLine = fs

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot locate repository root: %v\n", err)
		os.Exit(exitUnverified)
	}

	switch cmd {
	case "t0", "t1", "t2", "t3":
		tier := int(cmd[1] - '0')
		os.Exit(runTiers(root, []int{tier}, *noCache, *verbose))

	case "verify":
		os.Exit(runVerify(root, *noCache, *verbose))

	case "canary":
		os.Exit(runCanary())

	case "spec":
		os.Exit(reportSpecs(root))

	case "clear-cache":
		if err := loadCache(root).clear(); err != nil {
			fmt.Fprintf(os.Stderr, "clear cache: %v\n", err)
			os.Exit(exitUnverified)
		}
		fmt.Println("cache cleared")

	case "help", "-h", "--help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(exitUsage)
	}
}

func usage() {
	fmt.Fprint(flag.CommandLine.Output(), `gate — RemLinkAgent Loop Engineering verification

Usage:
  gate <command> [flags]

Commands:
  t0           Instant gates      — format, compile, vet
  t1           Inner loop         — lint, tests, conformance, fake-green
  t2           Convergence        — coverage ratchet, spec fidelity, licences
  t3           Heavy              — race detector, vulnerability scan
  verify       Everything, plus the checkpoint ① spec-ratification rule
  canary       Prove each gate still detects deliberate breakage
  spec         Spec artifact status
  clear-cache  Discard the selective regression cache

Flags:
  -no-cache    Re-run gates even when their inputs are unchanged
  -v           Show detail for passing gates

Exit codes: 0 passed · 1 failed · 4 could not verify
Reference: docs/development-loop.md
`)
}

// ── registry ────────────────────────────────────────────────────────────────

// registry is the complete set of gate definitions.
//
// These live in compiled Go, not in a config file, because the loop is
// rewarded for turning gates green and must not be able to weaken one. Editing
// this list is a code change that shows up in review.
//
//gate:spec-fixtures — canary SPEC ids below are planted data, not citations.
func registry() []Check {
	goInputs := []string{"cmd", "internal", "scripts", "go.mod"}

	return []Check{
		{
			ID: "gofmt", Tier: 0, Desc: "Go formatting",
			Inputs: goInputs, Run: checkGofmt,
			CanaryFiles: map[string]string{
				"bad.go": "package  x\nfunc  F( ) {\n\t\t}\n",
			},
		},
		{
			ID: "build", Tier: 0, Desc: "compiles",
			Inputs: goInputs, Run: checkBuild,
			CanaryFiles: map[string]string{
				"go.mod":    "module canary\n\ngo 1.24\n",
				"broken.go": "package canary\n\nfunc F() int { return }\n",
			},
		},
		{
			ID: "vet", Tier: 0, Desc: "suspicious constructs",
			Inputs: goInputs, Run: checkVet,
			CanaryFiles: map[string]string{
				"go.mod": "module canary\n\ngo 1.24\n",
				"vet.go": "package canary\n\nimport \"fmt\"\n\nfunc F() { fmt.Printf(\"%d\", \"not a number\") }\n",
			},
		},

		{ID: "lint", Tier: 1, Desc: "golangci-lint", Inputs: goInputs, Run: checkLint},
		{
			ID: "tests", Tier: 1, Desc: "unit tests",
			Inputs: goInputs, Run: checkTests,
			CanaryFiles: map[string]string{
				"go.mod":       "module canary\n\ngo 1.24\n",
				"fail_test.go": "package canary\n\nimport \"testing\"\n\nfunc TestPlanted(t *testing.T) { t.Fatal(\"planted failure\") }\n",
			},
		},
		{
			ID: "zero-touch-ai", Tier: 1, Desc: "relay cannot reach providers or keys",
			Inputs: goInputs, Run: checkZeroTouchAI,
			CanaryFiles: map[string]string{
				"internal/server/relay.go": "package server\n\nimport _ \"example.com/canary/internal/models\"\n",
			},
		},
		{
			ID: "secret-logging", Tier: 1, Desc: "no credentials in log calls",
			Inputs: goInputs, Run: checkSecretLogging,
			CanaryFiles: map[string]string{
				"leak.go": "package canary\n\nimport \"log/slog\"\n\nfunc F(apiKey string) { slog.Info(\"starting\", apiKey) }\n",
			},
		},
		{
			ID: "fake-green", Tier: 1, Desc: "tests that assert nothing",
			Inputs: goInputs, Run: checkFakeGreen,
			CanaryFiles: map[string]string{
				"hollow_test.go": "package canary\n\nimport \"testing\"\n\nfunc TestNothing(t *testing.T) {\n\t_ = 1 + 1\n}\n",
			},
		},
		{
			ID: "spec-hygiene", Tier: 1, Desc: "spec artifacts are well formed",
			Inputs: []string{".rla/specs"}, Run: checkSpecHygiene,
			CanaryFiles: map[string]string{
				".rla/specs/mismatch.md": "---\nid: wrong-name\ntitle: Canary\nstatus: ratified\n---\n\n## SPEC-wrong-name-01 — Something\n",
			},
		},

		{
			ID: "spec-fidelity", Tier: 2, Desc: "every ratified requirement is implemented",
			Inputs: []string{".rla/specs", "cmd", "internal", "scripts"}, Run: checkSpecFidelity,
			CanaryFiles: map[string]string{
				".rla/specs/orphan.md": "---\nid: orphan\ntitle: Canary\nstatus: ratified\n---\n\n## SPEC-orphan-01 — Never implemented\n",
			},
		},
		{
			ID: "coverage-ratchet", Tier: 2, Desc: "coverage must not regress",
			Run: checkCoverage,
			CanaryFiles: map[string]string{
				"go.mod":          "module canary\n\ngo 1.24\n",
				"x.go":            "package canary\n\nfunc Covered() int { return 1 }\n\nfunc Uncovered() int { return 2 }\n",
				"x_test.go":       "package canary\n\nimport \"testing\"\n\nfunc TestCovered(t *testing.T) {\n\tif Covered() != 1 {\n\t\tt.Fatal(\"unreachable\")\n\t}\n}\n",
				coverageFloorFile: "99.0\n",
			},
		},
		{ID: "doc-links", Tier: 2, Desc: "documentation links and anchors", Inputs: []string{"docs", "README.md", "CONTRIBUTING.md"}, Run: checkDocLinks},
		{
			ID: "licence-headers", Tier: 2, Desc: "AGPL headers present",
			Inputs: goInputs, Run: checkLicenceHeaders,
			CanaryFiles: map[string]string{
				"LICENSE_HEADER": "Canary\nCopyright (C) {{ .Year }} {{ .Holder }}\n",
				"cmd/bare.go":    "package main\n\nfunc main() {}\n",
			},
		},
		{
			ID: "licence-boundary", Tier: 2, Desc: "no AGPL code under mobile/",
			Inputs: []string{"mobile"}, Run: checkLicenceBoundary,
			CanaryFiles: map[string]string{
				"mobile/lib/leak.dart": "// GNU Affero General Public License\nvoid main() {}\n",
			},
		},

		{ID: "race", Tier: 3, Desc: "data race detector", Inputs: goInputs, Run: checkRace},
		{ID: "vulnerabilities", Tier: 3, Desc: "known CVEs in dependencies", Inputs: []string{"go.mod", "go.sum"}, Run: checkVulnerabilities},
	}
}

// judgementGates are the parts of the pipeline no script can decide. They are
// listed rather than silently omitted, because an unlisted obligation is
// indistinguishable from one that was met.
var judgementGates = []string{
	"backward spec diff — behaviour in the code that no SPEC id covers (Tier 2)",
	"architectural intent — does the design still match PRINCIPLES.md (Tier 2)",
	"black-box exploration — adversarial probing without reading the source (Tier 3)",
	"spec correctness — checkpoint ①, is the plan itself right (human)",
	"interface test — checkpoint ②, does it actually work end to end (human)",
}

// ── runners ─────────────────────────────────────────────────────────────────

func runTiers(root string, tiers []int, noCache, verbose bool) int {
	specs, err := loadSpecs(root)
	if err != nil {
		fmt.Printf("⚠️  spec artifacts unreadable: %v\n", err)
		specs = nil
	}
	var specIDs []string
	for _, s := range specs {
		if s.Active() {
			for _, r := range s.Requirements {
				specIDs = append(specIDs, r.ID)
			}
		}
	}

	cache := loadCache(root)
	worst := Pass
	ran := 0

	for _, tier := range tiers {
		checks := checksForTier(tier)
		if len(checks) == 0 {
			continue
		}
		fmt.Printf("\n\033[1mTier %d\033[0m — %s\n", tier, tierName(tier))

		for _, check := range checks {
			sig, sigErr := signature(root, check, specIDs)
			if sigErr == nil && !noCache && sig != "" {
				if summary, ok := cache.hit(check.ID, sig); ok {
					report(Result{Name: check.ID, Tier: tier, Status: Pass, Summary: summary, Cached: true}, verbose)
					continue
				}
			}

			start := time.Now()
			res := check.Run(root)
			res.Duration = time.Since(start)
			res.Name = check.ID
			res.Tier = tier
			ran++

			if res.Status == Pass && sigErr == nil && sig != "" {
				cache.store(check.ID, sig, res.Summary)
			}
			report(res, verbose)
			worst = worsen(worst, res.Status)
		}
	}

	if err := cache.save(); err != nil {
		fmt.Printf("⚠️  cache not saved: %v\n", err)
	}

	fmt.Printf("\n%s  %d gates run, %d cached\n", worst.icon(), ran, len(checksForTiers(tiers))-ran)
	return statusExit(worst)
}

func runVerify(root string, noCache, verbose bool) int {
	fmt.Println("\033[1mFull verification\033[0m — read-only sweep before checkpoint ②")

	code := runTiers(root, []int{0, 1, 2, 3}, noCache, verbose)

	// Checkpoint ①: a draft spec means the plan was never agreed, so no amount
	// of green gates can justify declaring the work ready.
	specs, err := loadSpecs(root)
	if err != nil {
		fmt.Printf("\n⚠️  spec artifacts unreadable: %v\n", err)
		return exitUnverified
	}

	var drafts []string
	for _, s := range specs {
		if s.Draft() {
			drafts = append(drafts, fmt.Sprintf("%s (%s)", s.ID, s.Phase))
		}
	}

	fmt.Println("\n\033[1mJudgement gates\033[0m — not decidable by script, still owed:")
	for _, g := range judgementGates {
		fmt.Printf("  ·  %s\n", g)
	}

	fmt.Println()
	switch {
	case code == exitUnverified:
		fmt.Println("⚠️   COULD NOT VERIFY — a gate did not run. Not ready.")
	case code != exitOK:
		fmt.Println("🔴  FAILED — gates are red. Not ready.")
	case len(drafts) > 0:
		fmt.Printf("🔴  NOT READY — %d spec(s) still awaiting ratification: %s\n",
			len(drafts), strings.Join(drafts, ", "))
		fmt.Println("    Checkpoint ① is unmet: ratify on mobile (or set status: ratified) first.")
		return exitFailed
	default:
		fmt.Println("🟢  READY FOR INTERFACE TEST — every automated gate passed.")
		fmt.Println("    The judgement gates above are yours. Green means \"follows the rules\",")
		fmt.Println("    not \"does the right thing\".")
	}
	return code
}

// runCanary deliberately takes no root: every canary runs against a freshly
// planted temporary tree, never against the repository.
func runCanary() int {
	fmt.Println("\033[1mCanary\033[0m — each gate is fed deliberate breakage and must catch it")

	worst := Pass
	for _, check := range registry() {
		if len(check.CanaryFiles) == 0 {
			fmt.Printf("  ⚠️  %-18s no canary — this gate's ability to detect breakage is unproven\n", check.ID)
			worst = worsen(worst, Unverified)
			continue
		}

		dir, err := plant(check.CanaryFiles)
		if err != nil {
			fmt.Printf("  ⚠️  %-18s canary setup failed: %v\n", check.ID, err)
			worst = worsen(worst, Unverified)
			continue
		}

		res := check.Run(dir)
		_ = os.RemoveAll(dir)

		if res.Status == Fail {
			fmt.Printf("  🟢 %-18s detected planted breakage\n", check.ID)
			continue
		}
		fmt.Printf("  🔴 %-18s reported %s on broken input — THIS GATE IS BLIND\n", check.ID, res.Status)
		worst = worsen(worst, Fail)
	}

	fmt.Println()
	switch worst {
	case Pass:
		fmt.Println("🟢  every gate with a canary can still detect breakage")
	case Unverified:
		fmt.Println("⚠️   some gates have no canary — their green is worth less than it looks")
	default:
		fmt.Println("🔴  a gate failed to detect deliberate breakage — fix it before trusting any pass")
	}
	return statusExit(worst)
}

func reportSpecs(root string) int {
	specs, err := loadSpecs(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return exitUnverified
	}
	if len(specs) == 0 {
		fmt.Println("no spec artifacts under .rla/specs/")
		return exitUnverified
	}

	index, err := annotationIndex(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "source scan: %v\n", err)
		return exitUnverified
	}

	fmt.Printf("\033[1m%-22s %-10s %-6s %s\033[0m\n", "SPEC", "STATUS", "PHASE", "IMPLEMENTED")
	drafts := 0
	for _, s := range specs {
		done := 0
		for _, r := range s.Requirements {
			if len(index[r.ID]) > 0 {
				done++
			}
		}
		mark := "  "
		switch {
		case s.Draft():
			mark, drafts = "① ", drafts+1
		case s.Active() && done == len(s.Requirements):
			mark = "🟢"
		case s.Active():
			mark = "🔴"
		}
		fmt.Printf("%s %-20s %-10s %-6s %d/%d\n", mark, s.ID, s.Status, s.Phase, done, len(s.Requirements))
	}

	if drafts > 0 {
		fmt.Printf("\n① %d spec(s) await ratification — the pipeline is halted at checkpoint ①.\n", drafts)
	}
	return exitOK
}

// ── helpers ─────────────────────────────────────────────────────────────────

func checksForTier(tier int) []Check {
	var out []Check
	for _, c := range registry() {
		if c.Tier == tier {
			out = append(out, c)
		}
	}
	return out
}

func checksForTiers(tiers []int) []Check {
	var out []Check
	for _, t := range tiers {
		out = append(out, checksForTier(t)...)
	}
	return out
}

func tierName(tier int) string {
	switch tier {
	case 0:
		return "instant · every edit"
	case 1:
		return "inner loop · every fix iteration"
	case 2:
		return "convergence · once per feature"
	case 3:
		return "heavy · once, candidate-complete"
	default:
		return "periodic"
	}
}

func report(r Result, verbose bool) {
	suffix := ""
	switch {
	case r.Cached:
		suffix = " \033[2m(cached)\033[0m"
	case r.Duration > 0:
		suffix = fmt.Sprintf(" \033[2m(%s)\033[0m", r.Duration.Round(10*time.Millisecond))
	}

	summary := r.Summary
	if summary == "" {
		summary = r.Status.String()
	}
	fmt.Printf("  %s %-18s %s%s\n", r.Status.icon(), r.Name, summary, suffix)

	if r.Status == Pass && !verbose {
		return
	}
	for _, f := range r.Findings {
		fmt.Printf("       %s\n", f)
	}
}

func worsen(current, next Status) Status {
	// Fail outranks Unverified: a known break is more actionable than an
	// unknown, but either one blocks.
	if next == Fail || current == Fail {
		return Fail
	}
	if next == Unverified || current == Unverified {
		return Unverified
	}
	return Pass
}

func statusExit(s Status) int {
	switch s {
	case Pass:
		return exitOK
	case Fail:
		return exitFailed
	default:
		return exitUnverified
	}
}

// splitCommand pulls the first non-flag argument out of argv and returns it
// along with everything else, so flags may precede or follow the subcommand.
func splitCommand(argv []string) (cmd string, rest []string) {
	for i, arg := range argv {
		if strings.HasPrefix(arg, "-") {
			rest = append(rest, arg)
			continue
		}
		rest = append(rest, argv[i+1:]...)
		return arg, rest
	}
	return "", rest
}

// plant writes canary files into a fresh temporary directory.
func plant(files map[string]string) (string, error) {
	dir, err := os.MkdirTemp("", "rla-canary-*")
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(files[name]), 0o600); err != nil {
			return "", err
		}
	}
	return dir, nil
}
