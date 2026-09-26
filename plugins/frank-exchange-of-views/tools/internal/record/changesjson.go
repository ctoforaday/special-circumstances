package record

import (
	"encoding/json"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE CHANGE LOG CARRIES THE CHANGE (gblock's ruling: "it's meant to have the diff").
//
// `show changes` rendered, per edit, a CHAR COUNT and blue's own prose description of what it did —
// `−106 chars` beside "Removed 'The research cost is…'". A reader auditing a repair was therefore
// taking the edited party's word for the edit, which is the one thing red's contract forbids, and the
// old/new text was in hand at the render (view/changes.go:83 reduced it to `delta()`).
//
// Measured on universe-m10: 23 of 26 `changes` reads were unscoped and got the char counts; 8 of them
// passed `--json`, which did not exist — the filters (`jq '.edits[]'`, `jq 'keys'`, `jq '.error'`)
// are a seat guessing the shape and then inspecting the refusal. Having got counts, the seat went to
// the report and grepped. Six calls to see a diff the record already held.
//
// The scoped form (`--id <gap>`) already rendered a proper diff. This is the same fact, unscoped and
// structured, so a seat can ask for all of it at once and filter it.
type ChangesJSON struct {
	// Edits is every recorded edit in record order. NOT omitempty: an empty list is "blue recorded
	// no edit", which on a run past epoch 0 is itself a finding, and an absent key would read as a
	// projection that does not carry edits.
	Edits  []EditJSON `json:"edits"`
	Counts struct {
		Edits int `json:"edits"`
	} `json:"counts"`
}

// EditJSON is one recorded edit, with the text on both sides.
type EditJSON struct {
	Seat    string `json:"seat"`
	Sitting int    `json:"sitting"`
	Epoch   int    `json:"epoch"`
	// Answers is the gap this edit responds to, empty for an edit answering none — which is a real
	// case ("blue's own") and not a missing value.
	Answers string `json:"answers"`
	// Old and New are the spans as blue typed them and as the tool matched them against the report.
	// They are the POINT of this projection: a reader comparing them is reading the change rather
	// than a description of it.
	Old string `json:"old"`
	New string `json:"new"`
	// Delta is New minus Old in characters, kept because it was the only thing the markdown gave and
	// a reader scanning for the big rewrites uses it.
	Delta int `json:"delta"`
	// Reason is blue's account of the edit. It sits BESIDE the diff rather than instead of it.
	Reason string `json:"reason"`
	// AppliedVerbatim and Accepted are how the edit came about — red's prescribed text taken as
	// given, and taken through the accept path. Estoppel is enforced on the first.
	AppliedVerbatim bool `json:"applied_verbatim"`
	Accepted        bool `json:"accepted"`
}

// ChangesJSONOf projects the recorded edits.
func ChangesJSONOf(evs []*Event) ChangesJSON {
	out := ChangesJSON{Edits: []EditJSON{}}
	var clk Clock
	for _, e := range Live(evs) {
		w := clk.Advance(e)
		ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		out.Edits = append(out.Edits, EditJSON{
			Seat: e.GetSeatId(), Sitting: w.Sitting, Epoch: w.Epoch,
			Answers: ed.GetAnswers(),
			Old:     ed.GetOld(), New: ed.GetNew(),
			Delta:           len([]rune(ed.GetNew())) - len([]rune(ed.GetOld())),
			Reason:          ed.GetText(),
			AppliedVerbatim: ed.GetAppliedVerbatim(), Accepted: ed.GetAccepted(),
		})
	}
	out.Counts.Edits = len(out.Edits)
	return out
}

// ChangesJSONBytes renders it as indented JSON, mirroring the other structured reads.
func ChangesJSONBytes(run Run) ([]byte, error) {
	m, err := MergedEvents(run)
	if err != nil {
		return nil, err
	}
	b, err := json.MarshalIndent(ChangesJSONOf(m.Events), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
