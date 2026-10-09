package cli

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/proof"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// mintQuote mints a gap at quote from the fixture lens, with any further flags, and returns its id.
func mintQuote(t *testing.T, runDir, key, quote string, extra ...string) (string, error) {
	t.Helper()
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	out, err := runMint(t, runDir, append([]string{"--run", runDir, "--seat-id", lensSeat,
		"--key", key, "--class", "overclaim", "--problem", "the defect", "--fix", "the fix",
		"--check-kind", "document", "--check", "the acceptance check", "--severity", "medium",
		"--likelihood", "medium", "--impact", "medium", "--complexity", "low", "--quote", quote}, extra...)...)
	return gapID(out), err
}

// openGap is gap id's entry on the board a seat reads.
func openGap(t *testing.T, runDir, id string) record.GapJSON {
	t.Helper()
	for _, g := range boardOf(t, runDir).Open {
		if g.ID == id {
			return g
		}
	}
	t.Fatalf("gap %s is not open on the board", id)
	return record.GapJSON{}
}

// A GAP'S ANCHOR SITS WHERE EVERY KIND'S DOES: right after the quote's last content character,
// before its punctuation, whether the quote is a fragment ending mid-sentence or the whole
// sentence. The board's location is the one sentence holding it, its passage that sentence's
// section, and the render read back from the record is the bytes the write placed.
func TestGapAnchorSitsAtTheQuoteEnd(t *testing.T) {
	const base = "# Costs\n\nQ1 was hard. Costs rose sharply in the first quarter, then fell.\n\nVolume grew.\n\n# Other\n\nNothing here.\n"
	runDir := newRun(t)
	writeReport(t, runDir, base)
	for _, c := range []struct{ quote, want, location string }{
		{"Costs rose sharply", "Costs rose sharply<!--gap:G1--> in the first quarter, then fell.", "Costs rose sharply in the first quarter, then fell."},
		{"Volume grew.", "Volume grew<!--gap:G2-->.", "Volume grew."},
	} {
		before := readReport(t, runDir)
		id, err := mintQuote(t, runDir, c.quote, c.quote)
		if err != nil {
			t.Fatalf("mint --quote %q: %v", c.quote, err)
		}
		after := readReport(t, runDir)
		if !strings.Contains(after, handleText(t, runDir, c.want)) {
			t.Errorf("mint --quote %q placed its anchor wrong:\n%s", c.quote, after)
		}
		if placed, err := anchortext.Attach(before, id, c.quote); err != nil || placed != after {
			t.Errorf("the render read back from the record is not the bytes the write placed (%v):\n write %q\nrender %q", err, placed, after)
		}
		g := openGap(t, runDir, id)
		if g.LocationState != record.LocationMarked || g.Location != c.location {
			t.Errorf("gap %s reads %s %q, want marked %q", id, g.LocationState, g.Location, c.location)
		}
		if !strings.HasPrefix(g.Passage, "# Costs") || strings.Contains(g.Passage, "Nothing here") {
			t.Errorf("gap %s's passage is not its own section:\n%s", id, g.Passage)
		}
	}
	// A GAP ANCHOR IS NO CLAIM AND NO EVIDENCE: the claim count does not move and assembly strips it.
	if got := claimcount.Count(readReport(t, runDir)); got != claimcount.Count(base) {
		t.Errorf("minting moved the claim count %d -> %d", claimcount.Count(base), got)
	}
	if asm := assembled(t, runDir); strings.Contains(asm, "<!--gap:") {
		t.Errorf("the assembled report holds a gap anchor:\n%s", asm)
	}
}

