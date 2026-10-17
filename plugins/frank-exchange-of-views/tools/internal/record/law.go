package record

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// LawDir is where setup stages the repository's law corpus for a run: the read-only copy the bench
// reads before ruling. Setup writes it and the ruling seats' work lists read it, through this one
// path.
func LawDir(run Run) string { return filepath.Join(run.Dir(), "inputs", "law") }

// lawLine is the ruling seats' statement of whether this run has law staged, and where.
//
// THE TOOL STATES IT SO THE SEAT NEVER PROBES FOR IT. The bench's prompt said "if inputs/law/
// exists, read it", so a bench on a run with no law listed the directory and met a failing `ls`
// (smoke 07d, event 382). Whether setup staged law is a fact of the run directory, known before any
// seat sits; the bench's and the chair's work lists say it, and the prompts name the act — read
// the law the list names.
//
// AN UNREADABLE DIRECTORY IS NOT AN EMPTY ONE. A read that fails for any reason but absence says
// NOT MEASURED, so a corpus the bench cannot see never reads as a run with none.
func lawLine(run Run) string {
	dir := LawDir(run)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return "no law is staged for this run — setup found no law corpus to mirror, so no statute or precedent binds or persuades your ruling beyond the parties' arguments, the artifacts and the leaf"
	}
	if err != nil {
		return fmt.Sprintf("whether this run has law staged is NOT MEASURED — %s cannot be read: %v — and that is not a report that it has none", dir, err)
	}
	files := 0
	for _, e := range entries {
		if !e.IsDir() {
			files++
		}
	}
	if files == 0 {
		return "no law is staged for this run — " + dir + " holds no file, so no statute or precedent binds or persuades your ruling beyond the parties' arguments, the artifacts and the leaf"
	}
	return fmt.Sprintf("this run's law is staged: %d file(s) in %s — read them before you rule", files, dir)
}
