package filechangedrearm

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

const note = `---
schema: 3
status: in-progress
---
## Validation loop
1. ` + "`go test ./...`" + `  → all packages ok  · re-armed by: any .go edit under tools/
   last run: pass
2. ` + "`qlty check`" + `  → no issues  · re-armed by: .qlty/qlty.toml
   last run: pass
3. ` + "`make release`" + `  → tagged  · re-armed by: a human deciding to ship
   last run: not yet run
`

var noon = time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)

// project builds a tree whose paths the resolver will actually find, because
// WatchTargets refuses to register anything that does not exist.
func project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{
		filepath.Join(".claude", "checkpoints"),
		filepath.Join("tools", "internal"),
		".qlty",
	} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), note)
	write(t, filepath.Join(dir, ".qlty", "qlty.toml"), "config_version = \"0\"\n")
	write(t, filepath.Join(dir, "tools", "internal", "x.go"), "package internal\n")
	return dir
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fire(t *testing.T, dir, filePath, event string) int {
	t.Helper()
	in, _ := json.Marshal(hookInput{FilePath: filePath, Event: event, SessionID: "s1"})
	var o, e bytes.Buffer
	return run(nil, bytes.NewReader(in), &o, &e, dir, noon)
}

func stateOf(t *testing.T, dir string) checkpoint.RearmState {
	t.Helper()
	s, err := checkpoint.LoadRearm(os.ReadFile, checkpoint.RearmPath(dir))
	if err != nil {
		t.Fatalf("LoadRearm: %v", err)
	}
	return s
}

// The core claim: a change under a check's trigger surface marks THAT check
// stale, and names the file that did it.
func TestChangeUnderASurfaceRearmsItsCheck(t *testing.T) {
	dir := project(t)
	if code := fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change"); code != 0 {
		t.Fatalf("exit %d — a FileChanged hook must never cost a tool call", code)
	}
	got := stateOf(t, dir).Sorted()
	if len(got) != 1 {
		t.Fatalf("rearmed = %+v, want exactly check 1", got)
	}
	if got[0].Index != 1 {
		t.Errorf("re-armed check %d, want 1 (the one claiming tools/)", got[0].Index)
	}
	if !strings.Contains(got[0].Check, "go test") {
		t.Errorf("check not cited verbatim: %q", got[0].Check)
	}
	if got[0].By != "tools/internal/x.go" {
		t.Errorf("by = %q, want the project-relative path", got[0].By)
	}
	if got[0].Event != "change" || got[0].At != "2026-07-29T12:00:00Z" {
		t.Errorf("event/at wrong: %+v", got[0])
	}
}

// A file watch on a literal file is as valid as a directory watch.
func TestALiteralFileSurfaceRearms(t *testing.T) {
	dir := project(t)
	fire(t, dir, filepath.Join(dir, ".qlty", "qlty.toml"), "change")
	got := stateOf(t, dir).Sorted()
	if len(got) != 1 || got[0].Index != 2 {
		t.Fatalf("rearmed = %+v, want check 2", got)
	}
}

// A directory watch is COARSER than the surface it stands in for, so unclaimed
// changes are expected. Recording them would turn a signal into a log.
func TestUnclaimedChangesAreNotRecorded(t *testing.T) {
	dir := project(t)
	write(t, filepath.Join(dir, "README.md"), "hello\n")
	fire(t, dir, filepath.Join(dir, "README.md"), "change")
	if got := stateOf(t, dir).Sorted(); len(got) != 0 {
		t.Errorf("recorded an unattributable change: %+v", got)
	}
}

// A surface expressed as prose has no path. The check simply never re-arms from
// the watcher — and the RESTORE side is what reports that gap, not this hook.
func TestProseOnlySurfaceNeverRearms(t *testing.T) {
	dir := project(t)
	write(t, filepath.Join(dir, "anything.txt"), "x\n")
	fire(t, dir, filepath.Join(dir, "anything.txt"), "change")
	for _, r := range stateOf(t, dir).Sorted() {
		if r.Index == 3 {
			t.Errorf("check 3 has no resolvable surface but was re-armed: %+v", r)
		}
	}
}

