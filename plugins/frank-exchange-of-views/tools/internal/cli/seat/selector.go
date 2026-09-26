package seat

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// ONE SELECTOR VOCABULARY ACROSS THE THREE KITCHEN SINKS (gblock's ruling).
//
// `report`, `board` and `changes` are the reads a seat should rarely need whole, and they shared no
// working way to ask for part of one. What they shared was an ADVERTISED one: `--id` is registered on
// every view's help and honoured only by `changes`, so a seat reads the flag on `show board --help`,
// tries it, and is refused. Measured on universe-m10: 47 of 74 report reads piped the whole document
// through grep/sed/head, and every board read began `jq '.open[]'`.
//
// # Why two flags and not one with a toggle
//
// A regex is what the work actually wanted: 4 of the 8 patterns seats wrote were alternations
// (`lenses\|audits\|revisions`), because a voice lens hunts a SET of tells. Exact-substring would
// have turned one call into seven.
//
// But regex alone silently finds the wrong thing. A research report is full of parentheses, and a
// seat searching for the literal `(this run)` under regex semantics gets a capture group matching
// `this run`; `O(√n)` matches `O√n` and returns nothing. A metacharacter eaten is a plausible zero,
// which is the class this surface keeps removing.
//
// So the seat says which it means. NOT a `--literal` toggle: whichever way a toggle defaults, half
// the callers get the other semantics without noticing, and the default is then the trap.
//
// # RE2, so the pattern is the seat's and the risk is not
//
// Go's regexp is RE2 — linear time, no backreferences or lookahead. The pattern here is
// AGENT-SUPPLIED, so that guarantee is the reason no timeout or complexity refusal is needed: a
// clumsy or hostile pattern cannot hang the tool.
//
// # Case-insensitive, and never capped
//
// Insensitive by default because a tell-hunt wants `Methodology` and `methodology` both, and a seat
// that has to remember `(?i)` finds half its hits and believes it found all of them.
//
// NOTHING IS EVER TRUNCATED OR RANKED. A selector that returns the first 10 of 40 hands back a
// complete-looking answer over an incomplete read — the same defect as the board's 100-character
// `problem`. Every caller reports the total it matched, so a seat can see its pattern was too broad
// rather than infer it from a short list.

// Selector is a seat's request for part of a projection: at most one of a regex and a literal.
type Selector struct {
	re *regexp.Regexp
	// Literal is the phrase as typed, kept for the message a caller prints when nothing matched —
	// "no line matches /x/" reads differently from "no line contains \"x\"".
	Literal string
}

// Active reports whether the seat asked for a subset at all.
func (s Selector) Active() bool { return s.re != nil }

// Describe is how a caller names the ask when it reports a count or an empty result.
func (s Selector) Describe() string {
	if s.Literal != "" {
		return "the phrase " + `"` + s.Literal + `"`
	}
	if s.re != nil {
		return "/" + s.re.String() + "/"
	}
	return ""
}

// Hits reports whether this text is selected. An inactive selector selects everything, so a caller
// can filter unconditionally.
//
// IT MATCHES THE VISIBLE TEXT, NOT THE RAW BYTES, and that is a correctness fix rather than a
// nicety. The report carries an invisible annotation layer — `negligible<!--fx:f-eee49716--><!--fx:
// f-051df802-->.` is one sentence with two anchors abutting inside it — so a seat quoting
// `overhead is negligible.` exactly as the page reads it matched NOTHING. Verified on m10's real
// report: that phrase returned "no line matches" while the sentence was right there.
//
// A silent zero from an invisible character is the class this whole surface keeps removing, and
// `blue edit` already solved it: anchortext.LocateSpan matches "the report minus its invisible
// annotation layer (markers/footnotes skipped, whitespace runs treated as a single separator)".
// This normalises the same two things for the same reason rather than inventing a third rule.
func (s Selector) Hits(text string) bool {
	if s.re == nil {
		return true
	}
	return s.re.MatchString(visibleText(text))
}

