package capture

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// recordWithRounds writes a minimal run whose board carries one minted gap per round, so the
// derived telemetry (view.Telemetry, which TelemetryAudit now computes) has one row per round.
func recordWithRounds(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	var evs []*recordpb.Event
	for r := 1; r <= n; r++ {
		seat := "red-merge-r" + itoa(r)
		// Every field the record REQUIRES, because it now refuses a mint that omits one. The
		// fixture used to name four; the other three were absent and nothing said so.
		evs = append(evs, recordtest.At(t, seat, r, seat+":mint:R"+itoa(r)+"-1", &recordpb.Mint{
			GapId:           proto.String("R" + itoa(r) + "-1"),
			Class:           proto.String("scope-creep"),
			Problem:         proto.String("p"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}))
	}
	recordtest.Seed(t, dir, evs...)
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
		`{"type":"result","result":{"ledger_closure_lines":`+itoa(ledgerLines)+`,"archive_blocks":`+itoa(archiveBlocks)+`,"log":["red-merge-r1: needed a PDF extractor for X"]}}`+"\n")
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
	if got := TelemetryAudit(runtest.Open(t, two), 2).Verdict; got != "PASS" {
		t.Errorf("2 telemetry rounds cover 2 red rounds: want PASS, got %s", got)
	}
	// One telemetry round, three red rounds on the record → FAIL.
	one := recordWithRounds(t, 1)
	if got := TelemetryAudit(runtest.Open(t, one), 3).Verdict; got != "FAIL" {
		t.Errorf("1 telemetry round vs 3 red: want FAIL, got %s", got)
	}
	// No board rounds with red rounds on record → FAIL (the derived series is empty); with
	// no red rounds → SKIP.
	empty := t.TempDir()
	if got := TelemetryAudit(runtest.Open(t, empty), 2).Verdict; got != "FAIL" {
		t.Errorf("empty telemetry with red rounds: want FAIL, got %s", got)
	}
	if got := TelemetryAudit(runtest.Open(t, empty), 0).Verdict; got != "SKIP" {
		t.Errorf("empty telemetry, no red rounds: want SKIP, got %s", got)
	}
}

