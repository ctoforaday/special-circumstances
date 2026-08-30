package setup

import (
	"errors"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func has(s, sub string) bool { return strings.Contains(s, sub) }

// BuildSkeleton creates a stub only for a file something later WRITES. The count has fallen
// twice for the same reason, and the reason is worth keeping: an artifact that moved onto the
// record leaves a stub behind, and the stub outlives its writer silently.
//
//	6 -> 5  blue/frontier.md (#297) — the frontier hypotheses became LINES OF INQUIRY, which carry an id,
//	        a fate and an argument red can rule on, as a markdown file never could
//	5 -> 3  debate.md and red/citation-ledger.md — rendered projections since the record
//	        migration, stubbed for months after the last writer went away
func TestBuildSkeletonCreatesStubsOnlyForFilesSomethingWrites(t *testing.T) {
	dir := t.TempDir()
	res := BuildSkeleton(runOf(t, dir), "test topic")
	if len(res.Created) != 2 {
		t.Fatalf("created = %d, want 3: %v", len(res.Created), res.Created)
	}
	if !has(read(t, filepath.Join(dir, "blue", "report.md")), "test topic") {
		t.Error("stub header missing the topic")
	}
	// Every one of these is rendered from the record on read. A stub here is not a harmless
	// placeholder: it survives to capture as an EMPTY artifact, which reads as "nothing
	// happened" rather than "this lives somewhere else".
	for _, p := range []string{"red/candidates", "red/ledger.md", "red/archive.md", "debate.md",
		"red/citation-ledger.md", "trajectories/board-telemetry.jsonl"} {
		if exists(filepath.Join(dir, filepath.FromSlash(p))) {
			t.Errorf("%s must NOT be created at setup — the tool renders it on read", p)
		}
	}
}

// BuildSkeleton is idempotent — pre-staged files are kept, not overwritten.
func TestBuildSkeletonIdempotent(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "blue", "report.md"), "PRE-STAGED CONTENT\n")
	res := BuildSkeleton(runOf(t, dir), "topic")
	found := false
	for _, s := range res.Skipped {
		if s == "blue/report.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("blue/report.md not reported skipped: %v", res.Skipped)
	}
	if len(res.Created) != 1 {
		t.Errorf("created = %d, want 1", len(res.Created))
	}
	if read(t, filepath.Join(dir, "blue", "report.md")) != "PRE-STAGED CONTENT\n" {
		t.Error("a pre-staged file was overwritten")
	}
}

// BuildPinned: HEAD row + per-cite pins honoring explicit @pin; pre-staged PINNED kept.
func TestBuildPinned(t *testing.T) {
	dir := t.TempDir()
	BuildSkeleton(runOf(t, dir), "topic")
	r := BuildPinned(runOf(t, dir), "abc1234", []string{"research/old-run@def5678", "ideas/backlog.md"})
	if !r.Written {
		t.Fatal("PINNED.md not written")
	}
	txt := read(t, r.Path)
	if !has(txt, "`abc1234`") || !has(txt, "`def5678`") {
		t.Error("explicit pin not honored / HEAD default not applied")
	}
	if !has(txt, "ideas/backlog.md") {
		t.Error("cite path missing")
	}
	if again := BuildPinned(runOf(t, dir), "zzz9999", nil); again.Written {
		t.Error("pre-staged PINNED was overwritten")
	}
}

// MirrorGapPatterns: concatenates memory files; absent/empty memory is a stated no-op.
func TestMirrorGapPatterns(t *testing.T) {
	dir := t.TempDir()
	BuildSkeleton(runOf(t, dir), "topic")
	mem := t.TempDir()
	write(t, filepath.Join(mem, "pattern_a.md"), "# pattern A\n")
	write(t, filepath.Join(mem, "pattern_b.md"), "# pattern B\n")
	r := MirrorGapPatterns([]string{mem}, runOf(t, dir))
	if r.Files != 2 {
		t.Fatalf("files = %d, want 2", r.Files)
	}
	out := read(t, filepath.Join(dir, "inputs", "red-gap-patterns.md"))
	if !has(out, "pattern A") || !has(out, "pattern B") || !has(out, "read-only copy") {
		t.Error("mirror content wrong")
	}
	if none := MirrorGapPatterns([]string{filepath.Join(mem, "nope")}, runOf(t, t.TempDir())); none.Written {
		t.Error("absent memory should be a no-op")
	}
}

