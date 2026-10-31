package seatprobe

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// ceilings are the largest a seat's dispatched prompt may be, per seat class.
//
// # Why a ratchet and not a target
//
// The prompts spent months growing. Every measured defect arrived as a paragraph, every paragraph
// was individually justified, and nothing ever counted the total — so the seats were reading
// 24,000 characters of instruction against a tool whose own --help said much of it again, in the
// flag's usage line, at the moment of use. The subtraction that cut them roughly in half is only
// as durable as something that NOTICES the next paragraph, because each one will be as justified
// as the last.
//
// So this fails on GROWTH and says nothing about shrinkage. It is not a quality measure: a short
// prompt is not a good one, and nothing here can tell whether what is left is the right half. What
// it can tell is that a prompt got bigger, which is the direction the failure comes from.
//
// # Raising one is a decision, not a formality
//
// If a ceiling has to move, the question to answer first is whether the paragraph belongs in the
// TOOL: does it teach a verb, a flag, a refusal, a rendering, or an output format? All of those
// are the help's, where the seat reads them while choosing rather than three thousand characters
// earlier. What belongs here is the JOB — the standard, the discipline, the measured indictment,
// the situation this round is in.
//
// Recorded 2026-08-21, immediately after the subtraction, with ~10% headroom over the measured
// size. Re-pinned 2026-09-08 for roundless (plans/roundless.md §III.B.3): minting, screening,
// closing and the originator rule moved from the chair's prompt to the lens's with the verbs,
// so the chair fell to 8,627 and the lens rose to 10,658 — the same paragraphs, read by the
// seat that now performs them. Both ceilings moved to measured +10%; the pair's total shrank. The numbers before it: blue-respond 24,356 · red-merge 24,193 · red-lens 13,931 ·
// judge 11,861.
//
// Raised 2026-09-11 for the lens bar (plans/feov-lens-bar.md III.4, D10): the lens reads its last
// sitting off its work view instead of a head the chair relayed, and its prompt now says how to
// scope the audit by each kind and that a fresh gap on text it already passed owes why it was
// missed — the JOB, which §V #6 requires the prompt to carry. The lens measured 11,846 (+465 net of
// the relayed head clause it replaced); the ceiling moved to 12,000, not +10%, to keep the ratchet
// tight.
//
// Raised 2026-09-15 for the lens bar (plans/feov-lens-bar.md III.5, §V #6): the chair relays a
// dispatch readied by each lens's last sitting against the head and records the PASS that permits, so its prompt says
// that a lens sits again only on a moved head and retires, that a PASS needs no lens ready, that the stale areas are
// read and named in the spot-check before it, and that the PASS lists every open gap that is not
// material by class — the JOB, which §V #6 requires the prompt to carry. The refusal that restated
// "the tool refuses a PASS the board does not permit" was cut from it; that refusal is the verdict
// help's. The chair measured 10,059 (+559); the ceiling moved to 10,200.
//
// Raised 2026-09-15 when the lens bar met OCR citations (plans/ocr-cite-derives-page.md): the lens's
// duty to check an OCR citation against one page's image, and its verbatim read of a PDF through the
// run's cache, landed under the 11,700 ceiling on main; with the last-sitting clause beside them the
// lens measured 12,166 (+320 over the lens bar's 11,846). The two are separate duties, both the JOB;
// the ceiling moved to 12,300.
// Raised 2026-09-18 for the shared scratchpad: the manual instruction every seat reads now names
// the file with the seat's own id and says why — the seats of a run share ONE scratchpad, so an
// unseated name is a name another seat is also writing. Measured on the 2026-09-17 smoke run,
// where nine seats wrote one manual.txt, one report.txt and one board.json between them; a lens
// diffed what it took to be its own earlier manual, found another seat's banner, and filed a tool
// defect against a tool that was correct. The read side is the cost that is not cosmetic: these
// files are re-read INSTEAD of re-running a projection, so a seat can read another seat's board.
//
// It is the JOB by this comment's own test — it teaches no verb, flag, refusal or rendering, it
// states the discipline and the property of the environment the discipline follows from. The
// instruction is shared by every seat, so every prompt pays it; only the judge had less than 39
// characters of headroom. 7,800 -> 7,850, the ratchet kept tight rather than +10%.
// Raised 2026-09-19 because the previous wording is REFUSED, which makes this the one entry here
// that is not discretionary growth. The paragraph read "YOUR REASONING IS PART OF EVERY ACT ...
// write your THINKING", and every blue lane dispatched on sonnet-5 returned stop_reason=refusal
// with zero output tokens, the API naming its own classifier: [reasoning_extraction]. The same
// text ran on the same model two days earlier, so the prompt did not change into a refusal — the
// guard tightened under it. Paired at one moment: the old wording refuses in five seconds, the new
// one runs; haiku-4.5 accepts both.
//
// What the seat is asked for is its CASE TO THE OTHER SEATS — grounds an opponent can answer — and
// never an account of how the model thinks. The old phrasing asked for the second while meaning the
// first. Every instruction survives; only the words naming the model's interior do not, and saying
// it takes more of them. red-lens-evidence absorbed the change inside its existing ceiling; the
// judge measured 7,879 (+29), so 7,850 -> 7,900.
//
// Lowered 2026-10-17 (issue #1268, phase D) to each seat's measured size rounded up to the next
// hundred, after the prompts stopped restating what the seat's constitution carries: blue-respond
// 9,419, red-chair 6,260, red-lens-evidence 10,419, judge 5,840. The ceilings had drifted far above
// the prompts as the clauses moved out, so growth up to the old numbers passed unseen.
//
// Held 2026-10-17 for the chair's final sitting: one sentence of duty — at final_sitting nobody sits
// after the chair, so it decides on the board as it is.
// Smoke m25's last chair filed a docket motion it said "the bench hears at the next dispatch", in the
// run's last epoch, and the run ended CEILING over a board that permitted a PASS. It fits inside
// 6,300 (red-chair 6,297): the prompt's account of epoch_limit_reached — that nobody further is
// dispatched — is the dispatch page's, and the sentence carries the duty that follows from it. The
// refusal, the petition the exit still hears, and the field are the tool's (motion and dispatch help).
// A ceiling rises only when the duty cannot fit in it.
//
// Raised 2026-10-17 for the owner's ruling that a bench holding reaches a seat only as a declaration
// on the record, on its work list with the board's status beside each gap it names. Every seat's
// prompt states the DUTY — this run's bench holdings bind how the record is read, each is on the
// work list, and the board, not a holding, says where a gap stands — 138 characters on the one
// clause every seat receives, the shortest wording that keeps all three and keeps a declaration
// apart from the law's holdings of earlier runs, which only persuade; the declarations themselves
// are the work list's, documented on its help page. No seat had that much headroom (blue-respond
// and red-lens-evidence 81, judge 65), so each ceiling rises by its measured excess to the measured
// size: blue-respond 9,557, red-lens-evidence 10,557, judge 5,973. The chair holds 6,300 with the
// duty and the final-sitting sentence both, measuring 6,284, by leaving two mechanics to the tool:
// its relay sentence names what it must not alter by class (anything dropped, added or altered fails
// the dispatch-parity audit, a blocker, pass_permitted or ceiling included) rather than field by
// field, and the PASS refusal of a board that does not permit one is the verdict page's.
//
// Raised 2026-10-30 for the owner's ruling that a bench ruling and a granted petition's relief reach
// a seat on the record, on its work list, and never through the engine. The prompts had carried them
// as clauses rendered only when a ruling or a grant existed, so no board here measured them; they are
// gone, and every seat's record clause states the duty instead — the rulings and orders bind what it
// does, beside the holdings that bind how it reads — 36 characters. The lens's ownership sentence
// names the bench as the one other seat that regrades a gap (26), and the bench's prompt carries the
// duty to set a grade it holds wrong before it rules (61); the lock, the window and the new-grounds
// check are the regrade and rule pages'. red-lens-evidence had 4 characters of headroom and the judge
// 12, so each rises to its measured size: red-lens-evidence 10,614, judge 6,057. The chair and
// blue-respond hold.
//
// Raised 2026-10-30 on merging the computation (epoch 49): the lens's check sentence speaks the
// computation — its program, and the bundle run once with this computation's output compared — 10
// characters more on the paragraph every lens receives, which this ruling's 10,614 had no room for.
// red-lens-evidence rises to its measured size, 10,616; the other ceilings hold.
//
// Raised 2026-10-31 for the halt read from the record (#1663 smoke, D4): a halt reaches the engine
// as the plan's halted, never as an envelope, so the chair's sentence on what an empty plan means
// names it beside pass_permitted and ceiling — "; with halted the bench ended it", 32 characters. A
// list of the reasons a plan is empty that leaves one out reads a halted run as a fault; the field's
// account is the dispatch page's. red-chair had 4 characters of headroom and rises to its measured
// size, 6,328; the other ceilings hold.
var ceilings = map[string]int{
	"blue-respond":      9557,
	"red-chair":         6328,
	"red-lens-evidence": 10616,
	"judge":             6057,
}

