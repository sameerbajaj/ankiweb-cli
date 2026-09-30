//go:build !windows

// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.


package testenv

import (
	"os"
	"testing"
)

func lockSandbox(t *testing.T, path string) {
	t.Helper()
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatalf("chmod sandbox %s: %v", path, err)
	}
}
