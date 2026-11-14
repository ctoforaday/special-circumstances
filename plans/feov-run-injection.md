# The run directory a seat never types

> The plan that built this, and the record of what changed since, is in
> [`plans/historical/feov-run-injection.md`](historical/feov-run-injection.md).

Current design of how a seat verb learns which run it is in, and how identity travels with it.
Verified against the tree 2026-09-05.

## Why nothing is hand-typed

Two measurements set the shape, and both are quoted in the code that answers them
(`internal/seatenv/seatenv.go`, package comment):

- The first live run recorded **55 tool-call errors in 534 executions, TEN of them one flag** —
  a seat copies the engine's `register --run <dir> --seat-id <id>` line and then improvises
  later verbs, dropping `--run`. `InferRunDir` was added for that and never fired on the real
  path: the prompt handed the seat an absolute `--run` at every call site, and an explicit flag
  always wins.
- **Worse than absence is a wrong value.** In the 2026-08-05 smoke, `blue-respond-r1` typed
  `special circumstances` — a space where the path has a hyphen — and the tool answered *"names
  gap R1-2, which no mint event created"*. The seat believed the tool, filed friction blaming a
  dangling-reference rule, and abandoned the manifest receipts for R1-3…R1-7. One typo cost five
  receipts and produced a false bug report, because a hand-typed path is an unmediated fact and
  nothing could refuse it ([[facts-are-fields]]).

## The hook, then the tool's own inference

The run directory reaches a seat's call through the PreToolUse hook, which rewrites every Bash
command in a live run the caller's session leads to carry `export FEOV_RUN='<runDir>';`, re-derived
per call from `run-live.json` (below). When the hook does not fire, the tool finds the run itself:
`runlive.InferRunDir` over the live rows the calling session (`CLAUDE_CODE_SESSION_ID`) leads. With
neither, a seat verb is refused with *"--run <runDir> is required"*; nothing else binds a run.

`binDir`, which the workflow interpolates into every seat's prompt as `"${binDir}/feov-record"`, is
the absolute directory of the binary `setup`'s preflight checked (`internal/setup/run.go`). A seat
runs that binary by path and never types the run.

