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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"google.golang.org/protobuf/proto"
)

// A READING THAT FINDS A DEFECT IN ITS SENTENCE MINTS ITS GAP IN THE SAME ACT. A lens reads a source
// for a sentence of the report in one verb, `verify`, by one of two subjects: a citation blue made
// (--anchor), or a source it found that blue never cited (--url, --title, --quote). The drive draws
// the gap flags exactly where the stored reading will mint — decided from the run's cache entry for
// the source, the entry the verb stamps the copy's completeness from, never from the outcome alone —
// and holds each act's answer to what the record then stores.

// readingClasses is the gap classes a reading of url with outcome mints in: the outcome's
// gap_classes for a reading of a citation, its found_source_gap_classes for a source the lens found,
// and none where the copy the run's cache holds for url is not the work (record.ReadingGapClasses on
// the body the verb will stamp). notTheWork reports that copy.
func (r *runner) readingClasses(url, outcome string, cited bool) (classes []string, notTheWork bool) {
	o, ok := record.SourceOutcomeOf(outcome)
	if !ok {
		return nil, false
	}
	held, _, cached, err := fetchcache.Lookup(r.run(), url)
	if err != nil {
		r.noteActErr("the run's cache could not be read to say which copy of " + url + " a reading resolves to: " + err.Error())
		return nil, false
	}
	completeness := recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED
	if cached {
		completeness = held.SourceCompleteness()
	}
	v := &recordpb.Verify{Outcome: &o, SourceCompleteness: &completeness}
	if cited {
		v.Anchor = proto.String("drawn")
	}
	for _, c := range record.ReadingGapClasses(v) {
		classes = append(classes, recordpb.Word(c))
	}
	return classes, completeness == recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK
}

// readingGapArgs is the gap flags a reading that mints passes: a class among the classes its outcome
// admits, the problem, the acceptance check and its kind, the three grades, the fix — the scenario
// blue's response reads back off the gap, as a mint's (directives) — and on a coin the optional
// complexity, as `mint` takes them. kind is the check kind drawn, which the caller notes for a
// computation gap.
func (r *runner) readingGapArgs(classes []string) (args []string, kind string) {
	kind = pick(r.rng, checkKinds)
	directive := pick(r.rng, []string{dirApply, dirCounter})
	// A FORCED-UNVERIFIED RUN'S GAPS ARE CONTESTED AND LOST, as its mints are (mintWith): the bench
	// declares the deadlock its verdict needs only over a contested docket.
	if r.forceUnverified {
		directive = dirDisputeLost
	}
	args = []string{"--class", pick(r.rng, classes),
		"--problem", "fuzz: the source says otherwise than the sentence",
		"--check-kind", kind, "--check", "the sentence states what the source says",
		"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
		"--fix", directive}
	if r.coin(30) {
		args = append(args, "--complexity", r.g())
	}
	return args, kind
}

// readingResult is a reading's --json result, as the drive reads it.
type readingResult struct {
	Anchor             string `json:"anchor"`
	URL                string `json:"url"`
	Outcome            string `json:"outcome"`
	Claim              string `json:"claim"`
	SourceCompleteness string `json:"source_completeness"`
	Gap                *struct {
		ID     string `json:"id"`
		Class  string `json:"class"`
		Minted bool   `json:"minted"`
	} `json:"gap"`
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

// storedMint is the Mint of gapID on the record.
func (r *runner) storedMint(gapID string) (*recordpb.Mint, bool) {
	b, ok := r.board()
	if !ok {
		return nil, false
	}
	for _, e := range b.Events {
		if m, isM := recordpb.BodyAs[*recordpb.Mint](e); isM && m.GetGapId() == gapID {
			return m, true
		}
	}
	return nil, false
}

// checkReading holds an admitted reading's result to the Verify the record stores for it: `gap`
// present exactly where carried_by is set, naming it, `minted` exactly where carrier_minted is true;
// and where the act minted its gap, the stored Mint's mint_reason is the --reason the drive passed and
// its fix_basis is `verified` where --new was passed and `proposed` where it was not. A minted gap is
// the minting lens's, noted as any landed mint is. It returns the stored reading.
func (r *runner) checkReading(seatID, out string, args []string) *recordpb.Verify {
	res, ok := readingResultOf(out)
	if !ok {
		r.noteActErr("an admitted reading's --json result does not parse: " + firstLine(out))
		return nil
	}
	anchorID, url := argOf(args, "--anchor"), argOf(args, "--url")
	claim := res.Claim
	v, _, found := r.storedReading(seatID, anchorID, url, claim)
	if !found {
		r.noteActErr(fmt.Sprintf("%s's admitted reading (anchor %q, url %q) is not on the record: %s", seatID, anchorID, url, firstLine(out)))
		return nil
	}
	switch {
	case (res.Gap != nil) != (v.GetCarriedBy() != ""):
		r.noteActErr(fmt.Sprintf("%s's reading printed gap %v and the stored reading's carried_by is %q", seatID, res.Gap, v.GetCarriedBy()))
		return v
	case res.Gap == nil:
		return v
	case res.Gap.ID != v.GetCarriedBy() || res.Gap.Minted != v.GetCarrierMinted():
		r.noteActErr(fmt.Sprintf("%s's reading printed gap %s (minted %v) and the stored reading names %s (carrier_minted %v)",
			seatID, res.Gap.ID, res.Gap.Minted, v.GetCarriedBy(), v.GetCarrierMinted()))
		return v
	}
	if !v.GetCarrierMinted() {
		return v
	}
	m, ok := r.storedMint(v.GetCarriedBy())
	if !ok {
		r.noteActErr("gap " + v.GetCarriedBy() + ", which " + seatID + "'s reading minted, has no Mint on the record")
		return v
	}
	wantBasis := "proposed"
	if argOf(args, "--new") != "" {
		wantBasis = "verified"
	}
	if m.GetMintReason() != argOf(args, "--reason") || m.GetFixBasis() != wantBasis {
		r.noteActErr(fmt.Sprintf("the gap %s's reading minted holds mint_reason %q and fix_basis %q; the reading passed --reason %q and the fix basis is %s",
			m.GetGapId(), m.GetMintReason(), m.GetFixBasis(), argOf(args, "--reason"), wantBasis))
	}
	r.noteMint(seatID, m.GetGapId())
	if argOf(args, "--check-kind") == "computation" {
		r.computationGaps.add(m.GetGapId())
	}
	return v
}

// ownFoundAgainstGapOpen is seatID's standing found-against check of citation anchorID whose reading
// had outcome, carried by a gap still open: the state in which seatID's second reading of the
// citation with that outcome is refused (record.requireNotARepeatReading). The check stands only
// while the sentence is unchanged — an edit clears it — so its standing says that too.
func (r *runner) ownFoundAgainstGapOpen(seatID, anchorID, outcome string) (record.Blocker, bool) {
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		r.noteActErr("the blockers could not be read to say whether " + anchorID + " stands found against: " + err.Error())
		return record.Blocker{}, false
	}
	open := map[string]bool{}
	for _, g := range r.openGaps() {
		open[g] = true
	}
	b, ok := r.board()
	if !ok {
		return record.Blocker{}, false
	}
	for _, f := range bs {
		if f.Kind != record.BlockerFailedCheck || f.CheckState() != record.CheckFoundAgainst || f.Subject != anchorID ||
			f.Owner != seatID || !open[f.AnsweredBy()] {
			continue
		}
		for _, e := range b.Events {
			if b.At.Of(e).ID != f.Since() {
				continue
			}
			if v, isV := recordpb.BodyAs[*recordpb.Verify](e); isV && recordpb.Word(v.GetOutcome()) == outcome {
				return f, true
			}
		}
	}
	return record.Blocker{}, false
}

