package bench

import (
	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// petition-rule: grant or deny a petition.
//
// A halt is deliberately NOT a value of --as. Ending the run is a different kind
// of act from granting relief, and giving it its own verb means it can never be
// reached by a typo in an enum.
//
// That last sentence was UNTRUE for as long as it was written here, in exactly the way
// `opinion` was: --as took any string, so `petition-rule --as halt` recorded a ruling
// whose word reads as the run's terminal act. `opinion` was patched by refusing the one
// word; here the set is genuinely closed (granted | denied, no "..."), so it is declared
// in record.EnumFields and enforced at the write path — which is what makes the claim
// above a property rather than an intention.
func newPetitionRule() *cobra.Command {
	c := seat.Prose(seat.New("petition-rule",
		"rule on a petition: --petitioner <seatId> --petition-class <class> --as "+record.MustEnum("petition-rule", "ruling").Spelling()+" --reason <opinion> (a halt is its own verb)",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			p := seat.SetSame(cmd, record.NewPayload(), flags.Petitioner)
			seat.Set(cmd, p, "class", flags.PetitionClass)
			seat.Set(cmd, p, "ruling", flags.As)
			p.Set("opinion", text)
			if _, err := record.Append(s.RunDir, s.SeatID, "petition-rule", p); err != nil {
				return nil, err
			}
			return petitionRuleResult{Ruling: seat.Str(cmd, flags.As), Petitioner: seat.Str(cmd, flags.Petitioner)}, nil
		}))

	c.Flags().String(flags.Petitioner, "", "the seat that filed the petition")
	c.Flags().String(flags.PetitionClass, "", record.MustEnum("petition-rule", "class").Usage("what the bench is being asked to hear"))
	c.Flags().String(flags.As, "", record.MustEnum("petition-rule", "ruling").Usage("a halt is its own verb"))
	return c
}

type petitionRuleResult struct {
	Ruling     string `json:"ruling"`
	Petitioner string `json:"petitioner"`
}

func (r petitionRuleResult) Human() string { return "petition " + r.Ruling + " (" + r.Petitioner + ")" }
