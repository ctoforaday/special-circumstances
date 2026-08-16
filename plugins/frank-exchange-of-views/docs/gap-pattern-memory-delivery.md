# Does red's gap-pattern memory work, and how should it be delivered?

Three questions were asked of the corpus, and they have three different answers. Keeping them
apart matters, because the corpus is good and that is not the same as the corpus working.

## 1. Is the corpus any good? — YES, but the first answer here was reached badly

**The retracted method.** This section originally read "55 of 57 carry a class; 36 distinct
classes; entries name a mechanism and a dated instance" and concluded *good*. That is a coverage
number and a shape check standing in for a judgement about whether an entry helps red decide
anything — the same substitution `CLAUDE.md` warns about where `internal/secrets` reported 100% of
statements while two of eight patterns could be deleted with the suite green. Classification
coverage cannot see instructional value. The verdict happened to survive re-reading; the method
would not have caught it if it hadn't.

**What re-reading the entries actually shows.** They are not a log of what red did. Almost every
one is a conditional with an ordered procedure — *"for each 'inherited/shared' cell in a
net-new-surface table, ask (1) was the baseline remediated since the cited incident? (2) does the
bespoke design restore authority the remediation removed? If yes to both, reclassify as net-new."*
That is decision guidance, and it is falsifiable at each step.

**What they are NOT, and this is the useful distinction.** The corpus is not guidance on how to
analyse an incoming report, and is not meant to be — the README says it is a source *compiled into
duty lines*. Red's analysis method lives in its constitution: leaf-node verification, graded trust,
`--check-kind`, estoppel, class-not-instance fix specs. The corpus is a set of pre-loaded
hypotheses to test, joined by class to the gap actually in front of a seat. Judging it as guidance
is a category error; the honest question is whether the join delivers.

**One measured caution about reach.** 41 of the 55 trigger clauses fire on a fresh report rather
than requiring a prior round — an earlier draft of this note claimed the opposite and was wrong.
But the subject matter is narrow: security controls, gates and guards, citation ledgers,
cost/token measurement, allowlists, deployment rungs, git state. All of it harvested from runs 2–5,
which audited this repo's tooling and security architecture. For a report outside that family the
class join returns little, and nothing measures how often that happens.

## 2. Is it delivered? — NOW YES, AND IT WAS NOT, TWICE

**Two unchecked writes, one commit apart, in one package.** Delivery is the part that was broken,
and the corpus's quality was never the issue.

`setup.MirrorGapPatterns` discarded the error from `os.WriteFile` and returned
`Written: true, Files: 55` having written nothing when `inputs/` did not exist. Production was
saved only by call ORDER — `run.go` happens to call `BuildSkeleton` first, which creates the
directory. Any caller that did not was told 55 files were mirrored into a path with no file in it.

The same class as everything else in this sweep: the failure and the success return the same
bytes. Fixed at 1.52.0 with a test that fails both ways (written-with-no-file, and
not-written-with-a-file).

**And the one that actually matters, found afterwards.** The mirror writes the flat file;
`inputs/gap-patterns-by-class.json` is the CLASS JOIN, which is what the engine hands a repairing
seat. Its write was `if b, err := marshalJSON(...); err == nil { os.WriteFile(...) }` — marshal
error skips the write, write error discarded — while the setup summary printed
`gap-pattern index: N class(es) -> inputs/gap-patterns-by-class.json` regardless, because N is the
in-memory index length and is never read back off disk. A run with no join and a run with a perfect
join printed the same line.

Setup already refuses a run whose corpus is mostly unclassified, in exactly these words: *"red
would open this run substantially blind while its memory directory looks full."* The gate guarded
the corpus and not the channel carrying it. Fixed at 1.54.0; the test is proven to fail against the
old code (exit 0, run built).

## 3. Does it change what a seat checks? — ONE RUN, n = 2 PER ARM, AND IT IS NOT ENOUGH

`red-auditor.md` asserts, as **measured**: *"duty-embedded patterns caught both warned classes in
round 1; the mounted file prevented nothing."* That is n = 1 per delivery form. This run is the
second look and does not settle it either. It is recorded here so the claim's evidence is
findable, not so the claim is upgraded.

Design: board `arithmetic` (whose gap classes `figure-recount-fails`,
`derivation-status-overclaim`, `self-attestation` draw 16 corpus entries), elicitation mode
(`seatprobe -ask` — the seat is asked what its options are, not watched taking them), three arms
× 2 replicates:

| arm | delivery |
|---|---|
| `none` | no corpus at all |
| `file` | corpus mirrored to `inputs/red-gap-patterns.md`, unmentioned |
| `duty` | corpus mirrored AND the pattern-duty clause in the sitting |

### What happened

| arm | tool calls | reached the board | distinctive output |
|---|---|---|---|
| `none` | 6, 9 | **2/2** | closure defined gap-by-gap |
| `file` | 13, 12 | 1/2 | pattern names offered as GUESSES at red's gaps |
| `duty` | 15, 10 | 1/2 | one cell added a CLASS-level completion test |

