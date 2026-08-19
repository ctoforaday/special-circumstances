package lens

import (
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
//   - WHICH citation was unrecordable. The source was free text a seat typed, so nothing joined
//     a verification to the `<!--cite:c-…-->` anchor it checked, and `show evidence` had to list
//     red's work beside blue's sources without connecting them (#382).
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
// # Why `corroborate` is a VERB and not a flag on this one
//
// Red does two different things with a source: it adjudicates a citation BLUE authored (which
// has an anchor to name), and it corroborates from a source red went and found itself (which has
// no anchor, and must therefore name the source outright). Those are different contracts —
// `--anchor` is required for one and meaningless for the other, `--url`/`--title` the reverse —
// and one verb could hold both only by policing the combination in its own body and leaving
// every flag optional to cobra.
//
// That is the shape this suite exists to remove, and `motion` already states the rule: cobra
// cannot express "required only when", so a mode-discerned verb puts its contract into
// hand-written validation where nothing can refuse it at parse. As two verbs, each marks what it
// genuinely requires and cobra refuses the nonsense before the handler runs.
func newVerify() *cobra.Command {
	c := seat.Prose(seat.New("verify",
		`adjudicate ONE citation blue authored: --anchor c-<hex> (from `+"`show evidence`"+`) --quote "..." --as supports|refutes|absent|… --confidence high|medium|low --reason "<what the source actually says>". `+
			`THIS VERB JUDGES A CITATION THAT EXISTS. A claim carrying NO citation at all is not verified as `+"`absent`"+` — `+
			"`absent` means you read blue's source and the claim is not in it. An unevidenced claim is not this verb's, and it is not "+
			"automatically a finding either: if a source exists and you can reach it, `corroborate` is how you go and get it, and it "+
			"answers whether the claim is true in the WORLD. `finding` is for when there is nothing to fetch, and answers whether the "+
			`TEXT stands up.`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := record.NewPayload().Set("anchor", strings.TrimSpace(seat.Str(cmd, flags.Anchor)))
			return writeVerify(s, cmd, p)
		}))

	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists), flags.Anchor, "the c-<hex> of the citation you checked, from the report's `<!--cite:c-…-->` token — resolve it with `lens show evidence`")
	_ = c.MarkFlagRequired(flags.Anchor)
	verifyAxes(c)
	return c
}

// corroborate: red reading a source IT found, for a claim blue made.
//
// The mirror of `verify` and not a mode of it. There is no anchor to name — blue never cited
// this — so the source is named the way every source in this system is named, by --url and
// --title, and both are required here because nothing else identifies what red read.
func newCorroborate() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("corroborate",
		`corroborate a claim from a source YOU found — one blue never cited, so there is no anchor: --url <u> --title <t> --quote "..." --as supports|refutes|absent|… --confidence high|medium|low --reason "<what the source actually says>". `+
			`THIS IS THE VERB FOR A CLAIM WITH NO CITATION when the source is obtainable — it answers whether the claim is true in the `+
			`WORLD, where a finding answers whether the TEXT stands up, and it is what makes "nobody cited this" checkable instead of `+
			"merely raised. To adjudicate a citation blue DID author, use `verify` instead: it names the citation by its anchor.",
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			p := record.NewPayload().Set("independent", true)
			seat.SetSame(cmd, p, flags.URL, flags.Title)
			return writeVerify(s, cmd, p)
		}), "verify"))

	c.Flags().String(flags.URL, "", flags.DescURL)
	c.Flags().String(flags.Title, "", flags.DescTitle)
	_ = c.MarkFlagRequired(flags.URL)
	_ = c.MarkFlagRequired(flags.Title)
	verifyAxes(c)
	return c
}

// verifyAxes are the flags both verbs share: WHAT was checked, what the source DID, how sure red
// is of that, and when it was read. Registered from one place so the two verbs cannot drift into
// describing the same four fields differently — which is how this vocabulary got into trouble.
func verifyAxes(c *cobra.Command) {
	c.Flags().String(flags.Quote, "", "REQUIRED — "+flags.DescQuote+" (the claim you are checking)")
	enumhelp.Flag(c, flags.As, record.MustEnum("verify", "outcome"), "REQUIRED — what the source ACTUALLY DID for the claim. It has a negative half: `refutes` and `absent` are findings, not failures to grade")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("verify", "confidence"), "REQUIRED — how sure you are of THAT determination, whichever it was. A separate question from --as: `refutes` you would defend and `refutes` you are unsure of are different facts")
	c.Flags().Var(&flags.DateValue{}, flags.AccessDate, "YYYY-MM-DD you actually read it; drives the staleness re-fetch trigger")
}

// writeVerify records what both verbs agree on. The caller has already put the fact that
// distinguishes them — an anchor, or the source it read — into the payload.
func writeVerify(s seat.Context, cmd *cobra.Command, p *record.Payload) (seat.Result, error) {
	seat.Set(cmd, p, "claim", flags.Quote)
	seat.Set(cmd, p, "access_date", flags.AccessDate)
	seat.Set(cmd, p, "outcome", flags.As)
	seat.Set(cmd, p, "confidence", flags.Confidence)
	if err := seat.SetReason(cmd, p, "reason"); err != nil {
		return nil, err
	}
	if _, err := record.Append(s.Identity(), "verify", p); err != nil {
		return nil, err
	}
	return verifyResult{Anchor: p.Str("anchor"), Source: p.Str("title"), Outcome: p.Str("outcome")}, nil
}

// citeAnchorShape is the citation id as it appears inside a `<!--cite:c-…-->` token. Checked
// before the record lookup so a seat that pasted the whole comment gets told what the id is,
// rather than being told the record has no such citation.
var citeAnchorShape = regexp.MustCompile(`^c-[0-9a-f]+$`)

type verifyResult struct {
	Anchor  string `json:"anchor,omitempty"`
	Source  string `json:"source,omitempty"`
	Outcome string `json:"outcome"`
}

func (r verifyResult) Human() string {
	subject := "citation " + r.Anchor
	if r.Anchor == "" {
		subject = "corroborating source " + r.Source
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
