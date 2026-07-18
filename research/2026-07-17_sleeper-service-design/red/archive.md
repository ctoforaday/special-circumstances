# red/archive.md — closed-gap prose record (APPEND-ONLY; never edit an existing block)

Immutable full-prose records of closed gaps: what was found, how verified, closure class.
The ledger's closure index is the screen; this file is the evidence of record.

(no closed gaps yet — round 1)

## Round 2 closures (red merge, 2026-07-17)

### R1-1 — backlog item count fails recount (40 → 25) — CLOSED
Found round 1 (L1): §1.3/[^IdeasCorpus] claimed "40 statused items"; recount at pin
`7bc501e` = 25 checkbox items / 39 lines. Round-1 repair: body + footnote restated
"25 statused checkbox items across 39 lines ... recounted at the pin, R1-1". Verified
round 2 by L1 re-count via `git show` (25/39 reproduced) and merge grep: "40 statused"
survives only inside §7's propagation-grep token log. Closure class: **closed**.

### R1-2 — scope-fusion overstatement on the PDF-gap recurrence claim — CLOSED
Found round 1 (L1): fused "red, blue, AND judge across two consecutive runs". Repair
splits the sentence into backlog 27c (three seats, one run) and 31h (two runs, red
merges), labeled "two sources, stated separately (round-1 correction R1-2)". Verified
round 2 by L1 at the pin: line 27(c) and line 31(h) verbatim; line-scheme validated by
two independent anchors. Closure class: **closed**.

### R1-3 — [^DGM] homes a Sakana-post quote to the arXiv paper — CLOSED
Repair: "improve themselves the more compute they are provided" moved to [^DGMSakana];
"(ICLR 2026)" dropped; [^DGM] carries the correction note. Verified round 2 by L1
(footnote text inspected; r1 leaf reads of record: quote verbatim at sakana.ai/dgm,
absent from abs and /html). §1.2's dual-cite [^DGM][^DGMSakana] acceptable — the
carrying source is cited. Closure class: **closed**.

### R1-4 — [^SICA] venue metadata contradicted the cited page — CLOSED
Repair: venue restated per the abs page's Comments field ("Submitted as a preprint to
NeurIPS 2025"); workshop tag dropped. Verified round 2 by L1 live re-fetch: Comments
verbatim; 17–53% SWE-bench-Verified-subset figure re-confirmed. Closure class: **closed**.

### R1-5 — "three MCP issues leaf-checked OPEN" while #32191 was CLOSED-duplicate — CLOSED
The sharpest round-1 citation defect (false verification claim). Repair: §3.3 restates
"TWO leaf-checked OPEN (#76239, #68375) ... ONE historical: #32191 ... CLOSED as
duplicate (canonical untraced; 2.1.58–2.1.71 era)"; §6 row 8 and [^McpHeadlessBugs]
propagated. Verified round 2: L2 confirms the restatement matches ledger statuses; L4
re-fetched both OPEN issues live (drift check — both still OPEN 2026-07-17); merge grep
confirms no stale "leaf-checked OPEN" claim covering #32191 remains. Closure class:
**closed**.

### R1-6 — exit code "0/1" attributed to a page with no exit-code docs — CLOSED
Repair: §3.2 requalified — 0-on-success probe-corroborated; "on error the CLI exits
nonzero, but the cli-reference page publishes no exit-code table ... the wrapper treats
ANY nonzero exit as failure"; [^CliReference] carries the correction. Verified round 2 by
L2 on a fresh cli-reference fetch: still no exit-code table; `--max-turns` still says
only "Exits with an error." Closure class: **closed**.

