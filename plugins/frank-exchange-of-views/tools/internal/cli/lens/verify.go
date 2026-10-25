package lens

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
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
// verifyAnchorAdvice is what a lens naming no citation is told: --anchor adjudicates a citation
// already on the record, so the act is to read the real label, not to create one.
var verifyAnchorAdvice = "--anchor names the citation you checked, so it is read off the record, not chosen: take the id from the report's `" + anchor.Token("C-…") + "` token beside the claim"

func newVerify() *cobra.Command {
	c := seat.Prose(seat.New("verify", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		body := &recordpb.Verify{Anchor: proto.String(strings.TrimSpace(seat.Str(cmd, flags.Anchor)))}
		claim, err := citedSentence(s, body.GetAnchor())
		if err != nil {
			return nil, err
		}
		body.Claim = proto.String(claim)
		if err := checkedPage(s, cmd, body); err != nil {
			return nil, err
		}
		source, err := citedURL(s, body.GetAnchor())
		if err != nil {
			return nil, err
		}
		return writeVerify(s, cmd, body, source, adjudates)
	}))

	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists(verifyAnchorAdvice)), flags.Anchor, "the id of the citation you checked, from the report's `"+anchor.Token("C-…")+"` token — resolve it with `show evidence`. The sentence holding that token is the claim recorded")
	seat.Require(c, flags.Anchor)
	// A verification is keyed on the citation it names.
	seat.KeysOn(c, "anchor")
	seat.Supplies(c, "claim", "the sentence holding the --anchor token in the report as it stands (citedSentence)")
	c.Flags().Int(flags.Page, 0, "for a citation with pages: the page whose image you checked, drawn first with render-page. Refused on a citation with none and on a corroboration, which holds none")
	verifyAxes(c)
	return c
}

// citedSentence is the claim a verification of citation id records: the sentence holding its anchor
// in the report as it stands, read the way an edit reads it to reopen the citation
// (bluedoc.SentenceAround). The lens names the citation and the tool reads the sentence, so a
// verdict cannot sit on a sentence the citation does not mark. A citation with no such sentence —
// its anchor out of the report, or bare, or the report unrendered — has no claim to read a source
// against, and is refused.
func citedSentence(s seat.Context, id string) (string, error) {
	run, err := s.Run()
	if err != nil {
		return "", err
	}
	report, err := reportproj.RenderFromRecord(run)
	if err != nil {
		return "", feov.Wrap(feov.Conflict, err, "the report does not render, so the sentence citation %s marks cannot be read, and nothing can be verified until it does: %v. "+
			"The board's anomalies say why; record that you could not verify it with `log`", id, err)
	}
	sentence, ok := bluedoc.SentenceAround(report, anchor.Token(id))
	switch {
	case !ok:
		return "", feov.Errorf(feov.Conflict, "citation %s's anchor is not in the report — blue retired its claim, or the anchor never placed — so the report holds no sentence for it and there is nothing to verify. "+
			"Verify the citations whose anchors the report carries", id)
	case slices.Contains(claimcount.BareAnchorIDs(report), id):
		return "", feov.Errorf(feov.Conflict, "citation %s's anchor stands bare in the report: the sentence it backed was cut, and it backs no prose, so there is nothing to verify. "+
			"The anchor leaves with blue's retire of its claim", id)
	}
	return sentence, nil
}

// citedURL is the url of the citation id names — blue's cite or a labelled corroboration — which
// is the source a verification of it reads.
func citedURL(s seat.Context, id string) (string, error) {
	run, err := s.Run()
	if err != nil {
		return "", err
	}
	sources, err := record.CitedSources(run)
	if err != nil {
		return "", err
	}
	for _, src := range sources {
		if src.Label == id {
			return src.URL, nil
		}
	}
	return "", feov.Errorf(feov.Conflict, "citation %s names no source on the record, so there is no source to have read", id)
}

// copyRead is whose bytes a reading was made against: the entry the run's fetch cache resolves the
// source's url to when the reading is written — the entry every read of that url through the tool
// returns — with the url it was fetched from where those bytes are not the url's own. A cache
// holding no copy is not_cached: red read the source outside the tool, or the cache that held it is
// gone, and the tool does not say which bytes those were.
func copyRead(e fetchcache.Entry, cached bool) (recordpb.CopySubstitution, *string) {
	if !cached {
		return recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_CACHED, nil
	}
	if e.Substitution != "" && e.CopyURL != "" {
		return e.CopySubstitution(), proto.String(e.CopyURL)
	}
	return e.CopySubstitution(), nil
}

