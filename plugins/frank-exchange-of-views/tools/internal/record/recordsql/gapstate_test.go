package recordsql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// installGapOracle installs the folded gap view (testdata/oracle_gap.sql) on conn as TEMP view
// oracle_gap.
func installGapOracle(t *testing.T, ctx context.Context, conn *sql.Conn) {
	t.Helper()
	oracle, err := os.ReadFile("testdata/oracle_gap.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, `DROP VIEW IF EXISTS temp."oracle_gap";`+string(oracle)); err != nil {
		t.Fatal(err)
	}
}

// gapStateOracleDiff holds the thin gap view against the folded view it replaces, every column the
// thin view has, on one connection because the oracle is TEMP. It returns "" when they agree row for
// row, else what differs.
func gapStateOracleDiff(t *testing.T, db *sql.DB) string {
	t.Helper()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	installGapOracle(t, ctx, conn)
	cols := gapColumns(t, ctx, conn)
	list := strings.Join(cols, ", ")
	var out []string
	for _, q := range []struct{ what, sql string }{
		{"gap rows the oracle lacks", `SELECT ` + list + ` FROM "gap" EXCEPT SELECT ` + list + ` FROM "oracle_gap"`},
		{"oracle rows the gap view lacks", `SELECT ` + list + ` FROM "oracle_gap" EXCEPT SELECT ` + list + ` FROM "gap"`},
		{"row counts (gap, oracle) that differ", `SELECT a, b FROM (SELECT (SELECT count(*) FROM "gap") AS a, (SELECT count(*) FROM "oracle_gap") AS b) WHERE a <> b`},
		{"gaps named and not flushed", `SELECT "gap_id" FROM "gap_dirty" UNION ALL SELECT "id" FROM "gap_flush"`},
	} {
		rows, err := conn.QueryContext(ctx, q.sql)
		if err != nil {
			t.Fatal(err)
		}
		rc, _ := rows.Columns()
		for rows.Next() {
			v := make([]any, len(rc))
			p := make([]any, len(rc))
			for i := range v {
				p[i] = &v[i]
			}
			if err := rows.Scan(p...); err != nil {
				t.Fatal(err)
			}
			out = append(out, fmt.Sprintf("%s: %v", q.what, v))
		}
		rows.Close()
	}
	return strings.Join(out, "\n")
}

