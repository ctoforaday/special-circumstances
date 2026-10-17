package record

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// Party is one seat the chair engages and the gaps it is engaged on. A lens engaged with no gaps
// has never sat or last sat behind the report head (retirement.go), or owes a finding on a
// contradiction it read, or — the evidence lens — the check of a citation or proof reopened since
// its last one, and sits to audit the current report.
type Party struct {
	SeatID string   `json:"seat_id"`
	GapIDs []string `json:"gap_ids" desc:"the gaps this party is engaged on; empty for a lens readied to audit the report — it has never sat, or the report head has moved since its last sitting — or by a contradiction it read, or by a reopened citation or proof" bench:"the gaps docketed for the bench; empty when it is convened for a petition alone"`
	// Occasions is what the bench is convened for this epoch, in the Occasion enum's words: `docket`
	// (the gap_ids, after blue sits) and `petition` (the unruled petitions, before any party sits).
	// The bench's one id cannot say which question it is asked, so the plan says it as a field and
	// the workflow routes on it — never on an empty gap list. Every other seat has no occasion, and
	// the field is absent.
	Occasions []string `json:"occasions,omitempty" desc:"absent, as the verb prints it: only the bench is convened for an occasion" bench:"what the bench is convened for — petition (before any other party sits), docket (its gap_ids, after blue)"`
}

// PlanBlocker is one entry of the gate's list (PassBlockers) as the plan relays it: what holds a
// PASS, what it is, and the seat whose act clears it ("" where no seat on the record can). The
// workflow reads the record only through the plan, so this is how it learns that a motion whose
// gavel is the bench's still stands at the exit.
type PlanBlocker struct {
	Kind    BlockerKind `json:"kind"`
	Subject string      `json:"subject" desc:"the reference that names it: a gap id, a motion id, a lens seat, an anchor, a contradiction's id; empty for the run-wide kinds"`
	Owner   string      `json:"owner"`
}

// RemandOwed is one gap whose remand's exchange the plan readies, and the direction that exchange
// owes — the gap view's `docket_direction`, which the record refuses blank on a remand.
type RemandOwed struct {
	GapID     string `json:"gap_id"`
	Direction string `json:"direction" pattern:"\\S" desc:"never blank: the record refuses a remand without one"`
}

// ExchangeCount is one gap's exchanges as the plan relays them: Count is the total had on the gap
// (GapExchanges.Exchanges — monotone, never reset by a remand), and Unresolved is the engaged
// sittings the record cannot close (GapExchanges.Unresolved), which are not counted and not zero. A
// gap with no counted exchange and an unresolved sitting is NOT MEASURED, and the two fields say so
// apart.
type ExchangeCount struct {
	GapID      string `json:"gap_id"`
	Count      int    `json:"count" min:"0"`
	Unresolved int    `json:"unresolved" min:"0"`
}

// exchangeCountOf is a gap's entry off the fold; x is nil for a gap no exchange fold names.
func exchangeCountOf(gapID string, x *GapExchanges) ExchangeCount {
	if x == nil {
		return ExchangeCount{GapID: gapID}
	}
	return ExchangeCount{GapID: gapID, Count: x.Exchanges, Unresolved: x.Unresolved}
}

// Plan is what `feov-record dispatch next` computes FROM THE BOARD (plans/roundless.md §III.B.1).
// The chair relays it; the workflow dispatches what it says; nothing in it is a seat's assertion.
// The record stores it as printed (recordpb.Plan), and capture compares every relay with that.
type Plan struct {
	// Key is the plan's identity: two plans carry one key when they decide the same — every field
	// but those tagged `key:"-"`, which say when the plan was asked (keyed). The engine's no-progress
	// valve compares keys across chair sittings, so a plan the record computes the same sitting after
	// sitting is one key repeating.
	Key string `json:"key" pattern:"^[0-9a-f]{16}$" desc:"the plan's identity: the same in two chair sittings whose plans decide the same — everything here but epoch and final_sitting"`
	// Head is the events.id of the latest blue_edit or base_ingest — the report the parties audit.
	Head int64 `json:"head" min:"0" desc:"events.id of the report head the parties audit"`
	// Parties is who to engage; empty is the termination signal, and empty is [], never null.
	Parties []Party `json:"parties"`
	// Docket names the gaps docketed at impasse THIS chair sitting: the verb files a docket motion
	// for each under the chair's authorship the moment impasse is first computed, so "docketed" and
	// "at impasse" are one fact and no seat's discretion sits between a stalled gap and the bench.
	// Asked again in the same sitting, the plan still names them — the chair relays the plan it asks
	// for last, and blue owes closings on what this sitting docketed.
	Docket []string `json:"docket" desc:"gaps the verb docketed for the bench at this sitting"`
	// RemandOwed names the gaps whose remand's one more exchange this plan readies (routeRemandOwed),
	// each with the research direction the remand states: their minting lens and blue are among
	// Parties for it. The workflow tells blue the remand's duty for these gaps and no others, so the
	// duty lasts exactly as long as the exchange is owed — and it quotes THIS direction, the
	// record's, never the bench envelope's account of its own ruling.
	RemandOwed []RemandOwed `json:"remand_owed" desc:"gaps the bench remanded whose one more exchange this plan readies, each with the research direction the remand states: their minting lens and blue are among the parties for it"`
	// Exchanges is one entry per gap a party is engaged on, in the order the parties first name
	// them. It is the engine's measure of movement that the head does not carry: a gap marching to
	// impasse keeps its head and its parties while its count advances, so the engine's no-progress
	// valve keys on these and capture holds each to the record.
	Exchanges []ExchangeCount `json:"exchanges" desc:"one entry per gap a party is engaged on: the exchanges had on it, and the engaged sittings the record cannot close — not counted, and not zero"`
	// ToFile is the part of Docket with no docket motion yet: what the verb files. Never relayed.
	ToFile []string `json:"-"`
	// PassPermitted: a report is ingested, nobody is dispatched, and every blocker left on the gate's
	// list (PassBlockers) is the chair's own to clear this sitting — its inquiry review, its
	// spot-check of the stale areas, its ruling on each unruled motion whose gavel is the chair's
	// (grade, inquiry). Its work list marks those blocking until they are done.
	PassPermitted bool `json:"pass_permitted"`
	// StaleAreas is every retired lens whose pin the head has moved past. The chair reads
	// the changes since each pin against that area's duties and names the areas in its spot-check;
	// a PASS is refused until it has.
	StaleAreas []StaleArea `json:"stale_areas" desc:"each retired lens whose pin the report head has moved past: the chair reads the changes since each pin and names the areas in its spot-check before a PASS"`
	// Ceiling is the run at its limit, for one of two reasons EpochLimitReached tells apart: every
	// open material gap is at impasse after the one more exchange its remand granted and the bench
	// has remanded it again (remandStageOf), or the run's epoch limit is reached with parties still
	// ready.
	Ceiling bool `json:"ceiling"`
	// MaxEpochs is the run's epoch limit (Params.MaxEpochs); 0 when the run is held to none.
	MaxEpochs int `json:"max_epochs" min:"0" desc:"the run's epoch limit, a term setup records; 0 when the run is held to none"`
	// Epoch is the epoch the record has reached: the chair sittings opened so far, this one included.
	Epoch int `json:"epoch" key:"-" min:"0" desc:"the epoch the record has reached: the chair sittings opened so far, this one included"`
	// FinalSitting: this chair sitting opens the run's last epoch (isFinalEpoch), whoever the board
	// readies. No seat is dispatched after it, so of the motions filed in it only a petition is heard
	// in time to change how the run ends — by the bench at the exit — and `motion … file` refuses the
	// others there (HeardInTime).
	FinalSitting bool `json:"final_sitting" key:"-" desc:"this chair sitting opens the run's last epoch: no seat is dispatched after it, and of the motions filed now only a petition is heard, by the bench at the exit"`
	// EpochLimitReached: this chair sitting opens the run's last epoch with parties the board
	// readies, and they are not dispatched — the run ends CEILING. A last epoch with nobody ready
	// is not at the limit: it ends as any empty plan does.
	EpochLimitReached bool `json:"epoch_limit_reached" desc:"this chair sitting opens the run's last epoch: nobody is dispatched, and ceiling is set for that reason"`
	// Blockers is everything holding a PASS, from the list the verdict gate refuses on, each with the
	// seat that clears it. A blocker whose owner is not the chair readies that owner, once per cause
	// (Why says which); one still standing when the debate ends with the bench as its owner is what
	// the terminal bench sitting rules.
	//
	// EMPTY ONLY WHEN A PASS IS PERMITTED. PassPermitted is a report ingested, nobody ready and every
	// blocker the chair's; without it the plan has a party (each readied by a blocker on this list),
	// no report (no_report), or a blocker another seat owns. The engine's schema refuses a relay that
	// breaks it: o3's chair, refused a nested relay, re-sent the fields with `parties` and `blockers`
	// emptied, and the engine read the fabricated plan as the record's word that nobody sits.
	Blockers []PlanBlocker `json:"blockers" emptyonlywith:"pass_permitted" desc:"everything holding a PASS, from the list the verdict gate refuses on, with the seat whose act clears each; empty only when pass_permitted is true"`
	Why      []string      `json:"why"` // the readiness of each source, in words a reader can check against the board
}