// checkedPage holds a verification of a citation with pages to the page image red checked.
//
// A CITATION WITH PAGES QUOTES OCR TEXT, which can misread, so reading the reading again proves
// nothing: the check is against the pixels, and the record says which page image it was. A page
// given on a citation with none has no referent, and a page outside the citation's pages is a
// check of something else. `unreachable` may omit the page — red could not draw or read the
// image, and a verification that cannot be written is worse than one saying nothing was checked —
// but a page it names must still have its image on disk.
func checkedPage(s seat.Context, cmd *cobra.Command, body *recordpb.Verify) error {
	run, err := s.Run()
	if err != nil {
		return err
	}
	cite, err := record.CiteByLabel(run, body.GetAnchor())
	if err != nil {
		return err
	}
	given := seat.Given(cmd, flags.Page)
	// A CORROBORATION HOLDS NO PAGES. Its quote is the report's sentence, never a span located in
	// an OCR reading, so a page names no image the corroboration was read from — and admitted, the
	// flag would be dropped with nothing recorded, which is the silent no-op `fetch --at` refuses.
	if cite == nil {
		if given {
			return feov.Errorf(feov.Validation, "--page names a page image, and %s is a corroboration — its quote is the report's sentence, not a span located in an OCR reading, so there is no image to check", body.GetAnchor())
		}
		return nil
	}
	pages := cite.GetPages()
	page, _ := cmd.Flags().GetInt(flags.Page)
	if len(pages) == 0 {
		if given {
			return feov.Errorf(feov.Validation, "--page names a page image, and citation %s has no pages — its text is not a quote located in an OCR reading, so there is no image to check", body.GetAnchor())
		}
		return nil
	}
	sha := cite.GetSha256()
	if !given {
		if seat.Str(cmd, flags.As) == recordpb.Word(recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE) {
			return nil
		}
		return feov.Errorf(feov.Validation, "citation %s quotes OCR text found on PDF %s, and the reading may be what is wrong — draw one of those pages with `render-page --sha %s --page %d`, check the image, and pass that page as --page", body.GetAnchor(), pageList(pages), sha, pages[0])
	}
	if !containsPage(pages, page) {
		return feov.Errorf(feov.Validation, "--page %d is not a page citation %s's quote was found on (PDF %s) — check one of those", page, body.GetAnchor(), pageList(pages))
	}
	img, err := os.ReadFile(fetchcache.PageImagePath(run, sha, page))
	if err != nil {
		return feov.Errorf(feov.Validation, "no image of page %d is on disk — draw it first with `render-page --sha %s --page %d`", page, sha, page)
	}
	rec, had, err := fetchcache.ReadReadingRecord(run, sha)
	if err != nil {
		return err
	}
	if !had || page > len(rec.RenderShas) {
		return feov.Errorf(feov.Validation, "the reading citation %s was located in is gone from the cache — `fetch --url %s` re-derives it", body.GetAnchor(), cite.GetUrl())
	}
	body.Page = proto.Int32(int32(page))
	body.PageRenderSha = proto.String(fetchcache.Sha(img))
	body.ReadingRenderSha = proto.String(rec.RenderShas[page-1])
	return nil
}

func containsPage(pages []int32, page int) bool {
	for _, p := range pages {
		if int(p) == page {
			return true
		}
	}
	return false
}

// pageList renders pages the way the report prints them: "p. 10" or "pp. 10, 34".
func pageList(pages []int32) string {
	parts := make([]string, len(pages))
	for i, p := range pages {
		parts[i] = strconv.Itoa(int(p))
	}
	if len(pages) == 1 {
		return "p. " + parts[0]
	}
	return "pp. " + strings.Join(parts, ", ")
}

