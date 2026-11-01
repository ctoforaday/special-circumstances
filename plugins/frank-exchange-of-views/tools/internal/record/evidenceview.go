package record

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE EVIDENCE VIEW — what backs the report, and what has been checked of it.
//
// # The anchor a seat could not resolve
//
// The report carries a tool-inserted anchor of every kind in the anchor kinds table. A seat
// reading a sentence sees the token. Only a finding's could ever be resolved — `show findings`
// is keyed by the f-label — and a source's and a computation's were opaque.
//
// Measured, in a seat's own words when asked what it could do next: "blocked by missing
// information (what does `c-29a72fe2` point to?)". The tool holds that citation's url, title,
// sha256, access date and the cached bytes themselves, and handed the seat an id.
//
// It is worse than an inconvenience for RED, whose terminal duty is verification at the leaf.
// The constitution tells red to re-read the exact bytes blue cited (`fetch --url <the cited
// url>`) — with no path from the anchor in front of it to that url. And `lens reproduce --id
// <sha256>` ASSERTED the way through: "every proof is listed in the report beside the sentence
// it backs". The report lists no sha. It carried an anchor id.
//
// This is the class that keeps paying out: the tool computes something, discards it, and asks
// the seat to reconstruct it.
//
// # Why one view and not two
//
// A citation and a computation are the same act at different strengths — blue offers evidence for
// a sentence, anchors it, and red checks it. Splitting them into two views would have grown the
// surface by two to answer one question ("what does this token point to"), and left a seat
// holding a `K-` id to guess which view took it. `citation-ledger` — red's
// verifications, and nothing of what blue cited — is subsumed here: anchored checks attach to the
// source they name, and red's own corroboration is the `independent` array.
//
// # Both halves now carry their join
//
// They did not. `lens verify` recorded a free-text `reference` rather than the anchor of the
// citation it checked, so this view listed red's work beside blue's sources without connecting
// them — and any join would have been a string-similarity guess presented as a fact, which reads
// identically to a right one whether it is right or not. #382 gave `verify` an `--anchor`, so a
// source now carries the verifications OF THAT SOURCE, and `verified: []` means nobody has
// checked it — stated, not inferred from an absence.
//
// Red's INDEPENDENT checks (corroboration it found itself, which blue never cited) have no
// anchor to join on and never will; they are the `independent` array, and that is a different
// fact from an unverified citation rather than a missing one.
//
// COMPUTATIONS carry the same join through `reproduce`'s `anchor`, and each act's comparison
// through it too: `rerun: not_rerun` means nobody executed the bundle, `rerun: not_compared` that the
// bundle was executed and nobody compared this act against it, `soundness: not_judged` that nobody
// judged this sentence. Every absence in this view is a stated field, because an unaudited thing
// and a clean thing are otherwise the same empty space.

