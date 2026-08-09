package record

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// Replay is where the log becomes state. Winner selection, dedup and the anomaly
// list are the parts that must never silently normalize: an anomaly the render
// hides is a divergence nobody can explain afterwards.

// writeShard puts a shard on disk directly, so a test can construct the
// multi-nonce and torn-file situations a crash produces.
func writeShard(t *testing.T, runDir, seatID, nonce string, evs []Event) string {
	t.Helper()
	if err := os.MkdirAll(recordsDir(runDir), 0o755); err != nil {
		t.Fatal(err)
	}
	p := shardPath(runDir, seatID, nonce)
	var b strings.Builder
	for _, ev := range evs {
		line, err := marshalEvent(ev)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func ev(seatID, nonce string, seq, round int, typ, key string, p *Payload) Event {
	if p == nil {
		p = NewPayload()
	}
	return Event{Seq: seq, SeatID: seatID, Nonce: nonce, Round: round, Type: typ, Key: key, Payload: p}
}

func TestReadShardDropsUnparseableLinesWithoutLosingTheRest(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "events-red-lens-r1-L1-aaaaaaaa.jsonl")
	good := `{"seq":0,"seatId":"red-lens-r1-L1","nonce":"aaaaaaaa","round":1,"type":"register","key":"k0","payload":{}}`
	good2 := `{"seq":2,"seatId":"red-lens-r1-L1","nonce":"aaaaaaaa","round":1,"type":"finding","key":"k2","payload":{"text":"kept"}}`
	content := strings.Join([]string{
		good,
		`{"seq":1,"seatId":"red-lens`, // torn mid-line
		"",                            // blank
		`not json at all`,
		`{"seq":9}`, // parseable but sparse — still an event
		good2,
	}, "\n") + "\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	evs, err := ReadShard(p)
	if err != nil {
		t.Fatalf("a shard with torn lines must not be an ERROR: %v", err)
	}
	if len(evs) != 3 {
		t.Fatalf("recovered %d events, want 3 (two whole plus the sparse one)", len(evs))
	}
	if evs[0].Type != "register" || evs[2].Payload.Str("text") != "kept" {
		t.Errorf("surviving events are wrong: %+v", evs)
	}
	// The file is NOT rewritten: the fragment stays visible on disk.
	after, _ := os.ReadFile(p)
	if string(after) != content {
		t.Error("ReadShard rewrote the shard; a torn fragment must stay visible for the audit")
	}
}

func TestReadShardOnMissingFileIsEmptyNotAnError(t *testing.T) {
	evs, err := ReadShard(filepath.Join(t.TempDir(), "nope.jsonl"))
	if err != nil {
		t.Fatalf("missing shard = %v, want no error", err)
	}
	if len(evs) != 0 {
		t.Errorf("missing shard yielded %d events", len(evs))
	}
}

// A shard with no trailing newline (the crash shape) must still yield its last
// whole event.
func TestReadShardHandlesAMissingFinalNewline(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "s.jsonl")
	line := `{"seq":0,"seatId":"a","type":"register","key":"k","payload":{}}`
	if err := os.WriteFile(p, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	evs, err := ReadShard(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Errorf("recovered %d events from an unterminated shard, want 1", len(evs))
	}
}

// Multi-nonce winner selection: the TERMINAL event decides, and mtime is only
// the explicit fallback. Getting this backwards would silently prefer a crashed
// instance's shard over the one that finished the seat's contract.
func TestMergedEventsWinnerSelection(t *testing.T) {
	t.Run("terminal event wins over a newer shard", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "red-merge-r1"
		withTerminal := writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
			ev(seatID, "aaaaaaaa", 0, 1, "register", seatID+":register:aaaaaaaa", nil),
			ev(seatID, "aaaaaaaa", 1, 1, "verdict", seatID+":verdict", NewPayload().Set("verdict", "PASS")),
		})
		withoutTerminal := writeShard(t, runDir, seatID, "bbbbbbbb", []Event{
			ev(seatID, "bbbbbbbb", 0, 1, "register", seatID+":register:bbbbbbbb", nil),
			ev(seatID, "bbbbbbbb", 1, 1, "position", seatID+":position", NewPayload().Set("text", "loser")),
		})
		// Make the NON-terminal shard newer, so mtime alone would pick the wrong one.
		old := time.Now().Add(-time.Hour)
		if err := os.Chtimes(withTerminal, old, old); err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		if err := os.Chtimes(withoutTerminal, now, now); err != nil {
			t.Fatal(err)
		}

		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range m.Events {
			if e.Payload.Str("text") == "loser" {
				t.Error("the shard WITHOUT the terminal event won despite being newer")
			}
		}
		var sawVerdict bool
		for _, e := range m.Events {
			if e.Type == "verdict" {
				sawVerdict = true
			}
		}
		if !sawVerdict {
			t.Error("the terminal event did not survive the merge")
		}
		if len(m.Anomalies) != 1 || !strings.Contains(m.Anomalies[0], "by terminal event") {
			t.Errorf("anomaly must state the selection basis, got %v", m.Anomalies)
		}
		if !strings.Contains(m.Anomalies[0], "2 dispatches") {
			t.Errorf("anomaly must count the dispatches, got %v", m.Anomalies)
		}
	})

	t.Run("with no terminal event, the latest mtime wins", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "red-lens-r1-L1"
		older := writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
			ev(seatID, "aaaaaaaa", 0, 1, "finding", seatID+":finding:F1", NewPayload().Set("label", "F1").Set("text", "older")),
		})
		newer := writeShard(t, runDir, seatID, "bbbbbbbb", []Event{
			ev(seatID, "bbbbbbbb", 0, 1, "finding", seatID+":finding:F2", NewPayload().Set("label", "F2").Set("text", "newer")),
		})
		old := time.Now().Add(-time.Hour)
		if err := os.Chtimes(older, old, old); err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		if err := os.Chtimes(newer, now, now); err != nil {
			t.Fatal(err)
		}
		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Events) != 1 || m.Events[0].Payload.Str("text") != "newer" {
			t.Errorf("mtime fallback picked the wrong shard: %+v", m.Events)
		}
		if len(m.Anomalies) != 1 || !strings.Contains(m.Anomalies[0], "by mtime fallback") {
			t.Errorf("anomaly must state the mtime fallback, got %v", m.Anomalies)
		}
	})

	t.Run("a revision is terminal too", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "blue-lane-1"
		writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
			ev(seatID, "aaaaaaaa", 0, 0, "revision", seatID+":revision", NewPayload().Set("text", "won")),
		})
		writeShard(t, runDir, seatID, "bbbbbbbb", []Event{
			ev(seatID, "bbbbbbbb", 0, 0, "friction", seatID+":friction:#1", NewPayload().Set("text", "lost")),
		})
		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Events) != 1 || m.Events[0].Type != "revision" {
			t.Errorf("revision was not treated as terminal: %+v", m.Events)
		}
	})

	t.Run("a single shard produces no anomaly", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "red-lens-r1-L1"
		writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
			ev(seatID, "aaaaaaaa", 0, 1, "finding", seatID+":finding:F1", NewPayload().Set("label", "F1")),
		})
		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Anomalies) != 0 {
			t.Errorf("a clean single-shard seat produced anomalies: %v", m.Anomalies)
		}
	})
}

