# migrate publishes its destination whole and synced

> STATUS: approved 2026-10-15 (audit: two consecutive PASSes, rounds 4 and 5), not started. Asked for by gblock on 2026-10-14 ("Plan it", then
> "build it properly"), after `faster-test-loop.md`'s follow-ups left `internal/record/migrate` as
> the local suite's floor. Audited in five rounds as `plans/migrate-bulk-load.md`, which paired this
> with a `synchronous=OFF` bulk mode. Round 5 found that bulk mode's reason obsolete — since #1379
> (`ce1f1be5`) `record.Append` validates inside its own write transaction and the store writes
> several acts in one caller-held transaction — and gblock ruled on 2026-10-15: ship this part now,
> and re-plan the speed half as batched replay through that path (see "Not in this plan").
>
> Rulings this plan carries: refuse a separated destination; Unix publishes by `rename(2)` over an
> existing empty `--to`, Windows removes it then renames; no automatic reaping of leftover stages;
> directory-entry durability is best-effort and a failure is a warning; macOS and Windows
> power-loss durability are stated residues; the Unix mount-point check is stated as untested.
> Audit round 1 of this plan FAILED on one gap (the stage's path leaking into published text);
> gblock ruled its three forks on 2026-10-15: no mode copy onto the stage, no working-directory
> refusal, no removal of created parents — each a stated residue. Rounds 2 and 3 FAILED on local
> gaps in the same class (text built with the stage's path reaching the operator: warnings, the
> error chain, the manifest returned to `--json`); round 4 PASSED with notes, folded in here. The
> speed half is #1382.

Paths are relative to `plugins/frank-exchange-of-views/tools/` unless they begin with `scripts/`,
`plans/` or `.github/`. Line numbers were taken at `41a815fc`; nothing this plan touches moved by `42b8dd18`, `origin/main` on 2026-10-15, except `difftest`'s manual golden (#1380), which step 6 regenerates after a rebase. Before any
command below that runs tests or the binary against a run, clear the seat environment:
`unset $(env | grep -o '^FEOV_[A-Z_]*')`.

## I. Summary & Goals

### The problem

`migrate.Migrate` writes straight into `--to`. A migration that is killed, or loses power, leaves a
partial run at `--to` that opens and reads like a run. Of what it writes, only the database is
synced (one fsync per event, `synchronous=FULL`); the copied channels, the translated
`class-registry.json` and `inputs/migration.json` are never synced at all.

It is also slow, and the slowness is that per-event fsync: on 2026-10-14, interleaved, on disk, the
test that migrates each of the 21 archived runs once took 49.95 s and 27.46 s as today, 15.49 s and
14.40 s with `synchronous=OFF` (a prototype, reverted), and 14.6 s with TMPDIR on tmpfs. That is the
re-planned speed half's problem; this plan is what makes any faster write safe to publish.

### Success criteria

| # | Criterion | Baseline | Target |
|---|---|---|---|
| M2 | A migration interrupted at any point — killed, or (Unix) power lost — leaves `--to` as it was | a partial run at `--to` | on Unix `--to` exactly as it was (absent, or the empty directory it was); on Windows also possibly absent in the instant between removing an empty `--to` and the rename. The partial work is in a `.<base>.migrating-<random>` sibling, which nothing removes automatically (ruling, 2026-10-15): the help names it as the tool's and says to delete it by hand |
| M3 | When `Migrate` returns (success, or the refusal path that publishes), the published run's FILES are on disk, and its directory entries are synced where the platform allows | the database synced per commit; no other file synced | Every regular file of the run fsynced after its last write — a failure is an error. Then directory syncs, BEST-EFFORT by ruling (2026-10-15): every directory of the run, then after the rename every directory whose entries `Migrate` changed (`--to`'s parent, each parent `Migrate` created, the nearest pre-existing ancestor). A directory sync that fails is a warning naming the directory, never an error. Linux on a local filesystem: expected to succeed and tested. macOS: attempted, `F_FULLFSYNC` semantics for directories undocumented, not run by CI — stated residue. Windows: not attempted; the rename is not write-through — stated residue |
| M4 | The migrated run is what today's code writes, sealed | — | every archive: event rows and body tables identical; the recursive listing with modes (`find <to> -printf '%M %P\n' \| sort`) identical except that the branch's record is sealed — no `records/record.db-wal` or `-shm`, which today's verb leaves because its process exits without closing |
| M5 | The live write path and every other opener of a record | — | untouched: nothing changes outside `migrate`, `recordroot.go`'s predicate and the verb |

