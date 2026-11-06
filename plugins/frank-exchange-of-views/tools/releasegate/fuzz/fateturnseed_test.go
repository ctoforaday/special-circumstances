package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
)

// fateTurnSeed is the seed whose run is FORCED through a lens's turn that a fate blue changed brought
// on a gap about a section the tool composes from fates (ruled by gblock 2026-11-05), the turn the
// sweep's fate_turns gate counts.
//
// Left to the draw it is a chain of coins: a lens mints about a composed section (half the section
// draws, once inquiries exist), the gap is material, blue's sitting engaged on it records a fate that
// section prints with a changed status, reason or method, and a later plan engages the lens on the gap.
// MEASURED on #1744's sweep (run 41988247415), with the tally reading only a lens engaged beside blue:
// 0 of 40 runs reached it once blue was held out of the last epoch that dispatches a lens, where main's
// sweep reached it on the draw.
//
// The run's first lens sitting mints a material gap about the section that prints one inquiry's fate
// (mintAboutFateSection), before any other mint of its sitting spends its budget. Blue's first sitting
// engaged on that gap records the inquiry's fate again with its status kept and its reason re-voiced
// (changeFateForTurn) — the m38 shape: the only way blue changes a composed section is a fate. The next
// plan engages the lens, whose turn that fate brought, alone or beside blue, and the lens-turn tally
// counts it.
const fateTurnSeed = 11

// mintAboutFateSection is the forced run's lens mint about the composed section that prints an
// inquiry's fate, once per run, by the first lens sitting with an inquiry on the record.
func (r *runner) mintAboutFateSection(seatID string) {
	r.mu.Lock()
	due := r.forceFateTurn && r.fateTurnGap == ""
	r.mu.Unlock()
	if !due {
		return
	}
	js, ok := r.inquiryJudgements()
	if !ok || len(js) == 0 {
		r.reachedFateTurnStage("a lens sat before any inquiry was on the record")
		return
	}
	q := js[0]
	section := reportheading.InquirySectionOf(q.Status)
	out, err := r.exec("--json", "mint", "--seat-id", seatID,
		"--problem", "fuzz: the section prints run machinery beside the inquiry's fate",
		"--check-kind", "document", "--check", "the section reads to a reader of the subject", "--fix", dirCounter,
		"--severity", "high", "--likelihood", "high", "--impact", "high", "--class", pick(r.rng, mintPoolClasses),
		"--about-kind", "section", "--about", section, "--reason", "fuzz: the composed section blue changes by a fate")
	if err != nil {
		r.reachedFateTurnStage(seatID + "'s mint about " + section + " was refused: " + firstLine(out))
		return
	}
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
		r.noteActErr("an admitted mint's --json result names no gap_id: " + firstLine(out))
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.minter == nil {
		r.minter = map[string]string{}
		r.lensMints = map[string]int{}
	}
	r.minter[env.Result.GapID] = seatID
	r.fateTurnGap, r.fateTurnInquiry = env.Result.GapID, q.ID
	r.fateTurnStage = fmt.Sprintf("%s minted %s about %s, which prints %s's fate; blue never sat engaged on it", seatID, env.Result.GapID, section, q.ID)
}

// changeFateForTurn is blue's answer to the forced gap: the inquiry's fate recorded again, its status
// kept and its reason re-voiced, at blue's first sitting engaged on the gap.
func (r *runner) changeFateForTurn(seatID string) {
	r.mu.Lock()
	gap, id := r.fateTurnGap, r.fateTurnInquiry
	due := r.forceFateTurn && gap != "" && !r.fateTurnChanged
	r.mu.Unlock()
	if !due || !slices.Contains(r.engaged(seatID), gap) {
		return
	}
	js, ok := r.inquiryJudgements()
	if !ok {
		return
	}
	for _, j := range js {
		if j.ID != id {
			continue
		}
		if _, ok := r.do("inquiry record", seatID).set("--id", id).set("--as", j.Status).
			set("--reason", "fuzz: the same fate, told to a reader of the subject, answering "+gap).admit(); !ok {
			r.reachedFateTurnStage("blue's fate on " + id + ", answering " + gap + ", was refused")
			return
		}
		r.mu.Lock()
		r.fateTurnChanged = true
		r.fateTurnStage = "blue changed " + id + "'s fate answering " + gap + "; no later plan engaged the lens on it"
		r.mu.Unlock()
		return
	}
}

// fateTurnNeedsInquiry reports whether the forced run's chain waits on an inquiry: forced, its gap not
// yet minted, and no inquiry on the record. A read that fails is no inquiry, and the chain's own stage
// then names the lens that sat without one.
func (r *runner) fateTurnNeedsInquiry() bool {
	r.mu.Lock()
	due := r.forceFateTurn && r.fateTurnGap == ""
	r.mu.Unlock()
	if !due {
		return false
	}
	js, ok := r.inquiryJudgements()
	return !ok || len(js) == 0
}

