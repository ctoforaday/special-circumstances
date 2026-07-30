package claims

import (
	"path/filepath"
	"strings"
	"testing"
)

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
		"/p/.claude/gray-area/trajectories-S1.jsonl": seatRow + "\n" + sessionRow("2026-07-30T09:00:00Z", "S1", "/t/old.jsonl", true) + "\n",
		"/p/.claude/gray-area/trajectories-S2.jsonl": sessionRow("2026-07-30T10:00:00Z", "S2", "/t/new.jsonl", true) + "\n",
	})
	got, err := ResolveSession("/p/.claude/gray-area", glob, open)
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
		"/p/.claude/gray-area/trajectories-S1.jsonl": seatRow + "\n" + seatRow + "\n",
	})
	_, err := ResolveSession("/p/.claude/gray-area", glob, open)
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
		"/p/.claude/gray-area/trajectories-S1.jsonl": sessionRow("2026-07-30T09:00:00Z", "S1", "/gone.jsonl", false) + "\n",
	})
	got, err := ResolveSession("/p/.claude/gray-area", glob, open)
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
	_, err := ResolveSession("/p/.claude/gray-area", glob, open)
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
		"/p/.claude/gray-area/trajectories-S1.jsonl": "{not json\n" + sessionRow("2026-07-30T09:00:00Z", "S1", "/t/ok.jsonl", true) + "\n",
	})
	got, err := ResolveSession("/p/.claude/gray-area", glob, open)
	if err != nil {
		t.Fatalf("one bad line lost the whole manifest: %v", err)
	}
	if got.TranscriptPath != "/t/ok.jsonl" {
		t.Errorf("resolved %q", got.TranscriptPath)
	}
}
