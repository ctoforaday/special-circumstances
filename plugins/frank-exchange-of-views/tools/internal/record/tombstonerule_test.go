package record

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// tombstoneDrive is a fixture stream covering each placing and ending arm of the tombstone rule: a
// bench ruling of each anchor kind of ruled gap and the act that ends each, a remand that places
// nothing, and estoppel placed, ended by a later verbatim edit answering another gap (L24), and placed
// again; and a pure cut ending each kind its removed_anchors reaches (F16): a ruled gap's own token, the
// anchor a ruled gap is about, and an estopped gap's token. Each step is one act, written through the write path (a Replay: writeIn, validate, no act
// judgement).
func tombstoneDrive() []struct {
	seat string
	body proto.Message
} {
	type step = struct {
		seat string
		body proto.Message
	}
	mint := func(gap string, edit func(*recordpb.Mint)) step {
		m := rawMint(gap)
		m.ClassMaterial = nil
		edit(m)
		return step{evLens, m}
	}
	quote := func(loc, fix string) func(*recordpb.Mint) {
		return func(m *recordpb.Mint) {
			m.Location = proto.String(loc)
			if fix != "" {
				m.FixNew = proto.String(fix)
			}
		}
	}
	about := func(kind recordpb.AboutKind, ref string) func(*recordpb.Mint) {
		return func(m *recordpb.Mint) { m.AboutKind, m.AboutRef = kind.Enum(), proto.String(ref) }
	}
	docket := func(id, gap string) step {
		return step{chairSeat, &recordpb.Docketing{DocketingId: proto.String(id), GapId: proto.String(gap),
			Cause: recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE.Enum(), Basis: proto.String("before the bench")}}
	}
	forBlue := func(id string) step {
		return step{benchSeat, ruling(id, "", func(d *recordpb.DocketRuling) {
			d.SidedWith, d.Disposition, d.Final = recordpb.SidedWith_SIDED_WITH_AUTHOR.Enum(), recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), proto.Bool(true)
		})}
	}
	remand := func(id string) step {
		return step{benchSeat, ruling(id, "", func(d *recordpb.DocketRuling) {
			d.Disposition, d.Direction = recordpb.Disposition_DISPOSITION_REMANDED.Enum(), proto.String("one more exchange")
		})}
	}
	edit := func(answers string, verbatim bool, reopened ...string) step {
		e := &recordpb.BlueEdit{Old: proto.String("o"), New: proto.String("n"), Text: proto.String("why"), Reopened: reopened}
		if answers != "" {
			e.Answers = proto.String(answers)
		}
		if verbatim {
			e.AppliedVerbatim = proto.Bool(true)
		}
		return step{blueRespondSeat, e}
	}
	cut := func(removed ...string) step {
		return step{blueRespondSeat, &recordpb.BlueEdit{Old: proto.String("o"), New: proto.String(""), Text: proto.String("why"), RemovedAnchors: removed}}
	}
	anchor, inquiry := recordpb.AboutKind_ABOUT_KIND_ANCHOR, recordpb.AboutKind_ABOUT_KIND_INQUIRY
	fix := strings.Repeat("the prescribed sentence carries its evidence ", 2)
	return []step{
		{blueRespondSeat, &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q?"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}},
		mint("G1", quote("Alpha holds.", "")),
		{blueRespondSeat, stamped(&recordpb.Cite{Label: proto.String("C-00000001"), Url: proto.String("https://example.org/a"), Title: proto.String("t"), CiteKey: proto.String("k-a"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()})},
		{blueRespondSeat, &recordpb.Computation{ComputationId: proto.String("K-00000002"), Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), Argv: []string{"true"}, BundleSha: proto.String("sha-k")}},
		mint("G2", about(anchor, "C-00000001")),
		mint("G3", about(anchor, "K-00000002")),
		mint("G4", about(inquiry, "Q-1")),
		mint("G5", quote("Echo holds.", "")),
		mint("P1", quote("Papa one holds.", fix+"one")),
		mint("P2", quote("Papa two holds.", fix+"two")),
		mint("B1", quote("Bravo holds.", fix+"bravo")),
		{blueRespondSeat, stamped(&recordpb.Cite{Label: proto.String("C-00000008"), Url: proto.String("https://example.org/h"), Title: proto.String("t"), CiteKey: proto.String("k-h"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()})},
		mint("G6", quote("Foxtrot holds.", "")),
		mint("G7", about(anchor, "C-00000008")),
		mint("P3", quote("Papa three holds.", fix+"three")),
		// a ruling on each, then the act that ends it
		docket("K1", "G1"), forBlue("K1"),
		edit("", false, "G5"), // a reword of another sentence ends nothing
		edit("", false, "G1"),
		docket("K2", "G2"), forBlue("K2"),
		edit("", false, "C-00000001"),
		docket("K3", "G3"), forBlue("K3"),
		{blueRespondSeat, &recordpb.Retire{Claim: proto.String("the computation"), Reason: proto.String("superseded"), Anchors: []string{"K-00000002"}}},
		docket("K4", "G4"), forBlue("K4"),
		{blueRespondSeat, &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED.Enum(),
			SupersedesStatus: proto.String("proposed"), Reason: proto.String("it held"), Method: proto.String("read")}},
		// a remand places nothing; the ruling after it does, and stands for the run
		docket("K5", "G5"), remand("K5"),
		docket("K6", "G5"), forBlue("K6"),
		// estoppel: placed, a reword of P2's sentence by an edit not answering it ends nothing; a verbatim
		// edit answering P2 whose reopened names P1 ends P1's and places P2's, its own reopened ending
		// nothing; P1's verbatim re-application places it again
		edit("P1", true, "P1"),
		edit("", false, "P2"),
		edit("P2", true, "P2", "P1"),
		edit("P1", true, "P1"),
		// a verbatim re-application where P1's stands writes a new placing row and no end
		edit("P1", true, "P1"),
		// a second reword of G1's sentence, with nothing standing, writes no row
		edit("", false, "G1"),
		// a pure cut ends what its removed_anchors reaches (F16): G6's own token, the citation G7 is
		// about, and P3's token
		docket("K7", "G6"), forBlue("K7"),
		cut("G6"),
		docket("K8", "G7"), forBlue("K8"),
		cut("C-00000008"),
		edit("P3", true, "P3"),
		cut("P3"),
		// the bench's verbatim application of red's closing text on B1, whose mint prescribes the same
		// text, places no estoppel: the bar on what the bench decided is its ruling's tombstone
		{benchSeat, &recordpb.BlueEdit{Old: proto.String("o"), New: proto.String("n"), Text: proto.String("red's text"),
			Answers: proto.String("B1"), AppliedVerbatim: proto.Bool(true), Reopened: []string{"B1"}}},
	}
}

