// Package setup ports setup-research-run.mjs — the mechanical half of /research
// steps 1-3 — into the feov-record binary. It creates the run's blackboard
// skeleton, pins the evidence base, mirrors red's gap-pattern + law + scorecard
// memory into inputs/, writes the .run-live marker, and preflights the record
// binary. Behaviour is preserved byte-for-byte against the mjs it replaces
// (the sole non-deterministic field is the marker's `started` timestamp, and the
// marker lives outside the run dir).
//
// Where the mjs used readdirSync (filesystem order, non-guaranteed), Go sorts the
// entries: a deterministic improvement over the JS's implicit order.
package setup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// dirs and stubs mirror the mjs constants exactly. friction.md is deliberately
// absent (friction lives on the record); ledger/archive/telemetry are red-merge-born.
//
// `records` is NOT in this list — it is created at its RESOLVED location by BuildSkeleton,
// because a separated run must not be handed an empty records/ in the run directory. An empty
// one is worse than none: it shows up in the seat's listing as the place to look, and it makes
// the run appear unseparated to anyone reading the tree.
var dirs = []string{"blue/candidates", "red", "trajectories", "inputs"}

// A STUB IS A PROMISE THAT SOMETHING WILL FILL IT. Two here were promises nobody kept:
// `debate.md` and `red/citation-ledger.md` became rendered projections (`show --view debate`,
// `--view citation-ledger`) and no writer remained, so setup laid down a one-line husk that
// survived to capture in every run since. Untouched-but-present is the worst state a file can
// be in — it reads as an empty artifact rather than an absent one, which is how `red/ledger.md`
// and `red/archive.md` took a capture audit down with them for months (see internal/capture,
// AUDIT 2). What is listed below is written by something after setup:
//
//	report.md          `bench assemble`
//	blue/report.md     the round-0 synthesizer, then every `blue edit`
var stubs = [][2]string{
	{"report.md", "report.md"},
	{"blue/report.md", "blue report"},
}

// marshalJSON matches JS `JSON.stringify(x, null, 2) + '\n'`: two-space indent, a
// trailing newline, and — critically — NO HTML escaping. Go's default escapes
// <, >, & to \u00xx; JS does not, and the law banner ("statute > precedent") and
// gap-pattern hooks carry '>'. json.Encoder.Encode already appends the newline.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

// ---- skeleton + pins ----

// SkeletonResult reports which stub files were created versus kept (pre-staged).
type SkeletonResult struct {
	Created []string
	Skipped []string
}

func BuildSkeleton(runDir, topic string) SkeletonResult {
	res := SkeletonResult{Created: []string{}, Skipped: []string{}}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(runDir, d), 0o755)
	}
	// Resolution failure is not fatal to a skeleton: setup's job is to lay out a directory, and
	// the first record verb reports an unreachable root far better than a half-built run does.
	if recDir, err := record.RecordsDir(runDir); err == nil {
		os.MkdirAll(recDir, 0o755)
	}
	for _, s := range stubs {
		rel, name := s[0], s[1]
		p := filepath.Join(runDir, filepath.FromSlash(rel))
		if exists(p) {
			res.Skipped = append(res.Skipped, rel)
			continue
		}
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(fmt.Sprintf("# %s — %s\n", name, topic)), 0o644)
		res.Created = append(res.Created, rel)
	}
	return res
}

// PinResult reports whether PINNED.md was written (false = pre-staged, kept).
type PinResult struct {
	Written bool
	Path    string
}

func BuildPinned(runDir, head string, cites []string) PinResult {
	p := filepath.Join(runDir, "inputs", "PINNED.md")
	if exists(p) {
		return PinResult{Written: false, Path: p}
	}
	lines := []string{
		"# PINNED — the evidence base, frozen at launch",
		"",
		"Agents MUST cite these commits for repo and cross-corpus references; the evidence base",
		"does not move under the report. Pushes to cited paths are frozen while the run is live",
		"(.run-live marker; hook-guarded).",
		"",
		"| Corpus | Pin |",
		"|---|---|",
		fmt.Sprintf("| Repo HEAD at launch | `%s` |", head),
	}
	for _, c := range cites {
		path, pin := splitPin(c)
		if pin == "" {
			pin = head
		}
		lines = append(lines, fmt.Sprintf("| `%s` | `%s` |", path, pin))
	}
	lines = append(lines, "")
	os.WriteFile(p, []byte(strings.Join(lines, "\n")), 0o644)
	return PinResult{Written: true, Path: p}
}

