# Catalogue shape rebuild — an older store is rebuilt, never re-stamped; the read path is really read-only

> STATUS: approved design pending implementation (2026-09-10), revision 8.3 — plan-audit PASS in
> round 11, after rounds 1–10. Split out of
> `plans/restart-recovery.md` (branch `plan/session-index`) by gblock's ruling: the live store
> defect ships on its own, first. The defect's author (the session that landed #867) agrees with
> rebuild-and-reproject over migration.

## I. Summary & Goals

**Defect 1 — the shape, measured 2026-09-10 on this box.** The gray-area catalogue at
`~/.local/state/special-circumstances/catalogue/catalogue.db` is stamped `user_version` = 2, but its
`session` table still has the shape-1 columns `first_seen`/`last_seen`. Commit `28ad35e5` (#867,
2026-09-09) renamed them to `ingested_first`/`ingested_last` in `Schema` and bumped `UserVersion`
1 → 2, but `CREATE TABLE IF NOT EXISTS` skips an existing table, `Open` (`open.go:39-69`) refuses
only a NEWER shape and stamps the current version on an OLDER one, and `v_session` is dropped and
recreated on every open (`schema.go:125`) against the new names. So `telepathy sql "SELECT
count(*) FROM v_session"` fails with `no such column: s.ingested_first`. Every store created before
`28ad35e5` is in this state and now stamped 2.

**Defect 2 — no connection this package opens is a URI, so neither path option nor path is safe
(found in plan-audit rounds 4 and 7, verified at the leaf).** modernc v1.57.0's `newConn`
(`conn.go:62-73`) cuts everything from the first `?` off any DSN not prefixed `file:`, opens the
remainder `SQLITE_OPEN_READWRITE|SQLITE_OPEN_CREATE`, and applies the `_pragma`s from the cut-off
query. Two consequences:
- **The read path is not read-only.** `OpenRead` opens `path + readDSN` (`open.go:75,86`); `mode=ro`
  never reaches SQLite, so only `query_only(1)` — which a statement can turn off — stands between a
  read verb and a write. Measured by the round-4 auditor: `PRAGMA query_only=0` + `INSERT` succeed,
  and a WAL file with pending frames was rewritten and its `-wal` removed by one open and close.
  `open_test.go:72-73` credits `mode=ro` for a refusal only `query_only` provides.
- **A path containing `?` names a different file.** `Open` opens `path + writeDSN` (`open.go:43`).
  Measured by the round-7 auditor: for a store at `…/a?b/catalogue.db` the write DSN created
  `…/a` (8192 bytes). The same truncation applies to `OpenRead` today.

**Objective.** A catalogue of any older shape — including one mis-stamped 2, or left partial at
stamp 0 by an interrupted first open — is rebuilt once and the reader is told it must be
backfilled; a file that is not recognised as a catalogue is never written; every connection opens
the file it was given; the read path is read-only in fact and never queries a shape it does not
know.

**Success criteria.**

1. `UserVersion` 2 → 3. A file recognised as an older catalogue (§III.1) is rebuilt by `Open`:
   every table, view and index dropped, `Schema` applied, `meta.rebuilt_at`/`meta.rebuilt_from`
   written, stamped 3.
2. **A file recognised as foreign or newer — by `Open` or `OpenRead` — is never written:** its
   database bytes are unchanged; an existing `-wal` or `-journal` is neither checkpointed, rolled
   back, rewritten nor removed. The only footprint allowed is what any SQLite reader leaves beside
   a WAL-mode file that had none: an empty `-wal` and a `-shm` (§II). A rollback-journal file gains
   nothing.
3. **Every connection — read and write — opens exactly the path it was given**, including a path
   containing `#` or `%`, a relative path, and — on Unix, since Windows forbids `?` in file and
   directory names — a path containing `?`.
4. A store stamped newer than 3 is recognised as NEWER from its table names alone and gets the
   version-skew refusal.
5. Exactly one rebuild under N concurrent openers of one old catalogue; none for N concurrent
   openers of a fresh path.
6. **A rebuilt store is loud until backfilled.** Every read verb — `agents`, `touched`, `sql`,
   `find`, `session` — prints one warning on stderr while `backfilled_at` is absent or strictly
   older than `rebuilt_at`; `backfill` clears it.
7. Classification and every read happen on a connection that is read-only in fact: `PRAGMA
   query_only=0` followed by a write fails with a read-only error. That connection sees a
   writer's uncheckpointed frames, so a live store is classified by its true state.
8. `OpenRead` never queries a shape it does not write, and its refusal never promises a rebuild
   that `Open` would refuse (it applies the same recognition, on the same kind of connection).
9. The newer-shape refusal no longer tells users to delete the store; messages say "stamped N".
10. No test in `plugins/gray-area/tools` opens the real store (§V.1 traces it).
11. On real data (§V.3): the preserved broken store is refused in words — by an absolute and by a
    relative `--store` — rebuilt once by `backfill`, queryable after, and warning-free; a real
    rollback-journal file whose only table is `act` is refused with its bytes unchanged and no side
    file created.

