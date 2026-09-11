package terms

import (
	"strings"
	"testing"
)

func TestTheEmbeddedRegistryLoads(t *testing.T) {
	r, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	gated := 0
	for _, e := range r.Entries {
		for _, b := range e.Bans {
			if b.Kind == Gated {
				gated++
			}
		}
	}
	if gated == 0 {
		t.Fatal("the registry gates no variant — the vocabulary gate would pass every surface")
	}
}

// Each case is a registry the loader must refuse, and the words the refusal must carry.
func TestTheRegistryRefusesWhatWouldMislead(t *testing.T) {
	entry := func(body string) string {
		return `{"entries":[{"term":"the report","definition":"The report is the prose.","seats":["blue"],"bans":[` + body + `],"collisions":[]}]}`
	}
	for _, c := range []struct {
		name, json, want string
	}{
		{"no entries", `{"entries":[]}`, "no entries"},
		{"no definition", `{"entries":[{"term":"x","definition":"","seats":["blue"],"bans":[],"collisions":[]}]}`, "no definition"},
		{"two sentences", `{"entries":[{"term":"x","definition":"One. Two.","seats":["blue"],"bans":[],"collisions":[]}]}`, "ONE sentence"},
		{"no seats", `{"entries":[{"term":"x","definition":"X is x.","seats":[],"bans":[],"collisions":[]}]}`, "no seat"},
		{"duplicate term", `{"entries":[{"term":"x","definition":"X is x.","seats":["blue"],"bans":[],"collisions":[]},{"term":"x","definition":"X is x.","seats":["blue"],"bans":[],"collisions":[]}]}`, "entered twice"},
		{"gated without pattern", entry(`{"variant":"living report","kind":"GATED"}`), "no pattern"},
		{"pattern not RE2", entry(`{"variant":"v","kind":"GATED","pattern":"(?=x)"}`), "not RE2"},
		{"unknown kind", entry(`{"variant":"v","kind":"SOFT","pattern":"v"}`), "kind"},
		{"allow without reason", entry(`{"variant":"v","kind":"GATED","pattern":"v","allow":[{"path":"README.md","reason":""}]}`), "without a reason"},
		{"allow without path", entry(`{"variant":"v","kind":"GATED","pattern":"v","allow":[{"path":"","reason":"r"}]}`), "names no path"},
		{"mask without reason", entry(`{"variant":"v","kind":"GATED","pattern":"round","masks":[{"phrase":"round-trip","reason":""}]}`), "mask missing"},
		{"mask the pattern cannot match", entry(`{"variant":"v","kind":"GATED","pattern":"round","masks":[{"phrase":"trip","reason":"r"}]}`), "blanks nothing"},
		{"registry-only with a pattern", entry(`{"variant":"v","kind":"REGISTRY-ONLY","pattern":"v"}`), "REGISTRY-ONLY"},
		{"misspelt field", entry(`{"variant":"v","kind":"GATED","pattren":"v"}`), "unknown field"},
	} {
		_, err := Parse([]byte(c.json))
		if err == nil {
			t.Errorf("%s: the registry loaded; it must be refused", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: refused with %q, which does not say %q", c.name, err, c.want)
		}
	}
}

func TestScanFoldsWhitespaceMasksAndAllows(t *testing.T) {
	r, err := Parse([]byte(`{"entries":[{"term":"epoch","definition":"An epoch is a cycle.","seats":["blue"],"bans":[
		{"variant":"round","kind":"GATED","pattern":"\\brounds?\\b","masks":[{"phrase":"round-trip","reason":"not a unit"}],
		 "allow":[{"path":"a/**/old.md","reason":"history"}]},
		{"variant":"living report","kind":"GATED","pattern":"living report"}],"collisions":[]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	u := NewUsage()
	hits := r.Scan("x.md", "one round-trip\nthe living\n   report\nnext Round", u)
	var got []string
	for _, h := range hits {
		got = append(got, h.Variant+"@"+string(rune('0'+h.Line)))
	}
	if want := "round@4 living report@2"; strings.Join(got, " ") != want {
		t.Errorf("hits = %q, want %q", strings.Join(got, " "), want)
	}
	if hs := r.Scan("a/b/c/old.md", "a round", u); len(hs) != 1 || hs[0].AllowedBy != "a/**/old.md" {
		t.Errorf("the allow did not cover a/b/c/old.md: %+v", hs)
	}
	if st := r.Stale(u); len(st) != 0 {
		t.Errorf("a mask and an allow both did work, yet Stale says %q", st)
	}
	if st := r.Stale(NewUsage()); len(st) != 2 {
		t.Errorf("with nothing scanned, the mask and the allow are both stale; Stale says %q", st)
	}
}

func TestForSeatDeliversOnlyThatSeatsEntries(t *testing.T) {
	r, err := Parse([]byte(`{"entries":[
		{"term":"a","definition":"A is a.","seats":["blue"],"bans":[],"collisions":[]},
		{"term":"b","definition":"B is b.","seats":["lens","blue"],"bans":[],"collisions":[]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if n := len(r.ForSeat("lens")); n != 1 {
		t.Errorf("lens gets %d entries, want 1", n)
	}
	if n := len(r.ForSeat("blue")); n != 2 {
		t.Errorf("blue gets %d entries, want 2", n)
	}
	if got := strings.Join(r.Seats(), ","); got != "blue,lens" {
		t.Errorf("Seats() = %q", got)
	}
}
