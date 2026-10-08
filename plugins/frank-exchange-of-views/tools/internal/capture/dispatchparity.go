package capture

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// DispatchParityAudit holds the workflow's RELAY to the record (plans/roundless.md §III.B.1). The
// chair's verb records who sits; the chair relays that into its envelope; the workflow dispatches
// what the envelope says — and the chair could drop, add or alter a party on the way. Two halves:
//
// REGISTERS AGAINST DISPATCH ROWS. Within the window from the first dispatch to termination, every
// register between one chair sitting's dispatches D and the next's D+1 is either a party named in D
// or the seat that authors D+1 (the chair registering to run the verb is exempt BY RULE, not by
// allowlist), and every party named in D SAT for it — record.DispatchGroup's Sat, the one "has this
// seat sat" predicate — before D+1. The bookends register outside the window by position: the base
// phase before the first dispatch, and the bench's terminal and assembly sittings after the last
// chair sitting.
//
// RELAYED FIELDS AGAINST DISPATCH ROWS. A chair result in the journal is the one whose `plan` is an
// object with a `parties` array — the chair envelope is the only one carrying a plan, and it is how
// the workflow itself recognises what to dispatch. Each relayed plan with non-empty parties pairs,
// in journal order, with a dispatch group, in stream order: the verb records a row per party and none
// for an empty plan, so the two sequences are one per chair sitting that dispatched. A count mismatch
// FAILs and nothing is compared, because the pairing is then unknown. Per pair, every relayed field
// with a row counterpart is compared: the seat_id set against the group's parties, each party's
// gap_ids and occasions (as sets) against its PartyRows row, and the plan's head against that row's
// pin. The row is the LAST one naming the party in the group — a plan that files a docket, or one that
// changed since the chair asked, is never standing, so a chair that asks twice can write two rows,
// and the later is the dispatch. Without this half a wrong gap
// id or head reached a seat and every register still matched, so the relay read as faithful.
//
// THE LAST PLAN'S BENCH BLOCKERS AGAINST THE RECORD AT THE END OF THE CHAIR'S LAST SITTING. The
// workflow convenes the terminal bench sitting off the last relayed plan's blockers the bench owns,
// and no row records a blocker. So the relayed set is held to the record's own unruled bench-gavel
// motions as they stood when the chair's last sitting ended — its last act, before the next seat's
// opening. A blocker dropped or altered in the relay, or a motion filed after the plan was asked
// for, departs here: the plan the chair relays is the one it asked for last.
//
// THE JOURNAL HOLDS ONE CHAIR RESULT PER CHAIR SITTING — the stated assumption the pairing rests on.
// A resume that re-journals a cached chair result shifts the pairing by one and FAILs on the count,
// loudly. With no journal there is nothing to compare, and the detail says the relayed fields were
// NOT compared: an absent journal never reads as a faithful relay.
//
// A CHAIR SITTING'S DISPATCH IS record.DispatchGroups', NOT THE EPOCH. A warm chair resuming its
// session opens no stored sitting until it registers, and on B3–B6 it registered once per run:
// grouped by epoch, each run is one chair sitting whose last dispatch is the run's last, every
// party sitting before it is discarded, and every party of "dispatch 1" reads as never having sat —
// the same FAIL on every warm run, whether or not the parties sat.
//
// The verb records the truth; this catches a relay that departed from it. A mismatch is a FAIL.
func DispatchParityAudit(run record.Run, results []map[string]any, journalPresent bool) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
	}
	notCompared := ""
	var relays []relayedPlan
	if journalPresent {
		relays = relayedPlans(results)
	} else {
		notCompared = " — no workflow journal, so the relayed party fields (seat_id, gap_ids, occasions, head) and the last plan's bench blockers were NOT compared"
	}
	groups := record.DispatchGroups(fam.Events, fam.At)
	if len(groups) == 0 && len(relays) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "SKIP", Detail: "no dispatch on the record — a run before the chair dispatched, or one that never reached the debate" + notCompared}
	}
	type reg struct {
		pos  int
		seat string
		// occasion is WHAT THE SITTING WAS CONVENED TO DO, off the register. Empty for every seat
		// whose id already says, and empty on a record written before the field existed.
		occasion string
	}
	var regs []reg
	// How many of the bench's registers said what they were for, so a record that predates the
	// field is reported as NOT MEASURED rather than read as though every bench sitting were a
	// docket ruling.
	var benchRegs, benchRegsWithOccasion int
	for i, e := range fam.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		r := reg{pos: i, seat: e.GetSeatId()}
		if b, ok := recordpb.BodyAs[*recordpb.Register](e); ok {
			r.occasion = recordpb.Word(b.GetOccasion())
		}
		if record.SeatOwesOccasion(r.seat) {
			benchRegs++
			if r.occasion != "" {
				benchRegsWithOccasion++
			}
		}
		regs = append(regs, r)
	}
	// THE OCCASION IS EITHER THERE FOR THE WHOLE BENCH OR IT IS NOT MEASURED. A record written
	// before the field carries none; a half-carrying one is worse than either, because it would
	// let some bench sittings be checked and others silently skipped under one verdict.
	occasionsMeasured := benchRegs > 0 && benchRegs == benchRegsWithOccasion
	// THE BOOKENDS ARE THE BENCH, and since the bench collapsed to one seat they are no longer
	// separable from it by id. This exempts a bench register in the FINAL dispatch group only,
	// which is where the terminal disposition and the assembly sit. The discrimination lost is
	// narrow and real: a genuinely stray bench register in that last group now reads as a bookend.
	// It cannot be recovered from an id that four sittings share — position is what distinguishes
	// them, and position is already what this test uses.
	// A BENCH SITTING THE ENGINE CONVENED IS NOT A STRAY. The terminal disposition and the assembly
	// are convened by the engine, not by the chair, so each lands in a dispatch window with no
	// dispatch row naming it. `docket` and `petition` are the occasions the chair dispatches — the
	// bench's row names which — so a bench register of either is a party's sitting, answered by a
	// row like any other, and one no row convened is a stray.
	//
	// THIS USED TO BE POSITION — "a bench register in the FINAL group" — which let a genuinely
	// stray bench register in that group read as a bookend. That was the discrimination the seat-id
	// collapse lost, and it is recovered here: the exemption is now exact rather than a window.
	engineConvened := func(seat, occasion string) bool {
		return record.SeatOwesOccasion(seat) && (occasion == "terminal" || occasion == "assemble")
	}
	var strays, absent, unmeasured []string
	for k, g := range groups {
		end := len(fam.Events)
		if k+1 < len(groups) {
			end = groups[k+1].First
		}
		party := map[string]bool{}
		for _, p := range g.Parties {
			party[p] = true
		}
		for _, r := range regs {
			if r.pos <= g.Last || r.pos >= end {
				continue
			}
			if r.seat == "red-chair" || engineConvened(r.seat, r.occasion) {
				continue
			}
			// The bench is a party for the occasions its row convened it for, and only those: a
			// petition sitting under a row that convened a docket answers nothing the chair asked.
			if party[r.seat] && (!occasionsMeasured || !record.SeatOwesOccasion(r.seat) || slices.Contains(g.PartyRows[r.seat].Occasions, r.occasion)) {
				continue
			}
			// An unmeasured bench register keeps the old positional exemption: on a record
			// carrying no occasions there is nothing better, and guessing would invent a stray.
			if !occasionsMeasured && k+1 == len(groups) && record.SeatOwesOccasion(r.seat) {
				continue
			}
			if party[r.seat] {
				strays = append(strays, fmt.Sprintf("%s registered for %s after dispatch %d, whose row convened it for %s", r.seat, r.occasion, k+1, setOf(g.PartyRows[r.seat].Occasions)))
				continue
			}
			strays = append(strays, fmt.Sprintf("%s registered after dispatch %d and was not a party to it", r.seat, k+1))
		}
		for _, p := range g.Parties {
			// A BENCH DISPATCHED FOR AN OCCASION MUST HAVE SAT FOR *THAT*, and the occasion is
			// what makes the question answerable.
			//
			// `Sat` is the party's first sitting opened after the dispatch. In the FINAL group the
			// bench's closing sittings — the terminal disposition and the assembly — register in
			// that same window, and once the bench collapsed to one seat they were no longer
			// separable from a docket ruling by id. So a bench dispatched onto a gap and never
			// sitting was SATISFIED BY ITS OWN BOOKEND, and "the bench sat" became unfalsifiable
			// on any run whose last dispatching chair sitting engaged it. The register that
			// answers a chair's dispatch is the one whose occasion the row convened the bench for
			// — EACH of them: a row convening it for a petition and a docket is sat twice, and a
			// petition-only row (no gaps) is answered by a petition sitting, not a docket one.
			if record.SeatOwesOccasion(p) {
				if !occasionsMeasured {
					unmeasured = append(unmeasured, fmt.Sprintf("%s was named in dispatch %d and its sitting is NOT MEASURED — this record predates the register's occasion, so the bench's closing sittings cannot be told from a docket ruling", p, k+1))
					continue
				}
				convened := g.PartyRows[p].Occasions
				if len(convened) == 0 {
					absent = append(absent, fmt.Sprintf("%s was named in dispatch %d by a row that convenes it for no occasion — nothing says which sitting answers it", p, k+1))
					continue
				}
				for _, occ := range convened {
					sat := false
					for _, r := range regs {
						if r.seat == p && r.occasion == occ && r.pos > g.Last && r.pos < end {
							sat = true
							break
						}
					}
					if !sat {
						absent = append(absent, fmt.Sprintf("%s was named in dispatch %d for %s and recorded no %s sitting before the next — its closing sittings do not answer a dispatch", p, k+1, occ, occ))
					}
				}
				continue
			}
			if at, sat := g.Sat[p]; !sat || at >= end {
				absent = append(absent, fmt.Sprintf("%s was named in dispatch %d and sat no sitting before the next", p, k+1))
			}
		}
	}
	sort.Strings(strays)
	sort.Strings(absent)
	sort.Strings(unmeasured)
	findings := append(strays, absent...)
	findings = append(findings, unmeasured...)
	compared := ""
	if journalPresent {
		if len(relays) != len(groups) {
			findings = append(findings, fmt.Sprintf("the journal relays %d plan(s) with parties and the record holds %d chair sitting dispatch(es) — the pairing is unknown, so no relayed party field was compared", len(relays), len(groups)))
		} else {
			for k := range groups {
				findings = append(findings, relayDepartures(k+1, relays[k], groups[k])...)
			}
			compared = fmt.Sprintf("; %d relayed plan(s) matched their dispatch rows on parties, gap_ids, occasions and head", len(relays))
		}
		if last, ok := lastRelayedPlan(results); ok {
			if d := benchBlockerDepartures(last, fam.Events, fam.At); d != "" {
				findings = append(findings, d)
			} else {
				compared += "; the last plan's bench blockers matched the record at the end of the chair's last sitting"
			}
		}
	}
	if len(findings) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "PASS", Detail: fmt.Sprintf("%d dispatch(es); every register between them was a named party or the chair, and every party sat", len(groups)) + compared + notCompared}
	}
	return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: strings.Join(findings, "; ") + notCompared}
}

