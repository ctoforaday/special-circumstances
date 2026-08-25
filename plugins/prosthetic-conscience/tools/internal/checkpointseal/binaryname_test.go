package checkpointseal

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// NO SOURCE LINE MAY SPELL THE RETIRED NAME.
//
// #201 step 3 split one merged sealer into three shims and retired "sc-checkpoint-seal".
// The -version line was corrected then — and eleven other lines were not, so every
// operator- and agent-facing message from all three binaries still announced a program
// that does not exist. Somebody grepping their logs for sc-precompact found nothing, which
// is the same nothing they would get from a hook that never ran.
//
// binaryFor's own fallback is the single legitimate use: it is the answer for an
// unlabelled invocation, which seals silently anyway.
func TestNoMessageSpellsTheRetiredName(t *testing.T) {
	const retired = `"sc-checkpoint-seal`
	for _, f := range []string{"main.go", "sealrow.go"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(b), "\n") {
			if !strings.Contains(line, retired) {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue // prose explaining the defect is not the defect
			}
			if strings.Contains(line, "return "+retired) {
				continue // binaryFor's fallback for an unlabelled invocation
			}
			t.Errorf("%s:%d names the retired binary in a message:\n\t%s", f, i+1, strings.TrimSpace(line))
		}
	}
}

// The three shims must each answer with their OWN name — the property #201 step 3 broke and
// hookmain now protects. Asserted here too because this is the package that got it wrong,
// and a test in hookmain proves the helper works, not that this caller uses it.
func TestEachShimIsNamedForItsEvent(t *testing.T) {
	for event, want := range map[string]string{
		"PreCompact": "sc-precompact", "SessionEnd": "sc-sessionend",
		"SubagentStop": "sc-subagentstop", "": "sc-checkpoint-seal",
	} {
		if got := binaryFor(event); got != want {
			t.Errorf("binaryFor(%q) = %q, want %q", event, got, want)
		}
	}
}

// The agent-facing drift advisories carry the invoking binary's name, so the reader can
// find the program that produced them.
func TestTheDriftAdvisoryNamesTheInvokingBinary(t *testing.T) {
	_, within := tree(t, "tools")
	got := drift("sc-subagentstop", "# a note with no validation loop\n", []string{"tools/x.go"}, within)
	if got == "" {
		t.Fatal("no advisory for a note with no validation loop")
	}
	if !strings.HasPrefix(got, "sc-subagentstop: ") {
		t.Errorf("advisory does not name its binary:\n\t%s", got)
	}
	if regexp.MustCompile(`sc-checkpoint-seal`).MatchString(got) {
		t.Errorf("advisory names the retired binary:\n\t%s", got)
	}
}
