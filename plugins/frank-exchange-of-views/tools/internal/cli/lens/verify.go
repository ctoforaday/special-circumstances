package lens

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
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
// # A reading records; a gap holds
//
// A reading records what the source did for the sentence and holds nothing by itself (gblock,
// 2026-11-06, #1727). A defect it shows is a gap `mint` opens, standing on the reading: a citation's
// `refutes` or `absent` is the ground of a gap minted about the citation, a found source's `refutes`
// the ground of a gap quoting the sentence (record.readingMintedClasses). A reading that finds against
// is its lens's DEBT until that gap or a re-read of the subject (record.readingDebtsOf): it holds the
// PASS as the lens's unfinished promise, on the chair's list with the lens's name, and the lens's
// envelope is refused while it stands. No reading places a citation: a source enters the report only
// by blue's cite.

// verifyAnchorAdvice is what a lens naming no citation is told: --anchor adjudicates a citation
// already on the record, so the act is to read the real label, not to create one.
var verifyAnchorAdvice = "--anchor names the citation you checked, so it is read off the record, not chosen: take the id from the report's `" + anchor.Token("C-…") + "` token beside the claim"

func newVerify() *cobra.Command {
	c := seat.Prose(seat.New("verify", verifyResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		if id := strings.TrimSpace(seat.Str(cmd, flags.Anchor)); id != "" {
			body := &recordpb.Verify{Anchor: proto.String(id)}
			run, err := s.Run()
			if err != nil {
				return nil, err
			}
			// The sentence the reading is OF: the one the seat judged, read before the write. The reading
			// lands only while the anchor still backs that sentence (record.CheckOfSentence).
			claim, err := record.LiveSentence(run, id)
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
			return writeVerify(s, cmd, body, source)
		}
		url := seat.Str(cmd, flags.URL)
		return writeVerify(s, cmd, &recordpb.Verify{
			Url:   proto.String(url),
			Title: proto.String(seat.Str(cmd, flags.Title)),
			Claim: proto.String(seat.Location(cmd)),
		}, url)
	}))

	const oneOf = "Give exactly one of --anchor (a citation blue made) and --url (a source read against a sentence that carries no citation of it: one you found, or one blue cites elsewhere)"
	c.Flags().Var(flags.CitationAnchor().WithCheck(record.CitationExists(verifyAnchorAdvice)), flags.Anchor, "the id of the citation you checked, from the report's `"+anchor.Token("C-…")+"` token — resolve it with `show evidence`. The sentence holding that token is the claim recorded. "+oneOf)
	c.Flags().String(flags.URL, "", flags.DescURL+". The source read for the sentence --quote names. "+oneOf)
	flags.Text(c, flags.Title, "the source's name, kept on your reading beside its url; a gap you mint on this reading names the source through it. It reaches no report text")
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
	seat.Supplies(c, "claim", "the sentence holding the --anchor token in the report as it stands (record.LiveSentence); on a source you found, your --quote")
	verifyAxes(c)
	return c
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
	enumhelp.Flag(c, flags.As, record.MustEnum("verify", "outcome"), "what the source ACTUALLY DID for the claim. It has a negative half — refutes and absent — recorded here, not left as prose: a reading of the work that finds the source against its sentence is the ground of the gap you mint on it, and your debt until that gap stands. Absent is refused on a copy fetch recorded as the work's abstract, or as not the work")
	enumhelp.Flag(c, flags.Bridge, record.MustEnum("verify", "bridge"), "WHAT a supports_with_bridge reading bridged — required with that outcome, refused with any other. A reading that needed a term moved to another object, a condition dropped or a relation reversed bridged nothing — record refutes or weak")
	enumhelp.Flag(c, flags.Confidence, record.MustEnum("verify", "confidence"), "how sure you are of THAT determination, whichever it was. A separate question from --as")
	c.Flags().Var(&flags.DateValue{}, flags.AccessDate, "YYYY-MM-DD you actually read it; drives the staleness re-fetch trigger")
}

// writeVerify records a reading. The caller has put the subject into the payload: an anchor and the
// sentence it marks, or the source it read and the sentence it quoted. source is the url the reading
// was made of: the citation's for --anchor, the typed one for --url.
func writeVerify(s seat.Context, cmd *cobra.Command, body *recordpb.Verify, source string) (seat.Result, error) {
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

	// A BACKING READING OF A MARKED SENTENCE IS REFUSED: a found source backs only an unmarked
	// sentence (refuseMarked). An `absent` one is recorded, and owes nothing on a marked sentence
	// (record.readingDebtsOf). Asked of the report as rendered now, and again under the write lock.
	if body.GetAnchor() == "" && record.VerifyReads(body) && !record.FindsAgainst(body) {
		if err := markedSentenceOf(run, body.GetClaim()); err != nil {
			return nil, err
		}
		return recordReading(s, run, body, uncitedSentence(body.GetClaim()))
	}
	return recordReading(s, run, body)
}

