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


## Round 3 closures (red merge, 2026-07-17)

### R1-7 — inputs/PINNED.md pins a path absent at the pin — ADJUDICATED (risk_argued)
Standing since round 1; carried with the LEAD through round 2. LEAD ruling (debate.md round-2
LEAD section): **risk_accepted**. The finding is valid and verified (`git cat-file -e` MISSING for
`plans/claude-port-plan.md` at `7bc501e`), but the fix (setup-script pin validation / staging
the port plan into inputs/) is run-infrastructure owned by the lead, not report text blue can
change; the affected quotes were re-verified verbatim at working-tree `6df52af` (r2 L1),
re-confirmed 6df52af-clean this session (r3 L1), so residual impact is a reproducibility
caveat already disclosed in [^PortPlan]. Recorded, not dropped: port-plan citations REMAIN
snapshot-grade in the final report; the lead owes pin validation before the next run.
Excluded from red's verdict pool per the run directive. Closure class: **risk_argued**.

### R2-1 — canary actor/observer/abort mechanism unspecified — CLOSED WITH REGRESSION
Found round 2 (L5, superseding R1-28): the step-0 canary had no mechanism connecting the
in-session write attempt to the wrapper's abort, and the header/label contradicted. Round-3
repair: blue specified the two-phase stream-json drive (`--input-format stream-json
--output-format stream-json`, flag pair leaf-verified by L2 in `claude --help` on CLI 2.1.212,
HIGH) plus a positive fired-record — the wrapper drives a canary prompt, parses the deny
event, confirms the sleeper-guard fired-record carries the nonce with decision=deny, only then
prompts real work. Verified round 3 (L2 flag HIGH; L6 mechanism sound). Regression: the pick
returns "as structured output (the phase drive's `--json-schema` leg)" — an undocumented
mid-drive schema-bound composition (structured output is documented only for `--output-format
json` final results, L2); and three behavioral legs of the drive plus two invariant-7 staging
edges are unprobed/unstated (L5). Closure class: **closed_with_regression** -> successor **R3-1**.

### R2-2 — the deny-outcome canary cannot isolate the hook fence — CLOSED
Found round 2 (L5+L6, superseding R1-28): layers 1/2 fence the identical boundary, so a
fence-dormant run passes the deny-outcome canary. Round-3 repair implements exactly the LEAD's
directed fix: a POSITIVE hook-liveness record — the sleeper-guard hook writes a fired-record
(nonce + decision=deny) that the wrapper confirms non-empty per run; marker loss -> no
fired-record -> ABORT (the H4-refuted permissions-only configuration now fails closed). The
smoke test (report lines 1211-1213) verifies the fence-dormant run aborts as designed —
resolving the R2-2 contradiction between OQ2's acceptance test and the claim. Verified round 3
by direct read (§4.3 layer 2, lines 424-427) + L6. Closure class: **closed**.

### R2-3 — step-4 FEOV locus never reconciled with "removes ALL script execution" — CLOSED
The round-2 HIGH (L5, superseding R1-16, R1-21). Round-3 repair: blue determined the FEOV
execution locus BY PROBE and stated the MIXED shape — session-Bash `node setup-research-run.mjs`
(step 2) and `node capture-research-run.mjs` (step 5), with `debate.js` executed via the
**Workflow** tool (harness-side runner, hook-scoped `scriptPath`), outside the Bash permission
surface. Verified round 3 (L2): read the pinned `research.md` (shipped 0.7.0 copy AND
`git show 7bc501e:` byte-identical) — step 2/3/5 loci exact; `--smoke` = 1 lane/1 round/haiku/
~50k tokens; capture emits cost.md + run-record-audit.md — the mixed-locus statement is exact,
HIGH. L6 concurs sound. The row-4 "read-only git argv" premise is separately challenged by the
`git log --output` write gadget — a NEW finding (R3-15), not an R2-3 residual. Closure class:
**closed**.

### R2-4 — §2.2 step 2 still executed a script in-session — CLOSED
Found round 2 (L5, superseding R1-16): propagation miss. Round-3 repair: step 2 reworded to
the wrapper-side architecture (read the staged scored table / model-side language-tier scoring,
no session script). Verified round 3 by L5 spot-check in place. Closure class: **closed**.

### R2-5 — authorship cap keyed on a dir-NAME glob — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-25): the loop's spawned sub-run landed untagged friction that
satisfied the corroboration gate. Round-3 repair: §1.5 rewritten to tag by CAUSAL PROVENANCE —
the wrapper stamps `inputs/.sleeper-origin` (carrying the parent run id) into EVERY run dir it
creates, and harvest additionally treats any dir whose creation timestamp falls inside a
wrapper-logged sleeper window as sleeper-origin even if markerless. Verified round 3 (L1
leaf-read `setup-research-run.mjs`: the wrapper choosing runDir + stamping is buildable as
claimed, HIGH). Regressions (three surfaces the marker/dir keying does not reach):
(1) at rung 0 — the DESIGN'S DEFAULT mode — there is no wrapper, so both manual run dirs land
markerless and outside any ledger window (L5 -> R3-2, jointly superseding R2-6, R2-16);
(2) the red gap-pattern MEMORY mirror is a separate surface with no run dir, so nightly-authored
patterns enter harvest un-tagged as "non-sleeper" (L6 -> R3-3); (3) the ledger-window fallback is
void for DEAD runs (no step-7 ledger write -> no logged window -> no porcelain sweep), letting a
model-authored in-session dir on a dying night escape markerless (L1 -> R3-4). Closure class:
**closed_with_regression** -> successors **R3-2, R3-3, R3-4**.

