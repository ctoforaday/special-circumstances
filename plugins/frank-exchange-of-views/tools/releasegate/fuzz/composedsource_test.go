package fuzz

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE VOICE LENS AUDITS THE TEXT THE TOOL COMPOSES (ruled by gblock 2026-11-10; record/composedsource.go).
// An inquiry's fate prints its reason and method in a section the tool composes, and an open gap's
// problem and fix print in the risk matrix; the voice lens is readied for each one recorded since it
// last sat, and its work list hands it each. universe-m40 shipped "A later run should search…" in a
// deferred inquiry's method because nothing sent the voice lens there.
//
// The drive counts every composed source the voice lens's list hands it (composedHanded). In a run
// whose verdict nothing forces, it finds narration in one handed inquiry's account, a third of the
// time, and mints the report_voice gap about that inquiry (composedGaps) — the m40 shape — and blue's
// first sitting engaged on that gap records the inquiry's fate again with its account re-voiced
// (composedAnswers), the act that answers a gap about an inquiry. The sweep gates each at zero.

// composedAuditor is the lens the record hands the composed sources to.
const composedAuditor = "red-lens-voice"

// auditComposedSources is the voice lens's audit of the composed sources its work list hands it.
func (r *runner) auditComposedSources(seatID string) {
	if seatID != composedAuditor {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil {
		return
	}
	var inquiries []string
	handed := 0
	for _, it := range w.Sitting.Open {
		if it.Kind != record.ItemComposedSource || it.Subject == nil {
			continue
		}
		handed++
		if it.Subject.Kind == record.SubjectInquiry {
			inquiries = append(inquiries, it.Subject.ID)
		}
	}
	r.mu.Lock()
	r.composedHanded += handed
	due := r.composedGap == "" && !r.forcedVerdict() && !r.forceEstoppel && !r.forceFateTurn
	r.mu.Unlock()
	if !due || len(inquiries) == 0 || !r.coin(35) {
		return
	}
	inquiry := inquiries[r.rng.Intn(len(inquiries))]
	// A REFUSAL IS THE RECORD'S AND THE NEXT SITTING TRIES AGAIN: an open report_voice gap about the
	// inquiry already (one gap per class on one inquiry).
	out, err := r.exec("--json", "mint", "--seat-id", seatID, "--class", "report_voice",
		"--problem", "The method recounts the search instead of stating what would settle the question.",
		"--fix", dirCounter, "--check-kind", "document", "--check", "the inquiry's method reads to a reader of the subject",
		"--severity", "medium", "--likelihood", "medium", "--impact", "low",
		"--about-kind", "inquiry", "--about", inquiry, "--reason", "fuzz: narration in a composed section's source")
	if err != nil {
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
	r.noteMint(seatID, env.Result.GapID)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.composedGap, r.composedInquiry = env.Result.GapID, inquiry
	r.composedGaps++
}

// answerComposedGap is blue's answer to the voice lens's gap about an inquiry's account: the fate
// recorded again, its status kept and its account re-voiced, at blue's first sitting engaged on it.
func (r *runner) answerComposedGap(seatID string) {
	r.mu.Lock()
	gap, id, done := r.composedGap, r.composedInquiry, r.composedAnswered
	r.mu.Unlock()
	if gap == "" || done || !slices.Contains(r.engaged(seatID), gap) {
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
			set("--reason", "fuzz: the same fate, its account stated about the subject").
			set("--method", "fuzz: what would settle the question").admit(); !ok {
			return
		}
		r.mu.Lock()
		r.composedAnswered = true
		r.composedAnswers++
		r.mu.Unlock()
		return
	}
}

// THE VOICE DRIVE REACHES ITS WHOLE CHAIN, run outside the release gate on the drive's own sittings:
// blue settles an inquiry with a reason, the voice lens's list hands it that inquiry's account, the
// lens mints the report_voice gap about it, and blue's sitting engaged on the gap records the fate
// again — each counted once.
func TestTheVoiceDriveMintsAboutAComposedInquiryAndBlueAnswers(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	const blue, chair = "blue-respond", "red-chair"
	r.sit("blue", blue)
	out, ok := r.do("inquiry propose", blue).set("--question", "fuzz: does it settle").set("--hypothesis", "fuzz: it settles").admit()
	id := inquiryIDOf(out)
	if !ok || id == "" {
		t.Fatalf("propose: %s (%s)", out, r.actErr)
	}
	if _, ok := r.do("inquiry record", blue).set("--id", id).set("--as", "deferred").set("--reason", "fuzz: why it waits").
		set("--method", "fuzz: a later run should search more").admit(); !ok {
		t.Fatalf("record: %s", r.actErr)
	}
	plan := func() {
		t.Helper()
		r.sit("chair", chair)
		r.planThisSitting = r.dispatchNext(chair)
	}
	plan()
	r.sit("lens", composedAuditor)
	r.auditRead(composedAuditor)
	for range 40 {
		if r.composedGap != "" {
			break
		}
		r.auditComposedSources(composedAuditor)
	}
	if r.composedHanded == 0 || r.composedGaps != 1 || r.composedInquiry != id {
		t.Fatalf("handed %d, minted %d about %q, want sources handed and one gap about %s (%s)", r.composedHanded, r.composedGaps, r.composedInquiry, id, r.actErr)
	}
	plan()
	if !slices.Contains(r.engaged(blue), r.composedGap) {
		t.Fatalf("the plan does not engage blue on the voice lens's gap %s: %v", r.composedGap, r.planThisSitting["parties"])
	}
	r.sit("blue", blue)
	r.answerComposedGap(blue)
	r.answerComposedGap(blue)
	if r.composedAnswers != 1 {
		t.Errorf("blue answered the gap about the inquiry's account %d time(s), want 1 (%s)", r.composedAnswers, r.actErr)
	}
	if r.actErr != "" {
		t.Errorf("the drive: %s", r.actErr)
	}
}
