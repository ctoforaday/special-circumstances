package lens

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
			"`absent` means you read the source and the claim is not in it. An UNEVIDENCED claim is a finding: raise it with `finding`, "+
			`which is the channel for "this assertion rests on nothing", exactly as it is for any other defect in the text.`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			return writeVerify(s, cmd, &recordpb.Verify{
				Anchor: proto.String(strings.TrimSpace(seat.Str(cmd, flags.Anchor))),
			}, adjudates)
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
			`To adjudicate a citation blue DID author, use `+"`verify`"+` instead: it names the citation by its anchor.`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			return writeVerify(s, cmd, &recordpb.Verify{
				Independent: proto.Bool(true),
				Url:         proto.String(seat.Str(cmd, flags.URL)),
				Title:       proto.String(seat.Str(cmd, flags.Title)),
			}, cites)
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
// citing says whether this verb may turn a supporting read into a FOOTNOTE.
//
// `corroborate` may: red found the source, blue never cited it, and a human reader cares that
// the text has appropriate references rather than which team inserted them. `verify` may not —
// it adjudicates a citation blue ALREADY made, which already has an anchor and a footnote; a
// second one would double the same source in the bibliography.
type citing bool

const (
	cites     citing = true
	adjudates citing = false
)

// backsTheClaim reports whether an outcome POINTS AT A SOURCE the reader can go and read, and so
// belongs in the bibliography.
//
// `weak` IS INCLUDED, and the argument is symmetry rather than strength. When BLUE cites a source
// and red grades it `weak` through `lens verify`, the footnote STAYS — verify adjudicates, it
// never touches the report. So excluding a weak corroboration would render the same (claim,
// source, grade) triple as a footnote when blue found the source and as NOTHING when red did,
// which is exactly the difference a reader should never be able to see. Held out, red's reading
// also reached no reader at all and carried no duty: the silent middle state.
//
// A footnote is a POINTER, not an endorsement. It says "this source bears on this sentence" —
// which is true of thin support — and red's judgement of how well it bears is preserved where
// judgements live, in the evidence view, exactly as it is for a blue cite graded weak.
//
// `refutes` and `absent` stay out because they are not pointers to supporting material: a source
// that contradicts the sentence, rendered in the bibliography, reads as backing it, and the
// report's assembly check already treats a live refuted citation as a failure. They owe a
// finding instead. `unreachable` stays out on its own merits — red could not read the thing, so
// there is nothing to point the reader at.
func backsTheClaim(o recordpb.SourceOutcome) bool {
	return o == recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS ||
		o == recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE ||
		o == recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK
}

func writeVerify(s seat.Context, cmd *cobra.Command, body *recordpb.Verify, mayCite citing) (seat.Result, error) {
	body.Claim = proto.String(seat.Str(cmd, flags.Quote))
	// ABSENT IS NOT EMPTY. Set unconditionally, an unpassed --access-date landed as "" and read
	// as "the seat dated this source to nothing" rather than "the seat did not date it" — the
	// same defect that made `merge close` write `successor = ''` and fail every closure once the
	// reference existed. Here nothing refuses it, so it simply reached every reader as a fact.
	body.AccessDate = seat.OptStr(cmd, flags.AccessDate)
	// BOTH WORDS ARE REFUSED RATHER THAN ZEROED. `refutes` and `absent` are the negative half this
	// axis was widened to carry, and the zero is UNSPECIFIED — a mistyped verdict recorded as the
	// zero is a citation nobody graded, which reads exactly like one nobody checked.
	if w := seat.Str(cmd, flags.As); w != "" {
		o, ok := record.SourceOutcomeOf(w)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "lens verify: %q is not a source outcome this record can carry", w)
		}
		body.Outcome = &o
	}
	if w := seat.Str(cmd, flags.Confidence); w != "" {
		cf, ok := record.ConfidenceOf(w)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "lens verify: %q is not a confidence this record can carry", w)
		}
		body.Confidence = &cf
	}
	text, err := seat.Reason(cmd)
	if err != nil {
		return nil, err
	}
	body.Text = proto.String(text)

	// THE FOOTNOTE, FOR A SOURCE THAT BACKS THE CLAIM. Minted and spliced BEFORE the append, in
	// the order `blue cite` uses and for the same reason: the label forms the marker, so it must
	// exist before the report write, and a rejected splice must leave no event behind.
	//
	// This is also what moves the event's key off the URL. `keyFields` is walked first-match and
	// `label` sits before `url`, so a labelled corroboration keys on the minted id — which is why
	// one source may now corroborate several claims. Keyed on the source, only the first recorded.
	if mayCite == cites && backsTheClaim(body.GetOutcome()) {
		// A RETRY RETURNS ITS OWN ANCHOR. The minted label is fresh every call, so without this
		// a crash-retried corroboration splices a SECOND anchor at the same sentence and records
		// a second event — the duplication the url key used to prevent. Checked before the mint,
		// as blue's cite checks its --key before the fetch.
		if prior, err := record.ExistingCorroborationLabel(s.RunDir, s.SeatID, body.GetUrl(), body.GetClaim()); err != nil {
			return nil, err
		} else if prior != "" {
			return verifyResult{Label: prior, Source: body.GetTitle(), Outcome: recordpb.Word(body.GetOutcome()), Idempotent: true}, nil
		}
		label := record.NewCitationID()
		marker := "<!--cite:" + label + "-->"
		if err := record.MutateBlueReport(s.RunDir, func(old []byte) ([]byte, error) {
			next, aerr := InsertAnchor(old, body.GetClaim(), marker)
			switch {
			case errors.Is(aerr, ErrMisQuote):
				return nil, feov.Errorf(feov.Validation,
					"lens corroborate: the quoted claim was not found in report.md — quote the EXACT sentence you are corroborating (via --quote); the whole string is matched, so a heading prepended to it matches nothing. A corroboration of a claim blue has since edited away is not spliced blind")
			case errors.Is(aerr, ErrInFence):
				return nil, feov.Errorf(feov.Validation, "lens corroborate: the quote resolves inside a code fence — corroborate a prose sentence, not code")
			}
			return next, aerr
		}); err != nil {
			return nil, err
		}
		body.Label = proto.String(label)
	}

	if _, err := record.Append(s.Identity(), body); err != nil {
		return nil, err
	}
	return verifyResult{
		Anchor:  body.GetAnchor(),
		Label:   body.GetLabel(),
		Source:  body.GetTitle(),
		Outcome: recordpb.Word(body.GetOutcome()),
	}, nil
}

// citeAnchorShape is the citation id as it appears inside a `<!--cite:c-…-->` token. Checked
// before the record lookup so a seat that pasted the whole comment gets told what the id is,
// rather than being told the record has no such citation.
var citeAnchorShape = regexp.MustCompile(`^c-[0-9a-f]+$`)

type verifyResult struct {
	Anchor     string `json:"anchor,omitempty"`
	Label      string `json:"label,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
	Source     string `json:"source,omitempty"`
	Outcome    string `json:"outcome"`
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
