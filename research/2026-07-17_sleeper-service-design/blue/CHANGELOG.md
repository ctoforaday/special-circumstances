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
