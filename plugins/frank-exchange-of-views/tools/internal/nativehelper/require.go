package nativehelper

import "os"

// TB is the part of testing.TB that Require uses; it keeps package testing out of the shipped
// binaries that import this package.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
	Skipf(format string, args ...any)
	Cleanup(func())
}

// Require is a running helper for a test, or the test's end:
//
//   - NATIVE_HELPER_DIR unset under GITHUB_ACTIONS: the test FAILS. CI never skips for want of a
//     helper.
//   - NATIVE_HELPER_DIR=none: the test skips, saying so.
//   - NATIVE_HELPER_DIR unset outside CI: the test skips, saying so.
//   - Otherwise the helper must be found there and pass its handshake, or the test fails; a
//     mistyped directory, or the variable set to the empty string, is a failure, never a skip.
//
// The client closes when the test ends.
func Require(t TB, helper string) *Client {
	t.Helper()
	pins, ok := helpers[helper]
	if !ok {
		t.Fatalf("nativehelper.Require: this binary carries no pins for a helper named %q; native/ holds no such helper", helper)
		return nil
	}
	return require(t, helper, pins)
}

func require(t TB, name string, pins Pins) *Client {
	t.Helper()
	switch dir, set := os.LookupEnv(EnvDir); {
	case !set && os.Getenv("GITHUB_ACTIONS") != "":
		t.Fatalf("%s is unset under GITHUB_ACTIONS: CI runs this test against a built %s. Set %s to the directory holding it, or to none to run without one.",
			EnvDir, name, EnvDir)
		return nil
	case !set:
		t.Skipf("SKIPPED, NOT PASSED: %s is unset, so this test has no %s. Set %s to the directory holding a built %s to run it.",
			EnvDir, name, EnvDir, name)
		return nil
	case dir == "none":
		t.Skipf("SKIPPED, NOT PASSED: %s=none runs without %s.", EnvDir, name)
		return nil
	}
	c := newClient(name, pins)
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Start(); err != nil {
		t.Fatalf("nativehelper.Require: %v", err)
		return nil
	}
	return c
}
