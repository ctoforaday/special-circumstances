# Hook failures reach someone — prosthetic-conscience and frank-exchange-of-views

## I. Summary and goals

A hook's stdout and stderr at exit 0 go to the client's debug log and nowhere else. Every hook in
this suite exits 0 by contract, so every failure any of them reports is, today, reported to nobody.
gray-area's capture hook was fixed this way in #990 (a failure record keyed by stage, announced as a
top-level `systemMessage` on the events the client displays). A sweep of the other plugins found the
same defect in ten prosthetic-conscience hook binaries and three frank-exchange-of-views ones, plus
nine places that say nothing on ANY channel. (`sc-doctor` is the eleventh binary under `cmd/` and is
not a hook: it is human-invoked, has real exit codes, and is the OTHER PATH for several items here.)

GOAL: every FAILURE (the hook could not do its job) and every ACTION-NEEDED message (a human must
run something) in those two plugins reaches a human, on the session it happened in where the event
displays, and on the next displaying event where it does not. ROUTINE diagnostics stay on stderr.

NON-GOALS: changing any hook's exit code, decision or blocking behaviour; making SubagentStart or
SubagentStop speak (measured: an emission there re-invokes the seat, 9x); rewriting the hook-log.

WHY THIS MACHINERY AND NOT LESS. Two cheaper fixes were considered and rejected on the evidence.
(a) *Say it on stderr, louder.* Refuted by measurement (§V.1, run): stderr at exit 0 never reaches
the agent and is not displayed; there is no volume that fixes a channel nobody reads. (b) *Emit a
systemMessage at the moment of failure and keep no record.* That works only for the five displaying
events, and SEVEN of the sites here fire on events that display nothing (PreCompact, PostCompact,
SessionEnd, SubagentStop, FileChanged) — including the whole seal group, which is where a failure
costs the most. A failure that cannot speak on its own event must be KEPT until an event that can,
and a kept fact needs a record that can be cleared, or the next session inherits a stale alarm. The
record is therefore the smallest thing that covers the non-displaying events; §III.2 keeps the
cheaper path for the two items that genuinely are about the call in hand. The generation step is not
new machinery: it is `buildidgen`'s existing pattern, and the alternative is three hand-kept copies
of one file, which is the defect `facts-are-fields` names one level up.

Measured basis, today, against the real client (`claude -p --settings`, transcript read back;
evidence in `~/.claude/scratch/hookvoice/stderr-measurement.md`):
- A top-level `systemMessage` is displayed as a `hook_system_message` attachment on SessionStart,
  Stop, PreToolUse and PostToolUse. PostToolUseFailure is documented with them and untested here.
- ONE PreToolUse response carries BOTH `systemMessage` and `hookSpecificOutput.permissionDecision`:
  the message was displayed, the decision was honoured and the tool ran. So an advisory does not
  have to displace a decision.

## II. Current state — the census

Displaying events: SessionStart, PreToolUse, PostToolUse, PostToolUseFailure, Stop. Not displaying:
PreCompact, PostCompact, SessionEnd, SubagentStart, SubagentStop, FileChanged.

**prosthetic-conscience**

