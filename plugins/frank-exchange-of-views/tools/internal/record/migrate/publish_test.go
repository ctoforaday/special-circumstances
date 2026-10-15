package migrate

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// publishArchive is the archive the publish's sync order is measured on: the smallest in the store,
// with a class registry, so its migrated run has channels, a registry translation, a database and a
// manifest — every kind of file the publish syncs.
const publishArchive = "2026-08-22_is-7-prime.tar.gz"

// tinySource writes a former-vocabulary record with one register and one friction — the smallest run
// that migrates clean — and, with refused set, one event of a word no vocabulary declares, which the
// migration refuses. The tests that do not need an archive's shapes use it, because each migration of
// an archive costs seconds of per-event fsync.
func tinySource(t *testing.T, refused bool) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	stmts := []string{
		`CREATE TABLE "events" ("id" INTEGER PRIMARY KEY, "seat_id" TEXT NOT NULL, "round" INTEGER NOT NULL, "ts" TEXT NOT NULL, "type" TEXT NOT NULL, "key" TEXT)`,
		`CREATE TABLE "enum_event_type" ("value" TEXT PRIMARY KEY, "means" TEXT NOT NULL)`,
		`CREATE TABLE "register" ("event_id" INTEGER PRIMARY KEY, "tool_version" TEXT)`,
		`CREATE TABLE "friction" ("event_id" INTEGER PRIMARY KEY, "text" TEXT, "kind" TEXT, "estopped_by" TEXT)`,
		`INSERT INTO "enum_event_type" VALUES ('register', 'former'), ('friction', 'former')`,
		`INSERT INTO "events" VALUES (1, 'blue-lane-1', 1, '2026-09-02T10:00:01.000000001Z', 'register', NULL)`,
		`INSERT INTO "register" VALUES (1, 'abc1234')`,
		`INSERT INTO "events" VALUES (2, 'blue-lane-1', 1, '2026-09-02T10:00:02.000000002Z', 'friction', NULL)`,
		`INSERT INTO "friction" VALUES (2, 'the fetch cache refused a legitimate URL', 'defect', NULL)`,
	}
	if refused {
		stmts = append(stmts,
			`INSERT INTO "enum_event_type" VALUES ('seance', 'a word no schema declares')`,
			`CREATE TABLE "seance" ("event_id" INTEGER PRIMARY KEY, "spirit" TEXT)`,
			`INSERT INTO "events" VALUES (3, 'blue-lane-1', 1, '2026-09-02T10:00:03.000000003Z', 'seance', NULL)`,
			`INSERT INTO "seance" VALUES (3, 'banquo')`)
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("tiny source: %s: %v", s, err)
		}
	}
	return dir
}

// seamCall is one call through the publish's seams, in order.
type seamCall struct {
	op   string // sync, rename or remove
	path string // for a rename, the build being moved
	dir  bool   // a sync of a directory
	sum  string // a sync of a regular file: the file's content hash at the moment of the sync
}

// recordSeams replaces the publish's seams with recorders that call through to the real ones, and
// restores them when the test ends.
func recordSeams(t *testing.T) *[]seamCall {
	t.Helper()
	calls := &[]seamCall{}
	prevSync, prevRename, prevRemove := syncFile, renameDir, removeBuild
	t.Cleanup(func() { syncFile, renameDir, removeBuild = prevSync, prevRename, prevRemove })
	syncFile = func(p string) error {
		c := seamCall{op: "sync", path: p}
		if fi, err := os.Lstat(p); err == nil && fi.IsDir() {
			c.dir = true
		} else {
			c.sum = hashFile(t, p)
		}
		*calls = append(*calls, c)
		return prevSync(p)
	}
	renameDir = func(from, to string) error {
		*calls = append(*calls, seamCall{op: "rename", path: from})
		return prevRename(from, to)
	}
	removeBuild = func(p string) error {
		*calls = append(*calls, seamCall{op: "remove", path: p})
		return prevRemove(p)
	}
	return calls
}

func hashFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("hashing %s: %v", p, err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// builds lists the build directories left beside to.
func builds(t *testing.T, to string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(filepath.Dir(to), "."+filepath.Base(to)+buildMarker+"*"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func listing(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		out = append(out, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// M3: every file of the published run is synced after its last write, every directory of it is
// synced, all of it before the rename; after the rename, --to's parent, each parent the migration
// created and the nearest ancestor that existed are synced. The --to is two levels below an existing
// directory so every kind of post-rename sync occurs. The recorders call through to the real seams,
// so the zero-warning assertion is the real seams' answer on this platform.
func TestThePublishSyncsEveryFileBeforeTheRename(t *testing.T) {
	src := recordtest.RequireArchive(t, publishArchive)
	base := recordtest.TmpRun(t)
	to := filepath.Join(base, "a", "b", "run")
	calls := recordSeams(t)

	m, err := Migrate(src, to, Entries(), Options{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	renames := slices.IndexFunc(*calls, func(c seamCall) bool { return c.op == "rename" })
	if renames < 0 {
		t.Fatal("no rename was recorded")
	}
	build := (*calls)[renames].path
	if !strings.HasPrefix(build, filepath.Join(base, "a", "b", ".run"+buildMarker)) {
		t.Fatalf("the build %s is not a .run%s sibling of --to", build, buildMarker)
	}
	if _, err := os.Lstat(build); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the build %s is still there after the publish (%v)", build, err)
	}
	atTo := func(p string) string { return strings.Replace(p, build, to, 1) }
	fileSyncs := map[string]string{} // --to path -> content hash at its sync
	dirSyncsBefore := map[string]bool{}
	for _, c := range (*calls)[:renames] {
		switch {
		case c.op == "sync" && c.dir:
			dirSyncsBefore[atTo(c.path)] = true
		case c.op == "sync":
			fileSyncs[atTo(c.path)] = c.sum
		default:
			t.Errorf("%s %s before the rename, which only syncs precede", c.op, c.path)
		}
	}
	files := 0
	for _, p := range listing(t, to) {
		fi, err := os.Lstat(p)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case fi.IsDir():
			if dirSyncs && !dirSyncsBefore[p] {
				t.Errorf("the directory %s was not synced before the rename", p)
			}
		case fi.Mode().IsRegular():
			files++
			sum, ok := fileSyncs[p]
			if !ok {
				t.Errorf("the file %s was not synced before the rename", p)
			} else if now := hashFile(t, p); now != sum {
				t.Errorf("the file %s was written after it was synced", p)
			}
		}
	}
	if files < 4 {
		t.Fatalf("the published run holds %d file(s) — too few for this to have measured the channels, registry, record and manifest", files)
	}
	if dirSyncs {
		after := map[string]bool{}
		for _, c := range (*calls)[renames+1:] {
			if c.op == "sync" && c.dir {
				after[c.path] = true
			}
		}
		for _, d := range []string{filepath.Join(base, "a", "b"), filepath.Join(base, "a"), base} {
			if !after[d] {
				t.Errorf("%s was not synced after the rename (synced: %v)", d, after)
			}
		}
	}
	if len(m.Warnings) != 0 {
		t.Errorf("the real seams warned on a local filesystem: %q", m.Warnings)
	}
	if strings.Contains(listingString(t, to), ".migrate-src") {
		t.Error("the source's scratch copy is in the published run")
	}
}

func listingString(t *testing.T, dir string) string { return strings.Join(listing(t, dir), "\n") }

// An existing empty --to is replaced by the rename, never removed first: when the rename is called
// it is still there. Windows removes it first by design, so this is Unix's claim.
func TestAnEmptyToStillExistsWhenTheRenameReplacesIt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows removes an empty --to before the rename")
	}
	src := tinySource(t, false)
	to := filepath.Join(recordtest.TmpRun(t), "out")
	if err := os.Mkdir(to, 0o755); err != nil {
		t.Fatal(err)
	}
	recordSeams(t)
	inner := renameDir
	checked := false
	renameDir = func(from, dst string) error {
		fi, err := os.Lstat(dst)
		if err != nil || !fi.IsDir() {
			t.Errorf("--to was gone when the rename was called (%v) — something removed it first", err)
		}
		checked = true
		return inner(from, dst)
	}
	if _, err := Migrate(src, to, Entries(), Options{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !checked {
		t.Fatal("the rename seam was never called")
	}
	if _, err := os.Stat(filepath.Join(to, "inputs", ManifestName)); err != nil {
		t.Fatalf("the published run has no manifest: %v", err)
	}
}

// A directory that cannot be synced is a warning, never an error, and the warning names the
// directory on the --to side, never the build.
func TestADirectorySyncThatFailsIsAWarningNamingTo(t *testing.T) {
	if !dirSyncs {
		t.Skip("no directory is synced on this platform")
	}
	src := tinySource(t, false)
	base := recordtest.TmpRun(t)
	to := filepath.Join(base, "out")
	recordSeams(t)
	inner := syncFile
	syncFile = func(p string) error {
		if fi, err := os.Lstat(p); err == nil && fi.IsDir() {
			return errors.New("injected directory sync failure")
		}
		return inner(p)
	}
	m, err := Migrate(src, to, Entries(), Options{})
	if err != nil {
		t.Fatalf("a directory that could not be synced failed the migration: %v", err)
	}
	if len(m.Warnings) == 0 {
		t.Fatal("no warning for the directories that could not be synced")
	}
	all := strings.Join(m.Warnings, "\n")
	if strings.Contains(all, buildMarker) {
		t.Errorf("a warning names the build, which no longer exists: %s", all)
	}
	for _, want := range []string{filepath.Join(to, "records"), base} {
		if !slices.ContainsFunc(m.Warnings, func(w string) bool { return strings.Contains(w, want+" ") }) {
			t.Errorf("no warning names %s: %s", want, all)
		}
	}
}

// A file that cannot be synced is an error, --to is left untouched and the build is removed. The
// returned error carries the injected one (errors.Is, and its feov code), names --to where the text
// named the build, and says the path belongs to the run that was being built. Any one file failing
// does it, so the first and the last file synced are each made to fail.
func TestAFileSyncThatFailsLeavesToUntouched(t *testing.T) {
	src := tinySource(t, false)
	calls := recordSeams(t)
	if _, err := Migrate(src, filepath.Join(recordtest.TmpRun(t), "count"), Entries(), Options{}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range *calls {
		if c.op == "sync" && !c.dir {
			n++
		}
	}
	for _, failAt := range []int{1, n} {
		t.Run(fmt.Sprintf("file %d of %d", failAt, n), func(t *testing.T) {
			to := filepath.Join(recordtest.TmpRun(t), "out")
			injected := feov.Errorf(feov.Conflict, "injected file sync failure")
			inner := syncFile
			seen := 0
			syncFile = func(p string) error {
				if fi, err := os.Lstat(p); err == nil && fi.Mode().IsRegular() {
					if seen++; seen == failAt {
						return injected
					}
				}
				return inner(p)
			}
			t.Cleanup(func() { syncFile = inner })
			m, err := Migrate(src, to, Entries(), Options{})
			if err == nil {
				t.Fatal("a file that could not be synced published the run")
			}
			if m != nil {
				t.Errorf("a failed migration returned a manifest: %+v", m)
			}
			if !errors.Is(err, injected) {
				t.Errorf("the returned error does not wrap the injected one: %v", err)
			}
			if got := feov.CodeOf(err); got != string(feov.Conflict) {
				t.Errorf("feov.CodeOf = %q, want %q", got, feov.Conflict)
			}
			msg := err.Error()
			if strings.Contains(msg, buildMarker) {
				t.Errorf("the error names the removed build: %s", msg)
			}
			if !strings.Contains(msg, to) || !strings.Contains(msg, "being built") {
				t.Errorf("the error does not say its path under %s belongs to the run being built: %s", to, msg)
			}
			if _, err := os.Lstat(to); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("--to was touched: %v", err)
			}
			if left := builds(t, to); len(left) != 0 {
				t.Errorf("the build was left behind: %v", left)
			}
		})
	}
}

// When the build cannot be removed after a failure, the returned error names it — the leftover is
// never hidden, even though the failure itself named no path.
func TestABuildThatCannotBeRemovedIsNamed(t *testing.T) {
	src := tinySource(t, false)
	to := filepath.Join(recordtest.TmpRun(t), "out")
	recordSeams(t)
	injected := errors.New("injected file sync failure")
	syncFile = func(string) error { return injected }
	removeBuild = func(string) error { return errors.New("injected removal failure") }
	_, err := Migrate(src, to, Entries(), Options{})
	if err == nil {
		t.Fatal("a failed sync published the run")
	}
	left := builds(t, to)
	if len(left) != 1 {
		t.Fatalf("want the one kept build, found %v", left)
	}
	if !strings.Contains(err.Error(), left[0]) {
		t.Errorf("the error does not name the kept build %s: %v", left[0], err)
	}
	if !errors.Is(err, injected) {
		t.Errorf("the error lost the failure that caused it: %v", err)
	}
}

// A refusal's text that names the run's path names --to in the published manifest, in the manifest
// returned to --json, and in nothing that names the build; the returned error points at the published
// manifest and does not say the run is unpublished. The class registry is made invalid in the source
// so every mint is refused naming the registry's path.
func TestARefusalNamesToNotTheBuild(t *testing.T) {
	src := recordtest.RequireArchive(t, publishArchive)
	reg := filepath.Join(src, "records", "class-registry.json")
	b, err := os.ReadFile(reg)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]any
	if err := json.Unmarshal(b, &top); err != nil {
		t.Fatal(err)
	}
	classes, _ := top["classes"].([]any)
	if len(classes) == 0 {
		t.Fatalf("%s has no classes to make invalid", reg)
	}
	for _, c := range classes {
		c.(map[string]any)["material_default"] = "sometimes"
	}
	out, _ := json.Marshal(top)
	if err := os.WriteFile(reg, out, 0o644); err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(recordtest.TmpRun(t), "out")

	m, refusedErr := Migrate(src, to, Entries(), Options{})
	if refusedErr == nil || m == nil || len(m.Refusals) == 0 {
		t.Fatalf("an invalid registry did not refuse: m=%v err=%v", m, refusedErr)
	}
	regAtTo := filepath.Join(to, "records", "class-registry.json")
	if !slices.ContainsFunc(m.Refusals, func(r Refusal) bool { return strings.Contains(r.Err, regAtTo) }) {
		t.Fatalf("no refusal names the registry at %s: %+v", regAtTo, m.Refusals[0])
	}
	published, err := os.ReadFile(filepath.Join(to, "inputs", ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(published), buildMarker) {
		t.Error("the published manifest names the build")
	}
	returned, err2 := json.Marshal(m)
	if err2 != nil {
		t.Fatal(err2)
	}
	if strings.Contains(string(returned), buildMarker) {
		t.Error("the returned manifest, which --json prints, names the build")
	}
	onDisk, err2 := ReadManifest(to)
	if err2 != nil {
		t.Fatal(err2)
	}
	if !reflect.DeepEqual(onDisk.Refusals, m.Refusals) {
		t.Error("the returned manifest's refusals differ from the published file's")
	}
	msg := refusedErr.Error()
	if strings.Contains(msg, buildMarker) || strings.Contains(msg, "being built") {
		t.Errorf("the refusal error of a published run names the build or calls it unpublished: %s", msg)
	}
	if !strings.Contains(msg, filepath.Join(to, "inputs", ManifestName)) {
		t.Errorf("the refusal error does not point at the published manifest: %s", msg)
	}
}

// A rename that fails keeps the build, complete, and every text that reaches the operator names it:
// the returned error, and — on the refusal path — the refusal count, which names the kept build's
// manifest and never a manifest at --to that does not exist.
func TestARenameThatFailsKeepsTheBuildAndNamesIt(t *testing.T) {
	for _, refused := range []bool{false, true} {
		t.Run(fmt.Sprintf("refused=%v", refused), func(t *testing.T) {
			src := tinySource(t, refused)
			to := filepath.Join(recordtest.TmpRun(t), "out")
			recordSeams(t)
			renameDir = func(string, string) error { return errors.New("injected rename failure") }
			m, err := Migrate(src, to, Entries(), Options{})
			if err == nil {
				t.Fatal("a failed rename reported success")
			}
			if m == nil {
				t.Fatal("a failed rename returned no manifest, so the verb cannot print its warnings")
			}
			left := builds(t, to)
			if len(left) != 1 {
				t.Fatalf("want the one kept build, found %v", left)
			}
			if _, err := os.Stat(filepath.Join(left[0], "inputs", ManifestName)); err != nil {
				t.Errorf("the kept build is not a complete run: %v", err)
			}
			if _, err := os.Lstat(to); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("--to was touched: %v", err)
			}
			msg := err.Error()
			if !strings.Contains(msg, left[0]) {
				t.Errorf("the error does not name the kept build %s: %s", left[0], msg)
			}
			if strings.Contains(msg, filepath.Join(to, "inputs", ManifestName)) {
				t.Errorf("the error names a manifest at --to, which was never published: %s", msg)
			}
			if refused {
				if !strings.Contains(msg, "1 event(s) refused") {
					t.Errorf("the error does not carry the refusal count: %s", msg)
				}
				if !strings.Contains(msg, filepath.Join(left[0], "inputs", ManifestName)) {
					t.Errorf("the refusal count does not name the kept build's manifest: %s", msg)
				}
			}
		})
	}
}

// V.7: what --to may be. Each refusal comes before anything is created or removed, and leaves what
// was found untouched.
func TestWhatToMayBe(t *testing.T) {
	src := tinySource(t, false)
	refusedUntouched := func(t *testing.T, base, to string) error {
		t.Helper()
		before := listing(t, base)
		_, err := Migrate(src, to, Entries(), Options{})
		if err == nil {
			t.Fatalf("--to %s was accepted", to)
		}
		if after := listing(t, base); !slices.Equal(before, after) {
			t.Errorf("a refusal created or removed something:\nbefore %v\nafter  %v", before, after)
		}
		return err
	}
	t.Run("a regular file", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		to := filepath.Join(base, "out")
		if err := os.WriteFile(to, []byte("keep me"), 0o644); err != nil {
			t.Fatal(err)
		}
		err := refusedUntouched(t, base, to)
		if !strings.Contains(err.Error(), "regular file") {
			t.Errorf("the refusal does not say what it found: %v", err)
		}
		if b, _ := os.ReadFile(to); string(b) != "keep me" {
			t.Errorf("the file at --to was changed: %q", b)
		}
	})
	t.Run("a symlink to an empty directory", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		target := filepath.Join(base, "target")
		if err := os.Mkdir(target, 0o755); err != nil {
			t.Fatal(err)
		}
		to := filepath.Join(base, "out")
		if err := os.Symlink(target, to); err != nil {
			t.Skipf("no symlinks here: %v", err)
		}
		err := refusedUntouched(t, base, to)
		if !strings.Contains(err.Error(), "symbolic link") {
			t.Errorf("the refusal does not say what it found: %v", err)
		}
		if es, err := os.ReadDir(target); err != nil || len(es) != 0 {
			t.Errorf("the link's target was written through: %v %v", es, err)
		}
	})
	t.Run("a non-empty directory", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		to := filepath.Join(base, "out")
		if err := os.MkdirAll(filepath.Join(to, "inputs"), 0o755); err != nil {
			t.Fatal(err)
		}
		err := refusedUntouched(t, base, to)
		if !strings.Contains(err.Error(), "not empty") {
			t.Errorf("the refusal does not say what it found: %v", err)
		}
	})
	t.Run("a separated record root", func(t *testing.T) {
		base := recordtest.TmpRun(t)
		root := filepath.Join(base, "root")
		t.Setenv(record.RecordRootEnv, root)
		err := refusedUntouched(t, base, filepath.Join(base, "x", "y"))
		if !strings.Contains(err.Error(), record.RecordRootEnv) {
			t.Errorf("the refusal does not name %s: %v", record.RecordRootEnv, err)
		}
	})
	t.Run("an empty directory takes the build's mode", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Unix permission bits")
		}
		base := recordtest.TmpRun(t)
		probe := filepath.Join(base, "probe")
		if err := os.Mkdir(probe, 0o755); err != nil {
			t.Fatal(err)
		}
		want, err := os.Stat(probe)
		if err != nil {
			t.Fatal(err)
		}
		to := filepath.Join(base, "out")
		if err := os.Mkdir(to, 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := Migrate(src, to, Entries(), Options{}); err != nil {
			t.Fatal(err)
		}
		got, err := os.Stat(to)
		if err != nil {
			t.Fatal(err)
		}
		// The stated residue: the replaced directory's mode is not carried. A change to that is a
		// change to the help, and this is where it is seen.
		if got.Mode().Perm() != want.Mode().Perm() {
			t.Errorf("--to's mode is %v, want the build's %v (0755 under the umask)", got.Mode().Perm(), want.Mode().Perm())
		}
	})
	t.Run("an unwritable parent", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("needs Unix permission bits enforced on this user")
		}
		base := recordtest.TmpRun(t)
		parent := filepath.Join(base, "ro")
		if err := os.Mkdir(parent, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(parent, 0o755) })
		err := refusedUntouched(t, base, filepath.Join(parent, "out"))
		if !errors.Is(err, fs.ErrPermission) || !strings.Contains(err.Error(), "must be writable") {
			t.Errorf("want the OS error wrapped and the reason given: %v", err)
		}
	})
	t.Run("a base name with no room for the suffix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Unix name limit")
		}
		base := recordtest.TmpRun(t)
		err := refusedUntouched(t, base, filepath.Join(base, strings.Repeat("n", 240)))
		if !errors.Is(err, syscall.ENAMETOOLONG) {
			t.Errorf("want ENAMETOOLONG wrapped: %v", err)
		}
	})
}

