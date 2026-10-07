// Package bluedoc holds the checks that decide whether a span replacement against
// blue/report.md is LEGAL — shared, because two roles now need the same answer.
//
// WHY IT EXISTS. `blue edit` has always validated its own old→new pair: the span must be
// present, unique, must not split a word, and must not change which immortal anchors exist.
// With #267 stage 3 red may attach a CONCRETE proposed fix to a gap, and a proposal red
// cannot state legally is a proposal blue cannot apply — so the same checks have to run at
// mint time, in the chair role.
//
// The alternative was `internal/cli/merge` importing `internal/cli/blue`, which makes two
// role packages depend on each other for a rule that belongs to neither: it belongs to the
// DOCUMENT. A second copy of the checks was never an option — the anchor invariant is the
// one thing standing between an edit and red's immortal audit record, and this repo has
// already paid for two readers of one rule more than once.
//
// What did NOT move: the splice hygiene (tidySeam) and the write path. Those are what blue
// does when APPLYING an edit; these are what makes an edit legal in the first place, and
// only the second question has two askers.
package bluedoc

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
)

// ErrMisQuote is the sentinel for "the old span is not present". `blue edit` distinguishes
// it to drive its crash-reconcile branch (old gone but new already present ⇒ the write
// landed), so it must stay identifiable rather than collapse into a generic error.
var ErrMisQuote = errors.New("the quoted span was not found in report.md — quote the EXACT current text you are replacing. Runs of whitespace and the invisible anchor layer are ignored, and a span MAY cross blank lines; every other character must match")

// LocateUnique resolves the one span `old` names, or explains why it cannot.
//
// verb prefixes every message, because the seat's only teacher is the error text and a
// chair seat told "blue edit: …" learns the wrong command.
//
// AMBIGUITY IS REFUSED, NOT GUESSED. Taking the first of several matches silently edits a
// site the author may not have meant — and blue is explicitly told to propagate corrections
// to every site stating a claim, so repeated text is the EXPECTED shape of a real report.
func LocateUnique(verb, report, old string) (int, int, error) {
	// AN EDIT MAY CROSS A PARAGRAPH BREAK; an anchor may not. Sharing one rule made this verb's
	// own two refusals jointly unsatisfiable — see anchortext.SpanScope for the measurement.
	start, end, err := anchortext.LocateOnce(report, old, anchortext.CrossParagraphs)
	switch err {
	case nil:
		return start, end, nil
	case anchortext.ErrMisQuote:
		return 0, 0, fmt.Errorf("%s: %w", verb, ErrMisQuote)
	case anchortext.ErrAmbiguous:
		return 0, 0, fmt.Errorf("%s: your quoted span appears MORE THAN ONCE in report.md, so the target is ambiguous — quote more surrounding context to pick out the one site you mean (to change every site, make one edit per site)", verb)
	}
	return 0, 0, fmt.Errorf("%s: %w", verb, anchortext.ErrSplitsWord)
}

