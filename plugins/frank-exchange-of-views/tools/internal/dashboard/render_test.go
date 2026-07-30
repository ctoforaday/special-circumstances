package dashboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fp(f float64) *float64 { return &f }
func ip(n int) *int         { return &n }

// telemetry fixture: two rounds, with new_mint.by_severity + accepted_deltas.
func telFixture() []map[string]any {
	return []map[string]any{
		{"round": 1.0, "mass": 4.0, "open_count": 3.0, "max_severity": "high",
			"new_mint": map[string]any{"count": 3.0, "by_severity": map[string]any{"high": 2.0, "medium": 1.0}}, "accepted_deltas": []any{}},
		{"round": 2.0, "mass": 6.0, "open_count": 2.0, "max_severity": "high",
			"new_mint": map[string]any{"count": 1.0, "by_severity": map[string]any{"medium": 1.0}}, "accepted_deltas": []any{map[string]any{}}},
	}
}

func judFixture() Judiciary {
	j := Judiciary{JudgeSittings: 1, Rulings: map[string]int{"carried": 2, "closed": 1},
		ChainSpans: map[int]int{1: 2, 2: 1}, Chains: 3, MigDown: 1, MigUp: 0, MigFlat: 0, LatestVerdict: "FAIL", VerdictRound: 2}
	j.Disputes.Raised, j.Disputes.Accepted, j.Disputes.Rejected = 2, 1, 1
	return j
}

