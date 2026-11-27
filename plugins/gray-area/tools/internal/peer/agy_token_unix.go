//go:build unix

package peer

import (
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
	if sys, ok := st.Sys().(*syscall.Stat_t); !ok || int(sys.Uid) != os.Getuid() {
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
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
