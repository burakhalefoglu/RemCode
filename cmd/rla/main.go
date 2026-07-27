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

// Command rla is the RemLinkAgent CLI: it runs the coding agent on the user's
// machine and links it to a paired mobile device.
//
// This is the P0 skeleton. Only `version` and `help` do anything; every other
// command is declared here so the surface is visible, and exits non-zero
// rather than pretending to work. Cobra replaces this dispatch in P0.3.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/burakhalefoglu/RemLinkAgent/internal/version"
)

// exitNotImplemented signals a command that exists in the design but has no
// implementation yet. Distinct from exitUsage so scripts can tell "you typed
// it wrong" apart from "this does not exist yet".
const (
	exitUsage          = 2
	exitNotImplemented = 3
)

// command is a CLI verb. Planned marks the roadmap phase that implements it.
type command struct {
	name    string
	summary string
	phase   string // "" when already implemented
}

var commands = []command{
	{"login", "Store AI provider API keys in the OS keychain", "P1.2"},
	{"logout", "Remove every stored credential from this host", "P1.2"},
	{"connect", "Show a QR code to pair a mobile device", "P1.10"},
	{"status", "Report daemon, relay and paired-device state", "P1.11"},
	{"device", "List or revoke paired devices", "P2.5"},
	{"daemon", "Run the agent daemon in the foreground", "P1.8"},
	{"version", "Print build metadata", ""},
	{"help", "Show this help", ""},
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(exitUsage)
	}

	switch name := args[0]; name {
	case "version", "--version", "-v":
		fmt.Println(version.Get())

	case "help", "--help", "-h":
		usage()

	default:
		for _, c := range commands {
			if c.name != name {
				continue
			}
			// Fail loud: never let a caller believe an unimplemented command
			// succeeded. This invariant outlives the skeleton.
			//
			// SPEC-build-scaffolding-03: a declared-but-absent command exits
			// non-zero and names the phase that will implement it.
			fmt.Fprintf(os.Stderr,
				"rla %s is not implemented yet (planned for %s).\nRoadmap: https://github.com/burakhalefoglu/RemLinkAgent/blob/main/docs/roadmap.md\n",
				c.name, c.phase)
			os.Exit(exitNotImplemented)
		}

		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
		usage()
		os.Exit(exitUsage)
	}
}

func usage() {
	w := flag.CommandLine.Output()

	fmt.Fprintf(w, "rla — RemLinkAgent CLI (%s)\n\n", version.Short())
	fmt.Fprintf(w, "An AI coding agent that runs on this machine and is driven from your phone.\n\n")
	fmt.Fprintf(w, "Usage:\n  rla <command> [flags]\n\nCommands:\n")

	for _, c := range commands {
		status := ""
		if c.phase != "" {
			status = fmt.Sprintf("  (planned: %s)", c.phase)
		}
		fmt.Fprintf(w, "  %-9s %s%s\n", c.name, c.summary, status)
	}

	fmt.Fprintf(w, "\nPre-alpha: only `version` and `help` are implemented.\n")
	fmt.Fprintf(w, "Docs: https://github.com/burakhalefoglu/RemLinkAgent/tree/main/docs\n")
}
