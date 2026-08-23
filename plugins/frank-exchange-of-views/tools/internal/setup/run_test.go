package setup

import (
	"bytes"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// THE GATE THAT NEVER FIRED.
//
// `Run` is the function that decides whether a version-skewed binary stops a run, and it had
// NO test — only `PreflightRecordBinary`, the pure helper it calls, did. So the helper was
// correct the whole time while the decision around it was armed on `--bin-dir`, a flag
// `grep -rn "bin-dir"` found nowhere in the plugin. The unit was tested; the wiring was not,
// which is the shape of every defect in this file's history.
//
// Measured consequence: the 2026-08-05 smoke's first `setup` printed
// "record binary: NOT AVAILABLE … (this run will not record through the tool)" and CREATED
// THE RUN ANYWAY. Had that gone unread, the debate would have taken the legacy prompt set,
// recorded nothing, and every axis added since would have been silently absent with every
// gate green — while `commands/research.md` promised the opposite in writing.

// runCfg is a Config with the environment injected, so no test touches real git or a real
// binary. Cwd/Home/ProjectDir point at t.TempDir() so mirrors and markers land in the sandbox.
func runCfg(t *testing.T, version string, exec ExecFunc) (Config, string) {
	t.Helper()
	home := t.TempDir()
	runDir := filepath.Join(home, "research", "2026-01-01_probe")
	return Config{
		RunDir:        runDir,
		Topic:         "a probe",
		Model:         "haiku",
		JudgmentModel: "haiku",
		MaxRounds:     "3", // required (#308): the bound CEILING is derived against
		Cwd:           home,
		Home:          home,
		ProjectDir:    home,
		ExpectVersion: version,
		Git:           func([]string) GitResult { return GitResult{Status: 0, Stdout: "abc1234\n"} },
		Exec:          exec,
		Now:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, runDir
}

func reports(version string) ExecFunc {
	return func(string, []string) ExecResult {
		return ExecResult{Status: 0, Stdout: "feov-record version " + version + "\n"}
	}
}

// DEFECT 1: the preflight must refuse WITHOUT anyone opting in. No --bin-dir is passed here,
// which is exactly the documented launch. Before the fix this returned 0 and built the run.
func TestSkewedBinaryRefusesEvenWithoutBinDir(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.34.0"))
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Errorf("exit %d, want 2 — a skewed binary must refuse on the DOCUMENTED launch, which passes no --bin-dir", code)
	}
	if !strings.Contains(errb.String(), "RECORD BINARY PREFLIGHT FAILED") {
		t.Errorf("the refusal did not name itself:\n%s", errb.String())
	}
	// The run must not exist: refusing after building it is the same failure with extra steps.
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("the run directory was created despite the refusal — the gate is supposed to fire BEFORE any run state")
	}
	if strings.Contains(out.String(), "NOT AVAILABLE") {
		t.Error("the warn-and-proceed path is still reachable")
	}
}

