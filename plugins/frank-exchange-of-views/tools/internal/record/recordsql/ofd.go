package recordsql

import (
	"errors"
	"fmt"

	"modernc.org/sqlite"
)

// THE RECORD'S FILE LOCKS BELONG TO THE CONNECTION, NOT TO THE PROCESS.
//
// SQLite's default unix locks are POSIX fcntl record locks, owned by the (process, inode) pair: the
// kernel drops every one the process holds on a file at ANY close(2) of ANY descriptor on it. An
// idle WAL connection holds a SHARED lock on record.db, and that lock is what another process's
// last-closer check reads before it checkpoints and resets the log. An os.Open/Close of record.db
// anywhere in the process — a hash, a copy, a read of a path a seat named — strips it silently, the
// connection keeps answering, and the other process truncates a log this one still writes to: rows
// committed through it vanish with nothing reporting it.
//
// An open-file-description (OFD) lock belongs to the descriptor SQLite opened and survives every
// other close. modernc's switch is process-wide and is fixed at the first lock attempt, so it is
// thrown here, in init: every binary that links the driver links this package (the
// TestEveryDriverImporterLinksTheLockSwitch gate), and init runs before any of their opens,
// including the ones that do not go through Open (migrate's sqlitesource).
//
// What it covers is the database file's locks: SHARED, RESERVED, PENDING and EXCLUSIVE, which in
// WAL mode are the connection's standing SHARED lock and the EXCLUSIVE lock a last closer or an
// exclusive-mode opener takes. The WAL coordination locks in record.db-shm (the writer lock, the
// read marks, the checkpointer and the DMS lock) stay POSIX in this library version, so a stray
// close of record.db-shm still drops them; recordfileguard keeps refusing every spelling of the
// record for that reason.
//
// OFD and POSIX locks on one file conflict with each other in the kernel, so a binary that locks
// POSIX and one that locks OFD exclude each other on a shared record
// (TestPosixAndOFDLocksExcludeEachOther). OFD locks exist on Linux only: elsewhere the library
// reports ErrOFDLockingUnavailable and changes nothing — POSIX locks on macOS and the BSDs, the
// Windows VFS's own handle locks on Windows; on a
// Linux kernel or filesystem that rejects them, the first lock falls back to POSIX for the rest of
// the process. Neither is an error: the record keeps the platform's default locks.
func init() { lockingErr = useOFDLocks() }

// lockingErr is the switch's refusal, held so that every open reports it: a process whose mode was
// fixed before this package's init would otherwise lock POSIX while every reader of this file
// believes it locks OFD.
var lockingErr error

func useOFDLocks() error {
	_, err := sqlite.OFDLocking(true)
	if err == nil || errors.Is(err, sqlite.ErrOFDLockingUnavailable) {
		return nil
	}
	return fmt.Errorf("opening the record: SQLite's lock mode could not be set to OFD locks: %w", err)
}
