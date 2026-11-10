package migrate

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A PROOF IS A COMPUTATION (epoch 49, plans/compute-and-show.md §III.1.6). A source before epoch 49
// records a `proof` of one script, stored under `proofs/<sha of the script>/`; the record holds a
// `computation` of a bundle, stored under `computations/<bundle_sha>/`, whose identity is over the
// command as well as the files. The translation is a function of the old event alone: the archived
// script's extension names the command, and the archived sha IS the one file's sha256, so the new
// identity needs no read of the store.
//
//   - proof_id `P-`/`p-` + 8 hex → computation_id `K-` + the same hex (remap.go)
//   - proof_key → key; text (`reason` in the shard era) → text; answers, cites, exit, location kept
//   - proof_basis `reproducible` → REPRODUCIBLE; `observed` and `not_portable` → VARIANT, its
//     invariant the drift the tool recorded (F4): the archive states no invariant, and its drift is
//     the one statement the record holds of what the two runs did
//   - the author's script path is not carried: the argv names the stored file, and the path named a
//     working file nothing kept
//   - the shard era's inline output is this act's stored output (writeStore), not a field; an
//     inline output that is the shard era's excerpt of proofs/<sha>/output stands for that file

// archivedArgv is the command an archived script's extension named: the interpreter the tool ran a
// script of that extension with, and the stored file. A frozen table — it states what the archive
// ran, and the live tool names its own command.
var archivedArgv = map[string][]string{
	".py":  {"python3"},
	".sh":  {"bash"},
	".js":  {"node"},
	".mjs": {"node"},
	".go":  {"go", "run"},
}

// archivedBundle is the bundle an archived proof translates to: its argv, the one file's name in
// the bundle, and the bundle's identity.
func archivedBundle(script, sha string) (argv []string, file, bundle string, err error) {
	ext := strings.ToLower(filepath.Ext(script))
	head, ok := archivedArgv[ext]
	if !ok {
		return nil, "", "", fmt.Errorf("the proof's script %q has the extension %q, which no archived interpreter ran — the archive names a command this translation does not know", script, ext)
	}
	file = "script" + filepath.Ext(script)
	argv = append(append([]string{}, head...), file)
	return argv, file, compute.BundleSha(argv, []compute.FileSum{{Path: file, Sha256: sha}}), nil
}

// proofEntry translates an archived proof, of either era, to a computation.
func proofEntry(old OldEvent) ([]proto.Message, error) {
	out := OldEvent{ID: old.ID, SeatID: old.SeatID, Round: old.Round, TS: old.TS, Key: old.Key,
		Word: "computation", Lists: old.Lists, Arms: old.Arms, Fields: map[string]any{}}
	var sha, script, basis string
	for col, v := range old.Fields {
		switch col {
		case "proof_id":
			out.Fields["computation_id"] = v
		case "proof_key":
			out.Fields["key"] = v
		case "reason":
			out.Fields["text"] = v
		case "proof_sha", "sha256":
			sha, _ = v.(string)
		case "script":
			script, _ = v.(string)
		case "proof_basis":
			basis, _ = v.(string)
		case "output":
			// The shard era's inline output is the act's stored output (writeStore).
		default:
			out.Fields[col] = v
		}
	}
	body, err := identityBody(out)
	if err != nil {
		return nil, err
	}
	c := body.(*recordpb.Computation)
	if sha == "" || script == "" {
		return nil, fmt.Errorf("proof %s names no script and sha — the computation it translates to is the bundle they name, and without them there is none", c.GetComputationId())
	}
	argv, _, bundle, err := archivedBundle(script, sha)
	if err != nil {
		return nil, err
	}
	c.Argv, c.BundleSha = argv, proto.String(bundle)
	switch basis {
	case "reproducible":
		c.Basis = recordpb.Basis_BASIS_REPRODUCIBLE.Enum()
	case "observed", "not_portable":
		c.Basis = recordpb.Basis_BASIS_VARIANT.Enum()
		c.Invariant = proto.String(c.GetDrift())
	default:
		return nil, fmt.Errorf("proof %s records the basis %q, which names no basis a computation carries", c.GetComputationId(), basis)
	}
	return []proto.Message{c}, nil
}

