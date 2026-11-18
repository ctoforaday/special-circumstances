package migrate_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// AN ARCHIVED RUN'S `hooks` KEY IS DROPPED, AND THE MANIFEST KEEPS WHAT IT SAID. Setup's observation of
// the plugin install cache is no key of the run's configuration, which the record reads strictly,
// so a copy carrying it would refuse every read of the migrated run's terms. The 11-11 run is the
// one whose observation named every hook binary missing while every seat's register carries a hook
// build: the source is checked to state the key first, so the test cannot pass on an archive that
// never had one.
func TestAnArchivedRunsHookObservationIsDroppedAndRecorded(t *testing.T) {
	const name = "2026-11-11_eiffel-tower-height.tar.gz"
	src := t.TempDir()
	untar(t, archivePath(t, name), src)
	var source map[string]json.RawMessage
	readJSON(t, filepath.Join(src, "inputs", "run-config.json"), &source)
	want, ok := source["hooks"]
	if !ok {
		t.Fatalf("%s states no `hooks` in its run-config, so nothing here is dropped", name)
	}

	res, run := migrateArchive(t, name)
	var migrated map[string]json.RawMessage
	readJSON(t, filepath.Join(run.Dir(), "inputs", "run-config.json"), &migrated)
	if v, ok := migrated["hooks"]; ok {
		t.Errorf("the migrated run-config still states `hooks`: %s", v)
	}
	if res.RunConfig == nil {
		t.Fatal("the manifest records no run-config translation")
	}
	got, ok := res.RunConfig.Dropped["hooks"]
	if !ok {
		t.Fatalf("the manifest's dropped keys %v do not name `hooks`", res.RunConfig.Dropped)
	}
	var g, w bytes.Buffer
	if err := json.Compact(&g, got); err != nil {
		t.Fatal(err)
	}
	if err := json.Compact(&w, want); err != nil {
		t.Fatal(err)
	}
	if g.String() != w.String() {
		t.Errorf("the manifest records `hooks` as %s, the source wrote %s", g.String(), w.String())
	}
}

func readJSON(t *testing.T, path string, into any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, into); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}
