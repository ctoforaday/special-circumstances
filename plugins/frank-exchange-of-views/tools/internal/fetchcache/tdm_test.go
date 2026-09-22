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
