package sittingwrite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func TestBothEndsOfASpanLand(t *testing.T) {
	run := newRun(t)
	for _, p := range []Phase{Open, Close} {
		if err := Write(run, p, "agent_01", "frank-exchange-of-views:lead-judge"); err != nil {
			t.Fatalf("writing the %s end: %v", p, err)
		}
	}
	open, closed := countByType(t, run)
	if open != 1 || closed != 1 {
		t.Errorf("want one open and one close, got %d and %d", open, closed)
	}
}

// THE ENVELOPE NAMES THE HOOK, NOT A SEAT. SubagentStart cannot know which seat it just started
// (#290), so the seat is recovered later by joining agent_id to the register event. An invented
// seat-shaped id here would be a guess written into a permanent record.
func TestTheSpanIsAttributedToTheHookAndCarriesTheAgent(t *testing.T) {
	run := newRun(t)
	if err := Write(run, Open, "agent_42", "frank-exchange-of-views:red-auditor"); err != nil {
		t.Fatal(err)
	}
	m, err := record.MergedEvents(mustRun(t, run))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_SITTING_OPEN {
			continue
		}
		if e.GetSeatId() != HookSeat {
			t.Errorf("seat_id = %q, want %q — the hook cannot name a seat and must not invent one", e.GetSeatId(), HookSeat)
		}
		if got := e.GetSittingOpen().GetAgentId(); got != "agent_42" {
			t.Errorf("agent_id = %q, want agent_42 — this is the join key the seat is recovered by", got)
		}
		if e.GetRound() != -1 {
			t.Errorf("round = %d, want -1 (UNKNOWN) — a hook fires outside any seat's round, and 0 is a real round", e.GetRound())
		}
		return
	}
	t.Fatal("no sitting-open event on the record")
}

// A SPAN WITH NO IDENTITY IS REFUSED RATHER THAN WRITTEN EMPTY. The hook is what decides an event
// is a seat; if that decision is ever bypassed, this says so loudly instead of putting a row on a
// permanent record that joins to nothing.
func TestASpanWithNoIdentityIsRefused(t *testing.T) {
	run := newRun(t)
	err := Write(run, Open, "", "")
	if err == nil {
		t.Fatal("a sitting with no agent identity was written")
	}
	if !strings.Contains(err.Error(), "agent identity") {
		t.Errorf("refusal does not name the cause: %v", err)
	}
}

func TestAnUnknownPhaseIsRefused(t *testing.T) {
	if err := Write(newRun(t), Phase("middle"), "a", "b"); err == nil {
		t.Fatal("a third end of a two-ended span was accepted")
	}
}

func newRun(t *testing.T) string {
	t.Helper()
	// recordtest.TmpRun is t.TempDir PLUS the cached-handle release, required of any test that
	// opens a record — without it the handle outlives the directory, which passes on Linux and
	// fails the Windows leg (#666).
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.CloseUnder(dir) })
	return dir
}

func mustRun(t *testing.T, dir string) record.Run {
	t.Helper()
	r, err := record.NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func countByType(t *testing.T, dir string) (open, closed int) {
	t.Helper()
	m, err := record.MergedEvents(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_SITTING_OPEN:
			open++
		case recordpb.EventType_EVENT_TYPE_SITTING_CLOSE:
			closed++
		}
	}
	return open, closed
}
