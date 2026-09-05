package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// seedRun is a run directory with a record and one gap on the board.
func seedRun(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	recordtest.Seed(t, dir, recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-1", &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Problem:         proto.String("p"),
		RequiredFix:     proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"),
		Class:           proto.String("self-attestation"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	return dir
}

// THE TWO QUESTIONS THE WHOLE DESIGN TURNS ON, ASKED OF THE REAL DATABASE.
//
// The record runs journal_mode=WAL, which makes both answers non-obvious and both failures
// silent. A committed write lands in record.db-wal and may never touch record.db until a
// checkpoint, so a fingerprint over record.db alone freezes the dashboard exactly when the run
// is busiest. And a read must leave the digest alone, or the watcher re-renders every tick on
// the strength of its own read.
//
// THE READ ARM ALSO DISPROVED A THEORY, which is why it is written as a question rather than an
// assertion about -shm. The first draft excluded record.db-shm on the reasoning that a WAL reader
// touches the shared-memory index; measured, it moves in neither size nor mtime across a full
// BuildModel, so nothing is excluded on that account and this arm is what would catch it if that
// ever changed. Neither answer can be settled by reading the driver's documentation.
func TestFingerprintSeesAWriteAndIgnoresOurOwnRead(t *testing.T) {
	dir := seedRun(t)
	transcripts := t.TempDir()

	before := Fingerprint(dir, transcripts)
	if before == "" {
		t.Fatal("a seeded run fingerprinted as unmeasurable")
	}

	// A READ must not move it, or the watcher never skips anything.
	BuildModel(runtest.Open(t, dir), transcripts, Config{}, 0)
	if after := Fingerprint(dir, transcripts); after != before {
		t.Error("reading the record changed the fingerprint — the watcher would re-render every " +
			"tick on the strength of its own read, which is the recompute this exists to stop")
	}

	// A WRITE must move it, or the dashboard silently stops updating.
	recordtest.Seed(t, dir, recordtest.At(t, "red-merge-r1", 1, "red-merge-r1:mint:R1-2", &recordpb.Mint{
		GapId:           proto.String("R1-2"),
		Problem:         proto.String("p2"),
		RequiredFix:     proto.String("f2"),
		AcceptanceCheck: proto.String("the check runs"),
		Class:           proto.String("self-attestation"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:        recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	if after := Fingerprint(dir, transcripts); after == before {
		t.Error("a gap was minted and the fingerprint did not move. In WAL mode the write lands " +
			"in record.db-wal, not record.db — a watcher keyed on record.db alone shows a frozen " +
			"board while the run is at its busiest")
	}
}

// A NEW TRANSCRIPT, AND A GROWN ONE, ARE BOTH CHANGES. Seats are append-only writers and a new
// seat is a new file, so a fingerprint that only stat'd files it had already seen would miss the
// arrival of a whole seat.
func TestFingerprintSeesTranscriptArrivalAndGrowth(t *testing.T) {
	dir := seedRun(t)
	transcripts := t.TempDir()
	base := Fingerprint(dir, transcripts)

	path := filepath.Join(transcripts, "agent-a1.jsonl")
	if err := os.WriteFile(path, []byte("{\"a\":1}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	arrived := Fingerprint(dir, transcripts)
	if arrived == base {
		t.Fatal("a new seat transcript did not change the fingerprint")
	}

	if err := os.WriteFile(path, []byte("{\"a\":1}\n{\"a\":2}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if grown := Fingerprint(dir, transcripts); grown == arrived {
		t.Error("an appended transcript did not change the fingerprint")
	}
}

// THE WATCHER MUST NOT CHASE ITS OWN TAIL. dashboard.html is what it writes; if writing it
// changed the fingerprint that decides whether to write it, no tick would ever be skipped and
// the whole change-detection would be inert while looking correct.
func TestFingerprintIgnoresThePageTheWatcherWrites(t *testing.T) {
	dir := seedRun(t)
	transcripts := t.TempDir()
	before := Fingerprint(dir, transcripts)

	if err := os.WriteFile(filepath.Join(dir, outputFileName), []byte("<html>a render</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if after := Fingerprint(dir, transcripts); after != before {
		t.Errorf("writing %s changed the fingerprint; the watcher would re-render forever on its "+
			"own output", outputFileName)
	}
}

// UNMEASURABLE IS NOT UNCHANGED, and Changed is the one place that is decided so no caller has
// to remember the direction. Rendering when we need not costs CPU; skipping when we should not
// costs the operator a page that quietly stopped being true.
func TestChangedFailsTowardsRendering(t *testing.T) {
	for _, tc := range []struct {
		name, previous, current string
		want                    bool
	}{
		{"identical digests", "abc", "abc", false},
		{"different digests", "abc", "def", true},
		{"no previous reading", "", "abc", true},
		{"current unmeasurable", "abc", "", true},
		{"neither measurable", "", "", true},
	} {
		if got := Changed(tc.previous, tc.current); got != tc.want {
			t.Errorf("%s: Changed(%q, %q) = %v, want %v", tc.name, tc.previous, tc.current, got, tc.want)
		}
	}
}

// A MISSING DIRECTORY IS UNMEASURABLE, NOT EMPTY. A transcript directory that does not exist yet
// must not fingerprint the same as one that does and is quiet.
func TestFingerprintOfNothingIsUnmeasurable(t *testing.T) {
	if got := Fingerprint(filepath.Join(t.TempDir(), "absent"), ""); got != "" {
		t.Errorf("a run directory that does not exist fingerprinted as %q, want \"\"", got)
	}
}

// THE HEADLINE TOTAL IS THE SUM OF THE BREAKDOWN PRINTED UNDER IT.
//
// It was not, and nothing checked. The total priced each MESSAGE at its own `model` field while
// the breakdown priced each FILE at one tier chosen for the whole transcript — two definitions of
// one run's cost, on one page, agreeing only while every transcript carried a single model. A
// transcript carrying two is exactly what `setup --allow-substitution` exists to permit.
//
// The fixture writes one, so this test fails against the old two-pass code rather than merely
// describing the new one.
func TestCostTotalEqualsTheSumOfItsBreakdown(t *testing.T) {
	dir := seedRun(t)
	transcripts := t.TempDir()

	// A seat whose transcript carries two different models — the substitution case.
	mixed := "" +
		`{"message":{"model":"claude-opus-4-1","usage":{"input_tokens":1000,"output_tokens":100,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}` + "\n" +
		`{"message":{"model":"claude-3-5-haiku-20241022","usage":{"input_tokens":1000,"output_tokens":100,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}` + "\n"
	if err := os.WriteFile(filepath.Join(transcripts, "agent-mixed.jsonl"), []byte(mixed), 0o644); err != nil {
		t.Fatal(err)
	}

	m := BuildModel(runtest.Open(t, dir), transcripts, Config{}, 0)

	var sum float64
	for _, r := range m.CostRows {
		sum += r.Cost
	}
	// Float equality is wrong for money; the two now come from the SAME addends, so the only
	// difference tolerable here is float association.
	if diff := m.Cost - sum; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("the page's total is %.6f and its own breakdown sums to %.6f (difference %.6g). "+
			"Two definitions of one run's cost on one page is the two-readers defect: the total "+
			"priced per MESSAGE, the breakdown priced per FILE, and a transcript carrying two "+
			"models made them disagree", m.Cost, sum, diff)
	}
	if m.Cost == 0 {
		t.Fatal("the fixture priced to zero, so this test would pass on any implementation")
	}
}
