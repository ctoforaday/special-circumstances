package record

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

// Durability and abort-safety.
//
// A seat's records are the only surviving account of what it did. Two ways they
// were losable before this file existed:
//
//  1. NO FSYNC. os.WriteFile + rename is atomic with respect to a CONCURRENT
//     READER, but not with respect to a CRASH: the rename can reach disk before
//     the bytes do, leaving a projection that is present, named correctly, and
//     empty. Same for an appended event — the shard grows, the content is zeroes.
//     Every durable write here is now write -> fsync(file) -> rename ->
//     fsync(dir), which is the only ordering that survives power loss.
//
//  2. NO SIGNAL DISCIPLINE. Seats are killed routinely — quota aborts, Ctrl-C, a
//     harness timeout, a workflow cancel. A SIGINT landing between "open shard for
//     append" and "write the line" truncates the record; one landing while a lock
//     is held leaves a .lock- directory that blocks the next seat for the full
//     bounded wait. Writes are now bracketed in CRITICAL SECTIONS: a signal sets a
//     shutdown flag, in-flight sections are allowed to FINISH, held locks are
//     released, and only then does the process exit.
//
// The cost of an interrupted seat is now a missing event, never a corrupt one.

var (
	guardMu      sync.Mutex
	guardCond    = sync.NewCond(&guardMu)
	guardDepth   int
	shuttingDown bool
	heldLocks    sync.Map // lock directory path -> struct{}
)

// InstallSignalGuard arms interrupt handling for the process. It returns a stop
// function for tests and for callers that want the handler removed.
//
// Signals handled: SIGINT and SIGTERM. Windows delivers Ctrl-C as SIGINT through
// the same mechanism; SIGKILL and a hard power loss are unhandleable by anyone,
// which is why the fsync ordering above has to stand on its own.
func InstallSignalGuard() (stop func()) {
	ch := make(chan os.Signal, 4)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-ch:
		case <-done:
			return
		}
		guardMu.Lock()
		shuttingDown = true
		for guardDepth > 0 { // let in-flight writes finish; they are bounded and short
			guardCond.Wait()
		}
		guardMu.Unlock()
		releaseHeldLocks()
		os.Exit(130) // 128 + SIGINT, the shell convention
	}()
	return func() {
		signal.Stop(ch)
		close(done)
	}
}

// enterCritical opens a section that a signal must not interrupt.
//
// Once shutdown has begun this BLOCKS rather than proceeding: the process is
// exiting as soon as the in-flight sections drain, so starting a new write here
// would be starting one we might not be allowed to finish — the exact truncation
// this guard exists to prevent. Blocked callers never increment the depth, so
// they cannot delay the exit they are waiting on.
func enterCritical() {
	guardMu.Lock()
	defer guardMu.Unlock()
	for shuttingDown {
		guardCond.Wait()
	}
	guardDepth++
}

func exitCritical() {
	guardMu.Lock()
	defer guardMu.Unlock()
	guardDepth--
	if guardDepth <= 0 {
		guardCond.Broadcast()
	}
}

func releaseHeldLocks() {
	heldLocks.Range(func(k, _ any) bool {
		os.Remove(k.(string))
		return true
	})
}

// fsyncFile flushes a file's contents to stable storage.
func fsyncFile(f *os.File) error { return f.Sync() }

// fsyncDir flushes a DIRECTORY entry, which is what makes a rename durable — the
// renamed name is metadata, and syncing the file does not sync the name.
//
// Windows has no directory-handle sync: opening a directory for this purpose
// fails, and NTFS commits the rename with the metadata transaction anyway, so the
// error is expected there and deliberately ignored rather than propagated as a
// write failure.
func fsyncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

// durableWrite writes content to path via a temp file, fsyncing before and after
// the rename. The whole sequence is one critical section.
func durableWrite(path string, content []byte, tmp string) error {
	enterCritical()
	defer exitCritical()
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := fsyncFile(f); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := renameWithRetry(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return fsyncDir(filepath.Dir(path))
}

// durableAppend appends one line to a shard and fsyncs it before returning, so a
// seat that reports success has a record that survives the machine dying.
func durableAppend(shard, line string) error {
	enterCritical()
	defer exitCritical()
	f, err := os.OpenFile(shard, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(line); err != nil {
		f.Close()
		return err
	}
	if err := fsyncFile(f); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return fsyncDir(filepath.Dir(shard))
}
