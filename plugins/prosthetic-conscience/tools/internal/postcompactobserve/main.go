// sc-postcompact-observe is a prosthetic-conscience PostCompact hook (Go binary).
//
// Contract (Design by Contract):
//
//	AFTER a compaction it MUST record how much of the checkpoint's content the
//	harness summary carried through, as one append-only observation row.
//	It MUST NOT attempt to restore anything. PostCompact cannot inject into the
//	model — absent from the client's hookSpecificOutput union, its stdout goes to
//	the user, and an end-to-end marker test found it NOT-SEEN (plans/
//	hook-surface-spike.md §3a). It also runs AFTER SessionStart(compact), so the
//	digest has already been handed over by the time this fires.
//	It MUST stay silent on stdout: that stream reaches the human, and a line per
//	compaction is noise they did not ask for.
//	It MUST NOT block — every path exits 0.
//
// # What the row is for
//
// Risk R4 (duplication between the digest and the harness summary) cannot be
// prevented at a boundary: the digest is injected before the summary exists. It
// CAN be measured across boundaries. A section the seal explicitly asked the
// summarizer to preserve, which the summary drops every time, is a signal about
// the seal's wording — and the only way to see it is to keep the score.
//
// # What the row is NOT
//
// Token overlap is a heuristic, not a reading. It answers "did the summary
// mention the same distinctive words", never "did the summary preserve the
// meaning". Rows are labelled with the probe that produced them so a later
// reader cannot mistake this for evidence of what a summary said. Under the
// suite's rule — exploration may summarize, adjudication must cite — these rows
// are exploration: they say where to look, and they can never back a finding.
//
// One confound is known and must not be read past. The restore digest is in
// context when the NEXT summary is produced, so a high ratio is consistent with
// two different stories: the summarizer preserved the note's content, or it
// echoed the digest that was already in front of it. The probe cannot tell them
// apart. First live run: two boundaries at 1.00 across every section (13.5KB and
// 15.4KB summaries), one at 0.10/0.00/0.12/0.14 (734B). The low row is the
// informative one — a short summary drops nearly everything, which is exactly
// when restore earns its place.
package postcompactobserve

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/freshness"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookenv"
)

// probe names the measurement, so a row can never be read as more than it is.
const probe = "token-overlap-v1"

// minToken is the shortest word counted as distinctive. Below this the overlap
// is dominated by grammar and every section scores high against every summary.
const minToken = 5

type hookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Trigger        string `json:"trigger"` // manual | auto
	CompactSummary string `json:"compact_summary"`
	AgentID        string `json:"agent_id"`
	CWD            string `json:"cwd"`
}

// observation is one compaction boundary, one row.
type observation struct {
	At        string           `json:"at"`
	Probe     string           `json:"probe"`
	SessionID string           `json:"session_id"`
	AgentID   string           `json:"agent_id,omitempty"`
	Trigger   string           `json:"trigger"`
	SummaryB  int              `json:"summary_bytes"`
	NotePath  string           `json:"note_path"`
	Sections  []sectionOverlap `json:"sections"`

	// The note's age at this boundary, in the same three units the seal row uses and
	// computed by the same gauge — a second implementation is how two records start
	// disagreeing about what "age" meant. Each is OMITTED when unmeasured rather than
	// written as a zero.
	NoteAgeTurns      *int `json:"note_age_turns,omitempty"`
	TurnsMeasured     bool `json:"turns_measured"`
	NoteGrowthTokens  *int `json:"note_growth_tokens,omitempty"`
	GrowthMeasured    bool `json:"growth_measured"`
	NoteBranchCommits *int `json:"note_branch_commits,omitempty"`
	BranchMeasured    bool `json:"branch_measured"`
	CeilingKnown      bool `json:"ceiling_known"`

	// NudgeEnabled says whether the nudge was live when this row was written.
	// Criterion 6 compares the two populations, and a row that does not say which it
	// belongs to cannot be put in either. Recorded per row rather than inferred from
	// dates, because "when did it ship" is exactly the kind of fact that gets
	// reconstructed wrongly a month later.
	NudgeEnabled bool `json:"nudge_enabled"`
}

