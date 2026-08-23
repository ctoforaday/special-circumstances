package seat

import (
	"strings"
	"testing"
)

// A SECOND SITTING IS TOLD WHAT THE FIRST ONE LOST.
//
// The shape is the one measured in research/2026-08-22_record-store-authority: a workflow killed
// between a seat writing its artifact and its call returning, then resumed — so the seat is
// dispatched twice, replay keeps one shard, and the other sitting's events survive nowhere. The
// second sitting there rewrote the report from scratch and never learned that four citations its
// predecessor had fetched were gone, one of them the only source in the run for its claim.
func TestASecondSittingIsToldWhatTheFirstLost(t *testing.T) {
	r := registerResult{
		SeatID: "blue-synthesize",
		Nonce:  "5060c5cd",
		PriorSitting: &priorSitting{
			Dispatches: 2,
			DiscardedKeys: []string{
				"blue-synthesize:cite:c-1e6a02f9",
				"blue-synthesize:cite:c-ffcf9002",
				"blue-synthesize:line-of-inquiry:#13",
				"blue-synthesize:blue_edit:#4",
				"blue-synthesize:blue_edit:#5",
			},
		},
	}
	h := r.Human()
	for _, want := range []string{
		"registered blue-synthesize (shard nonce 5060c5cd)",
		"YOU ARE SITTING 2 OF THIS SEAT",
		"5 event(s)",
		"2 cite, 1 line-of-inquiry, 2 blue_edit", // counted BY TYPE, in first-seen order
		"blue-synthesize:cite:c-1e6a02f9",        // and the keys themselves: the anchor is the actionable half
		"cite — a source was fetched",            // guidance for the kinds actually lost...
		"blue_edit — a change to the report",
	} {
		if !strings.Contains(h, want) {
			t.Errorf("the warning does not carry %q:\n%s", want, h)
		}
	}
	// ...AND NOT FOR KINDS THAT WERE NOT. A synthesizer that lost citations and edits is not
	// told what a dropped `mint` or `close` would have meant; the golden corpus caught the
	// unconditional version, where a merge seat losing one mint was lectured about citations.
	for _, unwanted := range []string{"mint —", "close —", "finding —"} {
		if strings.Contains(h, unwanted) {
			t.Errorf("the warning advises about %q, a kind this sitting did not lose:\n%s", unwanted, h)
		}
	}
}

// THE ORDINARY CASE STAYS SILENT. Every first sitting of every seat hits this path, so a warning
// here would be noise on every register in every run — and noise is how a real one stops landing.
func TestAFirstSittingIsToldNothing(t *testing.T) {
	h := registerResult{SeatID: "red-merge-r1", Nonce: "af1bfabf"}.Human()
	if h != "registered red-merge-r1 (shard nonce af1bfabf)" {
		t.Errorf("a clean register grew a warning:\n%s", h)
	}
}

// NOT MEASURED IS NOT A CLEAN BOARD, and this is the whole reason the field exists separately.
// A read failure that rendered as silence would be indistinguishable from "your predecessor lost
// nothing" — the defect class this warning was built to break, reproduced inside the warning.
func TestAFailedCheckSaysSoRatherThanReadingAsClean(t *testing.T) {
	h := registerResult{
		SeatID:              "blue-lane-1",
		Nonce:               "514729bc",
		PriorSittingUnknown: "record: shard unreadable",
	}.Human()
	if !strings.Contains(h, "COULD NOT BE CHECKED") || !strings.Contains(h, "NOT MEASURED") {
		t.Errorf("a failed check did not say so:\n%s", h)
	}
	if strings.Contains(h, "YOU ARE SITTING") {
		t.Errorf("a failed check rendered as a discard warning:\n%s", h)
	}
}

// A DISPATCHED SEAT WHOSE IDENTITY NEVER ARRIVED IS TOLD, WHILE ITS SITTING IS STILL RUNNING.
//
// The record already handles this correctly at the write (absent recorded as absent) and capture
// already reports it hours later. Nothing told the SEAT. Measured: in
// research/2026-08-22_is-7-prime all FOURTEEN registers carry no agent_id — the hook never fired
// for that run — so every later call was refused "this agent has not registered" whatever shape
// it took. That message was returned 92 times across the session and was false every time.
func TestASeatWhoseIdentityNeverArrivedIsToldWhatToDo(t *testing.T) {
	h := registerResult{SeatID: "red-lens-r1-L1", Nonce: "7a6b912e", IdentityAbsent: true}.Human()
	for _, want := range []string{
		"registered red-lens-r1-L1",
		"YOUR IDENTITY DID NOT REACH THE TOOL",
		"PASS --seat-id ON EVERY CALL",
		"DO NOT REGISTER AGAIN",
		"rotates your shard nonce",
	} {
		if !strings.Contains(h, want) {
			t.Errorf("the advisory does not carry %q:\n%s", want, h)
		}
	}
}

