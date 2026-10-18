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
