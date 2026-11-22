package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

func main() {
	evFlag := flag.String("event", "PreToolUse", "hook event name")
	flag.Parse()

	req := hookcore.Read(*evFlag, os.Stdin)

	// Diagnostic logging to stderr so we can verify execution in test traces
	fmt.Fprintf(os.Stderr, "[hookprobe] engine=%s event=%s tool=%s cmd=%s\n",
		req.Engine, req.Event, req.ToolName, req.Command)

	switch req.Event {
	case "PreTool", "PreToolUse":
		if strings.Contains(req.Command, "TRIGGER_BLOCK") {
			code := hookcore.Respond(req, hookcore.Response{
				Decision: hookcore.DecisionDeny,
				Reason:   "BLOCKED_BY_HOOKCORE_PROBE",
			}, os.Stdout, os.Stderr, nil)
			os.Exit(code)
		}
		if strings.Contains(req.Command, "TRIGGER_REWRITE") {
			code := hookcore.Respond(req, hookcore.Response{
				Decision:       hookcore.DecisionAllow,
				UpdatedCommand: "echo 'HOOKCORE_REWRITE_SUCCESS'",
			}, os.Stdout, os.Stderr, nil)
			os.Exit(code)
		}
		code := hookcore.Respond(req, hookcore.Response{
			Decision: hookcore.DecisionAllow,
		}, os.Stdout, os.Stderr, nil)
		os.Exit(code)

	case "PostTool", "PostToolUse":
		if strings.Contains(req.Command, "TRIGGER_FEEDBACK") {
			code := hookcore.Respond(req, hookcore.Response{
				Feedback: "HOOKCORE_FEEDBACK_DETECTED",
			}, os.Stdout, os.Stderr, nil)
			os.Exit(code)
		}
		code := hookcore.Respond(req, hookcore.Response{}, os.Stdout, os.Stderr, nil)
		os.Exit(code)

	case "SessionStart", "PreInvocation":
		ctxMsg := "HOOKCORE_CONTEXT_NOTICE"
		if req.Event == "PreInvocation" {
			var parts []string
			failed, _ := hookcore.DrainFailedToolSteps(req)
			for _, st := range failed {
				parts = append(parts, fmt.Sprintf("HOOKCORE_FAILED_STEP_CODE_%d:%s", st.ExitCode, st.Command))
			}
			if len(parts) > 0 {
				ctxMsg = strings.Join(parts, "; ")
			}
		}
		code := hookcore.Respond(req, hookcore.Response{
			Say:               "HOOKCORE_SAY_NOTICE",
			AdditionalContext: ctxMsg,
		}, os.Stdout, os.Stderr, nil)
		os.Exit(code)

	case "Stop", "PostInvocation":
		if strings.Contains(req.Command, "TRIGGER_STOP_NUDGE") {
			code := hookcore.Respond(req, hookcore.Response{
				AdditionalContext: "HOOKCORE_STOP_NUDGE",
			}, os.Stdout, os.Stderr, nil)
			os.Exit(code)
		}
		code := hookcore.Respond(req, hookcore.Response{}, os.Stdout, os.Stderr, nil)
		os.Exit(code)

	case "PreCompact":
		code := hookcore.Respond(req, hookcore.Response{
			PreCompactSteer: "HOOKCORE_STEER_ACTIVE",
		}, os.Stdout, os.Stderr, nil)
		os.Exit(code)

	default:
		code := hookcore.Respond(req, hookcore.Response{}, os.Stdout, os.Stderr, nil)
		os.Exit(code)
	}
}