| # | Site | Event(s) | Class | Other path today |
|---|---|---|---|---|
| 1 | `hookenv.Explain` (`hookenv.go:51`), 7 call sites | all | FAILURE — the hook is a no-op for the session | none |
| 2 | `pushfreezeguard` advisories (`main.go:155,158,161,165,168`) and the unreadable-freeze failure (`:152`), via `pretooluse/main.go:74` (exit always 0) | PreToolUse | ACTION-NEEDED; the guard never blocks, so this IS the mechanism | none |
| 3 | `hookunit.go:158` panic isolation | PreToolUse, PostToolUse, SessionStart | FAILURE — a crashed unit reads as a clean pass | none; on SessionStart `merge` drops the text before stderr |
| 4 | `qualitygate` "qlty not found" (`main.go:120`) | PostToolUse | ACTION-NEEDED | partial: sc-doctor, hook log |
| 5 | `checkpointseal` `main.go:416,421,433,548-550,565,567`, `sealrow.go:230,289-294` | PreCompact, SessionEnd, SubagentStop | FAILURE + ACTION-NEEDED | partial: next session's restore reports loop problems |
| 6 | `filechangedrearm` `main.go:169,183,227,234` | FileChanged | FAILURE + ACTION-NEEDED | partial, same as 5 |
| 7 | `postcompactobserve/main.go:172` | PostCompact | FAILURE (one corpus row) | none |
| 8a | `toolchainnudge/main.go:93-98` — a missing, unreadable or malformed `requirements.json` degrades to silence on every channel, so the nudge reports no tools AND no reason | SessionStart | FAILURE | none |
| 8b | `checkpointseal/main.go:522` — the note read that feeds steering fails silently (the seal's own read at `:421` reports) | PreCompact | FAILURE (steering is skipped) | partial: the seal's read |
| 8c | `checkpointrestore/main.go:568-571` — `compose` Stats the note, then `os.ReadFile` fails and it returns `"", nil`: the session's own cursor note vanishes and SessionStart says nothing | SessionStart | FAILURE | none |
| 8d | `runlive/runlive.go:131-133` — a `run-live.json` that is not JSON at all returns `State{}`, so `pushfreezeguard` sees no live run and warns about NOTHING while a freeze may be in force. The file's fail-open comment covers the MISSING marker, not an unparseable one | PreToolUse | FAILURE (a guard silently disarmed) | none |
| 8 | Silent on every channel: `stopnudge.go:162,190,274,287` (fail-closed), `sessionstart/main.go:100` and `strikecounter/main.go` encode errors, `stopnudge/main.go:132` | Stop, SessionStart, PostToolUseFailure | FAILURE | partial: `seals.jsonl` records the unreadable leg only |

**frank-exchange-of-views**

| # | Site | Event(s) | Class | Other path |
|---|---|---|---|---|
| 9 | `hookcmd.go:55,60,90` (panic, error funnel, turn-limit fault) | PreToolUse | FAILURE | none; the deny document still goes out |
| 10 | `sittinghook/sitting.go:174` — `exec.Command(...).Run()` with no stderr wired, so the writer's output goes to `/dev/null`; the comment claims it "reports to stderr" | SubagentStart, SubagentStop | FAILURE, silent on every channel | none; the sibling `Limit` path uses `CombinedOutput` and DOES capture it |
| 11 | `sittinghook/sitting.go:128,131-133,146-149` — silent returns, including a missing writer | SubagentStart, SubagentStop | FAILURE | none |
| 12 | `runlive/infer.go:68` — "marker present but unusable" returns `""`, and `sittinghook/sitting.go:142-145` then drops the seat's sitting silently. The SAME return value also means "no marker, not a live run", which is the ordinary case and MUST stay silent — so the two cannot be told apart at the call site today | SubagentStart, SubagentStop | FAILURE | none |

**The contradiction is SETTLED, by measurement (§V.1, run 2026-09-16).** `sealrow.go:277-279` claimed
stderr "is MEASURED to reach the agent: it arrives inside the tool result of whatever call was
running". It does not: on PreToolUse, the one event where a tool call really is in flight, the token
appeared in no tool result, and a live session asked immediately afterwards reported it had received
no such text. It IS stored in the transcript inside a `hook_success` record whose `content` is empty
— which is the record of the hook's streams, not text for the conversation, and is how the claim
survived. `checkpointseal/main.go:556` and `hooks/README.md` are right. PR1 DELETES the losing
comment, per the no-compat rule: no softening, no "formerly".

## III. Design

### III.1 One authored failure record, generated into each module (`scripts/hookfailgen`)

The three plugins are separate Go modules by design, so the record cannot be imported. The repo's
answer to that is generation, not copies: `scripts/buildidgen` and `scripts/fetchbingen` are the
precedent, and `scripts/check` gates each. `hookfailgen` authors ONE `failures.go` and writes it to
`plugins/<plugin>/tools/internal/hookfailures/failures.go` for prosthetic-conscience,
frank-exchange-of-views and gray-area, with `-check` failing on drift.

THE COPIES ARE BYTE-IDENTICAL, like `buildidgen`'s and `fetchbingen`'s. No templating is added: what
differs per plugin — the record directory's name — is a PARAMETER the caller passes
(`hookfailures.Open("prosthetic-conscience")`), never a value generated into the file. A generator
that rewrote a literal per target would make the three copies differ, and "identical bytes" is the
property `-check` can actually assert.

Carried over from #990 unchanged, because it is now proven in the field:
- Record at `<XDG_STATE_HOME or ~/.local/state>/special-circumstances/<plugin>/failures.json`,
  beside the plugin's state, never inside a directory the failure may be about.
- An entry is keyed by (stage, scope) — `scope` is the project for a project-scoped stage, empty for
  a machine-scoped one — and holds error, event, since, last, notified.
- A stage's own next success CLEARS its entry; an empty record removes the file. So "no entry"
  means the stage last worked, never "nothing looked".
- Displaying events announce entries due (never notified, or notified more than 10 minutes ago) as
  ONE top-level `systemMessage`. Non-displaying events record and print nothing.
- No record location, or an unwritable one: the invocation's own failures are still displayed,
  unthrottled, and say the record could not be kept. An unreadable record is replaced, never read as
  empty.
- Concurrent hooks race; last writer wins; a lost failure returns on the next failure.

Changes from gray-area's copy, made once here and applied to all three:
- `displays(event)` lists all five client-displaying events (gray-area's copy lists the two its own
  hook fires on, which is the same set intersected with its events).
- The record directory is the PLUGIN name. gray-area moves from `gray-area-capture/` to
  `gray-area/`; per the no-compat rule nothing reads the old path, and an orphaned file on a box
  that ran 0.14.0 is ephemeral state, not data.
- A stage is declared by its OWNER (the unit or package that can fail), not centrally, so a new
  failure path cannot be added without naming what it is.

### III.2 An advisory about THIS call is shown on THIS call, with no record (items 2, 4)

`pushfreezeguard`'s five warnings and `qualitygate`'s skip are not states that persist; they are
facts about the tool call in hand, and both events display. They travel in the same response as the
decision — measured today — as `systemMessage`, and the units' `Result.Stderr` keeps its debug-log
copy. `pushfreezeguard`'s unreadable-freeze case (`:152`) is a FAILURE and goes to the record as
well, because it persists until the marker is readable.

Warnings still survive a denial (today's merge policy), and the decision document is unchanged.

### III.3 Wiring, per binary (items 1, 3, 5–8)

- `hookenv.Explain` takes the recorder and records stage `project-root`; its doc comment, which
  currently endorses stderr as "the right channel", is rewritten to say why that was wrong.
- `hookunit.Run` records stage `unit-panic` scoped by unit name. `sessionstart.merge` stops dropping
  `Result.Stderr`.
- `checkpointseal`: stages `snapshot-dir`, `snapshot-read`, `snapshot-write`, `seal-row`,
  `note-check`, and ACTION-NEEDED stages `loop-problems`, `note-drift`, `written-at`. All three of
  its events record only; the next SessionStart or Stop announces. SubagentStop keeps writing
  nothing to stdout, and a test pins that.
- `filechangedrearm`: stages `loop-opens-no-checks`, `unclaimed-change`, `rearm-write`,
  `ambiguous-check-keys`. Its own event never displays.
- `postcompactobserve`: stage `observation-append`.
- `toolchainnudge`: stage `toolchain-manifest` (item 8a) — the nudge's silence becomes a reason.
- The note read that fails after the note was seen: stage `note-read`, from BOTH readers — the seal's
  steering read (item 8b) and `checkpointrestore.compose` (item 8c). One stage, because it is one
  fact about one file, and either reader's success clears it.
- `runlive.Read` (`runlive.go:114-148`, the package's one `json.Unmarshal`; `Live` is a pure method
  over what Read returned): stage `run-live-parse` when the marker exists and is not JSON (item 8d).
  A MISSING marker stays silent — that is the ordinary "no run" case, and the existing fail-open comment is
  right about it.
- `stopnudge`: the four fail-closed returns and the encode failure record stages `nudge-state-read`,
  `nudge-state-write`, `nudge-encode`. `sessionstart` and `strikecounter` record `encode`.

### III.4 frank-exchange-of-views (items 9–11)

- `hookcmd.Run` records `hook-panic` and `hook-error`; `Pre` records `turn-limit`. PreToolUse is
  FEOV's only displaying event, and it is where the announcement lands.
- `runlive.InferRunDir` must distinguish "no marker, not a live run" from "marker present but
  unusable" (item 12), because today both are `""` and only the second is a failure. It gains a
  second return value saying whether a marker was SEEN. This is the one API change in the plan, and
  it is forced: a call site cannot report a failure it cannot distinguish from the healthy zero.

  **Its consumer census** — four non-test callers, each RULED here, because a contract that gains a
  signal and wires it into one caller has moved the silence rather than ended it:

  | Caller | Surface | Ruling |
  |---|---|---|
  | `sittinghook/sitting.go:142` (`handoff`) | SubagentStart, SubagentStop | records `sitting-run-dir` when a marker was seen; the seat's sitting is lost otherwise |
  | `hookcmd/hookcmd.go:102` (`Pre`) | PreToolUse | RECORDS `run-dir-unusable`. `hookgate.PreOutcome` returns `(OutcomeNone, "")` on an empty runDir, so the injection simply does not happen — and `infer.go`'s own doc measures that class as ten of 55 tool-call errors in one run. PreToolUse displays, so this one is said on the call it happened to. |
  | `hookcmd/hookcmd.go:131` (`enforceLimit`) | PreToolUse | same stage, same reason: an unusable marker means the limit is not counted |
  | `cli/root.go:134`, `cli/seat/seat.go:262` | CLI, human-invoked | OUT OF SCOPE and stated: these have real exit codes and a human reading them. They take the mechanical `dir, _ :=`, and no stage — a CLI that cannot find the run says so through its own error path. |
- `sittinghook.spawn` switches to `CombinedOutput()` and records `sitting-write` on failure, which
  is what the sibling `Limit` path already does; the comment claiming the child "reports to stderr"
  is deleted. A missing writer records `sitting-writer-missing` from BOTH paths.
- The bootstrap window (the writer not yet fetched) is not special-cased: the entry clears as soon
  as a sitting is written, so a transient miss shows at most one message and a permanent one keeps
  showing.

### III.5 What this does NOT change

Exit codes, permission decisions, `additionalContext`, watchPaths, the hook log, the seal record,
hooks.json (no new binary, no new event), and every ROUTINE stderr line (torn-tail repair, "closed N
sessions", the strike counter's deliberate stderr duplicate of its `additionalContext`).

## IV. Risks and the alternatives rejected

- **Noise.** A broken machine could show several stages at once, every 10 minutes, on Stop. Bounded
  by: only FAILURE and ACTION-NEEDED reach the record; an entry clears on the stage's own success;
  one message lists every due stage.
- **A record that is itself broken** hides what it holds — hence the unpersistable and unreadable
  paths, both tested.
- **Races.** Hook events run in parallel; the last writer wins. Rejected alternative: a lock. A hook
  must never wait, and a lost entry returns on the next failure.
- **Rejected: put everything in `additionalContext`.** It reaches the model, costs session context
  on every turn, and asks the agent to relay a fact to the human. A systemMessage is for the human.
- **Rejected: one shared record for all plugins.** Separate modules, separate release cadence; a
  plugin must not parse another plugin's stage set.
- **Rejected: a hand-written copy per module.** That is the defect `facts-are-fields` names, one
  level up; generation is the repo's existing answer.
- **Windows.** The record is JSON at a `%LOCALAPPDATA%`-free path resolved through `os.UserHomeDir`,
  as gray-area's already is, and gray-area's suite runs on the Windows leg.

## V. Verification plan

1. **Settle the contradiction first** (§II): a scratch hook on PreCompact and on Stop writes a
   unique token to stderr and exits 0; read the session transcript and the debug log for the token.
   The comment that loses is deleted in PR1. Re-arms: any claim that stderr reaches the agent.
2. **The census is re-derived, not remembered.** Two audit passes each found sites the hand-written
   census had missed (8a, then 8c/8d/12), all of the same class, which is evidence about the METHOD:
   an enumerated list of failure sites is a document standing in for a sweep. So BEFORE each PR, the
   implementer re-runs the sweep over the binaries that PR touches and records its result in the PR
   body: for every wired `hookunit.Unit` and every hook entry point, every `return` that yields the
   healthy value on an error path, every `if err == nil` guard with no else, and every `_ =` on a
   call that can fail. Each hit is then classified FAILURE / ACTION-NEEDED / ROUTINE in the PR body,
   and a FAILURE with no stage is a defect in the PR, not a follow-up. The list in §II is the sweep's
   result at 52e2e751, not its definition.
3. **Per-item tests, breaking the real thing** (the #990 method, not injected failures): an
   unwritable state file, a note whose loop opens no checks, a snapshot directory that cannot be
   created, a `PATH` with no writer beside the hook, a panicking unit, a project root resolved from
   neither source. Each asserts: recorded under its stage; displayed on a displaying event; silent
   on a non-displaying one; cleared by that stage's own success.
4. **Mutation pass** — the repo's "delete the row, invert the branch, run the suite" discipline
   (CLAUDE.md, Verification), driven by a throwaway script over a copy of the module as in #990
   (there: 43 killed); there is no committed mutation tool and none is added. Over every new
   `fail`/`ok` call site and every branch of the generated record. Any survivor is either killed by a new test or named in the test
   file as unreachable, with the reason.
5. **SubagentStop stays mute**: a test drives `sc-subagentstop` and FEOV's SubagentStop hook with a
   failing store and asserts stdout is empty.
6. **Live client**, per plugin: a real `claude -p --settings` session with the built binaries wired,
   asserting the message arrives as a `hook_system_message` attachment on SessionStart, Stop,
   PreToolUse and PostToolUse — and that a PreToolUse deny still denies while carrying one.
7. **The mechanism is documented where the hooks' reasons live**: a section in
   `plugins/<plugin>/hooks/README.md` for prosthetic-conscience and frank-exchange-of-views, and the
   one #990 did not write for gray-area — what the record is, where it lives, which events announce,
   and that no file means every stage last worked.
8. `go -C scripts run ./check` (adds the `hookfailgen` gate) before every push; the gray-area and
   prosthetic-conscience suites also run on the Windows leg in CI.

### Sequence (four PRs, each green on its own)

1. `hookfailgen` + the generated record in all three modules + gray-area migrated onto it + the
   comment corrections the measurement settles (`sealrow.go:277-279` deleted, `hookenv`'s
   "stderr is the right channel", `sittinghook`'s "it reports to stderr") + the three README
   sections.
2. prosthetic-conscience advisories: pushfreezeguard and qualitygate through `systemMessage`
   (items 2, 4).
3. prosthetic-conscience failures: `hookenv`, `hookunit`, sessionstart/stop/strike encode, seal,
   filechanged-rearm, postcompact-observe, toolchain-nudge, checkpoint-restore's note read, the
   unparseable run-live marker (items 1, 3, 5–8d).
4. frank-exchange-of-views (items 9–12), including `InferRunDir`'s second return value.

A release of prosthetic-conscience and frank-exchange-of-views carries it; gray-area needs one only
if PR1's record-path move ships.