type openGap struct {
	id, mintedBy, severity string
	// material is the gap view's column — the one definition — and classMaterial the class's
	// default, which the reason names when the class alone makes the gap not material.
	material      bool
	classMaterial string
	// unruledFiled is the gap view's `unruled_docket_filed`: the events.id of the newest docket
	// motion on the gap that no ruling names, 0 when every one is ruled. The plan's "a docket stands
	// unruled" is unruledFiled > 0, and the bench's sitting for it is keyed on that filing
	// (benchSatFor).
	unruledFiled  int64
	unruledMotion string // the motion id of that filing, for the plan's reason; "" when none
	// direction is the gap view's `docket_direction`: the research direction the latest remand
	// states, which the plan's reason quotes when the remand's exchange is owed.
	direction string
	// supersededBy is the successor that names this gap as an ancestor, when one does and this gap
	// is still open — the gap view's `stranded`. The PASS gate refuses a verdict over one.
	supersededBy string
}

// PlanDispatch computes readiness from four sources — each cast lens the report head has moved past, each
// open material gap below its limits, each docketed gap awaiting the bench, and the owner of each
// blocker another seat must clear (the bench for a motion, a lens for a contradiction it read,
// the evidence lens for a reopened citation or proof) — and the two derived facts the termination turns on. It writes
// nothing.
//
// ONE READ TRANSACTION, as a narrowed view reads (readSnapshot). The plan folds the events and reads
// each gap's openness, docket and materiality off the `gap` view; asked of the record at two moments,
// a docket motion filed between them readied the bench for a gap whose motion the plan's own blocker
// list did not hold. Seats write in parallel, so every question the plan asks is asked of the one
// snapshot its events came off.
//
// THE FOLD RUNS AFTER THE TRANSACTION CLOSES. The transaction reads (planReadsAt) and nothing else:
// held across the fold, it would hold the handle's one connection and the WAL read mark for as long
// as the CPU takes over the whole record.
func PlanDispatch(run Run) (Plan, error) {
	var r planReads
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = planReadsAt(q)
		return err
	}); err != nil {
		return emptyPlan(), err
	}
	return planFrom(r, filesOf(run))
}

// planFrom is the plan from what planReadsAt read, under the run's terms — refused where the record
// holds no cast, or where the fold breaks a rule the plan's declaration states.
func planFrom(r planReads, files runFiles) (Plan, error) {
	if r.cast == nil {
		return emptyPlan(), errNoCast
	}
	params, err := files.params()
	if err != nil {
		return emptyPlan(), err
	}
	plan := keyed(foldPlan(params, r))
	if err := relayable(plan); err != nil {
		return emptyPlan(), err
	}
	return plan, nil
}

// errNoCast is the dispatch's refusal of a record that holds no cast.
var errNoCast = errors.New("the record holds no cast. setup writes the run's admissible seats before any seat registers; a run without one has nothing to dispatch against")

// planReads is every answer the dispatch plan takes off the record, read on one snapshot
// (planReadsAt) and folded after it closes (foldPlan).
type planReads struct {
	cast    []string
	head    int64
	evs     []*Event
	win     WindowIndex
	fresh   map[string]bool
	gaps    []openGap
	remands map[string][]remandRow
}

// planReadsAt reads the plan's answers off q. A record with no cast stops at the cast: the plan
// refuses it, and asks nothing further.
func planReadsAt(q recordsql.Querier) (planReads, error) {
	var r planReads
	var err error
	if r.cast, err = castAt(q); err != nil || r.cast == nil {
		return r, err
	}
	if r.head, err = reportHeadAt(q); err != nil {
		return r, err
	}
	if r.evs, r.win, err = eventsAt(q); err != nil {
		return r, err
	}
	if r.fresh, err = freshMaterialOf(q); err != nil {
		return r, err
	}
	if r.gaps, err = openGaps(q); err != nil {
		return r, err
	}
	r.remands, err = remandRulingsOf(q)
	return r, err
}

// emptyPlan is the plan's empty shape. ARRAYS, NEVER null, AT THE SOURCE. The chair relays this
// JSON and the engine type-checks every field it reads; a nil slice marshals as null, which the relay
// refuses as a missing array. B9's chair relayed a PASS-permitted plan verbatim, "parties": null, and
// the engine aborted the run at the sitting that should have ended it.
func emptyPlan() Plan {
	return Plan{Parties: []Party{}, Docket: []string{}, RemandOwed: []RemandOwed{}, Exchanges: []ExchangeCount{}, Why: []string{}, StaleAreas: []StaleArea{}, Blockers: []PlanBlocker{}}
}

