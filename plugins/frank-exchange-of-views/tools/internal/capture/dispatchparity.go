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
// SITTINGS AGAINST DISPATCH ROWS. Within the window from the first dispatch to termination, every
// sitting opened between one chair sitting's dispatches D and the next's D+1 — by the hook's bracket
// or by a register — is either a party named in D or the seat that authors D+1 (the chair opening
// its sitting to run the verb is exempt BY RULE, not by allowlist), and every party named in D SAT
// for it — record.DispatchGroup's Sat, the one "has this seat sat" predicate — before D+1. The
// bookends sit outside the window by position: the base phase before the first dispatch, and the
// bench's terminal and assembly sittings after the last chair sitting.
//
// THE RELAY AGAINST THE PLAN THE RECORD STORED. The verb stores every plan it prints, whole
// (record.StoredPlans), so the relay is held to the original by equality: each relayed plan, read
// as a plan, prints exactly as the plan the record stored LAST at the relayed plan's epoch — the one
// the chair asked for last in that sitting. A field dropped, added or altered in the relay, a stale
// plan relayed after a motion filed or ruled in the sitting, and a plan the record never stored all
// depart; so does a sitting whose stored plan the journal relays none of. A chair result in the
// journal is one whose `plan` is an object with a `parties` array — how the workflow itself
// recognises what to dispatch.
//
// With no journal there is nothing to compare, and the detail says the relayed plans were NOT
// compared: an absent journal never reads as a faithful relay. With no stored plan — a record
// migrated from before the record stored one — the detail says the same of the record.
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
	var relays []map[string]any
	if journalPresent {
		relays = relayedPlans(results)
	} else {
		notCompared = noJournalNote()
	}
	groups := record.DispatchGroups(fam.Events, fam.At)
	if len(groups) == 0 && len(relays) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "SKIP", Detail: "no dispatch on the record — a run before the chair dispatched, or one that never reached the debate" + notCompared}
	}
	// A SITTING IS THE STORED ONE (fam.At): the hook's bracket opens it where the configuration
	// seats one seat, a register where no bracket did. Each is read at its opening, so a seat that
	// sat by its bracket alone is a sitting here, and a register that joined its bracket's sitting
	// is not a second one.
	type sitting struct {
		pos  int
		seat string
		// occasion is WHAT THE SITTING WAS CONVENED TO DO, off the register in it that states one.
		// Empty for every seat whose id already says, for a bench sitting no register stated one
		// in, and on a record written before the field existed.
		occasion string
	}
	// How many of the bench's registers said what they were for, so a record that predates the
	// field is reported as NOT MEASURED rather than read as though every bench sitting were a
	// docket ruling.
	var benchRegs, benchRegsWithOccasion int
	occasionOf := map[int64]string{} // a sitting's opening id -> the occasion a register in it states
	for _, e := range fam.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_REGISTER {
			continue
		}
		occasion := ""
		if b, ok := recordpb.BodyAs[*recordpb.Register](e); ok {
			occasion = recordpb.Word(b.GetOccasion())
		}
		if record.SeatOwesOccasion(e.GetSeatId()) {
			benchRegs++
			if occasion != "" {
				benchRegsWithOccasion++
				occasionOf[fam.At.Of(e).SittingID] = occasion
			}
		}
	}
	var regs []sitting
	for i, e := range fam.Events {
		if seat, opens := fam.At.Opens(e); opens {
			regs = append(regs, sitting{pos: i, seat: seat, occasion: occasionOf[fam.At.Of(e).ID]})
		}
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
			// A BENCH SITTING NO REGISTER STATED AN OCCASION IN is not judged: whether the engine
			// or the chair convened it is the register's to say, and it wrote none. It answers no
			// dispatch either — the absent check below wants a sitting of the occasion.
			if occasionsMeasured && record.SeatOwesOccasion(r.seat) && r.occasion == "" {
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
				strays = append(strays, fmt.Sprintf("%s sat for %s after dispatch %d, whose row convened it for %s", r.seat, r.occasion, k+1, setOf(g.PartyRows[r.seat].Occasions)))
				continue
			}
			strays = append(strays, fmt.Sprintf("%s sat after dispatch %d and was not a party to it", r.seat, k+1))
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
		stored, err := record.StoredPlans(fam.Events)
		if err != nil {
			return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
		}
		departures, note := relayDepartures(relays, stored)
		findings = append(findings, departures...)
		compared = note
	}
	if len(findings) == 0 {
		return Audit{Check: "dispatch-parity", Verdict: "PASS", Detail: fmt.Sprintf("%d dispatch(es); every sitting opened between them was a named party's or the chair's, and every party sat", len(groups)) + compared + notCompared}
	}
	return Audit{Check: "dispatch-parity", Verdict: "FAIL", Detail: strings.Join(findings, "; ") + notCompared}
}

