package record

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// Replay is where the log becomes state. Winner selection, dedup and the anomaly
// list are the parts that must never silently normalize: an anomaly the render
// hides is a divergence nobody can explain afterwards.

// SEVEN TESTS AND THE `ev` HELPER ARE GONE, and every one named the shard layout.
//
//   - The three ReadShard tests: an unparseable line dropped without losing the rest, a missing
//     file read as empty rather than an error, a missing final newline handled. There are no lines
//     and no files.
//   - TestMergedEventsWinnerSelection and TestMultiNonceSeparatesACrashRetryFromLostWork: replay
//     picked ONE shard per seat, and the second told a healthy re-dispatch from one that lost work.
//     Nothing is discarded now — both sittings are rows — so neither can occur.
//   - TestMergedEventsGlobalOrderIsDeterministic: the order was reconstructed from (TS, SeatID,
//     Seq) because merging per file had lost it. It is `ORDER BY id` now.
//   - TestMergedEventsDedupByKeyExceptRegister: two events sharing a key were deduped on READ.
//     `events.key` is UNIQUE, so the second write is refused rather than silently dropped later.
//
// `writeShard` survives as a SEEDING helper — its call sites read it as "put these events in this
// run" — and takes neither a seat nor a nonce, because the record has neither.
func writeShard(t *testing.T, runDir string, evs []*Event) {
	t.Helper()
	recordtest.Seed(t, runDir, evs...)
}

func TestMergedEventsOnAnEmptyOrAbsentRun(t *testing.T) {
	m, err := MergedEvents(mustRun(t, filepath.Join(recordtest.TmpRun(t), "no-such-run")))
	if err != nil {
		t.Fatalf("absent run dir = %v, want no error", err)
	}
	if len(m.Events) != 0 {
		t.Errorf("absent run produced state: %+v", m)
	}

	runDir := newRun(t)
	if err := os.MkdirAll(recordsDirT(runDir), 0o755); err != nil {
		t.Fatal(err)
	}
	// Files that are not shards must be ignored, not parsed.
	for _, name := range []string{"ledger.md", ".lock-render", "class-registry.json", "events-bad.jsonl", "events-x-nothex.jsonl"} {
		if err := os.WriteFile(filepath.Join(recordsDirT(runDir), name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	m, err = MergedEvents(mustRun(t, runDir))
	if err != nil {
		t.Fatalf("a records dir of non-shards = %v, want no error", err)
	}
	if len(m.Events) != 0 {
		t.Errorf("a non-shard file was parsed as a shard: %+v", m.Events)
	}
}

func TestGapMassAndGradeStr(t *testing.T) {
	cases := []struct {
		likelihood, impact string
		want               float64
	}{
		{"medium", "high", 6},
		{"low", "low", 1},
		{"certain", "certain", 12.25},
		{"trivial", "trivial", 0.25},
		// realized is off-scale at 0: it already happened, so it carries no mass.
		{"realized", "high", 0},
		{"high", "realized", 0},
		// An unknown or absent grade contributes zero rather than erroring.
		{"", "high", 0},
		{"bogus", "high", 0},
		{"", "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.likelihood+"x"+tc.impact, func(t *testing.T) {
			if got := GapMass(tc.likelihood, tc.impact); got != tc.want {
				t.Errorf("GapMass(%q,%q) = %v, want %v", tc.likelihood, tc.impact, got, tc.want)
			}
		})
	}

	// GradeStr's GUARD IS THE TYPE NOW. It took `any` and returned "" for a nil, a bool, a number
	// or a slice — the shapes a payload map could hold in a grade key. A Grade is an enum, so the
	// only non-grade it can carry is the UNSPECIFIED zero, which is the ungraded case and must
	// still render as the empty word rather than as a grade.
	if got := GradeStr(recordpb.Grade_GRADE_UNSPECIFIED); got != "" {
		t.Errorf("an ungraded value renders %q, want empty", got)
	}
	if got := GradeStr(recordpb.Grade_GRADE_HIGH); got != "high" {
		t.Errorf("GradeStr(high) = %q", got)
	}
}

// A gap id names ONE gap over the run: one namespace, minted tool-side with no count of what the
// record holds. The collision class that produced four different ids for one gap cannot recur, and
// the chair sitting again opens no fresh namespace — there is no round left for one to belong to
// (plans/roundless.md §III.A.3). `mint.gap_id` is UNIQUE, so an id recorded twice is a refused
// write, before the chair re-sits and after.
func TestGapIdsAreRunGlobalAndNeverRoundShaped(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: seatID}, "", ""); err != nil {
		t.Fatal(err)
	}
	mint := func(id string) error {
		_, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seatID}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")})
		return err
	}
	shape := flags.GapID().Shape()
	var ids []string
	for i := 1; i <= 3; i++ {
		got := NewID("gap")
		if !shape.MatchString(got) {
			t.Fatalf("NewID(gap) = %q, which the gap-id shape %s refuses", got, shape)
		}
		if slices.Contains(ids, got) {
			t.Fatalf("NewID(gap) = %q a second time", got)
		}
		// The MINTED id, not a fixed one: each pass records the gap it just reserved.
		if err := mint(got); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, got)
	}
	if err := mint(ids[0]); err == nil {
		t.Errorf("%s was minted twice in one sitting — one id, two gaps", ids[0])
	}
	// The chair sits again. A fresh epoch is not a fresh namespace: an id spent before is still spent.
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: seatID}, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := mint(ids[1]); err == nil {
		t.Errorf("%s was minted again after the chair re-sat — the namespace is run-global", ids[1])
	}
	got := NewID("gap")
	if err := mint(got); err != nil {
		t.Fatalf("a fresh id after the chair re-sat was refused: %v", err)
	}
	have, err := allGapIDs(recordOf(t, mustRun(t, runDir)))
	if err != nil || len(have) != 4 {
		t.Fatalf("allGapIDs = (%v, %v), want the four minted", have, err)
	}
	for _, id := range append(ids, got) {
		if !have[id] {
			t.Errorf("%s is not on the record", id)
		}
	}
}

