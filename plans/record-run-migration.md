# record.Run — migrating 43 functions off `runDir string`

Status: **in progress.** The type exists and one caller-side package (`cli/seat`) has moved.
This file is the thread, because the change spans more pull requests than one context holds
and a PR boundary is not a completion boundary.

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

## III. Remaining

**43 functions in `internal/record` still take `runDir string`** — 26 exported, 17 unexported —
called from 24 packages:

```
cmd/seatprobe        internal/cli/enumhelp  internal/graph      internal/seatenv
internal/capture     internal/cli/lens      internal/hookcmd    internal/seatprobe
internal/cli         internal/cli/merge     internal/hookgate   internal/setup
internal/cli/bench   internal/cli/motion    internal/report     internal/verify
internal/cli/blue    internal/cli/seat      internal/scorecard  internal/view
internal/consistency internal/dashboard     internal/fetchcache internal/flags
```

`gopls` is not installed, so there is no rename-symbol sweep: each package migrates as its own
green step, driven by what the compiler refuses. Do NOT batch them — the value of the type is
in the refusals it forces at each site, and a sweep only compiles.

The 17 unexported ones move last and nearly for free: once every exported entry point holds a
`Run`, they are being handed a resolved directory already.

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
