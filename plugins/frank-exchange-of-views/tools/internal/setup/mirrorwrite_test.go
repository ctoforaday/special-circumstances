package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE MIRROR MUST NOT REPORT FILES IT DID NOT WRITE.
//
// MEASURED 2026-08-16, by reusing MirrorGapPatterns from the seat probe. `out` is
// <runDir>/inputs/red-gap-patterns.md, nothing here created `inputs/`, the os.WriteFile error was
// discarded, and the function returned {Written: true, Files: 55} with nothing on disk. Real runs
// were masked because BuildSkeleton makes the directory earlier in run.go — so the failure was
// reachable by any other caller and invisible to all of them.
//
// Red's ENTIRE accumulated memory ships through this one write. A caller that trusts Written has
// no way to discover the corpus never arrived, which is the same shape as everything in the corpus.
func TestTheMirrorDoesNotClaimFilesItDidNotWrite(t *testing.T) {
	corpus := t.TempDir()
	if err := os.WriteFile(filepath.Join(corpus, "a_pattern.md"), []byte("---\nname: p\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A run directory with NO inputs/ — the condition every caller outside run.go presents.
	run := t.TempDir()
	r := MirrorGapPatterns([]string{corpus}, run)

	staged := filepath.Join(run, "inputs", "red-gap-patterns.md")
	_, statErr := os.Stat(staged)
	switch {
	case r.Written && statErr != nil:
		t.Fatalf("reported Written with Files=%d and no file exists (%v) — red's memory would be absent from the run while setup reported it staged", r.Files, statErr)
	case !r.Written && statErr == nil:
		t.Fatalf("reported NOT written (%q) but the file is there — the honest failure and the silent success are swapped", r.Reason)
	case !r.Written:
		t.Fatalf("the mirror could not stage into a fresh run directory: %q. It must create inputs/ itself; every caller outside run.go arrives without it", r.Reason)
	}
	if r.Files != 1 {
		t.Errorf("Files = %d, want 1 — the count is what a caller believes about delivery", r.Files)
	}
}

// THE CLASS JOIN IS THE DELIVERY CHANNEL, AND ITS WRITE USED TO BE UNCHECKED.
//
// The engine hands a repairing seat only the patterns whose class matches the gap in front of it,
// via inputs/gap-patterns-by-class.json. That write was
// `if b, err := marshalJSON(...); err == nil { os.WriteFile(...) }` — the marshal error skipped
// the write, and the write error was discarded. The setup summary then printed
// `gap-pattern index: N class(es) -> inputs/gap-patterns-by-class.json` either way, because N is
// the length of the in-memory index and is never read back off disk.
//
// So a run with no join and a run with a perfect join printed the same line, and red would open
// the run with no patterns at all. Setup already refuses a run whose corpus is mostly
// unclassified, for exactly this reason stated in exactly these words — "red would open this run
// substantially blind while its memory directory looks full" — so a failed join must refuse too,
// or the gate guards the corpus and not the channel that carries it.
//
// The failure is induced the only way that does not depend on running as an unprivileged user:
// `inputs` is made a FILE, so MkdirAll and WriteFile cannot succeed at that path.
func TestAFailedClassJoinWriteRefusesTheRun(t *testing.T) {
	cfg, runDir := runCfg(t, reports("0.35.0"))
	mem := filepath.Join(t.TempDir(), "patterns")
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, "good.md"),
		[]byte("---\nclasses: [figure-recount-fails]\ndescription: a hook\n---\n# t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg.MemoryDir = mem

	// Stage the run directory with `inputs` occupied by a regular file.
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "inputs"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	code := Run(cfg, &out, &errb)

	if code == 0 {
		t.Fatalf("exit 0 — the class join could not be written and the run proceeded anyway.\n"+
			"stdout:\n%s\nstderr:\n%s", out.String(), errb.String())
	}
	if !strings.Contains(errb.String(), "class join") {
		t.Errorf("the refusal does not name the class join, so an operator cannot tell this from any "+
			"other setup failure:\nstderr:\n%s", errb.String())
	}
	// The summary line must not claim a delivery that did not happen.
	if strings.Contains(out.String(), "gap-pattern index:") {
		t.Errorf("setup printed the index summary for a join it failed to write — that line is the "+
			"whole reason the miss was invisible:\nstdout:\n%s", out.String())
	}
}
