package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE DEMAND IS ONE TEXT, AND EVERY NAMED CARRIER RENDERS IT (V1a′.8; ruling Close test). Each
// carrier is rendered the way the shipped path renders it, and the test holds it to
// recordpb.DemandClause verbatim, so a carrier that restates the demand in its own words fails here,
// named with its shape and check kind.
//
// ITS LIMIT: it reads these carriers only. Every other sentence that says when a gap closes — the
// other help pages, failedClears and the kind's refusal, blueAnswers, answersRetiredItems, the
// seatprobe expectations, the --answers flag texts, the Go comments, the agentgen sources, both
// preloaded skills, debate.js's authored text, FEOV's docs and the `failed check` term — is read by
// the census sweeps, by hand.
func TestEveryDemandCarrierRendersTheDemandClause(t *testing.T) {
	holds := func(t *testing.T, carrier, text string) {
		t.Helper()
		if !strings.Contains(text, recordpb.DemandClause) {
			t.Errorf("%s does not render the demand clause verbatim:\n%s", carrier, text)
		}
	}
	refusal := func(t *testing.T, carrier string, err error) string {
		t.Helper()
		if err == nil {
			t.Fatalf("%s: the close was admitted", carrier)
		}
		return err.Error()
	}
	// answer places an anchor of the gap's demanded kind on rdFree1 as g's answer, so the demand is
	// met and the close reaches the record's failed-check refusal.
	answer := func(t *testing.T, runDir, check, g string) {
		t.Helper()
		if check == "source" {
			placeAnswer(t, runDir, "citation", rdFree1, rdU7, g)
			return
		}
		placeAnswer(t, runDir, "computation", rdFree1, "", g)
	}
	// readingGap is lensL's refuting reading of the board's first citation, minting its gap under check.
	readingGap := func(t *testing.T, b readingBoard, check string) string {
		t.Helper()
		args := citeReading(b.c1, "refutes", "claim_contradicts_source")
		for i, a := range args {
			if a == "--check-kind" {
				args[i+1] = check
			}
		}
		r := b.mustRead(lensL, args...)
		if r.Gap == nil || r.Gap.ID == "" {
			t.Fatalf("the refuting reading minted no gap: %+v", r)
		}
		return r.Gap.ID
	}
	// unsoundComputation is a computation on rdFree2 that lensL judges unsound.
	unsoundComputation := func(t *testing.T, b readingBoard) string {
		t.Helper()
		k := placeAnswer(t, b.runDir, "computation", rdFree2, "", "")
		if _, err := run(t, "reproduce", "--run", b.runDir, "--seat-id", lensL, "--anchor", k, "--as", "unsound", "--reason", "it computes another claim"); err != nil {
			t.Fatalf("reproduce unsound: %v", err)
		}
		return k
	}

	for _, check := range []string{"source", "computation"} {
		t.Run("requireEvidenceAnswered through close, plain gap, "+check, func(t *testing.T) {
			b := readingRun(t)
			g := demandGap(t, b.runDir, "plain", check)
			holds(t, "the close refusal of a plain "+check+" gap", refusal(t, "plain "+check, lensClose(t, b.runDir, lensL, g, "repaired")))
		})
		t.Run("requireEvidenceAnswered through close, gap about an anchor, "+check, func(t *testing.T) {
			b := readingRun(t)
			g := demandGap(t, b.runDir, "about", check, "--about-kind", "anchor", "--about", b.c1)
			holds(t, "the close refusal of a "+check+" gap about an anchor", refusal(t, "about "+check, lensClose(t, b.runDir, lensL, g, "repaired")))
		})
		t.Run("requireEvidenceAnswered through close, gap carrying a reading, "+check, func(t *testing.T) {
			b := readingRun(t)
			g := readingGap(t, b, check)
			holds(t, "the close refusal of a "+check+" gap carrying a reading", refusal(t, "reading "+check, lensClose(t, b.runDir, lensL, g, "repaired")))
		})

		t.Run("requireFailedCheckAnswered through close, unread, "+check, func(t *testing.T) {
			b := readingRun(t)
			b.mustRead(lensL, citeReading(b.c1, "unreachable", "")...)
			g := demandGap(t, b.runDir, "unread", check, "--about-kind", "anchor", "--about", b.c1)
			answer(t, b.runDir, check, g)
			msg := refusal(t, "unread "+check, lensClose(t, b.runDir, lensL, g, "repaired"))
			if !strings.Contains(msg, "whose source no check since its sentence last changed has read") {
				t.Fatalf("the close was refused, but not by the unread check:\n%s", msg)
			}
			holds(t, "the unread failed-check close refusal of a "+check+" gap", msg)
		})
		t.Run("requireFailedCheckAnswered through close, unsound, "+check, func(t *testing.T) {
			b := readingRun(t)
			k := unsoundComputation(t, b)
			g := demandGap(t, b.runDir, "unsound", check, "--about-kind", "anchor", "--about", k)
			answer(t, b.runDir, check, g)
			msg := refusal(t, "unsound "+check, lensClose(t, b.runDir, lensL, g, "repaired"))
			if !strings.Contains(msg, "still calls unsound") {
				t.Fatalf("the close was refused, but not by the unsound check:\n%s", msg)
			}
			holds(t, "the unsound failed-check close refusal of a "+check+" gap", msg)
		})
		t.Run("requireFailedCheckAnswered through close, found_against, "+check, func(t *testing.T) {
			b := readingRun(t)
			g := readingGap(t, b, check)
			answer(t, b.runDir, check, g)
			msg := refusal(t, "found_against "+check, lensClose(t, b.runDir, lensL, g, "repaired"))
			if !strings.Contains(msg, "finds the source against") {
				t.Fatalf("the close was refused, but not by the found-against check:\n%s", msg)
			}
			holds(t, "the found-against failed-check close refusal of a "+check+" gap", msg)
		})

		t.Run("blue's sitting debt line through show work, "+check, func(t *testing.T) {
			b := readingRun(t)
			g := demandGap(t, b.runDir, "owed", check)
			out, err := run(t, "show", "work", "--run", b.runDir, "--seat-id", blueSeat)
			if err != nil {
				t.Fatalf("blue's show work: %v", err)
			}
			var line string
			for _, l := range strings.Split(out, "\n") {
				if strings.Contains(l, "gap "+g+" was minted --check-kind "+check) {
					line = l
				}
			}
			if line == "" {
				t.Fatalf("blue's work list holds no debt line for %s gap %s:\n%s", check, g, out)
			}
			holds(t, "blue's sitting debt line on a "+check+" gap", line)
		})
	}

	t.Run("carry's refusal of a gap no close has closed", func(t *testing.T) {
		runDir := citedRun(t)
		g := demandGap(t, runDir, "open", "source")
		must(t, runDir, "register", "--seat-id", "red-chair")
		_, err := run(t, "carry", "--run", runDir, "--seat-id", "red-chair", "--id", g, "--carried-from", "1", "--as", "repaired")
		holds(t, "carry's refusal", refusal(t, "carry", err))
	})

	t.Run("close --help", func(t *testing.T) {
		out, err := run(t, "close", "--seat-id", lensSeat, "--help")
		if err != nil {
			t.Fatal(err)
		}
		holds(t, "close --help", out)
	})

	// THE TWO (means), rendered: in mint --help, and in every generated lens definition that inlines
	// the mint page. Each value's own line must carry the clause.
	means := func(t *testing.T, carrier, text string) {
		t.Helper()
		for _, v := range []string{"source", "computation"} {
			var line string
			for _, l := range strings.Split(text, "\n") {
				if f := strings.Fields(l); len(f) > 1 && f[0] == v && strings.Contains(l, "settles it") {
					line = l
				}
			}
			if line == "" {
				t.Errorf("%s holds no `%s` check-kind line", carrier, v)
				continue
			}
			holds(t, carrier+"'s `"+v+"` line", line)
		}
	}
	t.Run("mint --help, from the check kinds' (means)", func(t *testing.T) {
		out, err := run(t, "mint", "--seat-id", lensSeat, "--help")
		if err != nil {
			t.Fatal(err)
		}
		means(t, "mint --help", out)
	})
	t.Run("the generated red-lens definitions", func(t *testing.T) {
		dir, err := repotree.Plugin("agents")
		if err != nil {
			t.Fatal(err)
		}
		paths, err := filepath.Glob(filepath.Join(dir, "red-lens-*.md"))
		if err != nil || len(paths) == 0 {
			t.Fatalf("no red-lens definitions under %s (%v)", dir, err)
		}
		for _, p := range paths {
			body, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			means(t, filepath.Base(p), string(body))
		}
	})
}