// foldPlan folds the plan from the answers planReadsAt read, under the run's terms. It asks the record
// nothing.
func foldPlan(params Params, r planReads) Plan {
	plan := emptyPlan()
	plan.Head = r.head
	evs, win, fresh := r.evs, r.win, r.fresh

	// Source 1: each cast lens that has never sat, or whose last sitting the head has moved past,
	// until it retires (retirement.go) — the same fold the PASS gate refuses from. A lens engaged
	// that has opened no sitting since has not sat, so its pin has not moved and it stays ready.
	ids := win.IDs(evs)
	dispatches, openings := dispatchLedger(evs, ids, win)
	benchOn := benchOccasionSittings(evs, ids, win)
	folds := lensStates(evs, ids, win, plan.Head, fresh)
	plan.StaleAreas = staleAreasOf(folds, plan.Head)
	parties := map[string][]string{}
	occasions := map[string][]string{}
	order := []string{}
	// engage readies a seat on gaps; occasion is what the bench is convened for, "" for every other
	// seat. A seat is one party however many sources ready it.
	engage := func(seat, occasion string, gaps ...string) {
		if _, seen := parties[seat]; !seen {
			order = append(order, seat)
			parties[seat] = []string{}
		}
		parties[seat] = append(parties[seat], gaps...)
		if occasion != "" && !slices.Contains(occasions[seat], occasion) {
			occasions[seat] = append(occasions[seat], occasion)
		}
	}
	for _, f := range folds {
		if f.ready {
			engage(f.seat, "")
		}
		plan.Why = append(plan.Why, f.why)
	}
	// The fold's roster is the cast through castOfEvents, which reads the same Cast event castAt read.

	// Source 2 and 3: the open gaps, with their materiality, their exchanges and their docket.
	gaps := r.gaps
	exch := exchangesOf(evs, ids, win, params, WhileRunning, r.remands)
	materialOpen, materialSettled := 0, 0
	statedMotion := map[string]bool{} // the docket motions a gap's reason below already names
	// debate readies a gap's two parties: its minting lens, and blue.
	debate := func(g openGap) {
		if g.mintedBy != "" {
			engage(g.mintedBy, "", g.id)
		}
		engage(blueRespondSeat, "", g.id)
	}
	// toBench dockets a gap at impasse: the chair files the motion, and the bench is convened for it.
	toBench := func(g openGap, why string) {
		plan.Docket = append(plan.Docket, g.id)
		plan.ToFile = append(plan.ToFile, g.id)
		engage(benchSeat, occasionDocket, g.id)
		plan.Why = append(plan.Why, why)
	}
	for _, g := range gaps {
		// A STRANDED GAP IS READY WORK WHATEVER ITS CLASS OR GRADE. Superseding is a promise to
		// replace, and the PASS gate refuses a verdict while the ancestor is open (refs.go) — so a
		// non-material ancestor nobody is dispatched to close would leave the plan saying "pass
		// permitted" and the gate saying no, forever, and the run ends UNVERIFIED with nobody ready.
		// Found by the release sweep. Held as material here: its minter and blue are engaged, its
		// exchanges count, and at impasse it reaches the bench like any other.
		stranded := g.supersededBy != ""
		x := exch[g.id]
		if x == nil {
			x = &GapExchanges{GapID: g.id}
		}
		// THE ROUTE IS routeOf, the one predicate the work lists read too.
		route := routeOf(g.material, stranded, g.unruledFiled > 0, x)
		if stranded && route != routeBench {
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open and superseded by %s — held as material until its minter closes it", g.id, g.supersededBy))
		}
		if g.material || stranded {
			materialOpen++
		}
		switch route {
		case routeBench:
			// A DOCKET MOTION IS THE ESCALATION ROUTE, and it readies the bench whether the gap is at
			// impasse or not. The dispatch files one at impasse; a party may file one earlier (`motion
			// docket file` is a red and blue verb, kept so the route to the bench is not the record's
			// discretion alone), and a trifle may be escalated too. Either way the motion stands until
			// the bench rules it, PASS is refused while it stands, and a bench that sat for it and
			// ruled nothing is not re-readied — the run cannot end in a verdict while it stands.
			statedMotion[g.unruledMotion] = true
			if benchSatFor(dispatches, benchOn, g.id, g.unruledFiled) {
				plan.Why = append(plan.Why, fmt.Sprintf("%s: docket motion %s stands unruled and the bench has sat since it was filed — one bench sitting per docketing, so this gap is not re-readied; the run cannot end in a verdict while it stands", g.id, g.unruledMotion))
				continue
			}
			engage(benchSeat, occasionDocket, g.id)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: docketed and unruled — the bench is ready", g.id))
			if docketedThisSitting(evs, ids, openings[chairSeat], g.unruledFiled) {
				plan.Docket = append(plan.Docket, g.id)
			}
		case routeNobody:
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open and not material (%s) — readies nobody", g.id, notMaterialBecause(g.classMaterial, g.severity)))
		case routeDebate:
			debate(g)
			if x.Remand == notRemanded {
				plan.Why = append(plan.Why, fmt.Sprintf("%s: open, material, %s — below its limits", g.id, x.Counted()))
			} else {
				plan.Why = append(plan.Why, fmt.Sprintf("%s: open, material, %s — the exchange its remand granted moved it, so its limits count from the bench's ruling, and it is below them", g.id, x.Counted()))
			}
		// AT IMPASSE, A GAP IS DOCKETED, OR OWED ITS REMAND'S EXCHANGE, OR AT ITS LIMIT. A gap the
		// bench has not remanded at impasse goes to the bench; one it remanded goes back to the debate
		// for one exchange between its minting lens and blue, carrying the ruling's direction, and back
		// to the bench if that exchange leaves it here; one remanded twice is at its limit.
		case routeDocket:
			toBench(g, fmt.Sprintf("%s: at impasse (%s) — docketed for the bench", g.id, x.Counted()))
		case routeRedocket:
			toBench(g, fmt.Sprintf("%s: at impasse (%s) after the exchange its remand granted — docketed for the bench again", g.id, x.Counted()))
		case routeRemandOwed:
			debate(g)
			plan.RemandOwed = append(plan.RemandOwed, RemandOwed{GapID: g.id, Direction: g.direction})
			plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse (%s) and remanded by the bench — its minting lens and blue are ready for the one more exchange the remand grants, on the ruling's direction: %s",
				g.id, x.Counted(), g.direction))
		case routeAtLimit:
			materialSettled++
			plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse and remanded again after the exchange its first remand granted — at its limit", g.id))
		}
	}
	// Source 4: THE SEAT THAT OWES ANOTHER SEAT'S BLOCKER SITS. A blocker on the gate's list that
	// the chair cannot clear — an unruled petition, a docket motion on a gap that has since closed,
	// a contradiction no finding raises, a citation or proof unchecked since its sentence changed
	// — holds the PASS until its owner acts, and an owner nobody
	// readies never acts: the run ended UNVERIFIED with the one seat that could clear it never asked
	// (#1202, #1203). Each owner is readied ONCE PER CAUSE: an owner that has sat since the blocker
	// arose and left it is not readied again, the reason says so, and the blocker still holds the
	// PASS — the precedent benchSatFor set for the docket.
	blockers := passBlockersOf(evs, ids, win, blockerGapsOfOpen(gaps), fresh)
	var reopenedSince int64 // the latest edit among the reopened blockers, all the evidence lens's
	for _, b := range blockers {
		switch {
		case b.Kind == BlockerReopened:
			reopenedSince = max(reopenedSince, b.since)
		case b.Kind == BlockerUnruledMotion && b.Owner == benchSeat:
			if statedMotion[b.Subject] {
				continue // a docket on an open gap: the gap's own reason above readied the bench
			}
			statedMotion[b.Subject] = true
			plan.Why = append(plan.Why, benchReadiness(b, dispatches, benchOn, engage))
		case b.Kind == BlockerContradiction && seatclass.RoleOfSeat(b.Owner) == "lens":
			if _, sat := firstAfter(openings[b.Owner], b.since); sat {
				plan.Why = append(plan.Why, fmt.Sprintf("%s has sat since it read a source contradicting %q (%s), and no finding raises it — readied once per contradiction, so not again", b.Owner, b.Detail, contradictionFact(b)))
				continue
			}
			engage(b.Owner, "")
			plan.Why = append(plan.Why, fmt.Sprintf("%s read a source contradicting %q (%s) and no finding raises it — ready, to raise it", b.Owner, b.Detail, contradictionFact(b)))
		}
	}
	// ONE SITTING FOR EVERY REOPENED ANCHOR, ONCE PER REOPENING. "Sat" is a sitting the lens opened
	// after the latest edit — the same opening a clearing check must come from (reopenedEvidence),
	// so a sitting that could not have cleared the item never spends the readying. What the lens
	// does in that sitting is not read: one that opens and checks nothing has still sat, the
	// blocker stands, and the run ends naming the anchor rather than readying the lens for ever.
	if reopenedSince > 0 {
		if _, sat := firstAfter(openings[evidenceSeat], reopenedSince); sat {
			plan.Why = append(plan.Why, fmt.Sprintf("%s has sat since edit %d and a citation or proof it reopened is still unchecked — readied once per reopening, so not again", evidenceSeat, reopenedSince))
		} else {
			engage(evidenceSeat, "")
			plan.Why = append(plan.Why, fmt.Sprintf("%s has not sat since edit %d, which left a citation or proof unchecked since its sentence changed — ready, to check each as it now stands", evidenceSeat, reopenedSince))
		}
	}
	for _, s := range order {
		plan.Parties = append(plan.Parties, Party{SeatID: s, GapIDs: parties[s], Occasions: occasions[s]})
	}
	// PASS_PERMITTED IS THE GATE'S OWN LIST, read through passBlockersOf: nobody is ready, and every
	// blocker left is the chair's to clear in this sitting (its inquiry review, its spot-check of
	// the stale areas, its ruling on a grade or inquiry motion). A blocker another seat must clear —
	// an unruled petition, a contradiction no lens has raised, a docket motion on a gap that has
	// since closed, a citation or proof unchecked since its sentence changed — holds it, and its
	// reason is stated here, so the plan never says PASS where the gate refuses one the chair
	// cannot clear (#1202).
	chairOnly := true
	for _, b := range blockers {
		chairOnly = chairOnly && b.ChairOwned()
		plan.Blockers = append(plan.Blockers, PlanBlocker{Kind: b.Kind, Subject: b.Subject, Owner: b.Owner})
		if why := kindOf(b.Kind).why; why != nil && !(b.Kind == BlockerUnruledMotion && statedMotion[b.Subject]) {
			plan.Why = append(plan.Why, why(b))
		}
	}
	plan.PassPermitted = plan.Head > 0 && len(plan.Parties) == 0 && chairOnly
	plan.Ceiling = len(plan.Parties) == 0 && materialOpen > 0 && materialSettled == materialOpen

	// THE EPOCH LIMIT IS A TERM OF THE RUN, read like k and kMax. The chair sitting that opens the
	// last epoch dispatches nobody: the parties the board readies stay undispatched and the run
	// ends CEILING. A board that permits PASS there readied nobody, so it is not at the limit and
	// the chair records the PASS as ever. The docket stands — filing a motion dispatches no seat,
	// and the terminal bench rules what stays unruled at the exit.
	plan.MaxEpochs = params.MaxEpochs
	plan.Epoch = win.LastEpoch(evs)
	plan.FinalSitting = isFinalEpoch(params, plan.Epoch)
	if plan.FinalSitting && len(plan.Parties) > 0 {
		plan.Why = append(plan.Why, fmt.Sprintf("epoch limit %d reached — this chair sitting opens the run's last epoch, so the %d party(ies) above are not dispatched", params.MaxEpochs, len(plan.Parties)))
		plan.Parties = []Party{}
		plan.EpochLimitReached, plan.Ceiling = true, true
	}
	// AFTER THE LIMIT HAS EMPTIED THE PARTIES: an entry is a gap a party of THIS plan is engaged on.
	seen := map[string]bool{}
	for _, p := range plan.Parties {
		for _, g := range p.GapIDs {
			if !seen[g] {
				seen[g] = true
				plan.Exchanges = append(plan.Exchanges, exchangeCountOf(g, exch[g]))
			}
		}
	}
	return plan
}

