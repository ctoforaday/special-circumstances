package record

import (
	"os"
	"path/filepath"
	"testing"
)

func runWith(t *testing.T, maxRounds string, evs []Event) string {
	t.Helper()
	dir := t.TempDir()
	if maxRounds != "" {
		if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"),
			[]byte(`{"maxRounds":"`+maxRounds+`"}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	recs := filepath.Join(dir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	byShard := map[string][]Event{}
	for _, e := range evs {
		byShard[e.SeatID] = append(byShard[e.SeatID], e)
	}
	for seat, list := range byShard {
		var body []byte
		for _, e := range list {
			b, err := MarshalEvent(e)
			if err != nil {
				t.Fatal(err)
			}
			body = append(append(body, b...), '\n')
		}
		if err := os.WriteFile(filepath.Join(recs, "events-"+seat+"-a0000000.jsonl"), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func vev(seat, typ string, round, seq int, p *Payload) Event {
	return Event{Seq: seq, SeatID: seat, Nonce: "a0000000", Round: round, Type: typ, Key: seat + ":" + typ, Payload: p}
}

// A PASS on the record is VERIFIED, without anyone saying so.
func TestVerifiedIsDerivedFromThePassEvent(t *testing.T) {
	dir := runWith(t, "3", []Event{vev("red-merge-r1", "verdict", 1, 0, NewPayload().Set("verdict", "PASS"))})
	got, why, ok := DeriveVerdict(dir)
	if !ok || got != "VERIFIED" {
		t.Errorf("got %q (ok=%v) — want VERIFIED: %s", got, ok, why)
	}
}

// Reaching the ceiling with no pass is CEILING — computed from the rounds on the record
// against the bound setup wrote, so nobody has to be told.
func TestCeilingIsDerivedFromTheRoundsAndTheConfiguredBound(t *testing.T) {
	dir := runWith(t, "2", []Event{
		vev("red-merge-r1", "position", 1, 0, NewPayload().Set("text", "x")),
		vev("red-merge-r2", "position", 2, 0, NewPayload().Set("text", "y")),
	})
	got, why, ok := DeriveVerdict(dir)
	if !ok || got != "CEILING" {
		t.Errorf("got %q (ok=%v) — want CEILING: %s", got, ok, why)
	}
}

// A HALT OUTRANKS A PASS. A run stopped on safety or integrity grounds did not end by
// passing, however clean the board looked when it stopped.
func TestHaltOutranksAPass(t *testing.T) {
	dir := runWith(t, "3", []Event{
		vev("red-merge-r1", "verdict", 1, 0, NewPayload().Set("verdict", "PASS")),
		vev("judge-r1", "halt", 1, 0, NewPayload().Set("opinion", "consent gate")),
	})
	got, _, ok := DeriveVerdict(dir)
	if !ok || got != "HALTED" {
		t.Errorf("got %q (ok=%v) — a halt must outrank a pass", got, ok)
	}
}

// THE ONE CASE THE RECORD CANNOT DECIDE, and it must say so rather than guess. A run that
// ends early with no pass and no halt ended on a judged deadlock — a determination that lives
// only in the bench's envelope and leaves no independent trace (#289).
func TestAJudgedDeadlockIsNotDerivable(t *testing.T) {
	dir := runWith(t, "5", []Event{vev("red-merge-r1", "position", 1, 0, NewPayload().Set("text", "x"))})
	got, why, ok := DeriveVerdict(dir)
	if ok {
		t.Errorf("derived %q from a record that cannot decide — the deadlock case must stay honest", got)
	}
	if why == "" {
		t.Error("the refusal must explain WHY it cannot decide, or the gap is invisible")
	}
}

// An absent or unparseable ceiling degrades CEILING to underivable rather than inventing a
// bound — the same posture as InferRunDir's "say nothing rather than guess".
func TestNoConfiguredCeilingMeansNoCeilingVerdict(t *testing.T) {
	dir := runWith(t, "", []Event{vev("red-merge-r9", "position", 9, 0, NewPayload().Set("text", "x"))})
	if _, _, ok := DeriveVerdict(dir); ok {
		t.Error("a ceiling verdict was derived with no configured ceiling")
	}
}
