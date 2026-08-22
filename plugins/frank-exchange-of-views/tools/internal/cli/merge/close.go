package merge

import (
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
	c := seat.New("close",
		`close a gap WITH its verification: --id R2-3 --as closed|closed_with_regression|... --verified-by L1 --verified-with "git show" --verified-against "7bc501e:path" [--superseded-by R3-1] --reason "<what was verified and why it holds>"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p, err := closurePayload(cmd)
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
			if _, err := record.Append(s.Identity(), p); err != nil {
				return nil, err
			}
			return closeResult{GapID: seat.Str(cmd, flags.ID), Class: recordpb.Word(p.GetClosureClass())}, nil
		})

	closureFlags(c)
	c.Flags().String(flags.VerifiedBy, "", "WHO verified it — the seat that read the evidence")
	c.Flags().String(flags.VerifiedWith, "", "WITH WHAT — the tool or command that showed it")
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
	// a CARRY — a carry restates a closure an earlier round already argued. Making it conditional
	// fixed the carry and cost THIS verb both its cobra refusal and its REQUIRED marker.
	// seat.ProseRequired restores the two together; separating them is how a parser ends up
	// holding a rule the help does not state.
	return seat.ProseRequired(c)
}

// carry: restate a closure made in an earlier round.
//
// Not a mode of `close`. It carries no fresh verification because it makes no fresh claim — the
// round it names already did, and the record is checked against that: a carry of a gap with no
// prior closure is refused, because otherwise it is a laundering path for exactly the seat that
// could not produce a verification triple.
func newCarry() *cobra.Command {
	c := seat.Records(seat.New("carry",
		`restate a closure from an earlier round: --id R2-3 --carried-from <round> [--as closed|...] [--superseded-by R3-1]. `+
			`It makes no fresh claim, so it needs no --verified-by triple — but the record must already hold a closure of this gap, or the carry is refused.`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p, err := closurePayload(cmd)
			if err != nil {
				return nil, err
			}
			p.CarriedFrom = proto.String(seat.Str(cmd, flags.CarriedFrom))
			if _, err := record.Append(s.Identity(), p); err != nil {
				return nil, err
			}
			return closeResult{GapID: seat.Str(cmd, flags.ID), Class: recordpb.Word(p.GetClosureClass()), Carried: true}, nil
		}), "close")

	closureFlags(c)
	c.Flags().String(flags.CarriedFrom, "", "the round whose closure this restates")
	_ = c.MarkFlagRequired(flags.CarriedFrom)
	return seat.Prose(c)
}

// closureFlags are what both verbs take: which gap, how it ended, and where the remainder went.
func closureFlags(c *cobra.Command) {
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	_ = c.MarkFlagRequired(flags.ID)
	enumhelp.Flag(c, flags.As, record.MustEnum("close", "closure_class"), ("HOW the gap ended. One vocabulary with the bench's dispositions since #342 — a reader no longer has to know which verb produced a closure before it can interpret the word"))
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.SupersededBy, "the gap id carrying the unresolved remainder forward")
}

// closurePayload builds what a closure records before either verb adds its own evidence.
func closurePayload(cmd *cobra.Command) (*recordpb.Close, error) {
	word := seat.Str(cmd, flags.As)
	if word == "" {
		word = "closed"
	}
	class, ok := record.DispositionOf(word)
	if !ok {
		return nil, feov.Errorf(feov.Validation,
			"merge close: %q is not a disposition — an unrecognized word lands in no bucket and the gap reads as closed for no stated reason", word)
	}
	// A MERGE MAY CLOSE AND MAY NOT CARRY, and the subset comes off the vocabulary rather than from
	// a word typed here. The database refuses the same value through the CHECK the schema generates
	// from `subset: "closes"`, so this refusal is the teaching copy of a constraint that holds even
	// against SQL written straight at the file.
	if !recordpb.Closes(class) {
		return nil, feov.Errorf(feov.Validation,
			"merge close: %q defers the gap instead of closing it, and deferring is the BENCH decision — a close asserts a verified repair. Rule it from the bench with `feov-record bench opinion --as %s`, or close it with a class that states what the repair was", word, word)
	}
	// The SHARED prose channel, not a private one. close hand-rolled its own --file read and so
	// was the only prose-bearing verb with no --text at all — a verb that opts out of the shared
	// helper drifts from it by construction.
	prose, err := seat.Reason(cmd)
	if err != nil {
		return nil, err
	}
	return &recordpb.Close{
		GapId:        proto.String(seat.Str(cmd, flags.ID)),
		ClosureClass: &class,
		Successor:    seat.OptStr(cmd, flags.SupersededBy),
		Prose:        proto.String(prose),
	}, nil
}

// closeResult names the closed gap and the closure class it was retired under.
type closeResult struct {
	GapID   string `json:"gap_id"`
	Class   string `json:"class"`
	Carried bool   `json:"carried,omitempty"`
}

func (r closeResult) Human() string {
	if r.Carried {
		return "carried the closure of " + r.GapID + " (" + r.Class + ")"
	}
	return "closed " + r.GapID + " (" + r.Class + ")"
}

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
	return g.Mint.GetCheckKind() == recordpb.CheckKind_CHECK_KIND_COMPUTATION, nil
}
