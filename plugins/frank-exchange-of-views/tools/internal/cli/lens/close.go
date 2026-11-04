package lens

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

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
	anchors := flags.IDList(flags.ClosureAnchorID)
	c := seat.New("close", seat.CloseResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
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
		// THE ANCHORS IT WAS VERIFIED WITH ARE A FIELD, each an id the write checks the record holds —
		// a report anchor or an inquiry: capture's attestation audit joins each to the other tool
		// calls of the closer's sitting, and an id read back out of --verified-with's words would be
		// a guess at what the seat meant.
		p.Anchors = anchors.Value()
		// A `SOURCE` OR `COMPUTATION` CHECK IS A DEMAND FOR A LIVE ANSWER, NOT A LABEL
		// (recordpb.DemandClause). A closure of such a gap with nothing meeting the demand would be
		// red accepting prose for the kind of evidence it declared prose cannot settle — the round the
		// 2026-08-05 smoke spent: R1-2 asked blue to "test it on a false claim", blue ASSERTED the test
		// had happened, and red's R2-2 refused it as "no evidence shown". It is asked under every
		// disposition, a close to a successor included, and before the record's failed-check refusal,
		// so a close the demand admits can still be refused by a check that stands.
		//
		// IT IS ASKED HERE, IN THE VERB, AND NOT IN record.validate: replay writes through the record
		// and never passes this verb, so every archived close replays as it was admitted, and `carry`
		// restates a closure the record already holds, which no open gap has.
		//
		// AN UNREADABLE RECORD REFUSES THE CLOSURE (gblock, 2026-10-23). A closure admitted
		// because the check could not be asked is a closure the tool vouched for on nothing.
		if err := requireEvidenceAnswered(seat.Str(cmd, flags.ID),
			func(id string) (record.EvidenceDemand, error) { return record.EvidenceDemandOf(run, id) }); err != nil {
			return nil, err
		}
		if _, err := record.Append(s.Identity(), p); err != nil {
			return nil, err
		}
		return seat.CloseResult{GapID: seat.Str(cmd, flags.ID), Class: recordpb.Word(p.GetClosureClass())}, nil
	})

	seat.ClosureFlags(c)
	c.Flags().String(flags.VerifiedBy, "", "WHO verified it — the seat that read the evidence")
	flags.Text(c, flags.VerifiedWith, "WITH WHAT — the tool or command that showed it")
	c.Flags().String(flags.VerifiedAgainst, "", "AGAINST WHAT — the object itself (the stored computation, the cited bytes, the named path), never a stand-in")
	c.Flags().Var(anchors, flags.Anchor, "the `ids` you verified it with, comma-separated — each the id inside a `"+anchor.Token("C-…")+"`, `"+anchor.Token("K-…")+"`, `"+anchor.Token("F-…")+"` or `"+anchor.Token("G-…")+"` token in the report, or the token itself, or an inquiry's `Q-…` id, of an act on this record; never the gap you are closing. The bare id is "+flags.BareID("F-…")+": `--anchor F-…,C-…`, or a token whole, quoted for the shell: `--anchor '"+anchor.Token("F-…")+"'`. Capture's attestation audit holds each to your other tool calls this sitting — one that read or named it, not this close; a closure verified against no anchor names none")
	// ALL THREE OR NONE, said by cobra rather than by a refusal after the fact. `validate` reads
	// them as one fact (`anchored`), so two of three was never a partial closure — it was an
	// unanchored one that spent the seat's turn before saying so.
	c.MarkFlagsRequiredTogether(flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst)
	seat.Require(c, flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst)
	// THE ARGUMENT IS UNCONDITIONAL FOR THIS VERB, AND CONDITIONAL FOR THE MESSAGE — the same
	// split `blue inquiry propose` makes for --question, and for the same reason.
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

// requireEvidenceAnswered refuses a closure of a gap whose check kind demands a live answer
// (recordpb.DemandClause) while nothing meets that demand, and any closure whose demand could not be
// read. The refusal names, by the kind's noun, every act that meets the demand on this gap.
func requireEvidenceAnswered(gapID string, demandOf func(string) (record.EvidenceDemand, error)) error {
	d, err := demandOf(gapID)
	if err != nil {
		return unreadableRefusal("the live-answer demand on "+gapID, err)
	}
	if d.AnswerKind == "" || d.Answered {
		return nil
	}
	return feov.Errorf(feov.Conflict, "%s", demandRefusal(gapID, d))
}

// demandRefusal is the close's refusal of gap gapID, whose demand d nothing meets: the demand in its
// one text, then what meets it on this gap — blue's anchor placed as its answer, a retire of each
// evidence anchor the gap reaches, and the check again by the seat whose check the gap carries — and
// the bench, for a demand the lens no longer holds.
func demandRefusal(gapID string, d record.EvidenceDemand) string {
	ways := []string{fmt.Sprintf("blue's `%s --answers %s`, a %s placed on the sentence as its answer (after a reword, on the reworded sentence)", d.AnswerVerb, gapID, d.AnswerNoun)}
	var retire []string
	if d.About != "" {
		retire = append(retire, fmt.Sprintf("blue's `retire --anchor %s`, the %s it is about", d.About, anchor.Kind(d.About)))
	}
	for _, r := range d.Readings {
		retire = append(retire, fmt.Sprintf("blue's `retire --anchor %s`, the citation whose reading by %s it carries", r.Anchor, r.Seat))
	}
	if len(retire) > 0 {
		ways = append(ways, strings.Join(retire, ", or "))
	}
	// THE CHECK AGAIN, BY ITS OWN SEAT, SINCE THE MINT, of an anchor of the demanded kind: a reading
	// that backs under `source`, a `sound` judgement under `computation`. Another seat's check meets
	// nothing.
	again := func(id, seat string) string {
		if anchor.Kind(id) != d.AnswerKind {
			return ""
		}
		if d.AnswerKind == "computation" {
			return fmt.Sprintf("a `sound` judgement of %s by %s, made since the gap was minted", id, seat)
		}
		return fmt.Sprintf("a reading of %s by %s, made since the gap was minted, that reads the work and backs the sentence", id, seat)
	}
	if d.About != "" {
		if a := again(d.About, d.MintedBy); a != "" {
			ways = append(ways, a)
		}
	}
	for _, r := range d.Readings {
		if a := again(r.Anchor, r.Seat); a != "" {
			ways = append(ways, a)
		}
	}
	return fmt.Sprintf("gap %s was minted --check-kind %s, and nothing meets its live-answer demand, so this close is refused under every disposition — `moot`, `not_a_defect` and a close to a successor included. Nothing is recorded. "+
		"The demand: %s. What meets it on this gap: %s; another seat's check meets nothing. "+
		"A demand you no longer hold is not closed: file `motion docket file` on the gap, and the bench's closing ruling, which is not held to the demand, disposes of it; a regrade keeps it open",
		gapID, d.CheckKind, recordpb.DemandClause, strings.Join(ways, "; or "))
}

// unreadableRefusal is the refusal a guard returns when the record it asks cannot be read: the
// act is not admitted on a check nobody could make, and the read error is named.
func unreadableRefusal(what string, err error) error {
	return feov.Wrap(feov.Internal, err, "the record could not be read to establish %s, so nothing is recorded — retry the call: %v", what, err)
}