func TestNoSeatPromptGrowsPastItsCeiling(t *testing.T) {
	got := map[string]int{}
	for _, b := range Boards() {
		d, err := ProductionPrompt(debateScriptForTest(), b, "/runs/x", "/bin", "haiku", "haiku")
		if err != nil {
			t.Errorf("%s: %v", b.Name, err)
			continue
		}
		if n := len(d.Prompt); n > got[d.SeatID] {
			got[d.SeatID] = n
		}
	}
	if len(got) == 0 {
		t.Fatal("no board produced a prompt — a size gate that measures nothing passes every time")
	}

	var table strings.Builder
	seats := make([]string, 0, len(got))
	for s := range got {
		seats = append(seats, s)
	}
	sort.Strings(seats)
	for _, s := range seats {
		fmt.Fprintf(&table, "  %-18s %6d (ceiling %d)\n", s, got[s], ceilings[s])
	}
	t.Log("dispatched prompt sizes:\n" + table.String())

	for _, s := range seats {
		limit, ok := ceilings[s]
		if !ok {
			t.Errorf("seat %s has no recorded ceiling (%d chars) — a seat the ratchet does not cover can grow without anyone seeing it", s, got[s])
			continue
		}
		if got[s] > limit {
			t.Errorf("%s is %d characters, past its ceiling of %d.\n\n"+
				"Before raising it: does the new text teach a VERB, a FLAG, a REFUSAL, or a RENDERING? "+
				"That belongs in the tool's --help, where the seat reads it while choosing. "+
				"The prompt carries the JOB — the standard, the discipline, and this round's situation.",
				s, got[s], limit)
		}
	}
	// A ceiling naming a seat production does not dispatch is a gate on nothing.
	for s := range ceilings {
		if _, ok := got[s]; !ok {
			t.Errorf("ceiling recorded for %s, which no board dispatches — the entry guards nothing", s)
		}
	}
}
