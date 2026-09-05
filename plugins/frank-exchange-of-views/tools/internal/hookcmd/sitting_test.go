package hookcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

func payload(t *testing.T, agentID, agentType, cwd string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"agent_id": agentID, "agent_type": agentType, "cwd": cwd})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// THE SILENCE IS THE CONTRACT. A SubagentStop hook that emits anything re-invokes the seat and
// fires again — nine firings for one seat in the measured case, the returned context discarded
// every time (plans/hook-surface-spike.md §10). This is the assertion that keeps a well-meaning
// "just tell the seat what we recorded" from turning one event into nine.
func TestTheSittingHooksEmitNothing(t *testing.T) {
	cwd, run := newRunDir(t)
	_ = run
	for _, tc := range []struct {
		name string
		fn   func(r *strings.Reader, w *bytes.Buffer) error
	}{
		{"SubagentStart", func(r *strings.Reader, w *bytes.Buffer) error { return SubagentStart(r, w) }},
		{"SubagentStop", func(r *strings.Reader, w *bytes.Buffer) error { return SubagentStop(r, w) }},
	} {
		var out bytes.Buffer
		in := strings.NewReader(payload(t, "agent_01", "frank-exchange-of-views:red-auditor", cwd))
		if err := tc.fn(in, &out); err != nil {
			t.Fatalf("%s returned an error; a hook must not fail on an event the seat cannot see: %v", tc.name, err)
		}
		if out.Len() != 0 {
			t.Errorf("%s wrote %q to stdout — an emission re-invokes the seat and the event fires nine times",
				tc.name, out.String())
		}
	}
}

// A MAIN-AGENT TURN END IS NOT A SITTING. SubagentStop fires at both, and agent_type is the only
// thing that tells them apart — 19 seats against 50 turn ends in one measured session. Without
// this filter a run's sitting count reads about 3.6x its seat count.
func TestATurnEndIsNotRecordedAsASitting(t *testing.T) {
	cwd, run := newRunDir(t)
	var out bytes.Buffer
	// A turn end: a minted id, and NO agent_type.
	if err := SubagentStop(strings.NewReader(payload(t, "minted_99", "", cwd)), &out); err != nil {
		t.Fatal(err)
	}
	if n := countSittings(t, run); n != 0 {
		t.Errorf("a turn end with no agent_type was recorded as %d sitting event(s)", n)
	}
}

func TestASeatsSpanIsRecordedAtBothEnds(t *testing.T) {
	cwd, run := newRunDir(t)
	var out bytes.Buffer
	p := payload(t, "agent_01", "frank-exchange-of-views:lead-judge", cwd)
	if err := SubagentStart(strings.NewReader(p), &out); err != nil {
		t.Fatal(err)
	}
	if err := SubagentStop(strings.NewReader(p), &out); err != nil {
		t.Fatal(err)
	}
	if n := countSittings(t, run); n != 2 {
		t.Fatalf("want 2 sitting events (open and close), got %d", n)
	}
}

// NO RUN, NO EVENT, NO ERROR. A subagent outside any run is the ordinary case in a normal
// session, and a hook that failed there would fail on every subagent the user ever launches.
func TestASubagentOutsideARunRecordsNothingAndDoesNotFail(t *testing.T) {
	var out bytes.Buffer
	if err := SubagentStart(strings.NewReader(payload(t, "agent_01", "x", t.TempDir())), &out); err != nil {
		t.Errorf("a subagent outside a run failed the hook: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("wrote %q", out.String())
	}
}

// newRunDir builds a project whose run-live marker points at a run, so InferRunDir resolves it
// through the package that owns the marker rather than a private decode of its shape — the
// distinction seat.go's own comment exists to make.
//
// It returns the CWD, not the run directory, because cwd is what the hook payload carries.
func newRunDir(t *testing.T) (cwd, runDir string) {
	t.Helper()
	// recordtest.TmpRun rather than a bare t.TempDir: it is t.TempDir PLUS the handle release,
	// and the hooks below open a record under it. Without the release the cached database handle
	// outlives the directory — which SUCCEEDS on Linux and fails the Windows leg (#666). The
	// guard sweep caught exactly this, in these tests, the first time they ran.
	cwd = recordtest.TmpRun(t)
	runDir = filepath.Join(cwd, "run")
	if err := os.MkdirAll(filepath.Join(runDir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	// The record lives under runDir, not under cwd, so the release has to name it.
	t.Cleanup(func() { _ = recordsql.CloseUnder(runDir) })
	runlive.WriteRunLiveMarker(cwd, runDir, nil, time.Now(), "run_test", "")
	return cwd, runDir
}

func countSittings(t *testing.T, runDir string) int {
	t.Helper()
	run, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	m, err := record.MergedEvents(run)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range m.Events {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_SITTING_OPEN, recordpb.EventType_EVENT_TYPE_SITTING_CLOSE:
			n++
		}
	}
	return n
}
