package checkpointseal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// tree builds a project whose directories actually exist, because WatchTargets refuses to
// register a surface it cannot stat — containment is os.Root, not string arithmetic.
func tree(t *testing.T, dirs ...string) (string, checkpoint.WithinRoot) {
	t.Helper()
	dir := t.TempDir()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	within, closeRoot := checkpoint.RootedWithin(dir)
	t.Cleanup(func() { _ = closeRoot() })
	return dir, within
}

func noteWith(loop string) string {
	return "---\nschema: 2\nobjective: something\n---\n## Validation loop\n" + loop + "\n## Open threads\n- none\n"
}

// THE #166 SCENARIO, as measured. The inherited note's loop watched
// plugins/prosthetic-conscience/tools/ while every edit in the window landed in
// plugins/frank-exchange-of-views/tools/ — so the checks reported `last run: pass` about a
// tree nobody had touched, and no check covered the tree that moved. 37 minutes, an
// auto-compaction inside the window, and every CI gate green throughout.
func TestDriftCatchesTheLoopPointedAtTheWrongTree(t *testing.T) {
	_, within := tree(t, "plugins/prosthetic-conscience/tools", "plugins/frank-exchange-of-views/tools")

	note := noteWith("1. `go test ./...`  → all ok  · re-armed by: plugins/prosthetic-conscience/tools/\n   last run: pass")
	written := []string{
		"plugins/frank-exchange-of-views/tools/internal/verify/verify.go",
		"plugins/frank-exchange-of-views/tools/internal/cli/verify.go",
	}

	got := drift(note, written, within)
	if got == "" {
		t.Fatal("the 03:30–04:07 case produced no line — this is the failure the check exists for")
	}
	for _, want := range []string{"NO check covers the work", "prosthetic-conscience/tools", "frank-exchange-of-views"} {
		if !strings.Contains(got, want) {
			t.Errorf("line missing %q:\n%s", want, got)
		}
	}
}

// The healthy case must stay SILENT. A check that fires on every seal is a check nobody
// reads by the third one.
func TestDriftIsSilentWhenTheLoopReachesTheWork(t *testing.T) {
	_, within := tree(t, "plugins/prosthetic-conscience/tools/internal")

	note := noteWith("1. `go test ./...`  → all ok  · re-armed by: plugins/prosthetic-conscience/tools/\n   last run: pass")
	cases := [][]string{
		{"plugins/prosthetic-conscience/tools/internal/checkpoint/checkpoint.go"},
		{"plugins/prosthetic-conscience/tools/internal"},                   // the surface's own directory
		{"plugins/prosthetic-conscience/tools/internal/x.go", "README.md"}, // one covered file is enough
	}
	for _, written := range cases {
		if got := drift(note, written, within); got != "" {
			t.Errorf("drift(%v) = %q; want silence", written, got)
		}
	}
}

// A prefix that is NOT a path boundary must not count as coverage. `tools/` watching
// `toolsmith/x.go` would be a false silence, which is worse than a false alarm here: the
// whole point is noticing that nothing covers the work.
func TestDriftRequiresAPathBoundary(t *testing.T) {
	_, within := tree(t, "tools", "toolsmith")
	note := noteWith("1. `go test`  → ok  · re-armed by: tools/\n   last run: pass")

	if got := drift(note, []string{"toolsmith/x.go"}, within); got == "" {
		t.Error("`tools/` must not be read as covering `toolsmith/` — that is a false silence")
	}
	if got := drift(note, []string{"tools/x.go"}, within); got != "" {
		t.Errorf("`tools/` must cover `tools/x.go`: %s", got)
	}
}

