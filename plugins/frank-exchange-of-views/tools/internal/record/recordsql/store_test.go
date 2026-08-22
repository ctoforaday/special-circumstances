package recordsql

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func store(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// event builds a fixture act.
//
// THE ORDINAL IS THE KEY NOW, not a `seq` field. It used to stamp `Event.Seq`, the seat's position
// within its own shard — a counter whose last reader was replay's tiebreak sort, which insertion
// order replaced. Keeping the parameter and dropping it on the floor would have left every call
// site stating a number the record does not hold, so it names the act instead, which is a fact the
// record does keep and a UNIQUE index enforces.
func event(t *testing.T, ord int32, typ recordpb.EventType, body proto.Message) *recordpb.Event {
	t.Helper()
	ev := &recordpb.Event{
		SeatId: proto.String("red-merge-r1"),
		Round:  proto.Int32(1),
		Ts:     proto.String("2026-01-01T00:00:00Z"),
		Key:    proto.String(fmt.Sprintf("red-merge-r1:act:#%d", ord)),
	}
	got, err := recordpb.SetBody(ev, body)
	if err != nil {
		t.Fatal(err)
	}
	if got != typ {
		t.Fatalf("SetBody derived %v, want %v", got, typ)
	}
	ev.Type = &got
	return ev
}

// A MINT LANDS AS COLUMNS, NOT AS A BLOB.
//
// The whole reason to be here rather than in a file is that the fields are queryable. If a body
// arrived as an opaque value, the move would have bought transactions and ordering and forfeited
// joins, constraints and aggregates — which is the trade this test exists to refuse.
func TestABodyIsWrittenAsColumns(t *testing.T) {
	db := store(t)
	id, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Class:           proto.String("scope-creep"),
		Problem:         proto.String("an absence of findings is reported as an absence of risk"),
		AcceptanceCheck: proto.String("the claim names the search that produced it"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_HIGH.Enum(),
		Supersedes:      []string{"R0-4", "R0-9"},
	}))
	if err != nil {
		t.Fatal(err)
	}

	var class, problem, kind, likelihood string
	if err := db.QueryRow(`SELECT class, problem, check_kind, likelihood FROM mint WHERE event_id = ?`, id).
		Scan(&class, &problem, &kind, &likelihood); err != nil {
		t.Fatalf("the mint is not readable as columns: %v", err)
	}
	if class != "scope-creep" || kind != "document" || likelihood != "high" {
		t.Errorf("mint = (%q, %q, %q), want (scope-creep, document, high)", class, kind, likelihood)
	}

	// The enum went in as its WORD, so it joins to the vocabulary that explains it.
	var means string
	if err := db.QueryRow(`SELECT v.means FROM mint m JOIN enum_grade v ON v.value = m.likelihood WHERE m.event_id = ?`, id).Scan(&means); err != nil {
		t.Fatalf("a grade does not join to its meaning: %v", err)
	}
	if means == "" {
		t.Error("the grade joins to an empty meaning")
	}

	// A repeated field is rows, in order, so lineage is queryable rather than a string to split.
	rows, err := db.Query(`SELECT value FROM mint_supersedes WHERE event_id = ? ORDER BY ord`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 2 || got[0] != "R0-4" || got[1] != "R0-9" {
		t.Errorf("supersedes = %v, want [R0-4 R0-9]", got)
	}
}