// corroborate: red reading a source IT found, for a claim blue made.
//
// The mirror of `verify` and not a mode of it. There is no anchor to name — blue never cited
// this — so the source is named the way every source in this system is named, by --url and
// --title, and both are required here because nothing else identifies what red read.
//
// It exists so that "nobody cited this" is checkable instead of merely raised. A soft tell in a
// title ("this run", "epoch 3") is flagged and not refused because a real source's title can use
// those words too.
func newCorroborate() *cobra.Command {
	c := seat.Prose(seat.Records(seat.New("corroborate", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		url := seat.Str(cmd, flags.URL)
		return writeVerify(s, cmd, &recordpb.Verify{
			Url:   proto.String(url),
			Title: proto.String(seat.Str(cmd, flags.Title)),
			Claim: proto.String(seat.Location(cmd)),
		}, url, cites)
	}), "verify"))

	c.Flags().String(flags.URL, "", flags.DescURL)
	flags.Text(c, flags.Title, flags.DescTitle)
	// The flag names whose sentence the quote is because, beside `refutes`, the sentence that does
	// the refuting is the source's, and writeVerify refuses it.
	flags.Text(c, flags.Quote, flags.DescQuote+". It is the REPORT's sentence under every --as value, never the source's")
	seat.Require(c, flags.URL, flags.Title)
	verifyAxes(c)
	// A corroboration is keyed on the source it read — or, when the source backs the claim, on the
	// label of the citation the tool then places (writeVerify).
	seat.KeysOn(c, "url", "label")
	return c
}

// verifyAxes are the flags both verbs share: what the source DID, how sure red is of that, and
// when it was read. Registered from one place so the two verbs cannot drift into describing the
// same three fields differently — which is how this vocabulary got into trouble.
func verifyAxes(c *cobra.Command) {
	enumhelp.Flag(c, flags.As, record.MustEnum("verify", "outcome"), "what the source ACTUALLY DID for the claim. It has a negative half: refutes and absent are findings, not failures to grade. Absent is refused on a copy fetch recorded as the work's abstract, or as not the work")
	enumhelp.Flag(c, flags.Bridge, record.MustEnum("verify", "bridge"), "WHAT a supports_with_bridge reading bridged — required with that outcome, refused with any other. A reading that needed a term moved to another object, a condition dropped or a relation reversed bridged nothing — record refutes or weak")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("verify", "confidence"), "how sure you are of THAT determination, whichever it was. A separate question from --as")
	c.Flags().Var(&flags.DateValue{}, flags.AccessDate, "YYYY-MM-DD you actually read it; drives the staleness re-fetch trigger")
}

// writeVerify records what both verbs agree on. The caller has already put the facts that
// distinguish them into the payload: an anchor and the sentence it marks, or the source it read and
// the claim it quoted. source is the url the reading was made of: the citation's for `verify`, the
// typed one for `corroborate`.
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

