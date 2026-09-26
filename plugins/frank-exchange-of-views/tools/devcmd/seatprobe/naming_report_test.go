package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// THE SURVIVING-NAME COUNT IS PRINTED WITH THE RESULT, and these pin the line.
//
// It used to guard an ARM: a redactor that had stopped matching would produce a `none` run
// byte-identical to `partial`, and both would report the same behaviour — a null manufactured by
// the instrument. The arms are gone, and the check that outlived them is the stronger one: the
// SHIPPED constitution must name no verb, so a name surviving is a regression rather than a
// treatment that failed to land.

func TestNamingTreatmentSaysNotMeasuredRatherThanZero(t *testing.T) {
	// "0 names survived" and "I could not open the constitution" are different facts, and only
	// one of them means the treatment worked. Reporting the second as the first is the exact
	// shape this whole report exists to avoid.
	got := namingTreatment("lens", "/nonexistent-constitutions", seatprobe.NewSurface(cli.CommandPaths()))
	if !strings.Contains(got, "NOT MEASURED") {
		t.Errorf("an unreadable constitution reported %q — it must say NOT MEASURED, never a count", got)
	}
}

// THE CONSTITUTIONS NOW CARRY THEIR WHOLE SURFACE, and the gate that forbade it is retired here.
//
// It held that a constitution names no verb, on a measurement worth keeping in view: a PARTIAL
// list in front of a seat satisfied its need to know what exists and stopped it looking — 58% of
// the surface seen against 95% with the list removed. That finding was about a SLICE, under a
// design where a seat learned its surface by asking the tool for help on one verb at a time.
//
// The design it guarded is the one being replaced. scripts/agentgen inlines the COMPLETE surface —
// byte-identical to what `manual` returns, and kept so by `agentgen -check` — because the fetch,
// not the content, was the cost: 171 tool calls in one run, 10% of every call made, and 69% of
// everything a barren lens read before it looked at the report. A whole surface cannot satisfy a
// seat into not looking at the part it was not shown; there is no such part.
//
// The prompt-side gates STAY and are unaffected: TestTheSeatPromptsNameNoVerb holds debate.js, and
// TestNoPromptGrowsItsCommandCatalogue still scans every constitution OUTSIDE the generated block,
// so a verb name creeping into authored prose fails exactly as it did.
