package main

import (
	"os"
	"testing"
)

// setupTestDir creates a temp directory and changes to it, returning cleanup function
func setupTestDir(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()
	oldWd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}
	return dir, func() { os.Chdir(oldWd) }
}