### Value case

Against doing nothing: a migration that can be interrupted without leaving a half-run at `--to`
that reads like a real one, and a published run whose files are on disk when the tool says it is
done. Against writing in place with better syncing: only a stage makes the interrupted case clean.
The speed half (#1382) does not depend on it: batched replay keeps `synchronous=FULL`.

## II. Technical Context

- Go 1.27 (`go.mod`: `go 1.27.1`), `modernc.org/sqlite`, WAL mode. Cached pools per absolute path
  (`openCache`, `internal/record/recordsql/store.go:125`); `CloseUnder` (`store.go:158`) closes
  every cached handle under a directory, and closing the last connection checkpoints the WAL into
  the database and deletes it.
- SQLite's sync behaviour is not relied on for the publish: after every handle is closed, this
  plan fsyncs every file itself. An `fsync` covers every dirty page of a file, whoever wrote it.
  (Round 1 established that a later `wal_checkpoint` finds nothing to sync once a close has folded
  the WAL; the strace is at `~/.claude/scratch/migrate-bulk-load-audit/checkpoint-after-close-fsync-trace.txt`.)
- Go's `os.Rename` Lstats the destination and returns EEXIST for any directory before the syscall
  (`os/file_unix.go:25-44`, go1.27.1), so replacing an empty directory needs `rename(2)` itself.
- `record.Migrating` (`record.go`, set in `migrate/replay.go:96-99`) is untouched.
- Repository rules that bind this: `migrate` is the only thing that upgrades a database and is
  self-contained; nothing is added to the live record for migration's sake.

## III. Proposed Changes

```
internal/record/
  recordroot.go         [MODIFY]  a pure exported predicate, "the record root is set", which RecordsDir uses
internal/record/migrate/
  migrate.go            [MODIFY]  refusals, stage, sync, publish
  publish.go            [NEW]     --to checks, stage creation, syncTree, publish
  publish_unix.go       [NEW]     //go:build unix: renameDir as rename(2) with EINTR retry; directory fsync; same-device check
  publish_windows.go    [NEW]     renameDir as RemoveDirectory then os.Rename; no directory fsync
  publish_test.go       [NEW]
  migrate.go also refuses a --to inside --from (found in review, #1386)
internal/record/recordroot_test.go [MODIFY]  RecordRootSet agrees with RecordsDir
internal/cli/migrate.go [MODIFY]  help sentence, --to flag text, warnings printed to stderr
internal/cli/migrate_test.go [NEW]  the verb's warning printer writes each warning to stderr, with and without --json
internal/difftest/testdata/manual_runs_every_help_page_live.golden [REGENERATE]
```

### The change

0. **What `--to` may be, decided before any work.** `--to` is cleaned and made absolute. It must be
   absent, or a real directory (by `os.Lstat`: not a symlink) that is empty. A regular file, a
   symlink, or a non-empty directory is refused with what was found. Today a regular file passes
   `os.ReadDir`'s check and is refused only later by `MkdirAll`; under staging nothing later would
   refuse it. **A separated destination is refused** (ruling, 2026-10-14): when the record root
   variable is set — the condition under which `record.NewRun` → `RecordsDir` → `adoptRoot` puts the
   record outside the run and writes pointers keyed by the run's absolute path
   (`recordroot.go:212,226,295`) — `Migrate` refuses before creating anything, naming the variable,
   because a stage renamed to `--to` would carry a marker whose pointer names the stage. No pure
   predicate exists today — `RecordsDir` with the variable set calls `adoptRoot`, which writes — so
   `recordroot.go` gains one (`RecordRootSet()`: the variable, trimmed, non-empty; for a stage that
   does not exist yet no marker or pointer can exist, so that is the whole condition) and
   `RecordsDir` uses it. `Migrate` calls it; it never calls `NewRun` to find out.
   What is accepted is an allow-list, nothing else: `os.Lstat(to)` absent, or
   `Mode().Type() == fs.ModeDir` with no entries — which excludes symlinks, files, and on Windows
   junctions and mount points (reported since Go 1.23 as `ModeIrregular`). On Unix `--to` must also
   be on the same device as its parent (`Stat_t.Dev`, in `publish_unix.go` behind a function the
   Windows file answers with "same"), so a mount point is refused here rather than failing the
   rename with EXDEV or EBUSY after the whole build. A Linux bind mount of a directory on the same
   filesystem has the same device and passes; its rename then fails with EBUSY after the build, the
   stage is kept and the error names it — stated in the comment. The device check is not exercised
   in CI (making a mount point needs privilege) — a stated residue, by ruling; a missing check
   fails loudly at the rename. A `--to` that is the operator's current directory is NOT refused
   (ruling, 2026-10-15): the Unix rename replaces it, and the operator's shell is left in the old,
   unlinked directory until it changes directory — stated in the help. An `os.Lstat(to)` error
   other than not-exist is a refusal, never "absent".
   Every refusal in this step runs before anything is created or removed. The parent of `--to`
   must be writable, since the stage is created there. That is decided by the stage's `os.Mkdir`
   succeeding, not by `access(2)` (ACLs and read-only mounts make that unreliable). A failure there
   is a refusal that wraps the OS error — EACCES, EROFS, EDQUOT and ENAMETOOLONG are different
   facts — and adds, for EACCES, that the stage is created beside `--to`. Two cases migratable today
   are now refused, stated in the help: a writable empty `--to` in a parent the operator cannot
   write, and a `--to` whose base name is within about 25 bytes of the filesystem's name limit (the
   stage's name adds `.` and `.migrating-<random>`).
1. **The stage.** It creates
   `--to`'s missing parents (`MkdirAll`, as `copyChannels` does today), records which parents it
   created, then creates `stage = <parent>/.<base>.migrating-<random>` with `os.Mkdir(stage, 0o755)` (a random
   suffix, retried on EEXIST) — the mode `copyChannels`' `MkdirAll(toDir, 0o755)` gives today, umask
   applied by the kernel; `os.MkdirTemp` would make it 0700, and reading the umask means changing it.
   When `--to` is an existing empty directory, it is REPLACED by the stage, which has that mode —
   nothing of the old directory is copied (ruling, 2026-10-15): its mode, owner, group, ACLs and
   extended attributes (an SELinux label) are the stage's. A residue stated in a comment and the
   help, with its one loud consequence: in a sticky parent such as `/tmp`, a `--to` owned by another
   user makes `rename(2)` fail EPERM after the build, the stage kept. Everything `Migrate` writes today goes to `stage` instead of `--to`:
   channels, the registry translation, `record.NewRun`, the replay (including `replayInOrder`'s
   close-delete-rebuild at `order.go:62`), seat turns, the manifest.
2. **The source's scratch copy is gone before the publish.** Today `scratch :=
   filepath.Join(toDir, ".migrate-src")` (`migrate.go:41`), opened by a plain `sql.Open` outside
   the cache (`sqlitesource.go:65`) and removed by a `defer` after return — after the publish, under
   staging. It stays inside the build, and the source is closed and the scratch removed BEFORE the
   publish step; a removal that fails is an error, so the copy is never synced or published. The
   published run must hold exactly the file set today's `Migrate` leaves (M4). (Amended in
   implementation, #1386: the plan first moved the copy to `os.MkdirTemp`, and verification found
   a killed migration then left a copy of the source record in the system temp directory, named
   nowhere. Inside the build, a killed migration leaves exactly the directory the help names.)
3. **Sync the tree, then publish.** In order, on success and on the refusal path (which today
   publishes its manifest into `--to` and returns the error, and still does):
   1. `recordsql.CloseUnder(stage)` — every handle under the stage closed, so SQLite has written
      everything it is going to write.
   2. `syncTree(stage)`: walk it; fsync every regular file (opened read-write on Windows, where
      `FlushFileBuffers` needs write access) — an error here is an error; then, on Unix, fsync every
      directory, deepest first — an error here is a warning. A WAL or shared-memory file left behind
      is synced like any other file.
   3. Publish, through one seam `renameDir(stage, to)`, in `publish_unix.go` (`//go:build unix` —
      `_unix` is not a file-name constraint) and `publish_windows.go`. Unix: `rename(2)` itself
      (`syscall.Rename`, retried on EINTR as `os.Rename` does internally),
      NOT `os.Rename`, which Lstats the destination and returns EEXIST for any directory before the
      syscall (`os/file_unix.go:25-44`, go1.27.1; probe in
      `~/.claude/scratch/migrate-bulk-load-audit/round3/probe/`). `rename(2)` replaces an existing
      EMPTY directory atomically and refuses a non-empty one (ENOTEMPTY) or a file (ENOTDIR) without
      deleting anything, so `--to` is never absent and nothing removes it first. Windows, which
      cannot rename over a directory: `syscall.RemoveDirectory(to)` if it exists — a directory-only
      call, so a `--to` that became a file since step 0 is refused, not deleted — then
      `os.Rename`. The raw call does not get `os`'s long-path handling, so a `--to` of 248 or more
      characters fails there after the build, the stage kept and `--to` intact — stated.
   4. Unix: fsync every directory whose entries changed — `--to`'s parent, each parent `Migrate`
      created in step 1 (deepest first), and the nearest ancestor that existed before, which holds
      the entry of the first created parent. Each failure is a warning. Windows: none.
   5. Warnings (a directory sync that failed) are collected on the
      returned `Manifest` in a field that is not serialised (`json:"-"`) — `inputs/migration.json` is
      written before the publish and cannot carry them — and the verb prints each to stderr, in
      `--json` mode as well (where `printManifest` is skipped), through `cmd.ErrOrStderr()`. The
      manifest's on-disk shape is unchanged, and no golden renders the field.
   6. **No text that outlives the stage names it.** Text built while the run is built embeds the
      run's path, which is the stage. The operator receives such text at exactly four exits:
      - **the refusals in the manifest** — `Refusal.Err` from `err.Error()` in
        `internal/record/migrate/replay.go:161,171,233,238,245`, some naming the class registry's
        absolute path built in `internal/record/replay.go:488,496,499`. They are rewritten IN THE
        `Manifest` STRUCT before `Manifest.Write`, so the on-disk `inputs/migration.json` and the
        `*Manifest` that `Migrate` returns — which the verb encodes to stdout under `--json`
        (`internal/cli/migrate.go:53-54`) and prints otherwise — carry the same text. `Refusals[].Err`
        is the only `Manifest` field that holds such text (round 3 checked every field);
      - **the published `inputs/migration.json`** — the same text, by the above;
      - **warnings** — a failed directory sync in step 3.2 names a stage directory;
      - **returned errors** — from `translateStagedRegistry` (`migrate.go:162,167`), `record.NewRun`
        (`record.go:212`), the `recordsql` errors that reach `Migrate` (`store.go:243,395,407`) and
        anything else.
      Warnings and returned errors are rewritten at `Migrate`'s return, and only when the run is no
      longer at the stage — published, or the stage removed successfully. When the stage is kept (a
      failed rename, or a failed removal) they keep naming the stage, so the leftover is never
      hidden. The refusals are rewritten at `Manifest.Write` regardless: they describe the run that
      is, or would have been, published at `--to`, and a comment says so — a kept stage's own
      `inputs/migration.json` therefore names `--to`.
      The rewrite of an error is a type, `atToError{err, msg}`, whose `Error()` is the rewritten
      text and whose `Unwrap()` returns the original, so `errors.Is`, `errors.As` and `feov.CodeOf`
      answer as before (the verb's refusal envelope takes its code from the chain,
      `internal/cli/seat/refusal.go:96`). It runs once, at `Migrate`'s return, after every
      `fmt.Errorf("…: %w")` inside `Migrate` has folded its text in. The replacement is of the
      stage's absolute path by `--to`'s — an exact prefix, unique by construction since the stage
      name carries a random suffix. Its comment records the constraint it depends on: every path
      under the stage is built by `filepath.Join`/`filepath.Abs`, never resolved through symlinks,
      URL-escaped (the `file:` DSN at `recordsql/store.go:103` appears in no error) or `%q`-quoted.
      A rewritten error from a failure BEFORE the publish names a path under `--to` that never
      existed there; every such error says the path belongs to the run that was being built for
      `--to`. The refusal path's error names `<to>/inputs/migration.json`; if that path's rename
      fails, `Migrate` returns the rename error naming the kept stage, joined with the refusal count
      (`errors.Join`), never the refusal error alone.
