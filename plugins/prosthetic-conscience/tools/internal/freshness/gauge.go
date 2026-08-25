package freshness

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/statefile"
)

// Of measures a note's age for a caller that has a project directory and a transcript.
//
// It lives here rather than in either caller because BOTH the seal record and the
// compaction row need the same three figures, and a second copy of this composition
// is how the two records start disagreeing about what "age" meant.
//
// Best-effort in every direction: a seal whose transcript is missing still writes its
// row with the age unmeasured, because the trigger, the time and the handles are facts
// the seal itself holds. Dropping the row because one measurement failed would lose the
// boundary from the baseline entirely — and the baseline is a distribution over
// boundaries, so a missing one is not a missing number, it is a missing observation.
func Of(projectDir, transcriptPath, note string, branch Branch, now time.Time) Measures {
	n := checkpoint.Parse(note)
	writtenAt, err := time.Parse(time.RFC3339, n.Get("written_at"))
	if err != nil {
		// No written_at — a note from before schema 3, or one written wrong. There is no
		// reference point, so there is no age. Unmeasured, never zero.
		return Measures{}
	}

	m, _ := ctxusage.Read(transcriptPath, writtenAt)

	stPath := filepath.Join(projectDir, ".claude", "checkpoints", "freshness.json")
	st, ok := readState(stPath)
	if !ok {
		// The record exists and could not be read. Stamping now would overwrite a reading
		// we cannot see with one taken at this moment, and every later growth figure would
		// measure from here rather than from the note. Report what needs no state, and
		// leave the file alone.
		return Gauge(State{}, m, branch)
	}
	st, justStamped := ObserveAndSay(st, writtenAt, m, now)
	writeState(stPath, st)

	// justStamped suppresses growth on this row: the reading it would measure against
	// was taken in this same call, so the answer would be a manufactured zero.
	return GaugeAfter(st, m, branch, justStamped)
}

// readState loads the gauge's state, reporting whether it could be read at all.
//
// The tri-state and the atomic write now live in internal/statefile, because this file
// and internal/stopnudge each grew their own copy and only this one learned — from CI —
// that a concurrent reader on Windows can transiently fail to open a file being renamed
// over. Two implementations of one idea, diverged in robustness within hours.
//
// The POLICY stays here, because it is not shared: an unreadable record means this gauge
// must not re-stamp, since stamping over a reading it cannot see makes every later growth
// figure measure from this moment rather than from the note.
func readState(path string) (State, bool) {
	st, status := statefile.Read[State](path)
	switch status {
	case statefile.Absent:
		return State{}, true // nothing stamped yet: an honest empty state
	case statefile.Present:
		return st, true
	}
	return State{}, false
}

func writeState(path string, st State) {
	// Best-effort, matching this package's posture: a lost write costs one observation,
	// and a hook must not fail an event over provenance.
	_ = statefile.Write(path, st)
}

// branchWork counts commits on THIS BRANCH'S line since the note's head.
//
// --first-parent for the same reason checkpointrestore uses it: a plain count is
// dominated by other people's work arriving through merges, and reports a note written
// before a routine merge as a hundred commits stale having done nothing.
func BranchWork(head string) Branch {
	if head == "" || head == "null" {
		return Branch{}
	}
	out, err := exec.Command("git", "rev-list", "--count", "--first-parent", head+"..HEAD").Output()
	if err != nil {
		return Branch{}
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return Branch{}
	}
	return Branch{Commits: n, Known: true}
}
