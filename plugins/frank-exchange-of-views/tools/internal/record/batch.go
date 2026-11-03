package record

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Batch is many acts written in ONE transaction its caller holds, each act inside a savepoint of
// its own: a migration's replay, which writes a whole run's events and commits them once.
//
// ONE COMMIT, NOT ONE PER ACT. Under synchronous=FULL every commit is an fsync, and a replay that
// commits each of a run's events pays one per event — most of a migration's time on disk. The
// batch's commit is synced like any other, so a finished migration is as durable as before; a
// migration that fails is discarded whole, so nothing is lost by committing once.
//
// A REFUSED ACT LEAVES NOTHING, AS IT DOES THROUGH Append. An act writes its `events` row before its
// body, and a constraint on the body can fail after it; the savepoint is rolled back on any refusal,
// so the act's rows go with it and the batch takes the next act.
//
// ONLY A MIGRATION HOLDS ONE. The batch holds the write lock from BeginBatch to Commit, across
// whatever its caller does between acts; no seat may hold the record that way, so BeginBatch
// refuses unless Migrating is set. While a batch is open, the run's handle is its transaction's: a
// read through the handle — anything that takes the Run and opens the record — waits on the batch
// forever. Reads go through Querier.
type Batch struct {
	run  Run
	tx   *sql.Tx
	err  error // the failure that ended the batch; every later call returns it
	done bool  // committed or rolled back
}

// ErrBatchFailed is wrapped by every error a Batch returns once the batch itself has failed — a
// savepoint statement that did not run, or a transaction SQLite ended on its own — as distinct from
// a refusal of one act, after which the batch takes the next.
var ErrBatchFailed = errors.New("the batch's transaction failed")

// SavepointTestHook is a TEST HOOK: when set, it runs before each savepoint statement a Batch issues
// and an error it returns stands in for SQLite's. The failure it stands in for — the disk full, an
// I/O error, the transaction rolled back under the batch — cannot be produced on demand by any act a
// batch is handed. Nil outside a test.
var SavepointTestHook func(stmt string) error

// actSavepoint names the one savepoint a batch holds at a time: acts are written one after another,
// never nested.
const actSavepoint = "act"

// BeginBatch opens a batch on run's record. It refuses unless Migrating is set. The record is opened
// (and, on a fresh one, its schema applied in a transaction of its own) before the batch's
// transaction begins.
func BeginBatch(run Run) (*Batch, error) {
	if !Migrating {
		return nil, feov.Errorf(feov.Internal,
			"a batch is opened by migrate and by nothing else — it holds the record's write lock across its caller's code between acts, which no seat may do")
	}
	if !run.Valid() {
		return nil, feov.Errorf(feov.MissingField, "this batch names no run")
	}
	db, err := openRun(run)
	if err != nil {
		return nil, err
	}
	WriteBegins()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &Batch{run: run, tx: tx}, nil
}

// Append writes one act inside the batch, as Append would through a transaction of its own: the same
// guards, the same validation asked of the batch's transaction. A refusal of the act rolls its
// savepoint back and is returned; the batch takes the next act. An error wrapping ErrBatchFailed is
// the batch's own failure, and every later call returns it.
//
// id.Run must be the batch's run: a write addressed to another run is refused, never written here.
func (b *Batch) Append(id Identity, body proto.Message) (*Event, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.done {
		return nil, fmt.Errorf("this batch is already committed or rolled back")
	}
	if id.Run != b.run {
		return nil, feov.Errorf(feov.Internal,
			"this act is addressed to the run at %s, and the batch writes the run at %s — an act lands on the run it names or nowhere",
			id.Run.Dir(), b.run.Dir())
	}
	if err := b.savepoint("SAVEPOINT " + actSavepoint); err != nil {
		return nil, err
	}
	ev, err := appendAct(id, body, func() (*sql.Tx, error) { return b.tx, nil }, nil, nil, nil)
	if err != nil {
		if ferr := b.savepoint("ROLLBACK TO " + actSavepoint); ferr != nil {
			return ev, ferr
		}
		if ferr := b.savepoint("RELEASE " + actSavepoint); ferr != nil {
			return ev, ferr
		}
		return ev, err
	}
	if err := b.savepoint("RELEASE " + actSavepoint); err != nil {
		return ev, err
	}
	return ev, nil
}

// StampedBody is one body of an act a batch writes whole (AppendBodies), with the stamp its source
// event carried: each body is written under its own.
type StampedBody struct {
	Body proto.Message
	At   time.Time
}

// WrittenBody is one event an act a batch wrote whole (AppendBodies) wrote, and its events.id.
type WrittenBody struct {
	Event *Event
	ID    int64
}

