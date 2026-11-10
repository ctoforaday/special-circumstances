package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// foundAgainstSeed is the seed whose run is FORCED through a found-against check's whole life
// (forceFoundAgainst), where a drawn run reaches the parts of it only on a chain of coins.
//
// Left to the draw, the citation owner's reading finds against a citation only while the lens has
// fresh stock (two mints a run, most spent in its first sitting), and blue's remedy — an edit or a
// retire — lands at blue's next sitting, which comes before the reader's next: so the reader's
// repeat over its own standing check, and its own backing re-read that clears it, need a later
// reading of the same citation with the same outcome while the check still stands. MEASURED at
// 59adcc3d7, c242e65e4 and 2ecae7a9b: zero repeats refused and zero closes admitted after the
// reader's own re-read on all three 40-run sweeps; one accept of a reading's fix on the twice-held
// sentence's cited copy across none of them, and `verify --new` passed on one sweep of the three.
//
// The run's blue cites a sentence at its first sitting (citeForFoundAgainst), and its citation owner
// draws no fresh mint until the chain has read what it is held to. In a sitting whose plan does not
// ready blue-respond — the first plan readies the lenses alone — the owner's reading of a citation
// finds against it and mints the gap, and the gap's close is refused while the check stands. That
// reading passes no --new: blue's accept of a prescription edits the sentence, and an edit clears the
// check the chain's next sitting reads. The plan readies blue on the standing check and on the gap,
// and readies the gap's lens only after blue has acted there, so the owner sits again before blue's
// remedy only where the plan readies it on its pin: the sitting the chain may start in serves the
// owner a selection of the report, never the whole (foundAgainstHoldsPin), and the chain starts only
// where the owner's pin is then behind the head. The lenses sit before blue in an epoch, so at the
// owner's next sitting its repeat with the same outcome is refused naming that reading, the close is
// refused again, and its own backing re-read clears the check, after which the close is admitted
// (foundAgainstCleared). A sitting holds one reading per citation, so the re-read is never the
// reading's own sitting. Once blue has made the twice-held sentence, the owner's reading of one copy
// finds against it with a --new (readRepeatedSentence) — the chain's one --new — and blue's next
// sitting accepts that fix (acceptRepeated). Every lane records its part, so the report holds the
// sentence twice.

// foundAgainstForcedRead is whether the forced run's citation owner has made both readings it is
// held to (its chain, and its reading of the twice-held sentence), after which it mints as any lens.
func (r *runner) foundAgainstForcedRead() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.foundAgainstChained && r.foundAgainstRepeatedRead
}

func (r *runner) reachedFoundAgainstStage(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.foundAgainstStage = s
}

// foundAgainstQuote is the words of lane 2's sentence blue's forced citation quotes: the part of it
// the edit drive's swap never rewrites.
const foundAgainstQuote = "across the whole sampled corpus"

// citeForFoundAgainst is blue's citation, at its first sitting in foundAgainstSeed's run, of the
// sentence the owner's chain reads against: a citation stands before the first lens sitting.
func (r *runner) citeForFoundAgainst(seatID string) {
	r.mu.Lock()
	due := r.forceFoundAgainst && r.foundAgainstCite == "" && !r.foundAgainstCiteTried
	r.mu.Unlock()
	if !due {
		return
	}
	// A REPORT THAT DOES NOT YET HOLD THE WORDS ONCE IS CITED AT A LATER BLUE SITTING.
	if cur, err := record.RenderRun(r.run()); err != nil || strings.Count(plainText(cur), foundAgainstQuote) != 1 {
		r.reachedFoundAgainstStage("no blue sitting so far found the chain's sentence once in the report")
		return
	}
	r.mu.Lock()
	r.foundAgainstCiteTried = true
	r.mu.Unlock()
	const path = "/found-against"
	out, ok := r.do("cite", seatID).set("--quote", foundAgainstQuote).set("--url", sourceURL(path)).set("--title", sourceTitle).
		set("--source-span", sourceBody(path)).bare("--json").admit()
	if !ok {
		r.reachedFoundAgainstStage("blue's citation of the chain's sentence was refused")
		return
	}
	var env struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.Label == "" {
		r.noteActErr("an admitted cite's --json result names no label: " + firstLine(out))
		return
	}
	r.mu.Lock()
	r.foundAgainstCite = env.Result.Label
	r.mu.Unlock()
}

