package fuzz

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// sourceDemandSeed is the seed whose run is FORCED to close a `source` gap nothing answers, and a
// `source` gap carrying a refuting reading nothing answers, each refused by the live-answer demand
// (forceSourceDemand). It is none of the other forced seeds.
//
// Left to the draw, closeGap tries a source gap's close unanswered on a coin, but only after the gap's
// lens is readied on it, and the plan readies the lens only once blue has acted on the gap: blue's
// sitting answers every engaged gap's demand (dischargeOwed) before the lens sits again, so the drawn
// close finds the demand met. MEASURED on eb4166e16: zero such closes refused across 40 runs.
const sourceDemandSeed = 18

// sourceDemandOpen is the open question blue's half of the forced arm writes its claim after, and
// sourceDemandClaim the claim: a sentence no other drive quotes, so neither of the arm's gaps meets
// another gap of its class on its sentence.
const (
	sourceDemandOpen  = "Whether the trend outlasts the sample"
	sourceDemandClaim = "Whether the sourced claim holds"
)

// sourceDemandDriven is sourceDemandStage once forcedSourceDemand has driven every step.
const sourceDemandDriven = "driven"

// forcedSourceDemand is sourceDemandSeed's arm, at the citation owner's first sitting, before any act
// of its own: blue adds sourceDemandClaim; the lens mints a `source` gap on it and closes it with nothing
// answering, which the live-answer demand refuses; blue cites the claim as the gap's answer and the
// close is admitted; the lens reads that citation as refuted, minting a `source` gap carrying the
// reading, whose close with nothing answering is refused the same way. That gap is left to the drive.
// It draws from the run's RNG only where the drawn drives it calls do.
func (r *runner) forcedSourceDemand(seatID string) {
	r.mu.Lock()
	due := r.forceSourceDemand && seatID == citationOwner && r.sourceDemandStage == ""
	if due {
		r.sourceDemandStage = "begun"
	}
	r.mu.Unlock()
	if !due {
		return
	}
	stage := r.driveSourceDemand(seatID)
	r.mu.Lock()
	r.sourceDemandStage = stage
	r.mu.Unlock()
}

// driveSourceDemand drives forcedSourceDemand's steps and returns how far it reached:
// sourceDemandDriven, or the step that failed.
func (r *runner) driveSourceDemand(seatID string) string {
	const blue = "blue-respond"
	r.register("blue", blue)
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+sourceDemandClaim).
		set("--reason", "fuzz: a claim added for a source to back").admit(); !ok {
		return "blue's edit adding the claim was refused"
	}
	out, err := r.exec("--json", "mint", "--seat-id", seatID, "--quote", sourceDemandClaim, "--class", mintPoolClasses[0],
		"--problem", "fuzz: a claim no source backs", "--check-kind", "source", "--check", "a source states the claim",
		"--fix", dirCounter, "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--reason", "fuzz: the claim needs a source")
	var minted struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if err != nil || json.Unmarshal([]byte(strings.TrimSpace(out)), &minted) != nil || minted.Result.GapID == "" {
		return "the source gap's mint was refused: " + firstLine(out)
	}
	id := minted.Result.GapID
	r.noteMint(seatID, id)

	if !r.demandRefused(seatID, id, false) {
		return "the close of source gap " + id + " with nothing answering it was not refused by its live-answer demand"
	}
	cited, ok := r.do("cite", blue).set("--quote", sourceDemandClaim).set("--url", sourceURL("/blue-shared")).set("--title", sourceTitle).
		set("--source-span", sourceBody("/blue-shared")).set("--answers", id).set("--reason", "fuzz: the source the claim needs").bare("--json").admit()
	var label struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(cited)), &label) != nil || label.Result.Label == "" {
		return "blue's cite --answers " + id + " was refused or named no label: " + firstLine(cited)
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.sourceAnswers++ })
	if _, ok := r.admit("close", "--seat-id", seatID, "--id", id, "--as", "repaired", "--reason", "fuzz: the citation answers it",
		"--verified-by", seatID, "--verified-with", "fuzz", "--verified-against", "rec"); !ok {
		return "the close of source gap " + id + " after blue's cite --answers was refused"
	}

	anchorID := label.Result.Label
	r.firstThisSitting(seatID, anchorID)
	classes, _ := r.readingClasses(r.citationURL(anchorID), "refutes", true)
	if len(classes) == 0 {
		return "a refuting reading of citation " + anchorID + " admits no gap class"
	}
	gap, _ := r.readingGapArgs(classes)
	setArg(gap, "--check-kind", "source")
	c := r.do("verify", seatID).set("--anchor", anchorID).set("--as", "refutes").set("--confidence", "high").
		set("--reason", "fuzz: the source says otherwise than the claim").bare("--json")
	c.args = append(c.args, gap...)
	read, err := c.result()
	if err != nil {
		return "the refuting reading of citation " + anchorID + " was refused: " + firstLine(read)
	}
	v := r.checkReading(seatID, read, c.args)
	if v == nil || v.GetCarriedBy() == "" {
		return "the refuting reading of citation " + anchorID + " is carried by no gap"
	}
	if !r.demandRefused(r.minterOf(v.GetCarriedBy()), v.GetCarriedBy(), true) {
		return "the close of source gap " + v.GetCarriedBy() + ", carrying the reading, with nothing answering it was not refused by its live-answer demand"
	}
	return sourceDemandDriven
}