// EvidenceSourceJSON is one source blue cited, keyed by the anchor token in the report.
type EvidenceSourceJSON struct {
	// Anchor is the c-<hex> exactly as it appears inside `<!--cite:…-->`. It is the lookup key
	// a seat holds when it reads the report, which is the whole reason this view exists.
	Anchor     string `json:"anchor"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Sha256     string `json:"sha256"`
	AccessDate string `json:"access_date"`
	// Location is the sentence the anchor sits at — what the source is offered as backing for.
	//
	// It is also the ONLY copy of that sentence. A cite carried it twice — `claim` re-quoted
	// what `location` already held to place the anchor — and Cite reserves field 8 to keep the
	// second spelling from coming back. There is no `claim` on a source here because there is
	// no second span to render; `Verify.claim` is a different field and is still on a
	// verification below.
	Location string `json:"location"`
	// BlueReason is blue's argument for the citation — why this source backs that sentence — from
	// `cite --reason`: BLUE'S WORDS ABOUT THE SOURCE, never the source's, and named so. A reason
	// that paraphrases a definition reads like the definition, and a lens that took one for the
	// source judged the sentence against blue's inverted paraphrase of it (universe-m30). Rendered on
	// every blue cite: "" is "blue offered no argument", which red reads when deciding what to
	// verify. Absent on a corroboration, whose reading is its check's. It is never printed in the
	// report.
	BlueReason *string `json:"blue_reason,omitempty"`
	SeatID     string  `json:"seat_id"`
	Epoch      int     `json:"epoch"`
	// SourceTextOrigin is where the cited text came from — embedded, ocr, none or not_recorded —
	// stamped by the tool. Empty on a corroboration, which red read itself.
	SourceTextOrigin string `json:"source_text_origin,omitempty"`
	// WorkStatus is what a maintained index says about the cited WORK — standing, retracted, or
	// not_checked. Red reads this view to decide what to verify, and a retraction is the one
	// thing here that no re-reading of the source can discover: the bytes are genuine and the
	// paper is withdrawn. It is listed for every source, including the ones the report's own
	// rendering stays silent about, because "nobody asked an index" is a lead for red and a
	// distraction for a reader of the subject.
	WorkStatus string `json:"work_status,omitempty"`
	// SourceCompleteness is which part of the work blue's copy is — full, abstract, unverified,
	// not_the_work, not_asked. A summary_only citation of an abstract and a leaf citation of the
	// paper differ in what red has to check, and only this field says which one it holds.
	SourceCompleteness string `json:"source_completeness,omitempty"`
	// CopySubstitution is WHOSE BYTES the citation read: `none` states they are the url's own,
	// and every other word names the route by which a copy stood in for it — followed_link,
	// recovered, fuller_copy, named_backend — with CopyURL naming where that copy was read.
	// not_recorded is a citation from before the field; not_cached, on a corroboration, a source
	// the run's cache held no copy of when red read it. Rendered on every source, because a lens
	// verifying against the url would otherwise check bytes nobody cited, and the blank that meant
	// "the url's own" would read the same as the blank that meant "nobody said".
	CopySubstitution string `json:"copy_substitution,omitempty"`
	CopyURL          string `json:"copy_url,omitempty"`
	// SourceText is WHAT THE SOURCE SAYS where blue quoted it: the span the tool found once in the
	// cached text, and the text around it, copied from the source when blue cited it. It is what a
	// lens reads the sentence against; blue_reason beside it is blue's account. Absent where the
	// tool holds no text for the source (source_text_origin says which), and on a citation recorded
	// before spans were located.
	SourceText *EvidenceSourceText `json:"source_text,omitempty"`
	// Pages are the PDF pages of an OCR reading the tool found the span on. A source with pages is
	// checked against a page image, not against the reading; OCREngine and OCRTextSha name the
	// reading the pages were found in.
	Pages      []int32 `json:"pages,omitempty"`
	OCREngine  string  `json:"ocr_engine,omitempty"`
	OCRTextSha string  `json:"ocr_text_sha,omitempty"`
	// CorroboratedBy names the red seat whose labelled corroboration placed this source's marker;
	// empty for a blue cite. A corroboration is a source of the report like any cite — the
	// assembler resolves its marker to a footnote — so it is listed here (gblock's ruling,
	// 2026-09-11; B9's evidence lens read its own marker as a citation with no source).
	CorroboratedBy string `json:"corroborated_by,omitempty"`

	// Verified is every `lens verify` naming THIS anchor. An empty slice is the honest zero and
	// it is rendered, not omitted: "nobody has checked this source" is what red reads to decide
	// where to spend its next pass, and an absent key would leave that to be inferred.
	Verified []EvidenceVerificationJSON `json:"verified"`
}

// EvidenceSourceText is the source's own text at a citation's span, as the tool found it.
type EvidenceSourceText struct {
	// Span is blue's span as typed; Matched is the same words as the source holds them (they
	// differ only by fetchcache.SpanNormalisation: whitespace, line-end hyphens, and typographic
	// quotes, dashes, ellipses, ligatures and soft hyphens). Before and After are the source's text
	// either side, up to a bounded window.
	Span    string `json:"span"`
	Before  string `json:"before"`
	Matched string `json:"matched"`
	After   string `json:"after"`
	// Start and End are byte offsets of Matched in the text whose sha256 is TextSha — the source's
	// text, or for OCR text the page's reading.
	Start   int32  `json:"start"`
	End     int32  `json:"end"`
	TextSha string `json:"text_sha"`
}

// sourceTextOf is a cite's located span as the evidence view shows it, or nil where it has none.
func sourceTextOf(c *recordpb.Cite) *EvidenceSourceText {
	if c.SpanStart == nil {
		return nil
	}
	return &EvidenceSourceText{
		Span: c.GetSourceSpan(), Before: c.GetSourceBefore(), Matched: c.GetSourceMatched(), After: c.GetSourceAfter(),
		Start: c.GetSpanStart(), End: c.GetSpanEnd(), TextSha: c.GetSpanTextSha(),
	}
}

// EvidenceComputationJSON is one computation blue recorded, keyed by its anchor, WITH red's comparison.
type EvidenceComputationJSON struct {
	// Anchor is the K-<hex> in the report and what `lens reproduce --anchor` takes; Sha256 is the
	// bundle's, which every computation on the same bundle shares.
	Anchor string   `json:"anchor"`
	Sha256 string   `json:"sha256"`
	Argv   []string `json:"argv"`
	Basis  string   `json:"basis"`
	// Cites is the citation anchor of the METHOD this computation applies, when blue named one —
	// the link between the two evidence classes, and the only one on the record.
	Cites string `json:"cites"`
	// Drift is WHAT MOVED when the tool's two runs of the bundle disagreed, and Invariant what blue
	// stated holds across them; both are empty on a reproducible computation.
	Drift     string `json:"drift"`
	Invariant string `json:"invariant"`
	SeatID    string `json:"seat_id"`
	Epoch     int    `json:"epoch"`
	// BundleBases is every basis the acts on this row's bundle were recorded with, where they differ:
	// one bundle, run twice by two acts, can move for one and not the other, and a reader judging
	// this act's comparison reads that the bundle's output is not settled. Absent where every act on
	// the bundle has this row's basis.
	BundleBases []string `json:"bundle_bases,omitempty"`

	// The bundle's files and this act's output are read at the computation's anchor, from the
	// store: see EvidenceComputationAtJSON. They are content, and every seat pulls this table.

	// THE VERDICT IS FLAT ON THE ROW, IN TWO HALVES (gblock, 2026-10-12: the anchor is about the
	// quote, the computation is about itself). It was one nested `verified` object keyed by the
	// sha, and it trapped readers twice over: two chair sittings filtered the row for
	// `{reproduced, sound}` and read nulls, and a lens's `unsound` on one sentence was overwritten by
	// another lens's `sound` on another sentence over the same bundle (#1358).
	//
	// Rerun is THIS ACT's comparison against its bundle's standing re-run: reproduced | differs |
	// calls_the_tool; not_compared where the bundle was executed and nobody compared this act;
	// not_rerun where nobody executed the bundle.
	Rerun string `json:"rerun"`
	// Soundness is THIS ANCHOR's sentence: sound | unsound | not_judged | reopened (an edit changed
	// the sentence after it was judged). A standing `unsound` from any seat reads unsound here until
	// that seat judges it again or an edit reopens the sentence.
	Soundness string `json:"soundness"`
	// Verified is the conjunction a reader asks for: this act's comparison reproduced AND this
	// sentence is judged sound.
	Verified bool `json:"verified"`
	// Checks is every `reproduce` naming this anchor, in record order, each marked whether it still
	// stands. An empty array is the honest "nobody judged this sentence".
	Checks []EvidenceReproductionJSON `json:"checks"`
}

// EvidenceReproductionJSON is one `reproduce` of one computation anchor — the two axes kept apart.
type EvidenceReproductionJSON struct {
	// Rerun is what this act's comparison found (COMPUTED by the tool): reproduced | differs |
	// calls_the_tool; Ran is whether the act executed the bundle, or compared against the bundle's
	// standing re-run. Soundness is RED'S JUDGEMENT of this anchor's sentence from reading the
	// program. A program that prints its conclusion reproduces forever, so the mechanical half alone
	// establishes nothing.
	Rerun     string `json:"rerun"`
	Ran       bool   `json:"ran"`
	Soundness string `json:"soundness"`
	// Standing is whether this verdict still stands on the anchor: it is its seat's latest, from a
	// sitting opened after the latest edit of the sentence.
	Standing bool   `json:"standing"`
	Note     string `json:"note"`
	SeatID   string `json:"seat_id"`
	Epoch    int    `json:"epoch"`
}

// EvidenceVerificationJSON is one `lens verify` — a claim red checked against a source.
type EvidenceVerificationJSON struct {
	Claim string `json:"claim"`
	// Anchor is the citation adjudicated, empty on an independent check.
	Anchor string `json:"anchor"`
	// Label is the footnote this reading MINTED, on a corroboration that supported the claim.
	// Empty on an anchored verify (it adjudicates a footnote blue already made) and on a reading
	// that did not back the claim — those splice nothing, by design.
	Label string `json:"label,omitempty"`
	// Outcome is what the source DID for the claim, and it has a negative half: `refutes` and
	// `absent` are the values that used to have nowhere to go, so the strongest finding on this
	// axis left as prose and the assembly screen looked for a verdict no field could carry.
	Outcome string `json:"outcome"`
	// Bridge is what a supports_with_bridge reading bridged — summary, secondary_citation or
	// near_restatement — absent on every other outcome, and on a reading recorded before the field.
	Bridge string `json:"bridge,omitempty"`
	// Confidence is how sure red is of that outcome — orthogonal to it. `refutes` at low
	// confidence is a call for more evidence; at high confidence it is a finding to act on.
	Confidence string `json:"confidence"`
	// Text is red's reading — required, because a verdict with nothing behind it is the
	// assertion the verb exists to replace.
	Text string `json:"text"`
	// URL and Title name the source red read. They replace `reference`, a free-text third
	// spelling of a source on the one verb that reads one red found itself: the corroboration a
	// run rests on was identified less precisely than the citations it audits. A source is named
	// by url + title here as it is on a cite.
	URL        string `json:"url"`
	Title      string `json:"title"`
	AccessDate string `json:"access_date"`
	SeatID     string `json:"seat_id"`
	Epoch      int    `json:"epoch"`
	// Page is the page image red checked for a citation with pages; PageRenderSha is that image's
	// sha256 and ReadingRenderSha the render the reading was made from. Equal hashes mean red
	// looked at the very pixels the reading came from.
	Page             int32  `json:"page,omitempty"`
	PageRenderSha    string `json:"page_render_sha,omitempty"`
	ReadingRenderSha string `json:"reading_render_sha,omitempty"`
	// WorkStatus is what a maintained index says about the work red read. It is here as well as
	// on the source row because a labelled corroboration is a source in its own right — red found
	// it, blue never cited it, and it renders in `independent`, which reads no source row at all.
	WorkStatus string `json:"work_status,omitempty"`
	// SourceCompleteness is which part of the work red's copy was. A `supports` read off an
	// abstract confirms what the abstract says, not what the study shows, and a reader of this
	// verdict needs to know which it is.
	SourceCompleteness string `json:"source_completeness,omitempty"`
	// CopySubstitution is WHOSE BYTES red's reading was made against — the copy the run's cache held
	// for the source when red recorded it: `none` the url's own, a substitution kind with CopyURL
	// where the copy was read, `not_cached` where the cache held none. A verdict on a recovered or
	// fuller copy is a verdict on that copy.
	CopySubstitution string `json:"copy_substitution,omitempty"`
	CopyURL          string `json:"copy_url,omitempty"`
}

// Refuted reports whether this verification found AGAINST the claim — the two outcomes that
// mean the report is carrying a citation it should not, when the reading stands. A named
// predicate rather than a string comparison at each call site: every reader of one reading must
// agree on what "found against" means, and they agree by calling this. Whether it STANDS is the
// anchor ledger's (CitationsFoundAgainst), which counts.sources_refuted reads.
func (v EvidenceVerificationJSON) Refuted() bool {
	return v.Outcome == "refutes" || v.Outcome == "absent"
}

// EvidenceJSON is the whole evidence layer of the report.
type EvidenceJSON struct {
	Sources      []EvidenceSourceJSON      `json:"sources"`
	Computations []EvidenceComputationJSON `json:"computations"`
	// Independent is red's corroboration — sources it went and found, which blue never cited and
	// which therefore have no anchor. A separate array rather than anchorless rows mixed into the
	// sources, because "checked something blue did not cite" and "checked a citation" answer
	// different questions and only one of them is about the report's own backing.
	Independent []EvidenceVerificationJSON `json:"independent"`
	// Unchecked are the citations and computations with no check of the sentence the report carries
	// (uncheckedAnchors): nobody has checked it, or an edit changed the sentence and no sitting
	// opened since has — the reference stands and its referent changed, so its check is of a
	// sentence the report no longer carries, which is stale, not refuted. It is the standing set: a
	// check removes the row. It holds a PASS (passBlockersOf).
	Unchecked []EvidenceUncheckedJSON `json:"unchecked"`
	// UnansweredContradictions are the claims where red read a source that CONTRADICTS or does
	// not support the report, and nothing has answered it: no gap was minted from the finding that
	// raised it, and the citation it was read against is neither retired nor backed
	// (unansweredContradictions).
	//
	// STATED HERE BECAUSE THIS IS WHERE RED LOOKS. The duty is enforced at the chair's PASS gate,
	// which is the right place to REFUSE — but a duty that only surfaces when someone else is
	// blocked at the end of the epoch is one the seat that owes it never sees. An empty array is
	// the honest "nothing outstanding"; without the field, nothing outstanding and nothing
	// checked are the same absence.
	//
	// EACH CARRIES ITS ID, and naming it is the only act that raises it (`finding --about-kind
	// contradiction --about <id>`), whether or not the report still holds the sentence. A finding
	// whose quote contains the claim raises nothing. A raised one carries the raising finding,
	// which the gap that answers it names in its found_by.
	UnansweredContradictions []EvidenceContradictionJSON `json:"unanswered_contradictions"`
	Counts                   struct {
		Sources      int `json:"sources"`
		Computations int `json:"computations"`
		// ComputationsUnverified is the count whose `verified` is false: the script not reproduced, or the
		// sentence not judged sound. It is stated because the honest zero and the unchecked case are
		// otherwise the same empty space.
		ComputationsUnverified int `json:"computations_unverified"`
		// SourcesUnverified is its citation twin: cited, and no reading of the sentence the report
		// carries has reached it — unchecked, or checked and nothing read (anchorStanding.state).
		SourcesUnverified int `json:"sources_unverified"`
		// SourcesRefuted counts the citations red finds against NOW: a standing `refutes` or
		// `absent` check of the sentence the report carries (CitationsFoundAgainst, the anchor
		// ledger's fold). One count, one meaning: the assembly screen fails on the same set, so a
		// refutation of a sentence an edit replaced and a re-check backed is not counted (gblock,
		// 2026-10-19; m30's C-4f4b8cfa).
		SourcesRefuted int `json:"sources_refuted"`
		Verifications  int `json:"verifications"`
	} `json:"counts"`
}

