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
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	skipDirs = map[string]bool{
		".git": true, "node_modules": true, "bin": true, "dist": true,
		"build": true, ".dart_tool": true, "vendor": true, "cache": true,
	}
	sourceExts = map[string]bool{
		".go": true, ".dart": true, ".kt": true, ".swift": true,
		".yml": true, ".yaml": true,
	}

	secretNameRe = regexp.MustCompile(`(?i)(api_?key|secret|passw(or)?d|token|credential|private_?key)`)
	publicishRe  = regexp.MustCompile(`(?i)(public|pub_?key|fingerprint|_?ttl$|hash)`)

	logCallRe = map[string]bool{
		"Print": true, "Printf": true, "Println": true, "Sprintf": true,
		"Info": true, "Warn": true, "Error": true, "Debug": true,
		"Infof": true, "Warnf": true, "Errorf": true, "Debugf": true,
		"Fatal": true, "Fatalf": true, "Log": true, "Logf": true,
	}
	assertCallRe = map[string]bool{
		"Error": true, "Errorf": true, "Fatal": true, "Fatalf": true,
		"Fail": true, "FailNow": true, "Skip": true, "Skipf": true, "SkipNow": true,
	}
)

// walkSource visits every source file in the repository.
func walkSource(root string, fn func(path string, body []byte) error) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !sourceExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		body, err := os.ReadFile(path) //nolint:gosec // walking our own repository
		if err != nil {
			return err
		}
		return fn(path, body)
	})
}

func hasDirective(body []byte, directive string) bool {
	return strings.Contains(string(body), directive)
}

// ── Tier 0: deterministic, sub-second, no tokens ────────────────────────────

func checkGofmt(root string) Result {
	res := Result{Name: "gofmt", Tier: 0}

	out := runCmd(root, "gofmt", "-l", ".")
	if out.missing {
		return res.unverified("gofmt is not on PATH")
	}
	if out.err != nil {
		return res.unverified("gofmt could not run: %v", out.err)
	}

	files := strings.Fields(strings.TrimSpace(out.stdout))
	if len(files) > 0 {
		return res.fail(prefixEach("needs formatting: ", files)...)
	}
	res.Summary = "all Go files formatted"
	return res.pass()
}

func checkBuild(root string) Result {
	res := Result{Name: "build", Tier: 0}

	out := runCmd(root, "go", "build", "./...")
	if out.missing {
		return res.unverified("go is not on PATH")
	}
	if out.exitCode != 0 || out.err != nil {
		return res.fail(tail(out.stdout, 20)...)
	}
	res.Summary = "compiles"
	return res.pass()
}

func checkVet(root string) Result {
	res := Result{Name: "vet", Tier: 0}

	out := runCmd(root, "go", "vet", "./...")
	if out.missing {
		return res.unverified("go is not on PATH")
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 20)...)
	}
	res.Summary = "no suspicious constructs"
	return res.pass()
}

// ── Tier 1: inner loop — every fix iteration ────────────────────────────────

func checkLint(root string) Result {
	res := Result{Name: "lint", Tier: 1}

	out := runCmd(root, "golangci-lint", "run", "./...")
	if out.missing {
		return res.unverified("golangci-lint is not installed — run `make tools`")
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 30)...)
	}
	res.Summary = "clean"
	return res.pass()
}

// checkTests runs the suite and, as importantly, counts it.
//
// `-count=1` disables Go's own test cache deliberately. A cached package emits
// no per-test events, so a warm cache would report a suite that shrank to
// nothing — and this gate's whole second job is noticing exactly that. When
// the gate's own signature says the inputs are unchanged it is skipped
// wholesale; when it runs, it really runs.
//
// SPEC-deterministic-backbone-02
func checkTests(root string) Result {
	res := Result{Name: "tests", Tier: 1}

	out := runCmd(root, "go", "test", "-json", "-count=1", "./...")
	if out.missing {
		return res.unverified("go is not on PATH")
	}

	tally := parseGoTestJSON(out.stdout)
	res = res.count("tests_run", tally.run).
		count("tests_skipped", tally.skipped).
		count("packages_tested", tally.packages)
	res.Summary = fmt.Sprintf("%d tests in %d packages, %d skipped",
		tally.run, tally.packages, tally.skipped)

	if out.exitCode != 0 {
		findings := tally.failures
		if len(findings) == 0 {
			// No test failed, yet the command did: a package that would not
			// build, or a toolchain error. Either way the suite is not
			// evidence, and the raw output is the only useful thing to show.
			findings = tail(plainLines(out.stdout), 20)
		}
		return res.fail(tail(strings.Join(findings, "\n"), 30)...)
	}
	return res.pass()
}