// visibleText is what a seat reads on the page: the annotation layer removed and whitespace runs
// collapsed, matching what anchortext tolerates when it locates a quote.
func visibleText(s string) string {
	return strings.Join(strings.Fields(claimcount.StripAnchors(s)), " ")
}

// AddSelectorFlags registers the pair on a view that supports them.
// THROUGH flags.Text, NOT c.Flags().String. Both values are composed by the seat in a shell, so both
// need the quoting rule that helper appends to the page — a regex is the value most likely to be eaten
// by a shell before the tool sees it. Registering them the plain way skipped that rule and skipped the
// free-text annotation, and two gates said so: TestEveryFreeTextVerbShowsTheQuotingRule and
// TestEveryRegisteredFlagIsInTheDeclaredVocabulary. The machinery for this already existed; building
// the behaviour again per command is what those gates exist to prevent.
func AddSelectorFlags(c *cobra.Command, what string) {
	flags.Text(c, flags.Match,
		"select only the "+what+" matching this `regex` (RE2, case-insensitive) — an alternation is one call where a phrase at a time is several. Every match is returned and the total is stated; nothing is ranked or cut")
	flags.Text(c, flags.Phrase,
		"select only the "+what+" containing this `text` LITERALLY (case-insensitive) — use this rather than --match whenever the text has (), ., *, ? or [] in it, which a regex would read as syntax and silently match something else")
}

// SelectorOf reads the pair, refusing the two ways a seat can get a wrong answer quietly: asking
// with both, and asking with a regex that does not compile.
func SelectorOf(c *cobra.Command) (Selector, error) {
	rx, _ := c.Flags().GetString(flags.Match)
	lit, _ := c.Flags().GetString(flags.Phrase)
	switch {
	case rx != "" && lit != "":
		return Selector{}, feov.Errorf(feov.Validation,
			"--match and --phrase ask the same question two ways and would disagree the moment the text "+
				"has a metacharacter in it. Pass ONE: --match for a regex (an alternation of tells), --phrase for "+
				"text containing (), ., * or [] that a regex would read as syntax")
	case lit != "":
		return Selector{re: regexp.MustCompile("(?i)" + regexp.QuoteMeta(lit)), Literal: lit}, nil
	case rx != "":
		re, err := regexp.Compile("(?i)" + rx)
		if err != nil {
			return Selector{}, feov.Errorf(feov.Validation,
				"--match %q is not a valid regex: %v. If you meant it literally — text with (), ., * or [] in "+
					"it — pass it as --phrase instead, which matches the characters as typed", rx, err)
		}
		return Selector{re: re}, nil
	}
	return Selector{}, nil
}

// selectReportLines renders every line the selector hits, under the heading it sits beneath, with
// its line number — the three things a gap's `location` must carry (the SKILL: "location MUST name
// the section heading and quote the challenged sentence").
//
// THE COUNT IS STATED AND NOTHING IS CUT. A pattern that selects most of the document says so by its
// total, which is the seat's signal to narrow; a capped list would read as a narrow pattern that
// found little. See the header of this file.
func selectReportLines(body string, sel Selector) (string, int) {
	lines := strings.Split(body, "\n")
	heading, hits := "", 0
	var b strings.Builder
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "#") {
			heading = strings.TrimSpace(ln)
		}
		if !sel.Hits(ln) || strings.TrimSpace(ln) == "" {
			continue
		}
		hits++
		if heading != "" {
			fmt.Fprintf(&b, "%s\n", heading)
		} else {
			fmt.Fprintf(&b, "(before the first heading)\n")
		}
		fmt.Fprintf(&b, "  %d: %s\n\n", i+1, strings.TrimRight(ln, " "))
	}
	if hits > 0 {
		fmt.Fprintf(&b, "_%d line(s) match %s, of %d in the report. Every match is above; nothing is ranked or cut._\n",
			hits, sel.Describe(), len(lines))
	}
	return b.String(), hits
}

