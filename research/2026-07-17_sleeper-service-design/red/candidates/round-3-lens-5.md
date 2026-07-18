# red candidate pass — round 3, lens 5 (logic & completeness)

Auditor: red L5 seat. Full living report re-read in three consecutive whole windows
(lines 1–608, 609–1108, 1109–1641 of `blue/report.md`, 1641 lines total) BEFORE audit;
`blue/CHANGELOG.md` used as navigation only; `red/ledger.md` (round-2 merge state) read for
lineage context. All round-2 textual repairs spot-verified in place: R2-4, R2-13, R2-14,
R2-15, R2-19, R2-20, R2-21, R2-22 read clean at their sites. The findings below are
lens-scoped candidates; stable R3-N ids and lineage rulings belong to the merge.

Overall shape of the round: the round-2 repairs are individually well-executed. What this
lens finds is almost entirely COMPOSITION residue — round-2 mechanisms that do not compose
with the design's own default mode (rung 0), with its abort paths, or with its fresh-dir
minting. Invariant 7 held up well under enumeration; its gaps are at the edges (abort-path
snapshots, harvest-staging failure), not the core.

---

## L5-F1 — the DEFAULT rung's execution shape is undefined and self-contradictory: §1.4 calls rung 0 "same code path" while §3.4 says manual runs "do not pass through the wrapper" — and steps 0/2/4/7 of the code path ARE the wrapper; consequence: the R2-5 origin-cap and the R2-6 corroboration gate are void in the design's default-and-possibly-terminal mode

- lineage candidates (merge to rule): amends closure of R2-5/R2-6 (mechanism holds at
  scheduled rungs only) and R2-16 (the ledger horn was owned; this is the unowned sibling);
  composes with R1-14's rung-0-default posture.
- location: §1.4 — "**rung 0 — manual `/self-improve`, same code path, zero standing cost —
  is the DEFAULT and may be terminal**"; vs §3.4 (R2-16b) — "rung-0 manual runs do not
  pass through the wrapper — manual-session spend is OUT-OF-LEDGER BY DESIGN ... the cheap
  upgrade is invoking manual runs through the same wrapper (the code path already exists)";
  vs §2.2 step 0 — "PREFLIGHT (wrapper, OUTSIDE the model session ...)"; §1.5 — "the
  wrapper stamps an origin marker file (`inputs/.sleeper-origin`, carrying the parent run
  id) into EVERY run dir it creates".
- problem, two legs:
  (a) **Contradiction.** After R1-16/R2-3/R2-4, steps 0 (preflight+harvest staging+canary
  drive), 2 (wrapper-staged scored table), 4 phase B (wrapper-side setup + origin stamp),
  and 7 (capture, ledger, snapshot compare, origin sweep) are all wrapper-hosted. A manual
  `/self-improve` that does not pass through the wrapper is NOT the same code path — as
  printed, the default mode has no executor for four of the eight steps (no docket, no
  staged red-memory mirror, no scored table for step 2 to read). The report specifies the
  scheduled shape in detail and the DEFAULT shape not at all.
  (b) **Guard void in default mode.** The R2-5 mechanism is wrapper-keyed (stamp at
  creation; step-7 porcelain sweep; ledger-window timestamp fallback — all wrapper
  artifacts). At rung 0 there is no wrapper, so a manual run's TWO run dirs (self-improve
  dir + FEOV sub-run dir) land markerless and outside any ledger window. Their
  loop-authored friction is then, at the next harvest: (i) untagged — counted at full
  recurrence, evading the 1-per-class cap; and (ii) "non-sleeper" — SATISFYING the R2-6
  corroboration gate for ordinary classes. The authorship of that friction is identical to
  the scheduled case (model-authored, human present only at invocation); the poisoning
  concern §1.5 argues from does not depend on the scheduler. This is the R2-5 circumvention
  shape recurring one proxy deeper: tag-by-wrapper-presence fails at rung 0 exactly as
  tag-by-dir-name failed for the sub-run — and rung 0 is the mode the report itself made
  DEFAULT and "possibly terminal" (R1-14), so this is normal operation, not an edge case.
  §6 row 10's grade ("the causal-provenance marker closes that hole") holds only for
  scheduled rungs; the row does not say so.
