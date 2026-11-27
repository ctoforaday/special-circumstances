package fuzz

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE JUDGED SENTENCE (plans/compute-and-show.md Part 1a″, III.1″.6): a check records the sentence it
// judged, and a standing check of a wording the report no longer holds fails the consistency oracle.
// Two arms, each with its tally gate.
//
//   - RECORDED SENTENCE: every check of an evidence anchor the drive records — a reproduce, a verify
//     --anchor — carries `claim` equal to its anchor's live sentence in the report as of its own event
//     (checkSentences); one that does not fails the run.
//     `reproducesWithClaim` = `reproduces`, `anchorVerifiesWithClaim` = `anchorVerifies`.
//   - STALENESS: one simulated debate per part is forced to hold a check of a sentence an edit then
//     changes (forcedStaleness), and the test seam takes the anchor out of that edit's `reopened` once
//     the debate ends (runtest.DropReopened, never a seat path); the oracle must report the debate.
//     `stalenessForged` = `stalenessReported`.

// stalenessSeeds (forcedseeds_test.go) are the seeds whose runs are FORCED through the staleness arm.

// stalenessClaim is the sentence the forced arm adds after sourceDemandOpen, and stalenessEdit the
// span of it blue's edit rewrites to stalenessEdited: a sentence no other drive quotes, so no drawn act
// judges or edits the arm's computation, and the drawn retire, which takes an anchor by its id, never
// draws it (drawnRetireAnchor).
const (
	stalenessClaim  = "The staleness claim holds"
	stalenessEdit   = "The staleness claim"
	stalenessEdited = "The edited staleness claim"
)

// stalenessJudge is the lens whose judgement the arm forges standing: not the computation's owner, so
// no drawn act of its judges the computation again, and its judgement made before the edit is its
// latest when the debate ends.
const stalenessJudge = otherLens

// stalenessDriven is stalenessStage once forcedStaleness has driven every step.
const stalenessDriven = "driven"

// forcedStaleness is a stalenessSeeds run's arm, at blue-respond's first sitting, before any act of its
// own: blue adds stalenessClaim and computes it; stalenessJudge, in a sitting of its own, judges the
// computation sound; blue edits the sentence, and its edit names the computation reopened. The record
// is honest when the arm ends — the judgement no longer stands — and the seam forges it after the
// debate (forgeStaleness). It draws nothing from the run's RNG.
func (r *runner) forcedStaleness(seatID string) {
	r.mu.Lock()
	due := r.forceStaleness && seatID == "blue-respond" && r.stalenessStage == ""
	if due {
		r.stalenessStage = "begun"
	}
	r.mu.Unlock()
	if !due {
		return
	}
	stage, anchor := r.driveStaleness(seatID)
	r.mu.Lock()
	r.stalenessStage, r.stalenessAnchor = stage, anchor
	r.mu.Unlock()
}

// driveStaleness drives forcedStaleness's steps and returns how far it reached — stalenessDriven, or
// the step that failed — and the computation it judged.
func (r *runner) driveStaleness(blue string) (string, string) {
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+stalenessClaim).
		set("--reason", "fuzz: the claim the staleness arm rests on").admit(); !ok {
		return "blue's edit adding the claim was refused", ""
	}
	k := r.lifecycleCompute(blue, stalenessClaim, "fuzz-staleness.js", "", "")
	if k == "" {
		return "blue's computation of the claim was refused", ""
	}
	if !r.openSitting(stalenessJudge) {
		return "the register of " + stalenessJudge + " was refused", k
	}
	if !r.lifecycleJudge(stalenessJudge, k, "sound") {
		return stalenessJudge + "'s judgement of " + k + " was refused", k
	}
	if _, ok := r.do("edit", blue).set("--quote", stalenessEdit).set("--new", stalenessEdited).
		set("--reason", "fuzz: the sentence the judgement was of, rewritten").admit(); !ok {
		return "blue's edit of the judged sentence was refused", k
	}
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		return "the record could not be read after the edit: " + err.Error(), k
	}
	for i := len(fam.Events) - 1; i >= 0; i-- {
		if e, ok := recordpb.BodyAs[*recordpb.BlueEdit](fam.Events[i]); ok {
			if !slices.Contains(e.GetReopened(), k) {
				return fmt.Sprintf("blue's edit of the judged sentence names %v reopened, not %s", e.GetReopened(), k), k
			}
			return stalenessDriven, k
		}
	}
	return "no edit of the judged sentence is on the record", k
}

