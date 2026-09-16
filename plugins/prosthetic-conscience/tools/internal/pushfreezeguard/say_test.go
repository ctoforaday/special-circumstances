package pushfreezeguard

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// THE GUARD NEVER BLOCKS, so its warning is the entire mechanism — and it used to go to stderr at
// exit 0, which reaches the debug log and nobody. The 2026-07-17 incident class it names is a
// data-loss event; a guard whose only output is invisible protects no one.

func liveRun(t *testing.T, marker string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "run-live.json"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runGuard(t *testing.T, project, command string, rec *hookfailures.Recorder) hookunit.Result {
	t.Helper()
	payload := `{"tool_name":"Bash","tool_input":{"command":` + strconv.Quote(command) + `}}`
	ctx := hookunit.NewCtx("PreToolUse", []byte(payload), project, time.Time{}, rec)
	return Unit().Run(ctx)
}

// Every advisory the guard has is SAID, not merely logged.
func TestEveryAdvisoryIsSaid(t *testing.T) {
	project := liveRun(t, `{"runs":[{"runDir":"/r","pinnedPaths":["a/b"]}]}`)
	for _, command := range []string{
		"git push", "git add -A", "git checkout main", "git switch main",
		"git stash -u", "git stash pop",
	} {
		t.Run(command, func(t *testing.T) {
			r := runGuard(t, project, command, testRecorder())
			if r.Say == "" {
				t.Fatalf("%q warned nobody: Say is empty (stderr %q)", command, r.Stderr)
			}
			if r.Say != r.Stderr {
				t.Errorf("the debug log and the human were told different things:\n say %q\n log %q", r.Say, r.Stderr)
			}
		})
	}
}

// A FREEZE THE GUARD CANNOT READ is not an advisory about this call — it is a broken guard, and it
// stays broken until the marker can be read. So it is RECORDED as well as said, and the record
// carries it to the next displaying event even if this call is never repeated.
func TestAnUnreadableFreezeIsRecorded(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	path, err := hookfailures.Path("prosthetic-conscience")
	if err != nil {
		t.Fatal(err)
	}

	project := liveRun(t, `{"runs":[{"runDir":"/r"}]}`) // live, and names NO pinned paths
	rec := hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now(), io.Discard)
	r := runGuard(t, project, "git push", rec)
	if !strings.Contains(r.Say, "could not read them") {
		t.Fatalf("the unreadable freeze was not said: %q", r.Say)
	}
	rec.Settle()

	var record hookfailures.Record
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("nothing was recorded: %v", err)
	}
	if err := json.Unmarshal(b, &record); err != nil {
		t.Fatal(err)
	}
	if len(record.Failures) != 1 || record.Failures[0].Stage != StageFreezeUnreadable {
		t.Fatalf("record = %+v", record.Failures)
	}

	// ...and a freeze it CAN read, IN THE SAME PROJECT, clears it: the guard is working again. A
	// readable marker in some other project says nothing about this one.
	rewriteMarker(t, project, `{"runs":[{"runDir":"/r","pinnedPaths":["a/b"]}]}`)
	rec = hookfailures.New("prosthetic-conscience", "sc-pretooluse", "PreToolUse", time.Now(), io.Discard)
	runGuard(t, project, "git push", rec)
	rec.Settle()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		b, _ := os.ReadFile(path)
		t.Errorf("a readable freeze left the entry behind: %s", b)
	}
}

// No live run: nothing to warn about, and nothing recorded. The guard is silent when the thing it
// guards is not happening — which is what makes its speech mean something.
func TestNoLiveRunSaysNothing(t *testing.T) {
	r := runGuard(t, t.TempDir(), "git push", testRecorder())
	if r.Say != "" || r.Stderr != "" {
		t.Errorf("say %q, stderr %q", r.Say, r.Stderr)
	}
}

func rewriteMarker(t *testing.T, project, marker string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(project, ".claude", "run-live.json"), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
}
