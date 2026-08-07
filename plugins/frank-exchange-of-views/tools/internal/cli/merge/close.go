package merge

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// close: retire a gap, WITH the evidence that it is retired.
//
// The anchor triple (who verified, with what, against what) is required because
// E0.5a found unanchored closures unauditable after the fact — the record said a
// thing was checked and could not say by whom, how, or against what. --carried-
// from is the honest alternative: this closure is not a fresh act, it is last
// round's, restated.
func newClose() *cobra.Command {
	c := seat.New("close",
		`close a gap WITH its verification anchor: --id R2-3 --as closed|closed_with_regression|... (--anchor-seat L1 --anchor-tool "git show" --anchor-target "7bc501e:path" | --carried-from <round>) [--successor R3-1] --reason "<what was verified and why it holds>"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			class := seat.Str(cmd, flags.As)
			if class == "" {
				class = "closed"
			}
			p := seat.Set(cmd, record.NewPayload(), "gap_id", flags.ID)
			p.Set("closure_class", class)
			seat.Set(cmd, p, "anchor_seat", flags.AnchorSeat)
			seat.Set(cmd, p, "anchor_tool", flags.AnchorTool)
			seat.Set(cmd, p, "anchor_target", flags.AnchorTarget)
			seat.Set(cmd, p, "carried_from", flags.CarriedFrom)
			seat.SetSame(cmd, p, flags.Successor)
			// The SHARED prose channel, not a private one. close hand-rolled its own
			// --file read and so was the only prose-bearing verb with no --text at all —
			// the same shape as the --prose-file divergence, one layer down: a verb that
			// opts out of the shared helper drifts from it by construction.
			if err := seat.SetReason(cmd, p, "prose"); err != nil {
				return nil, err
			}
			// A COMPUTATION CHECK CANNOT BE CLOSED BY PROSE.
			//
			// This is what makes --check-kind a demand rather than a label. Red asked for a
			// computation; a closure with no proof answering the gap would be red accepting
			// prose for the one class of check it declared prose cannot settle — which is
			// exactly the round the 2026-08-05 smoke spent: R1-2 asked blue to "test it on a
			// false claim", blue answered by ASSERTING the test had happened, and red's R2-2
			// correctly refused it as "no evidence shown". A full round, for something three
			// lines of trial division settle while leaving an artifact red can re-run.
			//
			// The guard reads the board, like the estoppel guard above it. A run whose board
			// cannot be read does not block the closure: refusing on infrastructure would
			// strand a round, and the check is a demand, not a safety property.
			if kind, gerr := computationGapKind(s.RunDir, seat.Str(cmd, flags.ID)); gerr == nil && kind {
				if !record.ProofAnswers(s.RunDir, seat.Str(cmd, flags.ID)) {
					return nil, fmt.Errorf("merge close: %s was minted --check-kind computation, and no proof answers it. Its acceptance check is settled by RUNNING something, not by reading the report — so closing it on prose would accept the one kind of evidence you declared insufficient. Blue settles it with `blue prove --location \"<the sentence>\" --script <path> --answers %s`; if the demand was wrong, regrade or supersede the gap rather than closing it unproved",
						seat.Str(cmd, flags.ID), seat.Str(cmd, flags.ID))
				}
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "close", p); err != nil {
				return nil, err
			}
			return closeResult{GapID: seat.Str(cmd, flags.ID), Class: class}, nil
		})

	c.Flags().String(flags.ID, "", "the gap id")
	c.Flags().String(flags.As, "", "closed | closed_with_regression | ... — the closure class")
	c.Flags().String(flags.AnchorSeat, "", "WHO verified the closure (the seat)")
	c.Flags().String(flags.AnchorTool, "", "WITH WHAT it was verified (the tool or command)")
	c.Flags().String(flags.AnchorTarget, "", "AGAINST WHAT — the exact file, ref or URL read")
	c.Flags().String(flags.CarriedFrom, "", "the round this closure was carried from, when it is not a fresh act")
	c.Flags().String(flags.Successor, "", "the gap id carrying the unresolved remainder forward")
	return seat.Prose(c)
}

// closeResult names the closed gap and the closure class it was retired under.
type closeResult struct {
	GapID string `json:"gap_id"`
	Class string `json:"class"`
}

func (r closeResult) Human() string { return "closed " + r.GapID + " (" + r.Class + ")" }

// computationGapKind reports whether the named gap was minted as a computation check.
func computationGapKind(runDir, gapID string) (bool, error) {
	if gapID == "" {
		return false, nil
	}
	b, err := record.BoardState(runDir)
	if err != nil {
		return false, err
	}
	g := b.Gaps[gapID]
	if g == nil || g.Mint == nil {
		return false, nil
	}
	return g.Mint.Str("check_kind") == "computation", nil
}