## II. Technical Context

- Module `plugins/gray-area/tools` (Go 1.25.13, `modernc.org/sqlite` v1.57.0). The write DSN
  (`open.go:32-36`) sets `_txlock=immediate`, `busy_timeout(5000)` and the PERSISTENT
  `journal_mode(WAL)`; modernc applies every `_pragma` at connection open (`conn.go:151`,
  `sqlite.go:433`), and parses the query of a `file:` DSN too (`_txlock` included). So both DSNs
  can travel in a `file:` URI's query, and nothing may be classified on a write-DSN connection.
- **What a `file:` `mode=ro` connection does and does not do** (measured by the round-4 to round-7
  auditors, fixtures under `~/scratch/plan-audit-rodsn/`): it refuses every write, including after
  `PRAGMA query_only=0`; it cannot checkpoint, so a WAL file with pending frames keeps its database
  bytes and its `-wal`; it SEES those pending frames, so a catalogue whose writer is still open, or
  one rebuilt but not yet checkpointed, or a foreign file whose schema is only in its `-wal`, is
  read in its true state; a rollback-journal file gains no sidecar; a cleanly closed WAL file with
  no sidecars gains an empty `-wal` and a 32 KiB `-shm` that remain after close. A read-only
  connection cannot roll back a hot journal (SQLite returns a read-only error).
- **Why not `immutable=1`** (revision 6): it ignores the `-wal` entirely and misclassified live,
  rebuilt-uncheckpointed and WAL-only-schema files (round-6 fixtures C–F). Correct classification
  needs the frames, so `mode=ro`, with its sidecar footprint accepted (criterion 2).
- **The URI builder has a precedent here.** frank-exchange-of-views' `recordsql.dsnFor`
  (`plugins/frank-exchange-of-views/tools/internal/record/recordsql/store.go:44-89`) builds its
  `file:` DSN with `net/url` (a hand-rolled escape once truncated paths at `?`/`#`), carries
  `_txlock` and the pragmas in `RawQuery`, and renders a Windows drive path as `file:///C:/…`. It
  prefixes `/` to any path that is not rooted, which would anchor a RELATIVE path at `/`; its only
  production caller passes an absolute path (`record/store.go:46`, `filepath.Join(abs, dbName)`),
  so it is unaffected. `telepathy --store` is taken as typed (`root.go:105`), so gray-area's
  `uriFor` calls `filepath.Abs` first. After `Abs`, a path's shape is OS-specific: on Unix
  `//srv/share/x.db` becomes `/srv/share/x.db` and `C:\Users\x.db` is a relative name; only
  Windows produces `file:///C:/…` or a UNC `file:////server/share/…` (round-7 auditor, measured on
  Linux).
- A read-decide-stamp transaction on the write connection is feasible: `db.Begin()` is `BEGIN
  IMMEDIATE` by DSN; `Schema` runs as one multi-statement `Exec`; `user_version` writes are
  transactional (confirmed by six auditors).
