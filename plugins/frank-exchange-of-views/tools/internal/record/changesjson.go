package record

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

// ChangesJSONOfDB projects the recorded edits from the `change` view.
//
// A QUERY, NOT A FOLD. This read used to load every event in the record through MergedEvents and walk
// them looking for one body type, which made the cost of asking about the edits scale with the size of
// the whole record — and pulled in the per-row clock derivation on the way. The view states the same
// three rules where every reader can see them: the clock comes from events_w, the act that STANDS
// after a correction comes from live_event, and the character delta is computed in SQL.
//
// ORDERED BY "pos", NOT BY EVENT ID, because that is what Live does: a corrected edit is replaced by
// its successor at the STRUCK act's position, so a repair does not reorder the log. LiveKeys pins the
// two orders against each other.
func ChangesJSONOfDB(db *sql.DB) (ChangesJSON, error) {
	out := ChangesJSON{Edits: []EditJSON{}}
	rows, err := db.Query(`SELECT "seat_id", "sitting", "epoch", "answers", "old", "new",
	    "delta", "reason", "applied_verbatim", "accepted"
	  FROM "change" ORDER BY "pos", "event_id"`)
	if err != nil {
		return out, fmt.Errorf("record: asking the change view: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ed EditJSON
		// THE SCAN ORDER IS THE SELECT'S ORDER. Two same-typed neighbours swapped here compile
		// silently and put one edit's text under another's heading, so the two lists are kept
		// literally adjacent above and below.
		if err := rows.Scan(&ed.Seat, &ed.Sitting, &ed.Epoch, &ed.Answers, &ed.Old, &ed.New,
			&ed.Delta, &ed.Reason, &ed.AppliedVerbatim, &ed.Accepted); err != nil {
			return out, err
		}
		out.Edits = append(out.Edits, ed)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	out.Counts.Edits = len(out.Edits)
	return out, nil
}

// ChangesJSONBytes renders it as indented JSON, mirroring the other structured reads.
func ChangesJSONBytes(run Run) ([]byte, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	// A RUN WITH NO RECORD YET IS AN EMPTY LOG, not an error: `show changes` before blue's first
	// edit is a legitimate read, and the empty `edits` list is the honest answer to it.
	out := ChangesJSON{Edits: []EditJSON{}}
	if db != nil {
		if out, err = ChangesJSONOfDB(db); err != nil {
			return nil, err
		}
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
