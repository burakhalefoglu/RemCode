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

package version

import (
	"strings"
	"testing"
)

func TestGetPopulatesRuntimeFields(t *testing.T) {
	info := Get()

	if info.GoVersion == "" {
		t.Error("GoVersion is empty")
	}
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform = %q, want GOOS/GOARCH", info.Platform)
	}
	if info.Protocol != Protocol {
		t.Errorf("Protocol = %d, want %d", info.Protocol, Protocol)
	}
}

func TestGetReportsStampedValues(t *testing.T) {
	// An unstamped build must still report something usable rather than empty
	// strings — a bug report with a blank version is not actionable.
	info := Get()

	for name, got := range map[string]string{
		"Version": info.Version,
		"Commit":  info.Commit,
		"Date":    info.Date,
	} {
		if got == "" {
			t.Errorf("%s is empty; defaults should apply when not stamped", name)
		}
	}

	if Short() != info.Version {
		t.Errorf("Short() = %q, Get().Version = %q; must agree", Short(), info.Version)
	}
}

func TestStringIncludesEveryField(t *testing.T) {
	got := Get().String()

	for _, want := range []string{"version:", "commit:", "built:", "go:", "platform:", "protocol:"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestProtocolIsPositive(t *testing.T) {
	// Protocol 0 would mean "unversioned", which the handshake treats as a
	// hard error. Guard against someone zeroing it during a refactor.
	if Protocol < 1 {
		t.Errorf("Protocol = %d, must be >= 1 (see docs/protocol.md)", Protocol)
	}
}
