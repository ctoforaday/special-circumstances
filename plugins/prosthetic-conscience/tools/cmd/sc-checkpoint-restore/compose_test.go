package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// clamp AT the boundary. The existing oversize test uses a note far past the cap and a
// tolerance of maxDigest+200, so it cannot distinguish `<=` from `<`; a mutation sweep
// flipped it with the suite green. One character either side of the cap is the whole
// question — and this function exists because an earlier version let text appended
// downstream escape the cap entirely.
func TestClampBoundary(t *testing.T) {
	const path = "projects/x/CHECKPOINT.md"

	exact := strings.Repeat("y", maxDigest)
	if got := clamp(exact, path); got != exact {
		t.Errorf("text of exactly maxDigest was modified: len=%d, truncated=%v", len(got), strings.Contains(got, "truncated"))
	}
	if got := clamp(exact[:maxDigest-1], path); got != exact[:maxDigest-1] {
		t.Error("text one char under the cap was modified")
	}

	over := clamp(exact+"z", path)
	if !strings.Contains(over, "truncated at") {
		t.Error("one char over the cap must truncate, and say so")
	}
	if !strings.Contains(over, path) {
		t.Error("a truncated digest must still point at the full note — otherwise the dropped content is unreachable")
	}
	if !strings.HasPrefix(over, exact) {
		t.Error("truncation must keep the head verbatim")
	}

	if got := clamp("", path); got != "" {
		t.Errorf("empty stays empty, got %q", got)
	}
}

// emit's two independent reasons to speak. The condition is
// `TrimSpace(text) == "" && len(watch) == 0`, and a mutation sweep flipped it with
// nothing failing — so nothing pinned the case the comment above it exists to explain:
// a note may be worth WATCHING even when there is nothing to say about it.
func TestEmitSpeaksWhenEitherHalfIsPresent(t *testing.T) {
	decode := func(t *testing.T, raw string) (text string, watch []string, spoke bool) {
		t.Helper()
		if strings.TrimSpace(raw) == "" {
			return "", nil, false
		}
		var out struct {
			HookSpecificOutput struct {
				HookEventName     string   `json:"hookEventName"`
				AdditionalContext string   `json:"additionalContext"`
				WatchPaths        []string `json:"watchPaths"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			t.Fatalf("emitted non-JSON %q: %v", raw, err)
		}
		if out.HookSpecificOutput.HookEventName != "SessionStart" {
			t.Errorf("hookEventName = %q, want SessionStart", out.HookSpecificOutput.HookEventName)
		}
		return out.HookSpecificOutput.AdditionalContext, out.HookSpecificOutput.WatchPaths, true
	}

	cases := []struct {
		name  string
		text  string
		watch []string
		spoke bool
	}{
		{"nothing at all stays silent", "", nil, false},
		{"whitespace-only text with no watch stays silent", "  \n\t ", nil, false},
		{"text alone speaks", "recovered state", nil, true},
		// The case the comment promises and no test held: worth watching, nothing to say.
		{"watch alone speaks", "", []string{"tools"}, true},
		{"whitespace text plus a watch still speaks", "   ", []string{"tools"}, true},
		{"both speak", "recovered state", []string{"tools"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var buf bytes.Buffer
			emit(&buf, c.text, c.watch)
			text, watch, spoke := decode(t, buf.String())
			if spoke != c.spoke {
				t.Fatalf("spoke=%v, want %v (emitted %q)", spoke, c.spoke, buf.String())
			}
			if !spoke {
				return
			}
			if text != c.text {
				t.Errorf("additionalContext = %q, want %q verbatim", text, c.text)
			}
			if len(watch) != len(c.watch) {
				t.Errorf("watchPaths = %v, want %v", watch, c.watch)
			}
		})
	}
}

// digest speaks when EITHER half is present: a section, or an objective. The condition
// is `len(parts) == 0 && objective == ""`, and a mutation sweep turned it into an `||`
// with nothing failing — under which a note carrying only one of the two goes silent,
// and silence is indistinguishable from "no checkpoint exists".
func TestDigestSpeaksOnEitherHalfAlone(t *testing.T) {
	empty := checkpoint.RearmState{}

	// Sections, no objective at all.
	sectionsOnly := "---\nschema: 2\n---\n## Next intended steps\n1. wire hooks.json\n"
	if got := digest(sectionsOnly, "p/CHECKPOINT.md", "startup", empty); got == "" {
		t.Error("a note with sections but no objective produced silence")
	} else if !strings.Contains(got, "wire hooks.json") {
		t.Errorf("the section body did not survive into the digest: %q", got)
	}

	// An objective, no sections at all.
	objectiveOnly := "---\nschema: 2\nobjective: prove the restore path fires\n---\n"
	got := digest(objectiveOnly, "p/CHECKPOINT.md", "startup", empty)
	if got == "" {
		t.Fatal("a note with an objective but no sections produced silence")
	}
	// And the objective is actually STATED — the `if objective != ""` branch that
	// prints it was itself unpinned.
	if !strings.Contains(got, "prove the restore path fires") {
		t.Errorf("the objective was not carried into the digest: %q", got)
	}

	// Neither: genuine silence, which must stay silent rather than emit a header.
	if got := digest("---\nschema: 2\n---\n", "p/CHECKPOINT.md", "startup", empty); got != "" {
		t.Errorf("a note with neither an objective nor a section must produce nothing, got %q", got)
	}
}

// pointer's objective fallback. The `obj == ""` branch was unpinned — a mutation sweep
// inverted it with the suite green, which would print the literal placeholder in place
// of a real objective. The pointer line is all the session gets in this mode, so the
// objective is the only thing distinguishing "checkpoint about X exists" from a bare
// path the operator has no reason to open.
func TestPointerObjectiveFallback(t *testing.T) {
	const path = "projects/x/CHECKPOINT.md"

	withObj := pointer("---\nobjective: ship the restore hook\n---\n## Open threads\n- none\n", path, "status: done")
	if !strings.Contains(withObj, `"ship the restore hook"`) {
		t.Errorf("a recorded objective must be quoted into the pointer: %q", withObj)
	}
	if strings.Contains(withObj, "no objective recorded") {
		t.Errorf("the fallback fired over a real objective: %q", withObj)
	}

	for _, raw := range []string{
		"## Open threads\n- none\n",                    // no frontmatter at all
		"---\nschema: 2\n---\n## Open threads\n",       // frontmatter without the key
		"---\nobjective:\n---\n## Open threads\n",      // key present, value empty
		"---\nobjective: \"\"\n---\n## Open threads\n", // explicitly empty
	} {
		got := pointer(raw, path, "status: done")
		if !strings.Contains(got, "no objective recorded") {
			t.Errorf("missing objective must be named as missing, not omitted: %q", got)
		}
	}

	// Both remaining facts survive in every case: the reason it was not read, and
	// where it is.
	for _, raw := range []string{"", "---\nobjective: x\n---\n"} {
		got := pointer(raw, path, "status: done")
		if !strings.Contains(got, path) || !strings.Contains(got, "status: done") {
			t.Errorf("pointer lost its path or its reason: %q", got)
		}
		if !strings.Contains(got, "/prosthetic-conscience:resume") {
			t.Errorf("pointer must name the command that prints the note: %q", got)
		}
	}
}
