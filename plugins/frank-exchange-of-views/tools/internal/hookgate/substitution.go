package hookgate

import "strings"

// A BACKTICK THE SHELL WILL RUN, found before the shell runs it.
//
// Measured over #861's smoke runs (2026-09-10): seats name the tool's verbs in backticks, because
// the tool's own help and prompts do, and they write that prose inside double quotes. Bash runs
// every such backtick as a command before feov-record sees the argument, and records the output —
// or nothing — in its place. Eight --reason values and one --fix were rewritten that way and stored
// silently; `prove` ran Perl's test harness, and `factor 91` pasted "91: 7 13" into a finding.
// Only the hook sees the command as the seat wrote it, so only the hook can refuse it intact.
//
// THIS IS A SCANNER, NOT A MATCHER, and the distinction is the one PreOutcome's note draws. The
// injection lost its matcher because a miss there was silent and destroyed a seat's identity. A
// miss here is today's behaviour — the substitution runs, as it always has — and a false hit is a
// refusal that names the text and the fix, which a seat can act on in the same turn. So the check
// errs toward precision about SHELL SEMANTICS rather than toward recognising the tool: it follows
// the quoting bash follows, and stays silent wherever bash would not run the backtick.
//
// WHAT IT HUNTS IS A BACKTICK IN TEXT: inside double quotes, inside a `${ }` within them, or in an
// unquoted heredoc body. Every measured corruption sat there. A backtick in COMMAND context —
// `cites=\`feov-record show …\`` — is a substitution the seat wrote as one, and a refusal there
// cost a measured shape its capture (an earlier cut of this scanner failed
// TestTheShapesThatOnceLostTheirIdentity on exactly that).
//
// WHAT IT FOLLOWS, as a stack of frames (see frame):
//   - single quotes (literal) and ANSI-C `$'…'` (literal, but `\'` does not end it);
//   - double quotes, where backticks run and `\“ does not, and where `$( )`, `$(( ))` and `${ }`
//     each open their own quoting — `"${v:-"x"}"` nests a second pair of double quotes;
//   - `$( )` with its own paren depth and `case … esac`, so a subshell's `)` or a case pattern's
//     `)` does not end the substitution early;
//   - arithmetic, `$(( ))` and `(( ))`, where `<<` is a shift and not a heredoc;
//   - comments, backslash escapes, here-strings (`<<<`, whose word is ordinary quoting), and
//     heredocs: a QUOTED delimiter (`<<'EOF'`, `<<"EOF"`, `<<\EOF`) makes the body literal; an
//     unquoted one makes the body behave like double-quoted text in which `"` and `'` are literal
//     but backticks, `$( )` and `${ }` still expand.
//
// That quoted-heredoc rule is the whole reason the recommended form passes:
// `"$(cat <<'EOF' … EOF)"` puts the seat's backticks inside a quoted heredoc inside a substitution
// inside double quotes.
//
// WHEN IT LOSES QUOTE SYNC — an unterminated single quote, `$'…'`, or heredoc delimiter quote — it
// does not guess. Bash would refuse such a command with a syntax error and run nothing, so refusing
// it too costs nothing, while "no backtick" would be a claim the scanner could no longer make. The
// policy is: from the point sync was lost, ANY backtick is refused.
//
// `$(` itself is never flagged. It is how the recommended form captures text, and a seat writing
// `$(` means to run something; a backtick inside prose almost never does.

// executedBacktick returns the byte offset of the first backtick in command that bash would
// execute inside text, and whether there is one.
func executedBacktick(command string) (int, bool) { return scan(command, kindCommand) }

// A frame is one quoting context the scanner is inside.
type frame struct {
	kind  byte
	depth int // unmatched `(` in a command or arithmetic frame
	cases int // open `case … esac` in a command frame, whose pattern `)` closes nothing
}

const (
	kindCommand  = 'n' // top level, or inside `$( )`
	kindDouble   = 'd' // inside double quotes
	kindBrace    = 'B' // inside `${ }` within text
	kindArith    = 'a' // inside `$(( ))` or `(( ))`
	kindHeredoc  = 'h' // an unquoted heredoc body: text, with no closing quote
	wordBoundary = " \t\n;&|()<>"
)

// reservedLead are the words after which a command still starts — `then case …` opens a case.
var reservedLead = map[string]bool{"then": true, "do": true, "else": true, "elif": true, "if": true, "while": true, "until": true, "time": true}

