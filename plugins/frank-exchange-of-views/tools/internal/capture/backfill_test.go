package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The audit this replaced had NO test — capture_test.go covered 14 things and neither
// RecordJoinAudit nor AttestationAudit was among them, which is how five independent defects
// lived in one regex (#223). These tests exist so the replacement cannot repeat that.

const bfStamp = "2006-01-02T15:04:05.000000000Z"

// bfSeat writes one seat's shard: a register at t0, then n events at the given offsets from t0.
// Stamps are written verbatim so a test can state the timing it means rather than depend on a
// clock.
func bfSeat(t *testing.T, dir, seat string, nonce string, t0 time.Time, offsets []time.Duration) {
	t.Helper()
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	var lines []byte
	add := func(e record.Event) {
		b, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, append(b, '\n')...)
	}
	add(record.Event{Seq: 0, TS: t0.Format(bfStamp), SeatID: seat, Nonce: nonce, Round: 1,
		Type: "register", Key: seat + ":register", Payload: record.NewPayload()})
	for i, off := range offsets {
		add(record.Event{Seq: i + 1, TS: t0.Add(off).Format(bfStamp), SeatID: seat, Nonce: nonce, Round: 1,
			Type: "finding", Key: seat + ":finding:F" + string(rune('1'+i)),
			Payload: record.NewPayload().Set("reason", "r")})
	}
	if err := os.WriteFile(filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl"), lines, 0o644); err != nil {
		t.Fatal(err)
	}
}