// forcedFoundAgainstChain is the citation owner's chain in foundAgainstSeed's run (foundAgainstSeed):
// the reading that finds against a citation, with no --new, then, at its next sitting, the repeat, the
// refused close and its own re-read. It draws the reading's gap flags as the drawn reading does, and pins the check
// kind of a gap it mints to a document check; a reading the one-gap rule holds is carried by another
// lens's gap, whose check kind is that gap's, so each close of the carrier answers its live-answer demand
// first (answerEvidenceDemand). It reads only a citation whose sentence holds no gap a scenario
// closes behind a reproduction (foundAgainstCitation), since the chain's close reproduces nothing.
func (r *runner) forcedFoundAgainstChain(seatID string) {
	r.mu.Lock()
	due := r.forceFoundAgainst && seatID == citationOwner && !r.foundAgainstChained
	anchorID, cite := r.foundAgainstAnchor, r.foundAgainstCite
	r.mu.Unlock()
	if !due {
		return
	}
	if anchorID != "" {
		r.foundAgainstReread(seatID, anchorID)
		return
	}
	// ONLY WHERE BLUE IS NOT READIED: the lenses sit before blue in an epoch, so a reading in an epoch
	// blue-respond sits in is remedied before the owner sits again.
	if r.planReadies("blue-respond") {
		r.reachedFoundAgainstStage("every owner sitting so far was in a plan that readies blue-respond")
		return
	}
	// ONLY WHERE THE NEXT PLAN READIES THE OWNER BEFORE BLUE'S REMEDY: a lens served the whole report
	// at the head is ready again only when the head moves, and blue's remedy is what moves it.
	if why := r.ownerPinHeld(seatID); why != "" {
		r.reachedFoundAgainstStage(why)
		return
	}
	cite, classes, why := r.foundAgainstCitation(seatID, cite)
	if cite == "" {
		r.reachedFoundAgainstStage(why)
		return
	}
	run := r.foundAgainstReading(seatID, cite, classes)
	out, ok := run.admit()
	if !ok {
		r.reachedFoundAgainstStage("the refuting reading of " + cite + " was refused")
		return
	}
	if v := r.checkReading(seatID, out, run.args); v == nil {
		r.reachedFoundAgainstStage("the refuting reading of " + cite + " named no gap")
		return
	}
	f, held := r.ownFoundAgainstGapOpen(seatID, cite, "refutes")
	if !held {
		r.noteActErr(fmt.Sprintf("%s's refuting reading of %s, carried by a gap, left no standing found-against check of its own on an open gap", seatID, cite))
		return
	}
	// THE CLOSE IS REFUSED while the check stands.
	r.refuseStandingClose(r.minterOf(f.AnsweredBy()), f.AnsweredBy(), "repaired", f)
	r.mu.Lock()
	r.foundAgainstAnchor = cite
	r.foundAgainstStage = "found against " + cite + "; the owner has not sat again"
	r.mu.Unlock()
}

// foundAgainstHoldsPin is whether seatID's audit this sitting reads a selection of the report rather
// than the whole (auditRead): the forced run's citation owner, in a sitting its chain's reading may
// start in. Not served the whole report at the head, the owner is ready at the head again, so the plan
// after its reading readies it, and it sits before blue's remedy of the sentence.
func (r *runner) foundAgainstHoldsPin(seatID string) bool {
	r.mu.Lock()
	due := r.forceFoundAgainst && seatID == citationOwner && !r.foundAgainstChained && r.foundAgainstAnchor == ""
	r.mu.Unlock()
	return due && !r.planReadies("blue-respond")
}

// ownerPinHeld is "" where seatID's pin, as its work list states it now, leaves it ready at the next
// plan's head — never served the report whole, or last served it behind the head this sitting was
// dispatched against, with no whole read this sitting — and otherwise why the chain waits.
func (r *runner) ownerPinHeld(seatID string) string {
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil || w.Sitting.LastRead == nil {
		return fmt.Sprintf("the owner's work list could not be read to say whether the next plan readies it (%v)", err)
	}
	switch k := w.Sitting.LastRead.Kind; k {
	case "served", "unchanged":
		return fmt.Sprintf("the owner's pin stands %s at head %d, so the next plan readies blue's remedy before the owner sits again", k, w.Sitting.LastRead.Pin)
	}
	return ""
}

// foundAgainstPlanStored names, in the chain's stage, a plan the record stores while the chain waits
// for its owner's next sitting that readies blue and not the owner: blue's remedy lands first, and the
// stage carries the plan's reason for the owner.
func (r *runner) foundAgainstPlanStored() {
	r.mu.Lock()
	due := r.forceFoundAgainst && r.foundAgainstAnchor != "" && !r.foundAgainstChained
	anchorID := r.foundAgainstAnchor
	r.mu.Unlock()
	if !due || r.planReadies(citationOwner) || !r.planReadies("blue-respond") {
		return
	}
	why := "the plan states no reason for " + citationOwner
	whys, _ := r.planThisSitting["why"].([]any)
	for _, w := range whys {
		if s, _ := w.(string); strings.HasPrefix(s, citationOwner+":") {
			why = s
			break
		}
	}
	r.reachedFoundAgainstStage("found against " + anchorID + "; a plan readied blue and not the owner — " + why)
}

// scenarioOwnsClose is whether a gap minted with scenario d is one whose scenario oracle judges how it
// closes: a COMPUTE or COMPUTE-DRIFTS gap closed by red without its computation reproducing fails the
// run. The chain's close of its carrier answers no reproduction, so it never reads where one carries it.
func scenarioOwnsClose(d string) bool { return d == dirCompute || d == dirComputeDrifts }

