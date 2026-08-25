package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// sandbox copies a module into a scratch directory so the sweep never writes to the tree
// anyone else is using.
//
// # Why a copy rather than a careful writer
//
// This tool's defining act is writing a deliberate defect into source and undoing it, and
// every hazard it has ever had comes from doing that IN PLACE. The undo can fail; a signal
// can land between choosing a file and recording how to restore it; and — the one that
// actually happens — a second agent or a human edits the tree while a sweep is running, so
// the "restore" writes a stale copy over somebody's live work. Hardening the undo answers
// the first two and cannot answer the third, because the third is not a bug in the undo. It
// is a consequence of sharing the working tree with other writers.
//
// A sweep in its own copy has none of them. An interrupt abandons a scratch directory; a
// failed write damages a scratch directory; a concurrent edit is simply not visible to it.
// Measured: the largest module here is 5.3 MB and copies in 0.06s, against a sweep that runs
// for minutes.
//
// # Why a copy and not `git worktree`
//
// A worktree carries COMMITTED state. Mutation testing is most useful on work in progress —
// the tests you just wrote, against the code you just wrote — and a worktree would silently
// measure something else. The copy takes the tree as it stands, uncommitted changes and all.
//
// # What the sweep still has to do
//
// Restore each file after mutating it, which stays for a reason that has nothing to do with
// safety: within one sweep, a file left mutated while the NEXT file is mutated means every
// later run carries two defects, and the measurement is silently wrong. The sandbox removes
// the danger, not that step.
func sandbox(moduleDir string) (dir string, cleanup func(), err error) {
	abs, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", nil, err
	}
	if _, err := os.Stat(filepath.Join(abs, "go.mod")); err != nil {
		return "", nil, fmt.Errorf("%s holds no go.mod, so it is not a module to sweep: %w", moduleDir, err)
	}
	tmp, err := os.MkdirTemp("", "sc-mutate-")
	if err != nil {
		return "", nil, err
	}
	dst := filepath.Join(tmp, filepath.Base(abs))
	if err := copyTree(abs, dst); err != nil {
		os.RemoveAll(tmp)
		return "", nil, fmt.Errorf("copying %s into a scratch directory: %w", moduleDir, err)
	}
	return dst, func() { os.RemoveAll(tmp) }, nil
}

// copyTree copies a directory recursively, preserving the executable bit.
//
// Symlinks are followed rather than recreated: a module here has none, and a copier that
// silently skipped one would produce a sandbox that is not the module it claims to be.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			// Resolve it: what matters is that the sandbox compiles like the original.
			real, err := filepath.EvalSymlinks(p)
			if err != nil {
				return err
			}
			info, err = os.Stat(real)
			if err != nil {
				return err
			}
			if info.IsDir() {
				return copyTree(real, target)
			}
			p = real
		}
		return copyFile(p, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// restoreFile puts a file back to the bytes it held before the sweep mutated it.
//
// Inside a sandbox this is about the MEASUREMENT, not about anyone's working tree: a file
// left mutated while the next one is mutated makes every later run carry two defects, and
// the survivors reported after that point are about a program nobody wrote. So the error is
// returned rather than discarded — it used to be `_ = os.WriteFile(...)`, under a handler
// that printed "interrupted — file restored" whether or not it had been.
func restoreFile(path string, body []byte) error {
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("restoring %s inside the sandbox: %w — every later mutant in this "+
			"sweep would be measured against a file that still carries this one", relOrPath(path), err)
	}
	return nil
}

// relOrPath shortens a sandbox path for a message, since the temp prefix is noise.
func relOrPath(p string) string {
	if i := strings.Index(p, "sc-mutate-"); i >= 0 {
		if j := strings.IndexByte(p[i:], filepath.Separator); j >= 0 {
			return p[i+j+1:]
		}
	}
	return p
}
