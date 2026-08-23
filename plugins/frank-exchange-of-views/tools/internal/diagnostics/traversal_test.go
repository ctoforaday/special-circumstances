package diagnostics

import "testing"

// THE SHAPES THAT DEFEATED THE FIRST DRAFT, each measured in
// research/2026-08-22_record-store-authority. Every one returned NO verb path, and a seat with no
// verb paths reports zero group crossings — which reads exactly like a seat that never left one
// group. These are regression fixtures for that silence, not for the parser's convenience.
func TestCommandWordsReadsTheShapesSeatsActuallyWrite(t *testing.T) {
	for _, c := range []struct {
		name, command string
		want          []string
	}{
		{
			// The majority shape. BinFields matches the path inside the ASSIGNMENT, so the fields
			// begin at `;` — and breaking there cost blue-synthesize 26 of its 27 invocations.
			name:    "assignment then invocation",
			command: `B="/tmp/x/runbin/feov-record"; $B --seat-id blue-synthesize show lines-of-inquiry 2>&1 | head -120`,
			want:    []string{"show", "lines-of-inquiry"},
		},
		{
			// Flags carried in a variable. `$S` is neither a dash-flag nor a verb token, and the
			// loop broke on it: red-merge-r1 made thirty recognised calls and one was classified.
			//
			// THE ASSIGNMENT IS PART OF THE FIXTURE ON PURPOSE. My first draft of this case wrote
			// the invocation alone — `"$B" $S show work` — and it failed, correctly: with no
			// literal `feov-record` anywhere on the line BinFields recognises no tool call at
			// all, and there is nothing for CommandWords to read. That is a real limit and it is
			// ToolUnrecognised's to report, not this function's. Every occurrence in the corpus
			// carries the assignment on the same line, which is why the alias is readable.
			name:    "flags expanded from a variable",
			command: `B="/tmp/x/feov-record"; S="--seat-id red-merge-r1"; "$B" $S show work 2>&1 | python3 -c "import json"`,
			want:    []string{"show", "work"},
		},
		{
			name:    "plain invocation still reads",
			command: `"/tmp/x/feov-record" --seat-id red-merge-r1 mint --id R1-1`,
			want:    []string{"mint"},
		},
		{
			// The guard that must NOT regress: a separator AFTER a verb still ends the path, or a
			// second command on the line would be read as a subcommand of the first.
			name:    "a separator after the verb still ends it",
			command: `feov-record show board && echo done`,
			want:    []string{"show", "board"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := CommandWords(BinFields(c.command, "feov-record"))
			if len(got) != len(c.want) {
				t.Fatalf("got %q, want %q", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("got %q, want %q", got, c.want)
				}
			}
		})
	}
}

// ZERO CROSSINGS IS THE ANSWER TO TWO QUESTIONS, and Unclassified is what keeps them apart.
func TestATraversalThatCouldNotReadTheCallsSaysSo(t *testing.T) {
	known := map[string]bool{"show": true, "motion": true, "mint": true}
	unreadable := traversalOf([]call{{path: nil}, {path: nil}, {path: nil}}, known)
	if unreadable.Crossings != 0 || unreadable.Unclassified != 3 {
		t.Errorf("crossings=%d unclassified=%d, want 0 and 3", unreadable.Crossings, unreadable.Unclassified)
	}
	oneGroup := traversalOf([]call{{path: []string{"show", "board"}}, {path: []string{"show", "work"}}}, known)
	if oneGroup.Crossings != 0 || oneGroup.Stays != 1 || oneGroup.Unclassified != 0 {
		t.Errorf("a seat that stayed in one group: crossings=%d stays=%d unclassified=%d",
			oneGroup.Crossings, oneGroup.Stays, oneGroup.Unclassified)
	}
	// The two are now distinguishable, which is the whole point of the field.
	if unreadable.Crossings == oneGroup.Crossings && unreadable.Unclassified == oneGroup.Unclassified {
		t.Error("an unreadable sitting and a single-group sitting still report identically")
	}
}

// HELP READS ARE NOT OPERATIONS. Surveying a surface is not moving across it, and counting the
// three-step help discipline as group-hopping would make every compliant seat look leaky.
func TestHelpReadsAreNotTraversal(t *testing.T) {
	got := traversalOf([]call{
		{path: []string{"show"}, isHelp: true},
		{path: []string{"motion"}, isHelp: true},
		{path: []string{"show", "board"}},
	}, map[string]bool{"show": true, "motion": true})
	if got.Crossings != 0 || got.Stays != 0 || got.Unclassified != 0 {
		t.Errorf("help reads leaked into the traversal: %+v", got)
	}
}

// EMPTY IS `[]`, NOT `null` — a consumer iterating every seat's pairs must not die on the first
// seat that never left a group.
func TestEmptyTraversalMarshalsAsEmptySlices(t *testing.T) {
	got := traversalOf(nil, nil)
	if got.Pairs == nil || got.Sequence == nil {
		t.Errorf("nil slices marshal as null: pairs=%v sequence=%v", got.Pairs, got.Sequence)
	}
}

// PROSE THAT ESCAPED A --reason BODY IS NOT A GROUP THE SEAT VISITED.
//
// CommandWords accepts any lowercase word, so a real corpus produced crossings named
// `and -> names` and `edit -> returns`. Checking each top against the seat's own verb set is what
// separates a command from a word, and an unrecognised token lands in Unclassified rather than
// being dropped — a token the tree does not know is exactly a call the traversal could not read.
func TestProseIsNotAGroup(t *testing.T) {
	known := map[string]bool{"show": true, "finding": true}
	got := traversalOf([]call{
		{path: []string{"show", "board"}},
		{path: []string{"and", "names"}}, // prose
		{path: []string{"returns"}},      // prose
		{path: []string{"finding", "--key"}},
	}, known)
	if got.Unclassified != 2 {
		t.Errorf("unclassified=%d, want 2 (the two prose tokens)", got.Unclassified)
	}
	// One real crossing survives: show -> finding. The prose must not have manufactured three.
	if got.Crossings != 1 {
		t.Errorf("crossings=%d, want 1", got.Crossings)
	}
	for _, p := range got.Pairs {
		if p.From == "and" || p.To == "names" || p.From == "returns" {
			t.Errorf("a prose token reached the pairs: %+v", p)
		}
	}
}
