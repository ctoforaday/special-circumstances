package hookmain

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestAnOrdinaryInvocationIsNotFinished(t *testing.T) {
	var out, errb bytes.Buffer
	if Preamble(nil, &out, &errb, Named("sc-thing")) {
		t.Error("a plain invocation reported itself finished")
	}
	if out.Len() != 0 {
		t.Errorf("wrote to stdout on a plain invocation: %q", out.String())
	}
}

// The version line names the binary that was INVOKED. This is the assertion the whole
// package exists for: three shims sharing one run() answered with the merged binary's
// retired name, and sc-doctor's table showed three rows with one name.
func TestVersionNamesTheBinaryAndFinishes(t *testing.T) {
	var out, errb bytes.Buffer
	if !Preamble([]string{"-version"}, &out, &errb, Named("sc-precompact")) {
		t.Fatal("-version did not finish the invocation")
	}
	if got := strings.Fields(out.String()); len(got) == 0 || got[0] != "sc-precompact" {
		t.Errorf("version line = %q, want it to start with sc-precompact", out.String())
	}
}

// THE CASE THAT BROKE BEFORE: a binary whose name is not knowable until a flag has been
// parsed. resolve() runs after parsing, so the version line can name the shim rather than
// the package.
//
// An API serving only the compile-time case would have left this caller writing the code
// that produced the defect — which is why resolve is a function and not a string.
func TestTheNameMayDependOnAFlagThatWasJustParsed(t *testing.T) {
	for _, tc := range []struct{ event, want string }{
		{"PreCompact", "sc-precompact"},
		{"SessionEnd", "sc-sessionend"},
		{"SubagentStop", "sc-subagentstop"},
	} {
		t.Run(tc.event, func(t *testing.T) {
			var out, errb bytes.Buffer
			var event *string
			resolve := func() string {
				switch {
				case event == nil || *event == "":
					return "sc-checkpoint-seal"
				case *event == "PreCompact":
					return "sc-precompact"
				case *event == "SessionEnd":
					return "sc-sessionend"
				}
				return "sc-subagentstop"
			}
			finished := Preamble([]string{"-event", tc.event, "-version"}, &out, &errb, resolve,
				func(fs *flag.FlagSet) { event = fs.String("event", "", "the event this invocation serves") })
			if !finished {
				t.Fatal("-version did not finish")
			}
			if got := strings.Fields(out.String()); len(got) == 0 || got[0] != tc.want {
				t.Errorf("version line = %q, want it to start with %s", out.String(), tc.want)
			}
		})
	}
}

// A declared flag is actually usable by the caller afterwards — the point of `declare` is
// that packages with extra flags do not have to opt out of the shared preamble.
func TestDeclaredFlagsReachTheCaller(t *testing.T) {
	var out, errb bytes.Buffer
	var event *string
	if Preamble([]string{"-event", "SessionEnd"}, &out, &errb, Named("sc-sessionend"),
		func(fs *flag.FlagSet) { event = fs.String("event", "", "") }) {
		t.Fatal("an ordinary invocation with a declared flag reported itself finished")
	}
	if event == nil || *event != "SessionEnd" {
		t.Errorf("declared flag did not reach the caller: %v", event)
	}
}

// A BAD FLAG IS NEVER WORTH FAILING THE EVENT OVER. Every caller's response to a parse
// error was already `return 0`; making that the shared behaviour removes the last place
// one could quietly decide otherwise — and a hook that exits non-zero can deny a tool call.
func TestABadFlagFinishesQuietlyRatherThanFailing(t *testing.T) {
	var out, errb bytes.Buffer
	if !Preamble([]string{"-nonsense"}, &out, &errb, Named("sc-thing")) {
		t.Error("a bad flag did not finish the invocation")
	}
	if out.Len() != 0 {
		t.Errorf("a bad flag wrote to STDOUT (%q); stdout is the client's channel", out.String())
	}
	if errb.Len() == 0 {
		t.Error("a bad flag said nothing on stderr; the operator gets no signal at all")
	}
}
