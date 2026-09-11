package hookgate

import (
	"strings"
	"testing"
)

// THE SHAPES THE SMOKE RUNS WROTE, and the ones the rule teaches instead.
//
// Every `deny` row is a command bash would have rewritten before feov-record saw it; every `allow`
// row is one bash passes through with the backtick literal. The measured rows are shortened from
// #861's transcripts (2026-09-10), where the verb exited 0 and stored the substituted text.
func TestABacktickTheShellWouldRunIsRefused(t *testing.T) {
	for _, c := range []struct {
		name, command string
		deny          bool
	}{
		// MEASURED: blue-respond's manifest row. `edit`, `retire` and `prove` ran; the row lost them.
		{"measured manifest-row, aliased binary", "B=\"/r/.bin/feov-record\"\n\"$B\" --seat-id blue-respond manifest-row --id G3 --reason \"the correct sequence is edit-to-bare-anchor then `retire` with the pre-edit quote\"", true},
		// MEASURED: red-lens-logic's finding. `factor 91` ran and its output was pasted in.
		{"measured finding", "\"/r/.bin/feov-record\" finding --quote \"91\" --reason \"confirmed with `factor 91` and `openssl prime 91`\"", true},
		{"backtick in another free-text flag", "feov-record mint --fix \"replace it with `x`\" --reason 'r'", true},
		// A backtick in COMMAND context is a substitution the seat wrote as one — #510's judge
		// captured the tool's own output this way — so it is not refused.
		{"backtick in command context is a command", "cites=`feov-record show evidence`\nfeov-record prove --cites \"$cites\" --reason 'r'", false},
		{"unquoted heredoc body runs its backticks", "X=$(cat <<EOF\nwe ran `prove`\nEOF\n)\nfeov-record position --reason \"$X\"", true},

		{"escaped backtick in double quotes", "feov-record position --reason \"uses \\`prove\\` here\"", false},
		{"single quotes are literal", "feov-record position --reason 'uses `prove` here'", false},
		// THE TAUGHT FORM, in both spellings a seat writes it.
		{"quoted heredoc captured into a variable", "X=$(cat <<'EOF'\nuses `prove`, and the report's figures\nEOF\n)\nfeov-record position --reason \"$X\"", false},
		{"quoted heredoc inline inside double quotes", "feov-record position --reason \"$(cat <<'EOF'\nuses `prove`\nEOF\n)\"", false},
		{"double-quoted delimiter", "X=$(cat <<\"EOF\"\n`prove`\nEOF\n)\nfeov-record position --reason \"$X\"", false},
		{"backslashed delimiter", "X=$(cat <<\\EOF\n`prove`\nEOF\n)\nfeov-record position --reason \"$X\"", false},
		{"dash heredoc, quoted, tab-indented", "X=$(cat <<-'EOF'\n\t`prove`\n\tEOF\n)\nfeov-record position --reason \"$X\"", false},
		{"a comment is not run", "feov-record position --reason x # see `prove`", false},
		{"$( ) is a capture, not a stray backtick", "feov-record position --reason \"$(date)\"", false},
		{"apostrophes inside double quotes", "feov-record position --reason \"the report's figures don't move\"", false},

		// HERE-STRINGS. `<<<` is not a heredoc: its word is ordinary quoting on the same line.
		{"double-quoted here-string runs its backtick", "feov-record x --reason \"$(cat)\" <<< \"it uses `prove`\"", true},
		{"a here-string does not swallow the lines after it", "feov-record x --reason \"$(cat)\" <<< \"ok\"\nfeov-record y --reason \"a `prove`\"", true},
		{"single-quoted here-string is literal", "feov-record x --reason \"$(cat)\" <<< 'it uses `prove`'", false},
		// ANSI-C QUOTES. `\'` does not end `$'…'`, and what follows it is scanned in step.
		{"escaped apostrophe in $'' then double quotes", "feov-record x --reason $'don\\'t' \"see `prove`\"", true},
		{"$'' holding a double quote, then a single-quoted backtick", "feov-record x --reason $'a\\'\"' '  b `x` '", false},
		{"$'' is literal", "feov-record x --reason $'uses `prove` here'", false},
		{"an unterminated quote refuses any backtick after it", "feov-record x --reason \"a\" 'b `prove`", true},
		// NESTED QUOTING inside text: ${ } and $( ) each open their own.
		{"double quotes nested in ${ } in double quotes", "feov-record x --reason \"${v:-\"`x`\"}\"", true},
		{"a case pattern's ) does not close $( )", "feov-record x --reason \"$(case x in x) echo \"`x`\";; esac)\"", true},
		{"a subshell's ) does not close $( )", "feov-record x --reason \"$( (echo a); echo \"`x`\" )\"", true},
		{"$(( closed by one paren is a subshell", "feov-record x --reason \"$((echo a); echo \"`x`\")\"", true},
		// ARITHMETIC. `<<` there is a shift, and opens no heredoc.
		{"arithmetic shift is not a heredoc", "x=$((1<<2))\nfeov-record x --reason 'see `x`'", false},
		{"arithmetic command shift is not a heredoc", "((x=1<<2))\nfeov-record x --reason 'see `x`'", false},
		{"arithmetic shift, then double quotes", "x=$((1<<2))\nfeov-record x --reason \"see `x`\"", true},
		// UNQUOTED HEREDOC BODIES are text: quotes literal, but $( ) opens full quoting again.
		{"single quotes inside $( ) in an unquoted body", "X=$(cat <<EOF\n$(echo '`x`')\nEOF\n)\nfeov-record x --reason \"$X\"", false},
		{"a quoted heredoc nested in an unquoted body", "X=$(cat <<EOF\n$(cat <<'I'\n`x`\nI\n)\nEOF\n)\nfeov-record x --reason \"$X\"", false},
		{"an apostrophe in an unquoted body does not hide a backtick", "X=$(cat <<EOF\ndon't `x`\nEOF\n)\nfeov-record x --reason \"$X\"", true},
		{"a comment after a redirection", "feov-record x --reason b >#out \"a `x`\"", false},
	} {
		at, runs := executedBacktick(c.command)
		if runs != c.deny {
			t.Errorf("%s: executedBacktick = %v, want %v\n%s", c.name, runs, c.deny, c.command)
			continue
		}
		if runs && c.command[at] != '`' {
			t.Errorf("%s: offset %d points at %q, not the backtick", c.name, at, c.command[at])
		}
		out, reason := PreOutcome(seatBash(t, c.command), liveRun)
		if c.deny {
			if out != OutcomeDeny {
				t.Errorf("%s: PreOutcome = %v, want a deny", c.name, out)
			}
			// THE REFUSAL TEACHES THE FORM THAT WORKS, or the seat retries the same thing.
			for _, want := range []string{"EXECUTE", "<<'EOF'", "\"$X\"", "Nothing was written"} {
				if !strings.Contains(reason, want) {
					t.Errorf("%s: the deny does not say %q:\n%s", c.name, want, reason)
				}
			}
		} else if out == OutcomeDeny {
			t.Errorf("%s: a command bash passes through was denied: %s", c.name, reason)
		}
	}
}

