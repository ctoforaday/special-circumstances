package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// RED'S ARGUMENT FOR A GAP MUST REACH THE SEATS THAT ANSWER AND WEIGH IT.
//
// `mint` took --reason and threw it away whenever --problem was also given, which is every ordinary
// mint. The write still answered "minted G1", and --reason's own help promises "the substance the
// report renders and the other side answers" — a promise this verb did not keep.
//
// MEASURED 2026-08-16 BY ASKING THE BENCH rather than by watching one. Dispatched to a petition
// sitting about what a required_fix may demand, it listed first among the things it could not find:
// "Red's closing or reasoning — Why did red mint this gap? … The gap's problem statement is clear,
// but red's intent for the fix isn't visible to me." It was being asked to adjudicate red's intent
// with red's intent unrecorded.
func TestAGapCarriesRedsArgumentAndNotOnlyItsProblem(t *testing.T) {
	runDir := newRun(t)
	t.Setenv("CLAUDE_PROJECT_DIR", recordtest.TmpRun(t))
	board := seatprobe.Boards()["audit"]
	exec := func(args ...string) (string, error) { return run(t, args...) }
	if err := seatprobe.Build(runtest.Open(t, runDir), board, exec); err != nil {
		t.Fatal(err)
	}

	const why = "blue may well have done the work; what is missing is the record of it"
	registerLensOnce(t, runDir)
	if _, err := runMint(t, "--run", runDir, "--seat-id", lensSeat,
		"--key", "argued", "--class", "spec_underspecification",
		"--quote", "The access log retains for 45 days.",
		"--problem", "the figure conflicts with the universal",
		"--fix", "reconcile them", "--check", "no two sections disagree", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", why); err != nil {
		t.Fatal(err)
	}

	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, why) {
		t.Errorf("a gap minted WITH --reason does not carry it to the board.\n\n"+
			"--reason's help promises \"the substance the report renders and the other side answers\"; if the verb "+
			"discards it whenever --problem is present, that promise is false for every ordinary mint and the write "+
			"still reports success.\n\nboard was:\n%s", out)
	}

	// --REASON IS NEVER THE PROBLEM. A mint that gives the argument and no --problem is refused, in
	// the problem's own words: the text it would have stored was red's argument, and filing it as
	// the problem — which the report prints — puts red's reasoning in the risk matrix.
	_, err = runMint(t, "--run", runDir, "--seat-id", lensSeat,
		"--key", "viareason", "--class", "spec_underspecification",
		"--quote", "The audit log retains for 30 days.",
		"--fix", "f", "--check", "c", "--check-kind", "document",
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low",
		"--reason", "the problem arrived through reason")
	if err == nil || !strings.Contains(err.Error(), "--problem") {
		t.Fatalf("a mint with --reason and no --problem = %v, want a refusal naming --problem", err)
	}
	out, err = run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "the problem arrived through reason") {
		t.Errorf("the refused mint reached the board:\n%s", out)
	}
}

