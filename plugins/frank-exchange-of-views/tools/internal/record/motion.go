package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// ONE ADJUDICATION MECHANISM, WITH AN ID (#344).
//
// The propose→rule exchange was implemented three times, with three vocabularies and no shared
// identity:
//
//	directions   blue avenue      -> merge avenue-rule       key `ruling`
//	governance   <seat> petition  -> bench petition-rule      key `ruling`
//	grades       blue dispute     -> merge dispute-respond    key `response`
//
// Two spellings of one concept, three renderers, and nothing tying an ask to its answer. That is
// the direct cause of a defect class fixed one instance at a time: #315 found the petition FILING
// unrendered while the avenue RULING was found unrendered SEPARATELY in the same sweep, because
// nothing said they were the same mechanism. #312 is the same root — `petition-rule` joins on
// `(petitioner, class)` with no id, which is why the report renders filings and rulings side by
// side rather than joined: pairing two filings by one seat in one round would be a guess.
//
// A motion has an ID. The ask and its answer join on it, once, and one renderer serves all three.

// MotionSubjects are what a motion can be ABOUT. Each carries its own required payload and its
// own ruler, which is why the CLI subgroups them rather than taking a --on flag: cobra cannot
// express "required only when --on=grade", so a flag-discerned subject would put three divergent
// contracts into hand-written RunE validation — a flag combination policed by prose, which is
// the shape this suite exists to remove.
var MotionSubjects = []string{"grade", "petition", "direction"}

// MotionVerdicts are the rulings, per subject. The KEY is `ruling` on every one of them and the
// flag is `--as` on every one of them, which is the point: §I of the plan names
// `ruling`/`ruling`/`response` as the structural defect, and a collapse that kept three
// spellings would have reproduced it inside the new group.
var MotionVerdicts = map[string][]string{
	"grade":     {"accepted", "rejected"},
	"petition":  {"granted", "denied"},
	"direction": {"endorsed", "out-of-scope", "too-thin"},
}

// MintMotionID assigns the next run-unique motion id (M1, M2 …).
//
// Run-unique rather than round-scoped, for the reason an avenue's is: a motion OUTLIVES the round
// that filed it — a grade dispute rejected in round 2 is re-disputed in round 3 and appealed to
// the bench in round 4 — so a round-scoped id would have to be re-minted to survive, and the
// re-mint is where the thread breaks.
func MintMotionID(runDir string) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	n := 0
	for _, e := range m.Events {
		if e.Type == "motion" && e.Payload.Str("motion_id") != "" {
			n++
		}
	}
	return fmt.Sprintf("M%d", n+1), nil
}

// Motion is one exchange after replay: what was asked, how it was ruled, and whether the filer
// pressed it further.
type Motion struct {
	ID      string
	Subject string
	Filer   string
	Round   int
	Basis   string // the ask, in the filer's words
	Relief  string // what the filer wants, stated so a seat can act on it without the argument

	// Ruling is empty until ruled. THE ABSENCE IS INFORMATION: a motion filed and never ruled
	// means the sitting did not happen, and the report says so rather than omitting the row.
	Ruling      string
	RulingBy    string
	RulingRound int
	Opinion     string

	// Appeal is the filer pressing on after a ruling — blue pursuing a direction ruled
	// out-of-scope, or re-disputing a rejected grade. `contests_ruling` was a bespoke field on
	// one of the three exchanges; here it is the same act on all of them.
	Appealed     bool
	AppealReason string

	// Subject-specific payload, carried rather than re-derived.
	Fields map[string]string
}

// Ruled reports whether the motion has an answer.
func (m Motion) Ruled() bool { return m.Ruling != "" }

// Motions replays the motion events into current state, in filing order.
//
// It reads the `motion` types only. Pre-collapse records carry `dispute`/`dispute-respond`,
// `petition`/`petition-rule` and `avenue-rule` instead, and those are handled by the DUAL-READ in
// compat.go — deliberately separate, so the shape of a motion is not bent to accommodate the
// three shapes it replaced.
func Motions(b *Board) []*Motion {
	byID := map[string]*Motion{}
	var order []string
	for _, e := range b.Events {
		id := e.Payload.Str("motion_id")
		if id == "" {
			continue
		}
		switch e.Type {
		case "motion":
			m, ok := byID[id]
			if !ok {
				m = &Motion{ID: id, Fields: map[string]string{}}
				byID[id] = m
				order = append(order, id)
			}
			m.Subject, m.Filer, m.Round = e.Payload.Str("subject"), e.SeatID, e.Round
			m.Basis, m.Relief = e.Payload.Str("basis"), e.Payload.Str("relief")
			for _, k := range []string{"gap_id", "dimension", "proposed", "class", "avenue_id"} {
				if v := e.Payload.Str(k); v != "" {
					m.Fields[k] = v
				}
			}
		case "motion-rule":
			m, ok := byID[id]
			if !ok {
				continue // a ruling naming no filing; refs.go refuses this at the write
			}
			m.Ruling, m.RulingBy, m.RulingRound = e.Payload.Str("ruling"), e.SeatID, e.Round
			m.Opinion = e.Payload.Str("opinion")
		case "motion-appeal":
			m, ok := byID[id]
			if !ok {
				continue
			}
			m.Appealed, m.AppealReason = true, e.Payload.Str("reason")
		}
	}
	out := make([]*Motion, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// RequireMotionRef refuses a ruling or appeal naming a motion no filing created — the same
// discipline every other cross-reference gets, for the same reason: a dangling reference is
// accepted at write time and dropped at replay, where nobody sees it go.
func RequireMotionRef(runDir, id string) error {
	if id == "" {
		return fmt.Errorf("record: --id is required — a ruling names the motion it answers, and that join is the whole of #312")
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.Type == "motion" && e.Payload.Str("motion_id") == id {
			return nil
		}
	}
	return fmt.Errorf("record: --id names motion %s, which no filing created — a dangling reference is accepted here and dropped at replay", id)
}

// MotionVerdictEnum builds the enum entry for a subject's ruling, so the CLI's help and the write
// check are generated from one table rather than stated twice.
func MotionVerdictEnum(subject string) EnumField {
	return EnumField{
		Key: "ruling", Flag: flags.As, Values: MotionVerdicts[subject],
		Why: "the ruling is what BINDS the coming seats, and every downstream reader switches on it; an unrecognized verdict reads as no ruling at all, so a refusal silently becomes permission",
	}
}
