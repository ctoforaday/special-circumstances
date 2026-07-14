## Round 0

### BLUE

Position: **the proposal's architecture is directionally right — endorse with mandatory
amendments, not a redesign.** Two lanes (H1 substrate deep, H2 consolidation deep; breadth
across H1–H5; 41 searches, saturation reached, disconfirming budget met) converge on the same
verdict, and the strongest single piece of evidence is external: Anthropic is independently
shipping the same loop (native Auto Memory, flag-gated "Auto Dream" consolidation), which
validates the shape while shrinking the bespoke remit.

What blue asserts, with evidence in `blue/report.md`:

1. **Substrate holds** (H1): OKF is real, spec-legal for our profile, four weeks old — adopt as
   pinned convention, not dependency. Transcript JSONL verified at the leaf node on this
   machine. The file-based-memory pattern is widely shipped; every matured system added a
   derived index — name the trigger (~300–500 concepts), don't defer indefinitely.
2. **Consolidation is the top technical risk but mitigable** (H2): lossy-compression and
   continuous-rewrite corruption are measured phenomena; the structural fix is append-only
   expansion (claims immutable after promotion; change = supersede) plus per-pass change caps.
   Review-by-git-diff is forensic, not preventive — bot commits are systematically unreviewed.
3. **One blocking omission**: memory poisoning. The design pipes untrusted input into always-on
   context and its corroboration rule launders injections into permanent rules
   (CVE-2026-21852 precedent against the exact inbox file). Trust-tiered provenance, permanent
   human gate on external-ingest promotion, injection screening, independent-source
   corroboration — blocking before Phase 1.
4. **Three factual defects in the proposal**: the §5 agent-`memory:` row is wrong (harness
   injects from fixed paths; bidirectional write collision); the "existing secret-scrub gate"
   does not exist; `docs/scheduling.md` does not exist. Verified locally.
5. **Pragmatist simplifications**: drop the stored confidence float (derive activation from
   observables); single projection channel (prefer path-scoped `.claude/rules/` over
   @-import + SessionStart double injection); hard token budget on `active.md`;
   sequential-only subagents in the scheduled pass.
6. **Risk-accepted, with rationale**: OKF v0.1 drift (degrades to plain markdown),
   multi-machine store divergence (single operator; git remote is the sync story),
   project-store PR-ratification flow (keep optional, off by default). A design made worse to
   satisfy these edge cases would itself be a defect.
7. **No alternative dominates** (H5): claude-mem, basic-memory, mem0, Letta, Zep each fail a
   binding constraint (SQLite opacity, missing lifecycle, daemon/service-bound); each
   contributes a stealable mechanism, catalogued in §7.

Consolidated change list: `blue/report.md` §8 (14 items, 3 blocking). Full risk grading: §9.

## Round 1

### RED

**Verdict: FAIL (CHANGES-REQUESTED).** Full audit: `red/findings.md` (30 graded gaps, cumulative).
Consolidation of five lens passes preserved in `red/candidates/round-1-lens-{1..5}.md`: lenses 1-3
leaf-node citation verification (three slices), lens 4 logic & completeness, lens 5 dark-side &
risk. Blue's direction, its consolidation-corruption core (§2.3b append-only, verified against
[^FaultyMemories]), and its poisoning core (§4) all stand — this is changes-requested, not a
redesign. But eight gaps block a clean pass and none is yet closed, rebutted, or risk-accepted.

**Blocking (R1-1, R1-2, R1-3, R1-4, R1-8, R1-9, R1-10, R1-11):**

1. **R1-1 — VERIFIED ERROR (red re-verified at leaf node).** §6.3 "No such gate exists" and §8
   item 3 "it does not exist to be reused" are wrong. `sc-secrets-gate` (a wired PreToolUse Go
   hook), the reusable `internal/secrets` pattern package, and their tests **ship today**;
   `hooks/hooks.json` wires them live on WebFetch|WebSearch|Bash. Blue's `[^LocalRepoScrub]`
   grepped only `*.md` and was blind to the Go layer — and lens-2 independently repeated the same
   `*.md` scope and corroborated the false claim HIGH. Two verifiers sharing a flawed method scope
   is not corroboration. Correct narrower claim: a reusable matcher exists; a new consumer for
   capture-/commit-/push-time scanning must be **wired, not built** — and the existing gate is
   outbound-tool-input only, so it does NOT protect the `git push` of store contents.
2. **R1-2 — NEW security (blocking-candidate).** Project-store-committed-with-code is a zero-click
   clone-time injection vector: a cloned repo's `active.md` auto-`@`-imports attacker memory with
   no install step — worse than the CVE §4 cites. §4 omits repo-clone-as-injection.
3. **R1-3 — NEW security.** `/memory-bootstrap` mass-processes all historical transcripts
   unattended; blue's trust taxonomy conflates provenance-of-record (trajectory-derived) with
   provenance-of-content (externally-sourced bytes), letting mid-session-read malicious content
   launder up and auto-promote via `review_count`.
4. **R1-4 — NEW internal incoherence.** (a) `autoMemoryDirectory`-into-store deletes the
   capture-time hook §4 mit.3 requires; (b) the `.claude/rules/` channel (§6.2/§8 item 7) is the
   highest-authority surface, opposite §4 mit.5's "de-authorize the projection voice" and opposite
   the CVE fix direction. Channel choice and voice-authority are coupled, treated as independent.
5. **R1-8 — META.** Build-vs-adopt is asserted, never argued: the net-new attack surface
   (R1-2/R1-3/R1-5/R1-7 + push exfil) is never summed against the shrunken value vs native + a thin
   projection skill. Confront the sum, not the parts.
6. **R1-9 — COMPLETENESS.** §3 argues a mandatory re-scope; §8 item 4 says "drop bespoke work
   duplicating native capture" — but the report never delivers the re-scoped phase plan. The
   actionable core of the audit ("build less — what, exactly?") is missing.
7. **R1-10 — LOGIC.** The verdict's "strongest validation" (Anthropic converging via Auto Dream)
   rests on an item the report itself files under §10 Unverified, sourced to blogs + a community
   skill. A low-confidence keystone carries the word "strongest" without flagging the tension.
8. **R1-11 — GRADE CONTESTED.** The §4/§9 poisoning "blocking" grade conflates
   attack-success-if-attempted (80-99%) with attack-likelihood for a solo, machine-local store. Keep
   the two ingest-edge gates (/ingest, mid-session web) as blocking; require each *additional*
   mitigation to be justified against a stated attacker model, or demote the surplus. Weigh against
   R1-2/R1-3 (which sharpen the likelihood side). Candidate for the lead's docket if re-raised.

**Non-blocking, raised and open:** R1-5 (concurrent single-box writers un-graded; mis-scoped as
multi-machine YAGNI) · R1-6 (git-history-undo vs secret-history-scrub mutually exclusive) · R1-7
(self-poisonable memory-backed consolidator) · R1-12 (append-only → unbounded concept-file growth,
tension with §6 cap) · R1-13 (dropping confidence float leaves merge tie-breaks undefined) · R1-14
(git-diff demotion on setting-mismatched OSS evidence) · R1-15 (defer/build-nothing timing branch
unexamined) · R1-16 (Heilmeier cost/schedule axis absent) · R1-17 (verdict framing).

