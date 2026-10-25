package setup

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claudedir"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookprovenance"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config is the parsed setup invocation. The environment fields (Cwd/Home/ExpectVersion)
// and the injectables (Git/Exec/Now) are resolved by the command from os;
// tests may supply them directly.
type Config struct {
	RunDir        string
	Topic         string
	Model         string
	JudgmentModel string
	Cites         []string
	Lanes         string
	// The run's terms (plans/roundless.md §III.B.2, §III.B.2.2): K, KMax, MintBudget and the
	// convergence fraction. Zero means "setup's default", which is written out so the file always
	// states what bound the run.
	K, KMax, MintBudget int
	ConvergenceFraction float64
	// MaxSittingCalls is the tool calls one seat may make in one sitting before the PreToolUse
	// hook refuses the rest (internal/sittingcap). Zero means sittingcap.DefaultMaxCalls.
	MaxSittingCalls int
	// MaxEpochs is the run's epoch limit; zero is setup's default unless MaxEpochsSet says the
	// operator passed it, and a value the operator passed below 1 is refused.
	MaxEpochs    int
	MaxEpochsSet bool
	BinDir       string
	// RunID and ScriptPath travel into the run-live marker so a STALE marker names how to
	// resume rather than only where something once ran. Optional: a launcher that does not
	// know them leaves them empty, and the marker omits the fields rather than carrying "".
	RunID      string
	ScriptPath string
	// AllowSubstitution is the operator's standing consent to a served model that is not the
	// configured tier. It is recorded on the run rather than passed per seat: the seat is the
	// party whose adversary strength is in question, so the decision cannot be its to make.
	AllowSubstitution bool
	// Session is the launching session's id (seatenv.Session), recorded on the run's live-run row
	// as its lead: the hooks attribute a payload to the run only when its session_id is this one.
	// Required — a run no session leads is a run none of whose seats can be attributed.
	Session string

	Cwd  string
	Home string
	Git  GitFunc
	Exec ExecFunc
	Now  time.Time
}