// foundAgainstCitation is the citation the owner's chain reads as refuting at this sitting, and the gap
// classes that reading mints in: preferred — blue's citation of the chain's sentence — where it
// qualifies, else the first other citation on the record that does. A citation qualifies where its
// token stands once in the report, inside a sentence that holds no open gap whose scenario owns its
// close (scenarioOwnsClose) — the one-gap rule would hold the reading on that gap and the chain would
// close it unreproduced — where the owner holds no standing found-against check of it, where its
// reading mints in some class, and where the owner has not read it this sitting. Where none does,
// cite is "" and why says what held the candidates back: the chain waits for a later sitting.
func (r *runner) foundAgainstCitation(seatID, preferred string) (cite string, classes []string, why string) {
	b, ok := r.board()
	report, err := record.RenderRun(r.run())
	if !ok || err != nil {
		return "", nil, "the record or its report could not be read to choose the chain's citation"
	}
	spans, bare := anchor.Sentences(report), claimcount.BareAnchorIDs(report)
	owned := map[int]string{}
	for _, g := range b.Gaps {
		if g == nil || !g.Open || g.Mint == nil || !scenarioOwnsClose(g.Mint.GetRequiredFix()) {
			continue
		}
		if i, ok := sentenceIndexOf(report, spans, bare, g.ID); ok {
			owned[i] = g.ID
		}
	}
	candidates := []string{}
	if preferred != "" {
		candidates = append(candidates, preferred)
	}
	for _, s := range record.EvidenceJSONOf(b.Events, b.At, "", nil).Sources {
		if s.Anchor != "" && s.Anchor != preferred {
			candidates = append(candidates, s.Anchor)
		}
	}
	if len(candidates) == 0 {
		return "", nil, "no citation stood at an owner sitting blue was not readied in"
	}
	held := ""
	for _, c := range candidates {
		if strings.Count(report, anchor.Token(c)) != 1 {
			continue
		}
		i, ok := sentenceIndexOf(report, spans, bare, c)
		if !ok {
			continue
		}
		if g := owned[i]; g != "" {
			if held == "" {
				held = c + " stands in a sentence holding open gap " + g + ", whose scenario closes it only behind a reproduction"
			}
			continue
		}
		if _, standing := r.ownFoundAgainstGapOpen(seatID, c, "refutes"); standing {
			continue
		}
		cl, notTheWork := r.readingClasses(r.citationURL(c), "refutes", true)
		if notTheWork || len(cl) == 0 || !r.firstThisSitting(seatID, c) {
			continue
		}
		return c, cl, ""
	}
	if held != "" {
		return "", nil, "no citation the chain may read stood at an owner sitting blue was not readied in, so it waits for a later one: " + held
	}
	return "", nil, "no citation the chain may read stood at an owner sitting blue was not readied in"
}

// foundAgainstReading is seatID's refuting reading of citation anchorID with the gap flags of classes
// drawn as a drawn reading draws them, its check a document check.
func (r *runner) foundAgainstReading(seatID, anchorID string, classes []string) *cmd {
	gap, _ := r.readingGapArgs(classes)
	setArg(gap, "--check-kind", "document")
	c := r.do("verify", seatID).set("--anchor", anchorID).set("--as", "refutes").set("--confidence", "high").
		set("--reason", "fuzz: the source says otherwise than the sentence").bare("--json")
	c.args = append(c.args, gap...)
	return c
}

// foundAgainstReread is the chain's second half, at the owner's sitting after its reading of
// anchorID: the repeat refused, the close refused, and its own backing re-read, after which the
// close is admitted. Where blue's remedy came first and the check no longer stands, the chain starts
// over at a later sitting.
func (r *runner) foundAgainstReread(seatID, anchorID string) {
	if !r.firstThisSitting(seatID, anchorID) {
		return
	}
	f, held := r.ownFoundAgainstGapOpen(seatID, anchorID, "refutes")
	if !held {
		r.mu.Lock()
		r.foundAgainstAnchor, r.foundAgainstCite = "", ""
		r.foundAgainstStage = "found against " + anchorID + ", and the check no longer stood at the owner's next sitting (before it: " + r.foundAgainstStage + ")"
		r.mu.Unlock()
		return
	}
	classes, _ := r.readingClasses(r.citationURL(anchorID), "refutes", true)
	r.mu.Lock()
	repeats, cleared := r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading
	r.mu.Unlock()
	// THE REPEAT, with the same outcome over the sentence unchanged, is refused naming the reading.
	r.repeatReadingRefused(seatID, anchorID, "refutes", f, r.foundAgainstReading(seatID, anchorID, classes))
	r.refuseStandingClose(r.minterOf(f.AnsweredBy()), f.AnsweredBy(), "repaired", f)
	// ITS OWN BACKING RE-READ CLEARS IT, and the close is admitted.
	r.foundAgainstCleared(seatID, anchorID, f)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case r.foundAgainstRepeats == repeats:
		r.foundAgainstStage = "found against " + anchorID + "; the repeat was not refused"
	case r.foundAgainstClearedByOwnReading == cleared:
		r.foundAgainstStage = "found against " + anchorID + ", the repeat refused; no close followed the owner's backing re-read"
	default:
		r.foundAgainstStage = "chained over " + anchorID
		r.foundAgainstChained = true
	}
}

// forcedRepeatedReading is the citation owner's reading of one copy of the twice-held sentence in
// foundAgainstSeed's run, once its chain has completed, until one mints its gap with a --new for
// blue to accept.
func (r *runner) forcedRepeatedReading(seatID string) {
	r.mu.Lock()
	due := r.forceFoundAgainst && seatID == citationOwner && r.foundAgainstChained && !r.foundAgainstRepeatedRead
	pending := len(r.pendingAccepts)
	r.mu.Unlock()
	if !due {
		return
	}
	why := r.readRepeatedSentence(seatID, true)
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.pendingAccepts) > pending {
		r.foundAgainstRepeatedRead = true
		r.foundAgainstStage = "chained, and read one copy of the twice-held sentence with a --new"
		return
	}
	r.foundAgainstStage = repeatedReadingStage(r.repeatedCites, why)
}

// repeatedReadingStage is the chain's stage after the owner's forced reading of the twice-held
// sentence minted no gap with a --new: what readRepeatedSentence found instead (why), so the run's
// error names the path — a retired copy, an open gap, a refusal — and not that no reading minted.
func repeatedReadingStage(cites [2]string, why string) string {
	switch {
	case cites[0] == "":
		return "chained; blue has not made the twice-held sentence"
	case why == "":
		return "chained; a reading of a copy of the twice-held sentence was admitted, and no gap minted with a --new was recorded for blue to accept"
	}
	return "chained; the owner's reading of the twice-held sentence minted no gap with a --new: " + why
}

