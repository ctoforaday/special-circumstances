package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// readingDebtSeed is the seed whose run is FORCED through a reading debt's whole life on a cited
// source (forceReadingDebt), where a drawn run reaches the parts of it only on a chain of coins:
// the citation owner's reading of a citation blue made finds against it and the owner mints
// nothing — the debt refuses its envelope and its sitting ends without one; the next plan readies
// the owner for the debt and refuses the chair's PASS naming it; at the owner's next sitting its
// claim_contradicts_source mint about the citation stands on the reading and discharges the debt,
// and the gap's `source` demand refuses its close; blue's retire of the citation answering the gap
// meets the demand through the gap's `about` arm; and the owner's close is admitted. Each step reads
// the record, never a count of sittings; the stage it reached names the run's error.
//
// The chain's citation is lane 1's closing sentence, which no drawn edit rewrites (blue's edit drive
// swaps a phrasing in lane 2's sentence), so blue's edit never discharges the debt before the owner's
// mint; the drawn drives leave the citation to the chain (reservedDebtSubject, chainCitation).

// debtChainQuote is the words of lane 1's closing sentence blue's forced citation quotes.
const debtChainQuote = "costs rising steeply"

// Stages of readingDebtSeed's chain (debtChainStage), in order.
const (
	debtChainCited   = "cited: blue cited the chain's sentence"
	debtChainRead    = "read: the owner's reading finds against the citation, the debt stands and the owner's envelope was refused"
	debtChainMinted  = "minted: the owner's gap about the citation stands on its reading, and the gap's demand refuses its close"
	debtChainRetired = "retired: blue's retire of the citation answering the gap meets its demand"
	debtChainClosed  = "closed: the owner's close of the gap was admitted"
)

// debtChainMintedYet is whether the forced run's citation owner has minted the gap its chain is
// held to, after which it mints as any lens.
func (r *runner) debtChainMintedYet() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.debtChainMinted
}

func (r *runner) reachedDebtChainStage(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debtChainStage = s
}

// chainCitation is the forced chain's citation while the chain reserves it, "" otherwise: the drawn
// citation reading leaves it to the chain.
func (r *runner) chainCitation() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.forceReadingDebt && !r.debtChainMinted {
		return r.debtChainCite
	}
	return ""
}

