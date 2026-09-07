package claims

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// FIXTURE PATHS ARE BUILT THE WAY THE CODE BUILDS THEM.
//
// The first draft keyed the fake filesystem on forward slashes while
// ResolveSession composes its glob with filepath.Join, so on Windows the pattern
// came out with backslashes and matched nothing — three tests green on Linux, red
// on the runner. This repo has hit that exact shape twice now, and the checkpoint
// note carries it as a foot-gun.
//
// The fix is to make the fixture platform-agnostic rather than to loosen the
// matcher: production is consistent, both sides coming from filepath.Join, so a
// separator-tolerant glob would paper over a test that was lying about its
// environment rather than fix anything real.
func manifestDir() string { return filepath.Join("p", ".claude", "gray-area") }

func manifestFile(session string) string {
	return filepath.Join(manifestDir(), "trajectories-"+session+".jsonl")
}

func fakeFS(files map[string]string) (func(string) ([]string, error), func(string) ([]byte, error)) {
	glob := func(pattern string) ([]string, error) {
		var out []string
		for p := range files {
			if ok, _ := filepath.Match(pattern, p); ok {
				out = append(out, p)
			}
		}
		return out, nil
	}
	open := func(p string) ([]byte, error) { return []byte(files[p]), nil }
	return glob, open
}

const seatRow = `{"schema":1,"kind":"seat","captured_at":"2026-07-30T09:00:00Z","session_id":"S1","agent_type":"red","agent_transcript_path":"/t/a.jsonl","resolved":true}`

func sessionRow(at, id, path string, resolved bool) string {
	r := `{"schema":1,"kind":"session","captured_at":"` + at + `","session_id":"` + id + `","transcript_path":"` + path + `","resolved":`
	if resolved {
		return r + "true}"
	}
	return r + `false,"capture_error":"stat: no such file"}`
}

// statOnly reports nil for the listed paths and an error for anything else, so a
// test states which transcripts exist ON DISK independently of what the manifest
// rows CLAIM. Those are different facts, and conflating them is the defect
// ResolveSession's re-check exists to fix (§11.9).
func statOnly(readable ...string) func(string) error {
	set := map[string]bool{}
	for _, p := range readable {
		set[p] = true
	}
	return func(p string) error {
		if set[p] {
			return nil
		}
		return fmt.Errorf("stat %s: no such file or directory", p)
	}
}

func TestResolvesTheNewestSessionRow(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): seatRow + "\n" + sessionRow("2026-07-30T09:00:00Z", "S1", "/t/old.jsonl", true) + "\n",
		manifestFile("S2"): sessionRow("2026-07-30T10:00:00Z", "S2", "/t/new.jsonl", true) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly("/t/old.jsonl", "/t/new.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if got.TranscriptPath != "/t/new.jsonl" {
		t.Errorf("resolved %q, want the newest by captured_at", got.TranscriptPath)
	}
	// The pick is a claim, so it must be citable.
	if got.Manifest == "" || got.Line == 0 {
		t.Errorf("the resolved row carries no provenance: %+v", got)
	}
}

// A manifest holding ONLY seat rows is the live case for any project whose
// SessionStart hook is not wired — and answering it with a seat's transcript
// would silently adjudicate the wrong document.
func TestSeatRowsAreNeverOfferedAsTheSessionsTranscript(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): seatRow + "\n" + seatRow + "\n",
	})
	_, err := ResolveSession(manifestDir(), glob, open, statOnly())
	if err == nil {
		t.Fatal("a seat row was offered as the session's transcript")
	}
	if !strings.Contains(err.Error(), "SessionStart") {
		t.Errorf("the error does not name the missing hook, so a reader cannot fix it: %v", err)
	}
}

// An UNRESOLVED row is returned rather than skipped, so the caller can say why it
// is unusable instead of reporting the same "nothing found" it would report for
// an empty directory. Those are different problems with different fixes.
func TestAnUnresolvedRowIsReturnedSoTheCallerCanExplainIt(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): sessionRow("2026-07-30T09:00:00Z", "S1", "/gone.jsonl", false) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly())
	if err != nil {
		t.Fatalf("an unresolved row was treated as no row at all: %v", err)
	}
	if got.Resolved {
		t.Error("the row claims to be resolved")
	}
	if got.CaptureError == "" {
		t.Error("no reason recorded, so the caller can only say \"nothing found\"")
	}
}

