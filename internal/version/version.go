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

// Package version exposes build metadata stamped in at link time.
//
// Values are injected by the build via -ldflags -X (see the Makefile). A
// binary built with a plain `go build` reports the "dev" defaults rather than
// failing — but release artefacts must always carry real values, because bug
// reports are unactionable without them.
package version

import (
	"fmt"
	"runtime"
)

// Stamped by -ldflags at build time. Unexported so nothing can mutate them at
// runtime; read through the accessors below.
var (
	version = "0.0.0-dev"
	commit  = "unknown"
	date    = "unknown"
)

// Info is a snapshot of the build metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`

	// Protocol is the wire protocol version this build speaks. It is
	// deliberately independent of Version: the CLI, the server and the mobile
	// app all ship on their own cadence and must negotiate compatibility on
	// the protocol number alone. See docs/protocol.md#versioning.
	Protocol int `json:"protocol"`
}

// Protocol is the wire protocol version implemented by this build.
const Protocol = 1

// Get returns the current build metadata.
func Get() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Protocol:  Protocol,
	}
}

// Short returns just the version string, for `--version` style output.
func Short() string { return version }

// String renders the full metadata as a human-readable multi-line block.
func (i Info) String() string {
	return fmt.Sprintf(
		"version:  %s\ncommit:   %s\nbuilt:    %s\ngo:       %s\nplatform: %s\nprotocol: v%d",
		i.Version, i.Commit, i.Date, i.GoVersion, i.Platform, i.Protocol,
	)
}