// stalenessForcedErr is why a stalenessSeeds run failed its arm, or "" where it drove it or is not
// forced. Called with r.mu held.
func (r *runner) stalenessForcedErr() string {
	switch {
	case !r.forceStaleness:
		return ""
	case r.stalenessStage == "":
		return "forced staleness: blue-respond never sat"
	case r.stalenessStage != stalenessDriven:
		return "forced staleness: " + r.stalenessStage
	}
	return ""
}

// forgeStaleness is the seam's half of the staleness arm, at the debate's end, where no seat writes:
// it takes the arm's computation out of every edit naming it reopened, so the judgement stalenessJudge
// made before the edit stands on the earlier wording. It returns whether it forged anything.
//
// A COMPUTATION OUT OF THE REPORT HAS NO JUDGEMENT TO FORGE: a check of an anchor a retire or a cut
// took stands nowhere (record.StandingChecks), so the seam refuses, naming the act, rather than count a
// forged staleness the oracle can never owe.
func forgeStaleness(o *outcome) (bool, string) {
	if o.stalenessAnchor == "" {
		return false, ""
	}
	run, err := record.OpenRun(o.runDir)
	if err != nil {
		return false, "forged staleness: the run does not resolve: " + err.Error()
	}
	removal, removed, err := record.AnchorRemovedAt(run, o.stalenessAnchor)
	if err != nil {
		return false, "forged staleness: whether " + o.stalenessAnchor + " is in the report could not be read: " + err.Error()
	}
	if removed {
		return false, fmt.Sprintf("forged staleness: %s left the report by a %s at event %d, so no judgement of it stands to forge; no drawn act may take the arm's computation", o.stalenessAnchor, removal.By, removal.Event)
	}
	edits, err := runtest.DropReopened(run, o.stalenessAnchor)
	if err != nil {
		return false, "forged staleness: the seam could not drop " + o.stalenessAnchor + " from its edits: " + err.Error()
	}
	if len(edits) == 0 {
		return false, "forged staleness: no edit names " + o.stalenessAnchor + " reopened, so there was nothing to forge"
	}
	return true, ""
}

// staleOf splits the oracle's violations into the staleness ones naming anchor — the forged
// disagreement the oracle owes the forged debate — and every other, each of which fails the run.
func staleOf(res consistency.Result, anchor string) (forged, other []string) {
	for _, v := range res.Violations {
		if anchor != "" && v.Rule == "staleness" && v.Anchor == anchor {
			forged = append(forged, v.String())
			continue
		}
		other = append(other, v.String())
	}
	return forged, other
}

// judgedSentences is the recorded-sentence arm's count over one run's record: the checks of an
// evidence anchor on it by verb, and those whose `claim` the arm admits (checkSentences).
type judgedSentences struct {
	// reproducesWithClaim is the reproduces whose claim is their anchor's sentence at their event.
	reproducesWithClaim int
	// anchorVerifies is the verify readings of a citation (`--anchor`), and anchorVerifiesWithClaim
	// those whose claim is their anchor's sentence at their event.
	anchorVerifies, anchorVerifiesWithClaim int
}