// goTestEvent is one line of `go test -json`.
type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// testTally is the count behind the exit code.
//
// Skips are counted apart from runs on purpose: a suite where a third of the
// tests quietly began skipping exits 0 and looks identical to a healthy one.
type testTally struct {
	run      int
	skipped  int
	packages int
	failures []string
}

func parseGoTestJSON(stdout string) testTally {
	var tally testTally
	outputs := map[string][]string{}
	counted := map[string]bool{}

	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var ev goTestEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}

		key := ev.Package + "." + ev.Test
		switch ev.Action {
		case "output":
			if ev.Test != "" {
				outputs[key] = append(outputs[key], strings.TrimRight(ev.Output, "\r\n"))
			}
		case "pass", "fail", "skip":
			if ev.Test == "" {
				if ev.Action != "skip" {
					tally.packages++
				}
				continue
			}
			if counted[key] {
				continue
			}
			counted[key] = true
			switch ev.Action {
			case "skip":
				tally.skipped++
			case "fail":
				tally.run++
				tally.failures = append(tally.failures, outputs[key]...)
			default:
				tally.run++
			}
		}
	}
	return tally
}

// plainLines keeps the non-JSON part of a `go test -json` stream — build
// errors arrive there, and printing raw event objects at a human helps nobody.
func plainLines(stdout string) string {
	var kept []string
	for _, line := range strings.Split(stdout, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "{") {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// checkZeroTouchAI enforces the project's flagship invariant structurally:
// relay code must not be able to reach an AI provider, because it must not be
// able to import the packages that talk to one.
//
// A promise in a document can be broken by accident. An import graph cannot.
func checkZeroTouchAI(root string) Result {
	res := Result{Name: "zero-touch-ai", Tier: 1}

	relayPrefixes := []string{
		"cmd/rla-server", "internal/server", "internal/stream", "internal/push",
	}
	forbidden := []string{
		"internal/models", // provider clients
		"internal/agent",  // agent loop
		"internal/tools",  // filesystem and exec
		"internal/crypto", // the relay never holds keys (ADR-004)
		"go-openai",       // any direct provider SDK
	}

	var findings []string
	relayFiles := 0

	err := walkGo(root, func(path string, file *ast.File) {
		relPath := relTo(root, path)
		if !hasAnyPrefix(relPath, relayPrefixes) {
			return
		}
		relayFiles++
		for _, imp := range file.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			for _, bad := range forbidden {
				if strings.Contains(p, bad) {
					findings = append(findings,
						fmt.Sprintf("%s imports %q — relay code must never reach provider, agent or key material", relPath, p))
				}
			}
		}
	})
	if err != nil {
		return res.unverified("import scan failed: %v", err)
	}

	// The count is recorded rather than guarded: this rule genuinely holds
	// vacuously until P2 creates the relay, and a floor here would fail a gate
	// for being early rather than for being blind. Recording it means the
	// artifact says "0 relay files" instead of implying the rule was tested.
	res = res.count("relay_files", relayFiles)

	if relayFiles == 0 {
		res.Summary = "no relay packages exist yet — rule holds vacuously (tripwire armed for P2)"
		return res.pass()
	}
	res.Summary = fmt.Sprintf("%d relay files, no forbidden imports", relayFiles)
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	return res.pass()
}

// checkSecretLogging looks for credentials interpolated into log or print
// calls. Redaction is asserted by test elsewhere; this catches the mistake at
// the point it is written.
func checkSecretLogging(root string) Result {
	res := Result{Name: "secret-logging", Tier: 1}

	var findings []string
	scanned := 0

	err := walkGoWithSource(root, func(path string, file *ast.File, fset *token.FileSet, body []byte) {
		if hasDirective(body, "//gate:allow-secret-log") {
			return
		}
		scanned++
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !logCallRe[sel.Sel.Name] {
				return true
			}
			for _, arg := range call.Args {
				name := exprName(arg)
				if name == "" || publicishRe.MatchString(name) {
					continue
				}
				if secretNameRe.MatchString(name) {
					pos := fset.Position(arg.Pos())
					findings = append(findings, fmt.Sprintf(
						"%s:%d: %q passed to %s — redact it, or annotate the file with //gate:allow-secret-log",
						rel(path), pos.Line, name, sel.Sel.Name))
				}
			}
			return true
		})
	})
	if err != nil {
		return res.unverified("AST scan failed: %v", err)
	}

	res = res.count("files_scanned", scanned)
	res.Summary = fmt.Sprintf("%d files scanned", scanned)
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	return res.pass()
}

