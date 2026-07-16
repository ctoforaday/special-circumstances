# Efficiency phase — the ratified cost levers, built without cheapening judgment

Source of authority: run 4, `research/2026-07-14_efficiency-investigation/report.md` (UNVERIFIED,
4 rounds, 10 residual gaps — none above MEDIUM). This plan implements what that debate RATIFIED,
defers what it HELD, and does not resurrect what it REJECTED. Cross-inputs: run 4's 39-entry
friction harvest, `cost.md` (run 3: $149.95; run 4: $414.97 measured), run-3 report §3,
`ideas/backlog.md`.

## I. Summary & Goals

**Problem.** Debate cost is dominated by (a) redundant re-reading of red's own closed cases at
the merge and judge seats (the archive compounds every round), (b) turn-fragmented candidate
ingestion, (c) a carried-gap re-docket loop in shipped code, and (d) rounds that continue past
the point of diminishing returns because the stop decision has no signal. Run 4 also measured
what the levers must never touch: distinct-lens coverage, red-merge depth, judge strength, the
full re-read of blue's report, and the spot-check floor.

**Success criteria (quantitative where the report gives numbers):**

1. Every §6.1 ranked item (1–6) is either implemented here or carries a named revisit trigger;
   the three REJECTED mechanisms (severity-floor auto-stop, mass-actuated throttle, best-of-N
   grading, collator seat) appear ONLY as telemetry/insurance groundwork, never as actuation.
2. Run 5's `cost.md` shows: merge-seat candidate ingestion collapsed to ≤2 read turns/round
   (batching, measured ≈$2.2/run at run-3 scale); the default merge read no longer includes the
   closed archive (sharding-addressable ≈$2–4/run modeled + the unpriced judge-read benefit).
3. `trajectories/board-telemetry.jsonl` exists in run 5 with one line per round, mapping-version
   stamped, reconcilable against `debate.md` round count (presence check), and sampled-recomputable
   from the git-tracked findings record.
4. The carried-gap re-docket loop is closed: a `carried` ruling persists and re-dispatches only
   on red grade change or a lineage successor (marginal docket growth eliminated; gate-erosion
   path closed).
5. Simulator: all existing tests green + new regression tests for every engine change (target
   ≥10 new); zero-token.
