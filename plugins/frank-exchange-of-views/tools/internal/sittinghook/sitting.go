// Package sittinghook is the SubagentStart/SubagentStop hook logic: decide whether this event is a
// seat's sitting in a live run, and if so hand the write to a separate process — and, on the opening
// end, deliver to the seat what that process rendered for it.
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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

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
	phaseLimit = "limit"
)

// Limit hands a sitting that has reached the run's tool-call limit to the writer, which puts it on
// the record. The PreToolUse hook calls it once per such sitting (sittingcap decides which call).
//
// UNLIKE THE SPAN ENDS, A MISSING WRITER IS REPORTED. The seat is being refused either way; what a
// missing writer loses is the record's account of why, and that absence should reach the hook log
// rather than read as a sitting that never hit the limit.
func Limit(runDir, agentID, agentType string, sitting, limit int) error {
	writer := writerPath()
	if writer == "" {
		return fmt.Errorf("%s is not beside this hook, so %s sitting %d reaching the limit of %d is not on the record",
			writerName, agentID, sitting, limit)
	}
	return spawnLimit(writer, runDir, agentID, agentType, sitting, limit)
}

// spawnLimit is a variable for the reason spawn is.
var spawnLimit = func(writer, runDir, agentID, agentType string, sitting, limit int) error {
	args := []string{
		"-run", runDir,
		"-phase", phaseLimit,
		"-agent-id", agentID,
		"-sitting", strconv.Itoa(sitting),
		"-limit", strconv.Itoa(limit),
	}
	if agentType != "" {
		args = append(args, "-agent-type", agentType)
	}
	if out, err := exec.Command(writer, args...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %v: %s", writerName, err, out)
	}
	return nil
}

// sittingInput is the subset of the SubagentStart/SubagentStop payload this needs.
type sittingInput struct {
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Cwd       string `json:"cwd"`
	// AgentTranscriptPath is the finished seat's own transcript, resolved by the harness. It
	// arrives on SubagentStop and is empty on SubagentStart, which is correct: a seat that has
	// just been dispatched has produced no turns to read.
	//
	// PASSED THROUGH, NEVER READ HERE. This package's whole design is that it links nothing
	// expensive; parsing a transcript is the writer's job, in the process that already carries
	// the record.
	AgentTranscriptPath string `json:"agent_transcript_path"`
	// WHERE THE SEAT SPOKE, both carried on every payload of both events and recorded by neither
	// until now. SessionID names the conversation the sitting happened in; PromptID names the
	// DISPATCH, and changes when an agent is re-prompted while its agent id does not.
	//
	// READ HERE AND NOT DERIVED LATER. A reader reconstructing the conversation from the filesystem
	// has to guess between two layouts — `subagents/` for a plain subagent and
	// `subagents/workflows/wf_<id>/` under the Workflow tool — and a guess that misses returns no
	// trajectory, which reads exactly like a seat that never ran.
	SessionID string `json:"session_id"`
	PromptID  string `json:"prompt_id"`
}

// Start records the moment the harness dispatched an agent.
//
// IT SPEAKS TO THE SEAT, and it is the only event in this plugin that may. The nine-firing loop §10
// of the hook surface spike measured is SubagentStop's: an emission there re-invoked the seat, its
// turn ended, the hook fired again, and the returned context was discarded every time. THIS EVENT IS
// THE OPPOSITE RESULT in the same section — one firing, and the marker arrives in the SEAT's own
// context. Verified three times: #500 and #507, and again 2026-09-25, where the injected marker
// landed as a hook_additional_context attachment on the seat's transcript and the seat returned it
// verbatim.
//
// WHAT IT SAYS IS THE SEAT'S WORK LIST (#1122), rendered by the writer this hands off to, and passed
// through here verbatim. Nothing in this package composes that text: a seat could not learn it owed
// nothing without asking, so the cheapest empty sitting was one call and never zero.
func Start(stdin io.Reader, stdout io.Writer, rec *hookfailures.Recorder) error {
	return handoff(stdin, phaseOpen, stdout, rec)
}

// Stop records the moment that agent returned, and MUST stay silent — here the measurement is of
// this very event: nine firings for one seat, nothing delivered anywhere.
func Stop(stdin io.Reader, stdout io.Writer, rec *hookfailures.Recorder) error {
	// NO WRITER PASSED, WHICH IS THE OUTER OF TWO REFUSALS. handoff cannot emit for an event that
	// handed it nowhere to write, and it also refuses on the phase — see the emission site.
	return handoff(stdin, phaseClose, nil, rec)
}

