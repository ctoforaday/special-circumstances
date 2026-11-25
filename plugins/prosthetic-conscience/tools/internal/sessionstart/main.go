// Package sessionstart is the ONE hook on the SessionStart event.
//
// It replaces sc-toolchain-nudge and sc-checkpoint-restore as separate binaries. They remain
// separate UNITS; what changes is that ONE response is now composed from both.
//
// # Why this event forced a correction
//
// The two binaries were emitting on DIFFERENT CHANNELS, and separate processes hid it:
//
//	sc-toolchain-nudge     printed a bare line to stdout, on the strength of a comment
//	                       reading "SessionStart stdout reaches the session"
//	sc-checkpoint-restore  emitted a JSON document carrying additionalContext + watchPaths
//
// One process cannot do both: a bare line on stdout corrupts the document beside it. And the
// claim the nudge rested on is NOT in plans/hook-surface-spike.md §5's verified list — what
// IS verified there is additionalContext ("hook_additional_context attachment, transcript
// line 42"). So the merge puts both on the channel that was actually measured, which is a
// behaviour change for the nudge and the reason this is not a pure refactor.
//
// # The decision merge, which is this event's alone
//
// COMPOSE, not pick-first. additionalContext is the concatenation of every unit's text in
// unit order; watchPaths is the union, deduplicated. That is the opposite of PreToolUse,
// where only one permission document may be emitted and the first denial wins, and different
// again from PostToolUse, where the merge is over exit codes. Three events, three policies —
// which is why hookunit still does not own one.
//
// Silence stays silent: an empty response would spend tokens telling the session there is
// nothing to tell it. Either half is enough to speak, because a note can be worth WATCHING
// with nothing to say about it.
package sessionstart

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpointrestore"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikecounter"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/strikes"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchainnudge"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string   `json:"hookEventName"`
		AdditionalContext string   `json:"additionalContext,omitempty"`
		WatchPaths        []string `json:"watchPaths,omitempty"`
	} `json:"hookSpecificOutput"`
	// SystemMessage is the failure record's, for the human. AdditionalContext above is for the
	// agent, and a hook that is broken is not the agent's business to relay.
	SystemMessage string `json:"systemMessage,omitempty"`
}

// StageEncode is recorded when this event's response cannot be encoded: a restored checkpoint that
// never arrives, which used to be discarded with `_ =` and read as a session with no note.
const StageEncode hookfailures.Stage = "encode"

// ensure is a variable so a test can answer the version check without a fetch-bin.sh on disk.
var ensure = hookfailures.Ensure

// merge composes one response from several units.
//
// IT USED TO READ ONLY Stdout AND Watch, which meant a unit's Stderr — the channel hookunit puts a
// PANIC in — was dropped here, before it even reached the debug log. A crashed unit was silent on
// every channel on this event. The panic is now recorded by hookunit itself, and the text is
// forwarded so the debug log keeps its copy.
func merge(results []hookunit.Result) (text string, watch []string, logged string) {
	var parts, logs []string
	seen := map[string]bool{}
	for _, r := range results {
		if t := strings.TrimRight(r.Stdout, "\n"); t != "" {
			parts = append(parts, t)
		}
		if l := strings.TrimRight(r.Stderr, "\n"); l != "" {
			logs = append(logs, l)
		}
		for _, p := range r.Watch {
			if !seen[p] {
				seen[p] = true
				watch = append(watch, p)
			}
		}
	}
	return strings.Join(parts, "\n\n"), watch, strings.Join(logs, "\n")
}

// stageResumeMarker records read/write failures on the Antigravity resume-restore marker so a
// broken marker fails loudly instead of silently skipping checkpoint restore on resume.
const stageResumeMarker hookfailures.Stage = "resume-restore-marker"

var hostProcessID = defaultHostProcessID

func resumeMarkerPath(pluginDataDir, sessionID string) string {
	safe := strings.NewReplacer("/", "_", "\\", "_", "..", "_").Replace(sessionID)
	return filepath.Join(pluginDataDir, "session_restored_"+safe+".txt")
}