// scan walks s from a bottom frame of the given kind. Heredoc bodies are scanned by a recursive
// call, because bash finds a body's terminator line BEFORE it expands anything inside it.
func scan(s string, bottom byte) (int, bool) {
	n := len(s)
	st := []frame{{kind: bottom}}
	var pending []heredoc
	wordStart, cmdStart := true, true
	for i := 0; i < n; {
		c := s[i]
		top := &st[len(st)-1]

		// TEXT: double quotes, `${ }` inside text, an unquoted heredoc body.
		if top.kind == kindDouble || top.kind == kindBrace || top.kind == kindHeredoc {
			switch {
			case c == '\\':
				i += 2
			case c == '`':
				return i, true
			case strings.HasPrefix(s[i:], "$(("):
				st = append(st, frame{kind: kindArith})
				i += 3
			case strings.HasPrefix(s[i:], "$("):
				st = append(st, frame{kind: kindCommand})
				i += 2
				wordStart, cmdStart = true, true
			case strings.HasPrefix(s[i:], "${"):
				st = append(st, frame{kind: kindBrace})
				i += 2
			case c == '"' && top.kind == kindDouble:
				st = st[:len(st)-1]
				i++
				wordStart, cmdStart = false, false
			case c == '"' && top.kind == kindBrace:
				st = append(st, frame{kind: kindDouble})
				i++
			case c == '}' && top.kind == kindBrace:
				st = st[:len(st)-1]
				i++
			default:
				i++
			}
			continue
		}

		// COMMAND AND ARITHMETIC share their quoting.
		switch {
		case c == '\\':
			i += 2
			wordStart, cmdStart = false, false
			continue
		case c == '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return lostSync(s, i)
			}
			i += j + 2
			wordStart, cmdStart = false, false
			continue
		case strings.HasPrefix(s[i:], "$'"):
			j, ok := ansiCEnd(s, i+2)
			if !ok {
				return lostSync(s, i)
			}
			i = j
			wordStart, cmdStart = false, false
			continue
		case c == '"':
			st = append(st, frame{kind: kindDouble})
			i++
			continue
		case c == '`':
			// IN COMMAND CONTEXT A BACKTICK IS A COMMAND THE SEAT CHOSE TO RUN, and it is left
			// alone: `cites=\`feov-record show evidence\`` is a measured shape (#510, the judge
			// capturing the tool's own output). Skip to its closing tick.
			j := i + 1
			for j < n && s[j] != '`' {
				if s[j] == '\\' {
					j++
				}
				j++
			}
			i = j + 1
			wordStart, cmdStart = false, false
			continue
		case strings.HasPrefix(s[i:], "$(("):
			st = append(st, frame{kind: kindArith})
			i += 3
			continue
		case strings.HasPrefix(s[i:], "$("):
			st = append(st, frame{kind: kindCommand})
			i += 2
			wordStart, cmdStart = true, true
			continue
		}

		if top.kind == kindArith {
			switch {
			case c == '(':
				top.depth++
			case c == ')' && top.depth > 0:
				top.depth--
			case c == ')' && i+1 < n && s[i+1] == ')':
				st = st[:len(st)-1]
				i += 2
				wordStart, cmdStart = false, false
				continue
			case c == ')':
				// `$((…)` closed by ONE paren was never arithmetic: bash re-reads it as `$( (…) …`,
				// a subshell inside a substitution, and that subshell has just closed.
				*top = frame{kind: kindCommand}
				wordStart, cmdStart = true, false
			}
			i++
			continue
		}

		switch {
		case c == '(' && cmdStart && i+1 < n && s[i+1] == '(':
			st = append(st, frame{kind: kindArith})
			i += 2
		case c == '(':
			top.depth++
			i++
			wordStart, cmdStart = true, true
		case c == ')':
			switch {
			case top.depth > 0:
				top.depth--
				wordStart, cmdStart = true, false
			case top.cases > 0:
				wordStart, cmdStart = true, true // a case pattern: its commands follow
			case len(st) > 1:
				st = st[:len(st)-1]
				wordStart, cmdStart = false, false
			}
			i++
		case c == '#' && wordStart:
			j := strings.IndexByte(s[i:], '\n')
			if j < 0 {
				return 0, false
			}
			i += j // the newline itself is handled below, so pending heredocs still start
		case strings.HasPrefix(s[i:], "<<<"):
			// A HERE-STRING, not a heredoc: its word follows on this line under ordinary quoting,
			// so a double-quoted one runs its backticks like any other.
			i += 3
			wordStart, cmdStart = true, false
		case strings.HasPrefix(s[i:], "<<"):
			h, adv, ok := parseHeredocOp(s[i+2:])
			if !ok {
				return lostSync(s, i)
			}
			if h.delim != "" {
				pending = append(pending, h)
			}
			i += 2 + adv
			wordStart, cmdStart = false, false
		case c == '\n':
			i++
			for _, h := range pending {
				body, next := heredocBody(s, i, h)
				if !h.quoted {
					if k, runs := scan(body, kindHeredoc); runs {
						return i + k, true
					}
				}
				i = next
			}
			pending = nil
			wordStart, cmdStart = true, true
		case cmdStart && isWordByte(c):
			j := i
			for j < n && isWordByte(s[j]) {
				j++
			}
			w := s[i:j]
			ends := j == n || strings.IndexByte(wordBoundary, s[j]) >= 0
			switch {
			case ends && w == "case":
				top.cases++
			case ends && w == "esac" && top.cases > 0:
				top.cases--
			}
			cmdStart = ends && reservedLead[w]
			wordStart = false
			i = j
		default:
			switch c {
			case ' ', '\t':
			case ';', '&', '|', '{', '!':
				cmdStart = true
			default:
				cmdStart = false
			}
			wordStart = strings.IndexByte(" \t;&|(<>", c) >= 0
			i++
		}
	}
	return 0, false
}