// isFinalEpoch reports whether epoch is the run's last under its epoch limit: the chair sitting that
// opens epoch MaxEpochs is the run's final sitting, and no seat is dispatched after it. The plan
// reads it for who is dispatched, the verdict for CEILING after the exit, and the write path for the
// motions nobody is left to hear. A run held to no limit has no final sitting.
func isFinalEpoch(params Params, epoch int) bool {
	return params.MaxEpochs > 0 && epoch >= params.MaxEpochs
}

// HeardAtTheExit reports whether a motion of subj filed at the run's final sitting still reaches a
// seat that can change how the run ends. A petition does: the terminal bench hears a petition standing
// at the exit, and it may halt the run there — a petition is the channel for an objection that the run
// must not go on, and the last sitting is no exception. A docket motion does not: the terminal bench
// rules it, and the run ends CEILING however it rules (epochLimitHeldTheBench). A grade motion does
// not: the chair rules it, and the chair does not sit again.
func HeardAtTheExit(subj recordpb.MotionSubject) bool {
	return subj == recordpb.MotionSubject_MOTION_SUBJECT_PETITION
}

// HeardInTime is the guard of a seat's motion filing: at the run's final sitting a motion of a subject
// not HeardAtTheExit is refused and nothing is written. The chair rules no motion after that sitting
// and the bench is convened for none before the exit, so such a motion holds the PASS until the run
// ends, and a docket motion ends the run CEILING however the terminal bench rules it
// (epochLimitHeldTheBench). Asked under the write lock, as every verb's check is. The docket the
// dispatch files at impasse is not a seat's motion and passes no such guard: it is the terminal
// bench's to rule.
func HeardInTime(subj recordpb.MotionSubject) Guard {
	return Check(func(l *Landing) error {
		if HeardAtTheExit(subj) {
			return nil
		}
		epochs, err := epochsAt(l.w.q)
		if err != nil || len(epochs) == 0 {
			return err
		}
		params, err := l.w.files.params()
		if err != nil {
			return err
		}
		epoch := epochs[len(epochs)-1]
		if !isFinalEpoch(params, epoch) {
			return nil
		}
		word := recordpb.Word(subj)
		return feov.Errorf(feov.Conflict, "a %s motion filed at the run's final sitting is not heard in time: the record is at epoch %d, the last under the run's epoch limit of %d, and no seat is dispatched after this sitting. "+
			"Recorded, it would hold the PASS until the run ends, and no ruling on it could change how the run ends. Nothing was recorded. "+
			"At the final sitting the chair records the verdict the board permits — a PASS where the plan permits one, or a FAIL — or none, and the run's outcome is derived from the record as it stands. "+
			"What the motion would have argued goes in a position, which the transcript carries to the bench at the exit; an objection that the run must not go on is a petition, which the bench hears at the exit", word, epoch, params.MaxEpochs)
	})
}

// epochLimitHeldTheBench reports whether the chair sitting that opened the run's last epoch left
// the bench ready for a motion the bench has ruled since — the epoch limit, read after the exit.
//
// THE LIMIT IS REACHED AT THE CHAIR'S SITTING, AND THE BENCH RULES AFTER IT. foldPlan asks whether
// anybody is ready NOW, which is the question the chair's verb asks at the sitting. The terminal
// bench sitting then rules every motion whose gavel is the bench's, and `bench outcome` derives the
// verdict after that: where the bench was the only party the limit left undispatched — one gap at
// impasse, docketed at the last sitting, or one petition — its ruling leaves nobody ready, the plan
// stops saying the limit was reached, and the CEILING the engine states is refused as a word the
// record cannot derive.
//
// The bench was ready for a motion when it had not had the sitting dispatch convenes it for —
// benchSatFor for a docket, a petition sitting for a petition — and no ruling changes that: the
// sittings the engine convenes at the exit answer no dispatch and state neither occasion. A motion
// the bench had already sat for and left readied nobody, so its ruling at the exit moves nothing
// here. A ruling counts when it follows the chair's last sitting opening; at the limit nobody is
// dispatched after that sitting, so every such ruling is the exit's.
func epochLimitHeldTheBench(params Params, r planReads) bool {
	evs, win := r.evs, r.win
	if !isFinalEpoch(params, win.LastEpoch(evs)) {
		return false
	}
	ids := win.IDs(evs)
	dispatches, openings := dispatchLedger(evs, ids, win)
	benchOn := benchOccasionSittings(evs, ids, win)
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil {
			continue
		}
		// Ruled since the chair's last sitting opened: an unruled motion's place is 0, before any.
		if _, chairSatSince := firstAfter(openings[chairSeat], m.ruled); chairSatSince {
			continue
		}
		subj, _ := MotionSubjectEnum(m.Subject) // benchSittingFor holds the bench-ruled subjects only
		switch benchSittingFor[subj] {
		case recordpb.Occasion_OCCASION_PETITION:
			if !benchSatOnOccasionSince(benchOn, recordpb.Occasion_OCCASION_PETITION, m.filed) {
				return true
			}
		case recordpb.Occasion_OCCASION_DOCKET:
			if !benchSatFor(dispatches, benchOn, m.GapID, m.filed) {
				return true
			}
		}
	}
	return false
}

// dispatchRow is one dispatch event: where it sits in the stream, the head it pinned, the seat it
// named and the gaps it engaged that seat on.
type dispatchRow struct {
	at, pin int64
	seat    string
	gaps    []string
	// occasions is what a bench row convened the bench for; empty on every other seat's row.
	occasions []recordpb.Occasion
}