// citationURL is the url of citation anchorID's source, off `show evidence`'s source row.
func (r *runner) citationURL(anchorID string) string {
	b, ok := r.board()
	if !ok {
		return ""
	}
	for _, s := range record.EvidenceJSONOf(b.Events, b.At).Sources {
		if s.Anchor == anchorID {
			return s.URL
		}
	}
	return ""
}

// citationFixNew is a replacement for the sentence citation anchorID's token sits in, as a reading's
// --new takes it: the sentence's words with every anchor that backs it kept in place, and the gap and
// finding anchors the accept moves left out; "" where the report does not hold the token once.
func (r *runner) citationFixNew(anchorID string) string {
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		return ""
	}
	tok := anchor.Token(anchorID)
	if strings.Count(cur, tok) != 1 {
		return ""
	}
	at := strings.Index(cur, tok)
	for _, sp := range anchor.Sentences(cur) {
		if at < sp[0] || at >= sp[1] {
			continue
		}
		text := cur[sp[0]:sp[1]]
		var b strings.Builder
		last := 0
		anchor.Each(text, func(start, end int, id string) {
			if !anchor.Backs(id) {
				b.WriteString(text[last:start])
				last = end
			}
		})
		b.WriteString(text[last:])
		if s := strings.TrimSpace(b.String()); s != "" {
			return "Read again, " + s
		}
	}
	return ""
}