// --key makes a retried mint idempotent: a seat whose message died after a
// successful mint must get the EXISTING id, not a second gap.
func TestExistingMintByKey(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: seatID}, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seatID}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1"), MintKey: proto.String("L1-F3"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}

	got, err := ExistingMintByKey(mustRun(t, runDir), seatID, "L1-F3")
	if err != nil {
		t.Fatal(err)
	}
	if got != "G1" {
		t.Errorf("ExistingMintByKey = %q, want G1", got)
	}
	// An empty key is not a lookup: it must never match a mint that had no key.
	if got, err := ExistingMintByKey(mustRun(t, runDir), seatID, ""); err != nil || got != "" {
		t.Errorf("empty key = (%q,%v), want empty and no error", got, err)
	}
	// A different key, and a different SEAT with the same key, are both misses:
	// the label is stable per seat, not globally.
	if got, err := ExistingMintByKey(mustRun(t, runDir), seatID, "L9-F9"); err != nil || got != "" {
		t.Errorf("unknown key = (%q,%v), want a miss", got, err)
	}
	if got, err := ExistingMintByKey(mustRun(t, runDir), "red-lens-evidence", "L1-F3"); err != nil || got != "" {
		t.Errorf("another seat's key matched: %q — mint keys are per-seat", got)
	}
}

func TestBoardStateReplaysGapLifecycle(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, seatID+":register:#1", &recordpb.Register{}),
		recordtest.At(t, seatID, seatID+":mint:G1", &recordpb.Mint{GapId: proto.String("G1"), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("p1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, seatID, seatID+":mint:G2", &recordpb.Mint{GapId: proto.String("G2"), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p2"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		// A regrade moves ONLY the keys it carries.
		recordtest.At(t, seatID, seatID+":regrade:G1", &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_CERTAIN), Basis: proto.String("new evidence")}),
		recordtest.At(t, seatID, seatID+":close:G2", &recordpb.Close{GapId: proto.String("G2"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")}),
		// A REGRADE AND A CLOSE OF AN UNKNOWN GAP USED TO BE SEEDED HERE, and the assertion was
		// that the replay IGNORED them rather than failing. Both are foreign keys onto the mint
		// now, so neither row can be written — the state is unrepresentable, and the replay's arm
		// for it is a hard error rather than a skip. Seeding them would fail the
		// fixture, not the assertion.
	})
	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Gaps) != 2 || b.Gaps[0].ID != "G1" {
		t.Fatalf("family order = %v, want mint order [G1 G2]", b.Gaps)
	}
	g1 := b.Gap("G1")
	if !g1.Open {
		t.Error("G1 should still be open")
	}
	if g1.Severity != recordpb.Grade_GRADE_CERTAIN {
		t.Errorf("regrade did not move severity: %v", g1.Severity)
	}
	// The keys the regrade did NOT carry are untouched.
	if g1.Likelihood != recordpb.Grade_GRADE_LOW || g1.Impact != recordpb.Grade_GRADE_LOW {
		t.Errorf("regrade clobbered a grade it did not carry: likelihood=%v impact=%v", g1.Likelihood, g1.Impact)
	}
	if len(g1.Regrades) != 1 {
		t.Errorf("regrade history not kept: %d entries", len(g1.Regrades))
	}
	g2 := b.Gap("G2")
	if g2.Open || !g2.HasClosed || g2.ClosedEpoch != 1 {
		t.Errorf("G2 closure not replayed: %+v", g2)
	}
	// Absence is renderable: a gap minted without --cx keeps nil, not "".
	if g2.ComplexityCost != recordpb.Grade_GRADE_UNSPECIFIED {
		t.Errorf("an unpassed grade became %v, want nil (it renders as \"undefined\")", g2.ComplexityCost)
	}
	// G3 was never minted, so it is not on the board — which is now true by construction rather
	// than by the replay choosing to skip it.
	if b.Gap("G3") != nil {
		t.Error("an event about an unknown gap created one")
	}
}

// FINDINGS REPLAY; DISPOSALS NO LONGER EXIST. `observe` and `dispose` are retired (#327), and
// with them the label-matching that resolved a disposal to its target — a mechanism that once
// attached 39 of 60 disposals by accident of ordering because 15 labels were reused across lens
// seats. A finding is now addressed by COALESCENCE alone: its label named in a gap's found_by.
//
// What must still hold is that every finding lands on the board with its identity intact, since
// the credit join is keyed on exactly that.
func TestBoardStateReplaysFindingsWithTheirLabels(t *testing.T) {
	runDir := newRun(t)
	lens := "red-lens-evidence"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, lens, lens+":finding:F-f0000001", &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("first")}),
		recordtest.At(t, lens, lens+":finding:F-f0000002", &recordpb.Finding{Id: proto.String("F-f0000002"), Text: proto.String("second")}),
	})
	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	obs := ObservationsOf(b.Events)
	if len(obs) != 2 {
		t.Fatalf("both findings must replay onto the board, got %d", len(obs))
	}
	for _, want := range []string{"F-f0000001", "F-f0000002"} {
		found := false
		for _, o := range obs {
			if o.Finding.GetId() == want {
				found = true
			}
		}
		if !found {
			t.Errorf("finding %s lost its id in replay — found_by credit is keyed on it", want)
		}
	}
}

