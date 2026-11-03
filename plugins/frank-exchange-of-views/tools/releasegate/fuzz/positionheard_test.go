package fuzz

import (
	"encoding/json"
	"fmt"
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
// one, and its bench rules the position the way the seed names.
const (
	positionLensSeed      = 10
	positionGapsSeed      = 11
	positionOverruledSeed = 12
)

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

// withholds is whether a chair whose plan permits the PASS withholds it. A run forced to a position
// withholds its first permitted PASS, and the gaps run only once a gap stands open for the bench to
// name; every other forced run never withholds, and a drawn run on a coin.
func (r *runner) withholds() bool {
	if r.forcePosition == "" {
		return !r.forcedVerdict() && r.coin(5)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.positionForced || (r.forcePosition == "gaps" && !r.heldTrifle) {
		return false
	}
	r.positionForced = true
	return true
}

// mintHeldTrifle is the gaps run's lens minting, once, a gap of a class that is never material, at
// a sitting after the first where it holds no gap of its own open (the budget refuses a second
// open one). A gap that is not material readies no seat and holds no PASS, so it stands open, with
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
	if r.mintWith(seatID, "", class) == "" {
		r.mu.Lock()
		r.heldTrifle = false
		r.mu.Unlock()
	}
}

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

// rulePositions rules every position the bench hears, at the sitting of the given occasion. A run
// forced to a position rules it the way it is forced; any other forced run's bench overrules, and a
// drawn run's rules on a coin. Upheld names open gaps the tool accepts, or a cast lens.
func (r *runner) rulePositions(seatID, occasion string) {
	if r.positionRuledAt == nil {
		r.positionRuledAt = map[string]string{}
	}
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
		return fmt.Sprintf("forced-position oracle: the run is forced to a %s ruling on a withheld PASS, and its bench made none (rulings %v, %d PASS withheld)", r.forcePosition, got, r.positionsWithheld)
	}
	open := r.openGaps()
	for _, g := range named {
		if slices.Contains(open, g) {
			return fmt.Sprintf("forced-position oracle: the bench upheld the withheld PASS's position naming gap %s, and the gap is still open at the run's end — the remand's exchange never closed it", g)
		}
	}
	return ""
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

// A FORCED CHAIR WITHHOLDS ONCE, and the gaps run's only once its lens holds a gap open to name.
func TestTheForcedChairWithholdsItsFirstPermittedPass(t *testing.T) {
	r := &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: "lens"}
	if !r.withholds() || r.withholds() {
		t.Error("the lens run's chair must withhold its first permitted PASS, and only that one")
	}
	r = &runner{rng: newLockedRand(1), forceVerified: true, forcePosition: "gaps"}
	if r.withholds() {
		t.Error("the gaps run's chair withheld before its lens minted the gap it holds open")
	}
	r.heldTrifle = true
	if !r.withholds() {
		t.Error("the gaps run's chair did not withhold once the gap stood open")
	}
	r = &runner{rng: newLockedRand(1), forceVerified: true}
	for range 50 {
		if r.withholds() {
			t.Fatal("a forced run with no position to force withheld a permitted PASS")
		}
	}
}

// A RUN FORCED TO A POSITION ITS BENCH NEVER RULED fails the oracle, naming the ruling it owed.
func TestTheForcedPositionOracleHoldsTheRuling(t *testing.T) {
	r, _ := converseRun(t, redClosesAndWithholds)
	r.forcePosition = "overruled"
	if m := r.forcedPositionOracle(); !strings.Contains(m, "overruled ruling") {
		t.Errorf("the oracle passed a forced run whose bench ruled nothing: %q", m)
	}
	r.forcePosition = ""
	if m := r.forcedPositionOracle(); m != "" {
		t.Errorf("the oracle fired on a run forced to no position: %s", m)
	}
}
