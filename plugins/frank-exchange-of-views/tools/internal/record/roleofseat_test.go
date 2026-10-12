package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// ONE ANSWER TO "WHAT ROLE IS THIS SEAT": every reader here gives the roster's, for a seat on it,
// a lane, and an id that only starts like one. A reader with its own table or its own prefix
// disagrees on the last kind first — `judge-x` is no bench seat and `red-lens-banana` no lens.
func TestOneAnswerToTheRoleOfASeat(t *testing.T) {
	ids := map[string]string{
		"red-chair": "chair", "blue-respond": "blue", "blue-synthesize": "blue", "frontier": "blue",
		"judge": "bench", "operator": "operator",
		"red-lens-adversary": "lens", "red-lens-architecture": "lens", "red-lens-computation": "lens",
		"red-lens-dark-side": "lens", "red-lens-evidence": "lens", "red-lens-logic": "lens", "red-lens-voice": "lens",
		"blue-lane-1": "blue", "blue-lane-2": "blue", "blue-lane-3": "blue",
		"blue-x": "", "judge-x": "", "judgement": "", "red-chair-x": "", "red-lens-banana": "",
		"frontierx": "", "operatorX": "", "harness": "", "": "", "blue-lane-0": "", "blue-lane-01": "",
	}
	for id := range seatclass.Seats {
		if _, ok := ids[id]; !ok {
			t.Errorf("roster seat %q is not in this table — every seat is held to the one answer", id)
		}
	}
	docket := recordpb.MotionSubject_MOTION_SUBJECT_DOCKET
	docketRuler, err := recordpb.SubjectRuler(docket)
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range ids {
		if got := seatclass.RoleOfSeat(id); got != want {
			t.Errorf("seatclass.RoleOfSeat(%q) = %q, want %q", id, got, want)
		}
		if got := PartyOf(&Event{SeatId: proto.String(id)}); got != want {
			t.Errorf("PartyOf(%q) = %q, want %q", id, got, want)
		}
		if seated := RequireDispatchedSeat(id) == nil; seated != (want != "") {
			t.Errorf("RequireDispatchedSeat(%q) admits = %v, and its role is %q", id, seated, want)
		}
		if owes := SeatOwesOccasion(id); owes != (want == benchRole) {
			t.Errorf("SeatOwesOccasion(%q) = %v, and its role is %q", id, owes, want)
		}
		// The seat that rules a subject files nothing on it: the write reads the filer's role.
		if refused := requireFilerHoldsNoGavel(id, docket, false) != nil; refused != (want == docketRuler) {
			t.Errorf("a docket filing by %q is refused = %v; its role is %q and the %s rules a docket", id, refused, want, docketRuler)
		}
		// Only a lens spends a mint budget, and an id off the roster is no lens.
		if lens := AreaOf(id) != ""; lens != (want == "lens") {
			t.Errorf("AreaOf(%q) says lens = %v, and its role is %q", id, lens, want)
		}
		if want == "" {
			continue
		}
		// A gavel rests with a seat of the role that holds it.
		if seat, err := gavelSeatOf(want); err == nil && seatclass.RoleOfSeat(seat) != want {
			t.Errorf("gavelSeatOf(%q) = %q, whose role is %q", want, seat, seatclass.RoleOfSeat(seat))
		}
	}
}

// THE BENCH'S ROLE WORD IS THE ROSTER'S. benchRole is what SeatOwesOccasion compares a seat's role
// with; a spelling no seat is seated under would stop the bench owing an occasion, silently.
func TestTheBenchRoleIsARosterRole(t *testing.T) {
	if len(seatclass.SeatsOfRole(benchRole)) == 0 {
		t.Errorf("benchRole is %q, and the roster seats nobody under it", benchRole)
	}
	if !SeatOwesOccasion(benchSeat) {
		t.Errorf("%s owes no occasion", benchSeat)
	}
}

// AN EVENT CARRIES NO ROLE. The party is read from the seat id at every read, so a stamped copy
// would be a second answer nothing holds to the first.
func TestAnEventCarriesNoRole(t *testing.T) {
	if f := (&recordpb.Event{}).ProtoReflect().Descriptor().Fields().ByName("role"); f != nil {
		t.Errorf("Event declares a role field (%d) — the party is PartyOf's answer, read from seat_id", f.Number())
	}
	if !(&recordpb.Event{}).ProtoReflect().Descriptor().ReservedNames().Has("role") {
		t.Error("Event does not reserve the name `role`, so a later field could reuse it")
	}
}
