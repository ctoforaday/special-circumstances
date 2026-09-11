package dashboard

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// preRenameJournalEnvs reads a real journal captured before the disposition keys, by repository
// path, skipping LOUDLY where the checkout does not hold it.
func preRenameJournalEnvs(t *testing.T) []map[string]any {
	t.Helper()
	const rel = "research/2026-08-23_sleeper-service-plan/trajectories/journal.jsonl"
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var body []byte
	for {
		if b, err := os.ReadFile(filepath.Join(dir, rel)); err == nil {
			body = b
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("%s not found above this checkout — the legacy-key check needs the repository's captured journal", rel)
		}
		dir = parent
	}
	var envs []map[string]any
	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for sc.Scan() {
		var m map[string]any
		if json.Unmarshal(sc.Bytes(), &m) == nil {
			envs = append(envs, m)
		}
	}
	return envs
}

// An old journal's bench sat four times; read under the current keys it would render "no judge
// sittings yet" and zero rulings, which is the page for a bench that never convened. The
// Judiciary table says what it found instead.
func TestAPreRenameJournalRendersNotMeasured(t *testing.T) {
	j := buildJudiciary(preRenameJournalEnvs(t))
	if len(j.Legacy) == 0 {
		t.Fatal("buildJudiciary did not notice the journal's pre-rename keys")
	}
	m := baseModel(t, fixtureRunDir(t))
	m.Judiciary = j
	h := RenderHTML(m)
	if strings.Contains(h, "no judge sittings yet") {
		t.Error("the page says no judge sat, over a journal holding four judge envelopes")
	}
	want := "not measured: this transcript's envelopes predate the disposition keys (resolutions/grade_disputes)"
	if !strings.Contains(h, want) {
		t.Errorf("the Judiciary table does not state what it could not read (want %q)", want)
	}
}
