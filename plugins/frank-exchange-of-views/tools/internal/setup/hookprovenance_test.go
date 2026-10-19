package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookprovenance"
)

// THE FIELD REACHES run-config.json AND COMES BACK OUT, which is a different question from whether
// the observer works. hookprovenance's tests drive Observe directly; this one marshals the struct
// the setup path actually writes and reads it with the reader run.md uses, so a field added to the
// observer and never wired into runConfig — or written under a key the reader does not ask for —
// fails here rather than shipping as a fix nobody gets.
//
// It asserts the SHAPE reaches the file, not what a particular machine's cache holds: the value
// depends on the developer's install state, and a test that asserted a version would pass or fail
// on where it ran.
func TestRunConfigCarriesHookProvenance(t *testing.T) {
	rc := runConfig{Topic: "t", Hooks: hookprovenance.Observe(t.TempDir(), "frank-exchange-of-views")}
	b, err := json.Marshal(rc)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inputs", "run-config.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	back, ok, err := hookprovenance.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("run-config.json carries no provenance under the key the reader asks for — the provenance is " +
			"observed and then dropped, which leaves the record exactly as silent about the hook side as before")
	}
	if back.State == "" {
		t.Error("the hook provenance round-trips with an empty state, so a reader cannot tell " +
			"'nothing was installed' from 'this binary did not look'")
	}
}