// tombstonesFromStream is the rule written over the events, not through tombstonesOwed: the
// "gap kind" pairs a tombstone stands on at the end of evs.
func tombstonesFromStream(evs []*Event) []string {
	type pair struct{ gap, kind string }
	stand := map[pair]bool{}
	mints := map[string]*recordpb.Mint{}
	minter := map[string]string{}
	docketed := map[string]string{}
	about := func(kind recordpb.AboutKind, ref string) []string {
		var out []string
		for g, m := range mints {
			if m.GetAboutKind() == kind && m.GetAboutRef() == ref {
				out = append(out, g)
			}
		}
		return out
	}
	const ruling, estoppel = "ruling", "estoppel"
	for _, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			mints[b.GetGapId()], minter[b.GetGapId()] = b, e.GetSeatId()
		case *recordpb.Docketing:
			docketed[b.GetDocketingId()] = b.GetGapId()
		case *recordpb.MotionRule:
			d := b.GetDocket()
			if g := docketed[d.GetDocketingId()]; g != "" && d.GetDisposition() != recordpb.Disposition_DISPOSITION_REMANDED && e.GetSeatId() != minter[g] {
				stand[pair{g, ruling}] = true
			}
		case *recordpb.BlueEdit:
			placed := ""
			if b.GetAppliedVerbatim() && mints[b.GetAnswers()].GetFixNew() != "" && e.GetSeatId() != benchSeat {
				placed = b.GetAnswers()
			}
			for _, r := range slices.Concat(b.GetReopened(), b.GetRemovedAnchors()) {
				delete(stand, pair{r, ruling})
				for _, g := range about(recordpb.AboutKind_ABOUT_KIND_ANCHOR, r) {
					delete(stand, pair{g, ruling})
				}
				if r != placed {
					delete(stand, pair{r, estoppel})
				}
			}
			if placed != "" {
				stand[pair{placed, estoppel}] = true
			}
		case *recordpb.Retire:
			for _, a := range b.GetAnchors() {
				for _, g := range about(recordpb.AboutKind_ABOUT_KIND_ANCHOR, a) {
					delete(stand, pair{g, ruling})
				}
			}
		case *recordpb.Inquiry:
			for _, g := range about(recordpb.AboutKind_ABOUT_KIND_INQUIRY, b.GetInquiryId()) {
				delete(stand, pair{g, ruling})
			}
		}
	}
	var out []string
	for p := range stand {
		out = append(out, p.gap+" "+p.kind)
	}
	slices.Sort(out)
	return out
}

