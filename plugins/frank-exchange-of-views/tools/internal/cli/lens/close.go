package lens

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// close: retire a gap, WITH the evidence that it is retired.
//
// The verification triple (who checked, with what, against what) is required because E0.5a found
// unanchored closures unauditable after the fact — the record said a thing was checked and could
// not say by whom, how, or against what.
//
// # Why `carry` is a VERB and not a flag on this one
//
// A carry is not a closure with a flag set; it is a DIFFERENT ACT with a different contract, and
// `validate` said so in four branches keyed off `--carried-from`: a carry needs no verification
// triple, is exempt from --reason (it restates a closure that already gave one), is exempt from
// the open-gap check (the gap is closed — that is the point), and must name a real prior closure.
//
// One verb held both, so cobra could require nothing: the triple was three optional flags, and
// the seat that could not produce them found `--carried-from` offered in the same help as the
// easier way out. Two verbs, and each requires what it actually requires.
func newClose() *cobra.Command {
	c := seat.New("close", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		p, err := seat.ClosurePayload(cmd)
		if err != nil {
			return nil, err
		}
		p.AnchorSeat = proto.String(seat.Str(cmd, flags.VerifiedBy))
		p.AnchorTool = proto.String(seat.Str(cmd, flags.VerifiedWith))
		p.AnchorTarget = proto.String(seat.Str(cmd, flags.VerifiedAgainst))
		// A COMPUTATION CHECK CANNOT BE CLOSED BY PROSE.
		//
		// This is what makes --check-kind a demand rather than a label. Red asked for a
		// computation; a closure with no proof answering the gap would be red accepting
		// prose for the one class of check it declared prose cannot settle — which is
		// exactly the round the 2026-08-05 smoke spent: R1-2 asked blue to "test it on a
		// false claim", blue answered by ASSERTING the test had happened, and red's R2-2
		// correctly refused it as "no evidence shown". A full epoch, for something three
		// lines of trial division settle while leaving an artifact red can re-run.
		//
		// The guard reads the board, like the estoppel guard above it. A run whose board
		// cannot be read does not block the closure: refusing on infrastructure would
		// strand an epoch, and the check is a demand, not a safety property.
		if kind, gerr := computationGapKind(run, seat.Str(cmd, flags.ID)); gerr == nil && kind {
			if !record.ProofAnswers(run, seat.Str(cmd, flags.ID)) {
				return nil, fmt.Errorf("lens close: %s was minted --check-kind computation, and no proof answers it. Its acceptance check is settled by RUNNING something, not by reading the report — so closing it on prose would accept the one kind of evidence you declared insufficient. Blue settles it with `blue prove --location \"<the sentence>\" --script <path> --answers %s`; if the demand was wrong, regrade or supersede the gap rather than closing it unproved",
					seat.Str(cmd, flags.ID), seat.Str(cmd, flags.ID))
			}
		}
		if _, err := record.Append(s.Identity(), p); err != nil {
			return nil, err
		}
		return seat.CloseResult{GapID: seat.Str(cmd, flags.ID), Class: recordpb.Word(p.GetClosureClass())}, nil
	})

	seat.ClosureFlags(c)
	c.Flags().String(flags.VerifiedBy, "", "WHO verified it — the seat that read the evidence")
	flags.Text(c, flags.VerifiedWith, "WITH WHAT — the tool or command that showed it")
	c.Flags().String(flags.VerifiedAgainst, "", "AGAINST WHAT — the exact file, ref or URL read")
	// ALL THREE OR NONE, said by cobra rather than by a refusal after the fact. `validate` reads
	// them as one fact (`anchored`), so two of three was never a partial closure — it was an
	// unanchored one that spent the seat's turn before saying so.
	c.MarkFlagsRequiredTogether(flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst)
	_ = c.MarkFlagRequired(flags.VerifiedBy)
	// THE ARGUMENT IS UNCONDITIONAL FOR THIS VERB, AND CONDITIONAL FOR THE MESSAGE — the same
	// split `blue line-of-inquiry propose` makes for --reason, and for the same reason.
	//
	// `Close.prose` carried `required: true`, which refuses unconditionally and therefore refused
	// a CARRY — a carry restates a closure an earlier sitting already argued. Making it conditional
	// fixed the carry and cost THIS verb both its cobra refusal and its REQUIRED marker.
	// seat.ProseRequired restores the two together; separating them is how a parser ends up
	// holding a rule the help does not state.
	return seat.ProseRequired(c)
}

// carry: restate a closure made in an earlier epoch.
//
// Not a mode of `close`. It carries no fresh verification because it makes no fresh claim — the
// epoch it names already did, and the record is checked against that: a carry of a gap with no
// prior closure is refused, because otherwise it is a laundering path for exactly the seat that
// could not produce a verification triple.

// closureFlags are what both verbs take: which gap, how it ended, and where the remainder went.

// closurePayload builds what a closure records before either verb adds its own evidence.

// closeResult names the closed gap and the closure class it was retired under.

// computationGapKind reports whether the named gap was minted as a computation check.
func computationGapKind(run record.Run, gapID string) (bool, error) {
	if gapID == "" {
		return false, nil
	}
	kind, err := record.MintCheckKind(run, gapID)
	if err != nil {
		return false, err
	}
	return kind == recordpb.CheckKind_CHECK_KIND_COMPUTATION, nil
}
