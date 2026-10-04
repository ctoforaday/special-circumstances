package record

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

func runWith(t *testing.T, maxRounds string, evs []*Event) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
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
	ev := recordtest.Event(t, seat, body)
	// One chair, several epochs: the key carries the round so two verdicts by red-chair are two rows.
	ev.Key = proto.String(seat + ":" + recordpb.Word(ev.GetType()) + ":" + strconv.Itoa(round))
	return ev
}

// A PASS on the record is VERIFIED, without anyone saying so.
func TestVerifiedIsDerivedFromThePassEvent(t *testing.T) {
	dir := runWith(t, "3", []*Event{vev(t, "red-chair", 1, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})})
	got, why, ok, err := DeriveVerdict(mustRun(t, dir))
	if err != nil || !ok || got != "VERIFIED" {
		t.Errorf("got %q (ok=%v, err=%v) — want VERIFIED: %s", got, ok, err, why)
	}
}

// A HALT OUTRANKS A PASS. A run stopped on safety or integrity grounds did not end by
// passing, however clean the board looked when it stopped.
func TestHaltOutranksAPass(t *testing.T) {
	dir := runWith(t, "3", []*Event{
		vev(t, "red-chair", 1, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}),
		vev(t, "judge", 1, &recordpb.Halt{Opinion: proto.String("consent gate")}),
	})
	got, _, ok, err := DeriveVerdict(mustRun(t, dir))
	if err != nil || !ok || got != "HALTED" {
		t.Errorf("got %q (ok=%v, err=%v) — a halt must outrank a pass", got, ok, err)
	}
}

// THE ONE CASE THE RECORD CANNOT DECIDE, and it must say so rather than guess. A run that
// ends early with no pass, no halt and no board at its ceiling ended UNVERIFIED — and the
// record cannot tell that from a run still in flight, so it refuses to derive rather than guess.
func TestARunThatEndedEarlyIsNotDerivable(t *testing.T) {
	dir := runWith(t, "5", []*Event{vev(t, "red-chair", 1, &recordpb.Position{Text: proto.String("x")})})
	got, why, ok, err := DeriveVerdict(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Errorf("derived %q from a record that cannot decide — the ended-early case must stay honest", got)
	}
	if why == "" {
		t.Error("the refusal must explain WHY it cannot decide, or the gap is invisible")
	}
}

// An absent or unparseable ceiling degrades CEILING to underivable rather than inventing a
// bound — the same posture as InferRunDir's "say nothing rather than guess".
