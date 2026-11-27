package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// PREINVOCATION ONLY HEALS: it writes no session row (SessionStart's work runs from SessionStart)
// and injects nothing into the conversation.
func TestPreInvocationWritesNoRowAndInjectsNothing(t *testing.T) {
	const conv = "1a2b3c4d-0000-4000-8000-00000000000f"
	dir := t.TempDir()
	transcript := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(transcript, []byte(`{"step_index":0}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := fmt.Sprintf(`{"conversationId":%q,"invocationNum":1,"workspacePaths":[%s],"transcriptPath":%s}`,
		conv, strconv.Quote(dir), strconv.Quote(transcript))
	stdout, stderr, code := call(t, in, "", okStat(10), "-event", "PreInvocation")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}
	if rows := readManifest(t, dir, conv); len(rows) != 0 {
		t.Errorf("PreInvocation wrote %d session rows", len(rows))
	}
	if strings.Contains(stdout, "ephemeralMessage") || strings.Contains(stdout, "telepathy") {
		t.Errorf("PreInvocation injected into the conversation: %s", stdout)
	}
}
