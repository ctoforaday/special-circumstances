package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These table tests carry the coverage of the deleted tests/simulator/{run-scripts,
// record-readers}.test.mjs. The JS SKIP-without-`--bin` cases have no Go analogue (Go always
// holds the record via BoardState); the record-backed audits are tested here with the projected
// counts the orchestrator feeds them (redRounds/blueBlocks/onRecord), which the differential
// pins row-for-row against the JS's spawned views.

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fixtureRun builds a runDir with the file-backed artifacts the file audits read.
func fixtureRun(t *testing.T, ledgerLines, archiveBlocks int) string {
	t.Helper()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "trajectories", "board-telemetry.jsonl"), `{"round":1,"mass":4}`+"\n"+`{"round":2,"mass":4}`+"\n")
	var lb strings.Builder
	lb.WriteString("# ledger\n## closure index\n")
	for i := 0; i < ledgerLines; i++ {
		lb.WriteString("R1-" + itoa(i+1) + " | closed | fixed | -\n")
	}
	write(t, filepath.Join(dir, "red", "ledger.md"), lb.String())
	var ab strings.Builder
	ab.WriteString("# archive\n")
	for i := 0; i < archiveBlocks; i++ {
		ab.WriteString("## R1-" + itoa(i+1) + " — closed\nprose\n")
	}
	write(t, filepath.Join(dir, "red", "archive.md"), ab.String())
	write(t, filepath.Join(dir, "blue", "CHANGELOG.md"), "## Round 1\nedits\n## Round 2\nedits\n")
	write(t, filepath.Join(dir, "trajectories", "journal.jsonl"),
		`{"type":"result","result":{"ledger_closure_lines":`+itoa(ledgerLines)+`,"archive_blocks":`+itoa(archiveBlocks)+`,"friction":["red-merge-r1: needed a PDF extractor for X"]}}`+"\n")
	return dir
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		return "-" + string(b)
	}
	return string(b)
}

func TestTelemetryAudit(t *testing.T) {
	dir := fixtureRun(t, 2, 2)
	if got := TelemetryAudit(dir, 2).Verdict; got != "PASS" {
		t.Errorf("2 telemetry rounds cover 2 red rounds: want PASS, got %s", got)
	}
	// One telemetry round, three red rounds on the record → FAIL.
	one := t.TempDir()
	write(t, filepath.Join(one, "trajectories", "board-telemetry.jsonl"), `{"round":1}`+"\n")
	if got := TelemetryAudit(one, 3).Verdict; got != "FAIL" {
		t.Errorf("1 telemetry round vs 3 red: want FAIL, got %s", got)
	}
	// Absent telemetry with red rounds on record → FAIL; with none → SKIP.
	noFile := t.TempDir()
	if got := TelemetryAudit(noFile, 2).Verdict; got != "FAIL" {
		t.Errorf("absent telemetry with red rounds: want FAIL, got %s", got)
	}
	if got := TelemetryAudit(noFile, 0).Verdict; got != "SKIP" {
		t.Errorf("absent telemetry, no red rounds: want SKIP, got %s", got)
	}
}

