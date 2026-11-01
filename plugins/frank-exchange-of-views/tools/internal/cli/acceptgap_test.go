package cli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

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

		t.Run("the finding the gap was minted from"+term, func(t *testing.T) {
			// m25 G-96ac6e01: a lens's finding on S, a gap minted from it prescribing S's rewording, and
			// blue's accept — refused as stale for dropping the finding, on a report nothing had changed.
			runDir := acceptRun(t)
			if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", acceptS+".",
				"--reason", "overclaim", "--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
				t.Fatalf("finding: %v", err)
			}
			fx := lastBody(t, runDir, &recordpb.Finding{}).GetId()
			gap, err := mintQuote(t, runDir, "G", loc, "--new", acceptN+term, "--found-by", fx)
			if err != nil {
				t.Fatal(err)
			}
			if m := lastBody(t, runDir, &recordpb.Mint{}); m.GetFixBasis() != "verified" {
				t.Errorf("the mint's fix_basis is %q, want verified", m.GetFixBasis())
			}
			tok, ftok := anchor.Token(gap), anchor.Token(fx)
			if old, new := boardPair(t, runDir, gap); old != acceptS+tok+ftok+term || new != acceptN+tok+ftok+term {
				t.Errorf("the board's pair is %q -> %q, want the finding carried with the gap's anchor", old, new)
			}
			if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
				t.Fatalf("accepting a fix on the sentence its finding marks: %v", err)
			}
			rep := readReport(t, runDir)
			if strings.Count(rep, ftok) != 1 || strings.Count(rep, tok) != 1 || !strings.Contains(rep, acceptN+tok+ftok+".") {
				t.Errorf("the accepted sentence does not carry the gap's and the finding's anchors once at its end:\n%s", rep)
			}
			b := lastBody(t, runDir, &recordpb.BlueEdit{})
			if !b.GetAccepted() || !b.GetAppliedVerbatim() || b.GetAnswers() != gap {
				t.Errorf("the accept: accepted=%v applied_verbatim=%v answers=%q, want true, true, %s", b.GetAccepted(), b.GetAppliedVerbatim(), b.GetAnswers(), gap)
			}
			if !slices.Contains(b.GetReopened(), fx) {
				t.Errorf("the accept does not record the finding's sentence as changed: reopened %v", b.GetReopened())
			}
			if g := openGap(t, runDir, gap); g.LocationState != record.LocationMarked || g.Location != acceptN+"." {
				t.Errorf("the gap reads %s %q, want marked on red's text", g.LocationState, g.Location)
			}
		})

		t.Run("a citation on S when red mints"+term, func(t *testing.T) {
			// The accept carries no citation onto red's rewording, so the mint refuses to verify one:
			// nothing is recorded, and the refusal names the citation and the three ways on.
			runDir := acceptRun(t)
			cite := citeSentence(t, runDir, acceptS+".", "https://costs/1")
			before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
			_, err := mintQuote(t, runDir, "G", loc, "--new", acceptN+term)
			if err == nil {
				t.Fatal("a mint verified a rewording the accept refuses: the citation on the sentence would not survive it")
			}
			// The form that writes the anchor into --new names the anchor as the seat types it, and asks
			// nothing of --quote (#1663 smoke, ev130/172/205).
			for _, want := range []string{"NOTHING IS RECORDED", cite, "word for word", "or write " + anchor.Token(cite) + " into --new after the text it marks — --quote stays as it is", "--fix"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not say %q: %v", want, err)
				}
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != before {
				t.Errorf("a refused mint recorded %d mint(s)", n-before)
			}

			// The two prescriptions the refusal offers are verified, and the accept applies each.
			for i, c := range []struct{ new, want func(ctok string) string }{
				{func(string) string { return acceptKeep + term }, func(ctok string) string { return acceptS + ctok + ". Analysts disagree about why" }},
				{func(ctok string) string { return acceptN + ctok + term }, func(string) string { return acceptN }},
			} {
				runDir := acceptRun(t)
				ctok := anchor.Token(citeSentence(t, runDir, acceptS+".", "https://costs/1"))
				gap, err := mintQuote(t, runDir, "G", loc, "--new", c.new(ctok))
				if err != nil {
					t.Fatalf("offered prescription %d: %v", i, err)
				}
				if err := acceptEdit(t, runDir, "A", gap, "--accept"); err != nil {
					t.Fatalf("offered prescription %d verified, then refused at accept: %v", i, err)
				}
				rep := readReport(t, runDir)
				if strings.Count(rep, ctok) != 1 || strings.Count(rep, anchor.Token(gap)) != 1 || !strings.Contains(rep, c.want(ctok)) {
					t.Errorf("offered prescription %d: the report does not hold red's text with each anchor once:\n%s", i, rep)
				}
			}
		})

		t.Run("a second gap on S"+term, func(t *testing.T) {
			runDir := acceptRun(t)
			g1, err := mintQuote(t, runDir, "G1", loc, "--new", acceptN+term)
			if err != nil {
				t.Fatal(err)
			}
			g2, err := mintQuote(t, runDir, "G2", loc, "--class", otherClass)
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
			if tok := anchor.Token(gap); old != acceptS+tok+term || new != acceptN+tok+term {
				t.Errorf("the board's pair is %q -> %q, want the location through its anchor, then its own terminator", old, new)
			}
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
				g2, err := mintQuote(t, runDir, "G2", acceptS+t1+term, "--new", "Costs rose, though less than feared"+t1+term, "--class", otherClass)
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

// BLUE'S WORK ITEM CARRIES THE PAIR THE BOARD SERVES — the pair `--accept` sends — read by `show work`
// as blue-respond over the real render, so blue acts on a proposed fix without a board read. A gap
// whose lens proposed no text says `proposed` and has no pair.
func TestShowWorkCarriesTheBoardsFixPairForBlue(t *testing.T) {
	runDir := acceptRunOf(t, "# Method\n\nThe procedure is proven correct. A second claim stands here.\n")
	quoted, err := mintQuote(t, runDir, "G", "The procedure is proven correct.", "--new", "The procedure is tested, not proven.")
	if err != nil {
		t.Fatal(err)
	}
	prose, err := mintQuote(t, runDir, "H", "A second claim stands here.")
	if err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "work", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("show work: %v", err)
	}
	var w struct {
		Open []map[string]any `json:"open"`
	}
	if err := json.Unmarshal([]byte(out), &w); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	items := map[string]map[string]any{}
	for _, g := range w.Open {
		items[g["id"].(string)] = g
	}
	old, new := boardPair(t, runDir, quoted)
	if !strings.Contains(old, anchor.Token(quoted)) {
		t.Fatalf("the board's fix_old does not run through the gap's anchor: %q", old)
	}
	if g := items[quoted]; g["fix_basis"] != "verified" || g["fix_old"] != old || g["fix_new"] != new {
		t.Errorf("work item %s fix = (%v, %q, %q), want the board's (verified, %q, %q)", quoted, g["fix_basis"], g["fix_old"], g["fix_new"], old, new)
	}
	g := items[prose]
	if g["fix_basis"] != "proposed" {
		t.Errorf("work item %s fix_basis = %v, want proposed", prose, g["fix_basis"])
	}
	for _, k := range []string{"fix_old", "fix_new"} {
		if v, has := g[k]; has {
			t.Errorf("work item %s has %s = %v; its lens proposed no text", prose, k, v)
		}
	}
}

// acceptFinding files a lens finding at quote and returns its id.
func acceptFinding(t *testing.T, runDir, key, quote string) string {
	t.Helper()
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", key, "--quote", quote,
		"--reason", "overclaim", "--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("finding: %v", err)
	}
	return lastBody(t, runDir, &recordpb.Finding{}).GetId()
}

