# Red — Round 4, Lens 4 (logic & completeness: leaps of faith, missing counterarguments, unexplored alternatives, template compliance)

**Surface audited:** full `blue/report.md` (1802 lines) re-read in context, focused on the Round-3
§14 responses (the answers to R3-1..R3-17) and their interaction with the standing design. Leaf-node
checks on the committed-projection / native-`@`-import path.

**Lens-4 disposition of the R3 gaps I own (4/L4 items):** R3-7 CLOSED (parser-derived taint is
sound — §14.1 session corollary; the self-report is genuinely removed from the trust path). R3-9
CLOSED (Threat A / Threat B separated cleanly, §14.5). R3-10 CLOSED (typing reclassified
surface-neutral/defense-enabling, §13.7). R3-11/R3-12 ADDRESSED, assembly-deferred (§8 forward
pointer + §14.8 operative-decisions table). R3-8 reconciled **but contingent** — see R4-1. **New R4
lens-4 gaps below.**

---

### R4-1 (lens 4 R4) — LEAP OF FAITH / COMPLETENESS: the clone-injection invariant is a *policy* with no *enforcement mechanism* against native `@`-import; "not new machinery — a removal of trust" is false [severity HIGH, blocking-candidate]

- **Location:** §14.1 — *"On clone/pull/merge of any store whose commits are not locally authored, every concept **loads clamped to reference/candidate tier**; its committed `status`/tier/`review_count` are **reset to candidate baseline**."* and *"This is **not new machinery** — it is a *removal* of trust."* Cross-read against §12.2 (accepted premise, never retracted) — *"`CLAUDE.md` `@`-imports the attacker's `active.md` at active authority with **no install step**"* and the withdrawn mechanism at line 807-808 — *"activation is gated on a **local, git-ignored ratification marker**."*
- **Problem:** the invariant states *what* must be true (committed tiers never inherit; the projection loads at reference tier) but the current design (§14.1/§14.8) names *no mechanism* that makes it true at the one moment it matters. The projection (`active.md`) is a **committed file loaded by native `@`-import at session open**, before any bespoke `/dream` runs. Three would-be enforcers all miss it:
  1. **The import corollary** ("every concept loads clamped... fields reset") describes a *bespoke re-derivation* over `knowledge/*.md` concepts. But native `@active.md` is resolved by the harness, not by `/dream`; nothing bespoke runs to "clamp" or "reset" anything at first open of a fresh clone. The reset applies to concept frontmatter at the *next local `/dream`* — which has not run yet.
  2. **mit.5 unconditional de-authorization** is a *generator-side* rendering property — it governs how *blue's* projector words concepts into a projection. On a clone the projection bytes were authored by the **attacker**, who writes imperatives directly into the committed `active.md`; native `@`-import loads them verbatim. De-authorization never touches attacker-authored projection bytes.
  3. **A SessionStart hook** cannot help: hooks are unreliable headless (§1.3), and even interactively `additionalContext` is *added* — it cannot *un-import* an `@`-imported committed file.
  The only concrete enforcement ever specified — the **git-ignored ratification marker** — lived in the **withdrawn §12.2** and was *not* carried into §14.1. Across three rounds the enforcement has been progressively hollowed: §12.2 had a concrete (flawed) git-ignored gate → §13.2 replaced it with an authorship check → §14.2 demoted authorship to "nudge-convenience, not activation" → §14.1 now asserts the outcome ("loads clamped") with *no* stated gate. The sentence "not new machinery — a removal of trust" is the leap of faith: enforcing "a committed, natively-`@`-imported projection loads at reference tier" is **precisely new machinery** (git-ignore the projection, or intercept the load), not a removal.
