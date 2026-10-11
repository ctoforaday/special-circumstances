# Faster test loop

Status: proposed 2026-10-11, not started. Two independently shippable parts; Part A ships first and
does not wait on Part B. Related: #1013, #1046 (both measure the same gate from the outside).

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
| B1 | Local `feov-record:race` gate, 16 cores, load ≤ 10 | red by timeout | green, at most 600 s |
| B2 | Local `feov-record:test` gate | 378 s (peer, 4 cores) / 215 s | at most 90 s |
| B3 | CI `feov-record (ubuntu)` job | 1,459 s | at most 780 s |
| B4 | Tests executed | the package's `-test.list` | the same set, every name with a verdict |

A target missed is reported with the measured figure; it is not grounds to widen the change.

### Value case

The naive answers are more cores (already tried: the critical path is single-threaded) and a longer
timeout (already tried: 10 → 30 minutes; the gate now fails at 30). Part A removes work outright, at
the cost of one test helper. Part B converts cores into wall clock without rewriting 900 tests, at
the cost of one dev tool and two workflow lines. Measured in a prototype with eight processes, all
green: `cli` 77 s → 21.7 s, `migrate` 205 s → 48.1 s.

## II. Technical Context

- Go 1.27, `GOTOOLCHAIN=local`. FEOV tools module `plugins/frank-exchange-of-views/tools`; dev
  tooling in the `scripts` module. Nothing here ships to a consumer; no `version` moves.
- The record is SQLite (modernc, WAL). A migrated record is a directory holding the database and
  its siblings.
- Gate set: `scripts/check/gates.go`, mirrored against `.github/workflows/hooks.yml` by
  `scripts/check/parity_test.go`. `scripts/golden`'s go leg is `go test -count=1 ./...`, which
  `check` subsumes under the module's `:test` gate (#626).
- Constraints: `-count=1` stays on every suite gate (`noCache`). The race scope stays the shipped
  binary's import graph (ruling 2026-09-03). No test is removed, skipped or made `-short`.
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
  remaplists_test.go       [MODIFY]  lines 336, 380 read the store
  testmain_test.go         [MODIFY]  removes the store's directory after m.Run