// THE READING'S GAP IS FILLED AS A PLAIN MINT'S. One lens gives the same judgment flags and --reason
// to a plain mint, a citation reading and a found-source reading; the three stored Mints agree on
// every field but those the subject decides and fix_new's presence. Walked off the descriptor, so a
// field added to Mint fails here until it has a disposition.
func TestAReadingsGapIsFilledAsAPlainMints(t *testing.T) {
	bySubject := map[string]bool{"gap_id": true, "location": true}
	byPresenceOnly := map[string]bool{"fix_new": true}
	judgment := func(extra ...string) []string {
		return append([]string{"--class", "claim_contradicts_source", "--problem", "the source says otherwise than the sentence",
			"--check-kind", "source", "--check", "the sentence states what the source says",
			"--severity", "high", "--likelihood", "high", "--impact", "medium", "--reason", "the leaf says otherwise"}, extra...)
	}
	three := func(t *testing.T, b readingBoard, plain, cite, found []string) [3]*recordpb.Mint {
		t.Helper()
		out, err := run(t, append([]string{"mint", "--run", b.runDir, "--seat-id", lensL, "--quote", rdFree2}, plain...)...)
		if err != nil {
			t.Fatalf("plain mint: %v\n%s", err, out)
		}
		g0 := gapID(out)
		g1 := b.mustRead(lensL, append([]string{"--anchor", b.c1, "--as", "refutes", "--confidence", "high"}, cite...)...).Gap.ID
		g2 := b.mustRead(lensL, append([]string{"--url", rdU3, "--title", "T", "--quote", rdFree1, "--as", "refutes", "--confidence", "high"}, found...)...).Gap.ID
		return [3]*recordpb.Mint{mintOf(t, b.runDir, g0), mintOf(t, b.runDir, g1), mintOf(t, b.runDir, g2)}
	}
	agree := func(t *testing.T, ms [3]*recordpb.Mint) {
		t.Helper()
		fields := ms[0].ProtoReflect().Descriptor().Fields()
		for i := range fields.Len() {
			f := fields.Get(i)
			name := string(f.Name())
			if bySubject[name] {
				continue
			}
			for k, m := range ms[1:] {
				pr, mr := ms[0].ProtoReflect(), m.ProtoReflect()
				if pr.Has(f) != mr.Has(f) {
					t.Errorf("%s: the plain mint has it %v, reading %d has it %v", name, pr.Has(f), k+1, mr.Has(f))
					continue
				}
				if !byPresenceOnly[name] && pr.Has(f) && !pr.Get(f).Equal(mr.Get(f)) {
					t.Errorf("%s: the plain mint holds %v, reading %d holds %v", name, pr.Get(f), k+1, mr.Get(f))
				}
			}
		}
	}
	t.Run("no --new, --complexity or --fix", func(t *testing.T) {
		b := readingRun(t)
		ms := three(t, b, judgment(), judgment(), judgment())
		agree(t, ms)
		for _, m := range ms {
			if m.FixNew != nil || m.ComplexityCost != nil || m.GetRequiredFix() != "" || m.GetFixBasis() != "proposed" {
				t.Errorf("%s: fix_new %v complexity %v fix %q fix_basis %q", m.GetGapId(), m.FixNew, m.ComplexityCost, m.GetRequiredFix(), m.GetFixBasis())
			}
		}
	})
	t.Run("all three given", func(t *testing.T) {
		b := readingRun(t)
		more := func(nw string) []string {
			return judgment("--fix", "restate the figure", "--complexity", "low", "--new", nw)
		}
		ms := three(t, b,
			more("The rain fell on the plain in spring."),
			more("Costs rose modestly in the first quarter"+anchor.Token(b.c1)+"."),
			more("Volume recovered in the autumn."))
		agree(t, ms)
		for _, m := range ms {
			if m.FixNew == nil || m.GetFixBasis() != "verified" {
				t.Errorf("%s: fix_new %v fix_basis %q, want set and verified", m.GetGapId(), m.FixNew, m.GetFixBasis())
			}
		}
	})
	t.Run("GAP 3: what the verb tells the seat back", func(t *testing.T) {
		b := readingRun(t)
		const check = "run `grep -c cost` and $(echo the count) against the source"
		const loud = "the earlier draft of this report said otherwise"
		flags := func() []string {
			return []string{"--class", "claim_contradicts_source", "--problem", loud, "--check-kind", "source", "--check", check,
				"--severity", "high", "--likelihood", "high", "--impact", "medium", "--reason", "the leaf says otherwise"}
		}
		echo := "acceptance check RECORDED as: " + check
		out, err := run(t, append([]string{"mint", "--run", b.runDir, "--seat-id", lensL, "--quote", rdFree2}, flags()...)...)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, echo) || !strings.Contains(out, "(read it back —") {
			t.Fatalf("mint does not echo its check:\n%s", out)
		}
		mintNote := strings.Contains(out, "the risk matrix, the first sentence of the problem")
		for _, args := range [][]string{
			append([]string{"--anchor", b.c1, "--as", "refutes", "--confidence", "high"}, flags()...),
			append([]string{"--url", rdU3, "--title", "T", "--quote", rdFree1, "--as", "refutes", "--confidence", "high"}, flags()...),
		} {
			r := b.mustRead(lensL, args...)
			if m := mintOf(t, b.runDir, r.Gap.ID); r.Gap.Check != m.GetAcceptanceCheck() || r.Gap.Check != check {
				t.Errorf("--json acceptance_check %q, stored %q", r.Gap.Check, m.GetAcceptanceCheck())
			}
		}
		// The human text is read on a fresh board's first reading of each form.
		for _, form := range []func(b readingBoard) []string{
			func(b readingBoard) []string {
				return append([]string{"--anchor", b.c1, "--as", "refutes", "--confidence", "high"}, flags()...)
			},
			func(b readingBoard) []string {
				return append([]string{"--url", rdU3, "--title", "T", "--quote", rdFree1, "--as", "refutes", "--confidence", "high"}, flags()...)
			},
		} {
			b := readingRun(t)
			out, err := b.verify(lensL, form(b)...)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, echo) || !strings.Contains(out, "(read it back —") {
				t.Errorf("the reading does not echo its gap's check:\n%s", out)
			}
			if mintNote && !strings.Contains(out, "the risk matrix, the first sentence of the problem") {
				t.Errorf("the reading does not carry mint's voice note under its gap:\n%s", out)
			}
			if strings.Contains(out, "the source's note and Bibliography entry") {
				t.Errorf("the reading's gap is advised as a footnote:\n%s", out)
			}
			out, err = b.verify(lensM, form(b)...)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out, "acceptance check RECORDED") {
				t.Errorf("a reading naming the standing gap echoes a check it did not record:\n%s", out)
			}
		}
		if !mintNote {
			t.Fatalf("the problem %q trips no voice tell on mint, so the note row measures nothing", loud)
		}
	})
}

