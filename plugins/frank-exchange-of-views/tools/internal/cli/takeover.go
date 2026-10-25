package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// newTakeOver makes the calling session the lead of a live run (runlive.TakeOver).
//
// # Why a verb, and why it must name the lead it replaces
//
// A run's members are the agents of the session that leads it: the hooks attribute a payload to a
// run only when its session_id is the lead the run's live-run row names. A resume can happen in a
// NEW session — the lead crashed, or was restarted — and the resumed workflow's seats then carry the
// new session's id, which no row names, so every sitting goes unbracketed. Re-running setup cannot
// move the lead (it refuses: a re-setup is its lead's own), and nothing may move it implicitly,
// because a session that took a run silently would take another live session's seats with it. So
// the move is its own act, and the caller names the session it replaces — read off the refusal, or
// the marker — which a mistaken caller cannot do by accident.
func newTakeOver() *cobra.Command {
	var from string
	c := &cobra.Command{
		Use:   "take-over <runDir> --from-session <the run's lead session>",
		Short: "make the calling session a live run's lead session, in place of the session named (operator; writes the live-run marker)",
		Long: "take-over <runDir> --from-session <id> makes the session running it lead the run in <runDir>, in place of the " +
			"session named.\n\n" +
			"The hooks attribute an agent to a run only when its hook payload's session_id is the session that leads the run, so " +
			"the seats of a run resumed in a NEW session — its lead session crashed or was restarted — are attributed to nothing until " +
			"that session takes the run over. Run it before the resume, from the Bash tool of the session that resumes, in the " +
			"project the run was set up in.\n\n" +
			"--from-session is the session that leads the run now; the refusal names it when it is not the one given, and so does " +
			"the run's row in .claude/run-live.json. A session takes over only from a lead session it names, so it cannot take a run from " +
			"another session by mistake. A session that already leads the run is told so, and nothing changes.",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			to := seatenv.Session()
			if to == "" {
				return feov.Errorf(feov.MissingField, "%s is unset, so the session to lead the run is unknown — run take-over from the Bash tool of the Claude Code session resuming the run", seatenv.SessionVar)
			}
			if !cmd.Flags().Changed(flags.FromSession) {
				return feov.Errorf(feov.MissingField, "requires --%s: the session that leads the run now, which its row in .claude/run-live.json names", flags.FromSession)
			}
			project, err := os.Getwd()
			if err != nil {
				return err
			}
			prior, err := runlive.TakeOver(project, args[0], from, to)
			var led *runlive.LedElsewhere
			switch {
			case errors.As(err, &led):
				return feov.Errorf(feov.Conflict, "%s is led by session %s, not %s — nothing changed. Take it over only from the run's lead session, named: --%s %s",
					args[0], orNoSession(led.Lead), orNoSession(from), flags.FromSession, orNoSession(led.Lead))
			case err != nil:
				return feov.Errorf(feov.Conflict, "cannot take %s over: %v", args[0], err)
			case prior == to:
				fmt.Fprintf(cmd.OutOrStdout(), "take-over: %s is already led by the calling session (%s) — nothing changed\n", args[0], to)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "take-over: %s is now led by the calling session (%s), in place of %s — its seats from here on are that session's\n", args[0], to, orNoSession(prior))
			}
			return nil
		},
	}
	c.Flags().StringVar(&from, flags.FromSession, "", "the session that leads the run now — its row in .claude/run-live.json names it")
	return c
}

// orNoSession names an empty session id as what it is.
func orNoSession(s string) string {
	if s == "" {
		return `"" (no session)`
	}
	return s
}