// AND IT IS AN ADVISORY, NOT A REFUSAL. Refusing would wedge a run whose only fault is a hook
// that did not fire — the seat can work perfectly well by passing --seat-id, and a run that
// cannot start is strictly worse than one that starts knowing.
func TestTheIdentityAdvisoryDoesNotDisplaceTheRegistration(t *testing.T) {
	h := registerResult{SeatID: "blue-lane-1", Nonce: "b647e93a", IdentityAbsent: true}.Human()
	if !strings.HasPrefix(h, "registered blue-lane-1 (shard nonce b647e93a)") {
		t.Errorf("the registration line is no longer first:\n%s", h)
	}
}

// THE ORDINARY CASE STAYS SILENT HERE TOO. Every seat in a healthy run registers with an agent
// id, so a warning on that path would be noise on every register in every run.
func TestAnArrivedIdentitySaysNothing(t *testing.T) {
	h := registerResult{SeatID: "blue-lane-2", Nonce: "a2d8351d"}.Human()
	if strings.Contains(h, "YOUR IDENTITY") {
		t.Errorf("a healthy register grew an advisory:\n%s", h)
	}
}

// THE HOOK IS NOT REACHING THIS RUN AT ALL — a bigger fact than a missing identity, and only
// visible from the identity's absence and an INFERRED run directory together.
//
// #512 is that this fact had no carrier. In research/2026-08-22_is-7-prime fourteen seats
// registered with no agent id, no stderr was written, and a day later nothing on the record could
// say whether the hook declined, was never invoked, or found no marker — three states and one
// silence. The seat is the first party that can see the pair, and the only one still running.
func TestASeatSeesThatTheHookIsMissingFromTheWholeRun(t *testing.T) {
	h := registerResult{SeatID: "red-lens-r1-L1", Nonce: "7a6b912e", IdentityAbsent: true, RunVia: "inferred", HookAbsent: true}.Human()
	for _, want := range []string{
		"YOUR IDENTITY DID NOT REACH THE TOOL", // the seat's own consequence still comes first
		"THE ENGINE'S HOOK IS NOT REACHING THIS RUN",
		"EVERY seat in it is affected",
		"the tool's own search for the run marker", // WHICH carrier stood in, so the operator can tell the two shapes apart
		"friction verb", // the one channel that carries it out of the run
	} {
		if !strings.Contains(h, want) {
			t.Errorf("the advisory does not carry %q:\n%s", want, h)
		}
	}
}

// UNDER A WRAPPER THE SAME FACT IS READABLE ON ITS OWN, and it must not read as alarming.
//
// setup bakes the run into <runDir>/.bin/feov-record, and the hook — when it fires — sets
// FEOV_RUN, which outranks it. So resolving by WRAPPER says the hook did not inject, with no
// pairing needed. The seat's work is fine in that state: the run directory is correct, which is
// the whole point of the wrapper, and telling a seat otherwise would spend a round on a
// non-problem.
func TestUnderAWrapperTheHookAbsenceIsStatedWithoutAlarm(t *testing.T) {
	h := registerResult{SeatID: "blue-lane-1", Nonce: "b647e93a", IdentityAbsent: true, RunVia: "wrapper", HookAbsent: true}.Human()
	for _, want := range []string{
		"THE ENGINE'S HOOK IS NOT REACHING THIS RUN",
		"this run's own wrapper",
		"YOUR WORK IS NOT AT RISK",
		"friction verb",
	} {
		if !strings.Contains(h, want) {
			t.Errorf("the advisory does not carry %q:\n%s", want, h)
		}
	}
	// It must not tell a wrapped seat the tool went looking for the marker — it did not.
	if strings.Contains(h, "own search for the run marker") {
		t.Errorf("a wrapped seat is told the run was inferred:\n%s", h)
	}
}

// A MISSING IDENTITY ALONE IS NOT A MISSING HOOK. A seat that typed --run on a healthy run looks
// identical on the identity alone, and telling it the whole run is broken would be a false
// diagnosis handed to the one party positioned to report it.
func TestAMissingIdentityAloneDoesNotAccuseTheHook(t *testing.T) {
	h := registerResult{SeatID: "red-lens-r1-L1", Nonce: "7a6b912e", IdentityAbsent: true, RunVia: "flag"}.Human()
	if strings.Contains(h, "NOT ONLY YOUR IDENTITY") {
		t.Errorf("a seat that typed --run is told the hook is absent:\n%s", h)
	}
}
