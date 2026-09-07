package seat

import (
	"strings"
	"testing"
)

// THE REGISTER'S WARNINGS, and what stopped being one.
//
// THE DISCARD WARNING IS GONE WITH THE THING IT WARNED ABOUT. Three tests here covered a second
// sitting being told what its first one lost — real under shards, where replay kept one shard and
// the other sitting's events survived nowhere. Under the store both sittings are rows and nothing
// selects a winner, so there is no discard to disclose; record/agentbinding.go carries the full
// reasoning and #533 carries the residual risk (a re-dispatched seat rewriting report.md, which
// loses the FILE and not the events). Deleting the tests with the feature is deliberate: a test
// retargeted at a mechanism that no longer exists is how a dead guard reads as a live one.

// A DISPATCHED SEAT WHOSE IDENTITY NEVER ARRIVED IS TOLD, WHILE ITS SITTING IS STILL RUNNING.
//
// The record already handles this correctly at the write (absent recorded as absent) and capture
// already reports it hours later. Nothing told the SEAT. Measured: in
// research/2026-08-22_is-7-prime all FOURTEEN registers carry no agent_id — the hook never fired
// for that run — so every later call was refused "this agent has not registered" whatever shape
// it took. That message was returned 92 times across the session and was false every time.
func TestASeatWhoseIdentityNeverArrivedIsToldWhatToDo(t *testing.T) {
	h := registerResult{SeatID: "red-lens-r1-evidence", IdentityAbsent: true}.Human()
	for _, want := range []string{
		"registered red-lens-r1-evidence",
		"YOUR IDENTITY DID NOT REACH THE TOOL",
		"PASS --seat-id ON EVERY CALL",
		"REGISTERING AGAIN WILL NOT FIX IT",
		// NOT "rotates your shard nonce". That consequence was true of shards and is false here:
		// a second register costs a dispatch count and no work, so the warning says the record
		// keeps both sittings instead of threatening the seat with a mechanism that is gone.
		"the record is append-only and keeps both sittings",
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
	h := registerResult{SeatID: "blue-lane-1", IdentityAbsent: true}.Human()
	if !strings.HasPrefix(h, "registered blue-lane-1") {
		t.Errorf("the registration line is no longer first:\n%s", h)
	}
}

// THE ORDINARY CASE STAYS SILENT HERE TOO. Every seat in a healthy run registers with an agent
// id, so a warning on that path would be noise on every register in every run.
func TestAnArrivedIdentitySaysNothing(t *testing.T) {
	h := registerResult{SeatID: "blue-lane-2"}.Human()
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
	h := registerResult{SeatID: "red-lens-r1-evidence", IdentityAbsent: true, RunVia: "inferred", HookAbsent: true}.Human()
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
	h := registerResult{SeatID: "blue-lane-1", IdentityAbsent: true, RunVia: "wrapper", HookAbsent: true}.Human()
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
	h := registerResult{SeatID: "red-lens-r1-evidence", IdentityAbsent: true, RunVia: "flag"}.Human()
	if strings.Contains(h, "NOT ONLY YOUR IDENTITY") {
		t.Errorf("a seat that typed --run is told the hook is absent:\n%s", h)
	}
}
