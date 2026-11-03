package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
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
// finds against it and mints the gap, and the gap's close is refused while the check stands. The
// lenses sit before blue in an epoch, so the owner's next sitting comes before blue's remedy: there its
// repeat with the same outcome is refused naming that reading, the close is refused again, and its own
// backing re-read clears the check, after which the close is admitted (foundAgainstCleared). A sitting
// holds one reading per citation, so the re-read is never the reading's own sitting. Once blue has made
// the twice-held sentence, the owner's reading of one copy finds against it with a --new
// (readRepeatedSentence), and blue's next sitting accepts that fix (acceptRepeated). Every lane
// records its part, so the report holds the sentence twice.
const foundAgainstSeed = 15

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
	if cur, err := reportproj.RenderFromRecord(r.run()); err != nil || strings.Count(plainText(cur), foundAgainstQuote) != 1 {
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
// the reading that finds against a citation, then, at its next sitting, the repeat, the refused close
// and its own re-read. It draws the reading's gap flags as the drawn reading does, and pins the check
// kind to a document check, so the close needs no computation.
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
	if cite == "" || !r.firstThisSitting(seatID, cite) {
		r.reachedFoundAgainstStage("no citation of the chain's sentence stood at an owner sitting blue was not readied in")
		return
	}
	classes, notTheWork := r.readingClasses(r.citationURL(cite), "refutes", true)
	if notTheWork || len(classes) == 0 {
		r.reachedFoundAgainstStage("the chain's citation reads as no gap class (" + cite + ")")
		return
	}
	run := r.foundAgainstReading(seatID, cite, classes)
	// --new WHERE THE SENTENCE GIVES ONE: the flag's own coverage rides on this reading too.
	if fixNew := r.citationFixNew(cite); fixNew != "" {
		run = &cmd{r: r, args: append(slices.Clone(run.args), "--new", fixNew)}
	}
	out, ok := run.admit()
	if !ok {
		r.reachedFoundAgainstStage("the refuting reading of " + cite + " was refused")
		return
	}
	if v := r.checkReading(seatID, out, run.args); v == nil || v.GetCarriedBy() == "" {
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
		r.foundAgainstStage = "found against " + anchorID + ", and the check no longer stood at the owner's next sitting"
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
	r.readRepeatedSentence(seatID, true)
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.pendingAccepts) > pending {
		r.foundAgainstRepeatedRead = true
		r.foundAgainstStage = "chained, and read one copy of the twice-held sentence with a --new"
	} else if r.repeatedCites[0] == "" {
		r.foundAgainstStage = "chained; blue has not made the twice-held sentence"
	} else {
		r.foundAgainstStage = "chained; no reading of a copy of the twice-held sentence minted its gap with a --new"
	}
}

// foundAgainstForcedErr is the run's error where foundAgainstSeed's run did not drive each path it is
// forced through, naming the last step the chain reached; "" otherwise and on every other run.
func (r *runner) foundAgainstForcedErr() string {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	for _, seed := range []int64{1, 2, 3, 4} {
		t.Run(fmt.Sprint("rng ", seed), func(t *testing.T) {
			r := stagedRunner(t)
			r.rng = newLockedRand(seed)
			r.forceFoundAgainst = true
			if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
				t.Fatalf("record the lanes' parts: %v", err)
			}
			const blue = "blue-respond"
			if msg := r.foundAgainstForcedErr(); !strings.Contains(msg, "never sat") {
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
			if !r.ownerMintedVerified() {
				t.Error("the chain's reading minted no gap with fix_basis verified — it passed no --new")
			}
			// BLUE MAKES THE TWICE-HELD SENTENCE; the owner reads one copy with a --new; blue accepts it.
			r.sit("blue", blue)
			r.stockRepeatedSentence(blue)
			r.sit("lens", citationOwner)
			r.forcedRepeatedReading(citationOwner)
			if !r.foundAgainstRepeatedRead || !r.mintsFresh(citationOwner) {
				t.Fatalf("the owner's reading of the twice-held sentence reached %q (pending %d)", r.foundAgainstStage, len(r.pendingAccepts))
			}
			r.sit("blue", blue)
			r.acceptRepeated(blue)
			if r.actErr != "" {
				t.Fatalf("the drive: %s", r.actErr)
			}
			if msg := r.foundAgainstForcedErr(); msg != "" {
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
		if r.foundAgainstRepeats+r.foundAgainstClearedByOwnReading != 0 || r.foundAgainstForcedErr() != "" {
			t.Errorf("a run not forced drove the chain (%d repeats, %d cleared) or owes it (%q)", r.foundAgainstRepeats, r.foundAgainstClearedByOwnReading, r.foundAgainstForcedErr())
		}
	})
}

// A CLOSE A SIBLING'S READING REFUSES BETWEEN THE DRIVE'S READ AND THE ACT IS THE REFUSAL IT IS (seeds 5
// and 6 of 2ecae7a9b's sweep): lens L's gap G stands on a cited sentence; L's close of G reads no
// standing check on it, and before the close lands the citation owner's refuting reading of that
// sentence is held by the one-gap rule and carried by G. The record refuses the close naming the
// reading; the drive tallies it as a close refused over a standing found-against check, not as an
// admitted act refused.
func TestACloseASiblingsReadingRefusesInTheWindowIsTallied(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-2")); err != nil {
		t.Fatal(err)
	}
	const blue, lens = "blue-respond", "red-lens-logic"
	r.sit("blue", blue)
	cite := r.citeAs(t, blue, "The cost is rising over time", "/window")
	r.sit("lens", lens)
	const class = "claim_contradicts_source"
	out, ok := r.admit("--json", "mint", "--seat-id", lens, "--quote", "rising over time", "--class", class,
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
	g := env.Result.GapID
	r.noteMint(lens, g)
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
	}
	if r.admitClose(g, []string{"close", "--seat-id", lens, "--id", g, "--as", "repaired", "--reason", "fuzz close as repaired",
		"--verified-by", lens, "--verified-with", "fuzz", "--verified-against", "rec"}) {
		t.Fatalf("the close of %s was admitted over the owner's reading it carries", g)
	}
	if r.actErr != "" {
		t.Fatalf("the refusal was read as the drive's failure: %s", r.actErr)
	}
	if r.foundAgainstClosesRefused != 1 {
		t.Errorf("found-against closes refused = %d, want 1", r.foundAgainstClosesRefused)
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
