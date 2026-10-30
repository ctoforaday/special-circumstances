package compute

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

// MaxBundle bounds a bundle's bytes. A bundle is the program and what it reads, committed with the
// run's archive; data larger than this is a source to cite, not a file to carry.
const MaxBundle = 1 << 20

// The tool's mounts in every copy: out/ is where a program writes files the act keeps, sources/ is
// where the tool places cited inputs. A bundle entry under either name would collide with them.
var mounts = map[string]bool{"out": true, "sources": true}

// Program is a bundle a seat named and the command that runs it, read without running anything.
type Program struct {
	Argv  []string
	Files []File
	// Sha is the bundle's identity (BundleSha) over Argv and Files.
	Sha string
}

// ReadProgram reads the bundle at dir for argv and states its identity, without running it, so a
// keyed retry tells a crash-retry from a collision before it executes anything.
//
// refuse is called with dir and then with every file's path BEFORE the file is opened: the caller
// refuses a path that is the run's own record, whose locks a read in this process would release.
//
// EVERY REFUSAL HERE IS ABOUT WHAT THE BUNDLE IS, and each is one a copy could not honour:
//   - a symlink reaches outside the bundle, so the copy would not hold what it names;
//   - a non-regular file (a device, a pipe) has no bytes to hash or copy;
//   - over MaxBundle, the bundle is data that belongs in a citation;
//   - `out` or `sources` at the top collides with the tool's mounts;
//   - `__pycache__` or a `.pyc` is an interpreter's cache, which changes between attempts and so
//     changes the sha while the program does not.
func ReadProgram(dir string, argv []string, refuse func(path string) error) (*Program, error) {
	if len(argv) == 0 || argv[0] == "" {
		return nil, feov.Errorf(feov.MissingField, "requires the command after `--`: the program and its arguments, as `-- python3 count.py` or `-- ./run.sh` — a bundle with no command runs nothing")
	}
	if err := refuse(dir); err != nil {
		return nil, err
	}
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, feov.Errorf(feov.Validation, "cannot read the bundle %s: %v", dir, err)
	}
	if !fi.IsDir() {
		return nil, feov.Errorf(feov.Validation, "%s is a file, and --dir names a DIRECTORY: put the program in a directory of its own (a one-file computation is a one-file directory) and name that", dir)
	}
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, feov.Errorf(feov.Validation, "cannot read the bundle %s: %v", dir, err)
	}
	var paths []string
	var size int64
	walk := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == root {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		switch {
		case mounts[rel]:
			return feov.Errorf(feov.Validation, "the bundle holds %s at its top, and the tool mounts its own %s/ in every copy — rename it", rel, rel)
		case d.Name() == "__pycache__" || strings.EqualFold(filepath.Ext(d.Name()), ".pyc"):
			return feov.Errorf(feov.Validation, "the bundle holds %s, an interpreter's cache: it changes between attempts while the program does not, so the bundle's sha would too. Delete it (or set PYTHONDONTWRITEBYTECODE=1 while you draft) and name the directory again", rel)
		case d.Type()&fs.ModeSymlink != 0:
			return feov.Errorf(feov.Validation, "the bundle holds %s, a symlink: a copy of the bundle holds what the bundle holds, and a symlink names something outside it. Copy the file into the bundle", rel)
		case d.IsDir():
			return nil
		case !d.Type().IsRegular():
			return feov.Errorf(feov.Validation, "the bundle holds %s, which is not a regular file — a bundle holds files a copy can carry", rel)
		}
		if err := refuse(p); err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if size += info.Size(); size > MaxBundle {
			return feov.Errorf(feov.Validation, "the bundle at %s is over %d bytes: a bundle is the program and the small files it reads, committed with the run. Data this size is a source — cite it, and compute over what the citation holds", dir, MaxBundle)
		}
		paths = append(paths, rel)
		return nil
	})
	if walk != nil {
		var fe *feov.Error
		if errors.As(walk, &fe) {
			return nil, walk
		}
		return nil, feov.Errorf(feov.Validation, "cannot read the bundle %s: %v", dir, walk)
	}
	if len(paths) == 0 {
		return nil, feov.Errorf(feov.Validation, "the bundle %s holds no file — a bundle is the program a reader re-runs, so it holds at least the program or the input it runs over", dir)
	}
	sort.Strings(paths)
	files := make([]File, 0, len(paths))
	for _, rel := range paths {
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, feov.Errorf(feov.Validation, "cannot read %s in the bundle: %v", rel, err)
		}
		files = append(files, File{Path: rel, Body: body})
	}
	if rel, ok := bundled(argv[0]); ok && !holds(files, rel) {
		return nil, feov.Errorf(feov.Validation, "%s names a file by a relative path, and the bundle holds no %s: a relative command runs from the bundle's copy, so it must be a file in the bundle. Name a program on PATH bare (`python3`), or one elsewhere by its absolute path", argv[0], rel)
	}
	argv = append([]string{}, argv...)
	return &Program{Argv: argv, Files: files, Sha: BundleSha(argv, Sums(files))}, nil
}

// bundled reports whether argv[0] names a bundle file — relative, with a separator, as `./run.sh` —
// and that file's slash-separated path inside the bundle.
func bundled(arg0 string) (string, bool) {
	if filepath.IsAbs(arg0) || !strings.ContainsAny(arg0, "/"+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(filepath.Clean(arg0)), true
}

func holds(files []File, rel string) bool {
	for _, f := range files {
		if f.Path == rel {
			return true
		}
	}
	return false
}

// String names a program by its command, for messages.
func (p *Program) String() string { return fmt.Sprintf("`%s`", strings.Join(p.Argv, " ")) }