func isWordByte(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// lostSync applies the policy for a quote the scanner cannot close: any backtick from here on.
func lostSync(s string, from int) (int, bool) {
	if k := strings.IndexByte(s[from:], '`'); k >= 0 {
		return from + k, true
	}
	return 0, false
}

// ansiCEnd returns the offset just past the `'` that closes a `$'…'` whose body starts at i. A
// backslash escapes the byte after it, so `$'don\'t'` is one word.
func ansiCEnd(s string, i int) (int, bool) {
	for ; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '\'':
			return i + 1, true
		}
	}
	return 0, false
}

type heredoc struct {
	delim  string
	quoted bool // any quoting of the delimiter makes the body literal
	dash   bool // <<- strips leading tabs from body lines and the terminator
}

// parseHeredocOp reads the delimiter after `<<`, returning it, how many bytes it consumed, and
// false when a quote in it never closes.
func parseHeredocOp(rest string) (heredoc, int, bool) {
	var h heredoc
	i := 0
	if i < len(rest) && rest[i] == '-' {
		h.dash = true
		i++
	}
	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}
	var word strings.Builder
	for i < len(rest) {
		c := rest[i]
		switch {
		case c == '\'' || c == '"':
			j := strings.IndexByte(rest[i+1:], c)
			if j < 0 {
				return heredoc{}, i, false
			}
			word.WriteString(rest[i+1 : i+1+j])
			h.quoted = true
			i += j + 2
			continue
		case c == '\\' && i+1 < len(rest):
			word.WriteByte(rest[i+1])
			h.quoted = true
			i += 2
			continue
		case strings.IndexByte(wordBoundary, c) >= 0:
			h.delim = word.String()
			return h, i, true
		}
		word.WriteByte(c)
		i++
	}
	h.delim = word.String()
	return h, i, true
}

// heredocBody returns the body starting at start and the offset just past its terminator line.
// A body with no terminator runs to the end of the command, as bash reads it.
func heredocBody(s string, start int, h heredoc) (string, int) {
	for p := start; p <= len(s); {
		eol := strings.IndexByte(s[p:], '\n')
		end := len(s)
		if eol >= 0 {
			end = p + eol
		}
		line := s[p:end]
		if h.dash {
			line = strings.TrimLeft(line, "\t")
		}
		if line == h.delim {
			next := end + 1
			if next > len(s) {
				next = len(s)
			}
			return s[start:p], next
		}
		if eol < 0 {
			break
		}
		p = end + 1
	}
	return s[start:], len(s)
}

// substitutionReason is the deny text: WHAT the shell would have run, and the one way that works
// for every free-text flag. The instruction is flags.ProseFooter's, stated where the seat is.
func substitutionReason(command string, at int) string {
	from, to := at-40, at+40
	if from < 0 {
		from = 0
	}
	if to > len(command) {
		to = len(command)
	}
	excerpt := strings.ReplaceAll(command[from:to], "\n", "⏎")
	return "feov-record: refused before it ran — bash would EXECUTE the backtick in …" + excerpt + "… " +
		"and record the command's output (or nothing) in place of your words. Nothing was written. " +
		"Pass free text by capturing it first with a QUOTED heredoc, then give the flag the variable: " +
		"X=$(cat <<'EOF' ⏎ your text, backticks and apostrophes literal ⏎ EOF ⏎ ) then \"$X\". " +
		"To keep one literal backtick inside double quotes, write \\` instead; to RUN a command and use its " +
		"output inside a string, write $( … )."
}
