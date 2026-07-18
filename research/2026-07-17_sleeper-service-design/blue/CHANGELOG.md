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

## Round 2 — revision (2026-07-17; CATCH-UP entry written at round 3)

Process defect, owned: the round-2 revision shipped in `blue/report.md` but this entry
and the `debate.md` ### BLUE block were never written — the change-summary channel ran
behind the living report and cost red's lens 3 a misjudged round state (logged by the
lead in friction.md). This entry is reconstructed at round 3 from the report's own R2-x
markers; the report text is the artifact of record.

All 22 round-2 gaps addressed (R2-1..R2-16 lead-carried, each owed direction executed;
R2-17..R2-22 red new-mints). Design changes: two-phase stream-json canary drive with
hook fired-record (R2-1/R2-2 — §2.2 step 0, §4.3 layer 2, OQ2 rewritten, OQ21 added);
FEOV execution locus probe-determined MIXED, session-Bash setup/capture re-hosted
wrapper-side, Workflow channel argued (R2-3 — §2.2 step 4, §4.3 layer 4, §6 row 4
re-graded Low→Low-Medium); step-2 "script" reword (R2-4); causal-provenance origin
markers on every spawned run dir (R2-5 — §1.5); severity-gated corroboration decided
(R2-6 — §1.5, §6 row 10); step-7 snapshot compare + TAMPER flag + fail-closed preflight
(R2-7 — §2.2 step 7, §4.3 layer 5); plugin-copy hash lifecycle (R2-8 — §3.2);
SessionStart staleness warning (R2-9 — §3.4); per-cause HALT (R2-10); graduation-queued
exemption (R2-11 — §1.4); wrapper-ledger idempotency completeness (R2-12); §6 row 5
cell (R2-13); §7 Pattern B/E upgrade note (R2-14); rung-1 label qualifier (R2-15);
rung-0 L2 cell split + out-of-ledger declaration (R2-16 — superseded round 3 by R3-2);
strict-mcp-config PDF trade stated (R2-17 — §3.3, §2.3); cap arithmetic reconciled
(R2-18 — §5.2); §0 enumeration totality (R2-19); backlog range fix (R2-20); DGM
admission-bar honesty clause (R2-21 — §2.4); run-3 figure artifact-of-record marker
(R2-22). Invariant 7 added at the lead's direction. claim_count was not re-derived at
round 2 (part of the same process defect); the round-3 figure below covers both
revisions.

## Round 3 — revision (2026-07-17)

All 17 round-3 gaps addressed additively; zero grade disputes filed (third consecutive
round — every required fix priced trivial-to-low, including R3-17 where red offered
risk-accept and the one-clause fix was cheaper than the acceptance argument). Pre-flight:
`inputs/red-gap-patterns.md` re-read (incomplete-repair both-directions greps,
false-equivalence disjuncts, identity-keyed detectors, exhaustive-sweep-omits-own-specimen
applied); debate.md ### RED + both ### LEAD sections read in full. Blue leaf
verifications before absorbing the two build-altering findings (critical-stance):
`git log -1 --oneline --output=/tmp/...` re-run on this box — exit 0, file created, NO
prompt (confirms R3-15 and shows the carve-out classifier itself passes `--output`, so
rule-pinning alone is insufficient); permissions doc re-fetched live — carve-out quoted
verbatim, set BROADER than the gap summary (adds ls/echo/pwd/wc/which/diff/stat/du/cd),
all deny-enumerated.

Design changes:

