# Round 4 — Lens 5 (logic & completeness) candidate findings

Audit surface: full re-read of `blue/report.md` (1893 lines, 4 consecutive Read windows,
2026-07-17, post-round-3 revision). CHANGELOG used as navigation hint only. Lens-scoped
ids; merge assigns stable R4-N ids and lineage.

Round-state note for merge: blue's round-3 revision IS present in report.md AND the
CHANGELOG carries a Round 3 entry — the round-2 desync did not fully recur; see L5-F8 for
the partial recurrence (one gap missing from the CHANGELOG enumeration).

---

## L5-F1 — /self-improve ships as an invocable command carrying the full loop payload; the unwrapped path R3-2 closed for rung 0 stays mechanically open via command invocation, guarded only by a prose header, and §3.4's containment claim for that path has INVERTED polarity (candidate supersedes: R3-2)

- location: §0 tree — "commands/ ... self-improve.md — daily loop: preflight → harvest →
  pick ONE → bounded research → idea stub" (no `disable-model-invocation` frontmatter;
  graduate.md carries it); §3.4 — "An operator bypassing the wrapper by pasting the
  command markdown into a bare interactive session is out-of-contract — stated in the
  command file's own header (the markdown is the wrapper's phase-1 payload, not a
  standalone entry point), and such a session carries the interactive profile's own
  permission gates rather than silently impersonating a sleeper run (no marker ⇒ the
  fence never engages, but also: no marker ⇒ its dirs are swept by the window log if any
  wrapper run brackets them, and its friction is ordinary-classed, requiring
  corroboration)."
- problem, three composing legs:
  (a) **Doctrine vs implementation.** "Not a standalone entry point" is contradicted by
  shipping the payload under `commands/` — the harness's entry-point directory. A named
  `/self-improve` command is invocable by the operator in any interactive session (an
  attractive nuisance: the natural way to "run it by hand" is typing the command, not
  `node sleeper-wrapper.mjs --manual`), AND — absent the `disable-model-invocation: true`
  frontmatter the design applies to its sibling graduate.md — invocable by the MODEL
  (SlashCommand-tool / scheduled-fire path; the report's own [^ScheduledTasks] quote
  establishes that only flagged skills "reach Claude as plain text"). The asymmetry
  (flag on graduate, none on self-improve) is never argued; after rounds 1–3 established
  that UNWRAPPED loop execution voids ledger, canary, fence, origin markers, and the
  corroboration gate, an unwrapped invocation is no longer harmless.
  (b) **Instructional gate on a refuted class.** The sole guard on this path is the
  command-file header — prose. §4.1's own doctrine: "Instructions in your prompt ...
  don't change what Claude Code allows"; "0/175"; the MUST-clause skipped live. The
  design defends its default-mode integrity with exactly the gate class it refutes.
  (c) **Inverted containment.** The reassurance clause is logically backwards. An
  out-of-contract run's dirs are markerless and (unless a concurrent wrapper window
  happens to bracket them — a coincidence, not a mechanism) classed NON-sleeper. Under
  §1.5's decided rule, an ordinary class needs "at least one non-sleeper occurrence" —
  so the out-of-contract run's friction does not FACE the corroboration requirement, it
  SELF-SATISFIES it (it *is* a non-sleeper occurrence) and can corroborate sleeper-origin
  occurrences of the same class into the top-3. "Its friction is ordinary-classed,
  requiring corroboration" states the threat vector as a mitigation. Additionally such a
  session runs the harvest→research→stub flow under the interactive profile — no fence,
  so the §1.5 never-edits rule (incl. backlog.md) is prompt-discipline-only there: the
  H4-refuted configuration executing the loop payload.
- required_fix: (i) make the command a thin trampoline — the command file instructs
  launching the wrapper (`node <operator-owned>/sleeper-wrapper.mjs --manual`) and
  carries NO loop logic; the phase-1 payload moves out of `commands/` (e.g.
  `prompts/self-improve-payload.md`, read by the wrapper only); (ii) add
  `disable-model-invocation: true` to self-improve.md regardless; (iii) correct §3.4's
  containment sentence — markerless out-of-contract dirs enter harvest as non-sleeper and
  CAN corroborate; if the residual is accepted, say so with that polarity, not the
  inverse; re-affirm §6 row 10 with this residual named.
