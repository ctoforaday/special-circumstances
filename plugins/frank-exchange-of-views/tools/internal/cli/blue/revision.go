package blue

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// revision: the round-record event (W1.7), and the last thing blue does in a sitting.
//
// A revision that is not on the record did not happen as far as the debate is
// concerned. Run 5 shipped a round-3 report revision with neither a debate.md
// section nor a CHANGELOG entry: a lens misjudged the round state from the
// absence, and the judge had to reconstruct blue's whole position from red's
// ledger. This event is the machine-checkable half of that fix.
func newRevision() *cobra.Command {
	// claim_count is NO LONGER carried here. It was hand-typed via --claim-count and
	// stored on the revision event, a second source for a number that only mattered in
	// the envelope (which debate.js reads to size dispatch); revision itself is not on
	// the live path — the round narrative is a position event and the round record was
	// hand-written in a file (retired, #251). #70 moved the count to the deterministic `count-claims` command and
	// dropped this flag, so there is exactly one way the number is produced.
	return seat.Prose(seat.New("revision",
		"the round record itself (--reason carries what changed this round) — singleton per seat-round; emit AFTER your report edits land",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			p := record.NewPayload().Set("text", text)
			if _, err := record.Append(s.Identity(), "revision", p); err != nil {
				return nil, err
			}
			// THE DEBT IS NAMED WHERE THE SITTING ENDS. `revision` is emitted after blue's
			// edits land, so it is the last moment blue can still act on an unanswered
			// computation demand — and the merge's refusal, which is excellent, arrives a seat
			// and a round later, when blue has gone.
			//
			// It REPORTS and does not refuse. Blue's legitimate answer to a computation demand
			// it thinks is wrong is to argue against the gap, and there is no verb for
			// contesting a check_kind — so a gate here would trap a seat that disagreed. What
			// was measured is not defiance but not-noticing: making `check_kind` visible moved
			// `prove` from 0 uses in eighteen sittings to 1 in nine, which says a seat reading
			// a property still does not read a debt.
			return revisionResult{Owed: record.GapsAwaitingProof(s.RunDir)}, nil
		}))
}

type revisionResult struct {
	// Owed is the open computation gaps with no proof against them, at the moment blue's
	// sitting closes. Empty on a clean round; a list is a debt blue can still discharge.
	Owed []string `json:"awaiting_proof,omitempty"`
}

func (r revisionResult) Human() string {
	m := "revision recorded — the round is on the record"
	if len(r.Owed) == 0 {
		return m
	}
	return m + "\n\nSTILL AWAITING A COMPUTATION: " + strings.Join(r.Owed, ", ") +
		". These gaps were minted --check-kind computation, which prose cannot close — the merge " +
		"is REFUSED if it tries, so an unanswered one carries into the next round rather than " +
		"settling. Settle each with `blue prove --location \"<the sentence>\" --script <path> " +
		"--answers <gap>`, or argue in your next edit's --reason why the demand is wrong."
}