// THE GRADE ENUM IS ENFORCED BY THE TYPE NOW, and this test says what it used to catch.
//
// It drove `validate` with `severity=catastrophic` and `severity=true`, asserting a refusal that
// named the field and the offending value — the guard for a payload map that could hold any string
// or any bool in a grade key. A Grade is a proto enum: `catastrophic` cannot be constructed, and
// neither can a bool. There is nothing left for a runtime check to refuse.
//
// What survives is the ABSENCE arm, which is not about the type: an absent graded field is fine,
// and only a present-and-wrong one was ever refused. Absence is still expressible (it is the
// UNSPECIFIED zero), so it is still asserted — and the canonical set is checked against the schema
// in enums_test rather than by writing every value through validate.
func TestAnAbsentGradeIsAccepted(t *testing.T) {
	if err := validateRun(mustRun(t, recordtest.TmpRun(t)), "red-chair", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}); err != nil {
		t.Errorf("absent grades were refused: %v", err)
	}
}

// validateContract is one write-path contract: a body, and what validate must say about it.
type validateContract struct {
	name    string
	typ     recordpb.EventType
	p       proto.Message
	wantErr string // empty means it must be ACCEPTED
}

// validateContractCases is the shared table. Extracted from TestValidateVerbContracts so
// TestNoRecordRefusalNamesAFlagASeatCannotType reads the SAME cases: a second hand-kept list of
// incomplete bodies would drift from this one, and the drift would show up as a flag-word gate
// that quietly stopped covering half the verbs.
func validateContractCases() []validateContract {
	return []validateContract{
		{"mint without --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1"), Problem: proto.String("p"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum()}, "requires --check"},
		{"mint with an empty --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1"), Problem: proto.String("p"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), AcceptanceCheck: proto.String("")}, "requires --check"},
		{"mint without --class", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1"), Problem: proto.String("p"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT)}, "requires --class"},
		{"mint complete", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}, ""},

		{"close without --id", recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{}, "requires --id"},
		{"regrade without --basis", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{}, "requires --reason"},
		{"regrade complete", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}, ""},

		// `--quote`, NOT `--claim`: the field is `claim` and the flag a seat types is `--quote`,
		// which is what the annotation now declares and what internal/cli/blue/retire.go registers.
		// This expectation was the field name, so it passed while the message and the parser
		// agreed with each other and disagreed with the test — a refusal naming a flag nobody can
		// pass is the class this branch has found four times, and the test asserted it.
		{"retire without --quote", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Reason: proto.String("r")}, "requires --quote"},
		{"retire without --reason", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c")}, "requires --reason"},
		{"retire complete", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c"), Reason: proto.String("r")}, ""},

		// An unknown status is UNREPRESENTABLE: InquiryStatus is an enum, so `shelved` cannot be built.
		// What remains is the ABSENT case below, which the type cannot refuse.
		{"an inquiry with no status at all", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Question: proto.String("l")}, "requires --as"},
		{"an inquiry with no id", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("l")}, "requires an id"},
		{"a deferred inquiry needs a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED), Question: proto.String("l")}, "requires --reason"},
		// THE REFUSAL NAMES --question. The payload key is `question` and the flag that fills it is
		// --question; the refusal derives the flag name, so it can only name a flag a surface has.
		{"an inquiry proposed without a question", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY)}, "requires --question"},
		{"a declined inquiry needs a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED), Question: proto.String("l")}, "requires --reason"},
		{"an abandoned inquiry needs a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED), Question: proto.String("l")}, "requires --reason"},
		{"a refuted inquiry needs a reason, and is asked for what the approach found", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Question: proto.String("l")}, "a refuted inquiry requires --reason — what the approach found"},
		{"a confirmed inquiry is asked for what the approach found", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED), Question: proto.String("l")}, "a confirmed inquiry requires --reason — what the approach found"},
		{"a refuted inquiry with a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Question: proto.String("l"), Reason: proto.String("what it found")}, ""},
		{"an underway inquiry does not need a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Question: proto.String("l")}, ""},
		{"a declined inquiry with a reason", recordpb.EventType_EVENT_TYPE_INQUIRY, &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED), Question: proto.String("l"), Reason: proto.String("why")}, ""},

		// The message must name the flag the PARSER accepts. It named --gap-id for as
		// long as that flag existed and kept naming it after the rename, because the
		// spelling was derived from the payload key rather than stated.
		{"a motion ruling missing every field", recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{}, "requires --id"},
		// AN UNKNOWN VERB IS UNREPRESENTABLE: the type is an enum and the body is a message, so there
		// is no "no-such-verb" to pass. The arm that ignored it is gone with the string.
	}
}

