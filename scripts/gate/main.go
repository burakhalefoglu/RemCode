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
// Tiers classify how often a check is worth running. Modes are what you
// actually run: `fast` after every change, `full` before anyone is asked for
// anything. Both write a machine-readable artifact under .rla/state/, which is
// what the judged half of the pipeline reads — a reviewing model reads the
// evidence, it does not re-run the tools.
//
//	go run ./scripts/gate fast      # every fix iteration
//	go run ./scripts/gate full      # at convergence
//	go run ./scripts/gate t0…t3     # one tier, when you want just that one
//	go run ./scripts/gate verify    # full, plus checkpoint ①
//	go run ./scripts/gate evidence  # the artifact a reviewer should read
//	go run ./scripts/gate canary    # prove the gates can still detect breakage
//	go run ./scripts/gate timings   # measure the wall clock, do not guess it
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

// runMode is a unit of running; a tier is a unit of classification. Conflating
// the two is what produces a pipeline nobody runs because every invocation
// costs the most expensive check in it.
//
// SPEC-deterministic-backbone-06
type runMode struct {
	Name  string
	Tiers []int
	// Budget is the mode's ceiling. It is not a target: the measured cost in
	// .rla/tool-timings.json is the number that matters, and a measurement
	// drifting towards its budget is the signal that work was quietly added to
	// a layer meant to stay cheap.
	Budget time.Duration
	Desc   string
}

var (
	modeFast = runMode{
		Name: "fast", Tiers: []int{0, 1}, Budget: 2 * time.Minute,
		Desc: "every fix iteration — static checks and the suite",
	}
	modeFull = runMode{
		Name: "full", Tiers: []int{0, 1, 2, 3}, Budget: 30 * time.Minute,
		Desc: "at convergence — everything, before a human is asked for anything",
	}
)

func tierMode(tier int) runMode {
	return runMode{Name: fmt.Sprintf("t%d", tier), Tiers: []int{tier}, Desc: tierName(tier)}
}

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
	mode := fs.String("mode", modeFull.Name, "which run's artifact to read (evidence)")
	record := fs.Bool("record", false, "commit the measurement as the new baseline (timings)")
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
		code, _ := runChecks(root, tierMode(int(cmd[1]-'0')), *noCache, *verbose)
		os.Exit(code)

	case "fast":
		code, _ := runChecks(root, modeFast, *noCache, *verbose)
		os.Exit(code)

	case "full":
		code, _ := runChecks(root, modeFull, *noCache, *verbose)
		os.Exit(code)

	case "verify":
		os.Exit(runVerify(root, *noCache, *verbose))

	case "evidence":
		os.Exit(reportEvidence(root, *mode))

	case "timings":
		os.Exit(runTimings(root, *record))

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

Run these:
  fast         Tiers 0–1  — after every change
  full         Tiers 0–3  — at convergence
  verify       full, plus the checkpoint ① spec-ratification rule

One tier at a time:
  t0           Instant            — format, compile, vet
  t1           Inner loop         — lint, tests, conformance, fake-green
  t2           Convergence        — coverage ratchet, spec fidelity, licences
  t3           Heavy              — race detector, vulnerability scan

About the run itself:
  evidence     The artifact a reviewing model should read instead of re-running
  timings      Measure each gate's wall clock (-record to set the baseline)
  canary       Prove each gate still detects deliberate breakage
  spec         Spec artifact status
  clear-cache  Discard the selective regression cache

