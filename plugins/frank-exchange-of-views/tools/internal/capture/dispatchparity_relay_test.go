package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// noJournalNote is what the detail carries when there is no journal to compare relayed fields from.
const noJournalNote = " — no workflow journal, so the relayed party fields (seat_id, gap_ids, occasions, head) and the last plan's bench blockers were NOT compared"

// row is one dispatch row a fixture records: the seat, the head it pins and the gaps it engages.
type row struct {
	seat      string
	pin       int64
	gaps      []string
	occasions []recordpb.Occasion
}

// relayRecord seeds a record in the dispatch parity fixture's shape: each group is one chair sitting
// — the chair registers, writes the group's rows with nobody registering between, and every party
// named sits once — and returns the run directory.
func relayRecord(t *testing.T, groups ...[]row) record.Run {
	t.Helper()
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	reg := func(seat string) *record.Event { return at(seat, &recordpb.Register{}) }
	evs := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
	}
	for _, g := range groups {
		evs = append(evs, reg("red-chair"))
		var sits []string
		seen := map[string]bool{}
		for _, r := range g {
			evs = append(evs, at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(r.pin), SeatId: proto.String(r.seat), GapIds: r.gaps, Occasions: r.occasions}))
			if !seen[r.seat] {
				seen[r.seat] = true
				sits = append(sits, r.seat)
			}
		}
		for _, s := range sits {
			evs = append(evs, reg(s))
		}
	}
	evs = append(evs, reg("red-chair"), reg("judge"), reg("judge"))
	dir := t.TempDir()
	recordtest.Seed(t, dir, evs...)
	return runtest.Open(t, dir)
}

// chairResult is a chair envelope's result as the workflow journals it: the relayed plan beside the
// envelope's other fields.
func chairResult(head int64, parties ...row) map[string]any {
	ps := []any{}
	for _, p := range parties {
		gs := []any{}
		for _, g := range p.gaps {
			gs = append(gs, g)
		}
		party := map[string]any{"seat_id": p.seat, "gap_ids": gs}
		if len(p.occasions) > 0 {
			os := []any{}
			for _, o := range p.occasions {
				os = append(os, recordpb.Word(o))
			}
			party["occasions"] = os
		}
		ps = append(ps, party)
	}
	return map[string]any{
		"plan": map[string]any{"head": head, "parties": ps, "docket": []any{}, "pass_permitted": false, "ceiling": false, "why": []any{}, "blockers": []any{}},
	}
}

// journalOf writes results as journal.jsonl lines and reads them back through ReadJournal, so the
// audit sees exactly what capture hands it — numbers as json.Number included.
func journalOf(t *testing.T, results ...map[string]any) []map[string]any {
	t.Helper()
	dir := t.TempDir()
	var b strings.Builder
	for i, r := range results {
		line, err := json.Marshal(map[string]any{"agentId": fmt.Sprintf("a%d", i), "result": r})
		if err != nil {
			t.Fatal(err)
		}
		b.Write(line)
		b.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "journal.jsonl"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := ReadJournal(dir)
	return got
}

// The fixture's two chair sittings: the lenses on nothing, then blue on G1, every row at head 2.
var (
	lensesRows = []row{{seat: "red-lens-evidence", pin: 2}, {seat: "red-lens-logic", pin: 2}}
	blueRows   = []row{{seat: "blue-respond", pin: 2, gaps: []string{"G1"}}}
)

func TestRelayedGapIDsDisagreeingWithDispatchRowFail(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	js := journalOf(t, chairResult(2, lensesRows...), chairResult(2, row{seat: "blue-respond", gaps: []string{"G2"}}))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 2: blue-respond relayed gap_ids [G2], its dispatch row recorded [G1]") {
		t.Fatalf("a relayed gap id the dispatch row does not hold = %s: %s", a.Verdict, a.Detail)
	}
}

func TestRelayedHeadDisagreeingWithPinFails(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	js := journalOf(t, chairResult(2, lensesRows...), chairResult(5, blueRows...))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 2: blue-respond relayed head 5, its dispatch row pinned 2") {
		t.Fatalf("a relayed head the dispatch row did not pin = %s: %s", a.Verdict, a.Detail)
	}
}

func TestRelayedPartySetDisagreeingWithDispatchFails(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	js := journalOf(t, chairResult(2, lensesRows[0]), chairResult(2, blueRows...))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 1: relayed parties [red-lens-evidence], the dispatch rows name [red-lens-evidence, red-lens-logic]") {
		t.Fatalf("a party dropped from the relay = %s: %s", a.Verdict, a.Detail)
	}
}

func TestRelayCountMismatchFails(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	js := journalOf(t, chairResult(2, lensesRows...))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "the journal relays 1 plan(s) with parties and the record holds 2 chair sitting dispatch(es)") {
		t.Fatalf("one relay for two dispatches = %s: %s", a.Verdict, a.Detail)
	}
}

// THE RESUME SHAPE: a resumed workflow re-journals a cached chair result. Every relay is faithful,
// and the pairing is off by one — the audit says so by the count rather than comparing sitting 2's
// relay against sitting 1's rows.
func TestRelayPairingOnDuplicatedChairResult(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	first := chairResult(2, lensesRows...)
	js := journalOf(t, first, first, chairResult(2, blueRows...))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "the journal relays 3 plan(s) with parties and the record holds 2 chair sitting dispatch(es)") {
		t.Fatalf("a chair result journaled twice = %s: %s", a.Verdict, a.Detail)
	}
}

