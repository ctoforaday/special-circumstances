package sittingwrite

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// Next is the phase that hands a seat its next item, from the PostToolUse hook after a call of the
// seat's, whichever tool (WriteNextItem).
const Next Phase = "next"

// deliveryDir holds, per seat sitting, what the seat was last handed: the current item's key and
// digest, the keys queued beside it, and the sitting's report-text delivery as cut. It lives under the
// run directory, keyed by the agent and its sitting; it is the writer's buffer, read by nothing else.
const deliveryDir = ".work-delivery"

// delivered is what a seat sitting was last handed.
type delivered struct {
	// Key and Digest are the current item's (record.QueueEntry); Key is "" when nothing was queued.
	Key    string `json:"key"`
	Digest string `json:"digest"`
	// Queue is every key queued when it was handed, in queue order.
	Queue []string `json:"queue"`
	// Report is the sitting's report-text delivery as cut at its first item, and how far it has been
	// handed: the cut text, kept so every later item comes from the one text (handReportText).
	Report *reportCut `json:"report,omitempty"`
	// Listed is whether the seat's whole list (standing, current item, index) has been handed this
	// sitting after report text held it back at the opening end.
	Listed bool `json:"listed,omitempty"`
}

// WriteNextItem prints what a seat is handed after any call of its, for the
// PostToolUse hook to pass to the seat — or nothing, where nothing it was handed has changed.
//
// THE CURRENT ITEM STAYS CURRENT UNTIL IT LEAVES THE QUEUE: an item that arrives ahead of it in queue
// order waits, because a seat moved off its work by every new arrival finishes none of it. So:
//
//   - the item it was handed is gone from its queue: the head of the queue is handed, with the index;
//   - the same item, its digest moved: the item again, marked changed_since_delivered;
//   - the same item, other keys gone from the queue: one line saying which left;
//   - nothing changed: nothing.
//
// A sitting with no record of what it was handed — the file lost, or a dispatch the opening end did
// not hand a queue — is handed the head of its queue. A list that cannot be read prints the
// unreadable statement, never silence: "I could not tell" and "nothing changed" are not the same
// bytes.
func WriteNextItem(runDir, agentID, agentType, engine string, out io.Writer) error {
	if err := deliveryAgent(agentID); err != nil {
		return err
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return nil
	}
	// THE SEAT IS THE RECORD'S BINDING OF THE AGENT, so a seat its configuration does not name is
	// reached from the call after its register; before it, the configuration names the seat where it
	// names one.
	seat, known, err := record.SeatOfAgent(run, agentID)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return nil
	}
	if !known {
		if seat, known = record.SeatOfAgentType(agentType); !known {
			return nil
		}
	}
	role := seatclass.RoleOfSeat(seat)
	w, err := record.WorkOfSeat(run, role, seat)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return nil
	}
	q := record.QueueOf(w)
	// REPORT TEXT IS HANDED ONLY WHERE IT REACHES THE SEAT'S CONTEXT AT ONCE: under Antigravity the
	// answer waits in hookcore's queue for the seat's next PreInvocation, so report-text entries are
	// passed over there and every other item is handed as here.
	if !DeliversReportText(engine) {
		q = slices.DeleteFunc(q, func(e record.QueueEntry) bool { return e.Kind == record.ItemReportText })
	}
	path := deliveryPath(runDir, agentID, w.SittingID())
	if len(q) > 0 && q[0].Kind == record.ItemReportText {
		return handReportItem(run, seat, w, q, path, out)
	}
	unlock, err := lockDelivery(path)
	if err != nil {
		return err
	}
	defer unlock()
	prev, had, err := readDelivered(path)
	if err != nil {
		return err
	}
	if !record.WorkArrivesAsItems(role, seat) && len(q) == 0 {
		return nil
	}
	var text string
	var now delivered
	if had && prev.Report != nil && !prev.Listed {
		// THE LIST ARRIVES WHOLE ONCE A SITTING: the opening end held it back behind the sitting's
		// report text, so the first item after that text carries the standing and the index with it.
		text, now = renderDispatch(w, q), delivered{Queue: queueKeys(q), Listed: true}
		if len(q) > 0 {
			now.Key, now.Digest = q[0].Key, q[0].Digest
		}
	} else {
		text, now = nextOf(w, q, prev, had)
		now.Listed = prev.Listed
	}
	if text == "" {
		return nil
	}
	now.Report = prev.Report
	if err := writeDelivered(path, now); err != nil {
		return err
	}
	_, err = io.WriteString(out, text)
	return err
}