Flags:
  -no-cache    Re-run gates even when their inputs are unchanged
  -v           Show detail for passing gates
  -mode        Which run's artifact to read: fast or full (evidence)
  -record      Write the measurement as the committed baseline (timings)

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
			Inputs: goInputs, Run: checkGofmt, Budget: time.Minute,
			CanaryFiles: map[string]string{
				"bad.go": "package  x\nfunc  F( ) {\n\t\t}\n",
			},
		},
		{
			ID: "build", Tier: 0, Desc: "compiles",
			Inputs: goInputs, Run: checkBuild, Budget: 5 * time.Minute,
			CanaryFiles: map[string]string{
				"go.mod":    "module canary\n\ngo 1.24\n",
				"broken.go": "package canary\n\nfunc F() int { return }\n",
			},
		},
		{
			ID: "vet", Tier: 0, Desc: "suspicious constructs",
			Inputs: goInputs, Run: checkVet, Budget: 5 * time.Minute,
			CanaryFiles: map[string]string{
				"go.mod": "module canary\n\ngo 1.24\n",
				"vet.go": "package canary\n\nimport \"fmt\"\n\nfunc F() { fmt.Printf(\"%d\", \"not a number\") }\n",
			},
		},

		{ID: "lint", Tier: 1, Desc: "golangci-lint", Inputs: goInputs, Run: checkLint, Budget: 8 * time.Minute},
		{
			ID: "tests", Tier: 1, Desc: "unit tests",
			Inputs: goInputs, Run: checkTests, Budget: 8 * time.Minute,
			// The suite's exit code says whether what ran passed. These say
			// whether anything ran, and whether it is still the same suite —
			// the two questions a green `go test` cannot answer.
			MinEvidence:  map[string]int{"tests_run": 1},
			BaselineKey:  "tests_run",
			BaselineFile: testBaselineFile,
			CanaryFiles: map[string]string{
				"go.mod":       "module canary\n\ngo 1.24\n",
				"fail_test.go": "package canary\n\nimport \"testing\"\n\nfunc TestPlanted(t *testing.T) { t.Fatal(\"planted failure\") }\n",
			},
		},
		{
			ID: "zero-touch-ai", Tier: 1, Desc: "relay cannot reach providers or keys",
			Inputs: goInputs, Run: checkZeroTouchAI, Budget: time.Minute,
			CanaryFiles: map[string]string{
				"internal/server/relay.go": "package server\n\nimport _ \"example.com/canary/internal/models\"\n",
			},
		},
		{
			ID: "secret-logging", Tier: 1, Desc: "no credentials in log calls",
			Inputs: goInputs, Run: checkSecretLogging, Budget: time.Minute,
			MinEvidence: map[string]int{"files_scanned": 1},
			CanaryFiles: map[string]string{
				"leak.go": "package canary\n\nimport \"log/slog\"\n\nfunc F(apiKey string) { slog.Info(\"starting\", apiKey) }\n",
			},
		},
		{
			ID: "fake-green", Tier: 1, Desc: "tests that assert nothing",
			Inputs: goInputs, Run: checkFakeGreen, Budget: time.Minute,
			MinEvidence: map[string]int{"tests_declared": 1},
			CanaryFiles: map[string]string{
				"hollow_test.go": "package canary\n\nimport \"testing\"\n\nfunc TestNothing(t *testing.T) {\n\t_ = 1 + 1\n}\n",
			},
		},
		{
			ID: "spec-hygiene", Tier: 1, Desc: "spec artifacts are well formed",
			Inputs: []string{".rla/specs"}, Run: checkSpecHygiene, Budget: 30 * time.Second,
			MinEvidence: map[string]int{"specs_parsed": 1},
			CanaryFiles: map[string]string{
				".rla/specs/mismatch.md": "---\nid: wrong-name\ntitle: Canary\nstatus: ratified\n---\n\n## SPEC-wrong-name-01 — Something\n",
			},
		},

		{
			ID: "spec-fidelity", Tier: 2, Desc: "every ratified requirement is implemented",
			Inputs: []string{".rla/specs", "cmd", "internal", "scripts"}, Run: checkSpecFidelity,
			Budget: time.Minute,
			// No evidence floor here: zero ratified requirements is a real and
			// legitimate state before checkpoint ①, and the check already
			// reports it as COULD NOT VERIFY rather than as a pass. A floor
			// would relabel "awaiting ratification" as "failed".
			CanaryFiles: map[string]string{
				".rla/specs/orphan.md": "---\nid: orphan\ntitle: Canary\nstatus: ratified\n---\n\n## SPEC-orphan-01 — Never implemented\n",
			},
		},
		{
			ID: "coverage-ratchet", Tier: 2, Desc: "coverage must not regress",
			Run: checkCoverage, Budget: 8 * time.Minute,
			MinEvidence: map[string]int{"functions_measured": 1},
			CanaryFiles: map[string]string{
				"go.mod":          "module canary\n\ngo 1.24\n",
				"x.go":            "package canary\n\nfunc Covered() int { return 1 }\n\nfunc Uncovered() int { return 2 }\n",
				"x_test.go":       "package canary\n\nimport \"testing\"\n\nfunc TestCovered(t *testing.T) {\n\tif Covered() != 1 {\n\t\tt.Fatal(\"unreachable\")\n\t}\n}\n",
				coverageFloorFile: "99.0\n",
			},
		},
		{
			ID: "doc-links", Tier: 2, Desc: "documentation links and anchors",
			Inputs: []string{"docs", "README.md", "CONTRIBUTING.md"}, Run: checkDocLinks,
			Budget: 2 * time.Minute, MinEvidence: map[string]int{"files_indexed": 1},
		},
		{
			ID: "licence-headers", Tier: 2, Desc: "AGPL headers present",
			Inputs: goInputs, Run: checkLicenceHeaders, Budget: 2 * time.Minute,
			CanaryFiles: map[string]string{
				"LICENSE_HEADER": "Canary\nCopyright (C) {{ .Year }} {{ .Holder }}\n",
				"cmd/bare.go":    "package main\n\nfunc main() {}\n",
			},
		},
		{
			ID: "licence-boundary", Tier: 2, Desc: "no AGPL code under mobile/",
			Inputs: []string{"mobile"}, Run: checkLicenceBoundary, Budget: time.Minute,
			CanaryFiles: map[string]string{
				"mobile/lib/leak.dart": "// GNU Affero General Public License\nvoid main() {}\n",
			},
		},

		{ID: "race", Tier: 3, Desc: "data race detector", Inputs: goInputs, Run: checkRace, Budget: 10 * time.Minute},
		{ID: "vulnerabilities", Tier: 3, Desc: "known CVEs in dependencies", Inputs: []string{"go.mod", "go.sum"}, Run: checkVulnerabilities, Budget: 8 * time.Minute},
	}
}

