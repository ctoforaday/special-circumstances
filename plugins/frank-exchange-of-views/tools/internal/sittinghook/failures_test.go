package sittinghook

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// NEITHER SITTING EVENT DISPLAYS ANYTHING TO A HUMAN, so every failure below goes to the record and
// is read out on the next displaying event. Each used to be a silent return or a writer whose output
// went to /dev/null.
//
// WHAT THE TESTS HERE ASSERT IS SILENCE ON THE FAILURE PATH, not silence as a law of the events.
// SubagentStop may never speak — an emission re-invokes the seat, nine firings measured. SubagentStart
// may, and does (#1122); what it must never do is turn a failure it could not record into output.

func sittingStages(t *testing.T) map[string]bool {
	t.Helper()
	path, err := hookfailures.Path("frank-exchange-of-views")
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return out
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Failures {
		out[string(f.Stage)+"@"+f.Scope] = true
	}
	return out
}

func settle(t *testing.T, event string, f func(rec *hookfailures.Recorder)) {
	t.Helper()
	rec := hookfailures.New("frank-exchange-of-views", "feov-"+strings.ToLower(event), event, time.Now(), io.Discard)
	f(rec)
	if msg := rec.Settle(); msg != "" {
		t.Fatalf("%s displays nothing, and Settle returned a message: %q", event, msg)
	}
}

func TestAnUnparsablePayloadIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	settle(t, "SubagentStop", func(rec *hookfailures.Recorder) {
		var out strings.Builder
		if err := Stop(strings.NewReader("{ not json"), &out, rec); err != nil || out.Len() != 0 {
			t.Fatalf("err %v, stdout %q — the hook must stay silent and never fail", err, out.String())
		}
	})
	if !sittingStages(t)[string(StageInput)+"@"] {
		t.Fatalf("an unparsable payload was not recorded: %v", sittingStages(t))
	}
}

// A MISSING WRITER DURING A LIVE RUN: every sitting of the run lost, and the span ends said nothing
// where Limit, on the same missing writer, always did.
func TestAMissingWriterIsRecordedAndClearsWhenItArrives(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, _ := liveRun(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(self), writerFileName())); err == nil {
		t.Skip("a writer is already beside the test binary; the missing case cannot be staged")
	}
	settle(t, "SubagentStart", func(rec *hookfailures.Recorder) {
		_ = Start(payload(t, "a1", "frank-exchange-of-views:red-auditor", cwd), io.Discard, rec)
	})
	if !sittingStages(t)[string(StageWriterMissing)+"@"] {
		t.Fatalf("a missing writer was not recorded: %v", sittingStages(t))
	}

	capture(t) // places a writer beside the test binary, and stubs the spawn
	settle(t, "SubagentStart", func(rec *hookfailures.Recorder) {
		_ = Start(payload(t, "a1", "frank-exchange-of-views:red-auditor", cwd), io.Discard, rec)
	})
	if sittingStages(t)[string(StageWriterMissing)+"@"] {
		t.Errorf("the writer arrived and the entry stayed: %v", sittingStages(t))
	}
}

// THE WRITER RAN AND FAILED. Its output used to go to /dev/null — exec.Cmd.Stderr left nil — so what
// it said about why was destroyed. It is on the record now, in the writer's own words.
func TestAWriterThatFailsIsRecordedInItsOwnWords(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, _ := liveRun(t)
	capture(t)
	prev := spawn
	spawn = func(string, string, string, string, string, string) ([]byte, error) {
		return nil, errors.New("feov-sitting-write: exit status 1: sittingwrite: ingesting 3 turns for a1: disk full")
	}
	t.Cleanup(func() { spawn = prev })

	settle(t, "SubagentStop", func(rec *hookfailures.Recorder) {
		_ = Stop(payload(t, "a1", "frank-exchange-of-views:red-auditor", cwd), io.Discard, rec)
	})
	path, _ := hookfailures.Path("frank-exchange-of-views")
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "disk full") || !strings.Contains(string(b), string(StageWrite)) {
		t.Fatalf("the writer's failure was not recorded in its own words: %s", b)
	}
}