// A docket plan is never standing, so a chair asking twice in one sitting writes two rows for one
// party. The later row is the dispatch, and the relay is held to it.
func TestRelayComparedAgainstLastRowOfDuplicatedGroup(t *testing.T) {
	twice := []row{{seat: "blue-respond", pin: 2, gaps: []string{"G1"}}, {seat: "blue-respond", pin: 2, gaps: []string{"G1", "G3"}}}
	run := relayRecord(t, lensesRows, twice)

	last := journalOf(t, chairResult(2, lensesRows...), chairResult(2, row{seat: "blue-respond", gaps: []string{"G3", "G1"}}))
	if a := DispatchParityAudit(run, last, true); a.Verdict != "PASS" {
		t.Fatalf("a relay matching the group's last row = %s: %s", a.Verdict, a.Detail)
	}
	first := journalOf(t, chairResult(2, lensesRows...), chairResult(2, twice[0]))
	if a := DispatchParityAudit(run, first, true); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 2: blue-respond relayed gap_ids [G1], its dispatch row recorded [G1, G3]") {
		t.Fatalf("a relay matching only the group's first row = %s: %s", a.Verdict, a.Detail)
	}
}

// A faithful relay passes, and the journal's other results — a lens envelope, the closing chair
// sitting's empty plan — pair with nothing.
func TestFaithfulRelayFieldsPass(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	lens := map[string]any{"findings": []any{}, "log": []any{}}
	js := journalOf(t, chairResult(2, lensesRows...), lens, chairResult(2, blueRows...), chairResult(2))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "PASS" || !strings.Contains(a.Detail, "2 relayed plan(s) matched their dispatch rows on parties, gap_ids, occasions and head") {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}
}

func TestNoJournalStatesFieldsNotCompared(t *testing.T) {
	run := relayRecord(t, lensesRows, blueRows)
	a := DispatchParityAudit(run, nil, false)
	if a.Verdict != "PASS" || !strings.HasSuffix(a.Detail, noJournalNote) || strings.Contains(a.Detail, "matched") {
		t.Fatalf("no journal must run the register half and say the relayed fields were not compared = %s: %s", a.Verdict, a.Detail)
	}
}

// THE OCCASIONS ARE RELAYED FIELDS WITH A ROW COUNTERPART. The workflow routes the bench on them, so
// a relay that drops or alters one is a departure like a wrong gap id — and a faithful one is not.
func TestRelayedOccasionsDisagreeingWithDispatchRowFail(t *testing.T) {
	bench := []row{{seat: "judge", pin: 2, gaps: []string{"G1"}, occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}}}
	run := relayRecord(t, lensesRows, bench)
	faithful := journalOf(t, chairResult(2, lensesRows...), chairResult(2, bench...))
	if a := DispatchParityAudit(run, faithful, true); strings.Contains(a.Detail, "relayed occasions") {
		t.Fatalf("a faithful occasion relay was reported: %s", a.Detail)
	}
	altered := bench[0]
	altered.occasions = []recordpb.Occasion{recordpb.Occasion_OCCASION_PETITION, recordpb.Occasion_OCCASION_DOCKET}
	js := journalOf(t, chairResult(2, lensesRows...), chairResult(2, altered))
	a := DispatchParityAudit(run, js, true)
	if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 2: judge relayed occasions [docket, petition], its dispatch row recorded [docket]") {
		t.Fatalf("an occasion added in the relay = %s: %s", a.Verdict, a.Detail)
	}
}

// THE LAST PLAN'S BENCH BLOCKERS ARE HELD TO THE RECORD AT THE END OF THE CHAIR'S LAST SITTING. No row
// records a blocker, and the workflow convenes the terminal bench off the relayed set — so a chair
// that asked for its plan, then filed a petition, and relayed the plan it asked for first has
// relayed a stale one, and a chair that drops a blocker in the relay has altered it. A plan asked
// for after the filing agrees with the record.
func TestTheLastPlansBenchBlockersAreHeldToTheRecord(t *testing.T) {
	n := 0
	at := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	evs := []*record.Event{
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		at("red-chair", &recordpb.Register{}),
		at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence")}),
		at("red-lens-evidence", &recordpb.Register{}),
		// The chair's last sitting: its plan is empty, and it files a petition after asking.
		at("red-chair", &recordpb.Register{}),
		at("red-chair", &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
			Basis: proto.String("b"), Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}}}),
		// The terminal sitting, which the stale plan would not have convened.
		at("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_TERMINAL.Enum()}),
	}
	dir := t.TempDir()
	recordtest.Seed(t, dir, evs...)
	run := runtest.Open(t, dir)
	final := func(blockers ...any) map[string]any {
		r := chairResult(2)
		r["plan"].(map[string]any)["blockers"] = blockers
		return r
	}
	lensRows := []row{{seat: "red-lens-evidence", pin: 2}}
	petition := map[string]any{"kind": "unruled_motion", "subject": "M1", "owner": "judge"}
	for _, c := range []struct {
		name     string
		last     map[string]any
		departed bool
	}{
		{"asked before the filing", final(), true},
		{"a blocker altered in the relay", final(map[string]any{"kind": "unruled_motion", "subject": "M9", "owner": "judge"}), true},
		{"asked after the filing", final(petition), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := DispatchParityAudit(run, journalOf(t, chairResult(2, lensRows...), c.last), true)
			got := strings.Contains(a.Detail, "and the record at the end of the chair's last sitting holds [M1] unruled")
			if got != c.departed {
				t.Fatalf("departure reported = %v, want %v: %s: %s", got, c.departed, a.Verdict, a.Detail)
			}
			if !c.departed && !strings.Contains(a.Detail, "the last plan's bench blockers matched the record") {
				t.Fatalf("an agreeing relay does not say it was compared: %s", a.Detail)
			}
		})
	}
}
