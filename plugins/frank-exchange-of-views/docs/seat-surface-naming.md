# What naming the surface in the constitution actually buys

Three measurements. **The 2026-08-20 run is the current record**; the two earlier runs are kept below
because it is what the design decisions were made on and its confound is still instructive.

Raw cell reports are not committed; regenerate with `seatprobe -naming <arm> [-help-directive]`.

---

# Current record, 2026-08-20 — the scoped tree, and the refusal as teacher

Instrument: `cmd/seatprobe`, 9 dispatches — 9 boards, one model (haiku), ONE configuration: the
shipped constitution against the shipped tree. **There are no arms.** The naming question is
settled and the alternatives were deleted from the probe rather than left dispatchable; keeping a
superseded configuration runnable is archaeology with a command-line flag. The second model is
gone for the same reason — two models answered "does this replicate", it did, and haiku alone is
the floor the surface has to work at.

Four things changed under the instrument since 2026-08-19, and each moves what a number MEANS:

- **The tree is scoped to the seat.** There is no role level: `feov-record --help` returns the
  seat's own verbs. A seat types `finding`, not `lens finding`.
- **Identity is bound at `register`** and read back from the record; the seat states `--seat-id`
  once and never again. A write from an unregistered agent is refused; reads and `--help` are not.
- **The surface-discovery duty is constitutional.** It used to live only in debate.js's dispatch
  prompt, so a constitution that named no verb also never said where the verbs were.
- **SEEN(leaf) is NOT comparable to the 2026-08-19 figure.** The denominator grew when `show`
  subviews entered the per-role act set (blue: 23 acts then, 35 now). The 99.0% and the 52.5%
  below are fractions of different things.

## Result

| | 2026-08-19 haiku `none` | 2026-08-20 shipped |
|---|---|---|
| opened | 9/9 | 9/9 |
| SEEN(top) | 100.0% | **100.0%** (132/132) |
| SEEN(leaf) | 99.0% | 52.5% (158/301) — *different denominator* |
| reached-for | 32.3% | **29.0%** (70/241) |
| help/sitting | 1.0 | **1.89** |
| registered | not measured | **9/9, eight of them at call 1** |
| refusals | not measured | 30 in 186 calls (16.1%) |

## What it says

**The refusal is the teacher, and that is measured rather than inferred.** Two of nine seats never
ran `--help` at all — and both still received a help listing, because the tool prints the surface
on an unknown command or a missing flag. `docket` asked for help ZERO times, was refused NINE
times, and reached the highest leaf exposure of any blue seat (26 of 35). SEEN(top) is 100% across
every sitting whether the seat asked or not.

That is the scoped tree paying off in a way the old surface could not: when the root IS the seat's
surface, any refusal at the root teaches the whole of it.

**Registration is not a cost.** All nine registered, eight on their first call, none skipped it,
and no sitting was refused for acting unbound. The one seat that skipped registration entirely in
an earlier run did so only while the harness was pre-registering it.

**Exposure saturates at about two help reads.** help/sitting fell from 3.4 (an earlier run of this
same configuration) to 1.89 once the prompt named `register` as the first act — and SEEN did not
move (54.2% -> 52.5% leaf, 100% -> 100% top). The extra reads were re-reads.

## What it does not say

n = 9, one model, one run, one replicate per board. **Two changes landed together** — the scoped
tree and the constitutional directive — and this run cannot attribute between them. `reached-for`
fell 32.3% -> 29.0%, which is within the noise of nine sittings and is NOT evidence of anything.

**Three instrument defects were found and fixed during this measurement, and each had produced a
number that read as a finding about seats:**

1. `ReadViewReads` matched `(?:lens|merge|blue|bench)\s+show`, which the scoped tree deleted. Every
   sitting reported "no projection opened at all" while one had run `show` fourteen times.
2. The harness pre-registered each seat under the handle it would later dispatch it with, so
   `register` was not the seat's first act and one seat in nine skipped it entirely — reported as
   compliance.
3. The acting prompt said "the record tool is <path>". Three seats read that as a directory and ran
   `cd <toolpath> && ./record register`; two then concluded the infrastructure was missing and
   abandoned the sitting after 1 and 4 calls.

The anomaly pass — reading the scored report for INTERNAL CONTRADICTION before trusting any number
— found all three. None was caught by a test.

## The finding that is not about the surface

Both seats that abandoned their sitting in (3) had caused the error themselves, and **neither used
the friction channel.** Their constitution names it in the sentence directly beneath the surface
directive: *if you need something that is not there, that is a finding about the tooling, and the
help names the channel that carries it.* Neither re-read the path. Neither tried the absolute form
the prompt had given them two lines earlier.

