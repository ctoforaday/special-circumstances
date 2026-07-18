---
name: pattern-repair-regression-citation
description: A repair that softens an unpinnable figure by attributing it to a NEW specific source can regress — the new source may report a materially different (contradicting) number, or the accurate number lives in an uncited paper
metadata:
  classes: [citation-figure-misattribution, incomplete-repair-propagation, undecided-disjunction]
  type: feedback
---

When blue "fixes" an unpinned/uncorroborated statistic by re-citing it to a specific new
source, YOU MUST re-verify the *new* source at the leaf node — the repair is a fresh
statement↔reference pair, not a closure.

**Why:** In round 2 of the memory-architecture audit, blue's R1-28 repair softened an
unpinnable "80–99% attack success" band to "up to ~90–95% (MINJA / environment-injection),
attributed." Following the two newly-cited footnotes: (1) the environment-injection paper
(arXiv 2604.02623, *Poison Once Exploit Forever*) reports ASR ≤32.5%, not ~90% — the repair
*introduced* a leaf-node contradiction that did not exist in the vaguer original; (2) the
accurate MINJA ~95%/~70% figure is real but lives in an uncited paper (arXiv 2503.03704),
while the survey blue did cite (2606.04329) states it gives no ASR numbers at all. So the fix
made one half contradicted and left the accurate half untraceable.

**How to apply:** Treat every R-numbered "repair" as a new claim to verify, never as a
resolved gap to trust. A softened/hedged number is still a number a skeptic must be able to
follow. "Contradicted at leaf node" is a real regression even when the verdict's disposition
survives (here R1-11 meant the blocking grade did not rest on the figure) — grade it, raise it,
do not let it stand just because it is non-load-bearing. Escalate the round-1 gap id rather
than inventing a wholly new one, so the thread stays navigable.

**Extension (FEOV-retrospective round 2) — adjacent-narrative count transposition:** a repair
touching a graded cell can import an attribute (here a "third occurrence" recurrence count)
from a *structurally adjacent* defect narrative that legitimately carries it (write-block and
ENAMETOOLONG were the paired Tier-B defects, discussed side by side everywhere). Blue's R1-13
fix re-graded ENAMETOOLONG's likelihood on "3 occurrences... per debate.md's merge-seat
friction" — the cited transcript section contains no such event; the true count was 2, and the
"third occurrence" language belonged to the write-block's separately-correct ledger. When two
defect stories travel together in a report, verify each one's counts/dates against its OWN
sources — proximity is a contamination vector during repairs. Same round also produced a
within-repair chronology self-contradiction ("this same round" vs. "two consecutive rounds" in
one paragraph): read the whole repaired paragraph for internal consistency, not just the
repaired clause.

**Red-side extension (FEOV-retrospective rounds 2–3) — red's own text is a claim surface too:**
two ways the adversary's own words became report defects. (1) Red's R2-4 `required_fix` proposed
a specific replacement citation (arXiv:2606.02646) for the "7 agents" figure *without leaf-
verifying it first*; blue fetched it, found it does NOT contain the figure (knee N≈10, plateau
~1.8 by N=30 — opposite direction), and rebutted the proposed fix while conceding the gap — the
correct outcome, but red proposed a citation that would have been a new miscitation. Leaf-verify
any source you name in a `required_fix` before proposing it. (2) Red's round-2 merge wrote "grep
'independent': zero hits outside the ledger line's own text" — an unverified-as-worded
characterization (actual result: zero hits INCLUDING the ledger clause); blue copied it verbatim
into the row-19 fix, and the imprecision survived two rounds before a round-3 lens re-ran the
grep. Blue copies red's characterizations verbatim when repairing — an imprecise red sentence
becomes a report defect with red's own authority behind it. Re-verify your own prior round's
phrasings at the same leaf-node standard as blue's.

**Red-side extension (FEOV-retrospective round 4) — an unresolved "or" in a required_fix ships
as an undecided fix:** red's R3-1 required fix said "treat as effective PASS-with-warning or
throw a distinguishing error." Blue's round-3 repair (§3 row 20) carried the disjunction forward
*verbatim* — the only fix that round that shipped an "or" where its siblings shipped decisions —
and the two branches were opposite failure philosophies (silently convert a degenerate FAIL
toward PASS vs. halt loudly). Third instance of blue shipping red's phrasing verbatim. When a
required_fix contains alternatives, YOU MUST name red's favored side (and why), while still
accepting an argued choice of the other branch — otherwise the disjunction itself becomes the
shipped behavior spec.

**Extension (run 4, round 2) — two more repair-regression shapes:** (1) *repair relocates the
defect*: a fix for an impossible sink ("emit into cost.md" — script has no fs access) replaced it
with a different unverified sink ("log() into journal.jsonl, consumed by cost-audit.mjs") — both
halves contradicted by first-hand check; verify a repaired mechanism claim end-to-end, not just that
the old error is gone. (2) *insertion severs the host sentence*: a correction record spliced into
the middle of a round-0 sentence orphaned its second clause ("But was measured-robust in this
loop" — no subject); after any inline correction-record insertion, read the WHOLE host paragraph
aloud for grammatical integrity, not just the inserted text. Also positive signal worth keeping:
the highest-risk repair class (re-citation to a brand-new source, R1-5→CathedralBazaar) verified
clean at the PDF leaf — arXiv-HTML garbled the key sentence ("at least 11" for "at least 1
*vector*"); when an HTML digest returns a semantically impossible number (CVSS has 8 base metrics),
suspect italic/subscript rendering artifacts and go to the PDF via pdftotext before flagging.

**Extension (run 4, round 4) — a repair's forensic DIAGNOSIS is a claim too:** blue's R3-1 fix
corrected a 4×-overstated dollar series (band right, floor/ceiling reproduce) but asserted the
bad series "reproduces only if cache-weighted BYTES are priced as tokens." Re-derivation from
the committed instrument: bytes-as-tokens gives $1.04/2.12/3.56/4.64 — NOT the printed
$1.40/2.60/4.10/4.10; cache-weighted-share × whole-merge-dollars regenerates both the printed
series and red lens-3's sibling recompute to ≤3%, and red's own ledger had recorded lens-3's
method as "share × merge $." When a repair explains WHY a prior figure was wrong, re-generate
the wrong figure under the claimed mechanism; if it doesn't regenerate, the diagnosis
misattributes — and may mischaracterize a sibling seat's ledgered method with red's authority
attached. Corrected numbers reproducing does not validate the error story wrapped around them.

**Merge-seat extension (run 4, round 2) — "matches red's required fix exactly" is NOT leaf
verification:** when red's own round-N finding embedded an error (R1-27 asserted a phrase lived
verbatim in TWO files; debate.js l.263 actually says "5 regressions", not "5 chains"), a
round-N+1 lens verified the repair HIGH on *fidelity to red's instruction* while a sibling lens's
first-hand source read contradicted the copied content. At merge, a fidelity-based grade and a
source-based grade are different evidence tiers: the source read wins, always — re-run it
first-hand before overruling. Corollary: every claim red's required_fix asserts about a source
(not just citations it proposes) must itself be leaf-verified, because blue copies it AND a
future lens may "verify" the copy against the instruction instead of the source, laundering
red's error through two seats.

Related: [[citation_status_and_misattribution_patterns]] (real figure miscited to wrong
source), [[gap_live_source_drift]] (re-follow to current primary),
[[pattern-within-source-condition-misattribution]] (right paper, wrong experimental arm),
[[pattern-identity-keyed-detector-lineage-blind]] (red's fresh-id convention defeating the
docket detector — another red-practice-becomes-system-defect case).
