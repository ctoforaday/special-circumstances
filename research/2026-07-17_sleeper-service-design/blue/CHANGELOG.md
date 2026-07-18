# blue CHANGELOG — How should sleeper-service, the autonomous learning loop (Phase 4), be designed?

## Round 0 — synthesis (2026-07-17)

Synthesized `blue/report.md` by UNION of the three method-lens lane candidates
(`blue/candidates/lane-1.md` adversarial-disconfirming-first, `lane-2.md`
primary-literature, `lane-3.md` local-repo critical-stance / live-probe). Structural merge
(append + dedup), not a free-form rewrite; no substantive lane content dropped.

Concrete edits:

- **Structure:** merged the three lane outlines into one H1–H5 report: §0 design summary +
  six loop invariants; §1 consumption (disconfirming pass, survival argument, input
  inventory, 3-stage pipeline, self-poisoning guard); §2 /self-improve + /graduate + stub
  contract; §3 headless (live probes, doc facts, traps, 5-rung scheduling ladder,
  idempotence); §4 consent gates (evidence line, permission profile, 7-layer gate stack
  with route-around answers); §5 cost discipline (ceilings, tiering, honest partials); §6
  merged risk matrix (12 rows + 5 rejected-scope items); §7 pre-flight self-audit; §8 open
  questions (15, union of all lanes plus one synthesis-minted).
- **Deduplication:** convergent claims (artifact-mining architecture, harvest/rank/pick-ONE
  pipeline, smoke-scale bounded research, stub contract, human-only /graduate, `claude -p`
  flag surface, `--bare` trap, dontAsk allowlist inversion, deny-set, PreToolUse fence,
  ledger preflight, $414.97 anchor, honest-partial abort pattern) collapsed to single
  statements.