// readCitation is the citation owner's reading of citation anchorID under outcome (axes sets the
// outcome, confidence, reason and bridge): the gap flags exactly where the reading finds against, a
// --new on a coin, the repeat over its own standing found-against check driven to its refusal, and the
// admitted reading held to what the record stores.
func (r *runner) readCitation(seatID, anchorID, outcome string, kindless bool, axes func(string, *cmd) *cmd) {
	// A SECOND READING THAT READS NOTHING, while the gap carrying this seat's first is open and the
	// sentence unchanged, is refused (answerFailed drives that refusal on its own coin); this drive
	// expects its reading admitted, so it reads the source.
	if outcome == "unreachable" && r.ownUnreadGapOpen(seatID, anchorID) {
		outcome = "weak"
	}
	url := r.citationURL(anchorID)
	// A SECOND READING THAT FINDS AGAINST, with the same outcome, while its own standing check's gap is
	// open and the sentence unchanged, is refused naming the earlier reading and that gap. By coin the
	// lens reads the source as backing instead — its own reading of the sentence as it stands, which
	// clears the check (foundAgainstCleared) — or drives the repeat to its refusal.
	if f, held := r.ownFoundAgainstGapOpen(seatID, anchorID, outcome); held {
		if r.coin(50) {
			r.foundAgainstCleared(seatID, anchorID, f)
			return
		}
		classes, _ := r.readingClasses(url, outcome, true)
		gap, _ := r.readingGapArgs(classes)
		c := axes(outcome, r.do("verify", seatID)).set("--anchor", anchorID).bare("--json")
		c.args = append(c.args, gap...)
		out, err := c.result()
		env, isEnv := refusalEnvelopeIn(out)
		switch {
		case err == nil:
			r.noteActErr(fmt.Sprintf("%s's repeat %s reading of %s, over its own standing check (event %d) carried by open gap %s, was admitted: %s",
				seatID, outcome, anchorID, f.Since(), f.AnsweredBy(), firstLine(out)))
		case !isEnv || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, "at event "+strconv.FormatInt(f.Since(), 10)) ||
			!strings.Contains(env.Error, "gap "+f.AnsweredBy()):
			r.noteActErr(fmt.Sprintf("%s's repeat %s reading of %s was refused without naming event %d and gap %s: %s",
				seatID, outcome, anchorID, f.Since(), f.AnsweredBy(), firstLine(out)))
		default:
			r.mu.Lock()
			r.foundAgainstRepeats++
			r.mu.Unlock()
		}
		return
	}
	classes, notTheWork := r.readingClasses(url, outcome, true)
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
	fixNew := ""
	if len(classes) > 0 {
		gap, _ := r.readingGapArgs(classes)
		c.args = append(c.args, gap...)
		if r.coin(30) {
			fixNew = r.citationFixNew(anchorID)
		}
	}
	// A COPY THAT IS NOT THE WORK READS NOTHING: by coin the lens passes the gap flags a negative
	// reading would take, and the record refuses them naming the copy.
	if notTheWork && !paged && !kindless && r.coin(50) {
		r.notTheWorkReading(seatID, anchorID, outcome, c)
		return
	}
	withNew := c
	if fixNew != "" {
		withNew = &cmd{r: r, args: append(slices.Clone(c.args), "--new", fixNew)}
	}
	// THE COPY THE READING NAMES IS THE ONE ITS CITATION'S SOURCE ROW NAMES: the sweep never clears the
	// cache and the first entry for a url is the one every read returns, so a reading and the act that
	// placed the citation resolve to one copy.
	wantCopy, wantCopyURL, cited := r.citationCopy(anchorID)
	out, err := withNew.result()
	// A --new READ OFF THE REPORT BEFORE THE ACT IS A FIX OF THE SENTENCE AS IT THEN STOOD: an edit
	// landing in between refuses it, and the lens reads again with the fix stated as prose.
	if err != nil && fixNew != "" && strings.Contains(out, "--new is not a fix") {
		noteEither(withNew.args, "a reading's --new read off a report an edit may move", false)
		withNew = c
		out, err = c.result()
	}
	if err == nil {
		if cited {
			r.noteCopy("verify --anchor", out, wantCopy, wantCopyURL)
		}
		v := r.checkReading(seatID, out, withNew.args)
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
	case !kindless && err != nil:
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
	}
}

// foundAgainstCleared is the found-against check's clearing by its own reader: seatID reads citation
// anchorID again, as backing, in a sitting after the reading that found against it, and the close of
// the gap carrying the check — refused while it stood — is admitted. By coin another lens's backing
// reading lands first, which is a disagreement and clears nothing: the close is still refused then.
func (r *runner) foundAgainstCleared(seatID, anchorID string, f record.Blocker) {
	g := f.AnsweredBy()
	closer := r.minterOf(g)
	if r.coin(50) {
		other := ""
		for _, l := range fuzzLensSeats {
			if l != seatID && l != closer {
				other = l
				break
			}
		}
		if other != "" && r.firstThisSitting(other, anchorID) {
			r.register("lens", other)
			r.do("verify", other).set("--anchor", anchorID).set("--as", "supports").set("--confidence", pick(r.rng, verifyConfidence)).
				set("--reason", "fuzz: another lens reads the source as backing the sentence").admit()
			still, ok := r.checkCarriedBy(g, anchorID, seatID)
			if !ok {
				r.noteActErr(fmt.Sprintf("another lens's backing reading of %s cleared %s's found-against check carried by %s", anchorID, seatID, g))
				return
			}
			r.refuseStandingClose(closer, g, pick(r.rng, []string{"repaired", "not_a_defect"}), still)
		}
	}
	out, ok := r.do("verify", seatID).set("--anchor", anchorID).set("--as", "supports").set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: read again, the source backs the sentence as it stands").bare("--json").admit()
	if !ok {
		return
	}
	r.checkReading(seatID, out, []string{"--anchor", anchorID, "--reason", "fuzz: read again, the source backs the sentence as it stands"})
	if _, still := r.checkCarriedBy(g, anchorID, seatID); still {
		r.noteActErr(fmt.Sprintf("%s's backing re-read of %s left its own found-against check standing on gap %s", seatID, anchorID, g))
		return
	}
	// ANOTHER CHECK THE GAP CARRIES STILL STANDS — another lens's reading the gap holds, or an unread
	// one — so the close stays refused, and is not this drive's to admit.
	if _, other := r.standingFailedCheckOf(g); other {
		return
	}
	// A COMPUTATION CHECK IS SATISFIED BEFORE IT IS CLOSED, as closeGap satisfies it.
	if r.computationGaps.has(g) {
		r.register("blue", "blue-respond")
		dir, argv, err := r.writeBundle("fuzz-cleared-"+g+".js", "console.log('answers "+g+"');")
		if err != nil {
			return
		}
		if _, ok := r.admit(append([]string{"compute", "--seat-id", "blue-respond", "--quote", "§ fuzz sentence",
			"--dir", dir, "--answers", g, "--reason", "fuzz: settling the computation check", "--"}, argv...)...); !ok {
			return
		}
	}
	if _, ok := r.admitFresh(g, "close", "--seat-id", closer, "--id", g, "--as", "not_a_defect",
		"--reason", "fuzz: the reader's own reading of the sentence as it stands backs it",
		"--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec"); ok {
		r.mu.Lock()
		r.foundAgainstClearedByOwnReading++
		r.mu.Unlock()
	}
}

// checkCarriedBy is the found-against check of citation anchorID by owner that gap g carries, while it
// stands.
func (r *runner) checkCarriedBy(g, anchorID, owner string) (record.Blocker, bool) {
	bs, err := record.PassBlockers(r.runHandle)
	if err != nil {
		return record.Blocker{}, false
	}
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.CheckState() == record.CheckFoundAgainst && b.AnsweredBy() == g &&
			b.Subject == anchorID && b.Owner == owner {
			return b, true
		}
	}
	return record.Blocker{}, false
}