// reproduceShape reads an archived reproduce's columns into the record's: the join columns take the
// computation's names, `not_run` is an act that ran nothing (its comparison is the standing
// re-run's, which replay supplies), and a re-run that differed carries what it wrote as its drift.
// The recorded output is not carried: it is the act's stored output.
func reproduceShape(col string, v any, out map[string]any) (handled bool, err error) {
	switch col {
	case "proof_sha":
		out["bundle_sha"] = v
	case "proof_id":
		out["computation_id"] = v
	case "rerun":
		w, _ := v.(string)
		if w == "not_run" {
			out["ran"] = int64(0)
			return true, nil
		}
		out["rerun"] = v
		if _, set := out["ran"]; !set {
			out["ran"] = int64(1)
		}
	case "recorded_output":
	case "observed_output":
		out[observedColumn] = v
	default:
		return false, nil
	}
	return true, nil
}

// observedColumn holds an archived re-run's output until reproduceEntry knows its comparison: the
// output of a re-run that differed is its drift, and of one that called the tool, nothing.
const observedColumn = "\x00observed_output"

// reproduceDrift is the drift an archived re-run that differed carries: what the program wrote, as
// a record string field carries a program's bytes (compute.Printable). Replay sets it after the ids
// are respelled, so the program's bytes are never respelled.
func reproduceDrift(observed string) string {
	return "on re-run the program wrote: " + compute.Printable(observed)
}

// migratedComputation is one computation replay wrote, and where its stored artifact comes from.
type migratedComputation struct {
	id, bundle string
	argv       []string
	// oldSha is the archived script's sha, its artifact under proofs/<oldSha>/; "" for a source that
	// already stores computations, whose artifact is under computations/<bundle>/.
	oldSha string
	// output is the shard era's inline output, which is this act's.
	output *string
	// exit is the act's recorded exit.
	exit int
}

// AN ARCHIVED RE-RUN THAT REPRODUCED IS THE EXECUTION THAT WROTE ITS ACT'S OUTPUT. An archived
// reproduce that names no execution compared the re-run's output byte for byte with the act's and
// recorded the result. The record names an execution by its sha (compute.Execution.Sha), over its
// output, exit and out/, and the archive supplies each: the output is the act's, which the re-run
// matched; out/ is the act's, empty for an archived proof, which wrote none; and the exit is the
// act's recorded exit, which no archived reproduce compared — the residue of this translation
// (gblock, 2026-11-10). A re-run that differed is stored only as the first 400 characters of its
// output, so it derives no execution and names none.

// derivedRerun is the execution an archived reproduce that ran and reproduced stands for: the act's
// stored output and out/ (sourceArtifact) and its recorded exit. ok is false where the source does
// not hold the act's artifact, which writeStore names as missing.
func derivedRerun(fromDir string, c migratedComputation) (e compute.Execution, ok bool) {
	_, output, out, _, err := sourceArtifact(fromDir, c)
	if err != nil {
		return compute.Execution{}, false
	}
	return compute.Execution{Output: string(output), Exit: c.exit, Out: out}, true
}

// deriveExecution names the execution an archived reproduce compared against, where it names none
// and the archive holds it, and returns the one it derived, for writeStore to store. One that ran and
// reproduced stands for its act's derived execution (derivedRerun), with the act's recorded exit as
// its rerun_exit; one that ran nothing names its bundle's standing re-run, as a live act does.
// standing is the latest reproduce written that ran each bundle.
func deriveExecution(r *recordpb.Reproduce, fromDir string, computationOf map[string]migratedComputation, standing map[string]*recordpb.Reproduce) *compute.Execution {
	if r.GetRerunSha() != "" {
		return nil
	}
	if !r.GetRan() {
		if st := standing[r.GetBundleSha()]; st != nil && st.GetRerunSha() != "" {
			r.RerunSha = proto.String(st.GetRerunSha())
		}
		return nil
	}
	c, ok := computationOf[r.GetComputationId()]
	if fromDir == "" || !ok || r.GetRerun() != recordpb.Rerun_RERUN_REPRODUCED {
		return nil
	}
	e, ok := derivedRerun(fromDir, c)
	if !ok {
		return nil
	}
	r.RerunSha, r.RerunExit = proto.String(e.Sha()), proto.Int32(int32(e.Exit))
	return &e
}

// shardExcerptCap and shardExcerptNote are how the shard era cut an inline output: the first 2000
// bytes and this line, the full output written to proofs/<sha>/output beside it. A frozen copy: it
// states what the archive wrote.
const (
	shardExcerptCap  = 2000
	shardExcerptNote = "\n… truncated; the full output is in <run>/proofs/<sha256>/output"
)

// isShardExcerpt is whether inline is the shard era's excerpt of full.
func isShardExcerpt(inline string, full []byte) bool {
	return len(full) > shardExcerptCap && inline == string(full[:shardExcerptCap])+shardExcerptNote
}

