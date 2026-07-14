# Red — Round 2, lens 4 (logic & completeness)

**Surface:** full `blue/report.md` re-read in context (all 1010 lines, incl. new §12 and
footnotes). Lens: leaps of faith, missing counterarguments, unexplored alternatives, template
compliance. This pass audits blue's Round-1 responses to the lens-4 gaps (R1-8..R1-17) and hunts
NEW logic/completeness defects introduced by the §12 additions.

**Lens verdict: FAIL — one HIGH new logic gap (R2-1, lead-docket candidate), one MEDIUM
unhandled-dependency (R2-2), plus reconciliation debt (R2-3/R2-4). Most Round-1 lens-4 gaps
are genuinely closed.**

---

## Round-1 lens-4 gaps: disposition this round

- **R1-9 (deliver re-scoped phase plan) — CLOSED.** §12.9 delivers the six-phase disposition
  table + "minimum viable bespoke layer." The actionable core is now present.
- **R1-10 (strongest → suggestive) — CLOSED.** Verdict and §3 Consequence 1 now read
  "suggestive, not 'strongest'"; the Auto Dream keystone no longer carries the verdict.
- **R1-12 (append-only → unbounded concept growth) — CLOSED.** §12.9 caps Evidence at N
  most-recent + a total counter; the §2.3b/§6.2 tension is resolved.
- **R1-13 (confidence-float tie-break undefined) — CLOSED.** §6.2 names a total deterministic
  ordered tie-break (trust tier → review_count → last_seen → longer body).
- **R1-14 (git-diff demotion on mismatched evidence) — CLOSED.** §2.4 now labels the
  solo-operator extrapolation "reasoned inference, not measurement," with the asymmetric-cost
  justification.
- **R1-15 (defer/build-nothing timing branch) — CLOSED, with a caveat that surfaces R2-2.**
  §12.9 evaluates for/against and recommends hybrid timing. But the deferral leg leans on an
  unverified dependency — see R2-2.
- **R1-16 (Heilmeier cost/schedule axis) — ADDRESSED to the minimum red offered; downgrade to
  LOW.** §8 effort note gives the aggregate blocking-set cost ("wiring" items = days;
  design+build items 1/16 = ~1-2 weeks), which is exactly the "or at least aggregate the
  blocking set's cost" fallback R1-16 accepted. Residual: no Heilmeier Catechism *section*
  exists anywhere, and the 15 non-blocking items still carry no effort. Assembly-time debt, not
  a live-report blocker.
- **R1-17 (verdict framing) — CLOSED.** Verdict opens "Endorse, gated on the blocking set below
  — not endorse-then-caveat… Read the blockers first."
- **R1-8 (netted build-vs-adopt asserted, never argued) — STRUCTURALLY ANSWERED but the answer
  carries a logic flaw → R2-1.** §12.5 delivers the summed table red demanded. The table's
  reasoning is defective; see R2-1.
