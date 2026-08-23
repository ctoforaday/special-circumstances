package diagnostics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// A trajectory, written the way the CLI writes one, so the test exercises the real parse rather
// than a convenient shape.
func traj(t *testing.T, steps ...any) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.jsonl")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, s := range steps {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		f.Write(append(b, '\n'))
	}
	return p
}

func use(id, cmd string) any {
	return map[string]any{"message": map[string]any{"content": []any{
		map[string]any{"type": "tool_use", "name": "Bash", "id": id,
			"input": map[string]any{"command": cmd}}}}}
}

func res(id, text string, isErr bool) any {
	return map[string]any{"message": map[string]any{"content": []any{
		map[string]any{"type": "tool_result", "tool_use_id": id, "is_error": isErr, "content": text}}}}
}

const rootHelp = `feov-record — red lens seats

Available Commands:
  finding     record a finding
  verify      adjudicate one citation
  fetch       fetch a source

Flags:
  -h, --help
`

const findingHelp = `record a finding

Flags:
      --quote string   REQUIRED
`

// THE CENTRAL DISTINCTION: knowing a verb EXISTS is not knowing how to run it. The root listing
// names every verb and gives no flags, so a first use backed only by it is blind — and that is
// exactly the state in which a seat guesses `--sha256` and is refused.
func TestTheRootListingIsNotHavingReadTheHelp(t *testing.T) {
	p := traj(t,
		use("a", `"/tmp/x/feov-record" --seat-id red-lens-r1-L1 --help`), res("a", rootHelp, false),
		use("b", `"/tmp/x/feov-record" fetch --url http://x --sha256 deadbeef`), res("b", "unknown flag: --sha256", true),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Asked {
		t.Fatal("the seat invoked a command and the survey says it asked nothing")
	}
	if len(s.FirstUses) != 1 || s.FirstUses[0].Command != "fetch" {
		t.Fatalf("first uses = %+v, want one for fetch", s.FirstUses)
	}
	if got := s.FirstUses[0].Depth; got != DepthTop {
		t.Errorf("depth = %q, want %q — the root listing names fetch and gives no flags", got, DepthTop)
	}
	if s.BlindFirst() != 1 {
		t.Errorf("blind-first = %d, want 1", s.BlindFirst())
	}
	if b, r := s.RefusedBlind(); b != 1 || r != 0 {
		t.Errorf("refusals blind/read = %d/%d, want 1/0", b, r)
	}
}

// AND HELP READ AFTER THE REFUSAL DOES NOT COUNT. This is the measured behaviour: the lens seat
// guessed a flag, was refused, and read the help as damage control. Scoring that as compliance
// forgives the exact act under test.
func TestHelpReadAfterTheRefusalDoesNotCount(t *testing.T) {
	p := traj(t,
		use("a", `/tmp/x/feov-record fetch --url http://x --sha256 deadbeef`), res("a", "unknown flag", true),
		use("b", `/tmp/x/feov-record fetch --help`), res("b", "fetch a source", false),
		use("c", `/tmp/x/feov-record fetch --url http://x`), res("c", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.FirstUses) != 1 {
		t.Fatalf("first uses = %+v, want exactly one (the FIRST invocation is the measurement)", s.FirstUses)
	}
	if got := s.FirstUses[0].Depth; got != DepthNone {
		t.Errorf("depth = %q, want %q — the help arrived after the call it should have preceded", got, DepthNone)
	}
}

// GROUP AND COMMAND ARE DIFFERENT STEPS and must not merge: an intervention that moves one and
// not the other is invisible if they share a bucket.
func TestGroupHelpIsNotCommandHelp(t *testing.T) {
	p := traj(t,
		use("a", `/tmp/x/feov-record motion --help`), res("a", "Available Commands:\n  petition  file one\n", false),
		use("b", `/tmp/x/feov-record motion petition file --class integrity`), res("b", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.FirstUses[0].Depth; got != DepthGroup {
		t.Errorf("depth = %q, want %q", got, DepthGroup)
	}
	p2 := traj(t,
		use("a", `/tmp/x/feov-record motion petition file --help`), res("a", "file a petition", false),
		use("b", `/tmp/x/feov-record motion petition file --class integrity`), res("b", "ok", false),
	)
	s2, _ := ReadSurvey(p2, "feov-record", nil)
	if got := s2.FirstUses[0].Depth; got != DepthCommand {
		t.Errorf("depth = %q, want %q", got, DepthCommand)
	}
}

// A SITTING THAT RAN NOTHING IS NOT A CLEAN SITTING. Asked=false is the loud form; a bare ratio
// would report 0 blind uses, which is what a perfect sitting also reports.
func TestASittingThatInvokedNothingSaysSoRatherThanScoringZero(t *testing.T) {
	p := traj(t, use("a", `/tmp/x/feov-record --seat-id x --help`), res("a", rootHelp, false))
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.Asked {
		t.Error("the seat invoked no command and the survey claims it did")
	}
	if s.BlindFirst() != 0 || len(s.FirstUses) != 0 {
		t.Errorf("first uses = %+v, want none", s.FirstUses)
	}
}

// SURVEYING MEANS OPENING A PAGE YOU DO NOT GO ON TO USE. A seat that only ever reads the help of
// verbs it has already chosen is confirming a decision, not weighing options — and that is the
// behaviour a per-command gate would entrench rather than fix.
func TestPagesOpenedForCommandsNeverRunAreTheSurveySignal(t *testing.T) {
	p := traj(t,
		use("a", `/tmp/x/feov-record --seat-id x --help`), res("a", rootHelp, false),
		use("b", `/tmp/x/feov-record finding --help`), res("b", findingHelp, false),
		use("c", `/tmp/x/feov-record verify --help`), res("c", "adjudicate one citation", false),
		use("d", `/tmp/x/feov-record finding --quote "x" --reason "y"`), res("d", "ok", false),
	)
	s, err := ReadSurvey(p, "feov-record", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.PagesNeverUsed != 1 {
		t.Errorf("pages never used = %d, want 1 (verify was read and not run)", s.PagesNeverUsed)
	}
	if s.LastHelpCall != 3 || s.TotalCalls != 4 {
		t.Errorf("lastHelpCall/totalCalls = %d/%d, want 3/4", s.LastHelpCall, s.TotalCalls)
	}
}

// THE TOKENISER IS THE FIRST THING THAT BREAKS. A scratch pass over these same trajectories
// reported `"`, `edit \` and a paragraph of a --reason body as commands the seat had run.
func TestQuotedPathsContinuationsAndProseAreNotVerbs(t *testing.T) {
	for _, tc := range []struct{ cmd, want string }{
		{`"/tmp/x/feov-record" --seat-id s finding --quote "a"`, "finding"},
		{"/tmp/x/feov-record edit \\\n  --key R1-1", "edit"},
		{`/tmp/x/feov-record position --reason $(cat <<EOF
Red identified a real gap: the report's load-bearing input cannot be inspected.
EOF
)`, "position"},
		{`/tmp/x/feov-record show debate 2>&1`, "show debate"},
		{`/tmp/x/feov-record count-claims`, "count-claims"},
	} {
		got := CommandWords(BinFields(tc.cmd, "feov-record"))
		if j := joinWords(got); j != tc.want {
			t.Errorf("CommandWords(%q) = %q, want %q", tc.cmd, j, tc.want)
		}
	}
}

func joinWords(w []string) string {
	out := ""
	for i, s := range w {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}

// QUOTED PROSE MUST NOT NOMINATE A VERB. `--reason "close it"` leaves `close` and `it` as bare
// tokens once quotes are stripped, and the tokeniser this replaced appended them: the caller that
// existed at the time was shielded by matching against a known verb set, and the survey is not.
// A phantom command is a first-use with no help read — an invented blind call.
func TestQuotedProseCannotNominateAVerb(t *testing.T) {
	for _, tc := range []struct{ cmd, want string }{
		{`/tmp/x/feov-record finding --quote "the corpus holds 340" --reason "close the gap"`, "finding"},
		{`/tmp/x/feov-record --run . --seat-id red-lens-r1-L1 lens finding --key F1`, "finding"},
		{`/tmp/x/feov-record motion petition file --class integrity --relief "strike the requirement"`, "motion petition file"},
		{`/tmp/x/feov-record friction --none --reason "reached for verify, found it"`, "friction"},
	} {
		if got := joinWords(CommandWords(BinFields(tc.cmd, "feov-record"))); got != tc.want {
			t.Errorf("CommandWords(%q) = %q, want %q", tc.cmd, got, tc.want)
		}
	}
}
