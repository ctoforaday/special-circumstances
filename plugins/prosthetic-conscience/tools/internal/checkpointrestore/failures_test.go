package checkpointrestore

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

// THE NOTE IS THERE AND CANNOT BE READ. compose Stats it, then the read fails, and it used to return
// the same ("", nil) as a project with no note at all — so a session lost its own cursor and was told
// nothing. The restore still degrades to no digest; the reason is now recorded, and a readable note
// clears it.
func TestANoteThatCannotBeReadIsRecordedNotReadAsNoNote(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	note := filepath.Join(cp, "CHECKPOINT.md")
	if err := os.WriteFile(note, []byte("---\nschema: 3\n---\n## Objective\nx\n"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(note, 0o644) })
	// A CAPABILITY check, not an identity check (see internal/unwritable): skip only if the read
	// still lands, whatever uid or platform this is.
	if _, err := os.ReadFile(note); err == nil {
		t.Skip("chmod 000 did not stop this process reading the note — the premise never took")
	}

	rec := hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now(), io.Discard)
	if text, _ := compose(dir, "startup", rec); text != "" {
		t.Errorf("an unreadable note produced a digest: %q", text)
	}
	if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageNoteRead)+" in ") {
		t.Fatalf("the lost cursor was not said: %q", msg)
	}

	if err := os.Chmod(note, 0o644); err != nil {
		t.Fatal(err)
	}
	rec = hookfailures.New("prosthetic-conscience", "sc-sessionstart", "SessionStart", time.Now().Add(time.Hour), io.Discard)
	compose(dir, "startup", rec)
	if msg := rec.Settle(); strings.Contains(msg, string(StageNoteRead)) {
		t.Errorf("a readable note left the entry: %q", msg)
	}
}
