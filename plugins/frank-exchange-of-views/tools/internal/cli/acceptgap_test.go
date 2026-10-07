package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

const (
	acceptS    = "Costs rose sharply across the first quarter"
	acceptN    = "Costs rose only modestly across the first quarter of the year"
	acceptKeep = acceptS + ". Analysts disagree about why"
)

// acceptRun is a report holding acceptS as one sentence of a paragraph, a lens seated to mint and
// blue seated to edit and cite.
func acceptRun(t *testing.T) string {
	t.Helper()
	return acceptRunOf(t, "# Costs\n\nThe year opened badly. "+acceptS+". Volume recovered later.\n")
}

func acceptRunOf(t *testing.T, body string) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, body)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://costs/1": []byte("<html>costs rose</html>")}})
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	return runDir
}

func acceptEdit(t *testing.T, runDir, key, gap string, args ...string) error {
	t.Helper()
	_, err := run(t, append([]string{"edit", "--run", runDir, "--seat-id", blueSeat, "--key", key,
		"--answers", gap, "--reason", "red's text is right"}, args...)...)
	return err
}

// --ACCEPT CARRIES THE GAP'S ANCHOR (R-13, R-17, R-24). The pair the board serves, the pair
// `--accept` sends and the pair a typed application is compared with are one function's output over
// one render: old is the location through the anchor run holding the gap's anchor, new is red's text
// with each of that run's gap anchors at its end. Every row runs with the location stored without
// and with its terminator.
func TestAcceptCarriesTheGapAnchor(t *testing.T) {
	for _, term := range []string{"", "."} {
		loc := acceptS + term
		t.Run("plain accept"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			gap, err := mintQuote(t, runDir, "G", loc, "--new", acceptN+term)
			if err != nil {
				t.Fatal(err)
			}
			if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
				t.Fatalf("accept: %v", err)
			}
			tok := anchor.Token(gap)
			rep := readReport(t, runDir)
			if strings.Count(rep, tok) != 1 || !strings.Contains(rep, acceptN+tok+".") {
				t.Errorf("the accepted text does not carry the gap's anchor once at its end:\n%s", rep)
			}
			if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() {
				t.Error("the accept is not applied_verbatim")
			}
			if g := openGap(t, runDir, gap); g.LocationState != record.LocationMarked || g.Location != acceptN+"." {
				t.Errorf("the gap reads %s %q, want marked on red's text", g.LocationState, g.Location)
			}
		})

		t.Run("a cite abutting S"+term, func(t *testing.T) {
			// new keeps S word for word and adds a sentence: the edit puts the cite back at S's end.
			runDir := acceptRun(t)
			gap, err := mintQuote(t, runDir, "G", loc, "--new", acceptKeep+term)
			if err != nil {
				t.Fatal(err)
			}
			cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
			if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
				t.Fatalf("an accept whose text keeps the cited sentence was refused: %v", err)
			}
			rep := readReport(t, runDir)
			if !strings.Contains(rep, acceptS+anchor.Token(cite)+". Analysts disagree about why"+anchor.Token(gap)+".") {
				t.Errorf("the cite is not back at S's end, or the gap's anchor not at the new text's end:\n%s", rep)
			}
			if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() {
				t.Error("the accept is not applied_verbatim")
			}

			// new rewrites S: the accept is refused as stale naming the cite and S, nothing is
			// recorded, and blue's hand edit carrying the cite applies.
			runDir = acceptRun(t)
			gap, err = mintQuote(t, runDir, "G", loc, "--new", acceptN+term)
			if err != nil {
				t.Fatal(err)
			}
			cite = citeSentence(t, runDir, acceptS+".", "https://costs/1")
			before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
			err = acceptEdit(t, runDir, "A", gap, "--accept")
			if err == nil || !strings.Contains(err.Error(), "stale") || !strings.Contains(err.Error(), cite) || !strings.Contains(err.Error(), acceptS) {
				t.Fatalf("an accept that would drop the cite = %v, want stale naming %s and its sentence", err, cite)
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != before {
				t.Errorf("a refused accept recorded %d edit(s)", n-before)
			}
			if err := acceptEdit(t, runDir, "H", gap, "--quote", acceptS+anchor.Token(cite)+anchor.Token(gap)+term,
				"--new", acceptN+anchor.Token(cite)+anchor.Token(gap)+term); err != nil {
				t.Errorf("blue's hand edit carrying the cite was refused: %v", err)
			}
		})

		t.Run("a second gap on S"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			g1, err := mintQuote(t, runDir, "G1", loc, "--new", acceptN+term)
			if err != nil {
				t.Fatal(err)
			}
			g2, err := mintQuote(t, runDir, "G2", loc)
			if err != nil {
				t.Fatal(err)
			}
			if err := acceptEdit(t, runDir, "A", g1, "--accept"); err != nil {
				t.Fatalf("accept after a second gap on the sentence: %v", err)
			}
			if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() {
				t.Error("the accept is not applied_verbatim")
			}
			if rep := readReport(t, runDir); !strings.Contains(rep, acceptN+anchor.Token(g2)+anchor.Token(g1)+".") {
				t.Errorf("both gap anchors are not at the accepted text's end:\n%s", rep)
			}
			if g := openGap(t, runDir, g2); g.LocationState != record.LocationMarked {
				t.Errorf("the second gap reads %s, want marked", g.LocationState)
			}
			w, err := record.WorkJSONOfRun(runtest.Open(t, runDir))
			if err != nil {
				t.Fatal(err)
			}
			for _, g := range w.Open {
				if g.ID == g2 && (len(g.EditedSince) != 1 || g.EditedSince[0].New == "") {
					t.Errorf("the accept is not in the second gap's edited_since: %+v", g.EditedSince)
				}
			}
		})

		t.Run("the board's pair typed"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			gap, err := mintQuote(t, runDir, "G", loc, "--new", acceptN+term)
			if err != nil {
				t.Fatal(err)
			}
			old, new := boardPair(t, runDir, gap)
			if err := acceptEdit(t, runDir, "T", gap, "--quote", old, "--new", new); err != nil {
				t.Fatalf("typing the board's pair: %v", err)
			}
			if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() || b.GetAccepted() {
				t.Errorf("the typed board pair: applied_verbatim=%v accepted=%v, want true and false", b.GetAppliedVerbatim(), b.GetAccepted())
			}
			if _, err := mintQuote(t, runDir, "AGAIN", acceptN+"."); err == nil || !strings.Contains(err.Error(), "estoppel") {
				t.Errorf("a mint against the applied prescription was not estopped: %v", err)
			}
		})

		t.Run("a gap quoting another's anchor"+term, func(t *testing.T) {
			for _, typed := range []bool{false, true} {
				runDir := acceptRun(t)
				g1, err := mintQuote(t, runDir, "G1", loc, "--new", acceptN+term)
				if err != nil {
					t.Fatal(err)
				}
				t1 := anchor.Token(g1)
				g2, err := mintQuote(t, runDir, "G2", acceptS+t1+term, "--new", "Costs rose, though less than feared"+t1+term)
				if err != nil {
					t.Fatalf("a mint quoting the first gap's anchor: %v", err)
				}
				args := []string{"--accept"}
				if typed {
					old, new := boardPair(t, runDir, g2)
					args = []string{"--quote", old, "--new", new}
				}
				if err := acceptEdit(t, runDir, "A", g2, args...); err != nil {
					t.Fatalf("applying the second gap's fix (typed=%v): %v", typed, err)
				}
				rep := readReport(t, runDir)
				if strings.Count(rep, t1) != 1 || strings.Count(rep, anchor.Token(g2)) != 1 {
					t.Errorf("typed=%v: each anchor is not there once:\n%s", typed, rep)
				}
				if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() {
					t.Errorf("typed=%v: not applied_verbatim", typed)
				}
			}
		})

		t.Run("b3 G1's wrapped list item"+term, func(t *testing.T) {
			// The report holds a list item wrapped with a continuation indent; the mint quotes it
			// without the indent, so its stored location differs from the render in whitespace.
			const wrapped = "- The procedure is proven to\n  correctly decide primality for every input"
			const quoted = "The procedure is proven to\ncorrectly decide primality for every input"
			const fix = "The procedure decides primality correctly for every input it was tested on"
			for _, accept := range []bool{true, false} {
				runDir := acceptRunOf(t, "# Method\n\n"+wrapped+".\n- A second item.\n")
				gap, err := mintQuote(t, runDir, "G", quoted+term, "--new", fix+term)
				if err != nil {
					t.Fatal(err)
				}
				args := []string{"--accept"}
				if !accept {
					old, new := boardPair(t, runDir, gap)
					if !strings.Contains(old, "\n  correctly") {
						t.Errorf("the board's fix_old is not the render's bytes: %q", old)
					}
					args = []string{"--quote", old, "--new", new}
				}
				if err := acceptEdit(t, runDir, "A", gap, args...); err != nil {
					t.Fatalf("accept=%v: %v", accept, err)
				}
				if b := lastBody(t, runDir, &recordpb.BlueEdit{}); !b.GetAppliedVerbatim() {
					t.Errorf("accept=%v: the application is not applied_verbatim", accept)
				}
			}
		})
	}
}