// Global ordering is round, then seat, then seq — deterministic across shards.
func TestMergedEventsGlobalOrderIsDeterministic(t *testing.T) {
	runDir := t.TempDir()
	writeShard(t, runDir, "red-lens-r2-L1", "aaaaaaaa", []Event{
		ev("red-lens-r2-L1", "aaaaaaaa", 1, 2, "finding", "b:1", NewPayload().Set("label", "r2-second")),
		ev("red-lens-r2-L1", "aaaaaaaa", 0, 2, "finding", "b:0", NewPayload().Set("label", "r2-first")),
	})
	writeShard(t, runDir, "red-lens-r1-L1", "bbbbbbbb", []Event{
		ev("red-lens-r1-L1", "bbbbbbbb", 0, 1, "finding", "a:0", NewPayload().Set("label", "r1-first")),
	})
	writeShard(t, runDir, "red-merge-r1", "cccccccc", []Event{
		ev("red-merge-r1", "cccccccc", 0, 1, "position", "c:0", NewPayload().Set("label", "r1-merge")),
	})
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"r1-first", "r1-merge", "r2-first", "r2-second"}
	if len(m.Events) != len(want) {
		t.Fatalf("got %d events, want %d", len(m.Events), len(want))
	}
	for i, w := range want {
		if got := m.Events[i].Payload.Str("label"); got != w {
			t.Errorf("event %d = %q, want %q", i, got, w)
		}
	}
}

