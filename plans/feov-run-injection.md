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

## Two carriers, and a gate between them

The run directory reaches a seat twice, by independent routes. That is deliberate: #500's
mid-run marker move (run 3's `assemble` wrote a foreign VERIFIED into a live run) was invisible
while there was one carrier and it was the one that had moved.

**1. The per-run wrapper — primary.** `setup` writes `<runDir>/.bin/feov-record` (and `.cmd`),
which execs the real binary with `FEOV_RUN_FROM_WRAPPER` set to this run, and prints that
directory as the `binDir` the workflow hands every seat (`internal/setup/wrapper.go`,
`WriteRunWrapper`; `WrapperDir = ".bin"`). The prompt already interpolates
`"${binDir}/feov-record"`, so a seat runs the wrapper without knowing it exists. The dispatcher
bakes the run in **before the run starts**; a mistyped wrapper path fails at the shell, where a
mistyped `--run` is well-formed and files the seat's work against nothing. Pointing `--bin-dir`
at a previous run's `.bin` is refused rather than wrapped.

**2. The PreToolUse hook — second.** It rewrites the seat's Bash command to carry
`export FEOV_RUN='<the live marker's runDir>';`, re-derived per call from `run-live.json`.

**The gate.** The two disagreeing is refused as *"the engine's live-run marker has moved since
this run started"* (`internal/seatenv/seatenv.go:122`) — exactly the #500 incident. On a healthy
run the wrapper is invisible: the hook sets `FEOV_RUN`, which outranks it.

**Why two variables rather than one.** `FEOV_RUN_FROM_WRAPPER` is distinct from `FEOV_RUN` so
that `run_via` stays a precise hook-absence detector (#512): if the wrapper set `FEOV_RUN` too,
every register would read "injected" and the hook-absence signal would go blind — the defect
that issue exists to end, reintroduced by the fix for its sibling. Neither variable appears on
any `--help` surface: a seat never types them, and documenting one would invite exactly the
hand-typed path this removes.

Measured on a live smoke run after #526: **249 commands through the wrapper, 0 containing
`--run`, 13/13 seats identified.**

## The hook half — `hookgate.PreOutcome`

`PreOutcome(in Input, runDir string) (Outcome, string)` is the single entry point for the
PreToolUse decision (`internal/hookgate/hookgate.go:246`); `internal/hookcmd/hookcmd.go:108`
is its one call site.

- **`runDir` is a parameter, never a field on `Input`.** `Input` is unmarshalled straight from
  the hook payload, so every field on it is wire-supplied; a CLI-computed member would leave a
  reader unable to tell a derived value from something the client sent. `PostDropped` takes its
  injected readers for the same reason. `hookcmd` resolves the directory from the payload's
  **`cwd`** — the seat's working directory, not the hook process's `os.Getwd()` — through
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