// notTheWorkReading drives a reading of a copy that is not the work with the gap flags a negative
// reading of the work would take, which the record refuses naming not_the_work, recording nothing.
func (r *runner) notTheWorkReading(seatID, anchorID, outcome string, c *cmd) {
	gapped := c
	if argOf(c.args, "--class") == "" {
		gap, _ := r.readingGapArgs(fuzzReadingClasses)
		gapped = &cmd{r: r, args: append(slices.Clone(c.args), gap...)}
	}
	_, before, had := r.storedReading(seatID, anchorID, "", "")
	out, err := gapped.result()
	env, isEnv := refusalEnvelopeIn(out)
	if err == nil || !isEnv || env.Code != string(feov.Validation) || !strings.Contains(env.Error, "not_the_work") {
		r.noteActErr(fmt.Sprintf("%s's %s reading of %s, whose copy is not the work, with gap flags: want refused naming not_the_work, got %v: %s", seatID, outcome, anchorID, err, firstLine(out)))
		return
	}
	if _, after, has := r.storedReading(seatID, anchorID, "", ""); has && (!had || after != before) {
		r.noteActErr(fmt.Sprintf("%s's refused reading of %s, whose copy is not the work, was recorded at event %d", seatID, anchorID, after))
		return
	}
	r.mu.Lock()
	r.notTheWorkFlagsRefused++
	r.mu.Unlock()
}

// fuzzReadingClasses is a class a negative reading of a citation admits, for flags passed on a
// reading the record refuses before it reads them.
var fuzzReadingClasses = []string{recordpb.Word(recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)}

// noteNotTheWorkUnread holds an admitted reading of a copy that is not the work to the record: no gap,
// no carrier, the copy stamped not_the_work, and citation anchorID on a failed check in state unread.
func (r *runner) noteNotTheWorkUnread(seatID, anchorID string, v *recordpb.Verify) {
	if v.GetCarriedBy() != "" || v.CarrierMinted != nil ||
		v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK {
		r.noteActErr(fmt.Sprintf("%s's reading of %s, whose copy is not the work, stored carried_by %q, carrier_minted %v, source_completeness %s",
			seatID, anchorID, v.GetCarriedBy(), v.CarrierMinted, recordpb.Word(v.GetSourceCompleteness())))
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

// mintsLeft is what lens seatID's work list states is left of its mint budget, and whether its
// standing read.
func (r *runner) mintsLeft(seatID string) (int, bool) {
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil || w.Sitting.MintBudget == nil || !w.Sitting.MintBudget.Measured {
		return 0, false
	}
	return w.Sitting.MintBudget.Left, true
}

// heldOn reports whether an open gap of class stands on the sentence quote's only occurrence ends in:
// the one-gap rule then names it, and the reading mints nothing.
func (r *runner) heldOn(class, quote string) bool {
	open, _ := r.sameClassAndSubject(class, []string{"--quote", quote})
	return len(open) > 0
}

// readFoundSource is a lens's reading of a source it found, which blue never cited, for a sentence
// of the report. Its gap is minted against the lens's budget: a backing outcome on a sentence no
// citation or computation marks mints unsourced_claim, refutes mints claim_contradicts_source on any
// sentence, absent mints nothing. The drive holds the act to the record's answer: a backing reading of
// a marked sentence refused naming its anchor; a minting reading at the cap refused for budget,
// nothing recorded; below it admitted and held to what is stored. Then, by coin, the identical reading
// again in the sitting is refused naming the stored one, and another source read for the same
// sentence is recorded, naming the gap the first one's reading left standing.
func (r *runner) readFoundSource(seatID, outcome string, kindless bool, axes func(string, *cmd) *cmd) {
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		r.noteActErr("the report could not be rendered for a found-source reading: " + err.Error())
		return
	}
	quote, marked := "", false
	for _, s := range foundSourceSentences {
		if m, once := sentenceMarked(cur, s); once {
			quote, marked = s, m
			break
		}
	}
	if quote == "" {
		return
	}
	read := func(url string) (*cmd, string, string, []string) {
		wantCopy, wantCopyURL := r.seedCopy(url)
		c := axes(outcome, r.do("verify", seatID)).set("--url", url).set("--title", sourceTitle).set("--quote", quote).bare("--json")
		classes, _ := r.readingClasses(url, outcome, false)
		if len(classes) > 0 {
			gap, _ := r.readingGapArgs(classes)
			c.args = append(c.args, gap...)
		}
		return c, wantCopy, wantCopyURL, classes
	}
	url := fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, foundSources.Add(1))
	// THE COPY THAT IS NOT THE WORK, read as a source the lens found: it reads nothing, mints no gap,
	// and places nothing among the report's sources.
	r.mu.Lock()
	ntw := r.notTheWorkCite != ""
	r.mu.Unlock()
	if ntw && r.coin(15) {
		r.readNotTheWorkFound(seatID, quote, outcome, axes)
		return
	}
	c, wantCopy, wantCopyURL, classes := read(url)
	backing := len(classes) > 0 && outcome != "refutes"
	class := argOf(c.args, "--class")
	left, measured := r.mintsLeft(seatID)
	atCap := len(classes) > 0 && measured && left == 0 && !r.heldOn(class, quote)
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
	case err != nil && strings.Contains(out, "its mint budget admits no more"):
		// THE LENSES SIT CONCURRENTLY, and its gaps only close under another seat: a budget read
		// before the act can only have grown, so a refusal is admissible only where it read none left.
		if !atCap {
			// ITS OWN SUCCESSOR MAY LAND IN BETWEEN: a regression close mints as the closer from the
			// chair's branch. A refusal is admissible where the budget now reads none left.
			if now, ok := r.mintsLeft(seatID); ok && now == 0 {
				noteEither(c.args, "a found-source mint whose budget a concurrent mint of the lens spent", false)
				return
			}
			r.noteActErr(fmt.Sprintf("%s's found-source reading was refused for budget, and its work list read %d mint(s) left before it: %s", seatID, left, firstLine(out)))
			return
		}
		if _, _, recorded := r.storedReading(seatID, "", url, quote); recorded {
			r.noteActErr(seatID + "'s found-source reading refused for budget was recorded: " + url)
			return
		}
		r.mu.Lock()
		r.foundSourceBudgetRefused++
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
		if now, rerr := reportproj.RenderFromRecord(r.run()); rerr == nil {
			if m, _ := sentenceMarked(now, quote); m {
				r.noteActErr(fmt.Sprintf("%s's backing found-source reading of a sentence an anchor marks was admitted: %s", seatID, firstLine(out)))
				return
			}
		}
	case atCap:
		noteEither(c.args, "a found-source mint at the cap, admitted where a close freed a mint in between", true)
	}
	r.noteBridge(seatID, outcome, kindless, out, err)
	r.noteCopy("verify --url", out, wantCopy, wantCopyURL)
	first := r.checkReading(seatID, out, c.args)
	if first == nil {
		return
	}
	// THE IDENTICAL READING AGAIN IN THIS SITTING IS A RETRY, refused naming the stored one.
	if r.coin(20) {
		_, at, _ := r.storedReading(seatID, "", url, quote)
		again, aerr := c.result()
		aenv, aok := refusalEnvelopeIn(again)
		switch {
		case aerr == nil:
			r.noteActErr(seatID + "'s identical found-source reading in one sitting was recorded again: " + firstLine(again))
		case aok && aenv.Code == string(feov.Validation) && strings.Contains(aenv.Error, fmt.Sprintf("you recorded a reading of %s for this sentence at event %d", url, at)):
			r.mu.Lock()
			r.foundSourceRetriesRefused++
			r.mu.Unlock()
		case strings.Contains(again, "its mint budget admits no more"):
			noteEither(c.args, "an identical found-source reading whose gap closed in between, refused for budget", false)
		default:
			r.noteActErr(fmt.Sprintf("%s's identical found-source reading was refused without naming its stored reading at event %d: %s", seatID, at, firstLine(again)))
		}
	}
	// ANOTHER SOURCE FOR THE SAME SENTENCE IS RECORDED, and where the first reading's gap still stands
	// with the second's class, it names that gap and mints none.
	if g := first.GetCarriedBy(); g != "" && r.coin(20) && slices.Contains(r.openGaps(), g) {
		url2 := fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, foundSources.Add(1))
		c2, want2, want2URL, _ := read(url2)
		setArg(c2.args, "--class", argOf(c.args, "--class"))
		out2, ok := c2.admit()
		if !ok {
			return
		}
		r.noteCopy("verify --url", out2, want2, want2URL)
		second := r.checkReading(seatID, out2, c2.args)
		if second != nil && second.GetCarriedBy() != g && slices.Contains(r.openGaps(), g) {
			r.noteActErr(fmt.Sprintf("%s's second source for one sentence was carried by %s, while %s, which carries the first, stands open", seatID, second.GetCarriedBy(), g))
		}
	}
}

