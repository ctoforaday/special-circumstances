package record

import (
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE ONLY TEST THAT MATTERS HERE IS THE ONE FOR STALE, and it is the branch a healthy run never
// takes — which is exactly why it shipped absent. A liveness check that can only be exercised by
// a dead run gets written, never fires, and is believed.

const liveStamp = "2006-01-02T15:04:05.000000000Z07:00"

// runWithCadence writes one shard whose events are `gap` apart, ending at `last`.
func runWithCadence(t *testing.T, n int, gap time.Duration, last time.Time) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	evs := make([]*Event, 0, n)
	for i := 0; i < n; i++ {
		ts := last.Add(-time.Duration(n-1-i) * gap).UTC().Format(liveStamp)
		evs = append(evs, recordtest.Stamped(
			recordtest.At(t, "red-lens-evidence", "red-lens-evidence:finding:k"+itoaT(i), &recordpb.Finding{
				Id:   proto.String("L1-F" + itoaT(i)),
				Text: proto.String("a finding"),
			}), ts))
	}
	writeShard(t, dir, evs)
	return dir
}

func itoaT(i int) string { return string(rune('a' + i%26)) }

func TestLastActivityRefusesARecordWithNoEvents(t *testing.T) {
	dir := recordtest.TmpRun(t)
	m, err := MergedEvents(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lastActivity(mustRun(t, dir), m); err == nil {
		t.Fatal("lastActivity returned no error for a run that has recorded nothing — a run that " +
			"never started and a run that went quiet would then arrive as the same age")
	}
}

func TestLastActivityReadsTheNewestEventsOwnTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	dir := runWithCadence(t, 8, 30*time.Second, now)
	m, err := MergedEvents(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	a, err := lastActivity(mustRun(t, dir), m)
	if err != nil {
		t.Fatal(err)
	}
	if !a.At.Equal(now) {
		t.Errorf("newest event at %s, want %s", a.At, now)
	}
	if a.Seat != "red-lens-evidence" {
		t.Errorf("seat = %q", a.Seat)
	}
	if a.Events != 8 {
		t.Errorf("counted %d events, want 8", a.Events)
	}
}

// ENDED HAS ONE SOURCE: THE MARK CAPTURE WRITES. A run whose bench recorded its outcome is decided,
// not ended — its assembly sitting may still be at work and capture has not run — and the write
// path, which refuses on the mark alone, takes its writes. A liveness verdict fed "has a verdict"
// called that run ENDED while the record still took writes; the two readings must be one.
func TestAssessSaysEndedOnlyOnceCaptured(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	dir := runWithCadence(t, 10, 20*time.Second, now)
	run := mustRun(t, dir)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "judge"}, "", "assemble"); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: run, SeatID: "judge"}, &recordpb.Outcome{
		Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED), Prose: proto.String("nobody was ready"),
	}); err != nil {
		t.Fatal(err)
	}
	if v, err := TerminalVerdict(run); err != nil || v == "" {
		t.Fatalf("the outcome is not on the record (%q, %v) — this test needs a decided run", v, err)
	}
	at := time.Now()
	if l := Assess(run, at); l.State == StateEnded {
		t.Fatalf("a run with an outcome and no capture mark reads %s (%s) — an outcome is not the run's end", l.State, l.Basis)
	}
	if err := MarkEnded(run, at.UTC().Format(time.RFC3339)); err != nil {
		t.Fatal(err)
	}
	if l := Assess(run, at); l.State != StateEnded {
		t.Errorf("a captured run reads %s (%s), want ENDED", l.State, l.Basis)
	}
	// The record's own activity is asked apart from the disposition: capture asks it of a run it is
	// capturing again, and a mark there must not answer for the record.
	if l := AssessActivity(run, at); l.State == StateEnded {
		t.Errorf("AssessActivity read the capture mark: %s", l.State)
	}
}

func TestAssessSaysLiveWhileTheRecordIsMoving(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	dir := runWithCadence(t, 10, 20*time.Second, now)
	l := Assess(mustRun(t, dir), now.Add(30*time.Second))
	if l.State != StateLive {
		t.Fatalf("state = %s (%s), want LIVE", l.State, l.Says())
	}
	if !strings.Contains(l.Says(), "live") {
		t.Errorf("Says() does not read as live: %s", l.Says())
	}
}

