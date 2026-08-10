# pre-motion-real-run — provenance

**This record cannot be reproduced.** The binary that wrote it does not exist on any branch after
#344's destructive half: it writes `dispute`, `dispute-respond`, `petition`, `petition-rule` and
`avenue-rule`, and those verbs are gone. If this directory is ever lost, recover it from git
history — there is no way to make another.

## What it is

The real-data half of the #344 compatibility check. `plans/command-groups.md` §V requires two
artifacts with different jobs:

- `testdata/pre-motion-run/` — produced by `TestFuzzDebate`. The mechanical compat regression.
- **this** — produced by real model-driven seats. The only thing that catches prose the harness
  never generates.

§V's reason, verbatim: *"fixtures prove logic while only real data surfaces data-shaped defects —
fallback collisions, harness sentinels, encoding — and a harness-produced record is precisely the
category that cannot surface a harness sentinel."*

## How it was made

`research/2026-08-10_pre-motion-real-record`, 2026-08-10, against `feov-record` **0.47.0** — a
binary with no `motion` verbs at all. Repo pin `0783a2c`.

Topic: *Does a permanent dual-read of a retired event vocabulary cost less than a one-shot
migration, for a log whose readers ship to projects the author cannot see?*

It was **not** driven by `debate.js`. The operator played `red-merge-r1` — minting three graded
gaps — and then dispatched real agents into the seats:

| seat | agent | what it did |
|---|---|---|
| `blue-respond-r1` | blue-researcher | contested 3 grades, proposed 6 directions, petitioned on integrity |
| `red-merge-r2` | red-auditor | ruled all 3 disputes (2 accepted, 1 rejected), ruled all 6 directions |
| `judge-petition` | lead-judge | heard the petition, granted in part |
| `blue-respond-r2` | blue-researcher | answered all 6 rulings |

No seat was told what to argue. The provocation was structural — gaps graded high on a `suspected`
axis, and locations naming report sections that did not exist — and the seats found the second one
themselves, which is why there is a petition on the record at all.

**There is no `cost.md` and no `run-record-audit.md`.** `capture` requires a workflow journal and
this run had none. The record is the artifact; the run directory is not a captured deliverable.

## What it carries

63 events across 5 shards. All five retiring types: `dispute`×3, `dispute-respond`×3,
`petition`×1, `petition-rule`×1, `avenue-rule`×6.

Prose the fuzz does not produce: backticks, angle brackets, pipes, bracketed references, em dashes
and other non-ASCII, and a single `opinion` field of **13,445 characters**.

**It does NOT carry `contests_ruling`** — zero occurrences. The merge ruled a line too-thin, blue
argued against that reasoning at the leaf and then declined the line anyway, and the field cannot
represent disagreeing-and-yielding because it is a side effect of moving a line to `pursued`. That
absence is a finding recorded on #344, not a hole here. The fuzz fixture carries the field.

## What reads it

`premotionreal_test.go`. Four tests: the vocabulary is still present, the record still replays,
the dual-read recovers every exchange with its ask AND its answer, and the prose survives the
round trip byte for byte.

Those tests assert **shape, never verdicts**. Five real seats decided what to argue; pinning their
conclusions would make the test a transcript of one afternoon's opinions rather than a compat
check.
