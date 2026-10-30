// Package archive writes and reads a run's archive: the gzipped tarball under run-archive/ that is
// the only part of a run that outlives its container.
//
// THE STORE IS ARCHIVED AS THE RECORD NAMES IT. A run's computations/ holds the directories its
// events name — each bundle's dir/, each act's acts/<K-id>/, each executed re-run's
// rerun/<rerun_sha>/ — and beside them staging (.tmp-*) and directories no event names: a re-run
// whose append failed after its rename, or an act whose event the record refused before it was taken
// back. Those are ignored by every reader, and an archive that walked the directory would carry them
// as if they were evidence, so Store lists only what the record names.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Entry is one file an archive holds: its slash-separated name in the tarball, and either the path
// its bytes are read from or the bytes themselves.
type Entry struct {
	Name string
	Path string
	Data []byte
}

// Write writes entries to out as a gzipped tarball, sorted by name, every file 0644: no mode is
// part of what an archive keeps.
func Write(out string, entries []Entry) (err error) {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	sorted := append([]Entry{}, entries...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for _, e := range sorted {
		b := e.Data
		if b == nil {
			if b, err = os.ReadFile(e.Path); err != nil {
				return err
			}
		}
		if err := tw.WriteHeader(&tar.Header{Name: e.Name, Mode: 0o644, Size: int64(len(b))}); err != nil {
			return err
		}
		if _, err := tw.Write(b); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// Extract unpacks a tarball Write wrote into dir. A name that is absolute or climbs out of dir is
// refused, and nothing but regular files is unpacked.
func Extract(tarball, dir string) (err error) {
	f, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("%s is not a gzipped archive: %w", tarball, err)
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		name := path.Clean(h.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") {
			return fmt.Errorf("%s holds %q, which is outside the directory it unpacks into", tarball, h.Name)
		}
		to := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return err
		}
		w, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		_, cerr := io.Copy(w, tr)
		if err := errors.Join(cerr, w.Close()); err != nil {
			return err
		}
	}
}

// Store is the run's computation store as its record names it, each file an Entry under
// computations/: every bundle's dir/, every act's acts/<K-id>/, and every executed re-run's
// rerun/<rerun_sha>/ — never a .tmp-* staging or a directory no event names. A directory the record
// names and the store lacks is left out: the archive keeps what the run holds, and a reader of it
// meets the absence the way a reader of the run does.
func Store(run record.Run) ([]Entry, error) {
	root := filepath.Join(run.Dir(), compute.StoreDir)
	if _, err := os.Stat(root); errors.Is(err, fs.ErrNotExist) {
		return nil, nil // a run that computed nothing has no store, and its record is not read for one
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		return nil, err
	}
	var dirs []string
	seen := map[string]bool{}
	name := func(rel string) {
		if !seen[rel] {
			seen[rel] = true
			dirs = append(dirs, rel)
		}
	}
	for _, e := range fam.Events {
		switch b := recordBody(e).(type) {
		case *recordpb.Computation:
			if b.GetBundleSha() == "" {
				continue
			}
			name(path.Join(b.GetBundleSha(), "dir"))
			if b.GetComputationId() != "" {
				name(path.Join(b.GetBundleSha(), "acts", b.GetComputationId()))
			}
		case *recordpb.Reproduce:
			if b.GetBundleSha() != "" && b.GetRerunSha() != "" {
				name(path.Join(b.GetBundleSha(), "rerun", b.GetRerunSha()))
			}
		}
	}
	var out []Entry
	for _, rel := range dirs {
		dir := filepath.Join(root, filepath.FromSlash(rel))
		err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.Type().IsRegular() {
				return nil
			}
			r, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			out = append(out, Entry{Name: compute.StoreDir + "/" + filepath.ToSlash(r), Path: p})
			return nil
		})
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// recordBody is an event's body, or nil.
func recordBody(e *record.Event) any {
	b, ok := recordpb.Body(e)
	if !ok {
		return nil
	}
	return b
}
