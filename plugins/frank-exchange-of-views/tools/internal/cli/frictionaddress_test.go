package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE CHANNEL FOR REPORTING THAT A CAPABILITY IS UNREACHABLE WAS ITSELF UNREACHABLE.
//
// `friction` at the root is the OPERATOR's read. The seat's write is `<role> friction`. All four
// constitutions taught the write as `friction --reason "<what blocked you>"` — no role in front —
// and that form lands on the read, where cobra rejected it at PARSE time: `unknown flag:
// --reason`, exit 2, before RunE and before the Long text one line up that says seats write
// `<role> friction`. The seat learned which flag was wrong, never that it was at the wrong
// address, and the thing it could not reach was the channel for reporting exactly this.
//
// Eighteen probed sittings recorded no friction at all. That was read as seats declining a duty.
// A seat that typed the form its own constitution taught got a parse error and no second guess.

func TestRootFrictionTeachesTheSeatItsAddress(t *testing.T) {
	for _, args := range [][]string{
		{"friction", "--reason", "the tool has no path for X"},
		{"friction", "--none", "--reason", "nothing blocked me"},
		{"friction", "--reason-file", "/tmp/x"},
		// An EMPTY reason is a seat at the wrong address just as much as a full one, which is
		// why the check keys on Changed rather than on emptiness.
		{"friction", "--reason", ""},
	} {
		out, err := run(t, append(args, "--run", t.TempDir())...)
		if err == nil {
			t.Errorf("%v was accepted by the operator's read", args)
			continue
		}
		if strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("%v still fails at the parser (%v) — the seat is told which flag is wrong, not where the verb lives", args, err)
			continue
		}
		// The refusal must carry the ADDRESS. "This takes no --reason" alone leaves the seat
		// exactly as stuck as the parse error did.
		all := out + err.Error()
		if !strings.Contains(all, "friction --reason") || !strings.Contains(all, "role") {
			t.Errorf("%v refused without naming `<role> friction`:\n%v", args, err)
		}
	}
}

// The operator's read is what this command IS. A refusal that also broke it would trade one
// unreachable channel for another.
func TestTheOperatorFrictionReadStillWorks(t *testing.T) {
	out, err := run(t, "friction", "--run", t.TempDir())
	if err != nil {
		t.Fatalf("the operator's friction read failed: %v", err)
	}
	if !strings.Contains(out, "friction") || !strings.Contains(out, "nothing_blocked") {
		t.Errorf("the read did not return the friction projection:\n%s", out)
	}
}

// NO CONSTITUTION MAY TEACH THE ROLELESS FORM AGAIN. The refusal above makes the mistake
// survivable; this makes it not happen. A backticked `friction --reason` in a seat's own
// constitution is an instruction to run a command that exits 2.
func TestNoConstitutionTeachesTheRolelessFrictionForm(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("..", "..", "..", "agents", "*.md"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("found no constitutions to check (%v) — a scan that reads nothing reports every file clean", err)
	}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, bad := range []string{"`friction --reason", "`friction --none"} {
			if strings.Contains(string(b), bad) {
				t.Errorf("%s teaches %s… — that form lands on the operator's read and exits 2; name the role, or state the duty and let the engine's prompt carry the invocation", filepath.Base(p), bad)
			}
		}
	}
}