### R1-8 — disableBypassPermissionsMode misplaced at top level (silent no-op) — CLOSED
Repair: key moved INSIDE the `permissions` object in the §4.2 sample (line "INSIDE
permissions — R1-8"), with prose explaining the silent-ignore failure. Verified at the
round-2 merge by direct read of the sample JSON against the r1 citation-ledger doc line
("permissions.disableBypassPermissionsMode", any scope). `disableAutoMode` correctly
deferred to OQ17 pending leaf-verify rather than asserted. Closure class: **closed**.

### R1-9 — "no endpoint to read" stale vs the live Rate Limits API — CLOSED WITH REGRESSION
Repair: §5.1 + H5 verdict requalified (spend limits: no API read/set; rate limits:
API-readable, read-only, Admin-key-only — nothing a subscription-auth scheduler can
poll); [^ConsoleLimits] amended; [^RateLimitsAPI] added. Verified round 1 at the leaf by
L3; round 2 L3 found the §6 row 5 likelihood cell still reads the pre-repair flat
"(no API)". Closure class: **closed_with_regression** → successor **R2-13**.

### R1-10 — [^CliReference] mislabeled --fallback-model print-only — CLOSED
Repair: label dropped; §5.1 bullet and footnote note the persistent `fallbackModel`
setting. Verified against the r1 leaf fetch (L3); no round-2 contradiction on the
re-fetched cli-reference page. Closure class: **closed**.

### R1-11 — "frontier ~$10/$50" true only of Fable/Mythos class — CLOSED WITH REGRESSION
Repair: §5.2 + [^Pricing] pinned to leaf figures, two frontier tiers named, tokenizer
+30% caveat added, Batch ≤24h honestly carried MEDIUM. Verified round 2 by L3 live
re-fetch: zero drift on all figures; the ≤24h sub-claim independently resolved HIGH on
the batch-processing page (bankable). Regression: §7's self-audit still says "pricing
figures graded MEDIUM" and its upgrade list omits R1-11 (L4). Closure class:
**closed_with_regression** → successor **R2-14**.

### R1-12 — bare-relative deny rules anchor at process cwd — CLOSED
Repair: all guardrail rules written `//`-absolute in the sample; anchoring physics
paragraph added (bare = cwd, `/` = settings source, `//` = root); wrapper-cd precondition
stated for scheduling.md beside the trust-dialog one. Verified at the round-2 merge by
direct read of §4.2 against the r1 citation-ledger anchor-table line. Closure class:
**closed**.

### R1-13 — chmod-444 dressed as a documented community pattern — CLOSED
Repair: [^PermAskBypass] splits the claims — PreToolUse exit-2 protected-files workaround
kept as thread-verbatim; chmod-readonly recast as "a DESIGN-PROPOSED measure of this
report's own, not community-sourced" in §4.3 layer 3. Verified at the round-2 merge by
direct read of layer 3 + footnote against the r1 full-thread gh check (L4/L3 concurring).
Closure class: **closed**.

### R1-14 — marginal value over the manual process never argued — CLOSED WITH REGRESSION
Repair: §1.4 gained the priced null-alternative paragraph — what automation buys (bounded
research + structured stubs, mechanical recurrence arithmetic, no single-point recall),
what it costs, rung 0 as DEFAULT and possibly terminal, cadence as hypothesis with a
named triage-rate revisit trigger. Verified by full read; the demanded honest paragraph
exists and engages the report's own Dependabot evidence. Regression: §3.4's ladder still
stamps rung 1 "RECOMMENDED default" unqualified (L5). Closure class:
**closed_with_regression** → successor **R2-15**.

### R1-15 — preferred --bare recipe made the repo copy the executing copy — CLOSED WITH REGRESSION
Repair: `--plugin-dir` pinned to an operator-owned, read-only copy outside the write
surface, "NEVER the repo's `plugins/` working tree"; §4.3 layer 2's route-around updated.
Verified by full read (§3.2 + layer 2 consistent). Regression: the new copy's lifecycle
is unbuilt — no creation/refresh owner, "the human-gated plugin update path" defined
nowhere, preflight asserts load not version (L5+L6 independently). Closure class:
**closed_with_regression** → successor **R2-8**.

### R1-16 — Bash(node scripts/*): write-capable channel, layout mismatch — CLOSED WITH REGRESSION
Repair: the allow rule deleted; harvest.mjs moved wrapper-side pre-session; session Bash
= pinned-argv read-only git; compound/redirection/traversal matching named as build-PR
test (OQ18). Verified by direct read of the §4.2 sample + §1.3 + merge grep (`Bash(node`
survives only in correction contexts). Regressions: §2.2 step 2 still reads "SCORE
(script...)" in-session (L5); step 4's FEOV invocation never reconciled with "removes ALL
script execution from the session" — the load-bearing composition defect of the round
(L5). Closure class: **closed_with_regression** → successors **R2-3, R2-4**.

### R1-17 — read/exfiltration channel never considered — CLOSED
Repair: Read/Grep/Glob allow-scoped to repo + plugin copy; belt Read-denies on
`~/.claude` settings/credentials/session transcripts; WebSearch dropped from the nightly
profile; §6 row 13 added with an argued RISK-ACCEPT (single-operator box, largely-public
repo, narrow enumerable session-logged residual egress). Verified by direct read of the
profile JSON + row 13. The acceptance is argued as demanded — red accepts the argument.
Closure class: **closed**.

### R1-18 — injection via the retrieval surface had no risk row — CLOSED
Repair: §6 row 14 added (graded, disposed); stub contract gained `origin-labels`
(web-derived = untrusted-origin, barred from ranking inputs); WebSearch nightly drop
shared with R1-17; §2.2 step 4 names the narrowed nightly retrieval surface. Verified by
direct read. Closure class: **closed**.

### R1-19 — monthly ledger inside the loop's write surface (invariant 4) — CLOSED WITH REGRESSION
Repair: ledger relocated to `~/.claude/sleeper/ledger.jsonl`, wrapper-written outside the
session, fail-closed on unparseable ledger; invariant-4 layer for the monthly guard named
(the wrapper); §0 invariant 5 updated. Verified by direct read of §5.1 + §2.2 step 7.
Regression: the gap's own text named the idempotency-gate sibling ("Same shape, lower
stakes"), which got neither fix nor argued acceptance — completeness signal and DEAD
marker location remain loop-writable/unstated (L5). Closure class:
**closed_with_regression** → successor **R2-12**.

### R1-20 — H3 "Confidence HIGH" attached to the untested bare path — CLOSED
Repair: verdict split — "HIGH for the probed non-bare path; OPEN for the preferred
`--bare --plugin-dir` recipe", with "A verdict stamp must not grade higher than its
recommended configuration's evidence." Verified round 2 by L2 (CLEAN). Closure class:
**closed**.

### R1-21 — row 4 accepted on the benign-actor premise §4.1 rejects — CLOSED WITH REGRESSION
Repair: row 4 re-argued explicitly WITHOUT actor intent — legs: (a) channel narrowness
(pinned git argv only), (b) detection outside the model's reach (R1-26 snapshots),
(c) L×I×Cx vs Windows sandbox complexity; "No premise about actor intent is used."
Verified by direct read — the re-argument is performed as demanded. Regression: leg (a)'s
premise is contested by the unreconciled step-4 FEOV execution locus (if a workflow
runner executes scripts outside the Bash surface, "the only subprocess reachable is the
pinned git set" is false), and leg (b) leans on a snapshot channel with no defined reader
(see R2-7). Closure class: **closed_with_regression** → successor **R2-3** (carries the
re-argument obligation; R2-7 covers leg (b)).

### R1-22 — "stubs age out visibly" was policy-without-mechanism — CLOSED WITH REGRESSION
Repair: aging specified — skip only stubs younger than 30 days; older untriaged stubs
auto-stale via harvest's filename-date sweep and re-enter the docket flagged
`stub-stale`; §2.3 status field carries the lifecycle; §6 row 3 updated. Verified by
direct read of §1.4 stage 2 + §2.3 + row 3. Regression: the window composes badly with
human graduation latency — a graduation-queued good stub is re-researched every 30 days
(L6). Closure class: **closed_with_regression** → successor **R2-11**.

### R1-23 — Batch API as a lever with no routing mechanism — CLOSED
Repair: demoted to an explicit FUTURE note ("if and when a routing mechanism exists, and
not before"). Verified by direct read of §5.2. Round 2 bonus: the ≤24h window sub-claim
resolved HIGH at the batch-processing page (L3). Closure class: **closed**.

### R1-24 — layer-6 row asserted what OQ3 admits is unverified — CLOSED
Repair: the qualifier carried inline in the table row ("doc-verified for SCHEDULED fires;
the in-session hostile-invocation case is exactly open question 3"), fallback layers
named in the route-around cell. Verified by direct read of §4.3 layer 6. Closure class:
**closed**.

### R1-25 — the loop authors its own future harvest inputs — CLOSED WITH REGRESSION
Repair: §1.5 extended to authorship — harvest tags `*_self-improve/` rows
`origin: sleeper`, caps sleeper-origin recurrence at 1 per class, requires non-sleeper
corroboration (or a `sleeper-only` flag) before the pick; §6 row 10 re-scoped and
re-graded; §4.3's "eats its own complaints" now carries the cap. Verified by direct read.
Regressions: (1) the tag is keyed on a dir-NAME glob and the loop's own spawned FEOV
sub-run dir escapes it — worse, the untagged sub-run friction satisfies the
"non-sleeper occurrence" gate, self-supplying the demanded corroboration (L6); (2) the
"requires X — or, failing that, flags" disjunction is undecided between a hard bar and a
labeled pass, and each horn has an unowned cost (L5+L6). Closure class:
**closed_with_regression** → successors **R2-5, R2-6**.

### R1-26 — uncommitted-working-tree tamper evidence was self-erasing — CLOSED WITH REGRESSION
Repair: wrapper takes out-of-session snapshots (git porcelain + SHA-256 guardrail hashes)
at run start and end, appended to an operator-owned log outside the repo; §4.3 layer 5
restates the durable-tamper-evidence property; heavier auto-commit alternative named.
Verified by direct read of layer 5 + §2.2 steps 0/7. Regression: the snapshot log has no
defined reader or comparison step — the same undefined-receiver defect R1-29 fixed for
loop death, unfixed here; row 4's leg (b) presupposes the missing watchman (L5+L6).
Closure class: **closed_with_regression** → successor **R2-7**.

### R1-27 — gate stack sold as rung-portable — CLOSED WITH REGRESSION
Repair: per-rung gate-survival table added to §3.4 (6 controls × 5 rungs), rung-4
settings-inversion named, "climbing past rung 2 is itself a graduation-grade decision —
never a config toggle"; §6 row 16 added. Verified by direct read; L2 cross-checked the
table's cells against ledger-verified [^RoutinesDocs]/[^ScheduledTasks] facts — no cell
contradicts a verified source. Regression: the rung-0 column — the design's own
newly-default mode — holds the soft cells: compound L2 row claims YES for a canary that
needs the absent wrapper, and manual-run spend never enters the monthly ledger, unstated
(L5). Closure class: **closed_with_regression** → successor **R2-16**.

### R1-28 — write-fence marker-keyed, fails OPEN on marker loss — CLOSED WITH REGRESSION
Repair: step-0 denial canary added (attempt one out-of-fence write, abort unless
DENIED); §0 tree, §2.2 step 0, §4.3 layer 2 all carry it; OQ2 reframed
verify-enforcement-not-presence; OQ16 (polarity inversion) added. Verified by direct
read. Regressions — the canary as specified does not deliver the closure it stamps:
(1) actor/observer/abort mechanism unspecified, with the step-0 header ("wrapper,
OUTSIDE the model session") contradicting the "First in-session action" label — as
printed it is either a post-hoc check or an instructional abort of the class §4.1
refutes (L5); (2) layers 1 and 2 fence the identical boundary, so a fence-dormant run
with live permissions PASSES the canary — it cannot isolate the layer it certifies, and
OQ2's own acceptance test fails as designed (L5+L6 independently). Closure class:
**closed_with_regression** → successors **R2-1, R2-2**.

### R1-29 — resume-forever livelock + undefined receiver of "loudly" — CLOSED WITH REGRESSION
Repair: resume cap k=3 → DEAD marker + dated abort record + fresh dir; dead-man surface
defined (wrapper-maintained `last-successful-run`; doctor reports "last successful
sleeper run: N days ago"); §6 row 15 added. Verified by direct read of §3.4 + row 15.
Regressions: (1) the reader is pull-only and "a surface the human already looks at" is
asserted, not evidenced — the automation's purpose removes exactly that habit (L6);
(2) the cap bounds per-DIR, not per-CAUSE — a deterministic wedge burns ~k×$5 nightly
via fresh dirs until the monthly cap trips, resetting monthly (L6). Closure class:
**closed_with_regression** → successors **R2-9, R2-10**.

### R1-30 — gap-pattern mirror line count off by one — CLOSED
Repair: corrected to 1,557 in §1.3 and [^RedPatterns] with the byte-exact recount note
(final byte 0x0a). Verified round 2 by L1 (`wc`: 1,557 lines / 119,418 bytes) and merge
grep: the only surviving "1,558" is §7's propagation-grep token log. Closure class:
**closed**.

