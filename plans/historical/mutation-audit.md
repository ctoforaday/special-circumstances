# The mutation audit — what it measured, what it cost, and why it was retired

> STATUS 2026-09-06: RETIRED, and the tool with it. Operator call, after the sweep was asked
> to justify its cost against the rest of this repository's gates and could not. §0 is the
> evidence; the rest is the record of what was built, kept because the *questions* it asked
> are still the right ones — see CLAUDE.md's coverage clause, which now asks them by hand.

## 0. Why it was retired (2026-09-06)

**It never found a defect here.** Two full sweeps on the day of the decision:
`prosthetic-conscience/tools` — 66 survivors, 88% killed; `gray-area/tools` — 70 survivors,
75% killed. Against that, this repository's other gates each named something specific to change,
repeatedly, in the same session: a roster citing a test that did not exist, a golden flipping
exit 0 to exit 2, three archaeology refusals, four plan-audit FAILs.

**A survivor was not the fact it read as.** A mutant is tried against its OWN package, so a
survivor means *survived its own package*. The worst-looking survivor in the
prosthetic-conscience sweep — `hookunit.go:71`, the flag deciding whether the secrets gate
scans an unparseable payload, i.e. the #211 bypass on the gate agent-guardrails is about — is
KILLED, by a test one package over. Fifteen of the sixteen packages holding survivors there
have importers, so most of that list was the same artifact. `-confirm` buys the true answer at
~8 minutes each: ~9 hours for the smallest module.

**It could not reach its own motivating defect.** Six operator flips cannot delete a table row,
and `internal/secrets` at 100% coverage with two of eight patterns deletable IS a deletion
finding. A deletion operator was written on the last day and immediately produced the one real
result of the exercise — ten unasserted fields of `gray-area`'s capture manifest row, filed as
#797. That finding is the argument for asking the question, not for owning an instrument that
asks it about 562 mutants at a time.

**What replaces it: nothing automated, deliberately.** The coverage clause in CLAUDE.md asks
the same question by hand, aimed at the code where a silently-wrong answer is the risk. A
targeted deletion takes a minute and answers about the thing you were already worried about.

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

### II.5 The operator set, and what it could not reach (2026-09-06)

Six flips — `&&`/`||`, `==`/`!=`, `>=`→`>`, `<=`→`<` — plus DELETION of one row of a composite
literal.

**The deletion operator exists because the tool could not rediscover its own motivating defect.**
`internal/secrets` at 100% of statements with two of eight patterns deletable is a *deletion*
finding, and nothing in the six flips removes a table row. Added 2026-09-06; the widening moves
the denominator, so survivor counts before and after that date are not comparable.

It paid on its first run. Sweeping `gray-area/tools`: 13 deletion survivors, 10 of them fields of
`buildRow` — the capture manifest row — including `SessionID`, `TranscriptPath`, `CapturedAt`,
`PromptID`, `StopHookActive` and `Effort`. Each can be dropped with the suite green, which means
nothing asserts the manifest carries it, which means it can silently stop being written and the
row will still read as a complete capture.

The candidate test is a masked line ending in a comma with balanced delimiters. It is over-eager
by design: a multi-line call's arguments qualify and mostly fail to compile, which the sweep
already discards. Import blocks are tracked by the scanner rather than inferred from the mask —
`"fmt",` and a row of a string-pattern table mask to the same bare comma, so no rule over the mask
can separate the two.

**Generated files are never swept.** Recognised by the `// Code generated ... DO NOT EDIT.` line,
which is the only marker the tools here share — `record.pb.go`, `schema_gen.go` and
`classes_gen.go` have no common suffix. A test written to pin protoc's output pins protoc, and
`record.pb.go` alone holds 731 operator sites in the largest and slowest module.

### II.6 Why a release cannot gate on this (measured 2026-09-06)

A mutant is tried against its OWN package. So a survivor means *survived its own package*, never
*survived the module*, and `-confirm` buys the wider answer at ~8 minutes each.

Measured over `prosthetic-conscience/tools`: 66 survivors. The worst-looking by a distance is
`hookunit.go:71` — the flag deciding whether the secrets gate scans an unparseable payload, i.e.
the #211 bypass on the one gate agent-guardrails is about. It is KILLED, by
`TestMalformedPayloadIsScannedNotWavedThrough` in `internal/secretsgate`, one package over.
Fifteen of the sixteen packages holding survivors there have importers, so most of the list is
the same artifact.

A release asked to explain that list in writing would file explanations that are not true — an
allowlist wearing a schema, which is the failure the explaining was supposed to prevent. Buying
a true list costs ~9 hours for the smallest of the three modules.

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

**ANSWERED 2026-09-05 by `-confirm`: all 21 are killed by the wider module. 0 survived.** So the
19 were an artifact of the narrow stage exactly as suspected, and `anchor.go` IS asserted — from
`internal/record` and `internal/cli`. The alarming number retires.

What survives it is milder and still true: `anchor.go` has no test in its OWN package, so anyone
refactoring `internal/anchor` in isolation gets no signal from the package they are editing and
must run `internal/record` and `internal/cli` to learn anything. That is a fact about where the
tests live, not about whether the code is asserted — and plain coverage would have shown it faster
than mutation testing did.

## V. Swept so far

| target | result | reading |
|---|---|---|
| `internal/record/citationid.go` | 32 mutants, 31 behavioural, **100% killed**, 1 non-compiling | the citation-id machinery is fully killed by its own package |
| `internal/record/refs.go` | 50 mutants, 50 behavioural, **100% killed**, 0 non-compiling | the other half of the citation machinery, same result — §V.7's original target is now swept whole |
| `internal/claimcount` | 1 survivor of 2 | equivalent mutant: `j >= 0` -> `j > 0` on a `strings.Index` result, where index 0 cannot occur for a `-->` closer preceded by content |
| `internal/anchor`, narrow | 19 survivors of 21 | an artifact of the narrow stage — see §IV |
| `internal/anchor`, `-confirm` | 21 behavioural, **100% killed**, 0 survivors | settles it: the module kills every one; the file is asserted from `internal/record` and `internal/cli` |

## VI. Not yet swept

- `internal/cli/blue/cite.go`, `internal/cli/lens/anchor.go` — the `internal/cli` suite runs 20+
  minutes ONCE, so per-mutant cost there is the open question the pool has not answered.
- `internal/cli`'s citation files (`blue/cite.go`, `lens/anchor.go`) remain — that package's suite
  runs 20+ minutes ONCE, so its per-mutant cost is the open question the pool has not answered.

**Done since:** the `-confirm` wide stage on `internal/anchor` (§IV, §V). Measured cost, which is
the number to plan the next wide stage from: **21 mutants in ~19 minutes on four workers**, most
paying 50-70s and one outlier at 642s. The pre-pool estimate for this was ~2.5 hours.

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
