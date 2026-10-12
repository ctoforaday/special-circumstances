package seatclass

import (
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
	// AgentType is the agent configuration the engine dispatches this seat under, spelled as the
	// harness delivers it on a hook payload. Empty for the operator, which nothing dispatches.
	//
	// THE SEAT NAMES ITS CONFIGURATION, AND NOTHING NAMES IT BACK. record's attestation table —
	// which roles a configuration may be seated as, and the one seat it always is — is computed
	// from this column, so a configuration and its seat are written once, here.
	AgentType string
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

// AgentTypePrefix is the plugin namespace the harness puts in front of every agent definition's
// name: `agents/red-chair.md` is dispatched, and attested on a hook payload, as this plus
// `red-chair`.
const AgentTypePrefix = "frank-exchange-of-views:"

// Seats is every seat id the engine can produce whose spelling is FIXED.
//
// The lens rows are GENERATED from flags.LensAreas rather than written out, so membership and
// shape stop being two questions with two answers: an id is a lens seat exactly when that list
// named its area. A pattern could bound the shape and never the membership, which is why
// `^red-lens-[a-z]+(?:-[a-z]+)*$` admitted `red-lens-banana` and a second list had to refuse it.
var Seats = roster()

func roster() map[string]Seat {
	m := map[string]Seat{
		"red-chair":       {Role: "chair", Base: "red-chair", AgentType: AgentTypePrefix + "red-chair"},
		"blue-respond":    {Role: "blue", Base: "blue-respond", AgentType: AgentTypePrefix + "blue-researcher"},
		"blue-synthesize": {Role: "blue", Base: "blue-synthesize", AgentType: AgentTypePrefix + "blue-synthesizer"},
		"frontier":        {Role: "blue", Base: "frontier", AgentType: AgentTypePrefix + "blue-researcher", CommonWord: true},
		"judge":           {Role: "bench", Base: "judge", AgentType: AgentTypePrefix + "lead-judge", CommonWord: true},
		OperatorRole:      {Role: OperatorRole, CommonWord: true},
	}
	for _, a := range flags.LensAreas {
		// One configuration per area, named for the seat: `agent_type` attests WHICH lens sat.
		m["red-lens-"+a] = Seat{Role: "lens", Base: "red-lens", Area: a, AgentType: AgentTypePrefix + "red-lens-" + a}
	}
	return m
}

// Lane is what every lane seat is, whatever its number. A lane's ID is generated from the run's
// lane count (record.LaneSeatIDs), so no row of Seats can carry one; its role, tier base and
// configuration do not vary with the number and are stated once, here.
var Lane = Seat{Role: "blue", Base: "blue-lane", AgentType: AgentTypePrefix + "blue-researcher"}

// LaneSeatPrefix is how a lane seat id is SPELLED, and the one place that says so.
//
// record.LaneSeatIDs composes ids from it; internal/reportvoice builds a prose pattern from it.
// Those are different acts and only one of them is the thing this change removed: SEARCHING TEXT
// for a name is not RECOVERING A FACT FROM AN ID. A report that says "blue-lane-2 drafted this"
// has named a seat to a reader of the subject, and a pattern is how prose gets searched; nothing
// downstream learns what that seat IS from the match.
const LaneSeatPrefix = "blue-lane-"

// THE LANE PATTERN IS GONE, AND SO IS THE LAST REGEX OVER A SEAT ID. Lane ids are GENERATED from
// the run's lane count (record.LaneSeatIDs), so "is this a lane" is a comparison against that set
// and "which lane" is a name the caller already holds. A pattern here could bound a shape and
// never membership, and the coverage audit read the lane's INDEX back out of its name.

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