func TestValidateVerbContracts(t *testing.T) {
	for _, tc := range validateContractCases() {
		t.Run(tc.name, func(t *testing.T) {
			// An inquiry is blue's to propose: the chair rules on them and files none.
			seat := "red-chair"
			if tc.typ == recordpb.EventType_EVENT_TYPE_INQUIRY {
				seat = "blue-respond"
			}
			err := validateRun(mustRun(t, newRun(t)), seat, tc.typ, tc.p)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("validate = %v, want accepted", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validate accepted %s with %v", recordpb.Word(tc.typ), tc.p)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// docketRunDir is a run in which G1 exists AND motion M1 has been docketed against it, so a
// bench ruling's references both resolve and the test can be about the missing FIELD rather than
// the missing gap or the missing filing.
//
// IT TAKES BOTH EVENTS, and that is the shape of the verb now. The gap rides the FILING and the
// disposition rides the RULING, so a fixture that seeds only one of them leaves the other end of
// the join empty — and an empty join reads exactly like a ruling nobody made.
func docketRunDir(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, "", ""); err != nil {
		t.Fatal(err)
	}
	id := NewID("gap")
	redID := Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}
	if _, err := Append(redID, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), AcceptanceCheck: proto.String("the check runs"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(redID, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Basis:    proto.String("red cannot settle this and puts it to the bench"),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(id)}},
	}); err != nil {
		t.Fatal(err)
	}
	return runDir
}

// The bench's ruling demands every field its schema marks required, and names the one that is
// missing with the flag spelling the seat actually typed.
//
// The expected spellings are LITERALS. This test used to compute them with the same
// underscore-to-hyphen transform the code under test used, which made it a tautology: it
// asserted the code agreed with itself, and passed happily while the message taught
// --gap-id and --disposition, two flags the parser had stopped accepting. A test that
// reimplements its subject cannot indict it.
//
// IT IS TWO EVENTS NOW, and the split is why the census is split with it. `gap_id` is the
// FILING's field — a docket motion that names no gap is an escalation of nothing — and the five
// reasoning fields are the RULING's. Asserting them all against one body would be asserting a
// shape the record no longer has.
func TestValidateDocketRulingNamesEachMissingField(t *testing.T) {
	// EACH FIELD CLEARED IN TURN, on a body that is otherwise complete. Clearing is `nil`, which
	// is what "the seat never passed it" means — a `proto.String("")` would SATISFY a presence
	// requirement, because an empty answer is an answer for `settled`, the one field that allows it.
	complete := func() *recordpb.MotionRule {
		return &recordpb.MotionRule{
			MotionId: proto.String("M1"),
			Opinion:  proto.String("the ruling's reasoning"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
				Principle:   proto.String("x"),
				Tension:     proto.String("x"),
				ReviewFlag:  proto.String("x"),
				Settled:     proto.String("the claim as it stood may not be re-asserted"),
				Final:       proto.Bool(true),
			}},
		}
	}
	for _, c := range []struct {
		field, wantFlag string
		clear           func(*recordpb.MotionRule)
	}{
		// The motion must EXIST and be NAMED CORRECTLY: references are checked at write time, so
		// motion_id carries the real filed id rather than a placeholder. With a bogus id the
		// reference check fires first and this would assert on the wrong refusal.
		{"motion_id", "--id", func(r *recordpb.MotionRule) { r.MotionId = nil }},
		{"opinion", "--reason", func(r *recordpb.MotionRule) { r.Opinion = nil }},
		{"disposition", "--as", func(r *recordpb.MotionRule) { r.GetDocket().Disposition = nil }},
		{"principle", "--principle", func(r *recordpb.MotionRule) { r.GetDocket().Principle = nil }},
		{"tension", "--tension", func(r *recordpb.MotionRule) { r.GetDocket().Tension = nil }},
		{"review_flag", "--review-flag", func(r *recordpb.MotionRule) { r.GetDocket().ReviewFlag = nil }},
		{"settled", "--settled", func(r *recordpb.MotionRule) { r.GetDocket().Settled = nil }},
	} {
		t.Run("missing "+c.field, func(t *testing.T) {
			o := complete()
			c.clear(o)
			err := validateRun(mustRun(t, docketRunDir(t)), "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, o)
			if err == nil {
				t.Fatalf("a docket ruling was accepted without %s", c.field)
			}
			if !strings.Contains(err.Error(), c.wantFlag) {
				t.Errorf("error %q does not name %q — the seat's only teacher is this string", err, c.wantFlag)
			}
		})
	}

	// THE FILING OWNS THE GAP, so its own omission is asserted on the filing.
	t.Run("a docket filing missing the gap", func(t *testing.T) {
		err := validateRun(mustRun(t, docketRunDir(t)), "red-chair", recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
			MotionId: proto.String("M2"),
			Basis:    proto.String("put it to the bench"),
			Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{}},
		})
		if err == nil {
			t.Fatal("a docket motion naming no gap was accepted — an escalation of nothing")
		}
		if !strings.Contains(err.Error(), "--id") {
			t.Errorf("error %q does not name --id", err)
		}
	})

	if err := validateRun(mustRun(t, docketRunDir(t)), "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, complete()); err != nil {
		t.Errorf("a complete docket ruling was refused: %v", err)
	}

	// AN EMPTY VALUE STILL COUNTS AS PRESENT for `settled` alone, the one field checked by
	// presence: a ruling may bar no proposition. TestTheBenchStatesItsRuleItsTensionAndItsReviewFlag
	// holds the fields an empty value does not satisfy.
	empty := complete()
	empty.GetDocket().Settled = proto.String("")
	if err := validateRun(mustRun(t, docketRunDir(t)), "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, empty); err != nil {
		t.Errorf("a docket ruling with settled present-but-empty was refused: %v", err)
	}
}

// Lineage is never dangling: a mint that supersedes an id no mint created is
// refused, because the supersedes chain is what the whole analysis reads.
func TestValidateRefusesDanglingLineage(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, seatID+":mint:G1", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1")}),
	})
	base := func(supersedes ...string) *recordpb.Mint {
		return &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			GapId: proto.String("G2"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Class:     recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Supersedes: supersedes,
		}
	}

	if err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_MINT, base("G1")); err != nil {
		t.Errorf("a real ancestor was refused: %v", err)
	}
	err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_MINT, base("G1", "G3"))
	if err == nil {
		t.Fatal("a dangling ancestor was accepted")
	}
	if !strings.Contains(err.Error(), "G3") || !strings.Contains(err.Error(), "dangling lineage refused") {
		t.Errorf("error must name the dangling id: %v", err)
	}
	// An empty lineage is not a dangling one.
	if err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_MINT, base()); err != nil {
		t.Errorf("an empty lineage was refused: %v", err)
	}
}