// frictionRun writes a run whose seats registered (optionally binding an agent handle) and
// optionally opened the friction channel.
func frictionRun(t *testing.T, seat, agentID string, wrote string) string {
	t.Helper()
	dir := t.TempDir()
	// agent_id IS A FIELD ON THE REGISTER now, not a payload key — and it is only SET when the
	// hook supplied one, because a run whose hook never fired must stay legible as "not measured"
	// rather than as an agent whose handle is the empty string.
	reg := &recordpb.Register{ToolVersion: proto.String("test")}
	if agentID != "" {
		reg.AgentId = proto.String(agentID)
	}
	evs := []*recordpb.Event{recordtest.At(t, seat, 1, seat+":register:#1", reg)}
	switch wrote {
	case "log":
		evs = append(evs, recordtest.At(t, seat, 1, seat+":friction:#1",
			&recordpb.Log{Text: proto.String("the seat's own words, recorded"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}))
	case "friction-none":
		evs = append(evs, recordtest.At(t, seat, 1, seat+":friction_none:#1",
			&recordpb.Log{Text: proto.String("the seat's own words, recorded"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}))
	case "":
	default:
		t.Fatalf("frictionRun does not know how to write %q", wrote)
	}
	recordtest.Seed(t, dir, evs...)
	return dir
}

func recordFriction(t *testing.T, runDir string) []record.LogEntryJSON {
	t.Helper()
	b, err := record.BoardState(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	fj := record.LogJSONOf(b.Events)
	return append(append([]record.LogEntryJSON{}, fj.Log...), fj.Log...)
}

// THE CASE THAT FAILED 5 OUT OF 5 IN PRODUCTION: a seat writes its friction to the record and then
// PARAPHRASES it into its return envelope. Comparing prose to prose called that a missing record.
func TestASeatThatParaphrasesItselfStillReconciles(t *testing.T) {
	run := frictionRun(t, "blue-synthesize", "a24445d32ad697bd4", "log")
	env := []EnvelopeLog{{AgentID: "a24445d32ad697bd4",
		Text: "blue-synthesize: citation-hygiene: entirely different wording from the record"}}
	got := LogAudit(runtest.Open(t, run), env, recordFriction(t, run))
	if got.Verdict != "PASS" {
		t.Errorf("the seat opened the channel; the envelope is a re-worded copy, not a second duty.\ngot %s: %s", got.Verdict, got.Detail)
	}
}

// AND THE REAL GAP STILL FAILS: a seat that reported friction to the harness and never opened the
// channel on the record.
func TestASeatThatToldOnlyTheHarnessIsAFinding(t *testing.T) {
	run := frictionRun(t, "red-merge-r1", "a78f5dfdc4aa2ea54", "")
	env := []EnvelopeLog{{AgentID: "a78f5dfdc4aa2ea54", Text: "needed a PDF extractor for X"}}
	got := LogAudit(runtest.Open(t, run), env, recordFriction(t, run))
	if got.Verdict != "FAIL" {
		t.Fatalf("friction the record never got: want FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	for _, want := range []string{"red-merge-r1", "needed a PDF extractor"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("the finding must name %q: %s", want, got.Detail)
		}
	}
}

// `friction-none` is the attested empty case — a seat closing the channel honestly has used it.
func TestTheAttestedEmptyCaseCountsAsOpeningTheChannel(t *testing.T) {
	run := frictionRun(t, "judge-r2", "a7e42caf6c06aec62", "friction-none")
	env := []EnvelopeLog{{AgentID: "a7e42caf6c06aec62", Text: "No capability gaps encountered."}}
	if got := LogAudit(runtest.Open(t, run), env, recordFriction(t, run)); got.Verdict != "PASS" {
		t.Errorf("a filed friction-none IS the channel being used; got %s: %s", got.Verdict, got.Detail)
	}
}

// NOT JOINED IS NOT A FINDING. A run whose PreToolUse hook never fired carries no agent_id on any
// register event, deliberately — so the entry cannot be attributed, and an audit that cannot see
// something must say so rather than accuse.
func TestAnUnjoinableEntryIsReportedRatherThanBlamed(t *testing.T) {
	run := frictionRun(t, "blue-respond-r2", "", "")
	env := []EnvelopeLog{{AgentID: "", Text: "Friction channel closed: no capability gaps."}}
	got := LogAudit(runtest.Open(t, run), env, recordFriction(t, run))
	if got.Verdict == "FAIL" {
		t.Errorf("an entry with no agent binding is unmeasurable, not a duty skipped.\ngot %s: %s", got.Verdict, got.Detail)
	}
	if !strings.Contains(got.Detail, "COULD NOT BE JOINED") {
		t.Errorf("the unjoinable entries must be counted out loud, or an audit that saw nothing "+
			"reads as a clean board:\n%s", got.Detail)
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
	for _, outcome := range []recordpb.SourceOutcome{
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES,
		recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT,
	} {
		t.Run(recordpb.Word(outcome), func(t *testing.T) {
			dir := screenRun(t, outcome, "https://example.test/refuted")
			// The assembled report still cites it.
			write(t, filepath.Join(dir, "report.md"), "A claim, sourced.[^1]\n\n[^1]: https://example.test/refuted\n")

			got := AssemblyScreen(runtest.Open(t, dir))
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
	dir := screenRun(t, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, "https://example.test/refuted")
	write(t, filepath.Join(dir, "report.md"), "A claim, now sourced elsewhere.[^1]\n\n[^1]: https://example.test/other\n")

	got := AssemblyScreen(runtest.Open(t, dir))
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
	dir := screenRun(t, recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK, "https://example.test/thin")
	write(t, filepath.Join(dir, "report.md"), "A claim.[^1]\n\n[^1]: https://example.test/thin\n")
	if got := AssemblyScreen(runtest.Open(t, dir)); got.Verdict != "PASS" {
		t.Errorf("verdict = %s on a `weak` verification, want PASS (%s)", got.Verdict, got.Detail)
	}
}

// THE TWO SKIPS STAY DISTINGUISHABLE: nothing to screen, and nothing assembled yet.
func TestAssemblyScreenSkipsAreDistinct(t *testing.T) {
	bare := t.TempDir()
	if got := AssemblyScreen(runtest.Open(t, bare)); got.Verdict != "SKIP" || !strings.Contains(got.Detail, "nothing to screen") {
		t.Errorf("no citations: want SKIP naming the empty set, got %s (%s)", got.Verdict, got.Detail)
	}
	unassembled := screenRun(t, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, "https://example.test/refuted")
	got := AssemblyScreen(runtest.Open(t, unassembled))
	if got.Verdict != "SKIP" || !strings.Contains(got.Detail, "no assembled document") {
		t.Errorf("pre-assembly: want SKIP naming the missing artifact, got %s (%s)", got.Verdict, got.Detail)
	}
}

// screenRun seeds a run with one blue citation and one red verification of it.
func screenRun(t *testing.T, outcome recordpb.SourceOutcome, url string) string {
	t.Helper()
	dir := t.TempDir()
	// SEEDED INTO THE RUN'S RECORD, not hand-marshalled into a shard file. The old helper built an
	// Event literal and appended a line; the literal's shape and the file are both gone, and a
	// fixture that still wrote the file would leave this run's board EMPTY while every assertion
	// below carried on passing.
	seed := func(seat string, body proto.Message) {
		recordtest.Seed(t, dir, recordtest.Event(t, seat, 1, body))
	}
	seed("blue-r1",
		&recordpb.Cite{Label: proto.String("c-1"), Url: proto.String(url), Title: proto.String("A Source")})
	seed("red-lens-r1-L1",
		&recordpb.Verify{
			Anchor:     proto.String("c-1"),
			Claim:      proto.String("a claim"),
			Outcome:    &outcome,
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Text:       proto.String("read it at the leaf"),
		})
	return dir
}

// seedRevisions writes N blue round records as EVENTS — the source record-parity now counts.
// It used to count heading matches in blue/CHANGELOG.md, which audits the seat's typing rather
// than the record; the two disagree (the 2026-08-05 run: a 6,847-byte CHANGELOG and one
// revision event from one of three eligible seats — see #268).
func seedRevisions(t *testing.T, runDir string, rounds int) {
	t.Helper()
	var evs []*recordpb.Event
	for r := 1; r <= rounds; r++ {
		seat := "blue-respond-r" + itoa(r)
		evs = append(evs, recordtest.At(t, seat, r, seat+":revision",
			&recordpb.Revision{Text: proto.String("round " + itoa(r) + " edits")}))
	}
	recordtest.Seed(t, runDir, evs...)
}

func TestRecordParityAudit(t *testing.T) {
	dir := fixtureRun(t, 2, 2)
	seedRevisions(t, dir, 2)
	if got := RecordParityAudit(runtest.Open(t, dir), 2, 2).Verdict; got != "PASS" {
		t.Errorf("2 red, 2 blue, 2 recorded round records: want PASS, got %s", got)
	}
	got := RecordParityAudit(runtest.Open(t, dir), 3, 1)
	if got.Verdict != "FAIL" {
		t.Errorf("3 red, 1 blue is below the redRounds-1 floor: want FAIL, got %s", got.Verdict)
	}
	if !strings.Contains(got.Detail, "3 red round(s)") {
		t.Errorf("detail should carry the red-round count: %s", got.Detail)
	}
	// PASS exit: 2 red, 1 blue (blue never took the final turn), 1 round record → floored PASS.
	passExit := fixtureRun(t, 2, 2)
	seedRevisions(t, passExit, 1)
	if got := RecordParityAudit(runtest.Open(t, passExit), 2, 1).Verdict; got != "PASS" {
		t.Errorf("a PASS exit is floored to redRounds-1: want PASS, got %s", got)
	}
	// THE DEFECT THE OLD SOURCE HID: a hand-written CHANGELOG present, round records absent.
	// Counting the file passed this; counting the record fails it, which is the point.
	unrecorded := fixtureRun(t, 2, 2)
	write(t, filepath.Join(unrecorded, "blue", "CHANGELOG.md"), "## Round 1"+"\n"+"edits"+"\n"+"## Round 2"+"\n"+"more"+"\n")
	if got := RecordParityAudit(runtest.Open(t, unrecorded), 2, 2); got.Verdict != "FAIL" {
		t.Errorf("a CHANGELOG with no revision events must FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	// No red rounds → SKIP.
	if got := RecordParityAudit(runtest.Open(t, dir), 0, 0).Verdict; got != "SKIP" {
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
	got := ModelTierAudit(runtest.Open(t, run), tr, []string{"agent-m.jsonl"})
	if got.Verdict != "FAIL" {
		t.Errorf("judgment seat dearer than configured: want FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	// No models in run-config → SKIP.
	bare := t.TempDir()
	write(t, filepath.Join(bare, "inputs", "run-config.json"), `{}`)
	if got := ModelTierAudit(runtest.Open(t, bare), tr, []string{"agent-m.jsonl"}).Verdict; got != "SKIP" {
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
	board := &record.Board{Events: []*record.Event{
		recordtest.Event(t, "judge-r2", 2, &recordpb.Opinion{
			GapId:       proto.String("R2-3"),
			Disposition: recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED.Enum(),
		}),
		// The petition's FILER is on the motion event, not on the ruling — the ruling names
		// only the motion. Harvesting the petitioner means joining the two.
		recordtest.Event(t, "blue-respond-r2", 2, &recordpb.Motion{
			MotionId: proto.String("M4"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
			Basis:    proto.String("the demand buries a hazard"),
			Filing:   &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}},
		}),
		recordtest.Event(t, "judge-r2", 2, &recordpb.MotionRule{
			MotionId: proto.String("M4"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
			Opinion:  proto.String("scope narrowed to shipped artifacts"),
			Ruling:   &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_GRANTED},
		}),
		recordtest.Event(t, "judge-r1", 1, &recordpb.Opinion{Disposition: recordpb.Disposition_DISPOSITION_CARRIED.Enum(), Rationale: proto.String(longRationale)}),
		// #361's verb. It moves no gap and has no envelope field, so it was unreachable by
		// construction — the one verb whose whole purpose is stating a holding.
		recordtest.Event(t, "judge-r2", 2, &recordpb.Declare{
			Holding: proto.String("verified means an act of looking"),
		}),
		// A grade ruling is deliberately NOT harvested: promoting it without the ask it
		// answered would strip its scope. If this ever starts appearing, it was a decision.
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.MotionRule{
			MotionId: proto.String("M1"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Opinion:  proto.String("disclosure does not lower likelihood"),
			Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED},
		}),
	}}

	r := HarvestPrecedents(runtest.New(t, runDir), nil, filepath.Join(repo, "law"), board)
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
	// A DISPOSITION IS NOT A HOLDING. This used to assert `holding: defect_accepted` and
	// `holding: granted` — docket statuses in the field law/README.md reserves for "the rule
	// applied". Nine rulings harvested across two real runs all read "holding: closed", which no
	// later bench could apply to anything, and the reasoning sat unextracted in `rationale`.
	for _, want := range []string{"disposition: defect_accepted", "disposition: granted"} {
		if !strings.Contains(body, want) {
			t.Errorf("a docket ruling must state its disposition as one: missing %q", want)
		}
	}
	if !strings.Contains(body, "holding: <reviewer: state the rule this ruling applied") {
		t.Errorf("a docket ruling's holding must be a placeholder the reviewer fills — the harvest " +
			"cannot synthesise the rule without inventing it")
	}
	// ...except a declaration, whose whole purpose is to state a holding.
	if !strings.Contains(body, "holding: verified means an act of looking") {
		t.Errorf("a declaration's text IS the holding and must not be replaced by a placeholder")
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
	noLaw := HarvestPrecedents(runtest.New(t, runDir), nil, filepath.Join(repo, "absent"), board)
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
		{"resolutions": []any{map[string]any{"gap_id": "R1-1", "resolution": "repaired", "reason": "fixed"}}},
		{"rulings": []any{map[string]any{"petitioner": "blue", "ruling": "denied", "reason": "no"}}},
	}

	r := HarvestPrecedents(runtest.New(t, runDir), claimed, filepath.Join(t.TempDir(), "law"), &record.Board{})
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
	quiet := HarvestPrecedents(runtest.New(t, runDir), nil, filepath.Join(t.TempDir(), "law"), &record.Board{})
	if quiet.Reason != "" || quiet.EnvelopeClaimed != 0 {
		t.Errorf("a genuinely quiet run must not be reported as a divergence: %+v", quiet)
	}
}

// A nil board is the harvest having no record to read at all. It must not panic and must not
// promote anything — the caller passes the same board the record-backed audits use.
func TestHarvestWithNoBoardPromotesNothing(t *testing.T) {
	r := HarvestPrecedents(runtest.New(t, filepath.Join(t.TempDir(), "2026-08-15_nil")), nil, t.TempDir(), nil)
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
	a := WriteScorecards(runtest.Open(t, runA), nil, memory, nil)
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
	WriteScorecards(runtest.Open(t, runB), nil, memory, nil)
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
	r := WriteScorecards(runtest.Open(t, runA), nil, filepath.Join(memory, "no-such-dir"), nil)
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
	// THE HEADING THE SLICE BRINGS IS DEMOTED, so the fold produces ONE section with a table
	// under it. Pasted as-is, "## Cost" is a heading with nothing beneath it and the table
	// belongs to a sibling — nine bytes of section, in every archived report.
	if !strings.Contains(string(got), "## Cost\n\n### Per seat-round") {
		t.Errorf("cost table not folded under ## Cost:\n%s", got)
	}
	if strings.Contains(string(got), "## Cost\n\n## ") {
		t.Errorf("## Cost shipped as an empty heading with the table under a sibling:\n%s", got)
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

// THE DISCARDED-EVENTS TEST IS GONE WITH THE AUDIT IT COVERED.
//
// It built two shard files under one seat id and asserted the audit FAILED the run, because the
// losing shard's events survived nowhere. There are no shards: both sittings are rows in one
// table, told apart by nothing that selects a winner, so the loss it detected is unrepresentable
// rather than merely absent. capture.go records the same reasoning where the audit used to be.
//
// Deleting the test rather than pinning it to a permanent PASS is the same decision as deleting
// the audit: a green check for a condition that cannot occur reads as evidence and is not.

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
		write(t, filepath.Join(cwd, ".claude", "run-live.json"), `{"runs":[{"runDir":"research/x"}]}`)
		got := closeRunLiveMarker(cwd, "research/x")
		if got != "run-live marker: removed" {
			t.Errorf("got %q", got)
		}
		if _, err := os.Stat(filepath.Join(cwd, ".claude", "run-live.json")); !os.IsNotExist(err) {
			t.Error("the marker must actually be gone, not merely reported gone")
		}
	})

	t.Run("absent is STATED, not silent", func(t *testing.T) {
		cwd := t.TempDir()
		got := closeRunLiveMarker(cwd, "research/x")
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

// ---- liveness ----

// writeRunForLiveness lays down `n` events `gap` apart, ending at `last`. `outcome` makes the last
// one a bench `outcome`, which is what separates a finished run from a killed one — see
// LivenessAudit.
//
// SEEDED THROUGH THE STORE. This hand-wrote an `events-<seat>-<nonce>.jsonl` line at a time with
// Fprintf. There are no shards, so it produced a run whose record was EMPTY and LivenessAudit
// answered "no events in this run" for every case — a fixture failing in a way that reads like the
// audit having an opinion.
func writeRunForLiveness(t *testing.T, n int, gap time.Duration, last time.Time, outcome bool) record.Run {
	t.Helper()
	dir := t.TempDir()
	stamp := "2006-01-02T15:04:05.000000000Z07:00"
	evs := make([]*recordpb.Event, 0, n)
	for i := 0; i < n; i++ {
		ts := last.Add(-time.Duration(n-1-i) * gap).UTC().Format(stamp)
		if outcome && i == n-1 {
			evs = append(evs, recordtest.Stamped(recordtest.At(t, "judge-terminal", 1, "judge-terminal:outcome:#1", &recordpb.Outcome{
				Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING),
				Prose:   proto.String("the round ceiling arrived before red could pass the final revision"),
			}), ts))
			continue
		}
		evs = append(evs, recordtest.Stamped(recordtest.At(t, "red-lens-r1-L1", 1, fmt.Sprintf("red-lens-r1-L1:finding:k%d", i), &recordpb.Finding{
			FindingId: proto.String(fmt.Sprintf("F%d", i)),
			Label:     proto.String(fmt.Sprintf("L1-F%d", i)),
			Text:      proto.String("a finding"),
		}), ts))
	}
	recordtest.Seed(t, dir, evs...)
	return runtest.Open(t, dir)
}

// THE PAIR IS THE POINT. Both runs are equally silent; only one of them finished. This audit's
// own first draft looked at silence alone and failed BOTH of the day's real runs — a gate firing
// on the healthy case, which is the defect class it was written to catch.
func TestLivenessSeparatesFinishedFromTerminatedAtEqualSilence(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	captureAt := now.Add(41 * time.Minute)

	killed := LivenessAudit(writeRunForLiveness(t, 10, 20*time.Second, now, false), captureAt)
	if killed.Verdict != "FAIL" {
		t.Errorf("a run with no terminal outcome, silent 41m, audited %s — want FAIL:\n%s", killed.Verdict, killed.Detail)
	}
	if !strings.Contains(killed.Detail, "TERMINATED") {
		t.Errorf("the FAIL does not say the run was terminated:\n%s", killed.Detail)
	}

	finished := LivenessAudit(writeRunForLiveness(t, 10, 20*time.Second, now, true), captureAt)
	if finished.Verdict != "PASS" {
		t.Errorf("a run that recorded its outcome and was captured 41m later audited %s — a finished "+
			"run is quiet BECAUSE it finished, and silence must not convict it:\n%s", finished.Verdict, finished.Detail)
	}
}

// Captured while seats are still writing: unusual, the operator's call, and not a finding.
func TestLivenessDoesNotConvictACaptureMidRun(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	a := LivenessAudit(writeRunForLiveness(t, 10, 20*time.Second, now, false), now.Add(15*time.Second))
	if a.Verdict != "PASS" {
		t.Errorf("captured 15s after the last event audited %s — want PASS:\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "mid-run") {
		t.Errorf("the PASS does not say it was captured mid-run:\n%s", a.Detail)
	}
}

// Too thin to judge reports itself. A capture over a record it cannot age must not tell a reader
// the run finished cleanly.
func TestLivenessReportsNotMeasuredRatherThanPassing(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	a := LivenessAudit(writeRunForLiveness(t, 3, time.Minute, now, false), now.Add(3*time.Hour))
	if a.Verdict != "SKIP" {
		t.Errorf("a 3-event record audited %s — want SKIP:\n%s", a.Verdict, a.Detail)
	}
}

// ---- record archive ----

func TestArchiveRecordKeepsTheShardsAndRefusesAnEmptyRun(t *testing.T) {
	repo := t.TempDir()
	run := filepath.Join(repo, "research", "2026-08-22_x")
	recs := filepath.Join(run, "records")
	if err := os.MkdirAll(filepath.Join(run, "proofs", "abc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}

	// AN EMPTY RECORD IS REFUSED, because an archive preserving nothing is indistinguishable
	// from a preserved run once it is sitting in run-archive/.
	if _, err := ArchiveRecord(runtest.Open(t, run), repo); err == nil {
		t.Fatal("ArchiveRecord wrote an archive for a run with no shards")
	}

	if err := os.WriteFile(filepath.Join(recs, "events-red-lens-r1-L1-aaaaaaaa.jsonl"),
		[]byte(`{"seq":0,"ts":"2026-08-22T12:00:00.000000000Z","seatId":"red-lens-r1-L1","nonce":"aaaaaaaa","round":1,"role":"lens","type":"finding","key":"k","payload":{}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "proofs", "abc", "script.py"), []byte("print(7)\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ArchiveRecord(runtest.Open(t, run), repo)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(out) != "2026-08-22_x.tar.gz" {
		t.Errorf("archive named %q — it must be findable by the run slug", filepath.Base(out))
	}
	if filepath.Dir(out) != filepath.Join(repo, "run-archive") {
		t.Errorf("archive landed in %q — it must sit OUTSIDE research/, which is gitignored and so "+
			"does not survive the container", filepath.Dir(out))
	}

	// And it holds what the audits re-read.
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			break
		}
		names[h.Name] = true
	}
	for _, want := range []string{"records/events-red-lens-r1-L1-aaaaaaaa.jsonl", "proofs/abc/script.py"} {
		if !names[want] {
			t.Errorf("archive is missing %q — it holds %v", want, names)
		}
	}
	// The cache is deliberately absent: 7.3 MB for a real run, re-fetchable, and every source's
	// sha256 is on the record so its integrity stays checkable without it.
	for n := range names {
		if strings.HasPrefix(n, "cache/") {
			t.Errorf("the fetched-source cache was archived (%s) — it is re-fetchable and dwarfs the record", n)
		}
	}
}

// CAPTURING ONE RUN MUST NOT CLOSE ANOTHER, and the guarantee is now STRUCTURAL.
//
// It used to be a path comparison standing between capture and the wrong marker: the file was a
// singleton naming the one open run, capture removed it by PATH without asking which run it
// named, and a guard was added after capturing an abandoned run A nearly lifted live run B's
// marker on 2026-08-22, eleven minutes into B's first round.
//
// Per-run rows remove the class rather than guard it (#529): the row this capture owns is the
// only one it can take. The guarantee is asserted here in its stronger form — BOTH runs listed,
// one captured, the other still live afterwards — because a class that can no longer occur is
// still a class that must be shown not to occur.
func TestCapturingOneRunLeavesTheOthersOpen(t *testing.T) {
	cwd := t.TempDir()
	marker := filepath.Join(cwd, ".claude", "run-live.json")
	write(t, marker, `{"runs":[{"runDir":"research/live-one"},{"runDir":"research/the-one-being-captured"}]}`)

	got := closeRunLiveMarker(cwd, "research/the-one-being-captured")
	if !strings.Contains(got, "1 other run") {
		t.Errorf("capture did not say another run is still open: %q", got)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("the marker file must survive while another run is live")
	}
	left := runlive.ReadRunLive(cwd)
	if len(left) != 1 || left[0].RunDir != "research/live-one" {
		t.Errorf("after capturing one run the marker holds %+v, want only research/live-one", left)
	}

	// A relative/absolute spelling of the SAME run still closes it — the marker stores whatever
	// it was given, so a string compare would refuse to close the run in progress.
	write(t, marker, `{"runs":[{"runDir":"research/same"}]}`)
	if got := closeRunLiveMarker(cwd, filepath.Join(cwd, "research", "same")); !strings.Contains(got, "removed") {
		t.Errorf("the same run spelled absolutely was refused: %q", got)
	}

	// AND A FILE NAMING NOTHING IS NOT A CLEAN CLOSE. An unreadable marker looks exactly like
	// this from here, and reporting "nothing to remove" would let it pass for an absent one.
	write(t, marker, `{"runs":[]}`)
	if got := closeRunLiveMarker(cwd, "research/x"); !strings.Contains(got, "names NO open run") {
		t.Errorf("a marker present but naming nothing was reported as an ordinary miss: %q", got)
	}
}

// AN UNREADABLE SCORECARD IS LEFT ALONE, NOT REPLACED WITH A FRESH HEADER.
//
// The read error used to fall through to scorecard.ChairHeader and the write below then
// replaced the file with it, so one unreadable moment discarded every earlier run's rows.
// The series IS the cross-run memory — TestWriteScorecardsAppends above asserts it is
// "appended, never overwritten" — and this is the path that overwrote it while reporting
// success. Absence is the only reason to start from a header.
func TestAnUnreadableScorecardIsNotOverwritten(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0000 does not deny reads the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root reads through 0000, so the failure cannot be staged")
	}
	memory := t.TempDir()
	runA := filepath.Join(t.TempDir(), "2026-07-01_first-run")
	if err := os.MkdirAll(runA, 0o755); err != nil {
		t.Fatal(err)
	}
	if r := WriteScorecards(runtest.Open(t, runA), nil, memory, nil); !r.Written {
		t.Fatalf("seeding the history failed: %+v", r)
	}
	card := filepath.Join(memory, "red-scorecard.md")
	before, err := os.ReadFile(card)
	if err != nil {
		t.Fatal(err)
	}

	// Make exactly one chair's card unreadable, then run a later capture over it.
	if err := os.Chmod(card, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(card, 0o644) })
	runB := filepath.Join(t.TempDir(), "2026-07-18_second-run")
	if err := os.MkdirAll(runB, 0o755); err != nil {
		t.Fatal(err)
	}
	got := WriteScorecards(runtest.Open(t, runB), nil, memory, nil)

	if got.Written {
		t.Error("a chair that could not be written must not report Written — that is the whole defect")
	}
	for _, want := range []string{"red", "cannot read", "left untouched"} {
		if !strings.Contains(got.Reason, want) {
			t.Errorf("the reason must carry %q so the operator knows which card and why: %q", want, got.Reason)
		}
	}
	// The other chairs still got their rows: a partial failure is partial, not total.
	if got.Chairs < 1 {
		t.Errorf("the readable chairs must still be written: %+v", got)
	}

	if err := os.Chmod(card, 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(card)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("the unreadable card was rewritten; its history is gone.\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if !strings.Contains(string(after), "2026-07-01_first-run") {
		t.Error("the earlier run's series must survive an unreadable moment")
	}
}

// A WRITE THAT FAILED MUST NOT REPORT Written: true.
//
// The error was discarded and the result asserted the write had landed, so a read-only
// memory dir or a full disk produced a capture that said the chair memory had been updated
// while nothing moved — and the next run reads the stale file as the whole history.
func TestAFailedScorecardWriteIsReported(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only file does not block writes the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root writes through read-only permissions")
	}
	memory := t.TempDir()
	runA := filepath.Join(t.TempDir(), "2026-07-01_first-run")
	if err := os.MkdirAll(runA, 0o755); err != nil {
		t.Fatal(err)
	}
	if r := WriteScorecards(runtest.Open(t, runA), nil, memory, nil); !r.Written {
		t.Fatalf("seeding failed: %+v", r)
	}
	card := filepath.Join(memory, "red-scorecard.md")
	// Readable, but not writable: the read succeeds and the write is refused.
	if err := os.Chmod(card, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(card, 0o644) })

	runB := filepath.Join(t.TempDir(), "2026-07-18_second-run")
	if err := os.MkdirAll(runB, 0o755); err != nil {
		t.Fatal(err)
	}
	got := WriteScorecards(runtest.Open(t, runB), nil, memory, nil)
	if got.Written {
		t.Error("a refused write must not report Written")
	}
	for _, want := range []string{"red", "cannot write"} {
		if !strings.Contains(got.Reason, want) {
			t.Errorf("the reason must name the chair and the failure: %q (missing %q)", got.Reason, want)
		}
	}
}

// A card that has never existed is the ordinary first-run path and must still be created
// from the header — the fix above must not turn "absent" into a failure.
func TestAnAbsentScorecardIsStillCreated(t *testing.T) {
	memory := t.TempDir()
	runA := filepath.Join(t.TempDir(), "2026-07-01_first-run")
	if err := os.MkdirAll(runA, 0o755); err != nil {
		t.Fatal(err)
	}
	got := WriteScorecards(runtest.Open(t, runA), nil, memory, nil)
	if !got.Written || got.Chairs < 1 || got.Reason != "" {
		t.Fatalf("a fresh memory dir must write every chair cleanly: %+v", got)
	}
	b, err := os.ReadFile(filepath.Join(memory, "red-scorecard.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "# red scorecard") {
		t.Errorf("a card born this run gets its header: %q", b)
	}
}

// THE RECORD'S ANSWER, AND WHY THE TRANSCRIPT PASS COULD NOT GIVE IT.
//
// The 2026-08-23 incident graded WARN here and capture exited 0: opus is CHEAPER than the
// configured fable, and cheaper had been filed as "verification may be discounted". Now `register`
// records what actually answered each seat, so the audit reads a declared substitution instead of
// deducing a tier from a price — and a substitution fails in either direction.
func TestModelTierAuditFailsOnASubstitutionTheRecordDeclares(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	write(t, filepath.Join(run, "inputs", "run-config.json"),
		`{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`)
	recordtest.Seed(t, run,
		recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:register:#1", &recordpb.Register{
			ToolVersion:    proto.String("test"),
			ServedModel:    proto.String("claude-opus-4-8"),
			RequestedModel: proto.String("claude-fable-5"),
		}),
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:register:#1", &recordpb.Register{
			ToolVersion: proto.String("test"),
			ServedModel: proto.String("claude-sonnet-5"),
		}),
	)
	got := ModelTierAudit(runtest.Open(t, run), tr, nil)
	if got.Verdict != "FAIL" {
		t.Fatalf("a declared substitution must FAIL (capture exits 2 on FAIL, and this one exited 0 for real): got %s — %s", got.Verdict, got.Detail)
	}
	for _, want := range []string{"blue-lane-1", "claude-fable-5", "claude-opus-4-8", "declared", "served measured on 2 of 2"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("detail must carry %q; got:\n%s", want, got.Detail)
		}
	}
	// The judgment seat was served as configured and must not be swept up with it.
	if strings.Contains(got.Detail, "red-merge-r1") {
		t.Errorf("a seat answered by its configured tier is not a finding; got:\n%s", got.Detail)
	}
}

// A RUN WHERE NOTHING LOOKED MUST NOT READ AS A RUN THAT MATCHED. This is the plausible zero the
// whole thread is about, one level up: the audit's own summary line.
func TestModelTierAuditSaysWhenTheServedModelWasNeverMeasured(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	write(t, filepath.Join(run, "inputs", "run-config.json"),
		`{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`)
	recordtest.Seed(t, run,
		recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:register:#1",
			&recordpb.Register{ToolVersion: proto.String("test")}),
	)
	got := ModelTierAudit(runtest.Open(t, run), tr, nil)
	if !strings.Contains(got.Detail, "NOT MEASURED") {
		t.Errorf("an unmeasured run must say so, not claim its seats matched; got:\n%s", got.Detail)
	}
	if got.Verdict == "FAIL" {
		t.Errorf("not measured is not a failure either; got %s", got.Verdict)
	}
}
