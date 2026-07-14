# Red — Round 2, lens 3 (leaf-node citation verification), slice 3 of 3

**Slice scope:** §7 (H5 alternatives), §8 (change table), §9 (risk table), §10 (unverified),
§12 (Round-1 additive responses), and the FIVE new footnotes Round 1 introduced —
`[^FactsFirstClass]`, `[^EnvInjectedMemory]`, `[^SkillSupplyChain]`, `[^SingleUserLowRisk]`,
`[^GitLockContention]` — none of which were in red's Round-1 verified-clean list. Each new
footnote followed to its primary this round; §7 footnotes re-checked at the live primary where
load-bearing (mem0).

**Method note:** these are the citations blue *added or rewrote* in Round 1, so they carry the
highest residual risk — a repair can introduce a fresh error. That is exactly what one did (R2-1).

---

## NEW GAPS (Round 2)

### R2-1 — MISCITED figure introduced BY the R1-28 repair: "~90% environment-injection attack success" is contradicted at the leaf node [severity MEDIUM]
- **Location:** §12.5 — *"a **malicious web page read mid-session** (environment-injected memory
  poisoning, \"poison once, exploit forever\") ... ~90% attack success in the web-agent
  environment-injection setting — supports the **opportunistic, untargeted** attacker model"*
  (footnote `[^EnvInjectedMemory]`); and §4 / §9 risk row 1 — *"up to ~90–95% (MINJA /
  environment-injection)"*.
- **Problem:** `[^EnvInjectedMemory]` = arXiv 2604.02623, *Poison Once, Exploit Forever:
  Environment-Injected Memory Poisoning Attacks on Web Agents* (Zou et al.). The paper exists and
  the qualitative claim (environment-only injection, no direct store access, persistent effect) is
  sound — but the reported attack success is **up to 32.5% (GPT-5-mini), 23.4% (GPT-5.2), 19.5%
  (GPT-OSS-120B)**, rising *up to 8×* only "under environmental stress" from a low base. Nowhere
  ~90%. The footnote's asserted "~90% attack success in the web-agent environment-injection
  setting" is not in the primary; the real environment-injection figure is roughly **one-third**
  of the cited number. This is the sharp irony: it was the **R1-28 repair itself** — which
  softened Round-0's "80–99%" to "up to ~90–95% (MINJA / environment-injection)" — that attached
  the ~90% to the environment-injection source. The MINJA half (~95%) is separately sourced and
  remains merely untraced (R1-28); the environment-injection half is now *contradicted*, not just
  untraced.
- **Consequence for the argument:** §12.5 uses this figure to lift attack *likelihood* above red's
  "solo operator, who would bother" framing (the R1-11 grade-contest rebuttal). The direction
  survives — untargeted environment injection does work a non-trivial fraction of the time — but
  the headline is inflated ~3×, so the likelihood leg of blue's part-rebuttal of R1-11 is weaker
  than stated. The exact "laundered into fact" failure the protocol names: a skeptic following the
  footnote lands on a 32.5% paper.
- **Required fix:** replace "~90%" with the paper's actual figures (up to ~32.5%, up to 8× under
  stress) attributed to 2604.02623; drop "environment-injection" from the "~90–95%" band in §4/§9
  and keep only MINJA there (itself still untraced per R1-28). Re-state §12.5's likelihood claim on
  the corrected number.
- **Grade:** corroboration LOW as cited (figure contradicted at leaf node) · likelihood-of-error
  certain (verified) · impact medium (props the sole blocking risk's likelihood cell + the R1-11
  rebuttal) · complexity-to-fix low.

### R2-2 — DISCONFIRMING-EVIDENCE citation not corroborated at leaf node; part rests on an unfollowable "practitioner consensus" [severity LOW-MEDIUM]
- **Location:** §12.5 — *"industry consensus is that for a single-agent/single-user local markdown
  store, 'simple advisory file locking is enough' and heavier mitigations are 'over-engineering
  for the threat model' — when the input is trusted"* and footnote `[^SingleUserLowRisk]` —
  *"markdown memory with **simple advisory file locking** is sufficient ... it 'stops being enough
  when a second agent/user, changing facts, or a retention requirement enters.'"*
