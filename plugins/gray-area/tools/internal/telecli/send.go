package telecli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
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
  antigravity-mailbox  the conversation's on-disk mailbox, ALWAYS, for every
                       Antigravity message. The text is QUEUED: the conversation reads
                       it at the start of its next turn, whatever starts that turn.

A QUEUED message is then WOKEN when it can be: send asks Antigravity's remote-control
daemon to start the conversation's turn now, with a fixed wake message ("You have
queued messages.") that asks for nothing; the turn it starts reads the mailbox
first. The text itself never travels live, so it cannot arrive twice. Printed under
QUEUED: "woken" and the daemon's address, or "not woken" and why, with when the
daemon started and when the conversation last recorded a step.

A wake needs, in this order: Linux; a conversation under the CLI daemon root
(~/.gemini/antigravity-cli); the daemon's token in
$XDG_CONFIG_HOME/gray-area/antigravity-daemon.env, which the human provisions with
telepathy agy-provision (it is read before any network contact); a running daemon;
its HTTP port, read from /proc; and no process but the daemon holding the
conversation — its presence/<id>.lock, or its conversations/<id>.db open. A
conversation another process holds gets no wake, because a language server sent a
message for a conversation another holds adopts it and the two then write it at
once; that holder's next turn reads the mailbox. A wake that fails leaves the message
queued and says why. A conversation hosted anywhere but the daemon (a standalone
agy, a TUI, the IDE) is read at its own next turn. At each daemon start the
provisioned daemon runs telepathy agy-wake, which wakes every conversation with mail
waiting; telepathy agents lists every conversation with messages still waiting.