- R3-2 (medium-high, the round's priority): rung 0 DECIDED — manual runs invoke the SAME
  wrapper via `--manual` (lead-endorsed R2-16b resolution). Edited: §1.4
  null-alternative, §2.2 step 0, §3.4 ladder row 0 + gate-survival R0 cells
  (L1/L2/wrapper rows) + the R2-16b paragraph replaced (out-of-ledger accept DISSOLVED),
  §6 row 10.
- R3-14 (medium-high): dontAsk read-only Bash carve-out absorbed — §3.2 sentence
  corrected; §4.2 JSON gains one deny per carve-out command + prose bullet rewritten on
  the true surface (git allow rules relabeled declared-intent/redundant); §4.3 layer 2
  unforgeability leg re-derived; layer 4 dontAsk parenthetical; §6 rows 4 + 13
  re-argued; [^PermissionsDoc] carries the verbatim carve-out quote; OQ18 extended
  (redirection-under-carve-out).
- R3-15 (medium-high): `git log --output` write gadget closed three ways — exact-argv
  pinning (wildcard dropped from the allow rule), `* --output*` belt denies,
  sleeper-guard Bash-write matcher extended to git-output-class flags. §4.2 (rule + new
  bullet), §4.3 layer 4 (i) re-scoped ("read-only EXCEPT one named channel"), §6 row 4
  leg (a) re-argued, OQ18 (b)/(c).
- R3-7 + R3-4 (folded per the lead): snapshots at EVERY wrapper-observed exit; compare
  chained to last-recorded-snapshot-of-any-kind, missing link fails closed; run-window
  START logged at step 0, END at any exit; origin sweep also at DEAD-marking. §0 tree +
  invariant 7 wording, §1.5, §2.2 steps 0/7, §4.3 layer 5.
- R3-8 (added to this list round 4, R4-16 — the bullet was omitted from the round-3
  summary though the report body addressed it): snapshot sensing scope stated honestly —
  the compare detects in-repo working-tree deltas + guardrail-file hashes only;
  OUT-OF-REPO subprocess writes (`~/.claude`, sibling trees, `/tmp`) are invisible to
  porcelain and the named-file hash set, so that residual is bounded by pinned code +
  no-remaining-write-gadget, NOT by detection. §4.3 layer 4 (iv), layer 5, §6 row 4.
- R3-3: provenance tagging extended to the red-memory mirror surface (snapshot-diff →
  sleeper-authored-patterns list; excluded from corroboration pool; staged mirror stays
  pre-run frozen). §1.5 new paragraph, §2.2 step 0, §6 row 10.
- R3-5: infra-class tag assigned SOLELY from the wrapper event log; friction text only
  ever ordinary-classed. §1.5.
- R3-11 + R3-6 (folded per the lead): dead-signature normalization specified (exit class
  + templated abort line, placeholders for dates/paths/ids/nonces/digits);
  zero-HALT-firings telemetry; alternating-cause residual owned; doctor line prints
  per-signature counts since last clear. §3.4 (two sites), §1.5 corroboration paragraph.
- R3-13: graduation-queued gains M-day (default 90) queued-stale re-surface. §1.4, §2.3
  status enum (which also gains the graduation-queued/queued-stale values the round-2
  edit had omitted), §6 row 3.
- R3-1: mid-drive `--json-schema` demoted to verify-at-build with fenced-block
  text-parse fallback (OQ22); three behavioral acceptance legs added (OQ23);
  harvest-failure = fail-closed skip; qmd degrade note destination/readers named. §2.2
  steps 0/3, §8.
- R3-9 (recommend-not-block): cheap parts absorbed — wrapper/harvest/configs named in
  layer 3's enumeration + snapshot hash set; absolute-path invocation from the
  operator-owned copy; SessionStart always-prints last-check timestamp; doctor prints
  wrapper-file hash dates. Self-verification limit OWNED as stated residual
  (box-compromise scenario, not loop-autonomy). §4.3 layer 3, §3.4.
- R3-10: §0 enumeration extended (SessionStart hook + hooks.json, CROSS-PLUGIN
  doctor-line delta, two operator-owned configs); SessionStart warning conditioned on
  scheduling-enabled. §0, §3.4.
- R3-12: in-night retry semantics stated (one resume per nightly fire); arithmetic
  recomputed (≤$5/night; cap ~10 ceiling nights; HALT night 12 or cap, whichever
  first). §3.4.
- R3-16: board-telemetry row corrected to SHIPPED (FEOV 0.7.0, present in this run's
  own trajectories/). §1.3.
- R3-17: tokenizer set completed (Opus 4.7+ members added). §5.2 + [^Pricing].
- Invariant 8 added (§0) at the lead's direction; round-2 and round-3 revision summary
  paragraphs added to the report header (reader orientation — part of the lens-3
  round-state repair).

Propagation greps (both directions, report-wide): `same code path` (only R3-2 correction
contexts), `out-of-ledger`/`OUT-OF-LEDGER` (only supersession notes), `k×$5` /
`~3 nights` (only the recomputation note), `auto-denies anything` (only the correction
bullet), `git log *` (only the correction comment; the allow rule now exact),
`Fable/Mythos/Sonnet-5 tokenizer` (0 standing — both sites carry the completed set),
`shipping in FEOV 0.6.0` (0 standing), `detection latency is one run` (0 standing — both
sites now per-exit), `step-7 auto-compare`/`step-7 snapshot compare` (§0 tree, invariant
7, layer 5, step 7, row 4 all updated to every-exit), `no Bash beyond` (0 standing),
`read-only git argv`/`read-only analysis tools`/`pinned-argv read-only` (only correction
contexts), `sole channel` (both §1.5 sites carry the R3-6 mechanism), `dontAsk` (all 11
sites checked: corrected, neutral, or correction-context). Sites checked: full report,
body AND footnotes.

**claim_count = 151** (same method, re-derived: 133 body claim lines — 49 top-level
bullets + 38 numbered items + 46 table body rows outside code blocks and footnotes,
6 header and 6 separator rows excluded — plus 18 code-block contract units:
/self-improve steps 0–7 and the 10 idea-stub contract fields). Floor, not ceiling.
Covers rounds 2+3 (round 2's figure was never derived — see the catch-up note above).

## Round 4 — revision (2026-07-17)

All 16 round-4 gaps addressed additively; ZERO grade disputes filed (fourth consecutive
round — every required fix priced trivial-to-low; the two weight-bearing catches, R4-2 and
R4-4, are genuine mechanism defects a dispute could not honestly deflect). Pre-flight:
`inputs/red-gap-patterns.md` re-read (invariant-soundness-by-enumeration → the allowlist
inversion for R4-2/R4-3; incomplete-repair both-directions greps; missing-root-invariant →
R4-9; sibling-repair-composition → R4-14; doctrine-vs-implementation → the §3.4 polarity
R4-1). debate.md ### RED + ### LEAD read in full; every carried gap's owed direction
executed.

Lead's cross-cutting direction followed: red's round-4 shape was "the round-3 repair closed
the named path but not its SIBLING"; answered with STRUCTURAL closes from invariants 6/8
rather than neighbor-by-neighbor enumeration.

**Design changes (build-altering):**

- R4-3 (medium, the round's STRUCTURAL priority): **bare `Bash` deny** added to the sleeper
  profile — a bare tool name removes the tool entirely (doc-verified), and §2.2's session
  steps never invoke Bash, so the whole dontAsk read-only-Bash carve-out class (enumerated
  members, doc-non-exhaustive unlisted members, read-only git, every git write gadget) is
  closed at the tool boundary. Enumerated per-command denies RETAINED as belt for rebuilt
  rungs; `sort`/`sed`/`file`/`readlink`/`strings`/`less` + `~/.ssh`/`.env` belt denies added.
  Edited: §0 round-4 header, §4.2 JSON deny block + parenthetical + two new prose bullets,
  §4.3 layer 2 + layer 4 header + layer 4 (i), §6 row 4 leg (a), §7 attribution note, OQ18(c).
- R4-2 (medium-high): git channel INVERTED to a read-ALLOWLIST at the hook (deny any git
  argv not an exact allowed read form); belt denies extended to `-o`/`-O` short forms and the
  `format-patch`/`archive`/`bundle`/`config`/`gc`/`repack`/`maintenance` writer family;
  row 4's "deny-enumerated per command" corrected to name the git exception with OQ18(c) as
  its standing test. Edited: §4.2 deny block + git-allowlist bullet, §4.3 layer 4 (i), §6
  row 4, OQ18(c).
- R4-15 (low): OQ18(c) extended to the git SUBCOMMAND boundary (config/gc/repack/maintenance
  named as probe cases); hook git-allowlist rejects them by default.
- R4-1 (medium): `/self-improve` becomes a thin trampoline (`node …/sleeper-wrapper.mjs
  --manual`) carrying `disable-model-invocation: true`, payload moved to the skill; §3.4
  containment POLARITY corrected (markerless out-of-contract dirs are NON-sleeper and CAN
  corroborate — the R2-5 4th surface, named residual = a human's deliberate paste-run).
  Edited: §0 tree + artifact enumeration, §3.4 containment paragraph, §6 row 10, OQ24 added.
- R4-4 (medium/certain): red-memory writes DECLARED DENIED by design under the sleeper
  profile (nightly seats do not learn; the fence-denial is a NORMAL fired-record class); the
  R3-3 machinery re-scoped to belt-for-drift; tagging at window-ADDED granularity (R1-22
  monotonic-blinding leg). Edited: §1.5 R3-3 paragraph rewritten, §6 row 10.
- R4-6 (low-medium): window END bounded by NEXT wrapper START; unobserved-exit (hard-kill)
  window flagged `retroactive-uncertain`, markerless sweep confined to sleeper date-key
  naming; snapshot-chain + resume backstop owned in step-7/layer-5. §1.5, §4.3 layer 5.
- R4-9 (low-medium): status re-surface ROOT INVARIANT stated once in §1.4 (no status
  timer-free); terminal states `graduated`→regression-on-recurrence, `rejected`→90-day/rate
  `rejected-recurring`. §1.4, §2.3 status enum, §6 row 3.
- R4-8 (low/certain): code-artifact count corrected THREE→FOUR (SessionStart hook counted +
  host plugin named sleeper-service). §0.
- R4-7 (low): qmd-degrade surfacing — §2.3 confidence field labeling clause + §3.4 doctor
  degrade-streak term.
- R4-10 (low/certain): cap/HALT arithmetic recomputed — cap trips FIRST (~night 10), death
  accrual pauses, HALT lands early NEXT month; worst-case ≈ one cap + ≤2 nights next month.
  §3.4.
- R4-11 (low): git `--output` gadget ATTRIBUTION re-scoped honest (both probes ran under
  `defaultMode: auto`, so consistent-with but not isolating; isolating probe deferred to
  OQ18(c)/OQ23; hook matcher chosen-conservative). §4.2 R3-15 bullet, §7, OQ23(d).
- R4-12 (low): est_complexity sourced — default 1 (inert) unless the matching backlog entry
  carries a human-recorded complexity note. §1.4 stage 2.
- R4-13 (low): gate-survival table gains a Provenance/origin-tag + corroboration gate row
  (YES R0/R1, NO R2-R4 absent rebuild); named in the graduation-grade adoption requirement.
  §3.4.
- R4-14 (low): dead-man flag custody — flag lives in the operator-owned dir (loop can't flip
  it); SessionStart line never fully silent (distinguishes disabled-by-choice from
  silent-disarm from flag-loss). §3.4.
- R4-5 (low-medium): postmortem corrected at BOTH sites (§4.2 prose bullet AND §6 row 13) —
  the round-3 `Bash(cat …/.claude/projects/…)` example was mis-diagnosed (that path was
  Read-denied, and Read denies extend to Bash file commands); re-pointed at un-named
  credentials-class paths; deny-reach clause noted.
- R4-16 (trivial): R3-8 bullet added to CHANGELOG Round 3.

**Propagation greps (both directions, report-wide):** `deny-enumerated per command`/`hence
deny-enumerated` (only §6 row 4 correction context + §0/§7 historical round-3 summaries now
carrying a supersession note), `exactly THREE`/`THREE new code` (0 standing — now FOUR),
`writes the SHARED agent-memory` (0 standing — R3-3 premise rewritten), `is ordinary-classed,
requiring corroboration` (0 standing — §3.4 polarity corrected), `AUTO-APPROVED under the
round-2` (only R4-5 correction context, BOTH §4.2 + §6 row 13 sites re-pointed),
`carve-out channel remains`/`rule-pinning ALONE`/`carve-out classifier itself passes` (0
standing — R4-11 both sites re-scoped), `bare `Bash` deny` (all sites consistent),
`night 12 OR the cap` (0 standing — R4-10 recomputed). Sites checked: full report, body AND
footnotes AND code-block comments.

**claim_count = 155** (same method, re-derived floor: round-3's 151 + 2 new §4.2 prose
bullets + 1 gate-survival table row (R4-13) + 1 open question (OQ24); the status enum gained
a value and many paragraphs grew, but those are prose within existing counted units, not new
countable lines). Floor, not ceiling.

## Round 5 (2026-07-17)

All 10 round-5 gaps addressed (all lead-carried with owed directions — each executed);
zero grade disputes a fifth round (every required fix priced trivial-to-low; the two
mediums earned structural fixes, not contests). Lead's cross-cutting direction followed:
state the invariant ONCE over the TRUE surface (seat population; corroboration pool),
not another per-surface lap. Pre-flight: inputs/red-gap-patterns.md re-checked — the
un-propagated-repair pattern (grep the retracted token report-wide) and
exhaustive-sweep-omits-own-specimen both applied below.

**Design changes (build-altering):**

- R5-1 (medium, the round's priority): seat-population HORN PICKED — the FEOV seat agents
  are INSIDE the sleeper boundary (bare `Bash` deny binds them) as a DESIGN REQUIREMENT,
  with the evidence honestly scoped (hook-layer seat coverage is interactive-leaf-verified;
  layer-1 `--settings` inheritance for headless-spawned seats is UNPROBED, now acceptance
  leg OQ23(e)); the sleeper-guard hook is GENERALIZED per invariant 6 to deny the WHOLE
  Bash channel for sleeper-marked runs, so seat-surface closure is inheritance-INDEPENDENT
  (either probe outcome leaves Bash closed on the seat surface; R1-13 read+egress does not
  re-open). Capability cost owned: nightly seats run BASH-FREE — smoke-lane + citation-pass
  work needs no Bash (repo state wrapper-staged), command-execution leaf probes are a
  graduation capability, and seat friction stays visible as an ordinary class. Sub-leg:
  the four `Bash(git ...)` allow rules re-labeled DEAD under the shipped profile (bare-name
  removal + deny supremacy — the "carve-out auto-approves read-only git regardless" comment
  REFUTED and withdrawn at BOTH sites); R4-2's reachability list corrected ("the Workflow
  seat agents" removed). Edited: §0 revision header, §0 tree (hook description), §2.2
  step 4, §4.2 JSON allow-comment + prose bullet + R4-2 bullet, §4.3 layer 4 (iii)
  rewritten, §5.2 bounded-pass row, §6 row 4 + row 13, OQ23(e).
- R5-3 (medium): corroboration-pool ROOT INVARIANT stated once in §1.5, mirroring R4-9 —
  the pool is POSITIVELY defined: a dir corroborates only with affirmative non-sleeper
  provenance ((a) creation before the sleeper deployment EPOCH recorded at schedule-setup;
  (b) interactive-origin marker `inputs/.run-origin` stamped at creation by
  setup-research-run.mjs — a one-line CROSS-PLUGIN edit to FEOV, enumerated in §0);
  anything unattributable is QUARANTINED counts-for-nothing, docket-flagged `unattributed`
  for one-keystroke attribution; sleeper-positive evidence DOMINATES an interactive marker
  (forgery precedence; conflict = quarantine + anomaly). Dissolves the R4-1/R4-6/R4-13
  per-surface residuals in one predicate; the rung-2 automatic-local-poisoning case is
  closed by mechanism, not adoption instruction. OQ24 RESOLVED by promotion (generalized
  form — no loop-shaped-dir schema heuristic, no schema false-positives). Edited: §0
  revision header + cross-plugin enumeration, §1.5 new root-invariant paragraph, §3.4
  R4-1 paragraph (paste-run residual DISSOLVED, not risk-accepted), §3.4 gate-survival
  provenance row (R2 NO to PARTIAL with quarantine noted; R3/R4 cells annotated), §6
  row 10, OQ24.
- R5-2 (low-medium/certain): §3.4 rung-0 cell REWRITTEN to the R4-1 thin-trampoline shape
  (payload in the wrapper's phase-1 prompt from skills/continuous-learning; the command is
  an inert trampoline); Phase-4 acceptance test RESTATED two-legged in §3.3 — (i) `node
  sleeper-wrapper.mjs --manual` produces run dir + stub touching only research/+ideas/;
  (ii) `claude -p "/self-improve"` produces NO run dir (trampoline INERTNESS as the
  verifiable property — the old printed gate's cheapest pass would have re-inlined the
  payload). [^PortPlan] footnote annotated (plan's verify step superseded as-written);
  §6 row 1 disposition re-pointed. "phase-1 prompt payload" added to the grep list.
- R5-4 (low-medium): §1.5 doctrine sentence QUALIFIED — name-keying permitted only to
  CONFINE retroactive-uncertain sweeps, never to assign origin outside one; the sub-run
  slug made KNOWABLE — the wrapper appends the sub-run dir PATH to the run-window log AT
  CREATION beside the START record (it creates the dir itself at phase B), so confinement
  matches recorded paths after a hard-kill; mkdir-to-stamp bound stated explicitly (one
  synchronous statement; a kill inside it leaves a dir the positive pool quarantines,
  fail-closed). §6 row 10 scoped accordingly.
- R5-5 (low): belt declared KNOWN-INCOMPLETE on ATTACHED short forms IN-FILE (`-o<value>`
  matches no pattern; leaf-reproduced twice by red) — deliberately NOT extended again
  (enumerate-and-extend cannot close the class; the escape recurred inside R4-2's own
  repair); the bare deny + hook read-allowlist named the closes of record; OQ18(c) gains
  probe leg (3): the hook's attached-form handling leaf-tested, not assumed. §4.2 deny
  block comment, §6 row 4 leg (a), OQ18(c).
- R5-6 (low-medium): `expected-memory-deny` bucket predicate NARROWED to actor/shape —
  ONLY the red-merge seat's record-new-patterns write shape (seat role + pattern-file
  target), bounded N<=2/run; any other actor/target-subdir/volume files as
  `unexpected-memory-write-attempt`, doctor-surfaced, PERSISTS like TAMPER. Restores
  tamper-attempt-must-be-visible on the memory surface (invariant 7's own test). §1.5.
- R5-7 (low): (i) sleeper-service's hooks.json wraps EVERY command in the bootstrap guard
  — stated in §0; §6 row 9 re-pointed at BOTH guarded files; OQ10 empty-bin acceptance
  check extended to cover both. (ii) flag-absent anomaly GATED on a prior
  `last-successful-run` record — no run history means never configured, silent/one-time
  notice; with history, anomaly (flag-lost). §3.4.
- R5-8 (low/certain): §0 tree gains `hooks/hooks.json` (registering the sleeper-guard
  PreToolUse fence AND the SessionStart staleness hook) and the `sc-sleeper-staleness`
  SessionStart executable — the tree now draws all FOUR code artifacts + registration
  (the enumeration was repaired three times while the tree never was).
- R5-9 (low/certain): dated status field — every human status change appends its date
  (`status: rejected 2026-07-17`); harvest parses it; both `rejected` re-surface arms key
  on it; dateless terminal status = fail-toward-re-surface (read as set-at-MINT, flagged
  `undated-status` once). Token DOMAINS pinned: stub-file statuses (human-edited, dated)
  vs docket flags (harvest-computed, never written to the stub) — `regression` is
  docket-only (stub stays `graduated`; a new stub may cite the ancestor);
  `rejected-recurring` setter is harvest, docket-only; re-confirm = re-dating the stub
  line. §1.4 (clause (b) + domain paragraph), §2.3 enum rewritten.
- R5-10 (trivial/certain): est_complexity note format stated — literal `cx:<1-5>` token in
  the backlog entry line, anything else = default 1 (inert); and the source's honest
  status at the pin stated (red's L1 leaf: NO structured field on any of the 25 items —
  the divisor is universally inert today; `cx:` is a forward-looking curation surface).
  §1.4 stage 2.

**Notes adopted (not gaps):** §7 gains its missing Round-4 update bullet plus a Round-5
bullet banking red's corroborations (ten living-source sets zero-drift; cap/HALT
arithmetic triple-recomputed; [^Pricing] zero-drift); [^McpHeadlessBugs] notes #68375's
new `stale` label (bot auto-close = drift risk, not a fix claim); [^MissedRun] re-pointed
at the StartWhenAvailable API page (settings URL 404s).

**Propagation greps (both directions, report-wide):** `read-only git regardless` (2
standing, BOTH inside correction contexts quoting the withdrawn claim — §4.2 JSON comment
+ §4.2 prose bullet; 0 live assertions), `the Workflow seat agents` (1 standing, inside
the R5-1 correction context), `full permission-engine + hook subjects` (0 standing —
layer 4 (iii) rewritten), `phase-1 prompt payload in EVERY mode` (0 standing — rung-0
cell rewritten), `counted NON-sleeper`/`counts them non-sleeper` (both remaining sites
carry the R5-3 supersession note), `PROVIDES the non-sleeper corroboration` (correction
context only), `never the dir name` (qualified at the doctrine site; row 10 scoped),
`sub-run slug` (correction context only), `Deferred, not built` (0 standing — OQ24
resolved), `produces a run dir` (remaining sites: §3.3 correction context + [^PortPlan]
annotated quote + §6 row 1 re-pointed), `expected-memory-deny` (§0 header + §1.5
predicate, consistent), `auto-approves` (0 live). Sites checked: full report, body AND
footnotes AND code-block comments AND table cells.

**claim_count = 161** (same method, re-derived floor: round-4's 155 + 1 §1.5 pool
root-invariant paragraph + 2 §0 tree artifact entries (hooks.json, sc-sleeper-staleness)
+ 1 OQ23(e) leg + 2 §7 update bullets; OQ24's resolution and the enum/cell rewrites are
changes within existing counted units). Floor, not ceiling.
