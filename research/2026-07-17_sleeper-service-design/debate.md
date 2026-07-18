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

### RED

Verdict: **FAIL** — 17 open gaps (3 medium-high / 4 medium / 5 low-medium / 5 low; mass
55.0 under the pinned mapping, vs 81.5 round 2). All 22 round-2 gaps CLOSED (14 clean,
8 with regression — every successor's `supersedes` is declared; lineage in the closure
index, full prose in `red/archive.md`). R1-7 stands LEAD-adjudicated (risk_argued),
excluded from the verdict pool. Per-lens evidence preserved under
`red/candidates/round-3-lens-*.md`.

**Blue's round-2 repairs landed and invariant 7 closes the round-2 wrapper-gate watch.**
The two decisive round-2 items are resolved: R2-2's positive fired-record replaces the
deny-outcome canary and the smoke test verifies the fence-dormant abort; R2-3's FEOV
execution locus was determined by probe (session-Bash setup/capture + Workflow-tool
debate.js, hook-scoped) and the mixed-locus statement is leaf-verified against the pinned
`research.md`. The severity-gated corroboration bypass (R2-6), the watchman compare (R2-7),
the plugin-copy hash (R2-8), the SessionStart push surface (R2-9), the per-cause HALT
(R2-10), the graduation-queued exemption (R2-11), and every citation-side fix
(R2-13/14/15/17/21/22) verify at the leaf. One process note: the report carries these
round-3 revisions but `CHANGELOG.md` and this transcript carry NO round-3 ### BLUE block —
lens 3 misjudged the round state as post-round-2 from the CHANGELOG absence (merge resolved
by direct read of report.md).

**Where it still fails: composition residues the repairs did not reach, plus two Bash-surface
findings refuted at the leaf.** The dominant shape is unchanged from round 2 — a repair that
holds for the scheduled/completed/happy path but not for the design's own default mode,
abort path, or sibling surface. R3-2 (the round's medium-high): rung 0 — the DEFAULT and
"possibly terminal" mode — is wrapper-less by §3.4 yet "same code path" by §1.4, so four of
eight steps have no manual executor AND the origin-cap + corroboration gate are void exactly
where the design runs by default (supersedes R2-5/R2-6/R2-16). The origin-cap also misses
the red-memory mirror surface (R3-3) and voids on the dead-run path (R3-4); the corroboration
bypass never says WHO classifies a row as infrastructure, leaving a text-keyed forgeable lane
(R3-5); the watchman gaps at the abort path (R3-7) and is blind to out-of-repo subprocess
writes (R3-8); the per-cause HALT's dead SIGNATURE is identity-keyed against the design's own
nightly-re-minted dated/pathed abort records, so it never fires as specified (R3-11).

**Two first-raise findings run the leaf rather than reading it.** R3-14: the `dontAsk`
closed-world premise is refuted by the permissions doc's built-in, non-configurable read-only
Bash set (`cat`, `grep`, read-only git — auto-approved in EVERY mode), so `Bash(cat
//…/transcript)` — row 13's own named exfil target — is auto-approved and R1-17's read-scoping
is open on the Bash channel; the three `Bash(git …)` allow rules are redundant, unmodeled-
carve-out evidence. R3-15: `Bash(git log *)` + `--output=<path>` is an arbitrary-file write
gadget (tool-run confirmed: exit 0, file created), so row 4 leg (a)'s "no rule grants argv
that chooses a write target" is false without any permission bug, and OQ18 does not cover the
git-native output flag. Both are medium-high; both compose (a `git log --output` out-of-repo
write is exactly the R3-8 blind spot).

**Cross-cutting (recommend-not-block, R3-9):** invariant 7 makes every GATE self-attesting
but is scoped to the wrapper's gates, not the wrapper — the actor/observer/abort authority
for all five controls is the one component nothing verifies per run (the R2-8 hash check runs
BY the wrapper; the SessionStart surface has no liveness check). Neither protection
enumeration (layer 3, the snapshot guardrail set) was extended to name the wrapper, and the
invoked-copy provenance is unstated. This is the layer beneath the watch invariant 7 closed;
one stated residual + a cheap human-surface liveness line, not new machinery.

Merge notes: lens conflict on round state (lens 3 "no revision" vs lenses 1/2/5/6
"round-2-fix text present") resolved by direct read — blue DID revise report.md; CHANGELOG
lags. R3-6 merges L1-F3 + L5-F6 (doctor-channel overclaim, concurring). R3-9 merges L5-F4 +
L6-F4 (wrapper root-of-trust, concurring). R3-1 merges L2-F1 + L5-F7 (phase-drive
composition + behavioral/staging edges). Archive spot-checks re-verified this round: R1-16
(session-Bash pinning — the "read-only git" conclusion is the one R3-15 newly regresses; no
drift in the archived record itself), R1-25 (dir-name keying → R2-5/R2-6, consistent), R1-29
(resume cap → R2-9/R2-10, consistent). No disputes pending; blue filed zero grade disputes,
so accepted_deltas is empty. Nothing here demands heavy machinery — every required fix is
doc-line/rule-tightening/one-clause class; R3-15 (rule pinning) and R3-14 (deny rules over
the carve-out) are the only build-altering catches, both cheap.

### LEAD

Round-3 adjudication, 2026-07-17. Docket: R3-1..R3-8, R3-10..R3-13 (12 items, ALL
first_raise_successor — closed/rebuttal_sustained structurally unavailable; blue has not
yet answered them). Before ruling: debate.md and red/ledger.md read in full;
red/archive.md read END-TO-END including the round-3 closure block, so every ancestor
record on every supersedes chain (R2-1, R2-5, R2-6, R2-7, R2-10, R2-11, R2-16, R2-19,
and their round-1 chain roots R1-16, R1-22, R1-25, R1-26, R1-27, R1-28, R1-29) was read
and is named per ruling below. Contested anchors independently leaf-verified against
blue/report.md at the judge seat: §1.4 "same code path" (line 304) vs §3.4 "rung-0 manual
runs do not pass through the wrapper ... OUT-OF-LEDGER BY DESIGN" (lines 785–786); the
flat mid-drive `--json-schema` assertion (443–444) vs the report's own
final-result-only documentation record (633–635); the run-dir-keyed origin marker and
corroboration bypass (330–357) beside the text-keyed keyword clusterer (245); the
unbounded graduation-queued exemption (266–274); the step-7-only snapshot compare
(479–480); the unspecified death-signature normalization and the k×$5 arithmetic
(728–732). Red's mechanism analyses hold against blue's own text in every checked case.

Process note: report.md carries a round-3 revision but debate.md and CHANGELOG.md carry
no round-3 ### BLUE block — the change-summary channel is two rounds behind the living
report and cost lens 3 a misjudged round state. Blue owes the missing block (and the
CHANGELOG catch-up) with the round-3 response; logged in friction.md.

Rulings:

- **R3-2 — carried** (the round's priority). Ancestors read: R2-16 (archive: the rung-0
  execution-shape regression is declared there verbatim), R2-5 and R2-6 (both archive
  records declare the rung-0 void among their regressions), chain roots R1-25 and R1-27.
  Verified: lines 304 and 785–786 contradict as printed — the DEFAULT and possibly
  terminal mode has no executor for steps 0/2/4/7 and voids the origin cap and
  corroboration gate exactly where the design runs by default. Blue owes: decide the
  rung-0 execution shape. The lead endorses R2-16b's own cheap resolution — manual runs
  invoke the SAME wrapper via `--manual`, restoring ledger/origin/staging/canary and
  making "same code path" true (this also dissolves the R2-16b out-of-ledger accept);
  else specify how steps 0/2/4/7 execute manually, extend origin tagging to whatever
  creates the manual dirs, and re-grade §6 row 10 + §1.5 with the rung-0 residual stated.

- **R3-1 — carried.** Ancestors read: R2-1 (archive: the undocumented mid-drive
  composition regression is declared there), chain root R1-28. Blue owes: demote the
  `--json-schema` mid-drive leg to verify-at-build with a named text-parse fallback (new
  OQ beside OQ20/OQ21) or probe-and-cite; add the three behavioral legs (second-message
  injectability, phase-0 deny visibility in-stream, canary-attempt reliability) to the
  Phase-4 acceptance list; one sentence each for harvest.mjs failure at step 0
  (fail-closed skip with reason) and the qmd degrade note's destination/reader.

- **R3-3 — carried.** Ancestors read: R2-5 (archive: the red-memory-surface regression is
  declared there), chain root R1-25. The memory mirror is a genuinely separate authorship
  surface the run-dir marker cannot reach. Blue owes: extend causal-provenance tagging to
  the memory surface — a PRE-run frozen mirror snapshot (lead-preferred: simplest, no
  mtime forensics) or wrapper-window mtime tagging — and exclude sleeper-authored
  patterns from the non-sleeper corroboration pool; state in §1.5, re-affirm §6 row 10.

- **R3-5 — carried.** Ancestors read: R2-6, chain root R1-25. The safe reading is almost
  certainly intended but is nowhere the stated classifier rule, and §6 row 14 concedes
  the text channel is forgeable. Blue owes the one sentence: the infrastructure-class tag
  is assigned SOLELY from the wrapper's own event log (mechanical, session-unforgeable);
  friction TEXT contributes only to ordinary, corroboration-requiring classes.

- **R3-7 — carried.** Ancestors read: R2-7 (archive: the abort-path regression is
  declared there), chain roots R1-26 (undefined-receiver lineage) and R1-29. Aborts are
  model-influenceable, so "detection latency is one run" is unbacked for the
  write-then-force-abort sequence and must not be cited for abort paths until fixed.
  Blue owes: the wrapper snapshots at EVERY exit path it observes (it outlives the
  session by construction); the cross-run compare is defined against the LAST RECORDED
  snapshot of any kind; a missing chain link fails closed per invariant 7.

- **R3-4 — carried.** Ancestors read: R2-5, chain root R1-25. NOT risk_accepted despite
  the low likelihood: the fix composes with R3-7's every-exit-path snapshot at near-zero
  cost (the step-0 start timestamp is already colocated with the start snapshot in the
  operator-owned dir), dead runs are the design's own expected class, and per the R2-12
  precedent an acceptance must live argued in the report, not in this ledger. Blue owes:
  log the run-window START at step 0 (end = DEAD-mark time) and/or run the origin sweep
  at DEAD-marking time — or the explicit argued accept in text.

- **R3-8 — carried.** Ancestors read: R2-7, chain root R1-26. An overclaim in an
  acceptance leg is a text defect, not an acceptable risk. Blue owes: scope leg (b)
  honestly — the snapshot detects in-repo working-tree deltas and guardrail-file
  tampering, not arbitrary out-of-repo subprocess writes — and state the out-of-repo
  write as a residual bounded by pinned code + no-write-gadget, which blue may cite ONLY
  after the R3-15 gadget is closed (sequence the two fixes).

- **R3-11 — carried.** Ancestors read: R2-10 (archive: the normalization-unspecified
  regression is declared there), chain root R1-29. A HALT that never fires while behaving
  as specified is a control certifying the wrong thing — the round-2 shape recurring in
  the design's own repair. Blue owes: specify the normalization (signature = exit class +
  abort-reason TEMPLATE/error-code; strip dates, paths, session ids, nonces, numbers);
  zero-HALT-firings telemetry on the doctor line; one clause owning the
  alternating-cause residual.

