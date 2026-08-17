// Package sessionstart is the ONE hook on the SessionStart event.
//
// It replaces sc-toolchain-nudge and sc-checkpoint-restore as separate binaries. They remain
// separate UNITS; what changes is that ONE response is now composed from both.
//
// # Why this event forced a correction
//
// The two binaries were emitting on DIFFERENT CHANNELS, and separate processes hid it:
//
//	sc-toolchain-nudge     printed a bare line to stdout, on the strength of a comment
//	                       reading "SessionStart stdout reaches the session"
//	sc-checkpoint-restore  emitted a JSON document carrying additionalContext + watchPaths
//
// One process cannot do both: a bare line on stdout corrupts the document beside it. And the
// claim the nudge rested on is NOT in plans/hook-surface-spike.md §5's verified list — what
// IS verified there is additionalContext ("hook_additional_context attachment, transcript
// line 42"). So the merge puts both on the channel that was actually measured, which is a
// behaviour change for the nudge and the reason this is not a pure refactor.
//
// # The decision merge, which is this event's alone
//
// COMPOSE, not pick-first. additionalContext is the concatenation of every unit's text in
// unit order; watchPaths is the union, deduplicated. That is the opposite of PreToolUse,
// where only one permission document may be emitted and the first denial wins, and different
// again from PostToolUse, where the merge is over exit codes. Three events, three policies —
// which is why hookunit still does not own one.
//
// Silence stays silent: an empty response would spend tokens telling the session there is
// nothing to tell it. Either half is enough to speak, because a note can be worth WATCHING
// with nothing to say about it.
package sessionstart

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpointrestore"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchainnudge"
)

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string   `json:"hookEventName"`
		AdditionalContext string   `json:"additionalContext,omitempty"`
		WatchPaths        []string `json:"watchPaths,omitempty"`
	} `json:"hookSpecificOutput"`
}

// merge composes one response from several units.
func merge(results []hookunit.Result) (text string, watch []string) {
	var parts []string
	seen := map[string]bool{}
	for _, r := range results {
		if t := strings.TrimRight(r.Stdout, "\n"); t != "" {
			parts = append(parts, t)
		}
		for _, p := range r.Watch {
			if !seen[p] {
				seen[p] = true
				watch = append(watch, p)
			}
		}
	}
	return strings.Join(parts, "\n\n"), watch
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	fs := flag.NewFlagSet("sc-sessionstart", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag never wedges a session start
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-sessionstart"))
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in struct {
		CWD string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &in)

	ctx := hookunit.NewCtx("SessionStart", raw, hookenv.ProjectDir(projectDir, in.CWD), now)
	if !hookenv.Explain(ctx.ProjectDir, stderr, "sc-sessionstart") {
		return 0
	}

	text, watch := merge(hookunit.Run(ctx, units))
	// Silence is a valid outcome and MUST stay silent. watchPaths alone is reason enough to
	// speak: a note can be worth watching with nothing to say about it.
	if strings.TrimSpace(text) == "" && len(watch) == 0 {
		return 0
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "SessionStart"
	out.HookSpecificOutput.AdditionalContext = text
	out.HookSpecificOutput.WatchPaths = watch
	_ = json.NewEncoder(stdout).Encode(out)
	return 0
}

// Units are this event's checks, in the order their text should appear.
func Units(pluginRoot string) []hookunit.Unit {
	return []hookunit.Unit{
		toolchainnudge.Unit(pluginRoot, toolchain.Probe),
		checkpointrestore.Unit(),
	}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now(), Units(os.Getenv("CLAUDE_PLUGIN_ROOT")))
}
