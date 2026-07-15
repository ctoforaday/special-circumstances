# Lane 3 — local-repo critical-stance audit of the five docketed levers

Run 4 (efficiency investigation), 2026-07-14. Method lens: local-repo critical-stance — every
claim below verified first-hand against the pinned artifacts; nothing inherited from the
backlog's or retrospective's own framing without a diff. Assigned order: hypothesis 3 first,
then breadth (H1, H2, H4, H5).

**Pin verification (performed before any evidence was read):** `git diff --stat bfa8a3b HEAD --
research/2026-07-12_feov-retrospective/` is empty and `git diff --stat 5396952 HEAD --
ideas/backlog.md plugins/frank-exchange-of-views/` is empty — the working tree equals the pins
for every cited path, so working-tree reads are pin-faithful.[^PinCheck]

**Pre-flight note:** red's gap-pattern memory was not readable from this seat (the project
memory directory exists but is empty — see friction). Substituted the run-3 corpus's own
documented red patterns as the checklist: incomplete propagation, unquoted holds, verbatim
shipping of the counterparty's phrasing, stale status re-assertion without a diff, recurrence
miscounts. This draft's counts each state their derivation; every status claim carries its pin.

---

## H3 — Grade-dispute channel: RATIFY (minimal envelope form). Best-of-N grading: REJECT.

### What the pinned record actually shows

**Judge dispatch count: zero, verified structurally.** `grep -n "^### " debate.md` at `bfa8a3b`
returns 11 anchored headers — 6 `### BLUE` (rounds 0–5), 5 `### RED` (rounds 1–5), zero
`### LEAD`. (An unanchored grep returns 5 hits — all prose mentions of the string inside
findings; the anchored count is the real one, the same L-vs-occurrence trap run 3's own R5-3
documents.)[^DebateTranscript]

**Grade-correction traffic ran one direction only: red→blue.** Every grade correction in run 3
targeted cells in blue's §3 docket — R2-1 (likelihood count deflated 3→2), R2-9 (row-10 impact
re-graded against the shipped ledger's real skip-trigger), R5-2 (stale blocking-evidence count
deflated ~3x) — and each rode the ordinary gap loop to a landed fix.[^FindingsGrades] In the
reverse direction the count is zero: blue never disputed a red gap grade. Round-4 BLUE states it
outright ("none was over-graded relative to its fix cost"); round-5 BLUE conceded all six
grades.[^DebateTranscript] The one grade blue *argued* (row 15's High likelihood on the honest
2/2 ENAMETOOLONG count, R2-1→R3-7) was argued in prose, accepted by red in-round, and never
deadlocked. **Predicted round-savings from the channel on run-3-shaped traffic: zero.**

**But the asymmetry the lever targets is structural, and its downstream consumer is
demonstrated.** Red owns `red/findings.md`; a gap red closes carries red's grade into the
permanent record with no machine-readable path for blue to contest it — blue prose in
`debate.md` has no reader in the script, and the risk-accept mechanism covers gap *dispositions*,
not grade *values*.[^EngineSource][^RedMandate] Grade integrity has a proven consumer: this very
run's docket was assembled from run 3's graded record — over- or under-graded closures propagate
into the next run's priorities. So the channel is insurance priced at its mechanism cost, not a
round-saver.

**Mechanism cost, measured against the actual source:** `debate.js` has no filesystem access by
design (comment at lines 32–34; all state rides envelopes), so "a dispute red rejects
auto-dockets" cannot be implemented by the script reading findings.md.[^EngineSource] The
envelope-only form: `BLUE_ENVELOPE` gains optional
`grade_disputes: [{gap_id, dimension, proposed, evidence}]`; `RED_ENVELOPE` gains optional
`dispute_responses: [{gap_id, dimension, response: accepted|rejected, rationale}]`; the script
holds rejected disputes one round and adds re-disputed ids to `contested` — set arithmetic
mirroring the existing lineage filter at lines 244–245. Two optional fields, one filter clause,
one judge-prompt sentence, one simulator case. Complexity: low. Doctrine: clean — it routes to
the existing judge and cheapens nothing.