// EvidenceContradictionJSON is one contradiction no gap answers. ID is its id — the
// events.id of the reading that opened it, the value `--about` takes; Claim is the sentence the
// reading was of; Reader is the seat whose reading last joined it, which owes the finding and the
// gap; Anchor is the citation that reading checked, "" for a corroboration; RaisedBy is the finding
// that raised it, "" until one has — the id the answering gap's found_by names.
type EvidenceContradictionJSON struct {
	ID       string `json:"id"`
	Claim    string `json:"claim"`
	Reader   string `json:"reader"`
	Anchor   string `json:"anchor"`
	RaisedBy string `json:"raised_by"`
}

// EvidenceUncheckedJSON is one unchecked citation or computation. Edit and LastCheck are events.id: the
// edit that changed its sentence, 0 when none did, and its latest check, 0 when it has none, which
// does not clear it.
type EvidenceUncheckedJSON struct {
	Anchor    string `json:"anchor"`
	Edit      int64  `json:"edit"`
	LastCheck int64  `json:"last_check"`
}

// EvidenceJSONOf projects the record's evidence layer in event order.
func EvidenceJSONOf(evs []*Event, win WindowIndex) EvidenceJSON {
	out := EvidenceJSON{
		Sources:      []EvidenceSourceJSON{},
		Computations: []EvidenceComputationJSON{},
		Independent:  []EvidenceVerificationJSON{},
	}
	out.UnansweredContradictions = []EvidenceContradictionJSON{}
	for _, c := range unansweredContradictions(evs, win) {
		out.UnansweredContradictions = append(out.UnansweredContradictions, EvidenceContradictionJSON{ID: c.idWord(), Claim: c.claim, Reader: c.reader, Anchor: c.anchor, RaisedBy: c.raisedBy})
	}
	out.Unchecked = []EvidenceUncheckedJSON{}
	for _, u := range uncheckedAnchors(evs, win) {
		out.Unchecked = append(out.Unchecked, EvidenceUncheckedJSON{Anchor: u.anchor, Edit: u.edit, LastCheck: u.lastCheck})
	}

	// Red's verifications, split by whether they name a citation. The anchored ones are indexed
	// so each source carries its own; the rest are corroboration and stand alone.
	byAnchor := map[string][]EvidenceVerificationJSON{}
	var corroborations []EvidenceSourceJSON
	for _, e := range evs {
		w := win.Of(e)
		// BodyAs returns false for BOTH no body and a body of another type. Neither is a
		// verification, and neither is rendered as a check with every field blank.
		vf, ok := recordpb.BodyAs[*recordpb.Verify](e)
		if !ok {
			continue
		}
		// OUTCOME AND CONFIDENCE ARE ENUMS NOW, AND THIS VIEW RENDERS THE SEAT'S WORD. Every
		// consumer — Refuted(), the assembly screen, a seat reading the JSON — speaks
		// `supports`/`refutes`/`high`, so the value's spelling is what crosses the boundary, not
		// `SOURCE_OUTCOME_REFUTES`. `Word` maps the UNSPECIFIED zero back to "" for the same
		// reason the old read did: the pre-migration record had no outcome word for an absent
		// outcome, and `unspecified` is a word no seat ever typed.
		//
		// THE HYPHEN CAVEAT IS RESOLVED. `supports-with-bridge` was spelled with a hyphen on the
		// seat-facing surfaces while `Word` derives `supports_with_bridge`, so the help offered a
		// word the write path refused. The declared set carries the schema's spelling now, and
		// enums_test asserts that every declared value is one the record can hold.
		v := EvidenceVerificationJSON{
			Claim:      vf.GetClaim(),
			Anchor:     vf.GetAnchor(),
			Label:      vf.GetLabel(),
			Outcome:    recordpb.Word(vf.GetOutcome()),
			Bridge:     recordpb.Word(vf.GetBridge()),
			Confidence: recordpb.Word(vf.GetConfidence()),
			Text:       vf.GetText(),
			URL:        vf.GetUrl(),
			Title:      vf.GetTitle(),
			AccessDate: vf.GetAccessDate(),
			SeatID:     e.GetSeatId(),
			Epoch:      w.Epoch,

			Page:             vf.GetPage(),
			PageRenderSha:    vf.GetPageRenderSha(),
			ReadingRenderSha: vf.GetReadingRenderSha(),
			WorkStatus:       recordpb.Word(vf.GetWorkStatus()),

			SourceCompleteness: recordpb.Word(vf.GetSourceCompleteness()),
			CopySubstitution:   recordpb.Word(vf.GetCopySubstitution()),
			CopyURL:            vf.GetCopyUrl(),
		}
		out.Counts.Verifications++
		// THE SPLIT IS ON THE ANCHOR, and the empty string is deliberate. A verification indexed
		// under "" would join no source and disappear from both arrays; sending it to
		// `independent` keeps it visible.
		// A LABELLED VERIFY WITH A CLAIM PLACED A MARKER in the report — it is the corroboration
		// its Anchor follows — and the assembler resolves it to a footnote like any cite. So it is a SOURCE
		// of the report as well as a check: listed with the cites below, marked as red's
		// corroboration, and still under independent[] where the check itself lives.
		if vf.GetLabel() != "" && vf.GetClaim() != "" {
			corroborations = append(corroborations, EvidenceSourceJSON{
				Anchor: vf.GetLabel(), URL: vf.GetUrl(), Title: vf.GetTitle(), Sha256: vf.GetSha256(), AccessDate: vf.GetAccessDate(),
				Location: vf.GetClaim(), SeatID: e.GetSeatId(), Epoch: w.Epoch,
				CorroboratedBy: e.GetSeatId(),
				// A corroboration's footnote rests on the copy red read, as a cite's rests on blue's.
				CopySubstitution: v.CopySubstitution, CopyURL: v.CopyURL,
				// ITS OWN CHECK COMES FIRST: a corroboration is red reading the source, so the source
				// is not "unverified" — the lens is not sent to check its own reading again.
				Verified: []EvidenceVerificationJSON{v},
			})
		}
		if v.Anchor == "" {
			out.Independent = append(out.Independent, v)
			continue
		}
		byAnchor[v.Anchor] = append(byAnchor[v.Anchor], v)
	}

	// Red's checks of each computation: the bundle's executions by sha, each act's comparison and its
	// sentence's verdicts by anchor, and each bundle's bases.
	ledger := anchorChecksOf(evs, win)
	reruns, acts := bundleRerunsOf(evs), actVerdictsOf(evs)
	bases := bundleBasesOf(evs)

	for _, e := range evs {
		w := win.Of(e)
		// TWO ARMS, so this stays `Body` plus a type switch rather than two `BodyAs` passes:
		// one walk of the events, and an event is a cite or a computation or neither.
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch bd := body.(type) {
		case *recordpb.Cite:
			// A cite with no label names no anchor, so it is not a citation a reader can reach.
			//
			// THIS IS NO LONGER "the label is what tells blue from red". It was, and citationid.go
			// used the same discriminator; since red's supporting corroborations mint a label too,
			// the two are told apart by EVENT TYPE (a Cite here, a Verify below) rather than by
			// whether a label is present. The filter stays for its own reason, stated above.
			label := bd.GetLabel()
			if label == "" {
				continue
			}
			// An empty slice, never nil: `"verified": []` is "nobody has checked this source",
			// which is a fact red acts on. A null would leave it to be inferred.
			checks := byAnchor[label]
			if checks == nil {
				checks = []EvidenceVerificationJSON{}
			}
			out.Sources = append(out.Sources, EvidenceSourceJSON{
				Anchor:     label,
				URL:        bd.GetUrl(),
				Title:      bd.GetTitle(),
				Sha256:     bd.GetSha256(),
				AccessDate: bd.GetAccessDate(),
				Location:   bd.GetLocation(),
				BlueReason: proto.String(bd.GetText()),
				SeatID:     e.GetSeatId(),
				Epoch:      w.Epoch,
				Verified:   checks,

				SourceTextOrigin:   recordpb.Word(bd.GetSourceTextOrigin()),
				WorkStatus:         recordpb.Word(bd.GetWorkStatus()),
				SourceCompleteness: recordpb.Word(bd.GetSourceCompleteness()),
				CopySubstitution:   recordpb.Word(bd.GetCopySubstitution()),
				CopyURL:            bd.GetCopyUrl(),
				SourceText:         sourceTextOf(bd),
				Pages:              bd.GetPages(),
				OCREngine:          bd.GetOcrEngine(),
				OCRTextSha:         bd.GetOcrTextSha(),
			})
		case *recordpb.Computation:
			// The anchor's verdict and its comparison are read by anchor; the bundle's executions by
			// the bundle_sha the record spells once, on the Computation and on each Reproduce of it.
			s := ledger.at(bd.GetComputationId())
			v := computationVerdictOf(bd.GetComputationId(), bd.GetBundleSha(), s, reruns, acts)
			checks := []EvidenceReproductionJSON{}
			if s != nil {
				standing := map[int64]bool{}
				for _, c := range v.judgements {
					standing[c.id] = true
				}
				for _, c := range s.checks {
					r, _ := recordpb.BodyAs[*recordpb.Reproduce](c.ev)
					checks = append(checks, EvidenceReproductionJSON{
						Rerun: recordpb.Word(r.GetRerun()), Ran: r.GetRan(), Soundness: recordpb.Word(c.soundness), Standing: standing[c.id],
						Note: r.GetNote(), SeatID: c.seat, Epoch: win.Of(c.ev).Epoch,
					})
				}
			}
			argv := bd.GetArgv()
			if argv == nil {
				argv = []string{}
			}
			out.Computations = append(out.Computations, EvidenceComputationJSON{
				Anchor:    bd.GetComputationId(),
				Sha256:    bd.GetBundleSha(),
				Argv:      argv,
				Basis:     recordpb.Word(bd.GetBasis()),
				Cites:     bd.GetCites(),
				Drift:     bd.GetDrift(),
				Invariant: bd.GetInvariant(),
				SeatID:    e.GetSeatId(),
				Epoch:     w.Epoch,
				BundleBases: func() []string {
					if b := bases[bd.GetBundleSha()]; len(b) > 1 {
						return b
					}
					return nil
				}(),
				Rerun:     v.rerunWord(),
				Soundness: v.soundness,
				Verified:  v.verified(),
				Checks:    checks,
			})
			// `verify` is handled in the indexing pass above — it has to be, because a source
			// must carry its verifications and a cite event may arrive after the check of it.
		}
	}
	// Red's corroborations are sources too, after blue's cites, each with every later check naming it.
	for _, c := range corroborations {
		c.Verified = append(c.Verified, byAnchor[c.Anchor]...)
		out.Sources = append(out.Sources, c)
	}

	out.Counts.Sources = len(out.Sources)
	out.Counts.Computations = len(out.Computations)
	for _, p := range out.Computations {
		if !p.Verified {
			out.Counts.ComputationsUnverified++
		}
	}
	for _, s := range ledger.order {
		if st := s.state(); !s.computation && (st == stateUnchecked || st == stateUnread) {
			out.Counts.SourcesUnverified++
		}
	}
	out.Counts.SourcesRefuted = len(CitationsFoundAgainst(evs, win))
	return out
}

