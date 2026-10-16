# migrate replays a run in one transaction

> STATUS: approved 2026-10-16 (audit: PASS in rounds 2 and 3), being implemented. Issue #1382. The second half of the work begun in
> `plans/migrate-staged-publish.md` (merged, #1383/#1386): gblock ruled on 2026-10-15 that the
> per-event fsync is removed by batching the replay through #1379's caller-held-transaction path,
> keeping `synchronous=FULL` and adding no process-wide switch.
>
> Audit round 1 FAILED on two local gaps (functions handed the destination `Run` while the batch is
> open; the factored act body's boundary). gblock ruled its three forks on 2026-10-16: `Batch` is
> enforced migrate-only; seat turns stay one commit per agent (stated); a failed reorder trial is
> ROLLED BACK instead of the record being closed, deleted and replayed. Round 2 PASSED with notes,
> folded in here; this is the text of the confirming round.

Paths are relative to `plugins/frank-exchange-of-views/tools/internal/` unless they begin with
`plans/` or `scripts/`. Census additions from round 2 (the `Translate` test call sites:
`entries_test.go:31,78,91,119,144,171,197`, `remanded_test.go:64`, `remaplists_test.go:278,284`,
`repairregister_test.go:39,57`, `publish_test.go:686`) are in
`~/.claude/scratch/migrate-batched-replay-audit/r2/`. Line numbers are at `a211b996`, `origin/main` on 2026-10-16. Before any
command that runs tests or the binary against a run: `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

## I. Summary & Goals

### The problem

`migrate` replays an old record through `record.Append`, one event at a time. Each `Append` opens
its own transaction and commits it (`insertNumbered`, `record/record.go:605`), and every commit
under `synchronous=FULL` is an fsync. A run of 400–800 events (the largest archives:
quadratic-formula 813, wave-b 706, m23 476, m20 437) pays roughly 900–1,700 commits — the gap
placer's anchors and an as-written fallback add more.

Measured 2026-10-14, interleaved, on disk: the test that migrates each of the 21 archived runs once
took 49.95 s and 27.46 s as today; 15.49 s and 14.40 s with `synchronous=OFF` (a prototype,
reverted); 14.6 s on tmpfs. The fsyncs are about two thirds of the time on disk.

### Success criteria

| # | Criterion | Baseline | Target |
|---|---|---|---|
| B1 | `go test -count=1 -run '^TestTheArchiveStoreMigratesEachArchiveOnce$' ./record/migrate/`, on disk, TMPDIR on disk, interleaved against `origin/main`, quiet box | 27–50 s | at most 20 s |
| B2 | Every migrated record is what today's code writes | — | every archive in `run-archive/`: `events` and every body table identical by `SELECT … ORDER BY` primary key (row ORDER of `seat_turn` excepted — it already varies between two runs of `main`), manifests identical minus timestamps and `tool_version`, recursive file listing identical |
| B3 | A refused event leaves nothing, and the events after it land as today | each refused `Append` rolls back its own transaction | a savepoint per act: a refused act rolls back to it, and the record afterwards is identical to today's for the same source (B2 covers the archives; a new test covers a refusal with a written row) |
| B4 | Durability of a finished migration | one fsync per commit | unchanged in kind: `synchronous=FULL`; the batch's commit is synced, then `migrate`'s publish fsyncs every file as today |
| B5 | Live writes | — | unchanged: `Append`, `DispatchNext` and every seat's path take the same transactions as today; `BeginBatch` refuses unless `record.Migrating` is set |
| B6 | A failed reorder trial | the record is closed, its files deleted, and the whole replay run again | the trial's batch is rolled back and the as-written replay runs in a fresh batch on the same database; no file is deleted |

### Value case

Against doing nothing: real migrations and the migrate tests stop paying an fsync per event —
about two thirds of their time on disk — wherever no tmpfs is used (macOS/Windows dev boxes,
`go test` outside `check`, an operator's migration). Against `synchronous=OFF` (planned and
abandoned, `plans/migrate-bulk-load.md` history in #1383): no relaxed durability mode, no
process-wide registration; the store's own stated preference (`applySchemaTx`'s comment,
`record/recordsql/store.go:742-758`) is exactly this — "the same guarantee, asked for once".

## II. Technical Context

- Go 1.27, `modernc.org/sqlite`, WAL, `synchronous=FULL` (the default). The cached handle holds ONE
  connection (`SetMaxOpenConns(1)`, `recordsql/store.go:389`): any read through the pool while a
  transaction holds that connection waits on itself forever (`store.go:381-388`,
  `record/writeview.go:17-19`).
- Since #1379 (`ce1f1be5`), every validation read of a write goes through the writing transaction:
  `writeIn(tx, …, check)` (`record/record.go:621`) with `check` = `validate(writeView{q: tx, files},
  …)`. `writeView` holds no `Run`, so a pool read cannot be written there. `DispatchNext`
  (`record/dispatchnext.go:26-131`) already writes several acts in one caller-held transaction:
  `openRun` and `filesOf(run)` before `Begin`, every read through the `*sql.Tx`, one `Commit`.
- What replay does between appends, outside any transaction:
  - the gap placer renders the report through the pool — `reportproj.RenderFromRecord(p.dst)`
    (`record/migrate/gapplace.go:64,102,119,138`) → `record.ReportProjection(run)`. A transactional
    form exists: `record.ReportProjectionAt(q recordsql.Querier)` (`record/queries.go:290`) and
    `reportproj.Replay(base, have, ops)` (`reportproj/render.go:264`);
  - the correction pair: `appendCorrected` (`record/correction.go:214`) reads through the pool
    before its own `Begin` — `correctionRetry(db, …)` (`:346`) and `readTarget(db, …)` (`:59`);
  - `record.SeatOwesOccasion` reads nothing (a roster lookup); registry `Translate` entries do not
    read the destination today (`eraentries.go:125,169`).
- After `replay` returns: `replayInOrder` reads `ReportProjection(dst)` (`order.go:48`) and
  `RenderFromRecord(dst)` (`order.go:75`), and may close, delete and rebuild the record
  (`order.go:59-69`); `AppendSeatTurns` (`record/seatturn.go:32`) opens one transaction per agent.
- Replay pins `record.Now` per event (`replay.go:111,180`) and sets `record.Migrating`; `envelope`
  reads `Now()` inside `writeIn`, so pinning holds inside a shared transaction.
- A refused act today leaves nothing: `insertNumbered` and `appendCorrected` roll back their own
  transactions. In a shared transaction, `InsertTx` writes the `events` row before the body, and
  its constraints can fail after it — body tables carry enum foreign keys (`foreign_keys(1)` is in
  the connection string), NOT NULL, CHECK and UNIQUE constraints, and `appendCorrected`'s second
  `InsertTx` can fail after its first — so an act's rows are undone with a `SAVEPOINT`. (The two
  `RAISE` triggers, `recordsql/schema.go:99-103`, are `BEFORE UPDATE`/`BEFORE DELETE` on `events`
  and never fire on an append; none uses `RAISE(ROLLBACK)` or `RAISE(FAIL)`, so nothing can end the
  whole transaction from inside an act.) SQLite supports savepoints inside `BEGIN`; through `database/sql` they are plain
  `tx.Exec` on the transaction's connection. Proven by the round-1 auditor on modernc v1.59.0 under
  `BEGIN IMMEDIATE` with `foreign_keys(1)`: `SAVEPOINT`/`ROLLBACK TO`/`RELEASE` work, constraint
  failures after the `events` INSERT are undone, `RELEASE` does not commit, and another connection
  sees nothing before `COMMIT` (`~/.claude/scratch/migrate-batched-replay-audit/r1/probe/`).
- Paths a live seat cannot reach: `appendCorrected` and the appeal are written only under
  `record.Migrating` (`record.go:546-562`). `RegisterSeat` is not on the replay path — replay sends
  register events through `Append` (`replay.go:67-73`).
- Repository rules: `migrate` is self-contained; nothing is added to the live RECORD for migration's
  sake. This plan adds a write API to the `record` package — how acts are committed, not anything
  stored — used by `migrate` only, as `DispatchNext` is used by the chair only.

## III. Proposed Changes

```
record/
  batch.go            [NEW]     Batch: one caller-held transaction for many acts, a savepoint per act
  batch_test.go       [NEW]
  record.go           [MODIFY]  Append's act body factored so Append and Batch.Append share it
  correction.go       [MODIFY]  appendCorrected's pre-Begin pool reads take the transaction; the dead retry-after-rollback block goes
  recordsql/bykey.go  [MODIFY]  EventByKey takes a Querier
  recordsql/store.go  [MODIFY]  the single-connection comment names a transaction held across replay
  writeview.go        [MODIFY]  its comment, likewise
reportproj/
  render.go           [MODIFY]  RenderFromRecord over the one Querier render (see step 2)
record/migrate/
  replay.go           [MODIFY]  replay writes through the Batch it is handed; Translate gets no destination
  registry.go         [MODIFY]  Entry.Translate(old) — the destination Run removed
  entries.go, eraentries.go [MODIFY]  every Translate drops its unused dst
  order.go            [MODIFY]  replayInOrder owns the batch: trial, keep or roll back, as-written
  gapplace.go         [MODIFY]  the placer renders through the batch's querier, never through dst
  replay_test.go or batch-specific tests [MODIFY/NEW]
```

### 1. `record.Batch`

```go
func BeginBatch(run Run) (*Batch, error)          // refuses unless Migrating; openRun, filesOf, writeBegins, db.Begin
func (b *Batch) Append(id Identity, body proto.Message) (*Event, error)
func (b *Batch) Querier() recordsql.Querier       // reads that must see the batch's own writes
func (b *Batch) Commit() error
func (b *Batch) Rollback() error                  // idempotent; after Commit, a no-op
```

- `BeginBatch` refuses unless `record.Migrating` is set (ruling, 2026-10-16): the batch holds the
  write lock across its caller's code between acts, which only `migrate` may do. Otherwise it does
  what `DispatchNext` does before its first write: `openRun(run)` (which applies the schema on a
  fresh database, in its own transaction, before this one begins), `files := filesOf(run)`,
  `writeBegins()`, `db.Begin()`.
- **One act body, two callers.** Everything `Append` does today (`record.go:529-585`) lives in one
  function, `appendAct(id, body, begin)`, where `begin` is how the caller obtains the transaction
  and is called only AFTER the guards — `run.Valid` before any handle is opened, so a write with no
  run is still refused with "this write names no run" rather than a bare `mkdir` error
  (`cli/seat/seat.go:226` depends on it; a test asserts the text). The guards: the `run.Valid` guard, `SetBody`'s refusal of a nil
  body, the appeal-only-under-migrate guard, the correction routing with its write-once
  `c.written` check, and then either the correction pair (step 3) or `writeIn(tx, …,
  validate(writeView{q: tx, files}, …))`. `Append` and `Batch.Append` differ only in how they get
  the transaction: for `Append`, `begin` opens the handle, calls `writeBegins()` (the test hook,
  for the correction pair too — it moves out of `appendCorrected`, `correction.go:233`) and
  `Begin`s, and `Append` commits;
  `Batch.Append` checks `id.Run == b.run` (`Run` is a comparable struct, `run.go:39`; a write
  addressed to another run is an error, never a silent write here), runs `SAVEPOINT act`,
  `appendAct` on the batch's transaction, then `RELEASE act` — or on any error `ROLLBACK TO act;
  RELEASE act` and returns the error. The batch stays usable after a refused act.
- **A savepoint failure poisons the batch.** If `SAVEPOINT`, `ROLLBACK TO` or `RELEASE` fails, or
  SQLite has rolled the transaction back on its own (FULL, IOERR, NOMEM), the batch is marked
  failed: every later `Batch.Append` returns that error and `Commit` fails. `Batch.Append`
  distinguishes the two kinds of error it returns — a refusal of the act (the batch is fine) and a
  failure of the batch — by a sentinel (`ErrBatchFailed`, wrapped), so replay records the first as
  a `Refusal` and returns the second as fatal. Without this, later savepoint pairs would run outside
  `BEGIN` and each commit on its own.
- The batch keeps its `run` unexported; replay holds `dst` only to build `Identity` values, and only
  `Batch.Append` receives it, using it only for the `id.Run == b.run` check. `writeSynthesizedCast`
  takes the batch and the run for its `Identity`, and appends through the batch.
- `DispatchNext` keeps calling `writeIn` directly: its acts are plans and dispatch records the chair
  writes, not seats' acts, and none of `appendAct`'s guards (appeal, correction routing) applies.
- The savepoint name is a constant; savepoints do not nest here (one act at a time).

### 2. Nothing called while the batch is open holds the destination `Run`

- **The render.** `record` already composes the render over a querier for the mint budget:
  `renderProjection(ReportProjectionAt(q))` (`record/mintbudget.go:201`, the renderer registered at
  `reportproj/register.go:9`). That composition is exported once as `record.RenderAt(q
  recordsql.Querier) (string, error)`, and `reportproj.RenderFromRecord(run)` becomes a call to an
  exported `record.RenderRun(run)` (`openRunForRead` + `RenderAt`, inside `record`, since
  `openRunForRead` is unexported), so there is one render over a querier, not two.
- **The placer** (`gapplace.go`) holds a `recordsql.Querier` instead of `dst`; replay hands it
  `batch.Querier()`. Its four render sites call `record.RenderAt`.
- **The registry.** `Entry.Translate(old OldEvent, dst record.Run)` (`registry.go:17-20,36`) is
  called while the batch is open (`replay.go:140,174`), and its doc invites reading `dst` — a read
  that would now wait forever. No entry reads it (`entries.go:41,83,122,158,177,190`,
  `eraentries.go:83,120,165`, and the test entry at `publish_test.go:678`), so the parameter is
  removed: `Translate(old OldEvent)`. A translation that ever needs what the migrated record holds
  is a change to this contract, made then, through the batch.
- The rule, written in `replay`'s doc: between `BeginBatch` and `Commit`, nothing is handed `dst`;
  reads go through `batch.Querier()`.

### 3. The correction pair inside a transaction it is handed

`appendCorrected(id, db, …)` reads through the pool before `Begin` — `correctionRetry` (`:346`)
and `readTarget` (`:59`), both through `recordsql.EventByKey(db *sql.DB, …)`
(`recordsql/bykey.go:13`). All three take a `recordsql.Querier`; `appendCorrected` takes the
transaction instead of the handle and no longer begins or commits. Its doc comment (`correction.go:202-213`) and `writeIn`'s
(`record.go:618-620`) are restated for the new model. The second retry check
(`correction.go:240-247`, `tx.Rollback(); return correctionRetry(db, …)`) is dead once the first
read is inside the transaction, and would end a batch's transaction if carried in; it is deleted.
Only `migrate` reaches this path (`record.Migrating`), so holding the write lock for these reads
changes nothing a live seat does.

### 4. replayInOrder owns the batch; a failed trial is rolled back

- `replayInOrder` sets `record.Migrating` and restores `record.Now` (moved out of `replay`, which
  now runs inside them), begins a `Batch` on `dst`, and calls `replay(src, reg, batch, opt, true)`.
  `replay` writes every event, anchor and the synthesized cast through `batch.Append`, records a
  refusal as today, and on a fatal error returns it; `replayInOrder` rolls the batch back on every
  exit that does not commit, including a panic (`defer batch.Rollback()`).
- The exported `Replay(src, reg, dst, opt)` (`replay.go:75`), which tests call directly, does the
  same around one pass: sets `Migrating`, pins `Now`, begins a batch, `replay`, commits — so no
  caller reaches `replay` without a batch it owns.
- The reorder decision reads through the batch: `record.ReportProjectionAt(batch.Querier())`
  instead of `ReportProjection(dst)` (`order.go:48`).
- Every path that keeps the trial commits before returning — `OrderReordered`, no report (`!have`,
  today's early return at `order.go:71-73`), and a base already in order — so the deferred rollback
  never discards a kept record. Not kept: **roll the batch back** (ruling, 2026-10-16), restore
  `record.Now` to the value it had before the trial (`writeSynthesizedCast` pins it only when the
  first timestamp parses), begin a fresh batch on the same database, `replay(…, false)`, commit. A
  batch failure (`ErrBatchFailed`) during the trial is fatal, never a reason to try as written: it
  is checked at all three append sites (event, anchor, synthesized cast). No `CloseUnder`, no file is deleted
  (`order.go:59-69` goes). The schema, applied by `openRun` in its own transaction before the first
  batch began, stays; the rolled-back events leave nothing, so the as-written replay starts from the
  same empty record the deleted-and-rebuilt one did — B2's comparison on every archive is what
  shows the ids and rows are the same.
- After the commit, the final `RenderFromRecord(dst)` (`order.go:75`), `AppendSeatTurns` (one
  commit per agent, about a dozen per run — left as it is by ruling) and the publish run as today.
- One transaction per replay pass, not chunks: a migration that fails is discarded whole (its build
  directory is removed, `plans/migrate-staged-publish.md` step 4), so a partial commit buys
  nothing; the WAL for the largest archive is a few megabytes.

### Alternatives rejected

| Alternative | Why not |
|---|---|
| `synchronous=OFF` for the destination | Planned and audited five rounds, then dropped by ruling (2026-10-15) for this |
| Commit every N events | A failed migration is discarded whole; partial commits only add fsyncs |
| One transaction with no savepoints | A refused act could leave its `events` row when a body or trigger fails after the INSERT |
| A `Batch` flag on `Append` | Puts the mode on the live path's signature; a type that holds the transaction keeps it off |
| Close, delete and rebuild the record after a failed reorder trial | Replaced by a rollback, by ruling |
| Seat turns inside the batch | About a dozen commits per run against ~1,000 removed; left by ruling |
| `Batch` exported without a guard | Ruled out: it would let any caller hold the write lock across arbitrary code |

### Consumer census

Contract changed: `appendCorrected` (takes the transaction), `correctionRetry`/`readTarget`/
`recordsql.EventByKey` (`*sql.DB` → `recordsql.Querier`); `RenderFromRecord` reimplemented over
`record.RenderAt`; `Entry.Translate` and `Registry.Translate` lose `dst`; `replay` takes a `Batch`;
the placer's fields; new `Batch`, `BeginBatch`, `RenderAt`, `ErrBatchFailed`. Round-1 census output:
`~/.claude/scratch/migrate-batched-replay-audit/r1/` (`c1`–`c3.txt`,
`auditor-census-additions.txt`, `eventbykey.txt`, `raise.txt`).

```
git grep -nE 'appendCorrected\(|correctionRetry\(|readTarget\(' -- plugins/frank-exchange-of-views/tools
git grep -nE 'RenderFromRecord\(|ReportProjectionAt\(' -- plugins/frank-exchange-of-views/tools
git grep -nE 'reportproj\.|placer|p\.dst' -- plugins/frank-exchange-of-views/tools/internal/record/migrate
git grep -nE 'EventByKey\(|Translate\(|renderProjection\(' -- plugins/frank-exchange-of-views/tools
```

The implementer pastes each command's output into the PR and states, per hit, whether it changes.

## IV. Risk & Mitigation

| # | Risk | L × I | Mitigation |
|---|---|---|---|
| R1 | A pool read inside the batch deadlocks (the single connection) | M × H | Reads in the batch go through `Querier()`; the placer holds no destination `Run`. Every migrate test and V.3 run with a timeout; a deadlock is a loud hang |
| R2 | A refused act leaves rows | L × H | Savepoint per act; V.2 forces a failure after the `events` INSERT and asserts the row count |
| R3 | The record differs from today's (ordering, ids, a validation that read differently) | L × H | V.3 compares every archive with `origin/main`'s binary |
| R4 | `Append` changes behaviour for live seats | L × H | The factored act body is the code `Append` runs today; the correction pair's reads move into its transaction — V.4 runs the record and cli suites and the release fuzz |
| R5 | SAVEPOINT misbehaves under modernc/`database/sql` | L × M | Proven by probe in round 1; V.2 proves it in the suite |
| R6 | A savepoint failure lets later acts commit one by one outside `BEGIN` | L × H | The batch is poisoned on any savepoint failure and replay returns it as fatal; V.2 injects one |
| R7 | The as-written replay after a rolled-back trial differs from today's delete-and-rebuild | L × M | V.3 compares every archive (several take the as-written path), and an order test asserts the trial left nothing |

## V. Verification Plan

1. B1: the all-archives test, interleaved against `origin/main`, TMPDIR on disk, quiet box, two
   pairs (`ps` for other sessions' fuzz runs first; a timing at load is void).
2. The batch: a `record` test begins a `Batch` (with `Migrating` set), appends a good act, then an
   act that fails after its `events` row is written (a body whose enum foreign key names no
   value, or a NOT NULL column left empty — the auditor's probe shows which the schema refuses
   after the INSERT), then a good act; it asserts the failed act left no `events` row and no body
   row, the two good acts are there with consecutive ids, and the batch commits. Also: an act
   refused by validation leaves the batch usable; a forced savepoint failure (a seam on the
   savepoint statement) poisons the batch — later appends and `Commit` return `ErrBatchFailed`;
   `BeginBatch` without `Migrating` is refused; `Batch.Append` to another run is refused; an appeal
   through `Batch.Append` with `Migrating=false` is impossible (the batch cannot exist) and a
   second same-type body through one correcting identity is refused, as through `Append`. `Append`
   with a zero `Run` returns "this write names no run". A trial batch poisoned mid-replay makes
   `migrate` fail, not publish `as_written`.
3. B2/B3 on real data: every archive migrated by `origin/main`'s binary and by the branch's,
   `--seat-id operator`, into scratch, after clearing the seat environment; `events` and every body table compared ordered by primary key
   (the `seat_turn` row-order exception stated); manifests compared minus `migrated_at` and
   `tool_version`; `find <to> -printf '%M %P\n' | sort` compared. Refusal counts per archive equal. The PR names which archives take the as-written
   path, so it is known that this covers the rolled-back trial.
4. The suites: `go test -count=1 ./record/... ./reportproj/ ./cli/` and `FEOV_RELEASE_GATE=1 go
   test ./releasegate/fuzz/` once; `go -C scripts run ./check` before push, FAIL COUNT read.
5. Mutations, each red then restored from a kept copy: drop the savepoint rollback on error (V.2
   red); render through `dst` instead of the querier in the placer (a hang — run with `-timeout
   2m`, red); commit per act inside the batch (B1 regresses — read the time, not a test); drop the
   `id.Run` check (V.2, red); drop the `Migrating` check in `BeginBatch` (V.2, red); drop the
   poisoning (V.2, red); keep a guard in `Append` but not in the shared act body (V.2's correction
   and appeal cases, red); delete the trial's rollback so the as-written pass runs on top of it (the
   order tests — this one HANGS, since the second `BeginBatch` waits on the open trial transaction
   on the one connection; run it with `-timeout 2m`).
6. Driveable: `feov-record migrate --seat-id operator` on m23 under `strace -c -f -e
   trace=fsync,fdatasync`: the count is a few per replay pass, plus one per agent for the seat
   turns, plus the publish's per-file syncs and the checkpoint at close — against roughly one per
   event on `origin/main`; both counts go into the PR.
