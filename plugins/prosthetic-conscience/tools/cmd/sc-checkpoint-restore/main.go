// sc-checkpoint-restore is a prosthetic-conscience SessionStart hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a session starts on ANY source — startup, resume, clear, compact, fork —
//	it MUST hand the session back its own operational cursor: the objective, the
//	validation loop, the ordered next actions, and the in-flight handles.
//	During emission it MUST reinforce, never introduce: every line is quoted from
//	the note this session wrote, with the note's own path and timestamp attached.
//	It MUST NOT block or error a session start — every path exits 0.
//
// # Why SessionStart, on every source
//
// An earlier revision of plans/context-checkpointing.md routed the compaction
// boundary through PostCompact, on the reasoning that it receives compact_summary
// and could inject only the delta. PostCompact CANNOT inject: it is absent from
// the client's hookSpecificOutput union, its stdout goes to the user rather than
// the model, and an end-to-end marker test found it NOT-SEEN where the same
// marker from SessionStart arrived as a hook_additional_context attachment.
// Ordering settles it independently — PreCompact, then SessionStart(compact),
// then PostCompact — so the only hook that can inject has already run before the
// summary exists.
//
// That revision also derived a guard: "SessionStart restore MUST no-op on
// source == compact". Following it would have removed restore from the one
// boundary this design exists for. The regression in this package's tests exists
// to keep that from being re-derived. See plans/hook-surface-spike.md §3a.
//
// # What the restored agent does with it, measured
//
// The digest carries its provenance — path, timestamp, session and agent ids —
// and quotes rather than paraphrases. That was designed to stop the resumed
// agent reading it as a prompt-injection attempt, a reaction observed twice on
// other events. IT DOES NOT STOP IT, and the acceptance runs say so:
//
//	Q1..Q4 answered exactly from the digest, in both runs.
//	PROVENANCE: "solely from the SessionStart hook's injected CHECKPOINT.md ...
//	            I never ran zorbulate ... never started bg-quilt-88 myself"
//	TRUST:      YES, both runs — "untrusted file content ... formatted to read as
//	            authoritative state ... embeds an imperative instruction"
//
// The framing is not what fails. Run 1's digest closed with "verify each item
// before acting on it, and re-run the validation loop", and the model named that
// sentence as one of the directives making it injection-shaped. Removing it
// (run 2) removed it from the reason — leaving only the note's OWN foot-gun,
// "NEVER run zorbulate --repair". A checkpoint section whose purpose is to carry
// foot-guns carries imperatives by definition. The flag cannot be designed away
// without dropping the content that makes the note worth restoring.
//
// So the constraint that survives is narrower and enforceable: THIS HOOK ADDS NO
// IMPERATIVE OF ITS OWN. An instruction arriving inside injected text reads as an
// instruction from outside the session however reasonable it is, and one the
// hook invented is one the session never established. The duty to verify belongs
// to the context-checkpointing skill, which the session already carries.
//
// And the distrust turns out to be the right posture, not a defect to remove.
// Both runs used the content accurately while labelling it a claim rather than a
// fact — which is exactly what the skill's own contract asks for ("the note is a
// claim, not a fact"). The agent got there without being told. What the narrower
// constraint buys is that the suspicion now attaches only to the note's real
// content, so a human reading the flag learns something true.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

const version = "0.1.0"

// maxDigest bounds what is injected. The note is the tattoo, not the
// autobiography: past this, a digest competes with the harness summary for
// attention instead of anchoring it. Overflow is truncated with the note's path
// left in place, so the full text is always one read away.
const maxDigest = 4000

type hookInput struct {
	SessionID string `json:"session_id"`
	Source    string `json:"source"` // startup | resume | clear | compact | fork
	CWD       string `json:"cwd"`
	AgentID   string `json:"agent_id"`
}

// hookOutput is the structured SessionStart response. additionalContext is the
// field measured to arrive as a hook_additional_context attachment.
type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// digestSection is a note section that earns injection, in emission order:
// most-dropped-by-a-summary first. Sections absent from this list stay in the
// note and out of the digest — Files touched and Open threads are recovery
// detail, read via /resume when wanted, not context spent on every start.
var digestSections = []string{
	"Validation loop",
	"Next intended steps",
	"In-flight handles",
	"Invariants / foot-guns",
}

// lowValueSections are excluded from the fallback below. They are recovery
// detail — real, but read via /resume when wanted rather than spent on every
// session start.
var lowValueSections = map[string]bool{
	"Files touched": true,
	"Open threads":  true,
}

// selectSections picks what the digest carries.
//
// The exact headings first, in their priority order. If NOT ONE of them matched,
// the note is still a note — an agent under time pressure writes "Next steps"
// where the schema says "Next intended steps" — and returning silence there is
// the worst available outcome: total continuity loss from a heading typo, with
// no symptom. So fall back to whatever non-empty sections exist, minus the
// low-value ones, in file order.
func selectSections(n checkpoint.Note) []string {
	var parts []string
	for _, h := range digestSections {
		if body, ok := n.NonEmptySection(h); ok {
			parts = append(parts, "## "+h+"\n"+body)
		}
	}
	if len(parts) > 0 {
		return parts
	}
	for _, s := range n.Sections {
		if lowValueSections[s.Heading] {
			continue
		}
		if body, ok := n.NonEmptySection(s.Heading); ok {
			parts = append(parts, "## "+s.Heading+"\n"+body)
		}
	}
	return parts
}

