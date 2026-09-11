package sittingcap

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

// runWithLimit is a run directory whose run-config states the limit.
func runWithLimit(t *testing.T, limit int) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"topic":"t","` + ConfigKey + `":` + strconv.Itoa(limit) + `}`
	if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func mustOpen(t *testing.T, run, agent, seat string, sitting int) {
	t.Helper()
	if err := Open(run, agent, Header{SeatID: seat, Sitting: sitting}); err != nil {
		t.Fatal(err)
	}
}

func count(t *testing.T, run, agent string) Decision {
	t.Helper()
	d, ok, err := Count(run, agent)
	if err != nil || !ok {
		t.Fatalf("Count(%s) = ok %v err %v, want a counted call", agent, ok, err)
	}
	return d
}

// CALL N RUNS AND CALL N+1 DOES NOT.
func TestCallNPassesAndCallNPlusOneIsOver(t *testing.T) {
	run := runWithLimit(t, 3)
	mustOpen(t, run, "agent_01", "red-lens-evidence-L1", 1)
	for i := 1; i <= 3; i++ {
		if d := count(t, run, "agent_01"); d.Over || d.Count != i {
			t.Fatalf("call %d: count %d over %v, want count %d and not over", i, d.Count, d.Over, i)
		}
	}
	d := count(t, run, "agent_01")
	if !d.Over || d.Count != 4 || d.Limit != 3 {
		t.Errorf("call 4: %+v, want over with count 4 and limit 3", d)
	}
	if d.SeatID != "red-lens-evidence-L1" || d.Sitting != 1 {
		t.Errorf("the decision names %s sitting %d, want red-lens-evidence-L1 sitting 1", d.SeatID, d.Sitting)
	}
}

// A NEW SITTING STARTS AT ZERO, including the second sitting of the same agent — a warm session
// resumed and registered again. The count belongs to the sitting, not to the agent.
func TestTheNextSittingOfTheSameAgentStartsAtZero(t *testing.T) {
	run := runWithLimit(t, 2)
	mustOpen(t, run, "agent_01", "blue-respond", 1)
	for i := 0; i < 3; i++ {
		count(t, run, "agent_01")
	}
	mustOpen(t, run, "agent_01", "blue-respond", 2)
	d := count(t, run, "agent_01")
	if d.Count != 1 || d.Over || d.Sitting != 2 {
		t.Errorf("first call of sitting 2: %+v, want count 1, sitting 2, not over", d)
	}
}

// THE LIMIT IS RECORDED ONCE PER SITTING, whatever number of calls cross it.
func TestExactlyOneCallPerSittingIsFirst(t *testing.T) {
	run := runWithLimit(t, 1)
	mustOpen(t, run, "agent_01", "red-chair", 1)
	count(t, run, "agent_01")
	var wg sync.WaitGroup
	var mu sync.Mutex
	firsts := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d, _, err := Count(run, "agent_01")
			if err != nil {
				t.Error(err)
				return
			}
			if !d.Over {
				t.Error("a call past the limit was not over")
			}
			if d.First {
				mu.Lock()
				firsts++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firsts != 1 {
		t.Errorf("%d calls claimed to be the first past the limit, want exactly 1", firsts)
	}
	// and the next sitting gets its own first
	mustOpen(t, run, "agent_01", "red-chair", 2)
	count(t, run, "agent_01")
	if d := count(t, run, "agent_01"); !d.First {
		t.Error("sitting 2's first call past the limit was not marked first")
	}
}

// CONCURRENT SEATS KEEP SEPARATE COUNTS, and no increment is lost to a race.
func TestConcurrentSeatsKeepSeparateCounts(t *testing.T) {
	run := runWithLimit(t, 1000)
	agents := []string{"agent_a", "agent_b", "agent_c"}
	for i, a := range agents {
		mustOpen(t, run, a, "red-lens-logic-L"+strconv.Itoa(i+1), 1)
	}
	var wg sync.WaitGroup
	for n, a := range agents {
		for i := 0; i < (n+1)*10; i++ {
			wg.Add(1)
			go func(a string) {
				defer wg.Done()
				if _, _, err := Count(run, a); err != nil {
					t.Error(err)
				}
			}(a)
		}
	}
	wg.Wait()
	for n, a := range agents {
		if d := count(t, run, a); d.Count != (n+1)*10+1 {
			t.Errorf("%s: count %d, want %d — calls leaked between seats or were lost", a, d.Count, (n+1)*10+1)
		}
	}
}

// NO REGISTER, NO LIMIT. The main session, an operator and an agent that never sat as a seat have
// no header, and none of them is counted.
func TestAnAgentWithNoSittingIsNotCounted(t *testing.T) {
	run := runWithLimit(t, 1)
	for _, agent := range []string{"agent_never_registered", ""} {
		if _, ok, err := Count(run, agent); ok || err != nil {
			t.Errorf("Count(%q) = ok %v err %v, want not counted", agent, ok, err)
		}
	}
	if _, ok, _ := Count("", "agent_01"); ok {
		t.Error("a call with no run directory was counted")
	}
}

func TestTheLimitIsTheRunTerm(t *testing.T) {
	if n, err := Limit(t.TempDir()); err != nil || n != DefaultMaxCalls {
		t.Errorf("no run-config: %d %v, want the default %d", n, err, DefaultMaxCalls)
	}
	if n, err := Limit(runWithLimit(t, 7)); err != nil || n != 7 {
		t.Errorf("limit 7 in run-config read as %d (%v)", n, err)
	}
	if _, err := Limit(runWithLimit(t, 0)); err == nil {
		t.Error("a limit of 0 on disk was accepted; setup refuses it, so one on disk was edited by hand")
	}
}

// THE AGENT ID IS A FILE NAME HERE, AND IT COMES OFF THE WIRE.
func TestAnAgentIDThatWouldEscapeTheDirectoryIsRefused(t *testing.T) {
	run := runWithLimit(t, 1)
	for _, bad := range []string{"../x", `a\b`, "..", "a/b", "a\nb"} {
		if err := Open(run, bad, Header{SeatID: "red-chair", Sitting: 1}); err == nil {
			t.Errorf("Open accepted agent id %q", bad)
		}
	}
}
