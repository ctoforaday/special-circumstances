package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE ARGUMENT VERBS ARE TOPOLOGY, NOT OBEDIENCE (gblock, 2026-10-20). The position — the argument on
// what is about no one gap — is the chair's, and no blue seat's surface holds it: blue-respond argues a
// gap in its closing, and a lane, the frontier and the synthesizer argue in what they write. The
// closing is a party's — the minting lens's and blue-respond's — and the chair, party to no dispute,
// holds no closing verb.
func TestThePositionIsTheChairsAndTheClosingThePartiesVerb(t *testing.T) {
	for _, seatID := range []string{"blue-respond", "blue-lane-1", "frontier", "blue-synthesize"} {
		t.Run(seatID+" has no position", func(t *testing.T) {
			runDir := newRun(t)
			_, err := run(t, "position", "--run", runDir, "--seat-id", seatID, "--reason", "what the bench is asked to weigh")
			if err == nil || !strings.Contains(err.Error(), `"position" is not on your surface — it is the chair seat's verb`) {
				t.Fatalf("%s filed a position, or was refused it for another reason: %v", seatID, err)
			}
			fam, ferr := record.FamilyOf(runtest.Open(t, runDir))
			if ferr != nil {
				t.Fatal(ferr)
			}
			for _, e := range fam.Events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
					t.Errorf("the refused position is on the record: %v", e)
				}
			}
		})
	}
	t.Run("the chair files a position", func(t *testing.T) {
		if _, err := run(t, "position", "--run", newRun(t), "--seat-id", "red-chair", "--reason", "why the PASS is withheld"); err != nil {
			t.Fatalf("the chair was refused its position: %v", err)
		}
	})
	t.Run("the chair has no closing", func(t *testing.T) {
		_, err := run(t, "closing", "--run", newRun(t), "--seat-id", "red-chair", "--id", "G1", "--reason", "the chair's case")
		if err == nil || !strings.Contains(err.Error(), `"closing" is not on your surface — it is the blue and lens seat's verb`) {
			t.Fatalf("the chair reached a closing verb: %v", err)
		}
	})
	for _, seatID := range []string{lensSeat, "blue-respond"} {
		t.Run(seatID+" holds the closing verb", func(t *testing.T) {
			help, err := run(t, "closing", "--help", "--seat-id", seatID)
			if err != nil || !strings.Contains(help, "owed once per docketing") {
				t.Fatalf("%s's surface does not carry the closing verb (err %v):\n%s", seatID, err, help)
			}
		})
	}
}