// splitPin splits "path@pin" into (path, pin); "path" alone yields pin "".
func splitPin(c string) (string, string) {
	if i := strings.Index(c, "@"); i >= 0 {
		return c[:i], c[i+1:]
	}
	return c, ""
}

// ---- mirrors ----

type MirrorResult struct {
	Written bool
	Reason  string
	Files   int
	Sources int
}

// MirrorLaw stages the repo's law/ corpus read-only into inputs/law.
func MirrorLaw(repoLawDir, runDir string) MirrorResult {
	outDir := filepath.Join(runDir, "inputs", "law")
	if exists(outDir) {
		return MirrorResult{Written: false, Reason: "already staged"}
	}
	if repoLawDir == "" || !exists(repoLawDir) {
		return MirrorResult{Written: false, Reason: "no law dir"}
	}
	os.MkdirAll(outDir, 0o755)
	files := 0
	for _, f := range mdFiles(repoLawDir, nil) {
		body, err := os.ReadFile(filepath.Join(repoLawDir, f))
		if err != nil {
			continue
		}
		os.WriteFile(filepath.Join(outDir, f), append([]byte("<!-- mirrored from law/ at run setup — read-only copy -->\n"), body...), 0o644)
		files++
	}
	return MirrorResult{Written: files > 0, Files: files}
}

// MirrorGapPatterns concatenates red's gap-pattern memory into inputs/red-gap-patterns.md,
// first-source-wins deduped by filename. dirs are tried in order (promoted before raw).
func MirrorGapPatterns(memoryDirs []string, runDir string) MirrorResult {
	out := filepath.Join(runDir, "inputs", "red-gap-patterns.md")
	if exists(out) {
		return MirrorResult{Written: false, Reason: "already staged"}
	}
	present := existingDirs(memoryDirs)
	if len(present) == 0 {
		return MirrorResult{Written: false, Reason: "no memory dir"}
	}
	var parts []string
	seen := map[string]bool{}
	for _, dir := range present {
		for _, f := range mdFiles(dir, map[string]bool{"README.md": true}) {
			if seen[f] {
				continue
			}
			seen[f] = true
			body, err := os.ReadFile(filepath.Join(dir, f))
			if err != nil {
				continue
			}
			parts = append(parts, fmt.Sprintf("\n<!-- mirrored from %s: %s -->\n%s", dir, f, string(body)))
		}
	}
	if len(parts) == 0 {
		return MirrorResult{Written: false, Reason: "memory dir empty"}
	}
	// THE DIRECTORY IS MADE HERE, AND THE WRITE IS CHECKED.
	//
	// It was neither. `out` is <runDir>/inputs/red-gap-patterns.md and nothing in this function
	// created `inputs/`; the write's error was discarded; and the function then returned
	// {Written: true, Files: 55} having written nothing at all.
	//
	// In a real run it is masked, because BuildSkeleton happens to create inputs/ earlier in
	// run.go. Any other caller — a probe, a test, a reordering of setup — silently loses red's
	// entire accumulated memory while the result says it staged fifty-five files. MEASURED
	// 2026-08-16: found by reusing this function from seatprobe, where the directory does not
	// exist; every dispatch reported a staged corpus and no file was ever on disk.
	//
	// This is the defect class the corpus itself catalogues: a write that fails, reports success,
	// and hands back a count of work it did not do. A caller cannot tell the difference, which is
	// why the count was believable.
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return MirrorResult{Written: false, Reason: "could not create inputs/: " + err.Error()}
	}
	if err := os.WriteFile(out, []byte("# red gap-pattern inventory (mirrored at run setup — read-only copy)\n"+strings.Join(parts, "\n")), 0o644); err != nil {
		return MirrorResult{Written: false, Reason: "could not write the mirror: " + err.Error()}
	}
	return MirrorResult{Written: true, Files: len(parts), Sources: len(present)}
}