// THE READING'S GAP CARRIES ITS ARGUMENT WHEREVER A MINT'S DOES: the board, blue's work list, the
// docket's open-gap block and closure index — and never the report.
func TestAReadingsGapCarriesItsArgumentWhereverAMintsDoes(t *testing.T) {
	const why = "the leaf says the rise was modest"
	for _, form := range []struct {
		name    string
		args    func(b readingBoard) []string
		reRead  func(b readingBoard) []string
		reading string
	}{
		{"verify --anchor",
			func(b readingBoard) []string { return citeReading(b.c1, "refutes", "claim_contradicts_source") },
			func(b readingBoard) []string { return citeReading(b.c1, "supports", "") }, rdCited},
		{"verify --url",
			func(b readingBoard) []string { return urlReading(rdU3, rdFree1, "refutes", "claim_contradicts_source") },
			nil, rdFree1},
	} {
		t.Run(form.name, func(t *testing.T) {
			b := readingRun(t)
			args := replaced(form.args(b), []string{"--reason", why})
			g := b.mustRead(lensL, args...).Gap.ID
			bg := openGap(t, b.runDir, g)
			if bg.MintReason != why || bg.FixBasis != "proposed" {
				t.Errorf("the board's row: mint_reason %q fix_basis %q", bg.MintReason, bg.FixBasis)
			}
			w, err := record.WorkOfSeat(runtest.Open(t, b.runDir), "blue", blueSeat)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(w.Open)
			if !strings.Contains(string(raw), `"id":"`+g+`"`) || !strings.Contains(string(raw), `"fix_basis":"proposed"`) {
				t.Errorf("blue's work item for %s carries no fix_basis proposed:\n%s", g, raw)
			}
			if rep := readReport(t, b.runDir); !strings.Contains(rep, anchor.Token(g)) || strings.Contains(rep, why) {
				t.Errorf("the live report: token present %v, argument present %v", strings.Contains(rep, anchor.Token(g)), strings.Contains(rep, why))
			}
			md := assembled(t, b.runDir)
			docket := readFile(t, filepath.Join(b.runDir, "docket.md"))
			block := gapBlock(docket, g)
			if !strings.Contains(block, "minted because: "+why) || !strings.Contains(block, "fix_basis: **proposed**") {
				t.Errorf("docket.md's block for %s:\n%s", g, block)
			}
			if strings.Contains(md, why) {
				t.Errorf("report.md carries the argument")
			}
			if form.reRead == nil {
				return
			}
			// After L's own backing re-read and its close, the closure index keeps the argument.
			b.sit(lensL)
			b.mustRead(lensL, form.reRead(b)...)
			if _, err := run(t, "close", "--run", b.runDir, "--seat-id", lensL, "--id", g, "--as", "repaired",
				"--verified-by", "L1", "--verified-with", "a reading of the source", "--verified-against", form.reading,
				"--reason", "the source backs the sentence now"); err != nil {
				t.Fatalf("close: %v", err)
			}
			assembled(t, b.runDir)
			docket = readFile(t, filepath.Join(b.runDir, "docket.md"))
			_, index, _ := strings.Cut(docket, "losure")
			if !strings.Contains(index, g) || !strings.Contains(index, "minted because: "+why) {
				t.Errorf("docket.md's closure index lost the argument of %s:\n%s", g, index)
			}
		})
	}
	t.Run("a second lens's identical reading", func(t *testing.T) {
		b := readingRun(t)
		args := citeReading(b.c1, "refutes", "claim_contradicts_source")
		g := b.mustRead(lensL, replaced(args, []string{"--reason", why})...).Gap.ID
		const why2 = "a second reader's argument"
		if r := b.mustRead(lensM, replaced(args, []string{"--reason", why2})...); r.Gap == nil || r.Gap.ID != g {
			t.Fatalf("M's reading = %+v, want %s named", r.Gap, g)
		}
		if v, _ := b.lastVerify(); v.GetText() != why2 {
			t.Errorf("M's reading holds %q, want %q", v.GetText(), why2)
		}
		if m := mintOf(t, b.runDir, g); m.GetMintReason() != why {
			t.Errorf("G's mint_reason moved to %q", m.GetMintReason())
		}
	})
	t.Run("no --reason", func(t *testing.T) {
		b := readingRun(t)
		args := []string{"--anchor", b.c1, "--as", "refutes", "--confidence", "high"}
		_, err := b.verify(lensL, append(args, gapFlagArgs("claim_contradicts_source")...)...)
		mustRefuse(t, err, "", "--reason")
		if c := b.counts(); c != [3]int{} {
			t.Errorf("the refusal wrote verify/mint/anchor %v", c)
		}
	})
}

// gapBlock is docket.md's text from g's first mention to the next blank-line-separated entry.
func gapBlock(docket, g string) string {
	i := strings.Index(docket, g)
	if i < 0 {
		return ""
	}
	rest := docket[i:]
	if j := strings.Index(rest, "\n\n"); j >= 0 {
		return rest[:j]
	}
	return rest
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
