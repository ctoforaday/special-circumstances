package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// A VERB'S OWN CHECK IS MADE AGAINST THE RECORD ITS WRITE LANDS ON (#1252). Each row runs a verb
// while another seat's act commits on that seat's own connection at the instant the verb's write
// opens its transaction (runtest.Overtaken). A check the verb made before its write has already
// answered by then, over a record that no longer holds; a check made inside the write (record.Guard)
// reads the other seat's act and answers as it would have had the act come first.

const (
	raceSentence = "Costs rose sharply across the first quarter"
	raceReport   = "# Costs\n\nThe year opened badly. " + raceSentence + ". Volume recovered later.\n"
)

// raceRun is raceReport frozen, with the chair, the evidence lens and blue seated.
func raceRun(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, raceReport)
	registerBlue(t, runDir)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	return runDir
}

// renders fails the test unless the report renders: an act replay cannot apply fails every render.
func renders(t *testing.T, runDir string) string {
	t.Helper()
	md, err := reportproj.RenderFromRecord(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("the report no longer renders — a write landed that replay cannot apply: %v", err)
	}
	return md
}

// blueEditBy is seat's recorded edit of old to new, as its write path records it.
func blueEditBy(t *testing.T, seat, key, old, new string) *recordpb.Event {
	return recordtest.At(t, seat, key, &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("the sentence overstates the rise")})
}

// findingAt is lens's finding id at loc and the Anchor that places it.
func findingAt(t *testing.T, lens, id, loc string) []*recordpb.Event {
	return []*recordpb.Event{
		recordtest.At(t, lens, lens+":finding:"+id, &recordpb.Finding{Id: proto.String(id), Location: proto.String(loc), Text: proto.String("the rise is not sourced")}),
		recordtest.At(t, lens, lens+":anchor:"+id, &recordpb.Anchor{Id: proto.String(id), Location: proto.String(loc)}),
	}
}

// V6: A FINDING'S QUOTE IS PLACED ON THE REPORT ITS ANCHOR LANDS ON. Blue's edit takes the quoted
// sentence out while the finding is written; placed against the report before the edit, the finding
// and its Anchor land on text the report no longer holds, and replay cannot place the anchor.
func TestAFindingIsPlacedOnTheReportItsAnchorLandsOn(t *testing.T) {
	runDir := raceRun(t)
	edit := blueEditBy(t, "blue-lane-1", "race:edit", raceSentence, "Costs rose only modestly in the first quarter")
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{edit}, func() error {
		_, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F", "--quote", raceSentence,
			"--reason", "the rise is not sourced", "--severity", "low", "--likelihood", "low", "--impact", "low")
		return err
	})
	md := renders(t, runDir)
	if err == nil {
		t.Fatalf("a finding quoting text an edit took out at that moment was recorded:\n%s", md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING); n != 0 {
		t.Errorf("the refused finding recorded %d finding(s)", n)
	}
}

// V6, ON A RETRY: THE ANCHOR A RECORDED ACT OWES IS PLACED ON THE REPORT IT LANDS ON. A finding on the
// record without its anchor is retried under its --key; blue's edit takes the stored location out at
// that moment, and an Anchor placed against the report before it cannot be replayed.
func TestAnOwedAnchorIsPlacedOnTheReportItLandsOn(t *testing.T) {
	runDir := raceRun(t)
	id := record.NewID("finding")
	recordtest.Seed(t, runDir, recordtest.At(t, lensSeat, lensSeat+":finding:"+id, &recordpb.Finding{
		Id: proto.String(id), FindingKey: proto.String("F"), Location: proto.String(raceSentence), Text: proto.String("the rise is not sourced")}))
	edit := blueEditBy(t, "blue-lane-1", "race:edit", raceSentence, "Costs rose only modestly in the first quarter")
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{edit}, func() error {
		_, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F", "--quote", raceSentence,
			"--reason", "the rise is not sourced", "--severity", "low", "--likelihood", "low", "--impact", "low")
		return err
	})
	md := renders(t, runDir)
	if err == nil || !strings.Contains(err.Error(), "does not place") {
		t.Fatalf("the retry placed an owed anchor at text an edit took out at that moment = %v\n%s", err, md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR); n != 0 {
		t.Errorf("the refused retry recorded %d anchor(s)", n)
	}
}

