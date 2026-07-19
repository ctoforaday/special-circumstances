package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE BOARD AS STRUCTURED STATE.
//
// A seat ACTS on the board: which gap is open, what its grades are, whether a closure
// carried its anchor. Every one of those questions used to be answered by parsing prose
// the seat had been told to read from a file path, and that parsing is where the scorecard
// defects came from — anchored_closures_pct read 0 against an 89 baseline because the
// metric parsed hand-written sentences while the anchors sat in structured fields one
// channel over.

type boardJSON struct {
	Open         []map[string]any `json:"open"`
	Closed       []map[string]any `json:"closed"`
	Observations []struct {
		ID       string         `json:"id"`
		Label    string         `json:"label"`
		Disposed bool           `json:"disposed"`
		Fate     map[string]any `json:"fate"`
	} `json:"observations"`
	Counts struct {
		Open             int `json:"open"`
		Closed           int `json:"closed"`
		ClosedByBench    int `json:"closed_by_bench"`
		UndisposedObserv int `json:"undisposed_observations"`
		Anomalies        int `json:"anomalies"`
	} `json:"counts"`
	Anomalies []string `json:"anomalies"`
}

func board(t *testing.T, runDir, seatRole, seatID string) boardJSON {
	t.Helper()
	out, err := run(t, seatRole, "show", "--run", runDir, "--seat-id", seatID, "--view", "board")
	if err != nil {
		t.Fatalf("show --view board: %v", err)
	}
	var b boardJSON
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		t.Fatalf("the board must be valid JSON — a seat parses this, and a half-formed document is worse than none: %v\n%s", err, out)
	}
	return b
}

func ids(rows []map[string]any) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		if s, ok := r["id"].(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// THE INVARIANT. The JSON board and the markdown ledger are two renderings of ONE replay,
// so they cannot disagree about which gaps are open. If they ever do, one of them is
// parsing the other — the second-reader defect this tool exists to remove, reintroduced.
func TestBoardJSONAndMarkdownLedgerAgreeOnWhatIsOpen(t *testing.T) {
	runDir := seatRun(t)
	open1 := mintGap(t, runDir, "stays-open", "json-vs-markdown")
	closedByRed := mintGap(t, runDir, "red-closes", "json-vs-markdown")
	closedByBench := mintGap(t, runDir, "bench-closes", "json-vs-markdown")

	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", closedByRed, "--as", "closed",
		"--anchor-seat", "L1", "--anchor-tool", "go test", "--anchor-target", "./internal/x",
		"--text", "the check passes"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := run(t, "bench", "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", closedByBench, "--as", "closed", "--principle", "p", "--tension", "t",
		"--review-flag", "no"); err != nil {
		t.Fatalf("bench opinion: %v", err)
	}

	b := board(t, runDir, "merge", "red-merge-r1")
	if got := ids(b.Open); len(got) != 1 || got[0] != open1 {
		t.Errorf("JSON board open = %v, want [%s]", got, open1)
	}
	if b.Counts.ClosedByBench != 1 {
		t.Errorf("closed_by_bench = %d, want 1 — WHO closed a gap decides whether red may close it again, and double-counting corrupts the repair_regression denominator", b.Counts.ClosedByBench)
	}

	// And the markdown must say the same thing about every one of them.
	for _, c := range []struct {
		id       string
		wantOpen bool
	}{{open1, true}, {closedByRed, false}, {closedByBench, false}} {
		if got := gapIsOpen(t, runDir, c.id); got != c.wantOpen {
			t.Errorf("markdown says gap %s open=%v, JSON says open=%v — two renderings of one replay have diverged", c.id, got, c.wantOpen)
		}
	}
}

// A closure's ANCHOR must arrive as fields, not as a sentence. The markdown flattened the
// triple into prose and the scorecard could not parse it back out, which is how a run with
// anchored closures scored 0% anchored.
func TestBoardJSONCarriesTheClosureAnchorAsFields(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "anchored", "anchor-as-fields")
	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "closed",
		"--anchor-seat", "L4", "--anchor-tool", "git show", "--anchor-target", "7bc501e:report.md",
		"--text", "re-read the cited source"); err != nil {
		t.Fatalf("close: %v", err)
	}

	b := board(t, runDir, "merge", "red-merge-r1")
	if len(b.Closed) != 1 {
		t.Fatalf("closed = %v, want one gap", ids(b.Closed))
	}
	closure, ok := b.Closed[0]["closure"].(map[string]any)
	if !ok {
		t.Fatalf("the closed gap carries no structured closure: %v", b.Closed[0])
	}
	for k, want := range map[string]string{
		"anchor_seat": "L4", "anchor_tool": "git show", "anchor_target": "7bc501e:report.md",
	} {
		if got, _ := closure[k].(string); got != want {
			t.Errorf("closure.%s = %q, want %q — an anchor a machine cannot read is why anchored_closures_pct scored 0 on a run whose closures WERE anchored", k, got, want)
		}
	}
}

