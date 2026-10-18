package secrets

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"regexp/syntax"
	"strings"
	"testing"
	"unicode"
	"unicode/utf16"
)

// WHY THESE SAMPLES ARE DRAWN FROM THE REGEX.
//
// gitleaks' config carries no example per rule, and a checked-in sample in a real vendor shape
// is what push protection and every scanner downstream flag. So each generated row's samples
// are drawn at test time from its own regex's structure (sampleOf), seeded by the rule id.
// The regex itself is pinned by secretsgen -check; what these tests prove is that the row is
// present and that its prefilter literals never hide a match of it.
//
// The list of rules comes from gitleaksRuleIDs, which secretsgen writes apart from the rows,
// so deleting a row fails here instead of taking its test with it.

const gitleaksSamples = 500

func gitleaksRow(t *testing.T, id string) Pattern {
	t.Helper()
	var rows []Pattern
	for _, p := range Patterns {
		if p.Class == id {
			rows = append(rows, p)
		}
	}
	if len(rows) != 1 {
		t.Fatalf("gitleaks rule %q has %d rows in Patterns, want 1 — regenerate with `go -C scripts run ./secretsgen`", id, len(rows))
	}
	return rows[0]
}

func TestEveryGitleaksRuleHasItsRow(t *testing.T) {
	if len(gitleaksRuleIDs) == 0 {
		t.Fatal("secretsgen selected no gitleaks rules")
	}
	if len(gitleaksPatterns) != len(gitleaksRuleIDs) {
		t.Errorf("gitleaksPatterns has %d rows for %d selected rules", len(gitleaksPatterns), len(gitleaksRuleIDs))
	}
	for _, id := range gitleaksRuleIDs {
		gitleaksRow(t, id)
	}
}

func seedOf(id string) int64 {
	h := fnv.New64a()
	h.Write([]byte(id))
	return int64(h.Sum64())
}

// Every sample of every generated row is caught as that row, bare, in each surrounding
// context in which the row's own regex matches, and once \uXXXX-escaped in a JSON payload.
func TestGitleaksSamplesMatchTheirRow(t *testing.T) {
	for _, id := range gitleaksRuleIDs {
		t.Run(id, func(t *testing.T) {
			row := gitleaksRow(t, id)
			r := rand.New(rand.NewSource(seedOf(id)))
			for i := 0; i < gitleaksSamples; i++ {
				s := mustSample(t, row, r)
				if !includes(Scan(s), id) {
					t.Fatalf("sample %d %q matches the row's regex but Scan does not report %q", i, s, id)
				}
				ctx := contexts[i%len(contexts)](s)
				if row.Re.MatchString(ctx) && !includes(Scan(ctx), id) {
					t.Fatalf("sample %d in context %q matches the row's regex but Scan does not report %q", i, ctx, id)
				}
			}
			s := mustSample(t, row, r)
			if got := ScanPayload([]byte(`{"command":"` + escapeAll(s) + `"}`)); !includes(got, id) {
				t.Errorf("ScanPayload of %q, \\uXXXX-escaped, = %v; want %q", s, got, id)
			}
		})
	}
}

// One row per vendor: a generated row's samples must not also be a hand row's. The hand
// formats are checked the other way round by TestGeneratedKeysMatchTheirRow, which asserts
// exactly one class per sample.
func TestGitleaksRowsDoNotDuplicateAHandRow(t *testing.T) {
	for _, id := range gitleaksRuleIDs {
		row := gitleaksRow(t, id)
		r := rand.New(rand.NewSource(seedOf(id) + 1))
		for i := 0; i < gitleaksSamples; i++ {
			s := mustSample(t, row, r)
			for _, h := range handPatterns {
				if h.Re.MatchString(s) {
					t.Fatalf("%s sample %q is also hand row %q — keep one row per vendor (secretsgen handCovered)", id, s, h.Class)
				}
			}
		}
	}
}

