package lens

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
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
	var ev CloseEvidence
	accept := flags.ProposalID()
	c := seat.New("close", seat.CloseResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		var p *recordpb.Close
		if m := accept.String(); m != "" {
			// THE RAISER TAKES THE AUTHOR'S PROPOSAL BY ITS OWN ACT: the tool supplies the close the
			// motion stores — its word and its verification — and the act carries accepts = M.
			// --reason is still the raiser's own argument.
			for _, f := range []string{flags.As, flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst, flags.Anchor} {
				if cmd.Flags().Changed(f) {
					return nil, feov.Errorf(feov.Validation, "--accept %s supplies the close motion %s proposes, its --as and its verification included — passing --%s as well would be a second source for one fact. Drop it, or drop --accept and close by your own judgment", m, m, f)
				}
			}
			stored, err := seat.AcceptedAct(s, m, seat.Str(cmd, flags.ID))
			if err != nil {
				return nil, err
			}
			cl, ok := stored.(*recordpb.Close)
			if !ok {
				return nil, feov.Errorf(feov.Validation, "motion %s proposes a regrade, which `regrade --accept %s` takes. Nothing was recorded", m, m)
			}
			prose, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			cl.Prose = proto.String(prose)
			cl.Accepts = proto.String(m)
			p = cl
		} else {
			if err := seat.RequireOwed(cmd); err != nil {
				return nil, err
			}
			p, err = seat.ClosurePayload(cmd)
			if err != nil {
				return nil, err
			}
			ev.fill(cmd, p)
		}
		// A `SOURCE` OR `COMPUTATION` CHECK IS A DEMAND FOR A LIVE ANSWER, NOT A LABEL
		// (recordpb.DemandClause). A closure of such a gap with nothing meeting the demand would be
		// red accepting prose for the kind of evidence it declared prose cannot settle — the round the
		// 2026-08-05 smoke spent: R1-2 asked blue to "test it on a false claim", blue ASSERTED the test
		// had happened, and red's R2-2 refused it as "no evidence shown". It is asked under every
		// disposition, a close to a successor included, and before the record's failed-check refusal,
		// so a close the demand admits can still be refused by a check that stands.
		//
		// IT IS ASKED HERE FIRST, SO IT REFUSES BEFORE THE RECORD'S SHAPE CHECKS, AND AGAIN IN THE
		// WRITE'S TRANSACTION (record.judgeAct), WHERE THE ANSWER HOLDS: a retire of the gap's only
		// answer can land between this read and the write. Neither is in record.validate: replay
		// never asks the act judgements, so every archived close replays as it was admitted, and
		// `carry` restates a closure the record already holds, which no open gap has.
		//
		// AN UNREADABLE RECORD REFUSES THE CLOSURE (gblock, 2026-10-23). A closure admitted
		// because the check could not be asked is a closure the tool vouched for on nothing.
		if err := WriteClose(s, run, p); err != nil {
			return nil, err
		}
		return seat.CloseResult{GapID: p.GetGapId(), Class: recordpb.Word(p.GetClosureClass()), Accepts: accept.String()}, nil
	})

	// --accept NAMES THE PROPOSAL, which names its gap, its word and its verification: each flag it
	// supplies is required only without it, and the handler refuses a close with neither.
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id")
	seat.RequiredUnlessGiven(c, flags.ID, flags.Accept)
	enumhelp.Flag(c, flags.As, record.MustEnum("close", "closure_class"), "HOW the gap ended — one vocabulary with every closure on the record")
	seat.RequiredUnlessGiven(c, flags.As, flags.Accept)
	c.Flags().Var(accept, flags.Accept, "take the author's proposed close of your gap EXACTLY as filed: the M- id its filing returned. The tool supplies its --as and its verification, and the act answers the proposal; under a docketing it is your concession, ending the docketing and the proposal with no ruling. --reason is still your own argument")
	ev.register(c, flags.Accept)
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

// CloseEvidence is a close's evidence as a verb takes it: the verification triple and the anchors it
// was verified with. `close` takes it, and so does `motion docket file close`, whose close the gap's
// raiser takes (--accept M) or the bench applies as it is stated.
type CloseEvidence struct{ anchors *flags.CSV }

// Register registers the evidence flags on c, each required.
func (ev *CloseEvidence) Register(c *cobra.Command) { ev.register(c, "") }

