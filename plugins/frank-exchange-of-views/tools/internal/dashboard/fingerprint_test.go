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
// The record runs journal_mode=WAL. That makes both answers non-obvious and both failures
// silent: a committed write lands in record.db-wal and may never touch record.db until a
// checkpoint, so a fingerprint over record.db alone freezes the dashboard exactly when the run
// is busiest; and a READER touches record.db-shm, so a fingerprint that includes it is dirtied
// by the very BuildModel call it was measuring and never skips a tick.
//
// Neither can be settled by reading the driver's documentation, so they are settled here.
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
