package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A COMPUTATION THE REPORT HOLDS NO SENTENCE FOR IS REFUSED, as a citation is (verifyclaim_test): one
// predicate, record.LiveSentence, refuses both check verbs, and names the anchor by its kind's noun.
// A bare computation is an edit's leftover; a retired one names its retire — event, reason, and
// whether its sentence was kept. The refusal records nothing.
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
	t.Run("bare", func(t *testing.T) {
		runDir, id := setup(t)
		gutToAnchor(t, runDir, sentence, anchor.Token(id))
		refused(t, runDir, id, "computation "+id, "stands bare", "nothing to check")
	})
	t.Run("retired", func(t *testing.T) {
		runDir, id := setup(t)
		gutToAnchor(t, runDir, sentence, anchor.Token(id))
		retireClaim(t, runDir, sentence)
		ev := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
		if ev != 1 {
			t.Fatalf("want one retire on the record, got %d", ev)
		}
		refused(t, runDir, id, "computation "+id+" is retired", "at event", `reason: "refuted"`, "with its claim", "in no document")
	})
	t.Run("live, admitted", func(t *testing.T) {
		runDir, id := setup(t)
		if out, err := run(t, "reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id,
			"--as", "sound", "--reason", "it computes what the sentence says"); err != nil {
			t.Fatalf("a judgement of a live computation is refused: %v\n%s", err, fmt.Sprint(out))
		}
	})
}
