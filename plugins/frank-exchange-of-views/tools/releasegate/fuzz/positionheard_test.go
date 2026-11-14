package fuzz

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHEN THE CHAIR HOLDS BACK A PASS ITS PLAN PERMITS, THE BENCH HEARS ITS POSITION (gblock,
// 2026-10-30). The drive stands in for each seat of that exchange: the chair's withheld PASS files the
// position (chairPosition), the plan it asks for next convenes the bench, and the bench's position
// sitting — or the terminal sitting, at the epoch limit — rules it upheld or overruled on a coin. An
// upheld ruling names a cast lens or open gaps, with a direction; an overruled one names nothing, and
// the chair's next sitting records its verdict, after the tool refuses a second position there.
// positionOracle holds every position the chair withheld a PASS on to that route.

// positionIDOf pulls the tool-assigned position id out of a position's JSON envelope.
func positionIDOf(out string) string {
	var env struct {
		Result struct {
			PositionID string `json:"position_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		return ""
	}
	return env.Result.PositionID
}

// unruledPositions is every position motion the record holds unruled: the positions the bench hears.
func (r *runner) unruledPositions() []string {
	fam, err := record.FamilyOf(r.run())
	if err != nil {
		r.noteActErr("the record could not be read for the positions the bench hears: " + err.Error())
		return nil
	}
	var ids []string
	for _, m := range record.MotionsOf(fam.Events, fam.At) {
		if m != nil && m.Subject == "position" && !m.Ruled() {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

// THE THREE RULINGS ARE FORCED, ONE RUN EACH, NOT DRAWN. A drawn run rarely reaches a plan that
// permits the PASS (38 of 40 end CEILING), so a coin over it drove one withheld PASS in a sweep and
// neither upheld form: `motion position rule --ids` went unpassed. Each position seed's run is a
// forced-VERIFIED run, which reaches a permitted PASS by construction; its chair withholds the first
// one its ruling's route fits the epoch limit (withholds), and its bench rules the position the way
// the seed names.

// positionFormOf is the ruling a seed's run forces its bench to: "lens" (upheld naming a lens),
// "gaps" (upheld naming open gaps) or "overruled"; "" for a seed that forces none.
func positionFormOf(seed int64) string {
	switch seed {
	case positionLensSeed:
		return "lens"
	case positionGapsSeed:
		return "gaps"
	case positionOverruledSeed:
		return "overruled"
	}
	return ""
}

// withholds is whether a chair whose plan permits the PASS, at the chair sitting that opened epoch
// of the run's limit maxEpochs, withholds it. A drawn run withholds on a coin and a forced run with no
// position to force never does. A run forced to a position withholds once, and only where its
// ruling's route still fits the limit (lastWithholdEpoch): a PASS withheld later sends the run to
// CEILING, which no ruling the bench makes can lift, and the forced-VERIFIED run fails on a draw. Its
// first permitted PASS in the window is the one withheld — the gaps run's once its lens holds a gap
// open for the bench to name, or at the window's last epoch whatever it holds, so the seed is
// exercised whenever the window holds a permitted PASS. A permitted PASS past the window is passed,
// and its epoch kept for the forced-position oracle to name.
func (r *runner) withholds(epoch, maxEpochs int) bool {
	if r.forcePosition == "" {
		return !r.forcedVerdict() && r.coin(5)
	}
	last := lastWithholdEpoch(r.forcePosition, maxEpochs)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.positionLimit = maxEpochs
	switch {
	case r.positionForced:
		return false
	case epoch > last:
		if r.positionPassedLate == 0 {
			r.positionPassedLate = epoch
		}
		return false
	case r.forcePosition == "gaps" && !r.heldTrifle && epoch < last:
		return false
	}
	r.positionForced = true
	r.positionWithheldAt = epoch
	return true
}

// planEpochs is the epoch a plan was asked in and the run's epoch limit, read off the plan's own
// epoch and max_epochs.
func planEpochs(plan map[string]any) (epoch, maxEpochs int) {
	e, _ := plan["epoch"].(float64)
	m, _ := plan["max_epochs"].(float64)
	return int(e), int(m)
}

// THE ROUTE A RULING ON A WITHHELD PASS SENDS THE RUN DOWN, as the record routes it
// (record/positionheard.go): one seat dispatched per epoch after the withholding chair sitting, in
// order, and then the chair's sitting that records the PASS.
//   - lens: the named lens owes an audit of its area (an upheld_position blocker), which readies it
//     at the next chair sitting until it is served the report in a sitting opened after the ruling.
//   - gaps: each named gap goes back for one more exchange — a remand whose plan readies blue until
//     it has acted, and then the gap's minting lens alone, whose re-audit closes it.
//   - overruled: the chair's next sitting decides on the PASS, and dispatches nobody first.
//
// The bench's position sitting falls inside the withholding epoch: the engine convenes it before the
// next chair sitting, so it takes no epoch of its own.
func positionRoute(form string) []string {
	switch form {
	case "lens":
		return []string{routeLens}
	case "gaps":
		return []string{routeBlue, routeLens}
	}
	return nil
}

// The roles a position's route dispatches: blue answering a remand, and a lens auditing.
const (
	routeBlue = "blue"
	routeLens = "lens"
)

// positionRouteSittings is the chair sittings a ruling's route needs after the one that withheld the
// PASS: one per seat it dispatches, each in an epoch of its own, and the one that records the PASS.
func positionRouteSittings(form string) int { return len(positionRoute(form)) + 1 }

// lastWithholdEpoch is the last epoch under the limit maxEpochs at which a chair withholding a PASS
// leaves the form's route room to reach the PASS: every seat it dispatches sits in an epoch that
// dispatches it — not the final sitting, which dispatches nobody, and for blue not the epoch it is
// held out of (record.Plan.BlueHeldOut) — and the PASS is recorded by the final sitting at the
// latest. 0 when no epoch leaves the room; maxEpochs 0 holds the run to no limit.
func lastWithholdEpoch(form string, maxEpochs int) int {
	if maxEpochs <= 0 {
		return math.MaxInt
	}
	route := positionRoute(form)
	fits := func(withheld int) bool {
		for i, seat := range route {
			e := withheld + 1 + i
			if finalEpoch(e, maxEpochs) || (seat == routeBlue && record.Plan{MaxEpochs: maxEpochs, Epoch: e}.BlueHeldOut()) {
				return false
			}
		}
		return withheld+len(route)+1 <= maxEpochs
	}
	for e := maxEpochs; e >= 1; e-- {
		if fits(e) {
			return e
		}
	}
	return 0
}

// finalEpoch is whether epoch is the run's final sitting under the limit maxEpochs — the chair
// sitting that opens epoch maxEpochs, after which no seat is dispatched (record.isFinalEpoch, which
// the record does not export; TestTheLensRouteFitsTheLimitAtItsLastWithholdEpoch drives the record
// to the boundary it draws).
func finalEpoch(epoch, maxEpochs int) bool { return maxEpochs > 0 && epoch >= maxEpochs }

// mintHeldTrifle is the gaps run's lens minting, once, a gap of a class that is never material, at
// a sitting after the first where it holds no gap of its own open. A gap that is not material readies no seat and holds no PASS, so it stands open, with
// no exchange on it, at the chair's next permitted PASS — the open gap the bench names. The bench's
// remand of it readies its lens and blue, who answer and close it, and the run then passes.
func (r *runner) mintHeldTrifle(seatID string) {
	if r.forcePosition != "gaps" || r.firstEpoch() {
		return
	}
	open := r.openGaps()
	r.mu.Lock()
	if r.heldTrifle || slices.ContainsFunc(open, func(id string) bool { return r.minter[id] == seatID }) {
		r.mu.Unlock()
		return
	}
	r.heldTrifle = true
	r.mu.Unlock()
	class := neverMaterialClass()
	if class == "" {
		r.noteActErr("the gap class vocabulary holds no class that is never material, so the gaps run has no gap to hold open for the bench to name")
		return
	}
	// THE MINT IS RETRIED IN THE SITTING. The lenses of an epoch sit concurrently, and the first to
	// reach here holds the others off, so one refused draw — a subject a gap of the class already
	// stands on, say — would cost the whole epoch; where the PASS is permitted at the next chair
	// sitting no lens sits again, and the chair passes it with nothing open for the bench to name.
	// Each try draws the mint afresh, and each refusal is kept for the oracle to name.
	for range trifleMintTries {
		if r.mintWith(seatID, class) != "" {
			return
		}
		r.mu.Lock()
		r.trifleRefusals = append(r.trifleRefusals, fmt.Sprintf("%s: %s", seatID, r.mintRefusal[seatID]))
		r.mu.Unlock()
	}
	r.mu.Lock()
	r.heldTrifle = false
	r.mu.Unlock()
}

// trifleMintTries is the draws of the held trifle's mint one lens sitting makes before it leaves the
// mint to the next sitting.
const trifleMintTries = 3

// closeNamedOnTurn is the minting lens's turn on a gap the bench named upholding a position: once
// blue has answered it, the lens audits the answer and closes it as any satisfied gap is closed. A
// gap that is not material is routed to nobody once its remand's exchange has begun, so the lens is
// readied by blue's answer as news, not engaged on the gap, and the engaged loop never reaches it.
func (r *runner) closeNamedOnTurn(seatID string, engaged []string) {
	r.mu.Lock()
	named := slices.Clone(r.positionNamedGaps)
	r.mu.Unlock()
	open := r.openGaps()
	for _, g := range named {
		if !slices.Contains(open, g) || slices.Contains(engaged, g) || r.minterOf(g) != seatID || !r.closableNow(g) {
			continue
		}
		r.closeGap(seatID, g, false)
	}
}

// neverMaterialClass is the first gap class whose material_default is `never`, read off the enum.
func neverMaterialClass() string {
	vals := recordpb.GapClass(0).Descriptor().Values()
	for i := 0; i < vals.Len(); i++ {
		c := recordpb.GapClass(vals.Get(i).Number())
		if recordpb.MaterialDefault(c) == recordpb.ClassMaterial_CLASS_MATERIAL_NEVER {
			return recordpb.Word(c)
		}
	}
	return ""
}

// positionClasses is every gap class word a position may name, read off the enum.
func positionClasses() []string {
	vals := recordpb.GapClass(0).Descriptor().Values()
	var out []string
	for i := 1; i < vals.Len(); i++ {
		out = append(out, recordpb.Word(recordpb.GapClass(vals.Get(i).Number())))
	}
	return out
}

// positionClassQuota is how many positions a drawn run's chair files whatever its coin says, each
// naming a class. THE SWEEP'S COVERAGE OF THE CLASS ENUM RESTS ON IT, not on the coin: a seed's
// rotation starts quota*seed classes in (positionClassOf), so consecutive seeds' first quota
// positions tile the enum, and the sweep tiles it twice over (TestTheSweepNamesEveryPositionClass)
// — a forced seed, which names none, or a run cut short leaves no class unnamed.
const positionClassQuota = 3

// positionClassOf is the class a drawn run's k-th classed position names: the rotation through
// positionClasses that starts positionClassQuota*seed classes in.
func positionClassOf(seed int64, k int) string {
	cs := positionClasses()
	return cs[(int(seed)*positionClassQuota+k)%len(cs)]
}

// alwaysMaterial is whether class's material default is `always`; false for "" or an unknown word.
func alwaysMaterial(class string) bool {
	c, ok := record.GapClassOf(class)
	return ok && recordpb.MaterialDefault(c) == recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS
}

// rulePositions rules every position the bench hears, at the sitting of the given occasion. A run
// forced to a position rules it the way it is forced; any other forced run's bench overrules, and a
// drawn run's rules on a coin. Upheld names open gaps the tool accepts, or a cast lens.
func (r *runner) rulePositions(seatID, occasion string) {
	if r.positionRuledAt == nil {
		r.positionRuledAt = map[string]string{}
	}
	// The bench sits between chair sittings: its ruling falls in the epoch the last one opened.
	epoch, _ := planEpochs(r.planThisSitting)
	for _, id := range r.unruledPositions() {
		const opinion = "fuzz: the bench's reasoning on the PASS withheld"
		open := r.openGaps()
		var form string
		switch {
		case r.forcePosition != "":
			form = r.forcePosition
		// A FORCED RUN'S BENCH OVERRULES, and drives no refusal: its chair's next sitting decides on
		// the PASS, so the forced run keeps the exit it is forced to (UNVERIFIED: its chair leaves its
		// own motions unruled, so that verdict is a FAIL or none).
		case r.forcedVerdict(), r.coin(50):
			form = "overruled"
		case len(open) > 0 && r.coin(50):
			form = "gaps"
		default:
			form = "lens"
		}
		// A POSITION WHOSE CLASS IS ALWAYS MATERIAL IS NOT OVERRULED (gblock, 2026-11-05): the overrule
		// is refused, and the bench upholds it naming a lens of the cast, retired or not.
		r.mu.Lock()
		class := r.positionClass[id]
		r.mu.Unlock()
		if form == "overruled" && alwaysMaterial(class) {
			r.positionRulingRefused(seatID, id, "overruled over a position of always-material class "+class, "--as", "overruled", "--reason", opinion)
			form = "lens"
		}
		if form == "overruled" {
			// AN OVERRULED RULING NAMES NO WORK, and one that names a lens is refused.
			if !r.forcedVerdict() {
				r.maybe(20, func() {
					r.positionRulingRefused(seatID, id, "overruled naming a lens", "--as", "overruled", "--lens", r.castLens(), "--reason", opinion)
				})
			}
			if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "overruled", "--reason", opinion); err != nil {
				r.noteActErr(fmt.Sprintf("the bench's overruling of position %s was refused: %v", id, err))
				continue
			}
			r.mu.Lock()
			r.positionRuledAt[id] = occasion
			r.positionRuledAtEpoch = epoch
			r.positionsOverruled++
			r.overruledPending = id
			r.mu.Unlock()
			continue
		}
		// AN UPHELD RULING NAMES ITS WORK ONE WAY: naming neither is refused.
		r.maybe(20, func() {
			r.positionRulingRefused(seatID, id, "upheld naming no work", "--as", "upheld", "--direction", "fuzz direction", "--reason", opinion)
		})
		const direction = "fuzz: audit the claim the chair would not pass"
		if form == "gaps" {
			// THE FIRST OPEN GAP THE TOOL ACCEPTS: one remanded at impasse already, or standing
			// docketed unruled, is refused, and the next is named instead.
			named := false
			for _, g := range open {
				if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "upheld",
					"--ids", g, "--direction", direction, "--reason", opinion); err == nil {
					named = true
					r.mu.Lock()
					r.positionNamedGaps = append(r.positionNamedGaps, g)
					if r.positionNamedAt == nil {
						r.positionNamedAt = map[string]string{}
					}
					r.positionNamedAt[g] = fmt.Sprintf("the %s sitting, in chair sitting %d", occasion, r.chairRegisters)
					r.mu.Unlock()
					r.remandRouted(g)
					break
				}
			}
			switch {
			case !named && r.forcePosition == "gaps":
				r.noteActErr(fmt.Sprintf("the gaps run's bench could name none of the open gaps %v in its upholding of position %s", open, id))
				continue
			case !named:
				form = "lens"
			}
		}
		if form == "lens" {
			if _, err := r.exec("motion", "position", "rule", "--seat-id", seatID, "--id", id, "--as", "upheld",
				"--lens", r.castLens(), "--direction", direction, "--reason", opinion); err != nil {
				r.noteActErr(fmt.Sprintf("the bench's upholding of position %s, naming a cast lens, was refused: %v", id, err))
				continue
			}
		}
		r.mu.Lock()
		r.positionRuledAt[id] = occasion
		r.positionRuledAtEpoch = epoch
		if form == "gaps" {
			r.positionsUpheldGaps++
		} else {
			r.positionsUpheldLens++
		}
		r.mu.Unlock()
	}
}

// castLens is a lens of the run's cast, drawn.
func (r *runner) castLens() string {
	cast, err := record.CastOf(r.run())
	if err != nil {
		return fuzzLensSeats[0]
	}
	var lenses []string
	for _, s := range cast {
		if strings.HasPrefix(s, "red-lens-") {
			lenses = append(lenses, s)
		}
	}
	if len(lenses) == 0 {
		return fuzzLensSeats[0]
	}
	return pick(r.rng, lenses)
}

// positionRulingRefused drives a ruling on position id that names the wrong work, and holds the tool
// to refusing it.
func (r *runner) positionRulingRefused(seatID, id, what string, args ...string) {
	out, err := r.exec(append([]string{"motion", "position", "rule", "--seat-id", seatID, "--id", id}, args...)...)
	if err == nil {
		r.noteActErr(fmt.Sprintf("the bench's ruling on position %s, %s, was admitted: %s", id, what, firstLine(out)))
		return
	}
	r.mu.Lock()
	r.positionRulingsRefused++
	r.mu.Unlock()
}

// repeatedPositionRefused files the chair's position at its first sitting after the bench overruled
// its last one, and holds the tool to refusing it as a conflict naming the overruled position.
func (r *runner) repeatedPositionRefused(seatID, overruled string) {
	out, err := r.exec("--json", "position", "--seat-id", seatID, "--reason", "fuzz: the same objection again")
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(fmt.Sprintf("the chair's position at its first sitting after the bench overruled %s was admitted: %s", overruled, firstLine(out)))
	case !ok || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, overruled):
		r.noteActErr(fmt.Sprintf("the chair's repeated position was refused without naming the overruled position %s as a conflict: %s", overruled, firstLine(out)))
	default:
		r.mu.Lock()
		r.repeatedPositionsRefused++
		r.mu.Unlock()
	}
}

// takeOverruled is the position the bench overruled since the chair last sat, cleared: the chair's
// next sitting owes the verdict.
func (r *runner) takeOverruled() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.overruledPending
	r.overruledPending = ""
	return id
}

// positionOracle holds every position the chair filed on a PASS it withheld to its route: the bench
// rules it, at the position sitting the engine convened or at the terminal sitting. A halted run is
// exempt, as the petition oracle exempts it. "" when it holds.
func (r *runner) positionOracle(halted bool) string {
	if halted {
		return ""
	}
	for _, id := range r.positionsHeard {
		switch at, ruled := r.positionRuledAt[id]; {
		case !ruled:
			return fmt.Sprintf("position oracle: the chair withheld a PASS its plan permitted and filed position %s, and the bench never ruled it", id)
		case at != "position" && at != "terminal":
			return fmt.Sprintf("position oracle: position %s was ruled at the %s sitting — the engine hears it as soon as nothing is running, or the terminal sitting hears it at the exit", id, at)
		}
	}
	return ""
}

// forcedPositionOracle holds a run forced to a position to its ruling's route: the chair withheld a
// PASS and the bench ruled it the way the seed names, and a gap the gaps run's bench named is closed
// by the run's end — its remand's exchange is what the ruling sent it to, and VERIFIED holds over a
// gap that is not material, so the verdict alone does not say it was had.
func (r *runner) forcedPositionOracle() string {
	if r.forcePosition == "" {
		return ""
	}
	r.mu.Lock()
	got := map[string]int{"lens": r.positionsUpheldLens, "gaps": r.positionsUpheldGaps, "overruled": r.positionsOverruled}
	named := slices.Clone(r.positionNamedGaps)
	r.mu.Unlock()
	if got[r.forcePosition] == 0 {
		return fmt.Sprintf("forced-position oracle: the run is forced to a %s ruling on a withheld PASS, and its bench made none (rulings %v, %d PASS withheld)%s%s%s", r.forcePosition, got, r.positionsWithheld, r.positionRouteClause(), r.trifleClause(), r.planTrail())
	}
	open := r.openGaps()
	for _, g := range named {
		if slices.Contains(open, g) {
			return fmt.Sprintf("forced-position oracle: the bench upheld the withheld PASS's position naming gap %s, and the gap is still open at the run's end — the remand's exchange never closed it%s", g, r.namedGapEvidence(g))
		}
	}
	return ""
}

// positionRouteClause is what a forced position run's failure says about its route against the epoch
// limit: the window its ruling's route fits, the epoch the PASS was withheld at and the epoch of the
// ruling, or the epoch whose permitted PASS came past the window and was passed. "" in a run forced to
// no position.
func (r *runner) positionRouteClause() string {
	if r.forcePosition == "" {
		return ""
	}
	r.mu.Lock()
	limit, withheld, ruled, late := r.positionLimit, r.positionWithheldAt, r.positionRuledAtEpoch, r.positionPassedLate
	r.mu.Unlock()
	if limit == 0 && withheld == 0 && late == 0 {
		return "; no chair sitting's plan permitted the PASS"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "; the %s route takes %d chair sitting(s) after the withholding one, so under the epoch limit %d the last epoch that can withhold is %d", r.forcePosition, positionRouteSittings(r.forcePosition), limit, lastWithholdEpoch(r.forcePosition, limit))
	switch {
	case withheld > 0 && ruled > 0:
		fmt.Fprintf(&b, "; the PASS was withheld at epoch %d and the bench ruled in epoch %d", withheld, ruled)
	case withheld > 0:
		fmt.Fprintf(&b, "; the PASS was withheld at epoch %d and the bench never ruled", withheld)
	}
	if late > 0 {
		fmt.Fprintf(&b, "; the first PASS permitted past that window, at epoch %d, was passed", late)
	}
	return b.String()
}

// trifleClause is what a gaps run's failure says about the gap its lens holds open for the bench to
// name: that none landed, with each refusal of its mint. "" in any other run, or once one landed.
func (r *runner) trifleClause() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.forcePosition != "gaps" || r.heldTrifle {
		return ""
	}
	if len(r.trifleRefusals) == 0 {
		return "; no lens sitting after the first minted the held trifle"
	}
	return fmt.Sprintf("; the held trifle never landed, its mint refused %d time(s): %s", len(r.trifleRefusals), strings.Join(r.trifleRefusals, " | "))
}

// positionEpochLimit is the epoch limit a position seed's run is set up with: past setup's default,
// so the window its ruling's route fits (lastWithholdEpoch) holds the first permitted PASS of every
// interleaving. A forced run mints its stock at its first sitting (findsFresh), but a gap blue
// counters goes to the bench at the fuzz's kMax of two and comes back on remand, so the march to a
// permitted PASS is not one exchange long: MEASURED on 120 runs of positionGapsSeed under the limit
// of twelve, the first permitted PASS came at epoch 5 in most, and at epoch 10 in three — past the
// gaps window of 9, where the chair passes it and the bench rules nothing. Sixteen leaves the gaps
// route a window to 13.
const positionEpochLimit = 16

// planTrail is what each plan the record stored held, one line per chair sitting: its epoch, whether
// it permitted the PASS, the parties it readied with their gaps, and the blockers holding the PASS
// with their owners. A forced run reaches its permitted PASS one march after its first sitting
// (findsFresh), and a run whose first permitted PASS comes later re-opened the march somewhere; the
// run directory does not outlive the CI job, so the failure carries the sittings that say where.
func (r *runner) planTrail() string {
	board, err := record.FamilyOf(r.runHandle)
	if err != nil {
		return "\n  the plans could not be read: " + err.Error()
	}
	stored, err := record.StoredPlans(board.Events)
	if err != nil {
		return "\n  the plans the record stored do not read: " + err.Error()
	}
	var b strings.Builder
	b.WriteString("\n  the plans, by chair sitting:")
	for _, sp := range stored {
		p := sp.Plan
		parties := make([]string, 0, len(p.Parties))
		for _, pa := range p.Parties {
			parties = append(parties, fmt.Sprintf("%s%v", pa.SeatID, pa.GapIDs))
		}
		held := make([]string, 0, len(p.Blockers))
		for _, bl := range p.Blockers {
			held = append(held, fmt.Sprintf("%s %s (%s)", bl.Kind, bl.Subject, bl.Owner))
		}
		fmt.Fprintf(&b, "\n    epoch %d: pass_permitted %t; parties %v; held by %v", p.Epoch, p.PassPermitted, parties, held)
	}
	return b.String()
}

// namedGapEvidence is what the forced-position oracle's failure carries about a named gap left open,
// because the run directory does not outlive the CI job: when the ruling was made against the run's
// limit, the gap's minter and scenario and whether the drive had blue's answer to close on, the last
// plan's word on the gap, and the gap's change log. It tells a remand cut off by the epoch limit from
// one whose exchange stalled.
func (r *runner) namedGapEvidence(g string) string {
	r.mu.Lock()
	at := r.positionNamedAt[g]
	sittings := r.chairRegisters
	r.mu.Unlock()
	limit := record.DefaultParams.MaxEpochs
	if r.maxEpochs > 0 {
		limit = r.maxEpochs
	}
	minter := r.minterOf(g)
	var b strings.Builder
	fmt.Fprintf(&b, "\n  ruled at %s; the chair sat %d times, the run's epoch limit is %d", at, sittings, limit)
	fmt.Fprintf(&b, "\n  minted by %s, scenario %q; blue's answer presented: %t; closable by the drive: %t", minter, r.scenarioOf(g), r.presented.has(g), r.closableNow(g))
	if r.planThisSitting != nil {
		var parties []string
		ps, _ := r.planThisSitting["parties"].([]any)
		for _, p := range ps {
			if m, _ := p.(map[string]any); m != nil {
				parties = append(parties, fmt.Sprintf("%v%v", m["seat_id"], m["gap_ids"]))
			}
		}
		fmt.Fprintf(&b, "\n  the last plan: epoch %v, parties %v", r.planThisSitting["epoch"], parties)
		whys, _ := r.planThisSitting["why"].([]any)
		for _, w := range whys {
			if s := fmt.Sprint(w); strings.Contains(s, g) {
				fmt.Fprintf(&b, "\n    why: %s", s)
			}
		}
	}
	if out, err := r.exec("show", "changes", "--seat-id", minter, "--id", g); err == nil {
		fmt.Fprintf(&b, "\n  its change log:\n%s", truncate(out))
	} else {
		fmt.Fprintf(&b, "\n  its change log could not be read: %v", err)
	}
	return b.String()
}

// THE DRIVE'S BENCH RULES THE POSITION IT HEARS, and the oracle holds it to the route. Pinned outside
// the sweep: a PASS withheld at the final sitting is ruled by the drive's terminal bench, the oracle
// passes it, and a position it never ruled fails the oracle.
func TestTheDrivesBenchRulesThePositionItHears(t *testing.T) {
	r, _ := converseRun(t, redClosesAndWithholds)
	ids := r.unruledPositions()
	if len(ids) != 1 {
		t.Fatalf("the record holds %v unruled positions, want the one the chair withheld the PASS on", ids)
	}
	r.positionsHeard = ids
	if m := r.positionOracle(false); m == "" || !strings.Contains(m, ids[0]) {
		t.Errorf("the oracle held over a position the bench never ruled: %q", m)
	}
	r.dispatchSits("bench", "judge", "terminal")
	r.rulePositions("judge", "terminal")
	r.dispatchReturns("judge", false)
	if r.actErr != "" {
		t.Fatalf("the drive's ruling: %s", r.actErr)
	}
	if left := r.unruledPositions(); len(left) != 0 {
		t.Fatalf("the drive's bench left %v unruled", left)
	}
	if r.positionRuledAt[ids[0]] != "terminal" || r.positionsUpheldLens+r.positionsUpheldGaps+r.positionsOverruled != 1 {
		t.Errorf("ruled at %q, %d upheld on a lens, %d on gaps and %d overruled; want one ruling at the terminal sitting", r.positionRuledAt[ids[0]], r.positionsUpheldLens, r.positionsUpheldGaps, r.positionsOverruled)
	}
	if m := r.positionOracle(false); m != "" {
		t.Errorf("the oracle fired over a position ruled at the terminal sitting: %s", m)
	}
}

// AN OVERRULED POSITION SENDS THE PASS BACK TO THE CHAIR, through the verbs the drive drives: the
// bench's position sitting overrules it, the chair's next sitting is refused a second position naming
// the overruled one, and records the PASS its plan permits.
func TestTheDrivesChairDecidesThePassAfterAnOverruling(t *testing.T) {
	r, last := converseRun(t, redClosesAndWithholdsEarly)
	if last.Ceiling || len(last.Parties) != 0 {
		t.Fatalf("the plan after the position = %+v, want nobody on it: a hearing is on no plan", last)
	}
	if read, err := record.ReadRecord(r.run()); err != nil || !slices.Equal(read.BenchOwed, []record.BenchHearing{"position"}) {
		t.Fatalf("the record's read after the position = %+v (%v), want the position owed alone", read, err)
	}
	id := r.unruledPositions()[0]
	r.dispatchSits("bench", "judge", "position")
	if _, err := r.exec("motion", "position", "rule", "--seat-id", "judge", "--id", id, "--as", "overruled", "--reason", "fuzz: the objection does not hold"); err != nil {
		t.Fatal(err)
	}
	r.dispatchReturns("judge", false)
	r.dispatchSits("chair", "red-chair")
	out, err := r.exec("--json", "dispatch", "next", "--seat-id", "red-chair")
	if err != nil || !strings.Contains(out, `"pass_permitted":true`) {
		t.Fatalf("the chair's next plan: %v\n%s", err, out)
	}
	r.repeatedPositionRefused("red-chair", id)
	if r.actErr != "" || r.repeatedPositionsRefused != 1 {
		t.Fatalf("the repeated position: %q, %d refused", r.actErr, r.repeatedPositionsRefused)
	}
	if _, err := r.exec("--json", "verdict", "--seat-id", "red-chair", "--as", "PASS"); err != nil {
		t.Fatalf("the PASS the overruling sent back: %v", err)
	}
	r.dispatchReturns("red-chair", false)
	if got, _, ok, err := record.DeriveVerdict(r.run()); err != nil || !ok || got != "VERIFIED" {
		t.Fatalf("derived (%q, %v, %v), want VERIFIED", got, ok, err)
	}
}

// A RUN FORCED TO A POSITION RULES IT THAT WAY, whatever its coins: the sweep's gate on each ruling
// holds a forced seed to it (positionFormOf), so the drive must rule the form the seed names.
func TestTheDrivesBenchRulesTheForcedPosition(t *testing.T) {
	for _, form := range []string{"lens", "gaps", "overruled"} {
		t.Run(form, func(t *testing.T) {
			r, _ := converseRun(t, redClosesAndWithholds)
			r.forcePosition = form
			// THE GAPS RUN'S LENS HOLDS A GAP OPEN for the bench to name: one of a class never
			// material, minted at a sitting after the first where it holds none of its own.
			if form == "gaps" {
				// A forced-VERIFIED run's mint is one blue repairs and its lens closes.
				r.forceVerified = true
				r.dispatchSits("lens", converseLens)
				r.mintHeldTrifle(converseLens)
				r.dispatchReturns(converseLens, false)
				if !r.heldTrifle || len(r.openGaps()) != 1 {
					t.Fatalf("the lens held %v open (minted: %t; %s), want the one gap it mints for the bench to name", r.openGaps(), r.heldTrifle, r.actErr)
				}
				bs, err := record.PassBlockers(r.run())
				if err != nil {
					t.Fatal(err)
				}
				for _, b := range bs {
					if b.Subject == r.openGaps()[0] {
						t.Fatalf("the gap held open holds the PASS as %s; it must be one no plan waits on", b.Kind)
					}
				}
			}
			r.positionsHeard = r.unruledPositions()
			r.dispatchSits("bench", "judge", "terminal")
			r.rulePositions("judge", "terminal")
			r.dispatchReturns("judge", false)
			if r.actErr != "" {
				t.Fatalf("the drive's ruling: %s", r.actErr)
			}
			got := map[string]int{"lens": r.positionsUpheldLens, "gaps": r.positionsUpheldGaps, "overruled": r.positionsOverruled}
			if got[form] != 1 || got["lens"]+got["gaps"]+got["overruled"] != 1 {
				t.Errorf("rulings %v, want the one %s ruling the run is forced to", got, form)
			}
			// THE ORACLE HOLDS THE ROUTE: the gap the bench named is still open here, its remand's
			// exchange never had, and the oracle names it; a lens or an overruling has no gap to hold.
			switch m := r.forcedPositionOracle(); {
			case form == "gaps" && (len(r.positionNamedGaps) != 1 || !strings.Contains(m, r.positionNamedGaps[0])):
				t.Errorf("the oracle passed a named gap left open: named %v, %q", r.positionNamedGaps, m)
			// AND ITS FAILURE CARRIES THE EVIDENCE that tells a remand the limit cut off from one that
			// stalled: when the ruling was made, and the gap's change log.
			case form == "gaps" && (!strings.Contains(m, "ruled at the terminal sitting, in chair sitting") || !strings.Contains(m, "its change log:")):
				t.Errorf("the oracle's failure carries no evidence of the named gap's route: %q", m)
			case form != "gaps" && m != "":
				t.Errorf("the oracle fired over a %s ruling: %s", form, m)
			}
			// AND THE LENS'S TURN CLOSES IT: blue has answered the named gap, and its minting lens,
			// readied on that news rather than engaged on the gap, audits the answer and closes it.
			if form == "gaps" {
				g := r.positionNamedGaps[0]
				r.presented.add(g)
				r.dispatchSits("lens", converseLens)
				r.closeNamedOnTurn(converseLens, nil)
				r.dispatchReturns(converseLens, false)
				if m := r.forcedPositionOracle(); m != "" || r.actErr != "" {
					t.Errorf("the lens's turn left the named gap open: %s %s", m, r.actErr)
				}
			}
		})
	}
}

// THE SWEEP NAMES EVERY POSITION CLASS, by construction: each class is among the first
// positionClassQuota positions of at least two drawn seeds of the quorum sweep, so the coverage
// gate (TestFuzzSweepTally's enum values) holds without a coin, and one short run cannot break it.
func TestTheSweepNamesEveryPositionClass(t *testing.T) {
	namedBy := map[string][]int64{}
	for seed := int64(1); seed <= surfaceQuorum; seed++ {
		if seed == unverifiedSeed || seed == verifiedSeed || positionFormOf(seed) != "" {
			continue
		}
		for k := range positionClassQuota {
			c := positionClassOf(seed, k)
			namedBy[c] = append(namedBy[c], seed)
		}
	}
	for _, c := range positionClasses() {
		if len(namedBy[c]) < 2 {
			t.Errorf("position class %s is among the first %d positions of drawn seeds %v of the %d-seed sweep — fewer than two, so the class's coverage rests on one run", c, positionClassQuota, namedBy[c], surfaceQuorum)
		}
	}
}

// A FORCED CHAIR WITHHOLDS ONCE, AND ONLY WHERE ITS ROUTE FITS THE LIMIT: the lens run's at its first
// permitted PASS up to the window's last epoch, the gaps run's once its lens holds a gap open or at that
// last epoch whatever it holds, and none past it — that PASS is passed, and its epoch kept.
func TestTheForcedChairWithholdsWhereItsRouteFitsTheLimit(t *testing.T) {
	limit := record.DefaultParams.MaxEpochs
	for _, form := range []string{"lens", "gaps", "overruled"} {
		last := lastWithholdEpoch(form, limit)
		t.Run(form, func(t *testing.T) {
			r := &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: form, heldTrifle: true}
			if r.withholds(last+1, limit) {
				t.Errorf("withheld at epoch %d, past %d, the last its route fits the limit %d", last+1, last, limit)
			}
			if r.positionPassedLate != last+1 {
				t.Errorf("the PASS passed late is kept at epoch %d, want %d", r.positionPassedLate, last+1)
			}
			r = &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: form, heldTrifle: true}
			if !r.withholds(last, limit) || r.withholds(last, limit) {
				t.Errorf("must withhold at epoch %d, the last its route fits, and only once", last)
			}
			if r.positionWithheldAt != last || r.positionPassedLate != 0 {
				t.Errorf("withheld at %d, passed late at %d; want withheld at %d", r.positionWithheldAt, r.positionPassedLate, last)
			}
			r = &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: form}
			switch got := r.withholds(last-1, limit); {
			case form == "gaps" && got:
				t.Errorf("the gaps run's chair withheld at epoch %d, before the window's last, with no gap held open to name", last-1)
			case form != "gaps" && !got:
				t.Errorf("the %s run's chair passed its first permitted PASS, at epoch %d, inside the window", form, last-1)
			}
			if form == "gaps" {
				if !r.withholds(last, limit) {
					t.Errorf("the gaps run's chair did not withhold at epoch %d, the window's last, with no gap held open: the seed would go unexercised", last)
				}
				r = &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: form, heldTrifle: true}
				if !r.withholds(last-1, limit) {
					t.Errorf("the gaps run's chair did not withhold at epoch %d once the gap stood open", last-1)
				}
			}
		})
	}
	r := &runner{rng: newLockedRand(1), forceVerified: true}
	for range 50 {
		if r.withholds(1, limit) {
			t.Fatal("a forced run with no position to force withheld a permitted PASS")
		}
	}
}

// THE WINDOW IS THE ROUTE'S SITTINGS COUNTED BACK FROM THE LIMIT: the PASS may be recorded at the final
// sitting, so a route of n chair sittings after the withholding one fits from epoch limit-n — the lens
// route's audit, the gaps route's remand answer and re-audit, the overruled route's next chair sitting.
// Where blue sits, the epoch it is held out of (record.Plan.BlueHeldOut) is the one before the final,
// and its route puts its lens there.
func TestThePositionRouteWindowIsItsSittings(t *testing.T) {
	want := map[string]int{"lens": 2, "gaps": 3, "overruled": 1}
	for _, limit := range []int{3, 4, record.DefaultParams.MaxEpochs} {
		for form, sittings := range want {
			if got := positionRouteSittings(form); got != sittings {
				t.Errorf("the %s route takes %d chair sittings, want %d", form, got, sittings)
			}
			last := lastWithholdEpoch(form, limit)
			if wantLast := max(limit-sittings, 0); last != wantLast {
				t.Errorf("limit %d: the %s route's last withholding epoch is %d, want %d (the limit less its %d sittings)", limit, form, last, wantLast, sittings)
			}
			for i, seat := range positionRoute(form) {
				e := last + 1 + i
				if seat == routeBlue && (record.Plan{MaxEpochs: limit, Epoch: e}.BlueHeldOut() || finalEpoch(e, limit)) {
					t.Errorf("limit %d: the %s route's blue sits at epoch %d, which does not dispatch it", limit, form, e)
				}
			}
		}
	}
	if got := lastWithholdEpoch("gaps", 0); got != math.MaxInt {
		t.Errorf("a run held to no limit withholds at any epoch; got a last epoch of %d", got)
	}
}

// THE LENS ROUTE FITS THE LIMIT AT ITS LAST WITHHOLDING EPOCH AND NOT ONE LATER, on the record: a PASS
// withheld at epoch 2 of a limit of 4 is upheld naming the lens, the lens is readied at epoch 3 and
// served the report, and the final sitting records the PASS; withheld at epoch 2 of a limit of 3, the
// final sitting has the lens readied and undispatched, and the run ends CEILING.
func TestTheLensRouteFitsTheLimitAtItsLastWithholdEpoch(t *testing.T) {
	for _, tc := range []struct {
		path  conversePath
		limit int
		want  string
	}{
		{redClosesAndWithholdsWithRoom, 4, "VERIFIED"},
		{redClosesAndWithholdsEarly, 3, "CEILING"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			r, _ := converseRun(t, tc.path)
			const withheldAt = 2
			if fits := withheldAt <= lastWithholdEpoch("lens", tc.limit); fits != (tc.want == "VERIFIED") {
				t.Fatalf("lastWithholdEpoch(lens, %d) = %d; the record's route ends %s from epoch %d", tc.limit, lastWithholdEpoch("lens", tc.limit), tc.want, withheldAt)
			}
			id := r.unruledPositions()[0]
			r.dispatchSits("bench", "judge", "position")
			if _, err := r.exec("motion", "position", "rule", "--seat-id", "judge", "--id", id, "--as", "upheld",
				"--lens", converseLens, "--direction", "fuzz: audit the claim the chair would not pass", "--reason", "fuzz: the objection holds"); err != nil {
				t.Fatal(err)
			}
			r.dispatchReturns("judge", false)
			r.dispatchSits("chair", "red-chair")
			var plan record.Plan
			if out, err := r.exec("--json", "dispatch", "next", "--seat-id", "red-chair"); err != nil || json.Unmarshal([]byte(out), &plan) != nil {
				t.Fatalf("the chair's plan after the ruling: %v\n%s", err, out)
			}
			r.dispatchReturns("red-chair", false)
			if tc.want == "CEILING" {
				if !plan.EpochLimitReached || plan.Epoch != tc.limit {
					t.Fatalf("the plan after the ruling = %+v, want the final sitting at the limit with the lens undispatched", plan)
				}
				if why := epochLimitWhy(plan); !strings.Contains(why, fmt.Sprintf("epoch %d of the limit %d", tc.limit, tc.limit)) || !strings.Contains(why, converseLens) {
					t.Errorf("the exit reason names neither the limit nor the lens it held: %q", why)
				}
			} else {
				if !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == converseLens }) {
					t.Fatalf("the plan after the ruling = %+v, want the named lens readied", plan)
				}
				r.dispatchSits("lens", converseLens)
				if _, err := r.exec("--json", "--seat-id", converseLens, "show", "report"); err != nil {
					t.Fatal(err)
				}
				r.dispatchReturns(converseLens, false)
				r.dispatchSits("chair", "red-chair")
				out, err := r.exec("--json", "dispatch", "next", "--seat-id", "red-chair")
				if err != nil || !strings.Contains(out, `"pass_permitted":true`) {
					t.Fatalf("the final sitting's plan: %v\n%s", err, out)
				}
				if _, err := r.exec("--json", "verdict", "--seat-id", "red-chair", "--as", "PASS"); err != nil {
					t.Fatalf("the PASS the lens's audit freed: %v", err)
				}
				r.dispatchReturns("red-chair", false)
			}
			requireEveryDispatchReturned(t, r)
			if got, why, ok, err := record.DeriveVerdict(r.run()); err != nil || !ok || got != tc.want {
				t.Fatalf("derived (%q, %q, %v, %v), want %s", got, why, ok, err, tc.want)
			}
		})
	}
}

// A FORCED RUN THAT ENDS AT ITS LIMIT SAYS SO: the exit reason is the limit with what it held, and a
// position run's names the route's window and the epochs it withheld and ruled at, or the PASS it
// passed past the window.
func TestTheForcedPositionFailureNamesItsRoute(t *testing.T) {
	r := &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: "lens"}
	r.withholds(10, 12)
	r.positionRuledAtEpoch = 10
	if got := r.positionRouteClause(); !strings.Contains(got, "the last epoch that can withhold is 10") || !strings.Contains(got, "withheld at epoch 10 and the bench ruled in epoch 10") {
		t.Errorf("the route clause: %q", got)
	}
	r = &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: "gaps"}
	r.withholds(11, 12)
	if got := r.positionRouteClause(); !strings.Contains(got, "the last epoch that can withhold is 9") || !strings.Contains(got, "at epoch 11, was passed") {
		t.Errorf("the route clause of a PASS passed past the window: %q", got)
	}
	if got := (&runner{}).positionRouteClause(); got != "" {
		t.Errorf("a run forced to no position has a route clause: %q", got)
	}
	why := epochLimitWhy(record.Plan{Epoch: 12, MaxEpochs: 12, EpochLimitReached: true,
		Blockers: []record.PlanBlocker{{Kind: record.BlockerUpheldPosition, Subject: "red-lens-logic", Owner: "red-lens-logic"}}})
	if !strings.Contains(why, "epoch 12 of the limit 12") || !strings.Contains(why, "red-lens-logic") {
		t.Errorf("the epoch-limit exit reason: %q", why)
	}
}

// A RUN FORCED TO A POSITION ITS BENCH NEVER RULED fails the oracle, naming the ruling it owed and
// carrying every stored plan: the sitting that permitted the PASS it withheld is among them.
func TestTheForcedPositionOracleHoldsTheRuling(t *testing.T) {
	r, _ := converseRun(t, redClosesAndWithholds)
	r.forcePosition = "overruled"
	m := r.forcedPositionOracle()
	if !strings.Contains(m, "overruled ruling") {
		t.Errorf("the oracle passed a forced run whose bench ruled nothing: %q", m)
	}
	if !strings.Contains(m, "the plans, by chair sitting:") || !strings.Contains(m, "pass_permitted true") {
		t.Errorf("the oracle's failure carries no plan that permitted the PASS: %q", m)
	}
	r.forcePosition = ""
	if m := r.forcedPositionOracle(); m != "" {
		t.Errorf("the oracle fired on a run forced to no position: %s", m)
	}
}

// A POSITION SEED'S RUN IS SET UP PAST SETUP'S LIMIT, so the window its ruling's route fits holds the
// latest first permitted PASS the sweep has measured (epoch 10); limitSeed keeps its own, and every
// other run setup's default.
func TestAPositionSeedsRunHasTheWindowItsRouteNeeds(t *testing.T) {
	for _, seed := range []int64{positionLensSeed, positionGapsSeed, positionOverruledSeed} {
		limit := runEpochLimit(seed)
		if last := lastWithholdEpoch(positionFormOf(seed), limit); last < 10 {
			t.Errorf("seed %d runs under the epoch limit %d, whose %s window ends at epoch %d, short of the measured 10", seed, limit, positionFormOf(seed), last)
		}
	}
	if got := runEpochLimit(limitSeed); got != forcedEpochLimit {
		t.Errorf("limitSeed's run is set up with the limit %d, want %d", got, forcedEpochLimit)
	}
	if got := runEpochLimit(verifiedSeed); got != 0 {
		t.Errorf("a run forced to no position or limit is set up with the limit %d, want setup's default (0)", got)
	}
}

// A GAPS RUN WHOSE HELD TRIFLE NEVER LANDED says so, with each refusal of its mint; one whose trifle
// landed, and a run forced to another ruling, say nothing of it.
func TestTheTrifleClauseNamesEachRefusedMint(t *testing.T) {
	r := &runner{forcePosition: "gaps"}
	if got := r.trifleClause(); !strings.Contains(got, "no lens sitting after the first minted") {
		t.Errorf("a gaps run that never tried the trifle: %q", got)
	}
	r.trifleRefusals = []string{"red-lens-logic: refused one", "red-lens-logic: refused two"}
	if got := r.trifleClause(); !strings.Contains(got, "refused 2 time(s)") || !strings.Contains(got, "refused two") {
		t.Errorf("a gaps run whose trifle mint was refused twice: %q", got)
	}
	r.heldTrifle = true
	if got := r.trifleClause(); got != "" {
		t.Errorf("a gaps run holding its trifle: %q", got)
	}
	if got := (&runner{forcePosition: "lens"}).trifleClause(); got != "" {
		t.Errorf("a lens run: %q", got)
	}
}
