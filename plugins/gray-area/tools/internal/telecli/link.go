package telecli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

func newLinkCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "make this Antigravity conversation wakeable by telepathy send",
		Long: `Records the language server hosting the Antigravity conversation whose shell runs
this, so telepathy send can deliver to it live and wake it when idle. Run it from an
Antigravity agent's own shell: the server's address and token reach nothing else —
not hooks, not MCP servers — so this is the one place they can be recorded.

The record lives in $XDG_RUNTIME_DIR/gray-area/antigravity-servers (directory 0700,
record 0600, gone at logout), keyed by server (PID and start time, or address),
storing port and token. Each link also leaves an empty marker for the conversation in
$XDG_RUNTIME_DIR/gray-area/antigravity-linked, which nothing deletes, so Gray Area's
hook can tell a stale link from none. On Linux, because a restarted remote-control daemon loads any
idle CLI conversation from disk when sent a message for it, linking any one
daemon-hosted conversation makes every idle conversation under the CLI daemon root
wakeable by telepathy send (a standalone interactive agy does not restore the daemon
route). Each record lists the conversations linked to its server, each with when it
last linked; linking again refreshes that time, and send prefers the running server
a conversation linked to last. A link lasts until the server exits: this verb and
every other one delete the records of servers that are gone, and keep a record
whose process cannot be judged. Every other telepathy verb run from an Antigravity
shell links as a side effect, and Gray Area's Antigravity hook asks for a link on
the first step of any turn while neither its server nor the conversation has a
running server's record. A record whose server cannot be judged running (no
Antigravity process above the shell, or a host with no machine id) is never used
by send, and link says so; the hook does not ask on such a host.

A Claude Code session needs no link: its inbox socket is advertised for it.

WHAT IT CANNOT TELL YOU. That the server will still be there when a message is sent:
a restart since the link makes telepathy send fall back to the mailbox until a
daemon-hosted Antigravity conversation links on the new server (Linux only), and it
says so.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			getenv := env.Getenv
			if getenv == nil {
				getenv = func(string) string { return "" }
			}
			s, got, err := peer.RegisterWith(env.RuntimeDir, getenv, env.hostServer(), env.Now())
			env.pruneServers(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			conv := getenv(peer.EnvConversation)
			out := cmd.OutOrStdout()
			unusable := ""
			if got == peer.Recorded || got == peer.AlreadyLinked {
				unusable = unusableRecord(s, env.recordLiveness()(s))
			}
			switch {
			case unusable != "":
				fmt.Fprintf(out, "linked antigravity conversation %s to its language server at %s, but telepathy send cannot use the record: %s\n", conv, s.Address, unusable)
			case got == peer.Recorded:
				fmt.Fprintf(out, "linked antigravity conversation %s to its language server at %s; telepathy send now wakes it\n", conv, s.Address)
			case got == peer.AlreadyLinked:
				fmt.Fprintf(out, "antigravity conversation %s is already linked to its language server at %s; its link time is refreshed\n", conv, s.Address)
			case got == peer.NoRuntimeDir:
				return errors.New("$XDG_RUNTIME_DIR is unset, so there is nowhere private to keep the link; telepathy send will queue to this conversation's mailbox instead")
			default:
				return fmt.Errorf("not an Antigravity agent's shell: %s, %s and %s are not all set; a Claude Code session needs no link",
					peer.EnvConversation, peer.EnvAddress, peer.EnvToken)
			}
			return nil
		},
	}
}

// unusableRecord says why telepathy send will not use the record s, whose liveness reads l, or ""
// when it will: send uses a record only while it reads Live.
func unusableRecord(s peer.Server, l catalogue.Liveness) string {
	switch {
	case l == catalogue.Live:
		return ""
	case s.PID <= 0:
		return "no Antigravity process was found above this shell, so the record names no server process whose running can be judged"
	case s.PidDomain == "":
		return "this host has no pid domain (no /etc/machine-id, or not Linux), so whether the server process runs cannot be judged"
	}
	return fmt.Sprintf("its server process reads %s", l)
}