- required_fix: decide the rung-0 execution shape and state it. The report's own R2-16b
  sentence names the cheap resolution: manual runs invoke the SAME wrapper (a `--manual`
  invocation; human present, canary and phase drive optional or kept — argue which), which
  restores ledger, origin stamping, staging, and makes "same code path" true. If instead
  rung 0 stays wrapper-less, then: specify how steps 0/2/4/7 execute manually, extend the
  origin-tag mechanism to cover manual runs (the marker can be stamped by whatever creates
  the dir, with the self-report limitation owned), and re-grade §6 row 10 + §1.5's "made
  true by mechanism" with the rung-0 residual stated.
- grading: certain (textual contradiction; composition present as printed) × medium
  (default-mode mechanics unspecified + the self-poisoning guard and corroboration gate
  void in default mode; bounded by human presence at invocation and the stub-gate human
  triage) × low (the fix is the sentence R2-16b already wrote) → severity **medium-high**

## L5-F2 — the R2-7 watchman compares only at step 7: the abort path takes no END snapshot and "start-vs-prior-run's-END" has no referent after a dead run — "run N+1 cannot fire past unexamined tamper evidence" is proven for COMPLETED runs only, and abort is a model-influenceable terminal state

- lineage candidate: amends closure of R2-7 (the watchman exists; its coverage over the
  abort path was never argued).
- location: §2.2 step 7 — "end snapshot AND COMPARE (R2-7): end-vs-start and
  start-vs-prior-run's-end guardrail hashes — on any mismatch, write a dated TAMPER flag";
  §4.3 layer 5 — "so run N+1 cannot fire past unexamined tamper evidence, and detection
  latency is one run, not one human review cadence"; vs §5.3 — the abort path enumerates
  blackboard writes, nonzero exit, session-id recording, resume — no snapshot, no compare.
- problem: a run that dies before step 7 (deterministic abort, `--max-budget-usd` kill,
  canary abort, wrapper crash, k-th failed resume) produces no END snapshot and runs no
  compare. The next run's specified compare is "start-vs-prior-run's-END" — which does not
  exist; whether the wrapper falls back to prior-START, to the last snapshot of any kind,
  or silently skips the cross-run leg is unstated. So the tamper-evidence chain has an
  unspecified link precisely at abnormal termination — and aborts are a state the session's
  own behavior can induce (turn/budget exhaustion), so "detection latency is one run"
  is unbacked for the write-then-force-abort sequence. Resume attempts compound the
  ambiguity: whether attempt 2 takes a fresh START snapshot (and what attempt 2's
  "prior end" is) is also unstated.
- required_fix: the wrapper snapshots at EVERY exit path it observes (abort included —
  the wrapper outlives the session by construction, so this is one code path), and the
  cross-run compare is defined against the LAST RECORDED snapshot regardless of kind
  (end, abort, or start), fail-closed on a missing chain link per invariant 7. One
  sentence in §2.2 step 7 + §4.3 layer 5.
- grading: low-medium (aborts are routine; the tamper premise additionally needs a layer-4
  writer) × medium-high (the property §4.3 layer 5 and row 4 leg (b) now lean on has a
  hole at exactly the abnormal path) × trivial → severity **medium**

## L5-F3 — the per-cause HALT's dead SIGNATURE is built from "exit class + first abort-record line," and abort records are dated and run-dir-addressed BY DESIGN: unless normalization strips the variable content the design itself re-mints nightly, no two deaths ever match and the HALT never fires

- lineage candidate: amends closure of R2-10 (the mechanism exists; its keying is
  unspecified where it decides everything).
- location: §3.4 — "the wrapper normalizes each death's abort reason (exit class + first
  abort-record line) into a dead SIGNATURE recorded beside the DEAD marker; after M=3
  consecutive fresh-dir deaths with the same signature the wrapper writes a `HALT`
  marker"; same section — the DEAD marker is written "in the OPERATOR-OWNED dir beside
  the ledger, dated" and the run is keyed by date ("research/<date>_self-improve/").
