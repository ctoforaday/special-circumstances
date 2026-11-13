package migrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
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
// reference to it down in a cascade. Everything except the database file set, an archived class
// registry, the computation store and inputs/run-config.json copies verbatim; the record is what
// replay rebuilds, and the store is rebuilt from it. The run's configuration is translated
// (translateRunConfig): the keys it no longer declares are dropped and its epoch is the migrated
// record's, and the manifest's run_config records both with the values the source wrote.
//
// A refusal publishes the run and its manifest and returns the manifest with an error. A failure
// before the rename removes the build; a failed rename keeps it, complete and synced, and names it.
func Migrate(fromDir, toDir string, reg Registry, opt Options) (*migration.Manifest, error) {
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
			return m, errors.Join(err, refusedError(len(m.Refusals), filepath.Join(b.dir, "inputs", migration.ManifestName)))
		}
		return m, err
	}
	for _, w := range warnings {
		m.Warnings = append(m.Warnings, b.atTo(w))
	}
	if len(m.Refusals) > 0 {
		return m, refusedError(len(m.Refusals), filepath.Join(to, "inputs", migration.ManifestName))
	}
	return m, nil
}

func refusedError(n int, manifest string) error {
	return fmt.Errorf("%d event(s) refused — the manifest at %s carries each one; "+
		"a refusal is the tool's OUTPUT, decided in review, and there is no flag that skips validation",
		n, manifest)
}

// fill builds the migrated run in the build directory and writes its manifest there.
func (b *build) fill(fromDir string, reg Registry, opt Options) (*migration.Manifest, error) {
	toDir := b.dir
	runConfig, err := copyChannels(fromDir, toDir)
	if err != nil {
		return nil, err
	}
	// The record's format picks the adapter: a database if one exists, else the era of
	// per-sitting shards. Both refusable — a directory holding neither is not a run.
	var src Source
	var unclassified []string
	var discarded []migration.DiscardedSitting
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
	res, err := replayInOrder(src, fromDir, reg, dst, opt)
	if err != nil {
		return nil, err
	}
	// THE STORE IS BUILT FROM THE RECORD replay wrote: each computation's bundle and its own output,
	// and each re-run a reproduce names (computations.go).
	missing, err := writeStore(fromDir, toDir, res.computations, res.reruns, res.derived)
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
	// A CAPTURED RUN STAYS CAPTURED, and its mark lands LAST. copyChannels leaves the mark behind
	// because the build's record refuses every write once it holds one (record.RunCaptured), the
	// replay's included; written after the record's last write, it says of the migrated run what it
	// said of the source. The source's bytes, so the capture's time is the one it states.
	if err := carryCapturedMark(fromDir, toDir); err != nil {
		return nil, err
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
	m.MissingArtifacts = missing
	m.SeatTurns = carried
	m.SourceEpoch = sourceEpoch
	m.RunConfig = runConfig
	if err := m.Write(toDir); err != nil {
		return nil, err
	}
	return m, nil
}

// carryCapturedMark writes the source's capture mark into the build, when the source holds one.
func carryCapturedMark(fromDir, toDir string) error {
	from := filepath.Join(fromDir, runlive.CapturedMark)
	if _, err := os.Stat(from); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("asking whether the source run was captured: %w", err)
	}
	if err := copyFile(from, filepath.Join(toDir, runlive.CapturedMark)); err != nil {
		return fmt.Errorf("carrying the capture mark into the migrated run: %w", err)
	}
	return nil
}

// copyChannels copies the run directory verbatim, EXCEPT the record itself — the database
// file set and the era's event shards. Replay rebuilds the record, and a stale copy beside
// the fresh one would be two records claiming one run. The capture mark waits for the replay
// (carryCapturedMark): beside the record before it, it refuses every event replay appends. The run's
// configuration is translated, not copied, and what the translation did is returned — nil when the
// source holds no configuration.
func copyChannels(fromDir, toDir string) (*migration.RunConfigTranslation, error) {
	var runConfig *migration.RunConfigTranslation
	skip := map[string]bool{filepath.Join(fromDir, runlive.CapturedMark): true}
	for _, n := range dbSiblings {
		skip[filepath.Join(fromDir, "records", n)] = true
	}
	// AN ARCHIVED RUN'S STAGED CLASS REGISTRY IS NOT A CHANNEL: a gap's class is a value of the
	// closed GapClass set, and nothing reads the file.
	skip[filepath.Join(fromDir, "records", "class-registry.json")] = true
	// THE COMPUTATION STORE IS REBUILT FROM THE RECORD (writeStore): an archived proofs/ is the old
	// layout, and computations/ holds staging and directories no event names, which are not copied.
	skipTree := map[string]bool{filepath.Join(fromDir, "proofs"): true, filepath.Join(fromDir, compute.StoreDir): true}
	err := filepath.WalkDir(fromDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if skip[path] {
			return nil
		}
		if skipTree[path] && d.IsDir() {
			return filepath.SkipDir
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
		if rel == runConfigPath {
			tr, err := translateRunConfig(path, dst)
			runConfig = tr
			return err
		}
		return copyFile(path, dst)
	})
	if err != nil {
		return nil, err
	}
	return runConfig, nil
}