// ScorecardResult reports the chairs staged and the per-chair prompt headlines.
type ScorecardResult struct {
	Written   bool
	Reason    string
	Chairs    []string
	Headlines map[string][]string
}

// MirrorScorecards stages each chair's scorecard into inputs/ and extracts the
// prompt headline — the emitted HEADLINE line where present, else the parsed rows.
func MirrorScorecards(memoryDir, runDir string) ScorecardResult {
	if memoryDir == "" || !exists(memoryDir) {
		return ScorecardResult{Written: false, Reason: "no feov-memory dir", Headlines: map[string][]string{}}
	}
	var staged []string
	headlines := map[string][]string{}
	for _, f := range scorecardFiles(memoryDir) {
		chair := strings.TrimSuffix(f, "-scorecard.md")
		body, err := os.ReadFile(filepath.Join(memoryDir, f))
		if err != nil {
			continue
		}
		os.WriteFile(filepath.Join(runDir, "inputs", f), body, 0o644)
		staged = append(staged, chair)
		latest := lastSection(string(body))
		if h := emittedHeadline(latest); h != nil {
			headlines[chair] = h
			continue
		}
		if picks := fallbackHeadline(latest); len(picks) > 0 {
			headlines[chair] = picks
		}
	}
	if len(staged) == 0 {
		return ScorecardResult{Written: false, Reason: "no scorecards yet — written at capture, consumed by the next run", Chairs: []string{}, Headlines: map[string][]string{}}
	}
	return ScorecardResult{Written: true, Chairs: staged, Headlines: headlines}
}

var sectionSplit = regexp.MustCompile(`(?m)^## `)

// lastSection returns the text after the final "## " heading (the most recent run).
func lastSection(body string) string {
	parts := sectionSplit.Split(body, -1)
	return parts[len(parts)-1]
}

var headlineLine = regexp.MustCompile(`(?m)^HEADLINE:\s*(.+)$`)

func emittedHeadline(section string) []string {
	m := headlineLine.FindStringSubmatch(section)
	if m == nil {
		return nil
	}
	var out []string
	for _, s := range strings.Split(m[1], " · ") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return firstN(out, 3)
}

func fallbackHeadline(section string) []string {
	var picks []string
	for _, r := range ParseRenderedRows(section) {
		if r.Value == nil || r.Cls == "measure" {
			continue
		}
		picks = append(picks, fmt.Sprintf("%s %s [%s]", r.Metric, strings.TrimSpace(*r.Value), strings.ToUpper(r.Cls)))
	}
	return firstN(picks, 3)
}

// ---- rendered-row parser (shared markdown contract with scorecards.mjs) ----

// RenderedRow is one parsed scorecard row; Value nil == "_not computed_".
type RenderedRow struct {
	Metric string
	Cls    string
	Clause string
	Value  *string
}

// parseRenderedRowsRe was ported verbatim from the now-deleted scorecards.mjs parseRenderedRows
// (its behaviour is canonical in scorecard.ParseRenderedRows): a lazy clause (.+?) terminates at
// the first colon actually followed by a value, so a clause that itself contains a colon
// ("LOSS: additive violations") is read whole.
var parseRenderedRowsRe = regexp.MustCompile("`([a-z_]+)`\\s*\\[(benchmark|detector|diagnostic|measure)\\]\\s*—\\s*(.+?):\\s*(?:\\*\\*([^*]+)\\*\\*|_not computed_)")

func ParseRenderedRows(section string) []RenderedRow {
	var out []RenderedRow
	for _, m := range parseRenderedRowsRe.FindAllStringSubmatch(section, -1) {
		row := RenderedRow{Metric: m[1], Cls: m[2], Clause: m[3]}
		if m[4] != "" {
			v := m[4]
			row.Value = &v
		}
		out = append(out, row)
	}
	return out
}

// ---- pattern index ----

