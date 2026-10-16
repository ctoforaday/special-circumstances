package migrate

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A MIGRATION IS BUILT BESIDE --to AND MOVED INTO PLACE WHOLE.
//
// Written in place, a migration that is killed or loses power leaves a partial run at --to that
// opens and reads like a run. So the run is built in a sibling directory, every file of it is
// synced, and one rename publishes it: an interrupted migration leaves --to as it was, and the
// partial work sits in a `.<base>.migrating-<random>` sibling that nothing reads. Nothing removes a
// leftover sibling automatically either — that would delete directories by name, an operator's
// included — so the help names it and says it is safe to delete by hand when no migration is running.
//
// The three seams below are the publish's only effects on the filesystem beyond the build itself,
// and are variables so a test can record their order and make each one fail.
var (
	// syncFile fsyncs one path, a regular file or a directory.
	syncFile = fsyncPath
	// renameDir moves the finished build onto --to.
	renameDir = publishRename
	// removeBuild removes an unfinished build.
	removeBuild = os.RemoveAll
)

// buildMarker is what names a build directory: `.<base>` + buildMarker + a random suffix.
const buildMarker = ".migrating-"

// build is one migration's directory under construction, and what publishing it must sync.
type build struct {
	to  string // the destination, cleaned and absolute
	dir string // the build directory, a sibling of to; every path under it is built by filepath.Join
	// created are the parents of to that this migration made, outermost first; ancestor is the
	// nearest one that existed before, which holds the entry of the first created.
	created  []string
	ancestor string
}

// destination decides what --to may be, before anything is created or removed.
//
// It is an allow-list: absent, or a real directory (by Lstat, so not a symlink, and on Windows not a
// junction or mount point, which Lstat reports as irregular) with no entries. A file, a link, a
// non-empty directory, and an Lstat error other than not-exist are each refused with what was found.
// On Unix an existing --to must also be on its parent's device, so a mount point is refused here
// rather than failing the rename after the whole build; a bind mount of a directory on the same
// filesystem has the same device and passes, and its rename then fails after the build with the
// build kept and named. That check is not exercised by a test: making a mount point needs privilege.
//
// A SEPARATED destination is refused: with the record root declared, the run's record would be
// written outside the build under pointers keyed by the build's path, and the rename would carry a
// marker whose pointer names a directory that no longer exists.
//
// Writability of --to's parent is not decided here: newBuild's Mkdir decides it, because access(2)
// is wrong under ACLs and read-only mounts.
func destination(toDir string) (string, error) {
	to, err := filepath.Abs(filepath.Clean(toDir))
	if err != nil {
		return "", err
	}
	if record.RecordRootSet() {
		return "", fmt.Errorf("%s is set, which would keep the migrated run's record outside %s — "+
			"a migration publishes its run by moving the whole directory into place, and a separated record "+
			"would stay bound to the directory it was built in. Unset %s to migrate",
			record.RecordRootEnv, to, record.RecordRootEnv)
	}
	fi, err := os.Lstat(to)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return to, nil
	case err != nil:
		return "", fmt.Errorf("--to %s cannot be examined, so it is not known to be absent or empty: %w", to, err)
	case fi.Mode().Type() != fs.ModeDir:
		return "", fmt.Errorf("--to %s is %s, not an absent path or an empty plain directory — "+
			"a migration publishes a fresh run there and writes through nothing", to, describeMode(fi.Mode()))
	}
	entries, err := os.ReadDir(to)
	if err != nil {
		return "", fmt.Errorf("--to %s cannot be listed, so it is not known to be empty: %w", to, err)
	}
	if len(entries) > 0 {
		return "", fmt.Errorf("%s is not empty — a migration writes a FRESH sibling, never into an existing run", to)
	}
	same, err := sameDevice(to, filepath.Dir(to))
	if err != nil {
		return "", fmt.Errorf("--to %s: %w", to, err)
	}
	if !same {
		return "", fmt.Errorf("--to %s is a mount point — the run is built beside it and moved into place, "+
			"and a move cannot cross devices. Name a directory inside it instead", to)
	}
	return to, nil
}

func describeMode(m fs.FileMode) string {
	switch t := m.Type(); {
	case t == 0:
		return "a regular file"
	case t&fs.ModeSymlink != 0:
		return "a symbolic link"
	case t&fs.ModeIrregular != 0:
		return "a junction, mount point or other irregular entry"
	default:
		return "a " + t.String() + " entry"
	}
}

// newBuild creates to's missing parents, then the build directory beside to.
//
// The build's Mkdir is where an unwritable parent, a read-only filesystem, an exhausted quota, or a
// base name too long to take the build's suffix is found, and each is refused with the OS error
// wrapped. These refusals come AFTER the parents are created, so a refusal here can leave created
// parents behind; they are left, as a migration that fails later leaves them.
//
// The build is made 0755 with the umask applied — the mode copyChannels' MkdirAll gave --to when the
// run was written in place. An existing empty --to is REPLACED by it, not filled: nothing of the old
// directory is carried, so its mode, owner, group, ACLs and extended attributes are the build's. In
// a sticky parent a --to owned by another user makes the rename fail after the build, the build kept.
func newBuild(to string) (*build, error) {
	b := &build{to: to}
	parent := filepath.Dir(to)
	missing, ancestor := missingParents(parent)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, err
	}
	b.created, b.ancestor = missing, ancestor
	for range 100 {
		var r [6]byte
		if _, err := rand.Read(r[:]); err != nil {
			return nil, err
		}
		dir := filepath.Join(parent, "."+filepath.Base(to)+buildMarker+hex.EncodeToString(r[:]))
		err := os.Mkdir(dir, 0o755)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			why := ""
			if errors.Is(err, fs.ErrPermission) {
				why = " — the run is built in a directory beside --to and moved into place, so --to's parent must be writable"
			}
			return nil, fmt.Errorf("--to %s: creating the directory the run is built in failed%s: %w", to, why, err)
		}
		b.dir = dir
		return b, nil
	}
	return nil, fmt.Errorf("--to %s: no free name for the directory the run is built in after 100 tries", to)
}

