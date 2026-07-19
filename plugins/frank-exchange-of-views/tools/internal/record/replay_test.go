package record

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
			Set("gap_id", got).Set("acceptance_check", "c").Set("class", "x")); err != nil {
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
		Set("gap_id", "R1-1").Set("mint_key", "L1-F3").Set("acceptance_check", "c").Set("class", "x")); err != nil {
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

// dispose targets the FIRST matching observation, by label when it has one and
// by event key otherwise.
func TestBoardStateDispositionMatching(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r1-L1"
	merge := "red-merge-r1"
	writeShard(t, runDir, lens, "aaaaaaaa", []Event{
		ev(lens, "aaaaaaaa", 0, 1, "finding", lens+":finding:F1", NewPayload().Set("label", "F1").Set("text", "first")),
		ev(lens, "aaaaaaaa", 1, 1, "observe", lens+":observe:#1", NewPayload().Set("text", "unlabelled")),
		ev(lens, "aaaaaaaa", 2, 1, "finding", lens+":finding:F1-dup", NewPayload().Set("label", "F1").Set("text", "same label again")),
	})
	writeShard(t, runDir, merge, "bbbbbbbb", []Event{
		ev(merge, "bbbbbbbb", 0, 1, "dispose", merge+":dispose:F1", NewPayload().Set("observation", "F1").Set("disposition", "minted-as")),
		ev(merge, "bbbbbbbb", 1, 1, "dispose", merge+":dispose:key", NewPayload().Set("observation", lens+":observe:#1").Set("disposition", "declined")),
		ev(merge, "bbbbbbbb", 2, 1, "dispose", merge+":dispose:none", NewPayload().Set("observation", "NOPE").Set("disposition", "declined")),
	})
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Observations) != 3 {
		t.Fatalf("%d observations, want 3", len(b.Observations))
	}
	byText := map[string]*Observation{}
	for _, o := range b.Observations {
		byText[o.Payload.Str("text")] = o
	}
	if o := byText["first"]; o.Disposition == nil || o.Disposition.Str("disposition") != "minted-as" {
		t.Errorf("the FIRST match on a label was not disposed: %+v", o.Disposition)
	}
	// find() semantics: the second observation sharing the label stays undisposed.
	if o := byText["same label again"]; o.Disposition != nil {
		t.Error("dispose reached a SECOND observation with the same label; find() takes the first only")
	}
	// An unlabelled observation is addressable by its event key.
	if o := byText["unlabelled"]; o.Disposition == nil || o.Disposition.Str("disposition") != "declined" {
		t.Errorf("an unlabelled observation could not be disposed by key: %+v", o.Disposition)
	}
	// A dispose naming nothing must not attach itself to an arbitrary observation.
	disposed := 0
	for _, o := range b.Observations {
		if o.Disposition != nil {
			disposed++
		}
	}
	if disposed != 2 {
		t.Errorf("%d observations disposed, want 2 — a non-matching dispose landed somewhere", disposed)
	}
}

