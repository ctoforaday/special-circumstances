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

// readAttempts and readBackoff absorb the transient case measured on Windows: rename is
// atomic for the FILE, but a concurrent open can still hit a sharing violation while it
// lands. CI saw 21 of 800 reads fail that way under four concurrent writers, and zero on
// Linux. A few milliseconds is far below any hook's budget and far above the window.
const (
	readAttempts = 3
	readBackoff  = 2 * time.Millisecond
)

// Read loads a JSON record, distinguishing absent from unreadable.
func Read[T any](path string) (T, Status) {
	var zero T
	for attempt := range readAttempts {
		b, err := os.ReadFile(path)
		switch {
		case errors.Is(err, os.ErrNotExist):
			return zero, Absent
		case err != nil:
			time.Sleep(time.Duration(attempt+1) * readBackoff)
			continue
		}
		var v T
		if err := json.Unmarshal(b, &v); err != nil {
			// Present and undecodable. Retrying will not help — a torn write has
			// already landed or the file is corrupt — and reporting it as empty would
			// invite the caller to overwrite it.
			return zero, Unreadable
		}
		return v, Present
	}
	return zero, Unreadable
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