// A note with no loop at all, while the session wrote. This is the more basic failure than
// #166's — there, a loop existed and pointed elsewhere; here validation-loop's "BEFORE
// implementing, YOU MUST write down the validation loop" was simply not done.
func TestDriftNamesAMissingLoop(t *testing.T) {
	_, within := tree(t, "tools")
	for _, note := range []string{
		"---\nschema: 2\n---\n## Open threads\n- none\n",                  // no section
		"---\nschema: 2\n---\n## Validation loop\n\n## Open threads\n",    // empty section
		"---\nschema: 2\n---\n## Validation loop\n← the checks go here\n", // scaffolding only
	} {
		got := drift(note, []string{"tools/x.go"}, within)
		if !strings.Contains(got, "NO validation loop") {
			t.Errorf("a note without a loop must say so, got %q for note %q", got, note)
		}
		if !strings.Contains(got, "BEFORE implementing") {
			t.Errorf("the line must cite the duty it is enforcing: %q", got)
		}
	}
}

// A loop whose every trigger surface is prose ("a human deciding to ship") resolves to
// nothing, so no check can claim any file. That is distinct from "the loop points
// elsewhere" and gets its own message — the operator's next move differs.
func TestDriftNamesUnresolvableSurfaces(t *testing.T) {
	_, within := tree(t, "tools")
	note := noteWith("1. `make release`  → tagged  · re-armed by: a human deciding to ship\n   last run: not yet run")

	got := drift(note, []string{"tools/x.go"}, within)
	if !strings.Contains(got, "resolvable trigger surface") {
		t.Errorf("an unresolvable loop must be named as such, got %q", got)
	}
}

// Silence when there is no evidence, NOT a claim of health. A session may have written
// plenty through Bash, which internal/transcript cannot see.
func TestDriftSaysNothingWithoutEvidence(t *testing.T) {
	_, within := tree(t, "tools")
	note := noteWith("1. `go test`  → ok  · re-armed by: tools/\n   last run: pass")

	if got := drift(note, nil, within); got != "" {
		t.Errorf("no observed writes must produce no line, got %q", got)
	}
	if got := drift("", nil, within); got != "" {
		t.Errorf("no note and no writes must produce no line, got %q", got)
	}
}

// The line is read at a seam, often mid-compaction. It must stay one bounded line.
func TestDriftLineStaysBounded(t *testing.T) {
	_, within := tree(t, "watched")
	note := noteWith("1. `go test`  → ok  · re-armed by: watched/\n   last run: pass")

	var many []string
	for _, n := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		many = append(many, "elsewhere/"+n+".go")
	}
	got := drift(note, many, within)
	if got == "" {
		t.Fatal("expected a drift line")
	}
	if strings.Count(got, "\n") != 0 {
		t.Errorf("the drift line must be ONE line: %q", got)
	}
	if !strings.Contains(got, "and 5 more") {
		t.Errorf("a long path list must be summarised, got %q", got)
	}
}

// The seal must never fail, whatever the drift check finds — and the check must reach
// stderr on ALL THREE events, because drift is drift at any seam.
func TestDriftReachesStderrOnEveryEventAndNeverBlocks(t *testing.T) {
	for _, event := range []string{evPreCompact, evSessionEnd, evSubagentStop} {
		t.Run(event, func(t *testing.T) {
			dir := t.TempDir()
			cp := filepath.Join(dir, ".claude", "checkpoints")
			if err := os.MkdirAll(cp, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(dir, "watched"), 0o755); err != nil {
				t.Fatal(err)
			}
			note := noteWith("1. `go test`  → ok  · re-armed by: watched/\n   last run: pass")
			if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(note), 0o644); err != nil {
				t.Fatal(err)
			}

			// A transcript recording a write OUTSIDE the watched surface.
			tp := filepath.Join(dir, "t.jsonl")
			rec := `{"message":{"content":[{"type":"tool_use","name":"Write","input":{"file_path":"` +
				filepath.ToSlash(filepath.Join(dir, "elsewhere", "x.go")) + `"}}]}}` + "\n"
			if err := os.WriteFile(tp, []byte(rec), 0o644); err != nil {
				t.Fatal(err)
			}

			stdout, stderr, code := call(t,
				`{"transcript_path":"`+filepath.ToSlash(tp)+`","session_id":"s1"}`,
				dir, noon, "-event", event)
			if code != 0 {
				t.Fatalf("the seal must never block: exit %d", code)
			}
			if !strings.Contains(stderr, "NO check covers the work") {
				t.Errorf("no drift line on %s:\nstderr=%q", event, stderr)
			}
			// Steering stays PreCompact-only; the drift line must not leak to stdout,
			// where on SubagentStop it would reach a seat mid-flight as a directive.
			if strings.Contains(stdout, "check covers") {
				t.Errorf("the drift line reached stdout on %s: %q", event, stdout)
			}
		})
	}
}

