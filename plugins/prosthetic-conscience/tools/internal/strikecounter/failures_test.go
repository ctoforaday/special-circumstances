package strikecounter

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikes"
)

func strikeStages(t *testing.T, scope string) map[hookfailures.Stage]bool {
	t.Helper()
	path, err := hookfailures.Path("prosthetic-conscience")
	if err != nil {
		t.Fatal(err)
	}
	out := map[hookfailures.Stage]bool{}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return out
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Failures {
		if f.Scope == scope {
			out[f.Stage] = true
		}
	}
	return out
}

// A STRIKES FILE THAT CANNOT BE READ OR WRITTEN DISABLES THE 3-STRIKE RULE for the project, and
// strikes.Load returned an empty state with no error at all — so it did so permanently and in
// silence. Counting still degrades to zero (a tool call is never the cost); the reason is recorded.
func TestAStrikesFileThatCannotBeKeptIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	path := strikes.Path(dir)
	if err := os.MkdirAll(path, 0o755); err != nil { // the strikes FILE as a directory: neither read nor write works
		t.Fatal(err)
	}
	if _, _, code := fire(t, dir, bash("make"), noon); code != 0 {
		t.Fatalf("a strikes file that cannot be kept must never cost the call: exit %d", code)
	}
	got := strikeStages(t, dir)
	if !got[StageStrikesRead] || !got[StageStrikesWrite] {
		t.Fatalf("read and write failures were not both recorded: %v", got)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	fire(t, dir, bash("make"), noon)
	if got := strikeStages(t, dir); got[StageStrikesRead] || got[StageStrikesWrite] {
		t.Errorf("a strikes file that works left entries: %v", got)
	}
}