// runTombstoneDrive writes the drive through a replay and, after every step, compares the standing
// rows with the stream's fold and checks each row sits right after the act that owed it, under its
// seat and stamp. It returns the first disagreement, "" for none.
func runTombstoneDrive(t *testing.T) string {
	t.Helper()
	run := mustRun(t, newRun(t))
	b := mustBegin(t, run)
	for i, s := range tombstoneDrive() {
		if _, _, err := b.Write(Identity{Run: run, SeatID: s.seat}, s.body, replayAt); err != nil {
			t.Fatalf("step %d (%T) refused: %v", i, s.body, err)
		}
		evs, _, err := eventsAt(b.Querier())
		if err != nil {
			t.Fatal(err)
		}
		var acts []*Event
		var act *Event
		for _, e := range evs {
			if tb, ok := recordpb.BodyAs[*recordpb.Tombstone](e); ok {
				if act == nil || act.GetSeatId() != e.GetSeatId() || act.GetTs() != e.GetTs() {
					return fmt.Sprintf("step %d: the row %v does not follow the act that owes it", i, tb)
				}
				continue
			}
			act = e
			acts = append(acts, e)
		}
		rows, err := b.Querier().Query(`SELECT "gap_id" || ' ' || "kind" FROM "tombstone_standing" ORDER BY 1`)
		if err != nil {
			t.Fatal(err)
		}
		var stored []string
		for rows.Next() {
			var r string
			if err := rows.Scan(&r); err != nil {
				t.Fatal(err)
			}
			stored = append(stored, r)
		}
		rows.Close()
		if want := tombstonesFromStream(acts); !slices.Equal(stored, want) {
			return fmt.Sprintf("after step %d (%T) the standing rows are %v, and the rule over the stream gives %v", i, s.body, stored, want)
		}
	}
	// NO ACT WROTE A ROW IT DID NOT OWE: an end only where one stood, so the counts are the drive's.
	var places, ends int
	if err := b.Querier().QueryRow(`SELECT count(*) FILTER (WHERE "ends" IS NULL), count(*) FILTER (WHERE "ends") FROM "tombstone"`).Scan(&places, &ends); err != nil {
		t.Fatal(err)
	}
	if places != 12 || ends != 8 {
		return fmt.Sprintf("the drive wrote %d placing and %d ending rows, want 12 and 8", places, ends)
	}
	return ""
}

// THE STORED TOMBSTONES ARE THE RULE (#1695 §III.B12). After every step of a drive covering each placing
// and ending act, the standing rows equal the rule folded over the event stream, and each row sits right
// after the act that owes it, in its transaction. A missed row reads as no bar, so this comparison is
// what makes the miss loud.
func TestTheStoredTombstonesAreTheRule(t *testing.T) {
	if miss := runTombstoneDrive(t); miss != "" {
		t.Fatal(miss)
	}
}

// EVERY TOMBSTONE ARM IS NEEDED: the drive, with each arm of tombstoneArms removed in turn, disagrees
// with the rule over the stream. An arm no step needs is untested.
func TestEveryTombstoneArmIsNeeded(t *testing.T) {
	all := tombstoneArms
	t.Cleanup(func() { tombstoneArms = all })
	for i, a := range all {
		t.Run(a.name, func(t *testing.T) {
			tombstoneArms = slices.Delete(slices.Clone(all), i, i+1)
			defer func() { tombstoneArms = all }()
			if runTombstoneDrive(t) == "" {
				t.Errorf("the drive agrees with the rule without the arm %q", a.name)
			}
		})
	}
}