For a seat that cannot reach its tool, the space of reasonable responses contained "declare the
environment broken and stop" and did not contain "report it on the channel provided". Friction
remains the one duty whose non-use is indistinguishable from a sitting that met none — which is
what the explicit `--none` form exists for, and neither seat used that either.

# Re-run, 2026-08-19 — two models, and the constitutions no longer name verbs

Instrument: `cmd/seatprobe`, 72 dispatches — 4 arms × 9 boards × 2 models (opus, haiku), one
replicate per cell. All four roles are represented (lens, merge, blue, bench) rather than two.

Three things changed under the instrument since the first run, and each moves what an arm MEANS:

- The constitutions were stripped of every command name, so `none` is now the SHIPPED bytes and
  `partial` is a CONSTRUCTED three-verb block appended to them. In the first run this was reversed:
  `partial` was the shipped file and `none` was a redaction of it. **The arm labels are stable; what
  they are relative to is not**, and the `partial` treatment in particular is much weaker now — three
  names in an appended block, against names woven through the shipped prose.
- Four verbs were split into eight (`verify`/`corroborate`, `close`/`carry`, and two subgroups), so
  the denominator grew: blue offered 18 verbs then and 23 acts now.
- Cobra now teaches on refusal for pre-handler errors too (`ExecuteRoot` + `seat.Taught`), so a
  missing required flag returns the verb's own help rather than a bare usage line.

## Two quantities, kept apart

**SEEN** is surface EXPOSURE — the verbs of the seat's own role that appeared in help output the
seat received. This is the 58/95/100 metric. **Reached-for** is USE. They are different questions
and the first run's headline figure answers only the first.

SEEN is measured per BLOCK, not per line, because the board stages ROOT help as setup material and
root's command list contains `friction` and `verify` — names that are also role verbs. A line-level
match credits every seat with two verbs before it acts, and flatters exactly the arms that open help
least. A block counts only when every name in it is one of the role's own.

## Result

| model | arm | opened | SEEN(top) | SEEN(leaf) | reached-for | help/sitting | 1st call | unmet |
|---|---|---|---|---|---|---|---|---|
| haiku | `none` | 9/9 | **100.0%** | 99.0% | 32.3% | 1.0 | 0/9 | 22/43 |
| haiku | `none+directive` | 9/9 | **100.0%** | 99.0% | 29.9% | 2.8 | 5/9 | 24/43 |
| haiku | `partial` | 8/9 | **88.9%** | 93.2% | 32.2% | 2.4 | 0/9 | 22/43 |
| haiku | `complete` | 6/9 | **66.7%** | 78.8% | 29.0% | 1.3 | 0/9 | 21/43 |
| opus | `none` | 9/9 | **100.0%** | 100.0% | 40.0% | 5.4 | 9/9 | 13/43 |
| opus | `none+directive` | 9/9 | **100.0%** | 100.0% | 39.9% | 5.8 | 9/9 | 13/43 |
| opus | `partial` | 9/9 | **100.0%** | 100.0% | 41.5% | 4.9 | 8/9 | 12/43 |
| opus | `complete` | 8/9 | **88.9%** | 95.1% | 37.9% | 4.9 | 9/9 | 13/43 |

`opened` is sittings that received a listing of their own role's verbs at all. SEEN(top) saturates
at one `<role> --help` call, so per sitting it is 0 or 100 and the arm figure restates `opened`;
SEEN(leaf) counts ACTS, so a subgroup child requires opening the subgroup and it does not saturate.

## What it says

**The direction of the first run replicates, and it was never an inversion.** Naming a partial list
lowered exposure then (`partial` 58% against `none` 95%) and lowers it now (88.9% against 100% on
haiku). The effect is far smaller, which is what a weaker treatment should do.

**`complete` is the worst arm on exposure, in both models.** The first run found that stating the
complete surface bought nothing over stating nothing, on the REACHED metric (t = 0.61). On exposure
it is not neutral but negative: 66.7% on haiku, 88.9% on opus. Told everything, the seat has no
occasion to open the page — and what it was told is a snapshot, while the page is the tree.

**The `--help` directive is now inert.** It took 95% → 100% in the first run. `none` now reaches
100% without it, so there is nothing left for it to add, and it adds nothing: exposure flat, unmet
22 → 24 on haiku, 13 → 13 on opus. It still moves the behaviour it names — haiku's first-call help
rate goes 0/9 → 5/9 — which is the first run's finding intact: **making a seat read the surface does
not make it use the surface.**

