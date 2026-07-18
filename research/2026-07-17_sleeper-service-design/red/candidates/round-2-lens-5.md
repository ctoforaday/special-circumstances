# red candidates — round 2, lens 5 (logic & completeness)

Audit surface: full re-read of `blue/report.md` (1387 lines, three consecutive Read
windows, 2026-07-17, post-round-1 revision). Lens: leaps of faith, missing
counterarguments, unexplored alternatives, template compliance — with round-2 emphasis on
whether the round-1 repairs COMPOSE (sibling-repair-composition, self-defeating-mitigation,
incomplete-repair patterns from red memory). Lens-scoped ids only; merge assigns stable
R2-N ids and rules on lineage (several findings below are successor-candidates on round-1
closures — flagged inline; the merge must verify against `red/ledger.md`/archive before
minting).

---

## L5-F1 — Denial canary: actor, observer, and abort mechanism all unspecified; as printed, the abort is an instructional gate of the class §4.1 refutes

- **location:** §2.2 step 0, sentence: "First in-session action: DENIAL CANARY — attempt
  one out-of-fence write, abort the run unless DENIED." Cross-anchor §4.3 layer 2: "fence
  ENFORCEMENT is proven per run by the wrapper's step-0 denial canary — one attempted
  out-of-fence write, run aborted unless DENIED".
- **finding:** The canary is listed under "PREFLIGHT (wrapper, OUTSIDE the model session)"
  yet is labeled the "first IN-SESSION action" — the write attempt must come from the model
  session (that is what is being tested), but the abort decision is attributed to the
  wrapper. No mechanism connects them. `claude -p` takes one prompt; the JSON envelope
  (with `permission_denials`) arrives at session END — a wrapper checking it post-hoc has
  already let the whole run execute with a possibly-dormant fence. If instead the abort is
  the model's own responsibility ("if your write succeeds, stop"), that is a prompt-level
  MUST — exactly the gate class §4.1 refutes with the report's own evidence ("the MUST-try
  clause has no observable"; 0/175 batching compliance). Workable mechanisms exist
  (stream-json two-phase drive: wrapper sends canary prompt, parses the denial event,
  only then sends the real prompt; or a separate probe-P2-style micro-session with an
  argued same-environment guarantee) — but the report specifies none, and the two named
  anchor points contradict each other on who acts. Successor-candidate on the R1-28 fix
  (closure-with-regression class: the fix's mechanism is underspecified at the seam that
  makes it a gate rather than a wish).
- **grading:** likelihood: medium-high (an implementer must invent the mechanism; the
  natural single-prompt reading yields either post-hoc check or instructional abort);
  impact: medium-high (this is the control that converted the fence from fail-open to
  fail-closed — R1-28's entire close); complexity: low (specify the two-phase drive or
  probe-session variant + the same-env argument). → severity **medium-high**
- **corroboration confidence:** high (in-document contradiction; both quotes verified in
  full read).

## L5-F2 — The canary proves "something denied one write," not "the hook fence is enforcing" — and the fence and layer 1 draw the SAME boundary, so a fence-dormant run with live permissions passes the canary

- **location:** §4.3 layer 2, sentence: "fence ENFORCEMENT is proven per run by the
  wrapper's step-0 denial canary … so marker loss … fails CLOSED instead of silently
  dropping to the permission-rules-only configuration H4 refuted (R1-28)." Cross-anchor
  §8 OQ2: "the smoke test must also verify the canary itself aborts on a deliberately
  fence-dormant run."
