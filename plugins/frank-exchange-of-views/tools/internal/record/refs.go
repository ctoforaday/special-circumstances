package record

import (
	"fmt"
	"sort"
	"strings"
)

// EVERY REFERENCE IS CHECKED AGAINST THE THING IT REFERENCES.
//
// This file exists because an audit of the whole surface found TWELVE OF TWELVE
// cross-references unvalidated. Only `close --id` and `mint --supersedes` were checked;
// every other field naming another entity — a gap, an observation, a finding, a seat —
// was accepted on the seat's say-so.
//
// That is not a hypothetical. It is the mechanism behind the 2026-07-18 run's worst
// damage: eight judicial closures were `opinion --id` events ACCEPTED AT WRITE TIME and
// then silently DROPPED at replay, because replay checks what the write path did not. The
// board was wrong by six gaps for three rounds and nothing on the surface said so.
//
// The split is the defect. A reference that replay will refuse must be refused when it is
// written, while the seat is still there to fix it — an event accepted into the log and
// discarded on the way out is the worst of both: it looks recorded and it does nothing.
//
// COST, MEASURED BEFORE ENFORCING: every reference in the run resolves — 6 successors,
// 20 fold-into targets, all minted; and zero forward references within a shard, so no
// seat named something it had not yet created. This would have refused nothing.

// requireGap refuses a reference to a gap no mint created.
func requireGap(runDir, id, verb, flag string) error {
	if id == "" {
		return nil
	}
	ids, err := allGapIDs(runDir)
	if err != nil {
		return err
	}
	if !ids[id] {
		return fmt.Errorf("record: %s %s names gap %s, which no mint event created — a dangling reference is accepted here and DROPPED at replay, which is how eight judicial closures vanished from a board that went on reporting them open", verb, flag, id)
	}
	return nil
}

// requireGaps is the list form, naming the first that does not resolve.
func requireGaps(runDir string, ids []string, verb, flag string) error {
	for _, id := range ids {
		if err := requireGap(runDir, id, verb, flag); err != nil {
			return err
		}
	}
	return nil
}

// requireObservation resolves a disposal's target, ROUND-SCOPED, and refuses ambiguity.
//
// The historical problem (2026-07-18 run): 15 labels were used by more than one lens seat
// (L5-F1 existed in rounds 1, 2 and 3), so 39 of 60 disposals named something that matched
// several findings; 13 named a label no event ever created; and 8 finding events carried no
// label at all. That is fixed at the SOURCE now: finding labels are tool-assigned, run-unique
// per role (findinglabel.go), so a label names exactly one finding run-wide — cross-round
// collision is impossible, not merely disambiguated.
//
// Round scoping therefore is no longer the disambiguator; it is simply the natural scope of
// a disposal — a merge coalesces THIS round's lens findings, so red-merge-r3 disposing L5-F3
// resolves it among round 3's findings. It stays as belt-and-suspenders: with unique labels
// it can only agree with a global lookup, and it keeps the refusal-on-ambiguity path honest.
//
// Where it still cannot tell, it REFUSES AND NAMES THE CANDIDATES rather than picking. A
// silent wrong pick writes the wrong finding's fate into the record and leaves the right
// one looking undisposed forever — invisible, and exactly the class of defect the whole
// tool exists to remove.
func requireObservation(runDir, label, seatID, verb, flag string) error {
	if label == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	round := RoundOf(seatID)
	var sameRound, anyRound []string
	for _, e := range m.Events {
		if e.Type != "observe" && e.Type != "finding" {
			continue
		}
		if e.Payload.Str("label") != label {
			continue
		}
		anyRound = append(anyRound, e.SeatID)
		if RoundOf(e.SeatID) == round {
			sameRound = append(sameRound, e.SeatID)
		}
	}
	switch {
	case len(sameRound) == 1:
		return nil
	case len(sameRound) > 1:
		return fmt.Errorf("record: %s %s names %s, which %d seats recorded THIS ROUND (%v) — the label does not identify one finding, so the disposal would write the wrong one's fate and leave the right one looking undisposed", verb, flag, label, len(sameRound), sameRound)
	case len(anyRound) == 1:
		return nil
	case len(anyRound) > 1:
		return fmt.Errorf("record: %s %s names %s, recorded by %v in other rounds and by nobody this round — name the finding from your own round, or the disposal is guesswork", verb, flag, label, anyRound)
	}
	return fmt.Errorf("record: %s %s names %s, which no observe or finding event created — the disposal would name nothing and the real observation would stay undisposed. If the lens wrote it in prose but never recorded it, the finding is the thing that is missing", verb, flag, label)
}

