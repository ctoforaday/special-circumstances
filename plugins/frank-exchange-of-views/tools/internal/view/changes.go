package view

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE DIFF STACK, MADE READABLE (#268 → #267).
//
// `blue edit` has appended a `blue_edit` event since 0.27.0 and NO PROJECTION RENDERED IT.
// The record's most detailed channel — every span blue replaced, with its reason — was
// written by one seat and read by nobody, while the `changelog` view rendered faithfully
// from a `revision` event no seat has emitted since the verb left the prompts. Two halves
// that never met.
//
// It matters now because estoppel cannot be adjudicated on evidence nobody can see. Red
// answers "did blue fix R1-1?" by re-reading the whole report and INFERRING. That full
// re-read stays mandatory — a decontextualized diff misleads on research prose, which is
// why the protocol forbids auditing from one. This view does not replace the read; it
// turns the follow-up question into a COMPARISON: here is the fix red required, and here
// is the change blue actually made, side by side, for the first time.
//
// Two forms, because they answer different questions:
//
//   - UNSCOPED — every recorded edit in round order. A navigation hint: what moved, who
//     moved it, and which gap each edit claims to answer.
//   - SCOPED to a gap (--id) — the comparison. Red's required_fix and acceptance check
//     above the edits blue recorded against them, with the exact old→new spans.

// changesMD renders the blue_edit stack. gapID "" is the unscoped form.
func changesMD(b *record.Board, gapID string) ([]byte, error) {
	if gapID != "" {
		return changesForGap(b, gapID)
	}

	out := []string{
		"# blue changes — RENDERED PROJECTION (source of truth: records/ event log)",
		"",
		"Every recorded edit to `blue/report.md`, in round order. This is a navigation hint, not",
		"a substitute for reading the report: a diff decontextualizes research prose, which is why",
		"the audit is a full re-read. To put red's required_fix and blue's actual change side by",
		"side, scope it — `show --view changes --id <gap>`.",
		"",
	}
	round, n, total := -1, 0, 0
	for _, e := range b.Events {
		if e.Type != "blue_edit" {
			continue
		}
		total++
		if e.Round != round {
			round, n = e.Round, 0
			out = append(out, fmt.Sprintf("## Round %d", e.Round), "")
		}
		n++
		answers := "_no gap — blue's own_"
		if id := e.Payload.Str("answers"); id != "" {
			answers = "answers **" + id + "**"
		}
		out = append(out, fmt.Sprintf("- **#%d** `%s` %s · %s — %s",
			n, e.SeatID, answers, delta(e.Payload.Str("old"), e.Payload.Str("new")), e.Payload.Str("text")))
	}
	if total == 0 {
		// SAID, NOT IMPLIED. An empty section reads as "nothing to show"; on a run past
		// round 0 it means blue changed the report through no recorded path, or did not
		// respond at all. Both are findings, and neither should look like formatting.
		out = append(out, "_No recorded edits. On a run past round 0 that is itself a finding:",
			"either blue answered nothing, or the report moved outside `blue edit`._", "")
	} else {
		out = append(out, "", fmt.Sprintf("_%d recorded edit(s)._", total), "")
	}
	return []byte(strings.Join(out, "\n")), nil
}

// changesForGap is the comparison form: what red required, then what blue did about it.
func changesForGap(b *record.Board, gapID string) ([]byte, error) {
	g, ok := b.Gaps[gapID]
	if !ok {
		return nil, fmt.Errorf("view changes: no gap %s on the board — --id takes a gap id (a mint created it), and a view that answered for an id nobody minted would be inventing a comparison", gapID)
	}

	out := []string{
		fmt.Sprintf("# blue changes answering %s — RENDERED PROJECTION", gapID),
		"",
		"## What red required",
		"",
	}
	if g.Mint != nil {
		// PAYLOAD KEY, NOT FLAG WORD. mint stores this under `required_fix` while the seat
		// types --fix (mint.go: seat.Set(cmd, p, "required_fix", flags.Fix)). Reading it as
		// "fix" compiles, renders, and silently reports every gap as unprescribed.
		if fix := g.Mint.Str("required_fix"); fix != "" {
			out = append(out, fix, "")
		} else {
			// A gap may legitimately carry no prescription — red states the defect and
			// leaves the remedy to blue. Naming that is not the same as an empty field.
			out = append(out, "_No required_fix was prescribed — red stated the defect and left the remedy to blue._", "")
		}
		// The CONCRETE proposal, when red made one. Shown as the diff blue could apply
		// verbatim — and labelled with what applying it costs and does not cost, because the
		// cheapest path is always compliance and that gradient should be visible, not felt.
		if fo, fn := g.Mint.Str("fix_old"), g.Mint.Str("fix_new"); fo != "" || fn != "" {
			out = append(out,
				"**Red proposed exact text** (`fix_basis: "+g.Mint.Str("fix_basis")+"` — red stated this against the live document, so it applies cleanly):",
				"",
				"```diff",
				"- "+oneLine(fo),
				"+ "+oneLine(fn),
				"```",
				"",
				"Blue is NOT obliged to apply it. Three paths stay open: apply it verbatim, counter-edit with",
				"blue's own fix, or dispute. Applying is the cheapest, which is exactly why declining has to",
				"stay a real option — a run in which blue never declines is manufacturing agreement, not earning it.",
				"")
		}
		if check := g.Mint.Str("acceptance_check"); check != "" {
			out = append(out, "**Acceptance check** (what red runs at re-audit): "+check, "")
		}
		if problem := g.Mint.Str("problem"); problem != "" {
			out = append(out, "**The defect**: "+problem, "")
		}
	}

	var edits []record.Event
	for _, e := range b.Events {
		if e.Type == "blue_edit" && e.Payload.Str("answers") == gapID {
			edits = append(edits, e)
		}
	}

	if len(edits) == 0 {
		out = append(out,
			"## Edits blue recorded against it: NONE",
			"",
			"_Blue recorded no edit answering this gap. That is not proof nothing changed — an edit",
			"that omitted `--answers` is invisible here — but it is the honest state of the join, and",
			"the unscoped `--view changes` lists every edit including the unattributed ones._", "")
		return []byte(strings.Join(out, "\n")), nil
	}

	out = append(out, fmt.Sprintf("## Edits blue recorded against it (%d)", len(edits)), "")
	for i, e := range edits {
		out = append(out,
			fmt.Sprintf("### %d. Round %d · `%s` · %s", i+1, e.Round, e.SeatID, delta(e.Payload.Str("old"), e.Payload.Str("new"))),
			"",
			"**Blue's reason**: "+e.Payload.Str("text"),
			"",
			"```diff",
			"- "+oneLine(e.Payload.Str("old")),
			"+ "+oneLine(e.Payload.Str("new")),
			"```",
			"")
	}
	return []byte(strings.Join(out, "\n")), nil
}

// delta reports the size change in CHARACTERS (runes, not bytes) — the unit a seat reading
// prose thinks in, and the one that does not make a multi-byte character look like an edit.
func delta(old, new string) string {
	d := utf8.RuneCountInString(new) - utf8.RuneCountInString(old)
	switch {
	case d > 0:
		return fmt.Sprintf("+%d chars", d)
	case d < 0:
		return fmt.Sprintf("%d chars", d)
	default:
		return "same length"
	}
}

// oneLine flattens a span onto a single line so the diff block stays a diff block.
//
// NOTHING IS ELIDED. A truncated span is exactly the decontextualized evidence this view
// exists to stop red reasoning from — a long edit is long, and the reader who wants it
// short can read the reason instead.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
