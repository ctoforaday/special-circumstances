package debatejs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// EVERY DISPATCH NAMES THE BUILDER THAT WROTE ITS PROMPT, and the builders are counted from the
// script. Site comes from the call stack of a drive and Sites from the syntax tree of the same
// wrapped source; the two are held together here on the shipped script, and each listed line is
// read back out of the file so a line that drifted by one fails on the text it points at.
func TestEveryDispatchSiteIsAnAgentCallInTheScript(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	script := ScriptPath(filepath.Join(wd, "..", ".."))
	sites, err := Sites(script)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	for _, l := range sites {
		if l < 1 || l > len(lines) || !strings.Contains(lines[l-1], "agent(") {
			t.Errorf("Sites lists line %d, which holds no agent( call", l)
		}
	}
	// The count the syntax tree finds is the count a reader of the file finds: a line of code — not
	// a comment — holding `agent(`. A call the walk does not descend into is missing from Sites and
	// present here.
	var inCode []int
	for i, l := range lines {
		if code, _, _ := strings.Cut(l, "//"); strings.Contains(code, "agent(") {
			inCode = append(inCode, i+1)
		}
	}
	if !slices.Equal(sites, inCode) {
		t.Errorf("Sites = %v; the lines of code holding agent( are %v", sites, inCode)
	}

	docketBench := map[string]any{"seat_id": "judge", "gap_ids": []any{"G1"}}
	ds, _ := drivePlans(t, []map[string]any{planWith([]any{lensParty, map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}}, docketBench}, []any{})}, nil)
	reached := map[int]bool{}
	for _, d := range ds {
		if !slices.Contains(sites, d.Site) {
			t.Errorf("%s was dispatched from line %d, which Sites (%v) does not list", d.Label, d.Site, sites)
		}
		reached[d.Site] = true
	}
	// One site per kind of sitting this drive holds: frontier, lane, synthesis, chair, lens, blue,
	// bench, assembly. Fewer means two builders were read as one line.
	if len(reached) < 8 {
		t.Errorf("a drive seating a lens, blue and the docket bench reached %d site(s) of %v, want at least 8", len(reached), sites)
	}
}

func TestSitesRefusesAScriptThatDispatchesNobody(t *testing.T) {
	p := filepath.Join(t.TempDir(), "debate.js")
	if err := os.WriteFile(p, []byte("export const meta = {}\nconst x = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := Sites(p); err == nil {
		t.Errorf("Sites = %v for a script with no agent() call, want an error — an empty list reads as every site covered", got)
	}
}
