package secrets

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"reflect"
	"regexp/syntax"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/zricethezav/gitleaks/v8/detect"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/secrets/gitleakstoml"
)

func all(*RuleSpec) bool { return true }

// fullDetector is the gate's detector holding every rule in the table.
func fullDetector(t testing.TB) *detect.Detector {
	t.Helper()
	d, err := detector(all)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// reference scans raw as ScanPayload does, with a detector built the way the gate built one
// before the table: gitleaks.toml translated by gitleaks, every rule compiled.
func reference(t testing.TB, raw []byte) []string {
	t.Helper()
	cfg, err := gitleakstoml.Config()
	if err != nil {
		t.Fatal(err)
	}
	ids, err := scanPayload(raw, func([]string) (*detect.Detector, error) {
		return configure(detect.NewDetector(cfg)), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

// The table is gitleaks.toml as gitleaks translates it. A stale table — gitleaks.toml edited,
// gitleaks upgraded, a rule dropped or altered by hand — fails here, naming the rules that differ.
func TestTheTableIsGitleaksTomlAsGitleaksTranslatesIt(t *testing.T) {
	cfg, err := gitleakstoml.Config()
	if err != nil {
		t.Fatal(err)
	}
	diffSpecs(t, SpecOf(cfg), table, "gitleaks.toml translated", "table.go")
}

// The configuration built from the table is the one the table was written from.
func TestTheTableBuildsTheConfigurationItHolds(t *testing.T) {
	cfg, err := table.Config(all)
	if err != nil {
		t.Fatal(err)
	}
	diffSpecs(t, SpecOf(cfg), table, "built from the table", "table.go")
}

func diffSpecs(t *testing.T, got, want Spec, gotName, wantName string) {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}
	g, w := map[string]RuleSpec{}, map[string]RuleSpec{}
	for _, r := range got.Rules {
		g[r.ID] = r
	}
	for _, r := range want.Rules {
		w[r.ID] = r
	}
	var differ []string
	for id, r := range g {
		if wr, ok := w[id]; !ok || !reflect.DeepEqual(r, wr) {
			differ = append(differ, id)
		}
	}
	for id := range w {
		if _, ok := g[id]; !ok {
			differ = append(differ, id)
		}
	}
	sort.Strings(differ)
	gr, wr := got, want
	gr.Rules, wr.Rules = nil, nil
	t.Fatalf("%s differs from %s (run `go run ./devcmd/secretsgen`): rules %v; configuration besides its rules equal: %t",
		gotName, wantName, differ, reflect.DeepEqual(gr, wr))
}

// Every rule with a content pattern refuses a sample of it through the gate, and the gate reports
// exactly what the detector holding every rule reports. Each sample is drawn from the rule's own
// pattern until the reference detector refuses it under that rule, so the rule's keywords,
// entropy threshold and allowlists all take part; the sample is then scanned as drawn and base64-
// encoded, where its keyword is only found on a decoding pass.
func TestEveryRuleFiresOnItsSampleThroughTheGate(t *testing.T) {
	var unsampled []string
	pathRules := 0
	for i := range table.Rules {
		r := &table.Rules[i]
		if r.Path != "" {
			// A rule with a path pattern holds only in a file whose path matches, and a payload
			// has no path.
			pathRules++
			continue
		}
		if r.Regex == "" {
			t.Errorf("rule %s has neither a content nor a path pattern", r.ID)
			continue
		}
		sample, ok := draw(t, r)
		if !ok {
			unsampled = append(unsampled, r.ID)
			continue
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(sample))
		for i, text := range []string{sample, "echo " + encoded + " | base64 -d"} {
			raw, err := json.Marshal(map[string]string{"command": text})
			if err != nil {
				t.Fatal(err)
			}
			want := reference(t, raw)
			got, err := ScanPayload(raw)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("rule %s: the gate reports %v where every rule compiled reports %v", r.ID, got, want)
			}
			if i == 0 && !slices.Contains(got, r.ID) {
				t.Errorf("rule %s: its sample is not refused under it (%v)", r.ID, got)
			}
		}
	}
	if len(unsampled) > 0 {
		t.Errorf("no sample drawn that the reference refuses, for %d rules: %v", len(unsampled), unsampled)
	}
	t.Logf("%d rules fired on their samples; %d rules hold only in a file whose path matches, and a payload has none", len(table.Rules)-pathRules-len(unsampled), pathRules)
}

// draw returns a string the reference detector refuses under r, drawn from r's pattern.
func draw(t *testing.T, r *RuleSpec) (string, bool) {
	t.Helper()
	re, err := syntax.Parse(r.Regex, syntax.Perl)
	if err != nil {
		t.Fatalf("rule %s: %v", r.ID, err)
	}
	rng := rand.New(rand.NewSource(int64(len(r.ID)) * 7919))
	for try := 0; try < 400; try++ {
		var b strings.Builder
		if len(r.Keywords) > 0 && try%2 == 1 {
			// A rule whose keyword is not in its pattern looks for it nearby.
			b.WriteString(r.Keywords[rng.Intn(len(r.Keywords))] + " ")
		}
		gen(&b, re, rng)
		s := "\n" + b.String() + "\n"
		raw, err := json.Marshal(map[string]string{"command": s})
		if err != nil {
			t.Fatal(err)
		}
		if slices.Contains(reference(t, raw), r.ID) {
			return s, true
		}
	}
	return "", false
}

// gen writes a random string matching re: the fewest repetitions a pattern allows, plus now and
// then one more, and printable ASCII from each class where the class has any.
func gen(b *strings.Builder, re *syntax.Regexp, rng *rand.Rand) {
	switch re.Op {
	case syntax.OpLiteral:
		for _, r := range re.Rune {
			if re.Flags&syntax.FoldCase != 0 && rng.Intn(2) == 0 {
				r = []rune(strings.ToUpper(string(r)))[0]
			}
			b.WriteRune(r)
		}
	case syntax.OpCharClass:
		b.WriteRune(pick(re.Rune, rng))
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		// An unescaped dot in a pattern is usually a literal one: hooks.slack.com.
		b.WriteByte('.')
	case syntax.OpCapture:
		gen(b, re.Sub[0], rng)
	case syntax.OpConcat:
		for _, s := range re.Sub {
			gen(b, s, rng)
		}
	case syntax.OpAlternate:
		gen(b, re.Sub[rng.Intn(len(re.Sub))], rng)
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		lo, hi := 0, 1
		switch re.Op {
		case syntax.OpPlus:
			lo, hi = 1, 2
		case syntax.OpRepeat:
			lo, hi = re.Min, re.Max
			if hi < 0 {
				hi = lo + 2
			}
		}
		n := lo
		if hi > lo && rng.Intn(4) == 0 {
			n++
		}
		for range n {
			gen(b, re.Sub[0], rng)
		}
	}
	// Empty-width operators (^ $ \b \A \z) and empty matches write nothing.
}

// pick returns a rune from the class's ranges, printable ASCII where the class has any.
func pick(ranges []rune, rng *rand.Rand) rune {
	var printable []rune
	for i := 0; i+1 < len(ranges); i += 2 {
		for r := max(ranges[i], 0x21); r <= min(ranges[i+1], 0x7e); r++ {
			printable = append(printable, r)
		}
	}
	if len(printable) > 0 {
		return printable[rng.Intn(len(printable))]
	}
	r := ranges[0]
	if !utf8.ValidRune(r) {
		r = 'x'
	}
	return r
}
