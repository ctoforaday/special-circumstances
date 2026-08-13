package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
// `flags.ReadPayload` understood `--reason-file -` all along, and every prose verb routes
// its payload through the one `seat.Reason` resolver, so stdin works everywhere for free —
// the capability no longer sits one package away from a second, drifted reader.

// hostile is the payload a seat actually has to pass: quotes, dollars, apostrophes,
// backticks and a newline — every character that makes shell quoting a hazard.
const hostile = "quotes \"like this\", $vars, 'apostrophes', `backticks`\nand a second line"

func TestPayloadArrivesIntactThroughStdin(t *testing.T) {
	runDir := seatRun(t)
	out, err := runStdin(t, hostile, "lens", "friction", "--run", runDir,
		"--seat-id", "red-lens-r1-L1", "--reason-file", "-")
	if err != nil {
		t.Fatalf("--reason-file - : %v (%s)", err, out)
	}
	if got := lastOfType(t, runDir, "friction").Payload.Str("text"); got != hostile {
		t.Errorf("the payload did not survive stdin.\n got: %q\nwant: %q", got, hostile)
	}
}

// The prose verbs whose justification field is genuinely long-form, each now reading it
// through the one --reason / --reason-file channel. These are the values a seat had to
// inline and escape, because the only alternative was the markdown. The payload KEY still
// differs per verb (reason/basis/rationale/evidence) — the WORD collapsed to --reason, the
// schema did not.
func TestLongFormFieldsAcceptThePayloadChannel(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "long-form", "payload-channel")
	// The STATE each verb needs, not just the referent. A ruling answers a motion, so M1 is
	// filed for the rule case to answer; the file case contests a DIFFERENT gap.
	undisputed := mintGap(t, runDir, "undisputed", "payload-channel")
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", id, "--dimension", "severity", "--proposed", "low", "--reason", "b"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, key string
		// typ is the event type to look for. It used to be args[1] — the verb word, read back
		// out of the argv the case had just composed. That works only while every command path
		// is <role> <verb> AND the verb word equals the event type; `motion grade file` breaks
		// both halves at once, and the failure was "no grade event in the log".
		typ  string
		args []string
	}{
		{"merge regrade", "basis", "regrade", []string{"merge", "regrade", "--seat-id", "red-merge-r1", "--id", id, "--severity", "low"}},
		{"motion grade rule", "opinion", "motion-rule", []string{"motion", "grade", "rule", "--seat-id", "red-merge-r1", "--id", "M1", "--as", "accepted"}},
		{"motion grade file", "basis", "motion", []string{"motion", "grade", "file", "--seat-id", "blue-respond-r1", "--id", undisputed, "--dimension", "severity", "--proposed", "low"}},
		{"motion petition file", "basis", "motion", []string{"motion", "petition", "file", "--seat-id", "red-merge-r1", "--petition-class", "safety", "--relief", "halt"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			// The path is however many leading non-flag words the case supplies.
			split := 0
			for split < len(c.args) && !strings.HasPrefix(c.args[split], "-") {
				split++
			}
			args := append(append([]string{}, c.args[:split]...), "--run", runDir)
			args = append(args, c.args[split:]...)
			args = append(args, "--reason-file", "-")
			if out, err := runStdin(t, hostile, args...); err != nil {
				t.Fatalf("%s via stdin: %v (%s)", c.name, err, out)
			}
			ev := lastOfType(t, runDir, c.typ)
			if got := ev.Payload.Str(c.key); got != hostile {
				t.Errorf("%s did not fill %s from the payload channel.\n got: %q\nwant: %q", c.name, c.key, got, hostile)
			}
		})
	}
}

// The two forms of the ONE prose field — --reason and --reason-file — must be REFUSED
// together, not silently ranked. A seat that passes both should be told which one this
// verb would have dropped, not discover it in a projection three rounds later.
func TestBothSpellingsOfOneFieldAreRefused(t *testing.T) {
	runDir := seatRun(t)
	// Driven through `merge position` since #327 retired `dispose`, which this used to use.
	// The rule is the seat.Prose contract's, not any one verb's — any prose verb proves it.
	both := filepath.Join(t.TempDir(), "prose.md")
	if werr := os.WriteFile(both, []byte("from a file"), 0o644); werr != nil {
		t.Fatal(werr)
	}
	_, err := run(t, "merge", "position", "--run", runDir, "--seat-id", "red-merge-r1",
		"--reason", "inline", "--reason-file", both)
	if err == nil {
		t.Fatal("passing --reason AND --reason-file was accepted; one of them was silently dropped")
	}
	for _, want := range []string{"--reason", "exactly one"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must name the field and the rule, got: %v", err)
		}
	}
}

// A verb that carries only short values wants NO payload channel — symmetry for its own sake
// would hand it a --reason with nothing to fill.
//
// `verify` WAS on this list, and the entry was load-bearing in the wrong direction: it recorded
// the belief that a verification is a label and a grade. It is not. It is a judgement about what
// a source says, and the judgement was the part that never reached the record — the verb accepted
// no flags at all, so a bare `lens verify` appended an event and counted as audit volume. It now
// requires the reading behind its verdict, which is exactly the payload channel this test used to
// forbid it.
func TestShortValueVerbsHaveNoPayloadChannel(t *testing.T) {
	for _, c := range [][2]string{{"merge", "verdict"}} {
		if h := help(t, c[0], c[1], "--help"); strings.Contains(h, "--reason ") {
			t.Errorf("%s %s grew a payload channel; its fields are a label and a grade, and --reason would have nothing to fill", c[0], c[1])
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

// ONE CONVENTION FOR STDIN: a `-` where a path goes. --reason-file - reads stdin, and it
// is the only reader — the universal --comment field (and its second stdin claimant) was
// retired in the 2026-07-20 vocabulary collapse, so the two-fields-on-one-stdin conflict it
// used to create no longer exists.
func TestReasonFileReadsStdinThroughTheDashConvention(t *testing.T) {
	runDir := seatRun(t)
	if _, err := runStdin(t, hostile, "lens", "friction", "--run", runDir,
		"--seat-id", "red-lens-r1-L1", "--reason-file", "-"); err != nil {
		t.Fatalf("--reason-file -: %v", err)
	}
	ev := lastOfType(t, runDir, "friction")
	if got := ev.Payload.Str("text"); got != hostile {
		t.Errorf("text = %q, want the stdin content intact", got)
	}
}
