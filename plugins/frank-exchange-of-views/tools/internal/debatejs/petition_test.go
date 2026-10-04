package debatejs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// planWith is a chair's relayed plan: the parties and blockers given, every other field as
// `dispatch next` prints it on a board with nothing else to say.
func planWith(parties, blockers []any) map[string]any {
	return map[string]any{"head": 2, "parties": parties, "docket": []any{}, "remand_owed": []any{}, "pass_permitted": false, "ceiling": false,
		"max_epochs": 0, "epoch_limit_reached": false, "why": []any{"test"}, "stale_areas": []any{}, "blockers": blockers}
}

var (
	lensParty     = map[string]any{"seat_id": "red-lens-logic", "gap_ids": []any{}}
	petitionBench = map[string]any{"seat_id": "judge", "gap_ids": []any{}, "occasions": []any{"petition"}}
	benchPetition = map[string]any{"kind": "unruled_motion", "subject": "M1", "owner": "judge"}
	chairGrade    = map[string]any{"kind": "unruled_motion", "subject": "M2", "owner": "red-chair"}
)

// drivePlans runs debate.js with the chair relaying plans[i] at its (i+1)th sitting and an empty
// plan after; every other seat returns an envelope carrying `extra`.
func drivePlans(t *testing.T, plans []map[string]any, extra Envelope) ([]Dispatch, Outcome) {
	t.Helper()
	chair := 0
	backend := func(seatID, label, prompt string) Envelope {
		e := Envelope{"synopsis": "t", "log": []any{}, "rulings": []any{}, "dispositions": []any{}, "holdings": []any{},
			"manifest": []any{}, "claim_count": 1, "saturation_reached": false, "sitting_record_appended": true, "open_gaps": 0}
		for k, v := range extra {
			e[k] = v
		}
		if seatID == "red-chair" {
			chair++
			if chair <= len(plans) {
				e["plan"] = plans[chair-1]
			} else {
				e["plan"] = planWith([]any{}, []any{})
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

// THE ENGINE HEARS THE PETITIONS THE PLAN CONVENES THE BENCH FOR, BEFORE ANY PARTY OF THE EPOCH
// (#1203). A lens files on the record; the chair's next plan readies the bench with occasion
// `petition`; the workflow seats that sitting first, then the lens.
func TestTheEngineHearsThePetitionsThePlanConvenes(t *testing.T) {
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, petitionBench}, []any{benchPetition})}, nil)
	petition := slices.IndexFunc(ds, func(d Dispatch) bool { return d.Occasion == "petition" })
	if petition < 0 {
		t.Fatalf("no petition sitting dispatched; occasions %q", occasions(ds))
	}
	if d := ds[petition]; d.SeatID != "judge" || !strings.HasPrefix(d.Label, "judge ·") {
		t.Errorf("petition sitting dispatched as %q (label %q), want the bench", d.SeatID, d.Label)
	}
	lens := slices.IndexFunc(ds, func(d Dispatch) bool { return d.SeatID == "red-lens-logic" })
	if lens < 0 || lens < petition {
		t.Errorf("the lens sat at %d and the petition sitting at %d — a petition is heard before any party of the epoch sits", lens, petition)
	}
	if n := strings.Count(strings.Join(occasions(ds), " "), "docket"); n != 0 {
		t.Errorf("%d docket sitting(s) for a bench convened only for a petition", n)
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
// no plan convenes the bench for one: no petition sitting is dispatched. The negative control
// against a second reader of the fact the record carries.
func TestAnEnvelopePetitionIsNotAChannel(t *testing.T) {
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}}}, []any{})},
		Envelope{"petitions": []any{map[string]any{"class": "safety", "basis": "b", "relief": "r"}}})
	if slices.Contains(occasions(ds), "petition") {
		t.Errorf("a petition sitting was dispatched off an envelope field: occasions %q", occasions(ds))
	}
}

// A BENCH PARTY WITH NO OCCASION IS REFUSED, NOT GUESSED AT. An empty gap list does not mean
// "petition"; a relay that dropped the field is altered in transit.
func TestABenchPartyWithoutAnOccasionIsRefused(t *testing.T) {
	wd, _ := os.Getwd()
	chair := 0
	_, _, err := CaptureRun(ScriptPath(filepath.Join(wd, "..", "..")), Config{Topic: "t", RunDir: t.TempDir(), BinDir: t.TempDir(),
		Lanes: 1, Model: "haiku", JudgmentModel: "haiku", Backend: func(seatID, label, prompt string) Envelope {
			e := Envelope{"claim_count": 1, "saturation_reached": false, "sitting_record_appended": true, "rulings": []any{}, "dispositions": []any{}, "open_gaps": 0}
			if seatID == "red-chair" {
				chair++
				e["plan"] = planWith([]any{map[string]any{"seat_id": "judge", "gap_ids": []any{}}}, []any{benchPetition})
			}
			return e
		}})
	if err == nil || !strings.Contains(err.Error(), "occasions") {
		t.Errorf("a bench party with no occasions drove to %v, want the relay refused naming the field", err)
	}
}
