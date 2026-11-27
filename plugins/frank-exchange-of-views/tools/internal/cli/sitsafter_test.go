package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// storePlan stores the chair's plan of epoch under an epoch limit of limit (0 for none), as the
// chair's dispatch stores the plan it prints: the plan a ruling reads its epoch and limit from.
func storePlan(t *testing.T, runDir string, epoch, limit int) {
	t.Helper()
	p := record.Plan{Epoch: epoch, MaxEpochs: limit, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{},
		Exchanges: []record.ExchangeCount{}, Why: []string{}, StaleAreas: []record.StaleArea{}, Blockers: []record.PlanBlocker{}}
	recordtest.Seed(t, runDir, recordtest.Event(t, "red-chair", &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}))
}

// T-B46: NO RAISER STATEMENT LANDS WHERE NO LENS SITS AFTER, AND NO REMAND WHERE THE GAP'S PARTIES DO
// NOT BOTH SIT AFTER (#1695 §III.B4, lensSitsAfter and partiesSitAfter). A quote gap, blue its author,
// docketed at impasse with red's text stated, is ruled in a docket sitting the plan of each epoch
// convenes under an epoch limit, and in the terminal and the assembly sittings. --for-raiser, which
// applies red's text, lands only where a lens sits after; --remand only where blue sits after for more
// than its closings, so in the epoch before the held-out one it is refused while --for-raiser lands;
// --for-author lands everywhere. The lens-author rows (a remand admitted in the epoch before the
// held-out one) stand on a gap about a gap, whose mint is R4's: TestPartiesSitAfterKeysOnTheAuthor
// holds the predicate's author conjunct.
func TestNoRaiserStatementLandsWhereNoLensSitsAfter(t *testing.T) {
	for _, c := range []struct {
		name              string
		epoch, limit      int
		occasion          string // "" for the docket sitting the dispatch convenes
		forRaiser, remand bool   // whether each lands
	}{
		{"no epoch limit", 3, 0, "", true, true},
		{"two epochs before the held-out one", 3, 6, "", true, true},
		{"the epoch before the held-out one", 4, 6, "", true, false},
		{"the held-out epoch", 5, 6, "", false, false},
		{"the final epoch", 6, 6, "", false, false},
		{"the terminal sitting", 3, 0, "terminal", false, false},
		{"the assembly", 3, 0, "assemble", false, false},
	} {
		for _, answer := range []string{"for-raiser", "remand", "for-author"} {
			t.Run(c.name+", "+answer, func(t *testing.T) {
				runDir, gap, d := positionsRun(t)
				if out, err := redStates(t, runDir, gap); err != nil {
					t.Fatalf("red's closing: %v\n%s", err, out)
				}
				storePlan(t, runDir, c.epoch, c.limit)
				if c.occasion == "" {
					conveneBench(t, runDir, gap)
				} else {
					asSeat(t, runDir, "judge", "register", "--occasion", c.occasion)
				}
				extra := []string{"--final"}
				if answer == "remand" {
					extra = []string{"--direction", "a reproduction"}
				}
				out, err := answerDocketing(t, runDir, d, answer, extra...)
				lands := map[string]bool{"for-raiser": c.forRaiser, "remand": c.remand, "for-author": true}[answer]
				switch {
				case lands && err != nil:
					t.Fatalf("--%s refused: %v\n%s", answer, err, out)
				case !lands && err == nil:
					t.Fatalf("--%s landed; want refused naming who does not sit after", answer)
				case !lands:
					want := map[string]string{"for-raiser": "no lens sits after this ruling", "remand": "a remand grants one more exchange"}[answer]
					if !strings.Contains(err.Error()+out, want) || !strings.Contains(err.Error()+out, "Nothing was recorded") {
						t.Errorf("--%s refused with %v\n%s; want it to name %q", answer, err, out, want)
					}
					if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)); n != 0 {
						t.Errorf("a refused --%s left %d bench edits on the record", answer, n)
					}
				}
			})
		}
	}

	// An answer for the raiser that applies nothing — red's text already stands — lands where no lens
	// sits after: nothing reaches the report that a lens has not audited. After the author's
	// concession (its --accept of red's stated pair) a later docketing of the gap finds red's text
	// applied, and --for-raiser in the terminal sitting writes the ruling alone.
	t.Run("the terminal sitting, red's text applied by the author's concession: --for-raiser lands with no edit", func(t *testing.T) {
		runDir, gap, d := positionsRun(t)
		if out, err := redStates(t, runDir, gap); err != nil {
			t.Fatalf("red's closing: %v\n%s", err, out)
		}
		asSeat(t, runDir, blueSeat, "register")
		if out, err := actOnce(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--accept", gap, "--reason", "red is right", "--json"); err != nil {
			t.Fatalf("the author's concession: %v\n%s", err, out)
		}
		d2 := docketNow(t, runDir, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
		if d2 == d {
			t.Fatal("the docketing again has the conceded one's id")
		}
		// The closing verb plans its pair on the report as it stands, where red's quote is gone: red's
		// statement is written as the record holds it, the same pair, through the record's write.
		asSeat(t, runDir, lensSeat, "register")
		st := &recordpb.Closing{GapId: proto.String(gap), Text: proto.String("red's case again"),
			Location: proto.String(acceptS + "."), FixNew: proto.String(acceptN + ".")}
		if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: lensSeat}, st); err != nil {
			t.Fatalf("red's closing on the later docketing: %v", err)
		}
		asSeat(t, runDir, "judge", "register", "--occasion", "terminal")
		if out, err := answerDocketing(t, runDir, d2, "for-raiser", "--final"); err != nil {
			t.Fatalf("--for-raiser with red's text standing: %v\n%s", err, out)
		}
		if n := len(benchActs(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)); n != 0 {
			t.Errorf("the bench wrote %d edits; red's text already stood", n)
		}
	})
}