// recordReading writes the reading alone and words what the record then holds.
func recordReading(s seat.Context, run record.Run, body *recordpb.Verify, guards ...record.Guard) (seat.Result, error) {
	// A READING OF A CITATION RECORDS THE SENTENCE THE SEAT JUDGED, read before the write. The landing
	// asks first, before every other guard, that the citation still backs that sentence
	// (record.CheckOfSentence): an edit or a retire committing since the read refuses the reading, so
	// the sentence recorded is the one the report holds at its event.
	var live []record.Guard
	if of := body.GetAnchor(); of != "" {
		live = append(live, record.CheckOfSentence(of, body.GetClaim()))
		if beforeAppend != nil {
			if err := beforeAppend(); err != nil {
				return nil, err
			}
		}
	}
	live = append(live, guards...)
	// A --url reading's quote is held to the report the reading lands on: it is the sentence the source
	// was read for, whatever it did. A reading of a citation is handed no quote: the tool read its
	// sentence off the anchor.
	quote := body.GetClaim()
	if body.GetAnchor() != "" {
		quote = ""
	}
	ev, err := seat.AppendPlaced(s, body, "", quote, claimRefusal, live...)
	if err != nil {
		return nil, err
	}
	stored, ok := recordpb.BodyAs[*recordpb.Verify](ev)
	if !ok {
		return nil, feov.Errorf(feov.Internal, "the reading was written as %T, not a verification", ev.GetBody())
	}
	at, err := record.EventID(run, ev)
	if err != nil {
		return nil, err
	}
	res := verifyResult{
		Event:              at,
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
	// WHAT THE RECORD NOW HOLDS ON THIS SUBJECT, read back after the write: every blocker on the
	// citation, and the debt this reading raised, each worded by its kind — the seat is told what
	// holds the PASS and whose act clears it here rather than left to find it in its work list.
	bs, err := record.PassBlockers(run)
	if err != nil {
		return nil, err
	}
	for _, b := range bs {
		switch {
		case b.Kind == record.BlockerReadingDebt && b.Owner == s.SeatID && b.Since() == at:
			res.Debt = b.WorkItem()
		case stored.GetAnchor() != "" && b.Subject == stored.GetAnchor():
			res.Blockers = append(res.Blockers, b.WorkItem())
		}
	}
	return res, nil
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
	return record.Check(func(l *record.Landing) error {
		report, err := l.Report()
		if err != nil {
			return err
		}
		return refuseMarked(report, quote)
	})
}

// refuseMarked refuses a backing reading of a source the lens found whose quote's sentence holds an
// evidence anchor — a citation or a computation token, read as anchors (anchor.BackingMark, the one
// predicate the reading debt reads too), never searched as text.
//
// WHY REFUSED, NOT RECORDED: a found source backs an uncited sentence, and the gap that asks blue for
// its citation (unsourced_claim) is true only of a sentence no evidence marks; on a marked one a
// backing reading tells no reader anything the sentence's own evidence does not. A quote the report
// does not hold once is the placement guard's to refuse, in its words.
func refuseMarked(report, quote string) error {
	_, end, err := anchortext.LocateOnce(report, quote, anchortext.StopAtParagraph)
	if err != nil {
		return nil
	}
	for _, sp := range anchor.Sentences(report) {
		if !(sp[0] < end && end <= sp[1]) {
			continue
		}
		marked := anchor.BackingMark(report[sp[0]:sp[1]])
		if marked == "" {
			return nil
		}
		act := "read that citation with `--anchor " + marked + "`"
		if anchor.Kind(marked) != "citation" {
			act = "its evidence is that " + anchor.Kind(marked) + ", which `reproduce` checks"
		}
		return feov.Errorf(feov.Validation, "this sentence is marked by %s; %s. A reading of a source you found that backs a sentence is the ground of an unsourced_claim gap, which is true only of a sentence no citation or computation marks. Nothing is recorded",
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
	// Event is the reading's events.id: the one event a reading writes, whatever its outcome.
	Event  int64  `json:"event"`
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
	// CopySubstitution is whose bytes the reading was made against, as the record holds them, and
	// CopyURL where that copy was read, when a copy stood in for the url.
	CopySubstitution string `json:"copy_substitution"`
	CopyURL          string `json:"copy_url,omitempty"`
	// Debt is the reading debt this reading raised, as the work list words it (record.Blocker.WorkItem):
	// a reading that finds against is the seat's own until a gap of its class stands on the subject.
	// Absent on a reading that owes nothing.
	Debt string `json:"debt,omitempty"`
	// Blockers are the PASS blockers standing on the citation after a reading of it, each worded by
	// its kind, this reading's own debt among them where another seat's reading raised one too.
	Blockers []string `json:"blockers,omitempty"`
	// reads is record.VerifyReads of the stored reading: a reading that reads nothing is worded so
	// before its outcome is.
	reads bool
}

func (r verifyResult) Human() string {
	out := r.human() + seat.CopyWords(r.CopySubstitution, r.CopyURL)
	if r.Debt != "" {
		out += "\nthis reading is your debt until a gap of its class stands on the subject: " + r.Debt
	}
	for _, b := range r.Blockers {
		out += "\n" + b
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
	recorded := " — recorded as event " + strconv.FormatInt(r.Event, 10)
	if r.Outcome == "unreachable" {
		return subject + " could not be read" + recorded + " as unreachable, with what you tried"
	}
	if !r.reads {
		return "this reading reads nothing of the source — " + subject + " is recorded as " + r.Outcome + " on a copy that is not the work (" + r.SourceCompleteness + "), on the claim " + claim + recorded
	}
	switch r.Outcome {
	case "refutes":
		return subject + " REFUTES the claim " + claim + recorded
	case "absent":
		return subject + " does NOT contain the claim " + claim + recorded + " as absent (the source says nothing about it; it does not say otherwise)"
	case "supports_with_bridge":
		return subject + " verified: " + r.Outcome + " (bridged: " + r.Bridge + "), on the claim " + claim + recorded
	default:
		return subject + " verified: " + r.Outcome + ", on the claim " + claim + recorded
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
