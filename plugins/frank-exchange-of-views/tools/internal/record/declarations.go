package record

import (
	"regexp"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S HOLDINGS REACH EVERY SEAT AS DECLARATIONS, ON THE WORK LIST, WITH THE BOARD BESIDE
// EACH GAP ONE NAMES (the owner's ruling of 2026-10-17).
//
// A holding binds every seat, and the record is the only place one is held: the bench records it
// with `declare`, and this list carries every declaration on the record, in record order, to every
// seat that reads its work list — the list the SubagentStart hook hands a seat as it is dispatched,
// so a seat holds the run's holdings before its first act.
//
// EVERY DECLARATION, NOT A "CURRENT" SUBSET. The record holds no link from a declaration to one it
// replaces — a declaration is free prose and names nothing structurally — so no reader can say which
// one a later one supersedes, and choosing would be a guess presented as a fold. A later declaration
// that replaces an earlier one says so in its own text, and the seat reads both.
//
// THE GAP'S STATUS IS THE BOARD'S, PLACED BESIDE THE PROSE BY THE TOOL. A holding states a gap's
// status in prose when the bench writes it, and the prose does not move when the gap does: on run
// m25 "G-a0cfb606 open … ruled before any PASS" stayed in front of every seat after the gap closed,
// and a chair acting on a declaration's "G-7199a24f remains open" after that gap closed ended a run
// at CEILING. Nothing here reads a status out of the prose. The gap IDS a declaration names are read
// from its text — the id shape is minted by the tool and recovered by anchor.IDPattern, the one id
// matcher — and each id's status comes off the same gap-view rows the work list's `open` and
// `estopped` are built from, so the list cannot state a gap one way here and another way there.
//
// A NAMED ID THE BOARD DOES NOT HOLD IS SAID, never dropped: a declaration naming a gap that does not
// exist is a fact a seat needs as much as a stale status, and silence would read as "names no gap".

// DeclarationJSON is one bench declaration as the work list carries it.
type DeclarationJSON struct {
	// Holding is the declaration's text exactly as the bench recorded it.
	Holding string `json:"holding"`
	// Epoch is the epoch the bench declared it in, so a seat can set the prose against when the
	// gaps it names last moved.
	Epoch int `json:"epoch"`
	// Gaps is each gap id the holding's text names, in order of first mention, with where the board
	// says that gap stands as the list is read. NOT omitempty: a holding that names no gap renders
	// `[]`, which is "names none", where an absent key would be "not computed".
	Gaps []DeclaredGapJSON `json:"gaps"`
}

// DeclaredGapJSON is one gap a declaration names, and the board's answer for it now. Its words are
// the near-match screen's: `status` open | closed, and on a closed gap `closed_by` bench | red with
// the `fate` that closed it — and `not_on_board` where the id names no gap the board holds.
type DeclaredGapJSON struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Fate, ClosedBy and ClosedEpoch are a closed gap's, and absent on an open one because there is
	// nothing to say — the same reading near-match's `closed_by` has.
	Fate        string `json:"fate,omitempty"`
	ClosedBy    string `json:"closed_by,omitempty"`
	ClosedEpoch int    `json:"closed_epoch,omitempty"`
}

// The words of DeclaredGapJSON.Status.
const (
	declaredGapOpen       = "open"
	declaredGapClosed     = "closed"
	declaredGapNotOnBoard = "not_on_board"
)

// gapIDInText is a gap id standing as a word in prose.
var gapIDInText = regexp.MustCompile(`\b` + anchor.IDPattern("gap") + `\b`)

// declarationsOf is every declaration in evs, in record order, each gap it names stated as gaps —
// the work list's own gap states — hold it.
func declarationsOf(evs []*Event, win WindowIndex, gaps []WorkGapState) []DeclarationJSON {
	byID := make(map[string]WorkGapState, len(gaps))
	for _, g := range gaps {
		byID[g.ID] = g
	}
	out := []DeclarationJSON{}
	for _, e := range evs {
		d, ok := recordpb.BodyAs[*recordpb.Declare](e)
		if !ok {
			continue
		}
		dj := DeclarationJSON{Holding: d.GetHolding(), Epoch: win.Of(e).Epoch, Gaps: []DeclaredGapJSON{}}
		seen := map[string]bool{}
		for _, id := range gapIDInText.FindAllString(dj.Holding, -1) {
			if seen[id] {
				continue
			}
			seen[id] = true
			dj.Gaps = append(dj.Gaps, declaredGapOf(id, byID))
		}
		out = append(out, dj)
	}
	return out
}

// declaredGapOf is the board's answer for one named id.
func declaredGapOf(id string, byID map[string]WorkGapState) DeclaredGapJSON {
	g, ok := byID[id]
	switch {
	case !ok:
		return DeclaredGapJSON{ID: id, Status: declaredGapNotOnBoard}
	case g.Open:
		return DeclaredGapJSON{ID: id, Status: declaredGapOpen}
	}
	by := "red"
	if g.ClosedByBench {
		by = "bench"
	}
	return DeclaredGapJSON{ID: id, Status: declaredGapClosed, Fate: g.Fate, ClosedBy: by, ClosedEpoch: g.closedEpoch}
}
