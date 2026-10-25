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

// newTakeOver makes the calling session a live run's lead session (runlive.TakeOver).
//
// # Why a verb, and why it must name the session it replaces
//
// A run's members are the agents of its lead session: the hooks attribute a payload to a run only
// when its session_id is the session the run's live-run row names. A resume can happen in a NEW
// session — the run's lead session crashed, or was restarted — and the resumed workflow's seats then
// carry the new session's id, which no row names, so every sitting goes unbracketed. Re-running setup
// cannot move the run's lead session (it refuses: a re-setup is that session's own), and nothing may
// move it implicitly, because a session that took a run silently would take another live session's
// seats with it. So the move is its own act, and the caller names the session it replaces — read off
// the refusal, or the marker. Naming it is a precondition stated, not a check that the named session
// has ended: the tool cannot see whether a session is still running, and says so.
func newTakeOver() *cobra.Command {
	var from string
	c := &cobra.Command{
		Use:   "take-over <runDir> --from-session <the run's lead session>",
		Short: "make the calling session a live run's lead session, in place of the session named (operator; writes the live-run marker)",
		Long: "take-over <runDir> --from-session <id> makes the session running it lead the run in <runDir>, in place of the " +
			"session named.\n\n" +
			"The hooks attribute an agent to a run only when its hook payload's session_id is the run's lead session, so " +
			"the seats of a run resumed in a NEW session — its lead session crashed or was restarted — are attributed to nothing until " +
			"that session takes the run over. Run it before the resume, from the Bash tool of the session that resumes, in the " +
			"project the run was set up in.\n\n" +
			"PRECONDITION: the run's lead session is no longer running. The tool cannot see whether it is; a take-over from a " +
			"session that is still running moves that session's seats off the run, and its sittings go unbracketed. A new run is " +
			"not a take-over: start it from another checkout, which is its own project.\n\n" +
			"--from-session is the run's lead session now; the refusal names it when it is not the one given, and so does " +
			"the run's row in .claude/run-live.json. A row that names no lead session is taken over with --from-session \"\". " +
			"A session that already leads the run is told so, and nothing changes; a session that leads another live run in the " +
			"project is refused, since the hooks attribute neither of two live runs one session leads.",
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
			var another *runlive.LeadsAnother
			switch {
			case errors.As(err, &led):
				return feov.Errorf(feov.Conflict, "%s's lead session is %s, not %s — nothing changed. Take the run over only to RESUME it, "+
					"once that session is no longer running, naming it: --%s %s. A new run starts from another checkout instead.",
					args[0], runlive.FromSessionArg(led.Lead), runlive.FromSessionArg(from), flags.FromSession, runlive.FromSessionArg(led.Lead))
			case errors.As(err, &another):
				return feov.Errorf(feov.Conflict, "cannot take %s over: the calling session already leads the live run %s — nothing changed. "+
					"Finish or capture that run first; the hooks attribute neither of two live runs one session leads.", args[0], another.Run)
			case err != nil:
				return feov.Errorf(feov.Conflict, "cannot take %s over: %v", args[0], err)
			case prior == to:
				fmt.Fprintf(cmd.OutOrStdout(), "take-over: %s is already led by the calling session (%s) — nothing changed\n", args[0], to)
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "take-over: %s is now led by the calling session (%s), in place of %s — its seats from here on are that session's\n", args[0], to, runlive.FromSessionArg(prior))
			}
			return nil
		},
	}
	c.Flags().StringVar(&from, flags.FromSession, "", "the run's lead session now — its row in .claude/run-live.json names it; \"\" for a row that names none")
	return c
}
