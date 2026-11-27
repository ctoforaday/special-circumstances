//go:build windows

package peer

// ConversationBusy fails closed on Windows where Unix flock and /proc fd inspection are unavailable.
func ConversationBusy(string, string, int) Holder {
	return Holder{Kind: HolderUninspectable, Detail: "presence locks and open conversation handles cannot be inspected on this platform"}
}
