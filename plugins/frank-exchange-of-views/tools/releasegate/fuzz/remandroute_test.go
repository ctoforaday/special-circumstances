package fuzz

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE BENCH'S REMAND, ON A GAP THAT IS NOT MATERIAL AND ON A CLOSED ONE (gblock, 2026-10-24).
//
// A remand of an open gap routes its minting lens and blue to one more exchange whatever its
// materiality, and holds the PASS until that exchange has begun; a remand of a closed gap is refused,
// naming the close, and the bench rules the motion with a disposition that closes. On universe-m33
// both remands were stored and never acted on, and the run passed against the bench's holding.
//
// The drive reaches both: blue escalates a gap that is not material to the bench, which the
// dispatch never dockets on its own, and the bench sitting convened on a docket whose gap has closed
// since it was filed reaches for a remand before it rules. The sweep gates each at zero.

// escalateTrifle files blue's docket motion on an open gap that is not material, so the bench's
// remand of such a gap is a path the drive takes. One with a remand already on it, a docket motion
// standing on it, a successor naming it, a closing of blue's this sitting, or benchOrdersSeed's
// construction holding it (arcHolds) is skipped.
func (r *runner) escalateTrifle(seatID string) {
	if !r.coin(25) {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "blue", seatID)
	if err != nil {
		r.noteActErr("blue's work list could not be read to escalate a gap that is not material: " + err.Error())
		return
	}
	stranded := map[string]bool{}
	if bs, err := record.PassBlockers(r.run()); err == nil {
		for _, b := range bs {
			if b.Kind == record.BlockerStrandedGap {
				stranded[b.Subject] = true
			}
		}
	}
	var trifles []string
	for _, g := range w.Open {
		if g.Material || g.Remanded || g.RemandStage != "" || stranded[g.ID] || r.arcHolds(g.ID) || r.closedThisSitting(seatID, g.ID) || len(r.unruledDocketMotions(g.ID)) > 0 {
			continue
		}
		trifles = append(trifles, g.ID)
	}
	if len(trifles) == 0 {
		return
	}
	id := trifles[r.rng.Intn(len(trifles))]
	r.admitFresh(id, "motion", "docket", "file", "--seat-id", seatID, "--id", id,
		"--reason", "fuzz: "+id+" is not material, and blue escalates it to the bench all the same")
}

// closeDocketed is a lens closing a gap of its own while a docket motion on it stands — a real act
// (closingowed.go: its lens closed it, the dispute is over, and the bench rules the motion on the
// record as it stands), and the one that puts a closed gap before the bench, whose sitting then
// reaches for the remand its scenario asks for. A gap is closed only where the lens's own close
// would close it (closableNow).
func (r *runner) closeDocketed(seatID string) {
	if !r.coin(30) {
		return
	}
	for _, id := range r.openGaps() {
		if r.minterOf(id) != seatID || !r.closableNow(id) || len(r.unruledDocketMotions(id)) == 0 {
			continue
		}
		r.closeGap(seatID, id, false)
		return
	}
}

// closableNow is the lens arm's close condition on gap id, by its scenario: satisfied once blue has
// sat on it, a computation once it has reproduced, a dispute red loses once blue has argued it.
func (r *runner) closableNow(id string) bool {
	switch d := r.scenarioOf(id); {
	case satisfied(d):
		return r.presented.has(id)
	case d == dirCompute:
		return r.reproduced.has(id)
	case d == dirDisputeWon:
		return r.presented.has(id)
	}
	return false
}

