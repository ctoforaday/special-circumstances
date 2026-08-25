// Package stopnudge is the guarded single emission at a turn boundary.
//
// THE HAZARD AND THE FEATURE ARE THE SAME MECHANISM. Injecting on `Stop` re-invokes the
// model, which is exactly what a nudge wants — one extra turn in which to act on what it
// was told — and is also a loop: measured on client 2.1.240, an unguarded injector fired
// **16 times, produced 35 assistant entries and burned 4,326 output tokens** on a single
// prompt (hook-surface-spike.md §13; 9 firings and 1,186 tokens on 2.1.235, so the cost
// is client-specific and grew).
//
// Everything here exists to make that emission happen at most once. The guards are
// deliberately redundant, because they fail for different reasons:
//
//   - `stop_hook_active` — the client's own flag, false on the first firing and true on
//     all fifteen re-entries (§13). It needs no state and cannot be lost.
//   - the band record — our state, which answers "have we already said this?" across
//     turns, where the client's flag cannot help.
//
// FAIL CLOSED, inverting this package family's posture on purpose. `checkpointseal` and
// `postcompactobserve` are recorders: a failed write costs one observation, so they
// report to stderr and continue. This is not a recorder. An unwritable state file, an
// unresolvable project directory or a marshal error means NO EMISSION — silence is a
// missed nudge, while emitting on an unrecorded band is the loop above.
package stopnudge

import (
	"errors"
	"path/filepath"
	"slices"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/statefile"
)

// Band names the severity that fired. They are ordered, and only the highest one that a
// reading crosses is emitted.
type Band string

const (
	BandNotice Band = "NOTICE"
	BandWarn   Band = "WARN"
	BandUrgent Band = "URGENT"
)

// Thresholds are the band edges, in each measure's own units.
//
// THEY SHIP UNSET, and unset means the nudge does not exist yet. Phase 1 collects the
// distribution these are supposed to come from, and §III fixes the RULE — P50/P75/P90 of
// that distribution, per measure, any-of — before any data exists, so the numbers cannot
// be chosen to suit the result once it does.
type Thresholds struct {
	TurnsNotice, TurnsWarn, TurnsUrgent    int
	GrowthNotice, GrowthWarn, GrowthUrgent int
}

// Configured reports whether any band edge has been set. An unconfigured nudge emits
// nothing AND WRITES NOTHING — see State's doc for why the second half matters.
func (t Thresholds) Configured() bool {
	return t.TurnsNotice > 0 || t.GrowthNotice > 0
}

// State is .claude/checkpoints/nudge.json, owned by this package alone.
//
// A SEPARATE FILE FROM freshness.json, and the separation is a safety property rather
// than tidiness. `freshness.json` is written by four record-writing binaries via
// json.Marshal of a fixed struct, so every write erases keys that struct does not
// declare. Band state living there would be destroyed by an unrelated hook firing on an
// unrelated event — and an erased band is a re-emission on `Stop`, which is the loop.
//
// ITS EXISTENCE IS ALSO A FACT OTHERS READ: the seal record derives `nudge_enabled` from
// whether this file is present, because the file is written by the thing whose liveness
// is the question. So an inert build must not create it — otherwise every row claims the
// nudge was live while it has never fired, and criterion 6 compares an "after" population
// against a "before" that is not one.
type State struct {
	// SessionID scopes everything below. Without it the counters outlive the session
	// they describe: four emissions EVER would suppress the nudge permanently in a
	// project, and "once per band per session" would have no session boundary.
	SessionID string `json:"session_id"`

	BandsSpent []Band `json:"bands_spent"`
	// AnsweredAtSeen is the note's newest timestamp at the moment a band was spent. Bands
	// re-arm by comparing it against the note's current timestamps — the band record is
	// NOT cleared on an answer, because `nudge_answered`'s derivation reads it to know a
	// band fired at all.
	AnsweredAtSeen time.Time `json:"answered_at_seen"`

	Emissions     int `json:"emissions_this_session"`
	EmissionBytes int `json:"emission_bytes_max"`

	BranchHeadSeen    string `json:"branch_head_seen"`
	BranchCommitsSeen int    `json:"branch_commits_seen"`
}

// maxEmissions is criterion 4's hard cap. It suppresses unconditionally, whatever the
// bands say: bands reset when the note is answered, so three bands per note-write cycle
// times two checkpoint writes already exceeds four without a cap that does not reset.
const maxEmissions = 4

// The F7 floor, fixed here rather than discovered later. A session that has done less
// than either of these has nothing a nudge could usefully say, and being told about a
// note it wrote twenty minutes ago is how a mechanism becomes wallpaper.
const (
	floorTurns  = 20
	floorGrowth = 50_000
)

// Decision is what a turn boundary produced.
type Decision struct {
	Emit string // "" means say nothing
	Band Band
}

