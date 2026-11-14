package fuzz

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A FAILED BLUE SITTING STALLS NO GAP (ruled by gblock 2026-11-11), driven through the binary and the
// writer the hooks spawn. Under k=1 one stalled exchange puts a gap at impasse: blue engaged on it
// sits once and records nothing. Returning its envelope, that sitting is a null turn and the gap is
// docketed; ending without it (dispatchReturns writes the close and no admission), it is no turn, the
// gap stands where it stood, and blue is engaged on it again.
func TestAFailedBlueSittingStallsNoGap(t *testing.T) {
	for _, c := range []struct {
		name     string
		admitted bool
		stalled  int
	}{
		{"blue returns its envelope: a null turn", true, 1},
		{"blue ends without its envelope: no turn", false, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, gap := blueEngagedRun(t)
			r.dispatchSits("blue", "blue-respond")
			r.dispatchReturns("blue-respond", c.admitted)
			if r.actErr != "" {
				t.Fatalf("blue's return: %s", r.actErr)
			}
			if failed, err := record.FailedSittings(r.run()); err != nil || len(failed) != 1-btoi(c.admitted) {
				t.Fatalf("the record reads %+v as failed (%v), want blue's sitting failed exactly when it ended without its envelope", failed, err)
			}
			p, err := record.RunParams(r.run())
			if err != nil {
				t.Fatal(err)
			}
			x, err := record.Exchanges(r.run(), p)
			if err != nil {
				t.Fatal(err)
			}
			if g := x[gap]; g == nil || g.Stalled != c.stalled || g.Exchanges != c.stalled || g.Impasse != (c.stalled > 0) {
				t.Fatalf("%s = %+v, want %d stalled exchange(s)", gap, g, c.stalled)
			}
			plan := chairPlan(t, r)
			if docketed := slices.Contains(plan.Docket, gap); docketed != (c.stalled > 0) {
				t.Errorf("the next plan dockets %v, want %s docketed only after a null turn", plan.Docket, gap)
			}
			if c.stalled == 0 && !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "blue-respond" && slices.Contains(p.GapIDs, gap) }) {
				t.Errorf("after its failed sitting blue is not engaged on %s again: parties %+v", gap, plan.Parties)
			}
		})
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// blueEngagedRun is a run under k=1 whose lens has minted one material gap, and whose chair's plan
// engages blue on it. Every opening carries an agent, so each return is what the hooks write.
func blueEngagedRun(t *testing.T) (*runner, string) {
	t.Helper()
	const lens, sentence = "red-lens-logic", "A § fuzz sentence to anchor findings."
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	if err := writeSetup(filepath.Join(runDir, "inputs", "run-config.json"),
		fuzzRunConfig(1, 1, 6, 0, 12)); err != nil {
		t.Fatal(err)
	}
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: []string{lens, "red-chair", "blue-lane-1", "blue-synthesize", "blue-respond", "judge"}, LaneSeatIds: []string{"blue-lane-1"}}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(7))
	r.openUnderAgent = true
	r.forceVerified = true // no draw leaves a sitting open or a return unadmitted
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	chairPlan(t, r)
	r.dispatchSits("lens", lens)
	for _, args := range [][]string{
		{"--json", "--seat-id", lens, "show", "report"},
		{"--json", "--seat-id", lens, "mint", "--quote", sentence, "--class", "false_universal", "--check-kind", "document",
			"--check", "acc", "--problem", "fuzz problem", "--fix", "fuzz fix",
			"--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason"},
	} {
		if out, err := r.exec(args...); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
	}
	r.dispatchReturns(lens, true)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	plan := chairPlan(t, r)
	if !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == "blue-respond" && slices.Contains(p.GapIDs, open[0]) }) {
		t.Fatalf("the plan does not engage blue on %s: parties %+v, why %q", open[0], plan.Parties, plan.Why)
	}
	return r, open[0]
}

// chairPlan is a chair sitting: it opens, asks the record who sits, and returns admitted.
func chairPlan(t *testing.T, r *runner) record.Plan {
	t.Helper()
	r.dispatchSits("chair", "red-chair")
	out, err := r.exec("--json", "dispatch", "next", "--seat-id", "red-chair")
	if err != nil {
		t.Fatalf("dispatch next: %v\n%s", err, out)
	}
	r.dispatchReturns("red-chair", true)
	var plan record.Plan
	if err := json.Unmarshal([]byte(out), &plan); err != nil {
		t.Fatalf("the plan does not read: %v\n%s", err, out)
	}
	return plan
}

