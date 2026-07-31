package checkpointrestore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

const note = `---
schema: 2
updated: 2026-07-29T04:00:00Z
session_id: sess-abc
agent_id: null
objective: "prove the restore path fires on every source"
plan: plans/context-checkpointing.md §13
beyond_plan: false
status: in-progress
---
## Validation loop
1. go test ./...  → all packages ok  · re-armed by: any .go edit
   last run: pass
## Next intended steps
1. wire hooks.json (issue #131)
## In-flight handles
- background task bg-77, PR #141
## Invariants / foot-guns
- PostCompact cannot inject; never route restore through it
## Files touched
- main.go
## Open threads
- none
`

func call(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &e, dir)
	return o.String(), e.String(), code
}

// withNote builds a project dir holding a live checkpoint at the fallback path.
func withNote(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// injected returns the additionalContext the hook emitted, or "" for silence.
func injected(t *testing.T, stdout string) string {
	t.Helper()
	if strings.TrimSpace(stdout) == "" {
		return ""
	}
	var out hookOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("stdout is not a hook response: %v\n%s", err, stdout)
	}
	if out.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Errorf("hookEventName = %q, want SessionStart",
			out.HookSpecificOutput.HookEventName)
	}
	return out.HookSpecificOutput.AdditionalContext
}

// THE regression. plans/hook-surface-spike.md §3 (2026-07-28) derived that
// SessionStart restore must no-op on source == "compact", leaving that boundary
// to PostCompact. PostCompact cannot inject, so obeying that would have removed
// restore from the ONLY boundary this design exists for — silently, with every
// other test still green. Withdrawn 2026-07-29 and guarded here.
func TestRestoreFiresOnCompact(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, code := call(t, dir, `{"source":"compact","session_id":"s1"}`)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := injected(t, stdout)
	if got == "" {
		t.Fatal("SILENT ON source=compact — the compaction boundary is the whole point of this hook")
	}
	if !strings.Contains(got, "go test ./...") {
		t.Errorf("the validation loop did not survive the boundary:\n%s", got)
	}
}

// Every source, not a list of blessed ones. A source-specific carve-out is the
// exact shape of the error that was just withdrawn, so it is tested as a class.
func TestRestoreFiresOnEverySourceThatContinuesWork(t *testing.T) {
	dir := withNote(t, note)
	for _, src := range []string{"startup", "resume", "compact", "fork", ""} {
		stdout, _, code := call(t, dir, `{"source":"`+src+`"}`)
		if code != 0 {
			t.Fatalf("source %q: exit %d", src, code)
		}
		if injected(t, stdout) == "" {
			t.Errorf("source %q: silent, want the digest", src)
		}
	}
}

// /clear is the human wiping context deliberately. Re-imposing the full digest
// fights that instruction; a pointer keeps continuity reachable without it.
// This carve-out is by INTENT, which is what the withdrawn one was not.
func TestClearGetsAPointerNotTheDigest(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, _ := call(t, dir, `{"source":"clear"}`)
	got := injected(t, stdout)
	if got == "" {
		t.Fatal("clear should still point at the note")
	}
	if strings.Contains(got, "go test ./...") {
		t.Errorf("clear re-imposed the digest the human just wiped:\n%s", got)
	}
	if !strings.Contains(got, "CHECKPOINT.md") {
		t.Errorf("pointer does not name the note:\n%s", got)
	}
}

// A note describing finished work must not be re-imposed as live state. The
// skill says to discard a completed note; a forgotten one would otherwise
// restore dead state on every session start, and a stale anchor is worse than
// no anchor because it reads as current.
func TestStatusDoneGetsAPointerNotTheDigest(t *testing.T) {
	dir := withNote(t, strings.Replace(note, "status: in-progress", "status: done", 1))
	stdout, _, _ := call(t, dir, `{"source":"startup"}`)
	got := injected(t, stdout)
	if strings.Contains(got, "go test ./...") {
		t.Errorf("restored a completed note as live state:\n%s", got)
	}
	if !strings.Contains(got, "status: done") {
		t.Errorf("pointer does not say why it is a pointer:\n%s", got)
	}
}

