// Package posttooluse is the ONE hook on the PostToolUse event.
//
// It replaced sc-quality-gate and sc-recall-index as separate binaries; the checks stayed
// separate UNITS — own packages, own tests — running in one process over one shared context.
// The recall-index unit was RETIRED with the retrieval layer (2026-08-04), so one unit ships
// today. The fan-in stays: it is this event's merge policy (below), not a headcount.
//
// # Why, given the client already ran them in parallel
//
// Measured (hook-surface-spike.md §4b): an event's hooks run in PARALLEL, so two trivial
// independent processes were already optimal and this merge must run its units concurrently
// simply to MATCH that. Parity is the floor.
//
// The gain is everything separate processes could not share:
//
//   - The payload is parsed ONCE, and the project root resolved ONCE, per event.
//   - ONE writer for the hook log. Two processes appending to one file on the same event
//     were concurrent BY DESIGN, not by accident — the units now RETURN their log lines and
//     this binary writes them in order.
//   - One place for the expensive reads to come (transcript, SQLite, git state): Ctx.Memo
//     makes the second unit to want one pay nothing, which process isolation cannot do at
//     any price.
//
// # The decision merge, which is this event's and not general
//
// PostToolUse exit 2 is a FEEDBACK channel: the write already happened, and 2 is how stderr
// reaches the model. So the merge is "any unit wanting 2 wins, and every unit's stderr
// survives" — losing one unit's message because another had nothing to say would be the
// regression that matters here. Other events merge differently (PreToolUse can DENY;
// SessionStart composes one response), which is why no general policy is invented.
package posttooluse

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hooklog"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/qualitygate"
)

// merge is the event's decision policy, kept pure so it can be tested without units.
//
// Exit: the highest any unit asked for. Stderr: every unit's, in unit order, so a quiet
// unit never suppresses a loud one. Log: every line, for a single ordered write.
func merge(results []hookunit.Result) (stderr string, exit int, logs []string) {
	var msgs []string
	for _, r := range results {
		if r.Exit > exit {
			exit = r.Exit
		}
		if s := strings.TrimRight(r.Stderr, "\n"); s != "" {
			msgs = append(msgs, s)
		}
		if r.Log != "" {
			logs = append(logs, r.Log)
		}
	}
	return strings.Join(msgs, "\n"), exit, logs
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-posttooluse")) {
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in struct {
		CWD string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &in)

	// Parsed once, resolved once — the point of the merge.
	ctx := hookunit.NewCtx("PostToolUse", raw, hookenv.ProjectDir(projectDir, in.CWD), now)

	feedback, exit, logs := merge(hookunit.Run(ctx, units))

	// ONE writer, in unit order.
	hooklog.Append(ctx.ProjectDir, logs...)

	if feedback != "" {
		fmt.Fprintln(stderr, feedback)
	}
	return exit
}

// Units are this event's checks, in the order their output should appear.
func Units() []hookunit.Unit {
	return []hookunit.Unit{qualitygate.Unit()}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), Units())
}
