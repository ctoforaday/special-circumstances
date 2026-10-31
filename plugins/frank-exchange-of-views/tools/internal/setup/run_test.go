package setup

import (
	"bytes"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
// gate green — while `skills/research/SKILL.md` promised the opposite in writing.

// runCfg is a Config with the environment injected, so no test touches real git or a real
// binary. Cwd/Home point at t.TempDir() so mirrors and markers land in the sandbox.
func runCfg(t *testing.T, exec ExecFunc) (Config, string) {
	t.Helper()
	home := t.TempDir()
	runDir := filepath.Join(home, "research", "2026-01-01_probe")

	// A PLUGIN-SHAPED LAYOUT, because that is the only shape setup runs in: the epoch comes
	// from the requirements.json one directory above the binary, and a fixture without one is
	// refused. Without BinDir the binary would resolve to the TEST binary's directory — a
	// go-build scratch dir with no plugin in it — which is exactly the configuration the old
	// compiled-in fallback let pass by comparing a number with itself.
	//
	// The manifest carries the epoch this binary compiles in, so the fixture tracks the
	// binary instead of pinning a number somebody would have to move by hand.
	binDir := filepath.Join(home, "plugin", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf(`{"eventSchema":%d,"tools":[{"name":"claude","min_version":"2.1.305"}]}`+"\n", record.EventSchema)
	if err := os.WriteFile(filepath.Join(home, "plugin", "requirements.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	return Config{
		BinDir:        binDir,
		RunDir:        runDir,
		Topic:         "a probe",
		Model:         "haiku",
		JudgmentModel: "haiku",
		Session:       testLead,
		Cwd:           home,
		Home:          home,
		Git:           func([]string) GitResult { return GitResult{Status: 0, Stdout: "abc1234\n"} },
		Exec:          exec,
		Now:           time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, runDir
}

// skewedEpoch is an epoch this binary CANNOT be writing, derived so it stays skewed.
//
// It was the literal "9", which stopped being a skew the day the schema reached 9: the preflight
// agreed with the fixture, refused nothing, and three assertions about a refusal failed at once.
// A test that hardcodes the value it is supposed to differ from has a shelf life.
var skewedEpoch = strconv.Itoa(record.EventSchema + 1)

func reports(epoch string) ExecFunc {
	return func(name string, _ []string) ExecResult {
		if name == clientBin {
			return ExecResult{Status: 0, Stdout: "2.1.309 (Claude Code)\n"}
		}
		// --schema answers with the epoch alone; the caller parses a number, not a sentence.
		return ExecResult{Status: 0, Stdout: epoch + "\n"}
	}
}

// DEFECT 1: the preflight refuses WITHOUT anyone opting in. Arming it used to depend on
// --bin-dir, a flag nothing in the plugin passed, so the documented launch returned 0 and
// built the run against a skewed binary.
//
// The skew is now an EPOCH difference rather than a version difference: this binary writes a
// shape the plugin beside it does not read, which is the only incompatibility that means
// anything when nothing promises backwards compatibility.
func TestSkewedBinaryRefusesEvenWithoutBinDirOptIn(t *testing.T) {
	cfg, runDir := runCfg(t, reports(skewedEpoch))
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
	// The plugin beside this binary reads schema 9 — the cache moved, the binary did not.
	if err := os.WriteFile(filepath.Join(root, "requirements.json"),
		[]byte(`{"eventSchema":9}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The binary reports 1 while the manifest reads 9. Under the old shape the expectation
	// could fall back to the binary's own number, which made this a tautology.
	cfg, _ := runCfg(t, reports(skewedEpoch))
	cfg.BinDir = binDir
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Errorf("exit %d, want 2 — the manifest beside the binary reads schema 9 and the binary writes 1", code)
	}
	if !strings.Contains(errb.String(), "9") {
		t.Errorf("the refusal did not name the manifest's expectation, so an operator cannot tell what to rebuild:\n%s", errb.String())
	}
}

// NO MANIFEST IS A REFUSAL, NOT A DEFAULT.
//
// The old shape fell back to the binary's own compiled-in number here, so a binary with no
// plugin beside it checked itself against itself and passed by construction. That is the
// plausible zero in its purest form: the check reports success precisely when it cannot be
// made.
//
// A dev build is not refused by this in practice — in a checkout the binary sits under
// plugins/<name>/tools, one level below the requirements.json it needs, so it finds it. What
// gets refused is a binary somewhere with no plugin at all, where there is genuinely nothing
// to check against and no honest way to say "checked".
func TestNoManifestIsRefusedRatherThanCheckedAgainstItself(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, _ := runCfg(t, reports(skewedEpoch))
	cfg.BinDir = binDir
	var out, errb bytes.Buffer

	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d, want 2 — with no manifest the epoch cannot be checked, and an "+
			"unmakeable check must not pass:\n%s", code, errb.String())
	}
	if !strings.Contains(errb.String(), "requirements.json") {
		t.Errorf("the refusal must name the file it could not read:\n%s", errb.String())
	}
}

// THE OTHER THREE FAIL-FAST GATES.
//
// The preflight was not specially unlucky: `Run` has four gates and NONE of them had a
// decision-level test, which is why one of them could sit disarmed through six runs. Each of
// these guards a promise `skills/research/SKILL.md` makes to an operator in writing, so each one
// silently not firing is a documented guarantee that is simply untrue.

func TestMissingModelTierRefuses(t *testing.T) {
	for _, drop := range []string{"model", "judgment"} {
		t.Run(drop, func(t *testing.T) {
			cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
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
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
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
		cfg, _ := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
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
	cfg, runDir := runCfg(t, func(string, []string) ExecResult {
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

// SETUP READS NOTHING IN feov-memory/ (plans/gap-class-enum.md §III.7). A gap's class is a value of
// the closed GapClass set the binary owns, so a consuming project with no feov-memory/ sets up
// exactly as one holding a stale or garbage class-registry.json: the same stdout, the same stderr,
// the same records, once each fixture's own root is masked.
func TestSetupReadsNoFeovMemory(t *testing.T) {
	type result struct {
		out, errb string
		records   []string
	}
	setup := func(t *testing.T, stage func(cwd string)) result {
		t.Helper()
		cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
		stage(cfg.Cwd)
		var out, errb bytes.Buffer
		if code := Run(cfg, &out, &errb); code != 0 {
			t.Fatalf("exit %d: %s", code, errb.String())
		}
		var names []string
		recs := filepath.Join(runDir, "records")
		if err := filepath.WalkDir(recs, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(recs, p)
			names = append(names, filepath.ToSlash(rel))
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		sort.Strings(names)
		mask := func(s string) string { return strings.ReplaceAll(s, cfg.Cwd, "<cwd>") }
		return result{mask(out.String()), mask(errb.String()), names}
	}
	bare := setup(t, func(string) {})
	garbage := setup(t, func(cwd string) {
		mem := filepath.Join(cwd, "feov-memory")
		if err := os.MkdirAll(mem, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(mem, "class-registry.json"), []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
	})
	if bare.out != garbage.out {
		t.Errorf("setup's stdout depends on feov-memory/:\n--- none ---\n%s\n--- garbage registry ---\n%s", bare.out, garbage.out)
	}
	if bare.errb != garbage.errb {
		t.Errorf("setup's stderr depends on feov-memory/:\n--- none ---\n%s\n--- garbage registry ---\n%s", bare.errb, garbage.errb)
	}
	if strings.Join(bare.records, "\n") != strings.Join(garbage.records, "\n") {
		t.Errorf("setup's records depend on feov-memory/:\n none: %v\n garbage: %v", bare.records, garbage.records)
	}
	if strings.Contains(bare.out, "class registry") {
		t.Errorf("setup prints a class-registry line:\n%s", bare.out)
	}
}

// A SECOND LIVE RUN FOR ONE SESSION IS REFUSED, NAMING THE FIRST AND HOW TO CLOSE IT (#1255). Driven
// end to end.
//
// The wrapper carries a seat's run into its own calls, but the hooks find a seat's run through this
// project's marker, among the runs its session leads, and with two of those they bracket no sitting
// and count no call for either. The live run here is the same session's; another session's is
// TestSetupRefusesBesideAnotherSessionsLiveRun.
// Setup used to open the second run with a notice; every seat of both runs then sat unrecorded.
func TestSetupRefusesASecondLiveRun(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	other := filepath.Join(cfg.Cwd, "research", "2026-08-01_running")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-08-01_running", nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "", testLead); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d, want 2 — a second live run must be refused:\n%s%s", code, out.String(), errb.String())
	}
	for _, want := range []string{
		"ANOTHER RUN IS LIVE IN THIS PROJECT",
		"research/2026-08-01_running",                     // names the live run
		"2026-08-01T09:00:00.000Z",                        // and when it started
		"feov-record capture research/2026-08-01_running", // how to close it
		"another checkout",                                // or how to run beside it
	} {
		if !strings.Contains(errb.String(), want) {
			t.Errorf("the refusal must carry %q:\n%s", want, errb.String())
		}
	}
	// NOTHING IS BUILT, AND THE LIVE RUN'S ROW STANDS.
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Errorf("the refused run was built: %v", err)
	}
	if rows := runlive.ReadRunLive(cfg.Cwd); len(rows) != 1 || rows[0].RunDir != "research/2026-08-01_running" {
		t.Errorf("the marker changed under a refusal: %+v", rows)
	}
}

// ROWS THAT ARE NO LONGER LIVE ARE PRUNED, AND NAMED, AND DO NOT STOP THE RUN. A run that was
// captured and a run whose directory is gone are residue nothing else will remove.
func TestSetupPrunesRowsThatAreNotLive(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	ended := filepath.Join(cfg.Cwd, "research", "2026-08-01_captured")
	if err := os.MkdirAll(ended, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ended, runlive.OutcomeMark), []byte(`{"agentId":"a-bench"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ended, runlive.CapturedMark), []byte("2026-08-01T10:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-08-01_captured", nil, at, "", "", testLead); err != nil {
		t.Fatal(err)
	}
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, "research/2026-07-01_deleted", nil, at, "", "", testLead); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — rows that are not live must not stop a run:\n%s", code, errb.String())
	}
	for _, want := range []string{
		"pruned the live-run marker's row for research/2026-08-01_captured",
		"it was captured",
		"pruned the live-run marker's row for research/2026-07-01_deleted",
		"its run directory is gone",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("setup must say what it pruned — missing %q:\n%s", want, out.String())
		}
	}
	if rows := runlive.ReadRunLive(cfg.Cwd); len(rows) != 1 || !runlive.SameRun(cfg.Cwd, rows[0].RunDir, runDir) {
		t.Errorf("the marker must hold the new run alone: %+v", rows)
	}
}

// A DECIDED RUN NOTHING CAPTURED REFUSES THE NEXT ONE (gblock, 2026-10-24: "Setup refuses B"),
// naming it and what closes it — capture when its assembly finished, a resume or a capture as it
// stands when it did not — and its row stands. Once captured, its row is pruned and setup goes on.
func TestSetupRefusesWhileADecidedRunIsNotCaptured(t *testing.T) {
	for _, assembled := range []bool{false, true} {
		t.Run(map[bool]string{false: "assembly unfinished", true: "assembly finished"}[assembled], func(t *testing.T) {
			cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
			const rel = "research/2026-08-01_decided"
			decided := filepath.Join(cfg.Cwd, rel)
			if err := os.MkdirAll(decided, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := runlive.WriteOutcomeMark(decided, "a-bench"); err != nil {
				t.Fatal(err)
			}
			if assembled {
				if err := os.WriteFile(filepath.Join(decided, report.FileRun), []byte("# run\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, rel, nil, time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC), "", "", testLead); err != nil {
				t.Fatal(err)
			}

			var out, errb bytes.Buffer
			if code := Run(cfg, &out, &errb); code != 2 {
				t.Fatalf("exit %d, want 2 — a decided run nothing captured must refuse the next:\n%s%s", code, out.String(), errb.String())
			}
			wants := []string{"HAS ITS OUTCOME AND WAS NOT CAPTURED", rel, "2026-08-01T09:00:00.000Z", "feov-record capture " + rel}
			if assembled {
				wants = append(wants, "nothing to resume")
			} else {
				wants = append(wants, "never finished", "resume it", "capture it as it stands")
			}
			for _, want := range wants {
				if !strings.Contains(errb.String(), want) {
					t.Errorf("the refusal must carry %q:\n%s", want, errb.String())
				}
			}
			if assembled && strings.Contains(errb.String(), "resume it") {
				t.Errorf("a finished assembly is sent to resume:\n%s", errb.String())
			}
			if _, err := os.Stat(runDir); !os.IsNotExist(err) {
				t.Errorf("the refused run was built: %v", err)
			}
			if rows := runlive.ReadRunLive(cfg.Cwd); len(rows) != 1 || rows[0].RunDir != rel {
				t.Errorf("the decided run's row did not stand: %+v", rows)
			}

			// CAPTURED, it is residue: pruned, named, and no refusal.
			if err := os.WriteFile(filepath.Join(decided, runlive.CapturedMark), []byte("2026-08-01T10:00:00Z\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			out.Reset()
			errb.Reset()
			if code := Run(cfg, &out, &errb); code != 0 {
				t.Fatalf("exit %d once the decided run was captured, want 0:\n%s", code, errb.String())
			}
			if !strings.Contains(out.String(), "pruned the live-run marker's row for "+rel) {
				t.Errorf("the captured run's row was not pruned and named:\n%s", out.String())
			}
		})
	}
}

// The same run is not a second run. Re-running setup on a run in progress is ordinary, and this
// gate must not be the thing that breaks it.
func TestSetupStaysIdempotentForTheRunItsMarkerNames(t *testing.T) {
	cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	if _, err := runlive.WriteRunLiveMarker(cfg.Cwd, runDir, nil, cfg.Now, "", "", testLead); err != nil {
		t.Fatal(err)
	}

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("exit %d, want 0 — setup on the run its own marker names must proceed:\n%s", code, errb.String())
	}
	if _, err := os.Stat(filepath.Join(runDir, "records")); err != nil {
		t.Errorf("the run was not built: %v", err)
	}
}

// A MARKER SETUP COULD NOT WRITE FAILS SETUP (#1552). The write's error used to be dropped: setup
// reported the run ready while no hook could find it.
func TestSetupFailsWhenTheMarkerCannotBeWritten(t *testing.T) {
	cfg, _ := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
	// A non-empty directory at the marker's path: no platform replaces it with a file.
	if err := os.MkdirAll(filepath.Join(cfg.Cwd, ".claude", "run-live.json", "occupant"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 1 {
		t.Fatalf("exit %d, want 1:\n%s%s", code, out.String(), errb.String())
	}
	if !strings.Contains(errb.String(), "live-run marker could not be written") {
		t.Errorf("the failure must name the marker:\n%s", errb.String())
	}
}