// relayedPlan is one chair result's plan as the journal holds it: the head as relayed (raw, so a
// value that is not an integer is reported as it stands) and each party in relay order.
type relayedPlan struct {
	head    any
	parties []relayedParty
}

type relayedParty struct {
	seat      string
	gaps      []string
	occasions []string
}

// relayedPlans is every chair result in the journal whose plan names at least one party, in journal
// order. An empty plan dispatches nobody and writes no row, so it has no group to pair with.
func relayedPlans(results []map[string]any) []relayedPlan {
	var out []relayedPlan
	for _, r := range results {
		plan, ok := r["plan"].(map[string]any)
		if !ok {
			continue
		}
		ps, ok := plan["parties"].([]any)
		if !ok || len(ps) == 0 {
			continue
		}
		rp := relayedPlan{head: plan["head"]}
		for _, p := range ps {
			pm, _ := p.(map[string]any)
			party := relayedParty{seat: jsString(pm["seat_id"])}
			if gs, ok := pm["gap_ids"].([]any); ok {
				for _, g := range gs {
					party.gaps = append(party.gaps, jsString(g))
				}
			}
			if os, ok := pm["occasions"].([]any); ok {
				for _, o := range os {
					party.occasions = append(party.occasions, jsString(o))
				}
			}
			rp.parties = append(rp.parties, party)
		}
		out = append(out, rp)
	}
	return out
}

