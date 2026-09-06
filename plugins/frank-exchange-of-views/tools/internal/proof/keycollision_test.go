package proof

import (
	"os"
	"path/filepath"
	"testing"
)

// A KEY IS ONE PROGRAM'S NAME, AND ScriptSha IS WHAT LETS prove SAY SO.
//
// Crash-retry and key collision are indistinguishable at the key and are opposite facts: one is
// the same program arriving twice, the other a DIFFERENT program under a key an earlier dispatch
// already burned. The second used to take the retry path — "already recorded as <sha>", recording
// nothing — so a seat that believed the message shipped a report claiming a proof it does not
// hold. Measured in research/2026-09-02_quadratic-formula, blue-respond-r1, --key P1.
func TestScriptShaDistinguishesARetryFromACollision(t *testing.T) {
	run := t.TempDir()
	dir := filepath.Join(run, "blue", "candidates")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(dir, "a.py")
	b := filepath.Join(dir, "b.py")
	if err := os.WriteFile(a, []byte("print(1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("print(2)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	shaA, err := ScriptSha(run, a)
	if err != nil {
		t.Fatal(err)
	}
	// The same bytes hash the same however they arrive — that is the retry, and it must stay
	// cheap: no execution, just a read.
	again, err := ScriptSha(run, a)
	if err != nil {
		t.Fatal(err)
	}
	if shaA != again {
		t.Errorf("the same script hashed differently on a second read: %s vs %s", shaA, again)
	}
	shaB, err := ScriptSha(run, b)
	if err != nil {
		t.Fatal(err)
	}
	if shaA == shaB {
		t.Fatal("two different programs hash the same — a collision would be invisible")
	}
}

// It must not need the script to RUN. A collision is decided before any execution, which is what
// makes refusing it cheaper than the wrong answer.
func TestScriptShaDoesNotExecute(t *testing.T) {
	run := t.TempDir()
	script := filepath.Join(run, "boom.py")
	if err := os.WriteFile(script, []byte("import sys\nsys.exit(3)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ScriptSha(run, script); err != nil {
		t.Errorf("hashing a script that would exit non-zero failed: %v", err)
	}
}