// THE self-trigger guard. Watching the project root was measured to catch the
// hooks' own writes and re-fire ten times; this closes the loop by rule so a
// note naming .claude as a surface cannot start one.
func TestChangesUnderDotClaudeAreIgnored(t *testing.T) {
	dir := project(t)
	for _, p := range []string{
		filepath.Join(dir, ".claude", "checkpoints", "rearmed.json"),
		filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		filepath.Join(dir, ".claude", "settings.json"),
	} {
		if code := fire(t, dir, p, "change"); code != 0 {
			t.Fatalf("exit %d", code)
		}
	}
	if _, err := os.Stat(checkpoint.RearmPath(dir)); err == nil {
		t.Error("wrote state in response to a change under .claude/ — that is the self-trigger loop")
	}
}

// The exclusion is `rel == ".claude" || strings.HasPrefix(rel, ".claude/")`, and it is
// only LOAD-BEARING when a watch surface actually claims a path under .claude/ —
// otherwise nothing is attributed and no state is written whether the guard fires or
// not. That is why the test above cannot pin it: its note watches `tools/` and
// `.qlty/`, so a mutation sweep disarmed the `||` and the suite stayed green.
//
// This note names `.claude/checkpoints/` as a trigger surface, which is the whole
// scenario: the hook writes rearmed.json INTO that directory, so without the exclusion
// its own write re-triggers it, attributes it to the check watching that directory, and
// re-arms forever. Both halves of the condition get their own firing.
func TestDotClaudeExclusionIsLoadBearing(t *testing.T) {
	selfWatching := "---\nschema: 3\nstatus: in-progress\n---\n" +
		"## Validation loop\n" +
		"1. `go test ./...`  → all ok  · re-armed by: .claude/checkpoints/\n" +
		"   last run: pass\n" +
		"2. `qlty check`  → no issues  · re-armed by: .claude\n" +
		"   last run: pass\n"

	for _, changed := range []string{
		filepath.Join(".claude", "checkpoints", "rearmed.json"), // the hook's OWN output
		filepath.Join(".claude", "checkpoints", "CHECKPOINT.md"),
		".claude", // the bare directory — the `rel == ".claude"` half
	} {
		t.Run(changed, func(t *testing.T) {
			dir := project(t)
			write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), selfWatching)

			if code := fire(t, dir, filepath.Join(dir, changed), "change"); code != 0 {
				t.Fatalf("exit %d", code)
			}
			// Sanity that the surface really does claim this path — otherwise this
			// test would pass for the same wrong reason as its predecessor.
			within, closeRoot := checkpoint.RootedWithin(dir)
			defer func() { _ = closeRoot() }()
			claimed, _ := checkpoint.WatchTargets(".claude/checkpoints/", within)
			if len(claimed) == 0 {
				t.Fatal("fixture is wrong: .claude/checkpoints/ did not resolve to a watch target, so this test cannot observe the exclusion")
			}

			if _, err := os.Stat(checkpoint.RearmPath(dir)); err == nil {
				t.Error("a change under .claude/ was recorded even though a check watches it — that is the self-trigger loop, and it runs forever")
			}
		})
	}

	// The other direction: a path that merely SHARES the prefix is ordinary watched
	// code and must still re-arm. A too-eager exclusion is a silent hole in the
	// re-arm, which is the failure the whole FileChanged path exists to prevent.
	dir := project(t)
	sibling := filepath.Join(dir, "tools", "internal", ".claudefiles.go")
	write(t, sibling, "package internal\n")
	if code := fire(t, dir, sibling, "change"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if _, err := os.Stat(checkpoint.RearmPath(dir)); err != nil {
		t.Errorf("a file merely sharing the .claude prefix was excluded: %v", err)
	}
}

// relTo is the display path, and its fallback matters: a file OUTSIDE the project must
// be shown as what it is, not as a relative path climbing out of the tree. A mutation
// sweep turned `err == nil && !HasPrefix(r, "..")` into an `||`, which reports "../x"
// as though it were a project path — a re-arm record citing a path that does not exist
// relative to the project it names.
func TestRelToFallsBackForPathsOutsideTheProject(t *testing.T) {
	dir := t.TempDir()

	if got := relTo(dir, filepath.Join(dir, "tools", "x.go")); got != "tools/x.go" {
		t.Errorf("an in-project path must be project-relative and slash-separated, got %q", got)
	}
	outside := filepath.Join(filepath.Dir(dir), "sibling", "x.go")
	got := relTo(dir, outside)
	if strings.HasPrefix(got, "..") {
		t.Errorf("relTo(%q) = %q — a path outside the project must not be reported as a relative climb", outside, got)
	}
	if got != filepath.ToSlash(outside) {
		t.Errorf("relTo(%q) = %q, want the path itself", outside, got)
	}
}

