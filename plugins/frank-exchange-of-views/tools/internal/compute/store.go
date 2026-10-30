package compute

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// THE STORE (plans/compute-and-show.md §III.0). A run's computations live under
// <runDir>/computations/<bundle_sha>/:
//
//	dir/                the bundle — content-addressed, written once
//	acts/<K-id>/        one per recorded act: its first run's output
//	rerun/<rerun_sha>/  content-addressed: one re-run's output
//	.tmp-*/             staging for all three; never read, never archived
//
// Every directory is staged under .tmp-* in its bundle's directory (the same filesystem) and
// renamed into place, so a reader never meets a half-written one. Every reader resolves a directory
// from the record — acts/ by computation_id, rerun/ by a reproduce's rerun_sha — never by listing,
// and a leftover nothing names is ignored rather than swept: a sweep races a concurrent writer.

// StoreDir is the store's directory under a run, relative to it.
const StoreDir = "computations"

// File is one file of a bundle, by its slash-separated path inside the bundle.
type File struct {
	Path string
	Body []byte
}

// FileSum is one file of a bundle by its path and the sha256 of its bytes, in hex — what a bundle's
// identity is over.
type FileSum struct {
	Path, Sha256 string
}

// Sums is each file's FileSum.
func Sums(files []File) []FileSum {
	out := make([]FileSum, len(files))
	for i, f := range files {
		sum := sha256.Sum256(f.Body)
		out[i] = FileSum{Path: f.Path, Sha256: hex.EncodeToString(sum[:])}
	}
	return out
}

// BundlePath is the directory a run holds a bundle's files in.
func BundlePath(runDir, bundle string) string {
	return filepath.Join(runDir, StoreDir, bundle, "dir")
}

// ActPath is the directory a run holds one recorded act's output in.
func ActPath(runDir, bundle, id string) string {
	return filepath.Join(runDir, StoreDir, bundle, "acts", id)
}

// BundleSha is a bundle's identity: sha256 over the sorted listing of its items, each item a
// sequence of length-prefixed fields (`<len>:<bytes>`) and itself length-prefixed, so a newline or a
// space inside an argument or a path forges nothing. The items are `argv <i> <arg>` for each
// argument and `file <path> <sha256>` for each file. No mode is identity: the one mode that matters,
// argv[0]'s executable bit, is set by the tool in every copy. It is computed without running
// anything, so a keyed retry tells a crash-retry from a collision before executing.
func BundleSha(argv []string, files []FileSum) string {
	var items []string
	for i, a := range argv {
		items = append(items, frame("argv", strconv.Itoa(i), a))
	}
	for _, f := range files {
		items = append(items, frame("file", f.Path, f.Sha256))
	}
	sort.Strings(items)
	h := sha256.New()
	for _, it := range items {
		fmt.Fprintf(h, "%d:%s", len(it), it)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// frame is one item: each field length-prefixed.
func frame(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&b, "%d:%s", len(f), f)
	}
	return b.String()
}

// WriteBundle puts a bundle's files at dir/, staged and renamed. A dir/ already present is another
// act's write of the same bundle, and stands only if it hashes to bundle under argv: content
// addressing makes the two writes the same bytes, and anything else is a store that lies about its
// name.
func WriteBundle(runDir, bundle string, argv []string, files []File) error {
	final := BundlePath(runDir, bundle)
	if _, err := os.Stat(final); err == nil {
		return checkBundle(runDir, bundle, argv)
	}
	stage, err := staging(runDir, bundle)
	if err != nil {
		return err
	}
	for _, f := range files {
		p := filepath.Join(stage, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			os.RemoveAll(stage)
			return err
		}
		if err := os.WriteFile(p, f.Body, 0o644); err != nil {
			os.RemoveAll(stage)
			return err
		}
	}
	if err := os.Rename(stage, final); err != nil {
		os.RemoveAll(stage)
		if _, serr := os.Stat(final); serr == nil {
			return checkBundle(runDir, bundle, argv)
		}
		return fmt.Errorf("storing bundle %s: %w", short(bundle), err)
	}
	return nil
}

// checkBundle refuses a present dir/ that does not hash to its name.
func checkBundle(runDir, bundle string, argv []string) error {
	files, err := Bundle(runDir, bundle)
	if err != nil {
		return err
	}
	if got := BundleSha(argv, Sums(files)); got != bundle {
		return fmt.Errorf("the store already holds a bundle under %s and it hashes to %s — the store's copy is not the bundle its name says, so nothing is written over it", short(bundle), short(got))
	}
	return nil
}

// WriteAct puts one act's output at acts/<id>/, staged and renamed. An id is minted fresh for every
// act, so a present directory is refused rather than overwritten.
func WriteAct(runDir, bundle, id string, output []byte) error {
	final := ActPath(runDir, bundle, id)
	if _, err := os.Stat(final); err == nil {
		return fmt.Errorf("the store already holds an act %s under bundle %s — an act's id is minted for it alone", id, short(bundle))
	}
	stage, err := staging(runDir, bundle)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stage, "output"), output, 0o644); err != nil {
		os.RemoveAll(stage)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		os.RemoveAll(stage)
		return err
	}
	if err := os.Rename(stage, final); err != nil {
		os.RemoveAll(stage)
		return fmt.Errorf("storing act %s: %w", id, err)
	}
	return nil
}

// RemoveAct takes back an act's directory whose event the record refused, so no acts/ entry is left
// that no event names.
func RemoveAct(runDir, bundle, id string) error {
	return os.RemoveAll(ActPath(runDir, bundle, id))
}

// staging is a fresh .tmp-* directory beside the bundle's final directories.
func staging(runDir, bundle string) (string, error) {
	parent := filepath.Join(runDir, StoreDir, bundle)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(parent, ".tmp-*")
}

// Bundle reads a stored bundle's files, sorted by path. A bundle the store does not hold is an
// error naming it, never an empty bundle.
func Bundle(runDir, bundle string) ([]File, error) {
	root := BundlePath(runDir, bundle)
	var files []File
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files = append(files, File{Path: filepath.ToSlash(rel), Body: body})
		return nil
	})
	if errors.Is(err, fs.ErrNotExist) || err == nil && len(files) == 0 {
		return nil, missing(bundle, "bundle", root)
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

// ActOutput reads one act's recorded output. An act the store does not hold is an error naming it.
func ActOutput(runDir, bundle, id string) ([]byte, error) {
	p := filepath.Join(ActPath(runDir, bundle, id), "output")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, missing(bundle, "recorded output of "+id, p)
	}
	return b, nil
}
