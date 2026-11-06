package secrets

import (
	"slices"
	"sort"
	"strings"

	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect/codec"
	"github.com/zricethezav/gitleaks/v8/regexp"
)

// Spec is a translated gitleaks configuration with every pattern held as its source text: what
// gitleaks' Translate builds from gitleaks.toml and gitleaks' default configuration, before a
// regular expression is compiled or an allowlist's stopword trie is built. table.go holds the
// gate's Spec, written by `go run ./devcmd/secretsgen` from gitleaks.toml, so no call parses
// TOML; the test that reads gitleaks.toml through gitleaks fails a stale table.
type Spec struct {
	Title        string
	Description  string
	Extend       config.Extend
	Path         string
	MinVersion   string
	Keywords     []string // every rule's keywords, lower-cased and sorted: the detector's prefilter
	OrderedRules []string
	Allowlists   []AllowSpec // the global allowlists
	Rules        []RuleSpec  // sorted by ID
}

// RuleSpec is a config.Rule with its patterns as source text.
type RuleSpec struct {
	ID          string
	Description string
	Regex       string // "" for a rule with no content pattern
	Path        string // "" for a rule with no path pattern
	SecretGroup int
	Entropy     float64
	Keywords    []string
	Tags        []string
	Allowlists  []AllowSpec
	Required    []RequiredSpec
	SkipReport  bool
}

// RequiredSpec is a config.Required.
type RequiredSpec struct {
	RuleID        string
	WithinLines   *int
	WithinColumns *int
}

// AllowSpec is a config.Allowlist with its patterns as source text. Commits and StopWords are
// sorted: gitleaks keeps them in map order.
type AllowSpec struct {
	Description    string
	MatchCondition config.AllowlistMatchCondition
	Commits        []string
	Paths          []string
	RegexTarget    string
	Regexes        []string
	StopWords      []string
}

// SpecOf is cfg with its patterns as source text.
func SpecOf(cfg config.Config) Spec {
	s := Spec{
		Title:        cfg.Title,
		Description:  cfg.Description,
		Extend:       cfg.Extend,
		Path:         cfg.Path,
		MinVersion:   cfg.MinVersion,
		OrderedRules: slices.Clone(cfg.OrderedRules),
		Allowlists:   allowSpecs(cfg.Allowlists),
	}
	for k := range cfg.Keywords {
		s.Keywords = append(s.Keywords, k)
	}
	sort.Strings(s.Keywords)
	for _, r := range cfg.Rules {
		rs := RuleSpec{
			ID:          r.RuleID,
			Description: r.Description,
			Regex:       source(r.Regex),
			Path:        source(r.Path),
			SecretGroup: r.SecretGroup,
			Entropy:     r.Entropy,
			Keywords:    slices.Clone(r.Keywords),
			Tags:        slices.Clone(r.Tags),
			Allowlists:  allowSpecs(r.Allowlists),
			SkipReport:  r.SkipReport,
		}
		for _, q := range r.RequiredRules {
			rs.Required = append(rs.Required, RequiredSpec{RuleID: q.RuleID, WithinLines: q.WithinLines, WithinColumns: q.WithinColumns})
		}
		s.Rules = append(s.Rules, rs)
	}
	sort.Slice(s.Rules, func(i, j int) bool { return s.Rules[i].ID < s.Rules[j].ID })
	return s
}

func allowSpecs(as []*config.Allowlist) []AllowSpec {
	var out []AllowSpec
	for _, a := range as {
		s := AllowSpec{
			Description:    a.Description,
			MatchCondition: a.MatchCondition,
			Commits:        sorted(a.Commits),
			RegexTarget:    a.RegexTarget,
			StopWords:      sorted(a.StopWords),
		}
		for _, p := range a.Paths {
			s.Paths = append(s.Paths, p.String())
		}
		for _, p := range a.Regexes {
			s.Regexes = append(s.Regexes, p.String())
		}
		out = append(out, s)
	}
	return out
}

func sorted(xs []string) []string {
	if xs == nil {
		return nil
	}
	out := slices.Clone(xs)
	sort.Strings(out)
	return out
}

func source(re *regexp.Regexp) string {
	if re == nil {
		return ""
	}
	return re.String()
}

