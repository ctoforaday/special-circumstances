package sittinghook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

type handoffArgs struct{ writer, runDir, phase, agentID, agentType string }

// capture swaps the spawn seam and records what the hook decided to hand over. It also puts a
// writer on disk beside the test binary, because writerPath's absence check is part of what is
// under test everywhere else.
func capture(t *testing.T) *[]handoffArgs {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	stub := filepath.Join(filepath.Dir(self), writerFileName())
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Skipf("cannot place a writer beside the test binary: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(stub) })

	var got []handoffArgs
	prev := spawn
	spawn = func(w, r, p, id, ty string) { got = append(got, handoffArgs{w, r, p, id, ty}) }
	t.Cleanup(func() { spawn = prev })
	return &got
}

func payload(t *testing.T, agentID, agentType, cwd string) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(map[string]string{"agent_id": agentID, "agent_type": agentType, "cwd": cwd})
	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(string(b))
}

// THE SILENCE IS THE CONTRACT. A SubagentStop hook that emits anything re-invokes the seat and
// fires again — nine firings for one seat in the measured case, the returned context discarded
// every time (plans/hook-surface-spike.md §10). This keeps a well-meaning "tell the seat what we
// recorded" from turning one event into nine.
func TestTheSittingHooksEmitNothing(t *testing.T) {
	got := capture(t)
	cwd, _ := liveRun(t)
	for _, tc := range []struct {
		name string
		fn   func(*strings.Reader, *bytes.Buffer) error
	}{
		{"Start", func(r *strings.Reader, w *bytes.Buffer) error { return Start(r, w) }},
		{"Stop", func(r *strings.Reader, w *bytes.Buffer) error { return Stop(r, w) }},
	} {
		var out bytes.Buffer
		if err := tc.fn(payload(t, "agent_01", "frank-exchange-of-views:red-auditor", cwd), &out); err != nil {
			t.Fatalf("%s returned an error; a hook must not fail on an event the seat cannot see: %v", tc.name, err)
		}
		if out.Len() != 0 {
			t.Errorf("%s wrote %q to stdout — an emission re-invokes the seat and the event fires nine times",
				tc.name, out.String())
		}
	}
	if len(*got) != 2 {
		t.Errorf("want both ends handed to the writer, got %d", len(*got))
	}
}

// A MAIN-AGENT TURN END IS NOT A SITTING, AND MUST NOT EVEN SPAWN. SubagentStop fires at both —
// 19 seats against 50 turn ends in one measured session — and agent_type is the only thing that
// tells them apart. This filter is also the whole frequency argument for keeping the hook light:
// every event it rejects here is a writer process not started.
func TestATurnEndIsNeitherRecordedNorSpawnedFor(t *testing.T) {
	got := capture(t)
	cwd, _ := liveRun(t)
	var out bytes.Buffer
	if err := Stop(payload(t, "minted_99", "", cwd), &out); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 0 {
		t.Errorf("a turn end with no agent_type spawned the writer %d time(s)", len(*got))
	}
}

// NO RUN, NO SPAWN. A subagent outside any run is the ordinary case in a normal session — this is
// the path taken by every subagent every user launches, and it must cost nothing.
func TestASubagentOutsideARunDoesNotSpawnTheWriter(t *testing.T) {
	got := capture(t)
	var out bytes.Buffer
	if err := Start(payload(t, "agent_01", "frank-exchange-of-views:lead-judge", t.TempDir()), &out); err != nil {
		t.Errorf("a subagent outside a run failed the hook: %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("a subagent outside a run spawned the writer %d time(s)", len(*got))
	}
}

// THE HANDOFF CARRIES THE FACTS THE WRITER NEEDS, spelled the way it parses them.
func TestTheHandoffNamesTheRunPhaseAndIdentity(t *testing.T) {
	got := capture(t)
	cwd, runDir := liveRun(t)
	var out bytes.Buffer
	if err := Start(payload(t, "agent_07", "frank-exchange-of-views:blue-researcher", cwd), &out); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("want one handoff, got %d", len(*got))
	}
	a := (*got)[0]
	if a.runDir != runDir || a.phase != phaseOpen || a.agentID != "agent_07" ||
		a.agentType != "frank-exchange-of-views:blue-researcher" {
		t.Errorf("handoff = %+v, want run=%s phase=%s agent_07/blue-researcher", a, runDir, phaseOpen)
	}
}

// THE PHASE STRINGS THE HOOK PASSES ARE THE ONES THE WRITER PARSES. sitting.go repeats them
// rather than importing them, to keep sittingwrite's graph — a SQLite driver and every protobuf
// descriptor — out of a binary that fires at every main-agent turn end in every session. This is
// the guard that makes that duplication safe, and it lives in the TEST binary where the heavy
// import costs nothing.
func TestThePhaseStringsMatchTheWriters(t *testing.T) {
	if phaseOpen != string(sittingwrite.Open) {
		t.Errorf("the hook sends phase %q and the writer parses %q — every open end would be refused", phaseOpen, sittingwrite.Open)
	}
	if phaseClose != string(sittingwrite.Close) {
		t.Errorf("the hook sends phase %q and the writer parses %q — every close end would be refused", phaseClose, sittingwrite.Close)
	}
}

// liveRun builds a project whose run-live marker points at a run, so InferRunDir resolves it
// through the package that owns the marker. Returns the CWD, which is what the payload carries.
func liveRun(t *testing.T) (cwd, runDir string) {
	t.Helper()
	cwd = t.TempDir()
	runDir = filepath.Join(cwd, "run")
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "")
	return cwd, runDir
}
