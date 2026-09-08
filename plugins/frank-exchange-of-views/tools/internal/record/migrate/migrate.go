package migrate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Migrate is the whole act: copy the run's channels, replay its record through the current
// write path, land the manifest. The source directory is never modified; the destination
// must not already hold anything.
//
// The channel copy comes FIRST, and the archive taught why: validation consults the run's
// own staged inputs — the gap-class registry lives beside the record (records/
// class-registry.json), and a mint replayed into a run without it refuses, taking every
// later reference to its gap down in a cascade. Everything except the database file set
// copies verbatim; the record is the one thing replay rebuilds.
func Migrate(fromDir, toDir string, reg Registry, opt Options) (*Manifest, error) {
	if entries, err := os.ReadDir(toDir); err == nil && len(entries) > 0 {
		return nil, fmt.Errorf("migrate: %s is not empty — a migration writes a FRESH sibling, never into an existing run", toDir)
	}
	if err := copyChannels(fromDir, toDir); err != nil {
		return nil, err
	}
	// The record's format picks the adapter: a database if one exists, else the era of
	// per-sitting shards. Both refusable — a directory holding neither is not a run.
	var src Source
	var unclassified []string
	var discarded []DiscardedSitting
	scratch := filepath.Join(toDir, ".migrate-src")
	if _, serr := os.Stat(filepath.Join(fromDir, "records", dbSiblings[0])); serr == nil {
		sq, err := OpenSQLite(filepath.Join(fromDir, "records"), scratch)
		if err != nil {
			return nil, err
		}
		src = sq
		defer func() {
			sq.Close()
			os.RemoveAll(scratch)
		}()
	} else {
		js, err := OpenJSONL(filepath.Join(fromDir, "records"))
		if err != nil {
			return nil, err
		}
		src = js
		discarded = js.Discarded()
	}

	dst, err := record.NewRun(toDir)
	if err != nil {
		return nil, err
	}
	res, err := Replay(src, reg, dst, opt)
	if err != nil {
		return nil, err
	}
	if sq, ok := src.(*SQLiteSource); ok {
		unclassified = sq.Unclassified()
	}
	m := NewManifest(fromDir, src.Files(), unclassified, res)
	m.Discarded = discarded
	if err := m.Write(toDir); err != nil {
		return nil, err
	}
	if len(res.Refusals) > 0 {
		return m, fmt.Errorf("migrate: %d event(s) refused — the manifest at %s carries each one; "+
			"a refusal is the tool's OUTPUT, decided in review, and there is no flag that skips validation",
			len(res.Refusals), filepath.Join(toDir, "inputs", ManifestName))
	}
	return m, nil
}

// copyChannels copies the run directory verbatim, EXCEPT the record itself — the database
// file set and the era's event shards. Replay rebuilds the record, and a stale copy beside
// the fresh one would be two records claiming one run.
func copyChannels(fromDir, toDir string) error {
	skip := map[string]bool{}
	for _, n := range dbSiblings {
		skip[filepath.Join(fromDir, "records", n)] = true
	}
	return filepath.WalkDir(fromDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skip[path] {
			return nil
		}
		// The era's shards are the old record, and the new record replaces them — copied
		// beside the fresh database they would be two records claiming one run, the exact
		// shape the db-file exclusion exists for. The archive keeps the originals.
		if filepath.Dir(path) == filepath.Join(fromDir, "records") && jsonlShardName.MatchString(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(fromDir, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(toDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		if !d.Type().IsRegular() {
			// A socket or symlink in a run directory is not a channel; the copy names it
			// rather than silently flattening or following it.
			return fmt.Errorf("migrate: %s is not a regular file — a run channel copy does not know what it would mean", path)
		}
		return copyFile(path, dst)
	})
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