- pattern refs: doctrine-vs-implementation; origin-tag scope (the untagged actor joins
  the trusted pool); policy-without-mechanism.
- grading: medium (human path likely over time — the command name invites it; model path
  possible) × medium (poisoning-gate erosion + unfenced loop execution; human present
  mitigates spend/writes) × low (trampoline + one frontmatter line + one sentence)
  → severity **medium**

## L5-F2 — row 4 leg (a)'s "carve-out ... hence deny-enumerated per command" is FALSE for its git member: read-only git is IN the carve-out, deliberately NOT deny-enumerated, and is an OPEN-ENDED platform-classified class in which R3-15 already proved one write gadget — sibling git-native output flags (`-o` short form: format-patch, archive; `bundle create`) are outside BOTH the belt denies and the hook matcher's named flag list (leaf-verified this box) (candidate supersedes: R3-14, R3-15)

- location: §6 row 4 — "the subprocess surface is the pinned EXACT git argv set + the
  built-in read-only Bash carve-out (non-configurable, hence deny-enumerated per command
  in §4.2)"; §4.3 layer 4 (i) — "the sleeper-guard's Bash-write matcher extended to
  git-output-class flags (`--output`, `--output-directory`, `-O`)"; §4.2 deny list —
  `"Bash(* --output=*)", "Bash(* --output *)", "Bash(* --output-directory*)"` (no `-o`
  form; no git-command denies beyond push/commit); [^PermissionsDoc] — the carve-out
  includes "read-only forms of `git`. The set is not configurable."
