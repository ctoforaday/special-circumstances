package fuzz

import (
	"maps"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE PROPOSED ACTS AND THEIR DOCKETINGS, GATED (#1695 §III.B9, §III.B13, §III.B14). Some keys are
// counted by the drives as they run (countGate): a proposal a ruling's tombstone bars, a side that stated nothing refused, the acts a closings-only
// sitting refuses and the closings filed there, the raiser's takes of the author's proposals and each
// party's concession of a docketing. The rest are read off the stream at the run's end
// (docketTally): the docketings by cause, the bench's answers by cause and side, the ruling tombstones
// by the side the ruling took. The sweep fails on any key at zero.

// proposalGateKeys is every key the sweep gates, with what a zero leaves unexercised.
var proposalGateKeys = []struct{ key, what string }{
	{"docketing:impasse", "docketings at impasse"},
	{"docketing:proposal", "docketings holding a proposal"},
	{"docketing:grade_movement", "docketings at grade movement"},
	{"answer:impasse:raiser", "bench answers for the raiser at impasse"},
	{"answer:impasse:author", "bench answers for the author at impasse"},
	{"answer:impasse:remand", "remands at impasse"},
	{"answer:proposal:raiser", "bench answers for the raiser on a proposal"},
	{"answer:proposal:author", "bench answers for the author on a proposal"},
	{"answer:grade_movement:kept", "grade-movement answers keeping the moved grades (--keep)"},
	{"answer:grade_movement:restored", "grade-movement answers restoring the prior grades (--restore)"},
	{"tombstone:placed:raiser", "ruling tombstones placed by a ruling for the raiser"},
	{"tombstone:placed:author", "ruling tombstones placed by a ruling for the author"},
	{"estopped_refused:regrade", "proposed regrades refused by a ruling's standing tombstone"},
	{"closing_no_text", "red closings stating no text (--no-text)"},
	{"closings_only_refused", "acts other than a closing refused in a closings-only sitting"},
	{"final_sitting_closings:red", "red closings filed in a closings-only sitting"},
	{"final_sitting_closings:blue", "blue closings filed in a closings-only sitting"},
	{"motion:blue:close", "close proposals filed by blue, the gaps' author"},
	{"motion:blue:regrade", "regrade proposals filed by blue, the gaps' author"},
	{"accept:proposal:regrade", "the author's proposed regrades the raiser took (regrade --accept)"},
	{"accept:proposal:close", "the author's proposed closes the raiser took (close --accept)"},
	{"accept:proposal:concession", "takes of a proposal its docketing held, the raiser's concession"},
	{"concession:raiser:close", "the raiser's closes of a docketed gap, its concession"},
	{"concession:author:edit", "the author's accepts of the raiser's text on a docketed gap (edit --accept), its concession"},
}

// proposalOracles is every proposalOracle key, each a defect at ANY count.
var proposalOracles = []struct{ key, what string }{
	{"oracle:two_answers", "a proposal answered by two rulings"},
	{"oracle:docketing_answered_twice", "a docketing ruled twice"},
	{"oracle:docket_motion_without_act", "a docket motion stating no act"},
}

// proposalStream is the stream's own count of the proposals and their answers, folded from the events
// alone — independent of record.DocketingsOf, so a fold that drops a ruling cannot hide a second one:
// each docket motion by its filer's role and act (motion:<role>:<act>), and each oracle of
// proposalOracles.
func proposalStream(evs []*record.Event) map[string]int {
	out := map[string]int{}
	byMotion, byDocketing := map[string]int{}, map[string]int{}
	for _, e := range evs {
		switch b, _ := recordpb.Body(e); b := b.(type) {
		case *recordpb.Motion:
			if b.GetDocket() == nil {
				continue
			}
			act := record.ActWord(b)
			if act == "" {
				out["oracle:docket_motion_without_act"]++
				continue
			}
			out["motion:"+seatclass.RoleOfSeat(e.GetSeatId())+":"+act]++
		case *recordpb.MotionRule:
			if id := b.GetMotionId(); id != "" {
				if byMotion[id]++; byMotion[id] == 2 {
					out["oracle:two_answers"]++
				}
			}
			if id := b.GetDocket().GetDocketingId(); id != "" {
				if byDocketing[id]++; byDocketing[id] == 2 {
					out["oracle:docketing_answered_twice"]++
				}
			}
		}
	}
	return out
}

// countGate adds one to key's count.
func (r *runner) countGate(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.proposalGates == nil {
		r.proposalGates = map[string]int{}
	}
	r.proposalGates[key]++
}

// docketTally is the run's gate counts: the drives' own (countGate), joined by what the stream holds at
// the run's end — each docketing by cause, each bench answer by its docketing's cause and its side
// (sided_with, moved_grades, or remand), and each ruling tombstone placed by the side its ruling took.
func (r *runner) docketTally() map[string]int {
	r.mu.Lock()
	out := maps.Clone(r.proposalGates)
	r.mu.Unlock()
	if out == nil {
		out = map[string]int{}
	}
	side := map[string]string{} // a docketing id -> the side its ruling took
	for _, d := range r.docketings() {
		cause := recordpb.Word(d.Cause)
		out["docketing:"+cause]++
		if !d.Ruled() {
			continue
		}
		dr := d.Ruling.GetDocket()
		answer := ""
		switch {
		case dr.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED:
			answer = "remand"
		case dr.GetMovedGrades() != recordpb.MovedGrades_MOVED_GRADES_UNSPECIFIED:
			answer = recordpb.Word(dr.GetMovedGrades())
		case dr.GetSidedWith() != recordpb.SidedWith_SIDED_WITH_UNSPECIFIED:
			answer = recordpb.Word(dr.GetSidedWith())
		default:
			answer = "unsided"
		}
		out["answer:"+cause+":"+answer]++
		side[d.ID] = answer
	}
	stream, _, err := record.EventsOf(r.run(), recordpb.EventType_EVENT_TYPE_MOTION, recordpb.EventType_EVENT_TYPE_MOTION_RULE)
	if err != nil {
		r.noteActErr("the record's proposals and rulings could not be read: " + err.Error())
		return out
	}
	for k, n := range proposalStream(stream) {
		out[k] += n
	}
	evs, _, err := record.EventsOf(r.run(), recordpb.EventType_EVENT_TYPE_TOMBSTONE)
	if err != nil {
		r.noteActErr("the record's tombstones could not be read: " + err.Error())
		return out
	}
	for _, e := range evs {
		t, ok := recordpb.BodyAs[*recordpb.Tombstone](e)
		if !ok || t.GetKind() != recordpb.TombstoneKind_TOMBSTONE_KIND_RULING {
			continue
		}
		if t.GetEnds() {
			out["tombstone:ended"]++
			continue
		}
		out["tombstone:placed:"+side[t.GetDocketingId()]]++
	}
	return out
}

// closingsOnlyClause is the words debate.js puts in a party's prompt for a sitting that admits its
// closings alone (closingsOnlyClause): every party's at the run's final epoch, and blue's in the epoch
// that holds it out.
const closingsOnlyClause = "THIS SITTING IS FOR YOUR CLOSINGS ALONE"

// closingsOnlyRefusal is the words the record refuses any other act in such a sitting with
// (record.requireClosingsOnlySitting), the final epoch's and the held-out epoch's alike.
const closingsOnlyRefusal = "for the closings it owes and nothing else"

// closingsOnlyAt reports whether a party of role sits for its closings alone in epoch under the run's
// epoch limit, by the record's definitions: every party at the final epoch, and blue in the epoch that
// holds it out (record.Plan.BlueHeldOut).
func closingsOnlyAt(epoch, limit int, role string) bool {
	return finalEpoch(epoch, limit) || role != "lens" && (record.Plan{MaxEpochs: limit, Epoch: epoch}).BlueHeldOut()
}

// blueClosingsOnly reports whether the epoch the latest chair plan opened holds blue to its closings:
// an act of blue's the drive issues outside blue's own sitting — an answer a lens's close waits on —
// is refused there, so the drive issues none.
func (r *runner) blueClosingsOnly() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return closingsOnlyAt(r.planEpoch, r.planLimit, "blue")
}

// closingsOnlySitting is a party's sitting whose prompt admits its closings alone: it reaches once for
// an act other than a closing — a petition, which needs nothing on the board — and the record refuses
// it, then files each closing it owes and ends. Its envelope is empty: it engaged nothing.
func (r *runner) closingsOnlySitting(role, seatID string) map[string]any {
	r.sit(role, seatID)
	out, err := r.exec("--json", "motion", "petition", "file", "--seat-id", seatID,
		"--class", pick(r.rng, petitionClasses), "--relief", "fuzz relief",
		"--reason", "fuzz: a petition in a sitting for closings alone")
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(seatID + "'s petition in a sitting for its closings alone was admitted: " + firstLine(out))
	case ok && env.Code == string(feov.Conflict) && strings.Contains(env.Error, closingsOnlyRefusal):
		r.countGate("closings_only_refused")
	default:
		r.noteActErr(seatID + "'s petition in a sitting for its closings alone was refused, and not as that sitting's: " + firstLine(out))
	}
	r.mu.Lock()
	before := r.closingsFiled
	r.mu.Unlock()
	r.fileClosings(seatID)
	r.mu.Lock()
	filed := r.closingsFiled - before
	r.mu.Unlock()
	side := "blue"
	if seatclass.RoleOfSeat(seatID) == "lens" {
		side = "red"
	}
	for range filed {
		r.countGate("final_sitting_closings:" + side)
	}
	return map[string]any{}
}

