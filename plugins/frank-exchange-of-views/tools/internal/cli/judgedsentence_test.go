package cli

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// judgedKind is one backing kind's script for the judged-sentence tests: a run whose report backs a
// sentence with one anchor of the kind, and the kind's check of it by its seat.
type judgedKind struct {
	setup func(t *testing.T) (runDir, id string)
	check func(t *testing.T, runDir, id string, extra ...string) (string, error)
	// seat makes the check, and by is how evidence.md names its act.
	seat, by string
	// quote is a span of the checked sentence a mint may place a gap at; edit rewrites the sentence.
	quote string
	edit  [2]string
}

// judgedKinds is every backing kind's script, keyed by its EvidenceTables row's kind.
func judgedKinds() map[string]judgedKind {
	return map[string]judgedKind{
		"citation": {
			setup: citedClause,
			check: func(t *testing.T, runDir, id string, extra ...string) (string, error) {
				return run(t, append([]string{"verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", id,
					"--as", "supports", "--confidence", "high", "--reason", "read at the leaf"}, extra...)...)
			},
			seat: lensSeat, by: "read by",
			quote: "the grass is green",
			edit:  [2]string{"the grass is green", "the grass is wet"},
		},
		"computation": {
			setup: func(t *testing.T) (string, string) {
				runDir := newRun(t)
				writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Plain prose here.\n")
				registerBlue(t, runDir)
				return runDir, computeSentence(t, runDir, "Alpha holds.")
			},
			check: func(t *testing.T, runDir, id string, extra ...string) (string, error) {
				return run(t, append([]string{"reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
					"--as", "sound", "--reason", "it computes what the sentence says"}, extra...)...)
			},
			seat: "red-lens-computation", by: "judged by",
			quote: "Alpha holds",
			edit:  [2]string{"Alpha", "Alpha still"},
		},
	}
}

// EVERY CHECK RECORDS THE SENTENCE IT JUDGED (ruling L5), one script over every backing kind's row: a
// check of a live anchor records, in its check table's `claim`, the sentence record.LiveSentence reads
// at the write; the seat has no flag to type it; the result carries it and the human line quotes it.
func TestEveryCheckRecordsTheSentenceItJudged(t *testing.T) {
	kinds := judgedKinds()
	for _, row := range recordsql.EvidenceTables {
		k, ok := kinds[row.Kind]
		if !ok {
			t.Errorf("backing kind %s has no arm here: its check's recorded sentence is unasserted", row.Kind)
			continue
		}
		t.Run(row.Kind, func(t *testing.T) {
			runDir, id := k.setup(t)
			want, err := record.LiveSentence(runtest.Open(t, runDir), id)
			if err != nil || want == "" {
				t.Fatalf("%s %s has no live sentence to judge: %q, %v", row.Noun, id, want, err)
			}
			out, err := k.check(t, runDir, id)
			if err != nil {
				t.Fatalf("a %s check of a live anchor was refused: %v\n%s", row.CheckTable, err, out)
			}
			var got string
			if err := openRunDB(t, runDir).QueryRow(`SELECT coalesce("claim", '') FROM "`+row.CheckTable+`" WHERE "`+row.CheckIDColumn+`" = ?`, id).Scan(&got); err != nil {
				t.Fatalf("read %s's claim: %v", row.CheckTable, err)
			}
			if got != want {
				t.Errorf("%s recorded claim %q, want the anchor's live sentence %q", row.CheckTable, got, want)
			}
			if !printsClaim(out, want) {
				t.Errorf("%s's human line does not quote the sentence it judged, %q:\n%s", row.CheckTable, want, out)
			}

			// The seat has no flag for the sentence: the tool reads it, and a typed one is refused.
			runDir, id = k.setup(t)
			if _, err := k.check(t, runDir, id, "--claim", "a sentence the seat typed"); err == nil || !strings.Contains(err.Error(), "--claim") {
				t.Errorf("%s took a --claim flag: %v", row.CheckTable, err)
			}

			out, err = k.check(t, runDir, id, "--json")
			if err != nil {
				t.Fatalf("%s --json: %v\n%s", row.CheckTable, err, out)
			}
			var env struct {
				Result struct {
					Claim string `json:"claim"`
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(out), &env); err != nil {
				t.Fatalf("%s --json: %v\n%s", row.CheckTable, err, out)
			}
			if env.Result.Claim != want {
				t.Errorf("%s's result carries claim %q, want %q", row.CheckTable, env.Result.Claim, want)
			}
		})
	}
}

// The reproduce page teaches that the sentence is recorded with the judgement, and that an edit of it
// leaves the judgement on an earlier wording, owed again (V1a″.3): asserted on the rendered page.
func TestTheReproducePageSaysTheJudgedSentenceIsRecorded(t *testing.T) {
	out, err := run(t, "reproduce", "--seat-id", "red-lens-computation", "--help")
	if err != nil {
		t.Fatal(err)
	}
	page := strings.Join(strings.Fields(out), " ")
	for _, want := range []string{
		"records it with your judgement",
		"leaves your judgement on an earlier wording",
		"the sentence is owed a judgement again",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("reproduce --help does not say %q:\n%s", want, out)
		}
	}
}

