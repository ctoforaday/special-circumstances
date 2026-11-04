package lens

import (
	"errors"
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
// # One verb, two subjects
//
// Red reads a source for a sentence of the report in two ways: it adjudicates a citation BLUE
// authored, named by --anchor, or it reads a source against a sentence that carries no citation of
// it — one it found, or one blue cites elsewhere — named by --url and --title with the sentence as --quote. Both write the same Verify with the same axes
// and the same gap flags, and the record already splits on the field — a reading with no anchor is a
// reading of a source the lens found — so the verb splits there too. The contract differs only in
// WHICH flag names the subject, and cobra states that at the parse, before the handler: exactly one
// of --anchor and --url (MarkFlagsOneRequired, MarkFlagsMutuallyExclusive), --url with its --title and
// --quote (MarkFlagsRequiredTogether), which together leave --anchor with neither, and no --url with a
// --page. Nothing is policed in the handler; a violated group is refused with each flag's usage
// (seat.RefuseFlagGroup).
//
// # A reading that finds a defect mints its gap
//
// A reading of the work that finds a citation's source against its sentence, and a reading of a
// source the lens found that backs an uncited sentence or refutes one, mints a gap of a real class on
// that sentence in its own act (record.AppendReading) — the gap flags are `mint`'s (gapFlags), required
// exactly where the stored reading's ReadingGapClasses is non-empty and refused where it is empty. No
// reading places a citation: a source enters the report only by blue's cite.

// verifyAnchorAdvice is what a lens naming no citation is told: --anchor adjudicates a citation
// already on the record, so the act is to read the real label, not to create one.
var verifyAnchorAdvice = "--anchor names the citation you checked, so it is read off the record, not chosen: take the id from the report's `" + anchor.Token("C-…") + "` token beside the claim"

// gapFlagsRequiredUnless is the condition under which a reading's seven required gap flags may be
// left off: the reading mints no gap.
// No backquotes: pflag reads a backquoted word in a usage as the flag's value placeholder.
const gapFlagsRequiredUnless = "unless this reading reads nothing of the work, backs a citation's sentence, or is absent on a source you found"

// readingRequiredGapFlags are the gap flags a reading that mints requires, in the order a refusal
// names the first missing one; `--fix`, `--new` and `--complexity` are optional, as on `mint`.
var readingRequiredGapFlags = []string{flags.Class, flags.Problem, flags.Check, flags.CheckKind, flags.Severity, flags.Likelihood, flags.Impact}

// readingGapFlags is every gap flag, required or not: a reading that mints nothing takes none of them.
var readingGapFlags = append(slices.Clone(readingRequiredGapFlags), flags.Fix, flags.New, flags.Complexity)

func newVerify() *cobra.Command {
	var gf gapFlags
	c := seat.Prose(seat.New("verify", verifyResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		if id := strings.TrimSpace(seat.Str(cmd, flags.Anchor)); id != "" {
			body := &recordpb.Verify{Anchor: proto.String(id)}
			claim, err := citedSentence(s, id)
			if err != nil {
				return nil, err
			}
			body.Claim = proto.String(claim)
			if err := checkedPage(s, cmd, body); err != nil {
				return nil, err
			}
			source, err := citedURL(s, id)
			if err != nil {
				return nil, err
			}
			return writeVerify(s, cmd, &gf, body, source)
		}
		url := seat.Str(cmd, flags.URL)
		return writeVerify(s, cmd, &gf, &recordpb.Verify{
			Url:   proto.String(url),
			Title: proto.String(seat.Str(cmd, flags.Title)),
			Claim: proto.String(seat.Location(cmd)),
		}, url)
	}))

	const oneOf = "Give exactly one of --anchor (a citation blue made) and --url (a source read against a sentence that carries no citation of it: one you found, or one blue cites elsewhere)"
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists(verifyAnchorAdvice)), flags.Anchor, "the id of the citation you checked, from the report's `"+anchor.Token("C-…")+"` token — resolve it with `show evidence`. The sentence holding that token is the claim recorded. "+oneOf)
	c.Flags().String(flags.URL, "", flags.DescURL+". The source read for the sentence --quote names. "+oneOf)
	flags.Text(c, flags.Title, "the source's name, kept on your reading beside its url — it names the source to whoever answers the gap your reading mints; it reaches no report text")
	flags.Text(c, flags.Quote, flags.DescQuote+". With --url only: the REPORT's sentence your source backs or refutes, under every --as value, never the source's")
	c.Flags().Int(flags.Page, 0, "for a citation with pages: the page whose image you checked, drawn first with render-page. Refused on a citation with none, and with --url")
	c.MarkFlagsOneRequired(flags.Anchor, flags.URL)
	c.MarkFlagsMutuallyExclusive(flags.Anchor, flags.URL)
	c.MarkFlagsMutuallyExclusive(flags.URL, flags.Page)
	c.MarkFlagsRequiredTogether(flags.URL, flags.Title, flags.Quote)
	// A reading of a citation is keyed on the citation it names; a reading of a source the lens found
	// takes the ordinal key, and the record holds it to one reading per url and sentence a sitting
	// (record.requireNotARetriedReading).
	seat.KeysOn(c, "anchor")
	seat.Supplies(c, "claim", "the sentence holding the --anchor token in the report as it stands (citedSentence); on a source you found, your --quote")
	seat.Supplies(c, "gap_id", "the tool mints the id of the gap a reading mints (record.NewID), at random")
	seat.Supplies(c, "class_material", "stamped on the reading's gap from its class's material default at the write path")
	seat.Supplies(c, "carried_by", "the tool names the gap that carries a reading that mints: its own act's, or the open gap of its class on the sentence")
	seat.Supplies(c, "carrier_minted", "the tool records whether the reading's own act minted the gap that carries it")
	verifyAxes(c)
	gf.register(c, readingSubject)
	for _, f := range readingRequiredGapFlags {
		seat.RequiredUnless(c, f, gapFlagsRequiredUnless)
	}
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
	report, err := record.RenderRun(run)
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

