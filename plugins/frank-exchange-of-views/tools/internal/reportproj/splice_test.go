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
