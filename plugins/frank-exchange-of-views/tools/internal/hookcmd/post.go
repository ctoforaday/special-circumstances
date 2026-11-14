package hookcmd

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"

	"github.com/ctoforaday/special-circumstances/shared/hookcore"
	"github.com/ctoforaday/special-circumstances/shared/hookfailures"
)

// Post is the PostToolUse entry point. It acts on two calls of a seat's, and on nothing else:
//
//   - the call returning a seat's envelope, SUCCEEDED, is recorded as the envelope's admission
//     (record.FailedSittings reads it);
//   - a lens's Bash call to the record hands the lens what it is handed next (nextItem).
//
// AFTER THE CALL, NOT BEFORE IT. The PreToolUse gate (gateEnvelope) asks the record whether it
// refuses the envelope; the harness checks the envelope against its schema inside the tool's own
// call, after every PreToolUse hook. PostToolUse fires on a call that succeeded, so an admission
// written here is an envelope both the record and the schema accepted. A call the schema rejects
// fails, and reaches PostToolUseFailure instead (PostFailure logs it).
//
// THE MATCHER IN hooks.json NAMES THE ENVELOPE TOOL AND Bash, so this binary is spawned for no other
// call; the tool name is still checked here, because the matcher is configuration and this is the
// claim. A call from no seat — no agent, no run — is no seat's, and nothing is written or said.
func Post(stdin io.Reader, stdout io.Writer, rec *hookfailures.Recorder) error {
	req := hookcore.Read("PostToolUse", stdin)
	if !req.Parsed {
		rec.Fail(StageAdmission, "cannot read or parse the PostToolUse payload — a seat's accepted envelope was not recorded as admitted")
		return nil
	}
	switch req.ToolName {
	case envelopeTool:
		admit(req, rec)
	case "Bash":
		nextItem(req, stdout, rec)
	}
	return nil
}

// admit records the seat's accepted envelope as admitted.
func admit(req hookcore.Request, rec *hookfailures.Recorder) {
	caller := callerOf(req.AgentID, req.AgentType, req.SessionID)
	if caller.ID == "" {
		return
	}
	runDir := seatRunDir(cwdOfRequest(req), caller, rec)
	if runDir == "" {
		return
	}
	if err := recordAdmission(runDir, caller.ID); err != nil {
		rec.Fail(StageAdmission, "a seat's envelope was accepted and its admission could not be recorded, so its sitting reads as failed: "+err.Error())
		return
	}
	rec.OK(StageAdmission)
}

// nextItem hands a seat what it is handed next after a call of its that went to the record: the
// sitting's report text one item per call while any is owed, then its next item when the record shows
// the current one discharged, the current one again when the record moved it, a line when another
// item left its list, and nothing when nothing changed (sittingwrite.WriteNextItem, through the
// writer, which links the record this binary may not).
//
// EVERY SEAT'S CALL TO THE RECORD SPAWNS THE WRITER, and no other call: an act is recorded only by a
// call to the record, so a call whose command does not name the tool (hookgate.CallsRecord) cannot
// have discharged anything. A call it misses — the tool reached some other way — costs a delivery
// until the seat's next call to the record, never the item: what the seat was handed stays on file.
//
// THE DELIVERY IS additionalContext ON THE CALL'S RESULT, which reaches a Workflow seat's context as it
// reaches an Agent subagent's. A delivery of report text that cannot advance ends the seat's turn
// ("continue": false); the writer has already logged it as the tool's defect. Under Antigravity a
// PostToolUse response reaches the agent through hookcore's queue, and the writer hands no report
// text there.
func nextItem(req hookcore.Request, stdout io.Writer, rec *hookfailures.Recorder) {
	caller := callerOf(req.AgentID, req.AgentType, req.SessionID)
	if caller.ID == "" {
		return
	}
	command, _ := req.RawToolInput["command"].(string)
	if !hookgate.CallsRecord(command) {
		return
	}
	runDir := seatRunDir(cwdOfRequest(req), caller, rec)
	if runDir == "" {
		return
	}
	engine := req.Engine
	if engine == "" {
		engine = hookcore.EngineClaude
	}
	text, err := handNextItem(runDir, caller.ID, caller.Type, string(engine))
	if errors.Is(err, hookgate.ErrDeliveryHeld) {
		// ANOTHER OF THE LENS'S CALLS HOLDS THE DELIVERY and is handing the item: this call hands
		// nothing, and nothing failed — nor did this call see the stage work.
		return
	}
	if errors.Is(err, hookgate.ErrDeliveryStuck) && caller.ID != req.SessionID {
		// A DELIVERY OF REPORT TEXT THAT CANNOT ADVANCE ENDS THE SEAT'S TURN, and only a seat's: the
		// writer has logged it as the tool's defect, and a seat handed nothing would read its envelope
		// as waiting on text that never comes. The lead session itself is never stopped here.
		rec.Fail(StageNextItem, "a seat's delivery of report text cannot advance, and its turn was ended: "+err.Error())
		hookcore.Respond(req, hookcore.Response{TerminateTurn: true, Reason: deliveryStuckStop + "\n\n" + err.Error()}, stdout, os.Stderr, nil)
		return
	}
	if err != nil {
		rec.Fail(StageNextItem, "a seat's call went to the record and what it is handed next could not be read: "+err.Error())
		// NEVER SILENCE IN ITS PLACE: a lens told nothing reads "nothing changed".
		text = nextItemUnreadable + "\n\n" + err.Error() + "\n"
	} else {
		rec.OK(StageNextItem)
	}
	if strings.TrimSpace(text) == "" {
		return
	}
	hookcore.Respond(req, hookcore.Response{AdditionalContext: text}, stdout, os.Stderr, nil)
}

// deliveryStuckStop is what a seat is told when its turn ends on a delivery that cannot advance.
const deliveryStuckStop = "THE REPORT TEXT OWED TO YOU THIS SITTING COULD NOT BE DELIVERED, so your turn ends here: " +
	"the tool logged it as its own defect, and the sitting is the operator's to look at."

// nextItemUnreadable is what a lens is told when the writer could not say what it is handed next.
const nextItemUnreadable = "WHAT YOU ARE HANDED NEXT COULD NOT BE READ, and that is NOT the same as nothing having " +
	"changed: nothing was determined. Read your work list from your own surface before you act."

// cwdOfRequest is the payload's working directory.
func cwdOfRequest(req hookcore.Request) string {
	if req.CWD != "" {
		return req.CWD
	}
	return cwdOf(req.Raw)
}