func writeResumeMarker(req hookcore.Request, hostID string, rec *hookfailures.Recorder) {
	if req.PluginDataDir == "" || req.SessionID == "" {
		return
	}
	p := resumeMarkerPath(req.PluginDataDir, req.SessionID)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		rec.Fail(stageResumeMarker, "cannot create resume marker dir: "+err.Error())
		return
	}
	if err := os.WriteFile(p, []byte(hostID+"\n"), 0o600); err != nil {
		rec.Fail(stageResumeMarker, "cannot write resume marker "+p+": "+err.Error())
		return
	}
	rec.OK(stageResumeMarker)
}

func needsResumeRestore(req hookcore.Request, hostID string, rec *hookfailures.Recorder) bool {
	if req.SessionID == "" {
		return false
	}
	// Only check on invocationNum == 0 (first model step of a turn), never mid-turn.
	var payload struct {
		InvocationNum *int `json:"invocationNum"`
	}
	_ = json.Unmarshal(req.Raw, &payload)
	if payload.InvocationNum != nil && *payload.InvocationNum != 0 {
		return false
	}
	if req.PluginDataDir == "" {
		rec.Fail(stageResumeMarker, "PLUGIN_DATA unset; cannot check resume-restore marker")
		return true
	}
	p := resumeMarkerPath(req.PluginDataDir, req.SessionID)
	b, err := os.ReadFile(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			rec.Fail(stageResumeMarker, "cannot read resume marker "+p+": "+err.Error())
		}
		return true
	}
	return strings.TrimSpace(string(b)) != hostID
}

