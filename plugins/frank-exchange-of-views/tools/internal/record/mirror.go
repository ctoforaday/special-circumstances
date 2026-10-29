package record

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// The CHECKPOINT MIRROR: where a run's record is snapshotted so a lost blackboard can be
// recovered, and the one place that path is spelled.
//
// It was spelled in two places and about to be spelled in a third — `chair verdict` composed
// it to write, `run-setup` composed the root to purge, and the reaper below wanted it again.
// A path assembled from parts at each end is a fact with no schema between the writer and the
// reader: get one segment wrong and the write still succeeds, the purge still returns 0, and
// nothing distinguishes that from a clean board.

// MirrorRoot is the directory holding every run's checkpoint mirror.
//
// The USER CACHE, deliberately, not TMPDIR: a temp purge must not void the sole recovery path
// for an uncommitted record. os.UserHomeDir rather than os.UserCacheDir because that is what
// the writer has always used, and on macOS the two are different directories — switching would
// silently orphan every mirror already on disk.
func MirrorRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "feov", "run-mirror"), nil
}

// MirrorDir is the checkpoint mirror for one run directory.
//
// THE KEY IS REPRODUCED EXACTLY, and that is a constraint rather than an accident: sha1 of the
// run directory as given, with no filepath.Clean and no case folding. The sibling key in
// recordroot.go does apply both, and copying that shape here would change every mirror's name
// — orphaning the ones already on disk while the writer went looking under a name that has
// never existed. The hash is not obfuscation; a run directory is an arbitrarily deep path and
// a cache filename is not.
func MirrorDir(runDir string) (string, error) {
	root, err := MirrorRoot()
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(runDir))
	return filepath.Join(root, hex.EncodeToString(sum[:])[:12]), nil
}

// WriteMirror snapshots the run's record into its checkpoint mirror and returns the mirror's
// directory.
//
// THE MIRROR IS ONE DATABASE FILE, taken through the record's own connection
// (recordsql.SnapshotInto), and nothing else. A file copy of records/ is the wrong instrument
// twice over: the database, its log and its shared memory are copied at three different moments
// from a run whose other seats are still writing, and closing the copy's descriptors drops every
// POSIX lock this process's connection holds on those files: the dead-man-switch lock on -shm, and
// off Linux the SHARED lock on record.db too (on Linux that one is an OFD lock and survives). The
// snapshot is a committed state with the log folded in, so the mirror opens on its own.
//
// THE OLD MIRROR STANDS UNTIL THE NEW ONE EXISTS: the snapshot is staged and synced beside it,
// then everything else in the directory — the previous snapshot's log or shared memory, which the
// next open would replay into the new database — is removed, then the staged file is renamed over
// record.db and the directory synced. A failure before the rename leaves the last good mirror.
func WriteMirror(run Run) (string, error) {
	mirror, err := MirrorDir(run.Dir())
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(mirror, 0o755); err != nil {
		return "", err
	}
	src := filepath.Join(run.Records(), dbName)
	if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
		return mirror, nil
	} else if err != nil {
		return "", err
	}
	staged := dbName + ".partial"
	if err := recordsql.SnapshotInto(src, filepath.Join(mirror, staged)); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(mirror)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.Name() == staged {
			continue
		}
		if err := os.RemoveAll(filepath.Join(mirror, e.Name())); err != nil {
			return "", err
		}
	}
	if err := os.Rename(filepath.Join(mirror, staged), filepath.Join(mirror, dbName)); err != nil {
		return "", err
	}
	return mirror, syncDir(mirror)
}

// syncDir makes the directory's entries durable — the rename that published the snapshot. Windows
// has no directory handle to sync, and NTFS journals the rename itself.
func syncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		_ = d.Close()
		return fmt.Errorf("syncing the mirror directory %s: %w", dir, err)
	}
	return d.Close()
}

// PurgeStaleMirrors removes mirrors untouched for maxAgeDays and reports how many went.
//
// AGE IS MODIFICATION TIME, which is what makes this safe to call while other runs are live: a
// mirror is rewritten every epoch, so an active run's is minutes old however long the run has
// been going. What ages out is a mirror whose run stopped — the orphan of a crashed or
// abandoned run, which is the only thing here that can never be copied back.
//
// A ZERO IS AMBIGUOUS ON ITS OWN and the caller is expected to say which zero it got: no root
// yet, an unreadable root, and a root of entirely fresh mirrors all return 0, and only the
// last of those is a healthy board.
func PurgeStaleMirrors(mirrorRoot string, now time.Time, maxAgeDays int) int {
	entries, err := os.ReadDir(mirrorRoot)
	if err != nil {
		return 0
	}
	purged := 0
	cutoff := time.Duration(maxAgeDays) * 24 * time.Hour
	for _, e := range entries {
		p := filepath.Join(mirrorRoot, e.Name())
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > cutoff {
			if os.RemoveAll(p) == nil {
				purged++
			}
		}
	}
	return purged
}
