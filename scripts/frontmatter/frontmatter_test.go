package main

import (
	"strings"
	"testing"
)

// THE THREE CASES #157 §2 MEASURED THE .mjs VERSION PASSING.
//
// Each is a genuine YAML parse failure, which means the loader drops every field in the
// block and loads the file anyway with empty metadata. The old escape hatch waved them
// through because the value STARTED with a quote or a bracket, without checking that it
// terminated — "already quoted, flow-style, or a block scalar" was assumed, not verified.
func TestCatchesTheParseFailuresTheHeuristicMissed(t *testing.T) {
	cases := []struct{ name, block string }{
		{"unterminated flow sequence", "name: probe\ntools: [Read, Write\n"},
		{"unclosed double quote", "name: probe\ndescription: \"unclosed quote\n"},
		{"tab-indented list item", "name: probe\ntools:\n\t- Read\n"},
		{"unterminated flow mapping", "name: probe\nmeta: {a: 1\n"},
		{"unclosed single quote", "name: probe\ndescription: 'unclosed\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			problems, scanned := check("x.md", "---\n"+c.block+"---\n")
			if !scanned {
				t.Fatal("the block should have been scanned")
			}
			if len(problems) == 0 {
				t.Fatalf("no problem reported — this is the silent whole-block drop the guard exists to catch:\n%s", c.block)
			}
			if !strings.Contains(string(problems[0]), "drops EVERY field") {
				t.Errorf("the message must name the consequence, got %q", problems[0])
			}
		})
	}
}

// The original case this guard was written for: an unquoted description containing ": "
// ends the scalar early. probe.md shipped broken this way.
func TestCatchesTheUnquotedColonThatBrokeProbe(t *testing.T) {
	problems, _ := check("probe.md", "---\nname: probe\ndescription: does things skills: Read and more\n---\n")
	if len(problems) == 0 {
		t.Fatal("an unquoted value containing \": \" must be reported")
	}
}

// What no parser can see: valid YAML that does not mean what it says. This is why the
// guard is not simply yaml.Unmarshal in a loop.
func TestCatchesCommentTruncationWhichIsValidYAML(t *testing.T) {
	problems, scanned := check("a.md", "---\nname: a\ndescription: does things # oops\n---\n")
	if !scanned {
		t.Fatal("valid YAML must still be scanned for truncation")
	}
	if len(problems) != 1 {
		t.Fatalf("expected exactly one truncation finding, got %v", problems)
	}
	if !strings.Contains(string(problems[0]), "loads truncated") {
		t.Errorf("message must explain the truncation, got %q", problems[0])
	}
	if !strings.Contains(string(problems[0]), "a.md:3") {
		t.Errorf("finding must cite the line, got %q", problems[0])
	}

	// Quoted, the "#" is literal and there is nothing to warn about. Flagging it would
	// train people to ignore this guard.
	for _, quoted := range []string{
		"---\nname: a\ndescription: \"does things # fine\"\n---\n",
		"---\nname: a\ndescription: 'does things # fine'\n---\n",
	} {
		if p, _ := check("a.md", quoted); len(p) != 0 {
			t.Errorf("a quoted value containing \" #\" must not be flagged: %v", p)
		}
	}
	// A "#" with no preceding space is not a comment in YAML.
	if p, _ := check("a.md", "---\nname: a\ntag: red#5\n---\n"); len(p) != 0 {
		t.Errorf("`red#5` is not a comment and must not be flagged: %v", p)
	}
}

func TestValidBlocksPass(t *testing.T) {
	valid := []string{
		"---\nname: probe\ndescription: \"does things skills: Read\"\ntools: [Read, Write]\n---\n",
		"---\nname: a\ntools:\n  - Read\n  - Write\n---\n",
		"---\nname: a\ndescription: >\n  a folded block\n  over two lines\n---\n",
		"---\nname: a\n# a comment line\nmodel: opus\n---\n",
		"---\nname: a\n---\n",
	}
	for _, block := range valid {
		problems, scanned := check("a.md", block)
		if !scanned {
			t.Errorf("block should be scanned:\n%s", block)
		}
		if len(problems) != 0 {
			t.Errorf("valid block reported %v:\n%s", problems, block)
		}
	}
}

func TestFrontmatterBoundaries(t *testing.T) {
	// No frontmatter at all is fine and is NOT counted as scanned.
	if p, scanned := check("a.md", "# just a heading\n"); len(p) != 0 || scanned {
		t.Errorf("a file without frontmatter must be silently skipped: problems=%v scanned=%v", p, scanned)
	}
	// Opens and never closes: the loader reads the entire file as YAML.
	p, scanned := check("a.md", "---\nname: a\nbody text with no closing marker\n")
	if len(p) != 1 || scanned {
		t.Fatalf("unclosed frontmatter: problems=%v scanned=%v", p, scanned)
	}
	if !strings.Contains(string(p[0]), "never closes") {
		t.Errorf("message must name the cause, got %q", p[0])
	}
	// A body containing "---" later must not be mistaken for the opener.
	if p, scanned := check("a.md", "text\n---\nnot frontmatter: true\n---\n"); len(p) != 0 || scanned {
		t.Errorf("frontmatter must be at the very start: problems=%v scanned=%v", p, scanned)
	}
}