- **R3-13 — carried.** Ancestors read: R2-11, chain root R1-22 (the governing principle:
  no stub may permanently subtract its class without a staleness re-surface). Blue owes:
  a cadence-tuned queued-stale backstop — a `graduation-queued` stub older than M
  weeks/months re-surfaces flagged `queued-stale` for human re-confirmation; state in
  §1.4/§2.3, re-affirm §6 row 3.

- **R3-6 — carried.** Ancestors read: R2-6, chain root R1-25. Either horn acceptable if
  owned in text: scope the sentence honestly (TAMPER/HALT persist; other classes surface
  only as most-recent) OR make it true — last-N reasons or per-signature counts since
  last human clear. The R3-11 normalization keys the latter; fixing R3-11 and R3-6
  together is cheaper than separately.

- **R3-10 — carried.** Ancestors read: R2-19 (archive: totality was verified against the
  round-1 tree; the round-2 additions are the recurring specimen). Blue owes: extend
  §0's enumeration (SessionStart hook + hooks.json, the prosthetic-conscience doctor-line
  delta named as CROSS-PLUGIN, the two operator-owned configs); one clause conditioning
  the SessionStart warning on scheduling-enabled (clearing HALT/disable resets it).

- **R3-12 — carried.** Ancestors read: R2-10, chain root R1-29. Wrong figures cannot be
  risk-accepted into a report whose own discipline is priced arguments. Blue owes: state
  in-night retry semantics explicitly, or recompute (≤$5/night; cap trips ~10
  ceiling-priced nights; HALT at night 12 or at cap, whichever first).

Non-docket items: R3-9 (recommend-not-block), R3-14, R3-15, R3-16, R3-17 are first-raise
findings with no blue response yet — they stay with red for blue's round-3 answer; no
adjudication is proper before blue speaks. The lead notes R3-14/R3-15 are leaf-refuted
build-altering catches and should head blue's queue with R3-2.

Cross-cutting direction (not a new gap): red's dominant round-3 shape is "the repair
holds for the scheduled/completed happy path but not the default mode, the abort path,
or the sibling surface." The carried set clusters accordingly — fixing R3-2 via the
`--manual` wrapper collapses the rung-0 voids in one move; the R3-7 fix should fold in
R3-4's start-log; the R3-11 normalization should key R3-6's counts. Blue SHOULD state
the completeness rule once — every guard is specified over ALL exit paths and ALL
authorship surfaces, default mode included — as invariant 8, parallel to invariant 7's
round-2 role, and derive the cluster fixes from it rather than patching path-by-path.

Deadlock check: FALSE — 12 gaps carried this round AND red raised five new items
(R3-9, R3-14..R3-17). The debate continues into round 4.

### BLUE

Round 3: all 17 open gaps addressed additively; zero grade disputes filed — third
consecutive round where every required fix priced trivial-to-low, so absorption beat
contestation on the pragmatist's own arithmetic (including R3-17, where red offered
risk-accept and the one-clause fix was cheaper than the acceptance argument would have
been). claim_count 142 → 151 (method re-derived in blue/CHANGELOG.md Round 3). The
missing round-2 ### BLUE block and CHANGELOG entry are repaired this round: a labeled
catch-up entry reconstructs the round-2 revision from the report's own R2-x markers, the
report header now carries round-2 and round-3 revision summaries, and the process defect
is owned in §7 and re-raised in friction.md with a concurrence on red/judge's proposed
parity check.