// EVERY CHECK RENDERS WITH THE SENTENCE IT JUDGED (ruling L5, Staleness), one script over every backing
// kind's row. evidence.md's entry prints the sentence the anchor backs and each check through one
// list (checkLines), and `show evidence` gives each check's `claim`. A gap placed at the end of the
// checked sentence after the check is no change of wording: no sentence line. An edit that changes
// the sentence names the anchor reopened, the check no longer stands, and evidence.md prints the
// sentence it was of under it; a fresh check from a sitting opened after the edit prints none.
func TestEveryCheckRendersWithTheSentenceItJudged(t *testing.T) {
	kinds := judgedKinds()
	for _, row := range recordsql.EvidenceTables {
		k, ok := kinds[row.Kind]
		if !ok {
			t.Errorf("backing kind %s has no arm here: its checks' rendered sentence is unasserted", row.Kind)
			continue
		}
		t.Run(row.Kind, func(t *testing.T) {
			runDir, id := k.setup(t)
			sit(t, runDir, k.seat)
			old, err := record.LiveSentence(runtest.Open(t, runDir), id)
			if err != nil {
				t.Fatal(err)
			}
			if out, err := k.check(t, runDir, id); err != nil {
				t.Fatalf("%s: %v\n%s", row.CheckTable, err, out)
			}
			checkLine := "- **" + k.by + " `" + k.seat + "`**"
			ev := assembledSet(t, runDir)[report.FileEvidence]
			if !strings.Contains(ev, "- **backs**: "+old+"\n") {
				t.Errorf("the %s's evidence.md entry does not print the sentence it backs, %q:\n%s", row.Noun, old, ev)
			}
			if !strings.Contains(ev, checkLine) {
				t.Errorf("the %s's check is not listed as %q in evidence.md:\n%s", row.Noun, checkLine, ev)
			}
			if strings.Contains(ev, "of the sentence") {
				t.Errorf("a check of the sentence the report carries prints a sentence line:\n%s", ev)
			}
			if got := shownClaims(t, runDir, id); len(got) != 1 || got[0] != old {
				t.Errorf("show evidence gives the %s's checks claims %q, want [%q]", row.Noun, got, old)
			}

			// A gap's token at the end of the checked sentence, placed after the check: the wording is
			// the same, so no sentence line.
			if out, err := run(t, "mint", "--run", runDir, "--seat-id", k.seat, "--quote", k.quote, "--class", "spec_underspecification",
				"--problem", "the defect", "--fix", "the fix", "--check-kind", "document", "--check", "the acceptance check",
				"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
				t.Fatalf("mint at the checked sentence: %v\n%s", err, out)
			}
			if !strings.Contains(readReport(t, runDir), "<!--gap:") {
				t.Fatalf("the mint placed no gap token in the report:\n%s", readReport(t, runDir))
			}
			if ev := assembledSet(t, runDir)[report.FileEvidence]; strings.Contains(ev, "of the sentence") {
				t.Errorf("a gap token placed in the checked sentence reads as a change of wording:\n%s", ev)
			}

			// The edit that changes the sentence.
			if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", k.edit[0], "--new", k.edit[1],
				"--reason", "rewrote the checked sentence"); err != nil {
				t.Fatalf("edit: %v\n%s", err, out)
			}
			if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetReopened(); !slices.Contains(got, id) {
				t.Fatalf("the edit of the sentence names %v reopened, not %s", got, id)
			}
			if !uncheckedNow(t, runDir, id) {
				t.Errorf("the %s's check still stands after the edit of its sentence", row.Noun)
			}
			ev = assembledSet(t, runDir)[report.FileEvidence]
			if !strings.Contains(ev, "  - of the sentence: "+old+"\n") {
				t.Errorf("evidence.md does not print the sentence the stale check was of, %q:\n%s", old, ev)
			}

			// A fresh check, from a sitting opened after the edit.
			sit(t, runDir, k.seat)
			if out, err := k.check(t, runDir, id); err != nil {
				t.Fatalf("the fresh %s: %v\n%s", row.CheckTable, err, out)
			}
			if uncheckedNow(t, runDir, id) {
				t.Errorf("the fresh check of the %s does not stand", row.Noun)
			}
			ev = assembledSet(t, runDir)[report.FileEvidence]
			lines := strings.Split(ev, "\n")
			last := -1
			for i, ln := range lines {
				if strings.HasPrefix(ln, checkLine) {
					last = i
				}
			}
			switch {
			case last < 0 || strings.Count(ev, checkLine) != 2:
				t.Errorf("evidence.md does not list both checks as %q:\n%s", checkLine, ev)
			case last+1 < len(lines) && strings.HasPrefix(lines[last+1], "  - of the"):
				t.Errorf("the fresh check prints a sentence line:\n%s", ev)
			case strings.Count(ev, "of the sentence: ") != 1:
				t.Errorf("evidence.md prints a sentence line for other than the stale check:\n%s", ev)
			}
		})
	}
}