// handReportItem hands the seat the next item of its sitting's report text. EVERY FAILURE HERE IS A
// DELIVERY THAT CANNOT ADVANCE — the seat's envelope waits on text no call can now hand it — so each
// returns ErrDeliveryStuck and is logged as the tool's defect exactly once, here; only a delivery
// another of the seat's calls holds is not (hookgate.ErrDeliveryHeld: that call is handing it).
func handReportItem(run record.Run, seat string, w record.WorkJSON, q []record.QueueEntry, path string, out io.Writer) error {
	err := func() error {
		unlock, err := lockDelivery(path)
		if errors.Is(err, hookgate.ErrDeliveryHeld) {
			return err
		}
		if err != nil {
			return stuck("the sitting's delivery file could not be locked: %v", err)
		}
		defer unlock()
		prev, _, err := readDelivered(path)
		if err != nil {
			return stuck("the sitting's delivery file could not be read: %v", err)
		}
		text, now, err := handReportText(run, seat, record.ReportTextSubjects(w), prev)
		if err != nil {
			return err
		}
		now.Key, now.Digest = q[0].Key, q[0].Digest
		now.Queue = queueKeys(q)
		now.Listed = prev.Listed
		if err := writeDelivered(path, now); err != nil {
			return stuck("the sitting's delivery file could not be written: %v", err)
		}
		_, err = io.WriteString(out, text)
		return err
	}()
	if reason, ok := IsDeliveryStuck(err); ok {
		if lerr := logDefect(run, seat, reason); lerr != nil {
			return stuck("%s; and its log could not be written: %v", reason, lerr)
		}
	}
	return err
}

// queueKeys is every entry's key, in queue order.
func queueKeys(q []record.QueueEntry) []string {
	keys := make([]string, len(q))
	for i, e := range q {
		keys[i] = e.Key
	}
	return keys
}

// nextOf decides what is handed after a call, given what was handed before, and what is then on
// record as handed.
func nextOf(w record.WorkJSON, q []record.QueueEntry, prev delivered, had bool) (string, delivered) {
	keys := make([]string, len(q))
	at := -1
	for i, e := range q {
		keys[i] = e.Key
		if had && e.Key == prev.Key && prev.Key != "" {
			at = i
		}
	}
	switch {
	case at >= 0 && q[at].Digest != prev.Digest:
		return renderNext(w, q, at, true), delivered{Key: q[at].Key, Digest: q[at].Digest, Queue: keys}
	case at >= 0:
		var left []string
		for _, k := range prev.Queue {
			if !slices.Contains(keys, k) {
				left = append(left, k)
			}
		}
		if len(left) == 0 {
			return "", prev
		}
		return fmt.Sprintf("%s left your list; your current item is still %s.\n", joinKeys(left), prev.Key),
			delivered{Key: prev.Key, Digest: prev.Digest, Queue: keys}
	case len(q) > 0:
		return renderNext(w, q, 0, false), delivered{Key: q[0].Key, Digest: q[0].Digest, Queue: keys}
	case had && prev.Key == "":
		return "", prev
	default:
		gone := "nothing is queued for you"
		if had {
			gone = prev.Key + " left your list, and nothing more is queued for you"
		}
		return fmt.Sprintf("%s: whether anything blocks you finishing is `complete` on your work list (now %v).\n", gone, w.Sitting.Complete),
			delivered{Queue: []string{}}
	}
}

func joinKeys(ks []string) string {
	b, _ := json.Marshal(ks)
	return string(b)
}

// deliveryAgent refuses an agent id that cannot name a file in deliveryDir.
func deliveryAgent(agentID string) error {
	if agentID == "" || agentID == "." || agentID == ".." || filepath.Base(agentID) != agentID {
		return fmt.Errorf("sittingwrite: %q names no agent whose delivery can be stored", agentID)
	}
	return nil
}

// deliveryPath is the file holding what one seat sitting was handed.
func deliveryPath(runDir, agentID string, sitting int64) string {
	return filepath.Join(runDir, deliveryDir, agentID+"-"+strconv.FormatInt(sitting, 10)+".json")
}

func readDelivered(path string) (delivered, bool, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return delivered{}, false, nil
	}
	if err != nil {
		return delivered{}, false, err
	}
	var d delivered
	if json.Unmarshal(b, &d) != nil {
		// A file that does not parse is a lost one: the head of the queue is handed again.
		return delivered{}, false, nil
	}
	return d, true, nil
}

func writeDelivered(path string, d delivered) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// lockDelivery holds an O_EXCL marker beside the sitting's file while what was handed is compared and
// rewritten, so two of the seat's calls returning together hand it the next item once. A marker
// older than staleLock is a holder that died, and is taken over; one still held past lockWait is
// hookgate.ErrDeliveryHeld, since the holder is handing the item.
func lockDelivery(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	lock := path + ".lock"
	deadline := time.Now().Add(lockWait)
	for {
		f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			f.Close()
			return func() { os.Remove(lock) }, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
		if fi, serr := os.Stat(lock); serr == nil && time.Since(fi.ModTime()) > staleLock {
			os.Remove(lock)
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w (%s) past %v", hookgate.ErrDeliveryHeld, lock, lockWait)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// lockWait is how long a call waits on another's hold before it hands nothing (a variable so a test
// need not wait it out); a hold older than staleLock is a holder that died.
var lockWait = 5 * time.Second

const staleLock = 30 * time.Second

// recordDelivered stores what the opening end handed the seat, so the first call after it compares
// against that.
func recordDelivered(runDir, agentID string, w record.WorkJSON, q []record.QueueEntry) error {
	if err := deliveryAgent(agentID); err != nil {
		return err
	}
	path := deliveryPath(runDir, agentID, w.SittingID())
	unlock, err := lockDelivery(path)
	if err != nil {
		return err
	}
	defer unlock()
	d := delivered{Queue: []string{}}
	for _, e := range q {
		d.Queue = append(d.Queue, e.Key)
	}
	if len(q) > 0 {
		d.Key, d.Digest = q[0].Key, q[0].Digest
	}
	return writeDelivered(path, d)
}
