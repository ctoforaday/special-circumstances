# red round 3 — lens 2 (leaf-node citation verification; slice 2 of 4: §2 + §3 + their footnote definitions)

Full report re-read whole (3 consecutive Read windows, lines 1–1641). Ledger consulted first;
prior-round HIGH claims in unchanged text carried (all access dates are today, so no
calendar-drift term). Blue changed §2.2/§2.3/§2.4/§3.2/§3.3/§3.4 this round (R2-1/3/4/8/9/
10/12/15/16/17/21) — every repair treated as a NEW claim and leaf-verified.

## Verified clean this round (leaf level)

- **R2-3 execution locus (§2.2 step 4, §4.3 layer 4, [^ResearchCommand])** — read the shipped
  plugin copy `frank-exchange-of-views/0.7.0/commands/research.md` AND
  `git show 7bc501e:plugins/frank-exchange-of-views/commands/research.md` (byte-identical):
  step 2 = session-Bash `node .../setup-research-run.mjs`, step 3 = **Workflow** tool with
  `scriptPath` = `.../debate.js`, step 5 = session-Bash `node .../capture-research-run.mjs`.
  Blue's MIXED-locus statement is exact. Also re-verified in the same file: `--smoke` = 1
  lane / 1 round / haiku / ~50k tokens; "for keeper runs, omit `model` entirely";
  stop-and-resume standing practice; capture emits cost.md + run-record-audit.md;
  script-vs-prose quote verbatim. **HIGH.**
- **R2-1 flag (§2.2 step 0, §3.2, [^CliReference])** — `claude --help` run live on this box:
  `--input-format <format>` — "Input format (only works with --print): 'text' (default), or
  'stream-json'". Flag exists at the pinned CLI as claimed. **HIGH.**
- **R2-21 DGM honesty clause (§2.4, [^DGM])** — arxiv.org/html/2505.22954 fetched live:
  "Only agents that compile successfully and retain the ability to edit a given codebase are
  added to the DGM archive"; low scorers deliberately retained ("archived solutions can serve
  as stepping stones"); "Parent selection is roughly proportional to each agent's performance
  score"; "Each newly generated agent is quantitatively evaluated on a chosen coding
  benchmark." Every leg of the repair corroborates; the retained lead sentence ("evaluates
  every change empirically... before archive entry") is ordering-true. Repair clean, no
  regression. **HIGH.**
- **§3.2 [^HeadlessDocs] set re-fetched live (living source):** `--bare` future-default note,
  10MB stdin cap, bg-workflow wait cap 10min v2.1.182+/env var, `system/init`
  plugins+plugin_errors fail-CI guidance, `total_cost_usd` in json output, `--json-schema`
  invalid-schema error with pre-2.1.205 silent-ignore history — all present verbatim, zero
  drift. **HIGH.**
- Carried without re-fetch (verified HIGH r1/r2, same-day access, section text unchanged):
  [^IdeaStudy], [^SmokeRecord], [^EfficiencyPlan], [^Reflexion], [^ScheduledTasks]
  (disable-model-invocation), [^PortPlan] (pin-absent caveat stands, lead-docketed),
  [^McpHeadlessBugs] #76239/#68375 OPEN (re-confirmed r2 same day), [^WindowsHang],
  [^SlashHeadlessIssues], [^WebSandbox], [^GhaSchedule], [^MissedRun], [^QmdDaemon] ~973MB,
  [^QmdFallback], [^Backlog] smoke shape, [^HeadlessProbe] (medium,
  disposition-of-record: re-run at build).
- Observation, no gap filed: §3.3(a)'s "ToolSearch discovers deferred tools only from
  DECLARED servers" is uncited but analytic — tools exist only for configured servers, and
  the `--strict-mcp-config` half is HIGH-cited. Demanding a citation would add nothing.

## Findings

### L2-F1 — MEDIUM — the phase drive's structured-output leg is an undocumented flag composition asserted as mechanism-of-record

- **Location:** §2.2 step 3 — "The pick returns to the wrapper as structured output (the
  phase drive's `--json-schema` leg), because step 4's staging needs it wrapper-side."
  Also CHANGELOG R2-3 ("after the phase-A structured-output pick").
- **Evidence (attempt line):** headless docs fetched live — structured output is documented
  ONLY as "use `--output-format json` with `--json-schema`... structured output in the
  `structured_output` field." The phase drive necessarily runs `--output-format stream-json`
  (the `--input-format stream-json` pair — the report's own R2-1 text). Nothing on the
  headless page or in `claude --help` documents (a) `structured_output` under stream-json
  output, or (b) per-turn structured output in a multi-turn streaming drive — and the pick
  is a MID-drive phase result (phases continue after it), not the final result the
  `--json-schema` docs describe. Second extraction path tried: `claude --help` full text —
  `--json-schema` entry carries no composition statement.
- **Class:** sibling-repair composition (red memory): R2-1's stream-json drive and the
  retained structured-output pick-return each verify alone; their composition is unverified
  and undocumented. The report hedges its OTHER two round-2 mechanism edges (OQ20 native
  rule scoping, OQ21 deny provenance) but asserts this one flat.
- **Grading:** likelihood MEDIUM (undocumented ≠ broken, but mid-stream schema-bound output
  is exactly the kind of thing that only binds the final result); impact MEDIUM-LOW (fallback
  is trivial — parse the pick from the phase-A result text, or restructure phase A as its own
  `--output-format json` invocation at some cache/session cost); complexity LOW.
- **Required fix:** demote the `--json-schema` leg to verify-at-build with the fallback named
  (a new OQ beside OQ20/OQ21), or probe it now and cite the probe. Do not carry an
  undocumented composition as the mechanism the wrapper's staging "needs."

### L2-F2 — MEDIUM-HIGH — the `dontAsk` closed-world premise is refuted by the built-in read-only Bash carve-out (live doc, leaf-quoted)

- **Location:** §3.2 — "`--permission-mode` (incl. `dontAsk` — auto-denies anything not
  allow-listed; documented for 'locked-down CI runs')." Load-bearing dependents outside my
  slice (merge to place): §4.2 "the default is closed, so a channel we forgot is denied";
  §6 row 4 "the subprocess surface is the pinned read-only git argv set plus the Workflow
  tool"; §4.3 layer 2 "unforgeable by the session (no Bash beyond pinned git argv...)";
  §6 row 13's closure story.