func writeVerify(s seat.Context, cmd *cobra.Command, body *recordpb.Verify, source string, mayCite citing) (seat.Result, error) {
	run, err := s.Run()
	if err != nil {
		return nil, err
	}
	// ABSENT IS NOT EMPTY. Set unconditionally, an unpassed --access-date landed as "" and read
	// as "the seat dated this source to nothing" rather than "the seat did not date it" — the
	// same defect that made `lens close` write `successor = ''` and fail every closure once the
	// reference existed. Here nothing refuses it, so it simply reached every reader as a fact.
	body.AccessDate = seat.OptStr(cmd, flags.AccessDate)
	// THE COPY THIS RUN HOLDS OF THE SOURCE, read once from the index and never fetched: a write verb
	// reaching the network would turn an unreachable host into a refused reading. Every fact below
	// about the copy — whether it can establish absence, what the literature says of the work, which
	// part of the work it is, whose bytes it is — is read off this one entry.
	held, _, cached, lerr := fetchcache.Lookup(run, source)
	if lerr != nil {
		return nil, unreadableRefusal("which copy of "+source+" this run holds", lerr)
	}
	// BOTH WORDS ARE REFUSED RATHER THAN ZEROED. `refutes` and `absent` are the negative half this
	// axis was widened to carry, and the zero is UNSPECIFIED — a mistyped verdict recorded as the
	// zero is a citation nobody graded, which reads exactly like one nobody checked.
	if w := seat.Str(cmd, flags.As); w != "" {
		o, ok := record.SourceOutcomeOf(w)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "%q is not a source outcome this record can carry", w)
		}
		body.Outcome = &o
		// SILENCE IN AN ABSTRACT IS NOT SILENCE IN THE WORK. `absent` says the claim is not in the
		// source; where the copy this run holds is recorded as the work's abstract, or as not the
		// work at all, nothing was read that could say so. Read from the index, never fetched —
		// the same rule the retraction stamp below keeps.
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT {
			if e := held; cached && e.WithoutBody() {
				return nil, feov.Errorf(feov.Validation, "the copy of %s this run holds is %s — %s. "+
					"Its silence is not the work's, so absent cannot be established from it; record what the copy does "+
					"establish, or find a copy that carries the body", source, e.Completeness, e.CompletenessReason)
			}
		}
	}
	if w := seat.Str(cmd, flags.Bridge); w != "" {
		k, ok := record.BridgeKindOf(w)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "%q is not a bridge this record can carry: summary | secondary_citation | near_restatement", w)
		}
		body.Bridge = &k
	}
	if w := seat.Str(cmd, flags.Confidence); w != "" {
		cf, ok := record.ConfidenceOf(w)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "%q is not a confidence this record can carry", w)
		}
		body.Confidence = &cf
	}
	text, err := seat.Reason(cmd)
	if err != nil {
		return nil, err
	}
	body.Text = proto.String(text)

	// WHOSE BYTES THE READING WAS MADE AGAINST. The seat names a source; the cache may hold the
	// full text its page pointed to, a recovered copy, a fuller one or a named backend's answer, and
	// a verdict on one of those is a verdict on that copy. Stamped before the retry check, which
	// holds a retry to the write's rules.
	kind, copyURL := copyRead(held, cached)
	body.CopySubstitution, body.CopyUrl = kind.Enum(), copyURL

	// THE FOOTNOTE, FOR A SOURCE THAT BACKS THE CLAIM. Minted and spliced BEFORE the append, in
	// the order `blue cite` uses and for the same reason: the label forms the marker, so it must
	// exist before the report write, and a rejected splice must leave no event behind.
	//
	// This is also what moves the event's key off the URL. `keyFields` is walked first-match and
	// `label` sits before `url`, so a labelled corroboration keys on the minted id — which is why
	// one source may now corroborate several claims. Keyed on the source, only the first recorded.
	var tells []string
	if mayCite == cites && backsTheClaim(body.GetOutcome()) {
		// WHAT THE LITERATURE SAYS ABOUT THE WORK RED FOUND. A labelled corroboration becomes an
		// ordinary footnote, so it owes the reader the same stamp a blue cite does — and this is
		// the worst case the field exists for: the strongest thing an adversarial process
		// produces, a claim confirmed by a source its author never chose, resting on withdrawn
		// work with nothing in the document saying so.
		//
		// READ FROM THE INDEX, NEVER FETCHED. Red is corroborating a source it has read, so the
		// run cache already holds it; a source outside the cache stamps NOT_CHECKED, which is the
		// honest answer and not a reassurance. A fetch here would make a write verb reach the
		// network and turn an unreachable host into a refused corroboration.
		if cached {
			body.WorkStatus = record.WorkStatusOf(held.Retraction()).Enum()
		} else {
			body.WorkStatus = recordpb.WorkStatus_WORK_STATUS_NOT_CHECKED.Enum()
		}
		// THE TITLE IS RED'S TEXT IN THE REPORT. A labelled corroboration's --title prints in the
		// source's note, "[^N]: <title>. <url> (accessed <date>)", where red's anchor is the first under
		// that note, and in the Bibliography line "- <title>. <url> (accessed <date>)" where no blue cite
		// names the URL (report/assemble.go weaveCitations, bibliographyEntry). The url and the date
		// are the source's identity; the claim is blue's sentence; the reason, outcome, confidence
		// and seat stay on the record. So the title is the one span of a corroboration that reaches
		// report.md, and it is held to the report's voice as a mint's problem, fix and prescription are:
		// the write path refuses the unambiguous tells (record.refuseCorroborationTitleVoice, which
		// reads the label this branch mints), and the ambiguous rest is advice (gblock, 2026-09-15).
		// Computed before the retry check, so a retry hears it.
		for _, f := range reportvoice.Advised(body.GetTitle()) {
			tells = append(tells, f.String())
		}
		// A RETRY RETURNS ITS OWN ANCHOR. The minted label is fresh every call, so without this
		// a crash-retried corroboration splices a SECOND anchor at the same sentence and records
		// a second event — the duplication the url key used to prevent. Checked before the mint,
		// as blue's cite checks its --key before the fetch.
		if prior, err := record.ExistingCorroboration(run, s.SeatID, body.GetUrl(), body.GetClaim()); err != nil {
			return nil, err
		} else if prior != nil {
			// A RETRY IS THE SAME ACT, held to every rule the first write was and to the reading it
			// repeats: a retry the write would refuse is refused, and one that reads the source
			// differently is a second reading, which this seat records by verifying the citation.
			// What comes back is the STORED row, never the request.
			if err := record.WouldRefuse(run, s.SeatID, recordpb.EventType_EVENT_TYPE_VERIFY, body); err != nil {
				return nil, err
			}
			if !sameReading(prior, body) {
				return nil, feov.Errorf(feov.Conflict, "you corroborated %s on this claim already, as citation %s: %s. A corroboration is one per source and claim, and its retry repeats it; a different reading of the same source is a verification of that citation — `verify --anchor %s`",
					body.GetUrl(), prior.GetLabel(), readingWords(prior), prior.GetLabel())
			}
			if was, now := statedOtherwise(prior, body); was != "" {
				return nil, feov.Errorf(feov.Conflict, "you corroborated %s on this claim already, as citation %s, stated with %s; this repeat states %s. "+
					"A corroboration is one per source and claim and its retry repeats it, so this one records nothing and is refused",
					body.GetUrl(), prior.GetLabel(), was, now)
			}
			// A corroboration on the record without its anchor — one migrate brought forward from an archive whose run crashed between the two — is finished by
			// the retry, at the claim it stored.
			if err := seat.PlaceOwed(s, prior.GetLabel(), ""); err != nil {
				return nil, err
			}
			return verifyResult{Label: prior.GetLabel(), Source: prior.GetTitle(), Outcome: recordpb.Word(prior.GetOutcome()), Bridge: bridgeWord(prior), Claim: prior.GetClaim(),
				CopySubstitution: recordpb.Word(prior.GetCopySubstitution()), CopyURL: prior.GetCopyUrl(), Idempotent: true, VoiceTells: tells}, nil
		}
		// The Anchor event appended after this one is the marker (it carries the claim and this
		// label); reportproj.Render re-places it on read. No file is spliced.
		body.Label = proto.String(record.NewID("citation"))
	}

	// WHICH PART OF THE WORK RED'S COPY IS, on every verdict that names a source — a `supports`
	// read off an abstract is a confirmation of the abstract, and the record says so rather than
	// the seat. Read from the index, never fetched; a source outside the cache was never asked.
	if body.GetUrl() != "" {
		completeness := recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED
		if cached {
			completeness = held.SourceCompleteness()
		}
		body.SourceCompleteness = completeness.Enum()
	}

	// A CORROBORATION'S QUOTE IS THE REPORT'S SENTENCE, WHATEVER THE SOURCE DID FOR IT. One
	// placement check, for every outcome, against the report the act and its Anchor land on; a
	// refusal records nothing. Where the source backs the claim, the quote is where the citation's
	// marker stands. Where the source contradicts the claim or is silent on it, the quote is the
	// claim of the contradiction the reading opens (record.contradictionsOf), and a claim the
	// report does not hold is a reading of nothing. A
	// verification places nothing and its claim needs no check: the tool read it from the report
	// (citedSentence), so it is handed no quote to hold.
	quote := body.GetClaim()
	if mayCite != cites {
		quote = ""
	}
	ev, err := seat.AppendPlaced(s, body, body.GetLabel(), quote, claimRefusal)
	if err != nil {
		return nil, err
	}
	// THE CONTRADICTION IS THE RECORD'S, read back after the write: only a finding naming its id
	// raises it, so the seat is told the id here rather than left to find it in its work list.
	c, err := record.ContradictionOfReading(run, ev)
	if err != nil {
		return nil, err
	}
	stored, ok := recordpb.BodyAs[*recordpb.Verify](ev)
	if !ok {
		return nil, feov.Errorf(feov.Internal, "the reading was written as %T, not a verification", ev.GetBody())
	}
	// WHAT THE RECORD NOW HOLDS OF THE CITATION, read back after the write: a reading that read
	// nothing leaves it on a failed check, and the seat is told whose gap it owes here rather than
	// left to find it in its work list.
	var failed *failedCheckWords
	if body.GetAnchor() != "" && !record.VerifyReads(stored) {
		bs, err := record.PassBlockers(run)
		if err != nil {
			return nil, err
		}
		for _, b := range bs {
			if b.Kind == record.BlockerFailedCheck && b.Subject == body.GetAnchor() {
				failed = &failedCheckWords{Owner: b.Owner, AnsweredBy: b.AnsweredBy(), seat: s.Identity().SeatID}
			}
		}
	}
	return verifyResult{
		Contradiction: contradictionOf(c),
		Anchor:        body.GetAnchor(),
		Label:         body.GetLabel(),
		Source:        body.GetTitle(),
		Outcome:       recordpb.Word(body.GetOutcome()),
		Bridge:        bridgeWord(body),
		Claim:         body.GetClaim(),
		VoiceTells:    tells,

		CopySubstitution: recordpb.Word(stored.GetCopySubstitution()),
		CopyURL:          stored.GetCopyUrl(),
		FailedCheck:      failed,
	}, nil
}

