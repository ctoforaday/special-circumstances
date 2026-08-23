package record

import (
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

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
	m, err := MergedEvents(filepath.Join(t.TempDir(), "no-such-run"))
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
	m, err = MergedEvents(runDir)
	if err != nil {
		t.Fatalf("a records dir of non-shards = %v, want no error", err)
	}
	if len(m.Events) != 0 {
		t.Errorf("a non-shard file was parsed as a shard: %+v", m.Events)
	}
}

// THE ROUND, AND WHETHER THE NAME ANSWERS IT AT ALL.
//
// This used to assert a bare int, so `assemble` and `judge-terminal` pinned to 0 — the same value
// `red-lens-r0-L1` pins to, and 0 is SYNTHESIS, a real round with real events. The two cases were
// indistinguishable by construction, and that is #327: a bench closure at run END read as a
// closure BEFORE round 1, put a phantom entry in the archive, and made the W1.8 spot-check floor
// demand samples from rounds whose seats had done nothing wrong.
//
// The three answers are now distinct: a round the name states, round 0 stated by a seat that
// genuinely runs in synthesis, and NOT ANSWERED.
func TestRoundOf(t *testing.T) {
	cases := []struct {
		in    string
		want  int
		known bool
	}{
		// The name states it.
		{"red-lens-r3-L5", 3, true},
		{"red-merge-r12", 12, true},
		{"judge-r1", 1, true},
		{"blue-respond-r7", 7, true},
		{"red-lens-r0-L1", 0, true},
		{"first-r2-then-r9", 2, true}, // the FIRST match wins, not the last
		// A petition sitting is named for its petitioner, so it inherits that seat's round —
		// which is why the bare `judge-petition` was retired (#394).
		{"judge-petition-red-merge-r1", 1, true},
		// Round 0 BY RULE rather than by accident: these are dispatched before the round loop
		// (debate.js puts `let round = 0` after them), so synthesis is exactly where they act.
		{"frontier", 0, true},
		{"blue-synthesize", 0, true},
		{"blue-lane-2", 0, true},
		// The name cannot answer. These act AFTER the last round, and answering 0 said the
		// opposite of the truth.
		{"assemble", 0, false},
		{"judge-terminal", 0, false},
		// Not a debate seat, and not a seat id at all.
		{"operator", 0, false},
		{"", 0, false},
		{"no-round-here", 0, false},
		// The round marker is a HYPHEN-DELIMITED segment of an engine-assigned seat id, so a
		// bare "r5" does not answer — there is no seat named that, and matching it would also
		// make "frontier" round 0 by accident rather than by rule.
		{"r5", 0, false},
		{"red-lens-rX-L1", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, known := RoundOf(tc.in)
			if known != tc.known {
				t.Fatalf("RoundOf(%q) known = %v, want %v — a name that cannot answer must say so, not answer 0", tc.in, known, tc.known)
			}
			if known && got != tc.want {
				t.Errorf("RoundOf(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
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

// Ids are minted tool-side and sequentially PER ROUND: the collision class that
// produced four different "R5-1"s cannot recur.
func TestMintGapIDIsSequentialPerRound(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-merge-r1"
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seatID, Round: RoundIn(runDir)(seatID)}, ""); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 3; i++ {
		got, err := MintGapID(runDir, 1)
		if err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("R1-%d", i); got != want {
			t.Fatalf("MintGapID = %q, want %q", got, want)
		}
		// The MINTED id, not a fixed one: each pass records the gap it just reserved, which is
		// what makes the ids sequential rather than one gap minted three times.
		if _, err := Append(Identity{RunDir: runDir, SeatID: seatID, Round: RoundIn(runDir)(seatID)}, &recordpb.Mint{GapId: proto.String(got), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
			t.Fatal(err)
		}
	}
	// A new round restarts the counter; the id namespace is per-round.
	seat2 := "red-merge-r2"
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seat2, Round: RoundIn(runDir)(seat2)}, ""); err != nil {
		t.Fatal(err)
	}
	got, err := MintGapID(runDir, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got != "R2-1" {
		t.Errorf("first mint of round 2 = %q, want R2-1", got)
	}
}

// --key makes a retried mint idempotent: a seat whose message died after a
// successful mint must get the EXISTING id, not a second gap.
func TestExistingMintByKey(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-merge-r1"
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seatID, Round: RoundIn(runDir)(seatID)}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: seatID, Round: RoundIn(runDir)(seatID)}, &recordpb.Mint{GapId: proto.String("R1-1"), MintKey: proto.String("L1-F3"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}

	got, err := ExistingMintByKey(runDir, seatID, "L1-F3")
	if err != nil {
		t.Fatal(err)
	}
	if got != "R1-1" {
		t.Errorf("ExistingMintByKey = %q, want R1-1", got)
	}
	// An empty key is not a lookup: it must never match a mint that had no key.
	if got, err := ExistingMintByKey(runDir, seatID, ""); err != nil || got != "" {
		t.Errorf("empty key = (%q,%v), want empty and no error", got, err)
	}
	// A different key, and a different SEAT with the same key, are both misses:
	// the label is stable per seat, not globally.
	if got, err := ExistingMintByKey(runDir, seatID, "L9-F9"); err != nil || got != "" {
		t.Errorf("unknown key = (%q,%v), want a miss", got, err)
	}
	if got, err := ExistingMintByKey(runDir, "red-merge-r2", "L1-F3"); err != nil || got != "" {
		t.Errorf("another seat's key matched: %q — mint keys are per-seat", got)
	}
}

func TestBoardStateReplaysGapLifecycle(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-merge-r1"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("p1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, seatID, 1, seatID+":mint:R1-2", &recordpb.Mint{GapId: proto.String("R1-2"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p2"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		// A regrade moves ONLY the keys it carries.
		recordtest.At(t, seatID, 1, seatID+":regrade:R1-1", &recordpb.Regrade{GapId: proto.String("R1-1"), Severity: recordtest.P(recordpb.Grade_GRADE_CERTAIN), Basis: proto.String("new evidence")}),
		recordtest.At(t, seatID, 1, seatID+":close:R1-2", &recordpb.Close{GapId: proto.String("R1-2"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")}),
		// A REGRADE AND A CLOSE OF AN UNKNOWN GAP USED TO BE SEEDED HERE, and the assertion was
		// that the replay IGNORED them rather than failing. Both are foreign keys onto the mint
		// now, so neither row can be written — the state is unrepresentable, and the replay's arm
		// for it is a hard error rather than a skip (see missingGap). Seeding them would fail the
		// fixture, not the assertion.
	})
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.GapOrder) != 2 || b.GapOrder[0] != "R1-1" {
		t.Fatalf("GapOrder = %v, want mint order [R1-1 R1-2]", b.GapOrder)
	}
	g1 := b.Gaps["R1-1"]
	if !g1.Open {
		t.Error("R1-1 should still be open")
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
	g2 := b.Gaps["R1-2"]
	if g2.Open || !g2.HasClosed || g2.ClosedRound != 1 {
		t.Errorf("R1-2 closure not replayed: %+v", g2)
	}
	// Absence is renderable: a gap minted without --cx keeps nil, not "".
	if g2.ComplexityCost != recordpb.Grade_GRADE_UNSPECIFIED {
		t.Errorf("an unpassed grade became %v, want nil (it renders as \"undefined\")", g2.ComplexityCost)
	}
	// R9-9 was never minted, so it is not on the board — which is now true by construction rather
	// than by the replay choosing to skip it.
	if b.Gaps["R9-9"] != nil {
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
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, lens, 1, lens+":finding:F1", &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("first")}),
		recordtest.At(t, lens, 1, lens+":finding:F2", &recordpb.Finding{Label: proto.String("F2"), Text: proto.String("second")}),
	})
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Observations) != 2 {
		t.Fatalf("both findings must replay onto the board, got %d", len(b.Observations))
	}
	for _, want := range []string{"F1", "F2"} {
		found := false
		for _, o := range b.Observations {
			if o.Finding.GetLabel() == want {
				found = true
			}
		}
		if !found {
			t.Errorf("finding %s lost its label in replay — found_by credit is keyed on it", want)
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
	if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}); err != nil {
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
		{"mint without --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("x")}, "mint requires --check"},
		{"mint with an empty --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("x"), AcceptanceCheck: proto.String("")}, "mint requires --check"},
		{"mint without --class", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT)}, "mint requires --class"},
		{"mint complete", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}, ""},

		{"close without --id", recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{}, "close requires --id"},
		{"regrade without --basis", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{}, "regrade requires --reason"},
		{"regrade complete", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}, ""},

		// `--quote`, NOT `--claim`: the field is `claim` and the flag a seat types is `--quote`,
		// which is what the annotation now declares and what internal/cli/blue/retire.go registers.
		// This expectation was the field name, so it passed while the message and the parser
		// agreed with each other and disagreed with the test — a refusal naming a flag nobody can
		// pass is the class this branch has found four times, and the test asserted it.
		{"retire without --quote", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Reason: proto.String("r")}, "retire requires --quote"},
		{"retire without --reason", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c")}, "retire requires --reason"},
		{"retire complete", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c"), Reason: proto.String("r")}, ""},

		// An unknown status is UNREPRESENTABLE: AvenueStatus is an enum, so `shelved` cannot be built.
		// What remains is the ABSENT case below, which the type cannot refuse.
		{"a line of inquiry with no status at all", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Line: proto.String("l")}, "requires --as"},
		{"a line of inquiry with no id", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("l")}, "requires an id"},
		{"a deferred line of inquiry needs a reason", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_DEFERRED), Line: proto.String("l")}, "requires --reason"},
		{"a line of inquiry without --line", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED)}, "requires --line"},
		{"a declined line of inquiry needs a reason", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_DECLINED), Line: proto.String("l")}, "requires --reason"},
		{"an abandoned line of inquiry needs a reason", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED), Line: proto.String("l")}, "requires --reason"},
		{"a PURSUED line of inquiry does not need a reason", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED), Line: proto.String("l")}, ""},
		{"a declined line of inquiry with a reason", recordpb.EventType_EVENT_TYPE_AVENUE, &recordpb.Avenue{AvenueId: proto.String("Q1"), Status: recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_DECLINED), Line: proto.String("l"), Reason: proto.String("why")}, ""},

		// The message must name the flag the PARSER accepts. It named --gap-id for as
		// long as that flag existed and kept naming it after the rename, because the
		// spelling was derived from the payload key rather than stated.
		{"opinion missing every field", recordpb.EventType_EVENT_TYPE_OPINION, &recordpb.Opinion{}, "opinion requires --id"},
		// AN UNKNOWN VERB IS UNREPRESENTABLE: the type is an enum and the body is a message, so there
		// is no "no-such-verb" to pass. The arm that ignored it is gone with the string.
	}
}