// checkSentences is the recorded-sentence arm, read at the run's end: each check of an evidence
// anchor on the record — a reproduce, a verify --anchor — carries `claim`, and it is the anchor's
// sentence in the report folded as of the check's own event. It returns the counts, and why the first
// check that does not fails the run.
//
// IT IS EQUALITY AT THE EVENT. The verb reads the sentence before its write, and the landing refuses
// the check unless the anchor still backs that sentence (record.CheckOfSentence), so a check that
// recorded a row was of the sentence the report held when it landed. A refused check writes no row and
// is not read here.
func checkSentences(run record.Run, fam record.Family) (got judgedSentences, msg string) {
	path := filepath.Join(run.Records(), "record.db")
	// A handle of its own, closed on return, so the process's cached handle — the one the rest of the
	// sweep reads through — is left as it was.
	db, err := recordsql.OpenSeparate(path)
	if err != nil {
		return got, "recorded sentence: the record could not be opened: " + err.Error()
	}
	defer func() {
		if err := db.Close(); err != nil && msg == "" {
			msg = "recorded sentence: the record's handle did not close: " + err.Error()
		}
	}()
	// The report as of a check's event is the replay of every op up to it from the empty text: no
	// reader of the record reads the report at a past event, so the gate folds the ops itself.
	reportAt := func(at int64) (string, error) {
		ops, err := recordsql.ReportOpsBetween(db, 0, at)
		if err != nil {
			return "", err
		}
		return reportproj.Replay("", false, ops)
	}
	for _, e := range fam.Events {
		var verb, id, claim string
		var admitted *int
		if rp, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
			verb, id, claim, admitted = "reproduce", rp.GetComputationId(), rp.GetClaim(), &got.reproducesWithClaim
		} else if vf, ok := recordpb.BodyAs[*recordpb.Verify](e); ok && vf.GetAnchor() != "" {
			got.anchorVerifies++
			verb, id, claim, admitted = "verify", vf.GetAnchor(), vf.GetClaim(), &got.anchorVerifiesWithClaim
		} else {
			continue
		}
		at := fam.At.Of(e).ID
		if claim == "" {
			if msg == "" {
				msg = fmt.Sprintf("recorded sentence: %s at event %d of %s recorded no claim", verb, at, id)
			}
			continue
		}
		report, err := reportAt(at)
		if err != nil {
			return got, fmt.Sprintf("recorded sentence: the report as of event %d does not render: %v", at, err)
		}
		live := record.SentencesIn(report, []string{id})
		if now, ok := live[id]; !ok || now != claim {
			if msg == "" {
				held := fmt.Sprintf("%q", now)
				if !ok {
					held = "no sentence for it"
				}
				msg = fmt.Sprintf("recorded sentence: %s at event %d of %s recorded claim %q, and the report as of that event held %s", verb, at, id, claim, held)
			}
			continue
		}
		*admitted++
	}
	return got, msg
}

// retireKeeps is why a retire of id, which recorded claim, did not keep its sentence, or "": the report
// after it is the report before it less id's one token, and claim is the sentence id stood on before
// it (record.SentencesIn, the reading the verb stamps it with).
func retireKeeps(before, after, id, claim string) string {
	if now, ok := record.SentencesIn(before, []string{id})[id]; !ok || now != claim {
		return fmt.Sprintf("recorded claim %q, and the report before it held %q for %s", claim, now, id)
	}
	if want := strings.Replace(before, anchor.Token(id), "", 1); after != want {
		return fmt.Sprintf("changed the report beyond taking %s's token off its sentence %q", id, claim)
	}
	return ""
}

// retiresKeepSentences is held at the run's end over every retire on the record, whichever drive made
// it: A RETIRE TAKES ONE ANCHOR AND KEEPS ITS SENTENCE (retireKeeps), at the retire's own event — the
// report folded through the ops before it and through its own. A sentence a later edit rewrites is
// that edit's (editHistory renders it); what the retire owes is the sentence as it left it. It returns
// the retires it held, and why the first that does not keep its sentence fails the run.
func retiresKeepSentences(run record.Run, fam record.Family) (held int, msg string) {
	db, err := recordsql.OpenSeparate(filepath.Join(run.Records(), "record.db"))
	if err != nil {
		return 0, "retire keeps its sentence: the record could not be opened: " + err.Error()
	}
	defer func() {
		if err := db.Close(); err != nil && msg == "" {
			msg = "retire keeps its sentence: the record's handle did not close: " + err.Error()
		}
	}()
	reportAt := func(at int64) (string, error) {
		ops, err := recordsql.ReportOpsBetween(db, 0, at)
		if err != nil {
			return "", err
		}
		return reportproj.Replay("", false, ops)
	}
	for _, e := range fam.Events {
		rt, ok := recordpb.BodyAs[*recordpb.Retire](e)
		if !ok {
			continue
		}
		at := fam.At.Of(e).ID
		if len(rt.GetAnchors()) != 1 {
			return held, fmt.Sprintf("retire keeps its sentence: the retire at event %d names %d anchors %v, and a retire takes one", at, len(rt.GetAnchors()), rt.GetAnchors())
		}
		before, err := reportAt(at - 1)
		if err != nil {
			return held, fmt.Sprintf("retire keeps its sentence: the report before event %d does not render: %v", at, err)
		}
		after, err := reportAt(at)
		if err != nil {
			return held, fmt.Sprintf("retire keeps its sentence: the report as of event %d does not render: %v", at, err)
		}
		if why := retireKeeps(before, after, rt.GetAnchors()[0], rt.GetClaim()); why != "" {
			return held, fmt.Sprintf("retire keeps its sentence: the retire at event %d %s", at, why)
		}
		held++
	}
	return held, ""
}