- problem: §4.2 deny-enumerates the 15 non-git carve-out commands, but read-only git —
  the one carve-out member that already yielded the R3-15 gadget — is retained
  un-enumerated (the design NEEDS git reads), so "deny-enumerated per command" overclaims:
  one member of the carve-out remains classifier-defined, and the classifier was PROVEN
  unreliable for write gadgets by R3-15 itself (`git log --output` passed as read-only).
  The R3-15 closure named exactly three flags for the hook matcher and wrote belt denies
  matching only the long forms. Leaf-verified on this box 2026-07-17:
  `git -C <repo> format-patch -1 -o /tmp/l5probe` → exit 0, patch file created at an
  arbitrary OUT-OF-REPO path. The `-o` short form matches none of the three belt-deny
  patterns and is not in the hook's named flag list; `git format-patch` with NO flag at
  all writes `0001-*.patch` to cwd; `git archive -o <file>` and `git bundle create
  <file>` are further git-native file-writers. Whether the dontAsk classifier passes
  these as "read-only git" is untested (exactly OQ18(c)'s point) — but the in-text
  narrowness claims (row 4 leg (a), layer 4 (i) "no surviving rule OR carve-out entry is
  known to grant ... a write target") read as if the carve-out were fully enumerated
  and the gadget class closed, when what was closed is ONE flag's long forms. This is
  the sibling-halo shape on the R3-15 closure: one audited gadget conferring unearned
  trust on its unaudited siblings, and out-of-repo targets sit in the R3-8-conceded
  snapshot blind spot.
- required_fix: invert to allowlist on the git channel — the sleeper-guard hook already
  parses Bash: deny any `git` invocation whose argv is not one of the exact allowed
  read forms (the design's own invariant 6: "allowlist-defined, never denylist-defined"
  — the current fix is a denylist over an open-ended class); at minimum extend the hook
  matcher + belt denies to `-o`/`--output`-class short forms and the
  format-patch/archive/bundle command family, and correct row 4's "hence deny-enumerated
  per command" to name the git exception with OQ18(c) as its standing test.
- pattern refs: audited-artifact sibling halo; invariant-soundness-by-enumeration
  (denylist under-includes — the design's own memory-mirrored lesson, §7 says it was
  applied); wildcard-allow write-gadget (extension: short-form flags).
- grading: medium (classifier untested over a class with one confirmed gadget; write
  capability of siblings leaf-verified) × high (arbitrary out-of-repo file write — same
  impact class as R3-15) × low (hook is already the enforcement of record; allowlist
  inversion is a matcher change) → severity **medium-high**

## L5-F3 — R3-3's memory-surface provenance tag is CHANGE-keyed, not content-keyed: extending an EXISTING human-established pattern during a sleeper window (the red memory system's documented modal write — "update, don't duplicate") reclassifies the WHOLE pattern sleeper-origin, capping it at 1 and EXCLUDING its accumulated human-present evidence from the corroboration pool (candidate supersedes: R3-3)

- location: §1.5 — "any pattern file or header that appears or changes inside a
  wrapper-logged sleeper window is appended to a `sleeper-authored-patterns` list beside
  the ledger; harvest tags matching pattern headers `origin: sleeper`, caps their
  recurrence contribution at 1 per class, and excludes them from the non-sleeper
  corroboration pool."
- problem: the trigger is file/header APPEARS-OR-CHANGES; the consequence is applied to
  the whole pattern header. Red's own memory discipline (mirrored into this run's
  inputs and cited by §1.4) is to EXTEND existing pattern files rather than mint
  duplicates — so the modal nightly red-seat write is an append to a pattern that
  already carries months of human-present-run evidence. Under the mechanism as written,
  one nightly "now also:" line converts that entire pattern to `origin: sleeper`:
  recurrence capped at 1 AND excluded from the non-sleeper corroboration pool — the
  highest-value, most-recurrent classes are precisely the ones a nightly pass will touch
  soonest, so the docket goes progressively blind to its best-attested signal. Polarity
  is conservative (fail toward exclusion — this is signal loss, not a breach), but it is
  the R1-22 monotonic-blinding failure arriving through the provenance guard itself, and
  §6 row 10's "re-affirming the grade" does not carry it.
- required_fix: tag at the granularity of what the window ADDED — the snapshot machinery
  already hashes content (R3-3): a pattern file that pre-exists the window keeps its
  original origin for its pre-window content/recurrence standing; only a pattern that
  APPEARS in-window (or a genuinely new header) is wholly sleeper-origin; an in-window
  EXTENSION contributes no NEW recurrence but does not strip pre-existing non-sleeper
  standing. One paragraph in §1.5; re-affirm row 10.
- pattern refs: origin-tag granularity (new wrinkle on origin-tag-naming-keyed);
  self-defeating-mitigation (a guard added round 3 starving the docket it protects).
- grading: medium (modal write pattern) × low-medium (docket blinding of top classes;
  safe polarity) × low → severity **low-medium**

## L5-F4 — "no status is timer-free" is asserted while two of the six statuses have neither timer nor stated dedupe semantics: `rejected` and `graduated` — the third consecutive per-status patch (R1-22 → R2-11 → R3-13) without the root invariant, and the enum's terminal members either permanently suppress their class or re-propose it nightly; the text decides neither (first-raise; lineage-adjacent to R3-13)

- location: §2.3 status field — "status: <open | stale | graduation-queued |
  queued-stale | graduated | rejected — ... (R3-13 — no status is timer-free); humans
  set the terminal states>"; §1.4 — "Skip any candidate with an open stub *younger than
  the staleness window*"; §1.4 stage 1 — "Anything already marked DONE/FIXED in
  backlog.md is closed signal — recurrence after a fix is its own high-value class:
  regression."
- problem: the skip/dedupe rule is stated for OPEN stubs (30d), `graduation-queued`
  (90d re-confirm), and backlog DONE/FIXED items (regression re-class). For STUBS a
  human sets to `rejected` or `graduated`, nothing states whether the stub keeps
  deduping its class. Both branches are defective: if terminal stubs dedupe forever, a
  single rejection permanently subtracts a class from the docket with no re-surface —
  exactly the monotonic-blinding failure R1-22's governing principle forbids and R3-13
  just re-patched for the queued state ("no status is timer-free" is false as written);
  if they don't dedupe, a recurring rejected class re-enters the docket and re-mints a
  stub every run — the Dependabot fatigue arm (re-proposing rejected ideas nightly).
  The backlog regression rule covers backlog items, not stub statuses; stubs and
  backlog are distinct dedupe surfaces by the design's own §1.4. Three rounds of
  per-status timers signal the missing root invariant: *every status's dedupe effect
  has a stated re-surface path*.
- required_fix: state the invariant once in §1.4 and give the two terminal states their
  semantics — e.g. `graduated`: class recurrence after graduation re-enters flagged
  `regression` (mirroring the backlog rule); `rejected`: dedupes for a cadence-tuned
  window (or until recurrence exceeds the pre-rejection rate), then re-surfaces flagged
  `rejected-recurring` for one-keystroke re-confirmation like queued-stale. Re-affirm
  §6 row 3.
- pattern refs: missing-root-invariant; waiver-graduation (class-conditional statuses).
- grading: medium (rejection is a modal triage outcome) × low-medium (class blinding or
  gate fatigue) × low → severity **low-medium**

## L5-F5 — the stage-2 ranking formula's `× (1 / est_complexity)` factor has no stated input source: a zero-token deterministic script cannot estimate complexity, and no harvested artifact carries a complexity field — policy-without-mechanism inside the formula the design calls "deterministic, stated ... so it is auditable" (first-raise)

- location: §1.4 stage 2 — "`score = recurrence_across_runs × severity_proxy
  (seat-classes affected) × staleness_decay`, with lane-2's additional factor
  `× (1 / est_complexity)` [minority: lane-2/primary-literature — the complexity
  divisor]"; stage 1's docket schema — "class | occurrences (run, seat) | seats affected
  | max severity seen | first/last seen | staleness | open backlog item? | existing
  disposition | score" (no complexity column, and none of the parsed inputs —
  friction.md, cost.md, telemetry, backlog checkboxes, pattern headers — carries one).
- problem: every other factor is derivable from the parsed inputs (occurrences, seat
  counts, dates); `est_complexity` is a judgment quantity with no source. Either the
  script hardcodes a guess (undermining "auditable"), a model estimates it (violating
  "ranking is arithmetic ... zero tokens" and moving judgment into the mechanics
  stage), or the factor is silently dropped at build (the printed formula is not the
  built formula). The report's own doctrine — "an LLM executing mechanics is an
  unenforced good-faith contract" — cuts the other way here: mechanics executing
  judgment is an unimplementable contract.
- required_fix: name the source or drop the factor — cheapest honest shape: default
  `est_complexity = 1` (factor inert) unless the class's matching backlog entry carries
  a human-recorded complexity note (the curation surface that already exists); one
  clause in §1.4.
- pattern refs: policy-without-mechanism.
- grading: certain (no source stated) × low (mis-weights one factor; the human sees the
  full ranked table and the pick is judgment anyway) × trivial → severity **low**

## L5-F6 — the §3.4 gate-survival table has no row for the §1.5 provenance/origin-marker + corroboration control — a gate the last two rounds made load-bearing (row 10's whole disposition) is absent from the artifact minted (R1-27) to stop rung migration silently shedding gates (first-raise; lineage-adjacent to R1-27/R2-5)

- location: §3.4 gate-survival table — rows enumerate "L1 permission profile / L2 hook
  fence + canary / Wrapper controls (preflight, ledger, idempotence, resume cap,
  snapshots) / L5 git-PR gate / L6 disable-model-invocation / Per-run budget"; §1.5 —
  "the wrapper stamps an origin marker file ... into EVERY run dir it creates."
- problem: origin stamping, the window log, the memory-surface snapshot diff, and the
  harvest-side corroboration gate are nowhere in the table's row set (the wrapper row's
  parenthetical omits them). At rungs 2–4 the wrapper is PARTIAL/NO — so sleeper runs
  there produce markerless dirs, and if their artifacts reach the local corpus (rung
  3/4 output arrives via human-merged branches; rung 2 runs locally), next-morning
  harvest counts them NON-sleeper: the poisoning gate is void at exactly the rungs the
  table exists to interrogate. §3.4's "adopting them means re-deriving the stack"
  clause bounds the risk, but the table IS the checklist that re-derivation reads — an
  absent row is a gate the adoption stub will not know to re-derive (the R1-27 failure
  mode, one artifact deeper).
- required_fix: add a "Provenance/origin-tag + corroboration gate (§1.5)" row (YES at
  R0/R1 via the wrapper; NO at R2–R4 absent rebuild) and name it in the
  graduation-grade adoption requirement.
- pattern refs: gate-stack rung non-portability; exhaustive-sweep-omits-hard-case.
- grading: low (rungs 2–4 adoption is itself human-gated and graduation-grade) × medium
  (poisoning-gate void if adopted) × trivial (one table row) → severity **low**

## L5-F7 — R3-12's recomputed bound treats cap-trip as terminal, but the monthly cap RESETS: cap-skip pauses death accrual at ~night 10 and the HALT lands early the NEXT month, so the worst-case burn on a deterministic cause is ~one cap PLUS the month-2 nights before HALT (~$55–60), and "whichever comes first" is not a stopping point (candidate supersedes: R3-12)

- location: §3.4 — "the $50 monthly cap trips after ~10 ceiling-priced nights (not ~3),
  and the per-cause HALT (initial fire + 3 resumes = 4 nights per dir, × M=3 dirs = 12
  nights) lands at night 12 OR at the cap, whichever comes first — the bounded-by-the-cap
  conclusion survives."
- problem: recompute at the ceiling: deaths land nights 4 and 8; the ledger preflight
  skips from ~night 11 (month-to-date ≥ $50 after 10 × $5); nights 11–30 are cap-skips,
  during which NO deaths accrue, so the third same-signature death — the HALT trigger —
  cannot occur in month 1. Month 2 the ledger resets and dir 3's remaining resume
  attempts burn until death 3 fires HALT (~$5–10 into month 2). Total worst-case burn
  ≈ $55–60 across two months, and there IS no "whichever comes first" race — the cap
  always pre-empts the HALT at ceiling pricing, then un-pre-empts itself at month
  rollover. The bounded conclusion survives (≈ one cap + ε); the printed trip-point
  race and the implied single-month bound are wrong as computed — the same
  repair-introduced-arithmetic class R3-12 itself fixed.
- required_fix: one clause — at ceiling pricing the cap trips first (~night 10),
  death accrual pauses with it, and the HALT lands early in the following month;
  worst-case deterministic-cause burn ≈ one monthly cap + ≤2 nights of the next month.
- pattern refs: unreconciled-numeric-floors (the reconciliation itself mis-composes —
  recompute, don't re-read); controller lookahead (a control gated on a monthly counter
  interacts with a detector gated on event counts).
- grading: certain (arithmetic) × trivial (conclusion survives; figures off) × trivial
  → severity **low**

## L5-F8 — CHANGELOG round-3 enumeration lists 16 of 17 gaps under an "all 17 addressed" header: R3-8 has no entry (the report body DOES address it at layer 4 (iv), layer 5, and row 4) — a partial recurrence of the round-2 change-summary desync blue owned (process; trivial)

- location: `blue/CHANGELOG.md` Round 3 — "All 17 round-3 gaps addressed additively"
  followed by per-gap bullets naming R3-1..R3-7 and R3-9..R3-17 (R3-2, R3-14, R3-15,
  R3-7+R3-4, R3-3, R3-5, R3-11+R3-6, R3-13, R3-1, R3-9, R3-10, R3-12, R3-16, R3-17);
  no R3-8 bullet. Report §7 round-3 update does enumerate R3-8 as executed, and the
  body carries the R3-8 edits ("bounded HONESTLY round 3 (R3-8)" — §4.3 layer 4 (iv);
  "Sensing scope stated honestly (R3-8)" — layer 5; §6 row 4 leg (b)).
- problem: the change-summary channel again runs behind the living report — one gap
  this time, not a whole round, and the report is the artifact of record, but this is
  the second consecutive round the channel under-reports, and lens seats were misled by
  exactly this channel in round 2.
- required_fix: add the R3-8 bullet to CHANGELOG Round 3 (content already exists in
  §7's list — a copy edit).
- pattern refs: change-summary desync.
- grading: certain × trivial (report body complete; navigation hint only) × trivial
  → severity **trivial** (process note)

---

## Lens self-report

- Round-3 repairs spot-checked in place and found SOUND (no finding): R3-7 chain-compare
  (crash-without-end-snapshot is covered by compare-vs-last-recorded = dead run's START,
  spanning the window); R3-2 `--manual` idempotence composition (a manual run
  date-suppresses that night's scheduled fire — correct by design); R3-10 §0 enumeration
  (now total over rounds 0–3 build artifacts); R3-11 normalization template (specified,
  zero-firings telemetry present, alternating-cause residual owned); R3-16/R3-17
  (verified corrected in place); R3-1 fenced-block fallback (fail-closed, both channels).
- Leaf work this pass: `git format-patch -1 -o /tmp/...` run live (exit 0, out-of-repo
  file created) backing L5-F2; all other findings are textual/logical, verified against
  the full report text.
- friction: none — template and protocol fit the material this round.
