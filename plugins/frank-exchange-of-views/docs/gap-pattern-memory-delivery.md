# Does red's gap-pattern memory work, and how should it be delivered?

Three questions were asked of the corpus, and they have three different answers. Keeping them
apart matters, because the corpus is good and that is not the same as the corpus working.

## 1. Is the corpus any good? — YES, and this is not in doubt

55 of 57 curated gap records carry a class; 36 distinct classes; entries name a mechanism and a
dated instance rather than a verdict, which is the shape `red-auditor.md` asks for (memory carries
QUESTIONS, never ANSWERS). Nothing below is a criticism of the material.

## 2. Is it delivered? — NOW YES, AND IT WAS NOT

`setup.MirrorGapPatterns` discarded the error from `os.WriteFile` and returned
`Written: true, Files: 55` having written nothing when `inputs/` did not exist. Production was
saved only by call ORDER — `run.go` happens to call `BuildSkeleton` first, which creates the
directory. Any caller that did not was told 55 files were mirrored into a path with no file in it.

The same class as everything else in this sweep: the failure and the success return the same
bytes. Fixed at 1.52.0 with a test that fails both ways (written-with-no-file, and
not-written-with-a-file).

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
single cell, and the strongest thing this run showed about the FILE form is that a seat which
cannot reach its board will use the corpus to fabricate one.

Before spending on a blue corpus, the cheaper question is whether blue reaches its board reliably
at all — this run says 4 of 6, with a now-fixed cause. Re-run the arms after the marker fix and
the comparison is at least about memory rather than about navigation.
