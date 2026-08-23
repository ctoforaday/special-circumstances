package checkpointseal

import (
	"strings"
	"testing"
	"time"
)

// The instruction REINFORCES what the session established and never INTRODUCES.
// Measured basis: a PreCompact instruction naming content absent from the
// conversation was flagged by the summarizer as a prompt-injection attempt, and
// the flag was written into the summary itself.
func TestSteerOnlyAsksForSectionsTheNoteCarries(t *testing.T) {
	cases := []struct {
		name      string
		note      string
		user      string
		wantHas   []string
		wantLacks []string
		wantEmpty bool
	}{
		{
			name:      "no note means no instruction at all",
			note:      "",
			wantEmpty: true,
		},
		{
			name:      "a note with no recognised section asks for nothing",
			note:      "## Shopping list\n- milk\n",
			wantEmpty: true,
		},
		{
			name:      "only the sections present are requested",
			note:      "## Validation loop\n1. go test ./...\n## Open threads\n- something\n",
			wantHas:   []string{"the validation loop", "the open threads"},
			wantLacks: []string{"in-flight task handles", "ordered next actions"},
		},
		{
			name: "every section present is requested, most-dropped first",
			note: "## Validation loop\n## Next intended steps\n## In-flight handles\n" +
				"## Invariants / foot-guns\n## Decisions made / rejected\n## Open threads\n",
			wantHas: []string{"validation loop", "ordered next actions", "in-flight task handles",
				"invariants and foot-guns", "re-litigated", "open threads"},
		},
		{
			name:    "a manual /compact instruction is carried through, not clobbered",
			note:    "## Validation loop\n1. go test ./...\n",
			user:    "focus on the auth refactor",
			wantHas: []string{"focus on the auth refactor", "the validation loop"},
		},
		{
			name:      "user instruction survives even when the note offers nothing",
			note:      "",
			user:      "keep the migration plan",
			wantHas:   []string{"keep the migration plan"},
			wantLacks: []string{"Also preserve"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := steer(tc.note, tc.user)
			if tc.wantEmpty {
				if got != "" {
					t.Fatalf("expected silence, got %q", got)
				}
				return
			}
			for _, w := range tc.wantHas {
				if !strings.Contains(got, w) {
					t.Errorf("missing %q in %q", w, got)
				}
			}
			for _, w := range tc.wantLacks {
				if strings.Contains(got, w) {
					t.Errorf("unexpected %q in %q", w, got)
				}
			}
		})
	}
}

// The instruction must never paste the note's content — that is what turns a
// reinforcement into an injection.
func TestSteerNeverEchoesNoteBody(t *testing.T) {
	note := "## Validation loop\n1. SECRET-COMMAND --do-the-thing\n## Open threads\n- LEAKY-DETAIL\n"
	got := steer(note, "")
	for _, leak := range []string{"SECRET-COMMAND", "LEAKY-DETAIL"} {
		if strings.Contains(got, leak) {
			t.Fatalf("instruction echoed note body (%q): %q", leak, got)
		}
	}
}

// Concurrent seats share the parent's session_id, so the seal name must carry
// agent_id or parallel seats overwrite each other.
func TestSnapshotNameDisambiguatesConcurrentSeats(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 34, 56, 0, time.UTC)
	a := snapshotName(now, "auto", "aaa111")
	b := snapshotName(now, "auto", "bbb222")
	if a == b {
		t.Fatalf("same name for two seats at the same instant: %q", a)
	}
	if want := "20260728T123456Z-auto-aaa111.md"; a != want {
		t.Errorf("got %q want %q", a, want)
	}
	if got := snapshotName(now, "manual", ""); got != "20260728T123456Z-manual.md" {
		t.Errorf("main session name: got %q", got)
	}
	if got := snapshotName(now, "", ""); !strings.Contains(got, "-unknown.md") {
		t.Errorf("missing trigger should not produce a bare name: %q", got)
	}
}

// Bounded collection, explicit lifecycle: the newest keep survive, oldest go.
func TestPruneKeepsNewest(t *testing.T) {
	names := []string{
		"20260728T000300Z-auto.md",
		"20260728T000100Z-auto.md",
		"20260728T000200Z-auto.md",
	}
	got := prune(names, 2)
	if len(got) != 1 || got[0] != "20260728T000100Z-auto.md" {
		t.Fatalf("prune(keep=2) = %v, want the oldest only", got)
	}
	if got := prune(names, 5); got != nil {
		t.Errorf("under the limit should delete nothing, got %v", got)
	}
	if got := prune(names, 0); len(got) != 3 {
		t.Errorf("keep=0 should delete all, got %v", got)
	}
	// The retention BOUNDARY. The cases above are at keep=2 (over), keep=5 (well
	// under) and keep=0, so none of them lands on len(names) == keep — a mutation
	// sweep confirmed `<=` could become `<` with the suite still green, which would
	// delete the oldest snapshot on every seal once the directory reached its limit.
	if got := prune(names, 3); got != nil {
		t.Errorf("exactly at the limit must delete nothing, got %v", got)
	}
	if got := prune(names, 4); got != nil {
		t.Errorf("one under the limit must delete nothing, got %v", got)
	}
	// A negative keep is nonsense input, and the answer to nonsense is to delete
	// nothing rather than to compute a negative slice bound and panic inside a hook
	// that must never fail.
	if got := prune(names, -1); got != nil {
		t.Errorf("a negative keep must delete nothing, got %v", got)
	}
	if got := prune(nil, 2); got != nil {
		t.Errorf("an empty directory must delete nothing, got %v", got)
	}
}