### R2-6 — undecided "requires corroboration — or, failing that, flags" disjunction — CLOSED WITH REGRESSION
Found round 2 (L5+L6, superseding R1-25). Round-3 repair: blue decided the branch — a
severity-gated bypass (infrastructure-failure classes enter flagged `sleeper-only` WITHOUT
corroboration; ordinary classes require a non-sleeper occurrence), the shape the LEAD endorsed
as a starting point. Verified round 3 by direct read (§1.5) + L6 (correct decision, owns both
horns). Regressions: (1) the report never states WHO classifies a row as infrastructure — the
only specified classifier is keyword clustering over friction TEXT, so loop-authored/injected
friction can self-route into the no-corroboration lane (L6 -> R3-5); (2) "every infrastructure
class ALSO surfaces independently on the doctor line" overclaims — the line prints only the
last skip reason, so transient events older than the most recent are on no channel but the stub
flag (L1+L5 -> R3-6); (3) the whole gate is void in rung-0 default mode (L5 -> R3-2). Closure
class: **closed_with_regression** -> successors **R3-2, R3-5, R3-6**.

### R2-7 — wrapper start/end snapshot had no reader or comparison step — CLOSED WITH REGRESSION
Found round 2 (L5+L6, superseding R1-26). Round-3 repair: the wrapper now COMPARES end-vs-start
and start-vs-prior-run's-end guardrail hashes at step 7; on mismatch it writes a dated TAMPER
flag beside the ledger and the next preflight fails closed (report line 479-483). Verified round
3 by direct read + L5/L6 (the watchman exists and is sound — not reopened). Regressions:
(1) the compare runs ONLY at step 7, so a run dying before it takes no END snapshot and
"start-vs-prior-END" has no referent — the tamper chain has an unspecified link at exactly the
abort path, which is model-influenceable (L5 -> R3-7); (2) the sensing surface (`git status
--porcelain` + guardrail-file hashes) is blind to out-of-repo subprocess writes, which is where
the layer-4 debate.js residual (leg a concedes it writes files) can land (L6 -> R3-8). Closure
class: **closed_with_regression** -> successors **R3-7, R3-8**.

### R2-8 — operator-owned plugin copy had no refresh owner or version check — CLOSED
Found round 2 (L5+L6, superseding R1-15). Round-3 repair: the wrapper preflight recomputes the
copy's content hash and FAILS CLOSED on mismatch with the operator-approved value (report line
611); the refresh step is added to scheduling.md's guardrail-PR merge checklist; a doctor line
reports copy staleness. Verified round 3 by direct read (line 611) + L6 (sound). The remaining
question — whether the wrapper VERIFYING the copy is itself executing from that copy
(self-verification) — is a NEW reflexivity finding about the wrapper's own root of trust
(R3-9), not an R2-8 residual. Closure class: **closed**.

### R2-9 — dead-man surface was pull-only — CLOSED
Found round 2 (L6, superseding R1-29). Round-3 repair: the staleness check now runs as a
SessionStart-hook warning line in every interactive Claude Code session (report line 745) — a
push surface the operator receives passively "whenever they use Claude Code at all" (invariant
7's human surface), directed by the LEAD. Verified round 3 by direct read. The warning's
enablement condition (an operator who schedules then stops leaves it firing forever) is folded
into the enumeration/hygiene successor R3-10, not carried here. Closure class: **closed**.

### R2-10 — resume cap bounds per-DIR not per-CAUSE — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-29). Round-3 repair: the wrapper normalizes each death's
abort reason into a dead SIGNATURE (exit class + first abort-record line); after M=3
consecutive fresh-dir deaths with the same signature it writes a HALT marker (preflight refuses
to launch until the human clears it); the "cannot ... forever" claim softened to per-dir scope
(report §3.4, line 409). Verified round 3 by direct read. Regressions: (1) abort records are
dated + run-dir-addressed by design, so unless the (asserted, unspecified) normalization strips
that variable content, no two signatures match and the HALT never fires — identity-keyed-
detector-lineage-blind applied to the design's own repair (L5 -> R3-11); (2) the per-night burn
arithmetic "k×$5/night" does not follow from one-resume-per-fire semantics (L2 -> R3-12). Closure
class: **closed_with_regression** -> successors **R3-11, R3-12**.

### R2-11 — 30-day auto-stale composes badly with graduation latency — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-22). Round-3 repair: a human-set `graduation-queued` status
exempts a stub from the 30-day auto-stale while still deduping its class (report line 267).
Verified round 3 by direct read. Regression: the exemption has NO upper bound — an abandoned
queued stub (queue-and-forget, operator turnover) is exempt forever while still deduping, so it
permanently subtracts its signal class with no backstop, re-opening for the queued subset the
exact monotonic-blinding failure R1-22 was minted to prevent (L6 -> R3-13). Closure class:
**closed_with_regression** -> successor **R3-13**.

### R2-12 — idempotency/resume gate still read loop-writable state — CLOSED
Found round 2 (L5, superseding R1-19). Round-3 repair: the wrapper derives completeness from
its own step-7 record in the operator-owned dir (not the loop-writable run dir), with the DEAD
marker location stated. Verified round 3 by L6 (sound). Closure class: **closed**.

### R2-13 — §6 row 5 likelihood cell carried the stale "(no API)" — CLOSED
Found round 2 (L3, superseding R1-9). Round-3 repair: cell requalified "(no spend-limit API;
rate-limit API unreachable at this auth tier — §5.1/R1-9, cell requalified round 2, R2-13)"
(report line 1122). Verified round 3 by direct read. Closure class: **closed**.

### R2-14 — §7 self-audit still characterized pricing as MEDIUM — CLOSED
Found round 2 (L4, superseding R1-11). Round-3 repair: the §7 Pattern B/E bullet now carries
"(upgraded to leaf-verified HIGH round 1, R1-11 — this bullet's lag fixed round 2, R2-14)" and
R1-11 is added to §7's banked-upgrade list. Verified round 3 by L4 (present, no residual).
Closure class: **closed**.

### R2-15 — §3.4 rung-1 label still "RECOMMENDED default" unqualified — CLOSED
Found round 2 (L5, superseding R1-14). Round-3 repair: label qualified "RECOMMENDED default
AMONG SCHEDULED RUNGS, once the human opts in." Verified round 3 by L5 spot-check. Closure
class: **closed**.

