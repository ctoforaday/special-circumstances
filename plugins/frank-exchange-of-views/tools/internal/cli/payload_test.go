package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// THE ESCAPING TAX.
//
// Measured in the 2026-07-18 run: 68 commands carried escaped quotes, 9 used heredocs, and
// 37 staged a temp file first — two of which failed because the staged file was not there.
// Prose into markdown costs nothing; prose through the tool meant fighting the shell, and
// evidence goes wherever it is cheap to put.
//
// The prose channel is ONE spelling now, --reason, and every prose verb routes its payload
// through the one `seat.Reason` resolver. The shell half of the fight is answered by the quoting
// rule in every prose verb's help (flags.ProseFooter), not by a second spelling of the flag.

// hostile is the payload a seat actually has to pass: quotes, dollars, apostrophes,
// backticks and a newline — every character that makes shell quoting a hazard. Argv in a Go
// test does not pass through a shell, so what arrives here is what the tool must record intact.
const hostile = "quotes \"like this\", $vars, 'apostrophes', `backticks`\nand a second line"

func TestPayloadArrivesIntactThroughReason(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "log", "--run", runDir,
		"--seat-id", "red-lens-evidence", "--type", "defect", "--reason", hostile)
	if err != nil {
		t.Fatalf("--reason: %v (%s)", err, out)
	}
	if got := lastBody(t, runDir, &recordpb.Log{}).GetText(); got != hostile {
		t.Errorf("the payload did not survive the channel.\n got: %q\nwant: %q", got, hostile)
	}
}

// The prose verbs whose justification field is genuinely long-form, each reading it through
// the one --reason channel. The payload KEY still differs per verb (reason/basis/rationale/
// evidence) — the WORD collapsed to --reason, the schema did not.
func TestLongFormFieldsAcceptThePayloadChannel(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "long-form", "payload-channel")
	// The STATE each verb needs, not just the referent. A ruling answers a motion, so M1 is
	// filed for the rule case to answer; the file case contests a DIFFERENT gap.
	undisputed := mintGap(t, runDir, "undisputed", "payload-channel")
	if _, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", "blue-respond",
		"--id", id, "--dimension", "severity", "--proposed", "low", "--reason", "b"); err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name, field string
		// typ is the event type to look for. It used to be args[1] — the verb word, read back
		// out of the argv the case had just composed. That works only while every command path
		// is <role> <verb> AND the verb word equals the event type; `motion grade file` breaks
		// both halves at once, and the failure was "no grade event in the log".
		typ  recordpb.EventType
		args []string
	}{
		// NO ROLE SEGMENT: the surface is the seat's, so `regrade` sits at the root of the lens
		// tree. `motion …` below keeps its path because motion is a real subgroup within it.
		{"lens regrade", "basis", recordpb.EventType_EVENT_TYPE_REGRADE, []string{"regrade", "--seat-id", lensSeat, "--id", id, "--severity", "low"}},
		{"motion grade rule", "opinion", recordpb.EventType_EVENT_TYPE_MOTION_RULE, []string{"motion", "grade", "rule", "--seat-id", "red-chair", "--id", "M1", "--as", "accepted"}},
		{"motion grade file", "basis", recordpb.EventType_EVENT_TYPE_MOTION, []string{"motion", "grade", "file", "--seat-id", "blue-respond", "--id", undisputed, "--dimension", "severity", "--proposed", "low"}},
		{"motion petition file", "basis", recordpb.EventType_EVENT_TYPE_MOTION, []string{"motion", "petition", "file", "--seat-id", "red-chair", "--class", "safety", "--relief", "halt"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			// The path is however many leading non-flag words the case supplies.
			split := 0
			for split < len(c.args) && !strings.HasPrefix(c.args[split], "-") {
				split++
			}
			args := append(append([]string{}, c.args[:split]...), "--run", runDir)
			args = append(args, c.args[split:]...)
			args = append(args, "--reason", hostile)
			if out, err := run(t, args...); err != nil {
				t.Fatalf("%s via --reason: %v (%s)", c.name, err, out)
			}
			// THE FIELD, NOT THE FLAG. `--reason` is what a seat types; the field it lands in is
			// spelled per verb (a regrade stores `basis`, a ruling `opinion`), which is exactly
			// the fold flags.ForPayloadKey exists for.
			ev := lastOfType(t, runDir, c.typ)
			body, ok := recordpb.Body(ev)
			if !ok {
				t.Fatalf("%s wrote an event with no body", c.name)
			}
			m := body.ProtoReflect()
			fd := m.Descriptor().Fields().ByName(protoreflect.Name(c.field))
			if fd == nil {
				t.Fatalf("%s: %s has no field %q", c.name, m.Descriptor().FullName(), c.field)
			}
			if got := m.Get(fd).String(); got != hostile {
				t.Errorf("%s did not fill %s from the prose channel.\n got: %q\nwant: %q", c.name, c.field, got, hostile)
			}
		})
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
	// (verb, a seat that holds it) — the role that used to precede the verb is now the identity
	// that selects the tree it is found in.
	for _, c := range [][2]string{{"verdict", "red-chair"}} {
		if h := help(t, c[0], "--help", "--seat-id", c[1]); strings.Contains(h, "--reason ") {
			t.Errorf("%s grew a payload channel; its fields are a label and a grade, and --reason would have nothing to fill", c[0])
		}
	}
}