- **Evidence (leaf quotes, both fetched live today):** headless doc: "`dontAsk` denies
  anything not in your `permissions.allow` rules **or the read-only command set**."
  Permissions doc #read-only-commands: "Claude Code recognizes a built-in set of Bash
  commands as read-only and runs them without a permission prompt **in every mode**. These
  include `ls`, `cat`, `echo`, `pwd`, `head`, `tail`, `grep`, `find`, `wc`, `which`, `diff`,
  `stat`, `du`, `cd`, and read-only forms of `git`. The set is not configurable; to require a
  prompt for one of these commands, add an `ask` or `deny` rule for it."
- **Consequences:** (a) under the §4.2 profile as printed, `Bash(cat
  //c/Users/gbloc/.claude/projects/...)` — the report's own named high-value exfil target,
  row 13 — is AUTO-APPROVED: the belt `Read(...)` denies match the Read tool, not Bash, and
  no Bash deny covers the carve-out set. The R1-17 read-scoping closure is therefore open on
  the Bash side, composable with the retained WebFetch(arxiv.org) egress. (b) The three
  `Bash(git status/diff/log)` allow rules are REDUNDANT (read-only git forms are
  auto-approved anyway) — evidence the carve-out was unmodeled, not judged. (c) The layer-2
  fired-record unforgeability leg cites "no Bash beyond pinned git argv," which is false;
  whether `echo ... >> <operator-owned fired-record>` survives read-only classification
  under redirection is exactly the OQ18 operator-semantics class and must join that build
  test. Write-fence impact is otherwise limited: exec-capable forms (`find -exec`, wrappers)
  still prompt (= deny under dontAsk), so this is a READ-surface breach, not a write breach.
- **Grading:** likelihood MEDIUM (carve-out is certain and non-configurable; exploitation
  requires the row-14 injection arm the report already grades Medium); impact MEDIUM-HIGH
  (row 13's own grade — box-local secret/transcript reads + narrow egress); complexity LOW
  (doc-sanctioned fix: `deny` rules over the carve-out set — deny precedence holds in every
  mode — or a bare `Bash` deny removing the tool entirely if the session can live without
  git, which the redundancy in (b) suggests it can).
- **Required fix:** correct §3.2's dontAsk sentence to carry the carve-out; re-derive §4.2's
  profile (enumerated Bash denies over the read-only set, or bare-Bash removal), §6 rows
  4/13, and §4.3 layer 2's unforgeability sentence on the true surface; add
  redirection-under-carve-out to the OQ18 test list.
- **Provenance note for the merge:** r1 graded the dontAsk sentence HIGH from the same pages;
  either live-source drift since the r1 fetch or an r1 under-read — the carve-out is on both
  pages TODAY, leaf-quoted above. Live-source-drift pattern applies either way.

### L2-F3 — LOW — R2-10's per-night burn arithmetic does not follow from the stated resume semantics

- **Location:** §3.4 — "a deterministic root cause that survives a fresh dir would otherwise
  burn up to k×$5/night until the monthly cap trips (~3 nights at the ceiling)."
- **Evidence:** internal recompute. The idempotence rule (§3.4) fires the wrapper once
  nightly and performs ONE resume attempt per fire ("on start... started-but-unrecorded →
  resume"); k=3 resumes accrue across three nightly fires, then "the next fire mints a fresh
  dir." Per-night worst case is therefore 1 × $5, not k × $5; a deterministic cause trips
  the $50 cap after ~10 ceiling-priced nights, not ~3, and the per-cause HALT (1 initial +
  3 resumes = 4 nights per dir × M=3 dirs = 12 nights) lands AFTER the cap would trip, not
  before. The "bounded by the cap" conclusion survives; both stated figures are wrong unless
  the wrapper retries k times WITHIN one invocation, which no text states.
- **Class:** repair-introduced arithmetic (recompute, don't re-read — red memory:
  unreconciled-numeric-floors / controller-pricing family).
- **Required fix:** either state in-night retry semantics explicitly, or recompute the
  sentence (≤$5/night; cap trips ~10 nights; HALT at night 12 or at cap, whichever first).

## Ledger appends

Written to `red/citation-ledger.md` under "## round 3 — lens 2".
