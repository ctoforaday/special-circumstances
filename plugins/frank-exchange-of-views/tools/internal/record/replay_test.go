package record

import (
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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
func writeShard(t *testing.T, runDir, seatID, nonce string, evs []*Event) string {
	t.Helper()
	if err := os.MkdirAll(recordsDirT(runDir), 0o755); err != nil {
		t.Fatal(err)
	}
	p := shardPath(recordsDirT(runDir), seatID, nonce)
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
	good2 := `{"seq":2,"seatId":"red-lens-r1-L1","nonce":"aaaaaaaa","round":1,"type":"finding","key":"k2","payload":{"reason":"kept"}}`
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
	evs, _, err := ReadShard(p)
	if err != nil {
		t.Fatalf("a shard with torn lines must not be an ERROR: %v", err)
	}
	if len(evs) != 3 {
		t.Fatalf("recovered %d events, want 3 (two whole plus the sparse one)", len(evs))
	}
	if evs[0].Type != "register" || evs[2].Payload.Str("reason") != "kept" {
		t.Errorf("surviving events are wrong: %+v", evs)
	}
	// The file is NOT rewritten: the fragment stays visible on disk.
	after, _ := os.ReadFile(p)
	if string(after) != content {
		t.Error("ReadShard rewrote the shard; a torn fragment must stay visible for the audit")
	}
}

func TestReadShardOnMissingFileIsEmptyNotAnError(t *testing.T) {
	evs, _, err := ReadShard(filepath.Join(t.TempDir(), "nope.jsonl"))
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
	evs, _, err := ReadShard(p)
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
		withTerminal := writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
			recordtest.At(t, seatID, 1, seatID+":register:aaaaaaaa", &recordpb.Register{}),
			recordtest.At(t, seatID, 1, seatID+":verdict", &recordpb.RoundVerdict{
				// The seat types PASS and the schema spells `pass`; recordpb.BySpelling is case-sensitive
				// by design, so the fold is the join's job and the fixture states the VALUE.
				Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS),
			}),
		})
		withoutTerminal := writeShard(t, runDir, seatID, "bbbbbbbb", []*Event{
			recordtest.At(t, seatID, 1, seatID+":register:bbbbbbbb", &recordpb.Register{}),
			recordtest.At(t, seatID, 1, seatID+":position", &recordpb.Position{Text: proto.String("loser")}),
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
			if e.Payload.Str("reason") == "loser" {
				t.Error("the shard WITHOUT the terminal event won despite being newer")
			}
		}
		var sawVerdict bool
		for _, e := range m.Events {
			if e.GetType() == recordpb.EventType_EVENT_TYPE_VERDICT {
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

	// WITH NO TERMINAL EVENT, THE RECORD'S OWN STAMP DECIDES — AND IT BEATS THE FILESYSTEM.
	//
	// This used to assert "the latest mtime wins", and mtime is metadata about the CARRIER rather
	// than a fact on the record: a git checkout, an rsync, a container copy or a restore from the
	// recovery mirror all rewrite it. It was also undecidable where two shards share an mtime,
	// which is what Windows produces for two writes in the same tick — measured 2026-08-16, the
	// same commit green on ubuntu and red on windows-latest.
	//
	// THE MTIMES ARE DELIBERATELY INVERTED HERE: the shard that must win is the OLDER file. A
	// selection that still consults the filesystem cannot pass this.
	t.Run("with no terminal event, the latest recorded stamp wins over the newer file", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "red-lens-r1-L1"
		early := recordtest.At(t, seatID, 1, seatID+":finding:F1", &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("older")})
		early.TS = "2026-01-01T00:00:00.000000000Z"
		late := recordtest.At(t, seatID, 1, seatID+":finding:F2", &recordpb.Finding{Label: proto.String("F2"), Text: proto.String("newer")})
		late.TS = "2026-01-01T00:00:00.000000001Z" // one nanosecond, which is what nextStamp issues on a tie
		loser := writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{early})
		winner := writeShard(t, runDir, seatID, "bbbbbbbb", []*Event{late})

		now := time.Now()
		if err := os.Chtimes(loser, now, now); err != nil { // the LOSER is the newest file
			t.Fatal(err)
		}
		old := now.Add(-time.Hour)
		if err := os.Chtimes(winner, old, old); err != nil {
			t.Fatal(err)
		}

		m, err := MergedEvents(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Events) != 1 || m.Events[0].Payload.Str("reason") != "newer" {
			t.Errorf("the shard with the later RECORDED stamp must win even though its file is older — the record owns this ordering, not the filesystem: %+v", m.Events)
		}
		if len(m.Anomalies) != 1 || !strings.Contains(m.Anomalies[0], "by latest recorded event") {
			t.Errorf("anomaly must state the selection basis, got %v", m.Anomalies)
		}
	})

	t.Run("a revision is terminal too", func(t *testing.T) {
		runDir := t.TempDir()
		seatID := "blue-lane-1"
		writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
			recordtest.At(t, seatID, 0, seatID+":revision", &recordpb.Revision{Text: proto.String("won")}),
		})
		writeShard(t, runDir, seatID, "bbbbbbbb", []*Event{
			recordtest.At(t, seatID, 0, seatID+":friction:#1", &recordpb.Friction{Text: proto.String("lost")}),
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
		writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
			recordtest.At(t, seatID, 1, seatID+":finding:F1", &recordpb.Finding{Label: proto.String("F1")}),
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
	writeShard(t, runDir, "red-lens-r2-L1", "aaaaaaaa", []*Event{
		recordtest.At(t, "red-lens-r2-L1", 2, "b:1", &recordpb.Finding{Label: proto.String("r2-second")}),
		recordtest.At(t, "red-lens-r2-L1", 2, "b:0", &recordpb.Finding{Label: proto.String("r2-first")}),
	})
	writeShard(t, runDir, "red-lens-r1-L1", "bbbbbbbb", []*Event{
		recordtest.At(t, "red-lens-r1-L1", 1, "a:0", &recordpb.Finding{Label: proto.String("r1-first")}),
	})
	writeShard(t, runDir, "red-merge-r1", "cccccccc", []*Event{
		recordtest.At(t, "red-merge-r1", 1, "c:0", &recordpb.Position{}),
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
	writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
		recordtest.At(t, seatID, 1, seatID+":register:aaaaaaaa", &recordpb.Register{}),
		recordtest.At(t, seatID, 1, seatID+":finding:F1", &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("first")}),
		recordtest.At(t, seatID, 1, seatID+":finding:F1", &recordpb.Finding{Label: proto.String("F1"), Text: proto.String("duplicate")}),
	})
	m, err := MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	findings := 0
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_FINDING {
			findings++
			if e.Payload.Str("reason") != "first" {
				t.Errorf("dedup kept the LATER duplicate: %q", e.Payload.Str("reason"))
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
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seatID, Round: RoundOf(seatID)}); err != nil {
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
		if _, err := Append(Identity{RunDir: runDir, SeatID: seatID, Round: RoundOf(seatID)}, &recordpb.Mint{AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
			t.Fatal(err)
		}
	}
	// A new round restarts the counter; the id namespace is per-round.
	seat2 := "red-merge-r2"
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seat2, Round: RoundOf(seat2)}); err != nil {
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
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: seatID, Round: RoundOf(seatID)}); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: seatID, Round: RoundOf(seatID)}, &recordpb.Mint{MintKey: proto.String("L1-F3"), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
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
	writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{Problem: proto.String("p1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, seatID, 1, seatID+":mint:R1-2", &recordpb.Mint{Problem: proto.String("p2"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		// A regrade moves ONLY the keys it carries.
		recordtest.At(t, seatID, 1, seatID+":regrade:R1-1", &recordpb.Regrade{Severity: recordtest.P(recordpb.Grade_GRADE_CERTAIN), Basis: proto.String("new evidence")}),
		recordtest.At(t, seatID, 1, seatID+":close:R1-2", &recordpb.Close{ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_CLOSED)}),
		// A regrade and a close of an UNKNOWN gap are ignored, not fatal.
		recordtest.At(t, seatID, 1, seatID+":regrade:R9-9", &recordpb.Regrade{GapId: proto.String("R9-9"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, seatID, 1, seatID+":close:R9-9", &recordpb.Close{GapId: proto.String("R9-9")}),
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
	writeShard(t, runDir, lens, "aaaaaaaa", []*Event{
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
	if err := validate(t.TempDir(), "red-merge-r1", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}); err != nil {
		t.Errorf("absent grades were refused: %v", err)
	}
}

func TestValidateVerbContracts(t *testing.T) {
	cases := []struct {
		name    string
		typ     recordpb.EventType
		p       proto.Message
		wantErr string // empty means it must be ACCEPTED
	}{
		{"mint without --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Class: proto.String("x")}, "mint requires --check"},
		{"mint with an empty --check", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{Class: proto.String("x"), AcceptanceCheck: proto.String("")}, "mint requires --check"},
		{"mint without --class", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT)}, "mint requires --class"},
		{"mint complete", recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}, ""},

		{"close without --id", recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{}, "close requires --id"},
		{"regrade without --basis", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{}, "regrade requires --reason"},
		{"regrade complete", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{Basis: proto.String("b")}, ""},

		{"retire without --claim", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Reason: proto.String("r")}, "retire requires --claim"},
		{"retire without --reason", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c")}, "retire requires --reason"},
		{"retire complete", recordpb.EventType_EVENT_TYPE_RETIRE, &recordpb.Retire{Claim: proto.String("c"), Reason: proto.String("r")}, ""},

		{"line-of-inquiry with an unknown status", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "shelved").Set("line", "l"), "line-of-inquiry requires --as"},
		{"line-of-inquiry with no status at all", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("line", "l"), "line-of-inquiry requires --as"},
		{"line-of-inquiry with no id", "line-of-inquiry", NewPayload().Set("status", "pursued").Set("line", "l"), "line-of-inquiry requires an id"},
		{"a deferred line of inquiry needs a reason", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "deferred").Set("line", "l"), "requires --reason"},
		{"line-of-inquiry without --line", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued"), "line-of-inquiry requires --line"},
		{"a declined line of inquiry needs a reason", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "declined").Set("line", "l"), "requires --reason"},
		{"an abandoned line of inquiry needs a reason", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "abandoned").Set("line", "l"), "requires --reason"},
		{"a PURSUED line of inquiry does not need a reason", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "pursued").Set("line", "l"), ""},
		{"a declined line of inquiry with a reason", "line-of-inquiry", NewPayload().Set("inquiry_id", "Q1").Set("status", "declined").Set("line", "l").Set("reason", "why"), ""},

		// The message must name the flag the PARSER accepts. It named --gap-id for as
		// long as that flag existed and kept naming it after the rename, because the
		// spelling was derived from the payload key rather than stated.
		{"opinion missing every field", recordpb.EventType_EVENT_TYPE_OPINION, &recordpb.Opinion{}, "opinion requires --id"},
		{"an unknown verb is not validated here", "no-such-verb", NewPayload(), ""},
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
	if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}); err != nil {
		t.Fatal(err)
	}
	id, err := MintGapID(runDir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: RoundOf("red-merge-r1")}, &recordpb.Mint{GapId: proto.String(id), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}); err != nil {
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
	p.Set("reason", "the ruling's reasoning")
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
	q.Set("reason", "the ruling's reasoning")
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
	writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1")}),
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
	writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1")}),
		recordtest.At(t, seatID, 2, seatID+":mint:R2-1", &recordpb.Mint{GapId: proto.String("R2-1")}),
	})
	anchored := func() *Payload {
		return NewPayload().Set("gap_id", "R1-1").
			Set("anchor_seat", "L1").Set("anchor_tool", "git show").Set("anchor_target", "7bc501e:path").
			Set("reason", "verified against the ref; the claim now holds")
	}
	cases := []struct {
		name    string
		p       proto.Message
		wantErr string
	}{
		{"unknown gap", NewPayload().Set("gap_id", "R9-9"), "close of unknown gap"},
		{"no anchor at all", NewPayload().Set("gap_id", "R1-1"), "requires the verification triple"},
		{"a PARTIAL anchor is not an anchor", NewPayload().Set("gap_id", "R1-1").Set("anchor_seat", "L1"), "requires the verification triple"},
		{"anchor missing its target", NewPayload().Set("gap_id", "R1-1").Set("anchor_seat", "L1").Set("anchor_tool", "git show"), "requires the verification triple"},
		{"a full anchor", anchored(), ""},
		// --carried-from remains the honest alternative to re-verifying, but it is a
		// CLAIM ABOUT THE RECORD and is now checked like one. This gap has never been
		// closed, so "carried from round 2" is simply false — and accepting it was a
		// laundering path: an unanchored first closure that scores as closed, which is
		// exactly what anchored_closures_pct exists to detect. The genuine case (close
		// with an anchor, then restate) is covered in required_test.go.
		{"--carried-from cannot invent an earlier closure", NewPayload().Set("gap_id", "R1-1").Set("carried_from", "2"), "no closure of it exists in the record"},
		{"closed_with_regression needs a successor", anchored().Set("closure_class", "closed_with_regression"), "requires --superseded-by"},
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
		if err := validate(t.TempDir(), "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("anything-at-all")})); err != nil {
			t.Errorf("advisory mode refused a class: %v", err)
		}
	})

	// AN UNPARSEABLE REGISTRY IS REFUSED. It used to degrade to advisory — accepting every class
	// slug for the whole run, silently — and the reason recorded for that was "as in the oracle".
	// The oracle is the JS engine retired in #121, so the justification outlived itself while the
	// behaviour stayed: a corrupt file turning off the gate that keeps the class vocabulary
	// honest, with nothing saying so.
	//
	// ABSENT stays advisory (the case above): a run with no registry has nothing to validate
	// against. PRESENT-BUT-BROKEN is different in kind — somebody staged it, so somebody meant it
	// to bind.
	t.Run("an unparseable registry is refused, not waved through", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, "{not json")
		err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("anything-at-all")}))
		if err == nil {
			t.Fatal("a corrupt registry accepted an arbitrary class — every --class passes while it stays that way, and the run reads as validated")
		}
		if !strings.Contains(err.Error(), "unreadable") {
			t.Errorf("the refusal must name the REGISTRY as the problem, or a seat re-reads its own --class: %v", err)
		}
	})

	t.Run("a known slug passes", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		if err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("scope-creep")})); err != nil {
			t.Errorf("a registry slug was refused: %v", err)
		}
	})

	t.Run("an unknown slug is refused with a hint", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("invented")}))
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
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		for _, missing := range []string{"definition", "neighbor", "distinguisher"} {
			p := NewPayload().Set("slug", "brand-new")
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
			err := validate(runDir, "red-merge-r1", "class-new", p)
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
		runDir := t.TempDir()
		writeRegistry(t, runDir, registry)
		p := NewPayload().Set("slug", "brand-new").
			Set("definition", "d").Set("neighbor", "not-a-class").Set("distinguisher", "q")
		err := validate(runDir, "red-merge-r1", "class-new", p)
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
		writeShard(t, runDir, seatID, "aaaaaaaa", []*Event{
			recordtest.At(t, seatID, 1, seatID+":class-new:x", &recordpb.ClassNew{Slug: proto.String("run-local-class")}),
		})
		if err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("run-local-class")})); err != nil {
			t.Errorf("a class minted in this run was refused: %v", err)
		}
		// And it is a valid neighbor for a further new class.
		p := NewPayload().Set("slug", "another").
			Set("definition", "d").Set("neighbor", "run-local-class").Set("distinguisher", "q")
		if err := validate(runDir, "red-merge-r1", "class-new", p); err != nil {
			t.Errorf("a run-local class was not a valid neighbor: %v", err)
		}
	})

	t.Run("a registry with fewer than six slugs does not slice out of range", func(t *testing.T) {
		runDir := t.TempDir()
		writeRegistry(t, runDir, `{"classes":[{"slug":"only-one"}]}`)
		err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("invented")}))
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
		if err := validate(runDir, "red-merge-r1", "mint", mint(&recordpb.Mint{Class: proto.String("invented")})); err == nil {
			t.Error("an empty registry accepted an invented class")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	cases := []struct {
		name   string
		typ    recordpb.EventType
		p      proto.Message
		prior  []*Event
		want   string
		seatID string
	}{
		{name: "singleton verbs key on seat+verb", typ: recordpb.EventType_EVENT_TYPE_POSITION, p: &recordpb.Position{}, seatID: "red-merge-r1", want: "red-merge-r1:position"},
		{name: "verdict is a singleton", typ: recordpb.EventType_EVENT_TYPE_VERDICT, p: &recordpb.RoundVerdict{}, seatID: "red-merge-r1", want: "red-merge-r1:verdict"},
		{name: "gap_id is the first label consulted", typ: recordpb.EventType_EVENT_TYPE_CLOSE, p: &recordpb.Close{GapId: proto.String("R1-1")}, seatID: "red-merge-r1", want: "red-merge-r1:close:R1-1"},
		{name: "label when there is no gap_id", typ: recordpb.EventType_EVENT_TYPE_FINDING, p: &recordpb.Finding{Label: proto.String("F1")}, seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:F1"},
		{name: "id, observation, anchor and url are also labels", typ: recordpb.EventType_EVENT_TYPE_CITE, p: &recordpb.Cite{Url: proto.String("https://x")}, seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:cite:https://x"},
		{
			name: "with no label at all, a per-shard ordinal", typ: recordpb.EventType_EVENT_TYPE_FRICTION, p: &recordpb.Friction{}, seatID: "blue-lane-1",
			prior: []*Event{recordtest.Event(t, "", 0, &recordpb.Friction{}), recordtest.Event(t, "", 0, &recordpb.Friction{}), recordtest.Event(t, "", 0, &recordpb.Position{})},
			want:  "blue-lane-1:friction:#3",
		},
		{
			name: "the ordinal counts only the SAME verb", typ: recordpb.EventType_EVENT_TYPE_FRICTION, p: &recordpb.Friction{}, seatID: "blue-lane-1",
			prior: []*Event{recordtest.Event(t, "", 0, &recordpb.Position{}), recordtest.Event(t, "", 0, &recordpb.Finding{})},
			want:  "blue-lane-1:friction:#1",
		},
		{name: "an empty label falls through to the ordinal", typ: recordpb.EventType_EVENT_TYPE_FINDING, p: &recordpb.Finding{Label: proto.String("")}, seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:#1"},
		{name: "a non-string label falls through", typ: recordpb.EventType_EVENT_TYPE_FINDING, p: &recordpb.Finding{Label: proto.String(true)}, seatID: "red-lens-r1-L1", want: "red-lens-r1-L1:finding:#1"},
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

// #394: MULTI-NONCE IS NORMAL; DISCARDING WORK IS NOT, AND THE TWO LOOKED IDENTICAL.
//
// A register rotates the nonce so a crash re-dispatch writes a fresh shard (8/50 in run 5). The
// retry rewrites the same idempotency keys, so nothing is lost. But when one seat id was used for
// two different SITTINGS, the losing shard held work that survives nowhere — and the replay said
// "multi-nonce seat X: 2 dispatches" for both cases, in the same words, with nothing gating on it.
func TestMultiNonceSeparatesACrashRetryFromLostWork(t *testing.T) {
	shard := func(t *testing.T, dir, seatID, nonce string, evs []Event) {
		t.Helper()
		var lines []string
		for _, e := range evs {
			b, err := MarshalEvent(e)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, string(b))
		}
		p := filepath.Join(dir, "records", "events-"+seatID+"-"+nonce+".jsonl")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	reg := func(seat, nonce string) Event {
		return recordtest.Stamped(recordtest.At(t, seat, 0, seat+":register:"+nonce, &recordpb.Register{}), "2026-01-01T00:00:00Z")
	}

	t.Run("a crash re-dispatch loses nothing", func(t *testing.T) {
		dir := t.TempDir()
		// Both shards carry the SAME work keys; only the register nonce differs. That is what a
		// retry looks like, and it must stay silent — otherwise the signal fires on the one case
		// it exists to permit.
		work := recordtest.Stamped(recordtest.At(t, "red-merge-r1", "", 0, 0, "red-merge-r1:mint:R1-1", &recordpb.Mint{}), "2026-01-01T00:00:01Z")
		a, b := work, work
		a.Nonce, b.Nonce = "aaaaaaaa", "bbbbbbbb"
		shard(t, dir, "red-merge-r1", "aaaaaaaa", []*Event{reg("red-merge-r1", "aaaaaaaa"), a})
		shard(t, dir, "red-merge-r1", "bbbbbbbb", []*Event{reg("red-merge-r1", "bbbbbbbb"), b})

		m, err := MergedEvents(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Anomalies) != 1 {
			t.Fatalf("multi-nonce is still reported: %v", m.Anomalies)
		}
		if len(m.Discarded) != 0 {
			t.Errorf("a retry rewrites the same keys — nothing is lost: %+v", m.Discarded)
		}
		if strings.Contains(m.Anomalies[0], "DISCARDED") {
			t.Errorf("a healthy retry must not be described as a loss: %q", m.Anomalies[0])
		}
	})

	t.Run("two sittings under one id lose the earlier one", func(t *testing.T) {
		dir := t.TempDir()
		// The #394 shape: distinct work under one seat id, and no terminal event, so selection
		// falls to mtime and the earlier sitting's rulings vanish.
		first := recordtest.Stamped(recordtest.At(t, "judge-petition", 0, "judge-petition:motion-rule:M1", &recordpb.MotionRule{}), "2026-01-01T00:00:01Z")
		second := recordtest.Stamped(recordtest.At(t, "judge-petition", 0, "judge-petition:motion-rule:M2", &recordpb.MotionRule{}), "2026-01-01T00:00:02Z")
		shard(t, dir, "judge-petition", "aaaaaaaa", []*Event{reg("judge-petition", "aaaaaaaa"), first})
		shard(t, dir, "judge-petition", "bbbbbbbb", []*Event{reg("judge-petition", "bbbbbbbb"), second})

		m, err := MergedEvents(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Discarded) != 1 {
			t.Fatalf("the earlier sitting's ruling is gone and must be reported: %+v", m.Discarded)
		}
		d := m.Discarded[0]
		if d.SeatID != "judge-petition" || d.Dispatches != 2 {
			t.Errorf("the loss names its seat and dispatch count: %+v", d)
		}
		if len(d.Keys) != 1 {
			t.Fatalf("exactly the one ruling is lost — the register is excluded: %v", d.Keys)
		}
		for _, k := range d.Keys {
			if strings.Contains(k, ":register:") {
				t.Errorf("a register key must never be counted as lost work: %q", k)
			}
		}
		if !strings.Contains(m.Anomalies[0], "DISCARDED") {
			t.Errorf("the anomaly line says so too, for a human: %q", m.Anomalies[0])
		}
	})
}
