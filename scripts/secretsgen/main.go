// Command secretsgen writes the secrets gate's gitleaks rows from a pinned copy of gitleaks'
// default ruleset.
//
// Dev tooling for this repository only.
//
// # Why
//
// gitleaks publishes ~220 vendor rules and keeps them current; a hand table covers what its
// author thought of. Most of those rules are meaningful only behind an entropy threshold or
// an allowlist, which a PreToolUse gate refusing on a bare match does not have, so only the
// rules that are accurate as bare regexes are taken: the selection criterion is select below,
// and every rule it rejects is listed with its reason in the generated file.
//
// # The pin
//
// gitleaks' config/gitleaks.toml is vendored at gitleaks/ with its MIT LICENSE, and pinVersion
// and pinSHA256 name the release and the file's digest. Vendoring rather than requiring the
// Go module keeps the generator and its -check gate offline, adds no dependency graph to this
// module, and puts every rule change of a bump into the PR diff where a reviewer reads it. A
// bump replaces the file from the new release, updates both constants, and regenerates.
//
// -check fails when the generated file is stale or hand-edited, or when the vendored copy no
// longer has the pinned digest.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	pinVersion = "v8.30.1"
	pinSHA256  = "e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf"
	vendored   = "scripts/secretsgen/gitleaks/gitleaks.toml"
	output     = "plugins/prosthetic-conscience/tools/internal/secrets/gitleaks.go"
	manifest   = "plugins/prosthetic-conscience/tools/internal/secrets/gitleaks_rules_test.go"
)

type rule struct {
	ID          string           `toml:"id"`
	Description string           `toml:"description"`
	Regex       string           `toml:"regex"`
	Path        string           `toml:"path"`
	Entropy     float64          `toml:"entropy"`
	Keywords    []string         `toml:"keywords"`
	SecretGroup int              `toml:"secretGroup"`
	Allowlists  []map[string]any `toml:"allowlists"`
}

type config struct {
	Rules []rule `toml:"rules"`
}

// handCovered names the gitleaks rules whose vendor a hand row in secrets.go already covers,
// with the hand row kept and why. One row per vendor: two rows for one format refuse the same
// payload twice and drift apart.
var handCovered = map[string]string{
	"anthropic-api-key":       "hand row `Anthropic api key` kept: it takes every sk-ant- key type (api03, admin01, oat01) with a body of 20 or more; this rule takes api03 with exactly 93+2",
	"anthropic-admin-api-key": "hand row `Anthropic api key` kept: it takes every sk-ant- key type (api03, admin01, oat01) with a body of 20 or more; this rule takes admin01 with exactly 93+2",
}

// selected is a rule that passed select, with the literals every match of it contains.
type selected struct {
	rule     rule
	literals []string
	fold     bool
}

// excluded is a rule select refused, and why.
type excluded struct {
	id, reason string
}

func main() {
	check := flag.Bool("check", false, "verify the generated files instead of writing them")
	flag.Parse()
	if err := run(*check); err != nil {
		fmt.Fprintln(os.Stderr, "secretsgen:", err)
		os.Exit(1)
	}
}

func run(check bool) error {
	root, err := gitx.Root()
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(filepath.Join(root, vendored))
	if err != nil {
		return err
	}
	if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != pinSHA256 {
		return fmt.Errorf("%s does not have the digest pinned for gitleaks %s — it was edited, or replaced without updating pinVersion and pinSHA256", vendored, pinVersion)
	}
	sel, exc, err := classify(raw)
	if err != nil {
		return err
	}
	files := map[string][]byte{}
	if files[output], err = renderRows(sel, exc); err != nil {
		return err
	}
	if files[manifest], err = renderManifest(sel); err != nil {
		return err
	}
	for _, rel := range []string{output, manifest} {
		path := filepath.Join(root, rel)
		if check {
			have, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(have, files[rel]) {
				return fmt.Errorf("%s is stale or hand-edited — run `go -C scripts run ./secretsgen`", rel)
			}
			continue
		}
		if err := os.WriteFile(path, files[rel], 0o644); err != nil {
			return err
		}
	}
	verb := "wrote"
	if check {
		verb = "verified"
	}
	fmt.Printf("secretsgen: %s %s from gitleaks %s — %d rules selected, %d excluded\n", verb, output, pinVersion, len(sel), len(exc))
	byReason := map[string]int{}
	for _, e := range exc {
		byReason[reasonClass(e.reason)]++
	}
	for _, k := range sortedKeys(byReason) {
		fmt.Printf("  excluded, %s: %d\n", k, byReason[k])
	}
	return nil
}