// The case fold the prefilter uses must cover every rune a case-insensitive regex folds onto
// an ASCII letter, or a (?i) row's literal would be missed in text spelled with that rune.
func TestFoldTextCoversEveryFoldOntoASCII(t *testing.T) {
	for r := rune(0); r <= unicode.MaxRune; r++ {
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if f >= 'a' && f <= 'z' {
				if got := foldText(string(r)); got != string(f) {
					t.Errorf("foldText(%U) = %q; a (?i) regex folds it onto %q", r, got, f)
				}
				break
			}
		}
	}
}

func includes(classes []string, id string) bool {
	for _, c := range classes {
		if c == id {
			return true
		}
	}
	return false
}

// escapeAll writes every character of s as a JSON \u escape.
func escapeAll(s string) string {
	var b strings.Builder
	for _, r := range s {
		for _, u := range utf16.Encode([]rune{r}) {
			fmt.Fprintf(&b, `\u%04x`, u)
		}
	}
	out := b.String()
	var back struct{ Command string }
	if err := json.Unmarshal([]byte(`{"command":"`+out+`"}`), &back); err != nil || back.Command != s {
		panic(fmt.Sprintf("escapeAll(%q) does not decode back: %v", s, err))
	}
	return out
}

// mustSample draws a string the row's regex matches, retrying a draw that lands on an
// assertion its neighbours cannot satisfy.
func mustSample(t *testing.T, row Pattern, r *rand.Rand) string {
	t.Helper()
	re, err := syntax.Parse(row.Re.String(), syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}
	for try := 0; try < 100; try++ {
		var b strings.Builder
		sampleOf(re, r, &b)
		if s := b.String(); row.Re.MatchString(s) {
			return s
		}
	}
	t.Fatalf("%s: 100 draws from its regex matched none — sampleOf cannot reach this shape", row.Class)
	return ""
}

// sampleOf writes one random string in the language of re: each class draws from its whole
// range, each case-insensitive literal draws each letter from its whole fold orbit (so the
// folded prefilter is exercised with U+017F and U+212A as well as upper case), and each
// repeat draws a count from its minimum to a few past it.
func sampleOf(re *syntax.Regexp, r *rand.Rand, b *strings.Builder) {
	switch re.Op {
	case syntax.OpLiteral:
		for _, c := range re.Rune {
			if re.Flags&syntax.FoldCase != 0 {
				orbit := []rune{c}
				for f := unicode.SimpleFold(c); f != c; f = unicode.SimpleFold(f) {
					orbit = append(orbit, f)
				}
				c = orbit[r.Intn(len(orbit))]
			}
			b.WriteRune(c)
		}
	case syntax.OpCharClass:
		var total int
		for i := 0; i+1 < len(re.Rune); i += 2 {
			total += int(re.Rune[i+1]-re.Rune[i]) + 1
		}
		n := r.Intn(total)
		for i := 0; i+1 < len(re.Rune); i += 2 {
			w := int(re.Rune[i+1]-re.Rune[i]) + 1
			if n < w {
				b.WriteRune(re.Rune[i] + rune(n))
				return
			}
			n -= w
		}
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		b.WriteByte('x')
	case syntax.OpCapture:
		sampleOf(re.Sub[0], r, b)
	case syntax.OpConcat:
		for _, s := range re.Sub {
			sampleOf(s, r, b)
		}
	case syntax.OpAlternate:
		// An alternative that consumes nothing (`$` closing a terminator group) is taken only
		// when there is no other, so the sample keeps its terminator inside any context.
		var subs []*syntax.Regexp
		for _, s := range re.Sub {
			if !zeroWidth(s) {
				subs = append(subs, s)
			}
		}
		if len(subs) == 0 {
			subs = re.Sub
		}
		sampleOf(subs[r.Intn(len(subs))], r, b)
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		lo, hi := re.Min, re.Max
		switch re.Op {
		case syntax.OpStar:
			lo, hi = 0, -1
		case syntax.OpPlus:
			lo, hi = 1, -1
		case syntax.OpQuest:
			lo, hi = 0, 1
		}
		if hi < 0 || hi > lo+3 {
			hi = lo + 3
		}
		for n := lo + r.Intn(hi-lo+1); n > 0; n-- {
			sampleOf(re.Sub[0], r, b)
		}
	}
}

func zeroWidth(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText,
		syntax.OpEndText, syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return true
	}
	return false
}
