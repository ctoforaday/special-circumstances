//go:build unix

package peer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// ReadDaemonToken reads the daemon's token from path. It opens the file without following a
// symlink and checks the open descriptor, refusing a file that is not regular, has any group or
// other permission bit, or belongs to another user. An absent file's error wraps fs.ErrNotExist.
// No error carries the token.
func ReadDaemonToken(path string) (string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", path)
	}
	if st.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("%s is readable or writable by group or others (mode %04o); it must be 0600", path, st.Mode().Perm())
	}
	if sys, ok := st.Sys().(*syscall.Stat_t); !ok || int(sys.Uid) != currentUID() {
		return "", fmt.Errorf("%s belongs to another user", path)
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(len(DaemonTokenVar))+maxTokenLen+3))
	if err != nil {
		return "", err
	}
	tok, err := parseDaemonToken(b)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return tok, nil
}

// currentUID is the caller's uid, against which the token file's owner is checked; a variable so a
// test can stand in another user without being root.
var currentUID = os.Getuid

// WriteDaemonToken writes a fresh random token to path (directory 0700, file 0600), replacing any
// token there by rename.
func WriteDaemonToken(path string) error {
	tok, err := newID()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".antigravity-daemon-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.WriteString(DaemonTokenVar + "=" + tok + "\n"); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// maxTokenLen bounds the token's value.
const maxTokenLen = 256

// parseDaemonToken reads the file's content: exactly one line DaemonTokenVar=<value> with at most
// one trailing newline, the value 1–maxTokenLen bytes of [A-Za-z0-9-]. No error carries the value.
func parseDaemonToken(b []byte) (string, error) {
	s := string(b)
	if n := len(s); n > 0 && s[n-1] == '\n' {
		s = s[:n-1]
	}
	prefix := DaemonTokenVar + "="
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return "", fmt.Errorf("it is not one line %s=<token>", DaemonTokenVar)
	}
	v := s[len(prefix):]
	if v == "" || len(v) > maxTokenLen {
		return "", fmt.Errorf("its token is empty or longer than %d bytes", maxTokenLen)
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return "", errors.New("its token holds a byte other than a letter, a digit or '-'")
		}
	}
	return v, nil
}