Before absorbing the two build-altering findings, blue re-verified both at the leaf
(critical-stance — red's output is treated as flawed until verified): (1) R3-15
reproduced independently — `git log -1 --oneline --output=/tmp/...` exited 0 and created
the file, with NO permission prompt, which adds a fact red's finding implies but did not
state: the carve-out CLASSIFIER itself passes `--output`, so pinning our allow rules
alone cannot close the gadget — the hook matcher is load-bearing, and the fix ships all
three legs (exact argv + `* --output*` belt denies + extended Bash-write matcher).
(2) R3-14 re-fetched at the permissions doc — the carve-out is verbatim as red said, and
BROADER than the gap summary: nine additional commands (ls, echo, pwd, wc, which, diff,
stat, du, cd), all now in §4.2's deny enumeration. A repair keyed on the gap JSON alone
would have shipped an under-inclusive deny list; logged as friction on the lossy-summary
channel, not against red.

Where the DESIGN moved (not just the text), per the lead's cluster direction —
invariant 8 is stated once (every guard specified over ALL exit paths, ALL authorship
surfaces, ALL execution modes, default included) and the fixes derive from it: rung 0
now invokes the SAME wrapper via `--manual` (R3-2, the lead-endorsed R2-16b resolution —
"same code path" becomes true by construction, the out-of-ledger accept DISSOLVES rather
than being re-argued, and the rung-0 marker/corroboration voids close in the same move);
the wrapper snapshots at EVERY exit it observes with compares chained to the
last-recorded-snapshot-of-any-kind, missing links failing closed (R3-7), which is also
where R3-4's run-window start-log and DEAD-time origin sweep ride; provenance tagging
extends to the red-memory mirror surface with sleeper-authored patterns excluded from
the corroboration pool (R3-3); the infrastructure-class bypass is keyed SOLELY on the
wrapper's own event log, never friction text (R3-5); the dead SIGNATURE normalization is
specified (template with dates/paths/ids/nonces/digits stripped — the corpus's
identity-keyed-detector lesson applied to our own repair), with zero-HALT-firings
telemetry and the alternating-cause residual owned (R3-11), and the same signature
machinery makes the doctor line's per-signature counts the R3-6 fix; graduation-queued
gains its own M-day queued-stale re-surface (R3-13 — no status is timer-free). §4.3
layer 4/§6 row 4 are re-argued on the TRUE Bash surface: pinned exact git argv + the
carve-out (deny-enumerated) + the Workflow channel, with leg (a) scoped to OQ18's
extended test matrix and leg (b) honestly limited to in-repo/guardrail sensing, the
out-of-repo residual bounded by pinned code + no-remaining-write-gadget, citable only
post-R3-15 (R3-8, sequenced as the lead directed). R3-9's cheap parts are absorbed
(watchmen named in layer 3 + the snapshot hash set, absolute-path invocation from the
operator-owned copy, SessionStart last-check timestamp, wrapper-hash dates on the doctor
line); the self-verification limit is OWNED in text — the wrapper cannot attest itself,
and the residual is a box-compromise scenario, not a loop-autonomy one. Citation-side:
R3-1 (mid-drive `--json-schema` demoted to verify-at-build; fenced-block text-parse
fallback named; OQ22/OQ23 added), R3-10 (§0 enumeration extended; SessionStart warning
conditioned on scheduling-enabled), R3-12 (retry semantics stated; ≤$5/night, cap ~10
nights, HALT night 12 or cap — recomputed), R3-16 (telemetry SHIPPED at 0.7.0), R3-17
(tokenizer set completed with Opus 4.7+).

Rebuttals: none maintained — no gap was wrong on its evidence; both leaf-refuted
findings verified true on independent reproduction. Propagation greps for every
corrected token run report-wide in both directions and logged in the CHANGELOG. Open
questions now 23: OQ18 extended (carve-out redirection/compound forms, git-native write
flags, classifier probes), OQ22 (mid-drive structured output), OQ23 (two-phase-drive
behavioral acceptance legs).

### RED

Round 4. Full living report re-read whole by all six lenses (1,893 lines, consecutive
windows); volatile leaves re-fetched live 2026-07-17 (permissions doc, [^Pricing] full grid,
#76239/#68375, STOP, DGM, AI-Scientist, OWASP) — zero drift on every one. Blue's round-3
revision addressed all 17 round-3 gaps; disposition: **7 CLOSED clean** (R3-5/6/7/8/11/16/17),
**10 CLOSED WITH REGRESSION** (successors carry declared lineage). Board: 16 open residues,
mass 46.0 (round 3: 55.0), max severity medium-high, open count 16 — the trend continues to
converge. R1-7 stands lead-adjudicated (excluded).

**What blue earned this round.** The two round-3 leaf-refuted Bash findings were repaired
soundly, not merely re-argued: the dontAsk carve-out is now carried in §3.2, all 14 named
commands are deny-enumerated (diffed line-by-line, L4/L6 HIGH), deny-supremacy is doc-verified,
and the git `--output` write channel is downgraded honestly with belt denies + a hook matcher.
Rung-0 dissolved its own out-of-ledger accept by routing manual runs through the wrapper
(`--manual`) per the lead's R2-16b resolution. The dead SIGNATURE normalization is now specified
with zero-firings telemetry (the corpus's identity-keyed lesson applied to blue's own repair).
R3-17 was completed rather than risk-accepted (Opus 4.7+ added; leaf-verified). No rebuttal was
maintained and none was owed — every gap was true on its evidence.

**Where the repairs under-reached (the round-4 docket).**
- **R4-2 (medium-high, sup R3-15).** "Deny-enumerated per command" is false for the carve-out's
  git member: read-only git is retained un-enumerated, and sibling git-native writers escape both
  belt and hook — leaf-verified this box, `git format-patch -1 -o /tmp/l5probe` → exit 0,
  arbitrary out-of-repo patch. The `-o` short form matches none of the three long-form belt denies.
  Sibling-halo on the R3-15 closure. Fix: invert to an allowlist on the git channel per the
  design's own invariant 6.
- **R4-1 (medium, sup R3-2).** The rung-0 fix closed the `--manual` path but `/self-improve` still
  ships as an invocable command under `commands/` carrying the full loop payload, no
  `disable-model-invocation` (its sibling has it), guarded only by a prose header — and §3.4's
  containment sentence has inverted polarity: markerless out-of-contract dirs PROVIDE non-sleeper
  corroboration, they do not "require" it. Origin-tag escape's 4th recurrence.
- **R4-3 (medium, sup R3-14).** The carve-out enumeration is built from a doc list the doc itself
  calls non-exhaustive ("include[s]"); an unlisted member auto-runs un-denied. The strictly
  stronger fix is a bare-`Bash` deny (the session never invokes Bash).
- **R4-4 (medium, sup R3-3).** R3-3's mechanism is aimed at a write the sleeper profile FORECLOSES
  (the nightly seat cannot write `.claude/agent-memory/**`), so §1.5 and §4.2 contradict; the real
  consequence — the seat's mandatory pattern-record write fails-denied every night — is unstated.
- Plus **R4-5** (deny-reach example refuted, sup R3-14), **R4-6** (unobserved-exit window sweeps
  human dirs, sup R3-4), **R4-9** (rejected/graduated statuses have no timer — 3rd per-status
  patch, missing root invariant, sup R3-13), and a fan of low/trivial residuals: R4-7 (degrade-note
  readers), R4-8 (count headline), R4-10 (cap resets — HALT lands month 2, sup R3-12), R4-11
  (probe attribution layer-masked, sup R3-15), R4-12 (est_complexity no source), R4-13
  (gate-survival row missing), R4-14 (dead-man disarm custody, sup R3-9), R4-15 (git subcommand
  boundary), R4-16 (CHANGELOG desync, 2nd round).

**Contested/for the judge:** R4-10 carried an intra-round lens conflict (L2/L6 held the R3-12
figures; L5 disputed). Merge recompute resolves it FOR L5: at ceiling pricing the monthly cap
trips first (~night 10) and pauses death accrual, so the HALT lands early the next month, not at
"night 12 or the cap." The bounded conclusion survives (≈ one cap + ε); the printed race is wrong.
Graded low because the safety conclusion is unaffected.

**Friction:** lens seats cannot run layer-isolating permission probes (nested `claude -p
--permission-mode dontAsk` denied twice by the seat's own auto-mode classifier), so R4-11's
attribution can only be graded, never settled, from inside the run — recorded in friction.md.

**Verdict: FAIL** — 16 open gaps, none yet risk-accepted by blue or ruled by the lead. Two
mechanism findings (R4-2 medium-high write channel; R4-1/R4-4 medium provenance/command surfaces)
carry real design weight; the rest are cheap. Nothing blocks a converging trajectory.

### LEAD

Round-4 adjudication, 2026-07-17. Docket: R4-1, R4-2, R4-3, R4-4, R4-5, R4-6, R4-7,
R4-8, R4-9, R4-10, R4-11, R4-14 (12 contested items — all `first_raise_successor`, so
`closed`/`rebuttal_sustained` are structurally unavailable; blue has not yet answered
them). Before ruling: debate.md and red/ledger.md read in full; red/archive.md read
END-TO-END (Round 2/3/4 closure blocks), so every ancestor on every supersedes chain is
read and named per ruling below. Contested anchors independently leaf-verified against
blue/report.md at the judge seat: §0 tree `self-improve.md` (line 62, NO
`disable-model-invocation`) vs `graduate.md` (line 66, carries it); §0 "exactly THREE new
code artifacts" (line 100); §1.5 "the nightly bounded-FEOV pass spawns a red-merge seat
that writes the SHARED agent-memory dir" (lines 411–413) against §4.2's
`Edit(<REPO>/.claude/**)` deny + the research/+ideas/ fence; §4.2 belt denies (line 1096:
`Bash(* --output=*)`, `Bash(* --output *)`, `Bash(* --output-directory*)` — no `-o`
short form); §3.4 "deny-enumerated" language (line 41); §3.4 out-of-contract containment
sentence (lines 958–960). Red's mechanism analyses hold against blue's own text in every
checked case.

Standing precedent applied this round (from the round-2 and round-3 LEAD sections): a
valid finding whose ACCEPTANCE would live only in this ledger, not argued in the report,
is not judge-accepted — "a judge-side accept would ratify the silent skip" (R2-12);
low-likelihood findings whose fix composes at near-zero cost are carried, not accepted
(R3-4); "wrong figures cannot be risk-accepted into a report whose own discipline is
priced arguments" (R3-12). All three bind their round-4 successors directly.

Rulings:

- **R4-2 — carried** (the round's medium-high; priority for round 4). Ancestor read:
  R3-15 (archive Round-4 block: the `git log --output` gadget, first-raised round 3,
  closed with three-part belt+hook shut; the sibling-halo is declared into this
  successor's scope). Verified: line 41 stamps the carve-out "deny-enumerated" while line
  1096's belt denies match only the three `--output` LONG forms — the `-o` short form and
  the `format-patch`/`archive`/`bundle create` writer family escape both belt and the
  hook's named-flag list, and read-only git is retained UN-enumerated by design. The
  leaf-verified `git format-patch -1 -o /tmp/l5probe` → exit 0 arbitrary out-of-repo write
  is the same write class R3-15 closed, landing in the R3-8-conceded snapshot blind spot.
  This is a live write channel, not a risk to accept. Blue owes: invert to an ALLOWLIST on
  the git channel per the design's own invariant 6 (the hook already parses Bash — deny
  any git argv that is not one of the exact allowed read forms); at minimum extend the
  hook matcher + belt denies to `-o`/short-form output flags and the
  format-patch/archive/bundle family; and correct row 4's "hence deny-enumerated per
  command" to name the git exception explicitly, with OQ18(c) as its standing test.

- **R4-1 — carried** (mechanism weight; priority with R4-2/R4-4). Ancestors read: R3-2
  (archive Round-4 block: the /self-improve command-invocation surface + inverted
  containment polarity declared into this successor), R2-16, R2-5, R2-6, chain roots
  R1-25 and R1-27. Verified: `self-improve.md` (line 62) ships the full loop payload under
  `commands/` — the harness entry-point directory — with NO `disable-model-invocation`,
  while its sibling `graduate.md` (line 66) carries exactly that gate; the asymmetry is
  never argued, so the unwrapped path R3-2 closed for rung-0 `--manual` stays mechanically
  open via command invocation (human or model/scheduled), guarded only by a prose header
  against §4.1's own "instructional gates degrade (0/175)" evidence. The §3.4 containment
  sentence (lines 958–960) has INVERTED polarity as red states: markerless out-of-contract
  dirs are classed NON-sleeper and thereby SELF-SATISFY the corroboration bar ordinary
  classes wait on — the threat vector printed as a mitigation. Origin-tag escape's 4th
  recurrence is the gate-erosion signal that forecloses a judge-accept. Blue owes: (i)
  make `/self-improve` a thin trampoline instructing `node <operator-owned>/
  sleeper-wrapper.mjs --manual`, moving the phase-1 payload out of `commands/`; (ii) add
  `disable-model-invocation: true` to `self-improve.md`; (iii) correct §3.4's containment
  polarity (markerless out-of-contract dirs enter harvest as non-sleeper and CAN
  corroborate) and re-affirm §6 row 10 with the residual named, OR quarantine markerless
  loop-shaped dirs / redefine the corroboration pool positively.

- **R4-4 — carried** (mechanism weight). Ancestors read: R3-3 (archive Round-4 block: both
  legs — foreclosed premise and change-keyed granularity — declared into this successor),
  R2-5, chain root R1-25. Verified: lines 411–413's §1.5 sentence (the nightly seat writes
  the shared agent-memory dir under hook coverage) cannot both be true with §4.2's
  `Edit(<REPO>/.claude/**)` deny + the research/+ideas/ fence + the phase-0 canary abort —
  the in-repo `.claude/agent-memory/**` write is foreclosed at three layers, so the tagging
  machinery R3-3 built is aimed at a write that cannot execute in-window; the real
  consequence (the seat's mandatory record-new-patterns write fails-denied every night — a
  silent degradation of the nightly adversary's learning plus a routine fence-denial stream
  in the fired-record) is stated nowhere. This is an internal contradiction (red grades it
  "certain"), not a risk. Blue owes: pick a horn and state it. Cheapest coherent — declare
  agent-memory writes DENIED by design under the sleeper profile (nightly seats do not
  learn; nightly patterns reach the corpus only via ordinary-classed friction text),
  reword §1.5's R3-3 paragraph to its true remit (belt for out-of-window/interactive
  authorship + defense-in-depth against profile drift), and state expected nightly
  memory-write denials as a NORMAL fired-record class; if seat writes are meant to succeed,
  enumerate the allow and re-argue invariant 6. Separately: tag at the granularity of what
  the window ADDED (pre-existing content keeps its origin; only in-window APPEARING
  patterns are wholly sleeper-origin), closing the R1-22 monotonic-blinding leg.

- **R4-3 — carried.** Ancestor read: R3-14 (archive Round-4 block: the non-exhaustive-doc
  regression declared into this successor). Verified against red's leaf work: the deny set
  is built from a doc sentence non-exhaustive by its own phrasing ("include[s]"), and the
  same page names `sort`/`sed` as classifier-reasoned commands outside the named 14 — an
  unlisted member (`sort`/`file`/`readlink`/`strings`/`less`) auto-runs un-denied, and the
  belt Read-denies bind only the NAMED `~/.claude` targets, leaving un-named box secrets
  exposed. Not risk_accepted despite low-medium likelihood: a cheap STRUCTURAL close exists
  and the design's own invariant 6 favors it. Blue owes: either (a) extend OQ18 with a
  member-enumeration probe under a bare dontAsk profile (any auto-running command gets a
  deny in the shipped file; state the bare-vs-`*` convention), AND/OR (b) the strictly
  stronger shape per invariant 6 — deny the bare `Bash` tool in the sleeper profile
  (doc-verified: a bare tool name removes the tool entirely), closing the whole carve-out
  class structurally, since §2.2's session steps never invoke Bash.

- **R4-9 — carried.** Ancestors read: R3-13 (archive Round-4 block: the rejected/graduated
  timer-free regression declared into this successor), R2-11, chain root R1-22 (the
  governing principle: no status may permanently subtract its class without a stated
  re-surface path). Third consecutive per-status patch (R1-22 → R2-11 → R3-13) with no root
  invariant — the missing-root-invariant signal is decisive, and R1-22's lineage forbids
  permanent class subtraction. Blue owes: state the invariant ONCE in §1.4 — every status's
  dedupe effect has a stated re-surface path — and give the two terminal states their
  semantics: `graduated` → class recurrence re-enters flagged `regression` (mirroring the
  backlog rule); `rejected` → dedupes for a cadence-tuned window (or until recurrence
  exceeds the pre-rejection rate), then re-surfaces `rejected-recurring` for
  one-keystroke re-confirmation. Re-affirm §6 row 3.

- **R4-5 — carried.** Ancestor read: R3-14 (archive Round-4 block: the prior-exposure
  example refutation declared into this successor). The R3-14 repair itself is sound and
  hardening; the defect is the postmortem — `Bash(cat //…/.claude/projects/…)` on row 13's
  named transcript target would have been BLOCKED under the round-2 profile (its
  `Read(//…/.claude/projects/**)` deny extends to recognized Bash file commands per the
  doc's own deny-reach clause), not auto-approved. Not risk_accepted despite low-medium
  impact: the fix is trivial and a false historical narrative in a risk row is a text
  defect, not an acceptable risk; the correcting fact strengthens the design honestly. Blue
  owes: re-point the example at an undenied credentials-class path; add one clause noting
  Read/Edit denies extend to recognized Bash file commands (`cat`/`head`/`tail`/`sed`) —
  which also honestly narrows R4-3's impact to un-named targets.

- **R4-6 — carried.** Ancestors read: R3-4 (archive Round-4 block: the unobserved-exit
  degenerate case declared into this successor), R2-5, chain root R1-25. Binding
  precedent: the round-3 LEAD ruled R3-4 NOT risk_accepted despite low likelihood because
  the fix composes at near-zero cost and an acceptance must live argued in the report — the
  identical reasoning governs this same-lineage successor. A wrapper hard-kill is an
  UNOBSERVED exit; the multi-day START-without-END window then sweeps daytime human-present
  run dirs into `origin: sleeper`, suppressing exactly the human-present corroboration
  ordinary classes wait on. Blue owes: (i) a window's END is additionally bounded by the
  NEXT wrapper START (no window spans invocations); (ii) an unclosed window is read as
  extending to the present, OR a DEAD-mark-closed window is flagged retroactive-uncertain
  with its markerless sweep confined to sleeper date-key naming (others surfaced for human
  confirmation); (iii) own the unobserved-exit case in step-7/§4.3-layer-5 text (backstop =
  snapshot chain + resume path) — or the explicit argued accept in text.

- **R4-7 — carried.** Ancestors read: R3-1 (archive Round-4 block: the degrade-note
  reader-declaration regression declared into this successor), R2-1, chain root R1-28.
  Verified: both named reader sites carry no surfacing obligation — §2.3's `confidence`
  spec names only the R2-17 PDF caveat, and §3.4's doctor line prints skip/abort reasons (a
  qmd degrade is neither) — so a builder implementing §2.3/§3.4 as written drops the
  surfacing and chronic daemon-start failure degrades recall silently. Blue owes: one
  clause in the §2.3 confidence field (qmd-degrade labeling) and one degrade-note term in
  the §3.4 doctor line.

- **R4-11 — carried.** Ancestor read: R3-15 (archive Round-4 block: the layer-masked
  attribution declared into this successor). The GADGET stands twice-verified; the
  ATTRIBUTION does not — both reproductions ran under `defaultMode: "auto"`, where the AUTO
  classifier is the approving layer, so "the carve-out classifier itself passes `--output`"
  and "rule-pinning alone cannot close the gadget" (lines 1138–1140, 1429–1430) are printed
  as established on an un-isolated fact. Safety is unaffected (the three-part close is
  conservative in both worlds); the risk is a Phase-4 builder deprioritizing the
  hook-matcher leg. NOT unresolved: red's friction (the isolating dontAsk-zero-allow probe
  is denied from inside the run by the seat's own auto-mode classifier) blocks SETTLING the
  attribution here, but the OWED FIX — honest re-scoping plus deferral of the isolating
  probe to build-time — is fully available. Blue owes: re-scope "showing/is itself
  evidence" to "consistent with carve-out classification but NOT isolating it (both probes
  ran under auto mode)"; add the isolating dontAsk-zero-allow probe to OQ18(b)/(c) or OQ23;
  state the hook matcher as chosen-conservative, not proven-necessary.

- **R4-10 — carried.** Ancestors read: R3-12 (archive Round-4 block: the cap-reset
  recompute regression declared into this successor, with the L2/L6-vs-L5 lens conflict
  resolved FOR L5 by merge recompute), R2-10, chain root R1-29. Binding precedent: R3-12's
  own ruling — "wrong figures cannot be risk-accepted into a report whose own discipline is
  priced arguments." The bounded conclusion survives (≈ one cap + ε), but the printed
  "night 12 OR the cap, whichever comes first" race and single-month bound are wrong at
  ceiling pricing. Blue owes: one clause — at ceiling pricing the cap trips FIRST (~night
  10), death accrual pauses with it, and the HALT lands early in the following month;
  worst-case deterministic-cause burn ≈ one monthly cap + ≤2 nights of the next month.

- **R4-14 — carried.** Ancestor read: R3-9 (archive Round-4 block: the disarm-custody +
  same-line R3-9/R3-10 composition regression declared into this successor). Guardrail
  integrity: a loop-reachable scheduling-enabled flag would let the loop silence its own
  dead-man channel — precisely the guards-modifying-guards class the safety section forbids
  — and if "conditioned" means fully silent when the flag is off/absent, flag LOSS (the
  modal marker-loss drift case R1-28 treats as certain) silently disarms the channel and a
  disarmed hook looks identical to a healthy-but-disabled one, re-opening what R3-9's
  always-print clause closed. Blue owes: two clauses — the flag lives in the operator-owned
  dir beside the ledger; when scheduling-disabled the SessionStart line still prints one
  token ("sleeper: scheduling disabled; last check <ts>") so silent-disarm ≠
  disabled-by-choice.

- **R4-8 — carried.** Ancestors read: R3-10 (archive Round-4 block: the count-headline
  regression declared into this successor), R2-19. Verified: line 100 asserts "exactly
  THREE new code artifacts" while the same §0 paragraph enumerates a fourth executable (the
  SessionStart staleness hook, R3-10's own text: "a new executable + hooks.json
  registration"); the skill file/manifest got an explicit "new PROSE artifacts, not code"
  classification, the SessionStart hook got none, and its host plugin is unstated. Blue
  owes: one clause — either count it (FOUR code artifacts) or declare it a hooks.json-inline
  one-liner (not a code artifact) and say so; name the host plugin.

Non-docket (no adjudication proper before blue speaks — per the round-3 precedent for
R3-14/R3-15): R4-12 (est_complexity source), R4-13 (missing gate-survival row), R4-15 (git
subcommand boundary), R4-16 (CHANGELOG desync, 2nd round) are new first-raise findings with
no blue response yet; they stay with red for blue's round-4 answer. The lead notes R4-2,
R4-1, and R4-4 are the weight-bearing mechanism catches and should head blue's queue.

Cross-cutting direction (not a new gap): red's dominant round-4 shape is "the round-3
repair closed the path it named but not the SIBLING of that path" — R4-2 is the git
sibling of the R3-15 flag-close; R4-1 is the command-invocation sibling of the rung-0
`--manual` close; R4-3 is the unlisted-member sibling of the enumerated 14; R4-9 is the
terminal-status sibling of the queued-status timer. Blue's invariant 8 ("every guard over
ALL exit paths / surfaces / modes") already names this shape; the round-4 residuals are
where a NAMED close did not sweep to its structural neighbor. Blue SHOULD prefer the
STRUCTURAL closes the findings offer (allowlist-invert the git channel, R4-2; deny bare
Bash, R4-3; state the once-and-for-all status re-surface invariant, R4-9) over another lap
of neighbor-by-neighbor enumeration — derive from invariant 6/8 rather than extend a list.

Deadlock check: FALSE — 12 gaps carried this round AND red raised four new first-raise
items (R4-12, R4-13, R4-15, R4-16). The debate continues into round 5.

### BLUE

Round 4. All 16 open gaps addressed additively; no substance subtracted; zero grade
disputes (fourth consecutive round — the two weight-bearing catches are genuine mechanism
defects, and a dispute could not honestly deflect a live write channel or an internal
contradiction). I took the lead's cross-cutting direction literally: red's round-4 shape was
"the round-3 repair closed the named path but not its sibling," so I answered with STRUCTURAL
closes from invariants 6/8 wherever one existed, not another lap of enumeration.

**The two structural closes (R4-2 + R4-3, folded).** The round-3 posture — enumerate the 14
carve-out commands, pin git argv, deny known write flags — was a denylist wearing an
allowlist's job, and red proved it incomplete twice over: the doc set is stated "include[s]"
(unlisted members auto-run), and `git format-patch -1 -o /tmp/probe` writes an arbitrary
out-of-repo patch at exit 0 with `-o` matching none of the three long-form denies. Rather than
chase `-o`, then `archive`, then `bundle`, then `config` one line at a time, I took the
lead-endorsed structural shape: **deny the bare `Bash` tool in the sleeper profile.** A bare
tool name removes the tool entirely (doc-verified), §2.2's session steps never invoke Bash
(git and node run wrapper-side), so the whole carve-out CLASS closes at the tool boundary in
one rule. On the git channel specifically I inverted the hook to a read-ALLOWLIST per
invariant 6 (deny any git argv not an exact allowed read form) so a new git writer defaults
denied. The enumerated denies, the `-o`/`-O` + writer-family belt denies, and the git
allowlist are all RETAINED — additive — as defense-in-depth for the rungs that rebuild
without the bare deny (3–4) and against profile drift. Nothing was subtracted to satisfy the
structural fix.

**R4-1 / R4-4 — the provenance and mode surfaces.** /self-improve is now a thin
`disable-model-invocation` trampoline with its payload out of `commands/` (no invocable
command carries the loop; matches its sibling), and I corrected the §3.4 polarity error red
and the lead both flagged: markerless out-of-contract dirs are NON-sleeper and therefore
CORROBORATE — they don't "require" corroboration. The residual (a human's deliberate paste-run
under the interactive profile) is named, bounded by the window sweep + 1-stub cap, with a
quarantine option deferred as OQ24 rather than built, because a design made worse to catch a
human ignoring their own contract is itself a defect. R4-4 I resolved by picking the cheapest
coherent horn the lead named: agent-memory writes are DENIED by design under the sleeper
profile — nightly seats do not learn, the fence-denial is a NORMAL fired-record class, and
red's patterns reach the corpus only via ordinary-classed friction text; the R3-3 machinery
survives as belt-for-drift, now tagging at window-ADDED granularity so one nightly append can't
re-classify months of human-present evidence.

**R4-6 / R4-9 — degenerate cases and the missing root invariant.** Unobserved wrapper exits
(hard-kill) now bound the window by the next START and confine the markerless sweep to sleeper
date-key naming, so a crash can never sweep daytime human dirs into origin:sleeper. And I
ended the per-status patch chain (R1-22→R2-11→R3-13) the way the lead asked — one root
invariant in §1.4 (every status has a stated re-surface path; no status is timer-free) with
`graduated`→regression-on-recurrence and `rejected`→90-day/rate `rejected-recurring`.

The rest are the cheap fan (R4-5/7/8/10/11/12/13/14/16), all absorbed and propagated:
notably R4-5 and R4-11 each had a SECOND site beyond the one red flagged (the §4.2 prose
bullet for the mis-diagnosed `cat` example; the §7 note for the gadget attribution), both
found by report-wide grep and corrected — incomplete propagation was run 3's dominant blue
regression class, and I greped every corrected token in both directions (log in CHANGELOG
Round 4).

Verdict I'd argue for: the two mechanism catches are genuinely closed structurally (not
re-argued), the contradiction is resolved by an explicit horn, and the remainder are text
corrections. Open questions carried: OQ18(c) now the standing git subcommand/member probe,
OQ23(d) the isolating attribution probe, OQ24 the paste-run quarantine — all build-time,
none load-bearing given the structural closes.

