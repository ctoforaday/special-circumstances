package migrate_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/verify"
	"google.golang.org/protobuf/proto"
)

// THE ARCHIVE IS THE FIXTURE (plan §V.4). The 2026-09-02 run is the one SQLite-era record
// whose vocabulary the current schema retired, and its mining pass is the ground truth this
// replay is held to: 813 events across 29 words WAL-inclusive — friction 35, friction_none
// 10, opinion 17 — which the docket-pair synthesis turns into 830.
//
// Skipped LOUDLY where the tarball is absent (a consumer checkout without run-archive/),
// never silently green.
func TestQuadraticFormulaArchive(t *testing.T) {
	tarball := archivePath(t, "2026-09-02_quadratic-formula.tar.gz")
	runDir := t.TempDir()
	untar(t, tarball, runDir)

	toDir := recordtest.TmpRun(t)
	res, merr := migrate.Migrate(runDir, toDir, migrate.Entries(), migrate.Options{})
	if res == nil {
		t.Fatalf("no manifest at all: %v", merr)
	}

	// The WAL-inclusive census, pinned: a shorter read means the file-set contract broke.
	total := 0
	for _, n := range res.In {
		total += n
	}
	if total != 813 || res.In["friction"] != 35 || res.In["friction_none"] != 10 || res.In["opinion"] != 17 {
		t.Fatalf("the source census moved: total=%d friction=%d friction_none=%d opinion=%d",
			total, res.In["friction"], res.In["friction_none"], res.In["opinion"])
	}

	// Refusals are the tool's output — and on THIS fixture the expected output is none.
	// Each one that appears is a registry lesson; the failure message carries the class.
	if len(res.Refusals) > 0 {
		byClass := map[string]int{}
		var sample []string
		for _, r := range res.Refusals {
			key := r.Word + ": " + firstLine(r.Err)
			byClass[key]++
			if byClass[key] == 1 && len(sample) < 12 {
				sample = append(sample, key)
			}
		}
		t.Fatalf("%d refusal(s) in %d class(es):\n  %s", len(res.Refusals), len(byClass), strings.Join(sample, "\n  "))
	}
	if merr != nil {
		t.Fatalf("Migrate returned an error with no refusals: %v", merr)
	}

	outTotal := 0
	for _, n := range res.Out {
		outTotal += n
	}
	if outTotal != 830 {
		t.Fatalf("813 in + 17 synthesized docket motions = 830 out; got %d (%v)", outTotal, res.Out)
	}
	if res.Out["log"] != 45 || res.Out["motion"] != res.In["motion"]+17 || res.Out["motion_rule"] != res.In["motion_rule"]+17 {
		t.Errorf("per-word arithmetic: out=%v in=%v", res.Out, res.In)
	}

	// The migrated record answers the CURRENT readers: the family loads, and the verify
	// suite runs over it exactly as over a native record.
	dst := runtest.Open(t, toDir)
	fam, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatalf("the migrated record does not read as a record: %v", err)
	}
	if len(fam.Events) != 830 {
		t.Errorf("family holds %d events, want 830", len(fam.Events))
	}
	assertRoundless(t, fam)
	if len(res.GapIDs) != 26 || res.GapIDs["R1-1"] != "G1" || res.GapIDs["R2-1"] != "G10" {
		t.Errorf("the manifest's gap-id table: %d entries, R1-1=%q R2-1=%q — want 26, G1, G10 (nine gaps minted in round 1)", len(res.GapIDs), res.GapIDs["R1-1"], res.GapIDs["R2-1"])
	}
	if res.Serialized["red-lens-r3-L2"] != 25 {
		t.Errorf("serialized instances: %v — the round-3 second citation instance has 25 events", res.Serialized)
	}
	checks := verify.Run(fam)
	for _, c := range checks {
		if !c.OK {
			t.Errorf("verify [%s] fails on the migrated record: %s", c.Name, c.Detail)
		}
	}

	// DETERMINISM (plan §V.6): the same source migrates to the same record. The events
	// comparison is the contract — byte-identical databases are SQLite's business.
	toDir2 := recordtest.TmpRun(t)
	res2, merr2 := migrate.Migrate(runDir, toDir2, migrate.Entries(), migrate.Options{})
	if merr2 != nil || res2.SourceHash != res.SourceHash {
		t.Fatalf("second migration: %v (hash %q vs %q)", merr2, res2.SourceHash, res.SourceHash)
	}
	again, err := record.FamilyOf(runtest.Open(t, toDir2))
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Events) != len(fam.Events) {
		t.Fatalf("two migrations, two sizes: %d vs %d", len(again.Events), len(fam.Events))
	}
	mo := proto.MarshalOptions{Deterministic: true}
	for i := range fam.Events {
		a, aerr := mo.Marshal(fam.Events[i])
		b, berr := mo.Marshal(again.Events[i])
		if aerr != nil || berr != nil || !bytes.Equal(a, b) {
			t.Fatalf("event %d differs between two migrations of one source", i)
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// archivePath walks up from the package to the repository's run-archive, skipping loudly
// when it is not there.
func archivePath(t *testing.T, name string) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "run-archive", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skipf("run-archive/%s not found above this checkout — the archive drive needs the repository's run-archive", name)
	return ""
}

func untar(t *testing.T, tarball, into string) {
	t.Helper()
	f, err := os.Open(tarball)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Clean(h.Name)
		if strings.HasPrefix(name, "..") {
			t.Fatalf("tar entry escapes: %q", h.Name)
		}
		dst := filepath.Join(into, name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dst, 0o755); err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				t.Fatal(err)
			}
			out, err := os.Create(dst)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec // a test fixture from this repo's own archive
				t.Fatal(err)
			}
			out.Close()
		}
	}
}
