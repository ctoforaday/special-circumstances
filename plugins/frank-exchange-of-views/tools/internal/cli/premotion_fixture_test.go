package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// THE PRE-MOTION COMPAT FIXTURE — a record carrying the vocabulary #344 retires.
//
// WHY THIS EXISTS AND WHY IT RUNS ONCE. Stage 4 collapses `blue dispute`,
// `merge dispute-respond`, `<seat> petition`, `bench petition-rule` and `merge avenue-rule` into
// one `motion` group. Every consumer of those types is a bare `switch e.Type` with no default, so
// a record written BEFORE the collapse would render an empty section and `0 filed / 0 ruled` —
// indistinguishable from a run that genuinely had no disputes. That is the plausible zero, and
// the dual-read exists to prevent it.
//
// Nothing could prove the dual-read works, because NO STORED RUN CARRIES THESE TYPES. Measured
// across all of `research/` before stage 4: dispute=0, dispute-respond=0, petition=0,
// petition-rule=0, avenue-rule=0. The corpus this project has accumulated cannot exercise the
// change, and a replay of it would produce a byte-identical report whose clean diff read as
// confirmation.
//
// So the fixture is PRODUCED, and it is produced NOW — before the verbs are deleted, because
// afterwards it is not producible at all and no later care recreates it. That ordering is the one
// thing no gate can recover.
//
// This generator is EXPECTED TO STOP COMPILING at stage 4, when the verbs it drives are gone.
// That is not a defect to fix by keeping the verbs alive in a test: delete the generator, keep
// its output. The fixture is the artifact; this is only the thing that made it.
//
//	FEOV_WRITE_PREMOTION_FIXTURE=1 go test ./internal/cli -run TestWritePreMotionFixture -count=1
func TestWritePreMotionFixture(t *testing.T) {
	if os.Getenv("FEOV_WRITE_PREMOTION_FIXTURE") == "" {
		t.Skip("generator; set FEOV_WRITE_PREMOTION_FIXTURE=1 to regenerate (impossible after #344)")
	}
	runDir := seatRun(t)
	gap := mintGap(t, runDir, "premotion-gap", "compat-fixture")

	// 1. A GRADE DISPUTE and its answer — blue contests a grade, red rules on it.
	if _, err := run(t, "blue", "dispute", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", gap, "--dimension", "severity", "--proposed", "low",
		"--reason", "the consequence is bounded by the caller's own validation"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "merge", "dispute-respond", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", gap, "--dimension", "severity", "--as", "rejected",
		"--reason", "the bound does not hold across the retry path"); err != nil {
		t.Fatal(err)
	}

	// 2. A PETITION and its ruling — the exchange whose filing and ruling #315 found
	//    rendering separately, and which #312 cannot join for want of an id.
	if _, err := run(t, "merge", "petition", "--run", runDir, "--seat-id", "red-merge-r1",
		"--petition-class", "integrity", "--relief", "strike the demand from the docket",
		"--reason", "the instruction would require asserting what I believe false"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "bench", "petition-rule", "--run", runDir, "--seat-id", "judge-r1",
		"--petitioner", "red-merge-r1", "--petition-class", "integrity", "--as", "granted",
		"--reason", "the objection is sound; the demand is struck"); err != nil {
		t.Fatal(err)
	}

	// 3. A DIRECTION and red's ruling on it, plus blue pursuing AGAINST that ruling —
	//    the `contests_ruling` field that becomes `motion direction appeal`.
	out, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--line", "survey the adjacent literature", "--hypothesis", "a prior art search closes two gaps",
		"--status", "proposed", "--reason", "worth one round's budget")
	if err != nil {
		t.Fatal(err)
	}
	avenueID := firstAvenueID(out)
	if avenueID == "" {
		t.Fatalf("no avenue id in %q", out)
	}
	if _, err := run(t, "merge", "avenue-rule", "--run", runDir, "--seat-id", "red-merge-r1",
		"--id", avenueID, "--ruling", "out-of-scope",
		"--reason", "a real question, but not THIS run's"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "blue", "avenue", "--run", runDir, "--seat-id", "blue-respond-r1",
		"--id", avenueID, "--status", "pursued",
		"--reason", "taking it anyway: the adjacent result is load-bearing for R1-1"); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join("..", "record", "testdata", "pre-motion-run")
	if err := os.RemoveAll(dest); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(filepath.Join(runDir, "records"), filepath.Join(dest, "records")); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote the pre-motion fixture to %s", dest)
}

// firstAvenueID pulls the tool-assigned avenue id (A1, A2 ...) out of a proposal's output.
func firstAvenueID(out string) string {
	return regexp.MustCompile(`A\d+`).FindString(out)
}

func copyTree(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(src, e.Name()))
		if rerr != nil {
			return rerr
		}
		if werr := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o644); werr != nil {
			return werr
		}
	}
	return nil
}