// No manifest at all names the actionable fix rather than just failing.
func TestNoManifestNamesTheFix(t *testing.T) {
	glob, open := fakeFS(map[string]string{})
	_, err := ResolveSession(manifestDir(), glob, open, statOnly())
	if err == nil {
		t.Fatal("an empty directory resolved to something")
	}
	if !strings.Contains(err.Error(), "SessionStart") {
		t.Errorf("the error does not name the hook that would fix it: %v", err)
	}
}

// Malformed lines are skipped, not fatal: a manifest is append-only and
// best-effort by design, so one bad row must not cost the whole lookup.
func TestAMalformedRowDoesNotCostTheWholeLookup(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): "{not json\n" + sessionRow("2026-07-30T09:00:00Z", "S1", "/t/ok.jsonl", true) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly("/t/ok.jsonl"))
	if err != nil {
		t.Fatalf("one bad line lost the whole manifest: %v", err)
	}
	if got.TranscriptPath != "/t/ok.jsonl" {
		t.Errorf("resolved %q", got.TranscriptPath)
	}
}

// THE FIXTURE CLASS, PARTLY GUARDED — and the limit is stated because a check
// that looks like protection and is not is worse than none.
//
// Twice now a fixture has keyed on a literal `/` while the code composed the same
// path with filepath.Join: green on Linux, red on the Windows runner, and the
// failure said nothing about the behaviour under test.
//
// ON A FORWARD-SLASH PLATFORM THIS TEST IS WEAK. It would pass against the broken
// fixture too, because there `/` IS the separator and both spellings agree. It
// earns its place on Windows, where the two diverge. The real check is CI's
// windows-latest job; `GOOS=windows go vet` does NOT substitute, because this is
// a runtime path-matching behaviour and not a compile error.
func TestFixturePathsAgreeWithHowTheCodeComposesThem(t *testing.T) {
	pattern := filepath.Join(manifestDir(), "trajectories-*.jsonl")
	name := manifestFile("S1")
	ok, err := filepath.Match(pattern, name)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("fixture %q does not match the glob %q the code builds — on this platform the suite would pass or fail for reasons unrelated to what it is testing", name, pattern)
	}
	if filepath.Separator != '/' && strings.Contains(name, "/") {
		t.Errorf("fixture path %q carries a hardcoded separator", name)
	}
}

// THE EMPTY BUILD GETS WORDS. A row from a pre-schema-5 hook binary has no capture_build,
// and rendering that as a blank beside "written by build" reads as a build whose name went
// missing rather than as a row that never recorded one — the two states the field was added
// to separate. Both directions are asserted because only one of them is ever exercised by a
// fresh install, so the other is the one that rots.
func TestWriterProvenanceSeparatesNoBuildFromABuild(t *testing.T) {
	if got := (SessionRow{CaptureBuild: "abc1234"}).WriterProvenance(); !strings.Contains(got, "abc1234") {
		t.Errorf("WriterProvenance() = %q, does not name the build", got)
	}
	got := SessionRow{}.WriterProvenance()
	if got == "" || strings.Contains(got, "written by build ") {
		t.Errorf("WriterProvenance() with no build = %q; an unrecorded writer must be stated, "+
			"never rendered as a build with an empty name", got)
	}
	if !strings.Contains(got, "not recorded") {
		t.Errorf("WriterProvenance() with no build = %q, want it to say the build was not recorded", got)
	}
}

// The projection must carry the field through, or the render above has nothing to render:
// ResolveSession is the only path by which a row reaches the verbs that cite it.
func TestResolvedSessionCarriesTheWritersBuild(t *testing.T) {
	dir := filepath.Join("p", ".claude", "gray-area")
	path := filepath.Join(dir, "trajectories-S9.jsonl")
	trace := filepath.Join("t", "s9.jsonl")
	row := `{"schema":5,"kind":"session","captured_at":"2026-09-07T09:00:00Z","session_id":"S9",` +
		`"transcript_path":"` + strings.ReplaceAll(trace, `\`, `\\`) + `","resolved":true,"capture_build":"dd44556"}`

	got, err := ResolveSession(dir,
		func(string) ([]string, error) { return []string{path}, nil },
		func(string) ([]byte, error) { return []byte(row), nil },
		func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if got.CaptureBuild != "dd44556" {
		t.Errorf("CaptureBuild = %q, want dd44556 — the row named its writer and the projection dropped it", got.CaptureBuild)
	}
}