// citedURL is the url of the citation id names — blue's cite, the only act that places one — which is
// the source a reading of it reads.
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
	if cite == nil {
		return feov.Errorf(feov.Conflict, "citation %s names no cite on the record, so there is no source to have read", body.GetAnchor())
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

// verifyAxes are the flags every reading takes: what the source DID, how sure red is of that, and
// when it was read.
func verifyAxes(c *cobra.Command) {
	enumhelp.Flag(c, flags.As, record.MustEnum("verify", "outcome"), "what the source ACTUALLY DID for the claim. It has a negative half — refutes and absent — recorded here, not left as prose: a reading of the work that finds a citation's source against its sentence mints that sentence's gap in this act, as a reading of a source you found does where it backs an uncited sentence or refutes one. Absent is refused on a copy fetch recorded as the work's abstract, or as not the work")
	enumhelp.Flag(c, flags.Bridge, record.MustEnum("verify", "bridge"), "WHAT a supports_with_bridge reading bridged — required with that outcome, refused with any other. A reading that needed a term moved to another object, a condition dropped or a relation reversed bridged nothing — record refutes or weak")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("verify", "confidence"), "how sure you are of THAT determination, whichever it was. A separate question from --as")
	c.Flags().Var(&flags.DateValue{}, flags.AccessDate, "YYYY-MM-DD you actually read it; drives the staleness re-fetch trigger")
}

