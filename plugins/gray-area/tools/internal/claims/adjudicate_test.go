package claims

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/trajectory"
)

// bash builds one transcript line carrying a Bash call.
func bash(uuid, ts, command string) string {
	return line(uuid, ts, "Bash", map[string]string{"command": command})
}

// write builds one transcript line carrying a Write call.
func write(uuid, ts, path string) string {
	return line(uuid, ts, "Write", map[string]string{"file_path": path, "content": "x"})
}

func line(uuid, ts, tool string, input map[string]string) string {
	in, _ := json.Marshal(input)
	blk, _ := json.Marshal([]map[string]any{{
		"type": "tool_use", "name": tool, "id": "toolu_" + uuid, "input": json.RawMessage(in),
	}})
	row, _ := json.Marshal(map[string]any{
		"uuid": uuid, "timestamp": ts, "type": "assistant",
		"message": map[string]any{"content": json.RawMessage(blk)},
	})
	return string(row)
}

func trace(t *testing.T, lines ...string) trajectory.Result {
	t.Helper()
	res, err := trajectory.Parse(strings.NewReader(strings.Join(lines, "\n")+"\n"), "session.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if res.BadLines > 0 {
		t.Fatalf("fixture did not parse cleanly: %d bad lines at %v", res.BadLines, res.BadSample)
	}
	return res
}

func one(t *testing.T, note string, res trajectory.Result) Finding {
	t.Helper()
	cs, err := Parse(strings.NewReader(note), "CHECKPOINT.md")
	if err != nil {
		t.Fatal(err)
	}
	fs := Adjudicate(cs, res)
	if len(fs) != 1 {
		t.Fatalf("got %d findings, want 1", len(fs))
	}
	return fs[0]
}

// A claim whose command IS in the trajectory cites the event that backs it.
func TestAClaimThatRanIsCitedWithItsUUID(t *testing.T) {
	note := "## Validation loop\n1. `go test -C plugins/gray-area/tools ./...` · re-armed by: plugins/gray-area/tools/\n   last run: pass 2026-07-30T05:00Z\n"
	res := trace(t,
		bash("aaa", "2026-07-30T04:50Z", "cd /repo && go test -C plugins/gray-area/tools ./... 2>&1 | grep FAIL"),
		bash("bbb", "2026-07-30T04:55Z", "git status"),
	)
	f := one(t, note, res)
	if f.Verdict != Cited {
		t.Fatalf("verdict = %s (%s)", f.Verdict, f.Note)
	}
	if f.UUID != "aaa" {
		t.Errorf("cited uuid = %q, want the matching Bash call", f.UUID)
	}
	// The piped, cd-prefixed spelling is the whole reason matching is by token
	// containment; an equality matcher would have missed this and accused the note.
	if !strings.Contains(f.Command, "go test") {
		t.Errorf("the cited command is not the one that matched: %q", f.Command)
	}
	if f.NoteLine == 0 {
		t.Error("no note provenance — the claim's own side must be citable too")
	}
}

// THE DELICATE ONE. A miss must report what it searched, or "not found" is
// indistinguishable from "did not look" — and a reader cannot tell whether the
// note lied or the matcher did.
func TestAMissReportsWhatItSearchedForAndHowMuch(t *testing.T) {
	note := "## Validation loop\n1. `node scripts/rule-sweep.mjs --selftest` · re-armed by: scripts/\n   last run: pass 2026-07-30T05:00Z\n"
	res := trace(t,
		bash("aaa", "2026-07-30T04:50Z", "go build ./..."),
		bash("bbb", "2026-07-30T04:51Z", "git log"),
	)
	f := one(t, note, res)
	if f.Verdict != NoEvidence {
		t.Fatalf("verdict = %s, want NO-EVIDENCE", f.Verdict)
	}
	if len(f.Searched) == 0 {
		t.Error("the miss does not say what it searched for")
	}
	if f.EventsSeen != 2 {
		t.Errorf("events_seen = %d, want 2 — a miss must say how much was searched", f.EventsSeen)
	}
	if f.UUID != "" {
		t.Error("a miss cited an event")
	}
	// The wording carries the epistemics. This is not decoration: the verdict is
	// read by a human deciding whether to distrust a note.
	if !strings.Contains(f.Note, "ABSENCE OF EVIDENCE") {
		t.Errorf("the miss reads as a finding of fact rather than an absence: %q", f.Note)
	}
}

