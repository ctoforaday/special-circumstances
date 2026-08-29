package testbuild

import (
	"os"
	"testing"
)

// This package calls Binary from its own tests — UNQUALIFIED, which is how it escaped the
// `testbuild.Binary` sweep that found the other four callers, and why its own guard refused
// the suite until this file existed. Run removes the build directory (#643).
func TestMain(m *testing.M) { os.Exit(Run(m)) }