// writeStore builds the migrated run's computations/ from the source's store: each bundle's dir/
// once, and each act's own output and out/ under acts/<id>/. A source at epoch 49 or later holds
// the store's own layout, and its directories the record names are copied as they are; .tmp-*
// staging and a directory no event names are not. An artifact the source does not hold is returned by name — the
// manifest carries it — and nothing is written in its place.
func writeStore(fromDir, toDir string, comps []migratedComputation, reruns map[string][]string, derived map[string][]compute.Execution) ([]string, error) {
	var missing []string
	for _, c := range comps {
		files, output, out, at, err := sourceArtifact(fromDir, c)
		if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, fmt.Sprintf("%s: %s", c.id, at))
			continue
		}
		if err != nil {
			return nil, err
		}
		if err := compute.WriteBundle(toDir, c.bundle, c.argv, files); err != nil {
			return nil, fmt.Errorf("computation %s: %w", c.id, err)
		}
		if err := compute.WriteAct(toDir, c.bundle, c.id, output, out); err != nil {
			return nil, fmt.Errorf("computation %s: %w", c.id, err)
		}
	}
	// A derived execution is stored as a live one is (compute.StageRerun), at rerun/<its sha>/.
	for bundle, es := range derived {
		for _, e := range es {
			staged, err := compute.StageRerun(toDir, bundle, e)
			if err != nil {
				return nil, err
			}
			if err := staged.Promote(); err != nil {
				return nil, err
			}
		}
	}
	for bundle, shas := range reruns {
		for _, sha := range shas {
			from := filepath.Join(fromDir, compute.StoreDir, bundle, "rerun", sha)
			if _, err := os.Stat(from); errors.Is(err, fs.ErrNotExist) {
				missing = append(missing, fmt.Sprintf("re-run %s of bundle %s: %s", sha, bundle, from))
				continue
			}
			if err := copyStoreTree(from, filepath.Join(toDir, compute.StoreDir, bundle, "rerun", sha)); err != nil {
				return nil, err
			}
		}
	}
	return missing, nil
}

// sourceArtifact reads one computation's bundle files, act output and out/ files from the source
// run, and names where it looked. An archived proof wrote no out/: the proof's store had none.
func sourceArtifact(fromDir string, c migratedComputation) (files []compute.File, output []byte, out []compute.File, at string, err error) {
	if c.oldSha == "" {
		if files, err = compute.Bundle(fromDir, c.bundle); err != nil {
			return nil, nil, nil, compute.BundlePath(fromDir, c.bundle), fs.ErrNotExist
		}
		if output, err = compute.ActOutput(fromDir, c.bundle, c.id); err != nil {
			return nil, nil, nil, compute.ActPath(fromDir, c.bundle, c.id), fs.ErrNotExist
		}
		if out, err = compute.ActOut(fromDir, c.bundle, c.id); err != nil {
			return nil, nil, nil, "", fmt.Errorf("computation %s: %w", c.id, err)
		}
		return files, output, out, "", nil
	}
	dir := filepath.Join(fromDir, "proofs", c.oldSha)
	file := c.argv[len(c.argv)-1]
	body, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return nil, nil, nil, filepath.Join(dir, file), fs.ErrNotExist
	}
	if output, at, err = actOutput(fromDir, c); err != nil {
		return nil, nil, nil, at, err
	}
	return []compute.File{{Path: file, Body: body}}, output, nil, "", nil
}

// actOutput is an archived proof's own output, and where it looked: the shard era's inline output,
// or proofs/<sha>/output where there is none or the inline output is the file's excerpt. The file is
// rewritten by every proof of the same script, so an inline output that is not its excerpt is this
// act's own.
func actOutput(fromDir string, c migratedComputation) ([]byte, string, error) {
	if c.oldSha == "" {
		output, err := compute.ActOutput(fromDir, c.bundle, c.id)
		if err != nil {
			return nil, compute.ActPath(fromDir, c.bundle, c.id), fs.ErrNotExist
		}
		return output, "", nil
	}
	at := filepath.Join(fromDir, "proofs", c.oldSha, "output")
	full, err := os.ReadFile(at)
	if c.output != nil && (err != nil || !isShardExcerpt(*c.output, full)) {
		return []byte(*c.output), "", nil
	}
	if err != nil {
		return nil, at, fs.ErrNotExist
	}
	return full, "", nil
}

// copyStoreTree copies a directory of regular files.
func copyStoreTree(from, to string) error {
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(to, rel), 0o755)
		}
		return copyFile(p, filepath.Join(to, rel))
	})
}

// computationEpoch is the event-schema epoch the computation replaced the proof at.
const computationEpoch = 49
