package seatprobe

import "fmt"

// ASKING THE SEAT WHAT IT THINKS ITS OPTIONS ARE, INSTEAD OF WATCHING WHICH IT TAKES.
//
// Every probe until now measured CHOICES: build a board, dispatch a seat, count the verbs it
// reached for. That number cannot distinguish the three things it might mean, and the difference
// decides what to fix:
//
//	NOT PERCEIVED   the option never entered the seat's model of what it could do
//	PERCEIVED, DECLINED   it saw the option and judged it not worth taking
//	WANTED, UNREACHABLE   it wanted the option and could not get to it
//
// A verb used zero times looks identical in all three. Fixing the first is a help-text or
// constitution problem; the second is a design question about whether the verb earns its place;
// the third is a capability gap. Guessing between them is how a verb gets deleted for being
// unloved when nothing ever told a seat it existed.
//
// So: same boards, same constitutions, same injected identity — and the seat is asked to
// ENUMERATE and ASSESS rather than act. What it lists is what it perceives. What it argues
// against is what it has weighed. What it says it wanted and could not do is friction it did not
// know how to file.
//
// # The prompt names no verb, and that is the whole method
//
// Listing the options for the seat would measure reading comprehension. The seat is given the
// situation and its own tools, and has to find the surface itself — which is exactly the act
// under test. It is told to read `--help` freely, because the question is not whether it
// remembers the surface but whether it VALUES what it finds there.

// ElicitPrompt is the sitting a seat is asked to reason about rather than perform.
//
// `sitting` is the prompt debate.js would have dispatched — passed in rather than composed here,
// so the acting arm and this one put the seat in the SAME situation and differ only in what they
// ask of it. Composing a second description of the situation is how the two arms would start
// measuring different boards while reporting one comparison.
//
// AND THE IDENTITY CONTRACT IS NOT RESTATED HERE. This wrapper used to carry its own paragraph
// about --seat-id, written when the situation was a two-line task. The situation is now
// debate.js's own prompt, which states the contract itself — and the two statements had already
// drifted apart: the wrapper said to pass --seat-id on every read, production says to state it
// once at register and never again. A seat handed both is being told to disobey one of them, and
// whichever it picks is scored as its judgement.
func ElicitPrompt(role, seatID, runDir, bin, sitting string) string {
	return fmt.Sprintf(`You are the %s seat in a frank-exchange-of-views run. Your seat id is %s.

Run directory: %s — use this ABSOLUTE path when you read files under it.

The record tool is the EXECUTABLE at %s. Invoke it by that absolute path, exactly as written — it is
a file, not a directory: do not cd into it, do not look for a binary inside it, and do not shorten
it.

DO NOT ACT. Record nothing, change nothing, write no files. This sitting is a question about
judgement, not a task — nothing you say here goes on the record. Everything below THE SITUATION is
the sitting you would have been given; read it as the situation, not as work to carry out.

First, read whatever you need in order to understand what is in front of you: the board, the
artifact under audit, your own available commands and their --help. Take as long as you need.

Then answer, in prose:

 1. WHAT ARE YOUR OPTIONS? Everything you could do about this situation, in your own words —
    including the ones you would not choose. If two options differ only in wording, say so.
 2. FOR EACH, what would it accomplish, what would it cost, and would you take it? Say plainly
    which you would do and which you would not, and why. "I would not bother" is a real answer
    and a useful one; so is "I do not understand what this would be for".
 3. IS ANYTHING MISSING? Something you would want to do about this situation and cannot see a
    way to do. Name it even if you suspect it exists and you have not found it.
 4. WOULD YOU KNOW WHEN YOU WERE FINISHED? Say what would tell you this sitting was complete,
    and whether anything actually tells you that.

Be concrete and be honest. An option you think is pointless is worth more to me than a polite
list of everything available.

THE SITUATION:

%s`, role, seatID, runDir, bin, sitting)
}