// setArg sets flag's value in args, where args carries it.
func setArg(args []string, flag, val string) {
	if i := slices.Index(args, flag); i >= 0 && i+1 < len(args) {
		args[i+1] = val
	}
}

// fixFoundAgainst is blue's remedy of a citation a lens's reading finds against: by coin, the edit
// that rewrites the sentence — which the record lists as reopening the citation — or the retire of
// the citation answering the gap that carries the check. One time in ten blue first tries to end its
// sitting with the check standing, and its envelope is refused naming it.
func (r *runner) fixFoundAgainst(seatID string, b record.Blocker) {
	if r.coin(10) {
		r.envelopeRefusedEarly(seatID, "reading of citation "+b.Subject+" at event")
	}
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		r.noteActErr("the report could not be rendered to fix citation " + b.Subject + ": " + err.Error())
		return
	}
	// A SENTENCE THE REPORT HOLDS TWICE HAS NO SPAN AN EDIT CAN QUOTE BEFORE THE ANCHOR ALONE, so its
	// citation is retired.
	words := wordsBefore(cur, b.Subject)
	if words == "" || r.coin(50) {
		rt := r.do("retire", seatID).set("--anchor", b.Subject).
			set("--reason", "fuzz: a lens's reading finds the source against the sentence, so the sentence stands without it")
		if g := b.AnsweredBy(); g != "" {
			rt = rt.set("--answers", g)
		}
		if _, ok := rt.admit(); ok {
			r.mu.Lock()
			r.foundAgainstRetires++
			r.mu.Unlock()
		}
		return
	}
	if _, ok := r.do("edit", seatID).set("--quote", words).set("--new", "Read against its source, "+words).
		set("--reason", "fuzz: the sentence now says what the source behind "+b.Subject+" says").admit(); !ok {
		return
	}
	if !r.lastEditReopened(seatID, b.Subject) {
		r.noteActErr("blue's edit of the sentence citation " + b.Subject + " backs does not list it reopened")
		return
	}
	r.mu.Lock()
	r.foundAgainstEdits++
	r.mu.Unlock()
}