// sameReading reports whether a corroboration retry reads the source as the stored one did: the
// same outcome, bridge and confidence.
func sameReading(stored, retry *recordpb.Verify) bool {
	return stored.GetOutcome() == retry.GetOutcome() && bridgeWord(stored) == bridgeWord(retry) &&
		stored.GetConfidence() == retry.GetConfidence()
}

// statedOtherwise is what a corroboration retry states besides its reading — the title, the date it
// was read and its reason — where that differs from the stored corroboration, as the flags that
// state it then and now; "" where every one repeats.
func statedOtherwise(stored, retry *recordpb.Verify) (was, now string) {
	var w, n []string
	for _, f := range []struct{ flag, s, r string }{
		{flags.Title, stored.GetTitle(), retry.GetTitle()},
		{flags.AccessDate, stored.GetAccessDate(), retry.GetAccessDate()},
		{flags.Reason, stored.GetText(), retry.GetText()},
	} {
		if f.s == f.r {
			continue
		}
		w, n = append(w, statedAs(f.flag, f.s)), append(n, statedAs(f.flag, f.r))
	}
	return strings.Join(w, " "), strings.Join(n, " ")
}

func statedAs(flag, v string) string {
	if v == "" {
		return "no --" + flag
	}
	return fmt.Sprintf("--%s %q", flag, v)
}

