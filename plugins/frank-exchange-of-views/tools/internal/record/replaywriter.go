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

// Replay is the migration's writer: a run's acts written in ONE transaction its caller holds, each
// act inside a savepoint of its own, each stamped with the time its caller hands it, and each held
// to the shape every writer is held to (validate) and to nothing else. The act judgements (judgeAct)
// are the live write's: an archived act is what a seat DID, and the replay writes it as it was done.
//
// ONE COMMIT, NOT ONE PER ACT. Under synchronous=FULL every commit is an fsync, and a replay that
// commits each of a run's events pays one per event — most of a migration's time on disk. The
// replay's commit is synced like any other, so a finished migration is as durable as before; a
// migration that fails is discarded whole, so nothing is lost by committing once.
//
// A REFUSED ACT LEAVES NOTHING, AS IT DOES THROUGH Append. An act writes its `events` row before its
// body, and a constraint on the body can fail after it; the savepoint is rolled back on any refusal,
// so the act's rows go with it and the replay takes the next act.
//
// ONLY THE MIGRATION HOLDS ONE. The replay holds the write lock from BeginReplay to Commit, across
// whatever its caller does between acts, which no seat may do; scripts/pluginparity refuses a
// reference to BeginReplay from anywhere but internal/record/migrate. While a replay is open, the
// run's handle is its transaction's: a read through the handle — anything that takes the Run and
// opens the record — waits on the replay forever. Reads go through Querier.
//
//pluginparity:migrate-only
type Replay struct {
	run  Run
	tx   *recordsql.Tx
	err  error // the failure that ended the replay; every later call returns it
	done bool  // committed or rolled back
}

// ErrReplayFailed is wrapped by every error a Replay returns once the replay itself has failed — a
// savepoint statement that did not run, or a transaction SQLite ended on its own — as distinct from
// a refusal of one act, after which the replay takes the next.
var ErrReplayFailed = errors.New("the replay's transaction failed")

// SavepointTestHook is a TEST HOOK: when set, it runs before each savepoint statement a Replay
// issues and an error it returns stands in for SQLite's. The failure it stands in for — the disk
// full, an I/O error, the transaction rolled back under the replay — cannot be produced on demand by
// any act a replay is handed. Nil outside a test.
var SavepointTestHook func(stmt string) error

// actSavepoint names the one savepoint a replay holds at a time: acts are written one after
// another, never nested.
const actSavepoint = "act"

// BeginReplay opens a replay on run's record. The record is opened (and, on a fresh one, its schema
// applied in a transaction of its own) before the replay's transaction begins.
//
//pluginparity:migrate-only
func BeginReplay(run Run) (*Replay, error) {
	if !run.Valid() {
		return nil, feov.Errorf(feov.MissingField, "this replay names no run")
	}
	db, err := openRun(run)
	if err != nil {
		return nil, err
	}
	WriteBegins()
	tx, err := recordsql.BeginWrite(db)
	if err != nil {
		return nil, err
	}
	return &Replay{run: run, tx: tx}, nil
}