**One enforcement lesson must be inherited at ratification time:** run 3's R5-5 established (and
PR #15 shipped the fix for) the unenforced-optional-field failure class — a schema'd field set
under prompt instruction alone goes silently unset (R3-2's friction field: three rounds
unnoticed).[^FindingsGrades] `dispute_responses` has the same shape: a red-merge that ignores
blue's disputes leaves the field unset, telemetry-invisible. The same structural check applies:
if `blueEnv.grade_disputes` was non-empty last round and `redEnv.dispute_responses` does not
address every disputed gap_id, treat the unaddressed disputes as REJECTED (auto-docket), not as
absent. Default-to-docket makes silent non-compliance loud at the judge instead of invisible.

### Best-of-N grading: rejected on the backlog's own stated condition

The backlog gates best-of-N on "runs 4–5 show lone-voice bias survives."[^DocketBacklog] The
pinned corpus contains no surviving-bias instance — and the "lone-voice" premise is itself
partially false structurally: run-3 grading already passed through at least two voices per gap
(lens grade → merge temper), and the merge exercised that power against its own lenses twice in
round 5 alone — R5-5 tempered HIGH→MEDIUM-HIGH, R5-6 tempered MEDIUM-HIGH→MEDIUM, both
downward — plus two outright lens-error overrules (lens 5's unquoted "no discrepancy" hold,
overruled by direct read; lens 2's six-id claim, overruled by mechanical extraction at the merge
seat, `citation-ledger.md` line 184).[^FindingsGrades][^CitationLedgerRun3] Multi-voice grade
correction is observed working; a graded panel would add a judgment-seat cost multiple
($10.60–$13.56/round is what the one existing judgment merge already costs[^CostAudit]) against
a defect class with zero occurrences. Keep the backlog's evidence trigger; the `grade_disputes`
records this channel creates are exactly the per-gap records the trigger needs to be judged
against after runs 4–5.

**Confidence: high** — every load-bearing claim above is a first-hand read or grep at a pin.

---

## H1 — Severity-floor termination: REJECT as auto-termination. Ratify only the advisory signal.

### The floor as specified never fires on the only corpus we have

Backlog item 30(1): fire "when every open gap is <= MEDIUM with trivial fix cost," claimed
saving "would have ended run 3 at round 3 for ~$10."[^DocketBacklog] The per-round board maxima,
read verbatim from red's grade lines:[^FindingsGrades]

| Board after merge | Open gaps | Max severity | Floor fires? |
|---|---|---|---|
| Round 1 | 20 | HIGH (R1-1, R1-2) | no |
| Round 2 | 11 | MEDIUM-HIGH (R2-1, R2-3, R2-7, R2-8, R2-9) | no |
| Round 3 | 10 | MEDIUM-HIGH (R3-1, R3-2) | no |
| Round 4 | 5 | HIGH (R4-1) | no |
| Round 5 | 6 | MEDIUM-HIGH (R5-5) | no |

Zero fires; realized saving $0. **The backlog's "ended run 3 at round 3" claim is contradicted
by the pinned record it cites** — round 3's board carried two MEDIUM-HIGH code-trace gaps
(R3-1's degenerate-envelope loop, R3-2's dropped friction seat), neither ≤ MEDIUM nor
trivial-fix (both were docketed code changes, complexity low, not trivial).

### Making it fire makes it wrong

The only threshold that realizes the claimed saving is ≤ MEDIUM-HIGH — and at that setting the
floor fires after round 3 and terminates the run before rounds 4–5, which minted **R4-1** (HIGH,
certain × high — the lineage-blind docket, the retrospective's own "single most consequential
finding," shipped as PR #15's core) and **R5-5** (MEDIUM-HIGH — the unenforced-supersedes
critique, shipped as PR #15's structural throw).[^FindingsGrades][^CeilingDisposition][^ShippedList]
The floor's implicit model — the current board's severity predicts the next round's *discovery*
severity — is directly falsified by the round-3→4 transition: a MEDIUM-HIGH-max board preceded a
fresh HIGH mint. A judge disposing the round-3 residual board does not audit; disposition
produces no R4-1.

The frontier's re-scoped variant (arm only on a no-new-gaps round) also never fires: every run-3
round minted new gaps (20/11/10/5/6).[^FindingsGrades]

### The intent already has two cheaper, judgment-preserving paths — both demonstrated

(a) Operator stop-and-resume with reduced `maxRounds`: measured ~$0 via cache replay, cut ~7
residual rounds of run 3 (cost.md finding 5).[^CostAudit] (b) Ceiling assembly already produces
the residual-board disposition table the floor's judge call would produce (report §"Outstanding
gaps at the ceiling": per-gap grading, blue's response, disposition, compromise
rationale).[^CeilingDisposition] What run 3's operator lacked was the *signal*, not the
mechanism. **Ratify instead:** a `log()` line per round with the board profile (open count, max
severity, computed mass — see H2) so the stop decision is made by a judge — human or lead-judge —
with the numbers in front of them. Composes with the still-unshipped log()-per-transition
heartbeat item;[^DocketBacklog] costs zero tokens. Doctrinal ground: the stop decision is
*judgment*; automating it on red's own grades cheapens judgment, which is the one thing the
constraint forbids. The severity-floor is not "cheaper redundancy" — it is an autopilot for the
call the judge exists to make.