func TestValidateGradeEnumOnEveryGradedField(t *testing.T) {
	for _, field := range []string{"severity", "likelihood", "impact", "complexity_cost"} {
		t.Run(field+"/rejects a non-grade", func(t *testing.T) {
			err := validate(t.TempDir(), "mint", NewPayload().Set(field, "catastrophic"))
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
			err := validate(t.TempDir(), "mint", NewPayload().Set(field, true))
			if err == nil {
				t.Fatalf("%s=true was accepted", field)
			}
			if !strings.Contains(err.Error(), "=true not in grade enum") {
				t.Errorf("a boolean must render bare in the message, got %q", err)
			}
		})
		t.Run(field+"/accepts every canonical grade", func(t *testing.T) {
			for _, g := range GRADES {
				p := NewPayload().Set(field, g).Set("acceptance_check", "c").Set("class", "x")
				if err := validate(t.TempDir(), "mint", p); err != nil {
					t.Errorf("%s=%s refused: %v", field, g, err)
				}
			}
		})
	}
	// An ABSENT graded field is fine; only a present-and-wrong one is refused.
	if err := validate(t.TempDir(), "regrade", NewPayload().Set("basis", "b")); err != nil {
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
		{"mint without --class", "mint", NewPayload().Set("acceptance_check", "c"), "mint requires --class"},
		{"mint complete", "mint", NewPayload().Set("acceptance_check", "c").Set("class", "x"), ""},

		{"close without --id", "close", NewPayload(), "close requires --id"},
		{"dispose without --as", "dispose", NewPayload(), "dispose requires --as"},
		{"dispose complete", "dispose", NewPayload().Set("disposition", "declined"), ""},
		{"regrade without --basis", "regrade", NewPayload(), "regrade requires --basis"},
		{"regrade complete", "regrade", NewPayload().Set("basis", "b"), ""},

		{"retire without --claim", "retire", NewPayload().Set("reason", "r"), "retire requires --claim"},
		{"retire without --reason", "retire", NewPayload().Set("claim", "c"), "retire requires --reason"},
		{"retire complete", "retire", NewPayload().Set("claim", "c").Set("reason", "r"), ""},

		{"avenue with an unknown status", "avenue", NewPayload().Set("status", "shelved").Set("line", "l"), "avenue requires --status declined|abandoned|pursued"},
		{"avenue with no status at all", "avenue", NewPayload().Set("line", "l"), "avenue requires --status"},
		{"avenue without --line", "avenue", NewPayload().Set("status", "pursued"), "avenue requires --line"},
		{"a declined avenue needs a reason", "avenue", NewPayload().Set("status", "declined").Set("line", "l"), "requires --reason"},
		{"an abandoned avenue needs a reason", "avenue", NewPayload().Set("status", "abandoned").Set("line", "l"), "requires --reason"},
		{"a PURSUED avenue does not need a reason", "avenue", NewPayload().Set("status", "pursued").Set("line", "l"), ""},
		{"a declined avenue with a reason", "avenue", NewPayload().Set("status", "declined").Set("line", "l").Set("reason", "why"), ""},

		{"opinion missing every field", "opinion", NewPayload(), "opinion requires --gap-id"},
		{"an unknown verb is not validated here", "friction", NewPayload(), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(t.TempDir(), tc.typ, tc.p)
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
func TestValidateOpinionNamesEachMissingField(t *testing.T) {
	all := []string{"gap_id", "disposition", "principle", "tension", "review_flag"}
	for _, missing := range all {
		t.Run("missing "+missing, func(t *testing.T) {
			p := NewPayload()
			for _, f := range all {
				if f != missing {
					p.Set(f, "x")
				}
			}
			err := validate(t.TempDir(), "opinion", p)
			if err == nil {
				t.Fatalf("opinion accepted without %s", missing)
			}
			wantFlag := "--" + strings.Replace(missing, "_", "-", 1)
			if !strings.Contains(err.Error(), wantFlag) {
				t.Errorf("error %q does not name %q", err, wantFlag)
			}
		})
	}
	p := NewPayload()
	for _, f := range all {
		p.Set(f, "x")
	}
	if err := validate(t.TempDir(), "opinion", p); err != nil {
		t.Errorf("a complete opinion was refused: %v", err)
	}
	// An EMPTY value still counts as present: the check is Has, not non-empty.
	q := NewPayload()
	for _, f := range all {
		q.Set(f, "")
	}
	if err := validate(t.TempDir(), "opinion", q); err != nil {
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
	base := func() *Payload { return NewPayload().Set("acceptance_check", "c").Set("class", "x") }

	if err := validate(runDir, "mint", base().Set("supersedes", []string{"R1-1"})); err != nil {
		t.Errorf("a real ancestor was refused: %v", err)
	}
	err := validate(runDir, "mint", base().Set("supersedes", []string{"R1-1", "R9-9"}))
	if err == nil {
		t.Fatal("a dangling ancestor was accepted")
	}
	if !strings.Contains(err.Error(), "R9-9") || !strings.Contains(err.Error(), "dangling lineage refused") {
		t.Errorf("error must name the dangling id: %v", err)
	}
	// An empty lineage is not a dangling one.
	if err := validate(runDir, "mint", base().Set("supersedes", []string{})); err != nil {
		t.Errorf("an empty lineage was refused: %v", err)
	}
}

func TestValidateCloseAnchorContract(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().Set("gap_id", "R1-1")),
	})
	anchored := func() *Payload {
		return NewPayload().Set("gap_id", "R1-1").
			Set("anchor_seat", "L1").Set("anchor_tool", "git show").Set("anchor_target", "7bc501e:path")
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
		{"--carried-from is the honest alternative", NewPayload().Set("gap_id", "R1-1").Set("carried_from", "2"), ""},
		{"closed_with_regression needs a successor", anchored().Set("closure_class", "closed_with_regression"), "requires --successor"},
		{"closed_with_regression with a successor", anchored().Set("closure_class", "closed_with_regression").Set("successor", "R2-1"), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(runDir, "close", tc.p)
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
	mint := func(p *Payload) *Payload { return p.Set("acceptance_check", "c") }

	t.Run("no registry staged is advisory, not strict", func(t *testing.T) {
		if err := validate(t.TempDir(), "mint", mint(NewPayload().Set("class", "anything-at-all"))); err != nil {
			t.Errorf("advisory mode refused a class: %v", err)
		}
	})

	t.Run("an unparseable registry degrades to advisory", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, "{not json")
		if err := validate(runDir, "mint", mint(NewPayload().Set("class", "anything-at-all"))); err != nil {
			t.Errorf("an unparseable registry made validation strict: %v", err)
		}
	})

	t.Run("a known slug passes", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		if err := validate(runDir, "mint", mint(NewPayload().Set("class", "scope-creep"))); err != nil {
			t.Errorf("a registry slug was refused: %v", err)
		}
	})

	t.Run("an unknown slug is refused with a hint", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		err := validate(runDir, "mint", mint(NewPayload().Set("class", "invented")))
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
			err := validate(runDir, "mint", p)
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
		err := validate(runDir, "mint", p)
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
		if err := validate(runDir, "mint", mint(NewPayload().Set("class", "run-local-class"))); err != nil {
			t.Errorf("a class minted in this run was refused: %v", err)
		}
		// And it is a valid neighbor for a further new class.
		p := mint(NewPayload().Set("class", "another").Set("class_new", true).
			Set("definition", "d").Set("neighbor", "run-local-class").Set("distinguisher", "q"))
		if err := validate(runDir, "mint", p); err != nil {
			t.Errorf("a run-local class was not a valid neighbor: %v", err)
		}
	})

	t.Run("a registry with fewer than six slugs does not slice out of range", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, `{"classes":[{"slug":"only-one"}]}`)
		err := validate(runDir, "mint", mint(NewPayload().Set("class", "invented")))
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
		if err := validate(runDir, "mint", mint(NewPayload().Set("class", "invented"))); err == nil {
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
