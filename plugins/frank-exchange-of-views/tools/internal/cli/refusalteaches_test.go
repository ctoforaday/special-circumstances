package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A REFUSAL TEACHES THE WHOLE SURFACE OR NAMES `--help`. It never hands out a bare list.
//
// # The rule, and why it is a rule
//
// MEASURED across nine probe sittings: seats do not learn this tool from `--help`. Every one of
// them read it once or twice in twenty to forty tool calls. They learn it from REFUSALS — a seat
// that guesses a verb is told it does not exist, and that is the moment it finds out what it has.
//
// So the refusal is the primary teaching channel, and it was the LOSSY one. `blue --help` gives
// every verb with what it does and its flags; the refusal gave fourteen bare names. A seat learns
// WHAT EXISTS and never WHAT IT IS FOR, which is precisely the measured behaviour: seats use the
// verbs they can name, in the ways they already assumed.
//
// A truncated list is worse than no list, because it reads as complete. A seat that has seen
// "available: register, edit, cite, …" stops looking, and everything it was not told about is a
// capability the run silently does without.
//
// # What counts as satisfying it
//
// Either the message carries the real help — which for a command means cobra's own, since that is
// generated from the tree and carries the descriptions and the enumerated-value menus — or it
// names `--help` and stops. What it may not do is enumerate names with no meanings.

// bareList matches three or more comma-separated bare words inside a parenthesis or after a
// colon — the shape of a hand-rolled "available: a, b, c" with nothing else said about any of them.
var bareList = regexp.MustCompile(`(?:available|one of|it has|yours are)[:]?\s*\(?\s*[a-z][a-z-]+(?:,\s*[a-z][a-z-]+){2,}`)

// refusals are the messages a seat meets when it gets something wrong. Each is (what it did,
// the command).
func refusals() []struct {
	name string
	args []string
} {
	return []struct {
		name string
		args []string
	}{
		{"a verb refusing on a missing required flag", []string{"blue", "line-of-inquiry", "propose", "--seat-id", "blue-respond-r1"}},
		{"a group invoked with no verb", []string{"blue", "line-of-inquiry", "--seat-id", "blue-respond-r1"}},
		{"a role with no verb", []string{"blue"}},
		{"an unknown top-level command", []string{"frobnicate"}},
		{"a motion subject with no such verb", []string{"motion", "petition", "appeal", "--seat-id", "blue-respond-r1"}},
		{"an unknown view", []string{"blue", "show", "nonesuch", "--seat-id", "blue-respond-r1"}},
	}
}

func TestNoRefusalHandsOutABareList(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := t.TempDir()

	for _, r := range refusals() {
		t.Run(r.name, func(t *testing.T) {
			args := append(append([]string{}, r.args...), "--run", runDir)
			out, err := run(t, args...)
			// stdout carries the help when the refusal printed it; the error carries the rest.
			text := out
			if err != nil {
				text += "\n" + err.Error()
			}
			if strings.TrimSpace(text) == "" {
				t.Fatalf("the refusal said NOTHING — a seat that gets silence learns nothing and retries the same thing")
			}

			if m := bareList.FindString(text); m != "" && !teaches(text) {
				t.Errorf("this refusal hands out a bare list and teaches nothing:\n\n  %s\n\nGive the seat the real help (cobra's own, which carries the descriptions and the enumerated-value menus) or name `--help` and stop. A truncated list reads as complete, so a seat that has seen it stops looking — and everything it was not told about becomes a capability the run silently does without.", m)
			}
			if !teaches(text) && !strings.Contains(text, "--help") {
				t.Errorf("this refusal neither carries the help nor names `--help`:\n\n%s\n\nMEASURED: seats learn this tool from refusals, not from help they were told to read. A refusal that teaches nothing is a dead end at the exact moment a seat is looking.", trim(text))
			}
		})
	}
}

// teaches reports whether the text carries real help: cobra's own rendering, or a menu with
// meanings beside the names.
func teaches(text string) bool {
	switch {
	case strings.Contains(text, "Available Commands:"):
		return true // cobra's help, which carries every Short
	case strings.Contains(text, "Enumerated values:"):
		return true // the enum menu
	case strings.Contains(text, "Usage:"):
		return true
	}
	// A per-line menu: at least three lines that are "  name  something meaningful".
	lines := 0
	for _, l := range strings.Split(text, "\n") {
		if m := regexp.MustCompile(`^\s{2,}[a-z][a-z-]+\s{2,}\S.{15,}`).FindString(l); m != "" {
			lines++
		}
	}
	return lines >= 3
}

func trim(s string) string {
	if len(s) > 600 {
		return s[:600] + "\n…"
	}
	return s
}

