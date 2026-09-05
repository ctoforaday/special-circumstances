// Package sittinghook is the SubagentStart/SubagentStop hook logic: decide whether this event is a
// seat's sitting in a live run, and if so hand the write to a separate process.
//
// IT LINKS NOTHING EXPENSIVE, and that is the whole design rather than a detail. These hooks fire
// far more often than they write:
//
//	SubagentStop fires at the MAIN AGENT'S TURN END as well as at a seat's return — 19 seats
//	against 50 turn ends in one measured session (plans/hook-surface-spike.md §7a) — and it fires
//	in EVERY session, including every session with no run at all.
//
// A hook that linked internal/record would pay a SQLite driver's init() and every protobuf
// descriptor on all of those, to discover it had nothing to write. Measured on this binary: 3.555
// ms and 13.06 MB with the record linked, against 1.189 ms and 2.94 MB without. That is #684 F2's
// defect exactly, on a different event, and #734 had just finished removing it from the PreToolUse
// hook when this was written.
//
// So the cheap facts are checked here — is there an agent type, is there a run — and the expensive
// process is spawned only once both are true. Roughly 38 times in a run, and never in a session
// that is not running a debate.
package sittinghook

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
)

// writerName is the binary that owns the write. It sits beside this hook in the plugin's bin
// directory, which is how it is found — the run's own wrapper would work too, but it exists only
// once `setup` has run, and a hook must behave the same before and after that.
const writerName = "feov-sitting-write"

// THE PHASE STRINGS ARE REPEATED HERE RATHER THAN IMPORTED, and that is deliberate in a way worth
// stating, because it looks like the duplication this codebase refuses everywhere else.
//
// sittingwrite owns these values. Importing them costs this package sittingwrite's entire graph —
// internal/record, a SQLite driver, every protobuf descriptor — for two string constants that are
// never read back here, only passed on the command line. That is the #684 F2 defect in miniature:
// nobody adds a heavy dependency, they add a convenient one, and the weight arrives behind it.
// Measured on this very binary during this change: importing the type took it from 1.189 ms and
// 2.94 MB to 3.555 ms and 13.06 MB, on an event that fires at every main-agent turn end in every
// session.
//
// The values cannot drift silently: TestThePhaseStringsMatchTheWriters asserts them against
// sittingwrite's own, from the test binary, where the heavy import costs nothing.
const (
	phaseOpen  = "open"
	phaseClose = "close"
)

// sittingInput is the subset of the SubagentStart/SubagentStop payload this needs.
type sittingInput struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Cwd       string `json:"cwd"`
}

// Start records the moment the harness dispatched an agent.
//
// IT WRITES NOTHING TO STDOUT, and that is load-bearing rather than minimal. §10 of the hook
// surface spike measured what an emission costs on the sibling event: a SubagentStop hook
// returning additionalContext re-invoked the seat, its turn ended, the hook fired again — NINE
// firings for one seat, the returned context discarded every time. Under a log-only hook the same
// launch fires exactly once. An observation hook that starts talking turns one event into nine.
func Start(stdin io.Reader, stdout io.Writer) error { return handoff(stdin, phaseOpen) }

// Stop records the moment that agent returned. Silent for the reason above — and here the
// measurement is of this very event rather than an analogy to it.
func Stop(stdin io.Reader, stdout io.Writer) error { return handoff(stdin, phaseClose) }

// handoff is the decision. Everything it rejects, it rejects BEFORE spawning anything.
//
// NOTHING HERE CAN FAIL THE HOOK. A hook's job is to observe; a seat is not blocked because the
// bookkeeping failed, and an error returned from here would reach the harness as a failed hook on
// an event the seat cannot even see.
func handoff(stdin io.Reader, phase string) error {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return nil
	}
	var in sittingInput
	if json.Unmarshal(raw, &in) != nil {
		return nil
	}
	// NOT A SEAT, AND THIS IS THE FILTER THE FREQUENCY ARGUMENT RESTS ON. Both halves are
	// required and for different reasons: no agent id means nothing to join a span to, and no
	// agent type means this is a main-agent turn boundary rather than a sitting. Without it a
	// run's sitting count reads about 3.6x its seat count, every extra one a turn end wearing a
	// seat's shape — and every one of them would have spawned a writer.
	if in.AgentID == "" || in.AgentType == "" {
		return nil
	}
	runDir := runlive.InferRunDir(in.Cwd)
	if runDir == "" {
		return nil
	}
	writer := writerPath()
	if writer == "" {
		return nil
	}
	spawn(writer, runDir, phase, in.AgentID, in.AgentType)
	return nil
}

// spawn is a variable so the DECISION can be tested without a built writer on disk. What matters
// about this function is which events reach it and with what — that a turn end never does, that a
// session with no run never does — and asserting that through a real subprocess would test the
// exec plumbing instead of the filter.
var spawn = func(writer, runDir, phase, agentID, agentType string) {
	// WAITED ON, not fired and forgotten: a detached child can be killed when the hook process
	// exits, and a span silently missing one end is worse than a hook that took another
	// millisecond. Its failure is deliberately discarded — it reports to stderr, and the hook's
	// contract is to stay silent whatever happened.
	_ = exec.Command(writer,
		"-run", runDir,
		"-phase", phase,
		"-agent-id", agentID,
		"-agent-type", agentType,
	).Run()
}

// writerPath locates the writer beside this executable, or "" when it cannot be found — which is
// the bootstrap window before `doctor --fix` has built the binaries, and is a silence rather than
// an error for the same reason the shell guard in hooks.json is.
func writerPath() string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}
	p := filepath.Join(filepath.Dir(self), writerName)
	if runtime.GOOS == "windows" {
		p += ".exe"
	}
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}