// writeVerify records a reading. The caller has put the subject into the payload: an anchor and the
// sentence it marks, or the source it read and the sentence it quoted. source is the url the reading
// was made of: the citation's for --anchor, the typed one for --url.
func writeVerify(s seat.Context, cmd *cobra.Command, gf *gapFlags, body *recordpb.Verify, source string) (seat.Result, error) {
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
	// about the copy — whether it can establish absence, which part of the work it is, whose bytes it
	// is — is read off this one entry.
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
		// work at all, nothing was read that could say so. Read from the index, never fetched.
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
	// a verdict on one of those is a verdict on that copy.
	kind, copyURL := copyRead(held, cached)
	body.CopySubstitution, body.CopyUrl = kind.Enum(), copyURL

	// WHICH PART OF THE WORK RED'S COPY IS, on every reading — the cited url's copy for --anchor,
	// --url's for --url. A `supports` read off an abstract is a confirmation of the abstract, and a
	// reading of a copy that is not the work reads nothing of it. Stamped BEFORE every decision below,
	// each of which reads the stamped body. Read from the index, never fetched; a source outside the
	// cache was never asked.
	completeness := recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED
	if cached {
		completeness = held.SourceCompleteness()
	}
	body.SourceCompleteness = completeness.Enum()

	// WHETHER THIS READING MINTS is the stored body's (record.ReadingGapClasses), never --as alone.
	classes := record.ReadingGapClasses(body)
	backing := body.GetAnchor() == "" && len(classes) > 0 && !record.FindsAgainst(body)
	// A BACKING READING OF A MARKED SENTENCE IS REFUSED FIRST: the gap flags would only meet this
	// refusal next. Asked of the report as rendered now, and again under the write lock.
	if backing {
		if err := markedSentenceOf(run, body.GetClaim()); err != nil {
			return nil, err
		}
	}
	if len(classes) == 0 {
		if given := givenGapFlags(cmd); len(given) > 0 {
			return nil, feov.Errorf(feov.Validation, "%s: %s, so nothing is recorded. Leave the gap flags off; the reading is recorded alone",
				strings.Join(given, ", "), mintsNothing(body))
		}
		return recordReading(s, run, body, nil, nil)
	}
	for _, f := range readingRequiredGapFlags {
		if !seat.Given(cmd, f) {
			return nil, feov.Errorf(feov.Validation, "--%s is REQUIRED: this reading mints its gap in the same act (%s, %s), and a gap is minted with its class, problem, acceptance check, check kind and grades. Nothing is recorded",
				f, recordpb.Word(body.GetOutcome()), formOf(body))
		}
	}
	gap, err := gf.fill(cmd, text, "")
	if err != nil {
		return nil, err
	}
	if !slices.Contains(classes, gap.GetClass()) {
		return nil, feov.Errorf(feov.Validation, "--class %s: a %s reading of %s mints its gap in %s — nothing is recorded",
			recordpb.Word(gap.GetClass()), recordpb.Word(body.GetOutcome()), formOf(body), classWords(classes))
	}
	return recordReading(s, run, body, gap, func() string { return seat.Str(cmd, flags.New) })
}

// recordReading writes the reading — with its gap's Mint and the Anchor that places it where gap is
// non-nil — and words what the record then holds.
//
// THE GAP IS DECIDED INSIDE THE WRITE. Where an open gap of the Mint's class already holds the
// sentence, the one-gap rule hands it back under the write lock and the reading lands alone, naming it
// (record.AppendReading); a gap read before the lock would be stale by the time the act lands.
//
// NO ESTOPPEL GUARD (gblock, 2026-10-31: "Exempt, stated"). A source that finds against the sentence is
// new evidence, not red relitigating its prescription, so a reading's gap opens also on text blue
// applied verbatim from a red --new; nothing is checked and nothing is logged.
func recordReading(s seat.Context, run record.Run, body *recordpb.Verify, gap *recordpb.Mint, fixNew func() string) (seat.Result, error) {
	var evs []*record.Event
	if gap == nil {
		// A reading that mints nothing places nothing. A --url reading's quote is still held to the
		// report the reading lands on: it is the sentence the source was read for, whatever it did. A
		// reading of a citation is handed no quote: the tool read its sentence off the anchor.
		quote := body.GetClaim()
		if body.GetAnchor() != "" {
			quote = ""
		}
		ev, err := seat.AppendPlaced(s, body, "", quote, claimRefusal)
		if err != nil {
			return nil, err
		}
		evs = []*record.Event{ev}
	} else {
		id := record.NewID("gap")
		gap.GapId = proto.String(id)
		gap.Location = proto.String(body.GetClaim())
		place := &recordpb.Anchor{Id: proto.String(id)}
		at := quotePlace(body.GetClaim())
		var guards []record.Guard
		if of := body.GetAnchor(); of != "" {
			place.SentenceOf = proto.String(of)
			at = sentenceOfPlace(body.GetClaim(), of)
			guards = append(guards, seat.PlacesToSentenceOf(id, of, sentenceOfRefusal))
		} else {
			place.Location = proto.String(body.GetClaim())
			guards = append(guards, seat.Places(id, body.GetClaim(), claimRefusal))
			if !record.FindsAgainst(body) {
				guards = append(guards, uncitedSentence(body.GetClaim()))
			}
		}
		if n := fixNew(); n != "" {
			guards = append(guards, record.CheckOnReport(func(l *record.Landing) error {
				report, err := l.Report()
				if err != nil {
					return err
				}
				return verifiedFix(gap, report, id, at, n)
			}))
		}
		var err error
		if evs, err = record.AppendReading(s.Identity(), body, gap, place, guards...); err != nil {
			return nil, err
		}
	}
	stored, ok := recordpb.BodyAs[*recordpb.Verify](evs[len(evs)-1])
	if !ok {
		return nil, feov.Errorf(feov.Internal, "the reading was written as %T, not a verification", evs[len(evs)-1].GetBody())
	}
	res := verifyResult{
		Anchor:             stored.GetAnchor(),
		URL:                stored.GetUrl(),
		Source:             stored.GetTitle(),
		Outcome:            recordpb.Word(stored.GetOutcome()),
		Bridge:             bridgeWord(stored),
		Confidence:         recordpb.Word(stored.GetConfidence()),
		Claim:              stored.GetClaim(),
		SourceCompleteness: recordpb.Word(stored.GetSourceCompleteness()),
		CopySubstitution:   recordpb.Word(stored.GetCopySubstitution()),
		CopyURL:            stored.GetCopyUrl(),
		reads:              record.VerifyReads(stored),
	}
	// THE GAP IS THE RECORD'S, read back after the write: set exactly when the stored reading names
	// one. Its class is the Mint's the reading stated — the one-gap rule names a standing gap only of
	// that class — and whether this act minted it is the stored carrier_minted, never the verb's guess.
	if g := stored.GetCarriedBy(); g != "" {
		res.Gap = &readingGap{ID: g, Class: recordpb.Word(gap.GetClass()), Minted: stored.GetCarrierMinted()}
		if res.Gap.Minted {
			if m, ok := recordpb.BodyAs[*recordpb.Mint](evs[0]); ok {
				res.Gap.Check = m.GetAcceptanceCheck()
				for _, t := range []string{m.GetProblem(), m.GetRequiredFix(), m.GetFixNew()} {
					for _, f := range reportvoice.Advised(t) {
						res.Gap.VoiceTells = append(res.Gap.VoiceTells, f.String())
					}
				}
			}
		}
	}
	// WHAT THE RECORD NOW HOLDS OF THE CITATION, read back after every reading of one: each failed check
	// standing on its anchor, worded by its state — the seat is told what holds the PASS and whose act
	// clears it here rather than left to find it in its work list.
	if id := stored.GetAnchor(); id != "" {
		bs, err := record.PassBlockers(run)
		if err != nil {
			return nil, err
		}
		for _, b := range bs {
			if b.Kind != record.BlockerFailedCheck || b.Subject != id {
				continue
			}
			// THIS READING'S OWN FINDING is worded by its Gap sentence. A seat's later reading that reads
			// replaces its own check (the anchor ledger), so the found-against check this seat owns after
			// a reading of its own that finds against is the one this reading raised.
			if b.CheckState() == record.CheckFoundAgainst && b.Owner == s.SeatID && record.FindsAgainst(stored) {
				continue
			}
			res.FailedChecks = append(res.FailedChecks, failedCheckWords{
				State: string(b.CheckState()), Owner: b.Owner, Since: b.Since(), AnsweredBy: b.AnsweredBy(),
				seat: s.SeatID, reads: res.reads, backs: res.reads && !record.FindsAgainst(stored),
			})
		}
	}
	return res, nil
}

