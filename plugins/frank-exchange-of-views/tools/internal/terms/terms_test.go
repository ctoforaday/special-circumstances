package terms

import (
	"strconv"
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

// A BANNED PHRASE IS BANNED HOWEVER ITS WORDS ARE JOINED. "operator channel" passed the gate as
// "operator-channel" in a seat's constitution (#1209). The scan folds the TEXT once — whitespace
// and hyphen runs to one space, a typographic apostrophe to an ASCII one — and bans and masks both
// run over that, so a mask covers every spelling its ban matches. A possessive is NOT folded away:
// "a sitting's record" is ordinary English for a different thing, so a ban that means the
// possessive says so in its own pattern.
func TestABanMatchesItsPhraseHowEverItIsJoined(t *testing.T) {
	r, err := Parse([]byte(`{"entries":[{"term":"the log","definition":"The log is entries.","seats":["blue"],"bans":[
		{"variant":"operator channel","kind":"GATED","pattern":"operator('s)? channel"},
		{"variant":"method lens","kind":"GATED","pattern":"method[- ]lens"},
		{"variant":"sitting record","kind":"GATED","pattern":"sitting records?\\b",
		 "masks":[{"phrase":"sitting-record repair","reason":"the term"}]}],"collisions":[]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	for text, match := range map[string]string{
		"the operator channel":         "operator channel",
		"the operator-channel entries": "operator-channel",
		"THE OPERATOR\n CHANNEL":       "OPERATOR\n CHANNEL",
		"the operator\u2019s channel":  "operator\u2019s channel",
		"the operator's - channel":     "operator's - channel",
		"a method-lens":                "method-lens",
		"a method lens":                "method lens",
		"the sitting-record":           "sitting-record",
	} {
		hs := r.Scan("x.md", text, nil)
		if len(hs) != 1 {
			t.Errorf("%q: %d hit(s), want 1", text, len(hs))
			continue
		}
		if hs[0].Match != match {
			t.Errorf("%q: reported %q, want the text as written, %q", text, hs[0].Match, match)
		}
	}
	for _, text := range []string{"the operatorchannel", "a method_lens", "a sitting record repair", "a sitting-record repair", "a Sitting-record\nrepair"} {
		if hs := r.Scan("x.md", text, nil); len(hs) != 0 {
			t.Errorf("%q: %d hit(s), want none", text, len(hs))
		}
	}
}

// A HYPHEN IN A PATTERN WOULD NEVER MATCH, because the text it runs over has none left. The
// loader refuses one rather than load a ban that passes everything it names; a hyphen inside a
// character class is a range or a member and is not refused.
func TestTheRegistryRefusesAHyphenTheTextNoLongerHas(t *testing.T) {
	for pat, refused := range map[string]bool{
		"diff-stack":          true,
		"red-merge|merged":    true,
		`round\-0`:            true,
		"diff stack":          false,
		"method[- ]lens":      false,
		"the judge([^a-z]|$)": false,
		"[[:alpha:] ]x":       false,
	} {
		_, err := Parse([]byte(`{"entries":[{"term":"t","definition":"T is t.","seats":["blue"],"bans":[{"variant":"v","kind":"GATED","pattern":` + strconv.Quote(pat) + `}],"collisions":[]}]}`))
		if got := err != nil && strings.Contains(err.Error(), "literal hyphen"); got != refused {
			t.Errorf("pattern %q: refused=%v (%v), want refused=%v", pat, got, err, refused)
		}
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

// definitionHits runs every GATED ban over every definition in the registry. Masks apply — they
// are part of what a ban means — and allows do not, because a definition has no path.
func definitionHits(r *Registry) []string {
	var out []string
	for _, e := range r.Entries {
		for _, h := range r.Scan("", e.Definition, nil) {
			out = append(out, strconv.Quote(e.Term)+" is defined with "+strconv.Quote(h.Match)+
				", the banned variant "+strconv.Quote(h.Variant)+" of "+strconv.Quote(h.Term))
		}
	}
	return out
}

// NO DEFINITION USES A VARIANT THE REGISTRY BANS. A definition is delivered to every seat it names
// and rendered into the vocabulary document, so a retired gloss left in one is handed out under the
// registry's own authority while the gate fails the same words everywhere else.
//
// This runs as a test and not inside Parse: recordpb resolves a value's `(defined_term)` through
// Load, so every record-tool process that renders the log types loads the registry, and scanning
// each definition against each ban there costs every one of them the scan.
func TestNoDefinitionUsesABannedVariant(t *testing.T) {
	r, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if hits := definitionHits(r); len(hits) > 0 {
		t.Errorf("%d definition(s) use a banned variant — reword the definition; it has no path to allow:\n  %s",
			len(hits), strings.Join(hits, "\n  "))
	}

	// The check itself, on the two glosses it exists to keep out and on a masked neighbour.
	planted, err := Parse([]byte(`{"entries":[
		{"term":"friction","definition":"Friction is one type of log entry: the work was impeded and the seat is noting it.","seats":["blue"],"bans":[{"variant":"impeded","kind":"GATED","pattern":"\\bimped(e|ed|es|ing|iment|iments)\\b"}],"collisions":[]},
		{"term":"the log","definition":"The log holds a defect in the tooling, or an impediment.","seats":["blue"],"bans":[],"collisions":[]},
		{"term":"exchange","definition":"An exchange is one round-trip between the parties.","seats":["blue"],"bans":[{"variant":"round","kind":"GATED","pattern":"\\bround\\b","masks":[{"phrase":"round-trip","reason":"a round-trip is a call and its answer"}]}],"collisions":[]}
	]}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := definitionHits(planted); len(got) != 2 {
		t.Errorf("the planted registry has two definitions using a banned variant and one masked neighbour; the check reported %d:\n  %s",
			len(got), strings.Join(got, "\n  "))
	}
}
