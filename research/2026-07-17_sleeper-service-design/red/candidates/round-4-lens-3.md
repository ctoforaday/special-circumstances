# round 4 — lens 3 (leaf-node citation verification; slice: §4/§5 + referenced footnote definitions)

Full report re-read whole (4 consecutive Read windows, 1,893 lines). Round-3 revision touched
this slice via R3-14/R3-15 (§4.2, §4.3 layers 2/4, [^PermissionsDoc] additions), R3-9
(§4.3 layer 3), R3-17 (§5.2 + [^Pricing]). Re-fetch policy applied: changed-section claims
leaf-verified live; >2-rounds-old living/academic leaves re-fetched (permissions doc, pricing,
cli-reference, #6631, #25621, usage-cost API, rate-limits API, STOP ar5iv, sakana ai-scientist,
sakana dgm, OWASP LLM06); pin-immutable internal leaves ([^SemanticConsent], [^PushGuard],
[^CostRecord], [^EfficiencyPlan], [^HooksJson], [^IdeasCorpus]) carried on pin immutability.

## Verification results (all live fetches 2026-07-17)

| Claim (section) | Source | Result |
|---|---|---|
| R3-14 carve-out quote + 14-command set + "not configurable" + ask/deny remedy ([^PermissionsDoc], §3.2, §4.2) | permissions doc live | VERBATIM — high |
| "a deny rule can't carry allowlist exceptions" ([^PermissionsDoc]) | permissions doc | VERBATIM — high |
| Unquoted globs auto-run only for all-read-only-flag commands; find/git still prompt ([^PermissionsDoc]) | permissions doc ("Commands with write-capable or exec-capable flags, such as `find`, `sort`, `sed`, and `git`, still prompt when an unquoted glob is present") | high |
| Exec wrappers + `find -exec`/`-delete` always prompt ([^PermissionsDoc]) | permissions doc | VERBATIM — high |
| `permissions.disableBypassPermissionsMode: "disable"` placement + value (§4.2 JSON) | permissions doc ("set `permissions.disableBypassPermissionsMode` or `permissions.disableAutoMode` to `"disable"` in any settings file") | high — and see L3-F3 |
| "If a tool is denied at any level, no other level can allow it" (§4.2) | permissions doc | VERBATIM — high |
| "A blocking hook … stops the tool call before permission rules are evaluated" (§4.3 layers 2/4) | permissions doc ("A blocking hook also takes precedence over allow rules. A hook that exits with code 2 stops the tool call before permission rules are evaluated") | high |
| R3-17 tokenizer set (§5.2 + [^Pricing]) | pricing page ("Claude Opus 4.7 and later Opus models, Claude Fable 5, Claude Mythos 5, Claude Mythos Preview, and Claude Sonnet 5 use a newer tokenizer … approximately 30% more tokens"; "Claude Sonnet 4.6 and earlier models use the previous tokenizer") | VERBATIM — high |
| [^Pricing] full grid zero-drift re-fetch (VOLATILE self-flag): Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 $2/$10 intro through 2026-08-31 → $3/$15 from 2026-09-01; Opus 4.5–4.8 $5/$25; Fable/Mythos 5 $10/$50; Batch flat 50%; cache read 0.1× | pricing page | high — zero drift |
| [^DenyRWIssue] #6631: closed; fix claimed via #4467 ("fix merged, should go out in next release", ant-kurt); reporter re-confirmed bypass at v1.0.93 Aug 2025 (§4.1) | github live | high — zero drift, r4 re-check |
| [^DenyBashIssue] #25621: CLOSED as duplicate, canonical not named on page (§4.1) | github live | high — zero drift, r4 re-check |
| [^CliReference] §5.1 flags: `--max-budget-usd` verbatim; `--max-turns` "Exits with an error"; `--fallback-model` NO print-only marker + persistent `fallbackModel` setting documented (R1-10 stands); no exit-code table (R1-6 stands) | cli-reference live | high — zero drift, r4 re-check |
| [^UsageAPI] §5.1: both endpoints; Admin key required; "The Admin API is unavailable for individual accounts"; ~5-min freshness; 1/min sustained polling | usage-cost-api live | high — r4 re-check |
| [^RateLimitsAPI] §5.1: `/v1/organizations/rate_limits` + workspace variant; READ-only ("Can I update rate limits with this API? **No.** … use the **Limits** tab"); Admin key; no spend-limit API documented | rate-limits-api live | high — r4 re-check; §5.1's requalified shape fully corroborated |
| [^STOP] §4.1 figures: 0.42% (Wilson 95% CI 0.31–0.57), 0.46% (0.35–0.61) with warning, 10,000 improvements, not significant (two-proportion z-test, α=0.05), syntactic check `use_sandbox=False`/`exec(` | ar5iv live | high — r4 re-check, exact |
| [^DGMSakana] §4.1/§2.4: fake-log quote, markers-removal quote incl. parenthetical, lineage quote, sandboxed-under-human-supervision, "improve themselves the more compute they are provided" | sakana.ai/dgm live | high — all five verbatim |
| [^AIScientist] §4.1: sandboxing recommendation VERBATIM ("These issues can be mitigated by sandboxing the operating environment of The AI Scientist"); both self-modification incidents still described on the live page ("it simply tried to modify its own code to extend the timeout period" returned verbatim; the "system call to run itself" fragment was captured verbatim at r1 and the extractor confirmed the behavior is still described) | sakana.ai/ai-scientist live | high — carried r1 verbatim + live phenomenon re-confirm |
| [^OWASP] §4.1: three root causes (excessive functionality/permissions/autonomy); least-privilege, human-in-the-loop-for-high-impact, authorization-in-downstream-systems, logging, rate-limiting all quoted; "draft-vs-execute separation" supported by the entry's manual-review/"hit 'send'" mitigation example | genai.owasp.org live | high — the footnote's draft-vs-execute gloss, previously unpinned, is now leaf-confirmed |

