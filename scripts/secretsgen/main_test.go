package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// Each clause of the criterion, on a rule written to fail exactly that clause.
func TestChooseAppliesEachClause(t *testing.T) {
	prefixed := `\b(acme_[a-z0-9]{32})(?:[\x60'"\s;]|\\[nr]|$)`
	for _, c := range []struct {
		name   string
		r      rule
		reason string // the reason's head; "" means selected
		lits   []string
		fold   bool
	}{
		{"vendor prefix", rule{ID: "a", Regex: prefixed, Keywords: []string{"acme_"}}, "", []string{"acme_"}, false},
		{"case-insensitive prefix", rule{ID: "a", Regex: `(?i)\b(ACME_[a-z0-9]{32})\b`, Keywords: []string{"acme_"}}, "", []string{"acme_"}, true},
		{"alternative prefixes", rule{ID: "a", Regex: `\b((?:acme|acmo)_x[a-z]{9})\b`, Keywords: []string{"acm"}}, "", []string{"acme_x", "acmo_x"}, false},
		{"path only", rule{ID: "a", Path: `\.pem$`}, rNoRegex, nil, false},
		{"entropy with a vendor literal", rule{ID: "a", Regex: prefixed, Entropy: 3, Keywords: []string{"acme_"}}, "", []string{"acme_"}, false},
		{"entropy, generic token", rule{ID: "a", Regex: `(?i)(?:acme)(?:[ \t\w.-]{0,20})[=:]\s*([a-f0-9]{32})\b`, Entropy: 3, Keywords: []string{"acme"}}, rEntropy, nil, false},
		{"entropy, literal not a keyword", rule{ID: "a", Regex: `\b(key-[a-z]{32})\b`, Entropy: 3, Keywords: []string{"acme"}}, rEntropy, nil, false},
		{"allowlist", rule{ID: "a", Regex: prefixed, Keywords: []string{"acme_"}, Allowlists: []map[string]any{{"regexes": []string{"x"}}}}, rAllowlist, nil, false},
		{"context keyword, generic token", rule{ID: "a", Regex: `(?i)(?:acme)(?:[ \t\w.-]{0,20})[=:]\s*([a-f0-9]{32})\b`, Keywords: []string{"acme"}}, rNoLiteral, nil, false},
		{"literal too short", rule{ID: "a", Regex: `\b(a_[a-z]{32})\b`, Keywords: []string{"a_"}}, rNoLiteral, nil, false},
		{"literal not a keyword", rule{ID: "a", Regex: `\b(key-[a-z]{32})\b`, Keywords: []string{"acme"}}, rNoLiteral, nil, false},
		{"covered by a hand row", rule{ID: "anthropic-api-key", Regex: `\b(sk-ant-api03-[a-zA-Z0-9_\-]{93}AA)\b`, Keywords: []string{"sk-ant-api03"}}, rHand, nil, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, reason := choose(c.r)
			if c.reason == "" {
				if reason != "" {
					t.Fatalf("excluded: %s", reason)
				}
				if !reflect.DeepEqual(s.literals, c.lits) || s.fold != c.fold {
					t.Errorf("literals %q fold %t; want %q fold %t", s.literals, s.fold, c.lits, c.fold)
				}
				return
			}
			if !strings.HasPrefix(reason, c.reason) {
				t.Errorf("reason %q; want it to start %q", reason, c.reason)
			}
		})
	}
}

// The context word is the stronger prefilter when the token's own prefix is common.
func TestChoosePrefersTheRarerRequiredLiteral(t *testing.T) {
	r := rule{ID: "a", Regex: `(?i)(?:gocardless)(?:[ \t\w.-]{0,20})[=:]\s*(live_[a-z0-9]{40})\b`, Keywords: []string{"live_"}}
	s, reason := choose(r)
	if reason != "" || !reflect.DeepEqual(s.literals, []string{"gocardless"}) || !s.fold {
		t.Errorf("got %q fold %t (%s); want the context word, folded", s.literals, s.fold, reason)
	}
}

// The pinned file selects rules, excludes the rest with a reason each, and every handCovered
// entry is live.
func TestClassifyThePinnedRuleset(t *testing.T) {
	root, err := gitx.Root()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, vendored))
	if err != nil {
		t.Fatal(err)
	}
	sel, exc, err := classify(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(sel) == 0 || len(exc) == 0 {
		t.Fatalf("selected %d, excluded %d", len(sel), len(exc))
	}
	for _, e := range exc {
		if e.reason == "" {
			t.Errorf("%s excluded with no reason", e.id)
		}
	}
	for _, s := range sel {
		if len(s.literals) == 0 {
			t.Errorf("%s selected with no prefilter literal", s.rule.ID)
		}
	}
}

// A handCovered entry for a rule the criterion already refuses is an error, not a no-op.
func TestClassifyRefusesADeadHandEntry(t *testing.T) {
	ruleset := func(allowlist string) []byte {
		var b strings.Builder
		for id := range handCovered {
			b.WriteString("[[rules]]\nid = \"" + id + "\"\n" +
				"regex = '''\\b(sk-ant-x-[a-zA-Z0-9_\\-]{93}AA)\\b'''\n" +
				"keywords = [\"sk-ant-x\"]\n" + allowlist)
		}
		return []byte(b.String())
	}
	if _, _, err := classify(ruleset("")); err != nil {
		t.Fatalf("both entries are live, yet classify refused: %v", err)
	}
	if _, _, err := classify(ruleset("[[rules.allowlists]]\nregexes = ['x']\n")); err == nil {
		t.Error("handCovered entries for allowlisted rules were accepted")
	}
}

// A rule gitleaks gates by entropy carries its secret's group and literals for the placeholder
// skip; a rule without a threshold does not skip.
func TestChooseGivesEntropyRulesAPlaceholderSkip(t *testing.T) {
	r := rule{ID: "a", Regex: `(?i)\b(acme_[a-z0-9]{32})(?:[\x60'"\s;]|$)`, Keywords: []string{"acme_"}}
	if s, reason := choose(r); reason != "" || s.placeholder != nil {
		t.Fatalf("no threshold: placeholder %+v (%s); want none", s.placeholder, reason)
	}
	r.Entropy = 3
	s, reason := choose(r)
	if reason != "" || s.placeholder == nil {
		t.Fatalf("threshold: no placeholder (%s)", reason)
	}
	if want := (placeholder{group: 1, literals: []string{"acme_"}, fold: true}); !reflect.DeepEqual(*s.placeholder, want) {
		t.Errorf("placeholder %+v; want %+v", *s.placeholder, want)
	}
}
