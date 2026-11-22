//go:build windows

package peer

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

// ConversationBusy fails closed on Windows where Unix flock and /proc fd inspection are unavailable.
func ConversationBusy(string, string, int) (bool, string) {
	return true, "cannot verify presence lock or open conversation handles on this platform"
}

// renameRecord replaces dst with src, retrying briefly when a concurrent reader on Windows holds
// dst open and causes MoveFileEx to return ERROR_ACCESS_DENIED or ERROR_SHARING_VIOLATION.
func renameRecord(src, dst string) error {
	var err error
	for attempt := 0; attempt < 20; attempt++ {
		err = os.Rename(src, dst)
		if err == nil {
			return nil
		}
		if !errors.Is(err, windows.ERROR_ACCESS_DENIED) && !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}
