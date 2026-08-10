package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// preMotionFixture is a record written by the PRE-#344 binary, carrying every type the motion
// collapse retires. See internal/cli/premotion_fixture_test.go for what produced it — and note
// that generator cannot be re-run once the verbs are gone.
const preMotionFixture = "testdata/pre-motion-run"

// THE FIXTURE MUST KEEP CARRYING WHAT IT WAS MADE FOR.
//
// A compat fixture that quietly stops containing the thing it tests is worse than none: every
// downstream assertion still passes, on a record that no longer exercises anything. This runs
// first and independently, so a truncated or regenerated fixture fails HERE with a clear reason
// rather than as a confusing pass somewhere else.
func TestThePreMotionFixtureStillCarriesTheRetiringVocabulary(t *testing.T) {
	var joined strings.Builder
	entries, err := os.ReadDir(filepath.Join(preMotionFixture, "records"))
	if err != nil {
		t.Fatalf("the pre-motion fixture is missing: %v — it CANNOT be regenerated after #344; recover it from git history", err)
	}
	for _, e := range entries {
		b, rerr := os.ReadFile(filepath.Join(preMotionFixture, "records", e.Name()))
		if rerr != nil {
			t.Fatal(rerr)
		}
		joined.Write(b)
	}
	text := joined.String()
	for _, typ := range []string{"dispute", "dispute-respond", "petition", "petition-rule", "avenue-rule"} {
		if !strings.Contains(text, `"type":"`+typ+`"`) {
			t.Errorf("the fixture no longer carries a %q event — it was made to prove a pre-collapse record still reads, and without this type it proves nothing", typ)
		}
	}
	if !strings.Contains(text, "contests_ruling") {
		t.Error("the fixture lost contests_ruling — the field that becomes `motion direction appeal`")
	}
}

// A PRE-COLLAPSE RECORD STILL REPLAYS, and its exchanges are still visible.
//
// This is the check the whole fixture exists for. Before #344 it passes trivially, which is the
// point: it is written NOW so that the moment the collapse lands, a regression in the dual-read
// fails here instead of rendering an empty section that reads exactly like a run with no
// disputes and no petitions.
func TestAPreCollapseRecordStillReplays(t *testing.T) {
	b, err := BoardState(preMotionFixture)
	if err != nil {
		t.Fatalf("a pre-collapse record must still replay: %v", err)
	}
	if len(b.Events) == 0 {
		t.Fatal("the fixture replayed to zero events — a plausible zero, which is the exact failure this guards")
	}
	counts := map[string]int{}
	for _, e := range b.Events {
		counts[e.Type]++
	}
	// Named individually: a loop over a slice would report "some type is missing" where the
	// useful message is WHICH exchange stopped being visible.
	for typ, what := range map[string]string{
		"dispute":         "blue contesting a grade",
		"dispute-respond": "red answering that contest",
		"petition":        "a seat's ethical/safety/integrity filing",
		"petition-rule":   "the bench's ruling on it",
		"avenue-rule":     "red's ruling on a proposed direction",
	} {
		if counts[typ] == 0 {
			t.Errorf("%s (%s) is not visible in the replayed record — a reader would see an empty section, which reads identically to a run that never had one", typ, what)
		}
	}
}
