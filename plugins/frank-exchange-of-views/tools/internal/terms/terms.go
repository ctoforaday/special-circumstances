// Package terms is the protocol's vocabulary: one entry per concept noun, the sentence that
// defines it, the seats that meet it, and the variants a seat-facing surface must not use for it.
//
// WHY A REGISTRY. The tool's names — verbs, flags, views, enum words — are held to the command
// tree by tests. The words for what those things ARE were held by nothing: one document was
// called by fourteen names across the prompts, constitutions and help, and "carried" meant a gap
// closed on one page and a gap kept open on the next. A seat reads all of those surfaces in one
// sitting and cannot tell a second name from a second thing.
//
// ONE SOURCE, THREE READERS. terms.json is embedded here and read by `manual`, which prints each
// seat the definitions its surface uses; by the gate in integration/surface, which fails on a
// GATED variant anywhere a seat reads; and by scripts/vocabdoc, which generates
// docs/vocabulary.md from the same file.
//
// A GATED ban is a phrase-exact pattern that cannot hit a legitimate sense of a word. Every text
// is normalised once before any ban or mask runs (fold): a run of whitespace and hyphens is one
// space, and a typographic apostrophe is an ASCII one, so a pattern spells a joined phrase with a
// space and a possessive with ' and matches every spelling of each. A word with
// a legitimate neighbour sense stays REGISTRY-ONLY: defined and delivered, not gated. RE2 has no
// lookahead, so a legitimate phrase that contains a banned one is a MASK, blanked before
// matching. An ALLOW exempts a path, and says why.
package terms

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"regexp/syntax"
	"strings"
)

//go:embed terms.json
var registryJSON []byte

// Kind says whether a ban is enforced by the gate or only recorded.
type Kind string

const (
	// Gated bans fail the vocabulary gate wherever they match unmasked and unallowed.
	Gated Kind = "GATED"
	// RegistryOnly bans are defined and delivered, and not gated: the word has a legitimate
	// neighbour sense no phrase-exact pattern can tell apart.
	RegistryOnly Kind = "REGISTRY-ONLY"
)

// Registry is the whole vocabulary.
type Registry struct {
	Entries []Entry `json:"entries"`
}

// Entry is one concept noun.
type Entry struct {
	Term       string      `json:"term"`
	Definition string      `json:"definition"`
	Seats      []string    `json:"seats"`
	Bans       []Ban       `json:"bans"`
	Collisions []Collision `json:"collisions"`
}

// Ban is one variant a seat-facing surface must not use for the entry's concept.
type Ban struct {
	Variant string  `json:"variant"`
	Kind    Kind    `json:"kind"`
	Pattern string  `json:"pattern,omitempty"`
	Masks   []Mask  `json:"masks,omitempty"`
	Allow   []Allow `json:"allow,omitempty"`

	re *regexp.Regexp
}

// Mask is a legitimate phrase containing a banned one, blanked before the ban's pattern runs.
type Mask struct {
	Phrase string `json:"phrase"`
	Reason string `json:"reason"`
}