// requireFindings refuses found_by attribution to a finding that does not exist.
//
// found_by is the credit chain from a lens's work to the board gap it earned, and it is
// read back by the capture-recapture estimate. An invented label inflates the count of
// lens-sourced gaps with a finding nobody made.
func requireFindings(runDir string, labels []string, verb, flag string) error {
	if len(labels) == 0 {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, e := range m.Events {
		if e.Type == "finding" || e.Type == "observe" {
			if l := e.Payload.Str("label"); l != "" {
				have[l] = true
			}
		}
	}
	for _, l := range labels {
		if !have[l] {
			return fmt.Errorf("record: %s %s names finding %s, which no lens recorded — attribution to a finding nobody made inflates the lens-sourced count the capture estimate reads", verb, flag, l)
		}
	}
	return nil
}

// requireSeat refuses attribution to a seat that never registered.
func requireSeat(runDir, seatID, verb, flag string) error {
	if seatID == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.SeatID == seatID {
			return nil
		}
	}
	return fmt.Errorf("record: %s %s names seat %s, which has recorded nothing in this run — a ruling attributed to a seat that never sat cannot be matched to the petition it answers", verb, flag, seatID)
}

// ---- STATE, not just existence ----
//
// A reference can resolve and still be wrong: the thing exists, but it is not in a state
// where the act makes sense. Regrading a closed gap moves a number nobody will read;
// answering a dispute nobody filed records half an argument; re-closing a closed gap
// double-counts closure history and, per replay.go, corrupts the repair_regression
// denominator that the whole repair metric divides by.
//
// EVERY RULE BELOW WAS MEASURED AGAINST THE 2026-07-18 RUN FIRST, and one candidate was
// DROPPED because of it: "supersedes must name a closed gap" sounded right and would have
// refused NINE OF NINE real mints — superseding a still-open predecessor is the normal
// pattern, not an error. The rules that survived had zero violations in the run, so they
// cost nothing today and exist for the cheaper tier.

// gapState answers what a gap IS, not merely whether it exists.
func gapState(runDir, id string) (closed bool, err error) {
	b, err := BoardState(runDir)
	if err != nil {
		return false, err
	}
	g, ok := b.Gaps[id]
	if !ok {
		return false, nil
	}
	return !g.Open, nil
}

// requireOpenGap refuses an act that only makes sense on a live gap.
func requireOpenGap(runDir, id, verb, flag, why string) error {
	if id == "" {
		return nil
	}
	closed, err := gapState(runDir, id)
	if err != nil {
		return err
	}
	if closed {
		return fmt.Errorf("record: %s %s names gap %s, which is already CLOSED — %s", verb, flag, id, why)
	}
	return nil
}

// requireClosedGap is the mirror: spot-check re-verifies the ARCHIVE, so naming a gap that
// was never closed is not a sample of anything.
func requireClosedGaps(runDir string, ids []string, verb, flag string) error {
	for _, id := range ids {
		if id == "" {
			continue
		}
		closed, err := gapState(runDir, id)
		if err != nil {
			return err
		}
		if !closed {
			return fmt.Errorf("record: %s %s names gap %s, which is still OPEN — the spot-check samples ARCHIVED closures, and re-verifying something that was never closed discharges the duty without doing it", verb, flag, id)
		}
	}
	return nil
}

// requirePriorDispute refuses an answer to an argument nobody made.
//
// The dispute channel fired ZERO times across two full runs, which is a finding about the
// channel rather than evidence that nobody disagreed. When it does fire, a response with no
// dispute behind it would record one half of an exchange and make the accounting — disputes
// raised against disputes answered — silently wrong in the flattering direction.
func requirePriorDispute(runDir, gapID string) error {
	if gapID == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.Type == "dispute" && e.Payload.Str("gap_id") == gapID {
			return nil
		}
	}
	return fmt.Errorf("record: dispute-respond --id names gap %s, on which no dispute was filed — answering an argument nobody made records half an exchange and inflates the answered-disputes count against a denominator of zero", gapID)
}