// evidenceView is the evidence read from the record: the sources, computations and their checks, the
// findings that raise a contradicting check and the mints that answer it (unanswered_contradictions), and the edits that
// change an anchor's sentence with the retires that take one out of the report (unchecked).
var evidenceView = declareNarrowedView("evidence", rendersEvents(EvidenceJSONOf),
	recordpb.EventType_EVENT_TYPE_CITE,
	recordpb.EventType_EVENT_TYPE_VERIFY,
	recordpb.EventType_EVENT_TYPE_COMPUTATION,
	recordpb.EventType_EVENT_TYPE_REPRODUCE,
	recordpb.EventType_EVENT_TYPE_FINDING,
	recordpb.EventType_EVENT_TYPE_MINT,
	recordpb.EventType_EVENT_TYPE_BLUE_EDIT,
	recordpb.EventType_EVENT_TYPE_RETIRE)

// EvidenceJSONBytes renders the evidence view as indented JSON.
func EvidenceJSONBytes(run Run) ([]byte, error) { return evidenceView.jsonBytes(run) }

// EvidenceAtJSON is the evidence view read AT ONE ANCHOR — the entry a seat holding that token is
// asking about, and nothing else on the table.
//
// IT IS A SECOND SHAPE BECAUSE IT ANSWERS A SECOND QUESTION. The table says what backs the report
// and what has been checked of it; this says what ONE anchor points at, in full. Narrowing the
// table's own arrays to one row would leave its run-wide counts and its unanswered contradictions
// beside a single entry, where "0 computations unverified" reads as a fact about that entry.
type EvidenceAtJSON struct {
	Anchor string `json:"anchor"`
	// Source is set when the anchor is a citation's (or a labelled corroboration's), Computation
	// when it is a computation's. Exactly one is present.
	Source      *EvidenceSourceJSON        `json:"source,omitempty"`
	Computation *EvidenceComputationAtJSON `json:"computation,omitempty"`
	// Unchecked is the anchor's row of the table's `unchecked`, absent when it owes no check — the
	// PASS refusal sends a seat here for an unchecked anchor, and this is where it reads which edit
	// and which check.
	Unchecked *EvidenceUncheckedJSON `json:"unchecked,omitempty"`
}