// An observation with no fate is the merge seat's outstanding work, and `disposed` states
// it rather than making the consumer test a null. Falsy-vs-absent confusion has produced
// three separate defects in this codebase already.
func TestBoardJSONStatesWhetherAnObservationHasAFate(t *testing.T) {
	runDir := seatRun(t)
	for _, l := range []string{"L1-O1", "L1-O2"} {
		if _, err := run(t, "lens", "observe", "--run", runDir, "--seat-id", "red-lens-r1-L1",
			"--label", l, "--kind", "note", "--text", "a thing worth noticing"); err != nil {
			t.Fatalf("observe %s: %v", l, err)
		}
	}
	if _, err := run(t, "merge", "dispose", "--run", runDir, "--seat-id", "red-merge-r1",
		"--observation", "L1-O1", "--as", "declined",
		"--reason", "checked at the leaf; correct as written"); err != nil {
		t.Fatalf("dispose: %v", err)
	}

	b := board(t, runDir, "merge", "red-merge-r1")
	if b.Counts.UndisposedObserv != 1 {
		t.Errorf("undisposed_observations = %d, want 1 — this count IS the merge seat's outstanding duty", b.Counts.UndisposedObserv)
	}
	for _, o := range b.Observations {
		switch o.Label {
		case "L1-O1":
			if !o.Disposed || o.Fate == nil {
				t.Errorf("L1-O1 was disposed but the board reports disposed=%v fate=%v", o.Disposed, o.Fate)
			}
		case "L1-O2":
			if o.Disposed {
				t.Error("L1-O2 has no disposal and must not report one")
			}
		}
	}
}

// ANOMALIES REACH THE SEAT. A dropped mutation used to vanish; the 2026-07-18 run spent
// three rounds on a board that was wrong by six gaps with nothing on the surface to say
// so. A seat that can see an anomaly can petition about it.
func TestBoardJSONSurfacesDroppedMutations(t *testing.T) {
	runDir := seatRun(t)

	// WRITTEN STRAIGHT TO A SHARD, because the CLI now refuses to create one.
	//
	// This test first drove the dangling opinion through the CLI and asserted the write
	// path let it through — "catching the dangling reference is the REPLAY's job". That
	// was wrong, and it was wrong in the specific way that cost the 2026-07-18 run six
	// gaps: an event accepted at write and dropped at replay looks recorded and does
	// nothing, and the seat is long gone by the time anyone notices.
	//
	// The write path refuses it now. But shards like this EXIST — the run's own records
	// carry twelve — so replay must still surface them rather than skip in silence, and
	// that is what this asserts.
	// APPENDED TO THE SEAT'S OWN SHARD, not a second one.
	//
	// The first version wrote events-judge-r1-deadbeef.jsonl beside the shard seatRun had
	// already registered, which made judge-r1 multi-nonce and left the winner to mtime.
	// It passed here and failed in CI, where the registered shard won and the dangling
	// opinion never replayed at all — a test whose outcome depended on filesystem
	// timestamp granularity, which is the exact defect this morning's golden fix removed.
	// One shard, one nonce, no race.
	nonce, err := os.ReadFile(filepath.Join(runDir, "records", ".active-judge-r1"))
	if err != nil {
		t.Fatal(err)
	}
	shard := filepath.Join(runDir, "records",
		"events-judge-r1-"+strings.TrimSpace(string(nonce))+".jsonl")
	line := `{"seq":1,"ts":"2026-07-19T12:00:00.000000000Z","seatId":"judge-r1","nonce":"` +
		strings.TrimSpace(string(nonce)) + `","round":1,` +
		`"type":"opinion","key":"judge-r1:opinion:R9-9","payload":{"gap_id":"R9-9","disposition":"closed",` +
		`"principle":"p","tension":"t","review_flag":"no"}}` + "\n"
	f, err := os.OpenFile(shard, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b := board(t, runDir, "merge", "red-merge-r1")
	if b.Counts.Anomalies == 0 {
		t.Fatal("a ruling on an unknown gap produced NO anomaly — this is the silent drop that let a board be wrong by six gaps for three rounds")
	}
	var found bool
	for _, a := range b.Anomalies {
		if strings.Contains(a, "R9-9") && strings.Contains(a, "DROPPED") {
			found = true
		}
	}
	if !found {
		t.Errorf("the anomaly must name the gap AND say the mutation was dropped, or a seat cannot tell what its board is missing: %v", b.Anomalies)
	}
}