6. Doctrine check passes: every change lands on instance-redundancy, residency of red's own
   closed cases, or mechanical collation (§6.3's three safe categories). Nothing reduces judge
   strength, red-merge depth, lens coverage, blue full-read, or the spot-check floor.
7. The write-guard preflight is executed from a live subagent seat and its result cited in the
   sharding PR (run 4 closed without the closing-act probe, so per §4.5 condition 6 the
   verify-seat-independence branch is PROMOTED to required).

**Explicitly out of scope (with owners):** round-scoped audit (HOLD; run-5 gate has three
mechanical conditions evaluated against run 4's propagation record at run-5 planning time —
§5.5); any mass/severity actuation (revisit trigger: runs 4–5 telemetry showing mass predicts
next-round value); best-of-N grading (trigger: surviving-bias instance in `grade_disputes`
records + a cross-family grader being reachable); bulk-tier model choice (human dial, already
documented — run 4's $415 vs run 3's $150 is ~all keeper freight, dwarfing every lever here;
stated for honesty, not actioned).

## II. Technical Context

- **Engine:** `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js`
  (workflow script — no filesystem access; prompts are the only lever on seat behavior; envelope
  schemas are the only structural enforcement). Simulator:
  `plugins/frank-exchange-of-views/tests/simulator/` (Node built-in test runner, stub harness).
- **Seats:** agent definitions under `plugins/frank-exchange-of-views/agents/` (tool grants,
  preloaded skills); prompts assembled in debate.js.
- **Hooks/toolchain:** Go, `plugins/prosthetic-conscience/tools/` (tested; CI-built release
  assets; version-gated plugin cache).
- **Constraints:** harness write-guard is filename-keyed, key set unenumerated, seat-class
  dependence UNVERIFIED (run-4 §4.5 cond 6); harness `log()` is console-ephemeral (verified —
  journal.jsonl carries lifecycle only); Read tool caps ~25k tokens/call; workflow agents reach
  MCP tools only via ToolSearch, which the FEOV seats' tool grants currently omit; attestation
  ceiling (§6.2): in-run enforcement reaches shape and consistency, never vacuity — vacuity's
  auditor is post-hoc over git-tracked artifacts.
- **Versioning:** FEOV 0.5.0 → 0.6.0 (content), PC 0.7.0 → 0.8.0 (hook + doctor changes, tag
  builds assets). Update dance required after merge; never mid-run.

## III. Proposed Changes (the spec)

Three PRs, ordered so each is independently shippable and simulator-verifiable.

### PR-A — engine: telemetry, disputes, carried rulings, sharding, batching (FEOV 0.6.0)

```
plugins/frank-exchange-of-views/
├── skills/research-protocol/scripts/debate.js        [MODIFY] — everything below
├── skills/research-protocol/SKILL.md                 [MODIFY] — blackboard tree gains red/ledger.md,
│                                                        red/archive.md, trajectories/board-telemetry.jsonl
├── agents/red-auditor.md                             [MODIFY] — archive demanded-read MUST; found_by duty
├── agents/lead-judge.md                              [MODIFY] — ancestor demanded-read MUST (rationale
│                                                        names archive records read)
└── tests/simulator/debate.test.mjs                   [MODIFY] — new regression tests (see V)
```

1. **[MODIFY] Board telemetry (§1.5 item 1, §2.5 items 1–2 — RATIFIED 3/3).** Red-merge prompt:
   append one JSON line per round to `trajectories/board-telemetry.jsonl` (neutral name;
   append-only `cat` path): `{round, open_count, max_severity, new_mint: {count, profile},
   mass, accepted_dispute_deltas, realized_open, excluded_mass_memo, found_by_summary,
   mapping_version}`. Mass mapping: pinned TOTAL over the 8-member GRADE enum before the first
   logged round (`realized` → 0 in mass but counted in board profile; `trivial` assigned; §8 Q6
   as decided); mapping version stamped per line; changed mapping ⇒ new series. Envelope:
   `RED_ENVELOPE` gains `found_by: ['L1',...]` per gap (auditable against preserved lens
   candidates; actuation reviews re-derive independently at a non-red seat — recorded as prompt
   text in both agent files). Consumers named: retrospective/next-run docket; cost-audit.mjs
   via PR-B item 3.
2. **[MODIFY] Grade-dispute channel, minimal form (§3.3/§3.5 — RATIFIED).** `BLUE_ENVELOPE`
   gains `grade_disputes: [{gap_id, dimension, proposed, evidence}]`; `RED_ENVELOPE` gains
   `dispute_responses` (red's per-dispute answer: accepted-with-delta / rejected-with-reason).
   Clauses (i)–(vii) from §3.3 verbatim into prompts + schema. Routing per the ratified design:
   an **explicitly REJECTED** dispute is held one round and enters the contested docket only if
   blue re-disputes; an **UNADDRESSED** dispute (`dispute_responses` silently unset for it)
   auto-dockets, treated as rejected — default-to-docket punishes silence, not disagreement.
   Accepted branch (clauses (v)–(vii)): accepted disputes carry a delta trail (old grade, new
   grade, round) into the telemetry line and findings; terminal-round disputes excluded from
   `carried` — they still auto-docket at exit, never silently dropped (clause (vi)). Judge
   resolution enum gains the dispute-resolution value.
3. **[MODIFY] Carried-ruling persistence (§6.4 item 6 — engine defect, graded MEDIUM).**
   `carried` rulings enter the adjudicated set; a carried gap re-dockets ONLY when (a) red's
   grade for it changed in `redEnv` (script-visible trigger), or (b) a successor names it in
   `supersedes` (existing lineage path). Kills the repeat-dispatch component and the
   carried→risk_accepted gate-erosion path. Judge enum/ordering fix rides here (run-4 friction,
   judge-r2): regression-successor traffic acknowledged — successors reach the judge only after
   blue has spoken to them once, OR the enum gains a first-raise class; pick at implementation
   after re-reading judge-r2's friction entry, record choice in the PR body.
4. **[MODIFY] Sharded findings (§4 — RATIFIED with seven conditions, all implemented as
   stated).** `red/findings.md` splits: `red/ledger.md` (authoritative: open gaps + compact
   closure index — one line per closed gap: id | class | summary | supersedes) and
   `red/archive.md` (immutable closed prose). Red-merge Writes BOTH on the first sharded run
   (not the skeleton — creation doubles as the live-seat guard probe); debate.js prompt refs at
   red-merge and judge seats updated (currently hardcode `red/findings.md`). Demanded-read
   MUSTs: red on any lineage/closure claim; judge on every chain ruling (names ancestors read
   in rationale — logged in `### LEAD`, zero schema change). **Near-match trigger (cond 3):** a
   closure-index near-match to a candidate gap forces a targeted archive read BEFORE a fresh id
   is minted — the index screens, it never decides. Reopen triggers (supersedes cite / blue
   cite / spot-check sample) **plus drift-trigger inheritance (cond 4): archived closures whose
   evidence cites volatile living sources inherit the ledger's drift/time re-check triggers.**
   Archive spot-check floor ≥1/round, never zero, rides `RED_ENVELOPE.archive_spot_checks`
   (required non-empty from round 2 — script shape-check; vacuity audited post-hoc per the
   attestation ceiling). **Count observables (cond 7):** the closure-index line count and
   archive block count ride `RED_ENVELOPE` as merge-reported integers; the script arithmetically
   compares them (catches self-inconsistent self-report only — tier stated; true counts audited
   post-hoc, or by an `sc-recall-index`-class hook if built). **Precondition (promoted, §4.5
   cond 6):** before this PR merges, a live subagent seat (red-merge class) test-Writes
   `red/ledger.md` and `red/archive.md` in the production harness; results cited in the PR
   body. If blocked: fall back to append-only creation via the skeleton + `cat` path — noting
   explicitly in the PR body that this fallback reintroduces the two-creator shape cond 6
   reconciled away, accepted only as a measured fallback with the guard's seat-(in)dependence
   recorded.
5. **[MODIFY] Read batching (§4.6 — collator REJECTED; degenerate form RATIFIED).** One
   red-merge prompt sentence: first action, `cat` all lens candidate files to an ABSOLUTE path
   in the seat's session scratchpad (outside `red/candidates/`, outside the recall index's
   markdown watch surface), then read the single file. Measured saving ≈$2.2/run; zero new
   seats.
6. **[MODIFY] Stop-and-resume as standing termination practice (§1.4/§1.5 item 2 — RATIFIED).**
   One paragraph in the debate.js header + research-protocol: the demonstrated ~$0 terminator
   is operator stop + cache-safe `maxRounds` resume; the telemetry line is the signal the
   operator reads. (The rejected severity floor's carried minority variant is NOT built; its
   arm conditions live in the report should telemetry ever show it firing.)

### PR-B — instruments & record hygiene (FEOV 0.6.0, same tag)

```
plugins/frank-exchange-of-views/
├── skills/research-protocol/scripts/cost-audit.mjs   [MODIFY] — fixes + extensions
├── skills/research-protocol/scripts/batch-collapse.mjs [NEW] — from run-4's committed instrument
├── skills/research-protocol/SKILL.md                 [MODIFY] — MUST-try observable; footnote
│                                                        namespaces; harness-contract reference
├── commands/research.md                              [MODIFY] — claim_count echo; telemetry capture
└── ideas/backlog.md                                  [MODIFY] — corrections (severity-floor claim,
                                                         stale 54KB figure)
```

1. **[MODIFY] cost-audit.mjs corrections (§6.4 item 1):** finding 2 → "merge cost tracks the
   cumulative archive" (its own table contradicts DISPUTE-size); finding 3 → 5× cache-write
   multiplier (12.5 is the absolute rate).
2. **[NEW] Standard instruments:** run-4's bespoke transcript parsers
   (`batch-collapse.mjs`, the merge-decomposition method) land as tested scripts — per-agent /
   per-source context timeline becomes standard cost-audit output (backlog 28(d); run-4
   friction: "run 4 should get this as standard output, not seat improvisation").
3. **[MODIFY] cost-audit.mjs telemetry join:** one added read of
   `<runDir>/trajectories/board-telemetry.jsonl`, joining per-round mass/board columns into the
   cost table (the named-consumer condition from §2.5 item 1).
4. **[MODIFY] MUST-try observable (run-4 friction, blue-respond-r1):** research-protocol PDF
   clause gains: every graded-down citation records an attempt-or-impossibility line (which
   tool was tried / why untriable). Kills the false-paywall class at round 0.
5. **[MODIFY] Footnote namespace convention (blue-synthesize friction):** lane dispatch prompt
   assigns lane-prefixed footnote labels; citation-slice dispatch names footnote-block
   ownership (lens-4 friction).
6. **[MODIFY] Harness-contract reference (three seats rediscovered it):** one referenceable
   paragraph in research-protocol: `log()` is console-ephemeral; `journal.jsonl` is the
   harness's lifecycle record (started/result only); transcripts are `agent-*.jsonl`. Plus the
   two recurring harness notes for lens prompts: Grep count mode counts lines (anchor your
   pattern), quoted-heredoc backslash hazard (prefer Write for scripts).
7. **[MODIFY] claim_count echo (red-merge-r2 friction):** blue's CHANGELOG line carries
   claim_count per round (tracked artifact; envelope-only today).
8. **[MODIFY] Pinned-artifact corrections (§6.4 items 2, 5):** backlog severity-floor savings
   claim corrected against the audited board; stale "54KB" figures replaced with the measured
   sizes.

### PR-C — seat capability & platform (PC 0.8.0 + FEOV agents)

```
plugins/frank-exchange-of-views/agents/
├── blue-researcher.md                                [MODIFY] — tools: + ToolSearch
├── red-auditor.md                                    [MODIFY] — tools: + ToolSearch
└── lead-judge.md                                     [MODIFY] — tools: + ToolSearch
plugins/prosthetic-conscience/
├── hooks/hooks.json                                  [MODIFY] — bootstrap guard wrapper (backlog)
├── tools/cmd/sc-doctor/                              [MODIFY] — cross-plugin requirements aggregation
└── .claude-plugin/plugin.json                        [MODIFY] — 0.8.0
plugins/frank-exchange-of-views/
├── .claude-plugin/plugin.json                        [MODIFY] — 0.6.0
├── requirements.json                                 [NEW] — FEOV's own deps (node; doctor aggregates)
└── commands/research.md                              [MODIFY] — red gap-pattern memory mirrored
                                                         into run inputs/ at pre-create
```

1. **[MODIFY] ToolSearch in seat tool grants** — the run-4 headliner (4 seat classes reported
   MCPs unreachable): seats can reach qmd/pdf-reader/arxiv-latex when the launching session has
   them; the MUST-try clause becomes fulfillable where it applies, and research-protocol's
   recall mode 1 becomes real at seats. Grant sized to duty (capability doctrine): ToolSearch
   only, servers still consent-gated at the session.
2. **[MODIFY] Red gap-pattern memory provisioning (4 seat classes couldn't read it):**
   `/research` pre-create step mirrors the red-auditor project-memory file into
   `<runDir>/inputs/red-gap-patterns.md` (read path that exists at every seat; source of truth
   stays the agent memory).
3. **[MODIFY] Hook bootstrap guard (backlogged crash-storm fix):** each hooks.json command
   wrapped: binary exists → exec; else one terse stderr line pointing at doctor --fix, exit 0.
   CI step runs every hooks.json command against an empty CLAUDE_PLUGIN_ROOT and asserts exit 0.
4. **[MODIFY] Doctor cross-plugin aggregation (backlog):** each plugin ships `requirements.json`;
   sc-doctor walks every installed SC plugin cache dir and aggregates. Go + tests.
5. **[MODIFY] Windowed-read discipline (25k Read cap friction at every audit seat):** documented
   as the sanctioned pattern in red-auditor.md (two-window read with stated boundaries satisfies
   the full-re-read MUST; no confidence discount for windowing alone). Harness feature requests
   are out of our hands; policy kills the friction class.

## IV. Risk & Mitigation

| # | Risk | L×I×C | Mitigation (linked step) |
|---|---|---|---|
| 1 | Write guard blocks `red/ledger.md`/`red/archive.md` at the live seat (guard key set unenumerated; seat-dependence unverified) | M × H × L | PR-A.4 precondition: live-seat probe BEFORE merge; fallback path named (skeleton + append). Probe result cited in PR body either way |
| 2 | Sharding drops ancestor context at judge ⇒ carried→risk_accepted erosion worsens | L × H × L | PR-A.4 demanded-read MUST at the judge (rationale names archive records); §4.5 cond 2 extended-to-judge clause implemented verbatim |
| 3 | Telemetry line becomes unaudited actuation evidence (attestation ceiling) | M × M × L | Presence check vs debate.md; recompute-on-actuation clause + window-reconciliation rule in prompts (PR-A.1); telemetry named "convenience copy, never evidence of record" |
| 4 | Grade-dispute accepted branch goes dark ⇒ blue deflation lever | L × H × L | Clauses (v)–(vii) are part of PR-A.2's schema/prompts, not optional; interlock recorded in SKILL.md |
| 5 | Carried-persistence suppresses a ruling that SHOULD be revisited | L × M × L | Re-dispatch triggers kept: red grade change (script-visible) + lineage successor; simulator test for both triggers (V) |
| 6 | ToolSearch grant widens seat attack surface | L × L × L | Grant is discovery-only; servers remain session-consent-gated; secrets-gate hook still fires PreToolUse on fetches |
| 7 | Batching concatenation pollutes candidates/ or the recall index | L × M × L | Absolute scratchpad path clause (PR-A.5, verbatim from §4.6's corrected sentence) |
| 8 | Version-gated cache skew after merge (empty-bin window) | M × M × L | PR-C.3 guard ships in the SAME tag that bumps versions; update-dance ordering documented |
| 9 | Simulator can't see prompt-level MUSTs (they're text, not code) | M × M × M | Envelope-level observables get schema checks (archive_spot_checks, grade_disputes, found_by) — simulator-testable; prompt MUSTs get the named post-hoc auditor per the attestation ceiling; no condition claims more than its tier |

## V. Verification Plan

Automated (exact commands, run per PR):

```
node --test plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs plugins/frank-exchange-of-views/tests/simulator/cost-audit.test.mjs
# PR-A adds tests: telemetry line emitted per round + mapping version; grade_disputes routing
#   (explicit reject held one round → dockets only on re-dispute; unaddressed dispute_responses
#   auto-dockets; accepted-delta trail present); carried ruling persists / re-dispatches on grade
#   change / on successor; sharded prompt refs (no red/findings.md string in red-merge or judge
#   prompts); near-match trigger text present in red-merge prompt; archive_spot_checks
#   required-non-empty from round 2; closure-index/archive count integers arithmetic-compared by
#   the script; batching sentence names absolute path; lineage throw still fires.
cd plugins/prosthetic-conscience/tools && gofmt -l . && go vet ./... && go test ./...   # PR-C
node .github/validate-json.mjs
```

Manual / live:

1. **Guard probe (PR-A precondition):** spawn a live red-merge-class subagent that Writes
   `red/ledger.md` + `red/archive.md` in a scratch run dir; record allowed/blocked in the PR.
2. **`--smoke` run** after PR-A+B merge and update dance: verify board-telemetry.jsonl (1 line),
   ledger/archive created by the merge seat, batching concat in scratchpad, cost.md gains the
   telemetry join columns.
3. **Doctor** after PR-C tag: aggregated table lists FEOV's requirements; empty-bin simulation
   (rename bin/, run any Edit) produces one warning line, no crash storm.
4. **Run 5 (the real gate):** keeper run; success criteria 2–4 measured from its cost.md and
   run record. Round-scoped audit's run-5 decision evaluated at plan time against run 4's
   propagation record (§5.5's three conditions — separate decision, documented with this plan
   as its pointer).

Auditor gate: `/plan-audit plans/efficiency-phase.md` must return PASS before implementation
begins.

## Appendix — write-guard preflight record (PR-A.4 precondition: SATISFIED)

Executed 2026-07-16 from a live red-auditor-class subagent in the production harness
(§4.5 cond 6's promoted branch), four real Write calls in a scratch run layout:

| Write | Result |
|---|---|
| `red/ledger.md` | ALLOWED (file created) |
| `red/archive.md` | ALLOWED (file created) |
| `red/findings.md` (known-blocked control) | BLOCKED — "Subagents should return findings as text, not write report files." |
| `report.md` (known-blocked control) | BLOCKED — identical error text |

Measurement: the guard keys on the filename token set, not path/directory/content (controls
blocked even in scratchpad — replicating run 3's experiment at this seat class); the guard was
demonstrably LIVE at the measuring seat (controls fired), so the shard-name pass is not
vacuous. The proposed names are clean; red-merge creates both files on the first sharded run
as specified; the skeleton+`cat` fallback is not needed.