- **Minority provenance manifest:** every claim appearing in exactly one lane draft is
  tagged `[minority: lane-N/<lens>]` inline — 62 minority markers total (20 lane-1, 15
  lane-2, 27 lane-3). Notable minority reports red should weigh: lane-3's live headless
  probes P1/P2 (the only run-verified headless evidence); lane-1's permission-enforcement
  issue trio (#22055 closed-not-planned, #6631, #25621) and the
  `disable-model-invocation`/background-wait-ceiling mechanisms; lane-2's DGM/SICA/STOP
  self-improvement corpus, STOP circumvention percentages, Dependabot fatigue evidence,
  and permission-doc drafting details (Edit-not-Write rules, Read-deny-blocks-Edit).
- **Cross-lane conflict carried openly** (§2.2 step 6 + open question 14): lane-2's
  "append one line to ideas/backlog.md" step vs lane-3's self-poisoning guard ("NEVER
  edits ideas/backlog.md"). Both preserved; synthesis proposes a loop-owned generated
  index (`ideas/stubs-index.md`) as reconciliation; ruling deferred to red/lead.
- **Footnotes merged:** 50 labels in a unified namespace; same-source citations across
  lanes merged to one label with all citing lanes noted (e.g. `[^HeadlessDocs]` lanes 1–3;
  `[^Reflexion]` lanes 1–3; `[^AIScientist]` lanes 1+3; `[^Voyager]` lanes 2+3;
  `[^SlashHeadlessIssues]` lanes 1+2). Lane-local labels (`[^L1...]`/`[^L2...]`/`[^L3...]`)
  retired; lane drafts preserved untouched in `blue/candidates/`.
- **claim_count = 132** (tracked copy of the envelope figure). Counting method,
  re-derivable: top-level bullets, numbered items, and table body rows outside code blocks
  and footnotes (121 matched lines − 5 table header rows = 116) plus code-block contract
  units (/self-improve steps 0–7 and the 8 idea-stub contract fields = 16). Prose-embedded
  sub-claims inside continuous paragraphs are not separately counted; the figure is a floor,
  not a ceiling.
- Pre-flight self-audit run against `inputs/red-gap-patterns.md` (read before merging);
  audit notes in report §7.

## Round 1 — revision (2026-07-17)

All 30 red gaps addressed additively; no substance dropped; zero grade disputes filed
(every required fix was trivial-to-low complexity, so absorption beat contestation on the
pragmatist's own arithmetic). Pre-flight: `inputs/red-gap-patterns.md` re-read; the
incomplete-repair/footnote-lag and repair-regression patterns drove the propagation greps
below. `debate.md` ### RED read in full; no LEAD section exists yet; no pending dispute
deltas.

**Design changes (the fixes that alter the build, not just the text):**

- R1-16 (high): harvest.mjs moved WRAPPER-SIDE, outside the model session; the in-session
  Bash allowlist is now pinned-argv read-only git commands only (`Bash(node scripts/*)`
  deleted — it was write-capable AND pattern-mismatched the printed layout). Edited: §0
  tree + artifact count (TWO→THREE, wrapper promoted to named artifact), §1.3
  red-memory provisioning, §2.2 steps 0/1/7, §4.2 profile + prose, §4.3 layer 4, §5.3,
  §6 row 4, OQ18 added.
- R1-25 (high): §1.5 extended to cover AUTHORSHIP (origin: sleeper tagging, recurrence
  contribution capped at 1, non-sleeper corroboration or sleeper-only flag before pick);
  §4.3 semantic-consent paragraph carries the cap; §6 row 10 re-graded Low→Low-Medium
  with the false-premise admission.
- R1-27 (high): per-rung gate-survival table added to §3.4 (6 controls × 5 rungs);
  rung-3/4 adoption declared graduation-grade; §6 row 16 added; OQ19 added.
- R1-15 (medium-high): `--plugin-dir` pinned to operator-owned read-only copy
  (`~/.claude/sleeper/plugins/`), never the working tree — §3.2 recipe, §4.3 layer 2,
  §4.2 allow rule for the copy.
- R1-19 (medium-high): ledger relocated `ideas/.sleeper-ledger.jsonl` →
  `~/.claude/sleeper/ledger.jsonl` (wrapper-written, fail-closed on unparse); invariant-4
  coverage named (the wrapper). Edited: §0 invariant 5, §2.2 steps 0/7, §5.1, §5.3,
  §6 row 12.
- R1-28 (medium-high): step-0 denial canary (attempt one out-of-fence write, abort unless
  DENIED) — marker loss now fails CLOSED. Edited: §0 tree, §2.2 step 0, §4.3 layer 2,
  OQ2 reframed (verify enforcement, not presence), OQ16 added (polarity inversion).
- R1-26 (medium): wrapper start/end snapshots (porcelain + guardrail hashes to
  operator-owned log); auto-commit-to-sleeper/-branch named as heavier alternative —
  §4.3 layers 4/5, §2.2 steps 0/7, §6 row 4.
- R1-29 (medium): resume cap k=3 + dead-mark + doctor dead-man line — §0 invariant 5,
  §2.2 step 0, §3.4 idempotence, §5.1, §6 row 15 added.
- R1-17/R1-18 (medium): Read/Grep/Glob allow-scoped to repo + plugin copy; belt
  Read-denies on `~/.claude` settings/credentials/transcripts; WebSearch dropped from the
  nightly default (kept for human-present graduation); stub contract gains
  `origin-labels` (untrusted-origin for web-derived); §6 rows 13/14 added with
  risk-accept rationale for the residual egress.
- R1-22 (low-medium): stub aging mechanism specified — `status` field in §2.3, 30-day
  auto-stale from dated filename, `stub-stale` docket re-entry; §1.4 skip rule rewritten;
  §6 row 3 updated.
- R1-14 (medium): null-alternative paragraph added to §1.4 (rung 0 default and possibly
  terminal; daily cadence is a hypothesis the stub-survival ledger tests; triage-rate
  revisit trigger named).
- R1-21 (medium): §6 row 4 acceptance re-argued on channel-narrowness + wrapper-snapshot
  detection + L×I×Cx, with an explicit "no premise about actor intent" clause.
- R1-12 (low-medium): all guardrail rules `//`-absolute; anchoring bullet added to §4.2;
  wrapper `cd` precondition stated for scheduling.md.
- R1-8 (medium): `disableBypassPermissionsMode` moved inside `permissions`;
  `disableAutoMode` deferred to OQ17 (leaf-verify before asserting).
- R1-20 (low-medium): H3 verdict split HIGH(non-bare)/OPEN(bare).
- R1-24 (low): OQ3 qualifier carried inline in §4.3 layer 6.
- R1-23 (low): Batch API demoted to future note in §5.2.

**Citation corrections:** R1-1 (backlog 25 items/39 lines — §1.3 + [^IdeasCorpus]);
R1-2 (PDF-gap scope fusion split: 27c one-run/three-seats vs 31h two-runs/red-merge —
§1.2); R1-3 (DGM quote moved [^DGM]→[^DGMSakana]; "(ICLR 2026)" dropped); R1-4 (SICA
venue as the abs page states: "Submitted as a preprint to NeurIPS 2025"); R1-5 (#32191
CLOSED duplicate, 2.1.58–2.1.71 era — §3.3, §6 row 8, §7, [^McpHeadlessBugs]); R1-6
(exit codes: nonzero=failure, exact code unpublished — §3.2 + [^CliReference]); R1-7
([^PortPlan] carries red's confirmation; defect routed to lead + re-raised in
friction.md); R1-9 (Rate Limits API requalification — §5.1, H5 verdict, [^ConsoleLimits],
new [^RateLimitsAPI]); R1-10 (--fallback-model print-only label dropped — §5.1 +
[^CliReference]); R1-11 (pricing leaf figures, both frontier tiers named, tokenizer +30%
note, ≤24h kept MEDIUM — §5.2 + [^Pricing]); R1-13 (chmod-444 recast design-proposed —
§4.3 layer 3 + [^PermAskBypass]); R1-30 (1,557 — §1.3 + [^RedPatterns]).

**Banked upgrades claimed:** STOP ar5iv re-pin (OQ8 RESOLVED; §4.1 + [^STOP] carry the
insignificantly-HIGHER precision); [^UsageAPI] and [^AIScientist] MEDIUM→HIGH; issue
statuses #837 CLOSED COMPLETED / #14246 CLOSED DUPLICATE / #23707 CLOSED NOT PLANNED /
#66395 CLOSED NOT PLANNED added to their footnotes and §3.4/§7.

**Propagation greps (both directions, report-wide):** `40 statused` (0 standing), `1,558`
(0 standing), `leaf-checked OPEN` (only the corrected TWO-open statement + correction
note), `print-only` (only correction notes), `sleeper-ledger` (only the relocation note),
`chmod 444`/`chmod` (only correction contexts), `Bash(node` (only correction notes),
`frontier ~$10/$50` (only the correction note), `no endpoint to read` (only requalified
context), `ICLR 2026`/`ICLR 2025` (only drop notes), `improve themselves the more
compute` ([^DGMSakana] carry + [^DGM] correction note only), `exit code 0/1` (0
standing), `Community defense` (0 standing), `two consecutive runs` (only the corrected
split sentence), `disableBypassPermissionsMode` (all sites show permissions-scoped
placement), `WebSearch` (all sites state the nightly drop). Sites checked: full report,
body AND footnotes.

**claim_count = 142** (same method as round 0, re-derived: 124 body claim lines —
top-level bullets, numbered items, table body rows outside code blocks and footnotes;
6 header rows and 6 separator rows excluded — plus 18 code-block contract units:
/self-improve steps 0–7 and the now-10 idea-stub contract fields). Floor, not ceiling.

## Round 2 — revision (2026-07-17)

All 22 round-2 gaps addressed additively (R2-1..R2-16 lead-carried with owed directions —
each direction executed; R2-17..R2-22 red-minted round 2); zero grade disputes filed
(round-2 grading accepted whole: every required fix priced trivial-to-low-medium, so
absorption again beat contestation on the pragmatist's arithmetic). Pre-flight:
`inputs/red-gap-patterns.md` re-read (window-without-a-watchman, false-equivalence
disjuncts, sibling-repair composition, exhaustive-sweep-omits-own-specimen,
self-referential-repo-drift, incomplete-repair both directions — all applied);
`debate.md` ### RED and ### LEAD round-2 sections read in full before drafting. No
pending dispute deltas (none filed round 1).

**Round-2 probes of record:**

- R2-1: `claude --help` on pinned CLI 2.1.212 confirms `--input-format <format>` with
  `stream-json` ("realtime streaming input", print-mode only) — the two-phase canary
  drive is buildable at the pinned version. Logged in [^CliReference].
- R2-3: FEOV execution locus determined from the shipped command file (plugin copy
  0.7.0, commands/research.md steps 2/3/5): setup-research-run.mjs and
  capture-research-run.mjs are session-Bash `node` invocations; the debate engine is a
  **Workflow tool** invocation (`scriptPath` = debate.js) — MIXED locus; neither of
  red's two branches alone. Logged in [^ResearchCommand].

**Design changes (build-altering):**

- R2-1 (medium-high): canary actor/observer/abort specified — the wrapper is a
  PHASE-DRIVEN stream-json session driver; phase 0 sends the canary-only message, parses
  the event stream + fired-record, and only on pass sends the real prompt; kill+abort
  otherwise. Header/label contradiction reconciled (attempt in-session; actor/observer/
  abort wrapper-side). Edited: §0 tree, §2.2 step 0, §4.3 layer 2.
- R2-2 (medium-high): layer-1-masks-layer-2 closed by a POSITIVE hook-liveness record —
  the guard appends a fired-record (tool, path, nonce, decision) to the operator-owned
  dir on every sleeper-fence deny; wrapper requires deny-in-stream AND nonce-in-record;
  marker loss means no record means ABORT. "Enforcement PROVEN" reworded to "fence
  LIVENESS proven" at §0 tree + §4.3 layer 2; OQ2's acceptance test rewritten to the
  fence-dormant case the new mechanism passes; deny-provenance-in-stream parked as belt
  (OQ21).
- R2-3 (high, the round's priority): mixed locus stated; the two Bash node scripts
  re-hosted WRAPPER-SIDE (§2.2 steps 0/4/7 — setup after the phase-A structured-output
  pick, topic sanitized to slug charset; capture at step 7); the debate engine rides the
  Workflow tool — `Workflow` added to the §4.2 allow list, sleeper-guard hook scopes
  scriptPath to the read-only plugin copy (native rule scoping = OQ20); seat-agent hook
  coverage argued from leaf evidence [^IdeasCorpus]; workflow-script fs writes classed
  layer-4 residual bounded by pinned code + R2-7 compare. §4.3 layer 4 re-scoped
  ("removes all MODEL-AUTHORED script execution"); §6 row 4 re-graded Low to Low-Medium
  and re-argued on the true premise (supersedes the R1-21 re-argument), leg (b) now
  backed by R2-7's watchman.
- R2-5 (medium-high): origin tagging moved from `*_self-improve/` name glob to
  wrapper-stamped causal-provenance markers (`inputs/.sleeper-origin` at creation for
  BOTH dirs — possible because R2-3 moved sub-run setup wrapper-side) + step-7
  porcelain-delta sweep + run-window fallback; harvest reads markers, never names.
  §1.5 restated; §6 row 10 re-graded on the new mechanism.
- R2-6 (medium): disjunction DECIDED — severity-gated bypass (red's shape,
  lead-endorsed): infrastructure-failure classes enter flagged `sleeper-only` without
  corroboration; ordinary classes need one non-sleeper occurrence; both horns' costs
  owned in §1.5; row 10 tracks the decision.
- R2-7 (medium): the snapshot delta gets its watchman — wrapper compares end-vs-start
  and start-vs-prior-end at step 7; mismatch raises a dated TAMPER flag + doctor line +
  NEXT preflight fails closed. §4.3 layer 5, §2.2 step 7, §6 row 4 leg (b)/revisit
  trigger.
- R2-8 (medium): plugin-copy lifecycle built — refresh-sleeper-copy step in
  scheduling.md's guardrail-PR checklist; operator-approved content hash; preflight
  recomputes and FAILS CLOSED on mismatch; doctor staleness line. §3.2, §2.2 step 0.
- R2-9 (low-medium): decision, both horns owned — SessionStart-hook staleness warning
  (N=7 days) in interactive sessions as a passively-received push-adjacent channel
  independent of the scheduler; residual latency stated honestly (unbounded only if the
  operator abandons the tool); §6 row 15 re-graded with the latency term. §3.4.
- R2-10 (low-medium): per-cause dead SIGNATURE (normalized abort reason); M=3
  consecutive same-signature fresh-dir deaths trigger a HALT marker, preflight refuses
  until human-cleared, dead-man flag with reason; "cannot be resumed nightly forever"
  softened to per-dir scope. §3.4, §2.2 step 0, §6 row 15.
- R2-11 (low-medium): `graduation-queued` human-set status exempts a stub from
  auto-stale while still deduping; window stays tunable. §1.4 stage 2, §2.3 status enum,
  §6 row 3.
- R2-12 (low): completeness read from the wrapper's OWN step-7 ledger record
  (operator-owned dir), never run-dir contents; DEAD marker located beside the ledger.
  §2.2 step 0, §3.4 idempotence.
- R2-17 (low-medium): trade CHOSEN — nightly profile stays strict qmd-only and ACCEPTS
  degraded PDF citation capability (vs re-opening #68375 unattended); false "ToolSearch
  reaches pdf/arxiv" parenthetical corrected (strict-mcp-config removes the declaration);
  standing confidence caveat added to the stub contract; full PDF tooling at graduation.
  §3.3(a), §2.3, §2.2 step 4.
- R2-18 (low-medium): expected per-run spend stated (~$0.10–0.50, from the measured
  ~50k-token smoke shape + probe P2's $0.058); $2–5 declared per-run CEILING/anomaly
  bound; 30 x expected = ~$3–15/mo vs the $50 cap, so cap-trip = anomaly signal; skip
  records carry REASON, printed on the dead-man line. §5.2 table, §1.4, §3.4.

**Text/enumeration/citation repairs:**

- R2-4: §2.2 step 2 reworded — "read the wrapper-staged scored table" (scoring ran
  wrapper-side in harvest.mjs); no in-session script.
- R2-13: §6 row 5 cell requalified "(no spend-limit API; rate-limit API unreachable at
  this auth tier — §5.1/R1-9)".
- R2-14: §7 Pattern B/E bullet upgraded ("leaf-verified HIGH round 1, R1-11"); R1-11
  added to §7's banked-upgrade list; §7 Round 2 update block added.
- R2-15: §3.4 rung-1 label now "RECOMMENDED default among SCHEDULED rungs, once the
  human opts in"; the table's "Reading:" paragraph aligned.
- R2-16: R0 L2 cell split ("fence YES (cache copy); canary n/a"); rung-0 manual spend
  declared OUT-OF-LEDGER BY DESIGN with consequences owned + cheap upgrade named. §3.4.
- R2-19: §0 artifact enumeration made total over the printed tree (+ continuous-learning
  skill file + plugin manifest).
- R2-20: [^Backlog] range 15–17 corrected to 15–18 with correction note.
- R2-21: §2.4 "exact" changed to "direct" + honesty clause (DGM admits low scorers;
  admission = compile+edit-ability; our /graduate is stricter, pass-required). No H2
  re-grade.
- R2-22: [^EfficiencyPlan] added beside the $149.95 run-3 figure in §5.2.
- Lead direction: invariant 7 added to §0 (checked, human-surfaced liveness/outcome
  record for every wrapper gate), with the four derived instances named; §0 Round-2
  revision header added.
- Banked round-2 upgrades claimed: [^WindowsHang] MEDIUM to HIGH (regression span
  body-confirmed), [^Pricing] zero-drift re-fetch + Batch 24h sub-claim MEDIUM to HIGH
  (footnote + §5.2 body both updated), [^CliReference] gains --input-format +
  --json-schema version marker, #76239/#68375 re-confirmed OPEN.

**Propagation greps (both directions, report-wide):** `enforcement PROVEN` (0 standing —
reworded at both sites; "proven" survives only inside correction contexts), `removes ALL
script` (0 standing; survives only inside the R2-3 correction quote), `*_self-improve/`
glob-as-mechanism (0 standing; the string survives as dir name + correction contexts),
`already looks at` (0 standing; survives only inside the R2-9 correction quote), `cannot
be resumed nightly forever` (now scoped "within one run dir"), `$2–5` (3 sites — §1.4,
§5.2 table, §3.4 R2-10 text — all now ceiling-labeled with expected spend stated),
`RECOMMENDED default` (qualified at the rung row AND the "Reading:" paragraph),
`analogy is exact` (0 — now "direct"), `no API` (§6 row 5 requalified; §5.1 sites were
already correct per R1-9), `15–17` (0), `carried MEDIUM pending` (0 — body + footnote
both resolved), `graduation-queued` (present at §1.4/§2.3/§6 row 3 — all three
stale-rule sites), `denial canary` (remaining mentions are the historical round-1 record
+ the gate-table row label; every mechanism site carries the fired-record).
claim_count 142 to 146 (method unchanged: +invariant 7, +§7 Round-2 bullet, +OQ20,
+OQ21).