// Run reproduces the mjs main(): the fail-fast gates (runDir→1; model tiers, launching session,
// epoch limit, another live run, a decided run uncaptured, pin miss, record-binary preflight→2),
// then the run-dir build + mirrors + marker, then the summary. Returns the process exit code; writes to the given streams.
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
		fmt.Fprintln(stderr, `usage: feov-record setup <runDir> --topic "<topic>" --model <m> --judgment-model <m> [--cite <path>[@pin]]... [--bin-dir <dir>]`)
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
			fmt.Fprintln(stderr, "  - --judgment-model is unset (the JUDGMENT tier: blue-synthesize, red-chair, judge, assemble)")
		}
		fmt.Fprintln(stderr, "  the engine does not guess a tier or inherit the session model. Pass both, e.g.")
		fmt.Fprintln(stderr, "  --model sonnet --judgment-model sonnet   (a smoke run passes --model haiku --judgment-model haiku)")
		return 2
	}

	// Gate: a launching session, before the run dir exists. The run's members are the agents of the
	// session that leads it, and the hooks tell them by the session_id on each payload; a run with no
	// lead would have none.
	if strings.TrimSpace(cfg.Session) == "" {
		fmt.Fprintf(stderr, "run-setup: NO LAUNCHING SESSION — refusing to create the run: %s is unset.\n", seatenv.SessionVar)
		fmt.Fprintln(stderr, "  The hooks attribute an agent to a run only when its payload's session_id is the session that")
		fmt.Fprintln(stderr, "  leads the run, and setup records that session from the client's environment. Run setup from")
		fmt.Fprintln(stderr, "  the Bash tool of the Claude Code session that will launch the run.")
		return 2
	}

	// Gate: the epoch limit is 1 or more, before the run dir exists. A limit below 1 would end the
	// run at its first chair sitting with nothing audited, and record.RunParams refuses to read it.
	if cfg.MaxEpochs < 0 || (cfg.MaxEpochsSet && cfg.MaxEpochs < 1) {
		fmt.Fprintf(stderr, "run-setup: --max-epochs %d refused — the run's epoch limit is 1 or more (default %d); refusing to create the run\n", cfg.MaxEpochs, record.DefaultParams.MaxEpochs)
		return 2
	}

	// ONE LIVE RUN PER PROJECT (#1255), WHATEVER SESSION LEADS IT. Rows that are no longer live are
	// pruned and named first; another LIVE run refuses this one, before any run state exists.
	//
	// The wrapper (#526) carries a seat's run into its own calls, but the HOOKS cannot use it: a
	// hook's payload names no run, so they find one through this project's marker, among the rows its
	// session leads — and with two of those live they bracket no sitting, count no call and inject no
	// run for either. Every seat of both runs then acts unbracketed and uncapped.
	//
	// Two runs led by DIFFERENT sessions are told apart by the hooks, and are still refused: they share
	// the project's other state. Two setups of one research/<date>_<slug> both pass this unlocked check
	// and build into one directory, and capture read-modify-writes feov-memory/*-scorecard.md and law/
	// with no lock. Two runs at once need two checkouts, each with its own marker and memory.
	//
	// A DECIDED RUN THAT WAS NEVER CAPTURED REFUSES THIS ONE TOO (gblock, 2026-10-24: "Setup refuses
	// B"), before the prune would take its row. Its outcome is recorded and nothing archived it:
	// pruning its row would leave the one record of it in a run directory the container does not
	// keep, and its assembly sitting — resumed — binds to the run through that row.
	if code := refuseADecidedRun(cfg, stderr); code != 0 {
		return code
	}
	pruned, perr := runlive.PruneRunLive(cfg.Cwd)
	if perr != nil {
		// The rows judged dead are still in the file, and a dead row the refusal below would read as
		// another live run is exactly what a prune exists to clear.
		fmt.Fprintln(stderr, "run-setup: COULD NOT PRUNE THE LIVE-RUN MARKER — refusing to create the run:")
		for _, p := range pruned {
			fmt.Fprintf(stderr, "  no longer live, still in the marker: %s (started %s) — %s\n", p.Row.RunDir, orUnknown(p.Row.Started), p.State.Why())
		}
		fmt.Fprintf(stderr, "  error: %v\n", perr)
		fmt.Fprintln(stderr, "  remedy: make .claude/run-live.json writable (or remove those rows by hand), then re-run setup")
		return 1
	}
	for _, p := range pruned {
		fmt.Fprintf(stdout, "run-setup: pruned the live-run marker's row for %s (started %s) — %s\n", p.Row.RunDir, orUnknown(p.Row.Started), p.State.Why())
	}
	lead := runlive.Caller{Session: cfg.Session}
	for _, m := range runlive.ReadRunLive(cfg.Cwd) {
		same := runlive.SameRun(cfg.Cwd, m.RunDir, cfg.RunDir)
		if same && lead.Leads(m) {
			continue // a re-setup is its own lead session's
		}
		if same {
			// THIS run's row, led by another session, refuses a re-setup here: setup would otherwise
			// move the run's lead session silently, which is a take-over without its refusal. The usual
			// caller is a second session launching a run of the same name, so the new checkout leads;
			// the take-over is for a resume, and only once the run's lead session has ended.
			fmt.Fprintln(stderr, "run-setup: THIS RUN IS LED BY ANOTHER SESSION — refusing to set it up again:")
			fmt.Fprintf(stderr, "  run: %s, led by session %s; this session is %s\n", m.RunDir, orUnknown(m.LeadSession), cfg.Session)
			fmt.Fprintln(stderr, "  remedies:")
			fmt.Fprintln(stderr, "    - start the new run from another checkout (`git worktree add`), which is its own project")
			fmt.Fprintln(stderr, "    - only if this session is RESUMING that run and its lead session has ended, take it over:")
			fmt.Fprintf(stderr, "        feov-record take-over %s --from-session %s --seat-id operator\n", m.RunDir, runlive.FromSessionArg(m.LeadSession))
			return 2
		}
		fmt.Fprintln(stderr, "run-setup: ANOTHER RUN IS LIVE IN THIS PROJECT — refusing to create the run:")
		fmt.Fprintf(stderr, "  live: %s (started %s), led by session %s\n", m.RunDir, orUnknown(m.Started), orUnknown(m.LeadSession))
		fmt.Fprintln(stderr, "  A project holds one live run: its seats find the run through this project's live-run marker,")
		fmt.Fprintln(stderr, "  and its capture writes this project's memory.")
		fmt.Fprintln(stderr, "  remedies:")
		fmt.Fprintln(stderr, "    - start this run from another checkout (`git worktree add`), which is its own project")
		fmt.Fprintf(stderr, "    - or finish that run: resume it to an outcome, or abandon it — then close it with\n")
		fmt.Fprintf(stderr, "        feov-record capture %s <its transcript dir> --seat-id operator\n", m.RunDir)
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
	// nothing. So the guarantee `skills/research/SKILL.md` states ("setup preflights the binary's
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
		fmt.Fprintln(stderr, "  (failing here costs a re-run; failing mid-run costs a seat its whole record)")
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
	// THE CAST IS WRITTEN FIRST, under the harness, before any seat registers: the run's admissible
	// seats, derived from the areas selected and the lane count. Every register and every dispatch
	// is checked against it from here on (plans/roundless.md §III.B.1).
	lanesN, _ := strconv.Atoi(cfg.Lanes)
	// THE LANES ARE NAMED, NOT LEFT TO BE RECOGNISED. Every other seat's role and tier come from the
	// static roster; a lane's cannot, because how many a run seats is a run parameter. setup is the
	// one process that knows, so it says so here rather than leaving a pattern to guess later.
	seats, laneSeats := record.CastFor(lanesN)
	if _, err := record.Append(record.Identity{Run: run, SeatID: record.HarnessSeat},
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats}); err != nil {
		fmt.Fprintf(stderr, "run-setup: could not write the cast: %v\n", err)
		return 2
	}
	// Setup holds no record handle past this write: the seats open their own.
	_ = recordsql.Close(filepath.Join(run.Dir(), "records", "record.db"))

	BuildSkeleton(run)
	if mirrorRoot, mErr := record.MirrorRoot(); mErr != nil {
		// LOUD, not folded into the zero: a purge that could not resolve its own directory has
		// not checked anything, and reporting that as "0 removed" is the same line a clean
		// board prints.
		fmt.Fprintf(stderr, "  mirror purge: NOT RUN — %v\n", mErr)
	} else if n := record.PurgeStaleMirrors(mirrorRoot, cfg.Now, 30); n > 0 {
		fmt.Fprintf(stdout, "  mirror purge: %d stale checkpoint mirror(s) removed\n", n)
	}
	pinned := BuildPinned(run, head, cfg.Cites)

	terms := record.DefaultParams
	if cfg.K > 0 {
		terms.K = cfg.K
	}
	if cfg.KMax > 0 {
		terms.KMax = cfg.KMax
	}
	if cfg.MintBudget > 0 {
		terms.MintBudget = cfg.MintBudget
	}
	if cfg.ConvergenceFraction > 0 {
		terms.ConvergenceFraction = cfg.ConvergenceFraction
	}
	if cfg.MaxEpochs > 0 {
		terms.MaxEpochs = cfg.MaxEpochs
	}
	maxCalls := sittingcap.DefaultMaxCalls
	if cfg.MaxSittingCalls > 0 {
		maxCalls = cfg.MaxSittingCalls
	}
	rc := runConfig{Topic: topic, RunDir: run.Dir(), Model: cfg.Model, JudgmentModel: cfg.JudgmentModel, Lanes: ptrOrNil(cfg.Lanes), EventSchema: expect, AllowModelSubstitution: cfg.AllowSubstitution,
		K: terms.K, KMax: terms.KMax, MintBudget: terms.MintBudget, ConvergenceFraction: terms.ConvergenceFraction, MaxEpochs: terms.MaxEpochs,
		MaxSittingCalls: maxCalls,
		Hooks:           hookprovenance.Observe(claudedir.PluginCache(os.Getenv, homeDir()), "frank-exchange-of-views"),
		Billing:         ReadBillingIdentity()}
	if b, err := marshalJSON(rc); err == nil {
		os.WriteFile(filepath.Join(run.Dir(), "inputs", "run-config.json"), b, 0o644)
	}
	reportBilling(rc.Billing, stdout, stderr)

	law := MirrorLaw(filepath.Join(cfg.Cwd, "law"), run)
	pinnedPaths := []string{}
	for _, c := range cfg.Cites {
		p, _ := splitPin(c)
		pinnedPaths = append(pinnedPaths, p)
	}
	marker, merr := runlive.WriteRunLiveMarker(cfg.Cwd, cfg.RunDir, pinnedPaths, cfg.Now, cfg.RunID, cfg.ScriptPath, cfg.Session)
	if merr != nil {
		// The run directory exists but the run is NOT live: the hooks find a seat's run only through
		// this marker, so a run launched now would bracket no sitting and count no call.
		fmt.Fprintf(stderr, "run-setup: the run directory %s is built, but the live-run marker could not be written: %v\n", cfg.RunDir, merr)
		fmt.Fprintln(stderr, "  The hooks cannot find this run until it is, so do not launch it.")
		fmt.Fprintln(stderr, "  remedy: make .claude/ writable, then re-run setup with the same arguments — setup is idempotent for its own run")
		return 1
	}
	// Written AFTER the marker deliberately: the wrapper is the carrier that does not move when
	// the marker does, and writing it second makes the ordering obvious to anyone reading for
	// which of the two is derived from the other. Neither is — that is the point.
	wrapper := WriteRunWrapper(run, recordBin)

	fmt.Fprintf(stdout, "run-setup: %s\n", cfg.RunDir)
	fmt.Fprintln(stdout, "  skeleton: directories only — blue/report.md is the synthesizer's to write, report.md the assembler's")
	// "red-merge-born" was true when the chair wrote those files. It stopped being true, and the
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
	if law.Written {
		fmt.Fprintf(stdout, "  law: %d file(s) mirrored (statute > precedent > argument)\n", law.Files)
	} else {
		fmt.Fprintf(stdout, "  law: %s\n", law.Reason)
	}
	// NO SCORECARDS ARE STAGED INTO THE RUN, because nothing read them.
	//
	// Three things went together here. inputs/scorecards.json existed so the skill could tell the
	// lead to paste its parsed contents into the Workflow args, and the engine destructured that
	// arg without ever reading it. The headline extraction that filled it computed a value with no
	// consumer. And the per-side <card>-scorecard.md copies mirrored into inputs/ had no reader
	// either: no prompt names them, and a seat reads its OWN scorecard through `show scorecard`,
	// computed live from THIS run's record — a prior run's numbers are Goodhart bait.
	//
	// The corpus still lives where it is written: capture's WriteScorecards keeps each card in
	// feov-memory/, which is the durable home and the next run's input.
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
	Lanes         *string `json:"lanes"`
	// The run's terms, always written (record.Params reads them; see that type for what each bounds).
	K                   int     `json:"k"`
	KMax                int     `json:"kMax"`
	MintBudget          int     `json:"mintBudget"`
	ConvergenceFraction float64 `json:"convergenceFraction"`
	MaxEpochs           int     `json:"maxEpochs"`
	// MaxSittingCalls is read by the PreToolUse hook through sittingcap.Limit, not by
	// record.Params: the hook may not link the record. The key is sittingcap.ConfigKey.
	MaxSittingCalls int `json:"maxSittingCalls"`
	// Hooks is what was INSTALLED on the hook side when this run was set up (#751). The record
	// binary is baked from the working tree and the hooks come from the version-gated install
	// cache, so the two can be days apart — and a run whose hooks were stale is otherwise
	// byte-identical on the record to one whose hooks were current.
	//
	// It says what was installed, never what RAN: a cache entry on disk does not prove the
	// harness invoked it. The liveness half is #555's. The key is hookprovenance.ConfigKey, and
	// run.md reads it back through hookprovenance.Read.
	Hooks hookprovenance.Provenance `json:"hooks"`
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
	// Billing is WHAT THIS RUN IS BILLED TO — the account, its organization and whether it is a
	// subscription. See BillingIdentity for the measurement that put it here: a run had no way to
	// say which bill it was on, so an isolated config silently authenticated against a different
	// organization and spent credits for months while reporting a dollar figure that read as
	// telemetry.
	Billing BillingIdentity `json:"billing"`
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

