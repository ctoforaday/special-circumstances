package cli

import (
	"encoding/json"
	"testing"
)

// THE BOARD NAMES EACH GAP'S MINTER, as the work list does. m28's lenses asked `show board` for
// `minted_by` 21 times and jq returned null: the key was on the work list's rows alone, and the
// board's `found_by` holds finding ids, which name the lens that found a defect and not the seat
// that minted the gap. The value is the work list's, read off the same column.
func TestTheBoardNamesEachGapsMinter(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "board-minter", "read-surface")
	out, err := run(t, "show", "board", "--run", runDir, "--seat-id", "red-chair")
	if err != nil {
		t.Fatal(err)
	}
	var board struct {
		Open []struct {
			ID       string `json:"id"`
			MintedBy string `json:"minted_by"`
		} `json:"open"`
	}
	if err := json.Unmarshal([]byte(out), &board); err != nil {
		t.Fatalf("show board: %v\n%s", err, out)
	}
	work, err := run(t, "show", "work", "--run", runDir, "--seat-id", lensSeat)
	if err != nil {
		t.Fatal(err)
	}
	var list struct {
		Open []struct {
			ID       string `json:"id"`
			MintedBy string `json:"minted_by"`
		} `json:"open"`
	}
	if err := json.Unmarshal([]byte(work), &list); err != nil {
		t.Fatalf("show work: %v\n%s", err, work)
	}
	if len(board.Open) != 1 || board.Open[0].ID != id || board.Open[0].MintedBy != lensSeat {
		t.Fatalf("the board does not name %s's minter %s: %+v", id, lensSeat, board.Open)
	}
	if len(list.Open) != 1 || list.Open[0].MintedBy != board.Open[0].MintedBy {
		t.Errorf("the board and the work list name different minters: board %+v, work %+v", board.Open, list.Open)
	}
}

// THE BOARD STATES WHAT THE WORK LIST STATES ABOUT A GAP, off the same gap-view columns: universe-m29's
// chair and bench asked `show board` for `material`, `remanded` and `docket_direction` and jq returned
// null. Each key is on the board's row, and its value is the work list's.
func TestTheBoardStatesTheGapFactsTheWorkListStates(t *testing.T) {
	runDir := seatRun(t)
	id := mintGap(t, runDir, "board-material", "read-surface")
	row := func(args ...string) map[string]json.RawMessage {
		out, err := run(t, append([]string{"show"}, append(args, "--run", runDir)...)...)
		if err != nil {
			t.Fatal(err)
		}
		var v struct {
			Open []map[string]json.RawMessage `json:"open"`
		}
		if err := json.Unmarshal([]byte(out), &v); err != nil || len(v.Open) != 1 {
			t.Fatalf("show %v: %v\n%s", args, err, out)
		}
		return v.Open[0]
	}
	board, work := row("board", "--seat-id", "red-chair"), row("work", "--seat-id", lensSeat)
	if string(board["id"]) != `"`+id+`"` || string(board["material"]) != "true" {
		t.Fatalf("the board's row for %s, graded medium, is not material: %s", id, board["material"])
	}
	for _, k := range []string{"minted_by", "material", "remanded", "docket_direction"} {
		b, ok := board[k]
		if !ok {
			t.Errorf("the board's row has no %q", k)
			continue
		}
		w, listed := work[k]
		if !listed { // the work list omits an empty direction
			w = json.RawMessage(`""`)
		}
		if string(b) != string(w) {
			t.Errorf("%s: the board says %s and the work list %s", k, b, w)
		}
	}
}
