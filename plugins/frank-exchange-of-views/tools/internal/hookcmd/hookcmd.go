// Package hookcmd is the PreToolUse/PostToolUse backend, as its own program rather than a verb
// on the seat tool.
//
// # Why this is not a subcommand any more
//
// It was `feov-record hook pretooluse`, and hooks.json invoked it with no --seat-id because a
// hook HAS no seat id — Claude Code calls it, not a seat and not an operator. That worked only
// while "nobody said who is asking" fell through to the operator surface. When identity started
// SELECTING the surface, no-identity stopped resolving any command: the root refused the shipped
// hook's own invocation, RefuseAndTeach exits 2, and exit 2 from a PreToolUse hook DENIES the
// tool call. Matcher Write|Edit|MultiEdit|NotebookEdit|Bash — so every mutating tool call in the
// session, for anyone whose plugin bin/ held the binary. The suite stayed green because the
// change taught eight test files to pass --seat-id, and hooks.json is the one caller that cannot
// be taught.
//
// Registering the verb outside the scoped surface fixed that instance and left the shape wrong.
// A hook backend living inside an identity-scoped command tree has to be exempted from identity,
// hidden from help, and then exempted AGAIN from the surface census — CommandPaths walks hidden
// commands, so `blue hook pretooluse` became a seat capability and three gates demanded a trigger
// row and a board that drives it. Every one of those exemptions is a symptom of the same category
// error: this is not a verb, it is a program.
//
// The suite's other eleven hooks are dedicated single-purpose binaries — prosthetic-conscience's
// nine and gray-area's one — and not one of them broke. This joins them.
//
// # The exit-code contract, enforced here rather than documented downstream
//
// EVERY ENTRY POINT RETURNS 0, whatever it decides. The decision travels in the stdout JSON; the
// exit status is not a channel. That invariant used to live in a doc comment on a leaf command,
// two levels below the argument parsing that broke it. Run() enforces it at the process boundary,
// including through a panic, because a hook that dies is a session that cannot edit files.
package hookcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittinghook"
	"github.com/ctoforaday/special-circumstances/shared/buildid"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// The stages Run records for any entry point. A hook that panicked, or returned an error, did not do
// its job — and both used to be a stderr line at exit 0, which reaches the debug log and nobody.
const (
	StagePanic     hookfailures.Stage = "hook-panic"
	StageError     hookfailures.Stage = "hook-error"
	StageTurnLimit hookfailures.Stage = "turn-limit"
	StageInput     hookfailures.Stage = "pretooluse-input"
	StageToolInput hookfailures.Stage = "tool-input"
	StageDeliver   hookfailures.Stage = "deliver"
	// StageRefusalLog is the write of a denial to the log, apart from the denial: a seat refused
	// correctly whose refusal could not be recorded is a different fault from a refusal that failed.
	StageRefusalLog hookfailures.Stage = "refusal-log"
	// StageFailureLog is the PostToolUseFailure hook's one job: a seat's failed call to the log.
	StageFailureLog hookfailures.Stage = "failure-log"
	// StageEnvelope is the check of a seat's envelope against the record (refuseEnvelope): a check
	// that could not run refuses the envelope naming why, and says so here as well.
	StageEnvelope hookfailures.Stage = "envelope-check"
	// StageAdmission is the PostToolUse hook's one job: a seat's accepted envelope recorded as admitted.
	StageAdmission hookfailures.Stage = "envelope-admission"
	// StageNextItem is the PostToolUse hook's other job: what a lens is handed after a call to the record.
	StageNextItem hookfailures.Stage = "next-item"
)

// Entry is a hook entry point: it reads the payload, writes its decision document (if any) to
// stdout, and records what it could not do on rec.
type Entry func(stdin io.Reader, stdout io.Writer, rec *hookfailures.Recorder) error

