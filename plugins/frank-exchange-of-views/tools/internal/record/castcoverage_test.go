package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// castWith seats exactly the named areas and returns the run.
func castWith(t *testing.T, areas ...string) Run {
	t.Helper()
	run := mustRun(t, recordtest.TmpRun(t))
	harness := Identity{Run: run, SeatID: HarnessSeat}
	if _, err := Append(harness, &recordpb.Cast{SeatIds: func() []string { s, _ := CastFor(areas, 1); return s }()}); err != nil {
		t.Fatal(err)
	}
	return run
}

// A NARROWED CAST IS A DIMENSION NOBODY AUDITED, and the subtraction is the whole fix: the cast and
// the roster were both on the record already and no reader compared them. #877's measured case is
// the first row — a run seated on evidence and logic shipped VERIFIED carrying 19 voice tells.
func TestUnseatedAreasIsTheRosterMinusTheCast(t *testing.T) {
	for _, tt := range []struct {
		name  string
		seat  []string
		want  []string
		quiet bool
	}{
		{name: "#877's cast: evidence and logic only", seat: []string{"evidence", "logic"},
			want: []string{"adversary", "architecture", "computation", "dark-side", "voice"}},
		{name: "one area dropped", seat: []string{"adversary", "architecture", "computation", "dark-side", "evidence", "logic"},
			want: []string{"voice"}},
		{name: "a whole cast says nothing", seat: DefaultCastAreas, want: nil, quiet: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, hasCast, err := UnseatedAreas(castWith(t, tt.seat...))
			if err != nil {
				t.Fatal(err)
			}
			if !hasCast {
				t.Error("a run WITH a cast reported hasCast=false")
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("unseated = %v, want %v", got, tt.want)
			}
			// THE QUIET CASE IS THE ONE THAT MUST NOT REGRESS. A note on a complete cast would
			// train every reader to skip the line, which is how the next real omission goes unread.
			if note := CoverageNote(got); tt.quiet && note != "" {
				t.Errorf("a complete cast produced a coverage note: %q", note)
			} else if !tt.quiet && note == "" {
				t.Error("a narrowed cast produced no coverage note")
			}
		})
	}
}

// NO CAST IS NOT FULL COVERAGE. A record with no cast has not narrowed anything — it has not
// started — and answering "nothing unseated" would hand a caller a clean bill for a run that never
// opened. That is the same distinction DeriveVerdict's CEILING arm already draws.
func TestARecordWithNoCastIsNotReportedAsFullyCovered(t *testing.T) {
	got, hasCast, err := UnseatedAreas(mustRun(t, recordtest.TmpRun(t)))
	if err != nil {
		t.Fatal(err)
	}
	// THE TWO ZEROES MUST DIFFER. An empty list alone cannot say which this is, and a caller that
	// reads "nothing unseated" off a run with no cast has been handed a clean bill for a run that
	// never opened.
	if hasCast {
		t.Error("a run with no cast reported hasCast=true — the zeroes are indistinguishable again")
	}
	if got != nil {
		t.Errorf("a run with no cast reported unseated areas %v — it has not narrowed anything", got)
	}
}

// THE STAMP MUST NOT READ THE SAME both ways. This is the defect in one assertion: before the
// subtraction, a terminal verdict on a cast missing two dimensions was byte-identical to one on a
// whole cast.
//
// HALT, NOT PASS, and the first draft of this test is why. A PASS is refused until a report has
// been ingested, so the fixture could not reach one and the assertion sat behind a t.Skipf —
// GREEN, and never run. A test that can skip is a test that can lie about having checked; the halt
// arm reaches the same composition in DeriveVerdict with preconditions a fixture can actually meet.
func TestATerminalVerdictStatesTheDimensionsNobodyAudited(t *testing.T) {
	halt := func(run Run) (string, string) {
		t.Helper()
		judge := Identity{Run: run, SeatID: "judge"}
		if _, _, err := RegisterSeat(judge, "", "docket"); err != nil {
			t.Fatal(err)
		}
		if _, err := Append(judge, &recordpb.Halt{Opinion: proto.String("consent gate")}); err != nil {
			t.Fatal(err)
		}
		v, why, ok := DeriveVerdict(run)
		if !ok || v != "HALTED" {
			t.Fatalf("verdict = %q %q ok=%v, want HALTED", v, why, ok)
		}
		return v, why
	}

	_, why := halt(castWith(t, "evidence", "logic"))
	for _, want := range []string{"voice", "no lens sat"} {
		if !strings.Contains(why, want) {
			t.Errorf("the basis does not say %q — a reader cannot tell it from a full audit:\n%s", want, why)
		}
	}

	// The control: a whole cast reaching the same verdict says nothing extra.
	if _, why2 := halt(castWith(t, DefaultCastAreas...)); strings.Contains(why2, "no lens sat") {
		t.Errorf("a complete cast carried a coverage note:\n%s", why2)
	}
}
