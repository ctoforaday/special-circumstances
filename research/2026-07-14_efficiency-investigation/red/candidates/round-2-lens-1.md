# Red audit — round 2, lens 1: leaf-node citation verification (slice 1 of 4: preamble + §0 + §1 + §2, with their footnotes)

Full living report re-read in context (all 1178 lines). Ledger consulted; round-1 HIGH
verifications held (≤1 round old, no drift trigger). Audit surface this pass: blue's round-1
edits in slice 1 (R1-1, R1-4/15, R1-5, R1-10, R1-12, R1-17, R1-20, R1-23/24/26, R1-33,
R1-35(a)) plus the new [^CathedralBazaar] citation — every repair re-verified as a new claim.

## Verified this pass (leaf-node, first-hand)

- **[^CathedralBazaar] (new round-1 citation, §2.2)** — arXiv:2607.05670 abs + /html/v1
  leaf-fetched: title/authors (Siqi Zhang, Fabio Massacci, Mengyuan Zhang) verbatim; "44,180
  CVEs under the Pairwise setting" and "72,122 CVEs under the Consumer-View setting" verbatim;
  "194/266 (73%) and 139/288 (48%) of CNAs have a median of at least 1 vector divergent"
  verbatim; divergence concentrates in Attack Complexity, User Interaction, Impact; "accuracy
  can drop by 40%" verbatim. **HIGH.** The R1-5 replacement citation is sound — the withdrawal
  record's replacement figures all trace.
- **[^AdaptiveStability] repaired gloss (R1-20)** — arXiv /html/2510.12697v1 re-fetched:
  criterion "Dt<0.05 for 2 consecutive rounds" verbatim; Table 2 stops at rounds 4–7 across
  enumerated configs; deltas vs fixed 10 rounds span −1.03%..0.00, largest loss −1.03%
  (Gemma-3-4B/BIG-Bench). "stops at rounds ~4–7 losing ≈1% or less in the reported
  configurations" — **HIGH** on the delta bound and criterion. On the "~4–7" band: my fetch
  enumerated only the BIG-Bench/JudgeBench rows (4–7) and its prose hinted "4–8"; the ledger's
  lens-4 entry (twice-fetched) shows Table 2's JudgeAnything rows stop at 2, 2, 8 — the full
  measured span is 2–8 and "~4–7" excludes reported configurations. I DEFER the band to that
  stronger evidence and surface the conflict to the merge rather than hold my softer read
  (unquoted-hold discipline): band **LOW as written**, convergent with lens-4's L4-F1.
- **§1.2 "~$78" counterfactual (R1-4)** — recomputed from cost.md: round 3 =
  $2.98+$9.46+$12.64 = $25.08 (~$25 ✓); rounds 4–5 = $53.00 ✓; sum ~$78 ✓. **HIGH.**
- **§1.2 "no threshold setting reproduces the round-3 stop"** — sound within the
  severity×fix-cost family: round-2 and round-3 boards are identical on both axes
  (MEDIUM-HIGH max, all complexity "low," per ledgered grades), so any floor admitting round 3
  fires at round 2 first. **HIGH** (logic on ledgered premises).
- **§1.5 live-code correction (R1-1)** — debate.js ll.236–258 direct read: PASS break (236)
  precedes contested filter (244–245); judge dispatched same round when contested non-empty
  (247–250); `carried` never enters `adjudicated` (252–253); ids accumulate at 258 so any gap
  surviving two rounds dockets. Every element of the correction paragraph verifies. **HIGH.**
- **§1.5 heartbeat attribution (R1-21 repair)** — ledger-held: backlog line 31, STILL OPEN,
  substance true at the corrected location. **HIGH** (ledger, round 1).
- **§2.3 "all three citation instances" for R4-1** — run-3 round-4 candidate headers direct
  read: lenses 1/2/3 are "leaf-node citation verification" instances, lens 5 "dark-side and
  risk"; R4-1 minted by 1/2/3/5 ⇒ all three citation instances ✓. **HIGH.**