func (r *runner) reachedFateTurnStage(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fateTurnStage = s
}

// fateTurnForcedErr is the run's error where fateTurnSeed's run counted no fate turn, naming the last
// step its chain reached; "" otherwise and on every other run. The caller holds r.mu.
func (r *runner) fateTurnForcedErr() string {
	if !r.forceFateTurn || r.fateTurns > 0 {
		return ""
	}
	stage := r.fateTurnStage
	if stage == "" {
		stage = "nothing: no lens sat"
	}
	return "forced fate turn: no lens turn on a gap about a section came with a fate blue changed; the chain reached " + stage
}

// THE FORCED FATE-TURN RUN COUNTS A FATE TURN, run outside the release gate on the drive's own
// sittings: the lens mints about the section that prints an inquiry's fate, blue's sitting engaged on
// the gap changes that fate, and the next plan engages the lens, whose turn the fate brought, and blue
// — which the lens-turn tally counts. With the run not forced, none of it is driven.
func TestTheFateTurnSeedReachesAFateTurn(t *testing.T) {
	r, plan := drivenToAChangedFate(t)
	r.tallyLensTurns(plan())
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.fateTurns != 1 {
		t.Errorf("fate turns %d, want 1: %s", r.fateTurns, r.fateTurnStage)
	}
	if msg := r.fateTurnForcedErr(); msg != "" {
		t.Errorf("the forced run drove its chain and still reads %q", msg)
	}
}

// A LENS ALONE ON THE GAP AFTER BLUE'S CHANGED FATE IS A FATE TURN TOO (ruled by gblock 2026-11-05):
// the epoch blue is held out of (#1744) re-audits the fate with the lens alone, and the tally counts it
// as it counts the lens engaged beside blue.
func TestALensAloneAfterAChangedFateIsAFateTurn(t *testing.T) {
	r, _ := drivenToAChangedFate(t)
	r.tallyLensTurns(map[string]any{"parties": []any{
		map[string]any{"seat_id": citationOwner, "gap_ids": []any{r.fateTurnGap}},
	}})
	if r.actErr != "" {
		t.Fatalf("the drive: %s", r.actErr)
	}
	if r.fateTurns != 1 || r.lensAlone != 1 {
		t.Errorf("fate turns %d and lens-alone engagements %d, want 1 and 1: %s", r.fateTurns, r.lensAlone, r.fateTurnStage)
	}
}

// drivenToAChangedFate is a forced run driven, on its own sittings, to blue's changed fate answering
// the lens's gap about the composed section that prints it, and plan, which sits the chair and returns
// the plan it relays.
func drivenToAChangedFate(t *testing.T) (*runner, func() map[string]any) {
	t.Helper()
	r := stagedRunner(t)
	r.forceFateTurn = true
	if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	const blue, chair, lens = "blue-respond", "red-chair", citationOwner
	if msg := r.fateTurnForcedErr(); !strings.Contains(msg, "no lens sat") {
		t.Fatalf("a forced run nothing has driven reads %q", msg)
	}
	r.sit("blue", blue)
	if !r.fateTurnNeedsInquiry() {
		t.Fatal("a forced run with no inquiry on the record does not have blue propose one")
	}
	out, ok := r.do("inquiry propose", blue).set("--question", "fuzz: does it settle").set("--hypothesis", "fuzz: it settles").admit()
	id := inquiryIDOf(out)
	if !ok || id == "" {
		t.Fatalf("propose: %s (%s)", out, r.actErr)
	}
	if _, ok := r.do("inquiry record", blue).set("--id", id).set("--as", "confirmed").set("--reason", "fuzz: what the approach found").admit(); !ok {
		t.Fatalf("record: %s", r.actErr)
	}
	if r.fateTurnNeedsInquiry() {
		t.Fatal("the forced run still waits on an inquiry with one on the record")
	}
	plan := func() map[string]any {
		t.Helper()
		r.sit("chair", chair)
		r.planThisSitting = r.dispatchNext(chair)
		return r.planThisSitting
	}
	plan()
	r.sit("lens", lens)
	r.auditRead(lens)
	r.mintAboutFateSection(lens)
	if r.fateTurnGap == "" {
		t.Fatalf("the lens minted no gap about the composed section: %s (%s)", r.fateTurnStage, r.actErr)
	}
	plan()
	r.sit("blue", blue)
	r.changeFateForTurn(blue)
	if !r.fateTurnChanged {
		t.Fatalf("blue changed no fate answering %s: %s (%s)", r.fateTurnGap, r.fateTurnStage, r.actErr)
	}
	return r, plan
}