### R2-16 — gate-survival table rung-0 cells overstated + manual spend out-of-ledger unstated — CLOSED WITH REGRESSION
Found round 2 (L5, superseding R1-27). Round-3 repair: the R0 L2 cell split ("fence YES (cache
copy); canary n/a — no wrapper at rung 0", report line 774) and manual-run spend declared
out-of-ledger by design with the cap-arithmetic composition stated. Verified round 3 by direct
read. Regression: the rung-0 EXECUTION SHAPE for steps 0/2/4/7 (all wrapper-hosted) is still
undefined while §1.4 calls rung 0 "same code path" — the unowned sibling of the ledger horn
(L5 -> R3-2, jointly with R2-5/R2-6). Closure class: **closed_with_regression** -> successor
**R3-2**.

### R2-17 — §3.3 granted and revoked pdf/arxiv ToolSearch reach in one sentence — CLOSED
Found round 2 (L5). Round-3 repair: blue picked the trade — the loop's MCP profile is
`--strict-mcp-config --mcp-config <sleeper-mcp.json>` naming qmd only, and the round-1
parenthetical "research subagents reach pdf/arxiv tools via ToolSearch" is corrected as false
under that flag (ToolSearch discovers only from DECLARED servers; strict-mcp-config ignores the
project `.mcp.json`) — report lines 676-680. Verified round 3 by direct read. Closure class:
**closed**.

### R2-18 — $2–5/night × 30 vs ~$50 cap unreconciled — CLOSED
Found round 2 (L5). Round-3 repair: expected per-run spend owned from the smoke figure, and the
ceiling-vs-cap composition stated as the intended month-end throttle / anomaly signal (not
death). Verified round 3: L1 recomputed (30×$0.10–0.50 = $3–15/mo, ≥3× cap headroom at the
band; $2–5×30 = $60–150 ceiling correctly owned as throttle) and L5 concurs. Closure class:
**closed**.

### R2-19 — §0 artifact enumeration omitted the skill file + plugin manifest — CLOSED WITH REGRESSION
Found round 2 (L1). Round-3 repair: enumeration made total over the round-1 tree (skill file +
manifest added; L1 recount = 8 entries, total, HIGH). Regression: the round-2 minted artifacts
(SessionStart staleness hook + hooks.json, the doctor-line delta EDITING prosthetic-conscience,
the two operator-owned JSON configs) fall outside the now-"total" enumeration — exhaustive-
sweep-omits-own-specimen recurring against the round-2 additions (L5 -> R3-10). Closure class:
**closed_with_regression** -> successor **R3-10**.

### R2-20 — [^Backlog] pin range "15–17" under-covered by one line — CLOSED
Found round 2 (L1). Round-3 repair: range -> "15–18". Verified round 3 by L5 spot-check. Closure
class: **closed**.

### R2-21 — "The DGM analogy is exact" overstated — CLOSED
Found round 2 (L2). Round-3 repair: "exact"->"direct" plus the honesty clause (DGM evaluates
every change before archiving but admits even low scorers for exploration; our promotion gate
is stricter, pass-required). Verified round 3 (L2) by live re-fetch of arxiv.org/html/2505.22954:
"Only agents that compile successfully and retain the ability to edit ... are added to the
archive"; low scorers deliberately retained; parent selection roughly proportional to score —
every leg corroborates, HIGH. Closure class: **closed**.

### R2-22 — one [^CostRecord] marker on two figures from different artifacts — CLOSED
Found round 2 (L3). Round-3 repair: [^EfficiencyPlan] added beside the run-3 $149.95 figure
(cost.md carries $414.97). Verified round 3 by L5 spot-check. Closure class: **closed**.


## Round 4 closures (red merge, 2026-07-17)

### R3-1 — canary --json-schema leg undocumented + unprobed behavioral legs — CLOSED WITH REGRESSION
Found round 3 (L2+L5, superseding R2-1). Round-3 revision: blue demoted the mid-drive
`--json-schema` leg to verify-at-build with a fenced-block text-parse fallback named as the design
of record, and added OQ22 (mid-drive structured output) + OQ23 (two-phase-drive behavioral
acceptance legs). Verified round 4 (L2): step-3 restatement faithful, no regression on the
structured-output leg — HIGH. Regression: the R3-1 degrade-note's "named readers" are unspecified
at both reader sites — §2.3's `confidence` field names only the R2-17 PDF caveat (no qmd-degrade
labeling obligation), and §3.4's doctor line prints skip/abort reasons (a qmd degrade is neither),
so the reader declaration exists only at the writer's site (L2 → R4-7). Closure class:
**closed_with_regression** -> successor **R4-7**.