// demandRefused is whether closer's close of gap id, driven by demandCloseRefused, was counted as
// refused by the live-answer demand — and, where reading is set, as the close of a gap carrying a
// reading.
func (r *runner) demandRefused(closer, id string, reading bool) bool {
	r.mu.Lock()
	closes, readings := r.sourceClosesRefused, r.readingDemandRefused
	r.mu.Unlock()
	r.demandCloseRefused(closer, id)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sourceClosesRefused > closes && (r.readingDemandRefused > readings) == reading
}

// sourceDemandForcedErr is why sourceDemandSeed's run failed its arm, or "" where it drove it or is not
// forced. Called with r.mu held.
func (r *runner) sourceDemandForcedErr() string {
	switch {
	case !r.forceSourceDemand || r.sourceDemandStage == sourceDemandDriven:
		return ""
	case r.sourceDemandStage == "":
		return "forced live-answer demand: " + citationOwner + " never sat"
	}
	return "forced live-answer demand: " + r.sourceDemandStage
}

// THE FORCED LIVE-ANSWER DEMAND RUN DRIVES BOTH REFUSALS at the citation owner's first sitting, as the
// sweep's sitting drives it (envelopeFor): the close of a `source` gap nothing answers, then, once blue's
// citation answers it and the close is admitted, the close of the `source` gap a refuting reading of that
// citation mints. With the run not forced, none of it is driven and the run owes nothing.
func TestTheSourceDemandSeedDrivesBothRefusals(t *testing.T) {
	stage := func(t *testing.T, force bool) *runner {
		t.Helper()
		r := stagedRunner(t)
		r.forceSourceDemand = force
		if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
			t.Fatalf("record the lanes' parts: %v", err)
		}
		// THE SYNTHESIS FRAMES THE PARTS, writing the open question blue's half adds its claim after.
		r.synthesize("blue-synthesize")
		if r.actErr != "" {
			t.Fatalf("the synthesis: %s", r.actErr)
		}
		return r
	}
	t.Run("forced", func(t *testing.T) {
		r := stage(t, true)
		r.mu.Lock()
		if msg := r.sourceDemandForcedErr(); !strings.Contains(msg, "never sat") {
			t.Fatalf("a forced run nothing has driven reads %q, want the sitting it never reached", msg)
		}
		r.mu.Unlock()
		r.envelopeFor(citationOwner, "")
		r.mu.Lock()
		msg, closes, readings := r.sourceDemandForcedErr(), r.sourceClosesRefused, r.readingDemandRefused
		r.mu.Unlock()
		if msg != "" || closes < 2 || readings < 1 || r.actErr != "" {
			t.Fatalf("the forced arm reached %q: %d close(s) refused, %d of a gap carrying a reading — want both driven (%s)", msg, closes, readings, r.actErr)
		}
		// THE FIRST GAP CLOSED once blue's citation answered it.
		closed := r.closedGapIDs()
		answered := 0
		for _, g := range closed {
			if d, err := record.EvidenceDemandOf(r.run(), g); err == nil && d.Answered {
				answered++
			}
		}
		if answered == 0 {
			t.Errorf("no gap the arm minted closed with its demand answered (closed: %v)", closed)
		}
	})
	t.Run("not forced", func(t *testing.T) {
		r := stage(t, false)
		r.forcedSourceDemand(citationOwner)
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.sourceDemandStage != "" || r.sourceClosesRefused != 0 || r.sourceDemandForcedErr() != "" {
			t.Errorf("a run not forced drove the arm (stage %q, %d closes refused) or owes it", r.sourceDemandStage, r.sourceClosesRefused)
		}
	})
}
