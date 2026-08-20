package seatprobe

// THE PERMISSION GRANT IS THE AGENT'S OWN DECLARATION, READ AT DISPATCH.
//
// Two different things wear the name "tools" and conflating them cost this probe a run in each
// direction.
//
//	AVAILABILITY  which tools the seat HAS. The agent definition's `tools:` line decides it, and
//	              nothing on the command line needs to — passing --allowedTools alongside --agent
//	              does not narrow it (measured: an agent dispatched with `--allowedTools Bash` still
//	              listed WebSearch and ToolSearch as its own).
//	PERMISSION    whether a call is allowed to RUN. Headless, with nobody at the keyboard, an
//	              ungranted Bash call comes back "This command requires approval".
//
// The old dispatch passed a hand-written list with --system-prompt-file, where it was doing BOTH
// jobs — and the list omitted WebSearch and ToolSearch, so `corroborate` was unperformable and every
// run scored the seat UNMET on it. Removing the list fixed availability and silently removed the
// grant: the seat's first two calls, both `--seat-id … --help`, came back needing approval, it
// concluded the record tool was not responding, and it spent the sitting reasoning about a blocker
// the harness had created. `--permission-mode dontAsk` and `auto` do not help — measured, both still
// refuse the record binary.
//
// So the grant is DERIVED from the definition rather than written beside it. A tool added to an
// agent is granted on the next run with nothing to keep in step.

import (
	"fmt"
	"os"
	"strings"
)

// GrantedTools reads the `tools:` frontmatter line from an agent definition.
//
// THE MISS IS AN ERROR. An empty grant is not a restrictive run — it is the approval prompt again,
// and the seat that meets it reports a broken tool rather than a missing permission. Those look
// nothing alike in the transcript and identical in the score.
func GrantedTools(constitutionPath string) ([]string, error) {
	b, err := os.ReadFile(constitutionPath)
	if err != nil {
		return nil, fmt.Errorf("read agent definition: %w", err)
	}
	src := string(b)
	if !strings.HasPrefix(src, "---") {
		return nil, fmt.Errorf("%s has no frontmatter, so it declares no tools to grant", constitutionPath)
	}
	end := strings.Index(src[3:], "\n---")
	if end < 0 {
		return nil, fmt.Errorf("%s has an unterminated frontmatter block", constitutionPath)
	}
	for _, line := range strings.Split(src[3:end+3], "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "tools:")
		if !ok {
			continue
		}
		var out []string
		for _, t := range strings.Split(rest, ",") {
			if t = strings.TrimSpace(t); t != "" {
				out = append(out, t)
			}
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("%s declares an EMPTY tools list — the seat would be dispatched with no permission to run anything and would report the record tool as broken", constitutionPath)
		}
		return out, nil
	}
	return nil, fmt.Errorf("%s declares no `tools:` line — nothing to grant, and a seat dispatched without a grant meets an approval prompt with nobody there to answer it", constitutionPath)
}
