package runlive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// endRun stands for `bench outcome` having run in dir: the mark, written by hand in the shape the
// verb writes, so these tests hold the predicate to the file and not to the writer's helper.
func endRun(t *testing.T, dir, agentID string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".outcome-recorded"), []byte(`{"agentId":`+quote(agentID)+`}`), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A RUN WHOSE OUTCOME IS RECORDED IS NOT LIVE, AND DOES NOT MAKE A LIVE RUN AMBIGUOUS (#1255). The
// row an uncaptured run leaves behind counted as an open run: beside the next run it made both
// ambiguous and the hooks recorded nothing for either; alone it resolved as THE run for every later
// session in the project.
func TestARunWhoseOutcomeIsRecordedYieldsToTheLiveRun(t *testing.T) {
	proj := t.TempDir()
	ended, live := filepath.Join(proj, "research", "ended"), filepath.Join(proj, "research", "live")
	mustMkdir(t, ended)
	mustMkdir(t, live)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	endRun(t, ended, "a-bench")
	writeMarker(t, proj, `{"runs":[{"runDir":"research/ended"},{"runDir":"research/live"}]}`)

	got := InferRunDir(proj, Caller{ID: "a-seat"})
	if got.Why != Resolved || got.Dir != live {
		t.Fatalf("got %v %q, want Resolved to %q — the ended run's row still counts", got.Why, got.Dir, live)
	}

	// ALONE, the ended run resolves nothing: a later session in this project is in no run.
	writeMarker(t, proj, `{"runs":[{"runDir":"research/ended"}]}`)
	if got := InferRunDir(proj, Caller{ID: "a-dev-session-agent"}); got.Why != NoRunOpen || got.Dir != "" {
		t.Fatalf("got %v %q, want NoRunOpen — a dead run's row attributed a later session to it", got.Why, got.Dir)
	}
}

// THE SEAT THAT RECORDED THE OUTCOME KEEPS ITS RUN. The assembly seat records the outcome and then
// assembles in the same sitting; ending the run for it too would strip its identity from every later
// call and leave its sitting open on the record.
func TestTheAgentThatRecordedTheOutcomeKeepsItsRun(t *testing.T) {
	proj := t.TempDir()
	run := filepath.Join(proj, "research", "r")
	mustMkdir(t, run)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"research/r"}]}`)
	endRun(t, run, "a-bench")

	for agent, want := range map[string]Inference{"a-bench": Resolved, "a-other": NoRunOpen, "": NoRunOpen} {
		if got := InferRunDir(proj, Caller{ID: agent}); got.Why != want {
			t.Errorf("agent %q: got %v, want %v", agent, got.Why, want)
		}
	}
	// A mark naming no agent — an outcome recorded where no identity reached the verb — ends it for all.
	endRun(t, run, "")
	if got := InferRunDir(proj, Caller{}); got.Why != NoRunOpen {
		t.Errorf("a mark naming no agent: got %v, want NoRunOpen", got.Why)
	}
	// A mark that does not parse still says the outcome is recorded; only the exception is lost.
	if err := os.WriteFile(filepath.Join(run, OutcomeMark), []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := InferRunDir(proj, Caller{ID: "a-bench"}); got.Why != NoRunOpen {
		t.Errorf("an unparsable mark: got %v, want NoRunOpen", got.Why)
	}
}

// THE BENCH KEEPS A DECIDED RUN, AND A CAPTURED RUN IS NOBODY'S. A resume dispatches an assembly
// sitting that stopped after its stamp again, as a fresh agent the mark does not name; the record
// admits its acts, so the hooks bind any agent of the bench's configuration to the decided run. No
// other configuration is bound, and nothing is once capture has marked the run ended.
func TestADecidedRunIsLiveForTheBenchUntilItIsCaptured(t *testing.T) {
	proj := t.TempDir()
	run := filepath.Join(proj, "research", "r")
	mustMkdir(t, run)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"research/r"}]}`)
	endRun(t, run, "a-bench")
	m := RunLiveMarker{RunDir: "research/r"}
	for _, tc := range []struct {
		caller Caller
		want   RowState
	}{
		{Caller{ID: "a-fresh", Type: BenchAgentType}, RowLive},
		{Caller{ID: "a-bench", Type: BenchAgentType}, RowLive},
		{Caller{ID: "a-lens", Type: "frank-exchange-of-views:red-lens-evidence"}, RowDecided},
		{Caller{Type: BenchAgentType}, RowDecided}, // no agent: a main session is no seat
		{Caller{}, RowDecided},
	} {
		if got := StateOf(proj, m, tc.caller); got != tc.want {
			t.Errorf("%+v after the outcome: %v, want %v", tc.caller, got, tc.want)
		}
	}
	if err := os.WriteFile(filepath.Join(run, EndedMark), []byte("2026-10-24T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []Caller{{ID: "a-fresh", Type: BenchAgentType}, {ID: "a-bench", Type: BenchAgentType}, {}} {
		if got := StateOf(proj, m, c); got != RowEnded {
			t.Errorf("%+v once captured: %v, want RowEnded", c, got)
		}
	}
	// Captured with no outcome recorded — a run that stopped before the bench decided it — is ended too.
	if err := os.Remove(filepath.Join(run, OutcomeMark)); err != nil {
		t.Fatal(err)
	}
	if got := StateOf(proj, m, Caller{ID: "a-lens", Type: "frank-exchange-of-views:red-lens-evidence"}); got != RowEnded {
		t.Errorf("a captured run with no outcome: %v, want RowEnded", got)
	}
}