// MINT AND FINDING REFUSE WHAT THE UNIQUE-QUOTE CHECK REFUSES, in the placement words, and record
// nothing. A quote that occurs twice in two paragraphs is ambiguous, including one whose first
// match crosses a blank line — it is told it repeats, never that it may not cross; #552's shape,
// a quote across a blank line holding a phrase that occurs earlier, is refused for crossing.
func TestMintRefusesWhatLocateUniqueRefused(t *testing.T) {
	const report = "# H\n\nThe cost is climbing sharply.\n\nBlue wrote that the cost is \"climbing sharply\"\n\nand gave no source for it.\n\n" +
		"Costs rose.\n\nCosts rose.\n\nPlain one.\n\nPlain two.\n\nWe saw costs\n\nrose. Then we saw costs rose again.\n\n```\ncode line here\n```\n"
	for _, c := range []struct{ name, quote, says string }{
		{"absent", "a sentence nobody wrote", "was not found in report.md"},
		{"twice in two paragraphs", "Costs rose.", "occurs more than once in the report"},
		{"splits a word", "ave no source", "starts or ends inside a word"},
		{"crosses a blank line", "Plain one. Plain two.", "runs across a blank line"},
		{"#552's shape", "Blue wrote that the cost is \"climbing sharply\" and gave no source for it.", "runs across a blank line"},
		{"first match crosses, also inside one paragraph", "saw costs rose", "occurs more than once in the report"},
		{"inside a fence", "code line here", "code fence"},
	} {
		for _, verb := range []string{"mint", "finding"} {
			t.Run(verb+"/"+c.name, func(t *testing.T) {
				runDir := newRun(t)
				writeReport(t, runDir, report)
				registerChairOnce(t, runDir)
				registerLensOnce(t, runDir)
				placed := func() int {
					return countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT) + countType(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING) +
						countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR)
				}
				before := placed()
				var err error
				if verb == "mint" {
					_, err = mintQuote(t, runDir, "K", c.quote)
				} else {
					_, err = run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", c.quote,
						"--reason", "r", "--severity", "low", "--likelihood", "low", "--impact", "low")
				}
				if err == nil || !strings.Contains(err.Error(), c.says) {
					t.Fatalf("refusal = %v, want it to say %q", err, c.says)
				}
				if strings.Contains(err.Error(), "one edit per site") {
					t.Errorf("a placement refusal advises one edit per site: %v", err)
				}
				if after := placed(); after != before {
					t.Errorf("a refused %s appended %d event(s)", verb, after-before)
				}
			})
		}
	}
}

// NO VERB PLACES AN ANCHOR ON A HEADING. Every placing verb refuses a quote that ends in a heading,
// appends nothing, and says so; a gap or a finding is told how to name the whole section instead.
func TestNoVerbPlacesAnAnchorOnAHeading(t *testing.T) {
	const report = "# Title\n\n## Method\n\nThe sieve runs once.\n"
	for _, c := range []struct {
		verb    string
		section bool
		act     func(t *testing.T, runDir string) error
	}{
		{"lens mint", true, func(t *testing.T, runDir string) error { _, err := mintQuote(t, runDir, "K", "Method"); return err }},
		{"lens finding", true, func(t *testing.T, runDir string) error {
			registerLensOnce(t, runDir)
			_, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", "Method",
				"--reason", "r", "--severity", "low", "--likelihood", "low", "--impact", "low")
			return err
		}},
		{"blue cite", false, func(t *testing.T, runDir string) error {
			registerBlue(t, runDir)
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://m/1": []byte("<html>the method</html>")}})
			_, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", "Method", "--url", "https://m/1", "--title", "M")
			return err
		}},
		{"blue prove", false, func(t *testing.T, runDir string) error {
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond"); err != nil {
				t.Fatal(err)
			}
			_, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond", "--quote", "Method",
				"--script", script(t, runDir, "m.js", "console.log(1)"), "--reason", "r")
			return err
		}},
		{"lens corroborate", false, func(t *testing.T, runDir string) error {
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
				t.Fatal(err)
			}
			_, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence", "--url", "https://m/2", "--title", "M",
				"--quote", "Method", "--as", "supports", "--confidence", "high", "--reason", "r")
			return err
		}},
	} {
		t.Run(c.verb, func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, report)
			err := c.act(t, runDir)
			if err == nil || !strings.Contains(err.Error(), c.verb+": the quote ends in a section heading") {
				t.Fatalf("refusal = %v, want %s to refuse the heading", err, c.verb)
			}
			if got := strings.Contains(err.Error(), "--about-kind section"); got != c.section {
				t.Errorf("refusal names --about-kind section = %v, want %v: %v", got, c.section, err)
			}
			if !strings.Contains(err.Error(), "quote a sentence of the section's text") {
				t.Errorf("refusal does not say what to quote: %v", err)
			}
			for _, typ := range []recordpb.EventType{recordpb.EventType_EVENT_TYPE_MINT, recordpb.EventType_EVENT_TYPE_FINDING,
				recordpb.EventType_EVENT_TYPE_ANCHOR, recordpb.EventType_EVENT_TYPE_CITE, recordpb.EventType_EVENT_TYPE_PROOF, recordpb.EventType_EVENT_TYPE_VERIFY} {
				if n := countType(t, runDir, typ); n != 0 {
					t.Errorf("a refused %s appended %d %v event(s)", c.verb, n, typ)
				}
			}
		})
	}
}