// givenGapFlags is each gap flag the call passed, as a refusal names it.
func givenGapFlags(cmd *cobra.Command) []string {
	var given []string
	for _, f := range readingGapFlags {
		if seat.Given(cmd, f) {
			given = append(given, "--"+f)
		}
	}
	return given
}

// mintsNothing is why a reading whose ReadingGapClasses is empty mints no gap.
func mintsNothing(v *recordpb.Verify) string {
	switch {
	case v.GetOutcome() == recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE:
		return "an unreachable source was not read and mints no gap"
	case !record.VerifyReads(v) && v.GetAnchor() != "":
		return "this copy is not the work (`not_the_work`), so the reading reads nothing of it and mints no gap — the citation stands unread, and the gap owed is about its anchor"
	case !record.VerifyReads(v):
		return "this copy is not the work (`not_the_work`), so the reading reads nothing of it and mints no gap"
	case v.GetAnchor() == "":
		return "a reading of a source you found that is `absent` records the reading and mints no gap"
	}
	return "a reading that backs the sentence mints nothing"
}

// formOf names which reading v is, for a refusal.
func formOf(v *recordpb.Verify) string {
	if v.GetAnchor() != "" {
		return "citation " + v.GetAnchor()
	}
	return "a source you found"
}

// classWords lists the classes a reading admits.
func classWords(classes []recordpb.GapClass) string {
	var names []string
	for _, c := range classes {
		names = append(names, recordpb.Word(c))
	}
	return strings.Join(names, " | ")
}

