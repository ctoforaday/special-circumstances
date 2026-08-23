package freshness

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const note3 = "---\nschema: 3\nwritten_at: 2026-08-23T00:10:00Z\nhead: %s\n---\n## Validation loop\n1. x\n"

// project builds a directory with a note and a transcript, and returns both paths.
func project(t *testing.T, head string, lines ...string) (dir, transcript string) {
	t.Helper()
	dir = t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Replace(note3, "%s", head, 1)
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	transcript = filepath.Join(dir, "t.jsonl")
	if err := os.WriteFile(transcript, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, transcript
}

func entry(min, tokens int) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":"2026-08-23T00:%02d:00Z","message":{"usage":{"input_tokens":0,"cache_read_input_tokens":%d,"cache_creation_input_tokens":0}}}`, min, tokens)
}

// Of is the composition every record writer calls, and its whole point is that the
// stamp PERSISTS between invocations — the two callers are separate processes at
// separate boundaries, so a reading held only in memory measures nothing.
func TestOfStampsOnceAndGaugesAgainstItNextTime(t *testing.T) {
	dir, tp := project(t, "", entry(20, 100_000))
	body, _ := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"))

	first := Of(dir, tp, string(body), Branch{}, ts(30))
	if first.GrowthKnown {
		t.Errorf("growth measured on the observation that created its own baseline: %d", first.Growth)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "freshness.json")); err != nil {
		t.Fatalf("no state written, so the next boundary has nothing to measure against: %v", err)
	}

	// A later boundary, more context: the stamp must survive and the growth be real.
	if err := os.WriteFile(tp, []byte(entry(20, 100_000)+"\n"+entry(40, 180_000)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second := Of(dir, tp, string(body), Branch{}, ts(50))
	if !second.GrowthKnown {
		t.Fatal("growth unmeasured on the second observation; the stamp did not survive the process boundary")
	}
	if second.Growth != 80_000 {
		t.Errorf("Growth = %d, want 80000", second.Growth)
	}
}

// A corrupt state file costs one note's growth measurement and must cost nothing else.
// Reading it as an empty state would RE-STAMP at the current count, and growth would
// then measure the interval since the corruption — a confident number for an interval
// nobody chose.
func TestOfTreatsACorruptStateFileAsUnmeasuredRatherThanRestamping(t *testing.T) {
	dir, tp := project(t, "", entry(20, 100_000))
	body, _ := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"))
	sp := filepath.Join(dir, ".claude", "checkpoints", "freshness.json")
	if err := os.WriteFile(sp, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := Of(dir, tp, string(body), Branch{}, ts(30))
	if m.GrowthKnown {
		t.Errorf("growth reported (%d) over a corrupt state file", m.Growth)
	}
}

// The write is atomic — temp file plus rename — because four binaries write this file
// and the client runs an event's hooks in parallel. A successful write must leave no
// temp files behind, or the checkpoints directory fills with debris that later reads
// have to ignore.
func TestWriteStateLeavesNoTempFilesBehind(t *testing.T) {
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(cp, "freshness.json")
	for i := range 3 {
		writeState(p, State{TokensAtWrite: 1000 * (i + 1), HasWriteReading: true})
	}
	entries, err := os.ReadDir(cp)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || names[0] != "freshness.json" {
		t.Errorf("checkpoints dir = %v, want just freshness.json — temp files were left behind", names)
	}
	got := readState(p)
	if got.TokensAtWrite != 3000 {
		t.Errorf("TokensAtWrite = %d, want 3000 from the last write", got.TokensAtWrite)
	}
}

// An unreadable state file is not an empty one, for the same reason a corrupt one is not.
func TestReadStateOnAMissingFileIsTheZeroState(t *testing.T) {
	st := readState(filepath.Join(t.TempDir(), "nope.json"))
	if st.HasWriteReading {
		t.Errorf("a missing state file reported a write reading: %+v", st)
	}
}

// BranchWork counts THIS BRANCH'S line. The same --first-parent argument as
// checkpointrestore: a plain count is dominated by other people's work arriving through
// merges, and both call sites must answer the same question or the two records disagree.
func TestBranchWorkCountsFirstParentOnly(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(cmd.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.invalid",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.invalid",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	git("commit", "-q", "--allow-empty", "-m", "A")
	base := git("rev-parse", "--short=7", "HEAD")
	git("checkout", "-q", "-b", "side")
	git("commit", "-q", "--allow-empty", "-m", "B")
	git("checkout", "-q", "main")
	git("commit", "-q", "--allow-empty", "-m", "C")
	git("merge", "-q", "--no-ff", "-m", "M", "side")
	t.Chdir(dir)

	b := BranchWork(base)
	if !b.Known {
		t.Fatal("branch work unknown in a repo that contains the head")
	}
	if b.Commits != 2 {
		t.Errorf("Commits = %d, want 2 (C and the merge) — a plain count would say 3 and "+
			"report the side branch's work as this session's", b.Commits)
	}
}

// An absent or unresolvable head is UNKNOWN, never zero: "the note names no commit" and
// "no work has landed since it" are different claims, and a zero would make a note whose
// head was rebased away look current.
func TestBranchWorkIsUnknownRatherThanZero(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	for _, head := range []string{"", "null", "0000000"} {
		if b := BranchWork(head); b.Known {
			t.Errorf("BranchWork(%q) reported known=%v commits=%d; want unknown", head, b.Known, b.Commits)
		}
	}
}
