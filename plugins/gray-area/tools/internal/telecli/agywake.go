package telecli

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

// portWait bounds agy-wake's wait for the daemon and its HTTP port, and portPoll is the interval
// between looks.
const (
	portWait = 45 * time.Second
	portPoll = 500 * time.Millisecond
)

func newAgyWakeCmd(env *Env) *cobra.Command {
	var tokenFile string
	cmd := &cobra.Command{
		Use:   "agy-wake",
		Short: "wake every Antigravity daemon conversation with mail waiting (systemd runs it at each daemon start)",
		Long: `Starts a turn, through Antigravity's remote-control daemon, in every conversation
under the CLI daemon root (~/.gemini/antigravity-cli) whose mailbox holds a message
from an agent or a shell that its client has not read; the turn reads the mailbox
first. The drop-in telepathy agy-provision writes runs it after each daemon start,
so a message queued while the daemon was down, or while nothing woke its
conversation, is read once the daemon is back. A hand run — by the human, or by an
agent with the human's agreement — spends a turn in each conversation it wakes.

It reads the daemon's token, waits up to 45 seconds for the daemon and its HTTP
port, then sends each such conversation the wake message telepathy send uses,
unless another process holds it (that holder's next turn reads the mailbox). It
does not wake for agy's own notices, for earlier wake messages, or for a message
whose file is missing or does not parse: none of those is mail an agent is owed.
Every request is cut to the time left of a 60-second bound, at most 5 seconds each,
and the probes of the daemon's ports run at once, because systemd stops the daemon
when a start command outlives the start timeout; the local reads (systemctl, /proc,
the mailboxes) are not cut, so the run ends at the bound plus at most the last of
them. A conversation not reached in time is named. --token-file names the token
file (the drop-in passes the one agy-provision wrote); without it, the default
path telepathy send reads. It prints one line per conversation to stderr (the
daemon's journal) and exits 0 whatever it found, so it never fails the daemon's
start.

WHAT IT CANNOT TELL YOU. That a woken turn ran: a wake answered 200 says only that
the daemon took it. It keeps no record between runs, so every daemon start wakes
every conversation still holding unread mail, however old.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if tokenFile != "" {
				env.TokenFile = tokenFile
			}
			runAgyWake(env, cmd.ErrOrStderr())
			return nil
		},
	}
	cmd.Flags().StringVar(&tokenFile, "token-file", "", "the daemon's token file (default $XDG_CONFIG_HOME/gray-area/antigravity-daemon.env)")
	return cmd
}

// runAgyWake is agy-wake's whole run, every line to log. Nothing it meets fails it.
func runAgyWake(env *Env, log io.Writer) {
	start := env.Now()
	deadline := start.Add(env.wakeBound())
	say := func(format string, a ...any) { fmt.Fprintf(log, "telepathy agy-wake: "+format+"\n", a...) }
	left := func() time.Duration { return deadline.Sub(env.Now()) }
	// A request's timeout is the time left, at most SendTimeout, and never zero, which an
	// http.Client reads as no timeout at all.
	client := func() *http.Client { return peer.NewAgyClient(max(min(env.sendTimeout(), left()), time.Millisecond)) }
	if env.goos() != "linux" {
		say("not run: a wake is sent on Linux only")
		return
	}
	if env.DaemonRoot == "" {
		say("not run: no home directory, so no CLI daemon root")
		return
	}
	tok, why := env.daemonToken()
	if why != "" {
		say("not run: %s", why)
		return
	}
	w := wakeDaemon{env: env, token: tok, looked: true}
	var addr string
	var lastErr string
	waitUntil := start.Add(min(portWait, env.wakeBound()))
	for {
		w.client = client()
		if w.id, w.found = env.currentDaemon(); !w.found {
			lastErr = "no daemon is running"
		} else if a, err := env.daemonAddress(w.client, w.id.PID); err != nil {
			lastErr = "the daemon's HTTP port cannot be told: " + err.Error()
		} else {
			addr = a
			break
		}
		if !env.Now().Add(portPoll).Before(waitUntil) {
			say("not run: %s (waited %s)", lastErr, env.Now().Sub(start).Round(time.Second))
			return
		}
		env.sleep(portPoll)
	}
	root := filepath.Clean(env.DaemonRoot)
	waiting, errs := peer.ScanMailboxes([]string{filepath.Join(root, "brain")}, "")
	for _, err := range errs {
		say("mailboxes not measured: %v", err)
	}
	for _, c := range waiting {
		if c.Messages-c.Unreadable <= 0 {
			continue
		}
		if left() <= 0 {
			say("not woken %s: the %s bound ran out", c.Conversation, env.wakeBound())
			continue
		}
		w.client = client()
		if why := w.wakeAt(addr, root, c.Conversation); why != "" {
			say("not woken %s: %s", c.Conversation, why)
			continue
		}
		say("woke %s (waiting: %d)", c.Conversation, c.Messages-c.Unreadable)
	}
}
