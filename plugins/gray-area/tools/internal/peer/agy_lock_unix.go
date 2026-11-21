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
// open under a PID other than excludePID. Sending through the daemon when true would create a
// second writer.
func ConversationBusy(agyRoot, conversation string, excludePID int) (bool, string) {
	if agyRoot == "" || !safeName(conversation) {
		return false, ""
	}
	lockPath := filepath.Join(agyRoot, "presence", conversation+".lock")
	if f, err := os.Open(lockPath); err == nil {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			f.Close()
			if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
				return true, fmt.Sprintf("presence/%s.lock is held by another live process", conversation)
			}
		} else {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			f.Close()
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		// Unreadable lock file: leave to DB-holder check rather than falsely blocking.
	}
	if pid, held := otherAgyDBHolder(agyRoot, conversation, excludePID); held {
		return true, fmt.Sprintf("another Antigravity process (pid %d) already has conversations/%s.db open", pid, conversation)
	}
	return false, ""
}
