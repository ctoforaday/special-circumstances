package runlive

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// readSharingDelete is os.ReadFile through openSharingDelete: the read never stops a writer
// removing or replacing the file under it.
func readSharingDelete(path string) ([]byte, error) {
	f, err := openSharingDelete(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// markerLock is the marker's lock file: frank-exchange-of-views's runlive.MarkerLock, which its
// writers hold exclusive from their read to their rename. That module cannot be imported here, so
// the name is restated.
const markerLock = "run-live.json.lock"

// readMarker reads .claude/run-live.json holding markerLock shared, as frank-exchange-of-views's
// own readers do. A rename is not atomic to a concurrent open on Windows: an open landing while the
// writer replaces the marker finds no file, and a guard reading that as "no run is live" fails
// open mid-run. Under the lock no open overlaps a rename. No lock file means no writer has written
// the marker; the reader never creates it.
func readMarker(projectDir string) ([]byte, error) {
	dir := filepath.Join(projectDir, ".claude")
	lf, err := os.Open(filepath.Join(dir, markerLock))
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return nil, err
	default:
		defer lf.Close() // closing the handle releases the lock
		if err := lockFile(lf, false); err != nil {
			return nil, err
		}
	}
	return readSharingDelete(filepath.Join(dir, "run-live.json"))
}