// classify sorts every rule into selected or excluded, in the config's order.
func classify(raw []byte) ([]selected, []excluded, error) {
	var c config
	if _, err := toml.Decode(string(raw), &c); err != nil {
		return nil, nil, err
	}
	if len(c.Rules) == 0 {
		return nil, nil, fmt.Errorf("%s holds no rules — refusing to generate nothing", vendored)
	}
	var sel []selected
	var exc []excluded
	used := map[string]bool{}
	for _, r := range c.Rules {
		s, reason := choose(r)
		if strings.HasPrefix(reason, rHand) {
			used[r.ID] = true
		}
		if reason != "" {
			exc = append(exc, excluded{r.ID, reason})
			continue
		}
		sel = append(sel, s)
	}
	// An entry for a rule the criterion already refuses, or one gitleaks no longer has, says
	// nothing about what ships: it is refused rather than left to read as a decision.
	for id := range handCovered {
		if !used[id] {
			return nil, nil, fmt.Errorf("handCovered names %q, which gitleaks %s does not have or which choose already excludes on another ground", id, pinVersion)
		}
	}
	return sel, exc, nil
}

// Reasons a rule is excluded. Each is the head of the reason string, which the generated file
// lists rule by rule.
const (
	rNoRegex   = "no regex"
	rEntropy   = "entropy threshold"
	rAllowlist = "allowlist"
	rNoLiteral = "no vendor literal in the secret"
	rHand      = "covered by a hand row"
)

// minLiteral is the shortest literal accepted as a vendor prefix. Shorter ones (`-`, `ey`)
// are structure, not a vendor's mark, and a prefilter on them passes nearly every payload.
const minLiteral = 3

// choose applies the selection criterion. A rule is taken only when all hold:
//
//  1. it has a regex (a path-only rule matches file names, which a tool payload has none of);
//  2. it has no entropy threshold — gitleaks reports its match only when the secret is random
//     enough, and a gate refusing on the bare match would refuse what gitleaks lets through;
//  3. it has no allowlist, for the same reason: gitleaks discards some of its matches;
//  4. every match of the SECRET's own span (secretGroup, else the first capture group, else
//     the whole match) contains one of a set of literals, each at least minLiteral long and
//     each containing one of the rule's keywords. That is a vendor prefix inside the token.
//     A rule whose literal is only in the context around the secret (`twitter ... = <hex>`)
//     fails here: the literal names the vendor, the token itself is generic. The literals
//     become the row's prefilter, so the regex runs only on text holding one;
//  5. no hand row covers its vendor (handCovered).
//
// gitleaks' global allowlist is not applied. Its paths name files, which a tool payload has
// none of; its regexes are anchored whole-secret placeholders (`true`, `$VAR`,
// `${{ secrets.X }}`, a run of one letter) that a secret holding a vendor literal cannot
// equal. Its two stopwords drop a secret CONTAINING them: a fixed UUID, and the lowercase
// alphabet in order. A selected-rule token holding either is a hand-typed placeholder, and
// this gate refuses it where gitleaks would not.
func choose(r rule) (selected, string) {
	switch {
	case r.Regex == "":
		return selected{}, rNoRegex + " (path-only rule)"
	case r.Entropy > 0:
		return selected{}, fmt.Sprintf("%s %.1f", rEntropy, r.Entropy)
	case len(r.Allowlists) > 0:
		return selected{}, rAllowlist
	}
	re, err := syntax.Parse(r.Regex, syntax.Perl)
	if err != nil {
		return selected{}, "regex does not parse under RE2: " + err.Error()
	}
	span := secretSpan(re, r.SecretGroup)
	set := required(span.Simplify())
	if set == nil {
		return selected{}, rNoLiteral + " (the literal is context around a generic token)"
	}
	for _, l := range set.strs {
		if len(l) < minLiteral {
			return selected{}, fmt.Sprintf("%s (shortest required literal %q is under %d)", rNoLiteral, l, minLiteral)
		}
		if !containsKeyword(l, r.Keywords) {
			return selected{}, fmt.Sprintf("%s (required literal %q holds none of the keywords %v)", rNoLiteral, l, r.Keywords)
		}
	}
	if !ascii(set.strs) {
		return selected{}, fmt.Sprintf("%s (a required literal in %q is not ASCII)", rNoLiteral, set.strs)
	}
	if why, ok := handCovered[r.ID]; ok {
		return selected{}, rHand + ": " + why
	}
	// The prefilter takes the stronger of the secret's literals and the whole match's: a
	// context rule's vendor name (`gocardless`) is rarer in ordinary text than its token's
	// prefix (`live_`). Both are required of every match, so either keeps the verdict.
	if whole := required(re.Simplify()); whole != nil && minLen(whole) > minLen(set) && ascii(whole.strs) {
		set = whole
	}
	return selected{rule: r, literals: set.strs, fold: set.fold}, ""
}

