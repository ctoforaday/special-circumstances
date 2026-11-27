package telecli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

// dropInName is the drop-in agy-provision writes into the daemon unit's drop-in directory.
const dropInName = "telepathy.conf"

// restartCommand is what the human runs for the daemon to read a provisioned token and drop-in.
const restartCommand = "systemctl --user daemon-reload && systemctl --user restart " + peer.DaemonUnit

func newAgyProvisionCmd(env *Env) *cobra.Command {
	return &cobra.Command{
		Use:   "agy-provision",
		Short: "give Antigravity's daemon a token telepathy send can wake it with (the human runs it)",
		Long: `Lets telepathy send wake Antigravity conversations on this box's remote-control
daemon. It writes files OUTSIDE any repository, rotates the daemon's token, and
takes effect only when the daemon restarts, which ends every turn running on it, so
it runs when the human runs it, or when an agent runs it with the human's agreement.
No hook and no other verb runs it.

On Linux it:

  1. writes a new random token to $XDG_CONFIG_HOME/gray-area/antigravity-daemon.env
     (else ~/.config; directory 0700, file 0600), one line
     ANTIGRAVITY_SERVER_CSRF_TOKEN=<token>. Antigravity's language server adopts
     that variable as its token; no hook, MCP server or agent shell receives it.
     A running daemon keeps its old token, so wakes fail until it restarts;
  2. writes the systemd drop-in
     $XDG_CONFIG_HOME/systemd/user/antigravity-cli-daemon.service.d/telepathy.conf,
     which gives the daemon that file as its environment and runs
     ~/.gemini/config/plugins/gray-area/bin/telepathy agy-wake (agy's own install,
     whose path does not change between releases) after each start;
  3. prints the restart the human runs. It runs no systemctl itself.

It warns when agy's install of telepathy is missing or does not know agy-wake: the
drop-in then runs nothing useful until a release carrying agy-wake is installed
there. The token is never printed.

WHAT IT CANNOT TELL YOU. Whether the daemon has restarted since: until it has, every
wake is refused and the message stays queued, saying so. The first start after the
restart wakes every daemon conversation with mail waiting, however old;
telepathy agents lists them beforehand.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if env.goos() != "linux" {
				return errors.New("the daemon's token is provisioned on Linux only")
			}
			for _, p := range []string{env.TokenFile, env.DropInDir, env.AgyTelepathy} {
				if !unitSafePath(p) {
					return fmt.Errorf("%q cannot be written into a systemd unit: it is empty, relative, or holds whitespace, a quote, a backslash or %%", p)
				}
			}
			if err := peer.WriteDaemonToken(env.TokenFile); err != nil {
				return fmt.Errorf("the token was not written: %w", err)
			}
			dropIn := filepath.Join(env.DropInDir, dropInName)
			body := "# Written by telepathy agy-provision; run it again to rotate the token.\n" +
				"[Service]\n" +
				"EnvironmentFile=-" + env.TokenFile + "\n" +
				"ExecStartPost=-" + env.AgyTelepathy + " agy-wake\n"
			if err := writeFileAtomic(dropIn, []byte(body), 0o644); err != nil {
				return fmt.Errorf("the token is written to %s, but the drop-in was not: %w", env.TokenFile, err)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "wrote a new daemon token to %s (a running daemon refuses wakes until it restarts)\n", env.TokenFile)
			fmt.Fprintf(out, "wrote %s: the daemon reads its token from that file and runs %s agy-wake after each start\n", dropIn, env.AgyTelepathy)
			if why := env.wakeVerbMissing(); why != "" {
				fmt.Fprintf(out, "WARNING: %s; the drop-in's agy-wake does nothing until a release carrying it is installed there\n", why)
			}
			fmt.Fprintf(out, "for live wakes the human runs:\n  %s\n", restartCommand)
			fmt.Fprintln(out, "the restart ends every turn running on the daemon, and the start wakes every daemon conversation with mail waiting (telepathy agents lists them)")
			return nil
		},
	}
}

// wakeVerbMissing says why agy's install of telepathy cannot run agy-wake, or "" when it can.
func (e *Env) wakeVerbMissing() string {
	st, err := os.Stat(e.AgyTelepathy)
	switch {
	case err != nil:
		return fmt.Sprintf("agy's install of telepathy is not at %s", e.AgyTelepathy)
	case st.Mode()&0o111 == 0:
		return fmt.Sprintf("%s is not executable", e.AgyTelepathy)
	}
	run := e.RunProgram
	if run == nil {
		run = runProgram
	}
	if err := run(e.AgyTelepathy, "agy-wake", "--help"); err != nil {
		return fmt.Sprintf("%s does not know agy-wake (%v)", e.AgyTelepathy, err)
	}
	return ""
}

// unitSafePath reports whether p can stand in a systemd unit line unquoted: absolute, and free of
// whitespace, quotes, backslashes and % (systemd's specifier character).
func unitSafePath(p string) bool {
	return filepath.IsAbs(p) && !strings.ContainsAny(p, " \t\r\n\"'\\%")
}

// writeFileAtomic writes b to path with mode, creating its directory, by rename of a temporary file.
func writeFileAtomic(path string, b []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