// AN ABSENT FIELD IS NULL, NOT A ZERO.
//
// Every column is nullable for this reason. A mint with no severity is UNGRADED on that axis, and
// writing a zero would record a grade the seat never gave — the distinction the whole schema exists
// to keep, and the one a plain Go struct could not have expressed.
func TestAnAbsentFieldIsNull(t *testing.T) {
	db := store(t)
	id, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Class:           proto.String("c"),
		Problem:         proto.String("p"),
		AcceptanceCheck: proto.String("a"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_MEDIUM.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var severity *string
	if err := db.QueryRow(`SELECT severity FROM mint WHERE event_id = ?`, id).Scan(&severity); err != nil {
		t.Fatal(err)
	}
	if severity != nil {
		t.Errorf("severity = %q for a mint that never graded it — an ungraded axis reads as graded", *severity)
	}
}

// THE ORDERING PROBLEM IS GONE, and this is the assertion the whole move was for.
//
// Filing and ruling were written by different seats into different files, and replay merged them by
// timestamp — so a ruling could replay BEFORE its filing. record/motion.go records that in capitals
// and works around it with a second pass. Here the state is simply unwritable.
func TestARulingCannotPrecedeItsFiling(t *testing.T) {
	db := store(t)
	rule := event(t, 0, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_ACCEPTED},
	})
	if _, err := Insert(db, rule); err == nil {
		t.Fatal("a ruling was recorded against a motion that does not exist — this is the ordering " +
			"hazard the sharded log could not refuse, and the reason Motions() needs a second pass")
	}

	// The gap the motion contests has to exist: `motion_grade.gap_id` references mint now.
	mintGap(t, db, 2, "R1-1")

	filing := event(t, 1, recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R1-1")}},
	})
	if _, err := Insert(db, filing); err != nil {
		t.Fatalf("the filing was refused: %v", err)
	}
	if _, err := Insert(db, rule); err != nil {
		t.Fatalf("the ruling was refused after its filing existed: %v", err)
	}

	// And the join it enables: the gap id lives on the filing, the verdict on the ruling, and one
	// query pairs them. That join is hand-written at eight readers in the file-backed record.
	var gap, verdict string
	if err := db.QueryRow(`
		SELECT g.gap_id, r.grade
		FROM motion_rule r
		JOIN motion m       ON m.motion_id = r.motion_id
		JOIN motion_grade g ON g.event_id  = m.event_id
		WHERE r.motion_id = 'M1'`).Scan(&gap, &verdict); err != nil {
		t.Fatalf("the filing/ruling join does not resolve: %v", err)
	}
	if gap != "R1-1" || verdict != "accepted" {
		t.Errorf("join = (%q, %q), want (R1-1, accepted)", gap, verdict)
	}
}

// A RULE THAT SPANS TWO FIELDS IS ENFORCED BY THE DATABASE, not only by the tool.
//
// "closed_with_regression requires a successor" cannot be NOT NULL — the column must stay nullable
// for every other closure class — so it lived as an `if` in validate. That left the DATABASE
// willing to store the exact state the tool refused, and for a record that is EVIDENCE the gap is
// real: anything writing SQL directly could create it, and the row would look legitimate.
func TestARegressionClosureMustNameItsSuccessor(t *testing.T) {
	db := store(t)
	// Every gap named below, INCLUDING the successor: `close.successor` references mint.gap_id now,
	// so "lineage never drops" no longer means "the field is non-empty". A successor naming a gap
	// that does not exist carried the remainder forward to an id no board ever had.
	for i, id := range []string{"R1-1", "R1-2", "R1-3", "R2-1"} {
		mintGap(t, db, int32(10+i), id)
	}
	mk := func(seq int32) *recordpb.Event {
		return event(t, seq, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
			GapId:        proto.String("R1-1"),
			ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED_WITH_REGRESSION.Enum(),
			Prose:        proto.String("repaired, and the retry path broke"),
		})
	}
	if _, err := Insert(db, mk(0)); err == nil {
		t.Fatal("a regression closure with no successor was stored — the thread from the defect to " +
			"the gap now carrying it is lost, and the repair_regression denominator counts a closure " +
			"that resolved nothing")
	}

	ok := event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-2"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED_WITH_REGRESSION.Enum(),
		Successor:    proto.String("R2-1"),
		Prose:        proto.String("repaired, and R2-1 carries the regression"),
	})
	if _, err := Insert(db, ok); err != nil {
		t.Fatalf("a regression closure that NAMES its successor was refused: %v", err)
	}

	// And an ordinary closure still needs no successor, or the rule is a wall rather than a gate.
	plain := event(t, 2, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-3"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED.Enum(),
		Prose:        proto.String("verified at the leaf"),
	})
	if _, err := Insert(db, plain); err != nil {
		t.Fatalf("an ordinary closure was refused for having no successor: %v", err)
	}
}

