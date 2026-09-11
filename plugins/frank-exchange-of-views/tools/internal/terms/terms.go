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
// A GATED ban is a phrase-exact pattern that cannot hit a legitimate sense of a word. A word with
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
		if !b.re.MatchString(m.Phrase) {
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
// Whitespace runs are folded to one space before matching, so a phrase hard-wrapped across two
// lines of markdown or a template literal is still one phrase. Line numbers are the original's.
func (r *Registry) Scan(path, text string, u *Usage) []Hit {
	norm, offs := fold(text)
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
				h := Hit{Term: e.Term, Variant: b.Variant, Match: norm[loc[0]:loc[1]], Line: lineAt(text, offs[loc[0]])}
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

// fold collapses each whitespace run to one space, returning the folded text and, for each of its
// bytes, the offset of the byte it came from.
func fold(s string) (string, []int) {
	var b strings.Builder
	offs := make([]int, 0, len(s))
	inSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if inSpace {
				continue
			}
			inSpace = true
			b.WriteByte(' ')
			offs = append(offs, i)
			continue
		}
		inSpace = false
		b.WriteByte(c)
		offs = append(offs, i)
	}
	return b.String(), offs
}

// blank replaces every case-insensitive occurrence of phrase (whitespace-folded) with NUL bytes of
// the same length, so offsets survive and no word boundary appears inside the blanked span.
func blank(s, phrase string) (string, int) {
	p, _ := fold(phrase)
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(p))
	n := 0
	out := re.ReplaceAllStringFunc(s, func(m string) string {
		n++
		return strings.Repeat("\x00", len(m))
	})
	return out, n
}

func lineAt(s string, off int) int { return strings.Count(s[:off], "\n") + 1 }
