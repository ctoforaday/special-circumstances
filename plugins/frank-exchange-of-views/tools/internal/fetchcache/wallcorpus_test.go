package fetchcache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE CORPUS IS REAL BYTES FROM REAL HOSTS, captured through this tool on 2026-09-21 by fetching
// 22 major scholarly sources. Synthetic fixtures cannot settle this question: the earlier tests
// here hand-wrote a skeleton and a prose page, and both passed while three of four live walls
// went through the detector as documents.
//
// EACH FIXTURE IS LABELLED BY WHAT THE PAGE IS, decided by reading it, and the label was fixed
// before the rules were rewritten around them. That order matters — a corpus labelled from the
// gate's own verdicts tests nothing, and the accept side is where that trap lives, because a
// wall announces itself and a document does not.
//
// THE CORPUS IS DELIBERATELY THE SMALL END. Nature's page is 394 KB and discriminates nothing;
// the whole difficulty is that the smallest REAL document pages (a NeurIPS abstract at 1,709
// visible characters, a CVF landing page at 2,447) are nearer in size to a wall than to an
// article. A gate that separates these separates the easy cases for free.
func TestTheDetectorOnRealCapturedPages(t *testing.T) {
	for _, tc := range []struct {
		file string
		wall bool
		what string
	}{
		// Walls. None of these carries a word of the paper its url names.
		{"pmc.html", true, "PubMed Central: an NCBI interstitial titled 'Checking your browser - reCAPTCHA'; NCBI publishes sanctioned machine routes to the same article"},
		{"openreview.html", true, "OpenReview: titled 'Verifying your browser', 288 characters, no paper"},
		{"jstor.html", true, "JSTOR: an F5/Shape page titled 'Client Challenge' asking for JavaScript"},
		{"inspire.html", true, "INSPIRE: the unrendered app shell, whose visible text is 'You need to enable JavaScript to run this app' plus build instructions"},

		// Documents. Each carries its paper's title, authors and abstract as prose.
		{"neurips.html", false, "NeurIPS proceedings page for Attention is All you Need: title, authors, abstract — the smallest real document in the survey at 1,709 characters"},
		{"cvf.html", false, "CVF open-access page for Deep Residual Learning: title, authors, abstract, and the repository's own notice"},
		{"pmlr.html", false, "PMLR page for the CLIP paper: title, authors, abstract, BibTeX"},
		{"acl.html", false, "ACL Anthology page for BERT: title, authors, abstract, and the anthology's navigation"},
		{"arxiv-control.html", false, "arXiv abstract page for Attention Is All You Need — the control; if this is ever refused the gate is broken"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Join("testdata", "walls", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			reason := ShellReason("text/html", body)
			switch {
			case tc.wall && reason == "":
				t.Errorf("A WALL WAS SERVED AS THE DOCUMENT. %s\nA seat would verify a citation against these "+
					"bytes and the miss reads exactly like an honest check.", tc.what)
			case !tc.wall && reason != "":
				t.Errorf("A REAL DOCUMENT WAS REFUSED. %s\nrefusal: %s\nThis is the worse direction: a wall "+
					"wrongly accepted is caught at the leaf when a seat reads it, while a document wrongly "+
					"refused becomes a fabricated absence nothing downstream can question.", tc.what, reason)
			}
		})
	}
}

// AND IT SAYS WHICH WALL, because the next act differs. An app skeleton means nothing will render
// without a browser; a challenge usually means the host publishes a machine route to the same
// text. PubMed Central was reported as the former for months while being the latter.
func TestAChallengeIsNamedAsOneNotAsAnApp(t *testing.T) {
	for _, f := range []string{"pmc.html", "openreview.html", "jstor.html"} {
		body, err := os.ReadFile(filepath.Join("testdata", "walls", f))
		if err != nil {
			t.Fatal(err)
		}
		reason := ShellReason("text/html", body)
		if !strings.Contains(reason, "ACCESS CHALLENGE") {
			t.Errorf("%s: reason = %q, want it named as a challenge", f, reason)
		}
		if !strings.Contains(reason, "never solve the challenge") {
			t.Errorf("%s: the refusal does not say not to solve it: %s", f, reason)
		}
	}
	// INSPIRE is a genuine unrendered app and must NOT be called a challenge — nobody is
	// gatekeeping it, the page simply needs a browser, and telling a seat to look for a
	// sanctioned route would send it hunting for something that does not exist.
	body, _ := os.ReadFile(filepath.Join("testdata", "walls", "inspire.html"))
	if r := ShellReason("text/html", body); strings.Contains(r, "ACCESS CHALLENGE") {
		t.Errorf("an unrendered app is reported as a challenge: %s", r)
	}
}