// missingParents lists the directories from dir upward that do not exist yet, outermost first, and
// the nearest one that does.
func missingParents(dir string) ([]string, string) {
	var missing []string
	for {
		if _, err := os.Lstat(dir); err == nil {
			break
		}
		up := filepath.Dir(dir)
		if up == dir {
			break
		}
		missing = append([]string{dir}, missing...)
		dir = up
	}
	return missing, dir
}

// syncTree fsyncs every regular file under dir, then — where the platform allows a directory to be
// synced — every directory, deepest first. A file that does not sync is an error; a directory that
// does not sync is a warning, because some filesystems (network and FUSE mounts) cannot sync one and
// the run's files are on disk regardless.
func syncTree(dir string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			dirs = append(dirs, path)
		case d.Type().IsRegular():
			if err := syncFile(path); err != nil {
				return fmt.Errorf("syncing %s: %w", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !dirSyncs {
		return nil, nil
	}
	sort.SliceStable(dirs, func(i, j int) bool { return depth(dirs[i]) > depth(dirs[j]) })
	var warnings []string
	for _, d := range dirs {
		warnings = append(warnings, syncDirWarning(d)...)
	}
	return warnings, nil
}

func depth(p string) int { return strings.Count(p, string(filepath.Separator)) }

func syncDirWarning(d string) []string {
	if err := syncFile(d); err != nil {
		return []string{fmt.Sprintf("the directory %s could not be synced (%v) — its entries may not survive a power loss", d, err)}
	}
	return nil
}

// publish is the build's last act: release every handle, sync the tree, move it onto --to, and sync
// the directories whose entries the move and newBuild changed. An error before the rename is the
// caller's to discard the build over; a rename error (kept) leaves the build in place, since it holds
// a complete synced run, and names it.
func (b *build) publish() (warnings []string, kept bool, err error) {
	if err := recordsql.CloseUnder(b.dir); err != nil {
		return nil, false, err
	}
	warnings, err = syncTree(b.dir)
	if err != nil {
		return nil, false, err
	}
	if err := renameDir(b.dir, b.to); err != nil {
		return warnings, true, fmt.Errorf("the migrated run is complete and synced in %s, but moving it to %s failed, so it is kept there: %w",
			b.dir, b.to, err)
	}
	if !dirSyncs {
		return warnings, false, nil
	}
	// The rename changed --to's parent; each created parent gained an entry; the nearest ancestor that
	// existed gained the first of them.
	seen := map[string]bool{}
	order := []string{filepath.Dir(b.to)}
	for i := len(b.created) - 1; i >= 0; i-- {
		order = append(order, b.created[i])
	}
	if len(b.created) > 0 {
		order = append(order, b.ancestor)
	}
	for _, d := range order {
		if seen[d] {
			continue
		}
		seen[d] = true
		warnings = append(warnings, syncDirWarning(d)...)
	}
	return warnings, false, nil
}

// discard removes an unfinished build after err. When the removal fails the build is kept and the
// returned error names it, joined to err, so the leftover is never hidden even when err names no
// path.
func (b *build) discard(err error) error {
	if cerr := recordsql.CloseUnder(b.dir); cerr != nil {
		err = errors.Join(err, cerr)
	}
	if rerr := removeBuild(b.dir); rerr != nil {
		return errors.Join(err, fmt.Errorf("the unfinished run in %s could not be removed and is kept — delete it by hand: %w", b.dir, rerr))
	}
	return b.unpublished(err)
}

// NO TEXT THAT OUTLIVES THE BUILD NAMES IT. Text built while the run is built embeds the run's path,
// which is the build's; once the build is published or removed that path names nothing. The rewrite
// replaces the build's absolute path by --to's: an exact prefix, unique by construction since the
// build's name carries a random suffix. It depends on every path under the build being built by
// filepath.Join or filepath.Abs — never resolved through symlinks, URL-escaped (the record's `file:`
// DSN appears in no error) or %q-quoted. A build that is KEPT is never rewritten out of text that
// reaches the operator, so the leftover stays named.
func (b *build) atTo(s string) string { return strings.ReplaceAll(s, b.dir, b.to) }

// unpublished rewrites an error from a build that was removed: a path it names under --to never
// existed there, and the error says so.
func (b *build) unpublished(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if !strings.Contains(msg, b.dir) {
		return err
	}
	return &atToError{err: err, msg: b.atTo(msg) + " (a path under " + b.to +
		" here belongs to the run that was being built for it; nothing was published there)"}
}

// atToError is an error whose text names --to where the original named the build. Unwrap returns
// the original, so errors.Is, errors.As and feov.CodeOf answer as they would have.
type atToError struct {
	err error
	msg string
}

func (e *atToError) Error() string { return e.msg }
func (e *atToError) Unwrap() error { return e.err }
