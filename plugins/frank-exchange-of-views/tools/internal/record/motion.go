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
var MotionVerdicts = map[string][]EnumValue{
	"grade": {
		Ev("accepted", "the filer is right and the grade moves — say so, then MOVE it with `merge regrade`; accepting without regrading is a channel with no consequence"),
		Ev("rejected", "the grade stands. Your --reason is what the filer appeals against, so it carries the argument, not the conclusion"),
	},
	"petition": {
		Ev("granted", "the objection holds. The relief BINDS the seats that come after, so state it as an instruction they can follow"),
		Ev("denied", "the objection does not hold, and your reason must say why at the leaf — a refusal without one is a decoration the petitioner cannot contest"),
	},
	"direction": {
		Ev("endorsed", "worth this run's time — blue should take it up"),
		Ev("out-of-scope", "a real question, but not THIS question"),
		Ev("too-thin", "in scope, and the hypothesis does not carry its budget as stated"),
	},
}

// MotionFields are the ENUMERATED payload fields a subject's filing carries, beyond the verdict.
//
// They live here for the reason MotionVerdicts does: the set is keyed on (subject, key) and
// EnumFields cannot express that — it is keyed by event TYPE, and one `motion` event carries a
// grade `dimension` or a petition `class` depending on what it is about.
//
// THIS TABLE EXISTS BECAUSE DELETING THE OLD VERBS EXPOSED THAT THE NEW ONES HAD LOST THEIR
// CONTRACT. `blue dispute --dimension` was enum-validated and `--proposed` was a pflag.Value that
// refused a non-grade at parse; `motion grade file` registered both as plain strings and accepted
// anything. The additive stage added the verb and nobody diffed its flag contract against the one
// it replaces — which is the half-state that reads as done, one layer below where that rule is
// usually applied.
var MotionFields = map[string]map[string][]EnumValue{
	"grade": {"dimension": {
		Ev("severity", "how bad the defect is in itself"),
		Ev("likelihood", "how likely the CONSEQUENCE is — never how likely the defect is to exist; that is the `existence` axis"),
		Ev("impact", "how bad the consequence is if it lands"),
		Ev("complexity_cost", "what fixing it costs — the axis to contest when the fix is worth more than the defect. THE FLAG IS `--cx`, not `--complexity_cost`: the other three dimensions ARE their flag names and this one is not, which is a trap a seat walks into by learning the pattern from the other three (measured)"),
	}},
	"petition": {"class": {
		Ev("ethical", "proceeding would require acting against the interests of someone the run affects"),
		Ev("safety", "proceeding would create or conceal a hazard"),
		Ev("integrity", "proceeding would require asserting what you believe false, or burying a real finding"),
		Ev("constitutional", "the instruction itself conflicts with the rules the run is bound by"),
	}},
}

