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
package checkpointrestore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookmain"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookunit"
)

// maxDigest bounds what is injected. The note is the tattoo, not the
// autobiography: past this, a digest competes with the harness summary for
// attention instead of anchoring it. Overflow is truncated with the note's path
// left in place, so the full text is always one read away.
const maxDigest = 4000

// truncationReserve keeps room for what digest() appends AFTER the sections — the
// shortened-to-fit line, re-arm notes, and the closing statement. Without it the sections
// would spend the whole budget and the clamp would cut the very lines explaining that
// something had been cut.
const truncationReserve = 400

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
		HookEventName     string   `json:"hookEventName"`
		AdditionalContext string   `json:"additionalContext,omitempty"`
		WatchPaths        []string `json:"watchPaths,omitempty"`
	} `json:"hookSpecificOutput"`
}

// digestSection is a note section that earns injection, in emission order:
// most-dropped-by-a-summary first. Sections absent from this list stay in the
// note and out of the digest — Files touched and Open threads are recovery
// detail, read via /resume when wanted, not context spent on every start.
// digestSections is the emission order, and it is also the BUDGET PRIORITY order when the
// digest does not fit (see fitSections).
//
// Foot-guns sit second, not last. They used to be last, which put the section a resumed
// session most needs BEFORE acting first in line to be cut — measured on this repo's own
// note: a 4052-char digest against a 4000 cap dropped the foot-guns entirely, including the
// line that stops a session amending published commits (#218). The schema marks the
// validation loop load-bearing, so it stays first; everything after it is ordered by what
// costs most to have missing.
var digestSections = []string{
	"Validation loop",
	"Invariants / foot-guns",
	"Next intended steps",
	"In-flight handles",
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
func digest(raw, path, source string, rearm checkpoint.RearmState) string {
	n := checkpoint.Parse(raw)

	parts := selectSections(n)
	objective := n.Get("objective")
	if len(parts) == 0 && objective == "" {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Recovered operational state for this session, from the checkpoint it wrote itself.\n")
	fmt.Fprintf(&b, "Source: %s · file: %s", source, path)
	// SCHEMA 3 SPLIT `updated:` IN TWO, because "something happened" was not a useful
	// fact. written_at moves when the BODY changes; reaffirmed_at moves when a session
	// judged the note still accurate without changing it. Rendering only the first
	// would make a re-affirmed note indistinguishable from an untouched one; rendering
	// them as one field is what schema 2 did.
	if w := n.Get("written_at"); w != "" {
		fmt.Fprintf(&b, " · written: %s", w)
	}
	if r := n.Get("reaffirmed_at"); r != "" && r != "null" {
		fmt.Fprintf(&b, " · reaffirmed: %s", r)
	}
	if age := staleness(n.Get("head"), commitsSince); age != "" {
		fmt.Fprintf(&b, " · %s", age)
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
	// Sections are fitted to what is left of the budget AFTER the provenance header, so a
	// long header cannot silently eat the sections. clamp() below remains as a final
	// backstop for everything appended after this point.
	fitted, shortened := fitSections(parts, maxDigest-b.Len()-truncationReserve, path)
	for _, p := range fitted {
		b.WriteString("\n" + p + "\n")
	}
	if len(shortened) > 0 {
		// Name what was reduced. A digest that drops what a session needed must not read
		// like a digest that had nothing to drop (#218).
		fmt.Fprintf(&b, "\nShortened to fit: %s — read %s for the full text.\n",
			strings.Join(shortened, ", "), path)
	}
	if rearmed := rearmLines(rearm); rearmed != "" {
		b.WriteString("\n" + rearmed)
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

	return b.String()
}

// staleness turns the note's recorded head into a fact a reader can act on.
//
// WHY THE FIELD EXISTS (#166 option 3). `objective:` is free prose, so "next is the FEOV
// verdict-case bypass" and "the FEOV verdict-case bypass is done" are indistinguishable to
// any checker — which is how a note describing work that had already shipped kept
// presenting itself as current for 37 minutes. The head is the one field that makes the
// note's age FALSIFIABLE without new plumbing or harness cooperation.
//
// Silence when the note records no head (an older note, or one written outside a repo),
// when the head is the current one, or when git cannot answer. A restore that guessed here
// would be inventing provenance, which is the one thing this hook must never do.
func staleness(head string, since func(string) (int, bool)) string {
	if strings.TrimSpace(head) == "" || head == "null" {
		return ""
	}
	n, ok := since(head)
	if !ok {
		// The commit is gone — rebased, amended, or from another branch. That is a
		// STRONGER staleness signal than a count, not a reason to stay quiet.
		return "written against " + head + ", which is no longer reachable from HEAD"
	}
	switch n {
	case 0:
		return "" // current; saying so would spend a line on "nothing has changed"
	case 1:
		return "written 1 commit ago on this branch (" + head + ")"
	}
	return fmt.Sprintf("written %d commits ago on this branch (%s)", n, head)
}

// THE DIGEST CANNOT CARRY THE FULL GAUGE, and this is a payload fact rather than a
// choice. SessionStart's stdin is {session_id, cwd, hook_event_name, source} — MEASURED,
// hook-surface-spike.md §9e correction 3 — with no transcript_path. Turns and growth both need the
// transcript, so at this event they cannot be computed at all.
//
// What the digest can say about age, it already says: staleness() renders the branch
// measure, which needs only git. The remaining two ride the seal record instead, where
// the sealing events do carry a transcript path.
//
// Deriving the transcript path from session_id and cwd was considered and refused: that
// recovers a path by assembling a string, which is the shape this suite is named after,
// and it would break silently the moment the projects-directory convention moved.

// commitsSince counts commits on THIS BRANCH'S OWN LINE from a recorded head to HEAD, reporting
// whether the commit is reachable at all. Best-effort: no git, no repo, or an unknown ref all mean
// "unreachable", and the restore hook must never fail over provenance.
//
// --first-parent IS THE WHOLE POINT, and it was missing. A plain count answers "how many commits
// are reachable from HEAD and not from the note", which on any branch that has merged is dominated
// by OTHER PEOPLE'S WORK ARRIVING. Measured on this repository: 109 reachable against 24 on the
// branch's own line — 85 of them merged-in side branches. A note written before a routine merge
// was reported 100+ commits stale having done nothing, which teaches a reader to ignore the number.
func commitsSince(head string) (int, bool) {
	out, err := exec.Command("git", "rev-list", "--count", "--first-parent", head+"..HEAD").Output()
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, false
	}
	return n, true
}

// clamp bounds the FINAL composed text. It lives at the end of composition
// rather than inside digest() because everything appended after a truncation
// would otherwise escape the cap — which is exactly what happened when the
// unresolved-surfaces line was added downstream of it.
func clamp(text, path string) string {
	if len(text) <= maxDigest {
		return text
	}
	return cut(text, maxDigest) + fmt.Sprintf("\n[truncated at %d chars — full note: %s]\n", maxDigest, path)
}

// cut truncates to at most n BYTES without splitting a rune.
//
// FIX (defect): this used to be text[:n], a byte slice. Checkpoint notes are routinely
// non-ASCII — em-dashes, arrows, check marks, the "←" the schema itself uses — so a cut
// landing mid-rune produced invalid UTF-8 in the one artifact a resumed session reads as
// authoritative. FEOV's record renderer carries the same lesson in its own truncate: byte
// slicing "silently cuts a different amount of text and can split a rune".
func cut(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// fitSections makes every selected section VISIBLE within the budget, truncating inside
// sections rather than letting the tail fall off the end.
//
// WHY (#218). The digest used to be joined and then cut at maxDigest, so the last sections
// vanished — heading and all — and the reader was told only that something had been cut, not
// what. Measured on this repo's own note: the foot-guns section disappeared silently, twice,
// and shortening an earlier section by ~180 chars moved the digest length by 2, because the
// cut point is fixed. A digest that drops the section a session needed reads exactly like a
// digest that had nothing to drop, which is the silent-absence class this suite keeps
// finding in everything except itself.
//
// Every section is now represented at minimum by its heading and a pointer. Budget beyond
// that goes in digestSections order, so the load-bearing sections keep their content and the
// ones that lose it still announce that they exist.
func fitSections(parts []string, budget int, path string) (out []string, shortened []string) {
	if len(parts) == 0 {
		return nil, nil
	}
	total := 0
	for _, p := range parts {
		total += len(p) + 2 // the blank line between sections
	}
	if total <= budget {
		return parts, nil
	}

	// Reserve the heading and a pointer for every section first: no section may vanish.
	type sec struct{ head, body, stub string }
	secs := make([]sec, 0, len(parts))
	reserved := 0
	for _, p := range parts {
		head, body, _ := strings.Cut(p, "\n")
		stub := head + "\n[omitted here — read " + path + "]"
		secs = append(secs, sec{head, body, stub})
		reserved += len(stub) + 2
	}
	spare := budget - reserved

	for i := range secs {
		full := len(secs[i].head) + 1 + len(secs[i].body)
		if spare >= full-len(secs[i].stub) {
			out = append(out, secs[i].head+"\n"+secs[i].body)
			spare -= full - len(secs[i].stub)
			continue
		}
		if spare > 0 {
			kept := cut(secs[i].body, spare)
			if kept != "" {
				out = append(out, secs[i].head+"\n"+kept+
					fmt.Sprintf("\n[section truncated, %d chars omitted — read %s]", len(secs[i].body)-len(kept), path))
				shortened = append(shortened, strings.TrimPrefix(secs[i].head, "## "))
				spare = 0
				continue
			}
		}
		out = append(out, secs[i].stub)
		shortened = append(shortened, strings.TrimPrefix(secs[i].head, "## "))
	}
	return out, shortened
}

// rearmLines states which checks have had their trigger surface move since the
// note recorded a result. Declarative, like everything else the hook emits: it
// says what changed and what recorded it, never what to do about it.
//
// Empty when nothing is re-armed, so a clean session start says nothing.
func rearmLines(s checkpoint.RearmState) string {
	rs := s.Sorted()
	if len(rs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Validation checks whose trigger surface changed after the recorded result — the `last run` above is stale for these:\n")
	for _, r := range rs {
		fmt.Fprintf(&b, "- check %d (%s) — %s %s at %s\n", r.Index, r.Check, r.Event, r.By, r.At)
	}
	return b.String()
}

// watchTargets resolves every check's trigger surface into paths the watcher
// will accept, and returns the surfaces it could not resolve.
//
// Unresolvable is DATA, not an error: a surface expressed as prose ("a new
// seal") has no path, and a session that silently watched nothing would look
// identical to one that watched everything. The unresolved list goes into the
// digest so the gap is visible to the agent that owns it.
func watchTargets(n checkpoint.Note, projectDir string) (paths []string, unresolved []string) {
	loop, ok := n.NonEmptySection("Validation loop")
	if !ok {
		return nil, nil
	}
	// Containment via os.Root, not path arithmetic: the note is agent-written,
	// and a symlink inside the project would defeat any lexical check.
	within, closeRoot := checkpoint.RootedWithin(projectDir)
	defer func() { _ = closeRoot() }()
	seen := map[string]bool{}
	for _, c := range checkpoint.ParseValidationLoop(loop) {
		got, why := checkpoint.WatchTargets(c.ReArmedBy, within)
		for _, p := range got {
			if !seen[p] {
				seen[p] = true
				paths = append(paths, p)
			}
		}
		if len(got) == 0 && why != "" {
			unresolved = append(unresolved, fmt.Sprintf("check %d: %s", c.Index, why))
		}
	}
	// A malformed label costs the same thing an unresolvable surface costs — a
	// check that is quietly not watched — so it is reported through the same
	// channel rather than a new one. See checkpoint.LoopProblems (#192, #193).
	unresolved = append(unresolved, checkpoint.LoopProblems(loop)...)
	return paths, unresolved
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
//
// watchPaths is independent of the digest: a note may be worth watching for even
// when there is nothing to say, so this emits when EITHER is non-empty. Both
// fields are measured to work in one response (hook-surface-spike.md §6).
func emit(stdout io.Writer, text string, watch []string) {
	if strings.TrimSpace(text) == "" && len(watch) == 0 {
		return
	}
	var out hookOutput
	out.HookSpecificOutput.HookEventName = "SessionStart"
	out.HookSpecificOutput.AdditionalContext = text
	out.HookSpecificOutput.WatchPaths = watch
	enc := json.NewEncoder(stdout)
	_ = enc.Encode(out)
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string) int {
	if hookmain.Preamble(args, stdout, stderr, hookmain.Named("sc-checkpoint-restore")) {
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-checkpoint-restore") {
		return 0
	}

	text, watch := compose(projectDir, in.Source)
	emit(stdout, text, watch)
	return 0
}

// compose is the restore's whole product — the digest text and the paths to watch — with no
// I/O of its own. The standalone binary and the merged SessionStart binary both run THIS, so
// the two cannot drift into restoring differently.
func compose(projectDir, source string) (string, []string) {
	path := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	if path == "" {
		return "", nil // no note: nothing was established, so say nothing
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}

	note := checkpoint.Parse(string(body))
	if pointerOnly(source, note.Get("status")) {
		why := "the context was cleared"
		if source != clearIsPointerOnly {
			why = "marked status: done"
		}
		// No watch on a pointer: the session is not resuming this work, so
		// registering its surfaces would collect re-arms nobody asked for.
		return pointer(string(body), path, why), nil
	}

	watch, unresolved := watchTargets(note, projectDir)
	// An unreadable state file must not block a restore — the digest is the
	// load-bearing half and re-arm history is an annotation on it. But it is
	// reported rather than swallowed: silence here is what let a torn file look
	// like "coverage stopped" for three sessions (#165).
	rearm, rearmErr := checkpoint.LoadRearm(os.ReadFile, checkpoint.RearmPath(projectDir))
	if rearmErr != nil {
		unresolved = append(unresolved, rearmErr.Error())
	}

	text := digest(string(body), path, source, rearm)
	if text != "" && len(unresolved) > 0 {
		text += "\nTrigger surfaces that could not be resolved to a watched path, so a change there is not recorded: " +
			strings.Join(unresolved, "; ") + ".\n"
	}
	return clamp(text, path), watch
}

// Unit exposes the restore to the merged SessionStart binary.
//
// It RETURNS its text and watch paths rather than emitting them, so the event's binary can
// compose ONE response: two hooks each emitting their own JSON document would have the
// second overwrite or corrupt the first, which is a thing separate processes got away with
// only because the harness merged their outputs for us.
func Unit() hookunit.Unit {
	return hookunit.Unit{
		Name: "sc-checkpoint-restore",
		Run: func(c *hookunit.Ctx) hookunit.Result {
			var src struct {
				Source string `json:"source"`
			}
			_ = json.Unmarshal(c.Raw, &src)
			text, watch := compose(c.ProjectDir, src.Source)
			return hookunit.Result{Name: "sc-checkpoint-restore", Stdout: text, Watch: watch}
		},
	}
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv("CLAUDE_PROJECT_DIR"))
}