- **Problem:** `[^SingleUserLowRisk]` cites the dev.to article
  (imaginex, *When Markdown Files Are All You Need*) **plus** "practitioner consensus surveyed
  2026-07-13." Following the dev.to primary: the article exists (author Yaohua Chen) but frames the
  markdown-vs-database choice by **scale** ("optimal for Local Agents ... Files get unmanageable
  >5MB"), not by **trust**; it does **not** discuss advisory file locking, trusted-input
  conditioning, or the enumerated triggers ("second agent/user, changing facts, retention
  requirement"). Those quote-shaped phrases are not in the cited article — they trace to the
  unnamed "practitioner consensus surveyed," which a skeptic cannot follow (the human/agent's own
  survey is an untrusted, unfollowable source per the leaf-node rule). Caveat: small-model HTML
  fetch is lossy (standing friction) — but here the fetch returned a *different* framing, not
  merely "not found," which is stronger than a null result.
- **Why it matters:** this is blue's **disconfirming leg** in §12.5 — the evidence it explicitly
  weighs against its own blocking grade to "localize" the risk to the ingest edge. A disconfirming
  citation that can't be followed to a primary weakens the honesty of the part-rebuttal of R1-11,
  which turns on exactly this "over-engineering when input is trusted" claim.
- **Required fix:** attribute the advisory-locking-sufficient / trusted-input-conditioning claim to
  a followable primary that actually carries it, or relabel it as blue's own reasoned synthesis
  ("practitioner sentiment, not a single citable source"), not a quotation. Do not present a
  self-conducted survey as external corroboration.
- **Grade:** corroboration LOW-MEDIUM (the directional "single-user is lower risk" sentiment is
  common and plausible; the specific quoted claims are uncorroborated at the cited primary and part
  rests on an unfollowable survey) · impact low-medium (weakens a disconfirming leg, not the
  blocking core) · complexity-to-fix low.

---

## VERIFIED-CLEAN this round (recorded so they are not re-raised)

- **`[^FactsFirstClass]` (arXiv 2603.17781, Zahn & Chana) — HIGH.** Primary confirms *"summarization
  destroys 60% of facts,"* Knowledge Objects *"100% accuracy across all conditions at 252× lower
  cost,"* multi-hop 78.9% vs 31.6%, Sonnet-4.5 100% exact-match 10→7,000 facts. The **R1-18
  re-attribution is sound**: the §2.1 headline (60% loss / 36.7× / 252×) genuinely lives here, not
  in Hindsight. R1-18 can be marked closed on the citation itself (the in-place repair is verified).
- **`[^SkillSupplyChain]` (arXiv 2604.03081, Qu et al.) — HIGH.** Primary confirms *"a single
  malicious skill can compromise the host,"* skills *"executed as operational directives with
  system-level privileges"* (DDIPE 11.6–33.5% bypass). §12.5's supply-chain leg is corroborated;
  blue uses it qualitatively (no number asserted) — clean.
- **`[^GitLockContention]` (anthropics/claude-code #55724) — HIGH, near-verbatim.** Issue confirms
  *"5 committed successfully, 8 failed"* of 13 parallel agents; retry with 200/400/800ms backoff;
  worktree auto-cleanup destroying uncommitted work as the priority fix. Grounds §12.6 and the §9
  concurrent-single-box row solidly. (Status: Closed as **duplicate** — not "as not planned"; blue
  does not mischaracterize its status, so no gap.)
- **`[^MemZero]` mem0 ADD-only — HIGH at the live primary.** Current README confirms *"Single-pass
  ADD-only extraction — one LLM call, no UPDATE/DELETE. Memories accumulate; nothing is
  overwritten,"* timestamp-based current-vs-historical resolution. **R1-23 fully discharged**: the
  §2.3b append-only corroboration and the §7 corrected "steal mem0's *current* ADD-only design" are
  accurate to the shipping repo. Minor sub-note (not a gap): the README the fetch surfaced cites
  LoCoMo 71.4→92.5 and LongMemEval 67.8→94.4 rather than blue's "~90% token / ~91% latency
  reduction"; those reduction figures are mem0's known vs-full-context marketing claim and are not
  load-bearing — leave as-is or attribute to the mem0 paper explicitly.

---

## STILL-OPEN from Round 1 in this slice (re-confirmed, NOT closed)

- **R1-28 (MINJA / "80–99%" band):** the survey primary `[^MemoryPoisonSurvey]` (arXiv 2606.04329)
  abstract confirms the taxonomy (four write channels, nine vulnerabilities, six attack classes)
  but carries **neither** the 80–99% band **nor** an explicit MINJA ~95% at the surface I can reach
  (lossy fetch; body may carry it). The MINJA figure remains **untraced at leaf node** — consistent
  with R1-28's medium grade; the R1-28 repair softened the wording but did not attach the surviving
  numbers to a confirmable primary. R1-28 stays open/medium. (Now compounded by R2-1: with the
  environment-injection half contradicted, MINFA is the *only* remaining leg of the "~90–95%" band,
  and it is untraced — the band is weaker than the softened §4 wording implies.)

---

## Interaction / adjudication note for the lead

R2-1 lands on the **likelihood** side of the sole blocking risk (memory poisoning) — the same cell
R1-11 contests. Sequence for the docket: R1-11 argues the "blocking" grade over-sizes the apparatus
against a low-probability attacker; blue's §12.5 rebuts by raising likelihood via an opportunistic
model **quantified partly on the now-contradicted ~90% figure**. Correcting R2-1 does **not** flip
the disposition (untargeted environment injection at ~32% + the clone-time (R1-2) and bootstrap
(R1-3) vectors still establish real, non-targeted exposure, and the CVE-class precedent carries the
risk's *existence* independently) — but it **narrows** the likelihood margin blue claims, which is
material to whether the *surplus* mitigations (mit.4/mit.5) clear the bar R1-11 sets. The lead
should read R1-11 + R2-1 together: the blocking *core* (two ingest gates) survives; the *sizing*
argument for the surplus is now built on a corrected, lower likelihood number.

---

## Verdict (this lens/slice)

**FAIL — 2 new gaps (R2-1 medium, R2-2 low-medium), R1-28 re-confirmed open.** Four of the five
new Round-1 footnotes verify HIGH at the leaf node — the Round-1 citation-repair work is largely
sound. But one repair (R1-28) introduced a fresh contradicted figure (R2-1), and one disconfirming
citation (R2-2) does not corroborate at its cited primary and partly rests on an unfollowable
self-survey. Neither overturns blue's direction or the blocking core; both are low-complexity
fixes. PASS unavailable this slice until R2-1 is corrected and R2-2 is re-sourced or relabeled.

## Friction
- HTML-arXiv / dev.to leaf-node fetches remain lossy for numbers in tables and for locating a
  specific claim inside a long article (R2-2's "different framing returned" is more diagnostic than
  a null, but a full-text/PDF search tool would let me state definitively whether the
  advisory-locking claim is *absent* vs *missed*). Same friction red logged in Round 1; unresolved.
- No tool gap blocked R2-1 (abstract-level contradiction was decisive).