// dispatchLedger reads the stream once for what "did this seat sit" is decided from: every
// dispatch row, and each seat's openings, in stream order. seq gives each event's place — the
// events."id" where the reader has them, the position where it holds only the stream. The
// predicate compares order and nothing else, so either answers it the same.
//
// THE OPENINGS ARE THE EVENTS THAT OPENED A SITTING, as the write path stored them (win): a hook's
// sitting_open, or a register no bracket of its agent preceded. A register under an agent the hook
// already bracketed JOINED that sitting, so a bracket and its paired register are one opening here,
// not two: read as two, the register would close the bracket's sitting, and every act after it
// would land in a sitting that answers no dispatch.
func dispatchLedger(evs []*Event, seq []int64, win WindowIndex) ([]dispatchRow, map[string][]int64) {
	var ds []dispatchRow
	openings := map[string][]int64{}
	for i, e := range evs {
		// A SITTING OPENS EITHER WAY, and the seat is whichever one this event opens a sitting for
		// — the registering seat, or the seat a hook's sitting_open resolves to through its
		// configuration. The second arm is how a no-op sitting costs nothing: every reader of "has
		// this seat sat" comes through this map, so a hook-opened sitting satisfies the dispatch,
		// the pin, the lens fold and the work list without the seat running a command.
		if seat, opens := win.Opens(e); opens {
			openings[seat] = append(openings[seat], seq[i])
		}
		if b, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok {
			ds = append(ds, dispatchRow{at: seq[i], pin: b.GetPin(), seat: b.GetSeatId(), gaps: b.GetGapIds(), occasions: b.GetOccasions()})
		}
	}
	return ds, openings
}

// firstAfter is the first of an ascending sequence that follows at.
func firstAfter(xs []int64, at int64) (int64, bool) {
	k := sort.Search(len(xs), func(i int) bool { return xs[i] > at })
	if k == len(xs) {
		return 0, false
	}
	return xs[k], true
}

// sittingFor IS THE ONE ANSWER TO "HAS THIS SEAT SAT FOR WHAT IT WAS DISPATCHED FOR": its sitting
// for a dispatch is the first sitting it opened after the dispatch row (dispatchLedger: the hook's
// bracket, or a register no bracket preceded), and a seat that has not opened one
// since has not sat. Every reader of the question asks it here — the lens's readiness fold
// (lensStates) and its last sitting (lastSittingBefore), the bench's sitting for a docketing
// (benchSatFor), the exchange count (exchangesOf), and the seat's own work list (owedSitting). An unsat dispatch is why dispatch
// readies the seat again, and it is why the seat's work list is not complete.
//
// A BENCH ROW IS SAT ONCE PER OCCASION, and benchRowSitting answers it. The bench is convened for a
// petition and a docket on ONE row and sits twice, petition first; the first sitting it opens after
// the row is the petition sitting, and read here it would count as the docket's too — a docket the
// bench never sat for would read as heard and never be readied again.
func sittingFor(openings []int64, d dispatchRow) (int64, bool) {
	return firstAfter(openings, d.at)
}

// benchOccasionSittings is the bench's sittings, by the occasion each one's register states, in stream
// order — the one ledger every reader of "has the bench sat for X" keys on. seq is the same place
// sequence dispatchLedger reads (events.id, or the stream position), so a filing's place and a
// sitting's compare under one rule. The place is the sitting's OPENING, the place dispatchLedger
// holds for it: a bracketed bench states its occasion at the register that joined the bracket's
// sitting, so the register's own place would be a second start for one sitting.
func benchOccasionSittings(evs []*Event, seq []int64, win WindowIndex) map[recordpb.Occasion][]int64 {
	place := placesOf(evs, seq, win)
	on := map[recordpb.Occasion][]int64{}
	for _, e := range evs {
		if e.GetSeatId() != benchSeat {
			continue
		}
		r, ok := recordpb.BodyAs[*recordpb.Register](e)
		if !ok || r.Occasion == nil {
			continue
		}
		if p, sat := place[win.Of(e).SittingID]; sat {
			on[r.GetOccasion()] = append(on[r.GetOccasion()], p)
		}
	}
	return on
}

// placesOf maps each event's row id to its place in seq, so a stored sitting_id — an event id —
// can be compared with the places a fold reads.
func placesOf(evs []*Event, seq []int64, win WindowIndex) map[int64]int64 {
	out := make(map[int64]int64, len(evs))
	for i, e := range evs {
		out[win.Of(e).ID] = seq[i]
	}
	return out
}

// benchRowSitting is the bench's sitting for a row: the opening of its first sitting OF EACH
// OCCASION the row convened it for (the occasion its register states), after the row. It has sat
// for the row once it has sat for every one of them, and the place returned is the latest. A row
// with no occasion is not a bench row, and reads sittingFor's way.
func benchRowSitting(openings []int64, on map[recordpb.Occasion][]int64, d dispatchRow) (int64, bool) {
	if len(d.occasions) == 0 {
		return sittingFor(openings, d)
	}
	var last int64
	for _, o := range d.occasions {
		r, ok := firstAfter(on[o], d.at)
		if !ok {
			return 0, false
		}
		last = max(last, r)
	}
	return last, true
}

// sittingCloser IS THE ONE ANSWER TO "WHERE DID THIS SEAT'S SITTING END", as sittingFor is to where
// it began. Every reader that bounds a seat's sitting asks it here — the exchange count
// (exchangesOf) and blue's sittings (BlueSittings, behind PositionSittings, the manifest's owed set and
// capture's record-parity audit) — so a change to what closes a sitting reaches all of them at once.
//
// ONLY FACTS ABOUT THIS SEAT CLOSE IT, AND THE EARLIER OF TWO DOES. A sitting ends at whichever
// comes first after it began:
//
//   - the opening of the seat's next sitting, as the write path stored it — the hook's bracket, or
//     a register no bracket preceded. A register that joined its own bracket's sitting is not one:
//     it is the same sitting, and closing there would cut the sitting at its own register.
//   - the stop of the agent that sat it: a sitting_close whose agent_id is the one on the event
//     that opened THIS sitting. The SubagentStop hook writes it when that agent returns, so it is
//     the harness's observation of this seat, not an act of another seat.
//
// NOTHING ANOTHER SEAT WRITES BOUNDS IT. A read keyed on another seat's act — the chair's next
// register, the chair's next dispatch row — depends on a rule enforced at that seat's write path,
// and a warm chair registers once per run: on five archived runs a chair-keyed exchange fold read
// every party sitting as still open and no gap could reach the bench by impasse (#1002).
//
// A STOP THAT JOINS NO OPENING CLOSES NOTHING. The hook records every typed subagent in a project
// whose run marker is live, a developer's own subagents included, and a register with no agent_id
// (a run the PreToolUse hook never reached) has nothing a stop can join. A headless seat run as a
// `claude -p` main session fires no SubagentStop at all, so its sittings close by their next
// opening alone.
//
// WITH NEITHER PAST IT, WHAT THE SITTING IS DEPENDS ON WHEN THE RECORD IS READ (ReadWhen), and the
// reader says which — the closer never guesses it:
//
//   - WhileRunning: the sitting is UNRESOLVED, not complete and not absent. It may be in flight, or
//     it may have ended with nothing recorded after it, and the record holds no fact that tells the
//     two apart. Each reader states that case as its own answer.
//   - AfterTheRun: the end of the record closes it. A finished run has no sitting in flight, so the
//     last act on the record is the last act of every sitting still open — that is the premise of
//     the read, not an act of another seat.
type sittingCloser struct {
	openings  map[string][]int64 // seat -> the places of the events that opened its sittings, ascending
	agentOf   map[int64]string   // a bracket's or register's place -> the agent_id it carries
	stops     map[string][]int64 // agent_id -> the places of its sitting_close events, ascending
	recordEnd int64              // one past the last place on the record
	when      ReadWhen
}