// NO EDIT PUTS AN ANCHOR ON A HEADING. An edit or a prescription that carries an anchor into a
// heading line, or makes a heading of the line an anchor stands on, is refused and records nothing;
// an anchor already on a heading — an archived carry — refuses no edit, its own heading's included.
func TestNoEditPutsAnAnchorOnAHeading(t *testing.T) {
	const report = "# Title\n\n## Method<!--fx:F-0000aaaa-->\n\nThe sieve runs once<!--fx:F-0000bbbb-->.\n"
	for _, c := range []struct{ name, quote, new string }{
		{"carried into a heading", "The sieve runs once<!--fx:F-0000bbbb-->", "## The sieve runs once<!--fx:F-0000bbbb-->"},
		{"made a heading around it", "The sieve", "## The sieve"},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, report)
			registerBlue(t, runDir)
			_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1", "--quote", c.quote, "--new", c.new, "--reason", "r")
			if want := "blue edit: this edit puts <!--fx:F-0000bbbb--> on a section heading"; err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("refusal = %v, want %q", err, want)
			}
			if !strings.Contains(err.Error(), "keep the anchor on its sentence in the section's text") {
				t.Errorf("refusal does not say where the anchor stays: %v", err)
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
				t.Errorf("a refused edit appended %d blue_edit event(s)", n)
			}
		})
	}
	t.Run("a prescription", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, report)
		_, err := mintQuote(t, runDir, "G", "The sieve", "--new", "## The sieve")
		if err == nil || !strings.Contains(err.Error(), "lens mint: this edit puts <!--fx:F-0000bbbb--> on a section heading") {
			t.Fatalf("refusal = %v, want lens mint to refuse the heading", err)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
			t.Errorf("a refused mint appended %d mint event(s)", n)
		}
	})
	for _, c := range []struct{ name, quote, new, want string }{
		{"an archived anchor on a heading, another edit", "sieve runs", "sieve ran", "## Method<!--fx:F-0000aaaa-->\n\nThe sieve ran once"},
		{"an archived anchor on a heading, its heading", "## Method<!--fx:F-0000aaaa-->", "## Methods<!--fx:F-0000aaaa-->", "## Methods<!--fx:F-0000aaaa-->\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, report)
			registerBlue(t, runDir)
			if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1", "--quote", c.quote, "--new", c.new, "--reason", "r"); err != nil {
				t.Fatalf("an edit that puts no anchor on a heading is refused: %v", err)
			}
			if rep := readReport(t, runDir); !strings.Contains(rep, c.want) {
				t.Errorf("report after the edit lacks %q:\n%s", c.want, rep)
			}
		})
	}
}

