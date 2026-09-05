# The mutation audit — what it can measure, what it costs, and what it has swept

> STATUS 2026-09-05: in progress. The instrument is fixed and the first target is settled
> (`citationid.go`, 100% killed). Unswept: `record/refs.go`, the `internal/cli` citation files, and
> the `-confirm` wide stage. Split out of `plans/historical/red-citations.md` §V.7, which was filed
> as historical while this half was still being implemented from.

## I. Summary & goals

Coverage says a line RAN. It cannot say a test would have NOTICED had the line been wrong —
`internal/secrets` reported 100.0% of statements while two of its eight secret patterns could be
deleted outright with the suite green. `scripts/mutate` asks the second question by flipping one
operator and running the narrowest suite that covers it.

It is an **on-demand audit, not a CI gate**. There is no threshold: survivors are a list to
EXPLAIN, and the ones nobody can explain are the findings. Only `-selftest` runs in CI, and it
proves the tool can still mutate and observe at all.

## II. The operational contract (fixed 2026-09-05, PR #722)

- **Every mutant prints its verdict**, against a denominator computed before any test runs.
  Before this, only survivors printed — so a working sweep, a wedged one and a dead one produced
  identical bytes: none.
- **`-jobs N`** (default `NumCPU`) gives each worker its own tree. Jobs dispatch in order and
  results reassemble in order, so the report never depends on which worker finished first.
- **The sweep runs in a sandbox copy.** It never writes to the tree you are using — which also
  means anything you measure by watching the working tree is measuring nothing.
- **Progress is stderr, the report is stdout**, so piping the report to a file keeps it a document.

## III. The cost model, measured rather than assumed

On `internal/record`, whose suite is one of the slowest in the module:

| | figure |
|---|---|
| per mutant | **~4.0s** — ~1.5s compile and link, ~2.7s the package's own tests executing |
| `citationid.go` whole file, `-jobs 1` | 121.6s |
| same, `-jobs 4` | **61.8s** (identical verdict) |

Not 4x, because `go test` parallelises internally and workers contend for the same cores.

**Compiler flags do not help, and one hurts badly.** Measured 2026-09-05:
`-gcflags='all=-N -l'` recompiles every dependency unoptimised and the tests then run **4.3x
slower** (11.8s against 2.7s) for a 31s wall. Package-only `-gcflags='-N -l'` is a wash: the
per-mutant rebuild is already ONE package, deps stay cached, so there is almost no compile left to
save — and test execution, which dominates, only gets worse. The pool is the lever; there is no
compiler win hiding behind it.

## IV. Reading a result — the trap the first sweep fell into

The narrow stage runs a mutant against **its own package only**. So a survivor means "nothing in
this package noticed", which is NOT the same as "nothing noticed".

`internal/anchor` scored 19 survivors of 21, which reads as worthless tests. The cause: the package
holds only `window_test.go`, so `anchor.go` — 13 of the 19 — has **no package-local test at all**,
and is exercised from `internal/record` and `internal/cli` instead. Before reading a survivor count,
check whether the file has tests in its own package. `-confirm` settles it properly by re-running
each survivor against the rest of the module, at ~8 minutes per survivor.

**That finding stands on its own, and plain coverage would have shown it faster:** `anchor.go`,
which places every citation and finding anchor, asserts nothing about itself in its own package.

## V. Swept so far

| target | result | reading |
|---|---|---|
| `internal/record/citationid.go` | 32 mutants, 31 behavioural, **100% killed**, 1 non-compiling | the citation-id machinery is fully killed by its own package |
| `internal/claimcount` | 1 survivor of 2 | equivalent mutant: `j >= 0` -> `j > 0` on a `strings.Index` result, where index 0 cannot occur for a `-->` closer preceded by content |
| `internal/anchor` | 19 survivors of 21 | a test-coverage finding, not an assertion-quality one — see §IV |

## VI. Not yet swept

- `internal/record/refs.go` — the other half of the citation machinery. Cheap now (~4s/mutant).
- `internal/cli/blue/cite.go`, `internal/cli/lens/anchor.go` — the `internal/cli` suite runs 20+
  minutes ONCE, so per-mutant cost there is the open question the pool has not answered.
- The `-confirm` wide stage on `internal/anchor`'s 19, which would separate "tested elsewhere"
  from "tested nowhere". It would also benefit from the worker pool and does not yet use it.

## VII. A standing warning, because it recurred three times in one session

Every wrong answer about this tool came from **measuring the wrong thing and believing the result**:

1. A 50-minute silence was written into a merged plan as `INFEASIBLE-AS-BUILT`. Silence was simply
   what a killed mutant printed.
2. Correcting that, a rate of "1 mutation per ~150s" was measured by polling the file in the REAL
   tree — which the sandboxed sweep never writes to. That number described a load average.
3. A dead sweep was twice reported as "still running", by counting the reporting shell's own
   `grep` as the process.

None of the three was a property of mutation testing. Before concluding anything about this tool's
cost, check the machine's load and confirm the sweep is emitting progress.

## VIII. Verification

    cd scripts && go run ./mutate -selftest                    # the CI gate
    cd scripts && go run ./mutate -module <mod> -filter <path> # a scoped sweep, progress on stderr
    go test ./mutate/ -count=1

Re-arms: any change under `scripts/mutate/**`. A sweep's numbers re-arm on the target package's
tests changing — which is the point of running it.