// occasion is what a reader uses months later to tell a compaction seal from a
// session's last words, and it is NEVER empty by contract. The per-event refinements
// were unpinned: a mutation sweep inverted `in.Reason != ""` with the suite green, which
// yields the dangling label "end-" for a SessionEnd carrying no reason — and `other` is
// the reason a headless `claude -p` run reports, so that is not a rare shape.
func TestOccasionIsNeverEmptyAndNamesItsCause(t *testing.T) {
	cases := []struct {
		event string
		in    hookInput
		want  string
	}{
		{evSessionEnd, hookInput{Reason: "clear"}, "end-clear"},
		{evSessionEnd, hookInput{Reason: "other"}, "end-other"},
		{evSessionEnd, hookInput{}, "end"},
		{evSubagentStop, hookInput{AgentType: "probe"}, "seat-probe"},
		{evSubagentStop, hookInput{}, "seat"},
		{evPreCompact, hookInput{Trigger: "auto"}, "auto"},
		{evPreCompact, hookInput{Trigger: "manual"}, "manual"},
		{evPreCompact, hookInput{}, "compact"},
		// An unlabelled invocation still has to produce something interpretable.
		{"", hookInput{Trigger: "auto"}, "auto"},
		{"", hookInput{}, "unknown"},
	}
	for _, c := range cases {
		got := occasion(c.event, c.in)
		if got != c.want {
			t.Errorf("occasion(%q, %+v) = %q, want %q", c.event, c.in, got, c.want)
		}
		if got == "" || strings.HasSuffix(got, "-") {
			t.Errorf("occasion(%q, %+v) = %q — an unlabelled or dangling occasion is one nobody can interpret later", c.event, c.in, got)
		}
	}
}

// sanitize is the filename-escape defence: agent_type is vendor-supplied and reaches
// this function on its way into a path. Its rune-class boundaries were entirely
// unpinned — a mutation sweep flipped every one of `a`/`z`/`A`/`Z`/`0`/`9` with the
// suite still green, so the character classes were asserted by nothing.
func TestSanitizeKeepsExactlyTheSafeClasses(t *testing.T) {
	// The boundary characters of every kept range, plus the two kept punctuation
	// marks. Each must survive VERBATIM.
	for _, keep := range []string{"a", "z", "A", "Z", "0", "9", "-", "_", "azAZ09-_"} {
		if got := sanitize(keep); got != keep {
			t.Errorf("sanitize(%q) = %q; every character here is safe on every target platform", keep, got)
		}
	}
	// The characters immediately OUTSIDE each range. `/` and `\` are the ones that
	// matter (a value with a separator must not escape the snapshot directory), and
	// the rest guard the range edges: `` ` `` precedes `a`, `{` follows `z`, `@`
	// precedes `A`, `[` follows `Z`, `/` precedes `0`, `:` follows `9`.
	for _, drop := range []string{"/", "\\", "`", "{", "@", "[", ":", ".", " ", "..", "é", "\x00"} {
		got := sanitize(drop)
		if strings.ContainsAny(got, `/\`) {
			t.Errorf("sanitize(%q) = %q; a path separator survived", drop, got)
		}
		for _, r := range got {
			if r != '-' {
				t.Errorf("sanitize(%q) = %q; every unsafe rune must become '-', got %q", drop, got, r)
			}
		}
	}
	// Replacement is per-rune, so length is preserved for ASCII and the result stays
	// a single path component.
	if got := sanitize("general-purpose/../etc"); strings.Contains(got, "..") && strings.Contains(got, "/") {
		t.Errorf("sanitize(%q) = %q; a traversal survived intact", "general-purpose/../etc", got)
	}
}

func TestHeadings(t *testing.T) {
	note := "---\nschema: 3\n---\n## One\ntext\n### Not a section\n## Two\n"
	got := headings(note)
	if len(got) != 2 || got[0] != "One" || got[1] != "Two" {
		t.Fatalf("headings = %v", got)
	}
}