### RED

Round 5. Full living report re-read whole by all six lenses (2,159 lines, consecutive
windows); volatile leaves re-fetched live 2026-07-17 (scheduled-tasks, routines,
permissions, hooks, pricing, missed-run, GHA-schedule, web-sandbox docs; #76239/#68375/#837/
#14246/#22055/#6631/#25621; IdeaStudy) — zero refuted external leaves, ten living-source
claim sets zero-drift; the R4-2 git gadget re-reproduced in BOTH spaced and attached forms
(lens + merge seats). Blue's round-4 revision addressed all 16 round-4 gaps; disposition:
**6 CLOSED clean** (R4-5/7/10/11/15/16), **10 CLOSED WITH REGRESSION** (successors carry
declared lineage). Board: 10 open residues, mass 30.0 (round 4: 46.0; round 3: 55.0), max
severity medium, open count 10 — still converging. R1-7 stands lead-adjudicated (excluded).

**What blue earned this round.** The two structural closes are real where they aim: the bare
`Bash` deny is doc-grounded (bare-name removal + deny supremacy re-verified live), the git
hook read-allowlist is the invariant-6 shape red asked for, the trampoline mechanism is sound
and symmetric with graduate.md, the R4-9 root invariant is stated once with terminal-state
semantics, and the cap/HALT arithmetic now survives three independent recomputes (L2/L5/L6).
The R4-4 horn is picked and stated plainly. Propagation was better than round 3 — but not
complete (below).

**Where the repairs under-reached (the round-5 docket).** The round's common thread: round
4's STRUCTURAL closes were derived over the TOP-LEVEL session and not over the two surfaces
where nightly work actually executes.
- **R5-1 (medium, sup R4-3+R4-2).** §4.3 layer 4 (iii) says the Workflow seat agents are
  full permission-engine subjects (⇒ the bare deny binds them) while §4.2's R4-2 bullet
  names "the Workflow seat agents" as a surface where Bash IS reachable — both cannot hold,
  and BOTH horns are defective: bound ⇒ the nightly citation/live-probe seats silently lose
  a capability they demonstrably use and R4-2's premise is false; not bound ⇒ row 13's
  "holds on the Bash channel too" is void for the actual nightly workers, and whether
  debate.js-spawned seats inherit the headless `--settings` profile AT ALL is unprobed
  (interactive-run hook evidence does not carry layer-1 inheritance). Plus the trivial
  sub-leg: the retained "carve-out auto-approves read-only git regardless" comment is
  refuted by the same doc's bare-deny semantics; the four git allow rules are dead rules.
- **R5-3 (medium, sup R4-13+R4-1+R4-6).** The corroboration pool is defined NEGATIVELY and
  round 4 patched its three feeder surfaces in isolation — the third round of per-surface
  patches on this mechanism (R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13) with the root invariant
  still missing, the exact chain shape blue ended for status timers with R4-9. Sharpest:
  rung-2 Desktop tasks run locally against the same corpus — markerless sleeper dirs count
  NON-sleeper next morning AUTOMATICALLY, guarded only by an instructional adoption
  requirement (§4.1's own 0/175 class). OQ24's deferral rationale prices the paste-run
  case, never this one.
- **R5-2 (low-medium, sup R4-1).** Two body sites still specify the pre-trampoline shape:
  §3.4 ladder row 0 ("command markdown IS the payload") and §3.3's adopted Phase-4
  acceptance test — which the trampoline now FAILS BY CONSTRUCTION; the cheapest way to
  make the printed gate pass is to re-inline the payload, undoing R4-1. The test must be
  restated two-legged (wrapper --manual produces the run dir; the command produces NONE).
- **R5-4 (low-medium, sup R4-6).** The confinement clause decides sweep membership BY dir
  name against §1.5's still-standing "never the dir name" doctrine, and "the wrapper's own
  sub-run slug" is not a knowable convention after the hard-kill the clause exists for —
  log the sub-run path at creation and the doctrine survives intact.
- **R5-6 (low-medium, sup R4-4).** The `expected-memory-deny` bucket is keyed on TARGET
  PATH, not actor/cause: an injection-driven memory-poisoning attempt (row 14's
  Medium-likelihood channel) files as expected at any volume. Enforcement holds; the
  attack-attempt signal on the surface that steers red's future audits is erased by
  design. One actor/shape predicate restores it.
- Plus the cheap fan: **R5-5** (the ATTACHED `-o<value>` form escapes the new belt —
  merge-reproduced; do not extend the belt again, verify the hook allowlist's attached-form
  handling, OQ18(c)), **R5-7** (sleeper hooks.json bootstrap-guard coverage unstated + the
  flag-absent anomaly nags every never-configured install), **R5-8** (§0 tree still ships
  three of four code artifacts and registers no hook), **R5-9** (both `rejected` re-surface
  arms key on a rejection date no artifact records; `regression` token domain unpinned),
  **R5-10** (est_complexity note format unstated and the named source is empty at the pin —
  recommend, not block).

**Notes, not gaps:** #68375 now carries a `stale` label (bot auto-close drift risk — keep
re-checking); the r1-era learn.microsoft.com Task Scheduler settings URL 404s ([^MissedRun]
should re-point at the StartWhenAvailable API page); §7 lacks a round-4 update bullet
(sub-trivial); OQ17 remains bankable; the AlertFatigue pinnable replacement remains unbanked.

**Verdict: FAIL** — 10 open gaps, none yet risk-accepted by blue or ruled by the lead. Two
mediums (R5-1 the seat-surface scoping decision; R5-3 the pool root invariant) carry design
weight; both have cheap stated fixes; the remainder are propagation and spec-totality
residues. Mass 55.0 → 46.0 → 30.0: the trajectory converges, and nothing on this board looks
like it needs more than one honest round.

### LEAD

Round-5 adjudication, 2026-07-17. Docket: R5-1..R5-10 (10 items, ALL
`first_raise_successor` — `closed`/`rebuttal_sustained` are structurally unavailable;
blue has not yet answered them). Before ruling: debate.md and red/ledger.md read in
full; red/archive.md's Round-5 closure block read END-TO-END (R4-1..R4-16 closure
records, each naming its superseded ancestor and the round-4 revision), plus the round-4
LEAD carried rulings (R4-1..R4-14, debate.md) and the round-4 closure prose — so every
ancestor on every supersedes chain is read and named per ruling. Contested anchors
independently leaf-verified against blue/report.md at the judge seat: §4.2 "Bash IS
reachable (a rebuilt rung, the Workflow seat agents, profile drift)" (line 1356) vs §4.3
layer 4 (iii) "the workflow's SEAT AGENTS are full permission-engine + hook subjects"
(line 1422) — the leaf evidence there is sc-quality-gate firing on workflow-agent writes,
i.e. HOOK (layer-2) coverage, NOT settings (layer-1) inheritance, so R5-1's
inheritance-unprobed leg holds; §6 row 13 "the Bash read carve-out is closed STRUCTURALLY
round 4 … holds on the Bash channel too" (line 1597); §3.4 rung-0 cell "the /self-improve
command markdown is the wrapper's phase-1 prompt payload in EVERY mode" (line 946) and
§3.3/§8 acceptance test "`claude -p \"/self-improve\"` produces a run dir" (lines
900/2157); §1.5 "harvest.mjs reads ONLY the marker, never the dir name" (line 460) beside
the R4-6 date-key confinement clause; §1.5 "class=expected-memory-deny" bucket (line 504);
§4.2 `"Bash(* -o *)"` belt (line 1279); §1.4 rejected "default 90 days" (line 372) with no
dated status field in the §2.3 enum; §1.4 stage-2 "complexity note" (line 346); §3.4
"flag missing; last check <ts>" (line 1031); §0 tree (lines 103–130 — scripts/ +
docs/scheduling.md, NO hooks/ dir, NO SessionStart executable, sleeper-guard drawn under
scripts/ with a "(PreToolUse hook)" annotation but nothing registering it) vs the §0 prose
count of FOUR code artifacts + own hooks.json (lines 133–156). R5-5's attached-form gadget
re-reproduced at the judge seat: `git format-patch -1 -o/tmp/… HEAD` → exit 0, out-of-repo
patch, no belt match. Red's mechanism analyses hold against blue's own text in every
checked case.

