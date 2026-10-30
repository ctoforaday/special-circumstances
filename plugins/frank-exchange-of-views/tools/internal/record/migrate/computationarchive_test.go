package migrate

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// archivedProofToken is a proof's anchor token, an archived note label, and a proof's id, as an
// archive before epoch 49 spells them. A token is one carrying an id: a seat's prose that names the
// token's shape (`<!--proof:p-…-->`) is the seat's words, and stays as written.
var (
	archivedProofToken = regexp.MustCompile(`<!--proof:[Pp]-[0-9a-f]{8}-->`)
	archivedProofNote  = regexp.MustCompile(`\[\^[Pp][0-9]+\]`)
	archivedProofID    = regexp.MustCompile(`\bP-[0-9a-f]{8}\b`)
)

// EVERY ARCHIVED PROOF MIGRATES TO A COMPUTATION THE STORE CAN SHOW (plans/compute-and-show.md
// §III.1.6). Over every archive: nothing is refused; no argument field keeps a proof's token, note
// label or id; every computation's id is `K-`; and every computation the manifest does not name as
// missing has its bundle's dir/ and its act's output in the migrated store.
func TestEveryArchivedProofMigratesToAComputationTheStoreCanShow(t *testing.T) {
	computations := 0
	for _, name := range archivedRuns(t) {
		res, dir := MigratedArchive(t, name)
		if len(res.Refusals) != 0 {
			t.Errorf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
			continue
		}
		merged, err := record.MergedEvents(runtest.Open(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		missing := strings.Join(res.MissingArtifacts, "\n")
		for _, e := range merged.Events {
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			eachString(body.ProtoReflect(), func(field, v string) {
				switch {
				case archivedProofToken.MatchString(v):
					t.Errorf("%s event %s: %s keeps a proof's anchor token: %.120q", name, e.GetKey(), field, v)
				case argumentFields[field] && archivedProofNote.MatchString(v):
					t.Errorf("%s event %s: %s keeps a proof's note label %q", name, e.GetKey(), field, archivedProofNote.FindString(v))
				case argumentFields[field] && archivedProofID.MatchString(v):
					t.Errorf("%s event %s: %s keeps a proof's id %q", name, e.GetKey(), field, archivedProofID.FindString(v))
				}
			})
			c, ok := recordpb.BodyAs[*recordpb.Computation](e)
			if !ok {
				continue
			}
			computations++
			id, bundle := c.GetComputationId(), c.GetBundleSha()
			if !strings.HasPrefix(id, "K-") {
				t.Errorf("%s: computation %q is not a computation's id", name, id)
			}
			if strings.Contains(missing, id+":") {
				continue
			}
			if _, err := compute.Bundle(dir, bundle); err != nil {
				t.Errorf("%s: computation %s: %v", name, id, err)
			}
			if _, err := compute.ActOutput(dir, bundle, id); err != nil {
				t.Errorf("%s: computation %s: %v", name, id, err)
			}
		}
	}
	if computations == 0 {
		t.Fatal("no archive migrated a computation — a pass that checked nothing")
	}
}

// migratedComputations is the named archive's migrated computations by id, and its run directory.
func migratedComputations(t *testing.T, name string) (map[string]*recordpb.Computation, []*recordpb.Reproduce, string) {
	t.Helper()
	res, dir := MigratedArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
	}
	merged, err := record.MergedEvents(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]*recordpb.Computation{}
	var reproduces []*recordpb.Reproduce
	for _, e := range merged.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Computation](e); ok {
			out[c.GetComputationId()] = c
		}
		if r, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
			reproduces = append(reproduces, r)
		}
	}
	return out, reproduces, dir
}

// A PROOF THE TOOL RECORDED `observed` OR `not_portable` IS A VARIANT COMPUTATION, its invariant the
// drift the tool recorded (F4): the archive states no invariant, and its drift is the one statement
// it holds of what the two runs did.
func TestAMovingArchivedProofIsVariantWithItsDriftAsInvariant(t *testing.T) {
	for _, c := range []struct{ archive, id, drift string }{
		{"2026-09-09_is-91-prime-b.tar.gz", "K-808ae840", ""},
		{"2026-08-22_is-7-prime-run2.tar.gz", "K-568517ec", "output differs from byte 131 between runs"},
	} {
		got, _, _ := migratedComputations(t, c.archive)
		k := got[c.id]
		if k == nil {
			t.Errorf("%s: no computation %s", c.archive, c.id)
			continue
		}
		if k.GetBasis() != recordpb.Basis_BASIS_VARIANT || k.GetDrift() == "" || k.GetInvariant() != k.GetDrift() {
			t.Errorf("%s %s: basis %s, drift %q, invariant %q — want variant with the drift as its invariant", c.archive, c.id, k.GetBasis(), k.GetDrift(), k.GetInvariant())
		}
		if c.drift != "" && k.GetDrift() != c.drift {
			t.Errorf("%s %s: drift %q, want %q", c.archive, c.id, k.GetDrift(), c.drift)
		}
	}
}

