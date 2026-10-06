// Package anchor is the immortal-anchor vocabulary: how an anchor is spelled in the report, what
// KIND of thing each id names, and how to read the report at one.
//
// # The kinds table
//
// Every kind of anchor is one row of `kinds`, and the row holds what the kind MEANS: its id's
// prefix and its token's tag (the spelling), the noun a message uses, what the assembly does with
// it, whether it makes its sentence a counted claim, and whether it is evidence standing behind
// its sentence. No column says how an anchor of the kind is placed, carried, protected or retired:
// every kind lives the same way, so a reader of that lifecycle walks every row alike. Each anchor
// is minted by its own verb — `lens finding`, `blue cite`, `blue prove`, `lens corroborate` — and
// read back by the edit guard, the board, the assembly and every seat that wants the live text at
// one.
//
// # Why it is a leaf
//
// That is a vocabulary several layers share, so it depends on NOTHING but the standard library.
// It was in `internal/bluedoc`, which reaches up into `internal/cli/lens` for its span locator;
// anything importing it inherited a command package, and `internal/cli/seat` could not import it at
// all without a cycle.
//
// The enforcement stayed where it belongs: `bluedoc` still owns the rule that an edit may carry an
// anchor but never drop, duplicate or invent one. This package owns only what an anchor IS.
package anchor

import (
	"regexp"
	"strings"
)

// Assembly is what the final report's assembly does with an anchor of a kind.
type Assembly int

const (
	// Strip takes the token out.
	Strip Assembly = iota
	// WeaveSource rewrites the token to a visible [^N] footnote on its source.
	WeaveSource
	// WeaveProof rewrites the token to a visible [^P<n>] reference to its computation.
	WeaveProof
)

// kind is one row of the kinds table: what an anchor of the kind MEANS.
type kind struct {
	name     string // Kind's answer
	prefix   string // the id's prefix; the class is carried by it
	tag      string // the token's tag: <!--TAG:ID-->
	label    string // the noun a message names an anchor of this kind by
	note     string // what the message adds after the id
	assembly Assembly
	claim    bool // a token of this kind makes the prose before it a counted claim
	backs    bool // the anchor is evidence standing behind its sentence
}

// kinds is the table, in the order every walk reports its kinds.
//
// A token is an HTML comment, not a footnote: a "[^id]" finding marker rendered as an undefined
// footnote AND red audited it as one. A comment renders as nothing and is no footnote, so no seat
// audits it.
var kinds = []kind{
	{name: "finding", prefix: "f-", tag: "fx", label: "finding-marker", assembly: Strip},
	{name: "citation", prefix: "c-", tag: "cite", label: "citation anchor",
		note:     " (citations are tool-managed — a cited claim leaves by blue's `edit` down to the bare anchor, then blue's `retire`, which takes the anchor out with it; never by a raw edit)",
		assembly: WeaveSource, claim: true, backs: true},
	{name: "proof", prefix: "p-", tag: "proof", label: "proof anchor",
		note:     " (a computation backs this sentence — the script and its output are cached; the claim leaves by blue's `edit` down to the bare anchor, then blue's `retire`, which takes the anchor out with it; never by a raw edit)",
		assembly: WeaveProof, backs: true},
}

const tokenOpen, tokenClose = "<!--", "-->"

// claimed is the row whose prefix id carries, or nil.
func claimed(id string) *kind {
	for i := range kinds {
		if strings.HasPrefix(id, kinds[i].prefix) {
			return &kinds[i]
		}
	}
	return nil
}

// rowOf is the row an id names. An id no row claims reads as a finding.
func rowOf(id string) *kind {
	if k := claimed(id); k != nil {
		return k
	}
	return &kinds[0]
}

// Token rebuilds the literal token for an anchor id, so a message can quote what must be
// reproduced verbatim.
//
// THE CLASS IS CARRIED BY THE ID's PREFIX, which is the one string-encoded fact here that is
// load-bearing on purpose: the minting verbs choose the prefix, and every reader — this function,
// Label, the edit guard's sweep — recovers the class from it rather than from a field. It is
// tolerable because the id and its token are minted together and never travel apart, and because
// an unknown prefix falls to the finding class rather than to a plausible zero.
func Token(id string) string { return tokenOpen + rowOf(id).tag + ":" + id + tokenClose }

// IDPattern is the one id matcher: a regular expression, unanchored and with no capture group,
// matching an anchor id of any kind in the table.
func IDPattern() string {
	alt := make([]string, len(kinds))
	for i, k := range kinds {
		alt[i] = regexp.QuoteMeta(k.prefix)
	}
	return `(?:` + strings.Join(alt, "|") + `)[0-9a-f]+`
}

// IDs are the anchor ids in s, in order, deduplicated.
//
// IT READS WHAT Token WRITES, and it is here rather than in the caller for that reason: the token
// format has one home. tokenLenAt recognises the table's kinds and nothing else; this walks with it.
func IDs(s string) []string {
	var out []string
	seen := map[string]bool{}
	Each(s, func(_, _ int, id string) {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	})
	return out
}

// Each visits every anchor token in s, in order, with its byte span and its id.
func Each(s string, visit func(start, end int, id string)) {
	for i := 0; i < len(s); {
		n := tokenLenAt(s, i)
		if n == 0 {
			i++
			continue
		}
		visit(i, i+n, idAt(s, i, n))
		i += n
	}
}