// MotionFieldEnum builds the enum entry for one of those fields, so the CLI's help and the write
// check are generated from one table rather than stated twice. It returns ok=false for a
// (subject, key) with no set, which is how a caller distinguishes "free text" from "empty set".
func MotionFieldEnum(subject, key string, flag string) (EnumField, bool) {
	values, ok := MotionFields[subject][key]
	if !ok {
		return EnumField{}, false
	}
	why := map[string]string{
		"dimension": "the ruling is matched to the filing on (gap, dimension) and the gap's grade is then read AT that dimension: an axis outside the four matches no ruling and reads no grade, so the motion is filed against nothing",
		"class":     "the class is what the seat is ASKING the bench to sit on, and the bench is convened per class; a fifth is a petition heard under whichever standard the ruling seat happened to imagine",
	}[key]
	return EnumField{Key: key, Flag: flag, Values: values, Why: why}, true
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

	// Avenue PROPOSALS, indexed by their id — the filing half of every direction motion. Gathered
	// in the same pass and read only when a ruling arrives, because a proposal nobody ruled on is
	// not a motion (see the motion-rule arm).
	type proposal struct {
		filer, basis string
		round        int
	}
	proposals := map[string]proposal{}

	// TWO PASSES, FOR THE REASON compat.go SPELLS OUT, AND I WROTE THE BUG HERE ANYWAY.
	//
	// Each seat writes its own shard and replay orders by TIMESTAMP across all of them, so the
	// shards interleave and A RULING CAN REPLAY BEFORE ITS FILING — the bench rules a petition blue
	// filed, and the two events live in different shards. A single pass that files-then-rules drops
	// every ruling that lands first, silently: the motion still renders, as an ask nobody answered.
	//
	// I documented exactly this in compat.go after the fixture caught it, and then shipped the same
	// single pass in the function compat.go exists to be the legacy twin of. What caught it was the
	// prose gate (#320) — "judge-petition/motion-rule prose absent from report" on 25 of 60 seeds —
	// not the reasoning that had already been written down one file over.
	// A PASS OF ITS OWN, for the same interleaving reason: blue proposes the avenue and the merge
	// rules it, so the two live in different shards and the ruling can replay first. Gathered
	// inside pass 1 this map was read before it was filled, and a direction motion came out with
	// no filer, no round and no ask — rendering as an answer to a question nobody asked.
	for _, e := range b.Events {
		if e.Type == "avenue" && e.Payload.Str("supersedes_status") == "" {
			if a := e.Payload.Str("avenue_id"); a != "" {
				proposals[a] = proposal{filer: e.SeatID, basis: e.Payload.Str("line"), round: e.Round}
			}
		}
	}

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
			// PASS 1 only CREATES. A direction motion has no filing event of its own, so its
			// ruling is also its creation; every other subject is created by its `motion` event,
			// which this pass may not have reached yet. The ruling itself is attached in pass 2.
			if _, ok := byID[id]; !ok {
				if e.Payload.Str("subject") != "direction" {
					continue // a ruling naming no filing; RequireMotionSubjectRef refuses this at the write
				}
				// A DIRECTION MOTION IS CREATED BY ITS RULING, and that is not a special case
				// bolted on — it is the only shape that does not invent identity. `direction` has
				// no `file` verb because the PROPOSAL is the filing (`blue avenue`), so the id it
				// joins on is the avenue's own A-id: minted, refusable, already on the record.
				//
				// The rejected alternative was minting an M-id at propose time. It would have
				// filed a motion for every line blue ever floated — ~60 a run in the fuzz against
				// 45 rulings — so the report's "N motions received no ruling" would have been
				// dominated by lines nobody was ever going to rule on, and the one signal that
				// message exists to carry (a sitting that did not happen) would have drowned in it.
				//
				// compat.go builds the legacy direction motion the same way, from `avenue-rule`.
				// One shape for both vocabularies, which is what makes the dual-read a translation
				// rather than a second model.
				m := &Motion{ID: id, Subject: "direction", Fields: map[string]string{"avenue_id": id}}
				if p, ok := proposals[id]; ok {
					m.Filer, m.Round, m.Basis = p.filer, p.round, p.basis
				}
				byID[id] = m
				order = append(order, id)
			}
		}
	}

	// PASS 2 attaches every answer. Order-independent by construction, which is the only safe
	// assumption about a multi-shard log.
	for _, e := range b.Events {
		id := e.Payload.Str("motion_id")
		if id == "" {
			continue
		}
		m, ok := byID[id]
		if !ok {
			continue // a ruling or appeal naming no filing; the write-side ref checks refuse these
		}
		switch e.Type {
		case "motion-rule":
			m.Ruling, m.RulingBy, m.RulingRound = e.Payload.Str("ruling"), e.SeatID, e.Round
			m.Opinion = e.Payload.Str("opinion")
		case "motion-appeal":
			m.Appealed, m.AppealReason = true, e.Payload.Str("reason")
		}
	}

	out := make([]*Motion, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// RequireMotionSubjectRef refuses a ruling or appeal naming a motion no filing created — the same
// discipline every other cross-reference gets, for the same reason: a dangling reference is
// accepted at write time and dropped at replay, where nobody sees it go.
//
// It takes the SUBJECT because the subjects do not share a filing verb. `grade` and `petition` are
// filed by `motion <subject> file` and join on the M-id it mints; `direction` has no file verb —
// the proposal is the filing — so it joins on the AVENUE's own id, and the thing that must exist
// is the avenue, not a motion event. Passing the subject keeps that difference in one place
// instead of pushing it into each RunE.
func RequireMotionSubjectRef(runDir, subject, id string) error {
	if id == "" {
		return fmt.Errorf("record: --id is required — a ruling names the motion it answers, and that join is the whole of #312")
	}
	if subject == "direction" {
		return RequireAvenueRef(runDir, id)
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

// RequireSubjectMatches refuses a ruling whose SUBGROUP disagrees with the motion's own subject.
//
// MEASURED BY PROBING, and it is the defect the collapse was supposed to remove, reintroduced one
// level down. `motion <subject> rule` takes its subject from its POSITION IN THE TREE — which
// subgroup you typed — and validated the verdict against that, never against the subject the
// motion was actually filed under. So `motion grade rule --id M1` on a PETITION motion was
// accepted: it bypassed the gavel (the merge ruling what only the bench may rule) AND the verdict
// vocabulary (`accepted`, which is not a petition ruling at all). The report then rendered
// "petition (safety) … ruled accepted".
//
// A fact recovered from tree position rather than read from the record is exactly what
// facts-are-fields is about. The record carries the subject; this reads it.
func RequireSubjectMatches(runDir, subject, id string) error {
	got, err := MotionSubjectOf(runDir, id)
	if err != nil {
		return err
	}
	if got != subject {
		return fmt.Errorf("record: motion %s was filed as a %s motion and you are ruling it as a %s — rule it under `motion %s rule`. The subject decides BOTH who holds the gavel and which verdicts exist, so ruling under the wrong one answers with a vocabulary the motion does not have",
			id, got, subject, got)
	}
	return nil
}

// MotionSubjectOf reports what a motion is ABOUT, from the record rather than from the caller.
//
// A direction has no filing event — the proposal is the filing — so an id that resolves to an
// avenue IS a direction motion by construction.
func MotionSubjectOf(runDir, id string) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for _, e := range m.Events {
		if e.Type == "motion" && e.Payload.Str("motion_id") == id {
			return e.Payload.Str("subject"), nil
		}
	}
	for _, e := range m.Events {
		if e.Type == "avenue" && e.Payload.Str("avenue_id") == id {
			return "direction", nil
		}
	}
	return "", fmt.Errorf("record: --id names motion %s, which no filing created — a dangling reference is accepted here and dropped at replay", id)
}

// RequireUnruledMotion refuses a SECOND ruling on a motion already answered.
//
// MEASURED BY PROBING: a petition ruled `accepted` by the merge was then ruled `denied` by the
// bench, both accepted, and the report showed ONE of them with no sign the other existed. A
// direction ruled `endorsed` was re-ruled `out-of-scope` the same way. Replay keeps whichever the
// ordering happens to favour, so the answer a reader sees is decided by shard interleaving.
//
// Two writers disagreeing about one fate is the defect the avenue code already guards against by
// giving moves a single writer; a ruling had no such guard. The escalation path is an APPEAL,
// which is a new event that preserves both positions, rather than a second ruling that erases one.
func RequireUnruledMotion(runDir, id string) error {
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.Type == "motion-rule" && e.Payload.Str("motion_id") == id {
			return fmt.Errorf("record: motion %s is already ruled %q by %s. A second ruling does not overturn the first — it replays as whichever the shard ordering favours, and the other disappears. To press it, `appeal` it: an appeal keeps both positions on the record, which is the whole reason a ruling is an argument rather than a command",
				id, e.Payload.Str("ruling"), e.SeatID)
		}
	}
	return nil
}

// RequireRuledMotion additionally refuses an appeal against a motion NOBODY HAS RULED.
//
// An appeal is the filer pressing on AFTER an answer; against no answer there is nothing to press
// against, and the event would replay as a motion both unruled and appealed — a state the report
// has no honest sentence for. The check covers every subject rather than the one that surfaced it:
// appealing an unruled grade is the same nonsense as appealing an unruled direction, and fixing
// only the instance is how the class survives.
func RequireRuledMotion(runDir, subject, id string) error {
	if err := RequireMotionSubjectRef(runDir, subject, id); err != nil {
		return err
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		switch e.Type {
		case "motion-rule":
			if e.Payload.Str("motion_id") == id {
				return nil
			}
		case "avenue-rule":
			// The pre-collapse spelling. An appeal filed under the new vocabulary against a
			// ruling written under the old one is a real case for as long as both are live.
			if e.Payload.Str("avenue_id") == id {
				return nil
			}
		}
	}
	return fmt.Errorf("record: %s motion %s has no ruling to appeal — an appeal presses on after an answer, and there is no answer on the record yet", subject, id)
}

// MotionVerdictEnum builds the enum entry for a subject's ruling, so the CLI's help and the write
// check are generated from one table rather than stated twice.
func MotionVerdictEnum(subject string) EnumField {
	return EnumField{
		Key: "ruling", Flag: flags.As, Values: MotionVerdicts[subject],
		Why: "the ruling is what BINDS the coming seats, and every downstream reader switches on it; an unrecognized verdict reads as no ruling at all, so a refusal silently becomes permission",
	}
}
