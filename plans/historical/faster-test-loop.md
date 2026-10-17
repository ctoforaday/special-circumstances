# Faster test loop

Status: shipped — historical record, 2026-10-17. Part A (#1321), Part R (#1326), Part B (#1332, local
only) and the OCR step change this work turned up (#1330) are merged, with the outcome PR #1334;
the follow-ups it filed (#1335-#1339) are closed — #1351, #1369, #1405, #1361, #1402. The
measured outcome and the rulings that followed it are under "Outcome" below; the sections after
it are the design as audited. Related: #1013, #1046 (both measure the same gate from the outside).

## I. Summary & Goals

### The problem

The development loop waits on one gate, and that gate cannot use the cores it is given.

Measured 2026-10-11 (raw data: `~/.claude/scratch/test-time/`):

| Evidence | Figure |
|---|---|
| `feov-record:race` across 109 session transcripts | 78 runs, 40 red, median 1,323 s, worst 2,842 s |
| Every other gate, same corpus | median under 160 s |
| CI run 38120976859, `feov-record (ubuntu)` | 1,459 s, of which the race step is 1,081 s |
| Inside that race step | `internal/record/migrate` 1,012 s, `internal/cli` 846 s, `internal/record` 408 s |
| Same three packages without `-race`, same run | 81 s, 61 s, 16 s |
| Local uncached FEOV suite, 16 cores, load ~10 | wall 215 s; `migrate` alone 205 s |
| Two peer sessions, each package alone under `-race` | `migrate` 2,183–2,662 s, `cli` 1,975–2,499 s, zero races — past the gate's 30-minute bound |

Two causes, both structural:

1. **One fixture is rebuilt many times.** `internal/record/migrate` migrates every archived run
   (19 tarballs on `6afe2976`) in each of nine sweep sites, about 20 s a sweep, plus 13 single-archive
   calls. Seven tests account for 169 s of the package's 205 s. The archive count grows with every
   captured run.
2. **The slow packages are serial.** `migrate` and `record` hold no parallel test; `cli` holds 20 of
   461, the rest serialised by `t.Setenv` (69 uses) and a package-level fetcher swap. Each runs on
   one core, so the wall clock is the slowest single package whatever the core count.

### Success criteria

Measured on the same box in one session, arms interleaved, suite run solo for the verdict.

| # | Criterion | Baseline | Target |
|---|---|---|---|
| A1 | `go test -count=1 ./internal/record/migrate/` wall | 205 s (load ~10) | at most 35% of an interleaved baseline run |
| A2 | Migrations per archive per test process through the helper | 7–9 | exactly 1, asserted by a test |
| A3 | CI `internal/record/migrate` under `-race` | 1,012 s | at most 400 s |
| R1 | CI race step, `feov-record (ubuntu)` | 1,081 s | at most 300 s — an aim from a one-run, file-level prototype; the gate is two runs over a somewhat larger set, so the first CI run restates it |
| R2 | Local `feov-record:race` gate, load ≤ 10 | red by timeout | green, at most 300 s, restated the same way |
| R3 | Concurrency starts in the shipped graph executed under `-race` | not measured; at least one is executed by no test | every one, asserted by the gate |
| R4 | Tests raced | every test in the 45 packages | the derived set, printed by the gate per package (file-level prototype: 104 tests in five packages; the rule above adds `capture` and the `seatprobe.Build` callers) |
| B1 | Local `feov-record:race` gate, 16 cores, load ≤ 10 | red by timeout | green, under the 30-minute bound; 600 s is the aim and is restated from Part A's measurement (R12) |
| B2 | Local `feov-record:test` gate | 378 s (peer, 4 cores) / 215 s | at most 90 s |
| B3 | CI `feov-record (ubuntu)` job | 1,459 s | at most 780 s |
| B4 | Tests executed | the package's `-test.list` | the same set, every name with a verdict |

A target missed is reported with the measured figure; it is not grounds to widen the change.

### Outcome (measured 2026-10-11 and 2026-10-12)

CI figures are the Ubuntu `feov-record` job. CI tests the merge with `main`, so every run after
#1321 merged includes Part A.

| # | Target | Measured | Verdict |
|---|---|---|---|
| A1 | at most 35% of baseline | 47 s and 57 s against 218 s and 285 s, interleaved (22%, 20%) | met |
| A2 | exactly 1, asserted | `TestTheArchiveStoreMigratesEachArchiveOnce` | met |
| A3 | at most 400 s | 360 s (from 1,012 s) | met |
| R1 | at most 300 s | 240 s, then 187 s (from 1,081 s; 889 s with Part A alone) | met |
| R2 | green, at most 300 s | 141 s solo; 165.5 s inside a full `check` at load 12–20 | met |
| R3 | every start executed | 5 of 5: `dashboard_serve.go:77` and `:79`, `seatprobe/build.go:349` (two), `debatejs/debatejs.go:331` | met |
| R4 | the derived set, printed | 61 tests in 6 packages; one selected test skips by design (`TestWriteSeatProbeFixture`) | met |
| B1 | — | moot: the race gate is Part R's | — |
| B2 | at most 90 s | 59 s, 62 s, 71 s against 94 s for plain `go test`; 66.9 s and 68.7 s as the gate inside `check` | met |
| B3 | at most 780 s | 526 s and 590 s with Parts A and R; Part B adds nothing in CI (below) | met without Part B |
| B4 | the same set, every name with a verdict | 71 packages, 2,198 top-level tests, equal pass/fail/skip counts on both paths | met |

The whole-graph race on Part R's commit, for the verdict comparison: 1,663 s locally, zero data
races, and all 61 scoped tests have the verdict there that they have in the scoped run.

**Part B does not help a four-core runner.** Through `testsplit` CI's suite step took 134 s and
111 s against 73–104 s for plain `go test`; `internal/cli` alone went from 57 s to 100 s. Plain
`go test` already runs four packages at once there, so every core is busy and more processes are
contention. Ruling 6 follows from that.

**The OCR engine step was the next cost, and it was not tests.** With the race step at 187–240 s,
the step "the OCR engine's own suite" (193–268 s) was as long. Its log held about 36–70 s of
tests and 1,230 lines of libc++ being compiled inside the first test binary's link: zig builds
its C runtime, libc++ and compiler-rt from source on first use and caches them under the home
directory, which a fresh runner does not have. #1330 keeps that cache with the C stack and runs
only the `fetchcache` tests that reach the real engine: 44 s with fresh links, 17 s re-running
identical code.

**What bounds each gate now.** Locally the FEOV suite's floor is `internal/fetchcache` at about
60 s, most of it one pacing test that sleeps a real crawl delay. In CI the Ubuntu job's largest
steps are the race gate and the plain suite, each under four minutes.

### Value case

The naive answers are more cores (already tried: the critical path is single-threaded) and a longer
timeout (already tried: 10 → 30 minutes; the gate now fails at 30). Part A removes work outright, at
the cost of one test helper. Part B converts cores into wall clock without rewriting 900 tests, at
the cost of one dev tool and two workflow lines. Measured in a prototype with eight processes, all
green: `cli` 77 s → 21.7 s.

### Why the race gate is re-scoped (Part R)

The race detector reports a race only between two goroutines that both run in the test binary.
Counted over the 45 packages the gate races today (text match, so a floor):

| Package | CI race s | `go` statements in source | In tests |
|---|---|---|---|
| `internal/record/migrate` | 1,012 | 0 | 0 |
| `internal/cli` | 846 | 1 (`dashboard_serve.go:76`) | 20 `t.Parallel`, 2 test HTTP servers |
| `internal/record` | 408 | 0 | 4 files |
| `internal/fetchcache` | 143 | 0 | 2 files, 22 test HTTP servers |
| `internal/record/recordsql` | 57 | 0 | 3 files |
| `internal/sittingcap` | 1 | 0 | 1 file |
| 39 others | 174 | 0 | 0 |

File-level prototype, 2026-10-11, load ~3: racing only the tests in test files that start a
goroutine — 104 tests in five packages, fewer than the rule in §III selects — took 122 s, 60 s, 23 s, 58 s and 2 s per package, all green, zero races.

And the whole-graph gate does not do what it is trusted to do: the ONE `go` statement in shipped
code, `go watchAndShutdown(srv, run)` (`dashboard_serve.go:76`, with `ListenAndServeTLS` two lines
below it), is executed by no test. The only test that calls
`serveDashboard` (`TestServeRefusesWhenTheRunIsNotLive`) takes the refusal path above it. Eighteen
minutes of race detection per pull request have never raced it.

## II. Technical Context

- Go 1.27, `GOTOOLCHAIN=local`. FEOV tools module `plugins/frank-exchange-of-views/tools`; dev
  tooling in the `scripts` module. Nothing here ships to a consumer; no `version` moves.
- The record is SQLite (modernc, WAL). A migrated record is a directory holding the database and
  its siblings.
- Gate set: `scripts/check/gates.go`, mirrored against `.github/workflows/hooks.yml` by
  `scripts/check/parity_test.go`. `scripts/golden`'s go leg is `go test -count=1 ./...`, which
  `check` subsumes under the module's `:test` gate (#626).
- Constraints: `-count=1` stays on every suite gate (`noCache`). The race gate's subject stays the
  shipped binary's import graph; ruling 3 changes which TESTS in it are raced on a pull request.
  No test is removed, skipped or made `-short`.
- Part R adds one dependency to the `scripts` module, `golang.org/x/tools/go/packages`, for typed
  syntax trees; nothing in a plugin imports it.
- Parts A and B change no shipped file. Part R changes one, `internal/cli/dashboard_serve.go`, to
  give its test two seams; behaviour is unchanged and no `version` moves.
- "Shard" already names the sharded-JSONL storage era in FEOV (`shardarchive_test.go`,
  `TestShardEraArchivesMigrate`). This plan says **split**; the tool is `testsplit` (census:
  `git grep -il testsplit` hits three files, each the substring of `TestSplit…`, none a concept).

## III. Proposed Changes

### Part A — each archive is migrated once per test process

```
plugins/frank-exchange-of-views/tools/internal/record/migrate/
  archiveonce_test.go      [NEW]     package migrate: the once-per-archive store and its copy-out
  archiveonce_ext_test.go  [NEW]     package migrate_test: migrateArchive, moved here, body replaced
  shardarchive_test.go     [MODIFY]  migrateArchive leaves this file
  findingarea_test.go      [MODIFY]  line 27 reads the store
  remaplists_test.go       [MODIFY]  line 336 reads the store; line 380 stays a direct call
  testmain_test.go         [MODIFY]  removes the store's directory before the process exits
plugins/frank-exchange-of-views/tools/internal/record/recordtest/
  recordtest.go            [MODIFY]  a variant of Main that runs a function after m.Run, before the guard
```

**Design.** A package-level store keyed by tarball name holds, per archive, a `sync.Once`, the
migrated directory (under one `os.MkdirTemp` root owned by `TestMain`), the `*Manifest`, and the
error. The first caller extracts and runs the real `Migrate(src, dst, Entries(), Options{})`, then
calls `recordsql.CloseUnder` on the destination — `Migrate` leaves that handle in
`recordsql.openCache`, keyed by absolute path, and a copy taken under an open handle is not the
record. Every caller, the first included, receives **its own copy** of the migrated
directory under `t.TempDir()` and opens that. A test that writes to its record therefore cannot
reach another test's, and no call site has to be audited for writes. The manifest is returned as a
copy for the same reason. Each copy registers a `CloseUnder` cleanup on its own directory, which is
what the orphaned-handle guard and Windows' temp-directory removal require. A copy's manifest keeps
the store's absolute `source_path`; no helper caller reads it (auditor's census, round 1).

**The store answers "this archive, migrated" and nothing else.** It holds one migration per
archive, so a test whose claim is agreement BETWEEN migrations never reads it: handed two copies
of one migration it compares a result with itself and is green forever.
`TestTwoMigrationsOfOneSourceMintTheSameIds` (`remaplists:374-402`) is that test and keeps its two
direct `Migrate` calls, as `archive:36,118` and `classmaterial:341` do.

The store's directory is created on the first migration, not at start-up, so a `-test.list` run
creates nothing. `recordtest.Main` runs `m.Run`, the orphaned-handle guard and `os.Exit` in one function
(`recordtest.go:183-198`), so a deferred removal never runs. `recordtest` gains a variant that
takes a function and runs it after `m.Run` and BEFORE `CheckOrphanedHandles`; `Main` keeps its
signature, so its other callers do not change. The order is the point: the store's directory is
removed first, so a handle the store left open is one the guard then reports.

`TestEveryArchivedRunMigrates` keeps its meaning: every tarball still goes through `Migrate`, and
its refusals are read from the stored manifest. A new test asserts A2 and drives the store itself,
so it holds under `-run` and in any process of a split: it asks the helper for every tarball in
`run-archive`, twice, then asserts the store migrated each exactly once.

**Alternatives rejected.**

| Alternative | Why not |
|---|---|
| Hand every test the one shared record, read-only | Seven test files in the package match a write pattern (`record.Append(`, `.Exec(`); one missed writer corrupts every later test, silently |
| Merge the nine sweeps into one test | Loses one-name-one-claim; a failure no longer says which property broke |
| `t.Parallel()` across sweeps | Same total work on more cores; leaves the `-race` cost in place |
| Commit pre-migrated records as fixtures | The tests exist to prove today's `migrate` on yesterday's records; a stored result proves nothing |
| Cache across processes on disk | A stale cache is a green that tested old code — the defect `noCache` exists to stop |

**Consumer census** — the contract changed is `migrateArchive`'s behaviour (signature unchanged) and
the two in-package sweep calls. Command, run in the package directory:

```
grep -nE 'migrateArchive\(t|archivedRuns\(t\)|Glob\(filepath.Join\(dir, "\*|(migrate\.|[^a-zA-Z.])Migrate\(' *_test.go
```

| Site | Changes? |
|---|---|
| `migrateArchive` callers (19): `archivedgaps:25,98,118` `prechange:14` `remanded:17` `classmaterial:36,62,72,84,291` `benchoccasion:27` `reopenedm19:17` `everyarchive:23` `schemaepoch:22` `everyrender:47` `seatturncarry:109` `gapplace:401` `shardarchive:24,43` | No edit; each now receives a copy of the once-migrated record |
| In-package sweeps: `findingarea:27`, `remaplists:336` | Yes — read the store instead of calling `Migrate` |
| `remaplists:380` (`TestTwoMigrationsOfOneSourceMintTheSameIds`, two migrations of b9 compared) | No — needs two real migrations |
| Sweep enumerators: `archivedRuns` (`remaplists:96`), three `Glob` sites (`benchoccasion:20`, `everyarchive:16`, `everyrender:39`) | No |
| Direct `Migrate` on a synthetic or chained source: `classmaterial:141,227,246,341` `correction:112,157,179,195` `gapplace:144,216` `seatturncarry:48` | No — not an archive, or deliberately a second migration |
| Direct `Migrate` on one archive: `archive:36,118,240,291` `repairregister:75` `avenuerename:57` | No — see scope note |
| Other packages extracting an archive: `internal/cli/prechange_test.go:18`, `internal/record/classmaterial_test.go:274` | No — they extract, they do not migrate |
| Other packages migrating an archive: `internal/cli/composedsection_test.go:241` (m20) — found by `git grep -nE 'migrate\.Migrate\(' -- '*_test.go'` outside the package | No — one call, another process; see scope note |

**Scope note.** The full concept — "no archive is migrated twice in one process" — would also route
the six direct single-archive calls in the package and the one in `internal/cli`. They are left: `archive:36,118` migrate the same run twice on
purpose to compare the results, and the rest cost about a second each. A2 is therefore stated for
the helper, not for the package.

### Part R — the race gate races concurrency, and proves it reached all of it

```
scripts/
  racescope/main.go            [NEW]     derives the race set, races it, proves the starts ran
  racescope/main_test.go       [NEW]
  internal/racescope/starts.go [NEW]     the table of library calls that run caller code on another goroutine
  check/gates.go               [MODIFY]  feov-record:race invokes racescope; raceScope's comment restated
  check/main.go                [MODIFY]  expandScope moves to scripts/internal/racescope
  check/parity_test.go         [MODIFY]  the race-scope and race-bound assertions read the new invocation
.github/workflows/hooks.yml         [MODIFY]  feov-record (ubuntu) race step
.github/workflows/race-nightly.yml  [NEW]     the whole-graph race, on schedule and on dispatch, its own workflow
.github/workflows/nightly-alarm.yml [MODIFY]  watches race-nightly as well as hooks
plugins/frank-exchange-of-views/tools/internal/cli/
  dashboard_serve.go           [MODIFY]  the bind address and the watcher's interval become package variables
  dashboard_serve_test.go      [MODIFY]  a test that runs the served dashboard to its self-teardown
```

`raceScope` already names what the race gate covers (`scripts/check/gates.go`); the tool extends
that concept and takes its name.

**The rule.** *Every concurrency start in the shipped binary's import graph, and every test that
makes one, runs under the race detector on every pull request.* Not every test in the graph.

**Design.** `go -C scripts run ./racescope -module <dir> -deps ./cmd/feov-record [-timeout D]`:

1. Scope: the module-local import graph of the target — today's derivation, unchanged. Empty is an
   error.
2. **Concurrency starts**, found by syntax tree and type information (`go/packages`), never by text:
   a `go` statement, or a call to an entry in `internal/racescope`'s table — `(*testing.T).Parallel`,
   `httptest.NewServer` / `NewTLSServer` / `NewUnstartedServer`, `(*http.Server).Serve*` and
   `ListenAndServe*`, `time.AfterFunc`, `signal.Notify`, `context.AfterFunc`, `(*errgroup.Group).Go`,
   `(*sync.WaitGroup).Go`. Each row carries its reason.
3. **The race set is derived.** A `Test`, `Fuzz` or `Example` function is in it when a start is
   reachable from it by these edges and no others:
   - a DIRECT call, resolved by `go/types`, to a function or to a method on a concrete receiver
     declared in a `_test.go` file of any
     package in the module, or in a non-test file of a module package OUTSIDE the shipped scope —
     the test-support packages (`seatprobe`, `debatejs`, `recordtest`, `runtest`, `testbuild`);
   - the body of a function literal counts as the body of the declaration that encloses it, so a
     `go` statement inside a closure, a `t.Run` body or a handler literal selects its test.

   Reachability does NOT enter non-test files of the shipped scope. `serveDashboard` is called
   from a literal inside `newDashboard()` (`dashboard.go:34,60`), which `root.go:262` calls
   directly; followed into source, every `cli` test that builds the root command would reach
   `dashboard_serve.go:76` and the set would be all of `cli`. A start in shipped source is owned
   instead by a declared test — `//racescope:covers <file>:<function>` above it — and by step 5,
   which refuses a start no raced test executed.

   This rule reaches `seatprobe/build.go:349` (`go srv.Serve`) from the five `cli` tests that
   call `seatprobe.Build`, `debatejs/debatejs.go:256` from `capture/relayfields_test.go`, and the
   six test starts that sit inside literals (`cli/adversarial_test.go:237`,
   `cli/refusalexit_test.go:60`, `record/motionguards_test.go:438`,
   `record/recordsql/windowrace_test.go:48`, `fetchcache/corpus_cgo_test.go:70`,
   `fetchcache/httpfetcher_test.go:225`). A call through a function VALUE or an interface is not
   followed; that residue is R14.
4. **The verdict run carries no coverage.** The set runs per package as `go test -race -count=1
   -json -timeout D -run '^(a|b|…)$' <pkg>`. Every selected name must come back with a `pass`,
   `fail` or `skip` event; a name with none — an `Example` with no `// Output:` is one — fails the
   gate as *not measured*.
5. **The evidence run carries no race detector.** The same set runs again as `go test -count=1
   -covermode=set -coverpkg=<every module-local package the set's binaries link> -coverprofile=…`.
   For every start found in a non-test file of those packages, the statement's count must be
   non-zero; one that is not fails the gate and is named with its file and line. A directive
   naming a start that does not exist fails too, and so does a profile that is missing or does not
   parse.

   The two runs are separate because measured together they are wrong: `-coverprofile` under
   `-race` forces atomic counters, and the detector treats each as a happens-before edge. Probe,
   go1.27.1, three runs each (`~/.claude/scratch/racescope-audit/probe`): plain `-race` reports the
   planted race 3 of 3; the same test with `-coverpkg -coverprofile` reports it 0 of 3.
6. The gate prints, per package, the set's size, the count of each verdict and the NAME of every
   selected test that skipped, and the starts it proved. A concurrency test that skips on its
   environment is unraced, as it is today; printed by name it is a line a reviewer reads, and a set
   that shrinks is a number that moved.

**Why reachability stops at shipped source.** Every `cli` test that builds or executes the root
command reaches every verb — by direct calls through command construction, and through cobra's
dispatch — so a call graph that enters source selects all 478 tests and the gain is gone. Starts
in step 5's scope are few (four today: `dashboard_serve.go:76` and `:78`, `seatprobe/build.go:349`,
`debatejs/debatejs.go:256`), so each is proved executed and a missing or wrong declaration is
refused.

