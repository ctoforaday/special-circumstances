package claims

import (
	"strings"
	"testing"
)

// The note format as prosthetic-conscience actually writes it, taken from a real
// sealed note rather than invented — including the wrapped continuation lines and
// the prose entry with no command, both of which broke earlier drafts.
//
// RENUMBERED 2026-07-31. This fixture was copied from a real note that was itself
// off-spec: labels 0, 1, 7 sitting in positions 1, 2, 3. Copying it faithfully
// carried the defect into the test data, which is why no test caught the
// off-by-one that made rearmed.json's key 2 mean the note's `1.` (#192). A
// fixture inherits its author's idea of what the format is; this one inherited a
// wrong one from a real file, which is the harder version of that trap.
const realNote = "---\nschema: 2\n---\n" + `
## Validation loop
1. ` + "`go test -C plugins/gray-area/tools ./...`" + `  → 3 packages ok  · re-armed by: plugins/gray-area/tools/
   last run: pass 2026-07-30T04:41Z (on merged main, 667c716)
2. ` + "`node .github/validate-json.mjs`" + `  → 16 files valid  · re-armed by: plugins/
   last run: pass 2026-07-30T04:03Z
3. The continuity loop itself: SessionStart digest arrives, watchPaths registers, a write under a
   watched surface produces rearmed.json naming the right check  · re-armed by: .claude/settings.local.json
   last run: PASS 2026-07-30T03:23Z — check 1 re-armed by an ` + "`add`" + ` event

## Next intended steps
1. Something that is not a validation claim and must not be parsed as one
`

func parse(t *testing.T, body string) []Claim {
	t.Helper()
	cs, err := Parse(strings.NewReader(body), "CHECKPOINT.md")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return cs
}

func TestParsesTheRealNoteFormat(t *testing.T) {
	cs := parse(t, realNote)
	if len(cs) != 3 {
		t.Fatalf("got %d claims, want 3: %+v", len(cs), cs)
	}

	if got, want := cs[0].Cmd, "go test -C plugins/gray-area/tools ./..."; got != want {
		t.Errorf("cmd = %q, want %q", got, want)
	}
	if got, want := cs[0].Surface, "plugins/gray-area/tools/"; got != want {
		t.Errorf("surface = %q, want %q", got, want)
	}
	if got, want := cs[0].At, "2026-07-30T04:41Z"; got != want {
		t.Errorf("at = %q, want %q", got, want)
	}
	// Every claim cites its own line, or nothing downstream can be checked.
	for _, c := range cs {
		if c.Line <= 0 || c.File == "" {
			t.Errorf("claim %s carries no provenance: %+v", c.Index, c)
		}
	}
}

// The third entry is prose. It must survive parsing AND refuse to pretend it
// has a command — a fabricated one would produce a confident false negative.
func TestAProseClaimIsKeptButHasNoCommand(t *testing.T) {
	cs := parse(t, realNote)
	prose := cs[2]
	if prose.Index != "3" {
		t.Fatalf("expected entry 3, got %s", prose.Index)
	}
	if prose.HasCommand() {
		t.Errorf("a prose check was given a command to look for: %q", prose.Cmd)
	}
	if prose.At != "2026-07-30T03:23Z" {
		t.Errorf("at = %q — the claimed time is on the continuation line and must still be read", prose.At)
	}
}

// Steps after the loop's heading are NOT claims. The "Next intended steps" list
// is numbered identically, so a parser that keys only on "N." swallows it.
func TestTheSectionEndsAtTheNextHeading(t *testing.T) {
	for _, c := range parse(t, realNote) {
		if strings.Contains(c.Text, "not a validation claim") {
			t.Fatalf("parsed past the section end: %q", c.Text)
		}
	}
}

// THE DRIFT ALARM. The note format belongs to another Go module, so this parser
// can and will fall behind it. Reporting "no claims" would read as "nothing to
// check" when the truth is "could not read the claims" — the same
// silence-means-success failure the rest of this suite keeps finding.
func TestAnUnreadableLoopIsAnErrorNotAnEmptyResult(t *testing.T) {
	drifted := "## Validation loop\n- item written some new way\n- another\n"
	if _, err := Parse(strings.NewReader(drifted), "CHECKPOINT.md"); err == nil {
		t.Fatal("a validation-loop heading with no parsable entries was accepted as empty")
	} else if !strings.Contains(err.Error(), "drifted") {
		t.Errorf("the error does not explain what happened: %v", err)
	}
}

// A note with no validation loop at all is a different thing from a broken one,
// and must not raise the alarm.
func TestANoteWithNoLoopIsNotAnError(t *testing.T) {
	cs, err := Parse(strings.NewReader("# notes\n\nsome prose\n"), "x.md")
	if err != nil {
		t.Fatalf("a note with no loop errored: %v", err)
	}
	if len(cs) != 0 {
		t.Errorf("invented %d claims", len(cs))
	}
}

// The timestamp must come from the "last run" line specifically. An entry that
// mentions another date would otherwise have it read as the run time — a wrong
// answer that looks exactly like a right one.
func TestTheTimestampComesFromTheLastRunLineOnly(t *testing.T) {
	body := "## Validation loop\n1. `go test ./...` — added 2020-01-01T00:00Z, re-armed by: x/\n   last run: pass 2026-07-30T04:41Z\n"
	cs := parse(t, body)
	if got, want := cs[0].At, "2026-07-30T04:41Z"; got != want {
		t.Errorf("at = %q, want %q — a date elsewhere in the entry was mistaken for the run time", got, want)
	}
}

// Outcome is kept as written. "FAILING, see #165" is a legitimate thing for a
// note to say, and flattening it to a boolean loses the reference a reader needs.
func TestOutcomeIsKeptVerbatimRatherThanFlattened(t *testing.T) {
	body := "## Validation loop\n1. `go test ./...` · re-armed by: x/\n   last run: **FAILING, see #165.** The mechanism is dead\n"
	cs := parse(t, body)
	if !strings.Contains(cs[0].Outcome, "#165") {
		t.Errorf("outcome = %q — the issue reference was dropped", cs[0].Outcome)
	}
}

func TestSignatureKeepsTheDistinguishingTokens(t *testing.T) {
	for _, c := range []struct {
		cmd  string
		want []string
	}{
		{"go test -C plugins/gray-area/tools ./...", []string{"go", "test", "plugins/gray-area/tools"}},
		{"node .github/validate-json.mjs", []string{"node", ".github/validate-json.mjs"}},
		{"node scripts/rule-sweep.mjs --selftest", []string{"node", "scripts/rule-sweep.mjs"}},
	} {
		got := Signature(c.cmd)
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("Signature(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

// A signature must not be so broad that it matches anything. `go` alone would
// make every go command in the transcript count as evidence for every go check.
func TestSignatureIsNotSoBroadItMatchesAnything(t *testing.T) {
	sig := Signature("go test -C plugins/gray-area/tools ./...")
	if len(sig) < 2 {
		t.Fatalf("signature %v is too thin to distinguish one check from another", sig)
	}
	if matches("go build ./...", sig) {
		t.Error("`go build` was accepted as evidence for a `go test` claim")
	}
	if matches("go test -C plugins/frank-exchange-of-views/tools ./...", sig) {
		t.Error("another module's test run was accepted as evidence for this one")
	}
}