- **finding:** The write surface of layer 1 (`Edit(<REPO>/research/**)`,
  `Edit(<REPO>/ideas/**)` under `dontAsk` auto-deny) and the hook fence ("deny writes
  outside research/ + ideas/") are the SAME boundary. Therefore every out-of-fence canary
  target is also outside layer 1's allow set, and whenever layer 1 is live the canary is
  denied REGARDLESS of whether the hook is dormant. The common drift case R1-28 was minted
  for — env marker lost, hook silently dormant, operator-owned settings file intact — is
  exactly the case where the canary passes and the run proceeds under the
  permission-rules-only configuration H4 "REFUTED at the leaf node," now with a false
  "enforcement PROVEN per run" stamp on it (§0 tree). The canary detects total gate
  collapse (both layers dead), which is real value, but the §4.3 sentence claims fence
  enforcement specifically. OQ2's own acceptance test ("canary aborts on a deliberately
  fence-dormant run") will FAIL as designed — on a fence-dormant, permissions-live run the
  canary is denied and does not abort — so the design's stated test contradicts its stated
  claim; one of them is wrong. Fix requires either deny-provenance discrimination (a
  PreToolUse hook block is distinguishable from a permission-rule denial in stream-json
  event/text — verify) or an honest downgrade of the claim to "proves at least one deny
  layer is live." Successor-candidate on R1-28 (self-defeating-mitigation class: the
  control cannot observe the specific failure it was added to catch).
- **grading:** likelihood: medium (marker loss with intact settings is the modal drift
  scenario — the layers fail independently and layer 1 is the more static artifact);
  impact: medium-high (silent reliance on the empirically-leaky #22055/#6631 layer, plus
  false per-run assurance — worse than round 0's honest fail-open in the assurance
  dimension); complexity: low-medium (parse deny provenance, or re-scope the claim and
  keep the canary as collapse-detection belt). → severity **medium-high**
- **corroboration confidence:** high (boundary coincidence is verifiable from the §4.2
  profile and §0 tree text side by side).

## L5-F3 — §2.2 step 2 still executes a script IN-SESSION; R1-16 propagation miss

- **location:** §2.2 step 2: "SCORE (script; cheap tier only if scoring needs language):
  the §1.4 formula; log the full scored table to the run dir so the pick is auditable."
