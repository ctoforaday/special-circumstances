package recordsql

import (
	"database/sql"
	"fmt"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE LIVE-ANSWER DEMAND IS ASKED OF ROWS, ARM BY ARM, INCLUDING ROWS THE WRITE PATH REFUSES. A gap
// whose check kind is `source` or `computation` is evidence_answered by a live evidence anchor of its
// kind placed as its answer (arm 1), by a retire of the anchor it is about (arm 2), by a retire of
// the anchor of a reading it carries (arm 3), or by its own check of that anchor made again since the
// mint and holding (arm 4, held level with the ledger by record's TestTheBackingArmIsTheLedgersHolds).
// Each subtest is one arm's rows and the arms' neighbours that must NOT meet it.

// demandRecord is a fixture record whose acts are inserted in order under the seat each names.
type demandRecord struct {
	t  *testing.T
	db *sql.DB
	n  int32
}

func newDemandRecord(t *testing.T) *demandRecord { return &demandRecord{t: t, db: store(t)} }

// add inserts body as seat's act and returns its events.id.
func (r *demandRecord) add(seat string, typ recordpb.EventType, body proto.Message) int64 {
	r.t.Helper()
	ev := event(r.t, r.n, typ, body)
	ev.SeatId = proto.String(seat)
	ev.Key = proto.String(fmt.Sprintf("%s:act:#%d", seat, r.n))
	r.n++
	id, err := Insert(r.db, ev)
	if err != nil {
		r.t.Fatal(err)
	}
	return id
}

func (r *demandRecord) mint(seat, gap string, kind recordpb.CheckKind, about string) int64 {
	m := &recordpb.Mint{ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
		GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem: proto.String("p"), AcceptanceCheck: proto.String("a"), CheckKind: kind.Enum(),
		Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
		AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings")}
	if about != "" {
		m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), proto.String(about)
	}
	return r.add(seat, recordpb.EventType_EVENT_TYPE_MINT, m)
}

func (r *demandRecord) cite(label, answers string) int64 {
	c := &recordpb.Cite{Label: proto.String(label)}
	if answers != "" {
		c.Answers = proto.String(answers)
	}
	return r.add("blue-respond", recordpb.EventType_EVENT_TYPE_CITE, c)
}

func (r *demandRecord) compute(id, answers string) int64 {
	c := &recordpb.Computation{ComputationId: proto.String(id), Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), Argv: []string{"true"}}
	if answers != "" {
		c.Answers = proto.String(answers)
	}
	return r.add("blue-respond", recordpb.EventType_EVENT_TYPE_COMPUTATION, c)
}