func selectCheckpointUnit(units []hookunit.Unit) []hookunit.Unit {
	if len(units) == 0 {
		return []hookunit.Unit{checkpointrestore.Unit()}
	}
	for _, u := range units {
		if u.Name == "sc-checkpoint-restore" {
			return []hookunit.Unit{u}
		}
	}
	return nil
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	eventName := "SessionStart"
	declare := func(fs *flag.FlagSet) {
		fs.StringVar(&eventName, "event", "SessionStart", "hook event name")
	}
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-sessionstart"), declare) {
		return 0
	}

	req := hookcore.Read(eventName, stdin)
	if projectDir == "" {
		projectDir = req.ProjectDir
	} else {
		projectDir = hookenv.ProjectDir(projectDir, "")
	}

	rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", eventName, now, stderr)
	// THE VERSION CHECK IS THIS HOOK'S TOO, so SessionStart has one hook in this plugin. It runs
	// before the project root is resolved: the binaries are the machine's, and a session with no
	// root still heals — only the message waits, because such a session writes nothing to stdout.
	healed, err := ensure(eventName)
	if err != nil {
		rec.Fail(hookfailures.StageEnsure, "sc-sessionstart: "+err.Error())
	} else {
		rec.OK(hookfailures.StageEnsure)
	}
	ctx := hookunit.NewCtx(eventName, req.Raw, projectDir, now, rec)
	if !hookenv.Explain(ctx.ProjectDir, rec, "sc-sessionstart") {
		// Recorded, and NOT said here: a hook with no project root writes nothing to stdout
		// (projectroot_test.go pins that for every binary). The next event that resolves one
		// announces it.
		rec.Persist()
		return 0
	}

	if eventName == "PreInvocation" && req.Engine == hookcore.EngineAntigravity {
		var feedbackParts []string
		scope := hookcore.DrainScopeKey(req)
		failedSteps, drainErr := hookcore.DrainFailedToolSteps(req)
		if drainErr != nil {
			rec.FailTTL(hookcore.StageFailedStepDrain(), scope, "failed-step drain: "+drainErr.Error())
		} else {
			rec.OKTTL(hookcore.StageFailedStepDrain(), scope)
		}
		sawHookErr := false
		for _, st := range failedSteps {
			if st.HookError {
				// Our hook's output failed to parse: record on hookfailures for the human,
				// never strike the agent for our own hook defect.
				sawHookErr = true
				rec.FailTTL(hookcore.StageHookUnmarshal(), scope, "hook result failed to parse: "+strings.TrimSpace(st.Output))
				continue
			}
			// Never strike tools without a command (e.g. Read or manage_task): their key would
			// collapse to the tool name alone and trip anti-spinning on unrelated calls.
			if ctx.ProjectDir == "" || strings.TrimSpace(st.Command) == "" {
				continue
			}
			path := strikes.Path(ctx.ProjectDir)
			state, err := strikes.Load(os.ReadFile, path)
			if err != nil {
				rec.FailIn(strikecounter.StageStrikesRead, ctx.ProjectDir, err.Error()+" — strikes are counted from zero until it can be read")
			} else {
				rec.OKIn(strikecounter.StageStrikesRead, ctx.ProjectDir)
			}
			rawInput, _ := json.Marshal(map[string]any{"command": st.Command})
			key := strikes.Key(st.ToolName, rawInput)
			count, tripped := strikes.Record(&state, key, now)
			if err := strikes.Save(path, state); err != nil {
				rec.FailIn(strikecounter.StageStrikesWrite, ctx.ProjectDir, "cannot write "+path+": "+err.Error()+
					" — strikes are not being kept, so the 3-strike rule cannot trip")
			} else {
				rec.OKIn(strikecounter.StageStrikesWrite, ctx.ProjectDir)
			}
			if tripped {
				errSummary := fmt.Sprintf("Command exited with code %d: %s", st.ExitCode, st.Output)
				feedbackParts = append(feedbackParts, strikecounter.Message(st.ToolName, errSummary, key, count))
			}
		}
		if !sawHookErr {
			rec.OKTTL(hookcore.StageHookUnmarshal(), scope)
		}

		// Antigravity does not fire SessionStart on conversation resume (--conversation) or after a
		// language-server restart (pinned by TestLiveAsymmetrySessionStartOnResume). Run
		// checkpoint-restore once on the first PreInvocation (invocationNum == 0) for a
		// (conversation, hosting-process) pair that has not yet been marked.
		var restoreText string
		var watch []string
		hostID := hostProcessID()
		if needsResumeRestore(req, hostID, rec) {
			if cpUnits := selectCheckpointUnit(units); len(cpUnits) > 0 {
				var logged string
				restoreText, watch, logged = merge(hookunit.Run(ctx, cpUnits))
				if logged != "" {
					fmt.Fprintln(stderr, logged)
				}
			}
			writeResumeMarker(req, hostID, rec)
		}

		said := strings.TrimSpace(hookfailures.JoinMessages(healed, rec.Settle()))
		resp := hookcore.Response{
			AdditionalContext: strings.TrimSpace(restoreText),
			WatchPaths:        watch,
			Feedback:          strings.Join(feedbackParts, "\n\n"),
			Say:               said,
		}
		return hookcore.Respond(req, resp, stdout, stderr, nil)
	}

	if req.Engine == hookcore.EngineAntigravity && eventName == "SessionStart" {
		writeResumeMarker(req, hostProcessID(), rec)
	}

	if len(units) == 0 {
		units = Units(req.PluginRoot)
	}
	text, watch, logged := merge(hookunit.Run(ctx, units))
	if logged != "" {
		fmt.Fprintln(stderr, logged)
	}
	said := hookfailures.JoinMessages(healed, rec.Settle())
	if strings.TrimSpace(text) == "" && len(watch) == 0 && strings.TrimSpace(said) == "" {
		return 0
	}

	resp := hookcore.Response{
		AdditionalContext: strings.TrimSpace(text),
		WatchPaths:        watch,
		Say:               strings.TrimSpace(said),
	}
	return hookcore.Respond(req, resp, stdout, stderr, nil)
}

// Units are this event's checks, in the order their text should appear.
func Units(pluginRoot string) []hookunit.Unit {
	return []hookunit.Unit{
		toolchainnudge.Unit(pluginRoot, toolchain.Probe),
		checkpointrestore.Unit(),
	}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), nil)
}
