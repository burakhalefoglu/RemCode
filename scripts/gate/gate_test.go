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

//gate:spec-fixtures — SPEC ids below are test data, not citations.

// Tests for the gate itself.
//
// A verification tool nobody verifies is the exact fake-green this system
// exists to catch. `gate canary` proves the gates react to planted breakage;
// these tests pin the logic underneath them.

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

// ── status semantics ────────────────────────────────────────────────────────

func TestUnverifiedIsNotAPass(t *testing.T) {
	// The whole fail-loud principle rests on this distinction. If Unverified
	// ever collapses into Pass, a gate that could not run starts reporting
	// success — which is the failure mode the system is built to prevent.
	if Unverified == Pass {
		t.Fatal("Unverified must be distinct from Pass")
	}
	if statusExit(Unverified) == exitOK {
		t.Errorf("statusExit(Unverified) = %d, must not be the success code", statusExit(Unverified))
	}
	if got := Unverified.String(); got != "COULD NOT VERIFY" {
		t.Errorf("Unverified.String() = %q", got)
	}
}

func TestCanaryVerdictSeparatesBlindFromUnrunnable(t *testing.T) {
	// A gate that could not run and a gate that ran and saw nothing are very
	// different findings. Collapsing them sends the reader hunting for a defect
	// in a gate that is fine, while the real cause — an uninstalled tool — goes
	// unnamed. Neither is a pass; only one is an accusation.
	cases := []struct {
		name          string
		onBrokenInput Status
		want          Status
	}{
		{"caught the breakage", Fail, Pass},
		{"ran and saw nothing — blind", Pass, Fail},
		{"could not run — proves nothing", Unverified, Unverified},
	}
	for _, c := range cases {
		if got := canaryVerdict(c.onBrokenInput); got != c.want {
			t.Errorf("%s: canaryVerdict(%v) = %v, want %v", c.name, c.onBrokenInput, got, c.want)
		}
	}
	if canaryVerdict(Unverified) == Pass {
		t.Error("an unrunnable canary must never clear the run")
	}
}

func TestWorsenNeverImproves(t *testing.T) {
	cases := []struct {
		current, next, want Status
	}{
		{Pass, Pass, Pass},
		{Pass, Unverified, Unverified},
		{Pass, Fail, Fail},
		{Unverified, Pass, Unverified},
		{Unverified, Fail, Fail},
		{Fail, Pass, Fail},
		{Fail, Unverified, Fail},
	}
	for _, c := range cases {
		if got := worsen(c.current, c.next); got != c.want {
			t.Errorf("worsen(%v, %v) = %v, want %v", c.current, c.next, got, c.want)
		}
	}
}

// ── spec parsing ────────────────────────────────────────────────────────────