// reportBilling says what the run will be billed to, BEFORE it runs.
//
// The failure this answers took thirty-five minutes and $7.09 to become visible, and then only as
// "Credit balance is too low" with the debate abandoned mid-epoch and no outcome row. Every fact
// needed to see it coming was on disk at setup; nothing read them aloud.
//
// A WARNING, NEVER A REFUSAL. An API-billed run is legitimate — CI, a second operator, a
// deliberately isolated environment — and refusing it would break those to protect a wallet the
// tool cannot see. What is not legitimate is doing it SILENTLY, so this makes the quiet case loud
// and leaves the decision where it belongs.
func reportBilling(b BillingIdentity, stdout, stderr io.Writer) {
	if b.Subscription() {
		where := "the default config"
		if b.Isolated {
			where = b.ConfigDir
		}
		fmt.Fprintf(stdout, "run-setup: billing to %s (%s), from %s\n",
			orUnknown(b.BillingType), orUnknown(b.SubscriptionType), where)
		return
	}
	fmt.Fprintln(stderr, "run-setup: THIS RUN IS NOT ON A SUBSCRIPTION.")
	if b.Isolated {
		fmt.Fprintf(stderr, "  The config directory is %s, which carries no subscription credential;\n", b.ConfigDir)
		fmt.Fprintln(stderr, "  the CLI will authenticate however else it can, which may be an API profile billed in CREDITS.")
	}
	if b.Unreadable != "" {
		fmt.Fprintf(stderr, "  %s\n", b.Unreadable)
	}
	fmt.Fprintln(stderr, "  A full run costs real money on that path. Point CLAUDE_CONFIG_DIR at a directory")
	fmt.Fprintln(stderr, "  holding your credential, or accept the charge deliberately.")
}