// citeForDebtChain is blue's citation, at its first sitting in readingDebtSeed's run, of the
// sentence the owner's chain reads against: a citation stands before the first lens sitting.
func (r *runner) citeForDebtChain(seatID string) {
	r.mu.Lock()
	due := r.forceReadingDebt && r.debtChainCite == "" && !r.debtChainCiteTried
	r.mu.Unlock()
	if !due {
		return
	}
	// A REPORT THAT DOES NOT YET HOLD THE WORDS ONCE IS CITED AT A LATER BLUE SITTING.
	if cur, err := record.RenderRun(r.run()); err != nil || strings.Count(plainText(cur), debtChainQuote) != 1 {
		r.reachedDebtChainStage("no blue sitting so far found the chain's sentence once in the report")
		return
	}
	r.mu.Lock()
	r.debtChainCiteTried = true
	r.mu.Unlock()
	const path = "/debt-chain"
	out, ok := r.do("cite", seatID).set("--quote", debtChainQuote).set("--url", sourceURL(path)).set("--title", sourceTitle).
		set("--source-span", sourceBody(path)).bare("--json").admit()
	if !ok {
		r.reachedDebtChainStage("blue's citation of the chain's sentence was refused")
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
	r.debtChainCite = env.Result.Label
	r.debtChainStage = debtChainCited
	r.mu.Unlock()
}

// forcedDebtChain is the citation owner's half of the chain, at each of its sittings: the
// reading that finds against, left standing; at the next sitting the mint on it and the refused
// close; once blue has retired the citation, the close. It draws nothing from the run's RNG.
func (r *runner) forcedDebtChain(seatID string) {
	r.mu.Lock()
	due := r.forceReadingDebt && seatID == citationOwner && !r.debtChainDone
	cite, stage, gap := r.debtChainCite, r.debtChainStage, r.debtChainGap
	r.mu.Unlock()
	if !due {
		return
	}
	switch stage {
	case debtChainCited:
		if r.readCited(seatID, cite, "refutes") == 0 {
			r.reachedDebtChainStage("the refuting reading of " + cite + " was refused")
			return
		}
		if _, stands := r.debtOf(seatID, cite); !stands {
			r.noteActErr(seatID + "'s refuting reading of " + cite + " left no reading debt of its own")
			return
		}
		// THE OWNER'S ENVELOPE IS REFUSED NAMING THE READING, and its sitting ends without one.
		if r.envelopeRefusedEarly(seatID, "reading of citation "+cite) {
			r.countReading(func(c *readingCounts) { c.envelopesRefusedOnDebt++ })
		}
		r.leaveDebt(seatID)
		r.reachedDebtChainStage(debtChainRead)
	case debtChainRead:
		if b, stands := r.debtOf(seatID, cite); !stands {
			r.reachedDebtChainStage("the debt on " + cite + " no longer stood at the owner's next sitting")
			return
		} else if r.readThisSitting(seatID, cite) {
			r.noteActErr(fmt.Sprintf("%s read %s again this sitting before the chain's mint (debt since %d)", seatID, cite, b.Since()))
			return
		}
		args := append(readingMintArgs(seatID, classContradicts), aboutArgs(cite)...)
		setArg(args, "--check-kind", "source")
		setArg(args, "--check", "a source states what the sentence states")
		out, err := r.exec(args...)
		if err != nil {
			r.reachedDebtChainStage("the mint about " + cite + " on the owner's reading was refused: " + firstLine(out))
			return
		}
		g := r.noteReadingMintLanded(seatID, out, classContradicts, "about")
		if g == "" {
			return
		}
		r.mu.Lock()
		r.debtChainGap, r.debtChainMinted = g, true
		r.mu.Unlock()
		if !r.debtDischargedBy(seatID, cite, "mint", "the owner's mint about the citation") {
			return
		}
		// THE GAP'S DEMAND REFUSES ITS CLOSE while nothing answers it.
		if !r.demandRefused(seatID, g) {
			r.reachedDebtChainStage("the close of source gap " + g + " with nothing answering it was not refused by its live-answer demand")
			return
		}
		r.reachedDebtChainStage(debtChainMinted)
	case debtChainRetired:
		// THE RECORD SAYS WHETHER THE GAP IS STILL THE OWNER'S TO CLOSE, asked at the act: a gap the
		// chain holds is never drawn to be superseded (heldByForcedChain), so one closed here was closed
		// by an act the drive does not expect, and the chain names it rather than issue a refused close.
		if !slices.Contains(r.openGaps(), gap) {
			r.reachedDebtChainStage("the owner's gap " + gap + " was closed before the owner's close: " + r.closedHow(gap))
			return
		}
		if _, ok := r.admitFresh(gap, "close", "--seat-id", seatID, "--id", gap, "--as", "repaired", "--reason", "fuzz: the citation left the sentence",
			"--verified-by", seatID, "--verified-with", "fuzz", "--verified-against", "rec"); !ok {
			return
		}
		r.mu.Lock()
		r.debtChainStage, r.debtChainDone = debtChainClosed, true
		r.mu.Unlock()
	}
}

// debtChainRetire is blue's half of the chain: once the owner's gap stands, blue retires the
// chain's citation answering it — its computations first — which meets the gap's `source` demand
// through the gap's `about` arm (reading_demand_met_by_retire). It draws nothing from the run's RNG.
func (r *runner) debtChainRetire(seatID string) {
	r.mu.Lock()
	due := r.forceReadingDebt && seatID == "blue-respond" && r.debtChainStage == debtChainMinted
	cite, gap := r.debtChainCite, r.debtChainGap
	r.mu.Unlock()
	if !due {
		return
	}
	if !r.retireCitingFirst(seatID, cite) {
		return
	}
	if _, ok := r.do("retire", seatID).set("--anchor", cite).set("--answers", gap).
		set("--reason", "fuzz: a lens's reading finds the source against the sentence, so the sentence stands without it").admit(); !ok {
		return
	}
	d, err := record.EvidenceDemandOf(r.run(), gap)
	switch {
	case err != nil:
		r.noteActErr("gap " + gap + "'s live-answer demand could not be read after blue's retire: " + err.Error())
		return
	case !d.Answered:
		r.noteActErr("blue's retire of " + cite + " answering " + gap + " did not meet the gap's demand through its about arm")
		return
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.readingDemandMetByRetire++ })
	r.reachedDebtChainStage(debtChainRetired)
}

