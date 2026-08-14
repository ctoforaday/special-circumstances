package lens

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// verify: red adjudicates ONE citation — which one, and what the source actually did for it.
//
// SPLIT FROM `cite` (#341), and the reason is the sharpest instance of facts-are-fields in the
// shipped tool. Blue authoring a citation and red verifying one shared the `cite` event type,
// told apart at read time by:
//
//	func IsVerifiedCite(e Event) bool { return e.Type == "cite" && e.Payload.Str("label") == "" }
//
// The distinction was the ABSENCE OF A FIELD. A blue cite written without a label counted as
// red's audit volume — a number red reads as how much work it did — with no error and no signal.
// Two acts, two event types now, and nothing infers which is which.
//
// # Three holes, one contract (0.60.0)
//
// The split fixed which ACT an event was and left the act itself unable to say anything
// definite. Measured on the shipped binary:
//
//	$ feov-record lens verify --run <dir> --seat-id red-lens-L1-r1
//	source verified:
//
// No flag was required. A verification of nothing, about nothing, recorded and counted.
//
//   - WHICH citation was unrecordable. `--reference` is free text a seat types, so nothing
//     joined a verification to the `<!--cite:c-…-->` anchor it checked, and `show evidence`
//     had to list red's work beside blue's sources without connecting them (#382).
//   - WHAT IT FOUND was unrecordable in the one direction that matters. The verdict field had
//     three values — high, medium, low — and no way to say the source does NOT hold up. So the
//     strongest finding on this axis had to leave as prose, and the capture audit built to catch
//     a report still carrying a refuted citation looked for a verdict no field could hold (#296).
//
// All three are the same defect wearing different clothes: the verb recorded that work
// happened rather than what the work concluded.
//
// # Two axes, and they are not the same question
//
// `--as` is WHAT THE SOURCE DID. `--confidence` is HOW SURE RED IS OF THAT. `refutes` at low
// confidence — this source may contradict the claim, I could not be certain — and `refutes` at
// high confidence are different facts, and a reader who cannot tell them apart cannot act on
// either. The plan specified exactly this: "for each statement ↔ reference pair it assigns a
// confidence that the source actually corroborates the statement (facts are rarely black and
// white); low confidence → needs more evidence, blue digs further, not an automatic fail".
//
// I COLLAPSED THEM AND WAS WRONG. The confidence field spent six releases named `--trust` — a
// rename made to dodge a collision with `blue confidence`, a verb deleted in 0.54.0 — and under
// that name its value descriptions drifted into a support scale ("the source supports the claim
// but you had to bridge something"). Read cold, it looked like an outcome enum with only its
// positive half, and folding it into `--as` looked like restoring a missing negative. It was not:
// it was deleting the orthogonal axis. The word is free again, so the field has its name back.
//
// # Why --independent rather than an optional --anchor
//
// Red verifies two different things: a citation BLUE authored (which has an anchor), and a
// source red went and found itself (corroboration, which has none). An optional --anchor
// collapses those into one shape where the absence means either "corroboration" or "I did not
// look it up" — the plausible zero again. So the anchor is REQUIRED, and the other case is
// stated with a flag of its own. `friction --none` is the same move for the same reason: an
// explicit negative is a fact; an empty field is a question.
func newVerify() *cobra.Command {
	c := seat.Prose(seat.New("verify",
		`adjudicate ONE citation: --anchor c-<hex> (from `+"`show evidence`"+`) or --independent, --claim "..." --as supports|refutes|absent|… --confidence high|medium|low --reason "<what the source actually says>"`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			anchor := strings.TrimSpace(seat.Str(cmd, flags.Anchor))
			independent, _ := cmd.Flags().GetBool(flags.Independent)

			switch {
			case anchor == "" && !independent:
				return nil, fmt.Errorf("lens verify requires --anchor <c-id> OR --independent: say WHICH citation you checked. Read the report, take the `<!--cite:c-…-->` token at the sentence, and resolve it with `lens show evidence --run <runDir>`. Pass --independent only for a source YOU found — corroboration blue never cited, which has no anchor to name")
			case anchor != "" && independent:
				return nil, fmt.Errorf("lens verify: --anchor and --independent are the two cases, not one. --anchor names a citation blue authored; --independent says this is a source you found yourself")
			}

			// THE SHAPE AND THE EXISTENCE ARE ON THE FLAG NOW. This block hand-rolled both — a
			// regexp for the c-<hex> form and a scan of CitationLabels — and both moved to the
			// typed flag (`flags.CitationAnchor().WithCheck(record.CitationExists)`), which
			// seat.Begin resolves before this handler runs. Keeping a second copy here would be
			// two readers of one rule, which is the thing this file's own header warns about.
			//
			// What stays is the part no flag can express: --anchor and --independent are the two
			// CASES, and exactly one must be chosen. That is a relation between flags, and a
			// pflag.Value sees only its own.

			p := seat.SetSame(cmd, record.NewPayload(), flags.Claim, flags.Reference)
			seat.Set(cmd, p, "access_date", flags.AccessDate)
			seat.Set(cmd, p, "outcome", flags.As)
			seat.Set(cmd, p, "confidence", flags.Confidence)
			if anchor != "" {
				p.Set("anchor", anchor)
			} else {
				p.Set("independent", true)
			}
			if err := seat.SetReason(cmd, p, "text"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.RunDir, s.SeatID, "verify", p); err != nil {
				return nil, err
			}
			return verifyResult{Anchor: anchor, Independent: independent,
				Outcome: p.Str("outcome"), Reference: seat.Str(cmd, flags.Reference)}, nil
		}))

	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists), flags.Anchor, "the c-<hex> of the citation you checked, from the report's `<!--cite:c-…-->` token — resolve it with `lens show evidence`. REQUIRED unless --independent")
	c.Flags().Bool(flags.Independent, false, "this is a source YOU found, not one blue cited, so there is no anchor to name. The explicit form of 'no anchor', because an empty field cannot say whether you looked")
	c.Flags().String(flags.Claim, "", "REQUIRED — the claim being verified, quoted from the report")
	c.Flags().String(flags.Reference, "", "the source the claim rests on (for an --independent check this is the only identification it has)")
	enumhelp.Flag(c, flags.As, record.MustEnum("verify", "outcome"), "REQUIRED — what the source ACTUALLY DID for the claim. It has a negative half: `refutes` and `absent` are findings, not failures to grade")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("verify", "confidence"), "REQUIRED — how sure you are of THAT determination, whichever it was. A separate question from --as: `refutes` you would defend and `refutes` you are unsure of are different facts")
	c.Flags().Var(&flags.DateValue{}, flags.AccessDate, "YYYY-MM-DD you actually fetched it; drives the staleness re-fetch trigger")
	return c
}

// citeAnchorShape is the citation id as it appears inside a `<!--cite:c-…-->` token. Checked
// before the record lookup so a seat that pasted the whole comment gets told what the id is,
// rather than being told the record has no such citation.
var citeAnchorShape = regexp.MustCompile(`^c-[0-9a-f]+$`)

type verifyResult struct {
	Anchor      string `json:"anchor,omitempty"`
	Independent bool   `json:"independent,omitempty"`
	Outcome     string `json:"outcome"`
	Reference   string `json:"reference,omitempty"`
}

func (r verifyResult) Human() string {
	subject := "citation " + r.Anchor
	if r.Independent {
		subject = "independent source " + r.Reference
	}
	switch r.Outcome {
	case "refutes":
		return subject + " REFUTES the claim — recorded. The report still carries it; that is a finding, and this event is the evidence for it"
	case "absent":
		return subject + " does NOT contain the claim — recorded as absent (silence, not contradiction)"
	case "unreachable":
		return subject + " could not be read — recorded as unreachable, with what you tried"
	default:
		return subject + " verified: " + r.Outcome
	}
}