// judgementGates are the parts of the pipeline no script can decide. They are
// listed rather than silently omitted, because an unlisted obligation is
// indistinguishable from one that was met.
//
// Two rules bound them, and both are cost decisions before they are design
// ones. A judged pass reads the artifact this command writes and **runs no
// tools**: re-running what a script already decided costs an order of
// magnitude more and returns a verdict that is not reproducible. And a judged
// pass **does not block the iteration** — its output is a report, while the
// authority to stop the work stays with the exit codes.
var judgementGates = []string{
	"backward spec diff — behaviour in the code that no SPEC id covers (Tier 2)",
	"architectural intent — does the design still match PRINCIPLES.md (Tier 2)",
	"black-box exploration — adversarial probing without reading the source (Tier 3)",
	"spec correctness — checkpoint ①, is the plan itself right (human)",
	"interface test — checkpoint ②, does it actually work end to end (human)",
}

// ── runners ─────────────────────────────────────────────────────────────────

// ratifiedRequirementIDs is the requirement set a cache signature is bound to:
// ratifying a spec changes what the gates are checking, so it must invalidate
// their passes.
func ratifiedRequirementIDs(root string) []string {
	specs, err := loadSpecs(root)
	if err != nil {
		fmt.Printf("⚠️  spec artifacts unreadable: %v\n", err)
		return nil
	}
	var ids []string
	for _, s := range specs {
		if !s.Active() {
			continue
		}
		for _, r := range s.Requirements {
			ids = append(ids, r.ID)
		}
	}
	return ids
}