// Age is deliberately not a criterion — a resume days later is when the note is
// most valuable. Guarded so a future "helpful" staleness timer has to argue.
func TestAgeAloneNeverSuppressesTheDigest(t *testing.T) {
	dir := withNote(t, strings.Replace(note, "updated: 2026-07-29T04:00:00Z", "updated: 2019-01-01T00:00:00Z", 1))
	stdout, _, _ := call(t, dir, `{"source":"resume"}`)
	if got := injected(t, stdout); !strings.Contains(got, "go test ./...") {
		t.Errorf("an old but unfinished note was suppressed:\n%s", got)
	}
}

// Reinforce, never introduce. Twice measured: content a model cannot account for
// is flagged as prompt injection, and the suspicion lands in the context the
// checkpoint is supposed to anchor. So the digest must carry its provenance.
func TestDigestCarriesItsProvenance(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, _ := call(t, dir, `{"source":"resume"}`)
	got := injected(t, stdout)
	for _, want := range []string{
		"checkpoint it wrote itself", // whose state this is
		"CHECKPOINT.md",              // where it came from
		"2026-07-29T04:00:00Z",       // when it was written
		"sess-abc",                   // which session wrote it
		"has not been re-verified since",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("digest missing %q:\n%s", want, got)
		}
	}
}

// THE mechanism for the no-imperative rule, so it is not prose that decays.
//
// Measured 2026-07-29: run 1's digest closed with "verify each item against
// reality before acting on it, and re-run the validation loop rather than
// trusting its recorded result", and the resumed agent named THAT SENTENCE among
// the directives making the payload injection-shaped. Deleting it removed it
// from the agent's reason, leaving only the note's own content.
//
// The rule is scoped, not absolute: imperatives QUOTED FROM THE NOTE are the
// content worth restoring (a foot-guns section is imperatives by definition).
// What is banned is an imperative the HOOK invented — the session never
// established it, so it arrives as a directive from outside.
func TestTheHookAddsNoImperativeOfItsOwn(t *testing.T) {
	// A note with no imperative anywhere in it, so anything commanding in the
	// output can only have come from the hook.
	declarative := `---
schema: 2
updated: 2026-07-29T04:00:00Z
objective: "measure the digest's own voice"
beyond_plan: true
status: in-progress
---
## Validation loop
1. go test ./...  → all packages ok
   last run: pass
## Next intended steps
1. wire hooks.json
`
	dir := withNote(t, declarative)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)

	// Second person and bare-imperative openers are how an invented directive
	// actually reads. Quoted note content is exempt by construction here: the
	// fixture contains none.
	banned := []string{
		"you must", "you should", "verify each", "re-run the", "do not ", "never ",
		"make sure", "ensure that", "remember to", "before acting",
	}
	low := strings.ToLower(got)
	for _, b := range banned {
		if strings.Contains(low, b) {
			t.Errorf("the hook introduced an imperative of its own (%q) — measured to read as injection:\n%s", b, got)
		}
	}
	if got == "" {
		t.Fatal("fixture produced no digest, so the assertion above proved nothing")
	}
}

// The companion: imperatives that came FROM the note must survive verbatim.
// Stripping them would drop the foot-guns, which is the content most worth
// having on the far side of a seam.
func TestImperativesQuotedFromTheNoteSurvive(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	if !strings.Contains(got, "never route restore through it") {
		t.Errorf("the note's own foot-gun was stripped:\n%s", got)
	}
}

// A note that established nothing gives the session nothing. Manufacturing a
// digest from an empty note is precisely the introduce-don't-reinforce failure.
func TestEmptyNoteProducesSilence(t *testing.T) {
	skeleton := "---\nschema: 2\nstatus: in-progress\n---\n" +
		"## Validation loop\n\n## Next intended steps\n\n## In-flight handles\n"
	dir := withNote(t, skeleton)
	stdout, _, code := call(t, dir, `{"source":"compact"}`)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if got := injected(t, stdout); got != "" {
		t.Errorf("manufactured a digest from an empty note:\n%s", got)
	}
}