**Model choice moves USE; the arm moves EXPOSURE.** These are close to orthogonal on this data. Opus
reaches for ~40% of its surface against haiku's ~30%, opens help 5 times a sitting against haiku's
1, opens it as its FIRST act in 9/9 sittings against haiku's 0/9, and leaves 13 of 43 board
expectations unmet against haiku's 21–24 — at IDENTICAL exposure (both 100% in `none`). Opus is also
markedly less sensitive to naming: `partial` costs it nothing at all, where it costs haiku a sitting.

**Naming changes where the knowledge came from, not how much work got done.** Unmet expectations
barely move across arms within a model (haiku 21–24, opus 12–13). This is not a result that naming
is harmful to outcomes. It is the narrower claim the design rests on: both routes inform the seat,
and only one of them cannot go stale when the tree grows a subgroup nobody updated the prose for.

## What it does not say

One replicate per cell — the first run had three, and traded board coverage for them. A 9-vs-8
`opened` count is one sitting, and nothing here separates a real effect from a coin flip at that
size. What carries weight is the direction agreeing across two models and 9 boards, not any single
cell.

`opus-partial`'s `docket` board exhausted its turn budget mid-sitting on the first attempt and was
re-dispatched onto a REBUILT board ~30 minutes later; same treatment, different point in time.

The `complete` arm appends 11–20 verb names depending on role, which also lengthens the
constitution. Naming volume and prompt length are not separated here.

Reach remains a proxy. A seat that reached for 11 verbs has not thereby done better work than one
that reached for 6.

---

# Original run, 2026-08-15 — one model, and the shipped constitution named verbs

Instrument: `cmd/seatprobe`, model haiku, 24 dispatches — 4 arms × 2 boards
(`arithmetic`, blue seat; `audit`, merge seat) × 3 replicates.

## Why this was run

`internal/cli/seat/menu.go` makes the refusal the primary teaching channel on a measured basis:
"seats do not learn this tool from `--help`. Every one of them read it once or twice in twenty to
forty tool calls." The observation is real. The inference is not supported by it, because every one
of those nine sittings ran with a PARTIAL list of verbs already in front of the seat — 2 to 4 names,
counted against the constitutions the probe dispatches under:

| seat | constitution | verbs named | reachable |
|---|---|---|---|
| blue | `blue-researcher.md` | 2 | 18 |
| bench | `lead-judge.md` | 2 | 11 |
| merge | `red-auditor.md` | 4 | 16 |
| lens | `red-auditor.md` | 1 | 9 |

A partial list is a plausible answer to the question `--help` answers completely. "Seats do not read
`--help`" and "seats stop when the partial list runs out" produce the same number and want opposite
fixes.

## The arms

| arm | constitution |
|---|---|
| `none` | shipped text with every verb NAME redacted, situations left standing |
| `partial` | shipped text, unmodified — the condition every prior probe ran under |
| `complete` | shipped text plus the whole role surface, GENERATED from the cobra tree |
| `none+directive` | `none`, plus production's "read `--help` before your first act" clause |

**These are the 2026-08-15 arms.** `partial` meaning "shipped text, unmodified" is what makes this
run's `partial` the strong treatment and the re-run's the weak one — see the re-run's note above.

## Result

| arm | n | surface reached (of 17) | range | `--help` reads | tool calls | refusals |
|---|---|---|---|---|---|---|
| `none` | 6 | **8.83** | 6–10 | 0.33 | 26.3 | 3.33 |
| `partial` | 6 | **6.83** | 5–9 | 0.67 | 11.7 | 2.17 |
| `complete` | 6 | **8.33** | 7–11 | 1.17 | 16.7 | 3.00 |
| `none+directive` | 6 | **7.50** | 7–9 | 2.33 | 24.7 | 3.33 |

Welch t on surface reached: `none` − `partial` = +2.00 (t = +2.35); `complete` − `partial` = +1.50
(t = +1.83); `none` − `complete` = +0.50 (t = +0.61).

## A third channel was uncontrolled, and it co-varied with the arm

**Read this before the section below.** `record.SittingOf` derives a `Duty{What, How}` for each live
circumstance — the situation, its consequence, and the exact command that discharges it. It carries
the same fact the constitution's naming carries, arriving only when it applies, and it rides on ONE
projection: `worklist`. `show board` does not carry it.

That channel was never varied, never measured, and is not a constant:

| arm | `worklist` reads/cell | cells with ≥1 | `board` reads/cell | reach |
|---|---|---|---|---|
| `none` | 2.00 | 5/6 | 4.00 | 8.83 |
| `partial` | 0.67 | 3/6 | 3.17 | 6.83 |
| `complete` | 0.33 | 2/6 | 4.33 | 8.33 |
| `none+directive` | 1.83 | 4/6 | 2.67 | 7.50 |

