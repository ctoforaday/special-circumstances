package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
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
	// RunID and ScriptPath travel into the run-live marker so a STALE marker names how to
	// resume rather than only where something once ran. Optional: a launcher that does not
	// know them leaves them empty, and the marker omits the fields rather than carrying "".
	RunID      string
	ScriptPath string
	// AllowSubstitution is the operator's standing consent to a served model that is not the
	// configured tier. It is recorded on the run rather than passed per seat: the seat is the
	// party whose adversary strength is in question, so the decision cannot be its to make.
	AllowSubstitution bool

	Cwd        string
	Home       string
	ProjectDir string // CLAUDE_PROJECT_DIR
	Git        GitFunc
	Exec       ExecFunc
	Now        time.Time
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

	// ANOTHER OPEN RUN IS REPORTED, NOT REFUSED (#529).
	//
	// The marker is commitment-as-state and its only retraction is `capture`, which is optional:
	// a run that is killed, throws, or is simply never captured leaves its row behind. That is
	// still worth telling a human at the one moment one is present, and it is collected here so
	// the summary can say it.
	//
	// It USED to refuse, for a reason its own comment gave: the stale marker was what
	// runlive.InferRunDir handed to every verb invoked without --run. #526 removed that — setup
	// bakes the run into <runDir>/.bin/feov-record, so a seat carries its run in its own
	// environment and never has to ask — and ReadRunLiveMarker now answers "no single run" when
	// more than one is open, so inference declines rather than guesses. With the misdirection
	// gone, refusing only forced the operator to defeat the guard by hand, which is how an
	// operator learns to reach for `rm` on the one file that says a run is open.
	var alsoLive []runlive.RunLiveMarker
	for _, m := range runlive.ReadRunLive(cfg.Cwd) {
		if !runlive.SameRun(cfg.Cwd, m.RunDir, cfg.RunDir) {
			alsoLive = append(alsoLive, m)
		}
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
	// The manifest beside the binary is the only authority, and an unreadable one REFUSES
	// rather than falling back to the binary's own number — that fallback made the check
	// compare a value with itself in the documented flow, where setup and the seats are the
	// same binary.
	expect, err := expectedSchema(recordBin)
	if err != nil {
		fmt.Fprintln(stderr, "run-setup: RECORD BINARY PREFLIGHT FAILED — refusing to create the run:")
		fmt.Fprintf(stderr, "  %v\n", err)
		fmt.Fprintln(stderr, "  remedy: reinstall the plugin so its requirements.json sits beside the binary")
		return 2
	}
	pre := PreflightRecordBinary(expect, recordBin, cfg.Exec)
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

	// RESOLVED ONCE, BEFORE ANYTHING IS WRITTEN. Everything below used to take the raw string
	// and each site decided for itself whether to make it absolute — so a run was laid out with
	// a mix of spellings, and the one absolute form was computed sixty lines later with a silent
	// fallback to the relative path when it failed.
	//
	// NewRun rather than OpenRun because this is the caller that legitimately holds a run before
	// there is anything on disk to open: BuildSkeleton, three lines down, is what creates it.
	run, err := record.NewRun(cfg.RunDir)
	if err != nil {
		fmt.Fprintf(stderr, "run-setup: %v\n", err)
		return 2
	}

	skel := BuildSkeleton(run, topic)
	if mirrorRoot, mErr := record.MirrorRoot(); mErr != nil {
		// LOUD, not folded into the zero: a purge that could not resolve its own directory has
		// not checked anything, and reporting that as "0 removed" is the same line a clean
		// board prints.
		fmt.Fprintf(stderr, "  mirror purge: NOT RUN — %v\n", mErr)
	} else if n := record.PurgeStaleMirrors(mirrorRoot, cfg.Now, 30); n > 0 {
		fmt.Fprintf(stdout, "  mirror purge: %d stale checkpoint mirror(s) removed\n", n)
	}
	pinned := BuildPinned(run, head, cfg.Cites)

	// The registry is staged BEFORE any seat can mint, because it is what makes `--class` mean
	// anything at all (#299).
	registry := StageClassRegistry(filepath.Join(cfg.Cwd, "feov-memory"), run)

	// The index was built and GATED above, before any run state existed; this only mirrors the
	// files into the run and writes the class join the engine hands to a repairing seat.
	mirror := MirrorGapPatterns(memDirs, run)

	// THE CLASS JOIN IS THE DELIVERY CHANNEL, so a failure to write it is the same condition the
	// unclassified gate forty lines up refuses the run over: red opens blind while its memory
	// looks full.
	//
	// This used to be `if b, err := marshalJSON(...); err == nil { os.WriteFile(...) }` — the
	// marshal error skipped the write silently, and the write error was discarded outright. The
	// summary below then printed `gap-pattern index: N class(es) -> inputs/gap-patterns-by-class.json`
	// regardless, because N comes from the in-memory index and never from the file. A run whose
	// join never landed and a run whose join landed perfectly printed the same line.
	//
	// Same defect, same package, one commit apart: MirrorGapPatterns returned `Written: true,
	// Files: 55` on a discarded write error. That one was caught by reusing it somewhere the
	// caller had not already created `inputs/`; this one sits on the line that actually feeds a
	// seat, and nothing was reusing it.
	joinPath := filepath.Join(run.Dir(), "inputs", "gap-patterns-by-class.json")
	b, err := marshalJSON(patternIndex.ByClass)
	if err != nil {
		fmt.Fprintf(stderr, "run-setup: could not encode the gap-pattern class join: %v\n", err)
		fmt.Fprintln(stderr, "  Delivery is class-indexed, so without this file red opens the run with no patterns at all.")
		return 2
	}
	if err := os.MkdirAll(filepath.Dir(joinPath), 0o755); err != nil {
		fmt.Fprintf(stderr, "run-setup: could not create inputs/ for the gap-pattern class join: %v\n", err)
		return 2
	}
	if err := os.WriteFile(joinPath, b, 0o644); err != nil {
		fmt.Fprintf(stderr, "run-setup: could not write the gap-pattern class join to %s: %v\n", joinPath, err)
		fmt.Fprintln(stderr, "  Delivery is class-indexed, so without this file red opens the run with no patterns at all.")
		return 2
	}

	rc := runConfig{Topic: topic, RunDir: run.Dir(), Model: cfg.Model, JudgmentModel: cfg.JudgmentModel, MaxRounds: ptrOrNil(cfg.MaxRounds), Lanes: ptrOrNil(cfg.Lanes), EventSchema: expect, AllowModelSubstitution: cfg.AllowSubstitution}
	if b, err := marshalJSON(rc); err == nil {
		os.WriteFile(filepath.Join(run.Dir(), "inputs", "run-config.json"), b, 0o644)
	}

	law := MirrorLaw(filepath.Join(cfg.Cwd, "law"), run)
	cards := MirrorScorecards(filepath.Join(cfg.Cwd, "feov-memory"), run)
	pinnedPaths := []string{}
	for _, c := range cfg.Cites {
		p, _ := splitPin(c)
		pinnedPaths = append(pinnedPaths, p)
	}
	marker := runlive.WriteRunLiveMarker(cfg.Cwd, cfg.RunDir, pinnedPaths, cfg.Now, cfg.RunID, cfg.ScriptPath)
	// Written AFTER the marker deliberately: the wrapper is the carrier that does not move when
	// the marker does, and writing it second makes the ordering obvious to anyone reading for
	// which of the two is derived from the other. Neither is — that is the point.
	wrapper := WriteRunWrapper(run, recordBin)

	fmt.Fprintf(stdout, "run-setup: %s\n", cfg.RunDir)
	fmt.Fprintf(stdout, "  skeleton: %d created, %d pre-staged (kept)\n", len(skel.Created), len(skel.Skipped))
	// "red-merge-born" was true when the merge wrote those files. It stopped being true, and the
	// line went on implying a writer would arrive — the same promise the husk stubs made.
	fmt.Fprintln(stdout, "  NOT created (rendered from the record on read, never materialized): the ledger,")
	fmt.Fprintln(stdout, "  debate transcript, evidence layer and board telemetry — `<role> show <name>`")
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
			os.WriteFile(filepath.Join(run.Dir(), "inputs", "scorecards.json"), b, 0o644)
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
	// LOUD, AND ACTIONABLE PER RUN. A notice that only said "something else is open" would be
	// the skimmable version of the refusal it replaces; each row names its own capture command.
	for _, m := range alsoLive {
		fmt.Fprintf(stdout, "  ALSO OPEN: %s (started %s) — its capture never ran, so nothing closed it.\n", m.RunDir, m.Started)
		fmt.Fprintf(stdout, "    close it with:  feov-record capture %s <its transcript dir>\n", m.RunDir)
	}
	if len(alsoLive) > 0 {
		fmt.Fprintln(stdout, "    Both runs proceed: each seat carries its own run in its wrapper, so neither can")
		fmt.Fprintln(stdout, "    file work against the other. A verb invoked with NO run directory now resolves")
		fmt.Fprintln(stdout, "    nothing rather than guessing between them.")
	}
	// THE binDir LINE IS AN INSTRUCTION, not a report, so it says what to do with the path.
	// Handing the workflow the raw binary directory instead of this one is silent: the run
	// works, and every seat is back to typing --run.
	if wrapper.BinDir != "" {
		fmt.Fprintln(stdout, `  binDir arg: pass THIS path as the workflow's "binDir" arg (not the record binary's directory)`)
		fmt.Fprintf(stdout, "    %s\n", wrapper.BinDir)
		fmt.Fprintln(stdout, "    It wraps the binary with this run baked in, so no seat types --run and a mistyped")
		fmt.Fprintln(stdout, "    path fails at the shell instead of filing work against nothing.")
	} else {
		// NOT SILENT ON THE MISS. Without this line the summary looks identical to a healthy
		// one and the operator hands over the raw directory believing it is the wrapper.
		fmt.Fprintf(stdout, "  binDir arg: NO WRAPPER WRITTEN — %s\n", wrapper.Skipped)
		fmt.Fprintln(stdout, "    Pass the record binary's own directory, and expect seats to carry --run themselves.")
	}
	return 0
}

