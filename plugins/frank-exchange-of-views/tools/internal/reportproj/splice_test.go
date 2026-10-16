package reportproj

import (
	"strings"
	"testing"
)

// The exact splice damage measured on the 2026-08-04 smoke, where blue spent 6 of 17 round-2
// edits repairing punctuation it had created itself. Each case is a seam an edit manufactures.
func TestTidySeamRemovesSpliceArtifacts(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"double period", "It is prime.. The next claim.", "It is prime. The next claim."},
		{"double colon", "Sources:: Cuemath", "Sources: Cuemath"},
		{"double semicolon", "one;; two", "one; two"},
		{"double comma", "a,, b", "a, b"},
		{"space before period", "It is prime . The next", "It is prime. The next"},
	} {
		t.Run(c.name, func(t *testing.T) {
			// the seam sits at the artifact
			at := 0
			for i := 1; i < len(c.in); i++ {
				if c.in[i] == c.in[i-1] && (c.in[i] == '.' || c.in[i] == ':' || c.in[i] == ';' || c.in[i] == ',') {
					at = i
					break
				}
				if c.in[i-1] == ' ' && (c.in[i] == '.' || c.in[i] == ',') {
					at = i
					break
				}
			}
			got, changed := tidySeam(c.in, at)
			if got != c.want {
				t.Errorf("tidySeam = %q, want %q", got, c.want)
			}
			if !changed {
				t.Error("changed = false, want true")
			}
		})
	}
}

// CONTENT IS NEVER TOUCHED. An ellipsis and deliberate emphasis are things a human meant to
// write; a prose normalizer that "fixes" them is worse than the artifacts it removes.
func TestTidySeamLeavesContentAlone(t *testing.T) {
	for _, s := range []string{
		"an ellipsis... is content",
		"emphasis!! stays",
		"a question?? stays",
		"an anchor<!--fx:F-00000abc--> is untouched",
		"no artifact here at all",
	} {
		for at := 1; at < len(s); at++ {
			if got, changed := tidySeam(s, at); changed || got != s {
				t.Fatalf("tidySeam(%q, %d) altered content -> %q", s, at, got)
			}
		}
	}
}

// A RUN OF ANCHORS BETWEEN TWO MARKS IS NOT CONTENT BETWEEN THEM, on either side of the seam. An
// edit that carries its sentence's anchors ends --new in them, so the seam sits AFTER the run and
// the report's terminator abuts it: the doubled mark tidies exactly as the bare pair does.
func TestTidySeamReadsADoubledMarkAcrossAnAnchorRun(t *testing.T) {
	const run = "<!--gap:G-0000000c--><!--fx:F-00000abc-->"
	for _, c := range []struct{ name, in, seam, want string }{
		{"control: the bare pair", "It is prime.. Next.", ". Next.", "It is prime. Next."},
		{"the run before the seam", "It is prime." + run + ". Next.", ". Next.", "It is prime" + run + ". Next."},
		{"the run after the seam", "It is prime." + run + ". Next.", run + ". Next.", "It is prime" + run + ". Next."},
		{"the run on both sides of the seam", "It is prime." + run + ". Next.", "<!--fx:F-00000abc-->. Next.", "It is prime" + run + ". Next."},
		{"two different marks are no pair", "It is prime:" + run + ". Next.", ". Next.", "It is prime:" + run + ". Next."},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got, _ := tidySeam(c.in, strings.LastIndex(c.in, c.seam)); got != c.want {
				t.Errorf("tidySeam = %q, want %q", got, c.want)
			}
		})
	}
}

// THE RUN A SPLICE LEAVES IS MEASURED ON THE VISIBLE TEXT. The m16 shape: the report's terminator
// stands after the sentence's anchors, the replacement ends in its own mark before them, and the
// reader sees "91:." with four anchors inside it.
func TestDoubledTerminatorReadsAcrossAnAnchorRun(t *testing.T) {
	const run = "<!--gap:G-0000000c--><!--fx:F-f8f18cf1--><!--fx:F-63bb2a68--><!--gap:G-00000004-->"
	before := "we searched for cases where it *fails* or is *limited*" + run + ".\n"
	if got := DoubledTerminator(before, "The following findings examine each method's performance on 91:"+run+".\n"); got != ":." {
		t.Errorf("DoubledTerminator = %q, want the run %q the reader sees", got, ":.")
	}
	if got := DoubledTerminator(before, "The following findings examine each method's performance on 91"+run+".\n"); got != "" {
		t.Errorf("DoubledTerminator = %q on an edit that leaves one terminator", got)
	}
}

// ONLY THE SEAMS ARE READ, and a run there doubles only when both sides carry a terminator or the
// same closing quote. Each row is the report a splice leaves, read against the one it replaced.
func TestDoubledTerminatorReadsOnlyTheSeam(t *testing.T) {
	for _, c := range []struct{ name, before, after, want string }{
		// universe-m24 ev246: the replacement's own `",` is prose, in a report with no run that long.
		{"a run inside the replacement", `A web search for "x" returned nine pages. Next.`, `For the query "x", nine pages come up. Next.`, ""},
		// The same, with the replacement sharing the text up to the quote: the seam is not the quote.
		{"a comma after a quote the edit kept", `He said "no" today. Next.`, `He said "no", today. Next.`, ""},
		// universe-m24 ev238: a quoted phrase ending the sentence before the report's period.
		{"a closing quote then the report's period", `Pages came up. Next.`, `Pages came up for "x". Next.`, ""},
		{"a terminator beside the report's", "It works. Next.", "It works!. Next.", "!."},
		{"the same closing quote on both sides", `He said "hello". Next.`, `He said "goodbye"". Next.`, `"".`},
		{"the same single quote on both sides", "He said 'hello'. Next.", "He said 'goodbye''. Next.", "''."},
		// research/2026-09-02_quadratic-formula: a doubled terminator made a tripled one.
		{"a doubled run made longer", `The count is 27."." Next.`, `The count is 28."."." Next.`, `."."."`},
		{"a deletion joining two marks", "A, B. C.", "A,. C.", ",."},
		// An edit that moves two anchors between its sentences: before and after part inside a token.
		{"a terminator beside the report's, past anchors the edit moved",
			"It is prime<!--fx:F-11111111-->. It is odd<!--fx:F-22222222-->. Next.",
			"It is prime<!--fx:F-22222222-->. It is odd!<!--fx:F-11111111-->. Next.", "!."},
		// A run as long elsewhere in the report does not license this one.
		{"a terminator beside the report's, in a report with longer runs", `He asked "why?". It works. Next.`, `He asked "why?". It works!. Next.`, "!."},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := DoubledTerminator(c.before, c.after); got != c.want {
				t.Errorf("DoubledTerminator = %q, want %q", got, c.want)
			}
		})
	}
}
