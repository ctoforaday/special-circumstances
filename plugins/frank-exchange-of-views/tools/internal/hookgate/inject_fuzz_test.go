package hookgate

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzInjectEnv fuzzes the rewrite with STATED INVARIANTS, not crash-only coverage.
//
// A crash-only fuzz over a matcher buys almost nothing: the interesting failures here are
// semantic, and every one of them is silent. Each invariant below is a way the rewrite could
// be wrong while still returning a string:
//
//  1. EVERY VALUE ROUND-TRIPS. A shell parsing the emitted prefix must recover each value
//     byte-for-byte. This is the security property: if the quoting can be escaped, the prefix
//     stops being an assignment and becomes a command.
//  2. THE ORIGINAL SURVIVES. The seat's command must appear verbatim as the suffix. A rewrite
//     that alters what the seat asked for is worse than no rewrite.
//  3. IDEMPOTENCE. Applying the rewrite to its own output must not stack a second export.
//  4. NO CONTROL CHARACTERS ESCAPE. A refused value must not be emitted at all.
//
// TWO VALUES NOW, AND THE SECOND IS WHY THE PER-VARIABLE ARM EXISTS. The identity is injected
// beside the run directory, and a command that already carries one must still receive the other
// — so idempotence is checked per variable rather than over the whole prefix. An all-or-nothing
// version passes this fuzz and ships a seat with no identity whenever it has a run directory.
func FuzzInjectEnv(f *testing.F) {
	f.Add(`cd /x && "/c/bin/feov-record" blue manifest-row --id R1-3`, `/runs/2026-08-05_smoke`, `agent_017a`)
	f.Add(`feov-record verify`, `/runs/it's here`, `agent_01`)
	f.Add(`grep -rn "feov-record" plugins/`, `/runs/x`, `agent_01`)
	f.Add("cd /tmp\nfeov-record blue register", `/runs/x`, `agent_01`)
	f.Add(`export FEOV_RUN='/runs/x'; feov-record verify`, `/runs/y`, `agent_01`)
	f.Add(`feov-record verify`, "/runs/x\nrm -rf /", `agent_01`)
	f.Add(`feov-record verify`, `/runs/x`, "agent\nrm -rf /")
	f.Add("cat <<'END' > d.md\nfeov-record blue cite\nEND", `/runs/x`, `agent_01`)
	f.Add(`feov-record verify`, ``, `agent_01`)
	f.Add(`feov-record verify`, `/runs/x`, ``)
	f.Add(``, ``, ``)

	f.Fuzz(func(t *testing.T, command, runDir, agentID string) {
		vars := [][2]string{{"FEOV_RUN", runDir}, {"FEOV_AGENT_ID", agentID}}
		got, ok := injectEnv(command, vars)
		if !ok {
			if got != "" {
				t.Errorf("refused but returned %q — a caller checking only the string would inject it", got)
			}
			return
		}

		// (2) The seat's command survives verbatim.
		if !strings.HasSuffix(got, command) {
			t.Fatalf("the original command was altered:\ncommand=%q\ngot=%q", command, got)
		}
		prefix := strings.TrimSuffix(got, command)

		// (1) + (4), per variable. A value that was emitted must round-trip and must have been
		// control-character free; a value that was NOT emitted must be absent entirely.
		for _, kv := range vars {
			name, value := kv[0], kv[1]
			recovered, err := parseExportedValue(prefix, name)
			if err != nil {
				// Not emitted. Legitimate only when there was nothing to emit, the command
				// already carried it, or the value was refused.
				if value != "" && !strings.Contains(command, "export "+name+"=") && !hasControl(value) {
					t.Fatalf("%s was droppable but should have been injected: value=%q prefix=%q", name, value, prefix)
				}
				continue
			}
			if hasControl(value) {
				t.Fatalf("emitted %s for a value containing a control character: %q", name, value)
			}
			if recovered != value {
				t.Fatalf("%s did not round-trip:\nwant %q\ngot  %q\nprefix %q", name, value, recovered, prefix)
			}
		}

		// (3) Idempotent, per variable: re-running must not stack anything already present.
		if second, ok2 := injectEnv(got, vars); ok2 {
			t.Fatalf("stacked a second export:\n%q", second)
		}
	})
}

