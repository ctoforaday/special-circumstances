package debatejs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// planWith is a chair's relayed plan: the parties and blockers given, every other field as
// `dispatch next` prints it on a board with nothing else to say — the parties' own blockers where
// none is given, and a PASS permitted where nobody is ready and nothing holds it.
func planWith(parties, blockers []any) map[string]any {
	if len(blockers) == 0 {
		blockers = HeldBy(parties)
	}
	return WithKey(map[string]any{"head": 2, "parties": parties, "docket": []any{}, "remand_owed": []any{}, "exchanges": []any{}, "pass_permitted": len(blockers) == 0, "ceiling": false,
		"max_epochs": 0, "epoch": 1, "final_sitting": false, "epoch_limit_reached": false, "why": []any{"test"}, "stale_areas": []any{}, "blockers": blockers})
}

var (
	lensParty     = map[string]any{"seat_id": "red-lens-logic", "gap_ids": []any{}}
	benchPetition = map[string]any{"kind": "unruled_motion", "subject": "M1", "owner": "judge"}
	chairGrade    = map[string]any{"kind": "unruled_motion", "subject": "M2", "owner": "red-chair"}
)

// drivePlans runs debate.js with the chair relaying plans[i] at its (i+1)th sitting and an empty
// plan after; every other seat returns an envelope carrying `extra`.
func drivePlans(t *testing.T, plans []map[string]any, extra Envelope) ([]Dispatch, Outcome) {
	t.Helper()
	chair := 0
	backend := func(seatID, label, prompt string) Envelope {
		e := Envelope{"rulings": []any{}, "dispositions": []any{},
			"manifest": []any{}, "open_gaps": 0}
		for k, v := range extra {
			e[k] = v
		}
		if seatID == "red-chair" {
			chair++
			if chair <= len(plans) {
				e.Relay(plans[chair-1])
			} else {
				e.Relay(planWith([]any{}, []any{}))
			}
		}
		return e
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	ds, out, err := CaptureRun(ScriptPath(filepath.Join(wd, "..", "..")), Config{Topic: "t", RunDir: t.TempDir(), BinDir: t.TempDir(),
		Lanes: 1, Model: "haiku", JudgmentModel: "haiku", Backend: backend})
	if err != nil {
		t.Fatalf("driving debate.js: %v", err)
	}
	return ds, out
}

func occasions(ds []Dispatch) []string {
	var out []string
	for _, d := range ds {
		if d.Occasion != "" {
			out = append(out, d.Occasion)
		}
	}
	return out
}

// owing is a plan whose chair's close reports these hearings owed, as the hook writes it.
func owing(plan map[string]any, occasions ...any) map[string]any {
	plan["record_read"] = map[string]any{"event": 1, "bench_owed": occasions, "halted": false}
	return plan
}

// THE ENGINE HEARS WHAT A CLOSE SAYS IS OWED, WITH NOTHING RUNNING (gblock, 2026-10-31). A lens files on
// the record; the read on the next close reports the hearing owed; the workflow seats that sitting
// before any party sits after that close, with no plan naming it.
func TestTheEngineHearsWhatACloseSaysIsOwed(t *testing.T) {
	ds, _ := drivePlans(t, []map[string]any{owing(planWith([]any{lensParty}, []any{benchPetition}), "petition")}, nil)
	petition := slices.IndexFunc(ds, func(d Dispatch) bool { return d.Occasion == "petition" })
	if petition < 0 {
		t.Fatalf("no petition sitting dispatched; occasions %q", occasions(ds))
	}
	if d := ds[petition]; d.SeatID != "judge" || !strings.HasPrefix(d.Label, "judge ·") {
		t.Errorf("petition sitting dispatched as %q (label %q), want the bench", d.SeatID, d.Label)
	}
	lens := slices.IndexFunc(ds, func(d Dispatch) bool { return d.SeatID == "red-lens-logic" })
	if lens < 0 || lens < petition {
		t.Errorf("the lens sat at %d and the petition sitting at %d — the chair's filing is heard before any party sits", lens, petition)
	}
	if n := strings.Count(strings.Join(occasions(ds), " "), "docket"); n != 0 {
		t.Errorf("%d docket sitting(s) for a bench that owes only a hearing", n)
	}
}

// THE TERMINAL BENCH SITTING FIRES IFF THE LAST PLAN HOLDS A BLOCKER THE BENCH OWNS — not on a
// count a seat reports, and not for a motion whose gavel is the chair's, which the bench has no
// verb to rule.
func TestTheTerminalBenchFiresOnABenchOwnedBlocker(t *testing.T) {
	for _, c := range []struct {
		name     string
		blockers []any
		want     bool
	}{
		{"a bench-owned motion stands", []any{benchPetition}, true},
		{"only the chair's motion stands", []any{chairGrade}, false},
		{"nothing stands", []any{}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			// The chair reports an unruled count no envelope schema carries; it must not matter.
			ds, _ := drivePlans(t, []map[string]any{planWith([]any{}, c.blockers)}, Envelope{"unruled_motions": 3})
			got := slices.Contains(occasions(ds), "terminal")
			if got != c.want {
				t.Errorf("terminal sitting dispatched = %v, want %v (occasions %q)", got, c.want, occasions(ds))
			}
		})
	}
}

// AN ENVELOPE IS NOT A PETITION CHANNEL. Every seat's envelope carries a `petitions` field here and
// no read reports one owed: no petition sitting is dispatched. The negative control
// against a second reader of the fact the record carries.
func TestAnEnvelopePetitionIsNotAChannel(t *testing.T) {
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}}}, []any{})},
		Envelope{"petitions": []any{map[string]any{"class": "safety", "basis": "b", "relief": "r"}}})
	if slices.Contains(occasions(ds), "petition") {
		t.Errorf("a petition sitting was dispatched off an envelope field: occasions %q", occasions(ds))
	}
}

// A BENCH PARTY WITH NO GAP IS REFUSED, NOT GUESSED AT. The chair convenes the bench for its docket and
// nothing else; a relay that emptied the list is altered in transit, and an empty gap list does not
// mean "petition".
func TestABenchPartyWithNoGapIsRefused(t *testing.T) {
	wd, _ := os.Getwd()
	_, _, err := CaptureRun(ScriptPath(filepath.Join(wd, "..", "..")), Config{Topic: "t", RunDir: t.TempDir(), BinDir: t.TempDir(),
		Lanes: 1, Model: "haiku", JudgmentModel: "haiku", Backend: func(seatID, label, prompt string) Envelope {
			e := Envelope{"rulings": []any{}, "dispositions": []any{}, "open_gaps": 0}
			if seatID == "red-chair" {
				e.Relay(planWith([]any{map[string]any{"seat_id": "judge", "gap_ids": []any{}}}, []any{benchPetition}))
			}
			return e
		}})
	if err == nil || !strings.Contains(err.Error(), "a bench party with no gap") {
		t.Errorf("a bench party with no gap drove to %v, want the relay refused naming it", err)
	}
}
