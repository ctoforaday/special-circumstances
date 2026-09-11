package telecli

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// bringBack is how long a restarted Remote Control server brings its own sessions back for: "about
// four hours after the server stopped" (https://code.claude.com/docs/en/remote-control). Neither
// the real daemon nor the limit itself was measured here, which is why every sentence built on it
// says "~4 h".
const bringBack = 4 * time.Hour

// errBootUnmeasured is --lost's refusal: without a boot there is no restart to measure from, and a
// guessed one would list the wrong sessions with the same confidence as the right ones.
var errBootUnmeasured = errors.New("boot time not measurable here — pass --boot")

func newAgentsCmd(env *Env) *cobra.Command {
	var lostOnly bool
	var window time.Duration
	var bootFlag int64
	c := &cobra.Command{
		Use:   "agents",
		Short: "who is running, in which worktree, doing what — and after a restart, what it cut off",
		Long: `Lists every session the client currently advertises, with its liveness, how much it
has done and what it did last — and below them, the sessions the last restart cut off.

Liveness comes from the client's session files and /proc, never from the store: a
transcript cannot say whether the process that wrote it is still there. 'live',
'ended' and 'unknown' are MEASURED, and 'unknown' is a third answer and not a
synonym for 'ended'.

'lost' is INFERRED, from the store and the boot time. A session is lost when its
last sign of life — its newest act, word or thought, in transcript time — precedes
the boot, lies within --window of the newest activity before the boot, was not
closed by the catalogue between that activity and the boot, and is not 'live'. An
advertised session that reads 'ended' and meets that test is listed once, as
'lost'; an 'unknown' one is never 'lost'. The list is candidates for a human: a
session that exited cleanly after the last closure sweep reads the same, and the
resume attempt is the check.

--lost prints only those, one block each: where to resume from, the cloud id when
capture recorded one, the last sign of life, the session's last recorded
permission mode, and the commands that bring it back in that mode. The boot is
/proc/stat's btime on Linux; elsewhere --lost refuses unless --boot is given.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if window <= 0 {
				return usagef("--window must be a positive duration, got %s", window)
			}
			if bootFlag < 0 {
				return usagef("--boot is a Unix time in seconds, got %d", bootFlag)
			}
			db, err := env.openRead(cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer db.Close()
			// The boot is resolved only AFTER the store opened, so a missing store is still the one
			// sentence it has always been, and an Env without a Boot never reaches it.
			boot, source, measured := env.boot(bootFlag)
			if lostOnly && !measured {
				return errBootUnmeasured
			}
			rows, err := catalogue.Agents(cmd.Context(), db, env.SessionsDir)
			if err != nil {
				return err
			}
			var rep catalogue.LostReport
			if measured {
				// ONLY `live` EXCLUDES, and `unknown` is not ours to judge. An `ended` advertisement
				// is exactly what a SIGKILL or a crash leaves behind, so it stays a candidate.
				exclude := map[string]bool{}
				for _, a := range rows {
					if a.Liveness != catalogue.Ended {
						exclude[a.SessionID] = true
					}
				}
				if rep, err = catalogue.FindLost(cmd.Context(), db, boot, window, exclude); err != nil {
					return err
				}
			}
			out := cmd.OutOrStdout()
			if lostOnly {
				printLost(out, env, rep, source)
				return nil
			}
			printAgents(out, env, rows, rep, measured)
			return nil
		},
	}
	c.Flags().BoolVar(&lostOnly, "lost", false,
		"print only the sessions the last restart cut off, each with the commands that bring it back")
	c.Flags().DurationVar(&window, "window", 48*time.Hour,
		"how long before the newest pre-boot activity a session may have gone quiet and still be listed")
	c.Flags().Int64Var(&bootFlag, "boot", 0,
		"the boot to measure from, in Unix seconds (default /proc/stat's btime, on Linux only)")
	return c
}

// boot resolves the boot a restart is measured from: --boot when given, else Env.Boot.
func (e *Env) boot(flag int64) (at int64, source string, ok bool) {
	if flag > 0 {
		return flag, "--boot", true
	}
	if e.Boot == nil {
		return 0, "", false
	}
	at, ok = e.Boot()
	return at, "/proc/stat", ok
}

func printAgents(out io.Writer, env *Env, rows []catalogue.Agent, rep catalogue.LostReport, measured bool) {
	lost := map[string]bool{}
	for _, l := range rep.Lost {
		lost[l.SessionID] = true
	}
	// AN ENDED ADVERTISEMENT THAT IS ALSO LOST IS SHOWN ONCE, as lost, where its commands are.
	var shown []catalogue.Agent
	for _, a := range rows {
		if a.Liveness == catalogue.Ended && lost[a.SessionID] {
			continue
		}
		shown = append(shown, a)
	}
	sort.Slice(shown, func(i, j int) bool {
		if shown[i].LastAt != shown[j].LastAt {
			return shown[i].LastAt > shown[j].LastAt
		}
		return shown[i].SessionID < shown[j].SessionID // a stable order for equal clocks
	})
	switch {
	case len(rows) == 0:
		// NOT an empty table. The client writes a session file per live session, so none means
		// nothing is running — a different fact from a store with no rows, and said in words.
		// With lost rows present it is printed ABOVE them, and stays true of advertised sessions.
		fmt.Fprintf(out, "no sessions are advertised in %s — nothing is running on this box\n", env.SessionsDir)
	case len(shown) == 0:
		// Every advertisement reads `ended` and moved below — what a crash leaves. "Nothing is
		// advertised" would be false: the files are there.
		fmt.Fprintf(out, "%d session file(s) are advertised and none is live — each was cut off by the restart and is listed below as lost\n", len(rows))
	}
	if len(shown) > 0 || len(rep.Lost) > 0 {
		fmt.Fprintf(out, "%-10s %-8s %7s  %-14s %s\n", "SESSION", "STATE", "ACTS", "LAST", "CWD")
	}
	for _, a := range shown {
		fmt.Fprintf(out, "%-10s %-8s %7d  %-14s %s\n",
			catalogue.Short(a.SessionID), a.Liveness, a.Acts, env.last(a.LastTool, a.LastAt), a.CWD)
	}
	for _, l := range rep.Lost {
		fmt.Fprintf(out, "%-10s %-8s %7d  %-14s %s\n",
			catalogue.Short(l.SessionID), "lost", l.Acts, env.last(l.LastTool, l.LastToolAt), resumeFrom(l))
	}
	for _, a := range shown {
		if a.Liveness == catalogue.Unknown {
			fmt.Fprintln(out, "\n  `unknown` means liveness could not be established — a foreign pid namespace, "+
				"or a platform where this is not implemented. It is not `ended`.")
			break
		}
	}
	if len(rep.Lost) > 0 {
		fmt.Fprintf(out, "\n  `lost` is inferred, not measured: last active before the boot at %s and not seen to end. "+
			"`telepathy agents --lost` prints each with the commands that bring it back.\n", utc(rep.Boot))
	}
	if !measured {
		fmt.Fprintf(out, "\n  %v to list the sessions a restart cut off.\n", errBootUnmeasured)
	}
}

// printLost is --lost: the boot and its source, the bring-back arithmetic the restart-recovery
// skill's first step needs — computed HERE, not by the agent reading it — and one block per lost
// session.
func printLost(out io.Writer, env *Env, rep catalogue.LostReport, source string) {
	fmt.Fprintf(out, "boot %s (from %s)\n", utc(rep.Boot), source)
	if rep.NewestSession == "" {
		fmt.Fprintln(out, "newest pre-boot activity: none in the store — run telepathy backfill, then ask again")
	} else {
		// NOW MINUS THE NEWEST PRE-BOOT ACTIVITY IS AN UPPER BOUND on the time since a server
		// stopped: it stopped at or after that activity. Inside the window, a restarted server brings
		// its sessions back when they are opened; past it, it may not.
		d := env.Now().Sub(time.Unix(rep.NewestPreBoot, 0))
		verdict := "past the ~4 h a restarted Remote Control server brings its sessions back in"
		if d <= bringBack {
			verdict = "inside the ~4 h window: open server-spawned sessions in the app first"
		}
		fmt.Fprintf(out, "newest pre-boot activity %s (session %s) — %s ago, %s\n",
			utc(rep.NewestPreBoot), catalogue.Short(rep.NewestSession), hm(d), verdict)
	}
	switch {
	case rep.Sessions == 0:
		fmt.Fprintln(out, "the store holds no sessions — run `telepathy backfill`, or capture has not run here")
		return
	case len(rep.Lost) == 0:
		fmt.Fprintf(out, "nothing was cut off by the restart at %s\n", utc(rep.Boot))
		return
	}
	for _, l := range rep.Lost {
		mode := ""
		if l.Transcript != "" {
			mode = catalogue.PermissionMode(l.Transcript)
		}
		cloud := catalogue.CloudID(l.BridgeSessionID)
		fmt.Fprintf(out, "\nlost %s\n", l.SessionID)
		fmt.Fprintf(out, "  RESUME FROM %s\n", resumeFrom(l))
		if cloud == "" {
			fmt.Fprintln(out, "  CLOUD unknown")
		} else {
			fmt.Fprintf(out, "  CLOUD %s\n", cloud)
		}
		seen := utc(l.LastAt)
		if l.Registered {
			seen += " (registered, no records)"
		}
		fmt.Fprintf(out, "  LAST SIGN OF LIFE %s\n", seen)
		fmt.Fprintf(out, "  %s\n", catalogue.ModeLine(mode))
		reattach, local := catalogue.RecoveryCommands(l.SessionID, cloud, mode)
		if reattach != "" {
			fmt.Fprintf(out, "  REATTACH %s\n", reattach)
		}
		fmt.Fprintf(out, "  LOCAL RESUME %s\n", local)
	}
}

// resumeFrom is where a lost session resumes from, labelled when it is only a hint.
func resumeFrom(l catalogue.LostSession) string {
	switch {
	case l.CWD == "" && l.ProjectDir != "":
		return "unknown (transcript folder " + l.ProjectDir + ")"
	case l.CWD == "":
		return "unknown"
	case !l.Verified:
		return l.CWD + " (unverified)"
	}
	return l.CWD
}

// last renders the LAST column: the last tool and how long ago, or a dash for a session with none.
func (e *Env) last(tool string, at int64) string {
	if tool == "" {
		return "—"
	}
	if at > 0 {
		return tool + " " + e.ago(at)
	}
	return tool
}

// utc renders an instant the way every line here states one: absolute, and in UTC.
func utc(ts int64) string { return time.Unix(ts, 0).UTC().Format("2006-01-02 15:04:05 UTC") }

// hm renders a span as hours and zero-padded minutes, so "7h29m" and "1h02m" line up.
func hm(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}