// blueFailsForcedErr is the run's error where blueFailsSeed's run did not end on blue's failed-sitting
// bound: every blue sitting in it ends without its envelope, so the engine stops the debate UNVERIFIED
// at the third, naming blue (debate.js seatStop), and nothing else may end the run first. result is the
// settled debate's return. runOne asks it holding r.mu, as it asks the other forced runs' errors, so
// it reads the runner's counts without taking the lock.
func (r *runner) blueFailsForcedErr(result map[string]any) string {
	if !r.forceBlueFails {
		return ""
	}
	sat := r.blueSittings
	term, _ := result["termination"].(map[string]any)
	stop, _ := term["seat_stop"].(map[string]any)
	failures, _ := stop["failures"].(int64)
	if verdict, _ := result["verdict"].(string); verdict != "UNVERIFIED" || stop["seat"] != "blue-respond" || failures != noProgressEpochs {
		return fmt.Sprintf("forced blue failures: seed %d's run ended %v with termination %v after %d blue sitting(s), want UNVERIFIED on blue-respond's %d failed sittings in a row",
			blueFailsSeed, result["verdict"], term, sat, noProgressEpochs)
	}
	return ""
}

// THE FORCED RUN'S ORACLE READS THE STOP OFF THE SETTLED RETURN, and fires on any other ending.
func TestBlueFailsForcedErrReadsTheSeatStop(t *testing.T) {
	stopped := func(seat string, failures int64, verdict string) map[string]any {
		return map[string]any{"verdict": verdict, "termination": map[string]any{
			"seat_stop": map[string]any{"seat": seat, "failures": failures}, "no_progress": nil}}
	}
	r := &runner{forceBlueFails: true}
	if msg := r.blueFailsForcedErr(stopped("blue-respond", noProgressEpochs, "UNVERIFIED")); msg != "" {
		t.Errorf("blue's stop read as a miss: %s", msg)
	}
	for name, res := range map[string]map[string]any{
		"the chair's stop":       stopped("red-chair", noProgressEpochs, "UNVERIFIED"),
		"a stop at two failures": stopped("blue-respond", 2, "UNVERIFIED"),
		"a CEILING":              stopped("blue-respond", noProgressEpochs, "CEILING"),
		"no stop":                {"verdict": "UNVERIFIED", "termination": map[string]any{"seat_stop": nil}},
		"no termination":         {"verdict": "UNVERIFIED"},
	} {
		if r.blueFailsForcedErr(res) == "" {
			t.Errorf("%s read as blue's stop", name)
		}
	}
	if (&runner{}).blueFailsForcedErr(map[string]any{}) != "" {
		t.Error("an unforced run is held to blue's stop")
	}
}

// THE LENS-TURN ORACLE READS A LENS SITTING THAT DID NOT FINISH AS NO TURN, as the record does
// (ruled by gblock 2026-11-11). Blue answers the gap, and the plan readies its lens to audit the answer.
// A lens sitting that returns its envelope in silence takes the turn: the next plan engages nobody on
// the gap. One that fails audited nothing: the next plan readies the lens on blue's answer again, and
// the oracle holds that plan to the same rule without firing.
func TestTheLensTurnOracleReadsANoTurnLensSittingAsTheRecordDoes(t *testing.T) {
	const lens, blue = "red-lens-logic", "blue-respond"
	for _, c := range []struct {
		name     string
		admitted bool
		readied  bool
	}{
		{"the lens returns its envelope: a null turn", true, false},
		{"the lens ends without its envelope: no turn", false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, gap := blueEngagedRun(t)
			r.dispatchSits("blue", blue)
			if _, ok := r.do("edit", blue).set("--quote", "fuzz sentence").set("--new", "fuzz sentences").
				set("--answers", gap).set("--reason", "fuzz: blue's answer to the gap").admit(); !ok {
				t.Fatalf("blue's edit: %s", r.actErr)
			}
			r.dispatchReturns(blue, true)
			if plan := chairPlan(t, r); !slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == lens && slices.Contains(p.GapIDs, gap) }) {
				t.Fatalf("after blue's answer the plan does not ready %s on %s: parties %+v, why %q", lens, gap, plan.Parties, plan.Why)
			}
			r.dispatchSits("lens", lens)
			r.dispatchReturns(lens, c.admitted)
			if r.actErr != "" {
				t.Fatalf("the lens's return: %s", r.actErr)
			}
			plan := chairPlan(t, r)
			if readied := slices.ContainsFunc(plan.Parties, func(p record.Party) bool { return p.SeatID == lens && slices.Contains(p.GapIDs, gap) }); readied != c.readied {
				t.Fatalf("after the lens's sitting the plan readies %s on %s: %v, want %v (parties %+v, why %q)", lens, gap, readied, c.readied, plan.Parties, plan.Why)
			}
			raw, err := json.Marshal(plan)
			if err != nil {
				t.Fatal(err)
			}
			var relayed map[string]any
			if err := json.Unmarshal(raw, &relayed); err != nil {
				t.Fatal(err)
			}
			r.tallyLensTurns(relayed)
			if r.actErr != "" {
				t.Fatalf("the oracle reads the plan as breaking the lens-turn rule: %s", r.actErr)
			}
		})
	}
}

