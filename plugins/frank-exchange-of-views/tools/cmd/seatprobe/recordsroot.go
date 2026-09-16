package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// THE PROBE'S RECORDS ARE SCRATCH, AND SCRATCH DOES NOT LIVE IN /tmp.
//
// The root is deliberately never deleted (see the dispatch path), which is exactly the case the
// repository's scratch rule is about: run directories, built binaries and measurement artifacts
// belong under a home-dir scratch area, because a kept artifact in the system temp directory is
// one a reboot or a tmpfiles sweep removes between the run and the reading of it.
func recordsBase(home string) string {
	return filepath.Join(home, ".claude", "scratch", "seatprobe")
}

// newRecordsRootIn creates a fresh, kept record root under home's scratch area.
func newRecordsRootIn(home string) (string, error) {
	base := recordsBase(home)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("no scratch area for the record root: %w", err)
	}
	return os.MkdirTemp(base, "records-")
}

func newRecordsRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home directory to root the probe's records in: %w", err)
	}
	return newRecordsRootIn(home)
}