// requireUndisposed refuses giving one finding a SECOND fate.
//
// This was unmeasurable before ids existed: the run showed 16 observations "disposed more
// than once", but every one was a label collision — three rounds of lens 5 each recording
// an L5-F1, and three disposals landing on whichever replayed first. That collision can no
// longer happen (finding labels are tool-assigned run-unique per role), and identity is
// assigned by the tool, so a second disposal of the same finding is unambiguous, and it is
// wrong: the first fate is already on the record and in the counts, and a second silently
// overwrites which one the projection shows.
//
// A seat that has changed its mind about a fate is describing a REGRADE of a judgement, and
// there is no verb for that yet — which is a finding about the tooling, and the friction
// verb is how it gets one.
func requireUndisposed(runDir, target string) error {
	if target == "" {
		return nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	for _, e := range m.Events {
		if e.Type == "dispose" && e.Payload.Str("observation") == target {
			return fmt.Errorf("record: dispose --observation names %s, which %s already gave the fate %q — a finding has one fate, and a second silently replaces which one the projection shows. If the first was wrong, record it with the friction verb: there is no verb for revising a disposal, and that absence is itself the finding",
				target, e.SeatID, e.Payload.Str("disposition"))
		}
	}
	return nil
}

// requireSupersededAreClosed is a COMPLETION duty, checked at the seat's terminal act.
//
// INVESTIGATED RATHER THAN ASSUMED. "Superseding an open gap" happened 9 times in the
// 2026-07-18 run and I first read the frequency as proof it was intended. Tracing what each
// one actually did splits them in two:
//
//	7 are STRUCTURALLY REQUIRED. The protocol is mint-the-successor, then close the
//	  ancestor naming it — so the ancestor is necessarily still open at mint time, and it
//	  cannot be otherwise, because the closure has to name a successor that already exists.
//	2 are a DEFECT nobody caught. R3-1 superseded R2-1 and R2-5 and neither was ever
//	  closed, so all three finished OPEN on the board. The run reported 9 open gaps; 7
//	  were distinct defects and one was counted three times.
//
// That is why the rule is not "supersedes must name a closed gap" — that would refuse all
// 9, including the 7 the protocol demands. The duty is that a superseded ancestor must not
// still be open when the seat FINISHES: superseding is a promise to replace, and a promise
// kept open inflates every count the board reports.
//
// Checked at verdict because that is the seat's terminal act and the last moment it is
// still there to close them.
func requireSupersededAreClosed(runDir string) error {
	b, err := BoardState(runDir)
	if err != nil {
		return err
	}
	superseded := map[string]string{} // ancestor -> the gap that claimed to replace it
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || g.Mint == nil {
			continue
		}
		for _, anc := range g.Mint.StrList("supersedes") {
			superseded[anc] = id
		}
	}
	var stranded []string
	for anc, successor := range superseded {
		if g, ok := b.Gaps[anc]; ok && g.Open {
			stranded = append(stranded, fmt.Sprintf("%s (superseded by %s)", anc, successor))
		}
	}
	if len(stranded) == 0 {
		return nil
	}
	sort.Strings(stranded)
	return fmt.Errorf("record: verdict refused — %d superseded gap(s) are still OPEN: %s. Superseding is a promise to replace, and an ancestor left open is the same defect counted twice: the 2026-07-18 run finished reporting 9 open gaps of which 7 were distinct. Close each ancestor (--successor names its replacement), or if it is genuinely still live, it was not superseded",
		len(stranded), strings.Join(stranded, ", "))
}

// requirePassClosesAllGaps refuses a PASS while ANY gap is still open. The protocol is "PASS
// only when every remaining gap is closed, evidence-rebutted, or risk-accepted", and all of
// those resolutions go through `close` (which sets the gap not-open) — so an open gap at PASS
// is an unadjudicated one. requireSupersededAreClosed catches only the lineage subset; the
// 2026-07-20 run recorded PASS with 9 PLAIN open gaps (one HIGH) that no lineage check saw,
// and the envelope then reported 0 outstanding. This is the complete enforcement, at the
// write path so no verdict route can bypass it. A FAIL is always allowed.
func requirePassClosesAllGaps(runDir string) error {
	b, err := BoardState(runDir)
	if err != nil {
		return err
	}
	var open []string
	for _, id := range b.GapOrder {
		if g := b.Gaps[id]; g != nil && g.Open {
			open = append(open, id)
		}
	}
	if len(open) == 0 {
		return nil
	}
	sort.Strings(open)
	return fmt.Errorf("record: verdict PASS refused — %d gap(s) still OPEN: %s. PASS requires every gap resolved through `close --id <id> --as closed|risk_accepted|rebuttal_sustained|routed_to_infrastructure`; close them, or issue `--as FAIL`",
		len(open), strings.Join(open, ", "))
}

// gapNamedIn returns the first gap id from the board that appears as a WHOLE TOKEN in
// prose, or "". It exists for one refusal: `blue edit` prose that names the gap the edit
// answers while `--answers` is empty (validate, case "blue_edit").
//
// It matches against the BOARD, never against a pattern. A regex for "gap-id-shaped" would
// have to guess the shape (R1-5, R1-11 today) and would fire on any prose that happens to
// look like one — a version number, a section reference, a matrix cell. Membership in the
// set of ids some mint actually created is exact, needs no shape at all, and costs the same
// read requireGap already does.
//
// Tokens break on anything outside [A-Za-z0-9-], so "R1-5: Quantify…" yields "R1-5" with
// its trailing colon shed, and a longer id is never matched by a shorter one's prefix.
func gapNamedIn(runDir, prose string) (string, error) {
	if strings.TrimSpace(prose) == "" {
		return "", nil
	}
	ids, err := allGapIDs(runDir)
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", nil
	}
	for _, tok := range strings.FieldsFunc(prose, func(r rune) bool {
		return !(r == '-' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) {
		if ids[tok] {
			return tok, nil
		}
	}
	return "", nil
}