func TestParseSpecReadsFrontMatterAndRequirements(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/demo.md": `---
id: demo
title: Demo feature
status: ratified
phase: P9
---

Prose that mentions nothing.

## SPEC-demo-01 — First requirement

Body.

## SPEC-demo-02 — Second requirement
`,
	})

	specs, err := loadSpecs(dir)
	if err != nil {
		t.Fatalf("loadSpecs: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}

	s := specs[0]
	if s.ID != "demo" || s.Title != "Demo feature" || s.Phase != "P9" {
		t.Errorf("front matter mis-parsed: %+v", s)
	}
	if !s.Active() || s.Draft() {
		t.Errorf("status %q should be active and not draft", s.Status)
	}
	if len(s.Requirements) != 2 {
		t.Fatalf("got %d requirements, want 2", len(s.Requirements))
	}
	if s.Requirements[0].ID != "SPEC-demo-01" || s.Requirements[0].Title != "First requirement" {
		t.Errorf("requirement 0 mis-parsed: %+v", s.Requirements[0])
	}
}

func TestParseSpecRejectsUnknownStatus(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/bad.md": "---\nid: bad\ntitle: T\nstatus: probably-fine\n---\n\n## SPEC-bad-01 — X\n",
	})

	if _, err := loadSpecs(dir); err == nil {
		t.Fatal("expected an error for an unknown status")
	}
}

func TestTemplateIsNotTreatedAsASpec(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/_TEMPLATE.md": "---\nid: <slug>\nstatus: draft\n---\n",
	})

	specs, err := loadSpecs(dir)
	if err != nil {
		t.Fatalf("loadSpecs: %v", err)
	}
	if len(specs) != 0 {
		t.Errorf("underscore-prefixed files must be skipped, got %d specs", len(specs))
	}
}

func TestDraftRequirementsAreNotEnforced(t *testing.T) {
	// A draft is a proposal. Holding code to it before a human agrees would
	// invert checkpoint ① — the plan must be ratified before it binds.
	dir := writeFiles(t, map[string]string{
		".rla/specs/draft.md": "---\nid: draft\ntitle: T\nstatus: draft\n---\n\n## SPEC-draft-01 — Not yet agreed\n",
		".rla/specs/live.md":  "---\nid: live\ntitle: T\nstatus: ratified\n---\n\n## SPEC-live-01 — Agreed\n",
		"internal/impl/x.go":  "package impl\n\n// SPEC-live-01: done.\n",
	})

	res := checkSpecFidelity(dir)
	if res.Status != Pass {
		t.Errorf("status = %v (%s), want Pass; findings: %v", res.Status, res.Summary, res.Findings)
	}
}

func TestUnimplementedRatifiedRequirementFails(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/live.md": "---\nid: live\ntitle: T\nstatus: ratified\n---\n\n## SPEC-live-01 — Never built\n",
		"internal/impl/x.go": "package impl\n",
	})

	res := checkSpecFidelity(dir)
	if res.Status != Fail {
		t.Fatalf("status = %v, want Fail", res.Status)
	}
	if !strings.Contains(strings.Join(res.Findings, "\n"), "SPEC-live-01") {
		t.Errorf("findings should name the uncovered requirement: %v", res.Findings)
	}
}

func TestDanglingCitationFails(t *testing.T) {
	// Code citing an id no spec declares means either a typo or a requirement
	// that was deleted while its marker stayed behind. Both are drift.
	dir := writeFiles(t, map[string]string{
		".rla/specs/live.md": "---\nid: live\ntitle: T\nstatus: ratified\n---\n\n## SPEC-live-01 — Built\n",
		"internal/impl/x.go": "package impl\n\n// SPEC-live-01 and SPEC-live-99 are cited here.\n",
	})

	res := checkSpecFidelity(dir)
	if res.Status != Fail {
		t.Fatalf("status = %v, want Fail", res.Status)
	}
	if !strings.Contains(strings.Join(res.Findings, "\n"), "SPEC-live-99") {
		t.Errorf("findings should name the dangling citation: %v", res.Findings)
	}
}

func TestSpecHygieneCatchesIDFileNameMismatch(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/actual-name.md": "---\nid: different\ntitle: T\nstatus: ratified\n---\n\n## SPEC-different-01 — X\n",
	})

	res := checkSpecHygiene(dir)
	if res.Status != Fail {
		t.Errorf("status = %v, want Fail for an id/filename mismatch", res.Status)
	}
}

func TestSpecHygieneCatchesForeignRequirementID(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		".rla/specs/mine.md": "---\nid: mine\ntitle: T\nstatus: ratified\n---\n\n## SPEC-yours-01 — Belongs elsewhere\n",
	})

	res := checkSpecHygiene(dir)
	if res.Status != Fail {
		t.Errorf("status = %v, want Fail for a requirement id from another spec", res.Status)
	}
}

// ── conformance checks ──────────────────────────────────────────────────────

func TestZeroTouchAIRejectsProviderImportInRelay(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"internal/server/relay.go": "package server\n\nimport _ \"example.com/x/internal/models\"\n",
	})

	res := checkZeroTouchAI(dir)
	if res.Status != Fail {
		t.Fatalf("status = %v, want Fail — the relay must not reach provider code", res.Status)
	}
}

func TestZeroTouchAIRejectsCryptoImportInRelay(t *testing.T) {
	// ADR-004: the relay holds no key material, ever.
	dir := writeFiles(t, map[string]string{
		"internal/server/relay.go": "package server\n\nimport _ \"example.com/x/internal/crypto\"\n",
	})

	if res := checkZeroTouchAI(dir); res.Status != Fail {
		t.Errorf("status = %v, want Fail", res.Status)
	}
}

func TestZeroTouchAIAllowsProviderImportOutsideRelay(t *testing.T) {
	// The CLI agent is *supposed* to talk to providers. The rule is about
	// location, not about the import existing anywhere in the tree.
	dir := writeFiles(t, map[string]string{
		"internal/agent/loop.go": "package agent\n\nimport _ \"example.com/x/internal/models\"\n",
	})

	if res := checkZeroTouchAI(dir); res.Status != Pass {
		t.Errorf("status = %v, want Pass; findings: %v", res.Status, res.Findings)
	}
}

func TestSecretLoggingCatchesCredentialInLogCall(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"x.go": "package x\n\nimport \"log/slog\"\n\nfunc F(apiKey string) { slog.Info(\"hi\", apiKey) }\n",
	})

	res := checkSecretLogging(dir)
	if res.Status != Fail {
		t.Fatalf("status = %v, want Fail", res.Status)
	}
}

func TestSecretLoggingAllowsPublicKeyMaterial(t *testing.T) {
	// Public keys and fingerprints are safe to log, and the pairing flow needs
	// to. A check that cannot tell them apart would be turned off within a week.
	dir := writeFiles(t, map[string]string{
		"x.go": "package x\n\nimport \"log/slog\"\n\nfunc F(publicKey, fingerprint string) { slog.Info(\"paired\", publicKey, fingerprint) }\n",
	})

	if res := checkSecretLogging(dir); res.Status != Pass {
		t.Errorf("status = %v, want Pass; findings: %v", res.Status, res.Findings)
	}
}

func TestSecretLoggingHonoursOptOut(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"x.go": "//gate:allow-secret-log\n\npackage x\n\nimport \"log/slog\"\n\nfunc F(token string) { slog.Info(\"hi\", token) }\n",
	})

	if res := checkSecretLogging(dir); res.Status != Pass {
		t.Errorf("status = %v, want Pass when the file opts out", res.Status)
	}
}

func TestFakeGreenCatchesAssertionlessTest(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestHollow(t *testing.T) { _ = 1 }\n",
	})

	res := checkFakeGreen(dir)
	if res.Status != Fail {
		t.Fatalf("status = %v, want Fail", res.Status)
	}
	if !strings.Contains(strings.Join(res.Findings, "\n"), "TestHollow") {
		t.Errorf("findings should name the hollow test: %v", res.Findings)
	}
}

func TestFakeGreenAcceptsAssertionInSubtest(t *testing.T) {
	// Table-driven tests assert inside a closure. Flagging those would make
	// the check useless against the dominant Go test style.
	dir := writeFiles(t, map[string]string{
		"x_test.go": `package x

import "testing"

func TestTable(t *testing.T) {
	t.Run("case", func(t *testing.T) {
		if 1 != 1 {
			t.Errorf("impossible")
		}
	})
}
`,
	})

	if res := checkFakeGreen(dir); res.Status != Pass {
		t.Errorf("status = %v, want Pass; findings: %v", res.Status, res.Findings)
	}
}

func TestLicenceBoundaryRejectsAGPLUnderMobile(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"mobile/lib/leak.dart": "// GNU Affero General Public License\nvoid main() {}\n",
	})

	if res := checkLicenceBoundary(dir); res.Status != Fail {
		t.Errorf("status = %v, want Fail — ADR-002 makes the boundary one-way", res.Status)
	}
}

func TestLicenceBoundaryPassesWithoutMobile(t *testing.T) {
	if res := checkLicenceBoundary(t.TempDir()); res.Status != Pass {
		t.Errorf("status = %v, want Pass when mobile/ does not exist", res.Status)
	}
}

// ── coverage ratchet ────────────────────────────────────────────────────────

func TestParseTotalCoverage(t *testing.T) {
	out := "github.com/x/y.go:12:\tF\t100.0%\ntotal:\t\t\t(statements)\t42.7%\n"

	got, ok := parseTotalCoverage(out)
	if !ok {
		t.Fatal("failed to parse a well-formed coverage summary")
	}
	if got != 42.7 {
		t.Errorf("got %v, want 42.7", got)
	}
}

func TestParseTotalCoverageReportsFailureOnGarbage(t *testing.T) {
	// Silently returning 0 would let the ratchet read a parse failure as a
	// catastrophic coverage drop, or worse, as a pass against a zero floor.
	if _, ok := parseTotalCoverage("no total line here"); ok {
		t.Error("expected ok=false when there is no total line")
	}
}

func TestReadCoverageFloorDefaultsToZero(t *testing.T) {
	if got := readCoverageFloor(t.TempDir()); got != 0 {
		t.Errorf("got %v, want 0 when the floor file is absent", got)
	}
}

func TestReadCoverageFloor(t *testing.T) {
	dir := writeFiles(t, map[string]string{coverageFloorFile: "37.5\n"})

	if got := readCoverageFloor(dir); got != 37.5 {
		t.Errorf("got %v, want 37.5", got)
	}
}

// ── cache ───────────────────────────────────────────────────────────────────

func TestSignatureChangesWithContent(t *testing.T) {
	check := Check{ID: "demo", Inputs: []string{"internal"}}

	dir := writeFiles(t, map[string]string{"internal/x.go": "package x\n"})
	before, err := signature(dir, check, nil)
	if err != nil {
		t.Fatalf("signature: %v", err)
	}

	if writeErr := os.WriteFile(filepath.Join(dir, "internal", "x.go"), []byte("package x // edited\n"), 0o600); writeErr != nil {
		t.Fatalf("rewrite: %v", writeErr)
	}
	after, err := signature(dir, check, nil)
	if err != nil {
		t.Fatalf("signature: %v", err)
	}

	if before == after {
		t.Error("signature must change when an input file changes")
	}
}

func TestSignatureChangesWithSpecSet(t *testing.T) {
	check := Check{ID: "demo", Inputs: []string{"internal"}}
	dir := writeFiles(t, map[string]string{"internal/x.go": "package x\n"})

	a, _ := signature(dir, check, []string{"SPEC-a-01"})
	b, _ := signature(dir, check, []string{"SPEC-a-01", "SPEC-a-02"})

	if a == b {
		t.Error("signature must change when the ratified requirement set changes")
	}
}

func TestSignatureIsOrderIndependentForSpecs(t *testing.T) {
	check := Check{ID: "demo", Inputs: []string{"internal"}}
	dir := writeFiles(t, map[string]string{"internal/x.go": "package x\n"})

	a, _ := signature(dir, check, []string{"SPEC-a-01", "SPEC-a-02"})
	b, _ := signature(dir, check, []string{"SPEC-a-02", "SPEC-a-01"})

	if a != b {
		t.Error("spec ordering must not affect the signature")
	}
}

func TestEmptyInputsOptOutOfCaching(t *testing.T) {
	sig, err := signature(t.TempDir(), Check{ID: "demo"}, nil)
	if err != nil {
		t.Fatalf("signature: %v", err)
	}
	if sig != "" {
		t.Errorf("got %q, want an empty signature so the gate is never cached", sig)
	}
}

func TestCacheOnlyMatchesItsOwnSignature(t *testing.T) {
	dir := t.TempDir()
	c := loadCache(dir)

	c.store("demo", "sig-a", Result{Summary: "all good"})

	if _, ok := c.hit("demo", "sig-b"); ok {
		t.Error("a stale signature must not hit")
	}
	entry, ok := c.hit("demo", "sig-a")
	if !ok || entry.Summary != "all good" {
		t.Errorf("hit = (%q, %v), want (\"all good\", true)", entry.Summary, ok)
	}
}

func TestCacheRoundTripsEvidenceThroughDisk(t *testing.T) {
	// The evidence has to survive the round trip, not just the summary: it is
	// what a later hit is re-judged against, and an entry that loses its
	// counts becomes a pass nobody can question.
	dir := t.TempDir()

	first := loadCache(dir)
	first.store("demo", "sig", Result{
		Summary:  "118 tests",
		Evidence: Evidence{"tests_run": 118},
		Duration: 1500 * time.Millisecond,
	})
	if err := first.save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	entry, ok := loadCache(dir).hit("demo", "sig")
	if !ok {
		t.Fatal("a saved entry should be readable by a fresh cache")
	}
	if entry.Evidence["tests_run"] != 118 {
		t.Errorf("tests_run = %d, want 118 — evidence must survive the cache", entry.Evidence["tests_run"])
	}
	if entry.duration() != 1500*time.Millisecond {
		t.Errorf("duration = %s, want 1.5s", entry.duration())
	}
}

func TestCorruptCacheIsTreatedAsEmpty(t *testing.T) {
	// A corrupt cache read as "everything passed" would be the worst possible
	// failure mode for a tool whose job is proving things.
	dir := writeFiles(t, map[string]string{cachePath: "{not json"})

	if _, ok := loadCache(dir).hit("demo", "sig"); ok {
		t.Error("a corrupt cache must never report a hit")
	}
}

// ── guards: the part an exit code cannot say ────────────────────────────────

func TestEmptyRunGuardFailsAGateThatExaminedNothing(t *testing.T) {
	// Both silent greens this project has seen were of this shape: a command
	// that exited 0 having looked at nothing. The exit code was honest; the
	// question it answered was not the one anybody meant to ask.
	check := Check{ID: "scanner", MinEvidence: map[string]int{"files_scanned": 1}}
	res := applyGuards(check, Result{Status: Pass, Evidence: Evidence{"files_scanned": 0}}, 0)

	if res.Status != Fail {
		t.Errorf("status = %v, want FAILED — a scanner that read nothing must not pass", res.Status)
	}
	if len(res.Findings) == 0 {
		t.Error("the guard must say what it saw, or the red is unactionable")
	}
}

func TestUnreportedEvidenceIsUnverifiedNotPassed(t *testing.T) {
	// A gate that declares a floor and then reports no count has not passed
	// the guard — it has evaded it.
	check := Check{ID: "scanner", MinEvidence: map[string]int{"files_scanned": 1}}
	res := applyGuards(check, Result{Status: Pass}, 0)

	if res.Status != Unverified {
		t.Errorf("status = %v, want COULD NOT VERIFY", res.Status)
	}
}

func TestBaselineGuardCatchesErosionNotJustCollapse(t *testing.T) {
	// A suite falling from 118 tests to 0 is caught by the floor. Falling to
	// 90 is caught by nothing else at all: every remaining test passes, the
	// command exits 0, and 28 tests left without a single red.
	check := Check{
		ID: "tests", BaselineKey: "tests_run", BaselineFile: testBaselineFile,
		MinEvidence: map[string]int{"tests_run": 1},
	}

	eroded := applyGuards(check, Result{Status: Pass, Evidence: Evidence{"tests_run": 90}}, 118)
	if eroded.Status != Fail {
		t.Errorf("status = %v, want FAILED when the suite shrank", eroded.Status)
	}

	grown := applyGuards(check, Result{Status: Pass, Evidence: Evidence{"tests_run": 130}}, 118)
	if grown.Status != Pass {
		t.Errorf("status = %v, want PASSED — the baseline is a ratchet, not a target", grown.Status)
	}
}

func TestMissingBaselineIsUnverified(t *testing.T) {
	// An unmeasurable guard reports that it could not measure. Treating a
	// missing baseline as "no erosion" would make deleting the file the
	// cheapest way to clear the gate.
	check := Check{ID: "tests", BaselineKey: "tests_run", BaselineFile: testBaselineFile}
	res := applyGuards(check, Result{Status: Pass, Evidence: Evidence{"tests_run": 10}}, -1)

	if res.Status != Unverified {
		t.Errorf("status = %v, want COULD NOT VERIFY when no baseline is recorded", res.Status)
	}
}

func TestBaselineFloorRejectsUnreadableFiles(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{"recorded", map[string]string{testBaselineFile: "118\n"}, 118},
		{"absent", nil, -1},
		{"garbage", map[string]string{testBaselineFile: "lots\n"}, -1},
		{"negative", map[string]string{testBaselineFile: "-5\n"}, -1},
	}
	check := Check{ID: "tests", BaselineKey: "tests_run", BaselineFile: testBaselineFile}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := baselineFloor(writeFiles(t, c.files), check); got != c.want {
				t.Errorf("baselineFloor = %d, want %d", got, c.want)
			}
		})
	}
}

func TestStepBudgetIsARedNotAWarning(t *testing.T) {
	// A step that outgrew its budget has changed which tier it belongs in, and
	// the tier split is the only reason the inner loop is affordable.
	check := Check{ID: "slow", Budget: time.Second}
	res := applyGuards(check, Result{Status: Pass, Duration: 2 * time.Second}, 0)

	if res.Status != Fail {
		t.Errorf("status = %v, want FAILED when a step exceeds its budget", res.Status)
	}
}

func TestGuardsNeverClearARed(t *testing.T) {
	// Guards are allowed to accuse, never to absolve. A passing count must not
	// rehabilitate a gate that already found something.
	check := Check{ID: "tests", MinEvidence: map[string]int{"tests_run": 1}, Budget: time.Hour}
	res := applyGuards(check, Result{Status: Fail, Evidence: Evidence{"tests_run": 500}}, 1)

	if res.Status != Fail {
		t.Errorf("status = %v, want FAILED — guards may only worsen a verdict", res.Status)
	}
}

func TestCachedPassIsRejectedWhenItsEvidenceWouldFailToday(t *testing.T) {
	// This is where a fake green becomes permanent. The files that make a
	// suite run zero tests do not change, so the signature keeps matching and
	// the empty pass is served for ever. Judging the recorded counts on every
	// hit is the only thing that expires it.
	check := Check{ID: "tests", MinEvidence: map[string]int{"tests_run": 1}}

	if guardsHold(check, Evidence{"tests_run": 0}, time.Second, 0) {
		t.Error("a cached pass that ran zero tests must not be reusable")
	}
	if !guardsHold(check, Evidence{"tests_run": 118}, time.Second, 0) {
		t.Error("a cached pass with real counts should still be usable")
	}
}

// ── counting the suite ──────────────────────────────────────────────────────

func TestParseGoTestJSONSeparatesRunFromSkipped(t *testing.T) {
	// Skips are counted apart from runs because a suite where a third of the
	// tests quietly began skipping exits 0 and reads exactly like a healthy
	// one.
	stream := strings.Join([]string{
		`{"Action":"run","Package":"p","Test":"TestA"}`,
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
		`{"Action":"output","Package":"p","Test":"TestB","Output":"    boom\n"}`,
		`{"Action":"fail","Package":"p","Test":"TestB"}`,
		`{"Action":"skip","Package":"p","Test":"TestC"}`,
		`{"Action":"pass","Package":"p"}`,
		`not json at all`,
	}, "\n")

	tally := parseGoTestJSON(stream)
	if tally.run != 2 {
		t.Errorf("run = %d, want 2 (pass + fail, skips excluded)", tally.run)
	}
	if tally.skipped != 1 {
		t.Errorf("skipped = %d, want 1", tally.skipped)
	}
	if tally.packages != 1 {
		t.Errorf("packages = %d, want 1", tally.packages)
	}
	if len(tally.failures) == 0 || !strings.Contains(strings.Join(tally.failures, "\n"), "boom") {
		t.Errorf("failures = %v, want the failing test's own output", tally.failures)
	}
}

func TestParseGoTestJSONCountsEachTestOnce(t *testing.T) {
	stream := strings.Join([]string{
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
		`{"Action":"pass","Package":"p","Test":"TestA"}`,
	}, "\n")

	if got := parseGoTestJSON(stream).run; got != 1 {
		t.Errorf("run = %d, want 1 — a repeated event is not a second test", got)
	}
}

func TestCountCoveredFunctions(t *testing.T) {
	// A percentage says nothing about how much was weighed: 100% of two
	// functions and 100% of two hundred are the same number, not the same
	// claim.
	out := "rla/main.go:12:\tmain\t100.0%\nrla/x.go:4:\tF\t0.0%\ntotal:\t(statements)\t61.4%\n"
	if got := countCoveredFunctions(out); got != 2 {
		t.Errorf("countCoveredFunctions = %d, want 2", got)
	}
}

func TestLeadingIntFallsBackToZero(t *testing.T) {
	cases := map[string]int{
		"42 files, 10 anchors indexed, 0 problems": 42,
		"no numbers here":                          0,
		"":                                         0,
	}
	for in, want := range cases {
		if got := leadingInt(in); got != want {
			t.Errorf("leadingInt(%q) = %d, want %d", in, got, want)
		}
	}
}

// ── run modes, deferral and the artifact ────────────────────────────────────

func TestFastModeDefersTheExpensiveTiers(t *testing.T) {
	// The fast layer's value is that it is cheap enough to run constantly.
	// Anything expensive that creeps into it is paid for a hundred times a day.
	fastTiers := map[int]bool{}
	for _, tier := range modeFast.Tiers {
		fastTiers[tier] = true
	}
	if fastTiers[2] || fastTiers[3] {
		t.Error("fast mode must not include tiers 2 or 3")
	}

	deferred := deferredChecks(modeFast)
	if len(deferred) == 0 {
		t.Fatal("fast mode defers work; an empty list would claim otherwise")
	}
	for _, d := range deferred {
		if d.ID == "" || d.Reason == "" {
			t.Errorf("deferred entry %+v must name the check and why it was skipped", d)
		}
		if fastTiers[d.Tier] {
			t.Errorf("%s runs in fast mode and must not be listed as deferred", d.ID)
		}
	}
}

func TestFullModeDefersNothing(t *testing.T) {
	if deferred := deferredChecks(modeFull); len(deferred) != 0 {
		t.Errorf("full mode deferred %d checks — it is what a readiness verdict rests on", len(deferred))
	}
}

func TestArtifactIsWrittenUnderItsFingerprint(t *testing.T) {
	dir := t.TempDir()
	written, err := writeArtifact(dir, runArtifact{
		Mode: "full", Fingerprint: "abcdef0123456789", Verdict: "PASSED", Exit: 0,
		Steps: []stepRecord{{ID: "tests", Verdict: "PASSED", Evidence: Evidence{"tests_run": 118}}},
	})
	if err != nil {
		t.Fatalf("writeArtifact: %v", err)
	}
	if !strings.Contains(written, "abcdef012345") {
		t.Errorf("path %q should carry the fingerprint", written)
	}

	got, fresh, err := readArtifact(dir, "full", "abcdef0123456789")
	if err != nil || !fresh {
		t.Fatalf("readArtifact: fresh = %v, err = %v", fresh, err)
	}
	if got.Steps[0].Evidence["tests_run"] != 118 {
		t.Error("the artifact must carry the counts, not only the verdicts")
	}
}

func TestArtifactForAnotherTreeIsNotFound(t *testing.T) {
	// Freshness is decided by fingerprint, never by timestamp: a report about
	// a tree that has moved on is evidence about something else.
	dir := t.TempDir()
	if _, err := writeArtifact(dir, runArtifact{Mode: "full", Fingerprint: "oldoldoldold"}); err != nil {
		t.Fatalf("writeArtifact: %v", err)
	}

	if _, _, err := readArtifact(dir, "full", "newnewnewnew"); err == nil {
		t.Error("an artifact from a different tree must not be readable as the current one")
	}
	stale := staleArtifacts(dir, "full", "newnewnewnew")
	if len(stale) != 1 {
		t.Fatalf("staleArtifacts = %v, want the old report named so a reader knows not to open it", stale)
	}
}

func TestArtifactIsMachineReadable(t *testing.T) {
	// The artifact exists so a reviewing model reads what ran instead of
	// running it again. If it stops being parseable, that reader silently goes
	// back to re-running everything.
	dir := t.TempDir()
	written, err := writeArtifact(dir, runArtifact{
		Mode: "fast", Fingerprint: "0123456789ab", Verdict: "FAILED", Exit: 1,
		Steps:    []stepRecord{{ID: "tests", Tier: 1, Verdict: "FAILED", Exit: 1}},
		Deferred: []deferredRecord{{ID: "race", Tier: 3, Reason: "tier 3 does not run in fast mode"}},
	})
	if err != nil {
		t.Fatalf("writeArtifact: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(written))) //nolint:gosec // test fixture
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("the artifact must be valid JSON: %v", err)
	}
	for _, key := range []string{"mode", "fingerprint", "verdict", "exit", "steps", "deferred", "judgement_owed"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("the artifact is missing %q", key)
		}
	}
}

func TestDriftIsMeasuredAgainstTheRecordedBaseline(t *testing.T) {
	// The budget is a ceiling; this is the sensor. A layer measured in seconds
	// that now takes half its budget did not get slower by accident.
	baseline := timings{Modes: map[string]int64{"fast": 1000}}

	if _, drifted := driftAgainst(baseline, "fast", 1100*time.Millisecond); drifted {
		t.Error("10% above baseline is noise, not drift")
	}
	if _, drifted := driftAgainst(baseline, "fast", 3*time.Second); !drifted {
		t.Error("3× the baseline must be reported as drift")
	}
	if line, _ := driftAgainst(baseline, "full", time.Second); line != "" {
		t.Errorf("got %q, want silence when nothing was measured for that mode", line)
	}
}

func TestTimingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := timings{
		Measured: "2026-07-28T00:00:00Z",
		Total:    4200,
		Modes:    map[string]int64{"fast": 1200, "full": 4200},
		Checks:   map[string]timingRecord{"tests": {Milliseconds: 900, Verdict: "PASSED"}},
	}
	if err := saveTimings(dir, timingsBaseline, want); err != nil {
		t.Fatalf("saveTimings: %v", err)
	}
	got, err := loadTimings(dir, timingsBaseline)
	if err != nil {
		t.Fatalf("loadTimings: %v", err)
	}
	if got.Modes["fast"] != 1200 || got.Checks["tests"].Milliseconds != 900 {
		t.Errorf("round trip lost data: %+v", got)
	}
}

// ── registry integrity ──────────────────────────────────────────────────────

func TestEveryCheckIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range registry() {
		switch {
		case c.ID == "":
			t.Error("a check has no id")
		case seen[c.ID]:
			t.Errorf("duplicate check id %q", c.ID)
		case c.Run == nil:
			t.Errorf("%s has no Run function", c.ID)
		case c.Tier < 0 || c.Tier > 4:
			t.Errorf("%s has tier %d, outside 0–4", c.ID, c.Tier)
		case c.Budget <= 0:
			// An unbounded step is how a tier stops meaning anything: it keeps
			// its name while its cost migrates into a different one.
			t.Errorf("%s declares no budget — every step must state its ceiling", c.ID)
		case c.BaselineKey != "" && c.BaselineFile == "":
			t.Errorf("%s ratchets %q against no file", c.ID, c.BaselineKey)
		}
		seen[c.ID] = true
	}
	if len(seen) == 0 {
		t.Fatal("the registry is empty")
	}
}

// ── argument parsing ────────────────────────────────────────────────────────

func TestSplitCommandAcceptsFlagsOnEitherSide(t *testing.T) {
	// `gate verify -no-cache` reads naturally and was silently dropping the
	// flag before this existed. A flag that is ignored rather than rejected
	// makes the cache look broken instead of the parser.
	cases := []struct {
		name     string
		argv     []string
		wantCmd  string
		wantRest []string
	}{
		{"flag after", []string{"verify", "-no-cache"}, "verify", []string{"-no-cache"}},
		{"flag before", []string{"-no-cache", "verify"}, "verify", []string{"-no-cache"}},
		{"flags both sides", []string{"-v", "t2", "-no-cache"}, "t2", []string{"-v", "-no-cache"}},
		{"no flags", []string{"canary"}, "canary", nil},
		{"nothing", nil, "", nil},
		{"flags only", []string{"-v"}, "", []string{"-v"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd, rest := splitCommand(c.argv)
			if cmd != c.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, c.wantCmd)
			}
			if strings.Join(rest, " ") != strings.Join(c.wantRest, " ") {
				t.Errorf("rest = %v, want %v", rest, c.wantRest)
			}
		})
	}
}

func TestTierNameCoversEveryTier(t *testing.T) {
	for tier := 0; tier <= 4; tier++ {
		if tierName(tier) == "" {
			t.Errorf("tier %d has no name", tier)
		}
	}
}

func TestJudgementGatesAreDeclared(t *testing.T) {
	// These are the obligations no script can discharge. An empty list would
	// mean `gate verify` silently claims full coverage it does not have.
	if len(judgementGates) == 0 {
		t.Fatal("judgementGates is empty — verify would imply nothing is owed")
	}
	for _, g := range judgementGates {
		if !strings.Contains(g, "Tier") && !strings.Contains(g, "human") {
			t.Errorf("%q should say which tier owns it, or that a human does", g)
		}
	}
}

func TestTierZeroStaysCheap(t *testing.T) {
	// Tier 0 runs after every edit. If something slow lands in it, the loop
	// stops running it, and the fastest feedback in the system disappears.
	allowed := map[string]bool{"gofmt": true, "build": true, "vet": true}

	for _, c := range checksForTier(0) {
		if !allowed[c.ID] {
			t.Errorf("%q was added to tier 0 — confirm it is sub-second, then update this test", c.ID)
		}
	}
}