// Run invokes a hook entry point and guarantees the process exit code is 0.
//
// A panic here is a bug worth fixing, and it is NOT worth denying every tool call in the session
// to report. It is swallowed deliberately — and RECORDED, because swallowed used to mean a stderr
// line nobody reads. The record is read out on PreToolUse, this plugin's only displaying event.
//
// The entry writes to a BUFFER, and Run writes the one response: the entry's document with the
// record's systemMessage added to it, or the message alone, or nothing. A second top-level object
// beside a permission document is not a response, and a decision is never risked for a message: a
// document Run cannot decode goes out exactly as the entry wrote it.
//
// SubagentStart and SubagentStop display nothing TO A HUMAN, so Settle returns "" and no
// systemMessage is added. That is not the same as muteness: SubagentStart's document is how a seat's
// work list reaches it (#1122), and only SubagentStop must stay silent outright — an emission there
// re-invokes the seat, nine firings measured.
func Run(bin, event string, f Entry, stdin io.Reader, stdout io.Writer) (code int) {
	for _, arg := range os.Args[1:] {
		if arg == "-version" || arg == "--version" {
			fmt.Fprintln(stdout, buildid.Line(bin))
			return 0
		}
	}
	rec := hookfailures.New("frank-exchange-of-views", bin, event, time.Now(), os.Stderr)
	var buf bytes.Buffer
	raw, _ := io.ReadAll(stdin)
	req := hookcore.Read(event, bytes.NewReader(raw))
	defer func() {
		if r := recover(); r != nil {
			rec.Fail(StagePanic, fmt.Sprintf("recovered from %v (exiting 0: a hook must never block on its own fault)", r))
		} else {
			rec.OK(StagePanic)
		}
		if err := respond(req, stdout, buf.Bytes(), rec.Settle()); err != nil {
			// The decision never reached the client. Recorded rather than said — the channel for
			// saying it is the thing that failed — and read out by the next call that can.
			//
			// PERSISTED, never settled: a second Settle on this displaying event would stamp the entry
			// as said while its message went nowhere, and the next call would then stay quiet about
			// it for ten minutes.
			rec.Fail(StageDeliver, "cannot write the response: "+err.Error())
			rec.Persist()
		} else {
			// Persisted AGAIN, because the Settle above ran before delivery was known: an OK recorded
			// after it would never reach the record, and a delivery failure would never clear. The
			// second pass is a load and, normally, no write.
			rec.OK(StageDeliver)
			rec.Persist()
		}
		code = 0
	}()
	if err := f(bytes.NewReader(raw), &buf, rec); err != nil {
		rec.Fail(StageError, err.Error()+" (exiting 0)")
	} else {
		rec.OK(StageError)
	}
	return 0
}