// WriteRunLiveMarker: commitment-as-state with the pinned paths for hook guards.
func TestWriteRunLiveMarker(t *testing.T) {
	project := t.TempDir()
	p := runlive.WriteRunLiveMarker(project, "research/x", []string{"research/old-run", "ideas/backlog.md"}, time.Now(), "", "")
	body := read(t, p)
	if !has(body, `"runDir": "research/x"`) {
		t.Errorf("runDir missing: %s", body)
	}
	if !has(body, "research/old-run") || !has(body, "ideas/backlog.md") {
		t.Errorf("pinnedPaths missing: %s", body)
	}
	// An empty pinnedPaths must render as [] not null.
	q := runlive.WriteRunLiveMarker(t.TempDir(), "r", nil, time.Now(), "", "")
	if !has(read(t, q), `"pinnedPaths": []`) {
		t.Error("empty pinnedPaths must be [], not null")
	}
}

// ValidatePins: missing path is named; explicit pin honored; non-git is a stated skip.
func TestValidatePins(t *testing.T) {
	var calls []string
	gitOK := func(args []string) GitResult {
		calls = append(calls, strings.Join(args, " "))
		return GitResult{Status: 0}
	}
	ok := ValidatePins([]string{"plans/x.md@abc1234", "ideas/y.md"}, "headddd", gitOK)
	if len(ok.Missing) != 0 || ok.Checked != 2 {
		t.Fatalf("clean run: missing=%v checked=%d", ok.Missing, ok.Checked)
	}
	if !has(calls[0], "abc1234:plans/x.md") {
		t.Error("explicit pin not used")
	}
	if !has(calls[1], "headddd:ideas/y.md") {
		t.Error("HEAD default not used")
	}
	gitMiss := func(args []string) GitResult {
		if has(strings.Join(args, " "), "plans/gone.md") {
			return GitResult{Status: 128}
		}
		return GitResult{Status: 0}
	}
	bad := ValidatePins([]string{"plans/gone.md@abc1234", "ideas/y.md"}, "headddd", gitMiss)
	if len(bad.Missing) != 1 || bad.Missing[0].Path != "plans/gone.md" {
		t.Fatalf("miss: %v", bad.Missing)
	}
	skip := ValidatePins([]string{"a@b"}, "unknown", nil)
	if !has(skip.Skipped, "UNVALIDATED") {
		t.Error("non-git context must be a stated skip")
	}
}

// MirrorLaw: repo law/ stages read-only into inputs/law; absent law is a stated no-op.
func TestMirrorLaw(t *testing.T) {
	repo := t.TempDir()
	write(t, filepath.Join(repo, "law", "README.md"), "# law\nstatute > precedent > argument\n")
	write(t, filepath.Join(repo, "law", "precedents.md"), "# precedents\n## some-holding [AFFIRMED 2026-07-18]\n")
	run := t.TempDir()
	r := MirrorLaw(filepath.Join(repo, "law"), runOf(t, run))
	if r.Files != 2 {
		t.Fatalf("files = %d, want 2", r.Files)
	}
	staged := read(t, filepath.Join(run, "inputs", "law", "precedents.md"))
	if !has(staged, "read-only copy") || !has(staged, "AFFIRMED") {
		t.Error("law not mirrored with provenance banner")
	}
	if none := MirrorLaw(filepath.Join(repo, "nope"), runOf(t, t.TempDir())); none.Written {
		t.Error("absent law dir should be a no-op")
	}
}

// BuildPatternIndex dedups promoted-first (order is the policy).
func TestBuildPatternIndexDedupsPromotedFirst(t *testing.T) {
	promoted, raw := t.TempDir(), t.TempDir()
	write(t, filepath.Join(promoted, "p.md"), "---\nmetadata:\n  classes: [false-universal]\ndescription: hook\n---\n# P\n")
	write(t, filepath.Join(raw, "p.md"), "---\ndescription: pre-promotion copy\n---\n# P\n")
	r := BuildPatternIndex([]string{promoted, raw})
	if len(r.Unclassified) != 0 {
		t.Errorf("the raw copy resurrected as a backlog item: %v", r.Unclassified)
	}
	if len(r.ByClass["false-universal"]) != 1 {
		t.Errorf("double-delivered: %v", r.ByClass["false-universal"])
	}
	rev := BuildPatternIndex([]string{raw, promoted})
	if len(rev.Unclassified) != 1 || rev.Unclassified[0] != "p.md" {
		t.Errorf("raw-first should surface the unclassified copy: %v", rev.Unclassified)
	}
}

