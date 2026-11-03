package planguard

import (
	"sort"
	"strings"
)

// Scope answers what a name in a plan line stands for.
//
// # Why the statement's text is read at all
//
// SQLite's plan names a FROM item by its ALIAS: `SCAN p`, `SEARCH ss USING AUTOMATIC COVERING
// INDEX (id=?)`. Whether `p` is the growing `proof` table or an eight-row vocabulary is not in the
// plan, and the bytecode does not carry it either — the plan's node ids are addresses in the
// EXPLAIN QUERY PLAN program, which is not the program plain EXPLAIN prints. The binding exists
// only in the SQL that declared it, so it is read from there: the statement, and every view the
// statement names, transitively.
//
// It is read by tokens, not by pattern: comments and string literals are skipped, quoted
// identifiers are unquoted, and a binding is taken only from a FROM or JOIN item. An alias the
// reading cannot bind resolves to itself and is reported under that name — a finding with an
// unhelpful name is read and fixed; a finding dropped because its name was not understood is a
// clean board that is not clean.
type Scope struct {
	binds map[string]map[string]bool // alias or name -> what the FROM items bound it to
}

// subqueryTarget is what an alias over a parenthesised FROM item is bound to. It names no table,
// so it never replaces the alias in a finding.
const subqueryTarget = "(subquery)"

// NewScope reads the bindings of stmt and of every view it reaches.
func NewScope(stmt string, views []View) *Scope {
	defs := map[string]string{}
	for _, v := range views {
		defs[strings.ToLower(v.Name)] = v.SQL
	}
	s := &Scope{binds: map[string]map[string]bool{}}
	seen := map[string]bool{}
	queue := []string{stmt}
	for len(queue) > 0 {
		text := queue[0]
		queue = queue[1:]
		toks := tokenize(text)
		s.bind(toks)
		for _, t := range toks {
			if t.kind != tokIdent && t.kind != tokWord {
				continue
			}
			name := strings.ToLower(t.text)
			if def, ok := defs[name]; ok && !seen[name] {
				seen[name] = true
				queue = append(queue, def)
			}
		}
	}
	return s
}

// Targets is what the plan's word may stand for, sorted. A word bound nowhere stands for itself.
func (s *Scope) Targets(word string) []string {
	set := s.binds[strings.ToLower(word)]
	if len(set) == 0 {
		return []string{strings.ToLower(word)}
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// Name is how a finding names the plan's word: the one table or CTE it is bound to, or, where the
// word is bound to several, the word itself with every candidate — so an ambiguous alias is
// reported, visibly ambiguous, rather than guessed.
func (s *Scope) Name(word string) string {
	var named []string
	for _, t := range s.Targets(word) {
		if t != subqueryTarget {
			named = append(named, t)
		}
	}
	switch len(named) {
	case 0:
		return word
	case 1:
		return named[0]
	}
	return word + "(" + strings.Join(named, "|") + ")"
}

// bind records every FROM and JOIN item's binding in toks.
func (s *Scope) bind(toks []token) {
	for i := 0; i < len(toks); i++ {
		if toks[i].kind != tokWord {
			continue
		}
		kw := strings.ToUpper(toks[i].text)
		if kw != "FROM" && kw != "JOIN" {
			continue
		}
		next, ok := s.item(toks, i+1)
		if !ok || kw != "FROM" {
			continue
		}
		// The rest of a FROM list: a comma at this nesting level, after the item and any JOIN ...
		// ON expression, opens the next item. JOIN items are read where their own JOIN is met.
		for j := next; j < len(toks); {
			t := toks[j]
			switch {
			case t.is("("):
				j = skipGroup(toks, j)
				continue
			case t.is(")"), t.kind == tokWord && endsFromList[strings.ToUpper(t.text)]:
				j = len(toks)
				continue
			case t.is(","):
				if after, ok := s.item(toks, j+1); ok {
					j = after
					continue
				}
			}
			j++
		}
	}
}

// endsFromList are the keywords that close a FROM list.
var endsFromList = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`WHERE GROUP ORDER LIMIT HAVING WINDOW UNION EXCEPT INTERSECT RETURNING SELECT VALUES`) {
		endsFromList[w] = true
	}
}