**The one thing the control never produced.** `duty-r2` set its own finishing condition as:
*"Red does not come back with new gaps in the same classes. If new gaps appear in
[figure-recount-fails], [derivation-status-overclaim], [self-attestation], or [false-universal],
then I missed something in this round."* Both `none` cells defined completion as closing the four
listed gaps. That is the duty clause doing the thing it is supposed to do — raising the seat's
unit of self-check from the instance to the class — and it is one cell.

**The one thing the mounted file produced, and it is not good.** `file-r2` never reached the board
and filled the hole with the corpus: it named `ephemeral-instrument`, `footnote-undercoverage`,
`verification-file-type-blindspot` and `live-source-drift` as gaps red had "almost certainly"
raised. None of them were on the board. `file-r1`, which did reach the board, read the mirror and
set it aside: *"a catalogue of pitfalls … not a fixing instruction for this report."*

So the corpus's failure mode, when it is present as a FILE, is that it is good enough to
reconstruct a plausible board from — which is exactly what a seat that cannot reach the real board
will do with it.

### The `file` arm was testing a form already documented as failed

This is the sharpest limit on the run and it was not stated when the results were first written up.
`README.md` in the corpus directory says, of the flat mirror: *"run 4's 'read red's accumulated gap
patterns' clause was unsatisfiable at four blue seats, and run 5 was worse — lanes verifiably READ
the file and committed both warned patterns anyway."* The `file` arm re-derived that. Presenting it
as a finding overstated it; only the mechanism is new — that a seat which cannot reach its board
uses the corpus to reconstruct a plausible one.

It also means the arm was not testing production. Production's operative channel is the class join,
which hands a seat only the patterns matching the gap in front of it; the flat mirror is the
secondary artifact. A three-arm comparison of `none` / flat-file / duty leaves the shipped channel
untested.

### What this does NOT support

- **"The duty beats the file."** Directionally consistent with the constitution's claim, at n = 2
  per arm, from one board, with one seat role. Two cells is an anecdote with a table around it.
- **"The file is harmful."** Both mirror arms cost ~1.7× the control's tool calls and both lost a
  cell to the board, but the board losses have a confirmed cause of their own (below) that is not
  the mirror, and 2 versus 4 cells cannot separate them.

### What it does support, and it is the reason the run was worth doing

The measure is not the problem this time; the **environment** was. See below.

## The defect this run actually found

`.records-elsewhere` — the marker the record separation drops in a run directory, whose entire job
is to tell a seat how to reach a board it cannot see — instructed:

    feov-record <role> show board --run <this directory>

The tool is not installed on PATH. Every dispatch prompt hands the seat an ABSOLUTE path. The
marker's line was a command that cannot execute, offered by the file that exists to unblock a
seat.

Three of six seats copied that form. Two recovered by hunting for the binary. One read
`feov-record: command not found` as **"that tool isn't available"**, and then answered the whole
sitting — options, costs, what would tell it it was finished — against a board it had never seen,
re-deriving the report's arithmetic by hand and guessing red's gaps from the pattern corpus.

Its output is not distinguishable from a sitting that read the board. Same structure, same
confidence, same four sections. The gaps are invented instead of quoted and nothing says so.

This is [[facts-are-fields]] at a marker: the invocation was COMPOSED into prose, from a fact the
`record` package does not hold, and recovered by a seat that had the real path in its prompt and
believed the file over it. The marker cannot carry the path — withholding it is the whole point of
the separation — so it must not compose the command at all. It now names the shape, points at the
handle the dispatch gave, and says out loud what `command not found` does and does not mean.

`recordroot.go` had already been corrected once for this exact file, for the `--view` rename. That
fix corrected the VERB and left the HEAD of the command alone.

Gated by `TestTheMarkerComposesNoInvocationTheSeatCannotRun`, which fails against the old text.

## Recommendation

**Do not build blue a second corpus on this evidence.** The one signal in favour of memory is a
single cell; the FILE arm re-derived a result the corpus README already records; and the shipped
channel — the class join — was not in the experiment at all.

Two cheaper questions come first, both of which this work turned up:

1. **Does blue reach its board reliably?** 4 of 6 here, with a now-fixed cause. A memory comparison
   run before that is measuring navigation.
2. **How often does the class join return nothing?** The corpus is narrow by construction — runs
   2–5, all auditing this repo's tooling and security architecture. For a report outside that
   family the join is thin or empty, and no instrument reports it. Setup refuses a corpus that is
   mostly unclassified; nothing notices a corpus that is fully classified and entirely
   inapplicable. That is the same shape one level out: an empty join and a well-matched join both
   produce a run that starts.

The thing red demonstrably has and blue lacked was never the corpus — it is a constitution dense
with method. Blue now has one. Whether either needs a second memory is a question worth asking
after the join is measured, not before.
