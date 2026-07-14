# Red audit — round 5, lens: dark-side and risk

**Scope re-read this round:** the full living `blue/report.md` (910 lines) in context, not the
CHANGELOG diff; `blue/CHANGELOG.md`'s Round 0–4 entries (navigation only); `red/findings.md` in
full (1076 lines, rounds 1–4); `debate.md` in full (647+ lines, headers re-grepped: 5x `### BLUE`,
4x `### RED`, zero `### LEAD`, confirming red's round-4 finding still holds). Live re-verification
against `origin/main` (HEAD `42dba2d`, unchanged since round 4's pin — `debate.js` byte-identical,
confirmed by direct full read, lines 1–218): the args guard, `blueEnv`/`redEnv` null-guards, the
absent `supersedes` field, the pure-string-equality `contested` line, and the friction
aggregation's three call sites (`red-merge`, `judge`, `blue-respond`; none at `blue-synthesize`
until the docketed R3-2 fix ships) are all exactly as the report describes. No code drift to
report; `ideas/backlog.md` and `debate.js` both stable at the round-4 pin.

**Round-4 gap closure verified first-hand (not re-litigated as gaps, listed for completeness):**
R4-1 (§2.1 rows split into (a)/(b), §3 row 23, §2.3 addition 15, §5 item 12) — the fix is
correctly scoped as additive to the `prevGapIds`-widening remedy, matches red's required fix.
R4-2 (§3 row 20 decided to throw, §2.3 addition 13 extended with the positive assertion) — decided,
reasoned, consistent with the report's own anti-silent-degradation argument used elsewhere.
R4-3 (§3 row 6's originating sentence edited at the source; grepped report-wide for the retired
"critical-stance/adversarial-disconfirming lens" slash-compound — zero remaining instances). R4-4
(the fifth "4th occurrence" instance corrected to "third occurrence (corrected R2-1)"). R4-5 (all
four locations + §1.2 now carry the `MA-` prefix; [^GapIdScheme] present). All five close clean —
no regression detected in any of the five this round.

This lens's job past that point: find what a clean-citation, clean-propagation report still gets
wrong on dark-side/risk grounds — specifically, whether round 4's *own fixes* introduce failure
modes of the exact class this report spends four rounds warning about.

---

## R5-1 — OPEN — HIGH — medium (an LLM told to set an optional field under prompt instruction
alone, with no schema or script-side enforcement, is exactly the discipline this corpus's own
history shows lapsing — R3-2's dropped `blue-synthesize` friction call, R4-2's un-decided
disjunction shipped verbatim, R3-6's imprecise phrasing surviving two rounds) x high (silent,
undetectable reversion to the *exact* defect the fix exists to close — `contested.length` stays 0,
indistinguishable from "no regression occurred") x medium (the naive fix — make `supersedes`
required — does not close the gap; see below) — corroboration: HIGH (direct read of `debate.js`
schema at `42dba2d`; row 23's own fix text read verbatim; cross-checked against the report's own
adjacent argument about good-faith-dependence)

