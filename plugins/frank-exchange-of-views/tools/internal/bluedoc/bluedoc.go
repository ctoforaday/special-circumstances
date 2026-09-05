// Package bluedoc holds the checks that decide whether a span replacement against
// blue/report.md is LEGAL — shared, because two roles now need the same answer.
//
// WHY IT EXISTS. `blue edit` has always validated its own old→new pair: the span must be
// present, unique, must not split a word, and must not change which immortal anchors exist.
// With #267 stage 3 red may attach a CONCRETE proposed fix to a gap, and a proposal red
// cannot state legally is a proposal blue cannot apply — so the same checks have to run at
// mint time, in the merge role.
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
	"regexp"
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
// merge seat told "blue edit: …" learns the wrong command.
//
// AMBIGUITY IS REFUSED, NOT GUESSED. Taking the first of several matches silently edits a
// site the author may not have meant — and blue is explicitly told to propagate corrections
// to every site stating a claim, so repeated text is the EXPECTED shape of a real report.
func LocateUnique(verb, report, old string) (int, int, error) {
	// AN EDIT MAY CROSS A PARAGRAPH BREAK; an anchor may not. Sharing one rule made this verb's
	// own two refusals jointly unsatisfiable — see anchortext.SpanScope for the measurement.
	start, end, ambiguous := anchortext.LocateSpanUniqueScoped(report, old, anchortext.CrossParagraphs)
	if start < 0 {
		return 0, 0, fmt.Errorf("%s: %w", verb, ErrMisQuote)
	}
	if ambiguous {
		return 0, 0, fmt.Errorf("%s: your quoted span appears MORE THAN ONCE in report.md, so the target is ambiguous — quote more surrounding context to pick out the one site you mean (to change every site, make one edit per site)", verb)
	}
	if !spanBoundaryOK(report, start, end) {
		return 0, 0, fmt.Errorf("%s: your span starts or ends inside a word — quote whole words. Editing letters rather than language produces one-byte ops that carry no meaning on the record", verb)
	}
	return start, end, nil
}

// LocateUniqueReplacing is LocateUnique for a caller that intends to REPLACE the span it finds.
//
// THE ANCHOR RULE BELONGS TO REPLACEMENT, NOT TO LOCATION. Baked into LocateUnique it also fired
// on `merge mint --quote`, which names the sentence a defect LIVES AT and rewrites nothing — so
// minting a gap about any already-anchored sentence was refused, with a message that spoke of
// "the text you are replacing". Caught by reading a regenerated golden that had recorded
// `minted R1-1` turning into `exit 2`, which is what the read-every-diff rule is for.
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
	m := anyAnchorToken.FindStringIndex(after)
	if m == nil || m[0] != 0 {
		return end, nil
	}
	// THE RUN, NOT THE FIRST TOKEN. Two lenses anchoring one sentence is an ordinary corpus shape
	// — `verification<!--fx:f-e4bc25ec--><!--fx:f-73a56bd3-->` is from a real report — and this
	// consumed one token deep. The seat then quoted the sentence exactly as `show report` prints
	// it, carried BOTH markers into --new as instructed, and was told the second one was an
	// INVENTION: it sat past the extended span, so AnchorsTransitUnchanged saw it appear from
	// nowhere. Following its own instruction was the thing that got it refused.
	run := m[1]
	for {
		next := anyAnchorToken.FindStringIndex(after[run:])
		if next == nil || next[0] != 0 {
			break
		}
		run += next[1]
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

	// IT DID NOT: refuse, and print the token to carry.
	return 0, fmt.Errorf("%s: the text you are replacing carries %s, and your quote stops just before it. "+
		"That anchor is ON this sentence: rewriting the sentence without it strands the reference beside prose it was never placed against. "+
		"Quote the sentence AS `show report` PRINTS IT — anchors included — and carry %s into --new unchanged. "+
		"To change the words around it and leave the anchor where it is, quote a FRAGMENT that does not reach it",
		verb, anchor.Label(idIn(tok)), tok)
}

// spanBoundaryOK rejects only a span that SPLITS A WORD.
//
// It is deliberately not a whitespace rule: normalizeQuote trims trailing punctuation, so a
// strict whitespace boundary would reject every sentence-final edit — measured, not feared.
func spanBoundaryOK(s string, start, end int) bool {
	word := func(b byte) bool {
		return b == '_' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}
	splits := func(i int) bool { return i > 0 && i < len(s) && word(s[i-1]) && word(s[i]) }
	return !splits(start) && !splits(end)
}

