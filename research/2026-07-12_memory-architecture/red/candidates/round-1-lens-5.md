# Red audit — Round 1, Lens 5: dark-side & risk

**Surface audited:** full `blue/report.md` (Round 0) re-read in context, against
`inputs/memory-architecture-proposal.md` and leaf-node checks on this machine.
**Lens:** failure modes, likelihood × impact × complexity grading, security & tradeoff blindspots.
**Verdict:** **CHANGES-REQUESTED (not PASS).** Blue's report is strong and directionally correct;
its poisoning and consolidation-corruption work is real. But it carries one verified leaf-node
error that miscasts a blocking item, two new poisoning vectors it did not surface, two
internal contradictions in its own mitigation set, and one un-confronted build/don't-build
tradeoff. None is a soft-pass; each is listed below and stays open until closed, rebutted with
evidence, or adjudicated.

---

## R1 — VERIFIED ERROR: the secret-scrub gate partially EXISTS; blue's leaf-node check was scoped to `*.md` and missed the Go tool layer

**Location:** §6.3 ("Two false premises found by local verification"), sentence:
> "Proposal §9.5 / Phase 5 cite 'the port plan's **existing** secret-scrub (`git grep` denylist)'. **No such gate exists.**"
and §8 item 3: "Build the secret-scrub gate ... — **it does not exist to be reused**."

**Leaf-node counter-evidence (read on this machine, high confidence):**
- `plugins/prosthetic-conscience/tools/cmd/sc-secrets-gate/main.go` — a shipping PreToolUse
  Go hook: "BEFORE an outbound-capable tool call (WebFetch, WebSearch, Bash) runs, its input
  MUST NOT contain matchable secrets ... the call is DENIED." Has `main_test.go`.
- `plugins/prosthetic-conscience/tools/internal/secrets/secrets.go` — a shared, high-precision
  pattern package (AWS/GitHub/Slack/Anthropic/OpenAI keys, private-key blocks) explicitly
  built for reuse: "Every consumer (sc-secrets-gate, future telemetry redaction, **any
  scrubber**) imports this."
- `plugins/prosthetic-conscience/hooks/hooks.json` — wires `sc-secrets-gate` as a live
  `PreToolUse` hook on `WebFetch|WebSearch|Bash` today.

