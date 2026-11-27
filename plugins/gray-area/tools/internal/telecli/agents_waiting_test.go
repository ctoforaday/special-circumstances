package telecli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	waitA = "aaaa1111-0000-4000-8000-00000000000a" // idle, two messages, a notice and an older wake
	waitB = "bbbb2222-0000-4000-8000-00000000000b" // a message, a missing file, a torn file
	waitC = "cccc3333-0000-4000-8000-00000000000c" // agy's own notices only
	waitD = "dddd4444-0000-4000-8000-00000000000d" // everything read
	waitE = "eeee5555-0000-4000-8000-00000000000e" // telepathy's wake messages only
)

// queueMessage writes one mailbox message to conv as the client lays it out; a "" body leaves the
// marker with no message file.
func queueMessage(t *testing.T, brain, conv, id, body string) {
	t.Helper()
	box := filepath.Join(brain, conv, ".system_generated", "messages")
	if err := os.MkdirAll(filepath.Join(box, "undelivered"), 0o755); err != nil {
		t.Fatal(err)
	}
	if body != "" {
		if err := os.WriteFile(filepath.Join(box, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(box, "undelivered", id), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mailboxJSON(sender, ts string) string {
	b, _ := json.Marshal(map[string]any{"id": "x", "recipient": "r", "sender": sender, "priority": "MESSAGE_PRIORITY_HIGH",
		"timestamp": ts, "renderDetails": map[string]string{"messageTitle": "t"}, "content": "c"})
	return string(b)
}

// waitingBrain is a brain directory with four conversations, none held by any process.
func waitingBrain(t *testing.T) string {
	t.Helper()
	brain := t.TempDir()
	queueMessage(t, brain, waitA, "a1", mailboxJSON("caller-1", "2026-09-09T09:00:00Z"))
	queueMessage(t, brain, waitA, "a2", mailboxJSON("telepathy", "2026-09-08T22:15:00.123Z"))
	queueMessage(t, brain, waitA, "a3", mailboxJSON("system", "2026-09-01T00:00:00Z"))
	queueMessage(t, brain, waitA, "a4", mailboxJSON("telepathy-wake", "2026-09-01T00:00:00Z"))
	queueMessage(t, brain, waitE, "e1", mailboxJSON("telepathy-wake", "2026-09-09T10:00:00Z"))
	queueMessage(t, brain, waitB, "b1", mailboxJSON("e1e1e1e1-0000-4000-8000-000000000001", "2026-09-09T11:00:00Z"))
	queueMessage(t, brain, waitB, "b2", "")
	queueMessage(t, brain, waitB, "b3", "{torn")
	queueMessage(t, brain, waitC, "c1", mailboxJSON("system", "2026-09-09T10:00:00Z"))
	queueMessage(t, brain, waitC, "c2", mailboxJSON("system", "2026-09-09T10:05:00Z"))
	// D's message was read: the client removed its marker and kept the file.
	queueMessage(t, brain, waitD, "d1", mailboxJSON("caller-1", "2026-09-09T08:00:00Z"))
	if err := os.Remove(filepath.Join(brain, waitD, ".system_generated", "messages", "undelivered", "d1")); err != nil {
		t.Fatal(err)
	}
	return brain
}

func TestAgentsWaitingGolden(t *testing.T) {
	h := newHarness(t)
	h.env.BrainDirs = []string{waitingBrain(t)}
	h.env.Getenv = func(k string) string { return map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}[k] }
	out, errOut, code := h.run(t, "agents")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	assertGolden(t, "agents-waiting", out)
}

// THE CONVERSATIONS WITH MAIL WAITING ARE LISTED WHETHER OR NOT ANYTHING RUNS THEM, and "yours" is
// the caller's own sender id on the message.
func TestAgentsWaitingListsIdleConversations(t *testing.T) {
	h := newHarness(t)
	h.env.BrainDirs = []string{waitingBrain(t)}
	h.env.Getenv = func(k string) string { return map[string]string{"CLAUDE_CODE_SESSION_ID": "caller-1"}[k] }
	out, _, _ := h.run(t, "agents")
	rows := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		for _, id := range []string{waitA, waitB, waitC, waitD, waitE} {
			if strings.HasPrefix(line, id[:8]+" ") {
				rows[id] = line
			}
		}
	}
	if !strings.Contains(rows[waitA], "      2  ") || !strings.Contains(rows[waitA], "1 yours, 1 from a shell, +1 agy notice, +1 wake") || !strings.Contains(rows[waitA], "2026-09-08 22:15:00 UTC") {
		t.Errorf("A's row = %q, want two waiting, the shell's the oldest (not the older wake), one yours, a wake apart", rows[waitA])
	}
	if !strings.Contains(rows[waitE], "      0  ") || !strings.Contains(rows[waitE], "+1 wake") {
		t.Errorf("E's row = %q, want a row for a wake alone, counting no message", rows[waitE])
	}
	if !strings.Contains(rows[waitC], "+2 agy notices") {
		t.Errorf("C's row = %q, want a row for agy's notices alone", rows[waitC])
	}
	if _, ok := rows[waitD]; ok {
		t.Errorf("D, whose message was read, is listed: %q", rows[waitD])
	}
	// From a plain shell nothing is "yours", and the shell's own messages are "from a shell".
	h.env.Getenv = func(string) string { return "" }
	out, _, _ = h.run(t, "agents")
	if strings.Contains(out, "yours") {
		t.Errorf("from a plain shell a message is marked yours:\n%s", out)
	}
	if out, _, _ := h.run(t, "agents", "--lost", "--boot", bootArg()); strings.Contains(out, "waiting for Antigravity") {
		t.Errorf("--lost printed the waiting section:\n%s", out)
	}
}

func TestAgentsWaitingUnreadableMessage(t *testing.T) {
	h := newHarness(t)
	h.env.BrainDirs = []string{waitingBrain(t)}
	out, _, _ := h.run(t, "agents")
	var row string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, waitB[:8]+" ") {
			row = line
		}
	}
	if fields := strings.Fields(row); len(fields) < 2 || fields[1] != "3" || !strings.Contains(row, "2 unreadable") {
		t.Errorf("B's row = %q, want 3 waiting, 2 of them unreadable", row)
	}
}

func TestAgentsWaitingScanError(t *testing.T) {
	h := newHarness(t)
	// A path no platform can open (it holds a NUL byte) fails the listing with an error that is not
	// "does not exist", the same on Unix and Windows.
	h.env.BrainDirs = []string{filepath.Join(t.TempDir(), "brain\x00")}
	out, _, code := h.run(t, "agents")
	if code != 0 || !strings.Contains(out, "waiting messages not measured: ") || strings.Contains(out, "no Antigravity conversation has a message waiting") {
		t.Errorf("exit %d, an unlistable brain directory reads:\n%s", code, out)
	}
}
