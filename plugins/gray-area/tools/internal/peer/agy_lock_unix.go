//go:build unix

package peer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// ConversationBusy reports whether another live process holds conversation's presence lock
// (<agyRoot>/presence/<conversation>.lock) or, on Linux, has <agyRoot>/conversations/<conversation>.db
// open under any PID other than excludePID.
//
// Fails CLOSED: any unexpected error opening or probing the lock file, reading /proc, or
// inspecting a same-UID process's file descriptors is treated as busy so telepathy send queues to
// the mailbox rather than risking a second writer.
//
// Check-then-send race: a second process can still open the conversation in the window between
// this check and the HTTP SendAgentMessage call. This guard prevents sending into an already-open
// second writer; eliminating the race completely requires a server-side lock inside Antigravity.
func ConversationBusy(agyRoot, conversation string, excludePID int) (bool, string) {
	if agyRoot == "" || !safeName(conversation) {
		return true, "conversation path cannot be validated for second-writer safety"
	}
	lockPath := filepath.Join(agyRoot, "presence", conversation+".lock")
	fl := flock.New(lockPath, flock.SetFlag(os.O_RDONLY))
	ok, err := fl.TryLock()
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return true, fmt.Sprintf("cannot inspect presence/%s.lock: %v", conversation, err)
		}
	} else if !ok {
		if !flockHeldOnlyBy(lockPath, excludePID) {
			return true, fmt.Sprintf("presence/%s.lock is held by a live process", conversation)
		}
	} else {
		_ = fl.Unlock()
	}
	if busy, why := otherDBHolder(agyRoot, conversation, excludePID); busy {
		return true, why
	}
	return false, ""
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
