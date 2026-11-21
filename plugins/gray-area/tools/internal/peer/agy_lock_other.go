//go:build !unix && !windows

package peer

import (
	"errors"
	"os"
)

// ConversationBusy fails closed on platforms where flock and /proc are unavailable.
func ConversationBusy(string, string, int) (bool, string) {
	return true, "cannot verify presence lock or open conversation handles on this platform"
}

// lockRegistry fails closed on platforms without an implemented file lock rather than returning a
// silent no-op that would lose concurrent updates.
func lockRegistry(string) (func(), error) {
	return nil, errors.New("server registry file locking is not implemented on this platform")
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