// DEFECT 2: the expectation must come from the manifest ON DISK, not from the constant
// compiled into whichever binary happens to run `setup`.
//
// In the documented flow that binary IS the one the seats will call, so comparing its own
// constant against its own --version passes by construction. That exact failure already
// shipped once — both sides read 0.1.0 for the whole of 2026-07-19 while the event schema
// changed underneath (see cli/versionsync_test.go). Arming a gate that compares a number to
// itself would have closed the issue and fixed nothing.
func TestTheExpectedVersionComesFromTheManifestNotTheRunningBinary(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// The plugin this binary belongs to expects 9.9.9 — the cache was updated, the binary was not.
	if err := os.WriteFile(filepath.Join(root, "requirements.json"),
		[]byte(`{"recordToolVersion":"9.9.9"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Compiled constant AND reported version agree at 0.35.0 — the tautology that used to pass.
	cfg, _ := runCfg(t, "0.35.0", reports("0.35.0"))
	cfg.BinDir = binDir
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Errorf("exit %d, want 2 — the manifest beside the binary says 9.9.9 and the binary reports 0.35.0", code)
	}
	if !strings.Contains(errb.String(), "9.9.9") {
		t.Errorf("the refusal did not name the manifest's expectation, so an operator cannot tell what to rebuild:\n%s", errb.String())
	}
}

// A binary outside any plugin layout has no manifest to disagree with, so the compiled-in
// constant remains the fallback — a dev build must not be refused for lacking a file.
func TestNoManifestFallsBackToTheCompiledConstant(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	cfg.BinDir = binDir
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — no manifest means nothing to disagree with:\n%s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(runDir, "records")); err != nil {
		t.Errorf("the run was not built on the passing path: %v", err)
	}
	if !strings.Contains(out.String(), "0.35.0") {
		t.Errorf("the summary did not report the verified version:\n%s", out.String())
	}
}

// THE OTHER THREE FAIL-FAST GATES.
//
// The preflight was not specially unlucky: `Run` has four gates and NONE of them had a
// decision-level test, which is why one of them could sit disarmed through six runs. Each of
// these guards a promise `commands/research.md` makes to an operator in writing, so each one
// silently not firing is a documented guarantee that is simply untrue.

func TestMissingModelTierRefuses(t *testing.T) {
	for _, drop := range []string{"model", "judgment"} {
		t.Run(drop, func(t *testing.T) {
			cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
			if drop == "model" {
				cfg.Model = ""
			} else {
				cfg.JudgmentModel = ""
			}
			var out, errb bytes.Buffer
			if code := Run(cfg, &out, &errb); code != 2 {
				t.Errorf("exit %d, want 2 — the engine never guesses a tier or inherits the session model", code)
			}
			if !strings.Contains(errb.String(), "MODEL TIERS REQUIRED") {
				t.Errorf("the refusal did not name itself:\n%s", errb.String())
			}
			if _, err := os.Stat(runDir); !os.IsNotExist(err) {
				t.Error("the run directory was created despite a missing tier")
			}
		})
	}
}

// A cite that does not exist at its pin refuses. Evidence that can still change underneath
// the run is not evidence, so this must fail BEFORE the run exists.
func TestPinMissRefuses(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	cfg.Cites = []string{"docs/gone.md@deadbeef"}
	// rev-parse resolves (so pins are CHECKED, not skipped); cat-file -e fails (so the cite misses).
	cfg.Git = func(args []string) GitResult {
		if len(args) > 0 && args[0] == "cat-file" {
			return GitResult{Status: 1}
		}
		return GitResult{Status: 0, Stdout: "abc1234\n"}
	}
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Errorf("exit %d, want 2 — a cite that does not exist at its pin is not evidence", code)
	}
	if !strings.Contains(errb.String(), "PIN VALIDATION FAILED") {
		t.Errorf("the refusal did not name itself:\n%s", errb.String())
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("the run directory was created despite an unresolvable cite")
	}
}

// The run directory argument itself. A flag in its place means the operator forgot it, and
// building a run called "--topic" is worse than refusing.
func TestMissingRunDirRefuses(t *testing.T) {
	for _, arg := range []string{"", "--topic"} {
		cfg, _ := runCfg(t, "0.35.0", reports("0.35.0"))
		cfg.RunDir = arg
		var out, errb bytes.Buffer
		if code := Run(cfg, &out, &errb); code != 1 {
			t.Errorf("RunDir %q: exit %d, want 1", arg, code)
		}
		if !strings.Contains(errb.String(), "usage:") {
			t.Errorf("RunDir %q: no usage line:\n%s", arg, errb.String())
		}
	}
}

// An unrunnable binary refuses too — previously this was the "NOT AVAILABLE" warning that
// created the run regardless.
func TestUnrunnableBinaryRefuses(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", func(string, []string) ExecResult {
		return ExecResult{Status: 1, Err: os.ErrNotExist}
	})
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Errorf("exit %d, want 2", code)
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("the run directory was created despite an unrunnable record binary")
	}
}

// A MOSTLY-UNCLASSIFIED CORPUS REFUSES THE RUN.
//
// Delivery is class-indexed, so an unclassified pattern reaches no seat. The count was already
// printed — "(59 UNCLASSIFIED, not delivered)" — and it read as a nag in a long summary rather
// than as "red is starting blind". The 2026-08-05 run shipped with 0 classes delivered.
func TestMostlyUnclassifiedCorpusRefuses(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	mem := filepath.Join(t.TempDir(), "patterns")
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	// One classified, three not — the shape a raw accrual has.
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(mem, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("good.md", "---\nclasses: [figure-recount-fails]\ndescription: a hook\n---\n# t\n")
	for _, n := range []string{"a.md", "b.md", "c.md"} {
		write(n, "# untitled pattern\nprose with no frontmatter\n")
	}
	cfg.MemoryDir = mem

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d, want 2 — a corpus that mostly cannot be delivered must stop the run", code)
	}
	if !strings.Contains(errb.String(), "MOSTLY UNCLASSIFIED") {
		t.Errorf("the refusal did not name itself:\n%s", errb.String())
	}
	// The sources must be listed: the measured cause was the WRONG SOURCES, not bad authoring.
	if !strings.Contains(errb.String(), mem) {
		t.Errorf("the refusal must list the directories it read, so a composition bug is visible:\n%s", errb.String())
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Error("the run was created despite the refusal")
	}
}

// --memory-dir ADDS a source. It used to REPLACE, and research.md documents passing it as the
// remedy when gap-patterns reports "no memory dir" — so following the documented advice
// discarded the curated corpus (57 files, 55 classified) for the raw accrual (60, 1).
func TestMemoryDirAddsRatherThanReplaces(t *testing.T) {
	cfg, _ := runCfg(t, "0.35.0", reports("0.35.0"))
	promoted := filepath.Join(cfg.Cwd, "feov-memory", "red-gap-patterns")
	if err := os.MkdirAll(promoted, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"p1.md", "p2.md", "p3.md"} {
		if err := os.WriteFile(filepath.Join(promoted, n),
			[]byte("---\nclasses: [figure-recount-fails]\ndescription: h\n---\n# t\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// An extra source with one unclassified file: alone it would be mostly-unclassified and
	// refuse; combined with the promoted corpus it is a minority and the run proceeds.
	extra := filepath.Join(t.TempDir(), "extra")
	if err := os.MkdirAll(extra, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extra, "raw.md"), []byte("# unclassified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg.MemoryDir = extra

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — --memory-dir must ADD to the promoted corpus, not replace it:\n%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "gap-patterns") {
		t.Errorf("the summary did not report the corpus:\n%s", out.String())
	}
}

// THE REGISTRY IS STAGED, so `--class` means something (#299).
//
// Nothing wrote records/class-registry.json until this. loadRegistry always returned nil,
// validateClass always took its advisory branch, and --class accepted any string on every run
// there has ever been. The cost was not tidiness: gap-pattern delivery is CLASS-INDEXED, red
// invented a fresh vocabulary each run, and both record-era runs had ZERO overlap between the
// classes minted and the classes the memory corpus is indexed by.
func TestClassRegistryIsStagedIntoTheRun(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	mem := filepath.Join(cfg.Cwd, "feov-memory")
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, "class-registry.json"),
		[]byte(`{"classes":[{"slug":"false-universal"},{"slug":"self-attestation"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	staged := filepath.Join(runDir, "records", "class-registry.json")
	if _, err := os.Stat(staged); err != nil {
		t.Fatalf("the registry was not staged where loadRegistry reads it: %v", err)
	}
	if !strings.Contains(out.String(), "2 class(es) staged") {
		t.Errorf("the summary must report the registry, since it decides whether --class validates:\n%s", out.String())
	}
}

// An ABSENT registry is reported LOUDLY rather than degrading in silence. The advisory branch
// was the pattern this suite keeps finding: a gate that reports nothing and passes everything
// when its input is missing.
//
// WHAT THE LINE HAS TO SAY CHANGED WITH THE BEHAVIOUR. It used to warn that `--class` would
// accept ANY string — the honest description of a run that had lost its registry and would mint
// happily against a vocabulary nothing recognised. There is no advisory branch now: an absent
// registry REFUSES every mint, so the operator reading this line is looking at a run that cannot
// do its work, not one that will quietly do it wrong. The assertion moved with it, because a test
// pinning the old sentence would have made removing the tolerance look like a regression.
func TestAbsentRegistryIsAnnounced(t *testing.T) {
	cfg, _ := runCfg(t, "0.35.0", reports("0.35.0"))
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "NOT STAGED") || !strings.Contains(out.String(), "REFUSED") {
		t.Errorf("an absent registry must say that every mint will be refused:\n%s", out.String())
	}
	if strings.Contains(out.String(), "ANY string") {
		t.Errorf("the summary still describes the advisory branch, which no longer exists — an operator "+
			"reading it would expect a run that mints loosely rather than one that cannot mint:\n%s", out.String())
	}
}

// THE JOIN'S HEALTH IS STATED. A corpus indexed by classes the registry does not contain
// reaches no seat however well it is composed — which is exactly what both record-era runs
// did, at zero overlap, while every summary line looked healthy.
func TestUnjoinablePatternClassesAreNamed(t *testing.T) {
	cfg, _ := runCfg(t, "0.35.0", reports("0.35.0"))
	mem := filepath.Join(cfg.Cwd, "feov-memory")
	patterns := filepath.Join(mem, "red-gap-patterns")
	if err := os.MkdirAll(patterns, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mem, "class-registry.json"),
		[]byte(`{"classes":[{"slug":"false-universal"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, class := range map[string]string{"ok.md": "false-universal", "orphan.md": "invented-last-run"} {
		if err := os.WriteFile(filepath.Join(patterns, name),
			[]byte("---\nclasses: ["+class+"]\ndescription: h\n---\n# t\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "1 of 2 indexed class(es) exist in the registry") {
		t.Errorf("the join ratio must be reported:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "invented-last-run") {
		t.Errorf("an unjoinable class must be NAMED, or the corpus looks healthy while reaching nobody:\n%s", out.String())
	}
}

// A SECOND RUN OPENS, AND THE FIRST ONE IS NAMED. Driven end to end (#529).
//
// This gate used to REFUSE, for a reason it stated itself: the stale marker was what
// seat.InferRunDir handed to every verb invoked without --run, so a run left open misdirected
// the next one's seats. #526 removed that — setup bakes the run into <runDir>/.bin/feov-record,
// so a seat carries its run in its own environment and never has to ask, and ReadRunLiveMarker
// now declines to answer at all when more than one run is open.
//
// With the misdirection gone, refusing only forced the operator to defeat the guard by hand;
// measured twice on consecutive days, and both times the remedy was `rm` on the one file that
// says a run is open. What survives is the half that was always worth keeping: telling a human,
// at the one moment one is present, that a previous run was never captured.
func TestSetupOpensASecondRunAndNamesTheUnclosedOne(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-08-01_abandoned", nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "")

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — a run left open no longer stops the next one:\n%s", code, errb.String())
	}
	msg := out.String()
	for _, want := range []string{
		"ALSO OPEN",
		"research/2026-08-01_abandoned", // names the other run
		"2026-08-01T09:00:00.000Z",      // and when it started
		"feov-record capture",           // with the remedy that KEEPS its record
		"resolves",                      // and what an unflagged verb now does instead of guessing
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the notice must carry %q so the operator can act on it:\n%s", want, msg)
		}
	}
	// THE RUN IS ACTUALLY BUILT. A notice that quietly still refused would be the worst of both.
	if _, err := os.Stat(filepath.Join(runDir, "records")); err != nil {
		t.Errorf("the second run was not built: %v", err)
	}
	// BOTH ARE LIVE, and the abandoned one is still capturable — its row must not be displaced.
	live := runlive.ReadRunLive(cfg.Cwd)
	if len(live) != 2 {
		t.Fatalf("the marker holds %d run(s), want both: %+v", len(live), live)
	}
	var names []string
	for _, m := range live {
		names = append(names, m.RunDir)
	}
	if !strings.Contains(strings.Join(names, " "), "research/2026-08-01_abandoned") {
		t.Errorf("opening a second run displaced the first one's row: %v", names)
	}

	// AND NEITHER IS INFERRED. Two open runs means there is no single answer, so the thing that
	// hands a run to an unflagged verb must decline rather than pick one.
	if m, ok := runlive.ReadRunLiveMarker(cfg.Cwd); ok {
		t.Errorf("with two runs open, inference still resolved one (%s) — that is the guess this replaces", m.RunDir)
	}
}

// The same run is not a second run. Re-running setup on a run in progress is ordinary, and this
// gate must not be the thing that breaks it.
func TestSetupStaysIdempotentForTheRunItsMarkerNames(t *testing.T) {
	cfg, runDir := runCfg(t, "0.35.0", reports("0.35.0"))
	runlive.WriteRunLiveMarker(cfg.Cwd, runDir, nil, cfg.Now, "", "")

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — setup on the run its own marker names must proceed:\n%s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(runDir, "records")); err != nil {
		t.Errorf("the run was not built: %v", err)
	}
}