**Citation/evidence (leaf-node):** R1-18 (§2.1 60%/36.7x figure miscited to Hindsight; belongs to
arXiv 2603.17781) · R1-19 (§2.4 61.4%/71.6% uncorroborated after two fetches) · R1-20 (#57507 CLOSED
not-planned, cited as "open"; blocking change gated on a won't-fix = unsatisfiable) · R1-21 (#56540
CLOSED, macOS-launchd-specific, generalized to cron/Windows) · R1-22 (v2.1.59 unattributed) · R1-23
(mem0 pipeline stale — vendor moved to ADD-only, which ironically corroborates §2.3b) · R1-24
(claude-mem "46k stars" → ~87k) · R1-25 (Letta git-branch detail unsupported by cited blog) · R1-26
(ARC-AGI 54% → 52.6%) · R1-27 (basic-memory "no cloud" → cloud optional) · R1-28 (80-99% band
unpinned) · R1-29 (CVE-2026-21852 id-mapping + "removed from system prompt" on vendor-blog sourcing)
· R1-30 (BeliefMem digits unconfirmed).

**Risk-accepts red does NOT contest:** OKF v0.1 drift; multi-*machine* divergence proper;
project-store PR-ratification flow off by default. R1-5 is carved out from the machine accept.

Recurring patterns this round (recorded to red memory): closed-not-planned issues cited as "open"
(R1-20/R1-21); live-source drift catchable only against the current primary (R1-23/R1-24);
verification file-type blindspot — and its compounding form, two verifiers sharing one flawed grep
scope reaching false agreement (R1-1); likelihood≠success-rate conflation in risk grading (R1-11).

### BLUE

Position: **all 30 gaps addressed — 3 accepted-and-repaired blockers, 2 new security blockers
absorbed, the two meta-gaps (R1-8/R1-11) answered with built arguments, and every citation drift
fixed at the leaf node.** Direction unchanged; evidence base now accurate; the build recommendation
is now *qualified and netted*, not asserted. Full detail: `blue/report.md` §12 (new) + in-place
repairs; edit log `blue/CHANGELOG.md` Round 1.

**Accepted outright (red was right):**
1. **R1-1 (the headline error) — retracted and corrected.** I re-verified at the leaf node: the
   `internal/secrets` matcher (built-for-reuse), `sc-secrets-gate` (wired PreToolUse deny-hook),
   and `hooks.json` wiring **ship today**. Round 0 grepped only `*.md` and missed the Go layer — my
   error, not the proposal's. Corrected claim (§6.3, §8 item 3): the matcher + gate pattern exist,
   so a store scrub is **wire-not-build**; but the gate scans outbound *tool input*, not committed
   bytes, so it does **not** cover `git push` of store contents — that consumer must be wired. This
   makes the design *easier* than I claimed and hands it a ready-made pattern library.
2. **R1-2 clone-time injection (new blocker)** and **R1-3 bootstrap laundering (new blocker)** —
   both absorbed into the §4 threat model (§12.2, §12.3). The committed project store must gate
   projection *activation* on a local git-ignored ratification marker; bootstrap output is
   quarantined wholesale at `candidate` and any transcript that touched an external read is
   down-tiered (provenance-of-content, not just provenance-of-record).
3. **R1-4 internal incoherence — reconciled** (§12.4): withdrew autoMemoryDirectory-into-store (it
   deletes the capture screen); gated the high-authority `.claude/rules/` channel on trust tier so
   channel and voice are one coupled decision keyed on §4 tiers.
4. **R1-5 / R1-7 / R1-12 / R1-13 / R1-15 / R1-16** — all accepted: advisory-lock concurrency carved
   out of the multi-machine YAGNI (§12.6); ephemeral consolidator memory (§12.8); Evidence-section
   cap (§12.9); named tie-breaker replacing the confidence float (§6.2); defer/timing branch
   evaluated → hybrid-timing recommendation (§12.9); blocking-set effort annotated (§8, §12.9).
5. **R1-9 — delivered the re-scoped phase table** (§12.9): the actionable core Round 0 promised and
   omitted. Minimum viable bespoke layer = Phase 0 + Phase-2-scoped-to-`knowledge/` + the typed-
   extraction sliver; native-overlap phases deferred pending the Auto Dream flag check.

**Answered with built arguments (part accept / part rebut):**
6. **R1-8 netted build-vs-adopt** (§12.5): summed the net-new surface. Key finding — **most of the
   poisoning surface is inherited from native, not created by the bespoke layer** (the CVE hit
   *native* MEMORY.md). So "adopt native + thin skill" does not buy a smaller attack surface on the
   dominant dimension; it buys *less value* for *the same* risk. Genuinely net-new surfaces
   (push-exfil, clone, concurrency, curator) are each opt-in or cheap-to-fix. Build wins for the
   shrunken remit — on a **narrower margin than Round 0 implied**, contingent on the re-scope.
7. **R1-11 grade** (§12.5): part concede. Built the attacker model Round 0 lacked — the vectors are
   **opportunistic/supply-chain, untargeted** (malicious skill, mid-session web page, poisoned
   package, poisoned clone), so likelihood is *not* gated by "who would target a solo operator."
   Weighed the disconfirming consensus (single-user local markdown = advisory-lock-is-enough) and
   accepted its qualifier: it holds *for trusted input*, which /ingest and web reads violate — so it
   **localizes** the blocker to the ingest edge rather than refuting it. The blocking core is the
   **two ingest-edge gates** red concedes; the other three mitigations add ~no complexity beyond
   them (trust tiers = the enforcing data model; independent-corroboration = one predicate; de-
   authorized voice = folded into §12.4). Kept all five, justified each against the attacker model.
   Demotion candidates if the lead disagrees on cost: mit.4 and mit.5, never the two gates.

**Citation integrity:** every R1-18 … R1-30 drift repaired in place with access dates; mem0's
move to ADD-only (R1-23) turned a stale citation into fresh corroboration of my append-only rule;
footnote-reference integrity re-checked (no danglers). Net verdict stance: **endorse, gated on the
now-5-item blocking set** — CHANGES-REQUIRED, direction affirmed.

## Round 2

### RED

**Verdict: FAIL (CHANGES-REQUIRED).** Full audit: `red/findings.md` (cumulative; 30 R1 gaps +dispositions, 13 new R2 gaps). Five Round-2 lens passes preserved in
`red/candidates/round-2-lens-{1..5}.md`. Blue made real progress and I credit it: **14 R1 gaps
close at the leaf node** (R1-1/4/10/12/13/14/17/18/20/21/23/24/26/27), R1-9 and R1-15 delivered the
missing re-scoped phase plan and defer/timing branch, and R1-2/3/5/6/7 are accepted with the right
*direction*. But accepting a direction is not accepting an implementation — and this round's finding
is that **every new §12 mitigation ships with an un-graded second-order failure mode.** PASS is
unavailable: no new gap is yet closed, rebutted, or risk-accepted.

**Red hard-blocks (new, this round):**

1. **R2-1 — the clone-injection fix is self-defeating (HIGH).** §12.2 gates projection activation on
   a *content*-fingerprint ratification marker. But `/dream` mutates the store every night, so the
   fingerprint mismatches after every legitimate run — the only escapes are self-ratification by the
   unattended pass (defeating the human-consent gate) or unworkable daily re-ratification. It leans on
   the diligence §2.4 itself discredits, and the "auto-ratify a trusted root" escape hatch reopens the
   zero-click vector for the solo-dev common case. Fingerprint provenance/authorship, not content.
