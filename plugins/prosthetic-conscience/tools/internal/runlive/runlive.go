// Package runlive reads .claude/run-live.json — the commitment-as-state pattern from the
// design-by-contract rule: promises about future behavior (push freeze on pinned paths, no
// plugin updates mid-run) are recorded as observable state by the frank-exchange-of-views
// `setup` verb, consulted by deterministic guards here, and removed by its capture verb.
// Memory is not an enforcement mechanism; this file is.
//
// THE FILE CROSSES A MODULE BOUNDARY AND THE SHAPE HAS ALREADY DRIFTED ONCE ACROSS IT.
// frank-exchange-of-views owns the writer; this module cannot import it (separate modules,
// both internal/), so the shape is restated here — and a restatement is a copy that can go
// stale. It did: #529 made the file a LIST of open runs, feov's own doc records that the
// change broke a private decoder inside that module, and THIS decoder — one module further
// out — was not migrated at all. json.Unmarshal of {"runs":[...]} into the old flat struct
// succeeds with every field zero, so both guards below kept firing with nothing in them:
// "a research run is LIVE (, started )" and "pinned paths are FROZEN: " followed by no
// paths. The warning still appeared, so it read as working.
//
// Two things follow, and both are load-bearing. Read understands the list shape AND the
// retired flat one, so a marker written by either vintage is still seen. And a file that
// is readable but matches NEITHER shape no longer parses to a contentless "live" — it
// reports Unrecognised, and the guards say so in as many words. Fail-open is preserved
// (an unreadable-shaped marker still means treat the run as live), but the miss is LOUD
// rather than a plausible zero. The next shape change announces itself instead of
// quietly emptying the parentheses.
//
// testdata/run-live.golden.json is the shape contract in a form both modules can hold: the
// feov writer's own test asserts it produces exactly those bytes, this package's test
// asserts it can read them, and scripts/check gates the two copies byte-for-byte. A shape
// change now fails on the writer's side first, and cannot reach a release without either
// this reader migrating or the gate going red.
package runlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Marker is one open run. It states only the fields the guards here consume; feov's writer
// records more (a run id and script path, so a stale marker is resumable rather than
// litter), and unknown fields are ignored by design — see State.Unrecognised for the case
// where NO known field is present, which is the one that must not pass silently.
type Marker struct {
	RunDir      string   `json:"runDir"`
	PinnedPaths []string `json:"pinnedPaths"`
	Started     string   `json:"started"`
}

// State is what the marker file says, including the case where it says something this
// reader does not understand. The bool is the whole point: without it, "no runs" and "a
// shape I cannot parse" are the same empty slice, and that equivalence is what shipped.
type State struct {
	// Runs are the open runs, in the order the file holds them.
	Runs []Marker
	// Unrecognised is set when the file was present and valid JSON but carried neither
	// the list shape nor the retired flat one. The run may well be live, so guards must
	// still warn — but they must say that the shape is unknown rather than print blanks.
	Unrecognised bool
}

// Live reports whether the guards should treat a run as in progress. An unrecognised
// shape counts as live: the file exists, something wrote it, and the safe reading of
// "I cannot tell" is that the freeze applies.
func (s State) Live() bool { return len(s.Runs) > 0 || s.Unrecognised }

// Describe names the live runs for a guard's message, and names the UNRECOGNISED case as
// itself. It lives here rather than in each guard because two packages phrasing the same
// state independently is how the shapes drifted in the first place — the package that owns
// the shape owns how it is said out loud.
//
// Empty when nothing is live, so a caller can use it as the whole test.
func (s State) Describe() string {
	if s.Unrecognised {
		return "a research run marker is present but its SHAPE is UNRECOGNISED by this guard " +
			"(the writer moved and this reader did not) — treat the run as LIVE"
	}
	switch len(s.Runs) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("a research run is LIVE (%s, started %s)", s.Runs[0].RunDir, s.Runs[0].Started)
	default:
		dirs := make([]string, 0, len(s.Runs))
		for _, r := range s.Runs {
			dirs = append(dirs, r.RunDir)
		}
		return fmt.Sprintf("%d research runs are LIVE (%s)", len(s.Runs), strings.Join(dirs, ", "))
	}
}

// Pinned is every pinned path across every open run, de-duplicated, first-seen order kept.
// Concurrent runs pin independently and a push touching either one violates that run's
// freeze, so the union is the honest answer rather than any single run's list.
func (s State) Pinned() []string {
	var out []string
	seen := map[string]bool{}
	for _, r := range s.Runs {
		for _, p := range r.PinnedPaths {
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
}

// Read returns what .claude/run-live.json under projectDir says.
//
// A missing or malformed file means "not live" — guards fail open, because a broken file
// is evidence of a broken file and not of an open run, and blocking on it would trade a
// silent hazard for a stuck operator.
func Read(projectDir string) State {
	if projectDir == "" {
		return State{}
	}
	raw, err := os.ReadFile(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return State{}
	}
	// Pointers, so presence of a KEY is distinguishable from its zero value: that is what
	// separates "a list of no runs" (the file says nothing is open) from "a shape with no
	// runs key at all" (this reader is out of date), which the old struct could not do.
	var probe struct {
		Runs        *[]Marker `json:"runs"`
		RunDir      *string   `json:"runDir"`
		PinnedPaths []string  `json:"pinnedPaths"`
		Started     *string   `json:"started"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return State{}
	}
	switch {
	case probe.Runs != nil:
		return State{Runs: *probe.Runs}
	case probe.RunDir != nil:
		// The retired singleton (pre-#529). Still read, so a marker left by an older
		// writer is not invisible to the guards.
		m := Marker{RunDir: *probe.RunDir, PinnedPaths: probe.PinnedPaths}
		if probe.Started != nil {
			m.Started = *probe.Started
		}
		return State{Runs: []Marker{m}}
	default:
		return State{Unrecognised: true}
	}
}