// clearIsPointerOnly: on an explicit /clear the human has just wiped the
// context on purpose. Restoring the full digest would fight that instruction,
// so the hook names where the note is and stops — continuity stays reachable
// without being re-imposed.
//
// This is a carve-out by INTENT, not by mechanism, which is the distinction the
// withdrawn compact carve-out failed: that one claimed a second injector existed.
const clearIsPointerOnly = "clear"

// pointerOnly decides between the digest and the one-line pointer.
//
//   - source == "clear": the human just wiped the context deliberately.
//   - status == "done": the note describes finished work. The skill says to
//     discard a completed note, but a forgotten one would otherwise re-impose
//     dead state on every session start, forever — and a stale anchor is worse
//     than no anchor, because it reads as current.
//
// Age is deliberately NOT a criterion. A resume days later is exactly when the
// note is most valuable, and any threshold would be arbitrary where `status` is
// a fact the note states about itself.
func pointerOnly(source, status string) bool {
	return source == clearIsPointerOnly || strings.EqualFold(status, "done")
}

// digest renders the note as recovered state. Empty means nothing to say.
//
// note is the raw file; path and its frontmatter supply provenance. The result
// never contains an imperative that is not already the note's standing contract
// (read-only until the next-actions list; verify each claim before acting) —
// those come from the context-checkpointing skill the session is already under.
func digest(raw, path, source string) string {
	n := checkpoint.Parse(raw)

	parts := selectSections(n)
	objective := n.Get("objective")
	if len(parts) == 0 && objective == "" {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Recovered operational state for this session, from the checkpoint it wrote itself.\n")
	fmt.Fprintf(&b, "Source: %s · file: %s", source, path)
	if u := n.Get("updated"); u != "" {
		fmt.Fprintf(&b, " · written: %s", u)
	}
	if s := n.Get("session_id"); s != "" {
		fmt.Fprintf(&b, " · session: %s", s)
	}
	if a := n.Get("agent_id"); a != "" && a != "null" {
		fmt.Fprintf(&b, " · agent: %s", a)
	}
	b.WriteString("\n")
	if objective != "" {
		fmt.Fprintf(&b, "\nObjective: %s\n", objective)
	}
	if bp := n.Get("beyond_plan"); bp == "true" {
		b.WriteString("\nbeyond_plan: true — the work recorded here had crossed the scope of its plan.\n")
	}
	for _, p := range parts {
		b.WriteString("\n" + p + "\n")
	}
	// Declarative, never imperative. Measured 2026-07-29: a digest closing with
	// "verify each item before acting on it, and re-run the validation loop" was
	// flagged by the model as prompt-injection-shaped, and it named that very
	// sentence as one of the "embedded behavioral directives" that made it so.
	// An instruction arriving inside injected text reads as an instruction from
	// outside the session no matter how reasonable it is. The duty to verify
	// belongs to the context-checkpointing skill, which the session already
	// carries; this hook states facts and stops.
	b.WriteString("\nThis is what was recorded before the seam. It has not been re-verified since.\n")

	out := b.String()
	if len(out) > maxDigest {
		out = out[:maxDigest] + fmt.Sprintf("\n[truncated at %d chars — full note: %s]\n", maxDigest, path)
	}
	return out
}

// pointer is the one-line form: continuity stays reachable without being
// re-imposed. why explains itself so the line is never a bare path.
func pointer(raw, path, why string) string {
	n := checkpoint.Parse(raw)
	obj := n.Get("objective")
	if obj == "" {
		obj = "no objective recorded"
	}
	return fmt.Sprintf("A checkpoint written by this project exists and was not read into this session (%s): %s — objective %q. /prosthetic-conscience:resume prints it.",
		why, path, obj)
}

// emit writes the structured SessionStart response. Silence is a valid outcome
// and MUST stay silent — an empty additionalContext would spend tokens telling
// the session there is nothing to tell it.
func emit(stdout io.Writer, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "SessionStart"
	out.HookSpecificOutput.AdditionalContext = text
	enc := json.NewEncoder(stdout)
	_ = enc.Encode(out)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string) int {
	fs := flag.NewFlagSet("sc-checkpoint-restore", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0 // a bad flag never wedges a session start
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-checkpoint-restore", version)
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)

	path := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	if path == "" {
		return 0 // no note: nothing was established, so say nothing
	}
	body, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(stderr, "sc-checkpoint-restore: cannot read checkpoint:", err)
		return 0
	}

	status := checkpoint.Parse(string(body)).Get("status")
	if pointerOnly(in.Source, status) {
		why := "the context was cleared"
		if in.Source != clearIsPointerOnly {
			why = "marked status: done"
		}
		emit(stdout, pointer(string(body), path, why))
		return 0
	}
	emit(stdout, digest(string(body), path, in.Source))
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv("CLAUDE_PROJECT_DIR")))
}