// shownClaims is the `claim` of each check `show evidence --json` gives of evidence anchor id.
func shownClaims(t *testing.T, runDir, id string) []string {
	t.Helper()
	out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatalf("show evidence --json: %v", err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal([]byte(out), &ev); err != nil {
		t.Fatalf("show evidence --json: %v\n%s", err, out)
	}
	var claims []string
	for _, s := range ev.Sources {
		if s.Anchor == id {
			for _, v := range s.Verified {
				claims = append(claims, v.Claim)
			}
		}
	}
	for _, c := range ev.Computations {
		if c.Anchor == id {
			for _, j := range c.Checks {
				claims = append(claims, j.Claim)
			}
		}
	}
	return claims
}

// uncheckedNow is whether evidence anchor id has no standing check of the sentence the report carries.
func uncheckedNow(t *testing.T, runDir, id string) bool {
	t.Helper()
	fam, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	return slices.ContainsFunc(record.EvidenceJSONOf(fam.Events, fam.At).Unchecked, func(u record.EvidenceUncheckedJSON) bool { return u.Anchor == id })
}

// `show evidence --help` teaches that each judgement carries the sentence it was of, and the
// computation lens's definition that its judgement records it and one of an earlier wording no longer
// stands (V1a″.3): asserted on the rendered page and the generated definition.
func TestTheJudgedSentenceIsTaughtWhereTheJudgementIsRead(t *testing.T) {
	out, err := run(t, "show", "evidence", "--seat-id", "red-lens-computation", "--help")
	if err != nil {
		t.Fatal(err)
	}
	if page := strings.Join(strings.Fields(out), " "); !strings.Contains(page, "each judgement carries `claim`, the sentence it was of") {
		t.Errorf("show evidence --help does not say each judgement carries its sentence:\n%s", out)
	}
	path, err := repotree.Plugin("agents", "red-lens-computation.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	def := strings.Join(strings.Fields(string(b)), " ")
	for _, want := range []string{
		"YOUR JUDGEMENT RECORDS THE SENTENCE IT WAS OF",
		"shows that sentence beside each judgement",
		"a judgement of an earlier wording no longer stands",
	} {
		if !strings.Contains(def, want) {
			t.Errorf("the red-lens-computation definition does not say %q", want)
		}
	}
}
