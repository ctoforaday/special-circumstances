package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE ORDER THE REPORT'S TEXT STANDS IN IS A RULING, written out here so a row that moves in the
// table is seen: the title line, TL;DR, the Catechism, the report's own sections in the order
// their author chooses, Open questions.
func TestTheOrderOfTheReportsTextIsTheRuling(t *testing.T) {
	want := "the title line (`# …`); `## TL;DR`; `## The Catechism`; the report's own sections, in the order their author chooses; `## Open questions`"
	if TextOrder != want {
		t.Errorf("the order reads\n  %s\nwant\n  %s", TextOrder, want)
	}
	if !strings.Contains(TextOrderRule, TextOrder) {
		t.Errorf("the rule on the help pages does not state the order: %s", TextOrderRule)
	}
}

// THE ORDER BLUE IS TOLD IS THE ORDER THE WRITE HOLDS. The write refuses a text out of the table's
// order, so the seat that writes the report and the seat that edits it are each told that order:
// the sections the table places, named in the table's order, from where the prompt states it.
// debate.js is hand-authored and cannot be generated from the table, so its sentence is held to
// it here through the prompts the simulator suite renders and pins.
func TestBluesPromptsStateTheOrderTheWriteHolds(t *testing.T) {
	rows, _ := textRows()
	for seat, from := range map[string]string{"blue-synthesize": "report.md OPENS with the H1 title", "blue-respond": "THEIR ORDER BINDS TOO:"} {
		prompt := strings.ReplaceAll(renderedPrompt(t, seat), "\n", " ")
		at := strings.Index(prompt, from)
		if at < 0 {
			t.Fatalf("the %s prompt does not state the order of the report's text (%q)", seat, from)
		}
		named := 0
		for _, s := range rows {
			if s.Line != "" || s.embeds {
				continue
			}
			i := strings.Index(prompt[at:], "`## "+s.Heading+"`")
			if i < 0 {
				t.Fatalf("the %s prompt does not name `## %s` in the table's order, after what the table places before it", seat, s.Heading)
			}
			at += i
			named++
		}
		if named != 3 {
			t.Errorf("the table places %d section(s) of the report's text by name; this test reads three", named)
		}
		if !strings.Contains(prompt[at:], "refused") {
			t.Errorf("the %s prompt states the order and not that a text out of it is refused", seat)
		}
	}
}

const (
	tl    = "## TL;DR\n\nNo.\n\n"
	cat   = "## The Catechism\n\nThese two factors settle it.\n\n"
	own1  = "## Is 91 divisible by a small prime?\n\nYes: 7 divides it.\n\n"
	own2  = "## What a Fermat test would have shown\n\nNothing decisive.\n\n"
	open  = "## Open questions\n\nNone.\n\n"
	title = "# Is 91 prime?\n\n"
)

// A TEXT IS OUT OF ORDER FOR A SECTION OUT OF PLACE, NEVER FOR ONE IT LACKS. universe-m21's blue
// wrote the Catechism after the sections it answers from; it shipped ahead of them, and no seat
// had read it there.
func TestATextInOrderIsNotRefused(t *testing.T) {
	for name, doc := range map[string]string{
		"all of it":                        title + tl + cat + own1 + own2 + open,
		"no text":                          "",
		"a title and nothing more":         title,
		"no title":                         tl + cat + own1 + open,
		"no heading":                       title + "One paragraph.\n\nAnother.\n",
		"no TL;DR":                         title + cat + own1 + open,
		"no Catechism":                     title + tl + own1 + own2 + open,
		"no Open questions":                title + tl + cat + own1,
		"none of the three":                title + own1 + own2,
		"only the three":                   title + tl + cat + open,
		"text before the first heading":    title + "A framing paragraph.\n\n" + own1 + open,
		"the names in another case":        title + "## tl;dr\n\nNo.\n\n## THE CATECHISM\n\nx.\n\n" + own1 + "##  open   questions (in full)\n\nNone.\n",
		"an anchored heading":              title + tl + "## The Catechism<!--fx:F-00000aaa-->\n\nx.\n\n" + own1 + open,
		"a row's heading inside a fence":   title + tl + own1 + "```\n## TL;DR\n# not a title\n## The Catechism\n```\n\n" + open,
		"a deeper heading of a row's":      title + tl + own1 + "### The Catechism\n\nx.\n\n### TL;DR\n\ny.\n\n" + open,
		"a section named as the body row":  title + tl + "## " + bodyHeading + "\n\nx.\n\n" + open,
		"a title-like line inside a fence": tl + own1 + "```\n# a comment in code\n```\n\n" + open,
		"a second title line":              title + tl + own1 + "# Another line that opens as a title\n\n" + open,
	} {
		if m, found := OutOfOrderIn("", doc); found {
			t.Errorf("%s: refused for %q %v %q", name, m.Line, m.After, m.Beside)
		}
	}
}