// lastEditReopened reports whether seatID's latest edit on the record lists anchorID reopened.
func (r *runner) lastEditReopened(seatID, anchorID string) bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		if e, isE := recordpb.BodyAs[*recordpb.BlueEdit](b.Events[i]); isE && b.Events[i].GetSeatId() == seatID {
			return slices.Contains(e.GetReopened(), anchorID)
		}
	}
	return false
}

// readingTally reads the reading counts off the stored record at the run's end, never off --as, and
// what held the run open (PR #1642's four, F7).
func readingTally(res *outcome, board record.Family, run record.Run) {
	mintOf := map[string]*recordpb.Mint{}
	for _, e := range board.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			mintOf[m.GetGapId()] = m
		}
	}
	byReading := map[string]bool{}
	for _, e := range board.Events {
		v, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok || v.GetCarriedBy() == "" {
			continue
		}
		if !v.GetCarrierMinted() {
			res.readingsHeld++
			continue
		}
		res.readingsMinted++
		byReading[v.GetCarriedBy()] = true
		if m := mintOf[v.GetCarriedBy()]; v.GetAnchor() == "" && m != nil &&
			(m.GetClass() == recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM || m.GetClass() == recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE) {
			res.foundSourceGaps++
		}
	}
	res.openByFoundBy, res.mintsByEpoch, res.closesByEpoch = map[string]int{}, map[int]int{}, map[int]int{}
	report, rerr := reportproj.RenderFromRecord(run)
	spans := anchor.Sentences(report)
	perClassSentence := map[string]int{}
	for _, g := range board.Gaps {
		res.mintsByEpoch[g.Epoch]++
		if g.HasClosed {
			res.closesByEpoch[g.ClosedEpoch]++
		}
		if !g.Open {
			continue
		}
		switch {
		case byReading[g.ID]:
			res.openByFoundBy["reading"]++
		case len(g.Mint.GetFoundBy()) > 0:
			res.openByFoundBy["finding"]++
		case g.Mint.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR && anchor.Kind(g.Mint.GetAboutRef()) == "computation":
			res.openByFoundBy["unsound_computation"]++
		default:
			res.openByFoundBy["none"]++
		}
		// ONE GAP PER CLASS IN ONE SENTENCE, read off the report: where the gap's token stands.
		at := strings.Index(report, anchor.Token(g.ID))
		if rerr != nil || at < 0 {
			continue
		}
		for i, sp := range spans {
			if sp[0] <= at && at < sp[1] {
				k := recordpb.Word(g.Mint.GetClass()) + "@" + strconv.Itoa(i)
				perClassSentence[k]++
				res.maxOpenPerClassSentence = max(res.maxOpenPerClassSentence, perClassSentence[k])
				break
			}
		}
	}
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

// repeatedSentenceSeeds are the two sentences lanes 1 and 3 close on: one word apart, so blue can cite
// each copy by its own quote before its edit makes the second read as the first (repeatedSentence).
var repeatedSentenceSeeds = [2]string{"The sample closed with costs rising steeply.", "The sample closed with costs rising sharply."}

// repeatedSentence is the sentence the report holds twice once blue has made its copies one wording.
const repeatedSentence = "The sample closed with costs rising sharply."

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
	e := fetchcache.EntryFor(r.run(), notTheWorkSource, &fetchcache.Attempt{Body: body, ContentType: "text/plain", TextRetrieved: true}, "")
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
// reads nothing, whatever its outcome. A negative outcome by coin first passes the gap flags, which
// the record refuses naming not_the_work, recording nothing; the reading without them is admitted
// with no gap and no carrier, the copy stamped not_the_work, and the citation left on an unread check.
func (r *runner) readNotTheWork(seatID, anchorID string) {
	outcome := pick(r.rng, []string{"refutes", "supports", "weak"})
	c := r.do("verify", seatID).set("--anchor", anchorID).set("--as", outcome).set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: the page the source answered with is its sign-in page").bare("--json")
	if outcome == "refutes" && r.coin(50) {
		r.notTheWorkReading(seatID, anchorID, outcome, c)
	}
	out, ok := c.admit()
	if !ok {
		return
	}
	if v := r.checkReading(seatID, out, c.args); v != nil {
		r.noteNotTheWorkUnread(seatID, anchorID, v)
	}
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
	for _, s := range record.EvidenceJSONOf(b.Events, b.At).Sources {
		if s.URL == url {
			n++
		}
	}
	return n
}

// stockRepeatedSentence is blue's making of the report's twice-held sentence, once a drawn run: it
// cites each copy of repeatedSentenceSeeds by its own quote while the two differ, then edits the first
// to the second's wording, so each copy of one sentence carries its own citation. It draws nothing
// from the run's RNG.
func (r *runner) stockRepeatedSentence(seatID string) {
	r.mu.Lock()
	due := !r.repeatedTried && !r.forcedVerdict() && !r.forceEstoppel
	r.repeatedTried = true
	r.mu.Unlock()
	if !due {
		return
	}
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil || strings.Count(plainText(cur), repeatedSentenceSeeds[0]) != 1 || strings.Count(plainText(cur), repeatedSentenceSeeds[1]) != 1 {
		return // a lane withheld its part, so the report holds one copy
	}
	var cites [2]string
	for i, quote := range []string{"costs rising steeply", "costs rising sharply"} {
		path := "/repeat-" + strconv.Itoa(i+1)
		out, ok := r.do("cite", seatID).set("--quote", quote).set("--url", sourceURL(path)).set("--title", sourceTitle).
			set("--source-span", sourceBody(path)).bare("--json").admit()
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
		cites[i] = env.Result.Label
	}
	if _, ok := r.do("edit", seatID).set("--quote", "costs rising steeply").set("--new", "costs rising sharply").
		set("--reason", "fuzz: one wording for the sample's close").admit(); !ok {
		return
	}
	r.mu.Lock()
	r.repeatedCites = cites
	r.mu.Unlock()
}