// A DOCKET MOTION FILED AFTER THE LAST PLAN IS THE ASSEMBLY'S TO RULE, and only that one. Blue, sitting
// on the plan's gap, files the docket motion itself; a stop then landing mid-turn convenes no terminal
// sitting for it, because the plan's blockers predate it. Once a plan is stored after it, that plan's
// blockers name it and the terminal sitting rules it, so the assembly leaves it alone.
func TestTheAssemblyRulesOnlyTheDocketsFiledAfterTheLastPlan(t *testing.T) {
	r, gap := blueEngagedRun(t)
	r.dispatchSits("blue", "blue-respond")
	out, ok := r.do("motion docket file", "blue-respond").set("--id", gap).set("--reason", "fuzz: blue escalates "+gap).bare("--json").admit()
	if !ok {
		t.Fatalf("blue's docket filing: %s", r.actErr)
	}
	var filed struct {
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &filed); err != nil || filed.Result.MotionID == "" {
		t.Fatalf("the filing names no motion (%v): %s", err, out)
	}
	if got := r.docketsSinceTheLastPlan(); !slices.Equal(got, []string{filed.Result.MotionID}) {
		t.Fatalf("after blue's filing the assembly would rule %v, want [%s]", got, filed.Result.MotionID)
	}
	r.dispatchReturns("blue-respond", true)
	chairPlan(t, r)
	if got := r.docketsSinceTheLastPlan(); len(got) != 0 {
		t.Fatalf("a plan stored after the filing names it, and the assembly would still rule %v", got)
	}
	if r.actErr != "" {
		t.Fatal(r.actErr)
	}
}

// THE FORCED-UNVERIFIED RUN'S SEVERITIES ARE THE RECORD'S MATERIAL ONES. For each pool class, a run
// mints a gap of it at each grade, one lens and one sentence to a grade, and reads back the record's
// materiality of each: material exactly at the grades materialGrades lists. A list that named one grade too low would
// let that run draw a gap the PASS ignores.
func TestTheMaterialGradesAreTheRecordsMaterialGrades(t *testing.T) {
	lenses := []string{"red-lens-logic", "red-lens-evidence", "red-lens-voice", "red-lens-adversary", "red-lens-architecture"}
	sentence := func(i int) string { return fmt.Sprintf("A § fuzz sentence number %d to anchor findings.", i) }
	var part strings.Builder
	for i := range grades {
		part.WriteString(sentence(i) + "\n\n")
	}
	if len(lenses) < len(grades) {
		t.Fatalf("%d lenses for %d grades: each grade needs a lens of its own", len(lenses), len(grades))
	}
	bin := buildBinary(t)
	for _, class := range mintPoolClasses {
		t.Run(class, func(t *testing.T) {
			runDir := recordtest.TmpRun(t)
			stageRun, err := record.NewRun(runDir)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat},
				&recordpb.Cast{SeatIds: append(slices.Clone(lenses), "red-chair", "blue-lane-1", "blue-synthesize", "blue-respond", "judge"), LaneSeatIds: []string{"blue-lane-1"}}); err != nil {
				t.Fatal(err)
			}
			r := newRunner(bin, runDir, newLockedRand(7))
			if err := r.stagePart(part.String()); err != nil {
				t.Fatalf("record the lane's part: %v", err)
			}
			for i, grade := range grades {
				lens := lenses[i]
				r.sit("lens", lens)
				out, err := r.exec("--json", "--seat-id", lens, "mint", "--quote", sentence(i), "--class", class, "--check-kind", "document",
					"--check", "acc", "--problem", "fuzz problem at "+grade, "--fix", "fuzz fix",
					"--severity", grade, "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
				if err != nil {
					t.Fatalf("mint at %s: %v\n%s", grade, err, out)
				}
				gap := mintedGapOf(t, out)
				b, ok := r.board()
				if !ok || b.Gap(gap) == nil {
					t.Fatalf("the board does not read %s", gap)
				}
				if got, want := b.Gap(gap).Material, slices.Contains(materialGrades, grade); got != want {
					t.Errorf("a %s gap at severity %s reads material %v, and materialGrades says %v", class, grade, got, want)
				}
			}
		})
	}
}

// mintedGapOf is the gap id a `--json mint` printed.
func mintedGapOf(t *testing.T, out string) string {
	t.Helper()
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.GapID == "" {
		t.Fatalf("the mint names no gap (%v): %s", err, out)
	}
	return env.Result.GapID
}