// Allow exempts the files a path glob names, relative to the repository root. `*` matches within
// one path segment and `**` across segments.
type Allow struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`

	re *regexp.Regexp
}

// Collision is a word with a second, legitimate sense, and how the two are kept apart.
type Collision struct {
	Word       string `json:"word"`
	Resolution string `json:"resolution"`
}

// Load parses and validates the embedded registry.
func Load() (*Registry, error) { return Parse(registryJSON) }

// Parse reads a registry and refuses one that would mislead a reader of it. An unknown field is
// refused too: a misspelt `alow` would otherwise parse as no allow at all and fail the gate
// somewhere far from the typo, or — for `pattren` — leave a ban that never runs.
func Parse(b []byte) (*Registry, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var r Registry
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("terms: %w", err)
	}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *Registry) validate() error {
	if len(r.Entries) == 0 {
		return fmt.Errorf("terms: the registry holds no entries — every reader would deliver and gate nothing")
	}
	seen := map[string]bool{}
	for i := range r.Entries {
		e := &r.Entries[i]
		if strings.TrimSpace(e.Term) == "" {
			return fmt.Errorf("terms: entry %d has no term", i)
		}
		if seen[e.Term] {
			return fmt.Errorf("terms: %q is entered twice — one concept, one entry", e.Term)
		}
		seen[e.Term] = true
		if err := checkDefinition(e); err != nil {
			return err
		}
		if len(e.Seats) == 0 {
			return fmt.Errorf("terms: %q names no seat, so no manual delivers it", e.Term)
		}
		for j := range e.Bans {
			if err := e.Bans[j].compile(e.Term); err != nil {
				return err
			}
		}
		for _, c := range e.Collisions {
			if strings.TrimSpace(c.Word) == "" || strings.TrimSpace(c.Resolution) == "" {
				return fmt.Errorf("terms: %q has a collision without a word or a resolution", e.Term)
			}
		}
	}
	return nil
}

// checkDefinition holds the definition to one present-tense sentence: non-empty, one line, one
// terminal period. A second sentence is where a definition turns into an essay nobody re-reads.
func checkDefinition(e *Entry) error {
	d := strings.TrimSpace(e.Definition)
	switch {
	case d == "":
		return fmt.Errorf("terms: %q has no definition — a term a seat is handed without its meaning is a second name, not a word", e.Term)
	case strings.ContainsAny(d, "\n\r"):
		return fmt.Errorf("terms: %q: the definition is one line", e.Term)
	case !strings.HasSuffix(d, "."):
		return fmt.Errorf("terms: %q: the definition is a sentence and ends with a period", e.Term)
	case strings.Contains(strings.TrimSuffix(d, "."), ". "):
		return fmt.Errorf("terms: %q: the definition is ONE sentence", e.Term)
	}
	return nil
}

func (b *Ban) compile(term string) error {
	if strings.TrimSpace(b.Variant) == "" {
		return fmt.Errorf("terms: %q has a ban with no variant", term)
	}
	switch b.Kind {
	case Gated:
		if strings.TrimSpace(b.Pattern) == "" {
			return fmt.Errorf("terms: %q bans %q as GATED with no pattern — a gate with nothing to match passes everything", term, b.Variant)
		}
		re, err := regexp.Compile("(?i)" + b.Pattern)
		if err != nil {
			return fmt.Errorf("terms: %q bans %q: the pattern is not RE2: %w", term, b.Variant, err)
		}
		// A HYPHEN IN A PATTERN NEVER MATCHES: fold turns every hyphen in the text into a space
		// before the pattern runs, so a literal one would make the ban pass everything it names.
		if literalHyphen(b.Pattern) {
			return fmt.Errorf("terms: %q bans %q: the pattern %q has a literal hyphen, which never matches — the scan folds every hyphen in the text to a space, so spell the join as a space", term, b.Variant, b.Pattern)
		}
		b.re = re
	case RegistryOnly:
		if b.Pattern != "" || len(b.Masks) > 0 || len(b.Allow) > 0 {
			return fmt.Errorf("terms: %q bans %q as REGISTRY-ONLY and gives it a pattern, a mask or an allow — nothing runs those, so they would read as enforcement that does not happen", term, b.Variant)
		}
	default:
		return fmt.Errorf("terms: %q bans %q with kind %q; the kinds are %s and %s", term, b.Variant, b.Kind, Gated, RegistryOnly)
	}
	for _, m := range b.Masks {
		if strings.TrimSpace(m.Phrase) == "" || strings.TrimSpace(m.Reason) == "" {
			return fmt.Errorf("terms: %q bans %q with a mask missing its phrase or its reason", term, b.Variant)
		}
		// A MASK THE PATTERN CANNOT MATCH BLANKS NOTHING. It would sit in the registry reading as
		// a legitimate neighbour the gate knows about while protecting nothing.
		if p, _, _ := fold(m.Phrase); !b.re.MatchString(p) {
			return fmt.Errorf("terms: %q bans %q: the mask %q does not contain a match for the pattern, so it blanks nothing", term, b.Variant, m.Phrase)
		}
	}
	for i := range b.Allow {
		a := &b.Allow[i]
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf("terms: %q bans %q with an allow that names no path", term, b.Variant)
		}
		if strings.TrimSpace(a.Reason) == "" {
			return fmt.Errorf("terms: %q bans %q and allows %s without a reason — an allow nobody can argue with is how a ban erodes", term, b.Variant, a.Path)
		}
		a.re = globRe(a.Path)
	}
	return nil
}

// literalHyphen reports whether a pattern can only match a hyphen it names outright: a literal
// '-' anywhere in its parsed tree. A hyphen inside a character class is a range or a member, and
// a class that also matches a space still matches the folded text, so classes are not refused.
func literalHyphen(pattern string) bool {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return false
	}
	var walk func(*syntax.Regexp) bool
	walk = func(r *syntax.Regexp) bool {
		if r.Op == syntax.OpLiteral && strings.ContainsRune(string(r.Rune), '-') {
			return true
		}
		for _, sub := range r.Sub {
			if walk(sub) {
				return true
			}
		}
		return false
	}
	return walk(re)
}

// globRe turns a path glob into an anchored regexp: `**` crosses segments, `*` and `?` do not.
func globRe(glob string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; {
		case c == '*' && i+1 < len(glob) && glob[i+1] == '*':
			b.WriteString(".*")
			i++
		case c == '*':
			b.WriteString("[^/]*")
		case c == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// Covers reports whether the allow exempts a repository-relative, slash-separated path.
func (a Allow) Covers(path string) bool { return a.re != nil && a.re.MatchString(path) }

// ForSeat returns the entries a seat's surface delivers, in registry order.
func (r *Registry) ForSeat(role string) []Entry {
	var out []Entry
	for _, e := range r.Entries {
		for _, s := range e.Seats {
			if s == role {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

// Seats is every seat value the registry names, in first-appearance order.
func (r *Registry) Seats() []string {
	var out []string
	seen := map[string]bool{}
	for _, e := range r.Entries {
		for _, s := range e.Seats {
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	return out
}

// Hit is one GATED variant found in a text.
type Hit struct {
	Term    string // the canonical term to use instead
	Variant string
	Match   string // the text that matched, as written
	Line    int    // 1-based, within the scanned text
	// AllowedBy is the allow glob that exempts this path, or "" when nothing does.
	AllowedBy string
}

// Usage counts which masks and allows did any work across a scan, so a stale one can be named.
type Usage struct {
	masks  map[string]int
	allows map[string]int
}

// NewUsage starts an empty count.
func NewUsage() *Usage { return &Usage{masks: map[string]int{}, allows: map[string]int{}} }

func maskKey(term, variant, phrase string) string { return term + "\x00" + variant + "\x00" + phrase }

// Scan runs every GATED ban over one text from one repository-relative path.
//
// The text is normalised once (fold) and every ban and mask runs over the same normalised bytes,
// so a phrase hard-wrapped across two lines, joined with a hyphen, or written with a typographic
// apostrophe is still the phrase. Line numbers and the reported match are the original's.
func (r *Registry) Scan(path, text string, u *Usage) []Hit {
	norm, starts, ends := fold(text)
	var hits []Hit
	for _, e := range r.Entries {
		for _, b := range e.Bans {
			if b.Kind != Gated {
				continue
			}
			masked := norm
			for _, m := range b.Masks {
				var n int
				masked, n = blank(masked, m.Phrase)
				if n > 0 && u != nil {
					u.masks[maskKey(e.Term, b.Variant, m.Phrase)] += n
				}
			}
			for _, loc := range b.re.FindAllStringIndex(masked, -1) {
				end := starts[loc[0]]
				if loc[1] > loc[0] {
					end = ends[loc[1]-1]
				}
				h := Hit{Term: e.Term, Variant: b.Variant, Match: text[starts[loc[0]]:end], Line: lineAt(text, starts[loc[0]])}
				for _, a := range b.Allow {
					if a.Covers(path) {
						h.AllowedBy = a.Path
						if u != nil {
							u.allows[maskKey(e.Term, b.Variant, a.Path)]++
						}
						break
					}
				}
				hits = append(hits, h)
			}
		}
	}
	return hits
}

// Stale names every mask that blanked nothing and every allow that exempted nothing across the
// scans u counted. A mask or allow that does no work is a hole waiting for a real hit to fall into
// it, and "delete the row and see if anything fails" is the question it has to answer.
func (r *Registry) Stale(u *Usage) []string {
	var out []string
	for _, e := range r.Entries {
		for _, b := range e.Bans {
			for _, m := range b.Masks {
				if u.masks[maskKey(e.Term, b.Variant, m.Phrase)] == 0 {
					out = append(out, fmt.Sprintf("%q / %q: the mask %q blanked nothing", e.Term, b.Variant, m.Phrase))
				}
			}
			for _, a := range b.Allow {
				if u.allows[maskKey(e.Term, b.Variant, a.Path)] == 0 {
					out = append(out, fmt.Sprintf("%q / %q: the allow %s exempted nothing", e.Term, b.Variant, a.Path))
				}
			}
		}
	}
	return out
}

// fold is the scan's ONE normalisation, applied to every text and every mask phrase before any
// pattern runs. A run of whitespace and hyphens becomes one space — "operator channel",
// "operator-channel" and a phrase wrapped across lines are one concept spelled three ways, and a
// pattern that matched only the first passed the second in a seat's constitution (#1209). A
// typographic apostrophe (U+2019) becomes an ASCII one, so "operator’s" is "operator's". A
// possessive is not otherwise folded: "a sitting's record" is plain English for something else,
// so a ban that means the possessive spells it in its own pattern.
//
// For each byte of the folded text it returns the offset of the original byte it came from and the
// offset just past the original character, so a match is reported as it was written.
func fold(s string) (string, []int, []int) {
	const rsquo = "\u2019"
	var b strings.Builder
	starts := make([]int, 0, len(s))
	ends := make([]int, 0, len(s))
	inSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '-' {
			if inSpace {
				continue
			}
			inSpace = true
			b.WriteByte(' ')
			starts = append(starts, i)
			ends = append(ends, i+1)
			continue
		}
		inSpace = false
		if strings.HasPrefix(s[i:], rsquo) {
			b.WriteByte('\'')
			starts = append(starts, i)
			ends = append(ends, i+len(rsquo))
			i += len(rsquo) - 1
			continue
		}
		b.WriteByte(c)
		starts = append(starts, i)
		ends = append(ends, i+1)
	}
	return b.String(), starts, ends
}

// blank replaces every case-insensitive occurrence of phrase (folded as the text is) with NUL bytes
// of the same length, so offsets survive and no word boundary appears inside the blanked span.
func blank(s, phrase string) (string, int) {
	p, _, _ := fold(phrase)
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(p))
	n := 0
	out := re.ReplaceAllStringFunc(s, func(m string) string {
		n++
		return strings.Repeat("\x00", len(m))
	})
	return out, n
}

func lineAt(s string, off int) int { return strings.Count(s[:off], "\n") + 1 }