WHAT IT CANNOT TELL YOU. Delivery is not reading: no route reports that the recipient
acted on the message, a wake that answered 200 does not say the turn ran, and the
Claude Code socket does not report a hold or a drop. The holder check sees only the
processes of send's own pid namespace and the locks of this host, and not a server
that has resumed the conversation without yet opening its database. A process outside
the namespace that has the conversation open is invisible, and so is a lock taken on
another host through a FUSE client of a distributed filesystem: from a container
that shares Antigravity's data with the host, or over such a filesystem, a wake can
create a second writer. A Claude Code session that is not running has no inbox and
is refused; telepathy agents shows which are. On native Windows a Claude Code inbox
takes posts only from its own session, so sending to one is refused there.`,
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
			if d.Route == peer.RouteClaudeSocket {
				fmt.Fprintf(out, "delivered to claude session %s over its inbox socket %s\n", target.id, d.Via)
				return nil
			}
			agyTitleSuffix := ""
			if t := env.agyTitle(target.id); t != "" {
				agyTitleSuffix = fmt.Sprintf(" (%q)", t)
			}
			fmt.Fprintf(out, "QUEUED for antigravity conversation %s%s in its mailbox (%s); it is read at the start of the conversation's next turn\n", target.id, agyTitleSuffix, d.Via)
			if d.wokenVia != "" {
				fmt.Fprintf(out, "  woken: the daemon's language server at %s was sent a wake, which starts that turn now\n", d.wokenVia)
				return nil
			}
			fmt.Fprintf(out, "  not woken: %s\n", d.notWoken)
			for _, line := range d.bounds {
				fmt.Fprintf(out, "  %s\n", line)
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

func (e *Env) currentDaemon() (peer.ServerIdentity, bool) {
	if e.CurrentDaemon != nil {
		return e.CurrentDaemon()
	}
	return peer.CurrentDaemon()
}

func (e *Env) conversationBusy(agyRoot, conversation string, excludePID int) peer.Holder {
	if e.ConversationBusy != nil {
		return e.ConversationBusy(agyRoot, conversation, excludePID)
	}
	return peer.ConversationBusy(agyRoot, conversation, excludePID)
}

func (e *Env) daemonAddress(client *http.Client, pid int) (string, error) {
	if e.DaemonAddress != nil {
		return e.DaemonAddress(client, pid)
	}
	return peer.DaemonAddress(client, pid)
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

// sent is what a send did: the delivery, and for an Antigravity message whether a wake went with it.
type sent struct {
	peer.Delivery
	// wokenVia is the daemon address a wake went to; empty when none went.
	wokenVia string
	// notWoken says why no wake went; bounds are the lines bounding the wait then.
	notWoken string
	bounds   []string
}

// deliver sends body to c by the one route c's harness has. An Antigravity message always goes to
// the conversation's mailbox; a wake follows when the daemon may start the conversation's turn.
func deliver(env *Env, c candidate, body, title string) (sent, error) {
	if c.harness == "claude" {
		d, err := peer.SendClaude(c.socket, body)
		return sent{Delivery: d}, err
	}
	m := peer.AgyMessage{Recipient: c.id, Sender: senderID(env), Title: title, Content: body}
	d, err := peer.EnqueueAgy(c.dir, m, env.Now())
	if err != nil {
		return sent{}, err
	}
	out := sent{Delivery: d}
	agyRoot := filepath.Dir(filepath.Dir(c.dir))
	dm, why := env.daemonForWake(agyRoot)
	if why == "" {
		var addr string
		if addr, why = dm.wake(agyRoot, c.id); why == "" {
			out.wokenVia = addr
			return out, nil
		}
	}
	out.notWoken = why
	out.bounds = []string{dm.daemonLine(), lastStepLine(c.dir)}
	return out, nil
}

// wakeDaemon is the daemon a wake goes through: its process, its token and its client.
type wakeDaemon struct {
	env *Env
	id  peer.ServerIdentity
	// looked says the daemon was looked for; found that it runs.
	looked, found bool
	token         string
	client        *http.Client
}

// daemonForWake checks, in order, everything a wake needs before any holder check, and says why
// none can go. The token file is read before any network contact, so an environment without one
// never probes a daemon.
func (e *Env) daemonForWake(agyRoot string) (wakeDaemon, string) {
	w := wakeDaemon{env: e, client: peer.NewAgyClient(e.sendTimeout())}
	if e.goos() != "linux" {
		return w, "a wake is sent on Linux only"
	}
	if !sameCanonicalPath(agyRoot, e.DaemonRoot) {
		return w, fmt.Sprintf("conversation root %s is not the CLI daemon root (%s), so the daemon cannot wake it", agyRoot, e.DaemonRoot)
	}
	tok, why := e.daemonToken()
	if why != "" {
		return w, why
	}
	w.token = tok
	w.id, w.found = e.currentDaemon()
	w.looked = true
	if !w.found {
		return w, "no daemon is running"
	}
	return w, ""
}

// daemonToken reads the daemon's token, or says why it cannot be used. Neither carries the token.
func (e *Env) daemonToken() (string, string) {
	if e.TokenFile == "" {
		return "", "no configuration directory, so no daemon token file"
	}
	tok, err := peer.ReadDaemonToken(e.TokenFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", fmt.Sprintf("no daemon token file at %s: the human provisions one with telepathy agy-provision and restarts the daemon, which ends every turn running on it", e.TokenFile)
	case err != nil:
		return "", fmt.Sprintf("the daemon token file is not used: %v", err)
	}
	return tok, ""
}

// wake sends the wake message for conversation through the daemon once nothing but the daemon is
// found holding it, and returns the address it went to, or why it did not go.
func (w wakeDaemon) wake(agyRoot, conversation string) (string, string) {
	addr, err := w.env.daemonAddress(w.client, w.id.PID)
	if err != nil {
		return "", fmt.Sprintf("the daemon's HTTP port cannot be told: %v", err)
	}
	return addr, w.wakeAt(addr, agyRoot, conversation)
}

// wakeAt sends the wake message for conversation to the daemon at addr, or says why it did not.
// The holder check runs last, just before the request, to keep the window in which a process can
// open the conversation unseen as short as it can be.
func (w wakeDaemon) wakeAt(addr, agyRoot, conversation string) string {
	switch h := w.env.conversationBusy(agyRoot, conversation, w.id.PID); h.Kind {
	case peer.HolderNone:
	case peer.HolderUninspectable:
		return "cannot inspect whether another process holds the conversation: " + h.Detail
	default:
		return fmt.Sprintf("held by pid %d, not the daemon, whose next turn reads the mailbox; a wake through the daemon would make a second writer", h.PID)
	}
	if err := peer.SendWake(w.client, addr, w.token, conversation); err != nil {
		return fmt.Sprintf("the wake failed: %v", err)
	}
	return ""
}

// daemonStarted is "started <T>", or "start not measured" when the boot or the start is unknown.
func (e *Env) daemonStarted(id peer.ServerIdentity) string {
	if id.StartTime <= 0 || e.Boot == nil {
		return "start not measured"
	}
	boot, ok := e.Boot()
	if !ok {
		return "start not measured"
	}
	return "started " + utc(peer.StartTimeAt(boot, id.StartTime).Unix())
}

// daemonLine bounds a queued message's wait: when the daemon last started. When the daemon was
// never looked for (an earlier check ended the wake), its absence is not a measurement.
func (w wakeDaemon) daemonLine() string {
	switch {
	case !w.looked:
		return "daemon start not measured"
	case !w.found:
		return "no daemon is running"
	}
	if s := w.env.daemonStarted(w.id); s != "start not measured" {
		return "daemon " + s
	}
	return "daemon start not measured"
}

// lastStepLine says when the conversation in dir last recorded a step: the created_at of the last
// line of its transcript, read from the end of the file.
func lastStepLine(dir string) string {
	at, ok, err := lastStepAt(filepath.Join(dir, ".system_generated", "logs", "transcript.jsonl"))
	switch {
	case err != nil:
		return "last step not measured: " + err.Error()
	case !ok:
		return "no step recorded"
	}
	return "last step recorded at " + utc(at.Unix())
}

func lastStepAt(path string) (time.Time, bool, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return time.Time{}, false, err
	}
	// A step line can be megabytes long, so the tail grows a chunk at a time until it holds a
	// newline before the last non-blank line, or the whole file.
	const chunk = 64 << 10
	var tail, line []byte
	for pos := st.Size(); pos > 0; {
		n := int64(chunk)
		if n > pos {
			n = pos
		}
		pos -= n
		b := make([]byte, n)
		if _, err := f.ReadAt(b, pos); err != nil && !errors.Is(err, io.EOF) {
			return time.Time{}, false, err
		}
		tail = append(b, tail...)
		trimmed := bytes.TrimRight(tail, " \t\r\n")
		if len(trimmed) == 0 {
			continue
		}
		if i := bytes.LastIndexByte(trimmed, '\n'); i >= 0 {
			line = trimmed[i+1:]
			break
		}
		if pos == 0 {
			line = trimmed
		}
	}
	if len(line) == 0 {
		return time.Time{}, false, nil
	}
	var step struct {
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(line, &step); err != nil {
		return time.Time{}, false, fmt.Errorf("the transcript's last line is not a step: %w", err)
	}
	at, err := time.Parse(time.RFC3339Nano, step.CreatedAt)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("the transcript's last step has no time: %w", err)
	}
	return at, true, nil
}

// senderID is the id an Antigravity conversation shows as the sender.
func senderID(env *Env) string {
	for _, k := range []string{peer.EnvConversation, "CLAUDE_CODE_SESSION_ID"} {
		if id := env.getenv(k); id != "" {
			return id
		}
	}
	return peer.ShellSender
}