// runChecks runs one mode and writes the artifact describing it.
//
// The artifact is the point. Everything printed here is for the person
// watching; the file is for the reviewer that comes next, and it is what keeps
// the judged half from rediscovering the project on every pass.
//
// SPEC-deterministic-backbone-01
func runChecks(root string, mode runMode, noCache, verbose bool) (int, string) {
	specIDs := ratifiedRequirementIDs(root)
	cache := loadCache(root)

	fingerprint, fpErr := treeFingerprint(root, specIDs)
	if fpErr != nil {
		fmt.Printf("⚠️  tree fingerprint unavailable: %v\n", fpErr)
	}

	started := time.Now()
	worst, ran, cached := Pass, 0, 0
	var steps []stepRecord

	for _, tier := range mode.Tiers {
		checks := checksForTier(tier)
		if len(checks) == 0 {
			continue
		}
		fmt.Printf("\n\033[1mTier %d\033[0m — %s\n", tier, tierName(tier))

		for _, check := range checks {
			res := runOne(root, check, cache, specIDs, noCache)
			if res.Cached {
				cached++
			} else {
				ran++
			}
			report(res, verbose)
			steps = append(steps, newStepRecord(res))
			worst = worsen(worst, res.Status)
		}
	}

	elapsed := time.Since(started)
	if err := cache.save(); err != nil {
		fmt.Printf("⚠️  cache not saved: %v\n", err)
	}

	deferred := deferredChecks(mode)
	reportDeferred(mode, deferred)

	fmt.Printf("\n%s  %d gates run, %d cached, %s\n", worst.icon(), ran, cached, elapsed.Round(10*time.Millisecond))
	reportDrift(root, mode, elapsed)

	artifact := ""
	if fpErr == nil {
		var err error
		artifact, err = writeArtifact(root, runArtifact{
			Mode:          mode.Name,
			Fingerprint:   fingerprint,
			Generated:     time.Now().UTC().Format(time.RFC3339),
			Verdict:       worst.String(),
			Exit:          statusExit(worst),
			DurationMS:    elapsed.Milliseconds(),
			Steps:         steps,
			Deferred:      deferred,
			JudgementOwed: judgementGates,
		})
		if err != nil {
			fmt.Printf("⚠️  evidence artifact not written: %v\n", err)
		} else {
			fmt.Printf("    evidence: %s\n", artifact)
		}
	}

	return statusExit(worst), artifact
}

// runOne runs a single gate, or serves a cached pass that still holds up.
func runOne(root string, check Check, cache *gateCache, specIDs []string, noCache bool) Result {
	floor := baselineFloor(root, check)

	sig, sigErr := signature(root, check, specIDs)
	if sigErr == nil && !noCache && sig != "" {
		if entry, ok := cache.hit(check.ID, sig); ok {
			if guardsHold(check, entry.Evidence, entry.duration(), floor) {
				return Result{
					Name: check.ID, Tier: check.Tier, Status: Pass,
					Summary: entry.Summary, Evidence: entry.Evidence,
					Duration: entry.duration(), Cached: true,
				}
			}
			// The signature matched, so nothing this gate reads has changed —
			// which is precisely how an empty pass becomes permanent. A cached
			// green whose recorded numbers would fail the guards today is not
			// a green, and re-running is the only honest response.
			fmt.Printf("  ⚠️  %-18s cached pass rejected — its recorded evidence no longer clears the guards\n", check.ID)
		}
	}

	start := time.Now()
	res := check.Run(root)
	res.Duration = time.Since(start)
	res.Name = check.ID
	res.Tier = check.Tier
	res = applyGuards(check, res, floor)

	if res.Status == Pass && sigErr == nil && sig != "" {
		cache.store(check.ID, sig, res)
	}
	return res
}

// deferredChecks names everything this mode did not run.
//
// A check that was skipped and a check that passed look identical in a green
// summary. Naming the skipped ones is the difference between "the fast layer
// is green" and "the fast layer is green and says nothing about races".
func deferredChecks(mode runMode) []deferredRecord {
	inMode := map[int]bool{}
	for _, t := range mode.Tiers {
		inMode[t] = true
	}

	var out []deferredRecord
	for _, c := range registry() {
		if inMode[c.Tier] {
			continue
		}
		out = append(out, deferredRecord{
			ID:   c.ID,
			Tier: c.Tier,
			Reason: fmt.Sprintf("tier %d does not run in %s mode — %s",
				c.Tier, mode.Name, c.Desc),
		})
	}
	return out
}

func reportDeferred(mode runMode, deferred []deferredRecord) {
	if len(deferred) == 0 {
		return
	}
	fmt.Printf("\n\033[1mDeferred\033[0m — not run in %s mode, and therefore unproven:\n", mode.Name)
	for _, d := range deferred {
		fmt.Printf("  ·  %-18s (tier %d)\n", d.ID, d.Tier)
	}
}

