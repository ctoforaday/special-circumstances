package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Config is the parsed setup invocation. The environment fields (Cwd/Home/ProjectDir/
// ExpectVersion) and the injectables (Git/Exec/Now) are resolved by the command from os;
// tests may supply them directly.
type Config struct {
	RunDir        string
	Topic         string
	Model         string
	JudgmentModel string
	Cites         []string
	MaxRounds     string
	Lanes         string
	BinDir        string
	MemoryDir     string

	Cwd           string
	Home          string
	ProjectDir    string // CLAUDE_PROJECT_DIR
	ExpectVersion string // the running binary's version (== requirements.json recordToolVersion)
	Git           GitFunc
	Exec          ExecFunc
	Now           time.Time
}

// Run reproduces the mjs main(): the four fail-fast gates (runDir→1; model tiers→2;
// pin miss→2; --bin-dir preflight→2), then the run-dir build + mirrors + marker,
// then the summary. Returns the process exit code; writes to the given streams.
func Run(cfg Config, stdout, stderr io.Writer) int {
	if cfg.Git == nil {
		cfg.Git = realGit(cfg.Cwd)
	}
	if cfg.Exec == nil {
		cfg.Exec = realExec
	}
	if cfg.Now.IsZero() {
		cfg.Now = time.Now()
	}

	if cfg.RunDir == "" || strings.HasPrefix(cfg.RunDir, "--") {
		fmt.Fprintln(stderr, `usage: feov-record setup <runDir> --topic "<topic>" --model <m> --judgment-model <m> [--cite <path>[@pin]]... [--bin-dir <dir>] [--memory-dir <dir>]`)
		return 1
	}
	topic := cfg.Topic
	if topic == "" {
		topic = "(topic not stated)"
	}
	head := gitHead(cfg.Cwd)

	// Gate: model tiers REQUIRED (#111), before the run dir exists.
	if cfg.Model == "" || cfg.JudgmentModel == "" {
		fmt.Fprintln(stderr, "run-setup: MODEL TIERS REQUIRED — refusing to create the run:")
		if cfg.Model == "" {
			fmt.Fprintln(stderr, "  - --model is unset (the BULK tier: frontier, blue lanes, red lenses, blue responses)")
		}
		if cfg.JudgmentModel == "" {
			fmt.Fprintln(stderr, "  - --judgment-model is unset (the JUDGMENT tier: blue-synthesize, red-merge, judge, assemble)")
		}
		fmt.Fprintln(stderr, "  the engine does not guess a tier or inherit the session model. Pass both, e.g.")
		fmt.Fprintln(stderr, "  --model sonnet --judgment-model sonnet   (a smoke run passes --model haiku --judgment-model haiku)")
		return 2
	}

	// Gate: pins validated before anything is built.
	pv := ValidatePins(cfg.Cites, head, cfg.Git)
	if len(pv.Missing) > 0 {
		fmt.Fprintln(stderr, "run-setup: PIN VALIDATION FAILED — refusing to create the run:")
		for _, m := range pv.Missing {
			fmt.Fprintf(stderr, "  - %s does not exist at pin %s (git cat-file -e %s:%s)\n", m.Path, m.Pin, m.Pin, m.Path)
		}
		fmt.Fprintln(stderr, "  remedies: fix the cite (right path / right pin), or stage the artifact into")
		fmt.Fprintln(stderr, "  <runDir>/inputs/ AND COMMIT IT before setup, then cite the committed copy.")
		fmt.Fprintln(stderr, "  Staging alone is NOT enough: an uncommitted file exists at no pin, so it")
		fmt.Fprintln(stderr, "  cannot be cited — evidence that can still change underneath the run is not")
		fmt.Fprintln(stderr, "  evidence. (setup keeps pre-staged files; committing is the missing step.)")
		return 2
	}

	// Gate: record-binary preflight — bound to INTENT (--bin-dir), before any run state.
	recordBin := "feov-record"
	if cfg.BinDir != "" {
		recordBin = filepath.Join(cfg.BinDir, "feov-record")
	}
	pre := PreflightRecordBinary(cfg.ExpectVersion, recordBin, cfg.Exec)
	if !pre.OK && cfg.BinDir != "" {
		fmt.Fprintln(stderr, "run-setup: RECORD BINARY PREFLIGHT FAILED — refusing to create the run:")
		fmt.Fprintf(stderr, "  %s\n", pre.Reason)
		fmt.Fprintf(stderr, "  remedy: %s\n", pre.Remedy)
		fmt.Fprintln(stderr, "  (failing here costs a re-run; failing mid-round costs a seat its whole record)")
		return 2
	}

	skel := BuildSkeleton(cfg.RunDir, topic)
	mirrorsPurged := PurgeStaleMirrors(filepath.Join(cfg.Home, ".cache", "feov", "run-mirror"), cfg.Now, 30)
	if mirrorsPurged > 0 {
		fmt.Fprintf(stdout, "  mirror purge: %d stale checkpoint mirror(s) removed\n", mirrorsPurged)
	}
	pinned := BuildPinned(cfg.RunDir, head, cfg.Cites)

	// Memory dirs: --memory-dir wins; else [promoted, raw], promoted first.
	memHome := func(d string) string {
		return filepath.Join(d, ".claude", "agent-memory", "frank-exchange-of-views-red-auditor")
	}
	promoted := filepath.Join(cfg.Cwd, "feov-memory", "red-gap-patterns")
	raw := ""
	for _, d := range []string{cfg.ProjectDir, cfg.Cwd} {
		if d == "" {
			continue
		}
		if h := memHome(d); exists(h) {
			raw = h
			break
		}
	}
	memDirs := []string{promoted, raw}
	if cfg.MemoryDir != "" {
		memDirs = []string{cfg.MemoryDir}
	}
	mirror := MirrorGapPatterns(memDirs, cfg.RunDir)
	patternIndex := BuildPatternIndex(memDirs)
	if b, err := marshalJSON(patternIndex.ByClass); err == nil {
		os.WriteFile(filepath.Join(cfg.RunDir, "inputs", "gap-patterns-by-class.json"), b, 0o644)
	}

	rc := runConfig{Topic: topic, Model: cfg.Model, JudgmentModel: cfg.JudgmentModel, MaxRounds: ptrOrNil(cfg.MaxRounds), Lanes: ptrOrNil(cfg.Lanes)}
	if b, err := marshalJSON(rc); err == nil {
		os.WriteFile(filepath.Join(cfg.RunDir, "inputs", "run-config.json"), b, 0o644)
	}

	law := MirrorLaw(filepath.Join(cfg.Cwd, "law"), cfg.RunDir)
	cards := MirrorScorecards(filepath.Join(cfg.Cwd, "feov-memory"), cfg.RunDir)
	pinnedPaths := []string{}
	for _, c := range cfg.Cites {
		p, _ := splitPin(c)
		pinnedPaths = append(pinnedPaths, p)
	}
	marker := WriteRunLiveMarker(cfg.Cwd, cfg.RunDir, pinnedPaths, cfg.Now)

	fmt.Fprintf(stdout, "run-setup: %s\n", cfg.RunDir)
	fmt.Fprintf(stdout, "  skeleton: %d created, %d pre-staged (kept)\n", len(skel.Created), len(skel.Skipped))
	fmt.Fprintln(stdout, "  NOT created (red-merge-born): red/ledger.md, red/archive.md, trajectories/board-telemetry.jsonl")
	if pinned.Written {
		fmt.Fprintf(stdout, "  pinned: HEAD %s + %d cited path(s)\n", head, len(cfg.Cites))
	} else {
		fmt.Fprintln(stdout, "  pinned: inputs/PINNED.md pre-staged (kept)")
	}
	if pv.Skipped != "" {
		fmt.Fprintf(stdout, "  pin validation: %s\n", pv.Skipped)
	} else {
		fmt.Fprintf(stdout, "  pin validation: %d cite(s) verified at their pins\n", pv.Checked)
	}
	if mirror.Written {
		fmt.Fprintf(stdout, "  gap-patterns: %d pattern(s) mirrored from %d source(s) (promoted corpus first)\n", mirror.Files, mirror.Sources)
	} else {
		fmt.Fprintf(stdout, "  gap-patterns: %s\n", mirror.Reason)
	}
	if law.Written {
		fmt.Fprintf(stdout, "  law: %d file(s) mirrored (statute > precedent > argument)\n", law.Files)
	} else {
		fmt.Fprintf(stdout, "  law: %s\n", law.Reason)
	}
	idxLine := fmt.Sprintf("  gap-pattern index: %d class(es) -> inputs/gap-patterns-by-class.json", len(patternIndex.ByClass))
	if len(patternIndex.Unclassified) > 0 {
		idxLine += fmt.Sprintf(" (%d UNCLASSIFIED, not delivered — classify them to make them bind)", len(patternIndex.Unclassified))
	}
	if len(patternIndex.HarnessLimit) > 0 {
		idxLine += fmt.Sprintf(" (%d harness-limit, classless by design)", len(patternIndex.HarnessLimit))
	}
	fmt.Fprintln(stdout, idxLine)
	if cards.Written {
		fmt.Fprintf(stdout, "  scorecards: %s staged into inputs/\n", strings.Join(cards.Chairs, ", "))
	} else {
		fmt.Fprintf(stdout, "  scorecards: %s\n", cards.Reason)
	}
	if len(cards.Headlines) > 0 {
		if b, err := marshalJSON(cards.Headlines); err == nil {
			os.WriteFile(filepath.Join(cfg.RunDir, "inputs", "scorecards.json"), b, 0o644)
		}
		fmt.Fprintln(stdout, `  scorecards arg: pass inputs/scorecards.json as the workflow's "scorecards" arg`)
		fmt.Fprintf(stdout, "    %s\n", compactJSON(cards.Headlines))
	}
	if pre.OK {
		v := pre.Version
		if v == "" {
			v = "(version unreported)"
		}
		fmt.Fprintf(stdout, "  record binary: %s %s\n", recordBin, v)
	} else {
		fmt.Fprintf(stdout, "  record binary: NOT AVAILABLE — %s (this run will not record through the tool; pass --bin-dir to require it)\n", pre.Reason)
	}
	fmt.Fprintf(stdout, "  run-live marker: %s\n", marker)
	return 0
}

type runConfig struct {
	Topic         string  `json:"topic"`
	Model         string  `json:"model"`
	JudgmentModel string  `json:"judgmentModel"`
	MaxRounds     *string `json:"maxRounds"`
	Lanes         *string `json:"lanes"`
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func okFailed(b bool) string {
	if b {
		return "ok"
	}
	return "FAILED"
}

// gitHead runs `git rev-parse --short HEAD` in cwd; "unknown" if it fails (non-git).
func gitHead(cwd string) string {
	c := exec.Command("git", "rev-parse", "--short", "HEAD")
	c.Dir = cwd
	out, err := c.Output()
	if err != nil {
		return "unknown"
	}
	h := strings.TrimSpace(string(out))
	if h == "" {
		return "unknown"
	}
	return h
}

// compactJSON matches JS `JSON.stringify(x)` — no indent, no HTML escaping.
func compactJSON(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if enc.Encode(v) != nil {
		return ""
	}
	return strings.TrimRight(buf.String(), "\n")
}