// respond writes the entry's document, with the record's message added when there is one.
func respond(req hookcore.Request, stdout io.Writer, doc []byte, msg string) error {
	if req.Engine == hookcore.EngineAntigravity {
		if msg != "" {
			hookcore.QueueSay(req.PluginDataDir, req.SessionID, msg)
		}
		if req.Event == "PostToolUse" || req.Event == "PostTool" || req.Event == "PostToolUseFailure" {
			_, err := stdout.Write([]byte("{}\n"))
			return err
		}
		if len(bytes.TrimSpace(doc)) > 0 {
			_, err := stdout.Write(doc)
			return err
		}
		return nil
	}
	if msg == "" {
		_, err := stdout.Write(doc)
		return err
	}
	if len(bytes.TrimSpace(doc)) == 0 {
		hookfailures.Emit(stdout, msg)
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(doc, &m); err != nil {
		_, werr := stdout.Write(doc) // the decision goes out intact; the message waits for the next call
		return werr
	}
	m["systemMessage"] = msg
	return json.NewEncoder(stdout).Encode(m)
}

// Pre injects the run directory and the calling agent's id into a Bash call made inside a live
// run, so a seat never mistypes the path and never loses the identity that binds it to its seat on
// the record — and it refuses two things: any tool call but the envelope's past the seat's
// per-sitting call limit (enforceLimit), and a tool command carrying a backtick the shell would run,
// which would rewrite the seat's prose before the tool saw it (hookgate/substitution.go). The call
// returning a seat's envelope goes to the record instead (gateEnvelope).
func Pre(stdin io.Reader, stdout io.Writer, rec *hookfailures.Recorder) error {
	req := hookcore.Read("PreToolUse", stdin)
	if !req.Parsed {
		// Not "nothing to do": the client always sends a payload, and one this hook cannot read
		// means no turn limit counted and no run directory injected for this call. The sitting
		// hooks record the same failure; this path, which carries both of those, did not.
		rec.Fail(StageInput, "cannot read or parse the PreToolUse payload — the turn limit and run-directory injection did not run for this call")
		return nil
	}
	rec.OK(StageInput)
	in := hookgate.Input{
		AgentID:   req.AgentID,
		AgentType: req.AgentType,
		SessionID: req.SessionID,
		ToolName:  req.ToolName,
		ToolInput: req.ToolInput,
	}
	cwd := req.CWD
	if cwd == "" {
		cwd = cwdOf(req.Raw)
	}
	// THE CALL THAT RETURNS A SEAT'S ENVELOPE ENDS ITS SITTING, so the record is asked whether it
	// refuses it, and otherwise what it says now, which the hook writes onto the call (gateEnvelope).
	// IT GOES BEFORE THE TURN LIMIT, which neither counts nor refuses it: the limit's refusal tells
	// the seat to return its envelope, and this is that call. The record's own refusal still stands.
	if in.ToolName == envelopeTool {
		gateEnvelope(in, cwd, stdout, rec, req)
		return nil
	}
	// THE TURN LIMIT covers every other tool, and comes before the injection: a refused call runs
	// nothing, so it needs nothing injected. A counting fault goes to stderr and the call proceeds.
	denied, err := enforceLimit(in, cwd, stdout, rec, req)
	if err != nil {
		// The seat is still refused when it should be — enforcement survives — but the record's
		// account of WHY does not, and that used to be a stderr line nobody read.
		rec.Fail(StageTurnLimit, "turn limit: "+err.Error())
	} else {
		rec.OK(StageTurnLimit)
	}
	if denied {
		return nil
	}
	// The gate says which calls it acts on — every Bash call, and a Read of the report file — so
	// every other tool stops here, before the marker search a Read would otherwise pay for.
	if !hookgate.Wants(in) {
		return nil
	}
	// The run directory is resolved from the payload's `cwd` — the SEAT's working directory,
	// which is wire-supplied and documented, never this hook process's os.Getwd(). Absent or
	// unusable marker → empty → no rewrite, matching InferRunDir's "say nothing rather than guess".
	inferred := runlive.InferRunDir(cwd, callerAgent(in))
	noteInference(inferred, rec)
	switch outcome, payload := hookgate.PreOutcome(in, inferred.Dir); outcome {
	case hookgate.OutcomeRewrite:
		emitPreRewrite(stdout, in.ToolInput, payload, req)
		rec.OK(StageToolInput)
	case hookgate.OutcomeDeny:
		emitPreDeny(stdout, payload, req)
		rec.OK(StageToolInput)
		// A DENIAL IS A REFUSAL, and the tool logs every refusal it gives a seat. This one is given
		// outside every verb, so it is written here or not at all.
		why, _, _ := strings.Cut(payload, ".")
		if err := recordRefusal(inferred.Dir, in.AgentID, "refused a `"+in.ToolName+"` the run does not serve: "+why); err != nil {
			rec.Fail(StageRefusalLog, "the hook refused a call and could not log the refusal: "+err.Error())
		} else {
			rec.OK(StageRefusalLog)
		}
	case hookgate.OutcomeUnparsable:
		rec.Fail(StageToolInput, "a Bash call's tool_input does not parse ("+payload+") — the run directory was not injected, and the seat will hit the missing-run refusal from another layer")
	default:
		rec.OK(StageToolInput)
	}
	return nil
}

// enforceLimit counts this call against the calling seat's sitting and refuses it when the sitting
// is past the run's per-sitting tool-call limit (internal/sittingcap).
//
// THE SEAT IS FOUND THE SAME WAY UNDER EITHER ENGINE. A Workflow seat is a subagent, and its
// payload carries agent_id. A seat run as the main session of a headless `claude -p` process has
// no agent_id on the payload; its identity and run are in the process environment this hook
// inherits, so those are read when the payload has none. An agent that never registered as a
// seat has no sitting open and is never counted — the main session, an operator, anything else.
//
// Two calls get through a limited sitting. The envelope's never reaches this count: Pre hands it to
// gateEnvelope first, because returning the envelope is what the limit's refusal tells the seat to
// do. A register is counted and let through (hookgate.InvokesRegister), because it opens the next
// sitting's count.
func enforceLimit(in hookgate.Input, cwd string, stdout io.Writer, rec *hookfailures.Recorder, req ...hookcore.Request) (denied bool, err error) {
	caller := callerAgent(in)
	agentID := caller.ID
	if agentID == "" {
		return false, nil
	}
	runDir := seatRunDir(cwd, caller, rec)
	d, counted, err := sittingcap.Count(runDir, agentID)
	if err != nil || !counted || !d.Over || hookgate.InvokesRegister(in) {
		return false, err
	}
	if d.MarkErr != nil {
		// The seat is still refused below. What is lost is the record of the sitting reaching its
		// limit, which only the first refused call writes — and a marker that cannot be created
		// means no call ever counts as first.
		err = fmt.Errorf("cannot mark %s's sitting %d as limited, so its limit is not on the record: %w", agentID, d.Sitting, d.MarkErr)
	}
	if d.First {
		err = recordLimit(runDir, agentID, d.Sitting, d.Limit)
	}
	emitPreDeny(stdout, hookgate.LimitReason(d.Count, d.Limit), req...)
	return true, err
}

// seatRunDir is the run a seat's call belongs to: the marker its working directory finds, else the
// run the engine put in the process environment. One resolution for every hook that attributes a
// seat's call, so the turn limit and the failure log cannot disagree about which run it was in.
func seatRunDir(cwd string, caller runlive.Caller, rec *hookfailures.Recorder) string {
	inferred := runlive.InferRunDir(cwd, caller)
	noteInference(inferred, rec)
	for _, dir := range []string{inferred.Dir, os.Getenv(seatenv.Var)} {
		if dir != "" {
			return dir
		}
	}
	return ""
}

// callerAgent is the agent a call is made by: the payload's, else the one a headless seat carries in
// its environment — the same identity enforceLimit counts against — with the configuration it was
// dispatched under, which decides whether a decided run is still live for it (runlive.StateOf).
func callerAgent(in hookgate.Input) runlive.Caller {
	return callerOf(in.AgentID, in.AgentType, in.SessionID)
}

// callerOf is callerAgent for a payload this package decodes itself. The type comes from the same
// place the id does: a payload's id with the environment's type would pair two agents.
//
// THE SESSION IS ALWAYS THE PAYLOAD'S: it is the one fact a hook has about which conversation the
// call is in, and the run's membership is decided by it (runlive.InferRunDir).
func callerOf(agentID, agentType, session string) runlive.Caller {
	if agentID != "" {
		return runlive.Caller{ID: agentID, Type: agentType, Session: session}
	}
	return runlive.Caller{ID: seatenv.AgentID(), Type: seatenv.AgentType(), Session: session}
}

// noteInference records a marker that is a FAULT, and clears the entry when the marker is usable.
//
// hookgate.PreOutcome says nothing on an empty run directory, so an unusable marker meant the run
// directory was simply never injected — and the seat then hit "no run directory" from a different
// layer, which names the wrong cause. InferRunDir's own comment measures that class at ten of 55
// tool-call errors in one run. The ordinary shapes — no marker, no live run — stay silent; two live
// runs is a fault, because neither run's seats get a run injected.
func noteInference(i runlive.Inferred, rec *hookfailures.Recorder) {
	switch {
	case i.Why.Fault():
		rec.FailIn(runlive.StageUnusable, i.MarkerDir, runlive.FaultDetail(i)+
			" — the run directory is not being injected into seats' commands")
	case i.MarkerDir != "":
		rec.OKIn(runlive.StageUnusable, i.MarkerDir)
	}
	runlive.NoteMembership(i, rec, "the run directory is not injected into its commands, and its calls are not counted")
}

// envelopeTool is the tool a Workflow seat returns its envelope with: the harness tells each seat to
// call it exactly once to end its work, and to call it again when the call is refused.
const envelopeTool = "StructuredOutput"

// gateEnvelope is the hook on the call returning a seat's envelope. It denies the call when the record
// refuses it (record.EnvelopeRefusal, record.ChairEnvelope, asked through the writer): a seat's sitting
// cannot end while a blocking item on its own work list stands, and a chair's envelope carries no copy
// of the plan. The writer logs the refusal. An envelope it lets through carries the record's read: the
// hook writes `record_read` onto the call's input, overwriting anything the seat typed there, and the
// harness returns the rewritten input to the engine, which reads what it knows of the record from that
// field and from nothing a seat composed (record.RecordRead). Where the record supplies the envelope
// itself — the chair's, holding the plan the record stored last in its sitting — the read is written onto
// that envelope and the call carries it instead of what the chair stated: the workflow script's only
// surface is a seat's envelope, so this is how the stored plan reaches it. The envelope is admitted only
// once the call succeeds — by the PostToolUse hook (Post), after the harness has checked it against its
// schema.
//
// A CHECK THAT CANNOT RUN REFUSES THE ENVELOPE, naming the error (gblock's standing rule: an
// unreadable record refuses). An envelope nothing checked is not admitted on the strength of the
// check having failed. A call from no seat — no agent, no run — is no seat's envelope, and passes
// with nothing written on it.
func gateEnvelope(in hookgate.Input, cwd string, stdout io.Writer, rec *hookfailures.Recorder, req ...hookcore.Request) {
	caller := callerAgent(in)
	agentID := caller.ID
	if agentID == "" {
		return
	}
	runDir := seatRunDir(cwd, caller, rec)
	if runDir == "" {
		return
	}
	answer, err := envelopeAnswer(runDir, agentID, in.ToolInput)
	if err != nil {
		rec.Fail(StageEnvelope, "a seat's envelope could not be checked against the record, and was refused: "+err.Error())
		if answer.Refusal == "" {
			answer.Refusal = envelopeReadRefusal(err)
		}
	} else {
		rec.OK(StageEnvelope)
	}
	switch {
	case answer.Refusal != "":
		emitPreDeny(stdout, answer.Refusal, req...)
	case answer.Read != nil:
		base := in.ToolInput
		if answer.Envelope != nil {
			base = answer.Envelope
		}
		emitEnvelopeRead(stdout, base, answer.Read, rec, req...)
	}
}

// emitEnvelopeRead replaces the envelope call's input with base — the call's own input, or the envelope
// the record supplied in its place — carrying the record's read. The field is the hook's: a value the
// seat typed there is overwritten, as a Bash call's run directory is. An input that does not parse
// cannot carry it, and the envelope is refused rather than returned with no read, which would stop the
// run over a call the seat can simply make again.
func emitEnvelopeRead(stdout io.Writer, base, read json.RawMessage, rec *hookfailures.Recorder, req ...hookcore.Request) {
	updated := map[string]any{}
	if err := json.Unmarshal(base, &updated); err != nil {
		rec.Fail(StageEnvelope, "a seat's envelope did not parse, so the record's read could not be written on it: "+err.Error())
		emitPreDeny(stdout, envelopeReadRefusal(err), req...)
		return
	}
	updated[hookgate.EnvelopeReadField] = read
	var r hookcore.Request
	if len(req) > 0 {
		r = req[0]
	}
	if r.Event == "" {
		r.Event = "PreToolUse"
	}
	hookcore.Respond(r, hookcore.Response{
		Decision:     hookcore.DecisionAllow,
		Reason:       "feov-record: the record's read is written by the hook, not typed by the seat",
		UpdatedInput: updated,
	}, stdout, os.Stderr, nil)
}

// envelopeReadRefusal is the refusal of an envelope whose check against the record could not run,
// naming the error.
func envelopeReadRefusal(err error) string {
	return "your envelope is refused: it could not be checked against your work list — " + err.Error() +
		". An envelope nothing checked is not admitted; return it again"
}

// envelopeAnswer is a variable for the reason recordLimit is.
var envelopeAnswer = sittinghook.Envelope

// recordLimit is a variable so the hook's handoff can be tested without a built writer on disk.
var recordLimit = sittinghook.Limit

// recordRefusal, recordFailure, recordAdmission and handNextItem are variables for the same reason.
var (
	recordRefusal   = sittinghook.Refusal
	recordFailure   = sittinghook.Failure
	recordAdmission = sittinghook.Admission
	handNextItem    = sittinghook.NextItem
)

// cwdOf pulls the seat's working directory out of the raw payload.
//
// It is read from the RAW bytes rather than added to hookgate.Input because Input is the gate's
// contract and cwd is this program's business — the gate takes a resolved directory as a
// parameter precisely so it never has to know how one is found.
func cwdOf(raw []byte) string {
	var p struct {
		Cwd string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &p)
	return p.Cwd
}

// emitPreRewrite replaces the tool's arguments with the same call plus the injected run
// directory. It emits `permissionDecision: allow` with `updatedInput` — ONE document, the same
// slot a deny would have used, which is why the ordering in PreOutcome is structural.
//
// The whole tool_input is preserved and only `command` is replaced: a Bash call may carry other
// fields (description, timeout, run_in_background) and dropping them would silently change how
// the seat's command runs.
func emitPreRewrite(stdout io.Writer, toolInput json.RawMessage, command string, req ...hookcore.Request) {
	var r hookcore.Request
	if len(req) > 0 {
		r = req[0]
	}
	if r.Event == "" {
		r.Event = "PreToolUse"
	}
	if len(r.RawToolInput) == 0 && len(toolInput) > 0 {
		var m map[string]any
		if err := json.Unmarshal(toolInput, &m); err == nil {
			r.RawToolInput = m
		}
	}
	hookcore.Respond(r, hookcore.Response{
		Decision:       hookcore.DecisionAllow,
		Reason:         "feov-record: the run directory is injected by the engine, not typed by the seat",
		UpdatedCommand: command,
	}, stdout, os.Stderr, nil)
}

// emitPreAsk writes the PreToolUse ask document: hand the call to the human WITH THE REAL CAUSE.
//
// FAIL-OPEN VS FAIL-CLOSED IS A FALSE BINARY: `ask` is the third option. A hook that cannot
// understand its input need not guess a direction — `permissionDecisionReason` is carried back to
// the model, so the agent sees the actual error and can correct or escalate instead of being
// silently allowed or silently blocked.
//
// This is the SYNCHRONOUS half of the answer; a log entry of type LOG_TYPE_FAILURE
// is the durable half. The reason string reaches the agent NOW; the log entry is what a
// later reader finds. Neither replaces the other.
func emitPreAsk(stdout io.Writer, reason string, req ...hookcore.Request) {
	var r hookcore.Request
	if len(req) > 0 {
		r = req[0]
	}
	if r.Event == "" {
		r.Event = "PreToolUse"
	}
	hookcore.Respond(r, hookcore.Response{
		Decision: hookcore.DecisionAsk,
		Reason:   reason,
	}, stdout, os.Stderr, nil)
}

// emitPreDeny writes the PreToolUse deny document. `permissionDecisionReason` is carried back to
// the model, so the seat reads what the shell would have run and the form that works, in the same
// turn — a refusal that teaches rather than one the seat has to reverse-engineer.
func emitPreDeny(stdout io.Writer, reason string, req ...hookcore.Request) {
	var r hookcore.Request
	if len(req) > 0 {
		r = req[0]
	}
	if r.Event == "" {
		r.Event = "PreToolUse"
	}
	hookcore.Respond(r, hookcore.Response{
		Decision: hookcore.DecisionDeny,
		Reason:   reason,
	}, stdout, os.Stderr, nil)
}
