package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// blueSatOwing puts blue-respond in a sitting engaged on G1 that filed `filed` and still owes the
// other half of its record — the state the engine re-prompts, and the one a repair is admitted for.
func blueSatOwing(t *testing.T, filed []string) string {
	t.Helper()
	runDir := newRun(t)
	stageEvents(t, runDir,
		recordtest.At(t, "red-chair", "red-chair:stage:d", &recordpb.Dispatch{Pin: proto.Int64(2),
			SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}))
	must(t, runDir, "register", "--seat-id", "blue-respond")
	must(t, runDir, append(append([]string{}, filed...), "--seat-id", "blue-respond")...)
	return runDir
}

// A SINGLETON ACT IS ONCE PER SITTING, AND A REPAIR IS THE SAME SITTING (#1026). The duty was
// enforced only by the idempotency key colliding, and that key counts TURNS: a repair is a turn of
// its own, so the ordinal advanced and the second act was written with nothing refusing it.
// Measured before this fix — two positions on one sitting, err=<nil>, both rendered as
// `blue-respond #2`, so the attribution #1020 introduced is what hid the double file.
//
// Each arm files ONE of the pair, which is what makes the repair admissible: a sitting that carries
// both owes nothing and checkRepair refuses the register itself.
func TestASingletonActIsRefusedASecondTimeInsideARepair(t *testing.T) {
	for _, c := range []struct {
		word string
		act  []string
		typ  recordpb.EventType
	}{
		{"position", []string{"position", "--reason", "the report answers G1"}, recordpb.EventType_EVENT_TYPE_POSITION},
		{"revision", []string{"revision", "--reason", "the sitting's edits"}, recordpb.EventType_EVENT_TYPE_REVISION},
	} {
		t.Run(c.word, func(t *testing.T) {
			runDir := blueSatOwing(t, c.act)
			must(t, runDir, "register", "--seat-id", "blue-respond", "--repair-sitting")

			_, err := run(t, append(append([]string{}, c.act...), "--run", runDir, "--seat-id", "blue-respond")...)
			if err == nil {
				t.Fatalf("a second %s inside the repair was accepted — the sitting now carries two answers to a once-per-sitting question", c.word)
			}
			if want := "blue-respond has already recorded a " + c.word + " this sitting"; !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal does not say what was wrong:\n%v", err)
			}
			n := 0
			for _, e := range events(t, runDir) {
				if e.GetType() == c.typ {
					n++
				}
			}
			if n != 1 {
				t.Errorf("the record carries %d %s events, want the one the sitting filed", n, c.word)
			}
		})
	}
}
