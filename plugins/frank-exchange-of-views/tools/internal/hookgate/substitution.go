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
// WHAT IT HUNTS IS A BACKTICK IN TEXT: inside double quotes, or in an unquoted heredoc body. Every
// measured corruption sat there. A backtick in COMMAND context — `cites=\`feov-record show …\`` — is
// a substitution the seat wrote as one, and a refusal there cost a measured shape its capture (an
// earlier cut of this scanner failed TestTheShapesThatOnceLostTheirIdentity on exactly that).
//
// What it follows: single quotes (literal), double quotes (backticks run, `\“ does not), `$( )`
// nesting inside either, comments, backslash escapes, and heredocs — a QUOTED delimiter
// (`<<'EOF'`, `<<"EOF"`, `<<\EOF`) makes the body literal, an unquoted one lets backticks run.
// That last pair is the whole reason the recommended form passes: `"$(cat <<'EOF' … EOF)"` puts
// the seat's backticks inside a quoted heredoc inside a substitution inside double quotes.
//
// `$(` itself is never flagged. It is how the recommended form captures text, and a seat writing
// `$(` means to run something; a backtick inside prose almost never does.

// executedBacktick returns the byte offset of the first backtick in command that bash would
// execute, and whether there is one.
func executedBacktick(command string) (int, bool) {
	s := command
	n := len(s)
	// ctx is a stack: 'n' is command context (top level, or inside `$( )`), 'd' is inside double
	// quotes. Single quotes and heredoc bodies are skipped whole, so they need no frame.
	ctx := []byte{'n'}
	var pending []heredoc
	wordStart := true
	for i := 0; i < n; {
		c := s[i]
		if ctx[len(ctx)-1] == 'd' {
			switch {
			case c == '\\':
				i += 2
			case c == '"':
				ctx = ctx[:len(ctx)-1]
				i++
				wordStart = false
			case c == '`':
				return i, true
			case c == '$' && i+1 < n && s[i+1] == '(':
				ctx = append(ctx, 'n')
				i += 2
				wordStart = true
			default:
				i++
			}
			continue
		}
		switch {
		case c == '\\':
			i += 2
			wordStart = false
		case c == '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return 0, false // unterminated: bash would refuse the command, not run it
			}
			i += j + 2
			wordStart = false
		case c == '"':
			ctx = append(ctx, 'd')
			i++
		case c == '`':
			// IN COMMAND CONTEXT A BACKTICK IS A COMMAND THE SEAT CHOSE TO RUN, and it is left
			// alone: `cites=\`feov-record show evidence\`` is a measured shape (#510, the judge
			// capturing the tool's own output). Skip to its closing tick. What the scanner hunts is
			// a backtick inside TEXT — double quotes, an unquoted heredoc body — where it is almost
			// always a name the seat meant literally.
			j := i + 1
			for j < n && s[j] != '`' {
				if s[j] == '\\' {
					j++
				}
				j++
			}
			i = j + 1
			wordStart = false
		case c == '$' && i+1 < n && s[i+1] == '(':
			ctx = append(ctx, 'n')
			i += 2
			wordStart = true
		case c == ')' && len(ctx) > 1:
			ctx = ctx[:len(ctx)-1]
			i++
			wordStart = false
		case c == '#' && wordStart:
			j := strings.IndexByte(s[i:], '\n')
			if j < 0 {
				return 0, false
			}
			i += j // the newline itself is handled below, so pending heredocs still start
		case c == '<' && strings.HasPrefix(s[i:], "<<") && !strings.HasPrefix(s[i:], "<<<"):
			h, adv := parseHeredocOp(s[i+2:])
			if h.delim != "" {
				pending = append(pending, h)
			}
			i += 2 + adv
			wordStart = false
		case c == '\n':
			i++
			for _, h := range pending {
				body, next := heredocBody(s, i, h)
				if !h.quoted {
					if k := unescapedBacktick(body); k >= 0 {
						return i + k, true
					}
				}
				i = next
			}
			pending = nil
			wordStart = true
		default:
			wordStart = c == ' ' || c == '\t' || c == ';' || c == '&' || c == '|' || c == '('
			i++
		}
	}
	return 0, false
}

type heredoc struct {
	delim  string
	quoted bool // any quoting of the delimiter makes the body literal
	dash   bool // <<- strips leading tabs from body lines and the terminator
}

// parseHeredocOp reads the delimiter after `<<`, returning it and how many bytes it consumed.
func parseHeredocOp(rest string) (heredoc, int) {
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
				return heredoc{}, i
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
		case strings.IndexByte(" \t\n;&|<>()", c) >= 0:
			h.delim = word.String()
			return h, i
		}
		word.WriteByte(c)
		i++
	}
	h.delim = word.String()
	return h, i
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

// unescapedBacktick is the first backtick in an unquoted heredoc body that a backslash does not
// escape, or -1.
func unescapedBacktick(body string) int {
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '\\':
			i++
		case '`':
			return i
		}
	}
	return -1
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