// selectJSONArrays keeps only the entries of EVERY top-level array whose JSON text the selector hits,
// and reports how many it kept of how many there were.
//
// NO PER-VIEW LIST OF KEY NAMES. Naming the arrays per projection would be a hand-kept table of what
// each view contains — the shape that goes stale the first time a projection grows a list, and the
// defect this surface keeps removing. Every top-level array is a collection of entries, and "which of
// these is about X" is the same question whichever view asked it.
//
// IT MATCHES THE ENTRY'S WHOLE JSON, deliberately: a seat hunting a class, a phrase in a problem, a
// seat id in found_by or a word in an acceptance check is asking one question — "which of these is
// about this?" — and a per-field selector would make it ask once per field, which is the round-trip
// tax this exists to remove.
//
// THE COUNTS GO INSIDE THE DOCUMENT, under `selection`. They were on stderr, on the argument that a
// count folded into the JSON changes the shape a consumer parses — and the seats settle it: measured
// on universe-m10 they habitually write `... 2>&1 | jq '…'`, which would fold the count line into the
// JSON and fail the parse. A read whose own diagnostics break the commonest way of reading it is
// worse than one whose shape gains a key you asked for by passing a flag.
func selectJSONArrays(body []byte, sel Selector) (out []byte, kept, of int) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(body, &doc); err != nil {
		return body, 0, 0 // not an object: hand back what we were given rather than a guess
	}
	for name, raw := range doc {
		var items []json.RawMessage
		if json.Unmarshal(raw, &items) != nil {
			continue // not an array: a scalar or an object is not a set of entries to select from
		}
		hit := []json.RawMessage{}
		for _, it := range items {
			of++
			if sel.Hits(string(it)) {
				hit = append(hit, it)
				kept++
			}
		}
		// NEVER omitempty: an array filtered to nothing renders `[]`, which is "none of these
		// matched", not the absent key a reader would read as "this projection has no such list".
		if b, err := json.Marshal(hit); err == nil {
			doc[name] = b
		}
	}
	// `selection` appears ONLY when a selector was passed, and says what was kept of what there was.
	// Without it a reader cannot tell a narrow pattern from a narrow board.
	if sb, err := json.Marshal(map[string]any{"criterion": sel.Describe(), "matched": kept, "of": of,
		"complete": true}); err == nil {
		doc["selection"] = sb
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return body, kept, of
	}
	return append(b, '\n'), kept, of
}

// WriteSelected is the ONE way a projection's JSON leaves this package.
//
// Ten sites in renderView wrote `cmd.OutOrStdout().Write(b)` directly, and adding selection to each
// would be the same behaviour built ten times — which is the objection that produced this file. Every
// site calls this instead, so a view gains selection by gaining the flags and nothing else, and a new
// view cannot ship without it by forgetting a line.
//
// AN INACTIVE SELECTOR WRITES THE BYTES UNTOUCHED, so a read with no selector is exactly what it was.
func WriteSelected(cmd *cobra.Command, b []byte) error {
	sel, err := SelectorOf(cmd)
	if err != nil {
		return err
	}
	if !sel.Active() {
		_, werr := cmd.OutOrStdout().Write(b)
		return werr
	}
	out, _, _ := selectJSONArrays(b, sel)
	_, werr := cmd.OutOrStdout().Write(out)
	return werr
}

// selectorNoun is what a view's entries are called, for the flag's own help.
//
// A SHORT TABLE WITH AN HONEST DEFAULT, not a name per view. Only the views whose rows have a word a
// seat already uses are named; everything else selects `entries`, which is true and does not rot when
// a view is added.
func selectorNoun(view string) string {
	switch view {
	case "report":
		return "report lines"
	case "board", "work":
		return "gaps"
	case "changes":
		return "edits"
	case "findings":
		return "findings"
	case "evidence":
		return "citations and proofs"
	case "motions":
		return "motions"
	case "lines-of-inquiry":
		return "lines of inquiry"
	}
	return "entries"
}
