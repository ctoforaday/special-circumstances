// Package statefile is one implementation of "a small JSON record that several
// processes read and write".
//
// It exists because there were two, and they had already diverged. `freshness` and
// `stopnudge` each grew a temp-file-plus-rename writer and a matching reader, hours apart
// and for the same reason — and only one of them learned, from CI, that a concurrent
// reader on Windows can transiently fail to open a file being renamed over. The fix went
// where the red was. The other copy is still missing it.
//
// That is the argument for extraction here and it is narrow: an atomic write is ONE idea,
// and the copies disagreed about nothing except how carefully they were written. It is
// deliberately NOT the argument for unifying the nine hookInput structs, which disagree
// about what their events actually carry and must keep disagreeing.
//
// What this package does NOT decide is what a failed read MEANS. `freshness` refuses to
// re-stamp its baseline; `stopnudge` suppresses an emission. Those are real policy
// differences and they stay with the callers — this returns the fact and lets them act.
package statefile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Status is what a read established. It is a tri-state because the two failures mean
// opposite things: "nothing has been written yet" invites writing, and "something is
// there and I cannot read it" forbids it. A bool would collapse them, and collapsing them
// is how a reader overwrites a record it could not see.
type Status int

const (
	// Absent: no file. An honest empty state — the caller may write.
	Absent Status = iota
	// Present: read and decoded.
	Present
	// Unreadable: the file exists and could not be read or decoded. NOT empty.
	Unreadable
)

func (s Status) String() string {
	switch s {
	case Absent:
		return "absent"
	case Present:
		return "present"
	}
	return "unreadable"
}

// readAttempts and readBackoff absorb the transient cases measured on Windows, and there
// are TWO of them. A concurrent open can hit a sharing violation while a rename lands: CI
// saw 21 of 800 reads fail that way under four concurrent writers, and zero on Linux. It
// can ALSO observe the name as not existing at all for an instant while the replace binds
// — measured once the CI temp dir moved to a RAM disk, where 800 renames land in 0.38s and
// the window is sampled often enough to see: 3 of 800 reads came back ErrNotExist on a
// file that existed before and after.
//
// That second case is why ErrNotExist is RETRIED rather than trusted. It is the dangerous
// one: a sharing violation reports Unreadable, which forbids a write, while a premature
// Absent invites the caller to overwrite a record it simply failed to see.
//
// Vars, not consts, so a test can widen the window deterministically.
var (
	readAttempts = 3
	readBackoff  = 2 * time.Millisecond
)

// Read loads a JSON record, distinguishing absent from unreadable.
func Read[T any](path string) (T, Status) {
	var zero T
	b, err := ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Still missing after every retry, so the absence is the file's own and not a
		// rename passing through.
		return zero, Absent
	case err != nil:
		return zero, Unreadable
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		// Present and undecodable. Retrying will not help — a torn write has already
		// landed or the file is corrupt — and reporting it as empty would invite the
		// caller to overwrite it.
		return zero, Unreadable
	}
	return v, Present
}

// ReadFile is os.ReadFile with this package's transient-miss retry, exported for the
// callers that read these records through somebody else's signature.
//
// checkpoint.LoadRearm takes its reader as an argument, and the one caller that reads the
// re-arm record OUTSIDE the re-arm lock (checkpointrestore) was passing os.ReadFile bare
// — so it could catch the same instant a rename was binding and conclude the record was
// empty, which is the reading LoadRearm's own doc calls the symptom that cost #165 three
// sessions. Giving it this reader closes that without moving the lock.
//
// The last error is the one returned, so a caller can still tell absent from unreadable.
func ReadFile(path string) ([]byte, error) {
	var err error
	for attempt := range readAttempts {
		var b []byte
		if b, err = os.ReadFile(path); err == nil {
			return b, nil
		}
		if attempt < readAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * readBackoff)
		}
	}
	return nil, err
}

// Write replaces the record ATOMICALLY: temp file, then rename.
//
// Several binaries write these files, and the client runs an event's hooks in PARALLEL
// (measured, hook-surface-spike.md §4b). A plain os.WriteFile truncates and then writes,
// so a concurrent reader sees an existing empty file and decodes the zero value — which
// is indistinguishable from "nothing written yet" unless somebody is looking for it.
//
// rename(2) is atomic within a directory, so a reader sees the old file or the new one.
// It does not ORDER two writers — last writer wins — and that is acceptable wherever both
// are recording a reading of the same thing. A caller for which it is not acceptable
// needs a lock, not this.
func Write[T any](path string, v T) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".statefile-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		_ = os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
		return err
	}
	return nil
}