// THE WRONG COMMAND IS ANSWERED BEFORE THE WRONG FLAG.
//
// Cobra parses flags before deciding a command is unknown, so `show board` reported
// `unknown flag: --view` — naming the one thing the caller had right, and sending a seat hunting
// through view names instead of learning the role prefix it was missing.
//
// This is the most common failure a real seat hits. Across nine probed seats: 9 of 21 non-zero
// exits were this message, and SIX OF NINE produced it on their first tool call.
func TestAnUnknownCommandIsNamedBeforeAnyFlagOnIt(t *testing.T) {
	root := newRoot()
	for _, tc := range []struct {
		name string
		argv []string
		want string
	}{
		// The measured slip: `show` is per-role, and the concept "show the board" carries no role.
		//
		// The wanted text is `"show" is not a top-level command`, NOT `no command named "show"`.
		// This case used to assert the second, which was the message before the refusal learned to
		// say WHERE the verb lives — and `show` is exactly the name that must not be denied, since
		// it exists on all four roles. TestABareSeatVerbIsToldWhereItLives holds that half; this
		// one holds the ORDERING (command before flag), which is a different property.
		{"real flags on a command that does not exist", []string{"feov-record", "show", "--run", "/x", "board"}, `"show" is not a top-level command`},
		{"no flags at all", []string{"feov-record", "show"}, `"show" is not a top-level command`},
		{"a command that does exist is left to cobra", []string{"feov-record", "blue", "--nonsuch"}, ""},
		{"a bare flag is left to cobra", []string{"feov-record", "--version"}, ""},
		{"help is left to cobra", []string{"feov-record", "--help"}, ""},
		{"no arguments at all", []string{"feov-record"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := refuseUnknownCommandFirst(root, tc.argv)
			switch {
			case tc.want == "" && err != nil:
				t.Fatalf("refused %v, which it must leave alone: %v", tc.argv[1:], err)
			case tc.want == "":
				return
			case err == nil:
				t.Fatalf("%v was not refused — cobra answers it with a flag error naming the "+
					"one part the caller had right", tc.argv[1:])
			case !strings.Contains(err.Error(), tc.want):
				t.Fatalf("refusal does not name the command: %v", err)
			}
			// The refusal must carry the surface, or it teaches nothing: this is the moment a
			// seat is definitively looking for what exists.
			if !strings.Contains(err.Error(), "blue") || !strings.Contains(err.Error(), "merge") {
				t.Errorf("the refusal does not list the commands that DO exist:\n%v", err)
			}
		})
	}
}

// The tool names itself by what it was actually invoked as. It used to claim "feov-record"
// unconditionally — a statement about the filesystem it is in no position to make, and the seat
// probe builds the binary under another name, so every refusal a probed seat read named a
// program that did not exist on that machine.
func TestTheToolNamesItselfByArgv0(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	// PATHS ARE BUILT NATIVE, because filepath.Base splits on the separator of the platform it
	// runs on — and it is RIGHT not to split a backslash on Linux, where that is an ordinary
	// character in a filename. A literal Windows path in this table passed on Windows and failed
	// on Linux: the test being wrong about portability, not the code being wrong about paths.
	for _, tc := range []struct{ argv0, want string }{
		{filepath.Join("C:", "tools", "fxr.exe"), "fxr"},
		{filepath.Join("usr", "local", "bin", "fxr"), "fxr"},
		{filepath.Join("opt", "bin", "feov-record"), "feov-record"},
		{filepath.Join("p", "feov-record.EXE"), "feov-record"}, // the extension match is case-insensitive
		{"feov-record", "feov-record"},                         // bare, no directory at all
		{"", "feov-record"},                                    // argv[0] can be empty; a name beats none
	} {
		os.Args = []string{tc.argv0}
		if got := InvokedAs(); got != tc.want {
			t.Errorf("InvokedAs() for %q = %q, want %q", tc.argv0, got, tc.want)
		}
	}
}

// A SEAT VERB TYPED WITHOUT ITS ROLE IS A WRONG ADDRESS, NOT A MISSING CAPABILITY.
//
// MEASURED 2026-08-17 on an elicitation probe. A red-merge seat holding a work list duty that named
// `inquiry-support` typed `feov-record inquiry-support --help` and was told
//
//	no command named "inquiry-support" exists
//
// It believed that — the message is unambiguous — went hunting, and settled on `motion inquiry
// rule`, which is a DIFFERENT ACT: whether the direction is worth the run's time, not whether the
// report still carries it. The seat did not lose a turn; it landed on the wrong verb confidently.
//
// root.go already documents the slip ("`show` is per-role while the concept carries no role, so
// dropping the prefix is the natural slip rather than a careless one", 6 of 9 seats on their FIRST
// call) and already routes it to the right refusal. The refusal itself still denied the verb
// existed. The role-level message had the right shape all along — "it exists, on the blue seat, but
// not for you. That is a wrong-seat error rather than a missing capability" — and the ROOT level,
// which is where a seat dropping the prefix actually lands, said the opposite.
//
// The genuinely-absent case must keep saying so, or this trades one false answer for another.
func TestABareSeatVerbIsToldWhereItLives(t *testing.T) {
	root := newRoot()

	for _, tc := range []struct{ name, wantPath string }{
		{"inquiry-support", "merge inquiry-support"},
		{"close", "merge close"},
		{"line-of-inquiry", "blue line-of-inquiry"},
		{"verify", "lens verify"},
	} {
		got := unknownCommandRefusal(root, tc.name)
		if !strings.Contains(got, tc.wantPath) {
			t.Errorf("a bare %q does not say where it lives (want %q):\n%s\n\n"+
				"A seat that believes the verb does not exist stops looking for it — measured: one went on to "+
				"use a different verb entirely.", tc.name, tc.wantPath, got)
		}
		if strings.Contains(got, "no command named") {
			t.Errorf("a bare %q is refused as nonexistent when it exists at %q:\n%s", tc.name, tc.wantPath, got)
		}
	}

	// AND A NAME THAT REALLY IS ABSENT MUST STILL SAY SO.
	absent := unknownCommandRefusal(root, "banana")
	if !strings.Contains(absent, "no command named") {
		t.Errorf("a genuinely absent command no longer says it is absent:\n%s\n\n"+
			"Pointing a seat at a verb that does not exist is the same defect facing the other way.", absent)
	}
}