**Confidence: high.** Open exposure: run 3 is one run (n=1); a future corpus whose late rounds
genuinely produce only trivia would strengthen the floor's case — the advisory signal collects
exactly that evidence for free.

---

## H2 — Risk-mass-proportional spend: REJECT as a spend throttle. Mass is telemetry, not a control input.

### The computed mass series, and what it would have throttled

Mapping grades to numbers (low=1, low-medium=1.5, medium=2, medium-high=2.5, high=3,
certain/realized=3.5; likelihood × impact per open gap, summed at each merge), from red's grade
lines verbatim:[^FindingsGrades]

| After round | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| Risk mass | ≈98 | ≈62 | ≈44 | ≈29 | ≈32 |

Not monotone: round 5 *rose*. The minimum-mass board (post-round-4, ≈29) preceded the round that
minted two new gate-tier/dark-side findings (R5-5, R5-6). A mass-driven throttle scoping round-5
lens count down to the floor would have cut exactly the dark-side lens pass that produced both.

### The conceptual defect: residue is not discovery

Open-gap mass measures *known residue*; lens passes exist for *unknown discovery*. The two are
uncorrelated in the pinned record — the lowest-mass merges preceded the highest-value mints
(R4-1: found by 4-of-5 lens convergence on a low-and-falling board; R5-5/R5-6: dark-side lens on
the minimum board).[^FindingsGrades][^DebateTranscript] Scoping lens count to residue mass
throttles the adversary's discovery capacity using a gauge that reads only what is already
found. It also degrades red's internal error correction: both round-5 lens errors were caught by
*sibling-lens comparison* at the merge (three lenses converging against lens 5's hold; lens 1
and lens 3 against lens 2's miscount)[^CitationLedgerRun3] — thin-lens rounds remove the
cross-checks that made those catches. That is cheapening the adversary: doctrine violation,
independent of the spot-check floor.

### The frontier's grade-noise objection is real but not the decisive one

Within-round temperings (R5-5, R5-6) move computed mass ~1–1.5 points on ≈31 (≈5%) — modest. The
corpus's famous grade-correction cycles (R2-1, R3-7, R5-2) were corrections to *blue's §3 docket
cells*, not to the red gap grades that would feed the mass computation. Noise alone would not
reject the lever; the residue-vs-discovery conflation does.

### What already ships is the correct proportionality

`citationPasses = Math.min(4, Math.max(1, Math.ceil(claim_count / 40)))`, recomputed every round
(`debate.js:198`, PR #15): spend scales with the **verification surface** (blue's live claim
count), not with residue mass.[^EngineSource][^ShippedList] Surface-proportional spend is the
ratifiable doctrine — the work tracks what there is to check. Blue-respond spend already tracks
board size naturally without a throttle ($3.95 → $3.96 → $2.98 → $3.05 → $4.27 across
rounds).[^CostAudit] **Ratify only:** compute and `log()` the mass per round (H1's advisory
signal), which also accumulates the per-gap grade records the backlog's own best-of-N trigger
needs. No spend control wire.

**Confidence: high** on the run-3 numbers; **medium** on generalization (single-run corpus; the
mapping constants are mine — stated so the computation is reproducible, and no conclusion above
turns on ±0.5 in any constant).

---

## H4 — Sharded findings (open ledger / closed archive): RATIFY with four conditions. Collator seat: REJECT; ratify prompt-level read batching instead.

### The held conflict is narrower than docketed

The full-re-read MUST protects **red-reads-BLUE**: `red-auditor.md` line 13 — "re-read the FULL
living report in context" — names blue's living report; the lens prompt (`debate.js:212`) makes
it explicit ("the FULL living report ${runDir}/blue/report.md"). No mandate anywhere in the
plugin covers red re-reading its own closed cases — verified by direct read of the agent file
and both red prompts at `5396952`.[^RedMandate][^EngineSource] Sharding red's own findings file
does not touch the principle row 18 held to protect. (Correction to this run's own frontier: the
54KB/three-windowed-reads friction entry — run-3 friction #15 — is about `blue/report.md` at the
merge seat, not findings.md; sharding does not fix that entry, and the frontier's attribution
should not survive into the synthesis.[^FrictionHarvest])

### The measured driver supports sharding — and contradicts cost.md's own finding

Merge cost grew with the cumulative archive, not the dispute: round 5's red-merge was the run's
most expensive ($13.56, 7.87M cache reads, 61 turns) on the run's *smallest* dispute board
(6 open gaps), while findings.md reached 1364 lines.[^CostAudit][^FindingsGrades] **cost.md's
finding 2 — "merge cost tracks DISPUTE size (peaked r2, fell after)" — is contradicted by its
own table** ($13.56 r5 > $13.22 r2; 7.87M r5 cache reads > 5.64M r2): a live defect in a pinned
artifact, flagged here for correction. The growth term is the closed archive plus the growing
transcript dragged through every merge turn at judgment-tier cache rates ($1/MTok read) —
exactly what an open-items ledger removes from residency. Honest caveat: the merge context also
contains blue's growing report and the lens candidates; per-turn decomposition needs the run-3
transcripts, which are gitignored and absent at `bfa8a3b` (see friction) — the attribution is
directional, not exact. Confidence on magnitude: medium.

### The disconfirming tests, run against the record

**(a) R5-1 (the red-reads-own-closed-cases catch) survives sharding.** The *discovery* was
blue-side — row 23's cell vs §2.1(b), both inside blue/report.md, covered by the untouched
full-re-read of blue.[^FindingsGrades] The closed-case closure records (R2-5's, R1-13's) were
used for *verification*, on-demand and targeted ("every chain link checked against this file's
own closure entries" — R5-1's corroboration line). The design keeps the archive readable on
demand; the catch replays identically.

**(b) The dedupe function is the real archive dependency — a condition, not a rejection.** The
merge assigns "fresh R{round}-N ids to genuinely new gaps only" and runs merge-time dedupe notes
every round[^EngineSource][^FindingsGrades] — knowing a candidate gap is *new* requires
comparison against ALL prior gaps, including closed ones. Under naive sharding the merge either
re-reads the archive (savings vanish) or mints duplicate ids for re-litigated closed ground.
Condition: the open ledger carries a **compact closure index** (one line per closed gap:
id | closure class | one-line summary | supersedes), full prose in the archive — dedupe stays
resident at ~10% of archive bytes.

**(c) The write path is safe if the shards are skeleton-born and neutrally named.** Run 3's
controlled experiment isolated the write-block as filename-keyed and path-independent
(`findings.md` refused even in scratchpad; neutral name succeeded — friction #4), and Edit on
pre-created files worked every round (friction #10).[^FrictionHarvest] Condition: pre-create
both shards in the blackboard skeleton (PR #14 pattern) with non-report-semantic names (e.g.
`red/ledger.md`, `red/archive.md`). The frontier's undercounted-cost worry resolves: no new
detour if this holds; possibly one fewer (a small open ledger is cheaper to Edit than a 54KB
monolith).

**(d) The citation-ledger precedent holds, with one inherited qualifier.** The skip-rule held
all prior confidences through round 4 with zero closed-citation regressions (friction
#11).[^FrictionHarvest] R5-2 was not a ledger regression — the stale MA-status claim was never a
ledgered pair (first MA-status entries appear in round-5 blocks, verified against
`citation-ledger.md`).[^CitationLedgerRun3] Qualifier: archived closures whose evidence cites
volatile living sources must inherit the drift triggers PR #15 added to the ledger clause
(re-check on rounds-elapsed or volatility), and the spot-check floor extends to the archive — N
sampled closed cases re-verified per round, never zero (replacing round-4's full 41-closure
sweep, which caught nothing that round but is the class of check the floor must keep alive).

### Collator: the digest seat is the wrong half of the idea

The merge's value-add is adversarial comparison *across* lens passes: run 5's two lens errors
were caught because the merge held conflicting lens outputs side by side and re-derived
first-hand (`citation-ledger.md` line 184's merge-seat overrule).[^CitationLedgerRun3] A
collator that **normalizes** resolves exactly those conflicts before the judgment seat sees
them — it cheapens the adversary's input, failing the frontier's own test (b). A collator that
only **concatenates** is not a seat: it is `cat red/candidates/round-N-*.md` — one sentence in
the merge prompt (backlog 28(d)'s lever 3, prompt-level read batching), saving the 5-file read
fan-out at zero dispatch cost.[^MergeSeatAnalysis] The agent-seat version would cost a bulk
dispatch that itself reads all five files — roughly the spend it saves — to produce a lossier
input. REJECT the seat; RATIFY the prompt line.

**Confidence: high** on mandate scope, write path, precedent, and the R5-1 replay; **medium** on
the dollar magnitude of savings (transcript decomposition unavailable).

---

## H5 — Round-scoped audit: REJECT unconditional scoping. Conditional ratification for run 5+, gated on run 4's own evidence.

### The blindness is structural and the type specimens are R4-4 and R5-2 — not R4-3

An unpropagated correction site sits, by definition, in a section blue did not change; a
changed-sections audit excludes it. Clean specimens from the record: **R4-4** — the fifth
"4th occurrence" numeral sat in §3's risk-accepted paragraph, untouched since round 2's
correction missed it, caught in round 4 by a report-wide grep at the merge seat ("grep
'4th|fourth' at merge: exactly one uncorrected instance").[^FindingsGrades] **R5-2** — §4 row
1's status, unchanged since round 2, went stale by cross-corpus drift and was caught in round 5
only by a lens re-reading the other corpus first-hand.[^FindingsGrades] Both catches came from
audit surface a changed-sections rule excludes. (Correction to this run's frontier: its R4-3
type specimen is *weaker* than claimed — the unedited sentence sat in the same row 6 that R3-5's
fix edited, so row-granularity scoping would have included it. The frontier's argument survives
on R4-4/R5-2; it should not ride on R4-3.)

### Three shipped mitigations now cover the observed miss classes — none yet has live evidence

1. **Blue propagation clause** (PR #15): in `blue-researcher.md` line 14 and the blue-respond
   prompt (`debate.js:263`) — prevention at the source for the R4-4 class. Live rounds of
   evidence: zero; run 4 is the first trial.[^BlueMandate][^EngineSource][^ShippedList]
2. **Corpus pinning** (PR #16): kills the R5-2 cross-corpus-drift class structurally — pinned
   corpora cannot move mid-run. This run's own PINNED.md is the first live trial.[^ShippedList]
3. **Ledger drift triggers** (PR #15): time/volatility re-check conditions in the ledger clause
   (`debate.js:205`).[^EngineSource]

Narrowing red's full re-read before mitigation 1 has a single live run of evidence removes the
backstop and the belt in the same release — the exact compound-change class this project's own
review doctrine flags. The frontier's ratification condition is correct, and run 4 *is* the
evidence run.

### The scoping rule that would be safe (the run-5 candidate)

Round 1 always full. Rounds 2+: (a) changed sections in full context; (b) contested/lineage
locations; (c) **propagation expansion** — for every correction accepted this round, grep the
corrected strings/figures report-wide and add every hit site to the surface (mechanically cheap;
run 3's merge demonstrated the exact move when it found R4-4); (d) spot-check floor: N
random *unchanged* sections per round, never zero. Lens grades from scoped rounds carry a stated
confidence discount, mirroring the row-16b documented-tradeoff pattern. Ratification for run 5
is conditional: if run 4's record shows the shipped propagation clause holding (zero
propagation-class regressions), adopt for run 5; if run 4 shows it failing, REJECT outright —
the winnow list's own audit trigger — because scoping would then remove the only check that
catches the engine's measured dominant regression class (5 chains in 5 rounds, each costing a
full audit round at $25–30).[^CostAudit][^FindingsGrades]

**Cost at stake, honestly stated for the tradeoff:** red-lens is the largest recurring line
($9.22–$11.05/round, 5 agents), and its cost tracks corpus size, not board size — the one half
of cost.md finding 2 the table does support.[^CostAudit]

**Confidence: high** on the miss-class analysis; the run-5 condition is evidence-gated by
construction.

---

## Cross-cutting findings (lane 3 originals)

1. **cost.md finding 2 is internally contradicted** ("merge cost tracks DISPUTE size (peaked
   r2, fell after)" vs its own table: r5 $13.56 > r2 $13.22; r5 cache reads 7.87M > r2 5.64M).
   Needs a correction in the pinned artifact's successor; it also flips the finding's lesson —
   merge cost tracks the *cumulative archive*, which is H4's whole case.[^CostAudit]
2. **The backlog's severity-floor savings claim ("ended run 3 at round 3 for ~$10") fails
   audit** against the round-3 board it cites (two MEDIUM-HIGH open gaps).[^DocketBacklog][^FindingsGrades]
3. **This run's frontier carries two misattributions** (friction #15 is blue/report.md not
   findings.md; R4-3 is a weak type specimen for H5) — both corrected above so they do not
   propagate into synthesis.

## Verdict summary (lane 3)

| Lever | Verdict | One-line ground |
|---|---|---|
| (1) Severity-floor termination | REJECT auto-form; ratify advisory board-profile `log()` | Never fires as specified on run 3 ($0 saved); the threshold that fires truncates the rounds that minted R4-1/R5-5; stop decisions are judgment |
| (2) Risk-mass-proportional spend | REJECT throttle; ratify mass as logged telemetry | Residue ≠ discovery: lowest-mass boards preceded highest-value mints; thin lenses lose the sibling-lens error correction observed working; correct proportionality (claims-scaled citationPasses) already ships |
| (3a) Grade-dispute channel | RATIFY minimal envelope form (+default-to-docket on unaddressed disputes) | Structural asymmetry real, consumer demonstrated (this run), cost ~two optional fields; zero observed traffic — insurance, not savings |
| (3b) Best-of-N grading | REJECT | Backlog's own trigger unmet: no surviving bias in the corpus; grading is already multi-voice (merge tempered/overruled lenses 4x in round 5) |
| (4a) Sharded findings | RATIFY with 4 conditions: closure index in ledger; skeleton-born neutral filenames; archive drift triggers; archive spot-check floor | Full-re-read MUST covers red-reads-blue only; merge cost tracks archive size (r5 = priciest merge on smallest board); R5-1 replays under sharding |
| (4b) Collator stage | REJECT the seat; RATIFY prompt-level concatenation line | Normalization pre-resolves the lens conflicts the merge's round-5 catches depended on; concatenation needs no dispatch |
| (5) Round-scoped audit | REJECT unconditional; CONDITIONAL ratify for run 5 (propagation-aware rule + nonzero unchanged-section floor), gated on run 4 showing the shipped propagation clause holds | Changed-sections scoping is structurally blind to the R4-4/R5-2 classes; three shipped mitigations cover them but have zero live evidence |

## Open questions (carried)

1. Merge-seat turn decomposition: which fraction of red-merge turns/cache-reads is own-archive
   vs blue-report vs candidates? Needs run-4 transcripts retained for analysis or a per-agent
   timeline in cost-audit (backlog 28(d) names it). Gates H4's savings estimate from
   directional to measured.
2. Does the shipped blue propagation clause hold in run 4? (Gates H5's run-5 ratification.)
3. Does the lineage docket arm in run 4 (first live trial), and does any natural grade dispute
   occur? (Recalibrates H3's zero-traffic estimate.)
4. Is qmd MCP `get`/`multi_get` the intended archive on-demand path for H4, and is it approved
   in run environments? (Plain Read suffices; qmd makes the targeted fetch cheaper.)

## Friction (lane 3)

1. **Run-3 agent transcripts unavailable** (`trajectories/agent-transcripts.tar.gz` gitignored
   and absent at `bfa8a3b`): could not decompose merge-seat turns to size H4's savings; wanted a
   per-turn/per-agent timeline emitted by `scripts/cost-audit.mjs` (backlog 28(d) already names
   it). Without it, every context-residency claim in this debate stays directional.
2. **Red's gap-pattern memory unreadable from this seat**: the project memory directory
   (`~/.claude/projects/C--Users-gbloc-Projects-special-circumstances/memory/`) exists but is
   empty of files — the blue pre-flight clause ("check red's accumulated gap-pattern memory")
   is unfulfillable as written from a lane seat. Wanted: the red-auditor memory path stated in
   the run skeleton, or the pattern list mirrored into the run's inputs.
3. **No qmd MCP tools offered at this seat**; all recall was full Read/Grep. Acceptable for a
   repo-local lane, but the protocol's retrieval mode 1 was unavailable rather than declined.

## Footnotes

[^PinCheck]: "Pin equivalence check" — `git diff --stat bfa8a3b HEAD -- research/2026-07-12_feov-retrospective/` (empty) and `git diff --stat 5396952 HEAD -- ideas/backlog.md plugins/frank-exchange-of-views/` (empty), run first-hand at the lane seat. Local repo `C:/Users/gbloc/Projects/special-circumstances`. Accessed 2026-07-14.
[^DocketBacklog]: "Backlog — run-3 termination & fairness levers (item 30) and merge-seat analysis (item 28)" — `ideas/backlog.md` @ `5396952`. Accessed 2026-07-14.
[^CostAudit]: "Cost audit — FEOV retrospective (run 3)" — `research/2026-07-12_feov-retrospective/cost.md` @ `bfa8a3b`. Per-seat-round table + findings 1–5. Accessed 2026-07-14.
[^FrictionHarvest]: "Friction — FEOV retrospective (run 3), 17 entries" — `research/2026-07-12_feov-retrospective/friction.md` @ `bfa8a3b`. Entries #4 (filename-keyed write-block experiment), #10 (Edit-path success), #11 (ledger skip-rule held), #15 (25k Read cap vs 54KB blue/report.md). Accessed 2026-07-14.
[^FindingsGrades]: "red findings — FEOV retrospective, per-gap grade lines rounds 1–5" — `research/2026-07-12_feov-retrospective/red/findings.md` @ `bfa8a3b` (grade blocks at lines ~135–250 (r5), ~425–530 (r4), ~715–890 (r3), ~1080–1210 (r2), ~1279–1360 (r1); closure records rounds 2–5). Accessed 2026-07-14.
[^DebateTranscript]: "debate.md — FEOV retrospective full transcript (11 anchored headers, zero ### LEAD)" — `research/2026-07-12_feov-retrospective/debate.md` @ `bfa8a3b`; round-4 BLUE item on grades; round-5 RED merge temperings. Accessed 2026-07-14.
[^CeilingDisposition]: "Outstanding gaps at the ceiling — disposition and compromise rationale" + TL;DR — `research/2026-07-12_feov-retrospective/report.md` @ `bfa8a3b`, lines 3–20. Accessed 2026-07-14.
[^EngineSource]: "debate.js (FEOV 0.5.0, 288 lines)" — `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` @ `5396952`: no-filesystem doctrine (lines 32–34), envelope schemas (62–144), citationPasses recompute (198), ledger drift clause (205), lens full-re-read prompt (212), lineage-contested filter (244–245), structural closure throw (231–235), blue-respond propagation sentence (263). Accessed 2026-07-14.
[^RedMandate]: "red-auditor agent contract" — `plugins/frank-exchange-of-views/agents/red-auditor.md` @ `5396952`, line 13 (full-re-read MUST names the living report = blue's). Accessed 2026-07-14.
[^BlueMandate]: "blue-researcher agent contract" — `plugins/frank-exchange-of-views/agents/blue-researcher.md` @ `5396952`, line 14 (propagation clause). Accessed 2026-07-14.
[^CitationLedgerRun3]: "red citation ledger, run 3" — `research/2026-07-12_feov-retrospective/red/citation-ledger.md` @ `bfa8a3b`, lines 159–187 (round-5 MA-status entries; line 184 merge-seat overrule of lens 2). Accessed 2026-07-14.
[^MergeSeatAnalysis]: "Backlog item 28(d) — merge-seat cost analysis: turns × context; levers 1–4" — `ideas/backlog.md` @ `5396952`. Accessed 2026-07-14.
[^ShippedList]: "Already shipped — winnow list for run 4 (PRs #14–#18)" — `research/2026-07-14_efficiency-investigation/inputs/already-shipped.md`, staged run input. Accessed 2026-07-14.
