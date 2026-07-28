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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Evidence, guards and the run artifact.
//
// The verdict of a deterministic gate is its exit code — nothing here produces
// an opinion. What this file adds is the part an exit code cannot say: *how
// much* the gate actually looked at. A suite that ran zero tests exits 0. A
// scanner that read zero files reports zero findings. Both are green, and both
// are lies of omission, so every gate also reports counts and the counts are
// held to declared floors.
//
// The artifact exists so the judged half of the pipeline — a model reviewing
// this work — reads what ran instead of running it again. Re-running is 10–30×
// the cost of reading, and its verdict is not reproducible.

const (
	// stateDir holds run artifacts. Regenerated, machine-local, never committed.
	stateDir = ".rla/state"

	// timingsBaseline is the committed wall-clock measurement every later run
	// is compared against. Recording a new one is a deliberate act
	// (`gate timings -record`), like lowering the coverage floor.
	timingsBaseline = ".rla/tool-timings.json"

	// testBaselineFile ratchets the number of tests that must run. Go exits
	// non-zero when a package fails to build, so the loud failure is already
	// covered; what this catches is the quiet one — a suite that shrank.
	testBaselineFile = ".rla/test-baseline.txt"
)

// Evidence is what a gate counted while it ran: tests executed, files read,
// requirements checked. Numbers only — a gate that writes a sentence here is
// producing judgement, which is not its job.
type Evidence map[string]int

// Guard is a rule evaluated against Evidence rather than against an exit code.
//
// Guards are the price of every convenience elsewhere in this system: caching,
// tier splitting and skipping all trade certainty for speed, and each one is
// only defensible while the counts are still checked.
type Guard struct {
	Name   string
	Status Status
	Detail string
}

// applyGuards holds a result to the counts it reported.
//
// A guard can only worsen a verdict. Clearing a red on the strength of a count
// would be a judged gate absolving a deterministic one, which is the exact
// inversion the pipeline forbids.
//
// SPEC-deterministic-backbone-02
func applyGuards(check Check, res Result, floor int) Result {
	for _, key := range sortedKeys(check.MinEvidence) {
		want := check.MinEvidence[key]
		got, reported := res.Evidence[key]
		switch {
		case !reported:
			res = addGuard(res, Guard{"empty-run", Unverified,
				fmt.Sprintf("%s reported no %q count — its green cannot be substantiated", check.ID, key)})
		case got < want:
			res = addGuard(res, Guard{"empty-run", Fail,
				fmt.Sprintf("%s: %s = %d, below the floor of %d — this gate examined nothing and still passed",
					check.ID, key, got, want)})
		default:
			res = addGuard(res, Guard{"empty-run", Pass, fmt.Sprintf("%s = %d", key, got)})
		}
	}

	if check.BaselineKey != "" {
		got := res.Evidence[check.BaselineKey]
		switch {
		case floor < 0:
			res = addGuard(res, Guard{"baseline", Unverified,
				fmt.Sprintf("no baseline recorded in %s — erosion of %s cannot be detected",
					check.BaselineFile, check.BaselineKey)})
		case got < floor:
			res = addGuard(res, Guard{"baseline", Fail,
				fmt.Sprintf("%s fell to %d from a baseline of %d — tests disappeared rather than failed; "+
					"restore them, or lower %s deliberately",
					check.BaselineKey, got, floor, check.BaselineFile)})
		default:
			res = addGuard(res, Guard{"baseline", Pass,
				fmt.Sprintf("%s = %d ≥ baseline %d", check.BaselineKey, got, floor)})
		}
	}

	if check.Budget > 0 {
		switch {
		case res.Duration > check.Budget:
			res = addGuard(res, Guard{"step-budget", Fail,
				fmt.Sprintf("%s took %s against a budget of %s — a step that outgrew its budget has "+
					"silently changed which tier it belongs in",
					check.ID, res.Duration.Round(time.Millisecond), check.Budget)})
		default:
			res = addGuard(res, Guard{"step-budget", Pass,
				fmt.Sprintf("%s of %s", res.Duration.Round(time.Millisecond), check.Budget)})
		}
	}

	return res
}

