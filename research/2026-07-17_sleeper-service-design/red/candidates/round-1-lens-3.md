# Red round-1, lens 3 — leaf-node citation verification, slice 3 of 4

Scope: sections divided evenly among 4 instances; slice 3 = **§4 (H4 consent gates)** and
**§5 (H5 cost discipline)**, plus the footnote definitions those sections reference
(HeadlessDocs, CliReference, PermissionsDoc, HooksDoc, ScheduledTasks, RoutinesDocs,
McpHeadlessBugs*, PermAskBypass, DenyRWIssue, DenyBashIssue, UsageAPI, ConsoleLimits,
Pricing, AIScientist, AIControl, OWASP, STOP, DGMSakana, CostRecord, FrictionRun4, Backlog,
EffReport, EfficiencyPlan, ResearchCommand, SemanticConsent, PushGuard, HooksJson,
RedPatterns, IdeasCorpus, HeadlessProbe*). Full report read whole (2 consecutive windows,
1111 lines). Citation ledger was empty at start (header only); every verification below
appended. (*McpHeadlessBugs and HeadlessProbe are §3-anchored — left to slice 2 except
where §4/§5 lean on them.)

Method: leaf-node fetch of every external reference (12 WebFetch targets + 1 gh thread
fetch); pinned-repo artifacts checked by direct read/grep against the working tree
(= pin `7bc501e`; no drift observed on any quoted line). MUST-try honored: STOP figures
extracted via ar5iv HTML after the abs page lacked them; #22055 comment thread reached via
`gh issue view --comments` after WebFetch returned body-only.

## Verification summary

The two sections are, on the whole, in very good citation health. 30+ statement↔reference
pairs verified at HIGH, including every load-bearing quote: the permissions-doc enforcement
note, dontAsk auto-deny, cross-level deny supremacy, Edit-not-Write startup warning
(2.1.210), Read-deny-blocks-Edit (2.1.208), subprocess non-coverage, blocking-hook
precedence, `//c/` Windows normalization, `-p` trust-dialog behavior, `Bash(git push *)`
syntax (the doc's own deny example — valid as written), issue statuses #22055
(closed-not-planned) / #6631 (closed, re-confirmed broken at v1.0.93) / #25621
(closed-duplicate), the DGM marker-removal and lineage quotes (verbatim), the AI-Scientist
self-relaunch + timeout-extension incidents, OWASP Excessive Agency mitigations, AI Control
abstract, all nine RoutinesDocs quotes, the disable-model-invocation plain-text clause
(v2.1.196+), hooks agent_id/agent_type + exit-2 deny, Usage/Cost Admin API details
(Admin-key-only, ~5-min freshness, 1/min), --max-budget-usd/--max-turns verbatim,
$414.97/42-agents/1975-api-turns/99%-cache, run-3 $149.95, the "dwarfing every lever"
quote, the semantic-consent clause, the push-guard contract comment, the hooks.json
bootstrap guard, and the run-4 friction specimens (4 seat classes gap-memory-unreadable;
the "MUST-try clause has no observable" line; the ABORT DISCLOSURE). Two claims the report
itself had flagged for re-pin are now settled UP: STOP percentages (0.42%, CI 0.31–0.57%;
0.46% with warning, CI 0.35–0.61%; 10,000 samples; syntactic `use_sandbox=False`/`exec(`
detection — ar5iv §6.2/Table 2, matches the report exactly → open question 8 resolvable)
and the #22055 community workaround (PreToolUse exit-2 protected-files hook, verbatim in
thread; the chmod-444 half of blue's phrase, however, appears in the thread only inside a
user's settings snippet — not as a recommended workaround; that sub-claim stays LOW).

## Findings

### L3-F1 — `disableBypassPermissionsMode` misplaced in the §4.2 profile JSON (MEDIUM)

- **Location:** §4.2 "Write surface and the permission profile", the sample
  `sleeper-permissions.json`: the line `"disableBypassPermissionsMode": "disable"` sits at
  the TOP level of the JSON, as a sibling of `"permissions"`, and the prose claims it
  "closes the escalate-to-bypassPermissions route".
- **Evidence:** the permissions doc places it INSIDE the permissions object: "set
  `permissions.disableBypassPermissionsMode` or `permissions.disableAutoMode` to
  `"disable"` in any settings file" (code.claude.com/docs/en/permissions, fetched
  2026-07-17). A key at the wrong level is silently ignored — the lockout the prose
  advertises never engages, and no warning is documented for unknown top-level keys.
