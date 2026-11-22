package record

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// rawRecord writes rows straight into a fresh record through recordsql.Insert, below the act judge,
// as recordsql's gap-state fixtures do: the rows the write path's verbs cannot yet reach (a Close of
// a docketed gap, P12b's) are written as rows.
type rawRecord struct {
	t  *testing.T
	db *sql.DB
	n  int
}

func newRawRecord(t *testing.T) *rawRecord {
	t.Helper()
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &rawRecord{t: t, db: db}
}

// add inserts body as seat, and returns the insert's error.
func (r *rawRecord) add(seat string, body proto.Message) error {
	r.t.Helper()
	r.n++
	ev := recordtest.Stamped(recordtest.At(r.t, seat, fmt.Sprintf("%s:raw:#%d", seat, r.n), body),
		fmt.Sprintf("2026-01-01T00:%02d:%02dZ", r.n/60, r.n%60))
	_, err := recordsql.Insert(r.db, ev)
	return err
}

// must is add, failing the test on a refusal.
func (r *rawRecord) must(seat string, body proto.Message) {
	r.t.Helper()
	if err := r.add(seat, body); err != nil {
		r.t.Fatalf("writing %T: %v", body, err)
	}
}

func (r *rawRecord) mint(gap string, supersedes ...string) {
	r.t.Helper()
	r.must(evLens, rawMint(gap, supersedes...))
}

// rawMint is a gap's mint, naming the one gap it supersedes where given.
func rawMint(gap string, supersedes ...string) *recordpb.Mint {
	m := &recordpb.Mint{GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(),
		ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("a"),
		CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(), Severity: recordpb.Grade_GRADE_HIGH.Enum(),
		Likelihood: recordpb.Grade_GRADE_HIGH.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}
	if len(supersedes) > 0 {
		m.Supersedes = proto.String(supersedes[0])
	}
	return m
}

func (r *rawRecord) docket(id, gap string, cause recordpb.DocketingCause, motion string) {
	r.t.Helper()
	d := &recordpb.Docketing{DocketingId: proto.String(id), GapId: proto.String(gap), Cause: cause.Enum(), Basis: proto.String("before the bench")}
	if motion != "" {
		d.MotionId = proto.String(motion)
	}
	r.must(chairSeat, d)
}

// propose files blue's proposal id on gap: a supersede by successor, or (successor "") a regrade.
func (r *rawRecord) propose(id, gap, successor string) {
	r.t.Helper()
	m := &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String("b"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}}}
	if successor == "" {
		m.Act = &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{GapId: proto.String(gap), Severity: recordpb.Grade_GRADE_LOW.Enum(), Basis: proto.String("lower")}}
	} else {
		m.Act = &recordpb.Motion_Supersede{Supersede: rawMint(successor, gap)}
	}
	r.must("blue-respond", m)
}

func (r *rawRecord) close(gap string, as recordpb.Disposition, carriedFrom string) {
	r.t.Helper()
	c := &recordpb.Close{GapId: proto.String(gap), ClosureClass: as.Enum(), Prose: proto.String("closed")}
	if carriedFrom != "" {
		c.CarriedFrom = proto.String(carriedFrom)
	}
	r.must(evLens, c)
}

// ruling is the bench's ruling of docketing id, naming motion where it holds one.
func ruling(id, motion string, answer func(*recordpb.DocketRuling)) *recordpb.MotionRule {
	d := &recordpb.DocketRuling{DocketingId: proto.String(id), Principle: proto.String("pr"), Tension: proto.String("tn"),
		ReviewFlag: proto.String("rf"), Settled: proto.String("st")}
	answer(d)
	r := &recordpb.MotionRule{Opinion: proto.String("ruled"), Ruling: &recordpb.MotionRule_Docket{Docket: d}}
	if motion != "" {
		r.MotionId = proto.String(motion)
	}
	return r
}