4. **Errors.** Before the rename (a build failure, a sync failure): `CloseUnder(stage)`, remove the
   stage; `--to` is untouched. Parents step 1 created are left, as today (ruling, 2026-10-15). The rename
   itself failing: the stage is KEPT and the error names it, since it holds a complete synced run
   (on Windows this includes a rename that fails after `syscall.RemoveDirectory(to)`, where `--to` is then absent and
   the error says so). After the rename (a directory fsync failing in step 3.4): `--to` is
   published and left alone. A directory fsync failure is never an error (best-effort, by ruling);
   it is a warning (step 3.5), so a filesystem without directory sync — some network and FUSE
   mounts, possibly WSL drvfs — still migrates and says what it could not guarantee. The source's scratch copy
   (step 2) is released by `defer` as well as explicitly, so error paths do not leak it.
   When the rename fails with no refusals, `Migrate` returns the `*Manifest` with the error, so the
   verb still prints its warnings (which name the kept stage). When the stage's removal fails after
   a build or file-sync failure, the removal error is joined (`errors.Join`) to the build error, so
   the returned error names the kept stage even when the build error names no path.
5. **No reaping** (ruling, 2026-10-15). A killed migration leaves its stage; nothing deletes a
   directory by name. The help says what the directory is and that it is safe to delete by hand.
