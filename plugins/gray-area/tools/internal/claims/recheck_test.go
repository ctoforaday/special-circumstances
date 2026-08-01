package claims

import (
	"strings"
	"testing"
)

// THE DEFECT (plans/gray-area.md §11.9, measured 2026-07-31).
//
// On the FIRST SessionStart of a newly minted session id, the harness hands over
// a transcript_path whose file does not exist yet: captured 03:43:31Z, present by
// 03:45:18Z. The capture hook wrote `resolved: false` and nothing ever revisited
// it, so `gray-area checkpoint` refused for that session's ENTIRE life — a
// one-time stat recorded as a permanent property.
//
// The refusal was correct about the row and wrong about the world.
func TestARowThatHasSinceResolvedIsUsable(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): sessionRow("2026-07-31T03:43:31Z", "S1", "/t/late.jsonl", false) + "\n",
	})
	// The file exists NOW, as it did by 03:45:18Z in the measured case.
	got, err := ResolveSession(manifestDir(), glob, open, statOnly("/t/late.jsonl"))
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if !got.Resolved {
		t.Fatal("the transcript is readable and the row was still reported unusable — this is the bug: a stat taken once, treated as a property")
	}
	if !got.Recovered {
		t.Error("the correction is invisible; a reader grepping the manifest will see resolved:false and not know why the tool proceeded")
	}
	if got.TranscriptPath != "/t/late.jsonl" {
		t.Errorf("resolved %q", got.TranscriptPath)
	}
}

// The question is which transcript can be READ, not which claim about one was
// written last. A newer row naming a file that is gone must not beat an older
// row naming a file that is there.
func TestAReadableOlderRowBeatsAnUnreadableNewerOne(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): sessionRow("2026-07-31T03:00:00Z", "S1", "/t/here.jsonl", true) + "\n" +
			sessionRow("2026-07-31T04:00:00Z", "S1", "/t/gone.jsonl", true) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly("/t/here.jsonl"))
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if got.TranscriptPath != "/t/here.jsonl" {
		t.Errorf("resolved %q — the newest row wins even when its transcript cannot be read", got.TranscriptPath)
	}
}

// The re-check runs in both directions, for the same reason. A row recorded as
// resolved whose transcript has since vanished must be demoted, and must say so
// — otherwise the failure surfaces later as an unexplained open error.
func TestARowWhoseTranscriptVanishedIsDemotedWithAReason(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): sessionRow("2026-07-31T03:00:00Z", "S1", "/t/deleted.jsonl", true) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly())
	if err != nil {
		t.Fatalf("an unreadable row must still be returned so the caller can explain it: %v", err)
	}
	if got.Resolved {
		t.Fatal("a transcript that is gone was reported as resolved, on the strength of a stat taken earlier")
	}
	if !strings.Contains(got.CaptureError, "not readable now") {
		t.Errorf("the reason does not distinguish 'never resolved' from 'resolved then vanished': %q", got.CaptureError)
	}
	if got.Recovered {
		t.Error("Recovered marks the opposite correction and must not be set here")
	}
}

// Recovery must not be claimed when nothing recovered — a row that was false and
// is still false reports exactly what it did before.
func TestARowThatStillDoesNotResolveIsNotMarkedRecovered(t *testing.T) {
	glob, open := fakeFS(map[string]string{
		manifestFile("S1"): sessionRow("2026-07-31T03:43:31Z", "S1", "/t/never.jsonl", false) + "\n",
	})
	got, err := ResolveSession(manifestDir(), glob, open, statOnly())
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if got.Resolved || got.Recovered {
		t.Errorf("resolved=%v recovered=%v; want both false", got.Resolved, got.Recovered)
	}
	if got.CaptureError == "" {
		t.Error("the original capture error was discarded, leaving the caller nothing to report")
	}
}
