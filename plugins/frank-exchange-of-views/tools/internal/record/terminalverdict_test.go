// MOVED FROM internal/dashboard with the function it tests. It reads only the record, and two
// other packages needed the answer; leaving the test behind would have left the implementation
// covered from a package that no longer owns it.
package record

import (
	"os"
	"path/filepath"
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
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal", Round: RoundIn(mustRun(t, runDir))("judge-terminal")}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "judge-terminal", Round: RoundIn(mustRun(t, runDir))("judge-terminal")}, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_HALTED), Prose: proto.String("ended on safety grounds")}); err != nil {
		t.Fatal(err)
	}
	// The rendered artifact says something else. It is the derived carrier; the event is the fact.
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Verdict:** VERIFIED — **derived from the record**, not claimed.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// THE SPELLING COMES FROM THE SCHEMA, not from a literal. The point here is that the RECORD
	// beats the rendered prose — report.md above says VERIFIED — and the casing is incidental to
	// it: a hardcoded "HALTED" was really asserting how the payload record happened to store the
	// seat's uppercase word.
	want := recordpb.Word(recordpb.RunOutcome_RUN_OUTCOME_HALTED)
	if got := TerminalVerdict(mustRun(t, runDir)); got != want {
		t.Errorf("readTerminalVerdict = %q, want %q — the record holds the verdict as a field and the report is a rendering of it", got, want)
	}
}

// AND WHEN THE RECORD CANNOT SAY, NOTHING IS INVENTED FROM THE PROSE.
//
// Measured across the 9 assembled runs in research/: five carry no terminal act at all while their
// reports read "UNVERIFIED" — a word backed by no record anywhere, written by a pre-#289 assembler.
// The old fallback's job was to hand that word to an operator as the run's verdict.
//
// Empty is the honest answer and the renderer already uses it well: it falls through to the round
// verdict off the record and relabels "final verdict" as "latest verdict (rN)", so the operator is
// shown a different claim rather than the same claim from a worse source.
func TestTerminalVerdictIsEmptyWhenTheRecordCannotSay(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	if err := os.WriteFile(filepath.Join(runDir, "report.md"),
		[]byte("# report\n\n**Verdict:** UNVERIFIED — the run ended without the question being answered.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := TerminalVerdict(mustRun(t, runDir)); got != "" {
		t.Errorf("readTerminalVerdict = %q from a run whose record carries no terminal act — the word was read out of prose no record backs", got)
	}
}
