# Command surface: make the tool-driven record beat the markdown

> STATUS 2026-09-02: shipped — historical record (§VI is the retrospective). Later shipped work
> reversed or renamed several specifics — `--comment`, the `--file`/`--text` spelling, the
> inline verb list, the positional `--seat-id` — each marked in place below.

## I. Summary & Goals

The 2026-07-18 run wrote every artifact twice — by hand as markdown, and through
`feov-record` as events. The hand-written copy won on richness:

| artifact | hand-written | rendered from events |
|---|---|---|
| `red/archive.md` | 34,086 bytes | 7,527 bytes |
| `red/ledger.md` | 24,715 bytes | 18,189 bytes |
| citation ledger | 29 rows | 12 rows |

That gap is the whole problem. The events are meant to become authoritative, and a
projection that carries a fifth of the archive's prose cannot replace it. **The goal is
not parity — it is for the tool-driven artifact to be BETTER than the markdown**, because
only the event stream is queryable, attributable, and machine-checkable.

Measured on the same run: **245 `feov-record` invocations, 63 failures — 25.7%**, against
10.3% across all tools. Our own tool is the least reliable thing the seats touch.

## II. Technical Context & Design

### The audit, by failure class

Every failure classified as: missing-required (make required or positional), misused
(fix the help), or missing (implement it).

| n | failure | class | disposition |
|---|---|---|---|
| 15 | `--run is required` | missing-required | **FIXED** — inferred from `.claude/run-live.json` |
| 5 | `--seat-id is required` | missing-required | make POSITIONAL (below) |
| 11 | verb outside this seat's role | misused | discoverability, not error text (below) |
| 2 | `unknown flag: --file` on `merge close` | misused | **naming defect** — `close` calls it `--prose-file` |
| 5 | `unknown flag: --note/--detail/--target/--line` | missing | universal `--comment` (below) |
| 2 | `--file` pointed at a temp file that did not exist | missing | stdin removes the staging step |
| 2 | unknown command | misused | discoverability |

### Why the markdown is richer: three causes, all fixable

**1. There is no stdin anywhere.** Every payload is an inline quoted string or a
`--file` pointing at a staged temp file. The transcripts show what that costs: **68
commands carrying escaped quotes, 9 heredocs, 37 `--file` uses** (each needing a separate
write first, twice failing because the staged file was not there). A seat writing rich
prose into markdown pays nothing; the same prose through the tool means fighting the
shell. Prose therefore goes where it is cheap, and the event gets the summary.

**2. Payload flags are not uniform.** `--file`/`--text` exist on `finding`, `friction`,
`observe`, `closing`, `mint`, `position`, `avenue`, `manifest-row` — but `merge close`
names the same concept `--prose-file`, and NINE verbs have no free-text channel at all
(`cite`, `dispose`, `dispute-respond`, `regrade`, `confidence`, `dispute`, `petition`,
`register`, `verdict`). A seat that learned `--file` on one verb is punished on another.

**3. There is nowhere to put what the schema did not anticipate.** Seats reached for
`--note`, `--detail`, `--target`, `--line` and were refused each time. That information
did not evaporate — it went into the markdown. Every such rejection is a row of evidence
moving from the queryable channel to the unqueryable one.

### The design

**A. Stdin for every payload.** Any verb accepting prose accepts it on stdin when the
payload flag is absent. `--file -` is the conventional spelling and needs no new flag.
This removes escaping from the seat's problem entirely and deletes the stage-a-temp-file
dance that failed twice.

**B. One payload vocabulary.** `--file` and `--text` on every verb that takes prose;
`--prose-file` becomes a deprecated alias on `close` rather than a second vocabulary.

> **SUPERSEDED (2026-09-02):** the shipped vocabulary is `--reason` / `--reason-file` (with
> `-` and bare stdin), one central resolver every prose verb routes through
> (`seat.Prose`/`flags.RegisterPayload`). The concept here — one channel, stdin, no
> `--prose-file` — landed; the `--file`/`--text` spelling did not survive the later
> prose-fan-out collapse to `reason`.

**C. A universal `--comment`.** Every verb, no exceptions, attached in `seat.New` so it
cannot be forgotten when a verb is added. It is the pressure valve: when the schema has
no field for what the seat knows, the note lands in the event rather than in prose the
tool will never see. Cheap to add, and each one is a candidate for a future first-class
field — the comments become the backlog of missing schema.

> **REVERSED (2026-09-02):** built, then REMOVED 2026-07-20 — no report ever surfaced a
> comment, so prose put there was lost to the reader; a missing field is now a friction entry
> (the removal note lives in `internal/cli/seat/seat.go`).

**D. `--seat-id` positional.** It cannot be inferred: shell state does not persist
between tool calls, cwd is reset, the engine is sandboxed, and every subagent shares the
parent's session id (all four verified). It is required on every verb, so it should be an
argument, not a flag — `feov-record lens finding red-lens-r1-L1 --location ...`. Cobra
then refuses a missing one with an arity error instead of a flag error, and there is no
flag NAME to forget. This is a breaking change to every seat prompt and 22 goldens, so it
lands on its own.

