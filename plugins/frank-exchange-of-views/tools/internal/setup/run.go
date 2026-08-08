package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	head := gitHead(cfg.Git)

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

	// Gate: the ROUND CEILING is required, for the reason the model tiers are (#308).
	//
	// It was optional, and debate.js defaults it to 12 in JS — so a run launched without it had
	// a ceiling nobody recorded, and CEILING became underivable from the record: the verdict
	// fell back to the seat's word. Found by the fuzz tripwire, which flagged 24 of 60 runs as
	// carrying an ASSERTED verdict purely because no ceiling was on file.
	//
	// The operator resolves this value anyway to hand the Workflow, and /research already says
	// to pass setup the SAME config. The engine does not guess a tier; it should not guess a
	// bound either.
	if cfg.MaxRounds == "" {
		fmt.Fprintln(stderr, "run-setup: --max-rounds REQUIRED — refusing to create the run:")
		fmt.Fprintln(stderr, "  the round ceiling is the bound the terminal CEILING verdict is derived against, and a")
		fmt.Fprintln(stderr, "  run that does not record it leaves its own outcome underivable — the verdict falls back")
		fmt.Fprintln(stderr, "  to whatever a seat says it was. Pass the same value you will hand the workflow, e.g.")
		fmt.Fprintln(stderr, "  --max-rounds 12   (a smoke run passes --max-rounds 2)")
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

	// Gate: record-binary preflight — UNCONDITIONAL, before any run state.
	//
	// It used to be armed by INTENT (`!pre.OK && cfg.BinDir != ""`), and `--bin-dir` was never
	// passed on the documented launch — `grep -rn "bin-dir"` across the plugin returned
	// nothing. So the guarantee `commands/research.md` states ("setup preflights the binary's
	// version before the run exists, so a missing or skewed one fails there rather than
	// mid-round") was false: the real path printed a WARNING and proceeded, and the 2026-08-05
	// smoke's first setup did exactly that. A run would then take the legacy prompt set,
	// record nothing through the tool, and every axis added since would be silently absent
	// with every gate green. `--bin-dir` now says only WHERE the seats' binary is; whether the
	// run records through the tool is the Workflow's `binDir`, which is a different decision
	// at a different seam.
	recordBin := "feov-record"
	if cfg.BinDir != "" {
		recordBin = filepath.Join(cfg.BinDir, "feov-record")
	} else if self, err := os.Executable(); err == nil {
		// The documented flow runs THIS binary and hands the seats the same one, so its own
		// directory is the honest default — better than a bare PATH lookup, which refuses an
		// operator who is demonstrably holding a working binary.
		recordBin = filepath.Join(filepath.Dir(self), "feov-record")
	}
	pre := PreflightRecordBinary(expectedRecordVersion(recordBin, cfg.ExpectVersion), recordBin, cfg.Exec)
	if !pre.OK {
		fmt.Fprintln(stderr, "run-setup: RECORD BINARY PREFLIGHT FAILED — refusing to create the run:")
		fmt.Fprintf(stderr, "  %s\n", pre.Reason)
		fmt.Fprintf(stderr, "  remedy: %s\n", pre.Remedy)
		fmt.Fprintln(stderr, "  (failing here costs a re-run; failing mid-round costs a seat its whole record)")
		return 2
	}

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
	// --memory-dir ADDS a source; it does NOT replace the promoted corpus.
	//
	// MEASURED, and it cost a whole run's memory. It used to replace, and
	// `commands/research.md` documents passing it as the remedy when gap-patterns reports "no
	// memory dir" — so an operator following the documented advice silently discarded the
	// curated corpus (57 files, 55 classified) in favour of the raw accrual (60 files, 1
	// classified). The 2026-08-05 run's inputs/gap-patterns-by-class.json: 0 classes, 0
	// entries. Red opened that run with nothing, and the setup summary said so in one line
	// nobody read.
	//
	// Promoted stays FIRST: BuildPatternIndex dedupes by filename, so the reviewed copy of a
	// pattern wins over the raw one it was promoted from.
	memDirs := []string{promoted, raw}
	if cfg.MemoryDir != "" {
		memDirs = append(memDirs, cfg.MemoryDir)
	}
	patternIndex := BuildPatternIndex(memDirs)
	// A corpus that is MOSTLY unclassified is a composition failure, not sloppy authoring.
	//
	// The count was already printed — "(59 UNCLASSIFIED, not delivered)" — and it is one line
	// in a long summary, so it read as a nag rather than as "red is starting this run blind".
	// Delivery is class-indexed: an unclassified pattern reaches no seat at all. A handful is
	// normal accrual; a majority means the sources are wrong, which is exactly what the old
	// replacing --memory-dir produced.
	if delivered := len(patternIndex.ByClass); len(patternIndex.Unclassified) > 0 && len(patternIndex.Unclassified) > delivered {
		fmt.Fprintln(stderr, "run-setup: GAP-PATTERN CORPUS MOSTLY UNCLASSIFIED — refusing to create the run:")
		fmt.Fprintf(stderr, "  %d unclassified pattern(s) against %d delivered class(es).\n", len(patternIndex.Unclassified), delivered)
		fmt.Fprintln(stderr, "  Delivery is class-indexed, so an unclassified pattern reaches no seat: red would open this run")
		fmt.Fprintln(stderr, "  substantially blind while its memory directory looks full. Sources read, in order:")
		for _, d := range memDirs {
			if d != "" {
				fmt.Fprintf(stderr, "    - %s\n", d)
			}
		}
		fmt.Fprintln(stderr, "  remedy: check the promoted corpus is among them (feov-memory/red-gap-patterns), or classify")
		fmt.Fprintln(stderr, "  the accrued files by adding `classes: [<slug>, ...]` to their frontmatter.")
		return 2
	}

	skel := BuildSkeleton(cfg.RunDir, topic)
	mirrorsPurged := PurgeStaleMirrors(filepath.Join(cfg.Home, ".cache", "feov", "run-mirror"), cfg.Now, 30)
	if mirrorsPurged > 0 {
		fmt.Fprintf(stdout, "  mirror purge: %d stale checkpoint mirror(s) removed\n", mirrorsPurged)
	}
	pinned := BuildPinned(cfg.RunDir, head, cfg.Cites)

	// The registry is staged BEFORE any seat can mint, because it is what makes `--class` mean
	// anything at all (#299).
	registry := StageClassRegistry(filepath.Join(cfg.Cwd, "feov-memory"), cfg.RunDir)

	// The index was built and GATED above, before any run state existed; this only mirrors the
	// files into the run and writes the class join the engine hands to a repairing seat.
	mirror := MirrorGapPatterns(memDirs, cfg.RunDir)
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
	// The registry decides whether `--class` means anything this run, so it is reported before
	// the corpus that joins on it.
	if registry.Written {
		fmt.Fprintf(stdout, "  class registry: %d class(es) staged — `--class` is validated; `--class-new` extends it\n", registry.Files)
	} else {
		fmt.Fprintf(stdout, "  class registry: NOT STAGED — %s\n", registry.Reason)
	}
	if mirror.Written {
		fmt.Fprintf(stdout, "  gap-patterns: %d pattern(s) mirrored from %d source(s) (promoted corpus first)\n", mirror.Files, mirror.Sources)
		// THE JOIN'S HEALTH, stated rather than assumed. Patterns are delivered by matching the
		// class of the gap in front of a seat, so a corpus indexed by classes the registry does
		// not contain reaches nobody however well it is composed — which is what both
		// record-era runs did, at zero overlap.
		if slugs := RegistrySlugs(filepath.Join(cfg.Cwd, "feov-memory")); len(slugs) > 0 {
			joinable, orphaned := 0, []string{}
			for class := range patternIndex.ByClass {
				if slugs[class] {
					joinable++
				} else {
					orphaned = append(orphaned, class)
				}
			}
			sort.Strings(orphaned)
			fmt.Fprintf(stdout, "    class join: %d of %d indexed class(es) exist in the registry\n", joinable, len(patternIndex.ByClass))
			if len(orphaned) > 0 {
				fmt.Fprintf(stdout, "    NOT joinable (indexed by a class no gap can carry): %s\n", strings.Join(firstN(orphaned, 8), ", "))
			}
		}
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
	// Reached only when the preflight PASSED — it refuses above now, so there is no
	// "NOT AVAILABLE" line any more. That line used to be the whole failure mode: it told the
	// operator the run would not record through the tool and then created the run anyway.
	v := pre.Version
	if v == "" {
		v = "(version unreported)"
	}
	fmt.Fprintf(stdout, "  record binary: %s %s\n", recordBin, v)
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

// gitHead resolves HEAD through the INJECTED git seam; "unknown" if it fails (non-git).
func gitHead(git GitFunc) string {
	r := git([]string{"rev-parse", "--short", "HEAD"})
	if r.Err != nil || r.Status != 0 {
		return "unknown"
	}
	h := strings.TrimSpace(r.Stdout)
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