func TestValidateVerbContracts(t *testing.T) {
	for _, tc := range validateContractCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(newRun(t), "red-merge-r1", tc.typ, tc.p)
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

// opinion demands all five fields, and names the one that is missing with the
// flag spelling the seat actually typed.
// The expected spellings are LITERALS. This test used to compute them with the same
// underscore-to-hyphen transform the code under test used, which made it a tautology: it
// asserted the code agreed with itself, and passed happily while the message taught
// --gap-id and --disposition, two flags the parser had stopped accepting. A test that
// reimplements its subject cannot indict it.
// opinionRunDir is a run in which R1-1 exists, so an opinion's reference resolves and the
// test can be about the missing FIELD rather than the missing gap.
func opinionRunDir(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, ""); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundIn(runDir)("red-merge-r1")}, &recordpb.Mint{AcceptanceCheck: proto.String("the check runs"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String(id), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
		t.Fatal(err)
	}
	return runDir
}

func TestValidateOpinionNamesEachMissingField(t *testing.T) {
	wantFlag := map[string]string{
		"gap_id":      "--id",
		"disposition": "--as",
		"principle":   "--principle",
		"tension":     "--tension",
		"review_flag": "--review-flag",
		"settled":     "--settled",
	}
	// EACH FIELD CLEARED IN TURN, on a body that is otherwise complete. Clearing is `nil`, which
	// is what "the seat never passed it" means — a `proto.String("")` would SATISFY the
	// requirement, because the check is presence and an empty answer is an answer.
	complete := func() *recordpb.Opinion {
		return &recordpb.Opinion{
			GapId:       proto.String("R1-1"),
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			Principle:   proto.String("x"),
			Tension:     proto.String("x"),
			ReviewFlag:  proto.String("x"),
			Rationale:   proto.String("the ruling's reasoning"),
		}
	}
	for _, c := range []struct {
		field string
		clear func(*recordpb.Opinion)
	}{
		// The gap must EXIST and be NAMED CORRECTLY: references are checked at write time, so
		// gap_id carries the real minted id rather than a placeholder. With a bogus id the
		// reference check fires first and this would assert on the wrong refusal.
		{"gap_id", func(o *recordpb.Opinion) { o.GapId = nil }},
		{"disposition", func(o *recordpb.Opinion) { o.Disposition = nil }},
		{"principle", func(o *recordpb.Opinion) { o.Principle = nil }},
		{"tension", func(o *recordpb.Opinion) { o.Tension = nil }},
		{"review_flag", func(o *recordpb.Opinion) { o.ReviewFlag = nil }},
	} {
		t.Run("missing "+c.field, func(t *testing.T) {
			o := complete()
			c.clear(o)
			err := validate(opinionRunDir(t), "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, o)
			if err == nil {
				t.Fatalf("opinion accepted without %s", c.field)
			}
			if !strings.Contains(err.Error(), wantFlag[c.field]) {
				t.Errorf("error %q does not name %q — the seat's only teacher is this string", err, wantFlag[c.field])
			}
		})
	}

	if err := validate(opinionRunDir(t), "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, complete()); err != nil {
		t.Errorf("a complete opinion was refused: %v", err)
	}

	// AN EMPTY VALUE STILL COUNTS AS PRESENT for the two fields checked by presence: `--review-flag
	// false` is a legitimate ruling, so the check is Has and not non-empty. `rationale` is the
	// exception — it is the prose the ruling turns on — and `disposition` is the exception to the
	// exception: an empty disposition rules nothing, and it is now an enum, so the empty case
	// cannot even be written.
	//
	// `principle` LEFT THIS SET on 2026-08-22 (operator's call). The comment above used to say
	// "the fields checked by presence" and mean three; it means two. See
	// TestTheOpinionDemandsARuleButNotAnInventedTension for the asymmetry and its argument —
	// stated there rather than restated here, so the two cannot drift into disagreeing about one
	// contract.
	empty := complete()
	empty.Tension = proto.String("")
	empty.ReviewFlag = proto.String("")
	if err := validate(opinionRunDir(t), "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, empty); err != nil {
		t.Errorf("opinion fields present-but-empty were refused: %v", err)
	}
}

// Lineage is never dangling: a mint that supersedes an id no mint created is
// refused, because the supersedes chain is what the whole analysis reads.
func TestValidateRefusesDanglingLineage(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-merge-r1"
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{Class: proto.String("overclaim"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("R1-1")}),
	})
	base := func(supersedes ...string) *recordpb.Mint {
		return &recordpb.Mint{
			GapId: proto.String("R1-2"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Class:     proto.String("x"), Problem: proto.String("p"),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Supersedes: supersedes,
		}
	}

	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, base("R1-1")); err != nil {
		t.Errorf("a real ancestor was refused: %v", err)
	}
	err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, base("R1-1", "R9-9"))
	if err == nil {
		t.Fatal("a dangling ancestor was accepted")
	}
	if !strings.Contains(err.Error(), "R9-9") || !strings.Contains(err.Error(), "dangling lineage refused") {
		t.Errorf("error must name the dangling id: %v", err)
	}
	// An empty lineage is not a dangling one.
	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, base()); err != nil {
		t.Errorf("an empty lineage was refused: %v", err)
	}
}

