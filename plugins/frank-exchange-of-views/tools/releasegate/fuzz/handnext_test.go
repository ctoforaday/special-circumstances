package fuzz

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// EVERY SEAT'S WORK ARRIVES THROUGH THE DELIVERY WRITER, one item after each of its calls — every call of the drive's is to the record —
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
	// the one handed before while that one is still queued (headMovesOnlyOnDischarge), with the record's
	// last event id, so a failure says whether anything landed between the oracle's reads.
	lens := strings.HasPrefix(seatID, "red-lens-")
	var reads queueReads
	sitting := ""
	var sittingID int64
	if lens {
		reads.beforeID, reads.idErr = r.lastEventID()
		// A LIST THAT DOES NOT READ IS SAID, never an oracle that silently holds nothing to account.
		if w, err := record.WorkOfSeat(r.run(), "lens", seatID); err != nil {
			r.noteActErr("queue oracle: " + seatID + "'s work list does not read before the writer runs: " + err.Error())
		} else {
			reads.before = queueSet(w)
			sittingID = w.SittingID()
			sitting = fmt.Sprintf("%s#%s#%d", seatID, agent, sittingID)
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
		var err error
		if reads.writerID, err = r.lastEventID(); err != nil && reads.idErr == nil {
			reads.idErr = err
		}
		// THE READ AFTER THE WRITER, made only where the oracle needs it; the last event id is read after
		// the list, so an id unmoved since the read before the writer means nothing landed between them.
		reads.after = func() (map[string]bool, int64, error) {
			w, err := record.WorkOfSeat(r.run(), "lens", seatID)
			id, idErr := r.lastEventID()
			if err == nil {
				err = idErr
			}
			return queueSet(w), id, err
		}
		reads.delivery = func() string { return r.storedDelivery(agent, sittingID) }
		reads.between = r.eventsBetween
		r.headMovesOnlyOnDischarge(seatID, sitting, text, reads)
	}
}

// queueSet is the keys queued on a lens's list.
func queueSet(w record.WorkJSON) map[string]bool {
	queued := map[string]bool{}
	for _, e := range record.QueueOf(w) {
		queued[e.Key] = true
	}
	return queued
}

// lastEventID is the id of the record's latest event.
func (r *runner) lastEventID() (int64, error) {
	db, err := recordsql.Open(filepath.Join(r.run().Records(), "record.db"))
	if err != nil {
		return 0, err
	}
	var id int64
	err = db.QueryRow(`SELECT COALESCE(MAX("id"), 0) FROM "events"`).Scan(&id)
	return id, err
}

// eventsBetween names the events after id a up to id b, each as its id, type and seat, for a failure to
// say what landed between two reads.
func (r *runner) eventsBetween(a, b int64) string {
	db, err := recordsql.Open(filepath.Join(r.run().Records(), "record.db"))
	if err != nil {
		return "not read: " + err.Error()
	}
	rows, err := db.Query(`SELECT "id", "type", "seat_id" FROM "events" WHERE "id" > ? AND "id" <= ? ORDER BY "id" LIMIT 40`, a, b)
	if err != nil {
		return "not read: " + err.Error()
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id int64
		var typ, seat string
		if err := rows.Scan(&id, &typ, &seat); err != nil {
			return "not read: " + err.Error()
		}
		out = append(out, fmt.Sprintf("%d %s %s", id, typ, seat))
	}
	if err := rows.Err(); err != nil {
		return "not read: " + err.Error()
	}
	return "[" + strings.Join(out, "; ") + "]"
}

// storedDelivery is what the writer stored as handed to agent's sitting — its current item's key and
// the queue it read — as the file's bytes, for a failure to show; never a verdict's input.
func (r *runner) storedDelivery(agent string, sitting int64) string {
	b, err := os.ReadFile(filepath.Join(r.runDir, ".work-delivery", fmt.Sprintf("%s-%d.json", agent, sitting)))
	if err != nil {
		return "not read: " + err.Error()
	}
	return strings.TrimSpace(string(b))
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
