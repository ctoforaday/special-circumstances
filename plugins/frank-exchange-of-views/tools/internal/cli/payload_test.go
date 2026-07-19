package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// THE ESCAPING TAX.
//
// Measured in the 2026-07-18 run: 68 commands carried escaped quotes, 9 used heredocs, and
// 37 staged a temp file first — two of which failed because the staged file was not there.
// Prose into markdown costs nothing; prose through the tool meant fighting the shell, and
// evidence goes wherever it is cheap to put.
//
// `flags.ReadPayload` understood `--file -` all along. `seat.Text` did its own os.ReadFile
// and never called it, so every verb silently lacked stdin while the capability sat one
// package away, written and tested. Two readers of one concept, drifted.

// hostile is the payload a seat actually has to pass: quotes, dollars, apostrophes,
// backticks and a newline — every character that makes shell quoting a hazard.
const hostile = "quotes \"like this\", $vars, 'apostrophes', `backticks`\nand a second line"

func TestPayloadArrivesIntactThroughStdin(t *testing.T) {
	runDir := seatRun(t)
	out, err := runStdin(t, hostile, "lens", "friction", "--run", runDir,
		"--seat-id", "red-lens-r1-L1", "--file", "-")
	if err != nil {
		t.Fatalf("--file - : %v (%s)", err, out)
	}
	if got := lastOfType(t, runDir, "friction").Payload.Str("text"); got != hostile {
		t.Errorf("the payload did not survive stdin.\n got: %q\nwant: %q", got, hostile)
	}
}

// The six verbs whose justification field is genuinely long-form. These are the values a
// seat had to inline and escape, because the only alternative was the markdown.
func TestLongFormFieldsAcceptThePayloadChannel(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "long-form", "payload-channel")
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--text", "o"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, key string
		args      []string
	}{
		{"merge dispose --reason", "reason", []string{"merge", "dispose", "--seat-id", "red-merge-r1", "--observation", "L1-O1", "--as", "declined"}},
		{"merge regrade --basis", "basis", []string{"merge", "regrade", "--seat-id", "red-merge-r1", "--id", id, "--severity", "low"}},
		{"merge dispute-respond --basis", "rationale", []string{"merge", "dispute-respond", "--seat-id", "red-merge-r1", "--id", id, "--as", "accepted"}},
		{"blue dispute --basis", "evidence", []string{"blue", "dispute", "--seat-id", "blue-respond-r1", "--id", id, "--dimension", "severity", "--proposed", "low"}},
		{"merge spot-check --notes", "notes", []string{"merge", "spot-check", "--seat-id", "red-merge-r1", "--ids", id}},
		{"lens petition --basis", "basis", []string{"lens", "petition", "--seat-id", "red-lens-r1-L1", "--petition-class", "safety", "--relief", "halt"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			args := append([]string{c.args[0], c.args[1], "--run", runDir}, c.args[2:]...)
			args = append(args, "--file", "-")
			if out, err := runStdin(t, hostile, args...); err != nil {
				t.Fatalf("%s via stdin: %v (%s)", c.name, err, out)
			}
			ev := lastOfType(t, runDir, c.args[1])
			if got := ev.Payload.Str(c.key); got != hostile {
				t.Errorf("%s did not fill %s from the payload channel.\n got: %q\nwant: %q", c.name, c.key, got, hostile)
			}
		})
	}
}

// Two spellings of one field must be REFUSED, not silently ranked. A seat that passes both
// should be told which one this verb would have dropped, not discover it in a projection
// three rounds later.
func TestBothSpellingsOfOneFieldAreRefused(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--label", "L1-O1", "--kind", "note", "--text", "o"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", "L1-O1", "--as", "declined",
		"--reason", "the named flag", "--text", "the payload channel")
	if err == nil {
		t.Fatal("passing --reason AND --text was accepted; one of them was silently dropped")
	}
	for _, want := range []string{"--reason", "exactly one"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the field and the rule, got: %v", err)
		}
	}
}

// The three verbs that carry only short values want NO payload channel. Symmetry for its
// own sake would give `confidence` a --file with nothing to fill.
func TestShortValueVerbsHaveNoPayloadChannel(t *testing.T) {
	for _, c := range [][2]string{{"lens", "cite"}, {"blue", "confidence"}, {"merge", "verdict"}} {
		if h := help(t, c[0], c[1], "--help"); strings.Contains(h, "--file ") {
			t.Errorf("%s %s grew a payload channel; its fields are a label and a grade, and --file would have nothing to fill", c[0], c[1])
		}
	}
}

// runStdin drives the CLI with a payload on stdin. cobra's InOrStdin() is what the
// resolver reads, so SetIn is enough — no process spawn, no real pipe.
func runStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	r, w, perr := os.Pipe()
	if perr != nil {
		t.Fatal(perr)
	}
	saved := os.Stdout
	os.Stdout = w

	root := newRoot()
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)
	err := root.Execute()

	os.Stdout = saved
	w.Close()
	var buf bytes.Buffer
	if _, cerr := buf.ReadFrom(r); cerr != nil {
		t.Fatal(cerr)
	}
	r.Close()
	return buf.String(), err
}

// ONE CONVENTION FOR STDIN: a `-` where a path goes.
//
// The comment field used to carry a dedicated --comment-stdin boolean while the payload
// spelled the same idea as `--file -`. Two spellings for one idea, in a package whose
// entire purpose is that a word means one thing — and the two functions contradicted each
// other in adjacent lines, RegisterPayload's comment saying stdin "costs no extra flag"
// directly above RegisterComment registering exactly that flag.
func TestCommentReadsStdinThroughTheSameDashConvention(t *testing.T) {
	runDir := seatRun(t)
	if _, err := runStdin(t, hostile, "lens", "friction", "--run", runDir,
		"--seat-id", "red-lens-r1-L1", "--text", "the payload", "--comment-file", "-"); err != nil {
		t.Fatalf("--comment-file -: %v", err)
	}
	ev := lastOfType(t, runDir, "friction")
	if got := ev.Payload.Str("comment"); got != hostile {
		t.Errorf("comment = %q, want the stdin content intact", got)
	}
	if got := ev.Payload.Str("text"); got != "the payload" {
		t.Errorf("the payload was disturbed by the comment reading stdin: %q", got)
	}
}

// THERE IS ONE STDIN. Two fields claiming it used to mean the first reader consumed
// everything and the second reported "stdin was empty" — an error naming the wrong flag,
// which sends a seat hunting a problem with its payload that does not exist.
func TestTwoFieldsCannotBothClaimStdin(t *testing.T) {
	runDir := seatRun(t)
	_, err := runStdin(t, "shared", "lens", "friction", "--run", runDir,
		"--seat-id", "red-lens-r1-L1", "--file", "-", "--comment-file", "-")
	if err == nil {
		t.Fatal("both fields read stdin and neither complained; one of them silently got nothing")
	}
	for _, want := range []string{"--file -", "--comment-file -", "only one"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name BOTH claimants and the reason, got: %v", err)
		}
	}
}
