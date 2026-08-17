package main

// THE TREE IS THE CONTRACT, AND A SIGNAL IS NOT AN EXCEPTION TO IT (#432 friction 4).
//
// review materialises proposals into the working tree and reverts everything it was not told to
// keep. Between those two moments the tree holds writes the tool has PROMISED TO TAKE BACK — and
// until this file existed, any signal in that window killed the process with the promise unkept,
// while the same tool's result line says "the tree is exactly as you found it".
//
// # Measured, and it is not a hypothetical
//
//	$ golden -review 2>&1 | head -200
//	exit 141
//	$ git status --short -- '*.golden'
//	 M ...prompt-assemble.golden        (and twelve more)
//
// Thirteen rewritten goldens left behind, and the NEXT run then refuses to start — its
// precondition sees dirty goldens and cannot tell the tool's own abandoned writes from the
// author's work, which is exactly the distinction that precondition exists to protect.
//
// # Why it read as intermittent
//
// SIGPIPE's default disposition kills the process at the next write to a closed stdout. Whether
// that lands inside the dangerous window depends on the PIPE BUFFER: at 64 KiB here, a review
// with one proposal (~15 KiB of output) is written entirely into the buffer, so the run reaches
// its revert before the signal is ever delivered and the tree comes out clean. Thirteen
// proposals is ~71 KiB, the writer blocks mid-REVIEW, and the signal lands with every proposal
// on disk. Same defect both times; only the buffer decided whether it showed.
//
// # Why the handler does not restore
//
// The obvious fix — revert inside the signal handler — races the child process that is still
// writing goldens, and would give this tool two writers of the working tree. Instead the handler
// does the two things that are safe from another goroutine: record the signal, and kill the leg
// that is running. The main flow then unwinds through the SAME restore every other path uses, so
// there is exactly one implementation of "take it back" and no window where two goroutines are
// both undoing things.
//
// # What this does NOT cover, enumerated rather than left to be discovered
//
// SIGKILL, a power loss, or a container reaped mid-run cannot be caught in-process, so those can
// still strand proposals. Closing that needs a different mechanism — a marker file naming what
// this run materialised, so a later run can tell its own abandoned writes from the author's and
// offer to undo them. That is a record, and it is filed rather than half-built here (#432).

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

// interrupted is the process-wide record of a signal and of the leg currently running.
//
// Package-level because signals are process-level: there is one stdout to lose and one child at
// a time, and threading that through every call site would buy nothing but ceremony.
var interrupted struct {
	mu  sync.Mutex
	got os.Signal
	cmd *exec.Cmd
}

// watchSignals arms the handler. SIGPIPE is the one that matters — and NOTIFYING it is itself
// half the fix, because a notified SIGPIPE stops being fatal: writes to a closed stdout return
// EPIPE instead of killing the process, which is what lets the main flow reach its restore at
// all. SIGINT and SIGTERM are here because Ctrl-C during a review has exactly the same shape.
func watchSignals() {
	ch := make(chan os.Signal, 4)
	signal.Notify(ch, syscall.SIGPIPE, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for s := range ch {
			interrupted.mu.Lock()
			if interrupted.got == nil {
				interrupted.got = s
			}
			// Kill the running leg so the main flow stops waiting on a process whose output has
			// nowhere to go. Without this, `go test` writing into a closed pipe can sit there
			// long after the reviewer has walked away.
			if c := interrupted.cmd; c != nil && c.Process != nil {
				_ = c.Process.Kill()
			}
			interrupted.mu.Unlock()
		}
	}()
}

// trackLeg records the child so a signal can end it. clearLeg drops it once it has exited.
func trackLeg(c *exec.Cmd) {
	interrupted.mu.Lock()
	interrupted.cmd = c
	interrupted.mu.Unlock()
}

func clearLeg() {
	interrupted.mu.Lock()
	interrupted.cmd = nil
	interrupted.mu.Unlock()
}

// stopped reports whether a signal arrived, and the exit code to leave with.
//
// 128+signal is the shell's convention, and using it means `| head` still reports 141 — the
// caller's view of what happened does not change, only the state of their working tree does.
func stopped() (int, bool) {
	interrupted.mu.Lock()
	defer interrupted.mu.Unlock()
	if interrupted.got == nil {
		return 0, false
	}
	if s, ok := interrupted.got.(syscall.Signal); ok {
		return 128 + int(s), true
	}
	return 1, true
}

func signalName() string {
	interrupted.mu.Lock()
	defer interrupted.mu.Unlock()
	if interrupted.got == nil {
		return ""
	}
	return interrupted.got.String()
}

// ---- the single restore ----

// held is what this run has materialised and not yet decided about. It is the list the restore
// works from, so a panic, a signal, and an ordinary refusal all take the same path out.
var held struct {
	mu    sync.Mutex
	root  string
	props []proposal
}

// hold claims the proposals this run created. Anything the tool did not create is never in here
// — proposalsBetween already excludes it — so the restore cannot touch the author's work.
func hold(root string, props []proposal) {
	held.mu.Lock()
	held.root, held.props = root, props
	held.mu.Unlock()
}

// release says the decision has been applied and the remaining files are meant to stay.
func release() {
	held.mu.Lock()
	held.props = nil
	held.mu.Unlock()
}

// restoreHeld reverts whatever is still claimed and reports how many. Idempotent, so running it
// from a defer costs nothing on the paths that already released.
func restoreHeld() int {
	held.mu.Lock()
	root, props := held.root, held.props
	held.props = nil
	held.mu.Unlock()

	n := 0
	for _, p := range props {
		if err := revert(root, p); err != nil {
			// Report and keep going: one file that cannot be put back must not strand the other
			// twelve. A partial restore is worse than a full one and far better than none.
			fmt.Fprintf(os.Stderr, "golden: could not restore %s: %v\n", p.path, err)
			continue
		}
		n++
	}
	return n
}
