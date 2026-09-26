package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

// ONE SELECTOR VOCABULARY ON THE THREE KITCHEN SINKS (#1122 follow-up, gblock's ruling).
//
// `report`, `board` and `changes` are the reads a seat should rarely need whole and shared no working
// way to ask for part of one — while all advertising `--id`, which only `changes` honours. Measured on
// universe-m10: 47 of 74 report reads piped the whole document through grep/sed/head, and every board
// read began `jq '.open[]'`.

// THE DIVERGENCE THAT DECIDED TWO FLAGS RATHER THAN ONE.
//
// A research report is full of parentheses. `O(√n)` as a REGEX means "O followed by √n" and does not
// match the text `O(√n)`; as a LITERAL it does. Verified against m10's real report before this test
// was written: --match found nothing, --quote found line 48. One flag with a `--literal` toggle
// would hand half the callers the other semantics silently, and a silent zero is the class this
// surface keeps removing.
func TestMatchIsARegexAndPhraseIsLiteral(t *testing.T) {
	runDir := seatRunReport(t, "# Findings\n\nTrial division is O(√n) for 91.\n")
	regexHit, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--match", `O(√n)`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(regexHit, "no line of the report matches") {
		t.Errorf("--match treated the parens as literal; a capture group must not match `O(√n)`:\n%s", regexHit)
	}
	literalHit, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--quote", `O(√n)`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(literalHit, "no line of the report matches") {
		t.Errorf("--quote did not match the characters as typed:\n%s", literalHit)
	}
	// AND THE EMPTY CASE SAYS WHAT IT IS. "found nothing" and "the report is empty" must not be the
	// same bytes — the seat that cannot tell them apart stops auditing.
	if !strings.Contains(regexHit, "not an empty report") {
		t.Errorf("an empty selection did not distinguish itself from an empty report:\n%s", regexHit)
	}
}

// A SELECTED READ CARRIES WHAT IT TAKES TO CITE THE HIT: the section heading, the line number and the
// text. That triple is exactly what a gap's `location` must contain, and assembling it by hand is
// what the grep-then-sed dance was doing — 20 greps and 5 seds across m10.
func TestASelectedReportLineCarriesItsHeadingAndLineNumber(t *testing.T) {
	runDir := seatRunReport(t, "# Findings\n\n## 7. Cost\n\nThe methodology would have a flaw.\n")
	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--match", "methodology")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## 7. Cost", "methodology would", "line(s) match"} {
		if !strings.Contains(out, want) {
			t.Errorf("a selected line is missing %q — a seat cannot write a location without it:\n%s", want, out)
		}
	}
	// CASE-INSENSITIVE, BOTH WAYS. A tell-hunt wants `Methodology` and `methodology` both, and a seat
	// that has to remember `(?i)` finds half its hits and believes it found all of them — which is a
	// silent partial read, not a missed feature. Unasserted, this was the one claim a mutation of
	// `(?i)` walked straight through.
	for _, flag := range []string{"--match", "--quote"} {
		up, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, flag, "METHODOLOGY")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(up, "no line of the report matches") {
			t.Errorf("%s METHODOLOGY missed the lowercase text — a case-sensitive search returns half a hunt and looks complete:\n%s", flag, up)
		}
	}
	// NOTHING IS CUT OR RANKED, and the total says so, so a pattern that selects most of the document
	// is visible as that rather than as a narrow pattern with few hits.
	if !strings.Contains(out, "nothing is ranked or cut") {
		t.Errorf("the selection does not state that it is complete:\n%s", out)
	}
}

// ASKING BOTH WAYS IS REFUSED, because the two disagree the moment the text has a metacharacter and
// a tool that picked one would be picking which answer the seat gets.
func TestTheTwoSelectorsMayNotBeCombined(t *testing.T) {
	runDir := seatRun(t)
	for _, v := range []string{"report", "board", "changes"} {
		args := []string{"show", v, "--run", runDir, "--seat-id", "red-chair", "--match", "x", "--quote", "y"}
		if v != "report" {
			args = append(args, "--json")
		}
		if _, err := run(t, args...); err == nil {
			t.Errorf("show %s accepted --match and --quote together", v)
		} else if !strings.Contains(err.Error(), "Pass ONE") {
			t.Errorf("show %s: the refusal does not say to pass one: %v", v, err)
		}
	}
}