// checkFakeGreen hunts tests that pass without proving anything — the most
// common way a suite becomes decorative.
func checkFakeGreen(root string) Result {
	res := Result{Name: "fake-green", Tier: 1}

	var findings []string
	tests := 0

	err := walkGoWithSource(root, func(path string, file *ast.File, fset *token.FileSet, body []byte) {
		if !strings.HasSuffix(path, "_test.go") || hasDirective(body, "//gate:allow-no-assert") {
			return
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			tests++
			if !hasAssertion(fn.Body) {
				pos := fset.Position(fn.Pos())
				findings = append(findings, fmt.Sprintf(
					"%s:%d: %s asserts nothing — a test that cannot fail is not evidence",
					rel(path), pos.Line, fn.Name.Name))
			}
		}
	})
	if err != nil {
		return res.unverified("AST scan failed: %v", err)
	}

	res = res.count("tests_declared", tests)
	if tests == 0 {
		return res.unverified("no test functions found — nothing to audit")
	}
	res.Summary = fmt.Sprintf("%d tests, all assert", tests)
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	return res.pass()
}

func hasAssertion(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && assertCallRe[sel.Sel.Name] {
			found = true
			return false
		}
		return true
	})
	return found
}

// ── Tier 2: convergence — once per feature ──────────────────────────────────

const coverageFloorFile = ".rla/coverage-floor.txt"

// checkCoverage ratchets rather than fixing an arbitrary threshold: coverage
// may rise freely, but a drop below the recorded floor fails the gate. This is
// the enforceable form of "coverage of changed code must not drop".
func checkCoverage(root string) Result {
	res := Result{Name: "coverage-ratchet", Tier: 2}

	profile := filepath.Join(root, ".rla", "cache", "coverage.out")
	if err := os.MkdirAll(filepath.Dir(profile), 0o750); err != nil {
		return res.unverified("cache dir: %v", err)
	}

	out := runCmd(root, "go", "test", "-count=1", "-coverprofile="+profile, "./...")
	if out.missing {
		return res.unverified("go is not on PATH")
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 20)...)
	}

	fnOut := runCmd(root, "go", "tool", "cover", "-func="+profile)
	if fnOut.exitCode != 0 {
		return res.unverified("go tool cover failed: %s", strings.TrimSpace(fnOut.stdout))
	}

	total, ok := parseTotalCoverage(fnOut.stdout)
	if !ok {
		return res.unverified("could not parse coverage total")
	}

	// A percentage says nothing about how much was weighed. 100% of two
	// functions and 100% of two hundred are the same number and not the same
	// claim, and the profile silently narrowing is how the second becomes the
	// first without anyone noticing.
	res = res.count("functions_measured", countCoveredFunctions(fnOut.stdout))

	floor := readCoverageFloor(root)
	res.Summary = fmt.Sprintf("%.1f%% statements (floor %.1f%%)", total, floor)

	// Half a point of slack absorbs rounding when unrelated code moves.
	if total < floor-0.5 {
		return res.fail(fmt.Sprintf(
			"coverage fell to %.1f%% from a floor of %.1f%% — add tests, or lower the floor deliberately in %s",
			total, floor, coverageFloorFile))
	}
	return res.pass()
}

func parseTotalCoverage(out string) (float64, bool) {
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "total:") {
			continue
		}
		fields := strings.Fields(line)
		raw := strings.TrimSuffix(fields[len(fields)-1], "%")
		v, err := strconv.ParseFloat(raw, 64)
		return v, err == nil
	}
	return 0, false
}

// countCoveredFunctions counts the function rows in `go tool cover -func`
// output — how much source the profile actually spanned, as distinct from the
// percentage of it that was exercised.
func countCoveredFunctions(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "total:") {
			continue
		}
		if strings.Contains(line, "%") {
			n++
		}
	}
	return n
}

func readCoverageFloor(root string) float64 {
	body, err := os.ReadFile(filepath.Join(root, coverageFloorFile)) //nolint:gosec // fixed path
	if err != nil {
		return 0
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(body)), 64)
	if err != nil {
		return 0
	}
	return v
}

func checkDocLinks(root string) Result {
	res := Result{Name: "doc-links", Tier: 2}

	out := runCmd(root, "go", "run", "./scripts/checkdocs")
	if out.missing {
		return res.unverified("go is not on PATH")
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 25)...)
	}
	res.Summary = strings.TrimSpace(lastLine(out.stdout))
	// checkdocs closes with "N files, M anchors indexed, K problems". A link
	// checker that indexed nothing reports no broken links.
	res = res.count("files_indexed", leadingInt(res.Summary))
	return res.pass()
}