```

**Design.** A package-level store keyed by tarball name holds, per archive, a `sync.Once`, the
migrated directory (under one `os.MkdirTemp` root owned by `TestMain`), the `*Manifest`, and the
error. The first caller extracts and runs the real `Migrate(src, dst, Entries(), Options{})`, then
closes every handle. Every caller, the first included, receives **its own copy** of the migrated
directory under `t.TempDir()` and opens that. A test that writes to its record therefore cannot
reach another test's, and no call site has to be audited for writes. The manifest is returned as a
copy for the same reason.

`TestEveryArchivedRunMigrates` keeps its meaning: every tarball still goes through `Migrate`, and
its refusals are read from the stored manifest. A new test asserts A2 — after the sweeps, the
store's migration count equals the number of tarballs in `run-archive`, each at 1.

**Alternatives rejected.**

| Alternative | Why not |
|---|---|
| Hand every test the one shared record, read-only | Seven test files in the package match a write pattern (`record.Append(`, `.Exec(`); one missed writer corrupts every later test, silently |
| Merge the nine sweeps into one test | Loses one-name-one-claim; a failure no longer says which property broke |
| `t.Parallel()` across sweeps | Same total work on more cores; leaves the `-race` cost in place |
| Commit pre-migrated records as fixtures | The tests exist to prove today's `migrate` on yesterday's records; a stored result proves nothing |
| Cache across processes on disk | A stale cache is a green that tested old code — the defect `noCache` exists to stop |

**Consumer census** — the contract changed is `migrateArchive`'s behaviour (signature unchanged) and
the three in-package sweep calls. Command, run in the package directory:

```
grep -nE 'migrateArchive\(t|archivedRuns\(t\)|Glob\(filepath.Join\(dir, "\*|(migrate\.|[^a-zA-Z.])Migrate\(' *_test.go
```

| Site | Changes? |
|---|---|
| `migrateArchive` callers (19): `archivedgaps:25,98,118` `prechange:14` `remanded:17` `classmaterial:36,62,72,84,291` `benchoccasion:27` `reopenedm19:17` `everyarchive:23` `schemaepoch:22` `everyrender:47` `seatturncarry:109` `gapplace:401` `shardarchive:24,43` | No edit; each now receives a copy of the once-migrated record |
| In-package sweeps: `findingarea:27`, `remaplists:336,380` | Yes — read the store instead of calling `Migrate` |
| Sweep enumerators: `archivedRuns` (`remaplists:96`), three `Glob` sites (`benchoccasion:20`, `everyarchive:16`, `everyrender:39`) | No |
| Direct `Migrate` on a synthetic or chained source: `classmaterial:141,227,246,341` `correction:112,157,179,195` `gapplace:144,216` `seatturncarry:48` | No — not an archive, or deliberately a second migration |
| Direct `Migrate` on one archive: `archive:36,118,240,291` `repairregister:75` `avenuerename:57` | No — see scope note |
| Other packages extracting an archive: `internal/cli/prechange_test.go:18`, `internal/record/classmaterial_test.go:274` | No — they extract, they do not migrate |

**Scope note.** The full concept — "no archive is migrated twice in one process" — would also route
the six direct single-archive calls. They are left: `archive:36,118` migrate the same run twice on
purpose to compare the results, and the rest cost about a second each. A2 is therefore stated for
the helper, not for the package.

### Part B — `testsplit` runs a serial package as several processes

```
scripts/
  testsplit/main.go             [NEW]     the runner
  testsplit/main_test.go        [NEW]
  internal/testsplit/table.go   [NEW]     which packages are split, per module, each with its reason
  check/gates.go                [MODIFY]  FEOV's :test and :race gates invoke testsplit
  check/main.go                 [MODIFY]  expandScope moves to the tool; subsumeGoldenLegs unchanged
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
3. Unsplit packages run as one `go test -count=1 [-race] -timeout D <pkgs>`, concurrently with:
4. Each split package: `go test -c`, then `<bin> -test.list '.*'`; `Test`, `Fuzz` and `Example`
   names are dealt round-robin into N lists; each list runs as `<bin> -test.count=1 -test.v
   -test.timeout D -test.run '^(a|b|…)$'` with the package directory as its working directory.
5. **Every listed name must come back with a top-level verdict.** A name with no `--- PASS`,
   `--- FAIL` or `--- SKIP` line fails the run as *not measured*. A process that exits non-zero
   fails the run and its output is printed whole.
6. `N` defaults to `min(8, NumCPU/2)`, at least 1; `-n 1` is the serial behaviour and is the
   comparison arm for §V.

The table (`internal/testsplit`) starts with three FEOV rows — `internal/cli`, `internal/record`,
`internal/record/migrate` — each carrying its measured serial seconds. It is a tuning table, not a
derived carrier: a missing row costs time and never coverage, because step 3 runs whatever step 4
does not. The tool prints each package's wall clock, so the next slow serial package is visible in
every report.

`check`, `golden` and the workflow all call the same command line; `parity_test.go` asserts the
workflow's line against the gate's args. Windows keeps `go test -short ./...` on pull requests: it is
262 s, off the critical path, and the tool is still built to run there (`.exe`, no shell).

**Alternatives rejected.**

| Alternative | Why not |
|---|---|
| Convert the packages to `t.Parallel()` | 69 `t.Setenv` and a package-level swap to unpick across ~900 tests; process isolation gives each split its own environment for free |
| Break the packages into smaller packages | Moves code and unexported access for a scheduling problem |
| Narrow the race scope | Ruled 2026-09-03; and #1046's 35 packages are not the time — two packages are |
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
| R3 | A failed migration is cached and every later test reports the same error | L × L | Intended: one cause, stated once per caller with the archive's name |
| R4 | Two processes of one package collide on state outside `t.TempDir()` (a fixed path, a port, a cache directory) — a concurrency no package has with itself today | M × H | Prototype was green once for `cli` and `migrate`; `record` and every `-race` arm are unmeasured. §V.6 runs each split package ten times at N=8 before merge; a collision is fixed at its source, never by lowering N |
| R5 | A test silently runs in no process | L × H | Step 5: every listed name needs a verdict; §V.7 deletes a name from a list and expects red |
| R6 | A hidden order dependency between tests | M × M | Same ten runs; round-robin keeps relative order inside a process |
| R7 | N race-instrumented processes exhaust memory on a 4-core runner | M × M | Default N is 2 there; §V.8 reads peak RSS from the first CI run |
| R8 | Workflow, `check` and `golden` drift onto three command lines | M × M | One tool, one table; parity asserts the workflow line against the gate args |
| R9 | `go test -c` skips the vet that `go test` runs by default | M × L | The `:vet` gates cover every module; §V.9 confirms by planting a vet failure |
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
   reader red, or A2's isolation assertion does.
5. **Real data.** The 19 archives are captured runs, not synthetic fixtures; step 2 is the driveable
   check on them. After merge, read `internal/record/migrate`'s line in the CI race step → A3.

**Part B**

6. `go -C scripts run ./testsplit -module $T -n 8` and the same with `-race -deps ./cmd/feov-record`,
   ten times each, on a quiet box → ten greens each. → B1, B2 against `-n 1` runs interleaved.
7. **Mutation checks.** (a) Drop one name from a list → *not measured*, red. (b) Plant a failing
   test in a split package → red, its output printed. (c) Name a package in the table that is not in
   scope → error. (d) Give `-deps` a target that resolves to nothing → error.
8. `go -C scripts run ./check` on a quiet box: read the FAIL COUNT and each FEOV gate's seconds.
   First CI run on the pull request: job and step durations (→ B3), peak RSS of the race step.
9. Plant a vet-only defect in `internal/cli` → `feov-record:vet` red.
10. B4: the union of names across all processes equals `-test.list` for each split package, printed
    by the tool and asserted in `main_test.go` against a fixture package.
11. `FEOV_RELEASE_GATE=1 go -C $T test ./releasegate/fuzz/` — unchanged by this plan; run once to
    confirm nothing in the harness depended on the suite's process layout.

**Both parts:** `go -C scripts run ./check` before each push; the auditor gate
(`/prosthetic-conscience:plan-audit`) on this document before any code.

### Rulings

Ruled by gblock, 2026-10-11:

1. **CI adopts `testsplit` as well as the local gate.** `check`, `golden` and the workflow call one
   command line.
2. **Part A ships alone, and is measured in CI before Part B is written.** Part B's targets (B1–B3)
   are re-read against that measurement and restated in this document if they move.