// remandOfClosedRefused drives the bench's remand of every docket motion standing on gapID, a gap
// that has closed since the motion was filed, and holds the tool to refusing each as a conflict that
// names the gap, its closing and that nothing was recorded.
func (r *runner) remandOfClosedRefused(seatID, gapID string) {
	for _, motionID := range r.unruledDocketMotions(gapID) {
		out, err := r.exec("motion", "docket", "rule", "--seat-id", seatID, "--id", motionID,
			"--as", "remanded", "--direction", "fuzz: one more exchange on a gap that has closed",
			"--principle", "correctness", "--tension", "cost", "--review-flag", "none: driven to be refused",
			"--settled", settledByFuzzBench, "--reason", "fuzz: a remand of a closed gap", "--json")
		env, ok := refusalEnvelopeIn(out)
		switch {
		case err == nil:
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on closed gap %s was admitted: %s", motionID, gapID, firstLine(out)))
		case !ok || env.Code != string(feov.Conflict):
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on closed gap %s was refused, and not as a conflict: %s", motionID, gapID, firstLine(out)))
		case !strings.Contains(env.Error, gapID) || !strings.Contains(env.Error, "CLOSED") || !strings.Contains(env.Error, "closed it at event") || !strings.Contains(env.Error, "Nothing was recorded"):
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on closed gap %s was refused without naming the gap and its closing: %s", motionID, gapID, firstLine(env.Error)))
		default:
			r.mu.Lock()
			r.closedRemandsRefused++
			r.mu.Unlock()
		}
	}
}

// remandRouted holds the gate's list to a remand the bench just ruled on open gap gapID: a gap that
// is not material — no material_gap or stranded_gap holds it — holds a remand_owed blocker until its
// exchange has begun, and the drive remembers it, so the chair's next plan is held to readying it.
func (r *runner) remandRouted(gapID string) {
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		r.noteActErr("the PASS gate's list could not be read after a remand: " + err.Error())
		return
	}
	material, owed := false, false
	for _, b := range bs {
		if b.Subject != gapID {
			continue
		}
		switch b.Kind {
		case record.BlockerMaterialGap, record.BlockerStrandedGap:
			material = true
		case record.BlockerRemandOwed:
			owed = true
		}
	}
	switch {
	case material && owed:
		r.noteActErr(fmt.Sprintf("gap %s holds the PASS both as material and as a remand owed — the remand_owed blocker is a gap's that is not material", gapID))
	case !material && !owed && r.remandStands(gapID):
		r.noteActErr(fmt.Sprintf("the bench remanded gap %s, which is not material, and the PASS gate's list holds no remand_owed for it — the remand would be stored and never acted on", gapID))
	case owed:
		r.mu.Lock()
		if r.trifleRemanded == nil {
			r.trifleRemanded = map[string]bool{}
		}
		r.trifleRemanded[gapID] = true
		r.mu.Unlock()
	}
}

// remandStands reports whether the record holds gapID open and remanded with nothing pending — the
// state a remand the bench just ruled leaves; a ruling the drive never landed is the act's own
// failure, reported where it was refused.
func (r *runner) remandStands(gapID string) bool {
	w, err := record.WorkOfSeat(r.run(), "blue", "blue-respond")
	if err != nil {
		return false
	}
	for _, g := range w.Open {
		if g.ID == gapID {
			return g.Remanded
		}
	}
	return false
}

// tallyTrifleRoutes counts each remanded gap that is not material whose one more exchange the plan
// readies: named in remand_owed, with its minting lens and blue among the parties for it.
func (r *runner) tallyTrifleRoutes(plan map[string]any) {
	r.mu.Lock()
	pending := map[string]bool{}
	for g := range r.trifleRemanded {
		pending[g] = true
	}
	r.mu.Unlock()
	// A plan at the epoch limit dispatches nobody, whatever it names as owed.
	if len(pending) == 0 || plan["epoch_limit_reached"] == true {
		return
	}
	raw, _ := json.Marshal(plan["remand_owed"])
	var owed []struct {
		GapID string `json:"gap_id"`
	}
	_ = json.Unmarshal(raw, &owed)
	engagedOn := map[string]map[string]bool{} // gap -> seats the plan engages on it
	parties, _ := plan["parties"].([]any)
	for _, p := range parties {
		m, _ := p.(map[string]any)
		gs, _ := m["gap_ids"].([]any)
		for _, g := range gs {
			id := fmt.Sprint(g)
			if engagedOn[id] == nil {
				engagedOn[id] = map[string]bool{}
			}
			engagedOn[id][fmt.Sprint(m["seat_id"])] = true
		}
	}
	for _, o := range owed {
		if !pending[o.GapID] {
			continue
		}
		if !engagedOn[o.GapID]["blue-respond"] || !engagedOn[o.GapID][r.minterOf(o.GapID)] {
			r.noteActErr(fmt.Sprintf("the plan names gap %s's remand as owed and does not engage its minting lens and blue on it: parties %v", o.GapID, parties))
			continue
		}
		r.mu.Lock()
		delete(r.trifleRemanded, o.GapID)
		r.trifleRemandsRouted++
		r.mu.Unlock()
	}
}