// Replace returns s with every anchor token replaced by repl(token, id).
func Replace(s string, repl func(token, id string) string) string {
	var b strings.Builder
	last := 0
	Each(s, func(start, end int, id string) {
		b.WriteString(s[last:start])
		b.WriteString(repl(s[start:end], id))
		last = end
	})
	if last == 0 {
		return s
	}
	b.WriteString(s[last:])
	return b.String()
}

// strippedToken is a token of a kind the assembly strips, whatever stands between its tag and its
// close: a seat that copies a token into record prose may elide its id ("<!--fx:…-->"), and the
// assembled report ships no such token either.
var strippedToken = func() *regexp.Regexp {
	var alt []string
	for _, k := range kinds {
		if k.assembly == Strip {
			alt = append(alt, regexp.QuoteMeta(k.tag))
		}
	}
	return regexp.MustCompile(regexp.QuoteMeta(tokenOpen) + `(?:` + strings.Join(alt, "|") + `):[^>]*` + regexp.QuoteMeta(tokenClose))
}()

// StripAssembled removes every token of a kind the assembly strips.
func StripAssembled(md string) string { return strippedToken.ReplaceAllString(md, "") }

// idAt is the id inside the token of length n beginning at i.
func idAt(s string, i, n int) string {
	tok := s[i : i+n]
	for _, k := range kinds {
		if p := tokenOpen + k.tag + ":"; strings.HasPrefix(tok, p) {
			return strings.TrimSuffix(tok[len(p):], tokenClose)
		}
	}
	return ""
}

// Kind is the kind an anchor id names: a row's name, read off the id's prefix.
func Kind(id string) string { return rowOf(id).name }

// AssemblyOf is what the assembly does with an anchor of id's kind.
func AssemblyOf(id string) Assembly { return rowOf(id).assembly }

// CountsAsClaim reports whether an anchor of id's kind, attached to prose, makes it a counted
// claim.
func CountsAsClaim(id string) bool { return rowOf(id).claim }

// Backs reports whether an anchor of id's kind is evidence standing behind its sentence.
func Backs(id string) bool { return rowOf(id).backs }

// Label describes an anchor id by its kind, so a seat is told which KIND of anchor its edit would
// have disturbed. A generic name is passed through unchanged.
func Label(id string) string {
	k := claimed(id)
	if k == nil {
		return id
	}
	return k.label + " " + id + k.note
}

// Kinds are the names of every kind in the table, in its order.
func Kinds() []string {
	out := make([]string, len(kinds))
	for i, k := range kinds {
		out[i] = k.name
	}
	return out
}

// SkipRun returns the index just past a maximal run of anchor tokens starting at i in s, or i
// itself when no token starts there.
//
// WHY A RUN AND NOT A TOKEN. Anchors abut: `verification<!--fx:f-e4bc25ec--><!--fx:f-73a56bd3-->`
// is a real span from the 2026-08-22 corpus, where two lenses anchored the same sentence. A
// caller stepping over "the anchor" would step over one of two and land inside the layer it
// was trying to see past.
func SkipRun(s string, i int) int {
	for i >= 0 && i < len(s) {
		n := tokenLenAt(s, i)
		if n == 0 {
			break
		}
		i += n
	}
	return i
}

// tokenLenAt is the length of the anchor token beginning at i, or 0. It recognises a token of a
// kind in the table whose id carries that kind's prefix and then hex, and nothing else: a stray
// HTML comment is not an anchor.
func tokenLenAt(s string, i int) int {
	if i < 0 || i >= len(s) {
		return 0
	}
	rest := s[i:]
	if !strings.HasPrefix(rest, tokenOpen) {
		return 0
	}
	body := rest[len(tokenOpen):]
	id, ok := "", false
	for _, k := range kinds {
		if strings.HasPrefix(body, k.tag+":") {
			id, ok = body[len(k.tag)+1:], true
			if !strings.HasPrefix(id, k.prefix) {
				return 0
			}
			id = id[len(k.prefix):]
			break
		}
	}
	if !ok {
		return 0
	}
	end := strings.Index(id, tokenClose)
	if end <= 0 {
		return 0
	}
	for i := 0; i < end; i++ {
		if c := id[i]; !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return 0
		}
	}
	return len(rest) - len(id) + end + len(tokenClose)
}

// Sentences splits text into its sentences: the [start, end) spans between boundaries, in order.
// A boundary is `.`, `!`, `?` or a newline outside an HTML comment — the `!` in `<!--` ends no
// sentence, and an anchor stays whole inside the sentence it sits in. A run of boundaries is one
// break. A text that opens or closes on a boundary has an empty first or last span, so the spans
// and the boundary runs alternate and every byte is in exactly one of them.
//
// It is the one splitter every reader of a sentence around an anchor reads: the claim count, the
// retire tidy, and the reopened set an edit records.
func Sentences(text string) [][2]int {
	boundary := make([]bool, len(text))
	for i := 0; i < len(text); {
		if strings.HasPrefix(text[i:], tokenOpen) {
			if j := strings.Index(text[i:], tokenClose); j >= 0 {
				i += j + len(tokenClose)
				continue
			}
		}
		switch text[i] {
		case '.', '!', '?', '\n':
			boundary[i] = true
		}
		i++
	}
	var out [][2]int
	start := 0
	for i, b := range boundary {
		if !b {
			continue
		}
		if i == 0 || !boundary[i-1] {
			out = append(out, [2]int{start, i})
		}
		start = i + 1
	}
	return append(out, [2]int{start, len(text)})
}