// An empty file_path is a payload with nothing to attribute. The hook must no-op
// rather than reason about "".
func TestEmptyFilePathIsANoOp(t *testing.T) {
	dir := project(t)
	if code := fire(t, dir, "", "change"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if _, err := os.Stat(checkpoint.RearmPath(dir)); err == nil {
		t.Error("wrote re-arm state for a payload carrying no file path")
	}
}

// Newest wins, and the file does not grow without bound: a check re-armed twice
// is re-armed. The count is not the signal; the fact is.
func TestRepeatedRearmsCollapseToTheNewest(t *testing.T) {
	dir := project(t)
	fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change")
	write(t, filepath.Join(dir, "tools", "internal", "y.go"), "package internal\n")
	fire(t, dir, filepath.Join(dir, "tools", "internal", "y.go"), "add")
	got := stateOf(t, dir).Sorted()
	if len(got) != 1 {
		t.Fatalf("rearmed = %d entries, want 1 per check", len(got))
	}
	if got[0].By != "tools/internal/y.go" || got[0].Event != "add" {
		t.Errorf("newest did not win: %+v", got[0])
	}
}

// The more specific claim wins: an author naming tools/internal meant that, not
// the parent a sibling check happens to name.
func TestLongestMatchingSurfaceWins(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints"), 0o755)
	os.MkdirAll(filepath.Join(dir, "tools", "internal"), 0o755)
	write(t, filepath.Join(dir, "tools", "internal", "x.go"), "package internal\n")
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		"## Validation loop\n1. broad  · re-armed by: tools/\n2. narrow  · re-armed by: tools/internal/\n")
	fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change")
	got := stateOf(t, dir).Sorted()
	if len(got) != 1 || got[0].Index != 2 {
		t.Fatalf("rearmed = %+v, want only the narrower check 2", got)
	}
}