// EVERY PLACING ACT IS FOLLOWED BY ITS ANCHOR. A citation, a proof and a corroboration each append
// their own event and then one Anchor carrying the act's id and its quote, from the same seat; the
// report holds the marker because of that Anchor, so an act with none places nothing.
func TestEveryPlacerAppendsItsAnchorAfterItsAct(t *testing.T) {
	const quote = "Costs rose sharply."
	for _, c := range []struct {
		seat string
		args []string
		id   func(*record.Event) string
	}{
		{blueSeat, []string{"cite", "--url", "https://example.org/s", "--title", "S"},
			func(e *record.Event) string { return e.GetCite().GetLabel() }},
		{blueSeat, []string{"prove", "--script", "p.js", "--reason", "r"},
			func(e *record.Event) string { return e.GetProof().GetProofId() }},
		{lensSeat, []string{"corroborate", "--url", "https://example.org/r", "--title", "R", "--as", "supports", "--confidence", "high", "--reason", "r"},
			func(e *record.Event) string { return e.GetVerify().GetLabel() }},
	} {
		t.Run(c.args[0], func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, "# H\n\n"+quote+"\n")
			registerLensOnce(t, runDir)
			registerBlue(t, runDir)
			script(t, runDir, "p.js", "console.log(1)")
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://example.org/s": []byte("<html>a source</html>")}})
			if _, err := run(t, append([]string{c.args[0], "--run", runDir, "--seat-id", c.seat, "--quote", quote}, c.args[1:]...)...); err != nil {
				t.Fatal(err)
			}
			m, err := record.MergedEvents(runtest.Open(t, runDir))
			if err != nil || len(m.Events) < 2 {
				t.Fatal(err)
			}
			act, placed := m.Events[len(m.Events)-2], m.Events[len(m.Events)-1].GetAnchor()
			if id := c.id(act); id == "" || placed.GetId() != id || placed.GetLocation() != quote || m.Events[len(m.Events)-1].GetSeatId() != c.seat {
				t.Fatalf("the act %s is followed by anchor {%q %q}, want its own id at %q from %s", id, placed.GetId(), placed.GetLocation(), quote, c.seat)
			}
			if !strings.Contains(readReport(t, runDir), "Costs rose sharply"+anchor.Token(placed.GetId())+".") {
				t.Errorf("the report does not hold the anchor:\n%s", readReport(t, runDir))
			}
			assertNoAnchorViolations(t, runDir)
		})
	}
}

// A STORED LOCATION IS THE QUOTE'S VISIBLE TEXT, for every act that places an anchor. A seat quotes a
// sentence as `show report` prints it, with the anchors already on it; the act and its Anchor store
// that quote with every anchor token out, and the new anchor stands where the typed quote puts it —
// at the end of the sentence's words, before the anchors it already carried. Reproduced from
// universe m16, where a mint, a finding and their Anchors stored `…rejects it)<!--proof:…-->`.
func TestEveryPlacerStoresItsQuoteWithoutAnchors(t *testing.T) {
	const held, visible = "<!--fx:F-0000beef-->", "Costs rose sharply."
	const quote = "Costs rose sharply" + held + "."
	for _, c := range []struct {
		name, seat string
		args       []string
		id, loc    func(*record.Event) string
	}{
		{"finding", lensSeat, []string{"finding", "--key", "K1", "--reason", "t", "--severity", "low", "--likelihood", "low", "--impact", "low"},
			func(e *record.Event) string { return e.GetFinding().GetId() }, func(e *record.Event) string { return e.GetFinding().GetLocation() }},
		{"mint", lensSeat, nil,
			func(e *record.Event) string { return e.GetMint().GetGapId() }, func(e *record.Event) string { return e.GetMint().GetLocation() }},
		{"cite", blueSeat, []string{"cite", "--url", "https://example.org/s", "--title", "S"},
			func(e *record.Event) string { return e.GetCite().GetLabel() }, func(e *record.Event) string { return e.GetCite().GetLocation() }},
		{"prove", blueSeat, []string{"prove", "--script", "p.js", "--reason", "r"},
			func(e *record.Event) string { return e.GetProof().GetProofId() }, func(e *record.Event) string { return e.GetProof().GetLocation() }},
		{"corroborate", lensSeat, []string{"corroborate", "--url", "https://example.org/r", "--title", "R", "--as", "supports", "--confidence", "high", "--reason", "r"},
			func(e *record.Event) string { return e.GetVerify().GetLabel() }, func(e *record.Event) string { return e.GetVerify().GetClaim() }},
	} {
		t.Run(c.name, func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, "# H\n\n"+quote+"\n")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			registerBlue(t, runDir)
			script(t, runDir, "p.js", "console.log(1)")
			withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://example.org/s": []byte("<html>a source</html>")}})
			var err error
			if c.args == nil {
				_, err = mintQuote(t, runDir, "K1", quote)
			} else {
				_, err = run(t, append([]string{c.args[0], "--run", runDir, "--seat-id", c.seat, "--quote", quote}, c.args[1:]...)...)
			}
			if err != nil {
				t.Fatal(err)
			}
			m, err := record.MergedEvents(runtest.Open(t, runDir))
			if err != nil || len(m.Events) < 2 {
				t.Fatal(err)
			}
			act, placed := m.Events[len(m.Events)-2], m.Events[len(m.Events)-1].GetAnchor()
			if id := c.id(act); id == "" || placed.GetId() != id || c.loc(act) != visible || placed.GetLocation() != visible {
				t.Fatalf("the act %s stores %q and its anchor %s stores %q, want %q in both", id, c.loc(act), placed.GetId(), placed.GetLocation(), visible)
			}
			if want := "Costs rose sharply" + anchor.Token(placed.GetId()) + held + "."; !strings.Contains(readReport(t, runDir), want) {
				t.Errorf("the anchor does not stand before the one the sentence carried, as %q:\n%s", want, readReport(t, runDir))
			}
		})
	}
}

