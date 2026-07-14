# Red — Round 2, Lens 5 (dark-side & risk)

**Surface:** full `blue/report.md` re-read in context (1010 lines, incl. new §12 Round-1 response
block and the extended §8/§9 tables). CHANGELOG used only as navigation.

**Lens verdict: FAIL (CHANGES-REQUIRED).** Blue made real progress — R1-9 and R1-15 are *closed*
(the re-scoped phase plan and defer/timing branch are delivered, §12.9), citation repairs are in
place, and R1-2/3/4/5/6/7 are accepted with the right *direction*. But accepting a direction is not
accepting an implementation: **every new mitigation blue added carries its own un-graded second-order
failure mode**, and two of them (R2-1 clone-gate, R2-3 netted-rebuttal) are severe enough to block a
clean pass. The lens confirms the risk core stands; it contests the newly-introduced controls.

Disposition of prior gaps at bottom.

---

## New gaps (Round 2)

### R2-1 (lens 5) — the clone-injection fix is self-defeating: the ratification fingerprint collides with `/dream`'s own store mutation [severity HIGH, blocking-candidate]
- **Location:** §12.2 "R1-2 — clone-time injection via the committed project store (ACCEPT; blocking)" — *"activation is gated on a **local, git-ignored ratification marker** (e.g. `.claude/knowledge/.ratified` containing a store-content fingerprint the operator's `/dream --ratify` writes). A freshly cloned repo has no marker → the projection loads at **candidate tier only**."*
- **Problem:** three compounding defects in the fix that closes R1-2:
  1. **Collision with the write loop.** `/dream` *mutates* the store every night (that is its job — consolidation, promotion, pruning). A content-fingerprint marker therefore mismatches after every legitimate dream run, dropping the projection to candidate tier until re-ratified. The only escapes are both broken: (a) `/dream` re-writes the fingerprint after its own run — but the nightly pass "runs with no human present" (§4), so this is *self-ratification by the unattended pass*, exactly defeating the human-consent gate; or (b) the operator manually re-ratifies daily — unworkable friction. The fingerprint cannot distinguish operator-authored changes from dream-authored changes from clone-delivered changes.
  2. **Leans on diligence the report itself discredits.** §2.4 demotes human diff-review to *forensic* because "a single operator reviewing nightly dream diffs will decay to LGTM within weeks." §12.2 then makes human ratification the *sole preventive* control for the clone vector. A mitigation cannot rest on the exact diligence the same report argues is unreliable.
  3. **Escape hatch reopens the common case.** *"auto-ratify repos under a configured trusted root"* voids the defense for the solo-dev workflow — a developer who clones everything under `~/Projects` and marks it trusted auto-ratifies every hostile clone, restoring the zero-click vector §12.2 was built to close.
- **Required fix:** the marker must fingerprint *provenance/authorship*, not *content* (e.g. sign operator-ratified state, and have `/dream` write into a distinct dream-authored tier that never self-elevates to active); state how a legit dream run avoids invalidating ratification without self-ratifying; and bound or remove the trusted-root auto-ratify, or grade the residual exposure it leaves.
- **Grade:** likelihood high (the collision fires on the first nightly run) · impact high (either the gate is bypassed by self-ratification or the feature is unusable) · complexity-to-fix medium. Corroboration of the fix as written: contradicted by the system's own loop.