6. **Help.** `migrate --help` gains: the run is built beside `--to` under a `.migrating-` name,
   synced, and moved into place when it is complete, so an interrupted migration leaves `--to` as
   it was (on Windows, possibly absent if it was an empty directory); its files are synced before it
   is moved, and its directory entries are synced where the filesystem allows — best effort, with a
   warning where it could not (on Windows, not at all); a leftover `.<name>.migrating-`
   directory beside `--to` is what an interrupted migration left and is safe to delete by hand.
   `--to` must be absent or an empty, plain directory (not a link, junction or mount point) in a
   writable parent, with a base name short enough to take the stage's suffix; a separated record
   root is refused. An existing empty `--to` is replaced, not filled: its mode, owner and
   attributes become the new directory's. On Unix a shell sitting in it is left in the old one
   until it changes directory; on Windows the directory cannot be removed while it is in use, so
   the publish fails after the build and the stage is kept. The `--to` flag's
   own text (`internal/cli/migrate.go:62`, today "refused if it already holds anything") is
   rewritten to the same rule. `difftest`'s manual golden is regenerated and each changed line
   read.

### Not in this plan: the speed half (re-planned)

By ruling (2026-10-15), the per-event fsync is to be removed by batching the replay through #1379's
caller-held-transaction path (`writeIn`, `record.go:621`; `Append` validates through its writing
transaction at `record.go:578`), which keeps `synchronous=FULL` and adds no process-wide switch. The
obstacles that plan must answer, found in round 5: replay's own reads between appends
(`record.SeatOwesOccasion` at `replay.go:216`, `ReportProjection` at `order.go:48`) go through the
run, not a transaction; `appendCorrected`, `RegisterSeat` and `AppendSeatTurns` have no transaction
form; and a refused event must not roll back its batch (a savepoint per event). It gets its own
plan and audit — #1382 — and nothing here depends on it.