// BuildPatternIndex keeps harness-limit distinct from unclassified.
func TestBuildPatternIndexHarnessLimitDistinct(t *testing.T) {
	d := t.TempDir()
	write(t, filepath.Join(d, "h.md"), "---\nmetadata:\n  classes: []\n  class_note: harness-limit — a tooling constraint\n---\n# H\n")
	write(t, filepath.Join(d, "u.md"), "---\nmetadata:\n  classes: []\n---\n# U\n")
	r := BuildPatternIndex([]string{d})
	if len(r.HarnessLimit) != 1 || r.HarnessLimit[0] != "h.md" {
		t.Errorf("harnessLimit = %v, want [h.md]", r.HarnessLimit)
	}
	if len(r.Unclassified) != 1 || r.Unclassified[0] != "u.md" {
		t.Errorf("unclassified = %v, want [u.md]", r.Unclassified)
	}
}

// PreflightRecordBinary compares EPOCHS: not runnable refuses, a different shape refuses, an
// equal one passes. There is no "how far behind" — nothing promises backwards compatibility,
// so the only answers are same shape or different shape.
func TestPreflightRecordBinary(t *testing.T) {
	missing := PreflightRecordBinary(1, "feov-record", func(string, []string) ExecResult {
		return ExecResult{Err: errors.New("ENOENT")}
	})
	if missing.OK || !has(missing.Reason, "not runnable") || !has(missing.Remedy, "doctor --fix") {
		t.Errorf("missing: %+v", missing)
	}
	skewed := PreflightRecordBinary(2, "feov-record", func(string, []string) ExecResult {
		return ExecResult{Status: 0, Stdout: "1\n"}
	})
	if skewed.OK || !has(skewed.Reason, "schema 1") || !has(skewed.Reason, "reads 2") {
		t.Errorf("skewed must name both shapes: %+v", skewed)
	}
	good := PreflightRecordBinary(1, "feov-record", func(string, []string) ExecResult {
		return ExecResult{Status: 0, Stdout: "1\n"}
	})
	if !good.OK || good.Version != "1" {
		t.Errorf("good: %+v", good)
	}
	// A BINARY THAT ANSWERS SOMETHING OTHER THAN A NUMBER IS REFUSED, not parsed hopefully.
	// This is what an older binary does — it has no --schema, so cobra prints usage — and
	// treating that output as an epoch is how a string-shaped check waves through the exact
	// binary it exists to catch.
	prose := PreflightRecordBinary(1, "feov-record", func(string, []string) ExecResult {
		return ExecResult{Status: 0, Stdout: "feov-record version 0.72.0\n"}
	})
	if prose.OK || !has(prose.Reason, "not an epoch") {
		t.Errorf("prose must be refused: %+v", prose)
	}
	absent := PreflightRecordBinary(1, "no-such-binary", func(string, []string) ExecResult {
		return ExecResult{Err: errors.New("exec: not found")}
	})
	if absent.OK || !has(absent.Reason, "not runnable") {
		t.Errorf("absent: %+v", absent)
	}
}

// MirrorScorecards: empty corpus explains itself instead of an undefined reason.
func TestMirrorScorecardsEmptyCorpus(t *testing.T) {
	mem, runDir := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755)
	r := MirrorScorecards(mem, runOf(t, runDir))
	if r.Written {
		t.Error("empty corpus should not be written")
	}
	if r.Reason == "" || !has(r.Reason, "capture") {
		t.Errorf("reason should explain where scorecards come from: %q", r.Reason)
	}
}

