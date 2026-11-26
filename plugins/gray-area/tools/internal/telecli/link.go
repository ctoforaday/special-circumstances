package telecli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

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
storing port and token. On Linux, because a restarted remote-control daemon loads any
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
running server's record.

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
			switch got {
			case peer.Recorded:
				fmt.Fprintf(out, "linked antigravity conversation %s to its language server at %s; telepathy send now wakes it\n", conv, s.Address)
			case peer.AlreadyLinked:
				fmt.Fprintf(out, "antigravity conversation %s is already linked to its language server at %s; its link time is refreshed\n", conv, s.Address)
			case peer.NoRuntimeDir:
				return errors.New("$XDG_RUNTIME_DIR is unset, so there is nowhere private to keep the link; telepathy send will queue to this conversation's mailbox instead")
			default:
				return fmt.Errorf("not an Antigravity agent's shell: %s, %s and %s are not all set; a Claude Code session needs no link",
					peer.EnvConversation, peer.EnvAddress, peer.EnvToken)
			}
			return nil
		},
	}
}