### Alternatives rejected

| Alternative | Why not |
|---|---|
| Write `--to` in place, syncing every file at the end | An interrupted migration still leaves a half-run at `--to` that opens |
| Trust SQLite's checkpoint for durability | Finds nothing to sync once a close has folded the WAL (§II) |
| Reap leftover stages automatically | Ruled out: it deletes directories by name, including an operator's that matches |
| Re-point a separated record root after the rename | Ruled out in favour of refusing a separated destination |

### Consumer census

Contract changed: where and how `Migrate` writes `toDir`, and what `--to` may be. Census output:
`~/.claude/scratch/migrate-bulk-load-audit/publish-r1/` (`census-migrate.txt`,
`census-migrating.txt`, `census-recordsdir.txt`, `census-verb-callers.txt`, `census-docs.txt`).
Name-only matches, not calls and unchanged: `shardarchive_test.go:21`, `everyarchive_test.go:34`,
`everyrender_test.go:49`, `reopenedm19_test.go:19`, `order_test.go:51,245`,
`archive_test.go:234,241`, `retireanchor_test.go:397,421`, `duplicatescreen_test.go:168-183`
(`record.Migrating`), `labelkey_test.go:15`, `replay.go:301`.

```
git grep -nE 'migrate\.Migrate\(|[^.]Migrate\(' -- plugins/frank-exchange-of-views/tools
git grep -n 'migrate-src\|migrating' -- plugins/frank-exchange-of-views
git grep -nE 'RecordsDir\(' -- plugins/frank-exchange-of-views/tools
```

