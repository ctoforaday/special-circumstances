//go:build !unix && !windows

package peer

import "os"

// ConversationBusy fails closed on platforms where flock and /proc are unavailable.
func ConversationBusy(string, string, int) (bool, string) {
	return true, "cannot verify presence lock or open conversation handles on this platform"
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