// debtChainForcedErr is the run's error where readingDebtSeed's run did not drive each path it is
// forced through, naming the last step the chain reached; "" otherwise and on every other run. The
// caller holds r.mu: runOne reads it beside the other forced-run errors under the run's lock.
func (r *runner) debtChainForcedErr() string {
	if !r.forceReadingDebt || r.debtChainDone {
		return ""
	}
	stage := r.debtChainStage
	if stage == "" {
		stage = "nothing: the citation owner never sat"
	}
	return "forced reading-debt chain: the chain reached " + stage
}

// forcedErrUnderTheRunLock reads debtChainForcedErr the way runOne does, holding r.mu, and fails
// rather than hangs if the read takes the lock again.
func forcedErrUnderTheRunLock(t *testing.T, r *runner) string {
	t.Helper()
	got := make(chan string, 1)
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		got <- r.debtChainForcedErr()
	}()
	select {
	case msg := <-got:
		return msg
	case <-time.After(10 * time.Second):
		t.Fatal("debtChainForcedErr did not return under the run's lock: it takes r.mu, which runOne already holds")
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
// the drive's own sittings: the owner's reading leaving the debt, its envelope refused over it, the
// mint on it at the next sitting with the gap's demand refusing the close, blue's retire meeting the
// demand, and the owner's close. With the run not forced, none of it is driven and the run owes
// nothing.
func TestTheReadingDebtSeedDrivesEachPathItIsForcedThrough(t *testing.T) {
	r := stagedRunner(t)
	r.forceReadingDebt = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue = "blue-respond"
	if msg := forcedErrUnderTheRunLock(t, r); !strings.Contains(msg, "never sat") {
		t.Fatalf("a forced run nothing has driven reads %q, want the stage it never reached", msg)
	}
	r.sit("blue", blue)
	r.citeForDebtChain(blue)
	if r.debtChainCite == "" || r.mintsFresh(citationOwner) {
		t.Fatalf("blue's chain citation %q (stage %q); the owner mints fresh %v before its chain", r.debtChainCite, r.debtChainStage, r.mintsFresh(citationOwner))
	}
	if r.chainCitation() != r.debtChainCite || !r.reservedDebtSubject(r.debtChainCite) {
		t.Fatal("the chain's citation is not reserved to the chain")
	}
	r.sit("lens", citationOwner)
	r.forcedDebtChain(citationOwner)
	if r.actErr != "" || r.debtChainStage != debtChainRead || r.reading.envelopesRefusedOnDebt != 1 {
		t.Fatalf("the chain's first sitting reached %q (envelopes refused %d): %s", r.debtChainStage, r.reading.envelopesRefusedOnDebt, r.actErr)
	}
	r.answerReadingDebt(citationOwner)
	if !r.leftDebtOnly(citationOwner) {
		t.Fatal("the owner's envelope does not stand refused over the left debt alone")
	}
	r.sit("lens", citationOwner)
	r.forcedDebtChain(citationOwner)
	if r.actErr != "" || r.debtChainStage != debtChainMinted || r.reading.debtsDischarged["mint"] != 1 || r.sourceClosesRefused != 1 || !r.mintsFresh(citationOwner) {
		t.Fatalf("the chain's second sitting reached %q (discharged by mint %d, closes refused %d, mints fresh %v): %s",
			r.debtChainStage, r.reading.debtsDischarged["mint"], r.sourceClosesRefused, r.mintsFresh(citationOwner), r.actErr)
	}
	r.sit("blue", blue)
	r.debtChainRetire(blue)
	if r.actErr != "" || r.debtChainStage != debtChainRetired || r.lifecycle.readingDemandMetByRetire != 1 {
		t.Fatalf("blue's retire reached %q (demands met by retire %d): %s", r.debtChainStage, r.lifecycle.readingDemandMetByRetire, r.actErr)
	}
	r.sit("lens", citationOwner)
	r.forcedDebtChain(citationOwner)
	if r.actErr != "" || !r.debtChainDone || slices.Contains(r.openGaps(), r.debtChainGap) {
		t.Fatalf("the chain's close reached %q: %s", r.debtChainStage, r.actErr)
	}
	if msg := forcedErrUnderTheRunLock(t, r); msg != "" {
		t.Errorf("the forced run drove every path and still reads %q", msg)
	}
	t.Run("not forced", func(t *testing.T) {
		r := stagedRunner(t)
		if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
			t.Fatal(err)
		}
		r.sit("blue", "blue-respond")
		r.citeForDebtChain("blue-respond")
		r.sit("lens", citationOwner)
		r.forcedDebtChain(citationOwner)
		if r.debtChainCite != "" || r.debtChainStage != "" || forcedErrUnderTheRunLock(t, r) != "" {
			t.Errorf("a run not forced cited the chain's sentence (%s), drove the chain (%q) or owes it", r.debtChainCite, r.debtChainStage)
		}
	})
}