// leadingInt reads the first integer in a summary line, or 0 when there is
// none — an unparsed summary must not be mistaken for a large count.
func leadingInt(s string) int {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSuffix(fields[0], ","))
	if err != nil {
		return 0
	}
	return n
}

func checkLicenceHeaders(root string) Result {
	res := Result{Name: "licence-headers", Tier: 2}

	args := []string{
		"-check", "-f", "LICENSE_HEADER", "-c", "Burak Halefoğlu", "-y", "2026",
		"-ignore", "**/*.md", "-ignore", "**/*.json", "-ignore", "**/*.lock",
		"-ignore", "**/vendor/**",
	}
	for _, dir := range []string{"cmd", "internal", "deploy", "scripts"} {
		if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
			args = append(args, dir)
		}
	}

	out := runCmd(root, "addlicense", args...)
	if out.missing {
		return res.unverified("addlicense is not installed — run `make tools`")
	}
	if out.exitCode != 0 {
		return res.fail(append([]string{"files missing an AGPL header:"}, tail(out.stdout, 20)...)...)
	}
	res.Summary = "AGPL headers present"
	return res.pass()
}

// checkLicenceBoundary guards the one-way rule from ADR-002: Apache-2.0 code
// may flow into the AGPL core, never the reverse.
func checkLicenceBoundary(root string) Result {
	res := Result{Name: "licence-boundary", Tier: 2}

	mobile := filepath.Join(root, "mobile")
	if _, err := os.Stat(mobile); os.IsNotExist(err) {
		res.Summary = "mobile/ does not exist yet — tripwire armed for P0.7"
		return res.pass()
	}

	var findings []string
	err := filepath.WalkDir(mobile, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".dart", ".kt", ".swift", ".gradle", ".yaml", ".yml":
		default:
			return nil
		}
		body, err := os.ReadFile(path) //nolint:gosec // walking our own repository
		if err != nil {
			return err
		}
		if strings.Contains(string(body), "GNU Affero General Public License") {
			findings = append(findings, rel(path)+" carries an AGPL header inside the Apache-2.0 boundary")
		}
		return nil
	})
	if err != nil {
		return res.unverified("mobile scan failed: %v", err)
	}

	res.Summary = "no AGPL code under mobile/"
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	return res.pass()
}

// ── Tier 3: heavy — once, when candidate-complete ───────────────────────────

func checkRace(root string) Result {
	res := Result{Name: "race", Tier: 3}

	out := runCmd(root, "go", "test", "-race", "./...")
	if out.missing {
		return res.unverified("go is not on PATH")
	}
	// The race detector needs cgo, which is commonly absent on Windows
	// toolchains. That is "could not verify", not "no races" — reporting a
	// pass here would be exactly the silent green this system exists to stop.
	if strings.Contains(out.stdout, "-race requires cgo") {
		return res.unverified("race detector unavailable: cgo is disabled (CI runs this on Linux)")
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 30)...)
	}
	res.Summary = "no data races detected"
	return res.pass()
}

func checkVulnerabilities(root string) Result {
	res := Result{Name: "vulnerabilities", Tier: 3}

	out := runCmd(root, "osv-scanner", "scan", "source", "-r", ".")
	if out.missing {
		out = runCmd(root, "govulncheck", "./...")
		if out.missing {
			return res.unverified("neither osv-scanner nor govulncheck is installed")
		}
	}
	if out.exitCode != 0 {
		return res.fail(tail(out.stdout, 25)...)
	}
	res.Summary = "no known vulnerable dependencies"
	return res.pass()
}

// ── helpers ─────────────────────────────────────────────────────────────────

func walkGo(root string, fn func(path string, file *ast.File)) error {
	return walkGoWithSource(root, func(path string, file *ast.File, _ *token.FileSet, _ []byte) {
		fn(path, file)
	})
}

func walkGoWithSource(root string, fn func(path string, file *ast.File, fset *token.FileSet, body []byte)) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		body, err := os.ReadFile(path) //nolint:gosec // walking our own repository
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(fset, path, body, parser.ParseComments)
		if err != nil {
			// A file that does not parse is the build gate's problem, not
			// ours: reporting it here would produce two red gates for one
			// cause, and the build gate's message is the useful one.
			return nil //nolint:nilerr // deliberate: parse errors belong to the build gate
		}
		fn(path, file, fset, body)
		return nil
	})
}

func exprName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	}
	return ""
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func prefixEach(prefix string, items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = prefix + s
	}
	return out
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}
