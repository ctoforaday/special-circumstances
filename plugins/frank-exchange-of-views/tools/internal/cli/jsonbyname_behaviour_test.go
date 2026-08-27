package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// THE MARK IS PINNED AGAINST BEHAVIOUR, BECAUSE SELF-CONSISTENCY HAS NO TEETH.
//
// The first version of this repair marked each JSON-by-name view with a field and tested that
// the field agreed with the generated help. Unmarking `work` passed that test — both halves
// simply agreed there was nothing to say — and it passed the existing one-way contract test too,
// because that demands an ERROR from `show work --json` and the unmarked fallback is also an
// error. Two green tests over a view that had quietly stopped being marked.
//
// That is the defect TestDebateJSONViewAndOneWayContract's own comment records ("a stale name
// checks nothing while reading as coverage"), reproduced by the fix for it. So the mark is
// checked against the only thing that cannot agree with it out of politeness: what the bare
// projection actually emits.
func TestJSONByNameMarkMatchesWhatTheBareViewEmits(t *testing.T) {
	runDir := seatRun(t)
	// A GAP, so the run has a round: `telemetry` is dense over rounds 1..current, so on a
	// fixture where nothing was ever minted it emits nothing at all and the loop below logs it
	// UNCHECKED. An unchecked view is the mark going untested in the one shape (JSONL) that the
	// whole-document parse cannot reach — exactly the case this test exists to hold.
	mintGap(t, runDir, "jsonbyname-mark", "read-surface")
	marked := map[string]bool{}
	for _, n := range seat.JSONByNameViews() {
		marked[n] = true
	}

	for _, view := range seat.ViewNames() {
		out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", view)
		if err != nil {
			// A view this seat's role cannot open says nothing about the mark.
			continue
		}
		trimmed := strings.TrimSpace(out)
		if trimmed == "" {
			// AN EMPTY PROJECTION DISCRIMINATES NOTHING and must not read as agreement: a fresh
			// fixture run has no telemetry rounds yet, and "no output" is neither JSON nor prose.
			t.Logf("show %s emitted nothing on this fixture — the mark is UNCHECKED here, not confirmed", view)
			continue
		}
		emitsJSON := emitsJSONOrJSONL(trimmed)

		switch {
		case emitsJSON && !marked[view]:
			t.Errorf("show %s emits JSON with no flag and is NOT marked jsonByName — a seat passing --json "+
				"gets the wrong refusal, and nothing warns it that the refusal parses as data", view)
		case !emitsJSON && marked[view]:
			t.Errorf("show %s is marked jsonByName and its bare form is not JSON — the warning it carries is false", view)
		}
	}
}

// AND THE REFUSAL ITSELF PARSES AS DATA, which is the fact the whole repair turns on. Pinned so
// nobody has to rediscover it from six crashed seats: a pipeline reading stdout without checking
// `ok` gets a well-formed object carrying none of the projection's keys.
func TestTheJSONByNameRefusalIsShapedLikeTheDataItIsNot(t *testing.T) {
	runDir := seatRun(t)
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", "work", "--json")
	if err == nil {
		t.Fatal("show work --json must refuse ([[one-way-no-aliases]])")
	}
	// The refusal travels in the structured envelope, which is correct for --json and is exactly
	// what makes it dangerous to a pipeline: it decodes.
	var envelope map[string]any
	if json.Unmarshal([]byte(err.Error()), &envelope) == nil {
		if _, hasSitting := envelope["sitting"]; hasSitting {
			t.Error("the refusal envelope carries a sitting key; this test no longer describes the trap")
		}
	}
	if !strings.Contains(err.Error(), "already JSON by name") {
		t.Errorf("the refusal must still teach the fix; got: %v", err)
	}
	// And it must now also name the consequence, which is what did not reach the six seats.
	if !strings.Contains(err.Error(), "checking `ok`") {
		t.Errorf("the refusal must say why a pipeline never sees it; got: %v", err)
	}
	_ = out
}

// WHOLE DOCUMENT FIRST, THEN PER LINE — in that order, because each check is wrong about the
// other's shape. The board/findings/work/motions/evidence projections are marshalled INDENTED, so
// every one of them is a multi-line object whose individual lines (`{`, `  "open": [`) are not
// JSON; a line-wise check convicts all five. `telemetry` is the mirror image: one object PER
// ROUND, a stream no single-document parse accepts. Checking one way only is how this test twice
// reported the mark false when the mark was right.
func emitsJSONOrJSONL(trimmed string) bool {
	var whole any
	if json.Unmarshal([]byte(trimmed), &whole) == nil {
		return true
	}
	lines := 0
	for _, line := range strings.Split(trimmed, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var one any
		if json.Unmarshal([]byte(line), &one) != nil {
			return false
		}
		lines++
	}
	// A single line that failed the whole-document parse is not a stream of one; it is prose.
	return lines > 1
}