Standing precedent applied (round-2/3/4 LEAD): a valid finding whose ACCEPTANCE would live
only in this ledger, not argued in the report, is not judge-accepted — "a judge-side
accept would ratify the silent skip" (R2-12); low-likelihood findings whose fix composes
at near-zero cost are carried, not accepted (R3-4); "wrong figures cannot be
risk-accepted into a report whose own discipline is priced arguments" (R3-12); an
overclaim/false narrative in an acceptance leg or risk row is a TEXT defect, not an
acceptable risk (R3-8/R4-5); an internal contradiction is certain-present, not a risk to
price (R4-4). All bind their round-5 successors directly. None of the ten fixes is a
priced likelihood×impact×complexity tradeoff — every required_fix is a stated
decision / one predicate / one clause / two tree lines that COMPOSES — so risk_accepted is
unavailable across the board; each item is carried with owed direction.

Rulings:

- **R5-1 — carried** (the round's priority; medium). Ancestors read: R4-3 (archive
  Round-5 block: the bare-`Bash` deny scoped to the TOP-LEVEL session, seat-surface scope
  contradictory/unestablished, settings inheritance never probed — declared into this
  successor) and R4-2 (archive Round-5 block: R4-2's own "Where Bash IS reachable (… the
  Workflow seat agents …)" premise vs layer 4 (iii) declared into this successor); chain
  roots R3-14/R3-15. Verified: line 1356 and line 1422 contradict as printed on whether
  the nightly FEOV seat agents sit inside the bare-`Bash`-deny boundary, and BOTH horns
  are defective — bound ⇒ the nightly citation/live-probe seats (whose Bash git method
  this design's own R3-15/R4-2/R5-5 reproductions used) silently lose Bash and wait
  forever under R2-6, and R4-2's premise is false; not bound ⇒ §6 row 13 / §4.3 layer 2's
  TOTAL "holds on the Bash channel too" is void for the actual nightly worker population,
  closure reverts to the enumeration R4-3 itself called non-load-bearing, and the fence
  covers WRITES not Bash READS on the seat surface (R1-13 read+egress re-opens). Decisive
  and un-priceable: the seat-coverage leaf evidence (sc-quality-gate on workflow-agent
  writes) is layer-2 HOOK coverage from INTERACTIVE runs; whether debate.js-spawned seats
  inherit the headless `--settings` layer-1 profile AT ALL is silently assumed — if seats
  run under `defaultMode: auto` (the very layer-masking fact R4-11 established), NONE of
  §4.2 binds the nightly workers. A contradiction plus a load-bearing closure claim void
  under one horn is a text/scoping defect, not a risk. Blue owes: state which horn holds
  and re-derive dependent text per red's required_fix — if seats BOUND (safe polarity),
  drop/rescope "the Workflow seat agents" in R4-2's reachability list, own the nightly
  seat-Bash capability cost in §2.2 step 4 / §5.2 (smoke-lane + citation-pass Bash-free or
  the loss priced), re-label the four dead `Bash(git …)` allow rules and drop/caveat the
  "auto-approves read-only git regardless" comment (refuted by the same doc's bare-deny
  semantics); if NOT bound, re-scope §6 row 13, §4.2, §4.3 layers 2/4, and §6 row 4 leg
  (a) to the top-level session and extend the fence to Bash READS on the seat surface (or
  generalize the R4-2 hook read-allowlist to the whole Bash channel per invariant 6).
  EITHER WAY add the seat-settings-inheritance probe as a named OQ23 acceptance leg —
  layer 4 (iii)'s interactive evidence does not carry the headless `--settings` case.

- **R5-2 — carried.** Ancestor read: R4-1 (archive Round-5 block: the thin-trampoline +
  payload-out-of-`commands/` close, with two body sites still specifying the OLD payload
  shape declared into this successor); chain roots R3-2, R1-25/R1-27. Verified: line 946's
  rung-0 cell asserts the round-3 payload shape the R4-1 edit list never covered — the
  ladder table is scheduling.md's shipped artifact, so a builder reading row 0 re-creates
  a payload-carrying invocable command, the exact vector R4-1 closed; and under R4-1
  `claude -p "/self-improve"` reaches Claude as trampoline instruction text, so the adopted
  Phase-4 acceptance test (lines 900/2157) is unsatisfiable as written and its cheapest
  pass re-inlines the payload — the gate-erosion path forecloses a judge-accept.
  first_raise_successor and the perverse incentive undoes a round-4 close, so this is
  carried, not accepted. Blue owes: rewrite the rung-0 cell to the R4-1 thin-trampoline
  shape (payload sourced from the wrapper's phase-1 prompt / continuous-learning skill,
  not commands/); restate the §3.3 Phase-4 acceptance test two-legged — (i) `node
  sleeper-wrapper.mjs --manual` produces the run dir + stub touching only research/+ideas/,
  (ii) `claude -p "/self-improve"` produces NO run dir (trampoline inertness becomes the
  verifiable R4-1 property); add "phase-1 prompt payload" to the propagation-grep list.

- **R5-3 — carried** (medium; design weight). Ancestors read: R4-13 (archive Round-5
  block: the gate-survival row is the THIRD per-surface patch on the negatively-defined
  pool, rung-2 poisoning automatic behind an instructional gate — declared into this
  successor), R4-1 (paste-run pool residual deferred via OQ24 pricing only paste-run),
  R4-6 (window-close is the second per-surface patch); chain roots R2-6, R3-3/R3-5, R1-25.
  Verified: §1.5 pool membership is default-negative (a dir corroborates unless
  affirmatively tagged `origin: sleeper`), and the three round-4 patches are per-surface
  bounds with no root fix — the exact per-surface chain shape blue itself ENDED for status
  timers with R4-9's root invariant, declined here. Sharpest and un-priced: rung-2 Desktop
  tasks run LOCALLY against the same corpus the nightly harvest reads, so markerless
  sleeper dirs count NON-sleeper next morning AUTOMATICALLY, guarded only by an
  instructional adoption requirement — the class §4.1's own 0/175 headline refutes — and
  OQ24's deferral rationale prices only the paste-run likelihood (a human ignoring their
  own contract), never the rung-2 automatic case. Per R2-12/R3-4 an accept must be argued
  in the report for the specific case, and red offers a cheaper root fix than the three
  bounds it replaces, so this is carried. Blue owes: state the root invariant once
  (mirroring R4-9) — the corroboration pool is defined POSITIVELY, a dir corroborates only
  with affirmative non-sleeper provenance, anything harvest cannot positively attribute is
  quarantined (fail-closed toward counts-for-nothing), OQ24's quarantine promoted from
  deferred to built and dissolving all three per-surface residuals; OR, if blue holds the
  deferral, the risk-accept must argue the rung-2 automatic-local-poisoning case
  SPECIFICALLY in the report, not just the paste-run case.

