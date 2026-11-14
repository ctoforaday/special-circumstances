package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The write path's fixture helpers: a run, a registered seat, an act appended through it.

func corrRun(t *testing.T) Run {
	t.Helper()
	return mustRun(t, newRun(t))
}

func sit(t *testing.T, run Run, seat string) Identity {
	t.Helper()
	id := Identity{Run: run, SeatID: seat}
	// The bench owes an OCCASION and no other seat may pass one, so this helper asks the same
	// question the write path does rather than carrying a list of which seats are the bench.
	occasion := ""
	if SeatOwesOccasion(seat) {
		occasion = "docket"
	}
	if _, _, err := RegisterSeat(id, "", occasion); err != nil {
		t.Fatal(err)
	}
	return id
}

func mustAppend(t *testing.T, id Identity, body proto.Message) *Event {
	t.Helper()
	ev, err := Append(id, body)
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func seatLog(text string) *recordpb.Log {
	return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
}

func corrMint(id string) *recordpb.Mint {
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p " + id), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
}

func mustRefuse(t *testing.T, err error, wants ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("accepted; want a refusal saying %q", wants)
	}
	for _, w := range wants {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
}

// sitToAssemble opens the bench's assembly sitting — the one sitting whose outcome the write takes
// (requireAssemblySitting).
func sitToAssemble(t *testing.T, run Run) Identity {
	t.Helper()
	id := Identity{Run: run, SeatID: benchSeat}
	if _, _, err := RegisterSeat(id, "", recordpb.Word(recordpb.Occasion_OCCASION_ASSEMBLE)); err != nil {
		t.Fatal(err)
	}
	return id
}

// withReport ingests a report holding the Findings section every corrMint is about: an act that
// stands on delivered text needs a report head to cut its delivery at.
func withReport(t *testing.T, run Run) Run {
	t.Helper()
	mustAppend(t, Identity{Run: run, SeatID: HarnessSeat}, &recordpb.ReportPart{Text: proto.String("# report\n\n## Findings\n\nThe finding stands.\n")})
	return run
}

// delivered completes id's delivery of report text in its open sitting, as the delivery writer does
// before the seat's act; a sitting that already holds its delivery is left as it is.
func delivered(t *testing.T, id Identity) {
	t.Helper()
	db, err := openRunForRead(id.Run)
	if err != nil {
		t.Fatal(err)
	}
	_, held, err := DeliveryThisSitting(db, id.SeatID)
	if err != nil {
		t.Fatal(err)
	}
	if !held {
		deliver(t, id.Run, id.SeatID)
	}
}

// sitRead opens seat's sitting and completes its delivery of report text in it: the sitting of a
// seat whose next act stands on text delivered to it.
func sitRead(t *testing.T, run Run, seat string) Identity {
	t.Helper()
	id := sit(t, run, seat)
	delivered(t, id)
	return id
}