// THE ANCHOR LIFECYCLE AND THE DEMAND, AS EACH PAGE SAYS THEM (V1a′.8). A failing assertion over the
// rendered pages, not a golden a person reads: drop one of these sentences and the page that lost
// it is named. Whitespace is folded, so a generated definition's line wrapping does not matter.
func TestTheAnchorLifecyclePagesSayIt(t *testing.T) {
	fold := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	page := func(seat string, path ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			t.Helper()
			out, err := run(t, append(append(append([]string{}, path...), "--seat-id", seat), "--help")...)
			if err != nil {
				t.Fatalf("%s --help: %v", strings.Join(path, " "), err)
			}
			return out
		}
	}
	generated := func(path ...string) func(t *testing.T) string {
		return func(t *testing.T) string {
			t.Helper()
			p, err := repotree.Plugin(path...)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			return string(b)
		}
	}
	label := func(kind string) func(t *testing.T) string {
		return func(t *testing.T) string { return anchor.Label(anchor.ID(kind, [4]byte{1, 2, 3, 4})) }
	}
	for _, c := range []struct {
		carrier string
		render  func(t *testing.T) string
		says    []string
	}{
		{"retire --help", page(blueSeat, "retire"), []string{
			"EVIDENCE LEAVES ON ITS OWN, ITS SENTENCE KEPT. A retire naming one citation or computation and no claim takes that anchor out of the report and keeps the sentence it stood on word for word",
			"A retired anchor is in no document",
			"A CITATION A COMPUTATION APPLIES STAYS WHILE THE COMPUTATION DOES. A retire taking a citation that a computation still in the report names as its method is refused, naming each such computation",
			"Retiring an anchor placed as an open gap's answer takes that answer away; with no other answer standing, the gap is owed again.",
			"Retiring the answers of a gap already closed leaves its closure standing, and the gap's lens is told.",
		}},
		{"edit --help", page(blueSeat, "edit"), []string{
			"EVIDENCE NEVER STANDS BARE. A replacement that would leave a citation's or computation's anchor with no prose before it in its sentence is refused, whether your quote holds it or not",
		}},
		{"reproduce --help", page(lensSeat, "reproduce"), []string{
			"A COMPUTATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED, and nothing is recorded",
		}},
		{"close --help", page(lensSeat, "close"), []string{
			"under every disposition — `moot`, `not_a_defect` and `repaired_with_regression` included",
			"a `source` gap's fix that rewords the sentence leaves it unanswered until blue cites on the reworded sentence as that gap's answer",
			"A demand you no longer hold is not closed: `motion docket file` on the gap puts it before the bench, whose closing ruling is not held to it.",
		}},
		{"mint --help", page(lensSeat, "mint"), []string{
			"if the defect it closed is still in the report, that is a new gap — mint it naming that gap as the one it supersedes",
			"A gap about a retired citation or computation is refused, naming its retire",
		}},
		{"cite --help", page(blueSeat, "cite"), []string{
			"Name the gap it answers, and the citation is placed as that gap's answer",
			"cite the source on the reworded sentence as that gap's answer",
			"A citation leaves by a `retire` naming it alone, its sentence kept",
		}},
		{"compute --help", page(blueSeat, "compute"), []string{
			"A computation leaves by a `retire` naming it alone, its sentence kept",
		}},
		{"show changes --help", page(blueSeat, "show", "changes"), []string{
			"each edit, and each citation or computation blue retired from a sentence the report keeps, named by its kind",
		}},
		{"show evidence --help", page(lensSeat, "show", "evidence"), []string{
			"A retired anchor is in no document, and in none of these lists or counts",
			"Read at its id with --anchor, a retired one is still its row, with `retired`",
		}},
		{"show board --help", page(lensSeat, "show", "board"), []string{
			"`answers_retired_event` is, on a closed gap, the event of the retire that left none of its answers live, null on every other gap: the closure stands, and its lens is told",
		}},
		{"the citation anchor's lifecycle label", label("citation"), []string{"it leaves by `retire --anchor` alone, which keeps its sentence word for word"}},
		{"the computation anchor's lifecycle label", label("computation"), []string{"it leaves by `retire --anchor` alone, which keeps its sentence word for word"}},
		{"the generated blue-researcher definition", generated("agents", "blue-researcher.md"), []string{
			"`source` or `computation` closes only on a live answer of its kind — a citation or a computation placed as that gap's answer and still in the report — never on prose",
			"`awaiting_evidence: true` with its `answer_kind` means that gap is waiting on evidence from YOU, a source (`citation`) or a program (`computation`)",
			"discharge each with a citation or computation placed as that gap's answer",
		}},
		{"the generated research-protocol skill", generated("skills", "research-protocol", "SKILL.md"), []string{
			"each document numbers the references it carries, citations and computations alike, in order of first appearance, so one computation can carry a different number in another document",
			"Name a computation by its id (`K-…`) and a citation by its id (`C-…`), never by a footnote number",
		}},
	} {
		t.Run(c.carrier, func(t *testing.T) {
			text := fold(c.render(t))
			for _, s := range c.says {
				if !strings.Contains(text, fold(s)) {
					t.Errorf("%s does not say %q", c.carrier, s)
				}
			}
		})
	}
}