// standing is the record's standing docketing and proposal per gap, stored (the gap view) and folded
// (docketingsAt, the Motions fold), each as "gap docketing@at proposal" lines.
func (r *rawRecord) standing() (stored, folded []string) {
	r.t.Helper()
	rows, err := r.db.Query(`SELECT "gap_id", COALESCE("standing_docketing", ''), COALESCE("standing_docketing_at", 0), COALESCE("unruled_docket_motion", '') FROM "gap"`)
	if err != nil {
		r.t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var g, d, m string
		var at int64
		if err := rows.Scan(&g, &d, &at, &m); err != nil {
			r.t.Fatal(err)
		}
		stored = append(stored, fmt.Sprintf("%s %s@%d %s", g, d, at, m))
	}
	evs, win, err := eventsAt(r.db)
	if err != nil {
		r.t.Fatal(err)
	}
	ids := win.IDs(evs)
	docketed := standingDocketings(docketingsAt(evs, ids))
	proposed := map[string]string{}
	for _, m := range motionsAt(evs, ids, win) {
		if m.Subject == "docket" && m.Standing() {
			proposed[m.GapID] = m.ID
		}
	}
	gaps := map[string]bool{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			gaps[m.GetGapId()] = true
		}
	}
	for g := range gaps {
		d, at := "", int64(0)
		if k := docketed[g]; k != nil {
			d, at = k.ID, k.at
		}
		folded = append(folded, fmt.Sprintf("%s %s@%d %s", g, d, at, proposed[g]))
	}
	sort.Strings(stored)
	sort.Strings(folded)
	return stored, folded
}

// standingOf is gap's stored "docketing proposal" pair.
func (r *rawRecord) standingOf(gap string) string {
	r.t.Helper()
	var d, m string
	if err := r.db.QueryRow(`SELECT COALESCE("standing_docketing", ''), COALESCE("unruled_docket_motion", '') FROM "gap" WHERE "gap_id" = ?`, gap).Scan(&d, &m); err != nil {
		r.t.Fatal(err)
	}
	return strings.TrimSpace(d + " " + m)
}

