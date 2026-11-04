package peer

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"time"
)

// nativeWindows is whether the inbox is a named pipe that demands the RECIPIENT's own token as its
// first line (code.claude.com/docs/en/cross-session-messaging, "The session's inbox socket"). Only
// the recipient's own child processes hold that token, so no other process can post there. A
// variable so the refusal is testable on every platform.
var nativeWindows = runtime.GOOS == "windows"

// claudeLine is one message on a Claude Code session's inbox socket: the user-message shape the
// session's own input stream uses.
type claudeLine struct {
	Type    string `json:"type"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

// SendClaude writes text to a running Claude Code session's inbox socket.
//
// The post carries no auth line: on macOS and Linux it is optional, and the token it would carry
// is the recipient's own, which proves only that the sender is the recipient's child. The
// recipient's inbound controls then decide what happens to it, and nothing on the socket reports
// their decision (see the send verb's help).
func SendClaude(socket, text string) (Delivery, error) {
	if nativeWindows {
		return Delivery{}, fmt.Errorf("on native Windows a session's inbox accepts only its own child processes, which hold its token; ask the session's human to relay instead")
	}
	if socket == "" {
		return Delivery{}, fmt.Errorf("the session advertises no messagingSocketPath, so it has no inbox to write to")
	}
	var line claudeLine
	line.Type = "user"
	line.Message.Role = "user"
	line.Message.Content = text
	b, err := json.Marshal(line)
	if err != nil {
		return Delivery{}, err
	}
	conn, err := net.DialTimeout("unix", socket, 5*time.Second)
	if err != nil {
		return Delivery{}, fmt.Errorf("its inbox socket %s refused the connection: %w", socket, err)
	}
	defer conn.Close()
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write(append(b, '\n')); err != nil {
		return Delivery{}, fmt.Errorf("writing to its inbox socket %s: %w", socket, err)
	}
	return Delivery{Route: RouteClaudeSocket, Via: socket}, nil
}
