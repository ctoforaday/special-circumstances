package migrate

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Migrate is the whole act: copy the run's channels, replay its record through the current
// write path, land the manifest, and publish the result at toDir. The source directory is never
// modified.
//
// toDir must be absent, or an empty plain directory — not a file, a symlink, a junction or a mount
// point — and the record root must not be declared; each is refused before anything is created
// (destination). The run is built in a `.<base>.migrating-<random>` directory beside toDir, in a
// parent that must therefore be writable, and is moved onto toDir only once every file of it is
// synced (publish), so an interrupted migration leaves toDir as it was. An existing empty toDir is
// replaced by the built directory, not filled.
//
// The channel copy comes FIRST: validation consults the run's own inputs (inputs/run-config.json,
// the run's terms), and an act replayed into a run without them refuses, taking every later
// reference to it down in a cascade. Everything except the database file set and an archived
// class registry copies verbatim; the record is the one thing replay rebuilds.
//
// A refusal publishes the run and its manifest and returns the manifest with an error. A failure
// before the rename removes the build; a failed rename keeps it, complete and synced, and names it.
func Migrate(fromDir, toDir string, reg Registry, opt Options) (*Manifest, error) {
	to, err := destination(toDir)
	if err != nil {
		return nil, err
	}
	// --to INSIDE --from IS REFUSED. The channel copy walks the source while the build grows beside
	// --to, so a build inside the source is a directory the walk copies into itself until a path
	// is too long — and the source is modified while it runs.
	if from, err := filepath.Abs(filepath.Clean(fromDir)); err == nil {
		if rel, err := filepath.Rel(from, to); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("--to %s is inside --from %s — a migration never writes into the run it reads", to, from)
		}
	}
	b, err := newBuild(to)
	if err != nil {
		return nil, err
	}
	m, err := b.fill(fromDir, reg, opt)
	if err != nil {
		return nil, b.discard(err)
	}
	warnings, kept, err := b.publish()
	if err != nil && !kept {
		return nil, b.discard(err)
	}
	if kept {
		// The build stays where it is, so every text that reaches the operator keeps naming it.
		m.Warnings = warnings
		if len(m.Refusals) > 0 {
			return m, errors.Join(err, refusedError(len(m.Refusals), filepath.Join(b.dir, "inputs", ManifestName)))
		}
		return m, err
	}
	for _, w := range warnings {
		m.Warnings = append(m.Warnings, b.atTo(w))
	}
	if len(m.Refusals) > 0 {
		return m, refusedError(len(m.Refusals), filepath.Join(to, "inputs", ManifestName))
	}
	return m, nil
}

func refusedError(n int, manifest string) error {
	return fmt.Errorf("%d event(s) refused — the manifest at %s carries each one; "+
		"a refusal is the tool's OUTPUT, decided in review, and there is no flag that skips validation",
		n, manifest)
}

// fill builds the migrated run in the build directory and writes its manifest there.
func (b *build) fill(fromDir string, reg Registry, opt Options) (*Manifest, error) {
	toDir := b.dir
	if err := copyChannels(fromDir, toDir); err != nil {
		return nil, err
	}
	// The record's format picks the adapter: a database if one exists, else the era of
	// per-sitting shards. Both refusable — a directory holding neither is not a run.
	var src Source
	var unclassified []string
	var discarded []DiscardedSitting
	// The source's scratch copy lives INSIDE the build, and is removed before the publish. Inside,
	// so that a killed migration leaves exactly one thing behind — the build directory the help
	// tells the operator to delete — and not a copy of the source record somewhere else nobody is
	// told about (a system temp directory did that). Removed before the publish, so the published
	// run holds exactly the file set the migration writes.
	// A release that fails is an error, never ignored: a scratch copy left in the build would be
	// synced and published as part of the run.
	release := func() error { return nil }
	defer func() { _ = release() }()
	if _, serr := os.Stat(filepath.Join(fromDir, "records", dbSiblings[0])); serr == nil {
		scratch := filepath.Join(toDir, ".migrate-src")
		release = func() error { return os.RemoveAll(scratch) }
		sq, err := OpenSQLite(filepath.Join(fromDir, "records"), scratch)
		if err != nil {
			return nil, err
		}
		src = sq
		release = func() error {
			return errors.Join(sq.Close(), os.RemoveAll(scratch))
		}
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
	res, err := replayInOrder(src, reg, dst, opt)
	if err != nil {
		return nil, err
	}
	carried := 0
	var sourceEpoch *int
	if sq, ok := src.(*SQLiteSource); ok {
		unclassified = sq.Unclassified()
		if e, ok := sq.SourceEpoch(); ok {
			sourceEpoch = &e
		}
		// THE PER-TURN MEASUREMENTS ARE CARRIED, NOT REPLAYED. They are not events — nothing in the
		// translation registry addresses them — and they are not re-derivable either, because the
		// transcripts they were read from are not part of a run. See SQLiteSource.SeatTurns.
		turns, err := sq.SeatTurns()
		if err != nil {
			return nil, err
		}
		for agent, ts := range turns {
			n, err := record.AppendSeatTurns(dst, agent, ts)
			if err != nil {
				return nil, fmt.Errorf("carrying %d per-turn measurement(s) for agent %s: %w", len(ts), agent, err)
			}
			carried += n
		}
	}
	// The source is released before the publish, not at return: nothing of it may be open or on
	// disk when the run is synced and moved.
	if err := release(); err != nil {
		release = func() error { return nil }
		return nil, fmt.Errorf("removing the source's scratch copy from the run before it is published: %w", err)
	}
	release = func() error { return nil }
	m := NewManifest(fromDir, src.Files(), unclassified, res)
	m.Discarded = discarded
	m.SeatTurns = carried
	m.SourceEpoch = sourceEpoch
	// The refusals describe the run that is, or would have been, published at --to, so their text
	// names --to wherever it named the build — in the struct, so the manifest written here and the one
	// Migrate returns say the same. A kept build's own manifest therefore names --to too.
	for i := range m.Refusals {
		m.Refusals[i].Err = b.atTo(m.Refusals[i].Err)
	}
	if err := m.Write(toDir); err != nil {
		return nil, err
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
	// AN ARCHIVED RUN'S STAGED CLASS REGISTRY IS NOT A CHANNEL: a gap's class is a value of the
	// closed GapClass set, and nothing reads the file.
	skip[filepath.Join(fromDir, "records", "class-registry.json")] = true
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
		if filepath.Dir(path) == filepath.Join(fromDir, "records") && !d.IsDir() && isShard(d.Name()) {
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
			return fmt.Errorf("%s is not a regular file — a run channel copy does not know what it would mean", path)
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
