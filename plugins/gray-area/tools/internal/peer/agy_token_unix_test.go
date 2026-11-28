//go:build unix

package peer

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadDaemonTokenRefuses(t *testing.T) {
	dir := t.TempDir()
	const secret = "s3cret-Value-42"
	write := func(name, body string, mode os.FileMode) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, mode); err != nil {
			t.Fatal(err)
		}
		return p
	}
	good := write("good", DaemonTokenVar+"="+secret+"\n", 0o600)
	if tok, err := ReadDaemonToken(good); err != nil || tok != secret {
		t.Fatalf("ReadDaemonToken(good) = %q, %v", tok, err)
	}
	if tok, err := ReadDaemonToken(write("nonl", DaemonTokenVar+"="+secret, 0o600)); err != nil || tok != secret {
		t.Errorf("no trailing newline = %q, %v", tok, err)
	}
	if _, err := ReadDaemonToken(filepath.Join(dir, "absent")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("absent = %v, want fs.ErrNotExist", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(good, link); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "adir")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{
		"symlink":      link,
		"directory":    sub,
		"0640":         write("g", DaemonTokenVar+"="+secret+"\n", 0o640),
		"0604":         write("o", DaemonTokenVar+"="+secret+"\n", 0o604),
		"two lines":    write("two", DaemonTokenVar+"="+secret+"\nX=1\n", 0o600),
		"wrong key":    write("key", "OTHER="+secret+"\n", 0o600),
		"empty value":  write("empty", DaemonTokenVar+"=\n", 0o600),
		"a space":      write("space", DaemonTokenVar+"="+secret+" x\n", 0o600),
		"quoted":       write("quoted", DaemonTokenVar+"=\""+secret+"\"\n", 0o600),
		"two newlines": write("nl2", DaemonTokenVar+"="+secret+"\n\n", 0o600),
	} {
		tok, err := ReadDaemonToken(path)
		if err == nil {
			t.Errorf("%s: read %q, want a refusal", name, tok)
			continue
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("%s: the refusal carries the token: %v", name, err)
		}
	}
	was := currentUID
	currentUID = func() int { return os.Getuid() + 1 }
	_, err := ReadDaemonToken(good)
	currentUID = was
	if err == nil || !strings.Contains(err.Error(), "belongs to another user") {
		t.Errorf("a file of another owner = %v, want it refused", err)
	}
	if _, err := ReadDaemonToken(sub); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("a directory = %v, want it refused as not a regular file", err)
	}
}

func TestWriteDaemonTokenModes(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gray-area")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "antigravity-daemon.env")
	if err := WriteDaemonToken(path); err != nil {
		t.Fatal(err)
	}
	first, err := ReadDaemonToken(path)
	if err != nil {
		t.Fatalf("a written token does not read back: %v", err)
	}
	if st, _ := os.Stat(dir); st.Mode().Perm() != 0o700 {
		t.Errorf("directory mode %04o, want 0700", st.Mode().Perm())
	}
	if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
		t.Errorf("file mode %04o, want 0600", st.Mode().Perm())
	}
	if err := WriteDaemonToken(path); err != nil {
		t.Fatal(err)
	}
	if second, _ := ReadDaemonToken(path); second == first || second == "" {
		t.Errorf("a second write did not rotate the token")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("the directory holds %d entries, want the token file alone", len(entries))
	}
}