// LocateUniqueReplacing is LocateUnique for a caller that intends to REPLACE the span it finds.
//
// THE ANCHOR RULE BELONGS TO REPLACEMENT, NOT TO LOCATION. Baked into LocateUnique it also fired
// on `merge mint --quote`, which names the sentence a defect LIVES AT and rewrites nothing — so
// minting a gap about any already-anchored sentence was refused, with a message that spoke of
// "the text you are replacing". Caught by reading a regenerated golden that had recorded
// `minted G1` turning into `exit 2`, which is what the read-every-diff rule is for.
func LocateUniqueReplacing(verb, report, old string) (int, int, error) {
	start, end, err := LocateUnique(verb, report, old)
	if err != nil {
		return 0, 0, err
	}
	end, err = settleAbuttingAnchor(verb, report, old, end)
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

// LocateLiteral resolves `old` BYTE FOR BYTE: no whitespace folding, no skipping of the anchor
// layer, and no trimming of trailing punctuation. It is the locate behind an edit recorded with
// exact_span, and the one replay uses for it, so the two cannot disagree about where it lands.
//
// It answers `found` rather than erroring on a miss or a repeat because its callers use it as a
// FALLBACK — the ordinary locate has already succeeded, and what they need to say about a
// literal quote that does not stand in is part of a larger message. count is how many times the
// literal quote occurs; a span is returned only when that is exactly one and the span splits no
// word. A replay caller treats anything else as the loud failure it is.
//
// The quote's LEADING AND TRAILING WHITESPACE is not part of the span, exactly as in the ordinary
// locate: a seat that copies a line with its newline means the line, and a literal span that
// swallowed the newline would join it to the next one.
func LocateLiteral(report, old string) (start, end, count int, ok bool) {
	old = strings.TrimSpace(old)
	if old == "" {
		return 0, 0, 0, false
	}
	count = strings.Count(report, old)
	if count != 1 {
		return 0, 0, count, false
	}
	start = strings.Index(report, old)
	end = start + len(old)
	if !anchortext.SpanBoundaryOK(report, start, end) {
		return 0, 0, count, false
	}
	return start, end, count, true
}

// requireAbuttingAnchor refuses a quote that stops JUST SHORT of the anchor attached to the text
// it is replacing.
//
// THE MARKERS ARE THE MECHANISM, and they are visible for this reason. `show report` prints
// anchors as they are, so a seat rewriting a sentence can see the token sitting in it and copy it
// into --new the same way it copies every other character. That is the whole model: an edit
// mimics how you edit any document — quote what is there, write what should be there.
//
// The tolerance broke it. normalizeQuote SKIPS annotation spans, so a quote that omits the marker
// still matches — and the span it locates then ENDS BEFORE the marker. Measured: the transit
// guard never fires on a whole-sentence edit (the commonest edit there is), the marker is stranded
// beside prose it was never placed against, and tidySeam cannot even collapse the doubled
// terminator because the marker sits between the two halves.
//
// Only the ABUTTING case is refused. A fragment edit inside a sentence does not touch what the
// anchor is attached to; the marker keeps its position and `reopened` records that its sentence
// moved. What is refused is rewriting the text an anchor is ON while pretending the anchor is not
// there.
func settleAbuttingAnchor(verb, report, quoted string, end int) (int, error) {
	// Trailing punctuation the quote legitimately omitted sits between the span and the marker:
	// InsertAnchor places the token BEFORE the terminator, so skip that run first. TrimLeft takes
	// a rune cutset — `…` is three bytes, and a byte-wise skip would half-consume it.
	tail := report[end:]
	after := strings.TrimLeft(tail, anchortext.TrailingPunct)
	// THE RUN, NOT THE FIRST TOKEN. Two lenses anchoring one sentence is an ordinary corpus shape
	// — `verification<!--fx:f-e4bc25ec--><!--fx:f-73a56bd3-->` is from a real report — and this
	// consumed one token deep. The seat then quoted the sentence exactly as `show report` prints
	// it, carried BOTH markers into --new as instructed, and was told the second one was an
	// INVENTION: it sat past the extended span, so AnchorsTransitUnchanged saw it appear from
	// nowhere. Following its own instruction was the thing that got it refused.
	run := anchor.SkipRun(after, 0)
	if run == 0 {
		return end, nil
	}
	tok := after[:run]

	// THE SEAT QUOTED IT: EXTEND THE SPAN TO COVER IT.
	//
	// normalizeQuote drops annotation spans from the quote as well as from the report, so a seat
	// that copied the sentence EXACTLY as `show report` prints it still locates a span ending
	// before the marker. The quote is the evidence of intent: if the token is in it, the seat
	// means to replace the text the anchor sits on, so the span swallows the punctuation run and
	// the token. AnchorsTransitUnchanged then sees the anchor and requires it in --new, and the
	// terminator goes with the replacement instead of being stranded past the marker — which is
	// what produced `now.<!--cite:c-…-->.`
	if strings.Contains(quoted, tok) {
		return end + (len(tail) - len(after)) + run, nil
	}

	// IT DID NOT: refuse, and print the token to carry, naming every anchor of the run.
	var held []string
	for _, id := range anchor.IDs(tok) {
		held = append(held, anchor.Label(id))
	}
	return 0, fmt.Errorf("%s: the text you are replacing carries %s, and your quote stops just before it. "+
		"That anchor is ON this sentence: rewriting the sentence without it strands the reference beside prose it was never placed against. "+
		"Quote the sentence AS `show report` PRINTS IT — anchors included — and carry %s into --new unchanged. "+
		"To change the words around it and leave the anchor where it is, quote a FRAGMENT that does not reach it",
		verb, strings.Join(held, " and "), tok)
}

// AnchorsTransitUnchanged enforces the one anchor invariant a replacement must satisfy: it
// may not change WHICH anchors exist. The multiset of anchor ids in the replaced span must
// equal the multiset in its replacement — so an anchor may be carried across an edit (and
// the prose around it rewritten), but never introduced, dropped or duplicated. It returns newText
// with every anchor AutoPlace put back, so the refusal of one it could not is true at every caller.
//
// Anchors are placed by the tool, never typed into a replacement, and leave only by retire.
// Transit is not authorship: the tool checks the bytes, so nothing is delegated to the model.
func AnchorsTransitUnchanged(verb, oldSpan, newText string) (string, error) {
	newText = AutoPlace(oldSpan, newText)
	count := func(s string) map[string]int {
		m := map[string]int{}
		for _, id := range claimcount.ProtectedAnchorIDs(s) {
			m[id] = strings.Count(s, anchor.Token(id))
		}
		return m
	}
	o, n := count(oldSpan), count(newText)
	for _, id := range anchor.IDs(oldSpan) { // in reading order, so a refusal names the first anchor dropped
		switch want, got := o[id], n[id]; {
		case got == 0:
			sent, _ := sentenceAround(oldSpan, anchor.Token(id))
			on, near := fmt.Sprintf("on the sentence %q", sent), ""
			if sent == "" {
				on = "bare of any sentence"
			}
			if s := nearestSentence(newText, sent); s != "" {
				near = fmt.Sprintf(" — its nearest sentence there reads %q", s)
			}
			return "", fmt.Errorf("%s: your old span carries %s %s, and the replacement neither carries it nor keeps that sentence word for word exactly once, so the tool cannot put it back. "+
				"Place %s where that claim now stands in the replacement%s. To take the claim itself out, make the replacement that anchor alone and then retire the claim with blue's `retire` — the retire takes the anchor out with it, and where the claim was a clause inside a sentence, name the anchor to the retire with --anchor",
				verb, anchor.Label(id), on, anchor.Token(id), near)
		case got != want:
			return "", fmt.Errorf("%s: %s appears %d time(s) in the old span but %d in the replacement — an anchor may not be duplicated or removed by an edit; carry each one across exactly once", verb, anchor.Label(id), want, got)
		}
	}
	for id, got := range n {
		if o[id] == 0 {
			return "", &ErrAnchorIntroduced{Verb: verb, ID: id, Count: got}
		}
	}
	return newText, nil
}

// ErrAnchorIntroduced names the anchor a replacement invented.
//
// IT IS TYPED BECAUSE THE GENERIC MESSAGE SENDS A SEAT IN A CIRCLE. Reproducing an anchor is
// mandatory when it sits INSIDE the replaced span and refused when it sits just outside, and a
// seat cannot see which: a quote's trailing punctuation is trimmed before the span is located,
// so the anchor before a sentence's final period — 40 of the 43 in the archived corpus — falls
// outside a span whose quote appeared to contain it. A seat that meets one refusal and follows
// its instruction meets the other. A caller holding the surrounding document can tell the two
// apart and say so; this type is what lets it.
type ErrAnchorIntroduced struct {
	Verb  string
	ID    string
	Count int
}

func (e *ErrAnchorIntroduced) Error() string {
	return fmt.Sprintf("%s: your replacement introduces %s, which was not in the span it replaces — anchors are placed by the tool, never typed into a replacement (got %d occurrence(s))", e.Verb, anchor.Label(e.ID), e.Count)
}

// MaxProposalGrowth bounds how much longer a CONCRETE proposed fix may be than the span it
// replaces, in characters.
//
// THIS IS THE LINE BETWEEN AN AUDIT AND AUTHORSHIP, and it is enforced rather than advised.
// Red may propose exact text for a TEXTUAL defect — an overclaim, a wrong figure, a
// contradiction — where compliance genuinely is the right answer and costs blue nothing.
// The moment red may hand over concrete text for a SUBSTANTIVE addition, blue becomes a
// typist by incentive: applying is instant and free, while a counter-edit costs an epoch and
// invites re-audit. The seat contract inverts, quietly, and the record still looks healthy.
//
// The number is measured, not chosen. Across the 2026-08-04 smoke's 26 recorded edits, every
// change red prescribed that was a genuine ADDITION landed at +285 characters or more
// ("acknowledge shared definition…" +823, "add mitigations" +623, "replace premature closure"
// +566, "answer actual costs" +285), while textual repairs clustered at +60 and below —
// rewordings, deletions, and punctuation. 120 sits well clear of both: generous enough for a
// rewording or an inserted qualifying clause, far below where authoring began in the sample.
//
// HONEST LIMIT: that is 26 edits from one trivial question, which is thin. The bound is set
// TIGHT on purpose because the two failure directions are not symmetric — too tight and red
// falls back to prose (`fix_basis: proposed`, exactly the status quo, no harm done); too
// loose and red writes blue's report while the decline rate quietly goes to zero.
const MaxProposalGrowth = 120

// ValidateProposal decides whether a CONCRETE proposed fix is one blue could actually apply,
// and one red is entitled to propose. It is the whole mechanism behind `fix_basis: verified`:
// the basis is DERIVED from passing this, never asserted by the seat that would benefit from
// the claim.
//
// Passing it means red read the real document — an exact, unique, non-word-splitting span
// cannot be written from memory of what the report probably says. That forced read is the
// point: all three of the smoke's round-2 gaps were contradictions between blue's new text
// and text red never re-read before prescribing.
func ValidateProposal(verb, report, old, new string) error {
	if old == new {
		return fmt.Errorf("%s: the proposed old and new text are identical — there is no change to propose", verb)
	}
	// REPLACING: red's proposal is text blue will apply verbatim, so it owes the same duty an
	// edit does — carry the anchors on the span it rewrites.
	start, end, err := LocateUniqueReplacing(verb, report, old)
	if err != nil {
		return err
	}
	if _, err := AnchorsTransitUnchanged(verb, report[start:end], new); err != nil {
		return err
	}
	if grew := utf8.RuneCountInString(new) - utf8.RuneCountInString(old); grew > MaxProposalGrowth {
		return fmt.Errorf("%s: a concrete proposal may add at most %d characters to the span it replaces, and this adds %d — that size is a substantive ADDITION, which is blue's to author. State it as prose in --fix instead: red says what is wrong and what must be true, blue writes the report",
			verb, MaxProposalGrowth, grew)
	}
	return nil
}

// ReopenedAnchors returns the anchors whose SENTENCE this edit changed — the referents that
// moved under their references.
//
// THE TWO CONCERNS ARE SEPARATE AND ONLY ONE WAS ENFORCED. An anchor is never lost:
// AnchorsTransitUnchanged refuses a replacement that drops one, and droppedMarker backstops the
// whole edit. That promise holds. But an anchor SURVIVING onto rewritten prose is a citation
// backing a sentence nobody read, and nothing said so — measured: a cite on "The sky is blue and
// the grass is green" followed the text to "The sky is green and the grass is on fire".
//
// COMPARING SENTENCES, NOT SPANS, IS THE POINT. The obvious implementation asks which anchors sat
// inside the replaced span, and it would find almost none: InsertAnchor places a marker BEFORE the
// terminal punctuation, and normalizeQuote trims trailing punctuation off the quote — so a
// whole-sentence edit locates a span that ENDS just before the anchor. The anchor is adjacent,
// not contained, which is why AnchorsTransitUnchanged does not fire on the commonest edit there
// is. Reading the sentence around each anchor sidesteps the boundary question entirely: if the
// words around the reference changed, the reference needs looking at again, however the offsets
// happened to fall.
func ReopenedAnchors(before, after string) []string {
	bare := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(after) {
		bare[id] = true
	}
	var out []string
	for _, id := range claimcount.ProtectedAnchorIDs(before) {
		b, okB := sentenceAround(before, anchor.Token(id))
		a, okA := sentenceAround(after, anchor.Token(id))
		// An anchor that is GONE from `after` is not reopened — it is dropped, which is a refusal
		// the caller has already made. Saying both would report one fault as two.
		if !okB || !okA {
			continue
		}
		// An anchor LEFT BARE is not reopened either: its sentence is gone, not moved, so there is
		// nothing for red to re-verify. It is a claim on its way out through `retire`, which
		// takes the anchor with it — and if the retire never comes, the claim count has already
		// fallen with no retire behind it, which is the loss detector's to report.
		if bare[id] {
			continue
		}
		if b != a {
			out = append(out, id)
		}
	}
	return out
}

// sentenceAround returns the sentence holding tok, as anchor.Sentences splits the document, with
// every anchor token stripped, so a SECOND anchor arriving in the same sentence does not read as
// the first one's referent changing.
func sentenceAround(doc, tok string) (string, bool) {
	a, b := sentenceOf(doc, strings.Index(doc, tok), len(tok))
	if a < 0 {
		return "", false
	}
	return flatText(doc[a:b]), true
}

// sentenceOf is the bounds of the sentence holding doc[i:i+n], or -1, -1.
func sentenceOf(doc string, i, n int) (int, int) {
	for _, sp := range anchor.Sentences(doc) {
		if sp[0] <= i && i+n <= sp[1] {
			return sp[0], sp[1]
		}
	}
	return -1, -1
}

// flatText is s with its anchors stripped and its whitespace runs read as one space.
func flatText(s string) string {
	return strings.Join(strings.Fields(claimcount.StripAnchors(s)), " ")
}

// AutoPlace puts back each anchor of span that new does not carry, where the anchor's sentence in
// span survives in new word for word and LocateOnce finds it there once — inside the one occurrence
// it counted, so "Costs rose sharply. Costs rose." is never placed on its first "Costs rose" — at
// the place the anchor held in that sentence. An anchor it cannot place stays out, for the transit
// check to refuse by its sentence. The sentence is read within span, the text being replaced: a
// fragment edit that keeps the fragment re-places the anchor, and ReopenedAnchors, reading the
// document's sentence, still records it.
func AutoPlace(span, new string) string {
	flat := flatText(new)
	for _, id := range anchor.IDs(span) {
		tok := anchor.Token(id)
		at := strings.Index(span, tok)
		a, b := sentenceOf(span, at, len(tok))
		if strings.Contains(new, tok) || a < 0 || !strings.Contains(flat, flatText(span[a:b])) {
			continue
		}
		if s, _, err := anchortext.LocateOnce(new, flatText(span[a:b]), anchortext.StopAtParagraph); err == nil {
			if j := inStep(span, a, at, new, s); j >= 0 {
				new = new[:j] + tok + new[j:]
			}
		}
	}
	return new
}

// inStep is the offset in new where the anchor at span[at] goes: span[a:at] and new from s are
// walked in step — anchors skipped on both sides, a whitespace run matching a whitespace run — to
// the text before the anchor, and then past the anchors new holds there that preceded it in span.
// It is -1 where the two texts part.
func inStep(span string, a, at int, new string, s int) int {
	j, run := s, a
	for k := a; k < at; {
		if e := anchor.SkipRun(span, k); e > k {
			k = e
			continue
		}
		j = anchor.SkipRun(new, j)
		switch {
		case anchortext.IsSpace(span[k]) && j < len(new) && anchortext.IsSpace(new[j]):
			for ; k < at && anchortext.IsSpace(span[k]); k++ {
			}
			for ; j < len(new) && anchortext.IsSpace(new[j]); j++ {
			}
		case j < len(new) && new[j] == span[k]:
			j, k = j+1, k+1
		default:
			return -1
		}
		run = k
	}
	for _, id := range anchor.IDs(span[run:at]) {
		if t := anchor.Token(id); strings.HasPrefix(new[j:], t) {
			j += len(t)
		}
	}
	return j
}

// nearestSentence is the sentence of text whose words overlap sent's most, by near-match's measure,
// or "" when none shares one — where a refused anchor's claim most likely now stands.
func nearestSentence(text, sent string) (best string) {
	words, most := anchortext.Tokenize(sent), 0.0
	for _, sp := range anchor.Sentences(text) {
		s := flatText(text[sp[0]:sp[1]])
		if n := anchortext.Jaccard(words, anchortext.Tokenize(s)); n > most {
			best, most = s, n
		}
	}
	return best
}