// foundAgainstForcedErr is the run's error where foundAgainstSeed's run did not drive each path it is
// forced through, naming the last step the chain reached; "" otherwise and on every other run. The
// caller holds r.mu: runOne reads it beside the other forced-run errors under the run's lock.
func (r *runner) foundAgainstForcedErr() string {
	if !r.forceFoundAgainst || r.foundAgainstRepeats > 0 && r.foundAgainstClearedByOwnReading > 0 && r.repeatedSentenceAccepts > 0 {
		return ""
	}
	stage := r.foundAgainstStage
	if stage == "" {
		stage = "nothing: the citation owner never sat"
	}
	return fmt.Sprintf("forced found-against chain: %d repeat(s) refused, %d close(s) after the owner's own re-read, %d accept(s) on the twice-held sentence; the chain reached %s",
		r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading, r.repeatedSentenceAccepts, stage)
}

// forcedErrUnderTheRunLock reads foundAgainstForcedErr the way runOne does, holding r.mu, and fails
// rather than hangs if the read takes the lock again.
func forcedErrUnderTheRunLock(t *testing.T, r *runner) string {
	t.Helper()
	got := make(chan string, 1)
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		got <- r.foundAgainstForcedErr()
	}()
	select {
	case msg := <-got:
		return msg
	case <-time.After(10 * time.Second):
		t.Fatal("foundAgainstForcedErr did not return under the run's lock: it takes r.mu, which runOne already holds")
		return ""
	}
}

// citeAs is blue's citation of quote, by a source at path, and its label.
func (r *runner) citeAs(t *testing.T, seatID, quote, path string) string {
	t.Helper()
	out, ok := r.do("cite", seatID).set("--quote", quote).set("--url", sourceURL(path)).set("--title", sourceTitle).
		set("--source-span", sourceBody(path)).bare("--json").admit()
	var env struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.Label == "" {
		t.Fatalf("blue's cite of %q: %s (%s)", quote, out, r.actErr)
	}
	return env.Result.Label
}

// THE FORCED FOUND-AGAINST RUN DRIVES EACH PATH IT IS FORCED THROUGH, run outside the release gate on
// the drive's own sittings: the citation owner's chain over one citation (the repeat refused, the
// close refused, its own re-read and the close admitted), blue's twice-held sentence, the owner's
// reading of one copy with a --new, and blue's accept of it. Each rng seed draws the chain's coins
// another way (another lens's backing reading first, or not). With the run not forced, none of it
// is driven and the run owes nothing.
func TestTheFoundAgainstSeedDrivesEachPathItIsForcedThrough(t *testing.T) {
	awaitsPhase1e(t)
	for _, seed := range []int64{1, 2, 3, 4} {
		t.Run(fmt.Sprint("rng ", seed), func(t *testing.T) {
			r := stagedRunner(t)
			r.rng = newLockedRand(seed)
			r.forceFoundAgainst = true
			if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
				t.Fatalf("record the lanes' parts: %v", err)
			}
			const blue = "blue-respond"
			if msg := forcedErrUnderTheRunLock(t, r); !strings.Contains(msg, "never sat") {
				t.Fatalf("a forced run nothing has driven reads %q, want the stage it never reached", msg)
			}
			r.sit("blue", blue)
			r.citeForFoundAgainst(blue)
			if r.mintsFresh(citationOwner) {
				t.Fatal("the forced run's citation owner mints fresh before its chain")
			}
			// ITS READING, and at its next sitting the re-read: a sitting holds one reading per citation.
			r.sit("lens", citationOwner)
			r.forcedFoundAgainstChain(citationOwner)
			if r.actErr != "" || r.foundAgainstAnchor == "" || r.foundAgainstClosesRefused != 1 {
				t.Fatalf("the chain's first sitting reached %q (closes refused %d): %s", r.foundAgainstStage, r.foundAgainstClosesRefused, r.actErr)
			}
			r.sit("lens", citationOwner)
			r.forcedFoundAgainstChain(citationOwner)
			if r.actErr != "" {
				t.Fatalf("the chain: %s", r.actErr)
			}
			if !r.foundAgainstChained || r.foundAgainstRepeats != 1 || r.foundAgainstClosesRefused < 1 || r.foundAgainstClearedByOwnReading != 1 {
				t.Fatalf("the chain reached %q: repeats %d, closes refused %d, cleared %d — want 1, at least 1, 1",
					r.foundAgainstStage, r.foundAgainstRepeats, r.foundAgainstClosesRefused, r.foundAgainstClearedByOwnReading)
			}
			// THE CHAIN'S READING PASSES NO --new: blue's accept of it edits the sentence, which clears
			// the check the chain's next sitting reads.
			if r.ownerMintedVerified() {
				t.Error("the chain's reading minted a gap with fix_basis verified — it passed a --new, which blue's accept applies to the chain's sentence")
			}
			// BLUE MAKES THE TWICE-HELD SENTENCE; the owner reads one copy with a --new; blue accepts it.
			r.sit("blue", blue)
			r.stockRepeatedSentence(blue)
			r.sit("lens", citationOwner)
			r.forcedRepeatedReading(citationOwner)
			if !r.foundAgainstRepeatedRead || !r.mintsFresh(citationOwner) {
				t.Fatalf("the owner's reading of the twice-held sentence reached %q (pending %d)", r.foundAgainstStage, len(r.pendingAccepts))
			}
			// THE CHAIN'S ONE --new is on the twice-held step's reading.
			if !r.ownerMintedVerified() {
				t.Error("the owner's reading of the twice-held sentence minted no gap with fix_basis verified — it passed no --new")
			}
			r.sit("blue", blue)
			r.acceptRepeated(blue)
			if r.actErr != "" {
				t.Fatalf("the drive: %s", r.actErr)
			}
			if msg := forcedErrUnderTheRunLock(t, r); msg != "" {
				t.Errorf("the forced run drove every path and still reads %q", msg)
			}
			if r.repeatedSentenceGaps != 1 || r.repeatedSentenceAccepts != 1 {
				t.Errorf("twice-held gaps %d, accepts %d — want 1 and 1", r.repeatedSentenceGaps, r.repeatedSentenceAccepts)
			}
		})
	}
	t.Run("not forced", func(t *testing.T) {
		r := stagedRunner(t)
		if err := r.stagePart(lanePartText("blue-lane-2")); err != nil {
			t.Fatal(err)
		}
		r.sit("blue", "blue-respond")
		r.citeForFoundAgainst("blue-respond")
		r.citeAs(t, "blue-respond", "The cost is rising over time", "/unforced")
		for i := 0; i < 2; i++ {
			r.sit("lens", citationOwner)
			r.forcedFoundAgainstChain(citationOwner)
			r.forcedRepeatedReading(citationOwner)
		}
		if r.foundAgainstCite != "" {
			t.Errorf("a run not forced cited the chain's sentence (%s)", r.foundAgainstCite)
		}
		if r.foundAgainstRepeats+r.foundAgainstClearedByOwnReading != 0 || forcedErrUnderTheRunLock(t, r) != "" {
			t.Errorf("a run not forced drove the chain (%d repeats, %d cleared) or owes it (%q)", r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading, forcedErrUnderTheRunLock(t, r))
		}
	})
}