// A REMAND'S ITEM HOLDS EACH SEAT IT READIES UNTIL THAT SEAT ACTS ON THE GAP (gblock, 2026-10-25).
//
// answerRemands is the act each owes before its envelope (record.EnvelopeRefusal refuses it while
// the item blocks): blue a computation answering the gap — an act that moves nothing, so a gap the scenario
// leaves unanswered still stalls, goes back to the bench and is closed there — and the minting lens a
// regrade that keeps the grade. It acts on a gap only where the item blocks this seat's list, so a
// seat that already acted, or whose dispatch was made before the ruling, is held to nothing here.
func (r *runner) answerRemands(seatID string) {
	role := "lens"
	if seatID == "blue-respond" {
		role = "blue"
	}
	w, err := record.WorkOfSeat(r.run(), role, seatID)
	if err != nil {
		r.noteActErr(seatID + "'s work list could not be read for the remands it owes: " + err.Error())
		return
	}
	for _, g := range w.Open {
		if g.RemandStage != "owed" || g.DocketDirection == "" || !blocksOn(w.Sitting, g.ID, g.DocketDirection) {
			continue
		}
		ok := false
		if role == "blue" {
			dir, argv, err := r.writeBundle("fuzz-remand-"+g.ID+".js", "console.log('the remand direction for "+g.ID+"');")
			if err != nil {
				continue
			}
			_, ok = r.admit(append([]string{"compute", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--dir", dir, "--answers", g.ID,
				"--reason", "fuzz: what the remand's direction found for " + g.ID, "--"}, argv...)...)
		} else {
			_, ok = r.admitFresh(g.ID, "regrade", "--seat-id", seatID, "--id", g.ID,
				"--reason", "fuzz: blue's answer re-audited against the remand's direction; the acceptance check still fails")
		}
		if !ok {
			continue
		}
		if w2, err := record.WorkOfSeat(r.run(), role, seatID); err == nil && blocksOn(w2.Sitting, g.ID, g.DocketDirection) {
			r.noteActErr(fmt.Sprintf("%s recorded an act on remanded gap %s and the remand's item still blocks its list", seatID, g.ID))
			continue
		}
		if role == "lens" {
			r.noteLensFirst(g.ID, g.DocketDirection)
		}
		r.mu.Lock()
		if r.remandActs == nil {
			r.remandActs = map[string]int{}
		}
		r.remandActs[role]++
		r.mu.Unlock()
	}
}

// blocksOn reports whether a blocking item on the list is the remand's on gapID: it names the gap
// and carries the ruling's direction, which no other item does.
func blocksOn(s record.SittingJSON, gapID, direction string) bool {
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, "gap "+gapID) && strings.Contains(it.What, direction) {
			return true
		}
	}
	return false
}

// remandPastLimit reports whether the bench's work list says a remand of gapID is refused — the
// gap's one remand at impasse is had — so the drive's bench reaches for the remand and is held to
// the refusal, and rules a disposition that closes instead.
func (r *runner) remandPastLimit(seatID, gapID string) bool {
	w, err := record.WorkOfSeat(r.run(), "bench", seatID)
	if err != nil {
		r.noteActErr("the bench's work list could not be read before it ruled: " + err.Error())
		return false
	}
	for _, it := range w.Sitting.Open {
		if strings.Contains(it.What, "a remand of "+gapID+" is refused") {
			return true
		}
	}
	return false
}

