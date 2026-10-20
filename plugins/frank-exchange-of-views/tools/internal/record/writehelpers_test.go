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

// replayed appends as migrate replays an archived act.
func replayed(id Identity, body proto.Message) (*Event, error) {
	Migrating = true
	defer func() { Migrating = false }()
	return Append(id, body)
}

func seatLog(text string) *recordpb.Log {
	return &recordpb.Log{Text: proto.String(text), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
}

func corrMint(id string) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p " + id), AcceptanceCheck: proto.String("c"),
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