| Consumer | Changes? |
|---|---|
| `internal/cli/migrate.go:51` (the verb) | No call change; help (step 6) |
| `internal/record/migrate/*_test.go`, 21 call sites; `internal/cli/composedsection_test.go:256` | No: each passes an absent path or an existing empty directory — `recordtest.TmpRun`, a `t.TempDir` or an `os.Mkdir`'d one, at eleven or more sites including `composedsection_test.go:255`, `correction_test.go:157`, `gapplace_test.go:216`, `avenuerename_test.go:56`, `order_test.go:181,242`, `classmaterial_test.go:140,226,245,340`, `remaplists_test.go:375` and the archive store's directory — and reads it after return. The publish replaces the empty directory, which `t.TempDir`'s cleanup tolerates. The tests are the gate (round 3's census: `round3/census-*.txt`) |
| `RecordsDir`'s callers | No: it computes the same answer through the new predicate |
| Readers of `Manifest.Refusals` (rewritten in the struct, step 3.6): the refusal-count error, `printManifest`, `capture/migrationline.go:38`, `order.go:55`, tests | No: each reads the count or a phrase, none a run path (`publish-r4/census-refusals.txt`) |
| `internal/cli/m23contradiction_test.go:42` (into a `recordtest.TmpRun`; `:47` matches the name only) | No; same class as the sites above, fails loudly if the publish breaks it |
| Verb-level callers: `internal/cli/retireanchor_test.go:419`, `internal/cli/migrateorder_test.go:41` (`run(t, "migrate", …)`, the latter into a `TmpRun`) | No; each fails loudly if the publish breaks it |
| `archiveonce_test.go:66` | No: `CloseUnder` on its destination after `Migrate` finds nothing open, a no-op |
| `root.go:270`, `classmaterial_test.go:238,267` | Name matches only; not calls |
| `migrate.go:41` and `sqlitesource.go:65` (`.migrate-src`) | Yes: step 2 |
| `plugins/frank-exchange-of-views/README.md:13` | Describes `migrate` without contradicting a staged publish; no change |

## IV. Risk & Mitigation

