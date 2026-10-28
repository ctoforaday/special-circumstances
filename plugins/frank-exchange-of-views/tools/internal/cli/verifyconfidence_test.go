package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING'S --json CARRIES THE CONFIDENCE IT RECORDS, read off the stored event. It echoed the
// outcome and the bridge and not the confidence, so a lens checking what it had recorded read the
// record again to find it (universe-m34). `medium` is no default, so an echo of anything but the
// stored value fails here.
func TestAReadingsJSONCarriesTheConfidenceItRecords(t *testing.T) {
	t.Run("verify", func(t *testing.T) {
		runDir, anchor := citedClause(t)
		r := jsonResult(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
			"--as", "supports", "--confidence", "medium", "--reason", "read at the leaf")
		if stored := recordpb.Word(firstVerify(t, runDir).GetConfidence()); stored != "medium" || r["confidence"] != stored {
			t.Errorf("verify --json confidence = %v, the record holds %q; want medium in both", r["confidence"], stored)
		}
	})
	t.Run("corroborate and its retry", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, fermatReport)
		cachedCopy(t, runDir, "", "")
		registerLensOnce(t, runDir)
		args := []string{"corroborate", "--run", runDir, "--seat-id", lensSeat, "--url", fermatURL, "--title", "Fermat pseudoprime",
			"--quote", fermatClaim, "--as", "supports", "--confidence", "medium", "--reason", "read at the leaf"}
		for _, pass := range []string{"first", "retry"} {
			r := jsonResult(t, args...)
			if r["confidence"] != "medium" {
				t.Errorf("%s corroborate --json confidence = %v, want medium", pass, r["confidence"])
			}
			if idem, _ := r["idempotent"].(bool); idem != (pass == "retry") {
				t.Errorf("%s corroborate --json idempotent = %v", pass, r["idempotent"])
			}
		}
	})
}
