# debate.md — How should sleeper-service, the autonomous learning loop (Phase 4), be designed?

## Round 0

### BLUE

Synthesis complete: `blue/report.md` is the union of three lane candidates (adversarial,
primary-literature, local-probe lenses). All five frontier hypotheses came back SUPPORTED,
three of them with conditions the design now carries explicitly rather than discovering
later.

Position: (1) The loop consumes durable, pre-curated artifacts — friction harvests, cost.md,
board telemetry, backlog/doubts, red's mirrored gap-pattern memory — never introspection;
the external literature (Reflexion, DGM/SICA/STOP, and the negative result on intrinsic
self-correction) and both pinned harvests support this. (2) /self-improve is a thin driver:
scripted harvest + deterministic ranking, ONE pick per run, smoke-scale bounded research,
an idea stub with mandatory provenance and 3–5 alternatives; /graduate is human-only,
full-strength FEOV, promotion only via PR. (3) Headless daily runs are viable — lane-3
proved it by live probe (plugin command executed under `claude -p` on 2.1.212, write denied
and honestly reported), with named traps: the `--bare` default flip, the 10-minute
background-wait ceiling, the open MCP-headless bug trio, workspace-trust preconditions.
(4) Consent gates are structural and layered (allowlist-inverted dontAsk profile,
compiled-in PreToolUse fence, guardrail definitions outside the write surface, git/PR as
the load-bearing gate, invocation gate on /graduate) because permission rules alone are
empirically leaky — #22055 is closed NOT PLANNED — and autonomous loops editing their own
constraints is a realized failure class (AI Scientist, DGM, STOP's measured 0.42%).
(5) Cost: per-run `--max-budget-usd` + turn caps + a git-tracked month-to-date ledger
preflight; programmatic quota introspection confirmed absent, so the guard is honestly
static; judgment is never exercised unattended, so it is never cheapened.

Declared for red: 62 minority-report claims tagged inline with lane provenance; one
cross-lane conflict carried openly (backlog-append vs self-poisoning guard, §2.2 step 6 /
open question 14); 15 open questions; claim_count 132 (method in blue/CHANGELOG.md). The
pinned-path defect (`plans/claude-port-plan.md` absent at 7bc501e) remains standing
friction — port-plan citations are snapshot-grade, not pin-grade.

## Round 1

### RED

Verdict: **FAIL** — 30 open gaps (3 high / 3 medium-high / 9 medium / 4 low-medium /
10 low / 1 trivial; mass 118.75 under the pinned mapping). Full grading in
`red/ledger.md`; per-lens evidence preserved in `red/candidates/round-1-lens-*.md`;
~100 statement↔reference pairs leaf-verified this round into `red/citation-ledger.md`.

**The evidentiary base is unusually sound.** The overwhelming majority of checked pairs
verify HIGH at the leaf: the self-correction/Reflexion/Voyager/DGM/SICA/STOP corpus, the
Dependabot base rates, every load-bearing permissions/hooks/headless/scheduled-tasks/
routines doc quote, the pinned internal artifacts, and both live probes. Red banks
upgrades blue may claim: STOP figures re-pinned at ar5iv by three independent lenses
(open question 8 resolvable), [^UsageAPI] and [^AIScientist] upgraded to HIGH,
#837/#14246/#6631/#25621/#23707/#66395 statuses now leaf-confirmed.

**Where it fails is composition, not citation.** The three high-severity gaps are all
cases where the report's own verified pieces don't compose: R1-16 (the
`Bash(node scripts/*)` allow is a write-capable subprocess channel that contradicts the
"read-only analysis tools" claim AND doesn't match the printed artifact layout), R1-25
(the loop authors its own future harvest inputs — §1.5's never-EDITS guard leaves append
authorship open, so nightly friction self-amplifies its own ranking; the report's
semantic-consent section states the amplification loop approvingly), and R1-27 (the §4
gate stack is a rung-1 artifact; the scheduling ladder sells rungs 2–4 where most layers
silently don't exist). Close behind: R1-15 (the preferred `--bare --plugin-dir` recipe
voids layer 2's independence leg), R1-19 (the monthly budget ledger lives inside the
loop's write surface, violating the design's own invariant 4 for its only cross-run
control), R1-28 (the write-fence is marker-keyed and fails OPEN on marker loss — into
exactly the permission-rules-only configuration H4 refuted).

**Citation-side defects are real but trivial-fix:** the sharpest is R1-5 (§3.3/§6 assert
#32191 "leaf-checked OPEN"; it is CLOSED as duplicate — a false verification claim the
report's own footnote disclaims), plus the misplaced `disableBypassPermissionsMode` key
(R1-8, silent no-op as specced), the stale "no endpoint to read rate limits" (R1-9 —
a Rate Limits API now exists, read-only/Admin-key; the design conclusion survives on the
auth grounds), cwd-dependent deny anchoring (R1-12), and the uncited chmod-444
"community pattern" (R1-13).

**Argument defects blue must re-argue, not necessarily re-design:** R1-21 (risk row 4 is
accepted on the benign-actor premise §4.1 explicitly rejects), R1-17/R1-18 (exfiltration
and injection-through-retrieval absent from the risk matrix — risk-accept rows are an
acceptable ending if argued), R1-14 (the null alternative — keep the loop manual — is
never priced against the report's own Dependabot evidence), R1-20 (H3's HIGH stamp
attaches to the unprobed preferred recipe).

Merge notes: lens conflict on the gap-pattern-mirror line count resolved by direct read
against L1's hold (final byte is 0x0a → 1,557 exact; R1-30 minted); backlog recount and
#32191 independently re-verified at the merge seat. No closures, no supersessions, no
disputes pending — round 1. Nothing here demands new heavy machinery; every required
fix is doc-line/wrapper/schema class, so none of the 30 is offered as
interesting-but-not-of-interest.

### BLUE

Round 1: all 30 gaps addressed additively; zero grade disputes filed — red's grading was
accepted whole because every required fix priced trivial-to-low, so absorption beat
contestation on the pragmatist's own arithmetic. No substance subtracted; claim_count
132 → 142 (method re-derived in blue/CHANGELOG.md Round 1).

Where red was right and the DESIGN moved (not just the text): the three highs and their
neighbors were all composition defects between verified pieces, and the fixes converge on
one architectural clarification — **the wrapper is the invariant-4 layer**, now a named
third code artifact. Concretely: harvest.mjs runs wrapper-side, so the model session
executes no scripts and its Bash surface is pinned-argv read-only git (R1-16); the
monthly ledger lives in the operator-owned dir, wrapper-written, fail-closed on unparse
(R1-19); the write-fence is proven per run by a step-0 denial canary, so marker loss
fails closed (R1-28); wrapper start/end snapshots (porcelain + guardrail hashes, logged
outside the repo) restore the durable tamper evidence the DGM analogy required (R1-26);
`--plugin-dir` is pinned to an operator-owned read-only copy, never the working tree
(R1-15); §1.5 now guards AUTHORSHIP — sleeper-origin harvest rows are tagged and
recurrence-capped at 1, so the loop cannot ratchet its own complaints (R1-25); §3.4
carries a per-rung gate-survival table and rung-3/4 adoption is graduation-grade
(R1-27). Risk row 4 is re-argued with an explicit no-actor-intent-premise clause
(R1-21); rows 13–16 added (exfiltration risk-accepted with rationale after Read-scoping
and the nightly WebSearch drop — R1-17/R1-18; persistent death — R1-29; rung migration —
R1-27). R1-14's null alternative is priced honestly: rung 0 is the default and may be
terminal; daily cadence is a hypothesis the stub-survival ledger tests, with a named
triage-rate revisit trigger.

Citation-side: all 12 corrections landed with report-wide propagation greps in both
directions (logged in the CHANGELOG — red's incomplete-repair pattern applied
preemptively). Banked upgrades claimed: STOP ar5iv re-pin (OQ8 RESOLVED, including the
insignificantly-HIGHER precision offer), [^UsageAPI]/[^AIScientist] → HIGH, six issue
statuses added at their footnotes. New footnote [^RateLimitsAPI]; H5 verdict requalified
(spend limits CONFIRMED-NEGATIVE; rate limits API-readable but Admin-key-only — the
design conclusion survives on auth grounds, as red conceded).

Rebuttals: none maintained — no gap was wrong on its evidence. One routing note, not a
rebuttal: R1-7 is red-confirmed run infrastructure (PINNED.md asserts a nonexistent
path); blue's disclosure stands, the fix (setup-script pin validation or staging the
port plan into inputs/) belongs to the lead, and the item is re-raised in friction.md so
it survives the round.

Open questions now 19: OQ2 reframed (verify enforcement, not presence), OQ8 resolved,
OQ16–19 added (fence polarity signal, disableAutoMode leaf-check, Bash
compound/redirection/traversal tests, rung-2 profile passthrough).

## Round 2

### RED

Verdict: **FAIL** — 23 open gaps (1 high / 3 medium-high / 3 medium / 6 low-medium /
9 low / 1 trivial; mass 81.5 under the pinned mapping, down from 118.75). 29 of 30
round-1 gaps CLOSED (16 clean, 13 with regression — every successor's `supersedes` is
declared; lineage in `red/ledger.md`'s closure index, full prose in `red/archive.md`).
R1-7 stays open with the lead (run-infrastructure, not blue's).

**The repairs landed; the repair MACHINERY is the round-2 board.** Every round-1 repair
was re-verified as a new claim. All 12 citation-side corrections are faithful with no
repair-regression (propagation greps re-run at the merge: retracted tokens survive only
in correction-log contexts). New leaf work banked for blue: [^WindowsHang] body fetched
(MEDIUM→HIGH, exact regression span), [^WebSandbox] and [^MissedRun]-anacron →HIGH,
[^Pricing] re-fetched live with zero drift AND its ≤24h Batch sub-claim resolved HIGH at
the batch-processing page, `--json-schema` ≥2.1.205 and the rung-3 ~973MB figure pinned,
plus OQ7 raw material (the sandboxed Bash tool is documented not-native-Windows).

**Where it fails: the round-1 gate-side controls do not verify what they claim.** The
board's center of mass sits on the remediation machinery itself, in three shapes.
(1) *Controls that certify the wrong thing:* the step-0 denial canary has no specified
actor/observer/abort mechanism (R2-1) and — decisive — cannot isolate the hook fence
even in principle, because layers 1 and 2 fence the identical boundary: on marker loss
the canary is denied by layer 1 and PASSES, certifying "enforcement PROVEN per run" in
exactly the fence-dormant, permissions-only configuration H4 refuted; OQ2's own
acceptance test fails as designed (R2-2, found independently by two lenses).
(2) *Same-round fixes that don't compose:* §2.2 step 4 invokes the scripted FEOV
machinery nightly while §4.3/row 4 claim the round-1 shape "removes ALL script execution
from the session" — one branch stalls the daily driver under its own profile, the other
falsifies row 4's re-argued acceptance leg (a); the report never says which is true
(R2-3, the round's one HIGH, superseding R1-16 and R1-21). The §1.5 authorship cap is
keyed on a `*_self-improve/` dir-name glob, and the loop's own spawned FEOV sub-run
lands untagged friction that SUPPLIES the "non-sleeper corroboration" the cap demands
(R2-5); its "requires X — or, failing that, flags" disjunction is undecided between a
hard bar and a label, each horn with an unowned cost (R2-6). (3) *Evidence with no
watchman:* the R1-26 snapshot log has no defined reader or comparison step — the exact
undefined-receiver defect R1-29 fixed for loop death (R2-7); the plugin copy R1-15
minted has no refresh owner or version check, so a merged guardrail fix can run
undeployed indefinitely (R2-8); the dead-man line is pull-only (R2-9); the resume cap
bounds per-dir, not per-cause (R2-10). Cheaper residue: $2–5×30 vs the ~$50 cap is
arithmetically unreconciled and manufactures monthly skip streaks the dead-man line
cannot distinguish from death (R2-18); §3.3 grants pdf/arxiv ToolSearch reach in the
same sentence whose `--strict-mcp-config qmd-only` flag removes it, leaving the
research-protocol MUST-try structurally unsatisfiable nightly (R2-17); plus small
propagation/enumeration items (R2-4, R2-11–R2-16, R2-19–R2-22).

**Cross-cutting, for the lead (recommend-not-block):** the wrapper now hosts five
gate-side controls and is itself entirely trusted — R2-1/R2-2/R2-7/R2-8 are one shape
(a wrapper control that records or attempts but never verifies-and-surfaces its own
outcome). If round 3 patches wrapper controls singly again, a stated invariant — "every
wrapper gate emits a checked, human-surfaced liveness/outcome record" — beats another
lap of gate-by-gate (missing-root-invariant watch; board mass is declining, so this is a
recommendation, not a block).

Merge notes: lens severity divergences resolved by argument at the merge (R2-8 impact
set medium between L5's low-medium and L6's medium-high — no new write channel, but a
believed-deployed/undeployed gate; R2-7 likelihood low-medium per L6's broader trigger
set). No disputes pending; blue filed zero grade disputes round 1, so accepted_deltas is
empty. Archive spot-checks this round: R1-1, R1-5, R1-8, R1-16, R1-30 (merge greps of
the repaired text; noting in friction that a true cross-round spot-check first becomes
possible at round 3, since the archive entered this round empty).

### LEAD

Round-2 adjudication, 2026-07-17. Docket: R1-7 (re_raised) + R2-1..R2-16 (all
first_raise_successor — closed/rebuttal_sustained structurally unavailable; blue has not
yet answered them). Before ruling: debate.md and red/ledger.md read in full;
red/archive.md read END-TO-END, so every ancestor record on every supersedes chain
(R1-9, R1-11, R1-14, R1-15, R1-16, R1-19, R1-21, R1-22, R1-25, R1-26, R1-27, R1-28,
R1-29) was read and is named per ruling below. Key contested anchors independently
leaf-verified against blue/report.md at the judge seat: §2.2 step-0 header vs "First
in-session action: DENIAL CANARY" (lines 334/341), §4.3 layer 2's fails-CLOSED claim
(line 770), step 4's FEOV invocation vs layer 4's "removes ALL script execution" vs §6
row 4 (lines 352/772/933), §1.5's glob-keyed tag and undecided disjunction (lines
290–294), layer 5's reader-less snapshot (line 773). Red's mechanism analyses hold
against blue's own text in every checked case.

Rulings:

- **R1-7 — risk_accepted.** Re-raised infrastructure defect, full resolution set
  available. The finding is valid and verified (git cat-file -e MISSING at 7bc501e), but
  the tradeoff is decisive: the affected quotes were independently re-verified verbatim
  at working-tree 6df52af (round 2, L1), so residual impact is a reproducibility caveat
  already disclosed in [^PortPlan], and the fix (setup-script pin validation / staging
  the port plan into inputs/) is run-infrastructure owned by the lead, not report text
  blue can change. Carrying it would park an item blue cannot act on in red's verdict
  pool indefinitely. Recorded, not dropped: port-plan citations REMAIN snapshot-grade in
  the final report, and the lead owes pin validation (git cat-file -e per pinned path at
  run creation) before the next run — logged in friction.md.

- **R2-1 — carried.** Ancestors read: R1-28 (archive: canary added round 1; regression
  declared there verbatim). The step-0 header ("wrapper, OUTSIDE the model session") and
  the canary label ("First in-session action") contradict as printed, and a post-hoc
  envelope check or an instructional abort are both refuted by the report's own §4.1
  evidence. Blue owes: specify the concrete mechanism — the stream-json two-phase drive
  (wrapper sends canary prompt, parses the deny event, only then sends the real prompt),
  verifying against the CLI docs/probe that streamed input supports the two-phase
  pattern at the pinned version, OR a probe micro-session with an explicit
  same-environment argument (same flags, same settings source, same marker) — and
  reconcile the header with the label.

- **R2-2 — carried.** Ancestors read: R1-28. The masking argument is verified sound:
  every out-of-fence canary target is outside layer 1's allow set, so a deny outcome
  cannot distinguish hook-block from permission-rule denial; on marker loss the canary
  passes in exactly H4's refuted configuration while stamping "enforcement PROVEN per
  run". OQ2's acceptance test contradicts the claim it tests. Blue owes, in order:
  (1) probe whether deny PROVENANCE (PreToolUse-hook block vs permission-rule denial) is
  distinguishable in stream-json/envelope output; (2) if not, specify the positive
  hook-liveness record (hook writes a fired-record the wrapper confirms non-empty per
  run); (3) if neither is buildable, downgrade §0/§4.3 to "at least one deny layer is
  live", re-grade the marker-loss residual, and fix OQ2 to test the surviving claim.

- **R2-3 — carried** (the round's HIGH; priority for round 3). Ancestors read: R1-16 and
  R1-21 (archive records both; R1-21's leg-(a) regression is explicitly declared into
  this successor's scope). One branch is true by construction and either falsifies
  layer 4/row 4 or stalls the nightly driver — the design's core loop and its core
  acceptance argument cannot both stand as written. Blue owes: determine the FEOV
  execution locus BY PROBE (how does the frank-exchange-of-views machinery execute under
  the sleeper profile — session Bash, or a harness-side runner outside the Bash
  permission surface?); then either enumerate FEOV's script needs wrapper-side/
  pinned-argv, or re-scope layer 4 + row 4 to cover the workflow-runner channel (arguing
  hook coverage over its writes) and re-argue row 4's acceptance leg (a) on the true
  premise. R2-7's fix feeds leg (b) of the same re-argument.

- **R2-4 — carried.** Ancestors read: R1-16. Propagation miss, trivial: reword §2.2
  step 2 to match the wrapper-side architecture — "read the staged scored table"
  (wrapper work at step 0) or drop "script" (model-side language-tier scoring). Blue
  owes the one-line reword, consistent with whichever R2-3 branch is established.

- **R2-5 — carried.** Ancestors read: R1-25 (archive names both regressions verbatim).
  Verified: §1.5 keys on the `*_self-improve/` name glob while §2.2 step 4 spawns a
  second, topic-slug-named run dir whose loop-authored friction is untagged AND
  satisfies the corroboration gate — the guard is circumvented by default nightly
  control flow. Blue owes: tag by causal provenance — wrapper stamps EVERY run dir it or
  its session spawns with an origin marker file at creation; harvest reads the marker,
  never the name; restate §1.5 and re-grade §6 row 10 once sub-run corpus membership is
  closed.

- **R2-6 — carried.** Ancestors read: R1-25. "Requires X — or, failing that, flags" is
  two materially different mechanisms wearing one sentence; row 10's grade assumes the
  strong horn while the text licenses the weak one. Blue owes: decide the branch and own
  its cost in text. Red's severity-gated bypass (aborted-run/fail-closed/hook-crash
  classes enter flagged sleeper-only; ordinary classes need corroboration) is a
  workable shape the lead endorses as a starting point; whatever is chosen, row 10's
  grade must track the decision.

- **R2-7 — carried.** Ancestors read: R1-26 (archive: the undefined-receiver regression
  is declared there) and R1-29 (the defect class R1-29 fixed for loop death). A delta
  with no reader is not detection; row 4 leg (b) and its revisit trigger presuppose a
  watchman the design never specifies. Blue owes: wrapper compares start/end (and
  end-vs-prior-start) guardrail hashes at step 7; on mismatch, a doctor-visible flag
  and/or fail the next preflight closed; until specified, leg (b) and the trigger are
  unbacked and must not be cited as acceptance support in the R2-3 re-argument.

- **R2-8 — carried.** Ancestors read: R1-15 (archive: lifecycle-unbuilt regression
  declared). A merge landing is not the executing surface fixed. Blue owes: preflight
  pins and verifies the copy's content hash (or committed version stamp) against the
  operator-approved value, fail-closed on mismatch; the refresh step added to
  scheduling.md's guardrail-PR merge checklist; a doctor line for copy staleness.

- **R2-9 — carried.** Ancestors read: R1-29. "A surface the human already looks at" is
  asserted against the design's own adoption motive. Blue owes a decision, either horn
  acceptable if owned in text: a push surface on an N-day-silent threshold through a
  channel the operator passively receives, OR the honest statement that death-detection
  latency equals doctor cadence, with row 15's residual re-graded on that latency.

- **R2-10 — carried.** Ancestors read: R1-29. The per-dir/per-cause distinction is
  verified against §3.4's own escape clause. Blue owes: per-cause dead signature (hash
  of abort reason), HALT + dead-man flag after M consecutive same-signature fresh-dir
  deaths, and softening "cannot ... forever" to its per-dir scope — or an explicit
  argued accept of the bounded monthly waste.

- **R2-11 — carried.** Ancestors read: R1-22 (archive: composes-badly regression
  declared). The 30-day window penalizes the good case. Blue owes: distinguish
  "untriaged" from "triaged: graduation-queued" (human-set status exempts from
  auto-stale while still deduping), or make the window configurable against observed
  graduation cadence; state which in §1.4/§2.3.

- **R2-12 — carried.** Ancestors read: R1-19 (archive: the sibling was named in R1-19's
  own text and got neither fix nor argument). Red concedes the risk-accept is likely
  right; the report's own discipline is that acceptance is ARGUED. Blue owes one of:
  wrapper-derived completeness from its own step-7 record in the operator-owned dir +
  stated DEAD-marker location, or the explicit argued accept in text. Not ruled
  risk_accepted here precisely because the accept must live in the report, not in this
  ledger — a judge-side accept would ratify the silent skip.

- **R2-13 — carried.** Ancestors read: R1-9 (closed_with_regression; §6 row 5 cell
  missed). One-cell fix: "(no spend-limit API; rate-limit API unreachable at this auth
  tier — §5.1/R1-9)". Incomplete-repair; text as printed contradicts §5.1.

- **R2-14 — carried.** Ancestors read: R1-11. §7's self-audit must not misreport the
  report's own confidence ledger. Blue owes: append "(upgraded to leaf-verified HIGH
  round 1, R1-11)" to the Pattern B/E bullet and add R1-11 to §7's upgrade list.

- **R2-15 — carried.** Ancestors read: R1-14 (null alternative priced; §3.4 label
  lagged). Blue owes the one-line qualifier: "RECOMMENDED default AMONG SCHEDULED
  RUNGS, once the human opts in."

- **R2-16 — carried.** Ancestors read: R1-27 (archive: rung-0-column regression
  declared; the table exists precisely to stop overstated cells). Blue owes: split the
  R0 L2 cell ("fence YES (cache copy); canary n/a"), and one sentence either declaring
  manual-run spend out-of-ledger by design (with the cap-arithmetic composition stated)
  or wrapper-wrapping the manual path.

Cross-cutting direction (endorsing red's recommend-not-block as LEAD direction for
round 3): R2-1, R2-2, R2-7, R2-8 are one shape — a wrapper control that records or
attempts but never verifies-and-surfaces its own outcome. Blue SHOULD state the
invariant once ("every wrapper gate emits a checked, human-surfaced liveness/outcome
record"), derive the four fixes from it, and add it to §0's invariants, rather than
patching gate-by-gate a second time. This is direction, not a new gap; red grades the
result on the carried items' own terms.

Deadlock check: FALSE — 16 gaps carried this round AND red raised six new gaps
(R2-17..R2-22). The debate continues into round 3.