// V5: BLUE'S EDIT IS PLANNED ON THE REPORT IT LANDS ON. A lens's finding places its anchor inside the
// span blue's edit replaces, at that moment; planned against the report before it, the edit's
// replacement does not carry the anchor, and replaying the edit takes the finding's anchor out of the
// report — the loss every edit is refused for.
func TestAnEditIsPlannedOnTheReportItLandsOn(t *testing.T) {
	runDir := raceRun(t)
	id := record.NewID("finding")
	err := runtest.Overtaken(t, runDir, findingAt(t, lensSeat, id, "Costs rose sharply"), func() error {
		_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E", "--quote", raceSentence,
			"--new", "Costs rose only modestly in the first quarter", "--reason", "the source gives a modest rise")
		return err
	})
	md := renders(t, runDir)
	if !strings.Contains(md, anchor.Token(id)) {
		t.Errorf("the finding's anchor is not in the report:\n%s", md)
	}
	if err == nil || !strings.Contains(err.Error(), anchor.Token(id)) {
		t.Errorf("an edit of a span a finding's anchor landed in at that moment = %v, want refused naming the anchor\n%s", err, md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Errorf("the refused edit recorded %d edit(s)", n)
	}
}

// V7: A RETIRE IS CHECKED ON THE REPORT IT LANDS ON. Another blue seat's edit writes the retired claim
// back into the report at that moment; checked against the report before it, the retire explains the
// removal of a claim the report holds.
func TestARetireIsCheckedOnTheReportItLandsOn(t *testing.T) {
	runDir := raceRun(t)
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E", "--quote", "The year opened badly. "+raceSentence+".",
		"--new", "The year opened badly.", "--reason", "the rise is not sourced"); err != nil {
		t.Fatal(err)
	}
	back := blueEditBy(t, "blue-lane-1", "race:edit", "Volume recovered later", raceSentence+". Volume recovered later")
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{back}, func() error {
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", raceSentence, "--reason", "no source gives the rise")
		return err
	})
	md := renders(t, runDir)
	if err == nil || !strings.Contains(err.Error(), "still in the report") {
		t.Fatalf("a retire of a claim another edit wrote back at that moment = %v\n%s", err, md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
		t.Errorf("the refused retire recorded %d retire(s)", n)
	}
}