// relayDepartures compares one relayed plan against the dispatch group it pairs with, naming the
// sitting (the 1-based pair index), the seat, and the relayed and recorded values of each field that
// departs.
func relayDepartures(sitting int, rp relayedPlan, g record.DispatchGroup) []string {
	var out []string
	relayed := map[string][]string{}
	relayedOcc := map[string][]string{}
	var relayedSeats []string
	for _, p := range rp.parties {
		if _, dup := relayed[p.seat]; dup {
			out = append(out, fmt.Sprintf("sitting %d: %s is relayed twice — the dispatch names each party once", sitting, p.seat))
			continue
		}
		relayed[p.seat] = p.gaps
		relayedOcc[p.seat] = p.occasions
		relayedSeats = append(relayedSeats, p.seat)
	}
	if !sameSet(relayedSeats, g.Parties) {
		out = append(out, fmt.Sprintf("sitting %d: relayed parties %s, the dispatch rows name %s", sitting, setOf(relayedSeats), setOf(g.Parties)))
	}
	head, headOK := relayedHead(rp.head)
	for _, seat := range g.Parties {
		gaps, ok := relayed[seat]
		if !ok {
			continue // named by the seat-set departure above
		}
		row := g.PartyRows[seat]
		if !sameSet(gaps, row.GapIDs) {
			out = append(out, fmt.Sprintf("sitting %d: %s relayed gap_ids %s, its dispatch row recorded %s", sitting, seat, setOf(gaps), setOf(row.GapIDs)))
		}
		if !sameSet(relayedOcc[seat], row.Occasions) {
			out = append(out, fmt.Sprintf("sitting %d: %s relayed occasions %s, its dispatch row recorded %s", sitting, seat, setOf(relayedOcc[seat]), setOf(row.Occasions)))
		}
		if !headOK || head != row.Pin {
			out = append(out, fmt.Sprintf("sitting %d: %s relayed head %s, its dispatch row pinned %d", sitting, seat, jsString(rp.head), row.Pin))
		}
	}
	return out
}

