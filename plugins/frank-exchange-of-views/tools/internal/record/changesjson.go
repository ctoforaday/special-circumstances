package record

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
	// Gap is red's side of the comparison on the form scoped to one gap (--id): what the gap's mint
	// prescribed, beside the edits answering it. null on the unscoped form, which is about no one gap —
	// NOT omitempty, so the key a seat read on the page is in every document, and null says "unscoped".
	Gap *ChangesGapJSON `json:"gap"`
	// Edits is every recorded change in record order — each edit, and each citation retire, which takes a
	// citation off a sentence the report keeps — and on the scoped form every one answering the gap. NOT omitempty: an empty list is "blue recorded
	// no edit", which on a run past epoch 0 is itself a finding, and an absent key would read as a
	// projection that does not carry edits.
	Edits  []EditJSON `json:"edits"`
	Counts struct {
		Edits int `json:"edits"`
	} `json:"counts"`
}

// ChangesGapJSON is what a gap's mint asked of blue: the defect, the fix, the exact text where the lens
// prescribed one, and the check it re-audits against. Empty strings are the mint's own empties — a gap
// minted with no required_fix or no prescribed text — never a field this projection failed to read.
type ChangesGapJSON struct {
	ID              string `json:"id"`
	Problem         string `json:"problem"`
	RequiredFix     string `json:"required_fix"`
	AcceptanceCheck string `json:"acceptance_check"`
	// Location and FixNew are the lens's prescribed pair, and FixBasis how far it was verified against
	// the report when it was minted. FixNew is empty where the lens prescribed no exact text.
	Location string `json:"location"`
	FixNew   string `json:"fix_new"`
	FixBasis string `json:"fix_basis"`
}

// EditJSON is one recorded change, with the text on both sides: an edit, or a citation retire (Retired).
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
	// Retired is the citation a retire took off a sentence it kept, whose words are Old and New alike;
	// empty on an edit.
	Retired string `json:"retired"`
}

// ChangesOf is `show changes`'s ONE SELECTION, and both its renderings read it: the JSON marshals it
// and the markdown formats it. gapID "" is the unscoped form, every recorded edit; a gap id scopes it to
// that gap's mint and the edits answering it.
//
// ONE SELECTION, BECAUSE TWO DRIFTED. The markdown selected its own edits off the event fold and the
// JSON read every row of the `change` view, and only the markdown read --id: as the judge, `show changes
// --id <gap> --json` returned all 55 edits of the run and no required_fix while the bare form beside it
// was scoped. A rendering that formats this value cannot select differently from its sibling.
//
// A QUERY, NOT A FOLD. The edits come from the `change` view, which states the clock (events_w) and
// the character delta where every reader can see them, so asking about the edits does not scale with
// the size of the whole record. The gap and its edits are read on one snapshot.
func ChangesOf(run Run, gapID string) (ChangesJSON, error) {
	out := ChangesJSON{Edits: []EditJSON{}}
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if gapID != "" {
			var g ChangesGapJSON
			var problem, fix, check, loc, fixNew, basis sql.NullString
			found := false
			if !noRecord(q) {
				var err error
				if found, err = queryRowAt(q, []any{&problem, &fix, &check, &loc, &fixNew, &basis},
					`SELECT "problem", "required_fix", "acceptance_check", "location", "fix_new", "fix_basis"
					   FROM "mint" WHERE "gap_id" = ?`, gapID); err != nil {
					return fmt.Errorf("reading gap %s's mint: %w", gapID, err)
				}
			}
			if !found {
				return fmt.Errorf("no gap %s on the board — --id takes a gap id (a mint created it), and a view that answered for an id nobody minted would be inventing a comparison", gapID)
			}
			g.ID, g.Problem, g.RequiredFix, g.AcceptanceCheck = gapID, problem.String, fix.String, check.String
			g.Location, g.FixNew, g.FixBasis = loc.String, fixNew.String, basis.String
			out.Gap = &g
		}
		if noRecord(q) {
			return nil
		}
		// THE SCOPE IS ONE PREDICATE ON THE ROWS, "" selecting them all, so the scoped form is the
		// unscoped one narrowed and never a second query that could read the rows differently.
		rows, err := q.Query(`SELECT "seat_id", "sitting", "epoch", "answers", "old", "new",
		    "delta", "reason", "applied_verbatim", "accepted", "retired"
		  FROM "change" WHERE ? = '' OR "answers" = ? ORDER BY "event_id"`, gapID, gapID)
		if err != nil {
			return fmt.Errorf("asking the change view: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var ed EditJSON
			// THE SCAN ORDER IS THE SELECT'S ORDER. Two same-typed neighbours swapped here compile
			// silently and put one edit's text under another's heading, so the two lists are kept
			// literally adjacent above and below.
			if err := rows.Scan(&ed.Seat, &ed.Sitting, &ed.Epoch, &ed.Answers, &ed.Old, &ed.New,
				&ed.Delta, &ed.Reason, &ed.AppliedVerbatim, &ed.Accepted, &ed.Retired); err != nil {
				return err
			}
			out.Edits = append(out.Edits, ed)
		}
		return rows.Err()
	})
	if err != nil {
		return ChangesJSON{}, err
	}
	out.Counts.Edits = len(out.Edits)
	return out, nil
}

// ChangesJSONBytes renders ChangesOf as indented JSON, mirroring the other structured reads. A run
// with no record yet is an empty log, not an error: `show changes` before blue's first edit is a
// legitimate read, and the empty `edits` list is the honest answer to it.
func ChangesJSONBytes(run Run, gapID string) ([]byte, error) {
	out, err := ChangesOf(run, gapID)
	if err != nil {
		return nil, err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
