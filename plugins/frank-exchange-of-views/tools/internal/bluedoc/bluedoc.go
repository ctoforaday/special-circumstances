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
	"strings"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
)

// ErrMisQuote is the sentinel for "the old span is not present". `blue edit` distinguishes
// it to drive its crash-reconcile branch (old gone but new already present ⇒ the write
// landed), so it must stay identifiable rather than collapse into a generic error.
var ErrMisQuote = errors.New("the quoted span was not found in report.md — quote the EXACT current text you are replacing (whitespace and the invisible anchor layer are ignored; everything else must match)")

// LocateUnique resolves the one span `old` names, or explains why it cannot.
//
// verb prefixes every message, because the seat's only teacher is the error text and a
// merge seat told "blue edit: …" learns the wrong command.
//
// AMBIGUITY IS REFUSED, NOT GUESSED. Taking the first of several matches silently edits a
// site the author may not have meant — and blue is explicitly told to propagate corrections
// to every site stating a claim, so repeated text is the EXPECTED shape of a real report.
func LocateUnique(verb, report, old string) (int, int, error) {
	start, end, ambiguous := lens.LocateSpanUnique(report, old)
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
			m[id] = strings.Count(s, AnchorToken(id))
		}
		return m
	}
	o, n := count(oldSpan), count(newText)
	for id, want := range o {
		switch got := n[id]; {
		case got == 0:
			return fmt.Errorf("%s: your old span contains %s but the replacement does not — an anchor may travel through an edit, but never be dropped by one. Reproduce it EXACTLY (%s) somewhere in the replacement", verb, AnchorLabel(id), AnchorToken(id))
		case got != want:
			return fmt.Errorf("%s: %s appears %d time(s) in the old span but %d in the replacement — an anchor may not be duplicated or removed by an edit; carry each one across exactly once", verb, AnchorLabel(id), want, got)
		}
	}
	for id, got := range n {
		if o[id] == 0 {
			return fmt.Errorf("%s: your replacement introduces %s, which was not in the span it replaces — anchors are placed by `lens finding` and `blue cite`, never typed into a replacement (got %d occurrence(s))", verb, AnchorLabel(id), got)
		}
	}
	return nil
}

// AnchorToken rebuilds the literal token for an anchor id, so a message can quote what must
// be reproduced verbatim.
func AnchorToken(id string) string {
	if strings.HasPrefix(id, "c-") {
		return "<!--cite:" + id + "-->"
	}
	return "<!--fx:" + id + "-->"
}

// AnchorLabel describes an anchor id by its class, so a seat is told which KIND of anchor
// its edit would have disturbed. A generic name is passed through unchanged.
func AnchorLabel(id string) string {
	switch {
	case strings.HasPrefix(id, "c-"):
		return "citation anchor " + id + " (citations are tool-managed — remove one with the tool, never a raw edit)"
	case strings.HasPrefix(id, "f-"):
		return "finding-marker " + id
	default:
		return id
	}
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
	start, end, err := LocateUnique(verb, report, old)
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