A covariate that moves threefold with the treatment is not an oversight to be waved off — it is a
rival explanation wearing the result's clothes. **The `none` over `partial` advantage cannot be
attributed to the constitution's naming.**

It is not a clean mediator either, and the opposite over-claim is refuted by the same table:
`complete` had the FEWEST worklist reads and the second-highest reach, so "reading the worklist
raises reach" does not hold.

`ReadViewReads` now prints this beside `HelpUse` on every probe report, so no future run states a
naming effect without the channel that competes with it.

**The design finding here is independent of the experiment.** `board` is described in this tool's
own words as "the form a seat acts on" and is read 2.7–4.3 times a sitting. `worklist` carries every
duty and is read 0.33–2.00 times. The one channel that delivers situation-plus-verb at the moment it
applies is the one the tool steers seats away from — and `SittingOf` can name only 4 of blue's 17
verbs, 5 of merge's 17, 3 of bench's 14, 2 of lens's 10, because a duty is derived only where
omission already carries a mechanical consequence (a refusal, or a capture score). Every verb whose
omission is merely a quality loss — `line-of-inquiry`, `manifest-row`, `spot-check`, `verify`, `reproduce`,
`closing`, `regrade` — gets no line, and those are the verbs the probe boards were built to bait.

## What it says

**The status quo is the worst arm — CONFOUNDED, see above.** A seat given the partial list reached
fewer verbs than a seat given no names at all, and did roughly half the work getting there (11.7
tool calls against 26.3). The partial answer does not merely fail to help; it appears to terminate
the search. But those same cells read the duty-carrying projection a third as often, so the effect
is not separable from the constitution's naming on this data. What is established is that the
shipped condition came LAST on reach; why it did is open.

**Stating the complete surface buys nothing over stating nothing** (t = 0.61). Delivery of the verb
list is not what caps a seat at ~8 of 17.

**The `--help` directive moves the behaviour it targets and not the outcome.** Help reads rose 0.33
→ 2.33 and moved to the first or second tool call — the seat orients rather than recovers — and
surface reach did not follow. Making a seat read the surface does not make it use the surface.

So the discovery explanation is the weak one. Across every arm the ceiling sits near 8 of 17
whether the seat is told nothing, told a little, told everything, or ordered to go and read it.
Whatever caps the reach is not knowledge of what exists.

## The mechanism, read off the trajectories

The obvious reading of the `none` arm — take the names away and the seat falls back on `--help` —
is wrong, and the trajectories say so. The `none` arm read `--help` LESS than any other arm (0.33).
What it did instead, from `none-arithmetic-r1`:

    ERR · ERR · blue show · blue close-gap · ERR · blue manifest-row · blue position ·
    blue friction · blue revision · …

`blue close-gap` does not exist. The refusal answered with the surface — `RefuseUnknownVerb` embeds
cobra's own help — and the seat's next four acts are real verbs it had not used before.

**The attribution to the refusal is withdrawn.** Two of those four — `friction` and `revision` — are
exactly the verbs `SittingOf` names, and this seat had already called `blue show` before it guessed.
The refusal and the duty list are both live at that moment, and the trace cannot separate them. What
the trace does establish is narrower and still worth having: the seat never asked for `--help`, so
whatever taught it was not the channel the constitution tells it to use.

The `partial` arm, same board, same model:

    show board · ERR · --help · blue show · blue edit ×5 · blue manifest-row ×4 ·
    blue revision · blue prove ×2 · blue position · blue friction · stop

It read help once, worked the short list, and stopped at 22 calls with 2 refusals. It never reached
past what it had been told, so it never met the channel that teaches.

**The two design decisions are in conflict, and that much does not depend on the attribution.** The
refusal was built as the primary teaching channel; the partial naming in the constitution is a
bypass around it. A seat that has been handed four verbs has no occasion to guess a fifth, and
guessing is what opens the channel. The `partial` arm's 2.17 refusals against `none`'s 3.33 is the
same story in the one number the confound does not touch.

## What it does not say

n = 6 per arm, one weak model, two boards, and the ranges overlap on every comparison. The
`partial` deficit is suggestive, confounded by the duty channel above, and not established; it
wants replicates AND a controlled duty channel before it carries a decision on its own. The next
run varies the duty lines and holds the naming fixed, which is the inverse of this one: this matrix
varied the channel seats barely read and held fixed the channel they do. The elicitation arm (`seatprobe -ask`) is the instrument for the question this
raises — whether the unreached verbs are unperceived or perceived-and-declined — and it has not
been run against these arms.

Reach is also a proxy. A seat that reaches for 11 verbs has not thereby done better work than one
that reached for 6; `internal/seatprobe` says so at length and it stays true here.