// A RETRY ANCHORS THE STORED LOCATION, for every act that places an anchor. An act appended without
// its Anchor — the crash window between the two appends — is named by the consistency check and
// finished by a retry under its key at the location the act stored, never at the retry's own
// --quote; where the report has since come to hold that location twice, the retry refuses and
// appends nothing. A corroboration's key is its source and claim, so its retry quotes the claim.
func TestRetryAnchorsTheStoredLocation(t *testing.T) {
	const stored, other = "Costs rose sharply.", "Volume grew."
	for _, c := range []struct {
		kind, id, seat string
		body           func(runDir string) proto.Message
		retry          []string
	}{
		{"finding", "F-0badf00d", lensSeat, func(string) proto.Message {
			return &recordpb.Finding{Id: proto.String("F-0badf00d"), FindingKey: proto.String("K1"), Location: proto.String(stored), Text: proto.String("t")}
		}, []string{"finding", "--key", "K1", "--quote", other, "--reason", "t", "--severity", "low", "--likelihood", "low", "--impact", "low"}},
		{"gap", "G-0badf00d", lensSeat, func(string) proto.Message {
			return &recordpb.Mint{GapId: proto.String("G-0badf00d"), MintKey: proto.String("K1"), Class: proto.String("overclaim"),
				Location: proto.String(stored), Problem: proto.String("p"), RequiredFix: proto.String("f"),
				AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
				Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
				Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
		}, nil},
		{"citation", "C-0badf00d", blueSeat, func(string) proto.Message {
			return &recordpb.Cite{Label: proto.String("C-0badf00d"), CiteKey: proto.String("K1"), Location: proto.String(stored),
				Url: proto.String("https://example.org/s"), Title: proto.String("S")}
		}, []string{"cite", "--key", "K1", "--quote", other, "--url", "https://example.org/s", "--title", "S"}},
		{"proof", "P-0badf00d", blueSeat, func(runDir string) proto.Message {
			sha, err := proof.ScriptSha(runDir, script(t, runDir, "p.js", "console.log(1)"))
			if err != nil {
				t.Fatal(err)
			}
			return &recordpb.Proof{ProofId: proto.String("P-0badf00d"), ProofKey: proto.String("K1"), ProofSha: proto.String(sha),
				Location: proto.String(stored), Script: proto.String("p.js")}
		}, []string{"prove", "--key", "K1", "--quote", other, "--script", "p.js", "--reason", "r"}},
		{"citation", "C-0badf00e", lensSeat, func(string) proto.Message {
			return &recordpb.Verify{Label: proto.String("C-0badf00e"), Url: proto.String("https://example.org/r"), Claim: proto.String(stored),
				Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH), Text: proto.String("r")}
		}, []string{"corroborate", "--url", "https://example.org/r", "--title", "R", "--quote", stored, "--as", "supports", "--confidence", "high", "--reason", "r"}},
	} {
		t.Run(c.kind+" "+c.id, func(t *testing.T) {
			for _, twice := range []bool{false, true} {
				runDir := newRun(t)
				writeReport(t, runDir, "# H\n\n"+stored+"\n\n"+other+"\n")
				registerChairOnce(t, runDir)
				registerLensOnce(t, runDir)
				registerBlue(t, runDir)
				id := c.id
				recordtest.Seed(t, runDir, recordtest.At(t, c.seat, c.seat+":seeded:"+id, c.body(runDir)))
				if twice {
					if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
						"--quote", other, "--new", stored, "--reason", "the same sentence again"); err != nil {
						t.Fatalf("doubling the stored sentence: %v", err)
					}
				}
				before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR)
				if v, err := consistency.Check(runtest.Open(t, runDir)); err != nil || !slices.Contains(v, "anchor-record: "+c.kind+" "+id+" has no anchor event") {
					t.Errorf("the consistency check does not name the half-appended %s %s: %v %v", c.kind, id, v, err)
				}
				var err error
				if c.retry == nil {
					_, err = mintQuote(t, runDir, "K1", other)
				} else {
					_, err = run(t, append([]string{c.retry[0], "--run", runDir, "--seat-id", c.seat}, c.retry[1:]...)...)
				}
				got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR) - before
				if twice {
					if err == nil || got != 0 {
						t.Errorf("with the stored location twice in the report the retry gave %v and appended %d anchor(s)", err, got)
					}
					continue
				}
				if err != nil || got != 1 {
					t.Fatalf("the retry gave %v and appended %d anchor(s), want one", err, got)
				}
				if a := lastBody(t, runDir, &recordpb.Anchor{}); a.GetId() != id || a.GetLocation() != stored {
					t.Errorf("the retry anchored %s at %q, want %s at the stored %q", a.GetId(), a.GetLocation(), id, stored)
				}
				if !strings.Contains(readReport(t, runDir), "Costs rose sharply"+anchor.Token(id)+".") {
					t.Errorf("the anchor is not at the stored location:\n%s", readReport(t, runDir))
				}
				assertNoAnchorViolations(t, runDir)
			}
		})
	}
}

