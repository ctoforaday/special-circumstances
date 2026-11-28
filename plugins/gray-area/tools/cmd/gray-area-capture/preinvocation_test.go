package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// PREINVOCATION ONLY HEALS: on a turn's first step and mid-turn alike it writes no session row
// (SessionStart's work runs from SessionStart) and injects nothing into the conversation.
func TestPreInvocationWritesNoRowAndInjectsNothing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fetch-bin.sh is a POSIX shell script")
	}
	// The version check a turn's first step runs, stood in by a fetch-bin.sh that reports success.
	root := t.TempDir()
	script := filepath.Join(root, "hooks", "fetch-bin.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte("[ \"$1\" = ensure ] && printf '{\"systemMessage\":\"pre-heal-ok\"}\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")
	t.Setenv("PLUGIN_ROOT", root)
	t.Setenv("PLUGIN_DATA", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	const conv = "1a2b3c4d-0000-4000-8000-00000000000f"
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(transcript, []byte(`{"step_index":0}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1} {
		in := fmt.Sprintf(`{"hookEventName":"PreInvocation","conversationId":%q,"invocationNum":%d,"workspacePaths":[%s],"transcriptPath":%s}`,
			conv, n, strconv.Quote(dir), strconv.Quote(transcript))
		stdout, stderr, code := call(t, in, "", okStat(10), "-event", "PreInvocation")
		if code != 0 {
			t.Fatalf("invocationNum %d: exit %d, stderr: %s", n, code, stderr)
		}
		if n == 0 && !strings.Contains(stdout, "pre-heal-ok") {
			t.Errorf("invocationNum 0 did not run the version check: %s", stdout)
		}
		if rows := readManifest(t, dir, conv); len(rows) != 0 {
			t.Errorf("invocationNum %d: PreInvocation wrote %d session rows", n, len(rows))
		}
		if strings.Contains(stdout, "ephemeralMessage") || strings.Contains(stdout, "telepathy") {
			t.Errorf("invocationNum %d: PreInvocation injected into the conversation: %s", n, stdout)
		}
	}
}