// MirrorScorecards: the emitted HEADLINE line is the ranked authority (detector leads).
func TestMirrorScorecardsHeadlineRanked(t *testing.T) {
	mem, runDir := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755)
	// A card whose emitted HEADLINE puts the tripped detector first.
	write(t, filepath.Join(mem, "blue-scorecard.md"),
		"# blue scorecard\n\n## run-6\n\n- `x` [benchmark] — A: **1**\n\nHEADLINE: tripped 7 [DETECTOR] · bench_one 1 [BENCHMARK] · bench_two 2 [BENCHMARK]\n")
	r := MirrorScorecards(mem, runOf(t, runDir))
	if len(r.Headlines["blue"]) != 3 || !strings.HasPrefix(r.Headlines["blue"][0], "tripped 7") {
		t.Errorf("headline not ranked from the emitted line: %v", r.Headlines["blue"])
	}
}

// MirrorScorecards: a pre-HEADLINE card still yields a prompt headline via the fallback,
// and a colon in the clause does not hide the metric.
func TestMirrorScorecardsFallbackParsesColonClause(t *testing.T) {
	mem, runDir := t.TempDir(), t.TempDir()
	os.MkdirAll(filepath.Join(runDir, "inputs"), 0o755)
	write(t, filepath.Join(mem, "red-scorecard.md"), strings.Join([]string{
		"# red scorecard", "", "## run-5", "",
		"- `unrecorded_claim_loss` [detector] — LOSS: additive violations: **4** (6 lost, 2 retired)",
		"- `anchored_closures_pct` [benchmark] — Attestation-format invariant: **89**",
		"",
	}, "\n"))
	r := MirrorScorecards(mem, runOf(t, runDir))
	joined := strings.Join(r.Headlines["red"], " | ")
	if !has(joined, "unrecorded_claim_loss 4") {
		t.Errorf("colon-bearing clause not read by the fallback: %v", r.Headlines["red"])
	}
	if !has(joined, "anchored_closures_pct 89") {
		t.Errorf("second row missing: %v", r.Headlines["red"])
	}
}

// #270: A RUN THAT WAS NEVER CLOSED MUST NOT BE OVERWRITTEN IN SILENCE.
//
// The marker's only remover is `capture`, and capture is optional. Setup used to overwrite it
// unconditionally — self-healing and silent, so nobody learned the previous run was never closed.
// Between the abandoned run and this setup, that stale marker is what seat.InferRunDir hands to
// every verb invoked without --run.
func TestSetupRefusesWhenAnotherRunIsStillOpen(t *testing.T) {
	cwd := t.TempDir()
	runlive.WriteRunLiveMarker(cwd, "research/2026-08-01_abandoned", nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "")

	m, ok := runlive.ReadRunLiveMarker(cwd)
	if !ok || m.RunDir != "research/2026-08-01_abandoned" {
		t.Fatalf("fixture: the marker must read back, got %+v ok=%v", m, ok)
	}
	if runlive.SameRun(cwd, m.RunDir, "research/2026-08-16_new") {
		t.Fatal("two different runs must not compare equal")
	}
}

// IDEMPOTENT FOR THE SAME RUN. Re-running setup on a run in progress is ordinary — and the marker
// stores whatever form it was given, so a string compare would refuse the very run it belongs to.
func TestTheSameRunIsRecognisedThroughPathForm(t *testing.T) {
	cwd := t.TempDir()
	abs := filepath.Join(cwd, "research", "2026-08-16_live")
	for _, form := range []string{"research/2026-08-16_live", abs, "./research/2026-08-16_live", filepath.FromSlash("research/2026-08-16_live")} {
		if !runlive.SameRun(cwd, form, "research/2026-08-16_live") {
			t.Errorf("%q must be recognised as the same run — setup on a live run must stay idempotent", form)
		}
	}
}

// A marker it cannot READ is not evidence of an open run. Blocking a new run on a corrupt file
// would trade a silent hazard for a stuck operator, and the gate can only name a run it can read.
func TestAnUnreadableMarkerIsNotAnOpenRun(t *testing.T) {
	for _, body := range []string{"", "{", "{}", `{"runDir":""}`, `{"runDir":"   "}`} {
		cwd := t.TempDir()
		if err := os.MkdirAll(filepath.Join(cwd, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cwd, ".claude", "run-live.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := runlive.ReadRunLiveMarker(cwd); ok {
			t.Errorf("%q must not read as an open run", body)
		}
	}
	if _, ok := runlive.ReadRunLiveMarker(t.TempDir()); ok {
		t.Error("an absent marker is not an open run")
	}
}