// The stages a sitting hook can fail at. NEITHER of its events displays anything TO A HUMAN, which
// is a different question from whether either may speak to the SEAT: SubagentStart may and
// SubagentStop may not. A failure here therefore waits on the record for FEOV's only displaying
// event — PreToolUse — which a seat's very next tool call fires.
const (
	StageInput         hookfailures.Stage = "sitting-input"
	StageWriterMissing hookfailures.Stage = "sitting-writer-missing"
	StageWrite         hookfailures.Stage = "sitting-write"
)

// handoff is the decision. Everything it rejects, it rejects BEFORE spawning anything.
//
// NOTHING HERE CAN FAIL THE HOOK. A hook's job is to observe; a seat is not blocked because the
// bookkeeping failed, and an error returned from here would reach the harness as a failed hook on
// an event the seat cannot even see.
func handoff(stdin io.Reader, phase string, seat io.Writer, rec *hookfailures.Recorder) error {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		rec.Fail(StageInput, "cannot read the hook payload: "+err.Error()+" — this sitting is not on the record")
		return nil
	}
	var in sittingInput
	if err := json.Unmarshal(raw, &in); err != nil {
		rec.Fail(StageInput, "the hook payload does not parse: "+err.Error()+" — this sitting is not on the record")
		return nil
	}
	rec.OK(StageInput)
	// NOT A SEAT, AND THIS IS THE FILTER THE FREQUENCY ARGUMENT RESTS ON. Both halves are
	// required and for different reasons: no agent id means nothing to join a span to, and no
	// agent type means this is a main-agent turn boundary rather than a sitting. Without it a
	// run's sitting count reads about 3.6x its seat count, every extra one a turn end wearing a
	// seat's shape — and every one of them would have spawned a writer.
	if in.AgentID == "" || in.AgentType == "" {
		return nil
	}
	inferred := runlive.InferRunDir(in.Cwd)
	if inferred.Why.Fault() {
		// A marker that cannot be used is not "no run": it is a run whose seats' sittings are all
		// being dropped. Only the two fault shapes are recorded — no marker, no open run and two
		// runs open are the ordinary shapes of a session, and stay silent.
		rec.FailIn(runlive.StageUnusable, inferred.MarkerDir, unusableDetail(inferred))
		return nil
	}
	if inferred.MarkerDir != "" {
		rec.OKIn(runlive.StageUnusable, inferred.MarkerDir)
	}
	if inferred.Dir == "" {
		return nil
	}
	writer := writerPath()
	if writer == "" {
		// Limit already reported this case; the span ends did not, and every sitting of the run
		// was lost with nothing said. It is the bootstrap window only until the fetch lands, and
		// the entry clears on the first sitting written after that — so a transient miss is shown
		// once, and a permanent one keeps showing.
		rec.Fail(StageWriterMissing, writerName+" is not beside this hook, so "+in.AgentID+"'s sitting is not on the record")
		return nil
	}
	rec.OK(StageWriterMissing)
	forSeat, err := spawn(writer, inferred.Dir, phase, in.AgentID, in.AgentType, in.AgentTranscriptPath, in.SessionID, in.PromptID)
	if err != nil {
		rec.FailIn(StageWrite, inferred.MarkerDir, err.Error())
		return nil
	}
	rec.OKIn(StageWrite, inferred.MarkerDir)
	// THE WRITER'S STDOUT IS THE SEAT'S WORK LIST, and it is passed through verbatim. This hook links
	// nothing that can read a record (see the package comment and the hookgraph allowlist), so the
	// projection is rendered in the process that already carries it and this one only delivers.
	//
	// NOTHING IS EMITTED FOR AN EMPTY PAYLOAD. A seat whose configuration seats several — blue's —
	// has no list to be handed, and an empty additionalContext would be a document saying nothing.
	//
	// THE PHASE IS CHECKED HERE AS WELL AS IN THE WRITER, and that is defence in depth rather than a
	// duplicated branch. Stop's silence is a contract whose breach costs nine firings of an event, and
	// the only thing enforcing it was a phase switch in a different process: a writer that started
	// printing anything on close — a diagnostic moved to stdout, a future phase — would have re-armed
	// the seat, and nothing here would have refused it.
	if phase == phaseOpen && seat != nil && len(strings.TrimSpace(string(forSeat))) > 0 {
		emitForSeat(seat, string(forSeat))
	}
	return nil
}