Blue's footnote `[^LocalRepoScrub]` states the check was `grep -i secret|scrub|denylist across
*.md`. That scope is structurally blind to compiled tooling; blue read the agent-guardrails
SKILL (future-tense "will enforce") and concluded nothing exists, when the deterministic gate
the SKILL describes had already shipped as a binary.

**The correct, narrower claim (which blue should adopt):** a reusable high-precision secret
matcher and a deny-gate pattern **already exist and are designed for reuse**. What does *not*
yet exist is a consumer covering the memory store's surface — **capture-time redaction of
writes into `short-term/`** and a **commit/push-time scan of store contents**. Critically, the
existing gate is *outbound-tool-input only*: it scans `tool_input` of WebFetch/WebSearch/Bash,
so it will **not** catch a secret sitting in a committed knowledge file being shipped by
`git push` (the push command line carries no secret). So blue's *conclusion* (a new
commit/capture-time gate is needed) survives, but its *premise and effort framing* ("build from
zero, it does not exist to be reused") are wrong and must be corrected — the work is "wire a new
consumer onto the existing `internal/secrets` package," not "build a scanner."

**Grade:** correctness / High confidence. Likelihood certain (verified) · Impact medium
(mis-scopes a blocking item; understates existing reuse; and the existing gate's outbound-only
scope is itself a latent gap the memory push path inherits) · Complexity to fix low.
**Disposition:** must be corrected before this item is closed.

---

## R2 — NEW (blocking-candidate): the project-store-committed-with-code is a wormable injection-distribution vector — the exact feature blue calls the defensible remit

**Location:** §3, sentence:
> "the bespoke layer's defensible remit shrinks to what native does not do: cross-project global
> knowledge as a reviewable git repo; ... **the project store committed with the code**."
and §7: "nothing surveyed offers project-store-committed-with-code."

**Dark side:** the project store lives in-repo and its `projections/active.md` is
`@`-imported by that project's `CLAUDE.md` (proposal §5). Therefore **cloning a malicious or
compromised repo and opening it with Claude Code auto-loads attacker-authored memory into
context — with no install step.** This is strictly worse than CVE-2026-21852's npm postinstall
vector blue cites in §4: that needed a package install; this needs only `git clone` + open. Every
supply-chain surface that ships code (dependencies vendored as submodules, template repos, forks)
becomes a memory-poisoning surface. Blue's §4 threat model addresses the *operator's own* ingest
and trajectory pipeline; it never addresses **repo-clone-as-injection**, and the very property it
markets as the surviving justification is the delivery mechanism.

**Grade:** Likelihood medium (any multi-repo / cloned-template workflow) · Impact high
(persistent context compromise, zero-click on clone) · Complexity medium (project-store
projections must be trust-tiered as `external-ingest` until the operator ratifies them; a cloned
project store must NOT auto-`@`-import at active authority). **Disposition:** surface as a
blocking-candidate for the §4 threat model; blue to rebut or absorb.

---

## R3 — NEW: `/memory-bootstrap` is the highest-concentration poisoning event, and blue's trust-tier taxonomy has a hole that lets it through

**Location:** blue §4 mitigation 2 ("External-ingest content never auto-promotes ... `/ingest`
output is quarantined at `candidate`") and risk table row "Memory poisoning via ingest/inbox
(§4) | Med".

**Dark side:** proposal §7.2 `/memory-bootstrap` "batch-fans-out `trajectory-review` over
**every transcript** under `~/.claude/projects/*/*.jsonl`" — i.e. the entire history of every
web page read, every `/ingest`ed doc, every pasted payload, processed **unattended in one pass**.
This is precisely where the reported 80–99% memory-poisoning success rates would compound. Blue's
gate keys on *provenance of record*: content tagged `external-ingest` is quarantined. But a
trajectory that *read a malicious web page mid-session* is classified `trajectory-derived`, not
`external-ingest` — its externally-sourced content is laundered into the higher-trust tier by the
bootstrap. **The taxonomy conflates provenance-of-record (which agent captured it) with
provenance-of-content (where the bytes came from).** The corroboration rule then makes it worse:
the same malicious page read across two historical sessions = `review_count: 2` = auto-promote.

**Grade:** Likelihood medium · Impact high (mass seeding of poisoned "corroborated" concepts) ·
Complexity medium (bootstrap must down-tier any trajectory whose transcript touched a `url:`/
external `file:` read; bootstrap output quarantined wholesale at `candidate`, never
auto-promoted). **Disposition:** open; blue's §4 must extend the tier taxonomy to
provenance-of-content and explicitly govern bootstrap.

---

## R4 — NEW: two of blue's own recommendations undercut its blocking poisoning mitigation

**(a) `autoMemoryDirectory`-into-store bypasses "injection screening at capture."**
**Location:** §3 ("consider pointing `autoMemoryDirectory` *into* the store's `short-term/`
(making native capture the ingest mechanism)") and §1.2 (same, "collapsing the ingest hop
entirely"), vs §4 mitigation 3 ("**Injection screening at capture** and at promotion").
If native Auto Memory writes *directly* into the git-tracked store, there is **no capture-time
hook to screen at** — Anthropic's writer, not ours, produces the file. Blue's cost-saving
collapse deletes the exact interception point its blocking mitigation requires.

**(b) `.claude/rules/` projection channel contradicts "de-authorize the projection voice."**
**Location:** §6.2 / §8 item 7 ("prefer generated, path-scoped `.claude/rules/` files over
`@`-import + SessionStart") vs §4 mitigation 5 ("**De-authorize the projection voice**:
projections render concepts as *reference knowledge*, not instruction-voiced text ... reduce the
authority of the surface").
`.claude/rules/` files "load at launch with **CLAUDE.md priority**" (blue §1.2) — they are the
*rules* channel; the name and load-order signal instruction intent. Post-CVE, Anthropic moved
authority **down** (removed memories from the system prompt); blue's rules-channel
recommendation moves it **up**. Blue treats channel choice (§6) and voice de-authorization (§4)
as independent free variables; they are coupled. Pick a coherent combination and state the
authority tradeoff, rather than recommending the highest-authority channel and low authority in
the same report.

**Grade:** Likelihood high (as written, both recommendations stand) · Impact medium (guts the
poisoning defense / internal incoherence) · Complexity low (reconcile). **Disposition:** open;
consistency defect in the recommendation set.

---

## R5 — NEW: concurrent single-box writers are un-graded; the "multi-machine" risk-accept mis-scopes the hazard

**Location:** risk table, row:
> "Multi-machine store divergence | Low (single operator, one box) | Low | Med (sync protocol) |
> **Risk-accept** — YAGNI; git remote is the sync story if ever needed."

**Dark side:** the risk-accept collapses "multiple machines" (genuinely YAGNI for one operator)
with "**multiple concurrent sessions on one box**" (routine — several terminals, worktrees, an
interactive session plus the scheduled nightly `/dream`). Blue's own §1.2 notes auto-memory is
"shared across worktrees." Concurrent commits to the *single* global store repo, plus an
unattended nightly `/dream` committing to the same repo, produce **git merge conflicts in an
automated repo with no lock, no merge driver, and no human present** — the nightly run either
fails silently (no-op night, matching blue's own "silent no-op nights" concern) or a naive
`git add -A && commit` races a concurrent writer. There is no concurrency-control story anywhere
in proposal or report.

**Grade:** Likelihood medium (any parallel-session workflow) · Impact medium (lost writes /
failed consolidation nights) · Complexity medium (advisory lock on the store, or Letta-style
isolated dream branch merged with a driver — blue already cites the isolated-branch pattern in
§5 but does not connect it to this hazard). **Disposition:** open; separate this from the
multi-machine YAGNI accept.

---

## R6 — NEW: "git history is the undo" contradicts the secret-history-scrub remediation on the same repo

**Location:** proposal §6 ("Git history retains it — nothing is truly gone") endorsed by blue
(§2.4 treats git-diff as forensic undo; risk table "OKF drift ... degrades to plain markdown"
logic leans on history). Against blue's own cited artifact: the git-proficiency CHEATSHEET
section "**Scrub a Folder from All History**" — the documented remediation when a secret is
committed to a *pushed* store.

**Dark side:** the two safety mechanisms are mutually exclusive on one repo. If a leaked
secret/PII in the pushed global store is ever remediated by history-scrub (filter-repo/BFG),
**every prior commit's hash changes and the "revert to yesterday's good knowledge" undo is
destroyed** for everything before the scrub. Blue's forensic-undo guard (§2.4) and its own
secret-leak posture (§6.3) cannot both hold. Neither proposal nor report reconciles them.

**Grade:** Likelihood low-medium (only when a scrub is triggered) · Impact medium (undo guarantee
silently voided exactly after a security incident, when you most want it) · Complexity medium
(separate the pushed/publishable store from the local forensic-history store, or accept that
push implies losing pre-scrub undo and say so). **Disposition:** open; surface the tension.

---

## R7 — NEW: the consolidator/curator agents are memory-backed and thus inside the poisonable surface (self-referential curation compromise)

**Location:** proposal §7 supporting agents: "**`memory-consolidator`** (`memory: project`, so
it learns the store's own shape over time) — the merge/dedup brain of the dream loop." Blue does
not flag this anywhere.

**Dark side:** the agent blue relies on as the curation/poisoning defense has its *own*
persistent memory *inside the store it curates*. A poisoned consolidator memory (e.g. a learned
heuristic "notes citing source X are reliable," or "prefer expanding concept Y") biases every
future merge/promote decision — a durable compromise of the mechanism, not just the data.
Combined with R4a (native writes bypassing screening), the loop can be steered rather than merely
fed. The defense is inside the attack surface.

**Grade:** Likelihood low (requires targeting the agent's memory file specifically) · Impact
high (systemic bias of all future consolidation) · Complexity medium (consolidator/curator run
with *read-only or ephemeral* memory during the pass; their learned memory, if any, is
operator-ratified, not self-written from trajectories). **Disposition:** surface-and-grade;
blue to argue whether the memory-backed consolidator is worth its self-poisoning surface.

---

## R8 — the build/don't-build tradeoff is asserted, never confronted, given native convergence

**Location:** §3 ("Bespoke remains justified for the shrunken remit; no external adoption
dominates") and Verdict ("the bespoke layer remains justified for a *shrunken* remit").

**Dark side / tradeoff blindspot:** blue grades ~13 risks individually but never **sums the
net-new attack surface against the shrunken value**. Native Auto Memory + flag-gated Auto Dream
(§3) now cover capture and consolidation for free. To recover the residual remit
(cross-project git-repo knowledge, typed concepts, external ingest, project-store-with-code) the
design **adds**: the inbound poisoning pipeline (§4), a `git push` exfil channel, a
concurrent-writer hazard (R5), a clone-time distribution vector (R2), and a self-poisonable
curator (R7). Blue asserts value > cost without quantifying either side. *Interesting is not the
same as of interest*: the honest adversarial question — does the shrunken remit justify the
*net-new security surface it introduces*, versus adopting native + a thin projection-only skill —
is never posed. This is not a demand to kill the design; it is a demand that the build decision
be **argued on the netted balance**, not assumed from per-risk mitigability.

**Grade:** strategic / meta. Likelihood n/a · Impact high (frames the entire go/no-go) ·
Complexity low (add a netted build-vs-adopt section). **Disposition:** open; blue or the judge
must confront the sum, not the parts.

---

## Evidence-confidence flags (stickler duty — "needs more evidence," not failure)

**R9 (medium confidence).** **Location:** §4, "CVE-2026-21852 ... fix (v2.1.50/v2.2) **removed
user memories from the system prompt**." This detail is load-bearing: it powers blue's claim that
`@`-import projections "still land ... with instruction-like authority (unlike post-fix auto
memory)," which in turn justifies mitigation §4.5 and colors the R4b channel choice. Sourced to
two vendor-blog-class posts (Cisco, omegamax.co), post-knowledge-cutoff, unverifiable from here.
Not a failure — tag the "removed from system prompt" mechanism as **medium-confidence** rather
than settled, since the differential-authority argument rests on it.

**R10 (high confidence, low grade).** **Location:** §7, "**claude-mem** (46k stars)." Live count
is ~85k and rising (Augment Code coverage shows 46.1k → 65.8k over the same period);
blue's figure is a stale point-in-time snapshot understating the plugin's prominence. Not
load-bearing (blue rejects claude-mem on binding constraints — SQLite opacity, no promotion
ladder — not popularity), so grade **low**; but a stickler notes the drift, and if anything it
strengthens the "strongest adopt-instead candidate" framing blue should own rather than round
down.

---

## Summary of dispositions

| # | Gap | Type | L×I×Complexity | Disposition |
|---|---|---|---|---|
| R1 | Secret-scrub gate partially EXISTS (`sc-secrets-gate` + `internal/secrets`); blue grepped only `*.md` | Verified error | certain × med × low | Correct claim; re-scope §8 item 3 |
| R2 | Project-store-with-code = zero-click clone-time injection vector | New security | med × high × med | Blocking-candidate; rebut/absorb |
| R3 | `/memory-bootstrap` mass-poisoning; trust taxonomy conflates provenance-of-record vs -of-content | New security | med × high × med | Open; extend §4 taxonomy |
| R4 | (a) `autoMemoryDirectory`-into-store bypasses capture screening; (b) `.claude/rules/` channel vs de-authorize-voice | Internal contradiction | high × med × low | Open; reconcile |
| R5 | Concurrent single-box writers un-graded; mis-scoped as multi-machine YAGNI | New failure mode | med × med × med | Open; separate from accept |
| R6 | "git history is undo" vs secret-history-scrub — mutually exclusive | New contradiction | low-med × med × med | Open; surface tension |
| R7 | Memory-backed consolidator is self-poisonable — defense inside attack surface | New security | low × high × med | Surface-and-grade |
| R8 | Build-vs-adopt netted tradeoff asserted, never argued | Strategic/meta | n/a × high × low | Open; must confront the sum |
| R9 | CVE-2026-21852 "removed from system prompt" — single-vendor sourcing, load-bearing | Evidence | medium confidence | Tag, don't launder |
| R10 | claude-mem "46k stars" stale (~85k live) | Evidence | high conf / low grade | Note the drift |

**Not-PASS.** R1 (verified error), R2/R3 (new blocking-candidate vectors), and R4 (mitigation
incoherence) each block a clean pass. R8 is the meta-gap the round should force into the open.
Blue's directional verdict and its consolidation/poisoning core stand — this is changes-requested,
not a redesign.