// markedSentenceOf is the marked-sentence guard asked of the report as rendered now, before the gap
// flags are: a backing reading of a source the lens found on a sentence an evidence anchor marks is
// refused with the act that reads that sentence's evidence.
func markedSentenceOf(run record.Run, quote string) error {
	report, err := record.RenderRun(run)
	if err != nil {
		return feov.Wrap(feov.Conflict, err, "the report does not render, so whether the sentence --quote names is cited cannot be told: %v", err)
	}
	return refuseMarked(report, quote)
}

// uncitedSentence is the marked-sentence guard asked of the report the reading lands on, where it is
// authoritative under the write lock: a citation placed on the sentence in between refuses the act.
func uncitedSentence(quote string) record.Guard {
	return record.CheckOnReport(func(l *record.Landing) error {
		report, err := l.Report()
		if err != nil {
			return err
		}
		return refuseMarked(report, quote)
	})
}

// refuseMarked refuses a backing reading of a source the lens found whose quote's sentence holds an
// evidence anchor — a citation or a computation token, read as anchors, never searched as text.
//
// WHY REFUSED, NOT RECORDED: such a reading mints `unsourced_claim`, true only of a sentence no
// evidence marks; on a marked one the class is false, and a reading that mints nothing tells no reader
// anything the sentence's own evidence does not. A quote the report does not hold once is the
// placement guard's to refuse, in its words.
func refuseMarked(report, quote string) error {
	_, end, err := anchortext.LocateOnce(report, quote, anchortext.StopAtParagraph)
	if err != nil {
		return nil
	}
	for _, sp := range anchor.Sentences(report) {
		if !(sp[0] < end && end <= sp[1]) {
			continue
		}
		var marked string
		anchor.Each(report[sp[0]:sp[1]], func(_, _ int, id string) {
			if marked == "" && anchor.Backs(id) {
				marked = id
			}
		})
		if marked == "" {
			return nil
		}
		act := "read that citation with `--anchor " + marked + "`"
		if anchor.Kind(marked) != "citation" {
			act = "its evidence is that " + anchor.Kind(marked) + ", which `reproduce` checks"
		}
		return feov.Errorf(feov.Validation, "this sentence is marked by %s; %s. A reading of a source you found that backs a sentence mints unsourced_claim, which is true only of a sentence no citation or computation marks. Nothing is recorded",
			anchor.Token(marked), act)
	}
	return nil
}

// bridgeWord is the bridge a reading names, "" where it names none.
func bridgeWord(body *recordpb.Verify) string {
	if body.Bridge == nil {
		return ""
	}
	return recordpb.Word(body.GetBridge())
}

type verifyResult struct {
	Anchor string `json:"anchor,omitempty"`
	// URL and Source are the source a reading of a source the lens found read, by url and title.
	URL     string `json:"url,omitempty"`
	Source  string `json:"source,omitempty"`
	Outcome string `json:"outcome"`
	// Bridge is what a supports_with_bridge reading bridged; absent on every other outcome.
	Bridge string `json:"bridge,omitempty"`
	// Confidence is the confidence the reading records, read off the stored event.
	Confidence string `json:"confidence"`
	// Claim is the sentence the reading sits on: for a reading of a citation, the one the tool read
	// off the anchor; for a source the lens found, its --quote.
	Claim string `json:"claim"`
	// SourceCompleteness is which part of the work the copy read was, as stamped on the stored reading.
	SourceCompleteness string `json:"source_completeness,omitempty"`
	// Gap is the gap that carries this reading — the one its act minted, or the open gap of its class
	// the sentence already held; absent on a reading that mints none.
	Gap *readingGap `json:"gap,omitempty"`
	// CopySubstitution is whose bytes the reading was made against, as the record holds them, and
	// CopyURL where that copy was read, when a copy stood in for the url.
	CopySubstitution string `json:"copy_substitution"`
	CopyURL          string `json:"copy_url,omitempty"`
	// FailedChecks are the failed checks standing on the citation's anchor after a reading of it, each
	// worded by its state; this reading's own found-against check is its Gap's.
	FailedChecks []failedCheckWords `json:"failed_checks,omitempty"`
	// reads is record.VerifyReads of the stored reading: a reading that reads nothing is worded so
	// before its outcome is.
	reads bool
}

// readingGap is the gap a reading's act names, with every field of mintResult a reading's gap has.
type readingGap struct {
	ID     string `json:"id"`
	Class  string `json:"class"`
	Minted bool   `json:"minted"`
	// Check is the acceptance check as the stored Mint records it, echoed for the reason mintResult's
	// is: a shell may rewrite it before the tool sees it. Absent where the reading named a standing
	// gap, whose check this act did not write.
	Check      string   `json:"acceptance_check,omitempty"`
	VoiceTells []string `json:"voice_tells,omitempty"`
}

