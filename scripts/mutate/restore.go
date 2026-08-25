package main

import (
	"fmt"
	"os"
	"sync"
)

// restorer holds the ONE file a sweep has mutated, so it can be put back.
//
// This tool's defining act is writing a deliberate defect into checked-in source and then
// undoing it. The undo is the whole safety property — the file header says it plainly: a
// mutant left in the working tree is worse than no measurement. Everything here exists to
// make that undo both RACE-FREE and ANSWERABLE.
//
// # Why a mutex
//
// The armed file was two package-locals assigned by the mutation loop and read by the SIGINT
// goroutine, unsynchronised. That is a data race in the plain sense — and the damaging kind,
// not the theoretical kind: an interrupt landing between the two assignments can pair one
// file's path with another file's bytes and write the wrong source over a real file. The
// sibling tool solved the same problem correctly (scripts/golden/interrupt.go guards its
// signal state with a mutex), so this is that pattern brought one directory over rather than
// a new invention.
//
// # Why restore reports
//
// The restore's error was discarded, and the interrupt handler then printed "interrupted —
// file restored" whether or not it had been. That is the repo's own most-feared shape: the
// failure folds into the success message, and the reassuring sentence is the only thing the
// operator sees. A restore that fails leaves a MUTANT in a tracked file, so the one moment
// it matters most is the one moment the old code was silent.
type restorer struct {
	mu   sync.Mutex
	path string
	body []byte
}

// arm records the file about to be mutated and the bytes that put it back.
func (r *restorer) arm(path string, body []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path, r.body = path, body
}

// restore rewrites the armed file with its original bytes and disarms.
//
// It returns the path it acted on so a caller can NAME the file that still holds a mutant —
// "restore failed" without a path sends someone hunting through a whole module. An empty
// path with a nil error means there was nothing armed, which is the honest "nothing to do"
// and is distinguishable from a successful restore.
//
// On failure the file stays ARMED: a later attempt (the deferred net, or the operator
// retrying) tries again rather than forgetting which file is dirty.
func (r *restorer) restore() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.path == "" {
		return "", nil
	}
	if err := os.WriteFile(r.path, r.body, 0o644); err != nil {
		return r.path, err
	}
	path := r.path
	r.path, r.body = "", nil
	return path, nil
}

// restore is restorer.restore with the message every caller would otherwise write itself.
//
// The failure has to say three things or it is not actionable: that the restore failed, WHICH
// file is dirty, and the one command that cleans it. "interrupted — file restored" said none
// of them, and said them cheerfully.
func restore(r *restorer) (string, error) {
	p, err := r.restore()
	if err != nil {
		return p, fmt.Errorf("RESTORE FAILED for %s: %w — that file STILL HOLDS A MUTANT; run `git checkout -- %s` before committing", p, err, p)
	}
	return p, nil
}
