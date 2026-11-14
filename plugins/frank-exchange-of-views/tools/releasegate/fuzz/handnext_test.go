package fuzz

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// EVERY SEAT'S WORK ARRIVES THROUGH THE DELIVERY WRITER, one item after each of its calls to the record,
// as the PostToolUse hook hands it (hookcmd.nextItem → sittingwrite.WriteNextItem). The drive spawns the
// shipped writer after every call of a seat's that the record admitted (execAs): report text is handed
// one item per call, and the writer writes the seat's report_read when it hands the last. Nothing in
// the drive writes a report_read itself, so a sitting holds one exactly when the writer completed it.
//
// THE SEAT IS RESOLVED AS THE HOOK RESOLVES IT: by the record's binding of the agent where the sitting
// opened under one (registerTwice, lookThenRegister), and otherwise by its agent configuration
// (record.SeatOfAgentType), which names every red seat. A sitting the drive opened by a plain register
// carries no agent, so its calls are attributed to one stable id per seat (deliveryAgentOf); the
// writer keys what it handed by agent and sitting, so each sitting still starts from the head of its
// queue. Blue's configurations seat several ids, so a blue sitting opened without an agent is handed
// nothing, as a live one would not be before its register binds it.

// handedEngine is the harness the drive's seats run under: the one whose hooks hand report text.
const handedEngine = string(hookcore.EngineClaude)

// maxDeliveryCalls bounds the calls readDelivery makes for one seat: a report the fuzz builds is cut
// into far fewer items, so a delivery still owed past it is a writer that stopped handing it.
const maxDeliveryCalls = 200

// deliveryAgentOf is the agent id a seat's calls are attributed to when the call itself carries none:
// its sitting's agent where the opening carried one, and otherwise one stable id for the seat.
func (r *runner) deliveryAgentOf(seatID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a := r.sittingAgent[seatID]; a != "" {
		return a
	}
	return "fuzz_seat_" + seatID
}

// seatOfCall is the seat a call names with --seat-id, or "" for a call of no seat's.
func seatOfCall(args []string) string {
	i := slices.Index(args, "--seat-id")
	if i < 0 || i+1 >= len(args) {
		return ""
	}
	if seat := args[i+1]; seat != "operator" {
		return seat
	}
	return ""
}

// handNext spawns the writer's next phase for seatID after one of its calls to the record, under
// agent (or the seat's own, deliveryAgentOf), with the argv the hook gives it (sittinghook). A held
// delivery hands nothing and is no failure: another of the seat's calls is handing the item. A
// delivery that cannot advance is the tool's defect, and the run fails on it.
func (r *runner) handNext(seatID, agent string) {
	if r.writer == "" {
		return
	}
	if agent == "" {
		agent = r.deliveryAgentOf(seatID)
	}
	// THE QUEUE OF A LENS IS READ BEFORE THE WRITER RUNS, for the oracle that holds its current item to
	// the one handed before while that one is still queued (headMovesOnlyOnDischarge).
	lens := strings.HasPrefix(seatID, "red-lens-")
	var queued map[string]bool
	sitting := ""
	if lens {
		if w, err := record.WorkOfSeat(r.run(), "lens", seatID); err == nil {
			queued = map[string]bool{}
			for _, e := range record.QueueOf(w) {
				queued[e.Key] = true
			}
			sitting = fmt.Sprintf("%s#%s#%d", seatID, agent, w.SittingID())
		}
	}
	cmd := exec.Command(r.writer, "-run", r.runDir, "-phase", string(sittingwrite.Next),
		"-agent-id", agent, "-agent-type", agentTypeOf(seatID), "-engine", handedEngine)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	r.progress.Store(time.Now().UnixNano())
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		switch exit.ExitCode() {
		case hookgate.ExitDeliveryHeld:
			return
		case hookgate.ExitDeliveryStuck:
			r.noteActErr(fmt.Sprintf("the writer could not advance %s's delivery of report text: %s", seatID, strings.TrimSpace(stderr.String())))
			return
		}
	}
	if err != nil {
		r.noteActErr(fmt.Sprintf("the writer's next phase failed for %s: %v: %s", seatID, err, strings.TrimSpace(stderr.String())))
		return
	}
	text := stdout.String()
	if strings.Contains(text, "COULD NOT BE READ") {
		r.noteActErr("the writer could not read what " + seatID + " is handed next: " + firstLine(text))
		return
	}
	if lens && sitting != "" {
		r.headMovesOnlyOnDischarge(seatID, sitting, queued, text)
	}
}

// owesReportText is whether seatID's sitting is owed report text still to be handed: the head of its
// queue is a report-text item, the writer's own test (WriteNextItem).
func (r *runner) owesReportText(seatID string) (bool, error) {
	w, err := record.WorkOfSeat(r.run(), seatclass.RoleOfSeat(seatID), seatID)
	if err != nil {
		return false, err
	}
	q := record.QueueOf(w)
	return len(q) > 0 && q[0].Kind == record.ItemReportText, nil
}

// readDelivery is a seat reading what it is owed: while report text is owed to its sitting, it calls
// the record, and the writer hands it the next item after each call (execAs → handNext). It returns
// once the writer has handed the last item and written the sitting's report_read.
func (r *runner) readDelivery(seatID, agent string) {
	for i := 0; ; i++ {
		owed, err := r.owesReportText(seatID)
		if err != nil {
			r.noteActErr("what " + seatID + " is owed does not read: " + err.Error())
			return
		}
		if !owed {
			return
		}
		if i == maxDeliveryCalls {
			r.noteActErr(fmt.Sprintf("%s is still owed report text after %d calls to the record: the writer stopped handing it", seatID, maxDeliveryCalls))
			return
		}
		if _, ok := r.admitAs(agent, "show", "--seat-id", seatID); !ok {
			return
		}
	}
}
