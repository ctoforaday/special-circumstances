package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE CITATIONS AS THEY WERE CHECKED. The report's notes print a source's title, URL and access
// date, and red's reading of each source against the sentence it backs reached no reader: across
// three archived runs, 82 readings and not one of their texts in any document (the reach audit,
// 2026-10-20). A reader could not tell a citation red found `refutes` from one nobody read. This
// section is the evidence view's sources (record.EvidenceJSONOf) as a document: each source with
// every reading of it, its outcome, its confidence and what the reader read, and red's readings of
// sources the report does not cite after them.

// citationsHeading is the section of evidence.md this file composes.
const citationsHeading = "Citations"

// citationsSection renders evidence.md's citations from the evidence view ev, each with the sentence
// it backs in the report (live, keyed by anchor), or "" for a run that cited nothing and read nothing.
func citationsSection(ev record.EvidenceJSON, live map[string]string) string {
	others := ev.Independent
	if len(ev.Sources) == 0 && len(others) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## " + citationsHeading + "\n\n")
	b.WriteString("Each source the report cites, with the sentence it backs and every reading of the source against that sentence. ")
	b.WriteString("A reading's outcome is what the source does for the sentence — `supports`, `supports_with_bridge` (it supports the sentence through a step the source does not state), `weak`, `refutes`, `absent` (the source says nothing on it) or `unreachable` — and its confidence is how sure the reading is.\n\n")
	for _, s := range ev.Sources {
		// THE SENTENCE IT BACKS IS THE REPORT'S NOW, as a computation's is: the cite's location is the
		// sentence as blue cited it, and an edit since changes the one a reading is measured against.
		// A citation whose token the report does not place has no sentence there, and shows its location.
		backs := live[s.Anchor]
		if backs == "" {
			backs = s.Location
		}
		fmt.Fprintf(&b, "### %s\n\n", firstNonEmpty(plain(s.Title), plainURL(s.URL)))
		fmt.Fprintf(&b, "- **source**: %s", plainURL(s.URL))
		if on := record.AccessedOn(s.FetchedAt); on != "" {
			fmt.Fprintf(&b, ", accessed %s", on)
		}
		if on := record.AccessedOn(s.CopyCapturedAt); on != "" {
			fmt.Fprintf(&b, ", from the web archive's copy captured %s", on)
		}
		b.WriteString("\n")
		if loc := plain(backs); loc != "" {
			fmt.Fprintf(&b, "- **backs**: %s\n", loc)
		}
		if len(s.Verified) == 0 {
			b.WriteString("- **not read against its sentence** — no reading of this source is on the record\n")
		}
		b.WriteString(checkLines(readingChecks(s.Verified), backs))
		b.WriteString("\n")
	}
	if len(others) > 0 {
		fmt.Fprintf(&b, "### Sources the report does not cite (%d)\n\n", len(others))
		b.WriteString("Sources red found and read against a sentence of the report that place no citation in it: a reading that does not support its sentence places none.\n\n")
		for _, v := range others {
			title := plain(v.Title)
			fmt.Fprintf(&b, "- **%s**", firstNonEmpty(title, plainURL(v.URL), "(no source named)"))
			if v.URL != "" && title != "" {
				fmt.Fprintf(&b, " — %s", plainURL(v.URL))
			}
			b.WriteString("\n")
			b.WriteString(indent(checkLines(independentChecks([]record.EvidenceVerificationJSON{v}), ""), "  "))
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// evidenceCheck is one check of an evidence anchor's sentence as evidence.md lists it: a reading of a
// cited source (`verify`) or a judgement of a computation (`reproduce`), from the evidence view's row.
type evidenceCheck struct {
	// by is the act, "read by" or "judged by"; verdict the reading's outcome or the judgement's
	// soundness.
	by, seat   string
	epoch      int
	verdict    string
	confidence string
	page       int32
	// text is what the reading read, or the judgement's note; claim the sentence the check was of,
	// "" where the record holds none (a migrated reproduce).
	text, claim string
	// falls is whether the check is of an anchor and does not stand on it: its seat checked the
	// sentence again, an edit changed the sentence since, or it read nothing. The view's `standing`
	// says which checks stand, by the anchor ledger's one rule.
	falls bool
}

// readingChecks is each reading of a cited source as a check, each marked where it does not stand.
func readingChecks(vs []record.EvidenceVerificationJSON) []evidenceCheck {
	out := independentChecks(vs)
	for i, v := range vs {
		out[i].falls = v.Standing == nil || !*v.Standing
	}
	return out
}

// independentChecks is each reading of a source the report does not cite as a check: it is of no
// anchor, so there is nothing for it to stand on and nothing to mark.
func independentChecks(vs []record.EvidenceVerificationJSON) []evidenceCheck {
	out := make([]evidenceCheck, len(vs))
	for i, v := range vs {
		out[i] = evidenceCheck{by: "read by", seat: v.SeatID, epoch: v.Epoch, verdict: v.Outcome, confidence: v.Confidence, page: v.Page, text: v.Text, claim: v.Claim}
	}
	return out
}

// judgementChecks is each judgement of a computation's sentence as a check.
func judgementChecks(js []record.EvidenceReproductionJSON) []evidenceCheck {
	out := make([]evidenceCheck, len(js))
	for i, j := range js {
		out[i] = evidenceCheck{by: "judged by", seat: j.SeatID, epoch: j.Epoch, verdict: j.Soundness, text: j.Note, claim: j.Claim, falls: !j.Standing}
	}
	return out
}

// checkLines renders every check of one anchor, a citation's or a computation's, as a list item: who
// made it and in which epoch, the outcome or the soundness and the confidence, the page when it read
// one, and what it read or noted. A text repeated from the check above it is said to be, rather than
// printed twice.
//
// A CHECK THAT DOES NOT STAND SAYS SO on its line: every check of the anchor is listed, a seat's
// earlier check under its later one and a check of the sentence before an edit under the headline
// that reports the anchor unchecked, and without the mark an earlier `unsound` reads as current
// beneath a headline that accepts the sentence.
//
// UNDER EACH, THE SENTENCE IT WAS OF, when that is not live — the sentence the anchor backs in the
// report now. A check of an earlier wording is how an edit of the sentence surfaces here; whether the
// check still stands is the anchor ledger's call. The two are compared in the form a check records
// its sentence (bluedoc.SameSentence), so a token placed in the sentence after the check — a reading
// gap's, at its end — is no change of wording. A check that recorded no sentence prints no line.
func checkLines(cs []evidenceCheck, live string) string {
	var b strings.Builder
	prevText, prevClaim := "", ""
	for _, c := range cs {
		fmt.Fprintf(&b, "- **%s `%s`** (epoch %d): **%s**", c.by, c.seat, c.epoch, firstNonEmpty(c.verdict, "no outcome"))
		if c.falls {
			b.WriteString(" (does not stand)")
		}
		if c.confidence != "" {
			fmt.Fprintf(&b, ", confidence %s", c.confidence)
		}
		if c.page > 0 {
			fmt.Fprintf(&b, ", page %d", c.page)
		}
		switch text := plain(c.text); {
		case text == "":
		case text == prevText:
			b.WriteString(" — the same text as the check above")
		default:
			b.WriteString(" — " + checkText(c.text))
		}
		prevText = plain(c.text)
		b.WriteString("\n")
		switch {
		case plain(c.claim) == "" || bluedoc.SameSentence(c.claim, live):
		case prevClaim != "" && bluedoc.SameSentence(c.claim, prevClaim):
			b.WriteString("  - of the same sentence as the check above\n")
		default:
			fmt.Fprintf(&b, "  - of the sentence: %s\n", plain(c.claim))
		}
		prevClaim = c.claim
	}
	return b.String()
}

// checkText is a check's text as it ships after the tool's on the check's line, without its tokens as
// it reads there (stripTokensAfter). Of one line, it stays on that line; of several, it runs on as the
// item's own paragraphs, a fence it leaves open closed where it ends (ownBlock, reading its first line
// as text after the tool's), so a reader's quoted block keeps its shape (itemLines).
func checkText(s string) string {
	t := strings.TrimSpace(lineEnds.Replace(stripTokensAfter(s)))
	if !strings.Contains(t, "\n") {
		return oneLine(t)
	}
	return itemLines(ownBlock("x" + t)[1:], "  ")
}

// plain is a record field as the tool writes it into a line it composes after the weave — a field of
// evidence.md, a note's text, a Bibliography entry — on one line, with no anchor token in it
// (stripTokensInLine): no weave follows to take one, and no strip.
//
// THE FIELD IS MADE ONE LINE, THEN STRIPPED: the strip keeps a token in code, and what is code is read
// from the line that ships. A token on an indented line is code in the field and prose on the one
// line, where it would ship as typed. The line is made one line again after: a token's removal leaves
// the spaces either side of it, and outside code a run of them is one space.
func plain(s string) string { return oneLine(stripTokensInLine(oneLine(s))) }

// indent prefixes every non-empty line of s.
func indent(s, by string) string {
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln != "" {
			lines[i] = by + ln
		}
	}
	return strings.Join(lines, "\n")
}