// ascii reports whether every literal is ASCII. The gate's case fold (secrets.foldText) is
// proven for ASCII letters only, so a non-ASCII literal is not a prefilter it can use.
func ascii(ls []string) bool {
	for _, l := range ls {
		for i := 0; i < len(l); i++ {
			if l[i] >= 0x80 {
				return false
			}
		}
	}
	return true
}

func containsKeyword(l string, kws []string) bool {
	for _, k := range kws {
		if k != "" && strings.Contains(strings.ToLower(l), strings.ToLower(k)) {
			return true
		}
	}
	return false
}

// secretSpan returns the subexpression gitleaks reports as the secret.
func secretSpan(re *syntax.Regexp, group int) *syntax.Regexp {
	if group == 0 {
		group = 1
	}
	if g := findCap(re, group); g != nil {
		return g
	}
	return re
}

func findCap(re *syntax.Regexp, n int) *syntax.Regexp {
	if re.Op == syntax.OpCapture && re.Cap == n {
		return re
	}
	for _, s := range re.Sub {
		if g := findCap(s, n); g != nil {
			return g
		}
	}
	return nil
}

// litSet is a set of strings, one of which every match contains (or, for exact, IS).
// fold marks strings lowercased because the regex matches them case-insensitively.
type litSet struct {
	strs []string
	fold bool
}

const maxSet = 64

// exact returns the finite set of strings re matches, or nil.
func exact(re *syntax.Regexp) *litSet {
	switch re.Op {
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText,
		syntax.OpEndText, syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return &litSet{strs: []string{""}}
	case syntax.OpLiteral:
		s := string(re.Rune)
		if re.Flags&syntax.FoldCase != 0 {
			return &litSet{strs: []string{strings.ToLower(s)}, fold: true}
		}
		return &litSet{strs: []string{s}}
	case syntax.OpCharClass:
		var out []string
		for i := 0; i+1 < len(re.Rune); i += 2 {
			for c := re.Rune[i]; c <= re.Rune[i+1]; c++ {
				if len(out) >= 8 {
					return nil
				}
				out = append(out, string(c))
			}
		}
		return &litSet{strs: out}
	case syntax.OpCapture:
		return exact(re.Sub[0])
	case syntax.OpConcat:
		acc := &litSet{strs: []string{""}}
		for _, s := range re.Sub {
			e := exact(s)
			if e == nil {
				return nil
			}
			if acc = cross(acc, e); acc == nil {
				return nil
			}
		}
		return acc
	case syntax.OpAlternate:
		acc := &litSet{}
		for _, s := range re.Sub {
			e := exact(s)
			if e == nil {
				return nil
			}
			acc = union(acc, e)
		}
		if len(acc.strs) > maxSet {
			return nil
		}
		return acc
	}
	return nil
}

func cross(a, b *litSet) *litSet {
	if len(a.strs)*len(b.strs) > maxSet {
		return nil
	}
	fold := a.fold || b.fold
	out := &litSet{fold: fold}
	for _, x := range a.strs {
		for _, y := range b.strs {
			out.strs = append(out.strs, x+y)
		}
	}
	return normalise(out)
}

func union(a, b *litSet) *litSet {
	return normalise(&litSet{strs: append(append([]string(nil), a.strs...), b.strs...), fold: a.fold || b.fold})
}