type runConfig struct {
	Topic string `json:"topic"`
	// RunDir is the ABSOLUTE path of this run.
	//
	// It is recorded because the run directory reaches a seat as a STRING each seat resolves
	// against its own working directory, and a relative one resolves differently per invocation.
	// Measured (#358): a seat whose shell cwd was the `tools/` directory resolved
	// `research/<slug>/` from there and built a second blackboard — the lane's entire draft, its
	// own shards, clock and locks — while the real run's candidates directory stayed empty. Two
	// shards of one seat class existed in both places.
	//
	// A seat can no longer CREATE a run directory (RegisterSeat refuses one that does not exist),
	// which turns that failure loud. This field is the other half: the operator and every
	// post-hoc reader can see which absolute path the run was set up at, rather than inferring it
	// from wherever they happen to be standing.
	RunDir        string  `json:"runDir"`
	Model         string  `json:"model"`
	JudgmentModel string  `json:"judgmentModel"`
	MaxRounds     *string `json:"maxRounds"`
	Lanes         *string `json:"lanes"`
	// EventSchema is the event-shape EPOCH this run's events were written under.
	//
	// A run directory is created by `setup` and does not outlive the schema that made it — but
	// `run-archive/` does, and CLAUDE.md calls it the only part of a run that survives the
	// container, re-read by every later audit. When the record's storage moved to a database,
	// six archived runs became unreadable by the current binary, and nothing in them said what
	// shape they were in — so a later reader could only guess which binary to build.
	//
	// It is the EPOCH and deliberately not a binary version (#597): a release number moves for
	// reasons the event shape does not care about, so recording one would tell a future reader
	// what shipped rather than what it can read. This is the same number the setup preflight
	// just checked the binary against, so the file states a fact the run has already verified.
	EventSchema int `json:"eventSchema,omitempty"`
	// AllowModelSubstitution records that the OPERATOR accepted, before the run, an environment
	// that may answer with a model other than the configured tier. Absent means no: a run whose
	// config predates the field never consented, and the failing direction of a gate has to be
	// the safe one.
	//
	// It is written here and read at `register` (record.allowSubstitution), which is what makes
	// the consent a FIELD an operator set rather than a flag a seat could type for itself.
	AllowModelSubstitution bool `json:"allowModelSubstitution,omitempty"`
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
