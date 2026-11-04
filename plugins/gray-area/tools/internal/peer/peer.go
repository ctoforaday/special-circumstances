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
// The record is keyed by conversation, never by server alone: a server that does not host a
// conversation ADOPTS it on SendAgentMessage and starts running it there, so sending a
// conversation to any server but the one its own shell reported is a defect, not a fallback.
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
	RouteAgyServer    Route = "antigravity-server"
	RouteAgyMailbox   Route = "antigravity-mailbox"
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