// THE #165 SHAPE, which is why this phase exists. A claim whose trigger surface
// was written to AFTER the claimed run time is stale — computed from the
// trajectory alone, needing nothing from the FileChanged mechanism whose
// coverage that issue records as unreliable.
func TestAClaimIsStaleWhenItsTriggerSurfaceMovedAfterwards(t *testing.T) {
	note := "## Validation loop\n1. `go test -C plugins/gray-area/tools ./...` · re-armed by: plugins/gray-area/tools/\n   last run: pass 2026-07-30T04:00Z\n"
	res := trace(t,
		bash("ran", "2026-07-30T03:59Z", "go test -C plugins/gray-area/tools ./..."),
		write("edit", "2026-07-30T04:30Z", "/home/u/repo/plugins/gray-area/tools/internal/claims/claims.go"),
	)
	f := one(t, note, res)
	if f.Verdict != Stale {
		t.Fatalf("verdict = %s, want STALE (%s)", f.Verdict, f.Note)
	}
	if f.UUID != "edit" {
		t.Errorf("STALE cited %q, want the write that invalidated the claim", f.UUID)
	}
	if !strings.Contains(f.Note, "plugins/gray-area/tools/") {
		t.Errorf("the finding does not name the surface that moved: %q", f.Note)
	}
}

// A write BEFORE the claimed run does not make it stale — the check ran after the
// change, which is the loop working. Getting this backwards would flag every
// healthy claim and make the verdict worthless.
func TestAWriteBeforeTheRunDoesNotMakeItStale(t *testing.T) {
	note := "## Validation loop\n1. `go test -C plugins/gray-area/tools ./...` · re-armed by: plugins/gray-area/tools/\n   last run: pass 2026-07-30T05:00Z\n"
	res := trace(t,
		write("edit", "2026-07-30T04:30Z", "/repo/plugins/gray-area/tools/internal/claims/claims.go"),
		bash("ran", "2026-07-30T04:59Z", "go test -C plugins/gray-area/tools ./..."),
	)
	if f := one(t, note, res); f.Verdict != Cited {
		t.Errorf("verdict = %s, want CITED — the check ran after the edit", f.Verdict)
	}
}

// A write OUTSIDE the surface is irrelevant to this claim. Without this, one
// unrelated edit would stale every check in the note.
func TestAWriteOutsideTheSurfaceIsNotStaleness(t *testing.T) {
	note := "## Validation loop\n1. `go test -C plugins/gray-area/tools ./...` · re-armed by: plugins/gray-area/tools/\n   last run: pass 2026-07-30T04:00Z\n"
	res := trace(t,
		bash("ran", "2026-07-30T03:59Z", "go test -C plugins/gray-area/tools ./..."),
		write("edit", "2026-07-30T04:30Z", "/repo/plugins/frank-exchange-of-views/tools/x.go"),
	)
	if f := one(t, note, res); f.Verdict != Cited {
		t.Errorf("verdict = %s — an edit in another plugin staled this claim", f.Verdict)
	}
}

// A prose claim is UNCHECKABLE, not NO-EVIDENCE. The two say different things: one
// is "there was nothing to look for", the other is "I looked and found nothing",
// and reporting the first as the second accuses a note of a gap it never had.
func TestAProseClaimIsUncheckableRatherThanAMiss(t *testing.T) {
	note := "## Validation loop\n1. The continuity loop itself, end to end · re-armed by: .claude/settings.local.json\n   last run: PASS 2026-07-30T03:23Z\n"
	f := one(t, note, trace(t, bash("a", "2026-07-30T04:00Z", "ls")))
	if f.Verdict != Uncheckable {
		t.Fatalf("verdict = %s, want UNCHECKABLE", f.Verdict)
	}
	if len(f.Searched) != 0 {
		t.Errorf("it claims to have searched for %v, but there was no command", f.Searched)
	}
}