func TestValidateCloseAnchorContract(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-chair"
	// BOTH gaps are minted: G1 is the one being closed, G2 the successor a
	// repaired_with_regression names. A successor is a reference like any other and is
	// now checked, so a fixture that names one has to create it.
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, seatID+":mint:G1", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G1")}),
		recordtest.At(t, seatID, seatID+":mint:G2", &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("G2")}),
	})
	anchored := func() *recordpb.Close {
		return &recordpb.Close{
			GapId: proto.String("G1"), AnchorSeat: proto.String("L1"),
			AnchorTool: proto.String("git show"), AnchorTarget: proto.String("7bc501e:path"),
			Prose: proto.String("verified against the ref; the claim now holds"),
		}
	}
	withClass := func(c recordpb.Disposition, successor string) *recordpb.Close {
		a := anchored()
		a.ClosureClass = c.Enum()
		if successor != "" {
			a.Successor = proto.String(successor)
		}
		return a
	}
	cases := []struct {
		name    string
		p       proto.Message
		wantErr string
	}{
		// EACH CASE CARRIES ITS PROSE, because the unconditional requirements are checked FIRST
		// now — `CheckRequired` runs before any verb's own rules, from the annotation on the
		// field. A fixture omitting --reason would be refused for that and never reach the rule
		// it names. The ordering is a deliberate change: the old code ran requiredness LAST so
		// the more specific refusal led, and this reverses it.
		{"unknown gap", &recordpb.Close{GapId: proto.String("G3"), Prose: proto.String("verified")}, "--id names gap G3, which no mint event created"},
		{"no anchor at all", &recordpb.Close{GapId: proto.String("G1"), Prose: proto.String("verified")}, "requires the verification triple"},
		{"a PARTIAL anchor is not an anchor", &recordpb.Close{GapId: proto.String("G1"), Prose: proto.String("verified"), AnchorSeat: proto.String("L1")}, "requires the verification triple"},
		{"anchor missing its target", &recordpb.Close{GapId: proto.String("G1"), Prose: proto.String("verified"), AnchorSeat: proto.String("L1"), AnchorTool: proto.String("git show")}, "requires the verification triple"},
		{"a full anchor", anchored(), ""},
		// --carried-from remains the honest alternative to re-verifying, but it is a
		// CLAIM ABOUT THE RECORD and is now checked like one. This gap has never been
		// closed, so "carried from round 2" is simply false — and accepting it was a
		// laundering path: an unanchored first closure that scores as closed, which is
		// exactly what anchored_closures_pct exists to detect. The genuine case (close
		// with an anchor, then restate) is covered in required_test.go.
		{"--carried-from cannot invent an earlier closure", &recordpb.Close{GapId: proto.String("G1"), Prose: proto.String("verified"), CarriedFrom: proto.String("2")}, "no closure of it exists in the record"},
		{"closed_with_regression needs a successor", withClass(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION, ""), "requires --superseded-by"},
		{"closed_with_regression with a successor", withClass(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION, "G2"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_CLOSE, tc.p)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("validate = %v, want accepted", err)
				}
				return
			}
			if err == nil {
				t.Fatal("validate accepted the close")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// THE KEY'S THREE ARMS, and the one that moved.
//
// deriveKey chooses: a singleton verb keys on seat+verb, a body carrying a label keys on it, and
// anything else falls to an ORDINAL. The first two are pure and are tested here. The third counts
// the seat's prior events of that type, and it counts them IN THE WRITING TRANSACTION now — the
// slice of `prior` events this test used to pass does not exist, because counting in Go from a
// list read earlier is a read-then-write with a gap in it. Its behaviour is exercised through the
// real write path (TestDeriveKeySameLabelNextRoundIsNotACollision below, and the cli suite).
//
// One case is gone rather than moved: "a non-string label falls through" passed `true` where a
// label was expected. Every label field is a string in the schema, so there is no non-string to
// fall through.
func TestTheKeyLabelIsTheFirstFieldTheBodyCarries(t *testing.T) {
	for _, c := range []struct {
		name string
		body proto.Message
		want string
	}{
		{"gap_id is the first label consulted", &recordpb.Close{GapId: proto.String("G1")}, "G1"},
		{"a finding's id when there is no gap_id", &recordpb.Finding{Id: proto.String("F-f0000001")}, "F-f0000001"},
		{"url is a label too", &recordpb.Cite{SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED.Enum(), WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum(), SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum(), Url: proto.String("https://x")}, "https://x"},
		{"an empty label is not a label", &recordpb.Finding{Id: proto.String("")}, ""},
		{"no label at all", &recordpb.Log{}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := keyLabel(c.body); got != c.want {
				t.Errorf("keyLabel = %q, want %q", got, c.want)
			}
		})
	}
}

// A SINGLETON VERB KEYS ON SEAT+VERB, so a second one collides rather than accumulating.
func TestSingletonVerbsAreDeclaredForTheVerbsThatAreOnce(t *testing.T) {
	for _, typ := range []recordpb.EventType{
		recordpb.EventType_EVENT_TYPE_POSITION,
		recordpb.EventType_EVENT_TYPE_VERDICT,
		recordpb.EventType_EVENT_TYPE_SPOT_CHECK,
	} {
		if !singleton[typ] {
			t.Errorf("%s is not declared a singleton; a seat could record two and both would stand",
				recordpb.Word(typ))
		}
	}
	// REGISTER IS NOT ONE, and that is the change: two registers for a seat are a legitimate
	// re-dispatch. It was a singleton with its key overridden to carry the nonce; the nonce is
	// gone, so the exception is gone with it and a re-dispatch gets `#2` from the ordinary rule.
	if singleton[recordpb.EventType_EVENT_TYPE_REGISTER] {
		t.Error("register is declared a singleton — a re-dispatched seat would collide with its own first registration and could not record at all")
	}
}