| # | Risk | L × I | Mitigation |
|---|---|---|---|
| K1 | A killed migration leaves a stage beside `--to` | M × L | Nothing reads a stage; the help names it and says to delete it by hand (no reaping, by ruling). V.2 |
| K2 | `Migrate` reports success before the run's files are on disk | L × H | `syncTree` after `CloseUnder`, then the rename; V.3 records every sync and the rename order, V.6 confirms with strace |
| K3 | The rename fails — across filesystems, onto a bind mount, a long Windows path, a sticky parent | L × M | The stage is a sibling of `--to`; mount points refused up front on Unix; every rename error keeps the stage and names it |
| K4 | Windows: no directory fsync; the rename is not write-through; `Sync` needs write access; antivirus may hold the directory | M × L | Directory-only removal, files opened read-write, the residue stated in M3 and the help. CI's Windows leg runs the migrate tests and may show a transient "Access is denied" |
| K5 | macOS: directory `Sync` is `F_FULLFSYNC`, undocumented for directories; no CI leg runs Go tests there | M × L | Directory syncs warn rather than fail (ruling); stated residue |
| K6 | The output differs from today's | L × H | V.1 compares rows, file listings and modes for every archive |
| K7 | A `--to` that is not an empty plain directory is deleted or written through | L × H | Step 0's allow-list before any work; the Unix rename refuses a non-empty directory or a file; Windows removal is directory-only. V.7 |
| K8 | A separated record root puts the database outside the stage | L × H | Step 0 refuses before writing, by ruling. V.7 |

## V. Verification Plan

1. M4, real data: every archive in `run-archive/` migrated by `origin/main`'s binary and by the
   branch's into scratch directories; `SELECT * FROM events ORDER BY id` and each body table dumped
   and compared — identical; manifests compared minus timestamps; `find <to> -printf '%M %P\n' |
   sort` identical except for `records/record.db-wal` and `records/record.db-shm`, which today's
   verb leaves and the branch's sealed record does not. A `.migrate-src` anywhere in the branch's
   output is a failure.
2. M2: a test runs `Migrate` in a subprocess and kills it mid-replay: a registry entry blocks on a
   chosen event after writing a marker file, and the parent waits for the marker before killing, so
   the kill lands after the stage exists and replay has begun. It then asserts `--to` absent or empty and a `.migrating-*`
   sibling present; a second `Migrate` into the same `--to` succeeds and leaves the first stage
   where it was (no reaping, by ruling).