// Placeholder annotations from the skill's template are scaffolding, not content.
func TestTemplateScaffoldingIsNotContent(t *testing.T) {
	tmpl := "---\nschema: 2\n---\n## Validation loop\n← load-bearing\n## Next intended steps\n← each item carries its queue pointer\n"
	dir := withNote(t, tmpl)
	stdout, _, _ := call(t, dir, `{"source":"resume"}`)
	if got := injected(t, stdout); got != "" {
		t.Errorf("restored the template's own annotations as state:\n%s", got)
	}
}

// beyond_plan is the flag saying the plan is no longer the authority. If the
// digest drops it, the resumed session re-scopes to the plan and silently loses
// the work that crossed it — the field failure this whole design came from.
func TestBeyondPlanSurvives(t *testing.T) {
	dir := withNote(t, strings.Replace(note, "beyond_plan: false", "beyond_plan: true", 1))
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	if got := injected(t, stdout); !strings.Contains(got, "beyond_plan: true") {
		t.Errorf("beyond_plan dropped from the digest:\n%s", got)
	}
}

// A heading typo must not cost the whole digest. An agent under time pressure
// writes "Next steps" where the schema says "Next intended steps", and silence
// there is total continuity loss with no symptom.
func TestOffSchemaHeadingsStillRestore(t *testing.T) {
	off := "---\nschema: 2\nobjective: \"ship it\"\n---\n" +
		"## Checks\n1. make verify → green\n" +
		"## Next steps\n1. wire the hook\n" +
		"## Files touched\n- main.go\n"
	dir := withNote(t, off)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	if !strings.Contains(got, "make verify") {
		t.Errorf("off-schema headings produced no digest:\n%s", got)
	}
	if strings.Contains(got, "main.go") {
		t.Errorf("fallback pulled in low-value recovery detail:\n%s", got)
	}
}

// The fallback must not fire when the schema headings DID match — otherwise
// every digest quietly grows to the whole note.
func TestFallbackDoesNotFireWhenSchemaHeadingsMatch(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	if strings.Contains(got, "Files touched") || strings.Contains(got, "Open threads") {
		t.Errorf("digest grew past the priority sections:\n%s", got)
	}
}