// readingWords renders a stored reading for the refusal that names it.
func readingWords(v *recordpb.Verify) string {
	out := "--as " + recordpb.Word(v.GetOutcome())
	if b := bridgeWord(v); b != "" {
		out += " --bridge " + b
	}
	return out + " --confidence " + recordpb.Word(v.GetConfidence())
}

// bridgeWord is the bridge a reading names, "" where it names none.
func bridgeWord(body *recordpb.Verify) string {
	if body.Bridge == nil {
		return ""
	}
	return recordpb.Word(body.GetBridge())
}

type verifyResult struct {
	Anchor     string `json:"anchor,omitempty"`
	Label      string `json:"label,omitempty"`
	Idempotent bool   `json:"idempotent,omitempty"`
	Source     string `json:"source,omitempty"`
	Outcome    string `json:"outcome"`
	// Bridge is what a supports_with_bridge reading bridged; absent on every other outcome.
	Bridge string `json:"bridge,omitempty"`
	// Claim is the sentence the verdict sits on: for a verification, the one the tool read off the
	// anchor, which a contradiction the reading opens names as its claim.
	Claim string `json:"claim"`
	// VoiceTells is ADVICE on a labelled corroboration's title, which prints in the source's note and Bibliography entry.
	VoiceTells []string `json:"voice_tells,omitempty"`
	// Contradiction is the contradiction this reading opened or joined; absent on a reading that
	// opened and joined none.
	Contradiction *readingContradiction `json:"contradiction,omitempty"`
	// CopySubstitution is whose bytes the reading was made against, as the record holds them, and
	// CopyURL where that copy was read, when a copy stood in for the url.
	CopySubstitution string `json:"copy_substitution"`
	CopyURL          string `json:"copy_url,omitempty"`
	// FailedCheck is the citation's failed check as the record holds it after a verification that
	// read nothing — whose check it is and the gap that carries it, "" until one does; absent on a
	// reading, and on a corroboration, which names no citation.
	FailedCheck *failedCheckWords `json:"failed_check,omitempty"`
}