// emitForSeat writes the one document SubagentStart may return: additionalContext for the subagent
// that was just dispatched.
//
// MARSHALLED, NEVER FORMATTED. The work list is JSON inside a JSON string, and every quote, newline
// and backslash in it has to survive; a Sprintf of this shape would corrupt the first list that
// carried a quoted claim, which is every list that carries a gap location.
func emitForSeat(stdout io.Writer, context string) {
	type hookSpecific struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	}
	b, err := json.Marshal(struct {
		HookSpecificOutput hookSpecific `json:"hookSpecificOutput"`
	}{hookSpecific{HookEventName: "SubagentStart", AdditionalContext: context}})
	if err != nil {
		// Unreachable for two strings, and silence is the right failure anyway: a malformed document
		// on this event is worse than no document, because the seat can still ask for its list.
		return
	}
	stdout.Write(append(b, '\n'))
}

// unusableDetail says which fault the inference hit, because each is fixed differently.
func unusableDetail(i runlive.Inferred) string {
	return runlive.FaultDetail(i) + " — seats' sittings are not being recorded"
}

// spawn is a variable so the DECISION can be tested without a built writer on disk. What matters
// about this function is which events reach it and with what — that a turn end never does, that a
// session with no run never does — and asserting that through a real subprocess would test the
// exec plumbing instead of the filter.
var spawn = func(writer, runDir, phase, agentID, agentType, transcript, sessionID, promptID string) ([]byte, error) {
	// WAITED ON, not fired and forgotten: a detached child can be killed when the hook process
	// exits, and a span silently missing one end is worse than a hook that took another
	// millisecond.
	//
	// ITS OUTPUT IS KEPT. This comment used to say the child's failure "reports to stderr" — and
	// Run() leaves exec.Cmd.Stderr nil, which Go connects to os.DevNull, so everything the writer
	// said was destroyed. The sibling Limit path always used CombinedOutput and kept it; the two
	// disagreed about whether the same writer's failure was worth reading. It is: the caller records
	// it, and the hook still writes nothing to its own stdout.
	args := []string{
		"-run", runDir,
		"-phase", phase,
		"-agent-id", agentID,
		"-agent-type", agentType,
	}
	// ONLY WHEN THERE IS ONE. SubagentStart carries no transcript, and an empty flag would make
	// the writer distinguish "not sent" from "sent empty" for no reason. The same rule for the
	// conversation fields: absent means the payload did not carry it, which is a real answer.
	if transcript != "" {
		args = append(args, "-transcript", transcript)
	}
	if sessionID != "" {
		args = append(args, "-session-id", sessionID)
	}
	if promptID != "" {
		args = append(args, "-prompt-id", promptID)
	}
	// STDOUT AND STDERR ARE SPLIT, and that is the change that makes this a delivery channel. They
	// were merged by CombinedOutput, which was right while the output was only ever a failure
	// report: now stdout is the seat's work list and stderr is the writer's diagnostics, and one
	// stray diagnostic line merged into the other would arrive in a seat's context as its work.
	cmd := exec.Command(writer, args...)
	var forSeat, diagnostics bytes.Buffer
	cmd.Stdout, cmd.Stderr = &forSeat, &diagnostics
	if err := cmd.Run(); err != nil {
		// BOTH STREAMS ON THE FAILURE PATH. The writer reports its faults on stderr, and anything it
		// had begun writing for the seat is evidence about the fault rather than a list.
		return nil, fmt.Errorf("%s: %v: %s", writerName, err,
			strings.TrimSpace(diagnostics.String()+" "+forSeat.String()))
	}
	return forSeat.Bytes(), nil
}

// writerFileName is the writer's name on this platform. It exists so the test that places a stub
// beside the test binary spells the name the same way this does — the two spelled it separately
// once, and the Windows leg failed while Linux passed, because only one of them added `.exe`.
func writerFileName() string {
	if runtime.GOOS == "windows" {
		return writerName + ".exe"
	}
	return writerName
}

// writerPath locates the writer beside this executable, or "" when it cannot be found — which is
// the bootstrap window before the hooks' fetch (or `doctor --fix`) has installed the binaries, and is a silence rather than
// an error for the same reason the shell guard in hooks.json is.
func writerPath() string {
	self, err := os.Executable()
	if err != nil {
		return ""
	}
	p := filepath.Join(filepath.Dir(self), writerFileName())
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}
