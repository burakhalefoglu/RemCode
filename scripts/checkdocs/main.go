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

// Command checkdocs verifies that every relative Markdown link in the
// repository resolves — both the file and, where present, the heading anchor.
//
// This exists because the project previously shipped a contributing guide
// whose documentation links all 404'd. Anchors are worse than paths: they rot
// silently when a heading is reworded, and nothing complains. So CI checks.
//
//	go run ./scripts/checkdocs          # whole repo
//	go run ./scripts/checkdocs docs     # a subtree
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// Markdown inline links: [text](target). Reference-style links and bare
	// URLs are out of scope — neither has caused a problem here.
	linkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)

	headingRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+?)\s*$`)

	// Fenced code blocks are stripped before scanning: a link inside an
	// example is documentation, not a reference to check.
	fenceRe = regexp.MustCompile("(?s)```.*?```|(?m)^~~~.*?^~~~")

	// GitHub's anchor slug keeps letters, digits, spaces, hyphens and
	// underscores; everything else is dropped, then spaces become hyphens.
	// Unicode letters survive, which matters for Turkish headings.
	slugDropRe = regexp.MustCompile(`[^\p{L}\p{N} _-]+`)

	skipDirs = map[string]bool{
		".git": true, "node_modules": true, "bin": true, "build": true,
		".dart_tool": true, "vendor": true,
	}
)

// slug reproduces GitHub's heading-to-anchor transformation.
func slug(heading string) string {
	s := strings.ToLower(heading)
	s = slugDropRe.ReplaceAllString(s, "")
	return strings.ReplaceAll(s, " ", "-")
}

type problem struct {
	file, link, reason string
	line               int
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	files, err := collect(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk %s: %v\n", root, err)
		os.Exit(1)
	}

	// Index every anchor in the repository first — a link may point anywhere.
	anchors := map[string]map[string]bool{}
	for _, f := range files {
		body, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			os.Exit(1)
		}
		set := map[string]bool{}
		for _, m := range headingRe.FindAllStringSubmatch(string(body), -1) {
			set[slug(m[2])] = true
		}
		anchors[filepath.ToSlash(f)] = set
	}

	var problems []problem
	for _, f := range files {
		problems = append(problems, check(f, anchors)...)
	}

	sort.Slice(problems, func(i, j int) bool {
		if problems[i].file != problems[j].file {
			return problems[i].file < problems[j].file
		}
		return problems[i].line < problems[j].line
	})

	for _, p := range problems {
		fmt.Printf("%s:%d  %s\n    %s\n", p.file, p.line, p.reason, p.link)
	}

	total := 0
	for _, set := range anchors {
		total += len(set)
	}
	fmt.Printf("\n%d files, %d anchors indexed, %d problems\n", len(files), total, len(problems))

	if len(problems) > 0 {
		os.Exit(1)
	}
}

func collect(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			out = append(out, path)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func check(file string, anchors map[string]map[string]bool) []problem {
	raw, err := os.ReadFile(file)
	if err != nil {
		return []problem{{file: file, reason: "unreadable", link: err.Error()}}
	}

	// Blank out fenced blocks while preserving newlines, so line numbers in
	// the report still match the real file.
	text := fenceRe.ReplaceAllStringFunc(string(raw), func(block string) string {
		return strings.Repeat("\n", strings.Count(block, "\n"))
	})

	dir := filepath.Dir(file)
	var problems []problem

	for _, line := range indexLines(text) {
		for _, m := range linkRe.FindAllStringSubmatch(line.text, -1) {
			target := m[1]

			switch {
			case strings.HasPrefix(target, "http://"),
				strings.HasPrefix(target, "https://"),
				strings.HasPrefix(target, "mailto:"),
				strings.HasPrefix(target, "data:"):
				continue
			// ../../issues style links resolve on GitHub but not on disk.
			case strings.Contains(target, "../../issues"),
				strings.Contains(target, "../../discussions"),
				strings.Contains(target, "../../pulls"):
				continue
			}

			path, anchor, hasAnchor := strings.Cut(target, "#")

			resolved := filepath.ToSlash(file)
			if path != "" {
				resolved = filepath.ToSlash(filepath.Clean(filepath.Join(dir, path)))
				if _, err := os.Stat(resolved); err != nil {
					problems = append(problems, problem{file, target, "missing file", line.n})
					continue
				}
			}

			if !hasAnchor || anchor == "" {
				continue
			}

			set, indexed := anchors[resolved]
			if !indexed {
				// Anchors can only be checked in Markdown we indexed.
				continue
			}
			if !set[anchor] {
				problems = append(problems, problem{file, target, "missing anchor", line.n})
			}
		}
	}
	return problems
}

type numberedLine struct {
	n    int
	text string
}

func indexLines(text string) []numberedLine {
	split := strings.Split(text, "\n")
	out := make([]numberedLine, len(split))
	for i, l := range split {
		out[i] = numberedLine{n: i + 1, text: l}
	}
	return out
}