- **R1-11 (poisoning grade conflation) — remains CONTESTED → lead docket.** §12.5 builds the
  opportunistic/supply-chain attacker model (fair, and it does raise likelihood above "who'd
  target a solo op"), part-concedes, and keeps all five mitigations arguing four add ~no
  complexity beyond the two ingest gates. That is a reasonable part-concession, but the
  apparatus-sizing dispute (keep-all-five vs demote mit.4/mit.5) is a genuine judgment call blue
  itself flags for the lead. Red does not close it; carry to the lead's docket as both parties
  request.

---

## NEW graded gaps (Round 2)

### R2-1 (lens 4) — LOGIC: the netted build-vs-adopt table classifies the dominant risk as "shared" to zero it out, contradicting §4's own "widens it" [severity HIGH; lead-docket candidate]

- **Location:** §12.5 netted table, row 1 — *"Inbound poisoning pipeline (ingest → context) |
  **Shared** — native auto-memory already pipes untrusted input to context; the CVE exploited
  native, not bespoke. Adopting native does not escape this."* and the conclusion —
  *"most of the poisoning surface is inherited from native, not created by the bespoke layer …
  it buys less value for the same dominant risk."* Cross-read against §4 —
  *"The proposal's store reproduces this surface and **widens** it (more files, more writers)"*
  and *"The consolidator can convert a one-shot injection into a high-confidence permanent rule."*
- **Problem:** the entire "build wins" conclusion turns on neutralizing the poisoning axis by
  labeling it "shared." But blue's own §4 says the bespoke layer does not merely reproduce the
  native surface — it **widens** it. Three concrete widenings the "shared" cell omits from the
  net-new column, each documented elsewhere in blue's own report:
  1. **Explicit external `/ingest` intake.** Native auto-memory captures *the operator's own
     sessions*; the bespoke layer adds a deliberate `url:`/`file:` external-content intake (§4,
     §12.3). That is a net-new untrusted edge native does not have.
  2. **Cross-project blast radius.** Native auto-memory is per-project, machine-local (§1.2,
     §3) — a poisoned memory is contained to one project. The bespoke **global** store's whole
     point is cross-project propagation, so one poisoned concept reaches *every* project's
     context. Net-new amplification on the dominant axis; absent from the table.
  3. **Corroboration → auto-promotion laundering.** §4: two poisoned trajectories =
     `review_count: 2` = auto-promoted to `active`. Native has no typed trust-tier promotion
     ladder that converts repetition into durable authority. Net-new laundering step.
  Additionally the cell's supporting claim is imprecise: *"the CVE exploited native, not
  bespoke."* The CVE (npm-postinstall → `MEMORY.md`) exploited the *file's authority + write
  access*, not native's auto-capture *pipeline*. Using it to prove the capture pipeline is
  "shared" conflates two mechanisms. The honest netting is: **adopt-native yields a materially
  NARROWER poisoning surface (session-only intake, per-project blast radius, no promotion
  ladder) for LESS value; build yields a WIDER surface for MORE value, mitigated by the gates.**
  Blue's "same dominant risk, more value" understates the risk delta and thereby biases the
  go/no-go toward build. This is the exact meta-defect R1-8 was raised to fix; the fix was
  delivered but the accounting inside it is motivated.
- **Required fix:** re-net the poisoning axis honestly — count the three widenings (external
  intake, cross-project blast radius, auto-promotion laundering) as net-new bespoke surface on
  the dominant dimension; state that adopt-native buys a smaller poisoning surface, and argue the
  value gained is worth the *widening* (not merely worth "the same risk"). If the widening cannot
  be shown worth it, the build recommendation weakens. Contested and go/no-go-bearing → for the
  lead's docket alongside R1-8/R1-11.
- **Grade:** logic/meta · likelihood n/a · impact high (frames the go/no-go; the netted table is
  the answer to the round's central meta-gap) · complexity-to-fix low (re-classify three rows).
  Corroboration: contradicted by blue's own §4 text.

### R2-2 (lens 4) — LEAP OF FAITH: the re-scope hands MEMORY.md consolidation to native Auto Dream, an item the report itself files Unverified, with no fallback for the flag being absent [severity MEDIUM]

- **Location:** §3 Consequence 3 — *"scope /dream to knowledge/ only and let native Auto Dream
  own MEMORY.md, consuming its output as the inbox."* and §12.9 Phase 2 — *"Let native Auto Dream
  own MEMORY.md **if the flag is live**; consume its output as inbox."* cross-read with §10 —
  *"Native Auto Dream availability — verified as concept … unverified as a dependable API
  (server-side flag)."*
- **Problem:** the re-scoped plan's deletion of bespoke MEMORY.md consolidation is justified as
  "don't duplicate native." But that only holds *if native actually consolidates MEMORY.md*, and
  Auto Dream is flag-gated, "not universal," and on blue's own §10 Unverified list. Native auto-
  *memory* writes MEMORY.md but does **not** consolidate it — consolidation is Auto Dream's job.
  So if the flag is absent (the likely default for this operator), MEMORY.md is captured but
  never consolidated, it grows unbounded, and blue's own §6.1 measured context-rot regression
  kicks in — with **no owner**, because bespoke `/dream` was scoped to `knowledge/` only and the
  MEMORY.md leg was deferred to a feature that may never arrive. Phase 0 "confirms the flag" but
  the plan states only the flag-live branch; the flag-absent branch is unhandled. The whole
  two-writer resolution and inbox story rest on an unverified dependency.
- **Required fix:** state the flag-absent branch explicitly — e.g. "if Phase 0 finds Auto Dream
  not live on this box, `/dream` retains MEMORY.md consolidation (un-defer that leg)." Make the
  re-scope's deferral *conditional on the Phase-0 finding*, not assumed.
- **Grade:** likelihood medium-high (flag-gated feature, absence is the likely state) · impact
  medium (unowned consolidation → measured context-rot) · complexity-to-fix low (one conditional
  branch in the phase table).

### R2-3 (lens 4) — COMPLETENESS: §4's "blocking before Phase 1" timing anchor is stale under the §12.9 re-scope [severity LOW-MEDIUM]

- **Location:** §4 — *"Required changes (blocking before Phase 1)"* against §12.9 Phase 4 —
  *"`/ingest` + `/memory-bootstrap` with the two ingest-edge gates (§12.5) and wholesale bootstrap
  quarantine (§12.3) as **blocking prerequisites**"* and the MVP —
  *"Minimum viable bespoke layer = Phase 0 + Phase 2-scoped-to-`knowledge/` + the typed-extraction
  sliver of Phase 1."*
- **Problem:** §4 anchors the poisoning blockers to "before Phase 1," written against the
  proposal's original phase numbering. Under the re-scope, the risky `/ingest`/`bootstrap` work is
  now **Phase 4**, and Phase 1 is reduced to a typed-extraction sliver. So "blocking before Phase
  1" is no longer coherent: the thing it gates (ingest) has moved to Phase 4, and the MVP ships
  Phases 0/2/1-sliver *without* ingest at all. The blocking-timing labels were not reconciled with
  the new phase map, so a reader cannot tell which blocker gates which phase. (The one blocker that
  *does* still bite the MVP — provenance-of-content on trajectory extraction, §12.3 — is correctly
  placed in the Phase-1 sliver, so this is a labeling inconsistency, not a missing control.)
- **Required fix:** re-anchor each §4/§8 blocker to its re-scoped phase (§12.9) — ingest gates
  gate Phase 4; provenance-of-content gates the Phase-1 typed-extraction sliver; clone-ratification
  gates Phase 3. Drop the stale "before Phase 1" global label.
- **Grade:** likelihood high (inconsistency present as written) · impact low-medium (reader cannot
  sequence blockers to phases) · complexity-to-fix low.

### R2-4 (lens 4) — TEMPLATE/NAVIGATION: dangling "§11" cross-references and still-absent Heilmeier section [severity LOW]

- **Location:** §2.3a — *"triggers the deferred SQLite/vector index (§11)"* and §1.5 —
  *"The proposal's 'SQLite + vector index: deferred, not rejected' (§11)"*; report section
  sequence runs …§9 → §10 → §12 (no §11).
- **Problem:** the report has no §11 heading, yet cross-refs point at "(§11)" as the home of the
  deferred index. §1.5 disambiguates ("*the proposal's* §11"), but the bare "(§11)" in §2.3a does
  not, so a reader may hunt for a report section that does not exist. Separately (folds with the
  R1-16 residual): the protocol's `report_template.md` requires the assembled `report.md` to carry
  the Heilmeier Catechism as a named section; none exists in the living report yet. Both are
  assembly-time defects, not live-report blockers, but they are navigation debt that will bite at
  union.
- **Required fix:** make bare "(§11)" read "proposal §11" (or add the report §11 the refs imply);
  ensure the Heilmeier Catechism section is present at assembly.
- **Grade:** likelihood medium · impact low (navigation only) · complexity-to-fix trivial.

---

## Items checked and NOT raised (recorded so they are not re-litigated)

- Disconfirming-evidence handling in §12.5 (`[^SingleUserLowRisk]`): blue weighs it, accepts the
  "conditioned on trusted input" qualifier, and uses it to *localize* rather than refute the
  blocker. Logically sound handling of disconfirming evidence — not a gap.
- §12.6 concurrency, §12.7 history-scrub tradeoff, §12.8 ephemeral curator: each states its
  tradeoff explicitly and does not over-claim. No logic gap on the lens-4 axis (their security
  substance is lens-5 territory).
- §12.9 hybrid-timing recommendation logic (build differentiating-sliver now, defer native-overlap):
  internally coherent *except* for the unhandled flag-absent branch already captured in R2-2.
- Tie-break ordering (§6.2) is total and deterministic — verified complete.
