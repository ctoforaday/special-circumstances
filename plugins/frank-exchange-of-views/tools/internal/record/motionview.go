package record

import (
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE MOTIONS VIEW, AND WHY IT HAD TO EXIST.
//
// A motion BLOCKS the terminal verb: `chair verdict` is refused while one is unruled,
// and the refusal names its id. Until this view, that id was all a seat could ever learn — there
// was no read verb (`motion` offers a subject's file and rule, no show) and
// no projection carried one. The ask was on the record and unreadable.
//
// A well-formed ruling on an unread motion is indistinguishable from a considered one, and nothing
// on the record marks it as written blind. So the view carries id, subject, grades and basis: the
// ask is readable before it is ruled.
//
// BASIS IS THE LOAD-BEARING FIELD. The ruling exists to answer the ask, so a view that carried
// id, subject and grades but not the argument would leave the seat exactly as blind while
// looking like it had solved the problem.

// MotionJSON is one motion and its answer, if it has one.
type MotionJSON struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Filer   string `json:"filer"`
	Epoch   int    `json:"epoch"`
	// Basis is the ask IN THE FILER'S WORDS. It is why this view exists.
	Basis  string `json:"basis"`
	Relief string `json:"relief"`

	// Ruled is false while the motion is outstanding, and an outstanding motion is what blocks
	// a PASS verdict — so the field a blocked seat is looking for is stated rather than inferred
	// from an empty ruling string.
	Ruled       bool   `json:"ruled"`
	Ruling      string `json:"ruling"`
	RulingBy    string `json:"ruling_by"`
	RulingEpoch int    `json:"ruling_epoch"`
	Opinion     string `json:"opinion"`
	// RulingRelief and RulingBinds are a granted petition's order and whom it binds; empty on every
	// other ruling. Each seat's work list carries the order too (WorkJSON.Relief).
	RulingRelief string `json:"ruling_relief"`
	RulingBinds  string `json:"ruling_binds"`

	// Fields carries the subject-specific payload — a docket motion's gap_id, a petition's class. Kept as the record holds them rather than flattened into named columns
	// that would differ per subject.
	Fields map[string]string `json:"fields"`
	// GapID is the gap the motion is ABOUT — a docket motion's — or absent for a subject about no
	// gap. A bench reading this page names the docketed gap each motion is for from it (the release
	// sweep's bench, given no gap here, filed fresh motions and ruled those while the dispatch's stood
	// unruled).
	GapID string `json:"gap_id,omitempty"`
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

// motionsJSONOf projects every motion on the record, current vocabulary and legacy alike.
func motionsJSONOf(evs []*Event, win WindowIndex) MotionsJSON {
	out := MotionsJSON{Motions: []MotionJSON{}}
	for _, m := range MotionsOf(evs, win) {
		if m == nil {
			continue
		}
		mj := MotionJSON{
			ID: m.ID, Subject: m.Subject, Filer: m.Filer, Epoch: m.Epoch,
			Basis: m.Basis, Relief: m.Relief,
			Ruled: m.Ruling != "", Ruling: m.Ruling, RulingBy: m.RulingBy,
			RulingEpoch: m.RulingEpoch, Opinion: m.Opinion,
			RulingRelief: m.RulingRelief, RulingBinds: m.RulingBinds,
		}
		mj.GapID = m.GapID
		if len(m.Fields) > 0 {
			mj.Fields = m.Fields
		}
		out.Motions = append(out.Motions, mj)
	}
	// Filing order is the order they must be answered in, and replay order is filing order —
	// but Motions reads the one live vocabulary, so a run holding
	// both would otherwise present two interleaved sequences as one.
	sort.SliceStable(out.Motions, func(i, j int) bool {
		if out.Motions[i].Epoch != out.Motions[j].Epoch {
			return out.Motions[i].Epoch < out.Motions[j].Epoch
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

// motionsView is the motions read from the record. The exchange lives on the families declared
// here; nothing else on the record decides a motion's state. A position is a motion where its
// sitting's plan permitted the PASS and the sitting records none (heardPositionsOf), so the
// position, the plan and the verdict are read with them.
var motionsView = declareNarrowedView("motions", rendersEvents(motionsJSONOf),
	recordpb.EventType_EVENT_TYPE_MOTION,
	recordpb.EventType_EVENT_TYPE_MOTION_RULE,
	recordpb.EventType_EVENT_TYPE_DOCKETING,
	recordpb.EventType_EVENT_TYPE_POSITION,
	recordpb.EventType_EVENT_TYPE_PLAN,
	recordpb.EventType_EVENT_TYPE_VERDICT)

// MotionsJSONBytes renders the motions view as indented JSON.
func MotionsJSONBytes(run Run) ([]byte, error) { return motionsView.jsonBytes(run) }