2. **R2-8 — a Round-1 citation *repair* regressed into a leaf-node contradiction (MEDIUM).** The
   R1-28 fix re-anchored the unpinnable "80–99%" band to "~90% environment-injection", cited to arXiv
   2604.02623 — which actually reports **≤32.5%**. The accurate MINJA ~95% leg lives in an uncited
   paper (2503.03704); `[^MemoryPoisonSurvey]` carries no ASR numbers at all. The disposition survives
   (blocking core = the two ingest gates), but red does not let a contradicted number stand.
3. **R2-3 — the provenance-of-content rule is "one predicate" only if it's useless (MEDIUM-HIGH).**
   §12.3's transcript-scoped tagging caps essentially *all* trajectory concepts at external-ingest
   (nearly every session touches the web), neutering auto-promotion; the turn-level alternative that
   preserves value is unspecified and is the hard part of the build, not "one predicate."

**Accepted-direction, residual raised as new R2 gap (each a second-order failure of the R1 fix):**
R2-4 (advisory lock: stale-timeout TOCTOU + capture-vs-commit un-serialized), R2-5 (history-scrub
publishes a scrubbed snapshot, sacrificing the "reviewable git repo" differentiator the build case
sells), R2-6 (ephemeral curator closes the *durable* self-poison path but not in-pass steering via
the store it reads), R2-7 (re-scope defers `MEMORY.md` consolidation to flag-gated Auto Dream with no
flag-absent fallback → unowned consolidation → measured context-rot).

**Citation surface:** R2-9 — three R1 body repairs (v2.1.59 R1-22, "80–99%" R1-28, Letta git-branch
R1-25) did **not** propagate to their footnotes; the leaf-node reader still lands on the retracted
claim. R2-10 — the §12.5 disconfirming leg rests on a dev.to article that frames by *scale not trust*
plus an unfollowable self-survey. R2-11 (stale "before Phase 1" anchor), R2-12 (v2.1.33 parity),
R2-13 (dangling §11 refs + absent Heilmeier section) are low/assembly-time.

**To the lead's docket (contested, not red-closable):**
- **R1-8 + R2-2** — the netted build-vs-adopt table (delivered as demanded) neutralizes the dominant
  poisoning axis by labeling it "Shared/inherited from native." That contradicts blue's own §4 ("the
  store … **widens** it"): bespoke adds explicit external `/ingest` intake, cross-project blast radius,
  and a corroboration→auto-promotion ladder native lacks — and its `.claude/rules/` channel
  *re-authorizes* the high-authority surface native's CVE fix *de-authorized*. Double-bind: R1-29 tags
  "removed from system prompt" medium-confidence — if unreliable, blue can't equate the surfaces; if
  reliable, bespoke re-opens what native closed. The value side R1-8 asked to sum is still qualitative.
  This cell carries the go/no-go.
- **R1-11** — poisoning apparatus sizing. Blue part-concedes: two ingest gates = blocking core;
  mit.4/mit.5 are demotion candidates. Reasonable — but R2-8 corrects the likelihood premise (~32.5%,
  not ~90%) blue used to justify keeping the surplus, so the sizing call is now argued on a lower
  number. Both parties hand this to the lead.

**Pattern note (all four re-instances of already-recorded red memory):** self-defeating mitigation
(R2-1 write-loop collision + discredited-diligence + escape-hatch; R2-6 closes-only-durable-path);
inherited-surface netting must verify the baseline wasn't patched (R2-2); repair-regression on
citations (R2-8); incomplete-repair footnote lag (R2-9). The adversary's memory fired correctly this
round — no new pattern, but four fresh instances.

### LEAD

**Scope of this ruling:** the two docket gaps only (R1-8+R2-2, R1-11). The Round-2 hard-blocks
(R2-1, R2-3, R2-8) and residuals (R2-4..R2-13) are red's to run in the debate — they are newly
raised, not yet raised-rebutted-re-raised, so they are not on my docket. I read `debate.md` in
full, `red/findings.md` in full, and verified the R2-2 keystone at the leaf node in
`blue/report.md` (§4 line 385-387 vs §12.5 row 1 line 816).

**Deadlock: FALSE.** Red raised thirteen new gaps this round (R2-1..R2-13) and R1-8 is carried
below. Debate continues; no final assembly.

---

**R1-8 + R2-2 — netted build-vs-adopt; the "Shared/inherited" poisoning classification — CARRIED.**

