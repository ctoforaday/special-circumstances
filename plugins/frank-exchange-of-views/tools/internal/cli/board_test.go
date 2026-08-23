package cli

import (
	"encoding/json"
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
		"--review-flag", "no", "--settled", "the proposition this ruling bars", "--final", "--reason", "closed on the merits"); err != nil {
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

// citations_checked is the record's, not red's self-report. The board tallies cite events so red
// reads the count from its native view instead of hand-counting a number that was fabricated on
// haiku. The count is DISTINCT sources verified.
//
// # A CAPABILITY QUESTION THIS TEST NOW PINS, and it is a real one
//
// A verify keys on its reference (`url` is in keyFields), so two verifications of one source in
// ONE SOURCE MAY CORROBORATE MANY CLAIMS, and this test used to assert the opposite.
//
// It pinned a REFUSAL: a corroboration keyed on its `url`, so a second one naming the same source
// in the same sitting collided and was refused. That was recorded as a real loss with three ways
// out, and the operator chose one (2026-08-22): red's supporting corroborations mint a label and
// become ordinary footnotes, exactly as blue's cites do. `keyFields` is walked first-match and
// `label` sits before `url`, so the key moved off the source and the collision is gone.
//
// The loss it described was never exotic: a strong source usually bears on several claims, so
// "one act per source per sitting" capped red's own evidence-gathering at the first one. The
// remaining half of the old contract still holds and is asserted below — the same source for the
// same CLAIM is still one act.
func TestBoardCountsCiteEvents(t *testing.T) {
	runDir := seatRun(t)
	// THE QUOTES ARE REAL SPANS from the seeded report. A supporting corroboration splices an
	// anchor at the claim, so the claim must be in the live document — the same rule blue's cite
	// is held to, and the reason a corroboration of text blue has since edited away is refused
	// rather than spliced blind.
	const (
		claimA = "§1 first — a finding sits in sec 1 here."
		claimB = "§2 the finding prose lands in a quoted sentence."
		claimC = "the parser accepts an empty body in this line."
	)
	cites := []struct{ claim, ref string }{
		{claimA, "https://example.com/a"},
		{claimB, "https://example.com/b"},
	}
	corroborate := func(claim, url, title string) error {
		// --independent by construction: these are sources red went and found, not citations blue
		// authored, so there is no anchor to name.
		_, err := run(t, "lens", "corroborate", "--run", runDir, "--seat-id", "red-lens-r1-L1",
			"--quote", claim, "--url", url, "--title", title,
			"--as", "supports", "--confidence", "high", "--reason", "read at the leaf",
			"--access-date", "2026-07-24")
		return err
	}
	for _, c := range cites {
		if err := corroborate(c.claim, c.ref, c.ref); err != nil {
			t.Fatalf("cite %q: %v", c.claim, err)
		}
	}
	// THE SAME SOURCE, A DIFFERENT CLAIM: this is what the label bought. It was refused.
	if err := corroborate(claimC, "https://example.com/a", "a"); err != nil {
		t.Fatalf("one source corroborating a SECOND claim was refused: %v\n"+
			"A strong source usually bears on several claims; keyed on the url, only the first could ever record.", err)
	}
	// AND THE SAME SOURCE FOR THE SAME CLAIM IS STILL ONE ACT. Not by refusal any more — the
	// minted label is fresh every call, so nothing collides — but by returning the anchor already
	// minted. That is the idempotency the url key used to provide, and losing it silently was the
	// cost that made "drop url from keyFields" the wrong answer to this in the first place.
	if err := corroborate(claimC, "https://example.com/a", "a"); err != nil {
		t.Fatalf("a retried corroboration was refused rather than returning its existing anchor: %v", err)
	}

	// THREE, NOT FOUR: three distinct (source, claim) corroborations, and the retry above added
	// nothing. This counted 2 when a source could bear on only one claim.
	if b := board(t, runDir, "red-merge-r1"); b.Counts.Citations != 3 {
		t.Errorf("counts.citations = %d, want 3 (three distinct source/claim readings, the retry adding none) — the board is the source for citations_checked", b.Counts.Citations)
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