- **§2.3 blue-respond series / second-smallest-board framing (R1-23)** — cost.md:
  $3.95/$3.96/$2.98/$3.05/$4.27, round 5 highest ✓ on the 6-open board vs round-4's 5 ✓. **HIGH.**
- **§2.3 36–75% band (R1-26)** — ledger round 1 (lens 2): cited arXiv source states 36–75%;
  the corrected text now quotes exactly that. **HIGH** (ledger).
- **§2.4 recomposition (R1-33, R1-35(a))** — cost.md: red-lens Σ rounds 1–5 = $49.48 ✓ (33% of
  $149.95 ✓); killed r6 spawn row = $0.61 ✓; seat total $50.09 ✓; per-round r3–r5
  $9.46/$10.47/$11.05 ⇒ 2-of-5 lens cut ≈ $3.8–4.4/round ⇒ $12/run ≈ 8% ✓. **HIGH.**
- **§2.5 item 1 first half (R1-10)** — debate.js no-filesystem doctrine (ledgered ll.32–34)
  ✓; cost-audit.mjs direct read: parses `agent-*.jsonl` usage records (token metering only) —
  never sees grades ✓. **HIGH** on those two sub-claims; see L1-F1 for the rest.
- Slice-1 claims unchanged since round 1 (board table, backlog quotes, R4-1/R5-5/R5-1 grade
  headers, mass tables, CvssInconsistent 68%, stop-resume, ceiling table, pin equivalence):
  **ledger-held HIGH**, not re-fetched (no drift trigger, <2 rounds).

## Findings (lens-scoped ids; merge assigns stable ids)

