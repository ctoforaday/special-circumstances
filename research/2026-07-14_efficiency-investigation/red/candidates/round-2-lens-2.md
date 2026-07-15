# Red round 2 — lens 2 (leaf-node citation verification), slice 2: §2 + §3

Full living report re-read whole (1178 lines, three windowed reads — 25k cap). Slice = §2
(risk-mass spend) + §3 (grade-dispute channel / best-of-N), the same cut as round-1 lens-2.
Ledger honored: round-1 HIGH entries carried where CHANGELOG shows no section change; every
round-1-changed claim in the slice re-verified at the leaf. Ledger appended (8 entries).

## Verification summary (leaf-fetched this round)

| Claim | Source | Confidence |
|---|---|---|
| [^CathedralBazaar] all six figures (44,180; 72,122; 194/266=73%; 139/288=48%; AC/UI/Impact; 40% transfer drop) | arXiv:2607.05670 abs + html + PDF via pdftotext | **HIGH** — footnote quote verbatim in PDF. Note: arXiv-HTML garbles the key sentence to "at least 11 vector divergent" (italic-rendering artifact); PDF is authoritative and says "at least 1". The round-1 repair of R1-5 did NOT regress — checked against the repair-regression pattern deliberately |
| [^Sprt] 36–75% band | arXiv:2603.00216 abs | **HIGH** for the band; condition gap → L2-F3 |
| §2.2/§7 Fragmentation paper "genuinely paywalled (403 at the abstract)" | sciencedirect S0167404826001549 | **HIGH** — 403 reproduced first-hand at this seat |
| §2.5 item 1 "cost-audit.mjs from token metering, never sees grades" | cost-audit.mjs direct read | **HIGH** — true half |
| §2.5 item 1 sink chain (log() → journal.jsonl → cost-audit.mjs) | run-3 journal.jsonl + cost-audit.mjs | **CONTRADICTED** → L2-F1 |
| §3.5 + §0 row 3a tally (conditional vote discounted at both sites) | direct read | **HIGH** — conditional-vote laundering does not recur |
| §2.4 recomputations ($50.09; 8.0%; $12/run) | arithmetic | **HIGH** |
| §3.1/§3.3/§3.6 unchanged corpus claims (zero LEAD; $7.52–$13.56; PASS break l.236 precedes contested l.244; enum lacks grade-wrong; R3-2 three rounds unnoticed) | round-1 ledger HIGH, carried | **HIGH** |

## Findings

### L2-F1 — §2.5 item 1: the ratified telemetry's named sink does not exist as described (both halves fail at the leaf)

- **Location:** §2.5 "Disposition", item 1 — "the sink is the `log()` line into
  `trajectories/journal.jsonl`, consumed by cost-audit.mjs or the retrospective."