// reportDrift compares this run against the committed measurement.
//
// The budget is a ceiling; this is the sensor. A layer that was measured at
// seconds and now takes half its budget did not get slower by accident —
// something moved into it, and the tier split stops paying for itself the day
// nobody notices.
func reportDrift(root string, mode runMode, elapsed time.Duration) {
	if mode.Budget <= 0 {
		return
	}
	fmt.Printf("    %s of a %s budget", elapsed.Round(10*time.Millisecond), mode.Budget)
	if baseline, err := loadTimings(root, timingsBaseline); err == nil {
		if line, drifted := driftAgainst(baseline, mode.Name, elapsed); line != "" {
			fmt.Printf(" · %s", line)
			if drifted {
				fmt.Print(" ⚠️  drifting — re-measure with `gate timings`")
			}
		}
	}
	fmt.Println()
}

func runVerify(root string, noCache, verbose bool) int {
	fmt.Println("\033[1mFull verification\033[0m — read-only sweep before checkpoint ②")

	code, artifact := runChecks(root, modeFull, noCache, verbose)

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
	if artifact != "" {
		fmt.Printf("\n    Hand the reviewer %s. It reads the evidence and runs no tools:\n", artifact)
		fmt.Println("    re-running what a script already decided costs an order of magnitude more")
		fmt.Println("    and answers a question that was already answered reproducibly.")
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

// reportEvidence is the reviewer's entry point: the artifact for the current
// working tree, or a refusal.
//
// Freshness is decided by fingerprint, not by timestamp. A report about a tree
// that has since moved on is not weak evidence — it is evidence about
// something else, and reading it is worse than reading nothing, because it
// arrives with the authority of a machine-generated file.
//
// SPEC-deterministic-backbone-04
func reportEvidence(root, mode string) int {
	fingerprint, err := treeFingerprint(root, ratifiedRequirementIDs(root))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot fingerprint the working tree: %v\n", err)
		return exitUnverified
	}

	artifact, fresh, err := readArtifact(root, mode, fingerprint)
	if err != nil || !fresh {
		fmt.Printf("⚠️   no fresh %s artifact for this working tree (fingerprint %s)\n", mode, short(fingerprint))
		for _, stale := range staleArtifacts(root, mode, fingerprint) {
			fmt.Printf("     stale: %s — describes a different tree\n", stale)
		}
		fmt.Printf("     run `gate %s` first; a judged pass on stale evidence proves nothing.\n", mode)
		return exitUnverified
	}

	fmt.Printf("\033[1m%s\033[0m — %s · fingerprint %s · %s\n",
		filepath.ToSlash(filepath.Join(stateDir, artifactName(mode, fingerprint))),
		artifact.Verdict, short(artifact.Fingerprint), artifact.Generated)

	for _, s := range artifact.Steps {
		evidence := ""
		if len(s.Evidence) > 0 {
			parts := make([]string, 0, len(s.Evidence))
			for _, k := range sortedKeys(s.Evidence) {
				parts = append(parts, fmt.Sprintf("%s=%d", k, s.Evidence[k]))
			}
			evidence = "  \033[2m" + strings.Join(parts, " ") + "\033[0m"
		}
		fmt.Printf("  %-18s %-18s %s%s\n", s.ID, s.Verdict, s.Summary, evidence)
	}

	if len(artifact.Deferred) > 0 {
		names := make([]string, 0, len(artifact.Deferred))
		for _, d := range artifact.Deferred {
			names = append(names, d.ID)
		}
		fmt.Printf("\n  deferred (unproven): %s\n", strings.Join(names, ", "))
	}

	fmt.Println("\n  Read this; do not re-run it. The verdict above is an exit code, which is")
	fmt.Println("  reproducible in a way a second opinion about it would not be.")
	return artifact.Exit
}

// staleArtifacts names the reports left over from earlier trees, so a reader
// who was about to open one is told why not to.
func staleArtifacts(root, mode, fingerprint string) []string {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(stateDir)))
	if err != nil {
		return nil
	}
	prefix := fmt.Sprintf("verify-%s-", mode)
	current := artifactName(mode, fingerprint)

	var out []string
	for _, e := range entries {
		if name := e.Name(); strings.HasPrefix(name, prefix) && name != current {
			out = append(out, filepath.ToSlash(filepath.Join(stateDir, name)))
		}
	}
	return out
}