// writeScorecardShadow drops a rendered scorecard into runDir/records/render-shadow/scorecards
// so scorecardSection → ParseRenderedRows/LatestSection actually fire.
func writeScorecardShadow(t *testing.T, runDir string) {
	dir := filepath.Join(runDir, "records", "render-shadow", "scorecards")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# red scorecard\n\n## this run\n\n- `anchored_closures_pct` [benchmark] — Attestation-format invariant: **100** (target 100)\n- `finding_precision` [detector] — Certification: **2**\n"
	if err := os.WriteFile(filepath.Join(dir, "red-scorecard.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func baseModel(runDir string) Model {
	tel := telFixture()
	return Model{
		RunDir: runDir, Telemetry: tel, Latest: tel[len(tel)-1],
		Cost: 12.34, APIRounds: 20, Agents: 5,
		Friction:   Friction{Count: 2, Last: "red-merge-r1: needed a PDF extractor"},
		Shards:     Shards{LedgerExists: true, OpenRows: 2, OpenBySeverity: map[string]int{"high": 1, "medium": 1}, Findings: 9, Citations: 4, ClosureIndexRows: 3, ArchiveRecords: 3},
		BlueClaims: ip(15),
		Steps:      []Step{{"frontier", "done"}, {"blue lanes", "done"}, {"synthesis", "live"}, {"round 1", "todo"}, {"assembly", "todo"}},
		Rates: []Rate{
			{Round: 1.0, Opened: 3, Closed: 0, Open: 3.0, CloseRate: 0},
			{Round: 2.0, Opened: 1, Closed: 2, Open: 2.0, CloseRate: 50},
		},
		Judiciary: judFixture(),
		Config:    Config{Topic: "does the widget converge", Model: "sonnet", JudgmentModel: "opus", MaxRounds: "8", Lanes: "3"},
		Generated: "2025-01-16T00:00:00.000Z",
	}
}

// TERMINAL fixture: report.md present → verdict-terminal, no live ETA; a superseded live seat
// (label matches a done one) and a never-finished one drive both terminal-label branches.
func TestRenderHTMLTerminal(t *testing.T) {
	runDir := t.TempDir()
	writeScorecardShadow(t, runDir)
	m := baseModel(runDir)
	m.Terminal = true
	m.TerminalVerdict = "VERIFIED"
	m.Eta = Eta{State: "complete"}
	m.Seats = []Seat{
		{Label: "frontier", Seat: "frontier", Done: true, Result: `{"verdict":"PASS","claim_count":15,"gaps":[]}`},
		{Label: "red-merge-r1", Seat: "red-merge", Round: 1, Done: true, Result: `{"verdict":"FAIL","gaps":[1,2],"resolutions":[1]}`},
		{Label: "red-merge-r1", Seat: "red-merge", Round: 1, Done: false, StartedMs: fp(1000)}, // superseded (a done one shares the label)
		{Label: "judge-r1", Seat: "judge", Round: 1, Done: false, StartedMs: fp(1000)},         // did not finish
	}
	h := RenderHTML(m)

	must := func(sub string) {
		t.Helper()
		if !strings.Contains(h, sub) {
			t.Errorf("terminal render missing %q", sub)
		}
	}
	// #1 h1 + title + topic + run-config
	must("<title>FEOV run — " + filepath.Base(runDir))
	must("FEOV run · " + filepath.Base(runDir))
	must("does the widget converge")
	must("<b>sonnet</b>")
	must("<b>opus</b>")
	must("8 rounds")
	// #3 progress bar (per-step states)
	must(`<div class="seg done"`)
	must(`<div class="seg live"`)
	must(`<div class="seg todo"`)
	// #4 the 11 tiles (cost SCALAR, verdict terminal, no projected-remaining on complete)
	must("<b>6</b><span>board mass")
	must("<b>2</b><span>open gaps")
	must("<b>high</b><span>max severity")
	must("<b>15</b><span>blue claims")
	must("<b>9</b><span>lens findings")
	must("<b>4</b><span>citations checked")
	must("<b>VERIFIED</b><span>final verdict")
	must("<b>2</b><span>friction entries")
	must("<b>$12.34</b><span>cost so far")
	must("<b>2/5</b><span>seats done")
	if strings.Contains(h, "projected remaining") {
		t.Error("terminal run must not show the projected-remaining tile")
	}
	// #5 mass SVG
	must("<polyline points=")
	must("board mass by round")
	// #6 open/close rates (incl. sevRow mint codes; round 1 opened 3 closed 0; round 2 rate 50%)
	must("<td>50%</td>")
	must("2h 1m") // round-1 mints high:2 medium:1 → sorted high-first
	// #7 Red's board — all three counts
	must("closed gaps (closure index)</td><td>3")
	must("archived closure records</td><td>3")
	must("open gaps on the board</td><td>2")
	// #8 Judiciary
	must("judge sittings</td><td>1")
	must("carried: 2")
	must("raised 2 · accepted 1 · rejected 1")
	must("down 1 · up 0 · flat 0")
	must("3 chains")
	// #9 Friction
	must("2 friction events on the record")
	must("needed a PDF extractor")
	// #9b Seats (run complete) — both terminal-label branches
	must("Seats (run complete)")
	must("superseded — a later attempt completed")
	must("did not finish")
	// #10 scorecard section (ParseRenderedRows fired)
	must("scorecards — this run")
	must("anchored_closures_pct")
	// #11 Recent completions (summarizeResult)
	must("Recent completions")
	must("verdict FAIL · 2 gaps · 1 ruling")
}

// LIVE fixture: no report.md → ETA running + running-minutes + n× live-seat collapse.
func TestRenderHTMLLive(t *testing.T) {
	runDir := t.TempDir()
	m := baseModel(runDir)
	m.Terminal = false
	m.Eta = Eta{State: "running", LowMin: 5, HighMin: 12, PerRoundLowMin: 8, PerRoundHighMin: 15, Basis: "3 completed seat(s) in this run", Unmeasured: []string{"assembly"}}
	// generated == the clock; oldest live seat started 3 min before → "running 3 min".
	nowMs := etaNow(m)
	m.Seats = []Seat{
		{Label: "red-lens-r1", Seat: "red-lens", Round: 1, Done: false, StartedMs: fp(nowMs - 3*60000)},
		{Label: "red-lens-r1", Seat: "red-lens", Round: 1, Done: false, StartedMs: fp(nowMs - 2*60000)}, // duplicate → 2×
	}
	h := RenderHTML(m)
	must := func(sub string) {
		t.Helper()
		if !strings.Contains(h, sub) {
			t.Errorf("live render missing %q", sub)
		}
	}
	// #2 ETA paragraph + projected-remaining tile
	must("projected <b>5–12 min</b> remaining")
	must("roughly 8–15 min")
	must("No completed precedent yet for: assembly")
	must("<b>5–12m</b><span>projected remaining")
	// #9b Seats live now: n× collapse + running-minutes
	must("Seats live now")
	must("2× red-lens-r1")
	must("running 3 min")
}

func TestSummarizeResult(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"verdict":"FAIL","claim_count":15,"gaps":[1,2,3],"resolutions":[1],"citations_checked":4,"friction":["x"]}`,
			"verdict FAIL · 15 claims · 3 gaps · 1 ruling · 4 citations checked · 1 friction"},
		{`{"resolutions":[1,2]}`, "2 rulings"},
		{`not json`, "not json"},
		{`{}`, "{}"},
	}
	for _, c := range cases {
		if got := summarizeResult(c.in); got != c.want {
			t.Errorf("summarizeResult(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// BuildModel integration: a fixture run dir + transcript dir → the model fields (config-merge,
// cost, seats, board unavailable when no record).
func TestBuildModelConfigAndCost(t *testing.T) {
	runDir := t.TempDir()
	tr := t.TempDir()
	// run-config on disk; CLI overrides lanes.
	_ = os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755)
	_ = os.WriteFile(filepath.Join(runDir, "inputs", "run-config.json"),
		[]byte(`{"topic":"t","model":"haiku","judgmentModel":"haiku","maxRounds":"6","lanes":"2"}`), 0o644)
	// one agent transcript with a usage record.
	_ = os.WriteFile(filepath.Join(tr, "agent-a.jsonl"),
		[]byte(`{"message":{"model":"claude-haiku-4-5","usage":{"input_tokens":1000,"output_tokens":500}}}`+"\n"), 0o644)
	_ = os.WriteFile(filepath.Join(tr, "journal.jsonl"),
		[]byte(`{"agentId":"a","result":{"verdict":"FAIL","gaps":[]}}`+"\n"), 0o644)

	m := BuildModel(runDir, tr, Config{Lanes: "9"}, 1737000000000)
	if m.Config.Model != "haiku" || m.Config.MaxRounds != "6" {
		t.Errorf("file config not merged: %+v", m.Config)
	}
	if m.Config.Lanes != "9" {
		t.Errorf("CLI --lanes should override file: got %q", m.Config.Lanes)
	}
	if m.Cost <= 0 {
		t.Errorf("cost should be > 0, got %v", m.Cost)
	}
	if m.Agents != 1 {
		t.Errorf("agents = %d, want 1", m.Agents)
	}
	if m.Shards.LedgerExists {
		t.Error("no record → shards unavailable")
	}
	if m.Friction.Count != -1 {
		t.Errorf("no record → friction unavailable (-1), got %d", m.Friction.Count)
	}
	if m.Generated != "2025-01-16T04:00:00.000Z" {
		t.Errorf("generated should reflect injected --now, got %q", m.Generated)
	}
}
