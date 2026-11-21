//go:build !unix

package peer

// ConversationBusy fails closed on non-Unix platforms where flock and /proc are unavailable.
func ConversationBusy(string, string, int) (bool, string) {
	return true, "cannot verify presence lock or open conversation handles on this platform"
}

func lockRegistry(string) (func(), error) { return func() {}, nil }
