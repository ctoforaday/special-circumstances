package record

import (
	"encoding/json"
	"sort"
)

// THE MOTIONS VIEW, AND WHY IT HAD TO EXIST.
//
// A motion BLOCKS the terminal verb: `merge verdict --as PASS` is refused while one is unruled,
// and the refusal names its id. Until this view, that id was all a seat could ever learn — there
// was no read verb (`motion` offers direction|grade|petition, each file/rule/appeal, no show) and
// no projection carried one. The ask was on the record and unreadable.
//
// Measured, from a probed merge seat blocked by exactly this. It tried `motion show --id M1`,
// read `motion --help` and concluded in its own words "there's a motion command but no show
// subcommand", then searched six views, `--json`, `jq` and three help pages across ten-plus
// calls. Then: "Since I don't know what direction blue proposed, let me reason about what's
// likely." It guessed the SUBJECT wrong, recovered the right one from a dangling-reference error,
// and ruled `rejected` on an argument it had never read — against blue's actual filing, which
// said the defect was presentational and `certain` severity priced a rewrite as a data error.
// The ruling asserts the precise proposition blue disputed.
//
// Nothing on the record marks that answer as written blind, and nothing could: a well-formed
// ruling on an unread motion is indistinguishable from a considered one.
//
// BASIS IS THE LOAD-BEARING FIELD. The ruling exists to answer the ask, so a view that carried
// id, subject and grades but not the argument would leave the seat exactly as blind while
// looking like it had solved the problem.

// MotionJSON is one motion and its answer, if it has one.
type MotionJSON struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Filer   string `json:"filer"`
	Round   int    `json:"round"`
	// Basis is the ask IN THE FILER'S WORDS. It is why this view exists.
	Basis  string `json:"basis,omitempty"`
	Relief string `json:"relief,omitempty"`

	// Ruled is false while the motion is outstanding, and an outstanding motion is what blocks
	// a PASS verdict — so the field a blocked seat is looking for is stated rather than inferred
	// from an empty ruling string.
	Ruled       bool   `json:"ruled"`
	Ruling      string `json:"ruling,omitempty"`
	RulingBy    string `json:"ruling_by,omitempty"`
	RulingRound int    `json:"ruling_round,omitempty"`
	Opinion     string `json:"opinion,omitempty"`

	Appealed     bool   `json:"appealed,omitempty"`
	AppealReason string `json:"appeal_reason,omitempty"`

	// Fields carries the subject-specific payload — a grade motion's gap_id/dimension/proposed,
	// a petition's class. Kept as the record holds them rather than flattened into named columns
	// that would differ per subject.
	Fields map[string]string `json:"fields,omitempty"`
}

// MotionsJSON is the whole docket, with the outstanding count a blocked seat needs.
type MotionsJSON struct {
	Motions []MotionJSON `json:"motions"`
	Counts  struct {
		Total       int `json:"total"`
		Ruled       int `json:"ruled"`
		Outstanding int `json:"outstanding"`
	} `json:"counts"`
}

// MotionsJSONOf projects every motion on the record, current vocabulary and legacy alike.
func MotionsJSONOf(b *Board) MotionsJSON {
	out := MotionsJSON{Motions: []MotionJSON{}}
	for _, m := range Motions(b) {
		if m == nil {
			continue
		}
		mj := MotionJSON{
			ID: m.ID, Subject: m.Subject, Filer: m.Filer, Round: m.Round,
			Basis: m.Basis, Relief: m.Relief,
			Ruled: m.Ruling != "", Ruling: m.Ruling, RulingBy: m.RulingBy,
			RulingRound: m.RulingRound, Opinion: m.Opinion,
			Appealed: m.Appealed, AppealReason: m.AppealReason,
		}
		if len(m.Fields) > 0 {
			mj.Fields = m.Fields
		}
		out.Motions = append(out.Motions, mj)
	}
	// Filing order is the order they must be answered in, and replay order is filing order —
	// but Motions reads the one live vocabulary, so a run holding
	// both would otherwise present two interleaved sequences as one.
	sort.SliceStable(out.Motions, func(i, j int) bool {
		if out.Motions[i].Round != out.Motions[j].Round {
			return out.Motions[i].Round < out.Motions[j].Round
		}
		return out.Motions[i].ID < out.Motions[j].ID
	})
	out.Counts.Total = len(out.Motions)
	for _, m := range out.Motions {
		if m.Ruled {
			out.Counts.Ruled++
		}
	}
	out.Counts.Outstanding = out.Counts.Total - out.Counts.Ruled
	return out
}

// MotionsJSONBytes renders the motions view as indented JSON.
func MotionsJSONBytes(runDir string) ([]byte, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(MotionsJSONOf(b), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
