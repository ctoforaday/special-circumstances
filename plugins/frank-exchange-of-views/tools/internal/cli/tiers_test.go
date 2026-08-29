package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The 2026-08-23 shape, in miniature: a bulk seat answered by something nobody asked for, a
// judgment seat answered as configured, and a seat nothing measured.
func tierFixture(t *testing.T) (string, *record.Board) {
	t.Helper()
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "inputs", "run-config.json"),
		[]byte(`{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, run,
		recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:register:#1", &recordpb.Register{
			ToolVersion:    proto.String("test"),
			ServedModel:    proto.String("claude-opus-4-8"),
			RequestedModel: proto.String("claude-fable-5"),
		}),
		recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:register:#1", &recordpb.Register{
			ToolVersion: proto.String("test"), ServedModel: proto.String("claude-sonnet-5"),
		}),
		recordtest.At(t, "judge-r1", 1, "judge-r1:register:#1", &recordpb.Register{
			ToolVersion: proto.String("test"),
		}),
	)
	b, err := record.BoardState(runtest.Open(t, run))
	if err != nil {
		t.Fatal(err)
	}
	return run, b
}

func TestTiersJoinsTheRequestAgainstTheService(t *testing.T) {
	run, b := tierFixture(t)
	r := tierReport(runtest.Open(t, run), b)
	if r.TierBound != 3 || r.Measured != 2 || r.Substituted != 1 {
		t.Fatalf("bound/measured/substituted = %d/%d/%d, want 3/2/1", r.TierBound, r.Measured, r.Substituted)
	}
	by := map[string]TierRow{}
	for _, s := range r.Seats {
		by[s.SeatID] = s
	}
	if s := by["blue-lane-1"]; s.Matches || !s.Declared || s.Served != "claude-opus-4-8" || s.Configured != "claude-fable-5" {
		t.Errorf("the substituted bulk seat: %+v", s)
	}
	if s := by["red-merge-r1"]; !s.Matches || s.Declared {
		t.Errorf("the judgment seat was answered as configured: %+v", s)
	}
	// THE ONE THAT MATTERS. An unmeasured seat must not count as a match and must not count as a
	// substitution — it is the third state, and folding it into either is the defect.
	if s := by["judge-r1"]; s.Measured || s.Matches || s.Served != "" {
		t.Errorf("an unmeasured seat is neither a match nor a substitution: %+v", s)
	}
}

func TestTiersRendersNotMeasuredRatherThanABlank(t *testing.T) {
	run, b := tierFixture(t)
	md := renderTiers(tierReport(runtest.Open(t, run), b))
	for _, want := range []string{"NOT MEASURED", "substitution declared by the harness", "served measured on 2 of 3"} {
		if !strings.Contains(md, want) {
			t.Errorf("the rendering must carry %q; got:\n%s", want, md)
		}
	}
}

// A run where NOTHING was measured says so in its own paragraph, because the per-row NOT MEASURED
// is easy to read past when every row carries it — and that is exactly the run that reads clean.
func TestTiersSaysWhenNothingLookedAtAll(t *testing.T) {
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "inputs", "run-config.json"),
		[]byte(`{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, run, recordtest.At(t, "blue-lane-1", 1, "blue-lane-1:register:#1",
		&recordpb.Register{ToolVersion: proto.String("test")}))
	b, err := record.BoardState(runtest.Open(t, run))
	if err != nil {
		t.Fatal(err)
	}
	md := renderTiers(tierReport(runtest.Open(t, run), b))
	if !strings.Contains(md, "NOTHING LOOKED") {
		t.Errorf("a run where nothing was measured must say so as a run, not only per row; got:\n%s", md)
	}
}