// A FINDING ID NAMES ONE FINDING OVER THE WHOLE RUN. A seat sitting again that records under an id
// it already spent is repeating an act, not filing a new finding, and another lens recording under
// it would make one name two findings — every found_by credit on it a coin flip. `finding.id` is
// UNIQUE, so either repeat is a refused write rather than a second row; a fresh id lands.
func TestAFindingIDIsRefusedASecondTimeAndAFreshOneLands(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	id := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(id, &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("x")}); err != nil {
		t.Fatal(err)
	}
	// A second sitting of the same seat.
	if _, _, err := RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(id, &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("y")}); err == nil {
		t.Fatal("F-f0000001 landed twice for one lens — a finding id is run-unique, and the second is a repeated act")
	}
	other := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-logic"}
	if _, _, err := RegisterSeat(other, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(other, &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("z")}); err == nil {
		t.Fatal("F-f0000001 landed for a second lens — one id now names two findings")
	}
	if _, err := Append(id, &recordpb.Finding{Id: proto.String("F-f0000002"), Text: proto.String("y")}); err != nil {
		t.Fatalf("a fresh id was refused in the later sitting: %v", err)
	}
	m, err := MergedEvents(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d findings on the record, want 2 (F-f0000001 once, F-f0000002 once)", n)
	}
}

// TestMASSKeysAreExactlyTheCanonicalGrades holds MASS's key set to the grade scale
// (recordpb.GradeWords), the words the grade flag accepts: the mass mapping carries no grade the
// flag refuses and misses none it accepts.
func TestMASSKeysAreExactlyTheCanonicalGrades(t *testing.T) {
	got := make([]string, 0, len(MASS))
	for k := range MASS {
		got = append(got, k)
	}
	sort.Strings(got)
	want := recordpb.GradeWords()
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("MASS has %d grades, the scale has %d:\n got=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("MASS grade set diverges from the scale:\n got=%v\nwant=%v", got, want)
		}
	}
}

// refusalFlagToken matches a `--flag` reference in a refusal. The leading boundary keeps it out
// of anchor tokens like `<!--proof:p-…-->`; Go's regexp has no lookbehind, so the NAME is group 1.
var refusalFlagToken = regexp.MustCompile("(?:^|[\\s(\"'`])--([a-z][a-z0-9-]*)")

// EVERY FLAG A RECORD-LAYER REFUSAL NAMES IS A WORD A SEAT CAN ACTUALLY TYPE.
//
// FOUR INSTANCES ON ONE BRANCH: `--as supports-with-bridge` advertised and refused;
// `--id Q1 --as supported|…` left in help after the schema retired both flags; `retire requires
// --claim`, which is the FIELD name where the flag is `--quote`; and the fuzz typing an enum's
// underscore spelling at a hyphenated flag. Every one is two spellings of one fact across a
// boundary with one side moved, and every one reads perfectly well.
//
// CHECKED HERE, not only in internal/cli, because these messages are UNREACHABLE from the command
// line: cobra marks the same fields required and refuses first, so the CLI sweep never renders
// them. Measured — renaming `--check-kind` to `--kind-of-check` in this file's subject left both
// CLI gates green. A message no gate can see is exactly where a wrong flag word survives.
//
// The vocabulary is RequiredFields', which derives the word from the same annotation the write
// path uses. A message naming something outside it is either a renamed flag whose message did not
// move, or a field name where a seat types a flag word.
func TestNoRecordRefusalNamesAFlagASeatCannotType(t *testing.T) {
	// Flag words declared by the schema, across every event type.
	known := map[string]bool{}
	vals := recordpb.EventType(0).Descriptor().Values()
	for i := 0; i < vals.Len(); i++ {
		for _, rf := range RequiredFields(recordpb.Word(recordpb.EventType(vals.Get(i).Number()))) {
			known[rf.Flag] = true
		}
	}
	if len(known) < 10 {
		t.Fatalf("only %d flag words derived from the schema — an empty vocabulary accepts every message forever", len(known))
	}
	// Words no `required` annotation can produce: prose and identity channels every verb shares,
	// and the CONDITIONALLY required fields whose obligation depends on another field's value, so
	// no unconditional annotation declares them. Listed rather than pattern-matched, so adding one
	// is a decision someone makes on purpose.
	for _, w := range []string{
		"reason", "run", "seat-id", "json", "help", "id", "as",
		"quote", "new", "question", "hypothesis", "method", "problem", "fix", "check",
		"superseded-by", "supersedes", "ids", "none", "cites", "access-date", "relief", "class",
	} {
		known[w] = true
	}

	inspect := func(label string, err error) bool {
		if err == nil {
			return false
		}
		for _, m := range refusalFlagToken.FindAllStringSubmatch(err.Error(), -1) {
			if !known[m[1]] {
				t.Errorf("%s: the refusal names `--%s`, which is not a flag word any verb declares:\n\n%v\n\n"+
					"A seat that obeys this is refused by cobra for a flag nobody accepts, and still does not know what it needed.", label, m[1], err)
			}
		}
		return true
	}

	var seen int
	// EVERY EVENT TYPE, driven with an EMPTY body, so whatever refuses first for that type is
	// read. The first draft ran only the hand-written contract table and covered fourteen types —
	// measured, `closing requires --id` could be renamed to a flag that does not exist and this
	// gate stayed green, because no case in the table produced that message. A gate over a
	// hand-kept list of cases inherits that list's blind spots.
	for i := 0; i < vals.Len(); i++ {
		typ := recordpb.EventType(vals.Get(i).Number())
		if typ == 0 {
			continue
		}
		word := recordpb.Word(typ)
		body, ok := recordpb.BodyField(word)
		if !ok {
			continue
		}
		if inspect(word+" (empty body)", validateRun(mustRun(t, recordtest.TmpRun(t)), "red-chair", typ, emptyBodyFor(t, body.Message()))) {
			seen++
		}
	}
	// AND the contract table on top, which reaches the SECOND and later refusals — the arms an
	// empty body never gets past.
	for _, tc := range validateContractCases() {
		if tc.wantErr == "" {
			continue
		}
		if inspect(tc.name, validateRun(mustRun(t, recordtest.TmpRun(t)), "red-chair", tc.typ, tc.p)) {
			seen++
		}
	}
	// Fourteen table cases plus one empty-body drive per event type. The floor sits below the
	// real total so retiring a contract is an ordinary edit, and far enough above zero that a
	// gutted table or a broken walk cannot leave this gate reading nothing and reporting clean.
	if seen < 25 {
		t.Fatalf("only %d refusals inspected — this gate reads refusals, so a case list that produces none passes it forever", seen)
	}
}