// parseExportedValue reads one named assignment out of the emitted prefix, and it validates the
// WHOLE prefix before answering.
//
// The strictness is the entire point and is easy to lose when generalising: an earlier version
// returned as soon as it found the name it wanted, so `export FEOV_RUN='/x'; rm -rf /; ` parsed
// as a clean `/x` and invariant (1) proved nothing about the bytes after it. A decoder that stops
// at the first match cannot see an escape that lands later in the same prefix.
//
// It exists so the invariant is checked by DECODING rather than by re-running the encoder's own
// logic, which would only prove the code agrees with itself. Inside single quotes every byte is
// literal and the only exit is a closing quote, so `'\”` reads as: close, escaped literal quote,
// reopen.
func parseExportedValue(prefix, name string) (string, error) {
	found, ok := "", false
	s := prefix
	for s != "" {
		const kw = "export "
		if !strings.HasPrefix(s, kw) {
			return "", errShape
		}
		rest := s[len(kw):]
		eq := strings.IndexByte(rest, '=')
		if eq < 0 {
			return "", errShape
		}
		got := rest[:eq]
		rest = rest[eq+1:]
		if !strings.HasPrefix(rest, "'") {
			return "", errShape
		}
		value, tail, err := readQuoted(rest[1:])
		if err != nil {
			return "", err
		}
		if got == name {
			found, ok = value, true
		}
		s = tail
	}
	if !ok {
		return "", errShape
	}
	return found, nil
}

// readQuoted consumes a single-quoted literal and its `; ` separator, returning the value and
// whatever follows.
func readQuoted(s string) (value, tail string, err error) {
	var b strings.Builder
	for {
		i := strings.IndexByte(s, '\'')
		if i < 0 {
			return "", "", errShape // never closed
		}
		b.WriteString(s[:i])
		s = s[i+1:]
		if strings.HasPrefix(s, `\''`) { // '\'' — an escaped literal quote, then reopen
			b.WriteByte('\'')
			s = s[3:]
			continue
		}
		if !strings.HasPrefix(s, "; ") { // the literal closed; only the separator may follow
			return "", "", errShape
		}
		return b.String(), s[2:], nil
	}
}

type shapeErr struct{}

func (shapeErr) Error() string {
	return "the prefix is not a sequence of well-formed `export NAME='…'; ` assignments"
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
		if _, err := parseExportedValue(bad, "FEOV_RUN"); err == nil {
			t.Errorf("the decoder accepted %q, so the round-trip invariant would pass on an escape", bad)
		}
	}
	if got, err := parseExportedValue(`export FEOV_RUN='/runs/it'\''s here'; `, "FEOV_RUN"); err != nil || got != `/runs/it's here` {
		t.Errorf("the decoder cannot read a correctly escaped value: %q, %v", got, err)
	}
	// AND IT MUST READ THE SECOND ASSIGNMENT, or the identity half of every invariant above is
	// checked against a prefix the decoder silently could not see.
	two := `export FEOV_RUN='/runs/x'; export FEOV_AGENT_ID='agent_01'; `
	if got, err := parseExportedValue(two, "FEOV_AGENT_ID"); err != nil || got != "agent_01" {
		t.Errorf("the decoder cannot read past the first assignment: %q, %v", got, err)
	}
	if got, err := parseExportedValue(two, "FEOV_RUN"); err != nil || got != "/runs/x" {
		t.Errorf("the decoder lost the first assignment when a second followed: %q, %v", got, err)
	}
	if _, err := parseExportedValue(two, "FEOV_NOT_THERE"); err == nil {
		t.Error("the decoder invented a value for a name the prefix does not carry")
	}
}

var _ = json.Marshal