// killChildEnv carries the kill test's arguments into its child process.
const killChildEnv = "MIGRATE_KILL_CHILD"

// M2: a migration killed mid-replay leaves --to as it was and its build beside it; the next migration
// into the same --to succeeds and leaves that build alone. The child blocks inside the replay's
// writing pass after writing a marker, so the kill lands after the build exists and replay has begun.
func TestAKilledMigrationLeavesToAsItWas(t *testing.T) {
	if args := os.Getenv(killChildEnv); args != "" {
		var a struct{ From, To, Marker string }
		if err := json.Unmarshal([]byte(args), &a); err != nil {
			t.Fatal(err)
		}
		base := Entries()
		reg := Registry{}
		for k, v := range base {
			reg[k] = v
		}
		visits := map[int64]int{}
		reg["friction"] = Entry{Translate: func(old OldEvent, dst record.Run) ([]proto.Message, error) {
			// The first visit is the pass that learns the ids; the second is the pass that writes.
			if visits[old.ID]++; visits[old.ID] == 2 {
				if err := os.WriteFile(a.Marker, nil, 0o644); err != nil {
					t.Fatal(err)
				}
				time.Sleep(time.Hour)
			}
			return base.Translate(old, dst)
		}}
		_, err := Migrate(a.From, a.To, reg, Options{})
		t.Fatalf("the migration was meant to block until killed, and returned %v", err)
	}

	src := tinySource(t, false)
	base := recordtest.TmpRun(t)
	to := filepath.Join(base, "out")
	if err := os.Mkdir(to, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(recordtest.TmpRun(t), "blocked")
	args, _ := json.Marshal(struct{ From, To, Marker string }{src, to, marker})
	cmd := exec.Command(os.Args[0], "-test.run=^TestAKilledMigrationLeavesToAsItWas$", "-test.count=1")
	cmd.Env = append(os.Environ(), killChildEnv+"="+string(args))
	var out strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	deadline := time.After(5 * time.Minute)
	for blocked := false; !blocked; {
		select {
		case err := <-exited:
			t.Fatalf("the child exited before blocking (%v):\n%s", err, out.String())
		case <-deadline:
			cmd.Process.Kill()
			t.Fatalf("the child never reached the replay:\n%s", out.String())
		case <-time.After(20 * time.Millisecond):
			_, err := os.Stat(marker)
			blocked = err == nil
		}
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-exited

	if es, err := os.ReadDir(to); err != nil || len(es) != 0 {
		t.Fatalf("--to is not the empty directory it was: %v %v", es, err)
	}
	first := builds(t, to)
	if len(first) != 1 {
		t.Fatalf("want the killed migration's build beside --to, found %v", first)
	}
	if _, err := Migrate(src, to, Entries(), Options{}); err != nil {
		t.Fatalf("a migration into the same --to after a killed one: %v", err)
	}
	if _, err := os.Stat(filepath.Join(to, "inputs", ManifestName)); err != nil {
		t.Fatalf("the second migration published no run: %v", err)
	}
	if after := builds(t, to); !slices.Equal(after, first) {
		t.Errorf("the killed migration's build was touched — nothing reaps one: before %v, after %v", first, after)
	}
}
