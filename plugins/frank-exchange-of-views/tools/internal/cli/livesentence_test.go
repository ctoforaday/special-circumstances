package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A COMPUTATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED, as a citation is (verifyclaim_test): one
// predicate, record.LiveSentence, refuses both check verbs, and names the anchor by its kind's noun.
// A computation out of the report names the act that took it — a cut of its sentence (event and reason)
// or a retire (event, reason, and the sentence kept). The refusal records nothing.
func TestAReproduceOfAComputationWithNoSentenceIsRefused(t *testing.T) {
	const sentence = "Alpha holds."
	refused := func(t *testing.T, runDir, id string, wants ...string) {
		t.Helper()
		before := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
		_, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
			"--as", "sound", "--reason", "it computes what the sentence says")
		if err == nil {
			t.Fatal("a judgement of a computation the report holds no sentence for was recorded")
		}
		for _, want := range wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("the refusal does not say %q: %v", want, err)
			}
		}
		if got := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); got != before {
			t.Errorf("the refusal left %d reproduce(s) on the record", got-before)
		}
	}
	setup := func(t *testing.T) (string, string) {
		runDir := newRun(t)
		writeReport(t, runDir, "# Findings\n\nAlpha holds. Beta holds. Plain prose here.\n")
		registerBlue(t, runDir)
		return runDir, computeSentence(t, runDir, sentence)
	}
	t.Run("cut", func(t *testing.T) {
		runDir, id := setup(t)
		if _, err := cut(t, runDir, sentence); err != nil {
			t.Fatalf("the cut of %q: %v", sentence, err)
		}
		refused(t, runDir, id, "computation "+id+" is out of the report", "blue's edit at event", "cut its sentence",
			`reason: "the claim cannot stand"`, "in no document")
	})
	t.Run("retired", func(t *testing.T) {
		runDir, id := setup(t)
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "refuted"); err != nil {
			t.Fatalf("retire --anchor %s: %v", id, err)
		}
		if ev := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); ev != 1 {
			t.Fatalf("want one retire on the record, got %d", ev)
		}
		refused(t, runDir, id, "computation "+id+" is out of the report", "blue retired it at event", `reason: "refuted"`,
			`its sentence "`+sentence+`" kept`, "in no document")
	})
	t.Run("live, admitted", func(t *testing.T) {
		runDir, id := setup(t)
		if out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
			"--as", "sound", "--reason", "it computes what the sentence says"); err != nil {
			t.Fatalf("a judgement of a live computation is refused: %v\n%s", err, fmt.Sprint(out))
		}
	})
}