// T-B34. THE STORED DOCKETING IS THE FOLD, AND A CLOSE ENDS IT. After every step the gap state's
// standing docketing and standing proposal are the Go folds' (docketingsAt, the Motions fold), and
// each step leaves the state it states: a non-carry Close of the gap written after a docketing of
// each cause ends it — once moot, once another word — and the proposal a PROPOSAL docketing holds,
// with no ruling and no tombstone written, and a bench ruling naming the ended docketing is refused;
// a carry Close after a docketing on a gap already closed (F8) ends nothing, and neither does the
// Close before the docketing's row.
//
// The duty readers of a standing docketing (closings owed, the docketed blocker, the freeze,
// benchHeld) read these two statements; their rows land with the dispatch's docketings.
func TestTheStoredDocketingIsTheFold(t *testing.T) {
	r := newRawRecord(t)
	impasse, movement, proposal := recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL
	moot, repaired, notADefect := recordpb.Disposition_DISPOSITION_MOOT, recordpb.Disposition_DISPOSITION_REPAIRED, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT
	for i, step := range []struct {
		name string
		act  func()
		want map[string]string // gap → its stored "docketing proposal"
	}{
		{"five gaps", func() {
			for _, g := range []string{"G1", "G2", "G3", "G4", "G5"} {
				r.mint(g)
			}
		}, map[string]string{"G1": "", "G3": ""}},
		{"an impasse docketing stands", func() { r.docket("K1", "G1", impasse, "") }, map[string]string{"G1": "K1"}},
		{"a grade-movement docketing stands", func() { r.docket("K2", "G2", movement, "") }, map[string]string{"G2": "K2"}},
		{"a proposal stands", func() { r.propose("M3", "G3", "") }, map[string]string{"G3": "M3"}},
		{"its docketing holds it", func() { r.docket("K3", "G3", proposal, "M3") }, map[string]string{"G3": "K3 M3"}},
		{"a moot close ends the impasse docketing", func() { r.close("G1", moot, "") }, map[string]string{"G1": ""}},
		{"a repaired close ends the grade-movement docketing", func() { r.close("G2", repaired, "") }, map[string]string{"G2": ""}},
		{"a close ends the proposal's docketing and the proposal", func() { r.close("G3", notADefect, "") }, map[string]string{"G3": ""}},
		{"a close before the docketing's row", func() { r.close("G4", repaired, "") }, map[string]string{"G4": ""}},
		{"a proposal on the closed gap (F8)", func() { r.propose("M4", "G4", "") }, map[string]string{"G4": "M4"}},
		{"its docketing stands on the closed gap", func() { r.docket("K4", "G4", proposal, "M4") }, map[string]string{"G4": "K4 M4"}},
		{"a carry close after it ends nothing", func() { r.close("G4", repaired, "G4") }, map[string]string{"G4": "K4 M4"}},
		{"a docketing the bench rules", func() { r.docket("K5", "G5", impasse, "") }, map[string]string{"G5": "K5"}},
		{"its ruling ends it", func() {
			r.must(benchSeat, ruling("K5", "", func(d *recordpb.DocketRuling) {
				d.SidedWith, d.Disposition, d.Final = recordpb.SidedWith_SIDED_WITH_BLUE.Enum(), notADefect.Enum(), proto.Bool(true)
			}))
		}, map[string]string{"G5": ""}},
	} {
		step.act()
		stored, folded := r.standing()
		if fmt.Sprint(stored) != fmt.Sprint(folded) {
			t.Fatalf("after step %d (%s) the stored standing is not the fold's:\nstored %v\nfolded %v", i, step.name, stored, folded)
		}
		for g, want := range step.want {
			if got := r.standingOf(g); got != want {
				t.Errorf("after step %d (%s) %s stands %q, want %q", i, step.name, g, got, want)
			}
		}
	}

	// A RULING OF AN ENDED DOCKETING IS REFUSED, and nothing of a ruling is written for it.
	for _, k := range []struct{ id, motion string }{{"K1", ""}, {"K3", "M3"}} {
		err := r.add(benchSeat, ruling(k.id, k.motion, func(d *recordpb.DocketRuling) {
			d.SidedWith, d.Disposition, d.Final = recordpb.SidedWith_SIDED_WITH_BLUE.Enum(), notADefect.Enum(), proto.Bool(true)
		}))
		if err == nil || !strings.Contains(err.Error(), "neither stands behind this ruling") {
			t.Errorf("a ruling of %s, which a close ended = %v, want refused", k.id, err)
		}
	}
	var rulings, tombstones int
	if err := r.db.QueryRow(`SELECT (SELECT count(*) FROM "motion_rule_docket" WHERE "docketing_id" IN ('K1', 'K2', 'K3')),
	    (SELECT count(*) FROM "tombstone")`).Scan(&rulings, &tombstones); err != nil {
		t.Fatal(err)
	}
	if rulings != 0 || tombstones != 0 {
		t.Errorf("the ended docketings hold %d ruling(s) and the record %d tombstone(s), want none", rulings, tombstones)
	}
}

