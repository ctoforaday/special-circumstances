package fuzz

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

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
//     --anchor — carries `claim` equal to its anchor's live sentence in some report state between its
//     sitting's opening and its event (checkSentences); one that does not fails the run.
//     `reproducesWithClaim` = `reproduces`, `anchorVerifiesWithClaim` = `anchorVerifies`.
//   - STALENESS: one simulated debate per part is forced to hold a check of a sentence an edit then
//     changes (forcedStaleness), and the test seam takes the anchor out of that edit's `reopened` once
//     the debate ends (runtest.DropReopened, never a seat path); the oracle must report the debate.
//     `stalenessForged` = `stalenessReported`.

// stalenessSeeds are the seeds whose runs are FORCED through the staleness arm: four consecutive
// seeds past every other forced seed, so each of the sweep's four parts (sweepPart.owns: seed-1 mod 4)
// runs exactly one.
var stalenessSeeds = []int64{20, 21, 22, 23}

// stalenessClaim is the sentence the forced arm adds after sourceDemandOpen, and stalenessEdit the
// span of it blue's edit rewrites to stalenessEdited: a sentence no other drive quotes, so no drawn act
// judges, edits or retires the arm's computation.
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
func forgeStaleness(o *outcome) (bool, string) {
	if o.stalenessAnchor == "" {
		return false, ""
	}
	run, err := record.OpenRun(o.runDir)
	if err != nil {
		return false, "forged staleness: the run does not resolve: " + err.Error()
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
	// reproducesWithClaim is the reproduces whose claim is their anchor's sentence in the window.
	reproducesWithClaim int
	// anchorVerifies is the verify readings of a citation (`--anchor`), and anchorVerifiesWithClaim
	// those whose claim is their anchor's sentence in the window.
	anchorVerifies, anchorVerifiesWithClaim int
}

// checkSentences is the recorded-sentence arm, read at the run's end: each check of an evidence
// anchor on the record — a reproduce, a verify --anchor — carries `claim`, and it is the sentence the
// seat judged: the anchor's sentence in the report as the verb read it before its write. It returns
// the counts, and why the first check that does not fails the run.
//
// THE READ IS A WINDOW, NOT AN EVENT. The verb reads the sentence (record.LiveSentence) and then
// appends, and the debate's seats act concurrently, so an edit may land between the two. The read is
// inside the seat's sitting, so the claim must be the anchor's sentence in the report as some state
// from the sitting's opening to the check itself held it: as of the opening, and after each report
// change in between.
func checkSentences(run record.Run, fam record.Family) (got judgedSentences, msg string) {
	db, err := recordsql.Open(filepath.Join(run.Records(), "record.db"))
	if err != nil {
		return got, "recorded sentence: the record could not be opened: " + err.Error()
	}
	reportAt := func(at int64) (string, error) {
		src, err := recordsql.ReportSourceAt(db, at)
		if err != nil {
			return "", err
		}
		return src.Fold(reportproj.Replay)
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
		w := fam.At.Of(e)
		if claim == "" {
			if msg == "" {
				msg = fmt.Sprintf("recorded sentence: %s at event %d of %s recorded no claim", verb, w.ID, id)
			}
			continue
		}
		// A check in no sitting says nothing about when it read: the write's own state is its window.
		ats := []int64{w.ID}
		if w.SittingID != 0 {
			ats = append(ats, w.SittingID)
			ops, err := recordsql.ReportOpsBetween(db, w.SittingID, w.ID)
			if err != nil {
				return got, "recorded sentence: the report's changes before event " + fmt.Sprint(w.ID) + " could not be read: " + err.Error()
			}
			for _, op := range ops {
				ats = append(ats, op.EventID)
			}
		}
		var held []string
		for i, at := range ats {
			report, err := reportAt(at)
			if err != nil {
				// The report the write read renders, or the verb would have refused; an earlier state
				// that does not holds no sentence the read could have taken.
				if i == 0 {
					return got, fmt.Sprintf("recorded sentence: the report as of event %d does not render: %v", at, err)
				}
				continue
			}
			if s, ok := record.LiveSentencesIn(report, []string{id})[id]; ok {
				held = append(held, s)
			}
		}
		if !slices.Contains(held, claim) {
			if msg == "" {
				msg = fmt.Sprintf("recorded sentence: %s at event %d of %s recorded claim %q, and the report held %q for it from its sitting's opening (event %d) to the write",
					verb, w.ID, id, claim, strings.Join(slices.Compact(held), `" or "`), w.SittingID)
			}
			continue
		}
		*admitted++
	}
	return got, msg
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