// sectionOverlap is how much of one section's distinctive vocabulary the summary
// reused. Counts travel with the ratio: 1/1 and 40/40 are not the same evidence.
type sectionOverlap struct {
	Heading string  `json:"heading"`
	Tokens  int     `json:"tokens"`
	Present int     `json:"present"`
	Ratio   float64 `json:"ratio"`
}

// tokens extracts the distinctive vocabulary of a body: lowercased words of at
// least minToken runes, deduplicated, sorted for a stable row.
func tokens(body string) []string {
	seen := map[string]bool{}
	for _, f := range strings.FieldsFunc(strings.ToLower(body), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len([]rune(f)) >= minToken {
			seen[f] = true
		}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// overlap scores every non-empty section of the note against the summary.
// A section with no distinctive tokens is reported with tokens: 0 and ratio 0
// rather than dropped — "nothing to measure" and "measured as absent" are
// different facts, and collapsing them is how a probe starts lying.
func overlap(note, summary string) []sectionOverlap {
	hay := strings.ToLower(summary)
	var out []sectionOverlap
	for _, s := range checkpoint.Parse(note).Sections {
		body, ok := checkpoint.Note{Sections: []checkpoint.Section{s}}.NonEmptySection(s.Heading)
		if !ok {
			continue
		}
		ts := tokens(body)
		present := 0
		for _, t := range ts {
			if strings.Contains(hay, t) {
				present++
			}
		}
		var ratio float64
		if len(ts) > 0 {
			ratio = float64(present) / float64(len(ts))
		}
		out = append(out, sectionOverlap{Heading: s.Heading, Tokens: len(ts), Present: present, Ratio: ratio})
	}
	return out
}

// appendRow writes one JSONL row. Best-effort: an observation that cannot be
// written must never cost the session anything.
func appendRow(projectDir string, o observation, stderr io.Writer) {
	dir := filepath.Join(projectDir, ".claude", "checkpoints")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(stderr, "sc-postcompact-observe: cannot create dir:", err)
		return
	}
	line, err := json.Marshal(o)
	if err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "compaction-observations.jsonl"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintln(stderr, "sc-postcompact-observe: cannot open observations:", err)
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, projectDir string, now time.Time) int {
	fs := flag.NewFlagSet("sc-postcompact-observe", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-postcompact-observe"))
		return 0
	}

	raw, _ := io.ReadAll(stdin)
	var in hookInput
	_ = json.Unmarshal(raw, &in)
	projectDir = hookenv.ProjectDir(projectDir, in.CWD)
	if !hookenv.Explain(projectDir, stderr, "sc-postcompact-observe") {
		return 0
	}

	path := checkpoint.NotePath(projectDir, func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}, filepath.Glob)
	if path == "" {
		return 0 // no note: nothing to score the summary against
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	age := freshness.Of(projectDir, in.TranscriptPath, string(body),
		freshness.BranchWork(checkpoint.Parse(string(body)).Get("head")))

	o := observation{
		At:        now.UTC().Format(time.RFC3339),
		Probe:     probe,
		SessionID: in.SessionID,
		AgentID:   in.AgentID,
		Trigger:   in.Trigger,
		SummaryB:  len(in.CompactSummary),
		NotePath:  path,
		Sections:  overlap(string(body), in.CompactSummary),

		TurnsMeasured:  age.TurnsMeasured,
		GrowthMeasured: age.GrowthKnown,
		BranchMeasured: age.BranchKnown,
		CeilingKnown:   age.ProximityKnown,
		// Phase 1 ships no nudge. This becomes a real reading when stopnudge exists.
		NudgeEnabled: false,
	}
	if age.TurnsMeasured {
		t := age.Turns
		o.NoteAgeTurns = &t
	}
	if age.GrowthKnown {
		g := age.Growth
		o.NoteGrowthTokens = &g
	}
	if age.BranchKnown {
		c := age.BranchCommits
		o.NoteBranchCommits = &c
	}
	appendRow(projectDir, o, stderr)
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr,
		os.Getenv("CLAUDE_PROJECT_DIR"), time.Now())
}
