package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"testing"
)

func runWith(t *testing.T, maxRounds string, evs []*Event) string {
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
	recordtest.Seed(t, dir, evs...)
	return dir
}

// vev builds a fixture act. The key is composed from what the event actually SAYS — the type is
// derived from the body — rather than from a word passed alongside it that could disagree.
func vev(t *testing.T, seat string, round int, body proto.Message) *Event {
	t.Helper()
	ev := recordtest.Event(t, seat, round, body)
	ev.Key = proto.String(seat + ":" + recordpb.Word(ev.GetType()))
	return ev
}

// A PASS on the record is VERIFIED, without anyone saying so.
func TestVerifiedIsDerivedFromThePassEvent(t *testing.T) {
	dir := runWith(t, "3", []*Event{vev(t, "red-merge-r1", 1, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})})
	got, why, ok := DeriveVerdict(dir)
	if !ok || got != "VERIFIED" {
		t.Errorf("got %q (ok=%v) — want VERIFIED: %s", got, ok, why)
	}
}

// Reaching the ceiling with no pass is CEILING — computed from the rounds on the record
// against the bound setup wrote, so nobody has to be told.
func TestCeilingIsDerivedFromTheRoundsAndTheConfiguredBound(t *testing.T) {
	dir := runWith(t, "2", []*Event{
		vev(t, "red-merge-r1", 1, &recordpb.Position{Text: proto.String("x")}),
		vev(t, "red-merge-r2", 2, &recordpb.Position{Text: proto.String("y")}),
	})
	got, why, ok := DeriveVerdict(dir)
	if !ok || got != "CEILING" {
		t.Errorf("got %q (ok=%v) — want CEILING: %s", got, ok, why)
	}
}

// A HALT OUTRANKS A PASS. A run stopped on safety or integrity grounds did not end by
// passing, however clean the board looked when it stopped.
func TestHaltOutranksAPass(t *testing.T) {
	dir := runWith(t, "3", []*Event{
		vev(t, "red-merge-r1", 1, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}),
		vev(t, "judge-r1", 1, &recordpb.Halt{Opinion: proto.String("consent gate")}),
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
	dir := runWith(t, "5", []*Event{vev(t, "red-merge-r1", 1, &recordpb.Position{Text: proto.String("x")})})
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
	dir := runWith(t, "", []*Event{vev(t, "red-merge-r9", 9, &recordpb.Position{Text: proto.String("x")})})
	if _, _, ok := DeriveVerdict(dir); ok {
		t.Error("a ceiling verdict was derived with no configured ceiling")
	}
}
