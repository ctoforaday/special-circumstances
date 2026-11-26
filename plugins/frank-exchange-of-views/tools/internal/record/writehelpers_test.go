package record

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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

// benchSits is the chair convening the bench on every gap an unruled docket motion is about, at the
// report head, and the bench sitting that docket with its delivery of those gaps' passages: the
// sitting a live bench rules from. The chair's sitting is the one it already holds.
func benchSits(t *testing.T, run Run) Identity {
	t.Helper()
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	head, err := ReportHeadAt(db)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT DISTINCT d."gap_id" FROM "motion" m JOIN "motion_docket" d ON d."event_id" = m."event_id"
  WHERE m."motion_id" NOT IN (SELECT "motion_id" FROM "motion_answers" WHERE "ruled_by" IS NOT NULL) ORDER BY m."event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	var gaps []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			t.Fatal(err)
		}
		gaps = append(gaps, g)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	// The chair's dispatch is committed as the record holds it: these fixtures cast no seats, and the
	// dispatch write admits only a cast seat.
	raw, err := recordsql.OpenSeparate(filepath.Join(run.Records(), dbName))
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	ev := &recordpb.Event{SeatId: proto.String("red-chair"), Ts: proto.String(Stamp()), Key: proto.String(fmt.Sprintf("red-chair:dispatch:%s:%s", benchSeat, Stamp()))}
	typ, err := recordpb.SetBody(ev, &recordpb.Dispatch{Pin: proto.Int64(head), SeatId: proto.String(benchSeat), GapIds: gaps})
	if err != nil {
		t.Fatal(err)
	}
	ev.Type = &typ
	tx, err := raw.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recordsql.InsertTx(tx, ev); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return sitRead(t, run, benchSeat)
}

// sitRead opens seat's sitting and completes its delivery of report text in it: the sitting of a
// seat whose next act stands on text delivered to it.
func sitRead(t *testing.T, run Run, seat string) Identity {
	t.Helper()
	id := sit(t, run, seat)
	delivered(t, id)
	return id
}

// repairClose is the gap's assignee closing it repaired.
func repairClose(gap string) *recordpb.Close {
	return &recordpb.Close{GapId: proto.String(gap), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}
}
