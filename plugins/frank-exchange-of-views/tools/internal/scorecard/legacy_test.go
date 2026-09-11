package scorecard

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// preRenameJournal is a real journal captured before the disposition keys: its judge envelopes
// carry `resolutions` and its blue envelopes `grade_disputes`. Read by repository path, skipped
// LOUDLY where the checkout does not hold it, never silently green.
func preRenameJournal(t *testing.T) []byte {
	t.Helper()
	const rel = "research/2026-08-23_sleeper-service-plan/trajectories/journal.jsonl"
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if b, err := os.ReadFile(filepath.Join(dir, rel)); err == nil {
			return b
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("%s not found above this checkout — the legacy-key check needs the repository's captured journal", rel)
		}
		dir = parent
	}
}

func resultsOf(body []byte) []map[string]any {
	var out []map[string]any
	for _, obj := range decodeJSONL(body) {
		if r, ok := obj["result"].(map[string]any); ok {
			out = append(out, r)
		}
	}
	return out
}

func TestAPreRenameJournalIsNamedByItsKeys(t *testing.T) {
	results := resultsOf(preRenameJournal(t))
	got := LegacyKeys(results)
	if !slices.Equal(got, []string{"resolutions", "grade_disputes"}) {
		t.Fatalf("LegacyKeys = %v, want both pre-rename keys this journal carries", got)
	}
	if k := LegacyKeys([]map[string]any{{"dispositions": []any{}, "grade_motions": []any{}}}); len(k) != 0 {
		t.Errorf("a current journal reported legacy keys %v", k)
	}
}

// THE OLD JOURNAL'S BENCH DID SIT. Read under the current keys its rulings are absent, and the
// row would say "the bench did not sit this run" — true of no run that holds four judge envelopes.
func TestTheBenchRowOnAPreRenameJournalIsStatedNotZero(t *testing.T) {
	results := resultsOf(preRenameJournal(t))
	r := rowByMetric(benchRows(results, nil), "remanded_share")
	if r == nil {
		t.Fatal("no remanded_share row")
	}
	if r.Value != nil {
		t.Errorf("remanded_share has a value %v on a journal whose rulings it could not read", r.Value)
	}
	want := "not measured: this transcript's envelopes predate the disposition keys (resolutions/grade_disputes)"
	if r.Note != want {
		t.Errorf("remanded_share note = %q, want %q", r.Note, want)
	}
}