// THE MEASURED CASE. Run 2's workflow was killed at 05:45 and every instrument still said "live"
// 41 minutes later. This is that run, in miniature.
func TestAssessSaysStaleWhenTheRecordStopped(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	dir := runWithCadence(t, 10, 20*time.Second, now)
	l := Assess(mustRun(t, dir), now.Add(41*time.Minute))
	if l.State != StateStale {
		t.Fatalf("state = %s, want STALE — a run silent for 41 minutes at a 20s cadence is not live", l.State)
	}
	says := l.Says()
	// THE VERDICT OWES ITS BASIS. "Stopped" is a judgement, and a reader who disagrees needs the
	// age and the threshold to argue with, not just the word.
	for _, want := range []string{"STALE", "41m", "Resume"} {
		if !strings.Contains(says, want) {
			t.Errorf("the STALE line does not carry %q — a reader cannot act on it:\n%s", want, says)
		}
	}
	if l.Age < 40*time.Minute {
		t.Errorf("age = %s, want ~41m", l.Age)
	}
}

// A THRESHOLD NOBODY CAN JUSTIFY IS THE SAME DEFECT AS A FACT NOBODY CAN REFUSE. Too few events
// to characterise the cadence must report itself, not fall back to LIVE — which is the exact
// substitution that let a dead run render as healthy.
func TestAssessSaysNotMeasuredRatherThanGuessing(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	dir := runWithCadence(t, 3, time.Minute, now)
	l := Assess(mustRun(t, dir), now.Add(2*time.Hour))
	if l.State != StateUnmeasured {
		t.Fatalf("state = %s, want NOT MEASURED with only 3 events", l.State)
	}
	if l.State == StateLive {
		t.Fatal("a run with too little record to judge must never read as LIVE")
	}
	if !strings.Contains(l.Basis, "too few") {
		t.Errorf("the basis does not say why it could not measure: %q", l.Basis)
	}
}

// THE THRESHOLD IS DERIVED, NOT A CONSTANT — a slow run must not be called dead for being slow.
func TestThresholdFollowsTheRunsOwnCadence(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	slow := Assess(mustRun(t, runWithCadence(t, 10, 8*time.Minute, now)), now.Add(time.Minute))
	fast := Assess(mustRun(t, runWithCadence(t, 10, 5*time.Second, now)), now.Add(time.Minute))
	if slow.Threshold <= fast.Threshold {
		t.Errorf("a run writing every 8m got threshold %s, one writing every 5s got %s — the "+
			"threshold is not following the cadence", slow.Threshold, fast.Threshold)
	}
	if fast.Threshold != minStaleFloor {
		t.Errorf("a fast run's threshold = %s, want the %s floor: 3x a 5s gap would call a run "+
			"dead during one ordinary seat", fast.Threshold, minStaleFloor)
	}
	// And the slow run, one minute quiet, is still LIVE — the point of deriving it.
	if slow.State != StateLive {
		t.Errorf("slow run state = %s, want LIVE one minute into an 8-minute cadence", slow.State)
	}
}

// THE ZERO VALUE IS NOT A MEASUREMENT. A caller that never assessed must not render as though the
// record was consulted and found wanting.
func TestZeroLivenessDoesNotClaimToHaveMeasured(t *testing.T) {
	var l Liveness
	s := l.Says()
	// Asserting on the absence of "live" was this test's own first draft, and it failed: "live"
	// is a substring of "liveness". A check that matches a word it did not mean is the defect
	// this package is about, one level up — so the assertion names the VERDICTS it must not make.
	for _, verdict := range []string{"NOT MEASURED", "STALE", "run complete", "stale after"} {
		if strings.Contains(s, verdict) {
			t.Errorf("the zero value says %q, which claims the verdict %q — it must read as "+
				"never-assessed, so a caller that forgot to assess is visible", s, verdict)
		}
	}
	if !strings.Contains(s, "not assessed") {
		t.Errorf("the zero value says %q — it should say plainly that nobody looked", s)
	}
}