// PatternEntry is one classified gap pattern under a class key.
type PatternEntry struct {
	File  string `json:"file"`
	Title string `json:"title"`
	Hook  string `json:"hook"`
}

// PatternIndex is the class-join lookup table plus the two classless buckets.
type PatternIndex struct {
	ByClass      map[string][]PatternEntry
	Unclassified []string
	HarnessLimit []string
}

var (
	frontmatterRe  = regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	titleRe        = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	descRe         = regexp.MustCompile(`(?m)^description:\s*(.+)$`)
	classesRe      = regexp.MustCompile(`(?m)^\s*classes:\s*\[([^\]]*)\]`)
	harnessLimitRe = regexp.MustCompile(`(?m)^\s*class_note:\s*harness-limit`)
	quoteTrimRe    = regexp.MustCompile(`^["']|["']$`)
)

func BuildPatternIndex(memoryDirs []string) PatternIndex {
	idx := PatternIndex{ByClass: map[string][]PatternEntry{}, Unclassified: []string{}, HarnessLimit: []string{}}
	seen := map[string]bool{}
	skip := map[string]bool{"README.md": true, "MEMORY.md": true}
	for _, dir := range existingDirs(memoryDirs) {
		for _, f := range mdFiles(dir, skip) {
			if seen[f] {
				continue
			}
			seen[f] = true
			body, err := os.ReadFile(filepath.Join(dir, f))
			if err != nil {
				continue
			}
			s := string(body)
			var fm string
			if m := frontmatterRe.FindStringSubmatch(s); m != nil {
				fm = m[1]
			}
			desc := ""
			if fm != "" {
				if m := descRe.FindStringSubmatch(fm); m != nil {
					desc = quoteTrimRe.ReplaceAllString(strings.TrimSpace(m[1]), "")
				}
			}
			title := f
			if m := titleRe.FindStringSubmatch(s); m != nil {
				title = m[1]
			}
			var classes []string
			if fm != "" {
				if m := classesRe.FindStringSubmatch(fm); m != nil {
					for _, c := range strings.Split(m[1], ",") {
						c = quoteTrimRe.ReplaceAllString(strings.TrimSpace(c), "")
						if c != "" {
							classes = append(classes, c)
						}
					}
				}
			}
			if len(classes) == 0 {
				if fm != "" && harnessLimitRe.MatchString(fm) {
					idx.HarnessLimit = append(idx.HarnessLimit, f)
				} else {
					idx.Unclassified = append(idx.Unclassified, f)
				}
				continue
			}
			for _, c := range classes {
				idx.ByClass[c] = append(idx.ByClass[c], PatternEntry{File: f, Title: title, Hook: desc})
			}
		}
	}
	return idx
}

// ---- pin validation ----

// GitResult mirrors the fields the mjs reads off spawnSync: an exit Status and an Err,
// plus the Stdout `rev-parse` needs. Stdout exists so HEAD resolution goes through the SAME
// injected seam as every other git call — it used to shell out directly, which meant the pin
// gate could not be reached in a test at all: a temp cwd is not a repo, HEAD came back
// "unknown", and ValidatePins skipped itself. A gate whose arming depends on ambient state
// nobody can control is the same defect this file's preflight had.
type GitResult struct {
	Status int
	Err    error
	Stdout string
}

// GitFunc runs a git invocation; injectable for tests, defaults to the real git.
type GitFunc func(args []string) GitResult

func realGit(cwd string) GitFunc {
	return func(args []string) GitResult {
		c := exec.Command("git", args...)
		c.Dir = cwd
		out, err := c.Output()
		status := 0
		if err != nil {
			status = 1
			if ee, ok := err.(*exec.ExitError); ok {
				status = ee.ExitCode()
			}
		}
		return GitResult{Status: status, Err: nil, Stdout: string(out)}
	}
}

// PinCheck names a cite whose path is absent at its pin.
type PinCheck struct {
	Path string
	Pin  string
}

// PinValidation is the pin-validation verdict; Skipped set == non-git context.
type PinValidation struct {
	Checked int
	Missing []PinCheck
	Skipped string
}