// pendingAccept is a gap a reading of one copy of the twice-held sentence minted with a --new, which
// blue's next sitting may accept.
type pendingAccept struct{ gap, cite, other, fixNew string }

// readRepeatedSentence is the citation owner's refuting reading of one copy of the twice-held
// sentence, by that copy's citation: the gap it mints or names stands at the end of the sentence the
// citation's token sits in — after its last word, before a citation token that ends it — in that
// copy and no other, read off the rendered report. A found-source reading quoting the sentence is
// refused as ambiguous.
func (r *runner) readRepeatedSentence(seatID string) {
	r.mu.Lock()
	cites := r.repeatedCites
	r.mu.Unlock()
	if cites[0] == "" || !r.coin(30) {
		return
	}
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil || strings.Count(plainText(cur), repeatedSentence) != 2 {
		return // an edit or a retire has since changed a copy
	}
	i := r.rng.Intn(2)
	cite, other := cites[i], cites[1-i]
	if strings.Count(cur, anchor.Token(cite)) != 1 || strings.Count(cur, anchor.Token(other)) != 1 {
		return
	}
	if _, held := r.ownFoundAgainstGapOpen(seatID, cite, "refutes"); held || !r.firstThisSitting(seatID, cite) {
		return
	}
	classes, notTheWork := r.readingClasses(r.citationURL(cite), "refutes", true)
	if notTheWork || len(classes) == 0 {
		return
	}
	gap, _ := r.readingGapArgs(classes)
	c := r.do("verify", seatID).set("--anchor", cite).set("--as", "refutes").set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: the source says otherwise than this copy of the sentence").bare("--json")
	c.args = append(c.args, gap...)
	fixNew := ""
	if r.coin(50) {
		fixNew = r.citationFixNew(cite)
	}
	run := c
	if fixNew != "" {
		run = &cmd{r: r, args: append(slices.Clone(c.args), "--new", fixNew)}
	}
	out, ok := run.admit()
	if !ok {
		return
	}
	v := r.checkReading(seatID, out, run.args)
	if v == nil || v.GetCarriedBy() == "" {
		return
	}
	g := v.GetCarriedBy()
	now, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		r.noteActErr("the report could not be rendered after a reading of the twice-held sentence: " + err.Error())
		return
	}
	if msg := gapEndsSentenceOf(now, g, cite); msg != "" {
		r.noteActErr(msg)
		return
	}
	r.mu.Lock()
	r.repeatedSentenceGaps++
	if fixNew != "" && v.GetCarrierMinted() {
		r.pendingAccepts = append(r.pendingAccepts, pendingAccept{gap: g, cite: cite, other: other, fixNew: fixNew})
	}
	r.mu.Unlock()
	// THE SENTENCE'S WORDS NAME NO ONE PLACE: quoted, it is refused as ambiguous.
	if r.coin(50) {
		url := fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, foundSources.Add(1))
		amb := r.do("verify", seatID).set("--url", url).set("--title", sourceTitle).set("--quote", repeatedSentence).
			set("--as", "refutes").set("--confidence", "high").set("--reason", "fuzz: a source the lens found, against the sentence's words").bare("--json")
		ambGap, _ := r.readingGapArgs([]string{recordpb.Word(recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)})
		amb.args = append(amb.args, ambGap...)
		out, err := amb.result()
		switch {
		case err == nil:
			r.noteActErr("a found-source reading quoting the twice-held sentence was recorded: " + firstLine(out))
		case strings.Contains(out, "its mint budget admits no more"):
			noteEither(amb.args, "a found-source reading of the twice-held sentence at the lens's cap", false)
		case !strings.Contains(out, "occurs more than once"):
			r.noteActErr("a found-source reading quoting the twice-held sentence was refused, and not as ambiguous: " + firstLine(out))
		}
	}
}

// gapEndsSentenceOf is "" where gap g's token stands once in report, in the sentence citation cite's
// token sits in, before that token, with nothing but the sentence's terminator visible after it; else
// what it found.
func gapEndsSentenceOf(report, g, cite string) string {
	tg, tc := anchor.Token(g), anchor.Token(cite)
	if n := strings.Count(report, tg); n != 1 {
		return fmt.Sprintf("gap %s's token stands %d times in the report, want once, in the sentence citation %s's token sits in", g, n, cite)
	}
	at := strings.Index(report, tc)
	for _, sp := range anchor.Sentences(report) {
		if at < sp[0] || at >= sp[1] {
			continue
		}
		s := report[sp[0]:sp[1]]
		gi, ci := strings.Index(s, tg), strings.Index(s, tc)
		switch {
		case gi < 0:
			return fmt.Sprintf("gap %s's token is not in the sentence citation %s's token sits in: %q", g, cite, s)
		case gi > ci:
			return fmt.Sprintf("gap %s's token stands after citation %s's, which ends the sentence: %q", g, cite, s)
		case strings.TrimSpace(plainText(s[gi:])) != ".":
			return fmt.Sprintf("gap %s's token does not stand at the end of its sentence: %q", g, s)
		}
		return ""
	}
	return fmt.Sprintf("citation %s's token stands in no sentence of the report", cite)
}