// AnchorsTransitUnchanged enforces the one anchor invariant a replacement must satisfy: it
// may not change WHICH anchors exist. The multiset of anchor ids in the replaced span must
// equal the multiset in its replacement — so an anchor may be carried across an edit (and
// the prose around it rewritten), but never introduced, dropped or duplicated.
//
// Anchors are still born ONLY from `lens finding` and `blue cite`, and still die only by
// tool. Transit is not authorship: the tool checks the bytes, so nothing is delegated to
// the model.
func AnchorsTransitUnchanged(verb, oldSpan, newText string) error {
	count := func(s string) map[string]int {
		m := map[string]int{}
		for _, id := range claimcount.ProtectedAnchorIDs(s) {
			m[id] = strings.Count(s, anchor.Token(id))
		}
		return m
	}
	o, n := count(oldSpan), count(newText)
	for id, want := range o {
		switch got := n[id]; {
		case got == 0:
			return fmt.Errorf("%s: your old span contains %s but the replacement does not — an anchor may travel through an edit, but never be dropped by one. Reproduce it EXACTLY (%s) somewhere in the replacement", verb, anchor.Label(id), anchor.Token(id))
		case got != want:
			return fmt.Errorf("%s: %s appears %d time(s) in the old span but %d in the replacement — an anchor may not be duplicated or removed by an edit; carry each one across exactly once", verb, anchor.Label(id), want, got)
		}
	}
	for id, got := range n {
		if o[id] == 0 {
			return &ErrAnchorIntroduced{Verb: verb, ID: id, Count: got}
		}
	}
	return nil
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
	return fmt.Sprintf("%s: your replacement introduces %s, which was not in the span it replaces — anchors are placed by `lens finding` and `blue cite`, never typed into a replacement (got %d occurrence(s))", e.Verb, anchor.Label(e.ID), e.Count)
}

// MaxProposalGrowth bounds how much longer a CONCRETE proposed fix may be than the span it
// replaces, in characters.
//
// THIS IS THE LINE BETWEEN AN AUDIT AND AUTHORSHIP, and it is enforced rather than advised.
// Red may propose exact text for a TEXTUAL defect — an overclaim, a wrong figure, a
// contradiction — where compliance genuinely is the right answer and costs blue nothing.
// The moment red may hand over concrete text for a SUBSTANTIVE addition, blue becomes a
// typist by incentive: applying is instant and free, while a counter-edit costs a round and
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
	if err := AnchorsTransitUnchanged(verb, report[start:end], new); err != nil {
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
	var out []string
	for _, id := range claimcount.ProtectedAnchorIDs(before) {
		b, okB := sentenceAround(before, anchor.Token(id))
		a, okA := sentenceAround(after, anchor.Token(id))
		// An anchor that is GONE from `after` is not reopened — it is dropped, which is a refusal
		// the caller has already made. Saying both would report one fault as two.
		if !okB || !okA {
			continue
		}
		if b != a {
			out = append(out, id)
		}
	}
	return out
}

// sentenceAround returns the text between sentence boundaries surrounding tok, with every anchor
// token stripped, so a SECOND anchor arriving in the same sentence does not read as the first
// one's referent changing.
func sentenceAround(doc, tok string) (string, bool) {
	if !strings.Contains(doc, tok) {
		return "", false
	}
	// MARKERS COME OUT BEFORE THE BOUNDARY SCAN, not after.
	//
	// `!` is a sentence terminator and `<!--` contains one, so scanning the raw document ends the
	// segment INSIDE a neighbouring marker — measured: a twin anchor truncated the sentence to
	// "…the grass is green<", which differs from the original and reported every cite as reopened
	// by its own neighbour. Stripping first removes the question.
	const mark = "\x00" // boundary-free, so the scan cannot end inside it
	cleaned := anyAnchorToken.ReplaceAllStringFunc(doc, func(m string) string {
		if m == tok {
			return mark
		}
		return ""
	})
	i := strings.Index(cleaned, mark)
	if i < 0 {
		return "", false
	}
	start := strings.LastIndexAny(cleaned[:i], ".!?\n")
	if start < 0 {
		start = 0
	} else {
		start++
	}
	rest := cleaned[i+len(mark):]
	end := strings.IndexAny(rest, ".!?\n")
	if end < 0 {
		end = len(rest)
	}
	return strings.Join(strings.Fields(cleaned[start:i]+rest[:end]), " "), true
}

// idIn returns the id inside an anchor token — `<!--cite:c-abc-->` yields `c-abc`.
func idIn(tok string) string {
	if i := strings.IndexByte(tok, ':'); i >= 0 {
		return strings.TrimSuffix(tok[i+1:], "-->")
	}
	return tok
}

// anyAnchorToken matches an invisible marker of EITHER class — a finding's `<!--fx:…-->` or a
// citation's `<!--cite:…-->`.
//
// By SHAPE, not by reconstructing tokens from ids. The first draft stripped
// `anchor.Token(id)` for each id ProtectedAnchorIDs found, which silently failed for whichever
// class Token does not render identically — and a token left behind reads as the sentence having
// changed, so every cite reopened its own neighbours. Caught by the twin case in the test rather
// than in review.
var anyAnchorToken = regexp.MustCompile(`<!--[a-z]+:[^>]*-->`)

// stripAnchors removes every anchor token from a segment, so a marker ARRIVING beside a
// reference is not read as that reference's text changing.
func stripAnchors(s string) string { return anyAnchorToken.ReplaceAllString(s, "") }