// failedCheckWords is record.Blocker's failed check as the verb prints it.
type failedCheckWords struct {
	Owner      string `json:"owner"`
	AnsweredBy string `json:"answered_by,omitempty"`
	seat       string
}

// words is the sentence the verb prints for a verification that left its citation unread: what the
// record now holds, and what is owed. "" where there is none.
func (f *failedCheckWords) words(anchor string) string {
	if f == nil {
		return ""
	}
	head := "\nthis reading read nothing of the source, so citation " + anchor + " stays unread: PASS is refused while the report cites it unread"
	switch {
	case f.AnsweredBy != "":
		return head + ", and gap " + f.AnsweredBy + " carries it to blue, which retires the citation or cites a source a lens can read"
	case f.Owner == f.seat:
		return head + ", and you owe a gap about it (`mint --about-kind anchor --about " + anchor + "`, which takes no mint of your budget) asking blue to retire it or cite a source a lens can read"
	}
	return head + ", and the gap about it is " + f.Owner + "'s to mint, whose check of it read nothing first"
}

// readingContradiction is record.ReadingContradiction as the verb prints it: the id as every
// surface spells it, the `--about` value of the finding that raises it.
type readingContradiction struct {
	ID       string `json:"id"`
	Opened   bool   `json:"opened"`
	OpenedBy string `json:"opened_by"`
}

func contradictionOf(c *record.ReadingContradiction) *readingContradiction {
	if c == nil {
		return nil
	}
	return &readingContradiction{ID: strconv.FormatInt(c.ID, 10), Opened: c.Opened, OpenedBy: c.Opener}
}

// words is the sentence that names the contradiction, "" where there is none.
func (c *readingContradiction) words() string {
	switch {
	case c == nil:
		return ""
	case c.Opened:
		return ". This reading opened contradiction " + c.ID
	default:
		return ". This reading joined contradiction " + c.ID + " (opened by " + c.OpenedBy + ")"
	}
}

func (r verifyResult) Human() string {
	return r.human() + seat.CopyWords(r.CopySubstitution, r.CopyURL) + r.Contradiction.words() + r.FailedCheck.words(r.Anchor) + reportvoice.Note("the source's note and Bibliography entry", r.VoiceTells)
}

func (r verifyResult) human() string {
	subject := "citation " + r.Anchor
	if r.Anchor == "" {
		subject = "corroborating source " + r.Source
	}
	claim := `"` + r.Claim + `"` // unescaped: the sentence as the report holds it
	switch r.Outcome {
	case "refutes":
		return subject + " REFUTES the claim " + claim + " — recorded. The report still carries it; that is a finding, and this event is the evidence for it"
	case "absent":
		return subject + " does NOT contain the claim " + claim + " — recorded as absent (the source says nothing about it; it does not say otherwise)"
	case "unreachable":
		return subject + " could not be read — recorded as unreachable, with what you tried"
	case "supports_with_bridge":
		return subject + " verified: " + r.Outcome + " (bridged: " + r.Bridge + "), on the claim " + claim
	default:
		return subject + " verified: " + r.Outcome + ", on the claim " + claim
	}
}

// claimRefusal is the refusal `corroborate` gives for a --quote Attach will not place. One wording
// for every outcome, so it says what the quote is under each.
func claimRefusal(err error) error {
	switch {
	case errors.Is(err, anchortext.ErrMisQuote):
		return feov.Errorf(feov.Validation,
			"--quote was not found in report.md. It is the REPORT's sentence — the claim you read the source against, in the report's exact words — whatever the source did for it: the sentence it backs, the sentence it contradicts, or the sentence it says nothing about. "+
				"What the SOURCE says is never the quote: its words go in --reason. "+
				"The whole string is matched, so a heading prepended to it matches nothing; a sentence blue has edited is quoted in its edited words, as `show report` prints it")
	case errors.Is(err, anchortext.ErrInFence):
		return feov.Errorf(feov.Validation, "the quote resolves inside a code fence — the claim a source is read against is a prose sentence, not code")
	}
	return anchortext.Refusal(err)
}