var (
	errNotConfigured   = errors.New("stopnudge: no thresholds configured")
	errUnreadableState = errors.New("stopnudge: state record exists and cannot be read")
)

// Decide is the whole guard, and it WRITES BEFORE IT RETURNS ANYTHING TO EMIT.
//
// The order is the point. A guard that emits and then records has re-emitted whenever the
// write fails or the process dies between the two — and on `Stop` a re-emission is not a
// duplicate nudge, it is the sixteen-firing loop.
func Decide(dir, sessionID string, stopHookActive bool, m freshness.Measures, notePath string,
	noteNewest time.Time, th Thresholds, now time.Time) Decision {

	// 1. The client's own re-entry flag. No state, nothing to lose, checked first.
	if stopHookActive {
		return Decision{}
	}
	// 2. An unconfigured nudge does not exist: no emission, and NO FILE, because the file
	//    is how the seal record learns the nudge is live.
	if !th.Configured() {
		return Decision{}
	}
	// 3. The floor. Below it the session has done too little to be worth interrupting.
	if !aboveFloor(m) {
		return Decision{}
	}

	st, err := load(dir, sessionID)
	if err != nil {
		return Decision{} // fail closed: cannot read our own record, so say nothing
	}
	// 4. Bands re-arm when the note has moved since we last spoke — an answer, of either
	//    kind. The record is not cleared; the comparison is what re-arms it.
	if !st.AnsweredAtSeen.IsZero() && noteNewest.After(st.AnsweredAtSeen) {
		st.BandsSpent = nil
	}
	// 5. The hard cap, which does not reset when bands do.
	if st.Emissions >= maxEmissions {
		return Decision{}
	}

	band, ok := highestBand(m, th)
	if !ok || slices.Contains(st.BandsSpent, band) {
		return Decision{}
	}

	line := freshness.Render(m, notePath)
	st.BandsSpent = append(st.BandsSpent, band)
	st.AnsweredAtSeen = noteNewest
	st.Emissions++
	if len(line) > st.EmissionBytes {
		st.EmissionBytes = len(line)
	}

	// WRITE BEFORE EMIT. If this fails, nothing is emitted — the band would be unrecorded
	// and the next boundary would say it again.
	if err := save(dir, st); err != nil {
		return Decision{}
	}
	return Decision{Emit: line, Band: band}
}

func aboveFloor(m freshness.Measures) bool {
	if m.TurnsMeasured && m.Turns >= floorTurns {
		return true
	}
	return m.GrowthKnown && m.Growth >= floorGrowth
}

// highestBand returns the most severe band any measure crosses — ANY-OF, taking the max.
//
// Any-of because the measures fail independently: a session that burned 400k tokens in
// twelve turns and one that took 300 turns without moving the token count are both stale,
// and an all-of rule would silence exactly the lopsided cases three measures exist to
// catch.
func highestBand(m freshness.Measures, th Thresholds) (Band, bool) {
	crosses := func(turns, growth int) bool {
		if turns > 0 && m.TurnsMeasured && m.Turns >= turns {
			return true
		}
		return growth > 0 && m.GrowthKnown && m.Growth >= growth
	}
	switch {
	case crosses(th.TurnsUrgent, th.GrowthUrgent):
		return BandUrgent, true
	case crosses(th.TurnsWarn, th.GrowthWarn):
		return BandWarn, true
	case crosses(th.TurnsNotice, th.GrowthNotice):
		return BandNotice, true
	}
	return "", false
}

func statePath(dir string) string { return filepath.Join(dir, ".claude", "checkpoints", "nudge.json") }

// load reads this session's state.
//
// A record belonging to a DIFFERENT session is not this session's state: the counters
// reset and the bands clear, because a new session has a fresh budget and has been told
// nothing yet.
//
// THIS COPY PREVIOUSLY LACKED THE RETRY AND THE TRI-STATE that freshness had. Both now
// come from internal/statefile, so a transient read failure on Windows no longer silently
// suppresses a nudge — and an unreadable record is still refused, which is this package's
// own policy and stricter than the gauge's.
func load(dir, sessionID string) (State, error) {
	st, status := statefile.Read[State](statePath(dir))
	switch status {
	case statefile.Absent:
		return State{SessionID: sessionID}, nil
	case statefile.Unreadable:
		// A corrupt or unreadable record is NOT an empty one. Treating it as empty
		// re-emits every band this session already spent, and on Stop a re-emission is a
		// loop rather than a duplicate.
		return State{}, errUnreadableState
	}
	if st.SessionID != sessionID {
		return State{SessionID: sessionID}, nil
	}
	return st, nil
}

// save writes the record, atomically, before anything is emitted.
//
// The atomicity matters for the same reason it does in the gauge — a truncating write
// leaves a window in which a concurrent reader sees an empty file and concludes no band
// has been spent — and the implementation is shared rather than repeated.
func save(dir string, st State) error {
	return statefile.Write(statePath(dir), st)
}