// THE CHAIN'S OWNER SITS AGAIN BEFORE BLUE'S REMEDY, on the plans the record relays: the plan after the
// chain's reading readies blue on its standing check and its gap, and readies the gap's lens only once
// blue has acted there, so the owner sits first only where the plan readies it on its pin. The
// sitting the chain reads in serves the owner a selection of the report (foundAgainstHoldsPin), so the
// next plan readies it beside blue; the lenses sit before blue, and the chain completes at that
// sitting. Blue then sits on what the plan engaged it on — its response to the gap and its remedy of
// any check — which leaves the completed chain as it stands. The chain's reading passes no --new; the
// twice-held step's reading passes one, and blue's next sitting accepts it.
func TestTheFoundAgainstChainSitsAgainBeforeBlue(t *testing.T) {
	awaitsPhase1e(t)
	r := stagedRunner(t)
	r.forceFoundAgainst = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, chair = "blue-respond", "red-chair"
	plan := func() {
		t.Helper()
		r.sit("chair", chair)
		r.planThisSitting = r.dispatchNext(chair)
		r.foundAgainstPlanStored()
	}
	blueSits := func() {
		t.Helper()
		r.sit("blue", blue)
		r.stockRepeatedSentence(blue)
		r.acceptRepeated(blue)
		r.blueRespondTo(blue, r.engaged(blue))
		r.fixFailed(blue)
		if r.actErr != "" {
			t.Fatalf("blue's sitting: %s", r.actErr)
		}
	}
	ownerSits := func() {
		t.Helper()
		r.sit("lens", citationOwner)
		r.auditRead(citationOwner)
		r.forcedFoundAgainstChain(citationOwner)
		r.forcedRepeatedReading(citationOwner)
		if r.actErr != "" {
			t.Fatalf("the owner's sitting: %s", r.actErr)
		}
	}
	r.sit("blue", blue)
	r.citeForFoundAgainst(blue)
	if r.foundAgainstCite == "" {
		t.Fatalf("blue's chain citation: %s (%s)", r.foundAgainstStage, r.actErr)
	}

	// THE CHAIN'S READING, in a plan that readies the owner and not blue.
	plan()
	if !r.planReadies(citationOwner) || r.planReadies(blue) {
		t.Fatalf("the first plan readies the owner %v and blue %v — want the owner alone: %v", r.planReadies(citationOwner), r.planReadies(blue), r.planThisSitting["why"])
	}
	ownerSits()
	if r.foundAgainstAnchor == "" || r.foundAgainstClosesRefused != 1 {
		t.Fatalf("the chain's first sitting reached %q (closes refused %d)", r.foundAgainstStage, r.foundAgainstClosesRefused)
	}
	if r.ownerMintedVerified() {
		t.Fatal("the chain's reading passed a --new: blue's accept of it edits the chain's sentence")
	}

	// THE NEXT PLAN READIES BLUE ON THE CHECK, AND THE OWNER BESIDE IT: the owner sits first.
	plan()
	if !r.planReadies(blue) {
		t.Fatalf("the plan after the chain's reading does not ready blue on its standing check: %v", r.planThisSitting["why"])
	}
	if !r.planReadies(citationOwner) {
		t.Fatalf("the plan after the chain's reading readies blue and not the owner, so blue's remedy lands first (stage %q)", r.foundAgainstStage)
	}
	ownerSits()
	if !r.foundAgainstChained || r.foundAgainstRepeats != 1 || r.foundAgainstClearedByOwnReading != 1 {
		t.Fatalf("the chain reached %q: repeats %d, cleared %d — want 1 and 1", r.foundAgainstStage, r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading)
	}
	blueSits()
	if !r.foundAgainstChained || r.foundAgainstRepeats != 1 || r.foundAgainstClearedByOwnReading != 1 {
		t.Fatalf("blue's sitting after the chain undid it: %q, repeats %d, cleared %d", r.foundAgainstStage, r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading)
	}

	// THE TWICE-HELD STEP: the owner's reading of one copy passes a --new, and blue accepts it.
	for i := 0; i < 3 && r.repeatedSentenceAccepts == 0; i++ {
		plan()
		if r.planReadies(citationOwner) {
			ownerSits()
		}
		blueSits()
	}
	if !r.foundAgainstRepeatedRead || !r.ownerMintedVerified() || r.repeatedSentenceAccepts != 1 {
		t.Fatalf("the twice-held step reached %q (verified mint %v, accepts %d)", r.foundAgainstStage, r.ownerMintedVerified(), r.repeatedSentenceAccepts)
	}
	if msg := forcedErrUnderTheRunLock(t, r); msg != "" {
		t.Errorf("the forced run drove every path and still reads %q", msg)
	}
}

