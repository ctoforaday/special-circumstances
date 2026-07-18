# Red round-2 — lens 4 (leaf-node citation verification)

Slice: §6 (risk matrix) / §7 (pre-flight self-audit) / §8 (open questions) + owned footnote [^HooksJson].
Full living report re-read in context (whole document, consecutive windows). Ledger read; round-1 HIGH
verifications honored except where drift-prone (issue trackers) or section changed.

## Method / re-verification log

- **Owned footnote [^HooksJson]** (§6 row 9, "Bootstrap guard already shipped in hooks.json"): pinned at
  7bc501e, verified HIGH round 1 (ledger). Pin unchanged, §6 row 9 unchanged this round → not re-fetched.
- **§6 row 8 OPEN-bug drift check (issue trackers = volatile source):** re-fetched the two OPEN issues the
  row rests on. #76239 — LIVE state OPEN; title "SDK headless: MCP tools silently missing on first turn when
  stdio server startup is slower than the (new) non-blocking pre-wait — regression … since CLI 2.1.144";
  matches §3.3/§6 row 8. #68375 — LIVE state OPEN; title "[BUG] stdio MCP tool call hangs indefinitely under
  claude -p when the full MCP fleet is loaded … works under --strict-mcp-config"; matches. Both HIGH, no
  drift. #32191 CLOSED-duplicate (round-1 HIGH, correction R1-5) not re-fetched — closed-terminal, low drift.
- STOP figures (§7 round-1 update, §8 OQ8), issue statuses #837/#14246/#23707/#66395/#22055/#6631/#25621,
  Dependabot base rate (§6 row 3), pricing (§5→referenced in §6/§7): all verified HIGH/graded round 1; the
  round-1 dispositions hold on re-read. No leaf-node source mismatch found in the §6 risk rows or §8 OQ text.

## Findings

### L4-F1 — §7 self-audit understates pricing citation confidence (stale MEDIUM; repair-lag)
- **Location:** §7 Pre-flight self-audit, Pattern B/E bullet — "*STOP percentages flagged for PDF re-pin;
  pricing figures graded MEDIUM with canonical source named; no number laundered.*"
- **Defect:** The round-1 citation correction R1-11 upgraded the pricing figures from aggregator-carried
  MEDIUM to leaf-verified HIGH — §5.2 states "leaf-verified against the platform pricing page at red's
  round-1 audit, upgrading round 0's aggregator-carried MEDIUM (R1-11)" and [^Pricing] says
  "leaf-fetched at red's round-1 audit … upgrading round 0's aggregator-carried MEDIUM (R1-11)." The §7
  self-audit still characterizes pricing as MEDIUM, and §7's own "Round 1 update" paragraph enumerates
  banked upgrades (STOP re-pin; [^UsageAPI] and [^AIScientist] MEDIUM→HIGH; issue statuses) but **omits the
  pricing MEDIUM→HIGH upgrade entirely.** A reader auditing confidence grades from §7 alone would carry a
  stale MEDIUM for a figure the report elsewhere now asserts HIGH.
- **Pattern:** incomplete-repair / footnote(narrative)-lag — a correction propagated to the body (§5.2) and
  the footnote ([^Pricing]) but not to the self-audit prose that inventories citation confidence.
- **Corroboration confidence for the underlying citation:** HIGH (pricing itself is leaf-verified round 1;
  no re-fetch triggered — pricing page is volatile but round-1 leaf grade stands, section unchanged, <2
  rounds elapsed). The defect is the §7 self-report of that grade, not the grade.
- **Grade:** LOW (likelihood the stale line misleads a downstream reader = low; impact = cosmetic/internal
  consistency; complexity-to-mitigate = trivial — one clause + add R1-11 to §7's round-1 upgrade list).
- **Required fix (blue):** in §7, either strike "pricing figures graded MEDIUM" or append "(upgraded to
  leaf-verified HIGH round 1, R1-11)"; add the pricing R1-11 upgrade to §7's Round-1 banked-upgrades list
  alongside STOP/[^UsageAPI]/[^AIScientist].

## No-gap confirmations (for merge signal)
- §6 rows 4/10/13/14/15/16 (round-1 added/re-graded) reference only claims verified elsewhere/round 1; the
  L×I×Cx gradings and dispositions represent the cited evidence faithfully. No mis-citation.
- §6 row 8 OPEN-bug framing re-confirmed live (#76239, #68375 both OPEN 2026-07-17).
- §8 OQ8 STOP figures match the round-1 ar5iv pin (0.42% CI 0.31–0.57%; 0.46% CI 0.35–0.61%; insignificantly
  HIGHER; 10,000; syntactic detection).
