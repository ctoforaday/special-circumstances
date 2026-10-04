package record

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A NARROWED VIEW LOADS WHAT IT RENDERS, and says so in one place.
//
// A projection that renders a few kinds of act reads only those families from the record rather
// than hauling the whole of it through the loader. The families it loads and the renderer that reads
// them are one declaration, because a family the renderer reads and the load leaves out is not an
// error: it renders as nothing, in the bytes the view uses for "none on the record". The debate
// printed every epoch's verdict empty that way — the gates were never loaded (#1243) — and the
// evidence view listed every contradiction as unanswered and no anchor as reopened, because neither
// the findings that answer one nor the edits that reopen one were.
//
// Every declaration is registered in narrowedViews, and narrowedview_test renders each one from its
// narrowed read and from the whole record and holds the two to the same bytes — so a family a
// renderer reads and its declaration leaves out fails there, on a record that carries it.
//
// ONE READ TRANSACTION, for the families and for every other question the renderer asks. The board
// reads its closures from the events and its openness from the `gap` table; asked on two snapshots, a
// close landing between them renders a gap closed with no closure and epoch 0. So the renderer is
// handed the transaction the events came off and asks the record nothing except through it. The run's
// handle is ONE connection (recordsql.Open), so a read around that transaction, from inside the
// renderer, waits on it forever.
type narrowedView[T any] struct {
	families []recordpb.EventType
	render   narrowedRenderer[T]
}

// narrowedRenderer renders a view from the families its declaration loaded, asking any further
// question of q — the read transaction they came off, nil on a run with no record yet.
type narrowedRenderer[T any] func(run Run, q recordsql.Querier, evs []*Event, win WindowIndex) (T, error)

// declaredView is a declared view with its result type erased — what the parity test walks.
type declaredView interface {
	declared() []recordpb.EventType
	ofAny(run Run) (any, error)
	renderAny(run Run, q recordsql.Querier, evs []*Event, win WindowIndex) (any, error)
}

// narrowedViews is every declared narrowed view, by the name it renders under.
var narrowedViews = map[string]declaredView{}

// declareNarrowedView declares a view that loads families and renders them, and registers it. A
// second declaration under one name would leave the parity test walking only the last, so it panics.
func declareNarrowedView[T any](name string, render narrowedRenderer[T], families ...recordpb.EventType) narrowedView[T] {
	if _, dup := narrowedViews[name]; dup {
		panic(fmt.Sprintf("record: the narrowed view %q is declared twice", name))
	}
	v := narrowedView[T]{families: families, render: render}
	narrowedViews[name] = v
	return v
}

func (v narrowedView[T]) declared() []recordpb.EventType { return v.families }

func (v narrowedView[T]) ofAny(run Run) (any, error) { return v.of(run) }

func (v narrowedView[T]) renderAny(run Run, q recordsql.Querier, evs []*Event, win WindowIndex) (any, error) {
	return v.render(run, q, evs, win)
}

// of renders the view from the families it declared, on one snapshot of the record.
func (v narrowedView[T]) of(run Run) (T, error) {
	var out T
	err := readSnapshot(run, func(q recordsql.Querier) error {
		evs, win, err := eventsOfAt(q, v.families...)
		if err != nil {
			return err
		}
		out, err = v.render(run, q, evs, win)
		return err
	})
	return out, err
}

// jsonBytes renders the view as indented JSON (a seat reads it in a terminal transcript).
func (v narrowedView[T]) jsonBytes(run Run) ([]byte, error) {
	view, err := v.of(run)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(view, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// readSnapshot runs read inside one read transaction on the run's record — every answer off one
// snapshot — or with a nil Querier on a run that has recorded nothing. ReadOnly is what makes the
// BEGIN deferred: the handle's `_txlock=immediate` would otherwise take the write lock for a read.
func readSnapshot(run Run, read func(q recordsql.Querier) error) error {
	db, err := openRunForRead(run)
	if err != nil {
		return err
	}
	if db == nil {
		return read(nil)
	}
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("record: opening a read of the record: %w", err)
	}
	defer tx.Rollback()
	return read(tx)
}

// rendersEvents adapts a renderer that needs only the events to the declaration's signature.
func rendersEvents[T any](f func([]*Event, WindowIndex) T) narrowedRenderer[T] {
	return func(_ Run, _ recordsql.Querier, evs []*Event, win WindowIndex) (T, error) { return f(evs, win), nil }
}