// THE ROUND TRIP MUST PRESERVE PRESENCE, not merely values.
//
// Every field is optional, so "the seat set this" and "the seat did not" are different states and
// the record's whole typing rests on telling them apart. A round trip that returns zeros for absent
// fields would collapse the two on the way home — a mint with no severity coming back GRADED, an
// outcome with no verdict coming back as a verdict — which is the defect the schema exists to
// remove, reintroduced at the last step.
func TestAnEventSurvivesTheRoundTripWithItsAbsencesIntact(t *testing.T) {
	db := store(t)
	original := &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Class:           proto.String("scope-creep"),
		Problem:         proto.String("an absence of findings is reported as an absence of risk"),
		AcceptanceCheck: proto.String("the claim names the search that produced it"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_HIGH.Enum(),
		Supersedes:      []string{"R0-4", "R0-9"},
		// severity and complexity_cost are deliberately NOT set: the axes this gap was never
		// graded on, which must come back ungraded.
	}
	if _, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, original)); err != nil {
		t.Fatal(err)
	}

	evs, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("read %d events, want 1", len(evs))
	}
	got, ok := recordpb.BodyAs[*recordpb.Mint](evs[0])
	if !ok {
		t.Fatal("the event came back with no Mint body")
	}
	if !proto.Equal(original, got) {
		t.Errorf("the mint did not survive the round trip:\n before %v\n after  %v", original, got)
	}
	// Stated separately, because proto.Equal treats unset and zero as equal for scalars and would
	// not catch a severity that came home as GRADE_UNSPECIFIED.
	if got.Severity != nil {
		t.Errorf("severity came back set (%v) for an axis the seat never graded", got.GetSeverity())
	}
	if evs[0].GetType() != recordpb.EventType_EVENT_TYPE_MINT {
		t.Errorf("type = %v", evs[0].GetType())
	}
}

// EVENTS COME BACK IN THE ORDER THEY WERE RECORDED, from one sequence rather than a merge.
//
// The shard format sorted by timestamp because nothing else spanned the files, and that sort IS
// the ordering hazard. Here `id` is assigned at insert, so read order is record order.
func TestEventsReadBackInRecordOrder(t *testing.T) {
	db := store(t)
	for i, seat := range []string{"blue-respond-r1", "red-merge-r1", "judge-r1"} {
		ev := event(t, int32(i), recordpb.EventType_EVENT_TYPE_POSITION, &recordpb.Position{
			Text: proto.String(seat + " speaks"),
		})
		ev.SeatId = proto.String(seat)
		// Every seat stamps the SAME instant. Under the old merge this was the ambiguous case that
		// needed a tiebreak; here the sequence is the database's and the clock does not decide.
		ev.Ts = proto.String("2026-01-01T00:00:00Z")
		if _, err := Insert(db, ev); err != nil {
			t.Fatal(err)
		}
	}
	evs, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, e := range evs {
		order = append(order, e.GetSeatId())
	}
	want := []string{"blue-respond-r1", "red-merge-r1", "judge-r1"}
	for i := range want {
		if i >= len(order) || order[i] != want[i] {
			t.Fatalf("read order = %v, want %v — identical timestamps must not decide the sequence", order, want)
		}
	}
}