- **finding:** R1-16's fix deleted `Bash(node scripts/*)` and asserts "the session never
  invokes node" (step 1) / "removes ALL script execution from the session" (§4.3 layer 4).
  Step 2 as printed is an in-session step whose primary mode is "script." Under the §4.2
  profile (Bash = pinned git argv only, dontAsk auto-deny) a session-side script
  invocation at step 2 is auto-denied and the loop stalls — the exact failure mode
  R1-16(d) named for the layout mismatch. The repair is available in the report's own
  text: §1.4 stage 1's wrapper-staged docket already carries a `score` column, so step 2
  is either wrapper-work already done at step 0 (and the step should say "read the staged
  scored table") or model-side language-tier scoring (and the word "script" must go).
  Incomplete-repair class: the R1-16 propagation grep (`Bash(node`) caught the profile
  sites but not this step's execution-locus word.
- **grading:** likelihood: medium (an implementer following the numbered steps builds the
  denied path); impact: low-medium (loop stalls loudly, or implementer improvises);
  complexity: trivial (reword the step). → severity **low-medium**
- **corroboration confidence:** high.

## L5-F4 — Step 4's FEOV invocation is never reconciled with "no script execution in the session" — one branch of the unstated disjunction breaks the loop, the other breaks risk row 4's re-argued acceptance

- **location:** §2.2 step 4: "invoke the frank-exchange-of-views machinery at smoke scale
  (1 lane, 1 round, 1 citation pass, bulk model — the measured ~50k-token smoke shape)";
  vs §4.3 layer 4: "the round-1 shape (R1-16) removes ALL script execution from the
  session"; vs §6 row 4: "the session's Bash surface is now pinned-argv read-only git
  commands — no script execution, no model-chosen subprocess write targets."
- **finding:** The FEOV machinery is scripted by doctrine — [^ResearchCommand] quotes the
  command's own contract ("the mechanics are scripted"), and a research run entails at
  minimum the workflow/setup script and the capture step "emitting cost.md and
  run-record-audit.md." The report never states HOW step 4 executes under the sleeper
  profile. Two possible readings, both unexamined: (a) FEOV's scripts run via the session's
  Bash tool → auto-denied under the git-only allowlist → the nightly loop's core research
  step stalls every night (the design's daily driver is incompatible with its own
  permission profile); or (b) workflow scripts execute through a harness-side runner
  outside the Bash permission surface → then "no script execution in the session" and row
  4's channel-narrowness leg ("the only subprocess reachable is the pinned git set") are
  both false — the model reaches an entire script-running workflow surface nightly, whose
  write behavior (cost.md, run-record-audit.md, telemetry appends — all legitimately
  in-fence, but executed by unfenced subprocesses) is exactly the layer-4 residual, and
  R1-21's re-argued acceptance (leg a) rests on a premise the loop's step 4 negates
  nightly. Whichever branch is factual, a round-1 conclusion fails: either R1-16's profile
  is incomplete for the loop's own spec, or row 4's Low re-grade and the "removes ALL
  script execution" sentence are overclaims. Sibling-repair-composition class: R1-16's
  hardening and step 4's requirement were each verified alone, never together.
- **grading:** likelihood: high (one branch is true by construction); impact: high (daily
  driver viability, or the design's largest accepted residual mis-argued); complexity:
  low-medium (determine the execution locus by probe; then either enumerate FEOV's script
  needs wrapper-side/pinned-argv, or re-scope row 4 and layer 4 to cover the workflow-
  runner channel and argue hook coverage over its writes). → severity **high**
- **corroboration confidence:** high on the in-document contradiction; medium on which
  branch is factual (execution locus not verifiable from the report alone — that is the
  gap).

## L5-F5 — §3.3(2a) grants and revokes the PDF/arXiv tool surface in the same sentence

- **location:** §3.3 item 2: "(a) the loop's MCP profile is `--strict-mcp-config
  --mcp-config <sleeper-mcp.json>` naming **qmd only** (fewer servers is the #68375
  mitigation; research subagents reach pdf/arxiv tools via ToolSearch per the shipped
  seat grants)".
- **finding:** ToolSearch discovers deferred tools from DECLARED servers; under
  `--strict-mcp-config` with a config naming qmd only, no pdf-reader/arxiv-latex tools
  exist to discover — the parenthetical asserts a capability the flag it annotates
  removes. Consequence unowned: the research-protocol skill preloaded into step 4 carries
  "YOU MUST try the document-extraction MCP tools before grading down on a lossy fetch" —
  a MUST that is structurally unsatisfiable in every nightly run, which (per the skill)
  makes every nightly citation audit that hits a PDF "an incomplete audit" by the
  protocol's own definition. Either the sleeper MCP config names the pdf/arxiv servers too
  (re-opening the #68375 several-servers trade the sentence just made), or nightly stubs
  carry a standing degraded-citation caveat — a fine trade to accept, but it must be
  chosen, stated, and reflected in the stub contract's confidence field.
- **grading:** likelihood: certain (textual contradiction); impact: low-medium (stub
  citation quality + a protocol MUST silently violated nightly); complexity: low (pick the
  trade; one sentence + one stub-contract note). → severity **low-medium**
- **corroboration confidence:** high.

## L5-F6 — $2–5/night ceiling × 30 nights = $60–150/month vs a ~$50/month cap: the arithmetic composition is never reconciled

- **location:** §5.2 tier table, row "whole daily run — `--max-budget-usd` **$2–5**;
  monthly ledger cap ~$50 (operator-tunable)"; cross-anchor §1.4: "What it costs:
  $2–5/night ceiling (ledger-capped ~$50/mo)".
- **finding:** At the stated per-run ceiling the monthly cap binds between day 10 (at $5)
  and day 25 (at $2) — i.e. the design's normal-operation envelope EXPECTS the monthly
  guard to trip mid-month, every month, unless typical spend is well under $1.67/night.
  Typical spend is never stated (the smoke shape "prices in single-digit dollars" is the
  only hint, and it is compatible with both sides). Downstream incoherence: §6 row 6
  treats skipped days as "costs nothing" anomalies and row 15's dead-man line
  distinguishes "missed day (harmless)" from "loop dead three weeks" — but a cap that
  trips routinely at day ~20 manufactures a legitimate multi-day skip streak every month,
  which the doctor line cannot distinguish from death without also surfacing the skip
  REASON. Unreconciled-numeric-floors class (red memory: recompute, don't re-read). Fix
  is one honest sentence: state expected per-run spend with the measured smoke figure,
  and either size the cap to cadence × expected (with headroom factor) or declare
  cap-trip-at-month-end the intended throttle and teach the dead-man surface the
  difference.
- **grading:** likelihood: medium-high (arithmetic is certain; whether real spend
  approaches the ceiling is the open variable); impact: low-medium (availability +
  monitoring semantics, not safety); complexity: trivial. → severity **low-medium**
- **corroboration confidence:** high (both figures quoted from the report; multiplication
  re-performed).

## L5-F7 — R1-25's cap mechanism is undecided between a hard bar and a labeled pass — an implementer cannot build it, and row 10's re-grade depends on which it is

- **location:** §1.5: "requires at least one non-sleeper occurrence before a sleeper-origin
  class enters the top-3 comparison — or, failing that, flags it `sleeper-only` in the
  stub provenance so the human sees the loop citing itself"; §6 row 10: "requires
  non-sleeper corroboration (or a visible `sleeper-only` provenance flag) before a
  sleeper-origin class reaches the pick".
- **finding:** "Requires X — or, failing that, flags" is not a requirement; it is a label.
  The two dispositions are materially different: a hard bar means a sleeper-only class can
  NEVER reach the pick (strong containment, but a real signal the loop alone observes is
  permanently unpickable — an unexplored cost the report should own); a flagged pass means
  the loop CAN pick its own uncorroborated complaint and the containment is the human
  noticing a provenance flag at stub-triage — under §1.1's own Dependabot fatigue
  base-rate, a weak reed. Row 10's Low-Medium re-grade silently assumes the strong
  reading while the text licenses the weak one. Undecided-"or" / false-equivalence-
  disjuncts class (red memory: no undecided "or" in a mechanism spec). Decide, state the
  cost of the chosen branch, and make row 10's grade track it.
- **grading:** likelihood: high (the ambiguity is textual and load-bearing); impact:
  medium (the guard bounding self-amplification — R1-25 was graded high); complexity:
  trivial (one decision, two sentence edits). → severity **medium**
- **corroboration confidence:** high.

## L5-F8 — The R1-26 snapshot log has no defined reader or comparison step — the same undefined-receiver defect R1-29 fixed for loop death, unfixed for tamper evidence

- **location:** §4.3 layer 5: "a write-then-revert inside the run shows as a start/end
  delta the model cannot reach, restoring the durable tamper-evidence property the DGM
  analogy actually requires."
- **finding:** Shows — to whom, when? No wrapper auto-compare is specified, no abort or
  flag on delta, no doctor line, no named review cadence. The report itself established
  (R1-29 fix, §3.4) that "loudly" needs a defined reader, and gave the dead-man surface
  one; the snapshot channel — which backstops risk row 4's acceptance leg (b), "detection
  outside the model's reach" — terminates in an operator-owned log nobody is tasked to
  read. Run N+1 fires before any human looks; a delta could sit unread indefinitely, and
  detection-that-nobody-reads is not detection (window-without-a-watchman class; also the
  DGM analogy's own force — the lineage was INSPECTED, which is how the marker-removal
  was caught). Cheap completion in the design's own idiom: the WRAPPER compares start/end
  hashes at step 7 and, on delta, writes the dead-mark + refuses the next fire until an
  operator clears it, and doctor gains a "last snapshot delta" line beside the dead-man
  line.
- **grading:** likelihood: low (deltas should be rare); impact: medium-high (row 4's
  acceptance leans on this leg; an unread log restores the self-erasing-evidence problem
  one level up); complexity: low (wrapper diff + doctor line). → severity **medium**
- **corroboration confidence:** high.

## L5-F9 — The operator-owned plugin copy has no creation/refresh owner, no freshness check, and "the human-gated plugin update path" names a path defined nowhere

- **location:** §3.2: "`--plugin-dir` pointing at an **operator-owned, read-only plugin
  copy outside the write surface** (e.g. `~/.claude/sleeper/plugins/`, refreshed only by
  the human-gated plugin update path — NEVER the repo's `plugins/` working tree".
- **finding:** R1-15's fix mints a new operational artifact and leaves it ownerless: who
  creates the copy, what "the human-gated plugin update path" concretely is for a
  hand-made directory copy (the marketplace update dance covers `~/.claude/plugins/`
  cache, not this bespoke dir), and what detects staleness. Consequence of the gap: guard
  and hook fixes merged to the repo by PR do NOT reach the executing copy until a human
  remembers an undefined refresh step — the fence the nightly run executes can lag every
  round-1-class hardening indefinitely, silently. Doctor gets no "sleeper plugin copy
  matches installed version" line; scheduling.md is not tasked with the refresh recipe.
  Policy-without-mechanism class, minor key: the pin was the right fix; the pin's
  lifecycle is unbuilt.
- **grading:** likelihood: medium (copy-drift is the default outcome of ownerless copies);
  impact: low-medium (stale guard code, delayed hardening — not a new write channel);
  complexity: low (doctor freshness line + a named refresh step in scheduling.md).
  → severity **low-medium**
- **corroboration confidence:** high.

## L5-F10 — Default-rung drift: §1.4 (R1-14 fix) makes rung 0 "the DEFAULT and may be terminal"; §3.4's ladder still stamps rung 1 "RECOMMENDED default"

- **location:** §1.4: "**rung 0 — manual `/self-improve`, same code path, zero standing
  cost — is the DEFAULT and may be terminal**"; §3.4 ladder, rung 1: "**RECOMMENDED
  default** — the only local option where every flag is explicit and version-pinnable".
- **finding:** The round-1 null-alternative paragraph re-based the design's default on
  rung 0 with cadence-as-hypothesis; the ladder table was not re-touched and still reads
  as recommending scheduled rung 1 outright. Reconcilable ("recommended default AMONG
  scheduled rungs, once the human opts in") but as printed the two sections answer "what
  do we ship as the default?" differently — and §3.4 is the table an implementer reads.
  Incomplete-repair class (body lags the repaired section).
- **grading:** likelihood: certain (textual); impact: low (one qualifier); complexity:
  trivial. → severity **low**
- **corroboration confidence:** high.

## L5-F11 — Gate-survival table, rung-0 column: the compound L2 cell claims YES for a canary that cannot exist without the wrapper; and rung-0 (the default mode) spend never enters the monthly ledger

- **location:** §3.4 gate-survival table, row "L2 hook fence + step-0 denial canary",
  R0 cell: "YES (cache copy)"; row "Wrapper controls (preflight, ledger, idempotence,
  resume cap, snapshots)", R0 cell: "n/a (human present)".
- **finding:** Two legs. (a) The L2 row is compound (fence + canary); at rung 0 there is
  no wrapper, hence no canary — the same table's wrapper row says so — so "YES" overstates
  the compound control by its verification half; honest cell is "fence YES (cache copy);
  canary n/a". The table exists (R1-27) precisely to stop rung cells from overstating
  gate presence. (b) The ledger is wrapper-written (§5.1); rung-0 manual runs bypass the
  wrapper, so in the design's own DEFAULT mode (per §1.4/R1-14) run costs are never
  appended — a later opt-in to scheduling starts against a ledger blind to all prior
  manual spend, and mixed manual+scheduled months undercount systematically. "n/a (human
  present)" is a fair risk posture for rung 0 alone but the composition with the cap's
  month-to-date arithmetic is unstated. Exhaustive-sweep-omits-hard-case class: the table
  swept 5 rungs × 6 controls and the report's own newly-default rung holds the soft cells.
- **grading:** likelihood: certain (textual) for (a), medium for (b) mattering; impact:
  low (calibration + a bounded undercount); complexity: trivial (cell edit; one sentence
  saying manual-run spend is out-of-ledger by design or wrapper-wrap the manual path).
  → severity **low**
- **corroboration confidence:** high.

## L5-F12 — The idempotency/resume gate still reads loop-writable state — R1-19's explicitly-named sibling was neither fixed nor risk-accepted

- **location:** §2.2 step 0: "today's run-dir idempotency check (exists+complete → exit 0;
  exists+incomplete → resume, capped at k=3 attempts, then mark dead + fresh dir next
  fire)"; §3.4: "the wrapper writes a `DEAD` marker plus a dated abort record".
- **finding:** R1-19's problem statement named it: "Same shape, lower stakes: the
  idempotency gate reads loop-writable `research/**` state." The round-1 fix relocated
  the LEDGER and stopped. The completeness signal the exit-0 decision reads lives in the
  run dir (loop-writable); the DEAD marker's location is unstated — row 15 puts "dated
  skip/abort records in the operator-owned dir" but the marker the NEXT FIRE's check
  consumes is naturally in-run-dir, and the report does not say. A run that mangles its
  own completeness state (no malice needed — a partial write at abort) flips the wrapper
  between resume/fresh/exit-0 on model-authored evidence. Stakes remain low (worst case:
  duplicate or skipped run, bounded by k=3 and the per-run budget) — which may well be
  the right risk-accept — but the report's own discipline is that risk-acceptance is
  argued, not silent, and this half of R1-19 got neither the fix nor the argument.
  Partial-closure residue on R1-19; merge should rule whether this travels as
  closure-with-regression lineage or a fresh low gap.
- **grading:** likelihood: low-medium (abort-time partial writes are the realistic path);
  impact: low (bounded by resume cap + budget); complexity: low (wrapper derives
  completeness from its own step-7 record in the operator-owned dir instead of run-dir
  state; state DEAD-marker location). → severity **low**
- **corroboration confidence:** high.

---

## Template/protocol compliance check (lens duty, no finding minted)

- §7 self-audit updated for round 1; §8 OQs renumbered coherently (OQ16–19 added, OQ8
  marked RESOLVED in place — correct practice).
- CHANGELOG claim_count re-derivation (142) checks arithmetically against its stated
  method (124 + 8 steps + 10 contract fields = 142). Not re-enumerated line-by-line this
  round; the method is stated and reproducible.
- Minority-marker convention maintained through round-1 insertions.
- No template misfit observed; no friction to report for this seat.

## Note to merge (lineage candidates)

- L5-F1/L5-F2 → successor-candidates on **R1-28** (canary fix): closure-with-regression
  if merge accepts either.
- L5-F3/L5-F4 → compose against **R1-16**; L5-F4 also destabilizes **R1-21**'s re-argued
  row-4 acceptance (leg a) and §4.3 layer 4's "removes ALL script execution."
- L5-F8 → sibling of **R1-26** (the fix's output channel lacks the reader R1-29 taught
  the design to name); touches row 4 acceptance leg (b).
- L5-F12 → partial-closure residue on **R1-19** (the gap's own text named the sibling).
- L5-F9 → residue on **R1-15**; L5-F10 → residue on **R1-14**; L5-F11(a) → residue on
  **R1-27**.
- Archive spot-check obligation: closure index is still empty (round-1 closures happen at
  this round's merge), so no archived closures existed to sample at this seat.
