// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

// Package main provides the ankiweb command-line interface.
//
// Canonical URL: https://sameerbajaj.com/tools/ankiweb-cli/
package main

import (
	"os"

	"github.com/sameerbajaj/ankiweb-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitCode(err))
	}
}