// recordedPrescription appends a gap prescribing fix at quote with no mint verb in between — the
// mint body of a prose-only gap on a scratch run of the same report, carrying the pair as recorded,
// with its anchor — so the accept can be asked about a pair the mint never judged.
func recordedPrescription(t *testing.T, body, runDir, quote, fix string) string {
	t.Helper()
	scratch := acceptRunOf(t, body)
	if _, err := mintQuote(t, scratch, "TEMPLATE", quote); err != nil {
		t.Fatalf("the template mint: %v", err)
	}
	m := proto.Clone(lastBody(t, scratch, &recordpb.Mint{})).(*recordpb.Mint)
	gap := record.NewID("gap")
	m.GapId, m.MintKey, m.FixNew, m.FixBasis = proto.String(gap), proto.String("RECORDED"), proto.String(fix), proto.String("verified")
	m.ClassMaterial = nil // stamped by the write path
	if _, err := record.AppendPlaced(record.Identity{Run: runtest.Open(t, runDir), SeatID: lensSeat}, m, gap, quote); err != nil {
		t.Fatalf("recording the prescription: %v", err)
	}
	return gap
}

// MINT VERIFIES EXACTLY WHAT THE ACCEPT APPLIES. Each row is a report, what stands on it when red
// mints, and red's pair. Run A mints it through the verb; run B, the same state, records the pair with
// no mint judgement and asks the accept. Verified in A ⇔ applied in B, and a fix verified in A is
// applied in A. The authoring bound is outside the equivalence: it is what red may prescribe, not
// whether the accept applies it.
func TestMintVerifiesExactlyWhatTheAcceptApplies(t *testing.T) {
	const costs = "# Costs\n\nThe year opened badly. " + acceptS + ". Volume recovered later.\n"
	const stray = "# H\n\nIntro.\n\n(They stand " + strayPeriod + "\n"
	const heading = "# Title\n\n## Method\n\nThe sieve runs once.\n"
	const wrapped = "# Method\n\n- The procedure is proven to\n  correctly decide primality for every input.\n- A second item.\n"
	type row struct {
		name, body string
		// on puts what stands on the report when red mints, and returns red's quote and text.
		on   func(t *testing.T, runDir string) (quote, fix string)
		want bool
	}
	var rows []row
	for _, term := range []string{"", "."} {
		loc := acceptS + term
		rows = append(rows,
			row{"plain" + term, costs, func(*testing.T, string) (string, string) { return loc, acceptN + term }, true},
			row{"the finding on S" + term, costs, func(t *testing.T, d string) (string, string) {
				acceptFinding(t, d, "F1", acceptS+".")
				return loc, acceptN + term
			}, true},
			row{"a finding inside S" + term, costs, func(t *testing.T, d string) (string, string) {
				acceptFinding(t, d, "F1", "Costs rose sharply")
				return loc, acceptN + term
			}, false},
			row{"a citation on S, reworded" + term, costs, func(t *testing.T, d string) (string, string) {
				citeSentence(t, d, acceptS+".", "https://costs/1")
				return loc, acceptN + term
			}, false},
			row{"a citation on S, kept" + term, costs, func(t *testing.T, d string) (string, string) {
				citeSentence(t, d, acceptS+".", "https://costs/1")
				return loc, acceptKeep + term
			}, true},
			row{"a citation on S, carried" + term, costs, func(t *testing.T, d string) (string, string) {
				return loc, acceptN + anchor.Token(citeSentence(t, d, acceptS+".", "https://costs/1")) + term
			}, true},
			row{"a second gap on S" + term, costs, func(t *testing.T, d string) (string, string) {
				if _, err := mintQuote(t, d, "G0", loc, "--class", otherClass); err != nil {
					t.Fatal(err)
				}
				return loc, acceptN + term
			}, true},
			row{"quoting another gap's anchor" + term, costs, func(t *testing.T, d string) (string, string) {
				g1, err := mintQuote(t, d, "G0", loc, "--new", acceptN+term, "--class", otherClass)
				if err != nil {
					t.Fatal(err)
				}
				t1 := anchor.Token(g1)
				return acceptS + t1 + term, "Costs rose, though less than feared" + t1 + term
			}, true},
			row{"a wrapped list item" + term, wrapped, func(*testing.T, string) (string, string) {
				return "The procedure is proven to\ncorrectly decide primality for every input" + term, "The procedure decides primality correctly for every input it was tested on" + term
			}, true},
		)
	}
	rows = append(rows,
		row{"a punctuation repair on the literal span", stray, func(*testing.T, string) (string, string) { return strayPeriod, "on their own.)" }, true},
		row{"a repair that changes nothing", stray, func(*testing.T, string) (string, string) { return "on their own.);", "on their own.)" }, false},
		row{"red's text made a heading", heading, func(*testing.T, string) (string, string) { return "The sieve", "## The sieve" }, false},
	)
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			a := acceptRunOf(t, r.body)
			quote, fix := r.on(t, a)
			before := countType(t, a, recordpb.EventType_EVENT_TYPE_MINT)
			gap, err := mintQuote(t, a, "G", quote, "--new", fix)
			verified := err == nil && lastBody(t, a, &recordpb.Mint{}).GetFixBasis() == "verified"
			if err != nil && countType(t, a, recordpb.EventType_EVENT_TYPE_MINT) != before {
				t.Errorf("a refused mint recorded a mint")
			}
			if verified {
				if err := acceptEdit(t, a, "A", gap, "--accept"); err != nil {
					t.Errorf("a fix the mint verified is refused at accept: %v", err)
				}
			}

			b := acceptRunOf(t, r.body)
			quote, fix = r.on(t, b)
			applied := acceptEdit(t, b, "A", recordedPrescription(t, r.body, b, quote, fix), "--accept") == nil
			if verified != applied || verified != r.want {
				t.Errorf("mint verified=%v (%v), accept applied=%v, want both %v", verified, err, applied, r.want)
			}
		})
	}
}
