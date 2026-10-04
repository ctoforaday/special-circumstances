package record

import (
	"encoding/json"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
type narrowedView[T any] struct {
	families []recordpb.EventType
	render   func(run Run, evs []*Event, win WindowIndex) (T, error)
}

// narrowedRender is a declared view with its result type erased — what the parity test walks.
type narrowedRender struct {
	families []recordpb.EventType
	render   func(run Run, evs []*Event, win WindowIndex) (any, error)
}

// narrowedViews is every declared narrowed view, by the name it renders under.
var narrowedViews = map[string]narrowedRender{}

// declareNarrowedView declares a view that loads families and renders them, and registers it.
func declareNarrowedView[T any](name string, render func(Run, []*Event, WindowIndex) (T, error), families ...recordpb.EventType) narrowedView[T] {
	narrowedViews[name] = narrowedRender{families: families, render: func(run Run, evs []*Event, win WindowIndex) (any, error) {
		return render(run, evs, win)
	}}
	return narrowedView[T]{families: families, render: render}
}

// of renders the view from the families it declared.
func (v narrowedView[T]) of(run Run) (T, error) {
	evs, win, err := EventsOf(run, v.families...)
	if err != nil {
		var zero T
		return zero, err
	}
	return v.render(run, evs, win)
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

// rendersEvents adapts a renderer that needs only the events to the declaration's signature.
func rendersEvents[T any](f func([]*Event, WindowIndex) T) func(Run, []*Event, WindowIndex) (T, error) {
	return func(_ Run, evs []*Event, win WindowIndex) (T, error) { return f(evs, win), nil }
}