func TestShardAudit(t *testing.T) {
	ok := fixtureRun(t, 2, 2)
	results, _ := ReadJournal(filepath.Join(ok, "trajectories"))
	if got := ShardAudit(ok, results); got.Verdict != "PASS" {
		t.Errorf("consistent shard self-report: want PASS, got %s (%s)", got.Verdict, got.Detail)
	}
	// Envelope claims 2/2 but files hold 3 index lines → divergence FAILs.
	lying := fixtureRun(t, 3, 2)
	write(t, filepath.Join(lying, "trajectories", "journal.jsonl"),
		`{"type":"result","result":{"ledger_closure_lines":2,"archive_blocks":2,"friction":[]}}`+"\n")
	lr, _ := ReadJournal(filepath.Join(lying, "trajectories"))
	got := ShardAudit(lying, lr)
	if got.Verdict != "FAIL" {
		t.Errorf("divergent self-report: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "measured (heuristic)") {
		t.Errorf("detail should be labeled heuristic: %s", got.Detail)
	}
	// Pre-sharding run (no ledger/archive) → SKIP.
	bare := t.TempDir()
	if got := ShardAudit(bare, nil).Verdict; got != "SKIP" {
		t.Errorf("no ledger/archive: want SKIP, got %s", got)
	}
}

func TestFrictionAudit(t *testing.T) {
	env := []string{"red-merge-r1: needed a PDF extractor for X"}
	if got := FrictionAudit(env, env).Verdict; got != "PASS" {
		t.Errorf("friction on record: want PASS, got %s", got)
	}
	got := FrictionAudit(env, []string{})
	if got.Verdict != "FAIL" {
		t.Errorf("friction the record never got: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "needed a PDF extractor") {
		t.Errorf("the missing entry must be named: %s", got.Detail)
	}
	if !strings.Contains(got.Detail, "never reached the record") {
		t.Errorf("the finding must state the pipeline gap: %s", got.Detail)
	}
}

func TestContextUse(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "agent-big.jsonl"),
		`{"message":{"role":"assistant","model":"claude-haiku-4-5","usage":{"input_tokens":1000,"cache_read_input_tokens":150000,"cache_creation_input_tokens":0,"output_tokens":10}}}`+"\n")
	write(t, filepath.Join(dir, "agent-small.jsonl"),
		`{"message":{"role":"assistant","model":"claude-fable-5","usage":{"input_tokens":1000,"cache_read_input_tokens":150000,"cache_creation_input_tokens":0,"output_tokens":10}}}`+"\n")
	got := ContextUse(dir, []string{"agent-big.jsonl", "agent-small.jsonl"})
	if got.Verdict != "WARN" {
		t.Errorf("haiku over 50%% of a 200k window: want WARN, got %s (%s)", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "big") || !strings.Contains(got.Detail, "50% tripwire") {
		t.Errorf("the breaching seat must be named: %s", got.Detail)
	}
	if got := ContextUse(dir, []string{"agent-small.jsonl"}).Verdict; got != "PASS" {
		t.Errorf("same tokens on a 1M window is 15%%: want PASS, got %s", got)
	}
	// No usage records → SKIP.
	empty := t.TempDir()
	write(t, filepath.Join(empty, "agent-none.jsonl"), `{"message":{"role":"assistant"}}`+"\n")
	if got := ContextUse(empty, []string{"agent-none.jsonl"}).Verdict; got != "SKIP" {
		t.Errorf("no usage records: want SKIP, got %s", got)
	}
}

func TestAssemblyScreen(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "red", "citation-ledger.md"),
		`"#32191 open / leaf-checked OPEN" | live fetch | LOW — REFUTED: Closed as duplicate | r1`+"\n"+`"solid claim" | source | high | r1`+"\n")
	write(t, filepath.Join(dir, "report.md"),
		"# assembled report\nThe open MCP-headless bug trio (#76239, #68375, #32191) blocks headless runs.\n\n## Blue team report (in full)\naudited body text\n")
	got := AssemblyScreen(dir)
	if got.Verdict != "WARN" {
		t.Errorf("REFUTED token in assembly-owned text: want WARN, got %s (%s)", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "#32191") {
		t.Errorf("the candidate regression token must be named: %s", got.Detail)
	}
	// Token only below the cut line (in the audited body) → PASS.
	write(t, filepath.Join(dir, "report.md"),
		"# assembled report\nTwo open bugs (#76239).\n\n## Blue team report (in full)\n#32191 appears only in the audited body.\n")
	if got := AssemblyScreen(dir).Verdict; got != "PASS" {
		t.Errorf("token below the cut is not screened: want PASS, got %s", got)
	}
	// Absent report/ledger → SKIP.
	bare := t.TempDir()
	if got := AssemblyScreen(bare).Verdict; got != "SKIP" {
		t.Errorf("absent report.md/citation-ledger.md: want SKIP, got %s", got)
	}
}

