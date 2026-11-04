package hookcmd

import (
	"io"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// Post is the PostToolUse entry point: a seat's call returning its envelope that SUCCEEDED is
// recorded as the envelope's admission (record.FailedSittings reads it).
//
// AFTER THE CALL, NOT BEFORE IT. The PreToolUse gate (gateEnvelope) asks the record whether it
// refuses the envelope; the harness checks the envelope against its schema inside the tool's own
// call, after every PreToolUse hook. PostToolUse fires on a call that succeeded, so an admission
// written here is an envelope both the record and the schema accepted. A call the schema rejects
// fails, and reaches PostToolUseFailure instead (PostFailure logs it).
//
// THE MATCHER IN hooks.json NAMES THE ENVELOPE TOOL, so this binary is spawned for no other call;
// the tool name is still checked here, because the matcher is configuration and this is the claim.
// A call from no seat — no agent, no run — is no seat's envelope, and nothing is written.
func Post(stdin io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
	req := hookcore.Read("PostToolUse", stdin)
	if !req.Parsed {
		rec.Fail(StageAdmission, "cannot read or parse the PostToolUse payload — a seat's accepted envelope was not recorded as admitted")
		return nil
	}
	if req.ToolName != envelopeTool {
		return nil
	}
	caller := callerOf(req.AgentID, req.AgentType, req.SessionID)
	if caller.ID == "" {
		return nil
	}
	cwd := req.CWD
	if cwd == "" {
		cwd = cwdOf(req.Raw)
	}
	runDir := seatRunDir(cwd, caller, rec)
	if runDir == "" {
		return nil
	}
	if err := recordAdmission(runDir, caller.ID); err != nil {
		rec.Fail(StageAdmission, "a seat's envelope was accepted and its admission could not be recorded, so its sitting reads as failed: "+err.Error())
		return nil
	}
	rec.OK(StageAdmission)
	return nil
}
