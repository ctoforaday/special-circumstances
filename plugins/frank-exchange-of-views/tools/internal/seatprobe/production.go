package seatprobe

// THE PROMPT IS NO LONGER WRITTEN HERE.
//
// Until now the probe handed its seat a prompt composed in the harness — a paraphrase of what
// debate.js says, roughly 950 characters against production's 12,800. Every help-reading number
// the probe ever reported was therefore a number about a prompt no seat is dispatched with. That is
// not a weak measurement; it is a different experiment reported under this one's name.
//
// What follows executes debate.js — the shipped file, the one a real run loads — and takes the
// prompt it renders for the board's seat. A clause edited in debate.js reaches the probe on the
// next run, because there is no second copy to keep in step.
//
// # The board is the state, and it is the state at BOTH ends
//
// debate.js embeds run state into its prompts: blue's open-gap JSON, the judge's contested docket.
// That state comes from the envelopes seats return, which under capture are stubs — so the stubs
// are built FROM THE BOARD that stages the record. A prompt naming gaps the record does not carry
// would put the seat in a situation the board contradicts, and the seat would be right to be
// confused and wrong to be scored for it.

import (
	"fmt"
	"os"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// backendFor answers debate.js's dispatches with envelopes derived from b.
//
// It is not a random walk like the fuzz's: the capture is a rendering, so the envelopes are the
// board, stated in the shape debate.js branches on. Two branches matter and are deliberate:
//
//	VERDICT FAIL   a PASS ends the run before red's chair dispatches anyone else.
//	GAPS REPEAT    red returns the SAME gap ids every epoch, so the second epoch sees them as
//	               re-raised. That is what fills the contested docket — and the docket is the
//	               ONLY thing that seats a judge at all. The first epoch cannot have one (nothing
//	               persists yet, and no dispute is pending), which is why no board may name `judge`.
func backendFor(b Board, ids []any) debatejs.Backend {
	// THE CHAIR RELAYS A PLAN (plans/roundless.md §III.B.1), and the plan is what dispatches the
	// seat under probe: its first sitting engages the board's seat on the board's gaps — a lens,
	// blue, or the bench (docketed) — and its second sitting permits PASS and records it, which
	// ends the run. A chair board is captured at the first sitting either way.
	engaged := []any{}
	docket := []any{}
	switch {
	case strings.HasPrefix(b.Seat, "red-lens-"), b.Seat == "blue-respond":
		engaged = append(engaged, map[string]any{"seat_id": b.Seat, "gap_ids": ids})
	case b.Seat == "judge":
		engaged = append(engaged, map[string]any{"seat_id": "judge", "gap_ids": ids, "occasions": []any{"docket"}})
		docket = ids
	}
	plan := func(parties []any, pass bool, dk []any) map[string]any {
		return map[string]any{"head": 2, "parties": parties, "docket": dk, "remand_owed": []any{}, "pass_permitted": pass, "ceiling": false,
			"max_epochs": 0, "epoch_limit_reached": false, "why": []any{"seatprobe capture"}, "stale_areas": []any{}, "blockers": []any{}}
	}
	return func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "seatprobe capture", "rulings": []any{},
			"dispositions": []any{}, "holdings": []any{},
			"manifest":           ids,
			"saturation_reached": false,
			"open_gaps":          0,
		}
		switch {
		case strings.HasPrefix(seatID, "red-chair"):
			if strings.Contains(label, "#1 ") || strings.HasSuffix(label, "#1") {
				e["plan"] = plan(engaged, false, docket)
				if len(engaged) == 0 {
					e["plan"] = plan([]any{}, true, []any{})
					e["verdict"] = "PASS"
				}
			} else {
				e["plan"] = plan([]any{}, true, []any{})
				e["verdict"] = "PASS"
			}
		}
		return e
	}
}

// gapIDs are the ids the record at runDir minted for the board's gaps, in mint order — the ids the
// seat will find on its board. A board not staged yet has no record, and its capture names each gap
// by position in the id shape, so the prompt renders at the size it will have.
func gapIDs(b Board, runDir string) ([]any, error) {
	ids := make([]any, len(b.Gaps))
	if _, err := os.Stat(runDir); err != nil {
		for i := range ids {
			ids[i] = anchor.ID("gap", [4]byte{3: byte(i + 1)})
		}
		return ids, nil
	}
	run, err := record.OpenRun(runDir)
	if err != nil {
		return nil, err
	}
	fam, err := record.FamilyOf(run)
	if err != nil || len(fam.Gaps) != len(b.Gaps) {
		return nil, fmt.Errorf("the record at %s holds %d gap(s) for a board of %d (%v)", runDir, len(fam.Gaps), len(b.Gaps), err)
	}
	for i, g := range fam.Gaps {
		ids[i] = g.ID
	}
	return ids, nil
}

// ProductionPrompt is the prompt debate.js hands b.Seat, and the routing it hands it under.
//
// THE MISS IS AN ERROR. A board naming a seat debate.js never dispatches gets no prompt and no
// fallback: dispatching a hand-written substitute is how the probe spent four runs measuring a
// seat production does not seat.
func ProductionPrompt(scriptPath string, b Board, runDir, binDir, model, judgmentModel string) (debatejs.Dispatch, error) {
	ids, err := gapIDs(b, runDir)
	if err != nil {
		return debatejs.Dispatch{}, fmt.Errorf("board %q: %w", b.Name, err)
	}
	ds, err := debatejs.Capture(scriptPath, debatejs.Config{
		Topic: b.Name, RunDir: runDir, BinDir: binDir, Lanes: 1,
		Model: model, JudgmentModel: judgmentModel, Backend: backendFor(b, ids),
	})
	if err != nil {
		return debatejs.Dispatch{}, fmt.Errorf("board %q: %w", b.Name, err)
	}
	d, err := debatejs.For(ds, b.Seat)
	if err != nil {
		return debatejs.Dispatch{}, fmt.Errorf("board %q: %w", b.Name, err)
	}
	if strings.TrimSpace(d.Prompt) == "" {
		return debatejs.Dispatch{}, fmt.Errorf("board %q: debate.js dispatched %s with an EMPTY prompt", b.Name, b.Seat)
	}
	if d.AgentType == "" {
		return debatejs.Dispatch{}, fmt.Errorf("board %q: debate.js dispatched %s with no agentType — the probe would fall back to the default agent and score a seat sitting under no constitution", b.Name, b.Seat)
	}
	return d, nil
}