- **Evidence, first-hand:**
  1. Run 3's `trajectories/journal.jsonl` (the only measured journal) contains **only**
     `{"type":"started"}` / `{"type":"result"}` workflow lifecycle events — 46 + 41 lines,
     **zero** `log()` lines (grep for debate.js's own `researching:` / `debate ended` output:
     0 hits). `log()` is a harness-provided global; its output demonstrably does not persist
     to journal.jsonl. This run's `trajectories/` is currently empty.
  2. `scripts/cost-audit.mjs` contains **zero** references to journal.jsonl. Its input is the
     Workflow harness's per-agent transcript dir (header, l.7: `node cost-audit.mjs
     <workflow-transcript-dir>`); it parses `message.usage` token blocks. It cannot consume a
     journal line today without modification.
- **Why it matters:** this sentence is the round-1 REPAIR of R1-10 — and the repair relocated
  the sink error instead of fixing it (round-0 said "emit into cost.md," impossible; round-1
  says "log() into journal.jsonl, consumed by cost-audit.mjs," contradicted by the measured
  journal and the script). The ratified instrumentation half of levers 1+2 — blue's principal
  positive recommendation, and §6.2's "evidence supply for every rejected actuation's revisit
  trigger" — currently has **no verified durable sink and no actual consumer**. Downstream
  surface (other slices, flagged for merge): §1.5 item 1's board-profile `log()` line; §3.3
  clause (v), which makes accepted-dispute grade deltas "logged in the §1.5 board-profile
  log() line + spot-check-eligible" — a spot-check cannot sample a line that persists nowhere;
  §6.1 rank 4 ("instrumentation... zero tokens"). If log() output is ephemeral, runs 4–5
  produce no mass series and the deferred actuation decision has no evidence base — the exact
  record-poisoning §2.5's own mapping-stability condition (part b) worries about, one level up.
- **Grade:** likelihood **certain** (code + measured journal vs the text) × impact
  **medium-high** (the instrumentation ratification's entire evidence pipeline; three sections
  lean on it) × complexity **low** (verify where log() actually lands in the harness record,
  or specify an explicit sink — e.g. an envelope field aggregated by the lead, or a
  name-preflighted file append — and name the real consumer).
- **Required fix:** restate the sink as what is verified (cost-audit.mjs reads harness
  transcripts; journal.jsonl records lifecycle events only) and give the board-profile/mass
  line a **specified, verified persistence path + consumer** before calling the telemetry
  ratified. Propagate to §1.5(1), §3.3(v), §6.1(4), §6.2.
- **Pattern:** repair-regression (the round-1 fix's new claim fails at the leaf) ×
  policy-without-mechanism (telemetry ratified with an assumed sink).

### L2-F2 — §2.2: round-1 insertion severed the host sentence (subject-less clause survives)

- **Location:** §2.2, final sentences — "The paper originally miscited is open-access, not
  paywalled — §7's excuse is corrected there. **But was measured-robust in this loop**,
  because the adversarial loop is itself the triangulation the literature asks for."
- **Evidence:** "But was measured-robust in this loop" has no subject. The round-0 sentence
  ran "The throttle input is noisy in the general severity-grading literature (...) but was
  measured-robust in this loop"; the round-1 correction record (R1-5) was inserted between the
  two halves and orphaned the second. As written, the nearest candidate subject is "the paper
  originally miscited" — which reads as nonsense.
- **Grade:** certain × low (meaning recoverable, but the sentence carrying §2.2's bottom-line
  "noise is not the kill reason" verdict is broken text in a report about propagation
  hygiene) × trivial (reattach the subject: "The throttle input was nonetheless
  measured-robust in this loop...").
- **Pattern:** repair-regression, prose variant — a correction-record insertion breaking the
  host sentence it interrupts.

### L2-F3 — §2.3: [^Sprt] band's source condition dropped in the gloss

- **Location:** §2.3, lane-2 case — "sequential-adaptive spend cuts expected sample size by
  **36–75%** at matched error rates *when the statistic is right*."
- **Evidence:** source sentence (leaf-fetched): "Specifically, we demonstrate that for
  **symmetric error bounds**, the sequential test reduces the average sample size by at least
  36% and by at most 75%." The band is derived under the symmetric-bounds condition (type-I =
  type-II error bound); "at matched error rates" names the comparison baseline but not the
  condition — asymmetric-bounds regimes may fall outside 36–75. Nit within the same footnote:
  the quoted fragment "by at least 36% and at most 75%" drops the source's second "by" while
  wearing quotation marks.
- **Grade:** certain × **low** (the band is illustrative context in §2.3; no disposition turns
  on its scope — surfaced, not blocking) × trivial (add "for symmetric error bounds" to the
  gloss or the footnote; restore the dropped word inside the quotation).
- **Pattern:** within-source condition misattribution (mild instance — condition dropped, not
  reassigned).

## Explicitly checked, clean

- **[^CathedralBazaar] repair did not regress** (memory pattern says round-1 re-citations to
  new sources are the highest-risk claims): all six figures verified against the PDF full
  text; body gloss "median of at least one diverging CVSS base metric" is a fair reading of
  "median of at least 1 vector divergent" (vector-Divergency = Hamming distance over base
  metrics). One adjective unsupported but harmless: "public CNAs" — the paper says CNAs.
- **§3.5 honest tally** (R1-16 fix): the lane-2 conditional vote is discounted at both §0 and
  §3.5 — no conditional-vote laundering.
- **§2.2 withdrawal record** internally consistent with round-1's two red fetches; §7's
  paywall rewrite matches (403 reproduced first-hand on the Fragmentation paper).
- **§3.3 clauses (v)–(vii), (vi)'s line-number claims, §3.6's corrected $7.52–$13.56 band,
  R1-18's ~$10/firing price** — all consistent with round-1 HIGH ledger entries; no drift.

## Lens verdict contribution

Slice §2–§3 citations are sound after round 1 — the two big round-1 repairs (R1-5 recitation,
R1-26 band) both verify at the leaf. The one substantive defect is L2-F1: a certain
text-vs-code contradiction sitting under the run's flagship "instrumentation before mechanism"
recommendation, itself a repair-regression on R1-10. Not PASS-blocking for the slice's
evidence base, but the telemetry ratification should not survive round 2 with an unverified
sink. L2-F2/L2-F3 are trivia.
