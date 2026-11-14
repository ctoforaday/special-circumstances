package record

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The reword chain reads what each edit changed off the paragraphs the edit stamped, never off the
// report rebuilt at a past event: an edit carrying no stamped paragraphs is a record the chain cannot
// read, and the read says so rather than taking the edit for one that touched nothing.
func TestTheRewordChainRefusesAnEditWithNoStampedParagraphs(t *testing.T) {
	const reworded = "The first claim rests on the audited figures from every month of the quarter."
	b := debtBoard(t).readFound(evLens, dbURL, dbFirst, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
	runOf(b)
	db, err := recordsql.Open(filepath.Join(b.runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	ev := recordtest.At(t, "blue-respond", "blue-respond:unstamped", &recordpb.BlueEdit{Old: proto.String(dbFirst), New: proto.String(reworded), Text: proto.String("edited")})
	ev.Ts = proto.String("2026-01-01T01:00:00Z")
	if _, err := recordsql.Insert(db, ev); err != nil {
		t.Fatal(err)
	}
	_, err = PassBlockers(runOf(b.register("red-chair")))
	if err == nil || !strings.Contains(err.Error(), "carries no stamped paragraphs") {
		t.Errorf("PassBlockers = %v, want the read's error naming the unstamped edit", err)
	}
}

// A cut that removes a sentence and leaves its gap's token bare removes the sentence: the bare token
// is no sentence the edit reworded it to, and the sentence beside it is not the reword (universe-m37,
// edit 984, where the token run left behind was read as the reword and the chain then followed it to
// the neighbouring sentence).
func TestACutLeavingItsTokensBareRemovesTheSentence(t *testing.T) {
	tok := anchor.Token("G-00000001")
	before := dbFirst + " " + strings.TrimSuffix(dbSecond, ".") + tok + "."
	m := editAt{id: 9, ed: &recordpb.BlueEdit{
		Old: proto.String(" " + strings.TrimSuffix(dbSecond, ".") + tok + "."), New: proto.String(tok),
		OldParagraphs: proto.String(before), NewParagraphs: proto.String(dbFirst + " " + tok),
	}}
	reworded, removed, err := rewordedTo(m)
	if err != nil || !removed || reworded != "" {
		t.Errorf("rewordedTo = (%q, %t, %v), want the sentence removed", reworded, removed, err)
	}
}

// The report at a lane's part is the parts recorded up to it, joined in order: a lane's part recorded
// after it is not on it, though it is on the report as it stands.
func TestTheReportAtAPartHoldsThePartsUpToIt(t *testing.T) {
	b := newStage(t).cast("blue-lane-1", "blue-lane-2").
		add("blue-lane-1", &recordpb.ReportPart{Text: proto.String("# report\n\nThe first lane's part stands here.\n")}).
		add("blue-lane-2", &recordpb.ReportPart{Text: proto.String("The second lane's part stands here.\n")})
	run := runOf(b)
	at, err := ReportAtPart(run, "blue-lane-1")
	if err != nil {
		t.Fatal(err)
	}
	now, err := RenderRun(run)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(at, "first lane's part") || strings.Contains(at, "second lane's part") || !strings.Contains(now, "second lane's part") {
		t.Errorf("the report at the first lane's part = %q (now %q), want the first part alone", at, now)
	}
}