func TestValidateCloseAnchorContract(t *testing.T) {
	runDir := newRun(t)
	seatID := "red-merge-r1"
	// BOTH gaps are minted: R1-1 is the one being closed, R2-1 the successor a
	// repaired_with_regression names. A successor is a reference like any other and is
	// now checked, so a fixture that names one has to create it.
	writeShard(t, runDir, []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{Class: proto.String("overclaim"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("R1-1")}),
		recordtest.At(t, seatID, 2, seatID+":mint:R2-1", &recordpb.Mint{Class: proto.String("overclaim"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("R2-1")}),
	})
	anchored := func() *recordpb.Close {
		return &recordpb.Close{
			GapId: proto.String("R1-1"), AnchorSeat: proto.String("L1"),
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
		{"unknown gap", &recordpb.Close{GapId: proto.String("R9-9"), Prose: proto.String("verified")}, "close of unknown gap"},
		{"no anchor at all", &recordpb.Close{GapId: proto.String("R1-1"), Prose: proto.String("verified")}, "requires the verification triple"},
		{"a PARTIAL anchor is not an anchor", &recordpb.Close{GapId: proto.String("R1-1"), Prose: proto.String("verified"), AnchorSeat: proto.String("L1")}, "requires the verification triple"},
		{"anchor missing its target", &recordpb.Close{GapId: proto.String("R1-1"), Prose: proto.String("verified"), AnchorSeat: proto.String("L1"), AnchorTool: proto.String("git show")}, "requires the verification triple"},
		{"a full anchor", anchored(), ""},
		// --carried-from remains the honest alternative to re-verifying, but it is a
		// CLAIM ABOUT THE RECORD and is now checked like one. This gap has never been
		// closed, so "carried from round 2" is simply false — and accepting it was a
		// laundering path: an unanchored first closure that scores as closed, which is
		// exactly what anchored_closures_pct exists to detect. The genuine case (close
		// with an anchor, then restate) is covered in required_test.go.
		{"--carried-from cannot invent an earlier closure", &recordpb.Close{GapId: proto.String("R1-1"), Prose: proto.String("verified"), CarriedFrom: proto.String("2")}, "no closure of it exists in the record"},
		{"closed_with_regression needs a successor", withClass(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION, ""), "requires --superseded-by"},
		{"closed_with_regression with a successor", withClass(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION, "R2-1"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, tc.p)
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

func TestValidateClassRegistry(t *testing.T) {
	writeRegistry := func(t *testing.T, runDir string, body string) {
		t.Helper()
		if err := os.MkdirAll(recordsDirT(runDir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(recordsDirT(runDir), "class-registry.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	registry := `{"classes":[{"slug":"scope-creep"},{"slug":"unfalsifiable"},{"slug":"stale-source"}]}`
	// The fields every valid mint carries, so each case states only what it is ABOUT.
	mint := func(m *recordpb.Mint) *recordpb.Mint {
		m.AcceptanceCheck = proto.String("c")
		m.CheckKind = recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT)
		m.Likelihood = recordtest.P(recordpb.Grade_GRADE_MEDIUM)
		m.Impact = recordtest.P(recordpb.Grade_GRADE_MEDIUM)
		m.Problem = proto.String("p")
		return m
	}

	t.Run("no registry staged is advisory, not strict", func(t *testing.T) {
		if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("anything-at-all")})); err != nil {
			t.Errorf("advisory mode refused a class: %v", err)
		}
	})

	// AN UNPARSEABLE REGISTRY IS REFUSED. It used to degrade to advisory — accepting every class
	// slug for the whole run, silently — and the reason recorded for that was "as in the oracle".
	// The oracle is the JS engine retired in #121, so the justification outlived itself while the
	// behaviour stayed: a corrupt file turning off the gate that keeps the class vocabulary
	// honest, with nothing saying so.
	//
	// BOTH SHAPES REFUSE NOW, and they still say different things: absent means the run was never
	// configured, present-but-broken means somebody staged it and the file is damaged. The seat
	// needs to know which, because the fixes are not the same.
	t.Run("an unparseable registry is refused, not waved through", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, "{not json")
		err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("anything-at-all")}))
		if err == nil {
			t.Fatal("a corrupt registry accepted an arbitrary class — every --class passes while it stays that way, and the run reads as validated")
		}
		if !strings.Contains(err.Error(), "unreadable") {
			t.Errorf("the refusal must name the REGISTRY as the problem, or a seat re-reads its own --class: %v", err)
		}
	})

	t.Run("a known slug passes", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, registry)
		if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("scope-creep")})); err != nil {
			t.Errorf("a registry slug was refused: %v", err)
		}
	})

	t.Run("an unknown slug is refused with a hint", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, registry)
		err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("invented")}))
		if err == nil {
			t.Fatal("an unknown class was accepted")
		}
		if !strings.Contains(err.Error(), "unknown class") || !strings.Contains(err.Error(), "scope-creep") {
			t.Errorf("the refusal must offer real slugs: %v", err)
		}
	})

	// COINING IS ITS OWN EVENT, so its contract is checked against a `class-new` payload rather
	// than against a mint that happened to carry four extra fields.
	t.Run("coining requires its full triple", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, registry)
		complete := func() *recordpb.ClassNew {
			return &recordpb.ClassNew{
				Slug: proto.String("brand-new"), Definition: proto.String("x"),
				Neighbor: proto.String("scope-creep"), Distinguisher: proto.String("x"),
			}
		}
		for _, c := range []struct {
			field string
			clear func(*recordpb.ClassNew)
		}{
			{"definition", func(n *recordpb.ClassNew) { n.Definition = nil }},
			{"neighbor", func(n *recordpb.ClassNew) { n.Neighbor = nil }},
			{"distinguisher", func(n *recordpb.ClassNew) { n.Distinguisher = nil }},
		} {
			missing := c.field
			n := complete()
			c.clear(n)
			err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLASS_NEW, n)
			if err == nil {
				t.Errorf("a coining was accepted without --%s", missing)
				continue
			}
			if !strings.Contains(err.Error(), "--"+missing) {
				t.Errorf("wrong refusal for missing --%s: %v", missing, err)
			}
		}
	})

	t.Run("coining needs a REAL neighbor", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, registry)
		n := &recordpb.ClassNew{
			Slug: proto.String("brand-new"), Definition: proto.String("d"),
			Neighbor: proto.String("not-a-class"), Distinguisher: proto.String("q"),
		}
		err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLASS_NEW, n)
		if err == nil {
			t.Fatal("an invented neighbor was accepted")
		}
		if !strings.Contains(err.Error(), "not a known class") {
			t.Errorf("wrong refusal: %v", err)
		}
	})

	t.Run("a class minted earlier in the run extends the registry", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, registry)
		seatID := "red-merge-r1"
		writeShard(t, runDir, []*Event{
			recordtest.At(t, seatID, 1, seatID+":class-new:x", &recordpb.ClassNew{Slug: proto.String("run-local-class")}),
		})
		if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("run-local-class")})); err != nil {
			t.Errorf("a class minted in this run was refused: %v", err)
		}
		// And it is a valid neighbor for a further new class.
		n := &recordpb.ClassNew{
			Slug: proto.String("another"), Definition: proto.String("d"),
			Neighbor: proto.String("run-local-class"), Distinguisher: proto.String("q"),
		}
		if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLASS_NEW, n); err != nil {
			t.Errorf("a run-local class was not a valid neighbor: %v", err)
		}
	})

	t.Run("a registry with fewer than six slugs does not slice out of range", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, `{"classes":[{"slug":"only-one"}]}`)
		err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("invented")}))
		if err == nil {
			t.Fatal("expected a refusal")
		}
		if !strings.Contains(err.Error(), "only-one") {
			t.Errorf("hint did not include the single slug: %v", err)
		}
	})

	t.Run("an EMPTY registry is still strict and does not panic", func(t *testing.T) {
		runDir := newRun(t)
		writeRegistry(t, runDir, `{"classes":[]}`)
		if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_MINT, mint(&recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Class: proto.String("invented")})); err == nil {
			t.Error("an empty registry accepted an invented class")
		}
	})
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
		{"gap_id is the first label consulted", &recordpb.Close{GapId: proto.String("R1-1")}, "R1-1"},
		{"label when there is no gap_id", &recordpb.Finding{Label: proto.String("F1")}, "F1"},
		{"url is a label too", &recordpb.Cite{Url: proto.String("https://x")}, "https://x"},
		{"an empty label is not a label", &recordpb.Finding{Label: proto.String("")}, ""},
		{"no label at all", &recordpb.Friction{}, ""},
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
		recordpb.EventType_EVENT_TYPE_REVISION,
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

