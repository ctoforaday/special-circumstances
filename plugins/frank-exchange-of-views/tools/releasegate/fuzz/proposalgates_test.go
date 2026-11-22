package fuzz

import (
	"maps"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE PROPOSED ACTS AND THEIR DOCKETINGS, GATED (#1695 §III.B9, §III.B13, §III.B14). Some keys are
// counted by the drives as they run (countGate): the successor a filer reads refused and answers, a
// proposal a ruling's tombstone bars, a side that stated nothing refused, the acts a closings-only
// sitting refuses and the closings filed there. The rest are read off the stream at the run's end
// (docketTally): the docketings by cause, the bench's answers by cause and side, the ruling tombstones
// by the side the ruling took. The sweep fails on any key at zero.

// proposalGateKeys is every key the sweep gates, with what a zero leaves unexercised.
var proposalGateKeys = []struct{ key, what string }{
	{"docketing:impasse", "docketings at impasse"},
	{"docketing:proposal", "docketings holding a proposal"},
	{"docketing:grade_movement", "docketings at grade movement"},
	{"answer:impasse:red", "bench answers for red at impasse"},
	{"answer:impasse:blue", "bench answers for blue at impasse"},
	{"answer:impasse:remand", "remands at impasse"},
	{"answer:proposal:red", "bench answers for red on a proposal"},
	{"answer:proposal:blue", "bench answers for blue on a proposal"},
	{"answer:grade_movement:kept", "grade-movement answers keeping the moved grades (--keep)"},
	{"answer:grade_movement:restored", "grade-movement answers restoring the prior grades (--restore)"},
	{"tombstone:placed:red", "ruling tombstones placed by a ruling for red"},
	{"tombstone:placed:blue", "ruling tombstones placed by a ruling for blue"},
	{"estopped_refused:regrade", "proposed regrades refused by a ruling's standing tombstone"},
	{"estopped_refused:supersede", "proposed successors refused by a ruling's standing tombstone"},
	{"closing_no_text", "red closings stating no text (--no-text)"},
	{"closings_only_refused", "acts other than a closing refused in a closings-only sitting"},
	{"final_sitting_closings:red", "red closings filed in a closings-only sitting"},
	{"final_sitting_closings:blue", "blue closings filed in a closings-only sitting"},
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
	for _, s := range []recordpb.SidedWith{recordpb.SidedWith_SIDED_WITH_RED, recordpb.SidedWith_SIDED_WITH_BLUE} {
		answers[recordpb.Word(s)] = true
	}
	for _, m := range []recordpb.MovedGrades{recordpb.MovedGrades_MOVED_GRADES_KEPT, recordpb.MovedGrades_MOVED_GRADES_RESTORED} {
		answers[recordpb.Word(m)] = true
	}
	drives := map[string]bool{"estopped_refused:regrade": true, "estopped_refused:supersede": true,
		"closing_no_text": true, "closings_only_refused": true,
		"final_sitting_closings:red": true, "final_sitting_closings:blue": true}
	for _, g := range proposalGateKeys {
		parts := strings.Split(g.key, ":")
		switch {
		case drives[g.key]:
		case parts[0] == "docketing" && len(parts) == 2 && causes[parts[1]]:
		case parts[0] == "answer" && len(parts) == 3 && causes[parts[1]] && answers[parts[2]]:
		case parts[0] == "tombstone" && len(parts) == 3 && parts[1] == "placed" && answers[parts[2]]:
		default:
			t.Errorf("gate %q (%s) is no cell a counter writes", g.key, g.what)
		}
	}
}