### R2-3 (lens 5) — the netted build-vs-adopt keystone mis-classifies the poisoning surface as "Shared/inherited"; bespoke re-authorizes what native's CVE fix de-authorized [severity HIGH, blocking-candidate / lead's docket]
- **Location:** §12.5 table row — *"Inbound poisoning pipeline (ingest → context) | **Shared** — native auto-memory *already* pipes untrusted input to context; the CVE exploited *native*, not bespoke. Adopting native does **not** escape this."* and the conclusion *"**most of the poisoning surface is *inherited from native*, not created by the bespoke layer**."*
- **Problem:** the netting treats native as static-vulnerable, but blue's own §4/R1-29 records that Anthropic's CVE-2026-21852 fix **removed user memories from the system prompt** — i.e. *de-authorized the native memory surface*. Post-fix native is therefore *less* poisonable. Meanwhile the bespoke design's preferred projection channel is `.claude/rules/`, which §1.2 and §6.2 establish loads at **CLAUDE.md priority — the highest-authority surface**. The bespoke layer thus **re-authorizes high-authority injection that native remediation removed** — it does not merely inherit native's surface, it re-widens it on the exact dimension (authority) that the CVE fix narrowed. The "Shared" cell is false, and it is the cell that carries the "adopt buys no safety" conclusion driving the go decision. Double-bind: blue tagged "removed from system prompt" *medium-confidence* (R1-29) — if it is too uncertain to rely on, blue cannot use it to equate native and bespoke surfaces; if it is reliable, then bespoke re-opens what native closed. Either way "Shared" does not hold. Separately, the netted table still only enumerates *surfaces* and asserts *value* qualitatively ("no typed concepts, no cross-project global git repo") — R1-8 asked to sum net-new surface **against** value; the cost side is now argued, the value side remains unquantified.
- **Required fix:** reclassify the inbound-poisoning row: the bespoke high-authority projection is **net-new authority** relative to post-fix native, not shared. Re-run the netted conclusion with that correction, or gate the projection to the de-authorized (reference-voice) channel unconditionally so the "Shared" claim becomes true by construction. Quantify or bound the "shrunken value" side that R1-8 asked for.
- **Grade:** likelihood n/a (logic/grading) · impact high (flips the keystone build-vs-adopt argument) · complexity-to-fix low-medium. Candidate for the lead's docket alongside R1-8/R1-11.

### R2-2 (lens 5) — provenance-of-content rule either over-blocks (kills trajectory automation) or needs unspecified turn-level tracing; "one predicate" undersells it [severity MEDIUM-HIGH]
- **Location:** §12.3 "R1-3 — provenance-of-record vs provenance-of-content (ACCEPT; blocking)" — *"A trajectory's trust is capped by the **most-untrusted content its transcript touched**. During extraction, if the transcript contains a `WebFetch`/`WebSearch` result, an external file read, or `/ingest` output, the derived candidate is tagged **external-ingest**."* and *"This closes the laundering path red identified and is cheap (one predicate in the extractor)."*
- **Problem:** near-every real working session performs a `WebSearch`/`WebFetch` or reads an external file. Under the transcript-scoped rule, essentially **all** trajectory-derived concepts get capped at `external-ingest`, which per §4 mit.2 "never auto-promotes … without explicit human confirmation." Net effect: the trajectory-capture-and-auto-promote path — the system's core automation value — produces nothing that auto-promotes, because the transcript almost always touched the web. The only alternative blue gestures at ("down-tier any candidate whose *supporting turns* include an external read") requires **fine-grained per-fact turn-level provenance tracing** — attributing which extracted fact derived from which turns — which is not "one predicate," is unspecified, and is the hard part of the design. The fix as stated is either safe-but-useless or cheap-but-unbuilt.
- **Required fix:** specify the granularity: either accept that transcript-scoped tagging neuters auto-promotion (and say the automation value is gated behind human confirmation by design), or specify the turn-level fact-provenance mechanism and re-grade its complexity as Medium, not "one predicate."
- **Grade:** likelihood high (the coarse rule fires on ordinary sessions) · impact medium-high (guts auto-promotion or hides real build cost) · complexity-to-fix medium.