// Every gated key names a cell its counter writes: a key no counter can write fails the sweep on every
// run, and reads as a path unexercised rather than a key misspelled.
func TestEveryProposalGateKeyIsOneItsCountersWrite(t *testing.T) {
	causes := map[string]bool{}
	for _, c := range []recordpb.DocketingCause{recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE,
		recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT} {
		causes[recordpb.Word(c)] = true
	}
	answers := map[string]bool{"remand": true}
	for _, s := range []recordpb.SidedWith{recordpb.SidedWith_SIDED_WITH_RAISER, recordpb.SidedWith_SIDED_WITH_AUTHOR} {
		answers[recordpb.Word(s)] = true
	}
	for _, m := range []recordpb.MovedGrades{recordpb.MovedGrades_MOVED_GRADES_KEPT, recordpb.MovedGrades_MOVED_GRADES_RESTORED} {
		answers[recordpb.Word(m)] = true
	}
	drives := map[string]bool{"estopped_refused:regrade": true,
		"closing_no_text": true, "closings_only_refused": true,
		"final_sitting_closings:red": true, "final_sitting_closings:blue": true,
		"accept:proposal:regrade": true, "accept:proposal:close": true, "accept:proposal:concession": true,
		"concession:raiser:close": true, "concession:author:edit": true}
	// ONLY THE AUTHOR FILES A PROPOSAL (record.requireFiledByAuthor), so blue is the one filer role.
	roles := map[string]bool{"blue": true}
	acts := map[string]bool{}
	for _, a := range []*recordpb.Motion{{Act: &recordpb.Motion_Close{}}, {Act: &recordpb.Motion_Regrade{}}} {
		acts[record.ActWord(a)] = true
	}
	for _, g := range proposalGateKeys {
		parts := strings.Split(g.key, ":")
		switch {
		case drives[g.key]:
		case parts[0] == "motion" && len(parts) == 3 && roles[parts[1]] && acts[parts[2]]:
		case parts[0] == "docketing" && len(parts) == 2 && causes[parts[1]]:
		case parts[0] == "answer" && len(parts) == 3 && causes[parts[1]] && answers[parts[2]]:
		case parts[0] == "tombstone" && len(parts) == 3 && parts[1] == "placed" && answers[parts[2]]:
		default:
			t.Errorf("gate %q (%s) is no cell a counter writes", g.key, g.what)
		}
	}
}

