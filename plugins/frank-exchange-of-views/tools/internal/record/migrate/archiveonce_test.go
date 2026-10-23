package migrate

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EACH ARCHIVED RUN IS MIGRATED ONCE PER TEST PROCESS. Nine tests sweep every tarball in
// run-archive and thirteen more ask for one; each used to run the whole migration again, so seven
// tests were 169 of this package's 205 seconds and the cost grew with every archived run. The
// migration is real — the first asker runs Migrate against the extracted tarball — and it is run
// once.
//
// EVERY ASKER GETS ITS OWN COPY, the first included. Seven files in this package write to a
// record; a shared one would let a single missed writer change what every later test reads, and
// nothing would say which. A copy is a file copy, and no call site has to be audited for writes.
//
// THE STORE ANSWERS "THIS ARCHIVE, MIGRATED" AND NOTHING ELSE. It holds one migration per archive,
// so a test whose claim is agreement BETWEEN migrations must not read it: handed two copies of one
// migration it compares a result with itself and passes whatever the code does. Such a test calls
// Migrate itself, once per migration it claims.
var archiveStore = struct {
	sync.Mutex
	root    string
	entries map[string]*archiveEntry
}{entries: map[string]*archiveEntry{}}

type archiveEntry struct {
	once       sync.Once
	dir        string
	manifest   *migration.Manifest
	err        error
	migrations int
}

// MigratedArchive returns the manifest of the named tarball's migration and a directory holding a
// private copy of the migrated record, released when the test ends. The manifest is a copy of the
// struct; its slices are the store's and are read, never written.
func MigratedArchive(t *testing.T, name string) (*migration.Manifest, string) {
	t.Helper()
	archiveStore.Lock()
	e := archiveStore.entries[name]
	if e == nil {
		e = &archiveEntry{}
		archiveStore.entries[name] = e
	}
	archiveStore.Unlock()

	e.once.Do(func() {
		e.migrations++
		src := recordtest.RequireArchive(t, name)
		dir, err := archiveStoreDir(name)
		if err != nil {
			e.err = err
			return
		}
		e.dir = dir
		e.manifest, e.err = Migrate(src, dir, Entries(), Options{})
		// Migrate closes every handle under the directory it built before it publishes it, so the
		// record layer's cache holds nothing for the destination and this finds nothing open. It is
		// kept because a copy taken under an open handle is not the record, and a handle left in
		// the cache would outlive the store.
		if cerr := recordsql.CloseUnder(dir); cerr != nil && e.err == nil {
			e.err = cerr
		}
	})
	if e.err != nil {
		t.Fatalf("migrate %s: %v (manifest: %+v)", name, e.err, e.manifest)
	}
	to := recordtest.TmpRun(t)
	if err := copyTree(e.dir, to); err != nil {
		t.Fatalf("copying the migrated %s: %v", name, err)
	}
	m := *e.manifest
	return &m, to
}

// archiveStoreDir makes the store's directory for one archive. The root is made at the first
// migration, so a process that lists tests and runs none leaves nothing behind.
func archiveStoreDir(name string) (string, error) {
	archiveStore.Lock()
	defer archiveStore.Unlock()
	if archiveStore.root == "" {
		root, err := os.MkdirTemp("", "feov-migrated-archives-")
		if err != nil {
			return "", err
		}
		archiveStore.root = root
	}
	dir := filepath.Join(archiveStore.root, name)
	return dir, os.Mkdir(dir, 0o755)
}

// ArchiveMigrations is how many times the store has migrated each archive it was asked for.
func ArchiveMigrations() map[string]int {
	archiveStore.Lock()
	defer archiveStore.Unlock()
	out := map[string]int{}
	for name, e := range archiveStore.entries {
		out[name] = e.migrations
	}
	return out
}

// RemoveArchiveStore deletes the store's directory. TestMain runs it BEFORE the orphaned-handle
// check: a handle the store left open is then a handle whose file is gone, which is what that
// check reports. A removal that fails is returned — on Windows an open handle makes the removal
// fail instead of orphaning.
func RemoveArchiveStore() error {
	archiveStore.Lock()
	defer archiveStore.Unlock()
	if archiveStore.root == "" {
		return nil
	}
	if err := os.RemoveAll(archiveStore.root); err != nil {
		return fmt.Errorf("removing the migrated-archive store: %w", err)
	}
	archiveStore.root = ""
	return nil
}

func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s is not a regular file or a directory", path)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		info, err := d.Info()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}

// THE STORE MIGRATES EACH ARCHIVE ONCE AND HANDS OUT SEPARATE COPIES. It asks for every tarball
// twice itself, so it holds whichever tests ran before it and in whichever process.
func TestTheArchiveStoreMigratesEachArchiveOnce(t *testing.T) {
	names := archivedRuns(t)
	if len(names) == 0 {
		t.Fatal("no archived run found")
	}
	for _, name := range names {
		_, a := MigratedArchive(t, name)
		_, b := MigratedArchive(t, name)
		if a == b {
			t.Errorf("%s: two askers were handed one directory, %s", name, a)
		}
		archiveStore.Lock()
		root := archiveStore.root
		archiveStore.Unlock()
		for _, dir := range []string{a, b} {
			if rel, err := filepath.Rel(root, dir); err == nil && filepath.IsLocal(rel) {
				t.Errorf("%s: an asker was handed %s, inside the store", name, dir)
			}
		}
		marker := filepath.Join(a, "written-by-one-asker")
		if err := os.WriteFile(marker, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(b, "written-by-one-asker")); err == nil {
			t.Errorf("%s: a write to one asker's copy reached another's", name)
		}
	}
	counts := ArchiveMigrations()
	for _, name := range names {
		if counts[name] != 1 {
			t.Errorf("%s was migrated %d times in this process, want 1", name, counts[name])
		}
	}
}