// A RETIRE KEEPS ITS SENTENCE, held at its event: the report less the one token, and the claim the
// sentence it stood on; a claim of another wording, or a text that lost more than the token, fails.
func TestRetireKeepsReadsTheSentenceAtTheRetire(t *testing.T) {
	const id = "K-00000001"
	tok := anchor.Token(id)
	before := "The cost is climbing sharply across the whole sampled corpus." + tok + " Another sentence holds.\n"
	claim := record.SentencesIn(before, []string{id})[id]
	if claim == "" {
		t.Fatalf("the fixture's anchor stands on no sentence: %q", before)
	}
	kept := strings.Replace(before, tok, "", 1)
	for _, c := range []struct {
		name, after, claim string
		fails              bool
	}{
		{"kept", kept, claim, false},
		{"another wording", kept, "The cost is rising over time across the whole sampled corpus.", true},
		{"the sentence went too", " Another sentence holds.\n", claim, true},
		{"the token stayed", before, claim, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			if why := retireKeeps(before, c.after, id, c.claim); (why != "") != c.fails {
				t.Errorf("retireKeeps = %q, want a failure %v", why, c.fails)
			}
		})
	}
}

// THE RETIRE ORACLE READS THE RECORD: a real retire of an anchor off its sentence is held and passes;
// the same record with the retire's claim doctored to another wording fails, naming the event.
func TestRetiresKeepSentencesHoldsEveryRetireOnTheRecord(t *testing.T) {
	r := stagedLifecycleRunner(t)
	cite := r.lifecycleCite("blue-respond", sourceDemandOpen, "/retire-keeps", "")
	if cite == "" {
		t.Fatalf("blue's citation was refused (%s)", r.actErr)
	}
	r.retireOneAnchor("blue-respond", cite)
	if r.actErr != "" || r.anchorRetiresKept != 1 {
		t.Fatalf("the retire of %s: kept %d (%s)", cite, r.anchorRetiresKept, r.actErr)
	}
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if held, msg := retiresKeepSentences(r.run(), fam); held != 1 || msg != "" {
		t.Fatalf("the honest record: held %d (%s), want 1 and no failure", held, msg)
	}
	path := filepath.Join(r.run().Records(), "record.db")
	db, err := recordsql.OpenSeparate(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE "retire" SET "claim" = 'A sentence the report never held.'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	// The process's cached handle decoded the retire before the doctoring: released, the next read
	// decodes it afresh (runtest.DropReopened).
	if err := recordsql.Close(path); err != nil {
		t.Fatal(err)
	}
	if fam, err = record.FamilyOf(r.run()); err != nil {
		t.Fatal(err)
	}
	if held, msg := retiresKeepSentences(r.run(), fam); held != 0 || !strings.Contains(msg, "A sentence the report never held.") {
		t.Fatalf("the doctored record: held %d (%q), want the retire refused naming its claim", held, msg)
	}
}

// THE STALENESS ARM AND THE RECORDED-SENTENCE ARM, against the built binary: a forced run's arm leaves
// an honest record the oracle has nothing to say of; the seam forges it, and the oracle reports the
// forged anchor; the judgement the arm recorded carries the sentence it judged. An unforced run
// drives nothing and owes nothing.
func TestTheStalenessArmForgesWhatTheOracleReports(t *testing.T) {
	t.Run("forced", func(t *testing.T) {
		r := stagedLifecycleRunner(t)
		r.forceStaleness = true
		r.mu.Lock()
		if msg := r.stalenessForcedErr(); !strings.Contains(msg, "never sat") {
			t.Fatalf("a forced run nothing has driven reads %q, want the sitting it never reached", msg)
		}
		r.mu.Unlock()
		r.forcedStaleness("blue-respond")
		r.mu.Lock()
		stage, anchor, msg := r.stalenessStage, r.stalenessAnchor, r.stalenessForcedErr()
		r.mu.Unlock()
		if stage != stalenessDriven || msg != "" || r.actErr != "" || anchor == "" {
			t.Fatalf("the forced arm reached %q on %q (%s; %s)", stage, anchor, msg, r.actErr)
		}

		fam, err := record.FamilyOf(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if got, msg := checkSentences(r.run(), fam); got.reproducesWithClaim != 1 || msg != "" {
			t.Errorf("the arm's judgement: %d reproduce(s) carry the sentence they judged (%s), want 1", got.reproducesWithClaim, msg)
		}

		o := outcome{runDir: r.runDir, stalenessAnchor: anchor}
		res, err := consistency.Check(r.run())
		if err != nil {
			t.Fatal(err)
		}
		if forged, other := staleOf(res, anchor); len(forged) != 0 || len(other) != 0 || len(res.Unmeasured) != 0 {
			t.Fatalf("the honest record: staleness %v, other violations %v, unmeasured %v", forged, other, res.Unmeasured)
		}
		forged, ferr := forgeStaleness(&o)
		if !forged || ferr != "" {
			t.Fatalf("the seam forged %v: %s", forged, ferr)
		}
		res, err = consistency.Check(r.run())
		if err != nil {
			t.Fatal(err)
		}
		stale, other := staleOf(res, anchor)
		if len(stale) != 1 || len(other) != 0 {
			t.Fatalf("the forged record: staleness %v, other violations %v; want one naming %s and nothing else", stale, other, anchor)
		}
		for _, want := range []string{stalenessJudge, `"` + stalenessClaim + `."`, stalenessEdited} {
			if !strings.Contains(stale[0], want) {
				t.Errorf("the violation does not name %s: %s", want, stale[0])
			}
		}
	})
	t.Run("not forced", func(t *testing.T) {
		r := stagedLifecycleRunner(t)
		r.forcedStaleness("blue-respond")
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.stalenessStage != "" || r.stalenessAnchor != "" || r.stalenessForcedErr() != "" {
			t.Errorf("an unforced run drove the arm: %q on %q", r.stalenessStage, r.stalenessAnchor)
		}
	})
}

// THE ARM'S COMPUTATION IS NEVER THE DRAWN RETIRE'S. A retire takes an anchor out of the report, and a
// check of an anchor out of it stands nowhere (record.StandingChecks), so the judgement the seam forges
// standing after a retire of the arm's computation is moot, and the oracle rightly reports nothing. CI
// run 45967095843 (seed 21): the seam took the computation out of its edit's reopened and the oracle
// reported no staleness — the drawn retire, which draws any anchor the report holds, had taken it.
func TestTheDrawnRetireLeavesTheStalenessArmsComputation(t *testing.T) {
	r := stagedLifecycleRunner(t)
	r.forceStaleness = true
	r.forcedStaleness("blue-respond")
	r.mu.Lock()
	k, msg := r.stalenessAnchor, r.stalenessForcedErr()
	r.mu.Unlock()
	if k == "" || msg != "" || r.actErr != "" {
		t.Fatalf("the forced arm left %q (%s; %s)", k, msg, r.actErr)
	}
	// AN ANCHOR BESIDE THE ARM'S, so the draw has one to take.
	if r.lifecycleCite("blue-respond", sourceDemandOpen, "/staleness-beside", "") == "" {
		t.Fatalf("blue's citation beside the arm's computation was refused (%s)", r.actErr)
	}
	drawn := 0
	for range 200 {
		switch id := r.drawnRetireAnchor(); id {
		case k:
			t.Fatalf("the drawn retire drew the staleness arm's computation %s", k)
		case "":
		default:
			drawn++
		}
	}
	if drawn == 0 {
		t.Fatal("the drawn retire drew nothing from a report holding the lanes' anchors, so the test measured no draw")
	}
}

// A FORGE OF A COMPUTATION OUT OF THE REPORT FAILS LOUD, naming the act that took it, rather than
// counting a forged staleness the oracle can never owe: a check of an anchor a retire or a cut took
// stands nowhere, so there is no staleness to forge. The oracle stays silent of it, as it must.
func TestTheSeamRefusesToForgeARemovedComputation(t *testing.T) {
	r := stagedLifecycleRunner(t)
	r.forceStaleness = true
	r.forcedStaleness("blue-respond")
	r.mu.Lock()
	k := r.stalenessAnchor
	r.mu.Unlock()
	if k == "" || r.actErr != "" {
		t.Fatalf("the forced arm left no computation (%s)", r.actErr)
	}
	r.retireOneAnchor("blue-respond", k)
	if r.actErr != "" {
		t.Fatalf("the retire of %s: %s", k, r.actErr)
	}
	o := outcome{runDir: r.runDir, stalenessAnchor: k}
	forged, ferr := forgeStaleness(&o)
	if forged || !strings.Contains(ferr, k) || !strings.Contains(ferr, record.RemovedByRetire) {
		t.Fatalf("the seam forged %v (%q) over a retired computation; want a refusal naming %s and its retire", forged, ferr, k)
	}
	res, err := consistency.Check(r.run())
	if err != nil {
		t.Fatal(err)
	}
	if stale, other := staleOf(res, k); len(stale) != 0 || len(other) != 0 {
		t.Fatalf("a retired computation's judgement: staleness %v, other violations %v; want none", stale, other)
	}
}

// NO FORCED ARM'S ANCHOR IS THE DRAWN RETIRE'S. Each forced arm below places an anchor and acts on it,
// or needs it standing, at a later sitting or at the run's end; the drawn retire, which takes any
// anchor the report holds by its id, would take it out first and leave the arm nothing to act on. The
// staleness arm's is TestTheDrawnRetireLeavesTheStalenessArmsComputation's; each other arm's is a
// subtest here, its flag set alone, over one report holding every arm's anchors and one free citation.
func TestTheDrawnRetireLeavesEveryForcedArmsAnchor(t *testing.T) {
	r := stagedLifecycleRunner(t)
	const blue = "blue-respond"
	const reopened, unsound, owed, limit, free = "The reopened claim holds", "The unsound claim holds", "The owed claim holds", "The limit claim holds", "The free claim holds"
	mintCite, mintComp := mintBareClaim("citation"), mintBareClaim("computation")
	added := strings.Join([]string{reopened, unsound, owed, limit, free, mintCite, mintComp}, ". ")
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+added).
		set("--reason", "fuzz: the forced arms' sentences").admit(); !ok {
		t.Fatalf("blue's edit adding the arms' sentences was refused (%s)", r.actErr)
	}
	bundle := func(claim, name, body string) string {
		t.Helper()
		dir, argv, err := r.writeBundle(name, body)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := r.do("compute", blue).set("--quote", claim).set("--dir", dir).set("--reason", "fuzz: the arm's computation").runs(argv).admit(); !ok {
			t.Fatalf("blue's computation of %q was refused (%s)", claim, r.actErr)
		}
		return lastComputationOn(r.run(), claim)
	}
	cite := func(claim, path string) string {
		t.Helper()
		c := r.lifecycleCite(blue, claim, path, "")
		if c == "" {
			t.Fatalf("blue's citation of %q was refused (%s)", claim, r.actErr)
		}
		return c
	}
	kReopened := bundle(reopened, reopenedComputationName, reopenedComputationScript)
	kUnsound := bundle(unsound, unsoundComputationName, unsoundComputationScript)
	kOwed := bundle(owed, "fuzz-owed-computation-1.js", "console.log('fuzz: owed computation 1');")
	kLimit := bundle(limit, "fuzz-limit-computation.js", "console.log('fuzz: the limit computation');")
	cLimit := cite(limit, "/forced-limit")
	cMint := cite(mintCite, "/forced-mint-bare")
	kMint := bundle(mintComp, "fuzz-mint-bare.js", "console.log('fuzz: the mint-bare computation');")
	cite(free, "/forced-free")
	if r.actErr != "" {
		t.Fatalf("placing the arms' anchors: %s", r.actErr)
	}
	for _, c := range []struct {
		name     string
		force    func(on bool)
		reserved []string
	}{
		{"reopened computation", func(on bool) { r.forceReopenedComputation = on }, []string{kReopened}},
		{"unsound computation", func(on bool) { r.forceUnsoundComputation = on }, []string{kUnsound}},
		{"owed computation", func(on bool) { r.forceOwedComputation = on }, []string{kReopened, kUnsound, kOwed, kLimit, kMint}},
		{"limit screen and debt", func(on bool) {
			r.forceScreenUnsound = on
			r.screenUnsoundComputation, r.limitDebtCite = "", ""
			if on {
				r.screenUnsoundComputation, r.limitDebtCite = kLimit, cLimit
			}
		}, []string{kLimit, cLimit}},
		{"mint bare", func(on bool) { r.forceMintBare = on }, []string{cMint, kMint}},
	} {
		t.Run(c.name, func(t *testing.T) {
			c.force(true)
			defer c.force(false)
			drawn := 0
			for range 200 {
				id := r.drawnRetireAnchor()
				if slices.Contains(c.reserved, id) {
					t.Fatalf("the drawn retire drew %s, the forced arm's", id)
				}
				if id != "" {
					drawn++
				}
			}
			if drawn == 0 {
				t.Fatal("the drawn retire drew nothing, so the test measured no draw")
			}
		})
	}
}
