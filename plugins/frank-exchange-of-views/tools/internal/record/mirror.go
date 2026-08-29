package record

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

// The CHECKPOINT MIRROR: where a run's records/ is copied so a lost blackboard can be
// recovered, and the one place that path is spelled.
//
// It was spelled in two places and about to be spelled in a third — `merge verdict` composed
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

// PurgeStaleMirrors removes mirrors untouched for maxAgeDays and reports how many went.
//
// AGE IS MODIFICATION TIME, which is what makes this safe to call while other runs are live: a
// mirror is rewritten every round, so an active run's is minutes old however long the run has
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