// T-B32. EVERY DOCKET RULING READER RESOLVES ITS GAP THROUGH THE DOCKETING. For each side's answer at
// impasse, each grade answer at grade movement (no motion_id), a remand, and each side's answer to a
// held proposal, the readers in this package return the ruling's gap and its word: DocketRulingGap,
// the docketings fold, the supersession fold's closed set (a successor minted after a closing ruling
// closes nothing), the gap state's bench arms (closed_by_bench: the closing arm, and the supersede arm
// on sided_with), motion_answers for the held proposal, and the trigger admitting each ruling (it is
// written). The readers outside this package (capture's precedents, the report's and view's LEAD
// lines, dashboard, scorecard, verify, graph) read DocketRulingGap and DocketingsOf.
func TestEveryDocketRulingReaderResolvesItsGapThroughTheDocketing(t *testing.T) {
	impasse, movement, proposal := recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL
	side := func(s recordpb.SidedWith, word recordpb.Disposition) func(*recordpb.DocketRuling) {
		return func(d *recordpb.DocketRuling) {
			d.SidedWith = s.Enum()
			if word != recordpb.Disposition_DISPOSITION_UNSPECIFIED {
				d.Disposition, d.Final = word.Enum(), proto.Bool(true)
			}
		}
	}
	grades := func(g recordpb.MovedGrades) func(*recordpb.DocketRuling) {
		return func(d *recordpb.DocketRuling) { d.MovedGrades = g.Enum() }
	}
	remand := func(d *recordpb.DocketRuling) {
		d.Disposition, d.Direction = recordpb.Disposition_DISPOSITION_REMANDED.Enum(), proto.String("one more exchange")
	}
	for _, c := range []struct {
		name      string
		cause     recordpb.DocketingCause
		supersede bool // the docketing holds blue's proposed supersede of the gap
		answer    func(*recordpb.DocketRuling)
		word      string
		closes    bool
	}{
		{"for blue at impasse", impasse, false, side(recordpb.SidedWith_SIDED_WITH_BLUE, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT), "not_a_defect", true},
		{"for red at impasse", impasse, false, side(recordpb.SidedWith_SIDED_WITH_RED, recordpb.Disposition_DISPOSITION_REPAIRED), "repaired", true},
		{"a remand at impasse", impasse, false, remand, "remanded", false},
		{"grades kept", movement, false, grades(recordpb.MovedGrades_MOVED_GRADES_KEPT), "grades kept", false},
		{"grades restored", movement, false, grades(recordpb.MovedGrades_MOVED_GRADES_RESTORED), "grades restored", false},
		{"for blue on its supersede", proposal, true, side(recordpb.SidedWith_SIDED_WITH_BLUE, recordpb.Disposition_DISPOSITION_UNSPECIFIED), "for blue", true},
		{"for red on blue's supersede", proposal, true, side(recordpb.SidedWith_SIDED_WITH_RED, recordpb.Disposition_DISPOSITION_REPAIRED), "repaired", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRawRecord(t)
			r.mint("G")
			motion := ""
			if c.supersede {
				motion = "M"
				r.propose(motion, "G", "S")
			}
			if c.cause == proposal {
				r.docket("K", "G", c.cause, motion)
			} else {
				r.docket("K", "G", c.cause, "")
			}
			rule := ruling("K", motion, c.answer)
			if err := r.add(benchSeat, rule); err != nil {
				t.Fatalf("the trigger refused the ruling: %v", err)
			}
			var open, byBench bool
			if err := r.db.QueryRow(`SELECT "open", "closed_by_bench" FROM "gap" WHERE "gap_id" = 'G'`).Scan(&open, &byBench); err != nil {
				t.Fatal(err)
			}
			if open == c.closes || byBench != c.closes {
				t.Errorf("G open=%v closed_by_bench=%v, want open=%v closed_by_bench=%v", open, byBench, !c.closes, c.closes)
			}
			// The successor the tool writes after a ruling siding with the supersede, or a mint
			// naming the gap after any closing ruling: lineage, closing nothing. After a ruling that
			// closes nothing, it closes the gap.
			r.mint("S", "G")
			evs, win, err := eventsAt(r.db)
			if err != nil {
				t.Fatal(err)
			}
			if got := DocketRulingGap(evs)(rule); got != "G" {
				t.Errorf("DocketRulingGap = %q, want G", got)
			}
			ds := docketingsAt(evs, win.IDs(evs))
			if len(ds) != 1 || !ds[0].Ruled() || ds[0].GapID != "G" || ds[0].Word() != c.word {
				t.Errorf("the docketings fold = %+v, want K on G ruled %q", ds, c.word)
			}
			if _, superseded := supersessionsOf(evs)["G"]; superseded == c.closes {
				t.Errorf("the supersession fold has S closing G = %v, want %v (a closing ruling precedes it)", superseded, !c.closes)
			}
			if motion != "" {
				var word string
				if err := r.db.QueryRow(`SELECT "ruling" FROM "motion_answers" WHERE "motion_id" = ?`, motion).Scan(&word); err != nil || word != c.word {
					t.Errorf("motion_answers' word for %s = %q (%v), want %q", motion, word, err, c.word)
				}
			}
		})
	}
}