// EDITED_SINCE IS THE EDITS THAT CHANGED THE SENTENCE HOLDING THE GAP'S ANCHOR. A fragment edit
// inside that sentence is listed though its old span excludes the anchor; an edit to an identical
// sentence elsewhere is not; the edit that cuts the sentence to the bare anchor is not listed, and
// the gap reads gone.
func TestEditedSinceSeesTheSentence(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nQ1 was hard. Costs rose sharply in the first quarter.\n\nIn Q3 the picture changed. Costs rose sharply in the first quarter.\n")
	id, err := mintQuote(t, runDir, "G", "Q1 was hard. Costs rose sharply in the first quarter.")
	if err != nil {
		t.Fatal(err)
	}
	tok := anchor.Token(id)
	registerBlue(t, runDir)
	edit := func(key, old, new string) {
		t.Helper()
		if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", key,
			"--quote", old, "--new", new, "--reason", "r"); err != nil {
			t.Fatalf("edit %s: %v", key, err)
		}
	}
	edited := func() record.WorkGapJSON {
		t.Helper()
		out, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
		if err != nil {
			t.Fatal(err)
		}
		var w record.WorkJSON
		if err := json.Unmarshal([]byte(out), &w); err != nil {
			t.Fatal(err)
		}
		if len(w.Open) != 1 {
			t.Fatalf("want one open gap: %+v", w.Open)
		}
		return w.Open[0]
	}
	edit("ELSEWHERE", "In Q3 the picture changed. Costs rose sharply in the first quarter.", "In Q3 the picture changed. Costs fell.")
	if g := edited(); len(g.EditedSince) != 0 {
		t.Errorf("an edit to an identical sentence elsewhere is listed: %+v", g.EditedSince)
	}
	edit("FRAGMENT", "rose sharply", "climbed")
	if g := edited(); len(g.EditedSince) != 1 || g.EditedSince[0].Old != "rose sharply" || g.LocationState != record.LocationMarked {
		t.Errorf("the fragment edit inside the gap's sentence is not listed: %s %+v", g.LocationState, g.EditedSince)
	}
	edit("CUT", "Costs climbed in the first quarter"+tok+".", tok)
	if g := edited(); len(g.EditedSince) != 1 || g.LocationState != record.LocationGone {
		t.Errorf("after the cut the gap reads %s with %d edit(s), want gone with the fragment edit alone", g.LocationState, len(g.EditedSince))
	}
}
