//go:build unix

package peer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
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
	f, err := os.Open(lockPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return true, fmt.Sprintf("cannot inspect presence/%s.lock: %v", conversation, err)
		}
	} else {
		flockErr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if flockErr != nil {
			f.Close()
			if errors.Is(flockErr, syscall.EWOULDBLOCK) || errors.Is(flockErr, syscall.EAGAIN) {
				if !flockHeldOnlyBy(lockPath, excludePID) {
					return true, fmt.Sprintf("presence/%s.lock is held by a live process", conversation)
				}
			} else {
				return true, fmt.Sprintf("cannot probe presence/%s.lock: %v", conversation, flockErr)
			}
		} else {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			f.Close()
		}
	}
	if busy, why := otherDBHolder(agyRoot, conversation, excludePID); busy {
		return true, why
	}
	return false, ""
}

// lockRegistry serialises concurrent RegisterWith read-modify-write updates in dir.
func lockRegistry(dir string) (func(), error) {
	f, err := os.OpenFile(filepath.Join(dir, ".register.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
