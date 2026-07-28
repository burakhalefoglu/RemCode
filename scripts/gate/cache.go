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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// gateVersion is the configuration signature. Bump it whenever a check's
// definition changes so every cached pass is invalidated — a gate that got
// stricter must not be skipped on the strength of an older, weaker run.
const gateVersion = "1"

const cachePath = ".rla/cache/gates.json"

// cacheEntry records a passing gate and the exact inputs that made it pass.
//
// The signature is computed from working-tree content, never from a commit
// SHA: `git commit` does not change what the gate read, so it must not
// invalidate anything.
//
// Evidence is stored alongside the verdict because a signature alone cannot
// tell a real pass from an empty one. The files that make a suite run zero
// tests do not change, so the signature keeps matching and the empty pass is
// served for ever. Recording the counts lets a hit be re-judged instead of
// trusted.
//
// SPEC-deterministic-backbone-03
type cacheEntry struct {
	Signature  string   `json:"signature"`
	Summary    string   `json:"summary"`
	Evidence   Evidence `json:"evidence,omitempty"`
	DurationMS int64    `json:"duration_ms,omitempty"`
}

func (e cacheEntry) duration() time.Duration {
	return time.Duration(e.DurationMS) * time.Millisecond
}

type gateCache struct {
	root    string
	entries map[string]cacheEntry
	dirty   bool
}

func loadCache(root string) *gateCache {
	c := &gateCache{root: root, entries: map[string]cacheEntry{}}

	body, err := os.ReadFile(filepath.Join(root, cachePath)) //nolint:gosec // fixed path
	if err != nil {
		return c
	}
	if err := json.Unmarshal(body, &c.entries); err != nil {
		// A corrupt cache must never be interpreted as "everything passed".
		c.entries = map[string]cacheEntry{}
	}
	return c
}

func (c *gateCache) save() error {
	if !c.dirty {
		return nil
	}
	path := filepath.Join(c.root, cachePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("cache dir: %w", err)
	}
	body, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode cache: %w", err)
	}
	return os.WriteFile(path, append(body, '\n'), 0o600)
}

// hit reports the cached pass when the signature still matches. Whether that
// pass may be *used* is a separate question, answered by the guards.
func (c *gateCache) hit(id, signature string) (cacheEntry, bool) {
	e, ok := c.entries[id]
	if !ok || e.Signature != signature {
		return cacheEntry{}, false
	}
	return e, true
}

// store records a pass together with what it saw. Failures and unverified
// results are never cached — they must be re-run until they genuinely pass.
func (c *gateCache) store(id, signature string, res Result) {
	c.entries[id] = cacheEntry{
		Signature:  signature,
		Summary:    res.Summary,
		Evidence:   res.Evidence,
		DurationMS: res.Duration.Milliseconds(),
	}
	c.dirty = true
}

func (c *gateCache) clear() error {
	c.entries = map[string]cacheEntry{}
	c.dirty = true
	return os.RemoveAll(filepath.Join(c.root, ".rla", "cache"))
}

// signature hashes gate version + ratified spec ids + the content of every
// file under the check's declared inputs.
func signature(root string, check Check, specIDs []string) (string, error) {
	if len(check.Inputs) == 0 {
		return "", nil // opt out of caching
	}

	h := sha256.New()
	fmt.Fprintf(h, "gate:%s\x00check:%s\x00", gateVersion, check.ID)

	sorted := append([]string(nil), specIDs...)
	sort.Strings(sorted)
	for _, id := range sorted {
		fmt.Fprintf(h, "spec:%s\x00", id)
	}

	var files []string
	for _, prefix := range check.Inputs {
		base := filepath.Join(root, filepath.FromSlash(prefix))
		info, err := os.Stat(base)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			files = append(files, base)
			continue
		}
		err = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if skipDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return "", err
		}
	}

	sort.Strings(files)
	for _, path := range files {
		body, err := os.ReadFile(path) //nolint:gosec // paths derive from check definitions
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(body)
		fmt.Fprintf(h, "%s:%s\x00", filepath.ToSlash(strings.TrimPrefix(path, root)), hex.EncodeToString(sum[:]))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