func bfT0(t *testing.T) time.Time {
	t.Helper()
	// A fixed instant: Now() is banned in these tests for the same reason it is banned in the
	// stamps — a timing assertion that depends on when it ran is not an assertion.
	ts, err := time.Parse(bfStamp, "2026-08-15T00:00:00.000000000Z")
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

// A seat that works for ten minutes and then writes everything in under a second is narrating.
func TestBackfillAuditWarnsOnAClosingBurst(t *testing.T) {
	dir := t.TempDir()
	t0 := bfT0(t)
	base := 10 * time.Minute
	offs := []time.Duration{base, base + 100*time.Millisecond, base + 200*time.Millisecond,
		base + 300*time.Millisecond, base + 400*time.Millisecond, base + 500*time.Millisecond}
	bfSeat(t, dir, "red-lens-r1-L1", "0000000a", t0, offs)

	a := BackfillAudit(dir)
	if a.Verdict != "WARN" {
		t.Fatalf("verdict = %q, want WARN\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "red-lens-r1-L1") {
		t.Errorf("detail does not name the seat:\n%s", a.Detail)
	}
	if !strings.Contains(a.Detail, "narration") {
		t.Errorf("detail does not say what the finding MEANS, only that it fired:\n%s", a.Detail)
	}
}

// The same six events spread across the sitting are contemporaneous and must not warn.
// This is the half that matters: an audit that warns on everything is not a signal.
func TestBackfillAuditPassesWhenRecordingIsSpreadAcrossTheSitting(t *testing.T) {
	dir := t.TempDir()
	t0 := bfT0(t)
	offs := []time.Duration{1 * time.Minute, 3 * time.Minute, 5 * time.Minute,
		7 * time.Minute, 9 * time.Minute, 10 * time.Minute}
	bfSeat(t, dir, "red-lens-r1-L1", "0000000b", t0, offs)

	a := BackfillAudit(dir)
	if a.Verdict != "PASS" {
		t.Fatalf("verdict = %q, want PASS\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "every measured seat") {
		t.Errorf("a PASS should say what it measured:\n%s", a.Detail)
	}
}

// Below the floor there is no difference between "clustered" and "this seat had three things to
// say", and the audit must say it measured NOTHING rather than report a clean board.
func TestBackfillAuditDoesNotJudgeASeatBelowTheEventFloor(t *testing.T) {
	dir := t.TempDir()
	t0 := bfT0(t)
	offs := []time.Duration{10 * time.Minute, 10*time.Minute + time.Millisecond, 10*time.Minute + 2*time.Millisecond}
	bfSeat(t, dir, "red-lens-r1-L1", "0000000c", t0, offs)

	a := BackfillAudit(dir)
	if a.Verdict != "PASS" {
		t.Fatalf("verdict = %q, want PASS (nothing measurable is not a failure)\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "nothing to measure") {
		t.Errorf("an unmeasured run must SAY so — a silent PASS here is the plausible zero this audit replaced:\n%s", a.Detail)
	}
	if !strings.Contains(a.Detail, "0 with enough events") {
		t.Errorf("detail should report the measured count:\n%s", a.Detail)
	}
}

// An unparseable stamp means the event was not seen. That must be LOUD: the shape this audit
// replaced folded its misses into the honest zero, which is what let it read as clean for months.
func TestBackfillAuditReportsUnparseableStampsRatherThanDroppingThem(t *testing.T) {
	dir := t.TempDir()
	t0 := bfT0(t)
	bfSeat(t, dir, "red-lens-r1-L1", "0000000d", t0, []time.Duration{time.Minute, 2 * time.Minute})

	recs := filepath.Join(dir, "records")
	bad := record.Event{Seq: 0, TS: "not-a-timestamp", SeatID: "red-merge-r1", Nonce: "0000000e",
		Round: 1, Type: "finding", Key: "red-merge-r1:finding:F1", Payload: record.NewPayload()}
	b, err := record.MarshalEvent(bad)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(recs, "events-red-merge-r1-0000000e.jsonl"), append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	a := BackfillAudit(dir)
	if !strings.Contains(a.Detail, "NOT MEASURED") {
		t.Fatalf("an unparseable ts must be reported, not silently skipped:\n%s", a.Detail)
	}
	if !strings.Contains(a.Detail, "1 event(s) carried an unparseable ts") {
		t.Errorf("the count of unseen events should be stated:\n%s", a.Detail)
	}
}

// A seat with no register event has no known start, so its lifetime is unknowable and it must not
// be judged — silently guessing a start is how a burst gets excused or invented.
func TestBackfillAuditSkipsASeatWithNoRegister(t *testing.T) {
	dir := t.TempDir()
	t0 := bfT0(t)
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	var lines []byte
	for i := 0; i < 6; i++ {
		e := record.Event{Seq: i, TS: t0.Add(10*time.Minute + time.Duration(i)*time.Millisecond).Format(bfStamp),
			SeatID: "red-lens-r1-L1", Nonce: "0000000f", Round: 1, Type: "finding",
			Key: "red-lens-r1-L1:finding:F" + string(rune('1'+i)), Payload: record.NewPayload()}
		b, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, append(b, '\n')...)
	}
	if err := os.WriteFile(filepath.Join(recs, "events-red-lens-r1-L1-0000000f.jsonl"), lines, 0o644); err != nil {
		t.Fatal(err)
	}

	a := BackfillAudit(dir)
	if a.Verdict != "PASS" {
		t.Fatalf("verdict = %q, want PASS\n%s", a.Verdict, a.Detail)
	}
	if !strings.Contains(a.Detail, "0 with enough events") {
		t.Errorf("a seat with no known start must not be counted as measured:\n%s", a.Detail)
	}
}

// A run with no record at all is SKIP, not PASS. Absent is not clean.
func TestBackfillAuditSkipsWhenThereIsNoRecord(t *testing.T) {
	a := BackfillAudit(t.TempDir())
	if a.Verdict != "SKIP" {
		t.Fatalf("verdict = %q, want SKIP — an unread record is not a passed audit\n%s", a.Verdict, a.Detail)
	}
}

// ParseStamp is the seam that keeps the layout in one place; a caller that re-declares it has
// forked the schema. Round-trip the exact format the tool writes.
func TestParseStampRoundTripsTheWrittenLayout(t *testing.T) {
	want := bfT0(t)
	got, err := record.ParseStamp(want.Format(bfStamp))
	if err != nil {
		t.Fatalf("ParseStamp on a stamp the tool itself writes: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("round trip = %v, want %v", got, want)
	}
	if _, err := record.ParseStamp("not-a-timestamp"); err == nil {
		t.Error("ParseStamp must return an error rather than the zero time — a caller cannot otherwise tell unparseable from the epoch")
	}
}