// normalise lowercases every string of a folded set (a case-sensitive literal lowercased and
// compared against lowercased text is still found, so the prefilter stays sound) and dedupes.
func normalise(s *litSet) *litSet {
	seen := map[string]bool{}
	var out []string
	for _, x := range s.strs {
		if s.fold {
			x = strings.ToLower(x)
		}
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	sort.Strings(out)
	return &litSet{strs: out, fold: s.fold}
}

// required returns a set of literals one of which every match of re contains, or nil.
func required(re *syntax.Regexp) *litSet {
	if e := exact(re); e != nil && minLen(e) > 0 {
		return e
	}
	switch re.Op {
	case syntax.OpCapture:
		return required(re.Sub[0])
	case syntax.OpPlus:
		return required(re.Sub[0])
	case syntax.OpRepeat:
		if re.Min >= 1 {
			return required(re.Sub[0])
		}
	case syntax.OpAlternate:
		acc := &litSet{}
		for _, s := range re.Sub {
			r := required(s)
			if r == nil {
				return nil
			}
			acc = union(acc, r)
		}
		if len(acc.strs) > maxSet {
			return nil
		}
		return acc
	case syntax.OpConcat:
		// Candidates: each run of children with an exact set, crossed together, and each
		// child's own required set. The best is the one whose shortest literal is longest.
		var best *litSet
		consider := func(c *litSet) {
			if c == nil || minLen(c) == 0 {
				return
			}
			if best == nil || minLen(c) > minLen(best) || (minLen(c) == minLen(best) && len(c.strs) < len(best.strs)) {
				best = c
			}
		}
		var run *litSet
		for _, s := range re.Sub {
			if e := exact(s); e != nil {
				if run == nil {
					run = e
				} else if run = cross(run, e); run == nil {
					run = e
				}
				consider(run)
				continue
			}
			run = nil
			consider(required(s))
		}
		return best
	}
	return nil
}

func minLen(s *litSet) int {
	if len(s.strs) == 0 {
		return 0
	}
	m := len(s.strs[0])
	for _, x := range s.strs {
		if len(x) < m {
			m = len(x)
		}
	}
	return m
}

func reasonClass(reason string) string {
	for _, k := range []string{rNoRegex, rEntropy, rAllowlist, rNoLiteral, rHand} {
		if strings.HasPrefix(reason, k) {
			return k
		}
	}
	return "regex does not parse"
}

func sortedKeys(m map[string]int) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func renderRows(sel []selected, exc []excluded) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by scripts/secretsgen from gitleaks %s config/gitleaks.toml (MIT, vendored at scripts/secretsgen/gitleaks/); DO NOT EDIT.\n\n", pinVersion)
	b.WriteString("package secrets\n\nimport \"regexp\"\n\n")
	fmt.Fprintf(&b, "// gitleaksPatterns are the %d gitleaks %s rules that are accurate as bare regexes; the\n", len(sel), pinVersion)
	b.WriteString("// criterion is choose in scripts/secretsgen. Each Class is the gitleaks rule id.\n//\n")
	fmt.Fprintf(&b, "// Excluded, %d rules — vendors this gate does NOT cover through gitleaks:\n//\n", len(exc))
	for _, e := range exc {
		fmt.Fprintf(&b, "//   - %s: %s\n", e.id, e.reason)
	}
	b.WriteString("var gitleaksPatterns = []Pattern{\n")
	for _, s := range sel {
		fmt.Fprintf(&b, "\t// %s\n", oneLine(s.rule.Description))
		fmt.Fprintf(&b, "\t{Class: %s, Literals: %s, Fold: %t,\n\t\tRe: regexp.MustCompile(%s)},\n",
			strconv.Quote(s.rule.ID), goStrings(s.literals), s.fold, goRaw(s.rule.Regex))
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

// renderManifest writes the selected rule ids into the test package. The tests read this list,
// not gitleaksPatterns, so a row deleted from the generated table fails a test rather than
// taking its own test with it.
func renderManifest(sel []selected) ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by scripts/secretsgen from gitleaks %s config/gitleaks.toml; DO NOT EDIT.\n\n", pinVersion)
	b.WriteString("package secrets\n\n")
	b.WriteString("// gitleaksRuleIDs is every rule secretsgen selected. A row missing from gitleaksPatterns\n// fails TestEveryGitleaksRuleHasItsRow.\n")
	b.WriteString("var gitleaksRuleIDs = []string{\n")
	for _, s := range sel {
		fmt.Fprintf(&b, "\t%s,\n", strconv.Quote(s.rule.ID))
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func goStrings(ss []string) string {
	q := make([]string, len(ss))
	for i, s := range ss {
		q[i] = strconv.Quote(s)
	}
	return "[]string{" + strings.Join(q, ", ") + "}"
}

func goRaw(s string) string {
	if !strings.Contains(s, "`") {
		return "`" + s + "`"
	}
	return strconv.Quote(s)
}
