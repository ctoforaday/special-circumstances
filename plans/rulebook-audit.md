# Rulebook audit — where our own conventions cost us

Operator-directed, 2026-07-18: scan the human-imposed rules for ones that do not
earn their keep. Evidence base: a full sweep of `research/*/friction.md` and the
E0.5 post-capture findings across runs 2-5 plus both smokes, classifying every
entry by root cause — (A) harness/tool limitation, (B) protocol-rule friction,
(C) genuine research difficulty. Only category B is in scope here; A is the
larger set and belongs to the efficiency/platform queue.

Nothing in this file is shipped. Each entry states the change and its status.

## THE META-FINDING (the reason this file exists)

Every protocol patch that shipped recurred in adjacent form:

| patched | recurred as | next run |
|---|---|---|
| lens-scoped finding labels (#15/#16) | blue-lane footnote namespace collision | run 4 |
| blue reads the transcript (#15/#16) | gap-JSON lossy summary, nine commands short | run 5 |
| grade enums widened (#15/#16) | mass mapping conflates existence with harm | run 5 |
| friction-to-file (#15/#16) | lens seats still had no write path | run 5 |

Four for four. We patch INSTANCES while the CLASS survives and re-emerges at the
adjacent seat — the exact argument in `ideas/gap-classes-proposal.md`, turned on
the rulebook itself rather than on debate gaps.

PROPOSED (status: OPEN, operator call not yet taken): extend the W2d class
registry to cover protocol-rule defects, so a rule patch must name its class and
sweep siblings before it ships. This is the only change in this file that stops
the recurrence pattern rather than adding a fifth patch to it.

## 1. The full-re-read mandate + additive-never-subtract (AGREED)

RULE: every audit seat re-reads the whole living report every round; blue never
subtracts substance.

EVIDENCE: 6 seat classes, every round, runs 3-5. The friction states the tradeoff
outright — "the full-read mandate outranks the token saving." The report grew
1178 -> 1668 lines in a single run (audit surface +75%); every audit seat pays
three windowed reads per round. The mandate has ALREADY grown two carve-outs
(findings sharding: red need not re-read its own closed cases; round-scoped audit,
held open as human-gated because it "trades against the full-re-read principle"),
and the cost recurred at every seat class after the sharding patch. The 25k Read
cap is harness (category A) and is NOT the complaint; the mandate is ours.

CHANGE: decouple the two rules that got conflated. "Additive" becomes a
CLAIMS-LEVEL invariant enforced by the record layer — no claim record may
disappear — rather than a PROSE-LEVEL one. Prose may then be compacted and
reorganized freely, because deletion is detectable structurally instead of by
forbidding rewriting. Re-read scope follows: a seat re-reads its own audit
surface, which is what the two existing carve-outs were already groping toward.

DEPENDS ON: the record layer (W2f / R2g) — the invariant needs claim records to
be enforceable. Until then the prose-level rule is the only thing preventing run
3's dominant failure class (blue quietly dropping content under pressure), so it
stays.

## 2. The disconfirming-search quota (AGREED, as amended)

RULE: spend at least one search in five on disconfirming evidence.

EVIDENCE — AND A RETRACTION: the first draft of this audit claimed the quota was
wasted volume. That claim is NOT supported and is withdrawn. What E0.5h actually
found is narrower and worse: the sections red NEVER gap-anchors are exactly the
self-critical material — both disconfirming passes, the human-gated pipeline, the
self-attested probe inventory. "Claims AGAINST the design attract no adversary."
The quota is fine; the material it produces is structurally exempt from audit.

CHANGE: bind the quota to the steelman-verification duty (already queued from
E0.5h) — producing a case-against creates an audit obligation on it. Red gains a
lens duty: is the case against honest and complete? The catechism audit found the
same shape at assembly (against-cases silently drop their strongest items).

CONVERGES WITH #3: declined and abandoned avenues ARE case-against material, so
the Lines of Inquiry record below is the surface this duty audits. One
instrument, not two.

## 3. Alternatives considered -> LINES OF INQUIRY (AGREED, expanded by operator)

RULE TODAY: `think-around-problem` mandates 3-5 genuinely distinct alternatives
before a significant decision; `terse-communication` forbids narrating options.
Required, invisible, unverifiable — a dead letter by the standard
`plans/scorecards.md` sets out ("a clause without an instrument and a feedback
path is a dead letter by construction"; cf. "confidence self-graded", mandated
and unpracticed at 5 markers in 1,892 lines).

CHANGE: exploration becomes a RECORD, and the record becomes a report section.
Operator's expansion (the better half of the idea): record not only the roads not
taken but the ones that WERE — pursued avenues act as counterpoints inside the
report, showing the exploration space rather than only its conclusions.

NAME: report section "## Lines of Inquiry"; record verb `avenue`. ("Avenue" alone
fails nominative determinism — an outsider cannot tell what it holds. "Lines of
inquiry" is standard research vocabulary and reads correctly cold.)

THREE STATUSES, each with a distinct audit question:

| status | meaning | audit question |
|---|---|---|
| `declined` | considered, not taken | was the rejection REASONED, or decoration? |
| `abandoned` | pursued, then died | what killed it? (the negative result) |
| `pursued` | became the report's spine | does the report reflect it honestly? |

`abandoned` is the most valuable and most routinely lost class: dead ends are
what a future run needs so it does not re-walk them, and they are direct
feedstock for the sleeper service's friction/subject mining.

INSTRUMENTATION (scorecard rows, blue chair):
- avenue count + method diversity per lane — DIAGNOSTIC, never a benchmark.
  Counting avenues invites avenue-padding; the Goodhart taxonomy exists for
  exactly this case.
- `declined` entries carrying no reason — DETECTOR (any nonzero is a finding).
- red's steelman duty (#2) reads the declined/abandoned set as its surface.

This is what finally instruments `think-around-problem`: the rule stops being
self-attested and starts leaving artifacts.

DEPENDS ON: blue's record CLI (W2f / R2g) — `avenue` is a new verb in blue's verb
set, and the report section is a projection like every other board.

## 4. semantic-consent's no-inference clause (AGREED)

RULE: "During ambiguity, YOU MUST NOT infer, guess, or decide unless the human
has explicitly made the decision yours."

EVIDENCE: no run-corpus evidence — this is a repo-development rule, not an engine
rule, so the friction sweep does not reach it. The case is structural: under a
long autonomous queue (the standing "land everything before the first run"
directive), a strict no-inference reading forces a stall at every small
ambiguity. Observed in practice this session: a round-2-vs-round-3 graduation
choice was decided, disclosed, and offered for reversal — which works, and which
the rule as written forbids.

CHANGE: split the clause by REVERSIBILITY. Decide-and-disclose for reversible,
in-scope calls; ask for irreversible, outward-facing, or scope-changing ones. The
rule should describe the behaviour that already works rather than prohibit it.
The consent boundary is unchanged where it matters — state-modifying and
destructive actions still require explicit semantic agreement.

## 5. Retry counts contradict (AGREED — SHIPPING NOW)

RULE: `plan-act-reflect` says "YOU MUST NOT retry more than twice without
changing the plan"; `anti-spinning` says "YOU MUST NOT attempt the same fix more
than 3 times" and escalate after the third. Two numbers for adjacent situations.

CHANGE: `anti-spinning` becomes the single authority on retry counts (3 strikes —
the incident-tested number); `plan-act-reflect` references it instead of stating
a rival number. No third rule invented.

STATUS: shipping as a prosthetic-conscience PR with a version bump.

## 6. Footnotes as a lived state (AGREED EARLIER — plans item, not queued)

RULE: semantic word-based footnotes, mandatory; `claim_count` is defined as
FOOTNOTED declarative claims; the citation ledger keys on footnote labels.

EVIDENCE: label collisions across blue lanes — `[^CaptureRecapture]` named two
different papers (Briand IEEE TSE vs Petersson JSS 2004), `[^BlueRound5]`
mislabeled a round-4 quote, manual reconciliation across ~50 definitions at
merge; footnote-block ownership unassigned at citation-slice dispatch (a lens
improvised a rule; patched in W2i); a grep miscount (64 vs 66) from markers and
definitions both matching. Note the class pattern: the lens-label half was
patched in #15/#16, the blue-lane namespace half was not and bit the next run.
Deeper cost: because the audit unit is the footnote, UNFOOTNOTED PROSE RECEIVES
ZERO ADVERSARIAL CONTACT (E0.5h: the frame "excludes unfootnoted mechanics
entirely"). A formatting convention is selecting what gets audited.

WHY NOT SIMPLY DROP THEM: the engine needs STABLE SOURCE IDENTITY — one source
cited eight times must be one ledger row, one verification, one drift trigger.
Inline links dedupe by URL string (fragile) and leave access dates and volatility
notes homeless. W2i's round-graduated sizing depends on the ledger working.

CHANGE: a citation becomes a RECORD ROW, not a markdown artifact; footnote
markdown is rendered by a projection pass. Collisions become impossible by
construction (ids minted by the tool, not chosen by lanes); slice ownership
becomes a query; linters only ever see rendered output; the audit surface becomes
"is a claim record" rather than "has a marker".

NOT TO BE SLIPPED IN: this changes `claim_count`, the citation ledger, and red's
PASS-bar evidence grade — the same class as W2.4, where the standing rule is
"semantic change -> debated, not slipped in." Compounding risk: the MASS
telemetry series just restarted at v2 (W2g); restarting the claim series in the
same wave leaves no comparable baseline on either axis at the first run.

MEASUREMENT FIRST: a footnote-mechanics friction row in the W2h scorecard, so the
next run says whether this is still bleeding post-mitigation. The friction
evidence above is all PRE-mitigation; the residual cost is unmeasured.

## Still open (operator call not taken)

- THE META-FINDING above: extend the class registry to protocol-rule defects.
- MEMORY-AS-READING IS MEASURED WORTHLESS: run 4 — the "read red's gap patterns"
  clause was unsatisfiable at four blue seats (no path provisioned by the dispatch
  that mandates the read). Run 5, worse — lanes verifiably READ the file and
  committed both warned patterns anyway; only red's DUTY-EMBEDDED patterns caught
  them. A rule saying "read this and be wiser" buys nothing; the same content
  compiled into a duty works. Bears directly on the untracked
  `.claude/agent-memory/` decision: the value of those 56 files is as compiled
  manifest lines (already specified in the constitutional reform), not as a
  corpus someone reads.
- TWO ENUM GAPS: no verdict class for ceiling-terminated runs (run 5 ended with a
  final revision no red pass ever audited, and no way to say so); no closure
  amendment class (late-discovered composition defects between two verified
  repairs masquerade as this-round closures).