// V3: ESTOPPEL IS ASKED OF THE RECORD THE MINT LANDS ON. Blue accepts the lens's prescription at the
// moment the lens mints a fresh gap on the text it prescribed; checked against the record before the
// accept, the mint opens a clean slate against the lens's own words.
func TestEstoppelIsAskedOfTheRecordTheMintLandsOn(t *testing.T) {
	const (
		was = "The index stood at ninety-one points in the spring survey of producers"
		fix = was + ", a revised figure"
	)
	runDir := newRun(t)
	writeReport(t, runDir, "# Index\n\nThe year opened badly. "+was+". Volume recovered later.\n")
	registerBlue(t, runDir)
	g1, err := mintQuote(t, runDir, "G1", was, "--new", fix)
	if err != nil {
		t.Fatal(err)
	}
	tok := anchor.Token(g1)
	accept := recordtest.At(t, "blue-lane-1", "race:accept", &recordpb.BlueEdit{Old: proto.String(was + tok), New: proto.String(fix + tok),
		Answers: proto.String(g1), Accepted: proto.Bool(true), AppliedVerbatim: proto.Bool(true), Text: proto.String("red's text is right")})
	err = runtest.Overtaken(t, runDir, []*recordpb.Event{accept}, func() error {
		_, err := mintQuote(t, runDir, "G2", was)
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "estoppel") {
		t.Fatalf("a mint on text the lens prescribed, accepted at that moment = %v", err)
	}
	if md := renders(t, runDir); !strings.Contains(md, fix+tok) {
		t.Errorf("the accepted fix does not stand whole in the report:\n%s", md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 1 {
		t.Errorf("the record holds %d mints, want the prescribing one alone", n)
	}
}

// A PRESCRIPTION IS VERIFIED AGAINST THE REPORT THE MINT LANDS ON. Another lens places a finding's
// anchor inside the span this lens prescribes a replacement for, at the moment it mints; verified
// against the report before it, the mint records a "verified" fix whose replacement would drop that
// anchor — a pair blue's --accept then refuses as stale.
func TestAPrescriptionIsVerifiedOnTheReportTheMintLandsOn(t *testing.T) {
	runDir := raceRun(t)
	id := record.NewID("finding")
	// THE TREE IS DRIVEN WITHOUT run's refusal checks: the refusal this race earns is the transit
	// refusal bluedoc words for blue (anchor.Label names `edit` and `retire`), which reaches a lens
	// through any --new over an anchored span, race or none — a wording defect of its own, not this
	// guard's.
	err := runtest.Overtaken(t, runDir, findingAt(t, "red-lens-logic", id, "Costs rose sharply"), func() error {
		t.Cleanup(func() { _ = recordsql.CloseUnder(runDir) })
		_, err := Invoke(NewRootFor(lensSeat), []string{"mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", "G1", "--class", "overclaim", "--problem", "the defect", "--fix", "the fix",
			"--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
			"--likelihood", "medium", "--impact", "medium", "--complexity", "low",
			"--quote", raceSentence + ".", "--new", "Costs rose modestly across the first quarter."}, lensSeat)
		return err
	})
	md := renders(t, runDir)
	if err == nil {
		m := lastBody(t, runDir, &recordpb.Mint{})
		t.Fatalf("a mint verified a fix that drops an anchor placed in its span at that moment: fix_basis %s\n%s", m.GetFixBasis(), md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
		t.Errorf("the refused mint recorded %d mint(s)", n)
	}
}

// A MINT RECORDS WHETHER THE RUN COINED ITS CLASS AS THE RECORD IT LANDS ON SAYS. Another lens coins the
// slug at the moment this one mints with it; read before the mint's write, class_new says the run did
// not coin a class the record it joins shows this run coined.
func TestAMintsClassNewIsReadOffTheRecordItLandsOn(t *testing.T) {
	runDir := raceRun(t)
	coin := recordtest.At(t, "red-lens-logic", "race:class", &recordpb.ClassNew{Slug: proto.String("race-coined"),
		Definition: proto.String("a defect of one shape"), Neighbor: proto.String("overclaim"), Distinguisher: proto.String("it is not an overclaim"),
		MaterialDefault: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum()})
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{coin}, func() error {
		_, err := runMint(t, runDir, "--run", runDir, "--seat-id", lensSeat, "--key", "G1", "--class", "race-coined",
			"--problem", "the defect", "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low", "--quote", raceSentence)
		return err
	})
	if err != nil {
		t.Fatalf("a mint naming a class coined at that moment: %v", err)
	}
	if m := lastBody(t, runDir, &recordpb.Mint{}); !m.GetClassNew() {
		t.Error("the mint records class_new false over a record that shows the run coined its class")
	}
}

// V2: THE OUTCOME'S VERDICT IS DERIVED FROM THE RECORD IT LANDS ON. The chair records a PASS at the
// moment the bench records the outcome UNVERIFIED; derived before the PASS, the outcome asserts a
// verdict the record it joins does not derive.
func TestTheOutcomeIsCrossCheckedOnTheRecordItLandsOn(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := actFixture(t)
	pass := recordtest.At(t, "red-chair", "race:gate", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()})
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{pass}, func() error {
		_, err := run(t, "outcome", "--run", runDir, "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "the run ended")
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "contradicts the record") {
		t.Fatalf("an UNVERIFIED outcome over a PASS recorded at that moment = %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_OUTCOME); n != 0 {
		t.Errorf("the refused outcome recorded %d outcome(s)", n)
	}
}

// V8: A SEAT'S REFUSAL IS LOGGED ONLY ON A RUN THAT IS NOT OVER, AS THE ENTRY LANDS. The bench records
// the outcome at the moment a lens's refused call is logged; asked before the entry's write, the entry
// lands after the outcome, on a record closed to seats.
func TestARefusalIsNotLoggedAfterTheOutcome(t *testing.T) {
	runDir := raceRun(t)
	outcome := recordtest.At(t, "judge", "race:outcome", &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(),
		VerdictBasis: proto.String(record.VerdictAsserted), Prose: proto.String("the run ended")})
	err := runtest.Overtaken(t, runDir, []*recordpb.Event{outcome}, func() error {
		_, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--quote", raceSentence)
		return err
	})
	if err == nil {
		t.Fatal("a finding with no --reason was recorded")
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG); n != 0 {
		t.Errorf("the refusal was logged on a record holding an outcome: %d log entr(ies)", n)
	}
}