// ReadWhen is when a reader reads the record: while the run can still be sitting, or after it
// ended. It decides one thing — whether a sitting nothing about its seat has closed is unresolved
// or closed by the end of the record (sittingCloser) — and every reader that bounds a sitting
// takes it from where it runs, never from a default.
type ReadWhen int

const (
	// WhileRunning is a read made while a sitting may be in flight: a seat's work list, the
	// chair's dispatch plan, a live view.
	WhileRunning ReadWhen = iota
	// AfterTheRun is a read of a finished run, which is capture's alone: capture runs once the run
	// has ended, so no sitting is still in flight.
	AfterTheRun
)

// sittingCloserOf reads the stream once for every fact that can close a sitting. seq and openings
// are dispatchLedger's: the same places the sitting's start was read at.
func sittingCloserOf(evs []*Event, seq []int64, win WindowIndex, openings map[string][]int64, when ReadWhen) sittingCloser {
	c := sittingCloser{openings: openings, agentOf: map[int64]string{}, stops: map[string][]int64{}, when: when}
	if n := len(seq); n > 0 {
		c.recordEnd = seq[n-1] + 1
	}
	for i, e := range evs {
		// THE AGENT THAT SAT IT is on whichever event opened the sitting — the hook's bracket, or a
		// register no bracket preceded.
		if a := recordpb.AgentOpening(e); a != "" {
			c.agentOf[seq[i]] = a
		}
		if b, ok := recordpb.BodyAs[*recordpb.SittingClose](e); ok && b.GetAgentId() != "" {
			c.stops[b.GetAgentId()] = append(c.stops[b.GetAgentId()], seq[i])
		}
	}
	return c
}

// span is a sitting's acts: the places from the event that opened it up to, not including, the act
// that closed it.
type span struct{ from, to int64 }

// holds reports whether a place falls in any of the spans.
func holds(spans []span, at int64) bool {
	for _, s := range spans {
		if at >= s.from && at < s.to {
			return true
		}
	}
	return false
}

// bounds is the seat's sitting that began at start: the one span its acts fall in, where it ends,
// and whether it is closed.
func (c sittingCloser) bounds(seat string, start int64) (spans []span, end int64, closed bool) {
	end, closed = c.end(seat, start)
	return []span{{from: start, to: end}}, end, closed
}

// end is where the sitting that began at from ended: the place of the act that closed it, which is outside it. With nothing about
// this seat past it, end is the end of the record, and closed is false while the run is read as
// running and true after it.
func (c sittingCloser) end(seat string, from int64) (int64, bool) {
	end, closed := firstAfter(c.openings[seat], from)
	if agent := c.agentOf[from]; agent != "" {
		if stop, stopped := firstAfter(c.stops[agent], from); stopped && (!closed || stop < end) {
			end, closed = stop, true
		}
	}
	if !closed {
		return c.recordEnd, c.when == AfterTheRun
	}
	return end, true
}

// DispatchGroup is one chair sitting's dispatch as the record delimits it: the dispatch rows
// written with NO SITTING OPENED BETWEEN THEM, by any seat. Somebody opening a sitting is somebody
// sitting — the chair opening a sitting, or a party sitting for what the chair dispatched — and the
// workflow sequences the two, so a dispatch row after an opening belongs to a later chair sitting,
// and one before any opening belongs to the same.
//
// NOT THE EPOCH, and the difference is measured. The epoch counts the chair's stored sittings, and
// a warm chair — a later sitting resuming the same session — opens none: it registered ONCE per run
// on all four archived is-91-prime B runs (B3–B6), so every chair sitting read as sitting 1 and
// every dispatch of the run as one. Nor one row set per `dispatch next`: the B5 and
// B6 chairs each ran the verb twice in their first sitting, once for the prose and once for
// --json, and wrote the plan twice, 4 s apart, with nobody sitting in between.
type DispatchGroup struct {
	First, Last int      // stream positions of the group's first and last dispatch row
	Parties     []string // each seat the group names, in the order first named
	// Sat is each party's sitting for the group, off the last row naming it: the stream position of
	// the first sitting it opened after that row (sittingFor) — for the bench, of the opening of
	// its sitting for each occasion the row convened it for, the latest of them (benchRowSitting).
	// A party absent here has not sat.
	Sat map[string]int
	// PartyRows is each party's LAST row in the group — the row Sat reads: a plan that files a
	// docket is never standing, so a chair that asks twice can leave two rows for one party, and
	// the later is the dispatch.
	PartyRows map[string]PartyRow
	rows      []dispatchRow
}

// PartyRow is one party's dispatch row as recorded: the head it was pinned to and the gaps it
// engaged the party on.
type PartyRow struct {
	Pin    int64
	GapIDs []string
	// Occasions is what the row convened the bench for, in the enum's words; empty for every other
	// seat. Capture holds the bench's registers of each to it.
	Occasions []string
}

// DispatchGroups is the record's dispatches, grouped as DispatchGroup says, in stream order.
func DispatchGroups(evs []*Event, win WindowIndex) []DispatchGroup {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, openings := dispatchLedger(evs, seq, win)
	on := benchOccasionSittings(evs, seq, win)
	var anyone []int64
	for _, rs := range openings {
		anyone = append(anyone, rs...)
	}
	sort.Slice(anyone, func(a, b int) bool { return anyone[a] < anyone[b] })
	var groups []DispatchGroup
	for i, d := range ds {
		if i == 0 || openedBetween(anyone, ds[i-1].at, d.at) {
			groups = append(groups, DispatchGroup{First: int(d.at), Sat: map[string]int{}, PartyRows: map[string]PartyRow{}})
		}
		g := &groups[len(groups)-1]
		g.Last = int(d.at)
		g.rows = append(g.rows, d)
	}
	for k := range groups {
		g := &groups[k]
		last := map[string]dispatchRow{}
		for _, d := range g.rows {
			if _, seen := last[d.seat]; !seen {
				g.Parties = append(g.Parties, d.seat)
			}
			last[d.seat] = d
		}
		for p, d := range last {
			g.PartyRows[p] = PartyRow{Pin: d.pin, GapIDs: append([]string{}, d.gaps...), Occasions: wordsOf(d.occasions)}
			if r, ok := benchRowSitting(openings[p], on, d); ok {
				g.Sat[p] = int(r)
			}
		}
	}
	return groups
}

// openedBetween reports whether any of the ascending openings falls strictly between a and b.
func openedBetween(openings []int64, a, b int64) bool {
	r, ok := firstAfter(openings, a)
	return ok && r < b
}

// chairSeat is the seat whose sittings are the record's epochs.
const chairSeat = "red-chair"

// benchSeat is the bench's seat — the one that holds the gavel for the subjects the schema gives
// the bench.
const benchSeat = "judge"

// evidenceSeat is the evidence lens's seat — the one that owes the check of a reopened citation or
// proof and is readied for it. Another lens's check from a sitting opened after the edit clears
// the anchor as well (reopenedEvidence).
const evidenceSeat = "red-lens-evidence"

// blueRespondSeat is blue's responding seat — the one the chair dispatches onto gaps.
const blueRespondSeat = "blue-respond"

// PartySeats are the seats a plan's party can name, in the two shapes a party takes: the bench,
// which carries occasions, and every lens seat on the roster with blue's responding seat, which
// carry none. The chair's envelope schema in debate.js states the same two lists: devcmd/enginegen
// writes its two party shapes from this one.
func PartySeats() (bench string, lensOrBlue []string) {
	lensOrBlue = append(seatclass.SeatsOfRole("lens"), blueRespondSeat)
	sort.Strings(lensOrBlue)
	return benchSeat, lensOrBlue
}