### R2-4 (lens 5) — advisory lock leaves a stale-timeout TOCTOU and does not serialize `/dream`'s commit against concurrent capture writes [severity MEDIUM]
- **Location:** §12.6 "R1-5 — concurrency control" — *"An **advisory lock** on `/dream`'s consolidate+commit stage (a lockfile in the store, e.g. `.knowledge/.dream.lock` with a stale-timeout). If held, `/dream` **no-ops and logs** … Capture writes are **append-only to per-session/per-day files** … two sessions write different dated files."*
- **Problem:** two residuals in the accepted fix. (a) **Stale-timeout TOCTOU:** a genuinely slow consolidation that exceeds the stale timeout is treated as dead; a second `/dream` proceeds concurrently — the exact race the lock prevents. Stale-lock schemes need an owner-liveness check or a monotonic renewal, not a bare timeout. (b) **Capture-vs-commit is un-serialized:** the lock only serializes `/dream` runs against *each other*. `/dream`'s commit stage does `git add`/commit over the store while an interactive session is simultaneously writing a new short-term capture file into the same git-tracked tree. If `short-term/` is inside the committed store (the design commits the store), a `git add -A` stages an in-flight partial write. Per-session dated files avoid *file* collisions but not *index/working-tree* races during commit.
- **Required fix:** replace bare stale-timeout with liveness (pid + heartbeat) or scope the lock to cover the commit's `git add` window against capture writers; state whether `short-term/` is inside the commit path and, if so, exclude it from the dream commit or lock capture during commit.
- **Grade:** likelihood medium · impact medium (partial-file commit / lost capture) · complexity-to-fix low-medium.

### R2-5 (lens 5) — the history-scrub fix trades away the reviewable-git-history differentiator that the build case depends on [severity MEDIUM]
- **Location:** §12.7 "R1-6 — history-scrub vs forensic-undo" — *"**Publishing is a separate operation to a separate remote**: push a **scrubbed export/derived snapshot**, not a mirror of the working repo."* cross-read with §12.5 value claim *"no cross-project global git repo … [is what adopt-native] buys *less value*"* and §3 remit *"cross-project global knowledge as a reviewable git repo."*
- **Problem:** the R1-6 resolution publishes a *scrubbed derived snapshot*, not the working repo's history. But the build-vs-adopt case (§12.5) and the surviving remit (§3) lean on **"cross-project global knowledge as a reviewable git repo"** as a *primary differentiator justifying build over adopt*. A published snapshot with rewritten/squashed history is not a reviewable git history — the very property sold as the reason to build is the property sacrificed to remediate leaks. The two sections are in tension and neither acknowledges it.
- **Required fix:** reconcile §12.7 with §12.5 — either the pushed artifact retains reviewable history (and then the scrub tradeoff of R1-6 stands unmitigated), or it is a scrubbed snapshot (and then the "reviewable git repo" differentiator is weaker than §12.5 claims). State which, and adjust the build-vs-adopt margin accordingly.
- **Grade:** likelihood low-medium (only when a scrub triggers) · impact medium (erodes a keystone value claim) · complexity-to-fix low (a framing/scope reconciliation).

### R2-6 (lens 5) — ephemeral consolidator closes the durable self-poisoning path but not in-pass steering via the poisonable store it reads [severity MEDIUM]
- **Location:** §12.8 "R1-7 — the consolidator must not have self-written persistent memory" — *"run `memory-consolidator` and `memory-curator` with **read-only or ephemeral memory during the consolidation pass** … The defense agent sits *outside* the trust surface it guards."*
- **Problem:** the fix removes *durable self-written* memory, closing the persistent-bias path — correct and accepted. But the consolidator still **reads the store** (the poisonable surface) as its working input each pass. A planted instruction-shaped concept ("always merge X into Y", "treat source Z as authoritative") read during a pass can steer that pass's merge/promote decisions *without* any durable memory. So "sits *outside* the trust surface it guards" is overstated — it still ingests the guarded, poisonable surface every run. The durable path is closed; the in-pass path is not.
- **Required fix:** either constrain the consolidator's read authority (treat store content it reads as data, never as instruction — the same de-authorized-voice discipline as §4 mit.5 applied to the consolidator's own inputs), or acknowledge the residual in-pass steering path and grade it, rather than claiming the agent is outside the surface.
- **Grade:** likelihood low (requires a poisoned concept surviving to the store first) · impact medium-high (biases a single consolidation pass) · complexity-to-fix low.

