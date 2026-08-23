package seatprobe

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE GRANT MUST COME FROM THE DEFINITION, AND IT MUST NOT BE EMPTY.
//
// An empty grant is not a restrictive run. It is the approval prompt, met by a seat with nobody at
// the keyboard: measured, the seat's first two calls — both `--seat-id … --help`, exactly the act
// under test — came back "This command requires approval", and it spent the sitting reasoning about
// a record tool it concluded was broken. That transcript and a transcript of a seat that chose not
// to read its help score identically.
func TestEveryDispatchedAgentGrantsTheToolsItsSeatNeeds(t *testing.T) {
	for _, agent := range []string{"red-auditor", "blue-researcher", "lead-judge"} {
		p, err := repotree.Plugin("agents", agent+".md")
		if err != nil {
			t.Fatal(err)
		}
		got, err := GrantedTools(p)
		if err != nil {
			t.Errorf("%s: %v", agent, err)
			continue
		}
		have := map[string]bool{}
		for _, g := range got {
			have[g] = true
		}
		// Bash is the one the record tool needs. Read is how the seat sees the artifact.
		for _, want := range []string{"Bash", "Read"} {
			if !have[want] {
				t.Errorf("%s declares no %s — a seat dispatched under it cannot run the record tool and will report the tool as broken", agent, want)
			}
		}
	}
}

// AND THE MISS IS AN ERROR AT EVERY SEAM, checked by producing each one.
func TestAMissingGrantIsLoud(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"no-frontmatter.md": "just prose\n",
		"unterminated.md":   "---\nname: x\ntools: Bash\n",
		"no-tools.md":       "---\nname: x\ndescription: y\n---\nbody\n",
		"empty-tools.md":    "---\nname: x\ntools:\n---\nbody\n",
	} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := GrantedTools(p)
		if err == nil {
			t.Errorf("%s: returned %v with no error — a silent empty grant dispatches a seat that cannot run anything", name, got)
		}
		if len(got) != 0 {
			t.Errorf("%s: returned %v alongside its error", name, got)
		}
	}
	// And a well-formed one parses to exactly what it declares.
	p := filepath.Join(dir, "ok.md")
	if err := os.WriteFile(p, []byte("---\nname: x\ntools: Read, Bash , WebSearch\nskills: [a]\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := GrantedTools(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, "|") != "Read|Bash|WebSearch" {
		t.Errorf("parsed %v, want [Read Bash WebSearch]", got)
	}
}
