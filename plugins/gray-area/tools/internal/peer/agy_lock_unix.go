//go:build unix

package peer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ConversationBusy reports the process, other than excludePID, that holds conversation: one holding
// its presence lock (<agyRoot>/presence/<conversation>.lock, named by its pid from /proc/locks), or
// one with <agyRoot>/conversations/<conversation>.db open (Linux). An excludePID of 0 excludes
// nobody.
//
// It never takes, or even opens, the presence lock: agy honours that lock, and a probe that held it
// for an instant could make agy itself fail to open the conversation. The holders are read from
// /proc/locks instead.
//
// Fails CLOSED: any unexpected error inspecting the lock file, reading /proc, or inspecting a
// same-UID process's file descriptors is HolderUninspectable, so telepathy send queues to the
// mailbox rather than risking a second writer.
//
// Check-then-send race: a second process can still open the conversation in the window between
// this check and the HTTP SendAgentMessage call. This guard prevents sending into an already-open
// second writer; eliminating the race completely requires a server-side lock inside Antigravity.
func ConversationBusy(agyRoot, conversation string, excludePID int) Holder {
	if agyRoot == "" || !safeName(conversation) {
		return Holder{Kind: HolderUninspectable, Detail: "the conversation's path cannot be validated"}
	}
	lockPath := filepath.Join(agyRoot, "presence", conversation+".lock")
	holders, err := flockHolders(lockPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// No lock file: nothing holds the presence lock.
	case err != nil:
		return Holder{Kind: HolderUninspectable, Detail: fmt.Sprintf("cannot inspect presence/%s.lock: %v", conversation, err)}
	}
	if h := holderFromLocks(conversation, holders, excludePID); h.Kind != HolderNone {
		return h
	}
	return otherDBHolder(agyRoot, conversation, excludePID)
}

// holderFromLocks is the presence lock's holder other than excludePID, from its holders' pids.
func holderFromLocks(conversation string, holders []int, excludePID int) Holder {
	for _, pid := range holders {
		if pid <= 0 {
			return Holder{Kind: HolderUninspectable, Detail: fmt.Sprintf("presence/%s.lock is held by a process outside this pid namespace", conversation)}
		}
		if pid != excludePID {
			return Holder{PID: pid, Kind: HolderPresenceLock, Detail: fmt.Sprintf("pid %d holds presence/%s.lock", pid, conversation)}
		}
	}
	return Holder{Kind: HolderNone}
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
