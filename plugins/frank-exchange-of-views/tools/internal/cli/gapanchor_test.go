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
		if !strings.Contains(after, c.want) {
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

// A RETRY ANCHORS THE STORED LOCATION. An act appended without its Anchor — the crash window between
// the two appends — is finished by a retry under its key at the location the act stored, never at
// the retry's own --quote; where the report has since come to hold that location twice, the retry
// refuses and appends nothing.
func TestRetryAnchorsTheStoredLocation(t *testing.T) {
	const stored, other = "Costs rose sharply.", "Volume grew."
	for _, kind := range []string{"finding", "gap"} {
		t.Run(kind, func(t *testing.T) {
			for _, twice := range []bool{false, true} {
				runDir := newRun(t)
				writeReport(t, runDir, "# H\n\n"+stored+"\n\n"+other+"\n")
				registerChairOnce(t, runDir)
				registerLensOnce(t, runDir)
				id := "f-0badf00d"
				var body proto.Message = &recordpb.Finding{Label: proto.String("evidence-F1"), FindingId: proto.String(id),
					FindingKey: proto.String("K1"), Location: proto.String(stored), Text: proto.String("t")}
				if kind == "gap" {
					id = "G1"
					body = &recordpb.Mint{GapId: proto.String(id), MintKey: proto.String("K1"), Class: proto.String("overclaim"),
						Location: proto.String(stored), Problem: proto.String("p"), RequiredFix: proto.String("f"),
						AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
						Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
						Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
				}
				recordtest.Seed(t, runDir, recordtest.At(t, lensSeat, lensSeat+":seeded:"+id, body))
				if twice {
					registerBlue(t, runDir)
					if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
						"--quote", other, "--new", stored, "--reason", "the same sentence again"); err != nil {
						t.Fatalf("doubling the stored sentence: %v", err)
					}
				}
				before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_ANCHOR)
				if v, err := consistency.Check(runtest.Open(t, runDir)); err != nil || !slices.Contains(v, "anchor-record: "+anchor.Kind(id)+" "+id+" has no anchor event") {
					t.Errorf("the consistency check does not name the half-appended %s %s: %v %v", kind, id, v, err)
				}
				var err error
				if kind == "gap" {
					_, err = mintQuote(t, runDir, "K1", other)
				} else {
					_, err = run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "K1", "--quote", other,
						"--reason", "t", "--severity", "low", "--likelihood", "low", "--impact", "low")
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
