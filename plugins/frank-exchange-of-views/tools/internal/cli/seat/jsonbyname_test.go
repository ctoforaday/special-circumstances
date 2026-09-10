package seat

import (
	"strings"
	"testing"
)

// SIX SEATS DID NOT GUESS THE SCHEMA WRONG.
//
// #593 read the `KeyError: 'sitting'` crashes — three seats in each of two runs — as seats
// inferring a schema "from something common and wrong", and proposed publishing the real one.
// The schema was already published and already correct: `show work`'s own description documents
// `sitting.open` and `sitting.complete`, and `sitting` is a real top-level key.
//
// What the six actually did was read a description shouting STRUCTURED JSON, reach for the
// tool's `--json` flag, and receive a REFUSAL — `{"ok":false,"code":…,"error":…}` at exit 2 —
// which parses exactly as well as the projection does. Piped into python, the refusal's advice
// never arrived; it surfaced as a missing key.
//
// These tests pin the two halves of the repair: the fact lives in ONE place, and the warning
// reaches the seat where the guess is formed.

// THE SET IS ONE FACT. It used to be two hand-kept carriers — the `long` prose and a switch case
// naming six views — with nothing holding them together. TestDebateJSONViewAndOneWayContract's
// comment records the cost: `friction` survived in a list like this after it stopped being a
// view, and the assertion kept passing because it demands an error and an unknown view errors too.
func TestJSONByNameSetIsDerivedFromTheTableAndNotEmpty(t *testing.T) {
	got := JSONByNameViews()
	if len(got) == 0 {
		t.Fatal("no view is marked jsonByName — the refusal would then never fire and would read as no views needing it")
	}
	for _, name := range got {
		if !viewIsJSONByName(name) {
			t.Errorf("%s is listed but does not answer the predicate — two readers of one field", name)
		}
	}
	// Every marked name must be a REAL view. A stale name here is the exact defect the contract
	// test recorded, and it is checkable now that both come from the table.
	real := map[string]bool{}
	for _, n := range ViewNames() {
		real[n] = true
	}
	for _, name := range got {
		if !real[name] {
			t.Errorf("%q is marked jsonByName and is not a view — a stale name checks nothing while reading as coverage", name)
		}
	}
}

// THE WARNING TRAVELS WITH THE MARK. A view that is JSON by name carries it; one that is not must
// not, or the warning stops meaning anything.
func TestEveryJSONByNameViewWarnsWhereTheGuessIsFormed(t *testing.T) {
	marked := map[string]bool{}
	for _, n := range JSONByNameViews() {
		marked[n] = true
	}
	for _, v := range views {
		help := v.long + jsonByNameHelpSuffix(v.jsonByName)
		carries := strings.Contains(help, "ALREADY THE JSON")
		if marked[v.name] && !carries {
			t.Errorf("%s is JSON by name and its help does not say so — the guess is formed here", v.name)
		}
		if !marked[v.name] && carries {
			t.Errorf("%s is not JSON by name and warns anyway", v.name)
		}
	}
}

// THE NOTE SAYS THE FLAG IS HARMLESS, and never that it is refused. It used to warn of a refusal
// that parsed as data; the refusal is gone, and a note still threatening one would teach a seat
// to avoid the flag that now works — a stale warning is a wrong instruction.
func TestTheNoteSaysSuccessIsIdenticalAndErrorsAreEnvelopes(t *testing.T) {
	// BOTH HALVES, because "changes nothing" was measured false on the error path: under --json an
	// error on this view is an {ok:false} envelope on stdout — the same shape that crashed six
	// pipelines as a missing key (#593). The note may call the flag harmless on success only.
	for _, want := range []string{"ALREADY THE JSON", "accepted", "byte-for-byte the same", "ERROR", "check `ok`"} {
		if !strings.Contains(jsonByNameWarning, want) {
			t.Errorf("the note must carry %q", want)
		}
	}
	for _, never := range []string{"REFUSED", "changes nothing"} {
		if strings.Contains(jsonByNameWarning, never) {
			t.Errorf("the note says %q, which is false", never)
		}
	}
}
