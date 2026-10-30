package fuzz

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
// standing on it, a successor naming it, or a closing of blue's this sitting is skipped.
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
		if g.Material || g.Remanded || g.RemandStage != "" || stranded[g.ID] || r.closedThisSitting(seatID, g.ID) || len(r.unruledDocketMotions(g.ID)) > 0 {
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
// readies: named in remand_owed, with blue among the parties for it — the exchange begins with
// blue's act, and the lens-turn oracle (tallyLensTurns) holds the minting lens to waiting for it.
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
		if !engagedOn[o.GapID]["blue-respond"] {
			r.noteActErr(fmt.Sprintf("the plan names gap %s's remand as owed and does not engage blue on it: parties %v", o.GapID, parties))
			continue
		}
		r.mu.Lock()
		delete(r.trifleRemanded, o.GapID)
		r.trifleRemandsRouted++
		r.mu.Unlock()
	}
}

// A REMAND'S ITEM HOLDS BLUE UNTIL BLUE ACTS ON THE GAP (gblock, 2026-10-25), AND NEVER HOLDS THE
// MINTING LENS (ruled by gblock 2026-10-30): the exchange begins with blue's act, and the lens's turn
// on the gap comes with it.
//
// answerRemands is the act blue owes before its envelope (record.EnvelopeRefusal refuses it while
// the item blocks): a computation answering the gap — an act that moves nothing, so a gap the scenario
// leaves unanswered still stalls, goes back to the bench and is closed there. It acts on a gap only
// where the item blocks blue's list, so a sitting that already acted, or whose dispatch was made
// before the ruling, is held to nothing here. On the minting lens's list the item must not block.
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
		if role == "lens" {
			r.noteActErr(fmt.Sprintf("%s's remand item on gap %s blocks its list — the exchange begins with blue's act, so the minting lens owes the gap nothing before it", seatID, g.ID))
			continue
		}
		dir, argv, err := r.writeBundle("fuzz-remand-"+g.ID+".js", "console.log('the remand direction for "+g.ID+"');")
		if err != nil {
			continue
		}
		if _, ok := r.admit(append([]string{"compute", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--dir", dir, "--answers", g.ID,
			"--reason", "fuzz: what the remand's direction found for " + g.ID, "--"}, argv...)...); !ok {
			continue
		}
		if w2, err := record.WorkOfSeat(r.run(), role, seatID); err == nil && blocksOn(w2.Sitting, g.ID, g.DocketDirection) {
			r.noteActErr(fmt.Sprintf("%s recorded an act on remanded gap %s and the remand's item still blocks its list", seatID, g.ID))
			continue
		}
		r.mu.Lock()
		r.blueRemandActs++
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

// A REMAND'S EXCHANGE BEGINS WITH BLUE (ruled by gblock 2026-10-27), and blue's item holds the sitting
// its dispatch owes until blue acts.
//
// remandHeldBeforeRegister reads blue's list BEFORE blue registers, when blue's sitting is not yet
// blue-respond's on the record, at the first blue sitting after each chair plan, for each gap that
// plan owes a remand's exchange and that is still owed with no docket motion standing on it. Whether
// the item must block is read off the record here, not off the plan: it blocks exactly while blue has
// recorded no act on the gap since the bench's latest remand of it AND that ruling precedes blue's
// latest dispatch. A plan can owe a remand blue has already acted on — the round holding blue's act
// is not yet counted — and the record then rightly says the item no longer holds blue. On
// universe-m35 the item read "the ruling landed after this sitting opened" and did not block while
// both held, and after blue's sitting closed it was gone; either direction fails here.
func (r *runner) remandHeldBeforeRegister() {
	r.mu.Lock()
	first := !r.blueSatSincePlan
	r.blueSatSincePlan = true
	plan := r.planThisSitting
	r.mu.Unlock()
	if !first || plan == nil {
		return
	}
	raw, _ := json.Marshal(plan["remand_owed"])
	var owedByPlan []struct {
		GapID     string `json:"gap_id"`
		Direction string `json:"direction"`
	}
	_ = json.Unmarshal(raw, &owedByPlan)
	if len(owedByPlan) == 0 {
		return
	}
	w, err := record.WorkOfSeat(r.run(), "blue", "blue-respond")
	if err != nil {
		r.noteActErr("blue's work list could not be read before it registered: " + err.Error())
		return
	}
	b, ok := r.board()
	if !ok {
		r.noteActErr("the record could not be read before blue registered")
		return
	}
	acts, _, _ := actsOnGaps(b.Events)
	ruled := remandRulings(b.Events)
	dispatched := -1 // blue's latest dispatch row, a stream position
	for i, e := range b.Events {
		if d, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok && d.GetSeatId() == "blue-respond" {
			dispatched = i
		}
	}
	owed := map[string]bool{}
	for _, g := range w.Open {
		owed[g.ID] = g.RemandStage == "owed"
	}
	for _, o := range owedByPlan {
		if !owed[o.GapID] || len(r.unruledDocketMotions(o.GapID)) > 0 {
			continue // closed or put before the bench again since: no exchange is owed, and no item stands
		}
		acted := false
		for _, a := range acts[o.GapID] {
			acted = acted || a.at > ruled[o.GapID] && seatclass.RoleOfSeat(a.seat) == "blue"
		}
		want := !acted && ruled[o.GapID] < dispatched
		switch got := blocksOn(w.Sitting, o.GapID, o.Direction); {
		case want && !got:
			r.noteActErr(fmt.Sprintf("the plan dispatching blue owes remanded gap %s its exchange, blue has recorded no act on it since the ruling, and before blue registered its remand item does not block its list", o.GapID))
		case !want && got:
			r.noteActErr(fmt.Sprintf("blue's remand item on gap %s blocks its list before blue registered, though %s", o.GapID,
				map[bool]string{true: "blue has recorded an act on it since the ruling", false: "the ruling landed after blue's latest dispatch"}[acted]))
		case got:
			r.mu.Lock()
			r.remandHolds++
			r.mu.Unlock()
		default:
			r.mu.Lock()
			r.remandReleases++
			r.mu.Unlock()
		}
	}
}
