package trajectory

import "testing"

func TestStripHeredocsReadsTheShellGrammar(t *testing.T) {
	cases := []struct {
		name, cmd, want string
		heredocs        int
	}{
		{"quoted", "python3 - <<'PY'\nprint(1)\nPY\ngo test ./...", "python3 - <<'PY'\ngo test ./...", 1},
		{"backslash", "cat <<\\EOF\nbody\nEOF\nls", "cat <<\\EOF\nls", 1},
		{"hyphenated delimiter", "cat <<'END-X'\nbody\nEND-X\nls", "cat <<'END-X'\nls", 1},
		{"tab-stripped", "cat <<-EOF\n\tbody\n\tEOF\nls", "cat <<-EOF\nls", 1},
		{"two on one line", "cat <<A <<B\na\nA\nb\nB\nls", "cat <<A <<B\nls", 2},
		{"last line", "cat <<EOF\nbody\nEOF", "cat <<EOF", 1},
		{"herestring and shift are not heredocs", "echo $((1<<3)); cat <<<word", "echo $((1<<3)); cat <<<word", 0},
		{"quoted << is not a heredoc", `echo "a <<EOF b"`, `echo "a <<EOF b"`, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, n, err := StripHeredocs(tc.cmd)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want || n != tc.heredocs {
				t.Errorf("got %q (%d heredocs), want %q (%d)", got, n, tc.want, tc.heredocs)
			}
		})
	}
}

// An unterminated heredoc is not shell bash would run; it is an error, so no
// caller can read its body as command.
func TestAnUnterminatedHeredocIsAnError(t *testing.T) {
	if _, _, err := StripHeredocs("cat <<'PY'\ngo test ./...\n"); err == nil {
		t.Error("an unterminated heredoc parsed")
	}
	if CommandKey("cat <<'PY'\ngo test ./...\n") != "" {
		t.Error("an unparseable command was keyed")
	}
}
