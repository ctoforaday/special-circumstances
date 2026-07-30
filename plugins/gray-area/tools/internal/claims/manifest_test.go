package claims

import (
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

func TestResolvesTheNewestSessionRow(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): seatRow + "\n" + sessionRow("2026-07-30T09:00:00Z", "S1", "/t/old.jsonl", true) + "\n",
		manifestFile("S2"): sessionRow("2026-07-30T10:00:00Z", "S2", "/t/new.jsonl", true) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open)
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
	_, err := ResolveSession(manifestDir(), glob, open)
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
	got, err := ResolveSession(manifestDir(), glob, open)
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
	_, err := ResolveSession(manifestDir(), glob, open)
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
	got, err := ResolveSession(manifestDir(), glob, open)
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
