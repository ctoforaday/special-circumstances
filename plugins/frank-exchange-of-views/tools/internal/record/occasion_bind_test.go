package record

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// occasionArg captures the OCCASION handed to recordClause — its second argument, which only the
// bench's four dispatches pass. The words are constants (DOCKET_OCCASION and its three siblings)
// so that each site's prompt and its dispatch label cannot drift apart; this reads the constant
// NAME, and occasionConst below reads what each one is bound to.
var occasionArg = regexp.MustCompile(`recordClause\([^,)]*,\s*([A-Za-z0-9_]+)\s*\)`)

// occasionConst reads `const X_OCCASION = 'word'`.
var occasionConst = regexp.MustCompile(`const\s+([A-Za-z0-9_]+_OCCASION)\s*=\s*'([a-z_]+)'`)

// THE ENGINE DISPATCHES THE WORDS THE WRITE PATH ACCEPTS, IN BOTH DIRECTIONS.
//
// A bench seat copies its OCCASION out of the prompt and types it at `register`, where checkOccasion
// resolves it against the Occasion enum and refuses anything else. So an engine that says a word the
// schema does not carry wedges every bench sitting in the run — at the bench's FIRST act, after the
// whole debate has been paid for.
//
// The other direction is the one that goes quiet rather than loud: a schema value no dispatch
// produces is a sitting kind nothing can ever be, and every reader that switches on it has a dead
// arm nobody will notice. That is the shape this repository keeps finding — `assemble` stopped being
// a seat id and five `Seat == "assemble"` comparisons in the dashboard became unsatisfiable, so every
// FINISHED run reported "running" forever and no test failed.
//
// WHY A BIND AND NOT GENERATION, since the repository's rule prefers generating a derived carrier:
// there is nothing here to generate. Which occasion the assembly sitting has is an authored fact
// about that one call site, not a copy of a list — debate.js holds four USES of the vocabulary, not
// a second copy of it. What can drift is whether those four words are the four the write path takes,
// and that is what this checks.
func TestTheEngineDispatchesTheSchemasOccasions(t *testing.T) {
	path, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read debate.js — the dispatch source this binds against: %v", err)
	}
	js := string(b)

	bound := map[string]string{}
	for _, m := range occasionConst.FindAllStringSubmatch(js, -1) {
		bound[m[1]] = m[2]
	}

	dispatched := map[string]bool{}
	for _, m := range occasionArg.FindAllStringSubmatch(js, -1) {
		word, ok := bound[m[1]]
		if !ok {
			t.Errorf("recordClause passes occasion %s, which no `const %s = '…'` in debate.js binds — "+
				"this bind cannot reduce it to a word, and an unread argument is a vocabulary that is no longer checked", m[1], m[1])
			continue
		}
		dispatched[word] = true
	}
	if len(dispatched) == 0 {
		t.Fatal("no recordClause call passes an occasion — either the bench stopped naming what its sittings are for, " +
			"or this bind stopped being able to read the call. Both are failures; neither is a clean pass")
	}

	schema := map[string]bool{}
	for _, w := range OccasionWords() {
		schema[w] = true
	}
	for w := range dispatched {
		if !schema[w] {
			t.Errorf("debate.js dispatches the bench on occasion %q, which the Occasion enum does not carry — "+
				"`register` refuses it, so that sitting cannot record its first act at all", w)
		}
	}
	for w := range schema {
		if !dispatched[w] {
			t.Errorf("the Occasion enum carries %q and no bench dispatch uses it — a sitting kind nothing can ever be, "+
				"and every reader switching on it has an arm no run reaches", w)
		}
	}

	// Named so a failure above reads against the actual sets rather than against a count.
	if t.Failed() {
		t.Logf("dispatched: %s; schema: %s", sortedOf(dispatched), sortedOf(schema))
	}
}

func sortedOf(m map[string]bool) string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return "[" + strings.Join(out, ", ") + "]"
}
