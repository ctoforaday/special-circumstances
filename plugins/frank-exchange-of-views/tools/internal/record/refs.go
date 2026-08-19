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
// THE EXPORTED FORMS, for flags that carry their own existence check.
//
// A typed flag declares what it is checked against (`flags.GapID().WithCheck(record.GapExists)`),
// and seat.Begin runs it once the run directory is resolved. These are THIN WRAPPERS over the
// same helpers `validate` uses — one implementation, two call sites — because validate remains
// the enforcer: it is the single write path every caller goes through, and a rule the CLI held
// alone would be one every other caller skips. Two enforcers drifting is the defect this
// codebase keeps finding; two callers of one enforcer is not.
//
// The subject is named as the FLAG, not the verb: this refusal arrives while the seat is looking
// at one command, so "--id names gap R9-9, which no mint event created" is the whole sentence it
// needs. validate's copy keeps the verb too, because there it can be reached by any caller.
func GapExists(runDir, id string) error { return requireGap(runDir, id, "the", "--id") }

// InquiryExists resolves a inquiry.
func InquiryExists(runDir, id string) error { return requireInquiry(runDir, id, "the", "--id") }

// CitationExists resolves a citation anchor.
func CitationExists(runDir, label string) error {
	return requireCitation(runDir, label, "the", "--cites")
}

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
		if e.Type == "finding" {
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
// requireCitation refuses a citation label that names no `blue cite` on the record.
//
// `blue prove --cites` names the METHOD a computation applies — "the source that says trial
// division decides primality". It was set into the payload and never checked, so a proof could
// cite a citation that does not exist and the assembled report would carry the link as though it
// meant something. That is the same hole `lens verify --anchor` had until 0.60.0, on the same
// axis, one verb over.
func requireCitation(runDir, label, verb, flag string) error {
	if label == "" {
		return nil
	}
	known, err := CitationLabels(runDir)
	if err != nil {
		return err
	}
	for _, k := range known {
		if k == label {
			return nil
		}
	}
	return fmt.Errorf("record: %s %s=%s names no citation on the record — blue has cited %d source(s), and `show evidence` lists them by anchor. Cite the method with `blue cite` first; a proof pointing at a citation that does not exist claims a provenance it does not have",
		verb, flag, label, len(known))
}

// requireInquiry refuses a move against a line of inquiry nobody proposed.
//
// `blue line-of-inquiry --id` required only that an id be PRESENT. A move naming an unknown line of inquiry wrote
// a status change for a line of inquiry that was never opened — and the lines-of-inquiry view
// renders it, so the run shows a direction being abandoned that nothing ever proposed.
func requireInquiry(runDir, id, verb, flag string) error {
	if id == "" {
		return nil
	}
	b, err := BoardState(runDir)
	if err != nil {
		return err
	}
	for _, a := range Inquiries(b) {
		if a != nil && a.ID == id {
			return nil
		}
	}
	return fmt.Errorf("record: %s %s=%s names no line of inquiry on the record — `show lines-of-inquiry` lists every one with its id and fate. Propose it first (`blue line of inquiry --line …`, which ASSIGNS the id); --id moves a line of inquiry that already exists",
		verb, flag, id)
}

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
	return fmt.Errorf("record: verdict refused — %d superseded gap(s) are still OPEN: %s. Superseding is a promise to replace, and an ancestor left open is the same defect counted twice: the 2026-07-18 run finished reporting 9 open gaps of which 7 were distinct. Close each ancestor (--superseded-by names its replacement), or if it is genuinely still live, it was not superseded",
		len(stranded), strings.Join(stranded, ", "))
}

// requirePassClosesAllGaps refuses a PASS while ANY gap is still open. The protocol is "PASS
// only when every remaining gap is closed, rebuttal_sustained, or risk_accepted", and all of
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
	if len(open) != 0 {
		sort.Strings(open)
		return fmt.Errorf("record: verdict PASS refused — %d gap(s) still OPEN: %s. PASS requires every gap resolved through `close --id <id> --as closed|risk_accepted|rebuttal_sustained|routed_to_infrastructure`; close them, or issue `--as FAIL`",
			len(open), strings.Join(open, ", "))
	}
	// AND EVERY MOTION ANSWERED, which this gate did not check and a probe walked straight
	// through: a run reached `verdict PASS` and `outcome VERIFIED` with a grade motion filed and
	// never ruled. PASS is a claim that nothing is left open, and an unanswered ask is exactly
	// that — the report named it, which is the only reason it was visible at all.
	//
	// The gate counts what is on the RECORD, both vocabularies, because a pre-collapse record
	// replayed under this binary must be judged by the same standard it was written to.
	var unruled []string
	for _, m := range Motions(b) {
		if !m.Ruled() {
			unruled = append(unruled, m.ID)
		}
	}
	if len(unruled) != 0 {
		sort.Strings(unruled)
		// THE REFUSAL NAMES THE READ. Until `--view motions` existed, this message handed a seat
		// an id and no way to look it up — and a probed merge seat, blocked here, searched six
		// views and three help pages, then ruled `rejected` on an argument it had never read.
		// A blocking message that does not say how to unblock is an invitation to guess.
		return fmt.Errorf("record: verdict PASS refused — %d motion(s) filed and never ruled: %s. "+
			"Read what each one asks with `show motions` (its `basis` is the filer's argument, which your ruling answers), "+
			"then rule it with `motion <subject> rule --id <id> --as <verdict> --reason \"...\"`. "+
			"A motion is answered before the debate moves on, so a PASS over an unanswered ask claims a settlement that did not happen; rule them, or issue `--as FAIL`",
			len(unruled), strings.Join(unruled, ", "))
	}
	// AND EVERY LINE OF INQUIRY VOTED THIS ROUND.
	//
	// The report's account of its own research — "we pursued X", "we deferred Y", "we abandoned Z"
	// — reaches the reader as a row `assemble` GENERATES from the record. It carries no citation
	// anchor, so `lens verify` cannot reach it and the ordinary adversarial route does not apply.
	// Without this gate it is the one class of claim in the document that nothing could refuse.
	//
	// PER ROUND, not once: the report is regenerated every round, so a verdict cast before this
	// round's edits answers a question about a document that no longer exists. That is the whole
	// content of "red verifies every turn" — a carried-forward vote would be a stale read wearing
	// the shape of a fresh one, which is this repository's recurring defect rather than a fix for
	// it.
	if unvoted := UnvotedInquiries(b); len(unvoted) != 0 {
		ids := make([]string, len(unvoted))
		for i, a := range unvoted {
			ids[i] = a.ID
		}
		sort.Strings(ids)
		return fmt.Errorf("record: verdict PASS refused — %d line(s) of inquiry have no support verdict this round: %s. "+
			"READ THE REPORT ONCE (`show report`), list the lines with `show lines-of-inquiry`, and vote every one "+
			"against that single read — this is one pass over the document, not one pass per line: "+
			"`inquiry-support --id <id> --as supported|weakened|unsupported|absent --reason \"<what the report says there>\"`. "+
			"A PASS claims the report is sound, and its account of what this run investigated is part of the report; "+
			"vote them, or issue `--as FAIL`",
			len(ids), strings.Join(ids, ", "))
	}
	return nil
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
