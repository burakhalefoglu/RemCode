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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Spec statuses. A pipeline halts at checkpoint ① until a draft is ratified.
const (
	StatusDraft    = "draft"
	StatusRatified = "ratified"
	StatusArchived = "archived"
)

var (
	// "## SPEC-<feature>-NN — Title"
	reqHeadingRe = regexp.MustCompile(`^##\s+(SPEC-[a-z0-9][a-z0-9-]*-\d{2})\s*[—:-]?\s*(.*)$`)
	specIDRe     = regexp.MustCompile(`SPEC-[a-z0-9][a-z0-9-]*-\d{2}`)
)

// Requirement is a single numbered obligation inside a spec.
type Requirement struct {
	ID    string
	Title string
	Line  int
}

// Spec is one `.rla/specs/<feature>.md` artifact.
type Spec struct {
	Path         string
	ID           string
	Title        string
	Status       string
	Phase        string
	Requirements []Requirement
}

// Draft reports whether this spec still awaits human ratification.
func (s Spec) Draft() bool { return s.Status == StatusDraft }

// Active reports whether the spec's requirements are currently binding.
func (s Spec) Active() bool { return s.Status == StatusRatified }

// loadSpecs reads every spec artifact under .rla/specs/.
func loadSpecs(root string) ([]Spec, error) {
	dir := filepath.Join(root, ".rla", "specs")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read spec dir: %w", err)
	}

	var specs []Spec
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "_") {
			continue
		}
		s, err := parseSpec(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
	return specs, nil
}