// lastRelayedPlan is the plan of the journal's last chair result — the one the workflow read its
// exit from, parties or none — and whether the journal holds one.
func lastRelayedPlan(results []map[string]any) (map[string]any, bool) {
	for i := len(results) - 1; i >= 0; i-- {
		if plan, ok := results[i]["plan"].(map[string]any); ok {
			if _, ok := plan["parties"].([]any); ok {
				return plan, true
			}
		}
	}
	return nil, false
}

// benchBlockerDepartures holds the last relayed plan's bench-owned motion blockers to the record's
// unruled bench-gavel motions at the end of the chair's last sitting: the first sitting another seat
// opens after the chair's last act, or the end of the record. "" when they agree.
func benchBlockerDepartures(plan map[string]any, evs []*record.Event, win record.WindowIndex) string {
	last := -1
	for i, e := range evs {
		if e.GetSeatId() == "red-chair" {
			last = i
		}
	}
	if last < 0 {
		return ""
	}
	end := len(evs)
	for i := last + 1; i < len(evs); i++ {
		if seat, opens := win.Opens(evs[i]); opens && seat != "red-chair" {
			end = i
			break
		}
	}
	var onRecord, relayed []string
	for _, b := range record.MotionBlockersOf(evs[:end], win) {
		if record.SeatOwesOccasion(b.Owner) {
			onRecord = append(onRecord, b.Subject)
		}
	}
	bs, _ := plan["blockers"].([]any)
	for _, x := range bs {
		b, _ := x.(map[string]any)
		if jsString(b["kind"]) == string(record.BlockerUnruledMotion) && record.SeatOwesOccasion(jsString(b["owner"])) {
			relayed = append(relayed, jsString(b["subject"]))
		}
	}
	if sameSet(relayed, onRecord) {
		return ""
	}
	return fmt.Sprintf("the last plan relays bench-owned motions %s, and the record at the end of the chair's last sitting holds %s unruled — the terminal bench sitting is convened off the relayed set, so the plan relayed is the one the chair asks for last", setOf(relayed), setOf(onRecord))
}

// relayedHead reads the relayed head as the integer the plan prints; ok is false for anything else.
func relayedHead(v any) (int64, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	h, err := n.Int64()
	return h, err == nil
}

// setOf renders xs as a sorted, de-duplicated list: the form both sides of a set comparison are
// named in.
func setOf(xs []string) string {
	seen := map[string]bool{}
	var s []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			s = append(s, x)
		}
	}
	sort.Strings(s)
	return "[" + strings.Join(s, ", ") + "]"
}

func sameSet(a, b []string) bool { return setOf(a) == setOf(b) }