// THE FOLD IS A VIEW, AND THE BOARD IS A QUERY.
//
// "Which gaps are open" is a fold over the event stream, and in the file-backed record every
// consumer folds it again with its own idea of what closing means — which is where `filed > ruled`
// came to compute `0 > 0` forever and a dispute counter sat at zero through every run. A fold
// nobody can see is a fold nobody checks.
func TestTheBoardIsAQuery(t *testing.T) {
	db := store(t)
	mint := func(seq int32, id string) {
		t.Helper()
		if _, err := Insert(db, event(t, seq, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
			GapId:           proto.String(id),
			Class:           proto.String("c"),
			Problem:         proto.String("p"),
			AcceptanceCheck: proto.String("a"),
			CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
			Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
		})); err != nil {
			t.Fatal(err)
		}
	}
	mint(0, "R1-1")
	mint(1, "R1-2")
	if _, err := Insert(db, event(t, 2, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-2"),
		ClosureClass: recordpb.Disposition_DISPOSITION_RISK_ACCEPTED.Enum(),
		Prose:        proto.String("the fix costs more than the defect"),
	})); err != nil {
		t.Fatal(err)
	}

	var open, closed int
	if err := db.QueryRow(`SELECT open_gaps, closed_gaps FROM board_counts`).Scan(&open, &closed); err != nil {
		t.Fatalf("the board does not count: %v", err)
	}
	if open != 1 || closed != 1 {
		t.Errorf("board_counts = (%d open, %d closed), want (1, 1)", open, closed)
	}

	// The closure's reason travels with the gap, so the closure index is a SELECT rather than a
	// second walk that has to agree with the first about what closed means.
	var class, why string
	if err := db.QueryRow(`
		SELECT g."closure_class", v."means"
		FROM "gap" g JOIN "enum_disposition" v ON v."value" = g."closure_class"
		WHERE g."gap_id" = 'R1-2'`).Scan(&class, &why); err != nil {
		t.Fatalf("a closed gap does not join to why it closed: %v", err)
	}
	if class != "risk_accepted" || why == "" {
		t.Errorf("closure = (%q, %q)", class, why)
	}

	// And an open gap reports no closure rather than an empty one — the distinction a fold that
	// defaults to "" cannot make.
	var cc *string
	if err := db.QueryRow(`SELECT "closure_class" FROM "gap" WHERE "gap_id" = 'R1-1'`).Scan(&cc); err != nil {
		t.Fatal(err)
	}
	if cc != nil {
		t.Errorf("an open gap reports closure_class %q", *cc)
	}
}

// AN UNRULED MOTION IS VISIBLE WITHOUT REPLAYING ANYTHING.
//
// `sitting.go` refuses a seat's PASS while a motion it rules is unanswered, and it computes that by
// folding the stream. Here it is a column, and the filing/ruling join comes with it.
func TestAnUnruledMotionIsAColumn(t *testing.T) {
	db := store(t)
	mintGap(t, db, 10, "R1-1")
	mintGap(t, db, 11, "R1-2")
	file := func(seq int32, id, gap string) {
		t.Helper()
		if _, err := Insert(db, event(t, seq, recordpb.EventType_EVENT_TYPE_MOTION, &recordpb.Motion{
			MotionId: proto.String(id),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String(gap)}},
		})); err != nil {
			t.Fatal(err)
		}
	}
	file(0, "M1", "R1-1")
	file(1, "M2", "R1-2")
	if _, err := Insert(db, event(t, 2, recordpb.EventType_EVENT_TYPE_MOTION_RULE, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_ACCEPTED},
	})); err != nil {
		t.Fatal(err)
	}

	var id, gap string
	if err := db.QueryRow(`SELECT motion_id, gap_id FROM motion_state WHERE unruled`).Scan(&id, &gap); err != nil {
		t.Fatalf("the unruled motion is not visible: %v", err)
	}
	if id != "M2" || gap != "R1-2" {
		t.Errorf("unruled = (%q, %q), want (M2, R1-2)", id, gap)
	}
}

// A DEFERRED GAP IS STILL OPEN, AND THE VIEW ASKS THE VOCABULARY RATHER THAN GUESSING.
//
// This is the defect that motivated typing `disposition`, driven through the write path and the
// projection a reader actually queries. The old predicate was "everything except `carried`
// closes", so this test would have passed for `carried` and silently failed for every deferring
// word added afterwards — which is what happened.
func TestABenchDispositionClosesTheGapOnlyIfTheVocabularySaysSo(t *testing.T) {
	for _, c := range []struct {
		as       recordpb.Disposition
		wantOpen bool
		why      string
	}{
		{recordpb.Disposition_DISPOSITION_CARRIED, true, "carried defers the question to a later round with a stated direction; the gap survives"},
		{recordpb.Disposition_DISPOSITION_RISK_ACCEPTED, false, "the risk is taken knowingly, with the argument on the record — there is nothing further to adjudicate"},
		{recordpb.Disposition_DISPOSITION_REBUTTAL_SUSTAINED, false, "blue's rebuttal held; nothing was repaired because nothing needed to be"},
	} {
		t.Run(recordpb.Word(c.as), func(t *testing.T) {
			db := store(t)
			if _, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
				GapId:           proto.String("R1-1"),
				Class:           proto.String("c"),
				Problem:         proto.String("p"),
				AcceptanceCheck: proto.String("a"),
				CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
				Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
				Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
			})); err != nil {
				t.Fatal(err)
			}
			if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_OPINION, &recordpb.Opinion{
				GapId:       proto.String("R1-1"),
				Disposition: c.as.Enum(),
				Principle:   proto.String("correctness over economy"),
				Tension:     proto.String("the repair costs a round"),
				ReviewFlag:  proto.String("no"),
				Rationale:   proto.String("stated"),
			})); err != nil {
				t.Fatalf("the record refused a disposition the bench is instructed to use: %v", err)
			}

			var open bool
			if err := db.QueryRow(`SELECT "open" FROM "gap" WHERE "gap_id" = 'R1-1'`).Scan(&open); err != nil {
				t.Fatal(err)
			}
			if open != c.wantOpen {
				t.Errorf("a %s ruling left the gap open=%v, want %v — %s.\n\nA gap the bench DEFERRED reading as closed retires work nobody decided to drop; a gap the bench ENDED reading as open keeps a settled question on the board forever.",
					recordpb.Word(c.as), open, c.wantOpen, c.why)
			}
		})
	}
}