Internal arithmetic re-checks (§5.2): 30×$0.10–0.50 = $3–15/mo ✓; $50/$15 ≈ 3.3× headroom ("≥3×") ✓; $2–5×30 = $60–150/mo ✓; $150–400×30 = $4.5k–12k ✓. Cache-read 0.1× = "cuts cached input ~90%" ✓.

## Findings

### L3-F1 — carve-out deny enumeration rests on a non-exhaustive doc list (MEDIUM: L Low-Medium × I Medium × Cx Low)

Location: §4.2, JSON comment — "One deny per carve-out command — the session's read surface
is the scoped native Read/Grep/Glob tools, never shell reads; full enumeration in the shipped
file"; and §4.3 layer 2 — "the carve-out's commands deny-enumerated in §4.2".

The doc sentence the enumeration is built from is non-exhaustive by its own phrasing: "These
**include** `ls`, `cat`, `echo`, `pwd`, `head`, `tail`, `grep`, `find`, `wc`, `which`,
`diff`, `stat`, `du`, `cd`, and read-only forms of `git`." The set is a platform classifier
("The set is not configurable") and the doc nowhere claims the list is complete — the same
page's unquoted-glob passage casually names `sort` and `sed` as commands the classifier
reasons about, neither of which is in the named carve-out list. An undocumented member (e.g.
`sort`, `file`, `readlink`, `strings`, `less`) would auto-run un-denied, re-opening the Bash
read channel that §6 row 13's repair ("closed by §4.2's enumerated carve-out denies") and
§4.3 layer 2's unforgeability leg both lean on. The belt Read-denies only bind the NAMED
`~/.claude` targets (and, per the same doc, recognized Bash file commands on those paths);
un-named box-local secrets (`~/.ssh`, stray `.env`) stay exposed to an unlisted member.
OQ18(c) probes "OTHER write-capable **flags** inside carve-out commands" — flags within known
members, not unlisted members. This is the invariant-soundness-by-enumeration pattern applied
to the design's own repair: R3-14 replaced a closed-world premise with an enumeration whose
source hedges.

Required fix (either leg is cheap; blue argues which):
(a) extend OQ18 with a member-enumeration probe — under a bare dontAsk profile with no allow
rules, attempt a candidate list of common read-only commands; any that auto-run get denies in
the shipped file; note the shipped file's bare-vs-`*` convention (the sample denies `Bash(ls)`
AND `Bash(ls *)` but only `Bash(cd *)`, only `Bash(pwd)`) so coverage is stated, not implied;
AND/OR
(b) the strictly stronger shape, consistent with invariant 6's allowlist doctrine: deny the
bare `Bash` tool in the sleeper profile — doc-verified this round: "A bare tool name like
`Bash` removes the tool from Claude's context entirely," which closes the ENTIRE carve-out
class structurally (unlisted members included). Cost audit: §2.2's session steps never invoke
Bash (steps 1–3/5 are read/Edit; step 4 is the Workflow tool, not Bash; the canary is an
Edit); the three git allow rules are already conceded "declared intent, not the enforcement
surface." The sleeper-guard's Bash-write matcher then becomes belt behind a removed tool
rather than the enforcement of record for an open classifier.