// A CLOSE A SIBLING'S READING REFUSES BETWEEN THE DRIVE'S READ AND THE ACT IS THE REFUSAL IT IS (seeds 5
// and 6 of 2ecae7a9b's sweep): lens L's gap G stands on a cited sentence; L's close of G reads no
// standing check on it, and before the close lands the citation owner's refuting reading of that
// sentence is held by the one-gap rule and carried by G. The record refuses the close naming the
// reading; the drive tallies it ONCE as a close refused over a standing found-against check, not as an
// admitted act refused — with or without another gap named as what the close was verified with.
// plainHeldClass is the class these drives give a plain gap a refuting reading is later held by: a
// class a refuting reading admits whose mint stands on no reading. A claim_contradicts_source mint
// stands on the minter's own reading of the sentence (record.readingMintedClasses), so a plain mint of
// it with no reading behind it is refused.
const plainHeldClass = "within_source_condition_misattribution"

func TestACloseASiblingsReadingRefusesInTheWindowIsTallied(t *testing.T) {
	awaitsPhase1e(t)
	for _, anchored := range []bool{false, true} {
		t.Run(map[bool]string{false: "plain", true: "verified with another gap"}[anchored], func(t *testing.T) {
			r := stagedRunner(t)
			if err := r.stagePart(lanePartText("blue-lane-2")); err != nil {
				t.Fatal(err)
			}
			const blue, lens = "blue-respond", "red-lens-logic"
			r.sit("blue", blue)
			cite := r.citeAs(t, blue, "The cost is rising over time", "/window")
			r.sit("lens", lens)
			const class = plainHeldClass
			mint := func(quote, class string) string {
				t.Helper()
				out, ok := r.admit("--json", "mint", "--seat-id", lens, "--quote", quote, "--class", class,
					"--problem", "fuzz: the sentence overstates the trend", "--check-kind", "document", "--check", "acc",
					"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
				var env struct {
					Result struct {
						GapID string `json:"gap_id"`
					} `json:"result"`
				}
				if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
					t.Fatalf("L's mint: %s (%s)", out, r.actErr)
				}
				r.noteMint(lens, env.Result.GapID)
				return env.Result.GapID
			}
			g := mint("rising over time", class)
			anchor := ""
			if anchored {
				anchor = mint("The cost is", "spec_underspecification")
			}
			if _, ok := r.standingFailedCheckOf(g); ok {
				t.Fatalf("gap %s carries a standing check before the reading", g)
			}
			r.betweenCloseReadAndAct = func(id string) {
				r.sit("lens", citationOwner)
				out, ok := r.do("verify", citationOwner).set("--anchor", cite).set("--as", "refutes").set("--confidence", "high").
					set("--reason", "fuzz: the source says the cost fell").bare("--json").
					set("--class", class).set("--problem", "fuzz: the source says the cost fell").
					set("--check-kind", "document").set("--check", "the sentence says what the source says").
					set("--severity", "high").set("--likelihood", "high").set("--impact", "high").admit()
				res, parsed := readingResultOf(out)
				if !ok || !parsed || res.Gap == nil || res.Gap.ID != id || res.Gap.Minted {
					t.Fatalf("the owner's reading was not held by %s: %s (%s)", id, out, r.actErr)
				}
				r.betweenCloseReadAndAct = nil
			}
			r.closeVerifiedWith(g, []string{"close", "--seat-id", lens, "--id", g, "--as", "repaired", "--reason", "fuzz close as repaired",
				"--verified-by", lens, "--verified-with", "fuzz", "--verified-against", "rec"}, anchor)
			if !slices.Contains(r.openGaps(), g) {
				t.Fatalf("the close of %s was admitted over the owner's reading it carries", g)
			}
			if r.actErr != "" {
				t.Fatalf("the refusal was read as the drive's failure: %s", r.actErr)
			}
			if r.foundAgainstClosesRefused != 1 {
				t.Errorf("found-against closes refused = %d, want 1: one close, one refusal", r.foundAgainstClosesRefused)
			}
		})
	}
}

// ownerMintedVerified reports whether the citation owner has minted a gap whose fix_basis is verified:
// one whose reading passed a --new the record validated.
func (r *runner) ownerMintedVerified() bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	for _, e := range b.Events {
		if m, isM := recordpb.BodyAs[*recordpb.Mint](e); isM && e.GetSeatId() == citationOwner && m.GetFixBasis() == "verified" {
			return true
		}
	}
	return false
}