**Location:** §3, row 23 (*"(Added round 4, R4-1) Lineage-following contested-gap detection — the
docket detector is id-string-equal, not lineage-aware"*) — required-fix cell: *"(1) add
`supersedes: { type: 'array', items: { type: 'string' } }` (**optional**) to `RED_ENVELOPE`'s gap
schema and **instruct red-merge to set it** when closing a gap 'WITH REGRESSION' and minting a
successor."*

**Problem:** the fix for the lineage-blindness defect is itself lineage-blind to its own failure
mode. Two sentences earlier in the *same row's own location* (§2.1, quoted at the top of this
lens's re-read), the report makes exactly the point that should have been turned on the fix
itself: *"The convergence this report has praised throughout ... is a property of this run's
actors' good faith, not a property the detector enforces — a less scrupulous or more fatigued pair
of seats could spin the same lineage indefinitely."* The proposed remedy for that exact
good-faith-dependence is: an **optional** schema field, populated only because red-merge is
**instructed** to remember to set it when it judges a closure to be "WITH REGRESSION." Nothing
validates that a gap actually closed-with-regression this round has a successor gap whose
`supersedes` array names it. This is not hypothetical skepticism about LLM reliability in the
abstract — this exact corpus has already demonstrated the failure class twice: R3-2 found a
schema-declared, seat-identical `friction` field (`BLUE_ENVELOPE`) going uncalled at one call site
for three full rounds before anyone noticed; R4-2 found red's own required-fix text shipped
verbatim as an undecided disjunction because nobody was forced to resolve it. An **optional** field
set purely by prompt instruction, with the merge agent under active cognitive load consolidating
lens passes, closing prior gaps, and drafting new ones in the same turn, is a materially weaker
guarantee than either of those two precedents — and both of those precedents *failed silently*
until a much later round's citation-lens pass happened to notice. The founding-suite case that
would validate this fix (§2.3 addition 15, *"three canned `redEnv` round objects where round 1
raises gap `X-1`, round 2's merge closes `X-1` 'WITH REGRESSION' and raises a fresh-id successor
`X-2`..."*) only tests the **detector's** logic given a correctly-populated `supersedes` field in
the canned input — it cannot test whether the live merge agent reliably populates that field in
the first place, because that is a real-model-reasoning question the report's own boundary
discipline (§2.1, *"a simulator that fakes... judgment content is the research problem itself"*)
places outside what a zero-token unit test can cover. So the report currently ships a Tier-A test
for a fix whose actual failure surface is Tier C, with no Tier B/C fallback named and no residual
risk flagged.

**Why "make it required" doesn't close this (anticipating the obvious counter):** requiring
`supersedes` to be present whenever a gap is closed "WITH REGRESSION" only forces its *presence*,
not its *correctness* — an agent under load can satisfy a required-field constraint with a vacuous
or incomplete array (e.g., listing only one of two actually-superseded prior ids), which is exactly
the R3-1 `{verdict:'FAIL', gaps:[]}` class of failure this same report already names: schema-legal,
semantically empty. There is no cheap, complete fix available from inside the script alone, because
the script does not track per-round gap dispositions as structured data (dispositions live only as
prose in `red/findings.md`, which the script never reads). This raises the honest complexity grade
from the "one schema field + one prompt sentence" the report currently prices it at (its complexity
cell states row 23's whole three-part fix, including this piece, at Medium) to something the report
should say plainly: the schema/prompt half of the fix is necessary but insufficient, and the
residual reliance on red-merge's unforced compliance is either (a) an explicit, named risk-accept
(consistent with how row 19's poisoning mitigation was honestly rescoped in R2-7, and row 13's PDF
MCP adoption got an explicit vetting-step precondition in R1-14 rather than being priced as a free
fix), or (b) closed by a structural cross-check the script *can* perform without reading prose:
e.g., record every gap id closed with a "WITH REGRESSION"-flagged disposition in the judge/merge
envelope itself (a small, required, machine-checkable field on the *closed* gap, not the successor)
and have the script assert, after each merge, that every regression-flagged closure this round has
a corresponding `supersedes` entry somewhere in the round's new gaps — failing loudly (a thrown
error, per this report's own R4-2 precedent) if the two lists don't reconcile, rather than trusting
either side alone.

**Required fix:** add one sentence to row 23 (or a new row) naming this residual risk explicitly —
either risk-accept it with the stated rationale (three-round track record of good-faith compliance,
argued not assumed, same as the report already does for other unenforced disciplines), or scope a
cheap structural cross-check (reconcile "regression"-flagged closures against `supersedes` entries
within the same merge turn, throw on mismatch) as part of the fix rather than leaving the schema
field's correctness entirely to prompt instruction. Either is acceptable; silence is not — this
report does not get to warn against unenforced good faith at line 65 and then ship an
unenforced-good-faith fix at row 23 without saying so.

---

## R5-2 — OPEN — MEDIUM-HIGH — medium-high (two of the three throw sites have already fired in
this project's history in some form — the `blueEnv`/`redEnv` null-crash class is exactly run 2's
own defect, and R4-2 adds a third, newly-armed throw site this round) x high (destroys the entire
run's accumulated self-improvement signal, not just the triggering round's) x low-medium (the
naive fix collides with the script's own stated no-filesystem-access design; a compliant fix needs
one extra design decision, not new infrastructure) — corroboration: HIGH (direct read of `debate.js`
lines 1–218 at `42dba2d`; `commands/research.md` step 5 read directly; cross-checked against the
report's own framing of what a "thrown" round costs)

**Location:** §2.1, Tier A (*"Friction aggregation: per-seat arrays namespaced by label and
concatenated, never dropped | source | Self-improvement input integrity [L1+L3]"*), as corrected by
R3-2 in place — and §3 row 2 (*"a mid-debate crash at the judge seat loses every paid round up to
that point, same as the original defect"*).

**Problem:** the report's own R3-2 correction already narrows "never dropped" to name one missing
call site (`blue-synthesize`). It does not address a second, structurally distinct way the same
`friction` array is dropped: **the entire aggregated array is held only in the script's local
scope and is never handed to anything outside the function until the very last statement —
`return { ..., friction }` — which a thrown error never reaches.** Direct read of `debate.js`
confirms three throw sites exist (line 136 `blueEnv` null-guard; the `redEnv` null-guard in the
loop; and, once R4-2's docketed fix lands, the new `FAIL`-with-empty-`gaps` guard at §3 row 20).
Any of the three firing mid-run discards every `friction` entry collected up to that point — not
just the triggering seat's own complaint, but every prior round's `red-merge`/`judge`/
`blue-respond` friction accumulated in the same in-memory array. `commands/research.md` step 5
(*"If the returned envelope carries `friction` entries, write them to
`<run directory>/friction.md`..."*) only fires on a **successful** return; a thrown Workflow-tool
invocation never reaches step 3–5 of the command, so `friction.md` is never written for that run at
all. The irony is exact: the three conditions that throw are the three conditions under which
something has already gone wrong enough to be worth reporting — a quota-walled agent, a malformed
merge, a degenerate FAIL-with-no-gaps — and those are precisely the runs whose friction data would
be most valuable to the self-improvement loop this report elsewhere praises as "functioning as
specified, not aspirational" (§1.4). The report's own language already treats a throw as lossy for
*round* progress (§3 row 2: "loses every paid round up to that point") but has never extended that
same recognition to the `friction` array specifically, and R4-2's newly-decided third throw site
(added this round, on red's own accepted disposition) makes the surface larger, not smaller, without
the report noting the tradeoff.

**Why the obvious fix needs a second look:** the natural fix — have `takeFriction` also append
incrementally to a file as it collects, so a later throw doesn't erase prior entries — runs against
the script's own stated architecture, quoted verbatim in `debate.js`'s header doctrine: *"The lead
is a script: mechanics, round-keeping, termination... the script has no filesystem access by
design."* A script-side incremental write would violate that principle the same round it would fix
this gap. The fix that stays inside the architecture is to move friction capture off the
script-aggregation path entirely: instruct each schema'd seat (`red-merge`, `judge`,
`blue-respond`, and once R3-2 ships, `blue-synthesize`) to append its own friction line directly to
`runDir/friction.md` via the same append-only-blackboard convention already adopted for the
write-block fix (§3 row 8b), in addition to (not instead of) returning it in the envelope for the
final assembly prompt's synopsis — decoupling "friction survives a crash" from "the script's
in-memory array survives to the final return."

**Required fix:** one paragraph, either at §2.1's friction row or as a new §3 row: name the
throw-loses-the-aggregate-friction-array gap, and adopt the agent-writes-directly variant above (or
an argued risk-accept if the team judges crash-time friction low-value relative to normal-path
friction — but that argument does not exist in the report today and should not be assumed by
silence, per this lens's own standing complaint about R4-2's disjunction).

---

## Noted, checked, not raised (round 5)

- **R4-1's four-chain enumeration vs. blue's `debate.md` round-4 BLUE correction:** blue's own
  round-4 response (`debate.md`, tail) records correcting its *own* first-pass chain enumeration to
  match red's, after cross-checking. Verified: the corrected enumeration (R2-5→R3-10, R2-7→R3-6,
  R2-8→R3-5→R4-3, plus R1-5→R2-4→R3-4/R3-9) matches §3 row 23's and §2.1's text exactly. No
  discrepancy found; a good-faith self-correction, not a gap.
- **R4-2's throw semantics vs. loss of the `friction` array specifically:** checked whether the
  throw itself (as opposed to the pre-existing two null-guard throws) introduces a *new* mechanism
  of loss — it does not; it shares the identical mechanism already latent in the two existing
  throws (R5-2 above treats all three as one class, not three separate gaps).
  `parallel()`/`Promise.all` semantics were checked and are irrelevant here — the throw is
  synchronous in the main loop body, not inside a `parallel` callback.
- **Whether R4-2's decision to throw (vs. logged-warning-PASS) could itself be gamed or
  mis-triggered by an over-cautious red-merge pass:** checked against the guard's actual condition
  (`verdict === 'FAIL' && gaps.length === 0`) — this cannot fire on a genuine PASS (handled earlier
  in the loop by the pre-existing `if (redEnv.verdict === 'PASS') break`), so no false-positive
  throw path exists. Held, not raised.
- **Whether §3 row 6's "lanes >= 5" arithmetic (R3-5/R4-3) recurs anywhere else uncorrected:**
  report-wide grep for the retired slash-compound and for "lanes >= 4" — zero remaining instances
  outside the one corrected cell. Held, not raised.
- **Whether the friction-loss gap (R5-2) also affects the assembled top-level `report.md`'s own
  claim of union-not-summary:** checked — the top-level `report.md` is confirmed still a
  35-byte stub pending PASS (unaffected either way; assembly hasn't run yet this trajectory).
- **Cost/incentive angle on R5-1 (considered, not separately raised):** a merge agent facing
  `judgment`-tier cost for judge dispatch has no stated incentive *against* setting `supersedes`
  correctly (judge dispatch is the system's intended remediation path, not a penalty the agent
  would rationally avoid triggering) — this is folded into R5-1's likelihood grade as an
  aggravating-but-unproven factor, not asserted as a distinct incentive-misalignment gap; no
  evidence in this corpus supports agents avoiding costly branches deliberately.

## Disposition

Two new gaps this round, both reflexivity findings against round 4's own fixes: **R5-1** (HIGH) —
the lineage-following fix trusts exactly the unenforced good-faith discipline the underlying defect
report indicts two sentences earlier — and **R5-2** (MEDIUM-HIGH) — three throw sites (two
pre-existing, one newly decided this round) silently discard the run's entire accumulated
self-improvement signal, with no incremental persistence and no stated tradeoff, colliding with the
script's own no-filesystem-access design if fixed naively. Both are report-side-correctable at low
cost (name the residual risk, or scope the structural/agent-side cross-check); neither disputes
H1–H5 or any of round 4's five closures, which hold clean on this lens's own re-verification.
Recommend: **FAIL** stands on these two findings pending at minimum an explicit named disposition
for each (risk-accept-with-rationale is an acceptable closure for either, consistent with this
report's own established pattern — silence is not).
