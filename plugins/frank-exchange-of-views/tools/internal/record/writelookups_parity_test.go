package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"google.golang.org/protobuf/proto"
)

// THE WRITE-PATH LOOKUPS, HELD AGAINST THE FOLDS THEY REPLACED — step 4's second group. These
// are the questions every mint, close and filing asks of the record before writing (the next
// free id, "did I already do this", "does the thing I name exist"), so the fixture exercises the
// edges those folds owned: a re-registered agent whose LAST binding wins, a moved line of
// inquiry that must not count as a proposal, a key lookup that must return the FIRST match, and
// a marker id the torn-splice heal asks about.
func TestWriteLookupsAgreeWithTheFoldsTheyReplaced(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-merge-r1", Round: 1}
	blue := Identity{Run: run, SeatID: "blue-respond-r1", Round: 1}

	// Two registers under one agent id: the binding is the most recent claim.
	t.Setenv(seatenv.AgentVar, "agent-007")
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "red-merge-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "blue-respond-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if seat, found, err := SeatOfAgent(run, "agent-007"); err != nil || !found || seat != "blue-respond-r1" {
		t.Errorf("SeatOfAgent = (%q, %v, %v), want the LATEST register to win", seat, found, err)
	}
	if _, found, err := SeatOfAgent(run, "agent-999"); err != nil || found {
		t.Errorf("SeatOfAgent on an unknown agent = (found=%v, %v)", found, err)
	}
	if seats, err := RegisteredSeats(run); err != nil || strings.Join(seats, ",") != "red-merge-r1,blue-respond-r1" {
		t.Errorf("RegisteredSeats = (%v, %v), want event order", seats, err)
	}

	// Mints: the id allocator counts THIS round's mints; the key lookup returns the first match.
	if id, err := MintGapID(run, 1); err != nil || id != "R1-1" {
		t.Errorf("MintGapID on an unminted round = (%q, %v)", id, err)
	}
	for _, gid := range []string{"R1-1", "R1-2"} {
		if _, err := Append(red, &recordpb.Mint{
			GapId:           proto.String(gid),
			MintKey:         proto.String("k-" + gid),
			Class:           proto.String("self-attestation"),
			Problem:         proto.String("p"),
			RequiredFix:     proto.String("f"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if id, err := MintGapID(run, 1); err != nil || id != "R1-3" {
		t.Errorf("MintGapID = (%q, %v), want R1-3 after two mints", id, err)
	}
	if id, err := MintGapID(run, 2); err != nil || id != "R2-1" {
		t.Errorf("MintGapID counts per ROUND: got (%q, %v)", id, err)
	}
	if id, err := ExistingMintByKey(run, "red-merge-r1", "k-R1-2"); err != nil || id != "R1-2" {
		t.Errorf("ExistingMintByKey = (%q, %v)", id, err)
	}
	if id, err := ExistingMintByKey(run, "blue-respond-r1", "k-R1-2"); err != nil || id != "" {
		t.Errorf("ExistingMintByKey must scope to the seat: got (%q, %v)", id, err)
	}
	if ids, err := allGapIDs(run); err != nil || len(ids) != 2 || !ids["R1-1"] || !ids["R1-2"] {
		t.Errorf("allGapIDs = (%v, %v)", ids, err)
	}

	// A close, then the prior-closure question a --carried-from claim is checked against.
	if _, err := Append(red, &recordpb.Close{GapId: proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
		AnchorSeat:   proto.String("L1"), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified at the leaf")}); err != nil {
		t.Fatal(err)
	}
	if rounds, err := priorClosureRounds(run, "R1-1"); err != nil || len(rounds) != 1 || rounds[0] != 1 {
		t.Errorf("priorClosureRounds = (%v, %v)", rounds, err)
	}
	if rounds, err := priorClosureRounds(run, "R1-2"); err != nil || rounds != nil {
		t.Errorf("priorClosureRounds on a never-closed gap = (%v, %v)", rounds, err)
	}

	// A proposal counts toward the next inquiry id; a MOVE of the same line must not.
	if _, err := Append(blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("a direction")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status:           recordpb.AvenueStatus_AVENUE_STATUS_PURSUED.Enum(),
		SupersedesStatus: proto.String(recordpb.Word(recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED))}); err != nil {
		t.Fatal(err)
	}
	if id, err := MintInquiryID(run); err != nil || id != "Q2" {
		t.Errorf("MintInquiryID = (%q, %v) — a move counted as a proposal, or the proposal was missed", id, err)
	}
	if err := RequireInquiryRef(run, "Q1"); err != nil {
		t.Errorf("RequireInquiryRef(Q1) = %v", err)
	}
	if err := RequireInquiryRef(run, "Q9"); err == nil {
		t.Error("RequireInquiryRef accepted a line of inquiry nobody proposed")
	}

	// Findings and their markers: the label allocator, the key lookup, the anchor pair heal.
	if _, err := Append(Identity{Run: run, SeatID: "red-lens-r1-evidence", Round: 1}, &recordpb.Finding{
		FindingId: proto.String("f-0a0a0a0a"), Label: proto.String("evidence-F1"), FindingKey: proto.String("fk-1"),
		Location: proto.String("L"), Text: proto.String("t"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}); err != nil {
		t.Fatal(err)
	}
	if label, err := NextFindingLabel(run, "red-lens-r1-evidence"); err != nil || label != "evidence-F2" {
		t.Errorf("NextFindingLabel = (%q, %v)", label, err)
	}
	if label, id, err := FindingByKey(run, "red-lens-r1-evidence", "fk-1"); err != nil || label != "evidence-F1" || id != "f-0a0a0a0a" {
		t.Errorf("FindingByKey = (%q, %q, %v)", label, id, err)
	}
	if err := requireFindings(run, []string{"evidence-F1"}, "mint", "--found-by"); err != nil {
		t.Errorf("requireFindings on a recorded label = %v", err)
	}
	if err := requireFindings(run, []string{"evidence-F1", "L9-F9"}, "mint", "--found-by"); err == nil || !strings.Contains(err.Error(), "L9-F9") {
		t.Errorf("requireFindings must name the first missing label: %v", err)
	}
	if exists, err := AnchorEventExists(run, "f-0a0a0a0a"); err != nil || exists {
		t.Errorf("AnchorEventExists = (%v, %v) — no anchor event was appended for this finding", exists, err)
	}

	if err := requireSeat(run, "red-merge-r1", "rule", "--by"); err != nil {
		t.Errorf("requireSeat on a seated seat = %v", err)
	}
	if err := requireSeat(run, "judge-r9", "rule", "--by"); err == nil {
		t.Error("requireSeat accepted a seat that never recorded anything")
	}

	if ClassCoinedInRun(run, "never-coined") {
		t.Error("ClassCoinedInRun found a class nobody coined")
	}
}
