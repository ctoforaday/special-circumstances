package pushfreezeguard

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

// A MARKER THAT IS NOT JSON AT ALL made runlive.Read return the zero State, so this guard saw no live
// run and warned about NOTHING — the flattering direction, on a guard whose silence means "nothing
// is frozen". It is recorded now, and a marker that parses again clears it.
func TestAnUnparsableMarkerIsRecordedNotReadAsNoRun(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	project := liveRun(t, "{ this is not json")
	rec := hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now(), io.Discard)
	runGuard(t, project, "git push", rec)
	msg := rec.Settle()
	if !strings.Contains(msg, "- "+string(StageMarkerUnparsable)+" in ") {
		t.Fatalf("an unparsable marker was not said: %q", msg)
	}

	rewriteMarker(t, project, `{"runs":[{"runDir":"/r","pinnedPaths":["a"]}]}`)
	rec = hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now().Add(time.Hour), io.Discard)
	runGuard(t, project, "git push", rec)
	if msg := rec.Settle(); strings.Contains(msg, string(StageMarkerUnparsable)) {
		t.Errorf("a parseable marker left the entry: %q", msg)
	}
}

// A FAILURE IN ONE PROJECT IS NOT CLEARED BY SUCCESS IN ANOTHER. This machine runs dozens of
// worktrees at once, each its own project; a healthy marker in one says nothing about a broken
// marker in the next, and a record keyed by stage alone would have let it wipe the entry.
func TestAHealthyProjectDoesNotClearABrokenOne(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	broken := liveRun(t, "{ not json")
	healthy := liveRun(t, `{"runs":[{"runDir":"/r","pinnedPaths":["a"]}]}`)

	rec := hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now(), io.Discard)
	runGuard(t, broken, "git push", rec)
	rec.Settle()
	rec = hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now().Add(time.Hour), io.Discard)
	runGuard(t, healthy, "git push", rec)
	if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageMarkerUnparsable)+" in "+broken) {
		t.Fatalf("a healthy marker elsewhere cleared the broken project's entry: %q", msg)
	}
}