// THE REAL spawn keeps the child's output. A writer that prints and exits non-zero must have what it
// printed in the error — that is the whole difference from Run(), which sent it to /dev/null.
func TestTheRealSpawnKeepsTheWritersOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in writer is a shell script; the recording path is covered on every platform by the stubbed-spawn test above")
	}
	dir := t.TempDir()
	writer := filepath.Join(dir, "writer")
	if err := os.WriteFile(writer, []byte("#!/bin/sh\necho 'the writer said why' >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := spawn(writer, dir, phaseClose, "a1", "t", "")
	if err == nil || !strings.Contains(err.Error(), "the writer said why") {
		t.Fatalf("spawn lost the writer's output: %v", err)
	}
}

// AN UNUSABLE MARKER IS A FAULT, SCOPED TO ITS PROJECT; the ordinary shapes are not. Two runs open at
// once is legitimate — recording it would raise a false alarm on every sitting of both runs.
func TestOnlyAFaultyMarkerIsRecorded(t *testing.T) {
	for name, tc := range map[string]struct {
		marker string
		fault  bool
	}{
		"unreadable":    {"{ not json", true},
		"stale":         {`{"runs":[{"runDir":"research/gone"}]}`, true},
		"no run open":   {`{"runs":[]}`, false},
		"two runs open": {`{"runs":[{"runDir":"a"},{"runDir":"b"}]}`, false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			proj := t.TempDir()
			if err := os.MkdirAll(filepath.Join(proj, ".claude"), 0o755); err != nil {
				t.Fatal(err)
			}
			for _, d := range []string{"a", "b"} {
				if err := os.MkdirAll(filepath.Join(proj, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"), []byte(tc.marker), 0o644); err != nil {
				t.Fatal(err)
			}
			capture(t)
			settle(t, "SubagentStart", func(rec *hookfailures.Recorder) {
				_ = Start(payload(t, "a1", "frank-exchange-of-views:red-auditor", proj), io.Discard, rec)
			})
			got := sittingStages(t)[string(runlive.StageUnusable)+"@"+proj]
			if got != tc.fault {
				t.Errorf("recorded = %v, want %v: %v", got, tc.fault, sittingStages(t))
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("stdin closed mid-read") }

// A PAYLOAD THAT CANNOT EVEN BE READ.
func TestAPayloadThatCannotBeReadIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	settle(t, "SubagentStart", func(rec *hookfailures.Recorder) {
		_ = Start(failingReader{}, io.Discard, rec)
	})
	if !sittingStages(t)[string(StageInput)+"@"] {
		t.Fatalf("an unreadable payload was not recorded: %v", sittingStages(t))
	}
}

// A SITTING THAT IS WRITTEN CLEARS WHAT EARLIER SITTINGS RECORDED — the payload, the marker, the
// write. Seeded, then driven clean, and asserted against the record file.
func TestAWrittenSittingClearsTheEntries(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	cwd, _ := liveRun(t)
	seed := hookfailures.New("frank-exchange-of-views", "feov-subagentstop", "SubagentStop", time.Now(), io.Discard)
	seed.Fail(StageInput, "seeded")
	seed.FailIn(runlive.StageUnusable, cwd, "seeded")
	seed.FailIn(StageWrite, cwd, "seeded")
	seed.Settle()

	capture(t)
	settle(t, "SubagentStop", func(rec *hookfailures.Recorder) {
		_ = Stop(payload(t, "a1", "frank-exchange-of-views:red-auditor", cwd), io.Discard, rec)
	})
	got := sittingStages(t)
	for _, k := range []string{string(StageInput) + "@", string(runlive.StageUnusable) + "@" + cwd, string(StageWrite) + "@" + cwd} {
		if got[k] {
			t.Errorf("a written sitting left %s on the record", k)
		}
	}
}
