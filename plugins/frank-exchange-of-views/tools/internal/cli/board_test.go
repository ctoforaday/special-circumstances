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
		Citations        int `json:"citations"`
	} `json:"counts"`
	Anomalies []string `json:"anomalies"`
}

// THE ROLE LEFT THE ARGS: the seat id selects the tree.
func board(t *testing.T, runDir, seatID string) boardJSON {
	t.Helper()
	out, err := run(t, "show", "--run", runDir, "--seat-id", seatID, "board")
	if err != nil {
		t.Fatalf("show board: %v", err)
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

	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", closedByRed, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := run(t, "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", closedByBench, "--as", "repaired", "--principle", "p", "--tension", "t",
		"--review-flag", "no", "--reason", "closed on the merits"); err != nil {
		t.Fatalf("bench opinion: %v", err)
	}

	b := board(t, runDir, "red-merge-r1")
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
	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", id, "--as", "repaired",
		"--verified-by", "L4", "--verified-with", "git show", "--verified-against", "7bc501e:report.md",
		"--reason", "re-read the cited source"); err != nil {
		t.Fatalf("close: %v", err)
	}

	b := board(t, runDir, "red-merge-r1")
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

// The findings view is the merge's structured read of the lens findings, replacing the
// red/candidates/*.md files — label (tool-assigned), role from the seat id, grades, text.
func TestFindingsViewProjectsLensFindings(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--quote", "§1", "--reason", "first", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-r1-L5"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L5",
		"--key", "F1", "--quote", "§2", "--reason", "second", "--severity", "high", "--likelihood", "high", "--impact", "high"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", "findings")
	if err != nil {
		t.Fatalf("show findings: %v", err)
	}
	var fv struct {
		Findings []struct{ Label, Role, Text string }
		Counts   struct{ Total int }
	}
	if err := json.Unmarshal([]byte(out), &fv); err != nil {
		t.Fatalf("findings view must be valid JSON the merge parses: %v\n%s", err, out)
	}
	if fv.Counts.Total != 2 || len(fv.Findings) != 2 {
		t.Fatalf("findings total = %d, want 2", fv.Counts.Total)
	}
	got := map[string]string{} // label -> role
	for _, f := range fv.Findings {
		got[f.Label] = f.Role
	}
	if got["L1-F1"] != "L1" || got["L5-F1"] != "L5" {
		t.Errorf("findings view mislabels/misroutes: %v — the tool assigns L{role}-F{N} and the role comes from the seat id", got)
	}
}