// runTimings measures. Every budget and every tier boundary in this file is
// downstream of these numbers, and the one ordering mistake that cannot be
// recovered from is setting them the other way round — deciding what a tier
// should cost and then never checking what it does.
//
// SPEC-deterministic-backbone-05
func runTimings(root string, record bool) int {
	fmt.Println("\033[1mTimings\033[0m — measured wall clock, cache bypassed")

	measured := timings{
		Measured: time.Now().UTC().Format(time.RFC3339),
		Modes:    map[string]int64{},
		Checks:   map[string]timingRecord{},
	}
	byCheck := map[string]time.Duration{}

	for _, check := range registry() {
		start := time.Now()
		res := check.Run(root)
		elapsed := time.Since(start)

		byCheck[check.ID] = elapsed
		measured.Total += elapsed.Milliseconds()
		measured.Checks[check.ID] = timingRecord{Milliseconds: elapsed.Milliseconds(), Verdict: res.Status.String()}
		if res.Status == Unverified {
			// A check that could not run contributes almost nothing to the
			// clock. Recording that silently would make the first complete run
			// on a full toolchain look like drift.
			measured.Unmeasured = append(measured.Unmeasured, check.ID)
		}

		note := ""
		if check.Budget > 0 && elapsed > check.Budget {
			note = fmt.Sprintf("  ⚠️  over its %s budget", check.Budget)
		}
		fmt.Printf("  %-18s %8s   %s%s\n", check.ID, elapsed.Round(time.Millisecond), res.Status, note)
	}

	for _, mode := range []runMode{modeFast, modeFull} {
		total := time.Duration(0)
		for _, tier := range mode.Tiers {
			for _, c := range checksForTier(tier) {
				total += byCheck[c.ID]
			}
		}
		measured.Modes[mode.Name] = total.Milliseconds()
		fmt.Printf("\n  %-6s %8s of a %s budget (%.0f%%)\n",
			mode.Name, total.Round(time.Millisecond), mode.Budget,
			float64(total)/float64(mode.Budget)*100)
	}

	if err := saveTimings(root, filepath.ToSlash(filepath.Join(stateDir, "tool-timings.json")), measured); err != nil {
		fmt.Printf("⚠️  measurement not saved: %v\n", err)
		return exitUnverified
	}

	if !record {
		fmt.Printf("\n  Written to %s/tool-timings.json. Pass -record to make it the committed\n", stateDir)
		fmt.Printf("  baseline in %s — that is a deliberate act, like lowering the coverage floor.\n", timingsBaseline)
		return exitOK
	}
	if err := saveTimings(root, timingsBaseline, measured); err != nil {
		fmt.Printf("⚠️  baseline not written: %v\n", err)
		return exitUnverified
	}
	fmt.Printf("\n  Baseline recorded in %s — commit it; drift is measured against it.\n", timingsBaseline)
	return exitOK
}

// runCanary deliberately takes no root: every canary runs against a freshly
// planted temporary tree, never against the repository.
// canaryVerdict maps a gate's result on deliberately broken input to what the
// canary may conclude about that gate.
//
// A gate that FAILED saw the breakage — that is the canary passing. A gate that
// PASSED is blind, which is the finding this whole command exists to surface.
//
// A gate that could not run has proved nothing in either direction. Reporting
// that as blindness would be a misdiagnosis: it sends the reader hunting for a
// defect in a gate that is very likely fine, while the real cause — usually an
// uninstalled tool — goes unnamed. It is still not a pass ([P3](fail loud)),
// so it degrades the run to COULD NOT VERIFY rather than clearing it.
func canaryVerdict(onBrokenInput Status) Status {
	switch onBrokenInput {
	case Fail:
		return Pass
	case Unverified:
		return Unverified
	default:
		return Fail
	}
}

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

		switch verdict := canaryVerdict(res.Status); verdict {
		case Pass:
			fmt.Printf("  🟢 %-18s detected planted breakage\n", check.ID)
		case Unverified:
			fmt.Printf("  ⚠️  %-18s canary could not run: %s\n", check.ID, res.Summary)
			worst = worsen(worst, verdict)
		default:
			fmt.Printf("  🔴 %-18s reported %s on broken input — THIS GATE IS BLIND\n", check.ID, res.Status)
			worst = worsen(worst, verdict)
		}
	}

	fmt.Println()
	switch worst {
	case Pass:
		fmt.Println("🟢  every gate with a canary can still detect breakage")
	case Unverified:
		fmt.Println("⚠️   some canaries did not run — those gates' green is worth less than it looks")
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
