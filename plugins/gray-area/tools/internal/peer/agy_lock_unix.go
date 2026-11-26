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

// ConversationBusy reports the process, other than excludePID, that holds conversation: one holding
// its presence lock (<agyRoot>/presence/<conversation>.lock, named by its pid from /proc/locks), or
// one with <agyRoot>/conversations/<conversation>.db open (Linux). An excludePID of 0 excludes
// nobody.
//
// Fails CLOSED: any unexpected error opening or probing the lock file, reading /proc, or
// inspecting a same-UID process's file descriptors is HolderUninspectable, so telepathy send queues
// to the mailbox rather than risking a second writer.
//
// Check-then-send race: a second process can still open the conversation in the window between
// this check and the HTTP SendAgentMessage call. This guard prevents sending into an already-open
// second writer; eliminating the race completely requires a server-side lock inside Antigravity.
func ConversationBusy(agyRoot, conversation string, excludePID int) Holder {
	if agyRoot == "" || !safeName(conversation) {
		return Holder{Kind: HolderUninspectable, Detail: "the conversation's path cannot be validated"}
	}
	lockPath := filepath.Join(agyRoot, "presence", conversation+".lock")
	fl := flock.New(lockPath, flock.SetFlag(os.O_RDONLY))
	ok, err := fl.TryLock()
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return Holder{Kind: HolderUninspectable, Detail: fmt.Sprintf("cannot inspect presence/%s.lock: %v", conversation, err)}
		}
	} else if !ok {
		pid, err := flockHolder(lockPath, excludePID)
		if err != nil {
			return Holder{Kind: HolderUninspectable, Detail: fmt.Sprintf("presence/%s.lock is held and its holder cannot be read: %v", conversation, err)}
		}
		if pid != 0 {
			return Holder{PID: pid, Kind: HolderPresenceLock, Detail: fmt.Sprintf("pid %d holds presence/%s.lock", pid, conversation)}
		}
	} else {
		_ = fl.Unlock()
	}
	return otherDBHolder(agyRoot, conversation, excludePID)
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