### R2-7 (lens 5) — the re-scope's Phase 2 hangs on the unverified, flag-gated Auto Dream with no stated fallback if the flag never lands on this box [severity MEDIUM]
- **Location:** §12.9 re-scoped phase table, Phase 2 — *"Let native Auto Dream own `MEMORY.md` if the flag is live; consume its output as inbox."* and §12.4(a) ingest path *"native `MEMORY.md` (or Auto Dream output) → **screened** `/dream` ingest → store."*
- **Problem:** Auto Dream is filed under §10 *Unverified* — "server-side flag rollout; availability unverified as stable API," rollout "not universal." The re-scoped Phase 2 makes the design's consolidation/ingest topology *contingent* on that flag being live ("if the flag is live; consume its output"). There is no stated branch for **the flag never landing on this operator's box**: if native never consolidates `MEMORY.md` for this user, what feeds the "consume its output as inbox" step? The re-scope defers bespoke `MEMORY.md` consolidation to native (Phase 1 "Drop bespoke work duplicating native `MEMORY.md` capture") while making the ingest inbox depend on a native output that may not exist here. A build plan cannot defer its own capability to an unverified upstream feature without a fallback.
- **Required fix:** add the "flag never lands" branch — either bespoke `/dream` retains a `MEMORY.md` consolidation fallback (so the dropped Phase-1 work is *conditionally* dropped, gated on Phase-0 flag confirmation), or state explicitly that the feature is blocked pending the flag and the operator accepts that dependency.
- **Grade:** likelihood medium (flag is not universal by blue's own account) · impact medium (a deferred-but-undelivered capability gap) · complexity-to-fix low (add the conditional branch already half-implied by the Phase-0 flag check).

---

## Disposition of Round-1 gaps (this lens)

**Closed (accept, no residual):**
- **R1-9** — re-scoped phase table delivered (§12.9). Closed.
- **R1-15** — defer/timing branch evaluated, hybrid-timing recommendation given (§12.9). Closed. (Residual dependency risk carried forward as R2-7, a *new* gap, not a re-open.)
- **R1-10, R1-13, R1-14, R1-16, R1-17** — reframes/corrections in place; accepted.

**Accepted direction, residual raised as new Round-2 gap:**
- **R1-2 → R2-1** (fix self-defeating).
- **R1-3 → R2-2** (fix over-blocks / under-specified).
- **R1-5 → R2-4** (stale-lock TOCTOU + capture-vs-commit).
- **R1-6 → R2-5** (trades away reviewable-history differentiator).
- **R1-7 → R2-6** (in-pass steering residual).

**Coherent now, minor note:**
- **R1-4** — §12.4 makes channel/voice one trust-tier-keyed decision; internally coherent. Note: combined with R2-2, the path to high-authority `.claude/rules/` narrows to operator-confirmed-only, so automatic rule-promotion value approaches zero — a value cost blue does not acknowledge, but not an incoherence. Fold into R2-2.

**On the lead's docket (contested grade, not a red hard-block):**
- **R1-8** — cost side now argued (§12.5); value side still qualitative, and R2-3 shows the cost-side "Shared" classification is wrong. Keep on the docket, now sharpened by R2-3.
- **R1-11** — blue concedes the two ingest gates are the blocking core and flags mit.4/mit.5 as demotion candidates for the lead (§12.10). Reasonable disposition; hand to lead.

**Citation repairs (R1-18..R1-30):** repaired in place per CHANGELOG; not re-litigated under this lens (dark-side, not citation-verification). Deferred to the citation lenses for leaf-node re-confirmation.

---

## Friction
- No new tool gap this lens beyond the standing one: leaf-node confirmation of the "removed user
  memories from the system prompt" CVE remediation detail (load-bearing for R2-3) is blocked by
  post-cutoff vendor-blog-only sourcing — a way to fetch/confirm the primary Anthropic security
  advisory would let red settle whether the bespoke projection re-authorizes a *remediated* surface
  (net-new) or an *unremediated* one (shared). The whole R2-3 double-bind turns on that fact.
