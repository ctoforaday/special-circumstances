# Round 5 — Red lens 4 (leaf-node citation verification), slice §6/§7/§8 + referenced footnotes

Verdict: **no new citation gaps.** Every load-bearing citation in the slice is corroborated
at HIGH; the two volatile living-source leaves were re-fetched this round zero-drift; the one
load-bearing local leaf (git output gadget) reproduces; round-4 edits to the slice introduced
no un-checked external citation. One volatility SIGNAL flagged for future rounds (not a gap).

## Method / what changed this round in-slice
Round-4 blue revision touched §6 (rows 3/4/10/13), §7 (attribution note), §8 (OQ18(c),
OQ23(d), OQ24 added). Those edits are internal-design or rest on citations already
leaf-verified round 4 (git gadget, permissions-doc deny-reach clause). No NEW external figure
or quote was minted in-slice at round 4 — so the drift/`>2-round`/section-changed triggers
resolve to: re-fetch the volatile GitHub trackers, re-run the local git gadget, carry the rest.

## Volatile leaves re-fetched this round (living issue trackers, 1 round since r4)
- **#76239 — OPEN, zero drift** (round-5 live WebFetch). Title verbatim: "SDK headless: MCP
  tools silently missing on first turn when stdio server startup is slower than the (new)
  non-blocking pre-wait — regression for single-turn sessions since CLI 2.1.144." Corroborates
  §6 row 8 / §3.3 / §7 Pattern A / [^McpHeadlessBugs]. Confidence HIGH.
- **#68375 — OPEN, zero drift** (round-5 live WebFetch). Title verbatim: "[BUG] stdio MCP tool
  call hangs indefinitely under `claude -p` when the full MCP fleet is loaded — regression in
  2.1.177, works under --strict-mcp-config." Corroborates §6 row 8 / §7 / [^McpHeadlessBugs].
  Confidence HIGH. **VOLATILITY SIGNAL (not a gap):** the issue now carries a GitHub `stale`
  label alongside `regression`/`has repro`. Still OPEN, so no defect today — but a bot
  auto-close is now a live drift risk; row-6-style "open ≠ will-be-fixed" framing is unaffected
  either way (the design owns the workaround). Future rounds should keep re-checking #68375.

## Local load-bearing leaf re-run (§6 row 4 / §8 OQ18 safety claim)
- **`git format-patch -1 -o /tmp/l4probe_r5 HEAD` → exit 0, patch written to
  `C:/Users/.../Temp/l4probe_r5/0001-*.patch`** (out-of-repo). Reproduces the R4-2 leaf claim
  blue absorbed verbatim ("git format-patch -1 -o <path> → exit 0 arbitrary out-of-repo patch,
  R4-2 leaf-verified") at §6 row 4 leg (a) and §8 OQ18(c). Confidence HIGH — blue's absorbed
  claim is faithful to the leaf. MUST-TRY: ran the leaf directly (not graded down).

## Carried HIGH (verified within ≤2 rounds, non-volatile / pin-immutable, section text stable)
- §6 row 7 / §8 OQ8 STOP figures (0.42% CI 0.31–0.57 / 0.46% CI 0.35–0.61 insignificantly
  higher / 10,000 / syntactic) — [^STOP] ar5iv §6.2/Table 2, verified r1×3+r2+r4, academic
  non-volatile. HIGH.
- §6 row 7 [^DGMSakana] (markers-removal, lineage) + [^AIScientist] (self-edit/timeout) —
  re-fetched zero-drift r4 (lens 3). HIGH.
- §6 row 1 [^SlashHeadlessIssues] #837 CLOSED COMPLETED / #14246 CLOSED DUPLICATE — closed =
  low-volatility, carried r1 (live gh classifier-blocked; story unaffected). HIGH.
- §6 row 8 #32191 CLOSED duplicate — [^McpHeadlessBugs], carried (closed, low-volatility). HIGH.
- §6 row 3 Dependabot base rate — [^Dependabot], verified r4 (lens 1). HIGH.
- §6 row 9 bootstrap guard shipped — [^HooksJson] @7bc501e pin-immutable. HIGH.
- §6 rows 11/14 cloud-routine connector-default + auth — [^RoutinesDocs], verified r1. HIGH.
- §6 row 13 / §7 R4-5 deny-reach clause ("Read and Edit deny rules apply … to file commands
  Claude Code recognizes in Bash, such as cat, head, tail, and sed") — [^PermissionsDoc],
  leaf-verified r4 (L2 + my own L4 concurrence, ledger r4); permissions doc re-fetched
  zero-drift ×4 seats r4; blue's R4-5 fix quotes the clause faithfully. HIGH.
- §8 OQ4 ~973MB models — [^QmdDaemon] backlog item 34 @7bc501e. HIGH.
- §7 Pattern A issue-status roster (#22055 CLOSED not-planned, #25621 CLOSED duplicate, #6631
  CLOSED, #66395/#23707 CLOSED not-planned) — [^PermAskBypass]/[^DenyBashIssue]/[^DenyRWIssue]/
  [^WindowsHang]/[^WebSandbox], closed statuses carried, self-audit roster accurate. HIGH.

## Archive spot-checks (floor not zero; sampled in-slice closures)
- **R2-13** (§6 row 5 cell requalified) — archive record matches report line 1589. Consistent.
- **R2-14** (§7 Pattern B/E pricing-lag bullet) — archive record matches report lines 1622–1625
  ("upgraded to leaf-verified HIGH round 1, R1-11"). Consistent.
- **R3-16** (§1.3 telemetry SHIPPED) — archive claims `trajectories/board-telemetry.jsonl`
  EXISTS; re-confirmed by Glob this round — file present. Consistent. (Sampled cross-slice.)
- archive_spot_checks: [R2-13, R2-14, R3-16]

## No downgrades
Nothing graded below HIGH in-slice this round, so no MUST-TRY impossibility lines are owed.
[^AlertFatigue] retains its standing LOW-on-number self-grade (disposition-of-record; the r3
pinnable-replacement offer remains unbanked in the NOTES section) — it is not an active
citation for any §6/§7/§8 claim (§6 row 3 cites Dependabot, not AlertFatigue), so not re-graded.

## Friction
None this round. WebFetch, live git, Glob, and archive reads all available and sufficient for
the slice.