// AppendBodies writes a reading's act whole inside the batch, as AppendReading wrote it live: its
// gap's Mint, the Anchor that places it and the reading, in that order, in one savepoint, each body
// validated with the act in hand (readingAct) — so the budget's admission of a citation reading's
// Mint and the reading's carrier_minted read the act, not a record the act is not on yet — and each
// stamped with its own At. A refusal of any body rolls the savepoint back whole and is returned; the
// batch takes the next act. It returns each event written and its events.id, read inside the
// savepoint after that body's insert, so a caller maps every source event to the event written from
// it and never reads the act's last id for its first.
func (b *Batch) AppendBodies(id Identity, bodies []StampedBody) ([]WrittenBody, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.done {
		return nil, fmt.Errorf("this batch is already committed or rolled back")
	}
	if id.Run != b.run {
		return nil, feov.Errorf(feov.Internal,
			"this act is addressed to the run at %s, and the batch writes the run at %s — an act lands on the run it names or nowhere",
			id.Run.Dir(), b.run.Dir())
	}
	act, err := readingActOf(bodies)
	if err != nil {
		return nil, err
	}
	if err := b.savepoint("SAVEPOINT " + actSavepoint); err != nil {
		return nil, err
	}
	was := Now
	defer func() { Now = was }()
	var out []WrittenBody
	for _, sb := range bodies {
		at := sb.At
		Now = func() time.Time { return at }
		ev, err := appendAct(id, sb.Body, func() (*sql.Tx, error) { return b.tx, nil }, nil, nil, act)
		if err == nil {
			var n int64
			if err = b.tx.QueryRow(`SELECT MAX("id") FROM "events"`).Scan(&n); err == nil {
				out = append(out, WrittenBody{Event: ev, ID: n})
				continue
			}
		}
		if ferr := b.savepoint("ROLLBACK TO " + actSavepoint); ferr != nil {
			return nil, ferr
		}
		if ferr := b.savepoint("RELEASE " + actSavepoint); ferr != nil {
			return nil, ferr
		}
		return nil, err
	}
	if err := b.savepoint("RELEASE " + actSavepoint); err != nil {
		return nil, err
	}
	return out, nil
}

// readingActOf is the reading's act its bodies are: a Mint, the Anchor whose id is its gap, and the
// Verify its own act minted that gap for (carrier_minted true, carried_by the gap), in that order.
func readingActOf(bodies []StampedBody) (*readingAct, error) {
	if len(bodies) == 3 {
		m, mok := bodies[0].Body.(*recordpb.Mint)
		a, aok := bodies[1].Body.(*recordpb.Anchor)
		v, vok := bodies[2].Body.(*recordpb.Verify)
		if mok && aok && vok && m.GetGapId() != "" && a.GetId() == m.GetGapId() &&
			v.GetCarriedBy() == m.GetGapId() && v.GetCarrierMinted() {
			return &readingAct{mint: m, anchor: a, verify: v}, nil
		}
	}
	return nil, feov.Errorf(feov.Internal,
		"an act a batch writes whole is a reading's: its gap's Mint, the Anchor whose id is that gap, and the Verify its act minted the gap for, in that order")
}

// savepoint runs one savepoint statement. A failure ends the batch: past it, the savepoints no
// longer bracket the acts, and SQLite may have ended the transaction itself — a later pair run
// outside BEGIN would commit each act on its own.
func (b *Batch) savepoint(stmt string) error {
	var err error
	if SavepointTestHook != nil {
		err = SavepointTestHook(stmt)
	}
	if err == nil {
		_, err = b.tx.Exec(stmt)
	}
	if err != nil {
		b.err = fmt.Errorf("%w: %s: %v", ErrBatchFailed, stmt, err)
		return b.err
	}
	return nil
}

// Querier is the batch's transaction, for the reads its caller makes between acts: they see the acts
// the batch has written, and they are the only reads that do not wait on it.
func (b *Batch) Querier() recordsql.Querier { return b.tx }

// Err is the batch's failure, if it has one: nil while every act's savepoint has held. A caller
// that records refused acts and carries on asks it at the end, so a failure that landed on any one
// act — whichever site wrote it — ends the replay rather than reading as a refusal.
func (b *Batch) Err() error { return b.err }

// Commit commits every act the batch wrote. A failed batch does not commit: its error is returned and
// the transaction rolled back.
func (b *Batch) Commit() error {
	if b.err != nil {
		_ = b.Rollback()
		return b.err
	}
	if b.done {
		return fmt.Errorf("this batch is already committed or rolled back")
	}
	b.done = true
	return b.tx.Commit()
}

// Rollback discards every act the batch wrote. It may be called any number of times, and after
// Commit it does nothing.
func (b *Batch) Rollback() error {
	if b.done {
		return nil
	}
	b.done = true
	// A failed batch's transaction may already be gone — SQLite ends it on a full disk or an I/O
	// error — so its rollback answers nothing the failure has not.
	if err := b.tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) && b.err == nil {
		return err
	}
	return nil
}
