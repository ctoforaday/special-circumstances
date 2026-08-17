// Package pretooluse is the ONE hook on the PreToolUse event.
//
// It replaces sc-secrets-gate and sc-push-freeze-guard as separate binaries. They remain
// separate UNITS, and their POSTURES ARE OPPOSITE — which is the whole difficulty of this
// event and the reason the merge policy is written here rather than generalised:
//
//	sc-secrets-gate      fails CLOSED. It can DENY, and it denies on an undecodable
//	                     payload rather than allowing (#211).
//	sc-push-freeze-guard NEVER blocks. The freeze is a commitment the human may
//	                     consciously override; the guard makes it impossible to forget,
//	                     not impossible to break.
//
// A merge that flattened those into one posture would either turn a security gate into a
// warning or turn an advisory into a blocker. Each unit keeps its own.
//
// # The decision merge
//
// DENY WINS, and only one decision can be emitted: the PreToolUse protocol reads a single
// permission document from stdout. So stdout is PICK-FIRST in unit order, not composed —
// unlike PostToolUse, where every unit's stderr must survive, and unlike SessionStart, where
// one response is assembled from several. Three events, three policies; that is why
// hookunit does not own one.
//
// Warnings are independent of the decision: a unit that only warns still warns when another
// unit denies, because the operator needs both facts.
//
// The exit code is ALWAYS 0. The block travels in the JSON, never in the status.
package pretooluse

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/pushfreezeguard"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secretsgate"
)

// merge is the event's decision policy, kept pure so it can be tested without units.
func merge(results []hookunit.Result) (decision, warnings string) {
	var warn []string
	for _, r := range results {
		if decision == "" && strings.TrimSpace(r.Stdout) != "" {
			decision = r.Stdout // first denial in unit order wins; only one may be emitted
		}
		if w := strings.TrimRight(r.Stderr, "\n"); w != "" {
			warn = append(warn, w)
		}
	}
	return decision, strings.Join(warn, "\n")
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	fs := flag.NewFlagSet("sc-pretooluse", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-pretooluse"))
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in struct {
		CWD string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &in)

	ctx := hookunit.NewCtx("PreToolUse", raw, hookenv.ProjectDir(projectDir, in.CWD), now)
	decision, warnings := merge(hookunit.Run(ctx, units))

	if warnings != "" {
		fmt.Fprintln(stderr, warnings)
	}
	if decision != "" {
		fmt.Fprint(stdout, decision)
	}
	// ALWAYS 0: the block travels in the JSON, never in the status.
	return 0
}

// Units are this event's checks, in the order a decision is taken from.
func Units() []hookunit.Unit {
	return []hookunit.Unit{secretsgate.Unit(), pushfreezeguard.Unit()}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), Units())
}
