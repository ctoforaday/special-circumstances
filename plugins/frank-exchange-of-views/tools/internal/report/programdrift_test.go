package report

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// fx builds compute's fixture program (compute/testdata/fx) into the test's own directory: a
// binary, so the arm runs on every platform without an interpreter.
func fx(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "fx")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, "../compute/testdata/fx").CombinedOutput(); err != nil {
		t.Fatalf("building the fixture program: %v\n%s", err, out)
	}
	return bin
}

// THE DRIFTS THE TOOL WRITES OPEN NOTHING IN evidence.md (T1). Both drifts here are the compute
// code's own, not a test's paraphrase of one: the computation's from two real runs of a program whose
// out/ file is named by the program — a backtick, a line break before `## X`, a link — and differs
// between the runs, so the drift names both names; the act's comparison from Compare, against an
// execution whose out/ file has another hostile name and whose output holds a fence, a `## V` line,
// a `====` rule and an indented traceback. The document's headings, `<a id>` anchors, links, images,
// raw HTML, footnotes, emphasis and fences are those of the same computation with plain names.
func TestTheToolsOwnDriftsOpenNothingInTheEvidenceDocument(t *testing.T) {
	bin := fx(t)
	doc := func(t *testing.T, hostile bool) string {
		name := func(h, plain string) string {
			if hostile {
				return h
			}
			return plain
		}
		runDir := newRun(t)
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("9\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		p, err := compute.ReadProgram(dir, []string{bin, "outdepth", name(hostileOut, "o")}, func(string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		res, err := compute.Run(runDir, p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(res.Drift, "written by the first only") || !strings.Contains(res.Drift, "written by the second only") {
			t.Fatalf("the two runs' out/ files do not differ by name: drift %q", res.Drift)
		}
		if err := compute.WriteBundle(runDir, res.Bundle, res.Argv, p.Files); err != nil {
			t.Fatal(err)
		}
		const id = "K-00000001"
		if err := compute.WriteAct(runDir, res.Bundle, id, []byte(res.Output), res.Out); err != nil {
			t.Fatal(err)
		}
		rerun := compute.Execution{
			Output: name("```\n## V\n====\n  Traceback (most recent call last):\n    File \"x\"\n", "other output\n"),
			Exit:   res.Exit,
			Out:    []compute.File{{Path: name(hostileFile, "p.py"), Body: []byte("y")}},
		}
		c, err := compute.Compare(runDir, res.Bundle, id, res.Exit, rerun)
		if err != nil {
			t.Fatal(err)
		}
		if c.Matches || !strings.Contains(c.Drift, "written by the re-run only") || !strings.Contains(c.Drift, "on re-run the program wrote: ") {
			t.Fatalf("the comparison does not name the re-run's file and output: %q", c.Drift)
		}
		staged, err := compute.StageRerun(runDir, res.Bundle, rerun)
		if err != nil {
			t.Fatal(err)
		}
		if err := staged.Promote(); err != nil {
			t.Fatal(err)
		}
		k := record.Computation{
			Label: id, SHA: res.Bundle, Argv: res.Argv, Basis: "variant", Reason: "counts",
			ToolchainPath: res.ToolchainPath, ToolchainSha256: res.ToolchainSha256,
			Drift: res.Drift, Invariant: "a file is written", Executed: true,
			Rerun: &record.ComputationRerun{Rerun: recordpb.Rerun_RERUN_DIFFERS, RerunSha: staged.Sha, Drift: c.Drift},
		}
		return resolveMarks(evidenceDoc(runtest.Open(t, runDir), []record.Computation{k}, map[string]bool{id: true}))
	}
	sameShapes(t, shapes(doc(t, false)), shapes(doc(t, true)))
}

// sameShapes fails on each node kind the second document holds a different count of than the first.
func sameShapes(t *testing.T, plain, hostile map[string]int) {
	t.Helper()
	for k, n := range plain {
		if hostile[k] != n {
			t.Errorf("the program's strings change %s: %d, the plain document's %d", k, hostile[k], n)
		}
	}
	for k, n := range hostile {
		if _, ok := plain[k]; !ok {
			t.Errorf("the program's strings open %d %s", n, k)
		}
	}
}
