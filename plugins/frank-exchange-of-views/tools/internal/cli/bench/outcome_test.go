package bench

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// A JUDGED DEADLOCK IS THE ONE TERMINAL VERDICT THE RECORD CANNOT DERIVE.
//
// DeriveVerdict says so itself: no pass, no halt, ceiling not reached means the run ended early,
// which only a judged deadlock does, "and that determination is not on the record (#289)". So at
// the single moment the bench is exercising unforced judgement, the tool had nowhere to put it —
// a bare boolean, and nothing else, forever.
//
// A bench seat found this from the other side and filed it as friction (#375), reaching for
// --reason and reporting the asymmetry with `opinion`. These tests pin the contract that answers
// it: required where the judgement is real, refused where it would argue someone else's
// conclusion, and — the arm I got wrong first — a refusal that does not itself misdescribe the
// record.
func TestOutcomeRequiresAnAccountOfAJudgedDeadlock(t *testing.T) {
	for _, tc := range []struct {
		name             string
		deadlocked       bool
		reason           string
		wantErr, wantSay string
	}{
		{
			name: "a deadlock with no account is refused", deadlocked: true, reason: "",
			wantErr: "yes", wantSay: "reason",
		},
		{
			name: "a deadlock with its account is recorded", deadlocked: true,
			reason: "blue and red read the same table oppositely and neither moved",
		},
		{
			name: "an ordinary stamp carries its account too", reason: "red passed on the second round and no gap survived it",
		},
		{
			name:    "no account at all is refused, deadlock or not",
			wantErr: "yes", wantSay: "reason",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := t.TempDir()
			if _, _, err := record.RegisterSeat(record.Identity{RunDir: runDir, SeatID: "judge-r1", Round: record.RoundOf("judge-r1")}); err != nil {
				t.Fatal(err)
			}
			// Identity comes from the injection, as it does in a real run: --run and --seat-id
			// are root persistent flags this subtree does not own, and the engine supplies both.
			t.Setenv(seatenv.Var, runDir)
			t.Setenv(seatenv.SeatVar, "judge-r1")
			args := []string{"outcome", "--as", "UNVERIFIED"}
			if tc.deadlocked {
				args = append(args, "--ended", "deadlock")
			}
			if tc.reason != "" {
				args = append(args, "--reason", tc.reason)
			}
			c := NewCommand()
			c.SetArgs(args)
			c.SetOut(&strings.Builder{})
			c.SetErr(&strings.Builder{})
			err := c.Execute()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("refused a valid outcome: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("accepted an outcome the contract forbids")
			}
			if !strings.Contains(err.Error(), tc.wantSay) {
				t.Errorf("the refusal does not name the missing flag %q: %v", tc.wantSay, err)
			}
			// AND THE REFUSAL MUST NOT MISDESCRIBE THE RECORD. The first version told a seat its
			// verdict was "DERIVED from the record ()" on a run where derivation had failed —
			// a message asserting a falsehood about the very thing it is teaching.
			if strings.Contains(err.Error(), "DERIVED from the record ()") {
				t.Errorf("the refusal claims a derivation that did not happen: %v", err)
			}
		})
	}
}

// The derivation's reasoning was computed on every call and used only to phrase an error, so a
// report could stamp a verdict and never say why it was that one.
func TestOutcomeRecordsWhyTheVerdictIsWhatItIs(t *testing.T) {
	runDir := t.TempDir()
	for _, s := range []string{"judge-r1", "red-merge-r1"} {
		if _, _, err := record.RegisterSeat(record.Identity{RunDir: runDir, SeatID: s, Round: record.RoundOf(s)}); err != nil {
			t.Fatal(err)
		}
	}
	// A PASS on the record makes VERIFIED derivable, with a stated basis.
	if _, err := record.Append(record.Identity{RunDir: runDir, SeatID: "red-merge-r1", Round: record.RoundOf("red-merge-r1")}, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}); err != nil {
		t.Fatal(err)
	}
	c := NewCommand()
	t.Setenv(seatenv.Var, runDir)
	t.Setenv(seatenv.SeatVar, "judge-r1")
	c.SetArgs([]string{"outcome", "--as", "VERIFIED", "--reason", "red passed and no gap survived it"})
	c.SetOut(&strings.Builder{})
	c.SetErr(&strings.Builder{})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}

	b, err := record.BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range b.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_OUTCOME {
			continue
		}
		o, ok := recordpb.BodyAs[*recordpb.Outcome](e)
		if !ok {
			t.Fatal("an outcome event carries no Outcome body")
		}
		if why := o.GetVerdictWhy(); why == "" {
			t.Fatal("the outcome records no verdict_why — the report stamps a verdict it cannot explain")
		} else if !strings.Contains(why, "PASS") {
			t.Errorf("verdict_why = %q, want it to name what decided the verdict", why)
		}
		return
	}
	t.Fatal("no outcome event was recorded")
}
