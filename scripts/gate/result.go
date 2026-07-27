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
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Status is the outcome of a gate.
//
// Unverified is deliberately distinct from Fail: a gate that could not run
// proves nothing, and treating it as a pass is precisely the silent-green
// failure this whole system exists to prevent.
type Status int

const (
	Pass Status = iota
	Fail
	Unverified
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "PASSED"
	case Fail:
		return "FAILED"
	default:
		return "COULD NOT VERIFY"
	}
}

func (s Status) icon() string {
	switch s {
	case Pass:
		return "🟢"
	case Fail:
		return "🔴"
	default:
		return "⚠️ "
	}
}

// Result is what a single gate reports.
type Result struct {
	Name     string
	Tier     int
	Status   Status
	Summary  string
	Findings []string
	Duration time.Duration
	Cached   bool
}

func (r Result) pass() Result {
	r.Status = Pass
	return r
}

func (r Result) fail(findings ...string) Result {
	r.Status = Fail
	r.Findings = append(r.Findings, findings...)
	return r
}

func (r Result) unverified(format string, args ...any) Result {
	r.Status = Unverified
	r.Summary = fmt.Sprintf(format, args...)
	return r
}

// Check is one gate definition.
//
// Definitions live in compiled Go rather than in a config file on purpose:
// weakening a gate must require a code change that shows up in review. The
// loop is rewarded for green gates, so it must not be able to edit them.
type Check struct {
	ID   string
	Tier int
	Desc string

	// Inputs are path prefixes whose content signs the cache entry. When all
	// of them are unchanged and the spec set is unchanged, the gate is skipped.
	// Empty means never cached.
	Inputs []string

	Run func(root string) Result

	// CanaryFiles are planted into an empty directory on which this check must
	// report Fail. A gate that cannot detect deliberate breakage proves
	// nothing when it reports green, so `gate canary` verifies each one.
	// Absent means the canary is unimplemented — reported as unverified, never
	// as a pass.
	CanaryFiles map[string]string
}

// ── external command helpers ────────────────────────────────────────────────

type cmdOutput struct {
	stdout   string
	exitCode int
	missing  bool
	err      error
}

// cmdTimeout bounds every external gate command. A hung `go test` must fail
// the gate rather than hang the loop — an unbounded wait is indistinguishable
// from a gate that will never answer.
const cmdTimeout = 10 * time.Minute

func runCmd(root, name string, args ...string) cmdOutput {
	if _, err := exec.LookPath(name); err != nil {
		return cmdOutput{missing: true, err: err}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // fixed tool names from gate definitions
	cmd.Dir = root

	out, err := cmd.CombinedOutput()
	result := cmdOutput{stdout: string(out)}

	if err != nil {
		var exitErr *exec.ExitError
		if ok := asExitError(err, &exitErr); ok {
			result.exitCode = exitErr.ExitCode()
			return result
		}
		result.err = err
		result.exitCode = -1
	}
	return result
}

func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok { //nolint:errorlint // direct type is sufficient here
		*target = e
		return true
	}
	return false
}

// tail keeps command output readable when a tool is verbose.
func tail(s string, lines int) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	all := strings.Split(s, "\n")
	if len(all) <= lines {
		return all
	}
	out := append([]string{fmt.Sprintf("… %d earlier lines omitted", len(all)-lines)}, all[len(all)-lines:]...)
	return out
}

// relTo renders path relative to an explicit root. Checks that reason about
// directory layout must use this rather than rel, so canary runs against a
// temporary tree see the same prefixes as a real run.
func relTo(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

// rel renders a path relative to the repository root for readable output.
func rel(path string) string {
	root, err := repoRoot()
	if err != nil {
		return filepath.ToSlash(path)
	}
	if r, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(path)
}

var cachedRoot string

func repoRoot() (string, error) {
	if cachedRoot != "" {
		return cachedRoot, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			cachedRoot = dir
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}
