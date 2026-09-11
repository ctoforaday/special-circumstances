package capture

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// The envelope claim is a cross-check on the record. A journal from before the disposition keys
// holds its claim under `resolutions`, so the claim reads as 0 — which, beside a record holding
// rulings, prints as a divergence the seats never made. The harvest says the claim was not read.
func TestAPreRenameClaimIsNotMeasuredRatherThanZero(t *testing.T) {
	run := runtest.New(t, recordtest.TmpRun(t))
	law := filepath.Join(t.TempDir(), "law")

	old := []map[string]any{{"resolutions": []any{map[string]any{"gap_id": "G1", "resolution": "repaired"}}}}
	r := HarvestPrecedents(run, old, law, nil)
	if !strings.Contains(r.EnvelopeUnmeasured, "not measured: this transcript's envelopes predate the disposition keys (resolutions)") {
		t.Errorf("EnvelopeUnmeasured = %q on a pre-rename journal", r.EnvelopeUnmeasured)
	}

	current := []map[string]any{{"dispositions": []any{map[string]any{"gap_id": "G1", "disposition": "repaired"}}}}
	if r := HarvestPrecedents(run, current, law, nil); r.EnvelopeUnmeasured != "" || r.EnvelopeClaimed != 1 {
		t.Errorf("a current journal: unmeasured %q, claimed %d — want none and 1", r.EnvelopeUnmeasured, r.EnvelopeClaimed)
	}
}
