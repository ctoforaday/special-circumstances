package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING RECORDS WHAT ONE SOURCE DID FOR ONE SENTENCE AND HOLDS NOTHING BY ITSELF
// (plans/gap-is-the-one-holder.md). A lens reads a source for a sentence of the report in one verb,
// `verify`, by one of two subjects: a citation blue made (--anchor), or a source it found that blue
// never cited (--url, --title, --quote). A reading that finds against is the lens's debt until the
// gap of the class it obliges stands on the subject, or the lens reads the subject again; the gap is
// a `mint` of its own, admitted on the lens's standing reading (mintOnReading, readingdebt_test.go).

// readingResult is a reading's --json result, as the drive reads it.
type readingResult struct {
	Event              int64  `json:"event"`
	Anchor             string `json:"anchor"`
	URL                string `json:"url"`
	Outcome            string `json:"outcome"`
	Claim              string `json:"claim"`
	SourceCompleteness string `json:"source_completeness"`
}

func readingResultOf(out string) (readingResult, bool) {
	var env struct {
		Result readingResult `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil {
		return readingResult{}, false
	}
	return env.Result, true
}

// storedReading is seatID's latest reading on the record of citation anchorID, or — anchorID "" — of
// url for the sentence claim, with its events.id; ok is false where the record holds none.
func (r *runner) storedReading(seatID, anchorID, url, claim string) (v *recordpb.Verify, at int64, ok bool) {
	b, bok := r.board()
	if !bok {
		return nil, 0, false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		e := b.Events[i]
		got, isV := recordpb.BodyAs[*recordpb.Verify](e)
		if !isV || e.GetSeatId() != seatID {
			continue
		}
		if anchorID != "" && got.GetAnchor() == anchorID ||
			anchorID == "" && got.GetAnchor() == "" && got.GetUrl() == url && strings.TrimSpace(got.GetClaim()) == strings.TrimSpace(claim) {
			return got, b.At.Of(e).ID, true
		}
	}
	return nil, 0, false
}

// checkReading holds an admitted reading's result to the record: the reading is stored under its
// seat and subject, and its result names the event it was recorded as. It returns the stored
// reading and its events.id.
func (r *runner) checkReading(seatID, out string, args []string) (*recordpb.Verify, int64) {
	res, ok := readingResultOf(out)
	if !ok {
		r.noteActErr("an admitted reading's --json result does not parse: " + firstLine(out))
		return nil, 0
	}
	anchorID, url := argOf(args, "--anchor"), argOf(args, "--url")
	v, at, found := r.storedReading(seatID, anchorID, url, res.Claim)
	if !found {
		r.noteActErr(fmt.Sprintf("%s's admitted reading (anchor %q, url %q) is not on the record: %s", seatID, anchorID, url, firstLine(out)))
		return nil, 0
	}
	if res.Event != at {
		r.noteActErr(fmt.Sprintf("%s's reading result names event %d, and the record stores it at %d", seatID, res.Event, at))
	}
	return v, at
}

// citationURL is the url of citation anchorID's source, off `show evidence`'s source row.
func (r *runner) citationURL(anchorID string) string {
	b, ok := r.board()
	if !ok {
		return ""
	}
	// The count reads urls alone; the rows' location is not asked, so no report is rendered.
	for _, s := range record.EvidenceJSONOf(b.Events, b.At, "", nil).Sources {
		if s.Anchor == anchorID {
			return s.URL
		}
	}
	return ""
}

// copyNotTheWork reports whether the copy the run's cache holds for url is not the work — the
// entry the verb stamps the reading's completeness from, never the outcome alone.
func (r *runner) copyNotTheWork(url string) bool {
	held, _, cached, err := fetchcache.Lookup(r.run(), url)
	if err != nil {
		r.noteActErr("the run's cache could not be read to say which copy of " + url + " a reading resolves to: " + err.Error())
		return false
	}
	return cached && held.SourceCompleteness() == recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
}

// readCitation is the citation owner's reading of citation anchorID under outcome (axes sets the
// outcome, confidence, reason and bridge): the admitted reading held to what the record stores, and
// where it finds against, the mint that stands on it or the debt it leaves (mintOnCitationReading).
// fresh is whether the lens still draws a fresh defect this sitting.
func (r *runner) readCitation(seatID, anchorID, outcome string, kindless, fresh bool, axes func(string, *cmd) *cmd) {
	// A SECOND READING THAT READS NOTHING, while the gap carrying this seat's first is open and the
	// sentence unchanged, is refused (answerFailed drives that refusal on its own coin); this drive
	// expects its reading admitted, so it reads the source.
	if outcome == "unreachable" && r.ownUnreadGapOpen(seatID, anchorID) {
		outcome = "weak"
	}
	url := r.citationURL(anchorID)
	notTheWork := r.copyNotTheWork(url)
	// A READING OF A COPY THAT IS NOT THE WORK READS NOTHING, whatever its outcome, so a second one of
	// this seat's while the gap carrying its first is open is the refused unread repeat: the lens has
	// nothing to read here this sitting.
	if notTheWork && r.ownUnreadGapOpen(seatID, anchorID) {
		return
	}
	// ABSENT IS REFUSED ON A COPY THAT IS NOT THE WORK — its silence is not the work's — so the drive
	// reads such a copy under another negative outcome, which the record admits and reads as nothing.
	if notTheWork && outcome == "absent" {
		outcome = "refutes"
	}
	// --page names a page image, and no citation this sweep writes has pages, so a verify carrying one
	// is refused: the flag and its refusal, driven.
	paged := r.coin(10)
	c := axes(outcome, r.do("verify", seatID)).set("--anchor", anchorID).bare("--json")
	if paged {
		c = c.set("--page", "1")
	}
	// THE COPY THE READING NAMES IS THE ONE ITS CITATION'S SOURCE ROW NAMES: the sweep never clears the
	// cache and the first entry for a url is the one every read returns, so a reading and the act that
	// placed the citation resolve to one copy.
	wantCopy, wantCopyURL, cited := r.citationCopy(anchorID)
	out, err := c.result()
	var v *recordpb.Verify
	var at int64
	if err == nil {
		if cited {
			r.noteCopy("verify --anchor", out, wantCopy, wantCopyURL)
		}
		v, at = r.checkReading(seatID, out, c.args)
		if notTheWork && v != nil {
			r.noteNotTheWorkUnread(seatID, anchorID, v)
		}
	}
	r.noteBridge(seatID, outcome, kindless, out, err)
	// A READING THE DRIVE DOES NOT DRIVE TO BE REFUSED IS ADMITTED: the kindless bridge is noteBridge's
	// to hold, and a --page is refused as a validation — no citation this sweep writes has pages.
	switch {
	case paged && err == nil:
		r.noteActErr("verify --page on a source with no pages was admitted: " + firstLine(out))
	case paged:
		if env, ok := refusalEnvelopeIn(out); !ok || env.Code != string(feov.Validation) {
			r.noteActErr("verify --page refused without a validation envelope: " + firstLine(out))
		}
	case !kindless && err != nil && r.retiredUnderTheDraw(anchorID, out):
		// BLUE RETIRED THE CITATION BETWEEN THE DRAW AND THE READING: the refusal is the record's
		// call on the report the reading would have landed on, and names the retire.
		noteEither(c.args, "a reading of a citation blue retired between the draw and the reading", false)
	case !kindless && err != nil:
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
	}
	if v != nil && !paged && !kindless {
		r.mintOnCitationReading(seatID, anchorID, v, at, fresh)
	}
}

// retiredUnderTheDraw reports whether a refusal of a reading of anchorID is the record's for a citation
// retired by the time the reading landed: a conflict naming the retire, and the record holds the
// anchor retired.
func (r *runner) retiredUnderTheDraw(anchorID, out string) bool {
	env, ok := refusalEnvelopeIn(out)
	if !ok || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, anchorID+" is retired") {
		return false
	}
	_, retired, err := record.AnchorRetiredAt(r.run(), anchorID)
	return err == nil && retired
}

// noteNotTheWorkUnread holds an admitted reading of a copy that is not the work to the record: the
// copy stamped not_the_work, and citation anchorID on a failed check in state unread.
func (r *runner) noteNotTheWorkUnread(seatID, anchorID string, v *recordpb.Verify) {
	if v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK {
		r.noteActErr(fmt.Sprintf("%s's reading of %s, whose copy is not the work, stored source_completeness %s",
			seatID, anchorID, recordpb.Word(v.GetSourceCompleteness())))
		return
	}
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		r.noteActErr("the blockers could not be read after a reading of a copy that is not the work: " + err.Error())
		return
	}
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.Subject == anchorID && b.CheckState() == record.CheckUnread {
			r.mu.Lock()
			r.notTheWorkReadingsUnread++
			r.mu.Unlock()
			return
		}
	}
	r.noteActErr(fmt.Sprintf("%s's reading of %s, whose copy is not the work, left no unread failed check on it", seatID, anchorID))
}

// foundSources numbers the sources the found-source drive reads, one fresh source per act.
var foundSources atomic.Int64

// foundSourceURL is a fresh url for a source seatID found.
func foundSourceURL(seatID string) string {
	return fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, foundSources.Add(1))
}

// foundSourceSentences are the report's sentences a found-source reading reads a source for, in the
// order the drive tries them: lane 3's sentence, which no act of the sweep cites, and the anchor
// sentence every cite marks. The first the report holds once is read.
var foundSourceSentences = []string{"Across the sampled corpus the rise holds in every year measured.", "A § fuzz sentence to anchor findings."}

// sentenceMarked reports whether the sentence quote's only occurrence ends in carries an anchor that
// backs it — a citation or a computation — and whether the report holds quote once at all.
func sentenceMarked(report, quote string) (marked, once bool) {
	_, end, err := anchortext.LocateOnce(report, quote, anchortext.StopAtParagraph)
	if err != nil {
		return false, false
	}
	for _, sp := range anchor.Sentences(report) {
		if !(sp[0] < end && end <= sp[1]) {
			continue
		}
		anchor.Each(report[sp[0]:sp[1]], func(_, _ int, id string) {
			marked = marked || anchor.Backs(id)
		})
		return marked, true
	}
	return false, true
}

// foundSourceQuote is the sentence a found-source reading of seatID's under outcome reads a source
// for, and whether an anchor marks it: the first of foundSourceSentences the report holds once — or,
// for an `absent` reading on a coin, the first it holds once that is marked, since a found page
// silent on a cited sentence owes nothing (found_absent_on_marked_no_debt). "" where none qualifies.
func (r *runner) foundSourceQuote(cur, outcome string) (quote string, marked bool) {
	wantMarked := outcome == "absent" && r.coin(50)
	for _, s := range foundSourceSentences {
		m, once := sentenceMarked(cur, s)
		if !once || wantMarked && !m {
			continue
		}
		return s, m
	}
	return "", false
}

// readFoundSource is a lens's reading of a source it found, which blue never cited, for a sentence
// of the report: a backing outcome on a sentence no citation or computation marks is recorded; one
// on a marked sentence is refused naming its anchor; a negative outcome is recorded on any sentence.
// Then, by coin, the identical reading again in the sitting is refused naming the stored one; and a
// reading that finds against is minted on, or left as the lens's debt (mintOnFoundReading).
func (r *runner) readFoundSource(seatID, outcome string, kindless, fresh bool, axes func(string, *cmd) *cmd) {
	cur, err := record.RenderRun(r.run())
	if err != nil {
		r.noteActErr("the report could not be rendered for a found-source reading: " + err.Error())
		return
	}
	quote, marked := r.foundSourceQuote(cur, outcome)
	if quote == "" {
		return
	}
	url := foundSourceURL(seatID)
	// THE COPY THAT IS NOT THE WORK, read as a source the lens found: it reads nothing and places
	// nothing among the report's sources.
	r.mu.Lock()
	ntw := r.notTheWorkCite != ""
	r.mu.Unlock()
	if ntw && r.coin(15) {
		r.readNotTheWorkFound(seatID, quote, outcome, axes)
		return
	}
	c, wantCopy, wantCopyURL := r.foundSourceReading(seatID, outcome, quote, url, axes)
	backing := outcome != "refutes" && outcome != "absent" && outcome != "unreachable"
	out, err := c.result()
	env, isEnv := refusalEnvelopeIn(out)
	switch {
	case kindless:
		r.noteBridge(seatID, outcome, kindless, out, err)
		return
	case err != nil && backing && isEnv && env.Code == string(feov.Validation) && strings.Contains(env.Error, "this sentence is marked by"):
		// A CITE LANDING IN BETWEEN MARKS AN UNMARKED SENTENCE: the refusal is the record's call on the
		// report the reading would have landed on.
		r.mu.Lock()
		r.markedSentenceRefused++
		r.mu.Unlock()
		return
	case err != nil && backing && marked:
		r.noteActErr(fmt.Sprintf("%s's backing found-source reading of a marked sentence was refused without naming its anchor: %s", seatID, firstLine(out)))
		return
	case err != nil:
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		return
	case backing && marked:
		// A CITE LANDING IN BETWEEN MARKS IT; ONE LEAVING UNMARKS IT. Admitted is the record's call
		// on the report it landed on, read again here.
		if now, rerr := record.RenderRun(r.run()); rerr == nil {
			if m, _ := sentenceMarked(now, quote); m {
				r.noteActErr(fmt.Sprintf("%s's backing found-source reading of a sentence an anchor marks was admitted: %s", seatID, firstLine(out)))
				return
			}
		}
	}
	r.noteBridge(seatID, outcome, kindless, out, err)
	r.noteCopy("verify --url", out, wantCopy, wantCopyURL)
	first, at := r.checkReading(seatID, out, c.args)
	if first == nil {
		return
	}
	// THE IDENTICAL READING AGAIN IN THIS SITTING IS A RETRY, refused naming the stored one.
	if r.coin(20) {
		again, aerr := c.result()
		aenv, aok := refusalEnvelopeIn(again)
		switch {
		case aerr == nil:
			r.noteActErr(seatID + "'s identical found-source reading in one sitting was recorded again: " + firstLine(again))
		case aok && aenv.Code == string(feov.Validation) && strings.Contains(aenv.Error, fmt.Sprintf("you recorded a reading of %s for this sentence at event %d", url, at)):
			r.mu.Lock()
			r.foundSourceRetriesRefused++
			r.mu.Unlock()
		default:
			r.noteActErr(fmt.Sprintf("%s's identical found-source reading was refused without naming its stored reading at event %d: %s", seatID, at, firstLine(again)))
		}
	}
	r.mintOnFoundReading(seatID, quote, first, at, marked, fresh)
}

// foundSourceReading is seatID's `verify --url` of url for the sentence quote with outcome, and the
// copy the act should name.
func (r *runner) foundSourceReading(seatID, outcome, quote, url string, axes func(string, *cmd) *cmd) (c *cmd, wantCopy, wantCopyURL string) {
	wantCopy, wantCopyURL = r.seedCopy(url)
	c = axes(outcome, r.do("verify", seatID)).set("--url", url).set("--title", sourceTitle).set("--quote", quote).bare("--json")
	return c, wantCopy, wantCopyURL
}

// readFound is seatID's reading of url, a fresh source it found, against the sentence quote with
// outcome, admitted; it returns the stored reading's events.id, 0 where the reading was refused
// (noted). The url's copy is seeded here; a url read before is read again by readFoundAgain.
func (r *runner) readFound(seatID, url, quote, outcome string) int64 {
	c, wantCopy, wantCopyURL := r.foundSourceReading(seatID, outcome, quote, url, plainAxes)
	out, ok := c.admit()
	if !ok {
		return 0
	}
	r.noteCopy("verify --url", out, wantCopy, wantCopyURL)
	_, at := r.checkReading(seatID, out, c.args)
	return at
}

// readFoundAgain is seatID's reading of url, a source read before, against the sentence quote with
// outcome, admitted: the cache keeps the url's first copy, so none is seeded and the copy is not
// held. It returns the stored reading's events.id, 0 where the reading was refused (noted).
func (r *runner) readFoundAgain(seatID, url, quote, outcome string) int64 {
	c := plainAxes(outcome, r.do("verify", seatID)).set("--url", url).set("--title", sourceTitle).set("--quote", quote).bare("--json")
	out, ok := c.admit()
	if !ok {
		return 0
	}
	_, at := r.checkReading(seatID, out, c.args)
	return at
}

// readFoundClearing is seatID's reading of a fresh source it found for the sentence quote that
// clears a debt of its on the sentence: `supports` on a sentence no anchor marks; `absent` on a
// marked one, where a backing found reading is refused naming the anchor (F6 (a)) and a found page
// silent on a cited sentence owes nothing (F7 (a)), so the reading replaces the one the debt stood
// on. It returns the stored reading's events.id, or 0 where the reading was refused — a cite
// marking the sentence in between, or blue's edit taking it out, each the record's call on the
// report the reading would have landed on (noted either way) — or placed nowhere.
func (r *runner) readFoundClearing(seatID, quote string) int64 {
	cur, err := record.RenderRun(r.run())
	if err != nil {
		r.noteActErr("the report could not be rendered for a clearing found-source reading: " + err.Error())
		return 0
	}
	marked, once := sentenceMarked(cur, quote)
	if !once {
		noteEither([]string{"verify"}, "a clearing found-source reading of a sentence the report no longer holds once", false)
		return 0
	}
	outcome := "supports"
	if marked {
		outcome = "absent"
	}
	c, wantCopy, wantCopyURL := r.foundSourceReading(seatID, outcome, quote, foundSourceURL(seatID), plainAxes)
	out, err := c.result()
	if err != nil {
		env, ok := refusalEnvelopeIn(out)
		switch {
		case ok && env.Code == string(feov.Validation) && strings.Contains(env.Error, "this sentence is marked by"):
			noteEither(c.args, "a clearing found-source reading, refused by a cite marking the sentence in between", false)
		case ok && env.Code == string(feov.Validation) && strings.Contains(env.Error, "was not found in report.md"):
			noteEither(c.args, "a clearing found-source reading, refused by the report's placement under blue's edit", false)
		default:
			r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		}
		return 0
	}
	r.noteCopy("verify --url", out, wantCopy, wantCopyURL)
	_, at := r.checkReading(seatID, out, c.args)
	return at
}

// readCited is seatID's reading of citation anchorID with outcome, admitted; it returns the stored
// reading's events.id, 0 where the reading was refused (noted). The sitting's one reading of the
// citation is marked taken (firstThisSitting).
func (r *runner) readCited(seatID, anchorID, outcome string) int64 {
	r.firstThisSitting(seatID, anchorID)
	c := plainAxes(outcome, r.do("verify", seatID)).set("--anchor", anchorID).bare("--json")
	out, ok := c.admit()
	if !ok {
		return 0
	}
	_, at := r.checkReading(seatID, out, c.args)
	return at
}

// plainAxes sets a reading's outcome, confidence and reason, with no bridge.
func plainAxes(outcome string, c *cmd) *cmd {
	return c.set("--as", outcome).set("--confidence", "high").set("--reason", "fuzz: what the source actually says")
}

// setArg sets flag's value in args, where args carries it.
func setArg(args []string, flag, val string) {
	if i := slices.Index(args, flag); i >= 0 && i+1 < len(args) {
		args[i+1] = val
	}
}

// gapTally reads the gap counts off the stored record at the run's end: what held the run open
// (PR #1642's four, F7), the gaps minted and closed by epoch, and the one-gap rule over the report.
func gapTally(res *outcome, board record.Family, run record.Run) {
	res.openByFoundBy, res.mintsByEpoch, res.closesByEpoch = map[string]int{}, map[int]int{}, map[int]int{}
	report, rerr := record.RenderRun(run)
	spans, bare := anchor.Sentences(report), claimcount.BareAnchorIDs(report)
	perClassSentence := map[string][]*record.Gap{}
	var keys []string
	for _, g := range board.Gaps {
		res.mintsByEpoch[g.Epoch]++
		if g.HasClosed {
			res.closesByEpoch[g.ClosedEpoch]++
		}
		if !g.Open {
			continue
		}
		switch {
		case slices.Contains(record.ReadingMintedClasses(), g.Mint.GetClass()):
			res.openByFoundBy["reading"]++
		case len(g.Mint.GetFoundBy()) > 0:
			res.openByFoundBy["finding"]++
		case g.Mint.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR && anchor.Kind(g.Mint.GetAboutRef()) == "computation":
			res.openByFoundBy["unsound_computation"]++
		default:
			res.openByFoundBy["none"]++
		}
		// ONE GAP PER CLASS IN ONE SENTENCE, read off the report: the sentence the gap's token stands
		// in, as the rule reads it (record's gapSentenceIn) — the whole token inside one sentence span,
		// and an anchor standing bare in none.
		if rerr != nil {
			continue
		}
		if i, ok := sentenceIndexOf(report, spans, bare, g.ID); ok {
			k := recordpb.Word(g.Mint.GetClass()) + "@" + strconv.Itoa(i)
			if perClassSentence[k] == nil {
				keys = append(keys, k)
			}
			perClassSentence[k] = append(perClassSentence[k], g)
		}
	}
	// A GAP ITS SUCCESSOR NAMES IN --supersedes IS NOT A SECOND GAP ON THE SUBJECT: the rule admits the
	// successor over it by design (record.requireOneGapPerClassAndSubject), as checkOneGap does at the
	// mint. Every other pair is the rule broken, and the run names it.
	for _, k := range keys {
		gs := perClassSentence[k]
		superseded := map[string]bool{}
		for _, g := range gs {
			for _, a := range g.Mint.GetSupersedes() {
				superseded[a] = true
			}
		}
		var standing []string
		for _, g := range gs {
			if !superseded[g.ID] {
				standing = append(standing, g.ID)
			}
		}
		res.maxOpenPerClassSentence = max(res.maxOpenPerClassSentence, len(standing))
		if len(standing) > 1 && res.oneGapBreach == "" {
			i, _ := strconv.Atoi(k[strings.LastIndex(k, "@")+1:])
			res.oneGapBreach = fmt.Sprintf("seed %d: open gaps %s, of class %s, stand in one sentence: %q",
				res.seed, strings.Join(standing, ", "), k[:strings.LastIndex(k, "@")], plainText(report[spans[i][0]:spans[i][1]]))
		}
	}
}

// sentenceIndexOf is the index in spans of the sentence gap id's token stands in, as the one-gap rule
// reads it (record's gapSentenceIn): the whole token inside the span, and an anchor standing bare in
// no sentence.
func sentenceIndexOf(report string, spans [][2]int, bare []string, id string) (int, bool) {
	tok := anchor.Token(id)
	at := strings.Index(report, tok)
	if at < 0 || slices.Contains(bare, id) {
		return 0, false
	}
	for i, sp := range spans {
		if sp[0] <= at && at+len(tok) <= sp[1] {
			return i, true
		}
	}
	return 0, false
}

// planBlockersAndParties is a relayed plan's blockers counted by kind and the seats it readies, in
// the order it names them.
func planBlockersAndParties(plan map[string]any) (map[string]int, []string) {
	blockers, parties := map[string]int{}, []string{}
	bs, _ := plan["blockers"].([]any)
	for _, b := range bs {
		if m, _ := b.(map[string]any); m != nil {
			kind, _ := m["kind"].(string)
			blockers[kind]++
		}
	}
	ps, _ := plan["parties"].([]any)
	for _, p := range ps {
		if m, _ := p.(map[string]any); m != nil {
			if id, ok := m["seat_id"].(string); ok {
				parties = append(parties, id)
			}
		}
	}
	return blockers, parties
}

// laneCloseSentences are the sentences lanes 1 and 3 close on (lanePartText): one word apart, so
// each is one place a quote names.
var laneCloseSentences = [2]string{"The sample closed with costs rising steeply.", "The sample closed with costs rising sharply."}

// notTheWorkSource is the url of the source stock's copy that is not the work: its cache entry is
// stored stamped not_the_work, and blue cites it once a drawn run (citeNotTheWork).
const notTheWorkSource = "https://fuzz.invalid/not-the-work"

// citeNotTheWork is blue's one cite, a drawn run, of the source whose copy the run's cache holds as
// not the work — a sign-in page — which the owner's reading then reads nothing of (readNotTheWork).
// Forced runs take none: the unread check it leaves is a chain their bound does not count. It draws
// nothing from the run's RNG.
func (r *runner) citeNotTheWork(seatID string) {
	r.mu.Lock()
	due := r.notTheWorkCite == "" && !r.notTheWorkTried && !r.forcedVerdict() && !r.forceEstoppel
	r.notTheWorkTried = true
	r.mu.Unlock()
	if !due {
		return
	}
	body := []byte(sourceBody(notTheWorkSource))
	e := fetchcache.EntryFor(r.run(), notTheWorkSource, &fetchcache.Attempt{FetchedAt: fetchcache.FetchTime(), Body: body, ContentType: "text/plain", TextRetrieved: true}, "")
	e.Completeness, e.CompletenessReason = fetchcache.CompletenessNotTheWork, "fuzz: the source answered with its sign-in page"
	if _, err := fetchcache.Store(r.run(), e, body); err != nil {
		r.noteActErr("seeding the cache with a copy that is not the work: " + err.Error())
		return
	}
	out, ok := r.do("cite", seatID).set("--quote", "§ fuzz sentence").set("--url", notTheWorkSource).set("--title", sourceTitle).
		set("--source-span", sourceBody(notTheWorkSource)).bare("--json").admit()
	if !ok {
		return
	}
	var env struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.Label == "" {
		r.noteActErr("an admitted cite's --json result names no label: " + firstLine(out))
		return
	}
	r.mu.Lock()
	r.notTheWorkCite = env.Result.Label
	r.mu.Unlock()
}

// readNotTheWork is the owner's check of the citation whose copy is not the work: a reading of it
// reads nothing, whatever its outcome — the copy stamped not_the_work, and the citation left on an
// unread check. The copy that is not the work is the one standing reading citation_figure_misattribution
// also admits (record.readingMintedClasses, F3 (a)): on a coin the owner mints it about the citation,
// and on the other a claim_contradicts_source mint, which asks a reading of the work, is refused.
func (r *runner) readNotTheWork(seatID, anchorID string) {
	outcome := pick(r.rng, []string{"refutes", "supports", "weak"})
	c := r.do("verify", seatID).set("--anchor", anchorID).set("--as", outcome).set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: the page the source answered with is its sign-in page").bare("--json")
	out, ok := c.admit()
	if !ok {
		return
	}
	v, at := r.checkReading(seatID, out, c.args)
	if v == nil {
		return
	}
	r.noteNotTheWorkUnread(seatID, anchorID, v)
	if r.coin(50) {
		r.readingMintRefused(seatID, classContradicts, aboutArgs(anchorID), "of a copy that is not the work")
		return
	}
	r.mintOnReading(seatID, classMisattribution, aboutArgs(anchorID), at, "about")
}

// isNotTheWorkCite reports whether anchorID is blue's cite of the copy that is not the work.
func (r *runner) isNotTheWorkCite(anchorID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return anchorID != "" && anchorID == r.notTheWorkCite
}

// sourceRows is how many of `show evidence`'s source rows name url.
func (r *runner) sourceRows(url string) int {
	b, ok := r.board()
	if !ok {
		return -1
	}
	n := 0
	for _, s := range record.EvidenceJSONOf(b.Events, b.At, "", nil).Sources {
		if s.URL == url {
			n++
		}
	}
	return n
}

// readOverPrescription is a lens's reading of a source it found that refutes text blue applied
// verbatim from a red --new: the reading is recorded — a source that finds against the sentence is
// new evidence — and the gap it obliges meets the estoppel guard like any `--quote` mint
// (plans/gap-is-the-one-holder.md D10, F4 (a)): the claim_contradicts_source mint quoting the
// sentence is refused naming the prescribing gap (found_source_mints_estopped), and the lens's
// debt clears by the mint `--supersedes` that gap where the lens minted it, else by the lens's
// reading of another source for the sentence.
func (r *runner) readOverPrescription(seatID, line, prescribing string) {
	quote := strings.TrimSpace(plainText(line))
	url := foundSourceURL(seatID)
	c := r.do("verify", seatID).set("--as", "refutes").set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: the source finds against the text the lens prescribed").bare("--json").
		set("--url", url).set("--title", sourceTitle).set("--quote", quote)
	out, err := c.result()
	if err != nil {
		// THE LINE'S WORDS NAME NO ONE PLACE under a concurrent edit; the refusal is the record's call on
		// the report the reading would have landed on, and is not this drive's.
		noteEither(c.args, "a found-source reading over a prescription, refused by the report's placement", false)
		return
	}
	v, _ := r.checkReading(seatID, out, c.args)
	if v == nil {
		return
	}
	args := append(readingMintArgs(seatID, classContradicts), "--quote", quote)
	mout, merr := r.exec(args...)
	switch {
	case merr == nil && !r.stillEstops(prescribing):
		// BLUE REWORDED THE APPLIED SENTENCE UNDER THE READING (F12): the span is blue's, the mint
		// lands on the reading, and it discharges the debt.
		noteEither(args[:2], "a found-source mint over a prescription, landed after blue reworded the applied sentence (F12)", true)
		r.noteReadingMintLanded(seatID, mout, classContradicts, "quote")
		r.debtDischargedBy(seatID, quote, "mint", "the mint quoting the sentence blue reworded")
		return
	case merr == nil:
		r.noteActErr("a claim_contradicts_source mint quoting text blue applied from a red --new was admitted with no --supersedes, and a found-source gap meets the estoppel guard like any quote mint: " + firstLine(mout))
		r.noteReadingMintLanded(seatID, mout, classContradicts, "quote")
		return
	case strings.Contains(mout, oneGapWords):
		// A SIBLING'S claim_contradicts_source GAP ON THE SENTENCE — its own mint --supersedes the
		// prescribing gap — met the one-gap rule first, and discharges this lens's debt (F9).
		noteEither(args[:2], "a found-source mint over a prescription, refused by the one-gap rule over a sibling's gap on the sentence", false)
		r.oneGapOnReading(seatID, classContradicts, mout, args, "quote")
		return
	case !strings.Contains(mout, estoppelWords):
		r.noteActErr("a claim_contradicts_source mint quoting text blue applied from a red --new was refused, and not as an estoppel: " + firstLine(mout))
		return
	}
	// THE GUARD NAMES ONE PRESCRIPTION THE SENTENCE HOLDS, and the sentence can hold two: another
	// gap's text applied verbatim into the same line estops the mint as well, and the refusal may
	// name that one. The debt is then that gap's to answer.
	named := ""
	for _, g := range r.estoppingGaps() {
		if strings.Contains(mout, g) {
			named = g
			break
		}
	}
	if named == "" {
		r.noteActErr("a claim_contradicts_source mint quoting text blue applied from a red --new was refused as an estoppel naming no prescription that still estops (the drive drew " + prescribing + "): " + firstLine(mout))
		return
	}
	prescribing = named
	r.countReading(func(c *readingCounts) { c.foundSourceMintsEstopped++ })
	if r.minterOf(prescribing) == seatID {
		sout, serr := r.exec(append(args, "--supersedes", prescribing)...)
		if serr == nil {
			r.noteReadingMintLanded(seatID, sout, classContradicts, "quote")
			r.debtDischargedBy(seatID, quote, "mint", "the mint --supersedes "+prescribing+" on the reading over it")
			return
		}
		noteEither(args[:2], "a found-source mint superseding the prescribing gap, refused", false)
	}
	// ANOTHER LENS PRESCRIBED THE TEXT, or its gap takes no successor now: the lens's debt clears by its
	// reading of another source for the sentence that leaves no debt (readFoundClearing).
	if r.readFoundClearing(seatID, quote) != 0 {
		r.debtDischargedBy(seatID, quote, "reread", "the lens's reading of another source for the prescribed sentence")
	}
}

// readNotTheWorkFound is a lens's reading of the copy that is not the work as a source it found: it
// is recorded, reads nothing, and the report's sources gain no row for it.
func (r *runner) readNotTheWorkFound(seatID, quote, outcome string, axes func(string, *cmd) *cmd) {
	// ABSENT IS REFUSED ON A COPY THAT IS NOT THE WORK, and a bridge is the bridge drive's.
	if outcome == "absent" || outcome == "supports_with_bridge" {
		outcome = "supports"
	}
	rows := r.sourceRows(notTheWorkSource)
	c := axes(outcome, r.do("verify", seatID)).set("--url", notTheWorkSource).set("--title", sourceTitle).set("--quote", quote).bare("--json")
	out, ok := c.admit()
	if !ok {
		return
	}
	v, _ := r.checkReading(seatID, out, c.args)
	switch {
	case v == nil:
	case r.sourceRows(notTheWorkSource) != rows:
		r.noteActErr(fmt.Sprintf("%s's found-source reading of %s changed the report's source rows for it (%d before, %d after)", seatID, notTheWorkSource, rows, r.sourceRows(notTheWorkSource)))
	}
}

// plainText is s as a reader sees it: every anchor token out, whitespace runs collapsed, its
// punctuation kept.
func plainText(s string) string {
	var b strings.Builder
	last := 0
	anchor.Each(s, func(start, end int, _ string) {
		b.WriteString(s[last:start])
		last = end
	})
	b.WriteString(s[last:])
	return collapseWS(b.String())
}