// Dedup is by key, and register is EXEMPT — two registers are the double
// dispatch, and collapsing them would hide it.
func TestMergedEventsDedupByKeyExceptRegister(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-lens-r1-L1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "register", seatID+":register:aaaaaaaa", nil),
		ev(seatID, "aaaaaaaa", 1, 1, "finding", seatID+":finding:F1", NewPayload().Set("label", "F1").Set("text", "first")),
		ev(seatID, "aaaaaaaa", 2, 1, "finding", seatID+":finding:F1", NewPayload().Set("label", "F1").Set("text", "duplicate")),
	})
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.Type == "finding" {
			findings++
			if e.Payload.Str("text") != "first" {
				t.Errorf("dedup kept the LATER duplicate: %q", e.Payload.Str("text"))
			}
		}
	}
	if findings != 1 {
		t.Errorf("%d findings survived dedup, want 1", findings)
	}
	if len(m.Anomalies) != 1 || !strings.Contains(m.Anomalies[0], "duplicate key dedup'd") {
		t.Errorf("dedup must be announced as an anomaly, got %v", m.Anomalies)
	}
}

func TestMergedEventsOnAnEmptyOrAbsentRun(t *testing.T) {
	m, err := MergedEvents(filepath.Join(t.TempDir(), "no-such-run"))
	if err != nil {
		t.Fatalf("absent run dir = %v, want no error", err)
	}
	if len(m.Events) != 0 || len(m.Anomalies) != 0 {
		t.Errorf("absent run produced state: %+v", m)
	}

	runDir := t.TempDir()
	if err := os.MkdirAll(recordsDir(runDir), 0o755); err != nil {
		t.Fatal(err)
	}
	// Files that are not shards must be ignored, not parsed.
	for _, name := range []string{"ledger.md", ".lock-render", "class-registry.json", "events-bad.jsonl", "events-x-nothex.jsonl"} {
		if err := os.WriteFile(filepath.Join(recordsDir(runDir), name), []byte("{}"), 0o644); err != nil {
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

func TestRoundOf(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"red-lens-r3-L5", 3},
		{"red-merge-r12", 12},
		{"judge-r1", 1},
		{"blue-respond-r7", 7},
		{"frontier", 0},
		{"blue-synthesize", 0},
		{"assemble", 0},
		{"judge-petition", 0},
		{"", 0},
		{"no-round-here", 0},
		{"red-lens-r0-L1", 0},
		{"first-r2-then-r9", 2}, // the FIRST match wins, not the last
		// The round marker is a HYPHEN-DELIMITED segment of an engine-assigned seat
		// id, so a bare "r5" is round 0 — there is no seat named that, and matching
		// it would also make "frontier" round 0 by accident rather than by rule.
		{"r5", 0},
		{"red-lens-rX-L1", 0},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := RoundOf(tc.in); got != tc.want {
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

	// GradeStr is the `MASS[g] ?? 0` guard: a non-string contributes nothing.
	for _, v := range []any{nil, true, 3, 3.5, []string{"high"}} {
		if got := GradeStr(v); got != "" {
			t.Errorf("GradeStr(%v) = %q, want empty", v, got)
		}
	}
	if got := GradeStr("high"); got != "high" {
		t.Errorf("GradeStr(\"high\") = %q", got)
	}
}

// Ids are minted tool-side and sequentially PER ROUND: the collision class that
// produced four different "R5-1"s cannot recur.
func TestMintGapIDIsSequentialPerRound(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, _, err := RegisterSeat(runDir, seatID); err != nil {
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
		if _, err := Append(runDir, seatID, "mint", NewPayload().
			Set("gap_id", got).Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
			t.Fatal(err)
		}
	}
	// A new round restarts the counter; the id namespace is per-round.
	seat2 := "red-merge-r2"
	if _, _, err := RegisterSeat(runDir, seat2); err != nil {
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
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	if _, _, err := RegisterSeat(runDir, seatID); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(runDir, seatID, "mint", NewPayload().
		Set("gap_id", "R1-1").Set("mint_key", "L1-F3").Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
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
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p1").Set("severity", "low").
			Set("likelihood", "low").Set("impact", "low")),
		ev(seatID, "aaaaaaaa", 1, 1, "mint", seatID+":mint:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("problem", "p2").Set("severity", "high")),
		// A regrade moves ONLY the keys it carries.
		ev(seatID, "aaaaaaaa", 2, 1, "regrade", seatID+":regrade:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("severity", "certain").Set("basis", "new evidence")),
		ev(seatID, "aaaaaaaa", 3, 1, "close", seatID+":close:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("closure_class", "closed")),
		// A regrade and a close of an UNKNOWN gap are ignored, not fatal.
		ev(seatID, "aaaaaaaa", 4, 1, "regrade", seatID+":regrade:R9-9", NewPayload().Set("gap_id", "R9-9").Set("severity", "low")),
		ev(seatID, "aaaaaaaa", 5, 1, "close", seatID+":close:R9-9", NewPayload().Set("gap_id", "R9-9")),
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
	if g1.Severity != "certain" {
		t.Errorf("regrade did not move severity: %v", g1.Severity)
	}
	// The keys the regrade did NOT carry are untouched.
	if g1.Likelihood != "low" || g1.Impact != "low" {
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
	if g2.ComplexityCost != nil {
		t.Errorf("an unpassed grade became %v, want nil (it renders as \"undefined\")", g2.ComplexityCost)
	}
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
	runDir := t.TempDir()
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []Event{
		ev(lens, "aaaaaaaa", 0, 1, "finding", lens+":finding:F1", NewPayload().Set("label", "F1").Set("text", "first")),
		ev(lens, "aaaaaaaa", 1, 1, "finding", lens+":finding:F2", NewPayload().Set("label", "F2").Set("text", "second")),
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
			if o.Payload.Str("label") == want {
				found = true
			}
		}
		if !found {
			t.Errorf("finding %s lost its label in replay — found_by credit is keyed on it", want)
		}
	}
}

func TestValidateGradeEnumOnEveryGradedField(t *testing.T) {
	for _, field := range []string{"severity", "likelihood", "impact", "complexity_cost"} {
		t.Run(field+"/rejects a non-grade", func(t *testing.T) {
			err := validate(t.TempDir(), "red-merge-r1", "mint", NewPayload().Set(field, "catastrophic"))
			if err == nil {
				t.Fatalf("%s=catastrophic was accepted", field)
			}
			want := fmt.Sprintf("mint.%s=%q not in grade enum", field, "catastrophic")
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to contain %q", err, want)
			}
		})
		t.Run(field+"/rejects a non-string", func(t *testing.T) {
			// A bare boolean flag must fail the same way `true` does in the oracle.
			err := validate(t.TempDir(), "red-merge-r1", "mint", NewPayload().Set(field, true))
			if err == nil {
				t.Fatalf("%s=true was accepted", field)
			}
			if !strings.Contains(err.Error(), "=true not in grade enum") {
				t.Errorf("a boolean must render bare in the message, got %q", err)
			}
		})
		t.Run(field+"/accepts every canonical grade", func(t *testing.T) {
			for _, g := range flags.GradeNames() {
				p := NewPayload().Set(field, g).Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")
				if err := validate(t.TempDir(), "red-merge-r1", "mint", p); err != nil {
					t.Errorf("%s=%s refused: %v", field, g, err)
				}
			}
		})
	}
	// An ABSENT graded field is fine; only a present-and-wrong one is refused.
	if err := validate(t.TempDir(), "red-merge-r1", "regrade", NewPayload().Set("basis", "b")); err != nil {
		t.Errorf("absent grades were refused: %v", err)
	}
}

func TestValidateVerbContracts(t *testing.T) {
	cases := []struct {
		name    string
		typ     string
		p       *Payload
		wantErr string // empty means it must be ACCEPTED
	}{
		{"mint without --check", "mint", NewPayload().Set("class", "x"), "mint requires --check"},
		{"mint with an empty --check", "mint", NewPayload().Set("class", "x").Set("acceptance_check", ""), "mint requires --check"},
		{"mint without --class", "mint", NewPayload().Set("acceptance_check", "c").Set("check_kind", "document"), "mint requires --class"},
		{"mint complete", "mint", NewPayload().Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p"), ""},

		{"close without --id", "close", NewPayload(), "close requires --id"},
		{"regrade without --basis", "regrade", NewPayload(), "regrade requires --reason"},
		{"regrade complete", "regrade", NewPayload().Set("basis", "b"), ""},

		{"retire without --claim", "retire", NewPayload().Set("reason", "r"), "retire requires --claim"},
		{"retire without --reason", "retire", NewPayload().Set("claim", "c"), "retire requires --reason"},
		{"retire complete", "retire", NewPayload().Set("claim", "c").Set("reason", "r"), ""},

		{"avenue with an unknown status", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "shelved").Set("line", "l"), "avenue requires --status"},
		{"avenue with no status at all", "avenue", NewPayload().Set("avenue_id", "A1").Set("line", "l"), "avenue requires --status"},
		{"avenue with no id", "avenue", NewPayload().Set("status", "pursued").Set("line", "l"), "avenue requires an id"},
		{"a deferred avenue needs a reason", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "deferred").Set("line", "l"), "requires --reason"},
		{"avenue without --line", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "pursued"), "avenue requires --line"},
		{"a declined avenue needs a reason", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "declined").Set("line", "l"), "requires --reason"},
		{"an abandoned avenue needs a reason", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "abandoned").Set("line", "l"), "requires --reason"},
		{"a PURSUED avenue does not need a reason", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "pursued").Set("line", "l"), ""},
		{"a declined avenue with a reason", "avenue", NewPayload().Set("avenue_id", "A1").Set("status", "declined").Set("line", "l").Set("reason", "why"), ""},

		// The message must name the flag the PARSER accepts. It named --gap-id for as
		// long as that flag existed and kept naming it after the rename, because the
		// spelling was derived from the payload key rather than stated.
		{"opinion missing every field", "opinion", NewPayload(), "opinion requires --id"},
		{"an unknown verb is not validated here", "friction", NewPayload(), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(t.TempDir(), "red-merge-r1", tc.typ, tc.p)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("validate = %v, want accepted", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validate accepted %s with %v", tc.typ, tc.p.Keys())
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
	runDir := t.TempDir()
	if _, _, err := RegisterSeat(runDir, "red-merge-r1"); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(runDir, "red-merge-r1", "mint", NewPayload().Set("gap_id", id).
		Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").
		Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")); err != nil {
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
	}
	all := []string{"gap_id", "disposition", "principle", "tension", "review_flag"}
	for _, missing := range all {
		t.Run("missing "+missing, func(t *testing.T) {
			// The gap must EXIST and be NAMED CORRECTLY: references are checked at
			// write time now, so gap_id carries the real minted id rather than the
			// "x" placeholder every other field uses. With a bogus id the reference
			// check fires first and this test would assert on the wrong refusal.
			runDir := opinionRunDir(t)
			p := NewPayload()
			for _, f := range all {
				if f == missing {
					continue
				}
				if f == "gap_id" {
					p.Set(f, "R1-1")
					continue
				}
				p.Set(f, "x")
			}
			err := validate(runDir, "judge-r1", "opinion", p)
			if err == nil {
				t.Fatalf("opinion accepted without %s", missing)
			}
			if !strings.Contains(err.Error(), wantFlag[missing]) {
				t.Errorf("error %q does not name %q — the seat's only teacher is this string", err, wantFlag[missing])
			}
		})
	}
	complete := opinionRunDir(t)
	p := NewPayload()
	for _, f := range all {
		if f == "gap_id" {
			p.Set(f, "R1-1")
			continue
		}
		p.Set(f, "x")
	}
	p.Set("rationale", "the ruling's reasoning")
	// disposition is a CLOSED set since #342 — a placeholder is no longer a legal value.
	p.Set("disposition", "closed")
	if err := validate(complete, "judge-r1", "opinion", p); err != nil {
		t.Errorf("a complete opinion was refused: %v", err)
	}
	// An EMPTY value still counts as present for the five Has-checked fields: the check
	// is Has, not non-empty (a --review-flag false is a legitimate ruling). rationale is
	// the exception — it is prose the ruling turns on, so it is required non-empty.
	q := NewPayload()
	for _, f := range all {
		q.Set(f, "")
	}
	q.Set("rationale", "the ruling's reasoning")
	// DISPOSITION IS THE EXCEPTION TO THE EXCEPTION (#342). The Has-not-empty rule exists for
	// fields like --review-flag, where "false" is a legitimate ruling. It never applied to the
	// disposition: an EMPTY disposition rules nothing, and it was only ever accepted because
	// the set was open. Now the set is closed, so it must be one of the words.
	q.Set("disposition", DispositionCarried)
	if err := validate(complete, "judge-r1", "opinion", q); err != nil {
		t.Errorf("opinion fields present-but-empty were refused: %v", err)
	}
}

// Lineage is never dangling: a mint that supersedes an id no mint created is
// refused, because the supersedes chain is what the whole analysis reads.
func TestValidateRefusesDanglingLineage(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().Set("gap_id", "R1-1")),
	})
	base := func() *Payload {
		return NewPayload().Set("acceptance_check", "c").Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")
	}

	if err := validate(runDir, "red-merge-r1", "mint", base().Set("supersedes", []string{"R1-1"})); err != nil {
		t.Errorf("a real ancestor was refused: %v", err)
	}
	err := validate(runDir, "red-merge-r1", "mint", base().Set("supersedes", []string{"R1-1", "R9-9"}))
	if err == nil {
		t.Fatal("a dangling ancestor was accepted")
	}
	if !strings.Contains(err.Error(), "R9-9") || !strings.Contains(err.Error(), "dangling lineage refused") {
		t.Errorf("error must name the dangling id: %v", err)
	}
	// An empty lineage is not a dangling one.
	if err := validate(runDir, "red-merge-r1", "mint", base().Set("supersedes", []string{})); err != nil {
		t.Errorf("an empty lineage was refused: %v", err)
	}
}

func TestValidateCloseAnchorContract(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	// BOTH gaps are minted: R1-1 is the one being closed, R2-1 the successor a
	// closed_with_regression names. A successor is a reference like any other and is
	// now checked, so a fixture that names one has to create it.
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().Set("gap_id", "R1-1")),
		ev(seatID, "aaaaaaaa", 1, 2, "mint", seatID+":mint:R2-1", NewPayload().Set("gap_id", "R2-1")),
	})
	anchored := func() *Payload {
		return NewPayload().Set("gap_id", "R1-1").
			Set("anchor_seat", "L1").Set("anchor_tool", "git show").Set("anchor_target", "7bc501e:path").
			Set("prose", "verified against the ref; the claim now holds")
	}
	cases := []struct {
		name    string
		p       *Payload
		wantErr string
	}{
		{"unknown gap", NewPayload().Set("gap_id", "R9-9"), "close of unknown gap"},
		{"no anchor at all", NewPayload().Set("gap_id", "R1-1"), "requires the attestation anchor"},
		{"a PARTIAL anchor is not an anchor", NewPayload().Set("gap_id", "R1-1").Set("anchor_seat", "L1"), "requires the attestation anchor"},
		{"anchor missing its target", NewPayload().Set("gap_id", "R1-1").Set("anchor_seat", "L1").Set("anchor_tool", "git show"), "requires the attestation anchor"},
		{"a full anchor", anchored(), ""},
		// --carried-from remains the honest alternative to re-verifying, but it is a
		// CLAIM ABOUT THE RECORD and is now checked like one. This gap has never been
		// closed, so "carried from round 2" is simply false — and accepting it was a
		// laundering path: an unanchored first closure that scores as closed, which is
		// exactly what anchored_closures_pct exists to detect. The genuine case (close
		// with an anchor, then restate) is covered in required_test.go.
		{"--carried-from cannot invent an earlier closure", NewPayload().Set("gap_id", "R1-1").Set("carried_from", "2"), "no closure of it exists in the record"},
		{"closed_with_regression needs a successor", anchored().Set("closure_class", "closed_with_regression"), "requires --successor"},
		{"closed_with_regression with a successor", anchored().Set("closure_class", "closed_with_regression").Set("successor", "R2-1"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(runDir, "red-merge-r1", "close", tc.p)
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
		if err := os.MkdirAll(recordsDir(runDir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(recordsDir(runDir), "class-registry.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	registry := `{"classes":[{"slug":"scope-creep"},{"slug":"unfalsifiable"},{"slug":"stale-source"}]}`
	mint := func(p *Payload) *Payload {
		return p.Set("acceptance_check", "c").Set("check_kind", "document").Set("likelihood", "medium").Set("impact", "medium").Set("problem", "p")
	}

	t.Run("no registry staged is advisory, not strict", func(t *testing.T) {
		if err := validate(t.TempDir(), "red-merge-r1", "mint", mint(NewPayload().Set("class", "anything-at-all"))); err != nil {
			t.Errorf("advisory mode refused a class: %v", err)
		}
	})

	t.Run("an unparseable registry degrades to advisory", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, "{not json")
		if err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "anything-at-all"))); err != nil {
			t.Errorf("an unparseable registry made validation strict: %v", err)
		}
	})

	t.Run("a known slug passes", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		if err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "scope-creep"))); err != nil {
			t.Errorf("a registry slug was refused: %v", err)
		}
	})

	t.Run("an unknown slug is refused with a hint", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "invented")))
		if err == nil {
			t.Fatal("an unknown class was accepted")
		}
		if !strings.Contains(err.Error(), "unknown class") || !strings.Contains(err.Error(), "scope-creep") {
			t.Errorf("the refusal must offer real slugs: %v", err)
		}
	})

	t.Run("--class-new requires its full triple", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		for _, missing := range []string{"definition", "neighbor", "distinguisher"} {
			p := mint(NewPayload().Set("class", "brand-new").Set("class_new", true))
			for _, f := range []string{"definition", "neighbor", "distinguisher"} {
				if f == missing {
					continue
				}
				v := "x"
				if f == "neighbor" {
					v = "scope-creep"
				}
				p.Set(f, v)
			}
			err := validate(runDir, "red-merge-r1", "mint", p)
			if err == nil {
				t.Errorf("--class-new accepted without --%s", missing)
				continue
			}
			if !strings.Contains(err.Error(), "--class-new requires") {
				t.Errorf("wrong refusal for missing --%s: %v", missing, err)
			}
		}
	})

	t.Run("--class-new needs a REAL neighbor", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		p := mint(NewPayload().Set("class", "brand-new").Set("class_new", true).
			Set("definition", "d").Set("neighbor", "not-a-class").Set("distinguisher", "q"))
		err := validate(runDir, "red-merge-r1", "mint", p)
		if err == nil {
			t.Fatal("an invented neighbor was accepted")
		}
		if !strings.Contains(err.Error(), "not a known class") {
			t.Errorf("wrong refusal: %v", err)
		}
	})

	t.Run("a class minted earlier in the run extends the registry", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		seatID := "red-merge-r1"
		writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
			ev(seatID, "aaaaaaaa", 0, 1, "class-new", seatID+":class-new:x", NewPayload().Set("slug", "run-local-class")),
		})
		if err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "run-local-class"))); err != nil {
			t.Errorf("a class minted in this run was refused: %v", err)
		}
		// And it is a valid neighbor for a further new class.
		p := mint(NewPayload().Set("class", "another").Set("class_new", true).
			Set("definition", "d").Set("neighbor", "run-local-class").Set("distinguisher", "q"))
		if err := validate(runDir, "red-merge-r1", "mint", p); err != nil {
			t.Errorf("a run-local class was not a valid neighbor: %v", err)
		}
	})

	t.Run("a registry with fewer than six slugs does not slice out of range", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, `{"classes":[{"slug":"only-one"}]}`)
		err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "invented")))
		if err == nil {
			t.Fatal("expected a refusal")
		}
		if !strings.Contains(err.Error(), "only-one") {
			t.Errorf("hint did not include the single slug: %v", err)
		}
	})

	t.Run("an EMPTY registry is still strict and does not panic", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, `{"classes":[]}`)
		if err := validate(runDir, "red-merge-r1", "mint", mint(NewPayload().Set("class", "invented"))); err == nil {
			t.Error("an empty registry accepted an invented class")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	cases := []struct {
		name   string
		typ    string
		p      *Payload
		prior  []Event
		want   string
		seatID string
	}{
		{name: "singleton verbs key on seat+verb", typ: "position", p: NewPayload(), seatID: "red-merge-r1", want: "red-merge-r1:position"},
		{name: "verdict is a singleton", typ: "verdict", p: NewPayload().Set("gap_id", "R1-1"), seatID: "red-merge-r1", want: "red-merge-r1:verdict"},
		{name: "gap_id is the first label consulted", typ: "close", p: NewPayload().Set("gap_id", "R1-1").Set("label", "ignored"), seatID: "red-merge-r1", want: "red-merge-r1:close:R1-1"},
		{name: "label when there is no gap_id", typ: "finding", p: NewPayload().Set("label", "F1"), seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:F1"},
		{name: "id, observation and reference are also labels", typ: "cite", p: NewPayload().Set("reference", "https://x"), seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:cite:https://x"},
		{
			name: "with no label at all, a per-shard ordinal", typ: "friction", p: NewPayload(), seatID: "blue-lane-1",
			prior: []Event{{Type: "friction"}, {Type: "friction"}, {Type: "position"}},
			want:  "blue-lane-1:friction:#3",
		},
		{
			name: "the ordinal counts only the SAME verb", typ: "friction", p: NewPayload(), seatID: "blue-lane-1",
			prior: []Event{{Type: "position"}, {Type: "finding"}},
			want:  "blue-lane-1:friction:#1",
		},
		{name: "an empty label falls through to the ordinal", typ: "finding", p: NewPayload().Set("label", ""), seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:#1"},
		{name: "a non-string label falls through", typ: "finding", p: NewPayload().Set("label", true), seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:#1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveKey(tc.seatID, tc.typ, tc.p, tc.prior); got != tc.want {
				t.Errorf("deriveKey = %q, want %q", got, tc.want)
			}
		})
	}
}

// The same label in a LATER round is not a collision: the key carries the seat
// id, and the seat id carries the round.
func TestDeriveKeySameLabelNextRoundIsNotACollision(t *testing.T) {
	p := NewPayload().Set("label", "F1")
	r1 := deriveKey("red-lens-r1-L1", "finding", p, nil)
	r2 := deriveKey("red-lens-r2-L1", "finding", p, nil)
	if r1 == r2 {
		t.Errorf("the same label in two rounds produced one key %q — the second finding would be dedup'd away", r1)
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
