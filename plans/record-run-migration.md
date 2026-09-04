# record.Run — migrating 43 functions off `runDir string`

> STATUS 2026-09-02: shipped — historical record (the two §III threads remain open, tracked below).

Status: **landed**, across #612, #621, #631 and #633 (re-landed as #641), with §III naming the four signatures that
keep a string on purpose. This file is the thread, because the change spanned more pull requests
than one context holds and a PR boundary is not a completion boundary.

**#633 did not reach `main` on its first merge, and the failure is worth keeping.** It was
stacked on #631's branch `record-run-read-leaves`. #631 merged to `main` at 17:00; #633 merged
at 19:47 INTO that branch, which by then was already merged and stale — so GitHub marked #633
MERGED and `main` never received a line of it. A stacked pull request whose base branch is
merged but not DELETED is not retargeted; the second merge lands somewhere nobody reads. The
tell is cheap and nothing checked it: `git merge-base --is-ancestor <merge-commit> origin/main`.
(CHECKED NOW, 2026-09-02: `scripts/stackguard` + a hooks.yml job guard exactly this, added
2026-08-29 with #633's re-land.)

## I. What the type is for

`record.Run` is a run directory that RESOLVED. A `runDir string` admits three states it should
not, and each has a filed incident behind it:

| Admitted state | What it cost |
|---|---|
| A path nobody dispatched | #526 — `show board --run "<typo>"` printed an empty board for a directory that never existed. An empty board is the one answer that is never true. |
| Two spellings of one run | #358 — the handle cache had to be keyed on the RESOLVED directory, as a rule each caller had to remember. |
| The empty string meaning two things | "nobody supplied a run" and "you were refused" in one byte, with the healthy reading winning by default. |

A `Run` cannot hold any of them: constructible only by resolution, zero value is not a run, and
not a string — so it cannot reach `filepath.Join` without a caller saying `Dir()` out loud.

**`OpenRun` checks the RUN directory, not the record directory.** The first draft checked
`records/`, and it was wrong in a way worth keeping written down: a write CREATES that directory
(`store.go`'s `MkdirAll`), so the check conflated *never dispatched* (a typo — refuse) with
*dispatched, nothing filed yet* (round 0 — an empty board that is honestly empty). Five tests
across fetch, friction and the operator read failed on runs that were entirely real.
`NewRun` is the escape hatch for `setup`, whose `BuildSkeleton` creates the directory itself.

## II. Done

- `internal/record/run.go` + tests — the type, both constructors, the refusal wording.
- `internal/cli/seat` — `Context.runDir` is PRIVATE; `Run() (record.Run, error)` is the only
  way to a path, so the refusal is unreachable-past rather than merely carried.
- ~30 call sites the compiler enumerated across `cli`, `cli/blue`, `cli/lens`, `cli/merge`,
  `cli/motion`, `cli/bench`.
- Two live defects fixed on the way, both of the same shape:
  - `blue/claimindex.go` answered a REFUSED run by inferring a different one. Regression test:
    `TestClaimIndexHonoursARefusedRunInsteadOfInferringOne`.
  - Six read verbs carried a SECOND `InferRunDir` fallback below `seat.Of`'s own — reachable
    only on a refusal, and therefore only ever wrong. Removed.
  - `cli/friction.go` told operators who HAD supplied `--run` that `--run` was required.

### `setup` (second step)

All seven exported functions take `record.Run`; the orchestrator resolves ONCE via `NewRun`
before anything is written, and `cmd/seatprobe` opens its probe run with `OpenRun` (the
directory exists by then, so the stronger constructor is the right one).

Two things fell out:

- **`absRun` is gone.** It was `filepath.Abs(cfg.RunDir)` computed sixty lines below first use,
  with `if absErr != nil { absRun = cfg.RunDir }` — a silent fallback to the relative path. Only
  two sites used it, so a run was laid out in a mix of spellings. `run.Dir()` is absolute by
  construction.
- **An unresolvable run used to build completely and report success.** Measured against the
  pre-migration code: a run directory carrying `.records-elsewhere` with no pointer and no
  `FEOV_RECORD_ROOT` — a copied run directory, or a cleaned cache — made `run-setup` **exit 0
  with empty stderr**, having written the skeleton, mirrors, run-live marker and class join.
  `StageClassRegistry`'s "cannot resolve the record directory" appeared as a reason string
  inside the success summary. `BuildSkeleton` had swallowed the resolution error deliberately
  (`if recDir, err := record.RecordsDir(runDir); err == nil`) on the grounds that laying out a
  directory is setup's job. Regression test:
  `TestAnUnresolvableRunIsRefusedBeforeTheSkeletonExists`, null-run against the old code.

**Left alone deliberately:** `runlive.WriteRunLiveMarker` still receives `cfg.RunDir` verbatim,
not `run.Dir()`. The marker stores the path as given and `SameRun` normalises when comparing, so
switching it to absolute would change what lands in `run-live.json` for every reader of that
file — a different concept with its own carriers ([[facts-are-fields]] clause 4). Worth doing;
not worth doing silently inside this one.

### The packages above `record` (third step)

`view`, `report`, `dashboard`, `scorecard`, `cost`, `consistency`, `capture`, plus
`seat.RequireRun` and the CLI verbs behind it. 34 files, +235/-190.

**This was planned as "the read-heavy leaves" and they are not leaves.** `view` is a shared
dependency, so changing `view.Telemetry` and `view.Markdown` propagated through eight packages
to the CLI boundary in one connected component. There was no smaller green cut once those two
signatures moved. Scoping a step by "which packages sound peripheral" does not survive contact
with the import graph; scope it by what the compiler can stop at.

**Two refusals added where nothing had ever validated the path.** `capture <runDir> …` and
`dashboard <runDir> …` take the run as a BARE ARGUMENT — the hook's injection reaches `--run`,
not `argv` — and passed `args[0]` straight through.

Measured by null run, and NOT what I first wrote here: neither verb reported an empty board.
Both proceeded and failed further down on errors that name the wrong thing — `capture` on a
missing `journal.jsonl` in the TRANSCRIPT directory, `dashboard` on being unable to write
`dashboard.html` into a directory that does not exist. The defect is a misdirected error, not a
plausible zero, and the distinction matters: I had written the stronger claim before the null
run produced the evidence, which is the same order of operations this rule exists to forbid.

This is #526's shape in the one place seat-resolution could not reach, since the value never
passes through `seat.Context`. Regression test:
`TestAPositionalRunDirectoryNobodyMadeIsRefused`. Its limit is worth stating: `OpenRun` refuses
a path that does not EXIST, so `--run /tmp` — a real directory that is not a run — still passes.

`seat.RequireRun` now returns a `record.Run` instead of a string, keeping its verb-named
"`--run` is required" for the unsupplied case and adding the existence refusal beneath it.

**`internal/record/runtest`** exists because `recordtest` cannot import `record`: files in
`package record` import `recordtest`, so the reverse edge would make record's own test binary a
cycle. `runtest` is imported only by tests of other packages.

### What the mechanical rewrite could not see

54 test call sites were wrapped by script. One could not be, and it is the instructive one:
`blueRows("")` passes the empty string DELIBERATELY — it means "no run", and it is the case
that test exists to cover. Wrapping it in a resolver would have fataled the fixture on exactly
its own subject; it became `record.Run{}`. The script had no way to tell that argument from the
53 others, and it surfaced only because the compiler stopped there.

### `internal/record` itself (fourth step)

77 signatures, plus the packages that call them and every test call site. `Identity` carries a
`Run` instead of a `RunDir string`.

**The defect it surfaced: `RegisterSeat` guarded that the run directory exists; `Append` never
did.** So `Identity{RunDir: ""}` reaching `Append` resolved `""` through `filepath.Abs("")` to
the WORKING DIRECTORY and wrote the event into `<cwd>/records` — the "second blackboard beside
the real run, and reported success" incident that `RegisterSeat`'s own refusal text describes,
still live on the write path. And `""` was reachable: `seat.Context.Identity()` read `c.runDir`,
which is empty exactly when the run was REFUSED, and five verbs call `.Identity()` without
resolving first. A `Run` cannot hold the empty string, so the state stops existing; `Append`
asserts `Valid()` anyway, because a zero `Run` can still arrive from a struct literal.

### CACHING THE RESOLUTION IS THE PART THAT NEEDED CARE, AND IT BROKE TWO THINGS

The type's premise — resolve once, at construction — is sound only while the INPUTS to that
resolution hold still. Twice they do not, and both were caught by tests rather than by review:

1. **The separation marker appears mid-flow.** The seat probe separates a run by letting the
   FIRST `register` subprocess adopt a root and write the marker, so the marker arrives AFTER
   the caller's handle was made. A handle resolved before it carries `<run>/records`, and
   `StageForRun` then wrote the class registry where the tool would never read — "no gap-class
   registry is staged" on every board, and `records/` left inside the run. `seatprobe.Build`'s
   own comment already warned that this ordering is load-bearing; caching the resolution made it
   load-bearing a second time, in a way ordering could no longer protect. `StageForRun`
   re-resolves deliberately and says why — it is the one function that straddles the moment
   separation is established.
2. **A root deleted after the handle was made.** Resolution moving to construction meant a
   SEPARATED run whose root vanished afterwards read as an honest empty board. `dashboard
   --watch` holds a handle across regenerations, so the window is real. `Run` now carries
   `separated`, and `openRunForRead` re-checks that case — the one place where a missing record
   directory means a lie rather than a fresh run.

3. **The cache itself became shared state.** Found when #633 was re-merged onto a `main` that
   had since taken #630 phase 3, which made the fuzz's seats concurrent. The fuzz runner
   resolved its handle LAZILY — `if !r.runHandle.Valid() { r.runHandle, _ = NewRun(...) }` —
   which was single-threaded when written and races once `envelopeFor` runs on a goroutine per
   seat without holding `r.mu`. Neither branch was wrong alone; the defect exists only in the
   merge, so no test on either side could have caught it. The runner resolves in its constructor
   now, and `runHandle` is the one field the mutex does not guard BECAUSE it is never written
   again.

   **The null run is the part worth keeping, because the first one FAILED TO REPRODUCE.** The
   lazy version under `-race` through two full debate runs reported no race at all. That is not
   evidence of safety and was nearly read as such: the window is narrow — the first seat to
   arrive writes before its siblings read — so the fuzz's own scheduling is not a detector for
   it. Eight goroutines calling `run()` with nothing else to do reproduced it immediately (read
   at the `Valid()` guard, write on the next line), and the same probe is clean against the
   constructor version. `TestTheRunHandleIsImmutableAfterConstruction` is that probe, and
   `hooks.yml` runs THAT ONE TEST on the race leg — the feov module races only
   `./internal/record/`, so left out it would have been a guard that cannot fire.

**The general rule, which the plan did not have before:** a cached resolution is a snapshot.
Where the thing resolved can change under it — a marker written by a subprocess, a root deleted
by an operator, an environment that differs across a process boundary — the read must re-resolve
or re-check, and must say which it is doing. And where the cache is filled LAZILY, the snapshot
is taken at a moment the caller does not control, on whichever goroutine arrives first: prefer
resolving at construction, where the timing is stated and the field can be immutable.

### Two more the mechanical pass could not see

- **Function TYPES.** `flags.Checker` and a table of minters in `idnamespace_test` are function
  types; Go requires exact parameter identity, so no call-site rewriting reaches them. The three
  flag validators keep `runDir string` because `flags` cannot import `record` (the reverse edge
  exists) and resolve at entry instead — the refusal is gained even though the parameter was not.
- **Fixtures whose paths are deliberately not on disk.** `HarvestClasses` and
  `HarvestPrecedents` take only `slugOf(run)`, and their fixtures pass synthetic paths like
  `/runs/run-alpha`. Resolving those with the existence check fatals the fixture on its own
  subject — the same shape as `blueRows("")`. They use the non-checking constructor.

### What the tooling got wrong, since the tooling is the method here

- Driving a fixer off `go vet` reports only the FIRST type error per package, so "4 errors"
  meant four PACKAGES and every fix revealed the next one behind it. `go test -c -gcflags=-e`
  gives them all; the flat count had looked like a plateau for ten passes.
- The fixer matched `variable of type string` but not `value of type string`, so every
  `t.TempDir()` argument was invisible to it and reappeared each pass untouched.
- It wrapped arguments in scopes with no `*testing.T`, producing `undefined: t`.
- It compared source text to the compiler's NORMALISED echo (`20 * time.Second` vs
  `20*time.Second`), silently declining twelve matches while reporting "skipped 8". **A tool that
  counts only the refusals it anticipated reports a plausible zero for the rest** — which is the
  defect class this whole workstream exists to remove, reproduced in the instrument.
- An argument splitter that understood `"` but not backticks mis-parsed a raw string containing
  `{`, ate a paren and mangled a function's closing braces. Found by running `gofmt -e` over
  every touched file, not by the compiler, which had stopped earlier for another reason.

## III. Remaining

**The migration is done, and what is left is named rather than implied.** Every package above
`record` holds a handle, `record` itself takes one, and `Identity` carries a `Run`. Four
signatures keep a `runDir string` ON PURPOSE, and each has a reason that is not "not yet":

- **`record.RecordsDir`** — it BUILDS a `Run`. The sweep migrated it and made it circular.
- **`record.GapExists`, `InquiryExists`, `CitationExists`** — the three `flags.Checker`
  validators. `flags` cannot import `record` (the reverse edge exists), and Go function types
  require exact parameter identity, so no amount of call-site rewriting reaches them. They
  resolve via `OpenRun` at entry, so the REFUSAL is gained even though the parameter was not.

**Two threads this migration deliberately did not pull, carried here so they are tracked rather
than remembered** ([[complete-the-concept]]: a PR boundary is not a completion boundary):

1. **`runlive.WriteRunLiveMarker` still receives `cfg.RunDir` verbatim**, not `run.Dir()`. The
   marker stores the path as given and `SameRun` normalises when comparing, so switching it to
   absolute changes what lands in `run-live.json` for every reader of that file — a different
   concept with its own carriers ([[facts-are-fields]] clause 4).
   **NOT DONE (2026-09-02):** `setup/run.go` still passes `cfg.RunDir` to the marker.
2. **The flag-name constants belong in a leaf package.** That is the real fix for the three
   validators above: it removes the import edge that forces `flags` to speak in strings. It is
   larger than this change was, and is recorded here rather than implied by their comment.
   **NOT DONE (2026-09-02):** `flags.Checker` is still `func(runDir, value string) error`;
   the names live in `internal/flags/names.go`, not a leaf `flags` can share with `record`.

## IV. Order

`setup` first among the writers (it is the only `NewRun` caller, so it proves the two-constructor
split earns its keep), then the read-heavy leaves (`verify`, `view`, `graph`, `report`,
`dashboard`), then the hook packages, then `record`-internal.

## V. Validation loop

Re-armed by: any edit under `plugins/frank-exchange-of-views/tools`.

```
cd plugins/frank-exchange-of-views/tools
export PATH=$PATH:/usr/local/go/bin
go vet ./...
TMPDIR=/dev/shm/feov-tmp go test -count=1 -timeout 25m ./...
```

**`TMPDIR` is a 2G tmpfs, and a killed sweep does not clean up after itself.** Measured: 98
leftover scratch trees filled `/dev/shm` to 100%, and the next sweep reported `[build failed]`
for EVERY package in the module — including ones with no dependency on anything that had
changed — while `go build ./...` exited 0. A full disk presents as a compiler failure. Check
`df -h /dev/shm` before believing a module-wide build break, and clear `/dev/shm/feov-tmp`
between killed runs.

**Keep the sweep's FULL output; do not pipe it straight through a filter.** The same run was
first read through `grep -E "^(---|FAIL)"`, which drops the `# package` header and the
`file.go:line:` text under it — so the failure arrived as an unexplained wall of FAIL with the
one line that explained it filtered out. Redirect to a file, then grep the file.

**Run the FULL module, not `-run` filtered, after the LAST edit.** Measured in this arc: a
filtered re-run after edits missed an entire package of setup failures that CI then caught.
The fuzz targets need more than the 600s default locally (~690s), hence `-timeout 25m`.

For each defect this migration fixes, the regression test MUST be null-run: reproduce the old
behaviour, watch the test fail, restore, watch it pass. A test written after the fix and never
seen red is a test that proves nothing.

**And the null run is the only thing that catches a test which fails for the wrong reason.**
Measured here: the claim-index regression test asserted `err != nil` on a refused run, and
passed with the swallow REINSTATED. The swallow falls through to an inference that walks up
from the working directory — under test that is a package directory, not a run — so it yielded
`""` and the read failed on a missing file instead. Two entirely different failures, one green
assertion. `err != nil` is rarely a discriminating claim about a verb that has many ways to
fail; assert on WHICH refusal, in words only the intended path can produce.