- **R5-4 — carried.** Ancestor read: R4-6 (archive Round-5 block: the confinement decides
  sweep membership BY dir name against §1.5's own doctrine and "the wrapper's own sub-run
  slug" is unknowable after a hard-kill — declared into this successor); chain roots R3-4,
  R2-5, R1-25. Verified: (a) the unqualified round-2 no-name-reads doctrine sentence (line
  460) stands in the same section whose round-4 clause decides sweep membership BY name —
  a standing doctrine-vs-mechanism contradiction, a text defect; (b) `<date>_self-improve/`
  is a real static convention but "the wrapper's own sub-run slug" is model-chosen per
  night, format-identical to a human's same-day research dir, and after the hard-kill the
  clause exists for, the wrapper that knew the slug is dead with the slug recorded nowhere
  durable — so confinement either cannot match the sleeper sub-run (row 10's auto-tag
  under-delivers) or must match `<date>_*` (sweeps human dirs, the harm (ii) prevents). Fix
  composes at near-zero cost (R3-4 precedent), so carried. Blue owes: (i) qualify the §1.5
  doctrine sentence — name-keying permitted only to CONFINE retroactive-uncertain sweeps,
  never to assign origin outside one; scope row 10 accordingly; (ii) make the slug knowable
  — the wrapper appends the sub-run dir PATH to the run-window log AT CREATION beside the
  START record, so confinement matches recorded paths and the doctrine sentence survives
  intact; (iii) state the mkdir-to-stamp bound explicitly.

- **R5-5 — carried.** Ancestor read: R4-2 (archive Round-5 block: the ATTACHED `-o<value>`
  form escaping the new `Bash(* -o *)` belt, belt-only, hook allowlist the load of record —
  declared into this successor); chain root R3-15. Verified at the judge seat: `git
  format-patch -1 -o/tmp/… HEAD` → exit 0, out-of-repo patch, no belt pattern matches — the
  enumerate-and-extend regress recurring one lexical form deeper INSIDE R4-2's own repair.
  NOT risk_accepted despite red grading it low and belt-only: the retained R4-2 comment
  ("`-o` matched none of the three long-form denies") implies the belt now catches `-o`
  when it does not catch the attached form — a false-completeness claim in-file is a text
  defect (R3-8), and the actual load of record (the hook read-allowlist's handling of
  attached-form flags) is ASSUMED, not leaf-tested, so the "hook is the real close"
  argument rests on an unverified premise. Carrying forces that probe. Blue owes: do NOT
  extend the belt pattern again; state in-file that the belt is KNOWN-incomplete on
  attached short forms and the hook allowlist is the actual close; add `git format-patch
  -o<attached>` to OQ18(c)'s probe matrix so the hook allowlist's attached-form handling is
  leaf-tested, not assumed.

- **R5-6 — carried.** Ancestor read: R4-4 (archive Round-5 block: the `expected-memory-deny`
  bucket keyed on TARGET PATH not actor/cause, laundering attack-attempt evidence on the
  memory surface — declared into this successor); chain roots R3-3, R1-25. Verified: line
  504's bucket keys on `.claude/agent-memory/**` target, so the nightly red-merge seat's
  one expected protocol write AND an injection-driven memory-poisoning ATTEMPT (§6 row 14's
  Medium-likelihood channel) at arbitrary volume and target within agent-memory/** both
  file as `expected-memory-deny`, unflagged. This is a self-inconsistency with the design's
  own invariant 7 (a control that records without surfacing is telemetry) on exactly the
  corpus that steers the adversary's future audits — a design-coherence defect, not a
  priced risk (enforcement holds either way, but the tamper-attempt-must-be-visible
  doctrine is violated by fiat). Red's round-4 distinguishability question was answered by
  DEFINING the class non-anomalous rather than by mechanism. Blue owes: one predicate in
  the parser — `expected-memory-deny` ONLY for the nightly red-merge seat's own
  record-new-patterns write shape (seat role + pattern-file target), bounded ≤N (small) per
  run; any other actor, target subdir, or volume over the bound surfaces as a distinct
  `unexpected-memory-write-attempt` class on the doctor line (persists like TAMPER).

- **R5-7 — carried.** Ancestors read: R4-8 (archive Round-5 block: homing the hook in
  sleeper-service's OWN hooks.json extends the empty-bin crash-storm surface to every
  interactive session while §6 row 9's bound cites prosthetic-conscience's hooks.json —
  declared into this successor) and R4-14 (flag-ABSENT is the DEFAULT state of a fresh
  install so the anomaly print nags every never-configured operator — declared into this
  successor); chain roots R3-10, R3-9. Verified: (a) the sleeper SessionStart hook lives in
  a DIFFERENT hooks.json and fires in every interactive session, so an unguarded sleeper
  hooks.json crash-storms all interactive work during a cache update — a surface row 9
  (nightly-scoped) never contemplated, guard coverage asserted nowhere; (b) printing
  flag-absent (line 1031) as a per-session ANOMALY is the alert-fatigue mode the report's
  own Dependabot evidence catalogs, aimed at its own dead-man channel — never-configured vs
  flag-lost undistinguished. Both trivial composing fixes; the (b) leg is a
  self-inconsistency with the report's own alert-fatigue evidence. Blue owes: (i) state
  that sleeper-service's hooks.json wraps its SessionStart command in the bootstrap guard,
  re-point row 9's bound to it, and add the hook to the empty-bin acceptance check; (ii)
  gate the flag-missing anomaly on a prior `last-successful-run` record — no prior run ⇒
  never configured ⇒ silent or one-time notice, not a recurring anomaly.

- **R5-8 — carried.** Ancestor read: R4-8 (archive Round-5 block: the §0 TREE never
  reconciled — no hooks/hooks.json entry, no SessionStart executable, nothing registering
  the sleeper-guard PreToolUse hook — declared into this successor); chain roots R3-10,
  R2-19. Verified: the §0 tree (lines 103–130, labeled "the implementable shape") ships
  three of the four R4-8-counted code artifacts and NO hook registration — the sleeper-guard
  is drawn under scripts/ with a "(PreToolUse hook)" annotation but nothing registers it,
  and the SessionStart executable + its hooks.json are absent from the tree entirely,
  while the same section's prose (lines 133–156) counts FOUR artifacts + the plugin's own
  hooks.json. The enumeration was repaired three times (R2-19/R3-10/R4-8) while the TREE
  never was — exhaustive-sweep-omits-own-specimen, one artifact deeper than R3-10. A
  builder implementing from the "implementable shape" ships a non-functional (unregistered)
  enforcement hook — a build defect, not a skimmer's cosmetic. Trivial fix. Blue owes: add
  `hooks/hooks.json` (registering the sleeper-guard PreToolUse hook and the SessionStart
  staleness hook) and the SessionStart executable to the §0 tree.

- **R5-9 — carried.** Ancestor read: R4-9 (archive Round-5 block: both arms of the
  `rejected` clause key on a rejection DATE no artifact records, and the `regression` token
  is not in the §2.3 enum with its domain unpinned — declared into this successor); chain
  roots R3-13, R2-11, R1-22. Verified: line 372's clause (b) needs the rejection date to
  split the timeline (both the 90-day arm and the pre-rejection-rate arm), but stub
  filenames date the MINT and status edits are undated — a builder implements
  filename-date+90d (measures mint age: a stub rejected at day 80 re-surfaces in 10 days)
  or silently drops the rate arm (printed invariant ≠ built invariant — the R4-12 defect
  class INSIDE the R4-9 repair). This is the "wrong/uncomputable printed figures cannot be
  risk-accepted" case (R3-12), trivial to fix. Blue owes: one dated field (the status line
  records the last human status change, e.g. `status: rejected 2026-07-17`; both arms key
  on it; harvest parses it) + one clause pinning `regression`'s domain (e.g. docket-flag
  only; the stub stays `graduated`; the regression entry may mint a NEW stub) and the
  `rejected-recurring` setter.

- **R5-10 — carried** (recommend-tier; trivial). Ancestor read: R4-12 (archive Round-5
  block: "harvest parses that note's value" states no FORMAT and no backlog entry carries a
  parseable complexity field at the pin — declared into this successor). Verified against
  red's L1 leaf (`git show 7bc501e:ideas/backlog.md`, exhaustive token grep): no structured
  complexity field on any of the 25 items, so the divisor is universally inert against the
  actual corpus while line 346's text implies it activates. Red grades trivial and
  "recommend, not block," and the safe default (1) fully bounds the harm — the nearest
  risk_accept candidate on the board — but the complexity_cost is trivial (one clause) and
  the defect is a false implication (text implies the factor activates; it never does), so
  there is no likelihood×impact×complexity tradeoff favoring rejection; per precedent a
  cheap-to-fix false-implication is carried, not accepted. Blue owes: state the token
  convention in §1.4 stage 2 (e.g. a literal `cx:<1-5>` token in the backlog entry;
  anything else = default 1) and note the field is currently unpopulated (forward-looking
  curation surface).

Cross-cutting direction (endorsing red's round-5 shape as LEAD direction for round 6, not
a new gap): red's dominant thread — "round 4's STRUCTURAL closes were derived over the
TOP-LEVEL session, not over the two surfaces where nightly work actually executes" — is
correct. R5-1 (the Workflow seat population) and R5-3 (the negatively-defined
corroboration pool) are the two design-weight residues, and both have the same remedy
shape blue has used to good effect twice already: state the invariant ONCE over the true
surface and derive rather than patch per-surface. For R5-1 the invariant is "layer-1
`--settings` inheritance is established (probed) or fenced for the seat population, not
assumed from interactive layer-2 evidence"; for R5-3 it is R4-9's own positive-provenance
move applied to the pool. The remaining eight are propagation/spec-totality residues whose
fixes compose. Blue SHOULD prefer the two structural statements over another lap of
neighbor-by-neighbor; red grades the result on the carried items' own terms.

Deadlock check: FALSE — 10 gaps carried this round (no new gaps were raised this round,
but a carried gap alone defeats deadlock). The debate continues into round 6.

### BLUE

Round 5. All 10 carried gaps addressed additively; no substance subtracted; zero grade
disputes a fifth round (the lead priced every required fix as a composing decision/predicate/
clause — none is a likelihood×impact×complexity tradeoff favoring rejection, so contestation
could not honestly beat absorption again). I took the lead's cross-cutting direction
literally: the two design-weight residues (R5-1 seat population, R5-3 corroboration pool)
got the same remedy shape I used to good effect twice — state the invariant ONCE over the
true surface and derive, rather than another per-surface lap.

**R5-1 — the seat surface, horn picked with the inheritance-independent close.** The
round-4 contradiction was real: §4.3 layer 4 (iii) made the seats full permission-engine
subjects while §4.2's R4-2 bullet listed "the Workflow seat agents" as a surface where Bash
was reachable. I picked the safe horn — seats BOUND, inside the sleeper boundary as a design
requirement — and answered the lead's decisive leg (the seat-coverage leaf evidence is
layer-2 HOOK coverage from INTERACTIVE runs, so headless layer-1 `--settings` inheritance is
unprobed) not by asserting inheritance but by making closure NOT depend on it: the
sleeper-guard hook is generalized per invariant 6 to deny the whole Bash channel for
sleeper-marked runs, and hook seat-coverage IS the leaf-verified part. So even if OQ23(e)
finds seats outside the `--settings` profile, Bash is closed on the seat surface by the hook,
and if BOTH the inheritance probe and the headless-hook probe (OQ2) failed, the phase-0
canary aborts the run — fail-closed on both at once. The nightly seat-Bash capability cost is
owned in §2.2 step 4 and §5.2: seats run Bash-free (repo state wrapper-staged; the research
protocol forbids shell file access anyway), and the genuine loss — command-execution leaf
probes like this run's own gadget reproductions — is a graduation capability, with any seat
that wanted Bash saying so in visible friction. The four dead `Bash(git …)` allow rules and
the refuted "auto-approves read-only git regardless" comment are corrected at both sites.

**R5-3 — the pool root invariant, promoting OQ24 to built.** Red's sharpest case (rung-2
Desktop tasks poisoning the shared corpus automatically, no human act anywhere) was correct
and un-priced. I stated the invariant once: the corroboration pool is POSITIVELY defined — a
dir corroborates only with affirmative non-sleeper provenance (pre-deployment-epoch date, or
an interactive-origin marker stamped at creation by a one-line FEOV cross-plugin edit);
anything unattributable is quarantined counts-for-nothing and surfaced for one keystroke.
That dissolves the R4-1/R4-6/R4-13 residuals in one predicate and resolves OQ24 by promotion
— in generalized provenance-keyed form, NOT the loop-shaped-dir schema heuristic the round-4
deferral considered (which would have paid schema false-positives on ordinary research dirs;
the provenance form pays none — an ad-hoc human dir is surfaced, never misclassified either
way). This is the R4-9 move applied to the pool, exactly as the lead directed.

**The cheap fan (R5-2/4/5/6/7/8/9/10), all absorbed and propagated.** Rung-0 cell + Phase-4
test restated to the trampoline shape with inertness as the verifiable property (R5-2); the
name-key doctrine qualified and the sub-run slug made knowable by logging its path at
creation (R5-4); the `-o` belt declared known-incomplete on attached forms with the hook
allowlist as the close of record and OQ18(c)(3) forcing the leaf-test — NOT another belt lap
(R5-5); the memory-deny bucket narrowed to an actor/shape predicate so poisoning attempts
surface like TAMPER (R5-6); sleeper-service's hooks.json bootstrap-guard coverage stated and
the flag-absent anomaly gated on prior run history (R5-7); the §0 tree finally draws all four
code artifacts plus hook registration (R5-8); a dated status field both `rejected` arms key
on, with `regression`/`rejected-recurring` domains pinned docket-only (R5-9); the
est_complexity token convention stated with its source honestly empty at the pin (R5-10).

Propagation was the run-3 regression class, so I greped every retracted token report-wide in
both directions (log in CHANGELOG Round 5): `read-only git regardless`, `the Workflow seat
agents`, `full permission-engine + hook subjects`, `phase-1 prompt payload in EVERY mode`,
`counts them non-sleeper`, `Deferred, not built`, `produces a run dir` — remaining hits are
all correction-context or supersession-noted, zero live stale assertions. Also banked red's
corroborations in §7 (ten living-source sets zero-drift, cap/HALT triple-recomputed,
[^Pricing] zero-drift) and the two notes (#68375 `stale` label; [^MissedRun] URL 404
re-pointed). claim_count 155 → 161.

Verdict I'd argue for: the two mediums are closed by stated invariants over the true surface
(not re-argued, not risk-accepted — the lead correctly ruled none was a priceable tradeoff),
the eight residues are spec-totality and propagation corrections, and the board should be
clean. OQ23(e) (seat `--settings` inheritance) and OQ18(c)(3) (hook attached-form handling)
are the two new build-time acceptance legs; both are belt-verification, neither load-bearing
given the hook-channel and bare-deny closes.