3. M3: the publish's syncs, its rename and the stage's removal go through package variables in
   `migrate` (`syncFile`, `renameDir`, `removeStage`), and a test replaces them with recorders. After a real migration of one archive it
   asserts: every regular file in the published run was synced, each after the last write to it
   (the recorder hashes the file's content at sync and the test re-hashes it after return); every
   directory of the run was synced; the rename came after the last of those; and after the rename
   `--to`'s parent, each created parent and the nearest pre-existing ancestor were synced — the
   test passes a `--to` two levels below an existing directory so all three kinds occur. Paths
   recorded at sync are stage paths, mapped to `--to` for the comparison. On Unix it also asserts,
   inside the `renameDir` recorder, that an existing empty `--to` still exists (`os.Lstat`) when the
   rename is called — nothing removed it first. Directory-sync assertions are skipped on Windows.
   Two cases, because one migration cannot have both: a `--to` two levels below an existing
   directory (absent; every directory-sync kind occurs), and an existing empty `--to` (the `Lstat`
   assertion). A directory sync made to fail through the seam produces a warning on the manifest and
   a nil error; a FILE sync made to fail produces an error and an untouched `--to`.
   `renameDir` itself is tested directly on Unix: over an empty directory it replaces it; over a
   non-empty one it fails with ENOTEMPTY and the contents are intact; over a file it fails and the
   file is intact. Fault injection: a failing `syncFile` on any one file
   makes `Migrate` return an error, leaves `--to` untouched, and leaves no `.migrating-*` sibling.
   With the real seams on Linux, a migration of one archive produces ZERO warnings — a directory
   sync that always failed would otherwise pass as a warning. A migration that refuses an event
   whose refusal text carries the run's path (a class registry made invalid in the source) publishes
   an `inputs/migration.json` in which no `.migrating-` appears and the path names `--to`, and
   `json.Marshal` of the returned `*Manifest` contains no `.migrating-` and its refusals equal the
   published file's; a build that fails after the stage exists — injected as an error carrying a
   non-default `feov` code — returns an error in which no `.migrating-` appears, which says the path
   belongs to the run being built for `--to`, for which `errors.Is(returned, injected)` holds and
   whose `feov.CodeOf` is the injected code; the refusal path's error (a published run) does NOT
   carry the "being built" clause — the clause is keyed on whether the run was published, decided
   at return, not on when the text was built; the injected directory-sync failure's
   warning contains no `.migrating-` and names the `--to`-side directory; with the stage's removal
   made to fail (through `removeStage`), the returned error and any warning name the kept stage;
   and on the refusal path with the rename made to fail, the returned error names the kept stage
   and the refusal count. The verb's
   warning printer is tested directly in `internal/cli` (it takes the `Manifest` and the command):
   each warning reaches `cmd.ErrOrStderr()`, with and without `--json` — the seams themselves stay
   unexported.
4. Mutations, each red then restored: drop `syncTree` (V.3 red); sync before `CloseUnder` (V.3 red:
   files written after their sync); publish before `syncTree` (V.3 red); drop the parent sync after
   the rename (V.3 red); drop the nearest-ancestor sync (V.3 red); remove `--to` before the Unix
   rename (V.3 red); drop the rewrite at any one of the four exits (V.3 red); rewrite the marshalled manifest bytes
   instead of the struct (V.3 red: the returned `*Manifest` names the stage); flatten the rewritten
   error with `errors.New` (V.3 red); put the scratch copy under the stage AND drop its
   explicit removal (V.1 red); leave the stage in place on error (V.3 red).
5. `go -C scripts run ./check` before push; read the FAIL COUNT. `FEOV_RELEASE_GATE=1 go test
   ./releasegate/fuzz/` once.
6. Driveable on real data, Linux: `feov-record migrate --seat-id operator --from <an archive
   extracted to scratch> --to <scratch>/out` under `strace -f -e trace=openat,write,pwrite64,copy_file_range,sendfile,splice,fsync,fdatasync,renameat,renameat2,rmdir,unlinkat`
   (`io.Copy` between files writes through the middle three on Linux), and read
   that every file under the stage has an fsync after its last write and before the rename, and that
   no `rmdir` or `unlinkat` of `--to` precedes the `renameat` (which the in-process recorder cannot
   see, because it replaces the shipped `renameDir`). Then
   `feov-record dashboard` on the output. Then the same migration killed partway (`timeout -s KILL`)
   and `ls -a` of the parent.
7. `--to` refusals: a regular file, a symlink to an empty directory, and a non-empty directory at
   `--to` are each refused before anything is created (no stage, no parents), and the file and the
   symlink's target are untouched. With the record root variable set, `Migrate` refuses before
   creating anything and names the variable; nothing is written under the root. An existing empty
   `--to` created 0700 is published with the stage's mode (0755, umask applied) — the stated
   residue, asserted so a change to it is seen. A `--to` whose parent is not writable is refused
   with the OS error wrapped and the reason; a `--to` whose base name leaves no room for the stage's
   suffix is refused with ENAMETOOLONG wrapped. The Unix mount-point refusal is NOT tested (making a
   mount point needs privilege) — stated residue, by ruling. On the Windows leg, a junction at
   `--to` (made with `mklink /J`, skipped if that is unavailable) is refused and its target
   untouched.
8. The source's scratch copy, inside the build, is gone before the publish: a test with a SQLite
   source (the shard-era archive the sync-order test uses has no copy) asserts it is neither synced
   nor present at the rename nor in the published run, and the copy of a killed migration is inside
   the build it left. A `--to` inside `--from` is refused and the source is unchanged.