// unopenedChairSitting is the chair's latest dispatch when a party has SAT for it (sittingFor) and
// no sitting of the chair's has opened since recording it. The workflow comes back to the chair
// only after the parties sit, so that is a new chair sitting, and the record holds it only once
// something opens it: the hook's bracket, or the chair's register where no bracket did. Unopened,
// its acts land in the previous epoch, and the dispatch groups would have only the parties'
// sittings to split on. A resumed sitting carries no bracket and opens by its register alone; that
// register is the chair's owed one.
//
// AN OPENING IS THE STORED ONE (win.Opens), the same answer the epoch reads: a chair sitting the
// bracket opened is the next epoch on the record, so the gate accepts it as opened.
//
// THE EXCHANGE COUNT DOES NOT DEPEND ON IT. exchangesOf closes a party's sitting at that seat's OWN
// next opening or its own agent's stop, whichever is first (#1002). This item stands on its own
// ground — the epoch and the groups — and the fold stands on the parties'.
func unopenedChairSitting(evs []*Event, win WindowIndex) (DispatchGroup, bool) {
	groups := DispatchGroups(evs, win)
	if len(groups) == 0 {
		return DispatchGroup{}, false
	}
	g := groups[len(groups)-1]
	if len(g.Sat) == 0 {
		return DispatchGroup{}, false // nobody has sat for it: still the sitting that recorded it
	}
	for i := len(evs) - 1; i > g.Last; i-- {
		if seat, opens := win.Opens(evs[i]); opens && seat == chairSeat {
			return DispatchGroup{}, false
		}
	}
	return g, true
}

// requireChairSittingOpened refuses `dispatch next` in a chair sitting nothing opened (see
// unopenedChairSitting), over the stream DispatchNext read in its transaction. It is the write-time
// half of the chair's "register for this sitting" work item: the verb is the chair's first act every
// sitting, so refusing it there puts the register ahead of every act a sitting no bracket opened
// records.
func requireChairSittingOpened(evs []*Event, win WindowIndex) error {
	g, owed := unopenedChairSitting(evs, win)
	if !owed {
		return nil
	}
	return fmt.Errorf("%s", chairRegisterOwed(g))
}

// chairRegisterOwed is the one wording of the chair's owed register, for the refusal and the work
// list alike.
func chairRegisterOwed(g DispatchGroup) string {
	sat := make([]string, 0, len(g.Sat))
	for p := range g.Sat {
		sat = append(sat, p)
	}
	sort.Strings(sat)
	return fmt.Sprintf("%s sat for the dispatch you recorded against report head %d, and no sitting of yours has opened since. The workflow came back to you, so this is a new sitting, and no bracket opened it — a resumed session's sitting is opened by your register alone. Register for this sitting, then ask again",
		strings.Join(sat, ", "), g.rows[0].pin)
}

// dispatchStands reports whether plan is already the record's standing dispatch: no sitting has
// opened since the latest dispatch rows — no bracket, no register opening one, by any seat — and
// those rows name exactly plan's parties, on the
// same gaps, at plan's head. `dispatch next` then records nothing new — asking twice in one
// sitting, for the prose and then for --json, is one decision, and the B5 and B6 chairs each wrote
// it twice. A plan that differs is a new decision and is recorded. It reads the stream DispatchNext
// read in the transaction the plan was folded in and the rows would be written in.
func dispatchStands(evs []*Event, win WindowIndex, plan Plan) bool {
	if len(plan.Parties) == 0 || len(plan.ToFile) > 0 || evs == nil {
		return false
	}
	groups := DispatchGroups(evs, win)
	if len(groups) == 0 {
		return false
	}
	g := groups[len(groups)-1]
	for _, e := range evs[g.Last+1:] {
		if _, opens := win.Opens(e); opens {
			return false
		}
	}
	// The occasions are part of the decision: the bench convened for a petition as well as its
	// docket is a new decision, on the same gaps. remand_owed is not in the key, and need not be: a
	// gap's remand exchange becomes owed only by a bench ruling, and the bench opens a sitting to
	// rule, which ends the standing decision above before any key is compared.
	key := func(seat string, pin int64, gaps, occasions []string) string {
		gs := append([]string{}, gaps...)
		sort.Strings(gs)
		os := append([]string{}, occasions...)
		sort.Strings(os)
		return fmt.Sprintf("%s@%d:%s/%s", seat, pin, strings.Join(gs, ","), strings.Join(os, ","))
	}
	standing := map[string]bool{}
	for _, d := range g.rows {
		standing[key(d.seat, d.pin, d.gaps, wordsOf(d.occasions))] = true
	}
	asked := map[string]bool{}
	for _, p := range plan.Parties {
		asked[key(p.SeatID, plan.Head, p.GapIDs, p.Occasions)] = true
	}
	if len(asked) != len(standing) {
		return false
	}
	for k := range asked {
		if !standing[k] {
			return false
		}
	}
	return true
}

// owedSitting is the latest dispatch naming seatID that the seat has not sat for, by sittingFor.
// The stream's positions stand in for events.id: evs is in id order.
func owedSitting(evs []*Event, win WindowIndex, seatID string) (dispatchRow, bool) {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, openings := dispatchLedger(evs, seq, win)
	on := benchOccasionSittings(evs, seq, win)
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].seat != seatID {
			continue
		}
		_, sat := benchRowSitting(openings[seatID], on, ds[i])
		return ds[i], !sat
	}
	return dispatchRow{}, false
}

// occasionDocket and occasionPetition are the Occasion enum's words for the two sittings dispatch
// convenes the bench for. The workflow routes a bench party on them, and the bench types the same
// word at its register.
var (
	occasionDocket   = recordpb.Word(recordpb.Occasion_OCCASION_DOCKET)
	occasionPetition = recordpb.Word(recordpb.Occasion_OCCASION_PETITION)
)

// benchSittingFor is the sitting dispatch convenes the bench for to rule a motion of each subject
// whose gavel the schema gives the bench. A bench-ruled subject with no row has no route to a
// ruling, and TestEveryBenchRuledSubjectHasASitting fails naming it.
var benchSittingFor = map[recordpb.MotionSubject]recordpb.Occasion{
	recordpb.MotionSubject_MOTION_SUBJECT_PETITION: recordpb.Occasion_OCCASION_PETITION,
	recordpb.MotionSubject_MOTION_SUBJECT_DOCKET:   recordpb.Occasion_OCCASION_DOCKET,
}

// benchReadiness readies the bench for one unruled motion whose gavel is the bench's and that no
// open gap's docket has already readied it for, and returns the plan's reason.
//
// A PETITION IS HEARD AT A PETITION SITTING, AND ONLY ONE COUNTS. The guard is a bench sitting
// opened after the filing whose register states occasion `petition` — not any bench sitting: a
// docket sitting that sat after the filing was convened for something else, and counting it would
// leave the petition unheard while the plan said it had been.
func benchReadiness(b Blocker, dispatches []dispatchRow, on map[recordpb.Occasion][]int64, engage func(seat, occasion string, gaps ...string)) string {
	subj, _ := MotionSubjectEnum(b.motionSubject)
	switch occ, routed := benchSittingFor[subj]; {
	case routed && occ == recordpb.Occasion_OCCASION_PETITION:
		if benchSatOnOccasionSince(on, recordpb.Occasion_OCCASION_PETITION, b.since) {
			return fmt.Sprintf("%s: petition, unruled, and the bench has sat to hear petitions since it was filed — one petition sitting per filing, so it is not convened again; the run cannot end in a verdict while it stands", b.Subject)
		}
		engage(benchSeat, occasionPetition)
		return fmt.Sprintf("%s: petition, unruled — the bench is ready to hear it, before any party of this epoch sits", b.Subject)
	case routed && occ == recordpb.Occasion_OCCASION_DOCKET:
		// An open gap's docket was readied above; this one's gap has closed since it was filed.
		if benchSatFor(dispatches, on, b.About, b.since) {
			return fmt.Sprintf("%s: docket motion %s stands unruled on a gap no longer open, and the bench has sat since it was filed — not re-readied; the run cannot end in a verdict while it stands", b.About, b.Subject)
		}
		engage(benchSeat, occasionDocket, b.About)
		return fmt.Sprintf("%s: docket motion %s stands unruled on a gap no longer open — the bench is ready", b.About, b.Subject)
	}
	return fmt.Sprintf("%s: %s motion, unruled, and no sitting convenes the bench for its subject — nobody is readied; the run cannot end in a verdict while it stands", b.Subject, b.motionSubject)
}