func addGuard(res Result, g Guard) Result {
	res.Guards = append(res.Guards, g)
	if g.Status != Pass {
		res.Status = worsen(res.Status, g.Status)
		res.Findings = append(res.Findings, "guard "+g.Name+": "+g.Detail)
	}
	return res
}

// guardsHold reports whether recorded evidence still satisfies a check's
// guards — the question asked of a *cache hit*.
//
// Without this, the cache is where a fake green goes to become permanent: the
// files that made a suite run zero tests do not change, so the signature keeps
// matching and the empty pass is served for ever. A cached green whose numbers
// would fail today is not a green.
//
// SPEC-deterministic-backbone-03
func guardsHold(check Check, ev Evidence, d time.Duration, floor int) bool {
	probe := Result{Name: check.ID, Status: Pass, Evidence: ev, Duration: d}
	return applyGuards(check, probe, floor).Status == Pass
}

// baselineFloor reads a check's recorded floor. A missing or unreadable file
// returns -1: an unmeasurable guard is reported as unverified, never as clear.
func baselineFloor(root string, check Check) int {
	if check.BaselineFile == "" {
		return 0
	}
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(check.BaselineFile))) //nolint:gosec // path from a gate definition
	if err != nil {
		return -1
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil || n < 0 {
		return -1
	}
	return n
}

// ── the run artifact ────────────────────────────────────────────────────────

// stepRecord is one gate's line in the artifact: what it decided, how it
// decided, and how much it saw.
type stepRecord struct {
	ID       string   `json:"id"`
	Tier     int      `json:"tier"`
	Verdict  string   `json:"verdict"`
	Exit     int      `json:"exit"`
	Duration int64    `json:"duration_ms"`
	Cached   bool     `json:"cached"`
	Summary  string   `json:"summary"`
	Evidence Evidence `json:"evidence,omitempty"`
	Guards   []struct {
		Name    string `json:"name"`
		Verdict string `json:"verdict"`
		Detail  string `json:"detail,omitempty"`
	} `json:"guards,omitempty"`
	Findings []string `json:"findings,omitempty"`
}

// deferredRecord names a gate this mode did not run.
//
// A skipped check that goes unnamed is indistinguishable from one that passed,
// and the reader of an artifact has no other way to tell.
type deferredRecord struct {
	ID     string `json:"id"`
	Tier   int    `json:"tier"`
	Reason string `json:"reason"`
}

// runArtifact is the complete machine-readable account of one run.
type runArtifact struct {
	Mode          string           `json:"mode"`
	Fingerprint   string           `json:"fingerprint"`
	Generated     string           `json:"generated"`
	Verdict       string           `json:"verdict"`
	Exit          int              `json:"exit"`
	DurationMS    int64            `json:"duration_ms"`
	Steps         []stepRecord     `json:"steps"`
	Deferred      []deferredRecord `json:"deferred"`
	JudgementOwed []string         `json:"judgement_owed"`
}

func newStepRecord(res Result) stepRecord {
	rec := stepRecord{
		ID:       res.Name,
		Tier:     res.Tier,
		Verdict:  res.Status.String(),
		Exit:     statusExit(res.Status),
		Duration: res.Duration.Milliseconds(),
		Cached:   res.Cached,
		Summary:  res.Summary,
		Evidence: res.Evidence,
		Findings: res.Findings,
	}
	for _, g := range res.Guards {
		rec.Guards = append(rec.Guards, struct {
			Name    string `json:"name"`
			Verdict string `json:"verdict"`
			Detail  string `json:"detail,omitempty"`
		}{Name: g.Name, Verdict: g.Status.String(), Detail: g.Detail})
	}
	return rec
}