// orUnknown keeps an absent field visible rather than rendering a confident blank.
func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

// homeDir is the home the plugin cache resolves against when neither CLAUDE_CODE_PLUGIN_CACHE_DIR nor
// CLAUDE_CONFIG_DIR names it. It reads $HOME rather than os.UserHomeDir so the suite's own HOME
// sandbox (testbuild) redirects it, which is what makes this testable at all — and what stops a test
// reading the developer's real plugin cache and reporting a different answer on every machine.
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}

// refuseADecidedRun refuses setup while another run in the project has its outcome recorded and
// was not captured, naming it and what closes it: capture when its assembly finished, a resume of
// its assembly sitting when it did not (report.AssemblyWritten, which engine-stopped reads too).
// It returns the exit code, 0 when no row is such a run.
func refuseADecidedRun(cfg Config, stderr io.Writer) int {
	for _, m := range runlive.ReadRunLive(cfg.Cwd) {
		if runlive.SameRun(cfg.Cwd, m.RunDir, cfg.RunDir) || runlive.StateOf(cfg.Cwd, m, runlive.Caller{}) != runlive.RowDecided {
			continue
		}
		fmt.Fprintln(stderr, "run-setup: ANOTHER RUN IN THIS PROJECT HAS ITS OUTCOME AND WAS NOT CAPTURED — refusing to create the run:")
		fmt.Fprintf(stderr, "  decided: %s (started %s)\n", m.RunDir, orUnknown(m.Started))
		fmt.Fprintln(stderr, "  Its record is archived, and its documents folded, only by capture; its row is what binds its")
		fmt.Fprintln(stderr, "  assembly sitting to it when the run is resumed.")
		capture := fmt.Sprintf("feov-record capture %s <its transcript dir> --seat-id operator", m.RunDir)
		run, rerr := record.OpenRun(runlive.Resolve(cfg.Cwd, m.RunDir))
		written := false
		if rerr == nil {
			written, rerr = report.AssemblyWritten(run)
		}
		switch {
		case rerr != nil:
			fmt.Fprintf(stderr, "  whether its assembly finished could not be read (%v), so neither remedy is chosen for you:\n", rerr)
			fmt.Fprintln(stderr, "    - resume it, and its assembly sitting assembles; then capture it with")
			fmt.Fprintf(stderr, "        %s\n", capture)
			fmt.Fprintln(stderr, "    - or capture it as it stands, with the same command")
		case written:
			fmt.Fprintln(stderr, "  Its assembly finished, and there is nothing to resume: capture it with")
			fmt.Fprintf(stderr, "      %s\n", capture)
		default:
			fmt.Fprintln(stderr, "  Its assembly never finished. remedies:")
			fmt.Fprintln(stderr, "    - resume it: the same workflow run resumes its assembly sitting, which assembles; then capture it with")
			fmt.Fprintf(stderr, "        %s\n", capture)
			fmt.Fprintln(stderr, "    - or capture it as it stands, with the same command — its documents are then never assembled")
		}
		return 2
	}
	return 0
}
