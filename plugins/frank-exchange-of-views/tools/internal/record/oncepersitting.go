package record

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE ONCE-PER-SITTING DUTY ASKS THE ATTRIBUTED SITTING, AND THE IDEMPOTENCY KEY COUNTS TURNS.
//
// They were ONE query. A singleton act's key is `<seat>:<word>:#<register count>` (deriveKey) and
// the duty was enforced only by that key colliding, so the two halves of #1002's ruling were being
// read off one number: which act this IS (an identifier, and a turn) and which sitting it belongs to
// (an attribution, and a repair is the sitting it repairs). A sitting-record repair moves the turn
// count and not the attribution, so inside a repair the ordinal advanced and nothing refused a
// second act. Measured through the CLI: two positions on one sitting, both attributed to
// `blue-respond #2`, and neither refused — invisible on the report, the changes view and the motion
// board, which now render them under one heading.
//
// So the key keeps counting turns — a repair's acts must not collide with the keys of the sitting it
// repairs — and the DUTY asks record.ActClock's question in SQL: has this seat already filed an act
// of this word since the register that opened the sitting its acts belong to. The refusal's own
// words ("has already recorded a %s this sitting") are true of that reading and were not true of the
// other.
//
// THE ACTS THIS COVERS ARE THE `singleton` SET — position, revision, verdict and spot_check —
// enumerated by TestTheOncePerSittingActsAreRefusedASecondTimeInOneSitting, which fails when a type
// joins that map without a row.

// oncePerSittingSQL is the standing act of this word in the seat's attributed sitting, or no row.
//
// The window opens at the seat's latest register that OPENS A SITTING, which is
// openingRegistersOfSeatSQL — the one SQL spelling of opensASitting, shared with correction.go's
// pair of subqueries. A seat with no opening register has no earlier sitting to borrow from, so
// COALESCE gives it the whole record as its sitting, exactly as seatDidThisSitting reads it.
//
// It returns the FIRST such act rather than a count: the refusal points the seat at the act that
// stands, and the correction chain is walked from there.
const oncePerSittingSQL = `SELECT e."key" FROM "events" e
     WHERE e."seat_id" = ?1 AND e."type" = ?2
       AND e."id" > COALESCE((SELECT max("id") FROM (` + openingRegistersOfSeatSQL + `)), 0)
     ORDER BY e."id" LIMIT 1`

// requireOncePerSitting refuses a second singleton act of this word attributed to the seat's current
// sitting. It runs inside the writing transaction, which holds the write lock from its BEGIN, so the
// read and the insert cannot straddle another write.
//
// A CORRECTION NEVER REACHES IT: appendCorrected writes the replacement itself, on the corrected
// act's key chain, so the seat that files its position and then corrects it is filing one act.
func requireOncePerSitting(q interface {
	QueryRow(string, ...any) *sql.Row
}, seatID string, typ recordpb.EventType, body proto.Message) error {
	if !singleton[typ] {
		return nil
	}
	word := recordpb.Word(typ)
	var stands string
	switch err := q.QueryRow(oncePerSittingSQL, seatID, word).Scan(&stands); {
	case errors.Is(err, sql.ErrNoRows):
		return nil
	case err != nil:
		return fmt.Errorf("record: asking whether %s has already recorded a %s this sitting: %w", seatID, word, err)
	}
	// THE TAIL DEPENDS ON THE TIER, as it does on every first-wins refusal: a correctable act gets
	// the key of the act that stands and the invocation that corrects it, and one that cannot be
	// corrected is superseded instead.
	tail := "If the first was wrong, say so in the act that supersedes it; the record is append-only and both stay visible"
	if Correctable(typ, body) {
		tail = correctionPointer(q, stands)
	}
	return feov.Errorf(feov.Validation,
		"record: %s has already recorded a %s this sitting, and it is a once-per-sitting act — "+
			"the record keeps your first one rather than quietly replacing it. %s",
		seatID, word, tail)
}