- problem: identity-keyed-detector-lineage-blind, applied to the design's own repair. The
  fresh-dir mechanism GUARANTEES the run-dir path and date differ every night; any abort
  reason whose first line embeds the run dir, the date, a session id, or a nonce (the
  natural content of an abort record — the design's own records are "dated" by
  specification) yields a unique signature per death, M=3-consecutive never matches, and
  the HALT — the entire R2-10 repair — never fires while behaving exactly as specified.
  "Normalizes" is asserted; the normalization is unspecified, and it is the whole
  mechanism. Secondary residual: "3 CONSECUTIVE same-signature" never trips on two
  deterministic causes that alternate (A,B,A,B) — lower likelihood, worth one owning
  clause, not machinery.
- required_fix: specify the normalization (strip dates, paths, ids/nonces, numbers —
  e.g. signature = exit class + abort-reason TEMPLATE/error-code, not its rendered
  string); add the wrapper's zero-HALT-firings telemetry to the doctor line so a
  never-firing detector is visible (the design's own zero-firing lesson); one clause
  owning the alternating-cause residual.
- grading: medium (natural implementations embed variable content; the design's own
  record format invites it) × low-medium (bounded waste resumes: up to ~k×ceiling/night
  until the monthly cap, monthly — the exact R2-10 harm) × trivial → severity
  **low-medium**

## L5-F4 — the wrapper now hosts the gate stack but is absent from both protection enumerations that predate its promotion: layer 3's protected-definitions list and the snapshot's "guardrail file set" name neither sleeper-wrapper.mjs nor harvest.mjs, and the report never states WHICH copy of the wrapper the OS scheduler invokes

- lineage candidate: composes with R2-8's closure (the plugin-copy hash verify is the
  wrapper verifying a copy — self-verification if the wrapper executes FROM that copy;
  unverified working-tree execution if not; the report never says which).
- location: §4.3 layer 3 — "Every file that *defines* layers 1–5 (permission profile,
  plugin hooks, scheduled-task definition, the command markdown of /self-improve and
  /graduate, the guard binary) is outside the write boundary" — a round-0 enumeration;
  §4.3 layer 5 — "SHA-256 hashes of the guardrail file set" — membership never
  enumerated; §0 tree — sleeper-wrapper.mjs lives under `plugins/sleeper-service/scripts/`;
  §3.4 rung 1 — "the launch script (Node, mechanics) owns preflight + budget ledger +
  idempotence + JSON-result logging" — invocation path unstated.
- problem: round 1 promoted the wrapper to a named artifact "because five gate-side
  controls now live in it"; round 2 added four more (canary drive, hash verify, watchman,
  HALT). Neither protection enumeration was extended: layer 3's list still reflects the
  round-0 design (no wrapper, no harvest.mjs), and the snapshot's guardrail set is
  undefined, so whether the wrapper's own hash is tamper-evidenced is unknowable from the
  text. The execution-provenance question is the sharper half: if the scheduled task
  invokes the repo working-tree wrapper, the design's most trusted component is the one
  copy OUTSIDE the R2-8 hash discipline (Edit-denied to the session, yes — but the
  layer-4 subprocess residual is exactly the channel row 4 risk-accepts, and the wrapper
  is its highest-value target); if it invokes the operator-owned copy, R2-8's check is
  self-verification and the integrity root is the scheduled-task definition + OS file
  permissions — defensible, but it must be SAID. The lead's round-2 watch ("the wrapper
  is itself entirely trusted") got invariant 7's outcome-record answer; the wrapper's own
  integrity/provenance was the other half and it is still unstated.
- required_fix: three sentences — (i) the scheduled task invokes the wrapper by absolute
  path from the operator-owned copy, and that task definition is already in layer 3's
  protected set (close the loop explicitly); (ii) add sleeper-wrapper.mjs + harvest.mjs +
  the sleeper-mcp/permissions JSONs to layer 3's enumeration and define the snapshot
  guardrail set to include them; (iii) own the self-verification limit: the wrapper cannot
  attest itself; its integrity root is the operator-owned task definition and OS
  permissions, outside the model's reach.
- grading: certain (textual: enumerations demonstrably predate the artifact) × low-medium
  (the copies are Edit-denied regardless; the exposure is the risk-accepted layer-4
  channel plus honesty of the protection claims) × low → severity **low-medium**

