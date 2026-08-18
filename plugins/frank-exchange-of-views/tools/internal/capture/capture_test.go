package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// recordWithRounds writes a minimal run whose board carries one minted gap per round, so the
// derived telemetry (view.Telemetry, which TelemetryAudit now computes) has one row per round.
func recordWithRounds(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	for r := 1; r <= n; r++ {
		seat := "red-merge-r" + itoa(r)
		// The shard filename's nonce must be exactly 8 hex chars (record's shardRe).
		nonce := "0000000" + string("0123456789abcdef"[r%16])
		e := record.Event{Seq: 0, SeatID: seat, Nonce: nonce, Round: r, Type: "mint",
			Key: seat + ":mint:R" + itoa(r) + "-1",
			Payload: record.NewPayload().Set("gap_id", "R"+itoa(r)+"-1").Set("problem", "p").
				Set("severity", "medium").Set("likelihood", "medium").Set("impact", "medium")}
		line, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl"), append(line, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

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
	// Telemetry is DERIVED from the record now, so the audit checks that the computed series
	// (one row per round with a minted gap) covers every red round.
	two := recordWithRounds(t, 2)
	if got := TelemetryAudit(two, 2).Verdict; got != "PASS" {
		t.Errorf("2 telemetry rounds cover 2 red rounds: want PASS, got %s", got)
	}
	// One telemetry round, three red rounds on the record → FAIL.
	one := recordWithRounds(t, 1)
	if got := TelemetryAudit(one, 3).Verdict; got != "FAIL" {
		t.Errorf("1 telemetry round vs 3 red: want FAIL, got %s", got)
	}
	// No board rounds with red rounds on record → FAIL (the derived series is empty); with
	// no red rounds → SKIP.
	empty := t.TempDir()
	if got := TelemetryAudit(empty, 2).Verdict; got != "FAIL" {
		t.Errorf("empty telemetry with red rounds: want FAIL, got %s", got)
	}
	if got := TelemetryAudit(empty, 0).Verdict; got != "SKIP" {
		t.Errorf("empty telemetry, no red rounds: want SKIP, got %s", got)
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

// THE SCREEN RUNS AGAIN, AND ON FIELDS.
//
// It caught a report still citing a source red had refuted, by regex-scanning the confidence
// column of red/citation-ledger.md for REFUTED|ABSENT. Red used to type its verdict there as
// prose ("LOW - REFUTED: closed as duplicate"), and real ledgers carried 5-11 such rows.
//
// Two changes killed it. The grade became a CLOSED ENUM of high|medium|low — three values that
// all mean the source SUPPORTS the claim — so there was nowhere to record that a verification
// failed; and the ledger became a rendered projection, so the file the screen read was a `setup`
// stub (46 bytes on the 2026-08-05 run). The result was PASS on every record-mode run. It was
// then made honest (SKIP, naming the gap), which stopped the false green and left the check
// checking nothing.
//
// `lens verify --as refutes|absent` plus `--anchor` is what brings it back (#296, #382), and the
// prose match does not come with it: the screen joins a source to the verifications OF THAT
// SOURCE and asks whether the assembled report still points a reader at it.
func TestAssemblyScreenFailsOnARefutedCitationStillInTheReport(t *testing.T) {
	for _, outcome := range []string{"refutes", "absent"} {
		t.Run(outcome, func(t *testing.T) {
			dir := screenRun(t, outcome, "https://example.test/refuted")
			// The assembled report still cites it.
			write(t, filepath.Join(dir, "report.md"), "A claim, sourced.[^1]\n\n[^1]: https://example.test/refuted\n")

			got := AssemblyScreen(dir)
			if got.Verdict != "FAIL" {
				t.Fatalf("verdict = %s, want FAIL — the report cites a source red found against (%s)", got.Verdict, got.Detail)
			}
			if !strings.Contains(got.Detail, "c-1") {
				t.Errorf("the detail must name WHICH citation, or the operator cannot act on it: %s", got.Detail)
			}
		})
	}
}

// AND PASSES WHEN THE REPORT DROPPED IT — a real PASS, over a real comparison.
func TestAssemblyScreenPassesWhenTheRefutedSourceIsGone(t *testing.T) {
	dir := screenRun(t, "refutes", "https://example.test/refuted")
	write(t, filepath.Join(dir, "report.md"), "A claim, now sourced elsewhere.[^1]\n\n[^1]: https://example.test/other\n")

	got := AssemblyScreen(dir)
	if got.Verdict != "PASS" {
		t.Fatalf("verdict = %s, want PASS (%s)", got.Verdict, got.Detail)
	}
	// The PASS must still SAY what it screened. "PASS" over an empty comparison is the shape
	// this check spent two releases in.
	if !strings.Contains(got.Detail, "1 found against") {
		t.Errorf("the PASS does not state what it compared: %s", got.Detail)
	}
}

// A SUPPORTED CITATION IS NOT SCREENED OUT. `weak` is thin support, not contradiction, and
// conflating them would turn a grading nuance into an assembly failure.
func TestAssemblyScreenIgnoresSupportingVerdicts(t *testing.T) {
	dir := screenRun(t, "weak", "https://example.test/thin")
	write(t, filepath.Join(dir, "report.md"), "A claim.[^1]\n\n[^1]: https://example.test/thin\n")
	if got := AssemblyScreen(dir); got.Verdict != "PASS" {
		t.Errorf("verdict = %s on a `weak` verification, want PASS (%s)", got.Verdict, got.Detail)
	}
}

// THE TWO SKIPS STAY DISTINGUISHABLE: nothing to screen, and nothing assembled yet.
func TestAssemblyScreenSkipsAreDistinct(t *testing.T) {
	bare := t.TempDir()
	if got := AssemblyScreen(bare); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "nothing to screen") {
		t.Errorf("no citations: want SKIP naming the empty set, got %s (%s)", got.Verdict, got.Detail)
	}
	unassembled := screenRun(t, "refutes", "https://example.test/refuted")
	got := AssemblyScreen(unassembled)
	if got.Verdict != "SKIP" || !strings.Contains(got.Detail, "no assembled report.md") {
		t.Errorf("pre-assembly: want SKIP naming the missing artifact, got %s (%s)", got.Verdict, got.Detail)
	}
}

// screenRun seeds a run with one blue citation and one red verification of it.
func screenRun(t *testing.T, outcome, url string) string {
	t.Helper()
	dir := t.TempDir()
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := func(seat, nonce string, seq int, typ string, p *record.Payload) {
		e := record.Event{Seq: seq, SeatID: seat, Nonce: nonce, Round: 1, Type: typ,
			Key: fmt.Sprintf("%s:%s:%d", seat, typ, seq), Payload: p}
		line, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		f := filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl")
		prior, _ := os.ReadFile(f)
		write(t, f, string(prior)+string(line)+"\n")
	}
	seed("blue-r1", "40000000", 0, "cite",
		record.NewPayload().Set("label", "c-1").Set("url", url).Set("title", "A Source"))
	seed("red-lens-r1-L1", "30000000", 0, "verify",
		record.NewPayload().Set("anchor", "c-1").Set("claim", "a claim").
			Set("outcome", outcome).Set("confidence", "high").Set("reason", "read it at the leaf"))
	return dir
}

// seedRevisions writes N blue round records as EVENTS — the source record-parity now counts.
// It used to count heading matches in blue/CHANGELOG.md, which audits the seat's typing rather
// than the record; the two disagree (the 2026-08-05 run: a 6,847-byte CHANGELOG and one
// revision event from one of three eligible seats — see #268).
func seedRevisions(t *testing.T, runDir string, rounds int) {
	t.Helper()
	recs := filepath.Join(runDir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	for r := 1; r <= rounds; r++ {
		seat := "blue-respond-r" + itoa(r)
		nonce := "2000000" + string("0123456789abcdef"[r%16])
		e := record.Event{Seq: 0, SeatID: seat, Nonce: nonce, Round: r, Type: "revision",
			Key: seat + ":revision", Payload: record.NewPayload().Set("reason", "round "+itoa(r)+" edits")}
		line, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl"), string(line)+"\n")
	}
}

func TestRecordParityAudit(t *testing.T) {
	dir := fixtureRun(t, 2, 2)
	seedRevisions(t, dir, 2)
	if got := RecordParityAudit(dir, 2, 2).Verdict; got != "PASS" {
		t.Errorf("2 red, 2 blue, 2 recorded round records: want PASS, got %s", got)
	}
	got := RecordParityAudit(dir, 3, 1)
	if got.Verdict != "FAIL" {
		t.Errorf("3 red, 1 blue is below the redRounds-1 floor: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "3 red round(s)") {
		t.Errorf("detail should carry the red-round count: %s", got.Detail)
	}
	// PASS exit: 2 red, 1 blue (blue never took the final turn), 1 round record → floored PASS.
	passExit := fixtureRun(t, 2, 2)
	seedRevisions(t, passExit, 1)
	if got := RecordParityAudit(passExit, 2, 1).Verdict; got != "PASS" {
		t.Errorf("a PASS exit is floored to redRounds-1: want PASS, got %s", got)
	}
	// THE DEFECT THE OLD SOURCE HID: a hand-written CHANGELOG present, round records absent.
	// Counting the file passed this; counting the record fails it, which is the point.
	unrecorded := fixtureRun(t, 2, 2)
	write(t, filepath.Join(unrecorded, "blue", "CHANGELOG.md"), "## Round 1"+"\n"+"edits"+"\n"+"## Round 2"+"\n"+"more"+"\n")
	if got := RecordParityAudit(unrecorded, 2, 2); got.Verdict != "FAIL" {
		t.Errorf("a CHANGELOG with no revision events must FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	// No red rounds → SKIP.
	if got := RecordParityAudit(dir, 0, 0).Verdict; got != "SKIP" {
		t.Errorf("no red rounds: want SKIP, got %s", got)
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

// The harvest reads THE RECORD. It used to read the harness journal's envelopes — a seat's own
// account of what it ruled — so an under-reported ruling was silently un-harvestable and the
// miss returned the same "0 rulings" as an honest quiet run (#413).
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
	board := &record.Board{Events: []record.Event{
		{Round: 2, Type: "opinion", SeatID: "judge-r2", Payload: record.NewPayload().
			Set("gap_id", "R2-3").Set("disposition", "risk_accepted").
			Set("reason", "complexity exceeds bounded likelihood x impact")},
		// The petition's FILER is on the motion event, not on the ruling — the ruling names
		// only the motion. Harvesting the petitioner means joining the two.
		{Round: 2, Type: "motion", SeatID: "blue-respond-r2", Payload: record.NewPayload().
			Set("motion_id", "M4").Set("subject", "petition").Set("reason", "the demand buries a hazard")},
		{Round: 2, Type: "motion-rule", SeatID: "judge-r2", Payload: record.NewPayload().
			Set("motion_id", "M4").Set("subject", "petition").Set("ruling", "granted").
			Set("reason", "scope narrowed to shipped artifacts")},
		{Round: 1, Type: "opinion", SeatID: "judge-r1", Payload: record.NewPayload().
			Set("gap_id", "R1-9").Set("disposition", "carried").Set("reason", longRationale)},
		// #361's verb. It moves no gap and has no envelope field, so it was unreachable by
		// construction — the one verb whose whole purpose is stating a holding.
		{Round: 2, Type: "declare", SeatID: "judge-r2", Payload: record.NewPayload().
			Set("holding", "verified means an act of looking, never confidence in the critique")},
		// A grade ruling is deliberately NOT harvested: promoting it without the ask it
		// answered would strip its scope. If this ever starts appearing, it was a decision.
		{Round: 1, Type: "motion-rule", SeatID: "red-merge-r1", Payload: record.NewPayload().
			Set("motion_id", "M1").Set("subject", "grade").Set("ruling", "refused").
			Set("reason", "disclosure does not lower likelihood")},
	}}

	r := HarvestPrecedents(runDir, nil, filepath.Join(repo, "law"), board)
	if r.Count != 4 {
		t.Fatalf("want 4 rulings harvested (2 opinions, 1 petition, 1 declaration), got %d", r.Count)
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
	if !strings.Contains(body, "petition by blue-respond-r2") {
		t.Errorf("the petitioner is joined from the motion event, not left blank:\n%s", body)
	}
	if !strings.Contains(body, "verified means an act of looking") {
		t.Errorf("a declared holding must reach the harvest (#361):\n%s", body)
	}
	if strings.Contains(body, "disclosure does not lower likelihood") {
		t.Errorf("a grade ruling is out of scope by decision — if that changed, change this test deliberately")
	}

	// No law/ dir → not written, reason names law.
	noLaw := HarvestPrecedents(runDir, nil, filepath.Join(repo, "absent"), board)
	if noLaw.Written || !strings.Contains(noLaw.Reason, "law") {
		t.Errorf("absent law dir: want not-written with law reason, got %+v", noLaw)
	}
}

// THE FAILURE THAT READ AS SUCCESS. Envelopes claiming rulings the record does not hold used to
// be indistinguishable from a run where nobody ruled: both printed "0 ruling(s)". The divergence
// is now stated, and carried in a FIELD so a reader compares numbers rather than parsing prose.
func TestHarvestNamesTheEnvelopeDivergence(t *testing.T) {
	runDir := filepath.Join(t.TempDir(), "2026-08-15_divergence")
	claimed := []map[string]any{
		{"resolutions": []any{map[string]any{"gap_id": "R1-1", "resolution": "closed", "reason": "fixed"}}},
		{"rulings": []any{map[string]any{"petitioner": "blue", "ruling": "denied", "reason": "no"}}},
	}

	r := HarvestPrecedents(runDir, claimed, filepath.Join(t.TempDir(), "law"), &record.Board{})
	if r.Written || r.Count != 0 {
		t.Fatalf("the record holds nothing, so nothing is promoted: %+v", r)
	}
	if r.EnvelopeClaimed != 2 {
		t.Errorf("the envelopes' claim is a field, not a sentence: want 2, got %d", r.EnvelopeClaimed)
	}
	if !strings.Contains(r.Reason, "envelopes claim 2") {
		t.Errorf("the divergence must be stated, not folded into the zero: %q", r.Reason)
	}

	// The honest quiet run: no record rulings AND no envelope claims. Silent, as it should be.
	quiet := HarvestPrecedents(runDir, nil, filepath.Join(t.TempDir(), "law"), &record.Board{})
	if quiet.Reason != "" || quiet.EnvelopeClaimed != 0 {
		t.Errorf("a genuinely quiet run must not be reported as a divergence: %+v", quiet)
	}
}

// A nil board is the harvest having no record to read at all. It must not panic and must not
// promote anything — the caller passes the same board the record-backed audits use.
func TestHarvestWithNoBoardPromotesNothing(t *testing.T) {
	r := HarvestPrecedents(filepath.Join(t.TempDir(), "2026-08-15_nil"), nil, t.TempDir(), nil)
	if r.Written || r.Count != 0 {
		t.Errorf("no board, no promotion: %+v", r)
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

func TestAppendCostToReport(t *testing.T) {
	dir := t.TempDir()
	report := filepath.Join(dir, "report.md")
	costMd := filepath.Join(dir, "cost.md")
	os.WriteFile(report, []byte("# Report\n\n## Analysis\n\nbody.\n"), 0o644)
	os.WriteFile(costMd, []byte("# Cost audit\n\n## Per seat-round\n\n| round | seat | $ |\n|---|---|---|\n| 1 | red-lens | $0.42 |\n| | **TOTAL** | **$0.42** |\n\n## Notes\n\n- cache stuff\n"), 0o644)

	msg := appendCostToReport(report, costMd)
	if msg == "" {
		t.Fatal("expected a fold-in message")
	}
	got, _ := os.ReadFile(report)
	if !strings.Contains(string(got), "## Cost\n\n## Per seat-round") {
		t.Errorf("cost table not folded under ## Cost:\n%s", got)
	}
	if !strings.Contains(string(got), "red-lens | $0.42") {
		t.Errorf("table rows missing:\n%s", got)
	}
	if strings.Contains(string(got), "## Notes") {
		t.Errorf("only the table (not Notes/tier) should fold in:\n%s", got)
	}
	// Idempotent: a second call must not double-append.
	before := string(got)
	appendCostToReport(report, costMd)
	after, _ := os.ReadFile(report)
	if string(after) != before {
		t.Error("second append changed report.md — not idempotent")
	}
	// No report.md → no-op, no panic.
	if appendCostToReport(filepath.Join(dir, "nope.md"), costMd) != "" {
		t.Error("absent report.md should be a silent no-op")
	}
}

// A STRAY IS A RECORDS TREE WITH NO RUN AROUND IT.
//
// Measured (#358): a seat resolved a relative --run against its own working directory and built a
// second blackboard beside the real one — the lane's whole draft, its own shards and locks —
// while the run it was dispatched into stayed empty. The run survived, which is why nothing
// noticed: work landing outside the run is indistinguishable from a seat that produced nothing.
func TestStrayRecordsAuditFindsShardsOutsideAnyRun(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "plugins"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The run being captured.
	runDir := filepath.Join(repo, "research", "the-run")
	write(t, filepath.Join(runDir, "inputs", "run-config.json"), `{"topic":"t"}`)
	write(t, filepath.Join(runDir, "records", "events-red-merge-r1-aaaaaaaa.jsonl"), "{}\n")

	if got := StrayRecordsAudit(repo, runDir); got.Verdict != "PASS" {
		t.Fatalf("a clean repo reported %s: %s", got.Verdict, got.Detail)
	}

	// A PAST RUN IS NOT A STRAY. Every run has records/, and reporting them would bury the
	// finding under the corpus — the discriminator is whether anything SET THE DIRECTORY UP.
	past := filepath.Join(repo, "research", "an-older-run")
	write(t, filepath.Join(past, "inputs", "run-config.json"), `{"topic":"t"}`)
	write(t, filepath.Join(past, "records", "events-blue-lane-1-bbbbbbbb.jsonl"), "{}\n")
	if got := StrayRecordsAudit(repo, runDir); got.Verdict != "PASS" {
		t.Fatalf("a past run was reported as a stray: %s", got.Detail)
	}

	// The measured shape: a records tree under tools/, from a relative path resolved there.
	write(t, filepath.Join(repo, "plugins", "tools", "research", "the-run", "records", "events-blue-lane-1-cccccccc.jsonl"), "{}\n")
	got := StrayRecordsAudit(repo, runDir)
	if got.Verdict != "FAIL" {
		t.Fatalf("a stray shard tree reported %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "tools/research/the-run/records") {
		t.Errorf("the detail must name WHERE the stray is, or an operator cannot go and look: %s", got.Detail)
	}
}

// #394: a run whose replay threw away recorded work must not report clean.
//
// The replay already reported "multi-nonce seat X: N dispatches" for BOTH a healthy crash
// re-dispatch and a seat id used for two sittings, in the same words — and nothing gated on
// either, so a run that lost a whole bench sitting finished with every audit green.
func TestDiscardedEventsAudit(t *testing.T) {
	shard := func(t *testing.T, runDir, name string, lines ...string) {
		t.Helper()
		write(t, filepath.Join(runDir, "records", name), strings.Join(lines, "\n")+"\n")
	}
	ev := func(seat, nonce, typ, key string) string {
		e := record.Event{TS: "2026-01-01T00:00:00Z", SeatID: seat, Nonce: nonce, Type: typ, Key: key, Payload: record.NewPayload()}
		b, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	t.Run("two sittings under one seat id FAIL", func(t *testing.T) {
		runDir := t.TempDir()
		shard(t, runDir, "events-judge-petition-aaaaaaaa.jsonl",
			ev("judge-petition", "aaaaaaaa", "register", "judge-petition:register:aaaaaaaa"),
			ev("judge-petition", "aaaaaaaa", "motion-rule", "judge-petition:motion-rule:M1"))
		shard(t, runDir, "events-judge-petition-bbbbbbbb.jsonl",
			ev("judge-petition", "bbbbbbbb", "register", "judge-petition:register:bbbbbbbb"),
			ev("judge-petition", "bbbbbbbb", "motion-rule", "judge-petition:motion-rule:M2"))

		a := DiscardedEventsAudit(runDir)
		if a.Verdict != "FAIL" {
			t.Fatalf("lost work must FAIL the run record: %+v", a)
		}
		if !strings.Contains(a.Detail, "judge-petition") || !strings.Contains(a.Detail, "motion-rule:M1") {
			t.Errorf("the detail must name the seat and the lost keys, so it is actionable: %q", a.Detail)
		}
	})

	t.Run("a crash re-dispatch PASSES", func(t *testing.T) {
		runDir := t.TempDir()
		// Same work key in both shards; only the register nonce differs.
		shard(t, runDir, "events-red-merge-r1-aaaaaaaa.jsonl",
			ev("red-merge-r1", "aaaaaaaa", "register", "red-merge-r1:register:aaaaaaaa"),
			ev("red-merge-r1", "aaaaaaaa", "mint", "red-merge-r1:mint:R1-1"))
		shard(t, runDir, "events-red-merge-r1-bbbbbbbb.jsonl",
			ev("red-merge-r1", "bbbbbbbb", "register", "red-merge-r1:register:bbbbbbbb"),
			ev("red-merge-r1", "bbbbbbbb", "mint", "red-merge-r1:mint:R1-1"))

		if a := DiscardedEventsAudit(runDir); a.Verdict != "PASS" {
			t.Errorf("a retry is the case this audit must permit, not punish: %+v", a)
		}
	})

	t.Run("a single-shard run PASSES", func(t *testing.T) {
		runDir := t.TempDir()
		shard(t, runDir, "events-blue-synthesize-aaaaaaaa.jsonl",
			ev("blue-synthesize", "aaaaaaaa", "register", "blue-synthesize:register:aaaaaaaa"))
		if a := DiscardedEventsAudit(runDir); a.Verdict != "PASS" {
			t.Errorf("the ordinary run must stay quiet: %+v", a)
		}
	})
}

// #270: capture is the step that CLOSES a run, and it used to be silent about the one piece of
// state that says the run is open.
//
// The old code appended a line only when it removed a marker. Absence printed nothing — so a
// capture run from a subdirectory (the path is cwd-rooted) found no marker, removed nothing, and
// reported a clean run, while the real marker stayed behind telling every later un-flagged verb
// it was still inside that run.
func TestCaptureSaysWhatHappenedToTheMarker(t *testing.T) {
	t.Run("removed", func(t *testing.T) {
		cwd := t.TempDir()
		write(t, filepath.Join(cwd, ".claude", "run-live.json"), `{"runDir":"research/x"}`)
		got := closeRunLiveMarker(cwd)
		if got != "run-live marker: removed" {
			t.Errorf("got %q", got)
		}
		if _, err := os.Stat(filepath.Join(cwd, ".claude", "run-live.json")); !os.IsNotExist(err) {
			t.Error("the marker must actually be gone, not merely reported gone")
		}
	})

	t.Run("absent is STATED, not silent", func(t *testing.T) {
		cwd := t.TempDir()
		got := closeRunLiveMarker(cwd)
		if got == "" {
			t.Fatal("silence is the defect: it reads identically to a successful removal being omitted")
		}
		if !strings.Contains(got, "none at") {
			t.Errorf("the line must say there was none: %q", got)
		}
		// The wrong-cwd case is the one that costs something, so the line has to raise it.
		if !strings.Contains(got, "not the project root") {
			t.Errorf("the line must name the reason a marker can be missed rather than absent: %q", got)
		}
		if !strings.Contains(got, cwd) {
			t.Errorf("the line must name the path it looked at, or the operator cannot check it: %q", got)
		}
	})
}
