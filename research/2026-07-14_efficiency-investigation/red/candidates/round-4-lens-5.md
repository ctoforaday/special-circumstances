# Red audit — round 4, lens 5: logic and completeness

Full re-read of `blue/report.md` (all 1535 lines, three windowed reads) plus first-hand leaf
verification at: the run directory's git state (`git status --porcelain`, `git ls-files`,
`git log -- trajectories/`), `debate.js` ll.55–60 (GRADE enum) and ll.205–220 (lens prompt
read surface), the pinned run-3 `red/findings.md` (awk byte-split recomputed), and this run's
round-3 `### LEAD` entry (debate.md ll.481–604). CHANGELOG used as navigation only.

**Lens verdict: FAIL** — the round-3 repairs are substantively faithful (every ruled sentence
landed; the big R3-1 correction is arithmetically clean end-to-end), but two repairs carry
regressions in their own new text, one freshly-decided design (the Q6 mapping pin) is not
total over the input population it maps, and two summary sentences fail the lead's own
round-4 bar (recomputed-from-the-printed-figures). Five findings, none above LOW-MEDIUM.

---

## L5-F1 — LOW-MEDIUM — certain (verified first-hand: `git status --porcelain` shows `??
trajectories/`; `git ls-files trajectories/` empty; `git log -- trajectories/decompose-merge.mjs`
empty) × low-medium (the #1-ranked lever's instrument attestation — the exact R3-1(b) class;
the §6.2-coherence argument now rests on a future commit event no stated obligation covers) ×
trivial — proposed lineage: supersedes R3-1 (closed WITH REGRESSION)

**Location:** §4.2 — *"The instrument is now **committed as `trajectories/decompose-merge.mjs`**
in this run's directory"* — propagated to [^MergeDecomposition] (*"The method is now committed
as **`trajectories/decompose-merge.mjs`**"*), §8 Q2 (*"The instrument is now committed"*), and
CHANGELOG round 3 (*"REBUILT and COMMITTED"*).

**Problem:** "committed" is false at audit time. The file exists in the working tree
(4,228 bytes, verified; header and method match the footnote, including the corrected
bytes→tokens step at its documented step 5) but the entire `trajectories/` directory is
untracked (`??`) and no commit touches it — the run's last commits are the staging pair
(`b162e50`, `5a08a0e`). The R3-1(b) defect was precisely that the measurement's audit
artifacts were "exactly the ones that are not git-tracked"; at this audit's timestamp the
replacement instrument is in the same state — one cleanup away from the round-2 scratchpad
condition. Mitigations, stated honestly: (i) mid-run, *nothing* in the run dir is committed
(all tracked files are modified-uncommitted), so blue plausibly cannot commit from its seat;
(ii) run-3 precedent (`trajectories/journal.jsonl` tracked at `bfa8a3b`, only the tarball
gitignored — .gitignore re-read first-hand) suggests the run-close sweep will capture it.
But no clause anywhere obligates the run-close commit to include it, and the sentence as
written invites exactly the check (`git log`) that refutes it. Vector honesty: red's R3-1
required fix and the lead's ruling both said "commit the parser into the run dir" — the verb
is red's; the leaf state is still the leaf state.

**Required fix:** restate at all three report sites: the instrument is *written to the run
directory at a git-tracked path* (only the tarball is gitignored) and *enters the object store
with the run-record commit* — plus one sentence making that inclusion an obligation of the
run-close commit (the same place the decomposition OUTPUT retention clause already lives), or
have the lead/operator commit it before close and cite the hash.

**found_by:** L5

## L5-F2 — LOW-MEDIUM — low this run (channel unbuilt, zero dispute traffic) rising to
medium-high under the actuation the interlock is mandatory for (sibling-consistent grading
with R3-3) × medium-high under actuation (an accepted, unlisted deflation flows into the mass
computation that sets red's lens budget — the harm the clause claims to prevent) ×
trivial-to-low — proposed lineage: supersedes R3-3 (closed WITH REGRESSION); fourth
generation of the clause-(v) chain R1-2 → R2-6 → R3-3 → this

**Location:** §3.3 clause (v), the round-3 repair text — *"pending-entry deltas are LISTED in
the round's `### RED` debate entry — a git-tracked surface every seat and the human operator
already read"* — and — *"an unlisted delta never enters mass (the listing is the precondition,
checkable against the telemetry line's delta record)"* — and the second guard's cumulative
threshold sentence.

**Problem:** three residuals in the repaired text, each a round-4-bar item (the lead: "no
design clause ships without naming who executes it and confirming the named channel can
physically carry it").
(a) **"Every seat … already read" is false for lens seats.** The engine's lens prompt
(debate.js l.212, re-read first-hand; confirmed by this lens's own dispatch) names
`blue/report.md` + CHANGELOG only — debate.md is not on the lens read surface (it is on
blue's, red-merge's marginally, the judge's when dispatched, and the lead/operator's). The
window has real watchers; it does not have the universal the sentence claims, and the claim
was broadened beyond red's R3-3 fix text ("already on blue's and the judge's read surfaces").
(b) **"Never enters mass" is policy without mechanism.** Mass is arithmetic over grades in
`red/findings.md`, computed by red-merge; an accepted delta changes the grade in the findings
record whether or not it is listed, so an unlisted delta enters mass by default — the clause's
guarantee has no executor. The stated check ("checkable against the telemetry line's delta
record") is same-writer consistency: the `### RED` listing and the telemetry line are both
red-merge outputs, so a coherent omission (delta absent from both) passes it. The genuinely
independent reconciliation exists and is unnamed: findings.md is git-tracked, so round-over-round
grade changes are diffable against the listed deltas — one sentence adding this reconciliation
to the §3.3(v)/§2.5 actuation-review duties (same family as the R3-7 recompute clause) closes
the hole at the tier §6.2 actually offers (post-hoc vacuity audit), and the clause should say
that is the tier it is buying.
(c) **The cumulative-magnitude auto-docket names no executor.** Plausibly the script CAN
compute it (old grade in the prior `redEnv.gaps`, proposed in `blueEnv.grade_disputes`,
acceptance in `redEnv.dispute_responses`, magnitude via the §8 Q6 pinned mapping — all
envelope-resident), which would make it a shape/consistency-tier check; but the clause does
not say who computes the sum or dockets the overflow.

**Required fix:** (i) restate the read-surface sentence to the seats that actually read it;
(ii) name the reconciliation: any actuation review MUST diff round-over-round grade values in
the git-tracked findings record against the listed deltas, and an unlisted accepted delta
disqualifies the affected rounds' series from the actuation case; (iii) one clause naming the
cumulative-threshold executor (the script, from envelope-resident values under the pinned
mapping — or wherever blue actually intends it).

**found_by:** L5

## L5-F3 — LOW-MEDIUM — medium (run 4's first logged round is the deadline the pin itself
names, and conditional/compound grades are corpus-normal — this run's own R3-3 grade is "low
this run … rising to medium-high") × low-medium (within-version ambiguity defeats the pin's
stated purpose — a comparable runs-4–5 series feeding the actuation decision — the same harm
the NEW-series rule exists to prevent) × trivial — proposed lineage: supersedes R3-8 (the Q6
decision is the R3-8 repair's own new text)

**Location:** §8 Q6 / §2.5 item 1 — *"The pinned mapping for runs 4–5's telemetry series:
low=1, low-medium=1.5, medium=2, medium-high=2.5, high=3, certain=3.5, realized=excluded"*.

**Problem:** the pinned mapping is not total over the input population it maps, on three axes.
(a) **The shipped GRADE enum has eight members** (debate.js l.60, re-read first-hand: `low,
low-medium, medium, medium-high, high, certain, realized, trivial`); `trivial` has no mapped
value. It is schema-legal as a likelihood or impact grade even if corpus convention reserves
it for complexity — an unmapped-token line is undefined, and undefined resolves by seat
convention, the thing the pin exists to remove.
(b) **The certain/realized boundary contradicts the exclusion's own rationale.** The corpus
defines `certain` AS realized: R4-1's pinned grading reads "certain (already realized in this
corpus, not projected)" and §2.1 item 4's own argument is "a text defect, once found, is
certain." The rationale for excluding `realized` — "realized risk is no longer a probability" —
applies verbatim to certain-as-realized, yet the mapping puts the two near-synonyms at
opposite extremes (3.5 vs excluded): a grader's word choice swings a gap by the mapping's full
range. Note the consequence for the metric's stated defect: §2.1 item 4's trivia-blindness
(certain×low nits scoring ~3.5) is *retained* by this pin while realized gaps drop out — if
the exclusion rationale were applied at the semantic boundary, it would also fix the
already-found-text-defect scoring §2.1 complains about.
(c) **Compound/conditional cells have no pinned reading convention.** §2.1's own two-lane
history demonstrates the sensitivity (lane extraction differences, ~65 vs ~62 at round 2, from
"compound cells read verbatim"), and conditional grades ("X this run rising to Y under
actuation") are this board's modal shape. Which branch enters the series is unpinned.

**Required fix:** three sentences in the Q6 pin: assign or exclude `trivial` explicitly; state
the certain/realized rule at the semantics, not the token (e.g. "a likelihood graded as
already-realized — whichever token carries it — is excluded from mass; `certain` counts 3.5
only for not-yet-realized certainty," or collapse the tokens and say so); pin the
conditional-cell convention (e.g. "the current-run branch of a conditional grade enters the
series; the conditional is preserved in the board-profile prose").

**found_by:** L5

## L5-F4 — LOW — medium (the run-end condition is plausibly realized — and per this report's
own §3.1/§6.1, adjudicated-but-open `carried` gaps are near-certain run-4 traffic, so the
mismatch is live, not hypothetical) × low (the error direction is conservative: the prediction
can fail to settle TRUE while the hardened arm would have armed and saved — under-recording
true firings delays a warranted build rather than falsely validating one) × trivial —
proposed lineage: supersedes R3-6 (composition residual in the restated prediction)

**Location:** §1.5 — arm condition (a): *"every **unadjudicated** open gap ≤ MEDIUM with
low/trivial fix cost"* — vs restated prediction condition (ii): *"a pre-final open board
all-≤-MEDIUM with low/trivial fix cost"*.

**Problem:** the R3-6 repair aligned the prediction's trigger cadence (two consecutive
zero-above-floor-mint rounds, red-health, disposed-not-carried caveat — all verified composed
this round) but dropped the arm's "unadjudicated" qualifier from the board condition. A
`carried` gap is adjudicated AND open; a pre-final board carrying an adjudicated MEDIUM-HIGH
satisfies arm (a) (which scopes to unadjudicated gaps only) while failing prediction (ii)
(which scopes to the whole open board). The prediction is therefore stricter than the arm it
registers evidence for — third-generation composition residual on the same paragraph
(R1-12/R1-17 → R3-6 → this), in the harmless direction this time, but a registered figure
that systematically under-counts the arm's firings still mis-feeds the §2.5-item-3
build-trigger record.

**Required fix:** one word — restate (ii): "a pre-final open board whose **unadjudicated**
gaps are all-≤-MEDIUM with low/trivial fix cost (adjudicated-open gaps — carried or
risk-accepted — sit outside the arm's scope by its own condition (a))."

**found_by:** L5

## L5-F5 — LOW — certain (recomputed from the report's own printed figures at this seat:
§4.6 item 2 prices batching at $1–2/round ≈ $5–10/run over five merge rounds; §6.1 item 1
prices sharding-addressable at $2–4/run — batching exceeds sharding ~2–3× on the measured
axis) × low (the #1 rank is *explicitly* re-based on the unpriced factors in the same
sentence, so no disposition turns on the word; but the gloss understates in the direction
that flatters the #1 lever) × trivial — new raise, R3-14's over-tight-gloss class (made a
named recurring class by red's own round-3 merge note), and a round-4-bar item (a summary
sentence about printed figures, not recomputed from them)

**Location:** §6.1 item 1 — *"sharding-addressable ≈$2–4/run at run-3 scale, comparable to
item 2's batching saving rather than dollar-dominant"*.

**Problem:** "comparable" is the wrong comparative. From the report's own figures the measured
ordering is batching > sharding by roughly 2–3× ($5–10/run vs $2–4/run) — the honest statement
is that on the measured-dollar axis sharding now ranks BELOW the item ranked under it, and #1
rests entirely on the disclosed unpriced factors (judge-read benefit, degradation-regime
quality, growth direction). The section already concedes the rank basis; the one word
"comparable" walks the concession halfway back.

**Required fix:** "…≈$2–4/run at run-3 scale — *below* item 2's ≈$5–10/run batching saving on
the measured axis; the #1 rank rests on the unpriced factors stated next" (or equivalent).

**found_by:** L5

---

## Checked and deliberately not raised (logic/completeness sweep)

- **Archive fraction 72.6% recomputed first-hand** at the pinned run-3 findings.md: awk split
  28,867 / 76,356 of 105,223 LF-normalized bytes = 72.57%; l.340 content matches the quoted
  boundary line exactly. The R3-1(a) repair's headline figure is clean.
- **The corrected dollar chain recomputes clean:** 0.26+0.53+0.89+1.16 = 2.84 ≈ $2.8;
  $4.10/$7.87 = 52.1%; $1.16/$7.87 = 14.7% ≈ "≈15%"; 72.6% × $3–6 = $2.2–4.4 ≈ "$2–4";
  72.6% × 23K = 16.7K ≈ "16–17K", below lane 1's 20K floor (R3-5 repair verified as restated).
- **R3-4/R3-10/R3-13/R3-14 repairs verified in place from the printed tables:** rounds 2/3/5
  largest + round-4 33/32/20 split at both sites; "× 3 throttled rounds" basis stated; residual
  row 86/80/91/91/96 vs parser's 13.6/19.5/9.0/8.6/4.0 (sums 99.5–100.1 — rounding, not
  raised); per-gap means 4.9/5.9/4.4/6.0/5.2 reproduce from §2.1's own table.
- **R3-2 repair verified:** the round-3 `### LEAD` entry exists (debate.md l.481), states the
  demonstration claim itself (ll.505–506), and the writer-capability audit's four
  observable→writer pairs are each physically writable as stated. The report's "demonstrates
  the form by construction" is a fair reading.
- **R3-6 repair's cadence composes:** conditions (i)/(iii) and the disposed-not-carried caveat
  now match the hardened arm; only the unadjudicated qualifier is residual (L5-F4).
- **R3-8/Q6 procedural half verified:** the lead's round-3 ruling did prefer deciding Q6 this
  run ("this run is the owner of record") and leaned realized-excluded; the report's citation
  is faithful. The mapping's *content* gaps are L5-F3, not a misquote.
- **§6.1 ranking item 3 behind item 2:** left waived per the round-3 merge's own
  checked-not-raised entry; L5-F5 concerns only the "comparable" gloss inside item 1.
- **§7 claim-count echo arithmetic:** ceil(166/40)=5 → min(4,5)=4 → 4+2=6 seats — holds.
- **§2.2's 194/266 = 72.9% ≈ "73%"** — holds as printed.

## Verification log (for the ledger)

| claim | reference | method | confidence |
|---|---|---|---|
| decompose-merge.mjs "committed" | run-dir git state | git status/ls-files/log first-hand | HIGH that claim is false at audit time (L5-F1) |
| archive fraction 72.6% | pinned findings.md @ working tree (pin-equal per [^PinCheck]) | awk recompute | HIGH (matches to 0.03pt) |
| GRADE enum members incl. `trivial`, `certain`, `realized` | debate.js l.60 @ 5396952-equal tree | direct read | HIGH |
| lens read surface = blue/report.md + CHANGELOG (no debate.md) | debate.js l.212 + this lens's own dispatch | direct read | HIGH |
| round-3 LEAD entry form-demonstration | debate.md ll.481–604 | direct read | HIGH |
| R4-1 "certain (already realized…)" | pinned run-3 findings.md l.~425 | prior-round ledger + report quote consistency | HIGH |

**Friction:** the 25k-token Read cap forced three windowed reads of the 1535-line
`blue/report.md` (56K tokens) — the run-3 friction-#15 class, still live at lens seats; a
single-file full-read affordance for the designated audit surface remains the shape the work
wants.
