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
const noJournalNote = " — no workflow journal, so the relayed party fields (seat_id, gap_ids, head) were NOT compared"

// row is one dispatch row a fixture records: the seat, the head it pins and the gaps it engages.
type row struct {
	seat string
	pin  int64
	gaps []string
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
		at("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge", "judge-terminal", "assemble"}}),
		at("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
	}
	for _, g := range groups {
		evs = append(evs, reg("red-chair"))
		var sits []string
		seen := map[string]bool{}
		for _, r := range g {
			evs = append(evs, at("red-chair", &recordpb.Dispatch{Pin: proto.Int64(r.pin), SeatId: proto.String(r.seat), GapIds: r.gaps}))
			if !seen[r.seat] {
				seen[r.seat] = true
				sits = append(sits, r.seat)
			}
		}
		for _, s := range sits {
			evs = append(evs, reg(s))
		}
	}
	evs = append(evs, reg("red-chair"), reg("judge-terminal"), reg("assemble"))
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
		ps = append(ps, map[string]any{"seat_id": p.seat, "gap_ids": gs})
	}
	return map[string]any{
		"plan":            map[string]any{"head": head, "parties": ps, "docket": []any{}, "pass_permitted": false, "ceiling": false, "why": []any{}},
		"unruled_motions": []any{},
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
	if a.Verdict != "PASS" || !strings.Contains(a.Detail, "2 relayed plan(s) matched their dispatch rows on parties, gap_ids and head") {
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
