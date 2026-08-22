package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

import "testing"

// check_kind IS THE FIELD WITH TEETH, AND IT REACHED NO VIEW.
//
// `computation` means the gap cannot be closed on prose at all — only a `blue prove --answers`
// settles it, and the merge is refused if it tries anything else. That gate works. But it fires
// at the MERGE, at close time, in the following round, and blue — the only seat that can satisfy
// it — could not see which gaps carried it.
//
// Measured: `prove` was invoked zero times across eighteen probed seat dispatches on boards built
// to demand it. One seat summed twelve integers inside its own reasoning, wrote the answer, and
// was satisfied. The answer was right, which is why nothing caught it.
func TestCheckKindReachesTheSeatThatMustSatisfyIt(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	// THE ID AND THE KIND ARE THE SUBJECT OF THIS TEST, and the earlier conversion dropped both
	// from the body — every gap would have been minted identical and unidentified, which is a
	// fixture that cannot fail the way the test intends.
	for _, c := range []struct {
		id   string
		kind recordpb.CheckKind
	}{
		{"R1-1", recordpb.CheckKind_CHECK_KIND_COMPUTATION},
		{"R1-2", recordpb.CheckKind_CHECK_KIND_DOCUMENT},
	} {
		if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{
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
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, g := range BoardJSONOf(b).Open {
		got[g.ID] = g.CheckKind
	}
	if got["R1-1"] != "computation" || got["R1-2"] != "document" {
		t.Errorf("board view check_kind = %v — a seat cannot know which gaps demand a program", got)
	}

	// The work list is the SCANNING read a seat plans its sitting from, so the demand's TYPE has
	// to be on it even though the rest of the acceptance check deliberately is not.
	got = map[string]string{}
	for _, g := range WorkJSONOf(b).Open {
		got[g.ID] = g.CheckKind
	}
	if got["R1-1"] != "computation" || got["R1-2"] != "document" {
		t.Errorf("work check_kind = %v — the read a seat plans from cannot say which gaps prose will not close", got)
	}
}

// AN EMPTY FRICTION LOG IS TWO DIFFERENT RUNS, and only one of them is fine.
func TestTheFrictionViewSeparatesSilenceFromAnAttestation(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "blue-respond-r1", Round: RoundOf("blue-respond-r1")}); err != nil {
		t.Fatal(err)
	}
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	j := FrictionJSONOf(b)
	if j.Counts.Total != 0 || j.Counts.Attested != 0 {
		t.Fatalf("a silent run: total=%d attested=%d, want 0/0", j.Counts.Total, j.Counts.Attested)
	}

	if _, err := Append(Identity{RunDir: runDir, SeatID: "blue-respond-r1", Round: RoundOf("blue-respond-r1")}, &recordpb.FrictionNone{Text: proto.String("read the board and my verb list; every refusal was my own error")}); err != nil {
		t.Fatal(err)
	}
	b, _ = BoardState(runDir)
	j = FrictionJSONOf(b)
	// The counts must now DIFFER from the silent run. Same total, different meaning — which is
	// the whole point: zero-with-an-attestation is a statement someone can be wrong about,
	// zero-alone is the absence of one.
	if j.Counts.Total != 0 {
		t.Errorf("an attestation must not count as a complaint: total=%d", j.Counts.Total)
	}
	if j.Counts.Attested != 1 || len(j.NothingBlocked) != 1 {
		t.Fatalf("the attestation did not reach the view: attested=%d entries=%d", j.Counts.Attested, len(j.NothingBlocked))
	}
	if j.NothingBlocked[0].SeatID != "blue-respond-r1" {
		t.Error("the attestation must name the seat that made it — an unattributed one cannot be weighed")
	}
}

// A PROPERTY IS NOT A DEBT.
//
// Projecting check_kind was necessary and not sufficient: with it visible, `prove` moved from
// 0 uses across eighteen sittings to 1 across nine. A seat reading `"check_kind": "computation"`
// learns something about the gap; it does not learn that IT owes a program, and only the second
// changes what the sitting produces.
func TestAwaitingProofTracksTheDebtAndAgreesWithTheGate(t *testing.T) {
	runDir := t.TempDir()
	for _, s := range []string{"red-merge-r1", "blue-respond-r1"} {
		if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: s, Round: RoundOf(s)}); err != nil {
			t.Fatal(err)
		}
	}
	mint := func(id string, kind recordpb.CheckKind) {
		t.Helper()
		if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{
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
	mint("R1-1", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	mint("R1-2", recordpb.CheckKind_CHECK_KIND_COMPUTATION)
	mint("R1-3", recordpb.CheckKind_CHECK_KIND_DOCUMENT)

	owed := GapsAwaitingProof(runDir)
	if len(owed) != 2 || owed[0] != "R1-1" || owed[1] != "R1-2" {
		t.Fatalf("owed = %v, want the two computation gaps in board order", owed)
	}
	// A document gap is never a proof debt — over-reporting would train seats to ignore it.
	for _, id := range owed {
		if id == "R1-3" {
			t.Error("a document-kind gap was reported as awaiting a computation")
		}
	}

	// THE PROOF ANSWERS A GAP, and the earlier conversion dropped `answers` — so the proof
	// discharged nothing and the debt could not move. `answers` is the whole join this test is
	// about: a proof that names no gap is a script that ran for no stated reason.
	if _, err := Append(Identity{RunDir: runDir, SeatID: "blue-respond-r1", Round: RoundOf("blue-respond-r1")}, &recordpb.Proof{
		Answers: proto.String("R1-1"),
		Script:  proto.String("s.py"),
	}); err != nil {
		t.Fatal(err)
	}
	if owed := GapsAwaitingProof(runDir); len(owed) != 1 || owed[0] != "R1-2" {
		t.Fatalf("after proving R1-1, owed = %v, want [R1-2]", owed)
	}

	// THE BOARD AND THE GATE MUST NOT DISAGREE about what is owed. They share one join and one
	// constant precisely so a seat cannot be told it owes nothing by the read and be refused by
	// the write — three bare "computation" literals used to make that possible.
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	fromBoard := map[string]bool{}
	for _, g := range BoardJSONOf(b).Open {
		if g.AwaitingProof {
			fromBoard[g.ID] = true
		}
	}
	if len(fromBoard) != 1 || !fromBoard["R1-2"] {
		t.Fatalf("board says %v awaits proof, the debt query says [R1-2]", fromBoard)
	}
	for _, g := range WorkJSONOf(b).Open {
		if g.AwaitingProof != fromBoard[g.ID] {
			t.Errorf("work and board disagree on %s: %v vs %v", g.ID, g.AwaitingProof, fromBoard[g.ID])
		}
	}

	// A CLOSED gap owes nothing, whatever its kind: the debt is what blue can still act on.
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Close{
		GapId:        proto.String("R1-2"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("x"),
		Prose:        proto.String("the computation ran and the check holds"),
	}); err != nil {
		t.Fatal(err)
	}
	if owed := GapsAwaitingProof(runDir); len(owed) != 0 {
		t.Errorf("a closed gap is still reported as owed: %v", owed)
	}
}
