package scorecard

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Blue's position prose is written once and kept, so a run recorded while the deferring
// disposition was spelled "carried" still says so. Direction uptake reads that prose: a pattern
// that knew only the current word would score such a run's uptake as zero, silently.
func TestDirectionUptakeReadsBothSpellingsOfTheDeferral(t *testing.T) {
	for _, prose := range []string{"the gap was remanded; acting on the stated research", "the gap was carried; acting on the stated research"} {
		dj := record.DebateJSON{Epochs: []record.DebateEpochJSON{
			{Lead: []record.DebateOpinionJSON{{GapID: "G1"}}, Blue: []string{prose}},
		}}
		if lead, blue := ComputeDirectionUptake(dj); lead != 1 || blue != 1 {
			t.Errorf("%q: uptake %d/%d, want 1/1", prose, blue, lead)
		}
	}
}

func rulingsResult(words ...string) []map[string]any {
	var rs []any
	for _, w := range words {
		rs = append(rs, map[string]any{"gap_id": "G1", "disposition": w})
	}
	return []map[string]any{{"dispositions": rs}}
}

func TestRemandedShareCountsTheCurrentWord(t *testing.T) {
	r := rowByMetric(benchRows(rulingsResult("remanded", "repaired"), nil), "remanded_share")
	if r == nil {
		t.Fatal("no remanded_share row")
	}
	if v, ok := r.Value.(float64); !ok || v != 0.5 {
		t.Errorf("remanded_share = %v (%s), want 0.5", r.Value, r.Note)
	}
}

// A disposition word this binary does not know makes the row UNMEASURED. A transcript is never
// migrated, so an old one holds "carried"; counted as not-remanded it would put the bench's
// deferrals in the denominator only, and a bench that deferred everything would score 0.
func TestAnUnknownDispositionIsNotMeasuredRatherThanZero(t *testing.T) {
	r := rowByMetric(benchRows(rulingsResult("carried", "carried", "repaired"), nil), "remanded_share")
	if r == nil {
		t.Fatal("no remanded_share row")
	}
	if r.Value != nil {
		t.Errorf("remanded_share has a value %v over a word outside the vocabulary; want none", r.Value)
	}
	if !strings.Contains(r.Note, `not measured: disposition "carried" is not in this binary's vocabulary`) {
		t.Errorf("the row does not say what it could not read: %q", r.Note)
	}
	if strings.Count(r.Note, `"carried"`) != 1 {
		t.Errorf("each unknown word is named once: %q", r.Note)
	}
}
