package cli

import (
	"encoding/json"
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

func board(t *testing.T, runDir, seatRole, seatID string) boardJSON {
	t.Helper()
	out, err := run(t, seatRole, "show", "--run", runDir, "--seat-id", seatID, "board")
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

	if _, err := run(t, "merge", "close", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", closedByRed, "--as", "closed",
		"--verified-by", "L1", "--verified-with", "go test", "--verified-against", "./internal/x",
		"--reason", "the check passes"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := run(t, "bench", "opinion", "--run", runDir, "--seat-id", "judge-r1",
		"--id", closedByBench, "--as", "closed", "--principle", "p", "--tension", "t",
		"--review-flag", "no", "--reason", "closed on the merits"); err != nil {
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
		"--verified-by", "L4", "--verified-with", "git show", "--verified-against", "7bc501e:report.md",
		"--reason", "re-read the cited source"); err != nil {
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

// The findings view is the merge's structured read of the lens findings, replacing the
// red/candidates/*.md files — label (tool-assigned), role from the seat id, grades, text.
func TestFindingsViewProjectsLensFindings(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--key", "F1", "--quote", "§1", "--reason", "first", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "lens", "register", "--run", runDir, "--seat-id", "red-lens-r1-L5"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "lens", "finding", "--run", runDir, "--seat-id", "red-lens-r1-L5",
		"--key", "F1", "--quote", "§2", "--reason", "second", "--severity", "high", "--likelihood", "high", "--impact", "high"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "merge", "show", "--run", runDir, "--seat-id", "red-merge-r1", "findings")
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

// citations_checked is the record's, not red's self-report. The board tallies cite events so red
// reads the count from its native view instead of hand-counting a number that was fabricated on
// haiku. The count is DISTINCT sources verified.
//
// # A CAPABILITY QUESTION THIS TEST NOW PINS, and it is a real one
//
// A verify keys on its reference (`url` is in keyFields), so two verifications of one source in
// one sitting share a key. Under shards both were written and the READER kept one — the header
// here called that "idempotent (updates in place)", which was a read-time illusion over an
// append-only log. `events.key` is UNIQUE now, so the second is REFUSED.
//
// That is a loss and it is stated rather than absorbed: a lens that re-reads a source mid-sitting
// and finds something different cannot record the second reading. The refusal follows from two
// deliberate choices (append-only, one act per key) and the alternative — dropping `url` from
// keyFields so verifies key on an ordinal — changes what a crash-retry does. Which way that should
// go is the operator's call; it is tracked in plans/record-sqlite.md rather than decided here.
func TestBoardCountsCiteEvents(t *testing.T) {
	runDir := seatRun(t)
	cites := []struct{ claim, ref string }{
		{"the API returns 200", "https://example.com/a"},
		{"the flag defaults off", "https://example.com/b"},
	}
	for _, c := range cites {
		// --independent: these are sources red went and found, not citations blue authored, so
		// there is no anchor to name. The explicit form, because an omitted --anchor cannot say
		// whether this was corroboration or a lookup the seat skipped.
		if _, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
			"--quote", c.claim, "--url", c.ref, "--title", c.ref,
			"--as", "supports", "--confidence", "high", "--reason", "read at the leaf",
			"--access-date", "2026-07-24"); err != nil {
			t.Fatalf("cite %q: %v", c.claim, err)
		}
	}
	// The same reference again in the SAME sitting is refused, with a message that says so.
	_, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
		"--quote", "the API still returns 200", "--url", "https://example.com/a", "--title", "a",
		"--as", "supports", "--confidence", "high", "--reason", "read again",
		"--access-date", "2026-07-24")
	if err == nil {
		t.Error("a second verification of one source in one sitting was accepted — two acts under one key")
	} else if !strings.Contains(err.Error(), "once-per-sitting") {
		t.Errorf("the refusal does not teach what was wrong:\n%v", err)
	}

	if b := board(t, runDir, "merge", "red-merge-r1"); b.Counts.Citations != 2 {
		t.Errorf("counts.citations = %d, want 2 (two distinct references) — the board is the source for citations_checked", b.Counts.Citations)
	}
}

// A DANGLING REFERENCE IS REFUSED AT THE WRITE, and there is no longer a dropped mutation for the
// seat to be told about.
//
// This test wrote a dangling opinion STRAIGHT INTO A SHARD — the CLI already refused one — to
// prove that replay surfaced it rather than skipping in silence. That mattered because a run's
// records carried twelve such events: accepted at write by an older binary, dropped at replay,
// looking recorded and doing nothing. The 2026-07-18 run spent three rounds on a board that was
// wrong by six gaps with nothing on the surface to say so.
//
// `opinion.gap_id` is a FOREIGN KEY onto `mint.gap_id` now. The row cannot be written — not by the
// CLI, not by a fixture, not by anything writing SQL at the file — so there is no dangling
// reference to surface, no anomaly channel to carry it, and the replay's arm for it is a hard
// error rather than a note (see missingGap). The state moved from "detected after the fact" to
// "unrepresentable", which is what the whole storage change was for.
//
// What survives is the write-path refusal, which has its own test: the CLI answers a `--id` naming
// no mint with a message that says so. This one is deleted rather than pinned to a permanent zero,
// because a green anomaly check for an anomaly that cannot occur reads as evidence and is not.

