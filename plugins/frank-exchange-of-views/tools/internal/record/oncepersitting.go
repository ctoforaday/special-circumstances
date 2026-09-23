package record

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE ONCE-PER-SITTING DUTY AND THE IDEMPOTENCY KEY NOW READ THE SAME FACT, AND THAT IS THE FIX.
//
// They were ONE query, then two, and the two drifted. The duty was enforced only by a singleton
// act's key colliding, so both halves of #1002's ruling came off one number: which act this IS (an
// identifier) and which sitting it belongs to (an attribution, and a repair is the sitting it
// repairs). A sitting-record repair moved the identifier and not the attribution, so inside a
// repair the ordinal advanced and nothing refused a second act. Measured through the CLI: two
// positions on one sitting, both attributed to `blue-respond #2`, and neither refused — invisible
// on the report, the changes view and the motion board, which render them under one heading.
//
// Splitting them gave each its own count of what opens a sitting, and the counts disagreed: the
// key's excluded a repair register and the events_w window did not. Both now read `sitting_id`,
// stamped once at the write (#1151), so the duty asks an equality and the key asks a rank over the
// same rows. The refusal's own words ("has already recorded a %s this sitting") are true of that
// reading, and a repair's acts belong to the sitting it repairs because the record says so rather
// than because two queries happened to agree.
//
// THE ACTS THIS COVERS ARE THE `singleton` SET — position, revision, verdict and spot_check —
// enumerated by TestTheOncePerSittingActsAreRefusedASecondTimeInOneSitting, which fails when a type
// joins that map without a row.

// oncePerSittingSQL is the standing act of this word in the seat's attributed sitting, or no row.
//
// IT ASKS FOR THE SITTING ITSELF. Every act carries the opening it belongs to as a field
// (schema.go), so "this sitting" is an equality on sitting_id rather than a window opened by
// recounting what a register means.
//
// THE COMPARISON IS `IS`, NOT `=`, AND THAT IS THE SEAT THAT NEVER OPENED ONE. Its acts carry NULL,
// and NULL = NULL is NULL — so a plain equality matches no row and the duty stops applying to
// exactly the seat with no sitting boundary to hide behind. `IS` is null-safe, which gives that
// seat the whole record as its sitting: the answer the old window reached by COALESCEing its start
// to 0, kept here rather than lost to a cleaner-looking operator.
//
// It returns the FIRST such act rather than a count: the refusal points the seat at the act that
// stands, and the correction chain is walked from there.
const oncePerSittingSQL = `SELECT e."key" FROM "events" e
     WHERE e."seat_id" = ?1 AND e."type" = ?2
       AND e."sitting_id" IS (SELECT max("id") FROM "sittings" WHERE "seat_id" = ?1)
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
