//go:build !unix && !windows

package peer

import "os"

// ConversationBusy fails closed on platforms where flock and /proc are unavailable.
func ConversationBusy(string, string, int) Holder {
	return Holder{Kind: HolderUninspectable, Detail: "presence locks and open conversation handles cannot be inspected on this platform"}
}

func renameRecord(src, dst string) error { return os.Rename(src, dst) }