### R3-2 — rung-0 execution shape undefined ("same code path" vs wrapper-less) — CLOSED WITH REGRESSION
Found round 3 (L5, superseding R2-5/R2-6/R2-16). Round-3 revision: blue adopted the lead-endorsed
R2-16b resolution — rung 0 invokes the SAME wrapper via `node sleeper-wrapper.mjs --manual`, so
"same code path" becomes true by construction, the out-of-ledger accept DISSOLVES, and the rung-0
marker/corroboration voids close in the same move (debate.md ### BLUE round 3). Verified round 4
(L1 cross-read §3.4 ladder row 0 + §2.2 step 0 — CONSISTENT; manual dirs marker-stamped).
Regression (a DISTINCT surface, not the manual-execution path): `/self-improve` ships as an
invocable command under `commands/` carrying the full loop payload, guarded only by a prose header,
with no `disable-model-invocation: true` (its sibling graduate.md carries it), so the unwrapped path
stays mechanically open via command invocation (human or model/scheduled); and §3.4's containment
sentence has INVERTED polarity — markerless out-of-contract dirs PROVIDE non-sleeper corroboration
rather than "requiring" it (L5+L6 → R4-1). Origin-tag escape's 4th surface. Closure class:
**closed_with_regression** -> successor **R4-1**.

### R3-3 — origin cap does not reach the red-memory surface — CLOSED WITH REGRESSION
Found round 3 (L6, superseding R2-5). Round-3 revision: blue extended provenance tagging to the
red-memory mirror — patterns appearing/changing in a wrapper-logged sleeper window are appended to
a `sleeper-authored-patterns` list, capped at 1 per class, excluded from the corroboration pool;
snapshots extended to the red-memory dir; mirror pre-run frozen. Verified round 4 (L1 cross-read
§2.2 step 0 + §6 row 10 — CONSISTENT with harvest timing). Regression, two legs: (a) FORECLOSED
PREMISE — the nightly red-merge seat CANNOT write `.claude/agent-memory/**` under the sleeper
profile (fence outside research/+ideas/, `Edit(<REPO>/.claude/**)` deny, phase-0 canary abort on
marker loss), so §1.5's premise (the seat writes the memory dir in-window) and §4.2's profile
cannot both be true; the real consequence (the seat's mandatory record-new-patterns write
fails-denied every night — a silent degradation of the nightly adversary's learning + a routine
fence-denial stream in the fired-record) is stated nowhere; R3-3's live remit shrinks to
out-of-window authorship and the snapshot-diff machinery is dead where aimed. (b) CHANGE-KEYED
GRANULARITY — the trigger is file/header APPEARS-OR-CHANGES applied to the whole pattern, but red's
memory discipline is to EXTEND existing files, so one nightly append converts a pattern carrying
months of human-present evidence to sleeper-origin (R1-22 monotonic-blinding through the guard).
(L5+L6 → R4-4.) Closure class: **closed_with_regression** -> successor **R4-4**.

### R3-4 — run-window fallback void for dead runs — CLOSED WITH REGRESSION
Found round 3 (L1, superseding R2-5). Round-3 revision: blue logs the run-window START in the
operator-owned dir at step 0 and the END at whatever exit it observes (step-7 record, abort, or
DEAD-mark time), and runs the origin-marker porcelain sweep at DEAD-marking time — so even a
resumeless death leaves its window on record. Verified round 4 (L1 both required-fix legs present).
Regression the repair itself mints: a wrapper hard-kill (SIGKILL/power loss/OS kill — §6 row 6's
own High-likelihood environment) is an UNOBSERVED exit, leaving a window with START and no END
until DEAD-mark time ~3–4 days later; the multi-day window then sweeps daytime HUMAN-present run
dirs into `origin: sleeper` (capped at 1, excluded from the corroboration pool — suppressing
exactly the human-present corroboration ordinary classes wait on), and harvest's reading of an
UNCLOSED window is unspecified (L1+L6 → R4-6). Closure class: **closed_with_regression** ->
successor **R4-6**.

### R3-5 — infrastructure-class classifier source unstated — CLOSED
Found round 3 (L6, superseding R2-6). Round-3 revision: blue stated that the infrastructure-class
tag is assigned SOLELY from the wrapper's own event log (mechanical, session-unforgeable); friction
TEXT never self-classifies into the no-corroboration lane. Verified round 4 (L1): verbatim sentence
present, CLEAN. Closure class: **closed**.

### R3-6 — doctor line "also surfaces independently" overclaimed (last-slot only) — CLOSED
Found round 3 (L1+L5, superseding R2-6). Round-3 revision: the doctor/dead-man line now prints
PER-SIGNATURE COUNTS since last human clear, keyed by the R3-11 normalization (one mechanism fixes
R3-6 + R3-11 as the lead directed), so a transient infra event followed by another is no longer on
the stub flag alone. Verified round 4 (L1 cross-read §3.4 dead-man paragraph — CONSISTENT). Closure
class: **closed**.

### R3-7 — abort-path watchman gap (no END snapshot after a dead run) — CLOSED
Found round 3 (L5, superseding R2-7). Round-3 revision: the wrapper snapshots at EVERY exit path it
observes (abort included — it outlives the session by construction), and the cross-run compare is
defined against the LAST RECORDED snapshot of any kind, failing closed on a missing chain link per
invariant 8. Verified round 4 (L5 spot-check): a crash-without-end-snapshot is covered by
compare-vs-last-recorded = the dead run's START snapshot, spanning the window — sound. (The
unobserved-hard-kill window-log residual rides on R3-4's successor R4-6, not here — the snapshot
chain leg holds.) Closure class: **closed**.

### R3-8 — leg (b) detection scoped to in-repo porcelain + guardrail hashes — CLOSED
Found round 3 (L6, superseding R2-7). Round-3 revision: blue scoped leg (b) honestly — the snapshot
detects in-repo working-tree deltas and guardrail-file tampering, NOT arbitrary out-of-repo
subprocess writes; the out-of-repo write is stated as a residual bounded by pinned code +
no-remaining-write-gadget (citable only post-R3-15), sequenced as the lead directed. Verified round
4: the body carries the edits at layer 4 (iv) ("bounded HONESTLY round 3 (R3-8)"), layer 5
("Sensing scope stated honestly"), and §6 row 4 leg (b). The CHANGELOG Round-3 enumeration omits
the R3-8 bullet (16 of 17 under an "all 17 addressed" header) — a navigation-hint defect, not a
report-body defect; raised separately as R4-16 (change-summary desync, 2nd consecutive round).
Closure class: **closed**.

### R3-9 — wrapper absent from protection enumerations; root of trust unverified — CLOSED WITH REGRESSION
Found round 3 (L5+L6, recommend-not-block). Round-3 revision: blue absorbed the cheap parts —
watchmen named in layer 3 + the snapshot hash set, absolute-path invocation from the operator-owned
copy, a SessionStart last-check timestamp, wrapper-hash dates on the doctor line — and OWNED the
self-verification limit in text (the wrapper cannot attest itself; the residual is a box-compromise
scenario, not a loop-autonomy one). Verified round 4 by direct read of the round-3 debate/report
text. Regression: the dead-man push channel's disarm switch (scheduling-enabled flag) has unstated
custody, and the same-round R3-9 (always-print liveness) and R3-10 (condition-on-flag) repairs to
the SAME SessionStart line have unstated composition — if "conditioned" means fully silent when the
flag is off/absent, flag loss silently disarms the dead-man channel and a disarmed hook looks
identical to a healthy-but-disabled one, re-opening what R3-9's always-print clause closed (L6 →
R4-14). Closure class: **closed_with_regression** -> successor **R4-14**.

