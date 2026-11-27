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

  claude-socket              a running Claude Code session's inbox socket. The
                             session's inbound controls then decide: by default the
                             text reaches it between tool calls, or starts a turn if it
                             is idle; a session that bypasses permission prompts HOLDS
                             it for its human's approval; crossSessionInbound "refuse"
                             drops it unseen.
  antigravity-daemon         the current remote-control daemon's language server,
                             through its record. It delivers the text, loading the
                             conversation if it has to, and wakes it if idle.
  antigravity-server-own     a language server whose record lists the conversation: of
                             those still running, the one it linked to last. It
                             delivers the text and wakes the conversation if idle.
  antigravity-server-holder  the language server that has the conversation open,
                             through a record that does not list it. It delivers the
                             text and wakes the conversation if idle.
  antigravity-mailbox        the conversation's on-disk mailbox. The text is QUEUED: it
                             is read at the conversation's next turn, and nothing wakes
                             it. Printed with it: why it was not delivered live, when
                             the daemon started, and when the conversation last
                             recorded a step — the two times that bound the wait.

Live delivery to an Antigravity conversation runs on Linux only; on other platforms
second-writer safety cannot be verified, so send always queues to the mailbox. On
Linux, send first finds who holds the conversation: its presence/<id>.lock, or its
conversations/<id>.db open in any process. When nobody does, it tries the running
server whose record lists the conversation and that it linked to last, then the
current remote-control daemon's record — written by any daemon-hosted conversation
that linked since the daemon started — which adopts conversations under the CLI
daemon root (~/.gemini/antigravity-cli). When a recorded, running server holds it,
that server delivers. Before every attempt send checks again that no process but
that server holds the conversation, because two writers corrupt it; any other
holder, or a holder it cannot inspect, queues the message and is named. Send moves
on from a server only when it could not be reached at all. A server that answers 429
(rate limited) declined without acting: send asks it again after the wait it names
(at most 5 seconds), checking the holder again first, and after its third 429 queues
the message — never through another server. A server that was sent the message and
did not confirm it — no answer, a dropped connection, any other status — may have
it: send then tries no other server, queues nothing, says the delivery is UNKNOWN
and exits non-zero, because a second copy could arrive. A record is
used only while its server process is running; a record whose process cannot be
judged (another pid namespace, a host with no machine id) is never used. The
server's token reaches nothing but an agent's shell commands, so when no
daemon-hosted conversation has linked since a restart (a standalone interactive agy
does not restore the daemon route), send queues to the mailbox and says so.
telepathy link --help says where the record lives; telepathy agents lists every
conversation with messages still waiting.

