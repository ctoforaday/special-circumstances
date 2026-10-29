package record

import (
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// RefuseRecordFile refuses a path a seat named, before this process reads it, when that path is
// one of the run's record files — record.db, its -wal or its -shm, reached directly, through a
// symlink, or by a hard link.
//
// A READ OF THE RECORD BESIDE ITS OWN CONNECTION RELEASES THE CONNECTION'S LOCKS. Under POSIX fcntl
// locking, closing any descriptor to a file drops every lock the process holds on it
// (recordsql.SnapshotInto's note), and the connection carries on as if it held them: the act that
// read the file then writes through a record another seat may judge unused, recover, or truncate.
// On Linux the database file's own locks are OFD locks and survive the close (recordsql's ofd.go);
// the -shm locks, and every lock on macOS and the BSDs, stay POSIX, so the refusal covers all
// three files on every platform. The record is never a program or a report, so the refusal costs a seat nothing it could mean.
//
// The check compares files, not names: os.SameFile on the stat of each, so every spelling of the
// path is the same answer. A path that does not exist is not refused here — the act's own read
// says so in its own words.
func RefuseRecordFile(run Run, path, act, named string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return nil
	}
	db := filepath.Join(run.Records(), dbName)
	for _, f := range []string{db, db + "-wal", db + "-shm"} {
		rf, err := os.Stat(f)
		if err != nil {
			continue
		}
		if os.SameFile(fi, rf) {
			return feov.Errorf(feov.Validation, "%s: %s is %s — this run's record, not %s. Reading the record "+
				"from inside the tool releases the locks its own connection holds on it, and the act would then "+
				"write through a record other seats may judge unused. Name the file itself",
				act, path, filepath.Base(f), named)
		}
	}
	return nil
}
