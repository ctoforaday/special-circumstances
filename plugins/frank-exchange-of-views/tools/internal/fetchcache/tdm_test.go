package fetchcache

import "testing"

// THE SHAPE ELSEVIER SENDS, from a live capture. Their own policy document, read the same day,
// carries `prohibition: null` and permits tdm:mine for EU DSM Article 3 — scientific research,
// the exception that cannot be reserved. What the meta tag reserves is Article 4, mining for any
// purpose. Neither reaches reading or quotation.
func TestATDMReservationIsReadAndItsPolicyKept(t *testing.T) {
	body := []byte(`<html><head>
<meta name="tdm-reservation" content="1" />
<meta name="tdm-policy" content="https://www.elsevier.com/tdm/tdmrep-policy.json" />
<title>Redirecting</title></head><body>x</body></html>`)
	reserved, policy := TDMReservation("text/html; charset=utf-8", body)
	if !reserved {
		t.Error("a declared reservation was not read")
	}
	if policy != "https://www.elsevier.com/tdm/tdmrep-policy.json" {
		t.Errorf("policy = %q, want the url the source states its terms at", policy)
	}
}

func TestNoReservationIsNotAReservation(t *testing.T) {
	for _, tc := range []struct{ name, ct, body string }{
		{"a page that says nothing", "text/html", "<html><body>a paper</body></html>"},
		// content="0" is the protocol's way of saying MINING IS PERMITTED. Reading it as a
		// reservation would invert the source's own statement.
		{"an explicit zero", "text/html", `<html><head><meta name="tdm-reservation" content="0"></head></html>`},
		{"not html", "application/pdf", `%PDF <meta name="tdm-reservation" content="1">`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if r, _ := TDMReservation(tc.ct, []byte(tc.body)); r {
				t.Error("reported a reservation the source did not make")
			}
		})
	}
}

// THE HEADER IS AN OPINION AND THE MAGIC BYTES ARE THE DOCUMENT.
//
// PubMed Central's open-access bucket — the route NCBI names as sanctioned — serves its PDFs as
// `binary/octet-stream`. Measured: a 983,106-byte body beginning `%PDF-1.4` was cached with that
// type, so the PDF extractor never ran and the OCR path could not fire either, because both ask
// for `application/pdf`. The tool found a paper three indexes had hidden and then could not read
// it.
func TestAGenericContentTypeIsSniffedFromTheBytes(t *testing.T) {
	pdf := []byte("%PDF-1.4\nbody")
	for _, declared := range []string{"binary/octet-stream", "application/octet-stream", "application/force-download", ""} {
		if got := SniffedMediaType(declared, pdf); got != "application/pdf" {
			t.Errorf("SniffedMediaType(%q, %%PDF…) = %q, want application/pdf", declared, got)
		}
	}
	// A SOURCE THAT IS SPECIFIC IS BELIEVED. Saying `text/html` is a claim; saying octet-stream
	// is saying it does not know. Overriding the first would be second-guessing the source about
	// its own document.
	if got := SniffedMediaType("text/html; charset=utf-8", []byte("<html>a page</html>")); got != "text/html" {
		t.Errorf("a specific content type was overridden: %q", got)
	}
	if got := SniffedMediaType("application/pdf", pdf); got != "application/pdf" {
		t.Errorf("a correct content type was changed: %q", got)
	}
	// AND AN UNRECOGNISED BODY KEEPS THE DECLARED TYPE rather than being guessed at.
	if got := SniffedMediaType("binary/octet-stream", []byte("\x00\x01\x02 not a format we know")); got != "binary/octet-stream" {
		t.Errorf("an unknown body was given a type: %q", got)
	}
}