// Every path exits 0 and writes nothing to stdout: this hook rides on every file
// change in a watched tree, so cost and noise are the whole risk.
func TestNeverBlocksAndNeverSpeaks(t *testing.T) {
	dir := project(t)
	cases := []struct{ name, projectDir, stdin string }{
		{"no project dir", "", `{"file_path":"/x","event":"change"}`},
		{"no file_path", dir, `{"event":"change"}`},
		{"malformed stdin", dir, `{not json`},
		{"empty stdin", dir, ``},
		{"no note", t.TempDir(), `{"file_path":"/x/y.go","event":"change"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var o, e bytes.Buffer
			code := run(nil, strings.NewReader(c.stdin), &o, &e, c.projectDir, noon)
			if code != 0 {
				t.Errorf("exit %d, want 0", code)
			}
			if o.String() != "" {
				t.Errorf("wrote to stdout: %q", o.String())
			}
		})
	}
}

func TestVersion(t *testing.T) {
	var o, e bytes.Buffer
	if code := run([]string{"-version"}, strings.NewReader(""), &o, &e, t.TempDir(), noon); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(o.String(), "sc-filechanged-rearm") {
		t.Errorf("version = %q", o.String())
	}
}

// #217. A check whose COMMAND is rewritten gets a new identity, so its old record can never
// match again — it just sits in the file forever, and every consumer that counts records
// (sc-doctor, the SessionStart summary) reads a stale check as still re-armed. The record
// must go when its check does.
//
// NARROWED BY #432: this used to reword the ANNOTATION ("all packages ok" -> "every package
// ok") and expect a prune. That is no longer an orphan and should never have been treated as
// one — the command did not move, so the check did not. What is exercised here is a real
// command change; the annotation case is TestRewordingAnAnnotationKeepsTheRecord.
func TestRecordsAreDroppedWhenTheirCheckIsRewritten(t *testing.T) {
	dir := project(t)
	if code := fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change"); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	before := stateOf(t, dir)
	if len(before.Rearmed) != 1 {
		t.Fatalf("setup: want one record, got %v", before.Rearmed)
	}

	// The human rewords check 1. Same check, new identity.
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		strings.Replace(note, "`go test ./...`", "`go test -race ./...`", 1))

	// An UNMATCHED event, deliberately: the prune must not wait for a change that happens
	// to re-arm something, or the orphan outlives the note that explained it.
	write(t, filepath.Join(dir, "unclaimed.txt"), "x")
	if code := fire(t, dir, filepath.Join(dir, "unclaimed.txt"), "change"); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if got := stateOf(t, dir).Rearmed; len(got) != 0 {
		t.Errorf("orphan survived the rewrite: %v", got)
	}

	// And the message names the cause, because a record vanishing is otherwise indis-
	// tinguishable from the hook never having recorded it.
	dir2 := project(t)
	fire(t, dir2, filepath.Join(dir2, "tools", "internal", "x.go"), "change")
	write(t, filepath.Join(dir2, ".claude", "checkpoints", "CHECKPOINT.md"),
		strings.Replace(note, "`go test ./...`", "`go test -race ./...`", 1))
	in, _ := json.Marshal(hookInput{FilePath: filepath.Join(dir2, ".qlty", "qlty.toml"), Event: "change"})
	var o, e bytes.Buffer
	run(nil, bytes.NewReader(in), &o, &e, dir2, noon)
	if !strings.Contains(e.String(), "no longer appears in the note") {
		t.Errorf("stderr = %q; a silent deletion is the failure mode this replaces", e.String())
	}
	// The event's own re-arm still landed — pruning is not a stop-the-world.
	if got := stateOf(t, dir2).Rearmed; len(got) != 1 {
		t.Errorf("state = %v; want only the check this event re-armed", got)
	}
}

// A note that stops parsing must not be read as "no checks exist". The hook returns before
// the update in that case, so nothing is pruned AND nothing is recorded.
func TestAnUnparsableNoteDeletesNothing(t *testing.T) {
	dir := project(t)
	fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change")
	if len(stateOf(t, dir).Rearmed) != 1 {
		t.Fatal("setup: want one record")
	}

	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), "---\nschema: 3\n---\n## Notes\nnothing here\n")
	fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change")

	if got := stateOf(t, dir).Rearmed; len(got) != 1 {
		t.Errorf("state = %v; a note we could not parse is not evidence that its checks are gone", got)
	}
}

// #219. This hook is where a malformed entry actually COSTS something: the entry opens no
// check, so nothing claims the files under its surface, so they are never re-armed. It used
// to say nothing at all — the complaint reached the operator once, at the next session start,
// from a different hook.
func TestAnUnmatchedChangeReportsWhatTheParserDropped(t *testing.T) {
	dir := project(t)
	// Entry 2b. names .qlty/ and opens no check, so a .qlty change re-arms nothing.
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), `---
schema: 3
---
## Validation loop
1. `+"`go test ./...`"+`  → ok  · re-armed by: any .go edit under tools/
   last run: pass
2b. `+"`qlty check`"+`  → no issues  · re-armed by: .qlty/qlty.toml
   last run: pass
## Next intended steps
1. one
2. two
`)

	in, _ := json.Marshal(hookInput{FilePath: filepath.Join(dir, ".qlty", "qlty.toml"), Event: "change"})
	var o, e bytes.Buffer
	if code := run(nil, bytes.NewReader(in), &o, &e, dir, noon); code != 0 {
		t.Fatalf("exit = %d; this hook must never block a tool call", code)
	}
	if !strings.Contains(e.String(), "2b.") || !strings.Contains(e.String(), "no check claimed it") {
		t.Errorf("stderr = %q; want the unclaimed file AND the entry that would have claimed it", e.String())
	}
	if o.String() != "" {
		t.Errorf("stdout = %q; this hook never writes to stdout", o.String())
	}
	// The numbered next-steps must not be read as loop entries.
	if strings.Contains(e.String(), "position 3") {
		t.Errorf("a section other than the loop was parsed: %q", e.String())
	}
}

// A change that DID re-arm a check is not the moment to complain: this hook fires once per
// file, and an unconditional warning would repeat dozens of times a turn and be tuned out.
func TestAMatchedChangeStaysQuietEvenWithAMalformedLoop(t *testing.T) {
	dir := project(t)
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		strings.Replace(note, "\n2. ", "\n2b. ", 1))

	in, _ := json.Marshal(hookInput{FilePath: filepath.Join(dir, "tools", "internal", "x.go"), Event: "change"})
	var o, e bytes.Buffer
	run(nil, bytes.NewReader(in), &o, &e, dir, noon)

	if e.String() != "" {
		t.Errorf("stderr = %q; a change that re-armed its check has nothing to answer for", e.String())
	}
	if len(stateOf(t, dir).Rearmed) != 1 {
		t.Error("the re-arm itself must still be recorded")
	}
}

// The severe case: a loop a reader counts and the parser opens NOTHING from. Every file
// change re-arms nothing, all session. The hook returned here before saying a word.
func TestALoopThatOpensNoChecksSaysSo(t *testing.T) {
	dir := project(t)
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), `---
schema: 3
---
## Validation loop
1a. `+"`go test ./...`"+`  → ok  · re-armed by: any .go edit under tools/
1b. `+"`qlty check`"+`  → no issues  · re-armed by: .qlty/qlty.toml
`)

	in, _ := json.Marshal(hookInput{FilePath: filepath.Join(dir, "tools", "internal", "x.go"), Event: "change"})
	var o, e bytes.Buffer
	if code := run(nil, bytes.NewReader(in), &o, &e, dir, noon); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(e.String(), "opens NO checks") {
		t.Errorf("stderr = %q; a loop that watches nothing must not be silent", e.String())
	}
	if !strings.Contains(e.String(), "1a.") || !strings.Contains(e.String(), "1b.") {
		t.Errorf("stderr = %q; name the entries, or there is nothing to fix", e.String())
	}
}

// A note with no problems and no match is silent — the common case, and the one that
// decides whether anybody still reads this channel.
func TestAnUnmatchedChangeUnderAHealthyLoopIsSilent(t *testing.T) {
	dir := project(t)
	write(t, filepath.Join(dir, "unclaimed.txt"), "x")
	in, _ := json.Marshal(hookInput{FilePath: filepath.Join(dir, "unclaimed.txt"), Event: "change"})
	var o, e bytes.Buffer
	run(nil, bytes.NewReader(in), &o, &e, dir, noon)
	if e.String() != "" {
		t.Errorf("stderr = %q; an unclaimed change under a healthy loop is expected, not a fault", e.String())
	}
}

// #432 friction 2 — THE CASE THE OLD KEY GOT WRONG FOUR TIMES OUT OF FIVE.
//
// Correcting a stale count is not a change to the check. Under the whole-line key it was, so
// the live state grew to fourteen records for twelve checks, and the note's authors learned
// to leave wrong numbers in place rather than orphan a record. A check's `last run` must
// survive its annotation being fixed.
func TestRewordingAnAnnotationKeepsTheRecord(t *testing.T) {
	dir := project(t)
	if code := fire(t, dir, filepath.Join(dir, "tools", "internal", "x.go"), "change"); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	before := stateOf(t, dir).Rearmed
	if len(before) != 1 {
		t.Fatalf("setup: want one record, got %v", before)
	}

	// Same command, corrected count — exactly the edit the loop had been avoiding.
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		strings.Replace(note, "→ all packages ok", "→ 30 packages ok", 1))

	write(t, filepath.Join(dir, "unclaimed.txt"), "x")
	if code := fire(t, dir, filepath.Join(dir, "unclaimed.txt"), "change"); code != 0 {
		t.Fatalf("exit = %d", code)
	}
	after := stateOf(t, dir).Rearmed
	if len(after) != 1 {
		t.Fatalf("state = %v; correcting an annotation must not orphan the record", after)
	}
	for k, r := range after {
		if r.At != recordAt(before) {
			t.Errorf("record %q was re-created rather than kept: at=%q want=%q", k, r.At, recordAt(before))
		}
	}
}

func recordAt(m map[string]checkpoint.Rearm) string {
	for _, r := range m {
		return r.At
	}
	return ""
}
