package fuzz

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
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
		[]byte(`{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":1,"k":1,"kMax":6,"convergenceFraction":0.25,"maxEpochs":12}`)); err != nil {
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
// settled debate's return.
func (r *runner) blueFailsForcedErr(result map[string]any) string {
	if !r.forceBlueFails {
		return ""
	}
	r.mu.Lock()
	sat := r.blueSittings
	r.mu.Unlock()
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