**#500's cross-run hazard is refused upstream.** Run 3's `assemble` wrote a foreign VERIFIED into a
live run after the singleton marker moved. A marker cannot move under a live run now: one live run
per project (#1385), setup refuses while a decided run is uncaptured (#1535), and a run's members are
its lead session's agents (#1582). The window that remains: a seat that outlives its run's capture,
in a session that has since set up a new run, is injected the new run. `register`-first refuses its
other writes there; its `register` succeeds.

**`run_via` records which path supplied the run** (`injected` | `flag` | `inferred`). A hook that
did not fire leaves `inferred` or `flag` and no `hook_version` stamp, which is the hook-absence
signal #512 asked for.

## The hook half — `hookgate.PreOutcome`

`PreOutcome(in Input, runDir string) (Outcome, string)` is the single entry point for the
PreToolUse decision (`internal/hookgate/hookgate.go:246`); `internal/hookcmd/hookcmd.go:108`
is its one call site.

- **`runDir` is a parameter, never a field on `Input`.** `Input` is unmarshalled straight from
  the hook payload, so every field on it is wire-supplied; a CLI-computed member would leave a
  reader unable to tell a derived value from something the client sent. `PostDropped` takes its
  injected readers for the same reason. `hookcmd` resolves the directory from the payload's
  **`cwd`** (delivered natively on both Claude Code and `agy 1.5.4+` hook payloads, or via
  `hookcore.ProjectDir` / `FEOV_RUN` for cross-harness `debate.py` seats — [#1866](https://github.com/ctoforaday/special-circumstances/issues/1866), [#1871](https://github.com/ctoforaday/special-circumstances/issues/1871)) — the seat's working directory, not the hook process's `os.Getwd()` — through
  `seat.InferRunDir` (`hookcmd.go:99-103`, `cwdOf`). No marker, or a marker naming a directory
  that does not exist → empty → **no rewrite**, matching `InferRunDir`'s "say nothing rather
  than guess".
- **Deny is first, structurally.** `PreOutcome` consults the deny arm before anything else and
  cannot return `OutcomeRewrite` when it fires. A rewrite would occupy the slot a deny needs —
  one `hookSpecificOutput` document, never two — and emitting one in place of a deny would
  silently open the blue-report lockdown. `PreDecision` stays exported for its existing callers
  and cannot emit a rewrite.
- **Injection is unconditional** on every Bash command in a live run — there is no matcher and
  nothing to look at. Nothing is ever *inserted* into the command: the prefix is prepended
  whole, so heredoc bodies, quoted prose and documentation being written to a file are
  byte-identical afterwards. That property is what lets it run unconditionally.
- **Two variables travel: `FEOV_RUN` and `FEOV_AGENT_ID`.** Identity is the half that was
  missing — `FEOV_SEAT` had readers and no writer, because only `register` knows which agent
  holds which seat and that is downstream of the first call. What the harness can hand over is
  `agent_id`, so that is what travels; the record resolves it to a seat, where the mapping is a
  field somebody wrote.
- **`export …;`, never an inline `VAR=x cmd` prefix.** Measured: blue's real command was
  `cd C:/… && "…/feov-record" blue manifest-row …`, where an inline prefix binds to `cd` and
  never crosses the `&&`. An export is a statement; it applies to everything after it.
- **Quoting is a security boundary, not formatting.** Each value is single-quoted with `'`
  escaped as `'\''`, and a value carrying a control character is **refused** — that variable is
  dropped — because a newline inside the emitted prefix would end the export statement and turn
  the remainder into a command. The run directory comes off disk and the agent id off the wire;
  neither is trusted for having been read by us.
- **Idempotent per variable, not per command.** A command already carrying `FEOV_RUN` still
  gets `FEOV_AGENT_ID`, which is what a seat copying an earlier rewritten command produces.
  Whole-command idempotency would have silently skipped the other variable — the plausible zero
  one layer down: identity absent, and the absence looking exactly like a main-session call.
- **An empty value is not injected as an empty string.** `export FEOV_AGENT_ID='';` would make
  "the main session, which has no agent id" indistinguishable from "an agent whose id is the
  empty string" at every reader downstream.

Fuzzed at `internal/hookgate/inject_fuzz_test.go` with stated invariants rather than
crash-only: the value round-trips through an independent single-quote decoder, the seat's
command survives verbatim, the rewrite is idempotent, and a refused value is never emitted.

## The reading half — `seatenv.Resolve`

`Resolve(flagRun, infer)` answers "which run am I in?" for a seat verb.

**Order: `FEOV_RUN` → `--run` → inference → empty** (the caller's own "`--run` is required").
`ResolveWithSource` additionally reports *which* path supplied it, which is what `run_via`
records.

**The disagreement is the whole point.** When both are present and differ, neither is trusted:
obeying the flag reinstates the typo, and silently overriding it would make a seat's own
argument vanish without a word. Both values are named in the refusal, so the operator can see
which is the typo. Trailing separators are tolerated on the comparison only — `…/run` and
`…/run/` are the same directory. Nothing else is normalised: case and separator style are
genuinely different paths on the platforms this runs on, and a guess here attaches a seat's
events to the wrong run, which is the outcome the whole mechanism exists to prevent.

**Scoped to seat verbs.** Operator commands (`verify`, `capture`, `dashboard`, `graph`,
`count-claims`, `scorecard`, `setup`) do not route through the seat path and are never refused;
an operator running against an archived run while another is live is unaffected.

## Known limits that still bind

- **The deny arm has the mention-vs-invocation confusion the rewrite arm no longer can.** Its
  write patterns match anywhere in a Bash command, so a heredoc *containing* `cp … blue/report.md`
  is refused as though it were a write. Found by hitting it, not by review; unfixed.
- **Prompts may still emit `--run`.** Where they do, a typo lands on the disagreement refusal —
  which is itself a tool error. "Zero run-path tool errors in a live run" is the follow-on's
  target, not this design's; this design's claim is only that **no wrong run directory ever
  wins silently**.

---

## Appendix A — the identity concept

Not part of the injection work; recorded here because it is where the principle was written
down, and `plans/record-protobuf.md` cites this appendix.

**The principle** (gb): *all of the "frank exchange" in Frank Exchange of Views is
tool-mediated.* Two failure classes: **markdown standing in for records**, and **string
concatenation / regexes standing in for structured-data paths**.

**Resolution** (gb): **`agent_id` is a collision-free identifier — if a UUID is needed, it is
that.** A per-`agent_id` counter namespace is unique by construction; identity lives in the
record as a **field**, not concatenated into a name. **Shipped** (#348): `PreOutcome` injects
`FEOV_AGENT_ID`, `internal/seatenv/identity.go` reads it, and round and role now arrive as
structured facts instead of being recovered at the append path. `FEOV_SEAT` and `FEOV_ROUND`
were deleted — both had readers and no writer, and an injected branch no run can reach is
scenery, not a fallback.

**Still standing, and each is deliberate rather than unnoticed:**

- `roundRe` (`record/round.go:8`) recovers the round from the seat id. `RoundOf` now returns
  `(round, known)` — a bare `0` meant both "round 0, which is a real round with real events"
  and "this name says nothing", and that ambiguity rendered a bench closure at run end as a
  closure *before round 1* (#327, found at 1 seed in 60). The regex is also no longer a hope
  about shape: the roster gate refuses a registered seat id unless it matches one of the
  engine's own patterns, so the segment being read is one the tool already validated.
- `roleRe` (`record/findinglabel.go:12`) recovers the lens role. The per-role prefix
  (`L1`, `L2`, `L3`) is **the uniqueness namespace that makes label allocation safe** under
  parallel dispatch — `NextFindingLabel` counts the prefix and returns `n+1` — not a
  comparability label. Removing it recreates the colliding `L5-F1`s that made 39 of 60 disposals
  ambiguous in run 3.
- `Sprintf("R%d-%d")` (`record/replay.go:525`) encodes the round into the gap id.

**Since resolved** (full reasoning in the historical document): the `"estoppel —"` prose-substring
count is now a field read (#283); `blue/frontier.md` became lines of inquiry, which carry an id
(#297); `blue/CHANGELOG.md` is gone and the round record is counted from `revision` events
(#251); red's gap-pattern memory is `inputs/gap-patterns-by-class.json`, handed to a repairing
seat by class.

**The envelope is protobuf, not lowerCamel JSON.** `internal/record/recordpb/record.proto`
defines `Event` with `seat_id`, and the store is SQLite (`plans/record-sqlite.md`). The earlier
claim in this appendix that `Event` is lowerCamel (`seatId`, `ts`) was true when written and is
false now; `plans/record-protobuf.md` tracks it as an outstanding `[MODIFY]`.

**Also open there**: red's agent-memory directory is keyed on the dispatched agent type
(`setup/run.go`, `commands/research.md`), so a per-lens split must name a distinct merge agent
rather than reuse `red-auditor`.