// gapColumns is the thin gap view's columns, quoted.
func gapColumns(t *testing.T, ctx context.Context, conn *sql.Conn) []string {
	t.Helper()
	rows, err := conn.QueryContext(ctx, `SELECT "name" FROM pragma_table_info('gap')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, `"`+c+`"`)
	}
	if len(cols) == 0 {
		t.Fatal("the gap view has no columns")
	}
	return cols
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

// gapStateSteps is a record that moves every column of gap_state through every marker trigger, one
// act per step, including the three paths no archived run holds: a reading carried by a gap, an
// upheld position's remand, and an evidence demand carried to an heir by supersession.
func gapStateSteps(r *demandRecord) []func() {
	src, comp, doc := recordpb.CheckKind_CHECK_KIND_SOURCE, recordpb.CheckKind_CHECK_KIND_COMPUTATION, recordpb.CheckKind_CHECK_KIND_DOCUMENT
	supports, refutes := recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS, recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES
	return []func(){
		// A check made before its anchor is placed meets nothing until the anchor is placed.
		func() { r.mintOf("A", src, "C1") },
		func() { r.verify(reader, "C1", "", supports) },
		func() { r.cite("C1", "") },
		// A recheck that does not hold replaces one that held.
		func() { r.verify(reader, "C1", "", refutes) },
		// The heir carries A's demand (same evidence kind) and its mint closes A.
		func() { r.mintOf("H", src, "", "A") },
		// A reading carried by the ancestor, after the heir's mint, by another seat: it meets A and,
		// through the inherited reach, H.
		func() { r.verify(other, "C1", "A", supports) },
		func() { r.verify(other, "C1", "A", refutes) },
		// A retire of the ancestor's anchor after the heir's mint meets the heir.
		func() { r.retire("C1") },
		// A computation placed as a gap's answer.
		func() { r.mintOf("P", comp, "") },
		func() { r.compute("K1", "P") },
		// A reproduce of the anchor a computation gap is about, by its minter.
		func() { r.compute("K2", "") },
		func() { r.mintOf("Q", comp, "K2") },
		func() { r.reproduce(reader, "K2", recordpb.Soundness_SOUNDNESS_SOUND) },
		// A citation placed as a gap's answer; the gap closed; its answer retired; a carry that does
		// not reset the retired answers; a judged close that does.
		func() { r.mintOf("R", src, "") },
		func() { r.cite("C2", "R") },
		func() { r.close("R", "") },
		func() { r.retire("C2") },
		func() { r.close("R", "R") },
		func() { r.close("R", "") },
		// A mint naming a closed gap closes nothing.
		func() { r.mintOf("V", doc, "", "R") },
		// The docket: two filings, the newer remanded (the older stands unruled again), then the
		// older ruled closing; a regrade moves the gap's grades and materiality.
		func() { r.mintOf("D", doc, "") },
		func() { r.regrade("D", recordpb.Grade_GRADE_LOW) },
		func() { r.fileDocket("M1", "D") },
		func() { r.fileDocket("M2", "D") },
		func() { r.ruleDocket("M2", recordpb.Disposition_DISPOSITION_REMANDED) },
		func() { r.ruleDocket("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED) },
		// A docket remand with nothing pending.
		func() { r.mintOf("E", doc, "") },
		func() { r.fileDocket("M3", "E") },
		func() { r.ruleDocket("M3", recordpb.Disposition_DISPOSITION_REMANDED) },
		// A red close after the bench's: whose word stands moves to red.
		func() { r.close("D", "") },
		// An upheld position remands every gap it lists.
		func() { r.mintOf("U", doc, "") },
		func() { r.upholdPosition("U-00000001", "U") },
	}
}

// THE STORED GAP STATE IS THE FOLDED VIEW'S, AT EVERY EVENT, ON THE SHAPES NO ARCHIVED RUN HOLDS. The
// archive sweep (migrate) holds the thin gap view equal to the oracle on every archived run; these
// fixtures move every stored column through every marker trigger, including a reading carried by a
// gap, an upheld position's remand, and an evidence demand an heir carries from its ancestor.
func TestTheStoredGapStateIsTheFoldedViews(t *testing.T) {
	r := newDemandRecord(t)
	for i, step := range gapStateSteps(r) {
		step()
		if d := gapStateOracleDiff(t, r.db); d != "" {
			t.Fatalf("after step %d the stored gap state is not the folded view's:\n%s", i, d)
		}
	}
	// The fixtures reach what they are for: an heir met through its ancestor's reading, an upheld
	// remand, a docket remand with nothing pending, a bench closure red's later close overruled, and
	// a closed gap whose answers were all retired.
	for _, c := range []struct{ what, sql string }{
		{"H answered through A's carried reading and retired anchor", `SELECT "evidence_answered" FROM "gap" WHERE "gap_id" = 'H'`},
		{"U remanded by the upheld position", `SELECT "remanded" AND "docket_direction" = 'source the figure' FROM "gap" WHERE "gap_id" = 'U'`},
		{"E remanded with nothing pending", `SELECT "remanded" FROM "gap" WHERE "gap_id" = 'E'`},
		{"D closed by the bench, red's word last", `SELECT "closed_by_bench" AND NOT "bench_closed_last" FROM "gap" WHERE "gap_id" = 'D'`},
		{"A closed by H's mint", `SELECT "superseded_by" = 'H' FROM "gap" WHERE "gap_id" = 'A'`},
		{"V's mint closed nothing", `SELECT "supersede_event" IS NULL FROM "gap" WHERE "gap_id" = 'R'`},
	} {
		var ok bool
		if err := r.db.QueryRow(c.sql).Scan(&ok); err != nil || !ok {
			t.Errorf("%s: got %v (%v)", c.what, ok, err)
		}
	}
}

// EVERY MARKER TRIGGER IS NEEDED: with any one dropped, the fixture record's gap state leaves the
// folded view's at some step. A marker no fixture misses is a marker nothing shows is right.
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
				step()
				if gapStateOracleDiff(t, r.db) != "" {
					return
				}
			}
			t.Errorf("with gap_dirty_%s dropped the stored gap state still equals the folded view at every step: no fixture shows the marker is needed", m.table)
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
	if d := gapStateOracleDiff(t, r.db); d != "" {
		t.Errorf("after the refused writes the stored gap state is not the folded view's:\n%s", d)
	}
}

// THE FLUSH COMPILES NO DEARER THAN THE FOLDED VIEW DID. Preparing the flush insert compiles the whole
// refresh (gap_refresh, through its trigger), and every gap-changing append pays it once; it is held
// to at most 1.2x preparing a read of the folded view the board used to pay on every read. Measured
// as the median ratio of interleaved prepare pairs, so load slows both arms alike.
func TestTheFlushCompilesNoDearerThanTheFoldedView(t *testing.T) {
	r := newDemandRecord(t)
	for _, step := range gapStateSteps(r) {
		step()
	}
	ctx := context.Background()
	conn, err := r.db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	installGapOracle(t, ctx, conn)
	prepare := func(q string) time.Duration {
		t0 := time.Now()
		st, err := conn.PrepareContext(ctx, q)
		d := time.Since(t0)
		if err != nil {
			t.Fatal(err)
		}
		st.Close()
		return d
	}
	const flush, folded = `INSERT INTO "gap_flush" ("id") VALUES (1)`, `SELECT * FROM "oracle_gap"`
	prepare(flush)
	prepare(folded)
	// One ratio per adjacent pair, its order alternating, and the median of the ratios: a stall on the
	// box lands on one pair, not on one arm.
	var ratios []float64
	var f, o []time.Duration
	for i := 0; i < 151; i++ {
		var a, b time.Duration
		if i%2 == 0 {
			a = prepare(flush)
			b = prepare(folded)
		} else {
			b = prepare(folded)
			a = prepare(flush)
		}
		f, o = append(f, a), append(o, b)
		ratios = append(ratios, float64(a)/float64(b))
	}
	sort.Float64s(ratios)
	ratio := ratios[len(ratios)/2]
	median := func(xs []time.Duration) time.Duration {
		sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
		return xs[len(xs)/2]
	}
	t.Logf("prepare medians: flush %v, folded view %v; median pair ratio x%.2f", median(f), median(o), ratio)
	if ratio > 1.2 {
		t.Errorf("preparing the flush takes x%.2f preparing the folded gap view (medians %v and %v): above x1.2, the refresh costs a gap-changing append more than the board's read it replaced", ratio, median(f), median(o))
	}
}