// citations_checked is the record's, not red's self-report. The board tallies cite
// events so red reads the count from its native view instead of hand-counting a number
// that was fabricated on haiku. Cite events are reference-keyed, so re-verifying a
// source is idempotent (updates in place): the count is DISTINCT sources verified —
// three cite calls over two references tally two.
func TestBoardCountsCiteEvents(t *testing.T) {
	runDir := seatRun(t)
	cites := []struct{ claim, ref string }{
		{"the API returns 200", "https://example.com/a"},
		{"the flag defaults off", "https://example.com/b"},
		{"the API still returns 200 (re-verified next round)", "https://example.com/a"}, // same ref → idempotent
	}
	for _, c := range cites {
		// --independent: these are sources red went and found, not citations blue authored, so
		// there is no anchor to name. The explicit form, because an omitted --anchor cannot say
		// whether this was corroboration or a lookup the seat skipped.
		if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
			"--quote", c.claim, "--url", c.ref, "--title", c.ref,
			"--as", "supports", "--confidence", "high", "--reason", "read at the leaf",
			"--access-date", "2026-07-24"); err != nil {
			t.Fatalf("cite %q: %v", c.claim, err)
		}
	}
	if b := board(t, runDir, "red-merge-r1"); b.Counts.Citations != 2 {
		t.Errorf("counts.citations = %d, want 2 (two distinct references) — the board is the source for citations_checked", b.Counts.Citations)
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
		`"type":"opinion","key":"judge-r1:opinion:R9-9","payload":{"gap_id":"R9-9","disposition":"repaired",` +
		`"principle":"p","tension":"t","review_flag":"no"}}` + "\n"
	f, err := os.OpenFile(shard, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(line); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b := board(t, runDir, "red-merge-r1")
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

// THE ESTOPPEL REGISTER MUST SAY WHO BARRED WHAT.
//
// `show work` is the projection every seat is told to run first and again before it stops, so
// its closed_index is the one carrier that reaches every board. It carried {id, location, class}
// — enough to say a gap is GONE, and not enough to say it is BARRED.
//
// Measured on the 2026-08-22 sqlite-schema run: a bench defect_owed_elsewhere ruling (still
// broken, not blue's to fix), a clean red closure, and a repaired_with_regression whose successor
// was still live all rendered as three identical three-field objects. A seat could not tell a
// ruling it must not relitigate from one of its own closures it may reopen on new evidence —
// and could not tell either from a gap nobody had ever raised, since all three are simply absent
// from the open set. The absent case and the healthy case, the same bytes, again.
func TestClosedIndexSaysWhoClosedItAndHow(t *testing.T) {
	runDir := seatRun(t)
	byRed := mintGap(t, runDir, "red-closes-this", "cross-seat-visibility")
	byBench := mintGap(t, runDir, "bench-rules-this", "cross-seat-visibility")

	if _, err := run(t, "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", byRed, "--as", "repaired",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := run(t, "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", byBench, "--as", "defect_owed_elsewhere", "--principle", "capability",
		"--tension", "correctness", "--review-flag", "yes",
		"--reason", "no verb can perform this fix at any cost"); err != nil {
		t.Fatalf("bench opinion: %v", err)
	}

	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r2", "work")
	if err != nil {
		t.Fatalf("show work: %v", err)
	}
	var w struct {
		ClosedIndex []struct {
			ID            string `json:"id"`
			Class         string `json:"class"`
			Fate          string `json:"fate"`
			ClosedBy      string `json:"closed_by"`
			ArtifactState string `json:"artifact_state"`
		} `json:"closed_index"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatalf("work must be valid JSON — a seat parses this: %v\n%s", err, out)
	}

	got := map[string][2]string{}
	for _, c := range w.ClosedIndex {
		got[c.ID] = [2]string{c.ClosedBy, c.Fate}
	}
	if len(got) != 2 {
		t.Fatalf("closed_index = %+v, want both closed gaps", w.ClosedIndex)
	}
	if g := got[byRed]; g[0] != "red" || g[1] != "repaired" {
		t.Errorf("red's own closure reads as {closed_by:%q fate:%q}, want {red repaired} — red may reopen "+
			"its own closure on new evidence, and cannot know that from an entry that will not say who closed it", g[0], g[1])
	}
	if g := got[byBench]; g[0] != "bench" || g[1] != "defect_owed_elsewhere" {
		t.Errorf("the bench ruling reads as {closed_by:%q fate:%q}, want {bench defect_owed_elsewhere} — "+
			"a bench ruling is ESTOPPED and re-raising it is relitigation, which a seat cannot avoid doing "+
			"if the register will not distinguish it from red's own act", g[0], g[1])
	}
	// AND THE TWO MUST NOT COLLIDE. The whole defect was that they rendered identically.
	if got[byRed] == got[byBench] {
		t.Errorf("a red closure and a bench ruling render identically as %v — this is the defect, restored", got[byRed])
	}

	// THE SECOND AXIS. The docket closed on both, and the ARTIFACT did not: red verified a
	// repair, while the bench ruled a real defect owed elsewhere. A board that cannot say that
	// reports open:0 over a report still carrying the defect — measured, on this run.
	art := map[string]string{}
	for _, c := range w.ClosedIndex {
		art[c.ID] = c.ArtifactState
	}
	if art[byRed] != "repaired" {
		t.Errorf("a verified repair reads artifact_state %q, want repaired", art[byRed])
	}
	if art[byBench] != "defect_live" {
		t.Errorf("defect_owed_elsewhere reads artifact_state %q, want defect_live — the dispute "+
			"is over and the defect is not, which is the whole point of the axis", art[byBench])
	}
}