// THE PLAN AFTER THE OWNER'S READING READIES IT FOR THE DEBT AND REFUSES THE PASS: the chair's plan,
// relayed, names the owner ready to mint the gap or read again (reading_debt_readied), permits no
// PASS, and the chair's PASS is refused naming the reading (pass_refused_on_reading_debt). Blue is
// readied by no debt.
func TestTheDebtReadiesItsOwnerAndRefusesThePass(t *testing.T) {
	r := stagedRunner(t)
	r.forceReadingDebt = true
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue, chair = "blue-respond", "red-chair"
	r.sit("blue", blue)
	r.citeForDebtChain(blue)
	if r.debtChainCite == "" {
		t.Fatalf("blue's chain citation: %s (%s)", r.debtChainStage, r.actErr)
	}
	r.sit("lens", citationOwner)
	r.forcedDebtChain(citationOwner)
	if r.actErr != "" || r.debtChainStage != debtChainRead {
		t.Fatalf("the chain's reading reached %q: %s", r.debtChainStage, r.actErr)
	}
	r.sit("chair", chair)
	r.planThisSitting = r.dispatchNext(chair)
	plan := r.planThisSitting
	r.tallyReadingReadied(plan)
	r.passRefusedOnReadingDebt(chair, plan)
	blockers, _ := planBlockersAndParties(plan)
	if !r.planReadies(citationOwner) || r.planReadies(blue) || blockers[string(record.BlockerReadingDebt)] != 1 {
		t.Fatalf("the plan readies the owner %v and blue %v, with %d reading debt(s) — want the owner alone on its debt: %v", r.planReadies(citationOwner), r.planReadies(blue), blockers[string(record.BlockerReadingDebt)], plan["why"])
	}
	if permitted, _ := plan["pass_permitted"].(bool); permitted {
		t.Error("the plan permits a PASS over the owner's reading debt")
	}
	if r.actErr != "" || r.reading.debtReadied != 1 || r.reading.passRefusedOnDebt != 1 {
		t.Errorf("readied %d, PASS refused %d, want 1 and 1: %s", r.reading.debtReadied, r.reading.passRefusedOnDebt, r.actErr)
	}
}

// A FORCED CHAIN'S FIRST PLAN READIES ITS OWNER AND NOT BLUE: the first plan, after the synthesis,
// readies the chain's owner and not blue-respond. An inquiry the synthesis leaves open readies
// blue-respond for as long as it is open, and every later plan readies blue on the gaps the first
// epoch mints, so a chain that missed the first plan reached its epoch limit with every owner sitting
// in a plan that readies blue (seeds 6 and 15). The inquiry here is proposed and left open, as the
// synthesizer's drawn proposal leaves one.
func TestAForcedChainsFirstPlanReadiesItsOwnerAndNotBlue(t *testing.T) {
	for _, c := range []struct {
		name, owner string
		force       func(*runner)
	}{
		{"unsound computation", computationOwner, func(r *runner) { r.forceUnsoundComputation = true }},
		{"reading debt", citationOwner, func(r *runner) { r.forceReadingDebt = true }},
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

// closedHow is how the record says gap id closed: its closure reason, and the successor carrying it
// forward where one does.
func (r *runner) closedHow(id string) string {
	b, ok := r.board()
	if !ok {
		return "the record could not be read to say how"
	}
	g := record.GapsByID(b.Gaps)[id]
	if g == nil {
		return "the record holds no such gap"
	}
	how := "closed " + g.ClosureReason()
	if s := g.Successor(); s != "" {
		how += ", carried forward by " + s
	}
	return how
}
