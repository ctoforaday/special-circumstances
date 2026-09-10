package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// manualName is the command's name. The survey that reads a manual out of a trajectory owns it,
// with the page grammar, so the writer and the reader cannot come to disagree about either.
const manualName = diagnostics.ManualCommand

// newManual prints every command on THIS surface with that command's own help, in one call.
//
// WHY IT EXISTS. Seats learn the tool by reading its help, and walking it a page at a time was
// measured in a real run (#861 B3): 142 help turns across 26 sittings, 14% of all turn time, a cold
// sitting's median seven help turns before its first act. One call replaces the walk.
//
// WHY IT RUNS THE HELP RATHER THAN SHOWING A COPY OF IT. An earlier proposal (#593) staged a
// snapshot of the help tree and was rejected because a snapshot drifts from the tool it describes.
// This executes each command's real `--help` at call time — the same binary, the same seat, the
// same run — so what it prints IS what the command would say, and there is nothing to keep in step.
//
// The surface is the one this tree was built for — the flag, else the seat this agent registered
// as, exactly as every invocation resolves it — so seatID is captured at construction rather than
// re-derived. It records nothing.
func newManual(seatID string) *cobra.Command {
	return &cobra.Command{
		Use:   manualName,
		Short: "every command on your surface, each with its own help, run live — read it once, first thing in a sitting",
		Long: manualName + " prints every command on this surface — the groups, and every command they hold, in tree order — each under a header naming the command exactly, followed by the help that command prints. The help is produced now, by running this binary with that command and --help, so it cannot disagree with the tool you are running. It records nothing.\n\n" +
			"Read it ONCE, at the start of a sitting and before you decide what to do: it is your whole surface in one call. A command's own --help is still there when you want to re-check one page.\n\n" +
			"The surface is your seat's — the seat you registered as, or the one this call names. A page whose help fails to run is printed with its error, never left out, and the command then exits non-zero.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printManual(cmd, seatID)
		},
	}
}

// printManual writes the manual: one line saying whose surface it is and how many commands it
// holds, then the surface's own page, then every command's page in tree order.
func printManual(cmd *cobra.Command, seatID string) error {
	// NO JSON FORM, SAID RATHER THAN IGNORED. What this prints is help pages — text for a reader —
	// and a --json that was quietly dropped would hand a machine consumer prose on a channel whose
	// contract is that it parses.
	if j, _ := cmd.Flags().GetBool(flags.JSON); j {
		return feov.Errorf(feov.Validation, "%s: there is no JSON form — it prints help pages, which are text to read. Run it without the JSON flag", manualName)
	}
	root := cmd.Root()
	// The surface's own page first: it carries what no command page does — the flags every command
	// inherits, what each means, and the footer on what to do about a capability you cannot find.
	pages := [][]string{nil}
	walkSurface(root, func(path []string, _ *cobra.Command) { pages = append(pages, path) })

	runDir, _ := cmd.Flags().GetString(flags.Run)
	w := cmd.OutOrStdout()
	bin := InvokedAs()
	fmt.Fprintf(w, "%s %s — the %s surface, seat %s: %d commands, each under a header naming it and followed by the help it prints, run live just now. The surface's own page comes first.\n",
		bin, manualName, RoleOfSeat(seatID), seatID, len(pages)-1)

	var failed []string
	for _, path := range pages {
		// THE SEAT IS PASSED EXPLICITLY, and the run with it where this call named one. A child that
		// had to rediscover who is asking could resolve a different surface from the one this call
		// was built for; an injected run arrives through the inherited environment either way.
		argv := append(append([]string{}, path...), "--help", "--"+flags.SeatID, seatID)
		if runDir != "" {
			argv = append(argv, "--"+flags.Run, runDir)
		}
		out, err := runHelp(argv)
		fmt.Fprintln(w, diagnostics.ManualRule)
		fmt.Fprintln(w, diagnostics.ManualHeader(bin, path))
		if err != nil {
			failed = append(failed, strings.TrimSpace(strings.Join(path, " ")))
			fmt.Fprintf(w, "THIS PAGE DID NOT RUN, so what follows is its error rather than its help: %v\n", err)
		}
		fmt.Fprint(w, out)
		if out != "" && !strings.HasSuffix(out, "\n") {
			fmt.Fprintln(w)
		}
	}
	if len(failed) > 0 {
		for i, f := range failed {
			if f == "" {
				failed[i] = "(the surface's own page)"
			}
		}
		return fmt.Errorf("%s: %d of %d help pages did not run, and each is printed above with its error: %s",
			manualName, len(failed), len(pages), strings.Join(failed, ", "))
	}
	return nil
}

// runHelp runs this tool with argv and returns everything it printed. It is a variable so a test
// can stand a failing one in; production never reassigns it.
//
// Assigned in init rather than at the declaration: the executor builds the tree, the tree mounts
// manual, and manual calls the executor, which Go refuses as an initialization cycle.
var runHelp func(argv []string) (string, error)

func init() { runHelp = selfExecHelp }

// selfExecHelp runs THIS BINARY again with argv — the real help, from the real tool, at call time.
//
// NOT INSIDE A TEST BINARY, and the reason is not speed. os.Executable() there is the test binary,
// and running it with a command path does not print that command's help — it runs the whole suite
// again, which calls manual again. The fuzz harness drives commands in-process through the same
// tree, so a test binary reaches this path without any test asking for it. In one, the pages come
// from the same tree built in-process, which is what the child would have built.
func selfExecHelp(argv []string) (string, error) {
	if testing.Testing() {
		return inProcessHelp(argv)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate this binary to run its help: %w", err)
	}
	c := exec.Command(exe, argv...)
	// ARGV[0] IS KEPT AS THIS PROCESS HAD IT, because InvokedAs reads it and every usage line on the
	// page names the tool through it. os.Executable resolves symlinks, so without this a tool run
	// through a link would print pages naming a program the seat never typed.
	c.Args[0] = os.Args[0]
	var out bytes.Buffer
	c.Stdout, c.Stderr = &out, &out
	if err := c.Run(); err != nil {
		return out.String(), fmt.Errorf("`%s` failed: %w", strings.Join(argv, " "), err)
	}
	return out.String(), nil
}

// inProcessHelp builds the tree for the seat argv names and runs argv against it — what the binary
// does around Execute, minus the process: the pre-parse refusal, the same ExecuteRoot, and the
// error line Execute prints.
func inProcessHelp(argv []string) (string, error) {
	seatID := seatenv.SeatIDIn(argv)
	root := NewRootFor(seatID)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(argv)
	err := refuseUnknownCommandFirst(root, append([]string{InvokedAs()}, argv...), seatID)
	if err == nil {
		err = ExecuteRoot(root)
	}
	if err != nil {
		fmt.Fprintf(&out, "%s: %v\n", InvokedAs(), err)
	}
	return out.String(), err
}