### R3-10 — round-2 minted artifacts outside §0's "total" enumeration — CLOSED WITH REGRESSION
Found round 3 (L5, superseding R2-19). Round-3 revision: blue extended §0's enumeration (SessionStart
staleness hook + hooks.json, the CROSS-PLUGIN prosthetic-conscience doctor-line delta, the two
operator-owned config files) and conditioned the SessionStart warning on scheduling-enabled.
Verified round 4 (L1 items present). Regression: the R3-10 repair extended the enumeration but never
reconciled the count HEADLINE — the same §0 paragraph simultaneously asserts "exactly THREE new code
artifacts" and enumerates a fourth executable (R3-10's own text: "a new executable + hooks.json
registration"); the skill file/manifest got an explicit "new PROSE artifacts, not code"
classification, the SessionStart hook got none, and its host plugin is unstated (L1 → R4-8).
Closure class: **closed_with_regression** -> successor **R4-8**.

### R3-11 — dead SIGNATURE normalization unspecified (HALT never fires) — CLOSED
Found round 3 (L5, superseding R2-10). Round-3 revision: blue specified the normalization —
signature = exit class + templated first abort line with `<date>`/`<path>`/`<id>`/`<n>` placeholders
(dates/paths/ids/nonces/digits stripped, the corpus's identity-keyed-detector lesson applied to the
design's own repair) — added zero-HALT-firings telemetry to the doctor line so a never-firing
detector is visible, and owned the alternating-cause (A,B,A,B) residual. Verified round 4 (L1/L2/L5
concur: concrete spec present, telemetry present, residual owned). Closure class: **closed**.

### R3-12 — per-night burn arithmetic wrong (k×$5 vs one-resume-per-fire) — CLOSED WITH REGRESSION
Found round 3 (L2, superseding R2-10). Round-3 revision: blue stated the in-night retry semantics
and recomputed (≤$5/night; cap trips ~10 nights; HALT at night 12 or the cap, whichever first).
Verified round 4 (L2 recomputed: one resume per fire → DEAD night 4/dir → HALT night 12; cap ~night
10 — the printed per-night figure now correct). Regression (merge recompute, resolving an L2/L6-held
vs L5-disputed conflict): the recomputed bound treats cap-trip as terminal, but the monthly cap
RESETS — at ceiling pricing the cap trips FIRST (~night 10) and PAUSES death accrual, so the third
same-signature death cannot occur in month 1 and the HALT lands early the NEXT month; worst-case ≈
$55–60 across two months and there is no "whichever comes first" race (the cap always pre-empts,
then un-pre-empts at rollover). The bounded conclusion survives (≈ one cap + ε); the printed race
and single-month bound are wrong — the same repair-introduced-arithmetic class R3-12 itself fixed
(L5 → R4-10). Closure class: **closed_with_regression** -> successor **R4-10**.

### R3-13 — graduation-queued exemption has no upper bound — CLOSED WITH REGRESSION
Found round 3 (L6, superseding R2-11). Round-3 revision: blue gave `graduation-queued` its own
M=90-day `queued-stale` re-surface for human re-confirmation (§1.4/§2.3), and labeled the field
"no status is timer-free." Verified round 4 (L1: both sites present, CLEAN for the queued state).
Regression: "no status is timer-free" is false as written — the terminal states `rejected` and
`graduated` have neither timer nor stated dedupe semantics, so a rejection either permanently
subtracts its class (the monotonic-blinding failure R1-22 forbids) or re-mints a stub every run
(the Dependabot-fatigue arm); the backlog regression rule covers backlog items, not stub statuses.
Third consecutive per-status patch (R1-22 → R2-11 → R3-13) signals the missing root invariant
(every status's dedupe effect has a stated re-surface path) (L5 → R4-9). Closure class:
**closed_with_regression** -> successor **R4-9**.

### R3-14 — dontAsk closed-world premise refuted; Bash read carve-out — CLOSED WITH REGRESSION
Found round 3 (L2, first-raise). Round-3 revision: blue corrected §3.2's dontAsk sentence to carry
the carve-out, re-derived §4.2's profile with enumerated `deny` rules over the 14-command read-only
set (all 14 deny-covered, diffed line-by-line by L4/L6), re-argued §6 rows 4/13 + layer 2 on the
true surface, and added [^PermissionsDoc] round-3 quotes (verbatim vs live doc, L2/L3/L4/L6 HIGH,
zero drift). Verified round 4. Regressions, two: (a) the enumeration rests on a NON-EXHAUSTIVE doc
list — the doc says the set "**include**[s]" 14 commands and never claims completeness, and the same
page names `sort`/`sed` as classifier-reasoned commands not in the 14; an unlisted member
(`sort`/`file`/`readlink`/`strings`/`less`) auto-runs un-denied, re-opening the Bash read channel
(invariant-soundness-by-enumeration applied to the design's own repair) (L3 → R4-3); (b) the
prior-exposure example is refuted at the leaf — `Bash(cat //…/.claude/projects/…)` on row 13's named
transcript target was NOT auto-approved under the round-2 profile, because that profile's
`Read(//…/.claude/projects/**)` deny extends to recognized Bash file commands per the doc's own
deny-reach clause; the real round-2 exposure was allow-scoped-but-not-Read-denied paths
(credentials-class), so both sites overstate what R3-14 closed (postmortem-misdiagnosis) (L2+L4 →
R4-5). Closure class: **closed_with_regression** -> successors **R4-3, R4-5**.

### R3-15 — git log --output write gadget refutes "pinned read-only git argv" — CLOSED WITH REGRESSION
Found round 3 (L4, first-raise; tool-run confirmed). Round-3 revision: blue added
`--output`/`--output-directory`/`-O`-class flags to the sleeper-guard hook's Bash-write matcher and
belt denies, added the channel to OQ18's named scope, and downgraded row 4 leg (a) + layer 4 (i) to
"read-only EXCEPT the write channel." Verified round 4. Regressions, two: (a) read-only git is
retained UN-enumerated in the carve-out (the design needs git reads), so "deny-enumerated per
command" overclaims for its git member — and sibling git-native output flags escape both belt and
hook: leaf-verified this box `git format-patch -1 -o /tmp/l5probe` → exit 0, arbitrary out-of-repo
patch (the `-o` short form matches none of the three long-form belt denies and is not in the hook's
named list); `git archive -o`/`git bundle create` are further writers — sibling-halo on the R3-15
closure, out-of-repo targets in the R3-8 blind spot (L5 → R4-2); (b) the "no prompt" reproduction is
attributed to the carve-out classifier without isolating the layer — both round-3 reproductions ran
under `defaultMode: "auto"`, where the AUTO classifier is the approving layer, so "showing the
carve-out classifier itself passes `--output`" and "rule-pinning alone cannot close it" are not
established by the probe (the isolating dontAsk-zero-allow probe was attempted twice from lens seats
and DENIED by the seat's own auto-mode classifier; not routed around) (L3+L4 → R4-11). Closure class:
**closed_with_regression** -> successors **R4-2, R4-11**.

### R3-16 — §1.3 telemetry-input row stale (0.6.0 future tense) — CLOSED
Found round 3 (L1, first-raise). Round-3 revision: blue updated the row to "SHIPPED as of FEOV 0.7.0
— present in this run's own trajectories/." Verified round 4 (L1): Glob confirms
`trajectories/board-telemetry.jsonl` EXISTS in this run dir — filesystem-verified, HIGH. Closure
class: **closed**.

### R3-17 — [^Pricing] tokenizer scope under-inclusive (omits Opus 4.7+) — CLOSED
Found round 3 (L3, first-raise; recommend risk-accept). Round-3 revision: rather than risk-accept,
blue completed the set — "tokenizer set completed with Opus 4.7+" (debate.md ### BLUE round 3).
Verified round 4 (L3 live re-fetch): the pricing page names "Opus 4.7 and later Opus models, Fable
5, Mythos 5, Mythos Preview, Sonnet 5" — verbatim, HIGH, zero drift. Closure class: **closed**.

## Round 5 closures (red merge, 2026-07-17)

### R4-1 — /self-improve invocable-command surface + inverted containment polarity — CLOSED WITH REGRESSION
Found round 4 (L5+L6, superseding R3-2). Round-4 revision: `self-improve.md` became a thin
trampoline carrying `disable-model-invocation: true` (matching graduate.md), the loop payload
moved to the wrapper's phase-1 prompt sourced from skills/continuous-learning, and §3.4's
containment polarity was corrected (markerless out-of-contract dirs are NON-sleeper and CAN
corroborate; residual = a human's deliberate paste-run, named, quarantine deferred as OQ24).
Verified round 5: L6 audited the trampoline mechanism directly — sound, symmetric with
graduate.md; L5/L2 verified the §0 tree + §3.4 paragraph + §6 row 10 + OQ24. Regressions:
(a) two body sites still specify the OLD payload shape — §3.4 ladder row 0 ("the /self-improve
command markdown is the wrapper's phase-1 prompt payload in EVERY mode") and §3.3's adopted
Phase-4 acceptance test (`claude -p "/self-improve"` produces a run dir — now fails by
construction; cheapest pass re-inlines the payload) (L2+L5 → R5-2); (b) the paste-run pool
residual was deferred via OQ24 on a rationale that prices only paste-run likelihood — the
rung-2 automatic-poisoning case is unargued, folded into the pool root-invariant finding
(L6 → R5-3). Closure class: **closed_with_regression** -> successors **R5-2, R5-3**.

### R4-2 — git carve-out member un-enumerated; sibling git-native writers escape belt+hook — CLOSED WITH REGRESSION
Found round 4 (L5, superseding R3-15). Round-4 revision: the git channel INVERTED to a
read-ALLOWLIST at the sleeper-guard hook (deny any git argv not an exact allowed read form —
the lead-endorsed invariant-6 shape); belt denies extended to `-o`/`-O` short forms and the
format-patch/archive/bundle/config/gc/repack/maintenance writer family; §6 row 4 corrected to
name the git exception with OQ18(c) as standing test. Verified round 5: L3 reproduced the
spaced-form gadget live (`git format-patch -1 -o /tmp/… → exit 0` out-of-repo) confirming the
absorbed leaf claim; L5/L6 verified the allowlist inversion text. Regressions: (a) R4-2's own
reachability premise — "Where Bash IS reachable (a rebuilt rung, the Workflow seat agents,
profile drift)" — contradicts §4.3 layer 4 (iii)'s "seat agents are full permission-engine +
hook subjects"; the seat-surface composition with R4-3's bare deny is undecided and
settings-inheritance unprobed (L3+L5+L6 → R5-1); (b) the ATTACHED `-o<value>` form (no space)
escapes the new `Bash(* -o *)` belt pattern — leaf-verified by L6 and re-reproduced at the
merge seat (`git format-patch -1 -o/tmp/r5mergeA HEAD` → exit 0, out-of-repo patch) — the
enumerate-and-extend regress recurring inside the repair; belt-only (hook allowlist is the
load of record) but the belt binds rebuilt rungs 3–4 (L6 → R5-5). Closure class:
**closed_with_regression** -> successors **R5-1, R5-5**.

### R4-3 — carve-out deny enumeration rests on a non-exhaustive doc list — CLOSED WITH REGRESSION
Found round 4 (L3, superseding R3-14). Round-4 revision: a bare `Bash` deny added to the
sleeper profile — doc-verified ("A bare tool name like `Bash` removes the tool from Claude's
context entirely"; `Bash(*)`≡`Bash` as deny) — closing the whole dontAsk read-only carve-out
class (enumerated + unlisted members + read-only git + every git write gadget) at the tool
boundary for the top-level session; enumerated denies retained as belt; `sort`/`sed`/`file`/
`readlink`/`strings`/`less` + `~/.ssh`/`.env` belt denies added; OQ18(c) gained the
member-enumeration probe. Verified round 5: L3 re-fetched the permissions doc live — every
load-bearing quote verbatim, zero drift; the "include[s]"/non-exhaustive premise and
deny-supremacy confirmed. Regression: the structural close is scoped to the TOP-LEVEL session
("§2.2's session steps never invoke Bash") while §6 row 13 and §4.3 layer 2 state TOTAL
claims ("holds on the Bash channel too") and §4.2's R4-2 bullet simultaneously concedes Bash
IS reachable in the Workflow seat agents — the two horns are undecided, and if seats do not
inherit the `--settings` profile, closure for the nightly worker population reverts to the
enumeration R4-3 itself declared non-load-bearing (L3+L5+L6 → R5-1; the stale "carve-out
auto-approves read-only git regardless" comment is the same finding's trivial sub-leg).
Closure class: **closed_with_regression** -> successor **R5-1**.

### R4-4 — provenance mechanism aimed at a profile-foreclosed write — CLOSED WITH REGRESSION
Found round 4 (L5+L6, superseding R3-3). Round-4 revision: blue picked the cheapest coherent
horn and stated it plainly — agent-memory writes are DENIED by design under the sleeper
profile; nightly seats do not learn; nightly patterns reach the corpus only via
ordinary-classed friction text; the R3-3 machinery re-scoped to belt-for-drift; tagging at
window-ADDED granularity (pre-existing content keeps its origin), closing the R1-22
monotonic-blinding leg. Verified round 5 (L5/L6 direct read of the rewritten §1.5 paragraph +
§6 row 10). Regression the repair mints: the mechanism making the denial stream un-alarming —
bucket every `.claude/agent-memory/**` denial as `origin=nightly-seat,
class=expected-memory-deny`, unflagged by the doctor line — is keyed on the TARGET PATH, not
the actor/cause, so an injection-driven memory-poisoning ATTEMPT (§6 row 14's
Medium-likelihood channel) at any volume is indistinguishable from the seat's one expected
protocol write; enforcement holds, attack-attempt telemetry on the adversary-steering surface
is erased (L5+L6 → R5-6). Closure class: **closed_with_regression** -> successor **R5-6**.

### R4-5 — prior-exposure example refuted by the doc's deny-reach clause — CLOSED
Found round 4 (L2+L4, superseding R3-14). Round-4 revision: postmortem corrected at BOTH
sites (§4.2 prose bullet AND §6 row 13) — the round-3 `Bash(cat …/.claude/projects/…)`
example acknowledged as mis-diagnosed (that path was Read-denied, and Read/Edit denies extend
to recognized Bash file commands per the doc); example re-pointed at un-named
credentials-class paths (`~/.aws/credentials`, stray `.env`); deny-reach clause noted in the
profile comments. Verified round 5 (L3): the deny-reach clause re-fetched verbatim live
("Read and Edit deny rules apply … and to file commands Claude Code recognizes in Bash, such
as `cat`, `head`, `tail`, and `sed`"); both corrected sites faithful to the doc. Closure
class: **closed**.

### R4-6 — unobserved-exit window sweeps human-present dirs — CLOSED WITH REGRESSION
Found round 4 (L1+L6, superseding R3-4). Round-4 revision: a window's END is additionally
bounded by the NEXT wrapper START (no window spans invocations); a DEAD-mark-closed window is
flagged `retroactive-uncertain` with its markerless sweep CONFINED to sleeper date-key naming
(other dirs surfaced for human confirmation); unclosed-window read specified
(extends-to-present for date-key-named dirs only); snapshot-chain + resume backstop owned in
step-7/layer-5 text. Verified round 5 (L5 direct read; all three owed legs present).
Regressions: (a) the confinement decides sweep membership BY dir name while §1.5's unqualified
doctrine sentence ("reads ONLY the marker, never the dir name") still stands, and the named
convention "the wrapper's own sub-run slug" is unknowable after a hard-kill (model-chosen
nightly, recorded nowhere durable; mkdir-to-stamp bound unstated) — row 10's "cannot sweep
human-present dirs" overclaims or the auto-tag under-delivers (L5 → R5-4); (b) the
window-close is the second of three per-surface pool patches folded into the pool
root-invariant finding (L6 → R5-3). Closure class: **closed_with_regression** -> successors
**R5-4, R5-3**.

### R4-7 — degrade-note reader sites carried no surfacing obligation — CLOSED
Found round 4 (L2, superseding R3-1). Round-4 revision: §2.3's confidence field gained the
"recall-degraded (qmd down; lexical Grep/Read only this run)" labeling clause; §3.4's doctor
line gained the qmd-degrade streak term ("qmd degraded on M of the last N runs"); §2.2 step
0's "named readers" sentence now has both readers specified. Verified round 5 (L2 + L5 direct
reads at both reader sites — the obligation exists where readers read). Closure class:
**closed**.

### R4-8 — "exactly THREE code artifacts" unreconciled with the fourth executable — CLOSED WITH REGRESSION
Found round 4 (L1, superseding R3-10). Round-4 revision: count corrected to FOUR code
artifacts (SessionStart staleness-warning hook counted); host plugin named — sleeper-service,
with the rationale (fires in interactive sessions independent of the scheduler, so it ships
with the sleeper plugin's own hooks.json). Verified round 5 (L5/L6 direct read of §0).
Regressions: (a) the §0 TREE (labeled "the implementable shape") was never reconciled — no
hooks/hooks.json entry, no SessionStart executable, and nothing registers the sleeper-guard
PreToolUse hook either; the enumeration has now been repaired three times (R2-19/R3-10/R4-8)
while the tree never was (L5 → R5-8); (b) homing the hook in sleeper-service's OWN hooks.json
extends the empty-bin crash-storm surface to every interactive session while §6 row 9's bound
cites prosthetic-conscience's hooks.json — guard coverage for the new file unstated
(L6 → R5-7). Closure class: **closed_with_regression** -> successors **R5-8, R5-7**.

### R4-9 — rejected/graduated statuses timer-free; missing root invariant — CLOSED WITH REGRESSION
Found round 4 (L5, superseding R3-13; chain root R1-22). Round-4 revision: the root invariant
stated ONCE in §1.4 — every status's dedupe effect has a stated re-surface path; no status
subtracts its class permanently — with terminal-state semantics: `graduated` → class
recurrence re-enters flagged `regression`; `rejected` → dedupes for a 90-day/rate window then
re-surfaces `rejected-recurring` for one-keystroke re-confirm. §2.3 enum + §6 row 3 carry it.
Verified round 5 (L5: invariant present once, semantics stated; L2: enum cross-read).
Regressions: (a) both arms of the `rejected` clause key on a rejection DATE no artifact
records (stub filenames date the MINT; status edits undated) — "pre-rejection rate" and the
90-day window are uncomputable by the zero-token harvest as specified (L5 → R5-9); (b) the
`regression` token is not in the §2.3 enum and its domain (docket flag vs stub status; the
`rejected-recurring` setter) is unpinned (L2 → R5-9). Closure class:
**closed_with_regression** -> successor **R5-9**.

### R4-10 — cap-trip treated as terminal but the monthly cap RESETS — CLOSED
Found round 4 (L5, superseding R3-12; intra-round lens conflict resolved for L5 by merge
recompute). Round-4 revision: §3.4 recomputed — at ceiling pricing deaths land nights 4 and 8;
the ledger preflight cap-skips from ~night 11; no death accrues during cap-skips, so the M=3
HALT lands early the NEXT month (~$5–10 in); worst-case ≈ one monthly cap + ≤2 nights
(~$55–60 across two months); the "whichever comes first" race removed. Verified round 5 by
THREE independent recomputes (L2, L5, L6) — all figures check. Closure class: **closed**.

### R4-11 — "no prompt" attributed to the carve-out classifier without layer isolation — CLOSED
Found round 4 (L3+L4, superseding R3-15). Round-4 revision: attribution re-scoped at all
three sites (§4.2 bullet, §7 round-3 update, OQ23(d)) to "consistent with carve-out
classification but does NOT isolate it (both probes ran under defaultMode: auto)"; the
isolating dontAsk-zero-allow probe deferred to build (OQ18(c)/OQ23(d)); hook matcher stated
chosen-conservative. Verified round 5 (L5 direct read of all three sites — faithful). Closure
class: **closed**.

### R4-12 — est_complexity factor had no stated input source — CLOSED WITH REGRESSION
Found round 4 (L5, first-raise). Round-4 revision: §1.4 stage 2 now states the source —
default 1 (inert; the factor vanishes) unless the class's matching ideas/backlog.md entry
carries a human-recorded complexity note, which harvest parses; ranking stays zero-token
arithmetic. Verified round 5 (L5: clause present and safe-defaulted). Regression: "harvest
parses that note's value" states no FORMAT (backlog entries are free prose — unparseable as
arithmetic), and L1's leaf check at the pin shows NO backlog entry carries any parseable
complexity field — the named source is a convention that does not yet exist, so the divisor
is universally inert while the text implies it activates (L1+L5 → R5-10). Closure class:
**closed_with_regression** -> successor **R5-10**.

### R4-13 — provenance/corroboration gate absent from the gate-survival table — CLOSED WITH REGRESSION
Found round 4 (L5, first-raise). Round-4 revision: the §3.4 gate-survival table gained a
"Provenance/origin-tag + corroboration gate (§1.5)" row (YES R0/R1 via the wrapper; NO R2–R4
absent rebuild) and the graduation-grade adoption requirement names rebuilding this gate as a
precondition. Verified round 5 (L5: row present, requirement named). Regression: the row +
instructional adoption requirement is the third per-surface patch on the negatively-defined
corroboration pool (R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13) with the root invariant still
missing; at rung 2 the poisoning is AUTOMATIC (local runs against the same harvested corpus,
markerless → counted non-sleeper next morning) behind an instructional-only gate — the class
§4.1's own 0/175 evidence refutes; OQ24's deferral rationale never argues this case
(L6 → R5-3). Closure class: **closed_with_regression** -> successor **R5-3**.

### R4-14 — dead-man disarm flag custody + R3-9/R3-10 composition unstated — CLOSED WITH REGRESSION
Found round 4 (L6, superseding R3-9). Round-4 revision: the scheduling-enabled flag lives in
the OPERATOR-OWNED dir beside the ledger (outside every session write surface); the
SessionStart line is NEVER fully silent — flag-off prints the disabled token, flag-absent
prints "sleeper: flag missing" (anomaly), flag-on-stale prints the staleness warning —
silent-disarm distinguishable from disabled-by-choice. Verified round 5 (L5: four states all
print; custody stated; composition with R3-9 explicit). Regression: flag-ABSENT is the
DEFAULT state of a fresh install before schedule-setup ever runs, so the anomaly print nags
every never-configured operator per session — never-configured vs flag-lost undistinguished;
the alert-fatigue mode the report catalogs, aimed at its own dead-man channel (L6 → R5-7).
Closure class: **closed_with_regression** -> successor **R5-7**.

### R4-15 — OQ18(c) flag-scoped while the black box is the git SUBCOMMAND boundary — CLOSED
Found round 4 (L4, first-raise; recommend-not-block). Round-4 revision: OQ18(c) extended to
the subcommand boundary of "read-only forms of git" with `config`/`gc`/`repack`/`maintenance`
named as probe cases (the `git config` pager/alias arbitrary-command-exec consequence
stated); the hook git read-ALLOWLIST rejects them by default (inside deny-by-default), and
the writer family entered the belt denies. Verified round 5 (L5/L3: OQ18(c) text + §4.2
carry it). Closure class: **closed**.

### R4-16 — CHANGELOG round-3 enumeration missing the R3-8 bullet — CLOSED
Found round 4 (L5, first-raise; process note). Round-4 revision: the R3-8 bullet added to
CHANGELOG Round 3 (content copied from §7's round-3 list). Verified round 5 (L5 + merge
direct read of CHANGELOG — bullet present; the change-summary channel matches the report for
the first time in three rounds). Closure class: **closed**.
