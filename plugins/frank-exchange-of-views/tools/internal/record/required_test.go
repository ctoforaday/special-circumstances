package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

import "testing"

// THE TABLE MUST BE TRUE.
//
// RequiredFields is what the CLI marks REQUIRED in a seat's help, and help is the seat's
// whole contract. A table that claims a field is required when validate would accept it
// missing teaches the seat to pass something it does not need; a table that omits a real
// requirement leaves the seat to discover it by failing. Either way the contract lies.
//
// So this is behavioural, not structural: for every field the table declares, a payload
// missing exactly that field must actually be REFUSED by validate.

// runWithGap is a run directory in which R1-1 actually exists. Since references are now
// checked at write time, a fixture that names a gap must create it — which is the point.
// seatFor picks a seat of the role that owns each verb, since validate now resolves
// round-scoped references and needs to know who is writing.
func seatFor(typ string) string {
	switch typ {
	case "opinion", "halt", "certify":
		return "judge-r1"
	case "retire", "line-of-inquiry", "manifest-row", "revision", "confidence":
		return "blue-respond-r1"
	default:
		return "red-merge-r1"
	}
}

func runWithGap(t *testing.T) string {
	t.Helper()
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{AcceptanceCheck: proto.String("the check runs"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}
	return runDir
}

// review_flag is required by PRESENCE, not by being non-empty, and that distinction is
// load-bearing: `--review-flag false` is a legitimate ruling ("no, a human need not look
// at this"), and a generic present-and-non-empty check would refuse it. Three separate
// defects in this codebase have come from treating a falsy value as an absent one.
func TestAFalsyReviewFlagSatisfiesTheRequirement(t *testing.T) {
	// `review_flag` is a STRING field carrying the seat's answer; "false" is an answer, and the
	// requirement is satisfied by the field being SET, not by it being non-empty.
	o := &recordpb.Opinion{
		GapId: proto.String("R1-1"), Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED),
		Principle: proto.String("p"), Tension: proto.String("t"),
		ReviewFlag: proto.String("false"), Rationale: proto.String("r"),
	}
	if err := validate(runWithGap(t), "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, o); err != nil {
		t.Errorf("a legitimately falsy review_flag was treated as missing: %v", err)
	}
}

// A CARRY IS A LINEAGE CLAIM. --carried-from was presence-only, so any value satisfied it
// and satisfying it skipped the anchor — a fresh, unverified closure wearing a carry's
// clothes, counted as closed by every projection and by anchored_closures_pct. The help
// offers it exactly where a seat that cannot produce an anchor will read it.
func TestCarriedFromCannotLaunderAnUnanchoredFirstClosure(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	mint := &recordpb.Mint{
		GapId: proto.String(id), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Problem: proto.String("p"),
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, mint); err != nil {
		t.Fatal(err)
	}

	// No prior closure exists, so a carry is a false claim about the record.
	carry := &recordpb.Close{GapId: proto.String(id), CarriedFrom: proto.String("1"), Prose: proto.String("verified at the leaf")}
	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, carry); err == nil {
		t.Error("an unanchored FIRST closure was accepted as a carry — that is the laundering path: no verification, no lineage, and it scores as closed")
	}

	// Anchored, it goes through — the escape hatch is closed, not the door.
	anchored := &recordpb.Close{
		GapId: proto.String(id), AnchorSeat: proto.String("L1"),
		AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"),
		Prose: proto.String("verified and holds"),
	}
	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, anchored); err != nil {
		t.Errorf("an anchored closure must still be accepted: %v", err)
	}
}

// And a GENUINE carry still works: close once with an anchor, then restate it.
func TestAGenuineCarryIsStillAccepted(t *testing.T) {
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{AcceptanceCheck: proto.String("the check runs"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Close{GapId: proto.String(id), AnchorTool: proto.String("go test"), AnchorTarget: proto.String("./x"), Prose: proto.String("verified at the leaf")}); err != nil {
		t.Fatal(err)
	}
	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{GapId: proto.String(id), CarriedFrom: proto.String("1"), Prose: proto.String("verified at the leaf")}); err != nil {
		t.Errorf("a carry restating a real earlier closure must be accepted: %v", err)
	}
}

// likelihood and impact MULTIPLY into GapMass, and an absent grade contributes zero — so
// an ungraded gap reads as harmless rather than ungraded. Severity and cx are reported
// rather than multiplied, so their absence is visible and they stay optional.
func TestMintRequiresTheGradesThatMultiplyIntoMass(t *testing.T) {
	base := func() *recordpb.Mint {
		return &recordpb.Mint{
			GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Class:     proto.String("scope-creep"), Problem: proto.String("p"),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}
	}
	// The two grades that MULTIPLY into mass, cleared one at a time. Clearing is `nil`, not the
	// zero value: an unset enum and one set to UNSPECIFIED are the same on the wire, and the
	// requirement is presence.
	for _, c := range []struct {
		name  string
		clear func(*recordpb.Mint)
	}{
		{"likelihood", func(m *recordpb.Mint) { m.Likelihood = nil }},
		{"impact", func(m *recordpb.Mint) { m.Impact = nil }},
	} {
		m := base()
		c.clear(m)
		if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, m); err == nil {
			t.Errorf("mint without --%s was accepted; its mass computes to ZERO and the gap sinks to the bottom of every ranking as though it were harmless", c.name)
		}
	}
	// Severity and cx remain optional: absent, they are SHOWN absent.
	if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, base()); err != nil {
		t.Errorf("severity and cx must stay optional — their absence is visible, not silently zero: %v", err)
	}
	if GapMass("", "medium") != 0 {
		t.Error("the premise of this rule has changed: an absent grade no longer contributes zero")
	}
}