// noJournalNote is what the detail carries when there is no journal to compare relayed plans from.
func noJournalNote() string {
	return " — no workflow journal, so the relayed plans were NOT compared with the plans the record stored"
}

// relayedPlans is every chair result's plan in the journal, in journal order.
func relayedPlans(results []map[string]any) []map[string]any {
	var out []map[string]any
	for _, r := range results {
		plan, ok := r["plan"].(map[string]any)
		if !ok {
			continue
		}
		if _, ok := plan["parties"].([]any); ok {
			out = append(out, plan)
		}
	}
	return out
}

// relayDepartures holds each relayed plan to the plan the record stored last at the relayed plan's
// epoch, by equality of their printings (record.ReadPlan), and each epoch the record stored a plan
// at to a relay of it. note is what the detail adds when the comparison ran.
func relayDepartures(relays []map[string]any, stored []record.StoredPlan) (out []string, note string) {
	if len(stored) == 0 {
		if len(relays) == 0 {
			return nil, ""
		}
		return nil, fmt.Sprintf(" — the record stores no plan, so the %d relayed plan(s) were NOT compared", len(relays))
	}
	last := map[int]record.StoredPlan{}
	var epochs []int
	for _, sp := range stored {
		if _, seen := last[sp.Plan.Epoch]; !seen {
			epochs = append(epochs, sp.Plan.Epoch)
		}
		last[sp.Plan.Epoch] = sp
	}
	relayed := map[int]bool{}
	for k, plan := range relays {
		raw, err := json.Marshal(plan)
		if err != nil {
			out = append(out, fmt.Sprintf("relay %d: the relayed plan does not encode: %v", k+1, err))
			continue
		}
		p, printed, err := record.ReadPlan(raw)
		if err != nil {
			out = append(out, fmt.Sprintf("relay %d: the relayed plan does not read as a plan: %v", k+1, err))
			continue
		}
		relayed[p.Epoch] = true
		sp, ok := last[p.Epoch]
		if !ok {
			out = append(out, fmt.Sprintf("relay %d: relays a plan at epoch %d, and the record stored none there", k+1, p.Epoch))
			continue
		}
		if printed != sp.Printed {
			out = append(out, fmt.Sprintf("epoch %d: the relayed plan departs from the plan the record stored last in that sitting on %s", p.Epoch, strings.Join(differing(printed, sp.Printed), ", ")))
		}
	}
	for _, e := range epochs {
		if !relayed[e] {
			out = append(out, fmt.Sprintf("epoch %d: the record stored a plan and the journal relays none", e))
		}
	}
	return out, fmt.Sprintf("; %d relayed plan(s) each equal to the plan the record stored last in its sitting", len(relays))
}

// differing is the keys of two printed plans whose values differ, in the plan's own order.
func differing(a, b string) []string {
	var x, y map[string]json.RawMessage
	_ = json.Unmarshal([]byte(a), &x) // both are a plan's printing
	_ = json.Unmarshal([]byte(b), &y)
	var out []string
	for _, f := range record.PlanFields() {
		if string(x[f.JSON]) != string(y[f.JSON]) {
			out = append(out, f.JSON)
		}
	}
	return out
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