// EvidenceComputationAtJSON is a computation's row WITH THE COMPUTATION ITSELF: the bundle the lens's
// `reproduce` executes and the output this act recorded.
//
// THE RECORD HOLDS NEITHER, AND THIS DOES NOT CHANGE THAT. They are content, addressed by the
// bundle_sha and the computation_id the `Computation` event carries and read from the store when a
// seat asks — the same join the assembler makes for the human's evidence document. They are here
// and not on the table's rows because every seat pulls the table, and a run's programs in every
// context is a cost the read at one anchor does not impose.
//
// PLAIN STRINGS, ALWAYS PRESENT. An artifact the store does not hold is refused before this type
// is filled (compute.Stored), so "" here means the stored bytes are empty and never that nobody
// could find them.
type EvidenceComputationAtJSON struct {
	EvidenceComputationJSON
	// Files is the bundle, each file by its path in the bundle.
	Files []EvidenceFileJSON `json:"files"`
	// Output is what this act's first run wrote.
	Output string `json:"output"`
}

// EvidenceFileJSON is one file of a computation's bundle.
type EvidenceFileJSON struct {
	Path string `json:"path"`
	Body string `json:"body"`
}

// EvidenceAt resolves one document anchor to the evidence entry it names. A computation comes back
// without its bundle: the artifact is the caller's to read, because the store is not the record.
//
// A MISS IS AN ERROR. The table answers "nothing here" with an empty array, which is right for a
// table and wrong for a lookup: a seat that asked what one anchor points at and received an empty
// document would read it as an anchor that points at nothing.
func EvidenceAt(run Run, id string) (EvidenceAtJSON, error) {
	ev, err := evidenceView.of(run)
	if err != nil {
		return EvidenceAtJSON{}, err
	}
	at := EvidenceAtJSON{Anchor: id}
	for i := range ev.Unchecked {
		if ev.Unchecked[i].Anchor == id {
			at.Unchecked = &ev.Unchecked[i]
		}
	}
	for i := range ev.Computations {
		if ev.Computations[i].Anchor == id {
			at.Computation = &EvidenceComputationAtJSON{EvidenceComputationJSON: ev.Computations[i]}
			return at, nil
		}
	}
	for i := range ev.Sources {
		if ev.Sources[i].Anchor == id {
			at.Source = &ev.Sources[i]
			return at, nil
		}
	}
	// A WELL-FORMED ID NOTHING HERE IS AT — a finding's, a gap's, or a citation or computation this
	// record does not hold. An id of the wrong shape, a computation's sha256 among them, never reaches this: the
	// flag that takes the anchor refuses it (flags.AnchorID).
	return EvidenceAtJSON{}, fmt.Errorf("no source or computation is at anchor %q on this record (%d source(s), %d computation(s)). "+
		"--anchor takes the id inside a citation or computation token in the report, exactly as it reads there; the bare projection lists every one with its anchor",
		id, len(ev.Sources), len(ev.Computations))
}

// bundleBasesOf is every bundle's bases, keyed by bundle_sha: each basis an act on it was recorded
// with, sorted.
func bundleBasesOf(evs []*Event) map[string][]string {
	seen := map[string]map[string]bool{}
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Computation](e); ok && c.GetBundleSha() != "" {
			if seen[c.GetBundleSha()] == nil {
				seen[c.GetBundleSha()] = map[string]bool{}
			}
			seen[c.GetBundleSha()][recordpb.Word(c.GetBasis())] = true
		}
	}
	out := map[string][]string{}
	for sha, b := range seen {
		for w := range b {
			out[sha] = append(out[sha], w)
		}
		sort.Strings(out[sha])
	}
	return out
}
