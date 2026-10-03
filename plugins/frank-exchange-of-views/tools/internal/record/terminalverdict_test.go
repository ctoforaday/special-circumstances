// MOVED FROM internal/dashboard with the function it tests. It reads only the record, and two
// other packages needed the answer; leaving the test behind would have left the implementation
// covered from a package that no longer owns it.
package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE VERDICT COMES OFF THE RECORD, NOT OUT OF THE PROSE IT WAS RENDERED INTO.
//
// The regex was the only reader, and it is coupled to a sentence assemble.go owns: basisNote
// appends " — **derived from the record**…" directly after the verdict word, so the pattern has to
// stop at an em-dash. A rewording there would blank the dashboard, and a blank dashboard is what an
// un-assembled run looks like too.
//
// The disagreement case is the one worth pinning: when the record says one thing and the rendered
// prose says another, the record wins. Anything else makes the dashboard a reader of a reader.
func TestTerminalVerdictPrefersTheRecordOverTheRenderedProse(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge"}, "", "docket"); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge"}, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED), Prose: proto.String("ended on safety grounds")}); err != nil {
		t.Fatal(err)
	}
	// The rendered artifact says something else. It is the derived carrier; the event is the fact.
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Outcome:** VERIFIED — **derived from the record**, not claimed.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// THE WORD COMES FROM THE SCHEMA, in the capitals TerminalVerdict answers in. The point here is
	// that the RECORD beats the rendered prose — report.md above says VERIFIED.
	want := strings.ToUpper(recordpb.Word(recordpb.RunOutcome_RUN_OUTCOME_HALTED))
	if got, err := TerminalVerdict(mustRun(t, runDir)); err != nil || got != want {
		t.Errorf("TerminalVerdict = (%q, %v), want %q — the record holds the verdict as a field and the report is a rendering of it", got, err, want)
	}
}

// AND WHEN THE RECORD CANNOT SAY, NOTHING IS INVENTED FROM THE PROSE.
//
// Measured across the 9 assembled runs in research/: five carry no terminal act at all while their
// reports read "UNVERIFIED" — a word backed by no record anywhere, written by a pre-#289 assembler.
// The old fallback's job was to hand that word to an operator as the run's verdict.
//
// Empty is the honest answer and the renderer already uses it well: it falls through to the round
// verdict off the record and relabels "final verdict" as "latest verdict (epoch N)", so the operator is
// shown a different claim rather than the same claim from a worse source.
func TestTerminalVerdictIsEmptyWhenTheRecordCannotSay(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Outcome:** UNVERIFIED — the run ended without the question being answered.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := TerminalVerdict(mustRun(t, runDir)); err != nil || got != "" {
		t.Errorf("TerminalVerdict = (%q, %v) from a run whose record carries no terminal act — the word was read out of prose no record backs", got, err)
	}
}

// EVERY READ DeriveVerdict MAKES SURFACES ITS FAILURE THROUGH TerminalVerdict, each failed alone.
//
// When the bench recorded no outcome, TerminalVerdict asks the record to decide, and that is four
// reads in order: the halt, the gate, the cast (UnseatedAreas and CastOf read it), and the
// dispatch plan. A fold on any one of them turns a busy record into ("", nil) — "not ended" — and
// the liveness audit convicts a finished run as TERMINATED. Failing only the first read proves
// nothing about the rest, because it returns first; so each case leaves the reads before it healthy
// and breaks exactly one, and the error must name what that read asked for.
//
// The fixture holds a cast and nothing terminal, so a healthy record answers ("", nil) only after
// reaching the plan read. A database read is failed by dropping its table through the run's cached
// handle, the way every later read of that run would then fail; the plan read is failed at the
// run's configuration, which it reads first.
func TestTerminalVerdictSurfacesEachReadTheDerivationMakes(t *testing.T) {
	healthy := castWith(t, "evidence")
	if v, err := TerminalVerdict(healthy); err != nil || v != "" {
		t.Fatalf("TerminalVerdict over a healthy cast-only record = (%q, %v), want (\"\", nil)", v, err)
	}
	if _, err := PlanDispatch(healthy); err != nil {
		t.Fatalf("the healthy fixture's plan read fails, so the plan case below would prove nothing: %v", err)
	}
	dropping := func(stmt string) func(*testing.T, Run) {
		return func(t *testing.T, run Run) {
			db, err := openRunForRead(run)
			if err != nil || db == nil {
				t.Fatalf("open the run's handle: (%v, %v)", db, err)
			}
			if _, err := db.Exec(stmt); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, tc := range []struct {
		read  string
		fail  func(*testing.T, Run)
		names string // what the error must carry, so it is THIS read that failed
	}{
		{"halt", dropping(`DROP TABLE "halt"`), "halt"},
		{"gate", dropping(`DROP TABLE "gate"`), "gate"},
		{"cast", dropping(`DROP TABLE "cast_seat_ids"`), "cast_seat_ids"},
		{"plan", func(t *testing.T, run Run) {
			if err := os.MkdirAll(filepath.Join(run.Dir(), "inputs"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(run.Dir(), "inputs", "run-config.json"), []byte("{not json"), 0o644); err != nil {
				t.Fatal(err)
			}
		}, "run-config.json"},
	} {
		t.Run(tc.read, func(t *testing.T) {
			run := castWith(t, "evidence")
			tc.fail(t, run)
			v, err := TerminalVerdict(run)
			if err == nil || v != "" {
				t.Fatalf("TerminalVerdict with the %s read failing = (%q, %v), want the error — \"\" here reads as a run in flight", tc.read, v, err)
			}
			if !strings.Contains(err.Error(), tc.names) {
				t.Errorf("the error does not name the %s read's %q, so another read failed: %v", tc.read, tc.names, err)
			}
		})
	}
}

// ONE SPELLING FROM EITHER READ. TerminalVerdict answers from the bench's recorded outcome, stored
// as the schema's lower-case word, or from DeriveVerdict, which names the verdict in capitals. Both
// reach an operator through the same consumers, so the same verdict through either path is the same
// string — the seat's capitals.
func TestTerminalVerdictAnswersOneSpellingFromEitherRead(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	recorded := mustRun(t, recordtest.TmpRun(t))
	if _, _, err := RegisterSeat(Identity{Run: recorded, SeatID: "judge"}, "", "docket"); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: recorded, SeatID: "judge"}, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED), Prose: proto.String("ended on safety grounds")}); err != nil {
		t.Fatal(err)
	}
	derived := mustRun(t, runWith(t, "3", []*Event{vev(t, "judge", 1, &recordpb.Halt{Opinion: proto.String("consent gate")})}))

	for _, tc := range []struct {
		path       string
		run        Run
		hasOutcome bool // which read answers: the recorded outcome, or the derivation
	}{
		{"recorded outcome", recorded, true},
		{"derived verdict", derived, false},
	} {
		t.Run(tc.path, func(t *testing.T) {
			if o, err := RecordedOutcome(tc.run); err != nil || (o != "") != tc.hasOutcome {
				t.Fatalf("RecordedOutcome = (%q, %v) — the fixture does not exercise the %s path", o, err, tc.path)
			}
			if got, err := TerminalVerdict(tc.run); err != nil || got != "HALTED" {
				t.Errorf("TerminalVerdict through the %s = (%q, %v), want \"HALTED\"", tc.path, got, err)
			}
		})
	}
}
