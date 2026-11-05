package record

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// everyInput stands in for the renderer with one whose text is every input it is handed, so two
// renders are equal exactly when their inputs are, and it counts the renders. It is a left fold of the
// ops from the empty text, as reportproj.Replay is.
func everyInput(t *testing.T) *int {
	t.Helper()
	calls := new(int)
	withRenderer(t, func(text string, parted bool, ops []reportproj.ReportOp) (string, error) {
		*calls++
		for _, op := range ops {
			parted = parted || op.Kind == "append"
			text += fmt.Sprintf(" | %s %q %q %v", op.Kind, op.A, op.B, op.Exact)
		}
		if !parted {
			return "no part", nil
		}
		return text, nil
	})
	return calls
}

// A GUARD READS THE REPORT ITS WRITE LANDS ON. Another lane's part, another seat's edit and an anchor
// commit at the instant the write opens its transaction (overtaken). The guard must read the report
// with all three in it — the report rendered whole once the write has committed nothing — and never
// one read before them.
func TestAGuardReadsTheReportItsWriteLandsOn(t *testing.T) {
	everyInput(t)
	run := newStage(t).cast(evLens, "blue-respond").ingest().
		add(evLens, &recordpb.Anchor{Id: proto.String("F-1"), Location: proto.String("# report")}).register(evLens).seed()
	var saw string
	var sawErr error
	err := overtaken(t, run, []*Event{
		recordtest.At(t, HarnessSeat, "race:part", &recordpb.ReportPart{Text: proto.String("## another part")}),
		recordtest.At(t, "blue-respond", "race:edit", &recordpb.BlueEdit{Old: proto.String("report"), New: proto.String("account"), EditKey: proto.String("E1")}),
		recordtest.At(t, "red-lens-logic", "race:anchor", &recordpb.Anchor{Id: proto.String("F-2"), Location: proto.String("# account")}),
	}, func() error {
		_, err := Append(Identity{Run: run, SeatID: evLens}, &recordpb.Anchor{Id: proto.String("F-9"), Location: proto.String("# report")},
			Check(func(l *Landing) error {
				saw, sawErr = l.Report()
				return ErrWritesNothing
			}))
		return err
	})
	if !errors.Is(err, ErrWritesNothing) {
		t.Fatalf("the write = %v, want the guard's ErrWritesNothing", err)
	}
	want, werr := RenderAt(recordOf(t, run))
	if sawErr != nil || werr != nil || saw != want {
		t.Errorf("the guard read %q (%v), want the report rendered whole, %q (%v)", saw, sawErr, want, werr)
	}
	for _, op := range []string{`append "## another part"`, `"report" "account"`, `"F-2"`} {
		if !strings.Contains(saw, op) {
			t.Errorf("the guard's report holds no %s, which committed before its write took the lock:\n%s", op, saw)
		}
	}
}

// A MINT RENDERS THE REPORT ONLY WHERE ITS MEASURE READS IT. The budget of a lens whose unit is in
// the report reads the report in the mint's validation (requireMintWithinBudget); a lens measured on
// rows of the record renders nothing.
func TestAMintRendersTheReportOnlyWhereItsMeasureReadsIt(t *testing.T) {
	calls := everyInput(t)
	run := newStage(t).cast("red-lens-voice").ingest().register("red-lens-voice").seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-lens-voice"}, highMint("G1")); err != nil {
		t.Fatalf("the mint: %v", err)
	}
	if *calls != 1 {
		t.Errorf("a voice lens's mint rendered the report %d times, want once", *calls)
	}
	*calls = 0
	if _, err := Append(Identity{Run: run, SeatID: evLens}, highMint("G2")); err != nil {
		t.Fatalf("the evidence mint: %v", err)
	}
	if *calls != 0 {
		t.Errorf("an evidence lens's mint, measured on citations, rendered the report %d times, want none", *calls)
	}
}