// words is the sentence naming the gap, with a minted gap's check echo and voice note.
func (g *readingGap) words() string {
	switch {
	case g == nil:
		return ""
	case g.Minted:
		return ". This reading mints gap " + g.ID + " (class " + g.Class + ")" + mintEcho(g.Check, g.VoiceTells)
	}
	return ". Gap " + g.ID + ", open on this sentence with class " + g.Class + ", carries this reading; no second gap is minted"
}

// failedCheckWords is a record.Blocker failed check standing on a citation, as the verb prints it.
type failedCheckWords struct {
	State      string `json:"state"`
	Owner      string `json:"owner"`
	Since      int64  `json:"since"`
	AnsweredBy string `json:"answered_by,omitempty"`
	// seat is the reading's own seat; reads and backs are whether this reading read the work, and
	// whether it backs the sentence.
	seat         string
	reads, backs bool
}

// words is the sentence for one failed check on citation anchor, chosen by its state.
func (f failedCheckWords) words(anchor string) string {
	if f.State == string(record.CheckFoundAgainst) {
		head := "\ncitation " + anchor
		if !f.reads {
			head = "\nthis reading reads nothing of the source, and citation " + anchor
		}
		out := head + " stays found against: " + f.Owner + "'s reading at event " + strconv.FormatInt(f.Since, 10) +
			" finds the source against this sentence, and PASS is refused while the report carries it"
		if f.AnsweredBy != "" {
			out += "; gap " + f.AnsweredBy + " carries it to blue, which edits the sentence or retires or replaces the citation"
		} else {
			out += "; no gap names it, and blue edits the sentence or retires or replaces the citation"
		}
		if f.backs && f.Owner != f.seat {
			out += ". Your backing reading beside " + f.Owner + "'s standing one is a disagreement and clears nothing"
		}
		return out
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

func (r verifyResult) Human() string {
	out := r.human() + seat.CopyWords(r.CopySubstitution, r.CopyURL) + r.Gap.words()
	for _, f := range r.FailedChecks {
		out += f.words(r.Anchor)
	}
	return out
}

// human words a reading by whether it read the work, and only then by its outcome: a `refutes` of a
// copy that is not the work refutes nothing.
func (r verifyResult) human() string {
	subject := "citation " + r.Anchor
	if r.Anchor == "" {
		subject = "source " + r.Source + " (" + r.URL + ")"
	}
	claim := `"` + r.Claim + `"` // unescaped: the sentence as the report holds it
	if r.Outcome == "unreachable" {
		return subject + " could not be read — recorded as unreachable, with what you tried; it mints nothing"
	}
	if !r.reads {
		return "this reading reads nothing of the source — " + subject + " is recorded as " + r.Outcome + " on a copy that is not the work (" + r.SourceCompleteness + ") and mints nothing, on the claim " + claim
	}
	switch r.Outcome {
	case "refutes":
		return subject + " REFUTES the claim " + claim + " — recorded"
	case "absent":
		return subject + " does NOT contain the claim " + claim + " — recorded as absent (the source says nothing about it; it does not say otherwise)"
	case "supports_with_bridge":
		return subject + " verified: " + r.Outcome + " (bridged: " + r.Bridge + "), on the claim " + claim
	default:
		return subject + " verified: " + r.Outcome + ", on the claim " + claim
	}
}

// claimRefusal is the refusal `verify --url` gives for a --quote Attach will not place. One wording
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

// sentenceOfRefusal is the refusal a reading of a citation meets where its gap's anchor will not stand
// at the end of the sentence the citation's token sits in, on the report the act lands on.
func sentenceOfRefusal(err error) error {
	switch {
	case errors.Is(err, anchortext.ErrAnchorGone):
		return feov.Errorf(feov.Conflict, "the citation's anchor left the report before this reading landed — blue retired its claim in between — so there is no sentence for its gap. Nothing is recorded: %v", err)
	case errors.Is(err, anchortext.ErrAnchorBare):
		return feov.Errorf(feov.Conflict, "the citation's anchor stands bare in the report — the sentence it backed was cut — so there is no sentence for its gap. Nothing is recorded: %v", err)
	}
	return feov.Wrap(feov.Validation, err, "the gap this reading mints cannot stand at the end of the citation's sentence: %v. Nothing is recorded", err)
}
