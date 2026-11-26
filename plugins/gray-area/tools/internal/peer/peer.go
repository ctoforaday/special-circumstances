// Package peer delivers a message to another agent session on this box, whichever harness runs it.
//
// Contract (Design by Contract):
//
//	A delivery MUST say which route carried it and what that route promises: a Claude Code
//	session's inbox socket hands the text to the running session; an Antigravity conversation's
//	language server wakes it; an Antigravity mailbox only queues it for the conversation's next
//	turn. Those are three different facts and a sender that cannot tell them apart will wait for
//	a reply that is not coming.
//	A target that cannot be resolved to exactly one session MUST be refused, never guessed.
//
// # The routes, and why each is the one it is
//
// Claude Code advertises each running session in ~/.claude/sessions/<pid>.json, and the
// advertisement names messagingSocketPath: a Unix socket that takes one NDJSON user message per
// line and needs no token. That is the whole Claude route.
//
// Antigravity serves every conversation it hosts from a language server on localhost, whose
// SendAgentMessage method delivers a message and wakes an idle conversation. The server demands an
// x-codeium-csrf-token header, minted per launch and never written down by Antigravity. It reaches
// exactly one place outside the server: the environment of the shell commands an agent runs
// (ANTIGRAVITY_LS_ADDRESS, ANTIGRAVITY_CSRF_TOKEN, ANTIGRAVITY_CONVERSATION_ID). Hooks receive
// none of the three and MCP servers receive no ANTIGRAVITY_ variable at all (both observed with
// agy 1.5.0). A PreToolUse hook could prefix every command to capture them, but its overwrite
// needs a decision and "allow" auto-approves the command, so that route would silently widen the
// agent's permissions. What is left is the honest one: when telepathy itself runs inside an
// Antigravity shell, it records that conversation's server (Register). A conversation that has
// never run telepathy has no record, and gets the mailbox.
//
// A record is keyed by server and lists the conversations that linked there, each with when it
// last linked. Any server ADOPTS a conversation it does not host on SendAgentMessage and starts
// running it there, so the one thing a send must never do is reach a conversation another process
// already has open: two processes writing one conversation is the defect this package guards.
//
// # Choosing the server
//
// Send first asks who holds the conversation (ConversationBusy, excluding nobody): its presence
// lock, or its conversations/<id>.db open in any process. Then:
//
//   - Nobody holds it: the Live record listing it that it linked to last, then the current
//     remote-control daemon's record, which adopts it. Each attempt is preceded by the same probe,
//     excluding only that server's pid, so a server that took the conversation while an earlier
//     attempt failed stops the send rather than meeting a second writer.
//   - A Live record's server holds it: that server delivers, once a probe excluding its pid finds
//     nobody else. It is never bypassed for another server, because it has the conversation open.
//   - The current daemon holds it and has no record: nothing can reach it until a conversation on
//     it links, and the message queues saying so.
//   - Anything else holds it, or the probe cannot inspect: the message queues, naming the holder.
//
// A record is a target only while RecordLiveness reads it Live. Route names which of the three
// kinds of record carried the message.
//
// The mailbox is Antigravity's own on-disk queue: messages/<id>.json under the conversation's
// brain directory, plus an empty undelivered/<id> marker. The client reads it at the
// conversation's next turn, marks the id in read.json and removes the marker. Writing it needs no
// token and runs no model.
package peer

import (
	"crypto/rand"
	"fmt"
)

// Route names how a message travelled.
type Route string

const (
	RouteClaudeSocket Route = "claude-socket"
	// RouteAgyDaemon: the current remote-control daemon's record.
	RouteAgyDaemon Route = "antigravity-daemon"
	// RouteAgyOwnRecord: a record that lists the conversation, other than the daemon's.
	RouteAgyOwnRecord Route = "antigravity-server-own"
	// RouteAgyHolderRecord: the record of the server holding the conversation, which does not list
	// it and is not the daemon's.
	RouteAgyHolderRecord Route = "antigravity-server-holder"
	RouteAgyMailbox      Route = "antigravity-mailbox"
)

// Delivery is what one send did.
type Delivery struct {
	Route Route
	// Via is the socket path, the server address, or the mailbox file.
	Via string
	// NotWoken says why an Antigravity message was queued rather than delivered live; empty on
	// every other route.
	NotWoken string
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
