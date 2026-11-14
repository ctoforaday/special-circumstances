package difftest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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

// replayDrive is what a drive step (cmd.do) acts through: the replay's run, its binaries, and the
// test its failures are reported on. Its calls are the drive's, not the sequence's: none is tallied as
// an arm or compared, and none is normalized, so none takes a placeholder from the sequence's own.
type replayDrive struct {
	t                   *testing.T
	bin, writer, runDir string
	heads               int // head moves this drive made, for each part's key
}

// call runs one command a seat types, as the seat does, and hands the seat its next item after it.
// A refusal fails the test: a drive step builds a board the arms stand on, and a step the record
// refused has built nothing.
func (d *replayDrive) call(args ...string) {
	d.t.Helper()
	c := exec.Command(d.bin, substitute(args, d.runDir)...)
	c.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+os.TempDir())
	if out, err := c.CombinedOutput(); err != nil {
		d.t.Errorf("drive step `feov-record %s`: %v\n%s", strings.Join(args, " "), err, out)
		return
	}
	handNext(d.t, d.writer, d.runDir, seatOfArgs(args))
}

// owes is whether seatID's sitting is owed report text still to be handed.
func (d *replayDrive) owes(seatID string) bool {
	d.t.Helper()
	owed, err := owesReportText(d.runDir, seatID)
	if err != nil {
		d.t.Errorf("what %s is owed does not read: %v", seatID, err)
	}
	return owed
}

// read is a seat reading what it is owed before an act that stands on delivered text: while report
// text is owed to its sitting, it reads its work list (`show work`), a call to the record, and the
// writer hands it the next item after it. It returns once the writer has handed the last item and
// written the sitting's report_read.
func (d *replayDrive) read(seatID string) {
	d.t.Helper()
	for i := 0; d.owes(seatID); i++ {
		if i == maxDeliveryCalls {
			d.t.Errorf("%s is still owed report text after %d calls to the record: the writer stopped handing it", seatID, maxDeliveryCalls)
			return
		}
		d.call("show", "work", "--run", "{RUN}", "--seat-id", seatID)
	}
}

// headMove appends a part to the report under the harness, as an ingest does: a head move every
// lens's pin is then behind.
func (d *replayDrive) headMove() {
	d.t.Helper()
	d.heads++
	recordtest.Seed(d.t, d.runDir, recordtest.At(d.t, record.HarnessSeat, fmt.Sprintf("%s:fuzz:head-move:%d", record.HarnessSeat, d.heads),
		&recordpb.ReportPart{Text: proto.String(fmt.Sprintf("\n## Addendum %d\n\nA paragraph the report gained after the lens's pin.\n", d.heads))}))
}

// readStep is the drive step in which seatID reads what its sitting is owed.
func readStep(seatID string) cmd {
	return cmd{drive: seatID + " reads its delivery", do: func(d *replayDrive) { d.read(seatID) }}
}

// chairOwesAStaleArea is the drive step that brings the tour's cast run to the state a chair's
// spot-check stands on: lens's area is stale on the chair's open sitting, and the chair has read the
// changes behind its pin. The chair is delivered report text only for a stale area's changes it has
// not borne, and a spot-check stands only on a sitting that holds that delivery, so a chair with
// nothing stale has no spot-check to record. Each of lens's sittings follows a chair dispatch naming
// it, as the workflow sends it, and takes the delivery it is owed (both whole deliveries first, then
// the changes since its pin); between sittings the head moves. Three sittings in a row with no fresh
// material mint retire the lens at its pin, and the next head move leaves its area behind that pin.
func chairOwesAStaleArea(lens string) cmd {
	return cmd{drive: "the chair reads " + lens + "'s stale area", do: func(d *replayDrive) {
		d.t.Helper()
		for i := 0; i < 6; i++ {
			if i > 1 {
				d.headMove()
			}
			d.call("register", "--run", "{RUN}", "--seat-id", "red-chair")
			d.call("dispatch", "next", "--run", "{RUN}", "--seat-id", "red-chair")
			if d.staleOnChair(lens) {
				d.read("red-chair")
				return
			}
			d.call("register", "--run", "{RUN}", "--seat-id", lens)
			d.read(lens)
		}
		d.t.Errorf("the drive left no stale area: after six sittings of %s the chair's list names none", lens)
	}}
}

// staleOnChair is whether the chair's open sitting lists lens's area as stale.
func (d *replayDrive) staleOnChair(lens string) bool {
	d.t.Helper()
	run, err := record.NewRun(d.runDir)
	if err != nil {
		d.t.Errorf("opening the run to read the chair's list: %v", err)
		return false
	}
	w, err := record.WorkOfSeat(run, "chair", "red-chair")
	if err != nil {
		d.t.Errorf("reading the chair's list: %v", err)
		return false
	}
	for _, it := range w.Sitting.Open {
		if it.Kind == record.ItemKind(record.BlockerStaleArea) && it.Subject != nil && it.Subject.ID == lens {
			return true
		}
	}
	return false
}
