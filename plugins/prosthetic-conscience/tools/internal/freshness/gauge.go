package freshness

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
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

// readState reports the state AND whether the file could be read at all.
//
// ABSENT AND UNREADABLE ARE DIFFERENT ANSWERS, which is the distinction this design
// enforces everywhere else and did not enforce here. A missing file is an honest empty
// state: nothing has been stamped yet. A file that exists but cannot be read is not
// empty — and treating it as empty makes Of() RE-STAMP tokens_at_write at the current
// count, after which growth measures the interval since the failed read.
//
// MEASURED ON WINDOWS, by CI. A reader can transiently fail to open a file that is being
// renamed over — rename is atomic for the FILE, but a concurrent open can still hit a
// sharing violation — and 21 of 800 reads came back "torn" under four concurrent writers.
// On Linux the same harness reports zero. The retry absorbs the transient case; the
// second return value makes the permanent one honest instead of silently zero.
func readState(path string) (State, bool) {
	var lastErr error
	for attempt := range 3 {
		b, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			return State{}, true // absent: an honest empty state
		}
		if err != nil {
			lastErr = err
			// Transient on Windows while a rename lands. A few milliseconds is far below
			// any hook's budget and far above the window.
			time.Sleep(time.Duration(attempt+1) * 2 * time.Millisecond)
			continue
		}
		var st State
		if err := json.Unmarshal(b, &st); err != nil {
			// Present but undecodable. NOT empty — see the doc comment.
			return State{}, false
		}
		return st, true
	}
	_ = lastErr
	return State{}, false
}

func writeState(path string, st State) {
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".freshness-*.tmp")
	if err != nil {
		return
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		_ = os.Remove(name)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)
		return
	}
	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)
	}
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