// A SessionStart hook that fails is a session that fails.
func TestNeverBlocksTheSession(t *testing.T) {
	cases := []struct{ name, dir, stdin string }{
		{"no project dir", "", `{"source":"startup"}`},
		{"no note", t.TempDir(), `{"source":"startup"}`},
		{"malformed stdin", withNote(t, note), `{not json`},
		{"empty stdin", withNote(t, note), ``},
		{"unknown source", withNote(t, note), `{"source":"teleport"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, code := call(t, c.dir, c.stdin)
			if code != 0 {
				t.Errorf("exit %d, want 0", code)
			}
		})
	}
}

// Silence must be silence: an empty additionalContext still costs tokens to say
// there is nothing to say.
func TestNoNoteEmitsNothingAtAll(t *testing.T) {
	stdout, _, _ := call(t, t.TempDir(), `{"source":"compact"}`)
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("emitted %q with no note present", stdout)
	}
}

// The note is the tattoo, not the autobiography — and truncation must leave the
// path, because a cut digest with no way back to the full note is worse than
// either.
func TestOversizeDigestIsTruncatedAndStillPointsHome(t *testing.T) {
	long := strings.Replace(note, "   last run: pass",
		"   last run: "+strings.Repeat("verbose ", 2000), 1)
	dir := withNote(t, long)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	if len(got) > maxDigest+200 {
		t.Errorf("digest length %d, want ~%d", len(got), maxDigest)
	}
	if !strings.Contains(got, "truncated") || !strings.Contains(got, "CHECKPOINT.md") {
		t.Errorf("truncated digest lost its pointer home:\n%s", got[max(0, len(got)-300):])
	}
}

func TestVersion(t *testing.T) {
	stdout, _, code := call(t, t.TempDir(), ``, "-version")
	if code != 0 || !strings.Contains(stdout, "sc-checkpoint-restore") {
		t.Errorf("version: code=%d stdout=%q", code, stdout)
	}
}

// withTree builds a project whose trigger surfaces actually exist, since
// WatchTargets refuses to register a path it cannot stat.
func withTree(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{filepath.Join(".claude", "checkpoints"), "tools", "manifest"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func watchOf(t *testing.T, stdout string) []string {
	t.Helper()
	var out hookOutput
	if strings.TrimSpace(stdout) == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("not a hook response: %v", err)
	}
	return out.HookSpecificOutput.WatchPaths
}

const watchNote = `---
schema: 2
status: in-progress
objective: "wire the watcher"
---
## Validation loop
1. ` + "`go test ./...`" + `  → ok  · re-armed by: any .go edit under tools/
   last run: pass
2. ` + "`make check`" + `  → ok  · re-armed by: manifest/*.yml
   last run: pass
`

// The measured constraint end-to-end: a note's trigger surfaces come back as
// DIRECTORIES the watcher accepts, never as the patterns the note wrote.
func TestWatchPathsAreRegisteredAsDirectories(t *testing.T) {
	dir := withTree(t, watchNote)
	stdout, _, code := call(t, dir, `{"source":"compact"}`)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := watchOf(t, stdout)
	if len(got) != 2 {
		t.Fatalf("watchPaths = %v, want tools and manifest", got)
	}
	for _, p := range got {
		if strings.ContainsAny(p, "*?[") {
			t.Errorf("registered a pattern %q — measured to register NOTHING, silently", p)
		}
	}
}

// A pointer session is not resuming this work, so registering its surfaces would
// collect re-arms nobody asked for.
func TestPointerSessionsRegisterNoWatch(t *testing.T) {
	dir := withTree(t, watchNote)
	stdout, _, _ := call(t, dir, `{"source":"clear"}`)
	if got := watchOf(t, stdout); len(got) != 0 {
		t.Errorf("watchPaths on a cleared session: %v", got)
	}
}

// An unresolvable surface is DATA. A session that silently watched nothing must
// not look identical to one that watched everything.
func TestUnresolvableSurfacesAreReported(t *testing.T) {
	dir := withTree(t, `---
schema: 2
status: in-progress
objective: "prose surfaces"
---
## Validation loop
1. `+"`make release`"+`  → tagged  · re-armed by: a human deciding to ship
   last run: pass
`)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	if !strings.Contains(got, "could not be resolved") {
		t.Errorf("silent about an unwatchable surface:\n%s", got)
	}
}

// Re-armed checks are surfaced, because a compacted agent reading `last run:
// pass` off its own note is exactly the failure I2 exists to catch.
func TestRearmedChecksAppearInTheDigest(t *testing.T) {
	dir := withTree(t, watchNote)
	state := checkpoint.RearmState{Rearmed: map[string]checkpoint.Rearm{
		"1": {Index: 1, Check: "1. `go test ./...`", By: "tools/x.go", Event: "change", At: "2026-07-29T12:00:00Z"},
	}}
	b, _ := json.Marshal(state)
	if err := os.WriteFile(checkpoint.RearmPath(dir), b, 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	got := injected(t, stdout)
	for _, want := range []string{"stale", "check 1", "tools/x.go", "change"} {
		if !strings.Contains(got, want) {
			t.Errorf("digest missing %q:\n%s", want, got)
		}
	}
}

// Nothing re-armed says nothing: a clean start must not spend tokens reporting
// the absence of a problem.
func TestNoRearmsIsSilentOnTheSubject(t *testing.T) {
	dir := withTree(t, watchNote)
	stdout, _, _ := call(t, dir, `{"source":"compact"}`)
	if got := injected(t, stdout); strings.Contains(got, "stale") {
		t.Errorf("reported staleness with no re-arms recorded:\n%s", got)
	}
}

// Corrupt re-arm state must not cost a session start.
func TestCorruptRearmStateDegradesToSilence(t *testing.T) {
	dir := withTree(t, watchNote)
	if err := os.WriteFile(checkpoint.RearmPath(dir), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, code := call(t, dir, `{"source":"compact"}`)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if got := injected(t, stdout); got == "" || strings.Contains(got, "stale") {
		t.Errorf("corrupt state changed the digest:\n%s", got)
	}
}