// THE CHAIN READS NO CITATION WHOSE SENTENCE HOLDS A GAP ITS SCENARIO CLOSES, and reads the first clean
// one once it stands. A lens's COMPUTE-scenario gap stands in the sentence blue's chain citation sits
// in: the owner's refuting reading there would be held by that gap, and the chain's close of its
// carrier — no reproduction behind it — is the close the COMPUTE oracle fails the run on. The chain
// waits; once blue cites another sentence, it reads that one, and the COMPUTE gap is never closed.
func TestTheFoundAgainstChainReadsNoCitationAScenarioGapStandsAt(t *testing.T) {
	awaitsPhase1e(t)
	r := stagedRunner(t)
	r.forceFoundAgainst = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, lens = "blue-respond", "red-lens-logic"
	r.sit("blue", blue)
	r.citeForFoundAgainst(blue)
	cite := r.foundAgainstCite
	if cite == "" {
		t.Fatalf("blue's chain citation: %s (%s)", r.foundAgainstStage, r.actErr)
	}
	classes, _ := r.readingClasses(r.citationURL(cite), "refutes", true)
	if !slices.Contains(classes, plainHeldClass) {
		t.Fatalf("a refuting reading of the chain's citation mints in %v, not %s", classes, plainHeldClass)
	}
	r.sit("lens", lens)
	out, ok := r.admit("--json", "mint", "--seat-id", lens, "--quote", "rising over time", "--class", plainHeldClass,
		"--problem", "fuzz: the trend's figure is computed wrong", "--check-kind", "computation", "--check", "the figure recomputes",
		"--fix", dirCompute, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		t.Fatalf("the COMPUTE gap's mint: %s (%s)", out, r.actErr)
	}
	g := env.Result.GapID
	r.noteMint(lens, g)
	open := func() bool { return slices.Contains(r.openGaps(), g) }
	readsOf := func(anchorID string) bool {
		_, _, read := r.storedReading(citationOwner, anchorID, "", "")
		return read
	}

	// THE ONLY CITATION STANDS AT THE COMPUTE GAP: the chain waits, reading nothing.
	r.sit("lens", citationOwner)
	r.forcedFoundAgainstChain(citationOwner)
	if r.actErr != "" {
		t.Fatalf("the chain: %s", r.actErr)
	}
	if readsOf(cite) || r.foundAgainstAnchor != "" {
		t.Fatalf("the chain read %s, whose sentence holds COMPUTE gap %s (stage %q)", cite, g, r.foundAgainstStage)
	}
	if !open() {
		t.Fatalf("COMPUTE gap %s was closed by the chain", g)
	}
	if !strings.Contains(r.foundAgainstStage, g) || !strings.Contains(r.foundAgainstStage, "waits") {
		t.Errorf("the chain's stage %q does not say it waits on %s", r.foundAgainstStage, g)
	}

	// A CLEAN CITATION STANDS: the chain reads it, and at the owner's next sitting completes over it.
	r.sit("blue", blue)
	clean := r.citeAs(t, blue, "the rise holds in every year measured", "/clean")
	r.sit("lens", citationOwner)
	r.forcedFoundAgainstChain(citationOwner)
	if r.actErr != "" || r.foundAgainstAnchor != clean {
		t.Fatalf("the chain read %q, want the clean citation %s (stage %q): %s", r.foundAgainstAnchor, clean, r.foundAgainstStage, r.actErr)
	}
	r.sit("lens", citationOwner)
	r.forcedFoundAgainstChain(citationOwner)
	if r.actErr != "" || !r.foundAgainstChained || r.foundAgainstClearedByOwnReading != 1 {
		t.Fatalf("the chain over %s reached %q (cleared %d): %s", clean, r.foundAgainstStage, r.foundAgainstClearedByOwnReading, r.actErr)
	}
	if readsOf(cite) {
		t.Errorf("the chain read %s, whose sentence holds COMPUTE gap %s", cite, g)
	}
	if !open() {
		t.Errorf("COMPUTE gap %s was closed by the chain", g)
	}
}

