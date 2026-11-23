// Package pretooluse is the ONE hook on the PreToolUse event.
//
// It replaces sc-secrets-gate and sc-push-freeze-guard as separate binaries. They remain
// separate UNITS, and their POSTURES ARE OPPOSITE — which is the whole difficulty of this
// event and the reason the merge policy is written here rather than generalised:
//
//	sc-secrets-gate      fails CLOSED. It can DENY, and it denies on an undecodable
//	                     payload rather than allowing (#211).
//	sc-push-freeze-guard NEVER blocks. The freeze is a commitment the human may
//	                     consciously override; the guard makes it impossible to forget,
//	                     not impossible to break.
//
// A merge that flattened those into one posture would either turn a security gate into a
// warning or turn an advisory into a blocker. Each unit keeps its own.
//
// # The decision merge
//
// DENY WINS, and only one decision can be emitted: the PreToolUse protocol reads a single
// permission document from stdout. So stdout is PICK-FIRST in unit order, not composed —
// unlike PostToolUse, where every unit's stderr must survive, and unlike SessionStart, where
// one response is assembled from several. Three events, three policies; that is why
// hookunit does not own one.
//
// Warnings are independent of the decision: a unit that only warns still warns when another
// unit denies, because the operator needs both facts.
//
// The exit code is ALWAYS 0. The block travels in the JSON, never in the status.
package pretooluse

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/pushfreezeguard"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secretsgate"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// StageDecision is a unit's decision document that this hook could not read.
const StageDecision hookfailures.Stage = "unit-decision"

// unreadableReason is what the agent is told when a unit's decision could not be read.
const unreadableReason = "sc-pretooluse: a PreToolUse check wrote a decision this hook cannot read, so the call is refused. " +
	"This is a defect in the hook, not in the call: tell the human, and do not retry or rephrase the call to get past it."

// readDecision turns a unit's decision document into the response hookcore renders for the
// calling harness.
//
// EVERY DECISION GOES THROUGH hookcore. A document written to stdout as-is is read by neither
// harness as a decision: Claude Code ignores stdout that is not its JSON and runs the call, and
// Antigravity fails the step with "failed to unmarshal result from hook" (#1927). The one unit
// that writes a decision is the secrets gate, which fails closed, so a document this hook cannot
// read — not JSON, or JSON naming no permissionDecision or one that is not allow, deny or ask — is a DENY, recorded as this hook's
// failure so the defect surfaces rather than recurring silently.
func readDecision(decision string, rec *hookfailures.Recorder) hookcore.Response {
	var doc struct {
		HookSpecificOutput struct {
			PermissionDecision       string         `json:"permissionDecision"`
			PermissionDecisionReason string         `json:"permissionDecisionReason"`
			UpdatedInput             map[string]any `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	err := json.Unmarshal([]byte(decision), &doc)
	if err == nil {
		switch hookcore.Decision(doc.HookSpecificOutput.PermissionDecision) {
		case hookcore.DecisionAllow, hookcore.DecisionDeny, hookcore.DecisionAsk:
		case hookcore.DecisionNone:
			err = fmt.Errorf("it names no hookSpecificOutput.permissionDecision")
		default:
			err = fmt.Errorf("permissionDecision %q is not allow, deny or ask", doc.HookSpecificOutput.PermissionDecision)
		}
	}
	if err != nil {
		rec.Fail(StageDecision, fmt.Sprintf("a unit's PreToolUse decision could not be read (%v), so the call was refused: %.200q", err, decision))
		return hookcore.Response{Decision: hookcore.DecisionDeny, Reason: unreadableReason}
	}
	rec.OK(StageDecision)
	resp := hookcore.Response{
		Decision: hookcore.Decision(doc.HookSpecificOutput.PermissionDecision),
		Reason:   doc.HookSpecificOutput.PermissionDecisionReason,
	}
	if cmd, ok := doc.HookSpecificOutput.UpdatedInput["command"].(string); ok {
		resp.UpdatedCommand = cmd
	}
	return resp
}

// merge is the event's decision policy, kept pure so it can be tested without units.
func merge(results []hookunit.Result) (decision, warnings, say string) {
	var warn, said []string
	for _, r := range results {
		if decision == "" && strings.TrimSpace(r.Stdout) != "" {
			decision = r.Stdout // first denial in unit order wins; only one may be emitted
		}
		if w := strings.TrimRight(r.Stderr, "\n"); w != "" {
			warn = append(warn, w)
		}
		if t := strings.TrimRight(r.Say, "\n"); t != "" {
			said = append(said, t)
		}
	}
	return decision, strings.Join(warn, "\n"), strings.Join(said, "\n")
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-pretooluse")) {
		return 0
	}

	req := hookcore.Read("PreToolUse", stdin)
	if projectDir == "" {
		projectDir = req.ProjectDir
	} else {
		projectDir = hookenv.ProjectDir(projectDir, "")
	}

	rec := hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", now, stderr)
	ctx := hookunit.NewCtx("PreToolUse", req.Raw, projectDir, now, rec)
	hookenv.Note(ctx.ProjectDir, rec) // recorded, never a reason to skip the secrets gate
	decision, warnings, say := merge(hookunit.Run(ctx, units))

	var resp hookcore.Response
	if decision != "" {
		resp = readDecision(decision, rec)
	} else {
		rec.OK(StageDecision)
	}
	resp.Warnings = warnings
	resp.Say = strings.TrimSpace(say + "\n\n" + rec.Settle())

	return hookcore.Respond(req, resp, stdout, stderr, nil)
}

// Units are this event's checks, in the order a decision is taken from.
func Units() []hookunit.Unit {
	return []hookunit.Unit{secretsgate.Unit(), pushfreezeguard.Unit()}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), Units())
}