- **What a catalogue looks like.** Table names across its history (each `schema.go` version):
  `e111c79a` act, file_offset, provisional_skip, session, thought, word; `bd478a05` adds meta;
  `28ad35e5` act, file_offset, meta, session, skip, thought, word. Every schema creates `session`
  FIRST; `session` has had `session_id, project_dir, cwd, closed_at, capture_build` in every version
  (#867 renamed its other two). Every past shape change added a table.
- **Stamp 0 is not proof of a fresh file.** Today's `Open` runs `Schema` and the stamp as separate
  autocommit statements (`open.go:60-67`); a first open killed part-way leaves a prefix of the
  tables — always including `session` — at stamp 0.
- **The generic-name hazard is real on this box.** `~/scratch/daemon-bench/def.db` is a
  rollback-journal SQLite file (header `01 01`) whose only table is `act`.
- The shape-1 DDL is `git show 28ad35e5^:plugins/gray-area/tools/internal/catalogue/schema.go`.
- **Why bump to 3 with no DDL change.** The stamp 2 is untrustworthy: old stores opened by a
  post-`28ad35e5` binary were stamped 2 with shape-1 tables.
- **What a rebuild costs.** The store is derived (`schema.go:16-19`) and reprojected by `telepathy
  backfill` — only from transcripts that still exist: the store keeps 30 days (`sweep.go:28`), the
  client's default transcript retention is also 30 days. On this box no retention setting is
  configured and the oldest transcript was last modified 2026-08-16 (`find ~/.claude/projects
  -name '*.jsonl' -printf '%TY-%Tm-%Td\n' | sort | head -1`, run 2026-09-10).
- **The rebuild notice is invisible from hooks.** A hook's exit-0 stderr goes to the debug log only
  (Claude Code hooks reference), and the first `Open` after an upgrade is almost always a hook's.
  The persisted `meta` markers are the signal (§III.3). `SubagentStop` never opens the catalogue
  (`main.go:488-535`).
- **Clocks.** `catalogue.Open` has no clock injection; `telecli.Env.Now` is frozen in tests
  (`root_test.go:185`). Both markers use the WALL clock; the warning prints an absolute UTC
  timestamp.
- **The golden harness today** (`telecli/golden_test.go`): `h.run` (:57) returns stdout, stderr and
  the exit code, but `TestGoldenOutput`/`TestGoldenHelp` fail on any non-zero exit (:170, :191) and
  `assertGolden` (:98) compares stdout only; `newHarness` (:45) builds a CURRENT store by running
  `backfill`. `find` needs ripgrep, which neither CI runner image ships and the `gray-area` job does
  not install; `TestGoldenFind` skips without it (`golden_test.go:202`). `scripts/golden` covers
  only `plugins/frank-exchange-of-views/tools` (`scripts/internal/goldenmods.Modules`); gray-area's
  goldens are gated by the `gray-area` CI job (`.github/workflows/hooks.yml:652-840`).
- **The tests touch the real store today.** `cmd/gray-area-capture/hook_test.go` SessionStart tests
  (`:167`/`:174`, `:194`/`:197`) drive `sweep`, which opens the catalogue at `DefaultDir()`, without
  redirecting `HOME`/`USERPROFILE`/`XDG_STATE_HOME`. Other paths are isolated or never open the
  store (re-run by seven auditors). `open_test.go`'s `newStore` keeps its writer open until cleanup.
  The real store's mtime moves only when a write is checkpointed, never on a read, and after a
  release it moves whenever this box's hooks write it — so it cannot prove isolation; `strace` is
  installed here (`/usr/bin/strace`) and can.
- **Preserved real artifact.** `~/.claude/scratch/restart-probe/broken-store-2026-09-10/catalogue.db`
  (mtime 2026-09-09 20:58); an empty `-wal`/`-shm` beside it came from an audit open.
- **Version skew after release.** Sessions begun before an update keep the old hook binary, which
  refuses a shape-3 store on every hook — stderr only — with its old *delete it* text.
- There is no `func Benchmark` in `internal/catalogue` today.

## III. Proposed Changes

```
README.md (repository root)                     [MODIFY] :127 backfill is also the remedy after an upgrade rebuild
plugins/gray-area/
├── README.md                                   [MODIFY] store paragraph (:69-78): upgrades may rebuild the store; telepathy warns until backfill
├── skills/telepathy/SKILL.md                   [MODIFY] "What it cannot tell you" (:79): an upgrade rebuilds the store empty; the warning; backfill
└── tools/
    ├── internal/catalogue/schema.go            [MODIFY] UserVersion 3; comments :16-19 and :144-147 (incl. the recognition contract)
    ├── internal/catalogue/open.go              [MODIFY] uriFor/readURI/writeURI; Open(path, notice); recognise(); rebuild; OpenRead via readURI + shape check; messages; comments :38, :52-54, :71-75, :77-80, :104
    ├── internal/catalogue/meta.go              [MODIFY] MetaRebuiltAt, MetaRebuiltFrom, MetaBackfilledAt; MarkRebuilt; MarkBackfilled; RebuildPending
    ├── internal/catalogue/open_test.go         [MODIFY] new cases; BenchmarkOpen; Open's new argument; comment :72-73 corrected
    ├── internal/catalogue/{meta,sweep,ingest}_test.go [MODIFY] Open's new argument; meta cases
    ├── internal/telecli/store.go               [MODIFY] openRead(w io.Writer) warns while a rebuild awaits backfill
    ├── internal/telecli/{agents,touched,sql,find,session}.go [MODIFY] pass cmd.ErrOrStderr() to openRead
    ├── internal/telecli/backfill.go            [MODIFY] Open(env.Store, cmd.ErrOrStderr()); MarkBackfilled on success; Long help (:15-23)
    ├── internal/telecli/root.go                [MODIFY] longDescription "WHAT IT CANNOT TELL YOU" (:66-70)
    ├── internal/telecli/golden_test.go         [MODIFY] store-variant harnesses; assertGoldenFull; the warning table test
    ├── internal/telecli/testdata/shape1.sql    [NEW] the 28ad35e5^ DDL, for the olderShape fixture
    ├── internal/telecli/testdata/golden/       [NEW] sql-older-shape, sql-not-a-catalogue, agents-rebuilt-warning; [MODIFY] help-root, help-backfill
    └── cmd/gray-area-capture/
        ├── catalogue.go                        [MODIFY] openCatalogue passes its stderr
        └── hook_test.go                        [MODIFY] TestMain isolates HOME/USERPROFILE/XDG_STATE_HOME
```

### III.1 URIs, recognition, then `Open(path string, notice io.Writer) (*sql.DB, error)` — contract change

**URIs — every connection this package opens is one.** `uriFor(path, rawQuery string) (string,
error)` mirrors `recordsql.dsnFor` with one correction: `abs, err := filepath.Abs(path)` first;
then `p := filepath.ToSlash(abs)`, prefix `/` when not rooted (a Windows `C:/…`), and
`url.URL{Scheme: "file", Path: p, RawQuery: rawQuery}.String()`.
- `readURI(path)` = `uriFor(path, strings.TrimPrefix(readDSN, "?"))` — `mode=ro`, `query_only(1)`,
  `_defensive=1`, `busy_timeout(2000)`.
- `writeURI(path)` = `uriFor(path, strings.TrimPrefix(writeDSN, "?"))` — `_txlock=immediate`,
  `busy_timeout(5000)`, `journal_mode(WAL)`, `synchronous(NORMAL)`, `foreign_keys(1)`, in the same
  order as today (busy_timeout before WAL, `open.go:28-31`).
- `open.go:43`'s `path+writeDSN` and `:86`'s `path+readDSN` are both replaced; no bare-path DSN
  remains in the package.

**`recognise(q querier) (v int, class Class, tables []string, err error)`**, shared by `Open` and
`OpenRead`. It reads `PRAGMA user_version` → `v`, the user tables `T` (`sqlite_master` type
`table`, name not `sqlite_%`) and, when needed, `session`'s columns. With `K` = {act, file_offset,
meta, provisional_skip, session, skip, thought, word} and `S` = {session_id, project_dir, cwd,
closed_at, capture_build}, **in this order**:

1. `T` empty → FRESH.
2. `v > UserVersion` → NEWER if {session, file_offset} ⊆ `T` (table names only — a future shape may
   add tables or rename `session`'s columns), else FOREIGN.
3. `session` ∉ `T`, or `session`'s columns ⊉ `S` → FOREIGN.
4. `T` ⊄ `K` → FOREIGN.
5. `v == 0` → OLDER (a complete or partial catalogue left unstamped).
6. `0 < v < UserVersion` → OLDER if `file_offset` ∈ `T`, else FOREIGN.
7. `v == UserVersion` → CURRENT if `file_offset` ∈ `T`, else FOREIGN.

Residual hazard (R2): a foreign file with a `session` table carrying all five columns, only
catalogue table names, and a stamp ≤ 2.

**`Open`:**
1. File absent → FRESH; go to step 3.
2. **Pre-classify read-only:** `sql.Open("sqlite", readURI(path))`, `recognise`, close. It sees any
   pending frames. FOREIGN → refuse: *`<path>` is not a gray-area catalogue (tables: …) — refusing
   to touch it; check `--store`.* NEWER → refuse: *`<path>` is stamped v by a newer gray-area (this
   binary writes UserVersion) — usually a session started after a plugin update while this one
   still runs the older binary; it resolves when this session restarts. Do not delete the store.*
   An error is returned as an error. In none of these cases is a write connection opened.
3. Only for FRESH, OLDER or CURRENT: `sql.Open("sqlite", writeURI(path))` — the same file step 2
   read; `db.Begin()` (IMMEDIATE); **`recognise` again** inside the transaction — the
   authoritative answer, since a concurrent opener may have rebuilt it (FOREIGN or NEWER now → roll
   back, refuse); then:
   - OLDER → drop every non-`sqlite_%` view, table and index (`DROP … IF EXISTS`, views first);
     apply `Schema`; `MarkRebuilt(tx, v, time.Now())`; stamp; mark `rebuilt`;
   - FRESH or CURRENT → apply `Schema`; stamp;
   - commit; if `rebuilt`, one line to `notice`: *catalogue: `<path>` was stamped v, rebuilt empty
     at shape UserVersion — run `telepathy backfill` to re-read the corpus*.

The window between step 2 and step 3 is the only place a foreign file could meet the write
connection: a catalogue replaced by a foreign file between two opens of the same path (R6).

Every error path rolls back and closes. `notice` may be `io.Discard`.

**Consumer census — `Open`** — `grep -rn -E '(catalogue\.)?\bOpen\(' plugins/gray-area/tools --include='*.go' | grep -v -E 'sql\.Open|os\.Open|OpenRead|OpenFile'`, run 2026-09-10 in this worktree — **14 lines** (matched by the round-4 to -7 auditors):

| Site | Changes? |
|---|---|
| `cmd/gray-area-capture/catalogue.go:31` `openCatalogue(stderr)` | Passes its `stderr` (debug log only — the signal is §III.3) |
| `internal/telecli/backfill.go:29` | Passes `cmd.ErrOrStderr()` |
| `internal/catalogue/open_test.go:15,179` | `io.Discard`, or a buffer where the notice is asserted |
| `internal/catalogue/sweep_test.go:14,182` | `io.Discard` |
| `internal/catalogue/meta_test.go:16` | `io.Discard` |
| `internal/catalogue/ingest_test.go:81,120,156,192,208,254` | `io.Discard` |
| `internal/catalogue/open.go:39` | The definition |

### III.2 `OpenRead` — behaviour change, signature unchanged

After the existing missing-store check: `sql.Open("sqlite", readURI(path))`, capped at one
connection as today (`PinnedConn`), and `recognise` on it — it sees a live writer's frames, so a
healthy store opened while hooks write it is CURRENT. CURRENT → return it. Otherwise close and
return:
- OLDER → *`<path>` is stamped v; this binary reads shape UserVersion. The next session start or
  turn end (gray-area's `SessionStart`, `Stop` and `SessionEnd` hooks) rebuilds it, or run
  `telepathy backfill`.* — true by construction: `Open` classifies the same file OLDER on the same
  kind of connection.
- NEWER → the §III.1 skew text. FOREIGN → the not-a-catalogue text.
- FRESH (an empty file) → *`<path>` holds no catalogue yet — let a session run or run
  `telepathy backfill`.*

**Consumer census — `OpenRead`/`openRead`** — `grep -rn -E 'OpenRead\(|openRead\(' plugins/gray-area/tools --include='*.go'`, run 2026-09-10 — **15 lines** (matched by the round-4 to -7 auditors):

| Site | Changes? |
|---|---|
| `internal/catalogue/open.go:81` | The definition |
| `internal/telecli/store.go:13,17` (`Env.openRead`) | → `openRead(w io.Writer)` (§III.3) |
| `internal/telecli/agents.go:24`, `touched.go:23`, `sql.go:60`, `find.go:79`, `session.go:22` | Pass `cmd.ErrOrStderr()`; each surfaces the refusals and the warning (§V.2). No test calls `openRead` directly |
| `cmd/gray-area-capture/catalogue_test.go:54`; `internal/catalogue/sweep_test.go:189`; `open_test.go:77,106,131,155` | No code change. Each reads a store `Open` created; in `open_test.go` the writer is still open (`newStore`), and `readURI` sees its frames, so each classifies CURRENT — pinned by a new test (§V.1). `open_test.go:77`'s write-refusal test now holds because of `mode=ro`, not only `query_only` |
| `internal/catalogue/open_test.go:187` | No — the missing-store case |

**Consumer census — the read-only claim** — `grep -rn -E 'read-only|query_only|mode=ro' README.md plugins/gray-area --include='*.md' --include='*.go' --include='*.golden' --include='*.json' | grep -v '_test.go'`, run 2026-09-10 — **10 lines**:

| Site | Changes? |
|---|---|
| `plugins/gray-area/README.md:55` (`telepathy sql … read-only`) | No — becomes true |
| `README.md:127` (root: "read-only, `query_only`, defensive, and refused `ATTACH`") | The `backfill` sentence changes (§III.4); the read-only clause becomes true |
| `plugins/gray-area/skills/telepathy/SKILL.md:53` (`sql` is read-only) | No — becomes true |
| `internal/telecli/sql.go:41` (Short), `help-root.golden:34` | No — becomes true |
| `internal/telecli/sql.go:48` (Long), `help-sql.golden:12` | No — becomes true |
| `internal/catalogue/open.go:75` (`readDSN`) | Unchanged value; now only reached through `readURI` |
| `internal/catalogue/open.go:88` (the "open read-only" error) | Unchanged |
| `internal/catalogue/open.go:104` (*mode=ro and query_only already refuse writes*) | Rewritten (§III.4) |

`open_test.go:72-73` is excluded by the grep's `_test.go` filter and is listed in §III.4.

### III.3 The rebuilt-store warning

- `meta.go` gains `MetaRebuiltAt = "rebuilt_at"`, `MetaRebuiltFrom = "rebuilt_from"`,
  `MetaBackfilledAt = "backfilled_at"` beside `MetaRetainedOn`; `MarkRebuilt(tx, from int, at
  time.Time)`; `MarkBackfilled(db, at time.Time)`; and
  `RebuildPending(db) (rebuiltAt time.Time, rebuiltFrom int, pending bool, err error)`: pending iff
  `rebuilt_at` exists AND (`backfilled_at` absent OR `backfilled_at` < `rebuilt_at`, strictly).
  Both are Unix seconds from the wall clock; a backfill finishing in the rebuild's second clears it.
- `Env.openRead(w io.Writer)`: `OpenRead`, then `RebuildPending`:
  - pending → one line to `w`: *telepathy: the store was rebuilt at <UTC RFC3339> (it was stamped
    N) and has not been backfilled since — sessions from before then are missing; run
    `telepathy backfill`*;
  - error → one line: *telepathy: could not read the store's rebuild markers (<err>) — whether it
    needs a backfill is not measured*;
  - the verb then runs either way, exit code unchanged.
- `backfill` calls `MarkBackfilled(db, time.Now())` after completing without error.

### III.4 Comments, help and prose that speak the old model

- `schema.go:16-19` → an older catalogue is REBUILT by `Open` and must be backfilled; a
  non-catalogue is never written; a newer stamp is refused. **The recognition contract:** every
  future shape MUST keep tables named `session` and `file_offset` (a newer stamp is recognised by
  those names alone); a stamp ≤ 3 is recognised by `session`'s columns ⊇ `S`.
- `schema.go:144-147` (`UserVersion`) → why 3: the stamp 2 was written onto shape-1 stores.
- `open.go:38` `Open` doc and `:52-54` → pre-classify on the read-only URI, decide inside the
  IMMEDIATE transaction on the write URI, and why (the WAL pragma is persistent; a write connection
  must never meet an unrecognised file); the sidecar footprint of a read-only reader.
- `open.go:71-75` (`readDSN` doc) and `:104` → both DSNs reach SQLite only inside a `file:` URI
  (modernc truncates a bare path at `?`); cite `recordsql.dsnFor` and the `filepath.Abs`
  correction.
- `open.go:77-80` (`OpenRead` doc: *READ-ONLY for the `sql` verb*) → it serves every read verb, is
  read-only through `readURI`, and classifies the file, refusing any class but CURRENT.
- `open_test.go:72-73` → the refusal comes from `mode=ro` (via `readURI`) and `query_only`.
- `telecli/root.go:66-70` `longDescription` → after a gray-area upgrade the store may have been
  rebuilt empty; `telepathy` warns until `telepathy backfill`. `help-root.golden` changes.
- `telecli/backfill.go:15-23` Long → also the remedy after a rebuild; clears the warning.
  `help-backfill.golden` changes.
- `plugins/gray-area/README.md` store paragraph and `skills/telepathy/SKILL.md` (:79) → an empty or
  thin result right after an upgrade is not evidence; `telepathy` warns until backfilled.
- Repository-root `README.md:127` → *`telepathy backfill` is the one explicit cold read — for a
  corpus that predates capture, and after an upgrade rebuilds the store*.
- `hooks.json` `_comment` mentions "reprojection" only as a cost comparison — unchanged.

### III.5 Test isolation

`cmd/gray-area-capture/hook_test.go` gains `TestMain(m *testing.M)`: a temp dir as `HOME`,
`USERPROFILE` and `XDG_STATE_HOME` for the whole package, removed afterwards. Tests that already
`t.Setenv` keep overriding it.

### III.6 Golden harness

- `newStoreHarness(t, variant)` builds the store the verbs run against: `olderShape` — the DDL in
  `testdata/shape1.sql` applied to an empty file, stamped 1; `foreign` — one unrelated table;
  `rebuiltPending` — `newHarness` as today (its `backfill` writes `backfilled_at`), then the test
  DELETES `backfilled_at` and writes `rebuilt_at`/`rebuilt_from` at a fixed instant.
- `assertGoldenFull(t, name, code, stdout, stderr)` writes and compares one file:
  `exit: <code>\n--- stdout\n<stdout>--- stderr\n<stderr>` (scrubbed as today). The three new
  goldens use it; existing goldens keep `assertGolden`.
- `TestTheRebuildWarningReachesEveryReadVerb`: on `rebuiltPending`, each of `agents`, `touched
  <path>`, `sql "SELECT 1"`, `find <term>` and `session <id>` prints the warning on stderr exactly
  once. Each exits 0 — except `find` when `exec.LookPath("rg")` fails, where it must instead exit
  non-zero with the existing no-ripgrep refusal (the warning is still asserted, since `openRead`
  runs first). After `backfill`, no verb prints it.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation |
|---|---|---|---|
| R1 | Every existing catalogue rebuilds once on upgrade; partial until a backfill; rows whose transcripts were removed do not come back | certain × low × low | Accepted for a derived store; LOUD — every read verb warns until `backfill`; help, skill, both READMEs say so |
| R2 | `Open` or `OpenRead` alters an unrelated SQLite file (a mistyped `--store`) | low × high × low | Classification only on a real `mode=ro` connection, which cannot write, checkpoint or roll back, and sees pending frames; no write connection unless that says FRESH, OLDER or CURRENT; FOREIGN requires missing the historical `session` table with five columns; tests with rollback, clean-WAL, pending-frames and WAL-only-schema fixtures, `{act}` and `{meta}` alone; §V.3 on the real `def.db`. Residuals, accepted and stated: (a) a WAL-mode foreign file that had no sidecars gains an empty `-wal` and a `-shm`, as from any SQLite reader — its database bytes unchanged; (b) a foreign file with a `session` table carrying all five columns, only catalogue names and a stamp ≤ 2 would be rebuilt — judged very unlikely |
| R3 | A future shape is called FOREIGN during version skew | low × medium × low | NEWER decided on table names alone before any column check; the contract is written into `schema.go`; tests with an extra table and a renamed `session` column |
| R4 | A partial first-open catalogue at stamp 0 is refused forever | low × medium × low | Stamp 0 + historical `session` + only catalogue names → OLDER; test with {session} alone |
| R5 | Two concurrent openers both rebuild | medium × medium × low | Re-classification inside one IMMEDIATE transaction; concurrent test under `-race` |
| R6 | The file is swapped between the read-only pre-classification and the write connection | very low × high × none | Accepted: a catalogue replaced by a foreign file in the milliseconds between two opens of the same path. The in-transaction re-classification still refuses before any DDL, but the write URI's WAL pragma will have converted the swapped-in file |
| R7 | The warning never clears, or clears wrongly | low × medium × low | One wall clock for both markers; strict comparison; equal-second case tested |
| R8 | Old binaries in running sessions refuse the shape-3 store with "delete it" | certain for a window × medium × low | stderr only, never blocking; the new text stops saying it for the next skew |
| R9 | The test suite opens the developer's real store | was certain × high × low | `TestMain` isolation; §V.1 traces every file the test binaries open |
| R10 | A hook stalls on the IMMEDIATE lock, or the extra read-only connection adds cost | low × low × low | `busy_timeout(5000)`; `BenchmarkOpen` reports `Open`'s cost on a current store (reported, not asserted). **Added after PR #894's Windows CI run:** SQLite does not apply `busy_timeout` to a lock upgrade, and the write URI's `journal_mode(WAL)` pragma on a not-yet-WAL file is one (shared lock to read the header, exclusive to convert) — 8 concurrent fresh openers failed in 0.17s with `SQLITE_BUSY`. `Open` therefore retries its classify-and-write phase on `SQLITE_BUSY` (extended codes included), jittered backoff, a 5s budget matching `busy_timeout`; `TestConcurrentOpenersRebuildOnce` on the Windows leg is the check |
| R11 | A connection opens a different file than the path given (`?`, `#`, `%`, relative, Windows) | was certain for `?` × high × low | Every connection goes through `uriFor` (`filepath.Abs`, then `net/url`, as `recordsql.dsnFor`); a test drives `Open` (creating, then pre-classifying) and `OpenRead` on paths containing `#`, `%` and — on Unix only, as Windows reserves it — `?`, and on a relative path under `t.Chdir`, asserting only the named file exists afterwards; per-OS URI tests pin both `writeURI` and `readURI` (§V.1); the Windows CI leg runs the suite |

## V. Verification Plan

1. **Unit and race** — re-arms on any change under `plugins/gray-area/tools`:
   `go -C plugins/gray-area/tools vet ./...`; then, traced,
   `strace -f -e trace=open,openat -o ~/.claude/scratch/restart-probe/test-trace.txt go -C plugins/gray-area/tools test -count=1 ./...`
   and `grep -c "${XDG_STATE_HOME:-$HOME/.local/state}/special-circumstances" ~/.claude/scratch/restart-probe/test-trace.txt`
   MUST print `0` — no test binary opened the real store (at the location `DefaultDir` resolves,
   whether or not `XDG_STATE_HOME` is set), whatever else wrote it meanwhile; then
   `CGO_ENABLED=1 go -C plugins/gray-area/tools test -count=1 -race ./...`. New tests in
   `open_test.go` and `meta_test.go`:
   - **URIs, per `runtime.GOOS` — a table in which EVERY input row is run through BOTH `readURI`
     and `writeURI`**, asserting for each that `url.Parse(got).Path` equals the expected `r` and
     that `RawQuery` equals that constructor's exact query (below). Rows: a relative path → `r` is
     `filepath.ToSlash(abs)` with a leading `/` added when it lacks one — `file:///home/…/rel.db`
     on Unix, `file:///C:/…/rel.db` on Windows — never `file:///rel.db`; `%`, `#` (and on Unix `?`)
     in a path component → percent-escaped in the URI; on
     Unix, `//srv/share/x.db` → `file:///srv/share/x.db` (the `//` collapses under `Abs`); on
     Windows only, `C:\Users\x.db` → `file:///C:/Users/x.db` and `\\server\share\x.db` →
     `file:////server/share/x.db`. The relative-path case compares `url.Parse(got).Path` with `r`
     (not the raw string), so a working directory containing a space, `#` or `%` — `t.TempDir` keeps
     them from the test name — still compares equal. `writeURI` carries the write DSN's query in the
     same order; `readURI` carries exactly `mode=ro&_pragma=query_only(1)&_defensive=1&_pragma=busy_timeout(2000)`
     on the same escaped path.
   - **The named file, and only it — write side AND read side:** subtests for a directory named
     `a#b`, `a%41b` (a `%` followed by two hex digits — the only form SQLite would decode, so an
     unescaped URI would open `aAb` and the subtest can fail), `a?b` (skipped when
     `runtime.GOOS == "windows"`, with the reason in `t.Skip`: Windows reserves `?` in file and
     directory names), and **a relative path**: `t.Chdir(<tmp>)` (Go ≥ 1.24; the module is 1.25.13)
     and the path given as `a#b/catalogue.db`. In each: (1) `Open` creates the store; close it.
     (2) `Open` again — the file now exists, so this runs the read-only pre-classify through
     `readURI` — must succeed and write NO notice, and `meta` must hold no `rebuilt_at` (that is how
     the test knows the class was CURRENT, since `Open` returns a bare `*sql.DB`); close it. (3)
     `OpenRead` on the same path must succeed (it refuses every class but CURRENT); close it. After
     each close, the only files anywhere under `<tmp>` MUST be `<tmp>/<dir>/catalogue.db` and, at
     most, its `-wal` and `-shm`: a truncated or mis-anchored read or write would have created or
     read another file (e.g. `<tmp>/a`, `<tmp>/aAb`, `/a#b/catalogue.db`). On `readURI`, `PRAGMA query_only=0` then an `INSERT` fails with a
     read-only error.
   - **Recognition:** FRESH (no tables); CURRENT; OLDER for a shape-1 catalogue stamped 1, a stamp-2
     catalogue with shape-1 tables (this box's state), a complete shape-1 catalogue at stamp 0, and
     a stamp-0 file holding only the shape-1 `session` table; NEWER for stamp 4 with today's tables,
     with an extra unknown table, and with a `session` column renamed; FOREIGN for `{act}` alone and
     `{meta}` alone at stamp 0, an unrelated table at stamp 1, a `session` table lacking one of the
     five columns, and stamp 4 without `file_offset`.
   - **Live stores classify by their true state:** `OpenRead` on a CURRENT store whose writer is
     still open (uncheckpointed frames) returns it; a store rebuilt by `Open` whose writer is still
     open is CURRENT to `OpenRead` and triggers the warning, not an OLDER refusal.
   - **No write on refusal**, for both `Open` and `OpenRead`, on four FOREIGN fixtures: a
     rollback-journal file (header `01 01`) — database sha256 unchanged and NO sidecar created; a
     cleanly closed WAL file — database sha256 unchanged, and the only new files are an empty `-wal`
     and a `-shm`; a WAL file with pending frames — database sha256 and `-wal` sha256 unchanged; a
     WAL file whose tables exist ONLY in its `-wal` — classified FOREIGN (not FRESH), database and
     `-wal` sha256 unchanged. The same for a NEWER fixture.
   - **Behaviour:** each OLDER case is rebuilt (notice once, stamp 3, `rebuilt_from` recorded,
     `v_session` queryable); NEWER refuses with the skew text (no "delete it"); CURRENT keeps its
     rows, no notice; N goroutines on one shape-1 catalogue → one notice, and a row inserted after
     the first return survives; N goroutines on one fresh path → no notice, no `rebuilt_at`;
     `OpenRead` returns the class-specific refusal for OLDER, NEWER, FOREIGN and FRESH, never a SQL
     error.
   - **`RebuildPending`:** no markers → not pending; `rebuilt_at` only → pending with `rebuiltFrom`;
     `backfilled_at` newer → not pending; older → pending; equal → not pending; a corrupt
     `rebuilt_at` → error.
   - **Benchmark** (reported, not asserted): `BenchmarkOpen` on a current store —
     `go -C plugins/gray-area/tools test -run '^$' -bench BenchmarkOpen ./internal/catalogue`.
2. **Goldens and the warning** — re-arms on `telecli` output, help, or the harness. Run uncached
   (memory *golden-test-cache-trap*): `go -C plugins/gray-area/tools test -count=1 ./internal/telecli
   -run 'TestGolden|TestTheRebuildWarningReachesEveryReadVerb'` — `sql-older-shape` (exit 1,
   refusal on stderr, `olderShape` store), `sql-not-a-catalogue` (exit 1, `foreign` store),
   `agents-rebuilt-warning` (exit 0, table on stdout, warning on stderr, `rebuiltPending` store), all
   through `assertGoldenFull`; updated `help-root` and `help-backfill`. `scripts/golden` does NOT
   cover gray-area; the CI gate for these goldens is the `gray-area` job
   (`.github/workflows/hooks.yml:652-840`) on Linux and Windows.
3. **Real data** — re-arms on `open.go`, `meta.go`, `store.go` or `backfill.go`. Build `telepathy`;
   let `$P=~/.claude/scratch/restart-probe/shape-check`.
   - **Relative path first, on a fresh copy:** copy
     `~/.claude/scratch/restart-probe/broken-store-2026-09-10/catalogue.db` (the `.db` alone; its
     `-wal` is empty) to `$P/rel/catalogue.db`; record its `sha256sum`; `cd $P/rel && telepathy
     --store catalogue.db sql "SELECT count(*) FROM v_session"` MUST print the worded OLDER
     refusal, exit 1, with the sha256 unchanged.
   - **The broken store, absolute:** a second fresh copy at `$B=$P/abs/catalogue.db`; `telepathy
     --store $B sql "SELECT count(*) FROM v_session"` MUST print the worded OLDER refusal, exit 1,
     with `$B`'s sha256 unchanged; `telepathy --store $B backfill` MUST print the rebuild notice once
     and succeed; `telepathy --store $B agents` MUST print no rebuild warning; `telepathy --store $B
     sql "SELECT count(*) FROM v_session"` MUST return a count; a second `backfill` MUST print no
     notice.
   - **The adversarial foreign file:** copy `~/scratch/daemon-bench/def.db` (rollback journal,
     header `01 01`, sole table `act`) to `$P/foreign.db`; record its `sha256sum` and `od -An -tx1
     -j18 -N2`; `telepathy --store $P/foreign.db backfill` and `telepathy --store $P/foreign.db
     agents` MUST each refuse naming the table `act`; the `sha256sum` and header MUST be unchanged;
     no `foreign.db-wal`/`-shm`/`-journal` MUST appear. The original is never opened.
   The live store is not touched.
4. **Parity** — `go -C scripts run ./pluginparity`; no `plugin.json` change.
5. **Auditor gate** — `/prosthetic-conscience:plan-audit plans/catalogue-shape-rebuild.md` → PASS
   before implementation.