> **SUPERSEDED (2026-09-02): never done, and no longer viable** — identity is DETECTED
> (hook-injected; #290/#480). `--seat-id` is passed once at `register`, which binds it on the
> record; after that it survives only as a cross-check that refuses a disagreeing value. A
> positional would put the seat back on every command line, the thing detection removed.

**E. Discoverability at the moment of need.** 32 `--help` invocations mid-task, plus 11
verb-outside-role errors, say the role's verb set is not visible when it is wanted. The
error text already lists the available verbs and is good; what is missing is the same
list BEFORE the failure. The engine's `recordClause` should carry the seat's verb list
inline, since it already knows the role.

## III. Implementation plan (file-by-file)

1. `internal/cli/seat/seat.go` — `New` attaches `--comment` to every verb and records it
   when present. One place, so a new verb inherits it.
2. `internal/cli/seat/seat.go` — payload resolution helper: `--text`, else `--file`
   (with `-` meaning stdin), else stdin when it is not a TTY.
3. `internal/cli/merge/close.go` — accept `--file`. ~~keep `--prose-file` as a hidden
   alias~~ — **superseded: the aliases were DELETED, not hidden.** An alias preserves a
   spelling for readers who remember the old one, and there are no humans in this loop:
   every caller is a prompt we rewrite in the same commit. A hidden alias would only let
   a stale prompt keep working silently, which is how the divergence lasted this long.
4. Verbs with no prose channel — add `--file`/`--text` where a payload is meaningful
   (`cite`, `dispose`, `dispute-respond`, `regrade`, `confidence`, `dispute`).
5. `debate.js` `recordClause` — inline the role's verb list. **SUPERSEDED (2026-09-02):** the
   shipped `recordClause` does NOT inline the list — it carries `SEAT_ID` plus a required
   read-the-whole-help-tree protocol, because the help is the one authoritative vocabulary and
   a prompt-inlined list is a second copy that drifts.
6. SEPARATE, LATER: `--seat-id` positional across all roles, with prompts and goldens.
   **SUPERSEDED (2026-09-02): see the marker under §II design D — detection replaced it.**

## IV. Risk & Mitigation (likelihood x impact x complexity-to-mitigate)

| risk | l x i | mitigation |
|---|---|---|
| stdin blocks forever when nothing is piped | med x **high** | only read stdin when it is NOT a terminal; never block on an interactive invocation |
| `--comment` becomes a dumping ground and schema work stops | med x med | it is a MEASURE, reported per run; a field that recurs in comments is a schema gap to close, and the count is the signal |
| a universal flag collides with an existing verb flag | low x med | `comment` is unused across all 30 verbs (verified by enumerating every verb's flags) |
| deprecated `--prose-file` silently diverges from `--file` | low x med | alias, not a second parameter — one destination, one code path |
| positional `--seat-id` breaks every prompt at once | high x high | separate change, its own PR, goldens re-recorded in their own commit |

## V. Verification plan

1. `go test -race -count=1 ./...` in `plugins/frank-exchange-of-views/tools`.
2. A test per failure class taken from the real run: `--file` accepted on `close`;
   `--comment` accepted on EVERY verb (enumerated, so a new verb without it fails);
   stdin payload equals the `--file` payload byte-for-byte; stdin not consumed when a TTY.
3. **The regression fixtures are the run's own errors** — `--note`, `--detail`,
   `--target`, `--line` must now be expressible, and `--file` on `close` must work.
4. Re-run `record-parity-check.mjs` against a future run: the archive byte-gap
   (34,086 vs 7,527) is the metric this plan is judged on, and it must close.
   **OBSOLETE (2026-09-02):** the parity scripts were retired (cf43e15a folded those audits
   into capture's record-reading audits), and the dual-write model itself is gone — the record
   is the sole channel ("routing around it into markdown is the failure this contract exists
   to prevent") and the `.md` files are human-verification projections. There is no byte-gap
   left to measure; the goal this metric guarded was overshot, not missed.
5. Goldens re-recorded in their own commit, with the help-text diff read line by line.

## VI. What the implementation taught (2026-07-19)

Three things the plan did not anticipate, all of the same shape — **a single source of
truth that nothing reads is not a source of truth.**

1. **The vocabulary chokepoint was inert.** `internal/flags/names.go` was written so a
   verb could not invent a private spelling without editing it. 20 of its 24 constants
   were never referenced; the verbs went on registering literals. It had already drifted
   internally — `Supersedes` held `"supersedes"` while every call site registered
   `superseded-by` — and nothing caught it, because nothing read it. Constants only bind
   at the call site.
2. **Error messages are a second, unaudited copy of the vocabulary.** `record.go` derived
   the flag name in `opinion requires --%s` from the payload key. The derivation was true
   when written and silently false after the rename. **The error string is the seat's
   only teacher** — it belongs to the vocabulary and must come from the same place.
3. **A test that reimplements its subject cannot indict it.** The test guarding those
   messages computed the expected spelling with the same transform the code used, so it
   asserted the code agreed with itself and passed straight through the rename.
   Expectations are literals wherever the subject is itself a derivation.

Verification note for the literal→constant sweep: **the goldens must not move.** No flag
name or payload key value changes, so any golden diff means the substitution altered
behaviour instead of expressing it — a bug to find, not a file to regenerate.