// A MERGE MAY CLOSE AND MAY NOT CARRY, AND THE DATABASE IS WHAT SAYS SO.
//
// `merge close` refuses `carried` in Go too, with a message that teaches. This asks the weaker but
// more important question: does the constraint hold against something writing the row DIRECTLY?
// The record is EVIDENCE, and an invariant that lives only in the CLI is one that anything else
// bypasses — which for a file-backed record was every invariant there was.
//
// The admitted words are expanded from the vocabulary's own `closes` annotation, so this cannot
// drift: adding a second deferring disposition tightens the CHECK with nobody editing it.
func TestAMergeCannotCloseAGapByCarryingIt(t *testing.T) {
	db := store(t)
	mintGap(t, db, 10, "R1-1")
	if _, err := Insert(db, event(t, 0, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CARRIED.Enum(),
		Prose:        proto.String("deferring, from the wrong seat"),
	})); err == nil {
		t.Fatal("the database accepted `merge close --as carried` — a close asserts a verified repair, and \"I repaired it by carrying it\" is not one. Deferring is the bench's decision, and a merge that can record it produces a gap that reads as closed with no repair behind it")
	}

	// The same column still takes a word that DOES close, so the CHECK is not simply refusing
	// everything — a constraint that admits nothing reads as strict and is broken.
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_RISK_ACCEPTED.Enum(),
		Prose:        proto.String("the fix costs more than the defect"),
	})); err != nil {
		t.Fatalf("the CHECK refuses a legitimate closure: %v", err)
	}
}

