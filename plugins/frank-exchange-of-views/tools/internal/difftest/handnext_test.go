package difftest

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// EVERY SEAT'S REPORT TEXT ARRIVES THROUGH THE DELIVERY WRITER, one item after each of its calls to the
// record, as the PostToolUse hook hands it (hookcmd.nextItem → sittingwrite.WriteNextItem). The replay
// spawns the shipped writer (`feov-sitting-write -phase next`) after every call of a seat's that the
// record admitted; a refused call reaches PostToolUseFailure, which hands nothing. The writer writes
// the seat's report_read when it hands the last item, and nothing in this package writes one itself,
// so a sitting holds a complete delivery exactly when the writer completed it.
//
// The fuzz's sittings open by a plain register, which binds no agent, so each seat's calls are
// attributed to one stable id (fuzzAgentOf) and the seat is resolved from its agent configuration
// (record.SeatOfAgentType), as the writer resolves a seat its binding does not name. Blue's
// configurations seat several ids and name no seat, so a blue sitting is handed nothing.

// handedEngine is the harness the fuzz's seats run under: the one whose hooks hand report text.
const handedEngine = string(hookcore.EngineClaude)

// maxDeliveryCalls bounds the calls readDelivery makes for one seat: a report the fuzz seeds is cut
// into far fewer items, so a delivery still owed past it is a writer that stopped handing it.
const maxDeliveryCalls = 200

// fuzzAgentOf is the agent id a seat's calls are attributed to: one stable id per seat, so two replays
// key the writer's buffer alike.
func fuzzAgentOf(seatID string) string { return "fuzz_seat_" + seatID }

// agentTypeOf is the agent configuration that dispatches seatID.
func agentTypeOf(seatID string) string {
	if s, ok := seatclass.Seats[seatID]; ok {
		return s.AgentType
	}
	return seatclass.Lane.AgentType
}

// seatOfArgs is the seat a call names with --seat-id, or "" for a call of no seat's.
func seatOfArgs(args []string) string {
	for i, a := range args {
		if a == "--seat-id" && i+1 < len(args) && args[i+1] != "operator" {
			return args[i+1]
		}
	}
	return ""
}

// handNext spawns the writer's next phase for seatID after one of its admitted calls to the record. A
// held delivery hands nothing and is no failure: another of the seat's calls is handing the item. A
// delivery that cannot advance, a writer that fails, and a list it cannot read are the tool's defects,
// and fail the test. What it hands is the seat's context, not the record's, so it is not compared.
func handNext(t *testing.T, writer, runDir, seatID string) {
	t.Helper()
	if seatID == "" {
		return
	}
	cmd := exec.Command(writer, "-run", runDir, "-phase", string(sittingwrite.Next),
		"-agent-id", fuzzAgentOf(seatID), "-agent-type", agentTypeOf(seatID), "-engine", handedEngine)
	// HERMETIC, as capture is: no live run on this machine may answer for the one the fuzz drives.
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+os.TempDir())
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		switch exit.ExitCode() {
		case hookgate.ExitDeliveryHeld:
			return
		case hookgate.ExitDeliveryStuck:
			t.Errorf("the writer could not advance %s's delivery of report text: %s", seatID, strings.TrimSpace(stderr.String()))
			return
		}
	}
	if err != nil {
		t.Errorf("the writer's next phase failed for %s: %v: %s", seatID, err, strings.TrimSpace(stderr.String()))
		return
	}
	if text := stdout.String(); strings.Contains(text, "COULD NOT BE READ") {
		t.Errorf("the writer could not read what %s is handed next: %s", seatID, firstLine(text))
	}
}

// owesReportText is whether seatID's sitting is owed report text still to be handed: the head of its
// queue is a report-text item, the writer's own test (WriteNextItem).
func owesReportText(runDir, seatID string) (bool, error) {
	run, err := record.NewRun(runDir)
	if err != nil {
		return false, err
	}
	w, err := record.WorkOfSeat(run, seatclass.RoleOfSeat(seatID), seatID)
	if err != nil {
		return false, err
	}
	q := record.QueueOf(w)
	return len(q) > 0 && q[0].Kind == record.ItemReportText, nil
}

// readDelivery is a seat reading what it is owed before an act that stands on delivered text: while
// report text is owed to its sitting, it calls the record (`show work`), and the writer hands it the
// next item after each call. It returns once the writer has handed the last item and written the
// sitting's report_read. Its calls are the drive's, not the sequence's: they are not tallied or
// compared, and their output names no id, so they move no placeholder.
func readDelivery(t *testing.T, bin, writer, runDir, seatID string) {
	t.Helper()
	for i := 0; ; i++ {
		owed, err := owesReportText(runDir, seatID)
		if err != nil {
			t.Errorf("what %s is owed does not read: %v", seatID, err)
			return
		}
		if !owed {
			return
		}
		if i == maxDeliveryCalls {
			t.Errorf("%s is still owed report text after %d calls to the record: the writer stopped handing it", seatID, maxDeliveryCalls)
			return
		}
		c := exec.Command(bin, "show", "work", "--run", runDir, "--seat-id", seatID)
		c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+os.TempDir())
		if out, err := c.CombinedOutput(); err != nil {
			t.Errorf("%s reading its work list: %v\n%s", seatID, err, out)
			return
		}
		handNext(t, writer, runDir, seatID)
	}
}