### L1-F1 — the ratified instrumentation's sink and consumer are both unsubstantiated; the pinned corpus contradicts the sink claim
- **Location:** §2.5 Disposition, item 1 (and the same sink underlies §1.5 RATIFY item 1's
  board-profile line and §0 rows 1–2's "RATIFY the instrumentation/telemetry").
- **Challenged sentence:** "the sink is the `log()` line into `trajectories/journal.jsonl`,
  consumed by cost-audit.mjs or the retrospective."
- **Evidence (first-hand):** (a) pinned `cost-audit.mjs` reads only `agent-*.jsonl` files from
  the workflow transcript dir (usage/token records); it never opens `journal.jsonl` — the named
  consumer cannot consume the named sink as shipped. (b) Run 3's
  `trajectories/journal.jsonl` — the corpus's only journal instance — contains 87 lines,
  exclusively `{"type":"started"}` / `{"type":"result"}` dispatch records; debate.js's existing
  `log()` output (l.52 "researching: …", l.271 "debate ended: …") is absent (grep count 0).
  The pinned record therefore affirmatively suggests harness `log()` does NOT persist to
  journal.jsonl.
- **Why it matters:** instrumentation-before-mechanism is blue's load-bearing RATIFY for
  levers 1+2; §2.5 item 3's revisit condition and §1.5's telemetry ratification both assume a
  durable, reviewable mass/board-profile series from runs 4–5. If `log()` is
  console-ephemeral, the ratified telemetry collects nothing, and the deferred-actuation
  decision arrives with no evidence base — the exact policy-without-mechanism failure this
  report flags elsewhere (§4.5 condition 7's own standard). The R1-10 repair relocated the
  sink from one impossible location (cost.md emission) to an unverified one the pinned record
  contradicts — a repair-regression.
- **Grade: MEDIUM** — likelihood medium-high (two independent contradicting observations; the
  harness's log() persistence contract is uninspectable from this seat, so a newer-harness
  journal schema could conceivably capture it — that unknown is the only thing keeping this
  below high) × impact medium (evidence base for both deferred actuations silently never
  materializes; discovered only at run-5 decision time) × complexity-to-fix low (verify
  first-hand where harness `log()` persists; if ephemeral, the spec must name a durable,
  git-tracked sink — e.g. a tracked telemetry file appended by the merge seat, or an explicit
  journal-write harness feature — and name a consumer that actually reads it).
- **Required fix:** replace the asserted sink/consumer with a verified persistence path, or
  mark the sink an open design item with a MUST-verify-before-run-4 condition. Do not re-cite
  cost-audit.mjs as consumer unless it is extended to parse the journal (state the extension).

### L1-F2 — R1-5 repair left a subjectless sentence in §2.2
- **Location:** §2.2, immediately after the correction record.
- **Challenged sentence:** "But was measured-robust in this loop, because the adversarial loop
  is itself the triangulation the literature asks for."
- **Evidence:** the clause has no subject; the round-1 rewrite of the ~34%-withdrawal passage
  orphaned it (the round-0 sentence's subject — the throttle input / grade noise — was edited
  away). Certain (text defect, read directly).
- **Grade: LOW** — certain × low × trivial (restore the subject: "But the throttle input was
  measured-robust…"). Incomplete-repair artifact; flagged because this is the same
  edit-regression class the report itself audits blue for.

### L1-F3 — "genuinely paywalled" overstates what a 403 shows
- **Location:** §2.2 correction record (echoed in §7 and [^ConflictingScores]).
- **Challenged sentence:** "…is genuinely paywalled (403 at the abstract) and remains
  unverified — not re-cited."
- **Evidence:** a 403 at a ScienceDirect abstract URL evidences bot-blocking from this seat,
  not a paywall — ScienceDirect abstracts are normally publicly readable; the journal
  (Computers & Security) is subscription, so the conclusion is plausible, but the stated
  mechanism ("403 at the abstract" ⇒ "genuinely paywalled") is the same
  access-failure-as-paywall conflation this run's own R1-5 record corrects in the other
  direction. Downstream handling is honest (unverified, cited for no figure).
- **Grade: LOW** — medium likelihood (the paper may in fact be fetchable elsewhere; e.g. an
  author preprint) × low impact (nothing rests on it) × trivial fix ("inaccessible from this
  seat (403); unverified" instead of "genuinely paywalled").

## Out-of-slice handoff (for the instance covering §5 / footnotes — evidence gathered incidentally, not a slice-1 finding)

- **[^PropagationChains] half-misquote:** the footnote asserts the quoted phrase "5 chains in
  5 rounds" lives at blue-researcher.md l.14 AND debate.js l.263. First-hand: blue-researcher.md
  l.14 reads "5 chains in run 3's 5 rounds" ✓; debate.js l.263 reads "5 **regressions** in
  5 rounds" — the phrase lives in one of the two named sources; the other is a paraphrase.
  Footnote-overattribution class; trivial fix.

## Observations (checked, no gap raised)

- §2.1 item 4 "mean mass per gap stays ~5 across all five rounds": actual per-round means span
  4.4–6.0 on the disclosed tables. Loose, but the item's conclusion (sum(L×I) cannot
  discriminate certain×low nits from medium×medium risks) turns on the 3.5-vs-4 comparison,
  not on the ~5 constant. Recorded so the merge knows it was checked, not missed.
- §1.5 degenerate-FAIL guard reference: confirmed live at debate.js l.225.

## Slice verdict input to merge

Slice 1's citation base is now fully leaf-verified: the R1-5 replacement citation
([^CathedralBazaar]) traces at HIGH on a third independent fetch; every round-1 dollar repair
in the slice recomputes correctly; the live-code correction verifies line-by-line. One round-1
repair regresses: the R1-20 "~4–7" band excludes reported configurations (deferred to lens-4's
stronger double-fetch — convergent, not a new finding from this seat). Slice-1 findings: one
MEDIUM (L1-F1 — sink/consumer of the ratified instrumentation contradicted by the pinned
record; independently convergent with lens-2's L2-F1), two LOW (repair artifacts). No HIGH.
Nothing in this slice blocks; L1-F1 must be answered before §2.5/§1.5's instrumentation
ratification can be called evidence-backed.

## Friction

- The harness's `log()` persistence contract is uninspectable from this seat: debate.js calls
  a harness-provided `log()` whose destination is documented nowhere in the pinned corpus; I
  could only test negatively (run-3 journal.jsonl lacks log output). A harness-behavior
  reference (where `log()` lands, what journal.jsonl's schema captures) would convert L1-F1
  from strong inference to direct check.
