package runlive

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A RUN'S MEMBERS ARE ITS LEAD SESSION'S AGENTS. The marker's location says which checkout a run is
// in; a worktree nested under that checkout finds the same marker from another conversation, and
// the session is what tells the two apart.
func TestOnlyTheLeadSessionsCallersBelongToTheRun(t *testing.T) {
	proj := t.TempDir()
	run := filepath.Join(proj, "research", "r")
	mustMkdir(t, run)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"research/r","leadSession":"session-lead-under-test"}]}`)
	nested := filepath.Join(proj, ".claude", "worktrees", "dev", "sub")
	mustMkdir(t, nested)

	for _, tc := range []struct {
		name   string
		caller Caller
		why    Inference
	}{
		{"a seat of the lead session", Caller{ID: "a-seat", Type: SeatTypePrefix + "red-chair", Session: testLead}, Resolved},
		{"the lead session itself", Caller{Session: testLead}, Resolved},
		{"another session's Explore agent", Caller{ID: "a-dev", Type: "Explore", Session: "dev-session"}, OtherSession},
		{"another session's main thread", Caller{Session: "dev-session"}, OtherSession},
		{"one of this plugin's agents from another session", Caller{ID: "a-x", Type: SeatTypePrefix + "red-chair", Session: "dev-session"}, SeatOfAnotherSession},
		{"a payload with no session", Caller{ID: "a-seat", Type: SeatTypePrefix + "red-chair"}, Sessionless},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := InferRunDir(nested, tc.caller)
			if got.Why != tc.why {
				t.Fatalf("Why = %v, want %v", got.Why, tc.why)
			}
			if (got.Dir != "") != (tc.why == Resolved) {
				t.Fatalf("Dir = %q for %v", got.Dir, got.Why)
			}
			if got.Why.Membership() != (tc.why == Sessionless || tc.why == SeatOfAnotherSession) {
				t.Fatalf("Membership() = %v for %v", got.Why.Membership(), got.Why)
			}
		})
	}
}

// TWO RUNS LED BY TWO SESSIONS ARE NOT AMBIGUOUS: each payload's session picks its own.
func TestTwoRunsLedByTwoSessionsResolveApart(t *testing.T) {
	proj := t.TempDir()
	mustMkdir(t, filepath.Join(proj, "a"))
	mustMkdir(t, filepath.Join(proj, "b"))
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"a","leadSession":"sa"},{"runDir":"b","leadSession":"sb"}]}`)
	for session, want := range map[string]string{"sa": filepath.Join(proj, "a"), "sb": filepath.Join(proj, "b")} {
		if got := InferRunDir(proj, Caller{ID: "x", Session: session}); got.Why != Resolved || got.Dir != want {
			t.Errorf("session %s: %v %q, want %q", session, got.Why, got.Dir, want)
		}
	}
}

// A MEMBERSHIP FAULT NAMES THE TAKE-OVER, because the one legitimate shape of a plugin agent from a
// session that leads nothing is a resume in a new session.
func TestASeatOfAnotherSessionNamesTheTakeOver(t *testing.T) {
	proj := t.TempDir()
	mustMkdir(t, filepath.Join(proj, "r"))
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"r","leadSession":"old-lead"}]}`)
	got := InferRunDir(proj, Caller{ID: "a", Type: BenchAgentType, Session: "new-lead"})
	detail := FaultDetail(got)
	for _, want := range []string{"old-lead", "new-lead", "take-over", "--from-session"} {
		if !strings.Contains(detail, want) {
			t.Errorf("the fault must carry %q: %s", want, detail)
		}
	}
}

// THE TAKE-OVER IS REFUSABLE: a session takes a run only from the lead it names.
func TestTakeOverMovesTheLeadOnlyFromTheSessionNamed(t *testing.T) {
	proj := t.TempDir()
	mustMkdir(t, filepath.Join(proj, "research", "r"))
	if _, err := WriteRunLiveMarker(proj, "research/r", nil, time.Now(), "", "", "old-lead"); err != nil {
		t.Fatal(err)
	}
	var led *LedElsewhere
	if _, err := TakeOver(proj, "research/r", "someone-else", "thief"); !errors.As(err, &led) || led.Lead != "old-lead" {
		t.Fatalf("a take-over naming the wrong lead: %v, want LedElsewhere{old-lead}", err)
	}
	if m, _ := RowOf(proj, "research/r"); m.LeadSession != "old-lead" {
		t.Fatalf("a refused take-over moved the lead to %q", m.LeadSession)
	}
	if _, err := TakeOver(proj, "research/r", "old-lead", ""); err == nil {
		t.Fatal("a take-over with no session to lead was accepted")
	}
	prior, err := TakeOver(proj, "research/r", "old-lead", "new-lead")
	if err != nil || prior != "old-lead" {
		t.Fatalf("TakeOver = %q, %v", prior, err)
	}
	if m, _ := RowOf(proj, "research/r"); m.LeadSession != "new-lead" {
		t.Fatalf("lead after the take-over: %q, want new-lead", m.LeadSession)
	}
	if got := InferRunDir(proj, Caller{ID: "a", Session: "new-lead"}); got.Why != Resolved {
		t.Fatalf("the new lead's seat after the take-over: %v", got.Why)
	}
	if got := InferRunDir(proj, Caller{ID: "a", Session: "old-lead"}); got.Why != OtherSession {
		t.Fatalf("the old lead's seat after the take-over: %v", got.Why)
	}
	// The lead taking over its own run changes nothing.
	if prior, err := TakeOver(proj, "research/r", "whatever", "new-lead"); err != nil || prior != "new-lead" {
		t.Fatalf("a take-over by the lead itself: %q, %v", prior, err)
	}
	if _, err := TakeOver(proj, "research/none", "", "new-lead"); err == nil {
		t.Fatal("a take-over of a run with no row was accepted")
	}
}

// A RE-SETUP IS ITS LEAD'S OWN: another session's write cannot move the lead of a live run.
func TestTheMarkerRefusesAnotherSessionsRewriteOfALiveRun(t *testing.T) {
	proj := t.TempDir()
	mustMkdir(t, filepath.Join(proj, "research", "r"))
	if _, err := WriteRunLiveMarker(proj, "research/r", nil, time.Now(), "", "", "lead"); err != nil {
		t.Fatal(err)
	}
	var led *LedElsewhere
	if _, err := WriteRunLiveMarker(proj, "research/r", nil, time.Now(), "", "", "other"); !errors.As(err, &led) {
		t.Fatalf("another session's re-setup: %v, want LedElsewhere", err)
	}
	if _, err := WriteRunLiveMarker(proj, "research/r", nil, time.Now(), "", "", "lead"); err != nil {
		t.Fatalf("the lead's own re-setup: %v", err)
	}
	if _, err := WriteRunLiveMarker(proj, "research/s", nil, time.Now(), "", "", " "); err == nil {
		t.Fatal("a row with no lead session was written")
	}
}