func (r *demandRecord) retire(anchors ...string) int64 {
	return r.add("blue-respond", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c"), Reason: proto.String("r"), Anchors: anchors})
}

func (r *demandRecord) verify(seat, anchor, carriedBy string, o recordpb.SourceOutcome) int64 {
	v := &recordpb.Verify{Claim: proto.String("c"), Anchor: proto.String(anchor), Outcome: o.Enum(),
		Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("t")}
	if carriedBy != "" {
		v.CarriedBy = proto.String(carriedBy)
	}
	return r.add(seat, recordpb.EventType_EVENT_TYPE_VERIFY, v)
}

func (r *demandRecord) reproduce(seat, id string, s recordpb.Soundness) int64 {
	return r.add(seat, recordpb.EventType_EVENT_TYPE_REPRODUCE, &recordpb.Reproduce{ComputationId: proto.String(id), Soundness: s.Enum(), Ran: proto.Bool(true)})
}

func (r *demandRecord) close(gap, carriedFrom string) int64 {
	c := &recordpb.Close{GapId: proto.String(gap), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(), Prose: proto.String("closed")}
	if carriedFrom != "" {
		c.CarriedFrom = proto.String(carriedFrom)
	}
	return r.add("red-lens-evidence", recordpb.EventType_EVENT_TYPE_CLOSE, c)
}

// demand is the gap's (evidence_answered, awaiting_evidence, answers_retired_event).
func (r *demandRecord) demand(gap string) (answered, awaiting bool, retired sql.NullInt64) {
	r.t.Helper()
	if err := r.db.QueryRow(`SELECT "evidence_answered", "awaiting_evidence", "answers_retired_event" FROM "gap" WHERE "gap_id" = ?`, gap).
		Scan(&answered, &awaiting, &retired); err != nil {
		r.t.Fatal(err)
	}
	return answered, awaiting, retired
}

func (r *demandRecord) wantAnswered(gap string, want bool, why string) {
	r.t.Helper()
	answered, awaiting, _ := r.demand(gap)
	if answered != want || awaiting == want {
		r.t.Errorf("%s: (evidence_answered=%v, awaiting_evidence=%v), want answered=%v: %s", gap, answered, awaiting, want, why)
	}
}

const (
	reader = "red-lens-evidence"
	other  = "red-lens-logic"
)

func TestTheLiveAnswerDemandIsMetByEachOfItsArms(t *testing.T) {
	src, comp := recordpb.CheckKind_CHECK_KIND_SOURCE, recordpb.CheckKind_CHECK_KIND_COMPUTATION
	t.Run("arm 1: a live answer of the gap's kind", func(t *testing.T) {
		r := newDemandRecord(t)
		r.mint(reader, "G-00000001", src, "")
		r.mint(reader, "G-00000002", comp, "")
		r.wantAnswered("G-00000001", false, "nothing answers it")
		r.compute("K-00000001", "G-00000001")
		r.wantAnswered("G-00000001", false, "a computation is not a source gap's answer")
		r.cite("C-00000001", "G-00000002")
		r.wantAnswered("G-00000002", false, "a citation is not a computation gap's answer")
		r.cite("C-00000002", "G-00000001")
		r.wantAnswered("G-00000001", true, "a live citation placed as its answer")
		r.compute("K-00000002", "G-00000002")
		r.wantAnswered("G-00000002", true, "a live computation placed as its answer")
		r.retire("C-00000002")
		r.wantAnswered("G-00000001", false, "its only answer is retired, so it is owed again")
	})
	t.Run("arm 2: the anchor the gap is about, retired", func(t *testing.T) {
		r := newDemandRecord(t)
		r.cite("C-00000003", "")
		r.compute("K-00000003", "")
		r.mint(reader, "G-00000003", src, "C-00000003")
		r.mint(reader, "G-00000004", comp, "K-00000003")
		r.wantAnswered("G-00000003", false, "its anchor stands")
		r.retire("C-00000003")
		r.retire("K-00000003")
		r.wantAnswered("G-00000003", true, "the citation it is about is retired")
		r.wantAnswered("G-00000004", true, "the computation it is about is retired")
	})
	t.Run("arm 3: the anchor of a reading the gap carries, retired", func(t *testing.T) {
		r := newDemandRecord(t)
		r.cite("C-00000005", "")
		r.cite("C-00000006", "")
		r.mint(reader, "G-00000005", src, "")
		r.verify(reader, "C-00000005", "G-00000005", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
		r.verify(other, "C-00000006", "", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
		r.wantAnswered("G-00000005", false, "the read citation stands")
		r.retire("C-00000006")
		r.wantAnswered("G-00000005", false, "a retire of a citation no reading it carries read")
		r.retire("C-00000005")
		r.wantAnswered("G-00000005", true, "the citation its reading read is retired")
	})
	t.Run("arm 4: the gap's own check again, holding", func(t *testing.T) {
		r := newDemandRecord(t)
		r.cite("C-00000007", "")
		r.compute("K-00000007", "")
		r.verify(reader, "C-00000007", "", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
		r.mint(reader, "G-00000007", src, "C-00000007")
		r.mint(reader, "G-00000008", comp, "K-00000007")
		r.mint(reader, "G-00000009", src, "K-00000007")
		r.wantAnswered("G-00000007", false, "a backing reading before the mint meets nothing")
		r.verify(other, "C-00000007", "", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
		r.reproduce(other, "K-00000007", recordpb.Soundness_SOUNDNESS_SOUND)
		r.wantAnswered("G-00000007", false, "another seat's backing reading meets nothing")
		r.wantAnswered("G-00000008", false, "another seat's sound judgement meets nothing")
		r.verify(reader, "C-00000007", "", recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)
		r.reproduce(reader, "K-00000007", recordpb.Soundness_SOUNDNESS_SOUND)
		r.wantAnswered("G-00000007", true, "its minter's own backing reading since the mint")
		r.wantAnswered("G-00000008", true, "its minter's own sound judgement since the mint")
		r.wantAnswered("G-00000009", false, "a judgement is never a source gap's answer")
	})
}

// THE ANSWERS TO A CLOSED GAP, ALL RETIRED, are the lens's news until a closing act other than a
// carry follows the retire; never on an open gap, and never while one answer stands.
func TestAnswersRetiredEventIsTheLastRetireAfterTheClosure(t *testing.T) {
	r := newDemandRecord(t)
	comp := recordpb.CheckKind_CHECK_KIND_COMPUTATION
	r.mint(reader, "G-0000000a", comp, "")
	r.mint(reader, "G-0000000b", comp, "")
	r.compute("K-0000000a", "G-0000000a")
	r.compute("K-0000000b", "G-0000000a")
	r.compute("K-0000000c", "G-0000000b")
	r.retire("K-0000000c")
	if _, _, ret := r.demand("G-0000000b"); ret.Valid {
		t.Errorf("an open gap whose answers are retired holds answers_retired_event %d: the column is a closed gap's", ret.Int64)
	}
	r.close("G-0000000a", "")
	r.retire("K-0000000a")
	if _, _, ret := r.demand("G-0000000a"); ret.Valid {
		t.Errorf("answers_retired_event = %d while K-0000000b still answers the gap", ret.Int64)
	}
	last := r.retire("K-0000000b")
	if _, _, ret := r.demand("G-0000000a"); ret.Int64 != last {
		t.Errorf("answers_retired_event = %v, want the last retire, event %d", ret, last)
	}
	r.close("G-0000000a", "e1")
	if _, _, ret := r.demand("G-0000000a"); ret.Int64 != last {
		t.Errorf("after a carry, answers_retired_event = %v, want %d: a carry restates the closure it carries", ret, last)
	}
	r.close("G-0000000a", "")
	if _, _, ret := r.demand("G-0000000a"); ret.Valid {
		t.Errorf("after a fresh close, answers_retired_event = %d, want null: that closer had the retirement before it", ret.Int64)
	}
}
