package claims

import (
	"strings"
	"testing"
)

// A suffixed loop heading IS the loop heading. Measured on the development box:
// 7 of 58 real loop headings carry a suffix, and each of those notes was
// reported "carries no validation loop" with exit 0.
func TestASuffixedLoopHeadingIsTheLoop(t *testing.T) {
	for _, h := range []string{
		"## Validation loop",
		"## Validation loop (all last run at ed3bc3ae)",
		"## validation LOOP — per branch",
		"## Validation loop        ← load-bearing",
		"### Validation loop:",
	} {
		cs, err := Parse(strings.NewReader(h+"\n1. `go test ./...` · re-armed by: a/\n"), "N.md")
		if err != nil {
			t.Errorf("%q: %v", h, err)
			continue
		}
		if len(cs) != 1 {
			t.Errorf("%q: %d claims, want 1", h, len(cs))
		}
	}
}

// A heading that spells the loop another way fails LOUDLY. Skipping it would
// report a note with no loop when the note has one this reader cannot see.
func TestANearMissLoopHeadingIsRefused(t *testing.T) {
	for _, h := range []string{
		"## Validation Loops",
		"## validation-loop",
		"## ValidationLoop",
		"## Validation  loop",
		"Validation loops\n----------------",
	} {
		_, err := Parse(strings.NewReader("intro\n\n"+h+"\n1. `go test ./...` · re-armed by: a/\n"), "N.md")
		if err == nil {
			t.Errorf("%q was read past as no loop", h)
			continue
		}
		if !strings.Contains(err.Error(), "N.md:3") || !strings.Contains(err.Error(), "Validation loop") {
			t.Errorf("%q: the refusal does not say where or what the heading must be: %v", h, err)
		}
	}
	// Other headings that merely mention the words are not near misses.
	for _, h := range []string{"## Committed; running the validation loop", "## Validator notes"} {
		if _, err := Parse(strings.NewReader(h+"\n1. x\n"), "N.md"); err != nil {
			t.Errorf("%q was refused: %v", h, err)
		}
	}
}

func TestTwoLoopHeadingsAreRefused(t *testing.T) {
	note := "## Validation loop\n1. `a b`\n## Validation loop (again)\n1. `c d`\n"
	if _, err := Parse(strings.NewReader(note), "N.md"); err == nil || !strings.Contains(err.Error(), "N.md:3") {
		t.Errorf("two loop headings: err = %v", err)
	}
}

// A deeper heading inside the loop does not end it; a heading of the loop's own
// rank does. The line pattern this replaces ended the loop at ANY heading, so a
// `### sub` dropped every later entry once one had parsed.
func TestTheLoopEndsAtAHeadingOfItsOwnRank(t *testing.T) {
	note := "## Validation loop\n1. `go test ./a`\n\n### fuzz\n\n2. `go test ./b`\n\n## Next intended steps\n1. not a claim\n"
	cs, err := Parse(strings.NewReader(note), "N.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 2 || cs[1].Cmd != "go test ./b" || cs[1].Line != 6 {
		t.Fatalf("claims = %+v", cs)
	}
}

// "1)" opens a list item, which prosthetic-conscience's hooks do not read as a
// check: refused, not adjudicated as if both readers agreed.
func TestAParenthesisedLabelIsRefused(t *testing.T) {
	_, err := Parse(strings.NewReader("## Validation loop\n1) `go test ./...`\n"), "N.md")
	if err == nil || !strings.Contains(err.Error(), "N.md:2") {
		t.Errorf("err = %v", err)
	}
}

// The frontmatter's written_at reaches every claim, and the line numbers still
// count the frontmatter's lines.
func TestWrittenAtReachesEveryClaimAndLinesCountTheFrontmatter(t *testing.T) {
	note := "---\nschema: 3\nwritten_at: 2026-10-18T10:00:00Z     # read from a clock\nstatus: done\n---\n## Validation loop\n1. `go test ./a`\n2. `go test ./b`\n"
	cs, err := Parse(strings.NewReader(note), "N.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 2 || cs[0].Line != 7 || cs[1].Line != 8 {
		t.Fatalf("claims = %+v", cs)
	}
	for _, c := range cs {
		if c.NoteWrittenAt != "2026-10-18T10:00:00Z     # read from a clock" {
			t.Errorf("written_at = %q", c.NoteWrittenAt)
		}
	}
}

// A note written with CRLF line endings reads the same as one with LF.
func TestCRLFNotesReadTheSame(t *testing.T) {
	note := "---\r\nwritten_at: 2026-10-18T10:00:00Z\r\n---\r\n## Validation loop (x)\r\n1. `go test ./a` · re-armed by: a/\r\n   last run: pass 2026-10-18T09:00Z\r\n"
	cs, err := Parse(strings.NewReader(note), "N.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 1 || cs[0].Line != 5 || cs[0].Surface != "a/" || cs[0].At != "2026-10-18T09:00Z" || cs[0].Outcome != "pass 2026-10-18T09:00Z" {
		t.Fatalf("claims = %+v", cs)
	}
}

// The trigger surface ends at "·", as prosthetic-conscience's reader ends it, so
// both readers record the same surface for one check.
func TestTheSurfaceEndsAtTheClauseSeparator(t *testing.T) {
	cs, err := Parse(strings.NewReader("## Validation loop\n1. `go test ./a` · re-armed by: a/ · owner: me\n"), "N.md")
	if err != nil {
		t.Fatal(err)
	}
	if cs[0].Surface != "a/" {
		t.Errorf("surface = %q", cs[0].Surface)
	}
}