func parseSpec(path string) (Spec, error) {
	f, err := os.Open(path) //nolint:gosec // path comes from our own spec directory
	if err != nil {
		return Spec{}, fmt.Errorf("open spec: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only

	spec := Spec{Path: path}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	inFrontMatter := false
	lineNo := 0

	for sc.Scan() {
		lineNo++
		line := sc.Text()
		trimmed := strings.TrimSpace(line)

		if lineNo == 1 && trimmed == "---" {
			inFrontMatter = true
			continue
		}
		if inFrontMatter {
			if trimmed == "---" {
				inFrontMatter = false
				continue
			}
			key, value, found := strings.Cut(trimmed, ":")
			if !found {
				continue
			}
			value = strings.TrimSpace(value)
			switch strings.TrimSpace(key) {
			case "id":
				spec.ID = value
			case "title":
				spec.Title = value
			case "status":
				spec.Status = value
			case "phase":
				spec.Phase = value
			}
			continue
		}

		if m := reqHeadingRe.FindStringSubmatch(line); m != nil {
			spec.Requirements = append(spec.Requirements, Requirement{
				ID:    m[1],
				Title: strings.TrimSpace(m[2]),
				Line:  lineNo,
			})
		}
	}
	if err := sc.Err(); err != nil {
		return Spec{}, fmt.Errorf("scan %s: %w", path, err)
	}

	if spec.ID == "" {
		return Spec{}, fmt.Errorf("%s: front matter is missing `id`", rel(path))
	}
	if spec.Status == "" {
		return Spec{}, fmt.Errorf("%s: front matter is missing `status`", rel(path))
	}
	switch spec.Status {
	case StatusDraft, StatusRatified, StatusArchived:
	default:
		return Spec{}, fmt.Errorf("%s: unknown status %q", rel(path), spec.Status)
	}

	return spec, nil
}

// annotationIndex maps every SPEC id found in source to the files citing it.
func annotationIndex(root string) (map[string][]string, error) {
	index := map[string][]string{}

	err := walkSource(root, func(path string, body []byte) error {
		// The spec artifacts themselves are the requirement, not a citation.
		if strings.Contains(filepath.ToSlash(path), "/.rla/specs/") {
			return nil
		}
		// Files that contain SPEC-shaped strings as test data — the gate's own
		// fixtures and canaries — must declare so. Counting a fixture as a
		// citation would let a requirement look implemented because a test
		// mentions it.
		if hasDirective(body, "//gate:spec-fixtures") {
			return nil
		}
		for _, id := range specIDRe.FindAllString(string(body), -1) {
			r := rel(path)
			if !contains(index[id], r) {
				index[id] = append(index[id], r)
			}
		}
		return nil
	})
	return index, err
}

// checkSpecFidelity implements the forward direction of the bidirectional diff:
// every requirement in a ratified spec must be cited somewhere in the source.
//
// The backward direction — behaviour in the code that no requirement covers —
// needs judgement and cannot be decided here. `gate verify` reports it as a
// review obligation rather than pretending it was checked.
func checkSpecFidelity(root string) Result {
	res := Result{Name: "spec-fidelity-forward", Tier: 2}

	specs, err := loadSpecs(root)
	if err != nil {
		return res.unverified("spec artifacts could not be parsed: %v", err)
	}
	if len(specs) == 0 {
		return res.unverified("no spec artifacts found under .rla/specs/")
	}

	index, err := annotationIndex(root)
	if err != nil {
		return res.unverified("source scan failed: %v", err)
	}

	seen := map[string]string{}
	var findings []string
	ratified, covered := 0, 0

	for _, s := range specs {
		for _, r := range s.Requirements {
			if prev, dup := seen[r.ID]; dup {
				findings = append(findings,
					fmt.Sprintf("%s is declared twice (%s and %s)", r.ID, prev, rel(s.Path)))
				continue
			}
			seen[r.ID] = rel(s.Path)

			if !s.Active() {
				continue
			}
			ratified++
			if len(index[r.ID]) == 0 {
				findings = append(findings,
					fmt.Sprintf("%s has no implementation citing it — %s:%d %q",
						r.ID, rel(s.Path), r.Line, r.Title))
				continue
			}
			covered++
		}
	}

	// A citation of an id that no spec declares is a dangling reference: either
	// a typo, or a requirement that was deleted while the code kept its marker.
	for id, files := range index {
		if _, declared := seen[id]; !declared {
			findings = append(findings,
				fmt.Sprintf("%s is cited by %s but no spec declares it", id, strings.Join(files, ", ")))
		}
	}

	res.Summary = fmt.Sprintf("%d/%d ratified requirements implemented", covered, ratified)
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	if ratified == 0 {
		return res.unverified("no ratified requirements to verify — every spec is still draft")
	}
	return res.pass()
}

// checkSpecHygiene validates the artifacts themselves. A malformed spec must
// never silently reduce what the fidelity gate believes it is checking.
func checkSpecHygiene(root string) Result {
	res := Result{Name: "spec-hygiene", Tier: 1}

	specs, err := loadSpecs(root)
	if err != nil {
		return res.fail(err.Error())
	}
	if len(specs) == 0 {
		return res.unverified("no spec artifacts found under .rla/specs/")
	}

	var findings []string
	for _, s := range specs {
		base := strings.TrimSuffix(filepath.Base(s.Path), ".md")
		if s.ID != base {
			findings = append(findings,
				fmt.Sprintf("%s: id %q does not match the file name", rel(s.Path), s.ID))
		}
		if s.Title == "" {
			findings = append(findings, fmt.Sprintf("%s: front matter is missing `title`", rel(s.Path)))
		}
		if len(s.Requirements) == 0 && s.Status != StatusArchived {
			findings = append(findings,
				fmt.Sprintf("%s: no SPEC-… requirements — a spec with nothing to verify is not a spec", rel(s.Path)))
		}
		for _, r := range s.Requirements {
			if !strings.HasPrefix(r.ID, "SPEC-"+s.ID+"-") {
				findings = append(findings,
					fmt.Sprintf("%s:%d: %s does not belong to spec %q", rel(s.Path), r.Line, r.ID, s.ID))
			}
			if r.Title == "" {
				findings = append(findings, fmt.Sprintf("%s:%d: %s has no title", rel(s.Path), r.Line, r.ID))
			}
		}
	}

	res.Summary = fmt.Sprintf("%d spec artifacts", len(specs))
	if len(findings) > 0 {
		return res.fail(findings...)
	}
	return res.pass()
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
