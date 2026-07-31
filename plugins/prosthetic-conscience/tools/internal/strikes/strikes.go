// Package strikes counts repeated tool failures so anti-spinning's 3-strike limit stops
// being enforced by the agent against itself.
//
// THE DEFECT (#127). anti-spinning says "During error or lint repair, YOU MUST NOT attempt
// the same fix more than 3 times" and "AFTER a third consecutive failure, YOU MUST stop".
// Nothing counted. The agent whose loop-detection is impaired is the same agent asked to
// notice it is looping — the failure mode the rule exists to catch, asked to catch itself.
//
// # "Consecutive" is not observable, and this says so rather than pretending
//
// PostToolUseFailure fires only on FAILURE. A success is never seen, so a literal
// consecutive-failure count cannot be computed from this event alone. Two ways out were
// considered:
//
//   - Register on PostToolUse as well, to reset a key on success. That fires a process on
//     EVERY tool call — 230 Bash calls in one measured session — to observe a reset that a
//     cheaper signal approximates.
//   - Bound the count by TIME. A spin is temporally dense: three failures of the same key
//     inside a few minutes is a loop; three across an hour, with work in between, is not.
//
// The window is the approximation, and it is stated in the message so nobody reads the
// count as something stricter than it is.
package strikes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Window bounds what counts as "the same attempt again". Three failures inside it are a
// spin; the same three spread across an afternoon are three separate problems.
const Window = 15 * time.Minute

// Threshold is anti-spinning's own number. It lives here so the counter and the rule cannot
// drift into disagreeing about what a third strike is.
const Threshold = 3

// maxKeys bounds the state file. A session that fails many DIFFERENT things is not spinning,
// and an unbounded map is the resource-ballooning failure this design has hit once already.
const maxKeys = 200

// File is the state's name, under .claude/.
const File = "strikes.json"

// State is the whole file: failure timestamps per key, newest last.
type State struct {
	Keys map[string][]string `json:"keys"`
}

// Path is where the state lives for a project.
func Path(projectDir string) string { return filepath.Join(projectDir, ".claude", File) }

// Load reads the state. A missing or corrupt file yields an empty state: losing strike
// history must never cost a tool call.
func Load(read func(string) ([]byte, error), path string) State {
	s := State{Keys: map[string][]string{}}
	b, err := read(path)
	if err != nil {
		return s
	}
	var parsed State
	if json.Unmarshal(b, &parsed) == nil && parsed.Keys != nil {
		s = parsed
	}
	return s
}

// Save writes the state, best-effort.
func Save(path string, s State) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, append(b, '\n'), 0o644)
}

// Key identifies "the same thing, tried again": the tool plus what it was aimed at.
//
// The target is the tool's own notion of a subject — the command for Bash, the path for an
// edit, the pattern for a search. Two different commands are two different attempts even
// under the same tool, which is the distinction the rule cares about: re-running the SAME
// action is what is forbidden, not failing twice at different things.
func Key(tool string, toolInput []byte) string {
	var in struct {
		Command  string `json:"command"`
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
		Pattern  string `json:"pattern"`
		URL      string `json:"url"`
		Query    string `json:"query"`
	}
	_ = json.Unmarshal(toolInput, &in)

	target := ""
	for _, candidate := range []string{in.Command, in.FilePath, in.Path, in.Pattern, in.URL, in.Query} {
		if strings.TrimSpace(candidate) != "" {
			target = candidate
			break
		}
	}
	// Collapse whitespace so a reformatted-but-identical command is the same attempt.
	target = strings.Join(strings.Fields(target), " ")
	if target == "" {
		return tool
	}
	return tool + "\x00" + target
}

// Record adds a failure and reports how many are live in the window.
//
// The key is RESET once it trips, so a persistent failure nags at 3, 6, 9 rather than on
// every subsequent attempt. A counter that fires forever is one people learn to ignore,
// which would leave the rule exactly as unenforced as it was.
func Record(s *State, key string, now time.Time) (count int, tripped bool) {
	if s.Keys == nil {
		s.Keys = map[string][]string{}
	}
	live := within(s.Keys[key], now)
	live = append(live, now.UTC().Format(time.RFC3339Nano))

	if len(live) >= Threshold {
		delete(s.Keys, key)
		prune(s, now)
		return len(live), true
	}
	s.Keys[key] = live
	prune(s, now)
	return len(live), false
}

// within drops timestamps that have fallen out of the window, and anything unparseable —
// a corrupt entry must not pin a key open forever.
func within(stamps []string, now time.Time) []string {
	var live []string
	for _, s := range stamps {
		t, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			continue
		}
		if now.Sub(t) < Window {
			live = append(live, s)
		}
	}
	return live
}

// prune bounds the file: expired keys go, and if that is not enough the oldest go too.
func prune(s *State, now time.Time) {
	for k, stamps := range s.Keys {
		live := within(stamps, now)
		if len(live) == 0 {
			delete(s.Keys, k)
			continue
		}
		s.Keys[k] = live
	}
	if len(s.Keys) <= maxKeys {
		return
	}
	type aged struct {
		key  string
		last string
	}
	var all []aged
	for k, stamps := range s.Keys {
		all = append(all, aged{k, stamps[len(stamps)-1]})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].last < all[j].last })
	for _, a := range all[:len(all)-maxKeys] {
		delete(s.Keys, a.key)
	}
}
