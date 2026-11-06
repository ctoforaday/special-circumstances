package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY CHECK RECORDS THE SENTENCE IT JUDGED (ruling L5), one script over every backing kind's row: a
// check of a live anchor records, in its check table's `claim`, the sentence record.LiveSentence reads
// at the write; the seat has no flag to type it; the result carries it and the human line quotes it.
func TestEveryCheckRecordsTheSentenceItJudged(t *testing.T) {
	const computed = "Alpha holds."
	type kind struct {
		setup func(t *testing.T) (runDir, id string)
		check func(t *testing.T, runDir, id string, extra ...string) (string, error)
	}
	kinds := map[string]kind{
		"citation": {
			setup: citedClause,
			check: func(t *testing.T, runDir, id string, extra ...string) (string, error) {
				return run(t, append([]string{"verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", id,
					"--as", "supports", "--confidence", "high", "--reason", "read at the leaf"}, extra...)...)
			},
		},
		"computation": {
			setup: func(t *testing.T) (string, string) {
				runDir := newRun(t)
				writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Plain prose here.\n")
				registerBlue(t, runDir)
				return runDir, computeSentence(t, runDir, computed)
			},
			check: func(t *testing.T, runDir, id string, extra ...string) (string, error) {
				return run(t, append([]string{"reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
					"--as", "sound", "--reason", "it computes what the sentence says"}, extra...)...)
			},
		},
	}
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
