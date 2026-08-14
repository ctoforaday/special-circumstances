package hookgate

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzInjectRunDir fuzzes the rewrite with STATED INVARIANTS, not crash-only coverage.
//
// A crash-only fuzz over a matcher buys almost nothing: the interesting failures here are
// semantic, and every one of them is silent. Each invariant below is a way the rewrite could
// be wrong while still returning a string:
//
//  1. THE VALUE ROUND-TRIPS. A shell parsing the emitted prefix must recover the run directory
//     byte-for-byte. This is the security property: if the quoting can be escaped, the prefix
//     stops being an assignment and becomes a command.
//  2. THE ORIGINAL SURVIVES. The seat's command must appear verbatim as the suffix. A rewrite
//     that alters what the seat asked for is worse than no rewrite.
//  3. IDEMPOTENCE. Applying the rewrite to its own output must not stack a second export.
//  4. NO CONTROL CHARACTERS ESCAPE. A refused value must not be emitted at all.
func FuzzInjectRunDir(f *testing.F) {
	f.Add(`cd /x && "/c/bin/feov-record" blue manifest-row --id R1-3`, `/runs/2026-08-05_smoke`)
	f.Add(`feov-record verify`, `/runs/it's here`)
	f.Add(`grep -rn "feov-record" plugins/`, `/runs/x`)
	f.Add("cd /tmp\nfeov-record blue register", `/runs/x`)
	f.Add(`export FEOV_RUN='/runs/x'; feov-record verify`, `/runs/y`)
	f.Add(`feov-record verify`, "/runs/x\nrm -rf /")
	f.Add("cat <<'END' > d.md\nfeov-record blue cite\nEND", `/runs/x`)
	f.Add(``, ``)

	f.Fuzz(func(t *testing.T, command, runDir string) {
		got, ok := injectRunDir(command, runDir)
		if !ok {
			if got != "" {
				t.Errorf("refused but returned %q — a caller checking only the string would inject it", got)
			}
			return
		}

		// (4) A rewrite implies the value was control-character free.
		for _, r := range runDir {
			if r < 0x20 || r == 0x7f {
				t.Fatalf("emitted a rewrite for a run directory containing %q", r)
			}
		}
		// (2) The seat's command survives verbatim.
		if !strings.HasSuffix(got, command) {
			t.Fatalf("the original command was altered:\ncommand=%q\ngot=%q", command, got)
		}
		// (1) The value round-trips through single-quote parsing.
		prefix := strings.TrimSuffix(got, command)
		if recovered, err := parseExportedValue(prefix); err != nil {
			t.Fatalf("the emitted prefix does not parse as one assignment (%v): %q", err, prefix)
		} else if recovered != runDir {
			t.Fatalf("value did not round-trip:\nwant %q\ngot  %q\nprefix %q", runDir, recovered, prefix)
		}
		// (3) Idempotent.
		if second, ok2 := injectRunDir(got, runDir); ok2 {
			t.Fatalf("stacked a second export:\n%q", second)
		}
	})
}

// parseExportedValue is a minimal POSIX single-quote reader for exactly the shape emitted:
// `export FEOV_RUN='<value>'; `. It exists so the invariant is checked by DECODING rather
// than by re-running the encoder's own logic, which would only prove the code agrees with
// itself. Inside single quotes every byte is literal and the only exit is a closing quote, so
// `'\”` reads as: close, escaped literal quote, reopen.
func parseExportedValue(prefix string) (string, error) {
	const head = "export FEOV_RUN='"
	if !strings.HasPrefix(prefix, head) {
		return "", errShape
	}
	s := prefix[len(head):]
	var b strings.Builder
	for {
		i := strings.IndexByte(s, '\'')
		if i < 0 {
			return "", errShape // never closed
		}
		b.WriteString(s[:i])
		s = s[i+1:]
		if strings.HasPrefix(s, `\''`) { // '\'' — an escaped literal quote, then reopen
			b.WriteByte('\'')
			s = s[3:]
			continue
		}
		if s != "; " { // the literal closed; nothing may follow but the separator
			return "", errShape
		}
		return b.String(), nil
	}
}

type shapeErr struct{}

func (shapeErr) Error() string {
	return "the prefix is not a single well-formed `export FEOV_RUN='…'; ` assignment"
}

var errShape = shapeErr{}

// The decoder must reject the shapes an escape would produce, or invariant (1) proves nothing.
func TestTheFuzzDecoderRejectsEscapes(t *testing.T) {
	for _, bad := range []string{
		`export FEOV_RUN='/x'; rm -rf /; `, // extra command after the assignment
		`export FEOV_RUN='/x`,              // unterminated
		`FEOV_RUN='/x'; `,                  // not an export
		`export FEOV_RUN='/x' '/y'; `,      // two words
	} {
		if _, err := parseExportedValue(bad); err == nil {
			t.Errorf("the decoder accepted %q, so the round-trip invariant would pass on an escape", bad)
		}
	}
	if got, err := parseExportedValue(`export FEOV_RUN='/runs/it'\''s here'; `); err != nil || got != `/runs/it's here` {
		t.Errorf("the decoder cannot read a correctly escaped value: %q, %v", got, err)
	}
}

// The tool_input shape the CLI hands back must stay valid JSON with the command replaced —
// checked here because a malformed updatedInput is silently ignored by the client.
func TestRewrittenCommandIsStillValidToolInput(t *testing.T) {
	in := bash(t, `cd /x && feov-record verify`)
	out, payload := PreOutcome(in, liveRun)
	if out != OutcomeRewrite {
		t.Fatal("expected a rewrite")
	}
	var ti map[string]any
	if err := json.Unmarshal(in.ToolInput, &ti); err != nil {
		t.Fatal(err)
	}
	ti["command"] = payload
	b, err := json.Marshal(ti)
	if err != nil {
		t.Fatalf("the rewritten tool_input does not marshal: %v", err)
	}
	if !json.Valid(b) {
		t.Error("the rewritten tool_input is not valid JSON")
	}
}
