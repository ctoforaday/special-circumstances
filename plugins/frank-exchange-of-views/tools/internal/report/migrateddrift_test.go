package report_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// This file is the external test package because migrate imports report.

// A MIGRATED DRIFT OPENS NOTHING IN evidence.md (T1). quadratic-formula's six archived `differs`
// migrate to DIFFERS whose drift carries up to 400 bytes of what the program wrote on re-run, `====`
// rules and indented tracebacks among them, unrespelled. Its evidence document parses to the same
// headings, `<a id>` anchors, links, images, raw HTML, footnotes, emphasis and fences as the same
// computations with every drift a plain word.
func TestAMigratedDriftOpensNothingInTheEvidenceDocument(t *testing.T) {
	to := filepath.Join(recordtest.TmpRun(t), "migrated")
	if _, err := migrate.Migrate(recordtest.ExtractArchive(t, "2026-09-02_quadratic-formula.tar.gz"), to, migrate.Entries(), migrate.Options{}); err != nil {
		t.Fatal(err)
	}
	run := runtest.Open(t, to)
	computations, err := record.RecordedComputations(run)
	if err != nil {
		t.Fatal(err)
	}
	bare := make([]record.Computation, len(computations))
	drifts, setext := 0, 0
	for i, k := range computations {
		bare[i] = k
		if k.Drift != "" {
			bare[i].Drift = "moved"
		}
		if k.Rerun != nil && k.Rerun.Rerun == recordpb.Rerun_RERUN_DIFFERS && k.Rerun.Drift != "" {
			drifts++
			if strings.Contains(k.Rerun.Drift, "\n====") {
				setext++
			}
			r := *k.Rerun
			r.Drift = "differs"
			bare[i].Rerun = &r
		}
	}
	// A line of text over a `====` line is a heading wherever it is spliced raw: the case measured.
	if drifts == 0 || setext == 0 {
		t.Fatalf("the migrated record holds %d DIFFERS with a drift, %d of them with a `====` rule; this measures nothing", drifts, setext)
	}
	report.SameShapes(t, report.Shapes(report.ResolveMarks(report.EvidenceDoc(run, bare))), report.Shapes(report.ResolveMarks(report.EvidenceDoc(run, computations))))
}