func TestRecordParityAudit(t *testing.T) {
	dir := fixtureRun(t, 2, 2) // CHANGELOG has 2 rounds
	if got := RecordParityAudit(dir, 2, 2).Verdict; got != "PASS" {
		t.Errorf("2 red, 2 blue, 2 CHANGELOG: want PASS, got %s", got)
	}
	got := RecordParityAudit(dir, 3, 1)
	if got.Verdict != "FAIL" {
		t.Errorf("3 red, 1 blue is below the redRounds-1 floor: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "3 red round(s)") {
		t.Errorf("detail should carry the red-round count: %s", got.Detail)
	}
	// PASS exit: 2 red, 1 blue (blue never took the final turn), CHANGELOG 1 round → floored PASS.
	passExit := fixtureRun(t, 2, 2)
	write(t, filepath.Join(passExit, "blue", "CHANGELOG.md"), "## Round 1\nedits\n")
	if got := RecordParityAudit(passExit, 2, 1).Verdict; got != "PASS" {
		t.Errorf("a PASS exit is floored to redRounds-1: want PASS, got %s", got)
	}
	// No red rounds → SKIP.
	if got := RecordParityAudit(dir, 0, 0).Verdict; got != "SKIP" {
		t.Errorf("no red rounds: want SKIP, got %s", got)
	}
}

func writeEvents(t *testing.T, runDir, seatID string, events []string, nonce string) {
	t.Helper()
	dir := filepath.Join(runDir, "records")
	write(t, filepath.Join(dir, "events-"+seatID+"-"+nonce+".jsonl"), strings.Join(events, "\n")+"\n")
}

func TestRecordJoinAudit(t *testing.T) {
	run := t.TempDir()
	tr := t.TempDir()
	writeEvents(t, run, "red-lens-r1-L1",
		[]string{`{"seq":0,"seatId":"red-lens-r1-L1","type":"finding","key":"red-lens-r1-L1:finding:L1-F1","payload":{"label":"L1-F1"}}`}, "aabbccdd")
	// Binary role-subcommand invocation shape.
	write(t, filepath.Join(tr, "agent-aaa.jsonl"),
		`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"feov-record lens register --run x --seat-id red-lens-r1-L1"}}]}}`+"\n"+
			`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"feov-record lens finding --key F1 --reason \"x\" --run x --seat-id red-lens-r1-L1"}}]}}`+"\n")
	if got := RecordJoinAudit(run, tr, []string{"agent-aaa.jsonl"}); got.Verdict != "PASS" {
		t.Errorf("binary invocation shape matched: want PASS, got %s (%s)", got.Verdict, got.Detail)
	}
	// mjs (pre-port) shape.
	run2, tr2 := t.TempDir(), t.TempDir()
	writeEvents(t, run2, "red-lens-r1-L1",
		[]string{`{"seq":0,"seatId":"red-lens-r1-L1","type":"finding","key":"k","payload":{}}`}, "aabbccdd")
	write(t, filepath.Join(tr2, "agent-old.jsonl"),
		`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"node tools/red-lens.mjs finding --key F1 --run x --seat-id red-lens-r1-L1"}}]}}`+"\n")
	if got := RecordJoinAudit(run2, tr2, []string{"agent-old.jsonl"}).Verdict; got != "PASS" {
		t.Errorf("mjs invocation shape matched: want PASS, got %s", got)
	}
	// Quoted binary path shape.
	run3, tr3 := t.TempDir(), t.TempDir()
	writeEvents(t, run3, "red-lens-r1-L1",
		[]string{`{"seq":0,"seatId":"red-lens-r1-L1","type":"finding","key":"k","payload":{}}`}, "aabbccdd")
	write(t, filepath.Join(tr3, "agent-q.jsonl"),
		`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Bash","input":{"command":"\"/plug/bin/feov-record\" lens finding --key F1 --run x --seat-id red-lens-r1-L1"}}]}}`+"\n")
	if got := RecordJoinAudit(run3, tr3, []string{"agent-q.jsonl"}).Verdict; got != "PASS" {
		t.Errorf("quoted binary path matched: want PASS, got %s", got)
	}
	// Orphan event no transcript invoked → FAIL, named.
	writeEvents(t, run, "red-merge-r1",
		[]string{`{"seq":0,"seatId":"red-merge-r1","type":"friction","key":"red-merge-r1:friction:x","payload":{"text":"nobody ran this"}}`}, "ddeeff00")
	got := RecordJoinAudit(run, tr, []string{"agent-aaa.jsonl"})
	if got.Verdict != "FAIL" {
		t.Errorf("orphan event: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "red-merge-r1 friction") {
		t.Errorf("the orphan event must be named: %s", got.Detail)
	}
	// No records/ → SKIP; empty records/ → SKIP (distinct detail strings).
	noRec := t.TempDir()
	if got := RecordJoinAudit(noRec, tr, nil); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "pre-record-tool run") {
		t.Errorf("no records/: want SKIP/pre-record-tool, got %s (%s)", got.Verdict, got.Detail)
	}
	emptyRec := t.TempDir()
	if err := os.MkdirAll(filepath.Join(emptyRec, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := RecordJoinAudit(emptyRec, tr, nil); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "empty") {
		t.Errorf("empty records/: want SKIP/empty, got %s (%s)", got.Verdict, got.Detail)
	}
}

func TestAttestationAudit(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	write(t, filepath.Join(run, "red", "archive.md"),
		"# archive\n\n## R1-1 — closed\nthe problem\nverification anchor: L1 | Read | report.md#SectionTwo\n"+
			"\n## R1-2 — closed\nother\nverification anchor: L5 | git show | 7bc501e:plans/record-tool.md\n")
	write(t, filepath.Join(tr, "agent-a.jsonl"),
		`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","input":{"file_path":"report.md#SectionTwo"}}]}}`+"\n")
	got := AttestationAudit(run, tr, []string{"agent-a.jsonl"}, 5)
	if got.Verdict != "FAIL" {
		t.Errorf("a claimed act with no matching tool call: want FAIL, got %s", got.Verdict)
	}
	if len(got.Unreconciled) != 1 || got.Unreconciled[0].ID != "R1-2" {
		t.Errorf("the unsupported claim R1-2 must be the one named: %+v", got.Unreconciled)
	}
	if !strings.Contains(got.Detail, "RECORD IS HONEST, never on the merits") {
		t.Errorf("the separation must be stated in the finding: %s", got.Detail)
	}
	// CARRIED closure asserts no fresh act → SKIP.
	write(t, filepath.Join(run, "red", "archive.md"),
		"# archive\n\n## R2-1 — closed\np\nverification anchor: CARRIED from round 1\n")
	if got := AttestationAudit(run, tr, []string{"agent-a.jsonl"}, 5).Verdict; got != "SKIP" {
		t.Errorf("all-carried closures: want SKIP, got %s", got)
	}
}

func TestModelTierAudit(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	// A judgment seat (red-merge) running on haiku while configured for opus → dearer? No: haiku
	// cheaper than opus → WARN. Configure judgment=haiku, seat on fable → dearer → FAIL.
	write(t, filepath.Join(run, "inputs", "run-config.json"), `{"model":"claude-haiku-4-5","judgmentModel":"claude-haiku-4-5"}`)
	write(t, filepath.Join(tr, "agent-m.jsonl"),
		`{"message":{"role":"assistant","model":"claude-fable-5","usage":{"input_tokens":10,"output_tokens":5}}}`+"\n"+
			`{"message":{"role":"assistant","model":"claude-fable-5"}}`+"\n")
	// The transcript head must classify to a tier-bound seat; prepend a red-merge marker.
	write(t, filepath.Join(tr, "agent-m.jsonl"),
		`{"prompt":"Red merge, round 1"}`+"\n"+
			`{"message":{"role":"assistant","model":"claude-fable-5","usage":{"input_tokens":10,"output_tokens":5}}}`+"\n")
	got := ModelTierAudit(run, tr, []string{"agent-m.jsonl"})
	if got.Verdict != "FAIL" {
		t.Errorf("judgment seat dearer than configured: want FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	// No models in run-config → SKIP.
	bare := t.TempDir()
	write(t, filepath.Join(bare, "inputs", "run-config.json"), `{}`)
	if got := ModelTierAudit(bare, tr, []string{"agent-m.jsonl"}).Verdict; got != "SKIP" {
		t.Errorf("no run-config models: want SKIP, got %s", got)
	}
}

func TestHarvestPrecedents(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "law"), 0o755); err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(repo, "research", "2026-07-18_law-test")
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	longRationale := strings.Repeat("context clause ", 45) + "Direction owed: TRAILING_ACTIONABLE_TAIL"
	if len(longRationale) <= 600 {
		t.Fatal("fixture must exceed the old 600-char cap")
	}
	results := []map[string]any{
		{"resolutions": []any{map[string]any{"gap_id": "R2-3", "resolution": "risk_accepted", "rationale": "complexity exceeds bounded likelihood x impact"}}},
		{"rulings": []any{map[string]any{"petitioner": "blue-respond-r2", "ruling": "granted", "opinion": "scope narrowed to shipped artifacts"}}},
		{"resolutions": []any{map[string]any{"gap_id": "R1-9", "resolution": "carried", "rationale": longRationale}}},
	}
	r := HarvestPrecedents(runDir, results, filepath.Join(repo, "law"))
	if r.Count != 3 {
		t.Fatalf("want 3 rulings harvested, got %d", r.Count)
	}
	out, err := os.ReadFile(r.Path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(out)
	if !strings.Contains(body, "[PERSUASIVE]") || strings.Contains(body, "[AFFIRMED") {
		t.Errorf("everything starts persuasive")
	}
	if !strings.Contains(body, "holding: risk_accepted") || !strings.Contains(body, "holding: granted") {
		t.Errorf("holdings carry their disposition")
	}
	if !strings.Contains(body, "facts: <reviewer: fill from the cited record") {
		t.Errorf("the harvest never invents facts")
	}
	if !strings.Contains(body, "source: 2026-07-18_law-test, R2-3") {
		t.Errorf("holdings carry their source anchors")
	}
	if !strings.Contains(body, "TRAILING_ACTIONABLE_TAIL") {
		t.Errorf("full rationale preserved — no truncation")
	}
	// No law/ dir → not written, reason names law.
	noLaw := HarvestPrecedents(runDir, results, filepath.Join(repo, "absent"))
	if noLaw.Written || !strings.Contains(noLaw.Reason, "law") {
		t.Errorf("absent law dir: want not-written with law reason, got %+v", noLaw)
	}
}

func TestWriteScorecardsAppends(t *testing.T) {
	memory := t.TempDir()
	runA := filepath.Join(t.TempDir(), "2026-07-01_first-run")
	runB := filepath.Join(t.TempDir(), "2026-07-18_second-run")
	if err := os.MkdirAll(runA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runB, 0o755); err != nil {
		t.Fatal(err)
	}
	a := WriteScorecards(runA, nil, memory, nil)
	if !a.Written || a.Chairs < 1 {
		t.Fatalf("every chair gets a card: %+v", a)
	}
	card := filepath.Join(memory, "red-scorecard.md")
	first, _ := os.ReadFile(card)
	if !strings.HasPrefix(string(first), "# red scorecard") {
		t.Errorf("a card born this run gets its header")
	}
	if !strings.Contains(string(first), "## 2026-07-01_first-run") {
		t.Errorf("the run directory name labels the series")
	}
	WriteScorecards(runB, nil, memory, nil)
	second, _ := os.ReadFile(card)
	s := string(second)
	if !strings.Contains(s, "## 2026-07-01_first-run") || !strings.Contains(s, "## 2026-07-18_second-run") {
		t.Errorf("both runs present — the series is appended, never overwritten")
	}
	if strings.Count(s, "# red scorecard") != 1 {
		t.Errorf("the header is written once, not re-stamped each run")
	}
	if strings.Index(s, "2026-07-01_first-run") > strings.Index(s, "2026-07-18_second-run") {
		t.Errorf("chronological, because appended")
	}
	// No memory dir → not written.
	r := WriteScorecards(runA, nil, filepath.Join(memory, "no-such-dir"), nil)
	if r.Written || !strings.Contains(r.Reason, "scorecards need the tracked memory dir") {
		t.Errorf("absent memory dir: want not-written, got %+v", r)
	}
}

// jsSlice must count UTF-16 code units, not bytes — JS String.slice(0,n) does. An em-dash
// (U+2014) is 3 UTF-8 bytes but 1 UTF-16 unit; a byte-based cut lands short and diverges from
// the JS this ports (the real-data differential caught it in the friction 60-char prefix match).
func TestJsSliceCountsUTF16Units(t *testing.T) {
	// 5 chars before the em-dash, then "—" (1 unit / 3 bytes), then more. slice(0,7) in JS keeps
	// "abcde—b" (7 units); a byte cut would stop 2 bytes into the em-dash's 3, dropping it.
	s := "abcde—bcdef"
	if got := jsSlice(s, 7); got != "abcde—b" {
		t.Fatalf("jsSlice(%q,7) = %q, want %q (em-dash is 1 UTF-16 unit, not 3 bytes)", s, got, "abcde—b")
	}
	// Cut landing exactly on the em-dash boundary keeps it.
	if got := jsSlice(s, 6); got != "abcde—" {
		t.Fatalf("jsSlice(%q,6) = %q, want %q", s, got, "abcde—")
	}
	// n past the end returns the whole string.
	if got := jsSlice(s, 100); got != s {
		t.Fatalf("jsSlice past end = %q, want %q", got, s)
	}
	// Pure ASCII still slices by count.
	if got := jsSlice("hello world", 5); got != "hello" {
		t.Fatalf("ascii jsSlice = %q, want %q", got, "hello")
	}
}