- **Required fix:** name the enforcement mechanism. The natural terminus of the invariant — **commit only raw concept bodies (`knowledge/*.md`); git-ignore the projection *and* every trust-elevating frontmatter field; regenerate the projection and re-derive all tiers locally** — makes R1-2/R2-1/R3-1/R3-2 structurally moot. Adopt it explicitly *and price its cost* (the projection no longer "travels with the repo" — the committed-store differentiator §13.7 sells shrinks to concepts-only, reviewable but not directly loadable). OR specify a reliable session-open interception. Do not present the invariant as mechanism-free while a committed `active.md` is still natively imported.
- **Grade:** likelihood high (a fresh clone's committed projection is imported at first open by construction) · impact high (zero-click active-authority load of attacker projection bytes — the original R1-2 vector, un-closed at the mechanism level) · complexity-to-fix low-medium (the git-ignore-projection decision is cheap; the cost is to a marketed differentiator, which must be stated). Corroboration of "not new machinery": contradicted by the committed-projection / native-import path. **Pattern: policy-without-mechanism (an invariant asserted as self-enforcing while the enforcing artifact was withdrawn).**
- **Amplifier:** R3-5 and R3-8 closures both rest on "the import clamp makes breadth-driven cloning safe-by-default." If R4-1 stands (no enforcing clamp), those closures are contingent, not complete.

### R4-2 (lens 4 R4) — TEMPLATE/ACCOUNTING: the verdict's "31 items, 5 blocking" cannot be reconciled against the superseding rows — the true blocking count is ~6 [severity LOW-MEDIUM]

- **Location:** Verdict — *"Consolidated required changes are in §8 (31 items, **5 blocking** — Round-2 fixes are items 21–27, Round-3 items 28–31)."* against §14.9 item 29 — *"**Blocking** (security; supersedes item-22 turn-level self-report)."*
- **Problem:** the original blocking five are §8 items 1, 2, 3, 15, 16. Item 21 supersedes 15; item 28 supersedes 21 (chain = one slot). But **item 29 is graded Blocking and supersedes item 22, which was graded *High*** — so a non-blocking row was replaced by a blocking one, adding a slot the "5" never counted. Net operative blocking set = {1, 2, 3, 16, 28, 29} = **6**. The headline "5 blocking" is stale, and — because the blocking rows are scattered across §8 / §13.11 / §14.9 with supersessions — the count is no longer verifiable from any single surface (the §14.8 operative table lists *decisions*, not a blocking tally).
- **Required fix:** recompute and state the operative blocking count once (reconciling supersessions), or add a blocking-set line to the §14.8 table. If 29 is meant to *fold into* 16 (both are provenance-of-content/taint) rather than stand alone, say so and drop the double-count.
- **Grade:** likelihood certain (present in the text) · impact low-medium (the go-decision headline miscounts the gating set) · complexity trivial. **Pattern: supersession-accounting drift (grade changed under a superseding row; headline count not re-derived).**

### R4-3 (lens 4 R4) — LEAP OF FAITH: the R3-6 recurring flag-check assumes a detectable "native-consolidation signature" for an unverified server-side feature [severity LOW-MEDIUM]

- **Location:** §14.7 — *"each `/dream` invocation detects Auto Dream's consolidation signature (**e.g. `MEMORY.md` mutated since last `/dream` by a writer other than `/dream`, or a native-consolidation marker/metadata**) and **stands down or re-scopes accordingly**."*
- **Problem:** the fix that closes R3-6 (volatile flag) depends on `/dream` reliably distinguishing native-Auto-Dream-consolidated `MEMORY.md` from bespoke-consolidated. Both proposed discriminators are speculative: (a) "mutated by a writer other than `/dream`" needs authorship, but `MEMORY.md` lives at `~/.claude/projects/<project>/memory/` — **not** in the project git repo, so there is no commit-authorship to read; distinguishing *native* mutation from *manual operator edits* or *other tooling* is unspecified. (b) "a native-consolidation marker/metadata" is asserted for **Auto Dream, which is on blue's own §10 Unverified list** — its output format and whether it leaves any marker are unknown. So the recurring check rests on a signature blue cannot confirm exists. The one-time→recurring upgrade is the right *direction*, but the detection primitive is hand-waved for a feature whose behavior is unverified.
- **Required fix:** state the detection primitive as an **unverified Phase-0 dependency** (test empirically whether Auto Dream leaves a distinguishable signature; if not, the recurring check degrades to a heuristic and the two-writer residual is not fully closed), rather than presenting "detect the signature" as a settled mechanism.
- **Grade:** likelihood medium (Auto Dream behavior unknown) · impact medium (undetected two-writer churn if the signature assumption fails) · complexity-to-fix low (relabel as tested-dependency + fallback). **Pattern: leap of faith on an unverified external feature's observable behavior.**

### R4-4 (lens 4 R4) — TEMPLATE/COHERENCE: the Heilmeier §0 headline still markets the *automatic* promotion ladder that §14.3 demoted to a near-empty-set convenience [severity LOW]

- **Location:** §0 Q3 — *"a promotion ladder (capture → corroborate → promote → decay) made physical as git commits"* (presented as the headline "what is new"), and Q5 — *"Learning compounds across projects."* Cross-read with §14.3 — auto-promotion *"downgraded from a load-bearing feature to a convenience that operates only on (i) **fully-untainted sessions** (no external read at all) and (ii) operator-confirmed concepts."*
- **Problem:** red's own established fact (R2-3, accepted by blue) is that *near-every real working session performs a `WebSearch`/`WebFetch` or external read* — so "fully-untainted sessions" is a near-empty set, and automatic corroborate→promote for trajectory-derived concepts now fires on approximately nothing; all durable promotion is operator-gated (`/remember` or human review). The §0 Heilmeier framing — the deliverable-facing pitch — still advertises the *automatic* ladder as the differentiating novelty, unreconciled with §14.3's concession. (The go-decision survives: §13.7 already made *human-gated* promotion the load-bearing value, so this is not a build-case defect — it is a marketing/coherence lag in the template-required section.) Trivial adjacent note: the report title (line 1) still reads *"living, **Round 1**"* three rounds on.
- **Required fix:** at assembly, reconcile §0 Q3/Q6 with §14.3 — frame the ladder as *capture → corroborate → **human-gated** promote → decay*, and state that unattended auto-promotion is a convenience over untainted sessions only; correct the "Round 1" title.
- **Grade:** likelihood certain (present) · impact low (Heilmeier over-sells vs operative design; go-decision unaffected) · complexity trivial. **Pattern: headline-lag (a template section not re-reconciled after a downstream concession narrowed the feature).**

---

## Lens-4 verdict for this pass: FAIL (CHANGES-REQUIRED)

The §14 invariant is the right *organizing idea* and genuinely closes the self-report/soundness holes
(R3-7/R3-9 closed; R3-3 honestly risk-accepted). But it is presented as self-enforcing when the
enforcing artifact was withdrawn two rounds ago (**R4-1**, blocking-candidate): a committed,
natively-`@`-imported projection has no mechanism clamping it to reference tier at first open. Three
lower-severity lens-4 gaps stand: the blocking count no longer reconciles (**R4-2**), the recurring
flag-check leans on an unverified signature (**R4-3**), and the Heilmeier headline over-sells the
demoted auto-ladder (**R4-4**). No new lens-4 gap is closed, rebutted, or risk-accepted this pass →
PASS unavailable on this lens.

## Friction
- Still no full-PDF-text / PDF-table extraction tool — carried from prior rounds (blocks definitive
  closure of R1-19/R1-28-residual). Not load-bearing for this lens's findings.
- No way to confirm Auto Dream's actual output behavior from here (§10 Unverified) — directly limits
  how hard R4-3 can be pressed; I can flag the assumption but not test it. A sandbox with the
  server-side flag toggled would let me verify whether a "native-consolidation signature" exists.