// emptyBodyFor builds a zero-valued body of the CONCRETE generated type, so validate's type
// switch matches it. A dynamicpb message carries the same descriptor and matches no arm at all,
// which would make a gate over it pass by never reaching the code it audits.
func emptyBodyFor(t *testing.T, md protoreflect.MessageDescriptor) proto.Message {
	t.Helper()
	mt, err := protoregistry.GlobalTypes.FindMessageByName(md.FullName())
	if err != nil {
		t.Fatalf("%s is not in the global type registry: %v", md.FullName(), err)
	}
	return mt.New().Interface()
}

// A CARRY NEEDS NO FRESH ARGUMENT, AND THE EXEMPTION MUST BE REACHABLE.
//
// `chair carry` restates a closure an earlier round already argued, so demanding the argument
// again asks the same thing twice — validate says so and exempts it. `Close.prose` was then
// annotated `required: true`, which refuses UNCONDITIONALLY and runs BEFORE the switch, so the
// exemption could not execute: `chair carry --id G1 --carried-from 2`, the invocation the
// verb's own help documents with --reason listed nowhere in it, was refused outright.
//
// The code read as live the whole time — present, commented, explaining itself, and unreachable.
// That is the failure mode this test exists for, not the refusal itself.
func TestACarryIsExemptFromTheClosureArgument(t *testing.T) {
	carry := &recordpb.Close{
		GapId:        proto.String("G1"),
		CarriedFrom:  proto.String("1"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("blue/report.md"),
	}
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, "", ""); err != nil {
		t.Fatal(err)
	}
	// A real gap on the record, so the reference check passes and the ARGUMENT rule is what answers.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId: proto.String("G1"), AcceptanceCheck: proto.String("the check runs"),
		Class: recordpb.GapClass_GAP_CLASS_RISK_COVERAGE_OMISSION.Enum(), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}
	if err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_CLOSE, carry); err != nil &&
		strings.Contains(err.Error(), "requires --reason") {
		t.Errorf("a carry was refused for the argument it explicitly does not owe: %v", err)
	}
	// And an ORDINARY closure still owes one — the exemption is for the carry, not a hole.
	ordinary := &recordpb.Close{
		GapId:        proto.String("G1"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("blue/report.md"),
	}
	err := validateRun(mustRun(t, runDir), "red-chair", recordpb.EventType_EVENT_TYPE_CLOSE, ordinary)
	if err == nil || !strings.Contains(err.Error(), "requires --reason") {
		t.Errorf("a closure with no argument was accepted; the exemption widened into a hole: %v", err)
	}
}

// `principle`, `tension` AND `review_flag` ARE STATED, NEVER BLANK (gblock, 2026-10-04).
//
// A ruling always applies SOME rule, and an empty `principle` is the decoration the bench's
// disposition exists to refuse — the measured failure is a bench that ruled `carried` on 64 of 65
// items, a router rather than a judge. An empty tension or review flag cannot be told from a bench
// that skipped the question, and each has something true to say where the bench sees no conflict
// or nothing for a human: the ruling's weakest point, and what on the record already settles it.
// Each refusal goes through the write and names the flag, so a bench can tell which one it owes.
func TestTheBenchStatesItsRuleItsTensionAndItsReviewFlag(t *testing.T) {
	full := func() *recordpb.MotionRule {
		return &recordpb.MotionRule{
			MotionId: proto.String("M1"),
			Opinion:  proto.String("the repair holds at the leaf"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
				Principle:   proto.String("a claim rests on its weakest citation"),
				Tension:     proto.String("correctness vs economy"),
				ReviewFlag:  proto.String("no human review needed"),
				Settled:     proto.String("the claim as it stood may not be re-asserted"),
				Final:       proto.Bool(true),
			}},
		}
	}
	// EMPTY, not absent: the point is that presence alone satisfies none of the three. A REAL GAP
	// AND A REAL FILING on each record, so the reference checks pass and the FIELD rules answer.
	for _, tc := range []struct {
		flag string
		set  func(*recordpb.DocketRuling)
	}{
		{"--principle", func(d *recordpb.DocketRuling) { d.Principle = proto.String("") }},
		{"--tension", func(d *recordpb.DocketRuling) { d.Tension = proto.String("") }},
		{"--review-flag", func(d *recordpb.DocketRuling) { d.ReviewFlag = proto.String("") }},
	} {
		t.Run("empty "+tc.flag, func(t *testing.T) {
			run := mustRun(t, docketRunDir(t))
			o := full()
			tc.set(o.GetDocket())
			_, err := Append(sit(t, run, "judge"), o)
			if err == nil {
				t.Fatalf("a docket ruling with an EMPTY %s reached the record — an empty answer cannot be told from a skipped question", tc.flag)
			}
			if !strings.Contains(err.Error(), "requires "+tc.flag+" to say something") {
				t.Errorf("the refusal does not name %s, so a bench cannot tell which field it missed: %v", tc.flag, err)
			}
		})
	}
	t.Run("all three stated", func(t *testing.T) {
		run := mustRun(t, docketRunDir(t))
		if _, err := Append(sit(t, run, "judge"), full()); err != nil {
			t.Errorf("a docket ruling stating its rule, tension and review flag was refused: %v", err)
		}
	})
}

