package stopnudge

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
)

// hookInput is the Stop payload. Every field here was MEASURED on client 2.1.240
// (hook-surface-spike.md §9c, §13) rather than taken from documentation.
//
// Note what is absent: no agent_id and no agent_type. Stop carries none of the agent_*
// fields, which is why the debounce is keyed on session_id alone — a composite key would
// have kept half of itself permanently empty.
type hookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	StopHookActive bool   `json:"stop_hook_active"`
}

type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext,omitempty"`
	} `json:"hookSpecificOutput"`
}

// configured reads the band edges.
//
// THEY DO NOT EXIST YET, and that is the current, correct state: Phase 1 is collecting
// the distribution they must come from, and §III fixes the RULE (P50/P75/P90 per measure,
// any-of) before the data exists so the numbers cannot be chosen to flatter the result.
// Until then this returns the zero value, the nudge emits nothing, and — because
// nudge_enabled is read from nudge.json's presence — it leaves no trace that would tell
// the seal record otherwise.
func configured() Thresholds { return Thresholds{} }

// run takes its thresholds as an ARGUMENT rather than reading configured() directly, so
// the emission path can be driven in a test.
//
// That is not test scaffolding for its own sake. Thresholds are unset today, so with a
// hardcoded source the ONLY reachable path is the silent one — and the branch that would
// go untested is the one that composes the response the client actually reads. A hook
// whose emit path has never run is a hook whose first real emission is its first
// execution of that code.
func run(stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time, th Thresholds) int {
	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-stop") {
		return 0
	}

	notePath := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	if notePath == "" {
		return 0 // no note: nothing whose age could be reported
	}
	body, err := os.ReadFile(notePath)
	if err != nil {
		return 0
	}

	n := checkpoint.Parse(string(body))
	m := freshness.Of(projectDir, in.TranscriptPath, string(body),
		freshness.BranchWork(n.Get("head")), now)

	d := Decide(projectDir, in.SessionID, in.StopHookActive, m, notePath, newest(n), th, now)
	if d.Emit == "" {
		return 0
	}

	var out hookOutput
	out.HookSpecificOutput.HookEventName = "Stop"
	out.HookSpecificOutput.AdditionalContext = d.Emit
	if err := json.NewEncoder(stdout).Encode(out); err != nil {
		fmt.Fprintln(stderr, "sc-stop: cannot encode response:", err)
	}
	return 0
}

// newest is the later of written_at and reaffirmed_at — the note's most recent ANSWER of
// either kind. Bands re-arm against this, so a re-affirmation closes a band exactly as a
// rewrite does, which is what the skill clause promises when it calls a reasoned "still
// accurate" a valid answer.
func newest(n checkpoint.Note) time.Time {
	var out time.Time
	for _, k := range []string{"written_at", "reaffirmed_at"} {
		if t, err := time.Parse(time.RFC3339, n.Get(k)); err == nil && t.After(out) {
			out = t
		}
	}
	return out
}

// Main is the process boundary: it wires the real environment in and returns the exit
// code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println(buildid.Line("sc-stop"))
		return 0
	}
	wd, _ := os.Getwd()
	return run(os.Stdin, os.Stdout, os.Stderr, wd, time.Now(), configured())
}
