package sittinghook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nonet"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
)

type handoffArgs struct {
	writer, runDir, phase, agentID, agentType, transcript string
	sessionID, promptID                                   string
}

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
	spawn = func(w, r, p, id, ty, tr, sid, pid string) ([]byte, error) {
		got = append(got, handoffArgs{w, r, p, id, ty, tr, sid, pid})
		return nil, nil
	}
	t.Cleanup(func() { spawn = prev })
	return &got
}

// payloadWith is payload plus the transcript path SubagentStop carries.
func payloadWith(t *testing.T, agentID, agentType, cwd, transcript string) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(map[string]string{
		"agent_id": agentID, "agent_type": agentType, "cwd": cwd, "agent_transcript_path": transcript,
	})
	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(string(b))
}

func payload(t *testing.T, agentID, agentType, cwd string) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(map[string]string{"agent_id": agentID, "agent_type": agentType, "cwd": cwd})
	if err != nil {
		t.Fatal(err)
	}
	return strings.NewReader(string(b))
}

// SubagentStop's SILENCE IS THE CONTRACT, AND IT IS ITS ALONE. A Stop hook that emits anything
// re-invokes the seat and fires again — nine firings for one seat in the measured case, the returned
// context discarded every time (plans/hook-surface-spike.md §10). This keeps a well-meaning "tell the
// seat what we recorded" from turning one event into nine.
//
// THE GATE USED TO COVER BOTH ENDS, WHICH STATED A MEASUREMENT OF ONE EVENT AS A LAW OF TWO. §10
// measured the opposite on Start in the same run: one firing, the marker delivered to the SEAT's
// context. Start's own contract is asserted by TestStartSpeaksTheSeatsWorkListAndNothingElse.
func TestSubagentStopEmitsNothing(t *testing.T) {
	got := capture(t)
	cwd, _ := liveRun(t)
	// THE WRITER HANDS BACK A PAYLOAD, which is the only version of this test that holds anything: a
	// stub returning nothing passes whatever Stop does with it.
	prev := spawn
	spawn = func(w, r, p, id, ty, tr, sid, pid string) ([]byte, error) {
		*got = append(*got, handoffArgs{w, r, p, id, ty, tr, sid, pid})
		return []byte(`{"sitting":{"seat":"red-lens-evidence"}}`), nil
	}
	t.Cleanup(func() { spawn = prev })
	var out bytes.Buffer
	if err := Stop(payload(t, "agent_01", "frank-exchange-of-views:red-auditor", cwd), &out, testRecorder()); err != nil {
		t.Fatalf("Stop returned an error; a hook must not fail on an event the seat cannot see: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("Stop wrote %q to stdout — an emission re-invokes the seat and the event fires nine times",
			out.String())
	}
	if len(*got) != 1 {
		t.Errorf("want the closing end handed to the writer, got %d", len(*got))
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
	if err := Stop(payload(t, "minted_99", "", cwd), &out, testRecorder()); err != nil {
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
	if err := Start(payload(t, "agent_01", "frank-exchange-of-views:lead-judge", t.TempDir()), &out, testRecorder()); err != nil {
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
	if err := Start(payload(t, "agent_07", "frank-exchange-of-views:blue-researcher", cwd), &out, testRecorder()); err != nil {
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
	if phaseLimit != string(sittingwrite.Limit) {
		t.Errorf("the hook sends phase %q and the writer parses %q — no sitting limit would reach the record", phaseLimit, sittingwrite.Limit)
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

// THE TRANSCRIPT PATH RIDES THE SPAWN, AND ONLY WHEN THERE IS ONE.
//
// It is what makes the turns readable during the run instead of at capture. The hook does not
// open it — that is the writer's job, in the process that already carries the record — but if the
// hook drops it the writer has nothing to ingest and the whole live path is dead.
func TestStopCarriesTheTranscriptPathAndStartDoesNot(t *testing.T) {
	got := capture(t)
	dir, _ := liveRun(t)

	body := payloadWith(t, "a1", "frank-exchange-of-views:red-auditor", dir, "/tmp/agent-a1.jsonl")
	if err := Stop(body, &bytes.Buffer{}, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("Stop spawned %d writers, want 1", len(*got))
	}
	if (*got)[0].transcript != "/tmp/agent-a1.jsonl" {
		t.Errorf("the transcript path did not reach the writer: %q", (*got)[0].transcript)
	}

	// SubagentStart carries no transcript — a seat just dispatched has produced no turns.
	*got = nil
	if err := Start(payload(t, "a1", "frank-exchange-of-views:red-auditor", dir), &bytes.Buffer{}, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("Start spawned %d writers, want 1", len(*got))
	}
	if (*got)[0].transcript != "" {
		t.Errorf("Start sent a transcript path %q; there are no turns at the opening end", (*got)[0].transcript)
	}
}

// testRecorder records into the package's isolated state directory (TestMain) and writes its lines
// nowhere.
func testRecorder() *hookfailures.Recorder {
	return hookfailures.New("frank-exchange-of-views", "test", "PreToolUse", time.Now(), io.Discard)
}

// No test in this package may write the developer's own failure record: every entry point here now
// settles into ~/.local/state unless something stops it.
//
// The join test opens a record, so the process must not exit while a cached database handle
// outlives the directory it lived in — invisible on Linux, a Windows-leg failure. This package
// needs its own TestMain for the state sandbox, so it keeps recordtest.Main's two guards around
// m.Run rather than giving them up: loopback-only networking before, the orphaned-handle check
// after (#666).
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "feov-hookfailures-test-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "TestMain:", err)
		os.Exit(1)
	}
	for _, k := range []string{"HOME", "USERPROFILE", "XDG_STATE_HOME"} {
		if err := os.Setenv(k, dir); err != nil {
			fmt.Fprintln(os.Stderr, "TestMain:", err)
			os.Exit(1)
		}
	}
	nonet.OnlyLoopback()
	code := m.Run()
	if err := recordtest.CheckOrphanedHandles(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			code = 1
		}
	}
	os.RemoveAll(dir)
	os.Exit(code)
}

// START SPEAKS THE SEAT'S WORK LIST, AND SPEAKS NOTHING ELSE (#1122).
//
// This is the contract the narrowed gate above leaves to be stated. Three properties, and each one
// is a way the channel has a plausible failure that looks like success:
//
//   - the writer's stdout reaches the seat VERBATIM, through a JSON string, so a list carrying a
//     quoted claim survives — every gap location quotes the sentence it challenges;
//   - nothing is emitted for an empty payload, because blue's configuration seats three seats and an
//     empty additionalContext is a document that says nothing;
//   - the writer's STDERR never reaches the seat. The streams were merged, and a diagnostic line
//     merged into the other would arrive in a seat's context as its work.
func TestStartSpeaksTheSeatsWorkListAndNothingElse(t *testing.T) {
	cwd, _ := liveRun(t)
	capture(t) // places a writer beside the test binary; writerPath's absence check gates the handoff
	for _, tc := range []struct {
		name, stdout string
		want         string
	}{
		{"a list", `{"sitting":{"seat":"red-lens-evidence","open":[{"what":"quote \"this\""}]}}`,
			`{"sitting":{"seat":"red-lens-evidence","open":[{"what":"quote \"this\""}]}}`},
		{"nothing", "", ""},
		{"whitespace only", "\n  \n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := spawn
			spawn = func(string, string, string, string, string, string, string, string) ([]byte, error) {
				return []byte(tc.stdout), nil
			}
			t.Cleanup(func() { spawn = prev })

			var out bytes.Buffer
			if err := Start(payload(t, "agent_01", "frank-exchange-of-views:red-lens-evidence", cwd), &out, testRecorder()); err != nil {
				t.Fatal(err)
			}
			if tc.want == "" {
				if out.Len() != 0 {
					t.Fatalf("an empty payload produced a document anyway: %q", out.String())
				}
				return
			}
			var doc struct {
				HookSpecificOutput struct {
					HookEventName     string `json:"hookEventName"`
					AdditionalContext string `json:"additionalContext"`
				} `json:"hookSpecificOutput"`
			}
			if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
				t.Fatalf("what Start emitted is not a hook document (%v): %q", err, out.String())
			}
			if doc.HookSpecificOutput.HookEventName != "SubagentStart" {
				t.Errorf("the document names %q; the client routes on this field",
					doc.HookSpecificOutput.HookEventName)
			}
			if got := doc.HookSpecificOutput.AdditionalContext; got != tc.want {
				t.Errorf("the seat's list did not survive the envelope:\n got %q\nwant %q", got, tc.want)
			}
		})
	}
}

// THE WRITER'S DIAGNOSTICS ARE NOT THE SEAT'S WORK. exec.Cmd's two streams were merged by
// CombinedOutput, which was right while the output was only ever a failure report. This drives the
// REAL spawn against a writer that prints on both, because the split is the thing under test and a
// stubbed spawn cannot have a stderr to lose.
func TestOnlyTheWritersStdoutReachesTheSeat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in writer is a shell script")
	}
	dir := t.TempDir()
	writer := filepath.Join(dir, "writer")
	if err := os.WriteFile(writer, []byte("#!/bin/sh\necho 'for the seat'\necho 'a diagnostic' >&2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	forSeat, err := spawn(writer, dir, phaseOpen, "a1", "t", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(forSeat))
	if got != "for the seat" {
		t.Errorf("the seat's channel carried %q — a diagnostic that reaches it arrives as the seat's work", got)
	}
}

// THE CONVERSATION FIELDS SURVIVE THE PAYLOAD AND REACH THE WRITER (#1122 follow-up).
//
// Both are on every SubagentStart and SubagentStop payload — measured against a live one — and were
// parsed by nothing. This drives the real parse rather than a constructed struct, because the defect
// it guards against is a json tag that does not match the wire name, which a struct literal hides.
func TestTheHandoffCarriesTheConversationAndTheDispatch(t *testing.T) {
	got := capture(t)
	cwd, _ := liveRun(t)
	raw, err := json.Marshal(map[string]string{
		"agent_id": "a5280059b8b60e1f7", "agent_type": "frank-exchange-of-views:red-lens-voice",
		"cwd": cwd, "session_id": "ab802afa-997d", "prompt_id": "e353381b-5bfa",
	})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := Start(strings.NewReader(string(raw)), &out, testRecorder()); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Fatalf("want one handoff, got %d", len(*got))
	}
	if a := (*got)[0]; a.sessionID != "ab802afa-997d" || a.promptID != "e353381b-5bfa" {
		t.Errorf("the handoff dropped the conversation: session=%q prompt=%q", a.sessionID, a.promptID)
	}
}
