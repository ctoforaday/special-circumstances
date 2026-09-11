package bench

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/spf13/cobra"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// UNVERIFIED IS THE ONE TERMINAL VERDICT THE RECORD CANNOT DERIVE, AND THE ONLY ONE THE BENCH
// MAY ASSERT.
//
// DeriveVerdict says so itself: no pass, no halt, no board at its ceiling means the record holds
// no terminal state — the run is in flight, or it ended before reaching one — and the record
// cannot tell those apart from inside. So the bench's word stands there, and ONLY there: any other
// word over such a record is a claim the tool would refuse if it could decide, arriving through
// the one gap where it cannot, and it is refused by name.
//
// A bench seat found the account channel missing from the other side and filed it as friction
// (#375). These tests pin the contract: the account is required, the asserted word is UNVERIFIED,
// and the refusals do not misdescribe the record.
func TestOutcomeAssertsOnlyUnverifiedAndAlwaysWithAnAccount(t *testing.T) {
	for _, tc := range []struct {
		name             string
		as, reason       string
		wantErr, wantSay string
	}{
		{
			name: "an ended-early run with no account is refused", as: "UNVERIFIED", reason: "",
			wantErr: "yes", wantSay: "reason",
		},
		{
			name: "an ended-early run with its account is recorded, asserted", as: "UNVERIFIED",
			reason: "the workflow stopped with nobody ready and neither PASS nor the ceiling held",
		},
		{
			name: "a verdict the record cannot derive is refused when it is not UNVERIFIED", as: "VERIFIED",
			reason: "the bench says it passed", wantErr: "yes", wantSay: "Only UNVERIFIED may be asserted",
		},
		{
			name: "CEILING cannot be asserted either — it is the board's to derive", as: "CEILING",
			reason: "the bench says the terms ran out", wantErr: "yes", wantSay: "Only UNVERIFIED may be asserted",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := recordtest.TmpRun(t)
			// Identity comes from the injection, as it does in a real run: the hook exports the
			// agent handle, `register` binds it to this seat, and every later call resolves the
			// seat from that binding rather than from a flag. So the handle is set BEFORE the
			// register that writes it — afterwards there would be nothing to bind.
			t.Setenv(seatenv.AgentVar, "agent_bench")
			if _, _, err := record.RegisterSeat(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, ""); err != nil {
				t.Fatal(err)
			}
			t.Setenv(seatenv.Var, runDir)
			args := []string{"outcome", "--as", tc.as}
			if tc.reason != "" {
				args = append(args, "--reason", tc.reason)
			}
			c := testRoot()
			c.SetArgs(args)
			c.SetOut(&strings.Builder{})
			c.SetErr(&strings.Builder{})
			err := c.Execute()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("refused a valid outcome: %v", err)
				}
				b, err := record.FamilyOf(runtest.Open(t, runDir))
				if err != nil {
					t.Fatal(err)
				}
				for _, e := range b.Events {
					if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok && o.GetVerdictBasis() != record.VerdictAsserted {
						t.Errorf("verdict_basis = %q, want %q — the record could not derive this word, and the outcome must say so", o.GetVerdictBasis(), record.VerdictAsserted)
					}
				}
				return
			}
			if err == nil {
				t.Fatal("accepted an outcome the contract forbids")
			}
			if !strings.Contains(err.Error(), tc.wantSay) {
				t.Errorf("the refusal does not say %q: %v", tc.wantSay, err)
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
	runDir := recordtest.TmpRun(t)
	// The bench's own handle is the one that must resolve; the chair is registered here only so
	// its verdict has a seat to hang on, and it binds a different agent for the same reason a run
	// does — two seats are two agents.
	t.Setenv(seatenv.AgentVar, "agent_merge")
	for _, s := range []string{"red-chair"} {
		if _, _, err := record.RegisterSeat(record.Identity{Run: runtest.Open(t, runDir), SeatID: s}, ""); err != nil {
			t.Fatal(err)
		}
	}
	// A PASS on the record makes VERIFIED derivable, with a stated basis.
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"}, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.AgentVar, "agent_bench")
	if _, _, err := record.RegisterSeat(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, ""); err != nil {
		t.Fatal(err)
	}
	c := testRoot()
	t.Setenv(seatenv.Var, runDir)
	c.SetArgs([]string{"outcome", "--as", "VERIFIED", "--reason", "red passed and no gap survived it"})
	c.SetOut(&strings.Builder{})
	c.SetErr(&strings.Builder{})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}

	b, err := record.FamilyOf(runtest.Open(t, runDir))
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

// testRoot mounts this seat's verbs on a bare root — what the CLI now does for a dispatched bench
// seat. The role GROUP this replaced no longer exists: the tree is scoped to the injected role.
func testRoot() *cobra.Command {
	c := &cobra.Command{Use: "bench", SilenceUsage: true, SilenceErrors: true}
	c.AddCommand(Verbs()...)
	return c
}