// THE ORACLES COUNT WHAT THEY NAME, and nothing on a clean stream: a docket motion with no act, a
// proposal answered twice and a docketing ruled twice are planted, and each is counted once; the
// filer census counts each motion by its filer's role and act.
func TestTheProposalOraclesCountAPlantedDefect(t *testing.T) {
	ev := func(seat string, body proto.Message) *record.Event {
		e := &record.Event{SeatId: proto.String(seat)}
		if _, err := recordpb.SetBody(e, body); err != nil {
			t.Fatal(err)
		}
		return e
	}
	docket := func(act bool) *recordpb.Motion {
		m := &recordpb.Motion{MotionId: proto.String("M-1"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G-1")}}}
		if act {
			m.Act = &recordpb.Motion_Close{Close: &recordpb.Close{}}
		}
		return m
	}
	rule := func(motion, docketing string) *recordpb.MotionRule {
		r := &recordpb.MotionRule{}
		if motion != "" {
			r.MotionId = proto.String(motion)
		}
		if docketing != "" {
			r.Ruling = &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{DocketingId: proto.String(docketing)}}
		}
		return r
	}
	clean := []*record.Event{ev("blue-respond", docket(true)), ev(fuzzLensSeats[0], rule("M-1", "")), ev("judge", rule("", "D-1"))}
	got := proposalStream(clean)
	for _, o := range proposalOracles {
		if got[o.key] != 0 {
			t.Errorf("a clean stream counts %s = %d", o.key, got[o.key])
		}
	}
	if got["motion:blue:close"] != 1 {
		t.Errorf("motion:blue:close = %d, want 1", got["motion:blue:close"])
	}
	planted := append(clean, ev("red-chair", docket(false)), ev("judge", rule("M-1", "")), ev("judge", rule("", "D-1")))
	got = proposalStream(planted)
	for _, o := range proposalOracles {
		if got[o.key] != 1 {
			t.Errorf("%s over a stream planted with one = %d, want 1", o.key, got[o.key])
		}
	}
}