- **Why it matters:** this is §4.2's own game — "Drafting details a naive spec would get
  wrong." The route-around table (layer 1) cites this setting as the close for the
  bypass-mode escalation; as specced it is a no-op.
- **Grading:** likelihood HIGH (the file as written misconfigures), impact MEDIUM (one
  belt among several; dontAsk + hook + deny set still stand), complexity TRIVIAL (move the
  key inside `permissions`). Also consider adding `disableAutoMode` alongside it — same
  doc line, same class of escape hatch.
- **Required fix:** correct the JSON; re-quote the doc's `permissions.`-prefixed form.

### L3-F2 — "no programmatic quota introspection" is partially contradicted by a live Rate Limits API (MEDIUM)

- **Location:** §5.1, "Console monthly spend limits exist at organization and workspace
  level — but they are **UI-only; the Admin API has no endpoint to read or set workspace
  spend/rate limits**, so pre-launch programmatic quota introspection is not available"
  (and the H5 verdict line "CONFIRMED-NEGATIVE on programmatic quota introspection").
- **Evidence:** platform.claude.com/docs/en/manage-claude/rate-limits-api (fetched
  2026-07-17) documents `/v1/organizations/rate_limits` and
  `/v1/organizations/workspaces/{id}/rate_limits` — Admin-key, READ access to configured
  org and workspace rate limits (requests/tokens per minute). It is read-only ("Can I
  update rate limits with this API? No.") and does NOT cover monthly spend limits; the
  workspaces doc still shows spend limits as Console-Limits-tab only.
- **So:** the "read ... rate limits" half of the denial is false as of today (live-source
  drift past the cited quickstarts#371 feature request); the "spend limits UI-only,
  no read/set" half stands, and the operative design conclusion survives on TWO grounds
  the report already names (Admin API unavailable to subscription auth AND to individual
  accounts). The falsifier verdict "CONFIRMED-NEGATIVE" is too strong as phrased.
- **Grading:** likelihood CERTAIN (doc contradicts the sentence), impact LOW-MEDIUM
  (conclusion survives; the sentence and the H5 stamp don't), complexity LOW (narrow the
  claim: spend limits have no API read/set; rate limits are now API-readable, read-only,
  Admin-key/org-only — still nothing a subscription-auth local scheduler can poll).
- **Required fix:** requalify the §5.1 bullet and the H5 verdict line; update
  [^ConsoleLimits] to cite the Rate Limits API page.

### L3-F3 — [^CliReference] mislabels `--fallback-model` as print-only (LOW)

- **Location:** §5.1 "**`--model` / `--fallback-model`** (print-only)" and the
  [^CliReference] footnote listing "`--fallback-model` (print-only)" among flags "quoted
  verbatim".
- **Evidence:** the current CLI reference description of `--fallback-model` carries no
  "(print mode only)" marker and documents a persistent `fallbackModel` setting
  ("To persist a chain across sessions..."). `--max-budget-usd` and `--max-turns` DO carry
  the marker — the label migrated onto a third flag. Attempt line: full flag description
  fetched verbatim from code.claude.com/docs/en/cli-reference 2026-07-17.
- **Grading:** likelihood CERTAIN, impact TRIVIAL (design uses it in print mode anyway;
  if anything the flag being broader helps), complexity TRIVIAL. Citation-fidelity fix
  only — but "quoted verbatim" footnotes must be verbatim.

### L3-F4 — pricing footnote can be upgraded from MEDIUM, with corrections (LOW)

- **Location:** §5.2 "List-rate reference points ... Haiku-class ~$1/$5 per MTok in/out,
  Sonnet-class ~$3/$15, frontier ~$10/$50 ... graded MEDIUM pending leaf fetch" and
  [^Pricing] "VOLATILE — re-fetch at citation-verification".
- **Evidence (leaf fetch of platform.claude.com/docs/en/about-claude/pricing,
  2026-07-17):** Haiku 4.5 $1/$5 CONFIRMED; Sonnet 4.5/4.6 $3/$15 CONFIRMED (Sonnet 5
  intro $2/$10 through 2026-08-31, then $3/$15); "frontier ~$10/$50" is true ONLY of
  Fable 5/Mythos 5 — the Opus 4.5–4.8 line is $5/$25. Batch API flat 50% CONFIRMED
  ("50% discount on both input and output tokens"); cache read 0.1x base CONFIRMED
  (~90% off). The "≤24h async turnaround" sub-claim is NOT on this page (it lives on the
  batch-processing page, not fetched — stays MEDIUM). Two additional facts material to
  §5.2's estimates: the Fable/Mythos/Sonnet-5 tokenizer produces ~30% more tokens for the
  same text (per the same page), and batch multipliers stack with caching.
- **Grading:** likelihood CERTAIN, impact LOW (the ~-qualified figures were honest;
  "frontier" should name Opus-class $5/$25 vs Fable-class $10/$50), complexity TRIVIAL.
- **Required fix:** pin the footnote to the leaf-fetched figures; correct the frontier
  line; either fetch the batch-processing page for the 24h window or keep that sub-claim
  MEDIUM-labeled.

### L3-F5 — mirror line count off by one (LOW, nit)

- **Location:** §1.3 input-inventory row (footnote block owned by this slice via
  [^RedPatterns]): "this run's mirror is 1,558 lines / 30+ named patterns".
- **Evidence:** line-anchored Grep count of `inputs/red-gap-patterns.md` = **1,557**
  lines. Recompute-don't-restate discipline; trivial.

### L3-F6 — deny-rule anchoring depends on process cwd, unstated (LOW-MEDIUM)

- **Location:** §4.2, the deny set written in bare-relative form (`Edit(.claude/**)`,
  `Edit(plugins/**)`, `Edit(.github/**)`, ...) inside an "operator-owned settings file
  outside the repo working directory" passed via `--settings`.
- **Evidence:** permissions doc pattern-anchor table (fetched 2026-07-17): bare `path` /
  `./path` rules anchor at the **current directory**; only `/path` anchors at the settings
  source (for a `--settings` file, the file's own directory) and `//path` at the
  filesystem root. So the §4.2 deny set protects the repo only when the scheduled process
  cwd IS the repo root. A Task-Scheduler recipe whose wrapper forgets to cd (default cwd
  is System32) silently relocates the entire relative deny set — layer 1 evaporates with
  no warning, precisely in the unattended mode it exists for. (Gitignore bare-filename
  depth-matching softens some rows but not the anchored ones like `.claude/**`.)
- **Grading:** likelihood LOW-MEDIUM (the wrapper is scripted and §3.2 does say "run
  non-bare explicitly from the repo root" — but that clause is about context loading, and
  the bare+--settings recipe is the preferred variant), impact MEDIUM (silent loss of
  layer 1; layer 2's compiled-in hook still covers), complexity TRIVIAL (write guardrail
  deny rules `//`-absolute or `~/`-anchored — the profile already does this for
  `~/.claude` — and state the cwd precondition in scheduling.md next to the trust-dialog
  precondition).

## Notes for merge (not findings)

- STOP figures: report matches ar5iv exactly; open question 8 can be marked resolved-at-
  ar5iv (PDF pin still outstanding if the lead wants publisher-PDF grade; arxiv-latex MCP
  not exposed at this seat).
- #22055 workaround sub-claim ("chmod 444 + PreToolUse hook ... in thread") SPLIT:
  the PreToolUse exit-2 protected-files script is in the thread verbatim (HIGH); chmod
  appears only inside a user's posted settings snippet, and no chmod-444 recommendation
  was found in the full `gh --comments` thread (LOW) — §4.3 layer 3's "chmod-readonly
  ... documented community pattern" leans on the LOW half. Concurs with the parallel
  lens-4 ledger line (L4-F2); merge should collapse these to one finding.
- §4.1's "closures invert the fix story" framing is well-supported by all three issue
  fetches; the report's own caveat ("behavior in current 2026 builds unverified") is
  accurate and honest.
- [^Backlog] cited items (15–17 smoke ~50k, 34 HTTP daemon, 39 batching) matched at line
  level in `ideas/backlog.md`; long lines elided by the harness — content spot-check
  MEDIUM-HIGH, no discrepancy observed. Item-level deep read left to the slice owning §1/§3.

## Friction

- arxiv-latex MCP not exposed at this lens seat (ToolSearch: no match) — the run-4
  standing class recurs; ar5iv WebFetch sufficed for STOP this time, but the
  table-faithful path remains unavailable where the MUST-try clause binds.
- The permissions-doc WebFetch answer overflowed the return channel (54KB persisted to a
  tool-results file) — workable via Read, but a page-scoped fetch prompt cannot prevent
  it; fine as-is, noting the detour cost one extra read.