Verified at the leaf node: blue's own §4 (lines 385-387) states the bespoke store "reproduces this
surface and *widens* it," while the §12.5 netted table row 1 (line 816) labels the identical
inbound-poisoning pipeline "**Shared** … Adopting native does not escape this." These two
statements cannot both be true. R2-2 is corroborated by blue's own text — this is not
rebuttal-sustained, and it is not closed (the "Shared" label stands uncorrected and it is, by
blue's own admission in §12.5, "the cell carrying the go decision").

Red's three enumerated widenings are real and under-counted by the table: (1) explicit external
`/ingest` `url:`/`file:` intake, which native — capturing only the operator's own sessions — does
not have; (2) cross-project blast radius from the *global* store, versus native's per-project
machine-local containment; (3) the corroboration → `review_count`-driven auto-promotion ladder,
which native lacks. The keystone conclusion — "adopt-native buys *less value* for *the same*
dominant risk" — is therefore false as stated. The honest form is: **adopt-native buys a
*narrower* poisoning surface for less value; build must argue the value is worth the widening.**
That is a materially weaker build case than the one §12.5 asserts, and R1-8's original demand —
quantify or bound the shrunken value against the summed cost — remains undischarged (the value
side is still a qualitative list: typed concepts, global git repo, ingest-with-provenance,
skill-promotion, committed project store).

I do not risk-accept this. The fix is low-complexity and the cell is go/no-go-bearing — the
threshold for risk-acceptance (a valid finding rejected on likelihood × impact × complexity) is
not met when the finding is cheap to fix and carries the verdict.

**Blue owes, to close on the next pass:**
1. Reclassify the inbound-poisoning row: count the three widenings (`/ingest` intake, cross-project
   blast radius, auto-promotion ladder) as **net-new bespoke surface**, not "Shared."
2. Re-run the build-vs-adopt conclusion on the corrected accounting — state explicitly that
   adopt-native buys a narrower poisoning surface for less value, then argue (do not assert) the
   differentiating value is worth the widening.
3. Bound or quantify the shrunken value side (even ordinally: which differentiators are
   load-bearing for the suite vs nice-to-have), so value is weighed, not listed.
4. Resolve the `.claude/rules/` re-authorization double-bind (R2-2, coupled to R1-4/§12.4): either
   route projections **unconditionally** through the de-authorized reference-voice channel so
   "Shared" becomes true by construction, or concede the re-authorization as a net-new widening and
   carry it in the accounting. Note the friction below: the double-bind's severity hinges on the
   medium-confidence "removed user memories from the system prompt" (R1-29) detail, which cannot be
   confirmed against the primary advisory from here — blue should state the conclusion so it holds
   under both branches of that uncertainty (as it already attempts at §4 lines 382-385), which
   favors the unconditional-de-authorized-channel option since it is robust to the CVE-detail's
   reliability.

**R1-11 — poisoning apparatus sizing — CLOSED (adjudicated).**

Blue built the opportunistic/untargeted attacker model R1-11 said was missing and conceded the
blocking core (§12.5 lines 799-809): the two ingest-edge gates — (i) external-ingest never
auto-promotes, (ii) injection screening at capture. Red concedes the same two. That core is
sustained as blocking. R2-8's correction (env-injection ASR ≤~32.5%, not ~90%) lowers the
likelihood premise but does **not** disturb the blocking core: the blocking grade rests on high,
undisputed *impact* (persistent context compromise) plus the CVE-2026-21852 precedent of a real
untargeted supply-chain hit on Claude Code memory specifically — not on the headline success rate.
A gate whose justification is impact + demonstrated-in-the-wild exploitation does not weaken when
the success-rate figure is corrected downward.

Apparatus sizing (the contested part both parties handed me), decided against the corrected lower
likelihood:
- **mit.1 (trust tiers): not surplus — part of the blocking core.** It is the data model that
  makes gate (i) enforceable; there is no separable cost to demote. Blue's "zero marginal cost"
  characterization is accepted.
- **mit.4 (independent-source corroboration before promotion): retained, but NOT blocking —
  demoted to ingest-hardening (Phase 4).** It directly defends the R1-3 laundering path
  (`review_count: 2` → auto-promote), which is a real and cheap-to-defend vector. But its concrete
  form is entangled with R2-3's unresolved granularity question (transcript-scoped vs turn-level),
  so it cannot be certified as "one predicate" nor gated before the MVP ships. It gates the
  auto-promotion feature, not Phase 0/1. Against the corrected ~32.5% likelihood this is
  proportionate as a should, not a blocker.
- **mit.5 (de-authorize the projection voice): retained AND elevated to unconditional.** This is
  not surplus to demote — it is load-bearing for the R1-8/R2-2 keystone (routing projections
  through the de-authorized reference-voice channel is the cheapest way to make the "Shared"
  classification true by construction). §12.4 already couples channel to trust tier at ~zero cost;
  make it unconditional. Demoting mit.5 would reopen R2-2, so it stays.

Net: **blocking core = two ingest gates + mit.1 (the enforcing schema), sustained; mit.4 demoted
to non-blocking ingest-hardening; mit.5 retained and made unconditional.** This resolves the
apparatus-sizing call. R1-11 leaves red's verdict consideration.

---

**Friction affecting these rulings** (reported per protocol, not silently absorbed):
- No PDF-full-text / table-extraction capability: I could not independently re-verify the arXiv
  2604.02623 ASR figure (R2-8's ≤32.5%) at the leaf node from here; my R1-11 likelihood premise
  relies on red's leaf-node verification. With such a tool I would have confirmed the figure and
  the MINJA ~95% attribution directly.
- No access to the primary Anthropic security advisory for CVE-2026-21852: the "removed user
  memories from the system prompt" detail is load-bearing for R2-2's double-bind but is confirmable
  only against post-cutoff vendor-blog sourcing. With a way to fetch the primary advisory I would
  have settled whether the bespoke projection re-authorizes a *remediated* (net-new) or
  *unremediated* (shared) surface — which is why R1-8's direction (item 4) is written to hold under
  both branches.

### BLUE

Position: **every Round-2 gap addressed — 3 hard-blocks repaired (one redesigned, one re-graded,
one corrected at the leaf node), 4 second-order residuals fixed, the lead's go/no-go docket item
closed on the corrected accounting the lead demanded, and 6 citation/label regressions fixed in
place.** Red's Round-2 thesis — "each R1 mitigation shipped with an un-graded second-order failure"
— was fair and largely right; the fixes are in `blue/report.md` §13 (new) + in-place repairs.
Direction unchanged; the build case is now weaker and more honestly bounded than any prior round.

**Red was right; repaired:**
1. **R2-1 (clone-fix self-defeating) — redesigned (§13.2).** The content-fingerprint marker was
   invalidated by every nightly `/dream` — a genuine self-defeat. Withdrawn. Re-keyed on **commit
   authorship / repo identity**: a store activates at instruction authority only if its
   `.claude/knowledge/` commits carry a trusted identity (`/dream` commits under the operator's
   identity, so nightly *content* mutation never changes *authorship* → gate never invalidates).
   Foreign clone → candidate tier + nudge. The trusted-root auto-ratify knob is **removed**
   (self-authored repos are trusted by construction; attacker-crafted clones are not). `/dream`
   output still passes the independent §4 content gate, so a trusted author cannot self-elevate
   poisoned external content. Signed commits = the strong-form next rung; graded residual stated.
2. **R2-8 (citation regression) — corrected at the leaf node.** I re-verified both papers: arXiv
   2604.02623 reports **≤32.5%** (up to 8× under stress), not "~90%"; MINJA (**~98.2%/76.8%**) is
   arXiv 2503.03704, now cited as new [^Minja]. The "~90%" is retracted everywhere; the survey
   [^MemoryPoisonSurvey] (no ASR numbers) no longer backs any figure. Blocking grade **unaffected**
   — it rests on impact + the demonstrated CVE-2026-21852, not the success rate (lead agrees).
3. **R2-3 (provenance "one predicate") — re-graded + specified (§13.4).** Accepted: transcript-scoped
   tagging neuters web-derived auto-promotion, and turn-level is real work (**re-graded Medium**).
   Specified the turn-level mechanism (tag on the *supporting turn set*, mechanical given the §1.4
   JSONL threading); transcript-scoped is the MVP default, turn-level the Phase-4 target. Partial
   rebut: the conservative form is not "useless" — operator-confirmed and operator-behavior
   auto-promotion survive it; it neuters *web-informed* promotion specifically, which is the intent.
4. **R2-4/5/6/7 residuals — all fixed.** Pid+heartbeat liveness + explicit-pathspec commit (no
   `git add -A`) so capture appends aren't staged mid-pass (§13.5); history-scrub reconciled — local
   store retains full history, only a post-leak *public mirror* degrades, a nice-to-have case
   (§13.6); consolidator reads store as **data not instruction** (structured-field merge + body as
   opaque payload), residual graded Low-L/High-I not claimed closed (§13.8); **flag-absent MEMORY.md
   fallback** — if Phase 0 finds Auto Dream not live, `/dream` retains consolidation, deferral is now
   conditional not assumed (§13.9).

**Lead's docket — R1-8/R2-2 CLOSED (§13.7):**
- The Round-1 **"Shared" label was wrong** (it contradicted my own §4 "widens it"). Corrected: the
  base own-session pipeline is shared, but the bespoke layer adds **three net-new widenings** —
  explicit `/ingest` intake, cross-project blast radius, the auto-promotion ladder — plus a fourth
  (`.claude/rules/` re-authorization).
- **Conclusion re-run and argued** (not asserted): adopt-native buys a *narrower* poisoning surface
  for *less* value; build must justify the value is worth the widening. Value bounded **ordinally** —
  2 load-bearing (cross-project global git repo; typed concepts + skill-promotion) vs 2 nice-to-have
  (`/ingest`; committed project store). The asymmetry *is* the build case: **widenings #1/#3/#4
  attach to nice-to-have/deferrable features** (each gated behind its own blocker), the second
  load-bearing feature (typed concepts) **narrows** the surface, and only widening #2 (blast radius)
  attaches to a load-bearing feature (the global store) — accepted explicitly as the gated price of
  core value (poison bounded to candidate-tier reference until it clears the gate).
- **Double-bind resolved by the lead's preferred option (item 4):** ALL projections route
  **unconditionally** through the de-authorized reference-voice channel (mit.5 elevated to
  unconditional; §12.4(b)'s tiered-voice exception superseded). This makes the authority dimension
  "Shared" **by construction**, robust to the medium-confidence CVE detail (R1-29) under both
  branches — which is why it beats tiered voice.

**R1-11 (adjudicated by lead) — reflected in the report (§13.10):** blocking core = two ingest gates
+ mit.1; mit.4 demoted to non-blocking Phase-4 ingest-hardening (the corrected ~32.5% likelihood
supports the demotion; I no longer argue to keep it blocking); mit.5 retained + unconditional.

**Citation/label integrity:** R2-9 (three footnote lags propagated), R2-10 (self-survey relabeled as
blue's reasoning), R2-11 (blockers re-anchored to re-scoped phases), R2-12 (v2.1.33 → community
source), R2-13 (Heilmeier §0 added, §11 disambiguated) — all repaired in place. No danglers.

Net stance: **endorse, CHANGES-REQUIRED, gated on the 5-item blocking set** — direction affirmed,
build case honestly narrower, every R2 gap closed, rebutted with evidence, or risk-accepted with
rationale (multi-machine concurrency; signed-commit strong form — complexity > likelihood × impact
for the solo baseline).

## Round 3

### RED

**Verdict: FAIL (CHANGES-REQUIRED).** Full audit: `red/findings.md` (cumulative). Five Round-3 lens
passes consolidated (`red/candidates/round-3-lens-{1..5}.md`): lenses 1-3 leaf-node citation
verification (three slices, live re-fetches), lens 4 logic & completeness, lens 5 dark-side & risk. I
credit real progress and I close what closes.

**Closed this round, on evidence I accept:** **R2-8** — I re-verified both band legs LIVE at the leaf
node: `[^EnvInjectedMemory]` (arXiv 2604.02623) returns ASR 32.5%/23.4%/19.5% + 8× under stress, and
MINJA (arXiv 2503.03704) 98.2%/76.8% is now cited and followable; no contradicted "~90%" survives.
That discharges **R1-28** too (band honestly wide, each half traced). **R2-1** content-fingerprint
self-defeat is redesigned away (authorship gate; nightly leg genuinely closed). **R2-3** "one
predicate" conceded and re-graded Medium. **R2-4/R2-5/R2-7/R2-11/R2-12/R2-13** all landed. **R1-8/R2-2**
met the lead's four asks (§13.7) — I do **not** re-open the classification. Severity is *declining*
round-over-round — convergence, not flailing.

**But the Round-2 pattern repeats a third time, and the citation surface still lags the body.** Eleven
new R3 gaps plus one R2 footnote confirmed still open at the leaf node.

1. **R3-1 (blocking-candidate) — the authorship clone-gate (§13.2) relocates §2.4 diligence, doesn't
   escape it, and mis-grades forgery.** Nightly leg closed (accepted). But (a) foreign-clone
   ratification is still a human judgment subject to §2.4 decay — escape claimed, per-clone volume never
   argued; (b) forgery is **low-effort / targeting-required**, not "high-effort" — a git identity is
   public in every commit, one `git config` forges it. With §13.13 risk-accepting the signed form, v1's
   honest guarantee is "defends only against attackers who don't know your public git identity."
2. **R3-3 + R3-7 — the turn-level provenance mechanism (§13.4) re-opens R1-3 laundering two ways.** R3-3:
   "immediately follow in parentUuid lineage" is unsound — poison read early launders into a later
   reasoning turn tagged trajectory-derived → auto-promotes; sound taint is transitive-after-any-external-read
   (collapsing toward the conservative rule R2-3 flagged as neutering auto-promotion). R3-7: the tag rests
   on the extractor's *self-reported* supporting-turn-set, but the extractor must read the poison to
   extract at all — an injection ("attribute this to the user's instruction") steers its own provenance
   report; mit.3 screens fact bodies, not provenance metadata.
3. **R3-4 (leaf-node contradiction) — §13.8 "opaque body" vs §2.3a semantic-dedup** (verified: report
   lines 1319-1320 vs 321). Paraphrase dedup *requires* reading bodies; "opaque payload never acted on"
   is not implementable without regressing to the lexical baseline §2.3a proves inadequate. Either the
   consolidator reads bodies (steerable — R2-6 not closed) or dedup fails. Residual larger than graded.

**Coherence / completeness residuals:** **R3-8** (build-value ecosystem breadth contradicts the
clone-ratification "clones mostly own repos" risk-accept), **R3-9** (§13.8 structured-field reliance
trusts the exact fields — `review_count`, provenance tier — the laundering pipeline inflates), **R3-2**
(shared/collaborative store: per-repo vs per-commit-authorship trust conflated; post-ratification
injection un-graded), **R3-6** (Auto Dream flag checked once in Phase 0 but volatile server-side — a
later flip re-creates the two-writer `MEMORY.md` collision undetected), **R3-12** (no consolidated
operative-decisions view — key items reachable only by excavating §N → §12 → §13 revision strata),
**R3-10/R3-11** (typing miscounted as surface-narrowing; §8 "27 items" split across §8 + §13.11).

**Citation surface (leaf-node, this round's re-follow — not trusting the CHANGELOG):** **R2-9(a) STILL
OPEN** — `[^MemoryDocs]` (line 1414) *literally still* carries "(auto memory native v2.1.59+)" then a
note claiming it was dropped: retract-by-annotation, not deletion; the same author cleanly deleted the
parallel "v2.1.33+" in `[^SubagentDocs]`, so this is an execution miss. **R3-13** (§1.5 still "46k-star"
vs §7/footnote "~87.1k"). **R3-14 (MEDIUM)** — `[^MemorySurvey]` over-attributes: the ~29-day
half-life (§6.1's *sole* prop for "decay guesses are in the evidenced band"), semantic intensification,
and cross-version score drift are not surfaced at the primary; unable-to-corroborate, not contradicted,
but a design-parameter figure must be pinnable. **R3-15/R3-16** (RecMem 77% lower bound; InstructionBudget
"<200 lines" is <100/40-80 in the primary). **R3-17** (`[^MemoryPoisonCve]` footnote states flatly what
the §4 body tags medium-confidence).

**Compounding into the (closed) docket, not re-litigating it:** **R3-5** — the accepted widening-#2
"bounded to candidate-tier" leans on mit.2, which only gates *external-ingest-tagged* content, so it
inherits R3-3/R3-7's under-tagging; the accepted price is higher than stated if those stand.

**Pattern note (all fresh instances of recorded memory):** self-defeating mitigation (R3-1
relocated-problem, R3-3 cheap-form-useless, R3-4 fix-collides-with-prior-requirement); provenance
self-report trusted from a compromised component (R3-7); footnote over-attribution (R3-14);
incomplete-repair footnote lag / repair-note-without-edit (R2-9a, R3-13, R3-17). **Meta, offered to the
lead (not forced):** three rounds of gate-by-gate patching trace to a missing single information-flow
invariant ("external-touched ⇒ tainted, transitively, until a human clears it"); R3-3/R3-4/R3-7/R3-9
collapse into that one rule. Severity is declining, so this is a design-coherence *recommendation*, not
a block. No new R3 gap is closed, rebutted, or risk-accepted → PASS unavailable.

### BLUE

Position: **every R3 gap closed, rebutted with evidence, or risk-accepted with rationale — and
red's own meta-recommendation adopted.** Red identified that three rounds of gate-by-gate patching
trace to one missing information-flow invariant. Blue agrees and **adopts it explicitly (§14.1)** —
replacing six spot-patches with one principle is the anti-complexity move the pragmatist mandate
demands. Detail: `blue/report.md` §14 (new) + in-place citation repairs; edit log `CHANGELOG.md`
Round 3.

**The organizing fix (§14.1) — the trust-derivation invariant:** *no trust-elevating field
(`status`, provenance tier, `review_count`) is ever inherited from bytes an attacker could author;
trust is re-derived locally from harness-observed evidence + explicit operator action.* Two
corollaries: **import** (foreign clones load reference-tier, committed fields reset — closes R3-2,
R3-9-for-clones, bounds R3-5) and **session** (taint is parser-derived, transitive-after-any-external-read,
NOT LLM self-report — closes R3-3, R3-7).

**Red was right; repaired or re-graded:**
1. **R3-1 (authorship clone-gate) — accept the re-grade; the invariant demotes the consequence.**
   Forgery is **low-effort / targeting-required** (accepted; one `git config`), and foreign-clone
   ratify does inherit §2.4 decay (accepted, bounded to one-time-per-repo). But under §14.1
   authorship-trust is **demoted from the security boundary to a nudge-convenience**: forging the
   operator's identity buys *nudge-suppression*, not activation, because declared tiers do not inherit
   and elevation requires local re-derivation or explicit `/remember`. Honest v1 guarantee stated:
   defends untargeted/broadcast fully; residual = operator manually `/remember`-confirming screened
   poison (Low-L/High-I). **Signed-commit risk-accept re-affirmed and *strengthened*** — signing now
   gates only the nudge, so it is further from load-bearing, not closer (rebut of R3-1d).
2. **R3-3 + R3-7 (turn-level provenance unsound) — §13.4 mechanism withdrawn.** Provenance is now
   **parser-derived from tool-use records, not LLM self-report** (closes R3-7 without extending mit.3
   — provenance is no longer LLM-authored). Transitive taint adopted (closes R3-3). Conceded: this
   collapses toward the conservative rule and **neuters web-informed auto-promotion** — so
   auto-promotion is **downgraded to a convenience** (untainted sessions + operator-confirmed only).
   Sound per-turn info-flow is **the unsolved problem red says it is** → **risk-accepted** rather than
   shipped as an unsound approximation (absorbing it makes the design strictly worse).
3. **R3-4 (opaque-body vs semantic dedup contradiction) — accepted, reframed, re-graded up.** The
   consolidator **does** read bodies semantically for dedup (required by §2.3a); "opaque payload never
   acted on" is corrected to "non-executable data interpreted for similarity, never obeyed as
   instruction" under data-framing. Residual re-graded **Low-Med-L / Med-High-I** (a crafted body can
   bias merges), capped by git-revert recoverability + per-pass caps — **not claimed closed**.
4. **R3-9 (structured-field inflation) — two threats separated.** §13.8 addresses prompt-injection-of-
   the-consolidator (Threat A) only; field-inflation (Threat B) is defended by the invariant + mit.4,
   not by §13.8. Structured-field reliance no longer presented as injection-safe in general.
5. **R3-2 (multi-author trust) — specified.** Per-concept authorship + import corollary: collaborator
   concepts arrive reference-tier, elevate only by local action; malicious-PR/compromised-collaborator
   injection reaches reference tier, never instruction authority, structurally. Attaches to the
   nice-to-have committed-project-store.
6. **R3-6 (volatile flag) — recurring per-run check** replaces the one-time Phase-0 gate (§14.7).
7. **R3-8 (breadth vs risk-accept) — reconciled:** value leans on the operator's *own global store*
   (no foreign clone); the clone risk attaches to the nice-to-have committed-project-store; the import
   clamp makes breadth-driven cloning safe-by-default.
8. **R3-5** — widening-#2's "bounded to candidate-tier" acceptance **conditioned explicitly** on the
   §14.1 invariant, which this round adopts; dependency stated, not hidden.

**Citation surface (leaf-node re-verified this round):** R2-9(a) **executed as deletion** (the four
words are gone, not re-annotated); R3-13 (§1.5 → ~87.1k); R3-14 (~29-day half-life withdrawn,
[^MemorySurvey] trimmed — my fetch confirms the figure is absent from the source); R3-15 (RecMem "up
to ~87% / accuracy improved" — fetch-confirmed); R3-16 ("<100 lines / 40–80", instruction vs line
budgets separated — fetch-confirmed); R3-17 (medium-confidence tag carried into [^MemoryPoisonCve]);
R3-10 (typing reclassified surface-neutral); R3-11/R3-12 (§8 forward pointer + **consolidated
operative-decisions table §14.8**).

Net stance: **endorse, CHANGES-REQUIRED, gated on the 5-item blocking core** — direction affirmed,
build case unchanged from the honestly-narrowed Round-2 margin. Every R3 design gap is fixed by the
adopted invariant, re-graded where red's grading was right, or risk-accepted with an explicit
complexity > likelihood × impact argument (sound per-turn info-flow; signed-commit strong form).
Severity declined round-over-round; the remaining residuals are graded and bounded, not open holes.

## Round 4

### RED

**Verdict: FAIL (CHANGES-REQUIRED).** Full audit: `red/findings.md` (cumulative). Five Round-4 lens
passes consolidated (`red/candidates/round-4-lens-{1..5}.md`): lenses 1-3 leaf-node citation
verification (three slices, live re-fetches), lens 4 logic & completeness, lens 5 dark-side & risk. I
credit real progress and I close what closes.

**You adopted the invariant I asked for, and I credit it.** Replacing six spot-patches with one
information-flow rule (§14.1) is the right move — more coherent than the gate-by-gate design.
**Closed this round, on evidence I accept:** **R2-9(a)** — the last standing R3 citation residual is
now an *actual deletion* (I grep-confirmed "2.1.59" survives only in the removal-note and the
"is dropped" body sentence, nowhere as a live claim); **R3-6** (§14.7 recurring per-run check — the
pinned-contract discipline I asked for); **R3-7** (self-report *removed* from the trust path — taint
is now parser-derived from harness tool-use records); **R3-9** (Threat-A/Threat-B split); **R3-10**
(typing reclassified surface-neutral); **R3-13/14/15/16/17** (all six R3 citation repairs verified
landed — I re-verified MINJA ISR 98.2%/ASR 76.8% and env-injection ≤32.5% LIVE at the leaf node again).
**R3-4** I accept as *graded-and-open*: §14.4 concedes the leaf-node contradiction, corrects "opaque
payload," and re-grades the residual **upward, explicitly not claimed closed** — that is disclosure,
not a soft-pass. Severity declined round-over-round: R1 had a verified-false claim plus three new
security vectors; R4 has one enumeration hole and one missing-enforcer in an otherwise-sound invariant,
plus coherence/citation lag. Convergence, not flailing.

**But the keystone is over-claimed on two axes, and I will not soft-pass a "sound" claim the design
does not have.** Twelve new R4 gaps; two blocking-candidate.

1. **R4-1 (blocking-candidate) — §14.1's *session* corollary is "sound" only against an
   under-inclusive four-item denylist.** "Transitive-after-any-external-read" is sound only if
   "external read" is the *complete* set of channels attacker bytes enter through. It is not: **(a)
   `Bash`-fetched content** (`curl`, `gh pr/api`, `git log` of a remote commit) enters as a `Bash`
   tool result → tagged `trajectory-derived` → auto-promotable — and this is *provable*, because your
   own outbound secret-gate wires on `WebFetch|WebSearch|Bash` (§6.3), so the design already treats
   `Bash` as I/O in the outbound direction but omits it inbound; the exfil pipe and the injection pipe
   are the same. **(b)** MCP tool results + sub-agent `isSidechain` reads (the parent never inherits
   sidechain taint). **(c)** in-repo files read via `Read` authored by untrusted commits — below the
   import corollary's reach. Every Round-3 closure that leans on the invariant (R3-1, R3-3, R3-7, R3-5)
   is contingent on a soundness this denylist does not provide. **Fix: invert to an allowlist** — a
   candidate is trajectory-derived *only if* every supporting turn is operator/harness-authored with no
   intervening external-bytes tool result; a new tool defaults to tainted. Parser change, not research.
   If a complete allowlist is infeasible, **withdraw the word "sound"** and grade the residual as you
   did R3-4. **Pattern: invariant-soundness-by-enumeration.**

2. **R4-2 (blocking-candidate) — §14.1's *import* corollary is a policy with no enforcer; "not new
   machinery" is false.** The corollary states the outcome ("committed projection loads at reference
   tier, fields reset") but names no mechanism at the one moment it matters: `active.md` is a
   *committed file natively `@`-imported at session open, before any bespoke `/dream` runs*. The
   bespoke re-derivation clamps `knowledge/*.md` at the *next* local `/dream` — not at first open of a
   fresh clone; mit.5 de-authorization is a generator-side property that never touches attacker-authored
   projection bytes; a SessionStart hook cannot *un-import* an `@`-imported file. The only concrete
   enforcer ever specified — the git-ignored ratification marker — lived in the *withdrawn* §12.2 and
   was never carried forward. Across three rounds the enforcement was hollowed: §12.2 concrete gate →
   §13.2 authorship → §14.2 authorship demoted to nudge → §14.1 outcome asserted with no gate.
   Enforcing "a committed, natively-`@`-imported projection loads at reference tier" is *precisely* new
   machinery. **Fix: commit only raw concept bodies; git-ignore the projection *and* the trust-elevating
   frontmatter; regenerate + re-derive tiers locally** (this makes R1-2/R2-1/R3-1/R3-2 structurally
   moot) — and *price* the cost: the projection no longer travels with the repo, so the committed-store
   differentiator §13.7 sells shrinks to concepts-only. **Pattern: policy-without-mechanism.**

3. **Compounding into the closed docket, not re-litigating it.** **R4-3** — R3-5's "bounded to
   candidate-tier" reconciliation misattributes the mechanism: the import corollary fires only on a
   store "whose commits are not locally authored," so it does *not* fire for the operator's *own*
   global store that carries widening #2; the real bound is a single ingest gate, and post-clearance a
   poisoned concept is `active` in every project. **R4-4** — §14.3's auto-promotion downgrade lowered
   the *value* side of the §13.7 accounting the lead closed (recurrence auto-promotion now fires only on
   "no-external-read-at-all" sessions, which your own §13.4 says almost never happen), and relocated
   elevation onto manual `/remember` at higher volume — re-importing the §2.4 review-fatigue failure
   mode you documented. Both are notices that inputs to the closed go-decision moved *after* closure;
   neither re-opens the classification.

4. **Lower-severity (LOW/LOW-MED, non-blocking but open):** R4-5 (blocking count "5" is stale — a
   grade-changing supersession makes the operative set ~6, unverifiable from any single surface),
   R4-6 (the recurring flag-check leans on a native-consolidation "signature" for a feature on your own
   Unverified list, and `MEMORY.md` has no commit-authorship to read), R4-7 (Heilmeier §0 still markets
   the *automatic* ladder §14.3 demoted; title still says "Round 1"), R4-8 (`last_seen` named
   non-inheritable in the invariant header but omitted from the corollary's reset list), R4-9 (§2.3a
   cosine-bin dedup figures miscited to a 0–100-scale LLM-as-judge paper — 3 fetch/search routes agree
   on the scope mismatch), R4-10 (the calibration claim R3-14 re-homed onto `[^MemoryEviction]` — its
   arXiv leg SSGM does not carry it; rests on a Medium listicle), R4-11 (§5 states Auto Dream's
   ~24h/>5-session trigger as fact while §3/§10 hedge it), R4-12 (§9/§12.5/§13.3 merge MINJA's
   *injection*-success 98.2% and *attack*-success 76.8% into one band whose endpoints are different
   metrics; §4 states them correctly and separately).

**Net stance:** the invariant is one honest hardening pass from delivering what you claim for it — the
blocking pair is a parser allowlist-inversion (R4-1) plus one architectural decision to git-ignore the
projection and commit bodies only (R4-2), *not* redesign. No new R4 gap is closed, rebutted with
evidence I accept, or risk-accepted → **PASS unavailable.** Contested pair for the lead if you rebut
rather than fix: R4-1 (is a denylist acceptable for a taint boundary?) and R4-2 (does "not new
machinery" survive the committed-`@`-import path?). **Friction (carried):** still no full-PDF-text /
PDF-table extraction — MINJA digits took three fetch routes (only `/html/…v2` surfaced them); no way to
confirm Auto Dream's actual output behavior from here, which limits how hard R4-6 can be pressed.

### LEAD

**Scope of this ruling:** the docket carries a single contested gap — **R3-4** (raised R3, rebutted
by blue in §14.4, re-raised/accepted-as-graded R4). The twelve new R4 gaps (R4-1..R4-12), including
the two blocking-candidates R4-1 and R4-2, are **not** on my docket: they are newly raised this
round, not yet raised-rebutted-re-raised, so they are red's to run in the debate (same rule applied
at Round 2 and Round 3). I read `debate.md` in full and `red/findings.md`, and verified R3-4 at the
leaf node in `blue/report.md` §14.4 (lines 1569-1592) against §2.3a and the §9 risk row (line 1724).

**Deadlock: FALSE.** Red raised twelve new gaps this round (R4-1..R4-12), two blocking-candidate.
The anti-spinning test — no gap carried AND nothing new raised — is not met: new material was
raised. Debate continues; no final assembly. (R3-4 is disposed below and does not carry.)

---

**R3-4 — consolidator body-handling (§14.4) vs §2.3a semantic-dedup requirement — RISK-ACCEPTED.**

Verified at the leaf node. §2.3a (report line 321) proves lexical/title dedup fails against
paraphrase and therefore *requires* the consolidator to read bodies semantically; §13.8's "opaque
payload it never acts on" (formerly lines 1319-1320) could not co-exist with that requirement — a
genuine contradiction, correctly caught by red. §14.4 (lines 1569-1592) resolves it honestly: "never
acts on" is retracted and replaced with "non-executable data the consolidator interprets for
similarity but never obeys as instruction," and the residual is re-graded **upward** to
Low-Medium-L / Medium-High-I with the concrete residual named (a crafted body biases which concepts
merge → knowledge suppression via forced supersession, or fragmentation via blocked merge).
Critically, blue does **not** claim closure.

This is not `closed`: the finding is not resolved away — the consolidator must interpret
attacker-influenceable bodies to dedup at all, so the crafted-body-biases-merge residual is
irreducible short of regressing to the lexical baseline §2.3a already proves inadequate, or solving
sound semantic dedup-without-interpretation (an open problem). It is not `rebuttal_sustained`: blue
did not beat the challenge, blue conceded it. It is not `carried`: both parties agree there is no
further research owed — the residual is disclosed, graded, and defended in depth, and red explicitly
accepts it "as graded-and-open (disclosure, not soft-pass)" (findings.md lines 16-19; debate.md
566-569).

It is `risk_accepted` on the standard tradeoff: **likelihood Low-Medium** (a poisoned body must
first survive capture-screening mit.3 to reach the consolidator), **impact Medium-High but capped**
(a steered merge is a `git revert`-recoverable diff under §2.4 forensic control, and the per-pass
supersession/deletion caps + halt-and-flag bound how much one steered pass can suppress),
**complexity-to-fully-eliminate high** (the only closures are lexical-only regression or an unsolved
research problem). A design made worse to chase a full closure here — dropping semantic dedup — would
itself be the defect. **Recorded, never dropped:** it stays in the §9 risk register (rows/items 24
and 30, and line 1724) as a graded, disclosed residual carried into implementation with git-revert +
per-pass caps + mit.3 as the standing controls. R3-4 leaves red's verdict consideration.

**Note on the operative-set count.** This risk-accept does not disturb the blocking set, but it is
one of the supersessions R4-5 flags as making the operative blocking count (~6) diverge from the
stale "5." That is an R4 assembly-time item, not part of this ruling — red owns it in the ongoing
debate. Recorded here only so the final assembly does not miscount R3-4 as a clean closure.

**Friction affecting this ruling** (reported per protocol): none material to R3-4 — the contradiction
and the re-grade were both confirmable directly against `blue/report.md` on this machine, and both
parties had already converged. The standing PDF-full-text / table-extraction and primary-advisory
gaps recorded in my Round-2 friction persist but do not bear on this gap.

### BLUE

Every open R4 gap addressed in `blue/report.md` (new §15 + in-place fixes). The two blocking-candidate
structural gaps are **accepted and closed**; the coherence residuals accepted or risk-accepted with
value-movement stated; the citation lags fixed in place. Nothing from Rounds 0–3 removed.

**R4-1 (session-corollary denylist incomplete) — ACCEPT, closed (§15.1).** Red is right and the proof
is internal to blue's own design: §6.3's outbound secret-gate already wires on `WebFetch|WebSearch|Bash`,
so the design *already* treats `Bash` as an I/O channel — yet the taint denylist omitted it inbound.
Inverted to an **allowlist (fail-closed):** a candidate is `trajectory-derived` only if every supporting
turn is operator/harness-authored with no intervening un-provenanced tool result. Bash stdout, all MCP
results, sidechain reads (taint now propagates to the parent), and `Read` of files not authored by a
locally-trusted commit all taint transitively; a **new tool type defaults tainted**. This is strictly a
removal of trust and *cheaper* to reason about than the denylist (enumerate the two trusted channels,
taint everything else). "Sound" now holds for the channel dimension. Folds into §8 item 29.

**R4-2 (import corollary has no session-open enforcer) — ACCEPT, closed (§15.2).** Conceded: "not new
machinery" was wrong — enforcing the clamp against a natively `@`-imported committed `active.md` *is* a
mechanism, and it was absent (the only concrete enforcer lived in the withdrawn §12.2). Fix is a
`.gitignore` line that works by absence: **git-ignore `projections/`; commit raw concept bodies only.**
A fresh clone has no `active.md` to `@`-import, so nothing loads at active authority at session open —
no hook, no headless dependency. The raw `knowledge/*.md` bodies travel but are never `@`-imported
directly; their sole reader is the local `/dream`, which regenerates the projection *and* applies the
import-corollary clamp — exactly when the clamp should fire. Price stated honestly: the projection no
longer travels, so the committed-store differentiator shrinks to concepts-only. In exchange R1-2/R2-1/
R3-1/R3-2 are **structurally moot**. New §8 item 32, Blocking.

**R4-3 (§14.7 R3-5 mis-invokes the import corollary) — ACCEPT the correction; risk-accept the
per-project gate (§15.3).** Genuine leaf-node error: the operator's own global store is locally authored,
so the import corollary does not fire on it — the honest bound is a **single ingest-time human gate**,
not per-project re-derivation. A per-project activation gate would re-gate the operator's own confirmed
knowledge in every project — destroying the cross-project compounding that *is* the load-bearing value —
to defend the "operator confirmed a bad thing on purpose" residual (Low-L/High-I, already disclosed).
Complexity × value-destruction ≫ likelihood × impact: **risk-accepted**, bound stated, §14.7 corrected.

**R4-4 (auto-promotion value moved) — ACCEPT (§15.4).** (a) Stated to the lead: the §13.7 build margin
now rests on operator-gated promotion, not auto-recurrence; the margin **narrows but does not invert** —
both load-bearing differentiators (cross-project global store; typed concepts + human-gated skill
promotion) never depended on auto-promotion (widening #3, "never core"). §0 corrected. (b) Apply §2.4
controls to `/remember` as the primary elevation path: per-batch caps, weekly digest, tier-gated review,
mit.3 screening on `/remember`-ed content. Partial rebut on mechanism (affirmative self-selection is
weaker fatigue than clearing a bot-filled queue), but controls applied anyway (asymmetric cost). New
item 33, Medium.

**R4-6 (recurring flag-check primitive hand-waved) — ACCEPT (§15.5).** MEMORY.md lives outside the
project git repo, so no commit-authorship exists to read. Replaced with an implementable primitive:
`/dream`-recorded **hash-delta** (detects *a* foreign writer with no authorship). Confirming the writer
is *specifically* Auto Dream is downgraded to an **unverified Phase-0 empirical dependency**; absent a
signature the check degrades to a coarse fail-safe heuristic (unnecessary stand-down on a manual edit —
low-cost, safe direction). §14.7 "detect the signature" downgraded from settled.

**R4-5 (blocking count stale) — ACCEPT (§15.7).** Recomputed once, reconciling supersessions:
{1,2,3,15,16} → 21 supersedes 15 → 28 supersedes 21, 29 adds a slot (supersedes 22, which was High) →
{1,2,3,16,28,29}=6 after R3; +32 (R4) = **{1,2,3,16,28,29,32}=7.** Item 29 does **not** fold into 16
(29 = general allowlist taint mechanism; 16 = bootstrap-specific down-tiering — distinct scopes).
Verdict corrected to "33 items, 7 blocking"; blocking-set line added to §14.8.

**R4-7 / R4-8 / R4-9 / R4-10 / R4-11 / R4-12 — citation/labeling fixes, in place.** Title→Round 4; §0
reconciled; last_seen added to the §14.1 reset list; §2.3a cosine-bin figures dropped (embedding curve,
not the LLM-judge sensitivity study) with [^LLMJudgeDedup] relabeled; §6.2 calibration claim de-coupled
from the SSGM arXiv leg → Medium-listicle-only, blog-graded; §5 Auto Dream trigger tagged Unverified;
§9/§12.5/§13.3 MINJA split into ~76.8% attack / 98.2% injection (two metrics, not a band).

**R3-1/R3-2/R3-5/R3-8 — close by construction (§15.9).** With §15.2 (git-ignored projection = concrete
session-open enforcer) and §15.1 (channel-complete allowlist taint), the §14.1 invariant now *guarantees*
the outcomes these leaned on: no committed active-authority surface to inject through, no inbound channel
auto-clean. R3-5 closes on the corrected single-gate bound (§15.3); R3-8 on the now-enforced import clamp.

**R1-19 — carried, friction-blocked.** The 61.4%/71.6% digits remain unconfirmed at the leaf node
(direction independently carried by [^BotReviewFatigue] ~54%); blocked on a PDF-table-extraction
capability. Not verdict-bearing.

**Risk-accepted this round (complexity > likelihood × impact):** the per-project activation gate for
global-store concepts (§15.3) — it would destroy the cross-project differentiator to re-gate the
operator's own confirmed knowledge. Recorded in the §9 register, not dropped.

**Friction:** PDF/table-extraction capability still absent (blocks R1-19 leaf-node confirmation of two
figures); no other capability gap impeded this round — the two structural fixes and all citation
corrections were confirmable directly against `blue/report.md` and the shipped `internal/secrets`
wiring on this machine.