func ValidatePins(cites []string, head string, git GitFunc) PinValidation {
	if head == "" || head == "unknown" {
		return PinValidation{Checked: 0, Missing: nil, Skipped: "not a git repo — pins UNVALIDATED"}
	}
	var missing []PinCheck
	for _, c := range cites {
		path, pin := splitPin(c)
		at := pin
		if at == "" {
			at = head
		}
		r := git([]string{"cat-file", "-e", fmt.Sprintf("%s:%s", at, path)})
		if r.Err != nil || r.Status != 0 {
			missing = append(missing, PinCheck{Path: path, Pin: at})
		}
	}
	return PinValidation{Checked: len(cites), Missing: missing}
}

// ---- run-live marker ----

// RunLiveMarker is the marker as a READ value. The writer's struct was anonymous, so every reader
// re-declared its own subset — seat.InferRunDir reads only RunDir, and the shape has never been
// stated in one place a new reader can find (#270).
type RunLiveMarker struct {
	RunDir      string   `json:"runDir"`
	PinnedPaths []string `json:"pinnedPaths"`
	Started     string   `json:"started"`
}

// ReadRunLiveMarker reports the open run, if the marker names one usably.
//
// ok=false covers absent, unreadable, unparseable and empty-runDir alike, and that is deliberate
// at THIS call site: setup's gate must refuse only when it can name the other run in its error.
// A marker it cannot read is not evidence of an open run — it is evidence of a broken file, and
// blocking a new run on one would trade a silent hazard for a stuck operator.
func ReadRunLiveMarker(projectDir string) (RunLiveMarker, bool) {
	b, err := os.ReadFile(filepath.Join(projectDir, ".claude", "run-live.json"))
	if err != nil {
		return RunLiveMarker{}, false
	}
	var m RunLiveMarker
	if json.Unmarshal(b, &m) != nil || strings.TrimSpace(m.RunDir) == "" {
		return RunLiveMarker{}, false
	}
	return m, true
}

// sameRun compares two run directories as PATHS, not as strings. The marker stores whatever it
// was given — `research/x` from one invocation, an absolute path from another — so a string
// compare would refuse a re-setup of the very run in progress, which is ordinary and must stay
// idempotent.
func sameRun(projectDir, a, b string) bool {
	abs := func(p string) string {
		if !filepath.IsAbs(p) {
			p = filepath.Join(projectDir, p)
		}
		if r, err := filepath.Abs(filepath.Clean(p)); err == nil {
			return r
		}
		return filepath.Clean(p)
	}
	return strings.EqualFold(abs(a), abs(b))
}

// WriteRunLiveMarker writes projectDir/.claude/run-live.json (commitment-as-state).
// `now` is injected so the sole non-deterministic field is controllable in tests.
func WriteRunLiveMarker(projectDir, runDir string, pinnedPaths []string, now time.Time) string {
	dir := filepath.Join(projectDir, ".claude")
	os.MkdirAll(dir, 0o755)
	p := filepath.Join(dir, "run-live.json")
	if pinnedPaths == nil {
		pinnedPaths = []string{}
	}
	marker := struct {
		RunDir      string   `json:"runDir"`
		PinnedPaths []string `json:"pinnedPaths"`
		Started     string   `json:"started"`
	}{runDir, pinnedPaths, now.UTC().Format("2006-01-02T15:04:05.000Z07:00")}
	b, _ := marshalJSON(marker)
	os.WriteFile(p, b, 0o644)
	return p
}

// ---- purge, preflight, version ----

// PurgeStaleMirrors removes checkpoint mirrors older than maxAgeDays (default 30).
func PurgeStaleMirrors(mirrorRoot string, now time.Time, maxAgeDays int) int {
	if !exists(mirrorRoot) {
		return 0
	}
	entries, err := os.ReadDir(mirrorRoot)
	if err != nil {
		return 0
	}
	purged := 0
	cutoff := time.Duration(maxAgeDays) * 24 * time.Hour
	for _, e := range entries {
		p := filepath.Join(mirrorRoot, e.Name())
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > cutoff {
			if os.RemoveAll(p) == nil {
				purged++
			}
		}
	}
	return purged
}