// A seat's writes are the SEAT's. SubagentStop carries agent_transcript_path — a per-seat
// file distinct from the parent's — so reading the parent's transcript would blame a seat
// for every file the parent touched, and would report drift against a loop the seat never
// owned. gray-area measured this field in the Phase 0 spike; this pins that we use it.
func TestDriftPrefersTheSeatsOwnTranscript(t *testing.T) {
	if got := driftTranscript(hookInput{TranscriptPath: "parent.jsonl", AgentTranscriptPath: "seat.jsonl"}); got != "seat.jsonl" {
		t.Errorf("SubagentStop must read the seat's own transcript, got %q", got)
	}
	if got := driftTranscript(hookInput{TranscriptPath: "parent.jsonl"}); got != "parent.jsonl" {
		t.Errorf("without a per-seat path, fall back to the session's own, got %q", got)
	}
	if got := driftTranscript(hookInput{}); got != "" {
		t.Errorf("no transcript at all must stay empty, got %q", got)
	}
}

// A reader that cannot read must SAY SO at the seam. Reporting nothing would present a
// broken check as a clean bill of health.
func TestSealReportsAnUnreadableTranscript(t *testing.T) {
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass")), 0o644); err != nil {
		t.Fatal(err)
	}
	gone := filepath.ToSlash(filepath.Join(dir, "no-such-transcript.jsonl"))

	_, stderr, code := call(t, `{"transcript_path":"`+gone+`"}`, dir, noon, "-event", evPreCompact)
	if code != 0 {
		t.Fatalf("the seal must never block: exit %d", code)
	}
	if !strings.Contains(stderr, "cannot check the note against this session's work") {
		t.Errorf("an unreadable transcript must be reported, got stderr=%q", stderr)
	}
}

// #219: the seal is the loudest moment to catch a malformed label, and the
// closest to the agent that just wrote it. Before this it reported nothing —
// LoopProblems existed but was wired only into SessionStart, so the author of a
// bad note learned about it from a LATER session, if at all.
func TestTheSealReportsALoopWhoseLabelsAreNotOrdinals(t *testing.T) {
	_, within := tree(t, "tools")

	// A loop numbered from 0, with a lettered sub-entry — the exact shape this
	// repo's own note carried while two of its checks went unwatched.
	note := noteWith("0. `go test ./...`  · re-armed by: tools/\n" +
		"0b. `go vet ./...`  · re-armed by: tools/\n" +
		"1. `go build ./...`  · re-armed by: tools/")
	got := drift(note, []string{"tools/x.go"}, within)

	if got == "" {
		t.Fatal("the seal accepted a loop whose labels are not ordinals; the checks it drops are neither watched nor adjudicated")
	}
	for _, want := range []string{`"0b."`, `"0."`} {
		if !strings.Contains(got, want) {
			t.Errorf("the report does not name %s, so the author cannot find it:\n%s", want, got)
		}
	}
}

// It must stay silent on a conforming loop — a check that cries wolf is a check
// that gets ignored when it matters.
func TestTheSealIsSilentOnAConformingLoop(t *testing.T) {
	_, within := tree(t, "tools")
	note := noteWith("1. `go test ./...`  · re-armed by: tools/\n2. `go vet ./...`  · re-armed by: tools/")
	if got := drift(note, []string{"tools/x.go"}, within); got != "" {
		t.Errorf("drift = %q; want silence on a well-formed loop", got)
	}
}
