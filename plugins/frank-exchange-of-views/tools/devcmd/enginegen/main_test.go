package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

func sample() source {
	return source{
		seats: map[string]seatclass.Seat{
			"judge":             {Role: "bench", Base: "judge", AgentType: "p:lead-judge"},
			"red-lens-evidence": {Role: "lens", Base: "red-lens", Area: "evidence", AgentType: "p:red-lens-evidence"},
			"operator":          {Role: seatclass.OperatorRole},
		},
		lane:       seatclass.Seat{Role: "blue", Base: "blue-lane", AgentType: "p:blue-researcher"},
		lanePrefix: "blue-lane-",
		classOf: func(base string) string {
			return map[string]string{"judge": "judgment", "red-lens": "bulk", "blue-lane": "bulk"}[base]
		},
		areas:       []string{"evidence"},
		occasions:   []string{"docket", "petition"},
		massVersion: "v2",
	}
}

// The block states each dispatched seat's row, sorted, and leaves the operator out: nothing
// dispatches it.
func TestTheBlockStatesEveryDispatchedSeatAndNoOther(t *testing.T) {
	got, err := render(sample())
	if err != nil {
		t.Fatal(err)
	}
	want := `const SEATS = {
  'judge': { role: 'bench', tier: 'judgment', agentType: 'p:lead-judge' },
  'red-lens-evidence': { role: 'lens', tier: 'bulk', agentType: 'p:red-lens-evidence', area: 'evidence' },
}
const LANE = { prefix: 'blue-lane-', role: 'blue', tier: 'bulk', agentType: 'p:blue-researcher' }
const RED_AREAS = ['evidence']
const OCCASION = { docket: 'docket', petition: 'petition' }
const MASS_MAPPING_VERSION = 'v2'
`
	if string(got) != want {
		t.Errorf("the block renders as:\n%s\nwant:\n%s", got, want)
	}
}

// Each roster defect the engine could not dispatch is refused at generation, naming it.
func TestARosterTheEngineCouldNotDispatchIsRefused(t *testing.T) {
	cases := []struct {
		name  string
		spoil func(*source)
		want  string
	}{
		{"a seat with no configuration", func(s *source) { s.seats["judge"] = seatclass.Seat{Role: "bench", Base: "judge"} }, "judge under no agent configuration"},
		{"a seat with no tier class", func(s *source) {
			s.seats["judge"] = seatclass.Seat{Role: "bench", Base: "bench", AgentType: "p:lead-judge"}
		}, "no tier class"},
		{"a lane with no configuration", func(s *source) { s.lane.AgentType = "" }, "a lane under no agent configuration"},
		{"no lens area", func(s *source) { s.areas = nil }, "declares no lens area"},
		{"an area with no seat", func(s *source) { s.areas = append(s.areas, "logic") }, "logic has no seat"},
		{"a lens seat for an undeclared area", func(s *source) {
			s.seats["red-lens-banana"] = seatclass.Seat{Role: "lens", Base: "red-lens", Area: "banana", AgentType: "p:red-lens-banana"}
		}, "red-lens-banana for the area banana"},
		{"no occasion", func(s *source) { s.occasions = nil }, "holds no word"},
		{"an occasion that is no key", func(s *source) { s.occasions = []string{"re-hearing"} }, `"re-hearing" is not a word`},
		{"no mass mapping version", func(s *source) { s.massVersion = "" }, "carries no version"},
		{"a value a script string cannot carry", func(s *source) { s.lane.AgentType = "p:blue'researcher" }, "cannot carry"},
	}
	for _, c := range cases {
		s := sample()
		c.spoil(&s)
		out, err := render(s)
		if err == nil {
			t.Errorf("%s: rendered\n%s", c.name, out)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: refused with %q, which does not name it (%q)", c.name, err, c.want)
		}
	}
}

// splice rewrites what stands between the markers and nothing else, and refuses a script that
// lacks either marker.
func TestOnlyTheRegionBetweenTheMarkersIsWritten(t *testing.T) {
	script := "const a = 1\n" + beginMark + "\nconst STALE = 0\n" + endMark + "\nconst b = 2\n"
	got, err := splice([]byte(script), []byte("const FRESH = 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := "const a = 1\n" + beginMark + "\nconst FRESH = 1\n" + endMark + "\nconst b = 2\n"
	if string(got) != want {
		t.Errorf("spliced:\n%s\nwant:\n%s", got, want)
	}
	if _, err := splice([]byte("const a = 1\n"), nil); err == nil {
		t.Error("a script with no opening marker was written")
	}
	if _, err := splice([]byte(beginMark+"\nconst a = 1\n"), nil); err == nil {
		t.Error("a script whose block never closes was written")
	}
}