// THE SCOPE: only a tool command, only in a live run.
func TestTheDenyIsScopedToToolCommandsInALiveRun(t *testing.T) {
	// A backtick in a command that does not touch the tool is the seat's own business: it still
	// gets its identity injected, and nothing is refused.
	if out, _ := PreOutcome(seatBash(t, "echo \"today is `date`\" >> notes.md"), liveRun); out != OutcomeRewrite {
		t.Errorf("an unrelated command with a backtick = %v, want the ordinary rewrite", out)
	}
	// Outside a live run the hook says nothing at all, deny included — a developer at a terminal is
	// not a seat.
	if out, _ := PreOutcome(seatBash(t, "feov-record position --reason \"`x`\""), ""); out != OutcomeNone {
		t.Errorf("outside a live run = %v, want no opinion", out)
	}
}

// NO INPUT MAKES THE SCANNER PANIC OR POINT AT SOMETHING ELSE. It reads every Bash command in a live
// run, so a crash here would be a crash of every seat's every command.
func FuzzExecutedBacktick(f *testing.F) {
	for _, s := range []string{
		"feov-record position --reason \"`x`\"",
		"X=$(cat <<'EOF'\n`y`\nEOF\n)\n",
		"<<", "<<-", "<<'", "\"$(", "'", "\\", "#`", "a <<E\n`\nE", "$(((", ")))",
		"<<<", "a <<< \"`x`\"\n`y`", "$'", "$'\\'", "$'\\'\" \"`x`\"", "\"${", "\"${\"`", "}\"`",
		"\"$(case x in x) \"`\";; esac)\"", "\"$( (a); \"`\" )\"", "\"$((a); \"`\")\"", "$((1<<2))\n'`'",
		"((", "((1<<2))\n`", "<<E\n$('`')\nE", "<<E\n$(cat <<'I'\n`\nI\n)\nE", "case", "esac)", "then case x in x)",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		at, runs := executedBacktick(s)
		if runs && (at < 0 || at >= len(s) || s[at] != '`') {
			t.Fatalf("offset %d for %q does not point at a backtick", at, s)
		}
		if runs {
			_ = substitutionReason(s, at) // the excerpt math must hold at every offset
		}
	})
}