### L3-F2 — "no prompt" reproduction attributed to the carve-out classifier without isolating the layer (LOW: evidence-grading, not design)

Location: §4.2 bullet — "blue's reproduction ran without any prompt, which is itself evidence
the carve-out classifier treats `git log --output` as read-only, so rule-pinning ALONE cannot
close it"; repeated in §7 round-3 update and §4.3 layer 4 (i).

Blue's seat runs under this run's own permission configuration, which demonstrably allows
broad Bash (the seat ran `git log --output`, `git show`, `wc` freely all round). A no-prompt
outcome under that session is equally explained by the session's allow rules; the probe
cannot distinguish the carve-out classifier from a co-located allow layer (verification-probe
layer-masking — the same defect class R2-2 fixed for the canary). The three-part close is
unaffected (belt denies + hook matcher bind under either explanation; and IF the classifier
does NOT pass `--output`, exact-argv pinning alone would already deny it under dontAsk — so
the "rule-pinning alone cannot close it" conclusion is also not established by this probe,
merely made prudent). Required fix: reword the evidentiary sentence to "consistent with the
classifier treating it as read-only" and fold a classifier-isolation probe into the OQ18/OQ23
build tests (run `git log --output` under a dontAsk profile with NO git allow rules; observe
auto-run vs prompt). Attempt-or-impossibility: I cannot isolate it from this seat either — my
session carries its own allow surface; the isolation requires the clean profile the build
test owns.

### L3-F3 — OQ17 is answerable at leaf today (LOW — banked upgrade, not a defect)

Location: §8 OQ17 — "`disableAutoMode` … leaf-verify the key's existence, name, and scope in
the current permissions doc before adding it to the profile"; §4.2 parenthetical.

Leaf-verified this round: the permissions doc documents it in the same sentence as the
bypass lockout — "set `permissions.disableBypassPermissionsMode` or
`permissions.disableAutoMode` to `"disable"` in any settings file." Key name
`disableAutoMode`, scope `permissions.`-object, value `"disable"`, any settings scope. The
cli-reference confirms `auto` is a live `--permission-mode` value, so the escape hatch the
key closes is real. Blue can add the line to §4.2's JSON and close OQ17 this round.

### L3-F4 — doc fact strengthening row 13, not carried (LOW — optional one-line absorb)

Location: §6 row 13 / §4.2 parenthetical (the Bash-read-channel story). The permissions doc
states: "Read and Edit deny rules apply to Claude's built-in file tools **and to file
commands Claude Code recognizes in Bash**, such as `cat`, `head`, `tail`, and `sed`." The
§4.2 belt Read-denies on `~/.claude` settings/credentials/transcripts therefore also bind
recognized Bash file commands on those paths — a second, independent closure of the named
exfil targets that holds even where L3-F1's unlisted-member risk applies. Worth one line in
§4.2; it also honestly narrows L3-F1's impact to un-named targets.

## Attempt lines (MUST-TRY observable)

Every external citation in slice graded HIGH this round; no citation graded down. L3-F2's LOW
is an internal-evidence grade with the isolation-impossibility stated above. The
[^AIScientist] extractor returned one fragment as behavior-described rather than verbatim;
graded HIGH on the r1 verbatim capture + this round's live confirmation of the passage's
presence (WebFetch, sakana.ai/ai-scientist, 2026-07-17) — not a down-grade.

## Repairs in slice verified clean (no regression)

R3-14 (§3.2 sentence, §4.2 denies, §4.3 layer-2 re-derivation, [^PermissionsDoc] quote),
R3-15 (rule pinning + belt denies + hook matcher; sequencing clause vs R3-8 coherent),
R3-9 (watchmen named in layer 3; self-verification residual owned), R3-17 ([^Pricing]
set verbatim). §4.2 JSON key placement and value both match the doc.
