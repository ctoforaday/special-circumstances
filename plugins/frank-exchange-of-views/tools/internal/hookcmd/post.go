package hookcmd

import (
	"encoding/json"
	"io"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookfailures"
)

// Post is the PostToolUse entry point: a seat's call returning its envelope that SUCCEEDED is
// recorded as the envelope's admission (record.FailedSittings reads it).
//
// AFTER THE CALL, NOT BEFORE IT. The PreToolUse gate (refuseEnvelope) asks the record whether it
// refuses the envelope; the harness checks the envelope against its schema inside the tool's own
// call, after every PreToolUse hook. PostToolUse fires on a call that succeeded, so an admission
// written here is an envelope both the record and the schema accepted. A call the schema rejects
// fails, and reaches PostToolUseFailure instead (PostFailure logs it).
//
// THE MATCHER IN hooks.json NAMES THE ENVELOPE TOOL, so this binary is spawned for no other call;
// the tool name is still checked here, because the matcher is configuration and this is the claim.
// A call from no seat — no agent, no run — is no seat's envelope, and nothing is written.
func Post(stdin io.Reader, _ io.Writer, rec *hookfailures.Recorder) error {
	var in struct {
		ToolName  string `json:"tool_name"`
		AgentID   string `json:"agent_id"`
		AgentType string `json:"agent_type"`
		Cwd       string `json:"cwd"`
		SessionID string `json:"session_id"`
	}
	b, err := io.ReadAll(stdin)
	if err != nil || json.Unmarshal(b, &in) != nil {
		rec.Fail(StageAdmission, "cannot read or parse the PostToolUse payload — a seat's accepted envelope was not recorded as admitted")
		return nil
	}
	if in.ToolName != envelopeTool {
		return nil
	}
	caller := callerOf(in.AgentID, in.AgentType, in.SessionID)
	if caller.ID == "" {
		return nil
	}
	runDir := seatRunDir(in.Cwd, caller, rec)
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