// An event with no provenance cannot back a finding, so it must not be counted as
// having been searched either — the count is part of the claim the tool makes
// about its own thoroughness.
func TestUncitableEventsAreNotCountedAsSearched(t *testing.T) {
	// A line with no uuid parses into an event that Cited() rejects.
	noUUID, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-30T04:00Z", "type": "assistant",
		"message": map[string]any{"content": []map[string]any{{
			"type": "tool_use", "name": "Bash", "id": "t1", "input": map[string]string{"command": "ls"},
		}}},
	})
	res := trace(t, bash("aaa", "2026-07-30T04:00Z", "ls"), string(noUUID))
	note := "## Validation loop\n1. `node scripts/rule-sweep.mjs --selftest` · re-armed by: scripts/\n   last run: pass 2026-07-30T05:00Z\n"
	f := one(t, note, res)
	if f.EventsSeen != 1 {
		t.Errorf("events_seen = %d, want 1 — an uncitable event was counted as searched", f.EventsSeen)
	}
}

// Order is the note's order, so a reader can follow the two documents side by
// side. Findings sorted by verdict would silently renumber the loop.
func TestFindingsComeBackInTheNotesOwnOrder(t *testing.T) {
	note := "## Validation loop\n" +
		"1. `node scripts/rule-sweep.mjs --selftest` · re-armed by: scripts/\n   last run: pass 2026-07-30T05:00Z\n" +
		"2. `go test ./...` · re-armed by: tools/\n   last run: pass 2026-07-30T05:00Z\n"
	cs, err := Parse(strings.NewReader(note), "CHECKPOINT.md")
	if err != nil {
		t.Fatal(err)
	}
	fs := Adjudicate(cs, trace(t, bash("a", "2026-07-30T04:00Z", "go test ./...")))
	if len(fs) != 2 || fs[0].Index != "1" || fs[1].Index != "2" {
		t.Fatalf("order = %+v", fs)
	}
	if fs[0].Verdict != NoEvidence || fs[1].Verdict != Cited {
		t.Errorf("verdicts = %s, %s", fs[0].Verdict, fs[1].Verdict)
	}
}

// A surface that is not path-shaped is not treated as a path — the same rule the
// restore hook applies. `scripts` is a word; `scripts/` is a place.
func TestABareWordSurfaceIsNotTreatedAsAPath(t *testing.T) {
	if underSurface("/repo/plugins/scripts/x.go", "scripts") {
		t.Error("a bare word was matched as a directory")
	}
	if !underSurface("/repo/scripts/x.mjs", "scripts/") {
		t.Error("a trailing-slash surface did not match a path under it")
	}
	if !underSurface("/repo/.claude/settings.local.json", ".claude/settings.local.json") {
		t.Error("a dot-prefixed file surface did not match itself")
	}
}

// A heredoc BODY is data, not command. Measured on this repo's own transcript:
// a `python3 - <<PY` script that merely mentioned plugins/prosthetic-conscience/tools
// was reported as CITED evidence that the tests there had been run.
//
// A false citation is worse than a miss: a miss shows its search and invites a
// check, while a wrong citation looks settled.
func TestAHeredocBodyIsNotEvidenceOfACommand(t *testing.T) {
	note := "## Validation loop\n1. `go test -C plugins/prosthetic-conscience/tools ./...` · re-armed by: nowhere/\n   last run: pass 2026-07-30T05:00Z\n"
	mention := "python3 - <<'PY'\np='cmd/x.go'\n# go test -C plugins/prosthetic-conscience/tools ./...\nPY\ngofmt -w cmd/"
	f := one(t, note, trace(t, bash("heredoc", "2026-07-30T04:00Z", mention)))
	if f.Verdict == Cited {
		t.Fatalf("a heredoc that only MENTIONS the path was cited as having run it:\n  %s", f.Command)
	}
	if f.Verdict != NoEvidence {
		t.Errorf("verdict = %s, want NO-EVIDENCE", f.Verdict)
	}
}

// ...but the command lines AROUND a heredoc are still command, so stripping must
// not throw away the real invocation that follows one.
func TestCommandsAroundAHeredocStillMatch(t *testing.T) {
	note := "## Validation loop\n1. `go test ./...` · re-armed by: nowhere/\n   last run: pass 2026-07-30T05:00Z\n"
	real := "python3 - <<'PY'\nprint('irrelevant')\nPY\ngo test ./... -count=1"
	if f := one(t, note, trace(t, bash("both", "2026-07-30T04:00Z", real))); f.Verdict != Cited {
		t.Errorf("verdict = %s — stripping the heredoc also removed the real command that followed it", f.Verdict)
	}
}

