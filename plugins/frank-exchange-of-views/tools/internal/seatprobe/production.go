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
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
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
func backendFor(b Board) debatejs.Backend {
	ids := make([]any, 0, len(b.Gaps))
	manifest := make([]any, 0, len(b.Gaps))
	for i := range b.Gaps {
		id := fmt.Sprintf("G%d", i+1)
		ids = append(ids, id)
		manifest = append(manifest, id)
	}
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
		engaged = append(engaged, map[string]any{"seat_id": "judge", "gap_ids": ids})
		docket = ids
	}
	plan := func(parties []any, pass bool, dk []any) map[string]any {
		return map[string]any{"head": 2, "parties": parties, "docket": dk, "pass_permitted": pass, "ceiling": false, "why": []any{"seatprobe capture"}}
	}
	return func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "seatprobe capture", "petitions": []any{}, "log": []any{}, "rulings": []any{},
			"resolutions": []any{}, "holdings": []any{},
			"manifest": manifest, "claim_count": len(b.Claims),
			"saturation_reached": false, "sitting_record_appended": true,
			"open_gaps": 0, "unruled_motions": 0,
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

// ProductionPrompt is the prompt debate.js hands b.Seat, and the routing it hands it under.
//
// THE MISS IS AN ERROR. A board naming a seat debate.js never dispatches gets no prompt and no
// fallback: dispatching a hand-written substitute is how the probe spent four runs measuring a
// seat production does not seat.
func ProductionPrompt(scriptPath string, b Board, runDir, binDir, model, judgmentModel string) (debatejs.Dispatch, error) {
	ds, err := debatejs.Capture(scriptPath, debatejs.Config{
		Topic: b.Name, RunDir: runDir, BinDir: binDir, Lanes: 1,
		Model: model, JudgmentModel: judgmentModel, Backend: backendFor(b),
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