WHAT IT CANNOT TELL YOU. Delivery is not reading: no route reports that the recipient
acted on the message, an UNKNOWN delivery does not say whether it arrived, and the
Claude Code socket does not report a hold or a drop. The holder check sees only the
processes of send's own pid namespace: a process outside it that has the
conversation open is invisible, so a container that shares Antigravity's data with
the host can create a second writer. A Claude Code session that is
not running has no inbox and is refused; telepathy agents shows which are. On native
Windows a Claude Code inbox takes posts only from its own session, so sending to one
is refused there.`,
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
			d, bounds, err := deliver(env, target, body, title)
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
			case peer.RouteAgyDaemon, peer.RouteAgyOwnRecord, peer.RouteAgyHolderRecord:
				fmt.Fprintf(out, "delivered to antigravity conversation %s%s via %s, the language server at %s; it wakes if idle\n", target.id, agyTitleSuffix, d.Route, d.Via)
			case peer.RouteAgyMailbox:
				fmt.Fprintf(out, "QUEUED for antigravity conversation %s%s in its mailbox (%s); it is read at the conversation's next turn and nothing wakes it\n", target.id, agyTitleSuffix, d.Via)
				fmt.Fprintf(out, "  not delivered live: %s\n", d.NotWoken)
				for _, line := range bounds {
					fmt.Fprintf(out, "  %s\n", line)
				}
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

func (e *Env) conversationBusy(agyRoot, conversation string, excludePID int) peer.Holder {
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

// restoreHint is what brings the daemon route back, said wherever its absence is the cause.
const restoreHint = "opening a daemon-hosted Antigravity conversation restores the live route (a standalone interactive agy does not)"

// deliver sends body to c by the route c's harness and state allow. For a message it QUEUED to an
// Antigravity mailbox it also returns the lines that bound the wait: when the daemon started, and
// when the conversation last recorded a step.
func deliver(env *Env, c candidate, body, title string) (peer.Delivery, []string, error) {
	if c.harness == "claude" {
		d, err := peer.SendClaude(c.socket, body)
		return d, nil, err
	}
	m := peer.AgyMessage{Recipient: c.id, Sender: senderID(env), Title: title, Content: body}
	daemonID, hasDaemon := env.currentDaemon()
	if env.goos() != "linux" {
		hasDaemon = false
	}
	r := &agyRouter{env: env, c: c, m: m, daemonID: daemonID, hasDaemon: hasDaemon,
		agyRoot: filepath.Dir(filepath.Dir(c.dir)), live: env.recordLiveness(),
		client: peer.NewAgyClient(env.sendTimeout())}
	if hasDaemon {
		r.daemonKey = peer.ServerKey(daemonID.PID, daemonID.StartTime, "")
	}
	d, ok := r.route()
	if ok {
		return d, nil, nil
	}
	if r.unknown != nil {
		// A second server, or the mailbox, could deliver the message twice: once through the server
		// that may already have it, and once more.
		return peer.Delivery{}, nil, fmt.Errorf("delivery UNKNOWN: %v — the message may have arrived, so no other server was tried and nothing was queued; check the conversation before sending it again", r.unknown)
	}
	q, err := peer.EnqueueAgy(c.dir, m, env.Now())
	q.NotWoken = r.cause
	return q, []string{r.daemonLine(), lastStepLine(c.dir)}, err
}

// agyRouter chooses the one server, if any, that may deliver to an Antigravity conversation, and
// otherwise holds the sentence saying why none may (peer's package doc states the model).
type agyRouter struct {
	env       *Env
	c         candidate
	m         peer.AgyMessage
	agyRoot   string
	live      func(peer.Server) catalogue.Liveness
	client    *http.Client
	servers   []peer.Server
	daemonID  peer.ServerIdentity
	hasDaemon bool
	daemonKey string
	// cause is why the message was not delivered live: the reason the last target tried stopped.
	cause string
	// unknown is the failure of a server that was sent the message and did not confirm it, which
	// ends the send with its arrival unknown.
	unknown error
}

func (r *agyRouter) queue(format string, a ...any) (peer.Delivery, bool) {
	r.cause = fmt.Sprintf(format, a...)
	return peer.Delivery{}, false
}

func (r *agyRouter) route() (peer.Delivery, bool) {
	if r.env.RuntimeDir == "" {
		return r.queue("no runtime directory ($XDG_RUNTIME_DIR is unset), so no server is ever recorded")
	}
	if r.env.goos() != "linux" {
		return r.queue("live delivery to an Antigravity conversation runs on Linux only")
	}
	servers, err := peer.ListServers(r.env.RuntimeDir)
	if err != nil {
		return r.queue("the server records cannot be read: %v", err)
	}
	r.servers = servers

	// ONE DISCOVERY, EXCLUDING NOBODY: who holds the conversation decides which server may reach it.
	h := r.env.conversationBusy(r.agyRoot, r.c.id, 0)
	if h.Kind == peer.HolderNone {
		return r.toUnheld()
	}
	if x := r.liveRecordOf(h.PID); x != nil {
		return r.throughHolder(*x, h.PID)
	}
	return r.queue("%s", r.heldBy(h))
}

// toUnheld tries, while nobody holds the conversation, the Live record it linked to last and then
// the current daemon's. A target that now finds the conversation held stops the send: the holder
// is the only server that may reach it, and it is reported rather than bypassed. So does a target
// that was sent the message and did not confirm it; only one that could not be reached at all
// passes the message on to the next.
func (r *agyRouter) toUnheld() (peer.Delivery, bool) {
	own := peer.NewestLiveListing(r.servers, r.c.id, r.live)
	if own != nil {
		d, ok, stop := r.attempt(*own)
		if ok || stop {
			return d, ok
		}
		if own.Key == r.daemonKey {
			return d, false // the daemon's record was the one tried
		}
	}
	daemon, why := r.daemonTarget()
	if daemon == nil {
		r.cause = why
		return peer.Delivery{}, false
	}
	d, ok, _ := r.attempt(*daemon)
	return d, ok
}

// throughHolder delivers through X, the Live record of the server holding the conversation, once a
// probe excluding X's pid finds nobody else. It never falls back to another server: X has the
// conversation open.
func (r *agyRouter) throughHolder(x peer.Server, pid int) (peer.Delivery, bool) {
	g := r.env.conversationBusy(r.agyRoot, r.c.id, pid)
	switch {
	case g.Kind == peer.HolderNone:
		d, ok, _ := r.send(x)
		return d, ok
	case g.Kind == peer.HolderUninspectable:
		return r.queue("%s", uninspectable(g))
	}
	return r.queue("held by the language server recorded at %s and also by pid %d", x.Address, g.PID)
}

// attempt sends through target once a probe excluding only target's pid finds nobody holding the
// conversation. stop reports that the send ends here: the probe found a holder, or target may have
// the message.
func (r *agyRouter) attempt(target peer.Server) (d peer.Delivery, ok, stop bool) {
	g := r.env.conversationBusy(r.agyRoot, r.c.id, target.PID)
	if g.Kind != peer.HolderNone {
		r.queue("%s", r.heldBy(g))
		return peer.Delivery{}, false, true
	}
	return r.send(target)
}

// rateLimitAttempts is how many times send asks one server that answers 429 (rate limited), and
// rateLimitWaitCap bounds the wait its Retry-After may ask for between them.
const (
	rateLimitAttempts = 3
	rateLimitWaitCap  = 5 * time.Second
	rateLimitWait     = time.Second // when the server names no wait
)

// send delivers through target, whose probe found nobody else holding the conversation. A target
// that could not be reached leaves the message undelivered, with that as the cause, and the send
// may go on. A target that answers 429 declined without acting: send asks it again, up to
// rateLimitAttempts times, probing before each request as before the first, and then queues — never
// another server, which the conversation could reach twice. Any other failure leaves its arrival
// unknown and stops the send.
func (r *agyRouter) send(target peer.Server) (d peer.Delivery, ok, stop bool) {
	for attempt := 1; ; attempt++ {
		d, err := peer.SendAgyServer(r.client, target, r.m, r.routeFor(target))
		var unreached *peer.UnreachedError
		var limited *peer.RateLimitedError
		switch {
		case err == nil:
			return d, true, false
		case errors.As(err, &unreached):
			r.queue("%v", err)
			return peer.Delivery{}, false, false
		case errors.As(err, &limited) && attempt == rateLimitAttempts:
			r.queue("the language server recorded at %s answered 429 (rate limited) %d times", target.Address, attempt)
			return peer.Delivery{}, false, true
		case errors.As(err, &limited):
			wait := limited.RetryAfter
			if wait <= 0 {
				wait = rateLimitWait
			}
			r.env.sleep(min(wait, rateLimitWaitCap))
			if g := r.env.conversationBusy(r.agyRoot, r.c.id, target.PID); g.Kind != peer.HolderNone {
				r.queue("%s", r.heldBy(g))
				return peer.Delivery{}, false, true
			}
			continue
		}
		r.unknown = err
		return peer.Delivery{}, false, true
	}
}

// daemonTarget is the current daemon's record when it may adopt the conversation, or nil and why not.
func (r *agyRouter) daemonTarget() (*peer.Server, string) {
	if !r.hasDaemon {
		return nil, "no daemon is running — " + restoreHint
	}
	if !sameCanonicalPath(r.agyRoot, r.env.DaemonRoot) {
		return nil, fmt.Sprintf("conversation root %s is not the CLI daemon root (%s), so the daemon cannot adopt it", r.agyRoot, r.env.DaemonRoot)
	}
	rec, why := r.daemonRecord()
	return rec, why
}

// daemonRecord is the current daemon's Live record, or nil and the sentence saying why there is none.
func (r *agyRouter) daemonRecord() (*peer.Server, string) {
	rec, err := peer.LookupDaemonServer(r.env.RuntimeDir, r.daemonID)
	switch {
	case err != nil:
		return nil, fmt.Sprintf("the daemon's server record cannot be read: %v", err)
	case rec == nil:
		return nil, fmt.Sprintf("the daemon (%s) has not been linked since it started — %s", r.daemonStarted(), restoreHint)
	}
	if l := r.live(*rec); l != catalogue.Live {
		return nil, fmt.Sprintf("the daemon's server record reads %s, and only a record whose server is running is used", l)
	}
	return rec, ""
}

// heldBy is the sentence for a conversation g holds that no send may reach.
func (r *agyRouter) heldBy(g peer.Holder) string {
	switch {
	case g.Kind == peer.HolderUninspectable:
		return uninspectable(g)
	case r.liveRecordOf(g.PID) != nil:
		return fmt.Sprintf("held by the language server recorded at %s (pid %d)", r.liveRecordOf(g.PID).Address, g.PID)
	case r.hasDaemon && g.PID == r.daemonID.PID:
		_, why := r.daemonRecord()
		if why == "" {
			why = fmt.Sprintf("held by the daemon (pid %d)", g.PID)
		}
		return why
	}
	return fmt.Sprintf("held by another process, pid %d; sending live would create a second writer", g.PID)
}

func uninspectable(g peer.Holder) string {
	return "cannot inspect whether another process holds the conversation: " + g.Detail
}

// liveRecordOf is the Live record of the server running as pid, or nil. Only a Live record's pid is
// ever excluded from a probe, so a pid from another namespace never stands in for a local process.
func (r *agyRouter) liveRecordOf(pid int) *peer.Server {
	if pid <= 0 {
		return nil
	}
	for i := range r.servers {
		if r.servers[i].PID == pid && r.live(r.servers[i]) == catalogue.Live {
			return &r.servers[i]
		}
	}
	return nil
}

// routeFor names the kind of record s is for this conversation.
func (r *agyRouter) routeFor(s peer.Server) peer.Route {
	switch {
	case r.daemonKey != "" && s.Key == r.daemonKey:
		return peer.RouteAgyDaemon
	case s.Lists(r.c.id):
		return peer.RouteAgyOwnRecord
	}
	return peer.RouteAgyHolderRecord
}

// daemonStarted is "started <T>", or "start not measured" when the boot or the start is unknown.
func (r *agyRouter) daemonStarted() string {
	if r.daemonID.StartTime <= 0 || r.env.Boot == nil {
		return "start not measured"
	}
	boot, ok := r.env.Boot()
	if !ok {
		return "start not measured"
	}
	return "started " + utc(peer.StartTimeAt(boot, r.daemonID.StartTime).Unix())
}

// daemonLine bounds a queued message's wait from below: nothing woke the conversation since then.
// Off Linux the daemon is never looked for, so its absence is not a measurement.
func (r *agyRouter) daemonLine() string {
	if r.env.goos() != "linux" {
		return "daemon start not measured"
	}
	if !r.hasDaemon {
		return "no daemon is running"
	}
	if s := r.daemonStarted(); s != "start not measured" {
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