## L5-F5 — round 2 minted artifacts outside §0's now-"total" enumeration: the SessionStart staleness hook, the doctor dead-man/TAMPER/staleness lines (a change to a DIFFERENT plugin), and the two operator-owned config files the build must author

- lineage candidate: same class as R2-19, recurring one round later against the round-2
  additions (R2-19's totality fix was verified against the round-1 tree).
- location: §0 — "exactly THREE new code artifacts ... plus two command prompts, a
  scheduling doc, **the continuous-learning skill file, and the plugin manifest** ...
  everything else reuses shipped machinery"; vs §3.4 (R2-9) — "the same staleness check
  runs as a SessionStart-hook warning line in every INTERACTIVE session"; §3.4/§2.2
  step 7 — doctor gains "last successful sleeper run" / TAMPER / "sleeper plugin copy"
  lines; §3.3 — `--mcp-config <sleeper-mcp.json>`; §4.2 — `sleeper-permissions.json`.
- problem: the SessionStart hook is a new executable artifact plus a hooks.json
  registration — none of the enumerated categories; the doctor lines modify
  prosthetic-conscience — "reuses shipped machinery" is false for a change that EDITS
  shipped machinery in another plugin (also a cross-plugin dependency §0's manifest line
  does not carry); the two operator-owned JSON configs are build-authored artifacts with
  no home in the enumeration. Exhaustive-sweep-omits-own-specimen, second instance.
  Secondary sub-defect, same R2-9 site: the SessionStart warning's ENABLEMENT condition is
  unstated — an operator who schedules then deliberately stops leaves `last-successful-run`
  frozen and the warning fires in every interactive session forever; by the report's own
  Dependabot/habituation evidence a channel that nags spuriously stops being read, which
  is the channel's whole value. One clause (warning conditioned on scheduling-enabled;
  clearing HALT/disable resets it) closes it.
- required_fix: extend §0's enumeration (SessionStart hook, doctor-line changes to
  prosthetic-conscience named as a cross-plugin delta, the two operator-owned configs);
  one clause on the warning's enablement/disable semantics.
- grading: certain × low (enumeration honesty + one unowned nag path) × trivial →
  severity **low**

## L5-F6 — R2-6's containment overstates the doctor channel: "every infrastructure class in the bypass list ALSO surfaces independently on the doctor/dead-man line," but the line prints only the LAST skip/abort reason — non-persistent infrastructure events older than the most recent are on no channel but the stub flag

- lineage candidate: amends closure of R2-6 (the decision stands; one supporting sentence
  overclaims).
- location: §1.5 — "every infrastructure class in the bypass list ALSO surfaces
  independently on the doctor/dead-man line (§3.4), so the flag is not the sole channel";
  vs §3.4 — doctor "reports 'last successful sleeper run: N days ago (last skip reason:
  <reason>)'".
- problem: TAMPER and HALT persist until human-cleared, so for those two the sentence
  holds. Canary aborts, ledger-unparse skips, hook crashes, and DEAD markers older than
  the most recent event scroll off a line that carries exactly one reason — for a class
  that fired last Tuesday and was followed by any other skip, the `sleeper-only` stub
  flag IS the sole channel, at precisely the Dependabot-base-rate reliance the R2-6
  decision's cost-ownership tried to bound. The severity-gated bypass decision itself is
  right; this sentence's "every ... ALSO surfaces" is doing containment work it has not
  earned.
- required_fix: either weaken the sentence to the classes it is true of (TAMPER/HALT
  persist; others surface only-as-most-recent), or make it true — the dead-man line
  prints the last-N reasons or a per-signature count since last human clear (the R2-10
  signature machinery already exists to key it).
- grading: certain × low (one overclaiming clause inside an otherwise owned residual) ×
  trivial → severity **low**

## L5-F7 — the two-phase stream-json drive's load-bearing BEHAVIORAL assumptions ride on flag existence only, and two invariant-7 edge steps have no stated failure outcome

