package secrets

import (
	"math/rand"
	"strings"
	"testing"
)

// The placeholders a person types into a command or a doc are skipped by the rows gitleaks
// gates by entropy, through Scan and through ScanPayload.
func TestPlaceholdersAreSkipped(t *testing.T) {
	for _, c := range []struct{ id, s string }{
		{"gitlab-pat", "glpat-" + strings.Repeat("x", 20)},
		{"stripe-access-token", "sk_live_" + strings.Repeat("0", 24)},
		{"npm-access-token", "npm_" + strings.Repeat("X", 36)},
		{"gitlab-pat", "glpat-" + strings.Repeat("xX0", 7)[:20]},
		{"shopify-access-token", "shpat_" + strings.Repeat("a", 32)},
	} {
		gitleaksRow(t, c.id)
		if got := Scan(c.s); len(got) != 0 {
			t.Errorf("Scan(%q) = %v; a placeholder is skipped", c.s, got)
		}
		if got := ScanPayload([]byte(`{"command":"echo ` + c.s + `"}`)); len(got) != 0 {
			t.Errorf("ScanPayload(%q) = %v; a placeholder is skipped", c.s, got)
		}
	}
}

// One character off a placeholder is not a placeholder: the test is fixed, not a score.
func TestNearPlaceholdersAreRefused(t *testing.T) {
	for _, c := range []struct{ id, s string }{
		{"gitlab-pat", "glpat-" + strings.Repeat("x", 19) + "y"},
		{"stripe-access-token", "sk_live_" + strings.Repeat("0", 12) + "1" + strings.Repeat("0", 11)},
		{"npm-access-token", "npm_" + strings.Repeat("X", 35) + "a"},
		{"shopify-access-token", "shpat_" + strings.Repeat("a", 31) + "b"},
	} {
		if got := Scan(c.s); !includes(got, c.id) {
			t.Errorf("Scan(%q) = %v; want %q", c.s, got, c.id)
		}
	}
}

// A skipped match does not hide a real one after it.
func TestAPlaceholderDoesNotHideARealSecret(t *testing.T) {
	s := "glpat-" + strings.Repeat("x", 20) + " glpat-" + "Zq3vR8mWk2Lp9TnB4cYd"
	if got := Scan(s); !includes(got, "gitlab-pat") {
		t.Errorf("Scan(%q) = %v; want gitlab-pat", s, got)
	}
}

// noPlaceholderShape names the Placeholder rows whose secret holds a fixed literal besides the
// vendor literal, so no match of them has a placeholder body and none is ever skipped: the
// row refuses on its literal, where gitleaks would drop a low-entropy match.
var noPlaceholderShape = map[string]string{
	"aws-amazon-bedrock-api-key-short-lived": "the whole secret is a literal",
	"jwt":                                    "the secret opens `ey` and its literal is the second segment's `.ey`",
	"sentry-org-token":                       "a region literal sits inside the body",
}

// Every row with a Placeholder, on a sample drawn from its own regex: the sample with its body
// rewritten to a placeholder is skipped, and the same placeholder with one character changed
// is refused.
func TestEveryPlaceholderRowSkipsOnlyPlaceholders(t *testing.T) {
	var n int
	for _, id := range gitleaksRuleIDs {
		row := gitleaksRow(t, id)
		if row.Placeholder == nil {
			continue
		}
		n++
		t.Run(id, func(t *testing.T) {
			r := rand.New(rand.NewSource(seedOf(id) + 2))
			for try := 0; try < 100; try++ {
				s := mustSample(t, row, r)
				body := bodyPositions(row, s)
				if body == nil {
					continue // the literal drew a non-ASCII fold; redraw
				}
				ph, ok := placeholderOf(row, s, body)
				if !ok {
					continue
				}
				if why, ok := noPlaceholderShape[id]; ok {
					t.Fatalf("%q is a placeholder of a row listed in noPlaceholderShape (%s) — remove the entry", ph, why)
				}
				if includes(Scan(ph), id) {
					t.Fatalf("placeholder %q (from %q) is refused as %q", ph, s, id)
				}
				near, ok := nearOf(row, ph, body)
				if !ok {
					continue
				}
				if !includes(Scan(near), id) {
					t.Fatalf("near-placeholder %q is not refused as %q", near, id)
				}
				return
			}
			if _, ok := noPlaceholderShape[id]; !ok {
				t.Fatal("100 draws gave no sample whose body could be rewritten to a placeholder the regex still matches")
			}
		})
	}
	if n == 0 {
		t.Fatal("no generated row has a Placeholder")
	}
}

// bodyPositions returns the byte offsets of the letters and digits of s's secret outside its
// vendor literal, or nil when the literal is not found in ASCII.
func bodyPositions(row Pattern, s string) []int {
	m := row.Re.FindStringSubmatchIndex(s)
	g := row.Placeholder.Group
	if m == nil || m[2*g] < 0 {
		return nil
	}
	lo, hi := m[2*g], m[2*g+1]
	secret := strings.ToLower(s[lo:hi])
	if !row.Placeholder.Fold {
		secret = s[lo:hi]
	}
	at, n := -1, 0
	for _, l := range row.Placeholder.Literals {
		if i := strings.Index(secret, l); i >= 0 && len(l) > n {
			at, n = i, len(l)
		}
	}
	if at < 0 {
		return nil
	}
	var out []int
	for i := lo; i < hi; i++ {
		if i >= lo+at && i < lo+at+n {
			continue
		}
		if c := s[i]; 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
			out = append(out, i)
		}
	}
	return out
}

// stillMatches reports whether the row's regex matches s with its first secret at the same span
// as in orig.
func stillMatches(row Pattern, orig, s string) bool {
	a, b := row.Re.FindStringSubmatchIndex(orig), row.Re.FindStringSubmatchIndex(s)
	g := row.Placeholder.Group
	return b != nil && a[2*g] == b[2*g] && a[2*g+1] == b[2*g+1]
}

// placeholderOf rewrites each body character to the first of x, X and 0 the regex still
// matches with, or else every body character to the body's first character.
func placeholderOf(row Pattern, s string, body []int) (string, bool) {
	if len(body) == 0 {
		return "", false
	}
	b := []byte(s)
	ok := true
	for _, i := range body {
		set := false
		for _, c := range []byte{'x', 'X', '0'} {
			was := b[i]
			b[i] = c
			if stillMatches(row, s, string(b)) {
				set = true
				break
			}
			b[i] = was
		}
		ok = ok && set
	}
	if ok {
		return string(b), true
	}
	b = []byte(s)
	for _, i := range body {
		b[i] = s[body[0]]
	}
	if out := string(b); stillMatches(row, s, out) {
		return out, true
	}
	return "", false
}

// nearOf changes one body character of a placeholder to one outside {x, X, 0} and different
// from the rest, keeping the regex's match.
func nearOf(row Pattern, ph string, body []int) (string, bool) {
	i := body[len(body)/2]
	for _, c := range []byte("ya1Bq7Zk") {
		if c == ph[body[0]] {
			continue
		}
		b := []byte(ph)
		b[i] = c
		if out := string(b); stillMatches(row, ph, out) {
			return out, true
		}
	}
	return "", false
}
