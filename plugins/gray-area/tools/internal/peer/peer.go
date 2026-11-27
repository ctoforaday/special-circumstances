// Package peer delivers a message to another agent session on this box, whichever harness runs it.
//
// Contract (Design by Contract):
//
//	A delivery MUST say which route carried it and what that route promises: a Claude Code
//	session's inbox socket hands the text to the running session; an Antigravity mailbox queues
//	it for the conversation's next turn, and a wake, when one is sent, starts that turn now. A
//	sender that cannot tell these apart will wait for a reply that is not coming.
//	A target that cannot be resolved to exactly one session MUST be refused, never guessed.
//
// # The routes, and why each is the one it is
//
// Claude Code advertises each running session in ~/.claude/sessions/<pid>.json, and the
// advertisement names messagingSocketPath: a Unix socket that takes one NDJSON user message per
// line and needs no token. That is the whole Claude route.
//
// An Antigravity message's content always goes to the conversation's mailbox: messages/<id>.json
// under its brain directory, plus an empty undelivered/<id> marker. The client reads it at the
// start of the conversation's next turn, marks the id in read.json and removes the marker. Writing
// it needs no token and runs no model, and the content cannot arrive twice or split the
// conversation between two servers.
//
// The turn that reads it can be started now: the remote-control daemon's language server takes a
// SendAgentMessage request carrying its CSRF token, and a fixed wake message (WakeContent) starts
// a turn whose first act is to read the mailbox (measured on agy 1.5.18). The token reaches no hook,
// MCP server or shell, so the daemon is started with ANTIGRAVITY_SERVER_CSRF_TOKEN, which its
// language server adopts, from a file the human provisions (ReadDaemonToken). A language server
// sent a message for a conversation another process holds adopts it anyway, so a wake goes only
// where ConversationBusy, excluding the daemon, finds no holder.
//
// Stated limit: the holder check sees only the processes of telepathy's own pid namespace, and a
// server that has resumed a conversation without yet opening its database. A process outside it
// that holds the conversation — an agy on the host while telepathy runs in a container sharing its
// data — is invisible, and a wake can then create a second writer.
package peer

import (
	"crypto/rand"
	"fmt"
)

// Route names how a message travelled.
type Route string

const (
	RouteClaudeSocket Route = "claude-socket"
	RouteAgyMailbox   Route = "antigravity-mailbox"
)

// Delivery is what one send did.
type Delivery struct {
	Route Route
	// Via is the socket path or the mailbox file.
	Via string
}

// newID returns a random RFC 4122 version 4 UUID, the shape Antigravity's own message ids take.
func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