- location: §2.2 step 0 — "THEN the wrapper OPENS the model session as a PHASE-DRIVEN
  stream-json drive (`--input-format stream-json --output-format stream-json`, flag pair
  verified in `claude --help` on the pinned CLI 2.1.212 — R2-1)"; same step — "qmd daemon
  /health (degrade note if down)"; "harvest.mjs staging (docket + red-memory mirror)".
- problem, three small legs of one class (specified mechanism, unprobed/unstated
  outcome):
  (a) R2-1's probe verified the FLAGS exist; the drive additionally assumes (i) a second
  user message can be injected after parsing phase-0 events within one session, (ii) the
  deny outcome is observable in-stream at phase-0 time (not only in the terminal
  envelope), (iii) the model reliably ATTEMPTS the canary edit on instruction. All three
  fail CLOSED (abort a healthy night — a reliability cost, not a safety hole; polarity is
  right and worth saying), but none is a named build/Phase-4 acceptance probe; OQ21
  covers deny PROVENANCE only, not deny visibility/timing.
  (b) harvest.mjs failure at step 0 has no stated outcome — abort, skip-with-reason, or
  proceed docket-less? Invariant 7's own test ("a control that records or attempts
  without verifying-and-surfacing its own outcome is telemetry") applied to the staging
  step the invariant's author skipped.
  (c) the qmd "degrade note" has no stated destination or reader (run-dir? ledger?
  doctor?) — a two-word fix under invariant 7.
- required_fix: add the two-phase drive (three behavioral legs) to the Phase-4
  acceptance-test list beside OQ1/OQ2; one sentence each for harvest-failure outcome
  (fail-closed skip with reason, per the ledger-unparse precedent) and the degrade note's
  destination.
- grading: medium (unprobed behavior on the pinned CLI; benign polarity) × low (worst
  case is spuriously aborted nights + an unspecified degrade path) × trivial → severity
  **low**

---

## Checks run that produced NO finding (for the merge's negative-space record)

- Invariant 7 enumeration sweep over all wrapper gates: canary/fired-record, snapshot
  watchman, hash verify, ledger preflight, HALT/dead-man all carry checked+surfaced
  records as claimed; the only residues are F2 (abort-path snapshot) and F7(b)/(c).
- R2-18 arithmetic recomputed: 30×$0.10=$3, 30×$0.50=$15, $50/$15≈3.3× — "≥3× headroom"
  holds; ceiling-vs-cap composition now correctly owned as anomaly signal.
- R2-2's hook-before-permission ordering is consistent with the quoted doc physics
  ("a blocking hook ... stops the tool call before permission rules are evaluated") —
  the fired-record mechanism is coherent as specified; fired-record unforgeability holds
  under the profile (operator-dir Edit-denied absolute).
- R2-5's marker erasability (marker sits INSIDE the loop-writable run dir) is adequately
  belted by the step-7 porcelain sweep + ledger-window timestamp fallback, both outside
  the session's reach — for SCHEDULED runs (the rung-0 hole is F1, not a marker defect).
- §0 artifact enumeration vs printed tree: total after R2-19 (the round-2 additions
  outside the tree are F5).
- claim_count 142→146 delta re-derived against the stated method: +invariant 7, +§7
  round-2 bullet, +OQ20, +OQ21 — consistent.
- OQ list: 21 items, OQ8 marked resolved inline — template-compliant.
- R2-21 honesty clause, R2-22 double-marker, R2-13/14/15/16/17/20 — read in place, clean.

## Envelope

```
verdict: FAIL (open findings above; F1 medium-high)
findings: [L5-F1 medium-high, L5-F2 medium, L5-F3 low-medium, L5-F4 low-medium,
           L5-F5 low, L5-F6 low, L5-F7 low]
lineage_candidates: F1→amends R2-5/R2-6/R2-16 closures; F2→amends R2-7; F3→amends R2-10;
           F4→composes with R2-8; F5→recurrence of R2-19 class; F6→amends R2-6
           (merge assigns supersedes/closure classes — lens proposes only)
archive_reads_needed_at_merge: R2-5, R2-6, R2-7, R2-10, R2-16, R2-19 closure records
           (near-match rule) before minting fresh ids
friction: none — template fit clean; no capability gaps impeded this pass
```
