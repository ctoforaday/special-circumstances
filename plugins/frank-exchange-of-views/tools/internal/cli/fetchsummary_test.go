package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
)

// THE WARNING IS TOLD ABOUT THE BACKEND THAT ANSWERED. The archive paragraph used to print over
// every recovered fetch, so a seat holding an arXiv PDF taken live from arxiv.org read that the
// live source had refused this container and that its bytes were a third party's snapshot —
// neither true. A warning that misnames what happened spends the seat's attention on the wrong
// risk, and it did so on the majority of recoveries, because only one of four backends is the
// archive.
func TestOnlyAnArchiveAnswerIsCalledASnapshot(t *testing.T) {
	const snapshot = "ARCHIVE SNAPSHOT"
	arx := fetchSummary{URL: "https://arxiv.org/abs/hep-th/9711200", Bytes: 246327,
		Backend: fetchcache.ViaArxiv, RetrievedVia: "arXiv hep-th/9711200 (…/pdf/…)", TextRetrieved: true}
	got := arx.render()
	if strings.Contains(got, snapshot) {
		t.Errorf("an arXiv PDF is reported as an archive snapshot:\n%s", got)
	}
	// AND THE PROVENANCE DUTY SURVIVES: these bytes still did not come from the url asked for.
	if !strings.Contains(got, "DID NOT COME FROM THE URL YOU ASKED FOR") {
		t.Errorf("gating the archive warning dropped the provenance warning entirely:\n%s", got)
	}

	arc := fetchSummary{URL: "https://ex.org/p", Bytes: 10,
		Backend: fetchcache.ViaArchive, RetrievedVia: "archive.org capture of 2019-05-20", TextRetrieved: true}
	if !strings.Contains(arc.render(), snapshot) {
		t.Errorf("an archive capture is no longer called a snapshot:\n%s", arc.render())
	}

	// A PLAIN LIVE FETCH GETS NEITHER. Nothing was recovered, so there is no provenance to warn
	// about and a warning here would be noise on the common path.
	live := fetchSummary{URL: "https://ex.org/p", Bytes: 10, TextRetrieved: true}
	if out := live.render(); strings.Contains(out, snapshot) || strings.Contains(out, "DID NOT COME FROM") {
		t.Errorf("a live fetch carries a provenance warning:\n%s", out)
	}
}