func TestATextOutOfOrderNamesTheLineWhereItStandsAndWhereItBelongs(t *testing.T) {
	const (
		first    = "nothing in the report's text stands before it"
		afterT   = "nothing in the report's text but the title line (`# …`) stands before it"
		afterTTL = "nothing in the report's text but the title line (`# …`) and `## TL;DR` stands before it"
		last     = "nothing in the report's text stands after it"
	)
	for name, tc := range map[string]struct {
		doc  string
		want Misplaced
	}{
		"the Catechism after the argument (universe-m21)": {title + tl + own1 + own2 + cat + open,
			Misplaced{"## The Catechism", "## Is 91 divisible by a small prime?", true, afterTTL}},
		"the Catechism before the TL;DR": {title + cat + tl + own1,
			Misplaced{"## TL;DR", "## The Catechism", true, afterT}},
		"the TL;DR after a section of the report's own": {title + own1 + tl,
			Misplaced{"## TL;DR", "## Is 91 divisible by a small prime?", true, afterT}},
		"a section after Open questions": {title + tl + own1 + open + own2,
			Misplaced{"## Open questions", "## What a Fermat test would have shown", false, last}},
		"Open questions before the Catechism (universe-m20)": {title + tl + open + own1 + cat,
			Misplaced{"## Open questions", "## Is 91 divisible by a small prime?", false, last}},
		"a second TL;DR after Open questions": {title + tl + own1 + open + "## TL;DR\n\nAgain.\n",
			Misplaced{"## TL;DR", "## TL;DR", true, afterT}},
		"the Catechism after Open questions": {title + tl + own1 + open + cat,
			Misplaced{"## The Catechism", "## Is 91 divisible by a small prime?", true, afterTTL}},
		"a second TL;DR": {title + tl + "## tl;dr\n\nAgain.\n\n" + own1,
			Misplaced{"## tl;dr", "## TL;DR", true, afterT}},
		"a second Open questions": {title + tl + own1 + open + "## Open Questions (in full)\n\nMore.\n",
			Misplaced{"## Open Questions (in full)", "## Open questions", true, last}},
		"the TL;DR after text before the first heading": {title + "A framing paragraph.\n\nA second.\n\n" + tl + own1,
			Misplaced{"## TL;DR", "A framing paragraph.", true, afterT}},
		"the Catechism after text before the first heading, no TL;DR": {title + "A framing paragraph.\n\n" + cat,
			Misplaced{"## The Catechism", "A framing paragraph.", true, afterTTL}},
		"the title after text": {"A line before the title.\n\n" + title + own1,
			Misplaced{"# Is 91 prime?", "A line before the title.", true, first}},
		"the title after a section": {tl + title + own1,
			Misplaced{"# Is 91 prime?", "## TL;DR", true, first}},
		"a fence before the TL;DR": {title + "```\ncode\n```\n\n" + tl,
			Misplaced{"## TL;DR", "```", true, afterT}},
	} {
		got, found := OutOfOrderIn("", tc.doc)
		if !found || got != tc.want {
			t.Errorf("%s: got %+v (found %v), want %+v", name, got, found, tc.want)
		}
	}
}

// A WRITE IS REFUSED FOR WHAT IT ADDS. A record migrated with its text as written can hold a
// section out of place: an edit that leaves it there or moves it to its place is not refused, and
// one that puts a further section out of place is.
func TestOnlyALineTheWritePutsOutOfOrderIsRefused(t *testing.T) {
	was := title + tl + own1 + cat + open
	for name, now := range map[string]string{
		"an unrelated change":             strings.Replace(was, "7 divides it", "7 divides it, 13 times", 1),
		"the section moved to its place":  title + tl + cat + own1 + open,
		"the section taken out":           title + tl + own1 + open,
		"a section added in its place":    title + tl + own1 + cat + own2 + open,
		"the misplaced section reworded":  strings.Replace(was, "These two factors", "The two factors", 1),
		"another section moved around it": title + tl + own1 + own2 + cat + open,
	} {
		if m, found := OutOfOrderIn(was, now); found {
			t.Errorf("%s: refused for %q beside %q", name, m.Line, m.Beside)
		}
	}
	for name, tc := range map[string]struct{ now, line string }{
		"a section added after Open questions": {was + own2, "## Open questions"},
		"a second Catechism":                   {title + tl + own1 + cat + cat + open, "## The Catechism"},
		"the TL;DR moved out of place as well": {title + own1 + tl + cat + open, "## TL;DR"},
	} {
		if m, found := OutOfOrderIn(was, tc.now); !found || m.Line != tc.line {
			t.Errorf("%s: got %q (found %v), want %q refused", name, m.Line, found, tc.line)
		}
	}
}

// A TEXT THE WRITE TAKES SHIPS IN THE ORDER IT IS WRITTEN: every arrangement of the sections a
// report can hold that is in order comes out of assembly with its lines in the order its author
// gave them. This is what holds the order the write refuses by to the order assembly ships.
func TestATextInOrderShipsInTheOrderWritten(t *testing.T) {
	parts := []string{tl, cat, own1, own2, open, "A framing paragraph.\n\n"}
	var arrange func(done []string, left []string)
	n := 0
	arrange = func(done, left []string) {
		doc := title + strings.Join(done, "")
		if _, out := OutOfOrderIn("", doc); !out {
			n++
			run := runtest.Open(t, newRun(t))
			id := record.Identity{Run: run, SeatID: "blue-synthesize"}
			if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
				t.Fatal(err)
			}
			if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(doc)}); err != nil {
				t.Fatal(err)
			}
			docs, err := AssembleAll(run)
			if err != nil {
				t.Fatal(err)
			}
			at := 0
			for _, ln := range strings.Split(doc, "\n")[1:] {
				if ln = strings.TrimSpace(anchor.StripAssembled(ln)); ln == "" {
					continue
				}
				i := strings.Index(docs[0].Body[at:], ln)
				if i < 0 {
					t.Fatalf("the text\n%s\nis in order, and report.md does not hold %q after the line before it:\n%s", doc, ln, docs[0].Body)
				}
				at += i + len(ln)
			}
		}
		for i := range left {
			rest := append(append([]string(nil), left[:i]...), left[i+1:]...)
			arrange(append(append([]string(nil), done...), left[i]), rest)
		}
	}
	arrange(nil, parts)
	// Of the arrangements of any of the six parts, these are in order: TL;DR, the Catechism and
	// Open questions each present or absent in their places, with the report's own three parts in
	// any order between them, the paragraph first among those when it is there.
	if n < 50 {
		t.Errorf("only %d arrangement(s) were in order and assembled", n)
	}
}