// item reads one FROM item at toks[i] — a name or a parenthesised subquery, then an optional alias
// — records its binding, and returns the index after it.
func (s *Scope) item(toks []token, i int) (int, bool) {
	if i >= len(toks) {
		return i, false
	}
	var target string
	switch {
	case toks[i].is("("):
		i = skipGroup(toks, i)
		target = subqueryTarget
	case toks[i].kind == tokIdent || (toks[i].kind == tokWord && !reservedAfterItem[strings.ToUpper(toks[i].text)]):
		target = strings.ToLower(toks[i].text)
		i++
		if i+1 < len(toks) && toks[i].is(".") { // schema-qualified: the table is the second part
			target = strings.ToLower(toks[i+1].text)
			i += 2
		}
		if i < len(toks) && toks[i].is("(") { // a table-valued function's arguments
			i = skipGroup(toks, i)
		}
	default:
		return i, false
	}
	alias := ""
	if i < len(toks) && toks[i].kind == tokWord && strings.EqualFold(toks[i].text, "AS") {
		i++
		if i < len(toks) && (toks[i].kind == tokIdent || toks[i].kind == tokWord) {
			alias = toks[i].text
			i++
		}
	} else if i < len(toks) && (toks[i].kind == tokIdent || (toks[i].kind == tokWord && !reservedAfterItem[strings.ToUpper(toks[i].text)])) {
		alias = toks[i].text
		i++
	}
	key := target
	if alias != "" {
		key = strings.ToLower(alias)
	}
	if key != subqueryTarget {
		if s.binds[key] == nil {
			s.binds[key] = map[string]bool{}
		}
		s.binds[key][target] = true
	}
	return i, true
}

// reservedAfterItem are the keywords that can follow a FROM item, so a bare word among them ends
// the item rather than naming its alias.
var reservedAfterItem = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`WHERE JOIN LEFT RIGHT FULL INNER OUTER CROSS NATURAL ON USING
		GROUP ORDER LIMIT OFFSET UNION EXCEPT INTERSECT WINDOW HAVING INDEXED NOT AS SELECT VALUES
		RETURNING SET WHEN THEN END ELSE AND OR WITH`) {
		reservedAfterItem[w] = true
	}
}

// skipGroup returns the index after the parenthesis group opening at toks[i].
func skipGroup(toks []token, i int) int {
	depth := 0
	for ; i < len(toks); i++ {
		switch {
		case toks[i].is("("):
			depth++
		case toks[i].is(")"):
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return i
}

type tokKind int

const (
	tokWord   tokKind = iota // a bare word: an identifier or a keyword
	tokIdent                 // a quoted identifier, never a keyword
	tokString                // a string literal
	tokNumber
	tokPunct
)

type token struct {
	kind tokKind
	text string
}

func (t token) is(p string) bool { return t.kind == tokPunct && t.text == p }

// tokenize splits SQL into the tokens the binding reader needs, dropping comments.
func tokenize(src string) []token {
	var out []token
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '-' && i+1 < len(src) && src[i+1] == '-':
			for i < len(src) && src[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				i = len(src)
			} else {
				i += end + 4
			}
		case c == '\'' || c == '"' || c == '`':
			text, next := quoted(src, i, c)
			kind := tokIdent
			if c == '\'' {
				kind = tokString
			}
			out = append(out, token{kind, text})
			i = next
		case c == '[':
			end := strings.IndexByte(src[i:], ']')
			if end < 0 {
				end = len(src) - i
			}
			out = append(out, token{tokIdent, src[i+1 : i+end]})
			i += end + 1
		case isWordStart(c):
			j := i + 1
			for j < len(src) && isWordPart(src[j]) {
				j++
			}
			out = append(out, token{tokWord, src[i:j]})
			i = j
		case c >= '0' && c <= '9':
			j := i + 1
			for j < len(src) && (isWordPart(src[j]) || src[j] == '.') {
				j++
			}
			out = append(out, token{tokNumber, src[i:j]})
			i = j
		default:
			out = append(out, token{tokPunct, string(c)})
			i++
		}
	}
	return out
}

// quoted reads a quoted token opening at src[i] with quote q, where a doubled quote is a literal
// one, and returns its unquoted text and the index after it.
func quoted(src string, i int, q byte) (string, int) {
	var b strings.Builder
	for j := i + 1; j < len(src); j++ {
		if src[j] == q {
			if j+1 < len(src) && src[j+1] == q {
				b.WriteByte(q)
				j++
				continue
			}
			return b.String(), j + 1
		}
		b.WriteByte(src[j])
	}
	return b.String(), len(src)
}

func isWordStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}

func isWordPart(c byte) bool { return isWordStart(c) || (c >= '0' && c <= '9') || c == '$' }
