package cli

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
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

// otherSeat is another seat's own connection on runDir's record, as a seat in another process holds.
func otherSeat(t *testing.T, runDir string) *sql.DB {
	t.Helper()
	db, err := recordsql.OpenSeparate(filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

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
	md, err := record.RenderRun(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("the report no longer renders — a write landed that replay cannot apply: %v", err)
	}
	return md
}

// blueEditBy is seat's recorded edit of old to new, as its write path records it.
func blueEditBy(t *testing.T, seat, key, old, new string) *recordpb.Event {
	return recordtest.At(t, seat, key, &recordpb.BlueEdit{Old: proto.String(old), New: proto.String(new), Text: proto.String("the sentence overstates the rise")})
}

// gapAt is lens's gap id minted at loc and the Anchor that places it.
func gapAt(t *testing.T, lens, id, loc string) []*recordpb.Event {
	return []*recordpb.Event{
		recordtest.At(t, lens, lens+":mint:"+id, &recordpb.Mint{GapId: proto.String(id), Location: proto.String(loc),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("the rise is not sourced"),
			RequiredFix: proto.String("source it"), AcceptanceCheck: proto.String("the rise cites a source"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_LOW),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, lens, lens+":anchor:"+id, &recordpb.Anchor{Id: proto.String(id), Location: proto.String(loc)}),
	}
}

// V6: A MINT'S QUOTE IS PLACED ON THE REPORT ITS ANCHOR LANDS ON. Blue's edit takes the quoted
// sentence out while the mint is written; placed against the report before the edit, the mint
// and its Anchor land on text the report no longer holds, and replay cannot place the anchor.
func TestAMintIsPlacedOnTheReportItsAnchorLandsOn(t *testing.T) {
	runDir := raceRun(t)
	edit := blueEditBy(t, "blue-lane-1", "race:edit", raceSentence, "Costs rose only modestly in the first quarter")
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{edit}, func() error {
		_, err := runMint(t, "--run", runDir, "--seat-id", lensSeat,
			"--key", "G1", "--class", "spec_underspecification", "--problem", "the rise is not sourced", "--fix", "source it",
			"--check-kind", "document", "--check", "the rise cites a source", "--severity", "low",
			"--likelihood", "low", "--impact", "low", "--quote", raceSentence)
		return err
	})
	md := renders(t, runDir)
	if err == nil {
		t.Fatalf("a mint quoting text an edit took out at that moment was recorded:\n%s", md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
		t.Errorf("the refused mint recorded %d mint(s)", n)
	}
}

// V5: BLUE'S EDIT IS PLANNED ON THE REPORT IT LANDS ON. A lens's gap places its anchor inside the
// span blue's edit replaces, at that moment; planned against the report before it, the edit's
// replacement does not carry the anchor, and replaying the edit takes the gap's anchor out of the
// report — the loss every edit is refused for.
func TestAnEditIsPlannedOnTheReportItLandsOn(t *testing.T) {
	runDir := raceRun(t)
	id := record.NewID("gap")
	err := runtest.Overtaken(t, otherSeat(t, runDir), gapAt(t, lensSeat, id, "Costs rose sharply"), func() error {
		_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E", "--quote", raceSentence,
			"--new", "Costs rose only modestly in the first quarter", "--reason", "the source gives a modest rise")
		return err
	})
	md := renders(t, runDir)
	if !strings.Contains(md, anchor.Token(id)) {
		t.Errorf("the gap's anchor is not in the report:\n%s", md)
	}
	if err == nil || !strings.Contains(err.Error(), anchor.Token(id)) {
		t.Errorf("an edit of a span a gap's anchor landed in at that moment = %v, want refused naming the anchor\n%s", err, md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Errorf("the refused edit recorded %d edit(s)", n)
	}
}

// V7: A RETIRE IS CHECKED ON THE REPORT IT LANDS ON. The retire reads the sentence its anchor stands
// on, which stays word for word, and records it. Another blue seat's write commits at that moment: a
// rewrite of the sentence, carrying the anchor, and the retire records the sentence the report holds
// when it lands; a cut of the sentence, taking the anchor, and the retire is refused — checked against
// the report before it, the retire takes an anchor the report no longer holds.
func TestARetireIsCheckedOnTheReportItLandsOn(t *testing.T) {
	t.Run("a rewrite carrying the anchor", func(t *testing.T) {
		runDir := raceRun(t)
		f := markClaim(t, runDir, "G1", raceSentence, "spec_underspecification")
		if err := lensClose(t, runDir, lensSeat, f, "not_a_defect"); err != nil {
			t.Fatalf("close %s: %v", f, err)
		}
		tok := anchor.Token(f)
		back := blueEditBy(t, "blue-lane-1", "race:edit", "Costs rose sharply across", "Costs rose across")
		err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{back}, func() error {
			_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", f, "--reason", "the gap's marker no longer applies")
			return err
		})
		md := renders(t, runDir)
		if err != nil {
			t.Fatalf("a retire of an anchor whose sentence another edit rewrote at that moment = %v\n%s", err, md)
		}
		if got, want := lastBody(t, runDir, &recordpb.Retire{}).GetClaim(), "Costs rose across the first quarter."; got != want {
			t.Errorf("the retire recorded the sentence %q, want the one the report held when it landed, %q", got, want)
		}
		if strings.Contains(md, tok) || !strings.Contains(md, "Costs rose across the first quarter.") {
			t.Errorf("the retire did not take %s off the rewritten sentence:\n%s", tok, md)
		}
	})
	t.Run("a cut taking the anchor", func(t *testing.T) {
		runDir := raceRun(t)
		f := markClaim(t, runDir, "G1", raceSentence, "spec_underspecification")
		if err := lensClose(t, runDir, lensSeat, f, "not_a_defect"); err != nil {
			t.Fatalf("close %s: %v", f, err)
		}
		tok := anchor.Token(f)
		gone := recordtest.At(t, "blue-lane-1", "race:cut", &recordpb.BlueEdit{Old: proto.String(" " + raceSentence + tok + "."), New: proto.String(""),
			RemovedAnchors: []string{f}, Text: proto.String("the rise is not sourced")})
		err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{gone}, func() error {
			_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", f, "--reason", "the gap's marker no longer applies")
			return err
		})
		md := renders(t, runDir)
		if err == nil || !strings.Contains(err.Error(), "the report holds no") || !strings.Contains(err.Error(), f) {
			t.Fatalf("a retire of an anchor another edit cut at that moment = %v\n%s", err, md)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != 0 {
			t.Errorf("the refused retire recorded %d retire(s)", n)
		}
	})
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
	// Overtaken inserts past the write path, so the estoppel row writeTombstones writes beside the
	// accept goes in with it, in its transaction.
	stone := recordtest.At(t, "blue-lane-1", "race:tombstone", &recordpb.Tombstone{GapId: proto.String(g1), Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL.Enum()})
	err = runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{accept, stone}, func() error {
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

// A PRESCRIPTION IS VERIFIED AGAINST THE REPORT THE MINT LANDS ON. Another lens places a gap's
// anchor inside the span this lens prescribes a replacement for, at the moment it mints; verified
// against the report before it, the mint records a "verified" fix whose replacement would drop that
// anchor — a pair blue's --accept then refuses as stale.
func TestAPrescriptionIsVerifiedOnTheReportTheMintLandsOn(t *testing.T) {
	runDir := raceRun(t)
	id := record.NewID("gap")
	err := runtest.Overtaken(t, otherSeat(t, runDir), gapAt(t, "red-lens-logic", id, "Costs rose sharply"), func() error {
		_, err := runMint(t, "--run", runDir, "--seat-id", lensSeat,
			"--key", "G1", "--class", "spec_underspecification", "--problem", "the defect", "--fix", "the fix",
			"--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
			"--likelihood", "medium", "--impact", "medium", "--complexity", "low",
			"--quote", raceSentence+".", "--new", "Costs rose modestly across the first quarter.")
		return err
	})
	md := renders(t, runDir)
	if err == nil {
		m := lastBody(t, runDir, &recordpb.Mint{})
		t.Fatalf("a mint verified a fix that drops an anchor placed in its span at that moment: fix_basis %s\n%s", m.GetFixBasis(), md)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 1 {
		t.Errorf("the record holds %d mints, want the overtaking lens's alone", n)
	}
}

// V2: THE OUTCOME'S VERDICT IS DERIVED FROM THE RECORD IT LANDS ON. The chair records a PASS at the
// moment the bench records the outcome UNVERIFIED; derived before the PASS, the outcome asserts a
// verdict the record it joins does not derive.
func TestTheOutcomeIsCrossCheckedOnTheRecordItLandsOn(t *testing.T) {
	t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())
	runDir := actFixture(t)
	sitToAssemble(t, runDir)
	pass := recordtest.At(t, "red-chair", "race:gate", &recordpb.Gate{})
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{pass}, func() error {
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

// V8: A SEAT'S REFUSAL IS LOGGED ON A RUN THAT IS OVER, TOO. The bench records the outcome at the
// moment a lens's refused call is logged; the entry lands on the record holding it, because a refusal
// after the run's end is friction the operator reads (record.requireOpenToSeats lets every log entry
// land).
func TestARefusalIsLoggedAfterTheOutcome(t *testing.T) {
	runDir := raceRun(t)
	outcome := recordtest.At(t, "judge", "race:outcome", &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(),
		VerdictBasis: proto.String(record.VerdictAsserted), Prose: proto.String("the run ended")})
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{outcome}, func() error {
		_, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat, "--quote", raceSentence)
		return err
	})
	if err == nil {
		t.Fatal("a mint with no --problem was recorded")
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG); n != 1 {
		t.Errorf("the refusal on a record holding an outcome: %d log entr(ies), want 1", n)
	}
}
