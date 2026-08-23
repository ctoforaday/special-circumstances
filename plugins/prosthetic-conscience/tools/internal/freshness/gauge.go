package freshness

import (
	"encoding/json"
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
func Of(projectDir, transcriptPath, note string, branch Branch) Measures {
	n := checkpoint.Parse(note)
	writtenAt, err := time.Parse(time.RFC3339, n.Get("written_at"))
	if err != nil {
		// No written_at — a note from before schema 3, or one written wrong. There is no
		// reference point, so there is no age. Unmeasured, never zero.
		return Measures{}
	}

	m, _ := ctxusage.Read(transcriptPath, writtenAt)

	stPath := filepath.Join(projectDir, ".claude", "checkpoints", "freshness.json")
	st := readState(stPath)
	st, justStamped := ObserveAndSay(st, writtenAt, m)
	writeState(stPath, st)

	// justStamped suppresses growth on this row: the reading it would measure against
	// was taken in this same call, so the answer would be a manufactured zero.
	return GaugeAfter(st, m, branch, justStamped)
}

// readState returns the zero State for anything it cannot read. A corrupt or absent
// state file costs one note's growth measurement, which the row then reports as
// unmeasured — the honest outcome, and better than a stale reading presented as current.
func readState(path string) State {
	b, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return State{}
	}
	return st
}

func writeState(path string, st State) {
	b, err := json.Marshal(st)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, b, 0o644)
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