// expectedRecordVersion resolves what the binary at recordBin OUGHT to report.
//
// THE SECOND HALF OF THE SAME DEFECT. Arming the preflight is not enough if the number it
// compares against comes from the binary running `setup` — in the documented flow that is the
// SAME binary the seats will use, so the check reduces to a number compared with itself and
// passes by construction. That exact failure already shipped once: both sides read 0.1.0 for
// the whole of 2026-07-19 while the event schema changed underneath (see versionsync_test).
//
// The authority is requirements.json sitting beside the binary's plugin root — `bin/` in an
// installed cache, `tools/` in a dev tree, both one level up. That is the manifest shipped
// with the PROMPTS this run will drive, which is the thing the binary has to agree with. The
// compiled-in constant remains the fallback for a binary living outside a plugin layout,
// where there is no manifest to disagree with.
func expectedRecordVersion(recordBin, compiledIn string) string {
	manifest := filepath.Join(filepath.Dir(filepath.Dir(recordBin)), "requirements.json")
	if v := RecordToolVersion(manifest); v != "" {
		return v
	}
	return compiledIn
}

// RecordToolVersion reads recordToolVersion from requirements.json (the version authority).
func RecordToolVersion(manifestPath string) string {
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return ""
	}
	var m struct {
		RecordToolVersion string `json:"recordToolVersion"`
	}
	if json.Unmarshal(b, &m) != nil {
		return ""
	}
	return m.RecordToolVersion
}

// ExecResult mirrors the spawnSync fields the mjs preflight reads.
type ExecResult struct {
	Status int
	Err    error
	Stdout string
}

// ExecFunc runs a binary with args; injectable for tests.
type ExecFunc func(bin string, args []string) ExecResult

func realExec(bin string, args []string) ExecResult {
	out, err := exec.Command(bin, args...).Output()
	res := ExecResult{Stdout: string(out)}
	if err == nil {
		return res
	}
	if ee, ok := err.(*exec.ExitError); ok {
		res.Status = ee.ExitCode()
		return res
	}
	// Could not run at all. Node's spawnSync surfaces error.code (ENOENT) for a missing
	// binary; match it so the "not runnable" detail is byte-identical across the port.
	res.Status = 1
	if errors.Is(err, exec.ErrNotFound) {
		res.Err = errors.New("ENOENT")
	} else {
		res.Err = err
	}
	return res
}

// Preflight is the record-binary preflight verdict.
type Preflight struct {
	OK      bool
	Reason  string
	Remedy  string
	Version string
}

func PreflightRecordBinary(expectVersion, bin string, run ExecFunc) Preflight {
	r := run(bin, []string{"--version"})
	if r.Err != nil || r.Status != 0 {
		detail := fmt.Sprintf("exit %d", r.Status)
		if r.Err != nil {
			detail = r.Err.Error()
		}
		return Preflight{
			OK:     false,
			Reason: fmt.Sprintf("%s not runnable (%s)", bin, detail),
			Remedy: "install it with /prosthetic-conscience:doctor --fix, or build it: " +
				"(cd plugins/frank-exchange-of-views/tools && go build ./cmd/feov-record)",
		}
	}
	got := lastField(strings.TrimSpace(r.Stdout))
	if expectVersion != "" && got != "" && got != expectVersion {
		// NAME WHAT THE OPERATOR LOSES, not just the two numbers (#407). "yours is 0.66.0, the
		// plugin expects 0.68.0" is true and useless — it does not say whether to care. The
		// capability deltas answer exactly that, and they used to be unreachable prose in
		// root.go's changelog.
		reason := fmt.Sprintf("%s is %s, the plugin expects %s — a skewed binary writes events under a different contract", bin, got, expectVersion)
		if missing := record.DeltasBetween(got, expectVersion); len(missing) > 0 {
			reason += fmt.Sprintf("\n  what yours cannot do (%d change(s) behind):", len(missing))
			for _, m := range missing {
				reason += "\n    - " + m
			}
		}
		return Preflight{
			OK:     false,
			Reason: reason,
			Remedy: "refresh it with /prosthetic-conscience:doctor --fix (never mid-run)",
		}
	}
	return Preflight{OK: true, Version: got}
}