// remandPastLimitRefused drives the bench's remand of every docket motion standing on gapID, a gap
// its work list says is at its remand limit, and holds the tool to refusing each as a conflict that
// names the motion, the gap, the limit and every disposition that closes the gap.
func (r *runner) remandPastLimitRefused(seatID, gapID string) {
	for _, motionID := range r.unruledDocketMotions(gapID) {
		out, err := r.exec("motion", "docket", "rule", "--seat-id", seatID, "--id", motionID,
			"--as", "remanded", "--direction", "fuzz: a second remand at impasse",
			"--principle", "correctness", "--tension", "cost", "--review-flag", "none: driven to be refused",
			"--settled", settledByFuzzBench, "--reason", "fuzz: a remand past the limit", "--json")
		env, ok := refusalEnvelopeIn(out)
		named := ok && strings.Contains(env.Error, motionID) && strings.Contains(env.Error, gapID) && strings.Contains(env.Error, record.RemandLimitRule)
		for _, d := range record.ClosingDispositions() {
			named = named && strings.Contains(env.Error, d)
		}
		switch {
		case err == nil:
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on gap %s, past its remand limit, was admitted: %s", motionID, gapID, firstLine(out)))
		case !ok || env.Code != string(feov.Conflict):
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on gap %s, past its remand limit, was refused, and not as a conflict: %s", motionID, gapID, firstLine(out)))
		case !named:
			r.noteActErr(fmt.Sprintf("the bench's remand of motion %s on gap %s was refused without naming the limit and the dispositions that close the gap: %s", motionID, gapID, firstLine(env.Error)))
		default:
			r.mu.Lock()
			r.pastLimitRemandsRefused++
			r.mu.Unlock()
		}
	}
}

// THE LENS SITS FIRST, AND ITS ACT BEGINS NO REMAND'S EXCHANGE (ruled by gblock 2026-10-27).
//
// debate.js dispatches the lenses before blue in an epoch, so the minting lens records its act on a
// remanded gap before blue has answered. noteLensFirst remembers each gap the lens acted on while
// blue's remand item on it still blocked — blue had not acted since the ruling — and the stage
// still reads owed: the lens's act alone has begun nothing. lensFirstHeld then reads blue's list
// BEFORE blue registers, when blue's sitting is not yet blue-respond's on the record: the item on each
// such gap must be there and must block, with no reason saying the ruling came after blue's dispatch.
// On universe-m35 it read "the ruling landed after this sitting opened" and did not block, and after
// blue's sitting closed it was gone.

// noteLensFirst records gapID as a remanded gap its minting lens acted on before blue, holding the
// stage to owed while blue's item still blocks.
func (r *runner) noteLensFirst(gapID, direction string) {
	w, err := record.WorkOfSeat(r.run(), "blue", "blue-respond")
	if err != nil {
		r.noteActErr("blue's work list could not be read after the lens acted on a remanded gap: " + err.Error())
		return
	}
	if !blocksOn(w.Sitting, gapID, direction) {
		return // blue acted on it since the ruling: the lens is not first
	}
	for _, g := range w.Open {
		if g.ID == gapID && g.RemandStage != "owed" {
			r.noteActErr(fmt.Sprintf("the minting lens acted on remanded gap %s before blue did, and its remand_stage reads %q — the lens's act alone began the exchange", gapID, g.RemandStage))
			return
		}
	}
	r.mu.Lock()
	if r.lensFirst == nil {
		r.lensFirst = map[string]string{}
	}
	r.lensFirst[gapID] = direction
	r.mu.Unlock()
}

// lensFirstHeld holds blue's list, read before blue registers, to each remanded gap the lens acted on
// first: while the remand is owed, its item is on the list and blocks.
func (r *runner) lensFirstHeld() {
	r.mu.Lock()
	pending := map[string]string{}
	for g, d := range r.lensFirst {
		pending[g] = d
	}
	r.mu.Unlock()
	if len(pending) == 0 {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "blue", "blue-respond")
	if err != nil {
		r.noteActErr("blue's work list could not be read before it registered: " + err.Error())
		return
	}
	owed := map[string]bool{}
	for _, g := range w.Open {
		owed[g.ID] = g.RemandStage == "owed"
	}
	for g, d := range pending {
		r.mu.Lock()
		delete(r.lensFirst, g)
		r.mu.Unlock()
		if !owed[g] {
			continue // closed or docketed again since: no exchange is owed, and no item stands
		}
		if !blocksOn(w.Sitting, g, d) {
			r.noteActErr(fmt.Sprintf("the minting lens acted on remanded gap %s first, and before blue registered its remand item does not block its list — the exchange is owed and blue has not acted", g))
			continue
		}
		r.mu.Lock()
		r.lensFirstHolds++
		r.mu.Unlock()
	}
}
