package seatclass

import (
	"regexp"
	"strconv"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// THE ROSTER LIVES HERE, AT THE BOTTOM, BECAUSE TWO SIDES OF THE TREE NEED IT.
//
// internal/record owns what a seat may DO; internal/reportvoice owns what a report may SAY, and
// one of the things it may not say is a seat id. record imports reportvoice, so the list of ids
// cannot live in record without reportvoice keeping a second copy — which is the defect, not a
// packaging inconvenience: the copy in reportvoice went stale the day `judge-terminal` was
// collapsed into `judge` (#1070), and for every run since, the voice gate has refused a name
// nothing can produce while `judge` and `frontier` walked past it.
//
// This package already holds the seat vocabulary's other half (the tier class), imports nothing
// but the areas the engine declares, and is imported by both sides. One list, one home.

// Seat is what a seat id used to carry in its SPELLING, held as fields instead.
//
// Base is the name SeatClass keys the tier class on. Area is the strategic area a lens audits and
// is empty for every other seat. Base is empty for the operator, which is not a debating seat and
// rides no tier.
type Seat struct {
	Role string
	Base string
	Area string
	// CommonWord marks an id that is ALSO an ordinary English word, so it cannot be guarded by
	// NAME anywhere. `judge`, `frontier` and `operator` are the three: a report about its subject
	// may say "a sitting judge", "the frontier of the field" or "the operator of the plant"
	// without naming anyone who acted in the run, and reportvoice's own clean-prose corpus
	// carries the first of those as prose it must not flag.
	//
	// IT IS A FIELD BECAUSE IT WAS AN ABSENCE. These ids were simply missing from the voice tell's
	// hand-written alternation, which reads identically to having been forgotten — and the same
	// list DID guard `judge-terminal`, a seat #1070 retired, so both a deliberate omission and a
	// stale entry sat in it with nothing to tell them apart. Stated, a reader can see which is
	// which.
	CommonWord bool
}

// OperatorRole is the identity a human or a script acts under, outside the debate.
const OperatorRole = "operator"

// Seats is every seat id the engine can produce whose spelling is FIXED.
//
// The lens rows are GENERATED from flags.LensAreas rather than written out, so membership and
// shape stop being two questions with two answers: an id is a lens seat exactly when that list
// named its area. A pattern could bound the shape and never the membership, which is why
// `^red-lens-[a-z]+(?:-[a-z]+)*$` admitted `red-lens-banana` and a second list had to refuse it.
var Seats = roster()

func roster() map[string]Seat {
	m := map[string]Seat{
		"red-chair":       {Role: "chair", Base: "red-chair"},
		"blue-respond":    {Role: "blue", Base: "blue-respond"},
		"blue-synthesize": {Role: "blue", Base: "blue-synthesize"},
		"frontier":        {Role: "blue", Base: "frontier", CommonWord: true},
		"judge":           {Role: "bench", Base: "judge", CommonWord: true},
		OperatorRole:      {Role: OperatorRole, CommonWord: true},
	}
	for _, a := range flags.LensAreas {
		m["red-lens-"+a] = Seat{Role: "lens", Base: "red-lens", Area: a}
	}
	return m
}

// LaneSeat is THE ONE PATTERN LEFT, and it is here for a reason that ends with #1153b.
//
// How many lanes a run has is a run PARAMETER (`--lanes`), so no table compiled into this binary
// can enumerate them. The run's cast does, and `dispatchableSeatID` could read it — but the tier
// join must answer `blue-lane` for a lane id, and the class is not a field on the cast yet.
// Deciding lane-ness by ELIMINATION instead (in the cast, absent from Seats) is a negative rule,
// and the one this tree already paid for was `benchClosesGap`: "everything except remanded"
// classified a later deferring disposition as closing and retired a gap the bench had
// deliberately kept alive.
var LaneSeat = regexp.MustCompile(`^blue-lane-(\d+)$`)

// LaneIndex is the lane a seat id names, and whether it names one.
//
// THE INDEX IS STILL LOAD-BEARING, WHICH IS WHY THIS EXISTS RATHER THAN THE CALLER PARSING. The
// lane-coverage audit reports WHICH lane never registered — "lane 3" sends a reader to that
// dispatch, where "2 of 3" sends them to the whole opening — so the fact is wanted, and until the
// cast carries it as a field (#1153b) it can only come off the id. What this removes is the
// SECOND copy of the pattern: internal/capture had its own, free to disagree with this one about
// what a lane id looks like.
func LaneIndex(seatID string) (int, bool) {
	m := LaneSeat.FindStringSubmatch(seatID)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

// SeatIDs is every seat id this roster names, for a reader that needs the NAMES rather than the
// facts behind them — reportvoice builds its tell from it, so a seat that leaves the roster
// leaves the gate with it and one that joins is guarded from its first run.
func SeatIDs() []string {
	ids := make([]string, 0, len(Seats))
	for id := range Seats {
		ids = append(ids, id)
	}
	return ids
}
