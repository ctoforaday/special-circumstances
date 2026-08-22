package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

import "testing"

// board assembles a Board directly from events, so these tests exercise the estoppel logic
// rather than the shard reader (which record_test.go already covers).
func board(t *testing.T, gapFix map[string]string, evs []*Event) *Board {
	t.Helper()
	b := &Board{Gaps: map[string]*Gap{}}
	for id, fixNew := range gapFix {
		m := &recordpb.Mint{GapId: proto.String(id), FixBasis: proto.String("proposed")}
		if fixNew != "" {
			// `fix_old` HAS NO FIELD: the span is the gap's own `location`, and the second copy
			// was retired with the second matcher that read it.
			m.FixBasis = proto.String("verified")
			m.Location = proto.String("OLD:" + id)
			m.FixNew = proto.String(fixNew)
		}
		b.Gaps[id] = &Gap{ID: id, Open: true, Mint: m}
	}
	b.Events = evs
	return b
}

const prescribed = "Five verification approaches agree, all sharing one definition of primality."

func edit(t *testing.T, gapID string, verbatim bool) *Event {
	t.Helper()
	be := &recordpb.BlueEdit{
		Answers: proto.String(gapID),
		Old:     proto.String("x"),
		New:     proto.String("y"),
		Text:    proto.String("r"),
	}
	if verbatim {
		// PRESENT AND TRUE, not merely true: an edit that never claimed verbatim application and
		// one that claimed it and was false are different facts, and estoppel turns on the claim.
		be.AppliedVerbatim = proto.Bool(true)
	}
	return recordtest.Event(t, "blue-respond-r1", 1, be)
}

// The core rule: a finding located in text red prescribed and blue applied VERBATIM names
// the gap that prescribed it.
func TestEstoppelCatchesRelitigationOfRedsOwnPrescription(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []*Event{edit(t, "R1-1", true)})

	id, got := EstoppelConflict(b, "Five verification approaches agree, all sharing one definition of primality.")
	if id != "R1-1" {
		t.Fatalf("EstoppelConflict = %q, want R1-1 — red re-raising its own prescribed text went undetected", id)
	}
	if got != prescribed {
		t.Errorf("the conflict must return the prescribed text so the refusal can quote it, got %q", got)
	}
}

// A fragment of the prescribed sentence is the same act as quoting the whole of it.
func TestEstoppelMatchesAFragmentOfThePrescribedText(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []*Event{edit(t, "R1-1", true)})
	if id, _ := EstoppelConflict(b, "agree, all sharing one definition of primality"); id != "R1-1" {
		t.Errorf("a quoted FRAGMENT of red's own prescription escaped the guard (got %q)", id)
	}
}

// ESTOPPEL ATTACHES TO RED'S OWN WORDS AND NOTHING ELSE. If blue counter-edited, the text is
// blue's authorship and red audits it normally — that is the right to disagree staying real.
func TestNoEstoppelWhenBlueCounterEditedInstead(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []*Event{edit(t, "R1-1", false)})
	if id, _ := EstoppelConflict(b, prescribed); id != "" {
		t.Errorf("red was estopped from auditing text BLUE authored (gap %q) — a counter-edit is not red's prescription", id)
	}
}

// Text red never prescribed is auditable, obviously — the guard must not become a general
// shield over the report.
func TestNoEstoppelForUnrelatedText(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []*Event{edit(t, "R1-1", true)})
	if id, _ := EstoppelConflict(b, "An entirely different sentence about sieve performance and its costs."); id != "" {
		t.Errorf("an unrelated finding was estopped by gap %q — the guard is over-broad", id)
	}
}

// The overlap floor exists so a short prescription cannot shield half the report. Refusing a
// real finding is worse than missing an estoppel, so the guard declines to fire here.
func TestShortPrescriptionsDoNotEstop(t *testing.T) {
	b := board(t, map[string]string{"R1-1": "7 is prime."}, []*Event{edit(t, "R1-1", true)})
	if id, _ := EstoppelConflict(b, "7 is prime."); id != "" {
		t.Errorf("a %d-character prescription estopped a finding (gap %q); the floor is %d",
			len("7 is prime."), id, minEstoppelOverlap)
	}
}

// The decline rate is the measurement that falsifies the design if it comes back zero.
func TestDeclineStatsSeparatesAppliedFromDeclinedFromUnanswered(t *testing.T) {
	b := board(t, map[string]string{
		"R1-1": prescribed,          // applied verbatim
		"R1-2": prescribed + " Two", // blue counter-edited
		"R1-3": prescribed + " Three",
		"R1-4": "", // prose only: not an offer, must not be counted
	}, []*Event{
		edit(t, "R1-1", true),
		edit(t, "R1-2", false),
		// R1-3 offered and never answered.
		recordtest.Event(t, "blue-respond-r1", 1, &recordpb.BlueEdit{Answers: proto.String("R1-4")}),
	})

	offered, applied, declined := DeclineStats(b)
	if offered != 3 {
		t.Errorf("offered = %d, want 3 — only gaps carrying a concrete proposal are offers", offered)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}
	if declined != 1 {
		t.Errorf("declined = %d, want 1 — a counter-edit IS a decline", declined)
	}
	// The unanswered one is neither: scoring silence as agreement is how a decline rate lies.
	if unanswered := offered - applied - declined; unanswered != 1 {
		t.Errorf("unanswered = %d, want 1 — an unanswered offer must not be scored as agreement", unanswered)
	}
}

func TestEstoppelCountSurvivesRewordingTheRefusal(t *testing.T) {
	// THE TEXT IS THE SUBJECT — this test is about a counter that survives the refusal being
	// REWORDED — and the earlier conversion dropped it, leaving two identical empty frictions.
	fr := func(text string) *Event {
		return recordtest.Event(t, "red-merge-r2", 2, &recordpb.Friction{
			Text: proto.String(text),
			Kind: recordtest.P(recordpb.FrictionKind_FRICTION_KIND_ESTOPPEL),
		})
	}
	b := board(t, nil, []*Event{
		fr("merge mint: estoppel — this gap's location is text YOU prescribed"),
		fr("Refused: you are raising a fresh gap against your own prescription."), // reworded
	})
	if got := EstoppelRejections(b); got != 2 {
		t.Errorf("EstoppelRejections = %d, want 2 — rewording the message must not move a number read as evidence about red", got)
	}
}

// A seat's ORDINARY friction is not an estoppel rejection, even when it talks about one.
// Under the substring rule, a seat quoting the refusal in its own complaint inflated the
// count — the same confusion between a mention and the thing itself that the hook's
// position matcher exists to avoid.
func TestASeatsOwnComplaintIsNotARejection(t *testing.T) {
	b := board(t, nil, []*Event{
		recordtest.Event(t, "blue-respond-r2", 2, &recordpb.Friction{}),
	})
	if got := EstoppelRejections(b); got != 0 {
		t.Errorf("EstoppelRejections = %d, want 0 — a seat QUOTING the refusal did not cause one", got)
	}
}

// Zero is a real answer and must be reachable, so the printed "0" means "the guard did not
// fire" rather than "the detector is broken".
func TestNoRejectionsCountsZero(t *testing.T) {
	b := board(t, nil, []*Event{
		recordtest.Event(t, "red-merge-r1", 1, &recordpb.Friction{Text: proto.String("the fetch cache refused an unreachable url")}),
	})
	if got := EstoppelRejections(b); got != 0 {
		t.Errorf("EstoppelRejections = %d, want 0", got)
	}
}