// A MALFORMED REGEX IS REFUSED AND POINTED AT THE OTHER FLAG, because the commonest cause is text
// with a bracket in it that the seat meant literally.
func TestAMalformedRegexNamesTheLiteralFlag(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--match", "O(√n")
	if err == nil {
		t.Fatal("an unbalanced regex was accepted")
	}
	if !strings.Contains(err.Error(), "--quote") {
		t.Errorf("the refusal does not point at the literal flag, which is the usual cause: %v", err)
	}
}

// THE CHANGE LOG CARRIES THE CHANGE. `show changes` gave a char count and blue's own prose account of
// its edit — so a reader auditing a repair was taking the edited party's word for it, which red's
// contract forbids. The old/new text was in hand at the render and reduced to a delta.
func TestTheChangeLogCarriesTheDiffAndIsSelectable(t *testing.T) {
	// answeredBoard seeds the cast `dispatch next` refuses without, plus a report with prose blue can
	// edit — the same fixture the blue-answered items use.
	runDir, id := answeredBoard(t)
	dispatchBlueOnto(t, runDir, id)
	editAnswering(t, runDir, id)

	out, err := run(t, "show", "changes", "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatalf("show changes --json: %v", err)
	}
	var page struct {
		Edits []struct {
			Old, New, Answers string
			Delta             int
		} `json:"edits"`
		Counts struct{ Edits int } `json:"counts"`
	}
	if e := json.Unmarshal([]byte(out), &page); e != nil {
		t.Fatalf("changes --json is not valid JSON (%v):\n%s", e, out)
	}
	if page.Counts.Edits == 0 || len(page.Edits) == 0 {
		t.Fatalf("the change log records no edit after one was made:\n%s", out)
	}
	e := page.Edits[0]
	if e.Old == "" || e.New == "" {
		t.Errorf("the change log carries no diff — a reader must take blue's word for the edit: %+v", e)
	}
	if e.Old == e.New {
		t.Errorf("old and new are identical, so the diff says nothing: %+v", e)
	}
	// AND IT IS SELECTABLE on the text of the change itself, which is the question a re-auditing
	// lens has: did anything touch this sentence?
	sel, err := run(t, "show", "changes", "--run", runDir, "--seat-id", lensSeat, "--json", "--quote", e.Old[:12])
	if err != nil {
		t.Fatalf("selecting the change log: %v", err)
	}
	var kept struct {
		Edits []json.RawMessage `json:"edits"`
	}
	if e := json.Unmarshal([]byte(sel), &kept); e != nil {
		t.Fatalf("a selected change log is not valid JSON (%v):\n%s", e, sel)
	}
	if len(kept.Edits) != 1 {
		t.Errorf("selecting on the removed text kept %d edit(s), want 1:\n%s", len(kept.Edits), sel)
	}
}

// THE BOARD IS SELECTABLE TOO, and an empty selection renders `[]` rather than dropping the key: a
// missing array reads as "this projection has no such list", which is a different answer from "none
// of them matched".
func TestTheBoardIsSelectableAndKeepsItsEmptyArrays(t *testing.T) {
	runDir := seatRun(t)
	mintGap(t, runDir, "board-sel", "c")
	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair", "--match", "no-such-class-anywhere")
	if err != nil {
		t.Fatal(err)
	}
	var b struct {
		Open   *[]json.RawMessage `json:"open"`
		Closed *[]json.RawMessage `json:"closed"`
	}
	if e := json.Unmarshal([]byte(out), &b); e != nil {
		t.Fatalf("a selected board is not valid JSON (%v):\n%s", e, out)
	}
	if b.Open == nil || b.Closed == nil {
		t.Fatalf("a selection that matched nothing dropped an array key instead of emptying it:\n%s", out)
	}
	if len(*b.Open) != 0 {
		t.Errorf("a pattern matching nothing kept %d open gap(s)", len(*b.Open))
	}
}

// THE SELECTOR MATCHES THE VISIBLE TEXT, NOT THE RAW BYTES.
//
// The report carries an invisible annotation layer, and anchors ABUT inside a sentence:
// `negligible<!--fx:f-eee49716--><!--fx:f-051df802-->.` is one sentence. A seat quoting
// `negligible.` exactly as the page reads it matched NOTHING — verified on m10's real report before
// this gate was written, and it is the silent zero this surface keeps removing. `blue edit` already
// tolerated both anchors and whitespace runs; the selector now normalises the same two things.
func TestASelectorMatchesThroughAnchorsAndWhitespace(t *testing.T) {
	runDir := seatRunReport(t, "# Findings\n\nOverhead is negligible<!--fx:f-eee49716--><!--fx:f-051df802-->.\n\nCost   is\n  spread over lines.\n")
	for _, tc := range []struct{ name, flag, value string }{
		{"a phrase spanning two abutting anchors", "--quote", "negligible."},
		{"a regex spanning an anchor", "--match", "negligible[.]"},
		{"a phrase across collapsed whitespace", "--quote", "Cost is"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, tc.flag, tc.value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out, "no line of the report matches") {
				t.Errorf("%s %q found nothing — an invisible character must not hide a sentence the seat can see:\n%s", tc.flag, tc.value, out)
			}
		})
	}
}