// register registers the evidence flags on c: required, or required under unless where it names the
// flag that supplies them (seat.RequiredUnless).
func (ev *CloseEvidence) register(c *cobra.Command, unless string) {
	ev.anchors = flags.IDList(flags.ClosureAnchorID)
	c.Flags().String(flags.VerifiedBy, "", "WHO verified it — the seat that read the evidence")
	flags.Text(c, flags.VerifiedWith, "WITH WHAT — the tool or command that showed it")
	c.Flags().String(flags.VerifiedAgainst, "", "AGAINST WHAT — the object itself (the stored computation, the cited bytes, the named path), never a stand-in")
	c.Flags().Var(ev.anchors, flags.Anchor, "the `ids` you verified it with, comma-separated — each the id inside a `"+anchor.Token("C-…")+"`, `"+anchor.Token("K-…")+"` or `"+anchor.Token("G-…")+"` token in the report, or the token itself, or an inquiry's `Q-…` id, of an act on this record; never the gap you are closing. The bare id is "+flags.BareID("C-…")+": `--anchor C-…,K-…`, or a token whole, quoted for the shell: `--anchor '"+anchor.Token("C-…")+"'`. Capture's attestation audit holds each to your other tool calls this sitting — one that read or named it, not this close; a closure verified against no anchor names none")
	// ALL THREE OR NONE, said by cobra rather than by a refusal after the fact. `validate` reads
	// them as one fact (`anchored`), so two of three was never a partial closure — it was an
	// unanchored one that spent the seat's turn before saying so.
	c.MarkFlagsRequiredTogether(flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst)
	if unless == "" {
		seat.Require(c, flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst)
		return
	}
	for _, f := range []string{flags.VerifiedBy, flags.VerifiedWith, flags.VerifiedAgainst} {
		seat.RequiredUnlessGiven(c, f, unless)
	}
}

// fill sets the evidence the flags state on p.
func (ev *CloseEvidence) fill(cmd *cobra.Command, p *recordpb.Close) {
	p.AnchorSeat = proto.String(seat.Str(cmd, flags.VerifiedBy))
	p.AnchorTool = proto.String(seat.Str(cmd, flags.VerifiedWith))
	p.AnchorTarget = proto.String(seat.Str(cmd, flags.VerifiedAgainst))
	// THE ANCHORS IT WAS VERIFIED WITH ARE A FIELD, each an id the write checks the record holds —
	// a report anchor or an inquiry: capture's attestation audit joins each to the other tool
	// calls of the closer's sitting, and an id read back out of --verified-with's words would be
	// a guess at what the seat meant.
	p.Anchors = ev.anchors.Value()
}

// Fill is fill, for another package's verb.
func (ev *CloseEvidence) Fill(cmd *cobra.Command, p *recordpb.Close) { ev.fill(cmd, p) }

// WriteClose is the close's write, whoever takes it: `close` itself, or the bench's ruling applying a
// proposed close, with before — the ruling — written ahead of it in the same transaction.
// The live-answer demand is asked first, and again in the write's transaction (record.judgeAct).
func WriteClose(s seat.Context, run record.Run, p *recordpb.Close, before ...record.Act) error {
	if err := requireEvidenceAnswered(p.GetGapId(),
		func(id string) (record.EvidenceDemand, error) { return record.EvidenceDemandOf(run, id) }); err != nil {
		return err
	}
	_, err := record.AppendActs(s.Identity(), append(before, record.Act{Body: p})...)
	return err
}

// requireEvidenceAnswered refuses a closure of a gap whose check kind demands a live answer
// (recordpb.DemandClause) while nothing meets that demand, and any closure whose demand could not be
// read. The refusal names, by the kind's noun, every act that meets the demand on this gap.
func requireEvidenceAnswered(gapID string, demandOf func(string) (record.EvidenceDemand, error)) error {
	d, err := demandOf(gapID)
	if err != nil {
		return unreadableRefusal("the live-answer demand on "+gapID, err)
	}
	if !d.Unmet() {
		return nil
	}
	return record.DemandRefusal(gapID, d)
}

// unreadableRefusal is the refusal a guard returns when the record it asks cannot be read: the
// act is not admitted on a check nobody could make, and the read error is named.
func unreadableRefusal(what string, err error) error {
	return feov.Wrap(feov.Internal, err, "the record could not be read to establish %s, so nothing is recorded — retry the call: %v", what, err)
}
