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
//	VERDICT FAIL   a PASS ends the run before red's merge dispatches anyone else.
//	GAPS REPEAT    red returns the SAME gap ids every round, so round 2 sees them as re-raised.
//	               That is what fills the contested docket — and the docket is the ONLY thing
//	               that seats a judge at all. Round 1 cannot have one (nothing persists yet, and
//	               no dispute is pending), which is why no board may name `judge-r1`.
func backendFor(b Board) debatejs.Backend {
	gaps := make([]any, 0, len(b.Gaps))
	manifest := make([]any, 0, len(b.Gaps))
	for i, g := range b.Gaps {
		// THE ID IS THE RECORD'S. `mint` assigns R<round>-<n> in board order, which is what the
		// staged board carries — so the JSON debate.js threads into blue's prompt names the same
		// gaps the seat will find under `show board`. A separate id space here would put the seat
		// in front of a docket that does not exist.
		id := fmt.Sprintf("R1-%d", i+1)
		// REFS ONLY, BECAUSE THAT IS THE SCHEMA. RED_ENVELOPE's gap items declare exactly
		// id/severity/likelihood/impact/complexity_cost/supersedes and debate.js says why in its
		// own comment: the prose — location, problem, required_fix, acceptance_check — is written
		// to the BOARD by `mint` and read back from the record, never round-tripped through the
		// envelope. Enriching this map would render a prompt fuller than the contract guarantees,
		// which is the same class of flattery the whole fidelity pass removed.
		gaps = append(gaps, map[string]any{
			"id": id, "severity": g.Severity, "likelihood": g.Likelihood,
			"impact": g.Impact, "complexity_cost": g.Complexity, "supersedes": []any{},
		})
		manifest = append(manifest, id)
	}
	disputes := make([]any, 0)
	for _, m := range b.Motions {
		if m.Subject == "grade" && m.GapID != "" {
			disputes = append(disputes, map[string]any{
				"gap_id": m.GapID, "dimension": m.Dimension, "proposed": m.Proposed,
			})
		}
	}
	return func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{
			"synopsis": "seatprobe capture", "verdict": "FAIL", "citations_checked": 0,
			"gaps": []any{}, "petitions": []any{}, "friction": []any{}, "rulings": []any{},
			"closures": []any{}, "dispute_responses": []any{}, "deadlock": false,
			"resolutions": []any{}, "grade_disputes": []any{},
			"manifest": manifest, "claim_count": len(b.Claims),
			"saturation_reached": false, "round_record_appended": true,
			"open_gaps": []any{},
		}
		switch {
		case strings.HasPrefix(seatID, "red-merge"):
			e["gaps"] = gaps
		case strings.HasPrefix(seatID, "blue-respond"):
			e["grade_disputes"] = disputes
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
		Topic: b.Name, RunDir: runDir, BinDir: binDir, Lanes: 1, MaxRounds: 3,
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