// TWO LIVE RUNS ARE NAMED. A fault that says "ambiguous" and not which runs leaves the operator to
// open the file to find out what to close.
func TestTwoLiveRunsAreAFaultThatNamesBoth(t *testing.T) {
	proj := t.TempDir()
	mustMkdir(t, filepath.Join(proj, "a"))
	mustMkdir(t, filepath.Join(proj, "b"))
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"a"},{"runDir":"b"}]}`)

	got := InferRunDir(proj, Caller{})
	if !got.Why.Fault() || got.Dir != "" {
		t.Fatalf("two live runs: Why %v, Dir %q — want a fault that resolves nothing", got.Why, got.Dir)
	}
	detail := FaultDetail(got)
	for _, want := range []string{filepath.Join(proj, "a"), filepath.Join(proj, "b"), "capture", "separate checkout"} {
		if !strings.Contains(detail, want) {
			t.Errorf("the fault must carry %q: %s", want, detail)
		}
	}
}

// PRUNING TAKES ONLY WHAT IS LIVE FOR NO ONE, AND SAYS WHY. The recording agent's exception does not
// keep a row from pruning: a pruner is no seat.
func TestPruneRemovesOnlyRowsThatAreNotLive(t *testing.T) {
	proj := t.TempDir()
	for _, d := range []string{"ended", "live"} {
		mustMkdir(t, filepath.Join(proj, d))
	}
	mustMkdir(t, filepath.Join(proj, ".claude"))
	if err := WriteOutcomeMark(filepath.Join(proj, "ended"), "a-bench"); err != nil {
		t.Fatal(err)
	}
	writeMarker(t, proj, `{"runs":[{"runDir":"gone"},{"runDir":"ended"},{"runDir":"live"}]}`)

	pruned, err := PruneRunLive(proj)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]RowState{}
	for _, p := range pruned {
		got[p.Row.RunDir] = p.State
	}
	if len(got) != 2 || got["gone"] != RowGone || got["ended"] != RowDecided {
		t.Fatalf("pruned %+v, want gone (RowGone) and ended (RowDecided)", pruned)
	}
	if rows := ReadRunLive(proj); len(rows) != 1 || rows[0].RunDir != "live" {
		t.Fatalf("kept %+v, want the live run alone", rows)
	}
	// Nothing left to prune: the file is untouched and nothing is reported.
	if again, err := PruneRunLive(proj); err != nil || len(again) != 0 {
		t.Errorf("a second prune removed %+v (%v)", again, err)
	}
	// Every row pruned: the file goes, as when capture closes the last run.
	if err := WriteOutcomeMark(filepath.Join(proj, "live"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := PruneRunLive(proj); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "run-live.json")); !os.IsNotExist(err) {
		t.Errorf("a marker with no rows left must be removed: %v", err)
	}
}