// Write writes one act inside the replay, stamped at: the shape check every writer runs, the key and
// the insert. A refusal of the act rolls its savepoint back and is returned; the replay takes the
// next act. An error wrapping ErrReplayFailed is the replay's own failure, and every later call
// returns it.
//
// id.Run must be the replay's run: a write addressed to another run is refused, never written here.
func (b *Replay) Write(id Identity, body proto.Message, at time.Time) (*Event, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.done {
		return nil, fmt.Errorf("this replay is already committed or rolled back")
	}
	if id.Run != b.run {
		return nil, feov.Errorf(feov.Internal,
			"this act is addressed to the run at %s, and the replay writes the run at %s — an act lands on the run it names or nowhere",
			id.Run.Dir(), b.run.Dir())
	}
	if err := b.savepoint("SAVEPOINT " + actSavepoint); err != nil {
		return nil, err
	}
	ts := at.UTC().Format(stampLayout)
	ev, err := appendAct(id, body, func() (*recordsql.Tx, error) { return b.tx, nil }, nil,
		writeMode{stamp: func() string { return ts }}, nil)
	if err != nil {
		if ferr := b.savepoint("ROLLBACK TO " + actSavepoint); ferr != nil {
			return ev, ferr
		}
		// The rolled-back act's ids are free again, so a stream read before the rollback is not a
		// prefix of the next read.
		b.tx.ForgetReads()
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

// StampedBody is one body of a reading's act a replay writes whole (WriteBodies), with the time its
// source event carried: each body is stamped with its own.
//
//pluginparity:migrate-only
type StampedBody struct {
	Body proto.Message
	At   time.Time
}

// WrittenBody is one event of a reading's act a replay wrote whole (WriteBodies), and its events.id.
//
//pluginparity:migrate-only
type WrittenBody struct {
	Event *Event
	ID    int64
}

// WriteBodies writes a reading's act whole inside the replay, as AppendReading writes it live: its
// gap's Mint, the Anchor that places it and the reading, in that order, in one savepoint, each body
// held to the shape check with the act in hand (readingAct) — so the reading's carrier_minted reads
// the act, not a record the act is not on yet — and each stamped with its own At. A refusal of any
// body rolls the savepoint back whole and is returned; the replay takes the next act. It returns each
// event written and its events.id, read inside the savepoint after that body's insert, so a caller
// maps every source event to the event written from it and never reads the act's last id for its
// first.
func (b *Replay) WriteBodies(id Identity, bodies []StampedBody) ([]WrittenBody, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.done {
		return nil, fmt.Errorf("this replay is already committed or rolled back")
	}
	if id.Run != b.run {
		return nil, feov.Errorf(feov.Internal,
			"this act is addressed to the run at %s, and the replay writes the run at %s — an act lands on the run it names or nowhere",
			id.Run.Dir(), b.run.Dir())
	}
	act, err := readingActOf(bodies)
	if err != nil {
		return nil, err
	}
	if err := b.savepoint("SAVEPOINT " + actSavepoint); err != nil {
		return nil, err
	}
	var out []WrittenBody
	for _, sb := range bodies {
		ts := sb.At.UTC().Format(stampLayout)
		ev, err := appendAct(id, sb.Body, func() (*recordsql.Tx, error) { return b.tx, nil }, nil,
			writeMode{stamp: func() string { return ts }}, act)
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
		b.tx.ForgetReads()
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
		"an act a replay writes whole is a reading's: its gap's Mint, the Anchor whose id is that gap, and the Verify its act minted the gap for, in that order")
}

// savepoint runs one savepoint statement. A failure ends the replay: past it, the savepoints no
// longer bracket the acts, and SQLite may have ended the transaction itself — a later pair run
// outside BEGIN would commit each act on its own.
func (b *Replay) savepoint(stmt string) error {
	var err error
	if SavepointTestHook != nil {
		err = SavepointTestHook(stmt)
	}
	if err == nil {
		_, err = b.tx.Exec(stmt)
	}
	if err != nil {
		b.err = fmt.Errorf("%w: %s: %v", ErrReplayFailed, stmt, err)
		return b.err
	}
	return nil
}

// Querier is the replay's transaction, for the reads its caller makes between acts: they see the acts
// the replay has written, and they are the only reads that do not wait on it.
func (b *Replay) Querier() recordsql.Querier { return b.tx }

// Err is the replay's failure, if it has one: nil while every act's savepoint has held. A caller
// that records refused acts and carries on asks it at the end, so a failure that landed on any one
// act — whichever site wrote it — ends the replay rather than reading as a refusal.
func (b *Replay) Err() error { return b.err }

// Commit commits every act the replay wrote. A failed replay does not commit: its error is returned
// and the transaction rolled back.
func (b *Replay) Commit() error {
	if b.err != nil {
		_ = b.Rollback()
		return b.err
	}
	if b.done {
		return fmt.Errorf("this replay is already committed or rolled back")
	}
	b.done = true
	return b.tx.Commit()
}

// Rollback discards every act the replay wrote. It may be called any number of times, and after
// Commit it does nothing.
func (b *Replay) Rollback() error {
	if b.done {
		return nil
	}
	b.done = true
	// A failed replay's transaction may already be gone — SQLite ends it on a full disk or an I/O
	// error — so its rollback answers nothing the failure has not.
	if err := b.tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) && b.err == nil {
		return err
	}
	return nil
}
