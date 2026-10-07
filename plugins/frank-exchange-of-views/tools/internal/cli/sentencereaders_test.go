package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY READER OF AN ANCHOR'S SENTENCE READS ONE WHOLE SENTENCE, through the real verbs. Each row
// fails on a reader that splits at an internal period ("3.5%", "13. ✓") or at a soft wrap: the
// edit's reopened set, the claim count, retire's bare set, the retire tidy, the risk matrix's lead
// sentence, the sentence an edit's left-out anchor is put back on, and a gap's location, backing and
// edited_since.
func TestEverySentenceReaderReadsOneSentence(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\n"+
		"Prices rose 3.5% in 2024. Then fell.\n\n"+
		"Costs rose 4.5%\nin 2023. Then fell back.\n\n"+
		"91 = 7 × 13. ✓ Verified twice.\n\n"+
		"**One.** Gone soon. Two.\n\n"+
		"1. Gone item.\n2. Kept item.\n")
	registerBlue(t, runDir)
	resp := map[string][]byte{}
	for _, u := range []string{"https://a/1", "https://a/2", "https://a/3", "https://a/4", "https://a/5"} {
		resp[u] = []byte("<html>a source at " + u + "</html>")
	}
	withFetcher(t, &fakeFetcher{resp: resp})
	cite := func(sentence, url string) string {
		t.Helper()
		if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat,
			"--quote", sentence, "--url", url, "--title", "Source "+url); err != nil {
			t.Fatalf("cite %q: %v", sentence, err)
		}
		return lastBody(t, runDir, &recordpb.Cite{}).GetLabel()
	}
	edit := func(quote, replacement string) *recordpb.BlueEdit {
		t.Helper()
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
			"--quote", quote, "--new", replacement, "--reason", "r"); err != nil {
			t.Fatalf("edit %q: %v", quote, err)
		}
		return lastBody(t, runDir, &recordpb.BlueEdit{})
	}
	retire := func(claim string) *recordpb.Retire {
		t.Helper()
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", claim, "--reason", "refuted"); err != nil {
			t.Fatalf("retire %q: %v", claim, err)
		}
		return lastBody(t, runDir, &recordpb.Retire{})
	}
	c1 := cite("Prices rose 3.5% in 2024", "https://a/1")
	c1w := cite("Costs rose 4.5% in 2023", "https://a/2")
	c2 := cite("91 = 7 × 13. ✓", "https://a/3")
	c3 := cite("Gone soon", "https://a/4")
	c4 := cite("Gone item", "https://a/5")

	t.Run("a sentence carrying a terminator inside it counts as one claim", func(t *testing.T) {
		if out, err := run(t, "count-claims", "--run", runDir, "--seat-id", blueSeat); err != nil || strings.TrimSpace(out) != "5" {
			t.Errorf("count-claims = %q (%v), want 5: \"91 = 7 × 13. ✓%s\" is a cited claim", out, err, anchor.Token(c2))
		}
	})
	t.Run("an anchor left out of a rewritten claim is refused by its whole sentence", func(t *testing.T) {
		// "5% in 2024" survives in the replacement, so a reader that split at "3." would put the
		// citation back on the rewritten claim.
		_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat,
			"--quote", "Prices rose 3.5% in 2024"+anchor.Token(c1), "--new", "Prices rose 1.5% in 2024", "--reason", "r")
		if err == nil || !strings.Contains(err.Error(), `"Prices rose 3.5% in 2024"`) {
			t.Errorf("edit = %v, want a refusal naming \"Prices rose 3.5%% in 2024\"", err)
		}
	})
	t.Run("an edit before an internal period reopens the citation", func(t *testing.T) {
		if got := edit("rose 3.5%", "rose 1.5%").GetReopened(); !slices.Contains(got, c1) {
			t.Errorf("reopened = %v, want %s: the edit changed the words its citation backs", got, c1)
		}
	})
	t.Run("an edit across a soft wrap reopens the citation", func(t *testing.T) {
		if got := edit("rose 4.5%", "rose 2.5%").GetReopened(); !slices.Contains(got, c1w) {
			t.Errorf("reopened = %v, want %s: the edit changed the words its citation backs, a line above it", got, c1w)
		}
	})
	t.Run("retire leaves an anchor attached after a terminator and a symbol", func(t *testing.T) {
		edit("✓"+anchor.Token(c2)+" Verified twice", "✓"+anchor.Token(c2))
		if got := retire("✓ Verified twice").GetAnchors(); slices.Contains(got, c2) {
			t.Errorf("retire took %v: %s backs \"91 = 7 × 13. ✓\", a live claim", got, c2)
		}
		if md := readReport(t, runDir); !strings.Contains(md, "91 = 7 × 13. ✓"+anchor.Token(c2)) {
			t.Errorf("the live claim lost its anchor:\n%s", md)
		}
	})
	t.Run("the retire tidy keeps the closer before the emptied sentence", func(t *testing.T) {
		tok := anchor.Token(c3)
		edit("Gone soon"+tok+".", tok) // the cut; the splice tidy takes its period
		edit("**One.** "+tok+" Two.", "**One.** "+tok+". Two.")
		if got := retire("Gone soon.").GetAnchors(); !slices.Equal(got, []string{c3}) {
			t.Errorf("retire took %v, want [%s]", got, c3)
		}
		if md := readReport(t, runDir); !strings.Contains(md, "\n**One.** Two.\n") {
			t.Errorf("want \"**One.** Two.\" after the retire:\n%s", md)
		}
	})
	t.Run("the retire tidy keeps an emptied ordered item's marker", func(t *testing.T) {
		edit("Gone item"+anchor.Token(c4)+".", anchor.Token(c4))
		retire("Gone item.")
		if md := readReport(t, runDir); !strings.Contains(md, "\n1.\n2. Kept item.") {
			t.Errorf("want the emptied item to read \"1.\" after the retire:\n%s", md)
		}
	})
	t.Run("the risk matrix cell is the whole first sentence", func(t *testing.T) {
		registerLensOnce(t, runDir)
		if _, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", "G1", "--class", "scope-creep", "--about-kind", "section", "--about", "Findings",
			"--problem", "The committee chair said “Stop.” Then everyone left the room.",
			"--check-kind", "document", "--check", "the report says who left",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
			t.Fatal(err)
		}
		if asm := assembled(t, runDir); !strings.Contains(asm, "| The committee chair said “Stop.” |") {
			t.Errorf("the risk matrix cell is not the first sentence, closing quote included:\n%s", asm)
		}
	})
	t.Run("a gap's location, backing and edited_since read its whole sentence", func(t *testing.T) {
		gap, err := mintQuote(t, runDir, "G2", "Costs rose 2.5% in 2023")
		if err != nil {
			t.Fatal(err)
		}
		g := openGap(t, runDir, gap)
		if g.LocationState != "marked" || g.Location != "Costs rose 2.5%\nin 2023." {
			t.Errorf("the gap reads %s %q, want the whole soft-wrapped sentence", g.LocationState, g.Location)
		}
		if len(g.Backing) != 1 || g.Backing[0].Anchor != c1w {
			t.Errorf("backing = %+v, want the citation in the gap's sentence, %s", g.Backing, c1w)
		}
		if got := edit("rose 2.5%", "rose 0.5%").GetReopened(); !slices.Contains(got, gap) {
			t.Errorf("reopened = %v, want %s: the edit changed the words of the gap's sentence, a line above its anchor", got, gap)
		}
	})
}