// writeArtifact records the run under a name that carries the fingerprint of
// the tree it describes, so a stale artifact cannot masquerade as a fresh one.
//
// SPEC-deterministic-backbone-01
func writeArtifact(root string, a runArtifact) (string, error) {
	dir := filepath.Join(root, filepath.FromSlash(stateDir))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("state dir: %w", err)
	}

	path := filepath.Join(dir, artifactName(a.Mode, a.Fingerprint))
	body, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode artifact: %w", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write artifact: %w", err)
	}
	return filepath.ToSlash(filepath.Join(stateDir, filepath.Base(path))), nil
}

func artifactName(mode, fingerprint string) string {
	return fmt.Sprintf("verify-%s-%s.json", mode, short(fingerprint))
}

func short(fingerprint string) string {
	if len(fingerprint) > 12 {
		return fingerprint[:12]
	}
	return fingerprint
}

// readArtifact loads a run artifact and reports whether it still describes the
// working tree. A report whose fingerprint has moved on is not weak evidence —
// it is evidence about a different tree, and reading it is worse than reading
// nothing.
//
// SPEC-deterministic-backbone-04
func readArtifact(root, mode, fingerprint string) (runArtifact, bool, error) {
	path := filepath.Join(root, filepath.FromSlash(stateDir), artifactName(mode, fingerprint))
	body, err := os.ReadFile(path) //nolint:gosec // fixed, derived path
	if err != nil {
		return runArtifact{}, false, err
	}
	var a runArtifact
	if err := json.Unmarshal(body, &a); err != nil {
		return runArtifact{}, false, fmt.Errorf("decode %s: %w", rel(path), err)
	}
	return a, a.Fingerprint == fingerprint, nil
}

// treeFingerprint hashes everything any gate reads, so one value answers
// "is this report about the tree I am looking at?".
func treeFingerprint(root string, specIDs []string) (string, error) {
	seen := map[string]bool{}
	var inputs []string
	for _, c := range registry() {
		for _, in := range c.Inputs {
			if !seen[in] {
				seen[in] = true
				inputs = append(inputs, in)
			}
		}
	}
	sort.Strings(inputs)
	return signature(root, Check{ID: "__tree", Inputs: inputs}, specIDs)
}

// ── measured wall clock ─────────────────────────────────────────────────────

// timingRecord is one measured step. Budgets in this repository were set from
// these numbers rather than from intuition; the reverse order is how a tier
// split ends up defending a cost nobody measured.
type timingRecord struct {
	Milliseconds int64  `json:"ms"`
	Verdict      string `json:"verdict"`
}

type timings struct {
	Measured string           `json:"measured"`
	Total    int64            `json:"total_ms"`
	Modes    map[string]int64 `json:"mode_ms"`
	// Unmeasured names the checks that could not run when this was recorded —
	// usually an uninstalled tool. Their cost is missing from the totals, and a
	// baseline that hides that would report the first complete run as drift.
	Unmeasured []string                `json:"unmeasured,omitempty"`
	Checks     map[string]timingRecord `json:"checks"`
}

func loadTimings(root, path string) (timings, error) {
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path))) //nolint:gosec // fixed path
	if err != nil {
		return timings{}, err
	}
	var t timings
	if err := json.Unmarshal(body, &t); err != nil {
		return timings{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return t, nil
}

func saveTimings(root, path string, t timings) error {
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return fmt.Errorf("timings dir: %w", err)
	}
	body, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("encode timings: %w", err)
	}
	return os.WriteFile(full, append(body, '\n'), 0o600)
}

// driftAgainst reports how far a mode's measured cost has moved from the
// committed baseline. The number matters more than it looks: a fast layer
// creeping towards its budget is the first sign that work was quietly added to
// the loop everyone runs a hundred times a day.
func driftAgainst(baseline timings, mode string, measured time.Duration) (string, bool) {
	base, ok := baseline.Modes[mode]
	if !ok || base <= 0 {
		return "", false
	}
	ratio := float64(measured.Milliseconds()) / float64(base)
	return fmt.Sprintf("%s: %s measured, baseline %s (%.0f%%)",
		mode, measured.Round(time.Millisecond), time.Duration(base)*time.Millisecond, ratio*100), ratio > 1.5
}

func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
