package record

import (
	"os"
	"regexp"
	"testing"
)

// THE MASS MAPPING'S STAMP LIVES IN TWO LANGUAGES.
//
// Go stamps every telemetry line with MassMappingVersion; debate.js names the same mapping to
// blue in its response prompt. The weights themselves live in one place — the Grade enum's `mass`
// facet, which Go's MASS is read off — so the stamp is the only half the engine still carries,
// and this test is what holds the pair equal.
//
// The parse is deliberately narrow. It reads the declaration by shape and FAILS LOUDLY if it
// cannot find it — a regex that silently matches nothing would report agreement with an empty
// string, which is the plausible zero this whole area exists to remove.

const debateJS = "../../../skills/research-protocol/scripts/debate.js"

var reMassVersion = regexp.MustCompile(`(?m)^const MASS_MAPPING_VERSION\s*=\s*'([^']*)'`)

func readDebateJS(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(debateJS)
	if err != nil {
		t.Fatalf("cannot read the engine at %s: %v\n\nThis test binds the Go mass mapping stamp to "+
			"the JavaScript one. If debate.js moved, POINT THIS TEST AT IT rather than deleting the "+
			"test.", debateJS, err)
	}
	return string(b)
}

func TestMassMappingVersionMatchesTheEngine(t *testing.T) {
	m := reMassVersion.FindStringSubmatch(readDebateJS(t))
	if m == nil {
		t.Fatalf("no `const MASS_MAPPING_VERSION = '...'` in %s — the declaration this test reads "+
			"has changed shape. A no-match here must FAIL, never pass quietly: the whole point is "+
			"that a silent miss reads exactly like agreement.", debateJS)
	}
	if m[1] != MassMappingVersion {
		t.Errorf("mass mapping stamp disagrees: Go says %q, debate.js says %q.\n\n"+
			"These name the SAME mapping and must move together: a split stamp puts a meaningless "+
			"split through every telemetry join keyed on mapping_version.",
			MassMappingVersion, m[1])
	}
}