// A DRAWN BACKING RE-READ CLEARS ITS CHECK AND LEAVES A COMPUTE CARRIER TO ITS SCENARIO: a lens's refuting
// reading held by the one-gap rule on a COMPUTE-scenario gap, cleared at its next sitting by its own
// backing re-read (readCitation's coin), closes nothing — the gap closes only behind a reproduction.
func TestABackingReReadLeavesAComputeCarrierOpen(t *testing.T) {
	awaitsPhase1e(t)
	r := stagedRunner(t)
	// LANE 1 HOLDS THE SENTENCE blue's computation of the carrier quotes (answerEvidenceDemand).
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2")); err != nil {
		t.Fatal(err)
	}
	const blue, lens = "blue-respond", "red-lens-logic"
	r.sit("blue", blue)
	cite := r.citeAs(t, blue, "The cost is rising over time", "/compute-carrier")
	classes, _ := r.readingClasses(r.citationURL(cite), "refutes", true)
	if !slices.Contains(classes, plainHeldClass) {
		t.Fatalf("a refuting reading of the citation mints in %v, not %s", classes, plainHeldClass)
	}
	r.sit("lens", lens)
	out, ok := r.admit("--json", "mint", "--seat-id", lens, "--quote", "across the whole sampled corpus", "--class", plainHeldClass,
		"--problem", "fuzz: the trend's figure is computed wrong", "--check-kind", "computation", "--check", "the figure recomputes",
		"--fix", dirCompute, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		t.Fatalf("the COMPUTE gap's mint: %s (%s)", out, r.actErr)
	}
	g := env.Result.GapID
	r.noteMint(lens, g)
	r.sit("lens", citationOwner)
	if out, ok := r.foundAgainstReading(citationOwner, cite, []string{plainHeldClass}).admit(); !ok {
		t.Fatalf("the owner's refuting reading: %s (%s)", out, r.actErr)
	}
	f, held := r.ownFoundAgainstGapOpen(citationOwner, cite, "refutes")
	if !held || f.AnsweredBy() != g {
		t.Fatalf("the owner's reading is not held by COMPUTE gap %s (held %v, carrier %q)", g, held, f.AnsweredBy())
	}
	r.sit("lens", citationOwner)
	r.foundAgainstCleared(citationOwner, cite, f)
	if r.actErr != "" {
		t.Fatalf("the re-read: %s", r.actErr)
	}
	if _, still := r.checkCarriedBy(g, cite, citationOwner); still {
		t.Errorf("the owner's backing re-read left its check standing on %s", g)
	}
	if !slices.Contains(r.openGaps(), g) || r.foundAgainstClearedByOwnReading != 0 {
		t.Errorf("COMPUTE gap %s was closed by the re-read (open %v, cleared %d)", g, slices.Contains(r.openGaps(), g), r.foundAgainstClearedByOwnReading)
	}
}

// TestAForcedChainsFirstPlanReadiesItsOwnerAndNotBlue holds the window both forced chains that want blue
// absent start in: the first plan, after the synthesis, readies the chain's owner and not blue-respond.
// An inquiry the synthesis leaves open readies blue-respond for as long as it is open, and every later
// plan readies blue on the gaps the first epoch mints, so a chain that missed the first plan reached its
// epoch limit with every owner sitting in a plan that readies blue (seeds 6 and 15). The inquiry here is
// proposed and left open, as the synthesizer's drawn proposal leaves one.
func TestAForcedChainsFirstPlanReadiesItsOwnerAndNotBlue(t *testing.T) {
	for _, c := range []struct {
		name, owner string
		force       func(*runner)
	}{
		{"unsound computation", computationOwner, func(r *runner) { r.forceUnsoundComputation = true }},
		{"found against", citationOwner, func(r *runner) { r.forceFoundAgainst = true }},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := stagedRunner(t)
			c.force(r)
			if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
				t.Fatalf("record the lanes' parts: %v", err)
			}
			const synth, chair = "blue-synthesize", "red-chair"
			if _, ok := r.do("inquiry propose", synth).set("--question", "fuzz inquiry left open").
				set("--hypothesis", "fuzz: what would be true if it paid off").admit(); !ok {
				t.Fatalf("propose the open inquiry: %s", r.actErr)
			}
			r.envelopeFor(synth, "")
			if r.actErr != "" {
				t.Fatalf("the synthesizer's sitting: %s", r.actErr)
			}
			r.sit("chair", chair)
			r.planThisSitting = r.dispatchNext(chair)
			if r.planReadies("blue-respond") || !r.planReadies(c.owner) {
				t.Fatalf("the first plan readies blue-respond %v and %s %v — want the owner alone: %v",
					r.planReadies("blue-respond"), c.owner, r.planReadies(c.owner), r.planThisSitting["why"])
			}
		})
	}
}

// A RETIRED COPY OF THE TWICE-HELD SENTENCE IS NAMED BY THE STAGE: blue retires one copy's citation
// (what fixFailed does for a citation no check could read), and the owner's forced reading of the
// sentence reads nothing at every later sitting — a retired citation's token stands nowhere in the
// report. The run's error says which citation, and that it stands no times, where one message once
// stood for every path that read nothing.
func TestTheForcedRepeatedReadingNamesTheRetiredCopy(t *testing.T) {
	r := stagedRunner(t)
	r.forceFoundAgainst = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue = "blue-respond"
	r.sit("blue", blue)
	r.stockRepeatedSentence(blue)
	if r.repeatedCites[0] == "" || r.repeatedCites[1] == "" {
		t.Fatalf("blue did not make the twice-held sentence (%s)", r.actErr)
	}
	retired := r.repeatedCites[0]
	if _, ok := r.do("retire", blue).set("--anchor", retired).set("--reason", "fuzz: no lens could read its source").admit(); !ok {
		t.Fatalf("blue's retire of %s: %s", retired, r.actErr)
	}
	// THE CHAIN IS COMPLETE; the owner's next duty is the twice-held sentence.
	r.foundAgainstChained = true
	r.sit("lens", citationOwner)
	r.forcedRepeatedReading(citationOwner)
	if r.actErr != "" {
		t.Fatalf("the owner's sitting: %s", r.actErr)
	}
	if r.foundAgainstRepeatedRead || len(r.pendingAccepts) != 0 {
		t.Fatalf("the owner read a copy of the sentence with a --new over a retired citation (stage %q)", r.foundAgainstStage)
	}
	msg := forcedErrUnderTheRunLock(t, r)
	for _, want := range []string{"the chain reached chained; the owner's reading of the twice-held sentence minted no gap with a --new: ", "citation " + retired + "'s token stands 0 time(s)"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the forced run's error %q does not say %q", msg, want)
		}
	}
	if strings.Contains(msg, "no reading of a copy of the twice-held sentence minted its gap") {
		t.Errorf("the forced run's error %q names no path", msg)
	}
}
