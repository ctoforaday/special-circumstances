//go:build unix

package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
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
// mailbox rather than risking a second writer. So is a lock /proc/locks may not show whole — a
// locker that has exited, a network filesystem (flockHolders says when).
//
// Stated limit: it sees only the processes of this procfs's pid namespace. A holder outside it is
// invisible, so a container that shares Antigravity's data with the host can create a second
// writer.
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
	for _, pid := range holders {
		if pid != excludePID {
			return Holder{PID: pid, Kind: HolderPresenceLock, Detail: fmt.Sprintf("pid %d holds presence/%s.lock", pid, conversation)}
		}
	}
	return otherDBHolder(agyRoot, conversation, excludePID)
}

// safeName refuses an id that could name a path outside the directory it is joined to.
func safeName(id string) bool {
	return id != "" && id != "." && id != ".." && filepath.Base(id) == id && !bytes.ContainsAny([]byte(id), `/\`)
}