// THE SAME LABEL IN A LATER ROUND IS NOT A COLLISION: the key carries the seat id, and the seat id
// carries the round.
//
// Driven through the WRITE PATH rather than through deriveKey, because deriveKey now counts inside
// the writing transaction and has no pure form. That is the better test anyway: `events.key` is
// UNIQUE, so a collision is not a wrong string — it is a refused act, and this asserts that both
// findings land.
func TestTheSameLabelInALaterRoundIsNotACollision(t *testing.T) {
	runDir := t.TempDir()
	for _, seat := range []string{"red-lens-r1-L1", "red-lens-r2-L1"} {
		id := Identity{RunDir: runDir, SeatID: seat, Round: RoundIn(runDir)(seat)}
		if _, _, err := RegisterSeat(id, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := Append(id, &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("x")}); err != nil {
			t.Fatalf("%s: the same label in a later round was refused: %v", seat, err)
		}
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			findings++
		}
	}
	if findings != 2 {
		t.Errorf("%d findings landed, want 2 — the second was dedup'd away by a key that does not carry the round", findings)
	}
}

func TestJsonish(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"plain", `"plain"`},
		{"with \"quotes\"", `"with \"quotes\""`},
		{true, "true"},
		{false, "false"},
		{3, "3"},
		{nil, "<nil>"},
		{json.Number("2.5"), "2.5"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.in), func(t *testing.T) {
			if got := jsonish(tc.in); got != tc.want {
				t.Errorf("jsonish(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestMASSKeysAreExactlyTheCanonicalGrades binds MASS's grade set to flags.Grades — the single
// grade authority. Add/remove/rename a grade in flags and this fails until MASS is updated, so the
// mass mapping can never carry a grade the validator rejects (or miss one it accepts). This is the
// cross-package binding record.GRADES used to lack (it was a second, untested grade list).
func TestMASSKeysAreExactlyTheCanonicalGrades(t *testing.T) {
	got := make([]string, 0, len(MASS))
	for k := range MASS {
		got = append(got, k)
	}
	sort.Strings(got)
	want := flags.GradeNames()
	sort.Strings(want)
	if len(got) != len(want) {
		t.Fatalf("MASS has %d grades, flags.Grades has %d:\n got=%v\nwant=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("MASS grade set diverges from flags.Grades:\n got=%v\nwant=%v", got, want)
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
		"reason", "reason-file", "run", "seat-id", "json", "help", "id", "as",
		"quote", "new", "line", "hypothesis", "method", "problem", "fix", "check",
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
		md, ok := bodyDescriptor(word)
		if !ok {
			continue
		}
		if inspect(word+" (empty body)", validate(t.TempDir(), "red-merge-r1", typ, emptyBodyFor(t, md))) {
			seen++
		}
	}
	// AND the contract table on top, which reaches the SECOND and later refusals — the arms an
	// empty body never gets past.
	for _, tc := range validateContractCases() {
		if tc.wantErr == "" {
			continue
		}
		if inspect(tc.name, validate(t.TempDir(), "red-merge-r1", tc.typ, tc.p)) {
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
// `merge carry` restates a closure an earlier round already argued, so demanding the argument
// again asks the same thing twice — validate says so and exempts it. `Close.prose` was then
// annotated `required: true`, which refuses UNCONDITIONALLY and runs BEFORE the switch, so the
// exemption could not execute: `merge carry --id R2-3 --carried-from 2`, the invocation the
// verb's own help documents with --reason listed nowhere in it, was refused outright.
//
// The code read as live the whole time — present, commented, explaining itself, and unreachable.
// That is the failure mode this test exists for, not the refusal itself.
func TestACarryIsExemptFromTheClosureArgument(t *testing.T) {
	carry := &recordpb.Close{
		GapId:        proto.String("R1-1"),
		CarriedFrom:  proto.String("1"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("blue/report.md"),
	}
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}
	// A real gap on the record, so the reference check passes and the ARGUMENT rule is what answers.
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: 1}, &recordpb.Mint{
		GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("the check runs"),
		Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}
	if err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, carry); err != nil &&
		strings.Contains(err.Error(), "requires --reason") {
		t.Errorf("a carry was refused for the argument it explicitly does not owe: %v", err)
	}
	// And an ORDINARY closure still owes one — the exemption is for the carry, not a hole.
	ordinary := &recordpb.Close{
		GapId:        proto.String("R1-1"),
		AnchorSeat:   proto.String("L1"),
		AnchorTool:   proto.String("Read"),
		AnchorTarget: proto.String("blue/report.md"),
	}
	err := validate(runDir, "red-merge-r1", recordpb.EventType_EVENT_TYPE_CLOSE, ordinary)
	if err == nil || !strings.Contains(err.Error(), "requires --reason") {
		t.Errorf("a closure with no argument was accepted; the exemption widened into a hole: %v", err)
	}
}

// `principle` IS NON-EMPTY; `tension` AND `review_flag` ARE PRESENCE-ONLY.
//
// The asymmetry is the whole content of the decision (operator, 2026-08-22), so it is asserted
// as an asymmetry rather than three separate facts. A ruling always applies SOME rule, and an
// empty `principle` is the decoration `bench opinion` exists to refuse — the measured failure is
// a bench that ruled `carried` on 64 of 65 items, a router rather than a judge. Demanding the
// other two would produce invented tension and pro-forma review flags, which read as reasoning
// and are worse than an honest blank.
func TestTheOpinionDemandsARuleButNotAnInventedTension(t *testing.T) {
	full := func() *recordpb.Opinion {
		return &recordpb.Opinion{
			GapId:       proto.String("R1-1"),
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			Principle:   proto.String("a claim rests on its weakest citation"),
			Tension:     proto.String("correctness vs economy"),
			ReviewFlag:  proto.String("no human review needed"),
			Rationale:   proto.String("the repair holds at the leaf"),
		}
	}
	// A REAL GAP on the record, so the reference check passes and the FIELD rules are what answer.
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "judge-r1", Round: 1}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: 1}, &recordpb.Mint{
		GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("the check runs"),
		Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}

	// EMPTY, not absent: the point is that presence alone no longer satisfies `principle`.
	empty := full()
	empty.Principle = proto.String("")
	err := validate(runDir, "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, empty)
	if err == nil {
		t.Error("an opinion with an EMPTY principle was accepted — a ruling with no stated rule is indistinguishable from a default, which is what this verb exists to prevent")
	} else if !strings.Contains(err.Error(), "principle") {
		t.Errorf("the refusal does not name --principle, so a bench cannot tell which field it missed: %v", err)
	}

	// And the other two still take an honest blank.
	for _, tc := range []struct {
		name string
		set  func(*recordpb.Opinion)
	}{
		{"tension", func(o *recordpb.Opinion) { o.Tension = proto.String("") }},
		{"review_flag", func(o *recordpb.Opinion) { o.ReviewFlag = proto.String("") }},
	} {
		o := full()
		tc.set(o)
		if err := validate(runDir, "judge-r1", recordpb.EventType_EVENT_TYPE_OPINION, o); err != nil {
			t.Errorf("an empty %s was refused: %v\nNot every ruling has two values in conflict, and most need no human to look — demanding these produces invented tension and pro-forma flags, which read as reasoning", tc.name, err)
		}
	}
}

// A GRADE MOTION MUST ASK FOR A CHANGE.
//
// Proposing the grade already on the board contests nothing, and the ruling it produces is
// unreadable: `rejected` on a no-op and `rejected` on the merits are the same word, in the one
// exchange built to make that distinction legible. Measured 2026-08-22 — with severity already
// `high`, `--dimension severity --proposed high` filed cleanly.
func TestAGradeMotionThatMovesNothingIsRefused(t *testing.T) {
	runDir := t.TempDir()
	for _, s := range []string{"red-merge-r1", "blue-respond-r1"} {
		if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: s, Round: 1}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: 1}, &recordpb.Mint{
		GapId: proto.String("R1-1"), AcceptanceCheck: proto.String("the check runs"),
		Class: proto.String("self-attestation"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}

	file := func(dim recordpb.GradeDimension, proposed recordpb.Grade) error {
		return validate(runDir, "blue-respond-r1", recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
			MotionId: proto.String("M1"),
			Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
			Basis:    proto.String("the grade overstates it"),
			Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
				GapId: proto.String("R1-1"), Dimension: &dim, Proposed: &proposed,
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
// walks Board.Events — record.Inquiries behind `show lines-of-inquiry`, report assembly, the
// scorecards, the graph renderer — read events ordered by how SEAT NAMES SORT.
//
// FOUND 2026-08-22 BY A BLUE SEAT mid-run, with this reproduction: `zulu` acts first in time,
// `alpha` acts second, and alphabetical order silently reverses them. The seat then caught the
// defect biting its own sitting — its line-of-inquiry moves replayed before the frontier
// proposals they answered, so the projection reported every line unmoved.
func TestBoardPublishesEventsInTheOrderItReducedIn(t *testing.T) {
	runDir := newRun(t)
	// zulu proposes FIRST in wall-clock; alpha moves it SECOND. Alphabetically alpha < zulu.
	writeShard(t, runDir, []*Event{
		recordtest.Stamped(recordtest.At(t, "zulu", 1, "zulu:line-of-inquiry:#1", &recordpb.Avenue{
			AvenueId: proto.String("Q1"),
			Status:   recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(),
			Line:     proto.String("opened"),
			Reason:   proto.String("opened"),
		}), "2026-08-22T10:00:00.000000000Z"),
		recordtest.Stamped(recordtest.At(t, "alpha", 1, "alpha:line-of-inquiry:#1", &recordpb.Avenue{
			AvenueId:         proto.String("Q1"),
			Status:           recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED.Enum(),
			SupersedesStatus: proto.String("proposed"),
			Reason:           proto.String("died"),
		}), "2026-08-22T11:00:00.000000000Z"),
	})

	b, err := BoardState(runDir)
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
	inq := Inquiries(b)
	if len(inq) != 1 {
		t.Fatalf("expected one line of inquiry, got %d", len(inq))
	}
	if got := inq[0].Status; got != "abandoned" {
		t.Errorf("Q1 status = %q, want abandoned — the later event must win. %q means the projection "+
			"replayed alpha's move BEFORE zulu's proposal because alpha sorts first, which is the "+
			"defect: a line reported as never moved when it was moved an hour later", got, got)
	}
}