func (r ExecResult) errored() bool { return r.Err != nil || r.Status != 0 }

// ---- small helpers ----

func firstN[T any](s []T, n int) []T {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func lastField(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// existingDirs filters to the memory dirs that exist, preserving order (promoted first).
func existingDirs(memoryDirs []string) []string {
	var out []string
	for _, d := range memoryDirs {
		if d != "" && exists(d) {
			out = append(out, d)
		}
	}
	return out
}

// mdFiles lists *.md files in dir, SORTED (deterministic — the JS used readdir order),
// excluding the names in skip.
func mdFiles(dir string, skip map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".md") || skip[n] {
			continue
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// scorecardFiles lists *-scorecard.md files in dir, SORTED.
func scorecardFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), "-scorecard.md") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// StageClassRegistry copies the gap class registry into the run, where record.loadRegistry
// reads it.
//
// IT DID NOT EXIST UNTIL 2026-08-07 (#299). `loadRegistry` reads
// <runDir>/records/class-registry.json; nothing ever wrote that file; so it always returned
// nil, `validateClass` always took its advisory branch, and `--class` accepted ANY STRING on
// every run there has ever been.
//
// The cost was not a tidy-taxonomy complaint. Delivery of red's accumulated gap patterns is
// CLASS-INDEXED — the design that replaced whole-corpus staging after run 5 measured that
// worthless — and with nothing constraining the key, red invented a fresh vocabulary each
// run. Measured across both record-era runs: 10 and 14 minted classes, and ZERO overlap with
// the 37 classes the memory corpus is indexed by. The corpus was built, the index was
// written, and the join has never once delivered a pattern.
//
// Staging the file is what makes the key shared. A class not in it is UNDECLARED rather than
// forbidden: `--class-new` introduces one with its definition, neighbour and distinguisher,
// which is how the taxonomy is supposed to grow.
func StageClassRegistry(repoMemoryDir, runDir string) MirrorResult {
	src := filepath.Join(repoMemoryDir, "class-registry.json")
	b, err := os.ReadFile(src)
	if err != nil {
		return MirrorResult{Written: false, Reason: "no class-registry.json — `--class` will accept ANY string this run (#299)"}
	}
	var reg struct {
		Classes []struct {
			Slug string `json:"slug"`
		} `json:"classes"`
	}
	if json.Unmarshal(b, &reg) != nil || len(reg.Classes) == 0 {
		return MirrorResult{Written: false, Reason: "class-registry.json is unreadable or declares no classes — `--class` will accept ANY string this run (#299)"}
	}
	// RESOLVED, NOT JOINED — and this is the same bug as the one documented above, one
	// separation later. `loadRegistry` reads the record directory the RESOLVER returns; a
	// hard-wired <runDir>/records here would stage the registry where a separated run never
	// looks, `validateClass` would take its advisory branch again, and `--class` would go back
	// to accepting any string on exactly the runs the seat probe is built to measure.
	recDir, err := record.RecordsDir(runDir)
	if err != nil {
		return MirrorResult{Written: false, Reason: "cannot resolve the record directory: " + err.Error()}
	}
	dst := filepath.Join(recDir, "class-registry.json")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return MirrorResult{Written: false, Reason: "cannot create the record directory: " + err.Error()}
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		return MirrorResult{Written: false, Reason: "cannot stage the registry: " + err.Error()}
	}
	return MirrorResult{Written: true, Files: len(reg.Classes), Sources: 1}
}

// RegistrySlugs reads the staged registry's slugs, for the memory-join report.
func RegistrySlugs(repoMemoryDir string) map[string]bool {
	out := map[string]bool{}
	b, err := os.ReadFile(filepath.Join(repoMemoryDir, "class-registry.json"))
	if err != nil {
		return out
	}
	var reg struct {
		Classes []struct {
			Slug string `json:"slug"`
		} `json:"classes"`
	}
	if json.Unmarshal(b, &reg) != nil {
		return out
	}
	for _, c := range reg.Classes {
		out[c.Slug] = true
	}
	return out
}