// runConfigPath is the run's terms file, relative to the run directory.
var runConfigPath = filepath.Join("inputs", "run-config.json")

// runConfigDrops are the keys of an archived run-config.json the run's configuration no longer
// declares. The record's reader of the file refuses a key it does not declare, so a key carried
// verbatim would refuse the migrated run's every read of its configuration. Why each key went is
// said here, beside it; the manifest's run_config records each key dropped with its value.
var runConfigDrops = map[string]struct{}{
	// A SOURCE BEFORE EPOCH 74 MAY STATE `maxSittingCalls`, the per-sitting tool-call limit, which
	// went at epoch 74: the agent definition's maxTurns bounds a sitting.
	"maxSittingCalls": {},
	// A SOURCE BEFORE EPOCH 75 MAY STATE `mintBudget`, the per-lens mint cap, which went at epoch 75:
	// lens retirement and the epoch limit bound a run.
	"mintBudget": {},
}

// translateRunConfig copies the run-config.json at from to to without the keys runConfigDrops
// names, every other key in its place and with its value as written — except `eventSchema`, which
// states the epoch the migrated record is written at (record.EventSchema), added at the end where the
// source states none: the record beside the file is at that epoch, and a file still naming the
// source's would say the run was set up under a shape it is no longer in. The source must be ONE
// JSON object and nothing after it (record.RunConfigObject), and the copy must read through
// record.DecodeRunConfig, the record's own reader, or nothing is written. What it did — each key
// dropped and the epoch the source stated, with the values the source wrote — is returned for the
// manifest.
func translateRunConfig(from, to string) (*migration.RunConfigTranslation, error) {
	b, err := os.ReadFile(from)
	if err != nil {
		return nil, err
	}
	raw, err := record.RunConfigObject(b)
	if err != nil {
		return nil, fmt.Errorf("%s %w", from, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil { // the object's `{`, checked above
		return nil, fmt.Errorf("%s: %w", from, err)
	}
	tr := &migration.RunConfigTranslation{Dropped: map[string]json.RawMessage{}}
	var out bytes.Buffer
	out.WriteByte('{')
	first := true
	put := func(key string, val []byte) error {
		if !first {
			out.WriteByte(',')
		}
		first = false
		k, err := json.Marshal(key)
		if err != nil {
			return err
		}
		out.Write(k)
		out.WriteByte(':')
		out.Write(val)
		return nil
	}
	epoch := []byte(strconv.Itoa(record.EventSchema))
	statedEpoch := false
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", from, err)
		}
		key, _ := tok.(string)
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, fmt.Errorf("%s: key %q: %w", from, key, err)
		}
		if _, drop := runConfigDrops[key]; drop {
			tr.Dropped[key] = val
			continue
		}
		if key == runConfigEpochKey {
			tr.SourceEventSchema = val
			val, statedEpoch = epoch, true
		}
		if err := put(key, val); err != nil {
			return nil, err
		}
	}
	if !statedEpoch {
		if err := put(runConfigEpochKey, epoch); err != nil {
			return nil, err
		}
	}
	out.WriteByte('}')
	var indented bytes.Buffer
	if err := json.Indent(&indented, out.Bytes(), "", "  "); err != nil {
		return nil, fmt.Errorf("%s: %w", from, err)
	}
	indented.WriteByte('\n')
	// THE COPY IS HELD TO THE RECORD'S OWN READER: a key no drop names, or a value of the wrong
	// type, would publish a run whose every read of its configuration is refused.
	if _, err := record.DecodeRunConfig(indented.Bytes()); err != nil {
		return nil, fmt.Errorf("%s: the migrated configuration %w", from, err)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(to, indented.Bytes(), 0o644); err != nil {
		return nil, err
	}
	return tr, nil
}

// runConfigEpochKey is record.RunConfig.EventSchema's key in the file.
const runConfigEpochKey = "eventSchema"

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
