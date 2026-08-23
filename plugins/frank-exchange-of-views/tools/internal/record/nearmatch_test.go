package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
	"testing"
)

// gapSpec is one gap to seed onto the board for the near-match tests.
type gapSpec struct {
	id, problem, location string
	open                  bool
}

// mintBoard writes all the specs as ONE merge shard (mints, then closes for the closed ones),
// mirroring a real merge round: one seat, one nonce, incrementing seq. Two shards for the
// same seat would each restart seq at 0 and not compose — the board is built by one writer.
func mintBoard(t *testing.T, runDir string, specs ...gapSpec) {
	t.Helper()
	const seat, nonce = "red-merge-r1", "aaaaaaaa"
	var evs []*Event
	seq := 0
	for _, s := range specs {
		evs = append(evs, recordtest.At(t, seat, 1, seat+":mint:"+s.id, &recordpb.Mint{
			GapId: proto.String(s.id), Class: proto.String("overclaim"),
			Problem: proto.String(s.problem), Location: proto.String(s.location),
			AcceptanceCheck: proto.String("check"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_HIGH),
			Impact:          recordtest.P(recordpb.Grade_GRADE_HIGH),
		}))
		seq++
	}
	for _, s := range specs {
		if !s.open {
			evs = append(evs, recordtest.At(t, seat, 1, seat+":close:"+s.id, &recordpb.Close{GapId: proto.String(s.id), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")}))
			seq++
		}
	}
	writeShard(t, runDir, evs)
}

// A near-duplicate of an existing gap must outrank an unrelated gap, and both open and
// closed gaps are scored (a reopen usually matches a closed one).
func TestNearMatchRanksDuplicateAboveUnrelated(t *testing.T) {
	runDir := newRun(t)
	mintBoard(t, runDir,
		gapSpec{"R1-1", "the cache eviction races the concurrent reader and returns stale entries", "cache.go:88", true},
		gapSpec{"R1-2", "the documentation heading uses the wrong capitalization", "README.md:1", true})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	got := NearMatch(b, "cache eviction races the reader returning stale entries", "cache.go", 5)
	if len(got) == 0 {
		t.Fatal("expected at least one near-match")
	}
	if got[0].ID != "R1-1" {
		t.Errorf("the duplicate R1-1 must rank first, got %q (full: %+v)", got[0].ID, got)
	}
	if got[0].Score <= 0 {
		t.Errorf("the top match must have a positive score, got %v", got[0].Score)
	}
	// The unrelated doc-capitalization gap should score below the duplicate (or not appear).
	for _, m := range got {
		if m.ID == "R1-2" && m.Score >= got[0].Score {
			t.Errorf("unrelated R1-2 (%v) must score below the duplicate R1-1 (%v)", m.Score, got[0].Score)
		}
	}
}

// The location bonus lifts a candidate that shares its section with a gap.
func TestNearMatchLocationBonus(t *testing.T) {
	runDir := newRun(t)
	mintBoard(t, runDir,
		gapSpec{"R1-1", "an off-by-one in the loop bound", "parser.go:42", true})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	// Same candidate text; only the presence of the matching --quote differs.
	without := NearMatch(b, "off-by-one in the loop bound", "", 5)
	with := NearMatch(b, "off-by-one in the loop bound", "parser.go:42", 5)
	if len(without) != 1 || len(with) != 1 {
		t.Fatalf("expected one match each, got without=%d with=%d", len(without), len(with))
	}
	if !(with[0].Score > without[0].Score) {
		t.Errorf("location match must raise the score: without=%v with=%v", without[0].Score, with[0].Score)
	}
}

// Closed gaps are scored (a reopen matches a closed gap) and carry status "closed".
func TestNearMatchScoresClosedGaps(t *testing.T) {
	runDir := newRun(t)
	mintBoard(t, runDir,
		gapSpec{"R1-1", "the retry backoff overflows on the tenth attempt", "retry.go:12", false})

	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	got := NearMatch(b, "retry backoff overflows on the tenth attempt", "", 5)
	if len(got) != 1 || got[0].ID != "R1-1" {
		t.Fatalf("closed gap must still be screened: %+v", got)
	}
	if got[0].Status != "closed" {
		t.Errorf("status = %q, want closed", got[0].Status)
	}
}

// topN caps the result set; a candidate with no overlap returns nothing (0 is not a match).
func TestNearMatchTopNAndNoOverlap(t *testing.T) {
	runDir := newRun(t)
	var specs []gapSpec
	for _, id := range []string{"R1-1", "R1-2", "R1-3", "R1-4", "R1-5", "R1-6"} {
		specs = append(specs, gapSpec{id, "shared token alpha beta gamma delta " + id, "file.go", true})
	}
	mintBoard(t, runDir, specs...)
	b, err := BoardState(runDir)
	if err != nil {
		t.Fatal(err)
	}
	got := NearMatch(b, "alpha beta gamma delta", "", 3)
	if len(got) != 3 {
		t.Errorf("topN=3 must cap the result at 3, got %d", len(got))
	}
	none := NearMatch(b, "wholly unrelated zulu yankee xray tokens", "", 5)
	if len(none) != 0 {
		t.Errorf("a candidate with no overlap must return nothing, got %+v", none)
	}
}