**The residue, stated.** A start reached only through a function value, or made by a library call
the table does not list, is NOT raced on the pull request. It is raced by the nightly, so a race
there is found at most a day after merge instead of before it.

**The backstop is its own workflow.** `race-nightly.yml` races the WHOLE graph — the command the
pull-request gate runs today, `-timeout=30m` per package unchanged — on `schedule` and on
`workflow_dispatch`, ubuntu only. It is where a start this tool cannot see would still surface: a
library that runs our callback on its own goroutine and is not in the table, or a start reached
only through a function value. It is NOT a step in `hooks`: the scheduled `hooks` run has been red
24 of its last 40 (#855), its `feov-record` matrix cancels the ubuntu leg when Windows fails, and
the race step is then skipped — a backstop there does not execute, and its red would be
indistinguishable from the standing one. `nightly-alarm.yml` gains `race-nightly` in its
`workflows:` list, fires on a dispatched `race-nightly` run as well as a scheduled one (its
condition today is `event == 'schedule'`, line 43), and its issue search, title and body are keyed
on the workflow that failed:
today it searches one marker ("nightly is red"), finds #855, and would file a race red there as a
comment about `FEOV_RELEASE_GATE`. A `race-nightly` red lands in an issue of its own that names
the workflow; the fix for one is a table row or a directive. The workflow reproduces the
`feov-record` job's setup (Go 1.27, `GOTOOLCHAIN=local`, cgo, the OCR C-stack cache). `check`
declares the gate and skips it locally with that reason; `parity_test.go`'s `workflow(t)` reads
one path today (`:39`) and reads both, so `TestEveryModuleRaceLegCarriesTheBound` sweeps both.

**The first red is real.** `dashboard_serve.go:76` is uncovered today, so the gate is red until
`dashboard_serve_test.go` gains a test that serves a live run, requests the page WHILE the watcher
is polling, ends the run, and observes the server close — so `render` on a connection goroutine
overlaps `runHasEnded` on the watcher's, which is the race this code could have. `serveDashboard`
binds `0.0.0.0` and ticks every three seconds; both become package variables (as `runLiveMarker`
already is) so the test listens on loopback only and does not wait three seconds. That is the one
shipped-source edit in this plan.

**Not changed:** the other three modules race `./...` in seconds to two minutes and keep doing so;
`feov-record:race-runhandle` keeps its one named test; Windows runs no FEOV race leg today and
gains none.

**Alternatives rejected.**

| Alternative | Why not |
|---|---|
| Race only packages that start goroutines | Measured: drops `migrate`, keeps `cli` and `record` whole — 1,081 s → about 846 s |
| A hand-kept list of concurrency tests | Stale the day a goroutine is added; a miss is silent |
| A naming convention (`*_race_test.go`) | The same list, spelled in filenames, with nothing to refuse a miss |
| A call graph that enters shipped source, or follows function values | Selects all of `cli` through command construction and dispatch |
| Select by per-test coverage of the plain suite | Go has no per-test coverage; running 1,193 tests one process each costs more than it saves |
| Keep the whole graph, make it faster (Part B) | Pays to race code that cannot race, and still never executes the one start in shipped code |

**Consumer census** — the contract changed is the FEOV race gate's invocation and what it claims.
Command: `git grep -nE 'go test -race|raceScope|raceTimeout|expandScope|depsScopePrefix|2026-09-03' -- .github scripts plugins/frank-exchange-of-views/tools CLAUDE.md docs`

| Consumer | Changes? |
|---|---|
| `hooks.yml:691` `xargs go test -race -timeout=30m` (feov-record), and the comment block above it (`hooks.yml:663-683`) | Yes — the tool; the whole-graph form moves to `race-nightly.yml` |
| `hooks.yml:311, 910, 1191` (other modules, `./...`), `hooks.yml:712` (narrow leg) | No |
| `check/gates.go`: `raceScope` and its comment (`gates.go:128-148`), `depsScopePrefix`, `raceTimeout`, `narrowRace`, the `:race` gate; a new declared gate for the nightly | Yes for FEOV's entry and its comment; `narrowRace` no |
| `check/main.go`: `expandScope`, `runGate`'s default branch | `expandScope` moves; `runGate` gains the tool kind for this gate |
| `check/parity_test.go`: `TestRaceScopesMatchTheWorkflow` (line 164 asserts the string `xargs go test -race`, which would stay green on the nightly copy — rewritten to assert the tool line in `hooks.yml` and the whole-graph line in `race-nightly.yml`), `TestEveryModuleRaceLegCarriesTheBound`, `TestExpandScopeDerivesTheModuleLocalGraph`, `TestTheNarrowRaceLegsAreDeclared` | First three yes; the last no |
| Prose stating the old scope, found by reading rather than by the token search: `internal/fetchcache/pdfextract.go:130-131` ("that is CI's race leg") | Restated |
| `releasegate/fuzz/runhandlerace_test.go:24` (cites the 2026-09-03 scope rule) | Comment restated: the carve-out stands, the rule it is carved from is this one |
| `CLAUDE.md` "Verification", `docs/local-development.md` | Read both; edit only where a stated command or claim changes |

### Part B — `testsplit` runs a serial package as several processes

**Conditional (ruling 3).** Written only if, with Parts A and R measured in CI, a gate still waits
on a single-core package. Its race arm is then moot — the race set is small — and B1 moves to R2.
What follows is the design as audited, kept for that decision.

```
scripts/
  testsplit/main.go             [NEW]     the runner
  testsplit/main_test.go        [NEW]
  internal/testsplit/table.go   [NEW]     which packages are split, per module, each with its reason
  check/gates.go                [MODIFY]  FEOV's :test and :race gates invoke testsplit
  check/main.go                 [MODIFY]  uses expandScope from scripts/internal/racescope (moved there by Part R)
  check/parity_test.go          [MODIFY]  race-scope and cache assertions read the new invocation
  golden/main.go                [MODIFY]  FEOV's go leg invokes testsplit
.github/workflows/hooks.yml     [MODIFY]  feov-record (ubuntu) test and race steps; goldens job
```

**Design.** `go -C scripts run ./testsplit -module <dir> [-race] [-deps <target>] [-n N] [-timeout D]`
runs the WHOLE scope it is given, so no package can fall between two invocations:

1. Resolve the scope: `./...`, or with `-deps` the module-local import graph of the target (the
   derivation `expandScope` performs today). An empty scope is an error.
2. Every package the table names for that module MUST be in the scope when the scope is `./...`;
   a table row naming no package is an error.
3. Unsplit packages run as one `go test -count=1 -json [-race] -timeout D -p P <pkgs>`, with
   `P = max(1, NumCPU − N)` so the two paths do not oversubscribe the box, concurrently with:
4. Each split package: `go test -c`, then `<bin> -test.list '.*'`; `Test`, `Fuzz` and `Example`
   names are dealt round-robin into N lists; each list runs as `<bin> -test.count=1
   -test.v=test2json -test.timeout D -test.run '^(a|b|…)$'`, piped through `go tool test2json`,
   with the package directory as its working directory and the caller's environment passed
   through unchanged (`UPDATE_GOLDENS`, `FEOV_RELEASE_GATE`).
5. **Every listed name must come back with a top-level verdict, read from the JSON event's
   `Action` field** — never scraped from a `--- PASS` line, which a test printing without a
   trailing newline hides. A name with no `pass`, `fail` or `skip` event fails the run as *not
   measured*. A process that exits non-zero fails the run and its output is printed whole. The
   tool prints the count of each verdict per package, so a test that starts skipping is a number
   that moved.
6. `N` defaults to `min(8, NumCPU/2)`, at least 1; `-n 1` is the serial behaviour and is the
   comparison arm for §V.

The table (`internal/testsplit`) starts with two FEOV rows — `internal/cli` and `internal/record` —
each carrying its measured serial seconds. `internal/record/migrate` has **no row**: Part A's store
is per process, so dealing its sweeps across N processes migrates every archive N times and
cancels Part A inside the split. Whether it ever needs a row is read from Part A's CI measurement
(ruling 2); a row would need the sweeps dealt to one process, which is a design change stated then,
not assumed now. It is a tuning table, not a
derived carrier: a missing row costs time and never coverage, because step 3 runs whatever step 4
does not. The tool prints each package's wall clock, so the next slow serial package is visible in
every report.

`check`, `golden` and the workflow all call the same command line; `parity_test.go` asserts the
workflow's line against the gate's args. `-count=1` is inside the tool on both paths (steps 3 and
4), and `main_test.go` asserts it on each — the gate's args no longer show it, so
`TestEverySuiteGateDefeatsTheTestCache` reads the tool's argv builder instead of the gate's. That
builder lives in `scripts/internal/testsplit`, beside the table, because a test in `scripts/check`
cannot import package `main`. The
scheduled run's `-timeout=40m` and `FEOV_RELEASE_GATE=1` (`hooks.yml:657-659`) carry over as the
tool's `-timeout` and the passed-through environment. `golden`'s recording leg (`-update`,
`-review`) uses the same invocation with `UPDATE_GOLDENS` in the environment.

Prose that states the old invocation and is rewritten with it: `scripts/golden/main.go:124`,
`scripts/internal/goldenmods/goldenmods.go:4`, `scripts/check/main.go:85`, and the `noCache`
comment in `scripts/check/gates.go`. Windows keeps `go test -short ./...` on pull requests: it is
262 s, off the critical path, and the tool is still built to run there (`.exe`, no shell).

**Alternatives rejected.**

| Alternative | Why not |
|---|---|
| Convert the packages to `t.Parallel()` | 69 `t.Setenv` and a package-level swap to unpick across ~900 tests; process isolation gives each split its own environment for free |
| Break the packages into smaller packages | Moves code and unexported access for a scheduling problem |
| Narrow the race scope | Reversed by ruling 3: it is Part R, and it comes first |
| Raise the timeout again | Done once; the gate is at 33–44 minutes per package against a 30-minute bound |
| A CI matrix per package | Fixes CI only; the local gate stays red |
| In-memory SQLite in tests | Stops testing WAL and locking, which is what several tests are for |

**Consumer census** — the contract changed is how FEOV's suite gates are invoked. Command:

```
git grep -nE 'go test' -- .github/workflows/hooks.yml scripts/check scripts/golden | grep -v _test.go
```

| Consumer | Changes? |
|---|---|
| `hooks.yml` `feov-record`: `go test … ./...` (line 657), `xargs go test -race -timeout=30m` (691) | Yes on ubuntu; Windows unchanged |
| `hooks.yml` `goldens` job (drives `scripts/golden`) | Yes, through `golden/main.go` |
| `hooks.yml` narrow race leg (712), tessocr steps (366, 613, 619), release fuzz (1687), other modules' steps (300, 311, 909, 910, 1014, 1185, 1191) | No |
| `check/gates.go` `gateSet` (`:test`, `:race` for FEOV), `raceScope`, `raceTimeout`, `noCache` | Yes |
| `check/main.go` `expandScope`, `subsumeGoldenLegs` | `expandScope` moves; subsumption holds because both legs are the same command |
| `check/parity_test.go` `TestRaceScopesMatchTheWorkflow`, `TestEverySuiteGateDefeatsTheTestCache`, `TestEveryModuleRaceLegCarriesTheBound`, `TestExpandScopeDerivesTheModuleLocalGraph` | Yes |
| `golden/main.go:133` | Yes |
| `CLAUDE.md` "Verification" (names `check`, `golden`) and `docs/local-development.md` | Read both; edit only where a stated command changes |

**Enumerated, not in this plan:** memory-backed `TMPDIR` for local runs (measured: a sweep 18.1 s →
9.4 s; CI on Linux already does it); a fake clock for `fetchcache`'s pacing test (53.6 s of 64 s);
a prebuilt empty database for fixtures (12% of `cli`'s CPU); `doctor`'s
`TestEveryPublishedPlatformCompiles` (50.7 s). Each is filed as its own issue when this plan is
approved.

## IV. Risk & Mitigation

| # | Risk | L × I | Mitigation |
|---|---|---|---|
| R1 | A copied record differs from a freshly migrated one (WAL not folded, a sibling file missed) | M × H | The store closes every handle before the first copy and copies every file in the directory; §V.3 compares the full suite's results against the serial, uncached baseline; `TestEveryArchivedRunRenders`' byte-pinned golden is the tripwire |
| R2 | The copy costs what the migration did | L × M | A1 is the gate; measured before the rest of Part A is written |
| R1b | A test that compares migrations is routed to the store and compares one result with itself | M × H | The rule above; §V.4(d) makes an id depend on the destination and expects `TestTwoMigrationsOfOneSourceMintTheSameIds` red |
| R3 | A failed migration is cached and every later test reports the same error | L × L | Intended: one cause, stated once per caller with the archive's name |
| R4 | Two processes of one package collide on state outside `t.TempDir()` (a fixed path, a port, a cache directory) — a concurrency no package has with itself today | M × H | Prototype was green once for `cli`; `record` and every `-race` arm are unmeasured. §V.6 runs each split package ten times at N=8 before merge; a collision is fixed at its source, never by lowering N |
| R5 | A test silently runs in no process | L × H | Step 5: every listed name needs a verdict; §V.7 deletes a name from a list and expects red |
| R6 | A hidden order dependency between tests | M × M | Same ten runs; round-robin keeps relative order inside a process |
| R7 | N race-instrumented processes exhaust memory on a 4-core runner | M × M | Default N is 2 there; §V.8 reads peak RSS from the first CI run |
| R8 | Workflow, `check` and `golden` drift onto three command lines | M × M | One tool, one table; parity asserts the workflow line against the gate args |
| R9 | `go test -c` skips the vet that `go test` runs by default | M × L | The `:vet` gates cover every module; §V.9 confirms by planting a vet failure |
| R11 | Split processes and the unsplit `go test` together oversubscribe the box and a timing-sensitive test (`fetchcache`'s pacing) goes red | M × M | `-p max(1, NumCPU − N)` on the unsplit path; §V.6's ten runs |
| R12 | Without a split row, `migrate` under `-race` alone keeps the local gate over B1 (35% of 2,183–2,662 s exceeds 600 s) | H × M | Ruling 2: B1 is restated from Part A's measurement; the choice then is dealing the sweeps to one process, or accepting the restated figure |
| R13 | A library call that runs our code on another goroutine is missing from the table, so its tests leave the pull-request race set | M × H | `race-nightly.yml`, which runs and alarms on its own; each table row is asserted by a fixture in `main_test.go`; a nightly red adds the row |
| R14 | A start reached only through a function value or interface is missed by direct-call reachability | M × M | In source: step 5 refuses an unexecuted start, so the miss is loud. In test code: the nightly; the gate prints the set, so a concurrent test that is absent is visible in review |
| R15 | The evidence run takes a different path from the verdict run, so a start counted as executed was not raced | L × M | Same tests, same `-run` expression, both `-count=1`; the directive names the test that owns each source start, and §V.R3(c) removes it |
| R19 | Coverage instrumentation hides the race it is meant to witness | measured | Steps 4 and 5 are separate runs; §V.R3(g) plants a race and expects the gate red |
| R16 | The new dashboard test is timing-sensitive (a 3-second ticker, a real listener) and flakes under load | M × M | The ticker interval becomes a package variable the test shortens, as `runLiveMarker` already is; ten runs under `-race` at load before merge |
| R17 | The nightly whole-graph race is red for a reason that is not a race (a package over the 30-minute bound, a load flake) | M × M | Its own workflow, so it is not hostage to #855; after Part A `migrate` is the measured figure from PR #1321's CI and `cli` is 846 s, both inside the bound; a timeout there is read as a timeout, never as a race |
| R18 | A real race exists today in code the scoped gate drops | L × H | By construction such code runs on one goroutine in its tests; the nightly run is the check on that argument, and the first scoped run is compared against a same-commit whole-graph run (§V.R5) |
| R10 | Another session's PR changes `migrate` tests or `hooks.yml` under this one (a new archive is in flight on `fix/m20-smoke-findings`) | H × L | `telepathy touched` on each file before editing; main merged in before the final CI |

## V. Verification Plan

Every command starts with `unset $(env | grep -o '^FEOV_[A-Z_]*')`. `T` is
`plugins/frank-exchange-of-views/tools`.

**Part A**

1. Baseline, before any edit: `go -C $T test -count=1 -json ./internal/record/migrate/ > base.json`,
   twice, interleaved with the candidate. → A1.
2. `go -C $T test -count=1 ./internal/record/migrate/` → `ok`, including the new A2 test.
3. Same set, same verdicts: every top-level test name and verdict in the candidate's `-json` equals
   the baseline's (script: `~/.claude/scratch/test-time/top.py`, extended to diff names).
4. **Mutation checks.** (a) Make the store return archive X's directory for archive Y → at least
   `TestEveryArchivedRunRenders` red. (b) Make the store skip one tarball → the A2 test red.
   (c) Make copy-out hand back the shared directory → a writing test (`classmaterial`) turns a later
   reader red, or A2's isolation assertion does. (d) Make a minted id depend on the destination
   directory → `TestTwoMigrationsOfOneSourceMintTheSameIds` red. (e) Remove the `CloseUnder`
   before the first copy → the orphaned-handle guard red, because the store's directory is removed
   before the guard runs. The render golden is NOT the detector: with no writer left, the database
   and its write-ahead log copy consistently.
5. **Real data.** The 19 archives are captured runs, not synthetic fixtures; step 2 is the driveable
   check on them. After merge, read `internal/record/migrate`'s line in the CI race step → A3.

**Part R**

R1. Baseline from today's command on the same commit: the whole-graph race step with `-json`,
    names and verdicts kept.
R2. `go -C scripts run ./racescope -module $T -deps ./cmd/feov-record` → green, the set and the
    proved starts printed; ten runs on a quiet box. → R2, R4.
R3. **Mutation checks.** (a) Add `go func(){}()` to a source function no raced test reaches → red,
    naming file and line. (b) Add a `go` statement to a test outside the set → that test is in the
    next set. (c) Remove the directive from the new dashboard test → red, `dashboard_serve.go`
    named. (d) Point a directive at a start that does not exist → red. (e) Delete one row of the
    table → the fixture test for that row red. (f) Make the profile unwritable → error, not green.
    (i) Move a test's `go` statement inside a function literal → the test stays in the set.
    (j) AFTER merge — `workflow_run` reads the alarm from the default branch — dispatch
    `race-nightly` on a branch carrying a planted race → an issue that is not #855 and names the
    workflow; the branch is deleted after.
    (g) Plant an unsynchronised write shared by two goroutines in a raced test → the gate red with
    `WARNING: DATA RACE`; this is the check that the verdict run carries no coverage. (h) Remove
    the request from the dashboard test → still green on step 5, so the overlap is asserted by the
    test itself (it fails if the page was not served before the run ended).
R4. The new dashboard test alone, ten times under `-race` at load → ten greens.
R5. **Real data.** On the pull request's commit, every test in the scoped set has the verdict it
    has in R1's whole-graph run, and R1 reports zero `WARNING: DATA RACE`. First CI run: the race
    step's seconds (→ R1). `race-nightly.yml` is run once by `workflow_dispatch` after merge and
    its conclusion read; it cannot run from the pull request's branch before that, so the
    backstop is unproven until that dispatch and the plan says so here.
R7. First run of the gate: a selected `Example` with no `// Output:` is red as *not measured*;
    each is given its output or leaves the set by losing its start, never by exemption.
R6. `go -C scripts run ./check` before each push; read the FAIL COUNT. A workflow edit needs CI's
    `qlty`/`actionlint` before any green claim.

**Part B**

6. Baseline, before any edit, from today's command: `go -C $T test -count=1 -json ./...` and the
   race step's `go test -race -count=1 -json <shipped graph>`. Then `go -C scripts run ./testsplit
   -module $T -n 8` and the same with `-race -deps ./cmd/feov-record`, ten times each, on a quiet
   box → ten greens each, and **every test name's verdict, `skip` included, equals the baseline's**
   in both arms. → B1, B2 against the baseline runs, interleaved.
7. **Mutation checks.** (a) Drop one name from a list → *not measured*, red. (e) Remove `-count=1`
   from the unsplit path → `main_test.go` red. (f) Run `golden -review` through the tool after
   editing one help line → the golden is rewritten, proving the environment reaches the unsplit
   path. (g) `main_test.go` runs a fixture package whose test fails unless a variable set by the
   caller is visible → proves it for the split path. (b) Plant a failing
   test in a split package → red, its output printed. (c) Name a package in the table that is not in
   scope → error. (d) Give `-deps` a target that resolves to nothing → error.
8. `go -C scripts run ./check` on a quiet box: read the FAIL COUNT and each FEOV gate's seconds.
   First CI run on the pull request: job and step durations (→ B3), peak RSS of the race step.
9. Plant a vet-only defect in `internal/cli` → `feov-record:vet` red.
10. B4: the union of names across all processes equals `-test.list` for each split package, printed
    by the tool and asserted in `main_test.go` against a fixture package.
11. `FEOV_RELEASE_GATE=1 go -C $T test ./releasegate/fuzz/` — unchanged by this plan; run once to
    confirm nothing in the harness depended on the suite's process layout.

**All parts:** `go -C scripts run ./check` before each push; the auditor gate
(`/prosthetic-conscience:plan-audit`) on this document before any code.

### Rulings

Ruled by gblock, 2026-10-11:

1. **CI adopts `testsplit` as well as the local gate.** `check`, `golden` and the workflow call one
   command line.
2. **Part A ships alone, and is measured in CI before Part B is written.** Part B's targets (B1–B3)
   are re-read against that measurement and restated in this document if they move.
3. **The race gate is scoped per test, and the scope is enforced** (Part R). Asked as "it should
   only cover things that actually spawn goroutines"; chosen over per-package scoping and over
   leaving the scope. This replaces the 2026-09-03 rule that the gate races the whole shipped
   import graph on every pull request; that form moves to a nightly workflow of its own. Part R is built before
   Part B, and Part B is written only if the measurement after R still calls for it.
4. **The residue is accepted: nightly-only detection.** A start reached only through a function
   value, or made by a library call the table does not list, is raced by `race-nightly` and not on
   the pull request, so such a race is found up to a day after merge. Chosen over following
   function values, which selects most of `internal/cli`.

Ruled by gblock, 2026-10-12, on the measurements under "Outcome":

5. **Part B is built anyway.** Asked with the recommendation to drop it, since no gate still waited
   minutes on a single-core package after Parts A and R; chosen over dropping it.
6. **Part B is local only.** This reverses ruling 1. `check` and `golden` run `testsplit`; the
   workflow's suite step stays plain `go test`, and the tool splits nothing on a box of four cores
   or fewer.
7. **The OCR engine step is narrowed now**, as its own pull request (#1330), without a path filter
   that would skip it.
