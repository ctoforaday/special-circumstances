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
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpointrestore"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
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
	// SystemMessage is the failure record's, for the human. AdditionalContext above is for the
	// agent, and a hook that is broken is not the agent's business to relay.
	SystemMessage string `json:"systemMessage,omitempty"`
}

// StageEncode is recorded when this event's response cannot be encoded: a restored checkpoint that
// never arrives, which used to be discarded with `_ =` and read as a session with no note.
const StageEncode hookfailures.Stage = "encode"

// merge composes one response from several units.
//
// IT USED TO READ ONLY Stdout AND Watch, which meant a unit's Stderr — the channel hookunit puts a
// PANIC in — was dropped here, before it even reached the debug log. A crashed unit was silent on
// every channel on this event. The panic is now recorded by hookunit itself, and the text is
// forwarded so the debug log keeps its copy.
func merge(results []hookunit.Result) (text string, watch []string, logged string) {
	var parts, logs []string
	seen := map[string]bool{}
	for _, r := range results {
		if t := strings.TrimRight(r.Stdout, "\n"); t != "" {
			parts = append(parts, t)
		}
		if l := strings.TrimRight(r.Stderr, "\n"); l != "" {
			logs = append(logs, l)
		}
		for _, p := range r.Watch {
			if !seen[p] {
				seen[p] = true
				watch = append(watch, p)
			}
		}
	}
	return strings.Join(parts, "\n\n"), watch, strings.Join(logs, "\n")
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, units []hookunit.Unit) int {
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-sessionstart")) {
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in struct {
		CWD string `json:"cwd"`
	}
	_ = json.Unmarshal(raw, &in)

	rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", now, stderr)
	ctx := hookunit.NewCtx("SessionStart", raw, hookenv.ProjectDir(projectDir, in.CWD), now, rec)
	if !hookenv.Explain(ctx.ProjectDir, rec, "sc-sessionstart") {
		// Recorded, and NOT said here: a hook with no project root writes nothing to stdout
		// (projectroot_test.go pins that for every binary). The next event that resolves one
		// announces it.
		rec.Persist()
		return 0
	}

	text, watch, logged := merge(hookunit.Run(ctx, units))
	if logged != "" {
		fmt.Fprintln(stderr, logged)
	}
	said := rec.Settle()
	// Silence is a valid outcome and MUST stay silent. watchPaths alone is reason enough to
	// speak: a note can be worth watching with nothing to say about it.
	if strings.TrimSpace(text) == "" && len(watch) == 0 && said == "" {
		return 0
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "SessionStart"
	out.HookSpecificOutput.AdditionalContext = text
	out.HookSpecificOutput.WatchPaths = watch
	out.SystemMessage = said
	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		// Discarded outright until now (`_ =`), which made a restored checkpoint that never
		// arrived look exactly like a session that had none.
		rec.Fail(StageEncode, "sc-sessionstart: cannot encode response: "+err.Error())
		rec.Persist()
	}
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