// THE COMMAND IS THE ONE THE ARCHIVED SCRIPT'S EXTENSION NAMED: a Go program is `go run`.
func TestAnArchivedGoProofRunsAsGoRun(t *testing.T) {
	got, _, _ := migratedComputations(t, "2026-09-11_is-91-prime-b6.tar.gz")
	seen := false
	for id, k := range got {
		if strings.HasSuffix(k.GetArgv()[len(k.GetArgv())-1], ".go") {
			seen = true
			if !slices.Equal(k.GetArgv(), []string{"go", "run", "script.go"}) {
				t.Errorf("%s: argv %q, want go run script.go", id, k.GetArgv())
			}
		}
	}
	if !seen {
		t.Fatal("b6 migrated no Go program — the fixture this test reads is gone")
	}
}

// ONE BUNDLE, MANY ACTS: the store holds a bundle's dir/ once and each act's output under its own id.
// 07d recorded nine proofs over six scripts; m32 seven over two.
func TestTheStoreHoldsEachBundleOnceAndEachActsOutput(t *testing.T) {
	for _, c := range []struct {
		archive    string
		dirs, acts int
	}{
		{"2026-10-17_is-91-prime-07d.tar.gz", 6, 9},
		{"2026-10-24_is-91-prime-m32.tar.gz", 2, 7},
	} {
		got, _, dir := migratedComputations(t, c.archive)
		dirs, _ := filepath.Glob(filepath.Join(dir, compute.StoreDir, "*", "dir"))
		acts, _ := filepath.Glob(filepath.Join(dir, compute.StoreDir, "*", "acts", "*", "output"))
		if len(dirs) != c.dirs || len(acts) != c.acts || len(got) != c.acts {
			t.Errorf("%s: %d dir/, %d act output(s), %d computation(s); want %d, %d, %d", c.archive, len(dirs), len(acts), len(got), c.dirs, c.acts, c.acts)
		}
		for _, d := range dirs {
			if entries, _ := os.ReadDir(d); len(entries) == 0 {
				t.Errorf("%s: %s is empty", c.archive, d)
			}
		}
	}
}

// AN ARCHIVED `not_run` IS AN ACT THAT RAN NOTHING, standing on the bundle's re-run: m32's
// judgements of a second sentence on a re-run bundle carry `ran` false and the re-run's comparison.
func TestAnArchivedNotRunStandsOnTheBundlesReRun(t *testing.T) {
	_, reproduces, _ := migratedComputations(t, "2026-10-24_is-91-prime-m32.tar.gz")
	standing := map[string]recordpb.Rerun{}
	stood := 0
	for _, r := range reproduces {
		if r.GetRan() {
			standing[r.GetBundleSha()] = r.GetRerun()
			continue
		}
		stood++
		if want, ok := standing[r.GetBundleSha()]; !ok || r.GetRerun() != want {
			t.Errorf("an act that ran nothing on %s states %s; the standing re-run is %s (held: %v)", r.GetBundleSha()[:12], r.GetRerun(), want, ok)
		}
	}
	if stood == 0 {
		t.Fatal("m32 carried no act that ran nothing — the fixture this test reads is gone")
	}
}

// AN ACT'S out/ SURVIVES A MIGRATION OF A STORE THAT HOLDS IT. A source at epoch 49 or later keeps
// each act's out/ beside its output; a migration that dropped it would have every later comparison
// read the re-run's out/ files as written by the re-run alone — a computation that reproduces,
// reported as one that does not.
func TestAnActsOutSurvivesTheMigrationOfAStoreThatHoldsIt(t *testing.T) {
	got, _, dir := migratedComputations(t, "2026-09-11_is-91-prime-b9.tar.gz")
	var k *recordpb.Computation
	for _, c := range got {
		if _, err := compute.ActOutput(dir, c.GetBundleSha(), c.GetComputationId()); err == nil {
			k = c
			break
		}
	}
	if k == nil {
		t.Fatal("the archive migrated no computation with its output — the fixture this test reads is gone")
	}
	want := []compute.File{{Path: "plot.svg", Body: []byte("<svg/>\n")}, {Path: "sub/n.txt", Body: []byte("91\n")}}
	for _, f := range want {
		p := filepath.Join(compute.ActPath(dir, k.GetBundleSha(), k.GetComputationId()), "out", filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, f.Body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	to := filepath.Join(t.TempDir(), "run")
	m, err := Migrate(dir, to, Entries(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Refusals) != 0 || len(m.MissingArtifacts) != 0 {
		t.Fatalf("re-migrating refused %d event(s) and lacked %q", len(m.Refusals), m.MissingArtifacts)
	}
	out, err := compute.ActOut(to, k.GetBundleSha(), k.GetComputationId())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.EqualFunc(out, want, func(a, b compute.File) bool { return a.Path == b.Path && string(a.Body) == string(b.Body) }) {
		t.Errorf("computation %s's out/ after the migration: %q, want %q", k.GetComputationId(), out, want)
	}
}
