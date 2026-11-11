package recordsql

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// gapStateSnapshot is every gap_state row, in mint order, one string each.
func gapStateSnapshot(t *testing.T, q interface {
	Query(string, ...any) (*sql.Rows, error)
}) []string {
	t.Helper()
	rows, err := q.Query(`SELECT * FROM "gap_state" ORDER BY "minted_event"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []string
	for rows.Next() {
		v := make([]any, len(cols))
		p := make([]any, len(cols))
		for i := range v {
			p[i] = &v[i]
		}
		if err := rows.Scan(p...); err != nil {
			t.Fatal(err)
		}
		out = append(out, fmt.Sprint(v))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// gapStateRefreshDiff holds the stored gap state, as the marker triggers kept it append by append,
// against gap_refresh asked of EVERY gap at once, in a transaction it rolls back. It returns "" when
// they agree row for row and nothing is left named or unflushed, else what differs. The refresh
// asked of every gap is the rules with no marker in the way, so a gap a marker failed to name stands
// out as the row the full refresh moves.
func gapStateRefreshDiff(t *testing.T, db *sql.DB) string {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	var out []string
	var pending int
	if err := tx.QueryRow(`SELECT (SELECT count(*) FROM "gap_dirty") + (SELECT count(*) FROM "gap_flush")`).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending != 0 {
		out = append(out, fmt.Sprintf("%d gap(s) named and not flushed", pending))
	}
	stored := gapStateSnapshot(t, tx)
	if _, err := tx.Exec(`INSERT OR IGNORE INTO "gap_dirty" ("gap_id") SELECT "gap_id" FROM "mint";
	    INSERT INTO "gap_flush" ("id") VALUES (1)`); err != nil {
		t.Fatal(err)
	}
	refreshed := gapStateSnapshot(t, tx)
	for i := 0; i < max(len(stored), len(refreshed)); i++ {
		var s, f string
		if i < len(stored) {
			s = stored[i]
		}
		if i < len(refreshed) {
			f = refreshed[i]
		}
		if s != f {
			out = append(out, fmt.Sprintf("stored %q, refreshed %q", s, f))
		}
	}
	return strings.Join(out, "\n")
}

// gapStateHolds is each of a step's truths that does not hold, or "" when all do.
func gapStateHolds(t *testing.T, db *sql.DB, holds []string) string {
	t.Helper()
	var out []string
	for _, q := range holds {
		var ok bool
		if err := db.QueryRow(q).Scan(&ok); err != nil || !ok {
			out = append(out, fmt.Sprintf("%s: got %v (%v)", q, ok, err))
		}
	}
	return strings.Join(out, "\n")
}

// fileDocket files docket motion id on gap.
func (r *demandRecord) fileDocket(id, gap string) int64 {
	return r.add("red-chair", recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
		MotionId: proto.String(id), Basis: proto.String("red cannot settle this one"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}}})
}

// ruleDocket rules docket motion id as disposition; a remand states its direction.
func (r *demandRecord) ruleDocket(id string, as recordpb.Disposition) int64 {
	dr := &recordpb.DocketRuling{Disposition: as.Enum(), Principle: proto.String("correctness over economy"),
		Tension: proto.String("the repair costs a round"), ReviewFlag: proto.String("no"),
		Settled: proto.String("the claim as it stood"), Final: proto.Bool(true)}
	if as == recordpb.Disposition_DISPOSITION_REMANDED {
		dr.Final, dr.Direction = nil, proto.String("a reproduction on the shipped binary")
	}
	return r.add("lead-judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String(id), Opinion: proto.String("stated"), Ruling: &recordpb.MotionRule_Docket{Docket: dr}})
}

// upholdPosition rules position motion id upheld, remanding gaps.
func (r *demandRecord) upholdPosition(id string, gaps ...string) int64 {
	return r.add("lead-judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String(id), Opinion: proto.String("the objection holds"),
		Ruling: &recordpb.MotionRule_Position{Position: &recordpb.PositionHearing{
			Ruling: recordpb.PositionRuling_POSITION_RULING_UPHELD.Enum(), GapIds: gaps, Direction: proto.String("source the figure")}}})
}

func (r *demandRecord) regrade(gap string, severity recordpb.Grade) int64 {
	return r.add("red-lens-logic", recordpb.EventType_EVENT_TYPE_REGRADE, &recordpb.Regrade{
		GapId: proto.String(gap), Severity: severity.Enum(), Basis: proto.String("moved")})
}

// gapStep is one act on the fixture record and what the gap view must hold after it: each a query
// returning one row of one truth.
type gapStep struct {
	act   func()
	holds func() []string
}

// gapIs is a truth about gap: cond over its gap view row. A gap with no row is no truth.
func gapIs(gap, cond string) string {
	return `SELECT ` + cond + ` FROM "gap" WHERE "gap_id" = '` + gap + `'`
}

// gapStateSteps is a record that moves every column of gap_state through every marker trigger, one
// act per step, with the state each act leaves, including the three paths no archived run holds: the
// minter's re-read of the anchor its gap is about, an upheld position's remand, and an evidence demand
// carried to an heir by supersession.
func gapStateSteps(r *demandRecord) []gapStep {
	src, comp, doc := recordpb.CheckKind_CHECK_KIND_SOURCE, recordpb.CheckKind_CHECK_KIND_COMPUTATION, recordpb.CheckKind_CHECK_KIND_DOCUMENT
	supports, refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	var hMint, rClose, rRetire, rCarry, m1, m1Rule, dClose int64
	at := func(v *int64) string { return fmt.Sprint(*v) }
	none := func() []string { return nil }
	return []gapStep{
		// A check made before its anchor is placed meets nothing until the anchor is placed.
		{func() { r.mintOf("A", src, "C1") }, func() []string {
			return []string{gapIs("A", `"open" AND NOT "evidence_answered" AND "awaiting_evidence" AND "supersedes_count" = 0
			  AND "close_event" IS NULL AND "closed_by_bench" = 0 AND "material"`)}
		}},
		{func() { r.verify(reader, "C1", "", supports) }, func() []string { return []string{gapIs("A", `NOT "evidence_answered"`)} }},
		{func() { r.cite("C1", "") }, func() []string { return []string{gapIs("A", `"evidence_answered" AND NOT "awaiting_evidence"`)} }},
		// A recheck that does not hold replaces one that held.
		{func() { r.verify(reader, "C1", "", refutes) }, func() []string { return []string{gapIs("A", `NOT "evidence_answered" AND "awaiting_evidence"`)} }},
		// The heir carries A's demand (same evidence kind) and its mint closes A; the reader's refuting
		// check came before the heir's mint, so it meets nothing for H.
		{func() { hMint = r.mintOf("H", src, "", "A") }, func() []string {
			return []string{
				gapIs("A", `NOT "open" AND "superseded_by" = 'H' AND "supersede_event" = `+at(&hMint)+` AND NOT "closed_by_bench" AND NOT "bench_closed_last" AND NOT "awaiting_evidence"`),
				gapIs("H", `"open" AND "supersedes_count" = 1 AND NOT "evidence_answered" AND "awaiting_evidence"`),
			}
		}},
		// Another seat's reading of the ancestor's anchor meets nothing: the demand is the minter's.
		{func() { r.verify(other, "C1", "", supports) }, func() []string {
			return []string{gapIs("A", `NOT "evidence_answered"`), gapIs("H", `NOT "evidence_answered" AND "awaiting_evidence"`)}
		}},
		// The minter's re-read of the anchor, after the heir's mint: it meets A and, through the
		// inherited reach, H; its refutation unmeets both.
		{func() { r.verify(reader, "C1", "", supports) }, func() []string {
			return []string{gapIs("A", `"evidence_answered"`), gapIs("H", `"evidence_answered" AND NOT "awaiting_evidence"`)}
		}},
		{func() { r.verify(reader, "C1", "", refutes) }, func() []string {
			return []string{gapIs("A", `NOT "evidence_answered"`), gapIs("H", `NOT "evidence_answered" AND "awaiting_evidence"`)}
		}},
		// A retire of the ancestor's anchor after the heir's mint meets the heir.
		{func() { r.retire("C1") }, func() []string {
			return []string{gapIs("A", `"evidence_answered"`), gapIs("H", `"evidence_answered" AND NOT "awaiting_evidence"`)}
		}},
		// A computation placed as a gap's answer.
		{func() { r.mintOf("P", comp, "") }, func() []string { return []string{gapIs("P", `"awaiting_evidence"`)} }},
		{func() { r.compute("K1", "P") }, func() []string { return []string{gapIs("P", `"evidence_answered" AND NOT "awaiting_evidence"`)} }},
		// A reproduce of the anchor a computation gap is about, by its minter.
		{func() { r.compute("K2", "") }, none},
		{func() { r.mintOf("Q", comp, "K2") }, func() []string { return []string{gapIs("Q", `NOT "evidence_answered"`)} }},
		{func() { r.reproduce(reader, "K2", recordpb.Soundness_SOUNDNESS_SOUND) }, func() []string { return []string{gapIs("Q", `"evidence_answered"`)} }},
		// A citation placed as a gap's answer; the gap closed; its answer retired; a carry that does
		// not reset the retired answers; a judged close that does.
		{func() { r.mintOf("R", src, "") }, none},
		{func() { r.cite("C2", "R") }, func() []string {
			return []string{gapIs("R", `"evidence_answered" AND "answers_retired_event" IS NULL`)}
		}},
		{func() { rClose = r.close("R", "") }, func() []string {
			return []string{gapIs("R", `NOT "open" AND "close_event" = `+at(&rClose)+` AND "answers_retired_event" IS NULL AND NOT "awaiting_evidence"`)}
		}},
		{func() { rRetire = r.retire("C2") }, func() []string { return []string{gapIs("R", `"answers_retired_event" = `+at(&rRetire))} }},
		{func() { rCarry = r.close("R", "R") }, func() []string {
			return []string{gapIs("R", `"answers_retired_event" = `+at(&rRetire)+` AND "close_event" = `+at(&rCarry))}
		}},
		{func() { r.close("R", "") }, func() []string { return []string{gapIs("R", `"answers_retired_event" IS NULL`)} }},
		// A mint naming a closed gap closes nothing.
		{func() { r.mintOf("V", doc, "", "R") }, func() []string {
			return []string{gapIs("R", `"supersede_event" IS NULL AND "superseded_by" IS NULL`), gapIs("V", `"open" AND "supersedes_count" = 1 AND NOT "awaiting_evidence"`)}
		}},
		// The docket: two filings, the newer remanded (the older stands unruled again), then the
		// older ruled closing; a regrade moves the gap's grades and materiality.
		{func() { r.mintOf("D", doc, "") }, func() []string { return []string{gapIs("D", `"current_severity" = 'medium' AND "material"`)} }},
		{func() { r.regrade("D", recordpb.Grade_GRADE_LOW) }, func() []string {
			return []string{gapIs("D", `"current_severity" = 'low' AND "severity" = 'medium' AND NOT "material"`)}
		}},
		{func() { m1 = r.fileDocket("M1", "D") }, func() []string {
			return []string{gapIs("D", `"unruled_docket_motion" = 'M1' AND "unruled_docket_filed" = `+at(&m1))}
		}},
		{func() { r.fileDocket("M2", "D") }, func() []string { return []string{gapIs("D", `"unruled_docket_motion" = 'M2'`)} }},
		{func() { r.ruleDocket("M2", recordpb.Disposition_DISPOSITION_REMANDED) }, func() []string {
			return []string{gapIs("D", `"open" AND "unruled_docket_motion" = 'M1' AND "unruled_docket_filed" = `+at(&m1)+`
			  AND "docket_direction" = 'a reproduction on the shipped binary' AND NOT "remanded"`)}
		}},
		{func() { m1Rule = r.ruleDocket("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED) }, func() []string {
			return []string{gapIs("D", `NOT "open" AND "bench_close_event" = `+at(&m1Rule)+` AND "closed_by_bench" AND "bench_closed_last"
			  AND "close_event" IS NULL AND "unruled_docket_motion" IS NULL AND "unruled_docket_filed" IS NULL AND NOT "remanded"`)}
		}},
		// A docket remand with nothing pending.
		{func() { r.mintOf("E", doc, "") }, none},
		{func() { r.fileDocket("M3", "E") }, func() []string { return []string{gapIs("E", `"unruled_docket_motion" = 'M3' AND NOT "remanded"`)} }},
		{func() { r.ruleDocket("M3", recordpb.Disposition_DISPOSITION_REMANDED) }, func() []string {
			return []string{gapIs("E", `"open" AND "remanded" AND "docket_settled" = 'the claim as it stood' AND "unruled_docket_motion" IS NULL`)}
		}},
		// A red close after the bench's: whose word stands moves to red, who closed it does not.
		{func() { dClose = r.close("D", "") }, func() []string {
			return []string{gapIs("D", `"closed_by_bench" AND NOT "bench_closed_last" AND "close_event" = `+at(&dClose)+` AND "bench_close_event" = `+at(&m1Rule))}
		}},
		// An upheld position remands every gap it lists.
		{func() { r.mintOf("U", doc, "") }, func() []string { return []string{gapIs("U", `NOT "remanded"`)} }},
		{func() { r.upholdPosition("U-00000001", "U") }, func() []string {
			return []string{gapIs("U", `"remanded" AND "docket_direction" = 'source the figure' AND "docket_settled" IS NULL`)}
		}},
	}
}

// gapStateFirstClosures is each gap's first closing act as gap_state holds it, after every step of
// gapStateSteps: the act convergence_at and the window index read as the gap's closing.
func gapStateFirstClosures(t *testing.T, db *sql.DB) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT s."gap_id", COALESCE(e."type", 'open') FROM "gap_state" s LEFT JOIN "events" e ON e."id" = s."closed_event"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var g, typ string
		if err := rows.Scan(&g, &typ); err != nil {
			t.Fatal(err)
		}
		out[g] = typ
	}
	return out
}

// THE STORED GAP STATE IS THE RULES', AT EVERY EVENT, ON EVERY SHAPE. Each step's act leaves the
// state the fixture states for it, and the stored state, kept append by append through the marker
// triggers, is the refresh asked of every gap at once: the minter's re-read of its gap's anchor, an
// upheld position's remand, an evidence demand an heir carries from its ancestor, a check before its anchor,
// a failing recheck, a carry close after the answers retire, a supersession of a closed gap, an older
// docket standing again under a remand, and a bench closure red's later close overrules.
func TestTheStoredGapStateIsTheRules(t *testing.T) {
	r := newDemandRecord(t)
	for i, step := range gapStateSteps(r) {
		step.act()
		if d := gapStateHolds(t, r.db, step.holds()); d != "" {
			t.Errorf("after step %d the gap state is not the one the step leaves:\n%s", i, d)
		}
		if d := gapStateRefreshDiff(t, r.db); d != "" {
			t.Fatalf("after step %d the stored gap state is not the refresh of every gap:\n%s", i, d)
		}
	}
	// THE FIRST CLOSING ACT, of each arm: A's by the mint superseding it, R's by red's first close
	// (two closes follow it), D's by the bench (red's close follows it), the rest open.
	want := map[string]string{"A": "mint", "H": "open", "P": "open", "Q": "open", "R": "close", "V": "open", "D": "motion_rule", "E": "open", "U": "open"}
	if got := gapStateFirstClosures(t, r.db); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("each gap's first closing act = %v, want %v", got, want)
	}
	var rFirst bool
	if err := r.db.QueryRow(`SELECT s."closed_event" = (SELECT min(c."event_id") FROM "close" c WHERE c."gap_id" = 'R')
	    FROM "gap_state" s WHERE s."gap_id" = 'R'`).Scan(&rFirst); err != nil || !rFirst {
		t.Errorf("R's closed_event is not its first close: %v (%v)", rFirst, err)
	}
}

// EVERY MARKER TRIGGER IS NEEDED: with any one dropped, some step of the fixture record leaves a gap
// unrefreshed, so the stored state leaves the refresh of every gap, or a step's state does not hold.
// A marker no fixture misses is a marker nothing shows is right.
func TestEveryGapMarkerIsNeeded(t *testing.T) {
	markers := gapMarkers()
	if len(markers) != 12 {
		t.Errorf("%d marker triggers, the census below covers each: update this count with the census", len(markers))
	}
	for _, m := range markers {
		t.Run(m.table, func(t *testing.T) {
			r := newDemandRecord(t)
			if _, err := r.db.Exec(`DROP TRIGGER "gap_dirty_` + m.table + `"`); err != nil {
				t.Fatal(err)
			}
			for _, step := range gapStateSteps(r) {
				step.act()
				if gapStateRefreshDiff(t, r.db) != "" || gapStateHolds(t, r.db, step.holds()) != "" {
					return
				}
			}
			t.Errorf("with gap_dirty_%s dropped every step still leaves its state and the refresh of every gap: no fixture shows the marker is needed", m.table)
		})
	}
}

// A DOCKET RULING ANSWERS A FILING ALREADY ON THE RECORD. Ruled first and filed after, the ruling would
// change its gap only once the filing named it; neither order is writable.
func TestADocketRulingAnswersAFilingOnTheRecord(t *testing.T) {
	r := newDemandRecord(t)
	r.mintOf("D", recordpb.CheckKind_CHECK_KIND_DOCUMENT, "")
	ev := event(t, 90, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("M9"), Opinion: proto.String("stated"),
		Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{Disposition: recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED.Enum(),
			Principle: proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("no"), Settled: proto.String("s"), Final: proto.Bool(true)}}})
	if _, err := Insert(r.db, ev); err == nil || !strings.Contains(err.Error(), "no docket filing carries") {
		t.Errorf("a docket ruling on a motion never filed = %v, want refused", err)
	}

	// A position ruling's motion has no filing; it is ruled first, and a docket filing of its id is refused.
	r.upholdPosition("M8", "D")
	ev = event(t, 91, recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{MotionId: proto.String("M8"), Basis: proto.String("b"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("D")}}})
	if _, err := Insert(r.db, ev); err == nil || !strings.Contains(err.Error(), "a ruling already answers this motion id") {
		t.Errorf("a docket filing of a motion id already ruled = %v, want refused", err)
	}
	if d := gapStateHolds(t, r.db, []string{gapIs("D", `"open" AND "remanded" AND "docket_direction" = 'source the figure'
	    AND "bench_close_event" IS NULL AND "unruled_docket_motion" IS NULL`)}); d != "" {
		t.Errorf("after the refused writes D is not the upheld position's remand alone:\n%s", d)
	}
	if d := gapStateRefreshDiff(t, r.db); d != "" {
		t.Errorf("after the refused writes the stored gap state is not the refresh of every gap:\n%s", d)
	}
}

// THE FLUSH COMPILES NO LARGER THAN THE FOLDED VIEW DID. Preparing the flush insert compiles the whole
// refresh (gap_refresh, through its trigger), and every gap-changing append pays it once in each
// process. It is held to the size of the program the board used to compile on every read: the gap
// view that folded every gap from the bodies, whose SELECT * compiled to 1769 VDBE instructions
// (measured 2026-11-10 on modernc SQLite, its prepare a median 1.20 ms beside the flush's 1.29 ms
// at 1744 instructions), with the x1.2 allowance its timed guard had.
//
// COUNTED, NOT TIMED: EXPLAIN of the flush insert lists the program its trigger compiles, and the
// count is the same on a loaded box as on an idle one, where a timed ratio needs a reference
// statement as dear as the refresh and nothing left on the record is. A SQLite upgrade can move the
// count; re-measure against the folded view's text in git history (recordsql/testdata/oracle_gap.sql
// at c40fc671a) before moving the ceiling.
func TestTheFlushCompilesNoLargerThanTheFoldedViewDid(t *testing.T) {
	const foldedViewInstructions = 1769
	ceiling := foldedViewInstructions * 12 / 10
	r := newDemandRecord(t)
	rows, err := r.db.Query(`EXPLAIN INSERT INTO "gap_flush" ("id") VALUES (1)`)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for rows.Next() {
		n++
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("the flush compiles to %d instructions; ceiling %d", n, ceiling)
	if n < foldedViewInstructions/2 {
		t.Errorf("the flush compiles to %d instructions: EXPLAIN no longer lists the refresh its trigger runs, and the count measures nothing", n)
	}
	if n > ceiling {
		t.Errorf("the flush compiles to %d instructions, above %d (x1.2 the folded gap view's %d): the refresh costs a gap-changing append more than the board's read it replaced", n, ceiling, foldedViewInstructions)
	}
}