// Config builds the gitleaks configuration s describes, holding only the rules keep selects and
// the rules they require. Each held rule's patterns are compiled and its allowlists validated,
// exactly as Translate does; a rule left out costs nothing.
//
// THE PREFILTER KEYWORDS ARE THE HELD RULES', and every rule's when keep holds them all. The
// detector builds an Aho-Corasick trie over them for each call, and reads its matches only to ask
// whether a held rule's keyword occurs, so a keyword no held rule names selects nothing; building
// the trie over every rule's was most of a small payload's scan.
func (s Spec) Config(keep func(*RuleSpec) bool) (config.Config, error) {
	cfg := config.Config{
		Title:        s.Title,
		Description:  s.Description,
		Extend:       s.Extend,
		Path:         s.Path,
		MinVersion:   s.MinVersion,
		Rules:        map[string]config.Rule{},
		Keywords:     map[string]struct{}{},
		OrderedRules: slices.Clone(s.OrderedRules),
	}
	var err error
	if cfg.Allowlists, err = allowlists(s.Allowlists); err != nil {
		return config.Config{}, err
	}
	byID := make(map[string]*RuleSpec, len(s.Rules))
	for i := range s.Rules {
		byID[s.Rules[i].ID] = &s.Rules[i]
	}
	var add func(*RuleSpec) error
	add = func(r *RuleSpec) error {
		if _, ok := cfg.Rules[r.ID]; ok {
			return nil
		}
		cr := config.Rule{
			RuleID:      r.ID,
			Description: r.Description,
			SecretGroup: r.SecretGroup,
			Entropy:     r.Entropy,
			Keywords:    slices.Clone(r.Keywords),
			Tags:        slices.Clone(r.Tags),
			SkipReport:  r.SkipReport,
		}
		if r.Regex != "" {
			cr.Regex = regexp.MustCompile(r.Regex)
		}
		if r.Path != "" {
			cr.Path = regexp.MustCompile(r.Path)
		}
		var err error
		if cr.Allowlists, err = allowlists(r.Allowlists); err != nil {
			return err
		}
		cfg.Rules[r.ID] = cr // held before its required rules, so a cycle ends
		for _, q := range r.Required {
			cr.RequiredRules = append(cr.RequiredRules, &config.Required{RuleID: q.RuleID, WithinLines: q.WithinLines, WithinColumns: q.WithinColumns})
			if err := add(byID[q.RuleID]); err != nil {
				return err
			}
		}
		cfg.Rules[r.ID] = cr
		return nil
	}
	for i := range s.Rules {
		if keep(&s.Rules[i]) {
			if err := add(&s.Rules[i]); err != nil {
				return config.Config{}, err
			}
		}
	}
	for _, r := range cfg.Rules {
		for _, k := range r.Keywords {
			cfg.Keywords[strings.ToLower(k)] = struct{}{}
		}
	}
	return cfg, nil
}

func allowlists(specs []AllowSpec) ([]*config.Allowlist, error) {
	var out []*config.Allowlist
	for _, s := range specs {
		a := &config.Allowlist{
			Description:    s.Description,
			MatchCondition: s.MatchCondition,
			Commits:        slices.Clone(s.Commits),
			RegexTarget:    s.RegexTarget,
			StopWords:      slices.Clone(s.StopWords),
		}
		for _, p := range s.Paths {
			a.Paths = append(a.Paths, regexp.MustCompile(p))
		}
		for _, p := range s.Regexes {
			a.Regexes = append(a.Regexes, regexp.MustCompile(p))
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// prefiltered returns, for texts, the keep function that holds every rule the detector could run
// on them: a rule with no keywords, and a rule with a keyword in a text the detector scans. That
// is the detector's own loop — each text lower-cased, then each decoding pass up to
// maxDecodeDepth, with gitleaks' own decoder — and a keyword its prefilter matches is a substring
// of that pass, so a rule left out is a rule the detector skips on every pass. A substring test
// per keyword replaces the prefilter's trie: building the trie costs more than the tests it saves
// on a tool payload, and it may hold a rule the trie would not, which the detector then skips.
func (s Spec) prefiltered(texts ...string) func(*RuleSpec) bool {
	found := map[string]bool{}
	for _, text := range texts {
		raw := text
		segments := []*codec.EncodedSegment{}
		decoder := codec.NewDecoder()
		for depth := 0; ; {
			lower := strings.ToLower(raw)
			for _, k := range s.Keywords {
				if !found[k] && strings.Contains(lower, k) {
					found[k] = true
				}
			}
			depth++
			if depth > maxDecodeDepth {
				break
			}
			if raw, segments = decoder.Decode(raw, segments); len(segments) == 0 {
				break
			}
		}
	}
	return func(r *RuleSpec) bool {
		if len(r.Keywords) == 0 {
			return true
		}
		for _, k := range r.Keywords {
			if found[strings.ToLower(k)] {
				return true
			}
		}
		return false
	}
}
