package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// grayAreaPlugin is the key gray-area registers itself under in `enabledPlugins`.
const grayAreaPlugin = "gray-area@special-circumstances"

// settingsFile is the shape this check reads. Both fields are POINTERS because the honest
// answer to each is three-state, and a plain value would collapse two of them.
//
// `showThinkingSummaries` absent is not the same as false. Absent is the default nobody
// changed; false is a decision somebody made. The warning is the same either way — the
// summaries are off — but a doctor that reports a deliberate choice in the same words as an
// oversight teaches the reader to stop reading it.
type settingsFile struct {
	ShowThinkingSummaries *bool           `json:"showThinkingSummaries"`
	EnabledPlugins        map[string]bool `json:"enabledPlugins"`
}

// settingsChain is the precedence order Claude Code merges, strongest first. The FIRST file
// that DEFINES a key decides it — which is why "does any file say true" is the wrong question
// and is not what this asks: a project-local `false` overrides a user-level `true`, and a
// check that OR-ed them together would report capture as live on a box where it is off.
func settingsChain(home, project string) []string {
	return []string{
		filepath.Join(project, ".claude", "settings.local.json"),
		filepath.Join(project, ".claude", "settings.json"),
		filepath.Join(home, ".claude", "settings.json"),
	}
}

// thinkingSummariesOn resolves the setting across the chain, and reports whether anything
// answered at all.
//
// The third return is what keeps this from being a coin flip reported as a measurement. A
// chain in which every file is missing or unparseable yields the same `false` as a chain that
// plainly says false, and those two must not reach the caller as one value — see
// [[facts-are-fields]] clause 3.
func thinkingSummariesOn(chain []string, read func(string) ([]byte, error)) (on bool, explicit bool, readable bool) {
	for _, p := range chain {
		body, err := read(p)
		if err != nil {
			continue // absent is normal for a chain; unreadable is handled by readable below
		}
		readable = true
		var s settingsFile
		if json.Unmarshal(body, &s) != nil {
			continue // malformed: it answered nothing, and a later file may answer
		}
		if s.ShowThinkingSummaries != nil {
			return *s.ShowThinkingSummaries, true, true
		}
	}
	return false, false, readable
}

// grayAreaEnabled reports whether gray-area is turned on anywhere in the chain. Unlike the
// setting above, enablement is genuinely a union: a plugin enabled at user level and not
// mentioned in a project file is enabled there too.
func grayAreaEnabled(chain []string, read func(string) ([]byte, error)) bool {
	for _, p := range chain {
		body, err := read(p)
		if err != nil {
			continue
		}
		var s settingsFile
		if json.Unmarshal(body, &s) != nil {
			continue
		}
		if s.EnabledPlugins[grayAreaPlugin] {
			return true
		}
	}
	return false
}

// captureWarnings reports when gray-area is installed to capture trajectories and the one
// setting that decides whether reasoning is IN them is off.
//
// # Why this is the doctor's business and not gray-area's own
//
// `showThinkingSummaries` defaults to false, and the client's own resolver is
// `AEn(explicitDisplay, isNonInteractive, ...)` — interactive sessions request summaries only
// when the setting is true. Measured on this repository's own box: of roughly 5,000 thinking
// blocks across recent transcripts, 161 were non-empty. The rest are `thinking: ""` beside a
// signature. Not redaction, not a model that declined to reason — a default nobody changed.
//
// The cost of leaving it off is the part worth stating in the warning. Thinking is enabled
// automatically for supported models (`alwaysThinkingEnabled` is true when absent) and its
// tokens are BILLED IN FULL whether or not a summary is kept. So the box is already paying for
// the reasoning and discarding the only readable record of it, and gray-area captures a
// trajectory whose deeds are complete and whose thought is empty.
//
// It belongs to the doctor because the setting is user-level configuration, outside any
// plugin's tree: gray-area can observe the consequence but cannot see the cause, and a plugin
// warning about a file it does not own would be guessing. `plans/reasoning-telemetry.md` §5
// has recommended this setting since it was written; nothing checked, so nothing applied it.
func captureWarnings(home, project string, read func(string) ([]byte, error)) []string {
	chain := settingsChain(home, project)
	if !grayAreaEnabled(chain, read) {
		// Nothing to warn about: the capture this setting feeds is not installed.
		return nil
	}
	on, explicit, readable := thinkingSummariesOn(chain, read)
	if !readable {
		// SAID, NOT SWALLOWED. gray-area resolved as enabled, so something was readable a moment
		// ago; if nothing is now, the check did not run and must not pass for silence.
		return []string{"THINKING CAPTURE NOT CHECKED: no settings file in the chain could be read, so whether `showThinkingSummaries` is on could not be determined — this is not the same as it being on."}
	}
	if on {
		return nil
	}
	how := "is not set (it defaults to off)"
	if explicit {
		how = "is set to false"
	}
	return []string{fmt.Sprintf(
		"THINKING CAPTURE OFF: gray-area is enabled but `showThinkingSummaries` %s, so captured trajectories record deeds and no reasoning. "+
			"Thinking still runs and is still billed — only the readable summary is discarded. Set \"showThinkingSummaries\": true in ~/.claude/settings.json.", how)}
}

// boxWarnings is the seam between the check and the run that prints it, and it is a variable
// so the WIRING can be pinned separately from the MECHANISM.
//
// The two fail apart. Delete the call in run() and every test in capture_test.go stays green
// while the doctor silently stops asking the question — which is the same shape as the defect
// being checked for: a capture that is off while everything reads as fine. Substituting this
// in a test is the only way to assert the warning actually reaches the verdict.
var boxWarnings = func() []string {
	// The project half of the chain is the directory the doctor was invoked from; a settings
	// file beside the work is the one that governs it.
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	return captureWarnings(homeDir(), cwd, os.ReadFile)
}

// homeDir is the user's home, or "" when it cannot be determined — in which case the chain
// simply carries no user-level entry rather than one rooted at a guess.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}