// dispatchOccasions is what the chair convenes the bench for, in schema order: a dispatch row
// names these and no other. The terminal and assembly sittings are the engine's own.
var dispatchOccasions = []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET, recordpb.Occasion_OCCASION_PETITION}

// DispatchOccasions is the words a plan's bench party can carry in `occasions` — the set
// requireDispatchOccasions holds a dispatch row to, which the engine script's party schema is
// generated from.
func DispatchOccasions() []string { return wordsOf(dispatchOccasions) }

// requireDispatchOccasions refuses a dispatch row whose occasions disagree with its seat, at the
// write. The bench's row says what it is convened for — `docket`, `petition`, or both — because its
// one id cannot; no other seat's row carries an occasion. The two the chair convenes are the only
// ones a row may name: the terminal and assembly sittings are the engine's, and no dispatch answers
// for them. A docket is the bench engaged on gaps, so `docket` and a non-empty gap list are one
// fact, and a row stating only one of them is refused rather than read either way.
func requireDispatchOccasions(b *recordpb.Dispatch) error {
	occ := b.GetOccasions()
	if !SeatOwesOccasion(b.GetSeatId()) {
		if len(occ) > 0 {
			return fmt.Errorf("%q is convened for no occasion; only the bench's row says what it sits for", b.GetSeatId())
		}
		return nil
	}
	if len(occ) == 0 {
		return fmt.Errorf("a bench row names what the bench is convened for (%s, %s, or both); its one seat id cannot say", occasionDocket, occasionPetition)
	}
	seen := map[recordpb.Occasion]bool{}
	for _, o := range occ {
		if !slices.Contains(dispatchOccasions, o) {
			return fmt.Errorf("the chair convenes the bench for %s or %s, and %s is a sitting the engine convenes", occasionDocket, occasionPetition, recordpb.Word(o))
		}
		if seen[o] {
			return fmt.Errorf("the row names %s twice; the bench sits once per occasion", recordpb.Word(o))
		}
		seen[o] = true
	}
	if docket, gaps := seen[recordpb.Occasion_OCCASION_DOCKET], len(b.GetGapIds()) > 0; docket != gaps {
		return fmt.Errorf("a docket sitting is the bench engaged on gaps, and this row has %s with %d gap(s)", strings.Join(wordsOf(occ), ", "), len(b.GetGapIds()))
	}
	return nil
}

// wordsOf spells occasions in the enum's words.
func wordsOf(occ []recordpb.Occasion) []string {
	out := make([]string, 0, len(occ))
	for _, o := range occ {
		out = append(out, recordpb.Word(o))
	}
	return out
}

// docketedThisSitting reports whether the docket motion filed at `filed` is the chair's own,
// filed in its current sitting — after the event that opened its latest one — which is how the dispatch
// verb dockets a gap at impasse. A plan asked for again in that sitting names the gap in Docket
// still, so the plan the chair relays last says what the sitting docketed.
func docketedThisSitting(evs []*Event, ids []int64, chairRegisters []int64, filed int64) bool {
	if len(chairRegisters) == 0 || filed <= chairRegisters[len(chairRegisters)-1] {
		return false
	}
	k, found := slices.BinarySearch(ids, filed)
	return found && evs[k].GetSeatId() == chairSeat
}

// benchSatOnOccasionSince reports whether the bench opened a sitting of the occasion after at — off
// the same register ledger, and the same place sequence, as benchSatFor.
func benchSatOnOccasionSince(on map[recordpb.Occasion][]int64, occ recordpb.Occasion, at int64) bool {
	_, ok := firstAfter(on[occ], at)
	return ok
}

// benchSatFor reports whether the bench has had its sitting for EVERY docket motion standing
// unruled on gapID — one bench sitting per docketing. A sitting counts for a docketing when the
// bench OPENED it AFTER BOTH a dispatch engaging it on the gap and the filing: for a dispatch
// naming gapID, the first docket sitting opened after it opens after unruledFiled, the events.id of
// the newest unruled docket motion on the gap (the gap view's `unruled_docket_filed`, off
// openGaps). The newest unruled filing is the key because a sitting that follows it follows every
// older one too. A bench that sat for a docket and ruled nothing is not re-readied for it; a filing
// the bench has not sat since is a new docketing.
//
// BOTH ORDERINGS, NOT THE DISPATCH'S ALONE. Keyed on the latest dispatch naming the gap, the
// bench that sat for G1's first docket and ruled M1 read as having sat for M2 too: blue filed M2
// after that sitting, a judge register stood after the last dispatch naming G1, and nobody was
// engaged while M2 stood unruled for the run (#1201). Keyed on the dispatch's place against the
// filing, a motion blue filed between the chair's dispatch and the bench's register — the bench
// sat with it on the record and ruled nothing — re-readied the bench for a docketing it had sat for.
// Keyed on the NEWEST filing rather than the newest unruled one, a motion blue filed into an open
// sitting and the bench ruled there re-readied the bench for the older motion it had left alone.
//
// THE SITTING THAT COUNTS IS A DOCKET SITTING. A row convening the bench for a petition and a
// docket is sat twice, petition first, and the petition sitting answers no docketing.
func benchSatFor(dispatches []dispatchRow, on map[recordpb.Occasion][]int64, gapID string, unruledFiled int64) bool {
	for _, d := range dispatches {
		if d.seat != benchSeat || !slices.Contains(d.gaps, gapID) {
			continue
		}
		if r, ok := firstAfter(on[recordpb.Occasion_OCCASION_DOCKET], d.at); ok && r > unruledFiled {
			return true
		}
	}
	return false
}

// openGaps reads the open gaps with what the plan needs of each: who minted it, its current
// severity, whether it is stranded, its docket and its remand's direction — all off the gap view, the
// same fold every reader uses (how many remands count is the exchange fold's, exchangesOf). "A docket stands unruled" is the view's `unruled_docket_filed` (recordsql/views.go),
// the one definition the view's own `remanded` is derived from; the PASS gate (motionsAt) asks
// the same per-motion question of the events, and
// TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled holds the two level:
// the bench is readied for a gap exactly when the gap's motion is among the plan's unruled ones.
func openGaps(db recordsql.Querier) ([]openGap, error) {
	rows, err := db.Query(`SELECT g."gap_id", COALESCE(g."minted_by", ''), COALESCE(g."current_severity", ''),
	    g."material", COALESCE(g."class_material", ''),
	    CASE WHEN g."stranded" THEN COALESCE(g."superseded_by", '') ELSE '' END,
	    COALESCE(g."unruled_docket_filed", 0),
	    COALESCE(g."unruled_docket_motion", ''),
	    COALESCE(g."docket_direction", '')
	  FROM "gap" g WHERE g."open" ORDER BY g."minted_event"`)
	if err != nil {
		return nil, fmt.Errorf("asking the record for its open gaps: %w", err)
	}
	defer rows.Close()
	var out []openGap
	for rows.Next() {
		var g openGap
		if err := rows.Scan(&g.id, &g.mintedBy, &g.severity, &g.material, &g.classMaterial, &g.supersededBy, &g.unruledFiled, &g.unruledMotion,
			&g.direction); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// notMaterialBecause is the one wording of WHY an open gap is not material: its class says never,
// or its class goes by grade and its current grade is below the floor. The dispatch's reasons and
// the chair's work list both say it.
func notMaterialBecause(classMaterial, severity string) string {
	if classMaterial == recordpb.Word(recordpb.ClassMaterial_CLASS_MATERIAL_NEVER) {
		return "its class is never material"
	}
	if severity == "" {
		return "ungraded"
	}
	return "graded " + severity
}
