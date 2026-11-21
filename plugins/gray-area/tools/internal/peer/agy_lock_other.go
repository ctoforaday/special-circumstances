//go:build !unix

package peer

// ConversationBusy reports not busy on non-Unix platforms where flock and /proc are unavailable.
func ConversationBusy(string, string, int) (bool, string) { return false, "" }
