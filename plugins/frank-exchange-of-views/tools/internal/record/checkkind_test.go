package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// check_kind IS THE FIELD WITH TEETH, AND IT REACHED NO VIEW.
//
// `computation` means the gap cannot be closed on prose at all — only a `blue prove --answers`
// settles it, and the chair is refused if it tries anything else. That gate works. But it fires
// at the CHAIR, at close time, in the following round, and blue — the only seat that can satisfy
// it — could not see which gaps carried it.
//
// Measured: `prove` was invoked zero times across eighteen probed seat dispatches on boards built
// to demand it. One seat summed twelve integers inside its own reasoning, wrote the answer, and
// was satisfied. The answer was right, which is why nothing caught it.
func TestCheckKindReachesTheSeatThatMustSatisfyIt(t *testing.T) {
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, "", ""); err != nil {
		t.Fatal(err)
	}
	// THE ID AND THE KIND ARE THE SUBJECT OF THIS TEST, and the earlier conversion dropped both
	// from the body — every gap would have been minted identical and unidentified, which is a
	// fixture that cannot fail the way the test intends.
	for _, c := range []struct {
		id   string
		kind recordpb.CheckKind
	}{
		{"G1", recordpb.CheckKind_CHECK_KIND_COMPUTATION},
		{"G2", recordpb.CheckKind_CHECK_KIND_DOCUMENT},
	} {
		if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			GapId:           proto.String(c.id),
			Class:           proto.String("self-attestation"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       c.kind.Enum(),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}); err != nil {
			t.Fatal(err)
		}
	}
	got := map[string]string{}
	for _, g := range mustBoardJSONT(t, mustRun(t, runDir)).Open {
		got[g.ID] = g.CheckKind
	}
	if got["G1"] != "computation" || got["G2"] != "document" {
		t.Errorf("board view check_kind = %v — a seat cannot know which gaps demand a program", got)
	}

	// The work list is the SCANNING read a seat plans its sitting from, so the demand's TYPE has
	// to be on it even though the rest of the acceptance check deliberately is not.
	got = map[string]string{}
	for _, g := range mustWorkJSONT(t, mustRun(t, runDir)).Open {
		got[g.ID] = g.CheckKind
	}
	if got["G1"] != "computation" || got["G2"] != "document" {
		t.Errorf("work check_kind = %v — the read a seat plans from cannot say which gaps prose will not close", got)
	}
}

// AN EMPTY LOG IS TWO DIFFERENT RUNS, and only one of them is fine. The distinction survives the
// retirement of the entry that used to carry it: a seat that SAT and filed nothing is clean, read
// off the harness bracket, and a run nobody sat in is silent. Neither costs the seat a call.
func TestTheLogViewSeparatesSilenceFromAnAttestation(t *testing.T) {
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, "", ""); err != nil {
		t.Fatal(err)
	}
	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	// THE SEAT HAS SAT, so it is already clean — it opened a sitting and has said nothing.
	j := LogJSONOf(b.Events)
	if j.Counts.Total != 0 || j.Counts.Clean != 1 {
		t.Fatalf("a seat that sat and said nothing: total=%d clean=%d, want 0/1", j.Counts.Total, j.Counts.Clean)
	}

	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, &recordpb.Log{Text: proto.String("read the board and my verb list; every refusal was my own error"), Type: recordpb.LogType_LOG_TYPE_FRICTION.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}); err != nil {
		t.Fatal(err)
	}
	b, _ = FamilyOf(mustRun(t, runDir))
	j = LogJSONOf(b.Events)
	// ONCE IT SPEAKS IT IS NO LONGER CLEAN, and what it filed asserts a problem. Every surviving
	// type does, so the entry counts toward the total and the seat leaves the clean set.
	if j.Counts.Total != 1 {
		t.Errorf("a filed entry must count: total=%d", j.Counts.Total)
	}
	if j.Counts.Clean != 0 || len(j.Log) != 1 {
		t.Fatalf("a seat that spoke is not clean: clean=%d entries=%d", j.Counts.Clean, len(j.Log))
	}
	if j.Log[0].SeatID != "blue-respond" {
		t.Error("the entry must name the seat that made it — an unattributed one cannot be weighed")
	}
}

// A PROPERTY IS NOT A DEBT.
//
// Projecting check_kind was necessary and not sufficient: with it visible, `prove` moved from
// 0 uses across eighteen sittings to 1 across nine. A seat reading `"check_kind": "computation"`
// learns something about the gap; it does not learn that IT owes a program, and only the second
// changes what the sitting produces.
func TestAwaitingProofTracksTheDebtAndAgreesWithTheGate(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"red-chair", "blue-respond"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: s}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	mint := func(id string, kind recordpb.CheckKind) {
		t.Helper()
		if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			GapId:           proto.String(id),
			Class:           proto.String("self-attestation"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       kind.Enum(),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}); err != nil {
			t.Fatal(err)
		}
	}
	mint("G1", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	mint("G2", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	mint("G3", recordpb.CheckKind_CHECK_KIND_DOCUMENT)

	owed := GapsAwaitingProof(mustRun(t, runDir))
	if len(owed) != 2 || owed[0] != "G1" || owed[1] != "G2" {
		t.Fatalf("owed = %v, want the two computation gaps in board order", owed)
	}
	// A document gap is never a proof debt — over-reporting would train seats to ignore it.
	for _, id := range owed {
		if id == "G3" {
			t.Error("a document-kind gap was reported as awaiting a computation")
		}
	}

	// THE PROOF ANSWERS A GAP, and the earlier conversion dropped `answers` — so the proof
	// discharged nothing and the debt could not move. `answers` is the whole join this test is
	// about: a proof that names no gap is a script that ran for no stated reason.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, &recordpb.Proof{
		Answers: proto.String("G1"),
		Script:  proto.String("s.py"),
	}); err != nil {
		t.Fatal(err)
	}
	if owed := GapsAwaitingProof(mustRun(t, runDir)); len(owed) != 1 || owed[0] != "G2" {
		t.Fatalf("after proving G1, owed = %v, want [G2]", owed)
	}

	// THE BOARD AND THE GATE MUST NOT DISAGREE about what is owed. They share one join and one
	// constant precisely so a seat cannot be told it owes nothing by the read and be refused by
	// the write — three bare "computation" literals used to make that possible.
	fromBoard := map[string]bool{}
	for _, g := range mustBoardJSONT(t, mustRun(t, runDir)).Open {
		if g.AwaitingProof {
			fromBoard[g.ID] = true
		}
	}
	if len(fromBoard) != 1 || !fromBoard["G2"] {
		t.Fatalf("board says %v awaits proof, the debt query says [G2]", fromBoard)
	}
	for _, g := range mustWorkJSONT(t, mustRun(t, runDir)).Open {
		if g.AwaitingProof != fromBoard[g.ID] {
			t.Errorf("work and board disagree on %s: %v vs %v", g.ID, g.AwaitingProof, fromBoard[g.ID])
		}
	}

	// A CLOSED gap owes nothing, whatever its kind: the debt is what blue can still act on.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Close{
		GapId:        proto.String("G2"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("x"),
		Prose:        proto.String("the computation ran and the check holds"),
	}); err != nil {
		t.Fatal(err)
	}
	if owed := GapsAwaitingProof(mustRun(t, runDir)); len(owed) != 0 {
		t.Errorf("a closed gap is still reported as owed: %v", owed)
	}
}