// acceptRepeated is blue's accept, on a coin, of a fix a reading prescribed for one copy of the
// twice-held sentence: applied to the cited copy only — the other keeps its words — and recorded as
// an edit of the span as written (exact_span).
func (r *runner) acceptRepeated(seatID string) {
	r.mu.Lock()
	pending := r.pendingAccepts
	r.pendingAccepts = nil
	r.mu.Unlock()
	for _, p := range pending {
		if !slices.Contains(r.openGaps(), p.gap) || !r.coin(50) {
			continue
		}
		if _, ok := r.do("edit", seatID).set("--answers", p.gap).bare("--accept").
			set("--reason", "fuzz: accepting the reading's fix of this copy of the sentence").admit(); !ok {
			continue
		}
		cur, err := reportproj.RenderFromRecord(r.run())
		if err != nil {
			r.noteActErr("the report could not be rendered after blue's accept of " + p.gap + ": " + err.Error())
			continue
		}
		got := func(id string) string {
			at := strings.Index(cur, anchor.Token(id))
			for _, sp := range anchor.Sentences(cur) {
				if at >= sp[0] && at < sp[1] {
					return strings.TrimSpace(plainText(cur[sp[0]:sp[1]]))
				}
			}
			return ""
		}
		if want, other := strings.TrimSpace(plainText(p.fixNew)), plainText(repeatedSentence); got(p.cite) != want || got(p.other) != other {
			r.noteActErr(fmt.Sprintf("blue's accept of %s left the cited copy %q (want %q) and the other %q (want %q)", p.gap, got(p.cite), want, got(p.other), other))
			continue
		}
		if !r.lastEditExact(seatID) {
			r.noteActErr("blue's accept of " + p.gap + " over one copy of the twice-held sentence is not recorded with exact_span")
			continue
		}
		r.mu.Lock()
		r.repeatedSentenceAccepts++
		r.mu.Unlock()
	}
}

// lastEditExact reports whether seatID's latest edit on the record replaced its span as written.
func (r *runner) lastEditExact(seatID string) bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		if e, isE := recordpb.BodyAs[*recordpb.BlueEdit](b.Events[i]); isE && b.Events[i].GetSeatId() == seatID {
			return e.GetExactSpan()
		}
	}
	return false
}

// readOverPrescription is a lens's negative reading of text blue applied verbatim from a red --new —
// by a citation the accept carried onto it where one stands there, else of a source the lens found,
// quoting it — which is admitted with its gap minted or named, and writes no estoppel entry: a source
// that finds against the sentence is new evidence, not the lens's prescription relitigated.
func (r *runner) readOverPrescription(seatID, line string) {
	if !r.mintsFresh(seatID) {
		return
	}
	estoppels := r.estoppelEntries(seatID)
	var args []string
	anchor.Each(line, func(_, _ int, id string) {
		if args == nil && anchor.Kind(id) == "citation" && !r.isNotTheWorkCite(id) {
			if _, held := r.ownFoundAgainstGapOpen(seatID, id, "refutes"); !held && r.firstThisSitting(seatID, id) {
				args = []string{"--anchor", id}
			}
		}
	})
	quote := strings.TrimSpace(plainText(line))
	if args == nil {
		args = []string{"--url", fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, foundSources.Add(1)), "--title", sourceTitle, "--quote", quote}
	}
	c := r.do("verify", seatID).set("--as", "refutes").set("--confidence", pick(r.rng, verifyConfidence)).
		set("--reason", "fuzz: the source finds against the text the lens prescribed").bare("--json")
	c.args = append(c.args, args...)
	gap, _ := r.readingGapArgs([]string{recordpb.Word(recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE)})
	c.args = append(c.args, gap...)
	out, err := c.result()
	switch {
	case err != nil && strings.Contains(out, "its mint budget admits no more"):
		noteEither(c.args, "a found-source reading over a prescription at the lens's cap", false)
		return
	case err != nil:
		r.noteActErr("a reading over text blue applied from a red --new was refused, and a reading's gap is exempt from estoppel: " + firstLine(out))
		return
	}
	v := r.checkReading(seatID, out, c.args)
	if v == nil || v.GetCarriedBy() == "" {
		r.noteActErr("a refuting reading over text blue applied from a red --new names no gap: " + firstLine(out))
		return
	}
	if n := r.estoppelEntries(seatID); n != estoppels {
		r.noteActErr(fmt.Sprintf("%s's reading over text blue applied from a red --new wrote an estoppel entry under its seat (%d before, %d after)", seatID, estoppels, n))
		return
	}
	r.mu.Lock()
	r.readingsOverPrescriptions++
	r.mu.Unlock()
}

// estoppelEntries is the tool's estoppel log entries on the record written under seatID — the seat
// whose mint the tool refused. The lenses sit concurrently and each lens's estopped mint writes one
// (mintEstopped), so a count over every seat moves with whichever sibling lands inside the window.
func (r *runner) estoppelEntries(seatID string) int {
	b, ok := r.board()
	if !ok {
		return -1
	}
	n := 0
	for _, e := range b.Events {
		if l, isL := recordpb.BodyAs[*recordpb.Log](e); isL && e.GetSeatId() == seatID &&
			l.GetType() == recordpb.LogType_LOG_TYPE_ESTOPPEL && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
			n++
		}
	}
	return n
}

// readNotTheWorkFound is a lens's reading of the copy that is not the work as a source it found: it
// is recorded with no gap and no carrier, and the report's sources gain no row for it.
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
	v := r.checkReading(seatID, out, c.args)
	switch {
	case v == nil:
	case v.GetCarriedBy() != "" || v.CarrierMinted != nil:
		r.noteActErr(fmt.Sprintf("%s's found-source reading of a copy that is not the work is carried by %q", seatID, v.GetCarriedBy()))
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
