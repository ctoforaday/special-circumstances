package telecli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

// candidate is one session an id prefix could name.
type candidate struct {
	harness string // "claude" or "antigravity"
	id      string
	socket  string // claude: the inbox socket
	dir     string // antigravity: the conversation's brain directory
}

func newSendCmd(env *Env) *cobra.Command {
	var title string
	cmd := &cobra.Command{
		Use:   "send <session-id> <message|->",
		Short: "deliver a message to another agent session, Claude Code or Antigravity",
		Long: `Delivers a message to another running agent session on this box, whichever
harness runs it. The id is a session id or any unique prefix of one; "-" reads the
message from stdin. The message arrives prefixed with who sent it and the command
that replies. When the recipient is an Antigravity conversation with a local annotation
file, its current title is printed beside the id in the delivery report (display only;
titles are never accepted as targets).

WHICH ROUTE CARRIED IT IS PRINTED, because the routes promise different things:

  claude-socket        a running Claude Code session's inbox socket. The session's
                       inbound controls then decide: by default the text reaches it
                       between tool calls, or starts a turn if it is idle; a session
                       that bypasses permission prompts HOLDS it for its human's
                       approval; crossSessionInbound "refuse" drops it unseen.
  antigravity-server   the language server hosting the Antigravity conversation. It
                       delivers the text and wakes the conversation if it is idle.
  antigravity-mailbox  the conversation's on-disk mailbox. The text is QUEUED: it is
                       read at the conversation's next turn, and nothing wakes it. The
                       reason it was not delivered live is printed with it.

Live delivery to an Antigravity conversation runs on Linux only; on other platforms
second-writer safety cannot be verified, so send always queues to the mailbox. On
Linux, an Antigravity conversation is reached live through its own server record when
that server is up, or through the current remote-control daemon's server record —
written by any daemon-hosted conversation that linked since the daemon started — for
conversations under the CLI daemon root (~/.gemini/antigravity-cli). Every live route
checks that presence/<id>.lock is not held and no other process has conversations/<id>.db
open (which would create a second writer). The server's token reaches nothing but an
agent's shell commands, so when no daemon-hosted conversation has linked since a
restart (a standalone interactive agy does not restore the daemon route), send queues
to the mailbox and says so. telepathy link --help says where the record lives.

WHAT IT CANNOT TELL YOU. Delivery is not reading: no route reports that the recipient
acted on the message, and the Claude Code socket does not report a hold or a drop. A
Claude Code session that is not running has no inbox and is refused; telepathy
agents shows which are. On native Windows a Claude Code inbox takes posts only from
its own session, so sending to one is refused there.`,
		Example: `  telepathy send 8a3d5199 "PR #1730 conflicts with mine; rebase after it merges"
  git diff --stat | telepathy send f13f9d4f -`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[1]
			if text == "-" {
				b, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				text = string(b)
			}
			if strings.TrimSpace(text) == "" {
				return usagef("the message is empty")
			}
			target, err := resolvePeer(env, args[0])
			if err != nil {
				return err
			}
			from, reply := env.sender()
			body := fmt.Sprintf("[telepathy] message from %s; %s\n\n%s", from, reply, text)
			if title == "" {
				title = "Message from " + from
			}
			d, err := deliver(env, target, body, title)
			if err != nil {
				return fmt.Errorf("%s session %s: %w", target.harness, target.id, err)
			}
			out := cmd.OutOrStdout()
			agyTitleSuffix := ""
			if target.harness == "antigravity" {
				if t := env.agyTitle(target.id); t != "" {
					agyTitleSuffix = fmt.Sprintf(" (%q)", t)
				}
			}
			switch d.Route {
			case peer.RouteClaudeSocket:
				fmt.Fprintf(out, "delivered to claude session %s over its inbox socket %s\n", target.id, d.Via)
			case peer.RouteAgyServer:
				fmt.Fprintf(out, "delivered to antigravity conversation %s%s through its language server at %s; it wakes if idle\n", target.id, agyTitleSuffix, d.Via)
			case peer.RouteAgyMailbox:
				fmt.Fprintf(out, "QUEUED for antigravity conversation %s%s in its mailbox (%s); it is read at the conversation's next turn and nothing wakes it\n", target.id, agyTitleSuffix, d.Via)
				fmt.Fprintf(out, "  not delivered live: %s\n", d.NotWoken)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", `the title an Antigravity conversation shows above the message (default "Message from <sender>")`)
	return cmd
}

// sender names the calling session and the command that answers it.
func (e *Env) sender() (from, reply string) {
	if id := e.getenv(peer.EnvConversation); id != "" {
		return "antigravity conversation " + id, "reply with: telepathy send " + id + " '<reply>'"
	}
	if id := e.getenv("CLAUDE_CODE_SESSION_ID"); id != "" {
		return "claude session " + id, "reply with: telepathy send " + id + " '<reply>'"
	}
	return "a process outside any agent session", "it has no session to reply to"
}

func (e *Env) getenv(k string) string {
	if e.Getenv == nil {
		return ""
	}
	return e.Getenv(k)
}

// resolvePeer finds the ONE session an id or prefix names, across running Claude Code sessions
// and Antigravity conversations. Two matches are refused with both named: a message sent to the
// wrong session arrives cleanly and says nothing.
func resolvePeer(env *Env, prefix string) (candidate, error) {
	if prefix == "" || strings.ContainsAny(prefix, `/\`) {
		return candidate{}, usagef("%q is not a session id", prefix)
	}
	var found []candidate
	files, err := catalogue.ReadSessionFiles(env.SessionsDir)
	if err != nil {
		return candidate{}, err
	}
	for _, sf := range files {
		if strings.HasPrefix(sf.SessionID, prefix) {
			found = append(found, candidate{harness: "claude", id: sf.SessionID, socket: sf.MessagingSocketPath})
		}
	}
	for _, root := range env.BrainDirs {
		matches, _ := filepath.Glob(filepath.Join(root, prefix+"*"))
		for _, m := range matches {
			if st, err := os.Stat(m); err == nil && st.IsDir() {
				found = append(found, candidate{harness: "antigravity", id: filepath.Base(m), dir: m})
			}
		}
	}
	switch len(found) {
	case 0:
		return candidate{}, fmt.Errorf("no running Claude Code session (advertised in %s) and no Antigravity conversation (in %s) has an id starting %q",
			env.SessionsDir, strings.Join(env.BrainDirs, ", "), prefix)
	case 1:
		return found[0], nil
	}
	names := make([]string, len(found))
	for i, c := range found {
		names[i] = c.harness + " " + c.id
	}
	sort.Strings(names)
	return candidate{}, usagef("%q names %d sessions — give more of the id:\n  %s", prefix, len(found), strings.Join(names, "\n  "))
}

func (e *Env) hostServer() func() peer.ServerIdentity {
	if e.HostServer != nil {
		return e.HostServer
	}
	return peer.HostServer
}

func (e *Env) currentDaemon() (peer.ServerIdentity, bool) {
	if e.CurrentDaemon != nil {
		return e.CurrentDaemon()
	}
	return peer.CurrentDaemon()
}

func (e *Env) conversationBusy(agyRoot, conversation string, excludePID int) (bool, string) {
	if e.ConversationBusy != nil {
		return e.ConversationBusy(agyRoot, conversation, excludePID)
	}
	return peer.ConversationBusy(agyRoot, conversation, excludePID)
}

func sameCanonicalPath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	ca, errA := filepath.EvalSymlinks(a)
	if errA != nil {
		ca = filepath.Clean(a)
	}
	cb, errB := filepath.EvalSymlinks(b)
	if errB != nil {
		cb = filepath.Clean(b)
	}
	return ca == cb
}

func deliver(env *Env, c candidate, body, title string) (peer.Delivery, error) {
	if c.harness == "claude" {
		return peer.SendClaude(c.socket, body)
	}
	m := peer.AgyMessage{Recipient: c.id, Sender: senderID(env), Title: title, Content: body}
	if env.RuntimeDir == "" {
		d, err := peer.EnqueueAgy(c.dir, m, env.Now())
		d.NotWoken = "no runtime directory ($XDG_RUNTIME_DIR is unset), so no server is ever recorded"
		return d, err
	}

	agyRoot := filepath.Dir(filepath.Dir(c.dir))
	client := &http.Client{Timeout: 15 * time.Second}
	own, ownErr := peer.LookupConversationServer(env.RuntimeDir, c.id)
	var ownWhy string
	var ownBusyWhy string
	if ownErr == nil && own != nil {
		if busy, busyWhy := env.conversationBusy(agyRoot, c.id, own.PID); busy {
			ownBusyWhy = busyWhy + " — sending live would create a second writer"
		} else {
			d, err := peer.SendAgyServer(client, *own, m)
			if err == nil {
				return d, nil
			}
			ownWhy = err.Error()
		}
	}

	var why string
	var triedDaemon bool
	var hasDaemon bool
	if ownBusyWhy != "" {
		why = ownBusyWhy
	} else if daemonID, ok := env.currentDaemon(); ok {
		hasDaemon = true
		daemonKey := peer.ServerKey(daemonID.PID, daemonID.StartTime, "")
		if own != nil && own.Key == daemonKey {
			// Route 1 already attempted the current daemon's record.
			triedDaemon = true
		} else if !sameCanonicalPath(agyRoot, env.DaemonRoot) {
			why = fmt.Sprintf("conversation root %s is not the CLI daemon root (%s)", agyRoot, env.DaemonRoot)
		} else {
			daemonSrv, dErr := peer.LookupDaemonServer(env.RuntimeDir, daemonID)
			if dErr != nil {
				why = dErr.Error()
			} else if daemonSrv != nil {
				triedDaemon = true
				if busy, busyWhy := env.conversationBusy(agyRoot, c.id, daemonSrv.PID); busy {
					why = busyWhy + " — sending through the daemon would create a second writer"
				} else {
					d, err := peer.SendAgyServer(client, *daemonSrv, m)
					if err == nil {
						return d, nil
					}
					why = err.Error()
				}
			}
		}
	}

	daemonStatus := "no daemon is running"
	if hasDaemon {
		daemonStatus = "no current daemon server is recorded"
	}
	restoreHint := "opening a daemon-hosted Antigravity conversation restores the live route (a standalone interactive agy does not)"
	if runtime.GOOS != "linux" {
		daemonStatus = "live delivery to an Antigravity conversation runs on Linux only"
		restoreHint = ""
	}
	suffix := daemonStatus
	if restoreHint != "" {
		suffix += " — " + restoreHint
	}

	switch {
	case why != "":
	case ownErr != nil:
		why = ownErr.Error()
	case ownWhy != "" && triedDaemon:
		why = ownWhy
	case ownWhy != "":
		why = ownWhy + "; " + suffix
	default:
		if hasDaemon && runtime.GOOS == "linux" {
			why = "no daemon server is recorded since the last restart — " + restoreHint
		} else {
			why = suffix
		}
	}
	d, err := peer.EnqueueAgy(c.dir, m, env.Now())
	d.NotWoken = why
	return d, err
}

// senderID is the id an Antigravity conversation shows as the sender.
func senderID(env *Env) string {
	for _, k := range []string{peer.EnvConversation, "CLAUDE_CODE_SESSION_ID"} {
		if id := env.getenv(k); id != "" {
			return id
		}
	}
	return "telepathy"
}