// An unterminated heredoc must not leak its body back into the match. The shell
// would feed the rest to the reader, so none of it is command.
func TestAnUnterminatedHeredocDoesNotLeak(t *testing.T) {
	note := "## Validation loop\n1. `go test ./...` · re-armed by: nowhere/\n   last run: pass 2026-07-30T05:00Z\n"
	leaky := "cat <<'PY'\ngo test ./...\n"
	if f := one(t, note, trace(t, bash("leak", "2026-07-30T04:00Z", leaky))); f.Verdict == Cited {
		t.Errorf("an unterminated heredoc's body was matched as a command: %q", f.Command)
	}
}

// The evidence shown must be the evidence matched on. A compound command that
// edits a file and then runs the check displayed as its FIRST line — a correct
// citation presented as an obviously wrong one, which costs the trust the
// provenance exists to build.
func TestTheCitedLineIsTheLineThatMatched(t *testing.T) {
	note := "## Validation loop\n1. `go test ./internal/difftest` · re-armed by: nowhere/\n   last run: pass 2026-07-30T05:00Z\n"
	compound := "python3 - <<'PY'\nedit_a_file()\nPY\ngofmt -w .\ngo test ./internal/difftest -run TestGolden"
	f := one(t, note, trace(t, bash("c", "2026-07-30T04:00Z", compound)))
	if f.Verdict != Cited {
		t.Fatalf("verdict = %s", f.Verdict)
	}
	if !strings.HasPrefix(f.Command, "go test") {
		t.Errorf("cited line = %q, want the line carrying the match", f.Command)
	}
}

// THE ACT/RESULT BOUNDARY, and it lives in Adjudicate so every caller inherits it.
//
// A validation-loop entry asserts an act AND an outcome. This record can check the
// act and cannot see the outcome — the trajectory holds invocations, never their
// output. A CITED row silent about the outcome reads as endorsing the claimed pass.
//
// It shipped in the pull request reader first and was missing here:
// `adjacent-seat-omission`, caught by the sibling sweep. This test is what stops
// the third caller from omitting it again.
func TestACitedClaimNamesTheOutcomeItCouldNotCheck(t *testing.T) {
	res := trace(t, `{"type":"assistant","uuid":"u1","timestamp":"2026-08-18T01:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go run ./check"}}]}}`)
	f := Adjudicate([]Claim{{
		File: "/w/n.md", Line: 3, Cmd: "go run ./check", Outcome: "pass 08-18T01:00Z",
	}}, res)
	if len(f) != 1 || f[0].Verdict != Cited {
		t.Fatalf("want one CITED, got %+v", f)
	}
	if len(f[0].Unmeasured) == 0 {
		t.Fatal("CITED with no boundary — the row reads as endorsing the claimed pass")
	}
	if !strings.Contains(f[0].Unmeasured[0], "pass 08-18T01:00Z") {
		t.Errorf("the boundary does not name the claimed outcome: %q", f[0].Unmeasured[0])
	}
	if !strings.Contains(f[0].Unmeasured[0], "not their output") {
		t.Errorf("the boundary does not say WHY it could not be checked: %q", f[0].Unmeasured[0])
	}
}

// A claim asserting no outcome must NOT gain a caveat. Everywhere is nowhere: a
// marker on every row is noise, and noise is how a real one gets skipped.
func TestAClaimWithNoAssertedOutcomeGetsNoBoundaryNote(t *testing.T) {
	res := trace(t, `{"type":"assistant","uuid":"u1","timestamp":"2026-08-18T01:00:00Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go run ./check"}}]}}`)
	f := Adjudicate([]Claim{{File: "/w/n.md", Line: 3, Cmd: "go run ./check"}}, res)
	if len(f[0].Unmeasured) != 0 {
		t.Errorf("invented a boundary note for a claim that asserted no outcome: %v", f[0].Unmeasured)
	}
}
