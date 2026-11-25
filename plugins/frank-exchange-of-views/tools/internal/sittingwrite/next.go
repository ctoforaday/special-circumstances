package sittingwrite

import (
	"context"
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

	"github.com/gofrs/flock"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// Next is the phase that hands a lens its next item, from the PostToolUse hook after a call of the
// lens's that went to the record (WriteNextItem).
const Next Phase = "next"

// deliveryDir holds, per lens sitting, what the lens was last handed: the current item's key and
// digest, and the keys queued beside it. It lives under the run directory, keyed by the agent and
// its sitting, and it holds identities only, never content.
const deliveryDir = ".work-delivery"

// delivered is what a lens sitting was last handed.
type delivered struct {
	// Key and Digest are the current item's (record.QueueEntry); Key is "" when nothing was queued.
	Key    string `json:"key"`
	Digest string `json:"digest"`
	// Queue is every key queued when it was handed, in queue order.
	Queue []string `json:"queue"`
}

// WriteNextItem prints what a lens is handed after a call of its that went to the record, for the
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
func WriteNextItem(runDir, agentID, agentType string, out io.Writer) error {
	if err := deliveryAgent(agentID); err != nil {
		return err
	}
	seat, known := record.SeatOfAgentType(agentType)
	if !known || seatclass.RoleOfSeat(seat) != "lens" {
		return nil
	}
	run, err := record.NewRun(runDir)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return nil
	}
	w, err := record.WorkOfSeat(run, "lens", seat)
	if err != nil {
		fmt.Fprintf(out, "%s\n\n%v\n", unreadable, err)
		return nil
	}
	path := deliveryPath(runDir, agentID, w.SittingID())
	unlock, err := lockDelivery(path)
	if err != nil {
		return err
	}
	defer unlock()
	q := record.QueueOf(w)
	prev, had, err := readDelivered(path)
	if err != nil {
		return err
	}
	text, now := nextOf(w, q, prev, had)
	if text == "" {
		return nil
	}
	if err := writeDelivered(path, now); err != nil {
		return err
	}
	_, err = io.WriteString(out, text)
	return err
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

// deliveryPath is the file holding what one lens sitting was handed.
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

// lockDelivery holds the lock file beside the sitting's file exclusive, through gofrs/flock, while
// what was handed is compared and rewritten, so two of the lens's calls returning together hand it
// the next item once. It waits up to lockWait. The operating system releases a lock when its
// holder's process exits, so a call that dies holding it blocks no one after it. The lock file stays
// after release: removing it would let a waiter lock the unlinked file while the next call locks a
// new one under the same name, and both would hand the item.
func lockDelivery(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	lock := path + ".lock"
	fl := flock.New(lock, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o644))
	ctx, cancel := context.WithTimeout(context.Background(), lockWait)
	defer cancel()
	ok, err := fl.TryLockContext(ctx, 20*time.Millisecond)
	switch {
	case ok:
		// A release that fails leaves the lock to the operating system, which drops it when this
		// process exits.
		return func() { _ = fl.Close() }, nil
	case errors.Is(err, context.DeadlineExceeded):
		return nil, fmt.Errorf("the lens's delivery is held by another call (%s) past %v", lock, lockWait)
	default:
		return nil, err
	}
}

const lockWait = 5 * time.Second

// recordDelivered stores what the opening end handed the lens, so the first call after it compares
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