// A GRADE MOTION MUST ASK FOR A CHANGE.
//
// Proposing the grade already on the board contests nothing, and the ruling it produces is
// unreadable: `rejected` on a no-op and `rejected` on the merits are the same word, in the one
// exchange built to make that distinction legible. Measured 2026-08-22 — with severity already
// `high`, `--dimension severity --proposed high` filed cleanly.
func TestAGradeMotionThatMovesNothingIsRefused(t *testing.T) {
	runDir := newRun(t)
	for _, s := range []string{"red-chair", "blue-respond"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: s}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "red-chair"}, &recordpb.Mint{
		GapId: proto.String("G1"), AcceptanceCheck: proto.String("the check runs"),
		Class: recordpb.GapClass_GAP_CLASS_RISK_COVERAGE_OMISSION.Enum(), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}

	file := func(dim recordpb.GradeDimension, proposed recordpb.Grade) error {
		return validateRun(mustRun(t, runDir), "blue-respond", recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
			MotionId: proto.String("M1"),
			Basis:    proto.String("the grade overstates it"),
			Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
				GapId: proto.String("G1"), Dimension: &dim, Proposed: &proposed,
			}},
		})
	}

	// The no-op: severity is already high.
	err := file(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY, recordpb.Grade_GRADE_HIGH)
	if err == nil {
		t.Error("a motion proposing the grade already on the board was accepted — the ruling on it cannot be told from a ruling on the merits")
	} else if !strings.Contains(err.Error(), "already") {
		t.Errorf("the refusal does not say the grade is already that value, so a seat cannot see what is wrong: %v", err)
	}

	// A REAL ask on the same axis still files.
	if err := file(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY, recordpb.Grade_GRADE_LOW); err != nil {
		t.Errorf("a motion asking for an actual change was refused: %v", err)
	}
	// And the check is PER AXIS: `high` is a real move on an axis that is `medium`.
	if err := file(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT, recordpb.Grade_GRADE_HIGH); err != nil {
		t.Errorf("the no-op check is reading the wrong axis — impact is medium, so proposing high is a change: %v", err)
	}
}

// EVERY GRADE AXIS IS READABLE, or the no-op check silently stops working on the one that is not.
//
// GradeAt's zero differs from any real proposal, so an unhandled dimension would accept every
// motion on that axis while reporting nothing. The second return exists to make that loud; this
// asserts the enum and the switch have not drifted apart.
func TestEveryGradeDimensionCanBeReadFromAGap(t *testing.T) {
	g := &Gap{}
	vals := recordpb.GradeDimension(0).Descriptor().Values()
	checked := 0
	for i := 0; i < vals.Len(); i++ {
		d := recordpb.GradeDimension(vals.Get(i).Number())
		if d == 0 {
			continue
		}
		if _, ok := g.GradeAt(d); !ok {
			t.Errorf("Gap.GradeAt cannot read %q — the no-op-motion check would accept every motion on that axis and say nothing", recordpb.Word(d))
		}
		checked++
	}
	if checked < 4 {
		t.Errorf("only %d dimensions swept — the four axes are the set this check exists over", checked)
	}
}

// THE BOARD MUST PUBLISH THE ORDER IT REDUCED IN.
//
// BoardState sorted a LOCAL COPY by (TS, SeatID, Seq) and published m.Events, which is
// (Round, SeatID, Seq). So the reduction saw the corrected chronology and every consumer that
// walks Board.Events — record.Inquiries behind `show inquiries`, report assembly, the
// scorecards, the graph renderer — read events ordered by how SEAT NAMES SORT.
//
// The reproduction: `zulu` acts first in time, `alpha` acts second, and alphabetical order
// silently reverses them. An inquiry's recorded fate would then replay before the proposal it
// answers, and the projection would report the inquiry with no fate recorded.
func TestBoardPublishesEventsInTheOrderItReducedIn(t *testing.T) {
	runDir := newRun(t)
	// zulu proposes FIRST in wall-clock; alpha records its fate SECOND. Alphabetically alpha < zulu.
	writeShard(t, runDir, []*Event{
		recordtest.Stamped(recordtest.At(t, "zulu", "zulu:inquiry:#1", &recordpb.Inquiry{
			InquiryId: proto.String("Q1"),
			Status:    recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(),
			Question:  proto.String("opened"),
			Reason:    proto.String("opened"),
		}), "2026-08-22T10:00:00.000000000Z"),
		recordtest.Stamped(recordtest.At(t, "alpha", "alpha:inquiry:#1", &recordpb.Inquiry{
			InquiryId:        proto.String("Q1"),
			Status:           recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED.Enum(),
			SupersedesStatus: proto.String("proposed"),
			Reason:           proto.String("died"),
		}), "2026-08-22T11:00:00.000000000Z"),
	})

	b, err := FamilyOf(mustRun(t, runDir))
	if err != nil {
		t.Fatal(err)
	}

	// The published slice must be in TIME order, not seat-name order.
	if len(b.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(b.Events))
	}
	if b.Events[0].GetSeatId() != "zulu" || b.Events[1].GetSeatId() != "alpha" {
		t.Errorf("Board.Events = [%s, %s], want [zulu, alpha] — zulu acted an hour EARLIER, and a "+
			"board published in seat-name order hands every consumer a chronology assembled from filenames",
			b.Events[0].GetSeatId(), b.Events[1].GetSeatId())
	}

	// AND THE CONSUMER THAT READS IT AGREES. This is the half that shipped broken: the reduction
	// was already correct, so only a consumer walking Board.Events could see the defect.
	inq := InquiriesOf(b.Events, b.At)
	if len(inq) != 1 {
		t.Fatalf("expected one inquiry, got %d", len(inq))
	}
	if got := inq[0].Status; got != "abandoned" {
		t.Errorf("Q1 status = %q, want abandoned — the later event must win. %q means the projection "+
			"replayed alpha's recorded fate BEFORE zulu's proposal because alpha sorts first, which is the "+
			"defect: an inquiry reported with no fate recorded when its fate was recorded an hour later", got, got)
	}
}