// EVERY WORD IN THE VOCABULARY ANSWERS WHETHER IT ENDS THE GAP.
//
// The column is NOT NULL and the schema refuses a partly-annotated set, so this cannot fail while
// the schema builds — which is the point, and is why it asserts the ANSWERS rather than only the
// presence. A test that checked `closes IS NOT NULL` would be asserting that NOT NULL works.
func TestTheVocabularySaysWhichWordsEndAGap(t *testing.T) {
	db := store(t)
	rows, err := db.Query(`SELECT "value", "closes" FROM "enum_disposition" ORDER BY "value"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]bool{}
	for rows.Next() {
		var word string
		var closes bool
		if err := rows.Scan(&word, &closes); err != nil {
			t.Fatal(err)
		}
		got[word] = closes
	}
	want := map[string]bool{
		"amends_prior":             true,
		"carried":                  false,
		"closed":                   true,
		"closed_with_regression":   true,
		"rebuttal_sustained":       true,
		"risk_accepted":            true,
		"routed_to_infrastructure": true,
	}
	if len(got) != len(want) {
		t.Fatalf("the vocabulary has %d words and this test knows %d — a new disposition must be added HERE with its answer, which is the whole reason the annotation is not defaulted", len(got), len(want))
	}
	for word, w := range want {
		if got[word] != w {
			t.Errorf("%q closes=%v, want %v", word, got[word], w)
		}
	}
}

// mintGap puts a real gap on the board, because every gap_id in the schema is a foreign key onto
// `mint.gap_id` now.
//
// That constraint broke five fixtures in this file when it landed, and each break was the same
// admission: they closed, ruled on and filed motions against gaps that had never been minted. The
// file-backed record met that case with a `missingGap` ANOMALY — a defect found on read, per
// reader, after the fact, which is how a whole run's bench closures once vanished into a list
// nobody displayed. The fixtures could be written that way because nothing refused them.
func mintGap(t *testing.T, db *sql.DB, seq int32, gapID string) {
	t.Helper()
	if _, err := Insert(db, event(t, seq, recordpb.EventType_EVENT_TYPE_MINT, &recordpb.Mint{
		GapId:           proto.String(gapID),
		Class:           proto.String("c"),
		Problem:         proto.String("p"),
		AcceptanceCheck: proto.String("a"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordpb.Grade_GRADE_HIGH.Enum(),
		Impact:          recordpb.Grade_GRADE_MEDIUM.Enum(),
	})); err != nil {
		t.Fatal(err)
	}
}

// A GAP THAT WAS NEVER MINTED CANNOT BE CLOSED, RULED ON, OR CONTESTED.
//
// The file-backed record met this case with a `missingGap` ANOMALY: the write succeeded, and the
// dangling reference was discovered later, by whichever reader happened to look, if any did. That
// is how a whole sitting's bench closures once vanished — the events were recorded correctly, the
// replay dropped them, and every projection downstream reported a board that had never existed.
//
// As a foreign key there is no anomaly to discover, because there is no row. That is the shape of
// the whole migration in one assertion: the invariant moves from something a reader might notice
// to something a writer cannot do.
func TestAnUnmintedGapCannotBeActedOn(t *testing.T) {
	for _, c := range []struct {
		what string
		body proto.Message
		typ  recordpb.EventType
	}{
		{"closed", &recordpb.Close{
			GapId:        proto.String("R9-9"),
			ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED.Enum(),
			Prose:        proto.String("closing a gap nobody minted"),
		}, recordpb.EventType_EVENT_TYPE_CLOSE},
		{"ruled on", &recordpb.Opinion{
			GapId:       proto.String("R9-9"),
			Disposition: recordpb.Disposition_DISPOSITION_RISK_ACCEPTED.Enum(),
			Principle:   proto.String("p"),
			Tension:     proto.String("t"),
			ReviewFlag:  proto.String("no"),
			Rationale:   proto.String("r"),
		}, recordpb.EventType_EVENT_TYPE_OPINION},
		{"contested by a grade motion", &recordpb.Motion{
			MotionId: proto.String("M1"),
			Subject:  recordpb.MotionSubject_MOTION_SUBJECT_GRADE.Enum(),
			Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("R9-9")}},
		}, recordpb.EventType_EVENT_TYPE_MOTION},
	} {
		t.Run(c.what, func(t *testing.T) {
			db := store(t)
			if _, err := Insert(db, event(t, 0, c.typ, c.body)); err == nil {
				t.Errorf("a gap that was never minted was %s — the reference dangles, and every count "+
					"derived from the board now includes an act against something that does not exist", c.what)
			}
		})
	}

	// And the same acts succeed once the gap is real, or the constraint is a wall rather than a gate.
	db := store(t)
	mintGap(t, db, 0, "R9-9")
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R9-9"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED.Enum(),
		Prose:        proto.String("the repair was verified at the leaf"),
	})); err != nil {
		t.Fatalf("a real gap could not be closed: %v", err)
	}
}

// A SUCCESSOR MUST BE A GAP, NOT A STRING THAT LOOKS LIKE ONE.
//
// "Lineage never drops" was enforced as "the field is not empty" — a rule a typo satisfies. The
// closure recorded a successor, the successor pointed at nothing, and the unresolved remainder was
// carried forward to a gap id no board has ever had. The CHECK and the foreign key are both needed
// and they say different things: one that a regression closure NAMES a successor, the other that
// what it names EXISTS.
func TestASuccessorMustBeAGapThatExists(t *testing.T) {
	db := store(t)
	mintGap(t, db, 0, "R1-1")

	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED_WITH_REGRESSION.Enum(),
		Successor:    proto.String("R2-7"),
		Prose:        proto.String("repaired, and R2-7 carries the regression"),
	})); err == nil {
		t.Fatal("a closure named R2-7 as carrying its regression forward and no such gap exists — the CHECK is satisfied, the lineage is broken, and the repair_regression denominator counts a closure that resolved nothing")
	}

	mintGap(t, db, 2, "R2-7")
	if _, err := Insert(db, event(t, 3, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED_WITH_REGRESSION.Enum(),
		Successor:    proto.String("R2-7"),
		Prose:        proto.String("repaired, and R2-7 carries the regression"),
	})); err != nil {
		t.Fatalf("a closure naming a REAL successor was refused: %v", err)
	}
}

// AN ABSENT FLAG MUST NOT BE STORED AS THE EMPTY STRING.
//
// `merge close` set `successor` unconditionally — `proto.String(seat.Str(cmd, flags.SupersededBy))`
// — so an ordinary closure recorded `successor = ''`. Once successor referenced `mint.gap_id` that
// became a hard refusal (no gap is named ''), which is how it was found; BEFORE the reference
// existed the same rows were written and read as a closure whose successor was the empty gap.
//
// This is the storage half of that guarantee: an unset field lands as NULL and comes back unset,
// so "the seat never said" and "the seat said nothing" stay different facts all the way down. The
// CLI half is TestAnAbsentFlagIsNotWrittenAsEmpty in internal/cli.
func TestAnUnsetOptionalFieldIsNullNotEmpty(t *testing.T) {
	db := store(t)
	mintGap(t, db, 0, "R1-1")
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_CLOSE, &recordpb.Close{
		GapId:        proto.String("R1-1"),
		ClosureClass: recordpb.Disposition_DISPOSITION_CLOSED.Enum(),
		Prose:        proto.String("the repair was verified at the leaf"),
	})); err != nil {
		t.Fatal(err)
	}

	// NULL, not ''. A '' here would satisfy the foreign key only because SQLite skips the check on
	// NULL — so the column reading '' is precisely the state that made every close fail.
	var successor *string
	if err := db.QueryRow(`SELECT "successor" FROM "close" WHERE "gap_id" = 'R1-1'`).Scan(&successor); err != nil {
		t.Fatal(err)
	}
	if successor != nil {
		t.Errorf("an unpassed --superseded-by stored %q — an absent flag written as a value is a fact "+
			"the seat never stated, and here it is a reference to a gap that cannot exist", *successor)
	}

	// And it survives the round trip as ABSENT rather than as an empty string.
	evs, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	var closes int
	for _, e := range evs {
		c, ok := recordpb.BodyAs[*recordpb.Close](e)
		if !ok {
			continue
		}
		closes++
		if c.Successor != nil {
			t.Errorf("the closure read back with successor set to %q, want unset", c.GetSuccessor())
		}
	}
	if closes != 1 {
		t.Fatalf("%d closures read back, want 1 — an empty traversal would pass the assertion above", closes)
	}
}

// THE DRIVER MUST BE REGISTERED BY THE PACKAGE THAT OPENS THE DATABASE, not by its tests.
//
// `_ "modernc.org/sqlite"` lived in schema_test.go, so `database/sql` had a registered driver
// throughout the suite and NONE in the shipped binary: every test passed and the first real
// `merge register` failed with `unknown driver "sqlite"`. A blank import is invisible to the
// compiler's unused check, which is why the wrong file was good enough.
//
// A test cannot catch that by opening a database — the test binary has the import either way. What
// it can do is ask the REGISTRY, which is process-global and populated by whoever imported the
// driver. It is a weak check on its own and it is paired with the strong one: cmd/feov-record is
// driven end to end in the cli suite, which is what actually failed.
func TestTheSqliteDriverIsRegistered(t *testing.T) {
	for _, d := range sql.Drivers() {
		if d == "sqlite" {
			return
		}
	}
	t.Fatalf("no \"sqlite\" driver is registered (have %v) — if this fails the import moved out of "+
		"the package that opens the database, and the binary will fail at the first write while the "+
		"suite stays green", sql.Drivers())
}
